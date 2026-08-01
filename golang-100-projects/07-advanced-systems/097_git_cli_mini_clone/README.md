# Project 097 — Git CLI Mini Clone

## 1. Project Name and Number
Project 097, `097_git_cli_mini_clone`. Build a small educational content-addressed object store and tracker inspired by Git's historical object format. The store is local, uses a `.minigit` directory under a supplied work tree, and pins the SHA-1 input shape and the canonical serialization rules so the educational object format is precise. This README is a learning guide only. It contains no implementation code, signatures, starter snippets, solution snippets, pseudocode, or implementation shell commands. Text-only input and output examples are permitted.

## 2. Project Idea
The caller supplies an existing empty directory. The program creates a `.minigit` directory under that root and refuses to operate otherwise. The store writes only within `.minigit`. The store reads explicitly requested work tree paths after containment and symlink checks; it never writes a work tree file and never reads an unstaged path, a host identity, an environment value, or the contents of a `.git` entry. Blob, tree, commit, and index objects are content-addressed by SHA-1 with a pinned header shape and stored as zlib-compressed files under `.minigit`. The index is a versioned document; the commit object references a tree built from the current live index and an optional single parent. SHA-1 is retained only to study Git's historical object format and is not a security recommendation.

## 3. Why This Project Now?
Projects 025 and 030 are the formal prerequisites: Project 025 contributes streaming and hashing filesystem discipline, and Project 030 contributes safe bounded file and container integrity. Project 096 is optional immediate-catalog-predecessor context, while Projects 087 and 088 are optional prior review for supplied-path operations, stable ordering, and atomic replacement.

## 4. Prerequisites
Projects 025 and 030 are the formal prerequisites. Project 025 provides streaming and hashing filesystem discipline; Project 030 provides safe bounded file and container integrity. Project 096 is optional immediate-catalog-predecessor context. Optional prior review includes Project 087 for supplying a path and a small operation set to that path and Project 088 for stable durable-state ordering. Be comfortable with byte slice handling, fixed format serialization, atomic file replacement through temp-file-and-rename, the difference between a manifest and an index, and the boundary between an educational format and a real Git installation.

## 5. What You Must Know Before Starting
- A content-addressed store addresses every object by a digest of its content. Matching digests identify a candidate existing object, but the store still verifies its kind, declared length, recomputed digest, and content bytes so corruption or the theoretical collision case is never silently accepted.
- The pinned SHA-1 input for every object in this store is the ASCII object kind, one ASCII space, the base-10 content byte length with no leading zero except the single digit `0`, one zero byte, and then the exact content bytes. The kind and the length prefix are part of the digest input.
- Object kinds in this educational store are exactly `blob`, `tree`, `commit`, and `index`. Object files are zlib-compressed on disk.
- SHA-1 is a historical format that Git used. New designs choose modern digest functions. This project retains SHA-1 only to study Git's historical object format and is honest that the choice is not a security recommendation.
- An "immutable object" means the bytes of the file under its digest never change. A second write of an existing digest that fully verifies is an idempotent no-op. A corrupt object or a same-digest different-content collision returns a typed integrity or collision outcome and is never overwritten.
- A "versioned index" is a small document whose revisions are themselves content-addressed and stored as objects. The live index is replaced atomically by temp-file-and-rename inside `.minigit`.
- A "deterministic commit" is one whose serialized object bytes depend only on the inputs and never on wall-clock time or process identifiers. The timestamp is injected by the caller as a normalized UTC value.
- "Atomic replacement" means writing to a temporary path in the same directory as the destination and renaming into place, so a reader sees either the old revision or the new revision and never a partial read.
- Path containment means every path supplied by the caller, after normalization, lives under the supplied work tree root. A path that escapes is rejected before any I/O.
- A symlink is rejected at staging if its target is absolute or resolves outside the work tree. The intermediate-path containment check rejects an intermediate path component that is itself a symlink escaping the work tree.
- The store writes only within `.minigit`. The store reads explicitly requested work tree paths after containment and symlink checks and never reads an unstaged path.

