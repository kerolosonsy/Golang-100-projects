# Project 010 — Password Generator

## 1. Project Name and Number
- Number: **010**, level 1 (language basics and CLI).
- Folder name in the table: **`010_pass_generator`**, matching `01-foundations/010_pass_generator/`.
- Kind: a small terminal program that generates a password of a documented length using a documented set of character categories, drawing randomness from a cryptographically secure source. The program rejects options that cannot be satisfied. The program does not use any pseudo-random mathematical source for the bytes that compose the password.

## 2. Project Idea
Build a terminal program that asks for a length and a set of character categories (lower-case letters, upper-case letters, digits, symbols, and similar documented groups), then produces a single password whose length and category composition match the request. The randomness comes from the standard library's cryptographically secure source. The program refuses requests that cannot be satisfied (for example, a length of zero, a negative length, or a request with no categories enabled). The output is one password per run.

## 3. Why This Project Now?
- First project in the path whose correctness depends on the choice of randomness source. The distinction between a secure source and a pseudo-random one is foundational for 030 (encryption), 050 (authentication), and every later project that handles secrets.
- Reuses the input-validation habit from earlier projects and adds the constraint that invalid options must be detected before any byte is produced.
- Introduces the pattern of options as flags with sensible defaults, building on the flag work from 009.
- Sets the expectation that the program never claims an entropy or security guarantee it cannot substantiate.

## 4. Prerequisites
- Project **009** (`009_system_info_cli`). Comfort with flags, defaults, and clear separation of concerns should already be in place.
- Familiarity with the custom-type and constant patterns from 008.
- No new tools or libraries beyond the standard library used in earlier projects.

## 5. What You Must Know Before Starting
- A cryptographically secure random source produces bytes that an attacker cannot predict, given reasonable assumptions about the attacker. The Go standard library provides such a source in `crypto/rand`. The package `math/rand` provides a pseudo-random source whose output is reproducible from a seed and is not appropriate for password generation.
- A character category is a set of symbols. The categories used here are documented: lower-case letters, upper-case letters, digits, and a documented set of punctuation symbols. Adding or removing categories is a design choice that must be reflected in the program and in this README.
- "Length" means the number of characters in the output, not the number of bytes or the number of random draws. In the typical case where every category is single-byte in UTF-8, length and bytes coincide; the project does not require Unicode-aware length accounting.
- An option is impossible when the user requests a length that the categories cannot fill, or asks for no categories, or asks for a length that is too small to contain at least one character from each requested category. Each kind of impossibility has its own message.
- "Entropy" is a property of the distribution, not of the output. A specific output string has no entropy in the strict information-theoretic sense; entropy belongs to the random process that produced it. The program does not claim a specific entropy value, and it does not equate length with entropy.
- A password generator is not a password manager. The program produces one password per run and forgets it. Storing, transmitting, or auditing passwords is out of scope for this project.

## 6. Explanation of New Concepts
- `crypto/rand`: the standard library package that reads bytes from a secure source maintained by the operating environment. The package's purpose is to provide cryptographic-quality randomness suitable for security-sensitive uses such as key material, nonces, and passwords. Its surface is small and documented; the relevant idea for this project is that the package exposes a way to obtain bytes that the program then translates into characters from the documented alphabet.
- `math/rand`: the standard library package that produces pseudo-random numbers from a seed. With a fixed seed, the sequence is reproducible, which is exactly what a password generator must not have. The package's presence in the standard library is not a recommendation for security-sensitive contexts.
- The distinction between "random enough for a game" and "random enough for a secret". Project 008 used the latter's predecessor casually; project 010 makes the distinction explicit.
- Options as flags: each character category is a boolean flag with a documented default. The combination of flags defines the alphabet; the alphabet defines what characters may appear in the output.
- Categories and guarantees: a guarantee that the output contains at least one character from each requested category is a stronger property than "characters are drawn from the union of the categories". The project's baseline includes the stronger guarantee: each requested category contributes at least one character.
- Failure modes: an invalid length (zero, negative, larger than some documented maximum), no categories enabled, or a request that cannot fit a character from each requested category into the requested length are all rejection paths. Each rejection path has its own message.
- Program boundaries: flag parsing, option validation, secure byte production, character selection, output. These are five distinct concerns.

