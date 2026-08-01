# Project 038 — Mutex Bank Account

## 1. Project Name and Number

Project 038 — Mutex Bank Account. The project teaches atomic account operations, integer money, invariant preservation, and safe ownership of mutex-containing state.

## 2. Project Idea

Model one bank account whose balance is stored in signed 64-bit integer cents. The account supports positive deposits, withdrawals that never make the balance negative, and balance reads. Every operation is atomic under a mutex. Invalid or non-positive amounts fail without mutation, and insufficient funds fail without mutation.

The account is intentionally small so the important lesson stays visible: validation, checking, and mutation must happen as one protected decision. The API must not expose its mutex or internal mutable state. Transfers and multiple-account lock ordering are outside the required scope.

## 3. Why This Project Now?

Project 037 was the immediate predecessor and required explicit ownership across concurrent goroutines. This project contrasts channel ownership with direct mutex protection of one invariant: balance can never become negative. It also creates a precise foundation for later atomic operations and database transaction projects.

## 4. Prerequisites

Complete Projects 031 and 037 first. Project 031 supplies the barriers, wait groups, and cancellation patterns reused for deterministic concurrent tests. Project 037 supplies the explicit channel ownership and coordinated shutdown discipline that this project extends to a single mutex-protected integer invariant. You must understand goroutines, barriers or coordination channels, data races, and deterministic concurrent tests. You should know integer types, errors, method receivers, pointer ownership, and why copying a value containing a mutex after first use is unsafe. Review the standard library synchronization primitives before beginning.

## 5. What You Must Know Before Starting

Know that money amounts must be represented as signed 64-bit integer cents, never floating-point values. Know the difference between an operation's input validation and its state mutation. Know that a mutex must protect the balance check and update together. Know that a read must also synchronize with writes. Know that an account value containing a mutex must not be copied after use, including accidental value returns or assignments. Know that the signed 64-bit maximum is the project's overflow boundary and that overflow checks must use it directly.

## 6. Explanation of New Concepts

An account invariant is a condition that must remain true after every completed operation. Here the central invariant is a non-negative balance, with the additional rule that a failed operation leaves the balance exactly unchanged. The invariant is preserved only when the withdrawal affordability check and subtraction occur under the same lock.

Integer cents make each amount exact. Positive means strictly greater than zero; zero is not a no-op success. Deposit overflow is a separate failure from invalid input: an amount can be positive yet impossible to add without exceeding the signed 64-bit maximum. Overflow must be checked before addition against that fixed boundary, and a failed overflow attempt must not mutate the account.

A mutex is part of the account's synchronization state. Callers receive operation results, not the lock or a reference to internal state. The account should be used through a stable pointer or another ownership model that guarantees the mutex is not copied after first use. A read-only balance method still needs synchronization because concurrent writes may occur.

Concurrent withdrawals are intentionally tested with controlled affordability. When the initial balance can fund only a known number of equal withdrawals, exactly that count may succeed and all other withdrawals must fail without making the balance negative. Scheduling determines which callers succeed, not the invariant or final balance.

## 7. Learning Objective

By completion, you can protect a financial invariant with a mutex, perform exact cent arithmetic, distinguish invalid and insufficient operations, handle addition overflow safely, and design concurrent tests that verify final state rather than assume execution order. You can explain why exposing or copying synchronization state is unsafe.

## 8. Functional Requirements

1. The account stores one balance in signed 64-bit integer cents.
2. The initial balance must be a non-negative value; a negative initial balance is invalid.
3. Deposits require a strictly positive cent amount.
4. A deposit fails if addition would exceed the signed 64-bit maximum and leaves balance unchanged.
5. Withdrawals require a strictly positive cent amount.
6. A withdrawal fails when the amount exceeds the current balance and leaves balance unchanged.
7. A successful withdrawal subtracts exactly its requested amount and never makes balance negative.
8. Balance reads return a synchronized snapshot.
9. Every operation is atomic under one mutex; no check-then-act decision occurs outside the lock.
10. Invalid amounts and insufficient funds return errors with no mutation.
11. The mutex and internal mutable state are not exposed to callers.
12. The account's ownership model prevents copying a mutex-containing account after first use.
13. Transfers and lock ordering across multiple accounts are not required; no account identity or name is part of the required contract.

## 9. Inputs and Outputs

Inputs are a valid non-negative initial balance and positive cent amounts for deposit or withdrawal operations. The project does not require any account identity or name; one is not part of the required contract.

Outputs are operation success or an error class for invalid amount, insufficient funds, or deposit overflow, plus a balance snapshot for a successful or explicitly valid balance read. A failed operation must leave the observable balance unchanged.

Text-only example: an account starts at zero. A withdrawal for one cent fails. A deposit for 250 cents succeeds. A withdrawal for exactly 250 cents succeeds and returns the balance to zero. A second withdrawal fails without changing zero.

## 10. Rules and Edge Cases

Zero and negative deposits are invalid. Zero and negative withdrawals are invalid. An initial negative balance is invalid. A withdrawal equal to the current balance succeeds; one cent more fails. An insufficient withdrawal must not partially subtract or reserve funds for a later caller.

A deposit that would exceed the signed 64-bit maximum is an overflow error, not a successful deposit with wrapped value. The overflow check occurs before addition while holding the mutex and uses the signed 64-bit maximum as the project's overflow boundary. If the initial balance is already at the maximum, every positive deposit fails for overflow while valid withdrawals remain possible.

