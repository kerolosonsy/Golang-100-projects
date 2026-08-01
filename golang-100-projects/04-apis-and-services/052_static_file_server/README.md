# Project 052 — Static File Server

## 1. Project Name and Number

Project 052 — Static File Server, located in `052_static_file_server`.

## 2. Project Idea

Build a small HTTP service that serves a fixed asset tree compiled into the binary through `embed.FS`. The service exposes the assets under `/assets/`, exposes one pinned index route at `GET`/`HEAD /` that serves the embedded `index.html`, and uses a custom HTML `404` page for every other path. The service generates strong ETag values deterministically from the embedded bytes at construction, pins one `Cache-Control` policy for normal assets and the index, and pins a distinct `no-store` policy for the custom `404` and method-error responses. The service supports `GET` and `HEAD`, applies `If-None-Match` only as exact strong-tag equality, and never lists directories or reads from disk at request time.

## 3. Why This Project Now?

Project 051 produced uploaded files whose integrity and naming policy this project must respect when they later become assets, and Project 046 is the `net/http` foundation this project extends. This project introduces embedding, deterministic cache headers, an exact route grammar, and exact `If-None-Match` semantics that the later CRUD and middleware projects rely on. The new step is to serve bytes without ever trusting request paths beyond a fixed prefix, to generate cache metadata from content, and to keep the asset tree inside the binary rather than on disk.

## 4. Prerequisites

Complete Projects 051 and 046 before starting. Earlier projects may be useful review, but they are not required prerequisites.

You should already be able to construct independently testable `net/http` handlers with `httptest`, distinguish `200`, `304 Not Modified`, `404`, and `405` semantically, set headers before body, validate request paths, and reason about byte-exact responses. You should also know how `embed.FS` compiles a directory tree into the binary and how `fs` and `io/fs` work for static content.

## 5. What You Must Know Before Starting

- `embed.FS` is a read-only filesystem compiled at build time. The directory layout at build time becomes the layout served at runtime. Nothing on disk is read by the serving process.
- The asset root is `/assets/` and the index route is `/`. Every other path is unmapped unless it points to a successfully matched asset under `/assets/`. Requests that look like `/assets/` itself or other directories are not listings and are not redirected to an index; they return the custom `404` page. There is no implicit directory index for any directory, including `/assets/`.
- A request that asks for `/assets/../something` or `/assets/foo/../bar` or `/assets/%2e%2e/etc` is a path-traversal attempt and must not reach the embedded tree. The service must not clean the path into a different resource; it must reject it.
- An encoded separator inside a request path is ambiguous. A request that uses percent-encoded slashes, backslashes, or dots to suggest a different path must be rejected, not silently decoded into a permitted path.
- MIME detection is deterministic. For embedded files whose extension is in the pinned map below, the response uses the mapped value exactly. For unknown extensions or no extension, the response uses bounded standard content detection on a prefix of the embedded bytes, and the type returned by the detector is used. The pinned map is authoritative for the listed extensions.
- The service supports exactly `GET` and `HEAD`. Any other method returns `405 Method Not Allowed` with a sorted `Allow` header.
- A `HEAD` response has the same status and the same headers as the corresponding `GET` response but no body. The server must not silently convert `HEAD` into `GET`; it must serve the headers and then stop.
- A strong ETag is a quoted lowercase hexadecimal SHA-256 digest of the exact embedded representation bytes, computed at construction or startup. It is the same on every request for the same file and changes only when the file content changes.
- `Cache-Control` is pinned exactly: normal successful asset responses and the index route use `public, max-age=3600`. The custom `404` page and method / error responses use `no-store`. The custom `404` is not long-cacheable.
- `If-None-Match` has exactly one consistent scope. A header value consisting of exactly the current single strong tag for the resource is a match. Any other header value — a weak tag, the wildcard form, a comma-separated list of one or more tags, a malformed value, or any other text — is unsupported and is treated as a non-match that produces the normal `200` response. The conditional handling applies to successful assets under `/assets/` and to the index `/`. The custom `404` remains `404` even though its body has an ETag computed from its own bytes.
- A matching `GET` or `HEAD` returns `304 Not Modified`. The response has no body, no `Content-Type`, and no `Content-Length`. It includes the current `ETag` and `Cache-Control`. Only the documented subset of `200` headers is included on `304`; the service does not copy every header from the `200` response to the `304` response.
- The handler must be testable through `httptest` without binding a fixed port.

