# Project 079 — Load Balancer Round Robin

## 1. Project Name and Number
Project 079, load_balancer_round_robin. This README is a learning guide only. You will create every source and test file yourself in `06-networking/079_load_balancer_round_robin/`. This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea
A reverse-proxy load balancer that distributes requests across a fixed non-empty list of backend URLs using a thread-safe round-robin selection. The selection is a cursor over the fixed backend order: the selection scans from the cursor position to the next healthy backend, chooses it, and then advances the cursor to the following physical slot. The cursor advances exactly once per successful selection, including when the request handled by that selection later fails. Health state starts healthy for every backend. Health state changes only because of an explicit probe or because of a selected request's transport failure; an ordinary successful transport response, including an upstream HTTP 5xx, does not change health, and only a later 2xx probe restores an unhealthy backend. The health URL for each backend is exactly that backend's origin followed by `/healthz`. The balancer preserves Project 078's corrected exact proxy safety contract.

## 3. Why This Project Now?
This project requires Project 078 (reverse_proxy) as the immediate predecessor, Project 071 (tcp_echo_server) for TCP framing, idle deadlines, and per-connection protocol error discipline, and Project 060 (graceful_shutdown_web) for graceful server shutdown and lifecycle ownership. Project 041 (context_timeout_example) is recommended review for context cancellation and deadlines and is optional review only, not a formal prerequisite. This project layers round-robin selection and explicit probe-driven health state on top of the fixed-upstream proxy discipline while preserving the corrected exact proxy safety rules from Project 078.

## 4. Prerequisites
Projects 078, 071, and 060 are required prerequisites. Project 078 is the immediate predecessor for the corrected exact proxy safety contract. Project 071 is required for TCP connection handling, byte framing, idle deadlines, accept-loop shutdown, and per-connection protocol error discipline. Project 060 is required for graceful server shutdown and lifecycle ownership. Project 041 is recommended review for context cancellation propagation but is not a formal prerequisite. No public network, no Docker, no environment variables for the backend list. The backend list is a configuration value supplied at construction. Health probes are explicit operations driven by tests through barriers; production may schedule them but the required behavior does not depend on a ticker.

## 5. What You Must Know Before Starting
Know the `net/http` request and response model, `httputil.ReverseProxy`, atomic counters or mutex-protected indices, context cancellation and deadline propagation, transport-level timeouts, the difference between idempotent and non-idempotent methods, the difference between probe-driven health and request-driven failure, the rule that 5xx is a successful transport response and not a transport failure, the rule that the all-unhealthy path does not perform a selection, and the race detector.

## 6. Explanation of New Concepts
Each backend URL is an origin-only `http` or `https` URL: a scheme and a host are required, and a default port for the scheme is filled in if absent. Userinfo, query, fragment, and non-root path are not permitted in the backend URL. Trailing root slashes are normalized away before duplicate detection. After this normalization, two backends that compare equal as origins are canonical duplicates and are rejected at startup. Invalid URLs are rejected at startup. An empty backend list is rejected at startup. There is exactly one health URL per backend and that health URL is exactly the backend's normalized origin followed by `/healthz`.

Each backend starts healthy. Health state changes only because of an explicit probe or because of a selected request's transport failure. An explicit probe performs one bodyless HTTP `GET` against the exact health URL using an injected client with dial, response-header, and whole-request timeouts. Any 2xx response marks the backend healthy; any non-2xx response, timeout, or transport failure marks it unhealthy. The response body does not carry health data. The probe discards at most 4,096 bytes and always closes the body; whether the transport can reuse that connection afterward is not part of the health contract. A larger body is not drained further and does not change the status-based verdict. No verdict or public error leaks the backend address.

The proxy call to a selected backend and the probe operation are distinct operations. A successful transport response from a selected request, including a 5xx, is forwarded through Project 078's response allowlist to the client and does not change the backend's health state. Only a transport failure during a selected request marks the backend unhealthy. A transport failure is observed when no upstream response was received at all, such as a dial error or a response header timeout. A 5xx response is a received response and is therefore not a transport failure.

