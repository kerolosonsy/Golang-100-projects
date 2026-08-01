# Project 080 — DNS Lookup Tool

## 1. Project Name and Number
Project 080, dns_lookup_tool. This README is a learning guide only. You will create every source and test file yourself in `06-networking/080_dns_lookup_tool/`. This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea
A CLI that resolves a single validated ASCII hostname using a small injected resolver boundary. The required scope is host addresses and the canonical name; A and AAAA addresses are reported separately after textual canonicalization, deduplication of canonical equivalents, and lexical sort of the canonical strings. Canonical name is optional metadata. Errors are categorized using precedence on `errors.Is` before `errors.As`, with a pinned category enum that includes `Canceled`. Invalid input is rejected without ever calling the resolver. Hostname validation accepts only ASCII LDH characters, rejects any ASCII whitespace anywhere, allows at most one trailing root dot, and lowercases the result. The CLI passes that same normalized hostname to the resolver and uses it for display. Deterministic output lines and ordering are pinned in prose.

## 3. Why This Project Now?
This project requires Project 079 (load_balancer_round_robin) as the immediate predecessor and Project 071 (tcp_echo_server) for TCP framing, idle deadlines, and per-connection protocol error discipline; it does not require Project 060 because this project is a CLI without an HTTP server. Project 041 (context_timeout_example) is recommended review for context cancellation and deadlines and is optional review only, not a formal prerequisite. This project introduces the discipline of injecting a resolver boundary so unit tests run without public DNS, plus the discipline of distinguishing DNS error categories through `errors.Is` and `errors.As` precedence rather than string matching.

## 4. Prerequisites
Projects 079 and 071 are required prerequisites. Project 079 is the immediate predecessor for network-boundary discipline. Project 071 is required for TCP connection handling, byte framing, idle deadlines, accept-loop shutdown, and per-connection protocol error discipline. Project 041 is recommended review for context cancellation propagation but is not a formal prerequisite. No public DNS in required unit tests. One optional integration test for `localhost` is permitted and must be tolerant of the local environment returning IPv4 only, IPv6 only, or both. The resolver boundary is injected; production wires the standard library resolver and tests wire a fake.

## 5. What You Must Know Before Starting
Know the `net` package and `net.Resolver`, the standard library `*net.DNSError` fields `IsNotFound`, `IsTimeout`, and `IsTemporary`, `errors.Is` and `errors.As` precedence, successful empty answers, context cancellation and deadlines, ASCII hostname limits, IPv4 and IPv6 textual forms, and the race detector.

## 6. Explanation of New Concepts
The CLI accepts a single positional hostname argument. Any ASCII whitespace anywhere in the argument is rejected: the CLI does not trim and accept the trimmed value. A scheme prefix, a path, a port, or a non-ASCII byte is also rejected. Validation accepts only an ASCII LDH hostname with an optional single trailing root dot, lowercases the input, removes that single trailing dot when present, and validates the normalized form against the 1..253-byte total hostname bound. Multiple trailing dots are invalid; an empty interior label is invalid; a label that starts or ends with a hyphen is invalid; a label longer than 63 bytes is invalid. The normalized hostname is exactly the value passed to the resolver boundary and the value used for any display. The input `EXAMPLE.com.` therefore results in `example.com` being passed to the resolver.

IDNA conversion requires an external package and is intentionally out of the required scope; non-ASCII input is rejected honestly with a clear message and the resolver is not called.

The resolver boundary is a small interface that accepts a context and a normalized hostname and returns A addresses as a slice of strings, AAAA addresses as a slice of strings, a canonical name as an optional string, and an error. The boundary never returns a payload that the caller has to interpret in a different way. The boundary return value is the complete source of truth for the call.