## 6. Explanation of New Concepts

`embed.FS` is a filesystem value built at compile time from the contents of one or more directories in the source tree. The resulting binary contains the files as data, not as paths on disk. The service uses the embedded tree rooted at the asset directory and serves it under `/assets/`. The embedded tree is the only source of bytes the service ever returns for asset requests.

A strong ETag is a token derived from the actual bytes of a resource. A quoted lowercase hexadecimal SHA-256 digest of the exact embedded representation bytes is a strong tag in this project. It changes whenever the bytes change, and it is the same across processes that embed the same content. A weak ETag is a token derived from semantic equivalence and may be shared by representations that differ in unimportant ways. The project uses strong tags from the embedded bytes only; it does not use timestamps, file sizes, or modification times that could vary between builds.

The route grammar has two mapped entries and one custom-404 sinkhole. `/assets/{path}` is matched against the embedded tree. `/` is the pinned index route, served only when the embedded `index.html` is present. Any other path, including `/assets/` itself, is answered with the custom `404` page or with `405` for an unsupported method.

Path containment is verified before the suffix is matched. The matcher accepts the prefix `/assets/` and discards the rest only when the rest has no `..` segments, no encoded separators, no raw backslashes, and no double-encoded forms. This prevents a request from being reinterpreted as a path outside the embedded tree.

`Content-Type` is decided by a deterministic rule. For files whose extension is in the pinned map below, the rule is the documented mapping. For unknown extensions or no extension, the rule is the bounded standard content-detection reading on a prefix of the embedded bytes, with the returned type used as the response `Content-Type`. The header is set explicitly; the handler does not let the response writer infer it from body bytes.

`Cache-Control` is a pinned directive pair. Successful assets and the index use `public, max-age=3600`. The custom `404` page and every method or error response use `no-store`. The pinned values are part of the contract and are asserted by tests.

`HEAD` semantics are honored explicitly. The handler runs the same path and method-resolution logic, computes the same headers, runs the same `If-None-Match` evaluation, writes the headers, and stops before the body. The handler does not delegate `HEAD` to `GET` and does not skip the ETag check.

`If-None-Match` is evaluated with one consistent rule. A request whose `If-None-Match` value is exactly the current strong ETag for the resource returns `304 Not Modified` with no body, no `Content-Type`, no `Content-Length`, the current ETag, and the current `Cache-Control`. Any other header value is treated as a non-match and produces the normal `200` response with the full asset body. Weak tags, the wildcard form, comma-separated lists, and malformed values all fall under this non-match rule. The same rule applies on `HEAD`, and a matching `HEAD` returns `304 Not Modified` with zero body bytes.

The custom `404` page is itself a pinned asset. It is served with its own ETag — the strong SHA-256 digest of its own embedded bytes — but the response status remains `404`, and the `Cache-Control` is `no-store`, not `public, max-age=3600`. Even though the body has an internal ETag, the page is never matched into `304` by the conditional machinery: conditional handling applies only to successful assets and `/`.

## 7. Learning Objective

By completion, you can serve a compiled asset tree under a fixed route prefix, generate deterministic strong ETag values from the embedded bytes at construction, pin exact `Cache-Control` policies for normal assets versus the custom `404` and method-error responses, handle `HEAD` and `GET` distinctly, reject path-traversal and encoded-separator attempts, return a custom HTML `404` for every unmapped path, and decide `304 Not Modified` based on exact single-tag `If-None-Match` equality. You can also explain why the asset tree is embedded rather than read from disk, why strong ETags are pinned to lowercase hex SHA-256 digests, why the custom `404` page uses `no-store`, and why `If-None-Match` does not accept weak tags, lists, wildcards, or malformed values.

## 8. Functional Requirements

