# Project 074 — WebSocket Live Chat

## 1. Project Name and Number

- Project 074, websocket_live_chat.
- This README is a learning guide only.
- You will create every source and test file yourself in `06-networking/074_websocket_live_chat/`.
- This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea

A WebSocket chat with a hub, a read pump, and a write pump per connection. The maintained WebSocket library is `github.com/coder/websocket` pinned to version `v1.8.15`. The first text message is the username: 1..24 ASCII letters, digits, underscore, or hyphen, with ASCII case-insensitive uniqueness and the accepted display spelling preserved. Later text messages are chat bodies: valid UTF-8, 1..4,096 bytes inclusive, and no LF byte; CR is allowed as content. The hub owns broadcast order and outbound queues. Bounded outbound queues and a lifecycle owner protect the server from slow clients. A server-initiated heartbeat keeps idle connections alive. The wire notice formats match Project 072 exactly.

## 3. Why This Project Now?

- This project requires Project 073 (udp_broadcast_server), Project 071 (tcp_echo_server), and Project 060 (graceful_shutdown_web).
- It applies prior network lifecycle and shutdown discipline to a connection-oriented WebSocket transport, which begins through an HTTP upgrade, with a library-managed handshake, framing, and heartbeat.
- It also raises the bar on origin validation, message size discipline, close-mapping consistency, and lifecycle ownership between the hub and the per-connection writer.

## 4. Prerequisites

- Projects 073, 071, and 060 are required prerequisites because WebSocket begins through an HTTP upgrade and depends on network lifecycle and graceful shutdown fundamentals.
- Project 072 is optional review only for the hub pattern and exact wire notice contract.
- The dependency `github.com/coder/websocket` is pinned at version `v1.8.15`.
- The tests need only `httptest` and an injected heartbeat boundary.
- No public network, no Docker, no environment variables.

## 5. What You Must Know Before Starting

- Know the WebSocket handshake, the difference between text and binary frames, the difference between a clean and an abnormal close, the coder/websocket library's framing, close, and ping behavior, the hub pattern from Project 072, lifecycle signaling between the hub and a per-connection writer, and the race detector.

## 6. Explanation of New Concepts

### Concepts

- The library is `github.com/coder/websocket` at version `v1.8.15`.
- This README does not implement WebSocket framing; the library handles fragmentation, reassembly, and frame boundaries.
- The library's public API for reading, writing, pinging, and closing is used exactly as documented.
- The application message limit is 4,096 bytes inclusive for each complete library-reassembled message; oversize detection happens on the reassembled message before application parsing.

- The first text message is interpreted as the username.
- A username is 1..24 bytes and contains only ASCII letters, ASCII digits, underscore, or hyphen.
- Uniqueness is compared with ASCII case folding, while the accepted spelling is preserved for display.
- Every later text message is a chat body: valid UTF-8, 1..4,096 bytes inclusive, and containing no LF byte.
- A CR byte is permitted as content.
- Disallowing LF keeps the one-line `<name>: <msg>` plus LF broadcast format unambiguous.

- Origin policy is injected as exact origins.
- The match is scheme plus host plus explicit port.
- No wildcard.
- No suffix matching.
- A missing, malformed, opaque, or non-allowlisted Origin is rejected before upgrade.
- Tests use exactly one allowed httptest origin.

- Close mapping is pinned.
- Any text or binary message with payload greater than 4,096 bytes is closed with 1009 before application parsing; a payload of exactly 4,096 bytes passes the size gate.
- A text message with invalid UTF-8 is closed with 1007.
- A binary frame within the size limit is closed with 1003.
- A valid-UTF-8 message whose application content is an invalid or duplicate username, an empty chat message, a chat body containing LF, or any other disallowed message policy is closed with 1008.
- A username longer than 24 bytes but within the 4,096-byte library-reassembled message limit is closed with 1008.
- A normal peer close is 1000.
- Server shutdown is 1001.
- An unexpected internal failure is 1011.
- Close reasons are stable, generic, and within protocol length.
- No internals are leaked.

- The wire notice formats embed the exact Project 072 contract.
- A join notice is the literal prefix `[hub] ` followed by the accepted display name followed by the literal word ` joined` followed by the LF byte.
- A leave notice is the literal prefix `[hub] ` followed by the accepted display name followed by the literal word ` left` followed by the LF byte.
- An accepted chat message is the accepted display name followed by the literal `: ` followed by the exact message followed by the LF byte.
- A pre-registration invalid or duplicate username rejection is the literal `[hub] username rejected` followed by the LF byte.
- The sender is included in join and chat events.
- No history is replayed.
- A sole client disconnecting may emit a leave notice at the hub level, but there are no remaining recipients, so the leave event has no observable wire recipient for that case.

