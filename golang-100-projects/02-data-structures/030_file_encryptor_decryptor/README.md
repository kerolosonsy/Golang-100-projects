# Project 030 — File Encryptor and Decryptor

## 1. Project Name and Number

Project **030** — `030_file_encryptor_decryptor`. The directory name and number must match exactly. The project encrypts and decrypts bounded-size files with AES-GCM using a caller-supplied raw AES key of exactly 16, 24, or 32 bytes. Passwords are not accepted. Key derivation is not performed. The project uses a small versioned binary container, authenticates the header as associated data, and replaces the destination file only after the full operation succeeds. Decryption validates format, version, and length before opening the AEAD, authenticates before releasing plaintext, and returns a generic authentication failure for wrong key or tampering. Empty files are supported. True large-file chunk framing is outside required scope.

## 2. Project Idea

The encryptor reads a bounded input (a reader or a file), seals the contents with AES-GCM, and writes a self-describing container to a destination. The decryptor reads the container, validates its format and version, opens the AES-GCM AEAD, authenticates the header and the ciphertext, and writes the plaintext to a destination. The container carries recognizable magic and version metadata, the random nonce, and the authenticated ciphertext-and-tag. The header is included in the authenticated associated data so that tampering with the metadata fails authentication rather than being silently accepted.

The project is honest about scope. AES-GCM in Go's standard library is an AEAD: it authenticates a complete message, not a stream of chunks. For this learning project, the input size is bounded by a modest maximum the project documents, the operation reads one bounded message, and the reader/writer boundaries are exposed for testing. True large-file chunk framing (per-chunk nonces, chunk sequence numbers, partial reads across gigabytes) is outside required scope.

Key handling is direct: the caller supplies a raw AES key of exactly 16, 24, or 32 bytes. The project does not accept passwords, does not invoke a key derivation function, does not stretch a short input into a long key, and does not invent a key from a passphrase. Production nonces come from `crypto/rand`, have the length AES-GCM requires, and are never deliberately reused with the same key.

The destination file is written through a temporary file in the destination's directory. The destination is replaced with the temporary file only after the full operation succeeds. On any failure, the temporary file is removed and the destination is left unchanged. The input path is never overwritten in place.

## 3. Why This Project Now?

Projects 001–029 established variables, functions, loops, structs, errors, slices, files, JSON, CSV, scanning, sorting, walking, hashing, shape-validated matrices, generic zero-value containers, a comparator-driven BST, and a pointer-linked list. None of them used authenticated encryption or worked with binary containers. Project 030 is the project's first encounter with AEAD, binary framing, deterministic randomness injection for testing, and the discipline of writing through a temporary file before replacing the destination.

The project also forces the learner to be honest about what AES-GCM is and is not. AES-GCM authenticates a complete message. It is not a streaming cipher. Treating it as a stream by encrypting chunks independently breaks the security model. The project pins the AEAD's actual semantics and bounds the input accordingly.

The temporary-file-then-replace discipline reuses the publication pattern from Project 017 — a discipline that the path's later projects (for example, the database-migration and event-driven projects) also rely on.

## 4. Prerequisites

Per the dependency map in `plan.md`, Project 030 formally requires:

- Completion of **029** (Linked List Implementation). The encryptor and decryptor are the next project after the linked list implementation and are the first project in Level 2 that uses authenticated encryption and a versioned binary container.
- Completion of **017** (JSON Todo Persister). The file-level decryptor writes the plaintext through a temporary file in the destination's directory and replaces the destination only after the full operation succeeds. This is the same safe-save publication discipline that Project 017 established for the JSON todo persister — write to a temporary file in the same directory, close with the close error checked, then rename over the target. Project 030 reuses the discipline for an authenticated binary container; the publication step is the same.

Project 025 (File Duplicate Finder) provides review/background context already encountered for streaming-I/O discipline over a `bytes.Buffer` or a file. It is not a formal prerequisite for Project 030; it informs the learner's design.

No prior knowledge of HTTP, databases, or concurrency.

## 5. What You Must Know Before Starting

- That AES is a block cipher. AES-GCM is an authenticated encryption mode that produces a ciphertext and an authentication tag, and verifies the tag before releasing the plaintext.
- That AES-GCM is an AEAD, not a stream cipher. The standard library's AEAD interface seals an entire plaintext (with optional associated data) in one call and opens the entire ciphertext in one call. The interface does not expose per-chunk or per-block operations.
- That the AEAD's nonce has a fixed length, supplied by the AEAD itself. The project must use that length, never a substitute.
- That the AEAD's tag is appended to the ciphertext by the standard library. The encryptor and decryptor must agree on the tag's position and length.
- That the AEAD's associated data is authenticated but not encrypted. The header is included as associated data so that tampering with the header fails authentication.
- That `crypto/rand` produces cryptographically secure random bytes and is the appropriate source for nonces. `math/rand` is not appropriate.
- That `crypto/aes` rejects keys that are not 16, 24, or 32 bytes. The encryptor and decryptor must validate the key length before constructing the cipher.
- That `os.CreateTemp` creates a file in the named directory. The temporary file's name is unique within the directory. The caller is responsible for closing and removing the temporary file on failure.
- That `os.Rename` is the standard atomic replace on most platforms but is not guaranteed to be atomic on every filesystem. The project documents the platform/filesystem limits honestly.
- That reading and writing through readers and writers (`io.Reader`, `io.Writer`) lets the operation be tested with in-memory buffers without touching the real filesystem for the core logic.

## 6. Explanation of New Concepts

### AES-GCM as an AEAD, not a stream cipher

The standard library's `cipher.AEAD` interface seals an entire plaintext in one call and opens the entire ciphertext in one call. The interface takes a nonce and an optional associated-data buffer. The encryptor reads the entire bounded input, calls `Seal`, and writes the resulting ciphertext-plus-tag to the destination. The decryptor reads the entire container, calls `Open`, and writes the plaintext to the destination.

