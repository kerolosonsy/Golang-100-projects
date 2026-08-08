# Project 082 — Simple Port Scanner

## 1. Project Name and Number

- Project 082, simple_port_scanner.
- This README is a learning guide only.
- You will create every source and test file yourself in `06-networking/082_simple_port_scanner/`.
- This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

- > **Authorization and safety note.** This tool is restricted to TCP connect scanning of loopback addresses on the local machine.
- It must never be used against public hosts, networks you do not own, or systems for which you lack written authorization.
- Scanning networks without authorization may violate law and policy.
- The required scope intentionally excludes raw SYN scans, UDP, service fingerprinting, stealth scans, CIDR expansion, hostname resolution, and any public target.

## 2. Project Idea

A bounded TCP connect scanner that accepts only numeric loopback IPv4 and IPv6 literals and rejects every hostname without resolving it. The scanner scans only the inclusive port range that the caller has explicitly enumerated, validates the lower and upper bounds and the span, and uses a bounded worker pool with positive concurrency and a positive per-dial timeout. Each port produces exactly one typed result; the categories are `Open`, `Closed`, or `Error` with cause `Timeout`, `Unreachable`, `Canceled`, or `Other`. A successful connection is closed exactly once immediately after classification. The result list is sorted ascending by port regardless of completion order. A unit test boundary uses a fake dialer so tests do not depend on real listeners; an optional integration test opens a few loopback ephemeral listeners and scans exactly the ports those listeners bound.

## 3. Why This Project Now?

- This follows Project 081 (ssh_remote_executor) as the immediate predecessor.
- Project 071 (tcp_echo_server) is the other formal prerequisite for loopback TCP listener lifecycle.
- Project 041 (context_timeout_example) is recommended review for context cancellation and deadlines, but is not a formal prerequisite.
- This project introduces the discipline of narrowing a scanning primitive to a safe local scope, the discipline of validating the host and the bounds before any dial, and the integration of bounded concurrency with the correct cancellation semantics.

## 4. Prerequisites

- Project 081 is the immediate predecessor and required prerequisite; Project 071 is also a formal prerequisite.
- Project 041 is recommended review for context cancellation and deadlines, but is not a formal prerequisite.
- The scanner accepts only numeric loopback IPv4 literals `127.0.0.1` and IPv6 literals `::1`, and other numeric IPs only when `net.IP.IsLoopback` returns true.
- Hostnames and the literal `localhost` are rejected without any DNS resolution.
- The required unit tests use a fake dialer; the optional integration test opens loopback ephemeral listeners and scans exactly the ports those listeners bound.

## 5. What You Must Know Before Starting

- Know the `net` package and the `net.Dialer` configuration, the difference between TCP connect scans and raw socket scans, numeric loopback address validation for IPv4 and IPv6 using `net.IP.IsLoopback`, the rejection of `localhost` and any other hostname without DNS resolution, inclusive port bounds, the difference between a connection refused outcome and a timeout outcome, the difference between a context cancellation outcome and a context deadline outcome, a bounded worker pool with a job channel and a results channel, classification through `errors.Is` and `errors.As` precedence rather than string matching, the race detector, and the principle that every opened connection is closed deterministically exactly once.

## 6. Explanation of New Concepts

### Concepts

- The host argument must be a numeric IP literal whose `net.IP.IsLoopback` returns true.
- The literals `127.0.0.1` and `::1` are always accepted; other numeric IPs are accepted only when their `net.IP.IsLoopback` returns true.
- The literal `localhost` is rejected without any DNS resolution; every other hostname is rejected without any DNS resolution; CIDR ranges and public addresses are rejected.
- Validation runs before dial and is the boundary that protects both the user and the targets.

