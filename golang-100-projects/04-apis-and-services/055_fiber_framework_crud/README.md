# Project 055 — Fiber Framework CRUD

## 1. Project Name and Number

- Project 055 — Fiber Framework CRUD, located in `055_fiber_framework_crud`.

## 2. Project Idea

Reimplement the Project 047 Notes CRUD contract on top of the Fiber web framework while keeping the domain logic and the in-memory store framework-independent. The Fiber adapter handles routing, binding, response writing, and middleware composition. The route matrix, methods, statuses, validation rules, strict JSON, body cap, identity and order rules, error envelope, and concurrency semantics must match Project 047 exactly. The Fiber dependency is pinned to a specific version in the learner implementation. Verification uses Fiber's in-process app test mechanism rather than a real listening port.

## 3. Why This Project Now?

- Project 054 ported the same contract to Gin and proved the parity approach for one popular framework.
- Project 047 is the contract being preserved.
- Project 046 is the `net/http` foundation whose boundary disciplines remain visible.
- The new step is to map the same Project 047 contract onto Fiber, which uses fasthttp rather than `net/http`, and to recognize honestly what changes about request context lifetime, header and body APIs, error handling, and standard-library middleware compatibility while preserving the external contract.

## 4. Prerequisites

- Complete Projects 054, 047, and 046 before starting.
- Earlier projects may be useful review, but they are not required prerequisites.

- You should already be able to build the Project 047 contract on `net/http` with `httptest`, distinguish the documented status codes, validate JSON strictly, manage an in-memory store with a mutex and copies, inject a deterministic clock, run tests under the race detector, and recognize when a framework's binding defaults are not strict.
- You should also know that Fiber's request and response APIs are built on fasthttp rather than `net/http`, that this has lifetime and interface implications that must be respected, and that Project 047's optional `body` field is allowed to be absent.

## 5. What You Must Know Before Starting

- The Project 047 contract is the source of truth. Route paths, methods, statuses, headers, body policies, validation rules, ID semantics, timestamp semantics, error envelope, and concurrency semantics are inherited without silent change. The Project 047 semantics for the optional `body` field are preserved: `body` may be omitted from create and full-replacement requests, and an omitted `body` becomes an empty stored body. An absent entire request document, a missing `title`, wrong types, unknown fields, and a trailing second JSON value are `400`.
- Fiber is built on fasthttp rather than `net/http`. Request context lifetime, header access, and body access differ from `net/http`. The adapter must copy any value it intends to use outside the handler into a standard-library data structure rather than retaining framework-owned pointers or buffers.
- A standard `net/http` middleware cannot be inserted directly into a Fiber chain. The adapter must use Fiber-compatible middleware signatures and must not pretend that a `net/http` middleware is interchangeable. Any reuse of middleware concepts from Project 048 or Project 054 must be rewritten for Fiber's middleware shape.
- Fiber's default JSON binder may not enforce strict JSON parsing. The adapter must configure strictness explicitly and must enforce the exact Project 047 body cap of exactly 1,048,576 bytes before binding, so an oversized request maps to the documented `413` without consuming binder memory.
- The body cap of exactly 1,048,576 bytes must be enforced before binding using a `BodyLimit`-style configuration or a middleware that wraps the body with a bounded reader. The cap value is the same as Project 047.
- Routes, methods, statuses, and `Allow` values match Project 047 exactly. The collection allows `GET, POST`; the item allows `DELETE, GET, PUT`. Unknown paths return `404`. `HEAD` returns the corresponding `405` policy with no body.
- Fiber's in-process app test mechanism exercises routes without binding a port. Tests do not bind a fixed port and do not depend on external networking. The adapter test entry point is the framework's in-process app test rather than a `httptest`-style server, although a parallel contract table can still use any in-memory test harness that respects the contract.
- The clock used for `created_at` and `updated_at` is read inside the store's mutation critical section. The contract requires exactly one clock read per successful creation and exactly one clock read per successful replacement, taken while the store lock is held.
- Middleware in this project is request-ID and error-boundary middleware with a deterministic injected ID source and a deterministic injected error reporter. The chain is composed once during setup and frozen before serving. The request-ID response header produced by the Fiber adapter is an explicitly additive, adapter-specific header; it is not part of the Project 047 envelope. The shared contract table compares every Project 047-required status, body, and header for both adapters. The request-ID response header is documented as additive and reconciled using the same approach as Project 054.
- The store mutex and concurrency semantics from Project 047 remain unchanged. The race detector must pass against the same store, and the adapter must not introduce a second store-owned state.
- Verification runs the same framework-neutral contract cases as Project 047 and Project 054, plus Fiber-specific lifecycle and error-middleware cases. The contract table asserts the same status, headers, body, content type, error envelope, ID, and timestamp for the same input.
- A comparison of the two adapters is provided as prose based on observed facts from this repository. The comparison covers adapter code shape, middleware model, test API, context compatibility, and dependency footprint. It does not assert universal speed or a winner without benchmarks; benchmarks are not correctness tests.

