# Project 018 — CSV Data Parser

## 1. Project Name and Number

- Project **018** — `018_csv_data_parser`.
- The directory name and number must match exactly.
- This project reads a small CSV file and produces a deterministic statistical report.

## 2. Project Idea

The program reads a CSV file through `encoding/csv`, validates every row against a fixed schema, and emits a deterministic report. The report contains the row count, the grand sum, the grand average, and one row per category with that category's count, sum, and average. Category rows are sorted lexicographically so two runs against the same input produce identical output.

The schema is small and pinned by the README: the header is exactly `category,value`, and each data row has exactly two fields. The first field is the category, normalized by trimming leading and trailing Unicode whitespace; an empty trimmed category is rejected. The second field is a numeric value. Quoted commas inside a category are accepted; the program does not hand-split on commas. Empty input (zero records) is rejected because there is no header to validate; header-only input is a valid zero-row report.

Malformed rows — wrong header, wrong field count, empty trimmed category, non-numeric value, NaN, or infinity — fail the parse and the report is not produced.

## 3. Why This Project Now?

- Projects 016 and 017 introduced structured collections and on-disk JSON.
- Project 018 introduces a different input format: a streaming, tabular, comma-separated format produced by tools that are not Go.
- The CSV format is deceptively simple — commas inside quoted fields, escaped quotes, line endings — and the discipline of using `encoding/csv` instead of `strings.Split` is exactly the discipline this project introduces.

- The project also introduces deterministic reporting.
- Two runs against the same input must produce identical output, including the order of category rows.
- Without an explicit ordering rule, the output depends on Go's map iteration order and is not reproducible.
- The README pins lexicographic order on categories, which is the simplest rule that produces a stable report.

- Finally, the project establishes a clear malformed-input policy.
- The program either produces a complete report or reports an error; it does not produce a partial report and pretend the run succeeded.
- That policy is the foundation that the log analyzer in project 021 builds on.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 018 therefore requires:

- Completion of **017** (JSON Todo Persister), including the discipline of an `io.Reader` seam and of streaming instead of buffering whole documents.
- No prior knowledge of HTTP, databases, generics, or concurrency.
- Familiarity with `encoding/csv` is helpful but not required; this project introduces it.

## 5. What You Must Know Before Starting

- What `encoding/csv` does: it wraps an `io.Reader`, splits it into records on commas, respects double-quoted fields, and unescapes doubled quotes inside a quoted field.
- That `csv.Reader.Read` returns one record at a time as `[]string`. The slice's length equals the number of fields in that row; a wrong-length row is the first signal of a malformed file.
- That `csv.Reader.FieldsPerRecord` can be set to a positive integer to enforce a fixed field count, with `FieldsPerRecord = -1` meaning "do not enforce".
- That `strconv.ParseFloat` returns a `float64` and an error. The error distinguishes "not a number" from "parsed to a special value". The project adds explicit checks for NaN and infinity on top of that.
- That floating-point arithmetic is approximate. Sums of decimal values are generally approximate too, not only averages, because most decimal fractions do not have an exact binary representation. The test asserts on counts and order exactly, and asserts on sums and averages within a documented tolerance.
- That two runs of the same program against the same input on the same machine produce identical output for the same floating-point operations; the report is deterministic and the README does not claim last-bit nondeterminism across runs.
- That `strings.Compare` or `bytes.Compare` gives lexicographic ordering over strings, which is what the README pins for category rows.

## 6. Explanation of New Concepts

### Concepts

#### Why use `encoding/csv` instead of `strings.Split`

A CSV field can contain a comma if it is quoted. `"a, b", 1` is a two-field row, not a three-field row. Hand-splitting on commas will treat `"a, b"` as the start of a quoted region and produce wrong results; it also does not handle escaped quotes (`""` inside a quoted field). `encoding/csv` handles both correctly. The discipline of the project is: never hand-split CSV; always go through `encoding/csv`.

