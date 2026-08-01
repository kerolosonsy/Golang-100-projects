# Project 059 — Session Cookie Auth

## 1. Project Name and Number

Project 059 — `session_cookie_auth`. Folder: `04-apis-and-services/059_session_cookie_auth/`. README only; the learner writes all source and tests.

## 2. Project Idea

Build server-side opaque sessions on top of `net/http` for a small JSON API. The service exposes exactly four endpoints: `POST /login`, `GET /me`, `POST /action`, `POST /logout`. The single fixed test user is username `learner`, ID `user-001`. Only a bcrypt hash of the user's password is stored. The credential comparison always runs bcrypt: for known users it compares against the stored hash; for unknown users it compares against a fixed dummy hash so the public behaviour is identical. The session ID is 32 bytes from `crypto/rand`. The CSRF token is an independent 32 bytes from `crypto/rand`. Both are encoded with unpadded URL-safe base64 and stored under the opaque session ID. The cookie name is exactly `__Host-session`; the cookie carries only the opaque ID. The session record stores only the user ID, the expiry, and the CSRF token. Session lifetime is exactly one hour. Successful login rotates any pre-existing session bound to the request. CSRF is enforced on `POST /action` and `POST /logout` via a custom header whose value must match the server-side token; `GET /me` does not require CSRF. The session store is thread-safe, bounded, and cleaned up explicitly. Login CSRF is documented as a known limitation of the design; CORS and `SameSite` are defence in depth.

## 3. Why This Project Now?

Projects 046 through 058 produced a documented HTTP API with a rate limiter and an OpenAPI contract. None of those projects decided what counts as "logged in". Project 059 introduces session state, secret material in cookies, password verification with the bcrypt dependency introduced in Project 050, and CSRF defence. Project 060 will compose the resulting service with graceful shutdown, where session cleanup at shutdown becomes important. By the end of Project 059 the learner can ship a logged-in web application whose auth model is conservative and well documented.

## 4. Prerequisites

Required earlier projects: Project 058, Project 050, and Project 046. Earlier HTTP, JSON envelope, middleware, and bcrypt projects are useful review but are not formally required. The learner must already understand `net/http` cookies, the JSON envelope from Project 049, middleware composition from Project 048, and the bcrypt usage from Project 050. The learner must also be comfortable with `crypto/rand`, `time`, and concurrency-safe maps.

## 5. What You Must Know Before Starting

- The bcrypt dependency from Project 050: how to hash a password, how to compare a password to a hash, and the cost parameter. The bcrypt dependency is `golang.org/x/crypto` pinned at version `v0.54.0`, used through the bcrypt package. It is the only third-party library allowed in this project.
- Cookie attributes: `HttpOnly`, `Secure`, `SameSite`, `Path`, `Domain`, `Max-Age`, `Expires`, and the meaning of each. The `__Host-` cookie prefix rules: the cookie must be `Secure`, must have `Path=/`, and must not have a `Domain`.
- The difference between opaque server-side sessions and signed tokens: the cookie here is opaque, the server holds the truth.
- CSRF: the attack, why `SameSite` is defence in depth, why a custom header bound to the session is the primary defence for state-changing requests, and why safe `GET` is exempt.
- Forced-logout CSRF: why the logout endpoint must also be CSRF-protected.
- Session fixation: why rotating the session on successful login matters and what the attacker can do otherwise.
- Login CSRF: the limitation that a login endpoint that sets a cookie after a `POST` is, by design, vulnerable to login CSRF. The mitigation in this project is session rotation plus the CORS policy from Project 056.
- `crypto/rand` versus `math/rand`: only `crypto/rand` may be used.
- `httptest.NewTLSServer` and cookie-jar behaviour over HTTPS in tests.
- `-race` and the discipline of holding locks only around the data structure.
- Unpadded URL-safe base64 encoding and decoding.

## 6. Explanation of New Concepts

**Opaque server-side sessions.** A session is a record held by the server that maps an opaque ID to a user ID, an expiry, and a CSRF token. The cookie carries only the opaque ID. Anyone who steals the cookie impersonates the user, so the ID has 32 bytes of entropy from `crypto/rand`.

