# Project 070 — Read/Write Splitting

## 1. Project Name and Number

- Project 070, read_write_splitting.
- This README is a learning guide only.
- You will create every source and test file yourself in `05-databases/070_read_write_splitting/`.
- This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea

Build a router over primary and replica interfaces. Every write goes once to primary. Ordinary reads prefer replica and fall back once to primary only for replica availability failure. Caller-carried positive version tokens provide explicit read-after-write routing without hidden session state.

## 3. Why This Project Now?

- Project 070 follows Project 069 in the catalog and applies its deterministic decision-making discipline to distributed-read routing, where availability, staleness, domain outcomes, and observer side effects must remain distinct.
- Projects 061 and 041 provide required persistence and service foundations.

## 4. Prerequisites

- Required prerequisites: Projects 069, 061, and 041.
- Optional review: Project 066 for transaction and idempotency contrasts; it is not required.
- All required tests use fakes only and need no database, replication system, Docker, network, or environment variables.

## 5. What You Must Know Before Starting

- Know interfaces, contexts, typed domain and availability errors, concurrency-safe observers, atomic or synchronized version boundaries, call-order assertions, caller-carried consistency metadata, and race testing.

## 6. Explanation of New Concepts

### Concepts

- Primary handles all writes.
- A successful write returns a positive version.
- The required version carried into reads is a signed 64-bit integer in this internal contract: negative is typed invalid input before any backend or boundary call, zero means absent or ordinary, positive means tokened.
- Serialization at an outer transport is out of scope.
- The router keeps no hidden per-session sticky state.
- Therefore one caller's write cannot affect another caller unless that second caller explicitly passes a positive token.

- Ordinary read goes directly to replica.
- Replica availability error causes exactly one fallback to primary and records that fallback.
- Replica domain not-found is returned without fallback.
- Other domain errors are also returned according to their classification rather than being disguised as availability.

- A tokened read first asks an injected thread-safe replica-version boundary whether the replica has applied at least the required positive version.
- The boundary returns a synchronized nonnegative signed 64-bit position or a typed state-unavailable outcome.
- If the boundary is unavailable, route directly to primary without replica data read.
- If the replica position is below required, route directly to primary without replica data read.
- Neither direct-primary route is an availability-triggered fallback and neither is recorded as such.
- If the replica position is at or above required and the boundary was available, replica is eligible.
- A boundary that returns a negative position is a backend-contract error: route directly to primary without replica data read, do not record availability fallback.

- An eligible replica read follows ordinary error rules: availability error falls back once to primary; domain not-found returns without fallback.
- There is no time-only catch-up promise.
- Version comparison is the consistency boundary.

- Context precedence is exact.
- A pre-cancelled context causes zero boundary, zero backend, and zero observer calls.
- If caller context becomes cancelled or deadline-exceeded and that cause is returned or discoverable from the boundary, replica, or primary, the router propagates that context error and does not classify it as replica availability or perform fallback.
- Only a typed replica availability error while the caller context remains active triggers fallback.

- Fallback recording uses an injected concurrency-safe observer.
- Record only actual availability-triggered replica-to-primary fallback, with enough routing context to distinguish ordinary and token-eligible reads.
- Observer failure must never replace, hide, or modify the read result or primary fallback result.

- Writes are never retried automatically because failure may have occurred after mutation.
- The router never sends writes to replica.
- A successful write must provide a positive version; if it reports success with zero or negative version, the router returns a typed backend-contract error with no usable token and does not retry.

## 7. Learning Objective

- Implement an explicit, stateless read/write routing matrix with honest stale-read semantics, caller-carried read-after-write consistency, exact fallback classification, observable routing, deterministic call order, and race-safe concurrent sessions.

## 8. Functional Requirements

1. Primary and replica are separate injected interfaces.
2. Every write calls primary exactly once and never calls replica.
3. No write is automatically retried, including on availability or context failure.
4. Successful primary write returns a positive version. Failed write returns no usable token.
5. A primary that reports success with a zero or negative version is a typed backend-contract error with no usable token; the router does not invent a token and does not retry.
6. Router stores no per-session or global sticky version state.
7. The required version carried into reads is a signed 64-bit integer: negative is typed invalid input before any boundary or backend call, zero means absent or ordinary, positive means tokened.
8. Ordinary read calls replica directly.
9. Ordinary replica availability failure falls back exactly once to primary.
10. Ordinary replica domain not-found returns not-found without primary fallback.
11. Tokened read asks the replica-version boundary before any replica data read.
12. Boundary unavailable routes directly to primary without replica data read. Direct-primary consistency routes are not recorded as availability fallback.
13. Replica version below required routes directly to primary without replica data read. Direct-primary consistency routes are not recorded as availability fallback.
14. A boundary that returns a negative position is a typed backend-contract error; the router routes directly to primary without replica data read and does not record availability fallback.
15. Replica version at or above required makes replica eligible.
16. Eligible replica availability failure falls back exactly once to primary.
17. Eligible replica domain not-found returns not-found without fallback.
18. No elapsed-time threshold promises replica catch-up.
19. Availability-triggered fallback is recorded through an injected concurrency-safe observer.
20. Observer failure never replaces or changes the read result.
21. Pre-cancelled context causes zero boundary, backend, and observer calls.
22. Caller cancellation or deadline context returned or discoverable from boundary, replica, or primary is propagated and is not classified as replica availability or used to trigger fallback. Only a typed replica availability error while the caller context remains active triggers fallback.
23. Unit tests use fakes only and run under the race detector.