The interface is not a streaming interface. A naive approach that reads chunks and seals each chunk independently produces ciphertexts that are not authenticatable as a whole: each chunk's tag authenticates only that chunk, not the others. A correct AEAD-based design either treats the input as one bounded message (the project's choice) or designs a chunking layer that maintains authentication across chunks. The latter is outside required scope.

### Bounded input size

The project imposes a maximum plaintext size of exactly **8 MiB** (8,388,608 bytes). The maximum is documented in the package, is consistent across encrypt and decrypt, and is small enough that the operation can hold the plaintext in memory and seal it in one AEAD call. The maximum is not a streaming limit; it is a "we read everything, then seal it" limit. The decryptor rejects any container whose declared plaintext length exceeds 8 MiB and any container whose total byte size exceeds 8 MiB plus the fixed documented overhead (magic, version, declared-length, nonce, and tag). True large-file support would require a chunked AEAD framing protocol that maintains authentication across chunks. That protocol is its own design problem and is outside required scope. Substituting a different AEAD construction does not solve the framing problem on its own; the framing is what has to be designed.

### Caller-supplied raw key

The caller supplies a raw AES key. The key must be exactly 16, 24, or 32 bytes. The project validates the key length and returns an error for any other length. The project does not accept passwords. The project does not run a key derivation function. The project does not stretch, hash, or transform the key.

### Production nonces from `crypto/rand`

Production nonces come from `crypto/rand`. The nonce length is whatever the AEAD reports as its required nonce size. Nonces are never deliberately reused with the same key. Reusing a nonce with the same key catastrophically breaks AES-GCM's security; the project pins the rule that each encryption uses a fresh nonce from `crypto/rand`.

### Deterministic randomness for tests

Tests need to exercise specific nonce values, specific error paths, and specific tampered regions. A test that depends on `crypto/rand` cannot pin those values. The project exposes a randomness source that tests can replace. Production wires `crypto/rand`; tests wire a deterministic source that returns planned bytes. The test's behavior is observable through the bytes written into the nonce field of the container. Tests do not use "two random outputs differ" as a proof.

### Versioned binary container (pinned contract)

The container has a recognizable magic, a version, a declared plaintext length, a nonce, and the authenticated ciphertext-and-tag. The conceptual layout is:

- **Magic** — four ASCII bytes that spell `G30F`. The decryptor recognizes the format by these four bytes.
- **Version** — one byte, value `1`. The decryptor recognizes the layout by this version.
- **Declared plaintext length** — eight bytes, interpreted as an unsigned big-endian integer, giving the byte length of the plaintext that was encrypted. The value is in the range `[0, 8388608]`. The value is authenticated as part of the header.
- **Nonce** — exactly the number of bytes that AES-GCM reports as its required nonce size. The decryptor reads exactly that many bytes.
- **Ciphertext-and-tag** — the bytes produced by AES-GCM `Seal`, which are the encrypted plaintext followed by the appended tag.

The bytes `G30F`, the version byte `1`, and the eight declared-length bytes together form the **authenticated header**. The nonce and the ciphertext-and-tag are not part of the associated data. The encryptor writes the header, the nonce, then the ciphertext-and-tag. The decryptor reads them in the same order and authenticates the header bytes alongside the ciphertext-and-tag.

The README pins this contract without prescribing the exact byte-by-byte procedure. The learner's implementation follows the contract; tests pin it through round-trip and tamper-detection cases.

### Header authentication as associated data

The header — `G30F`, the version byte, and the eight declared-length bytes — is the AEAD's associated data. The encryptor passes the header to `Seal` as associated data. The decryptor passes the header to `Open` as associated data. Tampering with the header bytes causes `Open` to fail authentication, not silently produce garbage plaintext.

### Structural format errors vs authenticated-header tampering

These are two distinct categories and the tests distinguish them.

**Structural format errors.** The decryptor detects these before opening the AEAD and returns a clear structural error:

- The container is too short to contain the magic, version, declared-length, and nonce.
- The four magic bytes are not `G30F`.
- The version byte is not `1`.
- The declared plaintext length is not in `[0, 8388608]`.
- The container's total byte size exceeds `8388608 + fixed overhead`, where the fixed overhead is the nonce length plus the tag length plus the header bytes.
- The container is too short to contain the declared plaintext length plus the tag.

These are fail-fast errors: invalid input is rejected with a clear message before any authentication is attempted.

**Authenticated-header tampering.** When the magic, version, and declared-length are all structurally valid but the declared-length bytes are then changed to another value within `[0, 8388608]`, the change is not a structural error — it is a change to authenticated associated data. The decryptor calls `Open` with the modified header and the AEAD fails authentication. The decryptor returns the same generic authentication failure it returns for wrong key, modified nonce, modified ciphertext, or modified tag. The decryptor does not distinguish "modified declared length" from "wrong key" from "tampered ciphertext".

### Authentication before releasing plaintext

The decryptor calls `Open` and inspects the error. A non-`nil` error from `Open` means the ciphertext was tampered with, the associated data was tampered with, or the key is wrong. The decryptor does not distinguish among these cases; it returns a generic authentication failure. The decryptor must not release any plaintext on authentication failure. The decryptor must not write partial plaintext to the destination.

### Post-authentication length check

After `Open` succeeds, the decryptor compares the number of plaintext bytes it produced against the declared plaintext length that was authenticated in the header. If the two differ, the decryptor returns a format/integrity error and publishes no plaintext. This check protects against a sealed plaintext whose length does not match the declared length, even though AEAD authentication passed.

### Wrong key as generic authentication failure

A wrong key produces the same generic authentication failure as tampering. The decryptor does not distinguish "wrong key" from "tampered ciphertext" or "tampered authenticated header". This is honest: distinguishing the cases leaks information to an attacker.

