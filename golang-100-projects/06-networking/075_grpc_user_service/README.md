# Project 075 — gRPC User Service

## 1. Project Name and Number

- Project 075, grpc_user_service.
- This README is a learning guide only.
- You will create every source and test file yourself in `06-networking/075_grpc_user_service/`.
- This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea

A protobuf User service with five unary RPCs: CreateUser, GetUser, ListUsers, UpdateUser, DeleteUser. The service is stored in an in-memory mutex-backed store with monotonic IDs, deterministic list order, and an injected clock. Status mapping is exact. Exactly three unary interceptors cover request ID, panic-to-Internal recovery, and ordinary completion logging. Tests use `bufconn` only and never bind a fixed port.

## 3. Why This Project Now?

- This project requires Project 074 (websocket_live_chat), Project 071 (tcp_echo_server), and Project 060 (graceful_shutdown_web).
- It builds on network lifecycle and shutdown discipline while introducing RPC through the protobuf contract, gRPC status code discipline, the three-interceptor chain, and the `bufconn` in-memory transport.
- It also raises the bar on status mapping, metadata handling, context precedence, and graceful-stop test discipline.

## 4. Prerequisites

- Projects 074, 071, and 060 are required prerequisites because this project is RPC and depends on network lifecycle and graceful shutdown fundamentals.
- Project 061 is optional review only for the reused name and email validation rules.
- The runtime modules are pinned: `google.golang.org/grpc` at version `v1.83.0` and `google.golang.org/protobuf` at version `v1.36.11`.
- The generation provenance pins the code generator plugins used by the learner to `protoc-gen-go` at the version aligned with the protobuf module `v1.36.11` and `protoc-gen-go-grpc` at version `v1.6.2`.
- The command module `google.golang.org/grpc/cmd/protoc-gen-go-grpc` is not a runtime import.
- The learner manages tool versions through an appropriate tooling mechanism.
- This guide does not require a specific `protoc` compiler release.
- The repository must contain no generated code in this README; tool installation and code generation belong to the learner.

## 5. What You Must Know Before Starting

- Know the protobuf type system, the gRPC status code discipline, unary interceptor composition including the exact order of three interceptors, the `bufconn` in-memory transport, `context` cancellation and deadline propagation with exact precedence between Canceled and DeadlineExceeded, Project 061 name and email validation rules, and the race detector.

## 6. Explanation of New Concepts

### Concepts

- The protobuf contract is pinned in prose without a proto snippet.
- The package is `tutorial.user.v1`.
- The service is `UserService`.
- The messages are `User`, `CreateUserRequest`, `GetUserRequest`, `ListUsersRequest`, `ListUsersResponse`, `UpdateUserRequest`, and `DeleteUserRequest`.
- Delete returns the standard `google.protobuf.Empty`.
- There are no optional patch fields; Update is full.

- Stable field names and field numbers are pinned in prose for compatibility.
- The `User` message has field `id` at number 1 as a positive signed-64 integer, field `name` at number 2 as a string, field `email` at number 3 as a string holding the normalized email, field `created_at` at number 4 as a protobuf Timestamp, and field `updated_at` at number 5 as a protobuf Timestamp. `CreateUserRequest` has field `name` at number 1 and field `email` at number 2. `GetUserRequest` has field `id` at number 1. `ListUsersResponse` has field `users` at number 1 as a repeated `User`. `UpdateUserRequest` has field `id` at number 1, field `name` at number 2, and field `email` at number 3. `DeleteUserRequest` has field `id` at number 1.
- Field order on the wire is not significant; the pinned field numbers are significant.

- The store is in-memory with a mutex.
- IDs start at 1 and increase without reuse.
- A checked overflow returns Internal without mutation.
- Create reads the injected clock once and produces two UTC protobuf Timestamps for `created_at` and `updated_at`.
- Update preserves ID and `created_at`, reads the injected clock once for `updated_at`, and writes a full User.
- A backward or equal injected time compared to the stored value is returned honestly without silent adjustment; that is, Update does not bump `updated_at` to a higher value when the injected clock returns a non-monotonic result.
- List ordering is by ID ascending.
- List with zero users contains zero users; the guide does not assert whether an empty generated repeated-field slice is nil after serialization.

