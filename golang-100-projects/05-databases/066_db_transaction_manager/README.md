# Project 066 — Database Transaction Manager

## 1. Project Name and Number
Project 066, db_transaction_manager. This README is a learning guide only. You will create every source, schema, Compose, and test file yourself in `05-databases/066_db_transaction_manager/`. This guide contains no implementation code, signatures, snippets, pseudocode, SQL, or solution commands.

## 2. Project Idea
Build an atomic PostgreSQL money-transfer service. A client supplies two distinct positive account IDs, a positive amount in integer cents, and a unique request ID. One transaction establishes idempotency, locks both accounts in deterministic ID order, validates existence and funds, debits one account, credits the other, records a terminal outcome, and commits.

## 3. Why This Project Now?
Project 066 follows Project 065 in the catalog and builds on its consistency concerns, moving from cache consistency around committed data to database-enforced atomicity, transaction serialization, idempotent requests, and retry boundaries. Projects 061 and 041 provide the required database and service foundations for this step.

## 4. Prerequisites
Required prerequisites: Projects 065, 061, and 041. Optional review: none. The normal unit gate must need no Docker, PostgreSQL, network, or environment variables. PostgreSQL integration is separate and opt-in.

## 5. What You Must Know Before Starting
Know contexts, typed errors, signed 64-bit integer arithmetic, transactions, isolation, row locks, unique constraints, injected clocks, and test doubles. Review PostgreSQL Serializable isolation and SQLSTATE classification through pgx error types.

## 6. Explanation of New Concepts
Atomic transfer means debit, credit, and transfer outcome become visible together or not at all. Balances and amounts use signed 64-bit integer cents; floating-point money is forbidden. Validate positive distinct account IDs and a positive amount, prevent negative balances, detect debit and credit overflow, and preserve total cents.

Idempotency is anchored by a unique request row. Its conceptual ledger shape contains request ID, exact source ID, destination ID, amount, status/result, transfer ID when successful, and one recorded-at UTC timestamp from an injected clock. Same request ID and same payload replays the original terminal result, including its stored recorded-at value, without moving money again. Same request ID with different payload is a typed conflict.

Success and deterministic business rejections are terminal and replayable. Missing source, missing destination, and insufficient funds are committed with no balance movement. A business rejection reads the injected clock exactly once and stores that UTC value as the ledger recorded-at timestamp. A successful transfer also reads the clock exactly once and uses that one UTC value for both the completed transfer's created-at timestamp and the request ledger's recorded-at timestamp. Validation failures that occur before transaction work are typed invalid input and make no database call. Internal, database, and context failures roll back and are not terminal business results. Rollback must leave no pending request row.

After locking the rows in ascending ID order, validate business rejection precedence in this exact order: missing source first, then missing destination, then insufficient funds. When both accounts are absent in the same attempt this pinning fixes the result to missing source. No retry is performed on any of these terminal outcomes regardless of which one is reached.

Use Serializable isolation. Within each attempt, create-or-observe the unique request ledger row inside the transaction and lock or serialize access to the matching row before comparing payload. A concurrent unique collision is not automatically a payload conflict. After the competing transaction resolves, inspect the stored payload and stored result, or retry on a recognized serialization abort, to decide replay versus conflict. A pending row is never exposed to callers.

The unique request row plus transaction serialization make concurrent identical requests safe: the second arrival blocks behind the first via the same unique request row, sees the committed terminal ledger on its next observation, and replays it without moving money.

Retry only the recognized aborted-transaction outcomes: typed pgx SQLSTATE serialization failure (40001) reported by a statement or by Commit, and typed pgx SQLSTATE deadlock detected (40P01) reported by an attempt. Never parse error text. Allow at most three total attempts, each with a fresh transaction, identical request ID, and identical payload. Backoff is small, bounded, context-aware, and injected so unit tests signal retries without sleeping. Validation, business rejection, conflict, and insufficient-funds outcomes are never retried.

