# Project 016 — Todo CLI

## 1. Project Name and Number

- Project **016** — `016_todo_cli`.
- The directory name and number must match exactly.
- This is the first project in the "data, files, and algorithms" level of the plan.

## 2. Project Idea

A command-line todo list that lives entirely in memory for the lifetime of a single process. The program reads command lines from an injected reader one at a time until the reader returns end-of-file or the user types `quit`. Each line begins with one of the four subcommands `add`, `list`, `complete`, or `delete`, followed by its arguments. The program responds to each line in turn, then waits for the next one.

State survives for the duration of the process only. There is no persistence. Two separate process invocations share no state, so anything that depends on cross-invocation state is out of scope.

Each task has a positive numeric ID, a non-empty title, a completion flag, and a creation position that is preserved across operations. Listing shows tasks in a deterministic order pinned by the README. Missing or unknown IDs, unknown subcommands, and malformed commands are reported as errors without mutating the in-memory collection.

The project deliberately separates three concerns: the domain operations on the task collection, the parsing of one command line at a time, and the read-print-loop that drives them. The domain operations are independently callable from tests without going through the loop.

## 3. Why This Project Now?

- Projects 001 through 015 walked through the language surface and through simple structured I/O.
- None of them built a small in-memory domain that another program could call.
- Project 016 is the first project whose operations must be exercisable from a test independently of the loop, because the same domain layer is reused by project 017 (JSON persistence) with a different I/O shell.

- The project also introduces the discipline of stable IDs: the plan's verification rules require that an ID, once issued, is never reused after deletion, and that listing order is deterministic.
- The learner practices those decisions now, while the project is small, so that the same discipline scales naturally when the same model becomes a persisted document in 017.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 016 therefore requires:

- Completion of **015** (CLI Counter), including the discipline of separating resource lifecycle from the code that uses the resource.
- No prior knowledge of HTTP, databases, generics, or concurrency.
- Familiarity with command-line argument parsing is helpful but not required; this project introduces the pattern of reading command lines from an injected reader.

## 5. What You Must Know Before Starting

- What a struct is in Go: a typed bundle of named fields, and how `t.Field = value` mutates a struct held through a pointer.
- What a slice is and how appending to it grows it; that the underlying array may change, so do not keep pointers into a slice while appending.
- The difference between a value receiver and a pointer receiver on a method, and why a stateful collection needs a pointer receiver.
- How `switch` on a string value matches exactly one case; an unknown subcommand falls into the `default` branch.
- How to return multiple values from a function, including an `error`, and why the caller is expected to check it before using the other values.
- That `bufio.Scanner` over an `io.Reader` returns one line at a time. Each successful `Scan` advances the reader; `Scan` returns `false` when the reader reaches EOF or an error.
- That `os.Stdin` is an `io.Reader` that production code wires to the scanner, while tests wire a `strings.Reader` or `bytes.Buffer` holding the command sequence.

## 6. Explanation of New Concepts

### Concepts

#### A loop driven by an injected reader

- The program runs a read-print loop.
- Production wires `os.Stdin` to a buffered scanner; tests wire a `strings.Reader` or `bytes.Buffer` that holds a fixed command sequence.
- The same loop body runs in both cases.
- The loop reads one line, dispatches it to the domain, prints the result, and reads the next line.
- The loop terminates cleanly when the scanner returns EOF; it also terminates when the user types `quit`.

- This seam is the same one project 015 introduced for the tick source: the production wiring is real, the test wiring is a fake.
- Here the seam is for input rather than for ticks, but the pattern is the same.

#### EOF ends cleanly

- EOF is the normal end of the loop.
- The scanner returns `false` because the underlying reader is at EOF, the loop exits, and the program returns from `main` with exit code zero.
- No special "exit" command is required; closing the input is enough.

#### `quit` is not a mutation

- The line `quit` ends the loop but is not a domain command.
- It does not add a task, does not delete a task, and does not change the collection.
- The test pins this: after `quit`, the collection still contains exactly the tasks it contained before the `quit` line was read.

