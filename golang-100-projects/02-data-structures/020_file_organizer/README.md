# Project 020 — File Organizer

## 1. Project Name and Number

Project **020** — `020_file_organizer`. The directory name and number must match exactly. This project builds a deterministic plan that groups regular files under a chosen root by lowercase extension, prints the plan, and only then performs moves under explicit confirmation.

## 2. Project Idea

The program walks an explicit root, classifies every regular file by its lowercase extension against a pinned whitelist, and produces a deterministic organization plan. Each regular file maps to a destination path of the form `<root>/<category>/<original-name>`, where `<category>` is the lowercase extension prefixed with an underscore. Two category names are reserved for files that do not fit the normal scheme: `_no_extension` for files with no extension at all, and `_other` for files whose extension is not in the pinned whitelist.

The program's safe mode is "no execution flag means dry-run". The absence of `--execute` keeps the program in dry-run: it prints the planned source-to-destination moves and performs zero filesystem mutations, including no destination-directory creation. With `--execute`, the program first builds and validates the complete plan, then prints the validated plan, then creates the needed directories and moves the files. Any destination collision, any two sources mapping to the same destination, or any destination that already exists and is not the source itself is a preflight error reported before any move happens. Symlinked directories are not followed. Symlinks and other special files are skipped and reported.

## 3. Why This Project Now?

Projects 016 through 019 introduced collections, persistence, CSV parsing, and text streaming. Project 020 introduces walking a real filesystem, building a deterministic plan, and executing it safely. The discipline of "validate the whole plan before moving anything" is the project's core lesson: a partial execution that leaves the directory half-organized is worse than refusing to start.

The project also introduces the dry-run-as-default discipline. A file organizer that mutates the filesystem on every run is dangerous; a file organizer that prints the plan first lets the user read what will happen before committing. The README pins the policy: absence of `--execute` is dry-run, presence of `--execute` is the only way to mutate the filesystem.

Finally, the project establishes the discipline of excluding its own destination directories. A second run on a root that already contains `_txt`, `_md`, and `_no_extension` directories must not move files into `_txt/_txt/`, and must not move files whose extension matches a category directory back into the root. The exclusion list makes re-runs idempotent.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 020 therefore requires:

- Completion of **019** (Word Frequency Counter), including the discipline of streaming input and producing deterministic output.
- No prior knowledge of HTTP, databases, generics, or concurrency.
- Familiarity with `filepath.WalkDir` is helpful but not required; this project introduces it.

## 5. What You Must Know Before Starting

- That `filepath.WalkDir` traverses a directory tree in lexical order, calling a callback for every entry. The callback receives the path, a `fs.DirEntry`, and an error from opening the directory.
- That `filepath.Ext` returns the file's extension, including the leading dot, or the empty string if there is no extension. `notes.txt` returns `.txt`; `Makefile` returns the empty string.
- That `os.Rename` moves a file within the same filesystem atomically from the kernel's point of view. If the source and destination are on different filesystems, `os.Rename` returns an error rather than copying the file. On platforms or filesystems where `os.Rename` cannot replace an existing destination, the call returns an error.
- That `os.Lstat` returns information about a path without following symlinks. `os.Stat` follows symlinks. The project uses `Lstat` to avoid loops.
- That symlinks can form loops; a directory whose symlink points back to its parent can trap a naive walker. The project uses `WalkDir`, which does not follow symlinked directories by default.
- That a `fs.DirEntry` exposes `IsDir()`, `Type()`, and `Info()` and is the right input to make decisions about whether an entry is a regular file, a directory, or a special file.
- That path containment is a path-aware concept: a destination is contained in a root when its cleaned relative form does not begin with `..` and is not absolute. A naïve prefix check on path strings is wrong; for example, the prefix `root` matches both `root/notes.txt` and `root2/notes.txt`.

## 6. Explanation of New Concepts

### A deterministic organization plan

The plan is a slice of moves. Each move has a source path, a destination path, and a category name. Two runs against the same root produce identical plans, with the moves listed in deterministic lexical order. The plan is the program's single source of truth for execution: nothing is moved until the plan is fully validated.