- The architecture is one reader and one writer per connection, plus one lifecycle owner.
- The hub owns registration, unregistration, and outbound queues.
- The reader never writes frames.
- The writer is the sole owner of data writes, close writes, and server heartbeat pings.
- Reader operations and writer operations use contexts and deadlines so shutdown cannot hang.

- Queue-full behavior is explicit.
- The outbound queue capacity is exactly 32 complete events.
- When the hub finds a client's outbound queue full during a nonblocking enqueue, the hub marks the client unregistered, closes that client's outbound queue exactly once, and signals the client's lifecycle owner.
- The lifecycle owner cancels or closes the connection so the reader and writer stop without requiring the writer to drain a full queue.
- Other clients receive the current broadcast event, then exactly one leave notice is ordered for the remaining membership.
- The hub performs no socket I/O.
- Repeated slow-client teardown signals for the same client are idempotent.

- Slow-client close control uses the generic policy close 1008 when it can be delivered under a bounded writer deadline.
- If the peer is unresponsive and the close cannot be delivered within the bounded writer deadline, the writer force-closes the connection without promising that the peer observes a WebSocket close code.

- The heartbeat is the coder/websocket library Ping operation issued by the sole writer under the 10-second ping-wait context while the reader remains active so the pong can be processed.
- The production defaults are configurable and validated: ping every 30 seconds with a 10-second ping-wait timeout.
- Both intervals must be positive and the timeout must be below the interval.
- A ping timeout or ping failure cancels or force-closes the connection and unregisters the client once; the server does not promise to deliver a WebSocket close code to an unresponsive peer after a force-close.
- Tests use an injected heartbeat trigger or clock; no sleep proves liveness.

- Server shutdown asks the writer to send close 1001 under a bounded deadline, then force-closes the connection if needed, and waits for pumps.
- Shutdown does not hang waiting for an unresponsive peer.

- Text-only protocol examples are permitted.
- As a prose shape: a client opens a WebSocket to the server with an allowed Origin.
- The first text message is `alice` followed by no terminator.
- The server validates the username and registers the client.
- The server then sends `[hub] alice joined` followed by LF to every registered client including the new one.
- The client then sends `hello there`; the server sends `alice: hello there` followed by LF to every registered client including the sender.
- On disconnect the server sends `[hub] alice left` followed by LF to the remaining members.
- A client that sends a binary frame within the size limit is closed with close code 1003.
- A client that sends a text frame whose reassembled payload is greater than 4,096 bytes is closed with 1009 before application parsing.
- A client with a disallowed Origin is rejected before upgrade.

## 7. Learning Objective

- Implement a deterministic WebSocket chat with library-pinned version, exact origin policy, exact close mapping, exact 072 wire notice formats, exact queued slow-client lifecycle, heartbeat discipline that uses injected boundaries for tests rather than sleep, and shutdown that cannot hang on an unresponsive peer.

## 8. Functional Requirements