## 6. Explanation of New Concepts
The work tree is supplied by the caller. The caller supplies an existing empty directory; the program refuses to operate on a non-empty directory. The program creates a `.minigit` directory inside the supplied root. The presence of a `.git` entry anywhere inside the supplied root is rejected by presence only; the program never opens or reads the contents of a `.git` entry. After successful initialization the learner may create work tree files and stage them.

The store reads only what the caller explicitly requests. Reads are confined to paths the caller names, after the containment check and the symlink check pass. The store never writes a work tree file. The store never reads an unstaged path, a host identity, an environment value, or a `.git` entry's contents.

The pinned SHA-1 input for an object is the byte sequence formed by:
- the ASCII object kind, exactly one of `blob`, `tree`, `commit`, or `index`;
- one ASCII space byte;
- the base-10 content byte length as plain text with no leading zero except the single digit `0` when the content is empty;
- one zero byte;
- the exact content bytes.

The digest is the SHA-1 of that input. The known empty-blob vector for the kind `blob` and empty content is the digest `e69de29bb2d1d6434b8b29ae775ad8c2e48c5391`.

Object files are zlib-compressed on disk. The on-disk path is fixed by the digest: the leading two hex characters of the digest form a subdirectory and the remaining hex characters form the file name. The store decompresses the on-disk bytes to recover the header and the content, verifies the header's kind and length, and verifies the recomputed digest against the file name. A failure on any of these steps returns a typed integrity outcome.

Staging takes a list of paths from the work tree. Stage paths are slash-separated relative UTF-8 paths, non-empty, with no absolute path, with no empty or dot or dot-dot segment, with no NUL byte, and with a maximum length of `255` bytes. A staged regular file is bounded to `16` MiB. A staged symlink is recorded by mode and stores a relative UTF-8 link target of 1-255 bytes; the target is never followed for blob content, and an absolute target or any target that resolves outside the work tree is rejected. Every path component is checked so an intermediate path component that is itself a symlink escaping the work tree is rejected. The index carries at most `10,000` staged paths.

Index paths are sorted by raw UTF-8 byte sequence in ascending order. The sort order is independent of how the caller supplied the paths and is independent of map iteration. Two indexes built from the same paths produce identical digests.

The canonical index and tree documents are compact UTF-8 JSON with no insignificant whitespace and no trailing newline. Each has fields in the fixed order `version`, then `entries`; version is exactly `1`. Entries are already sorted and each entry has fields in the fixed order `path`, `mode`, then `digest`. Mode is exactly `regular`, `executable`, or `symlink`. The canonical commit document is compact UTF-8 JSON with fields in the fixed order `version`, `tree`, `parent`, `author_name`, `author_email`, `committer_name`, `committer_email`, `occurred_at`, then `message`; version is exactly `1`, and `parent` is the empty string when there is no parent. Strings use standard JSON escaping. Names are 1-100 trimmed UTF-8 characters, email values are 3-254 trimmed ASCII characters, and the message is at most 4,096 UTF-8 characters. The caller supplies `occurred_at` in UTC RFC 3339 form with exactly millisecond precision and the `Z` suffix. No map iteration participates in serialization.

A commit builds the canonical tree document from the entries in the current live index; the caller does not supply an arbitrary tree digest. The tree and index have separate object kinds and therefore separate digests even when they describe the same entries. A commit carries the resulting tree digest, an optional single parent digest, an injected author tuple, an injected committer tuple, the normalized UTC timestamp supplied by the caller, and a message. A root commit with an empty index, or a later commit whose tree equals its parent's tree, is rejected unless the caller explicitly sets the allow-empty flag. The flag is the only way to permit either empty case. The store never reads the host's user identifier, hostname, environment variable, or wall-clock time.

The existing-object behavior is a full verification. Before the store accepts a write of a digest that already exists on disk, the store decompresses the on-disk object, verifies the header kind and length, verifies the recomputed digest against the file name, and verifies the bytes against the supplied content. Identical verified content is an idempotent no-op. Corrupt content or a same-digest different-content collision returns a typed integrity or collision outcome and is never overwritten.

