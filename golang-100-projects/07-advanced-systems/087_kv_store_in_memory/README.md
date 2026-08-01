# Project 087 — KV Store In Memory

## 1. Project Name and Number
Project 087, kv_store_in_memory. Build a concurrent in-memory store from string keys to byte values with Set, Get, Delete, TTL, Len, explicit expiry cleanup, deterministic JSON snapshots, atomic restore, and a narrow atomic-file-save boundary. This README is a learning guide only and contains no implementation code, signatures, snippets, pseudocode, or solution commands.

> **Scope.** This is a single-process educational store, not a database. It has no replication, transactions across keys, write-ahead log, distributed coordination, eviction policy, access control, or crash-proof durability.

## 2. Project Idea
Own every byte slice, make expiry decisions from an injected clock, and define one lock-protected linearization point for every state operation. Zero TTL means no expiry, positive TTL becomes an absolute expiration, and negative TTL is invalid without mutation. Reads lazily remove expired data; Len, explicit cleanup, and snapshot use the same expiry boundary. A versioned JSON snapshot round-trips semantic state, while restore validates a complete candidate before replacing memory atomically.

## 3. Why This Project Now?
This project takes its required foundation from Project 043 (thread_safe_cache), which supplies copy-out ownership, fixed TTL semantics, atomic test doubles, and race-detector-tested concurrent maps behind one documented concurrency model. The catalog's immediate predecessor is Project 086 (distributed_task_queue); Project 086 is referenced here only as optional context for mutex-protected state transitions, absolute logical times, copied payload ownership, and deterministic fake-clock testing. This project narrows those ideas into a reusable storage abstraction and adds snapshot consistency, strict schema validation, and honest filesystem durability boundaries.

## 4. Prerequisites
Project 043 (thread_safe_cache) is the required prerequisite. Project 086 (distributed_task_queue) is the immediate catalog predecessor and is useful for context but is not required. Be comfortable with mutexes and read/write mutexes, byte-slice aliasing, absolute time, JSON encoding rules, strict validation, stable sorting, file replacement, injected failure boundaries, temporary directories, and race-detector tests. Understand that serializing a map directly does not define deterministic entry order or duplicate-key validation.

## 5. What You Must Know Before Starting
Know the difference between copying a slice header and copying its bytes, the difference between a read lock and a write lock, and why lazy expiry can turn Get into a mutating operation. Understand exact expiration boundaries, integer overflow in time arithmetic, point-in-time snapshots, decode-then-validate workflows, all-or-nothing in-memory replacement, same-directory rename, checked write and close errors, file synchronization, and the limits of filesystem promises across operating systems and power loss.

## 6. Explanation of New Concepts
Keys are non-empty valid UTF-8 strings of at most 256 bytes. Each value is at most 1 MiB. The store holds at most 10,000 live entries and at most 64 MiB of live value bytes. Validation counts bytes, not characters. A rejected Set, snapshot, save, or restore does not partially apply its requested data change. Overwriting a live key accounts for the old value before enforcing total-byte capacity.

Set copies the input value before publishing it. Get returns a fresh copy. Snapshot also copies values into an isolated candidate. A caller that changes an original input slice or a returned slice therefore cannot mutate stored state. A zero-length value is valid and distinct from a missing key.

The store uses one injected clock and maintains a lock-protected high-watermark time. Every TTL-sensitive operation observes the clock once and uses the later of that observation and the prior high watermark. This effective time never moves backward. Moving a fake clock backward therefore cannot extend a TTL, make an expired entry visible, or resurrect an entry already removed.

Zero TTL stores an explicit non-expiring entry. Positive TTL stores an absolute expiration equal to effective time plus TTL, after overflow validation. Negative TTL is invalid and leaves the prior value, expiry, counters, and high watermark unchanged. An entry is expired when effective time is equal to or later than its absolute expiration. Exact equality is expired.

Get has one logical result boundary. Under exclusive locking, it observes effective time, checks the key, removes the entry and adjusts counters if expired, or copies and returns the live value. Missing and expired keys both report not found to callers. Delete removes a live entry and reports that it existed. Deleting an expired entry removes it but reports not found. Repeated Delete is harmless.

Len returns the number of live entries. It obtains exclusive locking, observes effective time once, removes every expired entry, updates counters, and then counts. Explicit cleanup follows the same rule and additionally reports how many entries it removed. Snapshot also performs a complete expiry sweep at one effective time before copying state. These policies ensure Len and snapshots never count expired entries.

