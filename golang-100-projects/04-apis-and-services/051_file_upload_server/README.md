# Project 051 — File Upload Server

## 1. Project Name and Number

- Project 051 — File Upload Server, located in `051_file_upload_server`.

## 2. Project Idea

Build a streaming HTTP service that accepts exactly one uploaded file per request on `POST /uploads`, validates the body, the file type, and the destination path, and persists the accepted file inside an explicit upload root that is exclusively owned by this service. The service returns a generated identifier and the saved metadata. It enforces a streaming file-content maximum and an independent total-request maximum, distinguishes oversized from malformed uploads, recognizes PNG, JPEG, and PDF by content bytes only, never trusts a client filename as a destination, never overwrites an existing destination, and leaves no partial state on any failure path including client disconnect or context cancellation.

## 3. Why This Project Now?

- Project 050 taught authentication discipline that protects routes, and Project 046 is the `net/http` foundation this project extends.
- Project 041 contributes cancellation and timeout handling, which become essential once a long request body is in flight and a client may disconnect mid-upload.
- Project 048 is useful background for middleware and request context but is not required.
- The new step is to combine bounded multipart parsing, content-based file typing, safe path generation, atomic file persistence under a publication lock, and cleanup under cancellation, all at a single HTTP boundary.
- The next project then serves static assets whose integrity must be controlled, so this project must first establish where files come from and how they are stored.

## 4. Prerequisites

- Complete Projects 050, 046, and 041 before starting.
- Earlier projects may be useful review, but they are not required prerequisites.

- You should already be able to construct independently testable `net/http` handlers with `httptest`, distinguish `400`, `404`, `405`, `413`, `415`, and `201` semantically, bound request bodies with `MaxBytesReader`, propagate cancellation through request context, and clean up resources when a goroutine path is abandoned.
- You should also know how to inspect a few bytes at the start of a file to recognize its format.

## 5. What You Must Know Before Starting

- `POST /uploads` is the only success path. The request must contain exactly one multipart part in total, and that single part must be a file part whose `Content-Disposition` field name is exactly `file` and whose filename parameter is non-empty. Any other shape — an absent part, an empty part, an extra part, a non-file part, a wrong field name, a missing filename, a missing boundary, or a malformed multipart document — is a `400 Bad Request`. There is no policy that interprets an extra form part permissively.
- The request body is capped at exactly 11 mebibytes, which is 11,534,336 bytes. The streaming file content is capped at exactly 10 mebibytes, which is 10,485,760 bytes. The one-part contract permits only the required `Content-Disposition` part header and one optional short `Content-Type` part header; duplicate or additional part headers are rejected. Together with the 255-byte filename and standard boundary limit, the 1 mebibyte headroom covers the pinned multipart framing. A total-limit error may be discovered only after some bytes have reached the temporary file, so the guarantee is cleanup with no published destination or leftover temporary file, not that no temporary bytes were ever written. Exactly 10 mebibytes of file content is eligible only when the complete request also fits the total cap; the first file byte beyond that is `413`. The two limits are independent and tested independently.
- Type classification uses the bytes of the file content only. A supported PNG, JPEG, or PDF byte signature is accepted even when the client `Content-Type` header is missing, false, or unrelated, and even when the client filename extension is missing, false, or unrelated. Unsupported or insufficient bytes — including an empty body — return `415 Unsupported Media Type`. The service never rejects a supported byte stream merely because client claims disagree, and never accepts an unsupported byte stream merely because client claims claim support.
- The client-supplied filename is untrusted input. It is preserved only as response metadata, after a documented validation policy applied to the raw declared value with no helper normalization in between. The raw value must be valid UTF-8, between 1 and 255 UTF-8 bytes, and contain no ASCII control characters, no NUL byte, no forward slash, no backslash, no absolute or drive form, and no parent-traversal segment. A filename that violates the policy causes a `400 Bad Request` before any temporary file is created. On success, the value is emitted through the standard JSON encoder, which performs normal JSON escaping; the service does not invent manual escaping rules and does not replace an invalid declared value with an empty string.
- The destination basename is generated server-side. It is composed of an identifier drawn from an injected cryptographic source, formatted as exactly 32 lowercase hexadecimal characters, plus one of three type-derived extensions: `.png`, `.jpg`, or `.pdf`. Publication permits at most three generated identifiers per upload. Collisions on the first two identifiers retry; collision on the third, an identifier-source failure, or invalid generated output returns generic `500 Internal Server Error`, removes the temporary file, and leaves existing destinations unchanged.
- The publication step is the final existence check plus rename, both performed under a service-level publication mutex so that two concurrent handlers cannot race past an existence check and overwrite one another. Standard `os.Rename` alone is not a portable no-replace primitive, and the service does not claim that it defeats an external process racing to mutate the configured upload root. Concurrent overwrite by handlers inside this service is prevented; mutation of the upload root by any external actor is outside the baseline contract.
- The temporary file lives inside the same directory as the final destination, is created with a unique name, is opened with explicit flags that do not follow symlinks and do not permit other readers, is written through a streaming copy that counts bytes, is synced, and is closed successfully before publication. Under the publication mutex, the service checks that the final destination is absent and then renames the temporary file. This service-owned-root protocol prevents in-process overwrite; the rename call itself is not described as portable no-replace. Any publication failure removes the temporary file. Directory-level durability across power loss is outside the baseline contract.
- The service accepts files; it does not serve them, list them, decode them, scan them, or extract archives. There is no authentication, no database, and no cloud storage adapter in this project.
- Methods other than `POST` on `/uploads` are not supported, and an unsupported method returns the documented `405 Method Not Allowed` with the sorted `Allow` header.

