# Project 099 — Distributed File Storage

## 1. Project Name and Number

- Project 099, `099_distributed_file_storage`.
- Build a deterministic local simulation of distributed content-addressed file storage.
- Nodes are isolated directories under a supplied test root; no network service is required.
- Bounded files are split into fixed-size chunks identified by SHA-256, identical chunks are deduplicated, every chunk is replicated across a configured number of distinct nodes chosen deterministically, and immutable content-addressed manifests are published through atomically replaced live references.
- This README is a learning guide only.
- It contains no implementation code, signatures, starter snippets, solution snippets, pseudocode, or implementation shell commands.
- Text-only input and output examples are permitted.

## 2. Project Idea

The caller supplies a test root under which the project constructs a fixed set of isolated node directories. A bounded file is presented through `Put`. The file is split into fixed-size chunks, each chunk's SHA-256 digest is computed, and the chunk's bytes are staged on the configured number of distinct node directories selected deterministically. The chunk's bytes are atomically renamed into place after size and digest verification. An immutable manifest object is computed and stored by digest, and a live reference for the file identifier and version is atomically published to that manifest digest. `Get`, `Verify`, `Repair`, and `Delete-metadata` resolve by file identifier and version, then verify the manifest digest before use. Get assembles the bytes only after every chunk verifies. The local simulation teaches invariants rather than real distributed consensus or crash-proof storage.

## 3. Why This Project Now?

- Projects 025, 034, and 087 are the formal prerequisites: Project 025 contributes content hashing, Project 034 contributes worker-pool discipline, and Project 087 contributes key/value ownership and storage.
- Project 098 is optional immediate-catalog-predecessor context, while Projects 088 and 089 are optional prior review for corruption recovery and invariant-based testing.

## 4. Prerequisites

- Projects 025, 034, and 087 are the formal prerequisites.
- Project 025 provides content hashing; Project 034 provides worker-pool discipline; Project 087 provides key/value ownership and storage.
- Project 098 is optional immediate-catalog-predecessor context.
- Optional prior review includes Project 088 for corruption and recovery discipline and Project 089 for invariant-based testing.
- Be comfortable with byte slice handling, fixed-size chunk boundaries, canonical UTF-8 JSON serialization, content-addressed storage, atomic file replacement through temp-file-and-rename, the difference between a manifest object and a live reference, and the difference between a missing replica and a corrupt replica.

## 5. What You Must Know Before Starting

- A content-addressed store addresses each chunk and each manifest by a digest of its content. A matching digest identifies a candidate reusable object, but every existing object is verified for size, digest, and bytes before reuse so corruption or the theoretical collision case is never silently accepted.
- A "replica" means a copy of a chunk placed on a distinct node directory. The replication factor is the count of distinct nodes that must hold a valid replica of each chunk.
- A "manifest" is an immutable content-addressed object that records, for one file identifier at one version, the chunk digests in order, the chunk sizes, and the selected node identifiers per chunk. A live reference points from a file identifier and version to the manifest digest.
- "Deterministic node placement" means the placement algorithm is a pure function of the chunk digest and the configured node set, reproducible across runs, and independent of map iteration order.
- "Atomic replacement" means writing to a temporary path in the same directory as the destination and renaming into place.
- "Staging" means writing chunk files to a target node directory and renaming them into place only after size and digest verification. The manifest reference is published only after all required replicas and the immutable manifest object exist.
- "Garbage collection" is an optional bounded operation that identifies chunks and manifest objects no longer referenced by any live reference and reclaims them.
- The simulation teaches invariants. It is not real distributed consensus, cross-node replication, crash-proof storage, erasure coding, encryption, consensus, a cloud object store, or production durability.

## 6. Explanation of New Concepts

### Concepts

- The configuration is validated before any filesystem mutation.
- Chunk size is exactly `1 MiB`.
- Maximum input is exactly `32 MiB`.
- Configured node identifiers are unique ASCII strings of 1-32 characters drawn from letters, digits, underscore, and hyphen, sorted ascending.
- Node count is 1 through 16 inclusive.
- Replication factor is 1 through `min(3, node count)` inclusive.
- File identifier is 1-64 ASCII characters drawn from letters, digits, period, underscore, and hyphen.
- Version is a decimal integer from 1 through the maximum signed 64-bit value.
- Any invalid configuration is rejected before any directory is created.

