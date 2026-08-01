# Project 050 — JWT Auth Server

## 1. Project Name and Number

Project 050 — JWT Auth Server, located in `050_jwt_auth_server`.

## 2. Project Idea

Build a small authentication service with a JSON login endpoint and one protected endpoint. A fixed credential-store boundary verifies one test user's password with bcrypt, a short-lived JWT is signed with HS256, and middleware authenticates a Bearer token before placing the verified subject in request context. The service authenticates only; it does not implement roles, authorization policy, registration, persistence, refresh, or revocation.

## 3. Why This Project Now?

Project 049 provides a stable JSON response and public-error boundary, while Project 048 provides route and middleware discipline. This project combines them at a security boundary where ambiguous parsing, leaked errors, weak secrets, and algorithm confusion have serious consequences. Project 046 remains the HTTP foundation. The result prepares the learner for later services that need to distinguish authentication from authorization without pretending that a signed token solves every security problem.

## 4. Prerequisites

Complete Projects 049 and 046 before starting. Earlier projects may be useful review, but they are not required prerequisites.

You should already understand strict JSON boundaries, response envelopes, `401` and `WWW-Authenticate`, request middleware and typed context values, injected clocks and IDs, concurrent `httptest` tests, and the difference between authentication and authorization.

## 5. What You Must Know Before Starting

- Passwords are not compared or stored as plaintext. A password hash is a one-way verifier representation, and bcrypt comparison must be performed by the maintained `golang.org/x/crypto/bcrypt` package.
- A JWT is a signed claims representation, not encryption. Anyone holding the token can generally read its payload, so passwords and secrets must never be placed in claims.
- The algorithm named in an incoming token is untrusted input. Verification must allowlist HS256 before accepting the signature and must reject `none` and every other algorithm.
- Registered claims have time and identity semantics. Expiration, not-before, issuer, audience, subject, issued-at, and token ID must be validated as claims, not merely decoded as arbitrary JSON.
- `401 Unauthorized` means the request lacks acceptable authentication and must include an authentication challenge. This project uses the exact Bearer challenge specified below. Authorization decisions such as roles and permissions are a separate concern.
- An `Authorization` header can have multiple field values or malformed whitespace. Parsing exactly one accepted form is safer than taking the first loosely parsed value.
- An environment variable can be required configuration, but its presence alone does not prove cryptographic strength. The decoded secret must contain at least 32 random bytes and must never be logged.
- Injected time and ID sources make expiry and token-identity tests deterministic. They do not make real deployment time or randomness optional.
- TLS is required for real deployment because Bearer tokens and passwords must be protected in transit. The learning server and its `httptest` cases may use in-memory HTTP because TLS termination is outside this project's boundary.

## 6. Explanation of New Concepts

Bcrypt is a password-hashing scheme designed to make guessing expensive. The server stores a bcrypt hash and asks the package to compare a presented password with that hash. It never reconstructs or decrypts a password. For an unknown username, the credential boundary should still perform comparison work against a fixed dummy bcrypt hash so ordinary wrong-user and wrong-password paths do not expose account existence through an obvious shortcut. Timing equality is not claimed as absolute over a network.

A JWT compact token has a header, claims payload, and signature. HS256 uses one shared secret for signing and verification. The token proves that a holder of the secret signed the claims; it does not prove that the token is encrypted, fresh beyond its time claims, or unrevoked.

This project pins one issuer string, one audience string, one fixed subject for the test user, a normalized issued-at time, an equal not-before time, an expiration exactly five minutes later, and a unique token ID. The parser must verify all required claims and their types as well as the signature. Membership in an audience list is not enough if extra audiences are permitted by accident; this contract requires the exact configured audience and no unexpected audience values.

A Bearer authentication scheme places the compact token after the exact scheme token and one separator space. The middleware rejects missing, duplicated, or malformed credentials before attempting verification. On every protected authentication failure it returns the same public `401` response and exact challenge, regardless of whether the token was absent, malformed, expired, incorrectly signed, or had a wrong claim.

