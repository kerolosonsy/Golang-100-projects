# Project 041 — Context Timeout Example

## 1. Project Name and Number

Project 041 — Context Timeout Example. The project is a learning lab for a single cooperative operation that waits for either a supplied work-result signal or the supplied context's done signal, returning success, the canceled sentinel, or the deadline-exceeded sentinel in a way that survives identity-aware inspection at the caller.

## 2. Project Idea

One operation accepts a parent context and an injected work-result signal. The operation returns the reported value when the work signal arrives before any cancellation. The operation returns the standard canceled sentinel when the parent's cancel was triggered without a deadline being passed. The operation returns the standard deadline-exceeded sentinel when the parent context's deadline elapsed before any value arrived. The work signal is an injectable seam so tests can produce a result, block indefinitely, or never produce a result without any real timing. The production demonstration wraps the operation in a finite-timeout parent context and calls the cancel function the context package returned.

## 3. Why This Project Now?

Project 040 was the immediate predecessor and compared mutex and atomic counters under contention. That project isolated how to synchronize shared state. Project 041 moves from synchronizing state to coordinating time and cancellation, the exact responsibility of the context package. The goroutine and barrier thinking the learner already used in Project 031 is reused here for deterministic cancellation tests, and the synchronization vocabulary from Project 040 is reused to argue why the operation must observe the supplied context at every blocking step. The synchronous-callback bus of Project 042 and the cooperative pipeline of Project 044 lean on the same honest context propagation this project establishes.

## 4. Prerequisites

Complete Projects 031 and 040 first. Project 031 introduces the goroutine primitives and barriers reused here for deterministic cancellation tests. Project 040 supplies the synchronization vocabulary that shows why cancellation must be observed at every blocking step. You must already be comfortable with goroutines, channels, wait groups, error sentinels, and the meaning of identity-aware inspection. Earlier projects such as the worker-pool project are useful review but are not required prerequisites.

## 5. What You Must Know Before Starting

Know that a context carries a deadline, a cancellation signal, and a request-scoped value store. Know that the done channel of a context is closed exactly once and that the channel is the place where cancellation and deadline expiry both publish their signal. Know that the context package exports two sentinel errors. Inspection with `errors.Is` is the required way to match those sentinels. The contract for this project is that this inspection continues to match the original sentinel.

Know that a function that performs blocking work must accept a context and must observe it on every blocking call. Know that replacing a caller-supplied context with a fresh background context silently disables the caller's cancellation and is a bug, not an optimization. Know that the parent context's err is the canceled sentinel when the parent was canceled without a deadline, or the deadline-exceeded sentinel when the parent's deadline elapsed, and that returning the err unchanged preserves the errors-aware inspection contract.

Know that any cancel function returned by the context package must be called by the function that created the context. The standard form is a deferred cancel call at the spot where the context was created. Calling the cancel function again is safe but unnecessary; the important contract is that the creator does not omit the call. The production demonstration in this project must create a finite parent timeout and call its cancel function. The core operation only accepts and propagates a context; it must not create a replacement background context or invent a cancel function of its own.

Know that an injected signal is the standard seam for deterministic tests. Its contract is to deliver exactly one usable result or remain blocked until the context ends; closing it without a value is outside this project's input contract and must never be mistaken for a successful zero value. This seam drives every required path without a timer, sleep, or real network call.

## 6. Explanation of New Concepts

The operation selects between the work-result signal and the context's done signal. Both events happen on channels; the select statement is the place where the operation waits for the first ready event and acts on it. When the work signal arrives, the operation returns the value. When the context's done channel fires, the operation consults the context's err and returns that err unchanged so that errors-aware inspection at the caller succeeds.

The injectable work-result signal is the seam that makes the operation testable. A test can supply a signal that already holds a value or one that remains blocked while the test triggers cancellation. The operation never depends on real elapsed time. A signal closed without a value is outside the contract and is not a third result path.

A barrier is used at the start of a test so the operation begins exactly when the test chooses and never races against fixture construction. A wait group at the end of a test awaits the operation's return before the test asserts on outcomes and before the test inspects leak indicators.

A test deadline may be set generously to guarantee that a deadlock fails the test. The test deadline is never the mechanism that decides correctness; every assertion must already be true by the time the operation returns. A test that waits a real second or relies on a real sleep is banned.

The production demonstration is the call site that proves the contract end to end. It creates a parent context with a finite timeout, arranges for the cancel function to be called, and starts the operation. The operation returns the deadline-exceeded sentinel when work outlives that deadline, or the value when work finishes first. The production call must include the real cancel call and the real finite timeout; deterministic tests still use already-expired deadlines and explicit channel coordination.

## 7. Learning Objective

By completion, you can write a cooperative operation that returns a value, the canceled sentinel, or the deadline-exceeded sentinel without breaking errors-aware inspection. You can propagate a parent context to every blocking step and refuse to replace it with a fresh background context. You can decide the production call site's responsibility for the cancel function and can call it correctly. You can write deterministic cancellation tests using injected signals and barriers, with no real timing. You can author a production demonstration that uses a finite parent timeout and exercises the contract end to end.

## 8. Functional Requirements

