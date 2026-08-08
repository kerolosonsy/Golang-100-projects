# Project 025 — File Duplicate Finder

## 1. Project Name and Number

- Project **025** — `025_file_duplicate_finder`.
- The directory name and number must match exactly.
- This project builds a deterministic duplicate finder that walks an explicitly supplied root, groups regular files by byte size, computes SHA-256 only for size groups with at least two candidates, and emits groups of two or more files that share both size and digest.
- Symlinks and non-regular files are skipped; the scan never mutates the filesystem; every error is surfaced with its identifying context.

## 2. Project Idea

The duplicate finder walks the supplied root with `filepath.WalkDir`, collects every regular file's path and byte size, and groups the files by size. Size groups with fewer than two files cannot contain duplicates and are not hashed. Size groups with two or more files are hashed: each file's SHA-256 is computed by streaming its contents through `crypto/sha256`, never by reading the whole file into memory. Files within a single size group that share the same SHA-256 form one duplicate group.

A duplicate group contains at least two files that share both size and SHA-256 digest. The project's practical criterion for "duplicate" is SHA-256 identity. SHA-256 collisions are not zero, but the probability of a real collision among ordinary files is so small that the project treats SHA-256 identity as the criterion for grouping. A byte-for-byte comparison after grouping is an optional refinement, not a guarantee in the required scope.

Empty files are valid files. They all have size `0` and all have the same SHA-256 digest, and they form one duplicate group whenever there are at least two empty files.

The output is deterministic. Groups are sorted by ascending size, then ascending digest, then ascending first path. Paths inside each group are sorted lexicographically. The output uses paths relative to the supplied root, displayed with the platform-native separator but in a form the test normalizes on display so the test is portable across platforms.

## 3. Why This Project Now?

- Project 024 introduced deterministic lexical walking with rendered tree output.
- Project 025 brings that walking discipline together with a new one: hash-based identity.
- The project forces the learner to think about the cost of hashing, the cost of skipping unnecessary hashing, and the difference between "the metadata said size X" and "the bytes actually hashed to digest Y".

- The project also introduces the discipline of "trusting metadata only as long as it can be verified".
- A file's size at walk time may differ from the size the hashing step sees, because another process can modify the file between the two reads.
- The project pins what to do in that case: the finder detects the inconsistency when practical and reports an unstable-file error rather than silently claiming an incomplete scan is complete.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 025 therefore requires:

- Completion of **024** (Directory Tree Printer). Earlier projects (for example 020's safe-walk pattern, 019's streaming discipline) are background concepts already encountered and may inform the learner's design, but they are not additional required completions for this project.
- No prior knowledge of HTTP, databases, generics, or concurrency.

## 5. What You Must Know Before Starting

- That `crypto/sha256` produces a 32-byte digest and exposes a `hash.Hash` value that can be written to incrementally. Pairing the hasher with a chunked read gives a streaming SHA-256 over any `io.Reader`.
- That the SHA-256 digest of the empty input is the well-known fixed value `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`. The project asserts this value directly in tests for empty files; the test does not depend on whatever the implementation happens to produce for the empty case.
- That `os.Open` returns an `io.ReadCloser` whose `Close` method must be called to release the file descriptor. A scan that opens hundreds of files without closing them exhausts the file-descriptor limit.
- That `fs.DirEntry.Type()` distinguishes regular files, directories, symlinks, and many special files on most platforms, but on some filesystems `Type()` may return `0` for an entry whose kind the filesystem did not classify. In that case `fs.DirEntry.Info()` (or, when needed, `os.Lstat`) reports the same kind without following symlinks. The walker treats `Type()&fs.ModeSymlink != 0` as the symlink signal; when `Type()` is inconclusive, `Info()` may be consulted, but the result must never be obtained by following the symlink to its target.
- That `filepath.WalkDir` calls a callback with a `fs.WalkDirFunc`. The callback returns `fs.SkipDir` to skip a directory's contents, an error to stop the walk, or `nil` to continue.
- That Go's `map` iteration order is randomized. The output must be sorted before being emitted.
- That comparing two SHA-256 digests with `bytes.Equal` is acceptable here because the digests are not secrets.

## 6. Explanation of New Concepts

### Concepts

#### Size pre-grouping

Grouping is two-stage:

- **By size.** The first pass walks the root and records every regular file's path and recorded metadata (the metadata recorded at walk time: size and, where available, modification time). Files are bucketed by size. Size groups with fewer than two files cannot contain duplicates and are dropped at this stage without ever being hashed.
- **By SHA-256.** For each size group with at least two files, the second pass opens each file and computes its SHA-256 by streaming while counting the bytes read. Files within the group are then re-bucketed by digest. Each digest bucket within the size group is one duplicate group.

The pre-grouping by size is the project's main efficiency choice. Hashing every file in the tree is wasted work when most files are unique. Hashing only the size groups that already have candidates is much cheaper on real inputs.

#### Streaming SHA-256

The SHA-256 is computed by streaming the file's contents through the hasher while counting the bytes read. The implementation reads the file in chunks, writes each chunk into the hasher, increments a running byte counter, and finishes with `Sum(nil)`. The file is never read into a single byte slice sized to the file. A test that hashes a multi-megabyte file confirms the streaming design and the byte counter.

#### Regular files only

The walker considers only regular files. Symlinks, directories, devices, sockets, named pipes, and any other non-regular entry are skipped without hashing. A symlink that points to a regular file is not a regular file in the sense the project uses; the walker sees a symlink and skips it without resolving the link to its target.

#### Empty files

Empty files have size `0` and the standard SHA-256 digest of the empty input (`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`). The test asserts this digest directly. Two or more empty files in the same scan form one duplicate group. The standard empty-digest value is a fixed known vector; the test does not depend on whatever the implementation happens to produce.

#### Deterministic ordering

The output's groups are sorted by ascending size, with ties broken by ascending digest and then by ascending first path. Within each group, paths are sorted lexicographically. The first path used as a tie-breaker is the lexicographically smallest path in the group.

Two runs against the same root produce byte-identical output. The sort is part of the contract; iterating a map directly violates the contract.

The output's paths are root-relative (relative to the supplied root). On platforms where the path separator differs, the test normalizes the separator on display before comparing expected and actual output; the underlying values are correct for the platform but the test's expected output uses a portable form.

#### Best-effort unstable-file detection

A file's size recorded in the first pass may differ from the size the hashing pass sees, because another process can modify the file between the two reads. The project pins a best-effort detection rule and is honest about its limits:

- The first pass records each regular file's size (and, where available, its modification time as a secondary metadata snapshot).
- The hashing pass opens each file, counts the bytes it actually reads while streaming the digest, and re-stats the file after the read for a post-read size and, where available, a post-read modification time.
- A change between the recorded size and the post-read size is reported as an unstable-file error. The error names the path and the two sizes.
- A change between the recorded modification time and the post-read modification time is reported as an unstable-file error. The error names the path and the two modification times.
- A change between the recorded size and the byte counter (the number of bytes actually streamed through the hasher) is reported as an unstable-file error. The error names the path, the recorded size, and the bytes-read count.

The project is honest about the limits of best-effort detection: no portable metadata check proves a file was unchanged across the read if the file's content changed and its metadata was restored to the recorded values. Snapshotting the file or locking it for the duration of the scan is out of scope; the project does not pretend the detection is exhaustive. The test marks metadata-restoration cases as "where practical" and pins the behavior the project does guarantee: when one of the three observable comparisons detects a difference, the finder reports it.

When detection is not practical because the file has been deleted between the two reads, the project surfaces the corresponding read or stat error with the path context.

#### Same size, different digest is not an error

Two files of the same size with different content have different SHA-256 digests. That is normal, not an unstable-file error: they simply belong to different digest buckets inside the size group, and they appear as singletons (or as members of other groups) rather than as a duplicate group. The test pins the digest difference directly and asserts that the two files are not emitted as a duplicate group.

#### Never modify the filesystem

The duplicate finder never opens a file for writing, never creates files, never deletes files, never renames files. It only opens files for reading and only writes its report to the injected output writer.

#### Errors with identifying context

Every error the duplicate finder can produce carries identifying context. Scan errors (traversal, open, read, unstable) carry the affected path or size group. Writer errors carry the renderer and the underlying writer failure.

- A traversal error names the entry whose `WalkDirFunc` failed.
- An open error names the file that could not be opened.
- A read error names the file that could not be read or hashed.
- An unstable-file error names the file and the compared values (the two sizes, the two modification times, or the recorded size vs the bytes-read count).
- A writer error names the renderer and the underlying writer failure.

