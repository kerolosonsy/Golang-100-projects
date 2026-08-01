# Project 007 — BMI Calculator

## 1. Project Name and Number
- Number: **007**, level 1 (language basics and CLI).
- Folder name in the table: **`007_bmi_calculator`**, matching `01-foundations/007_bmi_calculator/`.
- Kind: a small terminal program that asks for height and weight, computes a body-mass index, classifies the result into a category, and prints a clear, rounded value. The program is for educational practice, not medical advice.

## 2. Project Idea
Build a terminal program that prompts for height and weight, validates that both are positive numbers in their respective units, computes the index using the standard formula, and prints the value along with a category drawn from a documented set of thresholds. The required baseline covers: read two inputs, parse them to floating-point values, validate positivity, compute the index, classify it, and present a rounded result without losing intermediate precision.

## 3. Why This Project Now?
- First project in the path that uses `float64` end-to-end, exposing the usual pitfalls of floating-point arithmetic: rounding decisions, equality comparisons, and display formatting.
- Reuses the input-validation habit from 001, 002, and 014 (already a habit even though 014 is later in the path), but adds a numeric range check rather than just a parse check.
- Introduces the pattern of separating pure computation from input and output, a separation that becomes structural from project 011 onward.
- Sets the stage for later projects that reason about thresholds and categories (rate limiter in 036, search ranking in 069, scheduler priorities in 090).

## 4. Prerequisites
- Project **006** (`006_string_reverser`). Comfort with functions, slices, and clean input handling should already be in place.
- Familiarity with `int` and the basic distinction between integer and floating-point arithmetic.
- No new tools or libraries beyond the standard library used in earlier projects.

## 5. What You Must Know Before Starting
- The `float64` type is a 64-bit IEEE 754 binary floating-point value. Most decimal fractions cannot be represented exactly, and any computation that would produce an exact decimal result will produce a nearby binary approximation instead.
- The body-mass index used here is `weight / (height * height)` with weight in kilograms and height in metres. The formula is the same worldwide; the categories are conventional thresholds, not physiological truths.
- A common display convention is one or two decimal places. Rounding is a presentation decision; the underlying value should not be re-rounded at every step of the computation.
- Equality on floating-point values is risky: comparing `==` is almost always wrong. A small tolerance is the usual pattern, but for this project the right comparison is between the computed value and a category threshold, which is itself a `float64`.
- Parsing a string into a `float64` has the same shape as parsing into an integer: the conversion can fail, and the failure must be reported.
- A category boundary is itself a real number. The conventional rule assigns the boundary value itself to the upper band, so 18.5 is the first value counted as normal, 25 is the first value counted as overweight, and 30 is the first value counted as obese. Apply this rule consistently across the table.
- This program is educational. It does not diagnose, treat, or recommend anything. The wording around the result must not claim medical certainty.

## 6. Explanation of New Concepts
- The standard library's `strconv.ParseFloat`: where it lives, what error it returns, and how to react to it. Treat any parse failure as invalid input, not as a value of zero.
- The `fmt` verb families for floating-point output: some default to scientific notation, some to fixed-point, some to "shortest representation that round-trips". Pick a verb whose defaults match the rounding convention you documented.
- Floating-point rounding policy: rounding only when printing, not when computing, prevents compounding rounding error. The intermediate value is kept at full precision; the displayed value is rounded.
- Threshold comparison on floating-point values: comparing `x <= threshold` is the usual pattern for a category that includes the threshold, and `x < threshold` for one that excludes it. Pick one rule per boundary and apply it consistently.
- Numeric range validation: a value can parse correctly and still be nonsensical (zero, negative, or absurdly large). A parse check alone is not enough; a range check follows it.
- Program boundaries: the part of the program that parses input, the part that computes, and the part that prints are three distinct concerns. Keeping them in separate functions makes each one independently testable.

## 7. Learning Objective
By the end of the project you should be able to:
- Read two numeric inputs from the terminal and convert them to `float64` safely.
- Apply a documented positivity check to both values and reject zero or negative values with a clear message.
- Compute the index using the standard formula and classify the result using a documented threshold table.
- Round the value only when printing, keeping the intermediate computation at full precision.
- Explain, in plain language, why the program is an educational exercise and not a diagnostic tool.
- Distinguish three concerns: input parsing, computation, and presentation.

## 8. Functional Requirements
- F1: The program prompts for and reads a height in metres and a weight in kilograms. The wording of the prompts and the order of the prompts is your choice; the values themselves are unambiguous.
- F2: Both inputs are parsed to `float64`. A parse failure is handled with a clear message; the program does not silently treat invalid input as zero.
- F3: Both values must be strictly positive. Zero, negative, or non-finite values (infinity, NaN) are rejected with a clear message before any division takes place.
- F4: The index is computed as `weight / (height * height)`. The computation uses `float64`. The intermediate value is not rounded.
- F5: The computed value is classified into one of four categories: underweight (BMI strictly less than 18.5), normal (BMI at least 18.5 and strictly less than 25), overweight (BMI at least 25 and strictly less than 30), and obese (BMI at least 30). Exactly 18.5 is normal, exactly 25 is overweight, and exactly 30 is obese.
- F6: The program prints the computed value, rounded to a documented number of decimal places, followed by the category. The category wording is your choice but must be consistent with the threshold table.
- F7: The wording around the result makes it clear that this is an educational calculation, not a medical diagnosis.