- Deterministic node placement takes the first eight bytes of the chunk SHA-256 as an unsigned big-endian number, computes its remainder by the configured node count, and uses that remainder as the start index in the sorted node list.
- The placement then selects the configured replication-factor consecutive distinct nodes with wraparound.
- The placement is a pure function of the chunk digest and the configuration.

- Immutable manifests are content-addressed.
- A canonical manifest uses version `1`, a fixed field order, the file identifier, the version, the original total size, the chunk size, and an ordered array by zero-based chunk position.
- Each entry in the ordered array carries the chunk digest, the chunk byte length, and the selected node identifiers.
- The manifest is serialized as compact UTF-8 JSON with fixed field order and no insignificant whitespace.
- The manifest digest is the SHA-256 of the serialized bytes.
- The manifest object is stored immutably by its digest.
- The live reference for the file identifier and version is atomically published to the manifest digest.

- `Put` validates the configuration, splits the file into fixed-size chunks, computes each chunk's SHA-256 digest, and for each chunk stages the bytes on each placement-selected node.
- An existing replica is checked for exact byte length, digest, and byte equality.
- A valid identical replica is reused.
- A missing or corrupt replica is replaced atomically from the current verified input bytes; a same-digest different-content result returns an integrity-collision outcome and is never accepted.
- Each new or replacement chunk is renamed into place only after size and digest verification.
- The same full verification applies before reusing an existing manifest object.
- The live reference is published only after the manifest object verifies.
- A repeated `Put` with the same file identifier and version whose computed manifest digest equals the existing manifest digest is an idempotent no-op.
- A different digest returns `VersionConflict` and preserves the existing reference; changed content requires a different version.

- `Get` resolves the live reference for the file identifier and version, verifies the manifest digest, and for each chunk reads the chunk from one of the recorded node identifiers.
- The chunk's recorded byte length and digest are verified before the bytes are assembled into a private temporary buffer or file.
- Bytes are published to the caller only after every chunk has verified.
- If a chunk's verification fails and another recorded node holds a replica, the read switches to that replica.
- If no valid replica remains, `Get` returns a typed integrity outcome and publishes no partial bytes.

- `Verify` resolves the live reference, verifies the manifest digest, walks the manifest, and reports per-chunk integrity without writing any replicas. `Repair` resolves the live reference, verifies the manifest digest, and rewrites a corrupt or missing replica only to its originally selected configured node identifier.
- If that node is unavailable, `Repair` reports `RepairIncomplete`; it does not silently change placement, does not change the manifest, and does not alter the live reference. `Delete-metadata` removes only the live reference for the file identifier and version.
- Chunks and immutable manifest objects are not eagerly deleted.
- A repeated `Delete-metadata` is a typed no-op.

- Empty files have zero chunks and zero chunk locations.
- The manifest records an empty ordered array.
- The assembled output for an empty file is the empty byte sequence.
- No chunk files exist on any node for an empty file.

- On a `Put` failure, any newly created temporary files are removed and any previous reference is preserved.
- Successfully published immutable chunk objects that remain unreferenced are reclaimed only by the optional bounded garbage-collection step.
- Successfully published immutable manifest objects that are no longer referenced through any live reference are reclaimed only by the optional bounded garbage-collection step.
- The live reference is the only entry point for `Get`, `Verify`, `Repair`, and `Delete-metadata`.

## 7. Learning Objective

