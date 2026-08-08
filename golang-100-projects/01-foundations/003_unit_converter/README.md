# Project 003 — Unit Converter

## 1. Project Name and Number

- Number: **003**, level 1 (language basics and CLI).
- Folder name in the table: **`003_unit_converter`**, matching `01-foundations/003_unit_converter/`.
- Kind: a terminal program that converts a value from one unit to another across three categories: temperature, length, and weight. Conversions run in both directions within the chosen baseline pairs.

## 2. Project Idea

Build a terminal program that converts a value from one unit to another. The required baseline supports exactly one small unit pair per category: a temperature pair that involves both scaling and offset, a length pair that is linear scaling, and a weight pair that is linear scaling. Both directions of each pair must work. The conversion logic must be separated from input and output so that it can be tested in isolation. The program rejects an unknown category, an unknown unit, an unparseable value, and a value that is physically unreasonable for the category. Verification uses a tolerance, never direct float equality.

## 3. Why This Project Now?

- Generalises the calculator pattern (operate on numbers, surface errors) into value transformations that vary by category.
- Introduces named constants as an alternative to magic numbers spread through the code.
- Highlights the difference between linear conversions (scaling by a single factor) and offset conversions (scaling and translation), useful any time a project touches measurement.
- Sets up tolerance-based testing, a habit reused in any project where float equality is rarely exact.

## 4. Prerequisites

- 002 must be complete, especially the separation of computation from input and the error-handling pattern.
- Environment: Go installed and on `PATH`.

## 5. What You Must Know Before Starting

- Named constants: a way to give a name to a numeric constant and to keep magic numbers out of the body of code. Decide where each constant lives for readability.
- Floating-point numerics: 32-bit versus 64-bit floating-point types, and why one is the conventional default in Go for tasks that need precision.
- Floating-point precision: some decimal values do not have an exact binary representation. Direct equality on float results is usually wrong; a tolerance is the correct approach.
- Pure functions: a function that depends only on its inputs and produces a result without side effects is easier to test. Plan your conversion routines to be pure.
- Linear versus offset conversions: scaling the value by a single factor versus applying both a scale and an offset. The typical offset case appears in temperature conversions between common scales.
- Categories with different physical meaning: temperature has a reference point tied to physical phenomena; length and weight have natural zeros. Negative values are unphysical for length and weight (your design choice) but possible for temperature (also your design choice).

## 6. Explanation of New Concepts

### Concepts

- Conversion factors as named constants: a constant near the top of the file is more readable than a long floating-point number repeated throughout the body.
- The asymmetry of temperature: length and weight pairs usually relate by a single factor; some temperature pairs also include an offset, which makes the routine structurally different. Recognise this difference early.
- Tolerance in verification: tests that compare two floating-point results with direct equality are usually wrong; an explicit tolerance around the expected value is the standard practice in both mathematics and Go testing.
- Layered design: keep input parsing, unit selection, conversion logic, and output formatting in separate routines. This is what makes conversion logic testable without a terminal.
- The role of categories: each category (temperature, length, weight) needs its own conversion routine. A small dispatching routine picks the right one. This anticipates the dispatch style used in menu-driven projects later.

## 7. Learning Objective

By the end of the project you should be able to:
- Define conversion routines that take a numeric value and return a numeric value, with no terminal interaction inside them.
- Express conversion factors as named constants, not as magic literals.
- Recognise which conversion pairs are linear-only and which require an offset.
- Compose a small dispatching function that selects the right conversion logic given a category and a pair of units.
- Use a tolerance when comparing expected and actual conversion results.
- Handle an unknown unit or an unparseable input with a clear message rather than a crash.

## 8. Functional Requirements

1. F1: The program supports exactly one small unit pair per category in the baseline. Temperature: a pair whose conversion involves both a scale and an offset (Celsius and Fahrenheit is the natural baseline). Length: a linear-scale pair. Weight: a linear-scale pair. For each pair, conversions in both directions must be available. Document your chosen pair for each category in your program.
2. F2: The program reads a category choice, a source unit, a target unit, and a numeric value from the terminal.
3. F3: The program applies the chosen conversion and prints the result.
4. F4: An unknown category is rejected with a clear message.
5. F5: An unknown unit for a known category is rejected with a clear message.
6. F6: A value that cannot be parsed as a number is rejected with a clear message.
7. F7: A value that is physically unreasonable for the category (your design choice) is rejected with a clear message.
8. F8: Negative inputs follow your chosen policy per category: accepted in temperature (with whatever physical bound you choose), and rejected or accepted in length and weight according to the policy you document.
9. F9: After the conversion is reported, the program ends the run; the baseline does not include a repeat loop.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- A category selection in the format you choose.
- A source unit within the chosen category.
- A target unit within the chosen category.
- A numeric value.

#### Outputs

Text to standard output. Text-only examples:

- A baseline length conversion, source to target:
  - User selects the length category.
  - User selects the source unit and a different target unit.
  - User enters a positive numeric value.
  - Program prints a line naming the source value in source units and the converted value in target units.

- Unknown category:
  - User enters a category identifier the program does not recognise.
  - Program prints a message indicating the category is unsupported.

- Unknown unit:
  - User selects a valid category and a unit that the program does not know within that category.
  - Program prints a message indicating the unit is not part of the chosen category.

- Negative input where not appropriate:
  - User selects length and a negative value.
  - Program prints a message indicating the value is not valid for the category.

## 10. Rules and Edge Cases

