# Project 015 — CLI Counter

## 1. Project Name and Number

Project **015** — `015_cli_counter`. The directory name and number must match exactly.

## 2. Project Idea

A small command-line program that counts either up from a starting value to a target or down from a starting value to a target, one step at a time, and prints the current value at each step. The starting value is emitted immediately, before any tick. Every later value follows a tick. The counter runs until it reaches the target, until the user cancels it, or until the program decides there is nothing to count (the zero-count case where the starting value already equals the target).

The counter's tick source is injected from the outside. The production program uses a real `time.Ticker` driven by the system clock. Tests use a fake tick source that delivers deterministic tick events on demand and records when its `Stop` method was called. Cancellation flows through a `context.Context` so a test can stop the counter deterministically.

## 3. Why This Project Now?

Projects 011 through 014 introduced injected I/O, integer-cents arithmetic, deterministic time conversion, and per-field validation. None of them used the concurrency primitives that the rest of the path will rely on. Project 015 introduces the smallest useful piece of time-driven behavior: a loop that ticks at a chosen interval, can be cancelled through `context`, and releases its tick source cleanly when it stops.

This is the first project in which the **lifecycle of a resource** matters. The contract is: every tick source the program creates is stopped exactly once on the path that creates it, and the test can observe both the events and the cleanup. The counter never reads a wall clock directly; everything time-related flows through the injected tick source.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 015 therefore requires:

- Completion of **014** (Input Validator).
- No prior knowledge of HTTP, databases, or generics.
- Familiarity with `time.Ticker` is helpful but not required; this project introduces it.

## 5. What You Must Know Before Starting

- What a `time.Ticker` is in Go: a value that delivers "tick" events on a channel at a chosen interval, until it is stopped.
- What `time.NewTicker` and `Ticker.Stop` do. Modern Go can garbage-collect tickers that are no longer referenced, but explicit `Stop` is still the right discipline: it makes ownership obvious, releases the runtime timer slot promptly, and is what a test can observe.
- What a `context.Context` is and how `context.WithCancel` creates a cancellable context whose `Done` channel closes when cancelled.
- How `select` can wait on multiple channels: the tick channel from the tick source, the `Done` channel from the context, and any other channel the loop uses.
- The difference between counting up and counting down: how the loop decides when to stop, and how it handles the zero-count case where the target is already reached at start.
- Why `time.Sleep` is not appropriate inside the production loop if the loop must also respond to cancellation quickly.

## 6. Explanation of New Concepts

### Why an injected tick source

A `time.Ticker` is bound to real time. To test a counter without real sleeps, the program must be able to receive tick events on demand. The clean way to express that seam is to inject a small abstraction that has two responsibilities: deliver tick events and stop itself. Production wires it to a real `time.Ticker`; tests wire a fake that the test controls. The README does not prescribe a Go interface or a function signature — it only describes the contract:

- The tick source delivers tick events one at a time.
- The tick source has a `Stop` method that the counter calls on the path that created the source.
- The fake used in tests lets the test push a tick when it wants one and lets the test observe whether `Stop` was called.

### Why "round once at the boundary" applies here too

A counter that reads the wall clock directly is hard to test. The injected tick source means the only place the program touches real time is the production wiring, and the only thing tests need to control is the sequence of tick events. The counter itself does not call `time.Now` and does not own a wall-clock value.

### Cancellation through `context`

A `context.Context` carries a `Done` channel that closes when the context is cancelled. A subsequent `select` on that channel fires immediately. The counter's loop listens on both the tick channel and the `Done` channel; whichever fires first wins. When cancellation wins, the loop exits and the cleanup runs.

### Lifecycle of the tick source

The contract for the tick source's lifecycle is precise:

- Paths that do not create a tick source do not stop one. These paths are: invalid input, the zero-count case, and failure while writing the immediate starting value. On these paths the counter returns before source creation, so there is no `Stop` call to make.
- Paths that create a tick source stop it exactly once. These paths are: target reached after one or more ticks, context cancelled, and failure while writing a later value or final status. On these paths the counter calls `Stop` exactly once on the way out.