**Session and CSRF generation.** Two independent 32-byte values are generated from `crypto/rand`. The session ID is the lookup key. The CSRF token is bound to the session. Both are encoded with unpadded URL-safe base64 so they are safe in URLs and headers without further escaping. The generator is injected in tests so collisions and failures are deterministic.

**Cookie attributes.** The cookie name is exactly `__Host-session`. The cookie value is the encoded session ID only. `HttpOnly` is `true`. `Secure` is `true`. `SameSite` is `Lax`. `Path` is `/`. `Max-Age` is `3600`. `Expires` is the injected clock value plus one hour. `Domain` is unset. Tests that use a cookie jar use `httptest.NewTLSServer` so `Secure` is honoured; `Secure` is never turned off in tests.

**Deletion cookie.** When the session ends through logout or detected expiry, the server sends a deletion cookie with the same name and path, the same `Secure`, `HttpOnly`, and `SameSite=Lax` attributes, an empty value, an `Expires` value strictly in the past relative to the injected clock, and `Domain` unset. In the Go `http.Cookie` value the `MaxAge` field is set to `-1`. When `net/http` serialises that cookie, the wire attribute is `Max-Age=0`, which instructs the browser to delete the cookie immediately. The wire header is therefore `Max-Age=0`; the Go object's `MaxAge` field is `-1`. Tests assert both forms: the parsed `Cookie.MaxAge` is less than `0`, and the raw `Set-Cookie` header on the wire contains `Max-Age=0`.

**Password comparison boundary.** The bcrypt comparison happens at the credential boundary for every login attempt. For the known user `learner`, the comparison is against the stored hash. For any other username, the comparison is against a fixed dummy hash so that bcrypt runs in both cases and the public behaviour is identical. The login handler returns the same status and the same JSON envelope for both branches. The bcrypt comparison result, the password, the session ID, and the CSRF token are not logged.

**Session rotation on login.** When a client logs in successfully, the server creates a fresh session first (with a new ID and a new CSRF token), then atomically deletes the incoming old session and installs the fresh one. The new ID and the new CSRF token must differ from the old values. The fresh session is what the client receives in the cookie.

**Collision handling.** If the freshly generated session ID or CSRF token collides with an existing record, the server regenerates up to three times in total. After three collisions, or after the injected generator returns an error, the login returns `500` and does not create a new cookie. Any prior valid session for the same user is unchanged.

**CSRF token.** Each session carries a server-side CSRF token. The state-changing `POST /action` and `POST /logout` endpoints require a custom header (named exactly `X-CSRF-Token`) whose value matches the server-side token under constant-time comparison. The server rejects the request with `403 Forbidden` if the header is missing or wrong and does not mutate state. `GET /me` does not require the header. `SameSite=Lax` is defence in depth; it is not the primary defence.

**Login CSRF.** The login endpoint is, by design, vulnerable to login CSRF. The mitigation in this project is session rotation plus the CORS policy from Project 056. The README states this honestly. Login CSRF is not "fixed" by `SameSite` alone.

**Session store cleanup.** A `Cleanup(now)` function removes all sessions whose expiry is at or before `now`. The function is explicit and invoked by the application. No background goroutines. Expired sessions are also removed lazily on access.

**Store capacity.** The store is bounded at exactly `1000` sessions. When the store is at capacity, a login without a replaceable old session returns `503` and does not evict an existing session. Reclaiming capacity requires the application to call `Cleanup` explicitly.

**Login CSRF honestly.** The login endpoint accepts a JSON body over `application/json`. Combined with the CORS policy from Project 056, ordinary cross-origin JavaScript in a browser cannot complete the login request: the preflight fails for origins outside the Project 056 allowlist, and the simple request path does not include the credentials cookie cross-origin. `SameSite=Lax` is defence in depth. Despite these mitigations, robust pre-login CSRF and `Origin` validation remain a documented limitation of this design. Session rotation prevents session fixation; it does not by itself prevent login CSRF, and the README does not claim otherwise.

## 7. Learning Objective