## 6. Explanation of New Concepts

### Concepts

- A multipart form is a structured HTTP body composed of parts, each with headers and a payload.
- The `Content-Disposition` header identifies the field name and an optional filename parameter, and `Content-Type` may describe the part's media type.
- The request's own `Content-Type` indicates the boundary used to separate parts.
- This project accepts exactly one part, whose `Content-Disposition` contains `name="file"` and a non-empty `filename=` parameter.

- `http.MaxBytesReader` wraps a request body so that any read attempt beyond a configured cap fails.
- When applied to the request body as a whole, it provides the total-request limit.
- The cap must be larger than the maximum allowed file content because multipart framing contributes extra bytes; under the pinned one-part / header / filename policy used here, exactly 1 mebibyte of headroom is sufficient.
- The streaming file-content limit is enforced separately while the part's content is copied, and the two errors must be distinguishable: the first is `413 Payload Too Large` because the total body cap was crossed, while the second is `413 Payload Too Large` because the streaming file cap was crossed during copy.

- Magic bytes are short, fixed byte sequences at a known position inside a file.
- PNG files begin with a documented signature of eight bytes.
- JPEG files begin with a documented Start Of Image marker followed by an application segment.
- PDF files begin with the literal header that names the format and a version token.
- Reading only the leading bytes needed for recognition is enough to classify the file before committing to the rest of the stream.
- The classification is a property of the bytes, not of the client `Content-Type`, the client filename, or the client extension.

- Filename validation is a policy applied to the raw declared value before any helper normalization.
- The value is checked for UTF-8 validity, byte-length bounds, control characters, NUL, separators, absolute paths, drive forms, and traversal segments.
- A value that violates any rule is rejected and no temporary file is created.
- The validated value is then emitted through the standard JSON encoder, which performs normal JSON escaping.
- The service does not invent manual escaping rules and does not silently replace invalid declared values with empty strings.
- The valid declared value is therefore the exact string echoed back in the response, escaped by the standard JSON encoder.

- A cryptographic identifier source injects randomness into identifier generation.
- Tests inject a deterministic source that can be programmed to produce a fixed sequence or a colliding value.
- A production source reads from `crypto/rand`.
- The identifier is exactly 32 lowercase hexadecimal characters; a source failure or a format-invalid output becomes a generic internal error and prevents any file from being written.

- A publication mutex is a service-level lock that guards the final existence check plus rename pair.
- Under this lock, a handler verifies that no destination file exists, then performs the rename.
- While the lock is held, no other handler in this service can pass the existence check for the same identifier and overwrite the destination.
- The mutex does not extend to external actors or to a different process; claims about defending against processes external to this service are out of scope for the baseline contract.

