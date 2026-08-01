# Project 095 — Microservice with Event-Driven Outbox

## 1. Project Name and Number
Project 095, `095_microservice_event_driven`. The folder name is fixed by the curriculum table; do not rename the directory.

## 2. Project Idea
An Orders service that, in the same PostgreSQL transaction that creates an Order, also writes a versioned Outbox row representing the event the order should produce. A separate Publisher polls committed unpublished outbox rows in a deterministic per-aggregate order, claims each row safely with PostgreSQL row locking that satisfies a head-of-line predicate per aggregate, publishes each row to an injected broker, and marks the row `Published` only after the broker's acknowledgement. A Consumer validates the event's schema and version, then in the same PostgreSQL transaction that performs the side effect, attempts to insert a unique inbox row and either performs the side effect or no-ops on duplicate. Delivery is at-least-once; duplicates are absorbed by the inbox. Ordering is guaranteed only per aggregate key through both publisher serialization and the pinned broker-key assumption. The broker is never called inside `PlaceOrder` or before commit. A poison message or unknown schema version becomes a consumer `DeadLettered` outcome after a bounded number of deliveries; this is separate from the outbox row's `Failed` publisher state.

## 3. Why This Project Now?
This project is the transactional messaging capstone. It pulls the discipline of "atomic with the order" and the discipline of "publish after commit" into a single program, and forces honest accounting of what at-least-once delivery and per-aggregate ordering actually buy. The previous project pulled together the typed outcomes of a single-Redis lease and the absolute-secret rule for its token; this project pulls together transactional outbox semantics and conflict-aware inbox handling. The formal prerequisites — project 064 (database migrations), project 066 (transaction manager), and project 086 (distributed task queue) — each contribute one ingredient that this service combines. The immediate catalog predecessor, project 094, is optional context rather than a formal prerequisite.

## 4. Prerequisites
The formal prerequisites are projects 064, 066, and 086; project 094 is the immediate catalog predecessor and remains useful as optional context rather than a formal prerequisite.

## 5. What You Must Know Before Starting
- That a database transaction is the only atomic boundary available without a distributed transaction coordinator, and that the Outbox pattern trades "publish and commit" for "commit then publish" by writing the publish intent into the same transaction.
- That a unique constraint on the inbox table is the canonical way to make a side effect idempotent under at-least-once delivery.
- That PostgreSQL row locking with `SELECT ... FOR UPDATE SKIP LOCKED` is the canonical way to let multiple publishers contend for outbox rows without stepping on each other.
- That a broker's acknowledgement is a hint about delivery, not a guarantee of side-effect ordering, and that an at-least-once broker may redeliver after a crash.
- That schema versions let a consumer reject messages it does not understand without crashing.
- That per-aggregate ordering depends on two boundaries: the publisher must serialize each aggregate's outbox head, and the broker must preserve send order for a shared aggregate key.
- That a poison message is a message that fails processing repeatedly; the consumer must move it aside after a bounded number of attempts.
- That "publish before commit" and "mark before ack" are both bugs.