The category for a file is derived from its lowercase extension. The whitelist is pinned by the README and is not configurable:

- `.txt` maps to `_txt`
- `.md` maps to `_md`
- `.go` maps to `_go`
- `.json` maps to `_json`
- `.csv` maps to `_csv`
- `.png` maps to `_png`

Every other non-empty extension maps to `_other`. A file with no extension maps to `_no_extension`. The exact category name for a file with a whitelisted extension is the lowercase extension without the leading dot, prefixed with an underscore (for example, `_txt`).

### Safe mode is the default

The program's safe mode is "absence of `--execute` means dry-run". With dry-run, the program runs the full preflight validation but performs zero filesystem mutations. No destination directory is created. No file is moved. The program prints the plan and exits with code zero when the plan is valid.

With `--execute`, the program runs the full preflight validation. If validation succeeds, the program prints the validated plan, then creates the destination directories that are needed, then performs the moves in plan order. If a move fails after earlier moves have completed, the program stops and reports exactly which moves completed and which did not. The program does not roll back earlier moves; it does not copy across filesystems when `os.Rename` returns an error; it does not delete anything.

### Two-phase execution: validate, then move

When `--execute` is supplied, execution is in two phases:

1. **Validate.** Walk the root, build the plan, check for collisions (two sources mapping to the same destination), destination-already-exists errors (where the source is not the same file), and any path that escapes the root. If any validation check fails, the program reports the error and exits with a non-zero code. No directory is created, no file is moved.
2. **Move.** Print the validated plan. Create the destination directories that are needed. Then perform the moves in the plan's order. If any move fails, the program stops and reports exactly which moves completed and which did not.

Dry-run is validation without mutation. The dry-run path runs the same validation but does not create directories and does not move files.

### Excluding destination category directories

The walk skips only the exact top-level destination directories directly under the supplied root whose names are in the pinned category set. A root-level `_txt` is not re-walked into `_txt/_txt/`, but an unrelated nested source directory that happens to be named `_txt` is not silently ignored. Exclusion is based on the full cleaned destination path, not on a matching base name anywhere in the tree.

This makes a second run idempotent: running the organizer against an already-organized root produces a plan with zero moves and exits with code zero.

### Symlinks and special files

Symlinks, device files, sockets, named pipes, and any other non-regular entry are skipped and reported in a separate section of the plan. Symlinked directories are not followed. The walk uses `WalkDir` and the entry's `Type()` method to distinguish regular files from everything else.

### Path containment

The plan only includes moves whose destination is contained within the supplied root. Containment is a path-aware concept: a destination is contained in the root when its cleaned relative form (relative to the cleaned root) is not absolute and does not begin with `..`. A naïve string-prefix check is wrong: the prefix `root` matches both `root/notes.txt` and `root2/notes.txt`. The test in section 14 includes a sibling-prefix trap test to pin the correct behavior.

The program never touches any path outside the supplied root. A symlink that points outside the root is not followed.

### `os.Rename` and cross-filesystem behavior

All planned destinations are under the supplied root, so a well-formed plan stays within one filesystem in most cases. Nested mount points inside the root can still cause `os.Rename` to return a cross-device error, and the program must report that error. There is no cross-device copy fallback in the required scope; the program reports the failure, lists the moves that completed, and exits non-zero.

On a single filesystem, `os.Rename` preserves file contents and modes for the move. The verification in section 14 confirms that contents and modes are preserved after the move.

## 7. Learning Objective

After completing this project the learner can:

- Walk a directory tree with `filepath.WalkDir` in deterministic lexical order and classify entries by type.
- Build a deterministic plan of source-to-destination moves that groups files by lowercase extension against a pinned whitelist.
- Distinguish "missing file", "empty plan", and "plan with moves" and apply the right policy to each.
- Implement a two-phase executor that validates the plan in full before moving anything.
- Treat any destination collision or two sources mapping to the same destination as a preflight error.
- Use absence-of-`--execute` as dry-run and document that dry-run performs zero filesystem mutations, including no destination-directory creation.
- Exclude destination category directories from the walk so a second run is idempotent.
- Skip symlinks and special files without following them and without forming loops.
- Enforce path-aware containment of destinations, including the sibling-prefix case.
- Recognize that `os.Rename` returns a cross-device error rather than copying, and report rename failures without copying or removing.
- Write tests that use a per-test temporary directory and never touch the user's real files.

