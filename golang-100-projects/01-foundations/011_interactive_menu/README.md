# Project 011 — Interactive Menu

## 1. Project Name and Number

Project **011** — `011_interactive_menu`. The directory name and number must match exactly.

## 2. Project Idea

A small command-line application that presents the user with a numbered menu of actions, reads a selection, looks that selection up in a registry of menu items, and invokes the corresponding behavior. The loop continues until the user chooses an explicit exit option, and it also ends cleanly when input is exhausted so that a buffer-driven test does not have to know the exit token in advance. The set of actions and the input/output streams the menu uses are not hard-coded inside the loop; the registry and the streams are injected from the outside so the program can be exercised end-to-end without touching a real terminal.

The product is not a specific calculator, not a specific greeting, and not a specific greeting-with-farewell. It is the *plumbing*: a menu shell whose body and whose termination are defined by what is wired in, not by code baked into the loop itself.

## 3. Why This Project Now?

Up to project 010 the learner has written programs whose control flow is a single straight script: read, compute, print, exit. Project 011 introduces the first piece of reusable structure on top of that: keeping behavior *as data* and dispatching to it by key. Without this idea, later projects — every CLI with subcommands, every plug-in system, every router — become copy-pasted switches. With it, the menu becomes a single dispatch point and each action is an independent unit that can be reached by a single lookup.

This project also separates *what the program does* from *where it reads and writes*. That boundary is the seed of every later testable CLI: the production program uses the real standard streams, and the test code substitutes byte buffers so it can assert exactly what was asked and what was answered.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 011 therefore requires:

- Completion of **010** (Password Generator).
- No prior knowledge of HTTP, databases, channels, generics, or concurrency.
- This project does **not** require project 014. Validation, if needed later, is its own project; the menu loop in 011 reads opaque tokens and does not parse them.

## 5. What You Must Know Before Starting

- How to declare a function in Go and the difference between a function and a function value.
- What a `map` is in Go and what its zero value does when a key is missing.
- How `fmt.Fprintf` and `fmt.Fscan` differ from `fmt.Printf` and `fmt.Scan`, and why writing to an arbitrary `io.Writer` and reading from an arbitrary `io.Reader` matters.
- The basic shape of an `io.Reader` and `io.Writer` (bytes in, bytes out) so you can pass buffers in tests.
- The difference between `os.Args`, command-line flags, and interactive prompts; this project is interactive, so the choice comes from a prompt, not from a flag.
- How `for` loops work and how a `break` inside a loop exits that loop only.
- What "input exhausted" means for an `io.Reader`: a read past the end returns an `io.EOF`-style error, which is a normal termination signal, not a crash.
- What "explicit exit" means: a sentinel value the user types to leave, distinct from the loop noticing there is nothing left to read.

## 6. Explanation of New Concepts

### Functions as values

In Go, a function name without parentheses is a value. You can store it in a variable, pass it as an argument, return it from another function, or put it inside a struct or a map. This makes "behavior" a first-class thing you can move around like an integer. For a menu, this means an action — say, "print the current time" — is itself a value that the menu registry can hold next to its label.

### Maps as registries

A `map` in Go associates keys with values. Looking up a key that is not present returns the zero value of the value type; for a function type that zero value is `nil`. Calling a `nil` function panics, so a missing key must be handled explicitly before any invocation. This is the natural place to express "unknown choice" as a defined, observable outcome rather than a crash.

### Dispatch versus switch

A `switch` over a token works for a tiny, fixed menu. A `map` lookup expresses the same idea as data rather than as control flow: each action lives at one place, the loop does not branch per action, and adding or removing an item is a change to the registry, not a change to the loop's body. The pedagogical point is the *separation* — the loop never names the actions directly. The learner chooses how to represent the registry; this project does not require the registry to grow at runtime, only that the loop not embed a long `switch`.

### Injected I/O boundaries

