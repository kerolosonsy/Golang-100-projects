# Project 093 — Custom Interpreter and Evaluator

## 1. Project Name and Number
Project 093, `093_custom_interpreter_eval`. The folder name is fixed by the curriculum table; do not rename the directory.

## 2. Project Idea
A small interpreter for a tiny integer language. The language supports 64-bit integer literals, identifiers, assignment statements, expression statements, the binary operators plus, minus, asterisk, and slash, the unary plus and unary minus, and parentheses. Statements are separated by newlines or by semicolons under one pinned policy. The lexer emits tokens with a 1-based line and a 1-based rune-column position. The parser is a recursive-descent parser that builds a small AST. The evaluator runs the AST against an in-memory environment that persists within a single evaluation session, performs checked arithmetic, and produces stable error messages with positions.

## 3. Why This Project Now?
This project is the parsing and interpretation capstone for the early language work. It pulls the discipline of splitting a language into tokens, into an AST, and into values, with errors at every stage, into one program. The previous project pulled together interface narrowness and a CLI safety model; this project pulls together token positions, recursive-descent grammar, and checked arithmetic. The formal prerequisites — project 006 (Unicode-aware string handling) and project 028 (recursive data structure) — each contribute one ingredient that this interpreter combines. The immediate catalog predecessor, project 092, is optional context rather than a formal prerequisite.

## 4. Prerequisites
The formal prerequisites are projects 006 and 028; project 092 is the immediate catalog predecessor and remains useful as optional context rather than a formal prerequisite.

## 5. What You Must Know Before Starting
- That Go's `string` is a sequence of bytes and indexing it does not yield a Unicode code point; the lexer must decide whether it scans bytes, runes, or both.
- That integer literals in source text can overflow `int64` at parse time and that overflow is a lexer error, not a runtime panic.
- That a recursive-descent parser is a set of mutually recursive functions, one per grammar production, each consuming the tokens of its production and leaving the rest to the caller.
- That precedence in a grammar is encoded by the nesting of the parsing functions; tighter precedence is deeper in the call chain.
- That an environment is a mapping from identifier names to values, with assignment that mutates an existing binding or creates a new one.
- That evaluation order in expressions is left-to-right within a precedence level under the left-associative rule.
- That "no implicit multiplication" is a deliberate policy; `2(3+4)` is a syntax error.

## 6. Explanation of New Concepts
- Token as a typed value with a position: each token carries a kind (literal, identifier, plus, minus, asterisk, slash, left paren, right paren, assign, semicolon, newline, end of input, or a single error kind for unknown characters and invalid UTF-8), the literal text, a 1-based line, and a 1-based rune column. The last token is always the end-of-input token.
- Pinned lexical policy: both newline and semicolon are statement separators. Repeated separators and blank lines are ignored. Spaces and tabs are insignificant. CRLF counts as one newline. Line comments begin with `//` and end just before the newline that terminates them, so that terminating newline still separates statements. A bare carriage return that is not part of CRLF is invalid. No other comment syntax exists.
- ASCII identifier policy: the first character is an ASCII letter or underscore; the remaining characters are ASCII letters, ASCII digits, or underscore. A character outside this set is a lexical error, including non-ASCII letters. Unknown characters are lexical errors at the current position.
- Invalid UTF-8 policy: a byte sequence that is not valid UTF-8 is a lexical error at the current 1-based line and rune column.
- Decimal literal policy: a decimal digit literal is a non-negative magnitude consisting of one or more ASCII digits from zero through `MaxInt64`. A longer digit sequence is a lexical overflow error at the literal's position. Negative values are unary expressions; `MinInt64` is produced by checked expression evaluation rather than by accepting an out-of-range positive magnitude at lex time.
- Recursive-descent grammar: the grammar has productions for statements, expressions, terms, factors, and unary expressions. Unary binds tightest, then multiplication and division, then addition and subtraction, with left associativity at each level. Assignment is a statement, not an expression; only an identifier may appear on the left.
- AST as a small set of node types: a literal node, an identifier node, a binary node, a unary node, and an assignment statement node. Each node carries its source position.
- Checked arithmetic: every arithmetic operation on `int64` is checked for overflow, division by zero, and the special case of `MinInt64 / -1`. Each check produces a runtime error with the position pinned below.
- Pinned diagnostic positions: binary arithmetic errors point at the binary operator's position. Unary overflow errors point at the unary operator's position. Undefined variable errors point at the identifier's position. Missing close parenthesis errors point at the end-of-input token or the current token at the diagnostic moment.
- Evaluation under first-error semantics: the evaluator stops at the first diagnostic. Statements and assignments that completed successfully before the failure remain in the session environment and are observable on the next evaluation. The failing assignment makes no mutation, and no partial expression output is returned for the failing statement.

