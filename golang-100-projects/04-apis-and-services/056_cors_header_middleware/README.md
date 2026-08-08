# Project 056 — CORS Header Middleware

## 1. Project Name and Number

- Project 056 — `cors_header_middleware`.
- Folder: `04-apis-and-services/056_cors_header_middleware/`.
- README only; the learner writes all source and tests.

## 2. Project Idea

Build a single CORS middleware for a small JSON HTTP API on top of `net/http`. The middleware enforces one exact policy and nothing else. The allowed origins are exactly `https://app.example.com` and `https://admin.example.com`. The allowed methods are exactly `GET`, `POST`, `PUT`, `DELETE`. The allowed request headers are exactly `Content-Type`, `X-Request-ID`, `X-CSRF-Token`. Credentials are enabled. The preflight max age is exactly 600 seconds. A fixed set of security headers is emitted on every response the middleware writes. HSTS is emitted with the exact value `max-age=31536000; includeSubDomains`, only when the inbound `Request.TLS` is non-nil. The middleware is the sole gatekeeper for CORS; the application does not provide any loopback-echo or other insecure handler.

## 3. Why This Project Now?

- Projects 046 through 055 taught how to build handlers, route requests, return JSON envelopes, and authenticate with bearer tokens inside two web frameworks.
- None of those projects addressed what happens when a browser front-end hosted on a different origin calls the API.
- Project 056 is the first project in the plan whose security contract is decided almost entirely by HTTP response headers rather than by the response body.
- Understanding CORS deeply is the prerequisite for the per-client rate limiter in Project 057, which also has to decide who a request is for, and for the OpenAPI contract in Project 058, which has to describe the CORS behaviour of the API.

## 4. Prerequisites

- Required earlier projects: Project 055 and Project 046.
- Earlier HTTP, middleware, and JSON-envelope projects are useful review but are not formally required.
- The learner must already understand `net/http` handlers, the middleware chain pattern from Project 048, request and response header manipulation, the difference between a preflight and a simple request, and how to run a server under `httptest`.

## 5. What You Must Know Before Starting

- The CORS specification at the level of: `Origin`, `Access-Control-Request-Method`, `Access-Control-Request-Headers`, `Access-Control-Allow-Origin`, `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`, `Access-Control-Allow-Credentials`, `Access-Control-Max-Age`, and `Vary`.
- The fact that a preflight is an `OPTIONS` request that carries both `Origin` and `Access-Control-Request-Method`.
- Why `Access-Control-Allow-Origin: *` is incompatible with `Access-Control-Allow-Credentials: true`.
- Why `Vary: Origin` matters for intermediate caches when the allowlist contains more than one origin, and why the middleware must merge `Vary` into any pre-existing tokens rather than overwrite it.
- That browsers send `Origin` only in cross-origin and certain same-origin requests; absence of `Origin` means a same-origin or non-browser request.
- That HTTP method tokens are case-sensitive. `GET` and `get` are not the same token.
- That HTTP header names are case-insensitive.
- How to write tests using `httptest.NewRecorder` and `httptest.NewServer`, and how to assert headers with `http.Header.Get` plus case-insensitive helpers.
- How `net/http` exposes `Request.TLS`, `Request.RemoteAddr`, `Request.Method`, and `Request.Header`, and the `ResponseWriter` interface for setting headers.

## 6. Explanation of New Concepts

### Concepts

- **Simple request versus preflight.** Browsers classify cross-origin requests by HTTP method and by the set of "non-simple" request headers.
- Simple requests are sent straight to the server.
- Anything else triggers a preflight: the browser sends an `OPTIONS` request whose purpose is to ask permission before the real request is sent.
- A server that ignores preflights will appear to work with `curl` but fail inside a real browser.

- **Origin allowlist.** A server picks the exact list of origins it is willing to talk to.
- Anything outside that list is rejected at the middleware layer with no allow headers in the response.
- Reflecting the incoming `Origin` header verbatim into `Access-Control-Allow-Origin` is a security defect, not a convenience, because any origin can put any string there.

- **Credentialed CORS.** When `Access-Control-Allow-Credentials: true` is set, the browser will only honour a non-wildcard `Access-Control-Allow-Origin` and only when that origin value is concrete.
- The middleware echoes the exact configured allowed origin and never a wildcard.