Reading from `os.Stdin` and writing to `os.Stdout` couples a program to a terminal. To test what the program asked and what it answered, both endpoints must be replaceable. The standard library uses `io.Reader` and `io.Writer` for this. Anything that satisfies those interfaces — including a `bytes.Buffer` — can stand in for the real streams. The production wiring uses the real streams; the test wiring uses buffers.

### Explicit exit versus input exhaustion

Two situations end a menu: the user types an exit token, or the input stream runs out. Both are normal outcomes and the loop must handle each as a defined, observable state. A buffer that simply holds `1\n3\n` (two selections, no exit token) still terminates: the third read returns the end-of-input signal and the loop exits with that signal. A test that exercises the menu without supplying an exit token must see a clean end-of-input termination, not a hang and not an infinite string of "unknown choice" lines.

### Action values versus action results

The action held in the registry runs when its key is selected. The loop's job is to dispatch; the action's job is whatever it does. The loop does not need to interpret the action's result. The learner decides whether the action writes to the injected writer directly, returns a value the loop prints, or both; the contract only says the loop invokes the value held in the registry, not a per-action branch in its own body.

## 7. Learning Objective

After completing this project the learner can:

- Store function values inside a `map` and dispatch to them by key.
- Distinguish "no such choice" from "input exhausted" from "exit token chosen" and express each as a defined outcome.
- Inject input and output streams so a menu loop can be tested without a terminal.
- Explain, in plain English, why storing behavior as data separates the dispatch loop from the actions themselves.
- Read and write through an arbitrary `io.Reader` / `io.Writer` pair.

## 8. Functional Requirements

The program must satisfy each requirement below. Requirements are numbered so they can be referenced in tests and in review.

1. On start, print a header, then a numbered list of registered menu items, each identified by a label (and any extra fields the learner chooses to show).
2. Include an **exit** option whose key is part of the registry and whose selection ends the loop with an "exit" outcome.
3. After printing the menu, prompt the user for a selection.
4. Read the selection from an `io.Reader` injected from the outside, not directly from `os.Stdin`.
5. Look the selection up in a registry keyed by the option key.
6. If the selection matches a registered action, invoke that action and write the action's output through an `io.Writer` injected from the outside.
7. If the selection does not match any registered action or cannot be parsed, write a clear "unknown choice" message through the injected writer and continue the loop.
8. If the input stream is exhausted before the user picks the exit option, end the loop with an "end of input" outcome and exit cleanly. This path must not loop forever and must not panic on a read past the end.
9. The loop ends on the first of: exit token chosen, or input exhausted. These are distinct outcomes the test can observe.
10. The action functions themselves are stored as values inside the registry; the menu loop does not contain a long `switch` over the actions.
11. The program must run successfully when input is supplied from a buffer in a test (for example a `bytes.Buffer`) and when output is captured by another buffer.

## 9. Inputs and Outputs

### Inputs

- A stream of bytes the program reads selections from. In production this is `os.Stdin`; in tests it is a buffer the learner controls.
- Each selection is a token — typically a single character or short word matching a key in the registry. Whitespace around the token is tolerated.

### Outputs

- A header and a numbered menu printed to the injected writer.
- A prompt for the next selection.
- The result of the chosen action, written to the injected writer.
- A clear "unknown choice" line when the selection does not match.
- A final line indicating how the loop ended: exit token, or end of input.

### Example text-only success run with exit token

```
=== Menu ===
1. Greet
2. Show time
3. Exit
Choose: 1
Hello from action one.
Choose: 3
Exited via menu.
```

### Example text-only run that ends because input is exhausted

```
=== Menu ===
1. Greet
2. Show time
3. Exit
Choose: 1
Hello from action one.
Choose: 2
It is some moment in time.
End of input.
```

### Example text-only run with an unknown choice

```
Choose: 9
Unknown choice: 9
Choose: 0
Unknown choice: 0
Choose: 3
Exited via menu.
```

(The exact wording is a design decision for the learner. The contracts are: an "unknown choice" line is its own line, the loop does not exit on it, and a buffer that does not contain the exit token still ends the loop with an "end of input" outcome.)