- After completing this project you must be able to explain in your own words: why the configuration is pinned exactly and rejected before any filesystem mutation;
- Why deterministic node placement is a pure function of the chunk digest and the configuration;
- Why manifests are immutable and content-addressed and why live references are separated from manifest objects;
- Why each chunk is staged on each selected node and atomically renamed into place only after size and digest verification;
- Why the live reference is atomically published only after every required replica and the immutable manifest object exist;
- Why a repeated `Put` with the same file identifier and version is an idempotent no-op when the manifest digest matches and a `VersionConflict` otherwise;
- Why empty files have zero chunks and zero chunk locations;
- Why `Get` assembles into a private temporary buffer or file and publishes only after every chunk verifies;
- Why `Repair` rewrites only to the originally selected configured node identifier and reports `RepairIncomplete` when that node is unavailable;
- Why `Delete-metadata` removes only the live reference and never eagerly reclaims chunks or manifest objects;
- Why the local simulation teaches invariants rather than real distributed consensus or crash-proof storage;
- Why the project is not a network protocol, an erasure-coded object store, an encrypted object store, a consensus system, a cloud object store, or a production durability system;
- And how each invariant is tested deterministically using a supplied test root and temporary node directories.

## 8. Functional Requirements

1. The project operates on a supplied test root. Each node is an isolated directory beneath the test root. The node set is fixed at configuration time and is reproducible across runs.
2. The configuration is validated before any filesystem mutation. Chunk size is exactly `1 MiB`. Maximum input is exactly `32 MiB`. Configured node identifiers are unique ASCII strings of 1-32 characters drawn from letters, digits, underscore, and hyphen, sorted ascending. Node count is 1 through 16 inclusive. Replication factor is 1 through `min(3, node count)` inclusive. File identifier is 1-64 ASCII characters drawn from letters, digits, period, underscore, and hyphen. Version is a decimal integer from 1 through the maximum signed 64-bit value. Invalid configurations are rejected before any directory is created.
3. Deterministic node placement takes the first eight bytes of the chunk SHA-256 as an unsigned big-endian number, computes its remainder by the configured node count, and uses that remainder as the start index in the sorted node list. The placement then selects the configured replication-factor consecutive distinct nodes with wraparound. The placement is a pure function of the chunk digest and the configuration.
4. The manifest is an immutable content-addressed object. The canonical manifest uses version `1`, fixed field order, the file identifier, the version, the original total size, the chunk size, and an ordered array by zero-based chunk position. Each entry carries the chunk digest, the chunk byte length, and the selected node identifiers. The manifest is serialized as compact UTF-8 JSON with fixed field order and no insignificant whitespace. The manifest digest is the SHA-256 of the serialized bytes.
5. A live reference for the file identifier and version is atomically published to the manifest digest. The live reference is replaced atomically through temp-file-and-rename.
6. `Put` fully verifies an existing chunk or manifest before reuse. Identical verified content is reused; a missing or corrupt replica is atomically replaced from the current verified input; a same-digest different-content result returns an integrity-collision outcome and is never accepted. `Put` publishes the manifest reference only after all required replicas and the immutable manifest object verify. On failure, newly created temporary files are removed and any previous reference is preserved. Successfully published immutable objects may remain unreferenced until optional garbage collection.
7. A repeated `Put` with the same file identifier and version whose computed manifest digest equals the existing manifest digest is an idempotent no-op. A repeated `Put` with the same file identifier and version whose computed manifest digest differs returns `VersionConflict` and preserves the existing reference. A different positive version is required for changed content.
8. `Get` resolves the live reference, verifies the manifest digest, reads each chunk from one of the recorded node identifiers, verifies the chunk's recorded byte length and digest, and assembles the bytes into a private temporary buffer or file. Bytes are published to the caller only after every chunk has verified. On a failed verification the read switches to another recorded replica when one is available. On no valid replica `Get` returns a typed integrity outcome and publishes no partial bytes.
9. `Verify` resolves the live reference, verifies the manifest digest, walks the manifest, and reports per-chunk integrity without writing any replicas.
10. `Repair` resolves the live reference, verifies the manifest digest, and rewrites a corrupt or missing replica only to its originally selected configured node identifier. If that node is unavailable, `Repair` reports `RepairIncomplete`. `Repair` never changes placement, the manifest, or the live reference.
11. `Delete-metadata` removes only the live reference for the file identifier and version. Chunks and immutable manifest objects are not eagerly deleted. Repeated `Delete-metadata` is a typed no-op.
12. Empty files have zero chunks and zero chunk locations. The manifest records an empty ordered array. The assembled output for an empty file is the empty byte sequence. No chunk files exist on any node for an empty file.
13. The simulation is local. The project never opens a network socket, never resolves a hostname, and never speaks a distributed protocol.