The resolver error is handled before returned data: when the resolver returns a non-nil error, the CLI classifies that error by the pinned precedence and does not use accompanying address or canonical-name values. Only a nil-error result enters returned-data validation. Each address is parsed with the standard library address parser, verified to belong to its declared A or AAAA family, and re-emitted in canonical textual form. An unparseable or wrong-family value is a resolver-boundary contract violation categorized as `Other`, never silently discarded. Within each family, canonical equivalents deduplicate and canonical strings sort lexically. Empty valid slices with nil error are a successful empty answer.

Canonical name handling is optional metadata. An empty canonical name does not mean the input was already canonical and is not asserted to mean that; the canonical name is simply reported when the boundary provides a non-empty value. When the boundary provides a canonical name, the CLI normalizes it through the same lowercase rule and the same at-most-one trailing dot rule used for the hostname, then validates it against the same ASCII LDH rules. A canonical name that fails validation is recorded as `Other` and is not silently discarded; the resolver boundary is treated as the only source.

Error categorization uses `errors.Is` and `errors.As` precedence in the order pinned by this guide. The CLI first tests the resolver boundary's error against `context.Canceled` using `errors.Is`; on a match, the category is `Canceled`. The CLI then tests against `context.DeadlineExceeded` using `errors.Is`; on a match, the category is `Timeout`. The CLI then uses `errors.As` to extract a `*net.DNSError`; on match, the category is `NotFound` if `IsNotFound` is true, `Timeout` if `IsTimeout` is true, and `Temporary` if `IsTemporary` is true. Errors that do not match any of these branches are categorized as `Other`. The category enum is exactly `Canceled`, `NotFound`, `Timeout`, `Temporary`, and `Other`. A nil error with valid empty addresses is reported as a successful empty answer and is distinct from any category.

The required scope is host addresses plus canonical name. CNAME, MX, and TXT are not part of the required scope and are not produced unless the required contract explicitly pins those flags. The CLI does not parse raw DNS packets, does not run a DNS server, does not perform zone transfers, does not poison or manipulate caches, and does not implement a custom recursive resolver.

Output is deterministic for any given resolver return value. On success, the first line is the literal label `Host: ` followed by the normalized input. The next line is `Canonical: ` followed by the normalized canonical name or the single hyphen marker when absent. Each IPv4 address then gets one `A: ` line in sorted order, or one `A: -` line when none exist. Each IPv6 address then gets one `AAAA: ` line in sorted order, or one `AAAA: -` line when none exist. The final line is `Result: ok` when at least one address exists and `Result: empty` when both address families are empty. On resolver or returned-data failure, output is exactly one `Error: ` line followed by one of the five category names.

Unit tests use a fake resolver that records calls and returns pinned values. No test depends on public DNS. The fake resolver supports successful mixed IPv4 and IPv6 answers, duplicate and unordered answers, NXDOMAIN, timeout, cancellation, temporary failure, and other failure. Tests assert the categorical mapping through `errors.Is` and `errors.As`, the exact parsing and canonicalization, the exact sorting and deduplication, and the exact validation behavior including the no-call guarantee for invalid input.

One optional integration test resolves `localhost`. The integration test is tolerant of the local environment returning IPv4 only, IPv6 only, or both. The integration test does not assert external hostnames. The integration test is gated by a build tag or a test flag that allows opt-in execution; the default test run does not depend on it.

Text-only protocol examples are permitted. As a prose shape: `example.com` calls the resolver once with normalized `example.com`, then emits the exact `Host`, `Canonical`, sorted `A`, sorted `AAAA`, and `Result` lines. `EXAMPLE.com.` also resolves as `example.com`. Empty A and AAAA slices with no error produce `A: -`, `AAAA: -`, and `Result: empty`. Empty input prints usage and does not call the resolver; surrounding spaces, a scheme, an empty label, or multiple trailing dots are rejected without a call. A `*net.DNSError` with `IsNotFound` true produces `Error: NotFound`; one with `IsTimeout` true produces `Error: Timeout`; one with `IsTemporary` true produces `Error: Temporary`. An error wrapping `context.Canceled` produces `Error: Canceled`, one wrapping `context.DeadlineExceeded` produces `Error: Timeout`, and any remaining error produces `Error: Other`.