- **`Vary` and cache keying.** Public caches key a stored response on the request headers listed in `Vary`.
- When a server picks `Access-Control-Allow-Origin` per request, the cache must also key on `Origin`, otherwise it can serve one tenant's CORS headers to another tenant's browser.
- The middleware treats `Vary` as a case-insensitive set of tokens.
- Tokens already present from the downstream handler are preserved; tokens added by the middleware are merged in; the resulting header is serialised exactly once.
- The merged header must be finalised before the downstream handler is allowed to commit the response.
- The safe way to enforce this is to wrap the response writer at the middleware boundary, so that any later `WriteHeader` or `Write` call sees the merged `Vary` and cannot drop it.
- Setting `Vary` once before calling the downstream handler and assuming the handler will not overwrite it is unsafe; an explicit response-writer boundary is what makes the preservation observable.

- **Security headers.** Apart from CORS, the middleware writes a fixed set of defensive headers on every response it produces, including the `204` and `403` responses it writes itself: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, and `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`. `Strict-Transport-Security: max-age=31536000; includeSubDomains` is written only when `Request.TLS` is non-nil.
- HSTS is not a configuration toggle in this project.
- Plain-HTTP requests, including local test requests over `httptest`, never receive HSTS.

- **Preflight handling.** A preflight is only a request whose method is `OPTIONS` and that carries both a non-empty `Origin` header and a non-empty `Access-Control-Request-Method` header.
- The middleware validates the origin against the allowlist, validates the requested method against the configured method list using case-sensitive comparison, and validates the requested header set against the configured header list using case-insensitive comparison.
- The requested header set is parsed by splitting on commas, trimming whitespace from each item, and rejecting the whole preflight when any item is empty or any parsed name is outside the allowlist.
- A missing `Access-Control-Request-Headers` header means an empty requested set, which is valid.

## 7. Learning Objective

- After finishing this project, the learner can explain in their own words why a CORS middleware must never reflect arbitrary origins, why credentialed mode forbids wildcards, what `Vary: Origin` is for and why it must be merged through a response-writer boundary, what distinguishes a preflight from an ordinary `OPTIONS`, why HTTP method tokens are case-sensitive while header names are not, and why HSTS is only emitted when the request is actually over TLS.
- The learner can also write a deterministic test suite that pins every behaviour.

## 8. Functional Requirements

