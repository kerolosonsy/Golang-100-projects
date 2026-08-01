# Project 049 — JSON API Response Formatter

## 1. Project Name and Number

Project 049 — JSON API Response Formatter, located in `049_json_api_response_formatter`.

## 2. Project Idea

Build a reusable response boundary for HTTP JSON APIs. It emits one stable success envelope or one stable public error envelope, maps known domain-error categories to deliberate HTTP statuses, hides unknown internal causes, and commits a response only after its complete JSON representation is available in memory.

## 3. Why This Project Now?

Project 048 introduced a router and middleware chain that can identify a request and select an endpoint. This project gives those endpoints one response owner, preventing every handler from inventing a different JSON shape or writing a second response after an error. Project 047 supplies concrete REST status cases, and Project 046 supplies handler and response-writer fundamentals. Project 050 will rely on this boundary for authentication failures without exposing token or credential details.

## 4. Prerequisites

Complete Projects 048 and 046 before starting. Earlier projects may be useful review, but they are not required prerequisites.

You should already understand response commitment, `httptest`, JSON decoding and encoding, request IDs in context or headers, typed error classification, and why a handler must have one clear response owner.

## 5. What You Must Know Before Starting

- HTTP status and headers are committed before response bytes are sent. A later encoding failure cannot repair a response that has already been committed.
- JSON can fail during encoding even after a value has been accepted by application code. Unsupported values, invalid numeric values, and nested unsupported data must be treated as response-boundary failures.
- A public error is a deliberate contract, not the text returned by an internal error. Database paths, stack traces, parser details, secrets, and wrapped causes do not belong in a client envelope.
- Error classification can survive wrapping when the application uses typed or sentinel domain categories consistently. Unknown causes must take the generic internal path.
- A response writer permits one owner to select status, headers, and body. Multiple helpers writing independently create double responses and misleading logs.
- Buffering a complete representation before commitment improves failure handling but consumes memory proportional to the complete response. It is not streaming and is not an unlimited-response defense.
- A `204 No Content` response has no representation body and therefore must not receive a JSON envelope. A `HEAD` response is bodyless according to its selected handler policy.
- A `401 Unauthorized` response needs an authentication challenge. This formatter's unauthenticated mapping uses the exact `WWW-Authenticate` value documented below.

## 6. Explanation of New Concepts

An envelope is a stable outer JSON object that gives clients one predictable place to find application data or a public error. The success envelope contains a required `data` field and may contain a request identifier. The error envelope contains a required nested `error` object and may contain the same request identifier. The status remains in the HTTP response rather than being duplicated as an unstable body field.

The success data value is intentionally generic at this boundary. It may be a JSON object, array, string, number, boolean, or null when the selected endpoint contract permits it. The formatter is responsible for serializing it, not for deciding whether its domain meaning is valid.

A public error category is a controlled classification such as validation, authentication, not-found, or conflict. The category determines a fixed status, public code, and safe message. An internal cause may be attached to that category for logs, but it never changes the public message unless the contract explicitly defines a safe value.

The response boundary first prepares a complete envelope in a temporary memory buffer. Only a successful encoding of that complete buffer can proceed to header and status commitment. If success data cannot be encoded, the buffer is discarded and a generic internal-error envelope is prepared instead. This ordering makes an encoding failure observable as a clean `500` when no bytes have been written.

Commitment is one-way. After headers and status are selected and the body write begins, a write error can be recorded but cannot be turned into a second JSON response. The boundary therefore sets headers once, selects status once, and performs one body write for body-bearing responses.

An application boundary is the trusted place to log an internal cause. It includes the request ID when one exists, uses a structured logger or injected output, and records a safe event rather than copying internal text into the client message. The logger must not be used as a channel for passwords, tokens, or other secrets.

## 7. Learning Objective

By completion, you can define stable JSON response contracts, classify domain errors without leaking implementation detail, buffer before commitment, and enforce one response owner. You can explain the trade-off between clean encoding failure handling and memory use, implement bodyless `204` and `HEAD` behavior correctly, and verify headers, status, writes, logs, and public non-leakage deterministically.

## 8. Functional Requirements