- The port range is inclusive.
- The lower bound and the upper bound are 1..65535 inclusive; the lower bound is not greater than the upper bound; the inclusive span, which is the count of ports from the lower bound to the upper bound, is capped at exactly 1024 ports.
- A larger span is rejected before any dial.
- The concurrency is a positive integer between 1 and 128 inclusive; zero or negative values are rejected.
- The per-dial timeout is a positive duration; a non-positive value is rejected; the production default is 500 milliseconds.
- The effective peak in-flight dial count is the minimum of the configured concurrency and the configured span; a barrier test whose span is greater than or equal to the concurrency proves the exact configured peak and asserts that the in-flight peak never exceeds it.

- The classification is exact and final.
- A successful result has status `Open` only when the connection is a non-nil usable connection; the connection is closed exactly once immediately after classification, and a close failure is treated as observability only, never as a port reclassification.
- A connection refused signal is mapped to `Closed`.
- Errors are mapped to `Error` with a typed cause.
- The mapping uses `errors.Is` and `errors.As` precedence rather than string matching.
- The precedence is pinned: caller context `Canceled` first; caller or per-dial `DeadlineExceeded` or a timeout-capable network error next; a wrapped connection-refused signal maps to `Closed`; a wrapped network-unreachable or host-unreachable signal maps to `Unreachable`; any remaining error maps to `Other`.
- The implementation never inspects error text.

- Cancellation is correct end-to-end.
- On caller cancellation, the scanner stops dispatching new jobs, cancels or invalidates any in-flight dial, joins every worker, and returns a nil results slice together with a typed context cancellation or deadline error.
- The implementation never returns a partial result set on caller cancellation; the typed result is the only observation.
- The phrase "exactly one result per port" applies only to normal completion; on cancellation the result is nil and the typed error is the observation.

- Channel ownership is exact.
- One owner closes the jobs channel; one owner closes the results channel after every worker has finished.
- No goroutine sends on a closed channel; the implementation never leaks a worker past the join point.

- The integration test is optional and is loopback only.
- The test opens a few loopback ephemeral listeners, observes each bound port, and issues a one-port scan for each bound port; the asserted outcome for each bound port is `Open`.
- The integration test never scans a port that is not bound by a listener it opened; it never races by guessing an unused real port.
- The IPv6 integration variant is skipped only when loopback IPv6 is unavailable on the test host.
- The integration test never contacts a public target.

- Text-only protocol examples are permitted.
- As a prose shape: the scanner validates the numeric loopback address, the lower and upper bound, the span, the positive concurrency, and the positive per-dial timeout;
- It then runs a bounded worker pool that dials each port with a per-dial timeout, classifies the outcome by category, closes any successful connection exactly once immediately, and appends a result.
- On normal completion, results are sorted ascending by port and printed one per line in the form `port: <category>` for `Open` and `Closed` and `port: error: <typed-cause>` for `Error`.
- On caller cancellation, the scanner stops dispatching, joins the workers, and returns a nil results slice together with a typed context cancellation or deadline error.
- The optional integration test prints the same shape for the ports its ephemeral listeners bound.

## 7. Learning Objective

- Implement a TCP connect scanner restricted to numeric loopback IPv4 and IPv6 literals with explicit inclusive bounds, a 1024-port inclusive span cap, positive concurrency between 1 and 128, a positive per-dial timeout with a 500-millisecond production default, deterministic classification into `Open`, `Closed`, or `Error` with cause `Timeout`, `Unreachable`, `Canceled`, or `Other` through `errors.Is` and `errors.As` precedence rather than string matching, immediate close of every successful connection exactly once, results sorted ascending by port only on normal completion, correct cancellation that returns a nil results slice and a typed context error, channel ownership that prevents send-on-closed and worker leaks, and tests through a fake dialer plus an optional loopback ephemeral listener integration test.

## 8. Functional Requirements