## 7. Learning Objective
Implement a hostname resolver CLI with an injected resolver boundary, exact hostname validation that rejects whitespace and trims nothing, exact normalization with at most one trailing dot, exact address parsing and canonicalization before classification, exact `errors.Is` and `errors.As` precedence across the five categories including `Canceled`, exact deterministic output lines and ordering, and tests that pin every branch through a fake resolver without public DNS.

## 8. Functional Requirements
1. The CLI accepts a single positional hostname argument.
2. Validation rejects empty input; rejects any ASCII whitespace anywhere; rejects scheme prefixes; rejects paths; rejects ports; rejects non-ASCII bytes.
3. Validation accepts only ASCII LDH characters per label; each label is 1..63 bytes inclusive; the normalized total hostname is 1..253 bytes inclusive.
4. Empty interior labels are invalid. Labels that start or end with a hyphen are invalid. Multiple trailing dots are invalid. A single trailing root dot is permitted and is removed during normalization.
5. Normalization lowercases the input and removes at most one trailing dot; the normalized hostname is the value passed to the resolver boundary and the value used for display.
6. IDNA conversion is out of required scope; non-ASCII input is rejected without calling the resolver.
7. The resolver boundary is an injected small interface that accepts a context and the normalized hostname and returns A addresses, AAAA addresses, a canonical name, and an error.
8. A non-nil resolver error is classified before any accompanying data is used. On nil error, each address is parsed, verified as the declared IPv4 or IPv6 family, and re-emitted canonically; an unparseable or wrong-family address is `Other` and is not silently discarded.
9. Within each family, canonical equivalents deduplicate and the canonical strings are lexically sorted in ascending order.
10. Canonical name handling is optional metadata. An empty canonical name is reported as a single line and is not asserted to mean the input was already canonical.
11. When the canonical name is non-empty, the CLI normalizes it through the same lowercase and at-most-one-trailing-dot rule, validates it against the same ASCII LDH rules, and reports it as a single line. A canonical name that fails validation is recorded as `Other` and is not silently discarded.
12. The category enum is exactly `Canceled`, `NotFound`, `Timeout`, `Temporary`, and `Other`.
13. Precedence on errors uses `errors.Is` before `errors.As`: a match against `context.Canceled` is `Canceled`; a match against `context.DeadlineExceeded` is `Timeout`; then `errors.As` to `*net.DNSError` with `IsNotFound` true is `NotFound`; with `IsTimeout` true is `Timeout`; with `IsTemporary` true is `Temporary`; remaining errors are `Other`.
14. A nil error with valid empty addresses is reported as a successful empty answer and is distinct from any category.
15. Successful output order is exactly `Host`, `Canonical`, zero or more `A`, zero or more `AAAA`, then `Result`, using a hyphen line for an empty family and `Result: empty` only when both families are empty. Error output is exactly one `Error: <Category>` line.
16. CNAME, MX, and TXT are not produced unless the required contract explicitly pins those flags.
17. The CLI does not run a DNS server, does not perform zone transfers, does not poison or manipulate caches, does not parse raw DNS packets, and does not implement a custom recursive resolver.
18. Unit tests use a fake resolver; no public DNS in required tests.
19. One optional integration test resolves `localhost` and is tolerant of the local environment returning IPv4 only, IPv6 only, or both.
20. Invalid input never increments the fake resolver call count.

## 9. Inputs and Outputs
CLI input is a single positional hostname argument. Optional input is a context with timeout or cancellation. Resolver boundary input is a context and the normalized hostname. Successful output uses the exact `Host`, `Canonical`, `A`, `AAAA`, and `Result` ordering and labels, with deterministic hyphen markers for absent values. Resolver or returned-data failure output is exactly one `Error: <Category>` line drawn from `Canceled`, `NotFound`, `Timeout`, `Temporary`, or `Other`. A successful empty answer ends in `Result: empty` and is not an error category.