Round-robin selection is over the fixed backend order using a cursor. The cursor points at a physical slot. To make a selection, the algorithm scans from the cursor position forward through the backend list once. The first healthy backend encountered is chosen. The cursor is then advanced to the following physical slot, and that advancement is recorded as a successful selection. The cursor advances exactly once per successful selection. If a later request against the selected backend fails with a transport failure, the cursor has already advanced and that failure does not retroactively undo the cursor advance. When no healthy backend is found in the scan, the all-unhealthy path returns 503 without performing a selection; that path does not advance the cursor.

Under concurrency, the cursor advance is atomic with the chosen slot so that two concurrent requests cannot observe the same successful selection for the same cursor advance. The state and cursor are protected against data races; the round-robin selection is race-free under the race detector.

One backend is selected per successful selection. The selected backend is the upstream target for that request. The proxy call preserves Project 078's exact safety contract: prefix `/api`, configured outbound URL and Host, forwarding-header trust, hop-by-hop removal, response allowlist, 502 behavior, upstream status forwarding, and no retry. Caching, WebSocket upgrades, and dynamic discovery remain excluded; cursor-based round-robin over the fixed list is the only balancing behavior added here.

Transport failure on a selected request marks that backend unhealthy and returns 502 without retrying the current request. The mark is recorded after the transport failure is observed. The current request is not retried, so a `POST` cannot duplicate on another backend. A later selection observes that backend as unhealthy and skips it during the scan.

When all backends are unhealthy the balancer returns 503 without proxying or advancing the cursor. The response has `Content-Type: application/json` and a JSON object whose only field is `error` with value `service_unavailable`, with no trailing newline. It never includes backend addresses or internals.

The balancer adds no identity feature of its own. Backend test fixtures return distinct ordinary response bodies, and the proxy forwards those bodies verbatim through the allowlist. Production wiring adds or strips nothing for identity.

Concurrent aggregate counts need not assign a specific competing request-to-backend order. The aggregate count difference-at-most-one property applies only to a complete batch where the healthy set is stable and where no probe and no transport failure occurred during the batch; under those conditions, after `N` complete requests to `K` healthy backends, each backend's count is either `floor(N/K)` or `ceil(N/K)` and the difference between any two backends is at most one. The tests assert that stable property and assert no specific competing order.

Text-only protocol examples are permitted. As a prose shape: three healthy backends A, B, C with the cursor initially at A serve six sequential requests in the exact order A, B, C, A, B, C. Concurrent requests under the same stable health set need not have a pinned request-to-backend order, but their completed aggregate counts differ by at most one. After an explicit probe marks B unhealthy, later sequential selections skip B and rotate through A and C; a later 2xx probe restores B and makes it eligible again when the cursor scan next reaches it. After probes mark A, B, and C unhealthy, the next request returns 503 without an upstream call or cursor advance. A `POST` whose selected backend has a transport failure returns 502, marks that backend unhealthy, is not retried, and is not duplicated on another backend; the cursor advance from that selection stands.

## 7. Learning Objective
Implement a fixed-list round-robin load balancer with cursor-based selection, explicit probe-driven health state, transport-failure-driven health changes, single-successful-selection-per-cursor-advance, all-unhealthy 503 without selection, full preservation of Project 078's corrected exact proxy safety contract, and tests that pin cycle, skip, restoration, all-unhealthy, and concurrency behavior without sleep and without depending on a specific competing order.

