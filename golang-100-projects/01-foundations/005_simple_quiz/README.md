# Project 005 — Simple Quiz

## 1. Project Name and Number
- Number: **005**, level 1 (language basics and CLI).
- Folder name in the table: **`005_simple_quiz`**, matching `01-foundations/005_simple_quiz/`.
- Kind: a terminal quiz that walks an in-memory bank of multiple-choice questions, accepts the user's choice for each, and at the end prints the score and the corresponding percentage.

## 2. Project Idea
Build a terminal quiz that asks a small in-memory bank of multiple-choice questions, accepts the user's choice for each, and at the end prints the score and the corresponding percentage. The questions are stored as a collection of typed records that bundle the prompt, the choices, and the correct choice. The quiz must handle three situations gracefully: every answer is correct, every answer is wrong, and a question in the bank has no choices at all. The program reacts to each situation with a clear message rather than crashing.

## 3. Why This Project Now?
- Reuses everything from earlier projects and layers in a typed record bundling multiple fields under one name, the structural foundation for nearly every later project.
- Introduces a collection type for ordered, indexable storage of homogeneous records.
- Reinforces separating data (the bank) from logic (the asker, the evaluator, the reporter).
- Introduces an edge case that appears whenever data is hand-written: a question in the bank may be missing its choices, and the program must react to this rather than crash on it.

## 4. Prerequisites
- 004 must be complete, especially the loop termination habit and the input-validation style.
- Environment: Go installed and on `PATH`.

## 5. What You Must Know Before Starting
- Typed records: a way to bundle fields of different types under one name. Understand how to declare and instantiate them.
- Collection types: an ordered, indexable sequence of elements of the same type. You can traverse it with an index and a value, and you can pre-declare it as a literal.
- Traversal: how to visit each element once with a clear exit condition.
- Empty collections: an empty collection is a valid value that the program must handle without crashing.
- Output formatting: a way to report both the count (correct out of total) and the percentage, with your chosen precision. Both numbers describe the same underlying count, and the reporting is consistent across outcomes.
- Type choice for the percentage calculation: integer division versus floating-point division, and why direct equality on a percentage is rarely safe.

## 6. Explanation of New Concepts
- A typed record as data plus knowledge: putting the correct answer inside the record is convenient for a small program; in larger systems it is often separated. The decision belongs to you.
- A collection of records: storing an in-memory bank is a slice of records. The traversal order is the order of the collection, which is the order you choose to construct it in.
- Variable-content records: a record's choices field can carry any number of entries, including zero. The program must handle the zero case sensibly even though it is unusual for a multiple-choice question, because the data may contain a malformed entry.
- The arithmetic of a percentage: dividing two integers does not yield a float, so the numeric type must be chosen with intent.
- Two views of the same count: the program reports both the score as a count (correct out of total) and the corresponding percentage. Both numbers describe the same underlying count; the percentage is derived from it. Reporting the count in two inconsistent ways (for example, a fraction here and a percentage with different rounding there) is an avoidable source of inconsistency. Choose one rendering of the count and one rendering of the percentage and apply them uniformly across all outcomes.

## 7. Learning Objective
By the end of the project you should be able to:
- Declare a typed record that bundles a prompt, a list of choices, and a correct answer.
- Build a collection of such records and traverse it in order.
- Match the user's choice against the correct answer, applying your matching policy (whitespace normalisation, case folding, or both).
- Compute and report both a score count and a percentage, both derived from the same correct count, using one chosen numeric type.
- Handle a record whose choices list is empty without crashing.
- Handle an empty collection gracefully.

## 8. Functional Requirements
- F1: The program holds an in-memory bank of questions. The bank must contain at least one ordinary multiple-choice question and must also include a question whose choices list is empty, so that the no-choices case is present in the data and can be observed. The total bank size is your design choice.
- F2: The program walks the bank in order, presents each question, and reads the user's choice.
- F3: The program records whether each answer is correct.
- F4: After the last question, the program reports both the score (correct out of total) and the corresponding percentage. Both numbers reflect the same underlying count, and the reporting is consistent across all outcomes.
- F5: The program handles the all-correct case and the all-wrong case in the same form as partial scores.
- F6: A question with an empty list of choices is handled with a clear message rather than a crash.
- F7: An empty or unrecognised choice is handled with a clear message rather than a crash.
- F8: An empty bank is handled with a clear message rather than a crash.

## 9. Inputs and Outputs
**Inputs**:
- A choice string for each question, in the format your matcher accepts.

**Outputs**: text to standard output. Text-only examples:

- All answers correct:
  - The user answers every presented question correctly.
  - Program prints a final result line in the canonical form, naming the count and the percentage with the chosen precision.

- All answers wrong:
  - The user answers every presented question incorrectly.
  - Program prints a final result line in the same form as the all-correct case.

- Mixed answers:
  - The user answers some questions correctly and some incorrectly.
  - Program prints a final result line in the same form as the previous cases.

- A question with no choices:
  - The bank includes a question whose choices list is empty.
  - Program prints a clear message indicating that question is malformed (your wording) and continues with the rest, or exits, per your design.

- An empty bank:
  - The collection is empty.
  - Program prints a message indicating the bank contains no questions and exits without crashing.

