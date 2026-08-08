# Project 042 — Pub-Sub Event Bus

## 1. Project Name and Number

- Project 042 — Pub-Sub Event Bus.
- The project is a learning lab for an in-process topic event bus that delivers events synchronously to subscribers, exposes subscribe and unsubscribe as stable handles, and remains correct under concurrent publishers and subscribers without hidden locking at the callback boundary.

## 2. Project Idea

The bus keeps an internal list of subscriptions. A subscription pairs a topic with a callback. Subscribe returns a stable identifier the caller uses to unsubscribe. Publish takes a snapshot of the matching subscriptions at the moment of publishing and invokes each callback in subscription order without holding the bus mutex, so callbacks may subscribe, unsubscribe, or reenter the bus without deadlock. Delivery is synchronous, so a publisher blocks while callbacks run. The bus does not queue events for delivery later, does not retain events for late subscribers, does not match wildcards, and does not transmit events across a network.

## 3. Why This Project Now?

- Project 041 was the immediate predecessor and built a single cooperative operation that obeys context.
- Project 042 raises the level: instead of one operation, many subscribers want to know about many events.
- Project 041 taught honesty about cancellation and error identity; Project 042 teaches honesty about delivery, ordering, and isolation when one callback fails.
- The goroutine and barrier thinking from Project 031 is reused here for the concurrent test suite, and the mutex vocabulary from earlier projects is reused to keep the bus safe under concurrent use.
- Project 044 will use the channel-ownership thinking trained here when it decides who closes a channel.

## 4. Prerequisites

- Complete Projects 031 and 041 first.
- Project 031 introduces goroutines, channels, and barriers reused for the concurrent test suite.
- Project 041 supplies the error-identity and closure thinking the bus needs when a callback returns an error or panics.
- You must already be comfortable with goroutines, channels, mutexes, error sentinels, panic recovery, and barrier-driven tests with no sleep.
- Earlier projects that introduce mutex ownership are useful review but are not required prerequisites.

## 5. What You Must Know Before Starting

- Know that synchronization primitives must never be held while a user-supplied callback runs, because the callback may want to subscribe, unsubscribe, or publish through the same bus, and the bus must not deadlock against its own caller.
- The standard solution is to take a snapshot of the matching subscriptions under the lock and to deliver outside the lock.

- Know that an in-process bus cannot reason across processes, so every behavior described here is single-process only.
- Persistence, durable delivery, retries, backpressure, and cross-process transport are explicitly out of scope and must not be added.

- Know that a panic is a Go language feature separate from error returns.
- A panic inside a callback must be recovered at the bus boundary so other subscribers still receive the same publish, the publisher observes a delivery error rather than a crash, and the runtime continues executing.

- Know that duplicate callbacks are separate subscriptions.
- A callback registered twice is a separate subscriber and must be invoked twice for each matching publish.
- Removing one does not affect the other.

- Know that a stable subscription identifier is the bus's contract with its caller.
- The identifier is opaque to the caller but must remain unique across the lifetime of the bus and must be the only token the caller uses to remove that subscription later.

## 6. Explanation of New Concepts

### Concepts

- A subscription bundles three things: a topic the bus matches against published events, a callback the bus invokes synchronously when a matching event arrives, and an identifier the bus returns so the caller can remove the subscription.
- The bus stores subscriptions in an internal list ordered by registration time so that the subscription order is the delivery order.

- A publish snapshot is the list of matching subscriptions captured at the moment of publishing.
- The bus takes this snapshot while holding the bus mutex, releases the mutex, and iterates the snapshot outside the lock to invoke each callback.
- New subscriptions created during a publish, and existing subscriptions unsubscribed during a publish, only affect future publishes; the in-flight publish sees only what was true when it began.

- Subscription order is determined by the order in which subscriptions were registered with the bus.
- The publish order of two events is determined by the order in which the publisher invoked publish.
- The delivery order of callbacks inside one publish call is determined by subscription order, and that ordering is sequential within one publish.
- Concurrent publish calls have no global order across the bus and may invoke the same callback concurrently; the bus does not promise that distinct publish calls never interleave at the call-stack level.
- A callback may recursively publish through the same bus; that nested publish takes its own snapshot and only sees subscription changes made before it took its own snapshot.

- A delivery record is the per-subscription result of one publish.
- Each matching subscription produces exactly one delivery record in subscription order, and that record carries the stable subscription identifier, an outcome flag, and a non-nil error when the callback returned an error or the recovered panic was treated as a delivery error.
- Records are collected in subscription order so the publisher can see which subscriber failed and which subscribers after it still ran.
- A panic is recovered at the boundary, recorded as a non-nil delivery error associated with that subscription, and later snapshot callbacks still run.

