# Project 077 — HTTP/2 Push Server

## 1. Project Name and Number

- Project 077, http2_push_server.
- The directory name is historical; the title and scope of this project are HTTP/2 over TLS with multiplexed streams.
- This README is a learning guide only.
- You will create every source and test file yourself in `06-networking/077_http2_push_server/`.
- This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea

A TLS HTTP server whose ALPN list advertises `h2` first and `http/1.1` second, so that the required client negotiates `h2` and a deliberately HTTP/1.1-only trusted client succeeds as HTTP/1.1 on the health route. The server serves a health endpoint and a controllably blocked data endpoint. Tests prove that several concurrent requests share one HTTP/2 connection, that the connection identity is confirmed through `httptrace` rather than timing, and that graceful shutdown is observed by synchronization events. HTTP/2 server push is explicitly out of scope and is not the learning target.

## 3. Why This Project Now?

- This project requires Project 076 (grpc_streaming_api) as the immediate predecessor, Project 071 (tcp_echo_server) for TCP framing, idle deadlines, and per-connection protocol error discipline, and Project 060 (graceful_shutdown_web) for graceful server shutdown and lifecycle ownership.
- Project 046 (basic_http_server) is recommended review for `net/http` server configuration and is optional review only, not a formal prerequisite.
- This project extends the network discipline into HTTP/2: ALPN negotiation, TLS configuration, request body and header limits, and the conceptual difference between connection-wide and stream-level behavior.

## 4. Prerequisites

- Projects 076, 071, and 060 are required prerequisites.
- Project 076 is the immediate predecessor for gRPC streaming lifecycle.
- Project 071 is required for TCP connection handling, byte framing, idle deadlines, accept-loop shutdown, and per-connection protocol error discipline.
- Project 060 is required for graceful server shutdown and lifecycle ownership.
- Project 046 is recommended review for `net/http` server basics but is not a formal prerequisite.
- Test certificates are generated locally for test use only.
- No public network, no Docker, no production certificate authority.
- The learner manages TLS configuration and certificate generation through an appropriate tooling mechanism.

## 5. What You Must Know Before Starting

- Know the HTTP/1.1 request and response model, the HTTP/2 framing model at a conceptual level, TLS handshake and ALPN, certificate authorities and trust pools, server timeouts and their scope, the difference between connection-level and stream-level limits, `httptest.Server` and unstarted server lifecycle, the difference between protocol major version 1 and 2, graceful shutdown, and the race detector.

## 6. Explanation of New Concepts

### Concepts

- The historical directory name `http2_push_server` is preserved.
- The project title and scope are HTTP/2 over TLS with multiplexed streams.
- HTTP/2 server push is not the learning target.
- Server push via the `Pusher` interface is deprecated, is not broadly supported in browsers, and is intentionally excluded from required behavior and from tests.

- The server speaks TLS on a loopback address and its ALPN list is exactly `h2` then `http/1.1`.
- The minimum TLS version is TLS 1.2; the cipher configuration is the Go default for TLS 1.2 and TLS 1.3 with no user-supplied weakening.
- A generated test certificate is loaded at startup.
- The certificate is valid for `127.0.0.1` and for `localhost` over loopback.
- The required client negotiates `h2`; a deliberately HTTP/1.1-only trusted client that performs its own protocol negotiation succeeds as HTTP/1.1 on the health route because `http/1.1` appears second in the ALPN list.
- A separate client that trusts a different pool fails the handshake against the server, and that failure is observed in tests.
- The server is not described as requiring HTTP/2; the protocol selection is by ALPN, and HTTP/1.1 is permitted on the same listener.

- Routes include a health endpoint that responds with a fixed 200 status and a small JSON body, and a data endpoint that holds at an injected barrier until released.
- The data endpoint is the controllable block used to prove that one in-flight request does not prevent a sibling request from completing on the same HTTP/2 connection.
- The server runs an HTTP/2 transport that permits many concurrent streams per connection.

- A generic concurrent `http.Transport` may open additional connections when its internal configuration allows it, so tests constrain the transport to one connection for the origin.
- Because the transport uses a custom TLS trust pool, the required HTTP/2 transport explicitly enables HTTP/2 attempts rather than relying on automatic defaults.
- The test warms the connection with one request, then asserts through `httptrace` `GotConn` identity that every multiplexing request uses the same connection.
- The assertion is the connection object identity, not the `GotConn.Reused` flag or elapsed time.
- The separate fallback client is deliberately configured for HTTP/1.1 only.

