# Project 040 — Race Condition Detector

## 1. Project Name and Number

- Project 040 — Race Condition Detector.
- The project is a learning lab for observing an unsafe shared counter, replacing it with synchronized designs, and explaining what race-detector evidence does and does not prove.

## 2. Project Idea

The learner first creates a deliberately unsafe counter phase strictly outside the repository so the race detector can observe the unsynchronized access. That phase is temporary learning material and lives only in the learner's local workspace.

The completed project compares two safe counters with the same observable increment and load behavior. One protects its value with a mutex. The other uses standard-library atomic operations. Both must produce exact totals under concurrent load and both must pass the race detector. A separate logical-invariant example shows that eliminating a data race is not enough when a multi-step state change must be atomic.

## 3. Why This Project Now?

- Project 039 was the immediate predecessor and used a bounded worker pool with shared coordination.
- This project isolates the failure mode that synchronization must prevent, then connects it to the mutex lesson from Project 038 and the atomic concepts needed by later concurrency work.
- It is a controlled transition from building concurrent systems to measuring their correctness.

## 4. Prerequisites

- Complete Projects 031 and 039 first.
- Project 031 supplies the barriers, wait groups, and cancellation patterns reused for deterministic concurrent counter tests.
- Project 039 supplies the bounded worker-pool ownership that this project isolates into the smallest possible shared-state comparison.
- You must understand goroutines, barriers, wait groups, mutexes, channels, cancellation, and race-free resource ownership.
- You should know basic integer operations, method behavior, and how a test can coordinate many goroutines without sleep.
- Review the standard-library atomic counter API and its alignment and ownership requirements.

## 5. What You Must Know Before Starting

- Know that a data race occurs when concurrent conflicting accesses to the same memory location are not ordered by synchronization and at least one access writes.
- Know that a logical race can occur even when individual memory accesses are synchronized: a check followed by an update may still be interleaved incorrectly.
- Know that happens-before is the ordering relationship established by synchronization operations such as mutex unlock/lock, channel communication, and wait-group completion.

- Know that race-detector instrumentation observes executed memory-access paths and adds runtime and memory overhead.
- It can report a race that the test executes, but it does not prove that unexecuted paths are race-free.
- Know that a clean detector run is evidence for exercised behavior, not a universal correctness proof.
- Know that the temporary unsafe phase must not appear in the completed repository in any form.

## 6. Explanation of New Concepts

### Concepts

- The unsafe counter demonstrates why a read-modify-write increment is not one indivisible memory action.
- Multiple goroutines can read the same old value and overwrite each other's result, producing a logical total below the number of requested increments.
- The same unsynchronized access is also a data race because reads and writes conflict without a happens-before relationship.

- The mutex counter places increment and load behind one mutex.
- Unlock followed by a later lock establishes ordering between protected operations, and the entire increment decision is one critical section.
- The atomic counter uses the standard library's atomic increment and load operations on its counter value.
- Both implementations expose the same observable contract, allowing tests to compare correctness without requiring identical internals.

- A logical-invariant example uses a bounded reservation: a fixed pool has an available count and a reserved count whose sum must remain constant, and one reservation operation must check availability while moving one unit between those related values.
- Even if each field access is individually atomic, another goroutine may observe or change the relationship between them unless the whole reservation operation is synchronized.
- The example demonstrates why synchronization scope follows the invariant, not merely the machine word size.

## 7. Learning Objective

- By completion, you can distinguish data races from logical races, describe happens-before edges, use mutex and atomic synchronization for equivalent counter behavior, interpret race-detector limitations, and design high-contention tests that are deterministic without sleeps.
- You can document an observed unsafe phase without shipping it.

## 8. Functional Requirements