Atomic replacement applies to the live index path and the live commit-ref path. A temporary file is written, flushed, and renamed into place. A crash before rename preserves the previous revision; a crash after rename installs the new revision.

## 7. Learning Objective
After completing this project you must be able to explain in your own words: what the exact SHA-1 input is for a blob, a tree, a commit, and an index; why SHA-1 is retained for study only and not as a security recommendation; how a fully verified existing object behaves as an idempotent no-op and how a same-digest different-content collision is reported as a typed integrity outcome; why the index is itself content-addressed and versioned; why each index revision is replaced atomically; why stage paths are sorted by raw UTF-8 byte sequence; why an intermediate path component that is a symlink escaping the work tree is rejected even when the supplied path resolves inside the work tree; why the index caps a staged regular file at `16` MiB and the index itself at `10,000` paths; why the commit builds its tree from the current live index rather than from a caller-supplied tree digest; why the timestamp is injected by the caller as a normalized UTC value; why the store writes only within `.minigit` but reads only the explicitly requested work tree paths after containment and symlink checks; why the store never reads an unstaged path, host identity, environment value, or `.git` contents; and why the project is an educational store and not a Git implementation and makes no broad Git compatibility claim.

## 8. Functional Requirements
1. The program accepts an explicitly supplied work tree directory that exists and is empty. A non-empty directory is rejected. The presence of a `.git` entry anywhere inside the supplied root is rejected by presence only; the program never opens or reads the contents of a `.git` entry.
2. The program creates and uses only the `.minigit` path under the supplied work tree. The store writes only within `.minigit`. The store reads only the explicitly requested work tree paths after the containment check and the symlink check pass. The store never writes a work tree file, never reads an unstaged path, host identity, environment value, or `.git` contents.
3. The SHA-1 input for every object is the ASCII object kind, exactly one of `blob`, `tree`, `commit`, or `index`; one ASCII space byte; the base-10 content byte length with no leading zero except the single digit `0`; one zero byte; and the exact content bytes. The known empty-blob vector for `blob` and empty content is `e69de29bb2d1d6434b8b29ae775ad8c2e48c5391`.
4. Object files are zlib-compressed on disk. The on-disk path is the leading two hex characters of the digest as a subdirectory and the remaining hex characters as the file name.
5. Object files are immutable. Before accepting a write of an existing digest the store decompresses the on-disk object, verifies the header kind and length, verifies the recomputed digest, and verifies the bytes against the supplied content. Identical verified content is an idempotent no-op. Corrupt content or a same-digest different-content collision returns a typed integrity or collision outcome and is never overwritten.
6. The index is a content-addressed versioned document. Each index revision is stored as an object addressed by its digest. The live index is replaced atomically through temp-file-and-rename inside `.minigit`.
7. Stage paths are slash-separated relative UTF-8 paths, non-empty, with no absolute path, with no empty or dot or dot-dot segment, with no NUL byte, and with a maximum length of `255` bytes. A staged regular file is bounded to `16` MiB. The index carries at most `10,000` staged paths.
8. Index paths are sorted by raw UTF-8 byte sequence in ascending order. The sort is independent of supply order and of map iteration.
9. A staged symlink is recorded by mode and stores a relative UTF-8 link target of 1-255 bytes. The symlink target is never followed for blob content. An absolute, empty, overlong, invalid-UTF-8, or out-of-tree target is rejected. Every path component is checked so an intermediate path component that is itself a symlink escaping the work tree is rejected.
10. A tree carries an ordered list of entries. Each entry carries a path, a mode chosen from exactly `regular`, `executable`, or `symlink`, and a child digest. Tree entries are sorted by raw UTF-8 byte sequence in ascending order.
11. A commit carries a tree digest derived from the current live index, an optional single parent digest, an injected author tuple, an injected committer tuple, a normalized UTC timestamp supplied by the caller, a message, and an allow-empty decision. The caller does not supply an arbitrary tree digest. A root commit with an empty index and a later commit whose tree equals its parent's tree are rejected unless allow-empty is explicitly set.
12. The author tuple, the committer tuple, and the timestamp are caller-supplied values. The store never reads the host's user identifier, hostname, environment variable, or wall-clock time.
13. Index and tree serialization is compact UTF-8 JSON with no trailing newline, version `1`, fields ordered as `version` then `entries`, and entry fields ordered as `path`, `mode`, then `digest`. Commit serialization is compact UTF-8 JSON with version `1` and the fixed field order `version`, `tree`, `parent`, `author_name`, `author_email`, `committer_name`, `committer_email`, `occurred_at`, then `message`. Parent is the empty string when absent; strings use standard JSON escaping; the pinned identity, timestamp, and message bounds apply; map iteration never participates.
14. The live commit-ref is replaced atomically through temp-file-and-rename inside `.minigit`. A crash before rename preserves the previous live commit-ref; a crash after rename installs the new live commit-ref.
15. Path containment is enforced before any I/O. A normalized path that escapes the supplied work tree is rejected with a clear typed outcome.
16. On a failure inside an operation, the live index, the live commit-ref, and the live tree pointers remain at their pre-operation values. Newly written objects that were fully verified remain addressable by digest but are unreferenced.
17. The store operates locally. It never opens a network socket, never resolves a hostname, never reads environment variables, and never reads host identity.
18. The store is an educational format. It makes no broad Git compatibility claim beyond the pinned SHA-1 object identity shape and the fixed kind set.