- Atomic write uses a temporary file inside the same directory as the final destination.
- The temporary file is opened with explicit mode flags.
- Data is streamed in, synced, and closed.
- Only after a successful close does the publication mutex protect the final absence check and rename.
- An existing destination prevents the rename from being attempted.
- Any pre-publication or publication error removes the temporary file.
- The final destination is never created from an incomplete write.
- Directory-level durability across power loss is outside the baseline contract.

- Cleanup on disconnect and cancellation uses request context and explicit error paths.
- When the request context is cancelled mid-stream, the writer returns an error.
- The handler responds by removing the temporary file and returning no successful response, because the response status has not been committed yet.
- The same cleanup runs when the parser fails, when the magic-byte check fails, when the file size limit is crossed, when the write fails, when the close fails, and when the rename fails because the destination already exists.

## 7. Learning Objective

- By completion, you can combine bounded multipart parsing with content-based file typing, generate safe server-side filenames from a cryptographic source, write files atomically with rename under a publication lock, clean up partial state on every failure path including client disconnect, and explain why the two size limits are independent.
- You can also explain why client claims are never trusted for type, why raw filename validation happens before any helper normalization, why the JSON encoder is responsible for normal escaping, why a single quote of rename is not a portable no-replace primitive, and why a temporary file inside the destination directory is part of the atomic-write contract.

## 8. Functional Requirements

1. Expose `POST /uploads` as the only upload route. The request must contain exactly one multipart part in total. That part must be a file part whose `Content-Disposition` field name is exactly `file` and whose `filename` parameter is non-empty. Permit only that required part header and one optional short `Content-Type` part header; reject duplicate or additional part headers. A missing, empty, extra, non-file, wrongly named, or malformed part is `400 Bad Request` with a stable JSON error document.
2. Pin the total request body maximum to exactly 11,534,336 bytes, applied through `http.MaxBytesReader` over the request body. A request that exceeds this cap returns `413 Payload Too Large`. The total cap accounts for multipart framing under the pinned one-part / header / filename policy with exactly 1 mebibyte of headroom over the streaming file-content maximum.
3. Pin the streaming file-content maximum to exactly 10,485,760 bytes, enforced during copy. Files whose content size is exactly 10,485,760 bytes are eligible; the first byte beyond that during streaming is `413 Payload Too Large`. The two limits are independent and are tested independently: each has a boundary case, and the boundary between them is itself a testable contract.
4. Recognize PNG, JPEG, and PDF by inspecting magic bytes read from the part content. The classification decision uses the bytes only. A supported byte signature is accepted regardless of the client `Content-Type` header and regardless of the filename extension, including cases where those client claims are missing, false, or contradictory. A byte stream that does not match a supported signature, including an empty body, returns `415 Unsupported Media Type`.
5. Treat the client-supplied filename as metadata only. Validate the raw declared value with no helper normalization step. The raw value must be valid UTF-8, between 1 and 255 UTF-8 bytes, and free of ASCII control characters, NUL bytes, forward slashes, backslashes, absolute or drive forms, and parent-traversal segments. A raw value that violates the policy causes `400 Bad Request` before any temporary file is created.
6. Emit the valid declared filename through the standard JSON encoder on success. The JSON encoder performs normal escaping; the service does not invent manual escaping rules and does not replace an invalid declared value with an empty string. The metadata field in the success response is the validated raw string, escaped normally by the JSON encoder.
7. Generate the destination basename from an injected cryptographic identifier source. The identifier must be exactly 32 lowercase hexadecimal characters. The destination is the identifier joined with `.png`, `.jpg`, or `.pdf`. Permit at most three generated identifiers per upload: collisions on attempts one and two retry, while collision on attempt three returns generic `500 Internal Server Error`. A source failure or invalid output fails immediately. Every failure removes the temporary file and leaves existing destinations unchanged.
8. Perform the final existence check plus rename under a service-level publication mutex. The mutex prevents two concurrent handlers in this service from passing the existence check for the same destination and overwriting one another. The configured upload root is exclusively owned by this service within the baseline contract. An existing destination file at publication time blocks publication without overwriting.
9. Recognize that standard `os.Rename` alone is not a portable no-replace primitive. The publication step does not claim to defeat an external process racing to mutate the configured upload root. External mutation of the configured root is outside the baseline contract.
10. Stream the accepted content into a unique temporary file inside the explicit upload root, in the same directory as the final destination. Open the temporary file with explicit flags that do not follow symlinks and do not permit other readers. Apply the streaming limit during copy. Sync the temporary file before close. Close before rename.
11. Publish the temporary file only on complete success. Under the publication mutex, check that the destination is absent and only then rename. An existing destination prevents the rename call. This protocol prevents overwrite by handlers in the service-owned root; it does not claim the rename primitive itself is portable no-replace. Any failure removes the temporary file and leaves existing destinations unchanged.
12. Clean up partial state when the request context is cancelled, when the client disconnects, when the multipart parser fails, when the magic-byte check fails, when the file size limit is exceeded, when the write fails, when the sync fails, when the close fails, and when the rename or collision handling fails. No partial file remains in the upload root under any of these conditions.
13. Return `201 Created` on success with a JSON document containing the server-generated identifier, the saved name composed of the identifier and the type-derived extension, the detected type, the size in bytes, and the validated metadata filename as produced by the standard JSON encoder. Include `Content-Type: application/json; charset=utf-8` and a `Location` header whose value points to the canonical resource path for the new upload.
14. Reject unsupported methods on `/uploads` with `405 Method Not Allowed`, a sorted `Allow` header containing exactly `POST`, the documented JSON error envelope, and the exact JSON content type. Unknown paths return `404` with the same error envelope and content type. `HEAD` is not implied.
15. Place the upload root path under explicit configuration so tests can pass a temporary directory. The path is checked and prepared at startup. The service does not create files outside the configured root under any input.
16. The service does not serve uploaded files, list the upload root, decode images, scan for malware, extract archives, authenticate users, persist metadata to a database, or upload to cloud storage. Directory-level durability across power loss is outside the baseline contract.
17. Use a request-context-aware cleanup boundary. Every code path that opens a temporary file is paired with cleanup that runs when the handler returns, on cancel, on disconnect, on error, and after a rename that does not leave a live descriptor.