A typed context value carries the authenticated subject from middleware to the protected endpoint for one request. It is evidence that authentication completed; it is not a role system or a permission decision. The protected endpoint proves availability by returning the subject in its success data.

Secret loading is a startup boundary. The required environment value is decoded and validated once before the listener is made available. Tests inject the resulting secret directly whenever possible, avoiding process-global environment mutation and avoiding accidental logging of configuration.

## 7. Learning Objective

By completion, you can explain password-hash comparison, JWT signing versus encryption, algorithm allowlisting, registered-claim validation, Bearer parsing, generic authentication errors, secret configuration, and TLS deployment boundaries. You can build deterministic login and protected-route tests that reject forged, expired, malformed, and algorithm-confused tokens while remaining safe under concurrent verification.

## 8. Functional Requirements

1. Expose `POST /login` and `GET /protected`. Unknown paths return `404`. A known path with an unsupported method returns `405` with a sorted `Allow` header: `POST` for `/login` and `GET` for `/protected`. `HEAD` is not implied by `GET` and is rejected explicitly with no body.
2. The credential-store boundary contains exactly one fixed test user with username `learner`, subject/user ID `user-001`, and a bcrypt password hash. It has no registration operation and no database.
3. Login requests must have media type `application/json` with valid optional media-type parameters and a body no larger than exactly 16,384 bytes. Missing or unrelated content types return `415`; malformed, incomplete, unknown-field, wrong-type, or trailing-second-value JSON returns `400` using the public error envelope from Project 049.
4. The login JSON object contains exactly string fields `username` and `password`. No other fields are accepted. Password text is never trimmed, transformed, stored, or logged.
5. A correct fixture credential returns `200 OK` with a JSON success envelope. Its data object contains exactly `access_token`, `token_type`, and `expires_in`; `token_type` is `Bearer` and `expires_in` is the integer `300`.
6. A wrong username and a wrong password return the same `401` status, the same public JSON code `unauthenticated`, the same public message `authentication required`, and the same `WWW-Authenticate: Bearer` header. Neither response reveals which credential component failed.
7. Use `golang.org/x/crypto/bcrypt` for password-hash comparison. Do not compare plaintext to plaintext or implement password hashing, salting, or cryptographic primitives manually.
8. Use `github.com/golang-jwt/jwt/v5` for token creation and parsing. Do not hand-roll JWT serialization, signature verification, claim validation, or algorithm selection.
9. Sign tokens with HS256 only. The signing secret is loaded from required startup configuration and is never defaulted, hardcoded, printed, or logged.
10. The required environment variable is named `JWT_AUTH_SECRET`. Its value is standard-base64 text that decodes to at least 32 random bytes. Missing, malformed, or short decoded values prevent startup before the server accepts requests.
11. Every issued token contains required claims with exact meanings: issuer `project-050`, one audience value `project-050-protected`, subject `user-001`, issued-at `iat`, not-before `nbf`, expiration `exp`, and a unique non-empty token ID `jti`.
12. Token lifetime is exactly five minutes, or 300 whole seconds. `nbf` equals normalized `iat`, and `exp` equals `iat` plus 300 seconds. The injected clock supplies issuance time and parser time; required tests use zero clock leeway.
13. The token ID comes from an injected ID or random source, consists of exactly 32 lowercase hexadecimal ASCII characters, and is unique for each successful login under that source's contract. A source failure or invalid ID prevents token issuance and returns a generic `500` without exposing the cause.
14. Verification explicitly allowlists HS256 and validates the signature, required expiration, issuer, exact audience set, exact subject, issued-at, not-before, and a token ID of exactly 32 lowercase hexadecimal characters. It rejects `none`, wrong algorithms, bad signatures, missing claims, malformed claim types, expired tokens, future not-before values, and future issued-at values.
15. Protected middleware accepts exactly one `Authorization` header value in the form `Bearer` followed by one ASCII space and one compact token. It rejects missing, multiple, comma-separated, extra-whitespace, wrong-case-scheme, empty, or otherwise malformed credentials with generic `401` and `WWW-Authenticate: Bearer`.
16. Successful protected verification places subject `user-001` in request context under a private typed key. `GET /protected` reads that authenticated subject and returns `200` with a JSON success envelope whose data object contains exactly `subject` with value `user-001`.
17. Authentication failures never reveal whether a user exists, which claim failed, whether a signature was close, or why parsing failed. Login and protected errors use stable public messages from the formatter contract.
18. Never log passwords, password hashes, JWT compact strings, authorization headers, signing secrets, or raw request bodies. Internal failure logs may contain a safe category and request ID but must not contain secret material.
19. The service performs authentication only. Roles, authorization rules, revocation, refresh tokens, registration, password reset, persistent users, and TLS termination are outside the required scope.

