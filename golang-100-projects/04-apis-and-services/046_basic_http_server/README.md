# Project 046 — Basic HTTP Server

## 1. Project Name and Number

- Project 046 — Basic HTTP Server, located in `046_basic_http_server`.

## 2. Project Idea

Build a small `net/http` service with two public resources. `GET /healthz` returns a fixed plain-text health response, while `GET /` renders a supplied display message into an HTML page through `html/template`. The service must have precise routing and method behavior, independently testable handlers, and explicit production server limits.

## 3. Why This Project Now?

- This project begins the HTTP and REST stage of the roadmap.
- Project 014 contributes the validation mindset needed for exact request contracts, and Project 017 contributes separation between domain behavior and external I/O.
- The new step is to apply those habits to HTTP handlers, response metadata, templates, in-memory tests, and production server configuration.
- Earlier networking, context, and concurrency projects may be useful review, but they are not required prerequisites.

## 4. Prerequisites

- Complete Projects 014 and 017 before starting.
- No other project is a required prerequisite.

- You should already be able to validate inputs against an explicit contract, return and inspect errors, separate pure behavior from I/O startup, and write deterministic tests around an injected dependency.

## 5. What You Must Know Before Starting

- An HTTP response consists of a status, headers, and an optional body. Headers must be finalized before the first body write because a body write commits an implicit success status when no status was written explicitly.
- `http.Handler` is the testable boundary between a request and a response. Handler construction does not require a listening socket.
- `httptest` can exercise a handler entirely in memory. Such tests must not reserve a fixed port or depend on machine networking.
- `html/template` performs contextual HTML escaping. It is different from assembling markup with string concatenation and different from `text/template`.
- HTTP methods are case-sensitive tokens. This project supports only the method explicitly listed for each known path.
- A response to `HEAD` has no body even when the selected status is an error. This project rejects `HEAD` rather than treating it as an implied `GET`.
- A server timeout limits a particular phase of connection or response handling. No combination of server timeouts proves that every slow handler will stop promptly.
- Once response bytes have been committed, a later write error cannot be converted into a different clean HTTP response.

## 6. Explanation of New Concepts

### Concepts

- A handler is an object that receives one HTTP request and writes one HTTP response.
- Keeping handler construction separate from network startup makes routing and rendering testable without opening a socket.
- Production startup should receive a completed handler rather than construct hidden global routes.

- A template separates fixed markup from supplied data. `html/template` understands HTML contexts and escapes data before insertion.
- For example, angle brackets in a display message become text rather than executable markup.
- The template itself is trusted application material; the supplied display message is untrusted data.

- `httptest` supplies an in-memory request and response recorder.
- It is appropriate for status, header, body, routing, and escaping assertions.
- It is not a faithful way to prove kernel socket deadlines, so timeout values are verified by inspecting the production server configuration instead of waiting for real time to pass.

- `ReadHeaderTimeout` limits how long the server permits for reading request headers. `ReadTimeout` limits request reading, including the body, according to `net/http` server semantics. `WriteTimeout` limits the response-writing period according to those same semantics. `IdleTimeout` limits how long an idle keep-alive connection may wait for another request. `MaxHeaderBytes` bounds request-header parsing size.
- These controls reduce exposure to resource exhaustion, but they do not cancel every long computation, guarantee an upstream deadline, or stop a handler that ignores request cancellation.

- A normal server shutdown is represented by the standard server-closed sentinel error.
- Startup code treats that outcome as expected and distinguishes it from binding, configuration, or serving failures that must be reported.

## 7. Learning Objective

- By completion, you can define an exact HTTP contract, build handlers independently from a listener, render untrusted display data safely through `html/template`, and verify handler behavior with `httptest`.
- You can also explain what each production server timeout limits, what it does not limit, how maximum header size differs from body limits, and why a normal server close is not an operational failure.

## 8. Functional Requirements

1. Build one handler tree independently from all network startup. Tests receive that handler directly.
2. `GET /healthz` returns status `200 OK`, content type exactly `text/plain; charset=utf-8`, and a body consisting exactly of lowercase `ok` followed by one line-feed character.
3. `GET /` returns status `200 OK` and content type exactly `text/html; charset=utf-8`.
4. The root response is rendered from a pre-parsed `html/template`. The supplied display message appears as text in the page and is contextually escaped. Raw string concatenation must not place the message into markup.
5. A template parsing failure prevents handler construction or startup. A rendering failure before response commitment produces a controlled server error rather than a partial HTML page.
6. Any syntactically valid path other than exactly `/` or `/healthz` returns `404 Not Found`, content type `text/plain; charset=utf-8`, and lowercase `not found` followed by one line-feed character. Query parameters do not change path selection.
7. Any method other than `GET` on either known path returns `405 Method Not Allowed` and an `Allow` header whose value is exactly `GET`.
8. For unsupported methods other than `HEAD`, the `405` body is lowercase `method not allowed` followed by one line-feed character and uses `text/plain; charset=utf-8`.
9. `HEAD` is explicitly unsupported. `HEAD` on a known path returns `405` with `Allow: GET`, the plain-text content type, and no body. `HEAD` on an unknown path returns `404` and no body.
10. A response-body write failure is surfaced to an injected, test-observable reporting boundary where the handler design can observe it. The handler never attempts a second response after commitment.
11. Production startup constructs an `http.Server` with explicit nonzero read-header, read, write, and idle timeouts and an explicit positive maximum header size. These values are centralized so tests can inspect them without sleeping.
12. Returning the standard normal-server-close error is treated as expected shutdown. Every other startup or serving error is reported as a failure with its cause preserved.