1. The learner creates a deliberately unsafe counter strictly outside the repository so the race detector can observe the unsynchronized access.
2. The completed repository contains no intentionally racing source, test, example, build-tag variant, or default-disabled executable path; only a short prose learner note describing the observed race remains.
3. The completed project contains a mutex-protected counter and an atomic counter.
4. Both counters expose the same observable increment and load behavior.
5. Both counters produce exact totals under concurrent increments.
6. Both counters support zero increments and many increments.
7. Tests use start barriers and wait groups rather than sleeps.
8. Tests include repeated high-contention runs, exact final totals, and independent counter instances.
9. The full package passes the race detector.
10. An additional logical-invariant example demonstrates that synchronization must cover a multi-step invariant operation when useful.
11. The documentation explains data race, logical race, happens-before, instrumentation overhead, executed-path limitations, and the non-proof nature of a clean detector run.
12. Correctness tests do not depend on benchmark timing or universal performance claims.
13. Preflight rejects zero workers when the workload is positive; zero workers with zero workload completes cleanly.

## 9. Inputs and Outputs

### Interface Contract

- Inputs are an increment workload, a number of concurrent workers, and a counter instance.
- Zero workload is valid.
- A worker receives a defined number of increment requests and all workers begin from a start barrier so contention is deliberate rather than accidental.

- Outputs are the loaded final count and test evidence that it equals the exact expected total.
- The mutex and atomic counters must report the same value for the same workload.
- The learning note is a short prose description of the race observed in the temporary unsafe phase; it need not reproduce detector output verbatim.

- Text-only example: eight workers each perform 1,000 increments.
- The expected final count is 8,000 for both safe counters in every repeated run.
- Completion order is irrelevant, and no test assumes which worker increments first.

## 10. Rules and Edge Cases

- Zero workers with zero increments must complete cleanly.
- Zero workers with positive work is invalid preflight input and rejects the call before any goroutine is started.
- Zero increments leave the counter unchanged.
- Many repeated runs must not rely on one lucky schedule.
- Independent counter instances must not share state, locks, or atomic storage.

- The unsafe phase is educational only.
- The learner exercises it strictly outside the completed repository; the final package contains no intentionally racing source, test, example, build-tag variant, or default-disabled executable path, and must not require retaining failing code.
- The final tests run only synchronized implementations under the race detector.

- A data-race report is meaningful only for paths actually executed under instrumentation.
- A clean run does not establish absence of all races.
- Detector overhead can alter scheduling and resource use, so observed behavior under instrumentation is not a universal performance result.
- A logically invalid state transition remains invalid even if no data race is reported.

- The mutex and atomic counter APIs must have matching observable semantics.
- Do not compare them by assuming one is always faster.
- Do not use benchmark outcomes as correctness assertions.
- If an optional benchmark is added, it must state that results depend on hardware, workload, compiler, and runtime conditions.

## 11. Project Constraints

- Use only the Go standard library.
- Use barriers and wait groups, never sleep, to start concurrent work and await completion.
- Do not require copying race-detector output verbatim.
- Do not retain any intentionally racing source, test, example, build-tag variant, or default-disabled executable path.
- Do not enable a racing example through default tests, examples, or normal package execution.
- Do not use benchmarks as correctness tests and do not claim universal performance superiority.
- The completed package must pass the race detector.

## 12. Design Questions Before Coding

- What exact unsafe access pattern will be observed strictly outside the completed repository, and how is it ensured not to remain in any committed file?
- Which operations form the public contract shared by both safe counters?
- What happens-before edge protects each mutex operation?
- Which atomic ordering and value ownership rules apply?
- How will a start barrier create repeatable contention without sleep?
- What does the exact expected total equal for zero, one, and many workers, and how is zero workers with positive work rejected?
- How will independent instances prove state isolation?
- Which logical invariant requires a synchronized check-and-update rather than separate atomic field accesses?
- How will the README explain evidence without overstating detector coverage?

## 13. Implementation Milestones

1. Write the learner note and define the observable counter contract.
2. Create the unsafe counter strictly outside the completed repository, run it under instrumentation, and record a short prose observation of the reported conflict in the learner note.
3. Confirm the completed repository contains no intentionally racing source, test, example, build-tag variant, or default-disabled executable path.
4. Implement the mutex counter with synchronized increment and load behavior.
5. Implement the atomic counter with the same observable behavior.
6. Add barrier-driven tests for zero and many increments, exact totals, repeated contention, and independent instances; preflight rejects zero workers with positive work.
7. Add the logical-invariant example and a test that proves its multi-step rule remains atomic.
8. Run the complete package under the race detector and correct every issue in the final safe paths.
9. Keep any optional performance experiment separate from correctness criteria and avoid universal conclusions.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Test zero increments and a single increment for both safe counters.
- Test many workers with equal and uneven workloads, repeated high-contention runs, exact expected totals, and independent instances.
- Use a start barrier followed by a wait group; do not coordinate with sleep.
- Compare only observable correctness between mutex and atomic implementations.