## 9. Inputs and Outputs

The login path accepts one JSON object with exactly two string fields: `username` and `password`. The fixed test username is `learner`; the corresponding test password is supplied by the test fixture and exists only to exercise the stored bcrypt hash. Runtime credential state contains the hash, not that plaintext test value.

A successful login response is status `200`, content type exactly `application/json; charset=utf-8`, and the Project 049 success envelope. Its data object has an opaque compact JWT in `access_token`, exact string `Bearer` in `token_type`, and integer `300` in `expires_in`. It contains no password, hash, secret, or extra credential detail.

Text-only login example: a valid JSON login using the fixed fixture credentials returns one Bearer token whose issuer is `project-050`, whose audience is `project-050-protected`, whose subject is `user-001`, and whose expiration is exactly 300 seconds after issued-at. The token text itself is not printed in this guide or in logs.

A wrong-user and wrong-password response is the same public response in both cases: status `401`, JSON error code `unauthenticated`, message `authentication required`, content type `application/json; charset=utf-8`, and `WWW-Authenticate: Bearer`.

The protected path accepts a `GET` with exactly one valid Bearer credential. Its success data object has exactly one field, `subject`, containing `user-001`. Missing or invalid credentials use the same generic protected `401` response regardless of failure reason.

Malformed login documents use the Project 049 error code `invalid_request` and status `400`; over-limit bodies use `payload_too_large` and `413`; wrong media types use `unsupported_media_type` and `415`; internal configuration or dependency failures use `internal_error` and `500`. All body-bearing responses use the formatter's stable envelope and content type.

## 10. Rules and Edge Cases

The login decoder accepts one complete JSON value followed only by whitespace. Empty input, malformed syntax, a second JSON value, unknown field, missing field, null field, or non-string field is a `400` structural error. A syntactically valid but empty or impossible credential is treated as a generic credential failure under the same `401` contract rather than revealing account-validation detail. Password characters, including surrounding whitespace, are significant.

The 16,384-byte cap counts the encoded request body, including whitespace. A body that attempts to exceed the cap is rejected before state or token mutation. No partial JSON value can trigger a credential lookup or token issuance. The service does not accept form data, query-string credentials, credentials in a URL, or a second JSON document.

Credential comparison uses bcrypt for both known and unknown usernames, with a fixed non-matching comparison path for an unknown user. The test suite checks equal public behavior, not wall-clock equality. A malformed configured bcrypt hash or credential-store failure is a server configuration/dependency failure and maps to generic `500`, never to a misleading client `401`.

The environment secret is loaded once at startup. Base64 decoding is required, and the decoded byte sequence must contain at least 32 random bytes. An empty variable, fallback value, hardcoded development key, short value, or invalid encoding prevents startup. Startup can verify encoding and decoded length but cannot prove entropy; deployment configuration must obtain the bytes from a cryptographically secure random source. Tests inject a deterministic secret directly and do not require process-global environment changes unless they are specifically testing the loader boundary.

Issued claim times are normalized to whole UTC seconds. The token is valid only when current time is at least `nbf` and strictly before `exp`; a token at the expiration instant is expired. An issued-at value in the future is rejected. There is no implicit clock leeway in the required contract. A token ID is exactly 32 lowercase hexadecimal ASCII characters, and the injected production source must provide a fresh value for every successful login.

