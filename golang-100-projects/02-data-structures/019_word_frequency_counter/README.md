# Project 019 — Word Frequency Counter

## 1. Project Name and Number

- Project **019** — `019_word_frequency_counter`.
- The directory name and number must match exactly.
- This project reads a stream of text, counts the words, and emits a deterministic frequency report.

## 2. Project Idea

The program reads text from a source through an `io.Reader` and counts how often each word appears. It defines a single token rule: a word is a maximal run of Unicode letters or Unicode digits. Punctuation, symbols, and whitespace are separators. Each token is converted to Unicode lowercase before counting.

The output is a deterministic report: words are listed in descending order of frequency, with ties broken by lexicographically ascending word. Two runs against the same input produce byte-identical output. Empty text and punctuation-only text both produce an empty report: zero output rows and zero output bytes. The program streams its input and uses a buffered reader; it does not load the whole text into memory before counting.

A subtle but important caveat is pinned by the README: the standard library's Unicode lowercasing is used, but Unicode normalization is out of scope. Two canonically equivalent spellings (a precomposed character and its decomposed form) are not guaranteed to merge. The standard library's lowercasing is not the same as locale-specific full Unicode case folding.

The project's semantic token limit is `1 MiB` (1,048,576 bytes). A token at that boundary is accepted; a longer token produces a clear error and is never partially counted. The scanner's internal maximum must include enough headroom beyond the semantic limit to inspect a delimiter or EOF, rather than assuming both limits are identical.

## 3. Why This Project Now?

- Projects 016 through 018 built small domains with stable ordering and on-disk storage.
- Project 019 introduces a different kind of input: a long, unstructured stream of text whose structure must be inferred from Unicode rules.
- The discipline of pinning one token rule, pinning one lowercasing rule, and pinning one tie-breaking rule is the same discipline as in 016 (stable IDs) and 018 (lexicographic category order): deterministic output from a deterministic rule.

- The project also introduces the discipline of streaming with a buffer that respects a fixed token-size limit.
- A naive `bufio.Scanner` with the default buffer size stops scanning when a token exceeds the configured maximum and surfaces a `bufio.ErrTooLong` through `Scanner.Err`.
- The README pins the limit and the failure mode so the test can pin both.

- Finally, the project establishes that "what counts as a word" is a definition, not a fact.
- Two readers will disagree on whether a hyphenated phrase is one word or two; the README pins one rule and the test enforces it.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 019 therefore requires:

- Completion of **018** (CSV Data Parser), including the discipline of streaming records through an `io.Reader` and producing deterministic output.
- No prior knowledge of HTTP, databases, generics, or concurrency.
- Familiarity with `unicode` and `bufio.Scanner` is helpful but not required; this project introduces both.

## 5. What You Must Know Before Starting

- That `bufio.Scanner` reads from an underlying `io.Reader` one token at a time, where the token boundary is decided by a `Split` function. The default split is `ScanLines`, which is wrong for word counting.
- That `bufio.Scanner` has a configurable buffer size, set with `Scanner.Buffer`. The default maximum token size is `bufio.MaxScanTokenSize`, which is `64 * 1024` bytes. When a single token exceeds the configured maximum, `Scanner.Scan` returns `false` and `Scanner.Err` returns `bufio.ErrTooLong`. Scanning stops at that point; the oversized token is not partially returned.
- That `unicode.IsLetter` and `unicode.IsDigit` answer "is this rune a letter or digit?" under Unicode rules. The token rule uses both predicates.
- That `unicode.ToLower` lowers a rune to its lowercase form using Unicode lowercasing. It is not locale-specific full Unicode case folding. It is not Unicode normalization; two runes that look identical but differ in precomposed-vs-decomposed form do not compare equal under `ToLower` alone.
- That a `map[string]int` aggregates counts but does not guarantee iteration order. The report must sort before printing.
- That the buffer maximum is measured in bytes, not runes. A Unicode token can use multiple bytes per rune, so the maximum byte budget is not the same as the maximum rune count.