- A subscribe call returns a stable, unique identifier.
- The identifier is never reused during the bus's lifetime.
- If the identifier space is exhausted, subscribe rejects rather than wrapping or reusing an identifier.
- An unsubscribe call uses the identifier; removing an existing subscription returns true and changes the bus state, while removing an unknown identifier returns false and changes nothing.
- The policy is fixed and consistent across every unsubscribe path.

## 7. Learning Objective

- By completion, you can design an in-process pub-sub bus where publish takes a snapshot, delivers outside the lock, and lets callbacks subscribe, unsubscribe, or reenter the bus without deadlock.
- You can treat a panic as a recoverable delivery error rather than a system crash.
- You can keep identity stable and non-reusable for each subscription and reason about what subscribe and unsubscribe are allowed to do.
- You can write concurrent tests that drive many publishers and many subscribers without sleep and pass the race detector.

## 8. Functional Requirements

1. The bus is in-process and synchronous; no asynchronous queue, no persistence, no retry, no networking, no wildcard topics.
2. A subscription carries a topic string and a callback.
3. Subscribe returns a stable subscription identifier used to remove the subscription. The identifier is never reused during the bus's lifetime. If the identifier space is exhausted, subscribe rejects rather than reusing an identifier.
4. Unsubscribe with an existing subscription identifier returns true and removes that subscription. Unsubscribe with an unknown identifier returns false and changes nothing. The policy is fixed and consistent across every unsubscribe path.
5. Publish takes a snapshot of matching subscriptions and invokes them outside the bus mutex in subscription order.
6. Each matching subscription produces exactly one delivery record in subscription order, carrying the stable subscription identifier and the success or error outcome.
7. Within one publish call, callbacks run sequentially in subscription order. Concurrent publish calls have no global order and may invoke the same callback concurrently.
8. Subscription changes during a publish affect only later publishes; the in-flight publish sees only the snapshot taken at its start. A callback that calls publish recursively takes its own snapshot and only sees subscription changes made before it took its own snapshot.
9. Subscribe, unsubscribe, and publish are safe to call concurrently from many goroutines.
10. An empty topic is rejected by both subscribe and publish.
11. A nil callback is rejected at subscribe time.
12. A nil payload is allowed and is delivered to callbacks; the bus does not reject nil payloads.
13. A callback that returns an error contributes that error to the publisher's delivery record in subscription order; earlier callback errors do not skip later callbacks.
14. A callback that panics is recovered at the bus boundary, recorded as a non-nil delivery error associated with that subscription, and the panic does not stop later snapshot callbacks on the same publish.
15. Duplicate callbacks are separate subscriptions and are invoked once per subscription per matching publish.
16. The event or payload value is passed by value. The bus does not deep-copy payloads. Pointers, slices, maps, interfaces, functions, and other reference-containing payloads still refer to caller-owned mutable data, and the bus documents this honestly.
17. Subscribe, unsubscribe, and publish return the values expected by the documentation without panicking in normal operation.
18. Tests cover none, one, and many topics; subscription order; duplicate subscriptions; unsubscribe; changes during callback execution; callback errors; panic isolation; concurrent publishers and subscribers; and race-detector passes without sleep.

## 9. Inputs and Outputs

### Interface Contract

- The input is a topic string and a payload value supplied to publish.
- Subscribe takes a topic and a callback.
- Unsubscribe takes a subscription identifier.

- The output of subscribe is a stable subscription identifier, or a clear rejection when input is invalid or the identifier space is exhausted.
- The output of unsubscribe is true when an existing subscription was removed and false when the identifier was unknown.
- The output of publish is a list of delivery records in subscription order, one per matching subscription, each carrying the subscription identifier and the success or error outcome for that subscription.

- A side observable is the count of times the bus mutex would have been held during a callback, which tests assert is zero so callbacks are genuinely free to call back into the bus.

- Text-only example: subscribing callback A then callback B under topic T, publishing one event with payload P, yields exactly two delivery records in that order, the first carrying subscription A's identifier and the second carrying subscription B's identifier, both observed without subscription order depending on hashing or timing.

## 10. Rules and Edge Cases

- Subscribing an empty topic is rejected at subscribe time and the bus does not store anything.
- Publishing an empty topic is rejected and no callback is invoked.
- Subscribing a nil callback is rejected at subscribe time and the bus does not store anything.
- The bus does not reject a nil payload; a nil payload is delivered to callbacks as a value.

- Unsubscribe with an existing identifier returns true and removes that subscription.
- Unsubscribe with an unknown identifier returns false and changes nothing.
- The bus never panics on an unknown unsubscribe.