The parser validates the signing method before trusting claims. A token declaring `none`, an HMAC algorithm other than HS256, an asymmetric algorithm, or an unknown method is rejected even if its text has a plausible three-part shape. A signature is checked with the configured secret, and all required claims are checked for presence, exact type, expected identity, and expected time semantics.

The audience must be exactly the one configured value `project-050-protected`; missing, empty, malformed, or extra audience values are rejected. Issuer must equal `project-050`; subject must equal `user-001`; `iat`, `nbf`, and `exp` must be integral numeric dates; and `jti` must be exactly 32 lowercase hexadecimal ASCII characters. No claim is accepted merely because it can be decoded.

Authorization parsing is strict. There must be one and only one `Authorization` field value. The scheme spelling must be exactly `Bearer`, followed by exactly one ASCII space and a non-empty compact token with no additional whitespace or comma. Any missing, duplicated, malformed, or invalid token produces the same protected `401` response and exact Bearer challenge. The token is never placed in logs or context.

`HEAD` is rejected on both known paths with the relevant `Allow` value and zero body bytes. Unknown paths are `404`. The real service must run behind TLS or a trusted TLS terminator; in-memory `httptest` HTTP is a test transport, not a production security claim.

Concurrent verification may share immutable configuration but must not share mutable per-request claims, buffers, parser state, or context values. Each request receives its own authenticated subject. Authentication success does not grant any role or permission beyond access to this one protected demonstration route.

## 11. Project Constraints

Use the Go standard library plus exactly these two external dependencies: `golang.org/x/crypto/bcrypt` and `github.com/golang-jwt/jwt/v5`. Use maintained, pinned module versions selected by the learner's module configuration. Do not add a framework, another JWT library, another password library, an ORM, a database, or a password/crypto implementation.

Do not hand-roll cryptography, JWT parsing, signing, claim validation, Bearer parsing shortcuts, or password hashing. Do not put secrets, passwords, hashes, tokens, or implementation code in this README. Do not log sensitive headers or bodies. Tests use `httptest`, injected clock, injected secret, injected credential store, and injected ID/random source; they do not rely on real sleeps or external services.

TLS is required for real deployment and is intentionally terminated outside this learning server and its tests. Roles, authorization, registration, persistence, refresh, revocation, key rotation, and account recovery are not part of the baseline.

## 12. Design Questions Before Coding

- What boundary owns the fixed credential record, and how can unknown usernames take a comparable bcrypt path without exposing account existence?
- Which exact test user username and subject are stable, and where is the test-only plaintext kept so runtime state stores only its hash?
- What representation will the required environment variable use, and how will startup reject missing, malformed, or short decoded secrets before listening?
- Which claim values are constants, which come from the injected clock, and which come from the injected ID/random source?
- How will whole-second normalization make the five-minute lifetime and expiration boundary deterministic?
- How will the parser prove HS256 is allowlisted before signature verification and reject every alternative algorithm?
- How will exact audience membership differ from merely finding the expected value among unexpected values?
- Where will the authenticated subject live in context, and how will the protected endpoint prove it without exposing token contents?
- How will exactly one Authorization field value and exactly one separator space be recognized without accepting a comma list or extra whitespace?
- Which boundary owns public errors, `WWW-Authenticate`, request IDs, and internal logging so no helper writes a second response?
- How will concurrent verification avoid mutable shared parser or claims state?

## 13. Implementation Milestones

1. Record the two routes, method policies, status mappings, JSON envelopes, body cap, and authentication-header grammar.
2. Establish the fixed credential-store boundary and verify that only a bcrypt hash is retained at runtime.
3. Establish startup secret loading and validation, including required variable name, base64 representation, minimum decoded length, and no fallback.
4. Add injected clock and token-ID/random boundaries with exact five-minute claim rules.
5. Implement successful login behavior through the Project 049 response boundary.
6. Implement generic wrong-credential behavior and the unknown-user bcrypt comparison path.
7. Define token creation with the exact HS256 header/method policy and required claims.
8. Define strict token parsing and validation for algorithm, signature, identity claims, time claims, and token ID.
9. Add strict Bearer middleware, typed subject context, protected-route proof, and exact challenge responses.
10. Add deterministic malformed, security, concurrency, race-detector, and no-secret-leakage tests, then review the TLS deployment boundary.

