# Project 061 — SQLite CRUD

## 1. Project Name and Number
Project 061, sqlite_crud. This README is a learning guide only. You will create every Go source file, the SQL schema, and every test file yourself in `05-databases/061_sqlite_crud/`. The guide does not provide implementation code.

## 2. Project Idea
Implement a real User repository on top of Go's `database/sql` package against a freshly created temporary SQLite database file per test. Pin `modernc.org/sqlite` at `v1.55.0` as the only third-party dependency. Avoid the well-known `:memory:` multi-connection trap by using a real file path in every test. Treat the database as a precise component with deterministic connection limits and disciplined resource lifecycle. Tests must not require Docker, network, PostgreSQL, Redis, or environment variables.

## 3. Why This Project Now?
This follows Project 047 (an earlier data-handling project) and turns those application-side concepts into durable, constraint-backed storage. Project 060 (graceful shutdown for a web service) is related but optional review material; it is not formally required.

## 4. Prerequisites
Project 047 is required. No other project is formally required. The regular unit gate is `go test ./...`, which must pass with no Docker, no network access, no PostgreSQL, no Redis, and no environment variables.

## 5. What You Must Know Before Starting
You should be comfortable with Go structs, methods, interfaces, errors and error wrapping, the `context` package including cancellation and deadlines, basic SQL constraints, and table-driven tests. You should also understand the difference between package-level globals and per-test fixtures. You will create all source, schema, and test files yourself.

## 6. Explanation of New Concepts
The Go `database/sql` package is an interface, not a concrete driver. It owns a pool of connections to a single underlying SQLite file. The driver (`modernc.org/sqlite`) is registered once and is the only allowed third-party dependency. Even though SQLite is in-process, each `sql.DB` connection is still a connection, and certain statements are connection-scoped (for example, `PRAGMA foreign_keys = ON`, or any setup statement that you want to run before repository methods execute).

The connection pool versus the SQLite file: setting `SetMaxOpenConns` and `SetMaxIdleConns` to 1 forces all queries onto a single connection. This makes connection-scoped state deterministic and simplifies reasoning about locking. With more connections, the same file would still serialize writes through SQLite's file-level lock, but `:memory:` databases become per-connection private databases, which is the well-known trap you are deliberately avoiding.

Why on-disk files matter: a fresh temporary file per test gives every test an isolated, persisted-yet-disposable database. The file is removed when the test ends, so leakage between tests is impossible as long as you do not reuse paths.

Placeholders are mandatory. User-supplied values are sent to the driver as bound parameters, never by string concatenation. This avoids both syntax errors and SQL-injection-style mistakes.

Scanning and rows lifecycle: every query that returns rows produces a `*sql.Rows`. You must iterate, `Scan` every column you select, and then close the rows. You must also check `rows.Err()` after the loop because some failures only surface when the result set is fully consumed or closed. Closing rows releases the connection back to the pool.

Constraints as defense in depth: validating in Go is necessary but not sufficient. The schema must also enforce positive IDs, non-empty trimmed names, email already stored in normalized lowercase form, unique emails, and non-empty timestamps. If a future caller bypasses your repository, the database still rejects bad data.

Normalization: the repository normalizes email before persisting it. The database stores the normalized form. The repository does not re-normalize on read; Get and List scan and return the exact stored representation. Normalizing on read would hide corrupted or bypassed data and is forbidden.

Typed error mapping: a duplicate email is one outcome, a missing row is another, and an invalid value is a third. Each maps to its own typed error sentinel or type. You classify these from SQLite driver errors using the modernc driver error type and codes, never by matching error message substrings.

## 7. Learning Objective
Implement trustworthy create, get, list, update, and delete behavior with deterministic ascending ordering, explicit input validation, persistence that survives a close and reopen, honest context semantics, and stable typed outcomes for invalid input, duplicate, and not-found. Make resource lifecycle observable through tests, and run successfully under `go test -race`.