After finishing this project, the learner can explain in their own words why the session ID is opaque, why it must come from `crypto/rand`, why the cookie carries only the ID, why sessions are rotated on login, what each cookie attribute is for, what CSRF is and why a custom header is the primary defence, why `SameSite` is defence in depth, why the deletion cookie uses `Max-Age: 0` on the wire and `Max-Age: -1` in the Go object, and what login CSRF is and why the design documents it honestly. The learner can also implement an `httptest` test suite that drives the full flow with a cookie jar or explicit cookies.

## 8. Functional Requirements

1. The service exposes exactly four endpoints: `POST /login`, `GET /me`, `POST /action`, `POST /logout`. No other endpoints are exposed.
2. The fixed test user is `learner` with ID `user-001`. Only the bcrypt hash of the password is stored. No plaintext password is stored.
3. `POST /login` accepts a content type whose parsed base media type is exactly `application/json`; valid media-type parameters are allowed. A missing, malformed, or non-JSON media type returns `415` with `{ "code": "unsupported_media_type", "message": "unsupported media type" }`. The body must parse as a single JSON object containing exactly the fields `username` and `password`; any other shape, any unknown field, any trailing value, or any parse failure is rejected with status `400` and body `{ "code": "invalid_request", "message": "malformed login request" }`. The body size is bounded at exactly `4096` bytes; a valid JSON document at exactly that boundary reaches credential validation, while a larger body is rejected with status `413` and body `{ "code": "payload_too_large", "message": "payload too large" }`. No `Set-Cookie` is issued on any boundary failure.
4. On valid credentials the login handler creates a fresh session with a new 32-byte session ID and a new 32-byte CSRF token encoded with unpadded URL-safe base64, atomically deletes the incoming old session and installs the fresh one, sets the `__Host-session` cookie with the documented attributes, and returns status `200` with body `{ "user": { "id": "user-001", "username": "learner" }, "csrf_token": "<unpadded url-safe base64>" }`. On failure it returns status `401` with body exactly `{ "code": "invalid_credentials", "message": "invalid username or password" }`. The same status and body are returned for unknown users and for wrong passwords. No `Set-Cookie` is issued on failure.
5. `POST /login` returns status `500` with body `{ "code": "internal_error", "message": "session generation failed" }` when the injected generator returns an error or when three consecutive collisions occur. No `Set-Cookie` is issued. Any prior valid session for the same user is unchanged.
6. `POST /login` returns status `503` with body `{ "code": "session_capacity_reached", "message": "session store is full" }` when the store holds exactly `1000` records and the login cannot complete. A login that carries a valid old session cookie and will rotate that old session is allowed to succeed: rotation replaces the old record with the fresh one in a single state transition, so the size stays at `1000` and the new session is installed. A login without a replaceable old session — that is, a login that would create a brand-new record — is rejected with `503` and no `Set-Cookie`. No existing session is evicted. `Cleanup` is the documented way to reclaim expired slots; the size-rejection rule does not depend on a hidden "cleanup-called" state.
7. `GET /me` requires a valid session cookie. On success it returns status `200` with body `{ "user": { "id": "user-001", "username": "learner" } }`. On missing, unknown, or expired session it returns status `401` with body `{ "code": "authentication_required", "message": "authentication required" }` and sends the exact deletion cookie so the browser stops retrying a dead ID. The expired record is lazily removed on access. `GET /me` does not require the CSRF header.
8. `POST /action` requires both a valid session cookie and the matching `X-CSRF-Token` header. On valid CSRF it performs the operation and returns status `200` with body `{ "ok": true }`. On missing or wrong CSRF it returns status `403` with body `{ "code": "invalid_csrf", "message": "invalid or missing CSRF token" }`. State is never mutated on a `403`.
9. `POST /logout` requires both a valid session cookie and the matching `X-CSRF-Token` header. On success it deletes the server-side session record, sends the exact deletion cookie (same name, same path, same `Secure`, `HttpOnly`, `SameSite=Lax` attributes, empty value, `Expires` strictly in the past; Go `http.Cookie.MaxAge` set to `-1`, which serialises to wire `Max-Age=0`), and returns status `204` with zero body bytes. On missing, unknown, or expired session it returns status `401` with body `{ "code": "authentication_required", "message": "authentication required" }`, sends the exact deletion cookie, and lazily removes the expired record. On missing or wrong CSRF it returns status `403` with body `{ "code": "invalid_csrf", "message": "invalid or missing CSRF token" }` and leaves the session record unchanged.
10. The session store is a map guarded by a mutex. Expired sessions are removed lazily on access. An explicit `Cleanup(now)` function removes all sessions whose expiry is at or before `now` and is invoked by the application; tests call it directly.
11. The session store is bounded at exactly `1000` sessions. When the store is at capacity, a login without a replaceable old session is rejected with `503` and no existing session is evicted. Reclaiming capacity requires the application to call `Cleanup` explicitly.
12. The session ID generator is injected. The default generator reads from `crypto/rand`. The generator never falls back to `math/rand`.
13. The login handler always runs bcrypt: against the stored hash for known users, against a fixed dummy hash for unknown users. The bcrypt cost is pinned in configuration. Tests use a low cost to keep CI fast.
14. The login handler, the logout handler, the session store, and the CSRF middleware never log the cookie value, the session ID, the CSRF token, the plaintext password, or the bcrypt hash. Tests assert this where possible by capturing log output and asserting those strings are absent.
15. The CSRF token comparison uses constant-time comparison. The behaviour test asserts that wrong-length and wrong-value CSRF inputs both receive the identical `403` body and that no state mutation occurs. The test cannot prove the absence of a timing side channel; constant-time primitive use is a code-review and design requirement, not an observable property.