## 9. Inputs and Outputs
- Inputs are the supplied work tree directory, the list of paths to stage from the work tree, and the inputs to commit: the injected author tuple, the injected committer tuple, the injected normalized UTC timestamp, the message, the optional single parent digest, and the explicit allow-empty flag.
- Outputs are the digest of each staged blob, the digest of the index revision, the digest of the tree, the digest of the commit, and the new live commit-ref content. A failed operation returns a typed outcome that identifies the boundary at which it failed.
- Text-only behaviour example. Initialize the store on an empty temporary directory. Stage a small text file. Build a commit with no parent using the injected author and timestamp. Read the commit object and observe that its tree entries match the live index snapshot, while the tree and index retain distinct digests because their object kinds and canonical documents are distinct. The commit digest depends only on its inputs.
- Text-only behaviour example. Make a second commit with the same parent, the same file, the same author and committer, the same timestamp, and the same message. Observe that the commit digest is identical to the first commit. Make a third commit with the same parent, the same file, the same inputs, but a different timestamp. Observe that the commit digest differs and the tree digest remains identical.
- Text-only behaviour example. Stage a path that resolves outside the work tree. Observe that the operation is rejected with a containment error before any index revision or commit is written, and the live commit-ref remains at its previous value.
- Text-only behaviour example. Re-stage the same content twice. The existing object is fully verified, returns an idempotent no-op, and the digest is unchanged. Tamper with the on-disk object and re-stage the same content. The store returns a typed integrity outcome and never overwrites.
- Text-only behaviour example. Stage a path whose intermediate component is a symlink that escapes the work tree. The operation is rejected at the intermediate-path check before any blob or tree is written.

## 10. Rules and Edge Cases
- A work tree that is non-empty is rejected at initialization. A work tree that contains a `.git` entry is rejected by presence only, and the program never opens or reads the contents of the `.git` entry.
- A staged path that is absolute, contains an empty segment, contains a dot segment, contains a dot-dot segment, contains a NUL byte, or exceeds `255` bytes is rejected before any I/O.
- A staged regular file larger than `16` MiB is rejected. An index that would carry more than `10,000` staged paths is rejected.
- A staged symlink with an absolute target or a target resolving outside the work tree is rejected. An intermediate path component that is a symlink escaping the work tree is rejected.
- A staged blob whose existing on-disk object decompresses, verifies its header and digest, and matches the supplied content is an idempotent no-op. A staged blob whose existing on-disk object is corrupt or whose digest collides with different content returns a typed integrity or collision outcome and is never overwritten.
- A commit that fails after the commit object is written but before the live commit-ref replacement leaves the new commit object addressable but unreferenced.
- A malformed or truncated object file cannot be read back. Reading returns a typed integrity outcome and does not cascade into a partial commit or partial tree.
- Duplicate writes of the same fully verified blob content never overwrite existing files. The on-disk object set accumulates one file per unique digest.
- A commit whose parent digest does not match any stored commit object is rejected.
- A root commit with an empty live index, or a child commit whose tree matches its parent's tree, is rejected unless the explicit allow-empty flag is set.
- The empty blob's known digest vector is `e69de29bb2d1d6434b8b29ae775ad8c2e48c5391`. The empty tree's digest and the empty index's digest are defined by the pinned SHA-1 input and the pinned canonical serialization; their values are computed by the implementation under test.

