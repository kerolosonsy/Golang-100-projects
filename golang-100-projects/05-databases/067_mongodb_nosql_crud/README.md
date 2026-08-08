# Project 067 — MongoDB NoSQL CRUD

## 1. Project Name and Number

- Project 067, mongodb_nosql_crud.
- This README is a learning guide only.
- You will create every source, Compose, and test file yourself in `05-databases/067_mongodb_nosql_crud/`.
- This guide contains no implementation code, signatures, snippets, pseudocode, BSON fragments, or solution commands.

## 2. Project Idea

Build MongoDB CRUD for user-like documents with ObjectID identity, a normalized unique email, injected UTC timestamps, explicit index creation, and deterministic cursor pagination over creation time and ID.

## 3. Why This Project Now?

- Project 067 follows Project 066 in the catalog and contrasts its PostgreSQL transactional constraints with MongoDB document modeling, driver error classification, index-backed uniqueness, cursor lifecycle, and stable tuple pagination.
- Projects 061 and 041 remain required foundations, including the email-normalization contract from Project 061.

## 4. Prerequisites

- Required prerequisites: Projects 066, 061, and 041.
- Optional review: none.
- Reuse Project 061 email normalization exactly.
- The normal unit gate must need no Docker, MongoDB, network, or environment variables.
- MongoDB integration is separate and opt-in.

## 5. What You Must Know Before Starting

- Know contexts and deadlines, typed errors, interfaces, injected clocks, ObjectID, BSON-to-domain mapping, unique indexes, cursor iteration, tuple ordering, and opaque cursor encoding.
- Understand that domain values must not expose persistence-only BSON models.

## 6. Explanation of New Concepts

### Concepts

- The stored document contract contains ObjectID, trimmed non-empty name, Project-061-normalized unique email, created-at UTC timestamp, and updated-at UTC timestamp.
- Normalization trims surrounding whitespace, lowercases, requires exactly one `@`, non-empty local and domain parts, and no whitespace.
- Create owns ID and timestamps.
- Full update changes only name and email, preserves ID and created-at, and reads the injected clock once for updated-at.

- MongoDB BSON datetimes are stored as integer milliseconds since the Unix epoch.
- Every injected clock value is normalized to a UTC millisecond-precision instant before persistence and before it is returned.
- Create reads the injected clock once, normalizes that value to UTC millisecond precision, and uses the normalized instant for both created-at and updated-at.
- Update reads the injected clock once, normalizes it the same way, and uses it for updated-at only.
- Cursor RFC3339Nano text represents this millisecond-precision UTC instant.
- Sub-millisecond information supplied by the clock is truncated by normalization and is never exposed as nano precision; tests must not expect nanosecond round-trip and must prove the documented truncation.

- External IDs are canonical ObjectID hexadecimal strings.
- Reject malformed or noncanonical IDs as typed invalid input before any driver call.
- Repository results are domain values, never BSON models.

- Startup creates one ascending unique index named `uniq_users_email` on the stored normalized email.
- Index creation is verified by inspecting the resulting or listed definition; any creation, name, key, options, or verification mismatch is startup failure.
- Duplicate-key and not-found classification uses official driver error types, helpers, or documented result counts, never message text.

- List sorts ascending by the tuple of created-at and ObjectID.
- A cursor carries the last tuple and the next query starts strictly after it.
- Limit must be from 1 through 100 with no silent clamping.
- Fetch one extra document; return a next cursor only when more results exist.
- Skip/offset pagination is forbidden.

- Cursor text is versioned, URL-safe, opaque, and unauthenticated.
- It carries the millisecond-precision UTC RFC3339Nano created-at instant and the canonical ObjectID.
- Reject bad version, encoding, field count, timestamp, timezone, ID, and semantically impossible values before a driver call.
- Tamper resistance is out of scope.
- Pagination guarantees no duplicates or omissions only for a stable dataset; concurrent mutation can move page boundaries.

- Every driver operation gets a context with a bounded deadline.
- Always close list cursors, check terminal cursor error, and add operation context to decode failures.
- Update uses an explicit mutable field set so ID and created-at cannot change.

## 7. Learning Objective

