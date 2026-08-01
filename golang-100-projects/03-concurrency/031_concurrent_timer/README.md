# Project 031 — Concurrent Timer

## 1. Project Name and Number

Project 031 — Concurrent Timer. Lives in `03-concurrency/031_concurrent_timer`.

## 2. Project Idea

Build a driver that runs several named one-shot timers concurrently and gathers a structured result for every timer. Zero-duration timers complete immediately, negative durations are rejected before any goroutine starts, and cancellation stops every still-pending timer exactly once through the goroutine that owns it. The final aggregate is sorted by the original input order, never by nondeterministic completion order, and duplicate names are refused up front.

## 3. Why This Project Now?

Project 015 introduced the ticker with cancellation, and Project 027 introduced generic value containers. Both give you the channel and value-shape primitives. Project 031 is the first project of Level 3 that combines goroutines, channels, and the `context` package into a multi-result fan-in. Projects 032 and 033 reuse this fan-in pattern at smaller and larger scales. This project teaches the bridge between them: the many-goroutines, many-results, cancellation-on-everything shape that the next several concurrency projects reuse.

## 4. Prerequisites

The curriculum map's stated dependencies for this project: Project 015 (CLI counter with ticker and cancellation) and Project 027 (Custom Stack and Queue). You must be comfortable launching goroutines, sending and receiving on channels, and using `sync.WaitGroup`. A passing familiarity with `context.WithCancel` is required even though Project 041 has not yet been taught, because cancellation is the heart of this project. Self-study the relevant pages before starting.

## 5. What You Must Know Before Starting

The lifecycle of an unbuffered channel versus a buffered channel, and what happens to a send when no goroutine is receiving on the other end. How `sync.WaitGroup` moves its counter and the rule that positive `Add` calls accounting for already-launched goroutines must happen before those goroutines start and before any `Wait` that races with them. Adding later is unsound under that race, not categorically forbidden. How `time.Timer` is constructed and how its channel fires once, and how `time.Ticker`'s `Stop` method releases the underlying timing resource and prevents future ticks. That `context.Context` cancellation is a request, not a guarantee of immediate termination, and that goroutines observe it through the `Done` channel inside a `select`. That `select` picks one ready case pseudo-randomly and gives no application-level priority guarantee when multiple cases are ready. That `time.Timer.Stop`'s boolean return value does not, by itself, prove that the event was already received by some consumer.

## 6. Explanation of New Concepts

A `time.Timer` produces exactly one event on its channel after its duration elapses and then becomes inert. A `time.Ticker` produces events on a steady cadence; its `Stop` method releases the underlying timing resource and prevents future ticks without claiming a dedicated goroutine to free. Both have a documented stop method. The boolean return value alone does not, by itself, prove that the event was already received by any consumer. Do not assume an unconditional channel-drain recipe; the targeted Go version's `Timer` contract and the sequencing around `Stop` and the channel are version-dependent.

The pattern this project solidifies is "fan-out, fan-in". You spawn one goroutine per timer. Each goroutine waits on a pair of signals — its timer firing or the context being cancelled. Each goroutine sends one structured outcome into a shared channel. A collector assembles the outcomes. Because completion order is nondeterministic, the final slice must be sorted by the position the timer held in the input, not by the order the goroutines happened to send.

Cancellation in Go is cooperative. The driver must visit every outstanding timer on the cancellation path and apply `Stop` exactly once per timer if the timer is still pending, by the goroutine that owns the timer. Define the timer-versus-cancellation race by observable ownership and state: a completion already committed to a result slot remains completed; for still-pending timers, cancellation claims the slot and reports cancelled. Reason about the interleaving through the ownership boundary, not through an assumed `select` priority.

An explicitly owned `time.Timer` is required here, because cancellation has to call `Stop` on it and tests have to observe that the same goroutine invoked `Stop` exactly once. Runtime reclamation details for timer allocations differ by Go version; this project relies only on the ownership boundary, not on a specific reclamation recipe.

## 7. Learning Objective

After this project you can explain, in your own words, the difference between `time.Timer` and `time.Ticker`, including their lifecycle and stop semantics. You can describe who owns closing each channel and who consumes the outputs. You can run a fan-out, fan-in pipeline that respects context cancellation and reason about ordering independently of completion order. You can distinguish "the work completed" from "the work was cancelled" in a multi-result aggregate and you can write tests that prove this distinction without wall-clock waits, using a fake timer boundary the tests control.

## 8. Functional Requirements

