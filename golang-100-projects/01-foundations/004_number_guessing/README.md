# Project 004 — Number Guessing

## 1. Project Name and Number
- Number: **004**, level 1 (language basics and CLI).
- Folder name in the table: **`004_number_guessing`**, matching `01-foundations/004_number_guessing/`.
- Kind: a terminal guessing game where the program picks a number inside a bounded range and the user has a bounded number of attempts to guess it, receiving higher-or-lower hints after each guess until they win or run out.

## 2. Project Idea
Build a terminal game where the program picks a number inside a fixed range, and the user has a fixed number of attempts to guess it. After each guess, the program tells the user whether the secret is higher or lower than their guess, until the user wins or runs out of attempts. The source of randomness must be injectable so a test environment can supply a deterministic stream. The program must handle invalid input without crashing; whether an invalid guess consumes an attempt is a design choice the learner documents and applies consistently.

## 3. Why This Project Now?
- Reuses the input and printing skills from earlier projects and layers in a non-trivial loop that must terminate predictably.
- Introduces dependency injection in its simplest form: pass a "source of numbers" in, instead of calling a global generator. Without this habit, testing concurrent programs in later projects becomes painful.
- Reinforces state management inside a single function: the secret, the attempts remaining, and the most recent hint all live somewhere, and how you keep them straight matters.
- Models real software where randomness must be replaceable.

## 4. Prerequisites
- 003 must be complete, especially the input-cleaning habit and the separation of logic from terminal I/O.
- Environment: Go installed and on `PATH`.

## 5. What You Must Know Before Starting
- Loop shapes: Go offers a single loop construct with several idioms (a bounded counter, a condition, or a body that exits explicitly). Recognise which form best fits this project.
- Random sources: the standard library provides a pseudo-randomness package with both a global generator and an opt-in pattern that creates a generator from a seed. Understand why the opt-in pattern enables testing.
- Seeding: pseudo-random generators produce the same sequence for the same seed. A time-based seed varies the sequence per run; a fixed seed reproduces the sequence. Pick deliberately.
- Range semantics: decide whether your range bounds are inclusive or exclusive, and document it.
- Termination: understand which loop-exit condition guards against an infinite loop.
- Terminal end-of-file: receiving an end-of-file signal while reading is a valid scenario; the program must not panic on it.

## 6. Explanation of New Concepts
- Injectable randomness: instead of calling the global generator from inside the game logic, accept something that produces numbers, and pass a real generator at runtime and a fixed generator in tests. The shape of the injection is up to you; the principle is that "what produces numbers" is a parameter, not a hidden global.
- Seed choice: in production, vary the seed per run; in tests, freeze the seed to make results reproducible.
- Game state: the secret and the attempts counter are state; the hint is computed from each new guess. Keep state in clearly named variables or in a small data structure, and update it in one well-known place.
- Reading and reacting to invalid input: this is policy, not syntax. Decide whether an invalid guess consumes an attempt, apply that policy consistently, and document it in the code.
- Read-after-end-of-file handling: the terminal can signal end of input. Treat it as a real situation rather than as an impossible case.

## 7. Learning Objective
By the end of the project you should be able to:
- Construct a loop that terminates on either a correct guess or exhaustion of attempts.
- Inject a random number source into the game's logic so that the same secret emerges for the same input stream in tests.
- Update state without losing track of the secret, the attempts counter, or the hint direction.
- Distinguish between inputs that are numeric-but-out-of-range and inputs that are not numeric, and react to each.
- Make a defensible policy on whether invalid inputs consume an attempt, apply it consistently, and document it in the code.

## 8. Functional Requirements
- F1: The game uses a bounded numeric range for the secret and a bounded maximum number of attempts. Both bounds are fixed in the program.
- F2: The game uses a random source that can be replaced by a test source. The mechanism of replacement is your design.
- F3: On each attempt, the program reads the user's guess, compares it to the secret, and prints a hint that tells the user how to adjust. If the guess is above the secret, the hint says the secret is lower. If the guess is below the secret, the hint says the secret is higher. The hint is consistent with the comparison and does not reveal the secret.
- F4: A correct guess ends the game with a message that mentions the secret and the number of attempts used.
- F5: Exhaustion of attempts without a correct guess ends the game with a message that mentions the secret.
- F6: Non-numeric input is handled without crashing; the policy on whether it consumes an attempt is yours, must be applied consistently, and must be documented in the code.
- F7: Numeric input outside the chosen range is handled with a clear message.
- F8: An end-of-file signal on input is handled without crashing.

## 9. Inputs and Outputs
**Inputs**:
- Numeric guesses inside the chosen range.

**Outputs**: text to standard output. Text-only examples:

- A win scenario:
  - Program announces the range and the attempts allowed.
  - User enters guesses, receiving a hint after each.
  - User enters the correct guess before attempts are exhausted.
  - Program prints a win message that names the secret and states the attempts used.

- An exhaustion scenario:
  - Attempts are exhausted before the user guesses correctly.
  - Program prints a message that names the secret.

- A guess above the secret:
  - User enters a guess that is greater than the secret.
  - Program prints a hint that says the secret is lower, without revealing the secret.

- A guess below the secret:
  - User enters a guess that is less than the secret.
  - Program prints a hint that says the secret is higher, without revealing the secret.

- Non-numeric input:
  - User enters text that is not parseable as an integer.
  - Program prints a message indicating the input is not a valid guess; the reaction matches your chosen policy.

- Numeric input outside the range:
  - User enters a number below the lower bound or above the upper bound.
  - Program prints a message indicating the guess is outside the range.

- End-of-input:
  - The user signals end-of-file.
  - Program prints a message and exits without panicking.

