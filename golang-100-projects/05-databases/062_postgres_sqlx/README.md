# Project 062 — PostgreSQL sqlx

## 1. Project Name and Number

- Project 062, postgres_sqlx.
- This README is a learning guide only.
- You will create every Go source file, the SQL schema, the Compose definitions, and every test file yourself in `05-databases/062_postgres_sqlx/`.
- The guide does not provide implementation code.

## 2. Project Idea

Port the Project 061 User repository to PostgreSQL using `github.com/jmoiron/sqlx` at `v1.4.0` and `github.com/jackc/pgx/v5` at `v5.10.0`, using pgx's stdlib adapter through `database/sql`. Pin both versions exactly. Use explicit column lists and a single coherent named-parameter path: expand named parameters into positional arguments, then call `db.Rebind` exactly once to produce PostgreSQL placeholders, then send that query and arguments through the executor. Do not also call a helper that automatically rebinds an already-rebound query. Some sqlx helpers rebind internally; this project deliberately uses the explicit path so the executor boundary is observable and testable. Keep unit tests Docker-free by writing through a narrow repository-facing executor boundary. Tag integration tests behind both a build tag and an explicit safety guard, and never connect to a developer or shared database.

## 3. Why This Project Now?

- This follows Project 061 (sqlite_crud) and transfers a tested repository contract to a server database.
- Context propagation follows Project 041.
- No other project is formally required.

## 4. Prerequisites

- Projects 061 and 041 are required.
- No other project is formally required.
- The regular unit gate is `go test ./...`, which must pass with no Docker, no PostgreSQL, no Redis, no network, and no environment variables.

## 5. What You Must Know Before Starting

- You should know the exact Project 061 User contract (fields, normalization, ordering, clock behavior, typed outcomes).
- You should be comfortable with `database/sql` contexts and rows, basic PostgreSQL constraints and SQLSTATE, sqlx named arguments and `Rebind`, pool configuration including `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`, and `SetConnMaxIdleTime`, Go build tags, and safe isolation between tests using transactions or unique data.

## 6. Explanation of New Concepts

### Concepts

- sqlx adds three practical capabilities on top of `database/sql`: struct scanning by name, named parameter expansion in query text, and a small `Get`/`Select` helper.
- It does not change the underlying connection lifecycle or the requirement to close rows and check rows errors.

- pgx stdlib compatibility: pgx ships a `database/sql` driver and a standard `*sql.DB` adapter.
- When you register that driver and open a `*sql.DB`, every interaction still flows through `database/sql`, but pgx provides the driver implementation and its own error types.
- The pgx PostgreSQL error type is the recommended source of truth for SQLSTATE; you extract the SQLSTATE through type assertion, not by parsing message text.

- Named parameters and `Rebind`: sqlx expands named arguments (for example, `:name` or `:email`) into a rewritten query and an `[]any` slice.
- PostgreSQL uses `$1`, `$2`, ... placeholders, so after expansion you must call `db.Rebind` to translate `?` placeholders to `$N` placeholders.
- Without `Rebind`, the database will reject the query.

- Struct tags and explicit columns: every column you select has a matching struct tag (for example, `db:"id"` or `db:"created_at"`).
- The query lists every column explicitly. `SELECT *` is forbidden.

- Pool configuration: PostgreSQL tolerates multiple connections, so `SetMaxOpenConns` and `SetMaxIdleConns` may be larger than one.
- Lifetimes and idle times recycle stale connections.

- SQLSTATE: PostgreSQL reports error class codes such as `23505` (unique violation).
- You read the code from the pgx PostgreSQL error type using type assertion, then map to a typed outcome.

- Unit boundary versus real integration: handwritten fakes at the executor boundary can assert that the right query was sent, with the right arguments, to the right scan target, and that scan failures propagate.
- They cannot validate PostgreSQL syntax, real constraint semantics, or pgx's behavior.
- Integration covers those.

- DSN safety: integration must never silently connect to anything.
- You parse the DSN, require the database name to end in `_test`, and require a separate explicit opt-in guard before touching data.

## 7. Learning Objective

- Build a PostgreSQL repository with the same User semantics as Project 061 while clearly separating Docker-free unit confidence from opt-in integration confidence.
- Make every typed outcome, every pool decision, and every placeholder choice an explicit, tested decision.

## 8. Functional Requirements