## 6. Explanation of New Concepts

### Concepts

- A fasthttp-based framework uses zero-allocation or low-allocation request handling through reusable buffers.
- A fasthttp request and response have different lifetimes than `net/http` request and response objects.
- Buffers returned by the framework may be reused after the handler returns.
- Any value the adapter retains after the handler returns must be copied out of the framework-owned buffer into standard-library storage.
- Headers, body bytes, and URL components are all subject to this lifetime concern.

- Fiber provides routing, middleware, and response helpers built on fasthttp.
- The `fiber.Ctx` value carries the per-request state.
- It exposes the request body, headers, route parameters, query parameters, and response methods.
- The adapter extracts these values into standard-library types before invoking domain logic.

- Fiber's default JSON binder does not reject unknown fields or trailing second values.
- The adapter must enable strictness explicitly, either through binder configuration or through an explicit decoder that performs strict decoding.
- The strictness is part of the contract, not an assumed default.

- Body limits in Fiber are typically enforced through configuration such as `BodyLimit`.
- The cap value is exactly 1,048,576 bytes, the same as Project 047.
- A request whose body exceeds the configured limit is rejected with the framework's standard response or with a custom error middleware that maps the rejection to the documented `413` status and JSON error envelope.

- Fiber middleware is a function that takes a `fiber.Ctx` and returns an error.
- Middleware can call `c.Next()` to invoke the next handler in the chain, then run post-processing code on the way out.
- The chain is composed during setup and frozen before serving.
- The chain is composed in Fiber's middleware shape; a `net/http` middleware cannot be inserted into it directly.

- The Project 047 contract for the optional `body` field is preserved.
- Project 047 accepts a JSON object that may contain only `title`, may contain both `title` and `body`, and treats an absent `body` as an empty stored body.
- An absent entire document, missing `title`, wrong types, unknown fields, and trailing second values are `400`.
- The Fiber adapter preserves these semantics: it does not collapse "absent optional `body`" into a `400`.

- The Project 047 contract is documented in terms of route matrix, statuses, headers, validation, identity, ordering, timestamps, error envelope, and concurrency.
- The Fiber adapter must reproduce all of them.
- A deviation in any one of these dimensions is a contract failure, not a stylistic choice.

## 7. Learning Objective

- By completion, you can port a precise `net/http` contract to a fasthttp-based framework without changing the contract, including the Project 047 optional `body` field semantics, recognize the lifetime differences between fasthttp and `net/http` buffers and apply explicit copying where needed, configure Fiber binding to enforce strict JSON and the exact 1,048,576-byte body cap, compose Fiber middleware for request ID and error boundary with deterministic sources, run the same contract table against both the standard-library and framework handlers, isolate the parity baseline from a single explicitly additive adapter header, and pin a framework dependency to a specific version.
- You can also produce a comparison of adapter approaches based on observed facts, without claiming universal performance differences that benchmarks have not established.