## 7. Learning Objective
By the end of the project you should be able to:
- Explain, in plain language, the difference between `crypto/rand` and `math/rand` and why the difference matters for a password generator.
- Obtain all randomness from `crypto/rand`, handle source failures, and explain why the chosen character-selection design avoids bias.
- Enforce a documented length, a documented set of categories, and a guarantee that each requested category contributes at least one character.
- Detect every kind of impossible option and reject the request before any byte is produced.
- Refuse to claim an entropy value or a security guarantee the program cannot substantiate.
- Keep flag parsing, validation, randomness, selection, and output in separate steps.

## 8. Functional Requirements
- F1: The program reads a length and a set of character category flags. The length is a positive integer with a documented maximum. The categories include, at minimum, lower-case letters, upper-case letters, digits, and a documented set of symbols. The set of categories is exactly the set the program documents; no other categories exist.
- F2: The program uses `crypto/rand` as the only source of bytes for the password. No call to `math/rand` or to any other pseudo-random source is used to produce password bytes.
- F3: The output password has exactly the requested length, when the request is valid.
- F4: For each requested category, the output password contains at least one character from that category. The requirement is satisfied structurally; the program does not loop redrawing until a chance result happens to satisfy it.
- F5: Character selection is unbiased. The method the program uses to pick a character from the alphabet does not systematically favour some characters over others; the design explains why.
- F6: A request with no categories enabled is rejected before any byte is read from `crypto/rand`. The program does not fall back to a default alphabet.
- F7: A length of zero, a negative length, or a length larger than the documented maximum is rejected before any byte is read from `crypto/rand`.
- F8: A request whose length is too small to contain at least one character from each requested category is rejected before any byte is read from `crypto/rand`. The minimum length for a given set of categories is the number of categories, no less.
- F9: The program does not claim a specific entropy value, does not equate length with entropy, and does not advertise security properties beyond the documented behaviour.

## 9. Inputs and Outputs
**Inputs**:
- A length, as a positive integer within the documented range. Default may exist; the default is part of the contract.
- A set of category flags, each boolean. Each flag has a documented default.
- The categories are exactly the ones documented. There is no "custom alphabet" input by default.

**Outputs**: text printed to standard output. Text-only examples:

- Length 16, all four categories enabled:
  - Program prints a single line containing a sixteen-character password. The line contains at least one lower-case letter, at least one upper-case letter, at least one digit, and at least one symbol from the documented symbol set.

- Length 4, exactly one category enabled (for example, digits only):
  - Program prints a single line containing a four-character password composed entirely of digits.

- Length 1, exactly one category enabled:
  - Program prints a single line containing a one-character password from that category.

- Length 0, with any flag combination:
  - Program prints a rejection message; no password is printed.

- Length -5, with any flag combination:
  - Program prints a rejection message; no password is printed.

- Length 1000000, with any flag combination that would exceed the documented maximum:
  - Program prints a rejection message that mentions the documented maximum; no password is printed.

- All categories disabled:
  - Program prints a rejection message that explains that at least one category is required; no password is printed.

- Three categories enabled with length 2:
  - Program prints a rejection message that explains that the length is too small to include a character from each requested category; no password is printed.

- A run that successfully produces a password:
  - Two consecutive runs with the same flags produce different passwords, because the source is not seeded reproducibly for this purpose.

## 10. Rules and Edge Cases
- The byte source is `crypto/rand`. No path in the program produces password bytes through any other source.
- The category flags compose into the alphabet. The alphabet is the union of the categories that are enabled.
- The category-coverage guarantee is structural: every requested category appears at least once in the output. The implementation must not rely on chance to satisfy the requirement (for example, by redrawing until the result happens to cover every category); the requirement is satisfied by construction.
- The character selection is unbiased: each character in the alphabet must be equiprobable (or as close to equiprobable as the secure source allows). The program must not use a selection method that systematically favours some characters over others; the design must address selection bias as a property of the algorithm, not as a side effect.
- A rejection happens before any byte is read from `crypto/rand` for a rejected request. The program does not consume secure randomness it will not use.
- The documented maximum length exists to bound the allocation and the runtime of the generation step. A specific value is your choice; a few hundred is a reasonable upper bound for a baseline project.
- The program does not echo the options back to the user in production runs. Optional diagnostic output during development is your choice, but the documented production contract is one password line.
- The program does not store the password, does not write it to a file, and does not log it. The output channel is standard output.