## 9. Inputs and Outputs

### Interface Contract

- The request is `POST /uploads` with `Content-Type: multipart/form-data; boundary=...`, exactly one part whose `Content-Disposition` contains `name="file"` and a non-empty filename parameter.
- The request body must be a valid multipart document and must not exceed the total body maximum.
- The part's media type may be any value; it is not trusted.
- The filename extension may be any value; it is not trusted.
- The part's claimed `Content-Type` may be any value; it is not trusted.

- Text-only success example: a request whose only part is a `file` part with a non-empty filename and a small PDF whose magic bytes identify PDF produces a `201 Created` response.
- The body contains the 32-character identifier, the saved name composed of that identifier and the type-derived `.pdf` extension, the detected type `pdf`, the byte count, and the validated metadata filename as escaped by the standard JSON encoder.
- The `Location` value is the canonical resource path for that identifier.

- Text-only boundary example: a request whose file part contains exactly 10,485,760 bytes of supported content is eligible if the complete request remains within 11,534,336 bytes.
- One additional file byte is rejected with `413 Payload Too Large`.
- A request whose total body exceeds 11,534,336 bytes also returns `413`; cleanup removes any temporary bytes written before detection and nothing is published.

- Text-only failure examples: a request with no `file` part, a wrong field name, an extra part, an unsupported or duplicate part header, a non-file part, a missing filename parameter, or malformed multipart produces `400 Bad Request`.
- A request whose total body exceeds the total maximum returns `413`; it publishes nothing and cleanup removes any temporary bytes already written.
- A file that exceeds the streaming maximum returns `413` during copy.
- Unsupported content bytes return `415 Unsupported Media Type`.
- Invalid raw filename metadata returns `400` before any temporary file is created.
- An unsupported method on `/uploads` returns `405 Method Not Allowed` with `Allow: POST`; an unknown path returns `404 Not Found`.

- The successful JSON document contains exactly the fields `id`, `name`, `type`, `size`, and `original_name`. `type` is one of `png`, `jpeg`, or `pdf`. `id` is exactly 32 lowercase hexadecimal characters. `name` is `id` joined with the type-derived extension. `size` is the exact number of bytes written to the destination. `original_name` is the validated raw filename as escaped by the standard JSON encoder.

## 10. Rules and Edge Cases