1. The scanner accepts only numeric loopback IPv4 and IPv6 literals whose `net.IP.IsLoopback` returns true; the literals `127.0.0.1` and `::1` are accepted; the literal `localhost` and every other hostname are rejected without any DNS resolution.
2. The lower bound and upper bound are 1..65535 inclusive; the lower bound is not greater than the upper bound; the inclusive span is capped at exactly 1024 ports.
3. Concurrency is a positive integer between 1 and 128 inclusive; a non-positive value is rejected. The effective peak in-flight dial count is the minimum of the configured concurrency and the span.
4. The per-dial timeout is a positive duration; a non-positive value is rejected; the production default is 500 milliseconds.
5. The scanner dispatches at most the configured concurrency at any moment; a barrier test whose span is greater than or equal to the concurrency proves the exact configured peak and asserts that the in-flight peak never exceeds it.
6. Each port produces exactly one typed result on normal completion. The status is `Open` or `Closed`, or the status is `Error` with cause `Timeout`, `Unreachable`, `Canceled`, or `Other`.
7. Classification uses `errors.Is` and `errors.As` precedence rather than string matching. The precedence is: caller context `Canceled` first; caller or per-dial `DeadlineExceeded` or a timeout-capable network error next; wrapped connection-refused maps to `Closed`; wrapped network-unreachable or host-unreachable maps to `Unreachable`; remaining errors map to `Other`.
8. A `Open` result is emitted only when the connection is a non-nil usable connection; the connection is closed exactly once immediately after classification. A close failure is observability only and is never a port reclassification.
9. On caller cancellation the scanner stops dispatching new jobs, cancels or invalidates in-flight dials, joins every worker, and returns a nil results slice together with a typed context cancellation or deadline error. The implementation never returns a partial result set on caller cancellation.
10. One owner closes the jobs channel; one owner closes the results channel after every worker has finished. No goroutine sends on a closed channel; no worker leaks past the join point.
11. Results are sorted ascending by port only on normal completion; the returned slice is deterministic for any given input set, dialer set, and concurrency.
12. The scanner does not perform SYN scans, UDP scans, service fingerprinting, stealth scans, banner grabs, or CIDR expansion.
13. The required unit tests use a fake dialer. The optional integration test opens a few loopback ephemeral listeners and scans exactly the ports those listeners bound; the asserted outcome for each bound port is `Open`. Close, timeout, and unreachable are fake-dialer tests rather than races against unused real ports.
14. The IPv6 integration test is skipped only when loopback IPv6 is unavailable on the test host.
15. No required test contacts a public host.

## 9. Inputs and Outputs

### Interface Contract

- Input is the numeric loopback host literal, the inclusive lower bound, the inclusive upper bound, the positive concurrency between 1 and 128, the positive per-dial timeout, and the context.
- Output on normal completion is a sorted slice of results in ascending port order, where each result is a typed record containing the port number, the status, and the cause when the status is `Error`.
- Output on caller cancellation is a nil results slice together with a typed context cancellation or deadline error.
- The integration test output is the same shape for exactly the ports the ephemeral listeners bound.

## 10. Rules and Edge Cases

- A non-loopback identifier is rejected before dial.
- The literal `localhost` and every hostname are rejected without DNS resolution.
- An empty host is rejected.
- An invalid lower bound is rejected.
- An invalid upper bound is rejected.
- A lower bound greater than the upper bound is rejected.
- A span greater than 1024 is rejected.
- A non-positive concurrency is rejected.
- A concurrency greater than 128 is rejected.
- A non-positive per-dial timeout is rejected.
- Each port produces exactly one result on normal completion. `Open` is emitted only when the connection is non-nil and usable.
- The successful connection is closed exactly once immediately;
- A close failure is observability only. `Closed` is the status for a connection refused signal. `Timeout` is the cause for a per-dial deadline or a timeout-capable network error. `Unreachable` is the cause for a wrapped network-unreachable or host-unreachable signal. `Canceled` is the cause for a caller context cancellation.
- The remaining errors map to `Other`.
- The implementation never inspects error text.
- On caller cancellation the scanner stops dispatching, joins the workers, and returns a nil results slice together with a typed context cancellation or deadline error.
- The phrase "exactly one result per port" applies only to normal completion.
- The results are sorted ascending by port only on normal completion.
- The fake dialer records the in-flight peak so the concurrency assertion is by event.
- The optional integration test scans exactly the ports its ephemeral listeners bound.