## 8. Functional Requirements
1. Each backend URL is an origin-only `http` or `https` URL with required scheme and host, optional default port for the scheme, normalized trailing root; no userinfo, query, fragment, or non-root path.
2. Invalid URLs are rejected at startup. Canonical duplicates after normalization are rejected at startup. An empty backend list is rejected at startup.
3. Each backend's health URL is exactly that backend's normalized origin followed by `/healthz`.
4. Each backend starts healthy. Health state changes only because of an explicit probe or because of a selected request's transport failure.
5. An ordinary successful transport response from a selected request, including an upstream HTTP 5xx, does not change health; a later 2xx probe restores the backend to healthy.
6. The probe sends one `GET` to the health URL using an injected client with dial, response header, and request timeouts.
7. A probe is one bodyless `GET /healthz`; any 2xx is healthy and any non-2xx, timeout, or transport failure is unhealthy. Its body carries no health data, is discarded up to 4,096 bytes, and is always closed; a larger body is not drained further and does not change the status verdict.
8. The probe does not leak internal addresses in any verdict path.
9. The proxy operation and the probe operation are distinct; a selected-request 5xx is forwarded with the allowlist and does not mark the backend unhealthy.
10. Round-robin is a cursor over the fixed backend order. The selection scans from the cursor position forward once to the next healthy backend, chooses it, and advances the cursor to the following physical slot.
11. The cursor advances exactly once per successful selection, including when the request handled by that selection later fails with a transport failure.
12. The all-unhealthy path returns 503 without performing a selection; the cursor does not advance on that path.
13. Concurrent cursor advance and selection are atomic so two concurrent requests cannot observe the same successful selection for the same cursor advance; state and cursor are race-free under the race detector.
14. One backend is selected per successful selection; the selection is the upstream target for that one request.
15. Project 078's exact proxy safety contract is preserved end-to-end, including its 400 and 404 path outcomes, forwarding and hop-by-hop rules, response allowlist, 502 behavior, upstream status forwarding, and no retry. No cache, WebSocket, or dynamic discovery is added; round-robin over the fixed list is the only balancing behavior.
16. Transport failure on a selected request marks the selected backend unhealthy and returns 502 without retrying the current request.
17. A later selection skips a backend marked unhealthy.
18. All-unhealthy returns 503 with `Content-Type: application/json` and only `error: "service_unavailable"` in the JSON body, with no trailing newline, backend address, or internal detail.
19. The balancer adds no identity feature. Test backends return distinct ordinary bodies that the proxy forwards; production adds or strips nothing for identity.
20. Under concurrency with a stable healthy set and no probes or transport failures during the batch, aggregate counts differ by at most one after a complete batch; no specific competing order is asserted.

## 9. Inputs and Outputs
Balancer input is an `http.Request` on the listener. Configuration input is the fixed backend list, injected probe client, and injected upstream transport. Output is a successful proxy response, Project 078's exact 502 on selected pre-response transport failure, the exact `service_unavailable` 503 when all backends are unhealthy, a 404 for an off-boundary path, or a 400 for a suspicious in-boundary path.

## 10. Rules and Edge Cases
Empty backend list is a startup error. Invalid URL is a startup error. Canonical duplicate after normalization is a startup error. Backend starts healthy. Probe 2xx restores healthy. Probe non-2xx, timeout, or transport failure marks unhealthy. Probe response content does not affect the verdict, is discarded only up to 4,096 bytes, and is closed. Selected-request 5xx does not change health. Selected-request transport failure marks unhealthy and returns 502 without retry. `POST` is not duplicated on another backend. A later selection skips an unhealthy backend. The all-unhealthy path returns 503 without performing a selection and without advancing the cursor. The cursor advances exactly once per successful selection, including when that selected request later fails. Concurrent selection is race-free. Sequential selection from an initial cursor over stable healthy A, B, C is exactly A, B, C, A. Project 078's corrected exact proxy safety contract is preserved end-to-end. Upstream 4xx and 5xx are forwarded with their own status. The balancer adds no identity feature.

## 11. Project Constraints
Fixed backend list. No dynamic discovery. No public network. No Docker. No environment variables for the backend list. The probe client and upstream transport are injected. No retry. No cache. No WebSocket. No load balance beyond round-robin over the fixed list. The `httptest` package is used for backends and for the balancer listener. Production may schedule probes on an interval; the required behavior does not depend on a ticker.

## 12. Design Questions Before Coding
How is the backend URL normalized so that the default port, trailing root, and lowercase host are pinned before duplicate detection? How does the probe use only the status verdict, discard at most 4,096 response bytes, and close every body without leaking addresses? How is the cursor advance ordered relative to the proxy call so that the cursor advances exactly once per successful selection and not after the request outcome? How does an initial cursor prove the exact sequential A, B, C, A cycle? How is the all-unhealthy path implemented so it does not perform a cursor advance? How is a selected-request transport failure distinguished from a 5xx response that arrived successfully? How are concurrent cursor advances made atomic with the chosen slot? How is Project 078's corrected exact proxy safety contract preserved at the balancer layer without re-deriving any of its rules? How is the aggregate count difference-at-most-one property proved under a stable healthy set without asserting a specific competing order?