- Validation reuses Project 061 exactly.
- The name is trimmed of leading and trailing whitespace and must be nonempty; an empty name after trim is a validation failure.
- The email is trimmed, lowercased, must contain exactly one `@`, must have a nonempty local part, must have a nonempty domain part, and must contain no whitespace; an email that fails any rule is a validation failure.
- The store persists the normalized email.
- Validation runs before any clock read and before any store mutation.
- A nonpositive ID is InvalidArgument.
- Missing user on Get, Update, or Delete is NotFound.
- Duplicate normalized email on Create or Update is AlreadyExists.
- Delete is not idempotent: a second delete of a missing ID is NotFound.
- Update is full; the request supplies ID, name, and email.

- Context precedence is exact.
- A pre-cancelled caller context that is not a deadline maps to Canceled with no mutation.
- A caller context whose deadline is already expired maps to DeadlineExceeded with no mutation.
- Cancellation discovered before any commit or mutation maps appropriately.
- Canceled and DeadlineExceeded are not interchangeable.
- Context is observed before any mutation.

- Clock output is converted to UTC and must be valid for the protobuf Timestamp type.
- An injected time that does not yield a valid protobuf Timestamp is Internal with a generic public message and no mutation.
- Create reads the injected clock once for both timestamps; Update reads the injected clock once for `updated_at`.

- Status mapping is exact.
- InvalidArgument is used for validation failures and nonpositive IDs.
- NotFound is used for missing Get, Update, or Delete targets.
- AlreadyExists is used for normalized email collisions on Create or Update.
- Canceled and DeadlineExceeded are used from caller context.
- Internal with a generic public message is used for panic, ID exhaustion, or unexpected storage fault.
- Status codes are preserved and never wrapped into Unknown.

- Exactly three unary interceptors are pinned.
- The first interceptor handles the request ID.
- The second interceptor handles panic-to-Internal recovery.
- The third interceptor is the ordinary completion logging.
- The chain order is request ID outermost, recovery next, ordinary completion logging innermost.
- The request-ID context enrichment reaches recovery, logging, and the service.

- The request-ID metadata key is `x-request-id`.
- Accepted values are ASCII bytes in the inclusive range 33..126, which is printable non-space, with length 1..64.
- Multiple input values or an invalid value are InvalidArgument before the service is called.
- The accepted or generated effective ID is sent as response header metadata.
- The injected generator supplies a collision-resistant value when absent.
- A generator that returns an empty or invalid value is Internal with no service call.
- Because the request-ID interceptor is outermost and short-circuits on rejection, a rejected request produces no ordinary completion log; only the request-ID rejection path runs.
- This branch is tested explicitly.

- Logging discipline is pinned.
- The ordinary completion logging interceptor records request ID, full method, final status code, and injected duration only.
- It never records request payloads, response payloads, emails, or names.
- A normal call produces exactly one ordinary completion log.
- A panic produces exactly one recovery log and no ordinary completion log.
- Recovery catches panics that originate in the ordinary completion logging interceptor or in the service, and returns generic Internal.
- Recovery does not claim to catch a panic in the outer request-ID interceptor itself.

- Tests use `bufconn` only.
- Every RPC is exercised through `bufconn` with a finite dial context and deterministic cleanup.
- Reflection is disabled by default and, if explored, is only an optional extension.
- The graceful-stop test holds one RPC at an injected barrier, starts `GracefulStop` while that RPC is still active, and uses event synchronization to prove that graceful stop remains pending.
- The test then releases the RPC and proves that both the RPC and `GracefulStop` complete.
- A bounded test watchdog calls force `Stop` only if the proof fails and then fails the test, so the fallback is never mistaken for successful graceful behavior.

- Text-only protocol examples are permitted.
- As a prose shape: a client sends a CreateUser request whose `name` and `email` are valid by Project 061.
- The server returns a User with a positive `id`, the normalized email, and two Timestamps.
- A client sends a GetUser request with an `id`.
- The server returns a User.
- A client sends a ListUsers request.
- The server returns repeated Users ordered by `id`.
- A client sends an UpdateUser request with `id`, `name`, and `email`.
- The server returns the updated User with the same `id` and `created_at` and a fresh `updated_at`.
- A client sends a DeleteUser request with `id`.
- The server returns the standard Empty.
- A second DeleteUser for the same `id` returns NotFound.

## 7. Learning Objective

- Implement a deterministic unary User service with exact protobuf contract, exact Project 061 validation, exact status mapping, exact context precedence, exact monotonic ID discipline, the pinned three-interceptor chain, and tests that pin every RPC, status code, validation branch, metadata branch, and graceful-stop path through `bufconn` without sleep-based synchronization.

## 8. Functional Requirements

