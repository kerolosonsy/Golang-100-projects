# Project 036 — Rate Limiter Token Bucket

## 1. Project Name and Number

Project 036 — Rate Limiter Token Bucket. The project teaches a concurrency-safe token bucket whose time source can be controlled by tests.

## 2. Project Idea

Build a limiter with a positive whole-token capacity and a positive refill rate expressed as a positive whole-token count per positive duration. The bucket starts full. Each request asks to consume a positive whole-token amount. A request succeeds only when the bucket contains enough tokens at that instant; it then removes the requested amount as one atomic decision. A rejected request removes nothing.

Refill is lazy. An operation observes the injected monotonic time, credits elapsed time, and clamps the balance to capacity before deciding the request. Fractional credit remains stored with enough precision that many small elapsed intervals accumulate instead of being discarded. The core logic does not need a background ticker.

## 3. Why This Project Now?

Project 035 introduced concurrent work around an injected, controllable dependency. This project is the immediate next step: it applies the same deterministic testing discipline to elapsed time and adds shared mutable state protected by a mutex. It also prepares you for later middleware and cache projects that need explicit rate and lifecycle policies.

## 4. Prerequisites

Complete Projects 031 and 035 first. Project 031 supplies the barriers, wait groups, and cancellation patterns reused for deterministic tests. Project 035 supplies the deterministic-testing discipline and the injected-time dependency idea applied to elapsed-time accounting here. You should already understand structs, methods, errors, durations, interfaces or function values used as dependencies, and basic goroutines. Review the earlier channel and worker-pool projects when comparing mutex protection with channel ownership.

## 5. What You Must Know Before Starting

Know that a duration is an elapsed interval, not a wall-clock calendar value. Know that a monotonic reading must not move backward for refill accounting. Know how to represent a rate as a rational relationship between tokens and duration, how to preserve fractional credit, and how to prevent arithmetic overflow during elapsed-time calculations. Know that a mutex protects the complete read, refill, decision, and mutation sequence rather than only the final assignment. Know that tests can advance a fake clock directly and can coordinate callers with barriers or channels.

## 6. Explanation of New Concepts

A token bucket has two limits: capacity limits stored tokens, while refill rate limits how quickly credit returns. Starting full permits an initial burst up to capacity. A request consumes whole tokens, but elapsed refill may create a fractional remainder that is not spendable until enough credit reaches the next whole-token boundary.

The refill calculation uses elapsed monotonic time since the last accounted instant. Refill credit is defined as the exact rational total of elapsed nanoseconds multiplied by the configured whole-token rate and divided by the configured duration in nanoseconds. Only whole tokens from that total become spendable; the fractional remainder is retained between observations instead of being rounded away. When the balance reaches capacity, excess whole and fractional credit is discarded, because a full bucket must not bank credit above its burst limit. Checked arithmetic or an equivalent overflow-safe representation must preserve this meaning. Time at the same instant produces no credit. A reading earlier than the last accepted reading is clamped to the last accepted reading, produces no credit, and does not move the time marker backward.

Atomic admission means no caller can observe a balance check separated from its consumption. The mutex covers refill and request decision together. External I/O does not belong inside that critical section; this project has no need to perform I/O while deciding admission.

Lazy refill means operations perform accounting only when they arrive. A ticker goroutine would wake even when nobody requests admission, introduce a lifecycle and shutdown obligation, make fake-time tests harder to reason about, and create another shared communication path. It is unnecessary when elapsed time can be calculated at each operation.

## 7. Learning Objective

By completion, you can define precise token-bucket semantics, inject a monotonic clock, preserve fractional refill credit, reject invalid inputs without mutation, and protect admission decisions under concurrent calls. You can explain why lazy time accounting is simpler than a ticker and prove that accepted consumption never exceeds available credit.

## 8. Functional Requirements

1. Configuration requires a positive whole-token capacity and a positive refill rate represented as a positive whole-token count per positive duration.
2. The initial balance equals capacity, including any internal fractional representation used for later refill accounting.
3. Each admission request requires a positive whole-token amount; the API exposes no fractional request value.
4. Before every decision, elapsed time from the injected monotonic source is accounted for once.
5. Refill never makes the balance exceed capacity.
6. A request succeeds atomically only when enough whole tokens are available, then consumes exactly its requested amount.
7. A rejected request consumes no tokens and does not erase fractional credit.
8. The core logic receives an injected monotonic time source; production wiring may adapt the standard library clock, while tests advance fake time directly.
9. A time reading earlier than the last accepted reading is clamped to that last reading, creates no tokens, and never moves the marker backward.
10. Frozen time creates no refill.
11. All shared state is protected by a mutex. No lock is held while performing external I/O.
12. Invalid configuration and invalid request amounts return errors and leave limiter state unchanged.
13. Concurrent callers observe behavior consistent with serialized atomic decisions.

## 9. Inputs and Outputs

Inputs are the limiter configuration, an injected time source, and admission requests containing positive whole-token amounts. Time readings are consumed only as monotonic elapsed observations under the documented policy.

Outputs are an admission result and, where useful for teaching or diagnostics, an error identifying invalid configuration or invalid request. A successful result means exactly the requested number of tokens was consumed. A normal rejection means insufficient tokens and no mutation from that request. A backward time reading is clamped and is not an error.

Text-only example: capacity five starts with five tokens. A request for three succeeds and leaves two. A request for three immediately afterward is rejected and leaves two. After enough fake elapsed time for two and a half tokens, a request for four remains rejected, while the half-token remainder stays available for future refill.

## 10. Rules and Edge Cases