A typed pgx 40001 reported by Commit proves the transaction did not commit and is one of the recognized retryable outcomes. A typed pgx 40P01 from an attempt is also retryable. Any other commit-time failure that is not classified as one of those recognized aborted-transaction outcomes is an unknown outcome: the server may have committed even though confirmation failed. Unknown outcomes are not auto-retried. Surface the uncertainty so the caller may safely retry the same request ID and payload; idempotency resolves a completed outcome when visible. Do not blindly retry unknown outcomes under a new request ID.

## 7. Learning Objective
Define exact transaction, idempotency, retry, arithmetic, and failure contracts so atomic transfers remain safe under replay, concurrency, rollback, cancellation, serialization failures, deadlocks, and uncertain commits.

## 8. Functional Requirements
1. Use `github.com/jackc/pgx/v5` exactly at `v5.10.0` through its `stdlib` adapter with `database/sql`.
2. Inputs are context, unique non-empty client request ID, positive source account ID, positive destination account ID distinct from source, and positive signed 64-bit amount in cents.
3. Invalid input is rejected before beginning a transaction.
4. Use PostgreSQL Serializable isolation for every attempt and at most three total attempts.
5. Every retry creates a fresh transaction and preserves the exact request ID and payload.
6. Establish idempotency before balance mutation. A matching terminal row returns its stored result. A mismatched payload returns typed conflict.
7. Lock source and destination rows in ascending numeric ID order regardless of transfer direction.
8. Missing account and insufficient funds are deterministic terminal rejections, committed with exact payload, result, and a recorded-at UTC timestamp from one clock read, with no balance movement.
9. A successful transfer debits and credits exactly once, stores a transfer ID, records completed status/result, uses one UTC clock read for both transfer created-at and ledger recorded-at, and conserves total cents.
10. Check signed 64-bit underflow and overflow before mutation. No account may become negative.
11. Internal, driver, database, and context failures roll back and create no terminal business record. No pending row may survive rollback.
12. Retry only the recognized aborted-transaction outcomes: typed pgx SQLSTATE serialization failure (40001) reported by a statement or by Commit, and typed pgx SQLSTATE deadlock detected (40P01) reported by an attempt. Do not parse messages. Do not retry validation, conflict, missing-account, or insufficient-funds outcomes.
13. Retry backoff is bounded and context-aware. Tests inject the retry signal and perform no real sleep.
14. Any pre-commit failure requires rollback. Preserve both the primary failure and rollback failure contextually when both occur.
15. A typed pgx 40001 reported by Commit is a recognized retryable aborted-transaction outcome. A typed pgx 40P01 from an attempt is also retryable. Any other commit-time failure is unknown outcome and is not auto-retried; the caller may safely retry the same request ID and payload.
16. Concurrent identical requests produce one balance movement and equivalent replayed results. Same ID with another payload always conflicts after the competing transaction resolves and the stored payload is inspected.
17. Unit tests fake transaction boundaries and retry classification without PostgreSQL or Docker.
18. Integration tests are isolated behind the `integration` build tag and use disposable PostgreSQL scope only.
19. Schema constraints serve as defense in depth for positive account IDs, nonnegative signed 64-bit balances, nonempty unique request IDs, positive amount, distinct accounts, valid terminal status or result values, and exact stored payload. The README does not provide SQL.

## 9. Inputs and Outputs
Input consists of context, client request ID, source account ID, destination account ID, and amount in cents. Outcomes are completed transfer with transfer ID and resulting stored result, replay of that same result, typed invalid input, typed request conflict, typed missing account, typed insufficient funds, context failure, internal/database failure, or unknown commit outcome. Terminal outcomes include exact payload and the stored recorded-at UTC timestamp; successful transfer data also includes its created-at timestamp from the same clock read.

Example behavior: a transfer of 250 cents from account 10 with 1,000 cents to account 20 with 300 cents completes with balances 750 and 550. Repeating the same request and payload returns the original transfer result with balances unchanged. Reusing that request ID for 251 cents conflicts.