## 9. Inputs and Outputs

### Interface Contract

- The service accepts HTTP requests.
- It has no request-body contract, and both successful routes ignore query parameters.
- The display message is supplied when the handler is assembled; it is not read from a global variable or concatenated into HTML.

- Text-only health example: a `GET` request for `/healthz` produces status `200`, content type `text/plain; charset=utf-8`, and exactly three body bytes: `o`, `k`, and a line feed.

- Text-only page example: when the supplied message contains angle brackets around a script-like phrase, `GET /` produces HTML in which those characters are escaped and displayed as text.
- The raw script-like phrase must not appear as active markup.

- Text-only method example: a `POST` request for `/healthz` produces status `405`, an `Allow` value of `GET`, the plain-text error content type, and the pinned method-error body.
- A `HEAD` request for the same path produces the same status and `Allow` value but zero response-body bytes.

## 10. Rules and Edge Cases

- Path matching is exact and case-sensitive. `/healthz`, `/Healthz`, `/healthz/`, and `/anything` are different paths.
- Only the first is the health resource.
- The root path remains `/` even when a query string is present.

- A supplied display message may contain ampersands, quotes, angle brackets, Unicode, or an empty string.
- The page remains valid HTML and the value is handled as template data.
- Tests must distinguish escaped output from the unescaped source rather than assuming that a visual browser rendering is sufficient evidence.

- No request body is needed.
- A body attached to a valid `GET` does not alter the response contract and is not interpreted.
- This project does not add a body limit because no handler reads a body.

- The handler writes each normal body once.
- If the underlying writer reports a short write or an error after commitment, the error is observable where the design permits, but the handler does not call `http.Error`, rewrite the status, or append a second error document.

- Timeout tests inspect the configured values and their association with the correct server fields.
- They do not use sleeps, intentionally slow handlers, or assumptions about scheduler timing.
- Handler-level cancellation and graceful shutdown orchestration remain outside this project's required scope.

## 11. Project Constraints

- Use only the Go standard library.
- Use `net/http`, `net/http/httptest`, and `html/template` for their intended responsibilities.
- Do not use a web framework, third-party router, raw HTML concatenation for supplied data, package-level mutable router state, or fixed test ports.

- Do not add implementation examples, copied solutions, generated handlers, or timing-based tests to this guide.
- Production configuration must include all four required timeouts and a maximum header size, but the guide and implementation must not claim that those settings stop every slow handler or replace request-context deadlines.

## 12. Design Questions Before Coding

- Where will the parsed template and supplied display message live so handlers remain independent from network startup and global mutable state?
- What response information must be decided before any body bytes are committed?
- How will template parsing and rendering failures remain distinguishable from response-writer failures?
- What reporting boundary makes a write failure observable in a unit test without attempting an impossible second response?
- How will exact path selection avoid an accidental catch-all root handler?
- How will `HEAD` remain bodyless while still returning its pinned `404` or `405` status and headers?
- Where will production server limits be stored so tests can inspect them without opening a listener?
- Which startup result represents expected closure, and how will all other causes remain visible?

## 13. Implementation Milestones

1. Record the complete route, method, status, header, body, and `HEAD` policy as testable acceptance criteria.
2. Establish a handler-construction boundary that accepts the page data and required error-reporting dependency without starting a listener.
3. Prepare and validate the trusted HTML template during setup.
4. Complete the exact health response and its method handling.
5. Complete the escaped HTML response and its pre-commit rendering-failure behavior.
6. Complete exact unknown-path, unsupported-method, and bodyless `HEAD` behavior.
7. Make observable response-write failures stop after the first failed commitment attempt.
8. Define the production server configuration with all required limits and document each limit's scope.
9. Connect production startup to the completed handler and distinguish expected server closure from real serving failures.
10. Finish deterministic handler and configuration tests, then review all response contracts byte for byte where exactness is required.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Verify `GET /healthz` returns the exact status, exact content type, and exact three-byte body.
- Verify `GET /` returns `200`, the exact HTML content type, and a complete rendered page containing the supplied ordinary message.
- Supply a message containing angle brackets, quotes, an ampersand, and script-like text. Verify the response contains escaped text and does not contain the raw active markup.
- Verify an empty display message and a Unicode display message render successfully.
- Verify `/missing`, `/healthz/`, and a case-changed health path each return the exact `404` contract.
- Verify unsupported methods on both known paths return `405`, exact `Allow: GET`, the exact error content type, and the exact method-error body.
- Verify `HEAD` on each known path returns `405`, exact `Allow: GET`, and zero body bytes. Verify `HEAD` on an unknown path returns `404` and zero body bytes.
- Verify a query string does not change either known route's result.
- Force a template rendering failure before commitment using a test-controlled template value or writer boundary and verify a controlled server error with no partial HTML.
- Use a writer that reports a write failure in a path where the handler can observe it. Verify the failure is reported once and no second response is attempted.
- Inspect the production server configuration and verify each timeout is nonzero, associated with the intended field, and accompanied by a positive maximum header size. Do not use real delays.
- Verify the normal server-close sentinel is accepted as expected and a different serving error remains an error with its original cause discoverable.
- Run all handler tests through `httptest`; no test binds a fixed port or depends on external networking.

