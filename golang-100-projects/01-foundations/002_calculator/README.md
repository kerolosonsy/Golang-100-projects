# Project 002 — Calculator

## 1. Project Name and Number
- Number: **002**, level 1 (language basics and CLI).
- Folder name in the table: **`002_calculator`**, matching `01-foundations/002_calculator/`.
- Kind: a small one-shot terminal calculator that takes two numbers and an operator, applies one of the four basic arithmetic operations, and prints the result.

## 2. Project Idea
Build a terminal program that reads two numbers and an operator, then prints the result of applying that operator to the two numbers. The required baseline covers the four basic operations, identifies an operator it does not recognise, and handles division by zero explicitly rather than producing an undefined or special value. After the program completes (whether successfully or with a reported error), it ends the run; the baseline is a single operation per run.

## 3. Why This Project Now?
- Reuses the input and printing habits from 001 and layers in two central Go ideas: small functions as logical units, and errors as explicit return values rather than exceptions.
- Introduces the switch construct as a clean way to dispatch on a small set of cases; the same shape reappears in every later project.
- Establishes the habit of checking errors rather than ignoring them, which carries into validators, parsers, and network code later.
- Reinforces the separation of computation from input and output, a theme that deepens in 003.

## 4. Prerequisites
- 001 must be complete, especially the reading-conversion-error pattern learned there.
- Environment: Go installed and on `PATH`.

## 5. What You Must Know Before Starting
- Functions returning a value and functions returning an error: understand when a function should return only a number versus when it should also return an error value, and why the latter is sometimes the only honest signature.
- Floating-point behaviour for division by zero: a floating-point division by zero does not necessarily produce a runtime error in the language; the result depends on the dividend's sign. This means the divisor must be checked explicitly before the operation.
- switch in Go: each case ends by default without explicit break statements. Understand when fallthrough behaviour is appropriate (likely not in this project).
- Operators as text: the user supplies the operator as a string; you decide how to match it (single character, word, or both).
- Two independent input validations: one to verify each operand is a number, one to verify the operator is recognised.
- Output formatting for floats: several conventions exist; the choice is yours and should suit a calculator's purpose.

## 6. Explanation of New Concepts
- Multiple return values: a Go function can return more than one value. The idiomatic pattern for "I tried to compute, but here is what went wrong" is to return both. This keeps failure visible at the call site instead of hidden in exception machinery.
- The error interface: a built-in interface that any value implementing a single method satisfies. It is how Go's standard library signals recoverable failures; understanding it removes much of the mystery around how errors propagate.
- switch as a dispatch tool: when the control flow is "given a value, branch to one of several actions", switch is clearer than a chain of if and else. Tools and readmes also treat it well.
- Type choice for numeric values: integer versus floating-point changes the meaning of the division operation in particular. Choose with intent; do not rely on whatever the language gave you by default.
- The "stop processing and explain" habit: when input is invalid, the program should respond with a clear message and end the run cleanly, never silently substituting a placeholder value for the bad input.

## 7. Learning Objective
By the end of the project you should be able to:
- Define a function that returns a numeric result plus an error value.
- Use switch to dispatch on a string.
- Distinguish integer division from floating-point division and reason about both.
- Read a short operator token from the user and match it against a known set.
- Handle the failure modes (bad operand, unknown operator, division by zero) with clear messages rather than panics.
- Organise the code into small functions instead of one large block.

## 8. Functional Requirements
- F1: The program accepts two numeric operands and an operator string.
- F2: The program applies the operator to the operands using one of the four basic arithmetic operations.
- F3: For an unknown operator (anything outside the four supported operations), the program prints a message identifying the operator as unrecognised and does not print a numeric result.
- F4: For division, the program rejects a zero divisor with an explicit message before performing the operation; it does not produce a special-value output.
- F5: For an operand that cannot be parsed as a number, the program reports the failure clearly and ends the run.
- F6: On a successful operation, the program prints the result, then ends the run.
- F7: After any outcome, the program ends the run; the baseline does not include a repeat loop.

## 9. Inputs and Outputs
**Inputs**:
- A string representing the first operand.
- A string representing the second operand.
- A short string representing the operator.

**Outputs**: text printed to standard output. Text-only examples:

- Successful addition:
  - User enters the first operand, then the second, then a recognised addition operator.
  - Program prints the result of the operation.

- Unknown operator:
  - User enters two operands and an operator outside the four supported operations.
  - Program prints a message indicating the operator was not recognised; no numeric result appears.

- Division by zero:
  - User enters two operands where the second is zero, and the division operator.
  - Program prints a message indicating division by zero is not allowed; no numeric result appears.

- Non-numeric operand:
  - User enters text that is not parseable as a number.
  - Program prints a message indicating the operand was not a valid number, and the program ends.

## 10. Rules and Edge Cases
- An operator string with extra surrounding whitespace: your normalisation policy applies before matching.
- An operator in a different case (upper versus lower): your matching policy.
- Very large or very small operands: the program must react so that the user sees a readable output rather than unreadable artefacts; the policy is yours.
- Two operands that happen to be the same number: a valid input that must not break anything.
- A request that arrives with the operands in either order: same behaviour regardless of order.