- A callback that subscribes during delivery is recorded as a new subscription that begins only after the in-flight publish completes.
- A callback that unsubscribes itself during delivery removes its own subscription; the in-flight publish continues invoking later callbacks that were already in the snapshot.
- A callback that publishes during delivery starts a new publish call that takes its own snapshot, runs sequentially in subscription order within that call, and may interleave at the call-stack level with the outer publish.

- A callback that returns an error contributes that error to the matching delivery record and the next callback in subscription order still runs.
- A callback that panics is recovered, contributes a non-nil delivery error for that subscription, and the next callback in subscription order still runs.
- A bus must not let a panic escape into the publisher's goroutine.

- The bus must not hold its lock while running a user callback.
- It must not deliver asynchronously.
- It must not match wildcards.
- The payload value is passed by value to each callback; the bus does not deep-copy the payload and does not pretend that reference-containing payloads are isolated from the caller.

- Two publish calls running concurrently have no global order and may invoke the same callback concurrently.
- Tests must not assume one event ordering across many concurrent publishers; each publisher's events are delivered in the order that publisher called publish.
- Tests must not assume which goroutine wakes first.

## 11. Project Constraints

- Use only the Go standard library.
- Use the sync package and the sync/atomic package only when a single primitive is the natural choice; channels and mutexes are the expected primitives here.
- Do not introduce queues, timers, persistence, or wildcards.
- Do not introduce sleep into tests.
- Do not introduce networking.
- Do not deep-copy payloads.
- The completed code must pass the race detector.

## 12. Design Questions Before Coding

- What does subscribe return when the topic is empty or the callback is nil, and is that policy consistent with how publish rejects an empty topic?
- How is the snapshot taken under the mutex and released before any callback runs?
- How are delivery records collected in subscription order while callbacks may themselves mutate the bus or recursively publish?
- How is panic recovery scoped so the publisher observes a delivery error rather than a crash?
- How are subscription identifiers made stable, never reused, and exhausted gracefully?
- What is the fixed behavior of unsubscribe with an unknown identifier, and is the documentation explicit enough that callers can rely on it?
- How is duplicate-callback behavior explained so callers do not assume a deduplicated subscription?
- Where does the bus document that payload values are passed by value but reference-containing payloads still refer to caller-owned mutable data?

## 13. Implementation Milestones

1. Define the subscription type, the identifier type, and the bus type.
2. Implement subscribe with input validation for empty topic and nil callback, returning a stable unique identifier that is never reused.
3. Implement unsubscribe that returns true when an existing subscription was removed and false when the identifier was unknown, with the same policy on every code path.
4. Implement publish that takes a snapshot under the mutex and delivers outside the lock in subscription order, rejecting an empty topic.
5. Implement panic recovery at the callback boundary so other subscribers continue and the publisher observes a non-nil delivery error for the panicking subscription.
6. Implement delivery records in subscription order so callers can inspect who failed, each record carrying the stable subscription identifier and the success or error outcome.
7. Document the payload-by-value rule: the bus does not deep-copy the payload, and reference-containing payloads still refer to caller-owned mutable data.
8. Add tests covering none, one, and many topics; subscription order; duplicates; unsubscribe; subscription changes during callback execution; recursive publish; callback errors; panic isolation; concurrent publishers and subscribers; and assertion that the mutex is not held during callback execution.
9. Run the full package under the race detector and correct every issue reported.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Test the no-subscriber path: a publish with no matching subscriptions returns an empty delivery list and no panic.
- Test one-subscriber path: a publish invokes the single matching callback once with the supplied payload and produces one delivery record carrying the subscription identifier.
- Test many-subscriber path: many callbacks subscribe to the same topic and a single publish invokes them in subscription order with the same payload, producing one delivery record per matching subscription in that order.
- Test subscription order: register callbacks in a known order, publish once, and observe the delivery record sequence in that order.

- Test unsubscribe removes: subscribe, receive identifier, unsubscribe with the same identifier, publish, observe zero callbacks.
- Test unsubscribe unknown: unsubscribe with an identifier never issued returns false and does not panic.
- Test duplicate callbacks: register the same callback twice and observe that one publish invokes it twice and produces two records with distinct identifiers.
- Test changes during callback: a callback that subscribes a new callback or unsubscribes itself must not affect the in-flight publish, and a later publish must see those changes.

- Test empty-topic rejection: subscribe rejects an empty topic; publish rejects an empty topic and does not invoke any callback.
- Test nil-callback rejection: subscribe rejects a nil callback.
- Test nil-payload delivery: a nil payload is delivered to callbacks as a value, the bus does not reject it, and the published record reflects the nil payload.

- Test callback error: a callback that returns an error contributes the error to the publisher's delivery record and later callbacks still run.
- Test panic isolation: a callback that panics is recovered, the publisher observes a non-nil delivery error for that subscription, and later callbacks still run.
- Test recursive publish: a callback that calls publish through the same bus takes its own snapshot, sees only subscription changes made before its snapshot, and the outer publish sees only the subscription state at its own snapshot.