1. Define one success envelope with exactly one required top-level field named `data` and one optional top-level field named `request_id`.
2. Define one public error envelope with exactly one required top-level field named `error` and one optional top-level field named `request_id`. The `error` object has exactly `code` and `message`, both stable strings.
3. Omit `request_id` when no non-empty request ID is available. When present, include the same value in the envelope and echo it in the `X-Request-ID` response header.
4. Use content type exactly `application/json; charset=utf-8` for every body-bearing success and error response. Do not rely on automatic content-type sniffing.
5. Support the following public error mappings with fixed status, code, and message: validation maps to `400` / `invalid_request` / `invalid request`; unauthenticated maps to `401` / `unauthenticated` / `authentication required`; forbidden maps to `403` / `forbidden` / `forbidden`; not-found maps to `404` / `not_found` / `resource not found`; conflict maps to `409` / `conflict` / `resource conflict`; payload-too-large maps to `413` / `payload_too_large` / `payload too large`; unsupported-media-type maps to `415` / `unsupported_media_type` / `unsupported media type`; rate-limited maps to `429` / `rate_limited` / `too many requests`; and unavailable maps to `503` / `service_unavailable` / `service unavailable`.
6. The unauthenticated mapping always includes `WWW-Authenticate` with exact value `Bearer`.
7. An unknown internal error, including an unexpected dependency failure or an encoding failure, maps to `500` / `internal_error` / `internal server error`. No internal error string, stack trace, SQL path, credential, token, or secret appears in the client envelope.
8. Classify wrapped known errors by their documented typed or domain category. A cause that does not match a known category uses the generic `500` mapping.
9. Serialize the complete selected envelope into a temporary memory buffer before setting response headers, status, or body. If success-data encoding fails before commitment, discard it and emit the generic `500` envelope instead.
10. Set headers once, select status once, and write one complete body once for every body-bearing response. Do not call `http.Error` after a JSON response attempt and do not append a second error document.
11. A `204 No Content` response has no envelope, no body bytes, and no JSON content type. It may still carry a request-ID response header when one is available.
12. `HEAD` behavior follows the endpoint's explicit handler policy. The formatter does not infer `HEAD` from `GET`; an accepted `HEAD` has the selected status and representation headers but zero body bytes, while a rejected `HEAD` follows the endpoint's error policy without writing a body.
13. Prevent double-write ownership errors by arranging for one response boundary to own the writer, either because handlers return results or errors to it or because an equivalent ownership guard is enforced. Do not require a particular function signature.
14. Log the internal cause of unknown failures and pre-commit encoding failures at the application boundary with the request ID when available. The log is injected or otherwise test-observable and never substitutes internal text for the public message.
15. The formatter's behavior is safe for concurrent independent requests and does not use package-level mutable response state.

## 9. Inputs and Outputs

A success input consists of one of the supported HTTP success statuses `200 OK`, `201 Created`, or `202 Accepted`, a JSON-encodable data value, and an optional request ID. Another body-bearing success status is invalid formatter input and takes the generic pre-commit `500` path; `204` is handled separately and is always bodyless.

An error input consists of a typed or domain error category, an internal cause when one exists, and an optional request ID. The category selects the exact public mapping. A nil or otherwise absent error is not an error response and must not be silently converted into an arbitrary public failure.

Text-only success example: a successful request with data describing one note produces status `200`, content type `application/json; charset=utf-8`, and a JSON object whose only required top-level field is `data`. If a request ID is available, `request_id` is the additional top-level field and the response header carries the same value.

Text-only error example: a classified not-found cause produces status `404`, the JSON content type, and an error envelope whose nested code is `not_found` and whose nested message is exactly `resource not found`. The internal lookup detail is absent.

Text-only unknown-error example: a dependency reports a cause containing a private file path. The client receives status `500`, code `internal_error`, and message `internal server error`; only a trusted application log may record a sanitized internal cause, and sensitive material remains redacted.

A serialized body is one complete JSON document followed by one line-feed character. Field presence and names are stable; success fields are emitted in the order `data` then `request_id`, and error fields in the order `error` then `request_id`, with nested error fields `code` then `message`. Clients must parse JSON rather than depend on object order, but deterministic order makes byte-level tests and diagnostics stable.

## 10. Rules and Edge Cases

A success data value of null is valid if the endpoint's domain contract allows it. An unsupported value anywhere inside the data, an invalid floating-point value, or any other JSON encoding failure is an internal response failure. No prefix of the failed representation reaches the client.

If the generic error envelope itself cannot be encoded because of application-supplied data, the implementation has violated this contract: public generic fields are fixed safe strings and must remain encodable. The boundary must not fall back to plain-text `http.Error` after attempting JSON.

