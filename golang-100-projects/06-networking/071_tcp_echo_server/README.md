# Project 071 — TCP Echo Server

## 1. Project Name and Number

- Project 071, tcp_echo_server.
- This README is a learning guide only.
- You will create every source and test file yourself in `06-networking/071_tcp_echo_server/`.
- This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea

A TCP line protocol on a configurable loopback address, possibly ephemeral. Each complete byte line up to the limit is echoed exactly once with a newline. Over-limit closes that connection. Multiple lines per connection and concurrent clients are supported, with deadlined reads and writes per operation and clean shutdown.

## 3. Why This Project Now?

- This project requires Project 041 (context_timeout_example) for context cancellation and deadline discipline and Project 060 (graceful_shutdown_web) for graceful HTTP server shutdown and ownership discipline.
- It introduces TCP connection handling alongside explicit byte framing, idle deadlines, accept-loop shutdown, and the discipline of treating partial and malformed input as a per-connection protocol error rather than server-wide failure.
- Project 070 (read_write_splitting) is optional review only.

## 4. Prerequisites

- Projects 041 and 060 are required prerequisites.
- Project 070 is optional review only for read/write splitting and routing semantics.
- The server tests need only `net.Pipe` for the codec and loopback ephemeral endpoints on `127.0.0.1` for the integration check.
- No public network, no Docker, no environment variables.

## 5. What You Must Know Before Starting

- Know `net`, `net.Conn`, `bufio.Scanner` or byte-buffer reading, deadlines, the difference between EOF without buffered data and EOF with buffered data, listener accept loops, `context` cancellation, sync.WaitGroup, and the race detector.
- Know that `bufio.Scanner` discards the trailing newline but is error-prone for arbitrary bytes and bounded capacity.

## 6. Explanation of New Concepts

### Concepts

- The wire framing is line-based with exactly one delimiter: the LF byte.
- The delimiter is not CRLF.
- A CR immediately before LF is content and must be echoed, not stripped.
- Content may contain arbitrary bytes including NUL and invalid UTF-8.
- Content limit is exactly 65,536 bytes inclusive and excludes the LF.
- The 65,536-byte content limit is a fixed protocol constant; only the listen address and idle deadline policy are configurable.
- Internal framing capacity has at least one byte of delimiter headroom beyond that limit.
- An echo writes the exact content plus exactly one LF and handles partial writes cleanly.
- No error text is written to the peer for over-limit or partial-final input.
- The empty LF-only line is valid; accepted content length is 0..65,536 bytes inclusive.

- The protocol distinguishes two distinct kinds of EOF.
- EOF with no buffered bytes after complete lines is normal and ends the connection cleanly.
- EOF with buffered unterminated bytes is a protocol error for that connection: no echo for those bytes, no error text, the connection is closed, and the listener keeps serving.
- A single line that contains content followed by a line-feed is echoed exactly once.

- Idle policy is configurable.
- The production default is 30 seconds for both read and write.
- Each operation refreshes its own deadline: a read deadline before each line read, a write deadline before each full echo write.
- A timeout closes only the offending client connection; per-client read or write deadline expiry is a per-client outcome and never stops the listener.
- The server keeps an active-connection registry plus a handler WaitGroup.
- Caller context cancellation is the expected shutdown signal: the server closes the listener once, closes each registered connection once, waits for handlers, and returns a clean shutdown result.
- Unexpected accept or setup errors are surfaced through a separate error path; the expected shutdown result is not also reported as an error.

- Text-only protocol examples are permitted.
- As a prose shape: a client sends `hello` followed by the LF byte.
- The server replies `hello` followed by the LF byte.
- A client sends a line of 65,536 content bytes followed by LF; the server echoes it once.
- A client sends a line of 65,537 content bytes followed by LF; the server closes that connection with a contextual error and accepts new clients.
- A client connects, sends `partial`, and exits without a final LF; the server treats it as a protocol error, sends no echo, and closes the connection while continuing to serve.

## 7. Learning Objective

- Implement a deterministic TCP line-echo service with exact framing, exact content limits, exact idle and shutdown discipline, per-connection protocol error semantics, and tests that pin every branch of the protocol without relying on sleep-based synchronization.

## 8. Functional Requirements