1. The configuration holds the exact policy: an allowlist of exactly `https://app.example.com` and `https://admin.example.com`; an allowed-methods list of exactly `GET`, `POST`, `PUT`, `DELETE`; an allowed-headers list of exactly `Content-Type`, `X-Request-ID`, `X-CSRF-Token`; credentials enabled; and a preflight max age of exactly 600 seconds. There are no other configuration fields that change CORS behaviour.
2. The constructor normalises and stores these values, validates them, and rejects any contradictory input. The constructor rejects empty strings, malformed origins, wildcard origins, duplicate normalised entries in any of the three lists, empty tokens in the methods or headers lists, and any configuration that would emit `Access-Control-Allow-Origin: *` together with credentials enabled.
3. On every request the middleware classifies the origin as "disallowed", "no origin", or "allowed (exact match)". Origin comparison is exact and case-sensitive: `https://app.example.com` matches only that string and not `https://app.example.com/`, not `http://app.example.com`, not `https://APP.example.com`, and not any other case variant.
4. The middleware does not classify the actual method of an ordinary request. CORS does not authorise or block ordinary methods; routing does. Method-allowlist validation applies only to the `Access-Control-Request-Method` value on a preflight, where the comparison is case-sensitive and a token such as `get` (lowercase) is not equal to `GET` and is rejected.
5. Header-name comparison is case-insensitive. `content-type` and `Content-Type` are the same header name.
6. A request whose origin is in the disallowed set returns `403 Forbidden` with zero body bytes, no `Access-Control-Allow-*` headers, the fixed security headers including HSTS exactly when `Request.TLS` is non-nil, and the next handler is not called.
7. A request with no `Origin` header passes through to the next handler without any `Access-Control-Allow-*` headers. The fixed security headers are written; HSTS is written exactly when `Request.TLS` is non-nil. The next handler is called once.
8. An allowed-origin request whose method is `OPTIONS` and that carries both `Origin` and `Access-Control-Request-Method` is classified as a preflight. A preflight whose origin is allowed, whose requested method is one of the exact four tokens, and whose requested header set is a subset of the allowed-headers list returns `204 No Content` with zero body bytes, `Access-Control-Allow-Origin: <exact allowed origin>`, `Access-Control-Allow-Credentials: true`, `Access-Control-Allow-Methods: GET, POST, PUT, DELETE`, `Access-Control-Allow-Headers: Content-Type, X-Request-ID, X-CSRF-Token`, `Access-Control-Max-Age: 600`, the merged `Vary` header, and the fixed security headers including HSTS exactly when `Request.TLS` is non-nil. The next handler is not called.
9. A preflight whose origin is allowed but whose requested method is not one of the exact four tokens, or whose requested header set contains any name outside the allowed-headers list, or whose requested-header list contains any empty item, returns `403 Forbidden` with zero body bytes, no `Access-Control-Allow-*` headers, the merged `Vary` header, the fixed security headers including HSTS exactly when `Request.TLS` is non-nil, and the next handler is not called.
10. A preflight with a missing or empty `Origin` header, or with a missing or empty `Access-Control-Request-Method` header, is not a preflight by definition. Such a request is treated according to its origin classification and its method, not as a preflight.
11. An allowed-origin, non-preflight request forwards to the next handler once. The response written by the middleware includes `Access-Control-Allow-Origin: <exact allowed origin>`, `Access-Control-Allow-Credentials: true`, the merged `Vary` header, and the fixed security headers including HSTS exactly when `Request.TLS` is non-nil. The preflight-only fields `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`, and `Access-Control-Max-Age` are not emitted on ordinary responses. The next handler's status and body are preserved.
12. An `OPTIONS` request that does not carry both required preflight headers and for which a downstream handler is registered is forwarded to that handler. The middleware does not synthesise a preflight reply for an `OPTIONS` request that is missing `Origin` or `Access-Control-Request-Method`.
13. The middleware uses a response-writer boundary so that the merged `Vary` header, the CORS allow headers, the security headers, and HSTS exactly when `Request.TLS` is non-nil are written to the wire before the downstream handler is allowed to commit a status or body, and so that downstream code that attempts to overwrite any of these is observed.
14. The merged `Vary` header contains, at minimum, `Origin` for every response with a non-empty `Origin` header, `Access-Control-Request-Method` for every preflight response, and `Access-Control-Request-Headers` for every preflight response. Tokens already present from the downstream handler are preserved; the resulting set is de-duplicated case-insensitively and serialised exactly once. Tokens are separated by `, ` and appear in a stable order that the tests pin.
15. The middleware is safe for concurrent use. Configuration is set at construction time and treated as read-only at request time.

## 9. Inputs and Outputs

### Interface Contract

The middleware consumes an `http.Request` and an `http.ResponseWriter` and runs before the downstream handler. Inputs that matter to its decision are: the request method, the `Origin` header (optional), `Access-Control-Request-Method` (optional), and `Access-Control-Request-Headers` (optional, comma-separated list of names). Outputs are responses written directly for `403` cases and `204` cases, and response header additions on the response writer for allowed non-preflight cases. Example textual inputs and the expected textual outputs:

- Request: `OPTIONS /notes`, `Origin: https://app.example.com`, `Access-Control-Request-Method: GET`, `Access-Control-Request-Headers: Content-Type, X-CSRF-Token`, `Request.TLS` non-nil. Expected response: status `204`, body zero bytes, `Access-Control-Allow-Origin: https://app.example.com`, `Access-Control-Allow-Credentials: true`, `Access-Control-Allow-Methods: GET, POST, PUT, DELETE`, `Access-Control-Allow-Headers: Content-Type, X-Request-ID, X-CSRF-Token`, `Access-Control-Max-Age: 600`, `Vary` includes `Origin`, `Access-Control-Request-Method`, `Access-Control-Request-Headers`, `Strict-Transport-Security: max-age=31536000; includeSubDomains`, plus `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`. Handler is not called.
- Request: `OPTIONS /notes`, `Origin: https://other.example`, `Access-Control-Request-Method: GET`, `Request.TLS` non-nil. Expected response: status `403`, body zero bytes, no `Access-Control-Allow-*` headers, `Vary` includes `Origin`, the fixed security headers including HSTS. Handler is not called.
- Request: `OPTIONS /notes`, `Origin: https://app.example.com`, `Access-Control-Request-Method: PATCH`, `Request.TLS` non-nil. Expected response: status `403`, body zero bytes, no `Access-Control-Allow-*` headers, `Vary` includes `Origin` and `Access-Control-Request-Method`, the fixed security headers including HSTS. Handler is not called.
- Request: `OPTIONS /notes`, `Origin: https://app.example.com`, `Access-Control-Request-Method: GET`, `Access-Control-Request-Headers: Content-Type, , X-CSRF-Token` (note the empty item). Expected response: status `403`, body zero bytes, no `Access-Control-Allow-*` headers, the merged `Vary` header, the fixed security headers including HSTS. Handler is not called.
- Request: `GET /notes`, `Origin: https://app.example.com`, `Request.TLS` non-nil. Expected response: status from downstream handler, body from downstream handler, `Access-Control-Allow-Origin: https://app.example.com`, `Access-Control-Allow-Credentials: true`, `Vary` includes `Origin`, the fixed security headers including HSTS. The preflight-only fields `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`, and `Access-Control-Max-Age` are absent. Handler is called once.
- Request: `GET /health`, no `Origin` header, `Request.TLS` non-nil. Expected response: status from downstream handler, body from downstream handler, no `Access-Control-Allow-*` headers, the fixed security headers including HSTS. Handler is called once.
- Request: `GET /notes`, `Origin: https://app.example.com`, `Request.TLS` nil. Expected response: status from downstream handler, body from downstream handler, the CORS allow headers (`Access-Control-Allow-Origin`, `Access-Control-Allow-Credentials`) and `Vary` includes `Origin`, the fixed security headers except HSTS. HSTS is absent. The preflight-only fields remain absent.
- Request: `OPTIONS /custom`, no `Origin`, no `Access-Control-Request-Method`, where `/custom` has a registered handler. Expected response: status and body from that handler's normal routing, not a preflight reply. The middleware passes the request through because it is not a preflight by definition.
- Request: `OPTIONS /custom`, `Origin: https://app.example.com`, no `Access-Control-Request-Method`, where `/custom` has a registered handler. Expected response: status and body from that handler's normal routing, not a preflight reply. The middleware passes the request through because a preflight requires both `Origin` and `Access-Control-Request-Method`.

## 10. Rules and Edge Cases

- An empty `Origin` header is treated as no origin. The middleware never writes a blank `Access-Control-Allow-Origin`.
- An empty `Access-Control-Request-Method` header is treated as missing. The middleware does not classify the request as a preflight.
- An empty `Access-Control-Request-Headers` header is treated as an empty requested set, which is valid.
- A `,`-separated list with empty items after trimming (for example `Content-Type, , X-CSRF-Token`) is a rejected preflight.
- A requested header name whose lowercase form is outside the allowlist is a rejected preflight.
- Two requests with the same effective `Origin` must each independently reach the allowlist decision. There is no global short-circuit.
- The middleware must not crash on missing `Access-Control-Request-Headers`. Missing means empty requested set.
- A disallowed origin with credentials enabled still returns `403` and still omits allow headers; credentials do not turn a denial into an allowance.
- Responses written by the middleware itself (`204`, `403`) have empty bodies.
- The fixed security headers are written exactly once per response, with the exact values listed.
- HSTS is written exactly when `Request.TLS` is non-nil. The header value is exactly `max-age=31536000; includeSubDomains` and nothing else.
- The `Vary` header serialised on the wire is the result of a case-insensitive union of pre-existing tokens and the tokens the middleware adds. A token that is already present is not duplicated.

## 11. Project Constraints

- Standard library only. No third-party CORS libraries, no Gin, no Fiber, no middleware frameworks.
- The middleware is the only CORS-aware code in the project. There is no loopback-echo handler, no "developer-mode" bypass, no insecure fallback.
- No logging is required and none should be added inside the middleware.
- `Strict-Transport-Security` is not a configuration toggle. It is emitted exactly when `Request.TLS` is non-nil, with the exact value `max-age=31536000; includeSubDomains`.
- The middleware must not modify the request body or the response body written by the downstream handler. It may add response headers and may short-circuit with its own response.
- Tests use `httptest` and run locally. No real network is required.