1. Expose the embedded asset tree under the fixed prefix `/assets/`. Use `embed.FS` to compile the asset tree into the binary and serve its contents from there.
2. Expose one pinned index route at `GET` and `HEAD /` that returns the embedded `index.html`. The index route is the only non-`/assets/` route the server serves.
3. Serve a custom HTML `404` page for every path that is not a successfully matched asset under `/assets/` and is not the index `/`. This includes the asset root `/assets/`, other directories under `/assets/`, traversal-style paths, encoded-separator paths, and any other unmapped request. The custom `404` page is a pinned asset with its own ETag, but the response status remains `404` and the `Cache-Control` is `no-store`. There is no implicit directory index for any directory.
4. Reject path-traversal attempts. A request whose normalized path contains a `..` segment, an encoded separator, a raw backslash, or any double-encoded form that would resolve outside the asset prefix returns `404` with the custom `404` page. The service does not clean the path into a permitted resource.
5. Reject directory listings. A request that names a directory rather than a file returns `404` with the custom `404` page unless that directory is the index route `/`. The response never contains a list of entries and never contains a generated index page.
6. Use a deterministic `Content-Type` rule with an explicit pinned map for embedded known extensions: HTML `text/html; charset=utf-8`, CSS `text/css; charset=utf-8`, JavaScript `text/javascript; charset=utf-8`, plain text `text/plain; charset=utf-8`, PNG `image/png`, JPEG `image/jpeg`, and SVG `image/svg+xml`. For unknown extensions and for the custom `404` page, run a bounded standard content detection on a prefix of the embedded bytes and use the detector's returned type. The header is set before the body is written.
7. Support `GET` and `HEAD`. `GET` returns the asset bytes with `200 OK`, the explicit `Content-Type`, the strong ETag, the pinned `Cache-Control`, and `Content-Length`. `HEAD` returns the same status, the same headers, and zero body bytes.
8. Reject every other method with `405 Method Not Allowed`. The response includes a sorted `Allow` header whose value is `GET, HEAD`. The body is empty for `HEAD` and follows the documented error envelope for other methods. The response uses `no-store`.
9. Compute the strong ETag at construction or startup as a quoted lowercase hexadecimal SHA-256 digest of the exact embedded representation bytes for each asset and for the index and for the custom `404` page. The ETag is the same on every request for the same file and is recomputed only when the embedded content changes.
10. Pin `Cache-Control` for every successful asset response and for the index route to `public, max-age=3600`. Pin `Cache-Control` for the custom `404` page and for every method or error response to `no-store`. The two policies are distinct.
11. Apply `If-None-Match` exactly. A header value consisting of exactly the current single strong ETag for the resource is a match. Weak tags, the wildcard form, comma-separated lists of one or more tags, malformed values, missing values, and any other text are unsupported and are treated as non-matches that produce the normal `200` response. The conditional handling applies to successful assets and `/`.
12. A matching `GET` or `HEAD` returns `304 Not Modified`. The response has no body, no `Content-Type`, no `Content-Length`. It includes the current `ETag` and `Cache-Control`. Only the headers explicitly allowed for `304` are present. A non-matching `If-None-Match` returns the normal `200` with the full body.
13. The custom `404` page is `404`, not `304`, regardless of any internal ETag on its body. Conditional handling does not collapse `404` into `304`.
14. Use `net/http/httptest` for handler tests. Tests do not bind a fixed port and do not depend on external networking. Tests run the same handler concurrently and under the race detector.
15. Compose the handler tree so that the asset route and the index route are independent from the listener startup. The handler can be constructed, inspected, and tested without opening a socket.
16. The service does not serve files from disk at request time, does not serve files outside the embedded tree, does not allow uploads, does not implement authentication, does not generate listings, does not provide a directory index beyond the explicitly pinned `index.html`, does not implement other conditional headers such as `If-Modified-Since`, and does not proxy to another service.

## 9. Inputs and Outputs

The request is an HTTP method and a path. The path is matched against `/assets/` and then against the embedded tree, or against the pinned index route `/`. The request body is irrelevant for `GET` and `HEAD`; the server does not parse it.

Text-only asset example: a `GET` for an HTML asset under `/assets/` produces `200 OK`, the pinned `Content-Type` for HTML, the strong ETag token in quotes, `Cache-Control: public, max-age=3600`, the `Content-Length` value, and the file bytes in the body.

