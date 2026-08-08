# Project 043 — Thread-Safe Cache

## 1. Project Name and Number

- Project 043 — Thread-Safe Cache.
- The project is a learning lab for a generic, in-memory, comparable-key, thread-safe cache with a small, honest interface and a zero-value-usable starting state.
- The cache must distinguish a missing key from a stored zero value, support controlled concurrent readers and writers, and never reveal its internal map or mutex.

## 2. Project Idea

The cache maps comparable keys to arbitrary values. Read operations consult one read-lock at a time so many readers may share access. Write operations consult one exclusive lock so reads and writes do not conflict. The cache exposes a small set of operations: set a value under a key, look up the key and report whether the value is present, remove the key and report whether anything was removed, report the count of stored entries, and clear every entry while leaving the cache reusable.

The zero value of the cache type must be usable without initialization. The internal map may initially be nil; get, delete, len, and clear are safe on that state, and the first set lazily initializes the map while holding the write lock. The cache exposes nothing that would let a caller reach inside the map or the mutex.

## 3. Why This Project Now?

- Project 042 was the immediate predecessor and trained synchronous-callback delivery outside a lock.
- Project 043 turns that pattern into a long-lived shared structure where the lock is held briefly around each operation rather than around a callback loop.
- The goroutine and barrier thinking from Project 031 is reused here for the concurrent test suite.
- Project 045's atomic counter shares a verification goal, exact totals under contention, and the same race-detector discipline, which is why these two projects sit next to each other.

## 4. Prerequisites

- Complete Projects 031 and 042 first.
- Project 031 supplies goroutines, barriers, and wait groups reused for concurrent cache tests.
- Project 042 supplies the snapshot-outside-lock reasoning that this project adapts into lock-around-each-operation.
- You must already be comfortable with generics, comparable constraints, the read-write mutex, and barrier-driven concurrent tests with no sleep.
- Earlier projects that introduce mutex ownership are useful review but are not required prerequisites.

## 5. What You Must Know Before Starting

- Know that a type containing a mutex must not be copied after first use, because copying the mutex duplicates its internal state and defeats the synchronization.
- Documentation alone is not enough; the recommended discipline is to use a pointer to such a value anywhere it crosses a function boundary.

- Know that a read-write mutex allows many readers concurrently but allows at most one writer at a time and excludes readers during a write.
- Long-running write operations block readers, so writers must do as little as possible while holding the lock.
- The read-write mutex is the required learning primitive here, but the choice is not a universal performance claim.
- Under a workload dominated by writes, a plain mutex can be cheaper because the read-write mutex pays extra bookkeeping to support shared readers.
- The correct choice depends on the workload, and the project adopts the read-write mutex as the learning primitive without claiming it is universally faster.

- Know that a stored zero value is indistinguishable from a missing key only if the cache never carries a present-marker.
- The cache uses the standard map lookup result to distinguish: a lookup on the map itself returns the present flag along with the value, so the cache can return a value and a present flag without inventing a separate boolean field inside the map.
- Get reports the value and the present flag, never only the value.

- Know that the cache interface must be honest about reference-containing values.
- Returning a value by value gives the caller a copy of the value at the moment of return, but if that value contains a pointer, slice header, map header, channel, interface, or function, then the copy still refers to caller-mutable data.
- The cache documents this without trying to deep-clone caller values.

- Know that the zero value of the cache type must be usable.
- The internal map may start out nil; the first set lazily initializes the map while holding the write lock.
- Get, delete, len, and clear are safe on the nil-map state.

## 6. Explanation of New Concepts

### Concepts

- A generic type parameterizes the key and value types.
- The key type is constrained to comparable so the cache can use it as a map key.
- The value type is unconstrained because the cache never depends on value-type semantics beyond storage and retrieval.

- The map-lookup presence result is the standard technique for distinguishing absence from a stored zero value.
- The map lookup itself returns the stored value and a present flag indicating whether the key was present.
- The cache uses that result directly, so a stored zero value is reported as present and a missing key is reported as absent without any present-marker field of the cache's own invention.

- The lazy-initialize posture is the cache's answer to zero-value usability.
- The internal map may be nil when the cache is created.
- Get, delete, len, and clear must work on that nil-map state.
- The first set detects that the map is nil and initializes it while holding the write lock so subsequent operations see a usable map.