## 7. Learning Objective
After completing this project you must be able to explain in your own words: why lexer and parser are separate stages, how precedence is encoded by parser structure, why left associativity is a property of the grammar, why assignment is a statement and not an expression, why implicit multiplication is rejected, why checked arithmetic matters, why the environment persists across statements within a session, why evaluation stops at the first diagnostic and why prior successful state survives a later failure, and why each diagnostic class uses a pinned position.

## 8. Functional Requirements
1. The lexer accepts a UTF-8 source string and produces a token stream ending with the end-of-input token.
2. The lexer rejects invalid UTF-8, unknown characters, non-ASCII identifiers, and digit sequences longer than the maximum representable magnitude. Each rejection is a diagnostic with the offending position.
3. The lexer recognizes both newline and semicolon as statement separators, treats repeated separators and blank lines as insignificant, ignores spaces and tabs, treats CRLF as one newline, and treats `//`-to-end-of-line as a line comment so that the trailing newline still separates statements. A bare carriage return is invalid. No other comment syntax exists.
4. The parser accepts a token stream and produces an AST. The AST is a list of statements; each statement is either an expression statement or an assignment statement.
5. The parser enforces precedence: unary binds tighter than multiplication and division, which bind tighter than addition and subtraction, all left-associative.
6. The parser rejects implicit multiplication: a primary followed by another primary with no operator is a syntax error.
7. The parser rejects assignment to anything that is not a single identifier on the left.
8. The parser reports syntax errors with position and never panics on any input.
9. The evaluator accepts an AST and an environment, executes statements in order, and returns the final environment and the list of expression-statement results produced by successfully evaluated statements before the first diagnostic.
10. The evaluator performs checked `int64` arithmetic. Overflow, division by zero, and `MinInt64 / -1` are runtime errors with the positions pinned below.
11. The evaluator treats an undefined identifier as a runtime error at the identifier's position.
12. The evaluator mutates the environment only after the right-hand side of an assignment has succeeded. A failed right-hand side leaves the prior binding unchanged and produces no assignment diagnostic for that statement.
13. The interpreter exposes a single entry point that runs lex, parse, and evaluate and returns a structured result with diagnostics and the surviving session state.
14. Diagnostics carry a 1-based line and a 1-based rune column, and follow the pinned position classes.

## 9. Inputs and Outputs
- Input: a source string. Valid program examples:
  - Expression statement: `1 + 2 * 3`.
  - Assignment and use: `x = 10`, `y = x + 1`, `x + y`.
  - Unary: `-5`, `-x`, `+(1 - 2)`.
  - Parentheses: `(1 + 2) * 3`.
  - Multi-statement separated by newlines, by semicolons, or by any mix permitted by the pinned lexical policy: `x = 1`, newline, `y = 2`, newline, `x + y`.
- Output: a structured result with the list of expression-statement values from successfully evaluated statements before the first diagnostic, and an empty diagnostic list on full success. On failure, the diagnostic list contains the first error with its position, the surviving session environment is returned, and no partial expression value is reported for the failing statement.
- Successful expression value: the `int64` value of the last successfully evaluated expression statement.