## 8. Functional Requirements
1. Repository state is represented by exactly five fields: a positive 64-bit integer ID, a trimmed non-empty name, a normalized email, a created-at timestamp, and an updated-at timestamp.
2. Normalization of email means trimming surrounding whitespace and lowercasing the result. Validation of email requires the normalized form to be non-empty, contain exactly one `@`, have a non-empty local part before `@`, have a non-empty domain part after `@`, and contain no whitespace characters anywhere.
3. Name validation trims the input and rejects the result if it is empty.
4. Stored timestamps are UTC and rendered as RFC3339Nano text in the database.
5. The schema enforces a positive integer primary key, a non-empty trimmed name, an email equal to its trimmed lowercase form (so the database cannot store un-normalized email values), a unique index over the email column, and non-empty timestamp strings.
6. Create accepts only name and email from the caller. It must not accept an ID, created-at, or updated-at from the caller. The repository reads the injected clock once and uses the result for both created-at and updated-at.
7. Get accepts an ID and returns the corresponding user or a typed not-found outcome.
8. List returns all users in ascending ID order, scanning every column for every row. The returned collection must be non-nil even when no users exist. The rows iterator is closed and the terminal `rows.Err()` is checked.
9. Update is a full update of name and email for a given ID. It preserves the existing ID and created-at, normalizes the new email, trims and validates the new name, and reads the injected clock once for the new updated-at. Update returns typed not-found when the ID does not exist and a typed duplicate when the new email collides with another row.
10. Delete returns typed not-found when the ID does not exist and otherwise removes the row. Subsequent Get of the same ID returns typed not-found. A nonpositive ID on Get, Update, or Delete is a typed invalid-input outcome (not not-found).
11. All repository operations accept a `context.Context` and pass it to `database/sql`. Context cancellation before any call short-circuits with the context error. Context errors must remain discoverable through wrapping, not be swallowed into a generic error.
12. Duplicate classification covers both Create and Update. Invalid input is a distinct typed outcome from duplicate and from not-found. Duplicate is detected from SQLite constraint codes or driver error types provided by the modernc driver, never by message text.
13. Database constraints must be verified independently of repository validation by attempting direct inserts that bypass the repository. Such raw attempts are passed through the same SQLite error-classification helper used by the repository: a `CHECK` constraint violation (such as blank name, empty timestamp, non-positive ID, or un-normalized email) maps to typed invalid input; a `UNIQUE` constraint violation on email maps to typed duplicate. A bypassed raw insert does not magically return a repository typed error; it returns the underlying driver error which is then mapped by the helper.
14. The schema uses an `INTEGER PRIMARY KEY` column that includes SQLite's explicit `AUTOINCREMENT` keyword plus a positive-ID check constraint. The repository relies on the database-generated ID; it does not assign IDs itself. SQLite maintains `sqlite_sequence` to ensure a successfully deleted ID is never reused, including after close and reopen. This project intentionally accepts the `sqlite_sequence` overhead in exchange for the no-reuse guarantee; this tradeoff is documented and not an accident.
15. The connection pool uses `SetMaxOpenConns(1)` and `SetMaxIdleConns(1)`. The schema is created explicitly before any repository call. Startup uses `Ping` to confirm connectivity, and every test closes the database at the end.

## 9. Inputs and Outputs
Create inputs: context, name, email. Create outputs: a fully populated user with a positive ID and the two UTC timestamps set from a single clock read, or a typed invalid-input or duplicate outcome. Get inputs: context, ID. Get outputs: a fully populated user or typed not-found. List inputs: context. List outputs: a non-nil slice in ascending ID order, possibly empty, with each row scanned in full. Update inputs: context, ID, new name, new email. Update outputs: the updated user with ID and created-at preserved and updated-at from one clock read, or typed not-found, invalid-input, or duplicate. Delete inputs: context, ID. Delete outputs: typed not-found when absent, otherwise success.

## 10. Rules and Edge Cases
Reject a nonpositive ID on Get/Update/Delete. Reject a blank, whitespace-only, or otherwise empty name after trimming. Reject an email whose normalized form is empty, contains whitespace, contains zero or more than one `@`, or has empty local or domain parts. Classify SQLite constraint codes or driver error types, not message text. Pre-cancel context returns the context error. Backward or equal injected timestamps are reported honestly and never silently advanced. Concurrent calls share one configured database, accept SQLite's serialization, and never assert timing or busy-loop.

