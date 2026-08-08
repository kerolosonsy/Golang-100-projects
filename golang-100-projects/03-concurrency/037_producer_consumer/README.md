# Project 037 — Producer Consumer

## 1. Project Name and Number

- Project 037 — Producer Consumer.
- The project teaches channel ownership, bounded backpressure, cancellation, and deterministic reconciliation across multiple producers and consumers.

## 2. Project Idea

Build a pipeline in which multiple producers generate uniquely identified integer items and multiple consumers process every accepted item exactly once. The required simple transform is fixed: completed result value is the negative of its positive input ID. This makes correctness easy to check while keeping the learning focus on who owns each channel, when channels close, and how goroutines stop.

The work channel is bounded. Producers must feel backpressure when consumers cannot keep up. Results may arrive in scheduling order, but the final result is assembled and sorted by original input order. Cancellation stops new production, allows in-flight goroutines to terminate, and reports completed and cancelled items honestly.

## 3. Why This Project Now?

- Project 036 introduced deterministic time and mutex-protected shared decisions.
- This immediate predecessor is useful here because both projects require explicit state ownership and cancellation-safe behavior.
- Producer Consumer moves the same discipline into channel-based coordination and prepares you for later worker pools, event buses, and fan-out/fan-in pipelines.

## 4. Prerequisites

- Complete Projects 031 and 036 first.
- Project 031 supplies the barriers, wait groups, and cancellation patterns reused for deterministic tests of out-of-order completion and shutdown.
- Project 036 supplies the mutex-protected state and deterministic-time discipline that this project extends to channel coordination.
- You must understand goroutines, channel send and receive, buffered channels, channel closure, `select`, context cancellation, and basic race-detector reasoning.
- Review the worker-pool ownership ideas from Project 034, but do not rely on sleep-based tests or scheduler luck.

## 5. What You Must Know Before Starting

- Know that a channel close is a broadcast of no-more-sends, not a cancellation mechanism.
- Know that the sender side owns closure when ownership is unambiguous.
- Know how a coordinator can wait for all producers before closing shared input, and how a second wait group can wait for all consumers before closing results.
- Know that a bounded buffer limits queued work but does not by itself prevent goroutine leaks.
- Know how context cancellation affects both blocked sends and blocked receives.
- Know that scheduling order is nondeterministic and must not define result order.

## 6. Explanation of New Concepts

### Concepts

- Pipeline ownership is the central concept.
- Producers send work but never close the shared input because several producers may still be sending.
- A coordinator waits until every producer has stopped, then closes the work channel exactly once.
- Consumers receive until the work channel is closed or cancellation prevents further work.
- Consumers never close shared input.

- A separate result channel carries completed item reports.
- Consumers send reports and exit; a coordinator waits for every consumer to exit, then closes the result channel.
- This ordering prevents a result send from racing with result closure and prevents send-after-close panics.

- Backpressure comes from a bounded work channel.
- When its buffer is full, a producer must wait for a consumer or for cancellation.
- A zero-sized buffer is valid and creates direct handoff; a small positive buffer makes queueing visible.
- Neither setting permits dropping accepted work silently.

- The coordinator distinguishes exactly three per-item statuses: completed, cancelled, and not-accepted.
- An item is accepted only when its work has been placed into the work channel.
- An item is completed only when its transformed result has been committed before cancellation.
- An item is cancelled when it was accepted into the work channel but did not complete before cancellation.
- An item is not-accepted when cancellation prevented it from ever entering the work channel.
- There is no rejected status during a valid run; preflight rejection rejects the entire call before goroutines start and is not part of the per-item processing report.

## 7. Learning Objective

- By completion, you can design a bounded multi-producer and multi-consumer pipeline with one channel owner per closure, no send-after-close path, honest cancellation reporting, and deterministic input-order output.
- You can prove exactly-once processing for accepted items and show that all goroutines finish.

## 8. Functional Requirements