## 11. Project Constraints

- Loopback numeric IP only.
- No raw sockets.
- No SYN scans.
- No UDP scans.
- No service fingerprinting.
- No stealth scans.
- No banner grabs.
- No CIDR expansion.
- No public targets.
- No DNS resolution of any hostname.
- No `localhost` literal.
- The unit test boundary is a fake dialer.
- The optional integration test is loopback ephemeral listeners only.
- Tests use synchronization events, not sleep, to observe concurrency and cancellation.

## 12. Design Questions Before Coding

- Why is the validator the boundary that protects both the user and the targets, and why is DNS resolution of any hostname excluded?
- Why is the inclusive span capped at exactly 1024 ports and why is concurrency required to be between 1 and 128 inclusive?
- Why is the effective peak in-flight dial count the minimum of the configured concurrency and the span, and why is the concurrency assertion performed by a barrier test rather than by sleep?
- Why is the classification decided by `errors.Is` and `errors.As` precedence rather than string matching, and why are `Timeout`, `Unreachable`, `Canceled`, and `Other` typed causes rather than raw strings?
- Why is `Open` emitted only when the connection is a non-nil usable connection and why is the close failure observability only and never a port reclassification?
- Why does caller cancellation stop dispatching, join the workers, and return a nil results slice with a typed context error, and why is the phrase "exactly one result per port" restricted to normal completion?
- Why is the channel ownership single-owner-per-channel and why is send-on-closed forbidden?
- Why is the integration test optional and why are Closed, Timeout, and Unreachable fake-dialer tests rather than races against unused real ports?

## 13. Implementation Milestones

1. Define the host validator that accepts only numeric loopback IPv4 and IPv6 literals whose `net.IP.IsLoopback` returns true; reject `localhost` and every other hostname without DNS resolution.
2. Define the bound validator that requires lower and upper to be 1..65535 inclusive, requires lower not greater than upper, and caps the inclusive span at exactly 1024 ports.
3. Define the concurrency validator that requires a positive integer between 1 and 128 inclusive and rejects zero, negative, or out-of-range values.
4. Define the timeout validator that requires a positive duration and rejects zero or negative values; the production default is 500 milliseconds.
5. Define the worker pool with a job channel and a results channel whose effective peak in-flight dial count is the minimum of the configured concurrency and the span.
6. Define the classification through `errors.Is` and `errors.As` precedence rather than string matching: caller context `Canceled` first; caller or per-dial `DeadlineExceeded` or a timeout-capable network error next; wrapped connection-refused maps to `Closed`; wrapped network-unreachable or host-unreachable maps to `Unreachable`; remaining errors map to `Other`.
7. Define the close step that runs exactly once immediately after a successful classification and that never reclassifies the port on a close failure.
8. Define the cancellation discipline that stops dispatching, cancels or invalidates in-flight dials, joins every worker, and returns a nil results slice with a typed context cancellation or deadline error.
9. Define the channel ownership that closes the jobs channel from its single owner and closes the results channel from its single owner only after every worker has finished.
10. Define the deterministic sort that orders results ascending by port only on normal completion.
11. Define the fake dialer with pinned per-port outcomes, per-port invocation recording, and in-flight peak recording for the concurrency assertion; the optional loopback ephemeral listener integration test that scans exactly the ports its listeners bound and asserts `Open` for each bound port, with the IPv6 variant skipped only when loopback IPv6 is unavailable; and the full matrix of validation, bound, span, concurrency, timeout, open, closed, fake timeout, fake unreachable, real cancellation, ordering, concurrency peak, immediate close, and the optional loopback integration test.

## 14. Verification Cases the Learner Must Write

### Required Cases

