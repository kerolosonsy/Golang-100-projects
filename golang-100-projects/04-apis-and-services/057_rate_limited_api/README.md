# Project 057 — Rate Limited API

## 1. Project Name and Number

Project 057 — `rate_limited_api`. Folder: `04-apis-and-services/057_rate_limited_api/`. README only; the learner writes all source and tests.

## 2. Project Idea

Build a per-client HTTP middleware that applies a token-bucket rate limit to a JSON API. The required exercise instance uses rate 2 tokens per second, capacity 5, cost 1 per request, idle TTL 10 minutes, and `MaxClients` 1000. The middleware reuses the conceptual model from Project 036 but lifts it into the HTTP layer: every incoming request is attributed to a client identity, that identity has its own bucket, and the middleware decides whether to forward the request to the next handler or to return `429 Too Many Requests` with a documented JSON envelope and `Retry-After`. The clock is injected. Each new bucket starts full at capacity 5. The state map is concurrency-safe. Clients idle for at least 10 minutes are removed by an explicit `Cleanup(now)` function that the application invokes at safe boundaries; tests call cleanup directly, never via goroutines or sleeps. At 1000 records a brand-new identity is rejected with `503 Service Unavailable` and a documented JSON envelope; an existing identity continues to use its bucket. No eviction policy exists in this project.

## 3. Why This Project Now?

Projects 046 through 056 produced an HTTP service with growing cross-cutting concerns: routing, middleware, JSON envelopes, auth, CORS. None of those projects decided who a request was "for". Project 057 introduces the client-identity question, which is the same question authentication (Project 059) and the API gateway (Project 091) will answer again. Token-bucket logic was first written in Project 036; here the same logic is rebuilt under concurrency and HTTP semantics. The OpenAPI contract in Project 058 must describe the `429` and `503` responses and the rate headers. The session middleware in Project 059 will reuse the same map-cleanup discipline with TTL. The graceful shutdown in Project 060 will reuse the "test the orchestration, not the wall clock" pattern this project establishes.

## 4. Prerequisites

Required earlier projects: Project 056, Project 046, and Project 036. Earlier HTTP, middleware, JSON envelope, and concurrency projects are useful review but are not formally required. The learner must already understand `net/http` handler chaining, the token-bucket model from Project 036 (capacity, refill rate, cost, the meaning of "no tokens"), `context.Context`, and how to write `httptest` tests that drive the middleware through a real `http.Handler` chain.

## 5. What You Must Know Before Starting

- Token-bucket mechanics: capacity, refill rate, current tokens, cost per request, what happens when tokens go below cost.
- How to inject a clock into Go code without changing call sites at test time. The limiter accepts a clock interface that returns the current time; tests use a fake clock that advances only when the test advances it.
- `net/http` middleware chaining: how to wrap a handler, how to call the next handler, how to short-circuit before calling it, how to set the response status and headers.
- The semantics of `RemoteAddr`, including its `host:port` form, the difference between IPv4 and IPv6 textual forms, and the `net.SplitHostPort` helper.
- The semantics of `X-Forwarded-For`: it is a header that any client can forge, it contains a comma-separated list of addresses, and trusting it changes security posture.
- The semantics of `net.ParseIP` and how it distinguishes valid and invalid IP literals.
- Concurrency primitives in Go: `sync.Mutex`, the discipline of holding a mutex only around the data structure, and how to use `-race` to validate correctness.
- How `time.Duration` arithmetic works for computing "tokens to add since last seen" without losing precision in tests.

## 6. Explanation of New Concepts

**Per-client state.** Each client identity has its own bucket. The limiter keeps a map from identity to bucket state. The map is guarded by a single mutex; the per-bucket state lives inside the value type. When a request arrives, the limiter looks up the bucket, charges it, and updates `lastSeen`.

**Clock injection.** A production limiter reads `time.Now()`. A testable limiter reads from a clock interface that the constructor accepts. Tests use a fake clock that returns a fixed time and lets the test advance it explicitly. No test sleeps, no `time.After`, no real timers.

**Client identity.** The default identity is the IP host parsed from `Request.RemoteAddr` using `net.SplitHostPort`. The function supports IPv4 and bracketed IPv6. A missing port, an empty string, or any unparsable input is a `400 Bad Request` with the JSON envelope `{ "code": "invalid_client_address", "message": "client address could not be parsed" }`; no bucket is created and the next handler is not called.