## 10. Rules and Edge Cases
Empty input is rejected. Any ASCII whitespace anywhere is rejected. A scheme prefix is rejected. A path or port is rejected. A non-ASCII byte is rejected. An empty interior label is rejected. An over-length label is rejected. An over-length total hostname is rejected. Multiple trailing dots are rejected. A single trailing root dot is accepted and removed during normalization. The normalized hostname is exactly what is passed to the resolver and exactly what is used for display. Addresses returned by the resolver are parsed and canonicalized before classification; an unparseable address is `Other` and not silently discarded. Canonical equivalents deduplicate within each family. Lexical sort is ascending and is applied after canonicalization. An empty canonical name is reported as a single line; a non-empty canonical name is normalized and validated; an invalid canonical name is `Other`. The category enum is exactly `Canceled`, `NotFound`, `Timeout`, `Temporary`, and `Other`; precedence uses `errors.Is` before `errors.As`. A successful empty answer is distinct from any category. Context cancellation is observed and categorized through `errors.Is`. A deadline-exceeded signal is categorized as `Timeout` through `errors.Is`.

## 11. Project Constraints
No public DNS in required tests. The resolver boundary is injected; production wires the standard library resolver. The CLI is local and offline for required tests. IDNA is out of scope. CNAME, MX, and TXT are out of scope unless the required contract pins those flags. No DNS server, zone transfer, cache manipulation, raw packet parsing, or custom recursive resolver. No string matching on error text. No trimming of whitespace.

## 12. Design Questions Before Coding
How is the resolver boundary shaped so production wires the standard library and tests wire a fake? How is hostname validation ordered so invalid input never reaches the resolver? How is the normalized form guaranteed to be both the resolver argument and display value? Why is a non-nil resolver error classified before accompanying data, while nil-error data is then checked for parseability and correct address family? How does `errors.Is` precedence put cancellation and deadline outcomes before `*net.DNSError` classification? How is optional canonical metadata normalized and validated? How are the exact output labels, ordering, and empty markers kept deterministic? How does the fake prove invalid input makes no call? How is the optional `localhost` test gated?

## 13. Implementation Milestones
1. Define the validation rules: reject empty, reject any ASCII whitespace anywhere, reject scheme, reject path, reject port, reject non-ASCII bytes, accept only ASCII LDH labels with 1..63 bytes inclusive and a 1..253-byte total bound; reject empty interior labels, hyphen-leading or hyphen-trailing labels, and multiple trailing dots.
2. Define the normalization rules: lowercase, remove at most one trailing dot, validate the normalized result against the same ASCII LDH bound; the normalized value is exactly what is passed to the resolver boundary and used for display.
3. Define the resolver boundary as an injected small interface with the pinned return shape.
4. Define non-nil error classification before data use, then nil-error address parsing, family verification, canonicalization, per-family deduplication, and lexical sorting.
5. Define the canonical-name handling as optional metadata, normalized and validated through the same ASCII LDH rules when non-empty, and mapped to `Other` on validation failure.
6. Define the five-category enum `Canceled`, `NotFound`, `Timeout`, `Temporary`, `Other`.
7. Define the error precedence on `errors.Is` before `errors.As`: `context.Canceled` is `Canceled`; `context.DeadlineExceeded` is `Timeout`; then `errors.As` to `*net.DNSError` with `IsNotFound`, `IsTimeout`, and `IsTemporary`; remaining errors are `Other`.
8. Define the empty-successful-answer path as distinct from any category.
9. Define success output with exact `Host`, `Canonical`, `A`, `AAAA`, and `Result` labels and ordering, including hyphen markers and `ok` versus `empty`; define failure as exactly one `Error: <Category>` line.
10. Define the fake resolver with the pinned return shape and the call recording; ensure invalid input never increments the call count.
11. Define the full matrix of validation, canonicalization, sorting, deduplication, NXDOMAIN, timeout, temporary, cancellation, deadline-exceeded, empty answer, canonical-name metadata, unparseable address, invalid canonical name, and the optional `localhost` integration test.