- A non-loopback identifier is rejected before dial and never opens a socket.
- The literal `localhost` is rejected before dial without any DNS resolution.
- A hostname is rejected before dial without any DNS resolution.
- An empty host is rejected before dial.
- An invalid lower bound is rejected before dial.
- An invalid upper bound is rejected before dial.
- A lower bound greater than the upper bound is rejected before dial.
- A span greater than 1024 is rejected before dial.
- A non-positive concurrency is rejected before dial.
- A concurrency greater than 128 is rejected before dial.
- A non-positive per-dial timeout is rejected before dial.
- A single-port scan returns exactly one result on normal completion, classified by category, with the fake dialer recording exactly one invocation.
- An inclusive range scan returns exactly the requested count on normal completion, classified by category, with each port called exactly once.
- A connection refused signal is `Closed`.
- A per-dial deadline is `Timeout`.
- A wrapped network-unreachable or host-unreachable signal is `Unreachable`.
- A caller context cancellation observed during a dial is `Canceled`.
- A remaining error is `Other`.
- Classification uses `errors.Is` and `errors.As` precedence rather than string matching; an error whose text contains the substring "timeout" but whose cause is not a deadline is mapped through precedence and never by text.
- An `Open` result is emitted only when the connection is a non-nil usable connection; the successful connection is closed exactly once immediately; a close failure is observability only and never a port reclassification.
- The in-flight peak recorded by the fake dialer equals the minimum of the configured concurrency and the span; the peak never exceeds the configured concurrency across the test duration, asserted by event and not by sleep.
- A barrier test whose span is greater than or equal to the concurrency proves the exact configured peak and asserts that the in-flight peak never exceeds it.
- On caller cancellation the scanner stops dispatching, joins the workers, and returns a nil results slice together with a typed context cancellation or deadline error; a partial result set is never returned.
- On caller cancellation the jobs channel is closed by its single owner and the results channel is closed by its single owner only after every worker has finished; no goroutine sends on a closed channel and no worker leaks past the join point.
- Results are sorted ascending by port only on normal completion; the fake dialer returns outcomes in reverse order and the sorted output is still ascending.
- The optional loopback integration test opens a few ephemeral listeners and scans exactly the ports those listeners bound; the asserted outcome for each bound port is `Open`.
- The optional IPv6 integration variant is skipped only when loopback IPv6 is unavailable on the test host.
- The optional integration test never scans a port that is not bound by a listener it opened; it never races by guessing an unused real port.
- All tests pass under the race detector with no sleep synchronization.
- No required test contacts a public host.

## 15. Common Mistakes to Watch For

- Resolving arbitrary hostnames or the literal `localhost` through DNS;
- Accepting CIDR ranges;
- Allowing SYN scans through the design;
- Mixing timeout, unreachable, canceled, and other transport errors into a single errored text;
- Classifying by raw error strings;
- Sleeping to observe concurrency;
- Using a global mutex that masks the in-flight peak;
- Reading bytes from a successful connection;
- Closing a successful connection more than once or zero times;
- Reclassifying a port on a close failure;
- Returning a partial result set on caller cancellation;
- Sending on a closed results channel;
- Leaking a worker past the join point;
- Scanning public hosts to make tests pass;
- Disabling the validator because the test fixture happens to be loopback;
- Asserting the peak against the configured concurrency alone when the span is smaller than the concurrency;
- Guessing an unused real port for a closed or timeout test.

## 16. Topics and References for Study

- Study the `net` package and `net.Dialer`, `net.Listener`, numeric loopback address validation with `net.IP.IsLoopback`, the difference between TCP connect scans and raw socket scans, inclusive port bounds, the difference between a connection refused signal and a timeout signal, `errors.Is` and `errors.As` precedence rather than string matching, context cancellation and deadline, bounded worker pools with channels, channel ownership and send-on-closed prevention, synchronization events rather than sleep, and the race detector.
- Review the Go `net`, `context`, `errors`, and `sync` documentation.
- Read the prior README for Project 081 for the immediate predecessor discipline, Project 071 for the required TCP lifecycle foundation, and Project 041 for cancellation propagation as optional recommended review.

## 17. Self-Assessment Questions