**Trusted proxy CIDRs.** `X-Forwarded-For` is ignored by default. The optional configured trusted-proxy CIDR list is used only for one narrow rule: when the immediate peer IP is inside the allowlist, the middleware accepts `X-Forwarded-For` only if it contains exactly one comma-free, syntactically valid IP, and uses that IP as the client identity. If `X-Forwarded-For` is missing, empty, contains a comma, contains an unparsable IP, or contains more than one IP, the middleware falls back to the immediate peer IP. If the immediate peer IP is not in the trusted-proxy allowlist, `X-Forwarded-For` is ignored entirely. Multi-proxy chains are out of scope.

**TTL cleanup.** Each bucket records a `lastSeen` timestamp. A `Cleanup(now)` function removes any bucket whose `now - lastSeen` is greater than or equal to 10 minutes. The function is a pure function with no goroutine, no ticker, and no sleep. The application invokes cleanup at known boundaries. Tests call cleanup directly. `lastSeen` is updated on every successfully identified request, including a request that is rejected with `429` for rate exhaustion; it is not updated on requests that fail the identity check.

**Bounded memory.** `MaxClients` is exactly 1000. When the map already holds 1000 records and a request arrives whose identity is not already present, the middleware rejects the request with `503 Service Unavailable` and the JSON envelope `{ "code": "limiter_capacity_reached", "message": "rate limiter client capacity reached" }`. No bucket is created for that identity, and no `Retry-After` is emitted. An existing identity continues to use its bucket. The map size never exceeds 1000. Reclaiming capacity requires the application to call `Cleanup` explicitly; the limiter does not run cleanup itself.

**Rate headers.** Allowed responses and `429` responses both include three headers whose values are derived from the bucket state. `X-RateLimit-Limit` is exactly `5`. `X-RateLimit-Remaining` is an integer equal to the floor of the post-decision usable tokens. `X-RateLimit-Reset` is the Unix seconds value, in the current fake-clock frame, at which the next request would be admitted when at least one token remains; it is the current fake-clock Unix second when at least one token remains. These are legacy `X-` headers and the middleware does not claim they are an IETF standard.

**429 response.** Rejected requests due to rate exhaustion receive `429 Too Many Requests`, a JSON body exactly equal to `{ "code": "rate_limited", "message": "rate limit exceeded", "retry_after_seconds": N }` where `N` is the integer ceiling of the seconds needed for one full token to refill, with a minimum of 1, and the `Retry-After` header is exactly `N`. `N` and `Retry-After` must match exactly.

**Fractional tokens.** A bucket with `0.7` tokens after refill still cannot serve a request of cost 1. A bucket with `1.0` tokens can serve one request. The `retry_after_seconds` value rounds up so that after waiting `N` seconds the next request is admitted even if the calculation produces a fractional token count.

## 7. Learning Objective

After finishing this project, the learner can explain how a token-bucket limiter is lifted from a single-process design (Project 036) into a per-client HTTP middleware, why the clock must be injected, why `X-Forwarded-For` must not be trusted by default, why a TTL plus an explicit cleanup call bounds memory without leaking goroutines, and how the rate headers and `Retry-After` value are computed. The learner can also describe the exact behaviour at `MaxClients` and defend that choice.

## 8. Functional Requirements

