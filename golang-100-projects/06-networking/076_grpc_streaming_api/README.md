# Project 076 — gRPC Streaming API

## 1. Project Name and Number
Project 076, grpc_streaming_api. This README is a learning guide only. You will create every source and test file yourself in `06-networking/076_grpc_streaming_api/`. This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea
A single bidirectional gRPC stream named EventProcessor that accepts an ordered series of client events and emits an ordered series of server replies. Each client event carries a unique positive event ID and a bounded UTF-8 payload. Each server reply echoes the originating ID and returns a deterministic processed result or a per-event public processing failure. Response order is pinned to receive order. The stream enforces per-message validation and a per-stream count limit; every failure that arises at the message boundary is stream-fatal and terminates the RPC with the appropriate gRPC status. The stream does not issue a per-event response for a stream-fatal offending message.

## 3. Why This Project Now?
This project requires Project 075 (grpc_user_service) as the immediate predecessor, Project 071 (tcp_echo_server) for TCP framing, idle deadlines, and per-connection protocol error discipline, and Project 060 (graceful_shutdown_web) for graceful server shutdown and lifecycle ownership. Project 041 (context_timeout_example) is recommended review for cancellation and deadline propagation, and Project 034 (worker_pool_basic) is recommended review for backpressure and ownership of results; both are optional review only and are not formal prerequisites. This project extends the unary gRPC discipline into the streaming RPC discipline: ownership of send and receive, half-close semantics, flow control, and bounded queues instead of unbounded goroutines.

## 4. Prerequisites
Projects 075, 071, and 060 are required prerequisites. Project 075 is the immediate predecessor for unary gRPC status mapping. Project 071 is required for TCP connection handling, byte framing, idle deadlines, accept-loop shutdown, and per-connection protocol error discipline. Project 060 is required for graceful server shutdown and lifecycle ownership. Project 041 and Project 034 are recommended review for cancellation propagation and worker-pool ownership but are not formal prerequisites. The runtime modules are pinned: `google.golang.org/grpc` at version `v1.83.0` and `google.golang.org/protobuf` at version `v1.36.11`. The generation provenance pins the code generator plugins used by the learner to `protoc-gen-go` at the version aligned with the protobuf module `v1.36.11` and `protoc-gen-go-grpc` at version `v1.6.2`. The command module `google.golang.org/grpc/cmd/protoc-gen-go-grpc` is not a runtime import. The learner manages tool versions through an appropriate tooling mechanism. No specific `protoc` compiler release is required by this guide. Generated protobuf code belongs to the learner.

## 5. What You Must Know Before Starting
Know the gRPC streaming RPC model, the four streaming kinds, the bidirectional half-close discipline, server-streaming receive ownership, client-streaming send ownership, gRPC status codes, `context` cancellation and deadline propagation with exact precedence between Canceled and DeadlineExceeded, channel-based backpressure, the `bufconn` in-memory transport, the race detector, and Project 075 unary status mapping.

## 6. Explanation of New Concepts
The service is `EventProcessor` in package `tutorial.event.v1`. The RPC is bidirectional streaming named `Process`. The request message `Event` has exactly two fields: field `id` at number 1 as a positive signed-64 integer, and field `payload` at number 2 as a UTF-8 string. There is no audit field; ordering on the wire is server-assigned from receive order. The response message `EventResult` has exactly four fields: field `id` at number 1 as a signed-64 integer that echoes the originating request ID, field `code` at number 2 as a small enum, field `result` at number 3 as a string carrying the deterministic processed payload, and field `error_message` at number 4 as a string carrying a generic public message when processing fails. The code enum has exactly two values: `OK` and `PROCESSING_FAILED`. There are no per-event codes that correspond to validation rejection; validation rejection is stream-fatal.

The wire contract pins response order to receive order. The server emits at most one `EventResult` per accepted `Event`, and the order of `EventResult` on the wire matches the order in which the corresponding `Event` was read from the client. The server does not interleave results ahead of earlier events even when an earlier event blocks briefly inside the injected processor.