## 9. Inputs and Outputs

### Interface Contract

- Inputs are the validated configuration, the file identifier and version, the file bytes for `Put`, the file identifier and version for `Get`, the file identifier and version for `Verify`, and the file identifier and version for `Delete-metadata` and `Repair`.
- Outputs are the stored manifest digest for `Put`, the bytes for `Get` or a typed integrity outcome, the per-chunk integrity report for `Verify`, the typed `RepairIncomplete` outcome for `Repair` when the originally selected node is unavailable, the typed no-op outcome for repeated `Delete-metadata`, and the typed `VersionConflict` outcome for a repeated `Put` whose manifest digest differs.
- Text-only behaviour example. Put a small file with two chunks under a replication factor of two. The result observes two chunk files on each of the deterministically selected distinct nodes and one immutable manifest object, and the live reference for the file identifier and version is atomically published to that manifest digest. Get the same file. The result observes the original bytes. Verify the file. The result reports each chunk as valid against at least one replica.
- Text-only behaviour example. Put an empty file. The result observes one immutable manifest object whose ordered array is empty, zero chunk files on any node, and a live reference that resolves to that manifest digest. Get the empty file. The result observes the empty byte sequence and no partial output.
- Text-only behaviour example. Mark one replica of one chunk corrupt and call Get. The result observes the operation switching to the second replica, verifying the chunk, and returning the original bytes. Call Repair. The result observes the corrupt chunk rewritten from the valid source on its originally selected configured node, and the replication factor restored.
- Text-only behaviour example. Repeated `Put` with the same file identifier and version and the same content. The result observes an idempotent no-op. Repeated `Put` with the same file identifier and version and different content. The result observes `VersionConflict` and the existing reference preserved.

## 10. Rules and Edge Cases

- A `Put` whose staging, verification, manifest publication, or reference publication fails removes any newly created temporary files and preserves any previous reference. Successfully published immutable chunk objects that remain unreferenced are reclaimed only by the optional bounded garbage-collection step.
- A `Get` on a missing live reference returns a typed outcome and produces no output.
- A `Get` whose first chunk's verification fails on every recorded replica returns the typed integrity outcome and produces no output.
- A `Verify` on a manifest whose declared digest does not match its computed digest returns a typed integrity outcome for the manifest itself.
- A `Verify` whose chunk digest does not match the readback contents returns an integrity outcome for that chunk.
- A `Repair` whose originally selected configured node is unavailable for a chunk returns `RepairIncomplete` for that chunk without writing.
- A `Delete-metadata` whose deleted reference leaves some chunks or immutable manifest objects unreferenced marks them for the optional bounded garbage-collection step. Repeated calls produce no-op typed outcomes.
- An empty file's manifest and chunk references are well-defined: zero chunks and zero chunk locations.
- The live reference is the only entry point for `Get`, `Verify`, `Repair`, and `Delete-metadata`.

## 11. Project Constraints

- The simulation is local. The project owns the test root; the simulation never opens a network socket, never resolves a host name, and never speaks a distributed protocol.
- Nodes are isolated directories under the supplied test root. No node is a real machine, container, or service.
- Replication is on distinct node directories. Cross-node write requires the configured number of distinct paths to exist for the chunk digest.
- Chunks and immutable manifest objects are not eagerly deleted. They are reclaimed only by the optional bounded garbage-collection step.
- No erasure coding, no encryption, no consensus, no cross-machine replication, no cloud object store, no production durability, no claim of automatic cross-machine deployment.

## 12. Design Questions Before Coding

