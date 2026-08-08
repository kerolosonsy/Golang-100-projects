# Project 001 — Hello CLI

## 1. Project Name and Number

- Number: **001**, level 1 (language basics and CLI).
- Folder name in the table: **`001_hello_cli`**, matching `01-foundations/001_hello_cli/`.
- Kind: a small interactive terminal program that asks for a name and an age, then prints a personalised greeting that reports how many years remain until age 100.

## 2. Project Idea

Build a terminal program that prompts the user for a name and an age, then prints a personalised greeting that includes the number of years remaining until 100. The required baseline covers: read the inputs, greet the user, compute the remaining years, and react safely to an empty name, an age of exactly 100, and an age that cannot be parsed as a number.

## 3. Why This Project Now?

- First project in the path: it sets up the minimal shape of a runnable Go file — package declaration, entry point, importing the standard library, reading stdin, and printing to stdout.
- Introduces terminal input early, before more complex data types or control flow appear.
- Establishes the habit of validating input before using it; this habit is reused in every later project that touches user input.
- Provides a base of fluency needed for projects 002 onward, where multi-prompt interactive programs are the norm.

## 4. Prerequisites

- No prior project is required.
- Environment only: Go installed and on `PATH`, ability to run `go version` and interpret the output, familiarity with running a single command in a terminal and pressing Enter.

## 5. What You Must Know Before Starting

- The minimum runnable Go file: a package declaration and an entry point. Understand why these two pieces exist before writing any logic.
- Importing packages from the standard library: a single named import is usually enough for this project.
- Formatted output: the standard library offers several printing functions with different conventions for line endings, format verbs, and concatenation. Pick the one that fits each message; the choice is yours.
- Reading from standard input: the standard library offers several reading APIs, each with its own behaviour around whitespace, partial reads, and trailing newlines. The match between your prompt wording and the reading API you choose is the main design decision.
- Variables: declaration with an initial value, declaration with an explicit type, and the conventions for ignoring a returned value where applicable.
- Functions: why define a function rather than inline duplication, and how values flow in and out.
- The distinction between text that represents a number and the numeric value obtained by conversion.

## 6. Explanation of New Concepts

### Concepts

- The standard library's formatted I/O package: where it lives, what its main verb families cover (strings, integers, floats), and when each printing function is the natural choice. You do not need to memorise it; you do need to know what the documentation page contains.
- The first reading tool you adopt: each reading API in the package stops at a different delimiter (whitespace or newline, depending on the call). Choose deliberately and test the chosen API against multi-word inputs such as a name with a space.
- Error returns on input failure: many beginners ignore the secondary result. In Go, every read can fail, and the boolean or error result is what tells you so. Building the habit of checking protects you later when input comes from a file or the network.
- The arithmetic of years-remaining: simple subtraction, but the choice of numeric type (integer versus floating) and the policy for what happens when the current age has already passed 100 belong to you.
- The distinction between a message aimed at the user and a technical log or panic; thinking about what the user sees is part of basic user-experience care.

## 7. Learning Objective

By the end of the project you should be able to:
- Create a runnable Go source file from scratch without depending on a copied template.
- Read a line from the terminal and handle whatever is left in the input buffer.
- Convert a string read from the user into an integer safely and react to conversion failure.
- Compose a greeting that combines a string with one or more values, choosing the formatting that fits the message.
- Define small named functions that take parameters and return results, and explain each function's purpose.
- Distinguish three categories of output: prompt, success message, and error message.

## 8. Functional Requirements

1. F1: The program prompts for and reads the user's name as text.
2. F2: The program prompts for and reads the user's age as text and converts it to an integer safely.
3. F3: The program prints a personalised greeting that includes the years remaining until 100, using the values it read.
4. F4: When the age equals 100, the program prints a message that reflects reaching 100 instead of stating remaining years.
5. F5: When the age is greater than 100, the program does not print a negative "years remaining" in the output.
6. F6: An empty name (including one that becomes empty after normalisation) produces a sensible response rather than a crash or a leaked whitespace artifact.
7. F7: Age input that is not a meaningful age (non-integer, zero, negative, or any other value the program does not treat as valid) is handled with a clear message and a documented policy.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- A line representing the user's name. Any script is acceptable; surrounding whitespace is allowed.
- A line representing the age. On success, this parses as a non-negative integer.

