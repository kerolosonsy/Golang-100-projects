# Project 048 — Custom Router and Middleware

## 1. Project Name and Number

- Project 048 — Custom Router and Middleware, located in `048_custom_router_middleware`.

## 2. Project Idea

Build a deliberately small HTTP router on top of `net/http`. It recognizes exact static paths and paths containing named parameters that each consume one URL segment, chooses a deterministic route, and places decoded parameters in request context. A fixed middleware chain adds a conservative request ID and one structured access-log record without using globals or a full web framework.

## 3. Why This Project Now?

- Project 047 established precise REST paths, method status contracts, JSON boundary validation, and concurrent handler use.
- This project isolates the routing and cross-cutting concerns that a larger API needs before Project 049 standardizes response formatting.
- The learner must make ambiguity, path normalization, context ownership, and middleware order explicit instead of inheriting accidental behavior from a framework.

## 4. Prerequisites

- Complete Projects 047 and 046 before starting.
- Earlier projects may be useful review, but they are not required prerequisites.

- You should already be comfortable with `net/http` handlers, `httptest`, request context propagation, mutex-free immutable configuration after setup, HTTP methods, response headers, and deterministic concurrent tests.

## 5. What You Must Know Before Starting

- A route pattern is a path contract, not a filesystem path. It must not silently clean, collapse, or reinterpret a request into a different resource.
- A static segment and a named parameter segment are different match classes. Static specificity must win independently of registration order.
- A named parameter occupies one complete segment. It is not a wildcard, a suffix matcher, or permission to consume a slash.
- URL path escaping is distinct from query escaping. A plus sign in a path segment is not automatically a space.
- A path can contain malformed percent escapes or an encoded separator. Decoding policy must be stated before matching.
- `context.Context` carries request-scoped values across handler boundaries; it is not a global variable store and should not hold mutable application state.
- Middleware wraps a downstream handler. Code before the downstream call runs on the way in, and code after it runs while wrappers unwind.
- A response recorder observes status and body in tests, but a logging wrapper may need to observe an implicit status and byte count without changing handler semantics.
- Concurrent serving is different from concurrent registration. This project freezes route configuration before serving requests.

## 6. Explanation of New Concepts

### Concepts

- A static route matches a complete path whose segments are literal.
- A parameter route uses a named segment such as `{noteID}` as notation for one captured segment.
- The braces are route-pattern syntax, not a catch-all.
- A request must have the same number of segments as the pattern, and every parameter value must be non-empty after the prescribed decoding checks.

- Route selection happens in two conceptual stages.
- First, the path shape is selected using exact static specificity and the conflict rules.
- Then the request method is looked up on that selected shape.
- This prevents a parameter route from bypassing a more specific static route merely because it supports a different method.
- A selected path with no matching method is known and returns `405`, not `404`.

- Registration conflicts protect the routing table from ambiguous behavior.
- The same method and equivalent pattern cannot be registered twice.
- Parameter names are part of the context contract, so equivalent shapes with inconsistent names are rejected.
- Distinct parameter patterns that could match the same concrete path without a defined precedence are also rejected.
- A static route and a parameter route may coexist because static precedence resolves their overlap.

- URL decoding is a security boundary.
- The router examines path segments without treating an encoded slash or backslash as an ordinary character that can change segment count.
- It decodes each accepted parameter exactly once, rejects malformed escapes, and rejects separator ambiguity rather than cleaning it into a new path.

- Request-ID middleware gives one request a correlation value.
- A valid incoming value can be reused; an absent or invalid value is replaced by one generated through an injected source.
- Access logging runs after routing so it can record the selected status and byte count, and it receives the request ID through the request context.

- The required chain is ordered from outside to inside as request-ID middleware, structured access logging, and route dispatch.
- The request-ID pre-processing occurs first and its post-processing occurs last.
- The access logger's completion record is emitted after the endpoint returns, so its status and byte count describe the completed response.
- Every middleware invokes its downstream handler at most once.

