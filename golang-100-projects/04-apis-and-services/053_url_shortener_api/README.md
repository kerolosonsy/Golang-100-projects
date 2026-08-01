# Project 053 — URL Shortener API

## 1. Project Name and Number

Project 053 — URL Shortener API, located in `053_url_shortener_api`.

## 2. Project Idea

Build a small HTTP service that accepts a long URL through a strict JSON request, generates a fixed-length short code from a cryptographic random source, stores the code-to-URL mapping through an in-memory store interface, and redirects a follow-up `GET /{code}` to the stored URL with a pinned status. The service does not fetch the destination, does not resolve private networks, does not support custom aliases, expiry, analytics, or authentication, and is tested in memory with `httptest`.

## 3. Why This Project Now?

Project 052 served static assets with deterministic caching and exact route grammar, and Project 046 is the `net/http` foundation this project extends. This project introduces a clean store boundary, a generated-code contract, collision handling, and a small test-injectable cryptographic source. The new step is to combine strict input validation, deterministic random generation, an interface-based store, and uniform redirect semantics.

## 4. Prerequisites

Complete Projects 052 and 046 before starting. Earlier projects may be useful review, but they are not required prerequisites.

You should already be able to construct independently testable `net/http` handlers with `httptest`, distinguish `201`, `302`, `400`, `404`, `405`, `413`, `415`, `500`, and `503` semantically, validate JSON strictly, return a `Location` header on `201 Created`, and reason about map iteration order. You should also know how to read from `crypto/rand` with an unbiased mapping and how to inject a deterministic source into a constructor.

## 5. What You Must Know Before Starting

- `POST /links` accepts a JSON object whose only allowed field is the destination URL. The request body is capped at exactly 8,192 bytes, exactly 8 kibibytes. A body of exactly 8,192 bytes that is otherwise valid is accepted; the first byte beyond is `413 Payload Too Large`. The cap is enforced before JSON parsing so an oversized body is rejected without consuming parser memory.
- The URL must be an absolute `http` or `https` URL with a non-empty host. A URL with another scheme, an empty host, a missing scheme, or malformed syntax is invalid.
- The successful response is `201 Created` with a JSON document containing the generated code, the canonical short path, and the stored URL echoed back exactly as accepted through the standard JSON encoder. The response includes a `Location` header that points to the redirect endpoint for the generated code. The success body has exactly three string fields: `code`, `short_path`, and `url`.
- `GET /{code}` redirects to the stored URL with the pinned status `302 Found`. The response includes a `Location` header whose value is exactly the stored URL. There is no body. The stored URL is echoed back as the `Location` value; the service does not fetch the destination URL and must safely JSON-encode it whenever it appears in a JSON envelope.
- A code that does not exist in the store returns `404 Not Found`. The response uses the documented JSON error envelope.
- The store boundary distinguishes four outcomes for a lookup: found, not found, collision, and internal error. These outcomes are testable in memory without contacting external services.
- A short code is exactly 8 characters long. The alphabet is exactly this ordered sequence: `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`. The first ten characters are the digits `0` through `9`; the next twenty-six are the uppercase letters `A` through `Z`; the final twenty-six are the lowercase letters `a` through `z`, for a total of 62 symbols. Production code draws from `crypto/rand` using unbiased rejection sampling: random bytes are read, values that fall outside the largest whole multiple of the alphabet size are rejected, and the remaining values index into the alphabet directly. No naive modulo is used because it would skew the distribution. The implementation must use this rejection-sampling property over the documented alphabet; it must not be implemented as a modulo over a pseudo-random source.
- Collision handling is bounded. The service makes at most five total generation and store attempts per create. Attempts one through four retry on collision. Collision on attempt five returns `503 Service Unavailable`. Non-collision internal errors from the generator or the store immediately return a generic `500 Internal Server Error`. Existing entries are never overwritten, and successful creates never produce a code that collides with an existing entry.
- Code validation runs before the store lookup. A code that is not exactly 8 characters, contains a character outside the documented alphabet, contains a path separator, or is otherwise malformed is treated as not found.
- The service does not fetch the destination URL. There is no SSRF surface by construction. The service does not perform DNS resolution, TCP connection establishment, HTTP fetches, or previews.
- The service does not support custom aliases, expiry, analytics, authentication, a database, preview pages, or private-network access. They are out of scope.

## 6. Explanation of New Concepts

