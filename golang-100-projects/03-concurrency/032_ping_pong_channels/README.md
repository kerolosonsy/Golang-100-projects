# Project 032 — Ping-Pong Channels

## 1. Project Name and Number

Project 032 — Ping-Pong Channels. Lives in `03-concurrency/032_ping_pong_channels`.

## 2. Project Idea

Two goroutines exchange exactly one value each through channels for a precise positive number of rallies, producing a deterministic alternating transcript that begins with ping. A single coordinator owns startup, completion, and channel closure. The transcript is delivered through a structured result that distinguishes four cases: success with the full transcript and no error, validation rejection with an empty transcript and an error, cancellation with the transcript committed before cancellation and the context error/status, and zero-count success with an empty transcript and no goroutines. The rally may be cancelled mid-flight by a context.

## 3. Why This Project Now?

Project 031 gave you a multi-result fan-in that respects cancellation, and Project 032 distils that primitive to two participants so that every lifetime is visible. The next project, 033, reuses the channel ownership rules at fan-out scale against `net/http`. This project is the minimalist version where the players are exactly two and the channel lifetimes are right in front of you, and the result surface is the place to pin the closed-and-drained zero value behaviour.

## 4. Prerequisites

The curriculum map's stated dependency for this project: Project 031. The fan-out, fan-in primitives from Project 031 must be second nature. No new packages are introduced here; this project builds solely on what 031 establishes.

## 5. What You Must Know Before Starting

The semantics of sending on a closed channel — it panics — and receiving from a closed channel, which yields a value immediately while values remain and then yields the zero value immediately with a false receive status once the channel is drained. The rule that a channel is closed exactly once, by a designated owner, only after every sender that uses it has finished. The `range`-until-close pattern on the receive side and how it ends without ambiguity: `range` over a closed-and-empty channel terminates, and the loop never sees a "successful receive of the zero value" because the receive status is false. That `select` picks one ready case pseudo-randomly and gives no application-level priority guarantee when multiple cases are ready. That `context.Done` is a regular channel and integrates into `select` directly. That a goroutine that runs forever or that blocks forever after its work is done is a leak, observable under `-race` and through a bounded deadline test that fails on hang.

## 6. Explanation of New Concepts

Two participants pass a token back and forth through channels until a count is exhausted. The defining property is that the transcript is deterministic and fully ordered even though two goroutines run at the same time. The first item of the transcript is always the ping value. If the count is one, the transcript is exactly one item. If the count is two, the transcript is two items: ping then pong. A count of zero produces an empty transcript without spawning any goroutine. A negative count is invalid input and produces a validation rejection with no goroutines spawned.

Channel ownership in this project: multiple goroutines may legally send on a single channel. The close rule is that a designated owner closes the channel exactly once, only after every sender that uses it has finished. In this project both players may send transcript entries on the transcript channel, and the coordinator alone closes the transcript channel after both players exit. The coordinator is also the designated owner of any token channel the players exchange, because the coordinator is the goroutine that learns when both sides have ended. Players never close any shared channel.

Receiving from a closed-and-drained channel yields the zero value and a false receive status. The rally logic must not interpret the zero value as a rally hit; it uses the receive status instead. The result boundary distinguishes success, validation rejection, zero-count success, and cancellation, and surfaces the difference rather than collapsing it.

Cancellation is cooperative. When the coordinator cancels the rally before completion, both goroutines must terminate and the channels must close without further sends. "Both goroutines terminate" includes the case where one is mid-send and the other is mid-receive; both must observe the cancellation through a `select` against `ctx.Done()`. The final hit of a successful rally does not require an extra blocking send; every send and every receive that could wait also observes the context.

## 7. Learning Objective

After this project you can explain, in writing, which goroutine closes each channel and why. You can implement a producer/consumer pair whose transcript is captured through a channel rather than printed as it occurs, and whose result boundary distinguishes success, validation rejection, zero-count success, and cancellation. You can use `context` to interrupt both goroutines of a pair without deadlocking either side and without requiring a final blocking send. You can distinguish a count of zero from a negative count, and you can write tests for both.

## 8. Functional Requirements

