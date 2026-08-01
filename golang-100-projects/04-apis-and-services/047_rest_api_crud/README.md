# Project 047 — REST API CRUD

## 1. Project Name and Number

Project 047 — REST API CRUD, located in `047_rest_api_crud`.

## 2. Project Idea

Build an in-memory JSON REST API for Notes. A note has a stable positive identifier, a required title, an optional body, and creation and update timestamps. The service exposes collection operations and item operations with a fixed HTTP contract, strict JSON input, deterministic time injection, and a store safe for concurrent requests.

## 3. Why This Project Now?

Project 046 introduced independently testable `net/http` handlers and precise status and header behavior. This project applies that foundation to resource-oriented routing, JSON validation, HTTP method semantics, and synchronized mutable state. Project 046 remains a required foundation, while Project 014 supplies validation practice and Project 017 is useful review for JSON thinking; neither replaces the exact prerequisites below. Project 048 will later build a router and middleware layer on top of this API discipline.

## 4. Prerequisites

Complete Project 046 before starting. Earlier projects, including Projects 014 and 017, may be useful review, but they are not required prerequisites.

You should already be able to test a handler with `httptest`, set response headers before a body, distinguish `404` from `405`, and keep network startup outside handler tests.

## 5. What You Must Know Before Starting

- REST resource paths identify a collection or one item; the HTTP method identifies the operation on that resource.
- JSON decoding is not automatically strict. A decoder may accept unknown fields or leave a second value unread unless the boundary checks for them.
- A request media type describes the representation being sent. This API accepts JSON requests only and returns JSON for every response that has a body.
- `MaxBytesReader`-style request limiting protects the service from an unbounded body, but the limit must be applied before decoding and its failure must map to a deliberate status.
- A mutex protects the store's shared state, but it does not by itself make returned references safe. Results exposed to callers must be copies.
- A monotonic identifier is never moved backward or reused after a successful allocation, including after deletion.
- An injected clock makes timestamp behavior deterministic and prevents tests from depending on wall-clock scheduling.
- `201 Created`, `204 No Content`, `400 Bad Request`, `404 Not Found`, `405 Method Not Allowed`, `413 Payload Too Large`, and `415 Unsupported Media Type` have distinct meanings that must not be substituted casually.

## 6. Explanation of New Concepts

A collection endpoint represents all notes at `/notes`. Its `GET` operation returns an ordered snapshot, and its `POST` operation creates one note. An item endpoint represents one note at `/notes/{id}`. Its `GET` reads, its `PUT` replaces, and its `DELETE` removes. The path and method are evaluated before the request document is interpreted.

Full replacement with `PUT` means the submitted representation becomes the complete mutable note representation. This project does not define a partial-update operation. A missing optional body therefore means an empty body for the replacement, while a missing required title is invalid.

Strict JSON input has several independent checks: the media type is appropriate, the body is within the byte limit, the first JSON value has the expected shape and field types, no unknown field appears, and no second non-whitespace value follows it. Passing one check does not imply that the others passed.

The store owns note state. Its mutex protects ID allocation and maps or collections, and its public read operations return copies. List order is a contract rather than an accidental property of map iteration.

A clock dependency is a boundary around time acquisition. The API still emits timestamp strings in one documented format, but tests can supply a fixed sequence of times and assert which state transition used which reading.

## 7. Learning Objective

By completion, you can explain and implement a complete in-memory REST contract, validate JSON at an HTTP boundary, preserve resource identity during replacement, synchronize concurrent mutation, and produce deterministic output independent of map iteration and wall-clock timing. You can also explain why copies, injected time, request-size limits, and strict end-of-document checks matter.

## 8. Functional Requirements