No error is silently hidden. No error is rendered as a successful partial scan.

## 7. Learning Objective

After completing this project the learner can:

- Walk a directory tree with `filepath.WalkDir`, classify entries by `fs.DirEntry.Type()` (with `Info()` consulted only when `Type()` is inconclusive, never following symlinks), and skip symlinks and non-regular files.
- Group regular files by byte size using `fs.DirEntry.Info().Size()` or `os.Lstat`, then compute SHA-256 only for size groups with at least two candidates.
- Stream SHA-256 over any `io.Reader` using `crypto/sha256` and a chunked read, while counting the bytes read, without loading the whole file into memory.
- Treat empty files as ordinary regular files with size `0` and the standard SHA-256 digest of the empty input, asserted directly.
- Detect, when practical, that a file's size or metadata changed between walk and hash by comparing the recorded size, the bytes-read counter, and the post-read size and modification time, and report the inconsistency as an unstable-file error rather than trusting the metadata.
- Sort the output deterministically by size, digest, and first path, with paths inside each group sorted lexicographically. Use root-relative paths and normalize the separator on display so the test is portable.
- Surface every error with its identifying context (path or size-group for scan errors; renderer and writer failure for writer errors) and never silently hide an error or render it as success.
- Use an injected open boundary for tests that would otherwise depend on permission tricks or platform-sensitive behavior, marking those cases "where practical".
- Write tests that pin equal-content grouping, same-size-different-digest non-grouping, empty-file grouping with the standard digest, singleton omission, nested directories, deterministic ordering, symlink skip, injected open errors where practical, large streamed files, and zero filesystem mutation.

## 8. Functional Requirements

1. The duplicate finder accepts a root directory path and an `io.Writer`. Production wires a real path and standard output; tests wire a per-test temporary directory and a `bytes.Buffer`. Tests that need a controllable open step use an injected open boundary.
2. The walker uses `filepath.WalkDir` and considers only regular files. Symlinks, directories, and special files (devices, sockets, named pipes) are skipped without hashing. Symlinks are recognized via `Type()&fs.ModeSymlink != 0`, with `Info()` consulted only when `Type()` is inconclusive and never in a way that follows the link.
3. The first pass records every regular file's path, recorded size, and (where available) recorded modification time.
4. Size groups with fewer than two files are not hashed and do not appear in the output.
5. Size groups with at least two files are hashed. Each file's SHA-256 is computed by streaming its contents through `crypto/sha256` while counting the bytes read. The file is never read into a single byte slice sized to the file.
6. A duplicate group is a set of two or more files that share both size and SHA-256 digest.
7. Two files with the same size and different digests are not a duplicate group; they are not emitted. The size-then-digest grouping treats them as separate digest buckets.
8. Empty files are valid and form one duplicate group whenever there are at least two of them. The group's digest is the standard SHA-256 of the empty input.
9. Groups are sorted by ascending size, with ties broken by ascending digest and then by ascending first path. The first path is the lexicographically smallest path in the group. Paths within a group are sorted lexicographically. Paths are root-relative.
10. The output contains only groups of two or more files. Singletons are not emitted.
11. Best-effort unstable-file detection compares the recorded size with the post-read size and bytes-read count, and the recorded modification time with the post-read modification time. A difference is reported as an unstable-file error naming the path and the compared values. The detection is best-effort; snapshotting and locking are out of scope.
12. If an unstable, read, traversal, or other scan failure occurs, the duplicate finder returns the error and writes no report. The report is built only after a successful scan.
13. The duplicate finder never mutates the filesystem. It never writes to any path under the root and never deletes or renames anything.
14. Every error carries its identifying context: a traversal, open, read, or unstable-file error carries its path; a writer error carries the renderer and the underlying writer failure. No error is silently hidden or rendered as success.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- A root directory path. The directory must exist. The directory may contain regular files, directories, symlinks, and special files in any combination.
- An `io.Writer` to receive the report.

#### Outputs

- A deterministic report on the injected writer. The report contains one block per duplicate group. Each block names the size, the digest, and the sorted list of file paths. Paths are root-relative.
- An error returned to the caller for any of the failure modes listed in section 8.

#### Example text-only success run

