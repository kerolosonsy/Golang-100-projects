# Project 091 — API Gateway Service

## 1. Project Name and Number
Project 091, `091_api_gateway_service`. The folder name is fixed by the curriculum table; do not rename the directory.

## 2. Project Idea
A small HTTP gateway that sits in front of a fixed set of upstream services. A validated configuration maps URL path prefixes to fixed upstream origins. The gateway composes, in one `net/http` handler chain, request and correlation identifiers, panic recovery, access logging, route selection, an unauthenticated health response for the public prefix, JWT authentication for protected prefixes, a per-client rate limiter, and a reverse proxy to the chosen upstream. There is no service discovery and no open proxy. The client never chooses the target; the route is fully decided by the gateway from the request path and the validated configuration.

## 3. Why This Project Now?
This project is the composition capstone for the HTTP, auth, limits, and proxy work that came before. It pulls the wiring discipline, the header policy, the timeouts, and the per-route behavior into a single program. The previous projects taught each piece in isolation; this project is the lesson about how the pieces coexist, in what order, and with what error ownership. The formal prerequisites — project 050 (JWT authentication server), project 057 (rate-limited API), project 060 (graceful web shutdown), and project 078 (fixed-upstream reverse proxy) — each contribute one ingredient that this gateway combines. The immediate catalog predecessor, project 090, is optional context rather than a formal prerequisite.

## 4. Prerequisites
The formal prerequisites are projects 050, 057, 060, and 078; project 090 is the immediate catalog predecessor and remains useful as optional context rather than a formal prerequisite.

## 5. What You Must Know Before Starting
- That `net/http` middleware composition is a typed chain where each middleware takes the next `http.Handler` and returns a new `http.Handler`, and the order matters for both short-circuiting and observability.
- That a `ReverseProxy` from `net/http/httputil` will, by default, forward almost every header and writes the upstream response body directly to the client; both behaviors must be shaped deliberately.
- That JWT verification is verification of signature, expiry, issuer, and audience; a missing or invalid token is a 401 from the gateway, never an upstream failure.
- That the rate limit key is derived from a trusted signal and must not be attacker-controllable.
- That panic recovery belongs at the outermost layer of the chain and must still produce a clean JSON response when the panic was raised by a write failure.
- That hop-by-hop headers defined by RFC 7230 section 6.1 must not be forwarded to an upstream nor echoed back to the client.
- That graceful shutdown is `Server.Shutdown` waiting for in-flight handlers and refusing new connections; in-flight upstream calls must end on their own deadline.

## 6. Explanation of New Concepts
- Composition as a typed chain: each middleware takes the next handler and returns a new handler. Order becomes a property of the wiring, not of any individual handler, and so it is part of the gateway's contract.
- Segment-aware normalized prefix matching: configuration entries are validated before matching. A valid prefix has one leading slash, no trailing slash unless the prefix is exactly the root, no repeated slash, and no dot, dot-dot, or encoded-slash segment. Invalid spellings are rejected rather than silently rewritten. A prefix matches the path equal to it or a path that begins with the prefix followed by a slash. So `/api` matches `/api` and `/api/...`, but not `/apix` or `/apifoo`. Two entries with the exact same valid prefix are an invalid configuration. Two entries that are disjoint at the same length are valid. A longer prefix nested inside a shorter prefix is valid and the longer one wins for paths it matches.
- Explicit route selection as a stage: route selection is its own stage in the chain. It runs before authentication and rate limiting so that protected-route and public-route decisions are made before identity and limit decisions, and so that unmatched paths short-circuit to a 404 without ever reaching the auth or limit stages.
- Pinning the full middleware order: the order is recovery, request and correlation identifier handling, route selection, access logging, the health and unmatched short circuit, protected-route authentication, rate limiting, then the proxy. Access logging still records every outcome exactly once, even when a later middleware short-circuits.
- Header names and identifier validation: the request identifier header is exactly `X-Request-ID` and the correlation identifier header is exactly `X-Correlation-ID`. Both are echoed to the client and forwarded to the selected backend. An accepted inbound value is 1 through 64 characters drawn only from ASCII letters, digits, period, underscore, colon, and hyphen. Empty, overlong, repeated, or otherwise invalid inbound values are replaced by generated values. The same safe values appear in the response, log, and upstream request. Neither identifier is an authentication identity or a rate-limit key.
- JSON error ownership: every gateway-generated error has `Content-Type: application/json`, a compact one-field body, and no trailing newline. The exact bodies are `{"error":"not_found"}` for 404, `{"error":"unauthorized"}` for 401, `{"error":"rate_limited"}` for 429, `{"error":"internal_error"}` for 500, and `{"error":"bad_gateway"}` for 502. The 429 response also has a whole-seconds `Retry-After` value of at least one. Gateway-generated errors never forward upstream bodies. Upstream HTTP statuses and bodies are proxied faithfully when the transport interaction succeeds. Cancellation from the client is not rewritten as an unrelated success; it surfaces as the cancellation outcome and is logged as such.
- Configuration as a validated contract: the configuration is loaded and validated at startup. Validation fails closed: the process exits with a non-zero status and never serves traffic. Validation rejects a normalized prefix that duplicates another entry, an upstream URL that is not an absolute `http` or `https` origin with no userinfo, no query, and no fragment, a protected route that has no secret reference, and a rate-limit value that is not positive and finite. A normalized prefix whose path component collapses to empty is rejected. Timeouts for read, write, idle, and upstream are explicit and validated. There is no request-selected target and no open proxy.

