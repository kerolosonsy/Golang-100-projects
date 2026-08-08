# Project 008 — Rock Paper Scissors

## 1. Project Name and Number

- Number: **008**, level 1 (language basics and CLI).
- Folder name in the table: **`008_rock_paper_scissors`**, matching `01-foundations/008_rock_paper_scissors/`.
- Kind: a small terminal program that plays a single round of rock-paper-scissors against the computer, decides the outcome, and prints the result clearly. The computer's choice is produced by a randomness source that the program can replace for testing.

## 2. Project Idea

Build a terminal program that asks the user to pick rock, paper, or scissors, asks a replaceable randomness source to pick one of the three on behalf of the computer, decides who wins or whether the round is a draw, and prints the outcome with both choices visible. The required baseline covers: one round, a typed choice with three named values, constants for the values, replaceable randomness, and clear handling of input that does not correspond to a valid choice.

## 3. Why This Project Now?

- First project in the path that introduces a custom type with named values, which sets the pattern for typed enums used from 016 onward.
- First project that explicitly requires a replaceable source of randomness, laying the groundwork for dependency injection in 011 and 015, and for testable randomness in 010.
- Reuses input handling and validation from earlier projects and adds the concern of mapping user-typed tokens to typed values.
- Introduces `iota` in its natural context: a sequence of related named constants. The concept is small but reused widely later (HTTP status codes, file modes, and many more).

## 4. Prerequisites

- Project **007** (`007_bmi_calculator`). Comfort with parsing, validation, and clear error messaging should already be in place.
- Familiarity with functions, custom types, and conditional branching.
- No new tools or libraries beyond the standard library used in earlier projects.

## 5. What You Must Know Before Starting

- A custom type in Go is a named alias for an existing type. The compiler does not infer it from a string; a value of the custom type must be produced explicitly.
- `iota` is a special constant generator. Inside a `const` block, each untyped constant line that does not assign an expression repeats the previous expression, and `iota` increments by one per line. The result is a small, readable block of named integers.
- The user types text. The text must be mapped to one of the typed values. The mapping is your design choice (full word, single letter, abbreviations), but the rule is part of the contract.
- The "computer" choice must come from a source the program can replace. A concrete source and an interface (or a function value) describing the contract make the program testable. The randomness source itself is not the concern of this project; choosing it is the concern of project 010, where the security stakes matter. For 008, a non-security source is acceptable.
- The game's outcome is a function of the two choices. There are exactly nine ordered pairs of choices, three of which are draws and six of which have a winner. Enumerating them, either in a table or in a small set of rules, is part of the design.
- The decision is consistent: the rule that decides "rock beats scissors" gives the same outcome, namely "user wins", regardless of which side is the user and which is the computer when the choices are in that order. When the order is reversed, the win/lose direction inverts accordingly (or the round remains a draw). A consistent, complete rule that covers all nine ordered pairs is easier to defend than six separate scattered rules.

## 6. Explanation of New Concepts

### Concepts

- Custom type as a domain concept: the choice of rock, paper, or scissors is not a string and not an integer at the level of the game's logic. A named type makes the program's intent explicit and lets the compiler catch mistakes that a string-based design would let through.
- `iota` for related constants: a tidy way to give names to a small sequence of values without spelling out each value. The values themselves are arbitrary integers; what matters is that each named constant has a distinct identity the compiler can compare.
- Token mapping: the act of taking user input and turning it into a value of the typed domain. The mapping is finite, complete, and explicit. A mapping that accepts three inputs is correct; a mapping that silently accepts a fourth is not.
- Replaceable randomness: the program asks for a choice through an interface or a function value rather than calling the source directly. In production the source is the real one; in tests, the source is one the test controls, so the outcome is deterministic.
- Consistent, complete outcome rule: a rule that, given any ordered pair of choices, returns exactly one verdict. When the order of the two choices is swapped, the user/computer verdict inverts win/lose (or stays as a draw). The rule's coverage and order behaviour are part of its correctness.
- Program boundaries: the part of the program that reads and maps input, the part that produces the computer's choice, the part that decides the outcome, and the part that prints the result are four distinct concerns.

## 7. Learning Objective

By the end of the project you should be able to:
- Define a custom type with three named values, declared through `iota` or through equivalent explicit constants.
- Read user input, normalise it (case, surrounding whitespace), and map it to one of the typed values, rejecting anything that does not map cleanly.
- Decide the computer's choice through a replaceable source, so a test can substitute a deterministic source.
- Apply a single, consistent, complete rule that returns win, lose, or draw for any pair of choices.
- Print both choices and the verdict in a format that makes the round easy to follow.
- Explain why a custom type is preferable to a raw string for a small fixed set of values.

## 8. Functional Requirements