## 7. Learning Objective

- By completion, you can design a bounded route grammar, reject ambiguous registrations, enforce static-over-parameter precedence, decode parameters safely once, and explain why path cleaning can be a security bug.
- You can also build and test an ordered middleware chain that propagates request-scoped identity, emits deterministic structured access logs, and remains safe under concurrent requests.

## 8. Functional Requirements

1. Support exact static path patterns and patterns containing one or more named parameter segments. Each parameter occupies one complete segment and captures one value; wildcard, suffix, and catch-all patterns are unsupported.
2. Require patterns to use a canonical absolute-path form. The root path is allowed; non-root patterns cannot end in a trailing slash, contain empty segments, or contain dot segments.
3. Match path shapes exactly by segment count. A static segment matches only its registered literal, and a parameter segment matches one non-empty segment.
4. Give a static route precedence over a parameter route for the same concrete path, regardless of registration order.
5. Reject duplicate method-pattern registrations and every other ambiguous or conflicting registration defined by the route grammar before serving begins. Do not panic or silently replace a previous handler.
6. Select the path shape before method matching. A known selected path with a missing method returns `405 Method Not Allowed`; an unknown path returns `404 Not Found`.
7. Return a sorted `Allow` header for `405`. Sort method tokens in ascending bytewise order and join them with comma-space separators. Static and parameter route method sets must not be mixed after static precedence selects a path.
8. Do not imply `HEAD` from `GET`. `HEAD` is available only when explicitly registered, and a handler or router policy must preserve a bodyless `HEAD` response.
9. Decode each captured parameter exactly once using URL path semantics. Reject malformed percent escapes, decoded slash or backslash separators, and encoded-separator ambiguity such as a value that could become a separator under another decode.
10. Never redirect or clean a request path. A syntactically valid but noncanonical path containing a non-root trailing slash, repeated slash, raw dot segment, or raw backslash is treated as a distinct unmatched resource and returns `404`. A malformed percent escape, encoded slash or backslash, double-encoded separator ambiguity, or invalid decoded parameter returns `400`.
11. Place decoded route parameters in request context under a private, typed key owned by the router. No global parameter map is allowed, and one request's values must never be visible to another request.
12. Use a request-ID middleware as the outermost required middleware. Accept an incoming ID only when there is exactly one header value, its length is 16 through 64 ASCII characters, and every character is alphanumeric, hyphen, or underscore. Otherwise generate an ID through an injected source.
13. Generated IDs must satisfy the same format. An invalid or failed ID source produces `500 Internal Server Error`, plain-text content type, exact body `internal server error` followed by one line feed, and no `X-Request-ID` header. It does not call the access logger, router, or endpoint, and it reports the configuration failure once through a separate injected error boundary.
14. Echo the selected request ID in the `X-Request-ID` response header and make it available to the access logger and endpoint through request-scoped state.
15. Use structured access logging as the inner required middleware. Emit exactly one completion record for every request that reaches it, with request ID, method, path, status, and response byte count. A request-ID source failure is handled by the outer middleware before this logger and instead reaches the separate injected error boundary exactly once. Do not log query values or arbitrary request headers.
16. Preserve the middleware order: request ID before access logging before route dispatch. Each middleware calls its downstream handler at most once, and completion behavior unwinds in the reverse order.
17. Make the frozen router and required middleware safe for concurrent requests. Route registration is complete before serving and is not mutated concurrently with request handling.

## 9. Inputs and Outputs

### Interface Contract

- Route-registration input consists of a canonical pattern, one HTTP method, and one endpoint handler.
- A pattern may be a static path or contain complete named parameter segments.
- Registration either succeeds or returns a descriptive conflict error before serving.

- Request input is an HTTP method and escaped request path.
- A path query is not part of route matching and is not included in the access-log path field.
- A valid incoming `X-Request-ID` is reused only when it satisfies the full validation rule; all other incoming values are treated as absent for generation purposes.