#### Unknown or malformed commands continue the session

- A line that begins with anything other than the four known subcommands, or a line that fails to parse (for example, `complete` with no ID, or `add` with no title), produces an error line on standard error and the loop continues with the next line.
- The session does not exit on the first error.
- This invariant lets a test prove that earlier commands succeeded and that the failed command did not mutate state.

#### A domain separated from the loop

- The domain operations on the collection are plain Go.
- They take the current collection and arguments, return the updated collection plus a result plus an error, and never touch the reader, the writer, or the loop.
- A table-driven test calls the domain directly, asserts on the returned collection, and does not have to set up a scanner or capture standard output.

- This separation matters for one concrete reason: the tests in section 14 must be able to pin the invariant "errors never mutate state" by calling the domain in isolation.
- A test that runs through the loop has to capture standard output and parse it; a test that calls the domain directly inspects the returned collection and moves on.

#### Stable IDs and stable order

- The collection assigns IDs in a deterministic way: monotonically increasing positive integers, never reused after deletion.
- A counter remembers the next ID to issue; every `add` increments it; every `delete` removes a task but does not return its ID to a pool.
- The benefits are that an old ID still maps to "does not exist" after deletion, that listing order by ID is identical to creation order, and that a future JSON round trip in project 017 keeps the same IDs across runs.

- Listing order is the order of insertion, which is identical to ascending ID order under this policy.
- If the learner ever decides to expose a "sort by title" command, that sort is a separate view and does not change the canonical order.

#### Errors that do not mutate state

- Every domain operation that fails — empty title, missing ID, unknown ID, malformed command — returns an error without changing the collection.
- This invariant lets a test write one negative case and one positive case without resetting state in between.
- A test that runs `add a`, then `complete 999` (which fails), then `add b` confirms that the collection contains exactly `a` and `b` and that the failed `complete 999` did not consume the next ID.

## 7. Learning Objective

After completing this project the learner can:

- Build a small in-memory domain whose operations take and return a collection plus an error and never reach into a reader, a writer, or a loop variable.
- Drive a read-print loop from an injected reader, terminating cleanly on EOF and on the `quit` command.
- Choose and document a stable ID policy for a small collection, and keep that policy consistent across add, complete, delete, and list.
- Distinguish complete from incomplete tasks in the listed output in a way a test can assert on.
- Return clear errors for empty title, missing ID, unknown ID, malformed command, and unknown subcommand, and confirm in a test that the collection is unchanged on each error path.
- Write table-driven tests that drive the domain layer directly and a separate set of tests that drive the loop end-to-end through an injected reader.

## 8. Functional Requirements

1. Define a task type with at least four fields: a positive integer ID, a non-empty title string, a completion boolean, and a stable creation order that can be derived from the ID under the chosen policy.
2. Define a small domain whose operations on a task collection are: add a task, list all tasks, mark a task complete by ID, delete a task by ID. The domain returns the updated collection plus a result plus an error; it does not read from any input, does not write to any output, and does not mutate the collection on an error path.
3. Define a read-print loop with an injected input reader plus separate injected normal-output and error-output writers. It reads one line at a time, dispatches it, writes a success or command-level error to the appropriate writer, and continues.
4. The loop recognizes four subcommands: `add <title...>`, `list`, `complete <id>`, `delete <id>`. Any other first argument is an unknown subcommand.
5. The `add` line consumes the entire remainder of the line after `add` as the title. A title that contains internal spaces is accepted; a title that is empty or whitespace-only is rejected.
6. The IDs issued by `add` are positive integers that increase monotonically across the life of the process and are not reused after a task is deleted.
7. The `list` line writes all tasks in creation order, with each line clearly distinguishing a complete task from an incomplete one. A header line indicating the count is acceptable; the test pins the wording.
8. The `complete <id>` line marks the named task as complete. Completing a task that is already complete is not an error and does not change the collection beyond the no-op.
9. The `delete <id>` line removes the named task from the collection and does not reuse its ID for a future `add`.
10. The line `quit` ends the loop. The collection is not mutated by `quit`.
11. End-of-file on the input reader ends the loop. The collection is not mutated by EOF.
12. Empty title, missing or non-numeric ID, unknown ID, unknown subcommand, and any other malformed line each return a clear error message. None of these error paths may mutate the collection.
13. The loop continues after every error. A failed line does not end the session; the next line is processed normally.
14. After scanning stops, distinguish clean EOF from an underlying read error. Clean EOF and `quit` end the session successfully even if earlier command-level errors were already reported; an input-reader or output-writer failure stops the session and returns a process-level error.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- A stream of command lines from an injected reader. Each line is one command. Empty lines are treated as malformed input.