## 11. Project Constraints
- Libraries: the standard library only. `crypto/rand` for randomness, `flag` for options, `fmt` for output, `errors` (or equivalent) for explicit error returns.
- Prohibited: `math/rand`, any third-party randomness package, and any hand-rolled pseudo-random generator.
- Prohibited: claiming a specific entropy value, a specific number of bits of randomness, or any other unsubstantiated security property.
- Prohibited: persistence. The password is printed to standard output and forgotten.
- Network: none.
- Tests: optional in code; the verification section below lists scenarios the learner runs manually or as table-driven tests if tests are added.

## 12. Design Questions Before Coding
- What is the documented symbol set? Printable ASCII punctuation, a subset of it, or a Unicode block? A small printable-ASCII set is the baseline; the choice is yours and must be in the README.
- How will you represent the alphabet? A constant string, a slice of runes, a function that returns the slice on demand? Each shape has different ergonomics; the choice is yours.
- How will you satisfy the category-coverage requirement structurally, without relying on chance? The implementation must guarantee that every requested category appears at least once; random redraws are not acceptable. The exact mechanism is a design decision and must be argued, not just stated.
- How will you pick characters from the alphabet without introducing selection bias? A naive mapping from a single random draw onto a small alphabet typically biases some characters; your design must address this as a property of the algorithm. Research the topic before deciding; the project does not prescribe a method.
- How will you handle an error from `crypto/rand`? The error from the package is rare but real. Treat it as a failure with a clear message; the program does not silently substitute a different source.
- How will you bound the maximum length? A constant, a flag, or another mechanism? A constant is the simplest baseline; a flag is a reasonable extension.
- How will you keep flag parsing, validation, randomness, selection, and output as separate concerns without adding unnecessary structure?

## 13. Implementation Milestones
1. M1: Create the source file with the minimum required for a Go program to compile and run; verify with a build and run cycle.
2. M2: Declare the category flags and the length flag, each with documented types and defaults. Verify the flag package parses the command line on a smoke run.
3. M3: Define the alphabet as the union of the enabled categories. Verify that the alphabet matches the flag combination.
4. M4: Implement the validation step: length is positive and within the documented maximum; at least one category is enabled; length is at least the number of requested categories. Each failure has its own message. No bytes are read from `crypto/rand` until all validation passes.
5. M5: Integrate `crypto/rand` as the sole randomness source and handle any failure explicitly, without prescribing the selection algorithm in advance.
6. M6: Compose the password so that each requested category appears at least once, every character is selected from the alphabet without bias, and the total length is exactly the requested length.
7. M7: Print the composed password as a single line.
8. M8: Run every verification scenario from section 14 and confirm the program behaves as your design specifies.

## 14. Verification Cases the Learner Must Write
- Length 16, all four categories: the password is sixteen characters and each category contributes at least one character. Do not use “two outputs differ” as a correctness assertion because a secure random process may legitimately repeat an output; verify the randomness source by inspection instead.
- Length 1, exactly one category: the password is one character from that category.
- Length 4, exactly one category: the password is four characters from that category.
- Length equal to the number of categories, with each category enabled once: the password is one character from each category, no more, no less.
- Length 0, with any flag combination: rejected; no password is printed.
- Length negative, with any flag combination: rejected; no password is printed.
- Length above the documented maximum, with any flag combination: rejected; the message mentions the maximum; no password is printed.
- All categories disabled: rejected; the message explains that at least one category is required; no password is printed.
- Three categories enabled with length 2: rejected; the message explains that the length is too small; no password is printed.
- The source for password bytes is `crypto/rand`. The verification is by inspection of the imports and the code path: no `math/rand` is used for password bytes.
- Two runs with the same flags produce different passwords. The difference is observable in the output, not in any test framework.
- The program's output contains no claim of a specific entropy value and no claim of a specific security property.
- A rejection path does not call `crypto/rand`. Verification is by inspection of the code: the validation runs before the byte-reading step.
- The set of characters in the output is a subset of the documented alphabet. Verification is by inspection.