## 6. Explanation of New Concepts

### Concepts

#### A pinned token rule

- A word is a maximal run of Unicode letters or Unicode digits.
- Punctuation, symbols, and whitespace are separators.
- Hyphens split a phrase: `well-known` is two tokens (`well`, `known`), not one.
- Apostrophes split a contraction: `don't` is two tokens (`don`, `t`), not one.
- The rule is simple to state and simple to test: every rune in a token satisfies `unicode.IsLetter(r) || unicode.IsDigit(r)`, and a separator is every other rune.

- The token rule is the learner's responsibility to enforce.
- The recommended way is to configure the scanner with a `Split` function that emits one token per maximal run of letter-or-digit runes.
- The test in section 14 pins the rule across the project's edge cases.

#### Unicode lowercasing, not full case folding or normalization

- `unicode.ToLower` lowers the case of a rune using Unicode lowercasing rules.
- ASCII letters lower predictably; non-ASCII letters lower according to Unicode's tables.
- The project uses the standard library's lowercasing and does not perform Unicode normalization.
- Two strings that are visually identical but differ in precomposed vs decomposed form — for example, `é` written as a single codepoint versus as `e` followed by a combining acute accent — are not guaranteed to merge into the same count.
- The standard library's lowercasing is also not the same as locale-specific full Unicode case folding (for example, Turkish dotless-I rules), which is not applied here.
- The README documents this caveat; the test does not assert on it.

#### A bounded buffer for long tokens

- The default `bufio.Scanner` buffer is too small for this project's largest supported token.
- The project accepts a token up to and including `1 MiB`, while configuring the scanner before its first scan with a larger finite internal maximum, such as `2 MiB`.
- The extra capacity is framing headroom so the scanner can see the separator or EOF after an exactly `1 MiB` token; it does not raise the project's accepted token limit.

- A token above the semantic `1 MiB` limit is rejected with a clear error and is never added to the counts.
- If input exceeds the scanner's larger internal maximum before a token can be emitted, scanning stops and `Scanner.Err` reports the scanner error.
- Neither path returns or counts a partial token.

- The maximum is measured in bytes, not runes.
- A token composed of multi-byte UTF-8 runes can use several bytes per character, so the rune budget is smaller than the byte budget.
- The README documents the byte unit so the test pins it correctly.

#### Deterministic ordering

- The report orders tokens by descending count, then ascending token.
- Two runs against the same input produce identical output.
- The discipline is: never iterate a map to produce output; sort the keys first; the comparator handles the compound key.

#### Streaming

- The program reads the input through an `io.Reader`.
- It does not load the whole text into a buffer before counting.
- A test with a large generated input confirms correctness across multiple `Read` calls; the implementation's record-by-record processing establishes the streaming design.

#### Empty and punctuation-only inputs

- Empty input (zero bytes) produces an empty report: zero output rows and zero output bytes.
- Input that contains only separators — whitespace, punctuation, symbols, no letters or digits at all — also produces an empty report.
- The test pins both cases as "empty output".

## 7. Learning Objective

After completing this project the learner can:

- Configure a `bufio.Scanner` with a custom `Split` function that emits one token per maximal run of letter-or-digit runes, and set a finite internal maximum larger than the accepted `1 MiB` token limit before scanning begins.
- Convert tokens to lowercase using the standard library's `unicode.ToLower`.
- Aggregate counts in a `map[string]int` and emit a deterministic report sorted by descending count and ascending token.
- Reject every token above the `1 MiB` semantic limit, and surface a scanner error if input exceeds the larger internal buffer maximum, with no partial counting in either case.
- Stream a large input through an `io.Reader` without loading the whole text into memory, and confirm correctness across multiple `Read` calls.
- Document the difference between Unicode lowercasing and Unicode normalization, and between Unicode lowercasing and locale-specific full Unicode case folding.
- Write tests that pin the token rule, the lowercasing rule, the ordering rule, the empty-input policy, and the oversized-token policy.

