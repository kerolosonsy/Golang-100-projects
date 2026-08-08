# Project 006 — String Reverser

## 1. Project Name and Number

- Number: **006**, level 1 (language basics and CLI).
- Folder name in the table: **`006_string_reverser`**, matching `01-foundations/006_string_reverser/`.
- Kind: a small terminal program that reverses a piece of text and decides whether that text is a palindrome, treating Unicode characters and emoji correctly.

## 2. Project Idea

Build a terminal program that reads a line of text, prints it reversed in a Unicode-aware way, and prints whether the same text reads the same forward and backward after a simple normalisation step. The required baseline covers: read a line, reverse it at the level of Unicode code points, optionally lower-case it and trim surrounding whitespace for the palindrome check, and report the result clearly. Empty input is a documented case, not a crash.

## 3. Why This Project Now?

- First project in the path that touches real text content, where the difference between bytes and Unicode characters becomes visible.
- Reinforces input handling introduced in 001 and 005, but adds a new concern: the program must not silently corrupt non-ASCII input.
- Sets the foundation for later projects that compare or canonicalise text (search indexing in 019, normalisation in 014, file content round-trips in 017 and 030).
- Introduces the `strings` package alongside the existing `fmt` and `strconv` habits, expanding the surface area of the standard library the learner can reach for.

## 4. Prerequisites

- Project **005** (`005_simple_quiz`). Reading input and iterating over slices should already be comfortable.
- Comfort with the minimum runnable Go program, with functions, with slices, and with conditional branching from earlier projects.
- No new tools or libraries are required beyond the standard library already used in projects 001 to 005.

## 5. What You Must Know Before Starting

- The Go `string` type is a read-only sequence of bytes; it is not a sequence of characters in any meaningful human sense. Indexing it returns a byte, not a letter.
- A `rune` in Go is an alias for an integer type wide enough to hold any Unicode code point. A slice of runes therefore represents a sequence of code points.
- The conversion between a string and a slice of runes, and the conversion back, has a precise meaning at the code-point level and loses no information as long as the string is valid UTF-8.
- A grapheme cluster is a higher-level concept: a visible "character" as perceived by a user, which may consist of several code points (a base letter plus combining marks, or a base emoji plus skin-tone or gender modifiers). The standard library does not split strings into grapheme clusters; doing so requires extra packages.
- ASCII upper-case and lower-case differ only in one bit. Other scripts have case rules that a naive byte-level `+ 32` does not respect; a `strings`-package function does the right thing for Unicode.
- A palindrome, in the version used here, is text that reads the same forward and backward after lower-casing and trimming surrounding whitespace. Internal spaces and the distinction between upper and lower case are not part of the definition for this baseline.

## 6. Explanation of New Concepts

### Concepts

- UTF-8 encoding: the way Go stores strings internally. Each code point is encoded as one to four bytes, and the encoding is self-synchronising, meaning the boundaries between code points can always be recovered from the byte stream. Reversing bytes directly would scramble any non-ASCII text; reversing code points is the safe level of operation for this project.
- Rune slices versus byte slices: the same underlying text looks very different in the two representations. Choosing the right level of representation is a design decision, not a detail.
- Unicode case folding: the operation that brings text into a canonical form for comparison. For ASCII it is the familiar lower-case; for scripts with more elaborate case rules the standard library does the right thing.
- Whitespace trimming: a separate operation from lower-casing. Combining them in the wrong order, or omitting one of them, is a frequent source of false negatives in palindrome tests.
- Empty input: a string with length zero is a valid string in Go; it has a reverse that is also empty, and it is trivially a palindrome by any reasonable definition. The program must acknowledge this case explicitly rather than panic.
- The `strings` package: a collection of small, well-named helpers. Most of them work at the byte level; a smaller subset operates at the rune level, and it is the rune-level subset that is relevant here.

## 7. Learning Objective

By the end of the project you should be able to:
- Explain in your own words why reversing bytes is not the same as reversing characters, and why the choice matters for any non-ASCII input.
- Decide, for a given task, whether byte-level, code-point-level, or grapheme-level handling is appropriate, and justify the decision.
- Convert between strings and rune slices correctly and use the result without off-by-one mistakes.
- Apply Unicode-aware case folding and whitespace trimming from the standard library.
- Design a palindrome test that treats empty input, whitespace, and case differences in a documented way.
- Tell apart the three meanings of "character" in Unicode discussions: bytes, code points, and grapheme clusters.

