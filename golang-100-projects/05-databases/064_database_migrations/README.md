# Project 064 — Database Migrations

## 1. Project Name and Number
Project 064, database_migrations. This README is a learning guide only. You will create every Go source file, every embedded migration pair, and every test file yourself in `05-databases/064_database_migrations/`. The guide does not provide implementation code.

## 2. Project Idea
Build a small, single-process migration runner on top of `database/sql` for ordered, embedded migration pairs (Up and Down). Each pair has a unique monotonically increasing positive version, a name, and a SHA-256 checksum. A schema-migrations metadata table records what has been applied, when, and with which checksum. SQLite temp files are the deterministic test target.

## 3. Why This Project Now?
This follows Projects 063 (gorm_orm_basics) and 061 (sqlite_crud) and replaces the auto-bootstrap habit of 063 with disciplined, transactional migrations. Context handling follows Project 041. No other project is formally required.

## 4. Prerequisites
Projects 063, 061, and 041 are required. No other project is formally required. The regular unit gate is `go test ./...`, which must pass with no Docker, no PostgreSQL, no Redis, no network, and no environment variables.

## 5. What You Must Know Before Starting
You should be comfortable with `database/sql` transactions and how they interact with the SQLite driver, the documented SQLite behavior for transactional DDL, Go's `embed` package for migration files, and integrity checks via SHA-256.

## 6. Explanation of New Concepts
Embedded source discovery: the migration source is a directory tree embedded at compile time via `embed.FS`. Each migration is a pair of files. The runner reads the directory, validates the names, orders them by version, and exposes the bytes to the rest of the system.

Pair validation: every `.up.sql` has a matching `.down.sql` with the same version and name. Mismatched pairs are rejected before any database side effect.

Metadata ledger: the schema-migrations table records, for each applied version, the version number, the name, the checksum of the pair, and the time it was applied in UTC RFC3339Nano. The ledger is the source of truth for what has been applied.

Checksum and drift: a checksum hashes a precise framing of (version, name, Up bytes, Down bytes). If Up or Down bytes change after a migration is applied, the stored checksum no longer matches the computed checksum and the system reports drift. Drift is an error, never a silent overwrite.

Per-migration transaction boundaries: each pending Up runs in its own transaction. The migration SQL and the metadata insert share that single transaction. A failure rolls back the migration's data and its version row.

Up: from the current state and the source, validate the entire source and the metadata, then apply pending versions in ascending order. At the latest applied version, Up is a no-op and does not rewrite timestamps.

Down-one and Down-all: Down-one reverts only the latest applied known version. Down-all repeats Down-one in reverse order until either no applied versions remain or an error occurs. Down-all on empty is a no-op; Down-one on empty is a typed no-migrations outcome.

SQLite-specific DDL guarantees: SQLite supports transactional DDL for most statements; this project pins its guarantees to SQLite only. Other databases have different rules and are out of scope.

Missing cross-process lock: the runner is a single-process component. It does not claim safety under concurrent runners. Two runners racing on the same database can corrupt state, and this README must not promise otherwise. Within a single process, one `Runner` instance owns a mutex that serializes Up, Down-one, and Down-all operations; concurrent calls on that same instance are race-free and cannot double-apply. Two different `Runner` instances or two processes targeting the same database are outside the guarantee because there is no database-side or advisory lock.

## 7. Learning Objective
Implement a transparent migration runner whose behavior is verifiable from the source, the metadata ledger, and the schema and data of the test database. Make every failure mode honest and observable. State the single-process limitation explicitly.