## 10. Rules and Edge Cases
- A choice string with surrounding whitespace: handled per your matching policy.
- A choice string in a different case (upper versus lower): handled per your matching policy.
- A choice string that does not correspond to any of the listed choices: rejected with a clear message.
- An empty choice string: rejected with a clear message.
- A record whose correct-answer field is invalid: handled per your design.
- A record whose choices field is empty: handled per your design.
- An empty bank: handled with a clear message.
- An end-of-file signal during input: handled without crashing.

## 11. Project Constraints
- Libraries: the standard library only. The formatted I/O package and the strings package, if you choose to normalise inputs, are sufficient.
- Prohibited: any external package.
- Persistence: none. The bank is in the source code, not in a file.
- Network: none.
- Scope: a small in-memory quiz. Loading questions from a file, randomising order, and adaptive difficulty are not required by the baseline.

## 12. Design Questions Before Coding
- How many questions are in your bank, and how many choices per question? Justify the choice with readability and with the no-choices case you must cover.
- Where does the bank live? As a package-level value, a function-local value, or otherwise? Reason about clarity.
- Does the correct answer live inside the record, in a separate collection indexed somehow, or elsewhere? Each choice has trade-offs.
- How do you match a user's choice against the correct one? Exact match, or normalised match that ignores case and whitespace?
- What numeric type do you use for the percentage calculation, and why is direct equality rarely safe for percentages?
- How do you report the count and the percentage so that both describe the same underlying count, and so that the all-correct, all-wrong, and mixed cases use the same rendering?

## 13. Implementation Milestones
1. M1: Create the source file in the project folder with the minimum required for a Go program to compile and run; verify with a build and run cycle.
2. M2: Declare the typed record and define a bank of questions that contains at least one ordinary multiple-choice question and that also includes a no-choices case (a question whose choices list is empty).
3. M3: Implement traversal over the bank in order, presenting each question and reading the user's choice.
4. M4: Implement the matcher that compares the user's choice with the correct answer, applying your matching policy.
5. M5: Implement the score tracker that records the correct count and the total count.
6. M6: Implement the final result output that reports both the score count and the percentage, both reflecting the same underlying correct count, applied uniformly across the all-correct, all-wrong, and mixed cases.
7. M7: Implement the empty-bank branch and the malformed-question branch as clear messages, not crashes.
8. M8: Run every verification scenario in section 14 and confirm the documented behaviour.

## 14. Verification Cases the Learner Must Write
- All answers correct: the program prints the final result line in the canonical form with the appropriate count and percentage.
- All answers wrong: the same form, with the count at zero.
- Mixed answers: the same form, with a count strictly between zero and the total.
- A question whose choices list is empty: the program produces a clear message rather than a crash.
- A non-empty choice string that does not correspond to any listed choice: rejected with a clear message.
- An empty choice input: rejected with a clear message.
- A choice string in a different case: handled per your matching policy.
- A choice string with surrounding whitespace: handled per your matching policy.
- An end-of-file signal during input: handled without crashing.
- The percentage calculation produces a value that reflects the actual ratio, not a truncated integer-only division.

## 15. Common Mistakes to Watch For
- Using integer division for the percentage calculation, which truncates the fraction.
- Embedding the correct answer in a way that confuses reading and evaluation logic.
- Reporting the count and the percentage in inconsistent ways across the all-correct, all-wrong, and mixed cases. The two views must agree because they describe the same underlying count.
- Treating a question with no choices as a fatal error that crashes the program; it should be a defined branch.
- Having the traversal depend on a counter that gets out of sync with the collection's length.
- Comparing choices with case-sensitive or whitespace-sensitive equality by default; test the alternatives and pick a policy.
- Allowing mutation of the bank during traversal; the bank should be read-only during the quiz.
- Using a different rounding rule for the all-correct case than for partial cases; apply one rule.

## 16. Topics and References for Study
- A Tour of Go: Structs; A Tour of Go: Slices.
- Effective Go: Data structures.
- The official documentation for the strings package, particularly the normalisation helpers.
- General references: A Tour of Go on range-based traversal.
- Search terms: `Go struct slice quiz data model`, `Go int vs float64 percentage`, `Go whitespace case folding strings`.

## 17. Self-Assessment Questions
1. Why is it useful to bundle a question's prompt, choices, and correct answer in a single typed record rather than in parallel collections?
2. A question in the bank has no choices. How should the program react when it encounters it, and why is it useful to include this case in the bank rather than excluding it from the data?
3. Why does integer division produce the wrong percentage in most cases? Walk through an example.
4. The program reports both a count and a percentage at the end. How do these two views stay consistent with each other, and what would make them inconsistent?
5. If the bank grew from a few questions to a much larger set, what parts of the program would change and what would stay the same?
6. How do your matching policies on case and whitespace affect the user experience, and what is the principle behind the policy you chose?

## 18. Definition of Completion
- The program compiles and runs without compile errors.
- All-correct, all-wrong, and mixed cases produce a final result that reports both the count and the percentage, in the same form.
- A question with no choices and an empty bank each produce a clear message and never crash.
- Empty or unrecognised choices are rejected with clear messages.
- The bank is read-only during the quiz.
- You can explain how your typed record is shaped, how the traversal works, how the percentage is derived from the count, and how the no-choices case is handled.

## 19. Optional Extensions
- Optional 1: Add a random shuffle of the question order at the start of the quiz, keeping the same record shape and traversal.
- Optional 2: Add a second quiz round with a different bank, accessible by the same program run, sharing the typed record and the same reporting line.