## 9. Inputs and Outputs

### Interface Contract

- Write input is context plus domain mutation data; output is the primary result and a positive version on success, the primary error with no usable token on failure, or a typed backend-contract error with no usable token if success is reported with a non-positive version.
- Read input is context, lookup key, and optional required version.
- Required version is a signed 64-bit integer: negative is typed invalid input before any boundary or backend call, zero means absent or ordinary, positive means tokened.
- Output is domain value, typed not-found, availability or primary error, plus observable routing effects.
- The router does not return or retain hidden stickiness.

- Example behavior: caller A writes and receives version 42.
- A read carrying required version 42 asks replica version first.
- At replica position 41 it routes directly to primary; that direct-primary consistency route is not recorded as availability fallback.
- At replica position 42 it reads replica.
- Caller B omits the token, so required version is zero and the read is ordinary.
- Caller A's write created no hidden session state and never reaches caller B unless caller B passes the token explicitly.

## 10. Rules and Edge Cases

- Reject negative required versions as typed invalid input before any boundary or backend call.
- Treat zero required version as absent or ordinary.
- Never route writes to replica.
- Never automatically retry writes.
- Ask version boundary first for tokened reads.
- Do not call replica data read when boundary is unavailable, when replica position is below required, or when the boundary returns a negative position; route directly to primary in those cases.
- Direct-primary consistency routes from unavailable, behind, or backend-contract boundary outcomes are not recorded as availability fallbacks.
- Fall back only on typed replica availability errors while the caller context remains active.
- Return replica domain not-found without fallback when replica is eligible or read is ordinary.
- A primary that reports success with a non-positive version is a typed backend-contract error with no usable token.
- Pre-cancelled context causes zero calls at boundary, backend, and observer.
- Cancellation or deadline returned from boundary, replica, or primary is propagated and is not classified as replica availability.
- Ignore observer failure for result selection while making it test-observable.

## 11. Project Constraints

- This project models routing only.
- Actual replication, failover orchestration, transactions across stores, replication protocol, lag measurement implementation, hidden sessions, automatic write retry, and time-based catch-up promises are out of scope.
- No integration environment is required or claimed.
- All required verification uses fakes and no external services.

## 12. Design Questions Before Coding

- Which error classifications represent availability versus domain results versus backend-contract violations versus context cancellation?
- How does a tokened call prove boundary-before-data-read ordering?
- Which direct-primary routes are availability-fallback observations, and which are consistency routes that must not be recorded?
- How is a non-positive version from a successful write distinguished from ordinary primary failure?
- How is observer failure exposed for tests without changing the read result?
- How are version reads thread-safe and how is a boundary negative-position outcome handled?
- How do fake call logs prove session isolation and exactly-once behavior?

## 13. Implementation Milestones

1. Define primary, replica, version-boundary, observer, signed 64-bit required version, typed error, and routing-result contracts.
2. Implement write routing with exactly one primary call, positive-version validation, and typed backend-contract error on non-positive version from reported success.
3. Implement ordinary replica read, availability fallback, and domain not-found behavior.
4. Implement required version validation and boundary-first direct-primary routing for unavailable, behind, and boundary-contract-violation states, none of which records availability fallback.
5. Implement eligible-replica behavior and exactly-once availability fallback while the caller context remains active.
6. Add concurrency-safe observer calls whose failures cannot replace results.
7. Implement context precedence: pre-cancelled context causes zero calls; cancellation or deadline returned or discoverable from boundary, replica, or primary is propagated and never triggers fallback.
8. Complete routing-matrix, order, call-count, cancellation, boundary-contract-violation, concurrent-session, and race tests.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Successful write calls primary once, replica never, and returns a positive version.
- Failed write calls primary once, replica never, performs no retry, and returns no usable token.
- Primary that reports success with a zero version returns typed backend-contract error with no usable token and does not retry.
- Primary that reports success with a negative version returns typed backend-contract error with no usable token and does not retry.
- Ordinary read calls replica first and returns its success.
- Ordinary replica availability error calls primary exactly once and records one fallback.
- Ordinary replica domain not-found returns not-found with no primary call and no fallback record.
- Negative required version fails before any boundary or backend call; zero behaves as absent or ordinary.
- Tokened read calls version boundary before any backend data read.
- Boundary unavailable routes directly to primary with no replica data call and no availability-fallback record.
- Replica version below required routes directly to primary with no replica data call and no availability-fallback record.
- Boundary returning a negative position is a typed backend-contract error that routes directly to primary with no replica data call and no availability-fallback record.
- Replica version equal to or above required makes replica eligible.
- Eligible replica success returns without primary.
- Eligible replica availability error falls back once to primary and records fallback.
- Eligible replica domain not-found returns without fallback.
- Primary success, not-found, availability failure, domain failure, backend-contract failure, and context failure after fallback are returned exactly.
- Observer receives only actual availability-triggered fallback events.
- Observer failure does not replace replica or primary read result.
- Pre-cancelled context causes no backend, boundary, or observer call when checked at entry.
- Caller cancellation discoverable from boundary is propagated and is not classified as replica availability.
- Caller cancellation discoverable from replica is propagated and is not converted to generic availability.
- Caller cancellation discoverable from primary after fallback is propagated exactly.
- Call logs prove exact order and counts for every routing branch, including that direct-primary consistency routes are not recorded as availability fallback.
- Caller A's tokened read and caller B's ordinary read remain isolated under concurrency.
- Concurrent boundary updates and reads pass the race detector.
- No test uses elapsed time as proof of replica catch-up.