Per-message validation is stream-fatal and produces no per-event response for the offending message. A payload that is not valid UTF-8, a payload of length 0, a payload of length greater than 4,096 bytes, an ID that is not positive, and an ID that duplicates an ID already accepted on this stream are all detected at the message boundary and terminate the RPC with the appropriate gRPC status. The empty message boundary itself never produces an `EventResult`. A non-positive ID terminates with `InvalidArgument`. A duplicate ID terminates with `AlreadyExists`. An invalid UTF-8 payload, an empty payload, and an oversized payload each terminate with `InvalidArgument`. No `EventResult` is emitted for the offending message.

The per-stream count limit applies to accepted messages. The maximum number of otherwise-valid accepted messages on a single stream is exactly 10,000 inclusive. A message that is otherwise valid but would be the 10,001st accepted message on this stream terminates the RPC with `ResourceExhausted` and is stream-fatal with no `EventResult` emitted. The 10,000-message limit counts every otherwise-valid accepted message including a message whose processor later returns a per-event processing failure. Duplicate-ID tracking is bounded by the same 10,000-message limit so memory growth is capped at the stream limit.

The injected processor has a small public contract. Its input is an `Event` and an injected context. Its output is a deterministic processed result string and a processing error. A processing error is mapped to a per-event `EventResult` with code `PROCESSING_FAILED` and a generic public `error_message`; a per-event processing error does not by itself terminate the stream. The injected processor is sequential and not parallel; multiple processing workers are not part of the required design.

The application has bounded inbound and outbound queues to coordinate the receive loop and the single send owner. The inbound queue capacity is exactly 32 messages. The outbound queue capacity is exactly 32 messages. Both capacities are fixed at configuration time and are not unbounded. There is exactly one goroutine reading client `Event` messages, one goroutine applying the injected processor and forwarding into the outbound queue, and exactly one goroutine that owns the outbound send path. Exactly one owner calls send, and receive-order responses are preserved by construction.

The client may send several messages before reading any reply, and may read replies concurrently with sending. The half-close discipline is exact. When the client closes the send side, the server receive loop observes `io.EOF` on the next receive. The server then stops accepting further work, finishes every event it has already accepted into the inbound queue, sends the corresponding `EventResult` on the wire, returns `nil` from the handler, and only then does the client receive a clean `io.EOF` on its receive loop. The half-close on the client side does not by itself terminate the server before accepted work finishes.

Caller cancellation or a non-`io.EOF` send or receive failure observed on the stream context cancels all owned work and prevents any further send. There is no fallback to a per-event error response on cancellation or transport failure. A pre-cancelled non-deadline caller context maps to `Canceled`. An already-expired caller deadline maps to `DeadlineExceeded`. The two are not interchangeable.

HTTP/2 and gRPC flow control applies at the stream and connection level. The stream respects flow control through gRPC; the application layer still uses bounded queues to avoid application-side memory growth. The learner explains, in prose, that flow control limits bytes in flight per stream and per connection, and that application buffers remain bounded by the queue capacities defined in this guide.

The lifecycle test discipline is deterministic. Tests assert lifecycle through synchronization events rather than by counting goroutines directly. The graceful-stop test holds an active stream at an injected barrier, starts `GracefulStop` while the stream is still active, proves by synchronization events that the stop remains pending, then releases, finishes, or cancels the stream and proves that the stop completes. A bounded watchdog force-stops only on a test failure and then fails the test. An optional established leak checker may be used at the learner's discretion; the required tests do not depend on raw goroutine count equality between startup and shutdown.

Tests use `bufconn` only. Tests never bind a fixed port. Tests use channels to synchronize rather than sleep. Concurrent streams share the same server without interfering.

Text-only protocol examples are permitted. As a prose shape: a client sends three `Event` messages with IDs 1, 2, 3 in order, then closes the send side. The server replies with three `EventResult` messages in order echoing IDs 1, 2, 3 with code `OK`. The client then receives a clean `io.EOF`. A client sends a fourth `Event` with ID 4 whose processor returns a processing error; the server replies with one `EventResult` echoing ID 4 with code `PROCESSING_FAILED` and a generic public message, and the stream continues. A client sends an `Event` with ID 0; the server terminates the RPC with `InvalidArgument` and emits no `EventResult` for that message. A client sends an `Event` with ID 5 after ID 5 was already accepted; the server terminates the RPC with `AlreadyExists` and emits no `EventResult` for that message. A client sends a payload of 4,097 bytes; the server terminates the RPC with `InvalidArgument` and emits no `EventResult`. A client sends a 10,001st otherwise-valid `Event` after 10,000 were already accepted; the server terminates the RPC with `ResourceExhausted`. A client cancels its context mid-stream; the server cancels owned work and returns `Canceled`, and no further `EventResult` is attempted.