## 7. Learning Objective
After completing this project you must be able to explain in your own words: why segment-aware prefix matching is the correct shape rather than substring matching, why invalid prefix spellings and exact duplicate prefixes fail startup, why route selection must be its own stage before authentication and rate limiting, why both identifiers are echoed and forwarded under pinned header names, why neither identifier is ever an authentication or rate-limit identity, why each JSON error class is owned by the gateway or the upstream and never mixed, and why cancellation is not a hidden success.

## 8. Functional Requirements
1. The gateway is configured at startup from a single configuration file with a fixed schema. No configuration is changed at runtime and no admin endpoint exists.
2. Each route has a normalized path prefix, an absolute `http` or `https` upstream origin, a protected flag, a prefix policy of strip or preserve, and a rate-limit declaration including a key kind.
3. Routing uses segment-aware normalized prefix matching. The longest matching prefix wins. An unmatched path returns the gateway 404. Two entries with the exact same normalized prefix fail validation at startup. Two entries that are disjoint at the same length are allowed. A longer prefix nested inside a shorter prefix is allowed.
4. `GET /healthz` serves the built-in health document with status 200, `Content-Type: application/json`, and the exact body `{"status":"ok"}` with no trailing newline. It bypasses authentication, rate limiting, and proxy interaction. Other methods on `/healthz` return 405 with `Content-Type: application/json` and the exact body `{"error":"method_not_allowed"}` with no trailing newline, and never reach those stages.
5. The full middleware order is enforced: recovery, request and correlation identifier handling, route selection, access logging, health and unmatched short circuit, protected-route authentication, rate limiting, proxy.
6. Request and correlation identifiers are generated, named with pinned header names, echoed in the response, and forwarded to the selected backend. Invalid inbound values are replaced by generated values. The same safe values appear in the response headers, in the access log entry, and in the upstream request headers.
7. JWT authentication protects only routes marked protected. Verification accepts only HS256 and requires a valid signature, expiry, configured issuer, configured audience, and non-empty subject. Missing, malformed, expired, wrong-issuer, wrong-audience, wrong-algorithm, or signature-invalid tokens all return the exact gateway 401 contract. The verified subject is placed into the request context for downstream stages.
8. The rate limiter applies only to protected routes after authentication. Its key is the verified subject. The built-in health route bypasses the limiter and therefore has no rate-limit key. Per-subject state is isolated and two distinct subjects never share buckets.
9. The reverse proxy forwards method, path under the prefix policy, body, and a cleaned header set to the fixed upstream. The client never chooses the target. Hop-by-hop headers are stripped in both directions.
10. Upstream transport failures (connection refused, DNS failure, write or read timeout) are mapped to the gateway 502 with a stable code and no internal leakage. Upstream HTTP statuses and bodies are proxied faithfully when the transport interaction succeeds.
11. Client cancellation surfaces as cancellation and is recorded in the log; it is not rewritten as an unrelated success.
12. Access logging records one structured line per request including method, path, status, request identifier, correlation identifier, authenticated subject when present, chosen upstream when present, and the cancellation flag when relevant. Logging records each outcome exactly once.
13. Configuration validation at startup rejects duplicate normalized prefixes, invalid upstream origins, missing secret references on protected routes, and non-positive or non-finite limits. Timeouts are explicit and validated.
14. The HTTP server is started with explicit timeouts. `SIGINT` and `SIGTERM` trigger a bounded `Server.Shutdown` that waits for in-flight handlers and refuses new connections.

