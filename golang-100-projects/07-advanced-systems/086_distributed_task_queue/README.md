# Project 086 — Distributed Task Queue

## 1. Project Name and Number

- Project 086, distributed_task_queue.
- Build a deterministic in-memory simulation of one task queue shared by exactly two logical workers.
- This README is a learning guide only.
- It contains no implementation code, signatures, snippets, pseudocode, or solution commands.

- > **Scope.** Despite the directory name, this is not a production distributed broker.
- It has no network transport, external process coordination, disk persistence, broker protocol, or exactly-once guarantee.
- All queue state disappears when the process ends.

## 2. Project Idea

Model at-least-once task delivery with explicit task states, exclusive leases, token-checked acknowledgements, deterministic retries, and a dead-letter queue. Every task has one globally unique task ID that also serves as its idempotency key, a bounded payload, an attempt count, an absolute ready time, an immutable enqueue sequence, and one state. Two logical workers contend through the same mutex-protected state machine. Tests control time, token generation, handler outcomes, and lease reclamation without sleeping.

## 3. Why This Project Now?

- This project takes its required foundation from Projects 034 (worker_pool_basic), 041 (context_timeout_example), and 065 (redis_caching_layer): Project 034 supplies bounded worker-pool ownership, Project 041 supplies context cancellation and deadlines, and Project 065 supplies failure-aware external-cache and dependency boundaries, explicit partial outcomes, and Redis vocabulary.
- The catalog's immediate predecessor is Project 085 (packet_sniffer_basics); Project 085 is referenced here only as optional context for its bounds-first processing, typed failure outcomes, deterministic fixtures, and strict separation between supported behavior and unsafe production claims.
- This project applies both sets of ideas to mutable concurrent state: every transition has a guard, every retry has a deterministic time, and every delivery guarantee states its failure window honestly.

## 4. Prerequisites

- Projects 034 (worker_pool_basic), 041 (context_timeout_example), and 065 (redis_caching_layer) are the required prerequisites.
- Project 085 (packet_sniffer_basics) is the immediate catalog predecessor and is useful for context but is not required.
- Be comfortable with mutex-protected shared state, injected dependencies, finite-state machines, stable ordering, contexts, fake clocks, deterministic test doubles, and race-detector testing.
- Review the distinction between a queue delivery and a handler side effect before starting.

## 5. What You Must Know Before Starting

- Know that at-least-once delivery permits duplicate handler calls.
- A lease grants temporary processing authority; it does not transfer ownership forever.
- A lease token prevents an old worker from acknowledging a task after authority moved to another worker.
- An attempt is consumed when a lease is issued, not when a handler finishes.
- Idempotency belongs at the side-effect boundary because a queue cannot atomically commit an arbitrary external effect and its own acknowledgement.
- Also understand mutex invariants, overflow-safe duration arithmetic, deterministic priority ordering, and why tests driven by real time or goroutine scheduling are unreliable.

## 6. Explanation of New Concepts

### Concepts

- The queue accepts at most 10,000 task identities during its lifetime.
- A task ID is non-empty UTF-8 text of at most 128 bytes and is the task's idempotency key.
- A payload is at most 64 KiB and is copied on enqueue and delivery.
- Worker IDs are non-empty UTF-8 text of at most 128 bytes.
- The two configured worker IDs must be distinct.
- Duplicate task IDs are rejected without mutation, including IDs already completed, cancelled, or dead-lettered.
- Terminal records remain queryable until the simulation is discarded, which makes duplicate detection and duplicate-ack behavior deterministic.

- The task states are Ready, Leased, Completed, DeadLettered, and Cancelled.
- Enqueue creates Ready with attempt count zero, an absolute ready time, and the next monotonically increasing enqueue sequence.
- Ready may become Leased or Cancelled.
- Leased may become Ready after a retryable failure or expired lease, Completed after a valid acknowledgement, DeadLettered after the final failed attempt, or Cancelled.
- Completed, DeadLettered, and Cancelled are terminal.
- No transition leaves a task in two collections.

