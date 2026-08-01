# Project 021 — Log File Analyzer

## 1. Project Name and Number

Project **021** — `021_log_file_analyzer`. The directory name and number must match exactly. This project builds a deterministic streaming analyzer for a defined text-log format. It counts valid records per level, reports valid and malformed totals, and lists malformed-line diagnostics in input order.

## 2. Project Idea

The program reads a text stream and analyzes one record per line. Each line is expected to contain an RFC3339 timestamp, then exactly one ASCII space, then exactly one of the four supported uppercase levels (`DEBUG`, `INFO`, `WARN`, `ERROR`), then exactly one ASCII space, then a non-empty free-form message that may itself contain spaces. The program never reads the entire stream into memory; it consumes the input one line at a time through a scanner over an `io.Reader`.

The analyzer produces three blocks of output, always in this order:

1. A level block that lists every supported level in a fixed order with the count of valid records for that level.
2. A totals line that shows the count of valid records and the count of malformed lines.
3. A diagnostics block that lists every malformed line in input order, with its 1-based line number and a short reason.

Malformed lines are not fatal. They are skipped, do not contribute to the level counts, and appear in the diagnostics block. A line that exceeds the accepted content length, an underlying reader error, and a normal end-of-input are three distinct outcomes; the analyzer must keep them apart and never mistake one for another.

## 3. Why This Project Now?

Projects 019 and 020 introduced streaming discipline: a bounded buffer for tokens and a deterministic plan that is validated before any mutation. Project 021 turns that streaming discipline toward a long, weakly structured input: a text log whose records must be parsed one at a time without slurping the file. The discipline of pinning one line format, pinning one level set, and pinning one stable report order is the same discipline as 019 (one token rule, deterministic output) and 020 (deterministic plan, validate before mutating), applied here to parsing rather than to counting or moving.

The project also forces a clean separation between three outcomes that are routinely conflated:

- A normal end-of-input — no error.
- An ordinary malformed line — recorded in the diagnostics block, scanning continues.
- A structural failure — an over-limit line or an underlying reader error; the analyzer stops and returns an error.

Treating those three as one another is one of the most common log-analyzer bugs, and the project's correctness depends on distinguishing them deliberately.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 021 therefore requires:

- Completion of **020** (File Organizer). Earlier projects (for example 019's bounded-scanner discipline and 017's safe-file persistence) are background concepts already encountered and may inform the learner's design, but they are not additional required completions for this project.
- No prior knowledge of HTTP, databases, generics, or concurrency.

## 5. What You Must Know Before Starting

- That a "line" in a buffered scanner is whatever the line-terminator split considers a token. The default `ScanLines` split strips a single trailing `\r` and a single trailing `\n`. The line content is the bytes between the previous terminator and the terminator; the line terminator is not part of the content and is excluded from the line-length budget.
- That `bufio.Scanner` has a configurable buffer size, set with `Scanner.Buffer`. The default maximum token size is `bufio.MaxScanTokenSize`, which is `64 * 1024` bytes. When a single line exceeds the configured maximum, `Scanner.Scan` returns `false` and `Scanner.Err` reports a buffer-overflow condition.
- That the configured scanner maximum is a hard ceiling above which the scanner stops with an error, not the project's accepted-line-length limit. The two values must be set independently: a smaller accepted-line-length limit the analyzer enforces explicitly at `64 KiB`, and a larger scanner maximum at exactly `96 KiB` (98,304 bytes) that gives the scanner enough headroom to find the terminator or EOF after a line at or below the accepted limit.
- That an `io.Reader` is the only way to feed input. The reader can be a file, an in-memory buffer, or a custom incremental reader that returns data across multiple `Read` calls and may inject failures. A `Read` is allowed to return data and a non-`io.EOF` error together; the data delivered before that error is still part of the input.
- That `time.Parse` with the `time.RFC3339` layout accepts a fixed timezone offset (for example `Z` or `+02:00`) and rejects strings that do not match the layout exactly. The standard library's RFC3339 layout does not accept shortened forms and does not accept whitespace inside the timestamp.
- That a Go `map` has randomized iteration order. Any report derived from a map must sort its keys before emitting output, or the report will differ between runs.
- That line numbers in the input are 1-based: the first line the analyzer ever sees is line `1`, regardless of whether it is valid, malformed, blank, or absent because the input ended without a terminator. The diagnostics block uses 1-based line numbers for malformed lines.

## 6. Explanation of New Concepts

### The line format

Each record is exactly one line of text. A line is structured as three fields separated by exactly one ASCII space character (the byte `0x20`):

