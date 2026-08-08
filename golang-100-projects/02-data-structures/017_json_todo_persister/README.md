# Project 017 — JSON Todo Persister

## 1. Project Name and Number

- Project **017** — `017_json_todo_persister`.
- The directory name and number must match exactly.
- This project extends the in-memory todo session from project 016 with a JSON file store.

## 2. Project Idea

The todo session gains a single JSON file as its source of truth. On startup the program loads that file into the in-memory collection built in project 016; after every successful state change the program saves the updated collection back to the same file. The session itself is unchanged from 016: the user types one command per line into an injected reader, and the program processes the session until EOF or `quit`.

Two storage policies are pinned by the README. First, a missing data file means an empty collection, not a fatal error: the program treats the first run and a deleted file identically. Second, saving is atomic in the intended single-filesystem case: the program encodes to a temporary file in the destination's directory, finishes writing and closes it successfully, and only then renames it over the destination with `os.Rename`. On platforms or filesystems where `os.Rename` cannot replace an existing destination, the rename step returns an error; the program reports it and the previous valid destination is preserved. There is no remove-then-rename fallback, because that approach opens a window in which the destination is missing.

The on-disk JSON document explicitly carries the next-ID value as its own field. The next-ID is not derived from the loaded tasks; if it were, deleting the largest task before saving and then reloading would let the program reuse that ID, which violates the project 016 rule that an issued ID is never reused.

## 3. Why This Project Now?

- Project 016 built a stable in-memory domain with stable IDs and a deterministic order.
- That stability is exactly what makes a round trip on disk meaningful: the JSON file must preserve IDs, titles, completion flags, and order so that two consecutive sessions behave like one longer session.
- If 016's ID policy were to reuse IDs after deletion, a round trip on disk would silently change the meaning of an old ID and the project would lose its determinism.

- This project also introduces the discipline of "atomic file replace".
- A program that opens the destination for writing and crashes midway leaves a half-written file on disk; the next run loads a malformed JSON and either fails to start or silently overwrites valid data with an empty collection.
- The pattern this project introduces — write to a temporary file in the same directory as the destination, close it, then rename — is the simplest practical defense against partial writes, and is the foundation that later projects build on (CSV parser I/O, file organizer moves, contact book storage).

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. The map also lists `017` as the reference project for any later project that needs file storage. Project 017 therefore requires:

- Completion of **016** (Todo CLI), including the in-memory domain, the ID stability rules, and the read-print loop driven by an injected reader.
- No prior knowledge of HTTP, databases, generics, or concurrency.
- Familiarity with `encoding/json` is helpful but not required; this project introduces it.

## 5. What You Must Know Before Starting

- The structure of a JSON object: a map from string keys to typed values, and the corresponding Go representation as a `map[string]...` or as a struct with `json:` tags.
- That `encoding/json` is a streaming codec: it reads from an `io.Reader` and writes to an `io.Writer`, one token at a time, and never loads the whole document into memory unless the caller asks for it.
- The difference between "the file does not exist" (a missing-path error) and "the file exists but its contents are not valid JSON" (a parse error). The project treats these two cases very differently.
- That a same-filesystem rename provides atomic name replacement only where the operating system and filesystem support replacing the destination through `os.Rename`. Cross-filesystem renames return an error rather than falling back to copy-and-delete; platforms that cannot replace an existing destination also return an error.
- That a temporary file created in the same directory as the destination is renamed with the same atomicity as the destination itself: same directory, same filesystem, same rename call.
- That `os.WriteFile` is not atomic and must not be used for the save path. The save path writes to a temporary file and renames.
- That `defer` runs on the function's return path, and that closing a file inside a `defer` releases the file handle but does not roll back a partial write.

## 6. Explanation of New Concepts

### Concepts

#### A small storage interface

The project pins a single seam between the domain from 016 and the file: a storage interface with two methods, `Load` and `Save`. The domain layer does not know whether the implementation is JSON, an in-memory map, a database, or a network service. Tests substitute a fake that holds the collection in memory; the program wires a JSON file implementation.

The shape of that interface is the learner's choice. The contract is:

- `Load` returns the stored collection plus its next-ID value, an empty collection with a nil error if there is no file, or an error if the file exists but cannot be parsed.
- `Save` writes the collection plus its next-ID value to the destination. If the write, the close, or the rename step fails, the previous valid destination file is left untouched and the temporary file is removed.