A test can pin the contract by counting `Stop` calls and asserting the count is `0` on the no-source paths and `1` on the source-owning paths.

### Initial output before any tick

The starting value is emitted immediately, before the loop waits for the first tick. Every later value follows a tick. This makes "counting up from `0` to `3`" produce exactly four printed lines (`0, 1, 2, 3`) with three ticks between them, not three printed lines (`0, 1, 2`) followed by a tick that produces the final value.

### Verification without real sleeps

A test that calls `time.Sleep` to wait for the counter is a flaky test: it depends on scheduling and may time out on a slow machine. The contract for this project is: every verification case uses the injected tick source to produce a deterministic sequence of events. No test calls `time.Sleep` to advance the counter; no test waits for the counter to "finish on its own"; the test *causes* the events and then asserts the outcomes.

## 7. Learning Objective

After completing this project the learner can:

- Build a loop driven by an injected tick source, with the starting value emitted immediately and every later value following a tick.
- Cancel a tick-driven loop through `context` and stop the tick source exactly once on every path that created it.
- Express the tick source as a seam the test can drive and observe, without prescribing a Go interface or signature in this README.
- Handle invalid input and the zero-count case before any tick source is created, so there is nothing to stop on those paths.
- Write tests that drive a counter through cancellation, target reached, and writer failure by sending ticks from a fake source, with no real sleeps.
- Verify that `Stop` was called exactly once on each source-owning path and zero times on the no-source paths.

## 8. Functional Requirements

1. Accept a starting value, a target value, a direction (up or down), an interval, a `context.Context`, and an `io.Writer` for output.
2. Validate the inputs first. An invalid direction (target on the wrong side of start) and a non-positive interval are rejected with a clear error. On these paths no tick source is created and there is nothing to stop.
3. Handle the zero-count case next. If the starting value already equals the target, emit the starting value and the "already at target" line, then return. On this path no tick source is created and there is nothing to stop.
4. Emit the starting value through the injected `io.Writer` immediately, before any tick.
5. Create the tick source via a factory the program owns. The production factory builds a real `time.Ticker` from the interval; the test factory builds a fake the test controls.
6. On each tick event, compute the next value, write it through the injected writer, and stop if the new value equals the target.
7. Listen on both the tick channel and the context `Done` channel. Whichever fires first wins.
8. If cancellation wins, write the cancellation line and stop the tick source.
9. If the immediate starting-value write fails, return before creating a tick source. If a later write fails after source creation, stop that source exactly once and return.
10. If the target is reached after a tick, write the "reached target" line and stop the tick source.
11. On every path that created a tick source, call `Stop` exactly once.

## 9. Inputs and Outputs

### Inputs

- Direction: up or down.
- Starting value: an integer.
- Target value: an integer.
- Interval: a `time.Duration` typed in a human-readable form, for example `500ms`, `1s`, `2m`.
- Context: a `context.Context`.
- Writer: an `io.Writer`.

### Outputs

- One line containing the starting value, emitted immediately.
- One line per tick, containing the next value.
- A final line stating why the counter stopped: target reached, cancelled, or already at target.
- All output goes through the injected `io.Writer` so a test can capture it.

### Example text-only success run (counting up, 0 to 3, interval 1s)

```
Step: 0
Step: 1
Step: 2
Step: 3
Reached target 3.
```

### Example cancellation run

```
Step: 0
Step: 1
Cancelled at value 1.
```

### Example zero-count run

```
Step: 5
Already at target 5.
```

(The starting value `5` is emitted even though no tick is needed. There is no tick source on this path.)

## 10. Rules and Edge Cases