1. WebSocket library is `github.com/coder/websocket` at version `v1.8.15`.
2. Library's public framing, close, and ping behavior is used exactly; framing is not hand-rolled; the library handles fragmentation.
3. Application message limit is 4,096 bytes inclusive for each complete library-reassembled message; exactly 4,096 bytes passes the size gate.
4. Any text or binary message with payload greater than 4,096 bytes is closed with 1009 before application parsing.
5. A text message with invalid UTF-8 is closed with 1007.
6. A binary frame within the size limit is closed with 1003.
7. The first text message is a username of 1..24 ASCII letters, digits, underscore, or hyphen; uniqueness is ASCII case-insensitive and accepted display spelling is preserved. Later text messages are valid-UTF-8 chat bodies of 1..4,096 bytes inclusive with no LF byte; CR is allowed. A valid-UTF-8 message that violates these application rules is closed with 1008.
8. A username longer than 24 bytes but within the 4,096-byte library-reassembled message limit is closed with 1008.
9. Origin policy is injected as exact origins; match is scheme plus host plus explicit port; no wildcard; no suffix matching.
10. Missing, malformed, opaque, or non-allowlisted Origin is rejected before upgrade.
11. Close mapping is pinned: 1003 binary within limit, 1007 invalid UTF-8 text, 1008 valid-UTF-8 invalid or duplicate name or empty or disallowed message, 1009 message too big, 1000 normal peer close, 1001 server shutdown, 1011 unexpected internal failure.
12. Close reasons are stable, generic, within protocol length; no internals.
13. One reader and one writer per connection plus one lifecycle owner.
14. Reader never writes frames.
15. Writer is the sole owner of data writes, close writes, and server heartbeat pings.
16. Reader and writer operations use contexts and deadlines so shutdown cannot hang.
17. Outbound queue capacity is exactly 32.
18. Queue-full: hub marks the client unregistered, closes outbound queue once, signals lifecycle owner; reader and writer stop without draining the full queue; other clients receive the current event; exactly one leave notice is ordered for remaining membership.
19. Repeated slow-client teardown signals for the same client are idempotent.
20. Hub performs no socket I/O.
21. The wire notice formats match Project 072 exactly: `[hub] <name> joined` LF, `[hub] <name> left` LF, `<name>: <msg>` LF, `[hub] username rejected` LF; sender included in join and chat; no history is replayed.
22. A sole client disconnecting may emit a leave notice at the hub level with no observable wire recipient for that case.
23. Slow-client close control uses the generic policy close 1008 under a bounded writer deadline; on peer unresponsiveness the writer force-closes without promising the peer observes a close code.
24. Heartbeat is the coder/websocket library Ping from the sole writer under the 10-second ping-wait context while the reader remains active.
25. Production defaults: ping every 30 seconds, 10-second ping-wait timeout, both configurable and validated positive with timeout below interval.
26. Ping timeout or ping failure cancels or force-closes the connection and unregisters the client once; no promise to deliver a close code to an unresponsive peer.
27. Server shutdown asks the writer to send close 1001 under a bounded deadline, then force-closes if needed, and waits for pumps.
28. Tests use an injected heartbeat trigger or clock; no sleep proves liveness.
29. Tests use exactly one allowed httptest origin.
30. No history, no auth, no rooms, no persistence.

## 9. Inputs and Outputs

### Interface Contract

- Server input is a loopback HTTP listener address, an injected allowlist of exact origins, an injected ping interval, and an injected ping timeout.
- Codec input is a sequence of text and binary frames with library-managed reassembly.
- Codec output is queued events delivered to the writer.
- The wire contract is the exact Project 072 notice format set.

## 10. Rules and Edge Cases

- A client with a disallowed Origin is rejected before upgrade.
- A text or binary frame with reassembled payload greater than 4,096 bytes is closed with 1009 before application parsing; exactly 4,096 bytes passes the size gate.
- A binary frame within the size limit is closed with 1003.
- A text frame with invalid UTF-8 is closed with 1007.
- A username outside the 1..24-byte ASCII letters, digits, underscore, or hyphen rule is closed with 1008.
- A duplicate under ASCII case-insensitive comparison is closed with 1008 while accepted display spelling is preserved.
- An empty chat message or a valid-UTF-8 chat body containing LF is closed with 1008; CR remains allowed content.
- A slow client is dropped under the queue-full policy with lifecycle signaling and without affecting other clients.
- A sole client disconnecting may emit a leave notice at the hub level with no observable wire recipient.
- Server shutdown closes with 1001 under a bounded deadline and force-closes if needed; shutdown does not hang.
- A ping timeout or ping failure cancels or force-closes and unregisters once.
- Repeated unregister requests are idempotent.
- Repeated slow-client teardown signals for the same client are idempotent.
- The ping interval must be positive and the timeout must be below the interval.

## 11. Project Constraints

- Loopback only.
- The dependency is `github.com/coder/websocket` at version `v1.8.15`.
- The repository must contain no generated code in this README; tool installation and code generation belong to the learner.
- No history, no auth, no rooms, no persistence, no private messages.

## 12. Design Questions Before Coding

- How is the Origin header parsed into scheme, host, and explicit port without using a wildcard or suffix match?
- How is the first text message separated from later messages without state in the library?
- How is the 4,096-byte application message size enforced on the library-reassembled message before application parsing?
- How is the broadcast contract shared with Project 072 preserved exactly under the WebSocket transport?
- How is the outbound queue of capacity 32 mirrored from Project 072 with the same drop policy?
- How does the lifecycle owner stop the reader and writer without requiring the writer to drain a full queue?
- How is the slow-client close control 1008 delivered under a bounded writer deadline, and how is force-close distinguished from a delivered close?
- How is the unregister idempotent under both clean and abnormal read termination?
- How is the heartbeat boundary injected for tests without sleep?
- How does server shutdown bound the wait for an unresponsive peer?

