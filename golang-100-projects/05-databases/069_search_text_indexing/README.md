# Project 069 — Search Text Indexing

## 1. Project Name and Number

- Project 069, search_text_indexing.
- This README is a learning guide only.
- You will create every source and test file yourself in `05-databases/069_search_text_indexing/`.
- This guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 2. Project Idea

Build an in-memory immutable document index. Tokenize text using the Project 019 Unicode letter-or-digit rule, store per-document term frequency, query with AND semantics, rank deterministically, and paginate ranked results with a query-bound opaque cursor.

## 3. Why This Project Now?

- Project 069 follows Project 068 in the catalog and shifts from external-resource coordination to a pure in-memory data structure where normalization, immutability, ranking, deterministic ordering, and pagination can be tested exhaustively without services.
- Projects 061 and 041 remain required foundations; Project 019 contributes a focused tokenization rule for review.

## 4. Prerequisites

- Required prerequisites: Projects 068, 061, and 041.
- Optional review: Project 019 for its tokenization and lowercasing rule; it is not a formal prerequisite.
- Reuse that rule exactly.
- All tests use only in-memory data and require no Docker, database, network, or environment variables.

## 5. What You Must Know Before Starting

- Know Unicode rune classification, lowercase conversion, maps and slices, inverted indexes, term frequency, stable sorting, immutable ownership, opaque cursor encoding, deterministic fingerprints, and table-driven or property-style tests.

## 6. Explanation of New Concepts

### Concepts

- A document has a unique positive ID and original text.
- Empty document collections are valid and searchable with zero results.
- Build rejects nonpositive or duplicate IDs.
- It copies every caller-owned input needed by the index so later mutation of the original collection or text holders cannot change indexed terms or returned original text.

- Tokenization follows Project 019 exactly: a token is a maximal run of Unicode letters or digits; separators end tokens; tokens are lowercased.
- Unicode normalization beyond lowercasing is out of scope.
- The inverted index stores each term's frequency per document.

- A query is tokenized under the same rule and reduced to distinct normalized terms.
- An empty or separator-only query is typed invalid input.
- Repeated query terms count once for both matching and ranking.
- AND semantics require every distinct term.
- Score is the sum of per-document frequencies for the distinct query terms present in the document.
- Score and per-document term frequency are signed 64-bit positive integers.
- Build and search use checked addition and report overflow as a typed failure rather than depending on platform-sized int arithmetic.
- Results sort by score descending, then document ID ascending, and include ID, original text, and score.

- A valid last result for a nonempty AND query always has a strictly positive score.
- Cursor last score must therefore be positive: zero and negative are typed invalid cursor before any search traversal.
- Limit must be 1 through 100 with no silent clamping.
- Fetch or select one extra result to determine whether another page exists.
- A next cursor is returned only when more results remain.
- Pagination continues strictly after the ranking tuple: lower score follows higher score; at equal score, larger ID follows smaller ID.

- Cursor text is versioned, URL-safe, opaque, validated, and unauthenticated.
- It carries last score, last document ID, and a deterministic fingerprint of the normalized distinct query-term set.
- The fingerprint is computed by sorting the distinct normalized terms lexicographically, framing each term with an unambiguous byte-length prefix, hashing the resulting byte sequence with SHA-256, and presenting the digest as lowercase hex.
- Equivalent normalized term sets produce the same fingerprint value regardless of original order, case, separators, or repetition.
- Reject bad version, encoding, nonpositive ID, nonpositive score, impossible ranking state, and query mismatch as typed invalid cursor before search traversal.
- Tamper resistance is out of scope.
- The immutable index guarantees a stable pagination sequence; mutation support is outside scope.

## 7. Learning Objective

- Implement a deterministic immutable inverted index with exact token, ownership, matching, ranking, cursor, and pagination contracts, then prove correctness across edge cases and large corpora without external services.

## 8. Functional Requirements

