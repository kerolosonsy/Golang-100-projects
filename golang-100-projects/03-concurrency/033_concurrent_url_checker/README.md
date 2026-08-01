# Project 033 — Concurrent URL Checker

## 1. Project Name and Number

Project 033 — Concurrent URL Checker. Lives in `03-concurrency/033_concurrent_url_checker`.

## 2. Project Idea

Check a bounded list of URLs concurrently with HTTP HEAD using an injected `http.Client`, and return one result per input URL in the original input order including duplicates. The result distinguishes a small fixed set of health categories so callers can act on it without re-parsing the response. The input bound is 64 entries per call; longer inputs are rejected at preflight with no goroutines launched. A valid URL parses as an absolute `http` or `https` URL with a non-empty host; everything else is malformed per-item. The checker uses a no-follow redirect policy; the first 3xx is reachable and its body is closed. Caller-context cancellation is reported as cancelled; transport-level timeouts on an active caller context are transport error. Production wiring uses a client with an explicit timeout. Tests use a local `httptest.Server` and never touch the public internet.

## 3. Why This Project Now?

Project 031 gave you fan-out with cancellation, and Project 032 fixed channel ownership at small scale. This project scales the same primitives to real HTTP requests and replaces the manual signals with the `http.Client`. The next project, 034, tightens the same fan-out into a fixed worker pool; this project uses a goroutine-per-URL shape that the worker-pool project will refine. Now that you have cancellation habits, channel ownership, and a result-boundary habit, you can apply them to a real package boundary and watch what the `net/http` package does to the design.

## 4. Prerequisites

The curriculum map's stated dependencies for this project: Projects 031 and 032. You must be comfortable launching goroutines, fanning out with per-item indexed ownership, channel ownership and close rules for transcript-style channels, and `context` cancellation. Familiarity with the `net/http` and `httptest` packages is required.

## 5. What You Must Know Before Starting

That `http.Client` is the unit of configuration for timeouts, transport, and proxy policy. That `http.Client.CheckRedirect`, when set, controls redirect behavior; an unset function falls back to the Go default which follows redirects. That `http.NewRequest` paired with `client.Do` is the entry point for HEAD. That `http.Response.Body` is an `io.ReadCloser` that must be closed in every code path, including non-2xx responses and HEAD responses. That the `httptest` package exposes `httptest.NewServer` and how URLs are composed against it. That `context.WithTimeout` produces a context that cancels at the deadline and exposes `Done`. That `errors.Is` compares wrapped errors against a sentinel. That `url.Parse` validates structure but not reachability, and that an unparseable URL is a per-item error that does not abort the rest. That this project defines a hard input bound of 64 URL entries per call. That this project defines a valid URL as a string that parses as an absolute URL whose scheme is `http` or `https` and whose host is non-empty.

## 6. Explanation of New Concepts

A "concurrent URL checker" walks a list and returns, for each entry, a small categorical verdict. Five categories are pinned by this project: reachable for 2xx and 3xx, HTTP error status for 4xx and 5xx, malformed URL for entries that fail URL validation, transport error for failures to connect or read at the transport level, and context cancelled for entries cancelled through the supplied request context.

Input bound. The checker accepts at most 64 URL entries per call. Inputs longer than 64 entries are rejected at the call-level preflight step with no goroutines launched. Empty input is valid and short-circuits to an empty slice with no goroutines launched. The bound is the input ceiling that the worker-pool project will replace with a fixed worker count.

Valid URL. A valid URL is a string that parses as an absolute URL whose scheme is `http` or `https` and whose host is non-empty. Relative references, other schemes (for example `mailto` or `ftp`), missing-host URLs, and URLs that fail parsing are reported at their input positions with the malformed status, and the valid entries still run. The validation rule is per-item and synchronous so the result slice is in input order without surprises.