## 10. Rules and Edge Cases
- The exact bounds of the range and the exact maximum attempts are fixed in the program; they are documented in code, not as user input.
- A guess that matches the secret exactly is the only win condition.
- Invalid numeric input: your chosen policy applies uniformly across the run.
- Numeric input just outside the range: rejected with a clear message.
- Numeric input inside the range but equal to the secret: wins, regardless of attempts remaining.
- The hint follows a fixed convention: a guess above the secret produces a hint that says the secret is lower; a guess below the secret produces a hint that says the secret is higher. The hint is based only on the secret and the user's most recent guess.
- After the game ends (whether by win or exhaustion), the program ends the run; the baseline does not include another round.

## 11. Project Constraints
- Libraries: the standard library only. The formatted I/O package and the pseudo-randomness package are sufficient; the time package may be used for seeding.
- Prohibited: any external package; cryptographic randomness is not appropriate for this project.
- Persistence: none.
- Network: none.
- Scope: a single bounded guessing game per program run. The baseline does not include replay; a replay loop, if you want one, is a design choice beyond the baseline.

## 12. Design Questions Before Coding
- What is the range of the secret, and what is the maximum number of attempts? Both are design constants; choose with readability in mind.
- How do you inject the random source? As a value the game uses internally, as a function the game calls, or as an object with a method? Each is valid; pick one and reason about clarity and testability.
- What does your program do with an end-of-file signal? Treat it as immediate loss, treat it as nothing, or define some other behaviour? Make it explicit.
- Does non-numeric input consume an attempt under your chosen policy? Apply the policy uniformly once chosen, and document the choice in the code.
- How do you keep track of the secret and the attempts counter? In two variables, in a small data structure, or in another approach? Reason about clarity.
- How do you phrase higher-or-lower hints so that they do not leak information the user should not have at that moment?

## 13. Implementation Milestones
1. M1: Create the source file in the project folder with the minimum required for a Go program to compile and run; verify with a build and run cycle.
2. M2: Implement the random-source injection: a routine that, given a source of randomness (the shape is your choice), produces the secret.
3. M3: Implement the game's main loop with a clear exit condition that fires on a correct guess or on attempts running out.
4. M4: Implement the higher-or-lower hint logic, based on the secret and the user's most recent numeric guess.
5. M5: Implement the win and exhaustion messages.
6. M6: Implement the response to non-numeric input, applying your chosen policy consistently.
7. M7: Implement the response to numeric input outside the chosen range.
8. M8: Implement the response to an end-of-file signal without crashing.
9. M9: Run the verification scenarios in section 14 and confirm the documented behaviour.

## 14. Verification Cases the Learner Must Write
- A correct guess within the attempt budget prints a win message naming the secret and the attempts used.
- All-attempts exhaustion prints a message naming the secret.
- An injected deterministic source produces the same secret for the same input stream across two program runs.
- A guess above the secret yields a hint that says the secret is lower, and does not reveal the secret.
- A guess below the secret yields a hint that says the secret is higher, and does not reveal the secret.
- A non-numeric guess is handled without crashing, per your chosen policy.
- A numeric guess outside the range is rejected with a clear message.
- An end-of-file signal is handled without panicking.
- The loop terminates on every run; there is no path that loops forever.
- A guess equal to a previous wrong guess does not change the hint for the next guess.

## 15. Common Mistakes to Watch For
- Calling a global random generator from inside the game logic; this defeats testing.
- Letting the loop continue past the attempt limit because the exit condition is in the wrong place.
- Comparing user input as a string to a numeric secret without conversion; the result is nearly always wrong.
- Letting hints be inconsistent with the comparison. The convention is fixed: a guess above the secret produces a hint that says the secret is lower; a guess below the secret produces a hint that says the secret is higher. Following any other convention is a bug.
- Treating non-numeric input and out-of-range numeric input as the same case; they require different messages and may have different policies.
- Forgetting to react to end-of-file; this is the most common source of panics on input APIs.
- Sharing mutable state across the loop body in confusing ways; the secret and counter must remain coherent.
- Applying the policy on invalid input inconsistently across the run.

## 16. Topics and References for Study
- Effective Go: For loops.
- The official documentation for the pseudo-randomness package, especially the difference between the global generator and an opt-in source.
- The official documentation for the time package, if used for seeding.
- General references: A Tour of Go on flow control.
- Search terms: `Go rand source dependency injection`, `Go loop termination condition`, `Go stdin EOF handling`.

## 17. Self-Assessment Questions
1. Why is passing in a random source preferable to using a global generator, especially for tests?
2. What happens if your loop's exit condition does not check the attempts counter? Walk through a worst case.
3. Why does non-numeric input merit a different message from numeric-out-of-range input?
4. How do you reconcile the choice between "non-numeric input consumes an attempt" and "non-numeric input does not consume an attempt"? Pick one and apply it.
5. In your own words, state the hint convention: what does the hint say when the user's guess is above the secret, and what does it say when the guess is below the secret?
6. Why must the higher-or-lower hint be derived from the secret and the most recent guess, not from anything else?
7. If the same game is run twice with a fixed seed and the same guess sequence, the secret and the outcome must be reproducible. Explain why, and which part of the design enables it.

## 18. Definition of Completion
- The program compiles and runs without compile errors.
- Every scenario in section 14 produces the behaviour documented in the code.
- The random source is replaceable for testing.
- Hints are consistent with the comparison.
- The loop terminates on every run.
- You can explain why the source is injectable, how the loop terminates, how end-of-file is handled, and what your policy on invalid input is.

## 19. Optional Extensions
- Optional 1: Track which numbers the user has already guessed, and on a duplicate guess, print a message naming the previous guess without consuming an attempt.
- Optional 2: Add a small log of guesses outside the main flow that is testable separately; the policy on duplicates follows your extension design.