## 11. Project Constraints
- Local-only operation. No branches, no remotes, no network, no packfiles, no delta compression, no merge, no checkout, no credential handling, no hooks, no submodules, no mutation of a real Git repository.
- The store uses a single content-addressed object layer with zlib compression and a single small versioned index. The compression format is zlib; the kind set is fixed at `blob`, `tree`, `commit`, and `index`; the serialization is canonical and fixed.
- The store uses SHA-1 for object identity because that is Git's historical format. The store does not claim SHA-1 as a security primitive and is explicit that the choice is for study.
- Tests use temporary directories and own the work tree lifetime. No test reads from or writes to a real repository path.
- No `.git` entry is opened or read. The store's `.minigit` is the only state the project owns.
- No broad claim of Git compatibility is made. The store is an educational format whose object identity shape is shared with Git and whose operational surface is local-only.

## 12. Design Questions Before Coding
- What is the exact byte sequence that forms the SHA-1 input for `blob`, `tree`, `commit`, and `index`?
- What is the empty-blob digest vector and why is it pinned?
- How is the on-disk layout derived from the digest, and what does full verification of an existing object look like?
- How does the stage-path policy interact with absolute paths, dot segments, dot-dot segments, NUL bytes, and length?
- How does the intermediate-path symlink check reject a path whose middle component is a symlink escaping the work tree?
- How is the canonical serialization of the index, the tree, and the commit fixed in prose, and where in the code path does each format-version, field-order, and length-delimited rule live?
- How is the index sorted by raw UTF-8 byte sequence independent of supply order and map iteration?
- Why does the commit build its tree from the current live index rather than from a caller-supplied tree digest?
- What typed outcomes cover full verification, corruption, and same-digest different-content collision?
- What is the educational scope of this store, and what is explicitly not claimed about Git compatibility?

## 13. Implementation Milestones
1. Define the work tree boundary, the supplied-empty-directory check, the `.git` presence rejection, and the `.minigit` path the program creates and owns.
2. Pin the SHA-1 input bytes: kind, one space, base-10 length with no leading zero except `0`, one zero byte, content. Define the kind set `blob`, `tree`, `commit`, and `index`. Define the empty-blob digest vector.
3. Implement the on-disk layout: leading two hex characters of the digest as a subdirectory, remaining hex characters as the file name. Implement zlib compression and decompression on the object file contents.
4. Implement the existing-object full verification: decompress, verify the header kind and length, recompute the digest, compare with the file name, verify the bytes against the supplied content. Return the typed integrity or collision outcome on failure.
5. Implement the stage-path policy: relative UTF-8, non-empty, no absolute, no empty or dot or dot-dot segment, no NUL, max `255` bytes, max `16` MiB per regular file. Implement the index size cap of `10,000` paths.
6. Implement the raw-UTF-8 index sort and the exact version-1 compact JSON serialization, field ordering, string escaping, identity bounds, timestamp form, parent representation, and no-trailing-newline rules for index, tree, and commit.
7. Implement the symlink handling: store the link target's text bytes, set the symlink mode, never follow the target for blob content, reject absolute targets and targets resolving outside the work tree, and reject intermediate path components that are themselves symlinks escaping the work tree.
8. Implement the commit so it builds the tree from the current live index. The caller supplies only the author tuple, committer tuple, normalized UTC timestamp, message, optional single parent digest, and allow-empty flag.
9. Implement atomic replacement for the live index path and the live commit-ref path through temp-file-and-rename inside `.minigit`.
10. Write the unit test suite that owns temporary directories and asserts the empty-blob vector, full verification, collision reporting, stage-path policy, intermediate-symlink rejection, sorting determinism, deterministic commits with injected metadata, corrupt and truncated objects, containment, and failure-boundary preservation.
11. Verify under the race detector and reproduce the honest statement about SHA-1 choice and the absence of broad Git compatibility claims.