- The request must contain exactly one multipart part in total, and that part must be a file part named `file` with a non-empty filename parameter.
- Any other shape — a missing part, an empty part, a second part, a non-file part, a part whose `Content-Disposition` field name is not exactly `file`, a part whose `filename` parameter is empty or missing, or a malformed multipart document — is rejected with `400 Bad Request`.
- No interpretation of an extra part as a successful upload is permitted.

- The total body maximum of exactly 11,534,336 bytes wraps the body before multipart parsing, but the read that discovers excess may happen after earlier bytes were streamed to the temporary file.
- Detection returns `413`, cleanup removes the temporary file, and no destination is published.
- The separate 10,485,760-byte maximum is enforced while copying file content.
- Exactly that amount is eligible only when the total request also fits; the first file byte beyond is `413`.

- A file that begins with the PNG signature is accepted as PNG even when its declared `Content-Type` is something else and even when its filename extension is something else.
- The same is true for the JPEG signature and the PDF header.
- Anything else, including an empty body and a body whose first bytes are not a supported signature, is rejected with `415`.
- A request whose top-level `Content-Type` is not `multipart/form-data` is rejected before any part is read.

- The raw declared filename is validated with no helper normalization.
- Valid values are valid UTF-8, between 1 and 255 bytes, free of control characters and NUL, free of forward slashes, backslashes, absolute forms, drive forms, and parent-traversal segments.
- Invalid values produce `400 Bad Request` before any temporary file is created.
- The service does not silently rename an invalid value and does not replace it with an empty string; the metadata field is either the valid raw value as escaped by the JSON encoder, or the response is a `400`.
- The metadata field never contains an empty placeholder for an invalid declared value.

- The destination basename is generated server-side from the cryptographic identifier source plus the type-derived extension.
- The identifier is exactly 32 lowercase hexadecimal characters.
- Under the publication mutex, an existing destination is a collision.
- Collisions on attempts one and two cause a fresh identifier request; collision on attempt three returns generic `500 Internal Server Error`.
- A source failure or invalid output fails immediately.
- No failure modifies an existing destination or leaves a temporary file.

- Cleanup runs on every failure path.
- The temporary file is opened with a unique name inside the upload root, written through a streaming copy with a size counter, synced, and closed.
- If any step fails, the temporary file is removed.
- Under the publication mutex, an existing destination blocks the rename; any rename failure also removes the temporary file.
- The final destination is never created by an incomplete write.

- The upload root is configured.
- Tests pass a temporary directory created through the standard testing helpers.
- The service does not walk parent directories, does not follow symlinks into other roots, and does not create files outside the configured root.
- Filesystem permission errors are reported as internal errors and remove the temporary file.

## 11. Project Constraints

- Use only the Go standard library.
- Use `net/http`, `mime/multipart`, `io`, `path/filepath`, `crypto/rand`, `sync`, and the standard `net/http/httptest` package for tests.
- Do not use a web framework, an external multipart parser, an external MIME detector, an image decoder, an archive extractor, a malware scanner, an authentication library, a database, or a cloud storage SDK.
- The temporary-file path must use only standard-library functions.

- The exact 11,534,336-byte total maximum, the exact 10,485,760-byte streaming maximum, the exact 32-character lowercase-hex identifier format, and the three type-derived extensions are required learning contracts and are part of the document.
- Do not include other implementation code, function signatures, exact destination paths, exact magic byte sequences, or exact byte sequences for filenames in this guide.
- The guide states policies, not signatures.

## 12. Design Questions Before Coding

- What boundary owns the upload root path so tests can inject a temporary directory without affecting real filesystem state?
- Where does the multipart parser receive its body, and how does the total body maximum apply before the parser begins reading?
- Where does the streaming size counter live so the file maximum can be enforced during copy without buffering?
- What is the exact alphabet and length of the cryptographic identifier, and where is its source injectable in tests?
- How is the raw declared filename validated before any helper normalization, and where does the JSON encoder perform the response-side escaping?
- How is the temporary file given a unique name without relying on the random source that produces the final identifier?
- How is the temporary file opened with explicit flags that prevent symlink traversal and other readers?
- What guarantees that the rename never overwrites, and how is a collision detected and reported?
- What scope of the publication mutex covers the existence check plus rename, and what is explicitly outside that scope?
- What cleanup boundary removes the temporary file on cancel, disconnect, parser failure, type failure, size failure, write failure, close failure, and rename failure?
- What exact error envelope distinguishes `400`, `413`, `415`, `405`, and `404` without leaking internal causes?

