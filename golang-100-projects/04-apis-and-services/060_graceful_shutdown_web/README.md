# Project 060 — Graceful Shutdown Web

## 1. Project Name and Number

Project 060 — `graceful_shutdown_web`. Folder: `04-apis-and-services/060_graceful_shutdown_web/`. README only; the learner writes all source and tests.

## 2. Project Idea

Compose an HTTP server with explicit timeouts and one or more dependencies that must each close exactly once. The server values are pinned: `ReadHeaderTimeout` 2 seconds, `ReadTimeout` 10 seconds, `WriteTimeout` 10 seconds, `IdleTimeout` 30 seconds, and a graceful shutdown budget of 5 seconds. There is no separate configurable force-close deadline; when the graceful context ends, the orchestration calls `Server.Close` immediately. The production entry point converts `SIGINT` and `SIGTERM` into a lifecycle trigger. Tests never send OS signals; the orchestration accepts an injected listener and a trigger abstraction whose first call activates one shutdown notification and whose later calls are idempotently coalesced so that no caller blocks on a second trigger. The shutdown context is freshly derived from `context.Background()` at the moment of shutdown, never from the cancelled parent context and never from a request context. Each dependency is closed exactly once after HTTP serving has quiesced or force-close has completed. `http.ErrServerClosed` is treated as normal. `Run` returns one `errors.Join` aggregate. `Run` never calls `os.Exit`. Tests use real ephemeral listeners only where strictly necessary plus barrier channels inside handlers and a listener-close observer, and never use fixed ports, real sleeps, or real signals.

## 3. Why This Project Now?

Projects 046 through 059 produced an HTTP service with growing cross-cutting concerns: routing, middleware, JSON envelopes, auth, rate limiting, sessions, CSRF. None of those projects dealt with what happens when the process must stop. Project 060 introduces the discipline of treating the server lifecycle as a state machine with explicit transitions, of composing shutdown with dependency cleanup, and of writing tests that exercise the orchestration deterministically. By the end of Project 060 the learner can ship a service whose stop behaviour is provably correct.

## 4. Prerequisites

Required earlier projects: Project 059, Project 046, and Project 041. Earlier HTTP, middleware, and concurrency projects are useful review but are not formally required. The learner must already understand `net/http` server timeouts (`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`), `Server.ListenAndServe`, `Server.Shutdown`, `Server.Close`, `Server.Serve`, the meaning of `http.ErrServerClosed`, `context.Context` cancellation and deadlines, and how to use `httptest` with a custom listener.

## 5. What You Must Know Before Starting

- The difference between `Server.Shutdown` and `Server.Close`. Shutdown stops accepting new connections and waits for in-flight handlers to return. Close force-closes all connections immediately.
- That `Server.Shutdown` causes `Serve` to return `http.ErrServerClosed`. Any other error from `Serve` is a startup or runtime error.
- That `http.Server` is a struct. Timeouts are fields. The learner must set the four pinned values exactly.
- That `context.Context` values are not stored in structs. A shutdown context must be created fresh inside the orchestration.
- That `net.Listener` is an interface. `net.Listen("tcp", "127.0.0.1:0")` returns an ephemeral listener; the port is `0` and the kernel picks a free port.
- That `signal.Notify` is the production mechanism for receiving `SIGINT` and `SIGTERM`. Tests must not depend on signals.
- That `httptest.NewServer` allocates a port and tears down on `Close`, but does not exercise `Server.Shutdown` semantics. The learner must drive shutdown through the orchestration for the shutdown tests.
- That `errors.Join` aggregates multiple errors into one value.

## 6. Explanation of New Concepts

**Server timeouts.** `ReadHeaderTimeout` caps how long the server waits for request headers. `ReadTimeout` caps the total request read time. `WriteTimeout` caps the response write time. `IdleTimeout` caps keep-alive idle time. Without these, a slow or hostile client can hold a connection open indefinitely and exhaust resources. The configuration pins each one to a finite value.

**Shutdown context.** `Server.Shutdown` takes a `context.Context`. The orchestration creates a fresh context at the moment shutdown begins, derived from `context.Background()`, with a finite deadline equal to the configured graceful shutdown budget of 5 seconds. The orchestration does not derive the shutdown context from the parent `ctx`, because the parent may already be cancelled; nor does it derive it from any request context.