1. The operation accepts a parent context and an injected work-result signal.
2. The operation returns the value when the work signal arrives before any cancellation.
3. The operation returns the standard canceled sentinel when the parent's cancel was triggered without a deadline expiring.
4. The operation returns the standard deadline-exceeded sentinel when the parent's deadline elapsed before any value arrived.
5. The operation's returned error preserves the sentinel's identity, so errors-aware inspection at the caller matches the original sentinel.
6. The operation must not create a replacement background context inside its body.
7. The operation passes the parent context into every blocking call it makes. The operation does not wrap the parent's err into a new error type and does not rephrase it.
8. The operation must not start any new work when the parent context is already done at entry.
9. The operation observes the supplied context at every blocking step; if the operation spawns an owned goroutine for cooperative work, that goroutine observes the supplied context and the operation does not return until the goroutine has exited.
10. The operation does not call context.Background and does not invent a cancel function of its own.
11. The production demonstration creates a parent context with a finite timeout, calls the cancel function the context package returned, and exercises the operation end to end so the contract is observed in real code.
12. When a result and a cancellation are ready at the same instant, the operation treats the choice as nondeterministic unless the design adds an explicit policy; tests must not depend on which branch select picks.
13. Tests cover success, already-cancelled, explicit cancel while blocked, already-expired deadline, parent cancellation propagation, errors-aware sentinel identity, repeated runs without goroutine leaks, and use no sleep.
14. A generous test deadline may only guard against deadlock and must not be the mechanism that decides correctness.
15. No verification depends on real waiting or on a real sleep.

## 9. Inputs and Outputs

The input is a parent context and an injected work-result signal. The parent context may have no deadline, may support explicit cancellation, or may carry a deadline. The work signal either delivers one usable value or remains blocked until cancellation; it is never closed without a value under this project's contract.

The output is one of three outcomes. The reported value when the work signal beat cancellation. The canceled sentinel when the parent was canceled without a deadline expiring. The deadline-exceeded sentinel when the parent deadline elapsed before any value arrived. The returned error is the parent context's err unchanged in the cancellation cases, so errors-aware inspection at the caller succeeds.

A separate observable is whether the operation left any owned goroutine running after it returned. Tests run a fixed number of operations and assert that no owned goroutine outlives the operation under cancellation.

Text-only example: with a parent context canceled before the operation begins, the operation returns the context's canceled sentinel and errors-aware inspection at the caller matches the canceled sentinel. With a parent context bearing a finite deadline and a work signal that never delivers, the operation returns the deadline-exceeded sentinel and errors-aware inspection at the caller matches the deadline-exceeded sentinel; the test does not introduce any real sleep and any test deadline used guards only against deadlock.

## 10. Rules and Edge Cases

A parent context that is already canceled when the operation starts must cause an immediate return with the canceled sentinel. A parent context whose deadline has already elapsed must cause an immediate return with the deadline-exceeded sentinel. The operation must not block on the work signal in either case because the parent context is already done.

If both the work signal and the context's done channel are ready in the same select call, the runtime may pick either one. The operation treats this race as nondeterministic; tests must not assert that one path was taken when both were ready. Tests that need a deterministic check must ensure only one channel is ready at the instant of selection.

The operation must not wrap the parent's err. Returning ctx.Err() unchanged is the simplest and most correct choice. The contract is errors-aware inspection, not direct equality; a caller that uses the standard errors helpers continues to match the sentinel even if the operation ever introduced a wrapping layer, as long as the wrapping layer is honest.

The production caller, not the operation, owns the cancel function returned by the context package. The production caller must call that cancel function, typically with a deferred call at the spot where the context was created. Calling the cancel function twice is safe but unnecessary; the rule is that the creator must always call a cancel function at least once, not that the creator must call it under penalty.

An operation that starts an owned goroutine for cooperative work must observe the supplied context inside that goroutine and must not return until the goroutine has exited. An operation that only receives an injected signal need not invent a goroutine and need not invent a release or stop call.

A leaky owned goroutine is a bug. Tests run a fixed number of operations, then count active work or count unblocking signals, and fail if any work is still active under cancellation.

## 11. Project Constraints

Use only the Go standard library. Use the context package, channels, goroutines, and wait groups. Use an injected work-result signal so tests stay deterministic. Do not introduce real timers inside the operation. Do not introduce sleeps inside tests; if a test deadline is used it guards deadlock only. Do not create a replacement background context inside the operation. Do not store a context inside a struct field that outlives a single call. The completed code must pass the race detector.

## 12. Design Questions Before Coding

What is the smallest work-result signal seam that lets tests simulate a value or blocked work without timing? How does its contract exclude closure without a value? How does the operation distinguish canceled from deadline-exceeded using only the parent's error? How does the operation preserve `errors.Is` matching across every return path? When does the operation consult the parent's done channel without ever blocking on the work signal because the parent is already done? How does the production caller create a finite timeout and call its cancel function, and how is that responsibility distinguished from the core operation's responsibility? How does each test ensure only one channel is ready at the moment of selection so no test depends on which branch wins? How does the operation prove that no owned goroutine outlives its return when cancellation triggered the exit? Why is the operation allowed to receive an injected signal without inventing a goroutine of its own?

