# Project 028 — Binary Search Tree

## 1. Project Name and Number

- Project **028** — `028_binary_search_tree`.
- The directory name and number must match exactly.
- The project implements a generic binary search tree whose ordering is driven by an explicit comparator supplied by the caller.
- The tree supports insert, search, and in-order traversal.
- Duplicates are ignored: an attempt to insert a value the comparator considers equal to an existing value does not add a new node, the operation reports that no new node was added, and the size does not change.
- Balancing is out of scope.

## 2. Project Idea

The tree stores values whose order is defined by the comparator. The comparator is a function the caller provides when the tree is constructed (or, with a zero value, before the first insertion). The comparator returns a negative, zero, or positive integer that defines a consistent total order over the element type. The tree uses that comparator to decide where a new value belongs. Two values that the comparator considers equal are duplicates; the tree treats them as a single value and ignores the second insertion.

Insertion walks the tree from the root, comparing the new value to each node's value. When it finds a node whose value compares equal, insertion reports that no new node was added and returns. When it finds a position where the new value should go (an empty child slot), insertion creates a new node there and increments the size.

Search walks the tree from the root, comparing the target value to each node's value. A match (comparator returns zero) reports a hit; a miss (the walker reaches a `nil` child slot without a match) reports a miss.

In-order traversal walks the tree and produces a sequence of values in comparator order. The traversal returns an independent result that does not mutate the tree.

The project is honest about its scope: the tree is not balanced. A degenerate insertion order (for example sorted or reverse-sorted) produces a degenerate shape. Balancing is out of scope; the project does not claim balanced-tree guarantees.

## 3. Why This Project Now?

- Projects 001–027 established variables, functions, loops, structs, errors, slices, files, JSON, CSV, scanning, sorting, walking, hashing, shape-validated matrices, and generic zero-value containers.
- None of them implemented a recursive data structure with an externally supplied ordering rule.
- Project 028 is the first project that combines recursion, generics, and a comparator-driven invariant.

