# Project 078 — Reverse Proxy

## 1. Project Name and Number

- Project 078, reverse_proxy.
- This README is a learning guide only.
- You will create every source and test file yourself in `06-networking/078_reverse_proxy/`.
- This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea

A reverse proxy built on `net/http/httputil.ReverseProxy` that forwards requests to a single fixed configured upstream URL. The proxy is not an open proxy: it never accepts an arbitrary target from a request. The configured public prefix is exactly `/api`. The path mapping is pinned: the exact path `/api` maps to the upstream base path root, the path `/api` followed by a single forward slash and a safe suffix maps to the upstream root followed by that suffix, and the query bytes and semantics are preserved. Anything outside this boundary returns 404 without an upstream call. Malformed escaping, backslashes, encoded slash or backslash, and decoded dot segments are rejected before proxying. The proxy enforces a strict outbound Host and URL policy from the configured upstream, strips all inbound forwarding headers defensively, and synthesizes forwarding headers only for an exactly trusted direct peer. The proxy removes hop-by-hop headers per the HTTP specification with `Trailer` (not `Trailers`) and includes `Proxy-Connection` defensively; `TE` is stripped completely for this project. The proxy pins an exact response allowlist, returns a pinned 502 JSON public error on pre-response transport failure, and never retries for any method.

## 3. Why This Project Now?

- This project requires Project 077 (http2_push_server) as the immediate predecessor, Project 071 (tcp_echo_server) for TCP framing, idle deadlines, and per-connection protocol error discipline, and Project 060 (graceful_shutdown_web) for graceful server shutdown and lifecycle ownership.
- This project introduces the `httputil.ReverseProxy` discipline: a single fixed upstream, a pinned exact path mapping, an exact forwarding header policy, an exact response header allowlist, and an exact 502 envelope on pre-response transport failure.

## 4. Prerequisites

- Projects 077, 071, and 060 are required prerequisites.
- Project 077 is the immediate predecessor for HTTP/2 protocol and connection identity.
- Project 071 is required for TCP connection handling, byte framing, idle deadlines, accept-loop shutdown, and per-connection protocol error discipline.
- Project 060 is required for graceful server shutdown and lifecycle ownership.
- No public network, no Docker, no environment variables for the upstream URL.
- The upstream URL is a configuration value supplied at construction.

## 5. What You Must Know Before Starting

- Know the `net/http` request and response model, `httputil.ReverseProxy`, the modern rewrite boundary that sets the outbound URL and Host, hop-by-hop headers per the HTTP specification including the `Trailer` name and `Proxy-Connection`, the `X-Forwarded-*` family and its trust model, request body streaming, response body close, context cancellation propagation, transport-level timeouts, the difference between idempotent and non-idempotent methods, and the race detector.

## 6. Explanation of New Concepts

### Concepts

- The configured upstream is one origin-only `http` or `https` URL validated at construction.
- Scheme and host are required; userinfo, query, fragment, and any non-root path are forbidden, and a trailing root slash is normalized away.
- Invalid configuration is a startup error.
- The configured public prefix is exactly `/api`.
- An exact inbound path `/api` maps to upstream root `/`.
- An inbound path that begins with `/api/` followed by a safe suffix maps to the upstream root followed by that suffix.
- The inbound raw query is preserved verbatim, including ordering and percent-encoding.
- Anything else is not proxied and returns 404 without an upstream call.

- A safe suffix is one or more nonempty path segments separated only by literal forward slashes.
- Within each segment, the permitted literal characters are ASCII letters, digits, `-`, `.`, `_`, `~`, `!`, `$`, `&`, apostrophe, `(`, `)`, `*`, `+`, `,`, `;`, `=`, `:`, and `@`; a whole segment equal to `.` or `..` is forbidden.
- Percent-encoding is rejected anywhere in the path for this project, as are backslashes and empty interior segments, so encoded separators and double-decoding ambiguity never reach the upstream.
- Suspicious or malformed in-boundary paths return 400 without an upstream call; paths outside `/api` and `/api/` return 404.