- For normal routing errors, the router returns plain text with content type exactly `text/plain; charset=utf-8`.
- The `404` body is lowercase `not found` followed by one line feed.
- The `405` body is lowercase `method not allowed` followed by one line feed and includes the sorted `Allow` header.
- A malformed or ambiguous path returns `400` with lowercase `bad request` followed by one line feed. `HEAD` responses contain no body while retaining the selected status and relevant headers.

- The response always includes exactly one `X-Request-ID` value when ID middleware successfully establishes an ID.
- The access-log completion record contains the same value, the original method, the path without query text, the final status, and the number of response bytes observed by the logging boundary.
- The endpoint remains responsible for its own application response format.

- Text-only precedence example: if `/users/me` is registered as a static `GET` route and `/users/{id}` is registered as a parameter `GET` route, a `GET` for `/users/me` selects the static route and supplies no `id` parameter.
- A request for `/users/42` selects the parameter route and supplies `id` with value `42`.

- Text-only decoding example: a parameter containing a percent-encoded space is decoded once to a space.
- A parameter containing an encoded slash or backslash is rejected instead of becoming an additional path segment.
- A plus sign remains a plus sign in a path parameter.

## 10. Rules and Edge Cases

- Patterns use canonical unescaped static literals.
- Patterns and requests are case-sensitive.
- Static literals are not silently normalized, folded to lower case, or given alternate percent-encoded spellings; parameter segments follow the one-time decoding policy below.
- A non-root trailing slash, repeated slash, raw dot segment, or raw backslash is never redirected or cleaned.
- It remains a distinct syntactically valid unmatched path and returns `404`.

- Malformed percent escapes return `400`.
- An encoded slash or backslash in a parameter is rejected with `400`, as is a representation that still contains an encoded separator capable of changing meaning after a second decode.
- The router performs no second decode.
- Parameter values that decode to an empty string, dot segment, or dot-dot segment also return `400`.

- A route with a static segment has higher precedence than a parameter route at the same position.
- If a static path exists but does not support the requested method, the result is `405` using the static route's method set; the router does not fall through to a parameter route that happens to support that method.
- Ambiguous parameter patterns are registration errors rather than runtime lottery.

- An incoming request ID with multiple header values, a comma-separated list, whitespace, invalid characters, or an out-of-range length is not trusted.
- The middleware generates exactly one replacement through its injected source.
- A generated invalid value is a configuration failure for that request, not an opportunity to echo unvalidated data.

- The endpoint may write no body and may leave status implicit; the logging boundary records the resulting HTTP status according to `net/http` semantics.
- If an endpoint changes response headers after commitment, the middleware cannot retroactively repair them.
- The request-ID middleware owns its response header, and endpoint code must not replace it.

- The router has no catch-all route.
- A path with a different segment count, an absent static literal, or no registered parameter shape is `404`.
- Registration errors are reported before serving and do not alter an already frozen route table.

## 11. Project Constraints

- Use only the Go standard library, including `net/http`, `context`, `net/url`, `sort`, `strings`, `log/slog` or another standard structured logging facility, `sync`, and `net/http/httptest` as appropriate.
- Do not use a full router framework, reflection, wildcard or catch-all matching, package-level mutable state, hidden default loggers, or external request-ID packages.

- Do not silently clean paths, redirect trailing slashes, decode a parameter twice, or accept encoded separators.
- Do not prescribe or copy a framework-style API in the guide; the learner chooses types and boundaries.
- Registration must complete before concurrent serving, and tests use in-memory requests with no fixed ports, network calls, or sleeps.

## 12. Design Questions Before Coding