## 10. Rules and Edge Cases
- Empty input is valid and produces an empty statement list and an empty value list.
- Whitespace-only input is valid and produces the same result as empty input.
- A line comment at the end of the file is valid.
- A magnitude at the boundary of `int64` is valid; a magnitude strictly greater than `MaxInt64` is a lexical overflow error at the literal's position.
- A division by zero is a runtime error at the binary operator's position.
- The expression `MinInt64 / -1` is a runtime error at the binary operator's position.
- An overflow in addition, subtraction, multiplication, or unary negation is a runtime error at the offending operator's position.
- An undefined identifier is a runtime error at the identifier's position.
- An assignment to an undefined identifier creates the binding when the right-hand side succeeds.
- A failed right-hand side leaves prior bindings unchanged and produces no assignment commit; the session environment remains observable for the next evaluation.
- A syntax error in the middle of a program produces a single diagnostic at the offending token; the interpreter does not attempt to recover and continue.
- A non-ASCII character in an identifier position is a lexer error at that character.
- Implicit multiplication such as `2(3+4)` is a syntax error at the second primary.
- CRLF is one newline separator. A bare carriage return not part of CRLF is a lexer error.

## 11. Project Constraints
- The language is integer-only. There are no strings, no booleans, no floats, no functions, no loops, no conditionals, no arrays, no records.
- The interpreter has no I/O, no file access, no network, and no environment variables beyond what the entry point accepts as arguments.
- The interpreter does not import a parser generator or a lexer generator. The lexer and the parser are hand-written.
- The interpreter does not import a big-integer library. Arithmetic is `int64` and is checked at the operation boundary.
- The AST is plain data and the evaluator is a separate stage.
- The environment is a small in-memory map with deterministic iteration order where the language exposes iteration.
- Diagnostics are deterministic for a given input. The order of diagnostics is the order in which they are produced.
- Unit tests must run locally with no external services. Fuzz tests are encouraged for the lexer and the parser.

## 12. Design Questions Before Coding
- Will your lexer scan bytes or runes, and how will you keep the rune column accurate across multi-byte characters?
- How will you define the environment value type and ensure deterministic behavior under persistence across statements?
- How will you propagate errors through the parser without using panic?
- How will you detect overflow for addition and multiplication without importing a helper?
- How will you detect `MinInt64 / -1` and report it at the operator position?
- How will you guarantee that a failing assignment makes no mutation?
- How will you guarantee that evaluation stops at the first diagnostic and that prior state survives for the next evaluation?
- How will your fuzz tests avoid panicking on adversarial input?

## 13. Implementation Milestones
1. Define the token kinds, the token value type, and the lexer entry point that returns a slice of tokens ending with the end-of-input token.
2. Implement integer magnitude parsing with overflow detection at the lex layer, identifier parsing with the ASCII policy, comment and whitespace skipping per the pinned policy, and invalid UTF-8 and bare carriage return reporting.
3. Define the AST node types and the parser entry point that returns a list of statements.
4. Implement the statement parser: assignment of the form `identifier = expression`, and expression statement.
5. Implement the expression parser cascade preserving precedence and left associativity: addition and subtraction over multiplication and division, over unary, over primary. Implement the implicit-multiplication rejection.
6. Implement the primary parser: integer magnitude, identifier, and parenthesized expression.
7. Define the environment value type with deterministic lookup and assignment, and the evaluator entry point.
8. Implement integer literal evaluation, identifier lookup, and assignment with the evaluate-then-mutate rule.
9. Implement binary and unary evaluation with checked arithmetic for plus, minus, asterisk, slash, and unary negation, including the `MinInt64 / -1` case.
10. Wire the interpreter entry point that runs lex, parse, and evaluate, returning diagnostics and the surviving session environment.
11. Write the unit test suite and the fuzz tests for the lexer and the parser.