Read/write locking is disciplined rather than decorative. Operations that can remove expired entries require exclusive locking. A pure lookup that cannot mutate would be eligible for a read lock, but required Get is lazy-cleaning and therefore cannot remain under only a read lock when it finds expiry. No JSON encoding, decoding, file read, file write, synchronization, close, rename, callback, or other I/O occurs while the store lock is held.

Snapshot version one is a strict JSON document with exactly three top-level members: version, captured-at time, and an entries array. Captured-at and positive expirations are absolute signed Unix-nanosecond timestamps. A non-expiring entry has an explicit null expiration. Each entry has exactly a key, a canonical base64 byte value, and an expiration. Unknown or missing members, trailing JSON values, malformed base64, out-of-range timestamps, invalid UTF-8 keys, duplicate keys, unsorted or sorted input alike with duplicates, and unsupported versions are invalid. Input order is never trusted.

A generated snapshot sorts entries by raw UTF-8 key bytes, uses compact JSON, and ends with exactly one newline. This gives deterministic bytes for the same semantic state and captured time. Snapshot creation sweeps expiry and copies a stable state while holding the lock, then releases the lock before encoding. Concurrent mutations after that copy do not change the snapshot; they belong to a later state.

Restore first reads, bounds, decodes, and validates the entire candidate outside the store lock. Snapshot input is capped at 128 MiB. It validates version, schema, unique non-empty keys, key and value bounds, absolute expirations, entry count, and total live bytes. At the final replacement boundary it takes exclusive locking, computes effective time as the maximum of the existing high watermark, current clock observation, and snapshot captured-at time, omits entries expired at that time, rechecks live capacity, and swaps the complete candidate into memory. Any failure before the swap leaves all prior entries and counters unchanged.

Expired-at-restore entries are omitted, not installed and immediately exposed. Non-expiring entries remain non-expiring. Future absolute expirations preserve their original instant rather than receiving a fresh TTL. Restore is replacement, not merge. A valid empty snapshot atomically empties the store. Concurrent operations observe either the complete old state or the complete restored state at the lock boundary, never a mixture.

Atomic file save first obtains already-encoded snapshot bytes without holding the store lock during I/O. It creates a uniquely named temporary file in the destination directory, writes all bytes, checks file synchronization, checks close, and then renames over the destination. Any failure before rename leaves an existing destination unchanged and attempts to remove the temporary file. A successful same-directory rename provides atomic name replacement on supported local filesystems. This project does not claim atomicity on unusual filesystems, durability of directory metadata across power loss, network-filesystem semantics, or rollback after a rename that succeeded but whose result could not be reported.

Text-only state examples are permitted. A key set with zero TTL remains visible until overwritten or deleted. A key set with a positive TTL is visible immediately before its expiration and absent at the exact expiration. A snapshot taken before expiration may contain that absolute expiration, yet restore after the expiration omits the key. Advancing time, observing removal, and moving fake time backward leaves the key absent.

## 7. Learning Objective
Build and verify a bounded concurrent byte-value store with explicit ownership, exact TTL boundaries, non-resurrecting logical time, lazy and explicit cleanup, stable point-in-time snapshots, strict versioned restore, no I/O under lock, and narrow same-directory file replacement guarantees without presenting the result as a production database.

## 8. Functional Requirements
1. Support Set, Get, Delete, TTL, Len, explicit cleanup, snapshot, restore, save, and load behavior over string keys and byte values.
2. Validate non-empty UTF-8 keys up to 256 bytes, values up to 1 MiB, at most 10,000 live entries, at most 64 MiB of live value bytes, and snapshots up to 128 MiB.
3. Copy byte values on Set, Get, snapshot capture, and restore publication.
4. Interpret zero TTL as non-expiring, positive TTL as an absolute expiration from one effective-time observation, and negative TTL as invalid with no mutation.
5. Treat an entry as expired when effective time is equal to or later than expiration.
6. Maintain a non-decreasing high-watermark time so backward fake-clock movement cannot extend or resurrect data.
7. Lazily remove expired entries in Get and Delete; make Len, cleanup, and snapshot sweep all expired entries.
8. Keep entry count and total live-byte counters consistent on set, overwrite, expiry, delete, restore, and empty replacement.
9. Use locking that makes each operation linearizable and never perform I/O or JSON work while holding the lock.
10. Generate version-one JSON with absolute expiration timestamps, explicit non-expiring values, canonical byte encoding, sorted unique entries, compact encoding, and one trailing newline.
11. Validate the complete schema and candidate before restore, omit expired-at-restore entries, preserve non-expiring entries, and atomically replace rather than merge.
12. Save through a same-directory temporary file with checked write, synchronization, close, rename, and pre-rename cleanup behavior.
13. Leave an old destination unchanged on every injected save failure before rename.
14. Leave in-memory state unchanged on read, decode, schema, version, duplicate-key, bound, or storage failure during restore.