## 8. Functional Requirements

1. Reproduce the Project 047 Notes CRUD contract on Fiber. The route matrix, methods, statuses, headers, content types, body policies, validation rules, ID semantics, timestamp semantics, error envelope, and concurrency semantics match Project 047 exactly. The Project 047 semantics for the optional `body` field are preserved: `body` may be omitted, and an omitted `body` stores as an empty string.
2. Pin the Fiber dependency to a specific version in the learner implementation. The pinned version is recorded in the module manifest and is not replaced by a moving tag during testing.
3. Keep the domain and store framework-independent. The store implements the same interface used in Project 047 and is injected into the Fiber adapter without modification. The Fiber adapter calls framework-independent functions or methods on the domain.
4. Configure JSON binding for strictness. Unknown JSON fields are rejected with `400`. A trailing second JSON value after the first object is rejected with `400`. Wrong field types are rejected with `400`. The strictness is asserted by the contract table, not assumed from the binder default.
5. Enforce the exact Project 047 body cap of exactly 1,048,576 bytes through `BodyLimit` or middleware. A request whose body exceeds the cap returns `413` before any domain mutation. The cap is part of the contract.
6. Map media-type validation to the documented `415` policy with the documented JSON error envelope.
7. Map ID parsing to the documented canonical positive decimal rules from Project 047. An invalid ID is `400`. A missing item is `404`.
8. Map list output to ascending ID order. Empty list is an empty JSON array, not a JSON null value. Returned values are copies that cannot mutate store-owned state.
9. Map replacement to a full update that preserves ID and `created_at`, replaces `title` and `body`, and assigns a new `updated_at` from the injected clock. The clock is read inside the store's mutation critical section.
10. Map deletion to `204 No Content` with zero body bytes, no `Content-Type` header, and no JSON envelope.
11. Map unsupported methods to `405` with sorted `Allow` headers that match Project 047. `HEAD` returns the corresponding known-path `405` policy with no body.
12. Compose middleware in this order: request ID, error boundary, route dispatch. The middleware chain is composed in Fiber's middleware shape. The request-ID source is injected and deterministic. The error boundary reports once through an injected boundary and never writes a second response after commitment.
13. Treat the request-ID response header emitted by the Fiber adapter as an explicitly additive, adapter-specific header. It is not part of the Project 047 envelope. The shared parity comparison either ignores only that documented additive header or wraps the standard-library handler with an equivalent test middleware that adds the same header.
14. Use Fiber's in-process app test mechanism rather than a real listening port. No test binds a fixed port or depends on external networking.
15. Copy any framework-owned buffer or value out of the handler scope before retaining it. The handler does not retain pointers to fasthttp-owned memory beyond the handler return. Standard-library copies are used wherever the contract requires retention.
16. The Fiber adapter must not pretend that a `net/http` middleware is interchangeable. Any reuse of middleware concepts from earlier projects is rewritten for Fiber's middleware shape.
17. Verification runs the same framework-neutral contract cases as Project 047 and Project 054, plus Fiber-specific lifecycle and error-middleware cases. The contract table asserts identical status, headers, body, content type, error envelope, ID, and timestamp for the same input.
18. Provide a comparison in prose based on observed project facts. The comparison covers adapter code shape, middleware model, test API, context compatibility, dependency footprint measured in this repository, and other observed items. It does not assert universal speed or a winner without benchmarks; benchmarks are not correctness tests.
19. The optional additions must not change the parity baseline. They are documented separately and are not part of the required contract.

## 9. Inputs and Outputs

### Interface Contract

- The request inputs are identical to Project 047.
- The collection path is exactly `/notes`; the item path is exactly `/notes/{id}`.
- The request document is a JSON object with required string field `title` and optional string field `body`, no other fields.
- Omitting `body` is permitted and stores an empty body.
- The JSON body cap is exactly 1,048,576 bytes.