1. Protobuf package is `tutorial.user.v1`. Service is `UserService`. Messages are `User`, `CreateUserRequest`, `GetUserRequest`, `ListUsersRequest`, `ListUsersResponse`, `UpdateUserRequest`, and `DeleteUserRequest`. Delete returns `google.protobuf.Empty`.
2. Field numbers are pinned: User `id=1`, `name=2`, `email=3`, `created_at=4`, `updated_at=5`; CreateUserRequest `name=1`, `email=2`; GetUserRequest `id=1`; ListUsersResponse `users=1`; UpdateUserRequest `id=1`, `name=2`, `email=3`; DeleteUserRequest `id=1`.
3. Service has five unary RPCs: CreateUser, GetUser, ListUsers, UpdateUser, DeleteUser.
4. User message has positive signed-64 ID, name, normalized email, created-at Timestamp, updated-at Timestamp.
5. CreateUserRequest has name and email; response is User.
6. GetUserRequest has ID; response is User.
7. ListUsersRequest is empty; response is repeated Users ordered by ID.
8. UpdateUserRequest has ID, name, and email; response is User. Update is full.
9. DeleteUserRequest has ID; response is the standard protobuf Empty.
10. There are no optional patch fields.
11. Store is in-memory with a mutex.
12. IDs start at 1 and increase without reuse.
13. Checked overflow returns Internal without mutation.
14. Create reads the injected clock once and produces two UTC protobuf Timestamps for `created_at` and `updated_at`.
15. Update preserves ID and `created_at` and reads the injected clock once for `updated_at`.
16. A backward or equal injected time compared to the stored value is returned honestly without silent monotonic adjustment.
17. List ordering is by ID ascending.
18. Empty List contains zero users.
19. Validation reuses Project 061: name trimmed and nonempty; email trimmed, lowercased, exactly one `@`, nonempty local, nonempty domain, no whitespace; normalized email is stored.
20. Validation runs before any clock read and before any store mutation.
21. Nonpositive ID is InvalidArgument.
22. Missing user on Get, Update, or Delete is NotFound.
23. Delete is not idempotent; a second delete of a missing ID is NotFound.
24. Duplicate normalized email on Create or Update is AlreadyExists.
25. Pre-cancelled non-deadline caller context maps to Canceled with no mutation.
26. Already-expired caller deadline maps to DeadlineExceeded with no mutation.
27. Cancellation discovered before commit or mutation maps appropriately.
28. Canceled and DeadlineExceeded are not interchangeable.
29. Context is observed before any mutation.
30. Clock output is converted to UTC and must be valid for protobuf Timestamp; an injected time that does not yield a valid Timestamp is Internal with a generic public message and no mutation.
31. Status mapping: InvalidArgument for validation and nonpositive IDs; NotFound for missing Get, Update, or Delete; AlreadyExists for normalized email collision; Canceled and DeadlineExceeded from caller context; Internal with generic public message for panic, ID exhaustion, or unexpected storage fault.
32. Status codes are preserved and never wrapped into Unknown.
33. Exactly three unary interceptors are pinned: request ID, recovery, ordinary completion logging. Chain order is request ID outermost, recovery next, ordinary completion logging innermost.
34. Request-ID metadata key is `x-request-id`. Accepted value is ASCII 33..126 inclusive, length 1..64. Multiple input values or invalid value is InvalidArgument before the service.
35. Accepted or generated effective ID is sent as response header metadata.
36. Injected generator supplies a collision-resistant value when absent. A generator that returns an empty or invalid value is Internal and does not call the service.
37. The request-ID interceptor is outermost and short-circuits on rejection; a rejected request produces no ordinary completion log; this branch is tested explicitly.
38. Recovery catches panics that originate in the ordinary completion logging interceptor or in the service and returns generic Internal; recovery does not claim to catch a panic in the outer request-ID interceptor.
39. Ordinary completion logging records request ID, full method, final status code, and injected duration only.
40. Ordinary completion logging never records request or response payloads, emails, or names.
41. A normal call produces exactly one ordinary completion log.
42. A panic produces exactly one recovery log and no ordinary completion log.
43. Tests use `bufconn` only; no fixed port.
44. Tests use a finite dial context and deterministic cleanup.
45. Reflection is disabled by default and is only an optional extension.
46. The graceful-stop test holds one RPC at an injected barrier, starts `GracefulStop` while the RPC remains active, proves by events rather than sleep that graceful stop is still pending, releases the RPC, and proves completion; a bounded watchdog calls force `Stop` only on failure and then fails the test.