1. The coordinator accepts an input sequence of integer item IDs, positive producer and consumer counts, a positive or zero work-buffer size, and a context.
2. Every input item ID must be positive and unique; duplicate or non-positive IDs are rejected during preflight before any goroutine is started.
3. Empty work returns immediately without launching unnecessary pipeline goroutines.
4. Multiple producers divide the input work without generating duplicate accepted IDs.
5. Producers stop starting new work after context cancellation and report which items are not-accepted.
6. The work channel is bounded by the requested buffer size, including the valid zero-buffer case.
7. Only the coordinator closes the work channel, after all producers have finished sending.
8. Consumers process each accepted item exactly once by returning the negative of its positive input ID.
9. Consumers never close the shared work channel.
10. The result channel closes only after all consumers have exited.
11. No send occurs after either channel is closed.
12. Independent processing continues until cancellation prevents further safe progress.
13. Final results are sorted by original input order, never by completion or worker order.
14. Cancellation terminates blocked producers, blocked consumers, and the coordinator's collection path without goroutine leaks.
15. Every original input has exactly one final, input-ordered status: completed, cancelled, or not-accepted. No input has more than one status.

## 9. Inputs and Outputs

### Interface Contract

- Inputs are an ordered list of positive integer IDs, positive producer and consumer counts, a non-negative work-buffer capacity, and a context.
- The transform is deterministic negation: a completed item with ID 17 reports value -17.
- The source ID remains in its report so exactly-once processing can be checked.

- Outputs are a deterministic result collection in original input order, plus any preflight validation error from a rejected call.
- Each report identifies the input position and item ID, states exactly one of completed, cancelled, or not-accepted, and includes the transformed value only when processing completed.
- The output must not imply that an item was completed merely because it was generated or queued.

- Text-only example: four valid IDs enter a pipeline with two producers and three consumers.
- Consumer completion messages arrive in order four, one, three, two, but the returned collection is ordered one, two, three, four by the original input positions.

## 10. Rules and Edge Cases

- Producer and consumer counts must always be positive, including when input is empty.
- Preflight validates producer and consumer counts, the buffer capacity, IDs, and duplicates first; after successful preflight, empty input returns immediately before worker creation.
- A negative buffer size is invalid; zero is valid.
- Empty input produces an empty result and no deadlock.
- Duplicate or non-positive IDs fail preflight before any goroutine is started.

- The coordinator owns work-channel closure after producer completion.
- A consumer cannot close a channel it does not own.
- Result closure waits for all consumers, including consumers that exit because cancellation occurred.
- Every potentially blocking send or receive must have a cancellation path.

- More workers than items is valid; idle workers still terminate.
- With a zero buffer, producer and consumer synchronization must remain correct.
- Cancellation may occur before production, while producers are blocked, or while consumers are processing.
- On a non-cancelled valid run, every accepted item is completed exactly once.
- On cancellation, items whose transformed result was committed before cancellation are completed; items that were accepted into the work channel but did not complete are cancelled; items that never entered the work channel are not-accepted.
- No item has more than one status.

## 11. Project Constraints

- Use only the Go standard library.
- Use channels, goroutines, wait groups, and context; do not add third-party pipeline libraries.
- Use a bounded work channel.
- Tests must use barriers, channel handshakes, or other deterministic coordination and must never sleep.
- Do not close channels from consumers or producers that do not own them.
- Do not use shared mutable maps or slices without synchronization.
- Only the three per-item statuses completed, cancelled, and not-accepted appear in a valid run; preflight rejection is not a per-item status.
- The race detector must pass.

## 12. Design Questions Before Coding

- Which component owns each channel and the exact event that permits closure?
- How will producers divide input positions without duplicate production?
- At what moment is an item accepted?
- Which result statuses distinguish not-accepted from accepted but cancelled?
- How does a blocked producer observe cancellation while the work buffer is full?
- How does a consumer stop without leaving a result send blocked?
- How will result collection wait for closure without closing too early?
- How will input order be retained when processing completes out of order?
- How can tests force out-of-order completion without sleep?

## 13. Implementation Milestones