## 15. Common Mistakes to Watch For

- Registering a root handler as an unchecked catch-all can make unknown paths return the home page instead of `404`.
- Letting `net/http` infer content type can make exact contracts depend on body bytes.
- Writing a body before selecting a status silently commits `200`.
- Treating `HEAD` as an automatic `GET` violates this project's explicit method policy, while writing an error body for `HEAD` violates HTTP response semantics.

- Using `text/template` or string concatenation for HTML can turn supplied data into active markup.
- Parsing a template on every request adds avoidable work and moves configuration failures into request handling.
- Rendering directly to the response writer can expose a partial page before an encoding failure is known.

- Testing timeouts with sleeps creates slow, flaky tests and does not prove the configured field is correct.
- Claiming that server timeouts stop all handler work overstates their guarantees.
- Treating the normal server-close sentinel as a fatal startup error reports expected shutdown as failure.
- Calling `http.Error` after a failed body write creates a second, mixed response rather than repairing the committed one.

## 16. Topics and References for Study

- Study the official standard-library documentation for `net/http`, especially `Handler`, `HandlerFunc`, `ResponseWriter`, `Request`, `Server`, `ErrServerClosed`, status constants, and header commitment.
- Study `net/http/httptest` for in-memory handler requests and response recording.
- Study `html/template` for contextual autoescaping, trusted templates, and execution errors.
- Review the `errors` package for sentinel comparison and wrapped-cause inspection, and the `time` package for expressing server limits.
- Read the HTTP semantics for `GET`, `HEAD`, `404`, `405`, and the `Allow` header.

## 17. Self-Assessment Questions

1. Why can the handler tree be tested without starting the production server?
2. What exact event commits response headers, and why does its order matter?
3. Why is `html/template` appropriate for the supplied message while string concatenation is not?
4. Why does a known path with an unsupported method return `405`, while an unknown path returns `404`?
5. Why does this project return no body for `HEAD` even when the status is an error?
6. What phase does each of the four server timeouts limit?
7. Why does a maximum header size not limit a request body or a handler's computation time?
8. Why can a write failure after commitment be reported but not converted into a clean replacement response?
9. How do tests distinguish expected server closure from an actual serving failure?

## 18. Definition of Completion

- [ ] Both routes satisfy their exact method, status, content-type, and body contracts.
- [ ] Unknown paths, unsupported methods, the `Allow` header, and the explicit bodyless `HEAD` policy are deterministic and fully tested.
- [ ] The root page is produced by `html/template`, and hostile-looking display data is proven to be escaped rather than executed as markup.
- [ ] Handler construction is independent from network startup, and all route tests use `httptest` without a fixed port.
- [ ] Pre-commit template failure and observable writer failure behavior are tested without attempting a second response.
- [ ] Production uses an `http.Server` with explicit read-header, read, write, and idle timeouts and a positive maximum header size.
- [ ] Documentation and tests describe each limit honestly and use configuration inspection rather than real sleeps.
- [ ] Expected server closure is distinguished from all other startup and serving errors.
- [ ] The completed implementation uses only the standard library and passes its deterministic test suite.
- [ ] You can explain every contract and trade-off without referring to implementation syntax.

## 19. Optional Extensions

- Add a second safely rendered HTML page whose route has its own exact method and error contract.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 014 — Input Validator](../../01-foundations/014_input_validator/README.md#20-prerequisite-based-documentation-guide), [Project 017 — JSON Todo Persister](../../02-data-structures/017_json_todo_persister/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`net/http`](https://pkg.go.dev/net/http), [`net/http/httptest`](https://pkg.go.dev/net/http/httptest), [`html/template`](https://pkg.go.dev/html/template).
- **Standards and concept references:** [RFC 9110: HTTP semantics](https://www.rfc-editor.org/rfc/rfc9110.html).
- **Testing references:** [Go security best practices](https://go.dev/doc/security/best-practices).

### Project-specific learning focus

- **Learn now:** handlers, header commitment, GET and HEAD behavior, 404 and 405 responses, contextual escaping, server timeouts, and handler-level tests.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