- The lock-around-each-operation pattern is the cache's chosen posture.
- Set, get, delete, len, and clear each acquire the appropriate lock, perform their minimal work, release the lock, and return.
- The lock is never held across any external call, because the cache does not invoke user callbacks and does not do input or output.

- The reusable-after-clear posture means clearing the cache leaves the same cache value valid for further operations.
- The cache does not allocate a new map; it clears the existing one and keeps its existing mutex.

## 7. Learning Objective

- By completion, you can design a generic thread-safe cache that uses a read-write mutex to share readers and serialize writers.
- You can distinguish a missing key from a stored zero value using the map lookup presence result.
- You can hand the cache across function boundaries only by pointer, document the no-copy-after-first-use constraint, and refuse to copy the type in code review.
- You can write deterministic concurrent tests using disjoint keys or controlled phases and barriers, with no sleep and no dependence on which concurrent write happens last.

## 8. Functional Requirements

1. The cache is generic over a comparable key type and an unconstrained value type.
2. The zero value of the cache type is usable. Get, delete, len, and clear are safe on a freshly created cache whose internal map is nil. The first set lazily initializes the map while holding the write lock.
3. Set writes the value under the key regardless of whether an entry already exists; an existing entry is overwritten.
4. Get returns the stored value and a present flag; the present flag is true exactly when an entry exists.
5. Get uses the map lookup presence result to distinguish a missing key from a stored zero value of the value type; the cache does not invent a separate boolean field inside the map.
6. Delete removes the entry and returns a flag indicating whether an entry existed.
7. Len returns the number of stored entries at the moment of the call.
8. Clear removes every entry and leaves the cache reusable; the same cache value continues to work after a clear.
9. The internal map and the internal mutex are not exposed by the cache's public surface.
10. Values returned by get are returned by value; the cache documents honestly that reference-containing values can still refer to caller-owned mutable data.
11. Reads may share the read-lock; writes acquire the exclusive lock. The read-write mutex is the required learning primitive; the project does not claim it is universally faster than a plain mutex and explains the workload and overhead tradeoffs.
12. The cache does not invoke user callbacks, does not take a value-producing function on miss, does not persist, does not evict by time, does not distribute, and does not provide a close or stop operation.
13. The cache type must not be copied after first use because it contains a mutex; ownership crosses function boundaries only by pointer.
14. Tests cover zero-value usability, set and overwrite, zero-value stored versus missing, delete missing, delete present, len and clear with reuse, several key and value types, controlled concurrent readers and writers using disjoint keys or controlled phases, invariant checks after contention, and the race detector without sleep. Tests do not assert which concurrent write was last unless the test establishes that ordering explicitly.

## 9. Inputs and Outputs

### Interface Contract

- The input is a key, a value, an empty argument list, or a clear command.
- Set takes a key and a value and stores that pair.
- Get takes a key and returns the stored value plus a present flag.
- Delete takes a key and returns a present flag indicating whether anything was removed.
- Len takes no arguments and returns the current count.
- Clear takes no arguments and leaves the cache empty.

- The output of get is the stored value plus a present flag.
- The output of delete is a present flag indicating whether anything was removed.
- The output of len is the current size.
- The output of set and clear is the cache in its new state.
- No operation produces an error in normal use; misuse is prevented by the type system or by the documented contract.

- Text-only example: storing a zero value under a key and then asking the cache for that key returns the zero value and a present flag of true, while asking the cache for an unrelated key returns the zero value and a present flag of false.

## 10. Rules and Edge Cases

- Setting under a key that already has an entry overwrites the existing value; the count is unchanged.
- Deleting a key that has no entry returns a present flag of false and changes nothing.
- Deleting a key that has an entry returns a present flag of true and reduces the count by one.
- Clearing the cache resets the count to zero and leaves the cache ready for further operations.

- A fresh cache value exposes a nil internal map.
- Get on a missing key returns the zero value of the value type and a present flag of false.
- Delete on a missing key returns false.
- Len on a fresh cache returns zero.
- Clear on a fresh cache is a no-op.
- The first set on a fresh cache initializes the map while holding the write lock.

