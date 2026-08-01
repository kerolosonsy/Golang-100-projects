# Project 022 — Contact Book

## 1. Project Name and Number

Project **022** — `022_contact_book`. The directory name and number must match exactly. This project builds an in-memory contact book with create, read, update, delete, case-insensitive substring search, and JSON-file persistence. A small storage boundary isolates the in-memory logic from disk so behavior tests do not need a real file, while a separate JSON-file implementation preserves the safe-save pattern introduced in project 017.

## 2. Project Idea

A contact is a record with a positive integer ID, a non-empty name, an email, and a phone. The contact book is a small in-memory collection that assigns IDs on creation, supports lookup, update, and deletion, and supports a case-insensitive substring search across three fields. The book also persists itself to a JSON file.

Validation is intentionally modest. The contact book is a study project, not a production CRM. The validation rules below are enough to catch obvious mistakes (empty fields, missing `@`, phone numbers with letters) and to make the test cases deterministic. They are explicitly not a full RFC 5321/5322 implementation, and they do not attempt full international phone-number validation. The README pins this scope so the test can pin it too.

The contact book holds contacts in a stable, predictable order. Listings and search results are ordered by ascending ID. The next ID to assign is a counter that the book owns and persists; it is not derived from the stored IDs. Loading a JSON file with duplicate IDs, non-positive IDs, or a next-ID value that is not strictly greater than every stored ID is an error.

## 3. Why This Project Now?

Project 021 brought streaming parsing discipline. Project 022 revisits the CRUD-plus-persistence pair from project 017 with a different domain and adds three disciplines the previous pair did not need: case-insensitive substring search across multiple fields, a small storage boundary that lets the in-memory logic be tested without disk, and a persisted next-ID counter that is the source of truth rather than a derived value.

The project also forces the learner to think about validation deliberately and about the boundary between in-memory mutation and persistence deliberately. Pinning modest rules, and pinning the limits of those rules in writing, is the project's contribution to the learner's growing habit of "what does this code promise, and what does it explicitly not promise".

## 4. Prerequisites

Per the dependency map in `plan.md`, Project 022 formally requires:

- Completion of **021** (Log File Analyzer). The contact book is the next project after the log-file analyzer and is the first project in Level 2 that revisits the CRUD-plus-persistence pair with a search discipline.
- Completion of **014** (Input Validator). The contact book's validation rules — non-empty trimmed name, email containing exactly one `@` with non-empty parts and no whitespace, phone of digits plus allowed punctuation — are an exercise of the validator discipline established in Project 014. The modest scope of the validation rules (deliberately not a full RFC 5321/5322 implementation) is the "what does this code promise, and what does it explicitly not promise" habit that Project 014 established.
- Completion of **017** (JSON Todo Persister). The contact book persists itself to a JSON file using the narrow safe-save pattern from Project 017: write to a temporary file in the same directory, close with the close error checked, then rename over the target. The atomic-visibility-during-normal-operation rule applies.