A request ID is absent when it is empty. Whitespace-only values are not silently normalized into an ID. A non-empty value is treated as opaque response data only after its upstream request-ID policy has validated it; the formatter does not invent or sanitize IDs.

Mapped public messages are constants. Wrapping a known domain error with extra internal text does not alter its public code or message. An unknown error never becomes a public `400` merely because its text resembles a validation failure.

For `401`, the `Bearer` challenge is part of the status contract. Other security response headers are outside this formatter unless the endpoint's explicit boundary owns them before commitment. The formatter must not overwrite unrelated headers after they have been committed.

A `204` has no JSON content type and no body even when a caller supplies data. The formatter ignores that data for the bodyless response and does not attempt to encode it. A `HEAD` response is tested for zero bytes independently of the representation that would have been sent for `GET`.

A response writer failure after the one body write is not recoverable as another HTTP response. It is reported through the chosen application boundary and never triggers a second status, header set, body write, or `http.Error` call.

Response buffering covers the complete envelope for one response. It does not provide streaming, a general memory quota, or protection against an arbitrarily large JSON value. If an endpoint can produce large data, that endpoint must impose its own bounded contract or choose a different design outside this project's required scope.

## 11. Project Constraints

Use only the Go standard library. Appropriate standard facilities include `encoding/json`, `bytes`, `errors`, `fmt` only for controlled internal context, `net/http`, `net/http/httptest`, and `log/slog` or another injected standard logging destination. Do not use a web framework, third-party envelope library, `http.Error` for JSON responses, raw internal error strings, stack traces, SQL paths, or secrets in client output.

The required formatter is not a streaming writer and does not promise unlimited response safety. Do not prescribe function signatures, generated code, or a particular domain-error type layout in the guide. One response boundary must own the writer, and tests must be able to observe commitment order and write count without network access or sleeps.

## 12. Design Questions Before Coding

- What exact JSON field presence rules distinguish a success envelope from an error envelope?
- How will optional request IDs be represented when absent, and how will the response header remain consistent with the body?
- Which domain categories are stable enough to expose publicly, and which causes must remain generic?
- How will wrapped errors be classified without comparing unstable error strings?
- Where does the one response owner live, and how is a handler prevented from writing before or after it?
- How will a complete envelope be buffered without accidentally writing headers through an encoder attached to the real writer?
- What status and header policy applies when data is supplied with `204` or when a handler permits `HEAD`?
- How will a custom test writer prove that headers and status precede the one body write?
- Which internal cause is logged for an encoding failure, and how will logging avoid copying credentials or tokens?
- What response-size assumption is safe to document, and what must explicitly remain outside the formatter's promise?

## 13. Implementation Milestones

1. Write the exact success and error envelope contracts, field order, optional request-ID behavior, and public message table.
2. Define typed/domain error categories and their fixed status and public mappings, including the generic unknown path.
3. Establish one response ownership boundary and a test-observable internal logging boundary.
4. Add complete pre-commit success serialization and verify valid data shapes.
5. Add mapped error serialization and the generic non-leaking `500` path.
6. Add request-ID header and envelope behavior and the exact unauthenticated challenge.
7. Add bodyless `204` and endpoint-controlled `HEAD` behavior.
8. Add response commitment instrumentation, one-write enforcement, and post-commit write-failure reporting.
9. Add deterministic tests for encoding failure, status/header ordering, public non-leakage, and structured logs.
10. Review memory-scope documentation and run concurrent independent-request tests under the race detector.

## 14. Verification Cases the Learner Must Write

- Verify successful `200`, `201`, and `202` responses have the exact success envelope, stable field presence, content type, selected status, and one body write.
- Verify data values representing an object, array, string, number, boolean, and permitted null are encoded in the `data` field without an extra success field.
- Verify a non-empty request ID appears in the envelope and response header, while an absent ID is omitted from both places as documented.
- Verify every mapped error category: exact status, public code, public message, JSON content type, and any required `WWW-Authenticate` value for `401`.
- Verify wrapped known errors retain their mapping and unknown errors use generic `500`.
- Give the success path a JSON-unsupported value or invalid numeric value. Verify no status, headers, or body bytes were committed before the clean generic `500` response was selected.
- Verify the unknown client response does not contain an internal error string, stack trace marker, private file path, credential, token, or secret.
- Capture the structured log for an unknown error and verify the internal cause is recorded with the request ID while the client sees only the public generic message.
- Verify `204` has the selected status, zero body bytes, no JSON content type, and only the allowed request-ID header.
- Exercise an accepted `HEAD` policy and verify status and representation headers follow the selected response while body bytes remain zero. Exercise a rejected policy separately.
- Use a response recorder and an instrumented writer to verify headers and status are set before the body, exactly one body write occurs, and `http.Error` is never used after JSON handling.
- Force a body-writer error after commitment and verify it is reported once without a second write or status change.
- Attempt a second response from a handler or nested helper and verify the single-owner design prevents double-write behavior.
- Run independent concurrent success and error requests under the race detector and verify request IDs, logs, and envelopes do not cross-contaminate.
- Verify no test needs a sleep, fixed port, external service, or real network connection.