## 8. Functional Requirements
1. Standard library plus the already introduced `modernc.org/sqlite` `v1.55.0`. No third-party migration framework.
2. The runner core accepts an existing `*sql.DB` and an embedded migration source. It does not start a database by itself.
3. Migration filenames follow exactly: four decimal digits (the version), an underscore, a lowercase snake-case name starting with a letter, and the literal lowercase suffix `.up.sql` or `.down.sql`. Examples: `0001_create_users.up.sql`, `0001_create_users.down.sql`. Files with non-matching forms are rejected.
4. The version portion is a positive integer. Version zero is rejected. Source order is checked to be strictly increasing and free of duplicates before any side effect.
5. Each migration must have both an Up and a Down file with the same version and name. Missing pairs are rejected. Empty migration bodies are rejected before any mutation.
6. Version gaps are allowed; source order still must be strictly increasing and unique.
7. The metadata ledger table has exactly four columns: a positive version as primary key, a non-empty name, a checksum stored as exactly 64 lowercase hexadecimal characters, and an applied-at timestamp in UTC RFC3339Nano text. The ledger table itself enforces these constraints.
8. The checksum hashes a precise framing whose version field is the canonical base-10 numeric version with no leading zeros (the valid `0001` filename is parsed and contributes the version field `1`); the name bytes; the exact Up bytes; and the exact Down bytes. Each of the four fields is prefixed with its decimal byte length followed by an ASCII colon before being concatenated. The SHA-256 of those framed bytes is stored as exactly 64 lowercase hex characters. This avoids delimiter ambiguity. The checksum helper receives the parsed numeric version, not raw filename text, and a test proves that parsing the valid four-digit filename produces numeric version 1 and canonical checksum field `1`. A one-digit migration filename remains invalid. The framing is described in prose only; no code is provided here.
9. The metadata ledger bootstrap is idempotent (it may use `CREATE TABLE IF NOT EXISTS` for the ledger table itself). Migration bodies must not hide drift with broad `IF NOT EXISTS`.
10. The source is validated entirely before the metadata ledger is even bootstrapped, so a malformed source produces no database mutation.
11. After bootstrap, the runner validates every applied row against the source: an applied version absent from the source is an unknown-version failure; a changed checksum is drift; a changed name is drift; impossible ordering is drift. The meaningful invariant for the applied-history set is that it must be a prefix of the ordered source. If a later known source version is recorded in the ledger while an earlier source version is absent, the runner reports inconsistent history before any mutation. Version gaps in the source remain allowed. Any failure aborts before mutation.
12. Up applies each pending version in ascending order, in its own transaction, with the migration SQL and the metadata insert in the same transaction. The injected clock is read once per successfully applied Up inside that transaction. A failing migration rolls back its data and its version row; earlier successful versions remain committed.
13. Re-running Up at the latest applied version is a no-op and does not rewrite any timestamp.
14. Down-one reverts only the latest applied known version. Down SQL and the metadata deletion share one transaction. Down-all prevalidates once, then performs one migration per transaction in reverse order; if a later Down fails, already reverted later versions remain reverted and the failing version remains applied. State this honest partial progress.
15. Empty applied state with a non-empty source: Down-one returns a typed no-migrations outcome; Down-all is a no-op. Empty applied state with an empty source: source validation succeeds, Up bootstraps the ledger then returns no-op, Down-one returns no-migrations, and Down-all is a no-op. Empty source with an applied ledger is an unknown-version failure.
16. Context cancellation before starting the next migration stops before opening it, leaving earlier successful migrations committed. Context cancellation during a migration causes that single transaction to roll back; already committed earlier migrations remain. The runner does not pretend a transaction is open "between migrations" when none exists; it checks context before opening the next transaction and uses the standard transaction rollback on cancellation during an open transaction.
17. The runner does not claim cross-process or advisory-lock safety. DDL transactional guarantees are pinned to SQLite only.

## 9. Inputs and Outputs
Up inputs: context, `*sql.DB`, embedded migration source, injected clock. Up outputs: applied versions, a typed outcome (applied, no-op, drift, unknown, duplicate, malformed, cancelled, error), and exact metadata assertions. Down-one inputs: context, `*sql.DB`, embedded migration source. Down-one outputs: typed outcome (reverted, no-migrations, drift, unknown, cancelled, error). Down-all inputs: same as Down-one. Down-all outputs: typed outcome (reverted-all, no-op, partial-with-error, error).