## 9. Inputs and Outputs
**Inputs**:
- Height in metres, as a decimal number. Typical educational values are between 0.5 and 2.5 metres.
- Weight in kilograms, as a decimal number. Typical educational values are between 10 and 300 kilograms.
- Both inputs may be integers in textual form ("1" rather than "1.0"); both must be accepted.

**Outputs**: text printed to standard output. Text-only examples:

- A normal adult, height 1.75 m, weight 70 kg:
  - Program prints a single line containing the computed value, rounded to one or two decimal places as you documented, and the corresponding category.
  - The category matches the threshold table in section 10.

- A value exactly on a boundary, such as BMI exactly 25.0: by the conventional rule a BMI of exactly 25.0 belongs to the "overweight" band. The category in your program's output matches this rule; the rule is documented in this README and applied consistently in the program.

- A value below the lowest threshold, such as a clearly underweight example:
  - Program prints the rounded value and the corresponding "underweight" category.

- A value above the highest threshold, such as a clearly obese example:
  - Program prints the rounded value and the corresponding "obese" category.

- Height entered as `0`:
  - Program rejects the input with a clear message; no division is performed.

- Weight entered as a non-numeric string such as `seventy`:
  - Program rejects the input with a clear message; no division is performed.

## 10. Rules and Edge Cases
- Height and weight must both be strictly positive. Zero, negative, and non-finite values are invalid.
- Parse failure is invalid input. The program does not coerce invalid input to zero or to any other value.
- The threshold table is, in the version used here, the conventional four-band table with thresholds at 18.5, 25, and 30. The bands are defined explicitly as:
  - Strictly less than 18.5: underweight.
  - 18.5 or greater, and strictly less than 25: normal.
  - 25 or greater, and strictly less than 30: overweight.
  - 30 or greater: obese.
  By this rule, exactly 18.5 is normal, exactly 25 is overweight, and exactly 30 is obese.
- The intermediate computation is not rounded. Rounding is applied once, when the value is printed.
- Display precision is your documented choice (one or two decimal places). The same choice is used for every output line.
- NaN and ±Inf in the inputs (where parseable as a `float64`) are treated as invalid; the program does not print "NaN" or "Inf" as a result.
- The output never claims a medical diagnosis. Wording such as "according to the conventional thresholds", "educational result", or "for practice only" satisfies this requirement; wording such as "you are underweight" used in a clinical sense does not.

## 11. Project Constraints
- Libraries: the standard library only. `strconv.ParseFloat` for parsing, `fmt` for printing with a fixed-point verb, `math` if you choose to use it for explicit NaN/Inf checks.
- Prohibited: any external package. No third-party BMI library; the formula is two lines of arithmetic.
- Persistence: none. The program reads, computes, prints, and exits.
- Network: none. No ports, no requests.
- Tests: optional in code; the verification section below lists scenarios the learner runs manually or as table-driven tests if tests are added.

## 12. Design Questions Before Coding
- Will you keep the parsing, the computation, and the printing in three separate functions, or will you combine some of them? Three separate functions is the easier-to-test design.
- How will you represent a category? A string is the simplest representation. An integer constant per category, or a typed enum-like type, has its own advantages; the choice is yours.
- Where will the threshold table live? As a slice of `(upperBound, category)` pairs, as a `switch` over ranges, or in another shape? Each shape has different readability and testability trade-offs.
- How many decimal places will you print? One decimal place is conventional for body-mass index. Two decimal places is acceptable. Document the choice.
- What wording will you use around the category so the educational nature of the result is unmistakable? Reserve clinical-sounding language for real medical software.
- How will you handle inputs that parse correctly but are nonsense, such as `1e308` for the weight? A range check is part of the validation; without it, parse-only validation accepts absurd values.

## 13. Implementation Milestones
1. M1: Create the source file with the minimum required for a Go program to compile and run; verify with a build and run cycle.
2. M2: Read height and weight as text and parse each into a `float64`. Confirm that both valid and invalid inputs are handled.
3. M3: Validate that both values are strictly positive and finite. Reject zero, negative, NaN, and ±Inf with a clear message.
4. M4: Compute the index as `weight / (height * height)`. Keep the intermediate result at full precision; do not round during computation.
5. M5: Add a documented threshold table and a classification step that uses one consistent rule for boundaries.
6. M6: Print the rounded value and the category. Apply the same rounding rule on every output line.
7. M7: Add an explicit non-medical wording to the output. Verify that no line claims a clinical diagnosis.
8. M8: Run every verification scenario from section 14 and confirm the program behaves as your design specifies.

