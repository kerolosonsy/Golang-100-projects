# Project 012 — Bill Splitter

## 1. Project Name and Number

Project **012** — `012_bill_splitter`. The directory name and number must match exactly.

## 2. Project Idea

A small command-line tool that takes a bill amount, a tip percentage, and a number of people, and produces the per-person share that exactly splits the bill-plus-tip total among those people. The total is rounded to the nearest cent once, then expressed as an integer number of cents; that integer is divided into a base share and a remainder smaller than the number of people, and the remainder is distributed deterministically one cent at a time starting with the first person. The shares are reported as decimal amounts that sum exactly to the rounded total.

The split must work for a single diner as well as for several diners, must accept a tip of zero, and must reject an invalid number of people rather than guess. The chosen remainder policy is fixed and documented up front; the program does not switch policies per input.

## 3. Why This Project Now?

Project 011 taught the learner to keep behavior as data and to inject I/O boundaries. Project 012 adds the first piece of *domain logic* with rules the user can verify by hand: a tip on top of a bill, divided across N people, with a chosen rounding policy. It is also the first project where the choice of numeric type matters in a way the learner can see and measure.

Because the inputs and outputs are small and the rules are easy to state in plain language, this is a good place to develop the habit of writing test cases as tables before writing the implementation. Each row of the table — amount, tip, people, expected total in cents, expected shares in cents, expected sum — is a contract the implementation must honor exactly.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 012 therefore requires:

- Completion of **011** (Interactive Menu).
- No prior knowledge of databases, concurrency, or HTTP.
- This project does **not** require project 014. The bill-splitter rejects malformed inputs on its own; it does not depend on a separate validator package.

## 5. What You Must Know Before Starting

- How to declare a `struct` in Go and how to attach methods to it with a receiver.
- The difference between a value receiver and a pointer receiver, and which is appropriate when a method reads but does not mutate.
- The meaning of `float64` as an IEEE-754 binary floating-point number, and that a binary float cannot represent most decimal fractions exactly. Numbers like `0.10`, `0.20`, and their sum `0.30` illustrate the point.
- That "rounding to the nearest cent" means: multiply by 100, round to an integer, treat that integer as the number of cents. That integer is exact; everything downstream is integer arithmetic on cents.
- The meaning of "integer division with remainder": for non-negative integers, `a / b` is the floor of the division and `a % b` is the leftover.
- How `fmt.Scan` reads whitespace-separated tokens from a stream and how `fmt.Fscan` reads from an arbitrary `io.Reader`.
- How to write a function that returns a regular value plus an `error`, and how the caller decides what to do with that error.
- How to express a small fixed policy in plain language so the tests can mirror it.

## 6. Explanation of New Concepts

### Structs as small data carriers

A `struct` groups fields that travel together. For a bill, those fields are the amount, the tip percentage, and the number of people. For a result, those fields are the total in cents, the per-person share for each diner, and the size of the leftover that the chosen policy decided to distribute.

### Methods as operations on a value

A method attaches a function to a type's receiver. The method can read the fields, derive new values, and return them. The struct's fields stay immutable from the caller's point of view when the receiver is a value receiver, which is the right default for a bill that does not change after it is constructed.

### Binary floating point versus decimal money

A `float64` holds a binary approximation of a real number. Many decimal values — `0.10`, `0.20`, `0.30`, `1.10`, and so on — cannot be represented exactly in binary, so arithmetic on them accumulates small errors. The product `0.10 * 3` may print as `0.30000000000000004` even though the user thinks of it as `0.30`. The honest response for money is: do the rounding step at the boundary between the user-facing decimal and the internal arithmetic, then keep the rest of the work in integers.

### Rounding to cents at the boundary

The contract for this project is: the bill-plus-tip total is rounded to the nearest cent once, expressed as an integer number of cents, and that integer is the only money value the splitter manipulates. From that point on, the per-person shares are also integers (number of cents per person), and the sum of those integers is exactly the integer total. No further rounding is performed at any other step.

If the bill is represented by `B` and the tip percentage by `T`, first calculate the bill plus `B` multiplied by `T / 100`. Multiply that total by 100 and round it to the nearest whole cent using Go's `math.Round` behavior (half away from zero). The resulting integer cent amount is the only total used by the split. This rounding rule is fixed for the project, not chosen per implementation or per input.