## 9. Inputs and Outputs

### Interface Contract

- Service input is a unary request with a context.
- Service output is a unary response or a gRPC status with code and message.
- The status message for Internal is generic public.
- The store output is the bound service.
- Tests use `bufconn` listeners and clients.

## 10. Rules and Edge Cases

- A non-positive ID is InvalidArgument.
- A missing user on Get, Update, or Delete is NotFound.
- A duplicate normalized email on Create or Update is AlreadyExists.
- A pre-cancelled non-deadline caller context maps to Canceled with no mutation.
- An already-expired caller deadline maps to DeadlineExceeded with no mutation.
- An ID exhaustion attempt returns Internal without mutation.
- A panic that originates in the service or in ordinary completion logging returns Internal with a generic public message.
- The logging interceptor never logs payloads, emails, or names.
- The recovery log is produced once on a panic and replaces rather than duplicates the ordinary completion log.
- The request-ID interceptor short-circuits rejections before the service and produces no ordinary completion log for rejected requests.
- The request-ID effective ID is enriched on the context and sent as response header metadata.
- The chain order is the order of registration.
- The graceful-stop test begins while an RPC is held at an injected barrier, proves the stop remains pending, releases the RPC, and then proves completion; the watchdog fallback force-stops only on failure and fails the test.

## 11. Project Constraints

- Runtime modules are `google.golang.org/grpc v1.83.0` and `google.golang.org/protobuf v1.36.11`.
- The code generator plugins used by the learner are pinned to `protoc-gen-go` at the version aligned with the protobuf module `v1.36.11` and `protoc-gen-go-grpc v1.6.2`.
- The command module is not a runtime import.
- The learner manages tool versions through an appropriate tooling mechanism.
- No specific `protoc` compiler release is required by this guide.
- The repository must contain no generated code in this README; tool installation and code generation belong to the learner.
- No database, no TLS, no auth, no streaming.
- Reflection is disabled by default.

## 12. Design Questions Before Coding

- How is the protobuf Timestamp conversion validated before mutation, and how is a non-monotonic injected time preserved honestly?
- How is the normalized email computed and compared case-insensitively for AlreadyExists?
- How is the signed-64 ID range validated, and how is the overflow check ordered before any mutation?
- How is the injected clock read exactly once per operation?
- How is the metadata key read once and the value validated before default generation?
- How does the request-ID interceptor reject before the service and short-circuit ordinary logging?
- How is the recovery interceptor ordered so that a panic in logging or service never produces a duplicate ordinary log?
- How does an injected RPC barrier prove that `GracefulStop` waits for an active call before completing after release?
- How is the watchdog fallback bounded and used only as a test failure signal rather than as a successful graceful behavior?
- How is `bufconn` wired so test cleanup is deterministic?

## 13. Implementation Milestones

1. Define the protobuf contract in prose for package `tutorial.user.v1`, service `UserService`, and the seven messages including field numbers; generated code belongs to the learner.
2. Define the in-memory store with mutex, monotonic IDs, ordered list, and the Project 061 validation pipeline.
3. Define the validation and status mapping including the exact context precedence between Canceled and DeadlineExceeded.
4. Define the update and delete semantics including non-idempotent delete and full update.
5. Define the request ID interceptor with key, value validation against ASCII 33..126, length 1..64, default generator, response header, and InvalidArgument on multiple or invalid input.
6. Define the recovery interceptor with generic public message and exactly one recovery log for panics from logging or service.
7. Define the ordinary completion logging interceptor with request ID, full method, status code, and injected duration only.
8. Define the chain order as request ID outermost, recovery next, ordinary completion logging innermost, and confirm request-ID enrichment reaches recovery, logging, and service.
9. Define the `bufconn` test transport, finite dial context, and deterministic cleanup.
10. Define the graceful-stop test with an RPC held at an injected barrier, `GracefulStop` started while it is active, event-based proof that the stop remains pending, release and completion proof, and a bounded watchdog that force-stops only on failure and fails the test.
11. Define the full RPC matrix, status matrix, validation matrix, metadata tests, logging branch tests, concurrency tests, and race tests.

## 14. Verification Cases the Learner Must Write

### Required Cases