- The proxy is not an open proxy.
- The outbound URL and Host come only from the configured upstream.
- An inbound request cannot influence the target host, port, or scheme through any header.
- The proxy strips all inbound forwarding headers before composing the outbound request.
- The stripped set includes `Forwarded`, every `X-Forwarded-*` variant, and legacy target-influencing headers such as `X-Original-URL`, `X-Rewrite-URL`, and `X-Forwarded-Server`.

- Trusted-forwarding synthesis is exact.
- The proxy parses the direct peer IP from `Request.RemoteAddr`.
- If and only if that parsed peer IP is exactly in the configured trusted peer set, the proxy synthesizes a single outbound `X-Forwarded-For` containing that IP, a single outbound `X-Forwarded-Host` containing the validated inbound `Host`, and a single outbound `X-Forwarded-Proto` containing `https` if the inbound connection is over TLS or `http` otherwise.
- If the peer is not in the trusted set, the proxy sends no `X-Forwarded-*` headers at all.
- The proxy does not chain, append, or merge with any inbound forwarding value; there is no inbound forwarding value after stripping.

- The proxy removes hop-by-hop headers per the HTTP specification.
- The hard-coded removal list is `Connection`, `Keep-Alive`, `Proxy-Authenticate`, `Proxy-Authorization`, `Proxy-Connection`, `Transfer-Encoding`, and `Upgrade`.
- The removal list also includes `Trailer` (not `Trailers`) and any header named in a `Connection` header value. `TE` is stripped completely for this project as part of the hop-by-hop list; the proxy does not forward `TE`.
- Headers that are not hop-by-hop and not in the forwarding-header strip set are forwarded.

- The proxy pins an exact response allowlist rather than forwarding all upstream response headers.
- The allowlist is exactly `Content-Type`, `Content-Length`, `Content-Encoding`, `Cache-Control`, `ETag`, `Last-Modified`, `Retry-After`, and `Vary`.
- The proxy strips every other upstream response header and every hop-by-hop header from the upstream response.
- In particular, it does not forward an upstream `Location` that could expose the configured origin.
- The proxy sets exactly one `Via` value, the literal `1.1 tutorial-proxy`, on every successful response.
- It adds no invented request identifier, server, or other header.

- The public error contract is exact and applies only when transport fails before any upstream response headers have been received.
- The status is 502.
- The `Content-Type` is `application/json`.
- The body is a single JSON object whose only field is `error` with value `bad_gateway`, with no trailing newline.
- Upstream HTTP 4xx and 5xx responses are successful upstream responses and are forwarded through to the client with their own status, with the response header allowlist applied; they are not converted to 502.

- Failure paths and response lifetime are pinned precisely.
- For every upstream response it receives, `httputil.ReverseProxy` owns and closes the upstream response body after copying completes or aborts; learner code does not manually close a body that the proxy never received.
- A pre-response dial failure has no upstream response body and is reported through `ErrorHandler` with the pinned 502 envelope.
- A failure or cancellation observed after downstream streaming has begun cannot replace the already-written status with 502; the in-flight response terminates or is cancelled, and the proxy does not retroactively rewrite headers or status.
- Inbound client cancellation cancels the outbound request through the injected transport and triggers cleanup; it is not presented as a synthetic 502.

- The proxy does not retry for any method.
- It does not retry idempotent methods, does not retry non-idempotent methods, and does not support retry on transport failure.
- It does not cache responses, does not load balance, does not handle WebSocket upgrades, and does not discover upstream dynamically.
- These are explicit exclusions and not gaps in scope.

- The injected transport owns a 2-second dial timeout, 2-second TLS-handshake timeout, 5-second response-header timeout, and 30-second idle-connection timeout as production defaults.
- The inbound request context supplies whole-request cancellation and any earlier caller deadline; tests inject deterministic failure transports or shorter finite settings without sleeping for expiry.
- The proxy does not implement its own panic-recovery requirement; recovery from panic in the proxy path is not part of the contract and is not asserted by tests.