## 10. Rules and Edge Cases
Drift is an error, never a silent overwrite. Missing Down file is a typed failure before mutation. Stepwise and full Down must never go below zero applied versions. Cancellations before and between migrations are handled honestly. Re-running Up at latest is a no-op and does not rewrite timestamps. Empty migration bodies are rejected. The runner does not promise cross-process safety.

## 11. Project Constraints
Single-process runner only. No third-party migration framework. Do not use `IF NOT EXISTS` in migration bodies except inside the metadata ledger bootstrap. Do not silently advance timestamps. State honest semantics about data and schema state after any failure.

## 12. Design Questions Before Coding
What is the exact checksum framing? How are duplicate and out-of-order versions detected before any mutation? How are malformed names and suffixes detected? How does the runner distinguish "no migrations to apply" from "applied version unknown"? How is a Down failure distinguished from a partial commit? How is context cancellation between migrations handled? How is cross-process safety explicitly disclaimed?

## 13. Implementation Milestones
1. Define the source abstraction, the metadata ledger, the typed outcomes, and the in-process mutex that serializes Up, Down-one, and Down-all.
2. Implement filename validation, pair validation, ordering, gap tolerance, and the canonical base-10 version parsing used by the checksum framing.
3. Implement the checksum framing and a known-value test fixture, including a test that parses valid filename version `0001` to numeric version 1 and proves the checksum uses canonical field `1`; do not accept a one-digit migration filename.
4. Implement metadata ledger bootstrap with idempotency.
5. Implement Up with full source and metadata validation, per-migration transactions, and clock injection.
6. Implement Down-one and Down-all with empty-source / empty-ledger behavior and partial-progress semantics.
7. Implement the prefix-of-source invariant check over the applied set.
8. Add tests for every failure mode and every honest outcome, including single-instance concurrent calls under the mutex.
9. Run the unit gate and the race detector.