- **Timestamp.** An RFC3339 timestamp at the very start of the line, with no leading whitespace.
- **Level.** Exactly one of the four supported uppercase tokens: `DEBUG`, `INFO`, `WARN`, or `ERROR`.
- **Message.** Everything after the second space, up to (but not including) the line terminator. The message must contain at least one non-whitespace character. The message is free-form and may itself contain spaces, punctuation, and non-ASCII text.

The first space separates the timestamp from the level. The second space separates the level from the message. Because the timestamp itself never contains a space (RFC3339 places the offset at the end without spaces), the first two spaces in a well-formed line unambiguously mark the field boundaries. No other field structure, quoting, or escaping is supported.

A line is considered to have ended when the scanner encounters the line terminator (`\n`, or `\r\n` which is normalized to `\n` by `ScanLines`). A final line that has no terminator is treated as a regular line and is processed normally.

### The strict whitespace policy

The contract is strict: every separator is exactly one ASCII space, and every field is exactly what the contract describes.

- Leading whitespace before the timestamp is malformed.
- Trailing whitespace after the message (immediately before the terminator) is malformed. The analyzer does not trim and does not accept.
- A message whose only characters are whitespace is malformed (the same as empty).
- A blank line (empty or whitespace-only) is malformed.

The diagnostics block records the offending line with its reason. The contract makes the format unambiguous: any deviation from the strict shape is malformed, never silently coerced.

### Supported level set and fixed reporting order

The supported levels are exactly `DEBUG`, `INFO`, `WARN`, `ERROR`. The reporting order is pinned: the level block always lists `DEBUG`, `INFO`, `WARN`, `ERROR` in this order, with each level shown even when its count is zero. A count of zero is written as `0`, not omitted. The order is not the alphabetical order of the levels and is not derived from the input; it is a fixed contract.

### Valid records, malformed lines, and structural failures

The analyzer treats every line as exactly one of three things:

- **A valid record.** The line contains a timestamp that parses as RFC3339, a level that is one of the four supported tokens, and a message that contains at least one non-whitespace character. The record increments that level's count and the valid total.
- **A malformed line.** The line is well within the accepted line length and the scanner buffer, but its content does not match the format. Examples: blank lines, lines with leading whitespace, lines whose timestamp fails to parse, lines whose level is lowercase or unknown, lines whose message is empty or whitespace-only, lines whose trailing whitespace violates the strict policy. Malformed lines are skipped from the counts but are reported in the diagnostics block with a 1-based line number and a short reason. The analyzer continues scanning after a malformed line.
- **A structural failure.** Either the line's content exceeds the accepted line length (see "Content-length limit and scanner-buffer headroom" below) or the underlying reader returned an error. The analyzer stops and returns the failure as an error to the caller. Structural failure is not a malformed line and does not appear in the diagnostics block.

### Content-length limit and scanner-buffer headroom

Two values matter and they are different things:

- **Accepted line length.** The project accepts line content up to and including `64 KiB` (65,536 bytes), where "content" is the bytes of the line excluding the line terminator. The analyzer enforces this limit explicitly: when a line whose content exceeds `64 KiB` is produced by the scanner, the analyzer returns an explicit over-limit structural error naming the situation (for example, a typed or wrapped error that identifies it as "line exceeds 65536 bytes"). The over-limit line is not counted as valid, not counted as malformed, and not partially emitted into the output.
- **Scanner maximum.** The scanner's internal buffer is configured separately, with headroom beyond the accepted line length, so that the scanner can see the terminator or EOF after an exactly `64 KiB` line. The scanner's internal maximum is pinned to exactly `96 KiB` (98,304 bytes). Only when input pushes beyond that larger scanner maximum does the scanner itself stop and `Scanner.Err` report a buffer-overflow condition.

This separation has three observable consequences:

- A line whose content is `64 KiB` inclusive is accepted.
- A line whose content exceeds `64 KiB` but still fits within the configured scanner maximum is rejected by the analyzer's explicit over-limit check; it does not surface as `Scanner.Err`.
- Only input beyond the configured scanner maximum surfaces as `Scanner.Err`. The over-limit structural error and the `Scanner.Err` overflow error are both structural failures, but they are different errors with different messages.

### All-or-nothing output

The report is built in memory during the scan and is written to the injected `io.Writer` only after the analyzer has reached clean end-of-input with no structural failure. On any structural failure (over-limit line, reader error), the analyzer returns the error to the caller and writes nothing to the output writer. A partial report is never written. A caller that wants a guaranteed-correct report either receives the full report or receives the error and an empty writer; there is no in-between.