## 13. Implementation Milestones

1. Record the route, method policy, content-type rules, error envelope, status mapping, the exact size maximums, the exact identifier format, and the exact field names of the success and error documents as testable acceptance criteria.
2. Establish the upload-root configuration boundary, the publication mutex, and the cryptographic identifier source boundary that tests can inject.
3. Enforce the total body cap of exactly 11,534,336 bytes through `MaxBytesReader`, parse exactly one `file` part with a non-empty filename parameter, and verify oversized, missing, wrong-name, extra, non-file, empty-filename, and malformed requests publish nothing and leave no temporary file, including when oversize detection occurs after streaming begins.
4. Recognize PNG, JPEG, and PDF from a partial magic-byte read without consuming the rest of the stream, and verify classification ignores client claims while accepting every supported signature.
5. Stream content with a size counter that enforces exactly 10,485,760 bytes during copy, and verify the boundary between the total-body maximum and the streaming maximum.
6. Validate the raw declared filename for UTF-8 validity, the 1-to-255-byte length band, control characters, NUL, slashes, backslashes, absolute forms, drive forms, and parent-traversal segments.
7. Create a uniquely named temporary file with explicit flags inside the upload root and destination directory, then complete the streaming write, sync, and close.
8. Generate identifiers from the cryptographic source as exactly 32 lowercase hexadecimal characters joined with the type-derived extension.
9. Publish under the service-level publication mutex with the final existence check plus rename, preserving the no-overwrite guarantee and removing the temporary file on publication failure.
10. Ensure cleanup on cancel, disconnect, parser failure, type failure, size failure, generation failure, write failure, sync failure, close failure, and rename or collision failure.
11. Return the success JSON through the standard JSON encoder with no manual escaping or empty-string replacement, plus `Location` and the exact content type; return the documented `405` with `Allow: POST` and `404` for unknown paths.
12. Finish deterministic, content-based, boundary-based, concurrency, and race-detector verification and review the contract for honest size-limit, publication-mutex, and cleanup evidence.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Submit a valid PNG, a valid JPEG, and a valid PDF as the only multipart part with a non-empty filename parameter. Verify `201`, exact JSON fields, exact content type, and a `Location` value. Verify that `id` is exactly 32 lowercase hex characters, that `name` is the identifier joined with the appropriate type-derived extension, and that `original_name` is the validated raw filename escaped by the standard JSON encoder.
- Submit a file with the correct magic bytes but a misleading `Content-Type` and a misleading extension. Verify the request is accepted because classification is content-driven and ignores client claims.
- Submit a file whose bytes are not a recognized format with a misleading extension. Verify `415` and no file written.
- Submit a file that contains exactly 10,485,760 bytes of valid content. Verify `201` is reachable. Submit the same file with exactly one additional byte and verify `413 Payload Too Large` during streaming, with no destination file present.
- Submit a request whose total body exceeds 11,534,336 bytes after streaming has begun. Verify `413 Payload Too Large`, no published destination, and no temporary file remaining in the upload root.
- Submit a request with a missing `file` part, a wrong field name, a non-file part, an extra part, a missing filename parameter, an empty filename parameter, and a malformed multipart document. Verify each case maps to `400` and no file is written.
- Submit a request whose body uses `Content-Type: application/json` instead of `multipart/form-data`. Verify `400` and no file written.
- Simulate a client disconnect by cancelling the request context after the headers but before the body completes, and during the streaming copy. Verify no partial file remains in the upload root and no successful response is committed.
- Submit raw declared filenames that are not valid UTF-8, are empty, exceed 255 bytes, contain control characters, contain a NUL byte, contain a forward slash, contain a backslash, take an absolute form, take a drive form, and contain a parent-traversal segment. Verify each case returns `400 Bad Request` before any temporary file is created and that no destination file is produced.
- Submit a valid raw declared filename and verify the `original_name` field in the success response is the validated raw value escaped by the standard JSON encoder, with no manual escaping and no empty-string replacement.
- Inject a cryptographic identifier source that returns a fixed 32-character value. Verify the destination basename includes that value plus the validated extension.
- Inject a source whose output is invalid or fails. Verify generic `500 Internal Server Error` with no destination file and no temporary file.
- Inject existing destinations for generated attempts one and two, then a free third identifier, and verify successful publication under the third name. Inject collisions for all three attempts and verify generic `500`, no overwrite, and no temporary file.
- Run concurrent uploads against the same upload root with distinct injected identifier sources. Verify each file is written exactly once, no destination is shared, no partial file remains, the publication mutex prevents overwrites between handlers, and the race detector reports no issue.
- Verify unsupported methods on `/uploads` return `405` with sorted `Allow: POST`. Verify an unknown path returns `404`. Verify `HEAD` is not implied.
- Verify the upload root is the only directory touched by a successful upload and by every failure path. Verify no symlink target outside the root is created or followed.

