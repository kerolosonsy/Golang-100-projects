# Project 054 — Gin Framework CRUD

## 1. Project Name and Number

- Project 054 — Gin Framework CRUD, located in `054_gin_framework_crud`.

## 2. Project Idea

Reimplement the Project 047 Notes CRUD contract on top of the Gin web framework while keeping the domain logic and the in-memory store framework-independent. The Gin adapter handles routing, binding, response writing, and middleware composition. The route matrix, methods, statuses, validation rules, strict JSON, body cap, identity and order rules, error envelope, and concurrency semantics must match Project 047 exactly. The Gin dependency is pinned to a specific version in the learner implementation rather than to a moving `latest` tag.

## 3. Why This Project Now?

- Project 053 finished the short-link service whose interface-based store and strict validation discipline this project carries forward, and Project 047 is the exact contract this project must reproduce.
- Project 046 is the `net/http` foundation whose boundary disciplines remain visible inside the Gin adapter.
- The new step is to map the Project 047 contract onto a framework that performs its own binding and routing, while proving that the framework's defaults do not silently change behavior and that the same external contract emerges from a different routing layer.

## 4. Prerequisites

- Complete Projects 053, 047, and 046 before starting.
- Earlier projects may be useful review, but they are not required prerequisites.

- You should already be able to build the Project 047 contract on `net/http` with `httptest`, distinguish `201`, `204`, `400`, `404`, `405`, `413`, and `415` semantically, validate JSON strictly, manage an in-memory store with a mutex and copies, inject a deterministic clock, and run tests under the race detector.
- You should also understand the difference between framework binding defaults and explicit strictness, why pinning a dependency version matters, and how Project 047's optional `body` field is allowed to be absent.

## 5. What You Must Know Before Starting

- The Project 047 contract is the source of truth. Route paths, methods, statuses, headers, body policies, validation rules, ID semantics, timestamp semantics, error envelope, and concurrency semantics are inherited without silent change. The Project 047 semantics for the optional `body` field are preserved: `body` may be omitted from create and full-replacement requests, and an omitted `body` becomes an empty stored body. An absent entire request document, a missing `title`, wrong types, unknown fields, and a trailing second JSON value are `400`. The framework defaults are not assumed to match Project 047.
- The domain and store are framework-independent. The same store interface used by Project 047 must be usable inside the Gin adapter without modification. The Gin adapter does not replace the store; it adapts HTTP into store calls.
- Gin binding parses the request body into a Go value using a binder that may accept unknown fields and may not enforce strict JSON parsing by default. The adapter must configure strictness so that unknown fields, trailing second JSON values, and wrong types map to the documented `400` policy.
- A request body cap of exactly 1,048,576 bytes is part of the Project 047 contract. The adapter must enforce the cap before binding so an oversized request maps to `413` without consuming binder memory. The cap is the same number as Project 047, not a framework default.
- Routes, methods, statuses, and `Allow` values must match Project 047 exactly. The collection allows `GET, POST`; the item allows `DELETE, GET, PUT`. Unknown paths return `404`. `HEAD` returns the corresponding `405` policy with no body.
- The Gin test mode suppresses the framework's noisy request logging. The mode affects only logging verbosity and does not change routing or binding behavior. Even though the mode is global to Gin, the test-mode change must be performed exactly once in non-parallel package-level setup before any engine or test is created, and it must never be toggled again from inside parallel tests. Constructing a fresh engine alone does not isolate Gin's global mode; the global mode applies regardless of how many engines exist. Parallel tests must not mutate the global mode.
- The clock used for `created_at` and `updated_at` is read inside the store's mutation critical section. The contract requires exactly one clock read per successful creation and exactly one clock read per successful replacement, taken while the store lock is held.
- Middleware in this project is request-ID and error-boundary middleware with a deterministic injected ID source and a deterministic injected error reporter. The middleware chain is composed in a documented order and does not use mutable globals that parallel tests can race over. The request-ID response header produced by the Gin adapter is an explicitly adapter-added header; it is not part of the Project 047 envelope. The shared contract table compares every Project 047-required status, body, and header for both adapters. The request-ID response header is documented separately and verified in adapter-specific cases. The parity baseline does not silently change the Project 047 JSON envelope or any other Project 047 header.
- The store mutex and concurrency semantics from Project 047 remain unchanged. The race detector must pass against the same store, and the adapter must not introduce a second store-owned state.
- Verification is a contract table run against both the Project 047 standard-library handler and the Gin handler where practical. The contract table asserts the same Project 047 status, headers, body, content type, error envelope, ID, and timestamp for the same input.
- Optional additions do not change the parity baseline. They are documented separately and are not part of the required contract.