## 7. Learning Objective
Implement a deterministic bidirectional gRPC stream with exact per-message and per-stream limits, exact ordering, exact half-close semantics, exact per-event versus stream-fatal mapping, exact cancellation propagation, and tests that pin every branch through `bufconn` without sleep and without flaky raw goroutine-count assertions.

## 8. Functional Requirements
1. Service is `EventProcessor` in package `tutorial.event.v1`. RPC is bidirectional streaming named `Process`. Messages are `Event` and `EventResult`.
2. `Event` field numbers are pinned: `id=1` and `payload=2` only. There is no audit field.
3. `EventResult` field numbers are pinned: `id=1`, `code=2`, `result=3`, `error_message=4`.
4. The code enum has exactly two values: `OK` and `PROCESSING_FAILED`. There are no per-event codes for validation rejection.
5. Payload length is exactly 1..4,096 bytes inclusive and must be valid UTF-8.
6. Empty payload, invalid UTF-8, oversized payload, and non-positive ID are each stream-fatal with `InvalidArgument` and no `EventResult` for the offending message.
7. Duplicate ID on the same stream is stream-fatal with `AlreadyExists` and no `EventResult` for the offending message.
8. Maximum otherwise-valid accepted messages on one stream is exactly 10,000 inclusive; an otherwise-valid message that would be the 10,001st is stream-fatal with `ResourceExhausted` and no `EventResult`.
9. The 10,000-message limit counts every otherwise-valid accepted message including a message whose processor later returns a per-event processing failure.
10. Duplicate-ID tracking is bounded by the 10,000-message limit.
11. The injected processor returns a deterministic result and a processing error; a processing error maps to per-event `PROCESSING_FAILED` and does not terminate the stream by itself.
12. The injected processor is sequential and not parallel; multiple processing workers are not required.
13. Inbound and outbound application queue capacities are each exactly 32 messages and are not unbounded.
14. Exactly one goroutine owns the outbound send path. Receive-order responses are preserved by construction.
15. Half-close discipline: client `CloseSend` causes the server receive loop to observe `io.EOF`; the server then stops accepting further work, finishes every event already accepted into the inbound queue, sends every corresponding `EventResult`, returns `nil`, and only then does the client receive observe clean `io.EOF`.
16. Cancellation or a non-`io.EOF` send or receive failure cancels all owned work and prevents any further send; no per-event response is issued in place of cancellation.
17. Pre-cancelled non-deadline caller context maps to `Canceled`. Already-expired caller deadline maps to `DeadlineExceeded`. The two are not interchangeable.
18. Status mapping preserves codes exactly. Stream-fatal statuses are `InvalidArgument`, `AlreadyExists`, `ResourceExhausted`, `Canceled`, `DeadlineExceeded`, and `Internal` with a generic public message for unexpected faults. `Internal` is never wrapped into `Unknown`.
19. The lifecycle test discipline relies on synchronization events rather than on raw goroutine-count equality.
20. The graceful-stop test holds an active stream at an injected barrier, starts `GracefulStop` while the stream is still active, proves by synchronization events that the stop remains pending, releases, finishes, or cancels the stream, and proves that the stop completes; a bounded watchdog force-stops only on test failure and then fails the test.
21. Tests use `bufconn` only; no fixed port. Tests do not use sleep synchronization.

## 9. Inputs and Outputs
Stream input is a sequence of `Event` messages on a client stream context. Stream output is a sequence of `EventResult` messages on the server stream context followed by a clean `io.EOF` on success, or a status on stream-fatal termination. Per-event `EventResult.error_message` is generic public when present and never leaks processor internals. The processor boundary is injected; tests substitute a deterministic processor.

