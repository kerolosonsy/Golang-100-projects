# Project 039 — Concurrent Image Resizer

## 1. Project Name and Number

- Project 039 — Concurrent Image Resizer.
- The project teaches a bounded worker pool for local image files, deterministic discovery, safe output publication, and standard-library image processing.

## 2. Project Idea

Read regular PNG, JPG, and JPEG files from an explicit input directory and write resized PNG files to a separate explicit output directory. Use a bounded number of workers. Tests create small images inside temporary directories; the project never searches user folders and never uses a network.

Each supported candidate is decoded by content with standard-library image decoders, resized to fit inside a positive target box without upscaling, encoded as PNG, and published only after a successful temporary-file write and close. Independent file failures are reported while other jobs continue, unless context cancellation stops further work.

## 3. Why This Project Now?

- Project 038 was the immediate predecessor and made shared state safe with a mutex.
- This project applies bounded concurrency to filesystem and image resources, where ownership includes file handles, temporary files, and output names.
- It also combines the worker-pool lessons from Project 037 with deterministic preflight and cleanup.

## 4. Prerequisites

- Complete Projects 031 and 038 first.
- Project 031 supplies the barriers, wait groups, and cancellation patterns reused for deterministic worker limits and shutdown.
- Project 038 supplies the mutex-protected state ownership that this project extends to filesystem and worker-pool resources.
- You must understand goroutines, channels, wait groups, mutex-protected counters, context cancellation, temporary directories, file handles, and race-detector testing.
- You should know basic image dimensions, color models, PNG and JPEG encoding differences, and how standard-library decoders report configuration before full decoding.

## 5. What You Must Know Before Starting

- Know that an explicit directory is an input boundary and must not silently become a recursive traversal.
- Know how to inspect directory entries, distinguish regular files from symlinks, and sort names for deterministic order.
- Know that an extension is useful for candidate classification but cannot prove file content.
- Know that configuration decoding can reveal dimensions before allocating the full image and can enforce a pixel budget.
- Know that width multiplied by height must be computed without integer overflow before comparing the decoded pixel count against the caller's maximum, and that overflow or non-positive decoded dimensions must be rejected without allocating a huge image.
- Know that PNG preserves alpha while JPEG does not preserve transparency in the same way.

- Know the temporary-file publication pattern: create the temporary file inside the output directory, write the complete encoded output, close successfully, then rename to the final destination.
- Know that failure requires cleanup and that preflight rejection must not overwrite an existing destination.
- Know that a worker count bounds active jobs, not merely queued names.

## 6. Explanation of New Concepts

### Concepts

- The worker pool separates discovery and preflight from per-file processing.
- Discovery produces a stable list of every direct non-directory entry, including symlinks and unsupported regular files; these are the candidates that receive ordered reports.
- Supported regular image candidates are the subset eligible for decoding.
- Preflight validates both directories, dimensions, worker count, generated output names, and existing destinations before any worker begins image work or creates a temporary file.
- Processing then handles each supported candidate and records a result at its fixed input position.

- The resize policy is pinned.
- Let the source dimensions be positive and let the target box be positive.
- The image is fitted inside the box using the smaller of the width and height scale factors, with a maximum scale of one so no source is enlarged.
- Each scaled dimension is rounded to the nearest integer with half values rounded upward, then constrained to at least one pixel.
- This rule is documented so tests do not accept implementation-specific rounding.
- The result therefore fits within the target box and does not upscale; rounding can create a small ratio difference because dimensions are integral.

- Nearest-neighbor sampling is the required quality policy and must use standard-library image primitives.
- The guide intentionally specifies behavior, not an algorithm recipe: do not copy an implementation procedure or pseudocode into the README.
- Transparent source pixels remain meaningful where the source and destination color model support alpha, and tests should verify relevant transparent pixels rather than only dimensions.

- Content decoding is independent of extension trust.
- A file named with a supported suffix but containing invalid or different data is judged by decoder behavior.
- A file with an unsupported suffix is reported as skipped, even if its bytes happen to resemble an image; the supported extension list remains the candidate-selection policy.
- Symlinks are reported and skipped rather than followed.