## 12. Design Questions Before Coding

- Where is the configuration stored? As instance fields on the middleware value or as package-level constants? The configuration is fixed by the project; storing it as instance fields set by the constructor makes tests and documentation clearer.
- How is the response-writer boundary implemented? The middleware wraps the `ResponseWriter` so that any header mutations performed by the downstream handler are observed and the merged `Vary` and the fixed security headers cannot be silently overwritten.
- How is the merged `Vary` represented internally before serialisation? As a `map[string]struct{}` keyed by lowercase tokens, so the union is case-insensitive and the serialisation order is stable.
- How is the preflight method allowlist enforced? The middleware compares the `Access-Control-Request-Method` value against the four exact tokens using case-sensitive equality. The actual method of an ordinary request is not validated against the allowlist; routing decides whether an ordinary `GET`, `POST`, `PATCH`, or any other method is supported.
- How are the three exact header-name tokens stored? As a `map[string]struct{}` keyed by the lowercase form. The comparison is case-insensitive.
- How is origin matching done? An exact, case-sensitive equality check against the configured list.
- How is HSTS emitted? The middleware reads `Request.TLS` once at the start of each request and writes HSTS only when the field is non-nil.
- How is the next handler invocation counted in tests? Through a counter the test handler increments. The middleware must not double-call and must not call after a `403` or `204` decision.
- How are the pre-existing `Vary` tokens of the downstream handler observed? Through the response-writer wrapper. The wrapper records every header written through it and merges them into the final `Vary`.

## 13. Implementation Milestones

1. Sketch the configuration struct with the exact fields, exact values, and the constructor validation rules on paper.
2. Implement the constructor that normalises and validates the allowlist, methods, and headers, and rejects every invalid case listed in the rules.
3. Implement the origin, method, and preflight-requested-header classifiers together: an exact-match origin classifier returning "disallowed", "no origin", or "allowed (exact value)"; an exact-token method classifier returning true only for `GET`, `POST`, `PUT`, `DELETE`; a case-insensitive header-name set check using the lowercase form; and a requested-headers parser that splits on commas, trims whitespace, rejects empty items, lowercases the names, and decides whether the parsed set is a subset of the allowlist.
4. Implement the preflight detector: a function that returns true only when the method is `OPTIONS`, `Origin` is non-empty, and `Access-Control-Request-Method` is non-empty.
5. Implement the response-writer boundary used to observe downstream header writes and to guarantee the merged `Vary` and the fixed security headers reach the wire.
6. Centralise the fixed security headers and the HSTS-on-TLS rule so each branch calls the same code.
7. Implement the preflight writer for the success path: `204`, zero body, exact allow headers, exact security headers, merged `Vary`. Do not call the next handler.
8. Implement the preflight writer for the failure path and the disallowed-origin writer together: both return `403`, zero body, no allow headers, merged `Vary`, exact security headers. Do not call the next handler.
9. Implement the allowed-origin non-preflight pass-through: wrap the response writer, write the ordinary-response allow headers (`Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials`) and the security headers into the wrapper, then call the next handler exactly once. Do not emit the preflight-only fields on this path.
10. Implement the no-origin pass-through: wrap the response writer, write only the security headers into the wrapper, then call the next handler exactly once.
11. Wire the middleware around a no-op handler in tests and verify the call count for the preflight, allowed non-preflight, no-origin, and disallowed-origin paths is observable and matches the rules.
12. Review the verification list and confirm every item is covered before declaring the project complete.

## 14. Verification Cases the Learner Must Write

### Required Cases

Each item is a behavioural specification. The learner writes the corresponding `go test` code.