- Dequeue is one atomic critical section.
- At one clock observation it selects the eligible Ready task with the earliest ready time; ties use the lower immutable enqueue sequence.
- A task is eligible when its ready time is less than or equal to the observed time.
- Selection changes the task to Leased, increments its attempt count, increments its lease version, assigns the requesting worker, obtains a fresh opaque random token of at least 128 bits from the injected token source, and records an exclusive expiration.
- The lease duration is positive and validated before the simulation starts.
- If token generation or expiration arithmetic fails, dequeue makes no mutation.

- A lease is valid only before its expiration and only for the exact worker ID, token, and version stored on the task.
- The lease becomes invalid when observed time is equal to or later than its expiration, even if explicit reclamation has not run yet.
- A valid acknowledgement changes Leased to Completed.
- A valid negative acknowledgement records a non-empty error of at most 1,024 bytes and performs the retry-or-dead-letter transition.
- Any mismatched, expired, superseded, or otherwise old authority is stale and is rejected without mutation.

- Acknowledgement is idempotent only in one narrow case.
- After a valid acknowledgement reaches Completed, an exact repeat carrying the same worker ID, token, and version that completed the task returns the same successful outcome and changes nothing.
- This handles a lost acknowledgement response.
- Every other acknowledgement against a terminal task is rejected.
- Negative acknowledgement has no terminal duplicate-success exception.

- Attempts are capped by an injected retry policy whose validated maximum is between one and ten inclusive.
- The required deterministic policy uses a positive base delay, a maximum delay no smaller than the base delay, no jitter, and exponential growth: after failed attempt one it uses one base delay, after failed attempt two it uses twice the base delay, and it continues doubling with overflow-safe saturation at the maximum delay.
- A custom injected policy must return a non-negative delay no greater than its declared maximum.
- Retry ready time is the failure observation time plus that delay.
- Retries retain the original enqueue sequence.

- When a valid negative acknowledgement or lease reclamation observes an attempt below the maximum, the task returns to Ready.
- When the consumed attempt equals the maximum, the task moves exactly once to DeadLettered and retains its attempt count and last error.
- Lease expiry uses the fixed last error meaning lease expired.
- An explicit reclaim operation observes the clock once, processes all expired Leased tasks in expiration-time order with enqueue sequence as the tie-breaker, and applies the same retry or dead-letter rule.
- There is no background ticker.

- Cancellation is also explicit.
- Cancelling Ready changes it to Cancelled and removes it from eligibility.
- Cancelling Leased changes it to Cancelled and invalidates the lease; a later acknowledgement is stale.
- Repeating cancellation of Cancelled is harmless.
- Cancelling Completed or DeadLettered reports the existing terminal outcome and changes nothing.
- Queue cancellation cannot undo a handler side effect that has already started.

- A duplicate delivery can occur when a worker performs its side effect and crashes before acknowledgement.
- After lease expiry and reclamation, another worker may receive the same task.
- The handler therefore uses the task ID as an idempotency key in its own side-effect ledger: the first observation applies the effect and a later observation returns the prior result without applying it again.
- This is a handler contract, not an exactly-once queue promise.

- All queue transitions and inspections share one mutex-defined consistency boundary.
- No callback, handler, clock advancement, or blocking external work occurs while that mutex is held.
- At most one unexpired valid lease can exist for a task.
- At-least-once delivery means an accepted non-cancelled task remains eligible for attempts while explicit dequeue and reclaim operations continue, until it is acknowledged or exhausts its attempt cap; it does not mean infinite retries, durable retention, or exactly-once effects.

- Text-only state examples are permitted.
- A successful path is Ready, then Leased on attempt one, then Completed.
- A crash path is Ready, then Leased on attempt one, then an expired invalid lease, then Ready after reclaim and backoff, then Leased on attempt two.
- A final failure path ends in DeadLettered once.
- A worker holding the first lease cannot acknowledge after the second lease exists.

## 7. Learning Objective

