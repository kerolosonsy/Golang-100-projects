# Project 065 — Redis Caching Layer

## 1. Project Name and Number

- Project 065, redis_caching_layer.
- This README is a learning guide only.
- You will create every Go source file, the Compose definitions, and every test file yourself in `05-databases/065_redis_caching_layer/`.
- The guide does not provide implementation code.

## 2. Project Idea

Build a cache-aside decorator around the Project 061 User repository using `github.com/redis/go-redis/v9` `v9.21.0` as the only new third-party dependency. Use JSON values, a versioned namespaced key, and a fixed TTL of exactly 5 minutes. Make every failure mode honest: cache hits and misses, repository-first writes, partial success on invalidation failure, and a process-local bypass set for keys whose deletion cannot be confirmed.

## 3. Why This Project Now?

- This follows Projects 064 (database_migrations) and 061 (sqlite_crud) and adds a cache on top of a proven repository foundation.
- Context handling follows Project 041.
- No other project is formally required.

## 4. Prerequisites

- Projects 064, 061, and 041 are required.
- No other project is formally required.
- The regular unit gate is `go test ./...`, which must pass with no Docker, no Redis, no PostgreSQL, no network, and no environment variables.

## 5. What You Must Know Before Starting

- You should know the exact Project 061 User contract (fields, normalization, ordering, clock behavior, typed outcomes).
- You should be comfortable with `context` propagation and cancellation, JSON encoding for stable field sets, TTL semantics owned by Redis, and concurrency-safe in-process data structures.

## 6. Explanation of New Concepts

### Concepts

