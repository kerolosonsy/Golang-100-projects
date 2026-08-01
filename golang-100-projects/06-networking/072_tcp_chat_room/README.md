# Project 072 — TCP Chat Room

## 1. Project Name and Number
Project 072, tcp_chat_room. This README is a learning guide only. You will create every source and test file yourself in `06-networking/072_tcp_chat_room/`. This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea
A local line-based chat on a loopback address. The first line is the requested username with pinned length and character rules and unique case-insensitively. Subsequent lines are chat messages under a fixed limit. A hub goroutine owns membership, outbound queues, and broadcast ordering. Each client has exactly one writer owner; slow clients are torn down by the hub through a lifecycle owner without requiring the writer to drain a full queue.

## 3. Why This Project Now?
This follows Project 071 (tcp_echo_server). It reuses the LF-only framing and the per-connection lifecycle. It adds the hub pattern, broadcast ordering, slow-client teardown, and the discipline of one writer owner per connection with no socket I/O performed by the hub.

## 4. Prerequisites
Project 071 is the immediate predecessor and required prerequisite. The tests need only frozen time, fake clients for the hub, and an ephemeral loopback listener for the integration check. No public network, no Docker, no environment variables.

## 5. What You Must Know Before Starting
Know the line framing from Project 071, channels, goroutine ownership, the hub loop pattern, select with default for nonblocking sends, the difference between per-client slow-consumer drops and server-wide stalls, lifecycle signaling between hub and per-connection owners, and the race detector.

## 6. Explanation of New Concepts
The wire framing is the same as Project 071: LF only, no CRLF. A CR before LF is content. The username is 1..24 ASCII bytes, only letters, digits, underscore, and hyphen. No trimming or normalization beyond ASCII case-insensitive uniqueness. The accepted display spelling is preserved and used exactly as accepted on the wire. The chat message is 1..1,024 bytes excluding LF, must be valid UTF-8, and preserves spaces and content exactly. A CR before LF makes a username invalid because CR is not in the allowed character set, but a CR before LF may be valid message content if the rest of the message is valid UTF-8 and within the byte limit.

The wire notice formats are pinned exactly. A join notice is the literal prefix `[hub] ` followed by the accepted display name followed by the literal word ` joined` followed by the LF byte. A leave notice is the literal prefix `[hub] ` followed by the accepted display name followed by the literal word ` left` followed by the LF byte. An accepted chat message is the accepted display name followed by the literal `: ` followed by the exact message followed by the LF byte. A pre-registration invalid or duplicate username rejection is the literal `[hub] username rejected` followed by the LF byte. No alternative phrasing is permitted.

The hub goroutine owns the membership map, the per-client outbound queues, and the broadcast channel. The channel-receive order is authoritative. Join, leave, and chat events are enqueued in receive order. Once a client is registered, the hub enqueues a join notice for all registered clients including the joiner. Accepted chat becomes the exact chat format for all registered clients including the sender. An unregistered client never emits a leave. No history is replayed.

Each client has one reader goroutine, one writer owner, one outbound channel, and one lifecycle owner. The reader never writes to the socket directly. The writer is the sole owner of socket writes and close writes. The outbound channel capacity is exactly 32 complete events. The hub performs no socket I/O.

Slow-client teardown is implementable while preserving one writer. When the hub finds a client's outbound queue full during a nonblocking enqueue, the hub marks the client unregistered, closes that client's outbound queue exactly once, and signals the client's lifecycle owner. The lifecycle owner cancels or closes the connection so the reader and writer stop without requiring the writer to drain a full queue. The event that overflowed is not delivered to the disconnected client. The current broadcast continues to other clients before exactly one leave notice is ordered for the remaining membership. Repeated teardown for the same client is idempotent.

Message violations after registration are a protocol violation that disconnects and unregisters the offending client and produces exactly one leave notice for the remaining members. A post-registration violation covers empty message, invalid UTF-8 message, and over-1,024-byte message. Violations during pre-registration are handled by the rejection path below and never touch membership.