- Negative values: accepted in categories where they are physically meaningful and rejected where they are not, per your chosen policy.
- A unit symbol written in upper-case versus lower-case letters: your matching policy.
- A unit that shares a symbol across categories: rejected because the category context is required for the chosen pair, not because the symbol itself is invalid.
- A value that approaches the boundary of float precision (very large or very small): handled so the user sees readable output.
- Empty input where a value is expected: rejected with a clear message.

## 11. Project Constraints

- Libraries: the standard library only. The formatted I/O package, the string-to-number conversion package, and (if you choose) the math package for rounding helpers are sufficient.
- Prohibited: any external package.
- Persistence: none. No file I/O.
- Network: none.
- Scope: exactly one small pair per category in the baseline. Optional expansions belong to section 19.
- No hardcoded output precision is required; the choice is yours and should suit the conversions.
- After the conversion is reported, the program ends; the baseline does not include a repeat loop.

## 12. Design Questions Before Coding

- Which baseline unit pair will you support per category? Justify the choice with readability and with the offset-versus-linear distinction.
- Will you write separate conversion routines for each direction, or one routine that detects direction? What is the cost of each in testability?
- Where in the code do the conversion factors live? As named constants grouped by category? In a separate file? Justify the choice.
- Should the program accept same-unit-to-same-unit conversion as a recognised case, or only accept cases where the source and target differ? If the former, what wording makes the situation clear? The plan does not require either; both are acceptable as design choices.
- How will you represent negative-value policy per category? As a helper that validates ranges? Where does the helper belong?
- How do you choose the tolerance for each category? Is a single tolerance acceptable, or do different categories need different values?
- Will you let the user pick the category by name, by number, or both? Make it consistent.

## 13. Implementation Milestones

1. M1: Create the source file in the project folder with the minimum required for a Go program to compile and run; verify with a build and run cycle.
2. M2: Define the named constants for the baseline conversion factors and any offsets the baseline needs.
3. M3: Implement the conversion logic for each baseline pair, concentrating all conversion arithmetic in dedicated routines that do not interact with the terminal.
4. M4: Implement a small dispatcher that selects the right conversion given the category and the pair.
5. M5: Implement terminal prompts to read category, source unit, target unit, and value, applying your normalisation policy.
6. M6: Implement the rejection paths for unknown category, unknown unit, unparseable value, and physically unreasonable value.
7. M7: Add the output line that reports the source value and the converted value.
8. M8: Compose and run a tolerance-based check that matches section 14, demonstrating the expected precision for each direction.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Each baseline direction produces a value within your tolerance of the canonical expected result.
- Unknown category is rejected with a clear message.
- Unknown unit within a known category is rejected with a clear message.
- Unparseable numeric value is rejected with a clear message.
- A physically unreasonable value (your category-specific policy) is rejected.
- A negative input is handled according to your chosen per-category policy.
- An input with surrounding whitespace is treated consistently.
- A unit symbol written in upper-case is treated per your matching policy.
- A tolerance check compares the produced value against the canonical expected value, never using direct float equality.

## 15. Common Mistakes to Watch For

- Comparing two floating-point results with direct equality; always use a tolerance.
- Forgetting that one temperature conversion requires an offset while the other categories in your baseline do not.
- Pre-rounding inside the conversion routine, which loses precision for downstream transformations.
- Letting negative-value handling drift between categories.
- Letting unknown-unit handling crash on a lookup miss instead of producing a clear message.
- Embedding terminal reads inside the conversion routines, which prevents pure-function testing.
- Reusing the same absolute tolerance across very different magnitude categories (tolerance is category-dependent).
- Hardcoding an output precision that doesn't match the magnitudes produced by your conversions.

## 16. Topics and References for Study

- The official Go documentation pages for the formatted I/O package, the string-to-number conversion package, and the math package.
- The IEEE 754 representation, briefly: why decimal fractions like one tenth often have no exact binary representation in floating point, so direct equality on computed values is usually wrong.
- The formulas for the baseline conversion pairs you choose.
- General references: Effective Go on constants; A Tour of Go on numerical types.
- Search terms: `Go float tolerance`, `Go named constants group`, `temperature conversion formula`, `length weight conversion factor table`.

## 17. Self-Assessment Questions

1. Why must the conversion routines be free of terminal interaction if you want to test them with a tolerance-based check?
2. How would you recognise a conversion that needs both a scale and an offset versus one that only scales?
3. Why does the baseline include both directions for each pair? What breaks if one direction is missing?
4. What tolerance would you choose for a baseline check, and why? Would you change it for a different category?
5. If you add a new category later (for example, time), what stays the same in the program structure and what changes?
6. Why is per-category negative-value policy better than a uniform "reject negative" rule for all three categories?

## 18. Definition of Completion

- [ ] The program compiles and runs without compile errors.
- [ ] Each baseline conversion direction produces the expected result within tolerance.
- [ ] Each rejection path (unknown category, unknown unit, unparseable value, physically unreasonable value) produces a clear message and does not crash.
- [ ] The conversion routines are pure, with no terminal reads or prints inside them.
- [ ] You can explain why the chosen baseline pair is sufficient, why the offset applies where it does, and why your tolerance is appropriate.

## 19. Optional Extensions

- Optional 1: Add another unit per category to your baseline (for example, a third temperature unit), keeping the dispatch and tolerance pattern intact.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisite:** [Project 002 — Calculator](../../01-foundations/002_calculator/README.md#20-prerequisite-based-documentation-guide).

Read the linked guide first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **Standards and concept references:** [Go specification: floating-point types](https://go.dev/ref/spec#Floating-point_types).

### Project-specific learning focus

- **Learn now:** named conversion constants, dimensional analysis, precision and rounding, and tolerance-based tests.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