- Two injected per-request barriers are used in the overlap test.
- The two data handlers each receive their own barrier through configuration and each reports that it has entered through a synchronization event.
- The test then releases the second barrier while the first barrier remains held and confirms through a synchronization event that the second handler completes while the first remains blocked.
- The test then releases the first barrier and confirms that the first handler completes.
- Overlap is established by the sequence of events, not by the order in which barriers happen to be released in arbitrary ways.

- Request body and header limits are pinned.
- The body limit is exactly 1 MiB inclusive and is enforced at the handler boundary; a request whose body exceeds 1 MiB returns 413 from the handler.
- Header limits have two distinct layers.
- First, the application aggregate request-header budget is exactly 16 KiB inclusive; a request whose aggregate header names and values exceed that budget and still reaches the handler returns 431 from the handler.
- Second, `MaxHeaderBytes` and the corresponding HTTP/2 decoded-header-list defense are each configured at 64 KiB.
- A request rejected by either transport defense before reaching the handler may yield a protocol-level or connection-level error rather than a 431, and the test does not promise a 431 on that pre-handler rejection.
- The two layers are observed independently.

- Production defaults are pinned: 5-second read-header timeout, 15-second read timeout, 15-second write timeout, 60-second idle timeout, and a 5-second graceful-shutdown budget.
- The learner explains that `net/http` read and write deadlines do not cleanly map to a single HTTP/2 stream under all conditions and may have connection-wide interactions.
- The required tests do not wait for these wall-clock timeouts to prove protocol behavior; handlers rely on request contexts and the two injected barriers.
- Graceful shutdown, connection identity, and overlap are proved by synchronization events rather than by assuming deadline scope.

- Graceful shutdown is pinned.
- The graceful-shutdown test starts shutdown while a handler is held at an injected barrier, proves by synchronization events that the shutdown remains pending, releases the barrier, and proves that the shutdown completes.
- A bounded watchdog force-closes only on test failure and then fails the test, so the fallback is never mistaken for a successful graceful stop.

- Safe test hooks record protocol and connection identity without depending on timing.
- Tests use `httptrace` callbacks for `GetConn`, `GotConn`, and `GotFirstResponseByte`; the connection object observed by `GotConn` is the client-side identity assertion.
- The response itself reports the negotiated HTTP major version.
- A concurrency-safe `ConnState` hook counts server-side accepted TCP connections and confirms that the single constrained connection eventually closes, but it is not used as a per-stream state machine because HTTP/2 streams are not individual `ConnState` connections.

- Text-only protocol examples are permitted.
- As a prose shape: a required `h2` client opens one connection to `https://127.0.0.1:port`, negotiates `h2` over ALPN, sends a request to `/healthz`, receives 200 with a small body, then sends two parallel requests to `/data` that each enter their own injected barrier.
- The test releases the second barrier while the first barrier is still held; the second handler completes while the first remains blocked.
- The test then releases the first barrier; the first handler completes.
- The same connection identity is observed for every request through `httptrace`.
- A deliberately HTTP/1.1-only trusted client connects and succeeds on the health route as HTTP/1.1 because `http/1.1` is in the ALPN list.
- A client that trusts a different pool fails the handshake and observes a TLS error.
- A required `h2` client that sends a request body of more than 1 MiB receives 413 from the handler.
- A required `h2` client that sends aggregate headers over the application budget and that reaches the handler receives 431 from the handler.

## 7. Learning Objective

- Implement a TLS HTTP/2 server with the pinned ALPN list, multiplexed streams, an exact overlap test through two injected barriers, exact body and header limit semantics, deterministic graceful shutdown observed through synchronization events, and tests that prove connection identity through `httptrace` rather than timing assumptions.
- Exclude HTTP/2 server push from required behavior.

## 8. Functional Requirements