- Text-only protocol examples are permitted.
- As a prose shape: an inbound `GET /api/items?limit=10` is rewritten as `GET /items?limit=10` to the upstream base path with the configured outbound Host.
- An inbound exact `GET /api` is rewritten as `GET /` to the upstream base path with the configured outbound Host and with the query absent.
- An inbound `POST /api/v1/items` with a JSON body is forwarded as `POST /v1/items` with the body streamed and the inbound `Host` replaced by the configured upstream host.
- An inbound `GET /api/../etc` is rejected with 400 without an upstream call.
- An inbound `GET /api/%2fhost` is rejected with 400 without an upstream call.
- An inbound `GET /api/items` that carries `X-Forwarded-For: 9.9.9.9` from a loopback direct peer results in an outbound `X-Forwarded-For` containing the loopback peer, not `9.9.9.9`; if the peer were not trusted, no `X-Forwarded-For` would be sent.
- An inbound `GET /api/items` with `TE: trailers` results in `TE` not forwarded.
- An inbound `GET /api/items` with `Connection: close, X-Custom` results in `X-Custom` removed as a hop-by-hop header named in `Connection`.
- A successful upstream response includes exactly the response header allowlist plus `Via: 1.1 tutorial-proxy`, and `ReverseProxy` closes the upstream body it received.
- An upstream dial failure returns the exact 502 JSON envelope with no internal details.

## 7. Learning Objective

- Implement a fixed-upstream reverse proxy with a pinned exact path mapping, exact Host policy, exact forwarding-header policy, exact hop-by-hop header policy with `Trailer` and `Proxy-Connection` and a stripped `TE`, exact response allowlist, exact public 502 envelope on pre-response transport failure, no retry for any method, and tests that pin every branch through `httptest` without public network.

## 8. Functional Requirements

1. The proxy uses `net/http/httputil.ReverseProxy`.
2. The upstream is one origin-only `http` or `https` URL validated at construction: scheme and host required, no userinfo, query, fragment, or non-root path, trailing root normalized; invalid configuration is a startup error. The public prefix is exactly `/api`, and no request can choose a target.
3. Path mapping is exact: an inbound path of exactly `/api` maps to upstream root `/`; an inbound path of `/api/` followed by a safe suffix maps to upstream root followed by that safe suffix; any other inbound path returns 404 with no upstream call.
4. A safe suffix contains one or more nonempty segments separated by literal `/`; each segment uses only the enumerated ASCII path characters and is not exactly `.` or `..`.
5. Any percent-encoding, backslash, empty interior segment, or forbidden segment is rejected with 400 before proxying; an off-boundary path returns 404.
6. Method, query string bytes, and query semantics are preserved verbatim under the path mapping.
7. The outbound URL and Host come only from the configured upstream; the inbound `Host` is not forwarded.
8. All inbound forwarding headers are stripped, including `Forwarded`, every `X-Forwarded-*`, and legacy target-influencing headers.
9. If and only if the parsed direct peer IP from `Request.RemoteAddr` is exactly in the configured trusted peer set, exactly one outbound `X-Forwarded-For` is synthesized from that IP, exactly one outbound `X-Forwarded-Host` is synthesized from the validated inbound `Host`, and exactly one outbound `X-Forwarded-Proto` is synthesized as `https` or `http` from inbound TLS presence; otherwise no `X-Forwarded-*` headers are sent.
10. The hop-by-hop header removal list is `Connection`, `Keep-Alive`, `Proxy-Authenticate`, `Proxy-Authorization`, `Proxy-Connection`, `Transfer-Encoding`, `Upgrade`, `Trailer` (not `Trailers`), and any header named in a `Connection` header value; `TE` is stripped completely for this project.
11. Request context cancellation propagates through the injected transport.
12. Injected transport defaults are 2-second dial, 2-second TLS handshake, 5-second response header, and 30-second idle connection; inbound context cancellation and earlier deadlines still win.
13. Pre-response transport failures, including dial failure and response header timeout, return 502 with the pinned JSON envelope and no internals.
14. Upstream HTTP 4xx and 5xx responses are forwarded with their own status through the response header allowlist, never converted to 502.
15. Successful response bodies stream from upstream; for every upstream response it receives, `ReverseProxy` closes the body after copying completes or aborts. Pre-response dial failures have no body to close.
16. The response header allowlist is exactly `Content-Type`, `Content-Length`, `Content-Encoding`, `Cache-Control`, `ETag`, `Last-Modified`, `Retry-After`, and `Vary`. All other upstream response headers and all hop-by-hop headers are stripped; upstream `Location` is not forwarded.
17. The proxy sets exactly `Via: 1.1 tutorial-proxy` on every successful response and adds no invented request identifier, server, or other header.
18. The 502 envelope is exactly status 502, `Content-Type: application/json`, and a JSON object whose only field is `error` with value `bad_gateway`, with no trailing newline.
19. The proxy does not retry for any method. The proxy does not cache, does not load balance, does not handle WebSocket, and does not discover upstream dynamically.
20. Inbound client cancellation cancels the outbound request through the injected transport and triggers cleanup; the cancellation is not presented as a synthetic 502.
21. The proxy does not assert a panic-recovery contract; recovery in the proxy path is not part of the requirement.
22. Tests use `httptest` for the backend and the proxy listener; no public network, no fixed public port.

