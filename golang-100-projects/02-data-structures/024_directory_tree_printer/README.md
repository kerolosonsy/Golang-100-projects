# Project 024 — Directory Tree Printer

## 1. Project Name and Number

Project **024** — `024_directory_tree_printer`. The directory name and number must match exactly. This project builds a deterministic lexical tree printer rooted at an explicitly supplied directory. The printer walks the supplied root with `io/fs` and `path/filepath`, renders the tree into a string written to an injected `io.Writer`, marks symlinks without following them, and surfaces every error with its path context.

## 2. Project Idea

The printer takes a root directory path, an integer depth, and an `io.Writer`. It prints the root on the first line, then prints every child of the root that is within the depth limit, indented under the root, and so on recursively. Files and directories are both rendered. Symlinks are rendered with a visible mark and are not followed, so a symlink loop cannot trap the printer.

The output is deterministic. Sibling basenames are sorted lexicographically before rendering, so two runs against the same root produce byte-identical output. The output uses basenames and indentation only; it does not include platform-specific path separators in normal lines, so the test does not need to normalize a separator. The printer separates the traversal from the rendering so that traversal can be tested with synthetic directory listings and rendering can be tested with synthetic node trees, without depending on real directories.

Errors are explicit. A missing root, a root that is a file, an unreadable directory, a disappearing entry, and a writer error are all reported with the path or context they affect. The printer never silently hides an error or renders an error as success.

## 3. Why This Project Now?

Project 023 introduced a small language with a state machine and a pinned output layout. Project 024 revisits filesystem walking with a stricter determinism requirement: every sibling must be sorted, every error must carry its path, and the renderer must be decoupled enough from the walker that the renderer's output can be asserted against a synthetic tree. The project is also the first in the path that asks for a clear separation between "what to print" and "how to print it", with the split testable from both sides.

The project's discipline around symlinks is the same discipline as project 020: do not follow symlinked directories, do not loop on them, mark them visibly. Project 024 pins that discipline one more level by adding a visible mark and by adding explicit tests for symlink loops. The learner practices the same habit at a smaller scope.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 024 therefore requires:

- Completion of **023** (Markdown to HTML Converter). Earlier projects (for example 020's safe-walk pattern and the wider filesystem concepts already encountered) are background concepts and may inform the learner's design, but they are not additional required completions for this project.
- No prior knowledge of HTTP, databases, generics, or concurrency.

## 5. What You Must Know Before Starting

- That `filepath.WalkDir` traverses a directory tree and calls a callback for every entry. The callback receives the path, a `fs.DirEntry`, and an error from opening the directory. Within a single directory, `WalkDir` emits the children in lexical order after reading the directory; the project's observable requirement is that siblings appear in byte-wise lexicographic order in the rendered output.
- That `fs.DirEntry` exposes `IsDir()`, `Type()`, and `Info()`. `Type()` returns a `FileMode` that distinguishes regular files, directories, symlinks, and some special files on most platforms. On some filesystems `Type()` may return `0` for an entry whose kind the filesystem did not classify; in that case `Info()` (which calls `Lstat`) reports the same kind without following symlinks. The walker treats `Type()&fs.ModeSymlink != 0` as the symlink signal; when `Type()` is inconclusive, `Info()` may be consulted, but the result must never be obtained by following the symlink to its target.
- That a `fs.WalkDirFunc` returns `fs.SkipDir` to skip a directory's contents, an error to stop the walk, or `nil` to continue.
- That `io.Writer.Write` returns the number of bytes written and an error. A write that returns `0` and a non-`nil` error must be treated as a writer failure.
- That the output uses basenames and indentation rather than full path strings. The renderer does not embed path separators in normal lines, so the test does not need to normalize separators before comparing expected and actual output.

## 6. Explanation of New Concepts

### Depth semantics

The printer accepts a non-negative integer `depth`. The meaning of `depth` is pinned:

- `depth = 0` prints only the root line. The root is rendered once, with no children.
- `depth = 1` prints the root and every direct child of the root.
- `depth = N` for `N >= 1` prints the root and every entry up to `N` levels of nesting below it.
- `depth < 0` is invalid. The printer returns an error and writes nothing.

The contract is "non-negative integer only". There is no unlimited-depth sentinel. A negative depth is a hard error, not an unlimited-depth request.

### Lexical sibling sorting

Before rendering a directory's children, the printer sorts the children's basenames lexicographically. Lexical order means byte-wise comparison, not case-insensitive comparison and not Unicode-aware collation. The test pins the exact order against known directory contents. The renderer's observable contract is "siblings appear in byte-wise lexicographic order"; the implementation may use any standard-library primitive that produces that order.

The first rendered line is the root's basename. The second and later lines are the root's children, indented under the root. Each subsequent level of nesting is indented further by the pinned two-space unit.

### Files, directories, and special files

Every regular file is rendered. Every directory is rendered. Special files (devices, sockets, named pipes) are rendered but never descended into. The marks are pinned by this README:

- A directory is rendered as its basename.
- A regular file is rendered as its basename.
- A symlink is rendered as its basename followed by the trailing mark `@`.
- A special file is rendered as its basename followed by the trailing mark `!`.

The basename-and-mark rule keeps the output free of full paths and free of platform-specific path separators in normal lines.

### Symlink handling

A symlink is rendered with the trailing mark `@` immediately after the basename. The printer does not display the symlink's target in the rendered output. The printer never follows the symlink to its target; the symlink is recognized by `Type()&fs.ModeSymlink != 0` (with `Info()` consulted only when `Type()` is inconclusive and only in a non-following mode), and the walk does not descend into symlinked directories. This prevents loops. A symlink loop in the input is handled without hanging: the printer renders the symlink's basename with the `@` mark and stops descending at that point.

`filepath.WalkDir` provides no-follow traversal behavior for symlinked directories. The walker still explicitly classifies a symlink by `Type()&fs.ModeSymlink != 0` so it can render the trailing `@` suffix. The symlink's target is never followed.

### Rendering separated from traversal

The printer's two responsibilities are split:

- The walker turns a root path into a list of tree nodes. Each node has a basename, a kind (regular file, directory, symlink, special file), a depth, a parent reference, and the data the renderer needs.
- The renderer turns the list of nodes into the final text. The renderer is pure: given the same node list, it produces the same text.

This split makes both sides testable. The walker is tested with a per-test temporary directory and a small known tree, or with an injected filesystem boundary. The renderer is tested by feeding it a synthetic node list and asserting on its output. The renderer does not read from the filesystem.

### Output written to an injected writer

The printer writes the rendered tree to an `io.Writer`. The walker does not write to the writer; the renderer does. A writer error during rendering is reported as a printer error and is never rendered as success. Because the walker finishes before rendering begins, a writer error stops the rendering step, not a walk that already completed.

### Errors with path context

Every error the printer can produce carries enough context to identify the affected path or the affected input:

- A missing root path returns an error naming the root path.
- A root that exists but is not a directory returns an error naming the root path and identifying the kind.
- An unreadable directory returns an error naming the directory path.
- A disappearing entry (an entry that was present at one moment and missing the next, for example because another process removed it during the walk) returns an error naming the entry's path.
- A writer error returns an error identifying the renderer and the failure.

No error is silently hidden, silently swallowed, or rendered as a successful empty tree.

## 7. Learning Objective

After completing this project the learner can:

- Walk a directory tree with `filepath.WalkDir`, sort sibling basenames lexicographically, and produce a deterministic rendered tree.
- Pin depth as a non-negative integer, with `0` printing the root only and negative depths treated as a hard error.
- Recognize symlinks via `fs.DirEntry.Type()` (with `Info()` consulted only when `Type()` is inconclusive, never following the symlink) and render them with the trailing `@` mark without following them, preventing loops.
- Render the tree to an injected `io.Writer` while separating the traversal from the rendering so both sides are independently testable.
- Pin the indentation at two ASCII spaces per depth and the marks for each kind, so the test's expected output is stable.
- Surface every error with its path context and never silently hide an error or render an error as success.
- Use an injectable filesystem boundary for tests that would otherwise depend on permission tricks or platform-sensitive behavior, marking platform-sensitive cases as "where practical".
- Write tests that pin the exact ordering, indentation, depth behavior, symlink handling, and error reporting without relying on platform-specific home directories or directory enumeration order.

## 8. Functional Requirements

1. The printer accepts a root path, a depth, and an `io.Writer`. Production wires a real path and standard output; tests wire a per-test temporary directory and a `bytes.Buffer`. Tests that need a controllable walker use an injectable filesystem boundary.
2. `depth` is a non-negative integer. `depth = 0` renders only the root. `depth = N` for `N >= 1` renders the root and entries up to `N` levels of nesting. `depth < 0` returns an error and writes nothing.
3. The first rendered line is the root's basename. Each subsequent line is one entry under the root or under some deeper directory.
4. Sibling basenames are sorted lexicographically (byte-wise) before rendering. The order is deterministic across runs on the same machine.
5. Files, directories, symlinks, and special files are all rendered. Indentation is two ASCII spaces per depth level. Directories and regular files are rendered as their basename. Symlinks are rendered as basename followed by `@`. Special files are rendered as basename followed by `!`.
6. Symlinked directories are not descended into. Symlink loops are handled without hanging. The symlink's target is never displayed in the rendered output and is never resolved by following the link.
7. The printer writes to an injected `io.Writer`. A writer error during rendering is returned to the caller. The walk, which has already finished, is not restarted.
8. A missing root returns an error naming the root path.
9. A root that exists but is a regular file (or any non-directory kind) returns an error naming the root path and identifying the kind.
10. An unreadable directory returns an error naming the directory path.
11. A disappearing entry during the walk returns an error naming the entry's path.
12. The output is byte-identical across two runs against the same root.
13. The walker and the renderer are separate. The walker produces a node tree; the renderer turns the node tree into text. The renderer does not call the filesystem.

## 9. Inputs and Outputs

### Inputs

- A root directory path. The path must exist; the root must be a directory or the printer returns an error. The directory may be empty or contain a mix of regular files, subdirectories, symlinks, and special files.
- A non-negative integer `depth`.
- An `io.Writer` to receive the rendered tree.

### Outputs

- The rendered tree on the injected writer. The format includes the root on the first line, then one line per entry in lexicographic order at each level, indented by two ASCII spaces per depth level. The output uses basenames and indentation; it does not embed full paths or platform-specific separators in normal lines.
- An error returned to the caller for any of the failure modes listed in section 8.

### Example text-only success run

Input root:
```
project/
├── README.md
├── cmd/
│   └── main.go
└── go.mod
```

Output for `depth = 2`:
```
project
  README.md
  cmd
    main.go
  go.mod
```

Output for `depth = 0`:
```
project
```

Output for `depth = 1`:
```
project
  README.md
  cmd
  go.mod
```

### Example text-only symlink run

Input root:
```
project/
├── real/
│   └── file.txt
└── link -> real
```

Output:
```
project
  link@
  real
    file.txt
```

(The symlink `link` is rendered as `link@`. The target path is not displayed.)

### Example text-only error runs

```
$ tree --depth=-1 /tmp/empty
Error: depth must be non-negative.

$ tree /tmp/missing
Error: root not found: /tmp/missing.

$ tree /tmp/somefile.txt
Error: root is not a directory: /tmp/somefile.txt.
```

## 10. Rules and Edge Cases

- **Empty root.** A directory with no entries renders the root's basename on the first line and nothing else. No error is returned.
- **Depth zero.** The printer renders only the root's basename. No children are rendered.
- **Depth one.** The printer renders the root and every direct child of the root, sorted lexicographically. No grandchildren are rendered.
- **Depth N for large N.** The printer renders up to N levels of nesting. A tree deeper than N is truncated at level N.
- **Negative depth.** The printer returns an error and writes nothing.
- **Missing root.** The printer returns an error naming the root path. No output is written.
- **Root is a file.** The printer returns an error naming the root path and identifying the kind. No output is written.
- **Unreadable directory.** A directory that the user lacks permission to read returns an error naming the directory path. The walk stops at that directory.
- **Disappearing entry.** An entry that is removed during the walk returns an error naming the entry's path. The walk stops at that entry.
- **Writer error.** A `Write` that returns `0` and a non-`nil` error is surfaced as a renderer error. The walk has already finished; the error is reported and rendering stops.
- **Symlink loop.** A symlink whose target is a parent or ancestor is rendered with the `@` mark and is not descended into. The walk completes.
- **Symlink to a file.** A symlink to a regular file is rendered with the `@` mark. The walk does not descend.
- **Special file.** A device, socket, or named pipe is rendered with the trailing `!` mark. The walk does not descend into it.
- **Sibling sort.** Children of every directory are sorted by basename using byte-wise lexicographic order. The order is deterministic.
- **Indentation.** Each level of nesting is indented by exactly two ASCII spaces relative to its parent.
- **Output content.** Normal lines contain the basename and (where applicable) the kind mark. The output does not embed full paths or platform-specific path separators in normal lines.

## 11. Project Constraints

- Go standard library only. No third-party tree-printing libraries, no fancy Unicode drawing libraries, no realpath helpers beyond the standard library.
- The walker uses `filepath.WalkDir`. Symlinked directories are not followed.
- Depth is a non-negative integer. There is no unlimited-depth sentinel. Negative depths are invalid.
- Sibling basenames are sorted lexicographically (byte-wise).
- Symlinks are rendered with the trailing `@` mark immediately after the basename. The symlink's target is not displayed in the output. The mark rule keeps the output free of platform-dependent target paths.
- Special files are rendered with the trailing `!` mark and are not traversed.
- The renderer does not read from the filesystem. It consumes a node tree produced by the walker.
- The walker does not write to the output writer. The renderer writes.
- Errors are surfaced with path context. No error is silently hidden or rendered as success.
- The output is deterministic across runs on the same machine and does not embed platform-specific path separators in normal lines.

## 12. Design Questions Before Coding

- Where does the walker live? As a small `walk` function in the printer package, as a method on a tree type, or as a separate package? Which choice keeps the renderer's separation from the walker obvious?
- How is the node tree represented? As a slice of nodes, as a recursive struct with child slices, or as a flat slice with depth fields? Which choice keeps the renderer pure and easy to test?
- How is the lexical sort performed? Through a `sort.Strings` call on the basenames, through `sort.Slice` with a comparator, or through a custom comparator? Which choice keeps the byte-wise contract pinned without overspecifying the primitive?
- Where are the pinned kind marks represented? Decide how the renderer will apply the fixed `@` suffix for symlinks and the fixed `!` suffix for special files without making either mark configurable.
- Where is the pinned indentation unit represented? Decide how the renderer will apply exactly two ASCII spaces per depth without making the unit configurable.
- How is the disappearing-entry error surfaced? Through `fs.WalkDirFunc`'s error return, through a deferred check, or through a separate walk helper? Which choice keeps the walk's error contract obvious?
- How is the writer error surfaced? Through a check on every `Write`, through a buffered renderer that flushes at the end, or through a wrapper writer? Which choice catches partial-write failures without complicating the renderer?
- How is the depth contract pinned? Through a single constructor function that validates depth, through a check at the start of the printer, or through a typed depth value? Which choice makes negative depth impossible to misuse?
- How is the injectable filesystem boundary designed? Through a small interface the walker accepts, through a wrapper around `filepath.WalkDir`, or through a function-typed seam? Which choice keeps walker tests free of permission tricks and platform-sensitive behavior?

## 13. Implementation Milestones

1. Decide the package layout. Keep the walker and the renderer in the same package but as separate types. Keep `main` as a thin wrapper that parses flags, calls the printer, and prints any error.
2. Define the node type. The node carries the basename, the kind (regular file, directory, symlink, special file), the depth, the parent reference, and any extra data the renderer needs.
3. Implement the walker. Use `filepath.WalkDir`. Skip descending into symlinks and special files. Sort sibling basenames lexicographically at each directory level. Recognize symlinks via `Type()&fs.ModeSymlink != 0`, consulting `Info()` only when `Type()` is inconclusive and never in a way that follows the symlink.
4. Implement the depth check at the start of the printer. Negative depth returns an error; non-negative depth proceeds.
5. Implement the missing-root and root-is-not-directory checks. Both return errors naming the root path.
6. Implement the renderer. The renderer takes the node tree and writes the indented tree to the `io.Writer`, with two ASCII spaces per depth level and the pinned marks (`@` for symlinks; `!` for special files). The renderer checks the writer's error after every write.
7. Wire the disappearing-entry error. When `filepath.WalkDir` reports an error for an entry, the printer surfaces it with the entry's path.
8. Wire the writer error. The renderer checks `Write`'s return value and returns an error identifying the renderer and the underlying writer failure. The walk has already finished; only rendering stops.
9. Wire `main`. Accept a positional root argument and a `--depth` flag. Pass the root, the depth, and standard output to the printer. Print errors to standard error.
10. Add tests for every verification case in section 14, split between walker tests (using a per-test temporary directory or an injected filesystem boundary) and renderer tests (using a synthetic node tree).

## 14. Verification Cases the Learner Must Write

Each case is described in natural language. Walker tests use a per-test temporary directory; renderer tests use a synthetic node tree and a `bytes.Buffer`. Tests that would otherwise depend on permission tricks or platform-sensitive behavior use an injectable filesystem boundary and are marked "where practical".

### Ordering and indentation

- A test creates a temporary root with files and directories whose basenames are intentionally non-alphabetical on disk (created in random order) and runs the printer. The output lists the children in lexicographic order, indented by two ASCII spaces per level.
- A test creates a temporary root with files whose names differ only in case (for example `a.txt` and `A.txt`). The output uses byte-wise order, with the uppercase letter sorting before the lowercase letter because of ASCII byte values. The test pins the exact order.
- A test creates a temporary root with a mix of files and directories at the same level. The output lists all of them in the sorted order, with the indentation applied to every entry regardless of kind.

### Depth 0, 1, deeper

- A test creates a temporary root with three levels of nesting. With `depth = 0`, the output contains only the root line.
- With `depth = 1`, the output contains the root and every direct child. No grandchildren are present.
- With `depth = 2`, the output contains the root, the direct children, and the grandchildren. No great-grandchildren are present.
- With `depth = N` for a large `N`, the entire tree is rendered.
- With `depth = -1`, the printer returns an error and writes nothing.

### Empty root

- A test creates an empty temporary directory and runs the printer with `depth = 0`. The output contains only the root's basename on a single line. No error is returned.
- The same test with `depth = 1` produces the same single-line output. No children are listed because the directory is empty.

### Files and directories

- A test creates a temporary root with regular files and subdirectories. The output renders both kinds as their basenames, indented two ASCII spaces per depth level.
- A test creates a temporary root with a deeply nested directory. The output renders every entry along the path with progressively deeper two-space indentation.

### Symlink no-loop

- A test creates a temporary root with a symlink to a regular file inside the root. The output renders the symlink as `basename@`. The walk does not descend into the symlink.
- A test creates a temporary root with a symlink to a directory inside the root. The output renders the symlink as `basename@` and does not descend into it. The symlink's target's contents do not appear in the output.
- A test creates a temporary root with a symlink whose target is the root's parent (a symlink loop). The walk completes without hanging. The output renders the symlink as `basename@` and does not loop.
- A test creates a temporary root with a symlink whose target is a directory outside the root (for example, a sibling temporary directory). The output renders the symlink as `basename@` and does not include any entry from the target.

### Special files

- Where practical, a test creates a temporary root containing a named pipe or other special entry. The output renders the entry with the trailing `!` mark. The walk does not descend into it. (Where the test platform does not allow creating the special entry, the case is marked "where practical".)

### Invalid depth

- A test calls the printer with `depth = -1`. The printer returns an error and writes nothing. The error identifies the depth as invalid.

### Missing root

- A test calls the printer with a path that does not exist. The printer returns an error naming the missing path.
- A test calls the printer with a path that exists but is a regular file. The printer returns an error naming the path and identifying the kind.

### Writer error

- A test injects an `io.Writer` that returns an error on the first write. The printer returns an error identifying the renderer and the underlying writer failure.
- A test injects an `io.Writer` that returns an error after a few successful writes. The printer returns an error. The walk has already finished; the renderer stops.

### Permission and disappearing-entry behavior (where practical)

- Where practical, a test uses an injectable filesystem boundary to simulate an unreadable directory. The printer returns an error naming the directory path.
- Where practical, a test uses an injectable filesystem boundary to simulate a disappearing entry. The printer returns an error naming the entry's path.
- Where neither simulation is practical, the test marks the case "where practical" and pins the behavior through code review of the walker's error surfacing, rather than through a runtime test.

### Renderer separation

- A test feeds a synthetic node tree directly to the renderer and asserts on the rendered output. The test does not use the filesystem.
- A test feeds a synthetic node tree whose entries include a regular file, a directory with children, a symlink, and a special file. The output marks each kind according to the pinned rules.

### Determinism

- A test runs the printer twice against the same temporary root. The two outputs are byte-identical.
- A test runs the printer against a temporary root whose entries are created in randomized order across many runs. The output is always the same.

### Process

- An integration test runs the compiled binary against a per-test temporary directory and confirms the exit code is zero and the rendered tree is on standard output.
- An integration test runs the compiled binary with `--depth=-1` and confirms the exit code is non-zero and standard error names the invalid depth.

## 15. Common Mistakes to Watch For

- **Following symlinked directories.** A custom walker that resolves symlinks via `os.Stat` or otherwise follows the link defeats the no-follow rule. The project requires `WalkDir`'s default behavior plus the basename-only mark.
- **Treating negative depth as "unlimited".** The contract is "non-negative integer only". A negative depth is an error, not an unlimited-depth request.
- **Relying on directory enumeration order.** Children must be sorted before rendering. Iterating the walker's callback order directly produces non-deterministic output.
- **Relying on case-insensitive sort or Unicode-aware sort.** The contract is byte-wise lexicographic order. ASCII sort and Unicode collation are different.
- **Silently hiding a disappearing entry.** An entry that the walker cannot stat must be reported as an error with its path. Treating it as a missing-but-OK entry hides the race.
- **Silently hiding an unreadable directory.** A directory whose contents the user cannot read is an error with the directory's path. Treating it as empty hides the permission problem.
- **Silently hiding a writer error.** A `Write` that returns `0` and a non-`nil` error must stop the renderer. Treating it as "we tried" hides the partial-write failure.
- **Stopping the walk on a writer error.** The walk has already finished by the time the renderer writes. A writer error stops the renderer; it does not roll back a walk that already completed.
- **Mixing rendering with traversal.** The renderer must not call the filesystem. Letting the renderer stat nodes for additional information couples the two responsibilities and breaks the testability split.
- **Embedding full paths or platform-specific separators in normal output lines.** Normal lines contain the basename and (where applicable) the kind mark. Embedding full paths brings platform-dependent separators into the output and breaks the test's stable expectations.
- **Displaying the symlink's target in the mark.** The mark is the trailing `@`. The target path is not displayed; displaying it would bring platform-dependent paths into the output and would require following the symlink to learn the target.
- **Using real home directories or other OS-specific paths.** Tests must use per-test temporary directories.
- **Marking special files as if they were regular files.** Special files (devices, sockets, named pipes) have no children, but they still need the trailing `!` mark so the user knows they are not regular files.
- **Resolving a symlink's target by following the link.** Reading the target with `os.Readlink` (a non-following read) is acceptable only if the target is not displayed in the output; following the link with `os.Stat` defeats the no-follow rule.
- **Treating `filepath.SkipDir` as an error.** `fs.SkipDir` is a normal signal to skip a directory's contents, not a walk failure. The walker must distinguish it from a real error.
- **Sorting with `strings.ToLower` first.** Lowercase-then-sort is a different order than byte-wise sort. The contract is byte-wise.
- **Adding a "do not show hidden files" rule.** No such rule is pinned. Hidden files are rendered like any other entry. Adding the rule silently changes the output.
- **Testing permission failures with `chmod`.** Permission tricks are platform-sensitive. Use an injectable filesystem boundary; mark the case "where practical" when no clean seam is available.

## 16. Topics and References for Study

- A Tour of Go: "Methods", "Errors", "Reading files".
- Effective Go: "Errors", "Data".
- Package documentation: `io/fs` (`DirEntry`, `WalkDirFunc`, `WalkDir`, `FileMode`, `ModeSymlink`, `ModeDir`, `ModeIrregular`, `SkipDir`), `path/filepath` (`WalkDir`, `Walk`, `Base`, `Join`, `Clean`), `os` (`Lstat`, `IsNotExist`), `io` (`Writer`), `sort` (`Slice`, `SliceStable`).
- Tree rendering patterns: search for "Go directory tree printer", "Go filepath.WalkDir lexical order", "Go io.Writer error handling".
- Symlink-safe walking: search for "Go symlink loop walk", "Go WalkDir symlink handling", "Go Lstat vs Stat".
- Injectable filesystem boundaries: search for "Go fs.FS testing", "Go interface seam WalkDir", "Go chmod-free permission test".

## 17. Self-Assessment Questions

1. Why is depth pinned as a non-negative integer rather than as an integer with `-1` meaning unlimited?
2. Why must sibling basenames be sorted byte-wise rather than alphabetically case-insensitively?
3. Why must the renderer be separable from the walker (with the walker never writing to the output writer), and what does the synthetic-node-tree test prove about the renderer?
4. Why must a symlink be rendered with the trailing `@` mark (and not display its target), and what does the symlink-loop test pin?
5. Why must a disappearing entry be reported as an error rather than silently treated as missing?
6. Why must a writer error be surfaced rather than treated as a successful truncated output, and why does the walk not need to stop on a writer error?
7. Why does the output use basenames and indentation rather than full paths, and what does that choice pin about platform separators?
8. Why must the tests use per-test temporary directories rather than real home paths?
9. Why must special files (devices, sockets, named pipes) be marked visibly and not traversed?
10. Why must permission and disappearing-entry tests use an injectable filesystem boundary or be marked "where practical", instead of `chmod` tricks?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every functional requirement in section 8 is satisfied.
- Every verification case in section 14 has a corresponding test, with walker tests and renderer tests split. Permission and disappearing-entry tests use an injectable filesystem boundary or are marked "where practical".
- Depth is a non-negative integer. Negative depths return an error and write nothing.
- Sibling basenames are sorted byte-wise before rendering. The output is deterministic.
- Symlinks are rendered as `basename@` and never followed. Symlink loops are handled without hanging.
- Every error carries its path or context. No error is silently hidden or rendered as success.
- A writer error stops the renderer; the walk, which has already finished, is not restarted.
- The renderer does not call the filesystem. The walker does not write to the output writer.
- Indentation is two ASCII spaces per depth level. Normal lines contain basenames and (where applicable) the kind mark; they do not embed full paths or platform-specific path separators.
- Two runs against the same root produce byte-identical output.
- The package documentation states the depth contract, the lexical sort contract, the symlink mark and no-follow rule, the special-file rule, the indentation unit, and the error-context rule.
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Directory summary at the bottom.** Append a single summary line at the end of the rendered tree counting the regular files and directories shown in the tree. The summary line is omitted on error and is omitted when the writer fails before completion. Do not add counts for symlinks, special files, or sub-totals per directory.
- **Hidden-file filter flag.** Accept a `--all` flag that controls whether entries whose names begin with `.` are rendered. Without the flag, dot-files are skipped from rendering but their parent directory is still rendered. The flag does not change symlink handling or depth semantics. Do not add per-pattern filters or regex-based filters.
