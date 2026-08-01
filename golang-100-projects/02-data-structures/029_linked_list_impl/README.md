# Project 029 — Linked List Implementation

## 1. Project Name and Number

Project **029** — `029_linked_list_impl`. The directory name and number must match exactly. The project implements a singly linked generic list restricted to comparable values. The list reinforces generics while focusing on pointers and nodes. The list's zero value is usable immediately, with no constructor call required. Insertion is at a zero-based position; deletion removes the first matching value; find returns the first position; length returns the current size; values returns the values in order. Inserting at an invalid position returns an error and does not mutate the list. Deleting or finding a missing value has an explicit not-found outcome and does not mutate the list.

## 2. Project Idea

The list is a singly linked chain of nodes. Each node holds a value and a pointer to the next node. The list has a head pointer (the first node) and, if the learner chooses to track one, a tail pointer (the last node) and a length field. Operations walk the chain through the `next` pointers.

The list works only for comparable values. Comparable means the element type is one of Go's built-in or user-defined types usable with the `==` operator, and therefore satisfiable by the `comparable` constraint that the generic type parameter declares. This restriction lets the project focus on pointers and nodes without any separate comparator or comparison rule supplied by the caller. The same restriction also lets `find` and `delete` use `==` directly. Generics are still in play: the list works for many comparable types, including `int`, `string`, and structs whose fields are all comparable.

Insertion accepts a zero-based position and a value. Position `0` inserts before the current head. Position `length` inserts after the current tail. Positions in between insert at the corresponding chain position. Invalid positions (negative, greater than the current length) return an error and do not mutate the list.

Deletion accepts a value and removes the first node whose value equals the target by `==`. If no such node exists, deletion reports the not-found outcome and does not mutate the list.

Find accepts a target value and returns the position of the first node whose value equals the target by `==`. If no such node exists, find reports the not-found outcome.

Length returns the current size.

Values returns a slice of the values in chain order. The returned slice is independent: it does not expose the internal nodes, and a test that mutates the slice does not mutate the list.

## 3. Why This Project Now?

Projects 001–028 established variables, functions, loops, structs, errors, slices, files, JSON, CSV, scanning, sorting, walking, hashing, shape-validated matrices, generic zero-value containers, and a comparator-driven BST. None of them focused on pointer chains. Project 029 is the project's first encounter with a dynamic, pointer-linked data structure.

The project reinforces generics in a different setting: the element type must be comparable (not just any type), but the implementation is still generic across many comparable types. The project also forces the learner to reason about head and tail updates carefully: insertion at the head rewrites the head pointer; insertion at the tail rewrites the tail pointer; insertion in the middle rewrites the previous node's `next` pointer. Deletion can rewrite the head pointer (deleting the first node), the tail pointer (deleting the last node), or a middle node's `next` pointer. Each case must be handled explicitly.

The not-found outcome for find and delete is the same pattern established in project 027 (containers) and project 028 (BST). The list joins the project in the discipline of "empty reads and removals are explicit not-found outcomes, not panics and not fabricated values".

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 029 therefore requires:

- Completion of **028** (Binary Search Tree). Earlier projects (for example 027's container discipline, 022's interface discipline, 016's slice discipline) are background concepts already encountered and may inform the learner's design, but they are not additional required completions for this project.
- No prior knowledge of HTTP, databases, or concurrency.

## 5. What You Must Know Before Starting

- That a pointer is a value that holds the address of another value. Setting a pointer to `nil` means "no value". Following a `nil` pointer panics. The list operations check for `nil` before dereferencing a pointer.
- That a singly linked list node has a value and a `next` pointer. The `next` pointer points to the next node or is `nil` if there is no next node. Walking the chain follows `next` from the head until `next` is `nil`.
- That the list's head pointer is the entry point. An empty list has a `nil` head pointer. A non-empty list has a non-`nil` head pointer that points to the first node.
- That a tail pointer (if the learner chooses to track one) is the last node. Inserting at the tail rewrites the tail pointer. Deleting the tail rewrites the tail pointer to the previous node. An empty list has a `nil` tail pointer. A non-empty list has a non-`nil` tail pointer.
- That a length field (if the learner chooses to track one) holds the current size. The length grows by one on each successful insertion and shrinks by one on each successful deletion. A successful find or length query does not change the length.
- That type parameters with a `comparable` constraint accept types that support `==`. The constraint is enough for `int`, `string`, and structs whose fields are all comparable.
- That `errors.New` and `fmt.Errorf` produce errors the caller can match. A sentinel error (`var ErrInvalidIndex = errors.New(...)`) lets the caller use `errors.Is` to distinguish invalid-position errors from real failures.

## 6. Explanation of New Concepts

### Singly linked chain

A singly linked chain is a sequence of nodes where each node holds a value and a `next` pointer. The chain is walked by following `next` from the head. The chain ends when `next` is `nil`. There is no `prev` pointer; walking backward requires restarting from the head.

### Head and tail

The head pointer is the first node. The tail pointer (if tracked) is the last node. Insertion at position `0` creates a new node whose `next` points to the current head, then rewrites the head pointer to the new node. Insertion at the tail position rewrites the current tail's `next` pointer to a new node, then rewrites the tail pointer to the new node. Insertion in the middle walks to the node at position `pos - 1`, creates a new node whose `next` points to the node at position `pos`, and rewrites the node at position `pos - 1`'s `next` pointer to the new node. Each case rewrites a different pointer; getting the case wrong produces a broken chain.

Deletion of the first node rewrites the head pointer to the first node's `next`. Deletion of the last node walks to the second-to-last node and rewrites its `next` to `nil`, then rewrites the tail pointer to the second-to-last node. Deletion in the middle walks to the node at position `pos - 1` and rewrites its `next` to the node at position `pos + 1`.

### Comparable constraint

The element type is constrained to `comparable`. This means the element type supports `==` (or, in Go's generics vocabulary, the type is in the `comparable` constraint set). The constraint is enough for `int`, `string`, and structs whose fields are all comparable. It is not enough for slices, maps, or functions, which are not comparable. The project pins this restriction explicitly.

### Zero-based positions

Positions are zero-based. Position `0` is before the current head. Position `length` is after the current tail. Positions in between are at the corresponding chain position. The valid range is `[0, length]`. Position `length + 1` is invalid (no slot exists). Negative positions are invalid.

### Invalid position

An invalid position is one outside `[0, length]`. Inserting at an invalid position returns an error (for example a sentinel error matched with `errors.Is`) and does not mutate the list. The list's length, head, and tail are unchanged. The returned error names "invalid position" and includes the offending position and the allowed range.

### First-match deletion

Deletion removes the first node whose value equals the target by `==`. If the same value appears multiple times in the list, only the earliest occurrence is removed. The list's order is otherwise unchanged. If the value does not appear in the list, deletion reports the not-found outcome and does not mutate the list.

### First-match find

Find returns the position of the first node whose value equals the target by `==`. If the same value appears multiple times in the list, only the earliest occurrence's position is returned. If the value does not appear in the list, find reports the not-found outcome.

### Values in order

The `values` operation returns a slice of the values in chain order. The slice is independent: it does not expose the internal nodes. A test that mutates the returned slice does not mutate the list. The returned slice is a fresh allocation.

### Zero-value usability

The list's zero value is usable immediately. A freshly declared list has a `nil` head, a `nil` tail (if tracked), and length zero. Insertion at position `0` on an empty list creates the first node and sets head, tail, and length correctly.

### Head and tail updates for empty, head, middle, and end operations

For each insertion position, the operation must update the right pointer:

- **Position `0` on an empty list.** Creates the first node. Sets head to the new node, tail to the new node (if tracked), length to one.
- **Position `0` on a non-empty list.** Creates a new node whose `next` points to the current head. Sets head to the new node. Tail (if tracked) is unchanged. Length grows by one.
- **Position `length` on a non-empty list.** Walks to the current tail. Creates a new node. Sets the current tail's `next` to the new node. Sets tail to the new node (if tracked). Head is unchanged. Length grows by one.
- **Position in the middle.** Walks to the node at position `pos - 1`. Creates a new node whose `next` points to the node at position `pos`. Sets the node at position `pos - 1`'s `next` to the new node. Head and tail (if tracked) are unchanged. Length grows by one.

For each deletion case, the operation must update the right pointer:

- **Delete the first node.** Sets head to the first node's `next`. If the list becomes empty, sets tail to `nil` (if tracked). Length shrinks by one.
- **Delete the last node.** Walks to the second-to-last node. Sets the second-to-last node's `next` to `nil`. Sets tail to the second-to-last node (if tracked). Head is unchanged. Length shrinks by one.
- **Delete a middle node.** Walks to the node at position `pos - 1`. Sets the node at position `pos - 1`'s `next` to the node at position `pos + 1`. Head and tail (if tracked) are unchanged. Length shrinks by one.

### No internal node exposure

The `values` operation must not expose internal node pointers. The returned slice is a fresh allocation of values, not a slice that aliases the chain's node values. A test that mutates the returned slice does not mutate the list. A test that mutates a list value does not affect a previously returned slice.

### Multiple comparable types

The list works for many comparable types. Tests cover at least three distinct element types — for example `int`, `string`, and a small struct with two comparable fields. The same implementation works for all three types without per-type code.

## 7. Learning Objective

After completing this project the learner can:

- Implement a singly linked generic list whose element type is constrained to `comparable`. The implementation works for many comparable types without per-type code.
- Implement insertion at a zero-based position, handling the head, middle, and tail cases explicitly, and updating head, tail (if tracked), and length correctly for each case.
- Implement deletion of the first matching value by `==`, handling the head, middle, and tail cases explicitly, and updating head, tail (if tracked), and length correctly for each case.
- Implement find that returns the position of the first matching value by `==`.
- Implement length and values-in-order traversal. The values operation returns a fresh slice that does not expose internal nodes.
- Reject invalid insertion positions with an error that names the position and the allowed range. The list is not mutated.
- Return the not-found outcome for find and deletion of a missing value. The list is not mutated.
- Handle the zero value of the list: a freshly declared list has length zero, a `nil` head, and a `nil` tail (if tracked). Insertion at position `0` on an empty list creates the first node and sets head, tail, and length correctly.
- Use the `comparable` constraint in a generic type parameter so the implementation works for `int`, `string`, and comparable structs.
- Confirm that the values operation does not expose internal nodes by mutating a returned slice and re-running the operation.

## 8. Functional Requirements

1. The package defines a generic singly linked list type. The element type is constrained to `comparable`. The list's zero value is usable immediately, with no constructor call.
2. Insertion accepts a zero-based position and a value. Position `0` inserts before the current head. Position `length` inserts after the current tail. Positions in between insert at the corresponding chain position. Invalid positions (negative, greater than `length`) return a sentinel error and do not mutate the list.
3. Deletion accepts a value and removes the first node whose value equals the target by `==`. If the value is not in the list, deletion returns the not-found outcome (sentinel error matched with `errors.Is` or a clearly-named boolean) and does not mutate the list.
4. Find accepts a target value and returns the position of the first node whose value equals the target by `==`. If the value is not in the list, find returns the not-found outcome.
5. Length returns the current size. Length starts at zero (empty list). Each successful insertion increments length by one. Each successful deletion decrements length by one. A not-found find or a not-found deletion does not change length.
6. Values returns a slice of the values in chain order. The returned slice is a fresh allocation. The slice does not expose internal node pointers. Mutating the slice does not mutate the list.
7. The list updates head, tail (if tracked), and length correctly for insertion at position `0` on an empty list, insertion at position `0` on a non-empty list, insertion at the tail position, and insertion in the middle.
8. The list updates head, tail (if tracked), and length correctly for deletion of the first node, deletion of the last node, and deletion in the middle.
9. The list's element type is constrained to `comparable`. The list works for `int`, `string`, and comparable structs. The implementation does not bake in a single type.
10. The package documentation states the zero-based position rule, the valid position range `[0, length]`, the invalid-position error, the first-match deletion and find rules, the not-found outcome, and the zero-value-usable rule.
11. The list is not safe for concurrent use. Concurrency safety is out of scope.
12. Core logic is testable without terminal, real user directories, network, or any external service.

## 9. Inputs and Outputs

### Inputs

- For insertion: a zero-based position (integer) and a value (of the list's element type).
- For deletion: a target value (of the list's element type).
- For find: a target value (of the list's element type).
- For length: no value.
- For values: no value.

### Outputs

- For insertion: an error. Success returns `nil`. Invalid position returns a sentinel error that names the position and the allowed range.
- For deletion: an error or a clearly-named boolean indicating whether a node was removed. Success reports "removed". A not-found value reports "not removed" without changing the list.
- For find: an integer position and a clearly-named boolean (or an `errors.Is`-matchable error). Success returns the position of the first matching node and a "found" boolean. A not-found value returns the not-found outcome without changing the list.
- For length: the current size as an integer.
- For values: a fresh slice of values in chain order.

### Example text-only traces

List with `int` elements:

```
insert 0, 10  → ok, list: [10],             length=1
insert 1, 30  → ok, list: [10, 30],          length=2
insert 1, 20  → ok, list: [10, 20, 30],      length=3
insert 4, 99  → error: invalid position 4, allowed range [0, 3]
find 20       → position 1, found
find 99       → not found
delete 20     → removed, list: [10, 30],      length=2
delete 99     → not removed, list: [10, 30], length=2
values        → [10, 30]
```

List with duplicate values:

```
insert 0, 1   → ok, list: [1],               length=1
insert 1, 1   → ok, list: [1, 1],             length=2
insert 2, 1   → ok, list: [1, 1, 1],          length=3
find 1        → position 0, found            (first match)
delete 1      → removed, list: [1, 1],        length=2 (first match removed)
values        → [1, 1]
```

## 10. Rules and Edge Cases

- **Zero value usability.** A freshly declared list is empty: head is `nil`, tail is `nil` (if tracked), length is zero. Insertion at position `0` on an empty list creates the first node and sets head, tail, and length correctly.
- **Position range.** Valid positions are `[0, length]`. Position `0` is before the current head. Position `length` is after the current tail. Negative positions and positions greater than `length` are invalid.
- **Invalid position.** Insertion at an invalid position returns a sentinel error (matched with `errors.Is`). The list is not mutated. Length, head, and tail are unchanged.
- **Position `0` on an empty list.** Creates the first node. Head, tail, and length are set correctly.
- **Position `0` on a non-empty list.** Inserts before the current head. Head is rewritten. Tail (if tracked) is unchanged. Length grows by one.
- **Position `length` (tail).** Inserts after the current tail. Tail is rewritten. Head is unchanged. Length grows by one.
- **Position in the middle.** Inserts at the corresponding chain position. Head and tail are unchanged. Length grows by one.
- **Delete the first node.** Head is rewritten to the first node's `next`. Tail (if tracked) is unchanged unless the list becomes empty. Length shrinks by one.
- **Delete the last node.** Walks to the second-to-last node. The second-to-last node's `next` is rewritten to `nil`. Tail (if tracked) is rewritten to the second-to-last node. Head is unchanged. Length shrinks by one.
- **Delete in the middle.** Walks to the node at position `pos - 1`. The node at position `pos - 1`'s `next` is rewritten to the node at position `pos + 1`. Head and tail are unchanged. Length shrinks by one.
- **Delete the only node.** Head is rewritten to `nil`. Tail (if tracked) is rewritten to `nil`. Length shrinks to zero.
- **First-match deletion.** If a value appears multiple times, only the earliest occurrence is removed.
- **First-match find.** If a value appears multiple times, the position of the earliest occurrence is returned.
- **Not-found find.** Returns the not-found outcome. Length, head, and tail are unchanged.
- **Not-found delete.** Returns the not-found outcome. Length, head, and tail are unchanged.
- **Multiple element types.** The same generic implementation works for `int`, `string`, and a small struct with comparable fields. Behavior is identical across types.
- **Values in order.** The `values` operation returns a fresh slice of values in chain order. The slice does not expose internal node pointers.
- **No internal exposure.** Mutating a returned slice does not mutate the list. Mutating a list value does not affect a previously returned slice.
- **No panic.** The list operations never panic on invalid position, not-found value, or empty list. All outcomes are explicit returns.
- **No concurrency.** Concurrent access is undefined behavior. Concurrency safety is out of scope.

## 11. Project Constraints

- Go standard library only. No third-party list libraries, no `container/list`.
- The list is generic through type parameters with a `comparable` constraint. The implementation works for many comparable types without per-type code.
- The list is singly linked. There is no `prev` pointer.
- The list's zero value is usable immediately, with no constructor call.
- Invalid insertion positions return a sentinel error and do not mutate the list.
- Not-found find and not-found deletion return an explicit not-found outcome and do not mutate the list.
- The values operation returns a fresh slice that does not expose internal nodes.
- Concurrency safety is out of scope.
- Core logic is testable without terminal, real user directories, network, or any external service.
- Element types are restricted to `comparable`. Non-comparable types (slices, maps, functions) are not accepted by the implementation.

## 12. Design Questions Before Coding

- How is the list represented? As a struct with a head pointer (and optionally a tail pointer and a length field), or as a head pointer alone? Which choice makes the tail-update cases straightforward?
- How is length tracked? As a stored field on the list, as a recursive walk on demand, or both? A stored field is the conventional choice and makes length a constant-time accessor.
- How is the zero-value-usable property enforced? Are all internal fields valid as their zero values, or is there a one-time initialization the runtime can perform lazily?
- How is the invalid-position error shaped? Plain `error`, wrapped with `fmt.Errorf` and `%w`, or a typed error with named fields (offending position, allowed range)? Which choice keeps the error readable in tests and in production logs?
- How is the not-found outcome exposed? As a sentinel error matched with `errors.Is`, as a boolean paired with a position (or `nil` pointer), or both? Which choice is most discoverable for a caller reading the package?
- How is insertion at position `0` on an empty list distinguished from insertion at position `0` on a non-empty list? Are they handled as the same case or as two separate cases? Which choice keeps the head-update logic correct?
- How is deletion of the only node handled? As a separate case, or as a fall-through of the head-update and tail-update logic?
- How is the values operation implemented? As a fresh allocation walked from the head, or as a slice that aliases the chain's node values? Which choice satisfies the no-internal-exposure contract?
- How is the package documentation structured? Does it state the position rule, the invalid-position error, the first-match rules, the not-found outcome, the zero-value-usable rule, and the concurrency-unsafety rule in the package comment or in the type's doc comment?

## 13. Implementation Milestones

1. Decide the package layout. Keep the list as a generic struct with a head pointer, an optional tail pointer, and a length field. Keep `main` as a thin driver that inserts, deletes, finds, and prints values.
2. Pin the public contract as named constants or sentinel errors: the position rule, the valid position range `[0, length]`, the invalid-position error, the first-match rules, the not-found outcome, and the zero-value-usable rule.
3. Implement the node type. A node holds a value and a `next` pointer. Both fields are initialized to zero values.
4. Implement length. The length field starts at zero. Each successful insertion increments length. Each successful deletion decrements length. Not-found find or not-found deletion does not change length.
5. Implement insertion. Validate the position. On invalid position, return the sentinel error and do not mutate the list. On valid position, walk to the insertion point, create a new node, update head or tail (or a middle node's `next`) as appropriate, and increment length.
6. Implement deletion. Walk the chain to find the first matching node. If found, update head or tail (or a middle node's `next`) as appropriate, and decrement length. If not found, return the not-found outcome and do not mutate the list.
7. Implement find. Walk the chain to find the first matching node. If found, return the position and a "found" boolean. If not found, return the not-found outcome.
8. Implement values. Walk the chain from head to tail, copying each value into a fresh slice. Return the slice.
9. Verify the zero-value-usable contract. A test declares a list with the zero value, exercises empty-state operations, and confirms the not-found outcome.
10. Verify the head and tail updates. Tests cover insertion at position `0` on an empty list, insertion at position `0` on a non-empty list, insertion at the tail, insertion in the middle, deletion of the first node, deletion of the last node, deletion in the middle, and deletion of the only node.
11. Wire `main`. The driver exercises a small script of insertions, deletions, finds, and values. The driver is not part of the package's public contract.
12. Add tests for every verification case in section 14, with empty-list tests, position-validation tests, head/middle/end tests, deletion tests, find tests, length tests, multi-type tests, and no-internal-exposure tests separated.

## 14. Verification Cases the Learner Must Write

Each case is described in natural language. Tests use only in-memory elements and direct API calls; no terminal, no real user directories.

### Empty list

- A freshly declared list has length zero.
- Find on an empty list returns the not-found outcome.
- Delete on an empty list returns the not-found outcome.
- Values on an empty list returns an empty slice.
- Insertion at position `0` on an empty list creates the first node, sets head, sets tail (if tracked), and sets length to one.

### Insert head

- Insertion at position `0` on a non-empty list prepends the new value. The new value is at position `0` in the subsequent values traversal. Head is rewritten to the new node. Tail (if tracked) is unchanged unless the list was empty. Length grows by one.
- Insertion at position `0` when the list has one node creates a list of two nodes with the new value first.

### Insert middle

- Insertion at a position strictly between `0` and `length` places the new value at that position. The values before and after the insertion point are unchanged. Length grows by one.
- Insertion at position `1` on a list of length two produces a list of length three with the new value in the middle.

### Insert end (tail)

- Insertion at position `length` appends the new value. The new value is at the end of the subsequent values traversal. Tail (if tracked) is rewritten to the new node. Length grows by one.
- Insertion at position `length` on a list of one node produces a list of two nodes with the new value second.

### Invalid index

- Insertion at a negative position returns a sentinel error. The list is not mutated.
- Insertion at a position greater than `length` returns a sentinel error. The list is not mutated.
- Insertion at position `length + 1` returns a sentinel error. The list is not mutated.

### Duplicate values

- Inserting the same value multiple times produces a list with one node per insertion (duplicates are allowed in the list).
- Find on a list with duplicates returns the position of the first occurrence.
- Delete on a list with duplicates removes the first occurrence only. The remaining occurrences are still present in the same relative order.
- After deleting the first occurrence, a subsequent find for the same value returns the position of the new first occurrence.

### Delete only / first / middle / last

- Deletion of the only node (list of length one) returns the not-found outcome if the value does not match, or returns "removed" and produces an empty list if the value matches.
- Deletion of the first node (list of length two or more) rewrites head to the next node. Tail (if tracked) is unchanged unless the list becomes empty.
- Deletion of a middle node rewrites the previous node's `next` to the node after the deleted one. Head and tail are unchanged.
- Deletion of the last node rewrites the second-to-last node's `next` to `nil`. Tail (if tracked) is rewritten to the second-to-last node.
- After deletion, values returns the remaining values in chain order.

### Missing value

- Delete with a value not in the list returns the not-found outcome. The list is not mutated. Length, head, and tail are unchanged.
- Find with a value not in the list returns the not-found outcome. The list is not mutated.

### Length invariants

- Length starts at zero.
- Each successful insertion grows length by exactly one.
- Each successful deletion shrinks length by exactly one.
- An invalid-position insertion does not change length.
- A not-found deletion does not change length.
- A not-found find does not change length.
- Length matches the actual node count after every operation (a test can verify by counting nodes through a recursive walk or by relying on values' length as a proxy, depending on the package's surface).

### Zero-value use

- A freshly declared list is empty.
- Insertion at position `0` on the empty list creates the first node and sets head, tail, and length correctly.
- Subsequent operations on the same list work as on a list constructed explicitly.

### Multiple comparable types

- The same generic implementation works for `int`, `string`, and a small struct with two comparable fields.
- For each type, insertion, deletion, find, length, and values behave identically.
- For struct element types, equality by `==` is used (the constraint allows this for comparable structs).

### Stable order

- After a sequence of insertions, the values traversal returns values in the order they were inserted.
- After deleting the first occurrence of a value, the values traversal returns the remaining values in their original relative order.

### No internal exposure

- Mutating a slice returned by `values` does not change the list. A subsequent values traversal returns the original values.
- Mutating a list value (for example by inserting a new value) does not affect a previously returned slice.
- The test confirms that the returned slice is a fresh allocation, not a slice that aliases the chain's node values.

### Process

- A test runs the driver with a small script of insertions, deletions, and finds and confirms the printed values traversal matches the expected text-only form.

## 15. Common Mistakes to Watch For

- **Baking in a single element type.** `type List struct { head *NodeInt }` instead of a generic type is wrong. The list must work for many comparable types.
- **Forgetting the `comparable` constraint.** A type parameter without a constraint cannot use `==` on its values. Without the constraint, find and delete cannot compare values and the implementation does not compile.
- **Constructing the list through a `New` call only.** A zero-value-usable list does not require construction. A test that declares a list with the zero value and uses it immediately must succeed.
- **Updating head on insertion at position `0` but forgetting tail when the list becomes non-empty.** When the first node is created, both head and tail (if tracked) must be set.
- **Forgetting to update tail when inserting at the tail position.** Insertion at position `length` rewrites the current tail's `next` to the new node and rewrites the tail pointer to the new node. Forgetting either step leaves the list broken.
- **Forgetting to update the previous node's `next` on a middle insertion.** Middle insertion walks to position `pos - 1` and rewrites that node's `next`. Forgetting the rewrite leaves the new node unreachable from head.
- **Updating head on deletion of the first node but not resetting tail when the list becomes empty.** When the only node is deleted, both head and tail (if tracked) must be `nil`.
- **Leaving a deleted node's `next` pointing at the rest of the chain.** The deleted node itself is unreachable from the list's head, but it still holds a `next` pointer that reaches the rest of the chain. As long as any reference to that unreachable node exists, the rest of the chain is reachable through it. The implementation may overwrite `next` on deletion as a hygiene step; the project's tests do not require pinning this since Go's garbage collector handles unreachable chains, but it is worth knowing.
- **Returning a slice that aliases the chain's node values.** The `values` operation must return a fresh allocation. Returning a slice whose elements are the chain's node values directly (rather than copies) is wrong because a test that mutates the slice would mutate the list.
- **Returning a not-found outcome that does not distinguish from "found with a zero position".** A boolean paired with a position must be "found" only when the value was actually found. A "found" boolean with a position of `0` on an empty list is wrong; the boolean must be "not found" with the not-found position (or the sentinel error must be returned).
- **Panicking on invalid position.** The contract is an error return, not a panic.
- **Losing chain order across deletions.** After deleting a middle node, the previous node's `next` must point to the node after the deleted one. A test that walks the chain after deletion pins the order.
- **Allowing the list to become inconsistent (head, tail, length out of sync).** Each operation must keep head, tail (if tracked), and length consistent with the actual chain state.
- **Returning a non-comparable element type.** A test that tries to use a slice element type fails because slices are not comparable. The constraint pins the restriction.

## 16. Topics and References for Study

- A Tour of Go: "Generics", "Pointers", "Errors".
- Effective Go: "Generics", "Data", "Errors".
- Package documentation: `errors` (New, Is), `fmt` (Errorf, %w), `cmp` (Compare).
- Linked list patterns: search for "Go singly linked list generic", "Go comparable constraint list", "Go pointer chain head tail".
- Pointer reasoning: search for "Go pointer next nil walk", "Go singly linked insertion deletion", "Go linked list head tail update".
- Generics with constraints: search for "Go comparable constraint", "Go generics comparable type parameter", "Go generics struct comparable field".

## 17. Self-Assessment Questions

1. Why is the list's element type explicitly constrained to `comparable`, what does that enable, and what would an unconstrained parameter fail to guarantee at compile time?
2. Why is the list's zero value usable without a constructor, and what does this property buy the caller?
3. Why must insertion at position `0` on an empty list and insertion at position `0` on a non-empty list update different pointers, and what does each case rewrite?
4. Why do middle and tail insertions require different rewrites of the previous node's `next`, the new node's `next`, and the tail pointer, and what would forgetting one leave in the chain?
5. Why must deletion of the only node reset both head and tail to `nil`, and what would leaving tail pointing to a detached (stale, unreachable) node leave?
6. Why does the first-match rule mean duplicates are removed one at a time, and what does a duplicate-list test observe about the order of removal?
7. Why does the `values` operation return a fresh slice, and what does a no-internal-exposure test observe about mutation independence?
8. Why does the not-found outcome (sentinel error or boolean) distinguish "I read an empty position" from "I read a real value", and what would a single return value lose?
9. Why does invalid position return a sentinel error and not a panic, and what does the error's naming of the offending position and allowed range buy?
10. Why is concurrency safety out of scope for this project, and what would change about the implementation and the API if concurrency were required?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every functional requirement in section 8 is satisfied.
- Every verification case in section 14 has a corresponding test.
- The list is a generic type with a `comparable` constraint. The same implementation works for `int`, `string`, and a small struct with comparable fields.
- Insertion at position `0`, position `length`, and positions in between all work and update head, tail (if tracked), and length correctly.
- Invalid positions return a sentinel error matched with `errors.Is`. The list is not mutated.
- Deletion removes the first matching node and updates head, tail (if tracked), and length correctly. A missing value returns the not-found outcome and does not mutate the list.
- Find returns the position of the first matching node. A missing value returns the not-found outcome.
- Length matches the actual node count after every operation.
- The list's zero value is usable immediately, with no constructor call. Insertion at position `0` on the zero-value list creates the first node and sets head, tail, and length correctly.
- The values operation returns a fresh slice in chain order. The slice does not expose internal node pointers. Mutating the slice does not mutate the list.
- The package documentation states the position rule, the invalid-position error, the first-match rules, the not-found outcome, the zero-value-usable rule, and the concurrency-unsafety rule.
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Reverse values.** Add a method that returns the values in reverse chain order as a fresh slice. The method walks the chain once and appends to a fresh slice from the back. The method does not mutate the list and does not expose internal node pointers. Do not add a `Reverse()` in-place mutation or a doubly linked list.
- **Contains.** Add a method that returns a boolean indicating whether a target value appears in the list by `==`. The method does not mutate the list and does not change length. The method uses the same first-match rule as `find`. Do not add a `count` operation or a `find-all-positions` operation.