## 9. Inputs and Outputs

Inputs: HTTP requests to the four endpoints. Cookies are read by name; the CSRF header is read by name. Outputs: documented JSON bodies, cookies with the documented attributes, and the documented status codes. Example textual inputs and expected textual outputs:

- `POST /login` body `{"username":"learner","password":"correct"}` against the user whose stored hash matches. Expected: status `200`, body `{"user":{"id":"user-001","username":"learner"},"csrf_token":"<unpadded url-safe base64>"}`, `Set-Cookie: __Host-session=<opaque>; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=3600; Expires=<clock + 1 hour>`.
- `POST /login` body `{"username":"learner","password":"wrong"}`. Expected: status `401`, body exactly `{ "code": "invalid_credentials", "message": "invalid username or password" }`. No `Set-Cookie`.
- `POST /login` body `{"username":"nobody","password":"x"}`. Expected: status `401`, body identical to the previous case. No `Set-Cookie`.
- `POST /login` body `not-json`. Expected: status `400`, body `{ "code": "invalid_request", "message": "malformed login request" }`. No `Set-Cookie`.
- `POST /login` body that is JSON but not a single object (for example an array, a number, or a string). Expected: status `400`, body `{ "code": "invalid_request", "message": "malformed login request" }`. No `Set-Cookie`.
- `POST /login` body that is a single object with an unknown field (for example `{"username":"learner","password":"correct","role":"admin"}`). Expected: status `400`, body `{ "code": "invalid_request", "message": "malformed login request" }`. No `Set-Cookie`.
- `POST /login` body whose total request-body size is exactly `4097` bytes. Expected: status `413`, body `{ "code": "payload_too_large", "message": "payload too large" }`. No `Set-Cookie`.
- `GET /me` with no cookie. Expected: status `401`, body `{ "code": "authentication_required", "message": "authentication required" }`, plus the deletion cookie so the browser stops retrying.
- `GET /me` with a valid session cookie. Expected: status `200`, body `{"user":{"id":"user-001","username":"learner"}}`. No CSRF header required.
- `POST /action` with a valid session cookie but no `X-CSRF-Token` header. Expected: status `403`, body `{ "code": "invalid_csrf", "message": "invalid or missing CSRF token" }`. State unchanged.
- `POST /action` with a valid session cookie and a wrong `X-CSRF-Token` (any wrong value, including a wrong-length value). Expected: status `403`, body identical to the missing case. State unchanged.
- `POST /action` with a valid session cookie and the matching `X-CSRF-Token`. Expected: status `200`, body `{"ok":true}`.
- `POST /logout` with a valid session cookie and the matching `X-CSRF-Token`. Expected: status `204`, body zero bytes, `Set-Cookie: __Host-session=; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=0; Expires=<clock - 1 second>` on the wire. The session record is deleted. The parsed `http.Cookie.MaxAge` field is less than `0`.
- `POST /logout` with a missing cookie. Expected: status `401`, body `{ "code": "authentication_required", "message": "authentication required" }`, plus the deletion cookie. No session record existed to be removed.
- `POST /logout` with a valid cookie but missing CSRF. Expected: status `403`, body `{ "code": "invalid_csrf", "message": "invalid or missing CSRF token" }`. No `Set-Cookie`. State unchanged.