## 8. Functional Requirements

1. The program reads text from a source through an `io.Reader`. Production wires a file path; tests wire an `io.Reader` they control.
2. The program counts one occurrence per token. The token rule is "a maximal run of Unicode letters or Unicode digits"; separators are punctuation, symbols, and whitespace.
3. Tokens are converted to Unicode lowercase before counting, using the standard library's lowercasing. The program does not perform Unicode normalization and does not apply locale-specific full Unicode case folding.
4. The program emits a report listing each distinct token and its count. Tokens are ordered by descending count, with ties broken by ascending token.
5. The accepted token limit is exactly `1 MiB` (1,048,576 bytes), inclusive. Before scanning, configure a larger finite scanner maximum with delimiter/EOF headroom. Any token above the accepted limit is rejected clearly and never partially counted.
6. Empty input produces an empty report: zero output rows and zero output bytes. Input with only separators produces an empty report.
7. The sum of every count in the report equals the number of tokens read. The sum is an integer.
8. The program streams its input. A large input is processed correctly across multiple `Read` calls.
9. The report's output ordering is deterministic across runs on the same machine.
10. The output format is the learner's choice. The test pins the wording of the header (if any) and the row layout.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- A stream of text through an `io.Reader`. The text can be ASCII, non-ASCII, or a mix. The text can contain punctuation, symbols, whitespace, and any combination of letters and digits.

#### Outputs

- A deterministic report. Each non-empty line contains a token and its count. Tokens are sorted by descending count and ascending token for ties. The report is empty (zero output rows and zero output bytes) when there are no tokens to count.

#### Example text-only success run

```
$ wordfreq text.txt
Token          Count
the              6
of               4
and              3
a                3
quick            1
brown            1
fox              1
jumps            1
over             1
lazy             1
dog              1
```

The sum of counts (`6 + 4 + 3 + 3 + 1 + 1 + 1 + 1 + 1 + 1 + 1 = 23`) equals the number of tokens read.

#### Example text-only empty-output cases

```
$ wordfreq empty.txt
$ wordfreq punctuation-only.txt
```

Both invocations produce zero output rows and zero output bytes. There is no header line and no "(no tokens)" line in the empty report.

## 10. Rules and Edge Cases

- **Empty input.** Zero bytes of text. The report is empty: zero output rows and zero output bytes.
- **Punctuation-only input.** Input that contains only separators (no letters or digits) produces an empty report.
- **Whitespace-only input.** Treated the same as punctuation-only: empty report.
- **Single token.** A text containing exactly one token yields a report with one row whose count is `1`.
- **Repeated tokens.** A text with the same token repeated produces one row in the report with the correct count.
- **Mixed case.** Tokens that differ only in case ("The" and "the") are folded into a single count.
- **Hyphens and apostrophes.** Hyphens split. Apostrophes split. `well-known` is two tokens; `don't` is two tokens. The rule is pinned.
- **Unicode letters and digits.** A text containing non-ASCII letters or digits is tokenized using `unicode.IsLetter` and `unicode.IsDigit`. The test pins the rule with at least one non-ASCII example.
- **Mixed letters and digits.** `abc123` is one token. `abc123def` is one token. `abc 123` is two tokens.
- **Oversized token.** A token longer than the accepted `1 MiB` byte limit produces a clear error. If the larger scanner maximum itself is exceeded, that error comes from `Scanner.Err`. The token is never partially counted, returned, or silently truncated.
- **Streaming.** A large input is processed correctly across multiple `Read` calls. The verification case in section 14 confirms correctness.
- **Sum invariant.** The sum of every count in the report equals the number of tokens read. The test pins this invariant.

## 11. Project Constraints