Text-only HEAD example: a `HEAD` for the same asset produces the same status and headers with zero body bytes.

Text-only 304 example: a `GET` with `If-None-Match` set to exactly the asset's current strong ETag produces `304 Not Modified`. The response has no body, no `Content-Type`, no `Content-Length`, includes the current `ETag` and `Cache-Control: public, max-age=3600`.

Text-only non-match example: a `GET` with `If-None-Match` set to a weak tag, the wildcard, a comma-separated list, a malformed value, or missing produces the normal `200 OK` response with the full body.

Text-only index example: a `GET /` produces `200 OK` for the embedded `index.html` with `Content-Type: text/html; charset=utf-8`, the strong ETag for `index.html`, and `Cache-Control: public, max-age=3600`.

Text-only 404 example: a `GET` for an unknown path under `/assets/`, for `/assets/`, for any other directory under `/assets/`, for the index of a directory that does not have a pinned `index.html`, for a traversal-style path, or for any path outside the two mapped entries produces `404 Not Found` with the custom HTML content type, the strong ETag for the custom `404` page, `Cache-Control: no-store`, and the HTML body of the custom `404` page.

Text-only 405 example: a `POST` to an asset path or to `/` produces `405 Method Not Allowed`, sorted `Allow: GET, HEAD`, and `Cache-Control: no-store`. The body uses the documented error envelope.

## 10. Rules and Edge Cases

The path is matched by prefix and by exact segment count. A request whose asset suffix has an empty segment, an extra slash, or a missing trailing component is rejected by the documented route logic. A request for `/assets/` itself is a directory request and returns the custom `404`.

Path traversal is rejected before the embedded tree is consulted. A request whose normalized path contains `..`, an encoded slash, an encoded backslash, a raw backslash, or any double-encoded form is rejected as an unknown path. The server does not clean the path into a different resource.

A request whose asset suffix matches a directory under the embedded tree returns the custom `404` unless that directory is reached through `/`. The response does not list the directory's contents and does not redirect to a canonical trailing-slash form.

The ETag is a quoted lowercase hexadecimal SHA-256 digest of the exact embedded representation bytes. A weak tag in `If-None-Match` is not equivalent to the strong tag for matching purposes and is treated as a non-match. A wildcard value is not equivalent to the strong tag and is treated as a non-match. A comma-separated list of one or more tags is treated as a non-match unless the exact strong tag is the only listed value and the value equals the current strong tag in full; the implementation uses the simpler exact equality rule and treats every other form as a non-match. A malformed tag is a non-match. A missing header is a non-match and produces the normal `200`.

`HEAD` returns the same status and headers as `GET` with zero body bytes. A `HEAD` request for an unknown path returns `404` with the custom `404` headers and zero body bytes. A `HEAD` request that would otherwise produce `304 Not Modified` returns `304` with zero body bytes.

A matching `GET` or `HEAD` returns `304 Not Modified`. The `304` response includes only the headers explicitly allowed for `304`: `ETag` and `Cache-Control`. The `304` response does not include `Content-Type`, does not include `Content-Length`, and does not include any other header from the corresponding `200` response.

The custom `404` page is a pinned asset served with its own ETag and `Cache-Control: no-store`. Tests assert the cache policy for the custom `404` separately from the cache policy for normal assets. The custom `404` remains `404` even though its body has an ETag; conditional handling never collapses a `404` into `304`.

The `Content-Type` rule is deterministic. For embedded files whose extension is in the pinned map, the response uses the mapped value exactly. For unknown extensions and for the custom `404` page, the response uses the type returned by a bounded standard content detection on a prefix of the embedded bytes. The header is set before the body is written. The handler never infers `Content-Type` from body bytes outside this documented rule.

Concurrent reads of the same asset do not race because the embedded filesystem is immutable and the ETag is computed once at startup. The race detector passes under a parallel test suite.

## 11. Project Constraints