### All-or-nothing reader/writer decryption

For reader/writer decryption, the decryptor reads the container, validates the structural format, opens the AEAD with the header as associated data, calls `Open`, runs the post-authentication length check, and only then writes the plaintext to the destination. If any step fails, the destination receives no bytes. The first byte is written only after authentication and the post-authentication length check have both succeeded.

### Temporary file then replace (file-level decryption)

For file-level decryption, the decryptor creates a temporary file in the destination's directory, runs the all-or-nothing reader/writer decryption into that temporary file, and only after every write has succeeded and the temporary file has been closed does it rename the temporary file into place. If any step fails before the rename, the temporary file is removed and the destination is left unchanged. The input path is never overwritten in place.

### Normal-operation atomic visibility and filesystem limits

During normal operation on a single process and a single filesystem, the destination either contains the full plaintext produced by the decryptor or it contains the pre-operation content. The project documents that the temporary-file-then-rename discipline provides this normal-operation atomic visibility. The project does not claim universal crash durability. `os.Rename` is atomic on most Unix-like systems and on Windows for files on the same volume. It is not atomic across volumes, across filesystems, or in the presence of certain failure modes (power loss between rename and sync, for example). The tests verify normal-operation behavior on a single process and a single filesystem.

### Empty files

Empty files are supported. The plaintext is zero bytes. The declared plaintext length is zero. The nonce is freshly generated. The container is written. On decryption, the container is read, the structural format is validated, the AEAD is opened with the empty plaintext, and an empty plaintext is written to the destination.

### No partial output on failure

If any step fails (invalid key, oversized input, structural format error, randomness failure, generic authentication failure, post-authentication length mismatch), the decryptor does not write any partial plaintext to the destination. For reader/writer decryption, the destination receives no bytes. For file-level decryption, the temporary file is removed and the destination file is left unchanged.

### No password prompts, KDF, compression, directory recursion, network transport, custom cryptography, or production key management

The project does not prompt for passwords. The project does not run a key derivation function. The project does not compress. The project does not walk directory trees or recurse into subdirectories. The project does not transmit data over a network. The project does not implement a custom cipher. The project does not provide production-grade key management (key rotation, key escrow, key storage).

## 7. Learning Objective

After completing this project the learner can:

- Explain why AES-GCM is an AEAD that authenticates a complete message, and why it is not a general streaming cipher.
- Use `crypto/aes` and `crypto/cipher` to construct an AES-GCM AEAD with a raw key of exactly 16, 24, or 32 bytes, and reject other key lengths with a contextual error.
- Generate a nonce from `crypto/rand` with the length reported by the AEAD, and never reuse a nonce with the same key.
- Implement the pinned container contract: four ASCII magic bytes `G30F`, version byte `1`, eight-byte unsigned big-endian declared plaintext length, the AEAD's required nonce length, then the ciphertext-and-tag. The header (magic + version + declared length) is the AEAD's associated data.
- Enforce the 8 MiB (8,388,608 bytes) plaintext maximum on both encryption and decryption, including rejecting containers whose declared length exceeds 8 MiB and whose total byte size exceeds 8 MiB plus the fixed documented overhead.
- Distinguish structural format errors (rejected before AEAD `Open` with a clear message) from authenticated-header tampering (returned as the generic authentication failure, indistinguishable from wrong key, modified nonce, modified ciphertext, or modified tag).
- Run the post-authentication length check: if the plaintext produced by `Open` does not match the declared length, return a format/integrity error and publish no plaintext.
- Run reader/writer decryption all-or-nothing: read, validate structure, open the AEAD with the header as associated data, run `Open`, run the post-authentication length check, and only then write the first plaintext byte.
- For file-level decryption, write to a temporary file in the destination's directory, close it, then rename it into place. On failure, remove the temporary file and leave the destination unchanged.
- Inject a deterministic randomness source for tests, observe its effect through the nonce bytes written into the container, and exercise nonce-related error paths without relying on "two random outputs differ".
- Explain what normal-operation atomic visibility means on common platforms, what it does not promise, and what the project does not claim about crash durability.
- Write tests that cover round-trip for all key sizes and empty/binary content, deterministic known nonce through injection, invalid sizes/keys/format, structural format errors, authenticated-header tampering, all tamper regions, wrong key, post-authentication length mismatch, no partial output, temp cleanup, and preservation of an existing destination on failure.

## 8. Functional Requirements