## 6. Explanation of New Concepts
- Transactional outbox: `PlaceOrder` writes an Order and one Outbox row in the same PostgreSQL transaction. The Outbox row is the durable intent to publish. The broker is never called inside `PlaceOrder` and never called before commit.
- Pinned outbox row shape: each row carries a unique event identifier, an aggregate or order identifier, an aggregate-scoped sequence number, an event type, a schema version, an occurred-at timestamp, a JSON payload, a status chosen from `Pending`, `Published`, or `Failed`, a publisher attempt counter, and the necessary timestamps. `(aggregate_id, aggregate_sequence)` is unique. The event identifier is unique across rows.
- Pinned outbox states: a row is in exactly one of `Pending`, `Published`, or `Failed`. No other state values exist. `Pending` rows are eligible for the publisher. `Published` rows are terminal for the publisher. `Failed` rows block later rows for the same aggregate until an explicit operator policy resolves them; no silent reordering.
- Aggregate-scoped sequence allocation under a per-aggregate lock: the sequence for an aggregate is allocated inside `PlaceOrder`'s transaction under an aggregate-scoped row lock or counter so that two concurrent orders for the same aggregate cannot allocate the same sequence and cannot allocate out-of-order sequences. Two concurrent orders for different aggregates do not contend on the sequence allocator.
- Event JSON validation before insert: the JSON payload is validated before insert so a malformed payload fails `PlaceOrder` and never enters the outbox.
- Multi-publisher claiming with a head-of-line predicate: for each aggregate, the head is the minimum-sequence row whose status is not `Published`. That head is eligible only when its status is `Pending`; if it is `Failed`, the aggregate is blocked. The claim query applies this predicate, deterministic order, and `FOR UPDATE SKIP LOCKED`, and holds the selected row lock through publish plus mark. Two publishers can process different aggregates concurrently but cannot publish sequence N+1 while an earlier sequence is `Pending`, claimed, or `Failed`. Row locking alone without this predicate is insufficient.
- Pinned broker assumption: the broker accepts messages with the same aggregate identifier in send order and may redeliver on crash; there is no global order across aggregates. The publisher supplies the aggregate identifier as the message key. The per-aggregate database head-of-line rule plus the broker key assumption together define the ordering guarantee.
- Publisher flow: the publisher calls the injected broker while holding the claim transaction, then marks the row `Published` only after acknowledgement, then commits. A broker failure increments the publisher attempt and keeps the row `Pending` until the bound; at the bound the row becomes `Failed`. A crash or a database commit failure after broker acknowledgement but before the durable `Published` commit produces redelivery or a duplicate on the next publisher pass, which is expected at-least-once behavior.
- Consumer flow with conflict-aware inbox: the consumer begins a transaction and attempts the inbox insert with conflict-aware semantics. If the inbox insert reports already present, the consumer commits the no-op without performing the side effect and acknowledges success; this is the duplicate signal. If the inbox insert is newly performed, the consumer performs the database side effect in the same transaction, then commits. The consumer does not catch a unique violation inside an already-aborted transaction and continue.
- Consumer side-effect failure: a side-effect failure rolls back both the new inbox row and the side effect so the message is eligible for redelivery. Unknown schema or unknown version is rejected before payload decoding and before any side effect, and becomes the typed consumer outcome `DeadLettered` after the bounded delivery-attempt policy.
- Bounded delivery-attempt policy in the consumer: transient failures from the broker adapter retry; permanent or unknown-version failures become `DeadLettered` only after the bound, so there is no infinite hot loop. This delivery outcome is recorded by the consumer adapter or fake and is not a fourth outbox status. The publisher's attempt budget and the consumer's delivery-attempt budget are separate.

## 7. Learning Objective
After completing this project you must be able to explain in your own words: why the Order and the `Pending` outbox row are atomic, why the broker is never called inside `PlaceOrder` or before commit, why the sequence is allocated under an aggregate-scoped lock, why the claim query must enforce a head-of-line predicate per aggregate, why a `Failed` outbox head blocks its aggregate, why `SKIP LOCKED` plus deterministic order is necessary, why at-least-once is the honest delivery model for this pattern, why the inbox makes the consumer idempotent, why ordering is per aggregate and not global, why schema versions exist on events, why a poison or unknown-version delivery becomes `DeadLettered` after the bound, why the broker can remain a deterministic fake in the optional integration test, and why the publisher's attempt budget and the consumer's attempt budget are separate.

## 8. Functional Requirements
1. The schema includes an `orders` table, an `outbox` table, and an `inbox` table. Migrations bring the schema up and down.
2. `PlaceOrder` writes the Order and the `Pending` Outbox row in the same PostgreSQL transaction. The broker is never called inside `PlaceOrder` and never called before commit. The JSON payload is validated before insert.
3. Each Outbox row carries the unique event identifier, the aggregate identifier, an aggregate-scoped sequence, the event type, the schema version, the occurred-at timestamp, the JSON payload, the status from `Pending`, `Published`, or `Failed`, the publisher attempt counter, and the necessary timestamps. `(aggregate_id, aggregate_sequence)` is unique. The event identifier is unique across rows.
4. Sequence allocation for an aggregate occurs inside `PlaceOrder`'s transaction under an aggregate-scoped row lock or counter.
5. For each aggregate, the Publisher identifies the minimum-sequence row whose status is not `Published`. It may claim that head only when the head is `Pending`; a `Failed` head blocks later rows. The query uses deterministic order and `FOR UPDATE SKIP LOCKED`. The Publisher holds the row lock through the publish-plus-mark transaction.
6. The Publisher calls the injected broker while holding the claim transaction, marks `Published` only after the broker's acknowledgement, and then commits. The Publisher never publishes before commit and never marks before acknowledgement.
7. The Publisher increments the attempt on a broker failure and keeps the row `Pending` until the bound. At the bound the row becomes `Failed`. A `Failed` row blocks later rows for the same aggregate until an explicit operator policy resolves it; no silent reordering.
8. The broker used for at-least-once delivery and per-aggregate ordering is pinned to its assumption: messages with the same aggregate identifier are accepted in send order and may redeliver; there is no global order. The publisher supplies the aggregate identifier as the broker message key.
9. The Consumer validates the schema and version before payload decoding and before any side effect. An unknown schema or version becomes `DeadLettered` after the consumer's bounded delivery attempts. `DeadLettered` is a consumer delivery outcome, not an outbox status.
10. The Consumer begins a transaction and attempts the inbox insert with conflict-aware semantics. If the insert reports already present, the Consumer commits the no-op and acknowledges success. If the insert is newly performed, the Consumer performs the side effect in the same transaction and commits. The Consumer does not catch a unique violation inside an already-aborted transaction and continue.
11. The Consumer side-effect failure rolls back the inbox row and the side effect together so retry is possible. Transient delivery failures retry; permanent or unknown-version failures become `DeadLettered` after the bound. There is no infinite hot loop on the consumer side or on the publisher side.
12. The Publisher's pass is bounded per pass with a fixed batch size and a fixed maximum attempts per row. The Publisher does not block forever on a single hot row.
13. A crash or database commit failure after broker acknowledgement but before the durable `Published` commit produces redelivery or duplicate, which is at-least-once and is expected.
14. The status values of an outbox row are exactly `Pending`, `Published`, `Failed`. No other values exist.