- What route-pattern grammar is small enough to validate completely and still supports the required static and parameter cases?
- How will equivalent parameter shapes be identified when parameter names differ, and which overlaps are safe to reject at registration?
- Does path selection occur before method selection, and what exact `Allow` set follows from that choice?
- Which representation of the escaped request target lets the router detect encoded separators before a decoded value changes segment structure?
- What qualifies as one decoding operation, and how will double-decoding ambiguity be rejected?
- How will route parameters be scoped to one request without exposing a mutable global map?
- Which middleware is outermost, what does each wrapper do before and after its downstream call, and what order should the trace show?
- How will the access logger observe implicit status and byte count without claiming optional `ResponseWriter` capabilities it does not implement?
- What happens when an incoming ID is invalid or the injected generator returns an invalid value?
- How will the frozen route table and middleware dependencies remain safe while many requests execute concurrently?

## 13. Implementation Milestones

1. Specify the canonical pattern grammar, normalization rejection policy, and exact error responses.
2. Build registration validation for methods, names, duplicates, and ambiguous overlaps before enabling serving.
3. Add exact static matching and verify it independently from method selection.
4. Add one-segment parameter matching, one-time decoding, and private context propagation.
5. Add static-over-parameter precedence and sorted `Allow` behavior for known paths.
6. Add explicit `404`, `405`, `400`, trailing-slash, and `HEAD` policies.
7. Add request-ID validation, injected generation, response echoing, and failure handling.
8. Add the access-log wrapper with deterministic fields and reverse-order completion behavior.
9. Freeze configuration before serving and exercise concurrent requests under the race detector.
10. Complete the full `httptest` verification matrix without sleeps or network dependencies.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Register a static route and a parameter route that overlap, then verify the static route wins regardless of registration order.
- Match several parameter values and verify each named value is available to the endpoint through request-scoped context only.
- Verify a wrong segment count, unknown static segment, and syntactically valid unregistered path return `404`.
- Verify malformed percent escapes, encoded slash, encoded backslash, double-encoded separator ambiguity, and parameters that decode to empty or dot segments return `400`. Verify non-root trailing slashes, repeated slashes, raw dot segments, and raw backslashes remain distinct unmatched paths and return `404`. No case redirects or cleans into a registered resource.
- Verify a percent-encoded space decodes once and a plus sign remains a plus sign.
- Attempt duplicate method-pattern registration, equivalent parameter patterns with different names, and overlapping ambiguous parameter patterns. Verify each is rejected before the table is served.
- Register several methods on one path and verify wrong-method responses return `405` with bytewise-sorted `Allow` values and the exact error body.
- Verify static-path method selection does not fall through to a parameter route with another method.
- Verify implicit `HEAD` is not accepted, explicit `HEAD` follows the documented body policy, and unknown `HEAD` remains `404`.
- Supply a valid incoming request ID and verify it is echoed, placed in context, logged, and not replaced by the generator.
- Supply no ID, an invalid ID, multiple values, a comma-separated value, too-short and too-long values, and invalid characters. Verify generation occurs exactly when required and the endpoint sees only the selected ID.
- Make the injected ID source return an invalid value or failure and verify the exact plain-text `500` response, absence of `X-Request-ID`, no access-log or endpoint call, and exactly one report to the separate injected error boundary.
- Use an injected logger or output to verify one completion record per request, the exact field set, matching request ID, final status, byte count, and no query or arbitrary header leakage.
- Record middleware events and verify request-ID pre-processing precedes logging and dispatch, while completion logging and request-ID unwinding occur in reverse order. Verify no middleware calls its downstream more than once.
- Send many concurrent requests through static and parameter routes under the race detector and verify no cross-request parameter or request-ID contamination.
- Verify all router tests use `httptest` and no test sleeps or opens a fixed or external network endpoint.

## 15. Common Mistakes to Watch For

- Using first-registered wins makes behavior depend on setup order and lets a parameter route shadow a more specific static resource.
- Falling through after a static method mismatch can expose a different parameter endpoint.
- Replacing a duplicate registration silently hides configuration errors.
- Treating different parameter names as interchangeable can change what handlers find in context.

- Calling path cleaning functions can collapse `..`, repeated slashes, or encoded separators into a resource the client did not actually request.
- Splitting after an early decode lets `%2F` change segment count.
- Applying query decoding to path data turns plus signs into spaces.
- Decoding a value twice creates traversal and identity ambiguity.