## 15. Common Mistakes to Watch For

- Routing a write to replica, retrying uncertain writes, retaining a global sticky flag, treating a negative required version as a malformed token, treating zero as a positive required version, inventing a version from a primary that reported success with a non-positive version, reading replica before checking boundary state, falling back on domain not-found, treating every replica error as availability, treating caller cancellation as replica availability, treating caller cancellation discoverable from boundary or primary as availability, promising time-based catch-up, recording direct-primary consistency routing as availability fallback, treating a negative boundary position as a sync-only condition without record bookkeeping, letting observer failure replace data, or using unsynchronized version state.

## 16. Topics and References for Study

- Study primary-replica architectures, read-after-write consistency, monotonic version or log-position concepts, stale reads, typed error classification, context propagation, and idempotency limits for writes.
- Review Go interfaces, synchronization primitives, race detection, call-recording fakes, and observer patterns whose telemetry failure does not affect business outcomes.

## 17. Self-Assessment Questions

1. Why is the required version caller-carried rather than hidden, and why is it signed 64-bit?
2. Why does a tokened read ask the version boundary first?
3. Why does unavailable version state choose primary?
4. Why is a negative required version typed invalid input rather than treated as a malformed token?
5. Why is a non-positive version from a successful write a backend-contract error, not a retry?
6. Why is replica not-found not an availability failure?
7. Which routes count as availability fallback and which are direct-primary consistency routes that must not be recorded?
8. Why are writes not retried?
9. Why does caller cancellation never trigger fallback?
10. How can concurrent tests prove one caller's token never affects another?

## 18. Definition of Completion

- [ ] Every write goes to primary exactly once and is never automatically retried.
- [ ] Ordinary read, availability fallback, no-fallback domain not-found, and direct-primary consistency routes are tested with explicit record-bookkeeping checks.
- [ ] Signed 64-bit required version: negative typed invalid input, zero absent or ordinary, positive tokened. Stateless per-session storage.
- [ ] Primary reporting success with a non-positive version returns typed backend-contract error with no usable token.
- [ ] Tokened reads perform boundary-first routing and never read an ineligible replica; boundary negative position is backend-contract error and routes directly to primary without replica data read.
- [ ] At-or-above replica eligibility follows exact availability-versus-domain rules and only triggers fallback when caller context remains active.
- [ ] Caller cancellation discoverable from boundary, replica, or primary is propagated and never triggers fallback.
- [ ] Observer is concurrency-safe and its failure never changes read results; pre-cancelled context causes zero observer calls.
- [ ] Full routing matrix has exact call counts and order assertions.
- [ ] Concurrent sessions remain isolated and race tests pass.
- [ ] Actual replication, failover, transactions, and time-based catch-up claims remain out of scope.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions

- Add weighted routing across several eligible replicas with the same consistency boundary.
- Add routing metrics for direct-primary consistency decisions separately from availability fallbacks.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 069 — Search Text Indexing](../../05-databases/069_search_text_indexing/README.md#20-prerequisite-based-documentation-guide), [Project 061 — SQLite CRUD](../../05-databases/061_sqlite_crud/README.md#20-prerequisite-based-documentation-guide), [Project 041 — Context Timeout Example](../../03-concurrency/041_context_timeout_example/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **Standards and concept references:** [PostgreSQL warm standby and replication](https://www.postgresql.org/docs/current/warm-standby.html).

### Project-specific learning focus

- **Learn now:** primary-replica routing, read-after-write consistency, monotonic positions, stale reads, availability fallback, typed errors, idempotent writes, and non-blocking telemetry.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