## 10. Rules and Edge Cases
Never use floats. Reject zero or negative IDs, equal account IDs, empty request IDs, and nonpositive amounts. Detect source subtraction and destination addition overflow before mutation. Lock by sorted ID, not source-first. Keep request payload unchanged across attempts. Commit deterministic business rejection; roll back infrastructure failure. After locking both rows, evaluate business rejections in fixed order: missing source, missing destination, insufficient funds. A typed pgx 40001 from a statement or from Commit is retryable; a typed pgx 40P01 from an attempt is retryable. Any unclassified commit failure is unknown outcome and is not auto-retried. Never classify retryability from message text. A pending row is never exposed to callers.

## 11. Project Constraints
README only supplies behavior contracts. No ORM, distributed transaction, external queue, automatic write retry under a new request ID, or shared integration database. Normal tests require no Docker. Integration files use the `integration` build tag. No test logs credentials. No cleanup may target data outside a uniquely created disposable test scope.

## 12. Design Questions Before Coding
Where is the exact payload persisted before account mutation? How is a concurrent request distinguished as replay versus conflict when the unique request row collides? How are account locks ordered? Which failures become committed terminal outcomes and in what precedence? How are overflow and conservation checked? How is a typed pgx 40001 from a statement distinguished from a typed pgx 40001 from Commit, and how is each distinguished from an unclassified commit failure? How does an unknown commit differ from rollback? How can retry tests avoid sleeping?

## 13. Implementation Milestones
1. Define account, transfer request, ledger result, typed outcome, UTC clock, transaction boundary, retry-classifier, and schema-constraint contracts.
2. Add strict input and signed-integer arithmetic validation.
3. Build one-attempt orchestration with Serializable isolation, create-or-observe of the unique request row, locked payload comparison, and ascending-ID row locks.
4. Add replayable success, missing-source, missing-destination, and insufficient-funds terminal outcomes in fixed precedence. Each rejection uses one clock read for ledger recorded-at; success uses one clock read for both transfer created-at and ledger recorded-at.
5. Add rollback discipline and ensure pending state cannot survive failed attempts. Never expose a pending row.
6. Add pgx typed SQLSTATE classification covering 40001 from statement, 40001 from Commit, and 40P01 from attempt; three-attempt boundary; fresh transaction per attempt; identical request ID and payload; and injected context-aware backoff.
7. Add unclassified commit-time failure as unknown outcome, surface it without auto-retry, and let the caller safely retry the same request ID and payload.
8. Add schema constraints as defense in depth for the documented account, request, amount, status or result, and payload invariants.
9. Complete unit tests, then separately add guarded tagged PostgreSQL integration tests.

## 14. Verification Cases the Learner Must Write
Unit tests without Docker:
- Reject every invalid request before transaction begin.
- Transfer exact integer cents, conserve total, and prevent negative balances.
- Detect debit and credit overflow without mutation.
- Lock account rows in ascending ID order for both transfer directions.
- Commit success with exact payload, transfer ID, result, and one injected UTC clock value used for both transfer created-at and ledger recorded-at.
- Commit and replay missing-source, missing-destination, and insufficient-funds outcomes in fixed precedence, each with exactly one injected UTC clock read for ledger recorded-at and no balance movement.
- Replay same request and payload without another debit or credit.
- Return conflict for same request ID with different source, destination, or amount when the competing transaction is observed as already resolved.
- A unique-index collision in flight is not returned as conflict; the second arrival observes the resolved terminal ledger or retries on a typed serialization abort.
- Roll back injected failures after request establishment, after locks, after debit, after credit, and before commit; no pending or terminal row remains.
- Roll back context cancellation and preserve context classification.
- Retry typed pgx 40001 reported by a statement, retry typed pgx 40001 reported by Commit, and retry typed pgx 40P01 from an attempt; do not retry validation, conflict, business rejection, or insufficient funds; fresh transaction per attempt; no real sleep.
- Stop after three total attempts and preserve request ID and payload across all attempts.
- Surface an unclassified commit-time failure as unknown outcome without automatic retry; allow caller to retry the same request ID and payload.
- Concurrent fake orchestration proves one terminal result and equivalent replay semantics where the boundary supports synchronization.