## 15. Common Mistakes to Watch For

- Trusting the client `Content-Type` or the filename extension for type detection accepts a renamed malicious file.
- Rejecting a supported byte stream because client claims disagree defeats the content-based type policy.
- Reading magic bytes only after the whole file is buffered wastes memory and removes the streaming maximum's effect.
- Setting one limit that conflates total body and file content makes the boundary between the two `413` cases meaningless.

- Treating the client filename as already safe because it has been "passed through" a helper is a security bug.
- Normalizing or cleaning the raw value before validation lets an attacker construct a name whose normalized form passes.
- Replacing an invalid raw value with an empty placeholder string hides the issue and makes the metadata field lie about the input.
- Inventing manual escaping rules outside the standard JSON encoder risks breaking the response shape.

- Using the client filename in the destination creates path traversal, platform-separator, and NUL-byte attacks.
- Replacing it with a silent server name without policy is still hiding the issue.
- Treating the client filename as already safe because the policy "checked" it without an explicit raw-validation step is a security bug.

- Overwriting an existing destination destroys data and breaks the uniqueness contract.
- Using a non-atomic write leaves a half-written file visible to readers.
- Failing to remove the temporary file on cancel or disconnect leaves partial files in the upload root.
- Skipping the publication mutex lets two handlers pass the existence check for the same identifier and overwrite each other.
- Claiming `os.Rename` alone provides a portable no-replace guarantee is false; external mutation of the upload root is outside the baseline contract and must not be claimed.

- Skipping cleanup on the success path after rename can leave the temporary file if the rename succeeded for a different name or if a future operation reuses the temporary descriptor.
- Holding the temporary file descriptor open after rename may leak file handles and break the atomic-write contract.

- Reading the request body without a total maximum permits unbounded memory growth.
- Using a total maximum that only counts the file content misses framing overhead and lets oversize total requests through.
- Using a total maximum that only counts framing and ignores the file content prematurely rejects valid uploads.
- Mixing the two errors into a single code path hides the difference between total-body and streaming failures.

- Returning a successful response after a partial write or after cancel is dishonest.
- Committing headers before the body completes leaves the client unable to detect a truncation.
- Including internal causes in the response leaks filesystem and parser detail.

## 16. Topics and References for Study

- Study the official documentation for `net/http`, especially `MaxBytesReader`, `Request.MultipartReader`, `ResponseWriter`, and the standard patterns for limiting request bodies.
- Study `mime/multipart` for part iteration, `Content-Disposition`, framing rules, and partial reads.
- Study `io.Copy`, `io.LimitReader`, and `path/filepath` for safe path joining and bounded streaming copies.
- Study `crypto/rand`, `encoding/hex`, and the standard-library practices for generating fixed-length lowercase-hex identifier strings.
- Study `sync.Mutex` for the publication mutex and in-memory store synchronization.
- Study the standard `net/http/httptest` package for in-memory handler tests and context cancellation.
- Review the relevant sections of RFC 7578 for multipart form returns, RFC 7231 for HTTP method semantics, and the documented byte signatures for PNG, JPEG, and PDF.

## 17. Self-Assessment Questions