1. The middleware accepts a configuration that pins: rate `2` tokens per second, capacity `5`, cost `1` per request, TTL `10 minutes`, `MaxClients` `1000`, an injected clock, an optional configured trusted-proxy CIDR list (which may be empty), and the rate-limit envelope constants.
2. The constructor validates the configuration and rejects contradictory inputs.
3. The default client-identity function parses the host portion of `Request.RemoteAddr` using `net.SplitHostPort`. It supports IPv4 and bracketed IPv6. A missing port, an empty string, an unparsable host, or any other malformed input returns `400 Bad Request` with the JSON envelope `{ "code": "invalid_client_address", "message": "client address could not be parsed" }`. No bucket is created. The next handler is not called. `lastSeen` is not updated.
4. When a trusted-proxy CIDR allowlist is configured and the immediate peer IP is inside the allowlist, the middleware accepts `X-Forwarded-For` only if it contains exactly one comma-free, syntactically valid IP, and uses that IP as the client identity. Any other `X-Forwarded-For` value (missing, empty, contains a comma, contains an unparsable IP, contains more than one IP) falls back to the immediate peer IP. When the immediate peer IP is not in the allowlist, `X-Forwarded-For` is ignored entirely and the client identity is the immediate peer IP. Multi-proxy chains are out of scope.
5. The state map is guarded by a mutex. The mutex is acquired only around the data structure; it is not held while resolving identity, while calling the next handler, or while writing the response body.
6. Each bucket tracks `tokens`, `lastRefill`, and `lastSeen`. A new bucket starts with `tokens` equal to `5`. Refill is computed as `min(5, tokens + (now - lastRefill) * rate)` and applied on every allowed decision and every `429` decision. Cost is exactly `1` per allowed request. Tokens are never refunded. If the injected clock returns a value earlier than a bucket's `lastRefill` or `lastSeen`, the elapsed refill and idle durations are clamped to zero and neither stored timestamp is moved backward. Time moving backward never mints tokens, makes a recently seen client appear older, or causes `Cleanup` to remove a bucket.
7. A request whose bucket has fewer than `1` token after refill is rejected with `429 Too Many Requests`, the documented JSON envelope, and `Retry-After` equal to `N` where `N` is the integer ceiling of the seconds needed for one full token to refill, with a minimum of `1`. `N` and `Retry-After` match exactly. The `X-RateLimit-*` headers are present on the `429` response. The next handler is not called. `lastSeen` is updated.
8. A request whose bucket has at least `1` token after refill is allowed. The bucket is charged exactly `1` token. The `X-RateLimit-*` headers are present on the allowed response. The next handler is called exactly once. `lastSeen` is updated.
9. The middleware writes the `X-RateLimit-*` headers on every allowed response and every `429` response. The values are exactly as documented in the Explanation of New Concepts section. The headers are not written on `400`, `503`, or any other response where the limiter did not perform a per-bucket decision.
10. The middleware is consulted before the next handler runs. The next handler is called exactly once per allowed request and zero times per rejected request.
11. A `Cleanup(now)` function removes any bucket for which `now - lastSeen >= 10 minutes`. The function is pure: no goroutine, no ticker, no sleep, no background work. The application is responsible for invoking it. Tests call it directly. After `Cleanup` returns, the map size is exactly the number of buckets whose `lastSeen` is within the TTL window. The map size never exceeds `1000`. If `now` is earlier than a bucket's `lastSeen`, that bucket is never removed.
12. When the map already holds `1000` records and a request arrives whose identity is not already present, the middleware rejects the request with `503 Service Unavailable` and the JSON envelope `{ "code": "limiter_capacity_reached", "message": "rate limiter client capacity reached" }`. No bucket is created. No `Retry-After` is emitted. The `X-RateLimit-*` headers are not emitted on this response. The next handler is not called. `lastSeen` is not updated. An existing identity continues to use its bucket normally. Reclaiming capacity requires an explicit `Cleanup` call.
13. Allowed responses and `429` responses include the `X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset` headers with the exact values described above.
14. The middleware is safe for concurrent use.

## 9. Inputs and Outputs

Inputs the middleware reads from the request: `RemoteAddr`, optionally `X-Forwarded-For`, and the request method and path (the latter two only to set headers, not for routing decisions). Outputs are: either a short-circuit `400`, `429`, or `503` response with the documented envelope, or a forwarded request with the `X-RateLimit-*` headers added to the response. Example textual inputs and expected textual outputs:

- Request from `198.51.100.7:51234`. Configuration: rate 2, capacity 5, cost 1, TTL 10 minutes, MaxClients 1000. First five requests return from the downstream handler with `X-RateLimit-Limit: 5`, `X-RateLimit-Remaining` decreasing from 4 to 0, `X-RateLimit-Reset` equal to the current fake-clock Unix second. The sixth request at the same instant returns `429`, body exactly `{ "code": "rate_limited", "message": "rate limit exceeded", "retry_after_seconds": 1 }`, header `Retry-After: 1`, `X-RateLimit-Limit: 5`, `X-RateLimit-Remaining: 0`, `X-RateLimit-Reset` equal to the Unix second one tick in the future. Handler is not called.
- Request from `198.51.100.7:51234` followed by an immediate request from `203.0.113.4:51234`. The two clients have independent buckets. The second request is not affected by the first client's exhausted bucket.
- Request from `127.0.0.1:51234` while `127.0.0.1/32` is in the trusted-proxy allowlist and `X-Forwarded-For: 198.51.100.7`. The header contains exactly one comma-free, syntactically valid IP; the client identity is `198.51.100.7`.
- Request from `127.0.0.1:51234` while `127.0.0.1/32` is in the trusted-proxy allowlist and `X-Forwarded-For: 198.51.100.7, 10.0.0.1`. The header contains a comma; the client identity is the immediate peer `127.0.0.1`.
- Request from `127.0.0.1:51234` while `127.0.0.1/32` is in the trusted-proxy allowlist and `X-Forwarded-For: not-an-ip`. The header contains an unparsable IP; the client identity is the immediate peer `127.0.0.1`.
- Request from `127.0.0.1:51234` while `127.0.0.1/32` is in the trusted-proxy allowlist and no `X-Forwarded-For`. The header is missing; the client identity is the immediate peer `127.0.0.1`.
- Request from `198.51.100.7:51234` while `10.0.0.0/8` is in the trusted-proxy allowlist and `X-Forwarded-For: 203.0.113.4`. The immediate peer is not trusted; the client identity is the immediate peer `198.51.100.7`.
- Request from `[2001:db8::1]:443`. The bracketed IPv6 is parsed and used as the client identity.
- Request from a `RemoteAddr` that is empty or that lacks a port. The response is `400`, body exactly `{ "code": "invalid_client_address", "message": "client address could not be parsed" }`, no `X-RateLimit-*` headers, no `Retry-After`. The handler is not called. No bucket is created.
- Configuration MaxClients 1000, exactly 1000 distinct clients in the map. A request from a brand-new identity returns `503`, body exactly `{ "code": "limiter_capacity_reached", "message": "rate limiter client capacity reached" }`, no `Retry-After`, no `X-RateLimit-*` headers. The handler is not called. No bucket is created. The map size is still 1000. An existing identity continues to use its bucket.
- Configuration TTL 10 minutes. A bucket last seen exactly 10 minutes ago is removed by `Cleanup(now)`. A bucket last seen 9 minutes 59 seconds ago is preserved.

## 10. Rules and Edge Cases

- `RemoteAddr` that is empty, has no port, contains an unparsable host, or otherwise fails `net.SplitHostPort` is a `400` with the documented envelope. No bucket is created. `lastSeen` is not updated.
- `X-Forwarded-For` is empty, missing, contains a comma, contains an unparsable IP, or contains more than one IP: fall back to the immediate peer IP.
- When the immediate peer IP is not in the configured trusted-proxy CIDR list, `X-Forwarded-For` is ignored entirely.
- `Retry-After` is always a non-zero integer. It is never zero for a rejected request.
- A request that is allowed but whose downstream handler returns a `5xx` status still consumed its token. Tokens are never refunded.
- The clock is consulted only inside `Allow(now)`, `Cleanup(now)`, and any function the application calls. The clock is never consulted from a goroutine owned by the middleware.
- The map size, after any sequence of operations, never exceeds `MaxClients`. The `503` path does not transiently exceed `MaxClients`.
- No mutex is held while resolving identity, while calling the next handler, or while writing the response body.
- A request whose identity fails parsing produces `400` and is not affected by the rate-limit logic.
- A request that hits `MaxClients` produces `503` and is not affected by the rate-limit logic.
- The `X-RateLimit-Remaining` header is the floor of the post-decision usable tokens. After an allowed request that takes the bucket from `3.7` tokens to `2.7` tokens, the header is `2`. After a request that takes the bucket from `0.0` to `0.0` (because the bucket is exhausted), the request is rejected and the header is `0`.
- The `X-RateLimit-Reset` header is the Unix-seconds instant at which the next request would be admitted when at least one token remains. When at least one token already remains, `X-RateLimit-Reset` is the current fake-clock Unix second.

## 11. Project Constraints