#### JSON encoding choices

The on-disk representation is a single JSON object with two top-level fields: an array of task objects, and a numeric field that holds the next-ID value. Each task object carries the four required fields. The exact field names are the learner's choice and must be documented in the package documentation; the README recommends short, unambiguous names.

The next-ID field is mandatory. The program does not derive the next ID from the loaded tasks. Derivation looks attractive but is unsafe: if a task with the largest issued ID is deleted before the next save, the saved file contains no task with that ID, and a fresh load followed by an `add` would derive a next-ID that equals the deleted ID. That collision violates the project 016 rule. The on-disk JSON carries the next-ID explicitly so the rule survives the round trip.

#### Save: write, close, then rename

The save path is a sequence of five steps:

- Choose a temporary file name in the destination's directory. The name must be unique to the current save attempt.
- Encode the collection plus its next-ID value to that temporary file.
- Close the temporary file explicitly. Closing releases the file handle and pushes any buffered bytes to the kernel. A later step assumes the bytes are visible to the rename.
- Rename the temporary file over the destination path with `os.Rename`.
- If any step before step 4 fails, do not rename; remove the temporary file if it exists, and return the error to the caller. If step 4 fails, return the error; the previous destination file is left unchanged.

The temporary file lives in the destination's directory so the rename stays on a single filesystem. A temporary file in a different directory crosses a filesystem boundary and `os.Rename` returns an error rather than performing the rename.

#### What "atomic" means and what it does not mean

The save path's atomicity guarantee is narrow and honest. It is: on a single filesystem, `os.Rename` replaces the destination's directory entry without touching the file's contents. If the program is killed after step 4, the destination is the new file in full. If the program is killed before step 4, the destination is the old file in full. There is no moment at which the destination holds a partially written file.

This is not the same as crash durability. Before the rename, the save path has not intentionally modified the old destination. After the rename, a sudden power loss can still lose the new directory entry or newly written data unless the relevant file and directory are synchronized according to the platform's rules. The README therefore does not promise that the new bytes have reached stable storage; it promises only the atomic-visibility behavior supplied by a successful supported rename during normal operation. An optional extension in section 19 discusses durability without claiming to solve it universally.

#### Why a remove-then-rename fallback is not used

On platforms where `os.Rename` cannot replace an existing destination, an obvious alternative is "remove the destination, then rename the temp into place". That approach opens a window in which the destination does not exist on disk. A crash or concurrent read in that window sees a missing file instead of the last valid JSON, which is worse than the previous-valid-file outcome that `os.Rename`'s error path preserves. The README pins the policy: report the rename error and leave the previous destination file unchanged.

#### Missing file vs malformed file

The two failure modes are very different and must be handled very differently:

- **Missing file.** `os.Stat` returns an error of the "does not exist" kind. The program treats this as "no tasks yet" and starts with an empty collection and a next-ID of `1`. This is not a fatal error.
- **Empty existing file.** The file exists and has zero bytes. The program must distinguish this from "missing file" and treat it as an error: silently erasing data is worse than refusing to start.
- **Malformed JSON.** The file exists and has bytes but the JSON decoder reports a syntax error. The program treats this as an error and does not overwrite the file. The user must decide how to repair it.

A test pins all three cases.

## 7. Learning Objective

After completing this project the learner can:

- Define a small storage seam with `Load` and `Save` and substitute a fake implementation in tests.
- Encode a struct collection plus its next-ID value to JSON and decode it back without losing field values, IDs, order, or the next-ID value.
- Distinguish a missing file from an empty existing file from a malformed file, and apply the right policy to each.
- Implement a save path that writes to a same-directory temporary file, closes it, and renames it over the destination, with cleanup on failure.
- Recognize that `os.Rename` does not fall back to copy-and-delete across filesystems and report a rename error rather than silently replacing the destination.
- Reuse the domain layer from project 016 unchanged, with the storage layer as the only addition.
- Write tests that use a per-test temporary directory and never touch the user's real home or working directories.

## 8. Functional Requirements

