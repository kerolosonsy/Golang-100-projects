# Project 068 — Connection Pooling

## 1. Project Name and Number
Project 068, connection_pooling. This README is a learning guide only. You will create every source, Compose, and test file yourself in `05-databases/068_connection_pooling/`. This guide contains no implementation code, signatures, snippets, pseudocode, SQL, or solution commands.

## 2. Project Idea
Build a controlled PostgreSQL experiment around `database/sql` pool configuration and `DBStats`. Apply one pinned configuration, occupy every allowed connection behind synchronized database-side work, create a bounded waiter, observe cumulative wait statistics, release deterministically, and prove every resource becomes reusable.

## 3. Why This Project Now?
Project 068 follows Project 067 in the catalog, then returns to SQL to isolate connection-pool behavior from transaction business logic and make resource pressure, wait statistics, cancellation, and cleanup directly observable. Projects 061 and 041 supply required database and service foundations.

## 4. Prerequisites
Required prerequisites: Projects 067, 061, and 041. Optional review: Project 066 for recent PostgreSQL transaction context; it is not required. The normal unit gate must need no Docker, PostgreSQL, network, or environment variables. PostgreSQL integration is separate and opt-in.

## 5. What You Must Know Before Starting
Know `database/sql`, contexts, connection pools, query result lifecycle, concurrency barriers, channels, polling with deadlines, and test cleanup. Understand snapshots versus cumulative counters and why elapsed timing is not a correctness oracle.

## 6. Explanation of New Concepts
The experiment configuration is exact: maximum open connections 3, maximum idle connections 2, maximum connection lifetime 30 minutes, and maximum connection idle time 5 minutes. Validate maximum open above zero, maximum idle from zero through maximum open, and both durations nonnegative before applying anything.

`DBStats` exposes current and cumulative observations. Only `DBStats.MaxOpenConnections` directly reports a configured limit. `DBStats.OpenConnections` and the related InUse and Idle fields are current snapshots; `DBStats.Idle` is not the configured MaxIdle value and does not report it. `database/sql` exposes no getter and no `DBStats` field for configured Maximum Idle Connections, Maximum Connection Lifetime, or Maximum Connection Idle Time. Those three settings remain assertions on the validated configuration that the apply boundary records. WaitCount is cumulative. Do not claim WaitDuration or wall-clock timing is deterministic.

Deterministic observation starts with a baseline stats snapshot. Occupy exactly all three connections behind a synchronized database-side barrier. Start at least one fourth query with a bounded context. Wait through explicit events or poll-with-deadline until WaitCount increases from baseline. The contention test asserts `DBStats.MaxOpenConnections == 3`, asserts `DBStats.OpenConnections` and `DBStats.InUse` never exceed three while the controlled hold is active, and asserts a positive `DBStats.WaitCount` delta from baseline. Release the barrier and prove all work completes. Assert deltas, not global zero values.

Rows own a connection until consumed or closed. A unit fake can model Rows ownership in isolation but cannot prove real `database/sql` pool ownership; only the tagged PostgreSQL integration proves that a deliberately unconsumed real Rows occupies capacity and that closing it makes the connection reusable. The tagged integration temporarily keeps one result set unconsumed only until pool pressure is observed, then guaranteed cleanup closes it, checks terminal rows error where iteration occurs, and proves the connection serves another operation. No completed path leaks Rows or a database handle.

## 7. Learning Objective
Create a deterministic, resource-safe pool experiment that distinguishes validated configuration from observable stats, proves bounded concurrency and waiting without sleep-based races, and demonstrates how Rows lifecycle affects connection reuse.