Pre-registration rejection delivery is best-effort through a pre-registration writer or control path with a bounded delivery deadline. The literal rejection format `[hub] username rejected` followed by LF is sent at most once. If the rejection cannot be delivered within the bounded deadline, the connection is closed without blocking the hub. Case-insensitive duplicate preserves the existing member. The display name is used exactly as accepted on the wire.

Shutdown stops new registration, unregisters all clients once, closes each outbound queue once, closes the listener and connections, and waits for pumps. Active connections are closed once.

Text-only protocol examples are permitted. As a prose shape: a client connects and sends `alice` followed by LF. The server registers the user and sends `[hub] alice joined` followed by LF to every registered client including the new one. The client then sends `hello there` followed by LF. The server sends `alice: hello there` followed by LF to every registered client including the sender. A second client sends `BOB` followed by LF while one `bob` is already registered; the server sends `[hub] username rejected` followed by LF at most once, then closes that second client without joining.

## 7. Learning Objective
Implement a deterministic, hub-owned chat room with exact username rules, exact pinned wire notice formats, exact outbound queue capacity, exact slow-client teardown lifecycle, exact post-registration violation behavior, and tests that pin event order, queue behavior, and shutdown without sleep-based synchronization.

## 8. Functional Requirements
1. Server binds to a configurable loopback address; production default is `127.0.0.1:0`.
2. Wire framing reuses LF-only from Project 071.
3. Username is 1..24 ASCII bytes, only letters, digits, underscore, and hyphen.
4. Username uniqueness is case-insensitive ASCII; the accepted display spelling is preserved on the wire.
5. No trimming or normalization beyond case-insensitive uniqueness.
6. Chat message is 1..1,024 bytes excluding LF, valid UTF-8, spaces preserved.
7. Empty or invalid or over-limit input before registration is a pre-registration protocol violation handled by the rejection path.
8. Empty, invalid UTF-8, or over-1,024-byte message after registration is a post-registration protocol violation that disconnects and unregisters that client and produces exactly one leave notice for the remaining members.
9. A CR before LF makes a username invalid; a CR before LF may be valid message content.
10. The hub goroutine owns membership, per-client outbound queues, and broadcast ordering.
11. Cross-client arrival order equals hub receive order.
12. Join notice is enqueued for all registered clients including the joiner, after registration.
13. Accepted chat becomes display-name, `: `, exact message, LF, for all including sender.
14. An unregistered client never emits a leave.
15. On disconnect after join, exactly one leave notice is enqueued to remaining clients.
16. No history is replayed.
17. Each client has a bounded outbound channel, capacity 32, exactly one writer owner, and exactly one lifecycle owner.
18. A nonblocking enqueue that finds a full queue causes the hub to mark the client unregistered, close that client's outbound queue exactly once, and signal the client's lifecycle owner; the lifecycle owner cancels or closes the connection so the reader and writer stop without draining the full queue.
19. The event that overflowed is not delivered to the disconnected client.
20. The current broadcast continues to other clients before exactly one leave notice is ordered for the remaining membership.
21. Repeated teardown for the same client is idempotent.
22. The hub performs no socket I/O.
23. Duplicate or invalid name rejection occurs before membership.
24. Pre-registration rejection is the literal `[hub] username rejected` followed by LF, sent at most once through a pre-registration writer or control path with a bounded delivery deadline.
25. If the rejection cannot be delivered within the bounded deadline, the connection is closed without blocking the hub.
26. Case-insensitive duplicate preserves the existing member.
27. Shutdown stops new registration, unregisters all clients once, closes each outbound queue once, closes the listener and connections, and waits for pumps.
28. The reader never writes to the socket directly.
29. The writer owner is the sole writer of socket data and close writes.
30. Codec and hub are independently testable without a real listener.

## 9. Inputs and Outputs
Server input is a TCP loopback address. Codec input is a username line followed by zero or more chat lines. Codec output is one of the pinned wire notice formats per line on the wire: join, message, leave, or pre-registration rejection. Server output is a bound address. Tests use fake clients that inject hub events and observe outbound channel deliveries.