**Force-close.** When the graceful context expires, in-flight handlers may still be running. The orchestration calls `Server.Close` immediately. There is no separate configurable "force-close deadline". The contract is "after the graceful context ends, call `Server.Close` immediately".

**Trigger abstraction.** The orchestration exposes a single trigger abstraction. The first call activates one shutdown notification. Later calls are idempotently coalesced: they return without blocking, and they do not start a second shutdown. Production signals activate this abstraction; parent-context cancellation is the other documented event that can begin the same shutdown sequence.

**First-to-fire orchestration.** `Run` reacts to whichever happens first: an unexpected `Serve` failure, the parent context cancellation, or the first trigger. The parent context cancellation is treated as a trigger only; the shutdown context is still freshly derived from `context.Background()` at that moment.

**Dependency close order.** Each dependency that holds a resource (a database connection pool, a queue, an in-memory session store from Project 059) must be closed exactly once. Dependencies are closed only after HTTP serving has quiesced or force-close has completed. Closing a dependency before serving stops is a bug: in-flight handlers may still need it. Closing a dependency twice is a bug: many resources refuse double-close. The orchestration closes dependencies in declared order and continues closing remaining dependencies after an error, recording the error in the aggregate.

**`http.ErrServerClosed`.** When the server is shut down through `Server.Shutdown`, the `Serve` call returns `http.ErrServerClosed`. The orchestration treats this as the normal shutdown path and does not log it as an error. Any other error from `Serve` is logged and surfaced through the aggregate.

**Forced path return value.** When the graceful context ends before all in-flight handlers return, the orchestration calls `Server.Close`. The graceful context's error is `context.DeadlineExceeded`. The orchestration must surface that error through the aggregate; it must not silently return `nil`.

**Unexpected `Serve` failure before any trigger.** If `Serve` returns a non-`http.ErrServerClosed` error before any trigger fires, the orchestration does not wait forever for a trigger. It stops or closes the server as needed, closes dependencies exactly once, and returns the `Serve` error through the aggregate.

**Process exit discipline.** The process must wait for both the serve goroutine and the shutdown coordination to finish before exiting. A premature `os.Exit` skips the dependency cleanup and is forbidden inside the orchestration. `main` may exit after `Run` returns, but `Run` itself never exits the process.

**Test orchestration.** Tests do not use `signal.Notify`. They use the injected trigger and an injected listener so the test can drive every transition without races, without sleeps, and without fixed ports. Where a real `Listener` is required to exercise `Server.Shutdown` semantics, the test uses an ephemeral port on loopback and wraps the listener with a test observer whose `Close` event releases a barrier.

**Shutdown-context factory.** The orchestration accepts an injected shutdown-context factory. The production factory returns a fresh context derived from `context.Background()` with a 5-second timeout. The test factory returns a controllable context that the test can end after `Server.Shutdown` has started. The factory is the only place the shutdown context is created. The factory never weakens the production freshness rule.

**Barrier channels in handlers.** Handlers used by shutdown tests coordinate with the test through barrier channels. The handler signals when it has started and waits on a release channel before returning. The test signals start, calls the shutdown trigger, and then signals release. This proves that the in-flight handler completes before shutdown reports done. Force-path handlers observe context cancellation or have a release path so tests do not leak.

**Listener-close observer.** For new-request rejection tests, the test wraps the ephemeral listener with a small observer whose `Close` method signals a barrier. After the trigger, the test waits for the listener-closed barrier before attempting the new request. The test does not rely on timing.

**Goroutine discipline.** One serve goroutine is allowed and must be joined. No other background goroutine, ticker, or `time.AfterFunc` is owned by the orchestration. The lifecycle-owned goroutine and the lifecycle-owned completion channels are joined before `Run` returns. The tests assert this directly. `runtime.NumGoroutine` is not used as the primary leak assertion because it is noisy; lifecycle completion is the assertion.

## 7. Learning Objective

After finishing this project, the learner can explain why each `http.Server` timeout is set to the pinned value, why the shutdown context must be fresh from `context.Background()`, why dependencies are closed only after serving has quiesced, why each dependency is closed exactly once, why `http.ErrServerClosed` is not an error, why repeated shutdown triggers must be coalesced, and why `Run` never calls `os.Exit`. The learner can also write a test suite that proves every transition deterministically using ephemeral listeners and barrier channels.