1. F1: The program defines a custom type for the three choices: rock, paper, and scissors. The type has exactly three named values; no other value of the type exists at runtime.
2. F2: The program defines the three named values as constants. The constants are produced either through `iota` inside a `const` block or through three explicit declarations. Either form is acceptable; the chosen form is documented in your code.
3. F3: The program reads a line of text from the user and maps it to one of the three values. Mapping is case-insensitive and trims surrounding whitespace. The set of accepted inputs is documented (for example, full word, single letter, abbreviations).
4. F4: The program obtains the computer's choice from a randomness source that the program can replace. In production the source is the real one; in tests, the source is one the test controls.
5. F5: The program applies a single, consistent, complete rule that decides the outcome as win, lose, or draw for any pair of choices. All nine ordered pairs are covered, and swapping the user's and computer's choices inverts win/lose (or keeps it as a draw).
6. F6: The program prints both choices (the user's and the computer's) and the verdict in a single result line. The wording is your choice; the rule is that all three pieces of information appear.
7. F7: An input that does not map to any of the three values is rejected with a clear message. No silent default, no panic.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- A single line of text representing the user's choice. Acceptable inputs include the full name of the choice, a single letter, or whatever mapping you document. The input is read once; the program does not loop asking for re-entry by default.
- Surrounding whitespace is permitted; case is ignored.

#### Outputs

Text printed to standard output. Text-only examples:

- User picks rock, computer picks scissors:
  - The result line shows both choices and the verdict "user wins".

- User picks paper, computer picks rock:
  - The result line shows both choices and the verdict "user wins".

- User picks scissors, computer picks paper:
  - The result line shows both choices and the verdict "user wins".

- User picks rock, computer picks paper:
  - The result line shows both choices and the verdict "computer wins".

- User picks paper, computer picks scissors:
  - The result line shows both choices and the verdict "computer wins".

- User picks scissors, computer picks rock:
  - The result line shows both choices and the verdict "computer wins".

- Any choice against itself:
  - The result line shows both choices (the same) and the verdict "draw".

- An input that does not map to any of the three choices:
  - The program prints a clear message naming the invalid input and explaining the accepted forms; it does not print a verdict line.

## 10. Rules and Edge Cases

- The three valid choices are exactly the three named constants. There is no fourth, no "wildcard", no "quit" value inside the choice type.
- The input mapping is case-insensitive after trimming. Different cases of the same choice are equivalent.
- The mapping rule is exhaustive over the accepted set: every accepted input maps to exactly one value of the type.
- The randomness source is replaceable. The program obtains a value from the source, not from a global function. Tests substitute a source whose value is known.
- The outcome rule is consistent and complete: swapping the two inputs inverts the win/lose direction (or keeps it as a draw). Every ordered pair of choices has exactly one verdict.
- Invalid input does not produce a verdict. The program does not pick a default value and continue.
- The verdict for a draw is unambiguous; "draw" rather than "win" or "lose" for both sides.
- A second round, a replay loop, a quit command, or a farewell message is out of scope for this baseline. One round, one verdict.

## 11. Project Constraints

- Libraries: the standard library only. `math/rand` is acceptable as the production randomness source for this project; project 010 will revisit the choice for security-sensitive contexts.
- Prohibited: any external package. The game logic is small enough to live in the standard library's footprint.
- Prohibited: reading input more than once for the same round. The program does not enter a retry loop on invalid input by default.
- Persistence: none. The program reads, decides, prints, and exits.
- Network: none. No ports, no requests.
- Tests: optional in code; the verification section below lists scenarios the learner runs manually or as table-driven tests if tests are added.

## 12. Design Questions Before Coding

- How will you represent the three values? A `string`-based custom type is readable; an `int`-based custom type with `iota` is conventional. Either is fine; pick deliberately.
- What is the accepted mapping? Full word only, single letter only, or a mix? Each choice has a different ergonomics; the choice is yours, but the choice must be documented.
- Where does the randomness source live? As a global, as a field on a struct, or as a parameter to the function that decides the outcome? The parameter form is the most testable.
- How will you express the outcome rule? A lookup table indexed by the pair of choices, a `switch` over the two values, or a function that compares one value against another and returns "what beats what"? Each shape has different readability.
- How will the verdict be reported? A single string, a typed verdict, or an enum-like value? The shape affects how the print step uses it.
- How will you keep the input, the mapping, the randomness, the outcome, and the print step as separate concerns? Helper functions are the usual answer.

## 13. Implementation Milestones

1. M1: Create the source file with the minimum required for a Go program to compile and run; verify with a build and run cycle.
2. M2: Declare the custom type and the three named values. Confirm the type and the values exist as distinct entities the compiler can distinguish.
3. M3: Read the user's input and map it to one of the three values. Confirm that case differences and surrounding whitespace are normalised.
4. M4: Obtain the computer's choice from a replaceable randomness source. Confirm that the source can be substituted in a test.
5. M5: Apply a consistent, complete outcome rule that covers all nine ordered pairs and returns one of three results.
6. M6: Print the user's choice, the computer's choice, and the verdict in a single result line. The line must be readable and self-explanatory.
7. M7: Handle invalid input with a clear message and no verdict.
8. M8: Run every verification scenario from section 14 and confirm the program behaves as your design specifies.