- Text-only create example: a `POST /notes` with a JSON body whose `title` is whitespace-padded and `body` is absent returns `201`, the exact JSON note representation with empty `body`, `Content-Type: application/json; charset=utf-8`, and a `Location` header whose value is `/notes/{id}` for the new item.
- The behavior matches Project 047.

- Text-only optional-body replacement example: a `PUT /notes/{id}` with a JSON body that contains `title` and no `body` preserves the item's ID and `created_at`, replaces title, replaces body with the empty string, and assigns a new `updated_at` from the injected clock.
- The behavior matches Project 047.

- Text-only list example: a `GET /notes` returns `200` and a JSON array sorted by ascending ID.
- The behavior matches Project 047.

- Text-only error examples: a missing title returns `400` with code `invalid_request`; a missing item returns `404` with code `not_found`; a wrong media type returns `415` with code `unsupported_media_type`; an over-limit body returns `413` with code `payload_too_large`; an unsupported method returns `405` with sorted `Allow` and code `method_not_allowed`; an internal store error returns the documented service error envelope.
- The error envelope shape, codes, and messages match Project 047.

- The successful note JSON contains exactly `id`, `title`, `body`, `created_at`, and `updated_at`.
- Timestamps use the same UTC, RFC3339Nano format with a literal `Z` suffix as Project 047.
- The first create reading supplies both timestamps for the same item.

## 10. Rules and Edge Cases

- The strict JSON contract is enforced inside the Fiber adapter.
- An unknown field, a trailing second JSON value, a wrong type, a missing required `title`, an empty title after trimming, a non-string `title`, an absent entire request document, and a malformed JSON value are all rejected with the documented `400` policy.
- An omitted optional `body` field is accepted and stored as an empty string on create and on full replacement.

- The body cap of exactly 1,048,576 bytes is enforced before binding so an oversized body is rejected with `413` and no store mutation occurs.
- The cap is part of the contract.

- The ID parser accepts a canonical positive decimal segment.
- A segment that is empty, zero, negative, signed, leading-zero, overflow, or non-decimal is rejected with `400`.
- An item-shaped path with an extra trailing slash or extra segment returns `404`.

- Concurrency semantics are inherited from Project 047.
- Concurrent creates receive unique IDs, list output is sorted, and individual reads return copies.
- The store mutex protects ID allocation and note updates.
- The clock is read inside the same mutation critical section that performs the state transition it timestamps, and the read happens at most once per successful creation and at most once per successful replacement.
- The Fiber adapter does not introduce additional shared state.

- The middleware chain is composed once during setup and frozen before serving, all in Fiber's middleware shape.
- The request-ID source is deterministic in tests.
- The request-ID middleware emits its response header on Fiber responses; that header is adapter-specific additive and is not part of the Project 047 envelope.
- The error boundary reports a single error through an injected boundary on failure paths and never writes a second response after commitment.

- Fiber-owned buffers must be copied out before retention.
- The handler does not retain pointers to fasthttp-owned body bytes, header values, or URL components beyond the handler return.
- Standard-library copies are used wherever the contract requires retention.

- `HEAD` returns the corresponding known-path `405` policy with no body.
- Unknown paths return `404` with the documented envelope.
- The `Allow` header is sorted for `405` and matches Project 047.

- A failed validation, an unsupported media type, a malformed JSON, or an over-limit body does not allocate an ID, read the creation clock, or mutate any existing note.
- An unknown update does not read a replacement timestamp.
- An unknown delete does not alter the ID sequence or any note.

## 11. Project Constraints

- Use only the Go standard library plus exactly one pinned external dependency: the Fiber framework.
- Pin the dependency to a specific version in the learner module configuration.
- Do not add another router, another binder, another JSON library, another web framework, an ORM, a database, or a logging library beyond what Fiber provides.
- Do not use fasthttp directly without going through Fiber for request handling.

- Do not include implementation code, function signatures, exact byte sequences of error messages, or exact field names of the JSON envelope in this guide.
- The guide states policies, not literal values.