1. The User representation matches Project 061 exactly: positive 64-bit ID, trimmed non-empty name, normalized email, UTC RFC3339Nano created-at and updated-at.
2. Dependencies are pinned exactly: `github.com/jmoiron/sqlx` `v1.4.0` and `github.com/jackc/pgx/v5` `v5.10.0`. pgx is consumed via its stdlib adapter through `database/sql`. `lib/pq` is not used.
3. The PostgreSQL schema enforces the same semantic checks as the 061 SQLite schema for positive ID, non-empty trimmed name, email equal to its trimmed lowercase form, unique email, and non-empty timestamps.
4. Every query uses an explicit column list. `SELECT *` is never used.
5. Named arguments are used only where they genuinely improve clarity, never by default. The chosen path expands named parameters into positional arguments, calls `db.Rebind` exactly once to translate `?` placeholders to `$N` placeholders for PostgreSQL, and sends that query and arguments through the executor. No helper that auto-rebinds is layered on top of the already-rebound query.
6. Struct tags map columns to fields unambiguously. Tags must not collide.
7. Every repository call accepts a `context.Context`. Startup uses a finite context owned by the caller and calls `Ping`. Shutdown calls `Close`.
8. Pool configuration is finite and explicit. The chosen policy is `SetMaxOpenConns(8)`, `SetMaxIdleConns(4)`, `SetConnMaxLifetime(30 minutes)`, and `SetConnMaxIdleTime(5 minutes)`. These four values live in a validated application configuration object that the constructor applies once. Only `DBStats.MaxOpenConnections` directly reflects a configured value. The standard `database/sql` API does not expose getters for max idle, lifetime, or idle time, so the configuration object itself is the source of truth and is asserted at construction. Pool exhaustion does not produce a distinct error: when all 8 connections are busy, another operation waits until a connection is free or its context is cancelled or its deadline expires; the wait ends with the context error. `DBStats.WaitCount` records the number of times callers waited.
9. Duplicate is mapped from SQLSTATE `23505` obtained from the pgx PostgreSQL error type. Not-found is mapped from the absence of rows. Invalid input is mapped before any database call. No string matching is used.
10. The repository does not require a transaction feature in production code.
11. A narrow repository-facing SQL executor boundary exposes the operations the repository actually uses. The executor fake can prove only the final `$N` query and arguments it received. Exactly-once `Rebind` is observed one of two ways: (a) a tiny injected or recording query-preparation boundary that counts the named-expansion and `Rebind` preparation call, or (b) a bounded static review of the single preparation helper. The executor fake alone cannot observe a `Rebind` call because `Rebind` happens before the executor is invoked. Other assertions on the recorded query may normalize whitespace and case for semantic checks but must still assert: every query explicitly names its columns; item mutations and lookups include a `WHERE`; list queries include ascending `ORDER BY id`; wildcards are absent from the column list.
12. Handwritten fakes cannot validate PostgreSQL syntax, real placeholder behavior, real constraints, or real column-to-struct scan compatibility. Those are covered by tagged integration tests.
13. The `integration` build tag controls whether integration tests are compiled at all. Once compiled with that tag, activation requires a non-empty test DSN and the exact opt-in guard `I_UNDERSTAND_TEST_DB_WILL_BE_MODIFIED=yes`; there is no separate "address" input. If both the DSN and guard are absent, the integration test skips clearly. If either is supplied, then a missing or unparseable DSN, a database name that does not end in `_test`, a missing guard, or a guard whose value is not the exact expected string fails closed before any database connection, schema creation, or cleanup. Credentials must never be printed.
14. Schema setup is explicit and not driven by a third-party migration tool in this project. Ordinary integration cases isolate DML in transactions that roll back at the end. The concurrent-create case uses unique data plus explicit bounded cleanup because one transaction cannot honestly demonstrate pool concurrency.
15. The repository never defaults to a developer or shared database.

## 9. Inputs and Outputs

### Interface Contract

- Create: context, name, email.
- Output: user with PostgreSQL-generated ID and the two UTC timestamps from a single clock read, or typed invalid input or duplicate.
- Get: context, ID.
- Output: user or typed not-found.
- List: context.
- Output: non-nil slice in ascending ID order.
- Update: context, ID, new name, new email.
- Output: updated user with ID and created-at preserved and updated-at from one clock read, or typed not-found, invalid input, or duplicate.
- Delete: context, ID.
- Output: typed not-found or success.

## 10. Rules and Edge Cases

- Reject nonpositive IDs on Get/Update/Delete.
- Reject blank or invalid names and emails as in Project 061.
- Map SQLSTATE `23505` from the pgx PostgreSQL error type, not from message text.
- Treat no rows as typed not-found.
- Cancelled context returns the context error before any repository work.
- There is no distinct pool-exhaustion error from `database/sql`: when all 8 connections are busy, another operation waits until a connection is available or its context is cancelled or its deadline expires; the wait ends with the context error.
- With integration tests compiled, skip only when both runtime activation values are absent; any partially supplied or invalid activation fails closed.