- Go standard library only. No third-party text-processing libraries.
- The scanner uses `bufio.Scanner` with a custom `Split` function. The default split (`ScanLines`) is not appropriate for word counting and must not be used.
- The scanner's finite internal maximum is configured with headroom beyond the accepted `1 MiB` token limit before any scan. The program still rejects tokens above the semantic limit and never truncates them.
- Tokens are counted with a `map[string]int`. The output order is determined by sorting the keys with a comparator that handles the compound key (count descending, token ascending).
- The output is deterministic across runs on the same machine.
- The program does not perform Unicode normalization. The program does not apply locale-specific full Unicode case folding. The README documents both caveats.
- The program does not implement stop-word removal, stemming, or lemmatization.
- The empty report is zero output rows and zero output bytes. The program does not emit a header line or a "(no tokens)" line when the input is empty.

## 12. Design Questions Before Coding

- Where does the scanner live? In `main`, in a small package, or in two packages (a tokenizer and a counter)? Which choice lets the test drive the tokenizer with controlled inputs?
- How is the custom `Split` function shaped? Where does it live in the code? Which choice lets the test pin the token rule directly?
- How is the lowercasing performed? Token by token with `strings.ToLower`, rune by rune with `unicode.ToLower`, or through a wrapper that caches the result? Which choice is clearest to read?
- How is the compound ordering implemented? Which choice keeps the comparator obvious to read?
- How is the oversized-token case reported? Through `Scanner.Err` returned by the program, through a wrapped error, or through a sentinel? Which choice lets the test assert on the cause and confirm no partial counting occurred?
- How is the empty report handled? As zero output bytes, as a constant sentinel header, or as a single line "no tokens found"? Which choice matches the rule pinned in section 8?
- How is a large input injected for the streaming test? Through a custom `io.Reader` that yields the input across multiple `Read` calls, or through a `bytes.Reader`? Which choice tests correctness across reader boundaries without depending on real time?

## 13. Implementation Milestones

1. Decide the package layout: a small package for the tokenizer and the counter, and a thin `main` package that opens the file and prints the report.
2. Configure the scanner with a custom `Split` function that emits one token per maximal run of letter-or-digit runes. Apply `Scanner.Buffer` with a maximum of `1 MiB` (1,048,576 bytes) before any `Scan` call.
3. Iterate the scanner. For each token, convert it to lowercase and increment the count. Stop iterating when `Scanner.Scan` returns `false`; check `Scanner.Err` for an oversized-token error or any other read error.
4. Aggregate the counts in a `map[string]int`. Sum the counts as tokens are processed.
5. Sort the resulting counts with a comparator that orders tokens by descending count and ascending token for ties.
6. Format the report. The learner chooses the exact wording; the test pins it. The empty report is zero output rows and zero output bytes.
7. Wire the program's command-line interface: a single positional argument naming the input file, the report on standard output, errors on standard error.
8. Add the streaming test that drives a large generated input through a custom `io.Reader` that yields the input across multiple `Read` calls. The test asserts that the report's counts are correct, that the sum of counts equals the number of tokens, and that two runs of the same test produce byte-identical output.
9. Add boundary tests that accept a token exactly `1 MiB` long and reject a longer token with a clear error and no partial count. Also exercise the scanner-error path with input beyond its larger internal maximum.

## 14. Verification Cases the Learner Must Write

### Required Cases

Each case is described in natural language. The tokenizer is driven through a `strings.Reader` or a `bytes.Reader` for small inputs and through a custom incremental `io.Reader` for the large-input case. The program's command-line integration is a thin wrapper that opens the file. No test depends on real time, sleeps, or one-record-per-second pacing.

#### Token rule