A short code is a small identifier drawn from a fixed alphabet. The documented alphabet is 62 symbols ordered as the digits `0` through `9`, then the uppercase letters `A` through `Z`, then the lowercase letters `a` through `z`. The fixed length of exactly 8 characters with the fixed alphabet gives a fixed total address space and a fixed URL shape.

`crypto/rand` produces uniformly distributed random bytes. Mapping those bytes to a symbol alphabet without bias requires rejection sampling: bytes that fall in the top of the range, where a naive modulo would skew the distribution toward the lower symbols, are rejected and a new byte is read. The remaining bytes index into the alphabet directly. This property is what makes the code distribution uniform over the address space.

The store boundary is an interface with a small fixed surface area. The store distinguishes four outcomes for the documented methods: stored, found, not found, collision, and internal error. The boundary exists so tests can inject a deterministic store that simulates collisions, errors, and exhaustion. Returning a single error category for all four cases hides real failures and breaks the test boundary.

The 8,192-byte body cap protects the service from unbounded input. A body of exactly 8,192 bytes that is otherwise valid is accepted; the first byte beyond is `413 Payload Too Large`. The cap is enforced before JSON parsing so an oversized body is rejected without consuming parser memory.

Strict JSON validation rejects unknown fields, missing required fields, trailing whitespace that contains a second value, and wrong field types. A successful decode is exactly one JSON object with the documented fields, followed only by JSON whitespace.

The redirect status is `302 Found`. The service pins this status and does not vary it. The `Location` header value is the exact stored URL. The service does not modify the stored URL before emitting it because doing so would change the contract. The same URL is echoed back through the JSON encoder in the create response: the create body intentionally includes the accepted URL string, the standard JSON encoder performs normal escaping for safety, and the service must not invent manual escaping rules.

The service never resolves, fetches, or otherwise touches the destination URL. There is no DNS lookup, no TCP connection, no HTTP fetch, and no preview. By construction, the service is not a vector for SSRF. Tests do not require any external connectivity.

Internal and public responses are kept distinct. A non-collision internal generator or store error returns a generic `500 Internal Server Error` with the documented JSON error envelope and code `internal_error`. Collision exhaustion on attempt five returns `503 Service Unavailable` with code `service_unavailable`. Validation, body, media, and not-found failures keep their stable mappings. The success envelope never appears in error responses, and success has only `code`, `short_path`, and `url` as its top-level fields.

## 7. Learning Objective

By completion, you can build an interface-based in-memory store, generate unbiased fixed-length codes from a cryptographic source using rejection sampling, handle a small bounded number of collisions deterministically, validate a URL at the HTTP boundary, redirect with a pinned status, and explain why the absence of outbound fetching removes an entire class of risk. You can also explain why strict JSON validation matters, why the store interface must distinguish four outcomes, why the service must echo the accepted URL safely through the standard JSON encoder without manual escaping, and why the URL needs both URL-parser validation and JSON-encode-safe handling.

## 8. Functional Requirements