Redirect policy. The checker does not automatically follow redirects. The injected or derived client uses Go's standard no-follow behavior that returns the first redirect response to the caller without turning that response into a redirect error. The first 3xx response is classified as reachable, and its body is closed. Tests pin this policy with the standard `http.ErrUseLastResponse` sentinel rather than an arbitrary redirect error, because an arbitrary error would change the outcome into a client error instead of a reachable 3xx response.

Caller-context cancellation versus transport timeout. Cancellation through the supplied request context — `ctx.Done()` fires, or `ctx.Err()` reports cancelled or deadline exceeded — is reported as cancelled. A client-level timeout or any other transport-level timeout that fires while the caller's context is still active is reported as transport error, not as cancelled. The distinction is observable: the test asserts on the result category together with the request context's `Err()`.

Because the result slice is indexed by input position and each entry is written by exactly one goroutine, there is no shared mutable cell. Each goroutine writes to its own slot under indexed ownership, which is what makes the slice race-free even without locks.

The scope boundary between this project and the worker-pool project is that 033 uses a goroutine per URL, which is fine for the bounded input size that this project assumes. The worker-pool project trades the goroutine-per-URL shape for a fixed number of workers when the input is unbounded.

## 7. Learning Objective

After this project you can configure an injected `http.Client` with explicit timeouts and a no-follow redirect policy, and you can verify that response bodies are closed on every code path. You can apply a fixed categorical health policy to a fan-out of HTTP requests, including the cancel-versus-transport-timeout distinction. You can pre-validate the input without preventing the valid entries from running, and you can enforce the 64-entry input bound. You can decide between "goroutine per work item" and "fixed worker pool" and state when each is appropriate. You can prove overlap among requests without using `time.Sleep`.

## 8. Functional Requirements

1. The checker accepts a slice of URL strings and a context, and returns a slice of results of length equal to the input.
2. Hard input bound: the input is rejected at call-level preflight when its length exceeds 64 entries. Empty input is valid and returns an empty slice with no goroutines launched.
3. Each URL is pre-validated before any goroutine is launched. A valid URL is a string that parses as an absolute URL whose scheme is `http` or `https` and whose host is non-empty. Entries that fail validation are reported at their input positions with the malformed status, and the valid entries still run.
4. The checker accepts an injected `http.Client`. Production wiring constructs one with a `Timeout` and a no-follow redirect policy. Test wiring constructs one against `httptest.NewServer`. The checker must not depend on the default client or on the default Go redirect policy.
5. The HTTP method used by the checker is HEAD.
6. Responses with status in the 2xx or 3xx range are reported as reachable. The first 3xx response is reachable because the client does not follow redirects; its body is closed.
7. Responses with status in the 4xx or 5xx range are reported as HTTP error status. Transport-level failures (refused connection, DNS error, read failure) are reported as transport error.
8. Caller-context cancellation — the supplied request context is cancelled or deadline-exceeded — is reported as cancelled for any entry that had not produced a terminal status. A client or transport timeout that fires while the caller's context is still active is reported as transport error, not as cancelled.
9. The response body is closed in every code path, including for HEAD responses, for 3xx reachable responses, and for non-2xx responses.
10. The result slice is in the original input order. Duplicate URLs are valid input and produce independent results at their respective input positions.
11. Cancellation through the context cancels every outstanding request and returns the cancelled status at the corresponding input positions for entries that had not completed.

## 9. Inputs and Outputs

**Input** is a slice of URL strings and a context. The slice length must be at most 64; longer slices are rejected at preflight.

**Output** is a slice of length equal to the input. Each entry carries the input index, the URL string as supplied, and one of five pinned statuses: reachable, HTTP error, malformed URL, transport error, or cancelled. The slice is in input order.

**Behaviour example (text only).** With input `["http://a", ":/broken", "http://b", "http://a"]` of length four, the result slice has four entries in that order. The middle entry is malformed URL. The other three entries are reachable, HTTP error, or transport error depending on what the test server returned, each at the correct position.

**Behaviour example (text only).** With input `[]`, the result slice is empty and no goroutines are launched.