- Build exact document, validation, index, CRUD, error, context, cursor-resource, and pagination contracts that remain deterministic and testable without requiring MongoDB for the normal test suite.

## 8. Functional Requirements

1. Use the official `go.mongodb.org/mongo-driver/v2` exactly at `v2.8.0`.
2. Persist ObjectID, trimmed non-empty name, normalized unique email, created-at UTC, and updated-at UTC.
3. Create accepts name and email only, creates ObjectID, reads the injected clock once, normalizes it to UTC millisecond precision, and uses that normalized instant for both timestamps.
4. External IDs must be canonical ObjectID hex. Invalid or noncanonical text returns typed invalid input before any driver call.
5. Get returns a domain value or typed not-found.
6. Full update validates name and email, preserves ID and created-at, reads the injected clock once, normalizes it to UTC millisecond precision, and uses the normalized instant for updated-at.
7. Delete returns typed not-found when no document is removed.
8. Startup creates one ascending unique index named `uniq_users_email` on the stored normalized email. Startup verifies the resulting or listed index definition. Any creation error, name mismatch, key mismatch, options mismatch, or verification error is startup failure.
9. Duplicate key is mapped through driver-supported error inspection, not text. Not-found is mapped through documented driver outcomes, not text.
10. Every driver call has a bounded context deadline derived from the caller context.
11. List returns domain values sorted by created-at ascending and ObjectID ascending.
12. List limit must be 1 through 100. Invalid values return typed invalid input without a driver call; no clamping occurs.
13. Fetch limit plus one, return at most limit, and emit next cursor only when the extra document proves another page.
14. Cursor is versioned URL-safe opaque text carrying the millisecond-precision UTC RFC3339Nano created-at instant and the canonical ObjectID.
15. Reject malformed version, encoding, timestamp, timezone, ObjectID, and impossible cursor tuple before a driver call.
16. Query begins strictly after the cursor tuple. No skip or offset is used.
17. Always close cursors and check cursor error after iteration. Wrap decode failures with operation context.
18. Unit tests use repository or store boundary fakes and require no MongoDB or Docker. Integration is separately tagged and guarded.
19. Transactions and change streams are outside project scope.
20. Timestamp domain values are always UTC millisecond precision. Sub-millisecond data in the injected clock is truncated by normalization and never exposed as nano precision.

## 9. Inputs and Outputs

### Interface Contract

- Create takes context, name, and email and returns a complete domain document whose created-at and updated-at are the same UTC millisecond-precision instant from a single clock read, or typed invalid/duplicate/internal outcome.
- Get takes context and canonical ObjectID hex and returns a domain document or typed invalid/not-found/internal outcome.
- Full update takes context, ID, name, and email and returns the preserved identity and created-at with one new UTC millisecond-precision updated-at.
- Delete returns success or typed invalid/not-found/internal outcome.
- List takes context, limit, and optional cursor and returns ordered domain results plus a next cursor only when more exist.
- Returned timestamps represent millisecond-precision UTC instants.

- Example behavior: three documents sharing one created-at millisecond value are ordered by ObjectID.
- A page limit of two returns the first two and a cursor.
- The next page starts strictly after the second tuple and returns the third without repetition.

## 10. Rules and Edge Cases

- Reject blank names, invalid normalized emails, malformed IDs, malformed cursors, and limits outside 1 through 100 before a driver call.
- Normalize every injected clock value to UTC millisecond precision; sub-millisecond data is truncated and not exposed.
- Do not expose BSON models.
- Do not update ID or created-at.
- Do not silently clamp limits.
- Do not use skip/offset.
- Always close cursors even when decoding fails, and always check terminal cursor error.
- State stable-dataset pagination and unauthenticated cursor limitations honestly.

## 11. Project Constraints

- No ORM, transactions, change streams, cursor authentication, or shared integration database.
- Normal tests require no Docker.
- Integration files use the `integration` build tag.
- Runtime activation generates and validates a collision-resistant owned database name before any connection; partial activation, malformed connection settings, or activation that cannot prove owned scope fails closed without connecting.
- After connection the integration uses only that database and drops only that database through bounded cleanup.
- Never log credentials, drop shared databases, or clean data outside owned scope.