- Standard library only. No third-party rate-limit libraries, no Gin, no Fiber.
- No goroutines owned by the middleware. No background tickers, no background cleanups, no `time.AfterFunc`. The application owns the cleanup cadence.
- No real wall clock in tests. The clock is always injected.
- No external services. Tests run with `httptest` and in-process goroutines only.
- Configuration is set at construction time and treated as read-only at request time.

## 12. Design Questions Before Coding

- What does the clock interface look like? A single method returning the current time.
- How is the bucket stored? A struct embedded by value inside the map entry so the mutex protects both the map and the bucket.
- How is the trusted-proxy parsing rule expressed precisely? The middleware parses `X-Forwarded-For` as a single token. The token must be a single IP literal with no comma and no whitespace.
- How is `MaxClients` enforced? A size check immediately after resolving the identity and before any per-bucket decision.
- How is `Retry-After` computed? The middleware computes the integer ceiling of `(cost - tokens) / rate` seconds, with a minimum of `1`.
- How is the rate-limit response observed in tests? Through a custom recorder and by reading the response body and headers.
- How is concurrency safety validated? `-race` plus hammer tests that fire many goroutines at the same client and many goroutines at different clients.

## 13. Implementation Milestones

1. Sketch the configuration struct with the exact values on paper.
2. Define the clock interface and the fake clock used by tests.
3. Implement the default client-identity function with `RemoteAddr` parsing using `net.SplitHostPort` and malformed-input handling.
4. Implement the trusted-proxy-aware client-identity function behind the single documented parsing rule.
5. Implement the bucket value type with `tokens`, `lastRefill`, and `lastSeen`.
6. Implement the map-backed store guarded by a mutex, with the per-bucket `Allow` decision and the `Cleanup` pass described above.
7. Implement the response writers for `400`, `429`, and `503` together, with the documented JSON envelopes, the `Retry-After` header on `429`, and the rule that rate-limit headers are emitted only on the allowed and `429` responses.
8. Implement the `MaxClients` rejection branch and the middleware function: resolve identity, branch on outcome, set `X-RateLimit-*` headers, write the documented response or call the next handler exactly once.
9. Wire the middleware around a no-op handler in tests and verify the call count is observable for allowed, `429`, `400`, and `503` paths.
10. Review the verification list and confirm every item is covered before declaring the project complete.

## 14. Verification Cases the Learner Must Write

Each item is a behavioural specification. The learner writes the corresponding `go test` code.