## 9. Inputs and Outputs
- Configuration input: a single static file with a fixed schema. Example route entries:
  - `prefix: "/api/v1/notes"`, `upstream_origin: "http://127.0.0.1:9101"`, `protected: true`, `prefix_policy: "strip"`, `rate_limit: { rps: 10, burst: 20, key_kind: "subject" }`.
  - `/healthz` is not a configured proxy route; it is the single built-in public health path.
- Public health output: status 200, `Content-Type: application/json`, and exact body `{"status":"ok"}` with no trailing newline.
- Protected success output: the upstream response with the upstream status, the gateway's request identifier header, and the gateway's correlation identifier header.
- 401 output: a JSON body with a stable error code; the upstream is never called.
- 429 output: a JSON body with a stable error code and a `Retry-After` header; the upstream is never called.
- 404 output: a JSON body with a stable error code; the upstream is never called.
- 500 recovery output: a JSON body with a stable error code; the upstream is never called.
- 502 output: a JSON body with a stable error code and no internal leakage; the upstream error detail is in the log only.

## 10. Rules and Edge Cases
- Two configuration entries with the exact same normalized prefix fail validation at startup.
- Two entries that are same-length but disjoint are valid; the validator rejects only genuine ambiguity or duplication.
- A request for a path that matches no prefix returns the gateway 404 from the unmatched short circuit.
- The prefix policy is per route, declared in the configuration as strip or preserve, and is never inferred at runtime.
- A protected route with a missing token, an empty `Authorization` header, a header that is not `Bearer`, a `Bearer` token with the wrong number of segments, an expired token, a wrong-issuer token, a wrong-audience token, or a wrong-signature token all return 401 with the same stable code.
- The rate limiter is consulted only after authentication; an unauthenticated request is rejected at the auth stage and never reaches the limiter.
- A subject that exhausts its bucket returns 429 with `Retry-After` and never reaches the upstream.
- A request body with an unknown `Transfer-Encoding` value, a malformed `Content-Length`, or a body larger than the read limit is rejected before reaching the upstream.
- Client cancellation propagates as context cancellation into the upstream call and is recorded as the cancellation outcome, not as an unrelated success.
- A request received during shutdown is rejected by the listener before reaching the handler chain.
- An upstream that returns a `5xx` is forwarded as-is; the gateway does not retry and does not rewrite the body.
- An upstream that hangs longer than the upstream timeout returns 502.
- Hop-by-hop headers are stripped in both directions at the proxy.
- The access log never contains raw JWTs, raw `Authorization` headers, cookie values, or any field the user did not place in source.
- A panic in any later middleware still produces the gateway 500 JSON; recovery is the outermost stage.