- The input `"hello world"` produces two tokens: `hello` and `world`. The sum of counts is `2`.
- The input `"hello, world!"` produces the same two tokens; the comma and the exclamation mark are separators. The sum of counts is `2`.
- The input `"well-known"` produces two tokens: `well` and `known`. The sum of counts is `2`.
- The input `"don't"` produces two tokens: `don` and `t`. The sum of counts is `2`.
- The input `"abc123"` produces one token: `abc123`. The sum of counts is `1`.
- The input `"abc 123"` produces two tokens: `abc` and `123`. The sum of counts is `2`.
- The input `"hello.world"` produces two tokens: `hello` and `world`. The sum of counts is `2`.
- A non-ASCII letter, for example the Latin small letter `é`, counts as a letter and is part of a token.
- A non-ASCII digit, for example the Arabic-Indic digit `٤`, counts as a digit and is part of a token.

#### Lowercasing

- The input `"The the THE"` produces one token (`the`) with count `3`.
- The input `"Ångström"` and the input `"ångström"` both produce the token `ångström` after lowercasing; mixing them produces one token with the combined count.

#### Ordering

- A test runs the tokenizer against an input whose tokens have varying counts. The report is sorted by descending count and ascending token for ties. Two runs produce byte-identical output.
- A test runs the tokenizer against an input with a tie: three tokens each appearing once. The report lists them in ascending lexicographic order.

#### Edge cases

- Empty input (zero bytes) produces an empty report: zero output rows and zero output bytes.
- Input containing only punctuation (for example, `"!!! ???"`) produces an empty report.
- Input containing only whitespace (for example, spaces, tabs, newlines) produces an empty report.
- Input containing one token yields a report with one row whose count is `1`.

#### Sum invariant

- For every test that drives the tokenizer through a known input, the test asserts that the sum of every count in the report equals the number of tokens read.

#### Streaming

- A test constructs a large input and drives the tokenizer through a custom `io.Reader` that yields the input across multiple `Read` calls (for example, one token per `Read` or several tokens per `Read`). The test asserts that the report's counts are correct, that the sum of counts equals the number of tokens, and that two runs of the same test produce byte-identical output. The test does not rely on `time.Sleep`, wall-clock pacing, or timing.

#### Oversized token

- A test drives an input whose single token is longer than `1 MiB` (1,048,576 bytes). The test asserts that the program reports an oversized-token error, that no partial counting is reflected in the report, and that the failure is surfaced to the caller.
- A test drives an input whose token is exactly at the `1 MiB` boundary. The test asserts that the token is counted correctly and is not reported as an error.

#### Determinism

- A test runs the tokenizer twice against the same input and confirms the two outputs are byte-identical.

#### Process

- An integration test runs the compiled binary against a temporary file with valid data and confirms the exit code is zero and the report is on standard output.
- An integration test runs the compiled binary against an empty file and confirms the report is empty (zero output bytes) and the exit code is zero.

## 15. Common Mistakes to Watch For

- **Using the default `ScanLines` split.** Lines, not words, are emitted. The split function must emit one token per maximal run of letter-or-digit runes.
- **Making the scanner maximum identical to the accepted token limit.** The scanner may need to inspect a delimiter or EOF after the token. Give its finite internal buffer headroom beyond `1 MiB`, while still rejecting semantic tokens larger than `1 MiB`.
- **Configuring the scanner buffer after the first `Scan` call.** The buffer configuration must be applied before scanning begins; otherwise the scanner may have already allocated its internal buffer.
- **Silently truncating an oversized token.** Counting a prefix of an oversized token produces wrong results without telling the caller. Both semantic-limit and scanner-limit failures must surface an error with no partial count.
- **Lowercasing byte-by-byte.** Lowercasing at the byte level misses non-ASCII case differences. The lowercasing must operate on runes (for example, through `unicode.ToLower` or `strings.ToLower`).
- **Confusing Unicode lowercasing with locale-specific full Unicode case folding or with Unicode normalization.** The standard library's lowercasing is not the same as Turkish dotless-I rules or as `unicode.CaseRange` full case folding; it is also not normalization. The README documents the limits.
- **Iterating the count map directly to print.** Go's map iteration order is randomized; the output will differ between runs. The project requires sorting.
- **Treating hyphenated phrases as one token.** The rule is "hyphens split". `well-known` is two tokens, not one.
- **Treating contractions as one token.** The rule is "apostrophes split". `don't` is two tokens, not one.
- **Emitting a header line on empty input.** The empty report is zero output rows and zero output bytes. A header line or a "(no tokens)" line violates the rule.
- **Writing a timing-based streaming test.** A test that sleeps between records, paces records at a fixed interval, or otherwise depends on real time is flaky. Use a custom incremental reader and assert correctness across multiple `Read` calls.
- **Assuming a `bytes.Reader` proves the implementation never buffers everything.** A `bytes.Reader` confirms correctness; the streaming design is established by the implementation's record-by-record processing and by the test across multiple `Read` calls.