- New bucket starts full: a request from a brand-new identity shows `X-RateLimit-Remaining: 4` after one request, `X-RateLimit-Limit: 5`, and the downstream handler runs once.
- Burst up to capacity: a client makes 5 requests within the same tick; all 5 are allowed; the 6th request within the same tick is rejected with `429`, body exactly `{ "code": "rate_limited", "message": "rate limit exceeded", "retry_after_seconds": 1 }`, header `Retry-After: 1`, `X-RateLimit-Limit: 5`, `X-RateLimit-Remaining: 0`, and the downstream handler is not called.
- Refill exact: after advancing the fake clock by exactly `0.5` seconds at rate 2 tokens per second, the bucket has `1` new token; the next request is allowed. Advancing by `0.49` seconds leaves fewer than `1` new token; the next request is rejected with `Retry-After: 1`.
- Refill cap: a bucket idle for a long time refills to exactly `5` and never exceeds `5`.
- Boundary: `tokens` is exactly equal to `1` after refill. The request is allowed and the bucket is reduced to `0`. The next request at the same instant is rejected.
- Backward clock on refill: when the injected clock returns a value earlier than the bucket's `lastRefill` or `lastSeen`, the refill delta is clamped to zero. No tokens are minted, and neither stored timestamp moves backward. When the clock later returns to its previous value, cleanup still treats the intervening request as recent rather than aging the bucket from the earlier timestamp.
- Backward clock on cleanup: when the injected clock returns a value earlier than a bucket's `lastSeen`, the bucket is preserved by `Cleanup`. `now - lastSeen` is computed as zero, not as a negative number, and the bucket is not removed.
- `retry_after_seconds` ceiling: a request that arrives with `0.1` tokens remaining computes `retry_after_seconds` as the integer ceiling of `(1 - 0.1) / 2 = 0.45` seconds, which rounds up to `1`. `Retry-After` matches exactly.
- `retry_after_seconds` minimum: a request that arrives with `0.99` tokens remaining computes `retry_after_seconds` as the integer ceiling of `(1 - 0.99) / 2 = 0.005` seconds, which rounds up to `1`. `Retry-After` matches exactly.
- 429 envelope shape: the body parses as JSON and matches the exact documented envelope. `Retry-After` equals `retry_after_seconds`. The `X-RateLimit-*` headers are present.
- 503 envelope shape: when the map is at `MaxClients` and a brand-new identity arrives, the body is exactly `{ "code": "limiter_capacity_reached", "message": "rate limiter client capacity reached" }`, no `Retry-After`, no `X-RateLimit-*` headers, no bucket created, handler not called.
- 400 envelope shape: when `RemoteAddr` is empty or unparsable, the body is exactly `{ "code": "invalid_client_address", "message": "client address could not be parsed" }`, no `Retry-After`, no `X-RateLimit-*` headers, no bucket created, handler not called.
- Isolation: client A exhausts its bucket; client B's first request is allowed and shows `X-RateLimit-Remaining: 4`.
- Same-client concurrent: a hammer test fires `N` goroutines against the same client. The total number of allowed requests across the test is bounded by the bucket's capacity plus refill; no race is reported under `-race`.
- Different-client concurrent: a hammer test fires `N` goroutines against `M` distinct clients; each client sees an independent count.
- Spoofed `X-Forwarded-For` ignored: with no trusted-proxy configuration, two requests with different `X-Forwarded-For` values from the same peer share a bucket.
- Trusted proxy, single valid XFF: a request whose immediate peer is trusted and whose `X-Forwarded-For` is exactly one comma-free IP uses that IP as the identity.
- Trusted proxy, XFF missing: a request whose immediate peer is trusted and whose `X-Forwarded-For` is missing uses the immediate peer as the identity.
- Trusted proxy, XFF empty: a request whose immediate peer is trusted and whose `X-Forwarded-For` is empty uses the immediate peer as the identity.
- Trusted proxy, XFF with comma: a request whose immediate peer is trusted and whose `X-Forwarded-For` contains a comma uses the immediate peer as the identity.
- Trusted proxy, XFF unparsable: a request whose immediate peer is trusted and whose `X-Forwarded-For` contains an unparsable IP uses the immediate peer as the identity.
- Untrusted peer, XFF present: a request whose immediate peer is not trusted and whose `X-Forwarded-For` is present ignores the header and uses the immediate peer as the identity.
- IPv4 parsing: a request from `198.51.100.7:443` is attributed to `198.51.100.7`.
- IPv6 parsing: a request from `[2001:db8::1]:443` is attributed to `2001:db8::1`. A request from `[::1]:443` is attributed to `::1`.
- TTL boundary: a bucket last seen exactly 10 minutes ago is removed by `Cleanup(now)`. A bucket last seen 9 minutes 59 seconds ago is preserved.
- TTL active preservation: an empty bucket whose `lastSeen` is recent is preserved.
- `lastSeen` on 429: a bucket whose previous request was rejected with `429` has its `lastSeen` updated by the `429`. A subsequent `Cleanup(now)` does not remove it within the TTL window.
- `MaxClients` reject-new: with 1000 distinct clients, a brand-new identity receives `503` with the documented envelope; the map size is exactly 1000.
- Existing identity under `MaxClients`: with 1000 distinct clients, an existing identity continues to receive allowed responses with the `X-RateLimit-*` headers.
- Capacity reclaim after cleanup: after `Cleanup(now)` removes idle buckets, the freed slots can be used by new identities; an identity that was previously rejected with `503` is now allowed. The map size never exceeds 1000 at any point.
- Mutex discipline: a test invokes `Cleanup(now)` and the middleware concurrently under `-race`; no race is reported.
- Next-call count: rejected requests cause the next handler to be called zero times. Allowed requests cause the next handler to be called exactly once. `400` and `503` responses cause the next handler to be called zero times.
- No sleeps: every test uses the fake clock only. `go test -race` runs cleanly and finishes in well under a second.

## 15. Common Mistakes to Watch For