## 11. Project Constraints

- Pin exactly sqlx `v1.4.0` and pgx `v5.10.0` as direct dependencies.
- Never use `SELECT *`.
- No mocking dependency.
- No required transaction feature.
- No ORM.
- No HTTP layer.
- Compose material is conceptual only; this README contains no commands or YAML.
- Integration is opt-in and never run by the regular unit gate.

## 12. Design Questions Before Coding

- What is the minimum executor surface that still lets unit tests cover every repository operation?
- How will SQLSTATE be extracted via type assertion without message matching?
- Where is `Rebind` called for every named-parameter execution?
- How are DSN credentials kept out of test logs?
- How will rollback isolation prove that ordinary tests do not leak data?
- How will the concurrent-create case avoid pretending one transaction is pool concurrency?

## 13. Implementation Milestones

1. Define the executor boundary, the typed outcomes, and the user struct that mirrors Project 061.
2. Wire sqlx on top of pgx's stdlib adapter and configure the pool with the pinned numbers.
3. Implement the schema bootstrap with explicit columns and required constraints.
4. Implement Create with explicit column list and named/rebound parameters.
5. Implement Get, List (ascending ID, full scan, rows close and rows error), Update, and Delete.
6. Implement driver-error classification using the pgx PostgreSQL error type for SQLSTATE `23505`.
7. Write handwritten fakes for the executor boundary and unit tests for every typed outcome and bound argument.
8. Add tagged integration tests with the documented safety guards; isolate ordinary tests in transactions and use unique data plus bounded cleanup for concurrent creates.
9. Verify the unit gate and the race detector.

## 14. Verification Cases the Learner Must Write

### Required Cases

Unit tests:
- Successful Create issues the documented query text and bound arguments and returns a user with a positive ID and UTC timestamps equal to the single clock read.
- Create with blank or whitespace-only name returns typed invalid input.
- Create with invalid email (empty after trim, no `@`, multiple `@`, empty local, empty domain, internal whitespace) returns typed invalid input.
- Create normalizes mixed-case email and trims whitespace before binding.
- Create binds an explicit column list; `SELECT *` is rejected by a test that fails if the query text contains `* from`.
- Get by existing ID issues a parameterized query with `WHERE id = $1` and returns the user.
- Get by unknown ID returns typed not-found without calling the executor more than necessary.
- List issues `ORDER BY id ASC` and the explicit column list and returns a non-nil slice, possibly empty.
- Update preserves ID and created-at and reads the clock once for updated-at.
- Update with colliding email returns typed duplicate; the fake returns the constructed pgx PostgreSQL error with SQLSTATE `23505` so the mapping is exercised.
- Update with missing ID returns typed not-found and zero rows affected.
- Delete of an existing ID returns success.
- Delete of a missing ID returns typed not-found.
- Context cancelled before Create, Get, List, Update, and Delete returns the context error without useful database work.
- Scan-error propagation through the actual fake boundary: when the fake returns a result whose column types or arity do not match the destination, the resulting scan error is returned unchanged to the repository caller. No contrived mismatch that does not correspond to a real boundary failure is required.
- A fake that returns a `RowsAffected` of zero on Update is mapped to typed not-found.
- The query received by the fake uses `$N` placeholders.
- Exactly-once `Rebind` is proven by a tiny injected or recording query-preparation boundary that counts the preparation call, or by a bounded static review of the single preparation helper.
- The configuration object exposes the four pinned values and tests assert the configuration is applied at construction.

Tagged integration tests:
- Real Create/Get/List/Update/Delete on a fresh disposable PostgreSQL instance with timestamps in UTC RFC3339Nano.
- Duplicate on Create and on Update returns typed conflict mapped from SQLSTATE `23505`.
- Missing Get/Update/Delete returns typed not-found.
- Context cancelled mid-call returns the context error.
- Real Create and Update with input that normalizes to the same email return typed duplicate on the second call; the stored email in both rows equals its trimmed lowercase form.
- Real List on an empty database returns a non-nil empty slice.
- Ordinary cases isolate DML in transactions that roll back; concurrent creates use unique data and bounded explicit cleanup of only those rows.
- Pool-backed concurrent creates succeed and respect the configured max open connections without exhausting the pool. `DBStats.MaxOpenConnections` matches the configured value.
- Deterministic `WaitCount` test without sleeps: deliberately occupy all 8 pooled connections with synchronized in-flight operations (for example, long-context queries on a held connection), start an additional context-bounded waiter that must block because the pool is full, observe `DBStats.WaitCount` incrementing, then release the held connections and prove the waiter completes. Launching more fast operations than the pool size is not a deterministic way to observe `WaitCount` and is not relied upon.
- Exact lifetime recycling is not observed by sleep.
- Safety guard distinction: without the `integration` build tag the integration file is not compiled. With the tag present, an absent DSN and absent guard together cause a clear skip. If either runtime value is supplied, a missing or unparseable DSN, a database name not ending `_test`, a missing guard, or a guard value not exactly `I_UNDERSTAND_TEST_DB_WILL_BE_MODIFIED=yes` fails closed before any database connection, schema creation, or cleanup.
- `Ping` succeeds during startup with a finite context.
- `Close` is called in a defer.