## 8. Functional Requirements

1. The configuration pins `ReadHeaderTimeout` 2 seconds, `ReadTimeout` 10 seconds, `WriteTimeout` 10 seconds, `IdleTimeout` 30 seconds, and a graceful shutdown budget of 5 seconds. The constructor rejects non-positive values for any of these.
2. The configuration declares the dependencies to close in order, each implementing a single `Close` method.
3. The orchestration exposes a single entry point that accepts, in prose: the parent context, an injected listener, an injected trigger abstraction, an injected shutdown-context factory, and the dependency list. Production wraps this entry point with `signal.Notify` on `SIGINT` and `SIGTERM`. Tests call the entry point directly. The orchestration is the only place that knows the orchestration sequence; tests never reach into it through Go reflection or by reading internal state.
4. The trigger abstraction accepts one activation: the first call activates the shutdown notification; subsequent calls return immediately without blocking and do not start a second shutdown.
5. `Run` reacts to whichever happens first: an unexpected `Serve` failure, the parent context cancellation, or the first trigger. The parent context cancellation is treated as a trigger only. The shutdown context is freshly derived from `context.Background()` at the moment of shutdown, never from the parent context and never from a request context.
6. On shutdown the orchestration marks shutdown started, calls `Server.Shutdown` with the fresh finite context, and allows in-flight work to finish. If `Server.Shutdown` returns because its context ended, the orchestration calls `Server.Close` immediately.
7. The orchestration waits for `Serve` to return. Only then does it close dependencies, each exactly once, in declared order, continuing after errors and recording each error in the aggregate.
8. The orchestration returns one `errors.Join` aggregate containing unexpected `Serve` errors, the graceful context error when the graceful path timed out, unexpected `Close` errors, and dependency errors. Only `http.ErrServerClosed` is treated as normal and is not included in the aggregate.
9. A forced path returns the graceful context error through the aggregate. The orchestration does not return `nil` when the graceful context ended.
10. If `Serve` returns a non-`http.ErrServerClosed` error before any trigger fires, the orchestration stops or closes the server as needed, closes dependencies exactly once, and returns the `Serve` error through the aggregate. It does not wait forever for a trigger.
11. `main` waits for `Run` and reports its returned error normally. `Run` itself never calls `os.Exit`.
12. The HTTP server is configured with the four pinned timeouts. The configuration values are documented.
13. One serve goroutine is owned by the orchestration. The orchestration joins that goroutine before returning.
14. The orchestration starts no other background goroutine, ticker, or `time.AfterFunc`.
15. The lifecycle-owned completion channels and goroutines are joined before `Run` returns. The tests assert this directly; they do not rely on `runtime.NumGoroutine`.
16. Tests use the injected trigger and an ephemeral loopback listener. Real signals are never sent. Fixed ports are never used. `time.Sleep` is never used.

## 9. Inputs and Outputs

Inputs: the parent `ctx`, the `net.Listener`, the trigger abstraction, the injected shutdown-context factory, the configured timeouts, and the dependency list. Outputs: a returned error from `Run` (or nil on a clean shutdown), and the documented side effects on the listener, the server, and the dependencies. Example textual inputs and expected textual outputs:

- Trigger fires once; all in-flight handlers complete within the 5-second graceful budget. Expected: `Run` returns nil. All dependencies are closed exactly once. No error logged.
- Trigger fires once; one in-flight handler does not complete within the 5-second graceful budget. Expected: `Run` returns an `errors.Join` aggregate that contains the graceful context error (`context.DeadlineExceeded`). `Server.Close` was called. The handler's connection is force-closed. All dependencies are closed exactly once.
- Trigger fires twice in quick succession. Expected: only the first trigger starts the shutdown. The second trigger returns immediately without blocking. The orchestration does not panic and does not start a second shutdown.
- Parent context is cancelled before any trigger fires. Expected: parent cancellation is treated as a trigger. The shutdown context is still freshly derived from `context.Background()` with the full 5-second deadline. Dependencies are closed exactly once.
- `Server.Serve` returns a non-`http.ErrServerClosed` error. Expected: the error is included in the returned aggregate. Dependencies are still closed exactly once.
- A dependency's `Close` returns an error. Expected: the error is included in the aggregate. The remaining dependencies are still closed. `Run` returns a non-nil aggregate.
- Trigger fires before `Serve` is ready. Expected: the trigger call returns without blocking. Once `Serve` is running, the shutdown proceeds normally.
- Trigger fires after `Run` has returned. Expected: the trigger call returns without blocking. The orchestration does not panic.