Opt-in PostgreSQL integration tests:
- Integration files are excluded from the normal gate by the `integration` build tag.
- With that tag, both runtime activation values absent produce a clear skip. The values are a PostgreSQL connection setting and an explicit destructive-test guard. Partial activation, malformed connection settings, missing or wrong guard, or an activation that cannot prove a unique disposable scope fails closed before connecting or mutating.
- Generate a collision-resistant unique database or schema name, validate it before any connection, connect, create or use only that scope, and drop only it through bounded cleanup. Never print credentials, drop shared databases, truncate shared tables, or delete unowned data.
- Verify exact balances and conservation after success.
- Verify insufficient-funds and missing-account terminal replay without movement.
- Inject a mid-operation database error and prove full rollback with no pending request.
- Run concurrent opposite transfers and prove completion without deadlock leakage.
- Run concurrent identical requests, exact replay, and payload conflict, distinguishing observed-conflict from in-flight pending.
- Force or reliably induce a typed pgx 40001 from a statement, a typed pgx 40001 from Commit, and a typed pgx 40P01 from an attempt; verify bounded attempts and identical payload across all attempts.
- Verify that an unclassified commit-time failure is surfaced as unknown outcome without auto-retry.
- Verify context cancellation and resource cleanup.

## 15. Common Mistakes to Watch For
Using floating point, locking source first, inserting an uncommitted pending row outside the transaction, exposing a pending row to a caller, recording infrastructure errors as terminal business results, retrying every error, parsing SQLSTATE from text, reusing a transaction after failure, exceeding three attempts, changing request IDs, ignoring overflow, treating every commit failure as unknown outcome or refusing to retry typed 40001 from Commit, auto-retrying unclassified commit failures, applying the wrong business-rejection precedence, or cleaning shared integration data.

## 16. Topics and References for Study
Study PostgreSQL documentation for Serializable isolation, explicit row locking, transaction rollback, unique constraints, SQLSTATE serialization failure, and deadlock detection. Study Go `database/sql` transaction lifecycle, context cancellation, signed integer limits, and error wrapping. Study `github.com/jackc/pgx/v5` `v5.10.0`, its `stdlib` adapter, and exported PostgreSQL error type. Study idempotency-key and transactional outbox distinctions; this project implements only transactional request-result idempotency.

## 17. Self-Assessment Questions
Why are missing account and insufficient funds committed while driver failure is rolled back, and in what fixed precedence? Why must both accounts be locked in ID order? Why is a unique request row enough to serialize identical concurrent requests? Why is a concurrent unique collision not auto-conflict, and what is checked after the competing transaction resolves? Why can a commit error be uncertain, and why is a typed pgx 40001 reported by Commit still recognized as aborted? Why must safe caller retry reuse the same ID and payload? Which aborted-transaction outcomes are retryable, and how are they detected through the pgx error type? How is total-cent conservation proven? What schema constraints enforce the documented invariants as defense in depth?

## 18. Definition of Completion
- [ ] Exactly defined transfer, ledger, terminal outcome, business-rejection precedence, and clock contracts are implemented.
- [ ] Serializable isolation, deterministic locks, three total attempts, and fresh transaction per attempt are tested.
- [ ] Same-request replay and different-payload conflict are tested under concurrency and distinguished from in-flight unique collisions.
- [ ] Success and deterministic business rejection in fixed precedence are replayable; infrastructure failures and rolled-back attempts leave no terminal or pending row.
- [ ] Signed 64-bit overflow, no-negative-balance, and conservation checks pass.
- [ ] Typed pgx 40001 from a statement, typed pgx 40001 from Commit, and typed pgx 40P01 from an attempt are retried; unclassified commit-time failures surface as unknown outcome and are not auto-retried.
- [ ] Schema constraints enforce the documented invariants as defense in depth.
- [ ] Unit tests pass with no Docker, PostgreSQL, network, or environment variables, including the race detector.
- [ ] Tagged integration activation skips only when both values are absent and otherwise fails closed before unsafe access.
- [ ] Integration cleanup is bounded to unique disposable scope and credentials are never printed.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, SQL, Compose content, or solution commands.

## 19. Optional Extensions
Add a separately specified account statement read model derived from completed transfers. Add metrics for attempt counts and terminal outcome categories without changing transaction semantics.