## 15. Common Mistakes to Watch For
- Using `math/rand` because it is familiar or because a seed makes the program testable. For passwords, the source must be `crypto/rand`. Tests that need determinism must use a different design, such as testing the structural steps (validation, category coverage) separately from the randomness step.
- Quoting a specific entropy value or a specific number of bits for the output password. The program produces one concrete string per run; that string is a sample from a distribution, and the entropy is a property of the distribution, not of a single sample. The program does not advertise a number.
- Treating the category-coverage requirement as a probabilistic hope. The requirement is structural: every requested category must appear in the output by construction, not by chance.
- Using a selection method that introduces selection bias. A biased method systematically favours some characters over others, which weakens the output even when the underlying bytes are secure. The design must address bias as a property of the algorithm and argue why the chosen method is unbiased.
- Reading bytes before validation completes. The program must validate first, then read, so rejected requests do not consume secure randomness.
- Producing a password whose length is not exactly the requested length. The composition step must keep the total length exact.
- Conflating the alphabet with the per-category slice. The alphabet is the union of the enabled categories; the per-category slice is one component. Mixing the two leads to passwords that include characters outside the documented alphabet or that fail the category-coverage guarantee.
- Treating an error from `crypto/rand` as impossible. The error is rare but real. Handle it explicitly.
- Storing or logging the password. The output channel is standard output; the password is forgotten after printing.

## 16. Topics and References for Study
- A Tour of Go: Packages, basic types, and the introduction to errors.
- Effective Go: Names, control structures, and the section on errors.
- The `crypto/rand` package documentation: how to read bytes and how to handle the error.
- The `math/rand` package documentation: what it does and, importantly, what it is not designed for.
- The `flag` package documentation: how to declare flags of different types.
- A general reference on cryptographic randomness, including the role of secure sources and the difference between cryptographic randomness and seedable pseudo-randomness.
- A general reference on selection bias when mapping small alphabets to large random draws: why naive mappings introduce bias and what unbiased mappings look like.
- Search terms: `Go crypto/rand password generation`, `math/rand vs crypto/rand`, `unbiased random selection small alphabet`, `Go password generation category coverage`.

## 17. Self-Assessment Questions
1. Explain, in plain language, why `math/rand` is not appropriate for password generation, even when seeded from the current time.
2. Your program claims no entropy value. A reviewer says "a 16-character password from a 95-character alphabet has about 105 bits of entropy". Why is that statement not part of the program's documentation, and what would have to be true for the program to make it?
3. The category-coverage step is structural: every requested category appears in the output by construction. A reviewer suggests instead "draw the whole password, then verify, and retry if any category is missing". Why is the retry design incompatible with this project, and what does a structural design avoid?
4. The validation step runs before any byte is read from `crypto/rand`. A reviewer suggests moving the read earlier "to fail faster". What does the reviewer overlook, and why does the validation-first order matter?
5. A length of 3 with four categories enabled is impossible. The program rejects the request. Where exactly in the program does the rejection happen, and what does the user see?
6. Two runs of the program produce different passwords. What property of the source guarantees that, and what would change if the source were seeded?
7. The program uses `crypto/rand` but the rest of the standard library is also available. Why is the choice of source a security boundary, and why is "any randomness is fine" wrong?

## 18. Definition of Completion
- The program compiles and runs without compile errors.
- Every scenario in section 14 produces the behaviour documented in your code.
- No panic occurs in any documented scenario, including every impossible-option case.
- The password bytes are produced exclusively by `crypto/rand`. No other source is used.
- The validation step runs before any byte is read from `crypto/rand`.
- The category-coverage guarantee is structural and verifiable by inspection, not by chance.
- The character selection is unbiased. Your design argues why.
- The program's documentation and output make no claim of a specific entropy value or a specific security property beyond the documented behaviour.
- The set of categories is exactly the set documented in this README; no undocumented category exists.

## 19. Optional Extensions
- Optional 1: Add an option to exclude a documented set of ambiguous characters (for example, characters that look alike) from the alphabet. The option is opt-in; the default alphabet is unchanged.
- Optional 2: Add a flag that, when present, prints the alphabet and the length and asks the user to confirm before generating the password. The interactive confirmation is the only extension; the structural generation is unchanged.