## 8. Functional Requirements
1. Use `github.com/jackc/pgx/v5` exactly at `v5.10.0` through its `stdlib` adapter with `database/sql`.
2. Pin maximum open connections to 3, maximum idle connections to 2, maximum lifetime to 30 minutes, and maximum idle time to 5 minutes.
3. Validate maximum open is positive, maximum idle is nonnegative and no greater than maximum open, and durations are nonnegative before applying settings.
4. Invalid configuration applies nothing and returns typed invalid configuration.
5. Preserve the validated configuration as evidence for Maximum Idle Connections, Maximum Connection Lifetime, and Maximum Connection Idle Time because `database/sql` exposes no getters or `DBStats` fields for them.
6. Capture a baseline `DBStats` snapshot before contention.
7. Occupy exactly three connections behind a synchronized database-side or injected barrier.
8. Start at least one context-bounded fourth waiter only after all three occupants confirm arrival.
9. Observe WaitCount delta through events or polling with a deadline, not an unbounded loop or fixed sleep.
10. Assert `DBStats.MaxOpenConnections == 3`, `DBStats.OpenConnections <= 3`, and `DBStats.InUse <= 3` while the controlled hold is active.
11. Observe a positive `DBStats.WaitCount` delta from the baseline snapshot, not from a global zero assumption.
12. Release occupants through channels or a controlled database lock and prove occupants and waiters finish.
13. Assert cumulative deltas from baseline rather than assuming process-global counters begin at zero.
14. Do not assert deterministic WaitDuration, scheduling, or elapsed time.
15. Every query closes Rows and checks terminal rows error where rows are iterated.
16. The deliberate unconsumed-Rows demonstration is temporary, has guaranteed cleanup, and proves reuse after close. A unit fake can model this in isolation; only the tagged integration proves real `database/sql` pool ownership and connection reuse.
17. Every waiter has a context timeout or cancellation boundary.
18. The database handle is closed exactly once after experiment cleanup.
19. Unit tests validate configuration and orchestration with fakes and require no Docker. Integration is separately tagged and guarded.

## 9. Inputs and Outputs
Input is a context, the exact experiment configuration, and synchronization boundaries. Output is a structured observation containing baseline and final stats, maximum observed open connections, WaitCount delta, participant completion, waiter cancellation where requested, resource reuse evidence, and close outcome. Invalid configuration is distinct from context or database failure.

Example behavior: three occupants report that they hold all available connections. A fourth request begins and WaitCount rises relative to baseline while OpenConnections remains at most three. Releasing the barrier lets all operations finish and a follow-up query proves the pool remains usable.

## 10. Rules and Edge Cases
Validate before applying. Synchronize occupant arrival before starting waiters. Use bounded event waits or poll-with-deadline, never correctness sleeps. Treat stats as snapshots that can change immediately. Treat WaitCount as cumulative. Only `DBStats.MaxOpenConnections` reports a configured limit; `DBStats.Idle` is a snapshot, not the configured Maximum Idle limit. Never require WaitDuration to equal a duration. Close Rows on every path, including the deliberate hold. Close the database only after workers and cleanup finish.

## 11. Project Constraints
This is a controlled correctness experiment, not a benchmark and not guidance that 3 and 2 are universal production values. No throughput, latency, or ideal-sizing claims. Normal tests require no Docker. Integration files use the `integration` build tag. Integration uses unique disposable scope and bounded owned cleanup only. Never print credentials or alter shared data.

## 12. Design Questions Before Coding
Which configuration checks happen before any setter call? Which settings are directly visible in `DBStats` and which are only asserted through the validated configuration? How does each occupant prove it holds a connection? How does the fourth waiter prove pool contention without timing assumptions? How is `DBStats.MaxOpenConnections` sampled safely? How does guaranteed cleanup release deliberately held Rows? Where does a unit fake end and the tagged integration begin in proving real `database/sql` row ownership? How is database close ordered after worker completion?

## 13. Implementation Milestones
1. Define exact configuration, validation, apply boundary, observation result, and typed outcomes.
2. Build fake-based orchestration with baseline snapshot, three synchronized occupants, bounded waiter, WaitCount-delta observation, release, and completion.
3. Add maximum-open observation and `DBStats` snapshot assertions limited to what the API actually exposes.
4. Add the temporary Rows-hold experiment at the fake boundary with guaranteed close and reuse proof.
5. Add context cancellation for waiting work and exact database-close lifecycle.
6. Complete unit tests without PostgreSQL or Docker.
7. Separately add guarded tagged PostgreSQL integration that asserts `DBStats.MaxOpenConnections == 3`, snapshot ceilings on OpenConnections and InUse, a positive WaitCount delta, and real `database/sql` row ownership over a deliberately unconsumed real Rows.

## 14. Verification Cases the Learner Must Write
Unit tests without Docker:
- Accept the exact 3, 2, 30-minute, and 5-minute configuration.
- Reject nonpositive maximum open, negative maximum idle, idle above open, and negative durations before any apply call.
- Apply all four validated settings in defined order through a fake boundary.
- Preserve Maximum Idle, Maximum Lifetime, and Maximum Idle Time values in the validated configuration because no getter and no `DBStats` field exists for them.
- Capture baseline before occupants start.
- Require exactly three synchronized occupant-arrival events before starting a fourth waiter.
- Observe WaitCount as a positive delta from a nonzero fake baseline.
- Record maximum open at or below three through whatever field the fake boundary exposes, and never claim `DBStats.Idle` reports the MaxIdle limit.
- Release through an event and prove every participant completes within test deadlines.
- Cancel a waiting request and preserve context classification.
- Hold a fake Rows resource temporarily, observe blocked reuse, guarantee close, and prove reuse afterward at the fake boundary.
- Close every Rows resource on success, scan failure, context failure, and orchestration failure.
- Check terminal rows error after iteration.
- Close the database exactly once after all workers complete.
- Avoid assertions on WaitDuration and elapsed timing.