1. Store Notes in memory. Each Note has exactly these public representation fields: positive integer `id`, string `title`, string `body`, string `created_at`, and string `updated_at`.
2. Allocate IDs beginning at `1`. Every successfully created note receives a larger ID than every earlier successful creation. A deleted ID is never reused.
3. Accept `POST /notes` for creation and `GET /notes` for collection listing.
4. Accept `GET /notes/{id}` for one item, `PUT /notes/{id}` for full replacement, and `DELETE /notes/{id}` for deletion.
5. A create request must contain a non-empty title after surrounding whitespace is trimmed. Its body is optional and defaults to the empty string. Store the trimmed title and preserve the body value exactly.
6. A replacement request must contain the same title and optional body rules as creation. It preserves the item's ID and `created_at`, replaces title and body, and assigns a new `updated_at` from the injected clock.
7. A successful create returns `201 Created`, a JSON note representation, and `Location` set exactly to the path `/notes/{id}` for the new item.
8. A successful collection list returns `200 OK` and a JSON array sorted by ascending ID. An empty collection is represented as an empty JSON array, not a JSON null value.
9. A successful item read and replacement return `200 OK` and one JSON note representation.
10. A successful delete returns `204 No Content` with zero body bytes, no `Content-Type` header, and no JSON envelope.
11. Missing items return `404 Not Found` and a consistent JSON error document. A missing update or delete leaves all store state unchanged.
12. A known path with an unsupported method returns `405 Method Not Allowed`, a consistent JSON error document, and a sorted `Allow` header. The collection allows `GET, POST`; an item allows `DELETE, GET, PUT`.
13. Request bodies for create and replacement require an `application/json` media type, allow valid media-type parameters, and reject other or missing media types with `415 Unsupported Media Type`.
14. Request bodies are capped at exactly 1,048,576 bytes. A request that exceeds the cap returns `413 Payload Too Large` without creating or changing a note.
15. Malformed JSON, a non-object root, wrong field types, an empty title, unknown fields, and a trailing second JSON value return `400 Bad Request` with the documented JSON error shape.
16. Invalid item IDs return `400 Bad Request`. A valid ID is a canonical, positive, base-ten decimal path segment with no sign, no leading zero, and no numeric overflow.
17. All non-`204` responses have content type exactly `application/json; charset=utf-8`, including errors. `HEAD` is not implied by `GET`: it returns the corresponding known-path `405` policy with no body, while an unknown path remains `404` with no body.
18. Store operations are safe for concurrent HTTP requests. List results and individual reads are copies that cannot mutate store-owned state.
19. Inject the clock used for `created_at` and `updated_at`. A successful create reads it once and uses that reading for both timestamps; a successful replacement reads it once for `updated_at`. The store invokes the clock inside the same mutation critical section that protects ID allocation and note updates, so concurrent mutations do not race through a non-thread-safe test clock.

## 9. Inputs and Outputs

The collection path is exactly `/notes`; the item path is exactly `/notes/{id}`. The request document for create and replacement is a JSON object with required string field `title` and optional string field `body`, and no other fields. Title surrounding whitespace is removed before validation and storage. Body whitespace is not changed.

Note output fields use JSON names `id`, `title`, `body`, `created_at`, and `updated_at`. `id` is a positive JSON integer. The two timestamp strings use UTC and RFC3339Nano formatting with a literal `Z` suffix. The created note has equal creation and update timestamps because both use the one creation reading.

Text-only create example: submitting a JSON note with title `  groceries  ` and no body creates a note whose stored title is `groceries`, whose body is empty, and whose first successful ID is `1`. The response status is `201`, and its `Location` value is `/notes/1`.

Text-only list example: after successful creations with IDs `3` and `1`, the collection response lists ID `1` before ID `3`, regardless of request completion order. An empty list response is the JSON array with no elements.

Text-only error examples: a missing title is a `400` with code `invalid_request`; a missing item is a `404` with code `not_found`; a missing or non-JSON media type is a `415` with code `unsupported_media_type`; and an over-limit body is a `413` with code `payload_too_large`.

Every JSON error document has exactly two top-level string fields, `code` and `message`. Codes and messages are public stable values, not raw decoder, filesystem, mutex, or clock errors. The required mappings are `invalid_request` with `invalid request`, `invalid_id` with `invalid note id`, `not_found` with `note not found`, `method_not_allowed` with `method not allowed`, `unsupported_media_type` with `unsupported media type`, `payload_too_large` with `payload too large`, and `internal_error` with `internal server error`.

## 10. Rules and Edge Cases

The title is invalid when it is absent, not a JSON string, or empty after surrounding whitespace is trimmed. A title consisting only of whitespace is invalid. A title containing internal whitespace, Unicode, or punctuation is allowed unless the body limit is exceeded. The body may be absent or empty and is otherwise preserved as submitted.

The decoder accepts one complete JSON document followed only by JSON whitespace. An empty body, malformed value, second JSON value, unknown object field, or wrong type is rejected under the documented `400` policy. `null` is not a note object. Duplicate-known-field detection is outside the required scope. A body at the byte limit is eligible for decoding if it is otherwise valid; any attempt to exceed the limit is `413`, and no partial mutation is allowed.

IDs are parsed from the item path only. An item-shaped path with exactly one non-empty ID segment returns `400` when that segment is zero, negative, signed, leading-zero, non-decimal, overflowing, or otherwise non-canonical. An empty segment, trailing slash, or extra path segment does not identify an item route and returns `404`.