Input root:

```
photos/
├── a.jpg  (size 12345, digest D1)
├── b.jpg  (size 12345, digest D1)
├── c.jpg  (size 12345, digest D2)
└── d.txt  (size 99,   digest E1)
```

Output:

```
size=12345 digest=D1
  photos/a.jpg
  photos/b.jpg
```

(There is one duplicate group: `a.jpg` and `b.jpg`. `c.jpg` has the same size but a different digest; it is not emitted. `d.txt` is a singleton and is omitted.)

#### Example text-only empty-file run

Input root:

```
empties/
├── one.txt  (size 0)
├── two.txt  (size 0)
└── nested/
    └── three.txt  (size 0)
```

Output:

```
size=0 digest=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
  empties/nested/three.txt
  empties/one.txt
  empties/two.txt
```

#### Example text-only error runs

```
$ dedup /tmp/missing
Error: root not found: /tmp/missing.

$ dedup /tmp/somefile.txt
Error: root is not a directory: /tmp/somefile.txt.
```

## 10. Rules and Edge Cases

- **Empty root.** A directory with no entries produces an empty report. No groups, no header, zero output bytes. No error is returned.
- **Root with no duplicates.** A directory whose files are all unique (no two files share a size, or no two files share both size and digest) produces an empty report. No error is returned.
- **Single duplicate group.** Two files with the same size and digest produce one group with two paths.
- **Multiple duplicate groups.** Several independent groups are emitted in ascending-size order, with the size-then-digest-then-first-path tie-breaker.
- **Same size, different digest.** Two files of the same size with different digests are not a duplicate group; they are not emitted as a group. This is normal, not an unstable-file error.
- **Empty files.** Two or more empty files form one duplicate group with size `0` and the standard SHA-256 digest of the empty input.
- **Singleton.** A file with a unique size is not emitted.
- **Nested directories.** Files inside nested directories are walked and considered. The output uses root-relative paths.
- **Symlinks.** A symlink, whether to a file or to a directory, is skipped. The symlink's target is not opened, hashed, or listed.
- **Special files.** A device, socket, or named pipe is skipped.
- **Missing root.** The duplicate finder returns an error naming the root path. No output is written.
- **Root is a file.** The duplicate finder returns an error naming the root path and identifying the kind. No output is written.
- **Unreadable file.** A file that exists and is regular but cannot be opened or read returns an error naming the file's path. The scan stops. The report is not written.
- **File changed between walk and hash (size difference).** The file is reported as an unstable-file error naming the file's path and the recorded size and the post-read size. The scan stops. The report is not written.
- **File changed between walk and hash (bytes-read mismatch).** The byte counter at the end of the stream differs from the recorded size. The file is reported as an unstable-file error naming the path, the recorded size, and the bytes-read count. The scan stops. The report is not written.
- **File changed between walk and hash (modification time difference).** The post-read modification time differs from the recorded modification time. The file is reported as an unstable-file error naming the path and the two modification times. The scan stops. The report is not written.
- **File deleted between walk and hash.** The corresponding read or stat error is surfaced with the file's path. The scan stops. The report is not written.
- **Best-effort detection limits.** No portable metadata check proves a file was unchanged if content changed and metadata was restored to the recorded values. The detection is honest about its limits; snapshotting and locking are out of scope.
- **Writer error.** A `Write` that returns `0` and a non-`nil` error is surfaced as a renderer error. The scan has already finished; only the report emission stops.
- **Determinism.** Two runs against the same root produce byte-identical output.

## 11. Project Constraints

- Go standard library only. No third-party hashing libraries, no `xxhash`, no `md5`, no file-comparison frameworks.
- Hashing uses `crypto/sha256`. No other digest is accepted in the required scope.
- Files are streamed through the hasher; no file is loaded into a single byte slice sized to the file.
- The walker considers only regular files. Symlinks and non-regular files are skipped. Symlink recognition uses `Type()&fs.ModeSymlink != 0`, with `Info()` consulted only when `Type()` is inconclusive and never in a way that follows the link.
- Size pre-grouping is mandatory. SHA-256 is computed only for size groups with at least two candidates.
- The output uses root-relative paths and is byte-identical across runs on the same machine.
- The duplicate finder never mutates the filesystem. No open-for-write, no create, no delete, no rename, no chmod.
- Errors are surfaced with their identifying context: path or size-group for traversal, open, read, and unstable-file errors; renderer and writer failure for writer errors. No error is silently hidden or rendered as success.
- Empty files are valid regular files and participate in grouping normally. Their digest is the standard SHA-256 of the empty input and is asserted directly in tests.
- Best-effort unstable-file detection is performed when practical; snapshotting and locking are out of scope. The detection is honest about its limits.
- Tests that would otherwise depend on permission tricks or platform-sensitive behavior use an injected open boundary and are marked "where practical" when no clean seam is available.