## 10. Rules and Edge Cases
Empty payload is `InvalidArgument` at message boundary and stream-fatal with no per-event response. Invalid UTF-8 payload is `InvalidArgument` at message boundary and stream-fatal with no per-event response. Payload greater than 4,096 bytes is `InvalidArgument` at message boundary and stream-fatal with no per-event response. Non-positive ID is `InvalidArgument` at message boundary and stream-fatal with no per-event response. Duplicate ID on the same stream is `AlreadyExists` at message boundary and stream-fatal with no per-event response. The 10,001st otherwise-valid accepted message is `ResourceExhausted` and stream-fatal with no per-event response. Per-event processing error is `PROCESSING_FAILED`, counts as an accepted message, and does not terminate the stream by itself. Pre-cancelled caller context is `Canceled`. Expired deadline is `DeadlineExceeded`. Unexpected internal fault is `Internal` with a generic public message. Concurrent streams are isolated. Inbound and outbound queue capacities are exactly 32. Half-close does not require server immediate close; the server finishes accepted work before returning. Cancellation or non-`io.EOF` send or receive failure prevents any further send.

## 11. Project Constraints
Runtime modules are `google.golang.org/grpc v1.83.0` and `google.golang.org/protobuf v1.36.11`. The code generator plugins used by the learner are pinned to `protoc-gen-go` at the version aligned with `v1.36.11` and `protoc-gen-go-grpc v1.6.2`. The command module is not a runtime import. No database, no TLS, no auth, no unary RPCs in this project, no reflection by default. Generated code belongs to the learner. Tests use `bufconn` only.

## 12. Design Questions Before Coding
How is inbound backpressure bounded at exactly 32 so a fast sender cannot outrun the sequential processor? How is the single send owner coordinated with the inbound queue when the client half-closes before all replies are sent? How is per-event versus stream-fatal mapping distinguished so a per-event processing failure counts toward the 10,000 accepted-message limit without terminating the stream? How is duplicate-ID tracking bounded by the same 10,000-message limit so memory growth is capped? How is the sequential processor chosen so multiple workers are not required? How is the half-close observed on the server side so the server returns `nil` only after every accepted `EventResult` has been sent? How is cancellation or a non-`io.EOF` send or receive failure observed so no further send is attempted? How is `Canceled` distinguished from `DeadlineExceeded` on the stream context? How is the lifecycle test made deterministic through synchronization events instead of raw goroutine counts? How is the graceful-stop watchdog bounded so it force-stops only on test failure and then fails the test?

## 13. Implementation Milestones
1. Define the protobuf contract in prose for package `tutorial.event.v1`, service `EventProcessor`, bidirectional RPC `Process`, the two messages with pinned field numbers, and the two-value code enum; generated code belongs to the learner.
2. Define the bidirectional handler with exactly one send owner, one receive loop, one processor loop, inbound and outbound queues each of capacity 32, and exactly one outbound send path.
3. Define per-message validation: UTF-8 validity, payload length 1..4,096, positive ID, duplicate ID against a 10,000-message bounded set; each failure terminates the RPC with the pinned status and emits no per-event response.
4. Define the 10,000-message accepted-message counter and the `ResourceExhausted` termination on the 10,001st otherwise-valid message, counting per-event processing failures toward the limit and bounding duplicate tracking by the same limit.
5. Define the injected processor boundary with a deterministic result, a processing error, and a sequential rather than parallel execution model.
6. Define per-event processing-error mapping to `PROCESSING_FAILED` with a generic public message; no per-event response is issued for any stream-fatal event.
7. Define context discipline: pre-cancelled maps to `Canceled`; expired deadline maps to `DeadlineExceeded`; cancellation or non-`io.EOF` send or receive failure cancels owned work and prevents further send.
8. Define half-close and clean `io.EOF` semantics: the server receive loop observes `io.EOF` after the client closes the send side, the server stops accepting, finishes every event already accepted, sends every corresponding `EventResult`, returns `nil`, and only then does the client receive observe clean `io.EOF`.
9. Define the `bufconn` test transport with finite dial context and deterministic cleanup.
10. Define the graceful-stop test holding an active stream at an injected barrier, starting graceful stop while the stream is active, proving pending by synchronization events, releasing, finishing, or cancelling, and proving completion; the watchdog force-stops only on failure and then fails the test.
11. Define the full matrix of zero, one, and many message tests plus ordering, half-close, cancellation, duplicate, invalid UTF-8, empty payload, oversized payload, per-stream count, per-event processing error, concurrent streams, and clean shutdown.

