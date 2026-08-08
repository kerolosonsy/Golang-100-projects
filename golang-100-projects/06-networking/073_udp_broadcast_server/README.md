# Project 073 — UDP Broadcast Server

## 1. Project Name and Number

- Project 073, udp_broadcast_server.
- This README is a learning guide only.
- You will create every source and test file yourself in `06-networking/073_udp_broadcast_server/`.
- This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea

Despite the historical name, the required safety scope is a local discovery service over UDP loopback only. One datagram is one message. The protocol is a small versioned JSON request and response with exact field names. The server replies only to the source address reported by the packet API. The client retries because UDP loss, duplication, and reordering are possible, deduplicates by server identity, and sorts deterministically.

## 3. Why This Project Now?

- This project requires Project 072 (tcp_chat_room) and Project 071 (tcp_echo_server).
- It builds on network service lifecycle and concurrent-client discipline while introducing the datagram model, per-datagram framing, packet loss and duplicate handling, the discipline of replying only to the source address, the separation of server request diagnostics from client response diagnostics, and the deterministic deduplication policy that distinguishes this service from a TCP broadcast.

## 4. Prerequisites

- Projects 072 and 071 are required prerequisites.
- No additional project is required; earlier networking projects are optional review only.
- The tests need only `127.0.0.1` with ephemeral ports and an injectable packet boundary for retry simulation.
- No multicast, no LAN broadcast, no public network, no environment variables.

## 5. What You Must Know Before Starting

- Know `net.UDPConn`, the difference between length-prefixed byte slices and stream reads, deadlines, `context`, the use of `ReadMsgUDP` style truncation flags where available, the difference between a missing server and a slow server, and the race detector.

## 6. Explanation of New Concepts

### Concepts

- The wire format is a single top-level JSON object in each datagram.
- The request contains exactly two fields with the literal names `version` and `request_id` and no other fields.
- The response contains exactly five fields with the literal names `version`, `request_id`, `server_id`, `service_name`, and `service_address` and no other fields.
- Field order is not significant.
- Any unknown field, duplicate field, trailing JSON value or data, invalid UTF-8 in a string, wrong type for a known field, wrong version, or wrong request ID is rejected as malformed.
- Maximum application datagram payload is 1,024 bytes inclusive of the complete JSON bytes; IP and UDP headers are not part of the payload returned by the packet API and are not counted.

- The request `version` is exactly the numeric value 1.
- The request `request_id` is a string of 16..64 lowercase hexadecimal characters.
- The response `version` is exactly the numeric value 1.
- The response `request_id` is the exact same string echoed back.
- The response `server_id` and `service_name` are nonempty ASCII identifiers from 1..64 characters using letters, digits, underscore, and hyphen.
- The response `service_address` is a string in the form `127.0.0.1:<port>` where `<port>` is a positive integer in 1..65,535; the address is validated as the literal numeric loopback host `127.0.0.1` plus a positive port with no hostname resolution and no name lookup.

- One datagram is one complete message.
- The server reads up to the maximum payload size and uses a max-plus-one receive buffer plus `ReadMsgUDP`-style truncation flags where available to detect oversize or truncation with bounded behavior; the platform limitation is documented when truncation flags are unavailable.

- The server request-side diagnostic categories are exactly `oversized_or_truncated`, `malformed`, and `wrong_version`.
- The precedence is oversized or truncated first, then JSON shape, then version.
- A valid request increments no diagnostic category.
- The server sends no response to a malformed or wrong-version request.

- The client ignored-response diagnostic categories are exactly `wrong_request_id`, `wrong_version`, and `malformed`.
- Shape validation runs before version and request ID checks.
- One increment is recorded per ignored datagram.
- Valid responses and identical duplicate responses do not increment ignored counters.
- A conflicting valid response for one server ID is a typed conflict and is returned with no partial result.

- The client pins exactly 3 send attempts even if a valid server responds during an earlier attempt, because discovery cannot know all responders in advance.
- All 3 sends use the same request ID.
- Each attempt uses an injected per-attempt collection window.
- Expiration of an injected collection window is a normal attempt boundary, not caller context cancellation.
- After exactly three normal windows, any result means success and zero results means typed discovery-timeout.
- Caller-context termination and a typed conflicting-server result are the only early terminal outcomes; conflict stops further sends and returns no partial result.

- Caller context cancellation or deadline expiry is separate from injected collection-window expiration.
- Caller context cancellation aborts all attempts and discards any partial result.
- The client returns the context outcome and no partial result on caller context cancellation.

