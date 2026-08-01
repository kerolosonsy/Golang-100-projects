# Project 009 — System Info CLI

## 1. Project Name and Number
- Number: **009**, level 1 (language basics and CLI).
- Folder name in the table: **`009_system_info_cli`**, matching `01-foundations/009_system_info_cli/`.
- Kind: a small terminal program that prints information about the runtime environment and about a small, explicit allowlist of environment variables, controlled through command-line flags. Output is deterministic for fixed inputs.

## 2. Project Idea
Build a terminal program that prints the operating system name, the architecture, the runtime's Go version, and the value of a few named environment variables. The set of environment variables is fixed by an explicit allowlist in the program; the program never iterates over every variable in the environment. Each piece of information is selected through a documented flag, with a documented default behaviour for when the flag is absent. The output is formatted in a fixed, stable layout.

## 3. Why This Project Now?
- First project in the path that reads command-line flags. The flag package is reused heavily from 046 onward.
- First project that touches environment variables. The pattern of an explicit allowlist, rather than a dump, is the foundation for every later project that handles configuration safely (050, 060, 071).
- First project that mixes information from the standard library's runtime, OS, and flag packages. The boundaries between these packages are part of the lesson.
- Reinforces the discipline of presenting only what the program is designed to show, never everything that happens to be available.

## 4. Prerequisites
- Project **008** (`008_rock_paper_scissors`). Comfort with typed values, helper functions, and clear separation of concerns should already be in place.
- Familiarity with reading from standard input and writing to standard output.
- No new tools or libraries beyond the standard library used in earlier projects.

## 5. What You Must Know Before Starting
- Command-line flags are values the user passes to the program before it starts. They are not the same as runtime input read from standard input.
- A flag has a name, a type, a default value, and a usage string. The default is what the program uses when the flag is absent.
- Environment variables are name-value pairs held by the operating system and inherited by each process. They are not a substitute for flags, and flags are not a substitute for environment variables. They serve different roles.
- An environment variable may be present and empty, present and non-empty, or absent. All three cases are distinct and have different meanings.
- "Print everything in the environment" is unsafe as a default behaviour. Many environment variables carry sensitive values (tokens, credentials, paths to private keys, contents of `.env` files that have been sourced into a shell). A program that iterates over all variables broadens its handling of sensitive data and makes it easy to write such a value into a log line, an error message, or a structured output field by accident. The program must not iterate over all variables as a default.
- Reading a single, named environment variable by itself does not expose the variable's value beyond the process that already owns the environment; the concern is what the program does with the value afterwards, including where it might end up in output or logs. Reading the same value from the allowlist narrows that handling by design.
- An allowlist is a short, explicit list of names. The program reads only the names on the list; an absent name on the list is reported as absent, not as an error.
- The program's output must be deterministic: the same flags and the same environment produce the same output, run after run. Randomness, time-of-day, or environment leakage that changes between runs is a bug.
- "Format" here means the layout of each line: a stable key, a separator, a value, and a trailing newline. Different formats for different lines break tools that parse the output.

## 6. Explanation of New Concepts
- The `flag` package: where it lives, how a flag is declared with its default, how `flag.Parse` consumes `os.Args`, and what happens to unrecognised flags. The package is the canonical way to read command-line options in a small program.
- The `runtime` package: a small set of functions that report information about the running Go program itself, including the operating system target, the architecture, and the Go version. None of these are properties of the host hardware; they are properties of the compiled binary.
- The `os` package: a broad collection of helpers for the operating system surface. For this project, the relevant subset is `os.Args`, environment variable access, and the program's exit codes.
- Environment variable access: there is one function to read a single named variable and one function to read all variables. The first is the safe path; the second is the unsafe path that this project avoids.
- The difference between a present-but-empty variable and a missing one. Both produce an empty string from the read function; the `Lookup` family distinguishes them. The choice between the two is yours and must be documented.
- Determinism: the program's output does not change between two runs with the same inputs. Time, randomness, and any variable not in the allowlist do not influence the output.
- Program boundaries: the part of the program that parses flags, the part that gathers information, the part that formats each line, and the part that writes the output are four distinct concerns.

## 7. Learning Objective
By the end of the project you should be able to:
- Declare flags with explicit defaults and parse them in the canonical Go style.
- Read the value of a single, named environment variable and distinguish present-but-empty from absent.
- Never iterate over all environment variables in user-facing output.
- Produce a stable, line-oriented output format that does not change between runs.
- Explain why defaulting to an environment dump is unsafe and what kind of information typically appears in a process environment.
- Separate flag parsing, information gathering, formatting, and writing into distinct steps.