- Run the full final package under the race detector and verify it reports no race.
- Confirm the package's normal tests, examples, and default execution contain no intentionally racing source, test, example, build-tag variant, or default-disabled executable path.
- Keep a short learner note describing the observed unsafe access; do not require verbatim detector output or any retained unsafe file.

- Test the additional logical-invariant example with concurrent reservations that would over-reserve the pool if availability checking and movement were separated.
- Verify the synchronized operation preserves the constant total, never makes available count negative, and reports success for only the affordable reservation count in every repeated run.
- If an optional benchmark exists, ensure no correctness test reads its timing or treats one environment's result as universal.

## 15. Common Mistakes to Watch For

- Keeping any intentionally racing source, test, example, build-tag variant, or default-disabled executable path in the completed repository makes the package fail its own race-free requirement.
- Assuming one clean detector run proves absence of races ignores unexecuted paths.
- Treating detector output as a universal performance measurement ignores instrumentation overhead.
- Using atomic field accesses separately for a multi-step invariant still permits logical races.
- Starting goroutines without a barrier may produce weak contention and misleading tests.
- Using sleep makes race demonstrations and safety tests flaky.
- Comparing mutex and atomic counters by timing instead of observable behavior confuses performance experimentation with correctness.
- Copying atomic or mutex-containing state after use can invalidate synchronization assumptions.

## 16. Topics and References for Study

- Study the standard library documentation for `sync.Mutex`, `sync.WaitGroup`, channels, context, `sync/atomic`, and the testing package.
- Read Go's race-detector documentation, happens-before explanations, data-race definitions, and guidance on copying synchronization values.
- Review linearizability, critical sections, read-modify-write operations, logical invariants, and instrumentation overhead.
- Compare this project with Project 038's account invariant and Project 039's worker ownership.

## 17. Self-Assessment Questions

1. What is the difference between a data race and a logical race?
2. Which operations establish happens-before in the mutex and barrier tests?
3. Why can an atomic counter be race-free while a multi-step invariant remains logically unsafe?
4. What does detector instrumentation change?
5. Why does detector coverage depend on executed paths?
6. Why must the unsafe phase be removed from the completed project?
7. How are exact totals tested without relying on scheduling order?
8. Why must mutex and atomic implementations share an observable contract?
9. Why are benchmark results not correctness evidence?

## 18. Definition of Completion

- [ ] The learner has observed an unsafe counter race strictly outside the completed repository and documented it briefly in a learner note.
- [ ] The completed repository contains mutex and atomic counters with matching observable increment/load behavior, exact totals under zero and high contention, independent instances, barrier-driven tests, and a synchronized logical-invariant example.
- [ ] Preflight rejects zero workers with positive work; zero workers with zero workload completes cleanly.
- [ ] The full package passes the race detector.
- [ ] No intentionally racing source, test, example, build-tag variant, or default-disabled executable path remains, no failing code is required, no benchmark determines correctness, and documentation accurately states detector limitations and overhead.

## 19. Optional Extensions

- Add an optional benchmark comparing mutex and atomic counters while reporting results only for the measured environment and workload.
- Add a small teaching report that records which synchronization edges protect the counter and invariant, without retaining any unsafe source, test, example, build-tag variant, or default-disabled executable path.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 039 — Concurrent Image Resizer](../../03-concurrency/039_concurrent_image_resizer/README.md#20-prerequisite-based-documentation-guide), [Project 031 — Concurrent Timer](../../03-concurrency/031_concurrent_timer/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- None. This project applies already introduced APIs, standards, and testing practices in a new combination.

### Project-specific learning focus

- **Learn now:** happens-before relationships, read-modify-write races, critical sections, logical invariants, counter linearizability, and measurement overhead.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