#### A pinned schema

The schema is small enough to remember and rich enough to support useful statistics:

- The header is exactly the two fields `category` and `value`, in that order. The header itself is matched exactly and is not trimmed; any deviation from the exact two-field header is rejected.
- Each data row has exactly two fields. A row with one or three or more fields is rejected.
- The first field is the category. The category is normalized by trimming leading and trailing Unicode whitespace. The trimmed value must be non-empty. Internal spaces are preserved. The normalized category is the key under which per-category aggregates are stored and is the value emitted in the report.
- The second field is the value. It must parse to a finite `float64`. NaN and infinities are rejected; non-numeric text is rejected.

#### Streaming over buffering

The CSV reader streams records. The program reads one record at a time and updates aggregate state (count, sum, per-category count, per-category sum). It never builds a slice of every row. The benefit is that the program handles a large generated input without holding it in memory; the verification case in section 14 uses a large generated input through an `io.Reader` to confirm correctness across many rows.

#### Deterministic category ordering

The per-category rows in the report are sorted lexicographically by the normalized category. Two runs against the same input produce the same output. The discipline is: never iterate a map directly when producing deterministic output; always sort the keys first.

#### Malformed-input policy

The program either produces a complete report and exits with code zero, or reports a single error identifying the offending CSV record and exits with a non-zero code. It does not produce a partial report. The error message includes enough information for the user to locate the problem: the record number (1-based, where the header is record 1) and, when a field can be identified, the field name (`category` or `value`). When the parser cannot identify a field — for example, on a quoting error that has no field context — the program preserves and wraps the underlying `encoding/csv` error.

#### Numeric syntax and tolerance

The value field accepts every finite syntax that `strconv.ParseFloat` accepts at the program's selected bit size, including exponent notation. NaN and infinities are rejected explicitly because `strconv.ParseFloat` returns them as valid `float64` values.

Sums of decimal values are generally approximate because most decimal fractions do not have an exact binary representation. Counts and order are compared exactly. For sums and averages, this project fixes a combined tolerance: accept a difference no larger than `1e-12` for values near zero, or `1e-9` times the magnitude of the expected value for larger values. Two runs of the same program against the same input on the same machine produce identical output; the README does not claim that floating-point sums and averages vary between runs.

## 7. Learning Objective

After completing this project the learner can:

- Use `encoding/csv` through an `io.Reader` seam to stream a CSV file one record at a time.
- Pin a small schema and reject any input that does not match it, with an error message that identifies the record and field when available, and preserves the underlying parser error otherwise.
- Aggregate streaming records into deterministic per-category statistics without holding every row in memory.
- Sort category keys lexicographically so the report is reproducible.
- Distinguish "empty input" from "header-only input" and apply a different policy to each.
- Reject NaN and infinity explicitly, on top of `strconv.ParseFloat`'s built-in errors.
- Document a single fixed floating-point tolerance that the test uses to compare sums and averages, while asserting on counts and order exactly.
- Write a test that drives a large generated input through a reader and confirms the report's correctness across many rows.

## 8. Functional Requirements