### The chosen remainder policy

Once the rounded total in cents is known, integer division by the number of people gives the base share, while the integer remainder gives the number of extra cents still to allocate. The **first as many people as there are remainder cents** get one extra cent each; everyone else gets the base share. The sum of all shares is exactly the rounded total in cents.

This policy is the one the README requires. The learner does not choose among multiple policies; the contract is fixed here. The verification cases pin the outcomes.

### Worked examples

- `110.00` split three ways: `totalCents = 11000`. `base = 3666`, `remainder = 2`. Shares in cents: `3667, 3667, 3666`. In decimal: `36.67, 36.67, 36.66`. Sum: `110.00`.
- `100.00` split three ways: `totalCents = 10000`. `base = 3333`, `remainder = 1`. Shares in cents: `3334, 3333, 3333`. In decimal: `33.34, 33.33, 33.33`. Sum: `100.00`.
- `10.00` split two ways: `totalCents = 1000`. `base = 500`, `remainder = 0`. Shares: `500, 500`. In decimal: `5.00, 5.00`. Sum: `10.00`.

### Validating the people count

Zero or negative people is not a bill split; it is a malformed input. The program must reject it with a clear error and must not divide by zero. A fractional people count is also rejected; people count is an integer greater than zero.

## 7. Learning Objective

After completing this project the learner can:

- Define a small `struct` for the inputs and another for the result.
- Round a bill-plus-tip total to cents once, then keep the rest of the work in integer cents.
- Divide an integer total across N people so the shares are exact integers whose sum equals the total.
- Reject an invalid people count with a clear error rather than a crash or a guess.
- Explain, in plain English, why binary floats accumulate small errors and why the "round once at the boundary, then use integers" approach keeps the user-visible result honest.
- Express the chosen remainder policy as both documentation and testable behavior.

## 8. Functional Requirements

1. Read a bill amount, a tip percentage, and a number of people.
2. Compute the bill-plus-tip total as `bill + bill * tipPercent / 100` and round that total to the nearest cent as a single step. The result is the only money value the splitter manipulates after this point.
3. Reject a negative bill amount with a clear error before any rounding or division.
4. Accept a tip percentage of zero; do not require a positive tip.
5. Accept any non-negative tip percentage; reject a negative tip percentage.
6. Reject zero, negative, or non-integer people counts with a clear error.
7. Split `totalCents` across `people` using the documented policy: `base = totalCents / people`, `remainder = totalCents % people`. The first `remainder` people receive `base + 1` cents; the remaining people receive `base` cents.
8. Produce a result that includes the rounded total and the per-person shares in both cents and decimal form, so the caller can verify the sum.
9. The sum of the per-person shares must equal `totalCents` exactly for every valid input.
10. Make the splitting logic reachable from a test without going through the interactive prompt.
11. Declare the rounding rule and the remainder policy in the package documentation, in plain English, and apply them consistently.

## 9. Inputs and Outputs

### Inputs

- A bill amount: a non-negative decimal number the user types. Examples: `42.50`, `100`, `7.25`.
- A tip percentage: a non-negative decimal number. Examples: `0` (no tip), `10`, `12.5`, `20`.
- A number of people: a positive integer. Examples: `1`, `2`, `4`.

### Outputs

- The total including tip, rounded to the nearest cent, shown in decimal.
- For one person, a single per-person share equal to the total.
- For several people, a per-person share for each diner, in order. The first `remainder` entries are one cent larger than the rest. (For typical small bills this is a difference of `0.01`.)
- A line that confirms the sum of the shares equals the total exactly.

### Example text-only success run (three people, 100 with 10% tip)

```
Bill amount: 100
Tip percent: 10
People: 3
Total: 110.00
Per person: 36.67
Per person: 36.67
Per person: 36.66
Sum of shares: 110.00
```

### Example text-only success run (three people, 100 with no tip)

```
Bill amount: 100
Tip percent: 0
People: 3
Total: 100.00
Per person: 33.34
Per person: 33.33
Per person: 33.33
Sum of shares: 100.00
```

### Example text-only error run

```
Bill amount: 50
Tip percent: -5
Tip must be zero or positive.
```

```
Bill amount: 50
Tip percent: 10
People: 0
Number of people must be a positive integer.
```

