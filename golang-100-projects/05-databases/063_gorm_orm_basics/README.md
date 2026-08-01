# Project 063 — GORM ORM Basics

## 1. Project Name and Number
Project 063, gorm_orm_basics. This README is a learning guide only. You will create every Go source file, the schema bootstrap, and every test file yourself in `05-databases/063_gorm_orm_basics/`. The guide does not provide implementation code.

## 2. Project Idea
Model Users and Projects with GORM, where one User owns many Projects, the user email is normalized and unique, and the project owner foreign key uses the RESTRICT delete policy. Pin `gorm.io/gorm` `v1.31.2` and `gorm.io/driver/sqlite` `v1.6.0` as the only direct third-party dependencies. Use a fresh temporary SQLite file per test, enable foreign-key enforcement, and cap open and idle connections at 1. Make the N+1 problem observable through an injected counting GORM logger.

## 3. Why This Project Now?
This follows Project 062 (postgres_sqlx) and applies ORM concepts to the same User domain established in Project 061 (sqlite_crud). Context propagation follows Project 041. No other project is formally required.

## 4. Prerequisites
Projects 062, 061, and 041 are required. No other project is formally required. The regular unit gate is `go test ./...`, which must pass with no Docker, no PostgreSQL, no Redis, no network, and no environment variables.

## 5. What You Must Know Before Starting
You should know the Project 062 User semantics, basic GORM usage (models, associations, query chains, preloads), how GORM translates driver errors when configured to do so, Go context propagation, and how to enable SQLite foreign-key enforcement at the connection level.

## 6. Explanation of New Concepts
Models and tags versus domain: a GORM model is a struct with column tags and optional relationship tags. The domain-facing values used in your repository must be kept in distinct domain structs separate from the persistence model structs; explicit mapping is the separation mechanism. Storage concerns must not leak into domain types.

AutoMigrate limits: AutoMigrate is a development convenience. It can add missing columns and indexes, but it cannot drop columns, rename columns, or perform complex migrations safely. Its production limitations must be stated, and Project 064 will replace this habit with explicit, transactional migrations.

Association transactions: creating a parent and its children in one GORM call (or wrapped in `db.Transaction(...)`) ensures partial association creation rolls back on failure. Without a transaction, a failure mid-way leaves inconsistent state.

Foreign-key enforcement and RESTRICT: SQLite enforces foreign keys only when `PRAGMA foreign_keys = ON` is set on every connection that needs it. RESTRICT means a delete of the parent row is rejected if child rows exist; the parent's row remains untouched.

Preload and N+1: listing N users and their projects naively triggers one query for users and one query per user for projects, totaling N+1 queries. Preload tells GORM to issue one extra query for the projects of all listed users, turning N+1 into 2. The exact count depends on the operation; it must be observed, not guessed. Preload is registered on the users query chain before the terminal query executes; GORM then executes the users query followed by the projects query. Calling Preload after an already-executed user query is wrong and the project does not describe that pattern.

Counting logger: GORM emits a structured trace callback for each executed statement when configured. A counting logger increments a counter on each completed trace. It counts callbacks, not log text, and it is reset immediately before the operation under test.

GORM zero-value update behavior: GORM's single-column `Update` (for example `db.Update("archived", false)`) writes that single column even when the value is the zero value. The struct form `Updates(&Project{Archived: false})` omits zero-valued fields by default and silently does nothing. A map form `Updates(map[string]any{"archived": false})` writes the zero value. `Save` writes every column and behaves as an upsert. Core updates must not use unconstrained `Save` and must use an explicit selected-field update (single-column `Update` or a map) when the new value is the zero value. Demonstrating this pitfall is part of the project.

Deterministic NowFunc/time: GORM lets you inject a `time.Now` function. Using a controlled clock makes tests reproducible.

Error translation: when GORM is configured with `TranslateError: true`, the driver errors are translated into typed errors such as `gorm.ErrRecordNotFound`. Translation works only when enabled. Without it, you receive the raw driver error.