## 16. Topics and References for Study

- A Tour of Go: "Maps", "Reading files".
- Effective Go: "Data", "Sorting".
- Package documentation: `bufio` (`Scanner`, `Scanner.Split`, `Scanner.Buffer`, `Scanner.Scan`, `Scanner.Err`, `ScanWords`, `MaxScanTokenSize`, `ErrTooLong`), `unicode` (`IsLetter`, `IsDigit`, `ToLower`), `strings` (`ToLower`, `Map`), `sort` (`Slice`, `SliceStable`).
- Tokenization patterns: search for "Go bufio custom split", "Unicode letter categories", "Unicode normalization Go".
- Unicode case folding vs normalization vs lowercasing: search for "Go unicode normalization", "NFC NFD normalization", "Go strings.ToLower Unicode", "Unicode case folding locale".

## 17. Self-Assessment Questions

1. Why is the token rule pinned to "maximal run of Unicode letters or digits" instead of "split on whitespace"?
2. Why does the project use `bufio.Scanner` with a custom split instead of `strings.Fields`?
3. Why must the scanner's internal maximum be larger than the accepted `1 MiB` token limit, and what happens on each oversized-input path?
4. Why is the report sorted by descending count and ascending token, and what would break if the order were different?
5. Why is Unicode normalization out of scope, and what is the difference between Unicode lowercasing, full Unicode case folding, and Unicode normalization?
6. Why does the sum of every count equal the number of tokens read, and what would break if a token were counted twice?
7. Why does the streaming test use a custom incremental reader instead of relying on wall-clock pacing?
8. Why is the empty report zero output bytes, and what would a "(no tokens)" header line violate?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- [ ] The README's 19 sections are present in order; this file is the reference.
- [ ] Every functional requirement in section 8 is satisfied.
- [ ] Every verification case in section 14 has a corresponding test.
- [ ] The tokenizer uses a custom `Split` function that emits one token per maximal run of letter-or-digit runes.
- [ ] The scanner is configured before scanning with a finite internal maximum larger than the accepted `1 MiB` token limit.
- [ ] An oversized token surfaces a clear error, and no partial counting is reflected in the report.
- [ ] The sum of every count in the report equals the number of tokens read.
- [ ] The empty report is zero output rows and zero output bytes.
- [ ] The report's output ordering is deterministic across runs on the same machine.
- [ ] The package documentation states the token rule, the lowercasing rule, the `1 MiB` byte maximum, the oversized-token policy, and the distinction between Unicode lowercasing, full Unicode case folding, and Unicode normalization.
- [ ] The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Top-N report.** Accept a flag that limits the report to the top-N tokens by count. Tokens outside the top-N are omitted. The ordering rule is unchanged. Do not add filtering by token length or by token regex.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisite:** [Project 018 — CSV Data Parser](../../02-data-structures/018_csv_data_parser/README.md#20-prerequisite-based-documentation-guide).

Read the linked guide first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- None. This project applies already introduced APIs, standards, and testing practices in a new combination.

### Project-specific learning focus

- **Learn now:** streaming tokenization, scanner size limits, Unicode case handling and normalization, frequency maps, and stable tie-breaking.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