- Deduplication is by server ID.
- Identical duplicates collapse.
- Conflicting valid responses for one server ID are a typed conflict.
- Final results are sorted by server ID ascending.

- The server uses read deadlines and context shutdown.
- The server replies only to the source address reported by the packet API.
- The request carries no return address.
- The server binds `127.0.0.1:0` only.
- Context shutdown closes the packet connection once and waits for the serve loop.

- Text-only protocol examples are permitted.
- As a prose shape: a client sends one datagram whose JSON top-level object has only `version` set to 1 and `request_id` set to a 32-character lowercase hex string.
- The server replies with one datagram whose JSON top-level object has `version` set to 1, the same `request_id`, a `server_id`, a `service_name`, and a `service_address` of the form `127.0.0.1:<port>`.
- A datagram with payload above 1,024 bytes or with malformed JSON or wrong version produces no response and a single server-side diagnostic increment.

## 7. Learning Objective

- Implement a deterministic loopback UDP discovery service with exact JSON field contract, exact malformed policy, exact retry and dedup policy, exact separation of server request diagnostics from client response diagnostics, and tests that prove loss, duplicate, reordering, oversized, and shutdown without sleep-based synchronization.

## 8. Functional Requirements

1. Server binds to `127.0.0.1:0` only.
2. Wire format is a single top-level JSON object per datagram.
3. Request fields are exactly `version` and `request_id` and no other fields.
4. Response fields are exactly `version`, `request_id`, `server_id`, `service_name`, and `service_address` and no other fields.
5. Request `version` is exactly the numeric value 1; request `request_id` is 16..64 lowercase hexadecimal characters.
6. Response `version` is exactly the numeric value 1; response `request_id` is the exact echoed request ID.
7. Response `server_id` and `service_name` are nonempty ASCII identifiers from 1..64 characters using letters, digits, underscore, and hyphen.
8. Response `service_address` is the literal numeric loopback host `127.0.0.1` plus a positive port 1..65,535 with no hostname resolution.
9. Maximum payload is 1,024 bytes inclusive.
10. One datagram is one complete message.
11. Unknown fields, duplicate fields, trailing JSON values, invalid UTF-8, wrong types, wrong version, wrong IDs, and payload above 1,024 bytes are malformed.
12. Server request-side diagnostic categories are exactly `oversized_or_truncated`, `malformed`, and `wrong_version` with precedence oversized or truncated first, then JSON shape, then version.
13. A valid request increments no server-side diagnostic category.
14. Server sends no response to a malformed or wrong-version request.
15. Server replies only to the source address reported by the packet API.
16. Request carries no return address.
17. Server uses read deadlines and context shutdown.
18. Context shutdown closes the packet connection once and waits for the serve loop.
19. In a non-conflict run with an active caller context, the client sends exactly 3 times even if a valid server responds during an earlier attempt; all 3 sends use the same request ID. Caller-context termination or typed conflicting-server result stops further sends.
20. Each attempt uses an injected per-attempt collection window; window expiration is a normal attempt boundary, not caller context cancellation.
21. After exactly three normal windows, any result means success; zero results means typed discovery-timeout.
22. Caller context cancellation or deadline expiry aborts all attempts and discards partial results; the client returns the context outcome and no partial result.
23. Client ignored-response diagnostic categories are exactly `wrong_request_id`, `wrong_version`, and `malformed`, with shape validation before version and request ID checks.
24. One increment per ignored datagram; valid responses and identical duplicate responses do not increment ignored counters.
25. A conflicting valid response for one server ID is a typed conflict, stops further sends, and is returned with no partial result.
26. Deduplication is by server ID; identical duplicates collapse; conflicting valid responses for one server ID are a typed conflict.
27. Final results are sorted by server ID ascending.
28. Overflow detection uses `ReadMsgUDP`-style truncation flags when supported; otherwise a max-plus-one receive buffer provides bounded oversized detection and the platform limitation is documented.
29. Codec is independently testable without a real packet connection.

## 9. Inputs and Outputs

### Interface Contract

- Server input is the loopback bind address and a context.
- Server output is a bound address and a shutdown result.
- Request input is a context, a request ID, and a caller deadline.
- Request output is a sorted deduplicated list of typed discovery results, a typed conflict, a typed discovery-timeout, or the caller context outcome with no partial result.
- The client never attempts to discover beyond loopback.