## 8. Functional Requirements

1. The program accepts a single positional argument naming the root directory to organize. Production wires a real directory; tests wire a per-test temporary directory.
2. The program accepts an `--execute` flag. The absence of the flag is dry-run; its presence enables mutations.
3. The walk uses `filepath.WalkDir` and processes paths in lexical order. The plan's moves are emitted in the same lexical order.
4. The whitelist is pinned: `.txt` → `_txt`, `.md` → `_md`, `.go` → `_go`, `.json` → `_json`, `.csv` → `_csv`, `.png` → `_png`. Every other non-empty extension maps to `_other`. Files with no extension map to `_no_extension`.
5. Every regular file's destination is `<root>/<category>/<original-name>`, where `<category>` is one of the pinned category names and `<original-name>` is the file's basename preserved exactly.
6. Dry-run performs full preflight validation. It prints the planned moves and the list of skipped entries, performs zero filesystem mutations, creates no destination directory, and exits with code zero when the plan is valid.
7. `--execute` performs full preflight validation. If validation fails, no directory is created, no file is moved, and the program reports the failure. If validation succeeds, the program prints the validated plan, creates the needed destination directories, and performs the moves in plan order.
8. Any destination collision (two sources mapping to the same destination) is a preflight error reported before any move.
9. Any destination that already exists, where the source is not the same file, is a preflight error reported before any move.
10. Every destination path is contained within the supplied root under a path-aware containment rule. Sources outside the root are not part of the plan.
11. Symlinked directories are not followed. Symlinks and other special files are skipped and reported.
12. The walk excludes the exact root-level destination directories in the pinned category set. It does not exclude unrelated nested directories solely because they have the same base name. A second run on an already-organized root produces an empty plan and exits with code zero.
13. If a move fails after earlier moves have succeeded, the program stops and reports exactly which moves completed and which did not. The program does not roll back earlier moves and does not attempt a cross-filesystem copy fallback.

## 9. Inputs and Outputs

### Inputs

- A root directory path. The directory must exist and be readable. The directory may be empty or contain a mix of regular files, subdirectories, symlinks, and special files.
- An optional `--execute` flag.

### Outputs

- A list of planned moves on standard output. The format includes one line per move, naming the source and destination paths. The format also includes a section listing skipped symlinks and special files.
- Errors on standard error for any preflight validation failure or any execution failure.
- Exit code zero on success (including a successful dry run with an empty plan); non-zero on any failure.

### Example text-only dry-run output

```
$ organize /tmp/testroot
Plan for /tmp/testroot:
  /tmp/testroot/notes.txt       -> /tmp/testroot/_txt/notes.txt
  /tmp/testroot/Makefile        -> /tmp/testroot/_no_extension/Makefile
  /tmp/testroot/img.PNG         -> /tmp/testroot/_png/img.PNG
  /tmp/testroot/script.bak      -> /tmp/testroot/_other/script.bak
Skipped (not regular files):
  /tmp/testroot/link -> symlink
No changes performed (dry run).
```

### Example text-only `--execute` output

```
$ organize --execute /tmp/testroot
Validated plan for /tmp/testroot:
  /tmp/testroot/notes.txt       -> /tmp/testroot/_txt/notes.txt
  /tmp/testroot/Makefile        -> /tmp/testroot/_no_extension/Makefile
  /tmp/testroot/img.PNG         -> /tmp/testroot/_png/img.PNG
  /tmp/testroot/script.bak      -> /tmp/testroot/_other/script.bak
Executed 4 moves.
```

### Example text-only failure runs

```
$ organize /tmp/testroot
Error: destination collision: two sources map to /tmp/testroot/_txt/notes.txt: /tmp/testroot/one/notes.txt and /tmp/testroot/two/notes.txt.
$ organize /tmp/testroot
Error: destination already exists: /tmp/testroot/_txt/notes.txt (source is /tmp/testroot/other-notes.txt).
$ organize --execute /tmp/testroot
Validated plan for /tmp/testroot:
  /tmp/testroot/a.txt -> /tmp/testroot/_txt/a.txt
Executed 1 move; move 2 of 3 failed: <error>. Earlier moves remain in place.
```