- Test concurrent operations: many goroutines subscribe, unsubscribe, and publish simultaneously; the bus must not deadlock, must not race, and must pass the race detector.
- Test that the bus mutex is not held while callbacks run.
- One way is to have a callback that calls subscribe or unsubscribe and observe no deadlock, which proves the lock is not held.
- Do not introduce sleep into the tests.

## 15. Common Mistakes to Watch For

- Holding the bus mutex while running user callbacks deadlocks any callback that subscribes, unsubscribes, or publishes through the same bus.
- Forgetting to recover panics lets one faulty subscriber crash the publisher's goroutine.
- Treating duplicate callbacks as one subscription silently drops half of the publish's work.
- Letting a subscription change during a publish take effect immediately produces nondeterministic delivery that tests cannot reliably write.
- Panicking on an unknown unsubscribe breaks callers that retry cleanup.
- Using real timers to give callbacks a chance to run before asserting on results makes tests flaky.
- Returning errors through a different mechanism than delivery records hides which subscribers failed.
- Deep-copying payloads in the name of safety hides the contract and silently breaks callers that expect reference equality.
- Promising that separate publishes never interleave is a false global-order claim that tests cannot rely on.

## 16. Topics and References for Study

- Study the standard library documentation for the sync package, in particular the mutex and read-write mutex.
- Study the documentation for the recover built-in and how to scope panic recovery inside a function.
- Read Go's standard explanations of subscriber, observer, and listener patterns.
- Compare this project with Project 031's barriers, the mutex discipline from earlier projects, and Project 041's honest error identity.
- Project 044's pipeline decides who closes channels, which is exactly the kind of ownership thinking trained here.

## 17. Self-Assessment Questions

1. Why does publish take a snapshot under the mutex and deliver outside it?
2. How does the bus allow a callback to call subscribe, unsubscribe, or publish without deadlocking?
3. Why are duplicate callbacks separate subscriptions?
4. How does the bus recover a panicking callback and continue delivering?
5. Why must subscribe return a stable identifier that is never reused during the bus's lifetime, and what does the bus do when the identifier space is exhausted?
6. What is the fixed behavior of unsubscribe with an unknown identifier, and why is documentation alone not enough?
7. How are delivery records collected in subscription order, and what does that let the publisher see?
8. Why is it incorrect to claim that separate publish calls never interleave?
9. Why is the payload-by-value rule honest about reference-containing payloads?

## 18. Definition of Completion

- [ ] The bus exposes subscribe, unsubscribe, and publish as its entire surface.
- [ ] Subscribe rejects empty topics and nil callbacks, returns a stable unique identifier that is never reused, and rejects when the identifier space is exhausted.
- [ ] Publish rejects empty topics, takes a snapshot under the bus mutex, and delivers outside it in subscription order, producing one delivery record per matching subscription in that order.
- [ ] Each delivery record carries the matching subscription identifier and the success or error outcome for that subscription.
- [ ] Unsubscribe with an existing identifier returns true and removes the subscription.
- [ ] Unsubscribe with an unknown identifier returns false and changes nothing.
- [ ] The policy is fixed across every unsubscribe path and the bus never panics on an unknown unsubscribe.
- [ ] Subscription changes during a publish affect only later publishes.
- [ ] A callback that recursively publishes takes its own snapshot and only sees subscription changes made before it took its own snapshot.
- [ ] Concurrent subscribe, unsubscribe, and publish from many goroutines never deadlock, never race, and pass the race detector.
- [ ] Tests cover none, one, and many topics; subscription order; duplicates; unsubscribe; changes during callback; recursive publish; callback errors; panic isolation; concurrent operations; and an assertion that the mutex is not held during callback execution.
- [ ] Tests do not use sleep.

## 19. Optional Extensions

- Add a small teaching note that catalogs, in plain prose, every place the bus touches shared state and explains why the snapshot pattern keeps each one correct.
- Add a property-style check that randomizes many subscribe, unsubscribe, and publish operations and confirms the bus never returns a subscription identifier it issued twice.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 041 — Context Timeout Example](../../03-concurrency/041_context_timeout_example/README.md#20-prerequisite-based-documentation-guide), [Project 031 — Concurrent Timer](../../03-concurrency/031_concurrent_timer/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`runtime/debug`](https://pkg.go.dev/runtime/debug).
- **Standards and concept references:** [Go specification: handling panics](https://go.dev/ref/spec#Handling_panics).

### Project-specific learning focus

- **Learn now:** snapshot-under-lock delivery, subscriber identity, unsubscribe semantics, panic isolation, reentrancy, lock scope, and randomized invariant checks.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