## 13. Implementation Milestones

1. Define the origin allowlist with exact scheme, host, and explicit port matching.
2. Define the upgrade handler with origin rejection before handshake.
3. Define the codec with the exact username alphabet and case-folding rule, the valid-UTF-8 nonempty chat-body rule with LF forbidden and CR allowed, the 4,096-byte inclusive reassembled-message limit, and the pinned close mapping.
4. Define the wire notice format set matching Project 072 exactly.
5. Define the client structure with one reader, one writer, one lifecycle owner, and one outbound queue of capacity 32.
6. Define the hub with registration, nonblocking enqueue, slow-client teardown through the lifecycle owner, and post-registration violation handling.
7. Define the heartbeat as coder/websocket Ping from the sole writer under a 10-second ping-wait context with the reader active.
8. Define the close mapping with stable, generic, protocol-length reasons.
9. Define server shutdown that asks the writer to send close 1001 under a bounded deadline and force-closes if needed before waiting for pumps.
10. Define the `httptest` integration tests, race tests, and shutdown tests with no sleep.

## 14. Verification Cases the Learner Must Write

### Required Cases

- `go.mod` declares `github.com/coder/websocket` at version `v1.8.15`.
- Origin with allowed scheme, host, and explicit port is accepted; origin with disallowed host is rejected; origin with missing scheme is rejected; origin with wildcard attempt is rejected.
- First text message with a valid username is accepted and registers the client.
- First text message with a username longer than 24 bytes but within the 4,096-byte reassembled limit is closed with 1008.
- First text message with an invalid username within the size limit is closed with 1008.
- First text message with a duplicate case-insensitive username is closed with 1008.
- A text frame whose library-reassembled payload is greater than 4,096 bytes is closed with 1009 before application parsing; an exactly 4,096-byte valid chat body passes the size gate.
- A binary frame within the 4,096-byte limit is closed with 1003.
- A text frame with invalid UTF-8 within the 4,096-byte limit is closed with 1007.
- An empty chat message within the 4,096-byte limit is closed with 1008.
- A valid-UTF-8 chat message containing LF within the 4,096-byte limit is closed with 1008; a chat message containing CR but no LF is accepted.
- The join wire format is exactly `[hub] <name> joined` LF; the leave wire format is exactly `[hub] <name> left` LF; the chat wire format is exactly `<name>: <msg>` LF; sender is included in join and chat.
- No history is replayed to a new joiner.
- Two connected clients receive each other's accepted messages in hub order, including the sender.
- A full outbound queue triggers slow-client teardown: hub marks unregistered, closes outbound queue once, signals lifecycle owner, reader and writer stop without draining the full queue; other clients receive the current event, then exactly one leave notice.
- Repeated slow-client teardown signals for the same client are idempotent.
- Slow-client close control delivers 1008 under a bounded writer deadline; on peer unresponsiveness the writer force-closes and no promise is made that the peer observes a close code.
- Reader never writes frames; writer is the sole owner of data writes, close writes, and ping writes.
- Reader and writer operations use contexts and deadlines; shutdown cannot hang on an unresponsive peer.
- Ping interval and timeout are configurable and validated positive with timeout below interval.
- Heartbeat is coder/websocket Ping from the sole writer under the 10-second ping-wait context with the reader active; no separate invented pong callback is used.
- A simulated ping timeout cancels or force-closes and unregisters the client once; no promise is made to deliver a close code to the unresponsive peer.
- Server shutdown asks the writer to send close 1001 under a bounded deadline, force-closes if needed, and waits for pumps.
- A sole client disconnecting may emit a leave notice at the hub level with no observable wire recipient.
- Repeated unregister requests are idempotent.
- All tests use `httptest` and an injected heartbeat boundary; no test uses sleep to prove liveness.
- All tests pass under the race detector with no goroutine leak.

## 15. Common Mistakes to Watch For

- Using a wildcard or suffix for the Origin allowlist, hand-rolling WebSocket framing, leaking internals in close reasons, mapping empty or invalid messages to the wrong close code, applying 1008 to oversize payloads instead of 1009, applying 1009 before library reassembly, missing the timeout-below-interval validation, blocking the hub on a slow client, requiring the writer to drain a full queue, double-closing the outbound channel, leaking the reader or writer, using sleep for heartbeat tests, promising that an unresponsive peer observes a WebSocket close code after force-close, hanging server shutdown waiting for an unresponsive peer, inventing a separate pong callback, and forgetting the explicit port in the Origin match.