- What is the validated configuration shape, and at which boundary are invalid configurations rejected?
- How is deterministic node placement expressed as a pure function of the chunk digest and the configuration?
- What is the canonical manifest shape, and how is the manifest digest computed and verified before use?
- How does the staging layer ensure that a crash mid-write leaves the previous reference intact?
- How does `Get` switch to another replica on verification failure, and what is the typed outcome when no valid replica exists?
- How does `Repair` rewrite only to the originally selected configured node identifier, and what is the typed outcome when that node is unavailable?
- How does the simulation observe deduplication across two files that share a chunk digest?
- How does the simulation observe empty-file representation with zero chunks and zero chunk locations?
- How does the simulation observe chunks and immutable manifest objects that persist after `Delete-metadata` because no live reference points at them?
- Why is the simulation local and not a network protocol, and how is that boundary enforced in code?

## 13. Implementation Milestones

1. Define the validated configuration, the supplied test root, the node directory layout, and the assertion that invalid configurations are rejected before any filesystem mutation.
2. Implement the chunk-splitting logic with the fixed chunk size, the maximum file size, the empty-file convention, and the chunk digest computation using SHA-256.
3. Implement the deterministic placement function as a pure function of the chunk digest and the configuration. Verify that map iteration order does not affect placement.
4. Implement the immutable manifest object with the canonical serialization, the SHA-256 digest, and the immutability guarantee.
5. Implement the live reference with atomic publication through temp-file-and-rename.
6. Implement `Put`: stage each new chunk on each placement-selected node directory, verify size and digest, atomically rename into place, build and store the immutable manifest object, and atomically publish the live reference. Handle `VersionConflict` and idempotent no-op outcomes.
7. Implement `Get`: resolve the live reference, verify the manifest digest, read each chunk from a recorded node identifier, verify size and digest, assemble into a private temporary buffer or file, and publish bytes only after every chunk verifies.
8. Implement `Verify`: resolve the live reference, verify the manifest digest, walk the manifest, and report per-chunk integrity without writing.
9. Implement `Repair`: resolve the live reference, verify the manifest digest, rewrite a corrupt or missing replica only to its originally selected configured node identifier, and report `RepairIncomplete` when that node is unavailable.
10. Implement `Delete-metadata`: remove only the live reference and produce a typed no-op on repeat.
11. Write the unit test suite that owns the test root through temporary directories and exercises every pinned invariant.
12. Verify under the race detector and reproduce the honest statement about invariant teaching versus real distributed storage.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Configuration validation: configure an invalid chunk size, an invalid maximum input, an invalid node identifier, an out-of-range node count, an out-of-range replication factor, an invalid file identifier, an invalid version, and a configuration where the replication factor exceeds the node count; assert each is rejected before any directory is created.
- Single-chunk file: put a small file that fits in one chunk, get it back, and verify the chunk digest.
- Multi-chunk file: put a file that spans multiple chunks, get it back, and verify the assembled digest equals a digest computed over the original bytes.
- Empty file: put an empty file, get it back, observe the empty byte sequence, and assert no chunk files exist on any node and the manifest's ordered array is empty.
- Binary data: put a file containing arbitrary bytes, get it back, and verify the bytes round-trip exactly.
- Deduplication: put two files that share a chunk digest and assert the chunk lives on the same set of nodes for both files and that only one chunk file exists per node for that digest.
- Deterministic placement: run two `Put` operations of the same file under the same configuration and assert the placement set, the manifest digest, and the live reference are identical.
- Replication factor validation: configure a replication factor that exceeds `min(3, node count)` and assert the configuration is rejected before any directory is created.
- Node loss: remove one node directory and assert that `Get` still returns the original bytes by switching to a replica on another recorded node.
- One corrupt replica with fallback and repair: mutate one replica's bytes, assert `Get` returns the original bytes through the second replica, assert `Verify` reports the corrupt chunk, and assert `Repair` rewrites the chunk on its originally selected configured node identifier and restores the replication factor.
- All replicas invalid: corrupt every replica of one chunk and assert `Get` returns a typed integrity outcome and publishes no partial output.
- Truncated manifest: hand-craft a manifest whose bytes are too short and assert the read returns a typed integrity outcome.
- Wrong digest: hand-craft a manifest whose declared digest does not match its computed digest and assert the read returns a typed integrity outcome.
- Existing-object verification: reuse valid identical chunks and a valid identical manifest as idempotent objects; corrupt an existing replica and assert `Put` atomically restores it from the verified input; simulate same-digest different-content and assert an integrity-collision outcome with no publication.
- Repeated `Put` idempotent no-op: put the same file identifier and version with the same content twice and assert the second `Put` is an idempotent no-op and the live reference is unchanged.
- Repeated `Put` version conflict: put the same file identifier and version with different content and assert the second `Put` returns `VersionConflict` and the previous reference is preserved.
- Shared chunks: put three files where two of them share two chunks and assert those chunks live on the same node set for both files.
- Interrupted writes: simulate a mid-`Put` failure through staging and assert no new live reference points at staged chunks and that the previous reference is intact.
- No partial `Get` output: drive `Get` through a failure on every replica and assert no bytes are published and a typed integrity outcome is returned.
- Deterministic manifests: build two manifests from the same inputs and assert the digests are identical, and assert that perturbing one byte of metadata yields a different digest predictably.
- `RepairIncomplete`: take the originally selected configured node identifier for a chunk out of service and assert `Repair` reports `RepairIncomplete` without changing placement, the manifest, or the live reference.
- `Delete-metadata` no-op: call `Delete-metadata` twice and assert the second call is a typed no-op; assert that the underlying chunks and immutable manifest objects remain on disk and are not eagerly deleted.