## 9. Inputs and Outputs
- `PlaceOrder` input: an order identifier and the order's contents. Output: a stored Order, a stored Outbox row with status `Pending`, and a transaction commit.
- Publisher output: a sequence of broker publishes, each followed by a database update that sets `Published` and the published timestamp, after the broker has acknowledged. A failed publish leaves the row `Pending` and increments the publisher attempt; at the bound the row becomes `Failed`.
- Consumer input: an event payload from the broker including the unique event identifier, the aggregate identifier, the per-aggregate sequence, the event type, the schema version, and the JSON payload.
- Consumer output: a side effect in the database, an inbox row keyed by the unique event identifier, and a database commit on success. On duplicate delivery the inbox insert reports already present, the side effect is skipped, and the consumer acknowledges success.
- Behavior examples:
  - Place an order, observe one `Pending` outbox row. Run the publisher, observe the broker receive one message, then observe the row updated to `Published`.
  - Stop the publisher after the broker publish and before the database commit. Restart. The broker receives the same message again. The consumer absorbs the duplicate through the inbox.
  - Deliver a message with an unknown schema version. The consumer rejects it before payload decoding and, after the delivery bound, records the consumer outcome `DeadLettered` without changing the outbox status model.

## 10. Rules and Edge Cases
- `PlaceOrder` that fails inside the transaction leaves no Order and no outbox row.
- The broker is never called inside `PlaceOrder`.
- Concurrent orders for the same aggregate allocate distinct sequences in order under the aggregate-scoped lock.
- Two publishers concurrently claim disjoint aggregates; neither publishes sequence N+1 for an aggregate while N is `Pending` or claimed.
- A duplicate delivery is consumed by the inbox conflict-aware path; the side effect runs exactly once.
- A side-effect failure rolls back the inbox row and the side effect together.
- An unknown schema version is rejected before payload decoding.
- An empty outbox is a normal state; a publisher pass completes with no broker calls.
- A malformed JSON payload is rejected by `PlaceOrder` before insert; the broker never sees the event.
- A `Failed` row blocks later rows for the same aggregate until an explicit operator policy resolves the failure.

## 11. Project Constraints
- The database boundary is `database/sql` using the `pgx` standard-library adapter. At implementation time, select a currently supported `pgx` release and pin that version in the module; this guide does not invent a patch version.
- The broker is injected. In unit tests and in the optional integration test the broker is a deterministic fake. Project 098 introduces a real broker; this project does not.
- The publisher's invariants are enforced by code: "no publish before commit" is enforced by reading the outbox table after commit; "no mark before ack" is enforced by the control flow.
- The consumer's idempotency is enforced by a unique constraint on the inbox table.
- The publisher's per-row attempt counter is incremented on failure; the row becomes `Failed` after the bound. The consumer's attempt budget is separate.
- Unit tests must run locally with no Docker and no broker. They use a fake repository or fake driver and a fake broker.
- The opt-in integration test runs against real PostgreSQL started by Compose. Behavior proven: transaction rollback, unique inbox insert, aggregate-sequence allocation, head-of-line selection, `FOR UPDATE SKIP LOCKED`, and concurrent publishers. The broker remains a deterministic fake. The integration test is gated by a build tag and an environment flag.
- No Kafka until project 098. The broker assumption is pinned to a deterministic per-aggregate-key ordering plus at-least-once delivery.