Use only the Go standard library. Use `embed`, `io/fs`, `net/http`, `mime`, `crypto/sha256`, `encoding/hex`, `strings`, and `net/http/httptest`. Do not use a web framework, a third-party router, a third-party templating engine, a third-party asset bundler, a third-party MIME database, or a disk-based static file server. Do not implement range requests, compression negotiation, content negotiation, conditional requests beyond `If-None-Match`, or any caching protocol beyond the documented ETag and `Cache-Control` headers.

The exact `Cache-Control` policies, the exact MIME map, the exact status name `304 Not Modified`, and the exact rule that `304` carries only the documented header subset are required learning contracts and are part of this document. Do not include other implementation code, function signatures, exact byte sequences of the ETag, exact MIME strings, or the exact HTML of the custom `404` page in this guide. The guide states policies, not literal values.

## 12. Design Questions Before Coding

- Which directory in the source tree becomes the embedded asset root, and which files become pinned assets such as the index `index.html` and the custom `404` page?
- How is the sub-filesystem rooted so that the handler can never reach files outside the embedded tree?
- How is the strong ETag computed once at construction and shared with all handlers, and what hash algorithm is used?
- How is the route grammar separated between the `/assets/` prefix and the pinned index `/`, with everything else as a custom-404 sinkhole?
- How is path containment enforced before the embedded tree is consulted?
- How is `HEAD` served without duplicating the `GET` handler and without skipping the ETag check?
- How is `If-None-Match` evaluated as exact single-tag equality, with every other form treated as a non-match?
- How is the `304 Not Modified` response limited to the documented header subset, and how is the body omitted?
- How is the custom `404` page produced and pinned, and how are its ETag and `Cache-Control: no-store` distinguished from the asset policy?
- How is `Content-Type` determined from the pinned map and from content detection for unknown extensions?
- How is the handler tree composed so the listener startup is independent from handler construction?

## 13. Implementation Milestones

1. Record the route grammar, method policy, status mapping, header set, ETag rule, Cache-Control rule, exact MIME map, exact `If-None-Match` rule, and custom `404` policy as testable acceptance criteria.
2. Place the asset tree inside the source tree, place `index.html` at the root, and place the custom `404` page as a pinned asset. Embed it all through `embed.FS`.
3. Compute the strong ETag for each asset, for the index, and for the custom `404` page at construction or startup as a quoted lowercase hexadecimal SHA-256 digest of the exact embedded representation bytes, and store it in an immutable lookup table keyed by asset path.
4. Establish the path-containment matcher that accepts the `/assets/` prefix and rejects traversal, encoded separators, raw backslashes, and double-encoded forms.
5. Establish the route dispatch that maps `/` to the embedded index, maps `/assets/{path}` to the embedded tree, and sinks everything else to the custom `404`.
6. Apply the documented `Content-Type` map for known extensions and bounded standard content detection for unknown extensions and the custom `404` page.
7. Serve assets and the index through `GET` with explicit headers, ETag, `Cache-Control: public, max-age=3600`, `Content-Length`, and exact bytes.
8. Serve `HEAD` with headers matching `GET` for assets and the index, preserving ETag and `Cache-Control` while returning no body.
9. Apply exact single-tag `If-None-Match` equality, returning `304 Not Modified` with the documented header subset and no body on a match and normal `200` on every other value.
10. Serve the custom `404` page with its own ETag, `Cache-Control: no-store`, the documented `Content-Type` rule, and HTML body while ensuring it remains `404` and never becomes `304`.
11. Return `405` with sorted `Allow: GET, HEAD`, the documented error envelope, `Cache-Control: no-store`, and the required body policy.
12. Compose a listener-independent handler tree, then finish deterministic, byte-exact, concurrent, and race-detector verification and review the policy for honest strong-tag semantics.

## 14. Verification Cases the Learner Must Write

