# Project 044 — Fan-Out Fan-In

## 1. Project Name and Number

- Project 044 — Fan-Out Fan-In.
- The project is a learning lab for a small signed-64-bit pipeline that distributes work across a positive number of workers, lets each worker square its assigned value with overflow detection, and merges the per-worker results into a single ordered result set that matches the original input order even though workers complete out of order.

## 2. Project Idea

The pipeline accepts a slice of signed 64-bit integers paired with stable positions. The pipeline validates a positive worker count before any other step. The feed path sends each item to one of a positive number of workers through a coordinator-owned work channel and then closes the work channel after feeding stops. Each worker squares the integer it accepts, detects overflow on the signed 64-bit type, and produces a result that names the item, reports success or overflow-error, and carries the squared value when success or an overflow indicator when failure. A dedicated lifecycle closer waits for every worker to exit and then closes the result channel. A collector drains the result channel until closure and assembles the final ordered slice. The pipeline reports completed versus cancelled items honestly when the parent context is canceled.

## 3. Why This Project Now?

- Project 043 was the immediate predecessor and built a thread-safe cache with controlled concurrency.
- Project 044 raises the level: instead of one shared structure, many workers process many items in parallel while a coordinator and a lifecycle closer guard the channels.
- Project 041 supplied the context propagation policy this project depends on.
- The goroutine and barrier thinking from Project 031 is reused for deterministic concurrent tests.
- Project 045 follows next and uses the same exact-totals-under-contention testing approach for atomic counters.

## 4. Prerequisites

- Complete Projects 031 and 043 first.
- Project 031 supplies goroutines, barriers, and wait groups reused for deterministic concurrent pipeline tests.
- Project 043 supplies the lock-around-each-operation reasoning that this project adapts into per-worker output and the lifecycle closer that waits for workers to exit.
- You must already be comfortable with goroutines, channels, wait groups, context, error sentinels, and barrier-driven concurrent tests with no sleep.
- Earlier projects that introduce worker-pool ownership are useful review but are not required prerequisites.

## 5. What You Must Know Before Starting

- Know that order preservation in a fan-out fan-in pipeline comes from carrying each item's input position into its result and indexing the merged results by that position.
- Workers may complete out of order.
- The final slice is built by placing each result at its original index in a separate output buffer.

- Know that closure ownership is a hard rule that has three named owners in this pipeline.
- The feed path closes the work channel once feeding stops.
- Workers never close any shared channel.
- A dedicated lifecycle closer owned by the pipeline waits for every worker to exit and then closes the result channel exactly once.
- The collector only receives and drains results until the result channel is closed; it never closes the result channel.
- A worker or collector that closes a shared result channel can make a later worker send panic or can terminate collection before all results are committed.

- Know that the collector must not stop reading merely because the parent context is canceled.
- If the collector stops early, workers can block forever while sending final or cancellation results.
- The collector drains until the lifecycle closer closes the result channel and only then assembles the return.

- Know that the pipeline pins all inputs and squared results to signed 64-bit integers.
- Overflow detection is therefore architecture-independent.
- The pipeline refuses to perform an overflowing absolute-value operation on the signed 64-bit minimum, because the absolute value of that minimum does not fit in signed 64-bit.
- The boundary tests cover zero, positive, negative, the largest magnitude whose square fits, one beyond each positive boundary, one beyond each negative boundary, and the signed 64-bit minimum without overflowing the absolute-value operation.

- Know that the pipeline accepts an item only after it has been sent to the work channel.
- Every original input has exactly one input-ordered final status.
- A preflight-invalid call, including a non-positive worker count, rejects the whole call without per-item statuses.
- No item has multiple statuses.

- Know that context cancellation stops feeding new work, allows in-progress workers to terminate at their next observation point, and reports completed versus cancelled items honestly.
- Work that the worker is already computing cannot be magically preempted in this design; the project documents this honestly.

- Know that an empty input with a valid positive worker count is valid and produces an empty result without starting workers.
- Any zero or negative worker count is rejected before any goroutine starts, even when the input is empty.

## 6. Explanation of New Concepts

### Concepts

- An ordered item is the input paired with its input index.
- The input slice produces one ordered item per integer.
- The pipeline never loses that index and uses it to position results in the final slice.