## 7. Learning Objective
Implement explicit User and Project models with documented constraints, demonstrate relationship creation in a transaction, query the owner and projects deterministically, preload projects with a documented query-count bound, and prove the N+1 hypothesis empirically. Demonstrate the zero-value update pitfall and its intentional workaround using an explicit selected-field update. Email and name validation follow the exact Project 061 rules and run before any GORM call; persistence constraints remain as defense in depth.

## 8. Functional Requirements
1. Dependencies are pinned exactly: `gorm.io/gorm` `v1.31.2` and `gorm.io/driver/sqlite` `v1.6.0`.
2. Every test opens a fresh temporary SQLite file with the SQLite DSN configured to enable foreign-key enforcement on every connection that the driver opens, and caps `SetMaxOpenConns(1)` and `SetMaxIdleConns(1)`. After opening, the test queries the `PRAGMA foreign_keys` setting to verify it is on, and inspects the generated foreign-key metadata for the projects table to verify the on-delete action is `RESTRICT` rather than relying solely on a tag.
3. AutoMigrate is allowed only as this project's local and test bootstrap. Its production limitations are stated in the README.
4. User model: positive 64-bit integer ID, trimmed non-empty name, normalized unique email, created-at, updated-at. Project model: positive 64-bit integer ID, required `OwnerID` foreign key, trimmed non-empty name, boolean `Archived`, created-at, updated-at. Domain-facing values are kept in distinct domain structs separate from the persistence model structs; explicit mapping (not the `gorm:"-"` tag, which merely ignores a model field) is the separation mechanism.
5. Project model: positive 64-bit integer ID, required `OwnerID` foreign key, trimmed non-empty name, boolean `Archived`, created-at, updated-at.
6. The foreign-key delete policy is exactly RESTRICT. Deleting a user that still owns projects fails with a stable typed outcome. The User row and its Project rows are unchanged after a failed RESTRICT delete.
7. GORM is configured with `TranslateError: true`. Record-not-found, duplicated key, and foreign-key violation are mapped to stable domain outcomes without message matching.
8. A GORM time source is injected so that `created_at` and `updated_at` come from a deterministic clock. Each individual model create or update uses the injected source deterministically. An aggregate may invoke the source separately for the parent and for the association operations; the project does not promise one clock read for the entire aggregate unless the learner explicitly implements that boundary. Backward or equal injected timestamps are reported honestly and not silently advanced.
9. Context is applied to every operation.
10. Creating a user-with-projects aggregate accepts the new user's fields plus child project specifications without an `OwnerID` field. The aggregate input does not carry `OwnerID` because the user ID is generated by the database during this same transaction; the repository assigns the newly generated user ID to each child project inside the transaction and persists them as associations. The aggregate runs in one transaction so partial association creation rolls back. Any invalid project or association failure leaves no partial state.
11. Querying an individual owner and their projects uses explicit GORM calls rather than implicit joins. Querying users ordered by ID uses `ORDER BY id ASC`. Preloading projects on a users list registers `Preload("Projects")` on the users query chain before the terminal query executes; projects are ordered by ID.
12. Listing at least two seeded users with `Preload("Projects")` issues exactly two SQL statements: one users query and one projects query. The counter is the counting logger's callback count, reset immediately before the operation.
13. Listing zero users issues a single users query. Empty-list behavior is tested separately and is allowed to differ.
14. Core updates do not use unconstrained `Save`. The zero-value pitfall is demonstrated by deliberately changing `Archived` from true to false using an explicit selected-field update. The new false value persists.
15. Stable mappings for record-not-found, duplicate, and foreign-key violation exist and are exercised by tests.