## 14. Verification Cases the Learner Must Write

- Log in with the fixed test credentials and verify status, content type, success envelope shape, Bearer token type, five-minute `expires_in`, and absence of password or secret fields.
- Log in with the wrong password and wrong username. Verify byte-equivalent public error envelopes, identical status and challenge headers, and no account-enumeration detail. Do not assert network timing equality.
- Submit a missing media type, an unrelated media type, a malformed media type, malformed JSON, empty input, unknown fields, wrong field types, missing fields, a second JSON value, and a body beyond 16,384 bytes. Verify exact `415`, `400`, or `413` behavior and no token issuance. Submit `application/json` with a valid media-type parameter and verify it passes the media-type check.
- Inspect an issued token with the maintained JWT library and verify HS256, issuer `project-050`, exact audience `project-050-protected`, subject `user-001`, integral `iat`, `nbf` equal to `iat`, `exp` exactly 300 seconds later, and a unique `jti` with exactly 32 lowercase hexadecimal characters.
- Issue two successful tokens with distinct injected IDs and verify both IDs have the exact 32-character lowercase hexadecimal format and differ. Make the source fail or return an invalidly formatted value and verify generic `500` with no token or internal detail.
- Verify the token immediately before expiration, at the expiration instant, and after expiration. Confirm the exact boundary treats the expiration instant and later as expired and uses no hidden sleep or leeway.
- Verify a future `nbf` and future `iat` are rejected, while a token at its valid not-before time succeeds when all other claims are correct.
- Verify a wrong signature, altered payload, missing signature, `none` algorithm, HS384 or HS512 algorithm, asymmetric algorithm, and unknown algorithm are all rejected by the explicit HS256 allowlist.
- Verify missing, empty, malformed, wrong-type, or wrong-value issuer, audience, subject, issued-at, not-before, expiration, and token-ID claims are rejected. Verify an incorrectly formatted token ID and extra audience values are not accepted.
- Send protected requests with no Authorization header, multiple values, a comma list, wrong scheme, wrong capitalization, extra spaces, an empty token, a two-part token, and a token with trailing whitespace. Verify the same generic `401`, exact JSON error, and `WWW-Authenticate: Bearer` each time.
- Send a valid token to `/protected` and verify status, content type, success envelope, and exact authenticated subject `user-001`. Verify the subject is available through typed request context rather than a global.
- Verify wrong methods on `/login` and `/protected`, explicit `HEAD`, and unknown paths return the documented status, sorted `Allow` value, body policy, and stable public error.
- Test startup secret loading with missing, invalid-base64, and fewer-than-32-byte values. Verify startup fails before serving and no secret appears in an error or log.
- Capture logs for invalid credentials, malformed tokens, and internal verification failures. Verify passwords, hashes, compact tokens, Authorization values, and secrets never appear.
- Verify concurrent protected requests using one valid token and concurrent logins using injected sources. Run verification under the race detector and confirm no cross-request claim or context contamination.
- Verify all tests use injected time, secret, store, and ID/random sources; no test sleeps or contacts an external service.

## 15. Common Mistakes to Watch For

Comparing plaintext passwords, storing plaintext beside the hash, generating a hash on every request, or using a fast general hash for passwords defeats bcrypt's purpose. Returning “unknown user” and “wrong password” through different messages or statuses enables account enumeration. Logging a password, hash, token, Authorization header, or secret turns diagnostics into a credential leak.

Trusting the token's declared algorithm, accepting `none`, using an unchecked parse path, or selecting a verification key from untrusted token data creates algorithm-confusion vulnerabilities. Checking only the signature while ignoring issuer, audience, subject, time, or required-claim presence accepts a token outside its intended use.