1. Server binds to a configurable loopback address; production default is `127.0.0.1:0` (ephemeral).
2. Wire delimiter is exactly the LF byte. CRLF is not stripped; the CR is content before LF.
3. Content may include arbitrary bytes including NUL and invalid UTF-8. UTF-8 validity is not required.
4. Accepted content length is exactly 0..65,536 bytes inclusive, excluding the LF delimiter. The empty content line is valid and is echoed as one LF.
5. The 65,536-byte content limit is a fixed protocol constant; only the listen address and idle deadline policy are configurable.
6. Internal framing buffer has at least one byte of delimiter headroom beyond the content limit.
7. Scan path detects content length over the limit and closes that connection with a contextual error.
8. Over-limit detection never writes any error text to the peer.
9. Echo writes the exact content plus one LF and handles partial writes.
10. Multiple lines per connection are supported until the client closes, errors, or violates the protocol.
11. Multiple concurrent clients are supported.
12. Read deadline is refreshed before each line read; write deadline is refreshed before each full echo write.
13. Default idle policy is 30 seconds; production wiring uses that, while tests inject shorter or longer durations.
14. A deadline expiry closes only that client connection.
15. Server maintains an active-connection registry and a handler WaitGroup.
16. Context cancellation closes the listener, closes each registered connection once, waits for handlers, and returns a clean shutdown result.
17. Accept errors that are not listener-closed are surfaced; expected closed-listener shutdown is not treated as an error.
18. EOF without buffered bytes is normal.
19. EOF with buffered unterminated bytes is a protocol error for that connection. No echo is produced for those bytes. No error text is written. The connection is closed. The listener keeps serving.
20. No goroutine per operation leaks; each connection is closed once.
21. Unit tests use `net.Pipe` for codec-only checks.

## 9. Inputs and Outputs

### Interface Contract

- Server input is a TCP loopback address.
- Optional inputs are idle read and write durations and a context for shutdown.
- The content limit is a fixed protocol constant, not a runtime option.
- Codec input is a byte stream with LF-terminated lines.
- Codec output is the echoed content plus one LF per accepted line; rejected lines produce no response on the wire and close the connection.
- Server output is a bound address.
- A graceful shutdown returns a clean shutdown result.
- Unexpected accept or setup errors are surfaced through a separate error path and are not reported as the shutdown result.

## 10. Rules and Edge Cases

- LF is the only delimiter.
- A CR before LF is content and is preserved.
- Content from 0 up to and including 65,536 bytes is echoed once.
- Content of 65,537 bytes or more closes the connection with a contextual error and no peer-visible error text.
- NUL and invalid UTF-8 are valid content.
- Empty line is valid.
- A connection that closes cleanly after complete lines is normal.
- A connection that closes with buffered unterminated bytes is a protocol error for that connection; the listener continues.
- Deadlines apply per operation and are refreshed each time.
- Caller context cancellation is the expected shutdown signal; the listener close path is not an error.
- Per-client read or write deadline expiry is a per-client outcome and does not stop the listener.
- Unexpected accept or setup errors are surfaced through a separate error path.
- Slow or stuck clients do not stop the server.

## 11. Project Constraints

- Loopback only.
- No public network.
- No TLS.
- No authentication.
- No multipart messages, no binary frames, no command codes.
- The codec is line-only and the only response is an echo or a connection close.
- No retry logic, no recovery, no persistence.
- The repository must contain no generated code in this README; tool installation and code generation belong to the learner.

## 12. Design Questions Before Coding

- How is the framing buffer sized to give at least one byte of delimiter headroom over 65,536 content bytes?
- How does the read loop distinguish a content over-limit from a CRLF that drifted past a mistaken splitter?
- How does the read loop distinguish clean EOF from buffered unterminated EOF?
- How is single-write-permission handled under partial writes?
- How is the connection registry protected against concurrent close?
- How is the WaitGroup incremented and decremented so handlers never leak past shutdown?
- How is the listener-closed accept error filtered from real accept failures?
- How are deadlines reset so a slow echo does not consume the next read's budget?

## 13. Implementation Milestones

1. Define the line codec with LF-only delimiter, content cap, and headroom.
2. Define the per-connection read loop that distinguishes normal EOF from buffered unterminated EOF.
3. Define the per-connection write path that handles partial writes against a write deadline.
4. Define the server with listener, accept loop, connection registry, and handler WaitGroup.
5. Define the idle policy injection and the production default of 30 seconds.
6. Define the context-driven shutdown that closes the listener once, closes each registered connection once, waits for handlers, and returns a clean shutdown result.
7. Complete codec, server, shutdown, deadline, race, and integration tests using `net.Pipe` and `127.0.0.1:0`.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Codec over `net.Pipe` echoes a single short line exactly once with one LF.
- Codec over `net.Pipe` echoes the empty LF-only line as one LF.
- Codec over `net.Pipe` echoes multiple lines on one connection.
- Codec over `net.Pipe` echoes a line of 65,536 content bytes exactly once with one LF.
- Codec over `net.Pipe` closes the connection on a line of 65,537 content bytes and writes no error text.
- Codec over `net.Pipe` preserves NUL and invalid UTF-8 in content.
- Codec over `net.Pipe` preserves a CR before LF as content; no CRLF stripping.
- Codec over `net.Pipe` closes the connection on buffered unterminated bytes after EOF and writes no echo.
- Codec over `net.Pipe` returns successfully on clean EOF with no buffered bytes.
- Integration: connect to `127.0.0.1:0`, send three lines, verify three echoes on one connection.
- Integration: open many concurrent clients on the ephemeral port and verify each round-trip.
- Integration: send a 65,536-byte line and an over-limit line, verify the second connection closes and the listener keeps serving.
- Integration: send a line containing NUL and invalid UTF-8; verify the exact echo.
- Integration: inject a short deadline and verify a non-responsive client is closed while the server keeps serving.
- Integration: a per-client deadline expiry does not stop the listener.
- Integration: invoke shutdown via caller context cancellation, verify the listener stops, handlers finish, and shutdown returns a clean result.
- Integration: an unexpected accept or setup error is surfaced through the separate error path and is not reported as the shutdown result.
- No goroutine per operation leaks; the registry counts drop to zero on shutdown.
- All tests pass under the race detector.
- No test uses sleep to wait for protocol correctness.