### Streaming without slurping the input

The analyzer reads through an `io.Reader`. The analyzer does not call any "read the whole file" helper, does not allocate a slice sized to the input, and does not depend on the input fitting in memory. A test that drives the analyzer with a custom `io.Reader` returning the input across many small `Read` calls confirms the streaming design.

### Reader data-and-error semantics

An `io.Reader` may return some data together with a non-`io.EOF` error on the same `Read` call. The analyzer's scan then processes the lines that the scanner has already emitted from that delivered data, treating them per the normal rules. When the scanner subsequently stops because of the underlying reader error, the analyzer surfaces that non-`EOF` error as a structural failure. The level counts, totals, and diagnostics reflect only the lines that were actually delivered and emitted. The analyzer never writes a partial report; the structural failure produces no writer output.

When the scanner emits a line whose content exceeds the accepted `64 KiB` limit but still fits within the `96 KiB` scanner maximum, the analyzer knows its 1-based line number and includes that number in the explicit over-limit error. If the scanner itself stops because a token exceeds its larger maximum, that token was never emitted, so the analyzer reports the scanner failure with context such as the number of complete lines already processed without inventing an exact line number for the unseen token.

### Stable report ordering

The level block is in the pinned fixed order (`DEBUG`, `INFO`, `WARN`, `ERROR`). The totals line appears once, immediately after the level block. The diagnostics block lists malformed lines in input order; that is, ascending 1-based line number. Two runs against the same input produce byte-identical output.

## 7. Learning Objective

After completing this project the learner can:

- Define a small line-oriented format with three fields separated by exactly one ASCII space, and document the format precisely enough that a parser is unambiguous, including a strict whitespace policy.
- Configure a `bufio.Scanner` with a custom buffer whose maximum is independent of — and larger than — the accepted `64 KiB` line-length limit, applied before any `Scan` call.
- Enforce an explicit over-limit content check that produces its own error, separate from `Scanner.Err`, and that fires before the input reaches the scanner's larger buffer ceiling.
- Distinguish "ordinary malformed input", "over-limit line", and "underlying reader error" as three different outcomes with three different handling rules.
- Build the report in memory during the scan and write it only after clean EOF, so partial reports never appear.
- Process `io.Reader` data-and-error returns according to standard reader/scanner semantics, surfacing the non-`EOF` error without losing already-delivered data.
- Count valid records per supported level, count total valid records, and count malformed lines, with stable reporting order and with zero counts shown as zero.
- Report malformed lines with 1-based line numbers and a short reason, in input order, without aborting the scan.
- Stream a long input through an `io.Reader` without loading the whole input into memory.
- Write tests that pin every outcome in the contract: well-formed input, malformed input, blank lines, leading-whitespace lines, trailing-whitespace lines, empty-message lines, missing fields, exact-boundary supported lines, over-limit lines that still fit the scanner buffer, lines that exceed the scanner buffer, and injected reader failures.

## 8. Functional Requirements

1. The program reads input from an `io.Reader`. Production wires a file; tests wire a `strings.Reader`, a `bytes.Reader`, or a custom incremental reader.
2. A record line has the form `<RFC3339 timestamp> <LEVEL> <message>`, where `<LEVEL>` is one of `DEBUG`, `INFO`, `WARN`, `ERROR`. The two separators are exactly one ASCII space each. The message contains at least one non-whitespace character.
3. Leading whitespace before the timestamp is malformed. Trailing whitespace between the message and the line terminator is malformed. A blank line is malformed.
4. The accepted line content length is up to and including `64 KiB` (65,536 bytes), where the content excludes the line terminator. A line whose content exceeds `64 KiB` is a structural failure: the analyzer returns an explicit over-limit error and writes no report.
5. The scanner's internal maximum is configured before any `Scan` call to exactly `96 KiB` (98,304 bytes), giving headroom beyond the accepted `64 KiB` line length. Only input that exceeds that larger scanner maximum is reported by `Scanner.Err`; it is also a structural failure.
6. An underlying reader error is a structural failure. The analyzer stops scanning and returns an error that wraps or identifies the reader error. The failure is not reported in the diagnostics block.
7. A line that fits both limits but does not match the format is malformed. The analyzer increments the malformed total, records a short reason and the line's 1-based line number, and continues scanning.
8. A valid record increments exactly one level count and increments the valid total. A malformed line increments neither.
9. The report is built in memory during the scan and is written to the injected `io.Writer` only after clean end-of-input with no structural failure. On any structural failure, the analyzer returns the error and writes nothing.
10. The output, when written, contains three blocks in this fixed order:
    - The level block lists `DEBUG`, `INFO`, `WARN`, `ERROR` in this fixed order, with each level's count shown (zero is shown as `0`).
    - The totals line lists `valid=` followed by the count of valid records and `malformed=` followed by the count of malformed lines.
    - The diagnostics block lists malformed lines in input order. Each diagnostic carries its 1-based line number and a short reason. If there are no malformed lines, the diagnostics block is empty.