## 10. Rules and Edge Cases

- A session whose expiry is at or before the injected clock value is rejected with `401` and lazily removed. A second request with the same cookie continues to receive `401` and the record stays gone.
- An unknown session ID is rejected with `401`. No record is created.
- The CSRF header name is exactly `X-CSRF-Token`. The cookie name is exactly `__Host-session`. Tests assert both.
- `Max-Age` is exactly `3600` on a fresh cookie. The deletion cookie's Go `http.Cookie.MaxAge` field is set to `-1`, which serialises to the wire attribute `Max-Age=0`; the `Expires` value is strictly in the past relative to the injected clock.
- `Domain` is unset on every cookie.
- The session store never logs the session ID, the CSRF token, the password, the bcrypt hash, or the cookie value.
- A successful login with no existing session cookie behaves identically to a successful login with an existing session cookie: a fresh session is created.
- A successful login with an existing session cookie deletes the old record and installs the fresh one. The new session ID and CSRF token differ from the old values.
- A `GET /action` returns `405 Method Not Allowed` (or the documented status) and does not require CSRF.
- The login endpoint does not return `200 OK` for unauthenticated requests.
- The bcrypt cost is pinned by configuration. Tests use a low cost to keep CI fast.
- Collision retry: up to three total attempts to generate a unique session ID and CSRF token. After three collisions the login returns `500`.
- Generator failure: when the injected generator returns an error, the login returns `500` with no new cookie. Any prior valid session is unchanged.

## 11. Project Constraints

- Standard library plus the bcrypt dependency `golang.org/x/crypto` at version `v0.54.0`, used through the bcrypt package. No other third-party libraries.
- The session store is in-memory. Persistence is out of scope; shutdown cleanup is documented.
- The session ID and CSRF generators are injected. The default is `crypto/rand`.
- No background goroutines owned by the session middleware.
- The application invokes cleanup explicitly. Tests call cleanup directly.
- No real wall clock in tests. The clock is injected.
- Tests that use a cookie jar use `httptest.NewTLSServer`. `Secure` is never turned off.

## 12. Design Questions Before Coding

- How is the user store represented? A small in-memory map seeded from configuration. The known user is `learner` with ID `user-001`. Unknown users fall through to a fixed dummy bcrypt hash.
- How is the session record represented? A struct with `userID`, `expiresAt`, `csrfToken`. The mutex is held only around the map and the record.
- How is the login handler implemented as one branch that returns the same generic failure for every bad-credential path? The bcrypt comparison runs in both branches; the response status and body are identical.
- How is the CSRF token bound to the session? It is generated when the session is created and stored in the same record. Rotation produces a new token.
- How is the deletion cookie constructed? Same name, same path, same `Secure`, `HttpOnly`, `SameSite=Lax` attributes, empty value, `Expires` strictly in the past. The Go `http.Cookie.MaxAge` field is set to `-1`; when `net/http` writes the cookie, the wire attribute is `Max-Age=0`.
- How is `Cleanup` invoked by the application? On every Nth request, on shutdown, on a schedule. The session middleware does not own a ticker.
- How are the four endpoints registered? Through the router pattern from Project 048 with the session middleware applied to the protected routes.
- How is constant-time comparison used? `crypto/subtle.ConstantTimeCompare` returns 1 only when the slices are equal length and equal contents.

## 13. Implementation Milestones