1. A document contains a unique positive ID and original text.
2. Empty input collection is valid and builds a searchable index with zero results.
3. Reject any nonpositive or duplicate document ID.
4. Copy all caller-owned input needed for indexing and returned original text. Later caller mutation cannot affect results.
5. Tokenization uses maximal Unicode letter-or-digit runs and lowercase conversion exactly as Project 019. Unicode normalization is out of scope.
6. Store per-document frequency for every normalized term.
7. Query uses the same tokenization and deduplicates normalized terms.
8. Empty or separator-only query returns typed invalid input.
9. Multi-term query uses AND semantics; every distinct term must occur.
10. Repeated query terms count once for matching and score.
11. Score is the sum of per-document frequencies for distinct query terms. Score and per-document term frequency are signed 64-bit positive integers. Build and search use checked addition and report overflow as typed failure rather than relying on platform-sized int arithmetic.
12. Sort by score descending, then document ID ascending.
13. Results contain document ID, copied original text, and signed 64-bit positive integer score.
14. Limit must be 1 through 100. Invalid limits return typed invalid input; no clamping occurs.
15. Cursor is versioned URL-safe opaque text containing positive last score, positive last ID, and deterministic query-term fingerprint. Zero and negative score are typed invalid cursor.
16. Reject malformed or query-mismatched cursor as typed invalid cursor.
17. Continue strictly after score-descending, ID-ascending tuple and return next cursor only when more results exist.
18. Index is immutable after build; stable pagination depends on that immutability.
19. No external service or third-party dependency is required.
20. The query fingerprint hashes the lexicographically sorted distinct normalized terms, each framed by an unambiguous byte-length prefix, with SHA-256 and presented as lowercase hex. Equivalent normalized term sets produce the same fingerprint value.
21. The cursor serialization is versioned unpadded URL-safe Base64 over a documented data record containing version, positive score, positive ID, and fingerprint. Strict decoding rejects padding, extra or missing fields, unknown version, invalid fingerprint shape, and trailing data. The cursor is unauthenticated.

## 9. Inputs and Outputs

### Interface Contract

- Build input is a collection of ID and text documents.
- Output is an immutable searchable index or typed invalid-document or overflow outcome.
- Search input is query text, limit, and optional cursor.
- Output is ordered results with positive signed 64-bit integer scores and an optional next cursor whose last score is strictly positive, or typed invalid query, invalid limit, or invalid cursor.

- Example behavior: one document contains "Go go search" and another contains "Go search search."
- Query terms "GO search go" become the distinct terms "go" and "search."
- Both match; both score 3; lower document ID ranks first.
- Repetition in the query does not add score.
- The same distinct term set produces the same fingerprint whether entered as "GO search go" or "go search go search."

## 10. Rules and Edge Cases

- Accept empty collections.
- Reject duplicate and nonpositive IDs.
- Empty document text is valid but contributes no terms.
- Reject empty and separator-only queries.
- Unknown terms return an empty non-nil result with no cursor.
- Deduplicate query terms and sort them lexicographically before fingerprinting.
- Score and per-document term frequency are signed 64-bit positive integers; checked addition is the only allowed accumulation and must report overflow as typed failure.
- Cursor last score must be strictly positive for a valid last result of a nonempty AND query; zero and negative score are typed invalid cursor.
- Never depend on map iteration order.
- Never mutate after build.
- Never silently clamp limits.
- Never accept cursor reuse for another normalized distinct term set or with nonpositive score or ID.

## 11. Project Constraints

- In-memory only.
- No SQL, BM25, stemming, stop-word removal, locale-specific collation, Unicode normalization, fuzzy matching, index mutation, cursor authentication, or external services.
- Search correctness must not depend on input order or map order.
- Normal and race tests require no Docker or environment variables.

## 12. Design Questions Before Coding

- How is caller ownership severed during build?
- What exact rune rule forms tokens?
- How is empty query distinguished from no matches?
- How are distinct terms ordered, framed, and hashed to produce the fingerprint?
- How is the ranking tuple compared in both sort and continuation?
- Which cursor values are semantically impossible for a valid last result of a nonempty AND query?
- How is signed 64-bit overflow detected during build and search?
- How can large-corpus expected results be computed independently?

## 13. Implementation Milestones

