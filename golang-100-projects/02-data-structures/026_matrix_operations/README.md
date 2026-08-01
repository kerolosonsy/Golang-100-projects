# Project 026 — Matrix Operations

## 1. Project Name and Number

Project **026** — `026_matrix_operations`. The directory name and number must match exactly. The project works with rectangular `float64` matrices and supports addition, multiplication, and transpose. Inputs and outputs are two-dimensional `float64` slices. Operations return independent results, never mutate or alias the input rows, and never panic on invalid dimensions. Every invalid input is rejected with a contextual error.

## 2. Project Idea

The package models a matrix as a rectangular `[][]float64`: a slice of rows, each row a slice of columns. All three operations — add, multiply, transpose — produce a fresh matrix whose rows are independent slices; the result is decoupled from both inputs. Empty and ragged inputs are rejected with a contextual error rather than silently treated as zero rows. Equality of dimensions is checked before any element is read, so a mismatched shape is reported cleanly instead of producing a partial result. Floating-point comparisons in tests use a fixed tolerance; dimension checks use exact equality. Negative, decimal, and very large or very small `float64` values are valid inputs. The project never panics on an invalid shape.

## 3. Why This Project Now?

Projects 001–025 established variables, functions, loops, structs, errors, slices, files, JSON, CSV, scanning, sorting, walking, and hashing. None of them required structured numeric work with shape discipline. Project 026 introduces the first project in which the layout of a slice-of-slices is part of the contract. The learner must treat a matrix as a typed shape with rows and columns, not as "a slice of slices of numbers that happens to exist".

This is also the first project that forces the learner to be honest about shape: a ragged slice is not "a column count that happens to be zero", and an empty outer slice is not "a matrix with zero rows". Project 026 pins the rejection rule that subsequent projects in the path rely on.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 026 therefore requires:

- Completion of **025** (File Duplicate Finder). Earlier projects (for example 020's safe-walk pattern, 019's streaming discipline, 016's slice discipline) are background concepts already encountered and may inform the learner's design, but they are not additional required completions for this project.
- No prior knowledge of generics, HTTP, databases, or concurrency.

## 5. What You Must Know Before Starting

- That a `[][]float64` is a slice of row slices, not a contiguous rectangular block. Each row is its own slice with its own length, capacity, and backing array. Two rows can share the same backing array; aliasing a row means the result shares storage with an input.
- That `len(m)` is the row count and `len(m[i])` is the column count of row `i`. The two are independent. A matrix that claims to be 3×4 but whose rows have lengths `4, 4, 3` is ragged and not a valid matrix. A matrix whose every row has length `0` is empty (not ragged) and is also not a valid matrix.
- That an empty outer slice (`len(m) == 0`) has no rows and therefore no column count. Treating it as a zero-row matrix with a known width would be guessing.
- That `append` may reallocate a slice's backing array, which means a row that is grown with `append` no longer aliases the original backing array; a row that is taken by index and used directly does alias.
- That `float64` arithmetic is approximate. Equality comparisons in tests must use a tolerance, while dimension comparisons must use exact equality.
- That returning an `error` is the standard way to signal an invalid shape. The project does not recover from panics and does not pretend an out-of-range index is fine.
- That the built-in package `errors` (and `fmt.Errorf` with `%w` for wrapping) is enough for the errors this project produces. No third-party error library is needed.

## 6. Explanation of New Concepts

### Rectangular matrix as a shape, not a bag of rows

A matrix is a `[][]float64` whose rows all have the same length and whose row count is well-defined. The pair `(row count, column count)` is the shape. Every operation in this project is defined on shapes, not on slices: the result of adding two matrices exists only when the two shapes match exactly. The project pins shape as a contract so a later caller cannot rely on "it worked because the rows happened to line up".

### Empty matrices: zero rows and zero columns

A matrix is empty when it has zero rows, when every row has zero columns, or both. Both states are rejected with an "empty matrix" error. The project does not treat zero rows as "a matrix with zero rows and some known column count" — there is no usable column count to recover. The project does not treat every row having zero columns as a usable matrix either — a zero-column row cannot be added to, transposed, or multiplied against without inventing shape. Both states are empty; the error identifies "empty matrix" and names the offending operand.

### Rejecting ragged matrices

A ragged matrix is one whose rows have differing positive lengths, or whose rows mix zero and non-zero lengths. Ragged means "row lengths differ". A matrix whose rows are all length `0` is not ragged; it is empty. A matrix whose rows are all length `4` is not ragged. A matrix whose first row has length `4` and second row has length `3` is ragged and is rejected with an error that identifies "ragged matrix" and names the offending row index and the row length that differs.

### Adding two matrices

Two matrices can be added element-wise when their shapes are equal. The result has the same shape and contains element-wise sums. Inputs are not mutated; the result rows are independent slices. Shape mismatch produces a contextual error without reading any element.

### Multiplying two matrices

Two matrices can be multiplied when the left column count equals the right row count. The result has `left rows` rows and `right columns` columns. Each result element is the sum of the products of the left row and the right column at that position. Inputs are not mutated; the result rows are independent slices. Shape mismatch produces a contextual error without reading any element.

### Transposing a matrix

Transposing swaps rows and columns. A matrix with `r` rows and `c` columns becomes a matrix with `c` rows and `r` columns. The element at `(i, j)` in the result is the element at `(j, i)` in the input. The result rows are independent slices; the original input is not mutated. Transposing twice returns a matrix with the original shape and the original values, element-for-element.

### Independent results, never aliases

Every operation returns a fresh matrix whose rows are independent slices. The contract is verified behaviorally, not by inspecting internals. A test mutates a result cell and then re-reads the corresponding input cell: the input cell is unchanged. A test mutates a result cell in one row and then re-reads a different result row: the other row is unchanged. A test mutates an input cell and then re-reads a previously returned result: the result is unchanged. The behavioral test pins independence without depending on any particular memory layout, reflection trick, or pointer-identity expression.

### 1×1, row vector, column vector, and mixed-sign or decimal values

A 1×1 matrix is a single number; it transposes to itself and adds or multiplies with any other 1×1. A row vector (1×N) transposes to a column vector (N×1); the two have different shapes. A column vector (N×1) transposes to a row vector (1×N); the two have different shapes. Negative and decimal `float64` values are valid inputs and outputs. Tests must include negative values, decimals, and zeros in both inputs and expected results.

### Floating-point tolerance vs exact dimensions

Tests assert dimensions with exact equality and values with a fixed tolerance. The tolerance is a single number pinned by the project (for example an absolute tolerance appropriate for the operations). The tolerance is applied to every expected value, including negative, decimal, zero, and very-large or very-small results.

## 7. Learning Objective

After completing this project the learner can:

- Model a rectangular matrix as a `[][]float64`, treat its shape `(rows, columns)` as a contract, and reject empty or ragged inputs with a contextual error before reading any element.
- Implement matrix addition that requires exact shape equality, produces an independent result, and reports shape mismatch as an error without mutating the inputs.
- Implement matrix multiplication that requires the left column count to equal the right row count, produces a result with `left rows` rows and `right columns` columns, and reports shape mismatch as an error.
- Implement transpose that swaps rows and columns, produces an independent result, and preserves the original input.
- Explain why transposing twice returns the original shape and values, element-for-element.
- Reason about row independence: confirm that result rows do not alias input rows or each other through behavioral observation (mutate one and observe the others), without depending on internal layout or pointer inspection.
- Treat 1×1, row vector (1×N), column vector (N×1), negative, and decimal inputs as ordinary valid inputs.
- Use a fixed tolerance for `float64` value comparisons and exact equality for dimension comparisons in tests.
- Never panic on invalid dimensions: every out-of-shape path is an error return, not a runtime panic.
- Write tests that pin valid shapes, invalid shapes, empty (zero-row and zero-column) inputs, ragged inputs, identity and zero matrices, transpose-twice, non-mutation, and behavioral independence between inputs and results.

## 8. Functional Requirements

1. The package exposes operations that accept matrices as `[][]float64` and return matrices as `[][]float64`, returning an `error` on invalid shape.
2. Add accepts two matrices. Their shapes must be equal. The result has the same shape as the inputs. Each result element is the sum of the corresponding input elements. Inputs are not mutated. The result rows are independent slices. Shape mismatch returns a contextual error naming the two shapes and identifying the offending operand.
3. Multiply accepts two matrices. The left column count must equal the right row count. The result has `left row count` rows and `right column count` columns. Each result element is the dot product of the corresponding left row and right column. Inputs are not mutated. The result rows are independent slices. Shape mismatch returns a contextual error naming the shapes.
4. Transpose accepts one matrix. The result has the transposed shape: `c` rows and `r` columns for an input of `r` rows and `c` columns. Each result element is the input element at the swapped position. The input is not mutated. The result rows are independent slices.
5. Empty matrices are rejected for add, multiply, and transpose with an error identifying "empty matrix" and naming the empty operand. Empty covers zero rows, every row having zero columns, and zero rows with every row having zero columns. The project does not invent a column count for a zero-row input, does not invent a row count for a zero-column input, and does not treat "every row has zero columns" as a usable shape.
6. Ragged matrices (rows of differing non-zero lengths, or rows mixing zero and non-zero lengths) are rejected for add, multiply, and transpose with an error identifying "ragged matrix" and naming the offending row index and its actual length. A matrix whose rows are all length `0` is not ragged; it is empty. A matrix whose rows all share a positive length is not ragged. The project does not silently treat the longest row or the first row as the column count.
7. Operations never panic on invalid dimensions. Every invalid path returns an error. Tests that would panic in a careless implementation (for example a multi-dimensional index out of range) are part of the verification set.
8. Operations never mutate the input rows. The behavioral independence test mutates a result cell and re-reads the corresponding input cell: the input cell is unchanged.
9. Operations never mutate the result rows in response to later input mutations. The behavioral independence test mutates an input cell and re-reads a previously returned result cell: the result cell is unchanged.
10. Operations never propagate a mutation of one result row into another result row. The behavioral independence test mutates a result cell in one row and re-reads a different result row: the other row is unchanged.
11. Tests assert dimensions with exact equality and values with a fixed tolerance. The tolerance is pinned by the project and applied uniformly.
12. 1×1, row vector (1×N), column vector (N×1), negative, decimal, and zero values are valid inputs and produce correct outputs within the tolerance.

## 9. Inputs and Outputs

### Inputs

- Two `[][]float64` matrices for add and multiply. Each input is a slice of rows, each row a slice of columns. All rows in a single input must have the same length for the input to be valid.
- One `[][]float64` matrix for transpose. Same validity rule as above.
- Floating-point values may be negative, zero, positive, decimal, integer-valued, very large, or very small.

### Outputs

- A fresh `[][]float64` matrix. Result rows are independent slices and are not aliased to any input row or to each other.
- An `error`. A successful operation returns a `nil` error. An invalid input returns a contextual error that names the offending operand, the row index (for ragged inputs), and the compared shapes.

### Example text-only success runs

Input A:
```
1 2
3 4
```

Input B:
```
5 6
7 8
```

Add result:
```
 6  8
10 12
```

Multiply result (A × B):
```
19 22
43 50
```

Transpose of A:
```
1 3
2 4
```

### Example text-only error runs

```
Add: error: empty matrix: left operand has 0 rows.
Add: error: empty matrix: left operand has 0 columns.
Add: error: ragged matrix: row 2 has length 3, expected 4.
Multiply: error: shape mismatch: left columns 4 do not equal right rows 3.
Transpose: error: ragged matrix: row 0 has length 3, row 1 has length 4.
```

## 10. Rules and Edge Cases

- **Empty matrix (zero rows).** Rejected with an error that names the operand and identifies "empty matrix". The project does not invent a column count.
- **Empty matrix (every row zero columns).** Rejected with an error that names the operand and identifies "empty matrix". The project does not invent a row count for a zero-column input, does not treat "every row has zero columns" as a usable shape, and does not classify this state as ragged.
- **Mixed zero and non-zero rows.** A matrix whose outer slice has rows but whose rows mix zero and non-zero lengths is rejected. A mix of zero and non-zero row lengths is "ragged" if any two rows have different lengths, and "empty" if every row has zero length. The error identifies which state applies and names the offending row index and its actual length.
- **Ragged matrix.** Rejected with an error that names the offending row index and the expected length. The expected length is the first row's length. The error identifies "ragged matrix" so a caller can distinguish it from an empty-matrix error or a shape mismatch.
- **Shape mismatch (add).** Rejected with an error that names both shapes. No element is read.
- **Shape mismatch (multiply).** Rejected with an error that names the left column count and the right row count. No element is read.
- **Single element (1×1).** Add and multiply with another 1×1 produce a 1×1 result. Transpose produces an identical 1×1 result.
- **Row vector (1×N).** Transposed into a column vector (N×1). Add requires both inputs to be 1×N with the same N. Multiply of (1×N) × (N×M) produces a (1×M) result.
- **Column vector (N×1).** Transposed into a row vector (1×N). Add requires both inputs to be N×1 with the same N. Multiply of (N×M) × (M×1) produces an (N×1) result.
- **Identity matrix.** Multiplying an N×N matrix by the N×N identity matrix returns the original within tolerance. Multiplying the N×N identity matrix by an N×N matrix returns the original within tolerance.
- **Zero matrix.** Adding a zero matrix of the same shape to another matrix returns that matrix within tolerance. Multiplying a zero matrix of compatible shape with another matrix returns a zero matrix of the resulting shape within tolerance.
- **Negative and decimal values.** Treated as ordinary inputs. No special handling, no rejection.
- **Very large or very small values.** Treated as ordinary inputs. Tolerance is applied uniformly.
- **Transpose twice.** Returns a matrix with the original shape and the original values, element-for-element, within tolerance.
- **Non-mutation.** Mutating a result row does not change any input row and does not change any other result row.
- **Non-aliasing.** The result rows are not aliased to any input row or to any other result row. Behavioral independence tests confirm this.
- **No panic.** Every invalid path is an `error` return. The package does not panic on out-of-range, nil input, or mismatched shape.

## 11. Project Constraints

- Go standard library only. No third-party numeric libraries, no BLAS bindings, no `gonum`. The implementation uses the language's built-in `float64` arithmetic.
- Inputs and outputs are `[][]float64`. No custom matrix type is required for the package's public contract; the learner may introduce internal types, but the public operations must accept and return `[][]float64`.
- Operations never mutate input rows and do not alias them.
- Operations never panic on invalid input. Every invalid path returns an error.
- Empty and ragged inputs are always rejected with a contextual error. The project does not silently treat a zero-row slice as preserving a column count.
- Tests assert dimensions exactly and values within a fixed tolerance.
- Negative, decimal, integer-valued, very large, and very small `float64` values are valid inputs.
- The package does not depend on the terminal, real user directories, network, or any external service. Core logic is testable with in-memory slices.
- No panics are caught with `recover` to fake a successful path. The error return is the only signal.

## 12. Design Questions Before Coding

- How is shape validated up front so that no element is read before the shape is known to be valid? Does the validator return on the first issue, or collect all issues at once?
- Where does the empty-matrix check live? As a precondition for every operation, inside a shared helper, or in each operation independently?
- Where does the ragged-matrix check live? As a precondition, inside a shared helper that walks the rows once, or inline in each operation?
- How is the result allocated? As a fresh slice of row slices with row length pinned to the result's column count, or by appending to an empty outer slice?
- How is row independence guaranteed? Are result rows copied element-by-element into fresh slices, or are inputs copied first and the result derived from the copy? Which choice keeps the behavioral independence tests straightforward?
- How are operations made independent of each other? Does add reuse a helper that multiply also uses, or are they written independently with the same shape discipline?
- How is the error shaped? Plain `error`, wrapped with `fmt.Errorf` and `%w`, or a typed error with named fields (offending operand, row index, expected length, actual length)? Which choice keeps the error readable in tests and in production logs?
- How is the transpose implemented so that transposing twice returns the original? Is the second transpose a fresh allocation that mirrors the first, or does it rely on a property of the result rows?

## 13. Implementation Milestones

1. Decide the package layout. Keep `main` as a thin wrapper that reads matrices from standard input (or a fixed fixture) and prints results. Keep operations in a small package with shape validation, add, multiply, transpose, and the no-aliasing invariant pinned in one place.
2. Pin the public contract as named constants: the shape rule for add, the shape rule for multiply, the shape swap rule for transpose, the empty-matrix rejection rule, the ragged-matrix rejection rule, and the floating-point tolerance used in tests.
3. Implement shape validation. Reject empty matrices with a contextual error. Reject ragged matrices with a contextual error naming the offending row index and its actual length. Reject shape mismatches for add and multiply with contextual errors naming the compared shapes. No element is read before shape validation passes.
4. Implement add. Validate shapes, allocate a fresh result with the right shape, fill each element with the sum of the corresponding input elements, return the result and a `nil` error.
5. Implement multiply. Validate shapes, allocate a fresh result with `left rows` rows and `right columns` columns, fill each element with the dot product of the corresponding left row and right column, return the result and a `nil` error.
6. Implement transpose. Validate the input, allocate a fresh result with the swapped shape, fill each element with the input element at the swapped position, return the result and a `nil` error.
7. Verify the row-independence invariant. After every operation, the result rows are independent of input rows and of each other. The verification is behavioral: mutating a result cell does not change any input cell, mutating an input cell does not change any result cell, and mutating one result row does not change another result row.
8. Wire `main`. Accept matrices through a small driver (fixed fixture, file input, or standard input). Print successful results or errors. The driver is not part of the package's public contract; it is the demonstration harness.
9. Add tests for every verification case in section 14, with shape-validation tests, value-correctness tests, identity-matrix and zero-matrix tests, transpose-twice tests, and behavioral independence tests separated.

## 14. Verification Cases the Learner Must Write

Each case is described in natural language. Tests use in-memory matrices and assert dimensions exactly and values within the project's pinned tolerance.

### Shape validation

- Add accepts two matrices of the same shape.
- Add rejects a left operand that is empty (zero rows) with an error naming the operand and identifying "empty matrix".
- Add rejects a right operand that is empty with an error naming the operand and identifying "empty matrix".
- Add rejects a left operand whose rows are all zero columns with an error naming the operand and identifying "empty matrix".
- Add rejects a right operand whose rows are all zero columns with an error naming the operand and identifying "empty matrix".
- Add rejects a left operand that is ragged with an error naming the offending row index and length and identifying "ragged matrix".
- Add rejects a right operand that is ragged with an error naming the offending row index and length and identifying "ragged matrix".
- Add rejects two matrices of different shapes (for example 2×3 vs 3×2) with an error naming the two shapes.
- Multiply rejects a left operand that is empty with an error naming the operand.
- Multiply rejects a right operand that is empty with an error naming the operand.
- Multiply rejects a left operand whose rows are all zero columns with an error naming the operand.
- Multiply rejects a right operand whose rows are all zero columns with an error naming the operand.
- Multiply rejects a left operand that is ragged with an error naming the offending row index and length.
- Multiply rejects a right operand that is ragged with an error naming the offending row index and length.
- Multiply rejects two matrices whose left columns do not equal right rows with an error naming the compared counts.
- Transpose rejects a matrix that is empty (zero rows or all-zero-column rows) with an error naming the operand.
- Transpose rejects a matrix that is ragged with an error naming the offending row index and length.

### Value correctness

- Add produces the element-wise sum within tolerance.
- Multiply produces the dot-product result within tolerance, for at least one non-trivial pair (for example 2×3 times 3×2).
- Transpose swaps rows and columns correctly for a non-square matrix (for example 2×3 → 3×2).
- Transpose of a square matrix swaps rows and columns correctly.
- Multiply is consistent with add: A × (B + C) equals A × B + A × C within tolerance, for matrices where the shapes allow.
- Multiply is not commutative in general: A × B does not necessarily equal B × A, and the test pins this with concrete matrices.

### Identity and zero matrices

- Multiplying an N×N matrix by the N×N identity matrix returns the original within tolerance.
- Multiplying the N×N identity matrix by an N×N matrix returns the original within tolerance.
- Multiplying a zero matrix of compatible shape by another matrix returns a zero matrix of the resulting shape within tolerance.
- Adding a zero matrix of the same shape to another matrix returns that matrix within tolerance.

### 1×1, row vector, column vector

- A 1×1 matrix adds with another 1×1 to produce a 1×1 result within tolerance.
- A 1×1 matrix multiplies with another 1×1 to produce a 1×1 result within tolerance.
- Transpose of a 1×1 matrix is a 1×1 with the same value within tolerance.
- Transpose of a row vector (1×N) is a column vector (N×1).
- Transpose of a column vector (N×1) is a row vector (1×N).

### Negative, decimal, zero, very large, very small

- Add handles negative values correctly within tolerance.
- Multiply handles negative values correctly within tolerance.
- Transpose handles negative values correctly within tolerance.
- Operations handle zero values correctly.
- Operations handle decimal values correctly within tolerance.
- Operations handle very large values correctly within tolerance.
- Operations handle very small (close to zero) values correctly within tolerance.

### Transpose twice

- Transpose twice returns a matrix with the original shape.
- Transpose twice returns the original values, element-for-element, within tolerance.

### Non-mutation

- After add, the inputs are unchanged (lengths and values).
- After multiply, the inputs are unchanged.
- After transpose, the input is unchanged.
- Mutating a result row's element does not change the corresponding input row's element.

### Row-independence checks (behavioral)

- Mutate a result cell at position `(i, j)`. Re-read the corresponding input cell at position `(i, j)`. The input cell is unchanged. Run this for both add and multiply with non-trivial inputs and for transpose with a non-square input.
- Mutate a result cell in one result row. Re-read a different result row from the same operation. The other row is unchanged.
- Mutate an input cell at position `(i, j)`. Re-read a previously returned result cell at the corresponding position. The result cell is unchanged.
- Run the same behavioral checks for multiply and transpose.

### Process

- A test runs the driver against a small fixture and confirms the printed output matches the expected text-only form.
- A test runs the driver against invalid input and confirms the error message names the offending operand and the row index (when applicable).

## 15. Common Mistakes to Watch For

- **Aliasing input rows.** Returning a result that shares a row with an input means mutating the result mutates the input. The contract is behavioral independence: every result row must be a fresh slice so that mutating a result cell does not change an input cell.
- **Aliasing result rows to each other.** Two result rows from the same operation must not share a backing array. The result's outer slice must be a slice of independent row slices. The behavioral independence test catches this: mutating one result row and re-reading another must not show the change.
- **Treating a zero-row slice as preserving a column count.** The project rejects empty matrices. Guessing the column count from a different input is the bug the empty-matrix rule exists to prevent.
- **Treating every row having zero columns as a usable matrix.** The project rejects "every row has zero columns" as an empty matrix. Guessing a row count, using `len(m)` as a usable row count, or using zero-column rows in arithmetic is wrong.
- **Treating ragged slices as rectangular.** The project rejects ragged matrices with a contextual error. Silently using the first row's length, the longest row's length, or the shortest row's length is wrong. A matrix whose rows are all length `0` is not ragged; it is empty.
- **Reading elements before validating the shape.** A panic from an out-of-range index is a sign that the shape was not validated first. The package never panics on shape errors.
- **Using exact equality for floating-point values.** Values are compared within a pinned tolerance. Exact equality on `float64` arithmetic is fragile and produces flaky tests.
- **Using tolerance for dimension checks.** Dimensions are exact. Tolerance on a row count is meaningless.
- **Returning a generic error that does not name the operand.** Tests need to identify whether the left operand or the right operand is wrong. The error must name the operand (and, for ragged inputs, the row index).
- **Allowing add to work when one operand has zero rows and the other has rows.** The empty-matrix rule rejects this regardless of the other operand's shape.
- **Reusing a typed matrix type in the public contract.** The contract is `[][]float64`. A wrapper type can live internally but must not change the operation's input and output shape.
- **Recreating a known bug as "panics on invalid input".** The project never panics on invalid input. Every invalid path is an `error` return.
- **Producing a different shape than the documented result.** Add returns the same shape as the inputs. Multiply returns `left rows` rows and `right columns` columns. Transpose returns the swapped shape. Each is part of the contract.

## 16. Topics and References for Study

- A Tour of Go: "More types: pointers", "Slices", "Errors".
- Effective Go: "Data", "Errors", "Names".
- Package documentation: `errors` (New, Is, As), `fmt` (Errorf, %w), `math` (NaN, IsNaN, Inf, MaxFloat64, SmallestNonzeroFloat64).
- Floating-point comparison patterns: search for "Go float64 tolerance test", "Go nearly equal", "Go relative error float compare".
- Slice independence patterns: search for "Go slice copy row", "Go 2D slice independent rows", "Go slice backing array shared".
- 2D shape validation patterns: search for "Go rectangular matrix validation", "Go ragged slice error", "Go matrix empty input reject", "Go zero column row reject".

## 17. Self-Assessment Questions

1. Why does the project reject an empty matrix (zero rows, every row zero columns, or both) instead of treating it as "zero rows with a known column count" or "every row has zero columns but some known row count"?
2. Why does the project reject a ragged matrix, and why must its contextual error name the offending operand and row instead of reporting only a generic failure?
3. Why must add validate shapes before reading any element, and what would a panic-driven validation look like that the project forbids?
4. Why must multiply require left column count to equal right row count, and why is the resulting shape `left rows` rows and `right columns` columns?
5. Why does transposing twice return the original shape and values, and what does a "transpose twice" test prove about the implementation?
6. Why must result rows be independent of input rows and of each other, and how does the behavioral independence test pin independence without inspecting internals?
7. Why must floating-point value comparisons use a tolerance while dimension comparisons use exact equality?
8. Why are 1×1, row vector, column vector, negative, decimal, very large, and very small values ordinary valid inputs that need no special handling?
9. Why must invalid dimensions return errors without panicking or using `recover`, and what do this rule and its no-panic tests reveal about testability and hidden failures?
10. Why is the public contract `[][]float64` rather than a custom matrix type, and what would an internal type buy?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every functional requirement in section 8 is satisfied.
- Every verification case in section 14 has a corresponding test.
- Empty and ragged matrices are rejected with a contextual error that names the offending operand and (for ragged inputs) the row index. "Empty matrix" covers zero rows, every row having zero columns, and zero rows with every row having zero columns. "Ragged matrix" covers rows of differing positive lengths and rows mixing zero and non-zero lengths.
- Shape mismatches in add and multiply are rejected before any element is read, with errors that name the compared shapes.
- Result rows are independent of input rows and of each other. Behavioral independence tests confirm it: mutating a result cell does not change an input cell, mutating an input cell does not change a previously returned result cell, and mutating one result row does not change another result row.
- 1×1, row vector, column vector, identity, zero, negative, decimal, very large, and very small inputs produce correct outputs within the pinned tolerance.
- Transpose twice returns the original shape and values within tolerance.
- Tests assert dimensions exactly and values within the pinned tolerance.
- The package never panics on invalid input. Every invalid path is an `error` return.
- The package documentation states the shape rules, the empty-matrix rejection rule (zero rows, every row zero columns, both), the ragged-matrix rejection rule, the row-independence rule, and the tolerance rule.
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Scalar multiply.** Add a scalar-multiply operation that multiplies every element of a matrix by a `float64`. The result has the same shape. Negative, decimal, zero, very large, and very small scalars work. Result rows are independent slices. No panic on empty or ragged inputs. Do not add matrix-by-scalar and scalar-by-matrix as separate operations; one operation with the scalar on the right is enough.
- **Frobenius norm.** Add a function that returns the square root of the sum of squares of every element of a matrix as a `float64`. The function returns an error for empty or ragged inputs. Tests use the tolerance rule. Do not add other norms (1-norm, infinity-norm) and do not change the existing operations.