## 11. Project Constraints
- The gateway depends only on `net/http`, `net/http/httputil`, and the JWT library the learner chooses. The chosen JWT library must support signature verification, expiry, issuer, and audience.
- No service registry, no dynamic upstream discovery, and no open proxy. Upstreams are fixed in the configuration.
- No TLS termination inside the gateway process; the gateway runs plain HTTP behind a trusted reverse proxy or load balancer in this project.
- No request or response body transformation, no header rewriting beyond hop-by-hop strip, and no response aggregation across upstreams.
- No database, no admin endpoint, no metrics endpoint beyond the access log, and no control plane.
- The configuration is a single static file in exactly one format; the learner chooses the format and stays with it.
- Unit tests must run locally with no Docker, no Redis, no remote service, and no upstream process beyond `httptest.Server` for fake backends.
- The optional integration test must start a fake upstream on loopback; it is separate from the unit tests and is gated behind a build tag and an environment flag.

## 12. Design Questions Before Coding
- Which header names will you pin for request identifier and correlation identifier, and what is the maximum allowed length and the validation rule for inbound values?
- Which JWT library will you choose, and what is its exact verification contract including signing algorithm, mandatory claims, allowed clock skew, and error mapping?
- What is your normalized prefix representation as a value type, and what is your segment-aware matching algorithm and its worst case?
- Where will per-subject rate-limit state live, what is your eviction policy, and how do you avoid retaining state for subjects that no longer appear?
- How do you propagate the correlation identifier from the request context into the `Director` so the upstream request is the only place that overwrites it?
- How do you ensure that the panic recovery stage still produces a clean JSON response when the panic was caused by a write failure?
- How do you test configuration validation without coupling to file I/O?
- How do you test that access logging records exactly once even when later middleware short-circuits?

## 13. Implementation Milestones
1. Define the configuration value type and the loader. Build the validator that rejects duplicate normalized prefixes, invalid upstream origins, missing secret references on protected routes, and non-positive or non-finite limits. Validator operates on a parsed value, not a file path.
2. Implement segment-aware normalized prefix matching and the route selection stage that returns a matched route with its prefix policy or an unmatched decision. Tests cover equal-length disjoint prefixes, nested prefixes with the longer one winning, and unmatched paths.
3. Implement the request and correlation identifier stages with pinned header names, generation rules, echo to response, and replacement of invalid inbound values.
4. Implement panic recovery as the outermost stage. Recovery always produces the clean JSON 500.
5. Implement access logging as a single stage that records every request outcome exactly once.
6. Implement the health and unmatched short circuit for the public prefix and unmatched paths.
7. Implement JWT authentication for protected routes with the verified subject placed into the request context.
8. Implement per-subject rate limiting with an injected clock and an injected store. Tests use no real sleep.
9. Implement the reverse proxy with the fixed upstream, hop-by-hop strip in both directions, prefix policy application, and the 502 mapping for transport failures.
10. Wire the full pinned middleware order with the server: recovery, identifiers, route selection, access logging, health and unmatched short circuit, protected-route authentication, rate limiting, proxy. Implement startup validation, explicit timeouts, and graceful shutdown on `SIGINT` and `SIGTERM`.
11. Write the unit test suite including upstream-down, race on the limiter, cancellation, configuration validation, and identifier-echo-and-forward.
12. Write the opt-in integration test gated by build tag and environment flag against a fake upstream on loopback.

## 14. Verification Cases the Learner Must Write
- Routing: a longer normalized prefix wins over a shorter one; two disjoint same-length prefixes are both valid and each routes correctly; a duplicate normalized prefix is rejected by the validator; unmatched paths return the gateway 404 from the unmatched short circuit.
- Authentication: protected route with no token, malformed token, expired token, wrong issuer, wrong audience, and wrong signature all return the gateway 401 with the same stable code; the public route with no token succeeds.
- Rate limiting: two distinct subjects get isolated buckets; a subject that exhausts its bucket returns 429 with `Retry-After`; a rate-limited request never reaches the upstream; refill is observable with an injected clock and no real sleep.
- Upstream behavior: a closed upstream returns 502 with the stable code and no internal leakage; a hanging upstream returns 502 within the upstream timeout; an upstream returning 200 forwards 200 and the proxied body.
- Method and body: a POST with a body is proxied with the body intact; an oversize body is rejected before reaching the upstream; a malformed `Content-Length` is rejected before reaching the upstream.
- Identifier linkage: the response headers, the access log entry, and the upstream request headers for a given request all carry the same request identifier and the same correlation identifier; an inbound invalid identifier value is replaced and the replacement is what appears in all three places.
- Middleware order: a missing-token request returns 401 and never reaches the limiter or proxy; a rate-limited request returns 429 and never reaches the proxy; an unmatched request returns 404 and never reaches auth, limiter, or proxy.
- Cancellation: a canceled client request cancels the upstream call via context and the log records the cancellation outcome.
- Concurrency: a race test exercises many concurrent requests with a shared rate limiter and asserts no panic and stable accounting.
- Configuration validation: each failure class fails closed at startup with a message naming the class.
- Hop-by-hop headers: the proxy strips hop-by-hop headers in both directions; tests assert they never appear in the upstream request or in the response.