## 10. Rules and Edge Cases

- **Empty input**: a blank line is an unknown choice; the loop continues.
- **Whitespace around the token**: leading and trailing whitespace are tolerated.
- **Token that is not in the registry**: the "unknown choice" outcome; no silent exit, no panic.
- **Token that is in the registry but the value is missing**: in a correctly constructed registry this cannot happen; if it does, the program must not panic on a missing action.
- **Exit option**: selecting it ends the loop with the "exit" outcome; no "unknown choice" line is written for that selection.
- **Lower-bound and upper-bound selection**: any token outside the registered keys flows through the "unknown choice" path.
- **Repeated selections**: selecting the same non-exit item many times is allowed; the action runs each time.
- **Empty registry**: the menu must still print a header and offer the exit option; an empty registry must not crash the loop.
- **Input exhausted without exit token**: the loop ends with the "end of input" outcome; this is observable and is not a hang or an endless "unknown choice" loop.
- **I/O failure**: if a write to the injected writer fails, the program should stop cleanly without swallowing the error or panicking.

## 11. Project Constraints

- Go standard library only. No third-party packages, no command-line flag library required, no testing helpers beyond what `testing` provides.
- The action functions are stored as values. The menu loop must not embed a per-action `switch` over action names.
- Reading and writing must go through parameters typed as `io.Reader` and `io.Writer`. Direct use of `os.Stdin` or `os.Stdout` inside the menu loop is a design failure for this project, even if the production wiring still hands in the real streams.
- All behavior required by the verification section must be reachable from a test that uses buffers rather than a terminal, including the "end of input" outcome.
- No persistence, no configuration files, no environment-variable parsing, no signal handling — out of scope for this project.
- No reliance on terminal-specific escape sequences; output is plain text.

## 12. Design Questions Before Coding

Answer these on paper or in your own notes before writing any Go file.

- What type describes a single menu item? Label and the action value, and anything else you want to show — how do you express those together so the registry can carry them?
- Should the registry key be a string like `"1"` or an integer? What changes about the unknown-choice path when you pick one over the other?
- Where does the loop live? In `main` or in a function that takes the registry and the I/O pair so `main` only wires production streams and the tests wire buffers?
- How does the loop recognize end-of-input from a generic `io.Reader`? Where does that branch live so a buffer-driven test can hit it?
- Who owns the exit sentinel — the registry or each action? What happens if two actions both claim the exit key?
- What does the action return? A value, a write to the injected writer, or both? How does the menu know the action succeeded?
- How will a test prove that "Exit" really exits, that "end of input" really exits, and that the two outcomes are distinguishable?

## 13. Implementation Milestones

Each milestone is a behavior the learner can demonstrate. Do not move on until the current one is observable.

1. Define a menu item type that pairs a label and an action value in one structure.
2. Build a registry that holds several items, including the exit item, in a `map`.
3. Write a printing routine that writes the menu to an `io.Writer` it receives.
4. Write a reading routine that reads one token at a time from an `io.Reader` and returns it trimmed, or signals that input is exhausted.
5. Write the loop that prints, reads, looks up, and either invokes or reports "unknown choice".
6. Make the loop end cleanly on the exit token.
7. Make the loop end cleanly when input is exhausted, even if the exit token was never selected.
8. Wire the production program to pass `os.Stdin` and `os.Stdout` into the loop.
9. Confirm that a test can substitute buffers for both streams and observe each menu print, each prompt, each action output, and the chosen termination outcome.

## 14. Verification Cases the Learner Must Write

Describe each case in natural language first; only then write the test. Every case must be reachable with buffers for input and output — no live terminal.

