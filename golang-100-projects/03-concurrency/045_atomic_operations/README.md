# Project 045 — Atomic Operations

## 1. Project Name and Number

- Project 045 — Atomic Operations.
- The project is a learning lab for a single concurrent metrics component built from typed atomic primitives, the honest description of what atomicity covers and what it does not, and the careful verification of counters and a flag under contention without inventing a transaction across fields.

## 2. Project Idea

The component carries three values: a non-negative request counter, a non-negative error counter, and an enabled flag. Each counter is a typed atomic unsigned 64-bit value. The flag is a typed atomic boolean. The request counter and error counter support atomic increment and atomic load. The enabled flag supports atomic load, atomic store, atomic swap, and atomic compare-and-swap. The zero value of the component type is usable. Each value lives in its own typed atomic primitive and the component does not invent a transaction across those values.

## 3. Why This Project Now?

- Project 044 was the immediate predecessor and built a fan-out fan-in pipeline whose correctness depended on careful ownership and exact-once delivery.
- Project 045 lowers the level of abstraction to one primitive per value and forces the learner to describe what atomicity covers honestly.
- The goroutine and barrier thinking from Project 031 is reused here for the deterministic concurrent test suite.
- Project 043's verification approach for exact totals under contention is reused here for the typed atomic counters.

## 4. Prerequisites

- Complete Projects 031 and 044 first.
- Project 031 supplies goroutines, barriers, and wait groups reused here for high-contention counter tests.
- Project 044 supplies the exact-totals-under-contention testing approach adapted here for typed atomic counters.
- You must already be comfortable with goroutines, channels, wait groups, the typed atomic primitive types, and barrier-driven concurrent tests with no sleep.
- Earlier projects that introduce atomic versus mutex thinking are useful review but are not required prerequisites.

## 5. What You Must Know Before Starting

- Know that the typed atomic primitives provided by the standard library give single-variable atomicity and sequentially consistent ordering.
- All atomic operations behave as though they occur in one sequential order consistent with program order.
- This does not turn several atomic fields into one transaction.

- Know that several typed atomic fields together do not form one transaction.
- Reading two atomic fields gives the values of each at the moment of its own load; it does not give a snapshot of both together.
- A compound invariant that depends on related atomic fields must be protected by a mutex or designed differently.

- Know that typed atomic unsigned 64-bit counters wrap on overflow at the unsigned 64-bit boundary.
- The wrap is from the maximum value to zero.
- The project documents this honestly and does not present it as safe business behavior.
- Required tests stay safely below the maximum so the wrap is not silently assumed and not silently tested.

- Know that an atomic value, like a mutex, must not be copied after first use.
- The component's ownership crosses function boundaries by pointer.
- Documentation states this and code review treats any copying as a bug.

- Know that benchmarks measure the environment where they ran.
- They are not correctness evidence.
- They are not a universal ranking of atomic versus mutex.
- The project allows optional benchmarks but never treats one environment's result as universal.

## 6. Explanation of New Concepts

### Concepts

- A typed atomic primitive is a value type that carries an underlying integer or boolean and exposes atomic operations.
- The standard library provides a primitive for unsigned 64-bit integers, a primitive for signed integers as appropriate, and a primitive for booleans.
- Each typed primitive guarantees atomicity of its own operations and supplies memory ordering per the language specification.

- A non-negative counter is a typed atomic unsigned 64-bit value.
- Increment adds one to the counter's current value and returns the new value.
- Load returns the counter's current value.
- The counter wraps at the unsigned 64-bit maximum; the project documents that wrap and does not call it safe.

- The enabled flag is a typed atomic boolean.
- Load returns the current value.
- Store writes a new value.
- Swap writes a new value and returns the previous value.
- Compare-and-swap writes a new value only if the current value matches an expected value and returns whether the swap happened.
- The flag carries no implicit meaning; readers interpret it.

- A component carries three typed atomic primitives, one for each value.
- Each value stands alone.
- Operations on one value do not affect another.
- The component does not promise a snapshot across the three values, nor a transactional update across them.