1. Why is the validator the boundary that protects both the user and the targets, and why is DNS resolution of any hostname excluded?
2. Why is the inclusive span capped at exactly 1024 ports and why is concurrency required to be between 1 and 128 inclusive?
3. Why is the effective peak in-flight dial count the minimum of the configured concurrency and the span, and why is the concurrency assertion performed by a barrier test rather than by sleep?
4. Why is the classification decided by `errors.Is` and `errors.As` precedence rather than string matching, and why are `Timeout`, `Unreachable`, `Canceled`, and `Other` typed causes rather than raw strings?
5. Why is `Open` emitted only when the connection is a non-nil usable connection and why is the close failure observability only and never a port reclassification?
6. Why does caller cancellation stop dispatching, join the workers, and return a nil results slice with a typed context error, and why is the phrase "exactly one result per port" restricted to normal completion?
7. Why is the channel ownership single-owner-per-channel and why is send-on-closed forbidden?
8. Why is the integration test optional and why are Closed, Timeout, and Unreachable fake-dialer tests rather than races against unused real ports?

## 18. Definition of Completion

- [ ] The scanner accepts only numeric loopback IPv4 and IPv6 literals whose `net.IP.IsLoopback` returns true; the literal `localhost` and every hostname are rejected without DNS resolution.
- [ ] The lower bound and upper bound are 1..65535 inclusive; the lower bound is not greater than the upper bound; the inclusive span is capped at exactly 1024 ports.
- [ ] Concurrency is a positive integer between 1 and 128 inclusive; the effective peak in-flight dial count is the minimum of the configured concurrency and the span; the peak never exceeds the configured concurrency.
- [ ] The per-dial timeout is a positive duration with a 500-millisecond production default; context cancellation is observed.
- [ ] Each port produces exactly one typed result on normal completion; the status is `Open` or `Closed`, or the status is `Error` with cause `Timeout`, `Unreachable`, `Canceled`, or `Other`.
- [ ] Classification uses `errors.Is` and `errors.As` precedence rather than string matching: caller context `Canceled` first; caller or per-dial `DeadlineExceeded` or a timeout-capable network error next; wrapped connection-refused is `Closed`; wrapped network-unreachable or host-unreachable is `Unreachable`; remaining errors are `Other`.
- [ ] An `Open` result is emitted only when the connection is a non-nil usable connection; the successful connection is closed exactly once immediately; a close failure is observability only and never a port reclassification.
- [ ] On caller cancellation the scanner stops dispatching, cancels or invalidates in-flight dials, joins every worker, and returns a nil results slice together with a typed context cancellation or deadline error; a partial result set is never returned.
- [ ] The jobs channel is closed by its single owner; the results channel is closed by its single owner only after every worker has finished; no goroutine sends on a closed channel and no worker leaks past the join point.
- [ ] Results are sorted ascending by port only on normal completion; the returned slice is deterministic for any given input set, dialer set, and concurrency.
- [ ] Required tests use a fake dialer; the optional loopback integration test scans exactly the ports its ephemeral listeners bound and asserts `Open` for each bound port; the IPv6 variant is skipped only when loopback IPv6 is unavailable.
- [ ] No SYN, UDP, service fingerprinting, stealth, banner grab, or CIDR expansion is performed.
- [ ] All tests pass under the race detector with no sleep synchronization.
- [ ] No required test contacts a public host.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions

- Add an explicit per-port attempt count exposed through the fake dialer so the test can assert a single attempt per port even under cancellation.
- Add a small structured summary that reports the counts of each category but never raw error strings or target identifiers beyond the loopback numeric literal.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 081 — SSH Remote Executor](../../06-networking/081_ssh_remote_executor/README.md#20-prerequisite-based-documentation-guide), [Project 071 — TCP Echo Server](../../06-networking/071_tcp_echo_server/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- None. This project applies already introduced APIs, standards, and testing practices in a new combination.

### Project-specific learning focus

- **Learn now:** authorized loopback-only connect scans, inclusive port ranges, refusal versus timeout, typed network errors, bounded workers, single-attempt accounting, cancellation, and deterministic tests.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