## 6. Explanation of New Concepts

### Concepts

- A web framework provides a router, a binder, a response writer, and a middleware chain.
- Gin uses a `gin.Engine` as the router and `gin.Context` as the per-request value.
- The router dispatches based on method and path patterns.
- The binder decodes the request body into a Go value, applying rules about required fields, allowed types, and validation tags.

- Default binding in Gin does not reject unknown JSON fields unless the binder is configured to do so.
- The default JSON binding also does not reject a trailing second JSON value after the first object.
- The adapter must enable strictness through explicit options or through an explicit decoder that performs strict decoding.

- The body cap of exactly 1,048,576 bytes is enforced through a `MaxBytesReader`-style middleware or through reading the body into a buffer with a size check before binding.
- The cap is part of the contract and must map to the documented `413` status with the documented JSON error envelope.

- Gin's test mode is a single global switch that affects logging verbosity for every engine.
- Setting the mode once during non-parallel package-level setup ensures that parallel tests do not race on the global state.
- Constructing a fresh `gin.Engine` per test isolates the engine itself but does not change Gin's global mode; if the global mode is toggled from inside a parallel test, the change can race with other parallel tests and with the package-level setup.
- The parity baseline therefore requires one-and-only-one mode assignment outside any parallel test.

- Middleware is a function that takes a handler and returns a handler.
- The chain is built once during setup and frozen before serving.
- The request-ID middleware assigns an ID, the error-boundary middleware catches panics and unmapped errors, and the route handler is the innermost layer.
- The chain unwinds in reverse order on the way out.
- The Gin adapter's request-ID middleware emits a documented response header on Gin responses; that header is adapter-specific and is not part of the Project 047 envelope.
- The shared contract table checks every Project 047-required header for both adapters;
- The additive request-ID header is allowed to differ because it is explicitly documented as additive, and one of two reconciliation strategies is used: either ignore only the documented additive request-ID header when comparing baseline Project 047 parity, or wrap the standard-library handler with an equivalent test middleware that adds the same header.

- The Project 047 contract for the optional `body` field is preserved.
- Project 047 accepts a JSON object that may contain only `title`, may contain both `title` and `body`, and treats an absent `body` as an empty stored body.
- An absent entire document, missing `title`, wrong types, unknown fields, and trailing second values are `400`.
- The Gin adapter preserves these semantics: it does not collapse "absent optional `body`" into a `400`.

- The Project 047 contract is documented in terms of route matrix, statuses, headers, validation, identity, ordering, timestamps, error envelope, and concurrency.
- The Gin adapter must reproduce all of them.
- A deviation in any one of these dimensions is a contract failure, not a stylistic choice.

## 7. Learning Objective

- By completion, you can port a precise `net/http` contract to a third-party framework without changing the contract, including the Project 047 optional `body` field semantics, configure framework binding to enforce strict JSON, the exact 1,048,576-byte body cap, and validation rules, compose middleware for request ID and error boundary with deterministic sources, run the same contract table against both standard-library and framework handlers, isolate the parity baseline from a single explicitly additive adapter header, set Gin test mode exactly once in non-parallel package-level setup and not toggle it from parallel tests, and pin a framework dependency to a specific version.
- You can also explain why the domain remains framework-independent, why framework defaults are not assumed to match a contract, why the clock-under-mutation-lock rule carries into the adapter, and why the additive request-ID header does not break parity when reconciliation is explicit.

## 8. Functional Requirements