1. The package exposes an encrypt operation and a decrypt operation. Each accepts a source (reader or path), a destination (writer or path), and a raw key (exactly 16, 24, or 32 bytes).
2. The encrypt operation writes a container whose header is the four magic bytes `G30F`, the version byte `1`, and eight unsigned big-endian declared-plaintext-length bytes. The encrypt operation reads the source into memory, validates that the plaintext size is at most 8 MiB (8,388,608 bytes), seals the plaintext with AES-GCM using a fresh nonce from `crypto/rand`, authenticates the header as associated data, and writes the container to the destination.
3. The decrypt operation reads the container, validates its structural format (magic, version, declared length in `[0, 8388608]`, total size within `8388608 + fixed overhead`, and enough bytes for the fixed nonce plus the authentication tag), opens the AEAD with the header as associated data, calls `Open`, runs the post-authentication plaintext-length check, and writes the plaintext to the destination. It deliberately does not compare the ciphertext length with the declared plaintext length before `Open`, because doing so would misclassify an in-range declared-length modification as a structural error instead of authenticated-header tampering.
4. The encrypt operation rejects keys that are not 16, 24, or 32 bytes with an error naming the offending key length and the valid lengths.
5. The decrypt operation rejects keys that are not 16, 24, or 32 bytes with an error naming the offending key length and the valid lengths.
6. The encrypt operation rejects sources whose size exceeds 8 MiB with an error naming the source and the maximum.
7. The decrypt operation rejects containers with structural format errors (wrong magic, unsupported version, declared length out of `[0, 8388608]`, total size exceeding `8388608 + fixed overhead`, truncated body) with a clear structural error before opening the AEAD.
8. The decrypt operation returns a generic authentication failure for wrong key, modified nonce, modified ciphertext, modified tag, and authenticated-header tampering (such as changing the declared-length bytes to another in-range value). The decrypt operation does not distinguish among these cases.
9. The decrypt operation runs a post-authentication length check. If the plaintext produced by `Open` does not equal the declared plaintext length, the decrypt operation returns a format/integrity error and writes no plaintext.
10. The decrypt operation does not write any partial plaintext to the destination on any failure (invalid key, oversized input, structural format error, generic authentication failure, post-authentication length mismatch). Reader/writer destinations receive no bytes. File destinations are unchanged.
11. File-level writes use a temporary file in the destination's directory. The decrypt operation writes to the temporary file only after authentication and the post-authentication length check have succeeded. The temporary file is closed before the rename. The destination is replaced only after the rename succeeds. On failure, the temporary file is removed and the destination is left unchanged.
12. The input path is never overwritten in place. The encrypt operation reads from the source and writes to the destination; the decrypt operation reads from the source and writes to the destination. Source and destination are distinct.
13. Empty files are supported. The encrypt operation produces a container with declared length `0` and an empty ciphertext. The decrypt operation produces an empty plaintext.
14. Production nonces come from `crypto/rand` and have the AEAD-required length. Nonces are never deliberately reused with the same key. The package documentation states this rule.
15. The randomness source is injectable. Production wires `crypto/rand`. Tests wire a deterministic source that returns planned bytes.
16. The package documentation states the AES-GCM semantics (AEAD, not stream cipher), the 8 MiB plaintext maximum, the key-length validation, the container contract (`G30F` magic, version `1`, eight-byte big-endian declared length, AEAD-required nonce length, ciphertext-and-tag), the header-as-associated-data rule, the distinction between structural format errors and authenticated-header tampering, the post-authentication length check, the no-distinguish rule for authentication failures, the all-or-nothing reader/writer decryption rule, the temporary-file-then-replace discipline, the normal-operation atomic visibility rule, and the platform/filesystem limits.

## 9. Inputs and Outputs

### Inputs

- A source: an `io.Reader` or a file path containing the plaintext (for encrypt) or the container (for decrypt).
- A destination: an `io.Writer` or a file path to receive the container (for encrypt) or the plaintext (for decrypt).
- A raw key: exactly 16, 24, or 32 bytes.
- A randomness source: injectable. Production uses `crypto/rand`. Tests use a deterministic source.

### Outputs

- For encrypt: the number of bytes written to the destination, or an error. Errors include invalid key length, oversized input, randomness failure, and write failure.
- For decrypt: the number of bytes written to the destination, or an error. Errors include invalid key length, invalid format (missing magic, unknown version, truncated body, invalid nonce length), generic authentication failure (wrong key, tampering), and write failure.
- The destination file is replaced only on successful operation. On failure, the destination is unchanged.

### Example text-only round-trip

```
Source plaintext (UTF-8): "hello, world" (12 bytes)
Plaintext maximum: 8 MiB (8,388,608 bytes)
Key: 16 random bytes (deterministic in tests)

Encrypt:
  container bytes:
    magic       (4 ASCII bytes: G30F)
    version     (1 byte: 1)
    declared    (8 bytes big-endian: 12)
    nonce       (12 bytes, from crypto/rand)
    ciphertext  (12 bytes plaintext + 16 bytes tag = 28 bytes)

Decrypt (same key):
  plaintext: "hello, world" (12 bytes)

Decrypt (different key):
  error: authentication failure.

Decrypt (declared length bytes changed to another in-range value):
  error: authentication failure.   (authenticated-header tampering, indistinguishable from wrong key)
```

### Example text-only error cases

```
Encrypt:
  error: invalid key length 10, expected 16, 24, or 32.
  error: input exceeds maximum size of 8388608 bytes.
  error: randomness source failed.

Decrypt — structural format errors (rejected before AEAD Open):
  error: invalid container: wrong magic.
  error: invalid container: unsupported version.
  error: invalid container: declared length 99999999 exceeds maximum 8388608.
  error: invalid container: total size exceeds maximum plus overhead.
  error: invalid container: truncated body.

Decrypt — authenticated-header tampering and other authentication failures:
  error: authentication failure.

Decrypt — post-authentication length mismatch:
  error: plaintext length does not match declared length.
```

## 10. Rules and Edge Cases