## 9. Inputs and Outputs

### Interface Contract

- Proxy input is an `http.Request` on the listener.
- Configuration input is the upstream URL, the pinned prefix `/api`, the trusted peer set, and the injected transport with its timeouts.
- Proxy output is a successful streamed response with the response header allowlist plus the fixed `Via`, a 404 for a path outside the public boundary, a 400 for a suspicious or malformed in-boundary path, or a 502 with the pinned JSON envelope for pre-response transport failure.
- Upstream 4xx and 5xx are forwarded through to the client with their own status.
- The backend is exercised in tests through `httptest`.

## 10. Rules and Edge Cases

- An inbound path exactly `/api` maps to upstream root `/`.
- An inbound `/api/` followed by one or more safe path segments maps to upstream root plus those segments, preserving literal `/` separators.
- Any path outside the boundary returns 404 without an upstream call.
- Percent-encoding, backslashes, empty interior segments, and `.` or `..` segments are rejected with 400 before proxying.
- The outbound URL and Host come only from configured upstream; the inbound `Host` is not forwarded.
- All inbound forwarding headers are stripped.
- A peer in the trusted set produces exactly one outbound `X-Forwarded-For`, `X-Forwarded-Host`, and `X-Forwarded-Proto`; otherwise none are sent.
- Hop-by-hop headers are removed per the pinned list including `Trailer` and `Proxy-Connection`; `TE` is stripped completely. `Connection` is parsed and any header named in `Connection` is also removed.
- Only the exact response allowlist is forwarded; `Location` and all other headers are stripped.
- The proxy sets `Via: 1.1 tutorial-proxy` and no other invented header.
- A pre-response transport failure returns the exact 502 envelope with no trailing newline or internals.
- For every response it receives, `ReverseProxy` closes the body after copying completes or aborts; a pre-response dial failure has no body to close.
- Upstream 4xx and 5xx are forwarded with their own status.
- Inbound client cancellation cancels the outbound request and is not presented as a synthetic 502.
- The proxy never retries for any method.

## 11. Project Constraints

- Single validated origin-only upstream URL.
- No open proxy.
- No public network.
- No Docker.
- No environment variables for the upstream URL.
- The transport is injected; tests inject shorter durations.
- No WebSocket.
- No cache.
- No balance.
- No dynamic discovery.
- No retry for any method.
- No panic-recovery contract.