## 14. Verification Cases the Learner Must Write
- Generated descriptor exposes package `tutorial.event.v1`, service `EventProcessor`, RPC `Process`, the two messages with the pinned field numbers, and the two-value code enum.
- Zero-message stream: client closes the send side immediately; the server returns clean `io.EOF` with zero `EventResult`.
- One-message stream: a single `Event` produces exactly one `EventResult` echoing its ID with code `OK`.
- Many-message stream: 1,000 ordered `Event` messages produce 1,000 ordered `EventResult` messages with matching IDs and order.
- Per-event processor success and per-event processor error both produce `EventResult` without terminating the stream.
- Non-positive ID terminates the RPC with `InvalidArgument`; no `EventResult` is emitted for that message.
- Duplicate ID on the same stream terminates the RPC with `AlreadyExists`; no `EventResult` is emitted for the second offending message.
- Empty payload terminates the RPC with `InvalidArgument`; no `EventResult` is emitted for that message.
- Invalid UTF-8 payload terminates the RPC with `InvalidArgument`; no `EventResult` is emitted for that message.
- Oversized payload of 4,097 bytes terminates the RPC with `InvalidArgument`; no `EventResult` is emitted for that message.
- Stream exceeding 10,000 otherwise-valid accepted messages terminates the RPC with `ResourceExhausted`; a per-event processing failure in a prior message does not exempt that message from the count.
- Duplicate-ID tracking is bounded by the 10,000-message limit; the duplicate set does not grow beyond the stream limit.
- Response order is receive order even when the injected processor blocks briefly on an earlier event.
- Half-close discipline: the server receive loop observes `io.EOF` after the client closes the send side, finishes every accepted `EventResult`, and returns `nil` before the client receive observes clean `io.EOF`.
- Inbound and outbound queue capacities of 32 are observed by the test by holding the processor longer than the queue capacity and observing the receive loop block rather than spawn more goroutines.
- Exactly one send owner is observed by the test through the single outbound send path; receive-order responses are preserved.
- Client cancellation mid-stream cancels owned work and returns `Canceled`; pre-cancelled non-deadline context returns `Canceled`; already-expired deadline returns `DeadlineExceeded`.
- Cancellation or a non-`io.EOF` send or receive failure prevents any further send; no per-event response replaces the cancellation.
- Concurrent streams are isolated; state from one stream does not affect another.
- The graceful-stop test holds an active stream at an injected barrier, starts graceful stop while the stream is active, proves by synchronization events that graceful stop remains pending, releases, finishes, or cancels the stream, and proves that graceful stop completes; the watchdog force-stops only on failure and then fails the test.
- An optional established leak checker may be added at the learner's discretion; the required tests do not rely on raw goroutine-count equality.
- All tests pass under the race detector.
- No fixed port; all tests use `bufconn`.
- No sleep synchronization is used in tests.

## 15. Common Mistakes to Watch For
Adding an `Event.seq` audit field, adding per-event codes for validation rejection, allowing a payload that is not valid UTF-8, accepting an empty payload, treating an oversized payload as a per-event failure rather than a stream-fatal failure, treating a per-event processing failure as stream-fatal, treating a stream-fatal message as producing a per-event response, allowing the 10,000-message limit to count only successful events and ignore per-event processing failures, allowing duplicate-ID memory growth to exceed the 10,000-message limit, building multiple processing workers and making the design harder to reason about, allowing unbounded queues, calling send from more than one goroutine, allowing the server to return before every accepted `EventResult` has been sent on half-close, attempting to issue a per-event response after cancellation, treating `Canceled` and `DeadlineExceeded` as interchangeable, asserting raw goroutine-count equality between startup and shutdown, using a fixed port for tests, using sleep to synchronize tests, and ignoring HTTP/2 flow control responsibilities at the application layer.

