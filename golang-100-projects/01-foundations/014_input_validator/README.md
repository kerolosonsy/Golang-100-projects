# Project 014 — Input Validator

## 1. Project Name and Number

Project **014** — `014_input_validator`. The directory name and number must match exactly.

## 2. Project Idea

A small package that validates a single record — a small struct holding an **email**, a **phone number**, and a **date** — by running three independent validators, one per field, and aggregating their results into a single report. Each validator is a self-contained function with a narrow contract and is independently callable on its own input; the aggregator simply runs all three against the matching fields of the record and collects whatever errors they produce.

> **Scope reminder.** A learning validator is **not** a full RFC 5321/5322 email implementation, **not** a libphonenumber-grade international phone verifier, and **not** a calendar library. The contracts below are deliberately limited. The README and the package documentation must say so plainly. Real users who need international correctness should adopt a dedicated library; this project is about the *shape* of validation code, not about covering every edge case of every format.

## 3. Why This Project Now?

Projects 011 through 013 taught the learner to keep behavior as data, to inject I/O boundaries, and to handle domain logic with declared policies. Project 014 introduces a different recurring shape: a small set of independent checks, each with its own rules, and an outer layer that combines their results into one report.

This shape will reappear in every later project — every form validation, every API request validator, every configuration checker. Getting it right at this level means the later projects inherit good habits rather than reinvent the same plumbing.

This project is also the first one where regex appears with restraint: enough to express the contracts, not so much that the pattern becomes an unreadable wall. The companion discipline is *parsing*: when a real parser exists, prefer it to a regex.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 014 therefore requires:

- Completion of **013** (Time World Clock), which provides the date-parsing experience this project builds on.
- No prior knowledge of HTTP, databases, or concurrency.

## 5. What You Must Know Before Starting

- What a regular expression is in Go, the basics of `regexp.MatchString` and `regexp.MustCompile`, and the cost of compiling a pattern many times.
- The difference between "match the whole string" (`^...$` or full-string methods) and "match a substring".
- How `time.Parse` parses a date from a string into a `time.Time`, and what a *layout* looks like in Go (the reference time `Mon Jan 2 15:04:05 MST 2006`).
- That `time.Parse` returns an error for a wrong date layout or an out-of-range calendar field; the project only requires rejection, not fine-grained classification of which kind of date error occurred.
- How Go represents Unicode: a `string` is a sequence of bytes; a `[]rune` is a sequence of code points; ranging over a string yields runes; ranging over a `[]byte` yields bytes.
- The `unicode` package and the difference between ASCII whitespace (a small fixed set) and Unicode whitespace (a much larger set defined by the Unicode standard).
- How to return multiple errors from one function using `errors.Join` (Go 1.20+) or a comparable aggregation technique.

## 6. Explanation of New Concepts

### One validator per field

The package exposes three validator functions — one for email, one for phone, one for date — each of which accepts a string and returns either success or a clear, named error. The functions do not call each other and do not share state. Each is independently callable from a test without going through the aggregator.

The aggregator takes a small record that holds three separate fields (an email field, a phone field, a date field) and runs each validator against its matching field. The aggregator's job is to call the three validators and to combine their results; it does not change the rules of any individual validator.

### Why a record, not one string for everything

Running all three validators against the same string makes no sense: a phone validator applied to an email address will always fail, and the user would have to read three meaningless errors. A record with one field per validator means each validator sees only the value that belongs to it, and the user can read the per-field errors in order.

### Moderate regex use

A regex is appropriate when the contract is "a string that matches a pattern". It is not appropriate when the contract involves arithmetic (for example, "a date that exists in the calendar"). For email, a moderate pattern covers the common shape. For phone, the pattern is even narrower: digits, an optional leading `+`, optional separators. For date, **regex is not the right tool**; the validator delegates to a real date parser. The rule for the learner is: use regex where it fits, use a parser where parsing is the contract.

### Real date parsing

`time.Parse` accepts a layout and a string and returns a parsed `time.Time` plus an error. The contract for this project is: the input must look like `YYYY-MM-DD` and must parse to a real calendar date. The validator parses with the chosen layout; if the parser succeeds, the input is accepted as a valid date. If the parser fails, the validator reports the failure. The project does not require the validator to distinguish between "wrong layout" and "impossible calendar content" — only that both kinds of failure are rejected.

### Impossible dates

`2025-02-30` looks like a date because it has the right shape, but February has 28 days in 2025. A naive regex would accept it; a real parser, however, would fail because `2025-02-30` is not a real calendar day. The contract for this project is: an impossible date is rejected by the date validator. The verification cases must include `2025-02-30`, `2025-04-31`, `2025-13-01`, `2025-00-10`, and similar impossibilities. The learner is not required to classify which specific error the parser produced; rejecting is enough.