## 15. Common Mistakes to Watch For

Writing the status before discovering that JSON encoding fails leaves clients with a misleading success response. Encoding directly to the real writer can leak a partial document. Calling `http.Error` after a JSON write produces mixed content and may append bytes after headers are committed. Letting each handler write its own error creates competing response owners.

Returning `err.Error()` as the public message leaks implementation details and makes the API unstable. Mapping errors by string comparison breaks when causes are wrapped or wording changes. Logging a token, password, SQL path, or secret while trying to diagnose an unknown error defeats the boundary even if the client message is generic.

Adding a JSON envelope to `204` violates its bodyless semantics. Treating `HEAD` as an unconditional `GET` can write or record a body that must not be sent. Forgetting the authentication challenge on a `401` makes the response incomplete for clients.

Assuming buffering is a response-size limit hides memory risk. Using a real clock or sleep to test encoding and commitment order makes failures timing-dependent. A recorder that only checks final bytes may miss a header mutation or a second write, so commitment-order tests need an instrumented boundary.

Using a global request ID, logger, or response buffer allows concurrent requests to share state. Reusing a mutable envelope between requests can cause data races and cross-request leakage. Treating a successful race-detector run as proof that the public contract is correct misses status and non-leakage failures.

## 16. Topics and References for Study

Study the official documentation for `encoding/json`, `bytes.Buffer`, `net/http.ResponseWriter`, `net/http/httptest`, `errors.Is` and `errors.As`, and the standard structured logging package `log/slog`. Read HTTP semantics for response commitment, `204 No Content`, `HEAD`, `401` challenges, and content types. Search for JSON envelope design, typed domain errors, error classification across wrapping, response-writer instrumentation, buffering trade-offs, and safe public error messages. Review the route and request-ID boundaries from Project 048.

## 17. Self-Assessment Questions

1. Why must complete success serialization happen before headers and status are committed?
2. What information belongs in a public error envelope, and what information must remain internal?
3. Why is typed error classification more stable than comparing error strings?
4. Why does an unknown error become generic `500` rather than a guessed client error?
5. What makes `204` different from an ordinary successful response envelope?
6. How can an accepted `HEAD` preserve status and headers while writing zero body bytes?
7. Why is one response owner necessary even when each helper appears to handle a different failure?
8. What does a temporary buffer protect against, and what memory risk does it not solve?
9. Why should a `401` include `WWW-Authenticate` in this contract?
10. How can tests prove header-before-body ordering and a single write rather than merely inspecting final text?

## 18. Definition of Completion

- Success and error envelopes have exactly the documented fields, stable presence rules, public messages, and optional request-ID behavior.
- Every mapped category and unknown failure produces the pinned status, code, message, content type, and challenge behavior.
- Unknown causes, encoding failures, stack traces, private paths, credentials, tokens, and secrets never appear in client output.
- Complete envelopes are buffered before commitment; a pre-commit encoding failure becomes one clean generic `500` response.
- Headers and status are set once and body-bearing responses write once. No JSON response is followed by `http.Error` or a second response.
- `204` and endpoint-controlled `HEAD` behavior are bodyless and tested.
- Internal causes are logged at the application boundary with request ID without making logs a client-side leak.
- Tests prove commitment order, write count, mapping, non-leakage, logging, and concurrent safety using deterministic in-memory boundaries.
- The documented buffering scope is honest, and the implementation uses only standard-library facilities.

## 19. Optional Extensions

1. Add an explicit maximum serialized-response size that fails before commitment, with tests distinguishing it from the formatter's existing memory-buffer behavior.
2. Add content negotiation for one additional JSON-compatible media type while preserving the same envelope and error contracts.