## 12. Design Questions Before Coding

- How are persistence models separated from domain values?
- Where is Project 061 normalization reused?
- How is canonical ObjectID text checked before a driver call?
- How is the ascending unique email index named `uniq_users_email` created and verified at startup?
- How is UTC millisecond precision enforced on inbound clock values and on timestamps round-tripped to drivers and cursors?
- Which documented driver mechanisms classify duplicate and not-found?
- How is cursor closure guaranteed on decode failure?
- Why does stable-dataset pagination need an explicit limitation?
- How is the owned integration database name validated before connection?

## 13. Implementation Milestones

1. Define domain document, persistence mapping, typed outcomes, injected UTC clock, millisecond normalization, and store boundary.
2. Add name, email, canonical ObjectID, limit, cursor, and millisecond-precision timestamp normalization.
3. Add startup creation of the ascending unique index named `uniq_users_email` on stored normalized email, list or describe to verify the resulting definition, and propagate any mismatch as startup failure.
4. Add Create, Get, full Update, and Delete with exact immutable-field behavior and one normalized UTC millisecond-precision clock read for each timestamp-producing operation.
5. Add tuple-sorted list, fetch-one-extra logic, and versioned cursor encoding and validation that carries the millisecond-precision UTC RFC3339Nano instant.
6. Add bounded contexts, cursor closure, terminal cursor-error checks, and contextual decode errors.
7. Complete fake-based unit tests without MongoDB.
8. Separately add guarded tagged MongoDB integration that generates and validates a collision-resistant owned database name before connection, connects, uses only that database, and drops only that database through bounded cleanup.

## 14. Verification Cases the Learner Must Write

### Required Cases

Unit tests without Docker:
- Create trims name, normalizes email exactly as Project 061, creates ObjectID, normalizes the injected clock to UTC millisecond precision, and uses that one normalized instant for both timestamps.
- A clock with documented sub-millisecond data produces timestamps truncated to UTC millisecond precision and exposed as such; nano round-trip is not asserted.
- Blank name and every invalid email form fail before a store call.
- Canonical ObjectID succeeds; malformed, wrong-length, nonhex, and noncanonical IDs fail before a driver call.
- Get, update, and delete map not-found distinctly.
- Duplicate create and update map through simulated official driver classification, never text.
- Startup creates the ascending unique email index named `uniq_users_email`; any mismatch in name, key, options, or verification is startup failure.
- Full update changes name/email only, preserves ID/created-at, and reads and normalizes the clock once for updated-at.
- List returns domain values in created-at then ObjectID ascending order, with timestamps at UTC millisecond precision.
- Limits 1 and 100 work; zero, negative, and above 100 fail without clamping or driver access.
- Fetch one extra determines next cursor; no next cursor appears on the final page.
- Cursor round trip preserves the millisecond-precision UTC RFC3339Nano instant and canonical ObjectID.
- Bad cursor version, URL-safe decoding, timestamp, timezone, ID, and impossible values fail before a driver call.
- Queries continue strictly after the tuple without skip/offset.
- Traversing a stable dataset yields no duplicate or missing document.
- Cursor closes on success, decode failure, and iteration failure; terminal cursor error is checked.
- Caller cancellation and derived deadline errors remain discoverable.

Opt-in MongoDB integration tests:
- Integration files are excluded from the normal gate by the `integration` build tag.
- With that tag, both runtime activation values absent produce a clear skip. The values are a MongoDB connection setting and an explicit destructive-test guard. Partial activation, malformed connection settings, missing or wrong guard, or inability to generate and validate a collision-resistant owned database name fails closed before any connection or mutation. Database existence on the server is not asserted before connection; the owned name is generated and validated before connecting.
- Connect, create or use only that generated database, record ownership, and drop only that database through bounded cleanup. Never print credentials, drop shared databases, or destroy unowned data.
- Create and verify `uniq_users_email` and confirm its name, ascending email key, and unique option.
- Verify Create, Get, full Update, Delete, duplicate normalized email, and not-found, with timestamps at UTC millisecond precision.
- Verify invalid ObjectID is rejected before driver access.
- Verify pagination boundaries, tuple order, fetch-one-extra behavior, and full stable traversal.
- Verify context cancellation and cursor cleanup where observable.
- Verify concurrent inserts with unique emails and duplicate collision behavior.