## 14. Verification Cases the Learner Must Write
- Empty blob known vector: feed empty bytes through the `blob` digest and assert the digest equals `e69de29bb2d1d6434b8b29ae775ad8c2e48c5391`.
- Round trip: hash a blob, read it back through the zlib-compressed on-disk file, and assert the read content equals the original bytes.
- Duplicate object storage: hash the same content twice and assert the store owns exactly one object file for that digest and the digest is unchanged.
- Existing-object full verification: re-stage the same content; assert the store decompresses the existing file, verifies the header and digest, verifies the bytes, and returns an idempotent no-op.
- Collision reporting: tamper with the on-disk bytes of an existing object without changing the file name, attempt to write the digest with the original content, and assert the store returns a typed integrity outcome and never overwrites.
- Corrupt and truncated objects: hand-craft a partially written on-disk file, attempt to read it back, assert a typed integrity outcome, and assert no partial commit or tree is produced.
- Stage-path policy: supply an absolute path, a path with a parent-traversal segment, an empty segment, a dot segment, a dot-dot segment, a NUL byte, and a path longer than `255` bytes, and assert each is rejected before any I/O.
- Index size cap: stage more than `10,000` paths and assert the operation is rejected.
- Staged regular-file bound: stage a regular file larger than `16` MiB and assert rejection.
- Intermediate-symlink rejection: create a symlink inside the work tree that escapes the root as an intermediate path component, stage a path that goes through it, and assert rejection before any blob or tree is written.
- Absolute and out-of-tree symlink rejection: create a symlink with an absolute target, and a symlink with a relative target resolving outside the work tree, stage each, and assert rejection.
- Index ordering: stage the same paths in two different orders and assert the index revision digests are identical.
- Tree ordering: build trees from the same paths in two different orders and assert the tree digests are identical.
- Deterministic commits with injected metadata: build the same commit twice with identical injected values and assert the digests are identical; build the same commit with a different timestamp and assert the digests differ; build the same commit with identical inputs except the message and assert the digests differ.
- Commit builds from the live index: stage the index, build a commit, assert that the canonical tree entries match the live index entries, and assert that tree and index digests differ because their kinds and canonical documents differ; supplying an arbitrary tree digest is not an available input.
- Empty commit policy: reject a root commit with an empty index and a child commit whose tree equals its parent's tree when allow-empty is false; accept each when allow-empty is true.
- Preservation of existing metadata on failure: drive a staging failure, a commit object write failure, and a live commit-ref replacement failure, and assert the live index, live commit-ref, and live tree pointers remain at their previous values; assert any newly created fully verified objects remain addressable but unreferenced.
- Empty tree and empty index: assert the empty tree's digest and the empty index's digest match the values defined by the pinned SHA-1 input and the pinned canonical serialization, computed by the implementation under test.
- No `.git` content access: initialize the store inside a directory that contains a `.git` entry holding foreign contents, and assert the store refuses to operate by presence only and never opens or reads the contents of the `.git` entry.
- Reads only explicitly requested paths: assert that the store reads only the paths the caller names after the containment and symlink checks pass, and that no unstaged path, host identity, environment value, or `.git` contents are read.