- The feed path closes the work channel once feeding stops.
- The feed path is the part of the pipeline that owns the work channel, sends each item to a worker, and closes the work channel after every item has been sent.
- Workers read from the work channel until it is closed and then exit.
- Closing the work channel after feeding is the signal that there is no more work to do.

- The lifecycle closer is a dedicated goroutine that waits for every worker to exit and then closes the result channel exactly once.
- The lifecycle closer is the only code that closes the result channel.
- The collector never closes the result channel; it only drains until closure and assembles the return.

- The collector drains the result channel until the lifecycle closer closes it.
- Cancellation of the parent context does not stop the collector; the collector must keep draining so that workers can finish sending final or cancellation results.
- After the lifecycle closer closes the result channel, the collector assembles the final slice and returns.

- The processing boundary is an injectable interface that lets tests force out-of-order completion, drive cancellation, and observe that the pipeline places each result in the correct original index regardless of the order in which workers finished.
- Tests use in-memory channels to drive completion order and observe cancellation without sleep.

- The status model is the project's pinned reporting contract.
- An item is accepted only after it is sent to the work channel.
- Each accepted item produces exactly one input-ordered final status: success or overflow-error if accepted work commits a result; cancelled if accepted but cooperative processing ends without a committed square result due to context cancellation.
- An item that is not accepted because cancellation prevented it from being sent is reported as not-accepted.
- Preflight-invalid calls, including non-positive worker count, reject the whole call without per-item statuses.
- No item has multiple statuses.

- Cancellation propagation is the policy that says the parent context stops feeding, the feed path stops sending new items, workers observe cancellation at their next blocking step and exit, and the lifecycle closer waits for every worker to exit and closes the result channel.
- Cancellation is not a preemption primitive; it is a stop signal observed at convenient points.

## 7. Learning Objective

- By completion, you can distribute work across a positive number of workers, merge per-worker results into one ordered slice, and verify out-of-order completion deterministically with no sleep.
- You can decide who closes which channel and document that ownership clearly.
- You can validate the worker count before the empty-input fast path.
- You can propagate cancellation from a parent context through the feed path, the workers, and the lifecycle closer without hiding it behind a fresh background context.
- You can write tests that force out-of-order completion through an injected processing boundary and pass the race detector.

## 8. Functional Requirements

1. The pipeline accepts a slice of signed 64-bit integers and a positive worker count. Every input and every squared result is a signed 64-bit integer.
2. The pipeline validates the worker count before the empty-input fast path. Zero or negative worker counts reject the whole call without per-item statuses, even when the input is empty.
3. Empty input with a valid positive worker count returns an empty result without starting any worker.
4. The pipeline returns one final status per original input on every valid call, indexed by the original input position. Accepted inputs become success, overflow-error, or cancelled; inputs never accepted because cancellation stopped feeding become not-accepted.
5. Each result reports success with a squared signed 64-bit value, or an overflow-error without an overflowed value. A panic from overflow is not a permitted behavior.
6. The pipeline refuses to perform an overflowing absolute-value operation. The signed 64-bit minimum is handled without overflow.
7. The final returned results are in original input order even though worker completion order is not.
8. The feed path owns the work channel and closes it once feeding stops.
9. Workers never close any shared channel.
10. A dedicated lifecycle closer owned by the pipeline waits for every worker to exit and then closes the result channel exactly once.
11. The collector only receives and drains results until the result channel is closed; the collector never closes the result channel.
12. The collector must not stop reading merely because the parent context is canceled; it drains until the lifecycle closer closes the result channel and only then assembles the return.
13. Context cancellation stops feeding new items, lets workers terminate at their next observation point, and does not pretend to preempt workers already computing.
14. On a non-cancelled valid run, every input is accepted and produces exactly one success or overflow-error result.
15. On cancellation, the pipeline drains all worker results, reconciles accepted versus unaccepted positions, reports every original position once, and returns only after all owned goroutines exit.
16. Every original input has exactly one input-ordered final status: success or overflow-error if accepted work commits a result; cancelled if accepted but cooperative processing ends without a committed square result due to context cancellation; not-accepted if cancellation prevents it from being sent.
17. The pipeline does not create replacement background contexts.
18. The pipeline releases owned goroutines so none of them outlive the pipeline's return when cancellation triggered the exit.
19. Tests force out-of-order worker completion through an injected processing boundary with no sleep; cover one and many workers, empty input, invalid worker count, exact-once delivery, exact order, overflow detection including zero, positive, negative, the largest magnitude whose square fits, one beyond each positive boundary, one beyond each negative boundary, and the signed 64-bit minimum without an overflowing absolute-value operation; cancellation during feeding, cancellation during work, full closure on completion and on cancellation, and race-detector passes.