- The comparator pattern is reused elsewhere in the path (for example in project 069's text search ranking).
- The duplicate-ignore rule and the in-order-traversal contract are the simplest expression of "the comparator defines the order, the tree follows it".
- Project 028 is the project's first encounter with a data structure whose correctness depends on a property (the comparator invariants) supplied by the caller, not by the type system.

- Project 028 also forces the learner to be honest about an unbalanced BST.
- A degenerate insertion order is a worst case the project acknowledges, not a hidden surprise.
- Balancing is a separate concern.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 028 therefore requires:

- Completion of **027** (Custom Stack and Queue). Earlier projects (for example 022's interface discipline, 016's slice discipline, 014's validation discipline) are background concepts already encountered and may inform the learner's design, but they are not additional required completions for this project.
- No prior knowledge of HTTP, databases, or concurrency.

## 5. What You Must Know Before Starting

- That a binary search tree is a recursive structure: each node holds a value, a left child, and a right child. Children are themselves trees or `nil`. Operations are written as recursive functions or as iterative loops that mirror the recursion.
- That a comparator is a function the caller provides. The comparator returns a negative value when `a < b`, zero when `a == b`, and a positive value when `a > b`, where `<`, `==`, and `>` are the order the comparator defines. The comparator must define a consistent total order.
- That "consistent total order" means the comparator obeys the rules that make a total order work: antisymmetry, transitivity, totality (every pair of distinct values is comparable), and consistency of the returned sign with the order. A comparator that violates transitivity breaks the BST invariant silently and is a learner's responsibility to avoid.
- That `errors` and `fmt.Errorf` produce errors the caller can match. The project's not-found and duplicate-ignored outcomes use sentinel errors or clearly-named booleans.
- That the standard library's `sort.Slice` and the `cmp` package's `Compare` function exist and can be a reference for how a comparator's return value is interpreted. The tree does not depend on `sort` for its operations, but the comparator's convention matches the standard library's.
- That recursion on a tree is straightforward to write but easy to get wrong on the base cases. The empty-tree case (`nil`) must be handled before any access to a node's fields.

## 6. Explanation of New Concepts

### Concepts

#### Comparator-driven order

The BST's correctness depends on the comparator's invariants, not on the element type's built-in operators. Two values are "equal in the tree's order" when the comparator returns zero. Two values are "less" or "greater" when the comparator returns a negative or positive value. The tree stores and retrieves values according to the comparator's verdicts. If the comparator is inconsistent (for example, it returns zero for some pair on one call and a negative value on another call), the tree's behavior is undefined. The project pins this responsibility on the caller.

#### Insertion with a recursive walk

Insertion starts at the root. If the tree is empty, the new value becomes the root and the size becomes one. If the tree is not empty, insertion compares the new value to the current node's value. A "less" verdict recurses into the left subtree. A "greater" verdict recurses into the right subtree. A "equal" verdict reports that no new node was added and returns without changing the tree or the size. When the recursion reaches a `nil` child slot, a new node is created there and the size is incremented.

#### Duplicate-by-comparator policy

A duplicate-by-comparator value is a value the comparator considers equal to an existing node's value. The tree ignores duplicates: it does not add a second node, does not change the size, and does not change any existing node. The insertion operation reports the outcome through a clearly-named boolean or a sentinel error. The caller can distinguish "I added a new value" from "I tried to add a duplicate and the tree did not change".

#### Search

Search starts at the root. If the tree is empty, search reports a miss. If the tree is not empty, search compares the target value to the current node's value. A "equal" verdict reports a hit. A "less" verdict recurses into the left subtree. A "greater" verdict recurses into the right subtree. Reaching a `nil` child slot without a match reports a miss.

#### In-order traversal

In-order traversal visits the left subtree, then the current node, then the right subtree. The result is a sequence of values in comparator order. The traversal is independent of the tree: the returned sequence does not mutate the tree and is a fresh slice. In-order traversal on an empty tree returns an empty slice and no error.

#### Independent result, no mutation

The in-order traversal returns a fresh slice. A test that mutates the returned slice and re-runs the traversal observes the original values in order. A test that runs the traversal twice in a row observes the same values in the same order.

#### Size

Size is the number of nodes in the tree. The size starts at zero (empty tree). Each successful insertion (not a duplicate) increments size by one. Duplicates do not change size. The size accessor returns the current value. The size accessor is consistent with the actual node count.

#### Comparator invariants

The comparator defines a total order over the element type. The invariants are:

- **Antisymmetry.** If the comparator returns a negative value for `(a, b)`, it returns a positive value for `(b, a)`. If it returns zero for `(a, b)`, it returns zero for `(b, a)`.
- **Transitivity.** If the comparator returns a negative value for `(a, b)` and a negative value for `(b, c)`, it returns a negative value for `(a, c)`. The analogous rule holds for positive and zero.
- **Totality.** For any two values `a` and `b`, exactly one of "negative", "zero", or "positive" is returned. The comparator does not return "incomparable" or refuse to compare.
- **Consistency of sign.** The sign of the comparator's return value is what matters; the magnitude is not part of the order. A comparator that returns `1` for some pairs and `2` for others is fine as long as both are positive.

A comparator that violates any invariant produces an undefined tree. The project does not validate the comparator at runtime; the caller is responsible.

#### Multiple element types

The tree is generic through type parameters. Tests cover at least three distinct element types — for example `int`, `string`, and a small struct with a comparator that compares one of the struct's fields. The same implementation works for all three types without per-type code.

#### Ascending and descending comparators

The comparator can define ascending or descending order. The same tree, given an ascending comparator for `int`, stores and retrieves values in ascending order. The same tree, given a descending comparator for `int`, stores and retrieves values in descending order. The in-order traversal reflects whichever order the comparator defines.

#### Empty tree safety

Search and traversal on an empty tree are safe and well-defined. Search returns a miss. In-order traversal returns an empty slice. Size returns zero.

#### Degenerate insertion and balancing is out of scope

Inserting values in sorted (or reverse-sorted) order produces a degenerate shape: every node has at most one non-nil child, and the tree is effectively a linked list. The project is honest about this: the tree is not balanced, and a degenerate insertion is a worst case. Balancing (rotation, randomized insertion, red-black invariants) is out of scope.

## 7. Learning Objective

After completing this project the learner can:

- Design a generic BST driven by an externally supplied comparator, with the comparator defining a consistent total order.
- Explain the pinned missing-comparator policy: empty tree without a comparator supports safe search and traversal; insertion requires a non-`nil` comparator and returns a clear error without mutating the tree; `nil` comparator configuration is rejected; the comparator cannot be replaced once values exist.
- Implement insertion that places a new value at the correct child slot and ignores duplicates (values the comparator considers equal to an existing node's value).
- Implement search that returns a hit when the target equals a node's value by the comparator and a miss otherwise.
- Implement in-order traversal that returns values in comparator order and returns an independent result that does not mutate the tree.
- Handle the empty-tree case explicitly: search returns a miss, traversal returns an empty slice, size is zero.
- Reason about comparator invariants (antisymmetry, transitivity, totality, sign consistency) and explain why a comparator that violates them produces an undefined tree.
- Pin the duplicate-ignore policy as part of the contract: a duplicate insertion reports "not added" and leaves size unchanged.
- Use the same generic implementation across multiple element types, including ascending and descending comparators.
- Acknowledge that the tree is not balanced and that a degenerate insertion order produces a degenerate shape.
- Write tests that pin insert/search hit/miss, duplicate behavior, empty tree, ascending and descending comparators, in-order order, size transitions, negative/mixed values as appropriate, and no mutation from returned traversal data.

## 8. Functional Requirements

1. The package defines a generic BST type. The tree's ordering is driven by a comparator supplied by the caller. The tree's zero value is empty and supports safe search and in-order traversal (returning a miss and an empty slice respectively) without a configured comparator. Insertion requires a non-`nil` comparator.
2. Insertion places a new value at the correct child slot when a non-`nil` comparator is configured. If the comparator is `nil`, insertion returns a contextual error and does not mutate the tree. If the comparator considers the new value equal to an existing node's value, insertion reports that no new node was added and does not change the tree or the size. Insertion never panics.
3. The duplicate-ignore policy is pinned: duplicates are ignored, size does not change, and the operation reports the outcome through a clearly-named boolean or a sentinel error.
4. Search returns a hit when the target value compares equal to a node's value, and a miss otherwise. Search on an empty tree returns a miss.
5. In-order traversal returns a sequence of values in comparator order. The returned sequence is a fresh slice that does not mutate the tree. In-order traversal on an empty tree returns an empty slice.
6. Size returns the current number of nodes. Size starts at zero (empty tree). Each successful insertion increments size by one. Duplicate-ignored insertions do not change size.
7. The comparator defines a consistent total order. The package documentation states the comparator's invariants (antisymmetry, transitivity, totality, sign consistency) and that the tree's correctness depends on them.
8. The tree works for any element type supplied at the call site, with the comparator's element type matching the tree's element type. The implementation does not bake in a single type.
9. The tree is not balanced. A degenerate insertion order (sorted, reverse-sorted) produces a degenerate shape. Balancing is out of scope, and the package documentation states this rule.
10. Empty-tree operations (search and in-order traversal) are safe and well-defined.
11. The in-order traversal does not mutate the tree. A test that mutates the returned slice and re-runs the traversal observes the original values.
12. The package documentation states the comparator invariants, the duplicate-ignore policy, the in-order traversal contract, the empty-tree contract, and the no-balancing rule.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- For insertion: a value of the tree's element type.
- For search: a target value of the tree's element type.
- For in-order traversal: no value.
- For size: no value.
- The comparator is supplied at construction (or with a zero-value tree, set explicitly before the first non-empty use).

#### Outputs

- For insertion: a clearly-named boolean or a sentinel error indicating whether a new node was added. On a duplicate, the boolean is "not added" (or the sentinel error is returned) and size is unchanged. On a new value, the boolean is "added" (or no error) and size grows by one.
- For search: a hit or miss outcome through a clearly-named boolean, an `errors.Is`-matchable error, or a comparable result.
- For in-order traversal: a fresh slice of values in comparator order. Empty tree returns an empty slice.
- For size: the current node count as an integer.

#### Example text-only traces

Tree with ascending `int` comparator (returns negative if `a < b`):

```
insert 5 → added, size=1, tree: [5]
insert 3 → added, size=2, tree: [5; 3, _]
insert 7 → added, size=3, tree: [5; 3, 7]
insert 3 → not added, size=3 (duplicate)
insert 8 → added, size=4, tree: [5; 3, 7; _, 8]
search 7  → hit
search 4  → miss
in-order  → [3, 5, 7, 8]
```

Tree with descending `int` comparator (returns positive if `a < b`):

```
insert 5 → added, size=1
insert 3 → added, size=2
insert 7 → added, size=3
in-order  → [7, 5, 3]
```

## 10. Rules and Edge Cases

- **Comparator is required for insertion.** An empty tree without a configured comparator supports safe search (returns a miss) and in-order traversal (returns an empty slice). Insertion on a tree without a non-`nil` comparator returns a contextual error and does not mutate the tree. Insertion never panics. Configuring a `nil` comparator as the comparator is rejected with a contextual error.
- **Empty tree search.** Returns a miss. No error.
- **Empty tree traversal.** Returns an empty slice. No error.
- **Empty tree size.** Returns zero.
- **Single element.** Insertion grows size to one. Search for the inserted value is a hit; search for any other value is a miss. In-order traversal returns a slice with one element.
- **Duplicate insertion.** The tree does not change. Size does not change. The outcome is reported as "not added".
- **Sorted insertion order.** Inserting values in ascending comparator order produces a degenerate shape: each node has only a right child (or only a left child, for descending order). The project does not pretend this is balanced.
- **Reverse-sorted insertion order.** Same as sorted, in the opposite direction.
- **Ascending comparator.** In-order traversal returns values in ascending comparator order.
- **Descending comparator.** In-order traversal returns values in descending comparator order.
- **Multiple element types.** The same generic implementation works for `int`, `string`, and a struct (with a comparator that compares a chosen field). The duplicate-ignore policy applies identically across types.
- **Negative, decimal, mixed values.** Treated as ordinary inputs. No special handling. The comparator's invariants apply.
- **Comparator invariants.** The comparator must define a consistent total order. A comparator that violates the invariants produces an undefined tree. The project pins the invariants through documentation and tests but does not validate the comparator at runtime.
- **In-order traversal independence.** Mutating the returned slice does not change the tree. A subsequent traversal returns the original values.
- **Size consistency.** Size matches the actual node count after every insertion.

## 11. Project Constraints

- Go standard library only. No third-party tree libraries, no `google/btree`, no `golang.org/x/exp/constraints`.
- The tree is generic through type parameters. The implementation works for any element type without per-type code.
- The comparator is supplied by the caller. The implementation does not bake in a single ordering.
- Duplicates are ignored. The tree does not store multiple nodes with comparator-equal values.
- The tree is not balanced. Rotation, randomized insertion, red-black, AVL, and splay trees are out of scope.
- Concurrency safety is out of scope.
- Core logic is testable without terminal, real user directories, network, or any external service.
- The comparator's invariants are not validated at runtime. The caller is responsible.
- The in-order traversal returns an independent result and does not mutate the tree.

## 12. Design Questions Before Coding

- How is the comparator held? As a function-typed field, as an interface, or both? Which choice keeps the comparator's role explicit?
- How is the zero-value tree represented? The zero value is empty and supports safe search and traversal. Insertion requires a non-`nil` comparator; a constructor is not required. How is the comparator supplied without giving up the zero-value usability of the empty tree?
- How is the duplicate-ignore outcome exposed? As a boolean paired with the inserted value, as a sentinel error returned by the insertion operation, or as a typed result? Which choice is most discoverable for a caller reading the package?
- How is search's hit/miss outcome exposed? As a boolean, as an `errors.Is`-matchable error, or as a comparable result?
- How is the in-order traversal implemented? Recursively, with an iterative stack, or with a Morris-style traversal? Which choice is straightforward to test and matches the project's recursion focus?
- How is size tracked? As a stored field on the tree, as a recursive walk on demand, or both? A stored field is the conventional choice and makes size a constant-time accessor.
- How is the empty-tree case handled? As a `nil` root, as a tree struct with a `nil` root, or as a tree struct with an explicit `isEmpty` flag? The empty case must be handled before any access to a node's fields.
- How is the comparator's invariants documented? In the package comment, in the type's doc comment, or both?
- How is the no-balancing rule stated? In the package comment, in the type's doc comment, or both?

## 13. Implementation Milestones

1. Decide the package layout. Keep the tree as a generic type with a node type and a root pointer. Keep `main` as a thin driver that inserts a small script of values and prints the in-order result.
2. Pin the public contract as named constants or sentinel errors: the comparator requirement, the duplicate-ignore policy, the in-order traversal contract, the empty-tree contract, and the no-balancing rule.
3. Implement the node type. A node holds a value, a left child, and a right child. Both children start as `nil`.
4. Implement insertion. Compare the new value to the current node's value using the comparator. Recurse into the left subtree on a "less" verdict, the right subtree on a "greater" verdict. On a "equal" verdict, report "not added" and return without changing the tree. On reaching a `nil` child slot, create a new node and increment size.
5. Implement search. Compare the target value to the current node's value. Report a hit on "equal", recurse left on "less", recurse right on "greater". Report a miss on a `nil` child slot.
6. Implement in-order traversal. Recurse left, visit the current node, recurse right. Build a fresh slice. Return the slice.
7. Implement size. Track the node count in a stored field. Increment on successful insertion; do not change on duplicate-ignored insertion.
8. Verify comparator invariants through documentation. The package comment states antisymmetry, transitivity, totality, and sign consistency.
9. Verify the no-balancing rule through documentation. The package comment states that the tree is not balanced.
10. Verify the empty-tree contract. Insert on an empty tree creates the root. Search on an empty tree returns a miss. In-order traversal on an empty tree returns an empty slice. Size on an empty tree returns zero.
11. Wire `main`. The driver inserts a small script of values and prints the in-order result. The driver is not part of the package's public contract.
12. Add tests for every verification case in section 14, with insert/search tests, duplicate tests, empty-tree tests, comparator-order tests, in-order tests, size tests, multi-type tests, and no-mutation tests separated.

## 14. Verification Cases the Learner Must Write

### Required Cases

Each case is described in natural language. Tests use only in-memory elements and direct API calls; no terminal, no real user directories.

#### Insert and search

- Insert a value into an empty tree. Size becomes one. The tree's root is that value.
- Insert a second value that is "less" by the comparator. Size becomes two. The new value is the root's left child.
- Insert a second value that is "greater" by the comparator. Size becomes two. The new value is the root's right child.
- Search for an inserted value returns a hit.
- Search for a value not in the tree returns a miss.
- Search for a value in a tree with many nodes returns a hit when the value is present and a miss when it is not.
- Search on an empty tree returns a miss.

#### Duplicate behavior

- Insert a value, then insert the same value again. The second insertion reports "not added". Size does not change.
- Insert a value, then insert a comparator-equal value (for example a different `string` that the comparator considers equal to the first). The second insertion reports "not added". Size does not change.
- After several duplicates, the tree's structure and size are unchanged from the state before the duplicates.
- Insert duplicates in different positions (left subtree, right subtree, root). The duplicate-ignore rule applies at each position.

#### Empty tree

- An empty tree has size zero.
- Search on an empty tree returns a miss. No panic.
- In-order traversal on an empty tree returns an empty slice. No panic.
- Insertion on an empty tree creates a root and sets size to one.

#### Missing comparator (pinned policy)

- An empty tree without a configured comparator supports safe search (returns a miss) and in-order traversal (returns an empty slice). No panic.
- Insertion on a tree without a non-`nil` comparator returns a contextual error. The tree is not mutated. Size remains zero. No panic.
- Configuring the comparator with a `nil` comparator is rejected with a contextual error. The tree is not mutated. No panic.
- Once the tree holds any value, an attempt to replace the comparator returns a contextual error. The tree is not mutated. Size and existing nodes are unchanged. No panic.
- A test pins each of these outcomes directly. The tests do not exercise "comparator is optional", "comparator can be swapped after first insertion", or "panic on missing comparator". Those are not valid policies.

#### Ascending and descending comparators

- With an ascending `int` comparator, inserting `5, 3, 7, 3, 8` produces an in-order traversal of `[3, 5, 7, 8]`.
- With a descending `int` comparator (returns positive when `a < b`), inserting the same values produces an in-order traversal of `[8, 7, 5, 3]`.
- With an ascending `string` comparator, inserting values in mixed order produces an in-order traversal in lexicographic ascending order.
- The same tree, given different comparators at construction, reflects the comparator's order in its in-order traversal.

#### In-order order

- In-order traversal of a tree with a non-trivial shape (for example, a root with two subtrees, each with several nodes) returns values in comparator order.
- In-order traversal of a degenerate tree (sorted insertion) returns values in comparator order despite the shape.
- In-order traversal of a degenerate tree (reverse-sorted insertion) returns values in comparator order despite the shape.

#### Size

- After each successful insertion, size grows by exactly one.
- After a duplicate-ignored insertion, size is unchanged.
- Size matches the actual node count of the tree (a test can verify by counting nodes through a recursive walk or by relying on the in-order traversal's length as a proxy, depending on the package's surface).
- Size starts at zero for a freshly declared tree.

#### Negative and mixed values

- Inserting negative, zero, and positive `int` values produces an in-order traversal that includes all of them in the correct order.
- Inserting mixed-case strings produces an in-order traversal in the comparator's order.
- The duplicate-ignore rule applies to negative, zero, and positive values equally.

#### Multiple element types

- The same generic implementation works for `int`, `string`, and a struct with a comparator that compares a chosen field.
- For each type, insert and search behave identically.
- For struct element types, the comparator compares the chosen field and the in-order traversal reflects that field's order.

#### No mutation from returned traversal data

- Mutating the slice returned by in-order traversal does not change the tree's structure or values.
- A subsequent in-order traversal returns the original values in the original order.
- A subsequent search returns the same hits and misses as before the mutation.

#### Comparator invariants (documentation)

- The package documentation states the comparator's invariants (antisymmetry, transitivity, totality, sign consistency).
- The package documentation states that the tree's correctness depends on the comparator's invariants and that the implementation does not validate them.
- A test exercises a comparator that obeys the invariants and observes correct behavior; the test does not exercise a comparator that violates the invariants (that case is documented as undefined behavior).

#### No-balancing declaration

- The package documentation states that the tree is not balanced and that a degenerate insertion order produces a degenerate shape.
- The required scope does not test balancing. A test that inserts in sorted order and observes the degenerate shape is part of the required scope and pins the no-balancing rule's observable consequence.

#### Process

- A test runs the driver with a small script of insertions and confirms the printed in-order traversal matches the expected text-only form.

## 15. Common Mistakes to Watch For

- **Baking in a single ordering.** Using `<` or `>` on the element type instead of the comparator makes the tree non-generic and breaks the contract.
- **Storing duplicates.** A tree that stores two nodes with comparator-equal values breaks the duplicate-ignore rule. The comparator's zero verdict must lead to "do nothing".
- **Mutating size on a duplicate.** A duplicate-ignored insertion that increments size is wrong. Size changes only on a successful insertion.
- **Returning a slice that aliases internal nodes.** In-order traversal must return a fresh slice. Returning a slice that aliases the tree's internal storage is wrong and breaks the no-mutation contract.
- **Panicking on an empty tree.** Search and traversal on an empty tree are safe. A nil-dereference panic is wrong.
- **Forgetting to handle the empty-tree case.** Insertion on an empty tree must create the root. Search on an empty tree must return a miss without recursing into a nil node's children.
- **Confusing comparator-equal with element-equal.** A struct with multiple fields can be comparator-equal on one field and not-equal on another. The tree uses the comparator, not the element's equality operator.
- **Claiming balanced behavior.** The tree is not balanced. Comments or documentation that claim rotation or rebalancing are out of scope.
- **Validating the comparator at runtime.** The project does not validate the comparator. A test that tries to detect a buggy comparator through runtime checks is out of scope.
- **Using a recursive traversal that mutates the tree.** In-order traversal must not change the tree's structure.
- **Returning a hit on a value the comparator considers "less" or "greater".** A hit only happens on a comparator-zero verdict.
- **Returning a miss on a value the comparator considers equal.** A miss on a comparator-zero verdict is wrong.
- **Producing an in-order traversal that is not in comparator order.** A traversal that visits the right subtree before the current node, or that visits the current node twice, is wrong.
- **Allowing the comparator to be replaced once values exist.** Once the tree holds any value, replacing the comparator would invalidate the BST invariant because the existing nodes were placed using the old comparator's order. The package pins this: after a successful insertion, attempts to replace the comparator are rejected with a contextual error and the tree is not mutated. The implementation does not allow "comparator is optional after first insertion" or "comparator can be swapped" as valid policies.

## 16. Topics and References for Study

- A Tour of Go: "Generics", "Errors", "Recursion".
- Effective Go: "Generics", "Data", "Errors".
- Package documentation: `errors` (New, Is), `fmt` (Errorf, %w), `cmp` (Compare, Less), `sort` (Slice).
- Comparator patterns: search for "Go comparator function generic", "Go total order comparator", "Go sort comparator convention".
- BST patterns: search for "Go binary search tree generic", "Go BST insert search traverse", "Go recursive tree in-order".
- Comparator invariants: search for "Go comparator antisymmetry transitivity", "Go total order definition", "Go comparator sign consistency".
- Tree shape honesty: search for "Go unbalanced BST worst case", "Go degenerate tree insertion", "Go balanced BST out of scope".

## 17. Self-Assessment Questions

1. Why does BST correctness depend on comparator invariants such as transitivity, what does violating them do to the tree, and why does the implementation not attempt runtime validation?
2. Why must the comparator return a value whose sign (negative, zero, positive) — and not its magnitude — defines the order?
3. Why does the duplicate-ignore policy require that a comparator-equal insertion not change the tree or the size, and what does a duplicate test pin about that policy?
4. Why is the empty-tree case explicit (search returns a miss, traversal returns an empty slice, size is zero), and what does a nil-dereference panic imply about the implementation?
5. Why does the in-order traversal return a fresh slice, and what does a no-mutation test observe?
6. Why does a degenerate insertion order produce a degenerate shape, why is balancing out of scope, and why must the package documentation state that rule explicitly?
7. Why does the same generic implementation work for `int`, `string`, and a struct, and what does a multi-type test pin about the type parameter design?
8. Why does the same tree, given an ascending comparator, produce an ascending in-order traversal, and the same tree given a descending comparator produce a descending in-order traversal?
9. Why does size match the actual node count after every insertion, and what does a size-drift test catch?
10. Why must the missing-comparator policy be a single safe rule (empty tree safe, insertion rejected, `nil` comparator rejected, no panic, comparator not replaceable after first insertion), and what would a permissive "comparator is optional" or "comparator can be swapped" rule hide?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- [ ] The README's 19 sections are present in order; this file is the reference.
- [ ] Every functional requirement in section 8 is satisfied.
- [ ] Every verification case in section 14 has a corresponding test.
- [ ] The tree is generic through type parameters. The same implementation works for `int`, `string`, and a struct with a comparator that compares a chosen field.
- [ ] An empty tree without a configured comparator supports safe search (returns a miss) and in-order traversal (returns an empty slice). Insertion on a tree without a non-`nil` comparator returns a contextual error and does not mutate the tree. Configuring a `nil` comparator is rejected with a contextual error. Insertion never panics.
- [ ] Insertion places a new value at the correct child slot. A comparator-equal insertion reports "not added" and does not change the tree or the size.
- [ ] Search returns a hit on a comparator-equal value and a miss otherwise. Empty-tree search returns a miss.
- [ ] In-order traversal returns values in comparator order as a fresh slice. Empty-tree traversal returns an empty slice.
- [ ] Once any value is in the tree, the comparator cannot be replaced. An attempt to replace it returns a contextual error and does not mutate the tree.
- [ ] Ascending and descending comparators produce ascending and descending in-order traversals on the same tree.
- [ ] The package documentation states the comparator's invariants, the duplicate-ignore policy, the in-order traversal contract, the empty-tree contract, the missing-comparator policy (empty tree safe, insertion rejected, `nil` comparator rejected, comparator not replaceable after first insertion), and the no-balancing rule.
- [ ] The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Min and max.** Add a method each that returns the minimum and maximum value in the tree by the comparator's order. The method returns the not-found outcome (sentinel error or boolean) for an empty tree and does not mutate the tree. Do not add a `k`-th-smallest query or a rank operation.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisite:** [Project 027 — Custom Stack and Queue](../../02-data-structures/027_custom_stack_queue/README.md#20-prerequisite-based-documentation-guide).

Read the linked guide first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`cmp`](https://pkg.go.dev/cmp).

### Project-specific learning focus

- **Learn now:** comparator laws, insertion and search invariants, recursive in-order traversal, duplicate policy, and the worst case of an unbalanced tree.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