## 10. Rules and Edge Cases

- **Empty root.** A root that contains no entries produces an empty plan, prints "no moves planned", and exits with code zero.
- **Root with only directories.** A root that contains only directories produces an empty plan, exits with code zero, and creates no destination directories.
- **Files with no extension.** A file named `Makefile` maps to `_no_extension/Makefile`. The category name is the literal `_no_extension`.
- **Files with unknown extensions.** A file named `script.bak` maps to `_other/script.bak` because `.bak` is not in the pinned whitelist.
- **Mixed case extensions.** A file named `IMG.PNG` maps to `_png/img.PNG`. The category is derived from the lowercase extension. The original filename is preserved exactly.
- **Collision.** Two source files whose destinations coincide — for example, `one/notes.txt` and `two/notes.txt` both mapping to `_txt/notes.txt` — are a preflight error. The program reports both sources and the destination.
- **Existing destination.** A destination that already exists, where the source is not the same file, is a preflight error. The program reports the existing destination and the source.
- **Source is the same as destination.** A file already at its planned destination is not a collision; the plan excludes it.
- **Symlink to a file.** A symlink pointing to a regular file is not a regular file; it is a symlink. The plan skips it and reports the skip.
- **Symlink to a directory.** A symlink pointing to a directory is not a regular directory; the walk does not follow it. The plan skips it and reports the skip.
- **Symlink loop.** A directory tree that contains a symlink loop is handled by `WalkDir`'s default behavior, which does not follow symlinked directories. The plan completes.
- **Already-organized root.** A root that already contains `_txt`, `_md`, and `_no_extension` directories. The walk excludes those directories; the plan is empty. A second run is idempotent.
- **Sibling-prefix trap.** A root whose name is a prefix of another path (for example, the root is `root` and the test also creates `root2`) is handled by the path-aware containment rule. The plan only contains destinations under the cleaned root; nothing in `root2` is part of the plan.
- **Move failure mid-execution.** If a move fails after earlier moves have succeeded, the program stops and reports which moves completed and which did not. The program does not roll back earlier moves.
- **Cross-device rename error.** Even when the source and destination appear to share a root, nested mount points can make `os.Rename` return an error. The program reports the error and leaves earlier completed moves in place.
- **Source outside the root.** A file passed via a path that escapes the root is not part of the plan. The path-aware containment rule rejects it.

## 11. Project Constraints

- Go standard library only. No third-party filesystem-walking libraries.
- The walk uses `filepath.WalkDir` from the standard library. Symlinked directories are not followed.
- The program operates only inside the supplied root. It does not touch any path outside the root.
- Absence of `--execute` means dry-run. Presence of `--execute` is the only way to mutate the filesystem.
- Two-phase execution: validate in full, then move. No partial validation followed by partial move.
- The whitelist is pinned and is not configurable in the required scope. The category-name format is the lowercase extension without the leading dot prefixed with an underscore (for example, `_txt`), plus the reserved names `_no_extension` and `_other`.
- The plan's moves are emitted in deterministic lexical order.
- Path containment is path-aware: relative paths are cleaned and checked for an absolute form or a leading `..`. A naïve string-prefix check is not acceptable.
- `os.Rename` is the only move primitive. There is no cross-device copy fallback in the required scope. A rename error is reported as an execution failure; earlier completed moves remain in place.
- No deletion, recursive cleanup, content rewriting, watching, or cross-device copy fallback in the required scope.

## 12. Design Questions Before Coding