- Design and verify a mutex-safe, deterministic in-memory task queue simulation for two logical workers that demonstrates at-least-once delivery, exclusive expiring leases, stale-token rejection, idempotent acknowledgement responses, capped exponential retry, dead-lettering, cancellation, stable ready-time ordering, and handler-side idempotency without claiming persistence, exactly-once delivery, or production distribution.

## 8. Functional Requirements

1. Configure exactly two distinct logical worker IDs, an injected clock, an injected token source, a positive lease duration, and a validated deterministic retry policy.
2. Accept at most 10,000 globally unique task IDs; each ID is non-empty UTF-8 text of at most 128 bytes and also serves as the idempotency key.
3. Copy every payload on enqueue and delivery; reject payloads larger than 64 KiB without mutation.
4. Store attempt count, absolute ready time, immutable enqueue sequence, state, lease metadata when leased, and terminal metadata when applicable.
5. Order eligible tasks by ready time and then enqueue sequence. A task is eligible at the exact boundary where observed time equals ready time.
6. Lease selection, state change, attempt increment, version increment, token assignment, and expiration assignment are one atomic transition.
7. Treat a lease as valid only before expiration and only for its exact worker ID, random token, and version.
8. Permit only a valid current lease to acknowledge or negatively acknowledge. Reject stale authority without mutation.
9. Treat an exact repeat of the acknowledgement that completed a task as successful and harmless; reject every other terminal acknowledgement.
10. Compute retries from the injected policy and one clock observation. Use deterministic capped exponential backoff with no jitter for the required implementation.
11. Move a task to DeadLettered exactly once when its final consumed attempt fails or expires, retaining the bounded last error.
12. Reclaim expired leases only through an explicit operation; do not require a real ticker or background goroutine.
13. Cancel Ready or Leased tasks deterministically, invalidate cancelled leases, and make repeated cancellation of Cancelled harmless.
14. Require handlers to make side effects idempotent by task ID; never describe the queue as exactly once.
15. Protect the state machine with a mutex and never call a handler or advance a clock while the lock is held.
16. Return copied inspection values so callers cannot mutate queue payload or lease state.

## 9. Inputs and Outputs

### Interface Contract

- Inputs are validated task definitions, worker dequeue requests, lease acknowledgements, negative acknowledgements with bounded errors, cancellation requests, explicit reclaim requests, fake-clock advancement, deterministic token values, and deterministic handler outcomes.
- Outputs are accepted task identities, leased task copies with authority metadata, empty-ready results, typed validation or stale-lease outcomes, terminal task inspections, deterministic dead-letter contents, and an idempotent side-effect ledger used only by tests.
- No operation performs network or disk I/O.

## 10. Rules and Edge Cases

- The injected clock is treated as non-decreasing; a fake clock attempting to move backward is invalid.
- Ready-time and lease-expiration arithmetic must detect overflow.
- Equal ready times use enqueue sequence, never map iteration order.
- Expiration is exclusive: acknowledgement at the exact expiration is stale.
- Reclaim observes time once, so all decisions in one call use the same instant.
- An expired lease remains unavailable to dequeue until reclaim performs its transition.
- A failed token-source call leaves the task Ready and does not consume an attempt.
- An invalid negative-acknowledgement error leaves the lease unchanged.
- Attempt count never exceeds the policy maximum.
- DeadLettered receives one entry per task.
- A handler error and a lease expiry both consume the already-issued attempt.
- A handler success followed by a lost acknowledgement may lead to duplicate delivery.
- Cancellation does not roll back effects.
- No terminal task can return to Ready.

## 11. Project Constraints

- Single process, one in-memory queue, exactly two logical workers, and at most 10,000 lifetime task identities.
- No Redis, database, message broker, RPC, network transport, disk persistence, process failover, lease renewal, priorities beyond ready-time ordering, delayed background scheduler, real ticker, or unbounded retry.
- No promise of fairness beyond the documented ordering.
- No promise of exactly-once delivery or exactly-once side effects.
- Production systems also require authentication, authorization, observability, durable storage, backpressure, admission control, clock-skew handling, and operational recovery; those concerns are outside this project.

## 12. Design Questions Before Coding