#### Outputs

- For `add`: a single line confirming the new task, including its ID.
- For `list`: one line per task, in creation order, with a clear marker for complete vs incomplete, plus an optional header.
- For `complete`: a single line confirming the change, including the ID.
- For `delete`: a single line confirming the deletion, including the ID.
- For `quit` or EOF: no output.
- For any error: a single line on standard error describing the failure.

#### Example text-only success session (one process, many lines)

```
add buy milk
add write report
complete 1
list
quit
```

Expected output on standard output:

```
Added task 1: buy milk
Added task 2: write report
Completed task 1.
2 tasks:
[ ] 2: write report
[x] 1: buy milk
```

#### Example text-only error session (one process, many lines)

```
add
complete 99
party
complete 2
list
```

Expected output on standard error:

```
Error: title must not be empty.
Error: no task with ID 99.
Error: unknown command: party.
```

Expected output on standard output:

```
Completed task 2.
1 task:
[x] 2: write report
```

The session continues after every error. The `complete 2` line still succeeds because the failed `complete 99` did not mutate the collection.

## 10. Rules and Edge Cases

- **Empty title.** The `add` line with no remaining text after the subcommand is rejected with a clear error. The collection is unchanged.
- **Whitespace-only title.** A title whose trimmed length is zero is rejected. Whitespace handling is the learner's choice; the rule is that no task may have an empty title.
- **Title containing leading or trailing spaces.** The behavior is the learner's choice (preserve, trim, or reject). The README pins one rule and the test enforces it.
- **Repeated completion.** Completing an already-complete task returns success but does not change the collection beyond the no-op.
- **Deletion of an unknown ID.** Returns an error. The collection is unchanged.
- **Completion of an unknown ID.** Returns an error. The collection is unchanged.
- **Unknown subcommand.** Returns an error. The collection is unchanged.
- **Malformed line.** A line that does not match any known subcommand, or a known subcommand followed by an unparseable argument, returns an error. The collection is unchanged. The session continues.
- **Listing an empty collection.** `list` on an empty collection writes a single line such as "no tasks" and returns to the loop. The loop continues.
- **Non-numeric ID.** `complete abc` and `delete abc` are rejected as a malformed ID, with a clear error message.
- **Negative or zero ID.** Rejected as an out-of-range ID, with a clear error message.
- **ID reuse after deletion.** A future `add` after a `delete` issues an ID strictly greater than every ID ever issued in the process. The collection never contains two tasks with the same ID at any moment.
- **Listing order.** Tasks are always listed in ascending ID order, which under the chosen policy is identical to creation order.
- **EOF.** EOF on the input reader ends the loop cleanly. The collection is not mutated by EOF.
- **`quit`.** The line `quit` ends the loop. The collection is not mutated by `quit`.

## 11. Project Constraints

- Go standard library only. No third-party argument-parsing libraries.
- The domain operations must be callable directly from tests without going through any input reader, any output writer, or the loop. The learner chooses the package layout and the function signatures; the README does not prescribe them.
- The loop uses an injected input reader and an injected output writer. Production wires `os.Stdin` and `os.Stdout`; tests wire fakes.
- No persistence. The collection dies with the process. Persistence is the subject of project 017.
- No concurrency. A single goroutine owns the collection.
- No cross-process state. Two separate process invocations share no state. Tests do not assume any state survives across invocations.
- No third-party dependencies in `go.mod`.

## 12. Design Questions Before Coding