## 10. Rules and Edge Cases

- **One person**: the per-person share equals the total; `base = totalCents`, `remainder = 0`, no division step that could lose precision in a visible way.
- **Several people, totalCents divides evenly by people**: every share is the same integer; `remainder = 0`; the program does not invent extra cents.
- **Several people, totalCents does not divide evenly**: `remainder` is in `1 .. people-1`; the first `remainder` people pay one cent more than the rest.
- **Zero tip**: total equals bill; per-person shares are computed from the bill only; this is a normal run, not an error.
- **Tip with a fractional percentage**: allowed (for example `12.5`). The internal computation must handle this without silently truncating the tip.
- **Negative bill**: rejected with a clear error before any rounding or division.
- **Negative tip**: rejected with a clear error.
- **Zero, negative, or fractional people**: rejected with a clear error; the program does not divide by zero.
- **Non-numeric input**: treated as a parse error; the program explains which field failed.
- **Very large bill or very large people count**: behavior is whatever the chosen integer type can hold; the program does not promise arbitrary precision and must not pretend it does.

## 11. Project Constraints

- Go standard library only. No third-party decimal libraries. The point is to *learn* the rounding policy, not to outsource it.
- The bill-plus-tip total is rounded to the nearest cent exactly once. After that step, all money is in integer cents and arithmetic is integer arithmetic.
- The per-person shares are reported in decimal form. Each share is an exact conversion from a known integer number of cents.
- The sum of the per-person shares equals the rounded total exactly for every valid input.
- The splitting logic must be reachable from a test that does not depend on a terminal.
- No persistence; the program computes and reports, then exits.
- No currency conversion, no localization of the decimal separator, no storage of past bills — out of scope.

## 12. Design Questions Before Coding

- Where will the rounding step live — in the input parser, in the struct method, or in a private helper? Why is "round once at the boundary" easier to reason about than rounding at multiple places?
- Will you keep the total in cents as `int` or as `int64`? What does the choice imply for the largest bill-plus-tip value the program promises to handle?
- Where will the remainder distribution live — inside the splitter, or as a small helper that takes a total, a count, and a policy? Why does that separation make the tests easier to write?
- How will the program distinguish "the user typed zero people" from "the user typed a non-numeric value"? Both are errors, but the messages should be precise.
- How will the result type expose the per-person shares? A slice of shares, a slice of pairs of (index, share), or a single rounded share with a separate "remainder" field? Which choice makes the "first `remainder` people pay one cent more" rule obvious in the test?
- Will the test use table-driven cases where each row pins one policy decision? If a row has more than one decision in it, the test will be hard to debug.
- How will you document the precision contract in the package comment so the next reader knows the integer-cents rule is the honesty step, not a side note?

## 13. Implementation Milestones

1. Define the input struct (bill, tip percent, people count) with validation.
2. Define the result struct (total in cents, per-person shares in cents, the size of the remainder).
3. Write the validation step that returns a clear error for each invalid field.
4. Write the rounding step: compute `bill + bill * tipPercent / 100`, multiply by 100, round to the nearest integer, store as `totalCents`. This is the only money-related step that touches a floating-point value.
5. Compute `base = totalCents / people` and `remainder = totalCents % people` with integer arithmetic.
6. Build the slice of shares: the first `remainder` entries get `base + 1` cents; the remaining entries get `base` cents. The sum of the slice equals `totalCents` exactly.
7. Format the total and each share in decimal form (cents → `X.YY`).
8. Write the interactive prompt that reads three fields and prints the result.
9. Make sure a unit test can call the splitting logic with a struct literal and observe the integer-cents result without going through the prompt.

## 14. Verification Cases the Learner Must Write

Each case is a table row: bill, tip percent, people, expected total in cents, expected shares in cents, expected sum. The cases must pin the chosen policy exactly.