### Unicode cases

A learning validator that supports Unicode must do so *deliberately*, not by accident.

- **Email.** The contract for this project is: there is one non-empty local part, exactly one `@`, and a domain made of at least two non-empty labels separated by dots. Domain dots cannot lead, trail, or be consecutive. Other characters may be any code point that is not Unicode whitespace, `@`, or — inside a domain label — `.`. The program must check for Unicode whitespace explicitly (for example via the `unicode` standard package), because a pattern that only excludes ASCII whitespace would let through a Unicode space and pretend the address was clean.
- **Phone.** The contract stays ASCII-only: ASCII digits, an optional leading `+`, and a small set of separators. The digit-count limits count ASCII digits and ignore separators.
- **Date.** ASCII-only because the layout is fixed.

### Aggregated errors

When the aggregator runs the three validators against the three fields, it does not stop at the first failure. It collects every failure into one report, so the user fixes all problems at once. The aggregation must preserve each individual error message so the user can see which field and which check failed and why. The mechanism is left to the learner (`errors.Join`, a custom slice, or another technique), but the *behavior* is required: one call produces one combined result that lists every failing field.

## 7. Learning Objective

After completing this project the learner can:

- Write three independent validators, each with a narrow contract documented in plain English.
- Run each validator against its own field of a record, not against the same string.
- Use `regexp` with restraint and `time.Parse` for real date validation.
- Reject impossible dates without promising to classify the specific failure mode.
- Handle Unicode inputs intentionally, including an explicit Unicode-whitespace check for the email field.
- Aggregate per-field errors into a single report that preserves each individual failure.
- Explain in plain English why this learning validator is not a full RFC or international implementation.

## 8. Functional Requirements

1. Expose three validator functions — one for email, one for phone, one for date. Each accepts a string and returns either `nil` or a non-nil error that names the failing check. The exact names are the learner's choice; this README uses *the email validator*, *the phone validator*, and *the date validator* in the prose.
2. Expose a record type with three fields: an email field, a phone field, and a date field (each a string). The exact field names are the learner's choice.
3. Expose an aggregator function that takes a record, calls the three validators against the three fields, and returns a single combined result. The aggregator's name is the learner's choice.
4. The email validator's contract:
   - Local part: one or more code points that are not Unicode whitespace and not `@`.
   - Exactly one `@`.
   - Domain part: non-empty labels separated by at least one `.`, with no leading, trailing, or consecutive dots. Every label contains one or more code points that are not Unicode whitespace, `@`, or `.`.
   - Unicode whitespace is checked explicitly; an ASCII-only whitespace check is not enough.
   - The character set is "any non-whitespace, non-`@` code point"; it is not restricted to ASCII letters and digits.
5. The phone validator's contract:
   - Optional leading `+` followed by ASCII digits, or ASCII digits only.
   - Allowed separators: single ASCII spaces and ASCII hyphens between non-empty digit groups. A separator cannot lead, trail, or appear twice in a row. **Parentheses are not allowed** in the required scope; this keeps the balanced-parens rule out of the contract.
   - Digit count: between **7 and 15** ASCII digits inclusive. Limits count ASCII digits only; spaces and hyphens are not digits and do not count toward the limit.
   - The empty string is rejected.
6. The date validator's contract:
   - Layout: `YYYY-MM-DD`.
   - Real calendar date (so `2025-02-30` is rejected).
   - Reject malformed input (wrong layout, non-numeric, empty).
   - The validator does not promise to classify whether the failure was wrong-layout or impossible-content; either kind of failure produces a rejection.
7. The aggregator runs the three validators against the three fields and aggregates the failures into a single combined result.
8. The aggregated report is a single value that lists every failing field, in a stable order (for example `email`, then `phone`, then `date`).
9. The package documentation states explicitly that this is a learning validator and not a full RFC or international implementation.

## 9. Inputs and Outputs

### Inputs

A record (a small struct) with three fields:

- `Email`: a string the user typed. Example: `alice@example.com`.
- `Phone`: a string the user typed. Example: `+1-555-123-4567`.
- `Date`: a string in `YYYY-MM-DD` form. Example: `2025-01-15`.

For per-validator testing, each validator accepts its own kind of string on its own.

### Outputs

- A combined report that lists which fields failed and why.
- For each individual validator, success or a named error when called on its own.

### Example text-only all-fields-valid run

```
email "alice@example.com": OK
phone "+1-555-123-4567": OK
date  "2025-01-15": OK
```