- The cache does not use any eviction policy.
- The cache does not provide a way to ask whether storage is full.
- The cache does not provide a way to ask when an entry was stored.
- Adding any of those behaviors would require extending the interface, and the project refuses to add them now.

- Concurrent reads may overlap.
- A write excludes all reads and other writes.
- A read that begins after a write completes observes the new state.
- The cache does not guarantee any order between a write and a later concurrent reader, only that the reader observes a coherent state.

- The cache documents its no-copy-after-first-use constraint in the public documentation for the type.
- Code review treats any cached type passed by value across a function boundary after first use as a bug.
- The cache does not silently defend against misuse, because defense would require runtime checks at every method call and break the simple interface.

- If the cache holds values that contain pointers or other reference parts, the returned value shares the caller's underlying data.
- The cache cannot tell whether a value is safe to share; the caller owns that decision.
- The documentation states this without trying to deep-clone caller values.

## 11. Project Constraints

- Use only the Go standard library.
- Use a read-write mutex to synchronize access to the internal map.
- Do not expose the internal map or mutex on the cache's public surface.
- Do not invoke user callbacks from the cache.
- Do not add eviction, persistence, distribution, loading callbacks, or a close primitive.
- Do not copy the cache type after first use anywhere in the public surface.
- The completed code must pass the race detector.

## 12. Design Questions Before Coding

- How does the cache use the map lookup presence result to distinguish a stored zero value from a missing key without changing the public value type?
- How is the read-write lock scoped so reads share and writes exclude but neither holds the lock across an external call?
- How does the cache document and enforce the no-copy-after-first-use constraint in a way that the type system supports and the public surface reflects?
- How is the zero value usable as a cache without a constructor, with the map starting nil and the first set lazily initializing it?
- How does the test design use disjoint keys or controlled phases so concurrent readers and writers do not produce a value that depends on which write happens last?
- How does the cache describe what returning a value by value does and does not guarantee?
- What workload characteristics make the read-write mutex a reasonable learning choice, and where would a plain mutex be cheaper?

## 13. Implementation Milestones

1. Define the cache type with key and value type parameters, the comparable constraint on the key, and the internal map and read-write mutex.
2. Implement set with lazy initialization of the internal map under the write lock when the map is nil.
3. Implement get, delete, len, and clear that are safe on the nil-map state, using the map lookup presence result for get.
4. Use the write lock for set, delete, and clear; use the read lock for get and len. Each operation holds the lock for the minimum required scope.
5. Document the no-copy-after-first-use constraint on the cache type and ensure every method uses pointer receivers.
6. Add tests covering zero-value usability, set and overwrite, zero-value stored versus missing, delete missing and present, len and clear with reuse, and several key and value types.
7. Add controlled concurrent reader-and-writer tests using disjoint keys or controlled phases aligned by barriers and wait groups; no sleep is used and no test asserts an ordering that was not established by the test.
8. Add invariant checks after contention, such as len matching the number of stored entries for disjoint key sets.
9. Run the full package under the race detector and correct every issue reported.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Test the zero value: a fresh cache value supports get, delete, len, and clear without any initialization, and the first set lazily initializes the map and then works.
- Test set and overwrite: set a value, get it back, set a different value under the same key, get the new value.
- Test zero-value stored versus missing: set a zero value of the value type under a key, get returns the zero value and a present flag of true; get for an unrelated key returns the zero value and a present flag of false.

- Test delete missing: deleting a key with no entry returns false.
- Test delete present: deleting a key with an entry returns true.
- Test len and clear with reuse: set several entries, len matches, clear, len is zero, set again, len matches again.
- Test several key and value types: integer keys with string values, string keys with integer values, struct keys with struct values, and any other combination that exercises the comparable constraint.

- Test controlled concurrent readers and writers using disjoint keys: many goroutines perform a mix of read and write operations on disjoint key sets after a start barrier, all complete before the test continues, and the race detector reports no race.
- The final key set and values are known because the test partitioned the keys.
- Test controlled concurrent writers with a controlled final set: many goroutines set values on disjoint keys, then after all have joined a single controlled final set establishes the deterministic value for each key, and the test asserts that final value.
- Do not assert which concurrent write was last unless the test explicitly establishes that ordering.