- One diner, zero tip: `totalCents` equals the bill in cents; the single share equals `totalCents`.
- One diner, ten percent tip on a `100` bill: `totalCents = 11000`; the single share equals `11000`.
- Three diners, bill `100`, tip `10%`: `totalCents = 11000`; shares in cents are `3667, 3667, 3666`; sum equals `11000` exactly.
- Three diners, bill `100`, tip `0%`: `totalCents = 10000`; shares in cents are `3334, 3333, 3333`; sum equals `10000` exactly. There is one cent of remainder, and it goes to the first person.
- Two diners, bill `10`, tip `0%`: `totalCents = 1000`; shares are `500, 500`; remainder is `0`.
- Four diners, bill `10.01`, tip `0%`: `totalCents = 1001`; `base = 250`, `remainder = 1`; shares are `251, 250, 250, 250` cents; sum equals `1001` exactly.
- Three diners, bill `7.25`, tip `12.5%`: total is `7.25 + 0.90625 = 8.15625`, rounded to `8.16`, so `totalCents = 816`; `base = 272`, `remainder = 0`; shares are `272, 272, 272` cents.
- Zero tip is a normal case, not an error.
- Negative bill is rejected with a clear error before any rounding or division.
- Negative tip is rejected with a clear error.
- People count `0`, `-1`, and `1.5` are each rejected with a clear error.
- A bill amount typed as `42.5` parses the same as `42.50`; both round to the same `totalCents` for the same tip.
- Sum of per-person shares equals the rounded total exactly for every table row; this is asserted in cents, not with a tolerance.

## 15. Common Mistakes to Watch For

- **Rounding the per-person share instead of the total.** The contract is "round the total once to cents, then work in integer cents". Rounding each share independently reintroduces the very float-rounding error the integer-cents rule is meant to remove.
- **Working in floats after the rounding step.** Once `totalCents` is an integer, every later step is integer arithmetic. Do not convert back to a float for the division; the integer division with remainder is the whole point.
- **Distributing the remainder to the last person.** The contract is "first `remainder` people get the extra cent". The last-person-absorbs variant is a different policy and is not the one in this README.
- **Comparing floats with `==`.** Use a small tolerance or, better, compare the integer cents that the program actually computes.
- **Silent truncation of a fractional tip.** `12.5` percent is a normal input. The internal computation must not drop the `.5`.
- **Integer overflow.** If the chosen integer type cannot hold `totalCents`, the program must say so rather than wrap silently. Pick a type whose range comfortably covers the largest expected bill-plus-tip.
- **Letting `0` people produce `NaN` or `Inf`.** Validate before dividing.
- **Promising "exact" float arithmetic.** The honest promise is "exact in integer cents", and that is what the README says.

## 16. Topics and References for Study

- A Tour of Go: "Methods", "Structs", "Errors are values".
- Effective Go: "Data", "Errors", "Package documentation".
- Package documentation: `math` (`Round`, `Floor`, `Ceil`), `fmt` (`Fprintf`, `Fscan`), `errors` (`New`, `Is`).
- IEEE-754 background: search for "float64 decimal representation", "why 0.1 + 0.2 != 0.3".
- Rounding policies: search for "round half away from zero", "banker's rounding", "round half to even".
- Money in software: search for "fixed-point arithmetic", "integer cents money representation".

## 17. Self-Assessment Questions

1. Why does `0.10 + 0.20` not equal `0.30` exactly in Go, and what does that imply for money?
2. Where in the program is the single rounding step that turns a decimal money value into an integer number of cents?
3. After that step, what numeric type carries every money value in the program, and why is that type the right one?
4. What does the integer expression `totalCents / people` compute, and what does `totalCents % people` compute?
5. How does the chosen remainder policy distribute leftover cents, and which slice indices receive the extra cent?
6. How does the test prove that the sum of shares equals `totalCents` exactly, without using a tolerance?
7. If you had to switch the rounding rule from "round half away from zero" to "round half to even", which lines would change and which would stay the same?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every functional requirement in section 8 is satisfied.
- Every verification case in section 14 has a corresponding test, and each test pins the chosen rounding rule and the "first `remainder` people get the extra cent" policy in integer cents.
- The package documentation declares the rounding rule and the remainder policy in plain English.
- The splitting logic is reachable from a test that constructs the input struct directly, without going through the interactive prompt.
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Unequal shares by weights.** Accept a list of weights (for example `[1, 1, 2]`) and compute weighted shares in cents, still under the same "round the total to cents once, then distribute the remainder deterministically" rule. Do not add a more elaborate allocation algorithm.
- **Currency symbol and locale hint.** Accept a single command-line flag that selects a currency symbol printed next to the total. Keep the parser simple: one flag, one symbol, no localization of the decimal separator.