## 8. Functional Requirements

1. F1: The program reads one line of text from the terminal. The line may contain any printable characters, including Arabic script, Latin script with diacritics, and emoji.
2. F2: The program prints the input reversed at the Unicode code-point level. Every code point from the input appears in the output in reverse order; multi-byte sequences are never split.
3. F3: The program prints a verdict indicating whether the input is a palindrome, using the simple normalisation defined in section 10 (trim surrounding whitespace, lower-case using a Unicode-aware operation, then compare).
4. F4: An empty input (zero-length string after trimming) is accepted as a valid case. The program does not crash, the reversed output is also empty, and the palindrome verdict is consistent with your chosen definition for the empty case.
5. F5: Input that contains characters from different scripts (Latin, Arabic, emoji) is reversed and compared correctly. No code point is lost, duplicated, or reordered beyond what reversing produces.
6. F6: The reversal and the palindrome check are independent steps. Each can be reasoned about on its own, and the program reports each one separately.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- A single line of text, ending with a newline in the usual terminal fashion.
- The line may be empty, may contain only whitespace, may be pure ASCII, or may contain non-ASCII code points.
- The line is read as-is; the program does not require a specific encoding marker beyond the terminal's default.

#### Outputs

Text printed to standard output. Text-only examples:

- Normal Latin palindrome such as "level":
  - Program reads `level`.
  - Reversed line: `level`.
  - Palindrome verdict: a clear "yes" message.

- Mixed-script non-palindrome such as "café":
  - Reversed line: the code points of the input in reverse order. Whether the reversed output shows the accented `é` as a single code point or as two code points in reverse order depends on how the input was written (precomposed or decomposed). The reversal itself does not choose; it preserves whatever form the user typed.
  - Palindrome verdict: a clear "no" message.

- Arabic text or text with combining marks:
  - Reversed line: code points in reverse order. A combining mark, when present, is a separate code point that follows its base letter in the input; after code-point reversal, the combining mark precedes the letter it visually modifies. The visual attachment of a base letter and its combining mark is a grapheme-cluster property, not a code-point property, and the project's code-point baseline does not preserve that attachment.
  - Palindrome verdict: based on the code-point-level reading.

- Emoji input such as "ab😀cd":
  - Reversed line: code points in reverse order; a single-code-point emoji (such as the grinning face) stays as one code point in the output, in reverse position. An emoji that is itself a sequence of code points (for example, a base emoji plus a skin-tone modifier, or a family sequence joined by zero-width joiners) is composed of several code points and is reversed at that level, which can detach its parts visually.
  - Palindrome verdict: based on the code-point-level reading.

- Empty input:
  - Reversed line: empty.
  - Palindrome verdict: a clear "yes" message (an empty string is its own reverse).

## 10. Rules and Edge Cases

- Byte reversal is forbidden for the reversal output. Reversing at the byte level scrambles UTF-8 and produces invalid sequences.
- The palindrome comparison is performed on a normalised version of the input. Normalisation consists of trimming surrounding whitespace and applying Unicode-aware lower-casing, in the order you choose. Internal whitespace and the distinction between upper and lower case are not part of the definition.
- Empty string after trimming: treated as a palindrome by the project's definition. The verdict message is your wording, but it must be consistent across runs.
- Whitespace-only input: after trimming, the result is empty; the behaviour follows the empty case.
- Case difference inside the string (for example, "Level" vs "level"): ignored by the palindrome test because the normalisation lower-cases both sides.
- Internal spaces (for example, "race car" vs "racecar"): ignored by the simple normalisation used here. The comparison happens after lower-casing and trimming, not after stripping internal spaces. Decide whether "race car" is a palindrome under your rule; whatever you decide, document it.
- Combining marks and surrogate pairs in the input: the baseline reversal is at the code-point level. Grapheme-cluster reversal is out of scope for this baseline.
- A line longer than the terminal's usual display width wraps visually; that is a presentation concern, not a correctness concern.