- A bounded counter is named only as an optional extension whose separate contract and tests the learner would have to define.
- The project does not describe the bounded counter's algorithm, shipment, or assumptions.
- The default counter wraps at the unsigned 64-bit boundary.
- An optional benchmark comparing atomic and mutex is also named only as an optional extension; the project does not require that benchmark and does not present its absence as a deficiency.

## 7. Learning Objective

- By completion, you can describe atomicity and memory ordering for typed atomic primitives, use those primitives to implement a small concurrent metrics component, and avoid pretending that several atomic fields form one transaction.
- You can describe unsigned overflow honestly and refuse to use benchmarks as correctness evidence.
- You can write concrete high-contention tests that verify exactly one compare-and-swap win from a false-to-true race and that use a controlled final store after workers join.
- You can write high-contention tests with barriers and wait groups and pass the race detector with no sleep.

## 8. Functional Requirements

1. The component holds three independent values: a non-negative request counter, a non-negative error counter, and an enabled flag.
2. Both counters are typed atomic unsigned 64-bit values. The flag is a typed atomic boolean.
3. The zero value of the component type is usable; every operation works on a fresh value without initialization.
4. The request counter and the error counter support increment and load.
5. The enabled flag supports load, store, swap, and compare-and-swap.
6. Each operation is described behaviorally: what the operation does, what it returns, and what it does not do.
7. The default unsigned counters wrap at the unsigned 64-bit boundary. The required tests stay safely below that maximum so the wrap is not silently assumed and not silently tested.
8. The component does not silently include a saturating compare-and-swap loop. A bounded counter may be named only as an optional extension whose separate contract and tests the learner would have to define.
9. The component does not provide a transactional read of more than one field at once. Compound invariants that span multiple fields must be protected by a mutex or by a different design.
10. The component must not be copied after first use; ownership across function boundaries uses a pointer.
11. Tests cover the zero state, sequential operations, swap and compare-and-swap success and failure, high-concurrency exact totals, a false-to-true compare-and-swap race with exactly one winner, a mixed load/store/swap stress phase followed by a controlled final store after workers join, independent component instances, and the race detector. Tests do not use sleeps.
12. Benchmarks are optional, never a correctness test, and never a universal claim about any comparison. If present, they report results for the measured environment and workload only.

## 9. Inputs and Outputs

### Interface Contract

- The input to increment is nothing; the counter adds one to its current value and returns the new value.
- The input to load is nothing; the counter returns its current value.
- The input to store on the flag is a new value; the flag adopts that value.
- The input to swap on the flag is a new value; the flag adopts the new value and returns the previous value.
- The input to compare-and-swap on the flag is an expected value and a new value; the flag adopts the new value only if its current value equals the expected value, and the operation returns whether the swap happened.

- The output is the new counter value, the previous flag value, or the boolean result of the comparison.
- Each operation reports its own observation and does not report observations about any other field.
- Tests inspect each field on its own.

- Text-only example: starting from zero, eight goroutines each increment the request counter one thousand times; the test reports the final counter value and the expected value, the two must agree, and the comparison must pass on every run, not only on most runs.

## 10. Rules and Edge Cases

- The zero state of a fresh component is a request counter of zero, an error counter of zero, and an enabled flag of false.
- The component is not in an indeterminate state at any observable point during its lifetime.

- A counter whose increment would pass the unsigned 64-bit maximum wraps to zero.
- The project documents this behavior for the chosen unsigned 64-bit type.
- The required tests stay safely below the maximum so wrap is not silently assumed and not silently tested.
- A test that approaches but does not reach the maximum is welcome but is an optional bonus, not a required case.

- A bounded counter and the saturating compare-and-swap loop that implements it are named only as an optional extension.
- The project does not describe the algorithm, does not ship code for it, and does not present its absence as a deficiency.
- Any learner who chooses that extension must define its own contract and its own tests.

- A compare-and-swap with a mismatched expected value returns false and leaves the flag unchanged.
- A compare-and-swap with a matching expected value writes the new value and returns true.
- The flag's load after a failed compare-and-swap observes the value the flag actually held, which may have changed after the expected value was read.

- Concurrent operations on the same field are atomic at the field level.
- Concurrent operations on different fields are independent.
- The component does not give a snapshot across fields and does not pretend two atomic loads done back-to-back are one transactional read.