## 14. Verification Cases the Learner Must Write
- Full-source prevalidation rejects malformed names (wrong length, non-digit version, bad suffix, non-letter first character of name, uppercase, wrong extension) and exits before any database side effect.
- Prevalidation rejects version zero, duplicate versions, and non-increasing source order without touching the database.
- Prevalidation rejects a missing pair (Up without Down or vice versa) without touching the database.
- Prevalidation rejects empty migration bodies without touching the database.
- A checksum known-value test asserts the exact framing: same inputs yield the same hash; flipping a byte in Up or in Down, changing the numeric version, or changing the name changes the hash. The test calls the checksum helper with numeric version 1 and proves that parsing valid filename version `0001` supplies that value and contributes canonical field `1`; alternate one-digit filenames are rejected by filename validation rather than treated as checksum inputs.
- A framing-collision test asserts that field-boundary ambiguity is impossible: e.g., a name that ends in digits followed by a version that starts with digits cannot produce the same checksum as a different split.
- Fresh Up applies every pending version and writes a metadata row per version with the correct UTC RFC3339Nano timestamp and checksum.
- Idempotent Up: re-running Up on the latest applied state is a no-op and does not rewrite any timestamp.
- Persistence: closing and reopening the database preserves applied versions, schema, and data.
- Stepwise Down removes only the latest version and its metadata row, in one transaction.
- Full Down-all removes every version in reverse order, never going below zero, and ends at zero applied versions.
- Down-all on empty applied state is a no-op.
- Down-one on empty applied state returns a typed no-migrations outcome.
- Up SQL failure rolls back that migration's data and its metadata row while earlier successful versions remain committed.
- Down SQL failure preserves both the schema/data state and the version record for that migration.
- A metadata-insert failure induced deterministically by the test database (for example, by violating the ledger table's own constraints) rolls back the migration and leaves earlier versions committed.
- Down-all partial progress: a forced failure on a later Down leaves already reverted later versions reverted and the failing version applied; the runner reports this honestly.
- Checksum drift in Up bytes is detected and reported; no migration runs.
- Checksum drift in Down bytes is detected and reported even when Up bytes are unchanged; no migration runs.
- Name drift is detected and reported.
- Unknown applied version (metadata row whose version is absent from the source) is detected and reported.
- Inconsistent history is detected: a later known source version recorded in the ledger while an earlier source version is absent is reported before mutation.
- Empty source with empty ledger: source validation succeeds, Up bootstraps the ledger and returns no-op, Down-one returns no-migrations, Down-all is no-op.
- Empty source with applied ledger: Up (or any pre-mutation validation) reports unknown-version failure before any further action.
- Cancellation before starting a migration stops before opening it.
- Cancellation during a migration rolls back that single transaction and leaves earlier committed versions committed.
- Already-cancelled context before Up short-circuits without writing.
- Already-cancelled context before Down-one short-circuits without writing.
- Already-cancelled context before Down-all short-circuits without writing.
- Schema assertions: tables, indexes, and columns match the migration SQL.
- Data assertions: inserted rows match the migration SQL.
- Ledger assertions: every applied row has the exact version, name, checksum, and applied-at value.
- Race detector: `go test -race ./...` is clean; the runner is not claimed to be safe across multiple processes.
- Single-instance concurrency: concurrent calls on the same `Runner` instance are serialized by the in-process mutex and never double-apply; the test exercises parallel calls and asserts a stable end state. Multi-process or multi-instance safety is not claimed and is not exercised.
- No third-party migration framework is introduced.

## 15. Common Mistakes to Watch For
Hiding drift with `IF NOT EXISTS` inside migration bodies. Hashing only Up bytes or only Down bytes. Computing the checksum over a different framing and producing collisions. Including leading-zero version text as part of the checksum framing. Mixing SQL and metadata writes across transactions. Allowing Down below zero. Claiming cross-process safety. Inventing timestamps. Skipping prevalidation and bootstrapping the ledger on a malformed source. Reusing the same UTC applied-at across multiple migrations. Using `CREATE TABLE IF NOT EXISTS` inside migration bodies. Confusing "out-of-order applied versions" with the actual invariant (the applied set must be a prefix of the ordered source).

## 16. Topics and References for Study
Go `embed.FS` documentation. SQLite documentation on transactional DDL and the documented limits where DDL is not transactional. `database/sql` transaction handling and error mapping. SHA-256 framing patterns and the importance of length prefixes. UTC RFC3339Nano formatting. Single-process correctness versus distributed coordination.

## 17. Self-Assessment Questions
What does the checksum cover and why? Why is the framing length-prefixed rather than delimiter-separated? Why is the version field in the framing canonical base-10 with no leading zeros, and why is a filename-padding test required? Why is the metadata ledger bootstrap allowed to be idempotent but migration bodies are not? What is honest about transactional DDL on SQLite? What is honest about partial progress in Down-all? Why is cross-process safety explicitly disclaimed while single-process in-instance concurrency is serialized by a mutex? Why is the meaningful invariant a "prefix of the ordered source" rather than a row-order check? How does cancellation before starting a migration differ from cancellation during an open transaction?

## 18. Definition of Completion
- [ ] `go test ./...` passes with no Docker, network, PostgreSQL, Redis, or environment variables.
- [ ] `go test -race ./...` passes.
- [ ] Full-source prevalidation rejects malformed names, missing pairs, empty bodies, version zero, duplicates, and out-of-order versions without touching the database.
- [ ] Checksum framing is described in prose and verified by a known-value test.
- [ ] Each Up migration runs in its own transaction with the metadata insert; a failure rolls back both the migration and its metadata row while earlier versions remain committed.
- [ ] Down-one reverts only the latest version; Down-all never goes below zero.
- [ ] Down-all on empty applied state is a no-op; Down-one on empty returns a typed no-migrations outcome.
- [ ] Drift in Up bytes, Down bytes, or name is detected and reported without mutation.
- [ ] Unknown applied version is detected and reported without mutation.
- [ ] Pre-cancellation before the next migration and during an open transaction are handled honestly.
- [ ] The runner explicitly disclaims cross-process safety.
- [ ] No third-party migration framework is introduced.

## 19. Optional Extensions
Add a separately documented dry-run listing planned Up operations without applying them. Add a separately tagged experiment that records a per-migration timing trace, without weakening the unit gate.