1. Define the item, result, status, and input-order model in prose, including the exclusive completed/cancelled/not-accepted status set, before designing channel flow.
2. Implement preflight for counts, buffer size, empty input, duplicate IDs, and non-positive IDs.
3. Establish coordinator ownership of work and result channel closure.
4. Add producers that stop new production on cancellation and never close shared input.
5. Add consumers that transform each accepted item once and report exactly one status per input.
6. Add wait-group sequencing so work closes after producers and results close after consumers.
7. Add deterministic barriers for backpressure, zero-buffer handoff, and forced out-of-order processing.
8. Assemble final results in input order, guarantee one status per input, and verify status consistency.
9. Add cancellation and goroutine-completion tests, then run the package under the race detector.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Test one producer and one consumer, then many producers and many consumers.
- Test more workers than items.
- Test empty input, zero buffer, a small buffer, invalid counts, negative buffer size, duplicate IDs, and non-positive IDs.
- Verify preflight rejects invalid input without starting production.

- Test that every accepted item is transformed exactly once, that no item is lost or duplicated, and that result order follows original input order despite intentionally out-of-order processing.
- Use barriers and channels to hold selected consumers and release them in a chosen sequence; do not sleep.

- Test cancellation before production, while producers face a full buffer, and while consumers are active.
- Verify completed, cancelled, and not-accepted items are reported exactly once per input, with no input receiving more than one status.
- Verify work closes only after all producers exit, results close only after all consumers exit, no send-after-close occurs, and all goroutines complete.
- Run the full package with the race detector.

## 15. Common Mistakes to Watch For

- Letting any producer close the shared work channel can race with another producer's send.
- Closing results before consumers finish causes send-after-close panics.
- Treating channel closure as cancellation leaves blocked senders.
- Sending to a full buffer without a cancellation alternative leaks producers.
- Collecting results in arrival order makes output nondeterministic.
- Marking every generated item completed ignores cancellation during processing.
- Starting workers for empty input can create unnecessary shutdown paths.
- Using sleep to make one consumer "win" produces flaky tests.
- Protecting a shared result map only during writes but reading it concurrently still races.

## 16. Topics and References for Study

- Study the standard library documentation for channels, `select`, `sync.WaitGroup`, `context`, and `sync/atomic` when measuring counters.
- Review Project 034's worker-pool design and Project 035's cancellation and ordered results.
- Read about channel ownership, close semantics, backpressure, fan-out, fan-in, structured concurrency, and goroutine leak detection.
- Study the testing package's synchronization patterns.

## 17. Self-Assessment Questions

1. Who closes the work channel, and what proves no producer can send afterward?
2. Who closes the result channel, and what proves no consumer can send afterward?
3. What does the buffer size change, and what does it not change?
4. What is an accepted item?
5. How are not-accepted and accepted-but-cancelled items distinguished?
6. Why must result order be rebuilt from input positions?
7. How do zero-buffer tests differ from small-buffer tests?
8. Which blocked operations observe context cancellation?
9. How do barriers prove behavior without relying on scheduler timing?

## 18. Definition of Completion

- [ ] All functional requirements are implemented and covered by deterministic tests.
- [ ] Empty input returns immediately.
- [ ] Preflight rejects invalid input before any goroutine is started.
- [ ] Valid pipelines handle one or many producers and consumers, all buffer sizes including zero, and more workers than items.
- [ ] Every original input has exactly one final status, results are returned in input order, cancellation reports completed, cancelled, and not-accepted items exactly once, channels close in the required ownership order, no goroutine remains blocked, and the package passes the race detector.

## 19. Optional Extensions

- Add a bounded completed-result buffer with an explicit policy for cancellation when result collection is slow.
- Add per-item processing metadata such as producer and consumer labels while keeping the fixed one-status-per-input rule and exactly-once semantics unchanged.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 036 — Rate Limiter Token Bucket](../../03-concurrency/036_rate_limiter_token_bucket/README.md#20-prerequisite-based-documentation-guide), [Project 031 — Concurrent Timer](../../03-concurrency/031_concurrent_timer/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`sync/atomic`](https://pkg.go.dev/sync/atomic).

### Project-specific learning focus

- **Learn now:** producer and consumer ownership, backpressure, exactly-once accounting, fan-out and fan-in, structured cancellation, and goroutine-leak tests.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