1. Reproduce the Project 047 Notes CRUD contract on Gin. The route matrix, methods, statuses, headers, content types, body policies, validation rules, ID semantics, timestamp semantics, error envelope, and concurrency semantics match Project 047 exactly. The Project 047 semantics for the optional `body` field are preserved: `body` may be omitted, and an omitted `body` stores as an empty string.
2. Pin the Gin dependency to a specific version in the learner implementation. The pinned version is recorded in the module manifest and is not replaced by a moving tag during testing. Tests must not depend on `latest`.
3. Keep the domain and store framework-independent. The store implements the same interface used in Project 047 and is injected into the Gin adapter without modification. The Gin adapter calls framework-independent functions or methods on the domain.
4. Configure JSON binding for strictness. Unknown JSON fields are rejected with `400`. A trailing second JSON value after the first object is rejected with `400`. Wrong field types are rejected with `400`. The strictness is asserted by the contract table, not assumed from the binder default.
5. Enforce the exact Project 047 body cap of exactly 1,048,576 bytes. A request whose body exceeds the cap returns `413` before any domain mutation. The cap is part of the contract.
6. Map media-type validation to the documented `415` policy with the documented JSON error envelope.
7. Map ID parsing to the documented canonical positive decimal rules from Project 047. An invalid ID is `400`. A missing item is `404`.
8. Map list output to an ascending ID order. Empty list is an empty JSON array, not a JSON null value. Returned values are copies that cannot mutate store-owned state.
9. Map replacement to a full update that preserves ID and `created_at`, replaces `title` and `body`, and assigns a new `updated_at` from the injected clock. The clock is read inside the store's mutation critical section.
10. Map deletion to `204 No Content` with zero body bytes, no `Content-Type` header, and no JSON envelope.
11. Map unsupported methods to `405` with sorted `Allow` headers that match Project 047. `HEAD` returns the corresponding known-path `405` policy with no body.
12. Compose middleware in this order: request ID, error boundary, route dispatch. The request-ID source is injected and deterministic. The error boundary reports once through an injected boundary and never writes a second response after commitment.
13. Treat the request-ID response header emitted by the Gin adapter as an explicitly additive, adapter-specific header. It is not part of the Project 047 envelope. The shared parity comparison either ignores only that documented additive header or wraps the standard-library handler with an equivalent test middleware that adds the same header. The parity baseline does not silently change the Project 047 JSON envelope or any other Project 047 header.
14. Set Gin test mode exactly once in non-parallel package-level setup before any engine or test is created. Do not toggle the mode from inside any parallel test. The parity baseline runs under the test-mode effect on logging without racing the global state.
15. Verify the router through `httptest`. Tests bind no port and depend on no network. The same contract table runs against the Project 047 standard-library handler and the Gin handler where practical.
16. The optional additions must not change the parity baseline. They are documented separately and are not part of the required contract.

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

- The strict JSON contract is enforced inside the Gin adapter.
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
- The Gin adapter does not introduce additional shared state.

- The middleware chain is composed once during setup and frozen before serving.
- The request-ID source is deterministic in tests.
- The request-ID middleware emits its response header on Gin responses; that header is adapter-specific additive and is not part of the Project 047 envelope.
- The error boundary reports a single error through an injected boundary on failure paths and never writes a second response after commitment.

- `HEAD` returns the corresponding known-path `405` policy with no body.
- Unknown paths return `404` with the documented envelope.
- The `Allow` header is sorted for `405` and matches Project 047.

- A failed validation, an unsupported media type, a malformed JSON, or an over-limit body does not allocate an ID, read the creation clock, or mutate any existing note.
- An unknown update does not read a replacement timestamp.
- An unknown delete does not alter the ID sequence or any note.

## 11. Project Constraints

- Use only the Go standard library plus exactly one pinned external dependency: the Gin framework.
- Pin the dependency to a specific version in the learner module configuration.
- Do not add another router, another binder, another JSON library, another web framework, an ORM, a database, or a logging library beyond what Gin provides.

- Do not use Gin global state in parallel tests.
- Do not use `gin.Default` in production code that needs a frozen middleware chain.
- Do not assume the framework's default binding is strict.
- Do not replace the Project 047 contract with framework-specific shortcuts.

- Do not include implementation code, function signatures, exact byte sequences of error messages, or exact field names of the JSON envelope in this guide.
- The guide states policies, not literal values.

## 12. Design Questions Before Coding

- Where does the boundary between the Gin adapter and the framework-independent domain live, and what types cross it?
- How is the JSON binder configured for strictness, and what explicit options are required?
- How is the body cap of exactly 1,048,576 bytes enforced before binding, and how is an over-limit body mapped to the documented `413`?
- How is the Project 047 optional `body` field preserved across the Gin adapter so that omission succeeds and stores an empty body?
- How is the route matrix reproduced exactly, including the `Allow` header values and the item-shaped path?
- How is the request-ID middleware composed with the error-boundary middleware, and how is the ID source injected?
- How is Gin test mode set exactly once in non-parallel package-level setup so that parallel tests do not race the global mode, and how is the additive request-ID response header reconciled with the Project 047 parity baseline?
- How is the contract table constructed so the same input runs against both the Project 047 standard-library handler and the Gin handler, and how is the additive request-ID header normalized in the comparison?
- How is the store injected into the Gin handler without coupling it to the framework?
- How is the optional addition kept out of the parity baseline so contract tests do not regress?
- How is the dependency version pinned in the module manifest so a moving tag does not change behavior?

## 13. Implementation Milestones