- **Key length.** The key must be exactly 16, 24, or 32 bytes. Any other length returns an error naming the offending length and the valid lengths.
- **Plaintext maximum.** The plaintext must be at most 8 MiB (8,388,608 bytes). Encryption rejects inputs above this size with an error naming the source and the maximum. Decryption rejects any container whose declared length is above this maximum and any container whose total byte size is above `8388608 + fixed overhead` (the nonce length plus the tag length plus the header bytes). Empty plaintext (declared length `0`) is valid.
- **Empty files.** Encrypt produces a container with declared length `0` and an empty ciphertext (only the tag follows the nonce). Decrypt produces an empty plaintext.
- **Nonce length.** The nonce has the AEAD-required length, sourced from `crypto/rand` (or the injected source in tests). The decryptor reads exactly that length and rejects a container that does not have it.
- **Nonce freshness.** Each encrypt operation uses a fresh nonce. The same key with two encryptions produces two different nonces. The project does not deliberately reuse nonces.
- **Magic, version, declared length.** The encryptor writes the four magic bytes `G30F`, the version byte `1`, and the eight declared-length bytes. The decryptor validates all three before opening the AEAD. Wrong magic, unsupported version, and out-of-range declared length are structural format errors and are rejected with a clear message before the AEAD is opened.
- **Truncated container.** A container that cannot contain the header plus the nonce plus the tag is rejected as a structural format error before the AEAD is opened.
- **Total size limit.** A container whose total byte size exceeds `8388608 + fixed overhead` is rejected as a structural format error before the AEAD is opened.
- **Authenticated-header tampering.** Changing the declared-length bytes to another value within `[0, 8388608]` is not a structural error. The decryptor calls `Open` with the modified header and the AEAD fails authentication. The decryptor returns the same generic authentication failure used for wrong key, modified nonce, modified ciphertext, and modified tag.
- **Ciphertext authentication.** Tampering with any byte of the ciphertext or tag causes the AEAD to fail authentication.
- **Nonce tampering.** Tampering with any byte of the nonce causes the AEAD to fail authentication.
- **Wrong key.** A key that does not match the encryption key causes the AEAD to fail authentication. The decryptor returns a generic authentication failure, indistinguishable from tampering or authenticated-header tampering.
- **No distinguish.** The decryptor does not distinguish among "wrong key", "tampered authenticated header", "tampered nonce", "tampered ciphertext", and "tampered tag". All return the same generic authentication failure.
- **Post-authentication length check.** After `Open` succeeds, the decryptor compares the plaintext length against the declared length. A mismatch is a format/integrity error. No plaintext is published.
- **No partial output on failure.** On any failure, the decryptor does not write any plaintext to the destination. Reader/writer destinations receive no bytes. File destinations are unchanged (either untouched or in their pre-operation state).
- **All-or-nothing reader/writer decryption.** The first plaintext byte is written only after the structural format check, the AEAD `Open`, and the post-authentication length check have all succeeded.
- **Temporary file then replace.** File-level writes go through a temporary file in the destination's directory. The destination is replaced only after every write and the close of the temporary file have succeeded. On failure, the temporary file is removed and the destination is left unchanged.
- **Input path unchanged.** The encrypt operation reads from the source and writes to the destination. The source is never modified. The destination is a distinct path.
- **Normal-operation atomic visibility.** During normal operation on a single process and a single filesystem, the destination either contains the full plaintext produced by the decryptor or it contains the pre-operation content. The project does not claim universal crash durability. The tests verify normal-operation behavior on a single process and a single filesystem.
- **Deterministic randomness for tests.** Tests inject a deterministic randomness source. The tests observe the effect through the nonce bytes written into the container. Tests do not use "two random outputs differ" as a proof.
- **No password prompts, KDF, compression, directory recursion, network transport, custom cryptography, or production key management.**

## 11. Project Constraints

- Go standard library only. No third-party cryptography libraries. The implementation uses `crypto/aes`, `crypto/cipher`, `crypto/rand`, `errors`, `fmt`, `io`, `os`, `path/filepath` (where relevant), and the testing package.
- AES-GCM only. No other cipher mode is accepted in the required scope. No custom cipher.
- The caller supplies a raw AES key. The project does not accept passwords. The project does not run a key derivation function.
- The randomness source is injectable. Production wires `crypto/rand`. Tests wire a deterministic source.
- The plaintext maximum is exactly 8 MiB (8,388,608 bytes). Encryption and decryption both enforce it. Decryption also rejects containers whose total byte size exceeds `8388608 + fixed overhead`. True large-file support requires a chunked AEAD framing protocol and is out of scope.
- The container contract is pinned: four ASCII magic bytes `G30F`, version byte `1`, eight unsigned big-endian declared-plaintext-length bytes, the AEAD's required nonce length, then the ciphertext-and-tag.
- The header (`G30F` + version + declared length) is the AEAD's associated data. The nonce and the ciphertext-and-tag are not.
- File-level writes use a temporary file in the destination's directory. The destination is replaced only on success. The temporary file is closed before the rename.
- Reader/writer decryption is all-or-nothing: the first plaintext byte is written only after the structural format check, the AEAD `Open`, and the post-authentication length check have all succeeded.
- Structural format errors (wrong magic, unsupported version, out-of-range declared length, total size exceeding the documented ceiling, truncated body) are detected and returned before `Open`. Authenticated-header tampering is returned as the generic authentication failure, indistinguishable from wrong key, modified nonce, modified ciphertext, and modified tag.
- The project does not compress. The project does not recurse into directories. The project does not transmit over a network. The project does not implement custom cryptography. The project does not provide production key management.
- Core logic is testable through reader/writer boundaries. Tests use temporary directories for file-level cases.

## 12. Design Questions Before Coding

- How is the header laid out as associated data? The contract is `G30F` (4 bytes), version `1` (1 byte), declared length (8 bytes big-endian). How does the implementation produce and consume these bytes so the header is identical on both sides?
- How is the structural format check ordered against the AEAD `Open`? Magic, version, declared length, total size, then nonce and ciphertext presence, then `Open`? Which ordering keeps structural errors distinct from authenticated-header tampering?
- How is the declared length validated? Against `[0, 8388608]` before `Open`, against the actual plaintext length after `Open`. Which choice makes the two checks unambiguous?
- How is the randomness source injected? Through an interface the encrypt operation accepts, through a package-level variable tests can replace, or through a constructor that takes the source? Which choice keeps the production wiring simple and the test wiring obvious?
- How is the key length validated? Before constructing the cipher, as a precondition for every encrypt and decrypt, or inside a shared helper? Which choice ensures the error is returned with consistent naming?
- How is reader/writer decryption made all-or-nothing? The first destination byte is written only after `Open` and the post-authentication length check succeed. Which implementation pattern guarantees this?
- How is the temporary file managed? Created with `os.CreateTemp` in the destination's directory, written through the same all-or-nothing reader/writer path, closed, then renamed into place. Which choice ensures the destination is unchanged on failure?
- How is the no-distinguish rule enforced for authenticated-header tampering, wrong key, modified nonce, modified ciphertext, and modified tag? A single error type and message with no per-failure-mode wrapping. Which choice keeps the error message honest?
- How are tests structured? Round-trip tests for all key sizes, structural-format-error tests, authenticated-header-tampering tests, tamper tests for the nonce/ciphertext/tag regions, post-authentication length mismatch tests, and file-level tests through temporary directories. Which choice keeps the test coverage clear?