## 9. Inputs and Outputs
Set accepts a key, owned-copy value, and TTL classification. Get returns a copied value plus presence. Delete reports whether a live entry existed. Len returns live count after cleanup. Explicit cleanup reports removed-expired count. Snapshot returns deterministic versioned JSON for a stable state. Save targets a learner-owned path in a temporary test directory. Restore or load replaces state only after full validation. Validation, capacity, corruption, version, storage, and time-overflow failures are explicit outcomes; they never silently truncate or partially merge state.

## 10. Rules and Edge Cases
An empty key is invalid; a zero-length value is valid. Setting an existing key replaces both value and expiration. Setting with zero TTL clears an old positive expiration. Setting with negative TTL preserves the old entry exactly. At exact expiration Get reports missing and removes the entry. Len and snapshot may mutate only by removing expired entries and advancing the high watermark. A fake clock moving backward is clamped by the high watermark. Restore treats expiration equal to effective restore time as expired. Duplicate snapshot keys are invalid even if their values match. Unknown snapshot version, unknown fields, missing fields, malformed encoding, trailing JSON, and over-limit input are hard restore errors. A valid empty snapshot replaces all state. A pre-rename save failure preserves an old file. A post-rename reporting failure has an uncertain reported outcome and must not be described as rollback-safe. Temporary cleanup is required on known pre-rename failures but cannot be guaranteed after process death.

## 11. Project Constraints
Single process and one in-memory map protected by one locking discipline. No per-key transactions, compare-and-swap, range scans, eviction, disk-backed reads, write-ahead log, replication, encryption, compression, process sharing, or multi-process file locking. No background TTL ticker is required; expiry work is caused by Get, Delete, Len, cleanup, snapshot, or restore. JSON snapshots are educational and bounded, not an efficient database format. Same-directory rename has only the documented local-filesystem guarantee. Production storage would require stronger durability, migrations, access controls, monitoring, backup policy, and recovery testing.

## 12. Design Questions Before Coding
Why must Set and Get copy bytes? Why can lazy Get require exclusive locking? Which operations remove expired entries, and which exact time do they share? Why is expiration equality considered expired? Why maintain a high watermark instead of trusting a fake clock that can move backward? Why does snapshot copy under lock but encode after unlock? Why use an entries array rather than a JSON object? Why must restore validate duplicates and capacities before replacement? Why are absolute expirations preserved across restore? Why must the temporary file share the destination directory? Which failures guarantee the old file remains and which outcomes become uncertain after rename?

## 13. Implementation Milestones
1. Establish key, value, entry-count, total-byte, snapshot-size, timestamp, and TTL validation rules.
2. Establish owned byte copies, entry accounting, effective-time high watermark, and exact expiration checks.
3. Establish Set, lazy-cleaning Get, Delete, Len, and explicit cleanup under a consistent lock discipline.
4. Establish stable snapshot capture with one expiry sweep and copied entries sorted by key.
5. Establish strict version-one deterministic JSON encoding outside the lock.
6. Establish bounded decode and complete schema validation into an isolated restore candidate.
7. Establish atomic replacement under exclusive locking with restore-time expiry omission and capacity recheck.
8. Establish same-directory temporary save with injected write, sync, close, rename, and cleanup failures.
9. Establish load behavior that performs file I/O and validation before touching memory.
10. Complete fake-clock, semantic round-trip, corruption, failure-injection, concurrency, and race-detector verification.