1. Define document, immutable index, result, typed outcomes, signed 64-bit accumulation, fingerprint, cursor serialization, and exact ownership contracts.
2. Implement Project 019 tokenization and per-document term-frequency construction with checked 64-bit addition.
3. Validate empty collection, positive unique IDs, and deterministic build independent of input order.
4. Implement query tokenization, deduplication, invalid-empty policy, AND matching, signed 64-bit positive integer scoring with checked addition, and deterministic SHA-256 fingerprint over the lexicographically sorted distinct normalized terms with unambiguous byte-length framing, presented as lowercase hex.
5. Add deterministic score-descending and ID-ascending ranking.
6. Add limit validation, versioned unpadded URL-safe Base64 cursor over the documented record, strict cursor decoding that rejects padding, extra or missing fields, unknown version, invalid fingerprint shape, and trailing data, plus strict continuation and one-extra next-page detection. Cursor last score must be positive; zero and negative are typed invalid cursor.
7. Complete exhaustive and large-corpus tests, including race-safe concurrent reads.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Empty collection builds successfully and every valid query returns zero results.
- Duplicate and nonpositive document IDs are rejected.
- Caller mutation after build cannot change indexed terms or returned original text.
- Unicode letters and digits remain token content; punctuation and separators split tokens; case folds by the pinned lowercase rule.
- Unicode normalization is not claimed or silently added.
- One known term matches expected documents and frequencies.
- Multiple distinct terms use AND semantics.
- Unknown term and mixed known/unknown terms return no results.
- Repeated query terms count once for matching and scoring.
- Empty and separator-only queries return typed invalid input.
- Scores equal summed term frequencies and remain within signed 64-bit positive range when constructed with checked addition; overflow is typed failure.
- Higher score sorts first and equal score sorts by lower ID.
- Results contain exact copied original text.
- Limits 1 and 100 work; zero, negative, and above 100 fail without clamping.
- Cursor round trip preserves positive score, positive ID, and the SHA-256 lowercase hex fingerprint.
- Bad cursor version, unpadded URL-safe Base64 encoding, padding character, extra or missing fields, invalid fingerprint shape, trailing data, nonpositive ID, nonpositive score, impossible tuple, and query mismatch return typed invalid cursor.
- Equivalent queries with different case, separators, order, or repetition share the same distinct-term fingerprint when their normalized term set is identical.
- A genuinely different normalized term set rejects cursor reuse.
- Full page traversal has no repeats or omissions and emits no cursor after the final page.
- Build and search results are deterministic across shuffled document input.
- Large-corpus results match an independently calculated expected set, scores, and order.
- Concurrent searches over one immutable index pass the race detector.

## 15. Common Mistakes to Watch For

- Splitting only on ASCII spaces, forgetting lowercase, adding unstated Unicode normalization, counting repeated query terms, using OR semantics, storing only document presence instead of frequency, sorting ties by map order, accepting duplicate IDs, retaining mutable caller slices, using platform int for score or frequency, ignoring signed 64-bit overflow, clamping limits, accepting nonpositive cursor score, accepting padded Base64 cursor text, reusing a cursor across queries, authenticating without specification, or claiming pagination stability on a mutable index.

## 16. Topics and References for Study

- Study Go Unicode letter and digit classification, lowercase conversion, map and slice ownership, stable deterministic sorting, signed 64-bit arithmetic, and padded versus unpadded URL-safe Base64.
- Review inverted indexes, term frequency, AND retrieval, lexicographic tuple continuation, deterministic SHA-256 hashing of canonical term sets with unambiguous byte-length framing, immutable data structures, and Project 019 tokenization rules.

## 17. Self-Assessment Questions

1. Why is an empty collection valid while an empty query is invalid?
2. Why do repeated query terms count once?
3. How does AND matching differ from score calculation?
4. Why must distinct terms be sorted and length-prefix framed before hashing?
5. Why must the cursor last score be strictly positive for a valid nonempty AND result?
6. Why is checked signed 64-bit addition required instead of relying on platform int?
7. What tuple follows a result with equal score?
8. Why does copying input matter?
9. Which consistency guarantee comes from index immutability?

## 18. Definition of Completion

- [ ] Exact Unicode letter-or-digit tokenization and lowercase rule match Project 019.
- [ ] Empty collection, invalid IDs, copied ownership, and immutable search behavior are tested.
- [ ] AND matching, distinct query terms, signed 64-bit frequency sum, and deterministic ranking are correct.
- [ ] Limits 1 through 100 have no silent clamping.
- [ ] Versioned unpadded URL-safe Base64 cursor validates positive score, positive ID, encoding, version, padding rejection, extra or missing field rejection, invalid fingerprint shape, trailing data rejection, impossible states, and query fingerprint.
- [ ] Fingerprint is deterministic SHA-256 lowercase hex over lexicographically sorted distinct normalized terms with unambiguous byte-length framing.
- [ ] Stable full traversal has no duplicate or missing result.
- [ ] Large-corpus and input-order determinism tests pass.
- [ ] Unit and race tests pass with no Docker, database, network, or environment variables.
- [ ] Cursor authentication and Unicode normalization are honestly out of scope.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions

- Add phrase-position indexing as a separately specified ranking mode.
- Add snapshot serialization with version validation while preserving immutable ownership.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 068 — Connection Pooling](../../05-databases/068_connection_pooling/README.md#20-prerequisite-based-documentation-guide), [Project 061 — SQLite CRUD](../../05-databases/061_sqlite_crud/README.md#20-prerequisite-based-documentation-guide), [Project 041 — Context Timeout Example](../../03-concurrency/041_context_timeout_example/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`encoding/base64`](https://pkg.go.dev/encoding/base64).
- **Standards and concept references:** [Introduction to Information Retrieval](https://nlp.stanford.edu/IR-book/).

### Project-specific learning focus

- **Learn now:** token normalization, postings lists, AND retrieval, deterministic ranking, immutable snapshots, canonical hashes, cursor continuation, and overflow-safe scoring.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