1. The program accepts an input source through a seam — for example, a file path or an `io.Reader`. Production wires a file path; tests wire an `io.Reader` they control.
2. The program reads the input through `encoding/csv`. It does not hand-split on commas.
3. The first record must be the header `category,value` exactly. Any other header is rejected with a clear error.
4. Every subsequent record must have exactly two fields. A record with one or three or more fields is rejected with a clear error that includes the record number and the field count.
5. The category field is normalized by trimming leading and trailing Unicode whitespace. The trimmed category must be non-empty. An empty trimmed category is rejected with a clear error that includes the record number.
6. The value field must parse to a finite `float64` at the program's selected bit size, accepting exponent notation. A non-numeric value, NaN, or infinity is rejected with a clear error that includes the record number.
7. The program computes the grand count, grand sum, grand average, and for each category its count, sum, and average. Aggregates are computed as the records are read.
8. The report lists category rows in lexicographic order of the normalized category. The grand statistics are emitted in a fixed position; the test pins the layout.
9. Empty input (zero records) is rejected. The program reports that the input is missing the expected header and exits with a non-zero code. The report is not produced.
10. Header-only input (exactly one record, the header) is a valid zero-row report. The grand count is `0`, the grand sum is `0`, the grand average is `0` (explicitly represented, not omitted), and there are no category rows.
11. A malformed input fails the entire parse. The program emits one error message identifying the record and field (when available) and exits with a non-zero code. It does not produce a partial report.
12. The program streams records; it does not buffer the whole file. The verification in section 14 confirms correctness across a large generated input.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- A CSV file with a header `category,value` and zero or more data rows. Categories can contain spaces, punctuation, or commas if quoted. Values are decimal numbers in any finite syntax accepted by `strconv.ParseFloat`, including exponent notation.

#### Outputs

- A deterministic report containing the grand statistics and one row per normalized category. The exact layout is the learner's choice; the test pins the wording of the header and the order of the rows.

#### Example text-only success run

```
$ csvreport input.csv
Rows: 5
Grand sum: 30
Grand average: 6
Category report:
  books, count=2, sum=15, average=7.5
  food, count=3, sum=15, average=5
```

#### Example text-only failure runs

```
$ csvreport empty.csv
Error: input is empty; expected header row "category,value".
$ csvreport bad_header.csv
Error: header row does not match expected schema at record 1: expected "category,value".
$ csvreport short_row.csv
Error: record 4 has 1 field, expected 2.
$ csvreport bad_value.csv
Error: record 7 has non-numeric value in field "value".
$ csvreport nan_row.csv
Error: record 3 has non-finite value in field "value".
```

When the parser cannot identify a field — for example, on an unterminated quoted region — the program preserves and wraps the underlying error and includes the record number when one can be determined.

## 10. Rules and Edge Cases

- **Empty input.** Zero records (the file has zero bytes). The program reports that the input is missing the expected header and exits with a non-zero code. The report is not produced.
- **Header-only input.** Exactly one record, the header. The report is emitted with grand count `0`, grand sum `0`, grand average `0`, and no category rows. The exit code is zero.
- **Wrong header.** Any header other than the exact `category,value` is rejected. The error message names the offending record.
- **Duplicate header.** A header whose field names repeat (for example `category,category`) is rejected.
- **Missing header.** A file whose first record does not match the pinned header is rejected. The error message says "header row does not match expected schema".
- **Wrong field count.** A data row with one or three or more fields is rejected. The error message names the record number and the actual field count.
- **Empty category.** A data row whose first field is empty after trimming is rejected. The error message names the record number.
- **Non-numeric value.** A data row whose second field does not parse as a finite `float64` is rejected. The error message names the record number.
- **NaN or infinity.** A value that parses to a special floating-point value (for example, the strings `NaN`, `Inf`, `+Inf`, `-Inf`) is rejected explicitly. The error message names the record number.
- **Quoted comma.** A category like `"a, b"` is accepted. The category value is the unquoted string `a, b` (after trimming).
- **Escaped quote.** A category like `"he said ""hi"""` is accepted. The category value is `he said "hi"` (after trimming).
- **Trailing newline.** A file with a trailing empty line produces no extra empty record when the reader is configured to ignore empty lines; the test must not produce spurious empty rows.
- **Whitespace inside fields.** Leading and trailing Unicode whitespace in the category field is trimmed. Internal whitespace is preserved as part of the category. The header itself is not trimmed; the exact two-field header is required.
- **Empty value field.** A data row whose second field is the empty string is rejected as a non-numeric value.
- **Streaming.** A 100,000-row input is processed correctly. The verification case in section 14 confirms correctness across many rows.

## 11. Project Constraints