## 12. Design Questions Before Coding

- Where does the size-pre-grouping logic live? In the walker, in a small `group` package, or in `main`? Which choice keeps the two-stage grouping obvious?
- How is the size recorded? Through `fs.DirEntry.Info().Size()` in the walker's callback, through `os.Lstat` after the walk, or through both? Which choice minimizes syscalls while keeping the data available when needed?
- How is the streaming SHA-256 implemented? Through a chunked read loop that writes into the hasher and counts bytes, through `io.Copy` augmented with a counter, or through a wrapper that closes the file? Which choice ensures the file descriptor is always closed and the byte count is always accurate?
- How is the digest compared? Through `bytes.Equal`, through `hex.EncodeToString` followed by string comparison, or through a map keyed by the raw bytes? Which choice keeps the digest representation stable?
- How is the best-effort unstable detection implemented? Through a post-read `os.Lstat` that compares size and modification time to the recorded values, through a byte counter that compares streamed bytes to the recorded size, or both? Which choice pins the detection rule in one place?
- How are the groups sorted? Through `sort.Slice` with a compound comparator, through a stable sort with a key function, or through a custom comparator? Which choice keeps the size-then-digest-then-first-path order obvious?
- How is the writer error caught? Through a check after every `Write`, through a buffered renderer that flushes at the end, or through a wrapper writer? Which choice catches partial-write failures without complicating the renderer?
- How is the open boundary for injected errors designed? Through a small interface the hasher accepts, through a wrapper around `os.Open`, or through a function-typed seam? Which choice keeps injected-open tests deterministic and avoids permission tricks?

## 13. Implementation Milestones

1. Decide the package layout. Keep the walker, the size pre-grouper, the streaming hasher, the renderer, and the error-policy logic in the same package but as separate types. Keep `main` as a thin wrapper.
2. Pin the contract as named constants: the size-then-digest grouping rule, the streaming hasher's chunk size, the sort order, and the best-effort unstable-file detection rule. Pin the standard empty-input SHA-256 digest as a known vector.
3. Implement the walker. Use `filepath.WalkDir`. Skip non-regular entries via `Type()`, with `Info()` consulted only when `Type()` is inconclusive and never in a way that follows the link. Record each regular file's path, recorded size, and (where available) recorded modification time.
4. Implement the size pre-grouper. Build a map from size to a slice of paths (with each entry carrying its recorded metadata). Drop size groups with fewer than two paths.
5. Implement the streaming hasher. Open the file, read the contents in chunks into `sha256.New()` while counting the bytes read, close the file, return the digest and the byte count. The file is never read into a single byte slice sized to the file.
6. Implement the digest sub-grouper. For each size group with at least two files, hash every file and re-bucket by digest. Each digest bucket is one duplicate group.
7. Implement the best-effort unstable-file detection. After hashing a file, re-stat it with `os.Lstat` (which does not follow symlinks) and compare the post-read size to the recorded size, the post-read modification time to the recorded modification time, and the byte counter to the recorded size. A difference is reported as an unstable-file error naming the path and the compared values.
8. Build the report in memory during the scan. The report contains one block per duplicate group with the size, the digest, and the sorted list of root-relative paths.
9. On a successful scan, write the report to the injected `io.Writer`. On any scan failure (traversal, open, read, unstable), return the error and write no report.
10. Wire `main`. Accept a positional root argument. Pass the root and standard output to the duplicate finder. Print errors to standard error.
11. Add tests for every verification case in section 14, with walker tests, hashing tests, sorting tests, and error-policy tests separated.

## 14. Verification Cases the Learner Must Write

### Required Cases

Each case is described in natural language. Tests use a per-test temporary directory. Permission and platform-sensitive tests use an injected open boundary or are marked "where practical".