1. Expose `POST /links` for creation and `GET /{code}` for redirect. The two routes are the only routes the server serves.
2. Accept JSON request bodies whose media type is `application/json` with valid optional media-type parameters. Missing or unrelated media types return `415 Unsupported Media Type` with the documented JSON error envelope.
3. Pin the request body cap at exactly 8,192 bytes. A body of exactly 8,192 bytes that is otherwise valid is accepted; a body of 8,193 or more bytes returns `413 Payload Too Large`. The cap is enforced before JSON parsing.
4. Parse the request as a JSON object with exactly one allowed field named `url`. No other fields are accepted. A missing field, an extra field, a non-string value, an empty string, or a malformed JSON document returns `400 Bad Request`. Unknown fields and a trailing second JSON value after the first object are rejected.
5. Validate the URL as an absolute `http` or `https` URL with a non-empty host. A URL with another scheme, a missing scheme, an empty host, a malformed authority, or otherwise invalid syntax returns `400 Bad Request`. The validation result is independent of network reachability.
6. Generate a short code of exactly 8 characters from `crypto/rand`. The alphabet is the documented ordered alphabet of 62 symbols. The mapping uses unbiased rejection sampling over the alphabet: random bytes are read, out-of-range values are rejected and reread, and the accepted values index into the alphabet directly. No modulo mapping is used.
7. Inject the random source into the service constructor. Tests inject a deterministic source that returns a fixed sequence of bytes. A production source reads from `crypto/rand`. The contract for what the source returns is the same regardless of the implementation.
8. Store the code through the in-memory store boundary. The store distinguishes stored, found, not found, collision, and internal error outcomes for the documented methods. The boundary is an interface with a small fixed surface area.
9. Pin collision handling to at most five total generation and store attempts per create. Attempts one through four retry on collision. Collision on attempt five returns `503 Service Unavailable`. Non-collision internal errors from the generator or the store immediately return a generic `500 Internal Server Error`. Existing entries are never overwritten.
10. Return `201 Created` on success with a JSON document containing the generated `code`, the canonical short path, and the stored URL echoed back exactly as accepted. The JSON encoder performs normal escaping for safety; the service does not invent manual escaping rules. The response includes a `Location` header whose value is the redirect path for the generated code. The content type is `application/json; charset=utf-8`. The success envelope has only `code`, `short_path`, and `url` as top-level fields.
11. Reject unsupported methods on the two routes with `405 Method Not Allowed`. `POST /links` allows `POST`; `GET /{code}` allows `GET`. The `Allow` header is sorted. `HEAD` is not implied.
12. Validate the code segment of `GET /{code}` before consulting the store. A code that is not exactly 8 characters, contains a character outside the documented alphabet, contains a path separator, or is otherwise malformed is treated as not found and returns `404`.
13. Redirect with the pinned status `302 Found` and a `Location` header whose value is the stored URL exactly. The redirect does not fetch the destination, does not modify the stored URL, and does not include a body. The response is otherwise empty.
14. Return `404 Not Found` with the documented JSON error envelope when the code is not found in the store or when the code is malformed. The error envelope distinguishes not found from internal error.
15. Distinguish internal store errors from not found and from collision. A non-collision internal error returns a generic `500 Internal Server Error` with code `internal_error`. Collision exhaustion returns `503 Service Unavailable` with code `service_unavailable`. Validation, body, media, and not-found failures keep their stable mappings: `400` with `invalid_request`, `413` with `payload_too_large`, `415` with `unsupported_media_type`, and `404` with `not_found`.
16. Concurrency is safe at the store boundary. Concurrent creates may complete in any order, but each successful create stores a unique code. Concurrent reads are safe. The race detector passes under a parallel test suite.
17. The service does not fetch the destination, does not support custom aliases, does not support expiry, does not expose analytics, does not authenticate users, does not persist to a database, and does not render a preview page. Those concerns are out of the required scope.
18. The duplicate-long-URL policy is pinned: creating a second short code for an already-stored URL does not reuse the existing code, returns a fresh code with the same contract, and never returns the existing code as the result. The policy is documented in the verification cases.
19. The JSON success envelope contains exactly `code`, `short_path`, and `url`. The JSON error envelope contains exactly `code` and `message`. Error codes are stable public values: `invalid_request`, `payload_too_large`, `unsupported_media_type`, `not_found`, `method_not_allowed`, `service_unavailable`, and `internal_error`. There is no `created` error code, because `created` is not a success envelope field and not an error code.

## 9. Inputs and Outputs

The creation request is a JSON object with exactly one field, `url`, whose value is an absolute `http` or `https` URL string with a non-empty host. The request body must not exceed exactly 8,192 bytes. The request media type must be `application/json` with valid optional media-type parameters.

Text-only creation example: a request with a JSON body containing a single field `url` whose value is an absolute `https` URL with a non-empty host produces a `201 Created` response. The response body is a JSON document whose top-level fields are `code`, `short_path`, and `url`. The `url` value matches the request input exactly as the standard JSON encoder escapes it. The response includes a `Location` header whose value is the redirect path for the generated code.

Text-only redirect example: a `GET` to the redirect path with the generated code produces a `302 Found` response with a `Location` header whose value is exactly the stored URL and zero body bytes.

Text-only failure examples: a request with a missing `url` field produces `400` with code `invalid_request`; a request with a non-string `url` produces `400`; a request with an `ftp` URL or a non-absolute URL or a URL missing the host produces `400`; a request whose body is 8,193 or more bytes produces `413` with code `payload_too_large`; a body of exactly 8,192 bytes is accepted if otherwise valid; a request with a missing or wrong media type produces `415` with code `unsupported_media_type`; a request whose body is not a JSON object produces `400`; a request whose body contains unknown fields produces `400`; a request whose body contains a trailing second JSON value produces `400`; a request whose code is malformed produces `404` with code `not_found`; a request whose code is not found produces `404`; collision on attempt five produces `503` with code `service_unavailable`; a non-collision generator or store internal error produces `500` with code `internal_error`.