- Go standard library only. No third-party CSV libraries, no third-party statistics libraries.
- The CSV reader is `encoding/csv`. The program does not hand-split on commas.
- The input seam is an `io.Reader` for the parser and a file path for the program. Tests use a `strings.Reader`, a `bytes.Reader`, or a custom incremental reader to drive large inputs.
- Aggregates are computed on the fly; the program does not hold every row in memory after parsing it.
- Output ordering is deterministic. Lexicographic sort by normalized category is the rule.
- Floating-point tolerance is documented and bounded. The test asserts counts and order exactly, and asserts sums and averages within the documented tolerance.
- The malformed-input policy is "all or nothing": either a complete report or a single named error.

## 12. Design Questions Before Coding

- Where does the parser live? In `main`, in a small package, or in two packages (a CSV reader and a report builder)? Which choice lets the test drive the parser through an `io.Reader` without going through the file system?
- How is the schema enforced? Through `csv.Reader.FieldsPerRecord`, through an explicit check on the header, or through both? Which choice fails fast on the first malformed row?
- How is the per-category state stored? In a `map[string]CategoryStats`, in a slice of pairs, or in a tree? Which choice makes the report order deterministic without an extra sort?
- How is the value field validated? Through `strconv.ParseFloat` plus an explicit `IsNaN`/`IsInf` check, or through a wrapper function? Which choice produces the cleanest error message?
- How is the empty-input case reported? As a clear "missing header" error, with the report suppressed? Which choice keeps the malformed-input policy uniform?
- How is the malformed-record error formatted? Through `fmt.Errorf`, through a typed error, or through a sentinel? Which choice lets the test assert on the record number without coupling to the wording, and which choice lets the program preserve the underlying `encoding/csv` error when no field can be identified?
- How is a large generated input injected? Through a `bytes.Reader` holding a pre-built string, or through a custom `io.Reader` that yields one record at a time across multiple `Read` calls? Which choice tests correctness across reader boundaries without depending on real time?

## 13. Implementation Milestones

1. Decide the package layout: a small parser package that takes an `io.Reader` and returns a report or an error, and a thin `main` package that opens the file and prints the report.
2. Define the report type. The exact field set is the learner's choice; it must include grand count, grand sum, grand average, and a slice of per-category entries each with normalized category name, count, sum, and average.
3. Read the first record and validate the header against the pinned schema. Reject any deviation with a clear error that names the record. Reject an empty input by detecting the absence of a first record.
4. Read subsequent records one at a time. Validate each row's field count, trim the category, check that the trimmed category is non-empty, parse the value, and reject NaN and infinity explicitly before updating aggregates.
5. Reject NaN and infinity explicitly with a check on the parsed `float64`.
6. Compute the grand statistics as the records are read. A category that appears once has an average equal to its single value; the test pins this.
7. Sort the per-category entries by normalized category name in lexicographic order before producing the output.
8. Build the output layout. The learner chooses the exact wording; the test pins it. Header-only input emits count `0`, sum `0`, average `0`, and no category rows.
9. Wire the program's command-line interface: a single positional argument naming the input file, an error to standard error on malformed input, the report to standard output on success.
10. Add a large-input test that drives a generated 100,000-row input through a custom reader that yields the input across multiple `Read` calls. The test asserts that the report's grand count is `100,000` and that the grand sum equals the expected sum within the documented tolerance.

## 14. Verification Cases the Learner Must Write

### Required Cases

Each case is described in natural language. The parser is driven through an `io.Reader`; the file system is only used for the program's command-line integration test. No test depends on real time, sleeps, or one-record-per-second pacing.

#### Header

- An input whose first record is `category,value` is accepted.
- An input whose first record is `Category,Value` (different case) is rejected with a clear error.
- An input whose first record is `category,value,extra` is rejected with a clear error.
- An input whose first record is `category` (one field) is rejected.
- An input whose first record is `name,amount` (different field names) is rejected.
- An input whose first record is `category,category` (duplicate field names) is rejected.
- An empty input (zero records) is rejected with a clear error that names the missing header.