- Where does the collection live? Inside a package-level variable owned by `main`, inside a struct owned by `main`, or behind a small type the domain layer exposes? Which choice lets the test build a fresh collection per case without touching globals?
- How is the next ID stored and advanced? On the collection type, on a dedicated "store" struct, or in a small wrapper? Which choice keeps the "never reuse IDs" rule obvious to read?
- How is a complete task visually distinguished from an incomplete one? A leading marker, a trailing word, two separate columns? Which format survives a test that compares line-by-line?
- How is `add`'s title reconstructed from the line? One trim-then-string, one separator-joined string, or one explicit slice? Which choice keeps the test stable when the title contains spaces?
- How is "unknown subcommand" reported? A constant message, an `errors.New`, a typed error? Which choice lets the test pin the wording once and reuse it?
- How will the loop keep recoverable command errors separate from terminal I/O failures? A typo is reported and the next line runs, while a broken reader or writer ends the session with a process-level error.
- How is the loop terminated by EOF? Through `scanner.Scan` returning `false`, through a separate `io.EOF` check, or through the scanner's `Err` method? Which choice handles both clean EOF and read errors correctly?

## 13. Implementation Milestones

1. Decide the package layout. A small package for the domain (the collection type and the operations on it) and a thin `main` package for the loop and the standard I/O wiring.
2. Define the task type with the required fields. Decide whether completion is a `bool` or a richer status; the simplest correct choice is a `bool`.
3. Define the collection type and the next-ID field. Decide how the next ID is held and advanced.
4. Implement the domain `add` operation. It returns the updated collection plus the new task plus a nil error, or the unchanged collection plus an error if the title is empty.
5. Implement the domain `list` operation. It returns a stable, deterministic slice of the tasks in creation order plus a nil error.
6. Implement the domain `complete` operation. It returns the updated collection plus a result plus a nil error, or the unchanged collection plus an error for an unknown or invalid ID. Repeated completion is a no-op success.
7. Implement the domain `delete` operation. It returns the updated collection plus a result plus a nil error, or the unchanged collection plus an error for an unknown or invalid ID. The deleted ID must never reappear.
8. Build the loop. It reads and dispatches lines, writes successes and command errors to separate injected writers, ends successfully on `quit` or clean EOF, and checks the scanner error after scanning so an underlying read failure is not mistaken for EOF.
9. Wire the program's standard I/O: production wires `os.Stdin` to the input reader and `os.Stdout`/`os.Stderr` to the output writer; tests wire a `strings.Reader` or `bytes.Buffer` to the input and a `bytes.Buffer` to the output.
10. Confirm that one session can run `add`, `add`, `complete 1`, `list`, `quit` and that a second process invocation starts with an empty collection regardless of the first session's contents.
11. Write the table-driven test that drives the domain layer directly across every positive and negative case in section 14.
12. Write a separate set of loop tests that drive the loop end-to-end through an injected reader, capturing output through an injected writer, and asserting on the captured bytes.

## 14. Verification Cases the Learner Must Write

### Required Cases

Each case is described in natural language. The domain tests call the domain functions directly. The loop tests drive the loop end-to-end through an injected reader and capture output through an injected writer. No test depends on real user input, real home directories, or wall-clock time. No test assumes state survives across process invocations.

#### Domain — happy paths

- Adding a task with a non-empty title returns a collection of length one containing that task, with the chosen ID policy assigning ID `1` to the first task.
- Listing a freshly added collection returns one task in the same order it was added.
- Adding a second task yields IDs `1` and `2` in creation order; listing returns both in that order.
- Completing a task with an existing ID returns the same collection with that task's completion flag set.
- Deleting a task with an existing ID returns a shorter collection that no longer contains that task.
- After a deletion, adding a new task yields an ID strictly greater than every ID ever issued, so the deleted ID is not reused.
- A repeated completion of an already-complete task is a successful no-op: the collection is unchanged but no error is reported.

#### Domain — error paths