- Trusting `X-Forwarded-For` unconditionally. Any client can forge the header. The middleware trusts it only under the documented narrow rule.
- Spawning a background goroutine for cleanup. The application owns cleanup; the middleware must not own a goroutine or a ticker.
- Using `time.Sleep` in tests. Tests are nondeterministic and slow. Use the fake clock and call `Cleanup` directly.
- Holding a mutex across the next handler or across the response body write. That can deadlock under load and prevents the limiter from being safely composed with other middleware.
- Setting `Retry-After: 0` for a rejected request. The header would lie about the wait.
- Returning a `500` or `503` for rate limiting. The status is `429`.
- Returning `429` for `MaxClients`. The status for `MaxClients` is `503`.
- Refunding tokens when the downstream handler errors. The middleware consumes tokens before the handler runs and does not refund.
- Letting the map grow without bound. `MaxClients` plus a TTL plus a cleanup call is the bound; without all three, memory leaks.
- Using `time.Now()` inside the limiter logic instead of the injected clock. Tests will silently pass against a wall clock and fail in CI under load.
- Treating `X-Forwarded-For` as a multi-proxy chain. Multi-proxy chains are out of scope.
- Treating IPv6 addresses as if they were IPv4. The bracketed form must be parsed by `net.SplitHostPort`.
- Splitting `X-Forwarded-For` on commas. The rule is "exactly one comma-free IP". Splitting is the wrong operation.
- Computing `retry_after_seconds` as a fractional value. The value is an integer ceiling with a minimum of `1`. `Retry-After` must match exactly.

## 16. Topics and References for Study

- Go `net/http` package documentation, especially `Handler`, `ResponseWriter`, `http.Request.RemoteAddr`.
- Go `net` package documentation, especially `net.SplitHostPort` and `net.ParseIP`.
- Go `sync` and `sync/atomic` package documentation. The mutex discipline that holds a lock only around the data structure, never across user code.
- Token-bucket concepts and the `golang.org/x/time/rate` package documentation as a reference implementation, but the learner must not import it.
- RFC 6585 for the `429 Too Many Requests` status code semantics.
- RFC 7231 for `Retry-After`.
- The IETF draft `RateLimit-Header` for the convention used by `X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset` headers. The middleware does not claim these are an IETF standard.
- MDN reference for `X-Forwarded-For` and the warnings about trusting it.

## 17. Self-Assessment Questions

1. Why is the clock injected rather than read from `time.Now()` inside the limiter? What would the test suite look like otherwise?
2. What is the security risk of trusting `X-Forwarded-For` by default? Describe an attacker who controls the header.
3. Why must the cleanup function be invoked by the application rather than run as a background goroutine inside the middleware? Give one production scenario in which the goroutine approach would cause a leak.
4. Why are tokens not refunded when the downstream handler returns `5xx`? When would refunding be the wrong choice?
5. Why is `Retry-After` required to be a non-zero integer? What client behaviour does this enforce?
6. Why is the trusted-proxy parsing rule "exactly one comma-free IP" rather than "leftmost entry of a comma-separated list"? What changes if the rule is broadened?
7. Why is the `MaxClients` rejection a `503` and not a `429`? How would the client behave differently if it were a `429`?
8. Why must time moving backward never mint tokens and never remove a bucket? Give one production scenario in which a misconfigured clock could trip that rule.

## 18. Definition of Completion

The project is complete when, in addition to the rules above:

- Every item in the verification list is a passing test that the learner wrote themselves.
- The tests pass under `go test -race ./...` from the project folder.
- The middleware contains no third-party imports and starts no goroutines of its own.
- The clock interface and the fake clock are reused by every test, with no `time.Sleep` anywhere.
- The configuration struct has the exact fixed values and a constructor that rejects contradictory inputs.
- The map size never exceeds `MaxClients` at any point.
- The learner can answer every self-assessment question without rereading the README.

## 19. Optional Extensions

At most two. Pick one only if the core project is already complete and tested. Optional extensions must not change the core behaviour, must not add speculative eviction, and must not introduce another third-party dependency.

- Add a per-identity counters struct (`Allowed`, `Limited`, `Rejected`) that the application may read after `Cleanup` for structured logging.
- Add a `Stop` method that prevents new identities from being added and reports the map size. The map is not cleared by `Stop`.