## 9. Inputs and Outputs

### Interface Contract

- The input is a slice of signed 64-bit integers, a positive worker count, a parent context, and a processing boundary the test can inject to control completion order.
- The output is a slice of results in original input order.
- Each result carries the original index and a single status: success with the squared value, overflow-error without an overflowed value, cancelled, or not-accepted.

- A separate observable is the set of indices the pipeline reports as completed, cancelled, or not-accepted when cancellation arrives.
- The expectation is exactly the union of the three sets over the original indices, and each original index appears exactly once across the union.

- Text-only example: with input two, three, and five distributed across two workers where the second worker finishes its item first, the final ordered slice places a result for index zero first, a result for index one second, and a result for index two third, regardless of the order in which the workers wrote to the result channel.

## 10. Rules and Edge Cases

- An empty input with a valid positive worker count returns an empty result without starting workers and without holding any goroutine.
- Zero or negative worker count returns a clear rejection without starting any goroutine; the pipeline validates the worker count before the empty-input fast path, so a non-positive worker count rejects even when the input is empty.
- A non-positive worker count must not be normalized silently into one worker.

- Squaring overflow on signed 64-bit is detected at the worker before the operation is attempted.
- The pipeline accepts signed 64-bit inputs only, and the overflow check is architecture-independent.
- The pipeline refuses to take the absolute value of the signed 64-bit minimum because that operation would overflow.
- The boundary tests cover zero, positive, negative, the largest magnitude whose square fits, one beyond each positive boundary, one beyond each negative boundary, and the signed 64-bit minimum without performing an overflowing absolute-value operation.

- Cancellation during feeding stops the feed path before it has sent every item.
- Cancellation during work lets each worker observe cancellation at its next blocking step and exit.
- A worker that has already begun its squaring operation completes that operation; the pipeline does not pretend to preempt in-progress work.

- The pipeline reports every original input exactly once, in input order, even on cancellation.
- It never invents a square result for an item that was not accepted; it reports that position as not-accepted.
- It does not lose an accepted item that never committed a square result because of cancellation; it reports that position as cancelled.

- The feed path closes the work channel exactly once.
- The lifecycle closer closes the result channel exactly once after every worker has exited.
- The collector never closes the result channel.
- A worker that closes any shared channel is a bug.
- Tests count closure events to verify ownership.

- The collector must not stop reading merely because the parent context is canceled.
- The collector drains until the lifecycle closer closes the result channel and only then assembles the return.
- Stopping early can deadlock workers sending final or cancellation results.

## 11. Project Constraints

- Use only the Go standard library.
- Use goroutines, channels, wait groups, mutexes when needed for collector bookkeeping, and the context package.
- Use a processing boundary interface so tests can force out-of-order completion without sleep.
- Pin all inputs and squared results to signed 64-bit integers.
- Do not create replacement background contexts.
- Do not let workers close shared channels.
- Do not let the collector close the result channel.
- Do not pretend to preempt already-computing work.
- Do not introduce sleeps into tests.
- The completed code must pass the race detector.

## 12. Design Questions Before Coding

- How does the pipeline carry each item's input index into its result and into the final ordered slice?
- How does the feed path close the work channel once feeding stops, and how does the lifecycle closer close the result channel once every worker has exited?
- How does the collector know when to stop reading without observing the parent context?
- How does the pipeline validate the worker count before the empty-input fast path so a non-positive worker count rejects even when the input is empty?
- How does the pipeline guarantee that an item is accepted only after it is sent to the work channel?
- How does the pipeline refuse to perform an overflowing absolute-value operation on the signed 64-bit minimum?
- How does the pipeline reconcile accepted versus unaccepted positions after cancellation and report every original position once?
- How does the pipeline prove no owned goroutine outlives its return when cancellation triggered the exit?