## 8. Functional Requirements
- F1: The program defines a documented flag for each piece of information it can show. A typical baseline set includes: show the operating system, show the architecture, show the Go version, show a specific environment variable, and a flag to control whether all of the above are printed together.
- F2: Each flag has a documented default. The defaults compose into a documented overall default behaviour, stated in this README (for example, "by default the program prints operating system, architecture, and Go version, and does not read any environment variable").
- F3: The program's output for a fixed set of flags and a fixed environment is byte-for-byte reproducible across runs. No time stamp, no random value, no nondeterministic ordering of lines.
- F4: Environment variable access goes through an explicit, hard-coded allowlist of names. The allowlist lives in the program source. The program reads only the names on the list.
- F5: The program never calls any function that returns all environment variables, never iterates over the environment, and never prints a list of names that happen to be set.
- F6: A requested environment variable that is absent is reported with a documented marker (for example, `(unset)`); a present-but-empty variable is reported with a documented marker (for example, `(empty)`). The two cases are distinguishable in the output.
- F7: The output layout is fixed. Each line is a key, a separator, and a value, terminated by a newline. The same format is used for every line.
- F8: Unknown flags are rejected by the flag package with a clear error; the program does not silently accept unrecognised flags.

## 9. Inputs and Outputs
**Inputs**:
- Command-line arguments consumed by the flag package. Each flag has a documented name, a documented type (typically boolean for show/hide toggles and string for the environment variable name), and a documented default.
- The process environment, read only for the names on the allowlist.

**Outputs**: text printed to standard output, one line per piece of information requested. Text-only examples:

- Default flags, no environment override:
  - One line for the operating system: a fixed key, a fixed separator, and the value reported by the runtime.
  - One line for the architecture: the same layout.
  - One line for the Go version: the same layout.

- A flag that requests a specific environment variable by name, with the variable set to a non-empty value:
  - One line for the requested variable, with the layout described above and the actual value.

- The same flag, with the variable absent:
  - One line for the requested variable, with the layout described above and the documented `(unset)` marker.

- The same flag, with the variable present but empty:
  - One line for the requested variable, with the layout described above and the documented `(empty)` marker.

- A flag that asks for "all" or "everything":
  - One line per piece of information in the program's documented overall set, in a documented order.

- Unknown flag on the command line:
  - The flag package prints its own usage error and the program exits with a non-zero status. No partial output.

## 10. Rules and Edge Cases
- The allowlist is hard-coded. Adding a new variable to the allowlist is a code change, not a configuration change. This is deliberate; it forces a review whenever a new variable becomes visible.
- Reading an absent variable does not produce an error in the program's logic; it produces the documented `(unset)` marker. The same applies to a present-but-empty variable, which produces the documented `(empty)` marker.
- The output does not include a time stamp, a process identifier, a working directory, or any other piece of information that varies between runs unless the user has explicitly requested it through a flag.
- The output never includes a variable's name unless that variable is on the allowlist. There is no fallback or "show what is set" mode.
- The format of each line is fixed: the same key, the same separator, the same value formatting. Variations in punctuation, spacing, or quoting are bugs.
- A flag that takes a value (such as a string) is validated to the extent the program documents. An empty value for such a flag has a documented behaviour.
- The program's exit status is non-zero only on flag parsing errors or on other documented failures. A successful run prints the requested lines and exits zero.

## 11. Project Constraints
- Libraries: the standard library only. `flag`, `runtime`, `os`, and `fmt` are the relevant packages.
- Prohibited: any external package that claims to dump or enumerate environment variables.
- Prohibited: any call that returns all environment variables or any iteration that reads variables the allowlist does not name.
- Prohibited: writing any environment variable, mutating the process environment, or shelling out to read configuration that should come from flags.
- Persistence: none.
- Network: none.
- Tests: optional in code; the verification section below lists scenarios the learner runs manually or as table-driven tests if tests are added.

## 12. Design Questions Before Coding
- What is the exact set of flags? Each flag is a deliberate decision about what the program shows; the set is small and explicit.
- What is the documented default behaviour? "Show a fixed baseline set, do not read any environment variable" is one reasonable default; "show nothing, require at least one flag" is another. The choice is yours, but the choice is documented.
- How will you represent the allowlist? A slice of strings, a small map of display-name to environment-name, or another shape? The shape affects how the program prints each entry.
- How will you distinguish present-but-empty from absent? The `Lookup`-style functions distinguish them; the read-and-test-empty approach does not. Pick deliberately.
- What is the exact format of an output line? A `key = value` layout, a `key: value` layout, or another layout? Pick one and apply it everywhere.
- How will you separate the flag declarations from the output logic? Flag declarations at the top of the entry point and the gathering logic in helpers is one common shape; a struct that holds the parsed flags is another.
- How will you handle the "all" or "everything" case? Either as a separate flag that triggers a fixed sequence of prints, or as a default that simply runs every flag. The two designs have different ergonomics.