## 10. Rules and Edge Cases

- A trigger delivered before `Server.Serve` has returned is queued or absorbed. The implementation detail (buffered channel, atomic flag) is pinned and tested.
- A trigger delivered after `Run` has returned is a no-op and does not block.
- The shutdown context is independent of the parent `ctx`. If the parent `ctx` is already cancelled, the shutdown context still has its full 5-second deadline.
- The clean shutdown path never calls `Server.Close`. The orchestration calls `Server.Shutdown`, observes its return, waits for `Serve` to return, and proceeds to dependency close. `Server.Close` is invoked only when the graceful context ends before all in-flight handlers return, or when an unexpected `Serve` failure requires the listener to be torn down.
- A dependency that closes twice panics in some libraries. The orchestration guards against double-close.
- A handler that blocks past the graceful deadline is force-closed by `Server.Close`; its response body may be truncated. The orchestration does not wait for the handler to return after `Server.Close`.
- The listener is closed by `Server.Shutdown` or `Server.Close`. The orchestration does not close the listener separately.
- The orchestration does not start any goroutine other than the one for `Server.Serve`.
- The orchestration does not read or write to standard streams directly. Logging is delegated to a documented logger interface so tests can capture or suppress output.
- A forced path returns the graceful context error in the aggregate, not `nil`.

## 11. Project Constraints

- Standard library only. No third-party libraries.
- No fixed ports in tests. Tests use ephemeral loopback listeners.
- No real signals in tests. The trigger is injected.
- No `time.Sleep` in tests. The orchestration and the handler use barrier channels.
- No `os.Exit` inside the orchestration. The application may exit after `Run` returns.
- `runtime.NumGoroutine` is not the primary leak assertion. Lifecycle completion is.

## 12. Design Questions Before Coding

- How is the trigger represented? A small abstraction whose first call activates one notification and whose later calls are idempotently coalesced.
- How is "first trigger wins" implemented? An `atomic.Bool` flag or a buffered channel of size one. The choice is pinned and tested.
- How is the dependency list represented? A slice of a small `Closer` interface, iterated in declared order.
- How is the error aggregation implemented? `errors.Join` from the standard library.
- How is the handler barrier represented? A `started` channel that the handler closes on entry and a `release` channel that the handler blocks on.
- How is the listener-close observer implemented? A small wrapper that delegates to the underlying `Listener` and signals a barrier from `Close`.
- How is the shutdown-context factory represented? A function that returns a fresh context and a cancel function. Production returns a fresh `context.WithTimeout(context.Background(), 5*time.Second)`. Tests return a controllable context the test can end after the shutdown-start barrier.
- How is the production `main` organised? A thin `main` that builds the configuration, calls `signal.Notify`, calls the orchestration, and exits with a documented status code. Tests do not import `main`.

## 13. Implementation Milestones

1. Sketch the configuration, the trigger abstraction, the dependency list, the shutdown-context factory, and the orchestration state on paper.
2. Implement the configuration struct with documented defaults, the pinned timeouts, and constructor validation that rejects non-positive values.
3. Implement the dependency interface and the dependency-close loop with error aggregation using `errors.Join`.
4. Implement the trigger abstraction so that the first call activates one notification and later calls are idempotently coalesced.
5. Implement the shutdown-context factory so that the production version returns a fresh `context.Background()`-derived 5-second timeout context and the test version returns a controllable context.
6. Implement the orchestration's main loop: `Server.Serve`, select on `Serve` completion, parent cancellation, and the trigger; call `Server.Shutdown` with the fresh context; force-close on graceful context expiry; close dependencies; wait for the serve goroutine.
7. Implement the production `main` that wraps the orchestration with `signal.Notify` on `SIGINT` and `SIGTERM`. Tests must not depend on `main`.
8. Implement the test handler with barrier channels and a force-path handler that observes context cancellation.
9. Wire an ephemeral listener wrapped with a listener-close observer for the new-request rejection test.
10. Write the verification tests.
11. Run the test suite with `-race` and confirm every test passes deterministically.
12. Review the verification list and confirm every item is covered before declaring the project complete.