## 14. Verification Cases the Learner Must Write
- A valid hostname calls the resolver boundary once and prints exact `Host`, `Canonical`, sorted `A`, sorted `AAAA`, and `Result: ok` lines in order.
- A CLI invocation with mixed IPv4 and IPv6 addresses returns each family in canonical textual form, deduplicated within family and lexically sorted ascending.
- A CLI invocation with duplicate addresses returns each address once after canonicalization within each family.
- An absent canonical name emits exactly `Canonical: -` and does not invent a value.
- A CLI invocation with a non-empty canonical name that normalizes to the input is reported as a single canonical-name line.
- A CLI invocation with a non-empty canonical name that fails validation is recorded as `Other` and is not silently discarded.
- A CLI invocation with `EXAMPLE.com.` normalizes to `example.com` and the resolver receives `example.com`.
- A CLI invocation with `example..com` is rejected for an empty interior label and does not call the resolver.
- A CLI invocation with `example.com.` followed by an extra `.` is rejected for multiple trailing dots and does not call the resolver.
- A CLI invocation with an empty argument prints a usage message and does not call the resolver.
- A CLI invocation with `   example.com   ` is rejected for whitespace anywhere and does not call the resolver.
- A CLI invocation with `http://example.com` is rejected as a scheme prefix and does not call the resolver.
- A CLI invocation with `example.com:80` is rejected as a port and does not call the resolver.
- A CLI invocation with a non-ASCII byte is rejected and does not call the resolver.
- A CLI invocation with an over-length label is rejected and does not call the resolver.
- A CLI invocation with an over-length total hostname is rejected and does not call the resolver.
- A nil-error resolver return containing an unparseable or wrong-family address is `Other` and is not silently discarded.
- A non-nil resolver error is categorized by error precedence even if the fake also supplies data; accompanying data is not used.
- A resolver boundary return wrapping `context.Canceled` is categorized as `Canceled`.
- A resolver boundary return wrapping `context.DeadlineExceeded` is categorized as `Timeout`.
- A resolver boundary return of a `*net.DNSError` with `IsNotFound` true is categorized as `NotFound`.
- A resolver boundary return of a `*net.DNSError` with `IsTimeout` true is categorized as `Timeout`.
- A resolver boundary return of a `*net.DNSError` with `IsTemporary` true is categorized as `Temporary`.
- A resolver boundary return of any other error is categorized as `Other`.
- A nil error with empty A and AAAA slices emits `A: -`, `AAAA: -`, and `Result: empty`, distinct from any error category.
- A nonempty address result ends in `Result: ok`; each address gets its own family-labelled line in canonical lexical order.
- The fake resolver records the call count and the exact normalized hostname argument; invalid input never increments the call count.
- No test depends on public DNS.
- The optional `localhost` integration test is tolerant of IPv4-only, IPv6-only, and mixed local environments; it does not assert external hostnames.
- All tests pass under the race detector.

## 15. Common Mistakes to Watch For
Calling the resolver for invalid input; trimming and accepting whitespace; string matching error text; inspecting returned data before classifying a non-nil error; omitting `Canceled`; mapping deadline expiry outside `Timeout`; silently discarding an unparseable or wrong-family address; treating empty canonical metadata as proof of canonicality; leaving nonempty canonical metadata unvalidated; deviating from the exact `Host`, `Canonical`, `A`, `AAAA`, `Result`, and `Error` lines; conflating successful empty with error; using public DNS in unit tests; implementing a DNS server, zone transfer, raw packet parser, cache manipulation, or custom recursion; and incrementing the fake call count for invalid input.