1. The directory name remains `077_http2_push_server`; the project title and scope are HTTP/2 over TLS with multiplexed streams; HTTP/2 server push is excluded.
2. The server speaks TLS on a loopback address; the ALPN list is exactly `h2` then `http/1.1`.
3. The required client negotiates `h2`; a deliberately HTTP/1.1-only trusted client succeeds as HTTP/1.1 on the health route.
4. The server is not described as requiring HTTP/2; protocol selection is by ALPN with `http/1.1` permitted on the same listener.
5. Minimum TLS version is TLS 1.2; no user-supplied cipher weakening.
6. A generated test certificate is loaded at startup; the certificate is valid for `127.0.0.1` and `localhost` over loopback.
7. The client trusts the test CA through an explicit certificate pool; `InsecureSkipVerify` is not used as a solution.
8. Routes include a health endpoint and a controllably blocked data endpoint that holds at an injected per-request barrier.
9. The application aggregate request-header budget is exactly 16 KiB inclusive; greater requests that reach the handler return 431 from the handler.
10. `MaxHeaderBytes` and the corresponding HTTP/2 decoded-header-list defense are each 64 KiB; excess rejected before the handler may yield a protocol-level or connection-level error and has no promised 431.
11. The body limit is exactly 1 MiB inclusive and is enforced at the handler; over-limit returns 413 from the handler.
12. The required transport explicitly enables HTTP/2 attempts despite its custom TLS configuration, is constrained to one connection for the origin, and is warmed before the overlap test; `GotConn` object identity is asserted rather than `GotConn.Reused` or timing. The fallback transport permits only HTTP/1.1.
13. The overlap test uses two injected per-request barriers; both handlers report entry through synchronization events; releasing the second while the first remains held proves the second completes by event; releasing the first proves the first completes by event.
14. Production defaults are 5-second read-header timeout, 15-second read timeout, 15-second write timeout, 60-second idle timeout, and 5-second graceful-shutdown budget; the learner explains their possible connection-wide interactions and uses request contexts and injected barriers rather than timeout expiry to prove protocol behavior.
15. Graceful shutdown is context-driven; the test starts shutdown while a handler is held, proves pending by synchronization events, releases the handler, and proves completion; a bounded watchdog force-closes only on failure and then fails the test.
16. Tests use `httptest` unstarted servers with HTTP/2 enabled or loopback ephemeral TLS endpoints.
17. No browser; no public network; no HTTP/2 server push APIs.

## 9. Inputs and Outputs

### Interface Contract

- Server input is a loopback TLS address.
- Optional inputs are timeouts, body and header limits, ALPN protocols, and a context for shutdown.
- Server output is a bound address.
- Test observations are the HTTP major version on each response, the client-side connection identity from `httptrace`, the server-side accepted-connection count from `ConnState`, response status and body, and per-stream completion events synchronized through channels.

## 10. Rules and Edge Cases

- ALPN negotiates `h2` for the required client; the response reports HTTP major version 2.
- A deliberately HTTP/1.1-only trusted client succeeds as HTTP/1.1 on the health route because the ALPN list contains `http/1.1` second, and its response reports major version 1.
- A client that trusts a different CA pool fails the handshake.
- A request body of more than 1 MiB returns 413 from the handler.
- Aggregate headers greater than 16 KiB that reach the handler return 431.
- A header-list excess rejected by a 64 KiB transport defense before the handler may surface as a protocol-level or connection-level error and has no promised 431.
- A blocked handler does not block a sibling handler on the same HTTP/2 connection.
- Connection reuse is asserted through connection identity recorded by `httptrace`, not through `GotConn.Reused` alone or wall-clock timing.
- Graceful shutdown waits for in-flight handlers within its 5-second budget and force-closes after, observed by synchronization events.
- No HTTP/2 server push APIs are used.
- The watchdog force-closes only on test failure and then fails the test.

## 11. Project Constraints

- Loopback TLS only.
- Generated test certificate only; no production certificate authority.
- No public network.
- No browser.
- No HTTP/2 server push.
- The `Pusher` interface is excluded from required behavior.
- Reflection is disabled by default.
- No database, no auth.
- Tests use `httptest` unstarted servers with HTTP/2 enabled or loopback ephemeral TLS endpoints.

## 12. Design Questions Before Coding

- How is the ALPN list ordered so the required client observes `h2` and the HTTP/1.1-only client observes `http/1.1`?
- Why must HTTP/2 attempts be explicitly enabled when the transport has custom TLS configuration?
- How is the test CA separated so a wrong-pool client fails?
- How do two barriers prove overlap by events?
- How is the 16 KiB application header budget kept below the 64 KiB transport defenses?
- How is the 1 MiB body limit enforced at the handler?
- How is one connection enforced and identified?
- What does `ConnState` prove about TCP connections, and why is it not a stream state machine?
- How does graceful shutdown prove in-flight completion without sleep?

## 13. Implementation Milestones