## 13. Implementation Milestones

1. Define the ordered item type, the ordered result type, the status model, and the processing boundary interface.
2. Validate the worker count first, rejecting zero or negative counts before any goroutine starts and before the empty-input fast path.
3. Implement the feed path that starts the workers, sends each input item to the work channel, and closes the work channel once feeding stops.
4. Implement the workers that accept ordered items, perform the squaring operation with documented overflow detection on signed 64-bit, and write ordered results to the result channel.
5. Implement the lifecycle closer that waits for every worker to exit and closes the result channel exactly once.
6. Implement the collector that drains the result channel until the lifecycle closer closes it and assembles the final ordered slice indexed by each result's original index. The collector must not stop reading on parent cancellation.
7. Implement context propagation so the feed path stops feeding, workers observe cancellation at their next blocking step, and every owned goroutine exits before the pipeline returns.
8. Document and enforce closure ownership: the feed path closes the work channel; the lifecycle closer closes the result channel; the collector never closes the result channel; workers never close any shared channel.
9. Add tests covering zero, positive, negative, the largest magnitude whose square fits, one beyond each positive boundary, one beyond each negative boundary, and the signed 64-bit minimum without an overflowing absolute-value operation, plus forced out-of-order completion, cancellation during feeding, cancellation during work, full closure on completion and on cancellation, and race-detector passes.
10. Run the full package under the race detector and correct every issue reported.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Test forced out-of-order completion: an injected processing boundary drives a known completion order that is not the input order, and the pipeline places every result in the correct input index.
- Test one worker and many workers across the same input; both produce the same final ordered slice.

- Test empty input: an empty input with a valid positive worker count returns an empty result without starting any worker and without leaking goroutines.
- Test invalid worker count: zero and negative worker counts reject the whole call before any goroutine starts, even when the input is empty, and the result carries no per-item statuses.

- Test exact-once reporting: every original input produces exactly one final status in the returned slice, and the slice has the same size as the input on every valid call.
- On a non-cancelled run, every input is accepted and becomes exactly one success or overflow-error result.
- Test exact order: the final result slice is in input order regardless of completion order.
- Test overflow detection across the signed 64-bit boundary: zero is squared and committed;
- Positive values whose squares fit are committed;
- The first positive value whose square does not fit is reported as overflow-error;
- Negative values whose squares fit are committed;
- The first negative magnitude whose square does not fit is reported as overflow-error;
- The signed 64-bit minimum is reported as overflow-error without performing an overflowing absolute-value operation.
- The pipeline does not panic on any of these inputs.

- Test cancellation during feeding: cancel the parent context before the feed path finishes sending every item;
- The pipeline stops feeding, drains whatever worker results already arrived, reports every original position once as one of accepted success, accepted overflow-error, accepted cancelled, or not-accepted, and returns only after every owned goroutine has exited.
- Test cancellation during work: cancel the parent context while workers are processing;
- Workers exit at their next observation point, the lifecycle closer waits for every worker to exit and closes the result channel, the collector drains until closure and only then assembles the return, and the pipeline reports completed, cancelled, and not-accepted indices honestly.

- Test full closure on completion and on cancellation: closure events are counted, the feed path closes the work channel exactly once, the lifecycle closer closes the result channel exactly once, the collector never closes the result channel, and no worker closes any shared channel.
- Test that no owned goroutine outlives the pipeline's return when cancellation triggered the exit.
- Run every test under the race detector.

- Do not introduce sleep into the tests.
- Use barriers to align goroutine starts and wait groups to await completion.

## 15. Common Mistakes to Watch For

- Letting a worker close a shared channel leads to a closed-channel read on the coordinator or another worker.
- Letting the collector close the result channel leads to a closed-channel send on a worker that is still trying to report.
- Closing the work channel before every item has been sent loses work and makes the pipeline report incorrect cancelled counts.
- Closing the result channel before every worker has exited loses results and lets the collector exit early.
- Letting the collector stop reading on parent cancellation can deadlock workers sending final or cancellation results.
- Trying to preempt already-computing work is not supported by this design and any attempt hides real behavior.
- Creating a replacement background context disables parent cancellation and is a bug.
- Using sleep to force out-of-order completion introduces flaky behavior that masks real ordering bugs.
- Reporting cancelled items as completed, or reporting an item twice across statuses, breaks the honest reporting contract.
- Returning a single error instead of per-item statuses hides which inputs failed and which were never processed.
- Performing an overflowing absolute-value operation on the signed 64-bit minimum is a hidden bug.
- Validating the empty-input fast path before the worker count lets a non-positive worker count escape and start workers.