- Serve an existing HTML asset, a CSS asset, a JavaScript asset, a plain text asset, a PNG, a JPEG, and an SVG through `GET`. Verify the exact status, the exact `Content-Type` from the pinned map, the exact `Content-Length`, the strong ETag token, and `Cache-Control: public, max-age=3600`. Verify the body bytes match the embedded bytes exactly.
- Send `HEAD` for each served asset. Verify the status, all `200` headers, and zero body bytes. Verify the `Content-Length` header is present and matches the asset size.
- Send `GET` and `HEAD` for an unknown path under `/assets/`, for `/assets/`, for a directory under `/assets/` with no pinned index, for a traversal-style path with `..`, and for an encoded-separator path. Verify `404`, the custom HTML content type, the custom `404` ETag, `Cache-Control: no-store`, and the custom `404` HTML body. Verify zero body bytes for `HEAD`.
- Send `GET /`. Verify `200 OK`, the index `Content-Type`, the index ETag, `Cache-Control: public, max-age=3600`, and the exact body bytes of the embedded `index.html`. Send `HEAD /` and verify the same headers with zero body bytes.
- Send `GET` with `If-None-Match` set to the exact strong tag for an asset and for the index. Verify `304 Not Modified`, the current ETag, `Cache-Control: public, max-age=3600`, no body, no `Content-Type`, no `Content-Length`. Send `HEAD` with the same matching header and verify the same `304 Not Modified` with zero body bytes and the same allowed header subset.
- Send `GET` and `HEAD` with `If-None-Match` set to a weak tag with the same value, the wildcard, a comma-separated list with one matching tag, a comma-separated list with no matching tag, a malformed value, and no header at all. Verify `200 OK` for assets and for `/`, with the full body and the normal `200` headers; verify `304` does not appear.
- Verify the ETag is a quoted lowercase hexadecimal SHA-256 digest of the exact embedded bytes, is the same across requests for the same file, and changes when the embedded bytes change.
- Verify `POST`, `PUT`, `DELETE`, `PATCH`, and `OPTIONS` on an asset path and on `/` return `405` with sorted `Allow: GET, HEAD`, the documented error envelope, and `Cache-Control: no-store`. Verify the body policy.
- Verify the `Cache-Control` for the custom `404` page is exactly `no-store` and is distinct from `public, max-age=3600`.
- Verify that a matching conditional request returns `304` without `Content-Type`, without `Content-Length`, and without any header not on the documented `304` subset.
- Verify that the custom `404` page remains `404` even if its body has an internal ETag, and that conditional handling never collapses it into `304`.
- Run concurrent reads of the same asset under the race detector and verify no race is reported and the bytes are unchanged.
- Verify every test uses `httptest` and no test binds a fixed port or contacts external services.

## 15. Common Mistakes to Watch For

Trusting the request path after prefix matching without checking `..` segments, encoded separators, or raw backslashes allows path traversal out of the embedded tree. Cleaning the path into a permitted resource hides the attack and changes the contract. Using `http.FileServer` against a directory on disk re-introduces filesystem dependency and breaks the embed contract.

Serving `HEAD` by reusing `GET` and discarding the body is fragile. The standard library does not always guarantee that the implicit body discard is silent or that headers are written before the body, and the contract here is explicit. Skipping the ETag check on `HEAD` defeats conditional responses.

Treating any weak tag as equivalent to the strong tag violates HTTP semantics and the documented `If-None-Match` rule. Treating any `If-None-Match` value that contains the strong tag as a match is wrong here; the rule is exact single-tag equality. Treating the wildcard as a match, treating comma-separated lists as a match by any single element, or treating malformed values as a match violates the same rule. Pinning any of those as a match is wrong.

Returning a directory listing or a generated index page violates the no-listing rule. Returning `301` to a trailing-slash form changes the route grammar and the cache behavior. Serving the custom `404` page from disk re-introduces the dependency on the working directory.

Setting `Cache-Control: public, max-age=3600` for the custom `404` page makes outages sticky and contradicts the pinned `no-store` policy. Inferring `Content-Type` from body bytes silently changes the contract when the same byte sequence is also a valid HTML document. Treating the MIME map as advisory rather than authoritative breaks the byte-exact Content-Type promise.

Computing the ETag as anything other than a quoted lowercase hexadecimal SHA-256 digest of the exact embedded bytes makes the digest shape inconsistent with the contract. Embedding the asset tree but recomputing the ETag at request time is wasteful and may race. Computing the ETag from a timestamp, from a filesystem modification time, or from a hash of the path lets the same name in different builds share an ETag and is wrong.