- The component must not be copied after first use.
- Any method receiver must be a pointer.
- Documentation states the rule; code review enforces it.

## 11. Project Constraints

- Use only the Go standard library.
- Use the typed atomic primitives from the sync/atomic package.
- Do not invent cross-field transactions.
- Do not present unsigned wrap as safe business behavior.
- Do not silently include a saturating compare-and-swap loop.
- Do not use benchmark outcomes as correctness assertions.
- Do not introduce sleep into tests.
- The completed code must pass the race detector.

## 12. Design Questions Before Coding

- What typed atomic primitive holds each value, and what does the type guarantee about that primitive's atomicity and memory ordering?
- What does the component promise about cross-field consistency, and what does it explicitly refuse to promise?
- What unsigned type is chosen for the counters, and how is the wrap behavior documented?
- Is the bounded counter optionally adopted, and if so is the absence of its algorithm in this project explicit?
- How does the component enforce pointer ownership and refuse copies?
- How does each operation describe itself behaviorally so callers understand what they observe and what they do not?
- How is the race detector used to verify the safe behavior of the atomic fields without claiming absence of races in untested paths?
- How does the false-to-true compare-and-swap race test demonstrate exactly one winner?
- How does the mixed load/store/swap stress test use a controlled final store after workers join rather than asserting a scheduling-dependent value?

## 13. Implementation Milestones

1. Define the component type with three typed atomic fields: a non-negative request counter, a non-negative error counter, and an enabled flag. Use the typed atomic unsigned 64-bit primitive for both counters and the typed atomic boolean primitive for the flag.
2. Implement increment and load for each counter.
3. Implement load, store, swap, and compare-and-swap for the flag.
4. Document the unsigned 64-bit wrap behavior at the type boundary and the rule that required tests stay safely below the maximum.
5. Name the bounded counter with a saturating compare-and-swap loop only as an optional extension whose separate contract and tests the learner would have to define. Do not describe its algorithm.
6. Document the no-copy-after-first-use rule and ensure every method uses a pointer receiver.
7. Add tests covering the zero state, sequential operations, swap and compare-and-swap success and failure, high-concurrency exact totals, a false-to-true compare-and-swap race with exactly one winner, a mixed load/store/swap stress phase followed by a controlled final store after workers join, independent component instances, and the race detector.
8. If an optional benchmark is added, keep it separate from correctness and document it as environment-specific.
9. Run the full package under the race detector and correct every issue reported.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Test the zero state: a fresh component reports zero, zero, and false for the request counter, error counter, and enabled flag.
- Test sequential increment and load: increment returns the new value, load returns the same value, and the order is preserved.

- Test swap on the flag: set the flag, swap with a new value, observe the previous value returned and the new value in place.
- Test compare-and-swap success: set the flag, compare-and-swap with the matching expected value, observe the swap happened and the flag holds the new value.
- Test compare-and-swap failure: set the flag, compare-and-swap with a mismatched expected value, observe the swap did not happen and the flag holds the original value.

- Test high-concurrency exact totals: many goroutines increment each counter concurrently for a fixed workload, all complete before the test continues, and the final counter value equals the exact expected total on every run.

- Test the false-to-true compare-and-swap race: the flag begins false, many goroutines attempt a compare-and-swap from false to true, and exactly one call observes true as the swap result while every other call observes false.
- The post-condition is the flag holds true.
- The test does not depend on which goroutine won.

- Test the mixed load/store/swap stress phase: many goroutines perform a mix of load, store, swap, and compare-and-swap on the flag concurrently, all complete before the test continues, and the race detector reports no race.
- After the phase, the test performs a controlled final store on the flag and asserts that controlled final value.
- The test does not assert any scheduling-dependent intermediate or final value observed during the stress phase.

- Test independent component instances: changes to one component do not affect another.
- Test under the race detector with the test binary enabled for race detection and confirm no race is reported on the exercised paths.
- Do not introduce sleep into the tests.
- Use barriers to align goroutine starts and wait groups to await completion.

- If an optional benchmark is added, ensure no correctness test reads its timing or treats one environment's result as universal.