## 16. Topics and References for Study
Study the `net` package, `net.Resolver`, `*net.DNSError`, the `IsNotFound`, `IsTimeout`, and `IsTemporary` fields, `errors.Is` and `errors.As` precedence, successful empty answers, ASCII hostname limits, IPv4 and IPv6 textual canonicalization, context cancellation, and injected resolver boundaries for offline tests. Review the Go `net`, `context`, `errors`, and `sort` documentation. Read the prior README for Project 079 as the immediate predecessor for network-boundary discipline and Project 071 for TCP framing and protocol error discipline. Project 041 for cancellation propagation is optional review.

## 17. Self-Assessment Questions
Why is the resolver boundary injected rather than called directly from the CLI? Why is any ASCII whitespace rejected without trimming rather than trimmed and accepted? Why is the same normalized hostname passed to the resolver and used for display, and why does `EXAMPLE.com.` resolve as `example.com`? Why are addresses parsed and canonicalized before classification, and why is an unparseable address `Other` rather than silently discarded? Why is the precedence `errors.Is` before `errors.As`, why is `Canceled` an explicit category, and why is `context.DeadlineExceeded` resolved before the DNS branch? Why is the canonical name optional metadata rather than a guarantee that empty means the input was already canonical, and why is a non-empty canonical name normalized and validated through the same ASCII LDH rules? Why are the deterministic output lines and ordering pinned in prose, including empty sections, rather than left to impl-detail text? Why is CNAME, MX, and TXT out of required scope unless explicitly pinned, and why is IDNA out of required scope?

## 18. Definition of Completion
- [ ] Validation rejects empty input, any ASCII whitespace anywhere, schemes, paths, ports, non-ASCII bytes, empty interior labels, over-length labels, over-length total hostnames, and multiple trailing dots.
- [ ] Normalization lowercases the input and removes at most one trailing dot; the normalized value is exactly what is passed to the resolver boundary and used for display.
- [ ] The resolver boundary is an injected small interface; production wires the standard library resolver and tests wire a fake.
- [ ] Non-nil resolver errors are classified before accompanying data is used. Nil-error addresses are parsed, verified against their declared family, and emitted canonically; equivalents deduplicate and sort lexically.
- [ ] An unparseable or wrong-family address is `Other` and is not silently discarded.
- [ ] Canonical name handling is optional metadata; an empty canonical name is reported as a single line and is not asserted to mean the input was already canonical; a non-empty canonical name is normalized and validated through the same ASCII LDH rules; an invalid canonical name is `Other`.
- [ ] The category enum is exactly `Canceled`, `NotFound`, `Timeout`, `Temporary`, `Other`.
- [ ] Error precedence uses `errors.Is` before `errors.As`: `context.Canceled` is `Canceled`; `context.DeadlineExceeded` is `Timeout`; then `*net.DNSError` `IsNotFound`, `IsTimeout`, `IsTemporary`; remaining errors are `Other`.
- [ ] A successful empty answer is distinct from any category.
- [ ] Success output uses exact `Host`, `Canonical`, `A`, `AAAA`, and `Result` labels in order with hyphen markers; failure output is exactly one `Error: <Category>` line.
- [ ] CNAME, MX, and TXT are not produced unless the required contract explicitly pins those flags.
- [ ] No DNS server, zone transfer, cache manipulation, raw packet parsing, or custom recursive resolver.
- [ ] Unit tests use a fake resolver; no public DNS in required tests; invalid input never increments the call count.
- [ ] The optional `localhost` integration test is tolerant of the local environment returning IPv4 only, IPv6 only, or both; it does not assert external hostnames.
- [ ] All tests pass under the race detector.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions
Add an explicit CNAME flag whose behavior is pinned only when the required contract opts in, and whose default is off. Add a structured JSON output mode that preserves the same deterministic ordering as the text mode.