Capacity zero, negative capacity, zero or negative refill rate, and a non-positive duration are invalid. A request of zero or a negative amount is invalid rather than a rejection; a fractional request is not representable in the API. Invalid configuration or request input returns before changing balance or fractional credit. A backward time reading is valid input but is clamped to the last accepted time and produces no refill.

At exact capacity, elapsed refill is clamped and excess credit is not allowed to accumulate above the capacity ceiling. At an exact refill boundary, the newly earned whole tokens are available. Just before that boundary, the request must reflect the retained fractional remainder precisely. Repeated tiny intervals must equal one equivalent larger interval, subject to the chosen representation's documented arithmetic boundary.

A request requiring exactly the current available whole-token balance succeeds. A request requiring one more token fails without consumption. Backward time is clamped to the last accepted time, so it creates no credit and cannot cause a later interval to be counted twice. Concurrent admissions may succeed in any scheduling order, but their total accepted consumption cannot exceed initial credit plus valid elapsed refill, capped at capacity at each refill observation.

## 11. Project Constraints

Use only the Go standard library. Keep the core decision deterministic through injected time. Tests must advance fake time directly and must never sleep or depend on wall-clock scheduling. Use a mutex for shared state. Do not add a ticker goroutine, external I/O, network service, persistence, or third-party rate-limiting package. Do not expose the mutex or mutable internal state. The race detector must pass.

## 12. Design Questions Before Coding

What exact type and precision preserve the mathematically defined fractional refill credit without losing repeated small intervals? Which time reading is the last accepted point, and when is it advanced? How does clamping backward time prevent credit creation and double counting? Does an invalid request account for elapsed time, and can you state that choice consistently? How will exact-boundary behavior be distinguished from just-before-boundary behavior? Which state must the mutex protect so check and consume cannot separate? How will tests prove total admission under contention without relying on timing? Why would a ticker make ownership, cleanup, and fake-time control more complicated?

## 13. Implementation Milestones

1. Write the behavioral contract for valid configuration, initial fullness, request validation, refill, rejection, fractional remainder, and backward-time clamping.
2. Choose a representation that retains fractional refill credit and document its precision and boundary behavior.
3. Define the injected monotonic time dependency and fake-time test control.
4. Implement configuration validation and prove invalid configuration cannot create a usable limiter.
5. Implement lazy refill with capacity clamping and an explicit backward-time policy.
6. Implement atomic admission under one mutex-protected decision sequence.
7. Add deterministic tests for boundaries, fractional accumulation, frozen time, and invalid inputs.
8. Add concurrent admission tests that reconcile accepted tokens against available credit.
9. Run the complete package under the race detector and investigate every reported access.

## 14. Verification Cases the Learner Must Write

Test that a new limiter starts full and accepts exactly capacity whole tokens before rejecting the next request. Test rejection with no consumption by following a failed request with a request that would have fit before the failure. Test exact capacity, exact refill boundaries, just-before-boundary fractional credit, partial refill, full refill, and capacity clamping. Test many small fake-time advances against one equivalent elapsed interval and verify retained fractional credit.

Test zero, negative, and otherwise invalid capacity or refill configurations. Test zero and negative request amounts. Test frozen time, a backward reading under the documented clamping policy, and a later forward reading to show backward time never creates or double-counts tokens. Test concurrent callers with barriers or channels, then verify accepted consumption never exceeds initial credit plus permitted refill. Run all tests with the race detector. No test may use sleep to make time pass or to coordinate goroutines.

## 15. Common Mistakes to Watch For

Discarding fractional tokens on every refill loses credit. Refilling only when a whole token is available creates boundary drift. Using wall-clock jumps as elapsed credit can mint tokens after time moves backward. Checking balance outside the mutex creates a check-then-act race. Updating last-observed time inconsistently can count one interval twice or not at all. Allowing refill to exceed capacity breaks burst semantics. Treating an invalid request as a normal insufficient-token rejection hides caller errors. Adding a ticker creates cleanup and test-control problems without improving lazy behavior. Holding a lock around logging, callbacks, or other external I/O expands contention and risks deadlock.

## 16. Topics and References for Study

Study the standard library documentation for `time.Duration`, `time.Time` monotonic behavior, `sync.Mutex`, errors, and integer arithmetic. Review Project 035's injected dependencies and deterministic tests. Read about token buckets, burst capacity, fixed-point or rational accumulation, linearizability, and check-then-act races. Study the race detector documentation and the testing package's parallel execution behavior.

## 17. Self-Assessment Questions

Why does the bucket start full, and what burst does that permit? Why are the configured capacity and refill rate whole-token integers while refill credit may be fractional until it reaches a whole-token boundary? How can repeated small elapsed intervals avoid losing credit? What exactly happens at capacity after a long idle period? Which operation sequence must be atomic? What is your backward-time policy, and why can it not create tokens? Why is a ticker unnecessary? What credit bound should hold across concurrent admissions? Which state is hidden behind the mutex, and why must callers not receive it directly?

## 18. Definition of Completion

The limiter enforces every functional requirement, starts full, preserves mathematically defined fractional refill credit, clamps at capacity, and rejects invalid inputs without unintended mutation. Tests deterministically cover all listed refill, boundary, invalid-time, and concurrency cases without sleep. Backward time is clamped to the last accepted reading and never creates tokens. Concurrent accepted consumption respects available credit. The full package passes the race detector. The README contains no promise of ticker-driven behavior and no external dependency.

## 19. Optional Extensions

Add a deterministic admission report that records accepted amount, available whole tokens, and retained fractional credit without exposing mutable internal state. Add a configurable policy for reporting retry-after duration on rejection, with fake-time tests for exact boundary rounding.