1. The program loads its in-memory collection plus its next-ID value from an explicit file path supplied to the application. Tests always supply a path inside a per-test temporary directory; the program does not guess a home-directory or binary-relative location.
2. The session itself is unchanged from project 016: the program reads command lines from an injected reader until EOF or `quit`. The four subcommands `add`, `list`, `complete`, `delete` keep their semantics.
3. Treat a domain change and its save as one visible application operation: build the candidate state, save it, and only then publish it in the running session and print success. If saving fails, report the error and keep both the prior in-memory state and the prior destination as the visible state.
4. `Load` returns an empty collection and a next-ID of `1` if the file does not exist. It returns an error if the file exists and is empty, or if the file exists and contains malformed JSON, or if the file exists and contains valid JSON that does not match the expected schema.
5. The on-disk JSON document explicitly carries the next-ID value. The program does not derive the next ID from the loaded tasks.
6. `Save` encodes the collection plus its next-ID value to a temporary file in the destination's directory, closes it successfully, and only then renames it over the destination.
7. If any step of `Save` before the rename fails, the destination file is left unchanged, the temporary file is removed, and the error is reported.
8. If the rename step fails, the previous valid destination file is left unchanged. The program reports the rename error and does not attempt to remove the destination or to copy the temporary file into place.
9. The round-trip behavior must preserve IDs, titles, completion flags, the order of tasks, and the next-ID value. After a save followed by a load, an `add` issues an ID strictly greater than every ID that was loaded.
10. Loading validates domain invariants as well as JSON syntax: every task ID is positive and unique, and the stored next-ID value is positive and strictly greater than every stored task ID. An empty task array may legitimately carry a next-ID greater than `1` because deleted IDs remain part of its history.
11. The same domain layer from 016 is reused without modification. Tests call the domain layer directly with a fake storage and assert on its behavior.
12. All tests use a per-test temporary directory created by the testing framework, never the repository, the home directory, or any other real directory.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- A JSON file at the configured path. Four cases:
  - **Missing.** The file does not exist on disk.
  - **Empty.** The file exists but has zero bytes.
  - **Valid.** The file exists and contains a JSON object with the documented fields, including the next-ID value.
  - **Malformed.** The file exists and contains bytes that the JSON decoder rejects, or contains JSON that does not match the expected schema.

#### Outputs

- For `Load`: an in-memory collection of tasks plus a next-ID value plus a nil error in the missing case; the same plus a nil error in the valid case; an error in the empty, malformed, and schema-mismatch cases.
- For `Save`: nil on success; an error on any failure, with the destination file unchanged and the temporary file cleaned up.
- For each subcommand, the same line on standard output that 016 produced.

#### Example text-only success session (two sessions)

First session:

```
add buy milk
list
quit
```

Expected output:

```
Added task 1: buy milk
1 task:
[ ] 1: buy milk
```

The on-disk JSON after this session contains one task object and a next-ID value greater than the largest issued ID.

Second session, with the same JSON file present:

```
list
add write report
quit
```

Expected output:

```
1 task:
[ ] 1: buy milk
Added task 2: write report
```

The new task is ID `2`, not ID `1`, because the loaded next-ID value from the file is honored.

#### Example text-only failure cases

```
add buy milk
```

If the save fails because the rename step returned an error, the session writes:

```
Error: could not save data file: <rename error>.
```

The destination file is unchanged after this error: the next session loads the previous valid collection. The temporary file is removed.

## 10. Rules and Edge Cases

- **Missing data file.** `Load` returns an empty collection and a next-ID of `1`. The session starts normally and accepts new tasks.
- **Empty existing data file.** `Load` returns an error. The session does not start with an empty collection, because doing so would silently erase the user's data.
- **Malformed JSON.** `Load` returns an error. The session does not overwrite the malformed file.
- **Schema mismatch.** A JSON document whose top-level value is not an object with the expected fields is rejected with an error.
- **Missing next-ID field.** A JSON document whose top-level object lacks the next-ID field is rejected with a schema-mismatch error. The program does not silently default to a derived value.
- **Pre-rename failure.** A failure in the encode, write, or close step leaves the destination file untouched, removes the temporary file, and returns the error.
- **Rename failure.** A rename error leaves the previous valid destination file untouched and reports the error. The program does not attempt to remove the destination or to copy the temporary file into place.
- **Stale temporary file.** A leftover temporary file in the destination directory from a previous failed save must not interfere with the next save. The next save uses a fresh unique name; the next `Load` ignores any file that is not the destination path.
- **Cross-filesystem destination.** `os.Rename` returns an error rather than copying the file. The program reports the error and leaves the destination file unchanged. The README does not promise cross-filesystem atomicity.
- **Cross-device directory entry.** Even when the source path and destination directory are the same root, nested mount points can make the rename fail. The program reports the error and leaves the destination unchanged.
- **Concurrency.** Out of scope. The program assumes a single process holds the data file.
- **Permissions.** Out of scope. The program assumes the destination directory is writable. A failure caused by an unwritable directory is reported through the same pre-rename or rename error path.
- **Schema migration.** Out of scope. The program does not upgrade an old file format to a new one.
- **Encryption, signing, compression.** Out of scope. The file is plain JSON.