## 12. Design Questions Before Coding

- How is the exact `/api` mapping computed so that the boundary is exhaustive and so that off-boundary paths produce 404 without an upstream call?
- How is a safe suffix validated so that any suspicious byte is rejected before proxying?
- How is the outbound `URL` and `Host` enforced inside the modern rewrite boundary so the inbound `Host` cannot influence them?
- How is the trusted peer set checked so that a request through a non-trusted peer produces no `X-Forwarded-*` headers at all?
- How is `Connection` parsed so that any header listed in `Connection` is also removed?
- How is `Trailer` distinguished from `Trailers` and how is `TE` completely stripped for this project?
- How is the response allowlist applied so that arbitrary upstream headers and hop-by-hop headers are stripped before responding?
- How is the 502 envelope shaped so it never includes internals, and how is `ErrorHandler` used to surface pre-response transport failures?
- How is pre-response transport failure distinguished from in-flight cancellation so that a 502 does not retroactively replace an already-written status?
- How is the outbound cancellation observed through the injected transport without presenting it as a synthetic 502?

## 13. Implementation Milestones

1. Define origin-only upstream validation, prefix `/api`, segment-aware safe suffix rules, trusted peer set, and injected transport.
2. Define the exact path mapping inside the modern rewrite boundary that sets the outbound `URL` and `Host`.
3. Define segment-aware safe suffix validation, suspicious in-boundary 400, and off-boundary 404, all without an upstream call on rejection.
4. Define the inbound forwarding header stripping and the exact trusted-peer synthesis for `X-Forwarded-For`, `X-Forwarded-Host`, and `X-Forwarded-Proto`.
5. Define the hop-by-hop header removal list including `Trailer`, `Proxy-Connection`, and a complete `TE` strip, with `Connection` parsing.
6. Define context cancellation propagation through the injected transport.
7. Define the exact response header allowlist and `Via: 1.1 tutorial-proxy`; strip everything else, including upstream `Location`.
8. Define `ErrorHandler` for pre-response transport failures that yields the pinned 502 envelope; ensure upstream 4xx and 5xx are forwarded through to the client.
9. Define the pinned 502 envelope as status 502, `application/json`, a JSON object with only `error: "bad_gateway"`, and no trailing newline.
10. Define the `httptest` backend and proxy listener wiring for the test matrix.

## 14. Verification Cases the Learner Must Write

### Required Cases