## 14. Verification Cases the Learner Must Write

Each item is a behavioural specification. The learner writes the corresponding `go test` code.

- Clean in-flight completion: a request is in flight when the trigger fires; the handler completes within the 5-second graceful budget; the orchestration returns nil; the dependency's `Close` is called exactly once; the lifecycle-owned goroutine and channels are joined.
- New-request rejection after shutdown: the listener-close observer's barrier is released after the trigger; a subsequent request attempt against the listener fails to connect; the test does not rely on timing.
- Controlled force-close path: the test shutdown-context factory returns a controllable context. The test signals the shutdown-start barrier, then ends the controllable context. The orchestration calls `Server.Close` and joins. The returned aggregate contains `context.DeadlineExceeded` from the joined error. The dependency is still closed exactly once. The handler observes cancellation or releases. This is the test double for the production background-based 5-second timeout; the production factory is untouched by this test.
- Graceful context freshness without wall-clock observation: the parent context is cancelled before the trigger fires. The injected shutdown-context factory records that it was called. The factory returns a live controllable context whose `Done` channel is not closed at the moment of construction. A barrier-blocked in-flight handler is released and drains cleanly while the controllable context is still live. This proves the cancelled parent context was not reused as the shutdown context. The test does not measure wall-clock duration.
- Dependency close order: a test passes two dependencies whose `Close` methods record their order; the test asserts that the dependencies are closed only after `Serve` has returned and in declared order.
- Dependency close exactly once: a dependency whose `Close` increments a counter is closed exactly once across a successful shutdown, a force-close shutdown, and a repeated-trigger shutdown.
- Dependency close continues after error: a dependency whose `Close` returns an error is followed by another dependency whose `Close` is still called; the error from the first dependency appears in the aggregate.
- Unexpected forced `Close` error: the observed test listener's `Close` returns a sentinel error after signalling or closing the underlying listener; the returned aggregate retains that sentinel through the join. The test does not rely on an impossible random real-server error.
- Repeated trigger: a test activates the trigger twice in quick succession; only the first call activates the shutdown; the second call returns without blocking; the orchestration does not panic and does not start a second shutdown.
- Trigger before `Serve` is ready: a test activates the trigger before the orchestration has called `Server.Serve`; the trigger call returns without blocking; once `Serve` is running, the shutdown proceeds normally.
- Trigger after the orchestration has returned: a test activates the trigger after the orchestration returns; the trigger call returns without blocking; the orchestration does not panic.
- Unexpected `Serve` error before trigger: the listener is closed from outside before `Serve` returns; the orchestration returns an aggregate that contains the listener error; dependencies are closed exactly once; the orchestration does not wait forever for a trigger.
- Joined lifecycle: a test asserts that the lifecycle-owned goroutine and the lifecycle-owned completion channels are joined before the orchestration returns. The assertion is direct, not based on `runtime.NumGoroutine`.
- Error aggregation: a test forces three distinct errors (a graceful context error, an unexpected `Server.Close` error, and a dependency error) and asserts the returned `errors.Join` aggregate contains all three.
- Forced path returns non-nil: a forced-path test asserts the returned error is not `nil` and that `context.DeadlineExceeded` is reachable through `errors.Is` or `errors.As` against the aggregate.
- Race-free: every test above passes under `go test -race ./...`.
- Source-review static gate for `os.Exit` absence: the orchestration source is searched for the literal `os.Exit` and the search must return zero matches. The orchestration is the only Go file reviewed, so the gate is bounded and unambiguous. A process test that forks a subprocess to detect `os.Exit` is not required and is not part of this contract.
- Source-review static gate for `os/signal` isolation: the orchestration source is searched for an import of `os/signal` and the search must return zero matches. Only the production `main` imports `os/signal`; the orchestration and the tests do not.

## 15. Common Mistakes to Watch For