## 12. Design Questions Before Coding

- Where does the boundary between the Fiber adapter and the framework-independent domain live, and what types cross it?
- How are fasthttp-owned buffers copied into standard-library types before any value is retained outside the handler?
- How is the JSON binder configured for strictness, and what explicit options are required?
- How is the body cap of exactly 1,048,576 bytes enforced before binding, and how is an over-limit body mapped to the documented `413`?
- How is the Project 047 optional `body` field preserved across the Fiber adapter so that omission succeeds and stores an empty body?
- How is the route matrix reproduced exactly, including the `Allow` header values and the item-shaped path?
- How is the request-ID middleware composed with the error-boundary middleware in Fiber's middleware shape, and how is the additive request-ID response header reconciled with the Project 047 parity baseline?
- How is Fiber's in-process app test mechanism used without binding a fixed port or depending on external networking?
- How is the contract table constructed so the same input runs against both the Project 047 standard-library handler and the Fiber handler, and how is the additive request-ID header normalized in the comparison?
- How is the store injected into the Fiber handler without coupling it to the framework?
- How is the dependency version pinned in the module manifest so a moving tag does not change behavior?

## 13. Implementation Milestones

1. Record the Project 047 contract as a table of routes, methods, statuses, headers, content types, body policies, validation rules, ID semantics, timestamp semantics, error envelope, and concurrency semantics. Note the explicit treatment of the additive request-ID response header.
2. Pin the Fiber dependency to a specific version in the learner module configuration.
3. Build the framework-independent store and domain logic with the same interface used in Project 047.
4. Configure the Fiber binder for strict JSON parsing and bind the request body cap to the exact Project 047 value of 1,048,576 bytes through `BodyLimit` or middleware.
5. Map the collection routes, the item routes, and the method dispatch onto the Fiber app without changing the contract. Preserve the Project 047 optional `body` semantics.
6. Compose the request-ID middleware, the error-boundary middleware, and the route dispatch middleware in the documented order, all in Fiber's middleware shape. Document the additive request-ID response header.
7. Map the response writing for `201 Created`, `200 OK`, `204 No Content`, `404`, `405`, `400`, `413`, `415`, and the documented service errors. Copy any fasthttp-owned buffer into standard-library storage before retention.
8. Use Fiber's in-process app test mechanism for tests. Ensure no test binds a fixed port or contacts external services.
9. Build the contract table and run it against both the Project 047 standard-library handler and the Fiber handler. Reconcile the additive request-ID response header using either header-ignore for that one documented header or a wrapping test middleware on the standard-library handler. Run Fiber-specific lifecycle and error-middleware cases.
10. Write the prose comparison based on observed facts. Cover adapter code shape, middleware model, test API, context compatibility, and dependency footprint. Do not claim universal speed or a winner without benchmarks.
11. Finish deterministic, boundary, concurrency, and race-detector verification with the race detector.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Run the contract table against both the Project 047 standard-library handler and the Fiber handler with the same inputs. Verify identical Project 047 status, content type, `Location`, `Allow`, body shape, body bytes, ID, timestamp format, error code, and error message for every case. Reconcile the additive request-ID response header by either ignoring only that documented additive header or wrapping the standard-library handler with the same additive header.
- Submit a JSON body with `title` only and no `body` for create. Verify `201` with empty `body` stored. Submit the same body shape for replacement and verify the stored body becomes empty and `created_at` is preserved while `updated_at` advances. Verify the omission of optional `body` is not treated as `400`.
- Submit a JSON body with an unknown field, a trailing second JSON value, a wrong type, a missing `title`, an empty `title`, a non-string `title`, an absent entire request document, and a non-object value. Verify the documented `400` behavior from the Fiber adapter.
- Submit a body of exactly 1,048,576 bytes that is otherwise valid and verify it is accepted. Submit one additional byte and verify `413` from the Fiber adapter before any store mutation.
- Submit a missing media type and an unrelated media type. Verify `415` from the Fiber adapter.
- Submit an invalid ID and a missing item. Verify `400` and `404` from the Fiber adapter.
- Run concurrent creates, reads, updates, and deletes against the Fiber adapter under the race detector. Verify unique IDs, sorted list output, copies for read operations, and that the clock is read once per successful create and once per successful replacement inside the store's mutation critical section.
- Inspect the dependency manifest and verify the Fiber version is pinned to a specific value, not to `latest`.
- Verify the Fiber in-process app test mechanism is used and that no test binds a fixed port or contacts external services.
- Verify the request-ID middleware produces a deterministic ID from the injected source, that the additive request-ID response header is documented as adapter-additive rather than as part of the Project 047 envelope, and that the error-boundary middleware reports failures once and never writes a second response after commitment.
- Verify Fiber-specific lifecycle cases: server startup after route composition, request lifetime during middleware chain execution, and shutdown of the test app after the test run. Verify no framework-owned buffer leaks past the handler scope.
- Verify the prose comparison is grounded in observed facts from this repository and does not claim universal performance differences without benchmarks.
- Verify optional additions are documented separately and do not change the parity baseline.
- Verify every test uses the in-process test mechanism or an equivalent in-memory harness, no test binds a fixed port, and no test contacts external services.