## 13. Implementation Milestones

1. Define the work-result signal seam and document what it represents.
2. Implement the operation body that selects between the work signal and the context's done signal.
3. Return the value when the work signal arrives.
4. Return the parent's err unchanged when the context's done channel fires, choosing the canceled or deadline-exceeded sentinel automatically through the parent's err.
5. Skip the work signal entirely when the parent context is already done at entry.
6. If the operation spawns an owned goroutine for cooperative work, ensure the goroutine observes the supplied context and that the operation does not return until the goroutine has exited.
7. Author the production demonstration with a finite parent timeout and the cancel function call, so the contract is observed end to end.
8. Add tests covering success, already cancelled, explicit cancel while blocked, already-expired deadline, parent cancellation propagation, errors-aware sentinel identity, repeated runs, and zero sleep.
9. Run the full package under the race detector and correct every issue reported.

## 14. Verification Cases the Learner Must Write

Test the success path: an in-memory signal carries a value, the operation returns that value, and no goroutine outlives the operation. Test the already-canceled path: a parent context canceled before the operation begins, the operation returns the parent's error and `errors.Is` matches the canceled sentinel. Test the explicit-cancel-while-blocked path: a signal that never delivers and a cancel call arriving after a channel handshake proves the operation is blocked, the operation returns the parent's error and `errors.Is` matches the canceled sentinel. Test the already-expired-deadline path: a parent context with a deadline already in the past, the operation returns the parent's error and `errors.Is` matches the deadline-exceeded sentinel. No test waits for a future deadline to elapse.

Test parent-cancellation propagation through a derived context: a parent context canceled while the operation observes a child derived from it, the child done channel fires and the operation returns the canceled sentinel. Test errors-aware sentinel identity across every return path so a caller using the standard errors helpers matches the original sentinel. Test repeated runs that exercise the operation many times in sequence and confirm no owned goroutine outlives the operation in any run.

Use barriers to align tests with the operation's start. Use wait groups to confirm completion. Use a test-deadline guard only as a deadlock detector and never as the deciding assertion. Do not introduce sleep into the operation or its tests. Do not require a real second of waiting anywhere in the test suite.

## 15. Common Mistakes to Watch For

Creating a replacement background context inside the operation silently disables caller cancellation. Storing a context inside a struct field that survives the call lets a later caller see a stale signal. Replacing the parent's error or wrapping it without preserving an unwrap path breaks `errors.Is` matching; honest wrapping can preserve it, although this project simply returns the context error unchanged. Treating selection between a ready work signal and a ready context done channel as deterministic produces flaky tests. Treating a closed-without-value work signal as success invents a false zero-value result. Using sleep inside a cancellation test makes the test pass by accident and fail under load. Forgetting to call the cancel function in the production caller retains context resources unnecessarily. Inventing a release or stop contract that the injected signal does not require contradicts the simple seam. Relying on a test deadline as the correctness assertion hides logic bugs that would otherwise surface.

## 16. Topics and References for Study

Study the standard library documentation for the context package, including the deadlines, cancellation, and the done channel contract. Study the standard library documentation for the errors package and identity-aware inspection. Read Go's blog post on the context package and the standard explanation of the canceled and deadline-exceeded sentinels. Compare this project with Project 031's barriers and the synchronization vocabulary of Project 040. Project 044's pipeline propagates context in the same way and inherits the production-caller responsibility for the cancel function.

## 17. Self-Assessment Questions

Why must the operation not replace the parent context with a fresh background context? Why is the work signal an injectable seam, and what does that buy for tests? Why is select's choice between a ready work signal and a ready context done channel considered nondeterministic, and how does that shape tests? Why must the canceled and deadline-exceeded sentinels keep their error identity under the errors-aware inspection contract? How does the operation prove that no owned goroutine outlives its return when cancellation triggered the exit? Why is a test deadline only a deadlock guard and not a correctness assertion? Why must the production caller always call the cancel function the context package returned, and why is the operation itself not responsible for that call?

## 18. Definition of Completion

The operation accepts a parent context and an injected work-result signal and returns the value, the canceled sentinel, or the deadline-exceeded sentinel. The returned error in the cancellation cases is the parent context's err unchanged, so errors-aware inspection at the caller matches the original sentinel. The operation passes the parent context into every blocking call. The operation does not call context.Background and does not invent a cancel function of its own. The operation does not start new work when the parent context is already done at entry. The production demonstration creates a finite parent timeout, calls the cancel function the context package returned, and exercises the contract end to end. Tests cover success, already cancelled, explicit cancel while blocked, already-expired deadline, parent cancellation propagation, errors-aware sentinel identity, and repeated runs without goroutine leaks. Tests use no sleep and no verification depends on real waiting. The full package passes the race detector.

## 19. Optional Extensions

Add a small teaching note that explains, in plain prose, when a deadline and a cancel race for the same instant and how the operation's nondeterministic policy interacts with that race. Add repeated, explicitly coordinated cases where only one signal is ready at selection time and confirm the matching outcome without relying on randomized scheduling.