- Why is attempt count incremented when a lease is issued?
- Why does authority include worker, random token, and monotonic version?
- Why is acknowledgement at the expiration boundary stale?
- Why must reclaim be explicit and use one clock observation?
- Why does retry time start from the failure observation rather than from the original ready time?
- Why do retries retain enqueue sequence?
- Why is only an exact duplicate of the successful terminal acknowledgement harmless?
- Why can queue-level at-least-once delivery not guarantee exactly-once external effects?
- Why must handler work occur outside the mutex?
- What state must remain queryable to reject duplicate task IDs and recognize a lost acknowledgement response?

## 13. Implementation Milestones

1. Establish validated bounds, the injected clock and token source, the two worker identities, and the retry-policy contract.
2. Establish task records, terminal records, immutable enqueue sequence, copied payload ownership, and the five-state transition table.
3. Establish deterministic Ready selection by absolute ready time and enqueue sequence.
4. Establish the atomic lease transition with attempt count, worker identity, random token, version, and exclusive expiration.
5. Establish token-checked acknowledgement, the narrow duplicate-success rule, and stale-authority outcomes.
6. Establish negative acknowledgement, deterministic capped exponential backoff, and exactly-once dead-letter insertion.
7. Establish explicit expiration reclamation and deterministic processing order for multiple expired leases.
8. Establish cancellation and terminal inspection without exposing mutable internal values.
9. Establish an idempotent fake handler ledger keyed by task ID and controlled crash points around side effects and acknowledgement.
10. Complete deterministic fake-clock, two-worker, cancellation, dead-letter, and race-detector verification without sleeps.

## 14. Verification Cases the Learner Must Write

### Required Cases

- A task is leased once, handled successfully, acknowledged, and remains Completed.
- Ready-time equality is eligible; future tasks are not eligible.
- Tasks are selected by ready time and then immutable enqueue sequence, including after retries.
- The first failure schedules exactly one base delay; later failures double deterministically and saturate at the cap.
- Attempt count increments on lease issuance and never exceeds the maximum.
- A worker crash leaves the task Leased until fake time reaches expiration and explicit reclaim runs.
- A lease is valid immediately before expiration and stale at exact expiration.
- Reclaim requeues an expired task below the attempt cap and dead-letters an expired final attempt.
- A stale token, wrong worker, wrong version, and superseded lease are each rejected without mutation.
- Repeating the exact successful acknowledgement is harmless; a different terminal acknowledgement is rejected.
- A handler applies a side effect, loses its acknowledgement, receives a duplicate delivery, and does not apply the side effect twice.
- Two workers racing for one task never hold simultaneous valid leases.
- Multiple ready tasks and multiple expired leases have deterministic order independent of map iteration.
- A final negative acknowledgement moves a task to DeadLettered once with its last error.
- Cancelling Ready removes eligibility; cancelling Leased invalidates its authority; repeated cancellation is harmless.
- Duplicate IDs, oversize payloads, invalid worker IDs, invalid retry policy, invalid errors, token-source failure, and time overflow make no partial mutation.
- Returned payloads and inspection values cannot mutate queue state.
- All concurrent tests pass under the race detector and use barriers or channels rather than sleeps.
- No test starts a ticker, contacts a network, writes a file, or depends on wall-clock timing.

## 15. Common Mistakes to Watch For

- Removing a task from Ready before token generation can succeed;
- Incrementing attempts on failure rather than lease issuance;
- Accepting acknowledgement at exact expiration;
- Checking only task ID and not worker, token, and version;
- Allowing an old token after requeue;
- Making every repeated acknowledgement fail even when the successful response alone was lost;
- Applying backoff from wall-clock sleep;
- Adding jitter to deterministic tests;
- Assigning retry ordering from map iteration;
- Inserting the same task into the dead-letter queue twice;
- Holding the mutex during handler work;
- Returning internal payload slices;
- Cancelling a lease without invalidating it;
- Claiming handler idempotency makes queue delivery exactly once;
- Using a background ticker;
- Or treating in-memory state as durable.

## 16. Topics and References for Study