### Example text-only aggregated failure run

```
email "not an email":     local part or domain part contains whitespace or '@'
phone "abc":              contains non-digit, non-separator characters
date  "2025-02-30":       not a real calendar date
```

### Example impossible-date rejection

```
date "2025-02-30": not a real calendar date
```

## 10. Rules and Edge Cases

- **Empty input on any field**: every validator reports an "empty input" outcome for its own check.
- **Email without `@`**: rejected by the email validator.
- **Email with multiple `@`**: rejected by the email validator.
- **Email with a Unicode space (for example U+00A0 NO-BREAK SPACE) inside the local part**: rejected by the email validator, even though the space is not in the ASCII whitespace set.
- **Email with a Unicode local part such as `élève@example.com`**: accepted when the local part contains only non-whitespace, non-`@` code points.
- **Phone with letters**: rejected by the phone validator.
- **Phone with fewer than 7 ASCII digits**: rejected by the phone validator.
- **Phone with more than 15 ASCII digits**: rejected by the phone validator. The limit is inclusive.
- **Phone with parentheses**: rejected by the phone validator (parentheses are not in the allowed separator set).
- **Date with wrong layout** (for example `15-01-2025`): rejected by the date validator.
- **Date with right layout but impossible content** (`2025-02-30`, `2025-13-01`, `2025-00-10`): rejected by the date validator.
- **Aggregated behavior**: when more than one field fails, the report lists every failure, not just the first.
- **Aggregator order**: the order of the failures in the report is stable so tests can assert it.

## 11. Project Constraints

- Go standard library only. No third-party validation libraries. The point is to *write* the contracts, not to import them.
- The email, phone, and date validators are independent functions. The email function does not call the phone function or vice versa.
- Each validator is callable on its own from a test, with a string of its own kind. The aggregator is a separate function that takes the record and calls the three validators.
- Regex is used for email and phone; the date validator uses a real parser, not a regex.
- The package documentation explicitly states the limits: this is a learning validator, not a full RFC 5321/5322, E.164, or international date implementation.
- No persistence, no configuration files, no environment-variable parsing — out of scope.
- No internationalization of error messages — error wording is the learner's choice, but the *behavior* must be testable.

## 12. Design Questions Before Coding

- What are the field names of the record? Where is the record defined, and how does the aggregator reach the validators?
- Where does the phone digit limit live — in the function, in a package-level constant, or in a struct the caller constructs?
- Will the email regex anchor to the start and end of the string, or will you use a full-string match method? Which is more readable here?
- How will you reject Unicode whitespace when a regex on its own does not catch every Unicode space? Will the function walk the string rune by rune and check the whitespace property for each code point, in addition to or instead of the regex?
- How will the aggregator combine errors? Will you use `errors.Join`, a custom slice type, or something else? What does each error look like in the report?
- How will the Unicode cases be tested? Will you write the Unicode strings as literals or as escape sequences, and why?
- How will the package documentation make the "learning scope" visible at the top, so a future reader does not mistake the package for a production validator?
- How will the tests pin the contract — by table of records and expected outcomes, or by separate `TestEmail`, `TestPhone`, `TestDate`, `TestAggregator` functions? Which makes the per-field structure clearer?

## 13. Implementation Milestones

1. Write the package documentation that declares the three contracts, the per-field record, and the "learning scope" disclaimer.
2. Define the record type with `Email`, `Phone`, `Date` string fields.
3. Implement the email validator with the moderate pattern, plus an explicit Unicode-whitespace check that complements the regex.
4. Implement the phone validator with the declared ASCII-digit limits and the no-parentheses separator rule.
5. Implement the date validator using `time.Parse` and the `YYYY-MM-DD` layout; rejection is enough, no error classification required.
6. Implement the aggregator that takes a record, calls the three validators against the three fields, and combines the failures into a single result with stable order.
7. Confirm that each validator is reachable from a test without going through the aggregator.
8. Confirm that the aggregator is reachable from a test that constructs the record directly.

## 14. Verification Cases the Learner Must Write

Each case is described in natural language. The expected outcomes pin the contract.

### Per-validator cases