## 13. Implementation Milestones

1. Decide the package layout. Keep the container format, the encrypt operation, and the decrypt operation in the same package. Keep `main` as a thin driver that exercises encrypt and decrypt on a small file in a temporary directory.
2. Pin the public contract as named constants: the four magic bytes `G30F`, the version byte `1`, the plaintext maximum of 8 MiB (8,388,608 bytes), the fixed overhead (header bytes plus nonce length plus tag length), the AEAD nonce length (read from the AEAD), and the error sentinels or error types for structural format errors, generic authentication failure, and post-authentication length mismatch.
3. Implement the randomness source seam. Define an interface the encrypt operation accepts. Production wires `crypto/rand`. Tests wire a deterministic source.
4. Implement encryption input handling. Validate the key length and the input size against the 8 MiB maximum, then read the source into memory. Generate a fresh nonce from the randomness source.
5. Complete the encrypt operation. Build the header (`G30F`, version `1`, eight-byte big-endian declared length). Call `Seal` with the header as associated data. Write the header, the nonce, then the ciphertext-and-tag to the destination.
6. Implement decryption input handling. Validate the key length. Read the container. Run the structural format checks (magic, version, declared length in `[0, 8388608]`, total size within `8388608 + fixed overhead`, nonce and ciphertext present) before opening the AEAD.
7. Complete the decrypt operation. Open the AEAD with the header as associated data. Return the same generic authentication failure for wrong key and authenticated tampering. Run the post-authentication length check and return its distinct format/integrity error on mismatch. Write the plaintext only after every check succeeds; on any failure, write nothing.
8. Implement the temporary-file-then-replace discipline for file-level writes. The decrypt operation, when given paths, creates a temporary file in the destination's directory, runs the all-or-nothing reader/writer decryption into that temporary file, closes it, and renames it into place. On failure, the temporary file is removed and the destination is left unchanged.
9. Wire `main`. The driver exercises a small round-trip on a temporary file. The driver is not part of the package's public contract.
10. Add tests for every verification case in section 14, including all key sizes, empty and maximum-size content, structural errors, authenticated-header tampering, all other tamper regions, wrong keys, post-authentication length mismatch, no partial output, temporary cleanup, preservation of an existing destination, and file-level round trips.

## 14. Verification Cases the Learner Must Write

Each case is described in natural language. Tests use reader/writer boundaries for core logic and temporary directories for file-level cases.

### Round trip for all key sizes and content types

- Encrypt then decrypt a small plaintext with a 16-byte key. The decrypted plaintext equals the original byte-for-byte.
- Encrypt then decrypt a small plaintext with a 24-byte key. The decrypted plaintext equals the original.
- Encrypt then decrypt a small plaintext with a 32-byte key. The decrypted plaintext equals the original.
- Encrypt then decrypt binary content (every byte value from `0x00` to `0xFF`). The decrypted content equals the original.
- Encrypt then decrypt a plaintext whose size equals the documented 8 MiB maximum. The decrypted plaintext equals the original.

### Round trip for empty content

- Encrypt an empty plaintext. The container is written with declared length `0` and empty ciphertext.
- Decrypt that container. The decrypted plaintext is empty.
- Round trip is exact for all three key sizes.

### Plaintext maximum enforcement

- Encrypt a plaintext whose size is exactly 8 MiB (8,388,608 bytes). The encrypt operation succeeds.
- Encrypt a plaintext whose size is 8 MiB plus one byte. The encrypt operation returns an error naming the source and the maximum.
- Decrypt a container whose declared length is `8388609`. The decrypt operation returns a structural format error naming "declared length exceeds maximum" before `Open`.
- Decrypt a container whose total byte size exceeds `8388608 + fixed overhead`. The decrypt operation returns a structural format error before `Open`.
- Empty input is valid and is not rejected as oversized.

### Deterministic known nonce through injection

- The encrypt operation accepts an injected randomness source. A test wires a source that returns a fixed sequence of bytes. The nonce written into the container matches the planned bytes.
- The decrypt operation does not depend on the randomness source. A test uses the same key and the container produced by the deterministic-nonce encryption. Decryption succeeds.
- The test exercises a nonce-related error path by tampering with the nonce bytes. The decryptor returns the generic authentication failure.

### Invalid keys

- Encrypt with a 10-byte key returns an error naming the offending length and the valid lengths.
- Encrypt with a 33-byte key returns an error naming the offending length and the valid lengths.
- Decrypt with a 10-byte key returns an error naming the offending length and the valid lengths.
- Decrypt with a 33-byte key returns an error naming the offending length and the valid lengths.

### Structural format errors (rejected before AEAD Open)

- Decrypt a container whose magic bytes are not `G30F`. The decryptor returns a structural format error naming "wrong magic" before `Open`. No plaintext is written.
- Decrypt a container whose version byte is not `1`. The decryptor returns a structural format error naming "unsupported version" before `Open`. No plaintext is written.
- Decrypt a container whose declared length bytes are not a valid unsigned big-endian integer in `[0, 8388608]` (for example, a value above 8388608). The decryptor returns a structural format error naming "declared length out of range" before `Open`. No plaintext is written.
- Decrypt a container whose total byte size exceeds `8388608 + fixed overhead`. The decryptor returns a structural format error naming "total size exceeds maximum plus overhead" before `Open`. No plaintext is written.
- Decrypt a container whose body is shorter than `header + nonce + tag`. The decryptor returns a structural format error naming "truncated body" before `Open`. No plaintext is written.