- Adding a task with an empty title returns an error and leaves the collection unchanged.
- Adding a task with a whitespace-only title (per the rule the learner pinned) returns an error and leaves the collection unchanged.
- Completing a non-existent ID returns an error and leaves the collection unchanged.
- Deleting a non-existent ID returns an error and leaves the collection unchanged.
- Completing with a non-numeric or negative or zero ID returns an error and leaves the collection unchanged.
- Deleting with a non-numeric or negative or zero ID returns an error and leaves the collection unchanged.

#### Loop — happy paths

- A loop test feeds the sequence `add buy milk`, `add write report`, `complete 1`, `list`, `quit` through the injected reader. Standard output contains the two confirmations, the completion confirmation, and the two-task listing, in that order. The session ends after `quit`. Standard error is empty.
- A loop test feeds `add hello world` (title with a space) and then `list`. The output contains one task whose title is `hello world`.
- A loop test feeds `list` with an empty collection. The output contains the documented empty-listing line. Standard error is empty.

#### Loop — error paths and session continuity

- A loop test feeds `add` (no title) followed by `list`. Standard error contains a clear message for the empty title. Standard output contains the empty-listing line because the failed `add` did not mutate the collection.
- A loop test feeds `complete 999` against an empty collection followed by `add buy milk` and `list`. Standard error contains a clear message for the unknown ID. Standard output shows exactly one task. The failed `complete 999` did not consume the next ID.
- A loop test feeds `party` (unknown subcommand) followed by `add buy milk` and `list`. Standard error contains a clear message naming `party`. Standard output shows exactly one task. The session continues after the unknown command.
- A loop test feeds `complete abc` (non-numeric ID). Standard error contains a clear message. The collection is unchanged.
- A loop test feeds `complete -1` and `complete 0`. Both return a clear error. The collection is unchanged.
- A loop test feeds a line that is just whitespace. Standard error contains a clear malformed-line message. The collection is unchanged.

#### Loop — termination

- A loop test feeds the single line `quit`. Standard output is empty. Standard error is empty. The loop ends.
- A loop test feeds no lines (EOF immediately). Standard output is empty. Standard error is empty. The loop ends with exit code zero.
- A loop test feeds `add buy milk` then closes the input. Standard output contains the addition confirmation. The loop ends with exit code zero.
- A loop test feeds `add buy milk`, `quit`, and another line that the test cannot reach because the loop has ended. The unreachable line is not processed. The test confirms this by asserting that standard output contains only the addition confirmation.
- A custom reader that returns an error after a valid command causes the command's result to be emitted, then ends the session with a process-level read error rather than treating the failure as EOF.
- A writer that fails while receiving output ends the session with a process-level write error; the failure is not silently ignored.

#### Loop — error does not end the session

- A loop test feeds a sequence that interleaves successful commands, failed commands, and `quit`. The collection at the end of the loop contains exactly the tasks that the successful commands added, and the failed commands have no effect on the collection. The test pins this by capturing the final `list` output.

#### Stability

- A test runs `add a`, `add b`, `delete 1`, `add c`, `list` through the domain. The final collection contains exactly two tasks: the original `b` (ID `2`) and the new `c` (ID strictly greater than `2`, in that order). ID `1` does not appear.
- A test runs the operations in random order across many cases and confirms that the collection's tasks are always strictly increasing by ID, with no gaps filled.

#### Cross-invocation independence

- A loop test feeds `add buy milk` and `quit`. A second loop test, in a separate process-equivalent invocation (a fresh domain and a fresh reader), starts with an empty collection. The first session's tasks are not visible in the second.

## 15. Common Mistakes to Watch For