1. Define the TLS configuration with minimum version TLS 1.2, ALPN list exactly `h2` then `http/1.1`, and the generated test certificate loader.
2. Define the test certificate generation path and the test CA pool.
3. Define the health and data handlers with the two injected per-request barriers on the data handlers.
4. Define the server with read header, read, write, and idle timeouts; the learner explains the connection-wide interaction in prose.
5. Define the body limit of exactly 1 MiB enforced at the handler, returning 413.
6. Define the 16 KiB inclusive application aggregate request-header budget, returning 431 when a greater request reaches the handler.
7. Define the 64 KiB `MaxHeaderBytes` and corresponding HTTP/2 decoded-header-list defense for pre-handler rejection with no promised 431.
8. Define graceful shutdown owned by a server context that closes the listener, waits for handlers within the grace period, and force-closes after, observed through synchronization events; the watchdog force-closes only on failure and then fails the test.
9. Define `httptest` unstarted server wiring with HTTP/2 enabled, a required transport with explicit HTTP/2 attempts and one-connection constraint, and a separate HTTP/1.1-only fallback transport.
10. Define `httptrace` hooks for client-side connection identity and a concurrency-safe `ConnState` hook for accepted TCP-connection count and eventual close, without treating it as stream identity.
11. Define the full matrix of protocol, ALPN, cert trust, concurrent stream overlap, blocked-and-released, 413 from the handler, 431 from the handler, pre-handler header defense, HTTP/1.1-only fallback, and shutdown tests.

## 14. Verification Cases the Learner Must Write

### Required Cases

- TLS server binds a loopback address and the required `h2` client completes the handshake.
- ALPN negotiated protocol for the required client is `h2`; the test asserts the protocol major version is 2.
- The deliberately HTTP/1.1-only trusted client succeeds as HTTP/1.1 on the health route; its response reports protocol major version 1.
- A client that trusts a different CA pool fails the handshake; the failure is observed as a TLS error.
- `InsecureSkipVerify` is not used as a solution.
- The required transport explicitly enables HTTP/2 attempts with its custom trust pool, constrains the origin to one connection, and warms it; every multiplexing request then asserts the same `GotConn` object identity.
- The assertion is connection identity recorded by callbacks, not `GotConn.Reused` alone, and not wall-clock timing.
- The two-barrier overlap test reports entry of both handlers through events, releases the second barrier while the first remains held, and proves by event that the second handler completes while the first is still blocked; then the first barrier is released and the first handler completes.
- Request body over 1 MiB returns 413 from the handler.
- Aggregate headers greater than 16 KiB that reach the handler return 431 from the handler; exactly 16 KiB passes the application budget.
- An excess rejected by the server parser or transport before the handler is observed as a protocol-level or connection-level error and the test does not promise a 431 on that path.
- The health endpoint returns 200 with the expected body and a fixed Content-Type.
- Graceful shutdown waits for in-flight handlers; the test proves pending by synchronization events, releases the handler, and proves completion; the watchdog force-closes only on failure and then fails the test.
- No goroutine per request leaks past shutdown.
- A concurrency-safe `ConnState` hook observes one accepted TCP connection for the constrained HTTP/2 test and its eventual close; it is not used to count or identify individual HTTP/2 streams.
- All tests pass under the race detector.
- No test uses sleep to wait for protocol correctness.
- No test reaches a browser or public network.
- No HTTP/2 server push APIs are exercised.

## 15. Common Mistakes to Watch For

- Using `InsecureSkipVerify` as a solution, weakening TLS configuration, omitting ALPN or accepting only HTTP/1.1, describing the server as requiring HTTP/2 while also advertising `http/1.1` in ALPN, allowing the test transport to open more than one connection and asserting reuse by coincidence, asserting reuse through `GotConn.Reused` alone, asserting reuse through wall-clock timing, allowing a blocked handler to block a sibling on the same connection, conflating the application header budget with `MaxHeaderBytes`, promising 431 on a pre-handler parser or transport rejection, promising a clean per-stream deadline scope under HTTP/2 that the runtime does not actually provide, treating graceful-shutdown completion as instant without synchronization events, having the watchdog force-close as a success path, and using HTTP/2 server push APIs.

## 16. Topics and References for Study

- Study HTTP/2 framing at a conceptual level, ALPN ordering with `h2` then `http/1.1`, TLS 1.2 minimum configuration, certificate generation and trust pools, application header budgets versus `MaxHeaderBytes` and HTTP/2 transport defenses, request body limits at the handler, server timeouts and their connection-wide interaction, graceful shutdown observed through synchronization events, `httptest` unstarted server lifecycle, `httptrace` callbacks, and `ConnState`.
- Review the Go `crypto/tls`, `net/http`, `httptest`, and HTTP/2 documentation.
- Read the prior README for Project 076 as the immediate predecessor for stream lifecycle, Project 071 for TCP framing and protocol error discipline, and Project 060 for graceful server shutdown and lifecycle ownership.
- Project 046 for `net/http` server basics is optional review.
- Note that HTTP/2 server push is deprecated and intentionally excluded.