## 15. Common Mistakes to Watch For

- Using obsolete driver paths, storing raw email, accepting noncanonical IDs, letting callers set timestamps, asserting nanosecond round-trip against a millisecond-precision store, replacing whole documents during update, ignoring index setup failure or accepting a mismatched name or options, matching errors by text, using skip pagination, clamping limits, returning a cursor without an extra result, failing to close or check cursors, exposing BSON models, claiming mutation-safe pagination, attempting to prove a shared database exists before connection, dropping a shared database, or cleaning unowned data.

## 16. Topics and References for Study

- Study official `go.mongodb.org/mongo-driver/v2` `v2.8.0` documentation for ObjectID, BSON datetime millisecond precision, collection operations, index creation with name key and options, duplicate-key inspection, single-result not-found, cursor closure, cursor errors, and contexts.
- Study MongoDB compound sort and range pagination.
- Review RFC3339Nano encoding with millisecond precision, URL-safe encodings, stable-dataset cursor limitations, integration activation with unique disposable databases, and Project 061 email normalization.

## 17. Self-Assessment Questions

1. Why must an external ID be canonical before a driver call?
2. Why does update use explicit fields?
3. Why is the named unique index created and verified at startup?
4. Why is every injected clock value normalized to UTC millisecond precision, and why must tests prove truncation rather than round-trip?
5. Why is one extra document fetched?
6. What tuple defines strict continuation?
7. Which cursor failures are invalid input?
8. Why can concurrent updates invalidate stable pagination guarantees?
9. Why must domain results avoid BSON models?
10. Why is the owned integration database name validated before connection rather than asserting existence on the server?

## 18. Definition of Completion

- [ ] Official MongoDB driver is pinned exactly to `v2.8.0`.
- [ ] Document, normalization, immutable-field, millisecond-precision timestamp, and typed-error contracts are tested.
- [ ] The ascending unique email index named `uniq_users_email` is created and verified at startup; mismatch is startup failure.
- [ ] Limits, tuple order, one-extra pagination, and cursor validation are complete.
- [ ] Every cursor is closed and terminal cursor error is checked.
- [ ] Timestamp normalization to UTC millisecond precision is observable and tested.
- [ ] Unit tests pass without Docker, MongoDB, network, or environment variables, including the race detector.
- [ ] Tagged integration activation skips only when both values are absent and otherwise fails closed before any connection.
- [ ] Integration generates and validates a collision-resistant owned database name, uses only that database, and drops only that database through bounded cleanup without credential logging.
- [ ] Stable-dataset and unauthenticated-cursor limitations are documented.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, BSON, Compose content, or solution commands.

## 19. Optional Extensions

- Add projection-specific read models while preserving domain separation.
- Add an authenticated cursor format as a separately versioned compatibility exercise.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 066 — Database Transaction Manager](../../05-databases/066_db_transaction_manager/README.md#20-prerequisite-based-documentation-guide), [Project 061 — SQLite CRUD](../../05-databases/061_sqlite_crud/README.md#20-prerequisite-based-documentation-guide), [Project 041 — Context Timeout Example](../../03-concurrency/041_context_timeout_example/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`go.mongodb.org/mongo-driver/v2`](https://pkg.go.dev/go.mongodb.org/mongo-driver/v2), [`go.mongodb.org/mongo-driver/v2/bson`](https://pkg.go.dev/go.mongodb.org/mongo-driver/v2/bson).
- **Standards and concept references:** [MongoDB Go driver documentation](https://www.mongodb.com/docs/drivers/go/current/), [MongoDB indexes](https://www.mongodb.com/docs/manual/indexes/), [MongoDB cursor sorting](https://www.mongodb.com/docs/manual/reference/method/cursor.sort/).

### Project-specific learning focus

- **Learn now:** BSON types and time precision, ObjectIDs, unique indexes, duplicate-key errors, cursor cleanup, compound ordering, authenticated cursors, and disposable database tests.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