- `go.mod` declares runtime modules `google.golang.org/grpc v1.83.0` and `google.golang.org/protobuf v1.36.11`. Generation provenance records `protoc-gen-go` aligned with `v1.36.11` and `protoc-gen-go-grpc v1.6.2`. The command module is not a runtime import.
- Generated descriptor exposes package `tutorial.user.v1`, service `UserService`, and the seven messages with the pinned field numbers.
- CreateUser returns a User with positive ID and two Timestamps; the injected clock is read once.
- GetUser returns NotFound for a missing ID and returns the User for an existing ID.
- ListUsers returns repeated Users ordered by ID ascending; empty list returns zero users.
- UpdateUser preserves ID and `created_at`, reads the injected clock once for `updated_at`, returns the updated User, and rejects a duplicate normalized email with AlreadyExists.
- UpdateUser returns the injected time honestly when the injected time is backward or equal to the stored value with no silent monotonic adjustment.
- DeleteUser returns the standard Empty and removes the user.
- A second DeleteUser for the same ID returns NotFound; delete is not idempotent.
- Project 061 validation: empty name after trim is InvalidArgument; email with multiple `@` is InvalidArgument; email with empty local or domain is InvalidArgument; email with whitespace is InvalidArgument; email is normalized on store.
- Validation runs before any clock read and before any store mutation.
- Nonpositive ID on Get, Update, or Delete is InvalidArgument.
- Pre-cancelled non-deadline caller context maps to Canceled with no mutation on Create, Update, and Delete.
- Already-expired caller deadline maps to DeadlineExceeded with no mutation on Create, Update, and Delete.
- Canceled and DeadlineExceeded are not interchangeable in the mapping.
- An injected clock that yields an invalid protobuf Timestamp returns Internal with a generic public message and no mutation.
- Panic in the service returns Internal with a generic public message; recovery produces exactly one recovery log with the request ID; no ordinary completion log is produced.
- ID exhaustion attempt returns Internal without mutation.
- Request-ID interceptor accepts a valid value, rejects a value outside ASCII 33..126, rejects a value with length 0 or greater than 64, rejects multiple input values with InvalidArgument before the service, generates a collision-resistant value when absent, enriches the context, and sends the effective ID as response header metadata.
- A request-ID generator that returns an empty or invalid value yields Internal and does not call the service.
- The request-ID rejection branch produces no ordinary completion log; this branch is tested explicitly.
- Normal calls produce exactly one ordinary completion log.
- Ordinary completion logging records request ID, full method, final status code, and injected duration only; never records payloads, emails, or names.
- Chain order is preserved: request ID enriches context and reaches recovery, logging, and service; recovery catches panics from logging or service and returns generic Internal; recovery does not claim to catch a panic in the outer request-ID interceptor.
- A panic produces exactly one recovery log and no duplicate ordinary log.
- All tests use `bufconn`; no fixed port.
- Tests use a finite dial context and deterministic cleanup.
- Reflection is disabled by default.
- The graceful-stop test holds one RPC at an injected barrier, starts `GracefulStop` while the RPC is active, proves by synchronization events that graceful stop remains pending, releases the RPC, and proves both completions; on watchdog failure it calls force `Stop` and fails rather than reporting success.
- Concurrent creates do not duplicate IDs. Concurrent updates and lists are race-free, and every returned list is an ID-ascending snapshot of the set observed by that call. A test does not assert which competing update wins for the same user unless both updates use the same value or explicit synchronization establishes the order.

## 15. Common Mistakes to Watch For

- Wrapping gRPC status codes into Unknown, treating Canceled and DeadlineExceeded as interchangeable, mutating before context is observed, reusing IDs, allowing ID overflow to mutate, generating a request ID from a non-collision-resistant source, allowing a generator that returns an empty or invalid value to reach the service, sending payloads or emails in logs, logging twice on a panic, registering interceptors in the wrong order, producing an ordinary completion log on a rejected request-ID branch, claiming recovery catches a panic in the outer request-ID interceptor, silently adjusting a non-monotonic injected time on Update, treating delete as idempotent, using a fixed port for tests, mistaking the graceful-stop watchdog fallback for successful graceful behavior, forgetting the protobuf Timestamp conversion validation, and asserting nil-versus-empty on a generated repeated field.

## 16. Topics and References for Study

- Study the protobuf type system, Timestamp conversion, gRPC status codes, exact context precedence between Canceled and DeadlineExceeded, unary interceptor composition and chain order, `bufconn` in-memory transport, graceful-stop discipline with bounded watchdog fallback, Project 061 name and email validation, monotonic ID discipline, and the race detector.
- Review the official gRPC and protobuf documentation for the pinned versions.
- Read the prior READMEs for Projects 074, 071, and 060 as required foundations, including Project 074 for connection lifecycle and pump ownership, and Project 061 as optional review for validation rules.