## 11. Project Constraints

- Libraries: the standard library only. The `strings` package is the principal tool for case folding and trimming; `fmt` is used for output; `bufio` or another reading helper from the standard library is used for input.
- Prohibited: any external package that claims to split text into grapheme clusters. The project's baseline is code-point level, and the README makes this explicit.
- Prohibited: any operation that reverses a string at the byte level. The required baseline is code-point level.
- Persistence: none. The program reads one line, prints its outputs, and exits.
- Network: none. No ports, no requests.
- Tests: optional in code; the verification section below lists scenarios the learner runs manually or as table-driven tests if tests are added.

## 12. Design Questions Before Coding

- Will you keep the input as a string and convert it to a rune slice once, or will you work with it as a string everywhere and only convert for the reversal step? Each choice has a different feel; neither is wrong.
- How will you build the reversed output? The standard library offers several builders; choose deliberately and verify the choice against the emoji and Arabic examples.
- What is your definition of "the same text" for the palindrome check? Be explicit: trim surrounding whitespace only, or also collapse internal whitespace? Be explicit about case: lower-case both sides, or fold them using a stricter rule?
- How will you handle the empty case? Will you print a special verdict, or rely on the general verdict to cover it? Will the reversed output be the empty string, and is that the message you want to print?
- Where will you place the normalisation step: in a helper that takes a string and returns a normalised string, or inlined in the comparison? A helper is easier to test and easier to reuse; the inline form is shorter.
- How will you report two separate results (reversed text and palindrome verdict) on one screen? Two lines, one line with a separator, or another layout? The choice is yours; the requirement is that both pieces of information are present.

## 13. Implementation Milestones

1. M1: Create the source file with the minimum required for a Go program to compile and run; verify with a build and run cycle.
2. M2: Read a line of text from the terminal. Confirm that the line can hold any Unicode code point you type, including Arabic and emoji, without truncation.
3. M3: Convert the input to a sequence of code points and reverse it. Confirm the reversed output preserves every code point from the input in reverse order; no code point is lost or duplicated, and single-code-point emoji stay as one code point.
4. M4: Print the reversed output and the original input side by side or one above the other, with a clear separator.
5. M5: Add the palindrome check. Define a normalisation step (trim, then lower-case, in your chosen order) and apply it before comparing the input to its reverse.
6. M6: Print the palindrome verdict using wording that matches your documented definition, including the empty-case wording.
7. M7: Run every verification scenario from section 14 and confirm the program behaves as your design specifies.

## 14. Verification Cases the Learner Must Write

### Required Cases

- A pure-ASCII palindrome such as "level": reversed output equals the input; palindrome verdict is "yes".
- A pure-ASCII non-palindrome such as "go": reversed output is the reverse of the input; palindrome verdict is "no".
- An Arabic-script input such as the Arabic word for "book": the reversed output preserves the code points correctly; no byte-level scrambling.
- An input containing a single-code-point emoji such as "ab😀cd": the reversed output contains the emoji as one code point in reverse position; no code point is split or lost.
- An input with combining marks or diacritics: the reversal works at the code-point level, which means a combining mark, being a separate code point, can end up on the opposite side of the base letter it was modifying. The visible attachment of base and mark is a grapheme-cluster property and is outside the project's baseline.
- An empty input: the program accepts it, prints an empty reversed output, and prints a palindrome verdict consistent with the empty-case definition.
- A whitespace-only input: behaves like the empty case after trimming.
- Input with surrounding whitespace and mixed case such as "  Level  ": reversed output has no surrounding whitespace added back; palindrome verdict is "yes" under the simple normalisation.
- Input with internal spaces such as "race car": verdict matches your documented rule for internal spaces (typically "no" under the simple normalisation, "yes" only if you also collapse internal whitespace — pick one and document).
- A very long line (several hundred characters): no truncation, no panic, no slow path.
- Invalid UTF-8 input (a stray byte sequence): the program must not panic; behaviour for invalid sequences is documented in your code (replacement, error, or refusal to continue).

## 15. Common Mistakes to Watch For