## 15. Common Mistakes to Watch For
- Forgetting any part of the SHA-1 input header. The kind, the ASCII space, the base-10 length with no leading zero except `0`, the zero byte, and the content are all part of the digest input.
- Using leading zeros on the base-10 length except the single digit `0`. A length prefix of `0123` is incorrect.
- Reading any work tree path that the caller did not explicitly stage. The store reads only named paths after the containment and symlink checks pass.
- Overwriting an existing object file. The store never overwrites. The store decompresses, verifies, and either accepts the idempotent no-op or returns a typed integrity or collision outcome.
- Reporting a same-digest different-content collision as success. The typed integrity or collision outcome is the only correct response.
- Reading or opening a `.git` entry's contents. The presence check is the only check; contents are never opened or read.
- Reading an unstaged path, a host identity, or an environment value. The store reads only what the test names after the containment and symlink checks pass.
- Letting the sort order depend on supply order or on map iteration. The sort is raw UTF-8 byte sequence ascending.
- Treating a stage path's symlink target as something to follow for blob content. The target's text bytes are stored under symlink mode and never followed.
- Letting an intermediate symlink escape the work tree because only the leaf path was checked. Every path component is checked.
- Letting the commit accept an arbitrary tree digest supplied by the caller. The tree is built from the current live index.
- Letting the timestamp come from the host wall clock. The timestamp is injected by the caller as a normalized UTC value.
- Calling the result a Git implementation. The store is an educational format that shares only the SHA-1 object identity shape with Git, and the documentation must say so.
- Re-using real Git's `.git` directory layout to mean that the project is Git-compatible. The store uses `.minigit`.

## 16. Topics and References for Study
- The Git internals documentation covering object identity, the SHA-1 input header, the object directory layout, and the role of compression in object files. Treat the documentation as background for the historical format and the caveats about modern digest choices.
- The zlib documentation for the compression format used on every object file.
- The UTF-8 byte comparison discipline for the index and tree sort order.
- `filepath.Clean` and the relative-path semantics for the containment check and the intermediate-path symlink check.
- Atomic file replacement patterns using a temporary file in the same directory and `os.Rename`.
- Projects 025 and 030 are the formal prerequisites: Project 025 for streaming and hashing filesystem discipline and Project 030 for safe bounded file and container integrity. Project 096 is optional immediate-catalog-predecessor context; Projects 087 and 088 are optional study for supplied-path operations and stable durable-state ordering.

## 17. Self-Assessment Questions
- What exact byte sequence forms the SHA-1 input for each object kind, and why is SHA-1 retained only for historical study?
- Why are objects immutable and fully verified before idempotent reuse, and how are corruption and same-digest different-content collisions reported?
- Why is the index content-addressed and versioned, and why are index and commit-ref replacements atomic?
- Why does raw UTF-8 byte ordering make index and tree serialization independent of supply order and map iteration?
- How do containment checks reject absolute, traversal, empty-segment, dot, dot-dot, and out-of-tree paths before I/O?
- How does the intermediate-symlink policy protect the work-tree boundary, and how are absolute or out-of-tree leaf symlinks handled?
- What does the store write inside `.minigit`, what explicitly requested paths may it read, and what host or repository data must it never read?
- Why does a commit build its tree from the live index rather than accept a caller-supplied tree digest?
- Why are caller-supplied normalized UTC metadata and the allow-empty rule needed for deterministic commit behavior?
- What does the educational, local format guarantee, and what Git compatibility or production behavior does it explicitly not claim?

## 18. Definition of Completion
The project is complete when the program accepts an existing empty directory, creates `.minigit` inside it, refuses a non-empty directory, and never opens `.git` contents; when object identity uses the exact pinned header and the empty-blob vector; when zlib-compressed immutable objects are fully verified before idempotent reuse and corruption or collision is never overwritten; when the path, file-size, index-size, symlink-target, containment, and ordering rules are enforced; when index, tree, and commit use the exact version-1 compact JSON field orders and bounds pinned in this guide; when a commit derives its tree from the live index, keeps tree and index as distinct object kinds, and applies the allow-empty rule to both an empty root and an unchanged child; when live index and commit-ref replacement is atomic; when the store writes only inside `.minigit` and reads only explicitly staged safe paths; when the learner-written tests cover vectors, round trips, verification, corruption, path and symlink safety, deterministic serialization, empty commits, parent linkage, and failure preservation; when the race detector is clean; when the documentation retains the honest SHA-1 and non-Git limitations; and when this guide contains no implementation code, signatures, starter snippets, solution snippets, pseudocode, or implementation shell commands.

## 19. Optional Extensions
- A read-only object inspection command that accepts a digest and reports the verified kind, declared byte length, and a bounded escaped summary without mutating the object store or exposing arbitrary binary content.