The successful JSON document contains exactly `code`, `short_path`, and `url`. The error JSON document contains exactly `code` and `message`. The stable public error codes are `invalid_request`, `payload_too_large`, `unsupported_media_type`, `not_found`, `method_not_allowed`, `service_unavailable`, and `internal_error`.

## 10. Rules and Edge Cases

A URL is invalid when it is missing the scheme, has a non-`http` and non-`https` scheme, has an empty host, has an authority that cannot be parsed, or has invalid syntax. The validation does not attempt to resolve the host and does not require a public IP. The validation does not strip fragments or query strings; they are preserved as part of the stored URL.

The code length is exactly 8 characters. The alphabet is the documented ordered alphabet. The mapping uses rejection sampling so that every character is equally likely. Exactly 8 characters and the fixed alphabet give a fixed URL shape that does not change between processes.

A collision occurs when a generated code already exists in the store. The service retries generation at most five total attempts per create. Each retry uses a fresh sample from the random source. Collision on attempt five returns `503 Service Unavailable` with the documented envelope. A non-collision generator or store internal error at any point returns a generic `500 Internal Server Error` with the documented envelope. Existing entries are never overwritten.

A code that is not exactly 8 characters, contains a character outside the documented alphabet, contains a path separator, contains an encoded separator, or is otherwise malformed is treated as not found and returns `404`. The store is not consulted for malformed codes.

The duplicate-long-URL policy is pinned: a second create for an already-stored URL returns a fresh code with the same contract, not the existing code. The policy is tested in the verification cases. The service does not promise idempotency on creation.

The store boundary distinguishes stored, found, not found, collision, and internal error outcomes. The boundary returns are explicit so tests can inject a deterministic store. The boundary is not implemented as a plain map lookup that conflates not found with collision.

The service does not fetch the destination. There is no DNS lookup, no TCP connection, no HTTP fetch, and no preview. The stored URL is echoed through the JSON encoder in the create response and through the `Location` header on redirect. The service must safely JSON-encode the URL whenever it appears in a JSON envelope; the standard JSON encoder handles this. The service never offers a body on the redirect and never modifies the stored URL.

Internal responses are kept distinct from public responses. A non-collision internal error returns `500` with `internal_error` and a generic message. A collision exhaustion returns `503` with `service_unavailable`. Validation, body, media, and not-found failures keep their stable mappings.

## 11. Project Constraints

Use only the Go standard library. Use `net/http`, `encoding/json`, `crypto/rand`, `strings`, `net/url`, `sync`, and `net/http/httptest`. Do not use a web framework, a database, a third-party URL shortener library, a third-party random library, a third-party URL validator, or an external HTTP client. Do not perform any outbound HTTP, DNS, or TCP requests.

The exact 8,192-byte body cap, the exact 8-character code length, the exact documented alphabet order, the exact at-most-five total attempts collision policy, and the exact stable public error codes are required learning contracts and are part of this document. Do not include implementation code, function signatures, or pseudocode that describes the rejection-sampling algorithm step by step in this guide. The guide states the property and the alphabet, not the algorithm.

## 12. Design Questions Before Coding

- What interface surface does the store boundary expose, and how does it distinguish stored, found, not found, collision, and internal error?
- How is the documented alphabet ordered, and how is the rejection-sampling property enforced so that no modulo bias is introduced?
- How is the random source injectable into the constructor so tests can drive a deterministic sequence?
- How is the URL parsed and validated without contacting any network or DNS resolver?
- How is the 8,192-byte body cap applied before JSON parsing, and how is the JSON decoder configured for strict validation?
- How is the at-most-five retry counter bounded across generation and store attempts, and how is exhaustion reported as a service error?
- How is the `Location` header set on `201 Created` and on the redirect, and how are the two values distinguished?
- How is the code segment of `GET /{code}` validated before the store is consulted?
- How is concurrent creation safe at the store boundary without losing codes?
- How is the duplicate-long-URL policy enforced so a fresh code is returned for every successful create?
- How does the standard JSON encoder safely encode the stored URL in the create response, and why is manual escaping not invented?

## 13. Implementation Milestones