## 15. Common Mistakes to Watch For

- Letting map iteration order influence placement, so two runs place a chunk on different node sets.
- Forgetting to verify a chunk's recorded byte length and digest before assembling output.
- Returning partial bytes when verification fails on every replica. Only a fully verified set is publishable.
- Eagerly deleting chunks or immutable manifest objects at `Delete-metadata` and corrupting other manifests or other live references that still reference them.
- Renaming the live reference before every required replica exists or before the immutable manifest object exists.
- Treating a missing or corrupt replica as silence instead of a typed integrity outcome.
- Letting `Repair` change placement, the manifest, or the live reference when the originally selected node is unavailable.
- Writing a network protocol, a custom consensus implementation, an erasure-coded layout, an encryption layer, or any cross-machine replication. The project is local and invariant-driven.
- Calling the result a distributed system. It is a local simulation that teaches invariants.
- Reading real network configuration, environment variables, hostnames, or remote storage endpoints. The project is local.
- Letting the maximum file size or chunk size become unbounded. Configuration is validated up front.
- Treating repeated `Put` with different content as success when the file identifier and version match. The correct response is `VersionConflict` and the previous reference is preserved.
- Treating `Repair` as free to choose a different node. `Repair` rewrites only to the originally selected configured node identifier.
- Adding parity chunks or any erasure-coding step. The project has no erasure coding.

## 16. Topics and References for Study

- The SHA-256 specification used as a content digest for chunks and immutable manifests.
- The canonical UTF-8 JSON serialization discipline for immutable manifests.
- The discipline of snapshot and atomic replacement used in Project 087, and the discipline of staged on-disk publication and recovery ordering used in Project 088.
- The discipline of invariant-driven testing used in project 089.
- Projects 025, 034, and 087 are the formal prerequisites: Project 025 for content hashing, Project 034 for worker-pool discipline, and Project 087 for key/value ownership and storage. Project 098 is optional immediate-catalog-predecessor context; Projects 088 and 089 are optional study for corruption recovery and invariant-driven testing.

## 17. Self-Assessment Questions

1. Why is SHA-256 used for chunks and immutable manifests, and how does content-addressing enable deduplication?
2. Why is configuration validated before filesystem mutation, and why must replicas use distinct node directories?
3. Why is deterministic placement a pure function of digest and configuration, independent of map iteration?
4. Why are manifests immutable and content-addressed while live references remain separate and atomically replaceable?
5. Why must every chunk and manifest be size- and digest-verified before reuse, read, assembly, or publication?
6. How does `Get` fall back between replicas while preventing partial output, and what happens when none verifies?
7. Why does `Repair` write only to the originally selected node, and when does it return `RepairIncomplete`?
8. Why does `Delete-metadata` avoid eager reclamation, and how does that protect other live references?
9. How are empty files represented, and what invariants must hold for their manifests and chunk locations?
10. What does this local simulation teach, and which distributed, durability, deployment, encryption, erasure-coding, and cloud-storage claims are outside scope?