- Selecting a registered item invokes the corresponding action; the action's output appears in the captured writer.
- Selecting the exit item terminates the loop with the "exit" outcome; no "unknown choice" line is written for that selection.
- Selecting a token above the registered keys prints "unknown choice" and the loop continues.
- Selecting a non-numeric or non-token input prints "unknown choice" and the loop continues.
- An empty input line does not crash; it produces an "unknown choice" outcome and the loop continues.
- Two consecutive selections of the same item each produce the action's output, in order.
- A buffer that contains only selections and no exit token ends the loop with the "end of input" outcome; no infinite "unknown choice" loop, no hang.
- The same buffer that ends with an exit token ends the loop with the "exit" outcome, not the "end of input" outcome; the two outcomes are distinguishable in the captured writer.
- With a registry that contains only the exit item, the menu prints just that item and the loop exits on its selection.
- The action function for an item is stored as a value, not invoked at registration time: a test that registers an action which records "called" rather than "printed" can assert the action runs only when its key is selected.

## 15. Common Mistakes to Watch For

- **Long `switch` over actions inside the loop.** That hides the registry; if you find yourself writing `case "1": ... case "2": ...`, the design is wrong for this project.
- **Calling `os.Stdin` or `os.Stdout` directly inside the loop.** That blocks testing. Always pass the streams in.
- **Treating a missing key as "exit".** A missing key is an unknown choice; only the registered exit key exits.
- **Looping forever on end-of-input.** When a read returns the end-of-input signal, the loop must end; if the program keeps printing "unknown choice" past that point, the test will hang.
- **Panicking on a nil function.** If the registry allows nil actions, the menu must check before calling.
- **Mixing presentation with logic.** If the menu prints directly with `fmt.Println`, a test cannot replace the output stream; keep the writer as a parameter.
- **Buffering prompts but not menu prints.** Either buffer both or flush appropriately; mixed buffering leads to confusing test output.
- **Inventing behavior the plan did not ask for.** A retry count, a farewell message, a "press enter to continue" — those are out of scope unless the learner adds them as optional extensions.

## 16. Topics and References for Study

- A Tour of Go: "Function values" and "Maps".
- Effective Go: "Functions", "Data", and the section on I/O.
- Package documentation: `io` (`Reader`, `Writer`, `EOF`), `bufio` (`Scanner`, `Reader`), `fmt` (`Fprintf`, `Fscan`, `Fprintln`).
- Package documentation: `bytes` (`Buffer`) — used in tests as a stand-in for the real streams.
- Package documentation: `strings` — `TrimSpace` for cleaning input tokens.
- Search terms: "first-class functions Go", "map lookup zero value Go", "dependency injection testing stdin stdout", "io.Reader EOF in Go".

## 17. Self-Assessment Questions

1. Why is a function-without-parentheses a value in Go, and how does that let a menu store behavior as data?
2. What does a Go `map` return when the key is missing, and why must the menu handle that case explicitly before calling?
3. What does an `io.Reader` describe in one sentence, and what does an `io.Writer` describe?
4. Why does injecting the I/O pair matter for testing, even when the production program only ever passes the real streams?
5. What is the difference between "unknown choice", "exit token chosen", and "end of input", and where in the code is each distinction enforced?
6. What does the loop do when the read returns the end-of-input signal? Why is that branch as important as the exit-token branch?
7. Could the same registry be reused by a different program with a different header? What would have to change?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every numbered functional requirement in section 8 is satisfied by behavior the learner can demonstrate.
- Every verification case in section 14 has a corresponding test that uses buffers for input and output.
- No test depends on a real terminal, on real keyboard input, on a real sleep, or on the local clock.
- The loop has no `switch` over action names; actions are dispatched through a `map` lookup.
- The menu and prompts are written through an injected `io.Writer`; input is read from an injected `io.Reader`.
- The "end of input" outcome is exercised by at least one test that does not supply the exit token.
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Action results as data.** Let an action return a small result value that the menu loop logs through the writer, so the loop can react (for example, by printing a confirmation line) without knowing what the action did.
- **Sub-menus by registration.** Allow an action to register additional items before returning, so the next iteration of the loop sees a longer menu. Keep the extension tiny: no file-based configuration, no nested loops beyond one level.