## 16. Topics and References for Study

- Study the standard library documentation for the context package, in particular the cancellation propagation contract.
- Study the standard library documentation for channels and the rule that the sender closes and the receiver does not.
- Read Go's standard explanations of the fan-out fan-in pattern and the importance of carrying the input index with each result.
- Compare this project with the lock-around-each-operation reasoning of Project 043, the context propagation policy of Project 041, and the snapshot-outside-lock reasoning of Project 042.
- Project 045 follows next and uses the same exact-totals-under-contention testing approach.

## 17. Self-Assessment Questions

1. Why is the work channel closed by the feed path and the result channel closed by the lifecycle closer rather than by a worker or the collector?
2. Why must the collector drain until the lifecycle closer closes the result channel and not stop reading on parent cancellation?
3. Why is the worker count validated before the empty-input fast path, and why does a non-positive worker count reject even when the input is empty?
4. How does the pipeline guarantee that an item is accepted only after it is sent to the work channel and that every original input has exactly one input-ordered final status?
5. How does the pipeline handle the signed 64-bit minimum without performing an overflowing absolute-value operation?
6. How does cancellation during feeding differ from cancellation during work in what the pipeline reports?
7. How does the pipeline prove no owned goroutine outlives its return when cancellation triggered the exit?
8. Why must the pipeline not create a replacement background context?
9. Why is forcing out-of-order completion through an injected boundary the natural test approach, and why is sleep the wrong approach?

## 18. Definition of Completion

- [ ] The pipeline accepts a slice of signed 64-bit integers and a positive worker count, validates the worker count before the empty-input fast path, and returns one final status per original input in input order on every valid call.
- [ ] Each status is success with the squared value, overflow-error without an overflowed value, cancelled, or not-accepted.
- [ ] The feed path closes the work channel once feeding stops.
- [ ] Workers never close any shared channel.
- [ ] A dedicated lifecycle closer waits for every worker to exit and closes the result channel exactly once.
- [ ] The collector only receives and drains results until the lifecycle closer closes it and never closes it itself.
- [ ] The collector must not stop reading on parent cancellation.
- [ ] Empty input with a valid positive worker count returns an empty result without starting any goroutine.
- [ ] Zero or negative worker counts reject the whole call before any goroutine starts, even when the input is empty.
- [ ] On a non-cancelled valid run, every input is accepted and produces exactly one success or overflow-error result.
- [ ] On cancellation, the pipeline drains all worker results, reconciles accepted versus unaccepted positions, reports every original position once, and returns only after all owned goroutines exit.
- [ ] The pipeline refuses to perform an overflowing absolute-value operation and handles the signed 64-bit minimum without overflow.
- [ ] Tests force out-of-order completion through an injected processing boundary and cover empty input, invalid worker count, exact-once reporting, exact order, overflow detection across the signed 64-bit boundary, cancellation during feeding, cancellation during work, full closure on completion and on cancellation, and the race detector.
- [ ] Tests do not use sleep.

## 19. Optional Extensions

- Add a small teaching note that diagrams, in plain prose, the lifecycle of every channel and goroutine so the closure and cancellation rules are obvious to a new reader.
- Add repeated cancellation scenarios that confirm success, overflow-error, cancelled, and not-accepted positions together cover every input exactly once.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 043 — Thread-Safe Cache](../../03-concurrency/043_thread_safe_cache/README.md#20-prerequisite-based-documentation-guide), [Project 031 — Concurrent Timer](../../03-concurrency/031_concurrent_timer/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- None. This project applies already introduced APIs, standards, and testing practices in a new combination.

### Project-specific learning focus

- **Learn now:** fan-out, fan-in, stable input indexing, result completeness, overflow and cancellation categories, channel closure, and goroutine lifecycle diagrams.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