- Reversing at the byte level because the operation looks shorter. The result is invalid UTF-8 and visually scrambled output for any non-ASCII text.
- Confusing code points with grapheme clusters. The baseline reverses at the code-point level; an emoji plus a skin-tone modifier, when present, will appear separated after reversal. Document this rather than claiming full grapheme-aware reversal.
- Forgetting to lower-case before comparing, leading to "Level" being marked as "not a palindrome" when the project definition treats it as one.
- Forgetting to trim surrounding whitespace, leading to a single space making an otherwise palindromic string read as "not a palindrome".
- Treating the empty string as an error case. It is a valid string with a well-defined reverse and a well-defined palindrome status; the program must acknowledge it.
- Building the reversed output with the wrong size of buffer. The standard library builders grow as needed; using a manually sized slice invites off-by-one errors at the boundary.
- Mixing two normalisation rules: trimming in one place and lower-casing in another, so the helper used by the palindrome check and the helper used for display disagree.
- Letting internal whitespace affect the verdict silently. Either define it as irrelevant (simple normalisation) or strip it and document the choice; do not let the verdict depend on the order in which you happened to apply operations.

## 16. Topics and References for Study

- A Tour of Go: Strings, range, and the distinction between byte iteration and rune iteration.
- Effective Go: Strings, and the section on conversions.
- The Go specification: "String types", "Rune literals", and the encoding subsection that explains UTF-8.
- The `strings` package documentation: case-folding functions, trimming functions, and the rune-aware helpers.
- The Unicode glossary: definitions of "code point", "grapheme cluster", and "combining character".
- Search terms: `Go rune vs byte`, `Go UTF-8 string reversal`, `Unicode grapheme cluster Go`, `strings.ToLower Unicode`, `Go palindrome Unicode`.

## 17. Self-Assessment Questions

1. A reviewer reads your code and sees a step that builds a slice of runes from a string and then takes the third element. Why are these two operations not the same as indexing the original string at the third position? Describe, in prose, what each step actually returns and why the two answers differ for any non-pure-ASCII input.
2. Your program reverses a string that contains an emoji and prints the result. Describe what the user sees at the byte level and at the code-point level, and explain why those two descriptions differ.
3. Why does lower-casing the string before comparing, rather than after, give the same result for this project? When would the order matter?
4. The plan defines the palindrome check as "trim then lower-case, then compare". What changes in your program if you decide instead to also collapse internal whitespace? Which functions change, and which stay the same?
5. If a future requirement asks you to reverse at the grapheme-cluster level instead of the code-point level, which part of the design has to change, and which parts stay the same? Be specific.
6. Walk through what your program prints when the user presses Enter without typing anything. Why is the empty case a deliberate decision rather than a special case to suppress?
7. A reviewer claims your program "handles Unicode correctly". What is the most precise version of that claim that you can defend for this project, and what is the claim you cannot defend yet?

## 18. Definition of Completion

- [ ] The program compiles and runs without compile errors.
- [ ] Every scenario in section 14 produces the behaviour documented in your code.
- [ ] No panic occurs in any documented scenario, including invalid UTF-8 input.
- [ ] The reversal step is at the code-point level; no byte-level reversal is present.
- [ ] The palindrome check uses a single, documented normalisation rule.
- [ ] The empty case has a deliberate verdict, not a leftover from the general path.
- [ ] You can explain the difference between bytes, code points, and grapheme clusters without consulting a reference.

## 19. Optional Extensions

- Optional 1: Extend the palindrome check to the grapheme-cluster level for inputs that contain emoji modifiers or combining marks. Use a well-established external Unicode package and document why the extra dependency is justified; the standard library does not provide grapheme segmentation.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisite:** [Project 005 — Simple Quiz](../../01-foundations/005_simple_quiz/README.md#20-prerequisite-based-documentation-guide).

Read the linked guide first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`unicode`](https://pkg.go.dev/unicode), [`unicode/utf8`](https://pkg.go.dev/unicode/utf8).
- **Standards and concept references:** [Go blog: strings, bytes, runes and characters](https://go.dev/blog/strings), [Unicode text segmentation](https://unicode.org/reports/tr29/).

### Project-specific learning focus

- **Learn now:** bytes versus runes versus grapheme clusters, Unicode normalization, palindrome rules, and invalid UTF-8.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