## 7. Learning Objective

- By completion, you can build a bounded local file pipeline, preflight filesystem side effects, enforce a decoded-pixel budget, apply a precise no-upscale resize policy, preserve relevant transparency, publish outputs without partial files, and assemble deterministic per-candidate results under cancellation.

## 8. Functional Requirements

1. The operation accepts an existing input directory, an existing separate output directory, positive target width and height, a positive worker count, a positive maximum decoded-pixel count, and a context.
2. Discovery is limited to entries directly in the explicit input directory; recursive traversal is not required.
3. Candidate ordering is deterministic: directory entries are ordered lexically by name before results are assigned positions.
4. Regular files with `.png`, `.jpg`, or `.jpeg` extensions, case-insensitively, are supported candidates.
5. Unsupported extensions are skipped and reported; symlinks are skipped and reported, never followed.
6. Preflight validates both directories, confirms they are distinct directory boundaries, validates worker count, positive dimensions, and the positive pixel limit, derives unique output names, and rejects output-name collisions and existing destinations before work begins.
7. Preflight performs no image processing, output creation, temporary-file creation, or overwrite.
8. Supported candidates are decoded by content using standard-library decoders rather than trusting the extension alone.
9. The decoded pixel count must not exceed the caller's positive maximum decoded-pixel count; width multiplied by height is checked without integer overflow before the comparison, and any overflow or non-positive decoded dimension is an oversized or invalid per-file failure that prevents full decoding and any output publication; configuration decoding rejects oversized images before full decode when configuration information permits.
10. Resizing fits each image inside the target box without upscaling, uses nearest-neighbor behavior, and applies the documented nearest-integer, half-up rounding rule with a minimum dimension of one.
11. Each successful output is encoded as PNG.
12. Each output is written to a unique temporary file in the output directory, the file is successfully closed, and only then is it renamed to the final output name.
13. Temporary files are removed after write, encode, close, or rename failure; no partial final output is left.
14. Existing destinations are never overwritten. Output-name collisions are rejected before workers start.
15. One result is returned for every discovered candidate in deterministic input order, including skipped, rejected, failed, completed, and cancelled statuses as applicable.
16. Independent decode or write failures do not stop other jobs unless context cancellation prevents new or in-flight work from continuing.
17. Cancellation stops new work, allows active jobs to clean up, and reports completed versus cancelled files honestly.
18. The active worker count never exceeds the configured worker count.

## 9. Inputs and Outputs

### Interface Contract

- Inputs are explicit directory paths, positive target dimensions, a positive worker count, a positive maximum decoded-pixel count, and a context.
- The input directory is not inferred from the current environment.
- The output directory is separate and already exists.
- Each direct non-directory entry receives a stable position after lexical ordering.

- Outputs are PNG files named from the supported input basename with its extension replaced by `.png`, plus one status result per discovered candidate.
- A successful result identifies its source, destination, final dimensions, and output size if the chosen result contract includes size.
- A skip or failure identifies its reason.
- No result claims success before final rename.

- Text-only example: three directory entries sort as `a.jpg`, `b.png`, and `notes.txt`.
- The first two receive positions one and two and may finish in either order; the final result list remains in that order, while `notes.txt` reports unsupported and has no output.

## 10. Rules and Edge Cases

- The input and output paths must exist as directories when preflight begins, and they must be separate boundaries.
- A missing path, non-directory path, identical directory, non-positive worker count, non-positive target dimension, or non-positive pixel limit rejects the operation before work.
- The project does not create either directory.

- Only direct regular files with the supported extensions are eligible for image work.
- Direct subdirectories are outside the candidate list and are not traversed.
- A symlink is a reportable candidate but is skipped even if it points to a regular supported image.
- Unsupported regular files are reportable candidates and are skipped.
- Supported files with corrupt content fail independently during decode.
- Decoder content, not extension, determines whether bytes are valid PNG or JPEG data.