#### Outputs

Text printed to standard output. Text-only examples:

- Normal name and age below 100:
  - Program prompts for the name.
  - User enters the name.
  - Program prompts for the age.
  - User enters the age.
  - Program prints a single greeting line naming the user and stating the remaining years.

- Empty name:
  - User submits an empty line for the name.
  - Program prints a greeting that acknowledges the missing name with your chosen wording, then still asks for the age.

- Age exactly 100:
  - User enters a name, then `100` for age.
  - Program prints a message that acknowledges reaching 100 instead of remaining years.

- Invalid age (non-integer text):
  - User enters a name, then text that is not parseable as an integer.
  - Program prints a clear message that the input is not a valid integer, and exits without producing a numeric result.

## 10. Rules and Edge Cases

- Whitespace-only name: must not crash; handle it consistently with your empty-name behaviour.
- Empty name: not a crash; the chosen behaviour must be defined in the program, not left to undefined behaviour.
- Age of zero or negative: invalid/nonsensical input; the program handles it with a documented policy and a clear message rather than producing a nonsensical output.
- Age greater than 100: must not print a negative remaining count.
- Non-integer age text: must be rejected with a clear message.
- The leftover-buffer trap: when reading calls with different stopping behaviours are mixed, leftover input from one prompt may pollute the next. Whichever approach you choose, test multi-word names followed by an age on the next line.
- Non-ASCII characters in the name: must appear unchanged in the output.

## 11. Project Constraints

- Libraries: the standard library only. The package for formatted I/O is sufficient.
- Prohibited: any external package.
- Persistence: nothing is written to or read from disk.
- Network: none. No ports, no requests.
- Tests: optional for this entry project; the verification section below lists scenarios to run manually.

## 12. Design Questions Before Coding

- What is the right type for the age variable? Should it be a signed type that permits negatives and rely on you to validate, or an unsigned type? Each choice has trade-offs; neither eliminates the need to validate the range meaningfully.
- How do you intend to read multi-word names? Have you tested your chosen reading API against names that contain spaces and against names written without spaces? Some reading calls stop at the first whitespace token; that is a property of the call, not a bug to work around.
- How do you treat a name that becomes empty after stripping surrounding whitespace? Is it the same as an empty input line? Be explicit.
- Do you read the name and the age through helper functions, or directly in the entry function? What are the trade-offs in testability and clarity?
- What message do you show for invalid age input? Where does it appear, and does it reference the offending input?
- After rejecting invalid age input, do you exit immediately or do you offer one reprompt? The plan does not require either; pick one, justify it in a comment, and apply it consistently.

## 13. Implementation Milestones

1. M1: Create the source file in the project folder with the minimum required for a Go program to compile and run; verify with a build and run cycle.
2. M2: Read the name from the terminal and print it back as confirmation, having chosen and tested a reading API against multi-word input.
3. M3: Add age input. Accept the user's input as text first, then perform a safe conversion to an integer, and react clearly when conversion fails.
4. M4: Compose the personalised greeting message and print it after both inputs are valid.
5. M5: Branch on the age value to handle age 100 and ages above 100 so that no negative "years remaining" is ever displayed.
6. M6: Verify the empty-name behaviour: the program responds sensibly, without crashing and without printing raw whitespace.
7. M7: Run every verification scenario from section 14 and confirm the program behaves as your design specifies.

## 14. Verification Cases the Learner Must Write

### Required Cases