## 14. Verification Cases the Learner Must Write
- Precedence: `1 + 2 * 3` evaluates to 7; `2 * 3 + 1` evaluates to 7; `-2 * 3` evaluates to -6.
- Associativity: `10 - 3 - 2` evaluates to 5, not 9; `20 / 4 / 5` evaluates to 1, not 25.
- Parentheses: `(1 + 2) * 3` evaluates to 9.
- Unary: `-5 + 3` evaluates to -2; `-(2 + 3)` evaluates to -5.
- Variables: `x = 4`, `x + 1` evaluates to 5; `y = x`, `y` evaluates to 4.
- Reassignment: `x = 1`, `x = 2`, `x` evaluates to 2.
- Multi-statement: newline-, semicolon-, CRLF-, and mixed-separator programs produce the expected results where the pinned policy applies.
- Invalid token: a non-ASCII character in identifier position produces a lexer error at that character.
- Position: a syntax error on line 2 column 5 is reported with line 2 and rune column 5.
- Missing paren: `(1 + 2` is a syntax error at the end-of-input token or the current token at the diagnostic moment.
- Undefined variable: a reference to an undeclared identifier is a runtime error at the identifier's position.
- Division by zero: `1 / 0` is a runtime error at the binary operator's position.
- Overflow: an expression that overflows `int64` is a runtime error at the operator's position.
- Failed assignment no mutation: an assignment with a failing right-hand side leaves the prior binding unchanged; the next evaluation sees the prior state.
- First-error semantics: an evaluation that hits one diagnostic does not produce a partial expression value for the failing statement, returns the surviving session environment, and stops further evaluation; a subsequent evaluation sees that surviving environment.
- Deterministic diagnostics: the same input produces the same diagnostic order and content.
- Fuzz: a fuzz test on the lexer and a fuzz test on the parser never panic and always return either success or a structured diagnostic.

## 15. Common Mistakes to Watch For
- Letting the lexer panic on invalid UTF-8 or a bare carriage return.
- Letting the parser continue building an AST on a lex error.
- Using `int` instead of `int64`.
- Treating `MinInt64 / -1` as a single combined check with division by zero.
- Allowing implicit multiplication by accident.
- Letting an assignment mutate before the right-hand side succeeds.
- Reporting a runtime error at the statement position rather than at the operator, identifier, or current token pinned above.
- Using a non-deterministic `map` iteration where the language exposes iteration.
- Importing a parser generator or a big-integer library.
- Letting a runtime diagnostic cause a panic.
- Treating CRLF as two separators or rejecting CRLF; CRLF is one separator in this language.
- Treating a bare carriage return as part of a comment or as whitespace.

## 16. Topics and References for Study
- The Go Language Specification sections on integer types, constants, and overflow semantics.
- The `text/scanner` documentation, used as a reference for position tracking and token kinds; do not import it.
- The `go/ast` and `go/parser` packages, used only as references for AST node design; do not import them.
- A standard compiler textbook chapter on recursive-descent parsing and precedence climbing.
- The Go `testing/quick` package for property-based tests.

## 17. Self-Assessment Questions
- Why is precedence encoded by parser structure and not by the evaluator?
- Why is assignment a statement and not an expression in this language?
- Why is implicit multiplication rejected?
- Why does the assignment mutate only after the right-hand side succeeds?
- Why is `MinInt64 / -1` a separate check rather than an instance of division by zero?
- Why does the lexer carry a 1-based line and rune column, and what would break with 0-based positions?
- Why is the environment constructed once per evaluation session?
- Why does the interpreter produce a single diagnostic per evaluation and stop, rather than attempt to recover?
- Why does each diagnostic class use a pinned position rather than a position chosen at the diagnostic moment?
- Why is CRLF one separator and a bare carriage return invalid?

## 18. Definition of Completion
The project is complete when the lexer produces a token stream with 1-based positions and rejects invalid UTF-8, bare carriage returns, unknown characters, non-ASCII identifiers, and overflowing magnitudes with a structured diagnostic at the offending position; when the parser builds an AST for every valid program and produces a structured diagnostic on every invalid input without panicking; when the evaluator runs the AST, performs checked arithmetic, propagates runtime errors at the pinned positions, mutates the environment only on successful right-hand sides, persists the environment across statements within a session, stops at the first diagnostic, and leaves surviving state observable on the next evaluation; and when the unit test suite, including fuzz tests on the lexer and the parser, passes locally without external services.

## 19. Optional Extensions
- An additional binary operator such as modulo added with the same precedence and associativity rules; the lexer, parser, evaluator, and tests are updated together; the language still has no implicit multiplication.
- A block statement with braces whose body is evaluated in a fresh environment; a failed inner assignment does not mutate the outer environment, but a successful inner assignment remains observable within the block.