- Where does the plan live? As a slice of move structs inside `main`, in a small package returned by a `Plan` function, or in two packages (a planner and an executor)? Which choice lets the test drive the planner directly without touching the filesystem?
- How are collisions detected? By walking the plan and looking for duplicate destinations, by building a destination-to-source map, or by sorting and scanning for adjacent duplicates? Which choice scales?
- How is the walk's lexical order enforced? Through `WalkDir`'s default, through an explicit sort, or through `ReadDir` per directory? Which choice guarantees determinism?
- How is the exclusion list enforced? Through a set lookup against each directory's basename, through a switch on the basename, or through a separate filtering pass? Which choice keeps the category-name set in one place?
- How is `--execute` integrated? As a `flag.Bool` with a default of `false`, as a command-line argument, or as a separate subcommand? Which choice keeps the dry-run default safe?
- How are symlinks detected? Through `d.Type()&ModeSymlink != 0`, through `os.Lstat`, or through both? Which choice is robust against targets that have changed?
- How is path containment enforced? Through `filepath.Rel` to compute the relative path and reject when it is absolute or begins with `..`, through a prefix check, or through a containment helper? Which choice is correct on every platform and survives the sibling-prefix case?
- How is the failure path of a move reported? Through a typed error with the move's index, through a wrapped error, or through a sentinel? Which choice lets the test assert on the index?
- How is the partial-execution failure test simulated where practical? Through an injected move boundary the executor calls, through a pre-arranged read-only destination, or by marking the test as "where practical"? Which choice avoids permission tricks and remains reliable?

## 13. Implementation Milestones

1. Decide the package layout: a small planner package that takes a root path and returns a plan plus a validation result, and a thin `main` package that parses flags, calls the planner, and dispatches to dry-run or `--execute`.
2. Define the move type and the plan type. The plan exposes its moves in deterministic lexical order, its skipped entries, and a validation result that lists any preflight errors.
3. Implement the walk. Use `filepath.WalkDir` and process entries in lexical order. Exclude directories whose names are in the pinned category set (whitelist categories plus `_no_extension` and `_other`).
4. Classify each regular file. Compute the lowercase extension, look up the category from the pinned whitelist, default to `_other` for unknown non-empty extensions, and use `_no_extension` for files with no extension. Compute the destination path under the root.
5. Validate the plan. Detect destination collisions, destination-already-exists errors (excluding the source itself), and any destination whose path-aware containment fails. The validation result lists every error found.
6. Build the dry-run output. Print every move on a single line. Print the list of skipped entries. Print the empty-plan message when the plan has no moves. Print "No changes performed (dry run)." when the plan is non-empty.
7. Wire the `--execute` path. Validate the plan; if any validation fails, report all errors and exit without moving anything. If validation succeeds, print the validated plan, create the destination directories, and perform the moves in plan order. Report any failure with the move's index.
8. Wire the flag parsing. Absence of `--execute` means dry-run; presence of `--execute` enables mutations.
9. Add the integration tests that run the compiled binary against per-test temporary directories with a mix of regular files, symlinks, subdirectories, and collision cases. Add the sibling-prefix trap test.

## 14. Verification Cases the Learner Must Write

Each case is described in natural language. Every test uses a per-test temporary directory; no test touches the repository, the user's home, or any other real directory. No test depends on permission tricks or running as a particular user.

### Dry run

- A test creates a temporary root with three regular files of different extensions and runs the organizer without `--execute`. Standard output contains exactly the expected three moves, in lexical order. No destination directory is created. No file is moved. The exit code is zero.
- A test creates a temporary root with no entries and runs the organizer without `--execute`. Standard output contains the empty-plan message. No directory is created. The exit code is zero.

### Classification

- A file named `notes.txt` is planned for `_txt/notes.txt`.
- A file named `Makefile` (no extension) is planned for `_no_extension/Makefile`.
- A file named `script.bak` (extension not in the pinned whitelist) is planned for `_other/script.bak`.
- A file named `IMG.PNG` is planned for `_png/img.PNG` (lowercase extension, original filename preserved).
- A file with a hyphen in its name, for example `my-notes.txt`, is planned for `_txt/my-notes.txt`.
- A file named `data.json` is planned for `_json/data.json`.
- A file named `image.png` is planned for `_png/image.png`.
- A file named `report.csv` is planned for `_csv/report.csv`.

### Validation