## 9. Inputs and Outputs
User Create: context, name, email. Output: User with positive ID and UTC timestamps from a single clock read, or typed invalid input or duplicate. User Get: context, ID. Output: User or typed not-found. User List: context. Output: non-nil slice in ascending ID order. User Delete: context, ID. Output: typed not-found or success; RESTRICT failure leaves data unchanged. Aggregate User-with-Projects Create: context, user fields (name, email), and child project specifications without `OwnerID`. Output: aggregate (User with positive ID plus its created Projects) or typed invalid input, duplicate, or other typed errors. Standalone Project Create: context, existing `OwnerID`, name, archived. Output: Project or typed invalid owner (missing user), invalid input, or other typed errors. Project List by owner: context, `OwnerID`. Output: non-nil slice in ascending ID order. Project Update: context, ID, fields. Output: updated Project with timestamps set, or typed not-found or invalid.

## 10. Rules and Edge Cases
Reject a missing owner when creating a project against a nonexistent User ID. RESTRICT rejection leaves both the User and its Projects rows untouched. Duplicate email on Create and on Update is mapped to a stable typed duplicate. Record-not-found is mapped to a stable typed not-found. Zero-value updates require explicit field selection. Pre-cancelled context short-circuits before any operation. The counting logger must exclude setup statements: reset immediately before the operation under test.

## 11. Project Constraints
No HTTP layer. No third-party migration tool. No external services. Domain-facing types and persistence model types are distinct structs; explicit mapping is mandatory. `TranslateError: true` is enabled for the duration of the test database. Tests use temp files; no Docker, no network, no environment variables.

## 12. Design Questions Before Coding
Where exactly does domain-to-model translation happen? How will the counting logger be reset and read in a way that excludes setup? How will RESTRICT be confirmed rather than assumed? How will the zero-value pitfall be demonstrated and verified? How will partial-association rollback be exercised? How will the empty-list case be distinguished from the populated-list case for query counting?

## 13. Implementation Milestones
1. Define distinct domain types and distinct GORM models with explicit tags, with explicit mapping between them.
2. Configure the SQLite test database with foreign-key enforcement, single connection, and `TranslateError: true`. Verify the pragma after open.
3. Inject the GORM clock and the counting logger; ensure the logger counts only completed SQL trace callbacks.
4. Implement user Create, Get, List, and Delete with stable typed outcomes.
5. Implement aggregate user-with-projects create in one transaction; the input supplies user fields plus child project specifications without `OwnerID`, and the repository assigns the newly generated user ID to each child inside the transaction; cover the rollback path.
6. Implement project create against a user, project list by owner, and project update using an explicit selected-field update.
7. Add mapping for record-not-found, duplicate, and foreign-key violation with `TranslateError: true`.
8. Add tests for N+1 prevention, empty-list count, zero-value update, RESTRICT, ordering, and concurrency.

## 14. Verification Cases the Learner Must Write
- Relationship create-and-read: a user with two projects persists; the user is returned by ID, and the projects are returned for that owner in ascending order.
- Aggregate create in a transaction: invalid aggregate input is rejected before any write (a separate test). A separate test uses a test-only injected failure that occurs after the user insert but before or during project persistence, and asserts that no user row and no project row remain afterwards. Pre-validation alone does not prove rollback.
- Missing owner on project create returns a typed foreign-key or invalid-owner outcome without creating a project.
- Duplicate normalized email on user create returns typed duplicate; mixed-case input is normalized before binding.
- Duplicate normalized email on user update returns typed duplicate.
- `Preload("Projects")` registered on the users query chain before the terminal query executes returns projects in ascending ID order with the right owners after the users query runs.
- Listing two seeded users with `Preload("Projects")` issues exactly two SQL statements after the counting logger is reset immediately before the operation.
- Listing zero users issues a single users query; the counting logger records exactly one callback.
- A naive loop that fetches projects per user issues N+1 callbacks and is flagged by a separate test as the wrong pattern.
- Zero-value update: a project with `Archived=true` is updated to `Archived=false` using a selected-field update and the new value persists across a fresh `Get`.
- Unconstrained `Save` is not used in core; a bounded source-review or static gate over the core persistence files is required, and behavior tests cover the intentional selected-field update of the `Archived` boolean. Runtime interception of a method that is not behind the boundary is not required.
- RESTRICT deletion: deleting a user that owns projects fails with a stable typed outcome and leaves the user and projects unchanged.
- Foreign-key pragma is verified after opening the test database, not assumed from a tag.
- Deterministic ordering: list returns users in ascending ID order regardless of insertion order.
- Pre-cancelled context before a query returns the context error.
- Pre-cancelled context before a mutation returns the context error and does not write.
- Translated errors: record-not-found and duplicate each map to a stable typed outcome that callers can detect.
- Independent temp databases: two parallel test databases do not see each other's rows.
- Race detector: `go test -race ./...` is clean across all tests.