- The pixel budget is the caller's positive maximum decoded-pixel count.
- Width multiplied by height must be checked without integer overflow before that comparison; an overflow, a non-positive decoded width, or a non-positive decoded height is an oversized or invalid per-file failure that prevents full decoding and any output publication.
- Configuration decoding rejects an image above the budget or with overflowed dimensions before full decode where possible; if full decoding still discovers a violation or fails, the file result reports the specific failure and no output is published.
- Dimensions after resize are positive, inside the target box, and never larger than the source dimensions.
- The source aspect ratio is fitted using the pinned scale and rounding policy.

- Generated output names must be unique among supported candidates.
- Any collision or existing destination causes whole-batch preflight failure before workers, decoding, or temporary files; supported candidates report rejection or batch-blocked status according to the result contract, while skipped candidates retain their skip reports.
- A destination already present as a regular file, directory, or symlink is never replaced.
- A failure in one candidate does not cancel independent candidates after successful preflight.
- Cancellation may prevent queued candidates from starting and may stop active processing at safe cancellation points, but every active job must clean temporary state.

## 11. Project Constraints

- Use only the Go standard library, including `image`, `image/png`, `image/jpeg`, filesystem APIs, context, synchronization, and testing support.
- Use temporary directories generated by tests and no network or user folders.
- Do not upscale.
- Do not recurse.
- Do not trust extensions for decoding.
- Do not overwrite existing outputs.
- Do not leave temporary files after failure.
- Prove maximum concurrency with barriers and counters, never sleep.
- The race detector must pass.
- Do not provide algorithm pseudocode or implementation snippets in the guide.

## 12. Design Questions Before Coding

- What exactly is a candidate, and in what order are candidates reported?
- Which entries are skipped before processing, and which failures occur per file?
- How are output names derived and collision-checked?
- How are input and output directories proven distinct?
- What pixel count limit is applied and at which decode stage?
- How is width multiplied by height computed without integer overflow before the comparison, and how does that prevent allocating a huge image?
- What precise rounding rule makes dimensions deterministic?
- How is no-upscaling enforced for both dimensions?
- How are alpha pixels treated by the chosen image types and PNG encoder?
- Who owns each worker, job channel, result position, temporary file, and rename?
- How can cancellation stop queued work without abandoning cleanup?

## 13. Implementation Milestones

1. Define candidate discovery, lexical ordering, status values, output naming, and directory policies.
2. Implement read-only preflight for directories, dimensions, worker count, collisions, and existing destinations.
3. Add deterministic reports for unsupported entries and symlinks without following them.
4. Establish the bounded worker pool and fixed result positions.
5. Add content-based configuration and pixel-budget validation for supported image files, including overflow-safe width-times-height computation against the caller's positive maximum.
6. Add the pinned nearest-neighbor, no-upscale resize behavior and PNG encoding.
7. Add temporary output creation, successful close, rename, and cleanup for every failure path.
8. Add independent-error continuation and context cancellation with honest statuses.
9. Add barrier-based maximum-concurrency tests, deterministic result tests, and race-detector verification.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Create small PNG and JPEG images in a temporary input directory and verify successful PNG outputs.
- Test exact target dimensions, aspect-ratio fitting, no upscaling, the half-up rounding boundary, and transparent pixels where the source contains alpha.
- Test unsupported extensions, symlinks, corrupt supported files, and files over the positive configured pixel limit.
- Test the exact boundary where width times height equals the configured maximum, one pixel above the boundary, and an overflow-safe dimension scenario where declared width times height would overflow without the guard; in each overflow case, no full decode is performed and no output file is created.

- Test missing or non-directory paths, identical input and output directories, non-positive worker counts, non-positive dimensions, duplicate generated names such as same-stem supported files, and existing destination files, directories, and symlinks.
- Verify these preflight failures occur before workers or temporary outputs begin.