- Deriving the shutdown context from the parent `ctx`. If the parent is cancelled, the shutdown has no time to drain.
- Closing dependencies before serving has quiesced. In-flight handlers will fail or panic.
- Closing a dependency twice. Many resources refuse double-close; some panic.
- Calling `os.Exit` inside the orchestration. The dependency loop is skipped.
- Treating `http.ErrServerClosed` as an error. It is the normal shutdown signal.
- Using `time.Sleep` in tests. The tests become flaky and slow.
- Using fixed ports. The tests fail in CI when the port is busy.
- Using real signals in tests. The tests become unreliable across platforms and CI.
- Returning `nil` from a forced path. The graceful context error must be surfaced.
- Starting a background goroutine to invoke `Server.Shutdown` "in case the trigger blocks". The orchestration is synchronous with respect to the trigger.
- Forgetting to set `ReadHeaderTimeout`. A slow client can hold the server open.
- Calling `Server.Close` before `Server.Shutdown` has had a chance to drain. The two calls are ordered.
- Letting a second trigger block. The trigger abstraction must coalesce.
- Using `runtime.NumGoroutine` as the primary leak assertion. It is noisy and unreliable.

## 16. Topics and References for Study

- Go `net/http` package documentation, especially `Server`, `Server.ListenAndServe`, `Server.Serve`, `Server.Shutdown`, `Server.Close`, `Server.RegisterOnShutdown`, and the `Handler` interface.
- Go `context` package documentation, especially `WithTimeout`, `WithCancel`, and the rule that contexts are not stored in structs.
- Go `errors` package documentation for `errors.Join` and error wrapping.
- Go `os/signal` package documentation for `signal.Notify` and the channels it returns.
- Go `net` package documentation for `net.Listen`, `Listener`, and the loopback address family.
- Go `sync` and `sync/atomic` package documentation for `atomic.Bool`, `sync.Once`, and `sync.WaitGroup`.
- The Go blog post "Graceful shutdown in Go http.Server" (no link; the learner can search for it).

## 17. Self-Assessment Questions

1. Why must the shutdown context be created fresh from `context.Background()`, rather than derived from the parent `ctx`?
2. Why are dependencies closed only after serving has quiesced? What breaks if the order is reversed?
3. Why is each dependency closed exactly once? What is the failure mode of double-close?
4. Why is `http.ErrServerClosed` not an error? What would logging it as an error cause?
5. Why are repeated triggers idempotently coalesced? What would happen if a second trigger started a second shutdown?
6. Why are the four server timeouts (`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`) each pinned to a finite value? What attack does each one address?
7. Why is `os.Exit` forbidden inside the orchestration, and why are the `os.Exit` and `os/signal` gates source-review static checks rather than process-level tests? What does `os.Exit` skip, what does a process-level test buy that a static check does not, and why is that added complexity not worth it here?
8. Why is the forced path required to return a non-nil error? What client behaviour does this enforce?
9. Why is `runtime.NumGoroutine` not the primary leak assertion? What does the lifecycle-owned completion assertion measure that `NumGoroutine` does not?
10. Why is the freshness test a barrier-and-release test rather than a wall-clock observation test? What would a wall-clock observation prove or fail to prove?

## 18. Definition of Completion

The project is complete when, in addition to the rules above:

- Every item in the verification list is a passing test that the learner wrote themselves.
- The tests pass under `go test -race ./...` from the project folder.
- The orchestration contains no third-party imports and no `time.Sleep`.
- The orchestration never calls `os.Exit`.
- The configuration struct has the exact pinned values and a constructor that rejects non-positive values for any timeout or budget.
- The production `main` is the only place that imports `os/signal`; the orchestration and the tests do not import it. The source-review static gate for `os/signal` and the source-review static gate for `os.Exit` are both satisfied.
- The learner can answer every self-assessment question without rereading the README.

## 19. Optional Extensions

At most two. Pick one only if the core project is already complete and tested. Optional extensions must not weaken any documented contract.

- Add a `Ready` handler that returns `503` once the shutdown sequence has begun. Documented as a readiness probe hook. The handler does not require CSRF.
- Add a lifecycle event recorder that the orchestration calls at the documented transitions (serve started, shutdown started, `Server.Shutdown` returned, force-close started, dependency `i` closed, run returned). The recorder is an injected interface used by tests for assertions; it is not part of the production configuration and must not affect the shutdown sequence.