- A normal name and a normal age below 100: the greeting shows the name and the years remaining.
- An empty name: the greeting is produced, no crash, no leaked whitespace.
- A whitespace-only name: handled consistently with the empty-name case.
- An age equal to 100: the message reflects reaching 100 instead of remaining years.
- An age above 100: the message does not contain a negative number.
- An age of zero: handled per the program's documented policy for invalid ages; the program produces a clear message and does not produce a nonsensical output.
- A negative age: handled per the same documented policy; the program produces a clear message and does not produce a nonsensical output.
- A non-integer age: rejected, no incorrect numeric output is produced.
- A multi-word name in any script (for example, Latin with a space, or any non-Latin script with a space): the name is captured intact.
- A non-ASCII name: printed without distortion.
- The leftover-buffer trap: a name with trailing whitespace followed by an integer on the next line produces the expected two readings, not one merged read.

## 15. Common Mistakes to Watch For

- Treating all reading APIs in the standard I/O package as interchangeable. Each has its own stopping behaviour around whitespace and newlines. Pick deliberately, after testing against your scenarios.
- Ignoring the error result returned from a reading call. This silently swallows failures and produces baffling output later.
- Computing "years remaining" without checking whether the current age has already passed 100.
- Forgetting that an empty line is not the same as whitespace-only text, and that whitespace-only is not the same as a single-token name. Pick a normalisation strategy and apply it consistently.
- Choosing an unsigned numeric type for the age and assuming the type alone enforces meaningful validation. The range must still be validated; the type only narrows the representable values.
- Letting the greeting template accidentally drop the name when the user typed only whitespace; test the empty path explicitly.
- Mixing two printing styles inconsistently within the same program; pick the style that fits each message and stay with it.

## 16. Topics and References for Study

- A Tour of Go: Packages; A Tour of Go: Functions; A Tour of Go: Type inference.
- Effective Go: Names; Effective Go: Control structures.
- The official documentation page for the formatted I/O package, both the printing section and the scanning section.
- The standard library specification on the error interface and how it surfaces from input APIs.
- Search terms: `Go stdin reading`, `Go formatted output comparison`, `Go string whitespace normalization`, `Go strconv integer conversion error`.

## 17. Self-Assessment Questions

1. What is the conceptual difference between the reading APIs that stop at whitespace and the ones that consume a whole line? Which one fits a name prompt, and why?
2. Why does ignoring the error result returned by a reading call lead to silent bugs? Describe one such bug in the context of this project.
3. What is the difference between an empty string and a whitespace-only string after you normalise it, and why might you want to fold them together?
4. Walk through what your program should do when the user enters `150` for the age. Why is displaying a negative remainder wrong?
5. If a future requirement changes the milestone from 100 to a different value, what parts of the program would you update, and which stay the same?
6. How does the trade-off between letting the program exit immediately on invalid age input and offering one reprompt change the user experience for this project?

## 18. Definition of Completion

- [ ] The program compiles and runs without compile errors.
- [ ] Every scenario in section 14 produces the behaviour documented in the code.
- [ ] No panic occurs in any documented scenario.
- [ ] The code structure is proportional to the problem; any separation into multiple functions has a clear reason.
- [ ] You can explain every line in your code without consulting a reference.

## 19. Optional Extensions

- Optional 1: Accept a full birth date and compute the current age from it. Validation rules are yours to define.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** None. This project introduces the shared Go, standard-library, testing, and module documentation used by later projects.

### New documentation introduced in this project

- **Language foundations:** [A Tour of Go](https://go.dev/tour/), [How to Write Go Code](https://go.dev/doc/code), [Go language specification](https://go.dev/ref/spec), [Effective Go](https://go.dev/doc/effective_go).
- **API references:** [`fmt`](https://pkg.go.dev/fmt), [`os`](https://pkg.go.dev/os), [`strconv`](https://pkg.go.dev/strconv), [standard-library index](https://pkg.go.dev/std).
- **Testing references:** [Go testing](https://pkg.go.dev/testing), [subtests and table-driven tests](https://go.dev/blog/subtests).
- **Tooling and dependency references:** [Go Modules Reference](https://go.dev/ref/mod), [dependency-management tutorial](https://go.dev/doc/modules/managing-dependencies).

### Project-specific learning focus

- **Learn now:** separating standard input, output, and error streams; validating input; and writing table-driven CLI tests.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