Allowing the `304 Not Modified` response to carry `Content-Type`, `Content-Length`, or arbitrary `200` headers leaks representation metadata for a body that has none. The `304` header subset is exactly `ETag` and `Cache-Control`; everything else is omitted.

Allowing conditional handling to collapse `404 Not Found` into `304 Not Modified` because the `404` body has an internal ETag is wrong. The custom `404` page is a `404`, always.

## 16. Topics and References for Study

Study the official documentation for `embed.FS`, `fs.FS`, `fs.Sub`, and `fs.WalkDir` for compiling and reading the asset tree. Study `net/http` `Handler`, `ServeMux`, `FileServer`, `ServeContent`, and the documented behavior of `HEAD`. Study `crypto/sha256` and `encoding/hex` for deterministic strong ETag generation. Study `mime` for the standard extension mapping and content-detection primitives. Study `net/http/httptest` for in-memory handler testing. Review RFC 7232 for HTTP conditional requests, RFC 9110 for `If-None-Match` and ETag semantics, and RFC 9111 for Cache-Control directives. Note that the status name is `304 Not Modified` in standard documentation, not `304 Not Content` or `304 Not Found`.

## 17. Self-Assessment Questions

1. Why is the asset tree embedded at build time rather than read from disk at request time?
2. Why is the ETag a strong quoted lowercase hexadecimal SHA-256 digest of the exact embedded bytes rather than a timestamp, and how does it differ from a weak tag?
3. Why must the custom `404` page use `Cache-Control: no-store` rather than `public, max-age=3600`, retain its own ETag, and remain `404` rather than becoming `304`?
4. Why is `HEAD` served by writing the headers and stopping, rather than by reusing the `GET` handler and discarding the body?
5. Why is the exact `If-None-Match` value required for `304` rather than a substring, weak-equivalent, wildcard, or comma-list match?
6. Why is a directory request returned as `404` rather than redirected to an index file when no index is pinned?
7. Why does the `304 Not Modified` response include only `ETag` and `Cache-Control`, omitting `Content-Type`, `Content-Length`, and every other `200` header?
8. Why is path containment enforced before the embedded tree is consulted?
9. Why is the route grammar separated between the `/assets/` prefix and the pinned index `/`, with everything else as a `404` sinkhole?
10. Why must the handler be testable without binding a port, and how is that achieved with `httptest`?

## 18. Definition of Completion

- The asset tree is compiled through `embed.FS` and served under `/assets/`, with no disk access at request time.
- The strong ETag for each asset, the index, and the custom `404` page is computed at construction or startup as a quoted lowercase hexadecimal SHA-256 digest of the exact embedded representation bytes and is reused on every request.
- `GET` returns `200 OK` for assets and for `/` with the exact `Content-Type` from the pinned map, the exact `Content-Length`, the strong ETag, `Cache-Control: public, max-age=3600`, and exact body bytes.
- `HEAD` returns the same status and headers as `GET` for assets and `/`, with zero body bytes.
- `If-None-Match` returns `304 Not Modified` for the exact single strong tag and `200` for every other form, with the `304` response including only `ETag` and `Cache-Control` and no body, `Content-Type`, or `Content-Length`.
- Path traversal, encoded separators, raw backslashes, and double-encoded forms return `404` with the custom `404` page; the service never cleans a path into a different resource.
- Directory requests return `404` with the custom `404` page; there is no implicit directory index for any directory, including `/assets/`.
- The custom `404` page uses `Cache-Control: no-store`, retains its own ETag, and never becomes `304` through the conditional machinery.
- `405` returns the sorted `Allow: GET, HEAD` header with `Cache-Control: no-store` for unsupported methods.
- All tests use `httptest`, the race detector passes under concurrent reads, and the handler tree is composable without opening a listener.
- The implementation uses only the Go standard library, and the learner can explain each policy and trade-off without referring to implementation syntax.

## 19. Optional Extensions

1. Add a deterministic `Last-Modified` value derived from a stable build timestamp and a documented `If-Modified-Since` rule, while keeping the strong ETag primary and the existing `If-None-Match` rule unchanged.
2. Add a small `Content-Encoding: gzip` policy for known compressible text assets, computed once at startup, without changing the ETag or `Cache-Control` rules.