## 15. Common Mistakes to Watch For
- Letting the rate limit key be a client-controlled header; the key for protected routes is the verified subject.
- Letting a panic in a later stage leave the response half-written; recovery must still produce JSON.
- Forwarding hop-by-hop headers across the proxy.
- Trusting an inbound identifier header when the value is malformed; spoofed identifiers let one client poison another client's log.
- Letting the rate limiter run before authentication; that would let an unauthenticated attacker exhaust another subject's bucket.
- Marking an upstream `5xx` as a gateway error in the log when it is actually an upstream status.
- Logging the JWT itself; only the subject identifier is allowed.
- Using substring matching for prefixes; that is not segment-aware and would let `/api` match `/apix`.
- Rejecting all overlap or all equal-length routes; only genuine duplication or ambiguity is invalid.
- Returning the upstream body for a gateway-owned 401, 404, 429, 500, or 502; the gateway owns those bodies.
- Rewriting a client cancellation as a successful response; cancellation is its own outcome.

## 16. Topics and References for Study
- The Go blog post on `net/http` middleware composition and `http.Handler` as a value.
- The `net/http/httputil` documentation for `ReverseProxy`, `Director`, and `ErrorHandler`.
- RFC 7230 section 6.1 for the canonical hop-by-hop header list.
- RFC 7519 (JWT) and RFC 7515 (JWS) for the verification model.
- The OWASP "Authentication Cheat Sheet" and "Rate Limiting Cheat Sheet".
- The `net/http` documentation on `Server.Shutdown`.
- The Go documentation on `context.WithCancel` for client cancellation propagation.

## 17. Self-Assessment Questions
- Why is segment-aware normalized prefix matching the correct rule, and what fails with substring matching?
- Why does route selection run before authentication and rate limiting rather than after?
- Why must both identifiers be echoed and forwarded, and why must invalid inbound values be replaced?
- Why are the request identifier and correlation identifier not authentication identities or rate-limit keys?
- Why is each JSON error class owned by the gateway or by the upstream and never mixed?
- Why is client cancellation surfaced as cancellation rather than rewritten as success?
- Why must exact duplicate prefixes and invalid prefix spellings fail startup?
- Why does the configuration validator reject only genuine ambiguity rather than every overlap?

## 18. Definition of Completion
The project is complete when the gateway reads a validated configuration, starts the HTTP server with the pinned middleware order and explicit timeouts, serves the public prefix without authentication or rate limiting, returns 404 from the unmatched short circuit, returns 401 for protected requests with missing or invalid JWTs, returns 429 with `Retry-After` for rate-limited subjects, returns 500 from recovery, returns 502 for upstream transport failures without internal leakage, echoes and forwards request and correlation identifiers under pinned header names with the same safe values in the response, log, and upstream request, logs one structured line per request exactly once and without raw credentials, shuts down gracefully on `SIGINT` and `SIGTERM`, runs all unit tests locally without external services, and runs the opt-in integration test against a fake upstream on loopback when the gating flag and build tag are set.

## 19. Optional Extensions
- An additional authentication scheme on a second protected prefix, with the route selection stage choosing the verifier by prefix.
- A second route kind that is authenticated but not rate-limited, demonstrating that rate limiting is genuinely opt-in per route and that the route selection stage is the single source of truth for which chain applies.