Earlier projects (for example 016's stable-ID CRUD) provide review context already encountered and may inform the learner's design, but they are not additional required completions for this project.

No prior knowledge of HTTP, databases, generics, or concurrency.

## 5. What You Must Know Before Starting

- That `encoding/json` marshals and unmarshals Go structs using field names or `json:"..."` tags. The contact book's persisted document contains both the contacts and the next-ID value.
- That `json.Unmarshal` into a slice replaces the slice with the parsed contents. It does not merge with any pre-existing slice.
- That `os.CreateTemp` in a chosen directory produces a uniquely-named temporary file, and `os.Rename` replaces a target file when both paths are on the same filesystem. The project uses the narrow pattern from project 017: write to a temporary file in the same directory, close it (with a checked error), then rename over the target. The pattern provides atomic visibility during normal operation. Without explicitly syncing the temporary file and the parent directory, the pattern does not guarantee crash durability; a power loss can lose or reorder the newest persisted update.
- That `strings.ToLower` lowers a string rune by rune using Unicode lowercasing. It is not full locale-aware Unicode case folding; the Turkish dotless-I case is one well-known example where the two diverge.
- That `strings.Contains` reports whether a substring appears anywhere in a string. Combined with `strings.ToLower`, it gives a deterministic case-insensitive substring match.
- That a Go `map[uint64]Contact` (or similar ID-keyed map) is the natural in-memory representation for lookup by ID, but iteration order over a map is randomized. Any list-shaped output must be sorted by ID before printing.

## 6. Explanation of New Concepts

### The contact record

A contact has four fields:

- **ID.** A positive integer, unique within a single contact book instance. Assigned on creation by the contact book, never by the caller. The ID is just whatever value the book's next-ID counter holds at the moment of `Add`.
- **Name.** A non-empty string. The contact book trims surrounding whitespace before validating and rejects names that are empty after trimming.
- **Email.** A string that contains exactly one `@`, with a non-empty local part and a non-empty domain part, and no whitespace anywhere. This is deliberately modest: the rule is enough to catch obvious mistakes, and it is explicitly not a full RFC 5321/5322 implementation.
- **Phone.** A string consisting of digits and the allowed punctuation characters `+`, `-`, `(`, `)`, and spaces. The string must contain at least one digit. International formats are out of scope; the rule accepts a wide range of everyday formats without claiming correctness for any specific region.

### The persisted next-ID counter

The next ID to assign is a counter that the book owns. It is the source of truth for the next `Add` and is persisted alongside the contacts in the JSON file. The counter is not derived from the stored IDs and is not recomputed from a maximum on load or on `Add`.

Three rules together define the next-ID contract:

1. **Persisted exactly.** The JSON file contains both the contacts and the next-ID counter. Save writes both. Load reads both.
2. **Validated on load.** A loaded next-ID value must be strictly positive and strictly greater than every stored contact ID. A duplicate stored ID, a non-positive stored ID, a non-positive next ID, or a next ID that is not greater than every stored ID is a load error. The contact book is not usable until the load error is resolved (for example, by fixing the file or starting from an empty book).
3. **Incremented on commit.** `Add` reserves the current next ID for the new contact and increments the counter by one. The increment happens as part of committing the change, never as a pre-reservation.

Because the counter is the source of truth and is incremented only on commit, deleting a contact never changes the counter. The next `Add` receives the counter's current, greater value, which is the same value the counter held before the delete. The deleted ID is never reused. IDs are monotonically increasing and never reused across the lifetime of a single contact book instance. A loaded empty book whose contacts were deleted before save may have next ID `50` because the persisted counter was `50` when the last contact was deleted and the file was saved; the counter does not retroactively shrink. The contact book does not promise that every assigned ID is currently in use; it only promises that every currently-stored ID is unique and that the next ID to assign is the persisted counter value.

### Case-insensitive substring search

The search operates on three fields: name, email, and phone. The query is lowered with `strings.ToLower`. Each field is lowered with the same function. A match is reported if the lowered query appears as a substring in any of the three lowered fields. The result list is ordered by ascending ID. A search with no match returns an empty list, not an error.

The search is documented as Unicode-lowercase-based, not full Unicode case folding. The Turkish dotless-I case and similar locale-specific rules are not applied. Two strings that are visually identical but differ in precomposed vs decomposed form are not guaranteed to match. The README documents this caveat; the test does not assert on it.

### The small storage boundary

The contact book depends on a small storage interface with two methods, conceptually answering two questions: "give me the persisted contacts and the next-ID counter" and "replace the persisted contacts and the next-ID counter". The in-memory implementation returns empty results and next ID `1` until something has been added, and accepts any save without touching disk. The JSON-file implementation writes to a temporary file in the same directory, closes it, and renames over the target; on load it parses the file and validates both the ID contract and the next-ID invariant before returning.

The discipline this boundary enforces is testability. The contact book's CRUD and search logic is exercised through the in-memory store in unit tests, with no disk and no temporary directory. The JSON-file implementation is exercised in a separate set of integration tests against a per-test temporary directory, with the contact book's CRUD logic held fixed.

### Transactional mutation and persistence

Mutating operations (`Add`, `Update`, `Delete`) commit in two phases: build the candidate in-memory state, persist it through the store, then publish the in-memory state. The order is pinned:

1. **Validate.** Validate the candidate state without touching the store or the in-memory collection.
2. **Persist.** Call the store to write the candidate. If the store returns an error, abandon the candidate. The previous in-memory state and the previous persisted file remain visible.
3. **Publish.** Replace the in-memory state with the candidate. Return success.

A failed persist leaves the prior memory and the prior file intact. The caller sees the in-memory state matching the persisted file at every moment between operations; a partial publish is not observable.

For tests, the project requires an injected failure seam at the persistence boundary rather than read-only permission tricks. The JSON-file store is exercised through a small interface so a test can inject a store that returns a planned error from save and confirm that the prior in-memory state and the prior file are still visible after the failed save.

### The safe-save pattern reused

The JSON-file save follows the narrow pattern introduced in project 017: open a temporary file in the same directory as the target, write the document, close the file with the close error checked, then rename the temporary file over the target. On any failure before a successful rename, the previous target file is left untouched and the temporary file is removed. The pattern provides atomic visibility during normal operation: a successful rename replaces the previous target file in a single observable step. Without explicitly syncing the temporary file and the parent directory, the pattern does not guarantee crash durability; a power loss can lose or reorder the newest persisted update. The README documents this honestly; the test pins only the behavior the pattern guarantees during normal operation.

There is no copy fallback in the required scope and no remove-then-rename scheme. The rename is the atomic step; everything before it is fail-safe cleanup.

### Stable order

Every list-shaped result — listings, search results, and the results of any "all contacts" operation — is ordered by ascending ID. Two runs of the same operation against the same contact book return results in the same order.

### Missing and malformed storage

A missing storage file is not an error. It means the contact book starts empty with next ID `1`. A malformed JSON file is a hard error: the contact book does not silently overwrite the file with an empty book, and the caller is informed of the failure. A file with the right shape but with duplicate IDs, non-positive IDs, or a next-ID value that does not satisfy the persisted-counter invariant is also a hard error.

When load fails, the contact book's current in-memory state is unchanged. The failed load does not publish a partial in-memory state and does not touch the file.

## 7. Learning Objective

After completing this project the learner can:

- Define a small record type with four fields and pin the modest validation rules for each field in writing.
- Maintain an in-memory next-ID counter that is the source of truth, is persisted alongside the contacts, and is validated on load.
- Implement CRUD on an in-memory collection where mutating operations commit in two phases — persist first, then publish in-memory — so a failed persist leaves prior state intact.
- Implement case-insensitive substring search across multiple fields using the standard library's lowercasing, and document the limits of that approach.
- Introduce a small storage interface that isolates in-memory behavior from disk, and exercise both the in-memory and JSON-file implementations in their own tests.
- Reuse the narrow safe-save pattern from project 017 (temp file in same directory, checked close, rename over target) and document honestly that the pattern provides atomic visibility during normal operation but does not guarantee crash durability. Use the pattern without a copy fallback.
- Distinguish "missing file" (empty book) from "malformed file" (hard error) and treat each correctly. Distinguish "ID validation failed" from "next-ID invariant failed" and surface each correctly.
- Write tests that pin the next-ID policy, the validation rules, the search semantics, the JSON round trip, the transactional mutation contract, and the failure modes.

## 8. Functional Requirements

1. A contact is a struct with four fields: a positive integer ID, a non-empty name, an email, and a phone. The IDs are assigned by the contact book, not by the caller.
2. The contact book supports `Add`, `Get`, `Update`, `Delete`, `List`, `Search`, `Load`, and `Save`. `Add` accepts a contact without an ID and assigns the current next ID. `Get` returns the contact for a given ID, or an error if no contact has that ID. `Update` replaces the non-ID fields of an existing contact. `Delete` removes the contact by ID.
3. `Add` rejects a contact with empty name, empty email, empty phone, or any field that fails the modest validation rules.
4. The book owns a next-ID counter. The counter is the source of truth for the next `Add`. It is incremented on the commit of a successful `Add`. It is not derived from the stored IDs.
5. `Save` writes both the contacts and the current next-ID counter to the store. `Load` reads both. After a successful `Load`, the book's in-memory state exactly matches what was loaded: the contacts and the next-ID counter.
6. `Load` validates the persisted document before publishing it: every contact ID is positive and unique, and the next-ID counter is positive and strictly greater than every contact ID. A failed validation returns an error and leaves the current in-memory state unchanged. The file is not modified by a failed load.
7. Mutating operations commit in two phases: build the candidate in-memory state, persist it through the store, then publish the in-memory state. A failed persist returns an error and leaves both the prior in-memory state and the prior persisted file unchanged.
8. `List` returns every contact in ascending ID order. `Search` returns every contact whose name, email, or phone contains the query as a substring, in ascending ID order. The comparison is case-insensitive using `strings.ToLower` on both sides.
9. The contact book depends on a small storage interface with two conceptual operations: load (contacts + next ID) and save (contacts + next ID). The in-memory implementation supports behavior tests. The JSON-file implementation supports persistence tests against a temporary directory.
10. The JSON-file implementation reuses the narrow safe-save pattern from project 017: write to a temporary file in the same directory, close with the close error checked, then rename over the target. On any failure before the rename completes, the previous target file is left untouched and the temporary file is removed. The pattern provides atomic visibility during normal operation but does not guarantee crash durability; a power loss can lose or reorder the newest persisted update.
11. A missing storage file is not an error. The contact book starts empty with next ID `1`.
12. A malformed JSON file is an error. The contact book does not silently overwrite the file. The caller receives an error that identifies the parse failure or the validation failure.

## 9. Inputs and Outputs

### Inputs

- For the in-memory behavior tests: a fresh contact book (or a book pre-loaded with known contacts), a sequence of CRUD and search calls, and assertions on the returned contacts and errors.
- For the JSON-file tests: a per-test temporary directory, a JSON file produced by the encoder (or written by the test directly), and the contact book's load and save operations against that file.
- For the command-line integration: a JSON file path and CRUD/search commands on standard input or as command-line arguments. The exact CLI shape is the learner's choice; the test pins the learner's chosen shape.

### Outputs

- CRUD and search return contacts, lists, or errors. Errors carry enough context to identify the cause (for example, "missing name", "duplicate ID 5 in file", "next ID must be greater than every stored ID").
- The JSON file produced by the encoder contains every contact's ID, name, email, and phone, plus the next-ID counter.

### Example text-only search

```
$ contactbook search alice
1: Alice Adams  <alice@example.com>  +1-555-0101
7: Charlie Mallory-Alice  <cmallory@example.com>  +1-555-0107
```

### Example text-only validation error

```
$ contactbook add "" "not-an-email" "abc"
Error: name is empty.
```

### Example text-only load error

```
$ contactbook --file contacts.json list
Error: contact id 0 is not positive.
```

## 10. Rules and Edge Cases

- **Empty book.** A fresh contact book has zero contacts and next ID `1`. `List` returns an empty slice. `Search` for any query returns an empty slice.
- **Missing storage file.** A first `Load` against a path that does not exist is not an error. The book is empty with next ID `1`. The first `Save` creates the file.
- **Add with valid fields.** A contact with a non-empty trimmed name, an email containing exactly one `@` with non-empty parts and no whitespace, and a phone containing at least one digit plus the allowed punctuation is accepted and assigned the current next ID.
- **Add with empty name.** Rejected. The error message names the offending field.
- **Add with empty email.** Rejected.
- **Add with empty phone.** Rejected.
- **Add with whitespace-only name.** Rejected after trimming.
- **Add with email containing no `@`.** Rejected.
- **Add with email containing more than one `@`.** Rejected.
- **Add with email containing whitespace anywhere.** Rejected.
- **Add with phone containing letters or other disallowed punctuation.** Rejected.
- **Add with phone containing no digits.** Rejected.
- **Update on missing ID.** Rejected. The error names the missing ID.
- **Update on existing ID with valid fields.** The non-ID fields are replaced. The ID is preserved.
- **Update on existing ID with invalid fields.** Rejected; the stored contact is unchanged.
- **Delete on missing ID.** Rejected. The error names the missing ID.
- **Delete on existing ID.** The contact is removed. The next-ID counter is not modified.
- **Get on missing ID.** Rejected. The error names the missing ID.
- **Get on existing ID.** Returns the contact.
- **Search with no match.** Returns an empty slice, not an error.
- **Search with empty query.** Returns every contact, in ascending ID order.
- **Search results in stable order.** Two identical searches return results in the same order.
- **Next-ID source of truth.** The next ID to assign is the persisted counter, not a derived value. After `Add` assigns ID `1` and increments the counter to `2`, the book holds next ID `2` regardless of how many contacts are stored.
- **No ID reuse after delete.** After `Add` assigns IDs `1`, `2`, `3` and the counter is `4`, deleting contact `3` does not change the counter. The next `Add` produces a new contact with ID `4` (the counter's current value), not ID `3`. After `Delete` of contact `2`, the counter still reads `4`. After `Add`, the new contact has ID `4`, and the counter becomes `5`. The deleted ID `2` is never reused.
- **Loaded empty book with high counter.** A file that persisted an empty contact list with next ID `50` loads to a book with zero contacts and next ID `50`. The next `Add` assigns ID `50`. The counter does not retroactively shrink.
- **Load with duplicate ID.** Rejected. The error names the duplicate ID.
- **Load with non-positive ID.** Rejected. The error names the offending ID.
- **Load with non-positive next ID.** Rejected. The error names the offending next-ID value.
- **Load with next ID not greater than every stored ID.** Rejected. The error names the offending next-ID value.
- **Load with malformed JSON.** Rejected. The contact book's in-memory state is unchanged.
- **Save of empty current list.** Writes a JSON representation of an empty contact list plus the current next-ID counter (which may be `1`, may be `50`, or any positive value the book holds). The next-ID counter is not reset to `1` on save.
- **Save with failed persist.** The prior in-memory state and the prior persisted file remain intact. The caller sees the prior state and is informed of the failure.

## 11. Project Constraints

- Go standard library only. No third-party validation libraries, no regex helpers beyond the standard library, no JSON frameworks beyond `encoding/json`.
- Validation is the modest set pinned in section 10. It is not a full RFC 5321/5322 implementation, and it does not attempt full international phone-number validation.
- The next-ID counter is the source of truth. It is not derived from stored IDs at any point. It is persisted as-is and loaded as-is. It is validated on load against the rule "positive and strictly greater than every stored ID".
- The storage boundary is a small interface with two conceptual operations: load (contacts + next ID) and save (contacts + next ID). The interface is not a sprawling abstraction; it is the minimum needed to keep in-memory behavior tests off disk.
- The JSON-file implementation reuses the narrow safe-save pattern from project 017: write to a temporary file in the same directory, close with the close error checked, then rename over the target. On any failure before the rename, the previous file is untouched and the temporary file is removed. The pattern provides atomic visibility during normal operation but does not guarantee crash durability; no `fsync` is required. There is no copy fallback and no remove-then-rename scheme in the required scope.
- Mutating operations commit in two phases: persist through the store, then publish in memory. The transactional contract is verified by injected-failure tests, not by read-only permission tests.
- Search uses `strings.ToLower` on both the query and each field. Locale-aware case folding and Unicode normalization are out of scope; the README documents this caveat.
- Listing and search results are sorted by ascending ID before being returned.
- The contact book is not safe for concurrent use. Concurrency is introduced in later projects.

## 12. Design Questions Before Coding

- Where does the storage interface live? In the contact-book package, in a small `storage` subpackage, or in `main`? Which choice keeps the in-memory behavior tests free of disk dependencies while letting the JSON-file tests use a temporary directory?
- How is the next-ID counter represented? As a field on the contact book, as a value the store holds alongside the contacts, or both? Which choice makes "the counter is the source of truth" obvious in the code?
- How is the two-phase commit organized? As a private `commit(candidate)` method that persists then publishes, as separate `persist` and `publish` methods called in order, or as inline code in each mutating operation? Which choice keeps the transactional contract in one place?
- How is validation organized? As one `Validate` method, as per-field helpers, or as a small validator type? Which choice keeps the modest scope pinned in one place?
- How is the search implemented? Through three `strings.Contains` calls lowered on both sides, through a single lowered full-record string, or through a small match helper? Which choice keeps the three-field contract clear?
- How is the empty-query case handled? As a special case that returns all contacts, or as a natural consequence of "every field contains the empty string"? Which choice is clearest in the test?
- How is the load error surfaced? Through a typed error, through a wrapped error, through a sentinel, or through a result struct? Which choice lets the test assert on the cause (parse failure, duplicate ID, non-positive ID, non-positive next ID, next ID not greater than every stored ID)?
- How is the failure seam for the JSON-file store built? Through an interface boundary that the test substitutes with a failing store, or through a small `savable` interface that the JSON-file implementation wraps? Which choice keeps injected-failure tests deterministic without depending on read-only directories?
- How is the next-ID validation on load implemented? Through a single pass that walks the loaded contacts and the loaded counter together, or through separate validation steps? Which choice keeps the invariant check obvious in one place?

## 13. Implementation Milestones

1. Decide the package layout: a contact-book package that owns the contact record, validation, CRUD, search, and the two-phase commit; a small storage interface; an in-memory store implementation; and a JSON-file store implementation.
2. Define the contact struct with the four fields and the modest validation rules. Encode the rules as a small `Validate` method or per-field helpers.
3. Define the storage interface with two conceptual operations: load (contacts + next ID) and save (contacts + next ID). The interface's contract is "both values, together". Implement the in-memory store first; make it the default for unit tests.
4. Implement the contact book on top of the store. Track the in-memory contacts and the next-ID counter as fields. Implement `Add`, `Get`, `Update`, `Delete`, `List`, and `Search`.
5. Implement the two-phase commit. A private `commit` method persists the candidate state through the store, then publishes the in-memory state. Each mutating operation builds its candidate, calls `commit`, and returns the success or error. Validation happens before `commit`. The in-memory state and the persisted file stay consistent at every observable moment.
6. Implement `Search` with `strings.ToLower` on both the query and each field, returning matches in ascending ID order.
7. Implement the JSON-file store. Use `encoding/json` to encode and decode the document containing both the contacts and the next-ID counter. Reuse the narrow safe-save pattern from project 017: write to a temporary file in the same directory, close with the close error checked, then rename over the target. Clean up the temporary file on failure.
8. Wire load validation. After `json.Unmarshal`, walk the loaded contacts and the loaded next-ID counter together: every contact ID must be positive and unique; the next ID must be positive and strictly greater than every contact ID. Any failure returns an error and leaves the current in-memory state unchanged.
9. Wire the failure seam for the JSON-file store. Expose a small boundary the tests can substitute with a store that returns a planned error from save, so injected-failure tests can pin the two-phase commit without depending on read-only directories.
10. Wire the CLI. The exact CLI shape is the learner's choice; the integration tests pin the chosen shape against a per-test temporary file.
11. Add tests for every verification case in section 14, split between in-memory behavior tests and JSON-file persistence tests.

## 14. Verification Cases the Learner Must Write

Each case is described in natural language. In-memory tests use a fresh contact book or one pre-loaded with known contacts. JSON-file tests use a per-test temporary directory. Persistence failure tests use an injected store that returns a planned error from save.

### CRUD and validation

- `Add` with valid fields assigns ID `1`, `2`, `3` in three successive calls on a fresh book, and the counter holds `4` after the third `Add`.
- `Add` with empty name returns an error naming the name field. The book is unchanged and the counter is unchanged.
- `Add` with whitespace-only name returns an error naming the name field. The book is unchanged and the counter is unchanged.
- `Add` with empty email returns an error naming the email field. The book is unchanged.
- `Add` with email containing no `@` returns an error. The book is unchanged.
- `Add` with email containing two `@` returns an error. The book is unchanged.
- `Add` with email containing whitespace returns an error. The book is unchanged.
- `Add` with empty phone returns an error. The book is unchanged.
- `Add` with phone containing no digits returns an error. The book is unchanged.
- `Add` with phone containing letters returns an error. The book is unchanged.
- `Get` on an existing ID returns the contact.
- `Get` on a missing ID returns an error naming the ID.
- `Update` on an existing ID with valid fields replaces the non-ID fields and preserves the ID.
- `Update` on a missing ID returns an error naming the ID.
- `Update` with invalid fields returns an error; the stored contact is unchanged and the persisted file (if any) is unchanged.
- `Delete` on an existing ID removes the contact. `Get` on that ID afterwards returns an error. The next-ID counter is unchanged.
- `Delete` on a missing ID returns an error naming the ID.

### Next-ID policy

- After `Add` assigns IDs `1`, `2`, `3`, the counter holds `4`. After `Delete` of contact `2`, the counter still holds `4` and `Add` produces a new contact with ID `4`, then the counter holds `5`. The deleted ID `2` is never reused; the new contact is assigned the counter's current, greater value.
- After loading a file whose next-ID counter is `42` and adding one new contact, the new contact has ID `42` and the counter holds `43`.
- A file that persists an empty contact list with next ID `50` loads to a book with zero contacts and next ID `50`. The next `Add` assigns ID `50`.
- The next-ID counter is never derived from the stored IDs at any point. Adding a contact does not recompute the counter from the maximum stored ID; it uses the current counter value.

### Load validation

- A file with a duplicate stored ID (two records with ID `5`) returns a load error naming the duplicate ID. The in-memory state is unchanged.
- A file with a non-positive stored ID (a record with ID `0` or `-1`) returns a load error naming the offending ID. The in-memory state is unchanged.
- A file with a non-positive next ID (counter `0` or `-1`) returns a load error naming the offending next-ID value. The in-memory state is unchanged.
- A file whose next ID is not strictly greater than every stored ID (for example, counter `5` when a contact has ID `7`) returns a load error naming the offending next-ID value. The in-memory state is unchanged.

### Listing and search

- `List` on an empty book returns an empty slice.
- `List` on a book with contacts `3`, `1`, `2` returns them in order `1`, `2`, `3`.
- `Search` with a query that matches one contact's name returns that contact, in ascending ID order.
- `Search` with a query that matches a phone substring returns the contacts whose phones contain the substring.
- `Search` is case-insensitive: `Alice`, `ALICE`, and `alice` all match. `Search` with mixed case in the field also matches.
- `Search` with no match returns an empty slice, not an error.
- `Search` with an empty query returns every contact, in ascending ID order.
- Two identical searches against the same book return results in the same order.

### JSON persistence

- Save an empty book (next ID `1`) to a temporary directory, then load it. The book is empty with next ID `1`.
- Save a book with three contacts and next ID `4`, load it, and verify the three contacts plus next ID `4`.
- Save a book whose current contact list is empty but whose next-ID counter is `50`, load it, and verify the next-ID counter is `50` and the contact list is empty. The counter is not reset on save.
- Round-trip: save, load, save again. The two saved documents are semantically equivalent (the contacts and the next-ID counter match). Whitespace or representation differences the encoder is allowed to make are normalized by the test.

### Transactional mutation

- A test injects a store that returns a planned error from save. `Add` returns the error. The prior in-memory state is unchanged and the prior persisted file (if any) is unchanged.
- A test injects a store that returns a planned error from save after several successful operations. `Update` returns the error; the contact's stored fields and the persisted file are unchanged.
- A test injects a store that returns a planned error from save. `Delete` returns the error; the contact is still present in memory and the persisted file still records it.
- After a failed injected save, the next `Add` succeeds and uses the same next-ID counter the book held before the failed save.

### Failure modes

- A missing JSON file at load time is not an error; the book is empty with next ID `1`.
- A malformed JSON file at load time returns a parse error; the contact book's in-memory state is unchanged and the file is not overwritten.
- A JSON file with duplicate IDs returns an error naming the duplicate ID; the in-memory state is unchanged and the file is not overwritten.
- A JSON file with a non-positive ID returns an error naming the ID; the in-memory state is unchanged and the file is not overwritten.
- A JSON file with a non-positive next ID returns an error naming the offending next-ID value; the in-memory state is unchanged and the file is not overwritten.

### Process

- An integration test runs the compiled CLI against a temporary JSON file with a small set of known contacts and exercises `list` and `search` to confirm exit code zero and stable output.
- An integration test runs the compiled CLI against a malformed JSON file and confirms the exit code is non-zero and standard error names the parse failure.

## 15. Common Mistakes to Watch For

- **Building a sprawling storage abstraction.** The storage interface is two conceptual operations, not a generic persistence framework. The pattern from project 017 is reused; it is not duplicated into a full library.
- **Deriving the next-ID counter from the stored IDs.** The counter is the source of truth and is not derived at any point. Computing `max(stored IDs) + 1` or `len(contacts) + 1` is wrong because it silently reuses deleted IDs after a delete and produces a counter that contradicts the persisted file.
- **Treating `len(contacts)+1` as the next ID.** This is wrong because deletes invalidate the assumption. The counter is a separately-tracked value.
- **Failing to validate the next-ID counter on load.** A non-positive counter or a counter that is not strictly greater than every stored ID must be rejected. Treating it as "good enough" lets a corrupted file silently change the next assignment.
- **Silently overwriting a malformed file.** A malformed file is an error. The contact book must not save an empty book over the user's data.
- **Treating a missing file as an error.** A missing file is empty-state. The contact book starts fresh.
- **Publishing in-memory state before the persist succeeds.** The two-phase commit pins persist-then-publish. A `Save` that fails must not leave the in-memory state advanced beyond the persisted file.
- **Lowercasing the field instead of the query, or vice versa.** The search must lowercase both sides. Lowercasing only the query misses matches where the field has uppercase letters.
- **Lowercasing byte-by-byte.** Lowercasing must operate on runes. `strings.ToLower` does this. A byte-by-byte loop produces wrong results for non-ASCII text.
- **Conflating locale-aware case folding with `strings.ToLower`.** The project uses `strings.ToLower`. Locale-specific rules (Turkish dotless-I, etc.) are out of scope; the README documents this.
- **Sorting search results by frequency or by first match.** The contract is ascending ID. Any other order violates the stable-order requirement.
- **Validating after assigning the ID.** Validation must happen before any state mutation. An `Add` that validates after assigning the ID leaves a half-formed contact in the book on failure.
- **Writing directly to the target file during save.** A direct write to the target file is not the safe-save pattern from project 017. The pattern is temp-file in same directory → write → checked close → rename over target.
- **Mandating `fsync` as universal durability.** The pattern provides atomic visibility during normal operation but does not guarantee crash durability. No `fsync` is required, and a copy fallback is not in scope. The README documents the limits honestly.
- **Resetting the next-ID counter on `Save` of an empty list.** A saved empty list must persist the current next-ID counter, not always `1`.
- **Sharing one temporary directory across tests.** Each test must use its own temporary directory. Sharing paths causes cross-test pollution.
- **Testing save failure with a read-only directory.** Permission tricks are flaky and platform-sensitive. Use an injected store that returns a planned error from save; the test pins the two-phase commit deterministically.
- **Claiming full RFC 5321/5322 or full E.164 validation.** The validation is modest. The README pins the limits; over-claiming violates the contract.
- **Using regular expressions where simple string rules suffice.** The validation rules are simple character and substring checks. A regex obscures the modest scope and is harder to test deterministically.

## 16. Topics and References for Study

- A Tour of Go: "Methods", "Interfaces", "Errors".
- Effective Go: "Data", "Errors", "Packages".
- Package documentation: `encoding/json` (`Marshal`, `Unmarshal`, `MarshalIndent`, `Encoder`, `Decoder`), `strings` (`ToLower`, `Contains`, `TrimSpace`, `IndexByte`), `strconv` (`Itoa`, `Atoi`), `os` (`CreateTemp`, `Rename`, `ReadFile`, `Stat`), `errors` (`Is`, `As`, `New`).
- Persistence patterns: search for "Go safe-save temp file rename", "Go JSON round trip semantic equality", "Go transactional two-phase commit".
- Validation scope: search for "Go modest email validation", "Go phone number validation scope", "Go Unicode ToLower caveats".

## 17. Self-Assessment Questions

1. Why is the next-ID counter the persisted source of truth rather than derived from stored IDs (and never reset on save of an empty list), and what does that contract pin about deletes?
2. Why is a missing storage file treated as empty state while a malformed file is treated as a hard error?
3. Why is `strings.ToLower` documented as "not full Unicode case folding" rather than as "case-insensitive"?
4. Why must the storage interface stay small and keep the in-memory store as the default for unit tests, and what does that separation buy the JSON-file tests?
5. Why is validation done before any state mutation, and what would break if `Add` validated after assigning the ID?
6. Why must mutating operations commit in two phases (persist, then publish), and what does the injected-failure test pin?
7. Why must the JSON-file implementation reuse the narrow safe-save pattern from project 017 rather than writing directly to the target file, and why does the pattern provide atomic visibility during normal operation but not guarantee crash durability?
8. Why must `List` and `Search` results be sorted by ascending ID rather than by insertion order or by map iteration order?
9. Why are the modest validation rules pinned in writing rather than left to "common sense"?
10. Why must a contact book reject loading a file whose next-ID counter is not strictly greater than every stored ID, even if the JSON parses correctly?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every functional requirement in section 8 is satisfied.
- Every verification case in section 14 has a corresponding test, with in-memory tests and JSON-file tests separated.
- The next-ID counter is the source of truth. It is not derived from stored IDs at any point. It is persisted as-is and loaded as-is. It is validated on load against the rule "positive and strictly greater than every stored ID". The contact book does not promise that every assigned ID is currently in use.
- A missing storage file is not an error; a malformed JSON file is a hard error. A failed load leaves the current in-memory state unchanged.
- The JSON-file implementation uses the narrow safe-save pattern from project 017 (temp file in same directory, checked close, rename over target). The pattern provides atomic visibility during normal operation but does not guarantee crash durability; no `fsync` is required; there is no copy fallback and no remove-then-rename scheme.
- Mutating operations commit in two phases: persist, then publish. A failed persist leaves prior in-memory state and prior persisted file unchanged. The injected-failure test pins this deterministically.
- `List` and `Search` return results in ascending ID order with no other ordering.
- Validation happens before any state mutation. A rejected `Add` or `Update` leaves the book unchanged.
- The package documentation states the modest validation rules, the next-ID policy, the search semantics, the storage boundary, the two-phase commit contract, and the limits of `strings.ToLower`.
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Group-by initial letter.** Add a `Groups` operation that returns contacts grouped by the first letter of their name, with each group ordered by ascending ID and the groups themselves ordered alphabetically. The group operation does not mutate the book. Do not add further breakdowns (per-domain, per-phone-prefix) or a count summary per group.
- **Backup file on save.** Configure the JSON-file store to keep a single rolling backup (for example, `contacts.json.bak`) updated alongside each successful save. The backup is overwritten by the same safe-save pattern and is not consulted by `Load`. Do not add a backup history, restore command, or multi-file backup scheme.