## 10. Rules and Edge Cases
Empty username is rejected. Username longer than 24 bytes is rejected. Username with a disallowed character is rejected. A CR before LF makes a username invalid. Duplicate username case-insensitive is rejected without modifying the existing member. Empty message after registration is a post-registration violation that disconnects and unregisters. Invalid UTF-8 message after registration is a post-registration violation that disconnects and unregisters. Message longer than 1,024 bytes after registration is a post-registration violation that disconnects and unregisters. A full outbound queue during a nonblocking enqueue causes slow-client teardown through the lifecycle owner without requiring the writer to drain. Disconnect after join produces exactly one leave notice ordered for the remaining membership. Shutdown is ordered and idempotent. Repeated teardown for the same client is idempotent. Pre-registration rejection is sent at most once under a bounded delivery deadline and never blocks the hub.

## 11. Project Constraints
Loopback only. No public network. No TLS. No authentication. No history, no private messages, no rooms, no persistence. No real-time clock dependency. The repository must contain no generated code in this README; tool installation and code generation belong to the learner. The four pinned wire notice formats are exact and are not customizable.

## 12. Design Questions Before Coding
How is the membership map protected under concurrent registration and unregistration? How is the per-client outbound channel capacity exactly 32? How is a nonblocking enqueue implemented so the hub never blocks on a slow client? How does the lifecycle owner stop the reader and writer without requiring the writer to drain a full queue? How is the leave notice ordered after the rest of the current broadcast to other clients? How is teardown idempotent under repeated signals? How is registration closed before shutdown, and how is shutdown wait ensured to not lock on a slow pump? How is the display name preserved exactly on the wire? How is invalid-UTF-8 content detected without depending on the framing buffer's encoding? How does the bounded pre-registration delivery deadline guarantee the hub never blocks on rejection delivery?

## 13. Implementation Milestones
1. Define the line codec with username and message validation including the empty, invalid UTF-8, and over-limit cases.
2. Define the exact four pinned wire notice formats as a single internal contract shared by the codec, the hub, and the tests.
3. Define the client structure with one reader, one writer owner, one outbound channel, and one lifecycle owner.
4. Define the hub with ordered events, registration, nonblocking enqueue, slow-client teardown through the lifecycle owner, and post-registration violation handling.
5. Define the pre-registration rejection path with bounded delivery deadline and nonblocking close.
6. Define the server with the listener, accept loop, and shutdown orchestration.
7. Define the codec and hub tests with fake clients.
8. Define the integration tests on an ephemeral loopback port.
9. Run race detector and concurrency tests.

## 14. Verification Cases the Learner Must Write
- The join wire format is exactly the literal prefix `[hub] ` followed by the accepted display name followed by the literal word ` joined` followed by the LF byte, for all registered clients including the joiner.
- The leave wire format is exactly the literal prefix `[hub] ` followed by the accepted display name followed by the literal word ` left` followed by the LF byte, for remaining members only.
- The chat wire format is exactly the accepted display name followed by the literal `: ` followed by the exact message followed by the LF byte, for all registered clients including the sender.
- The pre-registration rejection wire format is exactly the literal `[hub] username rejected` followed by the LF byte, sent at most once.
- Codec validates a valid username and rejects an empty username, a 25-byte username, a username with a disallowed character, and a username with a CR before LF.
- Codec validates a valid UTF-8 message within the limit and rejects empty, over-limit, and invalid UTF-8 messages.
- A CR before LF is preserved as content in a valid message.
- Hub registers a user, sends a join notice to all registered clients including the joiner, and preserves the display spelling.
- Hub broadcasts a message in the exact chat format to all registered clients including the sender.
- Hub rejects a duplicate name case-insensitively without modifying the existing member using the exact pre-registration rejection format.
- Hub rejects an invalid name without joining, sends at most one rejection event through the pre-registration path, and closes without blocking the hub when delivery cannot complete within the bounded deadline.
- A post-registration empty message disconnects and unregisters the client and produces exactly one leave notice for remaining members.
- A post-registration invalid UTF-8 message disconnects and unregisters the client and produces exactly one leave notice for remaining members.
- A post-registration over-1,024-byte message disconnects and unregisters the client and produces exactly one leave notice for remaining members.
- A full outbound queue triggers slow-client teardown: hub marks unregistered, closes outbound queue once, signals lifecycle owner, reader and writer stop without draining the full queue; other clients receive the current event, then exactly one leave notice.
- Repeated slow-client teardown signals for the same client are idempotent.
- Hub receive order equals cross-client arrival order for any interleaved events.
- An unregistered client never produces a leave notice.
- A sole disconnecting client produces no leave recipient and the hub does not emit a leave to a non-existent remaining membership.
- No history is replayed to a new joiner.
- Shutdown stops new registration, unregisters once, closes queues once, closes connections, and waits for pumps.
- Fake-client tests verify hub behavior without a real listener.
- Integration tests bind `127.0.0.1:0` only.
- All tests pass under the race detector with no sleep synchronization.