1. Why are the 11,534,336-byte total maximum and the 10,485,760-byte streaming maximum enforced independently, and why is 1 mebibyte the right headroom under the pinned one-part / header / filename policy?
2. Why is magic-byte classification safer than trusting client `Content-Type` or extension, and why must a supported byte stream never be rejected on conflicting client claims?
3. Why is the client filename raw-validated before any helper normalization, and what breaks if normalization happens first?
4. Why does the standard JSON encoder perform the response-side escaping, and why must the service not invent manual escaping rules or replace an invalid declared value with an empty string?
5. Why is the destination basename always the cryptographic identifier plus the type-derived extension, and what classification rules pick `.png`, `.jpg`, or `.pdf`?
6. Why must the publication mutex cover the final existence check plus rename, what no-overwrite guarantee does that provide, and what remains outside its scope?
7. Why must cleanup run on cancel, disconnect, and every failure path, not only on success?
8. Why is a `415` response different from a `400` response for the same upload shape?
9. Why does this project not serve the uploaded files, and what would change if it did?
10. Why is the cryptographic identifier source injected, what deterministic control does that give tests, and why must source failures, invalid output, and collision exhaustion return generic `500` without leaving a temporary file?

## 18. Definition of Completion

- [ ] `POST /uploads` accepts exactly one multipart part, which must be a file part named `file` with a non-empty filename parameter, and returns `201` with the documented JSON document and `Location`.
- [ ] The total body maximum of exactly 11,534,336 bytes and the streaming maximum of exactly 10,485,760 bytes are enforced independently; the streaming boundary at exactly 10,485,760 bytes being eligible and the first byte beyond being `413` is verified.
- [ ] PNG, JPEG, and PDF are accepted by magic-byte detection and rejected otherwise, regardless of client `Content-Type` or extension claims.
- [ ] The raw declared filename is validated for UTF-8 validity, 1-to-255-byte length, control characters, NUL, slashes, backslashes, absolute forms, drive forms, and parent-traversal segments; invalid values return `400` before any temporary file is created; valid values are echoed through the standard JSON encoder with no manual escaping and no empty-string replacement.
- [ ] The destination is generated from the cryptographic source as exactly 32 lowercase hexadecimal characters joined with `.png`, `.jpg`, or `.pdf`, is verified under a service-level publication mutex, and never overwrites an existing destination; standard `os.Rename` is not claimed to be a portable no-replace primitive; external mutation of the configured upload root is explicitly outside the baseline contract.
- [ ] The temporary file is created in the same directory as the destination, opened with explicit flags, written under the streaming limit, synced, closed, then renamed on complete success; on any failure, the temporary file is removed.
- [ ] Cleanup runs on cancel, disconnect, parser failure, type failure, size failure, generation failure, write failure, sync failure, close failure, rename failure, and collision.
- [ ] Identifier-source failures, generation failures, and collision exhaustion return generic `500 Internal Server Error` with no destination file and no temporary file.
- [ ] Unsupported methods, unknown paths, missing or wrong field names, malformed multipart documents, and invalid raw filenames map to the documented statuses without leaking internal causes.
- [ ] The race detector passes under concurrent uploads against the same upload root with distinct identifier sources, and the publication mutex prevents overwrites between handlers.
- [ ] The service uses only the Go standard library, the upload root is configurable and prepared at startup, and no code path creates files outside it.
- [ ] The implementation does not serve, list, decode, scan, authenticate, persist to a database, or upload to cloud storage, and the learner can explain each boundary decision including the limits of the publication mutex.

## 19. Optional Extensions

- Add a deterministic batch-upload endpoint that accepts a small fixed number of parts in one request, reuses the same atomic-write pattern and publication mutex, and returns one summary document.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 050 — JWT Auth Server](../../04-apis-and-services/050_jwt_auth_server/README.md#20-prerequisite-based-documentation-guide), [Project 046 — Basic HTTP Server](../../04-apis-and-services/046_basic_http_server/README.md#20-prerequisite-based-documentation-guide), [Project 041 — Context Timeout Example](../../03-concurrency/041_context_timeout_example/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`mime/multipart`](https://pkg.go.dev/mime/multipart).
- **Standards and concept references:** [RFC 7578: multipart/form-data](https://www.rfc-editor.org/rfc/rfc7578.html).

### Project-specific learning focus

- **Learn now:** request and per-file byte caps, streaming multipart parts, filename distrust, signature validation, random identifiers, temporary files, atomic publication, and cleanup.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