- A canonical email like `alice@example.com` is accepted by the email validator.
- An email missing `@` is rejected by the email validator.
- An email with two `@` is rejected by the email validator.
- An email whose domain starts or ends with a dot, or contains two consecutive dots, is rejected by the email validator.
- An email with a Unicode local part such as `élève@example.com` is accepted by the email validator.
- An email with a non-ASCII space such as U+00A0 in the local part is rejected by the email validator.
- A canonical phone like `+1-555-123-4567` is accepted by the phone validator.
- A phone with letters is rejected by the phone validator.
- A phone with parentheses such as `+1 (555) 123-4567` is rejected by the phone validator (parentheses are not in the allowed separator set).
- A phone with fewer than 7 ASCII digits is rejected; a phone with more than 15 ASCII digits is rejected. Limits are inclusive and count ASCII digits only.
- A canonical date like `2025-01-15` is accepted by the date validator.
- An impossible date like `2025-02-30` is rejected by the date validator.
- A wrong-layout date like `15-01-2025` is rejected by the date validator.
- An out-of-range date like `2025-13-01` is rejected by the date validator.
- An empty input produces a non-nil error from every validator.

### Aggregator cases

- A record with all three fields valid produces a combined result with no failures; the "all-fields-valid" example from section 9 is reachable.
- A record with two invalid fields produces a combined result that lists both failures, in a stable order (for example `email` then `phone`).
- A record with all three fields invalid produces a combined result that lists all three failures, in the same stable order.
- The aggregator's report preserves the field order so tests can assert it.

## 15. Common Mistakes to Watch For

- **Running all three validators against the same string.** That destroys the per-field contract; an all-valid aggregate becomes impossible. Each validator runs against its own field of the record.
- **Using regex for dates.** A regex can check the layout but cannot check the calendar; impossible dates will slip through. Always parse.
- **Anchoring the email pattern weakly.** A pattern that does not anchor to the end of the string will accept trailing whitespace. Anchor explicitly.
- **Relying on the regex alone for whitespace.** Many Unicode space code points exist; an ASCII-only `[ \t]` class misses them. Add an explicit rune-by-rune Unicode whitespace check on top of the regex.
- **Promising fine-grained date error categories.** The project requires rejection, not classification. Do not promise to label each failure as "wrong layout" or "impossible content".
- **Stopping at the first failure.** The aggregator's whole point is to surface every problem. Returning on the first failure defeats that.
- **Counting Unicode bytes instead of runes.** A phone limit expressed in bytes will misjudge ASCII digits in ways that depend on the encoding; the rule is "count ASCII digits in the rune sequence, separators do not count".
- **Promising too much.** Saying "valid email" without the "learning scope" disclaimer misleads the next reader. The disclaimer is part of the contract.
- **Compiling the regex on every call.** The pattern should be compiled once at package level, not on every validation.

## 16. Topics and References for Study

- A Tour of Go: "Methods", "Errors", "Packages".
- Effective Go: "Package names", "Package comments", "Errors are values".
- Package documentation: `regexp` (`MatchString`, `MustCompile`, anchoring), `time` (`Parse`, layout reference time), `errors` (`Join` introduced in Go 1.20).
- Unicode in Go: search for "Go strings bytes runes", "unicode/utf8", "unicode.IsSpace".
- Aggregated errors: search for "errors.Join Go 1.20", "multi-error patterns".
- Honest validation scope: search for "RFC 5322 email regex is impossible", "libphonenumber scope".

## 17. Self-Assessment Questions

1. Why is a regex the wrong tool for date validation, and what is the right tool?
2. Why does the aggregator take a record with separate fields instead of running all three validators against one string?
3. Why must impossible dates be rejected by the date validator, and which part of the implementation enforces that?
4. Why is the order of errors in the aggregated report important, and how does your code make the order stable?
5. What does "learning scope" mean in the package documentation, and which lines make that scope visible?
6. How does the email validator handle Unicode whitespace? Why is "use the `unicode` package, not just a regex character class" the right answer?
7. How does the phone validator count digits? Why is the rule "ASCII digits only, separators do not count" the honest one for this project?
8. If you had to add a fourth validator (for example, a postal code), what shape would it have, and which aggregator contract would it satisfy?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every functional requirement in section 8 is satisfied.
- Every verification case in section 14 has a corresponding test, and the tests pin the per-field and aggregator contracts with concrete expected outcomes.
- The package documentation declares the three contracts, the per-field record, and the learning-scope disclaimer at the top.
- Each validator is reachable from a test that calls it on its own string.
- The aggregator is reachable from a test that constructs the record directly.
- The aggregator's behavior — every failure listed, order stable — is exercised by at least one test, and an all-valid record produces a result with no failures.
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Per-field severity.** Tag each error with a severity (`info`, `warning`, `error`) and let the aggregator filter by minimum severity. Keep the tag system trivial: an enum-like type and a switch, not a full rule engine.
- **Optional date layout.** Let the date validator accept a small set of alternative layouts (for example `YYYY/MM/DD` in addition to `YYYY-MM-DD`) and try each in turn. The aggregator's behavior is unchanged; only one validator grows a small list.