## 15. Common Mistakes to Watch For
Forgetting `PRAGMA foreign_keys = ON` and seeing RESTRICT not enforced. Setting `SetMaxOpenConns` greater than 1 and seeing non-deterministic behavior. Assuming a tag produced a foreign key without verifying the pragma and inspecting the generated on-delete action. Counting log strings instead of callbacks. Resetting the counter too early or too late. Using `:memory:` and seeing per-connection state. Using `Save` and silently overwriting fields. Using the struct form of `Updates` for a zero value and silently doing nothing. Treating zero values as "no change". Mixing domain types and GORM models and leaking storage concerns. Conflating `gorm:"-"` (which only ignores a field) with real domain/persistence separation. Using `TranslateError: false` and matching error messages by string. Confusing pool concurrency with transaction rollback semantics. Pretending pre-validation alone proves transaction rollback.

## 16. Topics and References for Study
The GORM `v1.31.2` documentation on models, tags, associations, transactions, preloads, and `TranslateError`. The SQLite driver's documentation on DSN options for foreign-key enforcement. SQLite documentation on `PRAGMA foreign_keys` and on foreign-key actions including RESTRICT. Go context propagation patterns. Counting and resetting shared state in tests. The pinned GORM and driver documentation.

## 17. Self-Assessment Questions
What does `TranslateError: true` change about error handling? Why is the foreign-key pragma verified and the on-delete action inspected rather than trusted? Why is "exactly two SQL statements" measured by callbacks rather than by time or log text? Why is `Save` rejected as the default update path? Why does the struct-form `Updates` silently skip zero values while single-column `Update` and the map form write them? How does RESTRICT preserve data? Why does empty-list count differ from populated-list count? Why must domain/persistence separation be implemented as distinct types with mapping rather than a `gorm:"-"` tag? Why does pre-validation alone not prove transaction rollback?

## 18. Definition of Completion
- [ ] `go test ./...` passes with no Docker, network, PostgreSQL, Redis, or environment variables.
- [ ] `go test -race ./...` passes.
- [ ] The foreign-key pragma is verified after open.
- [ ] Aggregate create runs in one transaction and rolls back on failure, demonstrated by a separate injected-failure test rather than only by pre-validation.
- [ ] `TranslateError: true` is enabled and mappings for record-not-found, duplicate, and foreign-key violation are tested.
- [ ] The two-statement bound for listing users with `Preload("Projects")` is observed by a callback count, not by time or log scraping.
- [ ] The zero-value update from `Archived=true` to `Archived=false` persists and is verified.
- [ ] RESTRICT deletion fails with a stable typed outcome and leaves data unchanged.
- [ ] No third-party dependency beyond `gorm.io/gorm v1.31.2` and `gorm.io/driver/sqlite v1.6.0` is introduced.
- [ ] No HTTP layer, ORM-managed migration tool, or implementation code appears in this README.

## 19. Optional Extensions
Add a separately tagged soft-delete experiment using a deleted-at column and a query scope. Add a separately documented benchmark that counts statements under a fixed scenario, without weakening the unit gate.