#### Equal and different content

- Two files with identical content (same size, same digest) form one duplicate group with two paths.
- Three files with identical content form one duplicate group with three paths.
- Two files with different content and different sizes are not emitted (singleton by size).
- Two files with different content but the same size are not emitted as a duplicate group; they are not errors, just two separate digest buckets that do not reach the two-or-more threshold.
- One file with unique content is not emitted (singleton).

#### Same size but different hash (not an error)

- Two files of the same size with different content have different SHA-256 digests. The test hashes both files independently and asserts the digests differ. The two files are not emitted as a duplicate group. This case is not an unstable-file error.

#### Empty files (standard digest asserted directly)

- Two empty files form one duplicate group with size `0` and the digest `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`. The test asserts the digest directly against this standard value.
- Three empty files inside a nested directory structure form one duplicate group with three paths. The output lists the paths sorted lexicographically.

#### Singleton omission

- A test creates a temporary root with several files whose sizes are all unique. The output is empty (zero output bytes). No error is returned.
- A test creates a temporary root with one file and many empty directories. The output is empty. No error is returned.

#### Nested directories

- A test creates a temporary root with duplicate files at multiple nesting levels (for example, two copies in `a/`, two copies in `a/b/c/`). The output reports the duplicates with root-relative paths.
- A test creates a temporary root with one duplicate group whose files are spread across nested directories. The output sorts the paths lexicographically and the first path is the lexicographically smallest.

#### Deterministic ordering

- A test creates a temporary root with several duplicate groups of different sizes. The output lists groups in ascending size order. The test pins the exact order.
- A test creates a temporary root with two groups of the same size (different digests). The output lists the groups by ascending digest, then by ascending first path. The test pins the exact order.
- A test runs the duplicate finder twice against the same root. The two outputs are byte-identical.

#### Symlink skip

- A test creates a temporary root with a symlink to a regular file inside the root. The symlink is not opened, not hashed, not listed. The output omits the symlink.
- A test creates a temporary root with a symlink to a directory containing regular files. The symlink is not followed. The output does not contain any file from the symlink's target.
- A test creates a temporary root with a symlink loop. The walker completes without hanging. The output omits the symlinks.

#### Injected open error (where practical)

- Where practical, a test injects an open boundary that returns a planned error on a chosen file's hash step. The duplicate finder surfaces the corresponding error with the file's path. The report is not written.
- Where injection is not practical, the test marks the case "where practical" and pins the error-surfacing contract through code review of the hasher, rather than through a runtime test.

#### Large streamed files

- A test creates a temporary root with a regular file large enough that a single-byte-slice read would be unwise (for example, several megabytes). The duplicate finder hashes the file, counts the bytes read, and reports its digest. The test asserts that the digest matches an independently-computed SHA-256 over the same content and that the bytes-read counter equals the recorded size. The test does not assert that the implementation avoids loading the whole file; the streaming design is established by the implementation's chunked read and the byte counter.

#### Best-effort unstable-file detection (where practical)

- Where practical, a test injects a controlled change between walk and hash: the recorded size is X but the file's content is altered before the hash step, so the byte counter or the post-read size differs. The duplicate finder surfaces an unstable-file error naming the path and the compared values. The report is not written.
- Where practical, a test injects a controlled change to the post-read modification time. The duplicate finder surfaces an unstable-file error naming the path and the two modification times. The report is not written.
- Where metadata-restoration cases cannot be set up cleanly, the test marks the case "where practical" and pins the behavior the project does guarantee.

#### All-or-nothing output

- A successful scan writes the full report to the injected `io.Writer`. No error is returned.
- A scan that ends with an unstable-file error, a read error, an open error, or a traversal error returns the error and writes no report.
- A scan that produces no duplicate groups writes an empty report (zero output bytes) and returns no error.

#### Writer error

- A test injects an `io.Writer` that returns an error on the first write. The duplicate finder returns a renderer error. The scan has already finished; only the report emission is affected.
- A test injects an `io.Writer` that returns an error after a few successful writes. The duplicate finder returns a renderer error.

#### No file mutation

- A test creates a temporary root with a known set of files. The test records each file's modification time, mode, and contents. The duplicate finder runs against the root. After the run, every file's modification time, mode, and contents are unchanged.
- A test creates a temporary root with files of varying permissions. The duplicate finder runs against the root. After the run, the file permissions are unchanged.