- Use barriers to hold workers and measure maximum active jobs without sleep.
- Test deterministic lexical result order even when completion is deliberately out of order.
- Test partial success when one file fails to decode or write.
- Test cancellation before queued work starts and during active work, then verify completed versus cancelled statuses and temporary-file cleanup.
- Verify no output is overwritten and no temporary file remains.
- Run the package under the race detector.

## 15. Common Mistakes to Watch For

- Trusting an extension instead of decoder content accepts mislabeled bytes.
- Following symlinks violates the candidate boundary.
- Recursing changes the input contract.
- Decoding before checking configuration can allocate excessively large images, and multiplying width by height without an overflow guard can let an oversized image slip through the pixel budget.
- Scaling by width or height alone can exceed the target box.
- Upscaling small images violates the required policy.
- Floating or library-default rounding makes boundary tests nondeterministic.
- JPEG cannot preserve alpha like PNG, so tests must state what is expected.
- Writing directly to the destination exposes partial output.
- Renaming before close can publish an incomplete file.
- Reusing a predictable temporary name creates collisions.
- Treating an existing destination as replaceable violates no-overwrite behavior.
- Letting one file error stop all independent jobs loses partial success.
- Using sleep to test concurrency creates flaky tests.

## 16. Topics and References for Study

- Study the standard library documentation for `os`, directory entries, file modes, symlink detection, temporary files, rename and cleanup behavior, `path/filepath`, `image`, `image/color`, `image/png`, `image/jpeg`, `context`, `sync.WaitGroup`, channels, and testing.
- Read about nearest-neighbor resampling, aspect-ratio fitting, alpha compositing, resource limits, atomic publication, and bounded worker pools.
- Review the race detector's treatment of synchronization and file-processing tests.

## 17. Self-Assessment Questions

1. Why must input and output directories be explicit and separate?
2. Which entries are candidates, and why are symlinks skipped?
3. Why is extension classification different from content decoding?
4. How does the pixel budget protect the process, and why must width times height be checked without overflow before the comparison?
5. What exact rounding rule governs output dimensions, and why is the scale capped at one?
6. Why must a temporary file live in the output directory?
7. What must happen if encoding succeeds but close fails?
8. Why can one decode failure coexist with other successful outputs?
9. How do barriers prove worker limits without sleep?
10. What does a result mean when cancellation arrives before a candidate starts?

## 18. Definition of Completion

- [ ] The implementation satisfies every functional requirement using standard-library packages only.
- [ ] Preflight validates directories, workers, dimensions, collisions, and existing destinations before work or mutation.
- [ ] Supported regular PNG and JPEG files are decoded by content, bounded by the caller's positive maximum decoded-pixel count using an overflow-safe width-times-height check, resized with deterministic nearest-neighbor no-upscale behavior, and published as complete PNG files through temporary files and rename.
- [ ] Unsupported entries and symlinks are reported without being processed.
- [ ] Results are one per candidate and input-ordered.
- [ ] Partial failures continue independently, cancellation terminates workers cleanly, temporary files are removed, outputs are never overwritten, concurrency is bounded, and the package passes the race detector.

## 19. Optional Extensions

- Add deterministic checksum metadata to successful reports without changing output publication or result ordering.
- Add a separately documented policy for preserving or flattening alpha in PNG outputs, with tests that distinguish transparent pixels from opaque color values.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 038 — Mutex Bank Account](../../03-concurrency/038_mutex_bank_account/README.md#20-prerequisite-based-documentation-guide), [Project 031 — Concurrent Timer](../../03-concurrency/031_concurrent_timer/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`image`](https://pkg.go.dev/image), [`image/draw`](https://pkg.go.dev/image/draw), [`image/png`](https://pkg.go.dev/image/png), [`image/jpeg`](https://pkg.go.dev/image/jpeg).
- **Standards and concept references:** [Go image package article](https://go.dev/blog/image).

### Project-specific learning focus

- **Learn now:** image bounds, nearest-neighbor resampling, aspect-ratio fitting, alpha handling, pixel limits, bounded workers, and atomic file publication.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