## 11. Project Constraints
Use only the standard library plus `modernc.org/sqlite` at `v1.55.0`. Use a fresh temporary on-disk file per test; never `:memory:`. Cap `SetMaxOpenConns` and `SetMaxIdleConns` at 1. Always close the database. Always close rows. Always check `rows.Err()`. Always scan every selected column. No ORM, no HTTP layer, no migration framework, no integration tag, no Docker, no network access.

## 12. Design Questions Before Coding
Where exactly does email normalization happen for each operation? Which validations belong in Go and which in SQLite? How do you classify modernc driver errors without string matching? What is the explicit monotonic non-reuse ID policy? How will you observe rows closure and the terminal rows error in tests? How will you prove persistence by closing and reopening the file? How will you distinguish context cancellation from other errors?

## 13. Implementation Milestones
1. Define typed outcomes for invalid input, duplicate, not-found, and any internal errors. Define the user struct, the injected clock boundary, and the repository interface.
2. Define an explicit schema bootstrap that creates the table, the unique index, and the documented check constraints, returning an error if any statement fails.
3. Configure the `sql.DB` with one open and one idle connection. Implement `Ping` during setup and `Close` on shutdown.
4. Implement Create: validate name and normalized email, read the injected clock once, perform a single INSERT whose bound parameters are the trimmed name, the normalized email, the created-at timestamp, and the updated-at timestamp, then read the database-generated ID, and return the user. The order is validate, read clock once, insert, read ID. A second write is never performed. A failed constraint may consume a clock reading; only the timestamps of successfully persisted users are part of the public contract.
5. Implement Get by ID and List in ascending ID order with full column scanning, rows close, and rows error check.
6. Implement full Update for name and email while preserving ID and created-at and reading the clock once for updated-at.
7. Implement Delete with typed not-found.
8. Implement driver-error classification using the modernc driver error type or codes for unique constraint, and a separate path for not-found.
9. Add unit tests covering every outcome, including direct schema-level duplicate attempts.
10. Add lifecycle tests: `Ping`, `Close`, reopen with the same file path, rows closure and terminal error, context cancellation before every operation class, and concurrent operations under one connection without timing claims.

## 14. Verification Cases the Learner Must Write
- Successful Create returns a user with a positive ID and UTC timestamps equal to the single clock read.
- Create with a blank name returns typed invalid input.
- Create with an invalid email (empty after trim, no `@`, multiple `@`, empty local, empty domain, internal whitespace) returns typed invalid input.
- Create normalizes mixed-case email and trims surrounding whitespace.
- Get by the just-created ID returns the same user.
- Get by an unknown ID returns typed not-found.
- Get by a nonpositive ID returns typed invalid input.
- Delete of a nonpositive ID returns typed invalid input.
- List after one create returns a non-nil slice with one element.
- List after zero creates returns a non-nil empty slice.
- List after multiple creates returns users in strictly ascending ID order.
- List scans every column for every row.
- Update of an existing user changes name and email, preserves ID and created-at, and sets updated-at from a single clock read.
- Update with a nonpositive ID returns typed invalid input.
- Update that collides on normalized email returns typed duplicate.
- Delete of an existing user succeeds; subsequent Get returns typed not-found; the deleted ID is not reused by any subsequent Create, including after close and reopen, because the schema uses SQLite's explicit `AUTOINCREMENT` plus the positive-ID check.
- Delete of an unknown ID returns typed not-found.
- Direct schema insert with an un-normalized email fails at the database level and is classified as typed invalid input or duplicate depending on collision.
- Direct schema insert with a blank name fails at the database level.
- Direct schema insert with an empty timestamp fails at the database level.
- Data persists across Close and reopen of the same file path.
- Context cancelled before Create returns the context error without writing.
- Context cancelled before Get returns the context error.
- Context cancelled before List returns the context error without partial iteration.
- Context cancelled before Update returns the context error without writing.
- Context cancelled before Delete returns the context error without writing.
- Rows lifecycle is observable: behavioral evidence that the sole pooled connection is released after List (for example, a synchronized follow-up operation on the same pool, or `DBStats.InUse` returning to zero after List completes) plus bounded source review confirming the terminal `rows.Err()` is checked. A normal SQLite test cannot force a mid-iteration driver error, and the project does not invent one.
- Concurrent goroutines calling different operations share the one-connection database without panic and finish; no timing or busy-loop assertion is made.
- Two independent temp-file databases do not see each other's rows.
- `Close` is called exactly once through test cleanup (for example, `t.Cleanup` or a deferred call that runs once per test), and a separate test demonstrates persistence across `Close` and reopen by opening a fresh handle against the same file path and reading the previously written data. Use of the closed handle is not part of the contract and is not required.
- The race detector (`go test -race`) is clean across all tests.