1. Record the Project 047 contract as a table of routes, methods, statuses, headers, content types, body policies, validation rules, ID semantics, timestamp semantics, error envelope, and concurrency semantics. Note the explicit treatment of the additive request-ID response header.
2. Pin the Gin dependency to a specific version in the learner module configuration.
3. Build the framework-independent store and domain logic with the same interface used in Project 047.
4. Configure the Gin binder for strict JSON parsing and bind the request body cap to the exact Project 047 value of 1,048,576 bytes.
5. Map the collection routes, the item routes, and the method dispatch onto the Gin engine without changing the contract. Preserve the Project 047 optional `body` semantics.
6. Compose the request-ID middleware, the error-boundary middleware, and the route dispatch middleware in the documented order. Document the additive request-ID response header.
7. Map the response writing for `201 Created`, `200 OK`, `204 No Content`, `404`, `405`, `400`, `413`, `415`, and the documented service errors.
8. Set Gin test mode exactly once in non-parallel package-level setup before any engine or test is created. Do not toggle the mode from inside any parallel test.
9. Build the contract table and run it against both the Project 047 standard-library handler and the Gin handler. Reconcile the additive request-ID response header using either header-ignore for that one documented header or a wrapping test middleware on the standard-library handler.
10. Add Gin-specific malformed-binding and middleware tests that exercise the strictness and the chain, including the additive request-ID header.
11. Finish deterministic, boundary, concurrency, and race-detector verification with the race detector.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Run the contract table against both the Project 047 standard-library handler and the Gin handler with the same inputs. Verify identical Project 047 status, content type, `Location`, `Allow`, body shape, body bytes, ID, timestamp format, error code, and error message for every case. Reconcile the additive request-ID response header by either ignoring only that documented additive header or wrapping the standard-library handler with the same additive header.
- Submit a JSON body with `title` only and no `body` for create. Verify `201` with empty `body` stored. Submit the same body shape for replacement and verify the stored body becomes empty and `created_at` is preserved while `updated_at` advances. Verify the omission of optional `body` is not treated as `400`.
- Submit a JSON body with an unknown field, a trailing second JSON value, a wrong type, a missing `title`, an empty `title`, a non-string `title`, an absent entire request document, and a non-object value. Verify the documented `400` behavior from the Gin adapter.
- Submit a body of exactly 1,048,576 bytes that is otherwise valid and verify it is accepted. Submit one additional byte and verify `413` from the Gin adapter before any store mutation.
- Submit a missing media type and an unrelated media type. Verify `415` from the Gin adapter.
- Submit an invalid ID and a missing item. Verify `400` and `404` from the Gin adapter.
- Run concurrent creates, reads, updates, and deletes against the Gin adapter under the race detector. Verify unique IDs, sorted list output, copies for read operations, and that the clock is read once per successful create and once per successful replacement inside the store's mutation critical section.
- Inspect the dependency manifest and verify the Gin version is pinned to a specific value, not to `latest`.
- Verify Gin test mode was set exactly once in non-parallel package-level setup before any engine or test was created. Verify that no parallel test toggles the mode, and that fresh engines alone do not isolate the global mode.
- Verify the request-ID middleware produces a deterministic ID from the injected source, that the same ID is observable from the response and from a test-injected boundary, that the error-boundary middleware reports failures once and never writes a second response after commitment, and that the request-ID response header is documented as adapter-additive rather than as part of the Project 047 envelope.
- Verify optional additions are documented separately and do not change the parity baseline.
- Verify every test uses `httptest`, no test binds a fixed port, and no test contacts external services.

## 15. Common Mistakes to Watch For

- Assuming the framework's default JSON binding is strict accepts unknown fields and trailing second values.
- Assuming the framework applies a Project-047-equivalent body cap silently accepts oversized requests.
- Inferring `Content-Type` from body bytes changes the contract and hides 415 vs 400 distinctions.

- Treating an absent optional `body` as `400` rejects requests that Project 047 accepts.
- Reading the body without the exact 1,048,576-byte cap permits unbounded growth.
- Inferring the body cap from a framework default lets oversized requests through.

- Using a Gin global engine in production with default middleware makes the middleware chain non-deterministic and violates the parity baseline.
- Toggling Gin test mode from inside parallel tests races the global state and produces flaky failures.
- Believing that constructing a fresh engine per test isolates Gin's global mode is wrong; the global mode applies regardless of how many engines exist.

- Reimplementing the domain inside the Gin handler couples the contract to the framework and prevents the contract table from running against both handlers.
- Returning pointers or slices backed by store memory breaks the copy contract.
- Incrementing IDs before validation creates gaps on failed requests.
- Reading the clock outside the store's mutation critical section changes the timestamp semantics and breaks Project 047 parity.