## 14. Verification Cases the Learner Must Write

### Required Cases

- All three user-wins pairings: rock vs scissors, paper vs rock, scissors vs paper. Each verdict is "user wins".
- All three user-loses pairings: rock vs paper, paper vs scissors, scissors vs rock. Each verdict is "computer wins".
- All three draw pairings: rock vs rock, paper vs paper, scissors vs scissors. Each verdict is "draw".
- Input with mixed case such as `Rock` or `SCISSORS`: maps to the same value as the lower-case form.
- Input with surrounding whitespace such as `  paper  `: maps to the same value as the trimmed form.
- Input that does not match any of the accepted forms: rejected with a clear message; no verdict printed.
- The randomness source is replaced by a deterministic source that always returns the same value: the program's output is reproducible; the same test input produces the same verdict every run.
- The consistency of the outcome rule is exercised: for any user/computer pair, swapping the two choices inverts the user/computer verdict for non-draw pairings, and the verdict stays "draw" for the matched pairings.
- Invalid input is not coerced to a default value; the program does not produce a verdict.

## 15. Common Mistakes to Watch For

- Treating the choice as a `string`. The compiler cannot tell "rock" from "Paper" when both are strings; a custom type can.
- Spelling out the named constants with explicit values rather than using `iota`. Either is acceptable, but using `iota` makes the relationship between the values explicit.
- Letting invalid input default to one of the three values silently. The user said something that is not a choice; the program must say so.
- Calling the randomness source directly rather than through an interface or function value. The program then cannot be tested deterministically.
- Expressing the outcome rule as six separate cases (one per non-draw pairing). A single, consistent, complete rule that covers all nine ordered pairs is shorter and harder to get wrong.
- Forgetting to trim or normalise the input. The user's ` Rock ` and `rock` should be the same value.
- Mixing up win/lose directions: a working rule for "rock beats scissors" must, when the order is reversed, give "computer wins". The order-reversal check in section 14 catches this.
- Printing only the verdict without both choices. The user wants to see what the computer picked.

## 16. Topics and References for Study

- A Tour of Go: Type declarations, constants, and `iota`.
- Effective Go: Names, control structures, and the section on constants.
- The `math/rand` package documentation: how a `Source` and a top-level convenience function relate.
- Search terms: `Go custom type enum`, `Go iota constants`, `Go replaceable randomness source`, `Go dependency injection small example`.

## 17. Self-Assessment Questions

1. Why is a custom type preferable to a `string` for representing the three choices, even though the program reads text from the user?
2. The outcome rule returns one of three values. If you replaced it with a single boolean for "did the user win?", what information would you lose, and why is the three-valued form better for this project?
3. Your program calls the randomness source through a function value. A reviewer asks why you did not just call the package-level function. What is the most precise answer?
4. The mapping accepts `r`, `p`, and `s` as well as the full words. A reviewer argues this is too permissive. Defend the choice or revise it; what would change in the program?
5. Walk through what your program prints when the user types `dynamite`. Why is "dynamite" not a fourth choice?
6. The consistency of the outcome rule is important: swapping the two choices inverts win/lose. Describe one test case that would fail if the rule were not consistent in this sense, and what the failure would look like.
7. A second round would require a loop. Why is that not part of this baseline, and which later project introduces the loop pattern you would use?

## 18. Definition of Completion

- [ ] The program compiles and runs without compile errors.
- [ ] Every scenario in section 14 produces the behaviour documented in your code.
- [ ] No panic occurs in any documented scenario, including invalid input.
- [ ] The three values are declared through a custom type and three named constants; the values are distinct.
- [ ] The randomness source is replaceable; substituting a deterministic source changes the program's output accordingly.
- [ ] The outcome rule is consistent and complete: every ordered pair has exactly one verdict, and swapping the user's and computer's choices inverts win/lose (or keeps it as a draw). All nine ordered pairs are covered.
- [ ] The output line contains both choices and the verdict; no partial output.

## 19. Optional Extensions

- Optional 1: Accept one additional input that prints the full outcome table without playing a round. The table is informational, not a gameplay mode, and is generated from the same outcome rule.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisite:** [Project 007 — BMI Calculator](../../01-foundations/007_bmi_calculator/README.md#20-prerequisite-based-documentation-guide).

Read the linked guide first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **Standards and concept references:** [Go specification: iota](https://go.dev/ref/spec#Iota).

### Project-specific learning focus

- **Learn now:** custom types as enums, exhaustive outcome tables, deterministic randomness, and pure game logic.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