## 11. Project Constraints
- Libraries: the standard library only. The formatted I/O package and the package that supports string-to-number conversion are sufficient.
- Prohibited: any external package.
- Persistence: none. No file I/O.
- Network: none.
- Scope: only the four basic arithmetic operations. Do not extend to exponentiation, roots, trigonometry, or any other operations within this baseline.
- `panic` and `recover` are not valid tools for handling the failure modes above; use error values instead.

## 12. Design Questions Before Coding
- Do you use the same numeric type for both operands and all operations, or separate types per operation? Each choice has trade-offs in clarity and behaviour.
- Do you write one function per operation, or a single function that switches internally? What does each choice buy you in testability?
- Does your operation function return only a result, or a result plus an error? How do you decide?
- What output form fits a calculator — fixed precision, dynamic precision, or default? Each has an audience; pick one and justify it.
- Do you accept the operator only as a symbol, also as a word (such as plus), or both? Pick one consistent policy.
- After reporting an operand-parsing failure, does your program end the run immediately? The plan does not require otherwise; decide and apply it.

## 13. Implementation Milestones
1. M1: Create the source file in the project folder with the minimum required for a Go program to compile and run; verify with a build and run cycle.
2. M2: Read one operand from the terminal as text first, then perform a safe conversion, with a clear reaction when conversion fails.
3. M3: Read the second operand with the same approach.
4. M4: Read the operator as a string and match it against your supported set using a switch.
5. M5: Implement the four operation computations so each can be exercised in isolation if you choose to write tests.
6. M6: Handle division-by-zero explicitly before performing division; react with a clear message.
7. M7: Handle unknown operators explicitly; react with a clear message.
8. M8: Print the result of a successful operation in your chosen form.
9. M9: Run every verification scenario in section 14 and confirm the documented behaviour.

## 14. Verification Cases the Learner Must Write
- Each of the four operators produces the mathematically correct result on representative inputs.
- An unknown operator symbol is rejected without producing a numeric result.
- Division by zero is rejected without producing a numeric or special value.
- A non-numeric operand is rejected with a clear message, and the program ends.
- An operator symbol with surrounding whitespace is treated as the same operator after normalisation.
- An operator in a different case is treated per your matching policy.
- A negative operand combined with each operation produces the correct result.
- An operand of zero with a non-division operator produces the correct result (for example, zero added to a number).
- Inputs that mix integers and decimals are handled per your chosen numeric type.

## 15. Common Mistakes to Watch For
- Confusing integer division with floating-point division; the difference changes the result of division drastically for non-trivial inputs.
- Letting division by zero yield a special value silently rather than checking the divisor explicitly before the operation.
- Using exception-like control flow to handle operator or operand errors instead of returning error values.
- Inconsistent input normalisation around the operator: trimming or case-folding sometimes but not always.
- Distributing the conversion logic across multiple call sites instead of one helper.
- Inconsistent error-message style across the three failure modes.
- Relying on default print formatting that is unreadable for chosen numerical ranges.

## 16. Topics and References for Study
- Effective Go: Multiple returns.
- A Tour of Go: switch.
- The official documentation for the standard error helpers and the formatted I/O package.
- The official documentation for the string-to-number conversion package.
- The official documentation for the math package, especially the section on detecting special float values.
- Search terms: `Go multiple return values`, `Go switch fallthrough`, `Go float division by zero`, `Go strconv number parsing`.

## 17. Self-Assessment Questions
1. Why does idiomatic Go prefer returning an error value over throwing an exception? Give a concrete benefit in this project.
2. For the division operator, what is the difference in result between integer division and floating-point division? Give an example from this project.
3. Why does the default switch in Go not require an explicit `break` at the end of each case?
4. Why must the program check the divisor before dividing, instead of letting the runtime produce whatever it produces?
5. How do you decide whether to accept the operator as a symbol only, a word only, or both? What is consistent in this project?
6. If you later add a new operation, what changes do you make in the program structure, and what stays the same?

## 18. Definition of Completion
- The program compiles and runs without compile errors.
- Each of the four basic operations is implemented and produces the correct result on representative inputs.
- Unknown operator and division-by-zero are handled with clear messages, not with crashes.
- A non-numeric operand is handled with a clear message and the program ends cleanly.
- The code is split into small functions with obvious responsibilities.
- You can explain why you chose a numeric type, why you return an error, and how your switch matches.

## 19. Optional Extensions
- Optional 1: Wrap the operation in a repeat-calculation loop so the user can run multiple operations in one program run. Each iteration prints either the result or the matching error message, and the loop ends when the user indicates they are done.
- Optional 2: Add one additional explicitly chosen operation (such as modulo, exponentiation, or square root) along with its matching error cases. For example, modulo by zero and square root of a negative number should each produce a clear, operation-specific message rather than a generic one.