1. Sketch the configuration, the user store, the session record, and the session store on paper.
2. Implement the injected session ID and CSRF token generator and the default `crypto/rand` implementation that returns unpadded URL-safe base64.
3. Implement the session store with `Create`, `Get`, `Delete`, `Rotate`, and `Cleanup`, all under a mutex.
4. Implement the cookie helper that writes the documented attributes from the configuration and the deletion cookie.
5. Implement the login handler: parse, look up the user, run bcrypt (against the stored hash or the dummy hash), branch on success and failure, generate session and CSRF, rotate any pre-existing session, set the cookie, return the documented body.
6. Implement the `GET /me` handler: read the cookie, look up the session, lazy-delete if expired, return the user or `401`.
7. Implement the CSRF middleware: read the `X-CSRF-Token` header, compare to the session record under constant-time comparison, allow or `403`.
8. Implement the `POST /action` handler: composed of the session middleware plus the CSRF middleware plus the handler.
9. Implement the `POST /logout` handler: composed of the session middleware plus the CSRF middleware plus the handler that deletes the record and writes the deletion cookie.
10. Wire the four handlers into the router with the appropriate middleware composition.
11. Write tests using `httptest.NewTLSServer` with a cookie jar or explicit cookie capture. Cover the verification list.
12. Review the verification list and confirm every item is covered before declaring the project complete.

## 14. Verification Cases the Learner Must Write

Each item is a behavioural specification. The learner writes the corresponding `go test` code.