1. The driver accepts a non-negative integer count and a context and returns a structured result that distinguishes four cases: success with the full transcript and no error, validation rejection with an empty transcript and an error, zero-count success with an empty transcript and no error, and cancellation with the transcript committed before cancellation and the context error/status.
2. Negative counts produce a result whose transcript is empty and whose error describes the validation failure; no goroutines start.
3. Count zero produces a successful result whose transcript is empty; no goroutines start.
4. Positive counts produce a transcript that alternates starting with the ping value, of length equal to the count, and the result carries no error.
5. Both players send transcript entries on the transcript channel. The coordinator alone closes the transcript channel, exactly once, after both players exit.
6. The coordinator is the designated owner of any token channel the players exchange; players never close any shared channel.
7. Receiving from the closed-and-drained transcript channel yields the zero value with a false receive status. The design does not interpret the zero value as a rally hit.
8. Cancellation cancels the rally before completion. The result carries the transcript committed before cancellation, plus the context error/status. Both goroutines terminate.
9. The final hit of a successful rally does not require an extra blocking send; every send and every receive that could wait observes the context through a `select`.

## 9. Inputs and Outputs

**Input** is a non-negative integer count and a context.

**Output** is a structured result that the test can read directly. The four cases are success, validation rejection, zero-count success, and cancellation. The transcript is the source of rally assertions in success and cancellation cases.

**Behaviour example (text only).** Count four with an unimpeded context. The result is success with no error. The transcript is exactly `["ping", "pong", "ping", "pong"]` in that order.

**Behaviour example (text only).** Count zero. The result is success with no error. The transcript is empty and zero goroutines are spawned.

**Behaviour example (text only).** Count minus one. The result is a validation rejection with the validation error; the transcript is empty and zero goroutines are spawned.

**Behaviour example (text only).** Count three with a cancellation that arrives after the first item has been produced. The result indicates cancellation and carries the context error/status. The transcript contains the items produced before cancellation. Both goroutines terminate.

## 10. Rules and Edge Cases

Count zero: empty transcript, no goroutines launched, result is success with no error. Count one: transcript of length one whose only item is the ping value. Count two: transcript of length two, ping then pong. Negative count: result is a validation rejection, transcript empty, no goroutines launched. Cancellation while a send is in flight: the in-flight send either completes or is drained; further sends are not attempted. Cancellation while a receive is in flight: the receive unblocks and the player returns without sending another value. The final hit of a successful rally does not require an extra blocking send; cancellation is observed by every send and every receive that could wait. Repeating the same rally many times in a row does not accumulate goroutines. Players never close any shared channel. Receiving from a closed-and-drained channel returns the zero value and a false receive status; the rally logic does not interpret that zero value as a rally hit.

## 11. Project Constraints

Standard library only. No `time.Sleep` or wall-clock `time.After` as the normal synchronization mechanism. A bounded test deadline may be used only as a deadlock guard that fails the test if the rally has not finished by that wall-clock point; it is never used as the synchronization mechanism itself. No goroutine leaks; every goroutine has a visible exit path that runs whether the rally finishes or is cancelled. The transcript is captured through a structured result and the tests read it directly; they do not depend on standard-output ordering. The race detector must report nothing under `-race`.

## 12. Design Questions Before Coding

Who creates the channels and who closes them, in terms of a designated owner that closes only after every sender that uses the channel has finished? What does each goroutine wait on besides the message channel? How does the coordinator learn that both players have finished, so it can close the channel at the right time? How is the transcript accumulated, and how does the test verify it without relying on output ordering? What is the exact termination sequence when the context is cancelled mid-rally, including the requirement that the final hit does not need an extra blocking send? Is the rule "first item is always ping and the rest alternate" the right rule, with the consequence that an even count ends on pong and an odd count ends on ping? How does the code distinguish cancellation that arrives before the first send from cancellation that arrives mid-rally? How does the receive side avoid interpreting the zero value from a closed-and-drained channel as a rally hit?

## 13. Implementation Milestones

1. Decide on a transcript representation that supports empty, length-one, and length-many slices, and that the test can read directly.
2. Decide on the result boundary that distinguishes the four cases; the boundary is the only place the test looks.
3. Implement the validation-rejection path for negative counts before any goroutine launches.
4. Implement the zero-count path that returns an empty success result with no goroutines.
5. Spawn the two players for positive counts, with the first player set to send the ping value first.
6. Implement the coordination between the two players so that each waits for the other's value before sending its own, and so that the final hit does not require an extra blocking send.
7. Implement the coordinator's wait for both players and the single close of the transcript channel by the coordinator.
8. Implement the coordinator's ownership of any token channel; players never close shared channels.
9. Add cancellation: both players observe context cancellation through `select` and terminate without sending further values.
10. Confirm under repeated runs that goroutines do not accumulate.
11. Run under `-race` and confirm no data races.