- An inbound exact `/api` is rewritten to upstream `/`; the query bytes are preserved verbatim.
- Construction rejects an upstream without `http` or `https` scheme or host and rejects userinfo, query, fragment, or non-root path.
- An inbound `/api/items` is rewritten to upstream `/items`; the method, query bytes, and query semantics are preserved verbatim.
- An inbound `/api/v1.0/items` is rewritten to upstream `/v1.0/items`.
- An inbound path that is not under the prefix returns 404 and does not call upstream.
- An inbound `/api/../etc` returns 400 with no upstream call.
- An inbound `/api/%2fetc` returns 400 with no upstream call.
- An inbound `/api/items` carrying `X-Forwarded-For: 9.9.9.9` from a loopback direct peer yields exactly one outbound `X-Forwarded-For` containing the loopback peer; the inbound value is not preserved or chained.
- An inbound `/api/items` from a non-trusted peer yields no `X-Forwarded-*` outbound headers.
- An inbound `/api/items` carrying `Forwarded: for=9.9.9.9` yields a stripped `Forwarded` and no outbound forwarding chain from it.
- An inbound request with legacy target-influencing headers such as `X-Original-URL` has them stripped before the outbound request.
- The outbound `Host` is the configured upstream host; the inbound `Host` is not forwarded.
- The pinned hop-by-hop list is removed, including `Trailer` and `Proxy-Connection`; `TE` is stripped completely.
- A `Connection: close, X-Custom` request results in `X-Custom` removed as a header named in `Connection`.
- The response header allowlist is exactly `Content-Type`, `Content-Length`, `Content-Encoding`, `Cache-Control`, `ETag`, `Last-Modified`, `Retry-After`, and `Vary`; `Location`, arbitrary upstream headers, and hop-by-hop headers are stripped; `Via` is exactly `1.1 tutorial-proxy`.
- Upstream 2xx, 3xx, 4xx, and 5xx responses are forwarded with their own status through the allowlist; an upstream 5xx is not converted to 502.
- An upstream dial failure returns 502 with the pinned JSON envelope; the envelope is status 502, `Content-Type: application/json`, body contains only `error: "bad_gateway"`.
- An upstream response header timeout returns 502 with the pinned JSON envelope.
- Pre-response transport failure is reported through `ErrorHandler`; there is no upstream response body to close because no body was received.
- Inbound context cancellation cancels the outbound transport; the cancellation is not presented as a synthetic 502.
- Failure after downstream streaming has begun terminates or cancels the in-flight response without rewriting the already-written status to 502.
- Successful response bodies stream from upstream; for every upstream response it receives, `ReverseProxy` closes the body after copy success or abort, including inbound cancellation.
- The proxy never retries for any method.
- Concurrent requests through the proxy do not race; the test runs under the race detector.
- No public network is contacted; the upstream is `httptest`.

## 15. Common Mistakes to Watch For

- Treating the proxy as an open proxy, accepting an arbitrary target from a request, forwarding the inbound `Host`, trusting inbound `Forwarded` or `X-Forwarded-*`, chaining inbound forwarding values with synthesized values, sending `X-Forwarded-*` for an untrusted peer, listing `Trailers` instead of `Trailer` in the hop-by-hop list, omitting `Proxy-Connection` from the defensive removal, forwarding `TE`, missing hop-by-hop headers named in `Connection`, forwarding arbitrary upstream response headers instead of pinning the allowlist, converting an upstream 4xx or 5xx to a 502, leaking internals in the 502 envelope, attempting to replace an already-written status with 502 on later failure, inventing a request identifier or server header, retrying for any method, attempting to cache responses, attempting to load balance, attempting WebSocket upgrades, attempting dynamic upstream discovery, using a fixed public port for tests, asserting a panic-recovery contract, and using sleep to synchronize tests.

## 16. Topics and References for Study

- Study `net/http/httputil.ReverseProxy`, the modern rewrite boundary that sets outbound `URL` and `Host`, hop-by-hop header rules per the HTTP specification including the exact `Trailer` name and `Proxy-Connection`, the `X-Forwarded-*` family and its trust model, request body streaming, response body close, context cancellation propagation, transport-level timeouts, idempotent versus non-idempotent methods, the exact 502 envelope contract, the response header allowlist, `ErrorHandler` semantics for pre-response transport failure, and the fact that recovery of an already-written status with 502 is not a supported behavior.
- Review the Go `net/http`, `httputil`, and `httptest` documentation.
- Read the prior README for Project 077 as the immediate predecessor for protocol and connection identity, Project 071 for TCP framing and protocol error discipline, and Project 060 for graceful server shutdown and lifecycle ownership.

## 17. Self-Assessment Questions

1. Why is the configured public prefix exactly `/api` and why is the mapping exact rather than a general rewrite?
2. Why is the proxy not an open proxy and how is that proved by tests for arbitrary-target and inbound-influence attempts?
3. Why is the outbound `URL` and `Host` fixed to the configured upstream and why is the inbound `Host` never forwarded?
4. Why are all inbound forwarding headers stripped before any synthesis, and why is the trusted synthesis limited to exactly one header of each name from the trusted peer only?
5. Why is `Trailer` correct and `Trailers` incorrect, why is `Proxy-Connection` defensively included, and why is `TE` stripped completely for this project?
6. Why is the response header allowlist pinned rather than allow arbitrary forwarding?
7. Why is upstream HTTP 4xx and 5xx a forwarded response and not a 502, and why is pre-response transport failure distinct from later in-flight failure?
8. Why does inbound client cancellation cancel the outbound request without being presented as a synthetic 502, and why does failure after downstream streaming has begun not rewrite an already-written status with 502?
9. Why does the proxy never retry for any method?