## 10. Rules and Edge Cases

- A valid request increments no server-side diagnostic category.
- A malformed request increments exactly one server-side diagnostic category with the pinned precedence.
- Oversized payloads and `ReadMsgUDP` truncation both count as oversized or truncated.
- A request with a wrong version but otherwise parseable JSON increments wrong version.
- A request with malformed JSON increments malformed.
- A request with valid JSON but an unknown field or duplicate field increments malformed.
- The server never sends a response to a malformed or wrong-version request.
- The client never claims success on a partial result.
- The client never returns a typed discovery-timeout while a valid server was found.
- The client never retries on top of a caller context outcome.
- The client never treats an injected collection-window expiration as caller context cancellation.
- Identical duplicate responses for the same server ID collapse and do not increment ignored counters.
- A conflicting valid response for one server ID stops further sends and returns typed conflict with no partial result.

## 11. Project Constraints

- Loopback only.
- No multicast, no LAN broadcast, no public network.
- No TLS.
- No authentication.
- No reliability claim.
- No deduplication across requests.
- No hostname resolution; the service address is the literal numeric loopback host `127.0.0.1` plus a positive port.
- This project contains no generated code and no tool installation step in the learner pipeline.

## 12. Design Questions Before Coding

- How is the JSON strict decoder configured to reject unknown fields and duplicate fields and accept only the exact field set per direction?
- How is the request ID validated as lowercase hexadecimal without scanning the whole string twice?
- How is the server-side precedence implemented so oversized or truncated always wins, then malformed JSON shape, then wrong version?
- How is the client-side ignored-response precedence implemented so shape validation runs before version and request ID checks?
- How is the source address captured once per `ReadFromUDP` and used exactly once for the reply?
- How is the injected per-attempt collection window exposed for tests without using sleep?
- How is the typed conflict surfaced without leaking internal detail?
- How is `ReadMsgUDP` truncation handled when the platform does not support it?
- How is the diagnostic counter concurrency-safe and observable in tests?
- How is caller context cancellation distinguished from injected collection-window expiration?

## 13. Implementation Milestones

1. Define the codec with strict JSON, exact field names per direction, version 1, request ID, server ID, service name, and address validation.
2. Define the server-side request diagnostic counter with the pinned precedence.
3. Define the client-side ignored-response diagnostic counter with the pinned precedence and shape-before-version-before-ID ordering.
4. Define the server with read deadline, source-address reply, and context shutdown.
5. Define the client with one unchanged request ID, exactly 3 attempts, injected collection window, and caller context binding.
6. Define the deduplication and sort policy including the typed conflict.
7. Define the codec and platform tests.
8. Define the integration tests on `127.0.0.1:0` only.
9. Run race detector and concurrency tests.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Codec accepts a valid request with exactly `version` and `request_id` and a valid response with exactly `version`, `request_id`, `server_id`, `service_name`, and `service_address`.
- Codec rejects an oversize payload above 1,024 bytes inclusive.
- Codec rejects unknown fields, duplicate fields, trailing JSON values, invalid UTF-8 in strings, wrong types for known fields, wrong version, and wrong request ID.
- Codec rejects a response `service_address` that is not the literal numeric `127.0.0.1` plus a positive port 1..65,535, including any hostname-form address.
- Server increments exactly one server-side counter for each malformed request with the pinned precedence; valid requests increment no counter.
- Server sends no response on malformed or wrong-version input.
- Server replies only to the source address reported by the packet API.
- Server uses `ReadMsgUDP`-style truncation flags where available, and otherwise detects oversized via a max-plus-one buffer.
- Server stops on context cancellation without leaking the serve loop.
- Client performs exactly 3 send attempts even when a valid server responds during the first attempt; all 3 sends use the same request ID; identical duplicate responses collapse; final results return after the third collection window.
- Client never returns early success because discovery cannot know all responders in advance.
- Client returns typed discovery-timeout after exactly three normal windows with zero results.
- Client returns the caller context outcome when the caller context cancels before the three-attempt process completes, with no partial result.
- Client distinguishes injected collection-window expiration from caller context cancellation and never treats one as the other.
- Client ignores responses with wrong request ID, wrong version, or invalid shape and counts them diagnostically; shape validation runs before version and request ID checks.
- Client does not increment ignored counters for valid responses or for identical duplicate responses for the same server ID.
- Client collapses identical duplicates; conflicting valid responses for one server ID stop further sends and return typed conflict with no partial result.
- Client sorts final results by server ID ascending.
- Integration tests bind `127.0.0.1:0` only; no multicast, no LAN broadcast.
- All tests pass under the race detector with no sleep synchronization; retry boundaries are injected.