### Authenticated-header tampering (generic authentication failure)

- A test takes a valid container and changes the declared-length bytes to another in-range value within `[0, 8388608]`. The header is structurally valid; the modification is authenticated-header tampering. The decryptor returns the generic authentication failure. No plaintext is written.
- The authenticated-header-tampering error message is the same as the wrong-key, modified-nonce, modified-ciphertext, and modified-tag error messages.

### Other tamper regions

- A test takes a valid container and modifies one byte of the nonce. Decryption returns the generic authentication failure. No plaintext is written.
- A test takes a valid container and modifies one byte of the ciphertext. Decryption returns the generic authentication failure. No plaintext is written.
- A test takes a valid container and modifies one byte of the tag. Decryption returns the generic authentication failure. No plaintext is written.

### Wrong key

- A test encrypts with one key and decrypts with a different key of the same length. Decryption returns the generic authentication failure. No plaintext is written.
- The wrong-key error message is the same as the tamper error messages.

### Post-authentication length mismatch

- A test constructs a container whose authenticated plaintext length differs from the declared length. `Open` succeeds; the post-authentication length check fails. The decryptor returns a format/integrity error. No plaintext is written.
- The format/integrity error is distinct from the structural format errors and from the generic authentication failure.

### No partial output

- A test forces an authentication failure and confirms the destination file (or writer) has zero bytes written. For writer-based tests, the writer's content is unchanged. For file-based tests, the destination file's content is unchanged from before the operation.
- A test forces a structural format error and confirms the destination receives no bytes.
- A test forces a post-authentication length mismatch and confirms the destination receives no bytes.
- A test forces an invalid-key error and confirms the destination receives no bytes.

### Temp cleanup

- A test forces a failure during encryption and confirms the temporary file in the destination's directory is removed after the operation.
- A test forces a failure during decryption (structural error, generic authentication failure, post-authentication length mismatch) and confirms the temporary file is removed.
- A test confirms the destination directory does not accumulate temporary files across multiple failing operations.

### Preservation of existing destination on failure

- A test creates an existing destination file with known content. The test then runs a failing decrypt operation against the destination. After the failure, the destination file's content is unchanged.
- A test confirms the destination file's modification time is unchanged after a failing operation.

### File-level round trip

- A test creates a temporary source file with known content. The test encrypts the source to a temporary destination. The test then decrypts the destination to another temporary destination. The final destination's content equals the original source's content byte-for-byte.
- The test exercises all three key sizes through the file-level path.

### Process

- A test runs the driver against a small source file and confirms the round-trip succeeds.
- A test runs the driver with a wrong key and confirms the driver exits with a non-zero status and the error message names the failure mode.

## 15. Common Mistakes to Watch For

- **Treating AES-GCM as a stream cipher.** Reading chunks and sealing each chunk independently produces ciphertexts that are not authenticatable as a whole. The project reads the entire bounded input and seals it in one `Seal` call.
- **Using a fixed nonce.** Reusing a nonce with the same key catastrophically breaks AES-GCM. The encrypt operation must use a fresh nonce from `crypto/rand` (or the injected source in tests) for every call.
- **Using `math/rand` for nonces.** `math/rand` is not cryptographically secure. The nonce must come from `crypto/rand`.
- **Accepting passwords or running a KDF.** The project does not accept passwords. The project does not run a key derivation function. The caller supplies the raw key.
- **Accepting keys of arbitrary length.** `crypto/aes` rejects keys that are not 16, 24, or 32 bytes. The encrypt and decrypt operations validate the key length explicitly and return a clear error for other lengths.
- **Treating structural format errors and authenticated-header tampering as the same category.** They are distinct. Wrong magic, unsupported version, declared length out of `[0, 8388608]`, total size above `8388608 + fixed overhead`, and truncated body are structural errors returned before `Open`. Changing the declared-length bytes to another in-range value is authenticated-header tampering, returned as the generic authentication failure. The tests must distinguish them.
- **Distinguishing authentication failures.** Wrong key, authenticated-header tampering, modified nonce, modified ciphertext, and modified tag all return the same generic authentication failure. The decryptor does not say "wrong key" or "tampered header" or "tampered ciphertext".
- **Skipping the post-authentication length check.** After `Open` succeeds, the plaintext length must equal the declared length. Skipping the check can release a plaintext whose length does not match the header. The check is required and produces a distinct format/integrity error.
- **Releasing partial plaintext on failure.** The decryptor does not write any plaintext to the destination on authentication failure. The destination is unchanged.
- **Writing the destination in place.** The project writes through a temporary file in the destination's directory and replaces the destination only after the full operation succeeds. Writing directly to the destination is wrong.
- **Overwriting the input path.** The encrypt operation reads from the source and writes to the destination. The source is never modified. Source and destination are distinct.
- **Leaving temporary files behind on failure.** The project removes the temporary file on failure. Leaving it behind pollutes the destination's directory.
- **Skipping magic, version, declared-length, or total-size validation.** The decryptor must validate all structural format checks before opening the AEAD. Skipping validation can lead to opaque authentication failures when the input is simply malformed.
- **Skipping associated data.** The header is part of the AEAD's input. Passing an empty associated data to `Seal` or `Open` defeats header authentication.
- **Using "two random outputs differ" as a proof in tests.** Tests inject a deterministic source and observe the planned bytes. A test that calls `crypto/rand` twice and asserts the outputs differ is not a valid proof.
- **Reading the source twice.** The encrypt operation reads the source once into memory. Reading twice (once for size, once for content) is wasteful and can produce inconsistent results on a changing source.
- **Claiming universal crash durability.** Atomic replace is not universal. The project documents the platform/filesystem limits honestly.
- **Including compression, directory recursion, or network transport.** The project's scope is bounded. Adding these is out of scope.