- Do not introduce sleep into the tests.
- Use barriers to align goroutine starts and wait groups to await completion.

## 15. Common Mistakes to Watch For

- Returning only the value from get hides the difference between a missing key and a stored zero value of the value type.
- Inventing a separate boolean field inside the map is unnecessary when the map lookup already returns a present flag.
- Exposing the internal map or mutex on the cache's public surface invites callers to touch state without holding the lock.
- Holding the read-write lock across a user callback or external call blocks other goroutines and is a bottleneck even if no panic happens.
- Copying a cache type after first use duplicates the mutex and breaks synchronization.
- Calling sleep in tests to give goroutines a chance to race introduces flaky behavior that masks real bugs.
- Asserting the last concurrent write as if the test had decided it produces a flaky test that depends on scheduling.
- Treating the cache as having a close or stop primitive when the project has none is a feature-creep mistake.
- Confusing the read-write mutex with a universal performance win ignores the workload and overhead tradeoff.

## 16. Topics and References for Study

- Study the standard library documentation for the sync package, in particular the read-write mutex and the workload characteristics that influence the choice between read-write and plain mutex.
- Study the standard library documentation for generics and the comparable type constraint.
- Read Go's standard explanations of the map lookup presence result and the no-copy-after-first-use rule for mutex-containing types.
- Compare this project with Project 042's snapshot-outside-lock reasoning, the synchronization vocabulary of Project 040, and the read-write versus plain mutex tradeoffs.
- Project 045's atomic counter shares the verification goal and uses similar disjoint-key partitioning for deterministic concurrent tests.

## 17. Self-Assessment Questions

1. Why does the cache use the map lookup presence result rather than a separate boolean field inside the map?
2. Why is a read-write mutex the natural learning choice for a cache, and why is it not a universal performance claim?
3. Why must the cache type not be copied after first use, and how is that constraint expressed in code rather than documentation alone?
4. Why is the zero value usable as a cache without a constructor, and how does the first set lazily initialize the map?
5. What does returning a value by value guarantee, and what does it not guarantee?
6. Why does the cache refuse to add eviction, persistence, distribution, loading callbacks, or a close primitive?
7. How does the test design use disjoint keys or controlled phases so concurrent readers and writers do not produce a value that depends on which write happens last?

## 18. Definition of Completion

- [ ] The cache exposes set, get, delete, len, and clear as its entire surface.
- [ ] The zero value of the cache type is usable; get, delete, len, and clear are safe on a freshly created cache whose internal map is nil, and the first set lazily initializes the map while holding the write lock.
- [ ] Get uses the map lookup presence result to distinguish a stored zero value from a missing key.
- [ ] Delete reports whether anything was removed.
- [ ] Clear leaves the cache reusable.
- [ ] The internal map and mutex are not exposed.
- [ ] The cache type is not copied after first use, and the public surface documents and supports that rule.
- [ ] The read-write mutex is the required learning primitive, with the workload and overhead tradeoffs documented.
- [ ] Tests cover zero-value usability, set and overwrite, zero-value stored versus missing, delete missing and present, len and clear with reuse, several key and value types, controlled concurrent readers and writers using disjoint keys or controlled phases, invariant checks after contention, and the race detector.
- [ ] Tests do not use sleep and do not assert which concurrent write was last unless the test establishes that ordering explicitly.

## 19. Optional Extensions

- Add a small teaching note that walks through, in plain prose, every internal access to the map and explains why each one holds the lock for the minimum required scope.
- Add a property-style check that randomizes many concurrent operations on disjoint keys and confirms the cache's invariants hold in every run.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 042 — Pub-Sub Event Bus](../../03-concurrency/042_pub_sub_event_bus/README.md#20-prerequisite-based-documentation-guide), [Project 031 — Concurrent Timer](../../03-concurrency/031_concurrent_timer/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- None. This project applies already introduced APIs, standards, and testing practices in a new combination.

### Project-specific learning focus

- **Learn now:** read-write lock tradeoffs, copy-in and copy-out ownership, zero-value cache design, TTL semantics, injected clocks, and disjoint-key race tests.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