## 13. Implementation Milestones
1. Define the startup validation: scheme and host required, no userinfo/query/fragment/non-root path, default port per scheme, trailing root normalization, lowercase host, canonical duplicate detection, empty list rejection.
2. Define the backend state struct with a health flag and a shared cursor with atomic or mutex protection.
3. Define the bodyless `GET /healthz` probe with injected timeouts, status-only verdict, at-most-4,096-byte discard, and unconditional body close.
4. Define the round-robin selection that scans from the cursor to the next healthy backend, chooses it, and advances the cursor to the following physical slot.
5. Define the proxy call to the selected backend with the injected transport and Project 078's corrected exact proxy safety rules.
6. Define the transport-failure-to-unhealthy marking after a successful selection returns a transport failure; the current request is not retried.
7. Define the all-unhealthy 503 path with exact `service_unavailable` JSON, no newline or internal detail, no selection, and no cursor advance.
8. Define the pinned 502 envelope inherited from Project 078 for pre-response transport failure on a selected request.
9. Define `httptest` backend wiring using distinct ordinary response bodies, balancer listener wiring, and the test matrix.
10. Define the full matrix of cycle, wrap, probe healthy to unhealthy to healthy, skip unhealthy, all unhealthy 503, transport failure marks unhealthy without retry, concurrent balanced counts under a stable healthy set, one backend, empty or invalid or duplicate at startup.

## 14. Verification Cases the Learner Must Write
- Startup with an empty list returns an error.
- Startup with an invalid URL returns an error.
- Startup with canonical duplicate URLs after normalization returns an error.
- A backend URL with userinfo, query, fragment, or non-root path returns a startup error.
- Each backend starts healthy.
- A 2xx probe marks the backend healthy.
- A non-2xx probe response marks the backend unhealthy.
- A probe that times out or returns a transport error marks the backend unhealthy and does not leak addresses.
- Probe response content does not affect the status verdict; at most 4,096 bytes are discarded and the body is always closed. Connection reuse after the probe is not asserted.
- After a transport failure during a selected request, the backend is unhealthy and the next selection skips it; the request that observed the failure is not retried on another backend.
- A selected-request 5xx is forwarded with the allowlist and does not change the backend's health.
- Only a later 2xx probe restores an unhealthy backend to healthy.
- The cursor advances exactly once per successful selection even when the request handled by that selection later fails with a transport failure.
- The cursor does not advance on the all-unhealthy path.
- The all-unhealthy path returns 503, exact `application/json`, and only `error: "service_unavailable"` with no trailing newline; no upstream call, address, or cursor advance occurs.
- A single-backend deployment serves all requests to that one backend when healthy and returns 502 or 503 as appropriate when unhealthy or transport-failed.
- Three sequential requests from an initial cursor over healthy A, B, C select exactly A, B, C; the next three select exactly A, B, C again.
- Three healthy backends with no probes or transport failures during the batch serve a complete batch with aggregate counts differing by at most one; no specific competing request-to-backend order is asserted.
- Method, query, body, and path mapping follow Project 078: exact `/api` and `/api/` plus safe suffix are the only proxied paths; off-boundary returns 404 with no upstream call.
- Forwarding headers follow Project 078: strip all inbound; synthesize from a trusted peer only as exactly one `X-Forwarded-For`, `X-Forwarded-Host`, and `X-Forwarded-Proto`.
- Hop-by-hop removal follows Project 078: `Trailer`, `Proxy-Connection`, and a complete `TE` strip; `Connection`-named headers are also removed.
- The response allowlist follows Project 078 exactly; the fixed `Via` is added; upstream 5xx is forwarded with the allowlist and not converted to 502.
- Context cancellation cancels the upstream transport for the selected request; the cancellation is not presented as a synthetic 502.
- Concurrent requests run race-free under the race detector.
- Backend test fixtures return distinct ordinary response bodies; the balancer adds no identity feature.
- No public network is contacted; backends and balancer are `httptest`.
- No sleep synchronization is used in tests.