## 15. Common Mistakes to Watch For

- Using `lib/pq`.
- Forgetting `Rebind` and seeing `$1` interpolation errors.
- Using `SELECT *`.
- Matching error messages by string.
- Forgetting `rows.Close` or `rows.Err`.
- Treating `RowsAffected` of zero on Update as success.
- Reusing a shared database name across developers.
- Running unit tests against a real PostgreSQL.
- Pretending one transaction demonstrates pool concurrency.
- Printing the DSN or credentials in test logs.
- Catching `context.Canceled` and remapping it as a generic internal error.
- Using `db.Unsafe()` to bypass sqlx safety without a documented reason.

## 16. Topics and References for Study

- The sqlx documentation on `Get`, `Select`, `NamedExec`, `NamedQuery`, `In`, and `Rebind`.
- The pgx `v5.10.0` documentation on its stdlib adapter and the PostgreSQL error type.
- PostgreSQL documentation on `SERIAL`/`BIGSERIAL` versus sequences, `CHECK` constraints, unique indexes, and SQLSTATE classes.
- Go `database/sql` pool configuration and `DBStats`.
- Go build tags.
- Safe DSN parsing and the rule that test database names must end in `_test`.

## 17. Self-Assessment Questions

1. Why is `Rebind` necessary after named expansion?
2. Why is `Rebind` called exactly once and not layered with an auto-rebind helper?
3. Why is the pgx PostgreSQL error type preferred over message matching?
4. What can a handwritten fake prove, and what can it not prove?
5. Why must the database name end in `_test`?
6. Why does the opt-in guard exist?
7. Why is a single transaction insufficient for proving pool concurrency?
8. Why is `SELECT *` rejected even when convenient?
9. What is the difference between a missing-row outcome and a `RowsAffected` zero outcome?
10. Why does pool exhaustion end with a context error rather than a distinct exhaustion error?

## 18. Definition of Completion

- [ ] `go test ./...` passes with no Docker, network, PostgreSQL, Redis, or environment variables.
- [ ] `go test -race ./...` passes.
- [ ] Exactly-once `Rebind` is proven by an injected or recording query-preparation boundary or by bounded static review of the single preparation helper.
- [ ] Deterministic `WaitCount` test holds all 8 connections and proves a context-bounded waiter waits, without sleeps.
- [ ] SQLSTATE `23505` is extracted from the pgx PostgreSQL error type, never by message text.
- [ ] Integration tests are behind the `integration` build tag and the documented DSN/guard checks.
- [ ] With the integration file compiled, tests skip clearly when both runtime activation values are absent and fail closed before any database connection, schema, or cleanup when activation is partial or unsafe.
- [ ] Ordinary integration cases isolate DML in transactions that roll back.
- [ ] The concurrent-create case uses unique data plus bounded cleanup of only its rows.
- [ ] No third-party mocking dependency is introduced.

## 19. Optional Extensions

- Add a separately documented cursor-style pagination experiment that proves stable boundary behavior.
- Add a separately tagged pool-observability integration that records `DBStats` before and after a burst, without weakening the unit gate.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 061 — SQLite CRUD](../../05-databases/061_sqlite_crud/README.md#20-prerequisite-based-documentation-guide), [Project 041 — Context Timeout Example](../../03-concurrency/041_context_timeout_example/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`github.com/jmoiron/sqlx`](https://pkg.go.dev/github.com/jmoiron/sqlx), [`github.com/jackc/pgx/v5/stdlib`](https://pkg.go.dev/github.com/jackc/pgx/v5/stdlib).
- **Standards and concept references:** [PostgreSQL documentation](https://www.postgresql.org/docs/current/).

### Project-specific learning focus

- **Learn now:** named queries, placeholder rebinding, SQLSTATE classification, constraints and sequences, stable pagination, pool statistics, safe DSNs, and guarded integration tests.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