#### Data rows

- An input with one data row produces a report with one category, count `1`, sum equal to the row's value (within the documented tolerance), average equal to that value.
- An input with five rows across two categories produces the correct grand count, grand sum (within tolerance), grand average (within tolerance), and correct per-category counts, sums, and averages.
- A quoted comma inside the category field is accepted and preserved as part of the category name after trimming.
- An escaped double quote inside the category field is accepted and reduced to a single quote.
- A leading or trailing space in the category field is trimmed; the trimmed category is the report key. Internal spaces are preserved.
- A category whose only characters are whitespace is rejected as an empty category.
- A row with one field is rejected with a record-numbered error.
- A row with three fields is rejected with a record-numbered error.
- A row with a non-numeric value (for example, the literal text `abc`) is rejected.
- A row with the literal text `NaN` is rejected with a clear "non-finite value" error.
- A row with the literal text `Inf` or `+Inf` or `-Inf` is rejected with a clear "non-finite value" error.
- A row whose value field uses exponent notation (for example, `1e3`) is accepted.

#### Order

- A test runs the parser against an input whose categories appear in a non-alphabetic order. The report's category rows are sorted lexicographically by normalized category. Two runs of the same test produce byte-identical output.

#### Edge cases

- A header-only input produces a report with grand count `0`, grand sum `0`, grand average `0`, and no category rows. The exit code is zero.
- An empty input is rejected with a clear error and a non-zero exit code.

#### Large input and streaming boundary

- A test constructs an input of 100,000 rows, each with a distinct or shared category. The test drives the parser through a custom `io.Reader` that yields the input across multiple `Read` calls (for example, one record per `Read`). The test asserts that the report's grand count is `100,000` and that the grand sum equals the expected sum within the documented tolerance. The test does not rely on `time.Sleep`, wall-clock pacing, or timing.

#### Tolerance

- A test asserts counts and order exactly. A test asserts sums and averages within the documented tolerance (small relative or absolute epsilon).
- A test runs the parser twice against the same input and confirms the two outputs are byte-identical.

#### Process

- An integration test runs the compiled binary against a temporary file with valid data and confirms the exit code is zero and standard output contains the report.
- An integration test runs the compiled binary against a temporary file with malformed data and confirms the exit code is non-zero and standard error contains the record number.

## 15. Common Mistakes to Watch For

- **Hand-splitting on commas.** `strings.Split(line, ",")` does not handle quoted commas, escaped quotes, or multi-line records. The project requires `encoding/csv`.
- **Forgetting to validate the header.** Reading the first record without comparing it to the pinned schema accepts any header. The schema check must happen before any data row is parsed.
- **Treating an empty input as a valid header-only input.** An empty input has no header and is rejected. A header-only input has exactly one record. The two cases are different.
- **Iterating a map directly to produce the report.** Go's map iteration order is randomized; the report will differ between runs. The project requires sorting the category keys before emitting them.
- **Treating NaN and infinity as valid floats.** `strconv.ParseFloat` accepts the strings `NaN`, `Inf`, `+Inf`, `-Inf` and returns a `float64`. The project requires an explicit check.
- **Producing a partial report on a malformed row.** The malformed-input policy is "all or nothing". A partial report is worse than no report.
- **Buffering the whole file in memory before parsing.** The streaming design is the point of the project. The verification in section 14 tests correctness across multiple `Read` calls without depending on real time.
- **Comparing floats with `==`.** Sums and averages are approximate; the test uses the documented tolerance. Counts and order compare exactly.
- **Using `fmt.Scan` or `bufio.Scanner` to read the CSV.** Both bypass `encoding/csv`'s quoting rules. The project requires `encoding/csv`.
- **Reporting the wrong record number.** Records are 1-based; the header is record 1; the first data row is record 2. The error messages must match this convention so a user can locate the row in their editor.
- **Silently trimming the header.** The header is matched exactly. Trimming applies only to the category field of data rows.
- **Silently ignoring parser errors that have no field context.** When the parser cannot identify a field, the program preserves and wraps the underlying `encoding/csv` error and includes the record number when one can be determined.
- **Writing a timing-based streaming test.** A test that writes one record per second, sleeps between records, or otherwise depends on real time is flaky and does not test streaming. Use a custom incremental reader and assert correctness across multiple `Read` calls.
- **Assuming a `bytes.Reader` proves the implementation never buffers everything.** A `bytes.Reader` confirms correctness; the streaming design is established by the implementation's structure (record-by-record processing, no slice of every row) and by the verification across multiple `Read` calls.