Using a default or hardcoded HMAC secret, accepting a short human phrase as sufficient entropy, loading the secret after serving begins, or printing it in startup logs weakens the entire system. Base64 text length is not the same as decoded random-byte length; validate the decoded bytes.

Parsing the first Authorization value, splitting on arbitrary whitespace, accepting comma-separated values, or treating a lowercase scheme as equivalent when the contract pins exact `Bearer` creates ambiguous authentication behavior. Returning a different protected error for expiration, signature failure, or missing credentials leaks verification detail.

Using wall-clock calls directly in issuance and parsing tests makes expiration cases flaky. Adding leeway without documenting it changes the exact lifetime boundary. Sharing mutable claims, parser state, buffers, or context maps across concurrent requests creates races and possible subject confusion.

Treating a JWT as encrypted, putting a password or secret in claims, assuming a valid token is revocable, or deploying bearer authentication over plain HTTP confuses separate security properties. TLS termination outside this learning server must still be configured and trusted in a real deployment.

## 16. Topics and References for Study

Study the official documentation for `golang.org/x/crypto/bcrypt`, especially password comparison and its input limits. Study `github.com/golang-jwt/jwt/v5` for registered claims, parser validation options, signing-method allowlists, NumericDate behavior, and time injection. Read RFC 7519 for JWT claims, RFC 6750 for Bearer authentication and challenges, and RFC 8725 for JWT best practices and algorithm validation. Review OWASP password-storage and authentication guidance, Go `encoding/json`, `net/http`, `context`, `httptest`, `crypto/rand`, and Project 049's response boundary.

## 17. Self-Assessment Questions

1. Why is bcrypt comparison different from decrypting or hashing a password once?
2. Why should an unknown username still take a bcrypt comparison path?
3. What does an HS256 JWT prove, and what does it not prove?
4. Why must the parser allowlist the algorithm before trusting a signature result?
5. Why are issuer, audience, subject, and token ID security checks rather than decorative fields?
6. What exact time condition makes a token expired, and why does injected time matter?
7. Why is decoded secret length more meaningful than environment-string length?
8. Why must multiple or malformed Authorization values be rejected instead of choosing one?
9. Why does a protected endpoint still need authorization concepts that this project deliberately omits?
10. Why is TLS outside the server's code boundary but still mandatory for real deployment?

## 18. Definition of Completion

- `/login` and `/protected` implement the exact route, method, status, header, JSON, and body policies, including explicit `HEAD` behavior.
- The only runtime credential record is the fixed test user and bcrypt hash; no registration or database exists.
- Login enforces strict JSON, exact content type, the 16,384-byte cap, generic wrong-credential behavior, and no sensitive output.
- Only HS256 is accepted, the required base64 environment secret is validated before startup, and tests can inject a secret directly without process-global state.
- Issued tokens have exact issuer, audience, subject, whole-second times, five-minute lifetime, and unique token ID. Injected clock and ID/random source make tests deterministic.
- Verification validates algorithm, signature, every required claim, exact audience and subject, expiration, not-before, and issued-at semantics, rejecting all listed malformed and forged variants.
- Protected middleware accepts exactly one `Bearer` form, returns generic `401` plus `WWW-Authenticate: Bearer` for every authentication failure, and exposes only the verified subject through typed context.
- No internal cause leaks to clients. Logs contain only safe categories and request identifiers for authentication failures and never contain a password, hash, compact token, Authorization value, signing secret, raw request body, stack trace, or other sensitive material.
- Concurrent login and verification pass under the race detector without real sleeps or external services.
- TLS is explicitly required at the real deployment boundary, and roles, authorization, revocation, refresh, registration, and persistence remain out of scope.
- The implementation uses only the two named external dependencies and the standard library, and the learner can explain every security decision.

## 19. Optional Extensions

1. Add HMAC key rotation with an explicit allowlist of configured keys and a documented key identifier, without accepting an algorithm outside HS256.
2. Add login-attempt throttling with deterministic injected time and per-identity limits, without changing generic authentication errors or adding revocation and refresh semantics.