## 16. Topics and References for Study
Study the gRPC streaming RPC model, half-close semantics, server-streaming and client-streaming ownership, gRPC status codes, exact context precedence between `Canceled` and `DeadlineExceeded`, channel-based backpressure and bounded queues, `bufconn` in-memory transport, deterministic lifecycle tests through synchronization events, bounded graceful-stop testing with a watchdog fallback that fails on a missed proof, and HTTP/2 flow control fundamentals. Review the official gRPC and protobuf documentation for the pinned versions. Read the prior README for Project 075 as the immediate predecessor for status mapping, Project 071 for TCP framing and protocol error discipline, and Project 060 for graceful server shutdown and lifecycle ownership. Project 041 for cancellation propagation and Project 034 for backpressure and worker-pool ownership are optional review.

## 17. Self-Assessment Questions
Why are `Event.seq` and per-event validation codes excluded from the contract, and why are validation failures stream-fatal rather than per-event? Why are inbound and outbound queue capacities pinned at exactly 32, and how does that cap backpressure at the application layer? Why is the injected processor sequential, and what does sequential processing guarantee about response ordering when the processor blocks? Why does the 10,000-message limit count per-event processing failures, and why is duplicate-ID tracking bounded by the same limit? Why does the client `CloseSend` lead to the server receive loop observing `io.EOF` and finishing accepted work before returning `nil`, and why does the client receive observe clean `io.EOF` only after the server returns? Why does cancellation or a non-`io.EOF` send or receive failure cancel owned work and prevent any further send, and why is there no per-event response in place of cancellation? Why are `Canceled` and `DeadlineExceeded` not interchangeable and how is precedence observed before mutation? Why are lifecycle tests deterministic through synchronization events rather than raw goroutine counts, and why does the graceful-stop watchdog force-stop only on failure and then fail the test?

## 18. Definition of Completion
- [ ] Generated descriptor exposes package `tutorial.event.v1`, service `EventProcessor`, RPC `Process`, the two messages with pinned field numbers, and the two-value code enum.
- [ ] `Event` has exactly `id=1` and `payload=2`; `EventResult` has exactly `id=1`, `code=2`, `result=3`, `error_message=4`.
- [ ] Per-message validation pins UTF-8 validity, payload length 1..4,096, positive ID, and duplicate ID detection; each failure is stream-fatal with the pinned status and emits no `EventResult` for the offending message.
- [ ] Per-stream count limit is exactly 10,000 otherwise-valid accepted messages and returns `ResourceExhausted` on excess, counting per-event processing failures toward the limit and bounding duplicate-ID tracking by the same limit.
- [ ] Response order is receive order regardless of per-event processor blocking.
- [ ] Inbound and outbound application queue capacities are each exactly 32 and not unbounded.
- [ ] Exactly one owner calls send; receive-order responses are preserved by construction.
- [ ] The injected processor is sequential and returns a deterministic result plus a processing error; per-event processing error maps to `PROCESSING_FAILED` with a generic public message and does not terminate the stream.
- [ ] Half-close and clean `io.EOF` from the client cause the server receive loop to observe `io.EOF`, stop accepting, finish accepted work, send every corresponding `EventResult`, return `nil`, and only then does the client receive observe clean `io.EOF`.
- [ ] Pre-cancelled non-deadline context maps to `Canceled`; expired deadline maps to `DeadlineExceeded`; the two are not interchangeable; cancellation or non-`io.EOF` send or receive failure cancels owned work and prevents any further send.
- [ ] Status mapping preserves codes; stream-fatal statuses are not wrapped and produce no per-event response.
- [ ] Lifecycle tests are deterministic through synchronization events; the graceful-stop test holds a stream at a barrier, starts graceful stop, proves pending, releases or finishes, and proves completion; the watchdog force-stops only on failure and then fails the test.
- [ ] Concurrent streams are isolated.
- [ ] All tests use `bufconn` with finite dial context and deterministic cleanup.
- [ ] All tests pass under the race detector with no sleep synchronization.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions
Add a unary health RPC that reports accepted-message and stream counts for observability without exposing internal addresses. Add a per-event timing field exposed in shutdown logs but never on the wire.