Concurrent deposits are all preserved unless an individual operation reaches the overflow boundary. Concurrent withdrawals are serialized by the mutex; at most the affordable count succeeds. The final balance equals the initial balance plus successful deposits minus successful withdrawals. Reads may occur concurrently and must never observe an unsynchronized or negative state.

Do not promise fairness or a particular winner among concurrent withdrawals. Do not add transfer behavior, multiple-account locking, overdraft, floating-point amounts, or implicit rounding.

## 11. Project Constraints

Use only the Go standard library. Store money in signed 64-bit integer cents and never use floating point. Protect all account state with a mutex. Do not expose the mutex, internal pointer, or mutable state. Do not copy the account after first use; use a stable ownership convention and make tests check for accidental copying where practical. Tests must use barriers, wait groups, and channels rather than sleep. The race detector must pass.

## 12. Design Questions Before Coding

How are cents represented in the signed 64-bit integer type, and what is the fixed overflow boundary? How is initial balance validated? Which errors distinguish invalid amount, insufficient funds, and overflow? What exact state does the mutex protect? Where do validation, affordability checking, overflow checking against the signed 64-bit maximum, and mutation occur? How can a balance read synchronize with writes? How will the API make copying after use difficult or clearly prohibited? How will controlled concurrent withdrawals prove only affordable operations succeed without assuming caller order?

## 13. Implementation Milestones

1. Define the balance, ownership, and error behavior in prose, including the non-negative invariant.
2. Pin the signed 64-bit integer cent type and the signed 64-bit maximum overflow boundary.
3. Define construction validation for the initial balance only; no identifier is required.
4. Define deposit, withdrawal, and balance-read operations without exposing mutable internals.
5. Protect each complete operation with the mutex, keeping check and mutation together.
6. Add sequential tests for zero balance, valid operations, invalid amounts, exact withdrawal, and insufficient funds.
7. Add overflow tests against the signed 64-bit maximum balance boundary.
8. Add concurrent deposit and controlled concurrent withdrawal tests with deterministic coordination.
9. Check ownership for mutex-copy hazards and run the entire package under the race detector.

## 14. Verification Cases the Learner Must Write

Test zero initial balance and a balance read. Test sequential deposits and withdrawals. Test zero and negative amounts for both operation types and verify no mutation. Test exact withdrawal success and one-cent-over withdrawal failure. Test deposit overflow at and near the signed 64-bit maximum and verify the failed operation leaves the prior balance unchanged.

Start many coordinated deposit goroutines and verify every valid deposit appears in the final balance. Start more withdrawal goroutines than the affordable count, release them through a barrier, and verify exactly the affordable total can succeed, no balance is negative, and the final balance is exact. Do not assert which goroutine wins. Run concurrent reads during writes and run the package under the race detector.

Add a review or compile-time-oriented test arrangement that makes accidental value copying visible, and document that a mutex-containing account must not be copied after first use. Verify independent account instances do not share balances or synchronization state.

## 15. Common Mistakes to Watch For

Using floating-point money creates rounding ambiguity. Checking affordability before locking allows two withdrawals to pass against one balance. Locking subtraction but not the preceding check still breaks the invariant. Adding before checking overflow against the signed 64-bit maximum can wrap the balance. Treating zero as a successful no-op hides invalid caller input. Returning internal state by pointer permits unsynchronized mutation. Exposing the mutex lets callers violate operation boundaries. Copying an account value after use copies synchronization state and can produce incorrect protection. Assuming concurrent withdrawals have a deterministic winner makes tests flaky. Holding a lock during unrelated external work needlessly increases contention.

## 16. Topics and References for Study

Study the standard library documentation for `sync.Mutex`, method receivers, signed 64-bit integer limits, errors, and the race detector. Review Go's guidance on copying values containing locks and static analysis for copy-lock hazards. Read about linearizability, invariants, atomic check-and-update operations, integer overflow against the signed 64-bit maximum, and concurrent test barriers. Compare this mutex-owned state with the channel-owned state in Project 037.

## 17. Self-Assessment Questions

Why are cents stored as signed 64-bit integers? Why must invalid and insufficient operations leave balance unchanged? Which exact statements belong inside the mutex-protected transaction? How is deposit overflow detected against the signed 64-bit maximum before addition? Why does a balance read need synchronization? What does the non-negative invariant guarantee under concurrent withdrawals? Why is the identity of the successful withdrawal caller unspecified? What happens when more callers request funds than the balance can afford? Why is copying a mutex-containing account after first use unsafe? Why are transfers and multiple-account lock ordering excluded?

## 18. Definition of Completion

The account uses signed 64-bit integer cents only, accepts only positive operation amounts, preserves a non-negative balance, handles overflow against the signed 64-bit maximum before addition, and returns errors without mutation for every invalid or unaffordable operation. Deposits, withdrawals, and reads are atomic under a mutex. Internal mutable state and the mutex remain private, and the ownership model prevents copy-after-use. Deterministic sequential and concurrent tests verify exact final balances, controlled withdrawal counts, independent instances, and race-free reads and writes. The full package passes the race detector.

## 19. Optional Extensions

Add an immutable transaction-history report that is produced from synchronized snapshots without exposing account internals. Add a single-account overdraft policy as a separate explicitly configured mode, with tests proving its invariant and error semantics remain distinct from the required account. No optional extension may introduce an account identifier, name, or multi-account coordination; those are explicitly out of scope.