- Study Go documentation for mutexes, the race detector, contexts, time values, cryptographic randomness, error classification, and heap-based priority queues.
- Study at-least-once delivery, visibility timeouts, lease fencing tokens, idempotency keys, dead-letter queues, retry storms, exponential backoff, and the transactional outbox pattern.
- Compare the educational model with documented semantics of mature queue systems, but do not import their production guarantees into this simulation.
- Review Project 034 (worker_pool_basic) for bounded worker-pool ownership, Project 041 (context_timeout_example) for context cancellation and deadlines, and Project 065 (redis_caching_layer) for failure-aware external-cache and dependency boundaries, explicit partial outcomes, and Redis vocabulary, and Project 085 (packet_sniffer_basics) for bounded input and deterministic diagnostic discipline.

## 17. Self-Assessment Questions

1. Can you state every valid transition from Ready and from Leased, including the terminal states?
2. At what instant does a lease stop being valid, and why does an expired task remain unavailable to dequeue until reclaim runs?
3. Which fields together prove current authority, and which event consumes an attempt?
4. What exact delay follows each failed attempt under the deterministic policy, and why does that policy need its declared maximum validated before observation begins?
5. Why does eligible-task ordering depend on ready time and immutable enqueue sequence rather than on insertion time or map iteration?
6. Why can a successful handler run more than once, and how does this queue keep that contract honest without claiming exactly-once delivery?
7. Which duplicate acknowledgement is harmless, under what exact match, and why is every other acknowledgement against a terminal task rejected?
8. What happens when cancellation races with acknowledgement, including a stale lease, and why does repeated cancellation of Cancelled remain harmless?
9. Which invariant prevents two simultaneous valid leases for one task, and what data is lost when the process exits?
10. Which broad production broker properties are intentionally absent from this educational model, and which of those does this project therefore refuse to promise?

## 18. Definition of Completion

- [ ] Projects 034 (worker_pool_basic), 041 (context_timeout_example), and 065 (redis_caching_layer) are treated as the required prerequisites.
- [ ] Queue models exactly two logical workers and no production distribution.
- [ ] IDs, payloads, worker IDs, errors, task count, lease duration, and retry policy are bounded and validated without partial mutation.
- [ ] Payloads and returned values are copied.
- [ ] Ready ordering is deterministic by ready time and enqueue sequence.
- [ ] Dequeue atomically creates one versioned random-token lease and consumes one attempt.
- [ ] Only unexpired matching authority can change a leased task.
- [ ] Exact duplicate successful acknowledgement is harmless; all stale acknowledgements are rejected.
- [ ] Expiry and negative acknowledgement share the same capped deterministic retry and dead-letter rules.
- [ ] Explicit reclaim, cancellation, terminal retention, and dead-letter insertion follow the documented state machine.
- [ ] Handler effects are idempotent by task ID, while queue semantics remain explicitly at least once.
- [ ] Fake-clock tests cover success, retry timing, crash expiry, stale tokens, duplicate handling, two-worker exclusion, dead-lettering, cancellation, and ordering.
- [ ] Concurrent tests pass under the race detector with no sleeps or real ticker.
- [ ] Guide and project make no persistence, production broker, or exactly-once claim.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions

- Add bounded lease renewal that requires the current worker, token, and version and never revives an expired lease.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 034 — Worker Pool Basic](../../03-concurrency/034_worker_pool_basic/README.md#20-prerequisite-based-documentation-guide), [Project 041 — Context Timeout Example](../../03-concurrency/041_context_timeout_example/README.md#20-prerequisite-based-documentation-guide), [Project 065 — Redis Caching Layer](../../05-databases/065_redis_caching_layer/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`container/heap`](https://pkg.go.dev/container/heap).
- **Standards and concept references:** [Amazon SQS visibility-timeout documentation](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-visibility-timeout.html).

### Project-specific learning focus

- **Learn now:** at-least-once delivery, leases and fencing tokens, retries and backoff, idempotency, dead-letter queues, deterministic clocks, queue invariants, and honest simulation limits.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