## 12. Design Questions Before Coding
- Which event identifier format do you choose, and why does it support broker deduplication and inbox uniqueness?
- Which JSON column type do you use, and at which boundary do you validate?
- Which currently supported `pgx` version will you pin for its `database/sql` adapter, and how will you verify its PostgreSQL compatibility?
- How do you express the aggregate-scoped sequence allocator so concurrent orders for the same aggregate cannot collide?
- How do you encode the head-of-line predicate per aggregate into your claim query, and how do you hold the lock through publish plus mark transaction?
- How does the consumer distinguish transient from permanent failure, and how do you pin a bounded delivery-attempt policy in the injected broker adapter?
- How do the unit tests prove the no-publish-before-commit and no-mark-before-ack invariants without a real broker?
- How does the optional integration test prove head-of-line selection and concurrent publishers without a real broker?

## 13. Implementation Milestones
1. Design the schema: `orders`, `outbox` with status from `Pending`, `Published`, `Failed`, and `inbox`. Write the up and down migrations. Include the uniqueness constraints on the event identifier and on `(aggregate_id, aggregate_sequence)`.
2. Implement the aggregate-scoped sequence allocator that runs inside `PlaceOrder`'s transaction and prevents collision and out-of-order allocation.
3. Implement `PlaceOrder` with payload validation before insert and the same-transaction write of Order plus `Pending` outbox row. The broker is never called.
4. Implement the Outbox row model with the pinned status set and the deterministic JSON encoder.
5. Implement the Publisher's claim query: deterministic order, head-of-line predicate per aggregate, `FOR UPDATE SKIP LOCKED`, lock held through publish plus mark transaction.
6. Implement the Publisher's publish loop: claim a batch, publish each row to the injected broker, mark `Published` only after acknowledgement, then commit. Increment the attempt on broker failure and move to `Failed` at the bound.
7. Implement the Consumer's schema and version validation that rejects unknown versions before payload decoding.
8. Implement the Consumer's conflict-aware inbox flow: begin a transaction, attempt the inbox insert, no-op on already-present, perform the side effect and commit on newly-inserted.
9. Implement the Consumer's bounded delivery-attempt policy. Transient failures retry; permanent or unknown-version failures become the separate consumer outcome `DeadLettered` after the bound.
10. Wire the wiring so that Orders, Publisher, and Consumer each receive their dependencies through injection. The Publisher's invariants are enforced in code.
11. Write the unit test suite with fake repository and fake broker, proving atomicity, rollback, unpublished polling, broker failure, crash-window replay, duplicate consumer, schema rejection, per-aggregate ordering, head-of-line selection, concurrent publishers, and the no-publish-before-commit and no-mark-before-ack invariants.
12. Write the opt-in integration test against real PostgreSQL started with Compose. The broker remains a deterministic fake. The test proves transaction rollback, unique inbox insert, aggregate-sequence allocation, head-of-line selection, `FOR UPDATE SKIP LOCKED`, and concurrent publishers.

## 14. Verification Cases the Learner Must Write
- Atomic Order plus outbox: a successful `PlaceOrder` leaves both rows; a failure inside the transaction leaves neither.
- Rollback: a simulated database failure during the transaction leaves the database in its prior state.
- Sequence allocation: concurrent orders for the same aggregate allocate distinct sequences in order; concurrent orders for different aggregates do not contend.
- Head-of-line selection: an earlier `Pending`, claimed, or `Failed` row blocks later rows for its aggregate; an earlier `Published` row does not. A `Failed` head yields no claim for that aggregate until an explicit operator policy resolves it.
- Unpublished polling: the publisher finds the minimum-sequence non-`Published` head per aggregate and claims it only when it is `Pending`, in deterministic order; an empty outbox returns an empty batch.
- Broker failure: a broker that returns an error leaves the row `Pending` and increments the publisher attempt; at the bound the row becomes `Failed`.
- Crash-window replay: a sequence in which the broker publishes but the database commit is skipped produces a duplicate on the next pass; the consumer absorbs the duplicate via the inbox.
- Duplicate consumer: the same event delivered twice inserts the inbox row once and performs the side effect once.
- Side-effect failure: a side-effect failure rolls back the inbox row and the effect together so the message is eligible for redelivery.
- Schema rejection: an event with an unknown schema version is rejected before payload decoding and becomes `DeadLettered` after the delivery bound without introducing another outbox status.
- Per-aggregate ordering: events for the same aggregate are published in per-aggregate sequence order; events for different aggregates may interleave.
- Multi-publisher: two publisher instances running concurrently each claim disjoint aggregates; neither marks the other's rows.
- No publish before commit: the publisher's first action in a unit test is a database read of committed rows; the broker is never called inside `PlaceOrder` and never called before commit.
- No mark before ack: a fake broker that fails after publish leaves the row `Pending`; the publisher does not mark it `Published`.
- Conflict-aware inbox: a programmed duplicate delivery takes the already-present branch and commits without a side effect; a fresh delivery takes the newly-inserted branch and commits with the side effect.
- Integration opt-in: with the gating flag and the gating build tag set, real PostgreSQL proves transaction rollback, unique inbox insert, aggregate-sequence allocation, head-of-line selection, `FOR UPDATE SKIP LOCKED`, and concurrent publishers; the broker remains a fake.