## 15. Common Mistakes to Watch For

- Treating several atomic loads done back-to-back as one transactional read produces a cross-field view that does not exist.
- Presenting unsigned wrap as safe business behavior hides a real bug.
- Silently including a saturating compare-and-swap loop without documenting it makes the component's behavior unclear to callers.
- Defining a wrong compare-and-swap transition in a saturating loop is a class of bug that the project avoids by not describing the algorithm at all.
- Copying a component value after first use duplicates the atomic fields and breaks synchronization.
- Using sleep to give goroutines a chance to race introduces flaky behavior that masks real bugs.
- Treating benchmark outcomes as correctness evidence hides real correctness issues.
- Asserting a scheduling-dependent intermediate or final value during a stress phase produces a flaky test; the controlled final store after workers join is the deterministic endpoint.
- Letting a counter test approach the type's maximum and then asserting on what happened across the wrap is a fragile test; the required suite stays safely below the maximum.

## 16. Topics and References for Study

- Study the standard library documentation for the sync/atomic package, in particular the typed atomic primitives and their memory ordering guarantees.
- Study the documentation for the testing package, in particular the race detector, the barrier pattern, and wait-group-based completion.
- Read the Go memory model documentation for the relationship between synchronization primitives and cross-goroutine ordering.
- Compare this project with Project 031's barrier pattern, the verification approach of Project 043, and the exact-totals-under-contention testing approach of Project 044.

## 17. Self-Assessment Questions

1. What does a typed atomic primitive guarantee about atomicity and memory ordering, and what does it not guarantee?
2. Why do several typed atomic fields together not form one transaction?
3. How is unsigned wrap documented honestly, and why is it not presented as safe business behavior?
4. Why is the bounded counter with a saturating compare-and-swap loop named only as an optional extension rather than silently assumed or shipped?
5. Why must the component be passed by pointer and never copied after first use?
6. What does the race detector prove on exercised paths, and what does it not prove on unexercised paths?
7. Why must benchmarks never be used as correctness evidence?
8. Why does the false-to-true compare-and-swap race test verify exactly one winner, and why does the mixed stress phase use a controlled final store after workers join instead of asserting a scheduling-dependent value?

## 18. Definition of Completion

- [ ] The component holds three typed atomic fields: a non-negative request counter and a non-negative error counter, both typed atomic unsigned 64-bit values, and an enabled flag that is a typed atomic boolean.
- [ ] The zero value of the component type is usable.
- [ ] The counters support increment and load.
- [ ] The flag supports load, store, swap, and compare-and-swap.
- [ ] Each operation is described behaviorally.
- [ ] Unsigned 64-bit wrap at the type boundary is documented and is not presented as safe business behavior.
- [ ] The required tests stay safely below the maximum.
- [ ] The component does not silently include a saturating compare-and-swap loop.
- [ ] The component does not provide cross-field transactional reads.
- [ ] The component is not copied after first use, and the public surface documents and supports that rule.
- [ ] Tests cover the zero state, sequential operations, swap and compare-and-swap success and failure, high-concurrency exact totals, a false-to-true compare-and-swap race with exactly one winner, a mixed load/store/swap stress phase followed by a controlled final store after workers join, independent component instances, and the race detector.
- [ ] Tests do not use sleep.
- [ ] Any optional benchmark is environment-specific and is never a correctness test.

## 19. Optional Extensions

- Define a bounded counter as an optional extension with its own explicit saturation contract and verification cases, without changing the default wrapping counters.
- Add an optional benchmark comparing atomic and mutex implementations whose results are reported only for the measured environment and workload and never treated as correctness evidence.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 044 — Fan-Out Fan-In](../../03-concurrency/044_fan_out_fan_in/README.md#20-prerequisite-based-documentation-guide), [Project 031 — Concurrent Timer](../../03-concurrency/031_concurrent_timer/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- None. This project applies already introduced APIs, standards, and testing practices in a new combination.

### Project-specific learning focus

- **Learn now:** typed atomics, atomicity versus compound invariants, memory ordering, wraparound contracts, contention tests, and correctness versus benchmark evidence.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