- **Direction up, target above start**: counts from start to target inclusive. The starting value is emitted immediately; each later value follows a tick. The last printed value equals the target.
- **Direction down, target below start**: same behavior, descending.
- **Direction up, target below start**: invalid direction; the program rejects the input with a clear error before any output and before any tick source is created.
- **Direction down, target above start**: invalid direction; same rejection.
- **Target equals start (zero-count case)**: the starting value is emitted, the "already at target" line is written, and the program returns. No tick source is created on this path.
- **Interval of zero**: rejected with a clear error before any tick source is created.
- **Negative interval**: rejected with a clear error before any tick source is created.
- **Cancellation before the first tick**: the starting value has already been emitted; the cancellation line is written and the tick source is stopped.
- **Cancellation after the target is reached**: the loop has already exited on the target-reached path and the tick source has already been stopped; the cancellation must not undo the cleanup.
- **Writer failure before source creation**: if writing the immediate starting value fails, the program returns and no source is created or stopped.
- **Writer failure after source creation**: if writing a later value or final status fails, the loop exits and the tick source is stopped exactly once.
- **Tick source lifecycle**: every source-owning path calls `Stop` exactly once; no-source paths never call `Stop`.

## 11. Project Constraints

- Go standard library only. No third-party timing libraries.
- The tick source is injected. The production program wires a real `time.Ticker`; the test program wires a fake. This README does not prescribe the Go interface or signature; the learner chooses one.
- The loop never calls `time.Sleep` to wait between steps; it waits on the tick channel.
- On every source-owning path, the tick source is stopped exactly once. On every no-source path, the tick source is never stopped. The package documentation must state both halves of this rule.
- No persistence, no scheduling across program runs, no alarms — out of scope.
- The interactive prompt accepts plain numeric and direction inputs; the validation rules are the learner's, but the rejection of invalid inputs is required.

## 12. Design Questions Before Coding

- Where does the tick source live — created inside the counter, created in the caller and handed in, or built by a factory the counter calls? Which shape lets the test observe `Stop` calls?
- How does the factory produce the production tick source and the fake tick source? How does the production wiring pick the production factory and the test wiring pick the fake factory?
- When the direction is up and the starting value is already past the target, does the counter reject the input or print nothing? How is the rejection reported?
- How is the "cancelled" line distinguished from the "target reached" line and from the "already at target" line? Does the test pin the wording?
- How does the test cancel the context? Through a cancellable context the test holds, or by reaching a deadline?
- How does the test confirm `Stop` was called exactly once on each source-owning path and zero times on each no-source path?
- How does the test deliver ticks on demand? A channel the test pushes to, a function the test calls, or another small mechanism?

## 13. Implementation Milestones

1. Define the inputs the counter takes: direction, start, target, interval, context, writer, and a tick-source factory.
2. Validate the inputs up front. Return a clear error for invalid direction and non-positive interval. Do not create a tick source on this path.
3. Handle the zero-count case: if start equals target, emit the starting value and the "already at target" line, then return. Do not create a tick source on this path.
4. Emit the starting value through the injected writer.
5. Create the tick source via the factory. This is the first source-owning path.
6. Build the loop: `select` on the tick channel and the context `Done` channel. On each tick, compute the next value, write it, and stop if it equals the target. On cancellation, write the cancellation line.
7. On every source-owning exit — target reached, cancelled, or a write failure after source creation — call `Stop` exactly once. A failure while writing the initial value returns before source creation.
8. Wire the production program to a real `time.Ticker` factory and a `context.Background` (or a `WithCancel` the user can trigger).
9. Confirm that a test can drive the counter through every exit path without any real sleep, using a fake tick source whose `Stop` calls are observable.

## 14. Verification Cases the Learner Must Write

Each case is described in natural language. The fake tick source, the context, and the writer are all controlled by the test; no case calls `time.Sleep` and no case waits for a real ticker to tick.

### Output and counting

- Counting up from `0` to `3` with three ticks delivered by the fake source produces the values `0, 1, 2, 3` and a "reached target" line, in that order. The starting value `0` is emitted before any tick.
- Counting down from `3` to `0` with three ticks produces the values `3, 2, 1, 0` and a "reached target" line. The starting value `3` is emitted before any tick.
- Cancelling the context after the starting value and one tick produces `0, 1` and the cancellation line.
- Cancelling the context before any tick is delivered still produces the starting value, then the cancellation line.

### Lifecycle

- On the target-reached path, the fake tick source's `Stop` was called exactly once.
- On the cancellation path, the fake tick source's `Stop` was called exactly once.
- When the writer rejects the immediate starting value, no tick source was created and `Stop` was called zero times.
- When a write fails after at least one tick, the fake tick source's `Stop` was called exactly once.
- On the zero-count path, no tick source was created and `Stop` was called zero times.
- On the invalid-direction path, no tick source was created and `Stop` was called zero times.
- On the non-positive-interval path, no tick source was created and `Stop` was called zero times.