## 14. Verification Cases the Learner Must Write
- A normal adult: height 1.75 m, weight 70 kg. The computed value, rounded to your documented precision, is in the conventional "normal" band.
- A clearly underweight example: height 1.75 m, weight 50 kg. The category is the lowest band in your table.
- A clearly overweight example: height 1.70 m, weight 80 kg. The category matches the upper bands of your table.
- A value exactly on a boundary between two categories: exactly 18.5 is normal, exactly 25 is overweight, and exactly 30 is obese. The verdict matches the rule stated in section 10 and is the same on every run for the same input.
- A value just below the boundary and a value just above the boundary: the two verdicts differ in the way your table specifies.
- Height of `0`: rejected with a clear message; no division by zero, no panic.
- Weight of `0`: rejected with a clear message; no division by zero, no panic.
- Negative height or negative weight: rejected with a clear message.
- Non-numeric height or weight (for example `seventy`): rejected with a clear message.
- A height such as `0.000001` and a weight such as `1e308`: rejected by the range check; the program does not overflow silently and does not print NaN or Inf.
- The same input run twice: the same rounded output and the same category. The computation is deterministic.
- Output wording: contains an explicit non-medical framing; contains no claim of clinical diagnosis.

## 15. Common Mistakes to Watch For
- Rounding the intermediate value as soon as it is computed. This compounds the rounding error across the formula. Round once, when printing.
- Comparing a `float64` to another `float64` with `==`. The threshold tables do not require equality comparisons; they require ordering comparisons. Use `<`, `<=`, `>`, `>=`, and document the boundary rule.
- Treating invalid input as zero. Zero is a valid value to reject; it is not a valid value to assume silently. A parse failure must produce a clear message and a non-result.
- Forgetting the range check. A parse check accepts `1e308` for the weight and divides by a tiny height, producing an infinite or absurdly large result. The range check exists precisely to prevent this.
- Choosing a different precision for different output lines. One rule, applied consistently.
- Using clinical-sounding wording for an educational result. The program must not pretend to diagnose.
- Letting the threshold table drift from the documentation. The thresholds in the code and the thresholds in the README must match. Pick a single source of truth.
- Returning a category as a bare string without a type. A typed category makes accidental misuse harder, but if you choose a bare string, document the set of values you use.

## 16. Topics and References for Study
- A Tour of Go: Type inference, basic types, and the introduction to floating-point numbers.
- Effective Go: Names, control structures, and the section on printing.
- The `strconv` package documentation: `ParseFloat` and its error reporting.
- The `fmt` package documentation: verbs for floating-point values and their precision and width modifiers.
- The Go blog or specification: a short note on IEEE 754 and on floating-point comparison pitfalls.
- The conventional body-mass index thresholds: confirm the four-band table (underweight, normal, overweight, obese) with the conventional boundary values 18.5, 25, and 30.
- Search terms: `Go ParseFloat error handling`, `Go fmt float precision`, `float64 comparison pitfalls`, `BMI category thresholds conventional`.

## 17. Self-Assessment Questions
1. Why is rounding only the printed value safer than rounding the intermediate result? Describe the failure mode that the project avoids by rounding once.
2. A user enters `1.70` for the height and `70.0000001` for the weight. The computed value, in `float64`, is very close to a boundary. Walk through what your program does, and explain whether your boundary rule can be defeated by a value that is one ulp away.
3. The conventional rule in section 10 says exactly 25 belongs to "overweight". A reviewer suggests instead that exactly 25 should be classified as "normal", arguing that the boundary should belong to the lower band. What changes in the program, what stays the same, and which alignment with section 10 is the right one for this project?
4. Why is treating an unparseable string as `0` for the height or weight a bug, not a convenience?
5. Your program accepts height `2.5` and weight `600`. The computation completes; the result is a real number; the category is the highest band. Should the program reject the input instead? Defend your answer in terms of the project's scope and the educational nature of the result.
6. The program prints "you are normal weight" without any qualifier. Why is this wording inappropriate for an educational tool, and what wording would you replace it with?
7. Walk through the case where the user enters `0` for the height. Where in your program does the rejection happen, and what does the user see?

## 18. Definition of Completion
- The program compiles and runs without compile errors.
- Every scenario in section 14 produces the behaviour documented in your code.
- No panic occurs in any documented scenario, including division-by-zero and absurd input.
- The threshold table in code matches the threshold table in this README, with exactly 18.5 normal, exactly 25 overweight, and exactly 30 obese.
- The intermediate computation is not rounded; rounding happens only at print time.
- The output includes an explicit non-medical framing; no clinical diagnosis is claimed.
- You can explain, in plain language, why the program is educational and not diagnostic.

## 19. Optional Extensions
- Optional 1: Accept either metric or imperial inputs (centimetres, or feet and inches, for height; kilograms or pounds for weight) and document the conversion rule. The category rule still uses the same body-mass index thresholds.
- Optional 2: Print a short, plain-language description of where the computed value sits between the two nearest thresholds, expressed in your own words. The description is educational and explicitly non-medical.