- A root with two files in different subdirectories whose destinations collide — for example, `one/notes.txt` and `two/notes.txt` both mapping to `_txt/notes.txt` — produces a preflight error that names both sources and the destination. No move is performed.
- A root with one regular file and one existing destination file whose name is the same as the planned destination (and which is not the source itself) produces a preflight error that names the existing destination and the source. No move is performed.
- A sibling-prefix trap test creates two sibling temporary roots, for example `root` and `root2`, places a file `notes.txt` in each, and asserts that the plan for the `root` directory contains only moves under `root` and nothing under `root2`. The test pins the path-aware containment rule.

### Execution

- A test creates a temporary root with three regular files and runs the organizer with `--execute`. After the run, each file is in its planned destination, the original locations are empty, and the file contents and modes are preserved.
- A test creates a temporary root with three regular files, runs the organizer with `--execute`, then runs it again without `--execute`. The second run produces an empty plan because the destination directories already exist and are excluded. The exit code is zero.
- A test creates a temporary root with mixed entries: regular files in two subdirectories and one symlink. The plan contains the regular files. The symlink is reported as skipped. No destination is inside the symlink's target.

### Symlinks

- A root containing a symlink to a regular file skips that symlink and reports the skip.
- A root containing a symlink to a directory does not traverse the symlinked directory and reports the skip.
- A root containing a symlink loop is handled without hanging. The walk completes and the plan contains no entries from the loop.

### Already-organized

- A root that already contains `_txt`, `_md`, `_go`, `_json`, `_csv`, `_png`, `_no_extension`, and `_other` directories at its top level: the walk excludes those directories. The plan contains only entries in the root itself, not in the category directories.
- A nested source directory whose base name is `_txt`, but whose full path is not the root-level `_txt` destination, is still traversed. Its eligible files appear in the plan normally.

### Determinism

- A test runs the planner twice against the same temporary root and confirms the two plans are byte-identical (modulo any absolute-path differences, which the test normalizes).
- A test runs the planner against a root whose entries are created in random order across many runs. The plan's moves are always in the same lexical order.

### Partial failure (where practical)

- Where practical, a test injects a move boundary the executor calls, so the test can simulate a single move failure deterministically without changing directory permissions. The test asserts that the program reports the failing move's index, lists the moves that completed, and exits with a non-zero code. Earlier completed moves remain in place.
- Where injecting a move boundary is not practical, the test marks the case as "where practical" and pins the partial-failure reporting shape through code review and inspection rather than a flaky runtime test.

### Process

- An integration test runs the compiled binary without `--execute` against a temporary root and confirms the exit code is zero and standard output contains the plan.
- An integration test runs the compiled binary with `--execute` against a temporary root with a valid plan and confirms the exit code is zero, the validated plan is printed before any mutation, and the moves completed.
- An integration test runs the compiled binary against a temporary root with a collision and confirms the exit code is non-zero and standard error names both sources.

## 15. Common Mistakes to Watch For

- **Following symlinked directories.** `filepath.WalkDir` does not follow them by default, but a custom walker that calls `os.Stat` instead of using the `DirEntry`'s `Type` will follow them. The project requires `WalkDir`'s default behavior.
- **Excluding by base name everywhere.** Only the root-level destination paths are excluded. Skipping every nested directory named `_txt` or `_other` can silently omit legitimate source files.
- **Lowercasing the filename instead of the extension.** The category is derived from the lowercase extension; the original filename is preserved. A program that lowercases the filename changes user data and is wrong.
- **Partial validation followed by partial move.** The project requires validate-in-full, then move. A program that validates and moves one entry at a time can leave the directory half-organized.
- **Overwriting existing destinations.** A preflight check must catch the case where the destination exists and is not the source.
- **Treating an already-correctly-placed file as a collision.** A file whose destination equals its source is not a collision; the plan excludes it.
- **Using a string-prefix check for containment.** A prefix check on the path string is wrong: the prefix `root` matches both `root/notes.txt` and `root2/notes.txt`. The containment rule must be path-aware, using a cleaned relative form that is not absolute and does not begin with `..`.
- **Using `filepath.Walk` instead of `filepath.WalkDir`.** Both walk, but `WalkDir` avoids an extra `os.Stat` per entry and exposes `Type` directly. Either works for this project; the discipline is to use the entry's `Type` rather than calling `os.Stat` again.
- **Assuming `os.Rename` is universally atomic or that it falls back to copy-and-delete.** `os.Rename` is atomic on a single filesystem; across filesystems it returns an error. The README does not promise cross-filesystem atomicity, and the program does not attempt a copy fallback in the required scope.
- **Silently skipping validation errors.** A preflight error is reported with enough information for the user to fix the input. A silent skip leaves the user confused.
- **Reporting failures without an index.** If the program stops mid-execution, it must say which move index failed and which moves already completed. The test pins this.
- **Modifying files outside the root.** A path-aware containment check is required. A symlink that points outside the root must not be followed.
- **Running a permission-based partial-failure test.** Changing directory or file permissions to simulate a failed move is flaky, depends on the test runner's user, and is unsafe to assume. Use an injected move boundary where practical; mark the case as "where practical" otherwise.
- **Touching the user's real files in tests.** Every test uses a per-test temporary directory. A test that runs against the user's home or the repository is a critical safety bug.