## 17. Self-Assessment Questions

1. Why is the ALPN list exactly `h2` then `http/1.1`, and why does that allow the required `h2` client and a deliberately HTTP/1.1-only trusted client to both succeed?
2. Why is the server not described as requiring HTTP/2 when `h2` is the required path?
3. Why is the transport constrained to one connection for the origin and why is the reuse assertion based on connection identity through `httptrace` rather than on `GotConn.Reused` or timing?
4. Why are two injected barriers used and why is the overlap proved through synchronization events rather than through arbitrary reverse release?
5. Why is the application aggregate header budget set strictly below the parser and transport maximum, and why does the parser or transport pre-handler rejection have no promised 431?
6. Why is the body limit enforced at the handler rather than at the parser, and why does that guarantee a 413 from the handler?
7. Why are server timeouts explained as having connection-wide interactions and why do the required tests rely on handlers and barriers rather than assumed deadline scope?
8. Why does the graceful-shutdown watchdog force-close only on failure and then fail the test?

## 18. Definition of Completion

- [ ] The server speaks TLS with ALPN list exactly `h2` then `http/1.1` and minimum TLS version 1.2.
- [ ] A generated test certificate is loaded; the client trusts the test CA through an explicit pool.
- [ ] `InsecureSkipVerify` is not used as a solution.
- [ ] The required client negotiates `h2`; the deliberately HTTP/1.1-only trusted client succeeds as HTTP/1.1 on the health route.
- [ ] Routes include a health endpoint and a controllably blocked data endpoint with two injected per-request barriers.
- [ ] The two-barrier overlap test proves by synchronization events that one blocked handler does not prevent another sibling handler from completing on the same HTTP/2 connection.
- [ ] The body limit is exactly 1 MiB inclusive and returns 413 from the handler.
- [ ] The application aggregate header budget is 16 KiB inclusive and returns 431 when exceeded at the handler; the transport defenses are 64 KiB and a pre-handler rejection has no promised 431.
- [ ] The test constrains the transport to one connection for the origin and asserts connection identity through `httptrace` rather than `GotConn.Reused` alone or wall-clock timing.
- [ ] Production defaults are 5-second read-header, 15-second read, 15-second write, 60-second idle, and 5-second graceful-shutdown limits; their possible connection-wide interaction is explained honestly.
- [ ] Graceful shutdown waits for in-flight handlers within the grace period and force-closes after; the watchdog force-closes only on failure and then fails the test.
- [ ] `httptest` unstarted servers with HTTP/2 enabled or loopback ephemeral TLS endpoints are used.
- [ ] Responses report protocol major version, `httptrace` records client-side connection identity, and `ConnState` records accepted TCP-connection count and eventual close; no timing assumptions or per-stream `ConnState` claims.
- [ ] All tests pass under the race detector with no sleep synchronization.
- [ ] No browser or public network is contacted; no HTTP/2 server push APIs are exercised.
- [ ] No goroutine per request leaks past shutdown.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions

- Add a small structured access log that records method, path, status, and protocol major version but never request bodies or headers.
- Add a configurable connection-level concurrency cap exposed for capacity planning tests.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 076 — gRPC Streaming API](../../06-networking/076_grpc_streaming_api/README.md#20-prerequisite-based-documentation-guide), [Project 071 — TCP Echo Server](../../06-networking/071_tcp_echo_server/README.md#20-prerequisite-based-documentation-guide), [Project 060 — Graceful Shutdown Web](../../04-apis-and-services/060_graceful_shutdown_web/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`crypto/tls`](https://pkg.go.dev/crypto/tls), [`golang.org/x/net/http2`](https://pkg.go.dev/golang.org/x/net/http2).
- **Standards and concept references:** [RFC 9113: HTTP/2](https://www.rfc-editor.org/rfc/rfc9113.html), [RFC 7301: ALPN](https://www.rfc-editor.org/rfc/rfc7301.html).

### Project-specific learning focus

- **Learn now:** frames and streams, TLS negotiation, certificate trust, header and body budgets, timeout scope, connection state, graceful shutdown, and why server push is excluded.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