- Allowed origin, allowed method, valid preflight: response is `204`, body has zero bytes, `Access-Control-Allow-Origin` equals the exact allowed origin, `Access-Control-Allow-Credentials: true`, `Access-Control-Allow-Methods: GET, POST, PUT, DELETE`, `Access-Control-Allow-Headers: Content-Type, X-Request-ID, X-CSRF-Token`, `Access-Control-Max-Age: 600`, `Vary` includes `Origin`, `Access-Control-Request-Method`, `Access-Control-Request-Headers`, the fixed security headers are present, HSTS is present when `Request.TLS` is non-nil, handler is not called.
- Disallowed origin preflight: response is `403`, body has zero bytes, no `Access-Control-Allow-*` headers, `Vary` includes `Origin`, the fixed security headers are present, handler is not called.
- Disallowed origin ordinary request: same `403` and header behaviour as the disallowed preflight; handler is not called.
- Allowed origin, requested method `PATCH`: preflight is `403`, body zero bytes, no allow headers, `Vary` includes `Origin` and `Access-Control-Request-Method`, fixed security headers, handler is not called.
- Allowed origin, requested method `get` (lowercase): preflight is `403` because the requested method is not an exact case-sensitive match. Body zero bytes, no allow headers, fixed security headers, handler is not called.
- Allowed origin, requested header `X-Other`: preflight is `403`, body zero bytes, no allow headers, fixed security headers, handler is not called.
- Allowed origin, requested headers `Content-Type, , X-CSRF-Token` with an empty item: preflight is `403` because the requested list contains an empty item. Body zero bytes, no allow headers, fixed security headers, handler is not called.
- Allowed origin, requested headers `Content-Type, x-csrf-token` (mixed case): preflight is `204` because header-name comparison is case-insensitive and `x-csrf-token` matches `X-CSRF-Token`. Allow headers are present, fixed security headers present, handler not called.
- Allowed origin, non-preflight `GET`: downstream handler runs exactly once, `Access-Control-Allow-Origin` equals the exact allowed origin, `Access-Control-Allow-Credentials: true`, `Vary` includes `Origin`, fixed security headers present, handler's status and body preserved. The preflight-only fields `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`, and `Access-Control-Max-Age` are absent on the response.
- No `Origin` header on a `GET`: downstream handler runs exactly once, no `Access-Control-Allow-*` headers, fixed security headers present, downstream response intact.
- Ordinary `OPTIONS` with no preflight headers and a registered handler: downstream handler runs, response comes from that handler, not from the preflight branch.
- Ordinary `OPTIONS` with `Origin` but no `Access-Control-Request-Method` and a registered handler: downstream handler runs, response comes from that handler.
- Exact origin match required: `https://app.example.com` matches only that string; `https://app.example.com/` does not match; `https://app.example.com:443` does not match; `http://app.example.com` does not match; `https://APP.example.com` does not match.
- HSTS TLS-only behaviour: a test with `Request.TLS` non-nil emits `Strict-Transport-Security: max-age=31536000; includeSubDomains`; a test with `Request.TLS` nil does not emit HSTS; the header value is exactly `max-age=31536000; includeSubDomains` and nothing else.
- `Vary` merging: a downstream handler that already sets `Vary: Accept-Encoding` results in a response whose `Vary` contains both `Accept-Encoding` and `Origin`. The merge is case-insensitive and a token present twice is serialised once.
- Downstream `Vary` overwrite attempt: a downstream handler that calls `Header().Set("Vary", "X-Custom")` results in a response whose `Vary` still contains the CORS tokens; the wrapper preserves the merged value.
- Constructor rejects empty or malformed origins: the constructor returns an error.
- Constructor rejects wildcard origins: the constructor returns an error.
- Constructor rejects duplicate normalised entries in any list: the constructor returns an error.
- Constructor rejects empty method or header tokens: the constructor returns an error.
- Constructor rejects contradictory credential settings: the constructor returns an error.
- Next-call count: a valid preflight causes the next handler to be called zero times; a simple `GET` from an allowed origin causes the next handler to be called exactly once; a disallowed origin request causes the next handler to be called zero times.
- Concurrency: a test that fires many goroutines against the middleware with a frozen configuration must not race. The middleware runs under `go test -race` cleanly.

## 15. Common Mistakes to Watch For