1. Record the routes, method policies, status mappings, error envelope, the exact 8,192-byte body cap, the exact 8-character code length, the exact alphabet order, the exact at-most-five retry bound, and the stable public error codes as testable acceptance criteria.
2. Establish injectable store and cryptographic random-source boundaries with the documented outcomes and identical production and test contracts.
3. Enforce media type, the exact 8,192-byte cap, single-object shape, unknown-field rejection, and trailing-second-value rejection for strict `POST /links` JSON validation.
4. Accept only absolute `http` or `https` URLs with a non-empty host, without contacting any network.
5. Generate exactly 8-character codes over the documented alphabet from the random source with the required rejection-sampling property.
6. Store each mapping with at most five attempts on collision, return `503 Service Unavailable` on exhaustion, and ensure duplicate long URLs receive fresh codes rather than existing ones.
7. Return generic `500 Internal Server Error` immediately for non-collision generator and store internal errors.
8. Return `201 Created` with `Location` and the documented JSON envelope encoded by the standard JSON encoder.
9. Redirect with the pinned `302 Found` status, the exact stored URL in `Location`, and zero body bytes.
10. Validate `GET /{code}` before store access and map malformed or missing codes to `404`.
11. Apply sorted `Allow` headers and the documented `400`, `404`, `405`, `413`, and `415` policies across known and unknown paths.
12. Finish deterministic, boundary, concurrency, and race-detector verification without any external network access.

## 14. Verification Cases the Learner Must Write

- Create a link with a valid `https` URL. Verify `201`, the exact JSON envelope with `code`, `short_path`, and `url` only, the `Location` header, and that the `url` field matches the input exactly as the standard JSON encoder escapes it.
- Create a link with a valid `http` URL. Verify the same contract as `https`.
- Submit a body of exactly 8,192 bytes that is otherwise valid. Verify it is accepted and a `201` is reachable. Submit a body of 8,193 bytes. Verify `413` before any JSON parse and no store mutation.
- Submit a missing media type, an unrelated media type, a malformed media type, and a media type with an invalid parameter. Verify `415` for each case.
- Submit a missing `url` field, a non-string `url`, an empty `url`, an unknown field, a `null` body, a non-object body, a malformed JSON document, a second JSON value, and trailing whitespace after a valid value. Verify `400` for each case and no store mutation.
- Submit a URL whose scheme is `ftp`, `file`, `data`, or empty. Verify `400`.
- Submit a URL with an empty host, a missing host, or a malformed authority. Verify `400`.
- Submit a URL whose fragment or query string is preserved. Verify the stored URL and the redirect target match the input.
- Redirect through the generated code. Verify `302`, the exact `Location` header, and zero body bytes.
- Submit a code that is not exactly 8 characters, contains a non-alphabet character, contains a path separator, or contains an encoded separator. Verify `404` without consulting the store.
- Submit a code that is not found. Verify `404` with the documented JSON envelope and content type.
- Inject a random source that returns a fixed sequence. Verify the generated code is exactly 8 characters drawn from the documented alphabet at the documented positions.
- Inject a store that reports a collision on attempt one and success on attempt two. Verify the success response and the final stored code.
- Inject a store that reports a collision on attempts one through four and a collision on attempt five. Verify `503 Service Unavailable` with code `service_unavailable` and that no entry is overwritten.
- Inject a generator that returns an invalid output or that fails. Verify generic `500 Internal Server Error` with code `internal_error` and no successful response.
- Inject a store that returns a non-collision internal error. Verify generic `500 Internal Server Error` with code `internal_error` and no successful response.
- Submit an unsupported method on `POST /links` and on `GET /{code}`. Verify `405`, sorted `Allow`, and the documented body policy.
- Create the same long URL twice. Verify each creation returns a fresh code and that the existing code is never reused.
- Run many concurrent creates against the same store. Verify unique codes, the exact count of successful creates, and that the race detector reports no issue.
- Verify no test performs an outbound HTTP request, a DNS lookup, or a TCP connection. Verify no test depends on the network or on a fixed port.

## 15. Common Mistakes to Watch For

Using a naive modulo to map random bytes to the documented alphabet biases the distribution toward the lower symbols and reduces the effective address space. Reading random bytes only once per code and not rejecting out-of-range values produces skewed codes. Generating a code with a length other than exactly 8 characters makes the URL shape unpredictable and the verification contract ambiguous.

Treating the store as a plain map lookup conflates not found, collision, and internal error. Returning a single error for all three cases hides real failures and breaks the test boundary. Returning collision as not found makes retry logic impossible to verify. Returning collision exhaustion as `500` collapses the distinct exhaustion contract into the generic internal-error contract.