## 14. Verification Cases the Learner Must Write
- Set and Get preserve exact bytes, including an empty value.
- Mutating an input slice after Set does not change stored data.
- Mutating a Get result does not change stored data.
- Overwrite changes value and TTL while preserving correct entry and byte counts.
- Delete distinguishes a live key from missing or expired data and is harmless when repeated.
- Zero TTL never expires; negative TTL returns an error and leaves prior state unchanged.
- Positive TTL is visible immediately before expiration and absent at the exact boundary.
- Get lazily removes expiry; Len, cleanup, and snapshot exclude all expired entries.
- Advancing past expiry, observing removal, then moving fake time backward never resurrects the entry.
- Capacity and overflow failures do not partially mutate data or counters.
- Concurrent Set, Get, Delete, Len, cleanup, and snapshot operations pass under the race detector.
- Snapshot round-trip preserves keys, bytes, non-expiring state, and future absolute expirations semantically.
- Snapshot entries are deterministic in raw key-byte order with compact encoding and one newline.
- Restore after an expiration omits that entry while preserving non-expiring entries.
- Corrupt JSON, trailing JSON, malformed byte encoding, unknown version, unknown or missing fields, duplicate keys, invalid keys, invalid timestamps, and over-limit content are rejected.
- Every failed restore leaves the old in-memory state and counters unchanged.
- A valid empty snapshot atomically replaces the store with empty state.
- Save failures during create, write, sync, close, or rename preserve an old destination when failure is before rename and remove known temporary files.
- Successful save and restore use only temporary directories owned by the test.
- No test sleeps or depends on wall-clock time.

## 15. Common Mistakes to Watch For
Storing or returning caller-owned slices; treating zero-length values as missing; allowing negative TTL to overwrite; checking expiry with a strict greater-than comparison; trusting backward clock values; reporting map length without cleanup; adjusting byte counters twice on overwrite; encoding while holding the lock; reading a file while holding the lock; serializing a map and assuming order; using a JSON object that hides duplicate keys; refreshing TTL during restore; merging a candidate into live state before validation finishes; installing expired entries; renaming a temporary file from another directory; ignoring sync or close errors; claiming a successful rename guarantees power-loss durability everywhere; or deleting the old destination before the replacement is ready.

## 16. Topics and References for Study
Study Go documentation for byte slices, read/write mutexes, time values, JSON, base64, sorting, file creation, file synchronization, close errors, rename behavior, temporary directories, and the race detector. Study linearizability, snapshot isolation at a single lock boundary, strict schema evolution, absolute versus relative expiration, cache expiry strategies, copy-in/copy-out ownership, and crash-consistency limitations. Review Project 043 (thread_safe_cache) for copy-out ownership and TTL semantics, and Project 086 (distributed_task_queue) for injected-clock and mutex-state-machine discipline.

## 17. Self-Assessment Questions
Which operations copy bytes, and what ownership does each copy preserve across the boundary? What does zero, positive, or negative TTL mean, and which of those leaves the prior entry unchanged? At what exact instant is a key expired, and what does that imply about equality with the absolute expiration? How does the high watermark handle backward fake time during normal operations and across restore, and why can a backward move neither extend a TTL nor resurrect a removed entry? Why does Len take an exclusive locking path, and which operations stay on that path versus a purely read path? At what point does a snapshot become stable, and which work is intentionally performed only after unlocking? Which schema errors must leave state unchanged, and why is the full candidate validated before any swap of memory? Why are snapshot expirations absolute rather than relative, and what does the restore path do with an entry already expired at the restore time? What does same-directory rename guarantee on supported local filesystems, and what does it not guarantee after power loss or a post-rename reporting failure? Which broad database properties are intentionally absent from this educational store, and which production guarantees does the project therefore refuse to claim?

## 18. Definition of Completion
- [ ] Project 043 (thread_safe_cache) is treated as the required prerequisite.
- [ ] Set, Get, Delete, Len, cleanup, snapshot, save, load, and restore follow one documented concurrency model.
- [ ] Keys, values, entry count, total bytes, snapshot size, TTL, and timestamps are bounded and validated.
- [ ] Every byte slice crossing the store boundary is copied.
- [ ] Zero TTL is non-expiring, positive TTL uses absolute time, and negative TTL makes no mutation.
- [ ] Expiration equality is expired and backward fake time cannot resurrect data.
- [ ] Get and Delete lazily remove expiry; Len, cleanup, and snapshot sweep expiry.
- [ ] No JSON or file I/O occurs while the store lock is held.
- [ ] Version-one snapshots preserve semantic state with deterministic key ordering and absolute expirations.
- [ ] Restore validates complete schema, uniqueness, bounds, and expiry before atomic replacement.
- [ ] Failed restore preserves memory; pre-rename failed save preserves the old file and cleans known temporary files.
- [ ] Temp-directory tests cover round-trip, corrupt state, unknown version, duplicates, expiry across restore, and injected save failures.
- [ ] Concurrent tests pass under the race detector with no sleeps.
- [ ] Production durability and database guarantees are not claimed.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions
- Add versioned compare-and-swap for one key while preserving copy ownership, TTL rules, and linearizable outcomes.
- Add a bounded least-recently-used eviction policy with deterministic fake-clock tests and snapshot semantics that explicitly distinguish eviction from expiry.