- Reflecting the incoming `Origin` header verbatim into `Access-Control-Allow-Origin`. Any origin can put any string there.
- Setting `Access-Control-Allow-Origin: *` together with `Access-Control-Allow-Credentials: true`. The browser will reject this combination.
- Treating HTTP method tokens as case-insensitive. `get` is not `GET`.
- Treating HTTP header names as case-sensitive. `content-type` is `Content-Type`.
- Splitting requested headers without trimming whitespace and without rejecting empty items.
- Overwriting an existing `Vary` header instead of merging it. The cache will then serve one tenant's CORS headers to another tenant's browser.
- Treating every `OPTIONS` as a preflight. That breaks endpoints that legitimately use `OPTIONS` for non-CORS purposes.
- Emitting HSTS over plain HTTP. HSTS has no useful effect on plain HTTP and pollutes the response.
- Trusting `Origin` for authentication decisions. `Origin` is easy to forge from a non-browser client and must drive CORS only, never authorisation.
- Running tests with shared, mutated configuration across goroutines. Even if the middleware is read-only at request time, the configuration struct must be immutable after construction.
- Setting `Vary` once before calling the downstream handler and assuming the handler cannot overwrite it. The merge must be enforced through a response-writer boundary.
- Allowing the downstream handler to drop the merged `Vary` or the fixed security headers. The response-writer boundary is what makes this observable.

## 16. Topics and References for Study

- MDN Web Docs, "Cross-Origin Resource Sharing (CORS)" and the page "CORS preflight request".
- The Fetch Living Standard, sections on CORS and preflight.
- OWASP Cheat Sheet, "Cross-Site Request Forgery Prevention" and "REST Security".
- Go `net/http` package documentation, especially `ResponseWriter.Header()`, `Request.Header`, `Request.Method`, `Request.TLS`, and `httptest`.
- RFC 6797 for `Strict-Transport-Security` and the rule that HSTS is meaningful only over HTTPS.
- The HTTP `Vary` header, MDN reference page, for the cache-keying contract.

## 17. Self-Assessment Questions

1. Why does the spec forbid `Access-Control-Allow-Origin: *` together with `Access-Control-Allow-Credentials: true`? What would happen if the server emitted both?
2. Why must `Vary: Origin` be merged into any existing `Vary` rather than overwriting it? Walk through the cache poisoning scenario.
3. What distinguishes a preflight from an ordinary `OPTIONS` request? How does the middleware decide which one it is?
4. Why are HTTP method tokens case-sensitive while HTTP header names are case-insensitive? Give one example for each.
5. Why is `Origin` not a safe input for authorisation decisions? Give an example of an attacker who controls the header.
6. Why is HSTS only emitted when `Request.TLS` is non-nil? What would the wire payload look like if it were emitted on plain HTTP, and why is that wrong?
7. Why is a response-writer boundary required to preserve `Vary` against a downstream handler that tries to overwrite it? What would happen without one?

## 18. Definition of Completion

The project is complete when, in addition to the rules above:

- [ ] Every item in the verification list is a passing test that the learner wrote themselves.
- [ ] The tests pass under `go test -race ./...` from the project folder.
- [ ] The middleware contains no third-party imports.
- [ ] The configuration struct has the exact fixed values and a constructor that rejects every invalid input listed in the rules.
- [ ] The middleware never emits HSTS on a request with `Request.TLS` nil.
- [ ] The middleware never emits `Access-Control-Allow-Origin: *` with credentials enabled.
- [ ] The `Vary` header serialised on the wire is the case-insensitive union of pre-existing tokens and the tokens the middleware adds, with no duplicates and a stable order that the tests pin.
- [ ] The learner can answer every self-assessment question without rereading the README.

## 19. Optional Extensions

At most two. Pick one only if the core project is already complete and tested. Optional extensions must not add speculative configuration fields, wildcard fallbacks, or insecure handlers.

- Add a tiny per-response decision struct (`Origin`, `Decision`, `AllowedHeaders`) that the application may read after `next` returns, useful for future structured logging.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 055 — Fiber Framework CRUD](../../04-apis-and-services/055_fiber_framework_crud/README.md#20-prerequisite-based-documentation-guide), [Project 046 — Basic HTTP Server](../../04-apis-and-services/046_basic_http_server/README.md#20-prerequisite-based-documentation-guide), [Project 034 — Worker Pool Basic](../../03-concurrency/034_worker_pool_basic/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **Standards and concept references:** [Fetch Standard: CORS protocol](https://fetch.spec.whatwg.org/#http-cors-protocol), [RFC 6797: HSTS](https://www.rfc-editor.org/rfc/rfc6797.html), [OWASP CSRF guidance](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html).

### Project-specific learning focus

- **Learn now:** simple versus preflight requests, exact-origin allowlists, credential rules, Vary cache keys, safe headers, HTTPS-only HSTS, and denial behavior.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