## 16. Topics and References for Study

- A Tour of Go: "Errors", "Readers", "Testing".
- Effective Go: "Errors", "Data".
- Package documentation: `crypto/aes` (NewCipher), `crypto/cipher` (AEAD, NewGCM, Seal, Open, NonceSize, Overhead), `crypto/rand` (Read), `errors` (New, Is), `fmt` (Errorf, %w), `io` (Reader, Writer, Copy), `os` (CreateTemp, Rename, Remove, Stat), `path/filepath` (Dir).
- AEAD patterns: search for "Go AES-GCM AEAD seal open", "Go associated data header authentication", "Go nonce length GCM".
- Atomic file replacement: search for "Go os.CreateTemp os.Rename atomic write", "Go temp file destination replace", "Go crash safe write".
- Deterministic randomness for tests: search for "Go crypto/rand injectable test", "Go deterministic nonce test", "Go crypto rand interface seam".
- Tamper detection patterns: search for "Go AEAD tamper detect", "Go GCM authentication failure", "Go associated data tamper".

## 17. Self-Assessment Questions

1. Why is AES-GCM an AEAD that authenticates a complete message, and why is it not a general streaming cipher?
2. Why must the plaintext maximum be exactly 8 MiB (8,388,608 bytes), and what does that bound buy for both encryption and decryption?
3. Why must the caller supply the raw AES key without a key derivation function, and why are compression, directory recursion, network transport, custom cryptography, and production key management outside this project's scope?
4. Why must production nonces come from `crypto/rand`, and what does reusing a nonce with the same key do to AES-GCM's security?
5. Why is the randomness source injectable, why is "two random outputs differ" not valid proof, and what does a deterministic source let tests observe instead?
6. Why are the four magic bytes `G30F`, the version byte `1`, and the eight declared-length bytes part of the AEAD's associated data, and what does changing those bytes do to authentication?
7. Why must the decryptor validate magic, version, declared length, total size, and truncation before `Open`, and how do fail-fast structural errors remain distinct from authenticated-header tampering in tests?
8. Why does the decryptor return the same generic authentication failure for wrong key, authenticated-header tampering, modified nonce, modified ciphertext, and modified tag, and what would distinguishing the cases leak?
9. Why must the post-authentication length check succeed before all-or-nothing reader/writer decryption releases any plaintext, and what do a mismatch and a no-partial-output test establish?
10. Why must file-level writes use a temporary file in the destination's directory, close it before replacement, and leave the input path untouched, and how do these rules keep the destination unchanged on failure without claiming universal crash durability?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every functional requirement in section 8 is satisfied.
- Every verification case in section 14 has a corresponding test.
- Encrypt then decrypt is exact for all three key sizes and for empty, binary, and 8 MiB maximum-size content.
- Empty files are supported. Encrypt produces a container with declared length `0` and empty ciphertext; decrypt produces an empty plaintext.
- Keys of any length other than 16, 24, or 32 bytes are rejected with an error naming the offending length and the valid lengths.
- Inputs larger than 8 MiB are rejected with an error naming the source and the maximum. Decryption also rejects containers whose declared length exceeds 8 MiB and containers whose total byte size exceeds `8388608 + fixed overhead`.
- Structural format errors (wrong magic, unsupported version, declared length out of `[0, 8388608]`, total size above `8388608 + fixed overhead`, truncated body) are rejected with a clear structural error before `Open`. The tests distinguish structural errors from authenticated-header tampering.
- Authenticated-header tampering (for example, changing the declared-length bytes to another in-range value) returns the same generic authentication failure as wrong key, modified nonce, modified ciphertext, and modified tag. No plaintext is written.
- The post-authentication length check compares the plaintext produced by `Open` against the declared length. A mismatch returns a format/integrity error and writes no plaintext.
- The decryptor does not write any partial plaintext to the destination on any failure. Reader/writer destinations receive no bytes. File destinations are unchanged.
- Reader/writer decryption is all-or-nothing: the first plaintext byte is written only after structural validation, `Open`, and the post-authentication length check have all succeeded.
- File-level writes use a temporary file in the destination's directory. The temporary file is closed before the rename. The destination is replaced only on success. On failure, the temporary file is removed and the destination is unchanged.
- The input path is never overwritten in place.
- Production nonces come from `crypto/rand`. Tests inject a deterministic source and observe the planned bytes through the container's nonce field.
- The package documentation states the AES-GCM semantics, the 8 MiB maximum, the key-length validation, the container contract (`G30F` magic, version `1`, eight-byte big-endian declared length, AEAD-required nonce, ciphertext-and-tag), the header-as-associated-data rule, the structural-error-vs-authenticated-header-tampering distinction, the post-authentication length check, the no-distinguish rule, the all-or-nothing reader/writer rule, the temporary-file-then-replace discipline, the normal-operation atomic visibility rule, and the platform/filesystem limits.
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Streaming seal with chunk framing (design problem).** True large-file support requires a chunked AEAD framing protocol. The design problem is to maintain authentication across chunks while still bounding memory and still providing all-or-nothing release of plaintext. A correct design must define: how each chunk's nonce is derived; how chunks are ordered; how the chunk sequence is authenticated as part of associated data; how out-of-order, missing, or duplicate chunks are detected; and how partial decryption is handled when a chunk fails authentication. The extension is the framing problem, not the cipher choice. The required scope does not solve this problem.
- **Overwrite destination on encrypt.** Allow the encrypt operation to overwrite the destination's existing content. The encrypt operation still uses a temporary file in the destination's directory and replaces the destination only on success. The "no partial output" rule for decrypt stays unchanged. Do not allow overwriting the input path in place; encrypt must always read from a source distinct from the destination. Do not add a `--force` flag or a confirmation prompt.