## 15. Common Mistakes to Watch For
Normalizing the username beyond case-insensitive uniqueness, trimming whitespace, rejecting CRLF-split lines, misordering the leave notice relative to the broadcast that overflowed, blocking the hub on a slow client, requiring the writer to drain a full queue, missing the nonblocking enqueue on full queue, leaking the writer or reader goroutine, double-closing the outbound channel, leaving the lifecycle owner unsignalled, replaying history, treating CR as a line delimiter, allowing customizable wire notice prefixes, using sleep to wait for tests, and producing a duplicate leave notice for the same disconnect.

## 16. Topics and References for Study
Study the hub pattern, broadcast channels, channel capacity, select with default, pump ownership, structured concurrency, lifecycle signaling between hub and per-connection owners, idempotent shutdown, and bounded delivery deadlines. Review Go's `net`, `sync`, `context`, and `unicode/utf8` documentation plus the race detector. Read the prior README for Project 071, the required foundation, for framing and lifecycle.

## 17. Self-Assessment Questions
Why is the hub the only owner of membership, outbound queues, and broadcast order, and why is the outbound channel capacity exactly 32? Why is a slow client torn down through a lifecycle owner instead of requiring the writer to drain? Why is the leave notice ordered after the current broadcast's other deliveries? Why are the four wire notice formats exact and not customizable? Why is a post-registration message violation a disconnect rather than just a rejection? Why is pre-registration rejection delivered under a bounded deadline, before membership, and without blocking the hub? Why is the display name preserved exactly on the wire? Why is a CR before LF allowed in a message but not in a username? Why is repeated teardown idempotent? How can tests prove hub order, wire formats, and teardown lifecycle without sleep?

## 18. Definition of Completion
- [ ] Username rules and chat message rules are exact and validated before any hub state change.
- [ ] The four pinned wire notice formats are exact: `[hub] <name> joined` LF, `[hub] <name> left` LF, `<name>: <msg>` LF, `[hub] username rejected` LF.
- [ ] Hub is the sole owner of membership, per-client outbound queues, and broadcast order; cross-client arrival order equals hub receive order.
- [ ] Outbound channel capacity is exactly 32; a nonblocking enqueue on a full queue triggers slow-client teardown through the lifecycle owner without requiring the writer to drain.
- [ ] Post-registration empty, invalid UTF-8, or over-1,024-byte message disconnects and unregisters the client and produces exactly one leave notice for remaining members.
- [ ] Duplicate or invalid name rejection is before membership and never touches the existing member.
- [ ] Pre-registration rejection is best-effort under a bounded delivery deadline and never blocks the hub.
- [ ] Repeated slow-client teardown signals for the same client are idempotent.
- [ ] The hub performs no socket I/O; the reader never writes the socket; the writer owner is the sole writer of socket data and close writes.
- [ ] Shutdown stops new registration, unregisters once, closes each queue once, and waits for pumps.
- [ ] Codec and hub are tested independently with fakes; integration tests bind `127.0.0.1:0` only.
- [ ] All tests pass under the race detector with no sleep synchronization.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions
Add a per-hub metric for joined, rejected, post-registration-violation, and dropped-slow clients visible at shutdown. Add a hub-side log line that records teardown reasons without weakening the pinned wire contract.