**Behaviour example (text only).** With input of length 65 — exceeding the input bound — the call is rejected at preflight, no goroutines are launched, and no requests are made.

## 10. Rules and Edge Cases

Hard input bound: a call with more than 64 entries is rejected at preflight with no goroutines launched. Empty input is valid and returns an empty slice with no goroutines launched. Duplicate URLs are independent; their results are at their input positions. Malformed URLs are detected synchronously in preflight and do not prevent valid entries from running. The HTTP method is HEAD; the body is closed in every code path. A response status in the 4xx or 5xx range is an HTTP error status, not a transport error. A connection refused, DNS failure, or read error is a transport error. Context cancellation that arrives before a request begins produces cancelled. Context cancellation that arrives while a request is in flight unblocks the in-flight request and produces cancelled for that entry. A client-level or transport-level timeout while the caller's context is still active is reported as transport error, not as cancelled. The first 3xx response is reachable under the no-follow redirect policy. Two entries that fail in different categories are reported in their respective categories; categories are never merged.

## 11. Project Constraints

Standard library only. Tests use `httptest.Server` only; no public internet. The injected `http.Client` is the only client configuration the checker depends on; no global default client and no default redirect policy. Every response body is closed, including for HEAD and for any reachable 3xx response and for error status. The result slice is race-free; each entry is written by exactly one goroutine under indexed ownership. No goroutine leak on the cancellation path. The package context is propagated as a parameter into the request and used to cancel the round-trip. The input size is bounded by the caller at 64 entries per call; the worker-pool project assumes a larger or unbounded input. The race detector must report nothing under `-race`.

## 12. Design Questions Before Coding

How is the categorical status represented so that all five cases are distinct at the type level? How does the checker encode a malformed URL versus a transport-level failure versus an HTTP error status, when the runtime cannot tell them apart from a single error string alone? Where does the injected `http.Client` live, and how does the test substitute its own client and own `CheckRedirect`? How is the result slice race-free when many goroutines each write to their own slot, and what synchronization, if any, is required to coordinate the collector? How does the request context cancellation reach the in-flight HTTP round-trip without being confused with a transport-level timeout? When the test server returns 5xx, what counts as HTTP error rather than transport error? Where does the 64-entry input bound live in the call, and how is the rejection path kept without goroutines?

## 13. Implementation Milestones

1. Decide on the result type that carries the input index, URL, status, and any error details.
2. Implement the call-level preflight: hard input bound of 64 entries, and empty-input short-circuit.
3. Implement the per-URL validation pass using the parse + scheme + host rule; mark invalid entries malformed at their positions.
4. Decide how the injected client is passed and where the no-follow redirect policy is set in production wiring.
5. Implement the per-entry fan-out. Each goroutine writes to its own slot under indexed ownership.
6. Implement the request, including the HEAD method, the body close, the no-follow redirect policy, and the status categorization.
7. Implement the cancel-versus-transport-timeout classification by observing both `ctx.Err()` and the transport error.
8. Implement the cancellation path that propagates the context into the request and converts it into the cancelled status.
9. Decide how to assemble the final slice in input order and confirm it is race-free.
10. Add tests with multiple handlers (200, 404, 500, 3xx) and synchronization channels that prove overlap without `time.Sleep`.
11. Run under `-race` and confirm no race report.

## 14. Verification Cases the Learner Must Write

Empty input returns an empty slice and launches no goroutines. Input of length 65 is rejected at preflight with no goroutines launched. All-reachable input against `httptest.NewServer` returns one reachable entry per URL. Mixed-status input returns reachable entries for 2xx and reachable entries for an un-followed 3xx, and HTTP error entries for 4xx and 5xx, in input order. Malformed URL entries — including relative references, non-http(s) schemes, and missing-host URLs — are reported as malformed URL at their positions while the rest of the slice still runs. A test server that closes the connection mid-HEAD returns transport error at the affected input position. A test with a controllable caller-context deadline that fires before some requests complete returns cancelled at the affected positions and reachable at the others. A test with the caller's context still active but the client timing out returns transport error at the affected positions, not cancelled. Duplicate URLs in the input produce duplicate result entries at their respective positions. The no-follow redirect policy uses the standard sentinel that returns the first response without a redirect error; the first 3xx is reported as reachable and its body is closed. Running under `-race` produces no race report. The handler is registered with synchronization channels or barriers that prove the requests overlapped, without `time.Sleep`. The injected client is honoured: a custom client configuration propagates to the requests.