- Composing middleware in a different order changes the request ID visibility and the error boundary coverage.
- Letting the framework's recovery middleware run inside the contract changes the panic-to-response behavior.
- Using a panic recovery that writes a second response after commitment hides the original error.
- Treating the additive request-ID response header as part of the Project 047 envelope breaks the parity table without making the additive header itself wrong.

- Returning `200` for creation, a body for deletion, no `Location` header on `201`, or implicit `HEAD` from `GET` violates the parity baseline.
- Returning different `Allow` values, different error codes, or different status codes for the same input breaks the contract table.

- Skipping the dependency pin and resolving Gin to `latest` introduces non-deterministic versions into tests.
- Using a non-pinned module path that points at a moving tag changes binding behavior between runs.

## 16. Topics and References for Study

- Study the official Gin documentation, especially `gin.Engine`, `gin.RouterGroup`, `gin.Context`, `ShouldBindJSON`, `BindJSON`, the `validator` package integration, and the documented strictness options.
- Study Gin's middleware signature and the documented behavior of `gin.Recovery`, `c.AbortWithStatus`, and `c.JSON`.
- Study the documented Gin test mode and its global scope.
- Study `encoding/json` decoder options for strict decoding.
- Study the official Project 047 README and contract to understand the exact parity target, including the optional `body` field and the body cap.
- Review the Go modules documentation for version pinning and reproducible builds.

## 17. Self-Assessment Questions

1. Why must the Project 047 contract, including the optional `body` semantics, be reproduced exactly on Gin and verified by the same contract table against both handlers?
2. Why is the domain kept framework-independent, and what does the boundary look like?
3. Why must strict JSON and the exact 1,048,576-byte body cap be enforced explicitly before binding rather than assumed from framework defaults?
4. Why is the clock read inside the store's mutation critical section, and why is at most one clock reading taken per successful create or replace?
5. Why is the middleware chain frozen at setup and not mutated during serving?
6. Why must Gin test mode be set exactly once in non-parallel package-level setup and not toggled from parallel tests, and why doesn't a fresh engine isolate the global mode?
7. Why is the additive request-ID response header documented as adapter-specific instead of being claimed as part of the Project 047 envelope, and how is it reconciled in the parity comparison?
8. Why is the Gin dependency pinned to a specific version rather than to `latest`?
9. Why must optional additions not change the parity baseline?
10. Why is the race detector still required even when the framework provides its own concurrency primitives?

## 18. Definition of Completion

- [ ] The Gin adapter reproduces the Project 047 contract for routes, methods, statuses, headers, content types, validation, body cap of exactly 1,048,576 bytes, ID semantics, timestamps, error envelope, and concurrency.
- [ ] The optional `body` field preserves Project 047 semantics: omitted `body` is accepted on create and replacement and stores an empty body; absent entire document, missing `title`, wrong types, unknown fields, and trailing second values are `400`.
- [ ] The Gin dependency is pinned to a specific version in the learner module manifest.
- [ ] JSON binding is configured for strict decoding and rejects unknown fields, trailing second values, and wrong types with the documented `400` policy.
- [ ] The body cap is enforced before binding and maps oversized requests to `413`.
- [ ] The request-ID and error-boundary middleware are composed in the documented order with injected sources and boundaries; the request-ID response header is documented as additive and reconciled in the parity comparison.
- [ ] The clock is read inside the store's mutation critical section: at most once per successful creation and at most once per successful replacement.
- [ ] Gin test mode is set exactly once in non-parallel package-level setup before any engine or test is created, and no parallel test toggles the mode.
- [ ] The contract table runs against both the Project 047 standard-library handler and the Gin handler with identical results for every Project 047-required status, header, and body.
- [ ] Concurrent requests pass the race detector with the same store and clock as Project 047.
- [ ] The domain and store are framework-independent and remain usable by other adapters in later projects.
- [ ] The implementation uses only the standard library plus the pinned Gin dependency, and the learner can explain each parity decision.

## 19. Optional Extensions

- Add a request-ID echo header on every response using the documented middleware composition, without changing any other contract item, and document it as additive.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 053 — URL Shortener API](../../04-apis-and-services/053_url_shortener_api/README.md#20-prerequisite-based-documentation-guide), [Project 046 — Basic HTTP Server](../../04-apis-and-services/046_basic_http_server/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`github.com/gin-gonic/gin`](https://pkg.go.dev/github.com/gin-gonic/gin).
- **Standards and concept references:** [Gin documentation](https://gin-gonic.com/en/docs/), [Go REST API with Gin tutorial](https://go.dev/doc/tutorial/web-service-gin).

### Project-specific learning focus

- **Learn now:** framework routing and context lifetimes, strict binding, middleware and recovery, validation, global test mode, version pinning, and contract parity.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