Trusting the URL without parsing or validation accepts non-`http` and non-`https` schemes, malformed authorities, and empty hosts. The service would still not fetch the destination, but the stored URL would no longer be a valid redirect target and the contract would be ambiguous.

Reading the body without the exact 8,192-byte cap permits unbounded memory growth. Treating the cap as advisory lets oversized bodies through. Decoding JSON with default options accepts unknown fields and trailing second values. Inferring the media type from the body accepts non-JSON requests.

Treating a duplicate long URL as an idempotent request and returning the existing code breaks the contract for fresh-code creation. Idempotency at the URL level is a separate feature and is not part of this project.

Redirecting with a status other than `302 Found`, modifying the stored URL before emitting it, including a body on the redirect, or fetching the destination as part of the redirect violates the contract. The redirect is a header-only response.

Inventing manual escaping rules outside the standard JSON encoder risks breaking the create response shape and the cross-handler consistency. Echoing the accepted URL through manual concatenation can produce an invalid JSON envelope for inputs that contain quotes, backslashes, or control characters.

Returning a successful response after a partial mutation or after an internal error leaks state. Holding a lock while doing unnecessary JSON encoding reduces concurrency. Using a global random source prevents deterministic tests.

## 16. Topics and References for Study

Study the official documentation for `net/http`, especially `ServeMux` patterns, `Redirect`, `NotFound`, `Error`, and the documented behavior of `Location`. Study `encoding/json` decoder options, strict field handling, and complete-document validation. Study `crypto/rand` for cryptographic random byte reads. Study `net/url` for `Parse`, `ParseRequestURI`, scheme rules, and host extraction. Study `sync.Mutex` and `sync.RWMutex` for ownership protection. Review RFC 3986 for URI syntax, RFC 7231 for redirect semantics, and the documented practices for unbiased random sampling over fixed alphabets.

## 17. Self-Assessment Questions

1. Why is rejection sampling required for unbiased code generation rather than a naive modulo, and why is the random source injected?
2. Why must the store distinguish stored, found, not found, collision, and internal error as separate outcomes?
3. Why does the service not fetch the destination URL, and what class of risk does that remove?
4. Why must the body cap of exactly 8,192 bytes be enforced before JSON parsing?
5. Why is the duplicate-long-URL policy not idempotent, and what does that mean for the contract?
6. Why is the redirect status `302 Found` rather than `301 Moved Permanently` or `307 Temporary Redirect`?
7. Why must the code segment be validated before the store is consulted?
8. Why is concurrency at the store boundary necessary even though the in-memory store is fast?
9. Why does the service echo the accepted URL in the create response while the redirect emits the same URL in `Location`, and how does the standard JSON encoder keep the echo safe?
10. Why are generation and storage limited to exactly five total attempts, and why is fifth-collision exhaustion a `503 Service Unavailable` rather than a generic `500 Internal Server Error`?

## 18. Definition of Completion

- `POST /links` accepts a strict JSON object with exactly one valid `url` field and returns `201` with the documented envelope and `Location`.
- The body cap is exactly 8,192 bytes; exactly that is accepted and one byte more is `413`.
- `GET /{code}` redirects with `302 Found`, the exact stored URL in `Location`, and zero body bytes.
- The store boundary distinguishes stored, found, not found, collision, and internal error and is testable through an injected implementation.
- Code generation uses rejection sampling over the documented alphabet of 62 symbols to produce exactly 8 characters.
- Retry on collision is bounded to at most five total attempts per create; collision on the fifth attempt returns `503`; non-collision generator or store internal errors return `500`.
- Malformed codes return `404` without consulting the store; missing codes return `404` with the documented envelope.
- Unsupported methods, unknown paths, missing or wrong media types, oversized bodies, malformed JSON, and unknown or trailing fields map to the documented statuses without leaking internal causes.
- The duplicate-long-URL policy returns a fresh code for every successful create and never returns the existing code.
- The accepted URL is echoed safely through the standard JSON encoder in the create response; no manual escaping is invented.
- Concurrent creates are safe under the race detector without external network access and without outbound HTTP.
- The service uses only the Go standard library, and the learner can explain each policy and trade-off without referring to implementation syntax.

## 19. Optional Extensions

1. Add a deterministic, in-memory analytics counter that records the number of redirects per code, with no expiration and no persistence.
2. Add a deterministic duplicate-URL short-circuit that returns the existing code for the same exact URL, with the existing-code behavior documented as a contract change.