## 16. Topics and References for Study

- Study the WebSocket protocol, the coder/websocket library documentation including its Ping and close APIs, upgrade handshake, origin header parsing, close codes, structured concurrency, hub ownership, lifecycle signaling between hub and per-connection writer, and library-managed framing.
- Review Go's `net/http`, `httptest`, `context`, and the race detector.
- Read the prior READMEs for Projects 073, 071, and 060 as required foundations, and Project 072 as optional review for the exact wire notice format set and slow-client teardown lifecycle.

## 17. Self-Assessment Questions

1. Why is the origin matched by scheme, host, and explicit port with no wildcard or suffix?
2. Why is the 4,096-byte limit applied on the library-reassembled message before application parsing, with 1009 used for any oversize text or binary payload rather than 1008?
3. Why does 1008 cover valid-UTF-8 but invalid or duplicate name, empty chat, or disallowed message policy?
4. Why is the wire notice format set identical to Project 072?
5. Why is a slow client torn down through a lifecycle owner instead of requiring the writer to drain?
6. Why is the slow-client close 1008 delivered under a bounded writer deadline, and why does force-close not promise the peer observes a code?
7. Why is the heartbeat coder/websocket Ping from the sole writer under a 10-second context with the reader active, with no separate pong callback?
8. Why do ping timeout and server shutdown bound work without promising a delivered close code to an unresponsive peer?
9. Why are unregister idempotence and a hub-level leave with no observable recipient important for a sole disconnecting client?
10. How can tests prove heartbeat, queue-full lifecycle, and shutdown without sleep?

## 18. Definition of Completion

- [ ] `go.mod` declares `github.com/coder/websocket` at version `v1.8.15`; library framing is used and not hand-rolled.
- [ ] Origin allowlist is exact scheme, host, and explicit port; missing or non-allowlisted origin is rejected before upgrade.
- [ ] First text message is a 1..24-byte username using only ASCII letters, digits, underscore, or hyphen; uniqueness is ASCII case-insensitive and accepted display spelling is preserved. Later text messages are valid UTF-8, 1..4,096 bytes inclusive, contain no LF, and may contain CR; binary is unsupported at the application layer.
- [ ] Application message limit is 4,096 bytes inclusive for each complete library-reassembled message; oversize is closed with 1009 before application parsing.
- [ ] Close mapping matches the pinned codes and reasons; no internals are leaked.
- [ ] The wire notice format set matches Project 072 exactly; sender is included; no history is replayed.
- [ ] Reader never writes frames; writer is the sole owner of data writes, close writes, and ping writes.
- [ ] Reader and writer operations use contexts and deadlines; shutdown cannot hang on an unresponsive peer.
- [ ] Outbound queue capacity is exactly 32; queue-full triggers slow-client teardown through the lifecycle owner without draining the full queue.
- [ ] Repeated slow-client teardown signals and unregister requests are idempotent.
- [ ] Slow-client close control uses 1008 under a bounded writer deadline and force-closes on peer unresponsiveness without promising the peer observes a code.
- [ ] Heartbeat is coder/websocket Ping from the sole writer under a 10-second context with the reader active; ping timeout or failure cancels or force-closes and unregisters once.
- [ ] Server shutdown asks the writer to send close 1001 under a bounded deadline and force-closes if needed before waiting for pumps.
- [ ] Tests use `httptest` and an injected heartbeat boundary; no sleep proves liveness.
- [ ] No history, no auth, no rooms, no persistence.
- [ ] All tests pass under the race detector with no goroutine leak.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions

- Add a per-hub metric for accepted, rejected, oversize, and dropped-slow clients visible at shutdown.
- Add a structured close-code table that documents the mapping for tests and operators.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 073 — UDP Broadcast Server](../../06-networking/073_udp_broadcast_server/README.md#20-prerequisite-based-documentation-guide), [Project 071 — TCP Echo Server](../../06-networking/071_tcp_echo_server/README.md#20-prerequisite-based-documentation-guide), [Project 060 — Graceful Shutdown Web](../../04-apis-and-services/060_graceful_shutdown_web/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`github.com/coder/websocket`](https://pkg.go.dev/github.com/coder/websocket).
- **Standards and concept references:** [RFC 6455: WebSocket](https://www.rfc-editor.org/rfc/rfc6455.html).

### Project-specific learning focus

- **Learn now:** upgrade and origin checks, message framing, close codes, ping behavior, size limits, hub and writer ownership, slow clients, and structured connection teardown.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