- Using a package-level parameter map or request-ID variable causes concurrent requests to overwrite one another.
- Storing mutable maps in context without ownership rules lets downstream code create races.
- Trusting any incoming request ID enables log injection and unbounded correlation values.
- Echoing an invalid generator result repeats the trust failure.

- Putting the access logger outside request-ID middleware prevents it from reliably seeing the established ID.
- Logging only before dispatch misses the final status and byte count.
- Calling downstream twice can duplicate side effects.
- A response wrapper that advertises unsupported optional interfaces can break handlers that inspect them.

- Testing only successful matches misses registration conflicts, `405` semantics, encoded separators, and middleware unwinding.
- Using sleeps to arrange concurrent requests produces timing-dependent tests instead of a synchronization boundary.
- Treating a race-detector pass as proof of correct route precedence confuses memory safety with behavior.

## 16. Topics and References for Study

- Study the official `net/http` handler and response-writer documentation, `net/http/httptest`, `context`, `net/url` path escaping and unescaping, `sort`, and the standard `log/slog` package.
- Read HTTP semantics for path identity, `404`, `405`, `Allow`, and `HEAD`.
- Search for route specificity, ambiguous route detection, request-scoped context values, response-writer instrumentation, structured access logging, and safe correlation IDs.
- Compare the handler boundary from Project 046 and the REST method matrix from Project 047.

## 17. Self-Assessment Questions

1. Why must static precedence be independent of registration order?
2. Why should an ambiguous parameter overlap be rejected at registration instead of resolved at request time?
3. What is the difference between a path segment and a query component when decoding a plus sign?
4. Why is encoded slash rejection necessary even when the parameter is decoded only once?
5. Why is a selected path with a wrong method `405` rather than `404`?
6. Why should route parameters live in a private typed context key instead of a global map?
7. What does the outer-to-inner middleware order imply about before and after events?
8. Why must an invalid incoming request ID be replaced rather than echoed?
9. What can a logging response wrapper observe, and what must it avoid pretending to support?
10. What does concurrent serving require after the route table is frozen?

## 18. Definition of Completion

- [ ] The router accepts only its documented static and one-segment parameter grammar and rejects duplicate or ambiguous registrations before serving.
- [ ] Static routes beat parameter routes, path selection precedes method selection, unknown paths return `404`, and known wrong methods return `405` with sorted `Allow` values.
- [ ] Syntactically valid noncanonical paths return `404`, malformed or encoded-separator paths return `400`, and no path is redirected or silently cleaned into another resource.
- [ ] Parameters are decoded exactly once, encoded separator ambiguity is rejected, and values are scoped under a private typed context key.
- [ ] Request-ID middleware validates the pinned format, generates through an injected source when needed, echoes exactly one value, and prevents invalid source output from reaching handlers.
- [ ] Structured access logging is injected, emits one complete record with the required fields, and follows the pinned middleware order and unwinding behavior.
- [ ] Every middleware calls its downstream at most once, no global state is used, and concurrent requests pass under the race detector.
- [ ] Tests cover precedence, parameters, decoding, registration errors, `404`, `405`, `Allow`, ID paths, order, concurrency, and no-sleep/no-network constraints.
- [ ] The implementation uses only the standard library and the learner can justify every normalization and trust-boundary decision.

## 19. Optional Extensions

- Add a bounded route-group facility for a fixed prefix while preserving exact matching, static precedence, and the same conflict rules.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 047 — REST API CRUD](../../04-apis-and-services/047_rest_api_crud/README.md#20-prerequisite-based-documentation-guide), [Project 046 — Basic HTTP Server](../../04-apis-and-services/046_basic_http_server/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`net/url`](https://pkg.go.dev/net/url), [`log/slog`](https://pkg.go.dev/log/slog).

### Project-specific learning focus

- **Learn now:** route specificity, ambiguity detection, escaped paths, 404 versus 405, middleware composition, request-scoped values, response capture, and safe logs.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