- **Mixing the domain and the loop.** If the domain reads from the injected reader or writes to the injected writer, tests cannot call it directly and the project's separation collapses.
- **Reusing IDs after deletion.** A future `add` after a `delete` issues the deleted ID again, which violates the stability rule and breaks the contract that 017 will rely on.
- **Mutating the collection on an error path.** Returning an error but still appending to the slice, or marking a task complete and then reporting "unknown ID", breaks the "errors never mutate state" invariant.
- **Ending the session on the first error.** The loop must continue after every error. A failed line is a local event, not a reason to exit.
- **Treating `quit` as a mutation.** `quit` ends the loop but does not change the collection.
- **Treating EOF as an error.** EOF is the normal end of the loop. The scanner returns `false` because the reader is at EOF, not because anything went wrong.
- **Sorting the slice in place as part of `list`.** If `list` sorts the underlying slice, the canonical order changes; tests that compare slices see inconsistent ordering across runs.
- **Treating an already-complete task as an error.** Repeated completion is a successful no-op, not an error.
- **Using a signed integer for ID with no policy on negatives.** A negative or zero ID is rejected, not "completed" or "deleted" silently.
- **Using `fmt.Scan` or a REPL.** `fmt.Scan` is not appropriate for line-based command input. `bufio.Scanner` over an injected reader is the right pattern.
- **Reading the same global collection in every test.** A test that mutates a global leaks state into the next test. Each test must build a fresh collection.
- **Writing to `os.Stdout` inside the domain.** The domain returns data; the loop writes it. Conflating the two makes negative-case tests impossible to write cleanly.
- **Assuming cross-process state.** State lives in memory only. A test that depends on a previous process's state will fail.

## 16. Topics and References for Study

- A Tour of Go: "Flow control: switch", "More types: structs", "Methods and interfaces".
- Effective Go: "Data", "Errors".
- Package documentation: `bufio` (`Scanner`, `Scanner.Scan`, `Scanner.Err`), `errors` (`New`, `Is`, `As`), `fmt` (`Errorf`, `Fprintln`, `Fscanf`), `io` (`Reader`, `Writer`), `os` (`Stdin`, `Stdout`, `Stderr`), `strconv` (`Atoi`, `ParseInt`), `strings` (`NewReader`, `Reader`).
- Go testing patterns: search for "table-driven tests Go", "Go CLI loop test injected reader", "test main package".
- Read-print loop patterns: search for "Go REPL pattern", "bufio.Scanner line input", "injected reader CLI".

## 17. Self-Assessment Questions

1. Why does the project separate the domain from the loop, and what does that separation buy the test suite?
2. Why must an ID, once issued, never be reused after deletion, even if it makes the ID counter grow without bound?
3. Why is the listing order identical to the creation order under the recommended ID policy, and what would change if the policy changed?
4. Why does a repeated `complete` succeed but `complete` of an unknown ID fail?
5. Why must the loop continue after every error instead of exiting, and how does a test prove that earlier commands still worked after a later command failed?
6. Why is "errors never mutate state" an invariant worth preserving, and how does a test enforce it?
7. What does the loop test gain that a pure domain test cannot provide, and what does the domain test gain that a loop test cannot?
8. Why is EOF not treated as an error, and what would a loop that returned non-zero on EOF imply?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- [ ] The README's 19 sections are present in order; this file is the reference.
- [ ] Every functional requirement in section 8 is satisfied.
- [ ] Every verification case in section 14 has a corresponding test, and the tests do not depend on real user input, real home directories, or cross-process state.
- [ ] The domain operations are callable from tests without touching any input reader, any output writer, or the loop.
- [ ] ID stability holds: across a representative randomized run, no ID is ever reused, and the listing order is always ascending ID order.
- [ ] Error paths return errors and leave the collection unchanged; a test confirms this for every error case in section 14.
- [ ] The loop continues after every error; a test confirms this for at least the empty-title, unknown-ID, and unknown-subcommand error cases.
- [ ] The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Filter flag on `list`.** Accept an optional flag on the `list` line that lists only incomplete tasks or only complete tasks. The flag must not change the canonical order, only the slice that is printed.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisite:** [Project 015 — CLI Counter](../../01-foundations/015_cli_counter/README.md#20-prerequisite-based-documentation-guide).

Read the linked guide first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **Standards and concept references:** [A Tour of Go: structs, methods and interfaces](https://go.dev/tour/methods/1).

### Project-specific learning focus

- **Learn now:** domain modeling, command parsing, stable identifiers, state transitions, and testing an interactive loop with injected streams.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