Failed validation, unsupported media type, malformed JSON, and an over-limit body do not allocate an ID, read the creation clock, or mutate an existing note. An unknown update does not read a replacement timestamp. An unknown delete does not alter the ID sequence or any note.

A replacement never changes ID or creation time. It performs a fresh update-time assignment even if the replacement title and body happen to equal the old values. Clock reads for successful creates and replacements occur while the store's mutation lock is held, alongside the state transition they timestamp. A clock supplied by tests may return equal or non-monotonic times; the API reports the injected values rather than inventing wall-clock ordering, and tests that require a changed serialized timestamp supply a later value.

If the representable ID space is exhausted, creation fails without wrapping or reusing an ID and returns the generic internal error contract. This condition is not simulated with a long-running test. Concurrent creates may complete in any order, but each receives one unique ID and every completed list is sorted.

## 11. Project Constraints

Use only the Go standard library, including `net/http`, `encoding/json`, `sync`, `time`, `mime`, and `net/http/httptest` as appropriate. Do not use a database, framework, global mutable store, generated API package, or third-party JSON validator.

Do not implement partial update semantics, pagination, filtering, persistence, authentication, or a custom router in the required project. Do not expose internal errors, map iteration order, mutable store references, or unbounded request bodies. Tests must use `httptest`, injected time, deterministic data, and the race detector; they must not bind fixed ports, sleep, or call external services.

## 12. Design Questions Before Coding

- What representation separates transport JSON from store-owned note state, if separation is useful?
- Where is ID allocation serialized, and how is non-reuse preserved after deletion and failed requests?
- How will a list snapshot remain sorted without relying on map iteration order?
- What exactly counts as one complete JSON document, and how will trailing whitespace differ from a second value?
- How will media-type parameters be recognized without accepting an unrelated type?
- At what point is the body limit applied, and how will an over-limit decode avoid partial mutation?
- Which timestamp is read for create and replacement, and how will a deterministic clock prove that choice?
- How will returned values be copied so a caller cannot mutate store-owned state?
- Which path and method are classified first when a request is both malformed and aimed at a missing item?
- How will response writing ensure headers and the body match the selected status, especially for `204` and `HEAD`?

## 13. Implementation Milestones

1. Write down the collection and item route matrix, including methods, statuses, `Allow` values, content types, and body policies.
2. Define the Note representation, validation rules, timestamp format, and public error contract.
3. Establish an injected-clock boundary and a store ownership boundary.
4. Implement successful creation with monotonic non-reused IDs and exact creation timestamps.
5. Implement deterministic collection listing and item retrieval using copies and ascending order.
6. Implement full replacement while preserving identity and creation time and changing update time.
7. Implement deletion and verify that missing mutations do not change state.
8. Add strict JSON, media-type, ID, and request-body-limit validation before any state mutation.
9. Add method dispatch, error responses, `Location`, `Allow`, `204`, and explicit `HEAD` behavior.
10. Add sequential, deterministic, concurrent, and race-detector verification for the complete API.

## 14. Verification Cases the Learner Must Write

- Create the first note and verify `201`, the exact JSON fields, trimmed title, empty default body, both timestamps, and exact path-only `Location`.
- Create several notes with an injected sequence of times and verify IDs are positive, increasing, non-reused, and timestamp values are exact.
- List an empty store and verify a JSON empty array. List notes created in a deliberately different completion order and verify ascending ID order.
- Read an existing item and verify all fields. Read a missing item and verify `404`, the exact JSON error fields, code, message, and content type.
- Replace an existing item with a new title and body and verify ID and `created_at` are unchanged while `updated_at` equals the next injected clock value.
- Replace with omitted body and verify the stored body becomes empty. Replace with an unchanged representation and verify the update timestamp still changes according to the injected clock.
- Delete an existing item and verify `204` has zero body bytes and that a later read returns `404`. Delete the same item again and verify no state or ID-sequence change.
- Attempt update and delete on an unknown ID and verify neither operation mutates the store or consumes a clock reading.
- Exercise every unsupported method on the collection and item paths. Verify `405`, exact sorted `Allow`, stable JSON error body, and content type.
- Exercise `HEAD` on known and unknown paths and verify the selected status and headers with zero body bytes.
- Send a missing media type, an unrelated media type, and a malformed media type and verify `415` without mutation. Send `application/json` with a valid media-type parameter and verify it passes the media-type check.
- Send malformed JSON, an empty body, a non-object, wrong field types, an unknown field, a second JSON value, and trailing whitespace after one valid value. Verify only the single valid document and its trailing-whitespace variant succeed. Duplicate-known-field detection is not a required verification case.
- Send empty and whitespace-only titles, titles with surrounding whitespace, Unicode titles, and bodies containing whitespace. Verify the exact trim and preservation rules.
- Send canonical, zero, negative, signed, leading-zero, overflow, and non-numeric IDs in an item-shaped path and verify `400`. Send an empty segment, trailing slash, or extra segment and verify `404`.
- Send a body exactly at the cap and another that exceeds it. Verify the larger request is `413` and neither request can cause partial state mutation.
- Run many concurrent creates against one store, verify unique IDs and the exact count, then list and verify sorted order. Run concurrent reads, updates, and deletes under the race detector.
- Verify two independent stores do not share IDs, timestamps, or note state.