#### Process

- An integration test runs the compiled binary against a temporary root with a known duplicate group and confirms the exit code is zero, the report is on standard output, and the paths match the expected sorted list.
- An integration test runs the compiled binary against a missing root and confirms the exit code is non-zero and standard error names the missing path.

## 15. Common Mistakes to Watch For

- **Hashing every file.** Files with unique sizes cannot be duplicates. Hashing them wastes work and obscures the two-stage grouping.
- **Loading the whole file into memory to hash it.** Streaming is the contract. A byte slice sized to the file defeats the streaming design.
- **Forgetting to close the file.** A scan that opens files without closing them leaks descriptors and eventually fails. Every open must have a paired close, even on error paths.
- **Following symlinks.** The contract is "regular files only". A symlink, whether to a file or to a directory, is not a regular file in the sense the project uses. The walker recognizes symlinks via `Type()&fs.ModeSymlink != 0`, with `Info()` consulted only when `Type()` is inconclusive and never in a way that follows the link.
- **Silently trusting metadata.** A file's size at walk time may differ from the size at hash time. The project requires best-effort detection and an explicit unstable-file error when the comparison detects a difference.
- **Treating "same size, different digest" as an unstable-file error.** Two files of the same size with different content are normal. They simply belong to different digest buckets inside the size group. They are not an error.
- **Writing a partial report on scan failure.** The report is built in memory during the scan and is written only after a successful scan. A failed scan returns the error and writes no report.
- **Silently hiding an unreadable file.** A file that exists but cannot be opened or read is an error with the path. Treating it as "not present" hides the permission or filesystem problem.
- **Silently hiding a writer error.** A `Write` that returns `0` and a non-`nil` error must stop the report emission. The scan has already finished; only the emission is affected. Treating the partial output as success hides the failure.
- **Iterating a map to build the output.** Map iteration is randomized. The output must be sorted, with the size-then-digest-then-first-path order and the lexicographic path order.
- **Treating SHA-256 collisions as guaranteed-impossible.** The project uses SHA-256 as the practical criterion. The probability of a real collision is small but not zero. Byte-for-byte comparison is an optional extension, not a guarantee in the required scope.
- **Treating empty files as special.** Empty files are valid regular files. They have size `0` and the standard SHA-256 digest of the empty input. The project does not skip them.
- **Mutating the filesystem.** No open-for-write, no create, no delete, no rename, no chmod. The finder is read-only.
- **Reading symlink targets with `os.Stat` (the following variant).** Resolving the target by following the link defeats the regular-files-only rule. Use `os.Lstat` (which does not follow) for the post-read metadata comparison.
- **Sorting by first path before sorting by digest.** The sort order is size, then digest, then first path. Reordering the tie-breakers produces a different output and breaks the contract.
- **Relying on `strings.ToLower` for path comparison.** Path sorting is byte-wise lexicographic. Lowercase-then-sort is a different order.
- **Producing a digest for empty files that does not match the standard value.** The standard empty SHA-256 digest is a fixed known vector. Any implementation that produces a different digest for the empty input is wrong.
- **Embedding platform-specific path separators in expected output.** The output uses the platform-native separator, but the test normalizes on display so the expected output is portable. The test does not assert a literal `/` on Windows or a literal `\` on Unix.
- **Using real home directories.** Tests must use per-test temporary directories.

## 16. Topics and References for Study

- A Tour of Go: "Methods", "Errors", "Reading files".
- Effective Go: "Errors", "Data".
- Package documentation: `crypto/sha256` (`New`, `Sum`, `Size`), `io` (`Copy`, `Reader`, `Writer`), `path/filepath` (`WalkDir`, `Join`, `Separator`, `Clean`), `io/fs` (`DirEntry`, `WalkDirFunc`, `WalkDir`, `FileMode`, `ModeSymlink`, `SkipDir`), `os` (`Open`, `Stat`, `Lstat`, `IsNotExist`), `bytes` (`Equal`), `sort` (`Slice`, `SliceStable`).
- Streaming hashing patterns: search for "Go crypto sha256 streaming", "Go io.Copy hash", "Go large file hash streaming".
- Duplicate-finding patterns: search for "Go duplicate file finder", "Go size-then-hash duplicate detection", "Go SHA-256 streaming scan".
- Best-effort unstable-file detection: search for "Go file race condition stat hash", "Go tocttou file scan", "Go stat after read inconsistency".

## 17. Self-Assessment Questions

1. Why is size pre-grouping mandatory before SHA-256, and what does hashing every file waste?
2. Why must the SHA-256 be computed by streaming while counting bytes (rather than reading the whole file), and why must the large-file test use several megabytes of content to prove the streaming design?
3. Why are symlinks skipped even when they point to regular files, and how does `Type()` plus `Info()` (when needed) support that without following links?
4. Why must empty files participate in grouping like any other regular file, and why is the standard empty-digest value a fixed known vector?
5. Why must the sort order be size, then digest, then first path, rather than alphabetical by path?
6. Why is best-effort unstable-file detection required, what does it pin about race conditions during the scan, and what does it not promise?
7. Why is SHA-256 identity the practical criterion for grouping (so that same size with different digests is normal and not an error), and what is the optional byte-for-byte refinement?
8. Why is the report built in memory and written only after a successful scan, and why must every error carry its identifying context (path or size-group for scan errors; renderer and writer failure for writer errors)?
9. Why must the duplicate finder never mutate the filesystem, and what does the no-mutation test verify?
10. Why must the test use root-relative paths and normalize the path separator on display, instead of asserting a literal separator?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- [ ] The README's 19 sections are present in order; this file is the reference.
- [ ] Every functional requirement in section 8 is satisfied.
- [ ] Every verification case in section 14 has a corresponding test. Tests that depend on platform-sensitive behavior use an injected open boundary or are marked "where practical".
- [ ] SHA-256 is computed only for size groups with at least two candidates. Files with unique sizes are never hashed.
- [ ] The hashing pass streams the file's contents through `crypto/sha256` while counting the bytes read, without loading the whole file into memory.
- [ ] Symlinks and non-regular files are skipped. Symlink recognition uses `Type()` (with `Info()` consulted only when `Type()` is inconclusive) and never follows links.
- [ ] Empty files form one duplicate group whenever there are at least two of them. The group's digest is the standard SHA-256 of the empty input, asserted directly.
- [ ] "Same size, different digest" is not an error. The two files are not emitted as a duplicate group; they belong to different digest buckets inside the size group.
- [ ] The output's groups are sorted by ascending size, ascending digest, ascending first path. Paths inside each group are sorted lexicographically. Paths are root-relative.
- [ ] Best-effort unstable-file detection compares recorded size with post-read size and bytes-read count, and recorded modification time with post-read modification time. A difference is reported as an unstable-file error naming the path and the compared values.
- [ ] The duplicate finder never mutates the filesystem. The no-mutation test confirms it.
- [ ] On any scan failure (traversal, open, read, unstable), the duplicate finder returns the error and writes no report. The report is written only after a successful scan.
- [ ] A writer error stops report emission; the scan, which has already finished, is not restarted.
- [ ] Every error carries its identifying context: a path or size-group for traversal, open, read, and unstable-file errors; a renderer and writer failure for writer errors. No error is silently hidden or rendered as success.
- [ ] The package documentation states the size-then-digest grouping rule, the streaming hasher's chunk size, the byte counter, the sort order, the best-effort unstable-file detection rule and its limits, and the no-mutation rule.
- [ ] The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Byte-for-byte verification.** After grouping by SHA-256, perform a byte-for-byte comparison of the files inside each group to confirm they are truly identical (since SHA-256 collisions are not zero). The verification streams one file at a time and stops with a clear error if a mismatch is detected. Do not replace SHA-256 with the byte comparison; the verification is an additional step after SHA-256 has already grouped the candidates.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisite:** [Project 024 — Directory Tree Printer](../../02-data-structures/024_directory_tree_printer/README.md#20-prerequisite-based-documentation-guide).

Read the linked guide first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`crypto/sha256`](https://pkg.go.dev/crypto/sha256).
- **Standards and concept references:** [NIST Secure Hash Standard](https://csrc.nist.gov/pubs/fips/180-4/upd1/final).

### Project-specific learning focus

- **Learn now:** size-then-hash grouping, streaming large files, deterministic reports, symlink policy, TOCTOU limitations, and content-addressed identity.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