Opt-in PostgreSQL integration tests:
- Integration files are excluded from the normal gate by the `integration` build tag.
- With that tag, both runtime activation values absent produce a clear skip. The values are a PostgreSQL connection setting and an explicit destructive-test guard. Partial activation, malformed connection settings, missing or wrong guard, or inability to prove unique disposable scope fails closed before connecting or mutating.
- Generate a collision-resistant owned database or schema name, validate it before any connection, connect, create or use only that scope, and drop only it through bounded cleanup. Never print credentials or destroy shared data.
- Capture baseline, occupy exactly three real connections behind a synchronized database-side barrier, and start a bounded fourth waiter.
- Assert `DBStats.MaxOpenConnections == 3` from real `database/sql`, assert `DBStats.OpenConnections <= 3` and `DBStats.InUse <= 3` snapshots during the controlled hold, and observe a positive `DBStats.WaitCount` delta using events or poll-with-deadline.
- Release without fixed sleeps and prove all work completes.
- Cancel a waiter through context and verify the pool remains usable.
- Keep a real result set deliberately unconsumed only until its pool effect is observed, close it through guaranteed cleanup, and prove connection reuse against real `database/sql`.
- Verify every Rows and the database handle are closed.

## 15. Common Mistakes to Watch For
Treating `DBStats` as immutable, asserting counters start at zero, using sleep as synchronization, claiming WaitDuration is deterministic, starting the waiter before all occupants arrive, leaking held Rows, failing to check terminal rows error, closing the database while workers run, claiming `DBStats.Idle` reports the configured MaxIdle limit, claiming getters or `DBStats` fields exist for Maximum Idle, Maximum Lifetime, or Maximum Idle Time, asserting a unit fake proves real `database/sql` row ownership, benchmarking instead of testing bounds, or cleaning shared PostgreSQL objects.

## 16. Topics and References for Study
Study Go `database/sql` documentation for pool setters, `DBStats` fields including only the `MaxOpenConnections` getter-equivalent, `Rows`, context cancellation, and `DB.Close`. Study `github.com/jackc/pgx/v5` `v5.10.0` and its `stdlib` adapter. Study concurrency barriers, channels, condition signaling, polling with deadlines, cumulative counters, and resource ownership in tests, including the explicit boundary between fake-based unit proofs and integration proofs of real pool behavior.

## 17. Self-Assessment Questions
Why is WaitCount asserted as a delta? Why can stats snapshots change immediately? Which three configuration values lack getters and `DBStats` fields? Why does `DBStats.Idle` not report the configured MaxIdle limit? Why must all three occupants signal before the fourth starts? Why is WaitDuration unsuitable as a deterministic assertion? How does an unclosed Rows hold capacity, and why can only the tagged integration prove that against a real `database/sql` pool? What proves cleanup made that connection reusable?

## 18. Definition of Completion
- [ ] Exact pool configuration and all pre-apply validation rules are tested.
- [ ] Baseline, three synchronized occupants, bounded fourth waiter, WaitCount delta, release, and completion are deterministic.
- [ ] Integration asserts `DBStats.MaxOpenConnections == 3` and snapshot ceilings on `DBStats.OpenConnections` and `DBStats.InUse` during the controlled hold.
- [ ] No test relies on fixed sleeps, WaitDuration, or global-zero counters.
- [ ] Every Rows is closed and terminal rows error is checked where applicable.
- [ ] Temporary resource hold has guaranteed cleanup and connection reuse proof; unit fake proves orchestration, tagged integration proves real `database/sql` ownership.
- [ ] Unit tests pass without Docker, PostgreSQL, network, or environment variables, including the race detector.
- [ ] Tagged integration activation skips only when both values are absent and otherwise fails closed before unsafe access.
- [ ] Integration cleanup is bounded to unique disposable scope and credentials are never printed.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, SQL, Compose content, or solution commands.

## 19. Optional Extensions
Add an experiment comparing one and several bounded waiters while preserving event-driven assertions. Add observability export for selected stats without making performance claims.