## 15. Common Mistakes to Watch For

Using a map directly for list output makes order nondeterministic and creates flaky tests. Returning pointers or a slice backed by store memory lets callers mutate state outside the mutex. Incrementing an ID before validation can create unexplained gaps and makes failed requests affect the contract. Reusing a deleted ID breaks identity and can make old references point at a new resource.

Treating `PUT` as an unspecified partial update creates incompatible clients. Trimming the body as well as the title changes user data. Updating `created_at` during replacement destroys resource history. Calling the clock multiple times in one operation makes deterministic timestamp assertions ambiguous.

A JSON decoder that stops after its first value accepts concatenated documents. Default decoder behavior may accept unknown fields. Checking only `Content-Type` for a string prefix can accept unrelated media types. Decoding an unbounded request before applying a limit permits avoidable memory growth.

Returning `200` for creation, a body for deletion, or no `Location` header loses useful HTTP semantics. Falling through from an unsupported method to `404` hides a known resource. Allowing implicit `HEAD` without documenting it makes handler and wire behavior diverge. Writing raw internal parse or mutex errors gives clients unstable information.

Locking individual map operations while returning a live collection is not enough. Holding a lock while doing unnecessary response encoding can reduce concurrency, while releasing it before copying can create a race. A race-detector pass is evidence for exercised paths, not permission to omit ownership reasoning.

## 16. Topics and References for Study

Study the official `net/http` and `net/http/httptest` documentation, HTTP method and status semantics, the `Allow` header, and `Location` on `201 Created`. Study `encoding/json` decoder behavior, strict field handling, complete-document validation, and JSON number and string rules. Study `mime.ParseMediaType` for media-type validation, `sync.Mutex` for ownership protection, `sort` for deterministic output, `time.Time` formatting, and `errors` for stable error classification. Review Project 046's handler boundary and Project 014's validation discipline.

## 17. Self-Assessment Questions

1. Why is `PUT` full replacement here, and what happens when its optional body is omitted?
2. Which state must remain unchanged when update or delete targets a missing ID?
3. Why can a successful create consume one clock reading for two timestamps while a replacement consumes one for its update?
4. Why must IDs be allocated under the store's synchronization boundary?
5. Why are copies necessary even when every store method uses a mutex?
6. How can one valid JSON value be distinguished from a valid value followed by another value?
7. Why is a missing or wrong media type different from malformed JSON, and why are their statuses different?
8. Why is list sorting part of the API contract rather than presentation polish?
9. What does the race detector add to concurrent tests, and what design property must still be explained manually?
10. Why must a successful delete have no response body?

## 18. Definition of Completion

- The collection and item routes implement exactly the pinned methods, statuses, `Allow` values, `Location`, content types, and body policies.
- Notes have positive monotonic non-reused IDs, trimmed non-empty titles, preserved bodies, and correctly formatted injected timestamps.
- Full replacement preserves ID and creation time and changes update time; unknown update and delete do not mutate state.
- Lists are empty arrays when empty, sorted ascending by ID, and made from copies.
- JSON requests enforce the exact media-type, 1,048,576-byte body limit, shape, unknown-field, complete-document, ID, and title rules.
- Every non-`204` response with a body uses the exact stable JSON error or success representation, and `HEAD` is explicitly tested.
- Concurrent creates, reads, updates, deletes, and list operations are safe under the race detector without sleeps or external networking.
- The clock and store boundaries are injected and deterministic in tests.
- Only standard-library facilities are used, and the learner can explain each status and edge-case decision.

## 19. Optional Extensions

1. Add bounded, deterministic pagination and filtering to collection listing, with a separately documented query contract and ordering tests.
2. Add a persistence adapter that saves and restores notes using the atomic JSON-file lessons from Project 017 without changing the in-memory API contract.