## 14. Verification Cases the Learner Must Write

Count zero produces a success result with an empty transcript and no goroutines launched. Count one produces a success result with a transcript of exactly one item whose value is ping. Count two produces a success result with a transcript of exactly `["ping", "pong"]`. Count many produces a success result with a transcript of correct length whose items alternate starting with ping. Negative count produces a validation rejection with an empty transcript and no goroutines launched. Cancellation before the first send produces a cancellation result with an empty transcript; both goroutines terminate. Cancellation mid-rally produces a cancellation result whose transcript contains only the items produced before the cancellation; both goroutines terminate; no further sends occur. Repeated runs of the same count produce the same result and do not leak goroutines. Running under `-race` produces no race report. No goroutine remains blocked on a send or receive after the rally terminates or is cancelled, verifiable through a bounded deadline test that fails on hang. The transcript channel is closed exactly once at the end of a successful rally, by the coordinator. No player closes any shared channel. The zero value received from the closed-and-drained transcript channel is never read as a rally hit; the receive status is used instead.

## 15. Common Mistakes to Watch For

Closing a channel from the receiving side after a `range` loop ends, or letting a player close the transcript channel "for safety"; only the coordinator closes after every sender has finished. Assuming the Go default allows only one sender per channel; multiple goroutines may legally send on a single channel, and the close rule is about the designated owner, not about sender count. Treating `select`'s pseudo-random ready-case choice as a priority signal; the choice is not biased. Interpreting the zero value received from a closed-and-drained channel as a rally hit; receive status, not the value, tells the loop that the channel is drained. Adding a final extra blocking send after the last rally step so the test "knows" the rally is over; the final hit must not require an extra send that could wait, and every send and receive that could wait must observe the context. Using `time.After` to "be safe" against deadlocks; that hides real bugs and makes tests flaky. Forgetting the bounded deadline as a deadlock guard, so a deadlock test passes forever. Treating `time.Sleep` as a synchronization mechanism; the project forbids it.

## 16. Topics and References for Study

The "share memory by communicating" section of Effective Go. The Go specification on channel close and on the comma-ok receive form `v, ok := <-ch`, including how `ok` is false on a closed-and-drained channel. The `context` package documentation, especially cancellation propagation across goroutines. The `sync` package documentation, especially `WaitGroup` and the placement of `Add`. The Go blog article on pipelines and cancellation, which presents the canonical fan-in pattern this project distils to two participants.

## 17. Self-Assessment Questions

Which goroutine is the designated owner that closes the transcript channel, and why is that the right owner under a rule that allows multiple senders on one channel? What is the difference between "the players have finished sending" and "the channel is closed", and what establishes the latter from the former? Why does count zero produce an empty success result rather than a transcript of `["ping"]`? If cancellation arrives while a player is sending, what guarantees does the rest of the rally have about further sends, and why does the final rally step not need an extra blocking send? Why does the design avoid interpreting the zero value from a closed-and-drained channel as a rally hit? Why must the test not depend on interleaved standard output? Why is `time.Sleep` disallowed as a synchronization mechanism, and what bounded primitive is permitted instead, and only as what? How do repeated runs of the same rally detect a goroutine leak? What does running the test under `-race` actually verify in this project that a successful run alone does not?

## 18. Definition of Completion

Every Functional Requirement is implemented and exercised by a passing test. The Behaviour Examples in this README hold. No `time.Sleep` is used as the synchronization mechanism; only a bounded test deadline may be used, and only as a deadlock guard. The result boundary distinguishes success, validation rejection, zero-count success, and cancellation. The transcript is captured through the result boundary and is the sole source for assertions. Tests run with `-race` and produce no race report. No goroutine remains blocked after a run terminates, even on cancellation. Repeated runs do not leak goroutines. The transcript channel is closed exactly once, by the coordinator, after both players exit. Players never close any shared channel. The zero value from a closed-and-drained channel is never treated as a rally hit; the receive status is used. You can answer every Self-Assessment Question without consulting the README.

## 19. Optional Extensions

Add a third participant that joins the rally as a relay at a configurable boundary, whose arrivals are visible in the transcript but who never closes a channel it does not own. Add a deterministic leading and trailing padding to the transcript, for example markers around the alternating core, whose presence does not affect the alternating invariant and whose zero-value receives are still not interpreted as rally hits.