## 16. Topics and References for Study

- A Tour of Go: "Packages", "Reading files".
- Effective Go: "Errors", "Data".
- Package documentation: `path/filepath` (`WalkDir`, `Walk`, `Ext`, `Rel`, `Base`, `Dir`, `IsAbs`, `Clean`), `os` (`Stat`, `Lstat`, `Rename`, `MkdirAll`, `ReadDir`), `io/fs` (`DirEntry`, `WalkDirFunc`, `FileMode`, `ModeSymlink`, `ModeDir`, `ModeIrregular`, `ModeDevice`), `flag` (`Bool`, `String`, `Parse`), `errors` (`Is`, `As`, `Join`).
- Filesystem traversal patterns: search for "Go filepath.WalkDir", "Go symlink loop safe walk", "Go os.Rename cross-device error".
- Idempotent CLI design: search for "dry run flag", "two-phase validation execution", "idempotent file operations".
- Path containment: search for "Go filepath.Rel containment", "Go path-aware root check", "sibling prefix trap file paths".

## 17. Self-Assessment Questions

1. Why is absence-of-`--execute` the dry-run default, and what does that buy a learner?
2. Why is the plan validated in full before any move, and what would break if validation happened move-by-move?
3. Why is the destination category derived from the lowercase extension while the original filename is preserved?
4. Why must the walk exclude destination category directories, and what would a second run look like without that exclusion?
5. Why does the program not follow symlinked directories, and how does `filepath.WalkDir` make that the default?
6. Why is a destination collision a preflight error instead of a runtime overwrite?
7. Why does the program report exactly which moves completed and which did not when a move fails mid-execution?
8. Why is the containment rule path-aware rather than a string-prefix check, and what does the sibling-prefix trap test pin?
9. Why does the program report a cross-device rename error rather than copying the file, and what guarantee would a copy fallback violate?
10. Why does the test use a per-test temporary directory, and what safety property would be lost if it used a real directory?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every functional requirement in section 8 is satisfied.
- Every verification case in section 14 has a corresponding test, and every test uses a per-test temporary directory.
- Absence of `--execute` performs zero filesystem mutations, including no destination-directory creation, while still running full preflight validation.
- `--execute` prints the validated plan before any mutation, then performs the moves in plan order.
- A preflight collision or destination-already-exists is reported before any move.
- Symlinked directories are not followed; symlinks and special files are skipped and reported.
- Every destination path in the plan is contained within the supplied root under the path-aware containment rule, including the sibling-prefix case.
- File contents and modes are preserved after the move on a single filesystem.
- The package documentation states the whitelist, the category-name format, the reserved names, the dry-run policy, and the rename-failure policy.
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Category whitelist via flag.** Accept a comma-separated flag that overrides the pinned whitelist. Extensions not in the override list still go to `_other`. The override is read-only at startup and does not change mid-execution. Do not add per-file rules or regex filters.
- **Verbose move log.** Accept a `--verbose` flag that prints a one-line confirmation for each completed move during execution. Dry-run mode is unchanged. Do not add progress bars, file counts, or timing.