11. The output is deterministic across runs on the same machine.
12. Empty input (zero bytes) produces: the full level block with every level at `0`, the totals line with `valid=0 malformed=0`, and an empty diagnostics block.
13. Normal end-of-input is not an error. The analyzer returns no error when the stream finishes cleanly.

## 9. Inputs and Outputs

### Inputs

- A stream of bytes through an `io.Reader`. The bytes form UTF-8 text whose lines conform to the line format above, or a mix of valid and malformed lines, or are empty. The stream may end with or without a final newline.

### Outputs

- The report written to an injected `io.Writer`. The report contains the level block, the totals line, and the diagnostics block, in that order. The exact wording of the report (header wording, separator characters, diagnostic format) is the learner's choice, but the test pins the learner's chosen format.

### Example text-only success run

```
$ analyze app.log
LEVEL    COUNT
DEBUG    2
INFO     5
WARN     1
ERROR    1
valid=9 malformed=0
```

### Example text-only mixed-input run

Input:
```
2026-07-29T10:00:00Z INFO request received
2026-07-29T10:00:01Z DEBUG cache hit
not-a-timestamp ERROR boom
2026-07-29T10:00:02Z warn lowercase level
2026-07-29T10:00:03Z ERROR
2026-07-29T10:00:04Z INFO request finished
```

Report:
```
LEVEL    COUNT
DEBUG    1
INFO     2
WARN     0
ERROR    1
valid=4 malformed=3
line 3: malformed timestamp
line 4: unsupported level (warn)
line 5: empty message
```

### Example text-only over-limit run

When the analyzer is given a line whose content exceeds `64 KiB` (and that still fits within the configured scanner maximum), the analyzer returns an over-limit structural error and writes nothing to the output writer. The error identifies the condition (for example, "line content exceeds 65536 bytes") and may include context such as "after processing N complete lines" when the analyzer can determine it.

### Example text-only scanner-overflow run

When input exceeds the configured scanner maximum, the analyzer returns a structural error that identifies the scanner-buffer overflow (for example, wrapped `bufio.ErrTooLong`). The analyzer writes no report. The line whose length pushed the input past the scanner maximum was not emitted as a parsed line, so its ordinal position is not promised.

## 10. Rules and Edge Cases