## 17. Self-Assessment Questions

1. Why are there exactly three unary interceptors in the order request ID outermost, recovery next, and ordinary completion logging innermost?
2. Why does the request-ID interceptor short-circuit and produce no ordinary completion log on rejection?
3. Why does recovery catch panics from logging or service but not from the outer request-ID interceptor?
4. Why is a non-monotonic injected time preserved honestly on Update?
5. Why is delete not idempotent?
6. Why are Canceled and DeadlineExceeded not interchangeable, and how is context precedence observed before mutation?
7. Why is Project 061 validation performed before any clock read and store mutation?
8. Why is the Internal status message generic and public, and why are payloads, emails, and names never logged?
9. Why does the graceful-stop test treat the watchdog fallback as a test failure rather than successful graceful behavior?
10. How can tests prove status mapping, validation branches, metadata branches, logging branches, and graceful-stop discipline without sleep?

## 18. Definition of Completion

- [ ] `go.mod` declares runtime modules `google.golang.org/grpc v1.83.0` and `google.golang.org/protobuf v1.36.11`; generation provenance pins `protoc-gen-go` aligned with `v1.36.11` and `protoc-gen-go-grpc v1.6.2`; the command module is not a runtime import.
- [ ] Generated descriptor exposes package `tutorial.user.v1`, service `UserService`, and the seven messages with the pinned field numbers.
- [ ] Five unary RPCs match the exact protobuf contract; Update is full; Delete returns the standard Empty.
- [ ] Project 061 validation runs before any clock read and before any store mutation; normalized email is stored.
- [ ] Store is in-memory with mutex; IDs are monotonic and never reused; checked overflow returns Internal without mutation.
- [ ] Status mapping matches the pinned codes; status codes are preserved; Canceled and DeadlineExceeded are not interchangeable; panic returns Internal with generic public message.
- [ ] An injected clock that yields an invalid protobuf Timestamp returns Internal with no mutation; a non-monotonic injected time is preserved honestly on Update.
- [ ] Exactly three unary interceptors are pinned in the order request ID, recovery, ordinary completion logging.
- [ ] Request-ID interceptor validates ASCII 33..126 length 1..64, rejects multiple or invalid values with InvalidArgument before the service, generates a collision-resistant default, treats an empty or invalid generator output as Internal with no service call, enriches the context, and sends the effective ID as response header metadata.
- [ ] Recovery interceptor catches panics from logging or service and returns generic Internal; a panic produces exactly one recovery log and no ordinary completion log.
- [ ] Ordinary completion logging interceptor records request ID, full method, final status code, and injected duration only; never records payloads, emails, or names.
- [ ] Request-ID rejection branch produces no ordinary completion log; normal calls produce exactly one ordinary completion log.
- [ ] All tests use `bufconn` with a finite dial context and deterministic cleanup.
- [ ] Reflection is disabled by default and is only an optional extension.
- [ ] Graceful-stop test holds an RPC at an injected barrier, starts `GracefulStop` while it remains active, proves the stop is pending, releases the RPC, and proves completion; the watchdog force-stops only on failure and fails the test.
- [ ] No database, no TLS, no auth, no streaming.
- [ ] All tests pass under the race detector.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions

- Add a streaming RPC extension that preserves the same three-interceptor chain and status mapping.
- Add a per-RPC metric for status codes visible at shutdown.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 074 — WebSocket Live Chat](../../06-networking/074_websocket_live_chat/README.md#20-prerequisite-based-documentation-guide), [Project 071 — TCP Echo Server](../../06-networking/071_tcp_echo_server/README.md#20-prerequisite-based-documentation-guide), [Project 060 — Graceful Shutdown Web](../../04-apis-and-services/060_graceful_shutdown_web/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`google.golang.org/grpc`](https://pkg.go.dev/google.golang.org/grpc).
- **Standards and concept references:** [gRPC for Go](https://grpc.io/docs/languages/go/), [Protocol Buffers proto3 guide](https://protobuf.dev/programming-guides/proto3/), [gRPC status codes](https://grpc.io/docs/guides/status-codes/), [gRPC interceptors](https://grpc.io/docs/guides/interceptors/).

### Project-specific learning focus

- **Learn now:** protobuf type design, timestamp conversion, unary interceptor order, context-status precedence, in-memory bufconn tests, monotonic IDs, and bounded graceful stop.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