1. The driver accepts a list of named timer specifications as input. Each specification pairs a string name with a duration.
2. The driver rejects, before starting any goroutine, any input that contains a duration less than zero or any duplicate name across the input.
3. A duration of exactly zero is accepted and is completed immediately in the same call without any real waiting.
4. The driver returns one structured result per input timer. The result slice is in the original input order even when timers complete in a different order.
5. Each result carries the input name, the input position, and one of three terminal statuses: completed, cancelled, or rejected.
6. The driver accepts a `context.Context`. When the context is cancelled, every still-pending timer is stopped exactly once by the goroutine that owns it, and the corresponding position in the result slice is marked cancelled.
7. Results that were already produced before cancellation are marked completed; they are never marked cancelled.
8. Empty input returns an empty slice with no goroutines launched.
9. Negative-duration entries and duplicate-name entries are reported as rejected at their input positions; no goroutines are started for any entry when the input fails preflight.

## 9. Inputs and Outputs

**Input** is a slice of named timer specifications and a context. A specification is a string name paired with a duration. The name must be non-empty; the duration may be zero or positive and is rejected if negative.

**Output** is a slice of length equal to the input length, in input order. Each entry carries the input position, the name, the terminal status, and for completed entries any value the timer generation step chooses to surface.

**Behaviour example (text only).** Input has three names in the order `["fast", "slow", "mid"]`. The `slow` timer is cancelled before it would have fired. The output slice has three entries in the order `["fast", "slow", "mid"]`. The `slow` entry reads `cancelled`. The `fast` and `mid` entries read `completed` once they have fired. None of the names leak into another entry's slot.

**Behaviour example (text only).** Input has two entries sharing the same name. The result slice has two entries. Both are `rejected`. No goroutine ever starts.

## 10. Rules and Edge Cases

A zero duration completes immediately; the immediate value of the timer channel is consumed rather than read on a deadline. A negative duration is treated as a preflight violation and the call returns without scheduling anything. Duplicates are detected across the entire input before scheduling; the call returns rejected entries at their positions without scheduling anything. Cancellation is observed at least once per outstanding goroutine; `Stop` is invoked exactly once per still-pending timer by the goroutine that owns it. A timer whose completion has already been committed to a result slot remains completed, regardless of when cancellation arrived. A still-pending timer whose cancellation arrives before it fires becomes cancelled. Empty input produces an empty slice and zero goroutines. Input order in the output is by the original index, not by completion time. The "completed versus cancelled" distinction is preserved exactly at the result-slice level, not collapsed for brevity. Recipes like "drain the timer channel after `Stop`" depend on the targeted Go version's `Timer` contract and are not assumed without a specific version reference.

## 11. Project Constraints

Standard library only. No third-party packages. No real waiting in tests; any timing primitive used by tests is replaced with a controllable seam so the test does not depend on wall-clock durations. The core logic accepts a timer-creation seam conceptually so tests can deliver events and observe stop calls without using real timers; the seam's interface and signature are not prescribed. No `time.Sleep` is used in any test. No reliance on the Go scheduler for ordering. Do not make absolute claims about drain patterns that ignore Go version or API races; rely on the current documentation for the version you target. The race detector must report nothing under `-race`. Cancellation reaches each outstanding timer through the goroutine that owns it; that goroutine invokes `Stop` at most once if the timer is still pending, and follows the targeted Go version's `Timer` contract. The cancellation path does not blindly drain any timer channel.

## 12. Design Questions Before Coding

How will you represent a single timer specification so that the name, the duration, and the result slot are all unambiguous — a single struct, a pair, or a generic value? How will three terminal statuses propagate through a single result type without confusing them at the call site? Where does the collector live, and how does it know how many results to expect without counting on completion order? How will you ensure that `Stop` is invoked exactly once per still-pending timer, including the case where the timer fired and the case where it did not? How will the final slice be ordered by input index rather than by completion time? What is the seam where the test substitutes a controllable timer for the real one, and what does the seam promise to deliver and to observe? How will you place the `WaitGroup.Add` call so that it occurs before each goroutine starts and before any `Wait` that races with it? How will you decide whether the collector returns early on preflight failure or schedules nothing and reports per-position rejections? What does "completed" mean for a zero-duration timer in terms of observable state, separate from whatever the timer channel's value carries?

## 13. Implementation Milestones

1. Write down the input and output types on paper, including the three terminal status values, before any code is written.
2. Implement preflight validation: empty input short-circuits, negative duration is rejected, duplicate names are rejected across the whole input.
3. Identify where the timer-creation seam will live in the core logic. Write down what the production code needs from the seam and what the test seam must deliver; do not pick a Go signature yet.
4. Implement the fan-out: one goroutine per accepted timer, each waiting on either the timer signal or the context cancellation through `select`.
5. Implement the collector: one result per goroutine, placed into a result buffer under indexed ownership.
6. Reorder results by input position before returning.
7. Layer in cancellation: each owning goroutine observes `ctx.Done()`, invokes `Stop` at most once if its timer is still pending, and marks its position as cancelled; already-committed positions remain as recorded.
8. Cover the zero-duration and empty-input paths without launching goroutines.
9. Make every status readable at the call site without ambiguity.
10. Run under `-race` and confirm no race report on the result buffer or on the per-timer stop flag.