## 15. Common Mistakes to Watch For

- Allowing multicast or LAN broadcast, treating the request as a stream, claiming reliability, retrying on top of a caller context outcome, partial results on cancellation, returning early success after the first attempt, missing the precedence on server-side diagnostic increments, sending a response to a malformed request, replying to a claimed return address, ignoring overflow detection on platforms without `ReadMsgUDP` truncation flags, treating injected collection-window expiration as caller context cancellation, resolving hostnames for the service address, incrementing ignored counters for valid or identical-duplicate responses, returning a typed conflict together with a partial result, using sleep to simulate retry, and confusing a TCP broadcast with a UDP loopback service.

## 16. Topics and References for Study

- Study the datagram model, packet loss, duplication, and reordering, JSON strict decoding with exact field allowlists, `ReadMsgUDP` truncation, deadlines, context cancellation, deterministic deduplication, separation of server request diagnostics from client response diagnostics, and the distinction between injected collection-window expiration and caller context cancellation.
- Review Go's `net`, `encoding/json`, `context`, and platform UDP documentation plus the race detector.
- Read the prior READMEs for Project 072 for hub ownership and pump semantics and Project 071 for framing and lifecycle; both are required foundations.

## 17. Self-Assessment Questions

1. Why is the request ID unchanged across exactly 3 attempts even after a valid response, and why is each collection window per-attempt rather than global?
2. Why are server-side request diagnostics and client-side ignored-response diagnostics separate?
3. Why does shape validation run before version and request ID checks on the client?
4. Why is a typed conflict surfaced instead of collapsing conflicting data?
5. Why is the packet source address the only return address?
6. Why is one malformed payload counted in exactly one server-side diagnostic category?
7. Why does the client never return a partial result on caller context cancellation?
8. Why is injected collection-window expiration not the same as caller context cancellation?
9. Why is `ReadMsgUDP` truncation preferred for overflow detection, and why is `127.0.0.1` validated as a literal numeric host without hostname resolution?
10. How can tests prove retry and dedup without sleep?

## 18. Definition of Completion

- [ ] Wire contract is a single top-level JSON object per datagram with the exact field names per direction.
- [ ] Maximum payload is 1,024 bytes inclusive; overflow detection uses `ReadMsgUDP`-style truncation flags where available, and a max-plus-one buffer otherwise.
- [ ] Server binds `127.0.0.1:0` only and replies only to the source address.
- [ ] Server-side request diagnostic categories are exactly `oversized_or_truncated`, `malformed`, and `wrong_version` with pinned precedence; valid requests increment no counter.
- [ ] Client-side ignored-response diagnostic categories are exactly `wrong_request_id`, `wrong_version`, and `malformed` with shape-before-version-before-ID ordering and one increment per ignored datagram.
- [ ] Client retries exactly 3 times with one unchanged request ID; identical duplicates collapse; results return after the third collection window; no early success.
- [ ] Client deduplicates by server ID, sorts ascending, surfaces a typed conflict with no partial result, and returns typed discovery-timeout when zero results.
- [ ] Client returns the caller context outcome and no partial result on caller context cancellation or deadline.
- [ ] Injected collection-window expiration is a normal attempt boundary and is never treated as caller context cancellation.
- [ ] Codec is independently testable; integration tests bind `127.0.0.1:0` only.
- [ ] No multicast, no LAN broadcast, no reliability claim, no hostname resolution.
- [ ] This project contains no generated code and no tool installation step in the learner pipeline.
- [ ] All tests pass under the race detector with no sleep synchronization.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions

- Add a per-server metric for accepted, malformed, and ignored responses visible at shutdown.
- Add a structured precedence table that documents both server-side and client-side ordering for tests and operators.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 072 — TCP Chat Room](../../06-networking/072_tcp_chat_room/README.md#20-prerequisite-based-documentation-guide), [Project 071 — TCP Echo Server](../../06-networking/071_tcp_echo_server/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **Standards and concept references:** [RFC 768: UDP](https://www.rfc-editor.org/rfc/rfc768.html).

### Project-specific learning focus

- **Learn now:** datagram boundaries, loss, duplication and reordering, truncation detection, strict schemas, deduplication, collection windows, diagnostics, and cancellation precedence.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