## 11. Project Constraints

- Go standard library only. No third-party JSON, no third-party file libraries.
- The save path uses `os.CreateTemp` (or the learner's chosen equivalent) inside the destination directory. The temporary file name is unique to the save attempt.
- The destination file is replaced only by an explicit `os.Rename` after the temporary file is closed. The program does not remove the destination before renaming; that approach opens a missing-file window.
- The on-disk JSON document explicitly carries the next-ID value. Derivation is not an option.
- The domain layer from project 016 is reused without modification. Tests call the domain layer directly with a fake storage; integration tests run the compiled binary against a temporary JSON file.
- No locking across processes. A second process touching the same file can corrupt it; this is documented in the package documentation as out of scope.
- No schema migration, no encryption, no database. Plain JSON file in the destination directory.

## 12. Design Questions Before Coding

- How is the storage interface shaped? Two methods (`Load`, `Save`)? Three (`Load`, `Save`, `Close`)? Which choice lets the test substitute a fake cleanly?
- Where does the destination path come from? A flag, an environment variable, a constant next to the binary? Which choice makes integration tests reliable?
- How is the next-ID value represented on disk? As a separate top-level field with a fixed name, as part of each task object, or as a sentinel inside the collection? Which choice survives a round trip and is easy to test?
- How is the temporary file named? A fixed suffix in the same directory, a randomized name, or a UUID-style identifier? Which choice avoids clashes when two saves race?
- How is "empty file" distinguished from "missing file"? Through `os.Stat` errors, through reading and checking length, or through a sentinel? Which choice is clearest to read?
- How is the failure path of `Save` reported? As a typed error, as a wrapped error with `fmt.Errorf`, or as a sentinel? Which choice lets the test assert on the cause without coupling to the wording?
- How does the test confirm that a failed save left the destination untouched? By reading the file's bytes before and after, by checking its modification time, or by loading it through the domain layer? Which choice pins the invariant?
- How does the test simulate a rename failure where practical? Through an injected move boundary that the storage layer calls, through a destination directory made unwritable by the test, or by leaving the rename-failure case marked as "where practical" and pinning the bytes invariant through a pre-rename failure instead?

## 13. Implementation Milestones

1. Re-read the domain layer from project 016 and confirm the four fields per task. Decide the on-disk JSON field names, including the explicit next-ID field, and write them down in the package documentation.
2. Define the storage interface with `Load` and `Save`. Decide the error type returned by `Load` for each of the failure modes (empty, malformed, schema mismatch, missing next-ID field).
3. Implement the JSON file storage. `Load` checks the file's existence and size before decoding. `Save` writes to a temporary file, closes it, and renames it over the destination. A failure before the rename removes the temporary file and returns the error. A rename error leaves the destination file untouched and returns the error.
4. Wire the session: load the collection plus its next-ID value once at startup, then enter the same loop as 016. Apply a command to a candidate state, save that candidate, and publish it in memory and print success only after the save succeeds. A failed save leaves the prior session state visible.
5. Handle the missing-file case in `Load`: empty collection, next-ID `1`, no error. Reject empty or malformed files, missing required fields, duplicate or non-positive task IDs, and a next-ID value that is non-positive or not greater than every stored task ID.
6. Add cleanup to `Save`: if the encode, write, close, or rename step fails, remove the temporary file if it exists and return the error without touching the destination.
7. Add the integration tests that run the compiled binary against a per-test JSON file in a temporary directory. Pin the round-trip, missing, empty, and malformed cases.
8. Add the domain tests that drive the domain layer with a fake storage. Pin the round-trip, the empty-file failure, the malformed-file failure, the missing-next-ID failure, and the failed-save cleanup.

## 14. Verification Cases the Learner Must Write

### Required Cases

Each case is described in natural language. All tests use the testing framework's per-test temporary directory; no test touches the user's real files.

#### Domain with fake storage

- `Save` then `Load` returns the same collection plus the same next-ID value: same length, same IDs in the same order, same titles, same completion flags.
- After `Save` then `Load`, an `add` issues an ID strictly greater than every ID that was loaded. The next-ID returned by `Load` matches the value the program wrote.
- A `Save` that fails inside the fake (for example, a writer that returns an error on `Close`) does not change the next `Load`'s output. The fake records that no temporary file was left behind.
- A session command whose candidate state cannot be saved reports the save error, does not print a success confirmation, and leaves the current in-memory collection and next-ID value unchanged.

#### File storage — happy paths

- A fresh temporary directory with no JSON file: `Load` returns an empty collection, a next-ID of `1`, and a nil error.
- After `add` against a fresh directory, the JSON file exists, is non-empty, is parseable, and contains one task object plus a next-ID value strictly greater than the largest issued ID.
- After `add`, `add`, `complete <id1>`, `delete <id2>`, the JSON file contains one task plus a next-ID value strictly greater than the largest ID the session ever issued (including the deleted ID).

#### File storage — error paths

- An empty existing file: `Load` returns an error, the program does not start with an empty collection, and the file's bytes are unchanged after the failed load.
- A file containing the literal text "not json": `Load` returns an error, and the file's bytes are unchanged.
- A file containing a JSON value that is not an object (for example, an array): `Load` returns an error.
- A file containing a JSON object that lacks the next-ID field: `Load` returns an error. The program does not silently default to a derived next-ID.
- A file containing a JSON object whose task array contains entries that miss required fields: `Load` returns an error.
- A task array with duplicate or non-positive IDs is rejected. A non-positive next-ID, or a next-ID not greater than every stored task ID, is rejected.
- A save that fails before the rename: `Load` on the destination path returns the previous valid collection, the destination file's bytes are unchanged, and no leftover temporary file remains in the directory.

#### Atomic save

- A test asserts that the destination file's bytes are equal to the previous valid bytes after a failed save. This is the central invariant for the pre-rename failure path.
- A test asserts that no file matching the temporary file's name pattern remains in the destination directory after a failed save.
- Where practical, a test injects a rename failure through a controllable seam (for example, a destination path that the test forces to fail the rename) and asserts that the previous valid bytes are unchanged, the rename error is reported, and the temporary file is removed. Tests that cannot reliably force a rename failure without permission tricks mark this case as "where practical" and pin the bytes invariant through a pre-rename failure instead.

#### Idempotence

- A test calls `Save` twice with the same collection. The destination file's contents are identical after both calls, and no leftover temporary file appears.
- A test calls `Load` twice without any intervening save. Both calls return the same collection plus the same next-ID value.

#### Next-ID round trip

- A test runs `add`, saves, then loads and confirms that the exact persisted next-ID value is preserved and is greater than the issued ID.
- A test runs `add a`, `add b`, `delete 2`, saves, then loads and runs `add c`. The new task's ID is strictly greater than `2`, even though `2` no longer exists in the loaded tasks.
- A test loads a file whose next-ID field equals `5` and whose task array is empty, then runs `add`. The new task's ID is `5`.

## 15. Common Mistakes to Watch For

- **Deriving the next ID from the loaded tasks.** A file whose largest task was deleted before saving would, after load and add, issue that deleted ID again. The project requires the next-ID value to be carried explicitly.
- **Using `os.WriteFile` for the save path.** That opens the destination for writing and truncates it immediately; a failure midway leaves a truncated file in place of the last valid one.
- **Creating the temporary file in `/tmp` or another directory.** `os.Rename` returns a cross-filesystem error rather than falling back to copy-and-delete. The temporary file must live in the destination's directory so the rename stays on one filesystem.
- **Treating a missing file and an empty file as the same case.** Empty means "the file exists and has zero bytes"; missing means "the file does not exist". Conflating the two erases data.
- **Overwriting the destination on a malformed-file `Load`.** The malformed file must remain on disk for the user to repair. The program reports the error and refuses to start.
- **Removing the destination before renaming.** A remove-then-rename fallback opens a window in which the destination is missing. The README pins the policy: report the rename error and leave the destination unchanged.
- **Forgetting to remove the temporary file on a pre-rename failure.** A leftover temporary file accumulates across runs and may be picked up by a careless implementation.
- **Closing the temporary file twice.** Closing twice returns an error on most platforms. Closing once in a `defer` and once explicitly is a common bug.
- **Using a single shared test directory across tests.** A failure in one test leaks state into the next. Each test must use its own temporary directory.
- **Reading the destination path's bytes inside the program to verify atomicity.** Tests should verify atomicity by checking that the destination file's contents match the previous valid bytes; reading them inside the program couples the test to the implementation.
- **Forcing permission-based rename failures in tests.** Changing permissions on a directory is flaky, depends on the test runner's user, and is unsafe to assume. Use an injected move boundary where practical; mark the case as "where practical" otherwise and pin the bytes invariant through a pre-rename failure instead.

## 16. Topics and References for Study

- A Tour of Go: "Methods and interfaces", "Reading and writing JSON".
- Effective Go: "Data", "Errors".
- Package documentation: `encoding/json` (`Marshal`, `Unmarshal`, `Encoder`, `Decoder`, `MarshalIndent`), `os` (`Stat`, `OpenFile`, `CreateTemp`, `Rename`, `ReadFile`, `WriteFile`), `errors` (`Is`, `As`), `io` (`Reader`, `Writer`, `Closer`).
- File replacement patterns: search for "Go atomic file write", "temp file rename", "POSIX rename atomicity".
- JSON design choices: search for "Go json struct tags", "versioned JSON documents", "JSON schema migration".

## 17. Self-Assessment Questions

1. Why does the project distinguish between "missing file" and "empty file", and why is the latter a hard error?
2. Why must the next-ID value be carried explicitly in the on-disk JSON, and what would go wrong if it were derived from the loaded tasks?
3. Why does the temporary file have to live in the destination's directory?
4. Why is `os.WriteFile` not acceptable for the save path, and what does `os.Rename` provide that `os.WriteFile` does not?
5. Why is a remove-then-rename fallback not used when `os.Rename` returns an error, and what window does that fallback open?
6. What does the test pin about the destination's bytes after a failed save, and why is that the central invariant?
7. How does the storage layer's `Load` return different errors for empty, malformed, schema-mismatch, and missing-next-ID cases, and how does the test distinguish them?
8. Why is reusing the domain layer from project 016 important, and what does the storage seam buy the test suite?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- [ ] The README's 19 sections are present in order; this file is the reference.
- [ ] Every functional requirement in section 8 is satisfied.
- [ ] Every verification case in section 14 has a corresponding test, and every test uses a per-test temporary directory.
- [ ] The domain layer from project 016 is reused without modification; only the storage layer is new.
- [ ] A pre-rename failure leaves the destination file's bytes equal to the previous valid bytes, and no leftover temporary file appears in the destination directory.
- [ ] A rename failure leaves the destination file's bytes unchanged and is reported; the program does not remove the destination or copy the temporary file into place.
- [ ] The on-disk JSON document explicitly carries the next-ID value, and a round trip preserves it.
- [ ] The package documentation states the storage seam, the on-disk JSON field names including the next-ID field, the pre-rename failure policy, the rename-failure policy, and the distinction between atomic visibility and crash durability.
- [ ] The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Backup before save.** Before renaming the temporary file over the destination, copy the previous destination file to a sibling file with a `.bak` suffix. The backup is overwritten on every save. Do not add rotation, retention, or restore commands.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisite:** [Project 016 — Todo CLI](../../02-data-structures/016_todo_cli/README.md#20-prerequisite-based-documentation-guide).

Read the linked guide first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`encoding/json`](https://pkg.go.dev/encoding/json).
- **Standards and concept references:** [Go blog: JSON and Go](https://go.dev/blog/json).

### Project-specific learning focus

- **Learn now:** JSON schema choices, round trips, atomic temp-file replacement, file permissions, idempotence, and recovery from malformed data.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