- Login right: valid credentials produce status `200`, body `{"user":{"id":"user-001","username":"learner"},"csrf_token":"<token>"}`, and a `Set-Cookie` whose name is `__Host-session`, value is the encoded session ID, and whose attributes are exactly `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`, `Max-Age=3600`, and `Expires` equal to the injected clock plus one hour. No `Domain`.
- Login wrong password: status `401`, body exactly `{ "code": "invalid_credentials", "message": "invalid username or password" }`. No `Set-Cookie`.
- Login unknown user: status `401`, body identical to the wrong-password case. No `Set-Cookie`.
- Login malformed body: status `400`, body `{ "code": "invalid_request", "message": "malformed login request" }`. No `Set-Cookie`. The malformed cases include non-JSON, JSON that is not a single object, single objects with unknown fields, and any other parse or shape failure.
- Login body boundary: an otherwise valid single-object request of exactly `4096` bytes reaches credential validation and does not return `413`; a `4097`-byte request returns `413` with `{ "code": "payload_too_large", "message": "payload too large" }` and no `Set-Cookie`.
- Login media type: missing, malformed, or non-JSON media types return `415` with `{ "code": "unsupported_media_type", "message": "unsupported media type" }`; `application/json` with a valid charset parameter passes the media-type check.
- Cookie attributes: parse the `Set-Cookie` and assert every attribute above. Tests run against `httptest.NewTLSServer`; `Secure` is asserted as present and never disabled.
- Session rotation on login: a request that arrives with a valid existing session cookie and then logs in receives a new session ID and a new CSRF token, and the old record is gone. A second request with the old cookie returns `401`.
- `GET /me` without cookie: status `401`, body `{ "code": "authentication_required", "message": "authentication required" }`.
- `GET /me` with valid cookie: status `200`, body `{"user":{"id":"user-001","username":"learner"}}`.
- `GET /me` with expired cookie (advanced by the fake clock): status `401`, the record is removed lazily, the next request with the same cookie is also `401`.
- `POST /action` missing CSRF: status `403`, body `{ "code": "invalid_csrf", "message": "invalid or missing CSRF token" }`. State unchanged.
- `POST /action` wrong CSRF: status `403`, body identical to the missing case. State unchanged.
- `POST /action` right CSRF: status `200`, body `{"ok":true}`. State updated.
- `GET /me` does not require CSRF: the protected `GET` returns `200` without an `X-CSRF-Token` header.
- `POST /logout` with valid cookie and matching CSRF: status `204`, body zero bytes, raw `Set-Cookie` contains `__Host-session=; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=0; Expires=<clock - 1 second>` (with no `Domain`). The parsed `http.Cookie.MaxAge` field is less than `0`. The session record is deleted.
- `POST /logout` with valid cookie and missing CSRF: status `403`, body `{ "code": "invalid_csrf", "message": "invalid or missing CSRF token" }`. No `Set-Cookie`. State unchanged.
- `POST /logout` with valid cookie and wrong CSRF: status `403`, body identical to the missing case. State unchanged.
- `POST /logout` with missing cookie: status `401`, body `{ "code": "authentication_required", "message": "authentication required" }`. State unchanged.
- `POST /logout` with an expired cookie: status `401`, body identical to the missing-cookie case, expired record removed lazily, and the exact deletion cookie sent. No unrelated session changes.
- Deletion cookie attributes on the wire: `__Host-session`, empty value, `Max-Age=0`, `Expires` strictly in the past relative to the injected clock, `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`. No `Domain`. The parsed `http.Cookie.MaxAge` field is less than `0`.
- CSRF token rotation: after a fresh login, the new session has a different CSRF token from any previous session. After logout, the CSRF token is no longer valid.
- Session isolation: two distinct logins for the fixed user receive distinct session IDs and CSRF tokens. Both authenticate as `learner`; logging out one session does not invalidate the other.
- Generator failure: when the injected generator returns an error, the login handler returns `500`, body `{ "code": "internal_error", "message": "session generation failed" }`, no `Set-Cookie`, and any prior valid session is unchanged.
- Generator collision retry: an injected generator that returns a colliding ID on the first attempt and a unique ID on the second succeeds; the recorded ID is the unique one.
- Generator collision exhaustion: an injected generator that returns colliding IDs for all three attempts makes the login return `500`, no `Set-Cookie`, and no record is created.
- TTL boundary at expiry: a session whose expiry equals the injected clock value is rejected with `401` and removed by `Cleanup(now)`.
- TTL boundary just before expiry: a session whose expiry is one nanosecond after the injected clock value is still valid.
- Lazy cleanup: an expired session is removed when the next request reads it.
- Explicit cleanup: `Cleanup(now)` removes every session whose expiry is at or before `now`.
- Store capacity: with the store holding `1000` records and the login request carrying no replaceable old session, the login returns `503`, body `{ "code": "session_capacity_reached", "message": "session store is full" }`, no `Set-Cookie`, no existing session is evicted.
- Capacity rotation: a login that carries a valid old session and would otherwise be rejected by the size check is allowed to complete by replacing the old record with the fresh one in a single state transition; the store stays at `1000`. The new session ID and CSRF token differ from the old values.
- Capacity reclaim after cleanup: after `Cleanup(now)` removes expired sessions, the freed slot can be used by a new login.
- Constant-time comparison observable: a wrong-length CSRF header and a wrong-value CSRF header both receive the identical `403` body and both leave state untouched. The behaviour test asserts these observable properties only. The test cannot prove the absence of a timing side channel; constant-time primitive use is a code-review and design requirement, not an observable property under a deterministic test.
- No-secret logs: capture the logger output during login, logout, and CSRF failure and assert that the session ID, the CSRF token, the plaintext password, the bcrypt hash, and the cookie value do not appear.
- Concurrency: a hammer test fires many goroutines that log in, access `GET /me`, hit `POST /action`, and log out; `-race` reports nothing.
- Forced-logout CSRF coverage: a cross-site request that submits `POST /logout` without the `X-CSRF-Token` header returns `403` and does not log the user out.

## 15. Common Mistakes to Watch For