- **Empty input.** Zero bytes. The level block lists all four levels at `0`. The totals line shows `valid=0 malformed=0`. The diagnostics block is empty. No error is returned.
- **Blank line.** A line that is empty or contains only whitespace is malformed. The diagnostic reason is "blank line" (or the learner's pinned wording). The line number is recorded.
- **Leading whitespace.** A line that begins with any whitespace character before the timestamp is malformed. The diagnostic reason is "leading whitespace".
- **Trailing whitespace.** A line whose message is followed by whitespace immediately before the line terminator is malformed. The diagnostic reason is "trailing whitespace". The message is not trimmed; the analyzer does not silently accept and trim.
- **Bad timestamp.** A line whose first field does not parse as RFC3339 is malformed. The diagnostic reason is "malformed timestamp".
- **Unknown level.** A line whose level is anything other than `DEBUG`, `INFO`, `WARN`, `ERROR` is malformed. The diagnostic reason is "unsupported level" and includes the offending token (for example "unsupported level (warn)").
- **Lowercase level.** A line whose level is one of the supported levels but written in lowercase (for example `info`) is malformed, not silently matched. The same diagnostic format is used as for unknown levels.
- **Missing message.** A line whose message is empty (no characters after the second space) is malformed. The diagnostic reason is "empty message".
- **Whitespace-only message.** A line whose message contains only whitespace is malformed. The diagnostic reason is "empty message".
- **Missing field separator.** A line with fewer than two spaces is malformed. The diagnostic reason is "missing field separator" or similar; the exact wording is the learner's choice and is pinned by the test.
- **Final line without newline.** A final line that lacks a trailing newline is processed normally. The analyzer does not require a final terminator.
- **Carriage returns.** A line ending in `\r\n` is normalized by the scanner; the timestamp, level, and message fields are unaffected.
- **Long supported line.** A line whose content is exactly `64 KiB` is accepted. A line whose content is exactly one byte longer than `64 KiB` is rejected as a structural failure with an explicit over-limit error, even when the configured scanner maximum can still hold it.
- **Injected reader error.** A custom `io.Reader` that returns a non-`io.EOF` error mid-stream is surfaced as a structural failure. The analyzer does not silently treat the error as EOF, and it does not write a partial report.
- **Reader returns data and error together.** When a `Read` call returns bytes and a non-`io.EOF` error in the same call, the analyzer processes the lines the scanner emits from the delivered data per the normal rules. The subsequent stop attributed to the reader error is surfaced as a structural failure. The level counts, totals, and diagnostics reflect only the lines actually emitted before the stop.
- **Normal EOF.** The scanner returning `false` with a `nil` error is normal completion. No error is returned. The full report is written to the output writer.
- **Determinism.** The level block, the totals line, and the diagnostics block are emitted in the same order on every run against the same input.

## 11. Project Constraints

- Go standard library only. No third-party log-parsing libraries, no regular-expression helpers beyond the standard library, no log frameworks.
- The scanner is `bufio.Scanner` with the default line-terminator split. The buffer is configured with `Scanner.Buffer` before the first `Scan` call.
- The accepted line content length is `64 KiB` (65,536 bytes), exclusive of the line terminator. The scanner's internal maximum is configured separately to exactly `96 KiB` (98,304 bytes), giving explicit headroom beyond `64 KiB`.
- The strict whitespace policy is enforced: leading whitespace before the timestamp is malformed, trailing whitespace after the message is malformed, an empty or whitespace-only message is malformed, and a blank line is malformed. No trimming or silent coercion is applied.
- The supported level set is exactly `DEBUG`, `INFO`, `WARN`, `ERROR`. No other level is recognized; the level block's order is fixed.
- An over-limit line (content over `64 KiB`) returns an explicit over-limit structural error and writes no report. A scanner-buffer overflow returns the scanner's error and writes no report. A reader error returns a wrapped reader error and writes no report.
- The diagnostics block lists malformed lines in input order (ascending 1-based line number). It is never reordered by frequency or by reason.
- The analyzer does not load the entire input into memory before processing.
- The output is written to an injected `io.Writer` only after clean end-of-input. The analyzer does not write to standard output or standard error directly.
- The output is deterministic across runs on the same machine.

## 12. Design Questions Before Coding

- Where does the line-format contract live? As constants in the analyzer package, as constants in `main`, or as a small `format` package that owns the layout and the supported level set? Which choice lets the test pin the format in one place?
- How is the level block emitted in fixed order? Through a fixed slice of level tokens iterated in order, through a sorted iteration of a map, or through a switch that prints each level? Which choice keeps the order independent of the input?
- How is the accepted line length distinguished from the scanner's internal maximum? Through a named constant for each, through a single value used for both, or through a configuration object passed at construction? Which choice keeps the two-value contract testable?
- How is the explicit over-limit check performed? Before parsing each line, after parsing each line, or as a separate scan step? Which choice keeps the over-limit error independent of `Scanner.Err`?
- How is the boundary between "malformed" and "structural failure" enforced? Through the scanner's own error reporting, through an explicit length check before parsing the line, or through a wrapper that distinguishes the three outcomes? Which choice keeps the outcomes clearly distinct?
- How are diagnostics accumulated? Through a slice of structs, through a slice of formatted strings, or through a small builder that formats on demand? Which choice makes the input-order requirement obvious in the code?
- How is the report accumulated in memory before writing? Through a `bytes.Buffer`, through a `strings.Builder`, or through a slice of lines? Which choice keeps the all-or-nothing output contract obvious?
- How is the injected `io.Writer` plumbed? Through a constructor parameter, through an analyzer struct, or through a method that accepts the writer per call? Which choice keeps the streaming design testable without real files?
- How is the streaming test built? Through a custom `io.Reader` that yields the input across many small `Read` calls, or through a `bytes.Reader`? Which choice confirms that the analyzer never reads everything into memory before parsing?
- How is the injected reader failure built? Through a reader that returns an error after N bytes, through a reader that interleaves a planned error mid-stream, or through a wrapper that fails at a chosen offset? Which choice keeps the test deterministic and avoids permission tricks?

## 13. Implementation Milestones

1. Decide the package layout. Define a small analyzer package that takes an `io.Reader` and an `io.Writer` and produces a report (or returns an error and writes nothing). Keep `main` as a thin wrapper that opens the file and prints any structural-failure error.
2. Pin the contract as named constants: the supported level set, the fixed reporting order, the accepted line-content length (`64 KiB` exclusive of the terminator), the strict whitespace policy, and the scanner's internal maximum at exactly `96 KiB` (98,304 bytes).
3. Build a `bufio.Scanner` over the reader. Configure the buffer to exactly `96 KiB` (98,304 bytes) before the first `Scan` call, giving headroom beyond the accepted `64 KiB` line length. Use the default line-terminator split.
4. Iterate lines. For each emitted line, perform the explicit over-limit content check first: if the line's content exceeds `64 KiB`, return an over-limit structural error and write nothing. Otherwise, attempt to parse the three fields by finding the first two ASCII spaces and validating each field against the strict whitespace policy. Increment the matching counters or the malformed counter accordingly.
5. Maintain a running 1-based line number across iterations. When a line is malformed, record the line number and a short reason in the diagnostics slice.
6. Accumulate the report in memory during the scan: the level block in fixed order, the totals line, and the diagnostics block in input order. Do not write to the output writer yet.
7. On every scanner step, check whether the scanner stopped because of a structural failure (over-limit already handled by the explicit check; remaining cases are scanner-buffer overflow and reader error). If so, return the failure as an error to the caller and write nothing.
8. On normal completion (clean EOF), write the accumulated report to the injected `io.Writer`.
9. Wire `main`: open the file, run the analyzer, and exit non-zero on any structural failure.
10. Add tests covering each verification case in section 14, including the explicit over-limit failure (line fits scanner buffer but exceeds `64 KiB`), the scanner-buffer-overflow failure (line exceeds the larger scanner maximum), the injected reader failure, and the all-or-nothing output contract.

## 14. Verification Cases the Learner Must Write

Each case is described in natural language. Tests drive the analyzer through an `io.Reader` and assert on the `io.Writer` output. No test depends on real files, real home directories, or wall-clock pacing.

### Format and counting

- Input with one `DEBUG`, two `INFO`, one `WARN`, and one `ERROR` line: the level block shows `DEBUG=1`, `INFO=2`, `WARN=1`, `ERROR=1`, the totals line shows `valid=5 malformed=0`, and the diagnostics block is empty.
- Input with no errors at all (every line is well-formed): the totals line shows `valid=N malformed=0` and the diagnostics block is empty, where `N` is the number of lines.
- Input with mixed valid and malformed lines: the totals line shows the correct `valid=` and `malformed=` values, and the diagnostics block lists only the malformed lines with the correct 1-based line numbers in input order.

### Empty input

- Input of zero bytes produces the full level block with every level at `0`, the totals line `valid=0 malformed=0`, and an empty diagnostics block. No error is returned. The output writer receives the report.

### Final line without newline

- Input where the last well-formed line has no trailing newline is processed normally and counted as one valid record. The totals line shows the correct counts.

### Strict whitespace policy

- A blank line is malformed. The diagnostic carries the correct line number and a "blank line" reason (or the learner's pinned wording).
- A line with leading whitespace before the timestamp is malformed. The diagnostic reason mentions the leading whitespace.
- A line whose message is followed by whitespace before the line terminator is malformed. The diagnostic reason mentions the trailing whitespace. The analyzer does not trim and does not silently accept.
- A line whose message contains only whitespace is malformed. The diagnostic reason mentions the empty message.

### Other malformed lines

- A line with a non-RFC3339 timestamp (for example `2026/07/29 10:00:00`) is malformed. The diagnostic reason mentions the malformed timestamp.
- A line with an unsupported level (`TRACE`, `FATAL`, `info`, `warn`, empty level, etc.) is malformed. The diagnostic reason mentions the unsupported level and includes the offending token.
- A line whose message is empty (the level is followed immediately by the line terminator) is malformed. The diagnostic reason mentions the empty message.
- A line that lacks one or both field separators is malformed. The diagnostic reason mentions the missing field separator or the wrong field count.

### Long supported line

- A line whose content totals exactly `64 KiB` (65,536 bytes) is accepted, counted, and not surfaced as an error.
- A line whose content is exactly one byte longer than `64 KiB` is rejected as a structural failure with an explicit over-limit error. The error identifies the over-limit condition. The over-limit line does not appear in the diagnostics block. The output writer receives nothing.
- The over-limit error is reported by the analyzer, not by `Scanner.Err`, when the over-limit line still fits inside the configured scanner maximum.

### Scanner-buffer overflow

- An input whose line content exceeds the configured scanner maximum is rejected as a structural failure that wraps or identifies the scanner-buffer overflow. The output writer receives nothing. The error is distinguishable from the explicit over-limit error.

### Stable ordering

- A test runs the analyzer twice against the same input. The two outputs are byte-identical.
- A test runs the analyzer against an input whose malformed lines appear in non-monotonic order in the input; the diagnostics block lists them in ascending 1-based line number (which is monotonic with input order).
- A test runs the analyzer against an input with all four levels present at varying counts; the level block is in the fixed order `DEBUG`, `INFO`, `WARN`, `ERROR`.

### All-or-nothing output

- A test runs the analyzer against an input that ends with an over-limit line. The output writer is empty; the analyzer returns the over-limit error.
- A test runs the analyzer against an input that ends with an injected reader error. The output writer is empty; the analyzer returns the wrapped reader error.
- A test runs the analyzer against a well-formed input. The output writer receives the full report; the analyzer returns no error.

### Injected reader failure

- A test injects a custom `io.Reader` that returns a planned non-`io.EOF` error mid-stream. The analyzer returns an error that wraps or identifies the reader failure. The reader failure does not appear in the diagnostics block. The output writer is empty.
- A test injects a custom `io.Reader` that returns some bytes together with a non-`io.EOF` error on the same `Read` call. The lines the scanner emits from the delivered bytes are processed per the normal rules; the analyzer then returns the reader error. The output writer is empty.
- A test injects a custom `io.Reader` that returns `io.EOF` after delivering some well-formed lines. The analyzer produces a normal report on those lines and returns no error.

### Process

- An integration test runs the compiled binary against a temporary file containing a known mix of valid and malformed lines and confirms the exit code is zero, the report is on standard output, and any structural-failure case exits non-zero with an error on standard error and no report on standard output.
- An integration test runs the compiled binary against an empty file and confirms the level block is full of zeros, the totals line shows `valid=0 malformed=0`, and the exit code is zero.

## 15. Common Mistakes to Watch For

- **Conflating malformed lines with structural failure.** A malformed line is recorded in the diagnostics block. An over-limit line, a scanner-buffer overflow, and a reader error are returned as errors and produce no report. Treating them as the same thing silently loses information.
- **Setting the scanner's internal maximum equal to the accepted line length.** The scanner buffer must be exactly `96 KiB` (98,304 bytes), giving headroom beyond `64 KiB`. If the scanner buffer equals `64 KiB`, the scanner cannot see the terminator or EOF after an exactly `64 KiB` line and may report it as overflow.
- **Configuring the scanner buffer after the first `Scan` call.** The buffer configuration must be applied before scanning begins; otherwise the scanner may have already allocated its internal buffer with the default size.
- **Silently truncating a long line.** Returning a prefix of an over-limit line as if it were a valid record produces wrong counts without telling the caller. The explicit over-limit check returns an error with no partial count and no partial report.
- **Treating a reader error as normal EOF.** A custom `io.Reader` that returns a non-`io.EOF` error mid-stream is a structural failure; the analyzer must return that error. Mistaking it for EOF loses information and may produce a partial report.
- **Treating a reader error as a malformed record.** A reader error is not a malformed line. It must not appear in the diagnostics block.
- **Writing a partial report on structural failure.** The contract is all-or-nothing. A partial report on the writer followed by a returned error lets a caller mistake partial output for a complete report.
- **Lowercasing or fuzzy-matching levels.** The contract is exact-match. `info`, `Info`, and `INFO` are different; only `INFO` is accepted. Fuzzy matching silently changes the report.
- **Trimming trailing whitespace from the message.** The strict policy forbids trimming. A line whose message ends with whitespace is malformed; the analyzer does not silently accept and trim.
- **Reordering the level block.** The block must be in the fixed order `DEBUG`, `INFO`, `WARN`, `ERROR`. Reordering by frequency, alphabetically, or by first-seen violates the contract.
- **Omitting zero counts.** A level with zero valid records is reported as `0`, not omitted. Omitting zero counts changes the row count and breaks downstream parsing of the report.
- **Iterating a map directly to build the diagnostics block.** Map iteration order is randomized. The diagnostics block must be a slice appended in input order, and its order must not depend on map iteration.
- **Loading the whole input into a buffer before scanning.** A `bytes.Buffer` filled by `os.ReadFile` is not a streaming design. The streaming test confirms correctness across many small `Read` calls.
- **Using regular expressions to parse the line.** A regex is unnecessary; the format has exactly two field separators and three fields. A regex obscures the contract and is harder to test deterministically.
- **Promising an exact line number for an over-limit line that was never emitted.** The over-limit line was never returned as a parsed line, so its ordinal position is not knowable from the line counter. Name the condition; offer context such as "after processing N complete lines" when the analyzer can determine it; otherwise omit the number.
- **Treating the final line's missing newline as an error.** A line without a trailing terminator is normal. The analyzer does not require a final newline.
- **Writing the report to standard output from inside the analyzer.** The analyzer writes to an injected `io.Writer`. Hard-coding `os.Stdout` makes the analyzer un-testable.

## 16. Topics and References for Study

- A Tour of Go: "Methods", "Errors", "Reading files".
- Effective Go: "Errors", "Data".
- Package documentation: `bufio` (`Scanner`, `Scanner.Buffer`, `Scanner.Scan`, `Scanner.Err`, `ScanLines`, `MaxScanTokenSize`, `ErrTooLong`), `time` (`Parse`, `RFC3339`, `Time`), `strings` (`SplitN`, `IndexByte`), `io` (`Reader`, `EOF`), `sort` (`SliceStable`), `errors` (`Is`, `As`, `Wrap`).
- Streaming design: search for "Go bufio.Scanner buffer", "Go streaming line parser", "Go scanner headroom".
- Reader data-and-error semantics: search for "Go io.Reader data and error", "Go bufio partial read error".
- Error categorization: search for "Go error wrapping errors.Is", "Go sentinel errors vs typed errors".

## 17. Self-Assessment Questions

1. Why are three different outcomes (normal EOF, malformed line, structural failure) treated as three different things instead of one?
2. Why is the accepted line length (`64 KiB`) set independently of the scanner's larger internal maximum (`96 KiB`), and what would break if they were the same value?
3. Why is a line whose content exceeds `64 KiB` returned as a distinct over-limit error rather than truncated (or surfaced as `Scanner.Err`), and what does the test pin about that distinction?
4. Why is the report built in memory during the scan and written only after clean EOF (and not line-by-line), and what would a partial report on the writer followed by a returned error violate?
5. Why is the level block in the fixed order `DEBUG`, `INFO`, `WARN`, `ERROR` instead of alphabetical or by frequency?
6. Why is the supported level set exactly four tokens, and what would fuzz-matching lowercase levels change?
7. Why are blank lines, leading-whitespace lines, trailing-whitespace lines, and empty-message lines each reported with their own reason instead of one generic "malformed" reason?
8. Why does the diagnostics block list lines in input order instead of by frequency or by reason?
9. Why is a reader error returned as an error rather than recorded as a malformed line?
10. Why is the line terminator excluded from the content-length budget, and what would change if it were included?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every functional requirement in section 8 is satisfied.
- Every verification case in section 14 has a corresponding test, and every test drives the analyzer through an `io.Reader` and asserts on an `io.Writer`.
- The strict whitespace policy is enforced: leading whitespace before the timestamp is malformed, trailing whitespace after the message is malformed, an empty or whitespace-only message is malformed, and a blank line is malformed.
- The scanner buffer is configured to exactly `96 KiB` (98,304 bytes) before the first `Scan` call, giving headroom beyond the accepted `64 KiB` line-content length, and the accepted line-content length is set independently at `64 KiB` exclusive of the terminator.
- A line whose content exceeds `64 KiB` returns an explicit over-limit structural error and writes no report, even when the line still fits the scanner's larger buffer.
- A line whose content exceeds the scanner's larger maximum returns a wrapped `Scanner.Err` and writes no report.
- A reader error mid-stream returns a wrapped reader error and writes no report. The data delivered in the same `Read` call is processed per the normal rules before the stop.
- The level block is in the fixed order `DEBUG`, `INFO`, `WARN`, `ERROR`, with zero counts shown as `0`.
- The diagnostics block lists malformed lines in input order with 1-based line numbers and short reasons.
- The output is byte-identical across two runs against the same input.
- The package documentation states the supported level set, the level block's fixed order, the `64 KiB` accepted line-content length, the `96 KiB` scanner-buffer maximum, the strict whitespace policy, and the three-way outcome policy (normal EOF, malformed, structural failure).
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Per-level filter flag.** Accept a flag that lists the levels to include in the report. Levels not listed are still counted in the valid total but are not printed in the level block. The diagnostics block is unchanged. Do not add filtering by message content or by timestamp range.
- **ISO-date summary line.** Add one summary line at the bottom of the report listing the earliest and latest timestamps among the valid records, in RFC3339 form. The line is omitted when there are no valid records. Do not add timezone conversion or date-format changes.