## 13. Implementation Milestones
1. M1: Create the source file with the minimum required for a Go program to compile and run; verify with a build and run cycle.
2. M2: Declare the documented set of flags with their types and defaults. Verify that the flag package parses the command line without errors on a smoke run.
3. M3: Implement a function that reads a single named environment variable and returns either its value, an `(empty)` marker, or an `(unset)` marker, with the three cases distinguishable.
4. M4: Implement the allowlist: a hard-coded set of names. Reading any name not on the allowlist is not possible because the program does not have a path that does so.
5. M5: Implement the format function that turns a key and a value into a single line of the documented layout. Apply the same function to every output line.
6. M6: Implement the "all" or "everything" behaviour, if your design includes it. The sequence of lines is fixed and documented.
7. M7: Run the program under several documented flag combinations and confirm that the output is deterministic across two consecutive runs.
8. M8: Run every verification scenario from section 14 and confirm the program behaves as your design specifies.

## 14. Verification Cases the Learner Must Write
- Default flags: prints exactly the documented baseline set, in the documented order, with no environment lines.
- A flag that requests a specific environment variable, with the variable set to a non-empty value: the line shows the value, not a marker.
- A flag that requests a specific environment variable, with the variable absent: the line shows the documented `(unset)` marker.
- A flag that requests a specific environment variable, with the variable present but empty: the line shows the documented `(empty)` marker, distinguishable from `(unset)`.
- The "all" or "everything" flag, if your design includes it: prints the documented sequence of lines.
- Two consecutive runs with identical flags and identical environment: output is byte-for-byte identical.
- An unknown flag on the command line: the flag package reports the error and the program exits non-zero; no partial output.
- A flag that takes a string value, with an empty value provided: documented behaviour, no crash.
- A run with an environment that contains a sensitive-looking variable name: the variable is not printed, because it is not on the allowlist. The behaviour is verified by inspection, not by output.
- A run where the allowlist contains an entry that happens to be unset: the printed line shows the `(unset)` marker, not a panic and not a stray empty value.

## 15. Common Mistakes to Watch For
- Iterating over all environment variables. The iteration broadens handling of sensitive values and makes it easy for one of them to end up in a log line, an error message, or an output field by accident, even when a filter exists.
- Confusing "absent" with "present-but-empty". Both produce empty strings from a naive read; the `Lookup`-style functions distinguish them. Choose deliberately.
- Printing the value of an environment variable without a marker when it is absent. The user cannot tell whether the program failed to read or whether the variable simply was not set. Markers make the difference explicit.
- Including a time stamp or a process identifier in the output. The output is supposed to be deterministic; a time stamp breaks that promise.
- Mixing two output formats. Each line must use the same layout.
- Letting the flag defaults grow undocumented. Every default is a design choice and must be in the README.
- Letting the allowlist be configurable. A configurable allowlist is a re-introduction of the dump problem in a smaller form.
- Treating "all" or "everything" as "every environment variable that happens to be set". The two are not the same; the program's documented set is a small subset.

## 16. Topics and References for Study
- A Tour of Go: Packages, basic types, and the introduction to command-line arguments.
- Effective Go: Names, control structures, and the section on flags.
- The `flag` package documentation: how to declare flags, how to parse, and what happens on errors.
- The `runtime` package documentation: the `GOOS`, `GOARCH`, and version-reporting functions.
- The `os` package documentation: environment variable access functions, including the `Lookup` family that distinguishes present-but-empty from absent.
- Search terms: `Go flag package tutorial`, `Go runtime GOOS GOARCH`, `Go os.Getenv vs LookupEnv`, `Go environment variable safety`.

## 17. Self-Assessment Questions
1. The allowlist is hard-coded. Why is a hard-coded allowlist safer than a configurable allowlist, even when the configurable form is convenient?
2. A reviewer suggests adding a `--show-all-env` flag that prints every environment variable. Explain why this suggestion is incompatible with the project's constraints, regardless of how the implementation is written.
3. The program distinguishes `present-but-empty` from `absent`. Why is that distinction worth the extra code?
4. Two consecutive runs produce byte-for-byte identical output. Name three kinds of information that would break that promise if added without a flag.
5. The flag package rejects unknown flags. What is the user experience trade-off of accepting and ignoring unknown flags instead, and why does the canonical Go style reject them?
6. A future requirement adds a new piece of information to the program's output, such as the number of CPUs. What changes in the program and what stays the same? Which change is a design change and which is purely mechanical?
7. Walk through what your program prints when the environment contains a variable that is not on the allowlist, such as a credential-style name. Why is "nothing about that variable" the safe answer, and how is that answer enforced by the program's structure?

## 18. Definition of Completion
- The program compiles and runs without compile errors.
- Every scenario in section 14 produces the behaviour documented in your code.
- No panic occurs in any documented scenario, including unknown flags and absent variables.
- The allowlist is hard-coded in the program source. No path in the program reads an environment variable outside the allowlist.
- The output is deterministic: two runs with the same inputs produce byte-for-byte identical output.
- The output layout is consistent across every line.
- The program's documented defaults match its actual defaults.

## 19. Optional Extensions
- Optional 1: Add a flag that prints the names of the allowlist entries that are currently set in the environment, but never their values. The flag is informational and explicitly safe.
- Optional 2: Add a flag that selects among a small, predefined set of output formats (for example, `key=value` lines, a JSON object, or a table layout). Each format is implemented once; the flag's contract is documented; no new environment access is added.