- Using `math/rand` for session IDs. The entropy is too low and the ID is predictable.
- Logging session IDs, CSRF tokens, passwords, or cookie values. Tests catch this; the implementation must not.
- Setting `SameSite=None` to "fix" a cross-site integration. The cookie must remain `Lax` and the integration must use a different mechanism.
- Skipping session rotation on login. Session fixation reappears.
- Setting `Domain` on the cookie. The `__Host-` prefix requires no `Domain`, and a parent domain broadens the attack surface.
- Setting `HttpOnly=false` because the front-end needs to read the session. The cookie is opaque; the front-end must not read it.
- Treating `SameSite` as the primary CSRF defence. The custom header is the primary defence; `SameSite` is defence in depth.
- Comparing CSRF tokens with `==` after a constant-time helper is available. Use `crypto/subtle.ConstantTimeCompare`.
- Skipping the lazy deletion of expired sessions. Memory grows over time.
- Letting the session middleware start a background goroutine. The application owns the cleanup cadence.
- Returning different bodies for "no such user" and "wrong password". User enumeration is the result.
- Sending a literal `Max-Age: -1` on the wire. The Go object's `MaxAge` field is `-1`, but `net/http` writes `Max-Age=0` on the wire. The two forms are consistent: the parsed `MaxAge < 0` and the raw header `Max-Age=0` both instruct deletion.
- Skipping CSRF on `POST /logout`. Forced-logout CSRF would succeed.
- Turning `Secure` off in tests. Use `httptest.NewTLSServer` instead.
- Treating login CSRF as fixed by `SameSite` alone. It is not. The README says so honestly. Session rotation prevents fixation, not login CSRF.
- Treating session rotation as a mitigation for login CSRF. Rotation prevents fixation; the login endpoint itself remains vulnerable to login CSRF by design, and the README treats that as a documented limitation mitigated by `application/json`, the CORS policy from Project 056, and `SameSite=Lax` as defence in depth.

## 16. Topics and References for Study

- OWASP Cheat Sheet, "Session Management" and "Cross-Site Request Forgery Prevention".
- OWASP Authentication Cheat Sheet, particularly the generic-failure guidance and the `__Host-` cookie prefix.
- The bcrypt dependency documentation from Project 050 and the `golang.org/x/crypto` release notes for `v0.54.0`.
- The `net/http` package documentation for `Cookie`, `SameSite`, `SetCookie`, `Cookie.MaxAge`, and `http.Request.Cookie`.
- RFC 6265 for HTTP cookies, including the `Max-Age` semantics and the deletion pattern.
- The Fetch Living Standard for `SameSite` semantics.
- Go `crypto/rand` and `crypto/subtle` documentation.
- The OpenAPI contract from Project 058 for documenting the four endpoints.

## 17. Self-Assessment Questions

1. Why is the session ID opaque? Why must it come from `crypto/rand`?
2. Why is the cookie `HttpOnly` and `Secure`? What attack does each attribute prevent?
3. Why is the session rotated on successful login? Walk through a fixation attack.
4. Why is the CSRF defence a custom header rather than relying on `SameSite` alone?
5. Why is the logout endpoint CSRF-protected? Walk through a forced-logout attack.
6. Why is the login endpoint vulnerable to login CSRF even with the documented mitigations? What does this project actually do about it, and what does it explicitly not claim?
7. Why are expired sessions removed lazily on access and explicitly by cleanup? What does each address?
8. Why must the session ID, the CSRF token, the cookie value, the plaintext password, and the bcrypt hash never appear in logs?
9. The deletion cookie's Go `http.Cookie.MaxAge` field is `-1` and the wire attribute is `Max-Age=0`. Why are both correct? What does each form instruct?
10. Why can a deterministic behaviour test prove that wrong-length and wrong-value CSRF inputs both get the identical `403` body, but cannot prove that there is no timing side channel?

## 18. Definition of Completion

The project is complete when, in addition to the rules above:

- Every item in the verification list is a passing test that the learner wrote themselves.
- The tests pass under `go test -race ./...` from the project folder.
- The only third-party dependency in `go.mod` is `golang.org/x/crypto` at version `v0.54.0`, used through the bcrypt package.
- The session store contains no background goroutines and no real wall clock.
- The cookie attributes are asserted in tests, including `Secure` asserted as present on every cookie the server emits.
- The learner can answer every self-assessment question without rereading the README.
- The README or the test comments include the honest statement that login CSRF is a known limitation of the design. The mitigations are `application/json` plus the CORS policy from Project 056 plus `SameSite=Lax` as defence in depth. Session rotation prevents fixation, not login CSRF, and the README does not claim otherwise.

## 19. Optional Extensions

At most two. Pick one only if the core project is already complete and tested. Optional extensions must not change the documented contracts.

- Add a documented "active sessions" endpoint that lists the current user's sessions and supports explicit revocation. The endpoint is CSRF-protected on state-changing actions.
- Add a per-user "last login at" timestamp returned in the `GET /me` body. The timestamp comes from the injected clock and is not persisted.