### Independence from real time

- Across all the above cases, the test file contains zero calls to `time.Sleep` (the learner can search the file to confirm).
- Across all the above cases, no production ticker is ever constructed; the fake source is the only source the counter sees.

## 15. Common Mistakes to Watch For

- **Reading the wall clock directly in the counter.** The counter only knows about tick events; it does not call `time.Now`. If you find yourself reaching for the current time, the design has drifted away from the injected tick source.
- **Emitting the starting value after the first tick.** The contract is "starting value immediately, every later value follows a tick". Reversing the order produces wrong outputs and wrong test expectations.
- **Forgetting to call `Stop` on a source-owning path.** Each source-owning path calls `Stop` exactly once; a missing call leaks the underlying timer slot and breaks the lifecycle test.
- **Calling `Stop` on a no-source path.** Paths that never created a tick source must not call `Stop`; doing so either panics on a nil source or calls `Stop` on something that was never started.
- **Calling `Stop` more than once.** The lifecycle contract is "exactly once on each source-owning path". Multiple calls do not free anything twice but signal a confused ownership story.
- **Using `time.Sleep` inside the loop.** That makes the loop unresponsive to cancellation; the wait must come from the tick channel.
- **Reading the tick channel without selecting on the context `Done` channel.** A loop that only reads the tick channel is uncancellable; the test will hang when it tries to cancel.
- **Closing the writer from inside the loop.** The writer belongs to the caller; closing it is the caller's responsibility.
- **Promising real-time accuracy.** A real ticker that ticks every `500ms` ticks *roughly* every 500ms; tests must not depend on the exact wall-clock interval. The fake source has no such uncertainty.
- **Promising "no leaks" by relying on the garbage collector.** Modern Go can collect unreferenced tickers, but that is not the lifecycle contract; explicit `Stop` is what makes ownership testable.

## 16. Topics and References for Study

- A Tour of Go: "Concurrency", "Select".
- Effective Go: "Concurrency", "Cancellation".
- Package documentation: `time` (`NewTicker`, `Ticker.Stop`, `Tick`, `After`, `Since`), `context` (`Background`, `WithCancel`, `Done`, `Err`, `Deadline`).
- Lifecycle of timers: search for "Go ticker leak", "time.Ticker Stop responsibility", "context cancellation patterns".
- Test patterns: search for "deterministic ticker test Go", "fake ticker Go", "fake clock Go".

## 17. Self-Assessment Questions

1. What does `time.Ticker.Stop` do, and what does the lifecycle contract require the counter to do with it?
2. Why is `time.Sleep` a poor choice for the wait inside a cancellable loop?
3. What does a `context.Context` `Done` channel signal, and how does `select` use it?
4. What is the zero-count case, and why does the contract say "no tick source is created on this path"?
5. How does the test deliver ticks on demand, and how does it confirm the counter stopped the tick source on every source-owning path?
6. Why is "starting value emitted immediately, every later value follows a tick" the right output ordering?
7. If you had to count for an unbounded duration with no target, what would change about the lifecycle contract?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every functional requirement in section 8 is satisfied.
- Every verification case in section 14 has a corresponding test, and no test calls `time.Sleep` to advance the counter.
- The package documentation states both halves of the lifecycle rule: every source-owning path calls `Stop` exactly once, every no-source path calls `Stop` zero times.
- The counter's loop is reachable from a test that controls the tick source, the context, and the writer.
- The test file contains zero calls to `time.Sleep` (the learner can search the file to confirm).
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Step callback.** Add an optional hook that is called with each value before it is printed; the hook can transform the value or record it for the test. Keep the hook simple — one function value, no interface ceremony.
- **Step limit.** Accept an optional maximum number of ticks; if the counter would exceed that limit before reaching the target, it stops with a "step limit reached" line and stops the tick source exactly once. Do not add scheduling, alarms, or real-time guarantees.