## 15. Common Mistakes to Watch For
Using `:memory:` and assuming connections share state. Setting `SetMaxOpenConns` greater than 1 and observing non-deterministic behavior. Interpolating user values into SQL strings. Returning errors matched by `strings.Contains` against the modernc driver's message text. Forgetting `rows.Close` or `rows.Err`. Selecting fewer columns than you scan. Relying solely on Go validation when the database would otherwise accept bad data. Changing the ID or created-at during Update. Reading the injected clock twice and accidentally treating those as a timestamp range. Asserting that timestamps are strictly monotonic. Forgetting to call `Close` in a defer, or calling `Close` more than once. Using shared package-level database variables across tests. Declaring `AUTOINCREMENT` semantics that the schema does not actually declare. Re-normalizing email on read and silently masking corrupted stored data. Catching `context.Canceled` and converting it to a generic internal error.

## 16. Topics and References for Study
The Go `database/sql` package documentation, including `DB`, `Tx`, `Stmt`, `Rows`, `Row.Scan`, and the role of `Rows.Err`. The modernc.org/sqlite driver documentation for `v1.55.0`, especially the driver's exported error type and how SQLite extended codes are exposed. RFC 3339 and the `time.Time.Format` constant `time.RFC3339Nano`. SQLite documentation on `CHECK` constraints, `UNIQUE` indexes, `INTEGER PRIMARY KEY`, the documented behavior of `AUTOINCREMENT`, and the per-connection meaning of `PRAGMA foreign_keys`. Go testing patterns including `t.TempDir`, table-driven tests, and `t.Cleanup`. Go error wrapping with `fmt.Errorf` and `%w`, and the use of `errors.Is` and `errors.As`. The SQLite documentation on `sqlite_sequence`.

## 17. Self-Assessment Questions
Why must the schema enforce already-normalized email storage rather than only validating in Go? How do you distinguish duplicate from invalid input when both surface as constraint failures? What does `rows.Err()` add after the loop ends? Why must the injected clock be read once per Create and once per Update? What behavior of `:memory:` makes a one-connection pool still ambiguous? What is the honest meaning of "monotonic non-reuse" for this project's IDs, and what is the role of `sqlite_sequence` in guaranteeing it? How does `SetMaxIdleConns(1)` interact with `SetMaxOpenConns(1)` under contention? Why is message-text matching a fragile classification strategy? How would a direct schema insert prove defense in depth?

## 18. Definition of Completion
- [ ] `go test ./...` passes with no Docker, network, PostgreSQL, Redis, or environment variables.
- [ ] `go test -race ./...` passes.
- [ ] Every required repository operation has at least one positive and one negative test.
- [ ] Duplicate classification is exercised on both Create and Update.
- [ ] Direct schema inserts that bypass the repository are demonstrated to fail with typed outcomes.
- [ ] Every operation class short-circuits on a pre-cancelled context.
- [ ] Behavioral tests prove List releases the sole pooled connection, and bounded source review confirms rows are closed and terminal `rows.Err()` is checked.
- [ ] Persistence is proven by opening a fresh handle against the same file path after `Close` and reading previously written data.
- [ ] Every test creates its own temporary database file and closes it exactly once.
- [ ] No third-party dependency beyond `modernc.org/sqlite v1.55.0` is introduced.
- [ ] No ORM, HTTP layer, migration framework, or implementation code is present in this README.

## 19. Optional Extensions
Add a separately documented property test that seeds many randomized users and asserts ascending order, unique normalized emails, and ID monotonicity. Add a separately documented process-level experiment that observes `SetMaxOpenConns` effects with a counting wrapper, without weakening the unit gate.