## 15. Common Mistakes to Watch For

- Assuming the framework's default JSON binding is strict accepts unknown fields and trailing second values.
- Treating an absent optional `body` as `400` rejects requests that Project 047 accepts.
- Assuming the framework applies a Project-047-equivalent body cap silently accepts oversized requests.
- Inferring `Content-Type` from body bytes changes the contract and hides 415 vs 400 distinctions.

- Retaining pointers to fasthttp-owned buffers past the handler return reads memory that the framework reuses.
- Headers, body bytes, and URL components must be copied out of the framework before retention.
- Treating a Fiber context value as a long-lived reference violates the framework's lifetime contract.

- Treating a `net/http` middleware as interchangeable with a Fiber middleware breaks the chain.
- Standard `net/http` middleware cannot be inserted directly into a Fiber app.
- Reusing middleware concepts is fine, but the implementation must use Fiber's middleware shape.

- Reimplementing the domain inside the Fiber handler couples the contract to the framework and prevents the contract table from running against both handlers.
- Returning pointers or slices backed by store memory breaks the copy contract.
- Incrementing IDs before validation creates gaps on failed requests.
- Reading the clock outside the store's mutation critical section changes the timestamp semantics and breaks Project 047 parity.

- Composing middleware in a different order changes the request ID visibility and the error boundary coverage.
- Letting the framework's recovery middleware run inside the contract changes the panic-to-response behavior.
- Using a panic recovery that writes a second response after commitment hides the original error.
- Treating the additive request-ID response header as part of the Project 047 envelope breaks the parity table without making the additive header itself wrong.

- Returning `200` for creation, a body for deletion, no `Location` header on `201`, or implicit `HEAD` from `GET` violates the parity baseline.
- Returning different `Allow` values, different error codes, or different status codes for the same input breaks the contract table.

- Skipping the dependency pin and resolving Fiber to `latest` introduces non-deterministic versions into tests.
- Using a non-pinned module path that points at a moving tag changes binding behavior between runs.

- Claiming universal speed superiority or framework superiority without benchmarks overstates what is observed.
- A comparison that lists observed facts about adapter shape, middleware model, test API, and dependency footprint is honest; a comparison that asserts a universal winner is not.

## 16. Topics and References for Study

- Study the official Fiber documentation, especially `fiber.App`, `fiber.Ctx`, `fiber.Router`, `BodyLimit`, the JSON binder options, middleware signatures, and the documented in-process app test mechanism.
- Study fasthttp's request and response lifetime, buffer reuse, and the recommended pattern of copying values before retention.
- Study `encoding/json` decoder options for strict decoding.
- Study the official Project 047 README and contract to understand the exact parity target, including the optional `body` field and the body cap.
- Review the official Project 054 README and the prose comparison principles used there, especially how a single additive adapter header is reconciled in the parity comparison without changing the Project 047 envelope.