- redis.Nil miss versus other Redis errors: `redis.Nil` is a documented sentinel meaning the key was not found.
- It is treated as a confirmed miss.
- Any other Redis error on the Get path (transport failure, server error, timeout that is not the request's own context cancellation) is treated as a non-confirmed-miss outage: the decorator falls back to the repository and returns its result, but does not attempt Set during that call because the cache state was not a confirmed miss.
- Best-effort repopulation only occurs after `redis.Nil`, after a corrupt entry whose stale deletion succeeds, or after a successful bypass deletion.

- Cache-aside read: the cache is consulted first.
- A hit returns the decoded value without touching the repository.
- A miss loads from the repository and returns it, and best-effort writes the value into the cache with the TTL.
- The TTL is the only expiry policy; the core does not invent local expiry checks.

- Best-effort fill: writing to the cache after a repository miss is best-effort.
- A Set failure never turns a successful repository read into a failure.
- The repository result is what the caller sees.

- Repository-first write: on update or delete, the repository is written first.
- If the repository fails, the cache is not invalidated because the source of truth did not change.
- If the repository succeeds, the cache key is deleted.

- Invalidation partial success: after a successful repository write, the cache key is deleted.
- If deletion succeeds, normal success is returned.
- If deletion fails, the database change is already committed.
- The decorator returns a typed partial-success or invalidation warning that explicitly carries the committed result where applicable, and adds the key to a concurrency-safe process-local bypass set.

- Local bypass: while a key is in bypass, Get never trusts its cached value.
- It reads the repository, attempts deletion again, and clears bypass only on successful deletion.
- The same Get returns the repository result.
- Other processes may still serve stale data; the README explicitly disclaims multi-process and distributed safety.

- Negative caching is not part of this project.
- A repository not-found during a miss returns not-found and does not write a sentinel into the cache.
- Negative caching is entirely outside the required implementation, not an optional branch.

## 7. Learning Objective

- Implement a cache-aside decorator whose hits, misses, invalidations, fallbacks, partial-success outcomes, and bypass recovery are precisely defined, testable through fakes, and observably distinct.
- Make every Redis call subject to context rules.
- Keep the boundary honest about what it can and cannot guarantee.

## 8. Functional Requirements

1. Pin `github.com/redis/go-redis/v9` `v9.21.0` as the only new third-party dependency. No mocking library is introduced.
2. The cache key is exactly `go-tutorial:v1:user:<positive-decimal-id>`. The TTL passed to the cache boundary is exactly 5 minutes. Real TTL is owned by Redis server time; the core does not maintain a local expiry.
3. Cached JSON contains exactly the public User fields: `id`, `name`, `email`, `created_at`, `updated_at`. There is no wrapper object and no internal field. Timestamps are serialized as UTC RFC3339Nano.
4. Input validation rejects a nonpositive or otherwise invalid ID before any cache or repository access.
5. Read on a hit validates the decoded payload against the full Project 061 public invariants: the requested positive matching ID; a trimmed non-empty name; an email that is exactly the trimmed lowercase normalized form and otherwise valid; timestamps that parse as RFC3339Nano and are in UTC. Empty or unparsable timestamps are corrupt. A structurally valid but semantically wrong cached value (for example, a mismatched ID, a non-normalized email, or a non-UTC timestamp) is treated as corrupt and follows the same delete/bypass/repository path as a syntax-level corruption.
6. Read on a `redis.Nil` miss loads the repository, returns its result, and best-effort Sets the JSON with the 5-minute TTL.
7. Read on a Redis Get outage (any Redis error that is not `redis.Nil` and not the request's own context cancellation/deadline) falls back to the repository. The repository result is returned. No best-effort Set is attempted during this call because the cache state was not a confirmed miss; the next read may produce `redis.Nil` and then the best-effort fill applies.
8. Read on a Redis Set failure (after a successful repository read) returns the repository result; the cache outcome is not surfaced as an error.
9. Read on a Redis cancellation/deadline error that matches the request context propagates the context error immediately and does not fall back to the repository. Context cancellation is not a cache outage.
10. Corrupt cached JSON (syntax-level or semantic-mismatch) is never returned. The decorator attempts to delete the key. If the delete succeeds, it loads the repository and best-effort repopulates from that repository result. If the delete fails, it marks local bypass, loads the repository, and best-effort repopulates only after the stale deletion succeeds.
11. Update and delete write the repository first. If the repository fails, no invalidation is attempted. The repository error is returned. The cache state is unchanged.
12. After a successful repository update or delete, the decorator attempts to delete the key. On successful delete, normal success is returned.
13. After a successful repository update or delete, if the delete fails (any Redis error, including context cancellation), the database change is already committed. The decorator adds the key to a concurrency-safe process-local bypass set and returns a typed partial-success or invalidation warning that explicitly carries the committed result where applicable.
14. While a key is in bypass, Get calls the repository first. If the repository returns a non-not-found error (including a context cancellation/deadline error), the decorator retains bypass and returns that error immediately without making any Redis call. If the repository returns a User, the decorator retries Delete; on Delete success it clears bypass and best-effort Sets the repository result, then returns the User; on Delete failure it retains bypass and returns the repository User without any Set. If the repository returns not-found, the decorator retries Delete; on Delete success it clears bypass and returns not-found without a Set; on Delete failure it retains bypass and returns not-found without a Set. The bypass-aware Get never returns a value read from Redis during that call.
15. Corrupt-entry ordering follows the same exact outcomes: initial Delete, then repository. If initial Delete succeeds and the repository returns a User, best-effort Set the repository result. If the repository returns not-found, return not-found without a Set. If the repository returns any error, return it. If initial Delete fails, mark bypass, read the repository, and never Set during that call; the bypass-aware path takes over on subsequent reads.
16. Context rules across read and write paths. Pre-cancelled context causes zero calls. On a read, the initial Redis Get is checked first: if it returns the request's own cancellation or deadline error, the decorator propagates that error immediately and never reaches the repository. If the initial Get returns `redis.Nil`, the repository is read; a context cancellation/deadline error from that repository call is returned. A best-effort Set belongs only to a confirmed `redis.Nil` miss and to the explicitly defined successful corrupt-entry and successful bypass-recovery paths. If that best-effort Set returns the request context error after the repository User was already obtained, the User remains the result. On update or delete, the repository runs first. If the repository returns a context error before any commit, that error is returned and no cache call occurs. Once the repository has reported success, the subsequent Delete may return any error including a context cancellation/deadline error; the typed partial-success warning is returned because the source-of-truth change has already happened.
17. The typed partial-success outcome conceptually carries the operation kind, a committed flag, the updated User for update (delete has no User), and the wrapped cache cause. Callers must be able to detect partial success and the underlying cause. The README does not give a Go type or signature.
18. The bypass set is concurrency-safe but may grow during a prolonged outage. This limitation is stated honestly. Bounded reconciliation is a possible extension rather than a guaranteed guarantee.
19. The implementation does not claim a DB/Redis transaction, distributed invalidation, stampede protection, or penetration-prevention.
20. The cache key for integration tests keeps the required prefix `go-tutorial:v1:user:`, derives unique positive decimal IDs for the run, records the exact keys it writes, and deletes only the recorded keys. It never flushes or scans unrelated keys.

## 9. Inputs and Outputs

### Interface Contract

- Get inputs: context, ID.
- Get outputs: User or typed not-found, invalid-ID, or repository error.
- Update inputs: context, ID, new name, new email.
- Update outputs: normal success (with the updated User), typed invalid-ID or invalid-input, repository error, or typed partial-success warning carrying the committed User and the wrapped cache cause.
- Delete inputs: context, ID.
- Delete outputs: typed not-found, repository error, normal success, or typed partial-success warning carrying a committed flag and the wrapped cache cause.

## 10. Rules and Edge Cases

- Reject nonpositive IDs before any cache or repository call.
- Distinguish `redis.Nil` from other Redis errors.
- Distinguish context cancellation from outage.
- Never return corrupt cached JSON.
- Never set a negative cache entry on repository not-found.
- Never invalidate before a successful repository write.
- Never claim multi-process or distributed safety.
- The bypass set may grow under prolonged outage.
- Cache contents are never read from Redis while a key is in bypass.

## 11. Project Constraints

- No mocking library.
- No Docker for unit tests.
- Compose material is conceptual only; this README contains no commands or YAML.
- Integration is opt-in.
- The core uses the pinned dependency and only the pinned dependency as a new direct third-party import.

## 12. Design Questions Before Coding

- How does the decorator distinguish `redis.Nil` from other Redis errors without string matching?
- How does it distinguish context cancellation from outage?
- Where is bypass state held and how is it cleared?
- How is the typed partial-success outcome defined conceptually so callers can detect it?
- How is the bypass growth limitation disclosed?
- How does the unit test fake record calls, values, and TTLs without taking a dependency on a mocking library?

## 13. Implementation Milestones

1. Define the cache and repository boundaries with handwritten fakes that record calls, keys, values, and TTLs.
2. Implement key construction, JSON shape, and input validation that rejects nonpositive IDs before any call.
3. Implement the read path: hit, miss with best-effort fill, Get outage fallback, Set failure handling, context propagation, and corrupt-JSON deletion.
4. Implement bypass add, bypass check, bypass clear, and the bypass-aware read path.
5. Implement update and delete with repository-first ordering, normal invalidation success, and partial-success warning with committed flag and wrapped cause.
6. Implement the conceptual typed partial-success outcome with the operation kind, committed flag, updated User (where applicable), and wrapped cause.
7. Add unit tests covering every typed outcome, including concurrency on the same key and on different keys.
8. Add tagged integration tests that exercise real serialization, hit/miss, invalidation, and TTL behavior under the documented key prefix and guard, then clean up only their own keys.

## 14. Verification Cases the Learner Must Write

### Required Cases

Unit tests:
- Exact key construction: a Get, Set, or Delete call uses the literal key `go-tutorial:v1:user:<id>` for positive decimal IDs.
- Hit returns a decoded User whose fields match the JSON.
- Hit validation rejects a cached JSON whose ID is nonpositive or mismatched; the decorator follows the corrupt-JSON path.
- Hit validation rejects a cached JSON whose email is not normalized.
- Miss (redis.Nil) calls the repository and best-effort Sets the JSON with the 5-minute TTL.
- Miss returns repository not-found without any Set.
- Set failure after a repository miss does not change the returned result.
- Get outage (a non-`redis.Nil` Redis error that is not the request's own context cancellation/deadline) falls back to the repository, returns its result, and makes no Set during this call.
- Get outage that is a context cancellation/deadline matching the request context propagates immediately and does not call the repository.
- Corrupt JSON syntax: the decorator attempts Delete; on Delete success it loads the repository and best-effort repopulates from that result; on Delete failure it marks bypass, loads the repository, and best-effort repopulates only after the stale deletion succeeds.
- Corrupt JSON with mismatched ID follows the same path.
- Repository error during a miss is returned to the caller.
- Successful invalidation on update and on delete removes the key and returns normal success.
- Repository failure on update or delete (including a context cancellation/deadline returned before any commit) returns the repository error and makes no cache call.
- Write ordering for update and delete: pre-cancelled context prevents the repository call; a context error before commit returns that error with no cache call; once the repository reports success, any subsequent Delete failure including context cancellation/deadline becomes the typed partial-success warning.
- Partial-success warning on update: repository commits, Delete fails; the typed outcome carries a committed flag, the updated User, and the wrapped cache cause; the key is in bypass.
- Partial-success warning on delete: repository commits, Delete fails; the typed outcome carries a committed flag and the wrapped cache cause; the key is in bypass.
- Read cancellation from the initial Redis Get that matches the request context propagates immediately and never reaches the repository.
- Read cancellation from a best-effort Set that belongs only to a confirmed `redis.Nil` miss (or a defined successful corrupt/bypass recovery path) does not replace the already-obtained repository User; the User remains the result.
- Bypass-aware Get calls the repository first; on a non-not-found repository error it retains bypass and returns that error with no Redis call; on a User it retries Delete, clears bypass and best-effort Sets on Delete success, or retains bypass and returns the User without a Set on Delete failure; on not-found it retries Delete, clears bypass without a Set on success, or retains bypass and returns not-found without a Set on failure.
- Bypass-aware Get never reads from Redis while a key is in bypass.
- Corrupt-entry ordering: initial Delete, then repository; if initial Delete succeeds, repository is loaded and best-effort Set only when the repository returned a User; if initial Delete fails, bypass is marked and no Set occurs.
- Invalid ID (zero, negative) is rejected before any cache or repository call.
- Pre-cancelled context causes zero cache and repository calls.
- Concurrent access on the same key under `-race` is clean.
- Concurrent access on different keys under `-race` is clean.
- Bypass set growth under a simulated prolonged outage is observable and stated as a limitation.
- TTL passed to the cache boundary is exactly 5 minutes.

Tagged integration tests:
- Activation: without the `integration` build tag the integration file is not compiled. With the tag present, an absent connection string and absent guard together cause a clear skip. If either runtime value is supplied, a missing or unparseable connection string, a missing guard, or a guard whose value is not exactly `I_UNDERSTAND_TEST_REDIS_WILL_BE_MODIFIED=yes` fails closed before any Redis connection. Connection-string credentials are never logged. Unique positive decimal IDs are derived for the run in a way that resists collision with concurrent or repeated runs.
- The integration suite records the exact keys it writes under the required `go-tutorial:v1:user:` prefix, using unique positive decimal IDs for the run.
- Real Set followed by real Get returns the cached JSON.
- Real TTL is positive and no greater than 5 minutes, verified by reading the TTL value, without using sleeps.
- Real Delete on update and on delete removes the key.
- The integration suite cleans up only the recorded keys. It never flushes the database and never scans or deletes unrelated keys.

## 15. Common Mistakes to Watch For

- Treating context cancellation as an outage.
- Returning corrupt cached JSON.
- Invalidating before writing the repository.
- Setting a negative cache entry on repository not-found.
- Returning a value read from Redis while the key is in bypass.
- Attempting Set during a Get outage call when the cache state was not a confirmed miss.
- Reading from Redis during a corrupt-entry or bypass recovery call.
- Flushing the Redis test database.
- Sleeping to assert TTL.
- Using a mocking library.
- Claiming distributed or multi-process safety.
- Hiding the bypass growth limitation.
- Logging credentials or addresses in test logs.
- Forgetting to clear bypass after a successful Delete.

## 16. Topics and References for Study

- The go-redis `v9.21.0` documentation, especially `redis.Nil`, context propagation, `Set` with TTL, and `Del`.
- Cache-aside patterns and their honest limits.
- JSON encoding for stable field sets and versioned keys.
- Concurrency-safe in-process sets.
- Process-local versus distributed coordination.
- Safe integration activation with build tags and explicit guards.

## 17. Self-Assessment Questions

1. Why must the repository write happen before invalidation?
2. Why is corrupt JSON deleted rather than repaired?
3. Why is negative caching entirely outside this project?
4. Why is local bypass only process-local?
5. Why is context cancellation distinct from outage?
6. Why does a Get outage not attempt Set during that call?
7. What is honest about this implementation?
8. Why is the bypass set's unbounded growth under prolonged outage stated as a limitation rather than hidden?
9. Why does the integration suite record and delete only its own keys?
10. Why does a miss after Set failure not retry Set inside the same call?

## 18. Definition of Completion

- [ ] `go test ./...` passes with no Docker, network, PostgreSQL, Redis, or environment variables.
- [ ] `go test -race ./...` passes.
- [ ] Every required typed outcome has a unit test.
- [ ] The exact key `go-tutorial:v1:user:<id>` is asserted.
- [ ] The 5-minute TTL is asserted on every Set call.
- [ ] Corrupt JSON syntax and semantic-mismatch cases are covered.
- [ ] Repository-first ordering is asserted for update and delete.
- [ ] No cache invalidation occurs when the repository write fails.
- [ ] Partial-success warning carries the committed flag and wrapped cache cause and is detectable by callers.
- [ ] Bypass add, bypass-aware read, and bypass clear are tested.
- [ ] Context cancellation is propagated immediately and never treated as an outage.
- [ ] No mocking library is introduced.
- [ ] Integration is tagged and guarded; absent activation inputs skip clearly; any supplied-but-invalid activation input fails closed before any Redis connection; it cleans up only its own keys and never flushes the database.
- [ ] No implementation code, full function signatures, typed return signatures, SQL snippets, YAML, Compose, or shell commands appear in this README.

## 19. Optional Extensions

- Add a separately documented cache warm-up experiment under the same key prefix.
- Add a separately tagged observability integration that records hit/miss counters without weakening the unit gate.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 064 — Database Migrations](../../05-databases/064_database_migrations/README.md#20-prerequisite-based-documentation-guide), [Project 061 — SQLite CRUD](../../05-databases/061_sqlite_crud/README.md#20-prerequisite-based-documentation-guide), [Project 041 — Context Timeout Example](../../03-concurrency/041_context_timeout_example/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`github.com/redis/go-redis/v9`](https://pkg.go.dev/github.com/redis/go-redis/v9).
- **Standards and concept references:** [Redis Go client guide](https://redis.io/docs/latest/develop/clients/go/).

### Project-specific learning focus

- **Learn now:** cache-aside reads, cache misses versus failures, versioned key namespaces, TTL policy, invalidation order, stampedes, stale data, and integration-test isolation.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