## 16. Topics and References for Study

- A Tour of Go: "Reading files", "Packages".
- Effective Go: "Errors", "Data".
- Package documentation: `encoding/csv` (`Reader`, `Reader.Read`, `Reader.FieldsPerRecord`, `Reader.TrimLeadingSpace`), `strconv` (`ParseFloat`), `math` (`IsNaN`, `IsInf`), `sort` (`Strings`, `Slice`), `bufio` (`NewReader`), `strings` (`TrimSpace`, `ToLower`), `unicode` (`IsSpace`).
- CSV pitfalls: search for "Go csv quoted comma", "csv injection", "RFC 4180 CSV format".
- Floating-point behavior: search for "Go float64 epsilon", "IEEE 754 NaN representation", "ParseFloat NaN Inf", "decimal binary representation".

## 17. Self-Assessment Questions

1. Why does the project require `encoding/csv` instead of `strings.Split` on commas?
2. Why is the header pinned to the exact string `category,value` and not trimmed, while category values are trimmed?
3. Why are NaN and infinity rejected explicitly, and what does `strconv.ParseFloat` accept by default?
4. Why are the per-category report rows sorted lexicographically, and what would break if they were emitted in map iteration order?
5. Why is "all or nothing" the right malformed-input policy, and what would a partial report imply?
6. Why does the test assert counts and order exactly but sums and averages only within a tolerance?
7. Why is an empty input rejected while a header-only input is accepted, and what distinguishes the two cases?
8. Why does the large-input test use a custom incremental reader instead of a `bytes.Reader`, and what does it confirm about correctness?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- [ ] The README's 19 sections are present in order; this file is the reference.
- [ ] Every functional requirement in section 8 is satisfied.
- [ ] Every verification case in section 14 has a corresponding test.
- [ ] The parser is driven through an `io.Reader`; the program's command-line integration is a thin wrapper that opens the file and prints the report.
- [ ] The report's output ordering is deterministic across runs on the same machine.
- [ ] Counts and order compare exactly; sums and averages compare within the documented tolerance.
- [ ] The package documentation states the schema, the empty-input policy, the header-only policy, the category normalization rule, the numeric syntax policy, the malformed-input policy, and the floating-point tolerance.
- [ ] The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Top-N categories.** Accept a flag that limits the report to the top-N categories by sum (with stable tie-breaking by category name). Out-of-top categories are aggregated into a single "other" row whose category name is fixed. Do not add filtering by value range or threshold.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisite:** [Project 017 — JSON Todo Persister](../../02-data-structures/017_json_todo_persister/README.md#20-prerequisite-based-documentation-guide).

Read the linked guide first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`encoding/csv`](https://pkg.go.dev/encoding/csv), [`sort`](https://pkg.go.dev/sort).
- **Standards and concept references:** [RFC 4180: CSV format](https://www.rfc-editor.org/rfc/rfc4180.html).

### Project-specific learning focus

- **Learn now:** streaming records, quoted fields, strict headers, numeric rejection, deterministic aggregation, and spreadsheet-formula injection risks.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