## 15. Common Mistakes to Watch For
Accepting an empty, invalid, or canonical-duplicate backend list at startup; allowing userinfo, query, fragment, or non-root path in a backend URL; treating a 5xx response as a transport failure; using probe body content as the health verdict; reading an unbounded probe body; failing to close a probe body; leaking addresses in the probe verdict path; advancing the cursor on the all-unhealthy path or after the request outcome; failing the exact sequential cycle; treating transport failure as a retry trigger; duplicating a `POST` on another backend; including backend addresses in the 503 envelope; inheriting a relaxed proxy policy; converting an upstream 5xx to 502; asserting a specific competing order under concurrency; using sleep to synchronize tests; and re-deriving Project 078's rules inconsistently.

## 16. Topics and References for Study
Study cursor-based round-robin selection, atomic or mutex-protected indices, explicit probe-driven health state, the distinction between 5xx and transport failure, status-only health verdicts, bounded probe body discard and close, transport-failure mapping without retry, the all-unhealthy 503 path with no cursor advance, aggregate count difference-at-most-one under a stable healthy set, and Project 078's corrected exact proxy safety rules. Review the Go `net/http`, `httputil`, `sync`, `sync/atomic`, and `httptest` documentation. Read the prior README for Project 078 as the immediate predecessor for the corrected exact proxy safety rules, Project 071 for TCP framing and protocol error discipline, and Project 060 for graceful server shutdown and lifecycle ownership. Project 041 for context cancellation propagation is optional review.

## 17. Self-Assessment Questions
Why is the backend URL normalized to origin only with required scheme and host, and why are userinfo, query, fragment, and non-root path rejected? Why is health changed only by an explicit probe or a selected request's transport failure, and why is a 5xx response not a transport failure? Why is the probe verdict based only on status, why are at most 4,096 body bytes discarded, and why is the body always closed? Why does the cursor advance exactly once per successful selection and not after the request outcome, and why does the all-unhealthy path not advance? Why does the initial cursor yield the exact sequential cycle? Why does selected-request transport failure return 502 without retry? Why does the aggregate balance property require a stable healthy set with no probes or failures? Why is Project 078's corrected contract preserved rather than re-derived?

## 18. Definition of Completion
- [ ] Backend URLs are origin-only with required scheme and host, no userinfo/query/fragment/non-root path, normalized before duplicate detection; empty list and canonical duplicates produce startup errors.
- [ ] Each backend's health URL is exactly the normalized origin followed by `/healthz`.
- [ ] Each backend starts healthy; health changes only because of an explicit probe or a selected request's transport failure.
- [ ] A selected-request 5xx does not change health; only a later 2xx probe restores an unhealthy backend.
- [ ] The probe is one bodyless `GET /healthz` with injected dial, response-header, and whole-request timeouts; 2xx is healthy and non-2xx, timeout, or transport failure is unhealthy; at most 4,096 response bytes are discarded and the body is always closed; no addresses leak.
- [ ] Round-robin is a cursor over the fixed backend order; selection scans from the cursor to the next healthy backend, chooses it, and advances the cursor to the following physical slot.
- [ ] With an initial cursor and healthy A, B, C, sequential requests select exactly A, B, C and wrap to A.
- [ ] The cursor advances exactly once per successful selection even when the request handled by that selection later fails with a transport failure.
- [ ] The all-unhealthy path returns exact `application/json` 503 with only `error: "service_unavailable"`, no trailing newline or internal detail, no upstream call, and no cursor advance.
- [ ] Concurrent cursor advance and selection are atomic; state and cursor are race-free under the race detector.
- [ ] Transport failure on a selected request marks the selected backend unhealthy and returns 502 without retrying the current request.
- [ ] A later selection skips a backend marked unhealthy.
- [ ] Project 078's corrected exact proxy safety contract is preserved end-to-end.
- [ ] The balancer adds no identity feature; test backends return distinct ordinary bodies and the proxy forwards them.
- [ ] Under a stable healthy set with no probes or transport failures during the batch, aggregate counts differ by at most one after a complete batch; no specific competing order is asserted.
- [ ] Tests use `httptest` for backends and balancer listener; no public network, no fixed public port.
- [ ] All tests pass under the race detector with no sleep synchronization.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions
Add a per-backend latency histogram exposed at shutdown for capacity planning tests. Add a structured access log that records method, path, status, and duration but never request or response bodies.