## 17. Self-Assessment Questions

1. Why must the Project 047 contract be reproduced exactly rather than re-interpreted on Fiber, including the optional `body` field semantics?
2. Why is the domain kept framework-independent, and what does the boundary look like?
3. Why must fasthttp-owned buffers be copied into standard-library types before retention?
4. Why must strict JSON and the exact 1,048,576-byte body cap be enforced explicitly before binding rather than assumed from framework defaults?
5. Why is the clock read inside the store's mutation critical section, and why is at most one clock reading taken per successful create or replace?
6. Why must the middleware chain be frozen at setup, and why is the race detector still required despite framework concurrency primitives?
7. Why is the additive request-ID response header documented as adapter-specific instead of being claimed as part of the Project 047 envelope, and how is it reconciled in the parity comparison?
8. Why is the Fiber dependency pinned to a specific version rather than to `latest`?
9. Why is the in-process app test mechanism used instead of binding a real port?
10. Why is the prose comparison grounded in observed facts rather than universal performance claims?

## 18. Definition of Completion

- [ ] The Fiber adapter reproduces the Project 047 contract for routes, methods, statuses, headers, content types, validation, body cap of exactly 1,048,576 bytes, ID semantics, timestamps, error envelope, and concurrency.
- [ ] The optional `body` field preserves Project 047 semantics: omitted `body` is accepted on create and replacement and stores an empty body; absent entire document, missing `title`, wrong types, unknown fields, and trailing second values are `400`.
- [ ] The Fiber dependency is pinned to a specific version in the learner module manifest.
- [ ] JSON binding is configured for strict decoding and rejects unknown fields, trailing second values, and wrong types with the documented `400` policy.
- [ ] The body cap is enforced before binding through `BodyLimit` or middleware and maps oversized requests to `413`.
- [ ] The request-ID and error-boundary middleware are composed in the documented order with injected sources and boundaries, in Fiber's middleware shape; the request-ID response header is documented as additive and reconciled in the parity comparison.
- [ ] The clock is read inside the store's mutation critical section: at most once per successful creation and at most once per successful replacement.
- [ ] No fasthttp-owned buffer or value is retained past the handler return without being copied into standard-library storage.
- [ ] Tests use Fiber's in-process app test mechanism. No test binds a fixed port and no test contacts external services.
- [ ] The contract table runs against both the Project 047 standard-library handler and the Fiber handler with identical results for every Project 047-required status, header, and body, plus Fiber-specific lifecycle and error-middleware cases.
- [ ] The prose comparison is grounded in observed facts from this repository and does not claim universal performance differences without benchmarks.
- [ ] Concurrent requests pass the race detector with the same store and clock as Project 047.
- [ ] The domain and store are framework-independent and remain usable by other adapters in later projects.
- [ ] The implementation uses only the standard library plus the pinned Fiber dependency, and the learner can explain each parity decision.

## 19. Optional Extensions

- Add a structured access-log middleware that records request ID, method, path, status, and duration using an injected logger, without changing the response contract.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 054 — Gin Framework CRUD](../../04-apis-and-services/054_gin_framework_crud/README.md#20-prerequisite-based-documentation-guide), [Project 046 — Basic HTTP Server](../../04-apis-and-services/046_basic_http_server/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`github.com/gofiber/fiber/v3`](https://pkg.go.dev/github.com/gofiber/fiber/v3), [`github.com/valyala/fasthttp`](https://pkg.go.dev/github.com/valyala/fasthttp).
- **Standards and concept references:** [Fiber documentation](https://docs.gofiber.io/), [Fiber application testing API](https://docs.gofiber.io/api/app/#test).

### Project-specific learning focus

- **Learn now:** handler and middleware lifetimes, body limits, strict JSON, buffer reuse and copying, in-process tests, framework-specific behavior, and API parity.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