## 15. Common Mistakes to Watch For

Forgetting to close the response body on non-2xx status and on the un-followed 3xx. The `defer` only runs at the function exit, and an early return on a "non-2xx" branch can leak. Confusing HEAD with GET; GET forces a body to be drained, which this project does not use. Treating 4xx as transport error rather than as HTTP error status. Letting two goroutines write to the same result slot under "first wins"; indexed ownership is mandatory. Classifying a transport-level timeout as cancelled because the request context's `Err()` happened to read as a deadline; the category is caller-context cancellation only, while the caller's `Err()` remains nil. Letting the test client follow redirects by relying on the Go default redirect policy. The project pins a no-follow policy for test and production clients; the first 3xx is reachable. Treating the input as unbounded or as larger than 64 entries. Using `http.DefaultClient` in production, which has no timeout and uses the default redirect policy. Failing to register a `CheckRedirect` whose behaviour the test pins.

## 16. Topics and References for Study

The `net/http` package documentation, especially `Client`, `Request`, `Response`, `CheckRedirect`, and `Transport`. The `httptest` package documentation, especially `httptest.NewServer`, `httptest.NewRecorder`, and URL composition against the test server. The `context` package documentation, especially `context.WithTimeout`, `context.WithCancel`, and how requests carry a context. The Effective Go notes on HTTP clients. The blog post on the complete guide to `net/http` timeouts for the distinction between transport, request, and response-header timeouts, and the relationship between client timeout and the request context.

## 17. Self-Assessment Questions

What is the difference between HEAD and GET in this project, and which paths is it easy to forget to close a body on? What is the difference between transport error and HTTP error status, and how does the test distinguish them? What is the difference between caller-context cancellation and a transport-level timeout, and how does the test prove the distinction? Why is malformed URL a per-item preflight check rather than a runtime check, and why does it not block the rest of the slice? How does the no-follow redirect policy affect the classification of a 3xx response, and what proves the policy is in place? How does the result slice stay race-free under many concurrent writes? Why is `time.Sleep` disallowed in the overlap proof, and what primitive is used instead? When the caller context is cancelled mid-request, what does the cancelled status mean for an entry whose handler already produced 200? What is the role of the 64-entry input bound, and where in the call does the preflight reject larger inputs? Why is the worker-pool project the right next step, rather than reusing the goroutine-per-URL shape here?

## 18. Definition of Completion

Every Functional Requirement is implemented and exercised by a passing test. The Behaviour Examples in this README hold. Tests run with `-race` produce no race report. No public internet is touched in any test. The result slice is always exactly as long as the input and always in input order. The injected `http.Client` is the only client configuration the checker depends on, and the no-follow redirect policy is pinned by the test. Every code path closes the response body, including for non-2xx HEAD responses and for un-followed 3xx reachable responses. The cancellation path unblocks in-flight requests and reports cancelled at the affected positions; transport-level timeouts on an active caller context are reported as transport error. Empty input returns an empty slice with no goroutines launched. Input above 64 entries is rejected at preflight with no goroutines launched. You can answer every Self-Assessment Question without consulting the README.

## 19. Optional Extensions

Add a connectivity-type breakdown (TCP, TLS, certificate name mismatch) that the checker surfaces as a sub-category alongside the top-level status, with tests against `httptest.NewTLSServer` and against a deliberately misnamed certificate. Add a per-host concurrent slot limit so a single host with many URLs does not monopolize the fan-out, with a test that proves the slot count never exceeds the limit and that the slots above the limit report transport error or any category the design pins.