## 15. Common Mistakes to Watch For

- Treating CRLF as the delimiter, stripping a CR before LF, treating NUL or invalid UTF-8 as invalid content, off-by-one on the 65,536 limit, forgetting the delimiter headroom, conflating clean EOF with buffered unterminated EOF, writing a textual error to the peer on violation, sharing one deadline across read and write, leaking the WaitGroup on shutdown, double-closing connections, retrying the listener loop on a closed-listener error, using sleep to wait for tests, and forgetting to free the framing buffer between connections.

## 16. Topics and References for Study

- Study TCP framing, line scanners, byte-buffer reads, partial writes, deadlines, context cancellation, structured concurrency, and the difference between `EOF` as a transport signal and `EOF` as a protocol signal.
- Review the Go `net` package, `bufio`, `sync`, and `context` documentation plus the race detector.
- Read the prior READMEs for Projects 041 and 060 as required foundations for context cancellation, deadline discipline, and shutdown ownership, and Project 070 as optional catalog context for routing semantics.

## 17. Self-Assessment Questions

1. Why is LF the only delimiter, and why is a CR before LF preserved?
2. Why is the content limit exactly 65,536 bytes inclusive of zero, and why is one byte of headroom required?
3. Why is the content limit a fixed protocol constant while the listen address and deadlines are configurable?
4. Why is buffered unterminated EOF a protocol error while clean EOF is normal?
5. Why is no error text written to the peer on violation?
6. Why is the deadline refreshed per operation rather than per connection?
7. Why does caller context cancellation return a clean shutdown result while unexpected accept or setup errors are surfaced through a separate error path?
8. Why is per-client deadline expiry a per-client outcome that does not stop the listener?
9. Why is shutdown owned by the listener plus the registry plus the WaitGroup rather than by a global flag?
10. How can tests prove correctness without sleep?

## 18. Definition of Completion

- [ ] Wire framing is LF-only with the CR preserved as content.
- [ ] Content limit is exactly 65,536 bytes inclusive with at least one byte of delimiter headroom; empty line is valid.
- [ ] Over-limit lines close the connection with a contextual error and no peer-visible error text.
- [ ] Buffered unterminated EOF is a per-connection protocol error; clean EOF is normal.
- [ ] Read and write deadlines are refreshed per operation; default production idle is 30 seconds; tests inject durations; per-client deadline expiry does not stop the listener.
- [ ] Caller context cancellation closes the listener once, closes each registered connection once, waits for handlers, and returns a clean shutdown result.
- [ ] Unexpected accept or setup errors are surfaced through a separate error path and are not reported as the shutdown result.
- [ ] Multiple lines per connection and many concurrent clients are supported.
- [ ] Codec tests use `net.Pipe`; integration tests bind `127.0.0.1:0` only.
- [ ] No goroutine per operation leaks; the registry counts drop to zero on shutdown.
- [ ] All tests pass under the race detector with no sleep synchronization.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions

- Add structured logging that records accept, connect, close, violation, and shutdown events with the bound address.
- Add a per-connection byte-and-line counter exposed at shutdown for capacity planning.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 041 — Context Timeout Example](../../03-concurrency/041_context_timeout_example/README.md#20-prerequisite-based-documentation-guide), [Project 060 — Graceful Shutdown Web](../../04-apis-and-services/060_graceful_shutdown_web/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **Standards and concept references:** [RFC 9293: TCP](https://www.rfc-editor.org/rfc/rfc9293.html).

### Project-specific learning focus

- **Learn now:** application framing over streams, partial reads and writes, EOF meanings, deadlines, accept-loop shutdown, per-connection ownership, byte limits, and race-tested cleanup.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