## 15. Common Mistakes to Watch For
- Calling the broker inside `PlaceOrder` or before commit.
- Marking `Published` before the broker acknowledges.
- Treating a duplicate event as an error rather than absorbing it through the inbox.
- Using a non-unique inbox key.
- Hot-looping on a poison message; the publisher moves its outbox row to `Failed`, while the consumer records the separate `DeadLettered` delivery outcome after its own bound.
- Conflating the publisher's attempt budget with the consumer's attempt budget.
- Locking outbox rows without enforcing the head-of-line predicate per aggregate.
- Letting sequence allocation race under contention and produce collisions or out-of-order allocations.
- Accepting a malformed JSON payload into the outbox.
- Using a third broker state beyond `Pending`, `Published`, `Failed`.
- Letting the consumer catch a unique violation inside an already-aborted transaction and continue.
- Calling the pattern "exactly-once"; the pattern is at-least-once with idempotent consumers.
- Adding a real broker to the optional integration test.
- Inventing a database driver version in the documentation; the learner selects and pins a currently supported version.

## 16. Topics and References for Study
- The PostgreSQL documentation for `SELECT ... FOR UPDATE SKIP LOCKED`, transactional `INSERT`, and unique-constraint conflict handling.
- The Chris Richardson "Microservices Patterns" chapter on the Transactional Outbox.
- The Debezium documentation on change-data-capture-based outbox publishing, used only as background; this project uses polling.
- The `database/sql` documentation for the driver you pin.
- The migration discipline from project 064, applied to bring the schema up before any publisher or consumer runs.

## 17. Self-Assessment Questions
- Why must the Order and the `Pending` outbox row be written in the same transaction, and why is the broker never called inside `PlaceOrder` or before commit?
- Why is the sequence allocated under an aggregate-scoped lock so concurrent orders for the same aggregate cannot collide or allocate out of order?
- Why does the claim query enforce a head-of-line predicate per aggregate together with `FOR UPDATE SKIP LOCKED` in deterministic order?
- Why is the delivery model at-least-once and not exactly-once?
- Why does the inbox make the consumer idempotent, why must it be conflict-aware, and what is wrong with catching a unique violation inside an aborted transaction?
- Why is ordering per aggregate and not global?
- Why is the broker message key the aggregate identifier?
- Why does the schema version exist on each event?
- Why is the `Failed` state terminal for the row and what blocks later rows for the same aggregate until an explicit operator policy resolves the failure?
- Why must the publisher's and consumer's attempt budgets be separate?

## 18. Definition of Completion
The project is complete when the schema is brought up by a versioned migration; when `PlaceOrder` writes the Order and a `Pending` outbox row in the same transaction and never calls the broker; when sequence allocation under an aggregate-scoped lock prevents collisions and out-of-order allocations; when the publisher's claim query selects only a `Pending` minimum-sequence non-`Published` head per aggregate, leaves a `Failed` head blocking, uses `FOR UPDATE SKIP LOCKED` in deterministic order, holds the row lock through publish plus mark, marks `Published` only after acknowledgement, increments the attempt on failure, and moves the outbox row to `Failed` at the bound; when the consumer validates schema and version before decoding, uses the conflict-aware inbox, no-ops on duplicate, performs the side effect on a fresh insert, rolls back both on side-effect failure, and records `DeadLettered` after its separate bound for permanent or unknown-version deliveries; when the unit suite passes locally with no Docker and no broker; when the opt-in PostgreSQL integration suite passes with its gates and the broker remaining a fake; when the honest at-least-once statement is reproduced in the project's own documentation; and when no real broker is introduced before project 098.

## 19. Optional Extensions
- A second event type added to the same outbox table, with the Consumer dispatching on event type after version validation; the inbox uniqueness key is unchanged and the publisher's loop is unchanged.
- A small set of read-model projections written by the Consumer, each in its own transaction with its own inbox row, demonstrating that one event can drive multiple projections idempotently.