## 18. Definition of Completion

- [ ] The project is complete when the simulation owns a supplied test root and creates isolated node directories beneath it;
- [ ] When the configuration is pinned exactly and invalid configurations are rejected before any filesystem mutation;
- [ ] When deterministic node placement takes the first eight bytes of the chunk SHA-256 as an unsigned big-endian number, computes its remainder by the configured node count, and selects the configured replication-factor consecutive distinct nodes with wraparound;
- [ ] When the manifest is an immutable content-addressed object serialized as compact UTF-8 JSON with fixed field order and no insignificant whitespace, hashed with SHA-256, stored immutably by digest, and atomically published through a live reference;
- [ ] When `Put` stages each new chunk on each placement-selected node directory, verifies size and digest, atomically renames into place, builds and stores the immutable manifest object, and atomically publishes the live reference only after every required replica and the immutable manifest object exist;
- [ ] When a repeated `Put` with the same file identifier and version whose manifest digest matches is an idempotent no-op, when a repeated `Put` whose manifest digest differs returns `VersionConflict` and preserves the previous reference, and when a different positive version is required for changed content;
- [ ] When `Get` resolves the live reference, verifies the manifest digest, reads each chunk from a recorded node identifier, verifies size and digest, assembles into a private temporary buffer or file, and publishes bytes only after every chunk verifies;
- [ ] When `Verify` resolves the live reference, verifies the manifest digest, walks the manifest, and reports per-chunk integrity without writing;
- [ ] When `Repair` resolves the live reference, verifies the manifest digest, rewrites a corrupt or missing replica only to its originally selected configured node identifier, and reports `RepairIncomplete` when that node is unavailable without changing placement, the manifest, or the live reference;
- [ ] When `Delete-metadata` removes only the live reference and produces a typed no-op on repeat without eagerly deleting chunks or immutable manifest objects;
- [ ] When empty files have zero chunks and zero chunk locations and the assembled output is the empty byte sequence;
- [ ] When the unit tests pass with a supplied test root covering configuration validation, single-chunk files, multi-chunk files, empty files, binary data, deduplication, deterministic placement, replication factor validation, node loss, one corrupt replica with fallback and repair, all replicas invalid, truncated manifest, wrong digest, repeated `Put` idempotent no-op, repeated `Put` version conflict, shared chunks, interrupted writes, no partial `Get` output, deterministic manifests, `RepairIncomplete`, and `Delete-metadata` no-op;
- [ ] When the race detector is clean;
- [ ] When the project documentation reproduces the honest statement that the simulation teaches invariants and is not real distributed consensus, crash-proof storage, encryption, erasure coding, a cloud object store, or a production durability system;
- [ ] And when this guide contains no implementation code, signatures, starter snippets, solution snippets, pseudocode, or implementation shell commands.

## 19. Optional Extensions

- A bounded mark-and-sweep garbage-collection step that identifies chunks and immutable manifest objects that are no longer reachable from any live reference and reclaims them, demonstrating that the optional bounded garbage-collection path is consistent with the no-eager-deletion invariant.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 025 — File Duplicate Finder](../../02-data-structures/025_file_duplicate_finder/README.md#20-prerequisite-based-documentation-guide), [Project 034 — Worker Pool Basic](../../03-concurrency/034_worker_pool_basic/README.md#20-prerequisite-based-documentation-guide), [Project 087 — KV Store In Memory](../../07-advanced-systems/087_kv_store_in_memory/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- None. This project applies already introduced APIs, standards, and testing practices in a new combination.

### Project-specific learning focus

- **Learn now:** content-defined identity, fixed-size chunking, immutable canonical manifests, copy ownership, replication and repair, quorum limitations, atomic references, garbage collection, and invariant tests.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