## 18. Definition of Completion

- [ ] The proxy uses `httputil.ReverseProxy` with one construction-validated origin-only upstream; invalid scheme, missing host, userinfo, query, fragment, or non-root path is a startup error.
- [ ] The configured public prefix is exactly `/api`. Exact `/api` and `/api/` followed by one or more safe segments are the only proxied paths; paths outside the boundary return 404 without an upstream call.
- [ ] Any percent-encoding, backslash, empty interior segment, or `.` or `..` segment in an in-boundary path returns 400 without an upstream call.
- [ ] Method, query bytes, and query semantics are preserved verbatim under the path mapping.
- [ ] The outbound `URL` and `Host` come only from the configured upstream; the inbound `Host` is not forwarded.
- [ ] All inbound forwarding headers including `Forwarded` and every `X-Forwarded-*` and legacy target-influencing headers are stripped.
- [ ] For an exactly trusted peer, exactly one outbound `X-Forwarded-For`, `X-Forwarded-Host`, and `X-Forwarded-Proto` are synthesized; otherwise none are sent.
- [ ] Hop-by-hop headers are removed per the pinned list including `Trailer` and `Proxy-Connection`; `TE` is stripped completely; `Connection`-named headers are also removed.
- [ ] Context cancellation propagates through the injected transport with dial, response header, and idle timeouts.
- [ ] Pre-response transport failure returns 502 with the pinned JSON envelope; upstream 4xx and 5xx are forwarded with their own status; later in-flight failure does not rewrite an already-written status with 502.
- [ ] The 502 envelope is exactly status 502, `Content-Type: application/json`, a JSON object with only `error: "bad_gateway"`, and no trailing newline.
- [ ] Successful response bodies stream; for every upstream response it receives, `ReverseProxy` closes the body after copying completes or aborts; pre-response dial failures have no body to close.
- [ ] The response allowlist is exactly `Content-Type`, `Content-Length`, `Content-Encoding`, `Cache-Control`, `ETag`, `Last-Modified`, `Retry-After`, and `Vary`; `Location` and everything else are stripped.
- [ ] The proxy sets exactly `Via: 1.1 tutorial-proxy` on every successful response and adds no invented request identifier, server, or other header.
- [ ] Inbound client cancellation cancels the outbound request and is not presented as a synthetic 502.
- [ ] The proxy does not retry for any method, does not cache, does not load balance, does not handle WebSocket, and does not discover upstream dynamically.
- [ ] No panic-recovery contract is asserted.
- [ ] Tests use `httptest` for backend and proxy listener; no public network, no fixed public port.
- [ ] All tests pass under the race detector with no sleep synchronization.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions

- Add a per-upstream latency histogram exposed at shutdown for capacity planning tests.
- Add a structured access log that records method, path, status, and duration but never request or response bodies.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 077 — HTTP/2 Push Server](../../06-networking/077_http2_push_server/README.md#20-prerequisite-based-documentation-guide), [Project 071 — TCP Echo Server](../../06-networking/071_tcp_echo_server/README.md#20-prerequisite-based-documentation-guide), [Project 060 — Graceful Shutdown Web](../../04-apis-and-services/060_graceful_shutdown_web/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`net/http/httputil`](https://pkg.go.dev/net/http/httputil).

### Project-specific learning focus

- **Learn now:** safe outbound rewrites, Host and forwarding trust, hop-by-hop header removal, streaming bodies, cancellation, transport timeouts, response allowlists, and pre-response failures.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