## 14. Verification Cases the Learner Must Write

Empty input returns an empty slice and launches no goroutines. A single zero-duration timer returns one completed result at position zero without any real waiting. Several timers with mixed durations, controlled by an injected boundary that releases events in a chosen order, complete in that order and the output preserves input order, not completion order. Duplicate-name input is rejected for every input position, with no goroutines started. Negative-duration input is rejected at the preflight step with no goroutines started. Cancellation before any timer fires produces one cancelled result per input position, with `Stop` called exactly once per still-pending timer by the goroutine that owns it. Cancellation after some timers have already fired produces completed results for the timers that fired and cancelled results for the rest, with `Stop` called only for the timers that had not yet fired. If the timer-creation seam supports a factory failure, a failure produces a rejected result at the failing position without leaking goroutines. Repeated runs of the same input produce the same status pattern and never deadlock. The final slice is always exactly as long as the input and always sorted by input index. Running under `-race` produces no race report. No real `time.Sleep` is used in any test. The fake timer seam the tests substitute observes that `Stop` is invoked exactly once by the owning goroutine.

## 15. Common Mistakes to Watch For

Placing a `sync.WaitGroup.Add` so late that a racing `Wait` may observe an incomplete counter. The safe shape is to place `Add` before each goroutine launches and before any `Wait` that races with it; absolute rules forbidding every later `Add` are over-broad. Treating cancellation as a silent kill: a cancelled goroutine does not blindly drain the timer channel; its single owner invokes `Stop` at most once if the timer is still pending and follows the targeted Go version's `Timer` contract. Calling `Stop` more than once on the same timer, or treating the boolean return value as conclusive proof that the event was already received by a consumer. Allowing two goroutines to write into the same result slot without indexed ownership. Reading the final slice straight off the output channel and assuming it is in input order — channel order is delivery order, not input order. Embedding a `context` inside a struct instead of passing it as a parameter, which breaks cancellation propagation. Hard-coding an unconditional channel-drain recipe that ignores Go version behaviour. Trusting that `time.After` allocations require manual reclamation on long paths — reclamation details vary by Go version and this project relies only on explicit ownership through `time.Timer`.

## 16. Topics and References for Study

The `time` package documentation, especially the entries for `Timer`, `Ticker`, and `After`. The `sync` package documentation, especially `WaitGroup`, the ordering guarantee on `Wait`, and the rules for `Add`. The `context` package documentation, especially `WithCancel`, `Done`, and the cancellation-propagation rules. The Go specification on channel operations, send and receive semantics, and the `select` statement. The Go blog or release notes that introduced the current timer implementation, including any notes about the runtime call inside a timer. The Effective Go notes on channels, goroutines, and the fan-in pattern.

## 17. Self-Assessment Questions

What is the difference between `time.Timer` and `time.Ticker` in terms of how many events they emit, what stops each, and what `Stop` actually releases? Why does the final result slice have to be sorted by input position rather than by the order the goroutines happened to send? If a timer fires and the same instant the context is cancelled, which status takes precedence, and why is the tie resolved by the observed ownership boundary rather than by an assumed `select` priority? Where in this design is the `WaitGroup.Add` call placed, and what would happen if it were placed inside a goroutine so that `Wait` could observe an incomplete counter? If the test seam provides a controllable timer, what does the seam need to expose so the test can verify that `Stop` was called exactly once by the owning goroutine on still-pending timers? Why is "duplicate name" a preflight rejection rather than a per-item rejection at collection time? What is the consequence of allowing two goroutines to write into the same result slot without indexed ownership? Why does empty input not launch goroutines, and how would a test prove that? Why does this project depend on an explicit `time.Timer` instance rather than on `time.After`, given that cancellation must call `Stop` and tests must observe that ownership?

## 18. Definition of Completion

Every Functional Requirement is implemented and exercised by a passing test. The Behaviour Examples in this README hold under the documented inputs. Tests run with `-race` and produce no race report. No `time.Sleep` or wall-clock waiting exists in any test. The final slice is always equal in length to the input and always ordered by input index. A preflight violation (negative duration, duplicate name) returns the per-position rejected statuses without scheduling any goroutine. Cancellation invokes `Stop` at most once per still-pending timer, by the goroutine that owns the timer, and never blindly drains any timer channel. Already-committed completions remain completed even when cancellation arrives. Empty input returns an empty slice with zero goroutines launched. The seam between the core logic and real timers is exercised by tests, and the tests prove the stop-ownership invariant and the exact-once-by-the-owner invariant. You can answer every Self-Assessment Question without consulting the README.

## 19. Optional Extensions

Add a "reset" path for a timer that was started with a long duration and is later resized to a shorter one, surfacing the change as a distinct status while preserving the original input position. Add support for grouping timers by tag, where cancellation cancels only the timers in the named tag and the rest proceed to completion.
