# Project 027 — Custom Stack and Queue

## 1. Project Name and Number

- Project **027** — `027_custom_stack_queue`.
- The directory name and number must match exactly.
- The project implements generic stack and FIFO queue containers whose zero values are usable immediately, with no constructor call required.
- The containers work for any element type the caller supplies.
- Empty reads and removals return an explicit not-found outcome and never panic or invent a value.
- The stack is LIFO; the queue is FIFO.
- Concurrency safety is out of scope.

## 2. Project Idea

Two generic containers share a design philosophy: their zero values are ready to use, every empty-state operation is a clearly-named not-found outcome, and the containers preserve order across interleaved operations. The stack supports push, pop, peek, length, and empty-state behavior. The queue supports enqueue, dequeue, front (peek), length, and empty-state behavior. Each container works for arbitrary element types through type parameters.

The stack uses a slice as backing storage. The queue also uses a slice as backing storage, but the design acknowledges a queue-specific concern: removed elements should not be retained indefinitely, and the underlying slice's capacity should not grow without bound as elements are enqueued and dequeued. The project discusses this concern and the trade-offs of different release strategies, without prescribing one exact implementation. The discussion is conceptual; the implementation choice is the learner's.

The project is concurrency-unsafe by design. No locks, no channels, no atomics. Concurrent access is undefined behavior. The required scope does not test concurrency.

## 3. Why This Project Now?

- Projects 001–026 established variables, functions, loops, structs, errors, slices, files, JSON, CSV, scanning, sorting, walking, hashing, and shape-validated matrices.
- None of them used type parameters.
- Project 027 is the project's first encounter with generics.
- The learner must reason about a single implementation that works across many element types and must design a zero-value-usable API rather than relying on a constructor to set up internal state.

- The stack and queue are also the first containers the learner writes.
- They have to reason about what an empty read or removal should mean, what a peek should mean, how length transitions are observable, and what slice backing storage actually does when elements are added and removed.
- The queue's release-of-removed-elements discussion is the project's first explicit encounter with the gap between "I can read this element" and "I should retain this element".

- Project 027 is referenced by project 031 (concurrent timer uses channels but the container discipline is background) and by later projects whose data structures depend on understanding how a slice grows, shrinks, and holds references.
- The container discipline carries forward into project 028 (binary search tree) and project 029 (linked list), both of which build on the same "what does an empty read mean" pattern.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 027 therefore requires:

- Completion of **026** (Matrix Operations). Earlier projects (for example 022's interface discipline, 016's slice discipline, 014's validation discipline) are background concepts already encountered and may inform the learner's design, but they are not additional required completions for this project.
- No prior knowledge of generics, HTTP, databases, or concurrency.

## 5. What You Must Know Before Starting

- That a type parameter is a placeholder for an element type chosen at the call site. The package declares the type parameter once, and the caller picks the concrete type when constructing (or, with a zero value, never constructing) the container.
- That the zero value of a struct is usable without an explicit constructor, provided the fields are valid for immediate use. A slice field's zero value is `nil`, which has length `0` and can be appended to.
- That `append` may reallocate the underlying array when capacity is exceeded, and that the slice header still points to the same logical position after reallocation.
- That removing an element from a slice by re-slicing (for example shrinking the length after a stack pop, or advancing the head after a queue dequeue) does not by itself clear the underlying slot in the backing array. The slot still holds the removed element's value until something overwrites it. The slot is reachable through the backing array for as long as the backing array itself remains reachable. A container that relies on the runtime's garbage collector to clean up reachable slots has not released the reference.
- That "empty read" and "empty removal" are distinct states. A peek on an empty container is a read of "nothing", a pop on an empty container is a removal of "nothing". Both must be representable without panic.
- That `errors.New` and `fmt.Errorf` produce errors the caller can match. A sentinel error (`var ErrEmpty = errors.New(...)`) lets the caller use `errors.Is` to distinguish empty-state outcomes from real failures.
- That the standard library's `container/list` exists but is doubly-linked and not the focus of this project. The project implements its own slice-backed containers.

## 6. Explanation of New Concepts

### Concepts

#### Zero-value usability

The containers' zero values must work. A caller can declare a variable, use it immediately, and observe empty-state behavior on first use. No constructor, no `New()` call, no initialization step is required. The cost is that the internal fields must be valid as their zero values: a slice field that is `nil` works because `append` handles `nil` slices, and a length field that is `0` works because the container starts empty.

#### Not-found outcomes for empty operations

Reads and removals on an empty container return an explicit not-found outcome. The outcome is a clearly-named value the caller can check (for example an `ok` boolean, a sentinel error, or an `Optional`-style pair). The project does not panic, does not return a zero-valued element by pretending one was found, and does not invent a default value the caller did not ask for. A test that calls pop on an empty stack observes a not-found result, not a panic and not a returned zero value.

#### LIFO stack

The stack is last-in, first-out. The most recently pushed element is the next popped or peeked element. Order across interleaved operations is preserved: if the caller pushes `a`, pushes `b`, pops one element, and then pushes `c`, the next pop returns `c`, the next pop returns `a`, and the stack is empty after that. The test pins this order directly.

#### FIFO queue

The queue is first-in, first-out. The earliest enqueued element is the next dequeued or fronted element. Order across interleaved operations is preserved: if the caller enqueues `a`, enqueues `b`, dequeues one element, and then enqueues `c`, the next dequeue returns `b`, the next dequeue returns `c`, and the queue is empty after that. The test pins this order directly.

#### Backing storage and reference release

A slice is a header that points into a backing array. Two operations on the same slice can share the same backing array even when the slice header's length changes. Re-slicing only changes the header's length (and possibly its start offset); it does not touch the backing array's slots that are no longer covered by the header. As long as the backing array is reachable, the slots it covers still hold the values that were last written there. A reference in those slots is retained until the slot is overwritten or the backing array becomes unreachable.

For the queue, this is a hard requirement, not an optional refinement. A long-running queue that enqueues and dequeues millions of elements must not hold the dequeued elements through its backing array indefinitely once the queue's logical length has dropped. The queue's reclamation policy is part of the contract. The learner chooses one bounded-reclamation strategy and implements it:

- **Compact-on-dequeue.** After each dequeue, or after a threshold, copy the live elements forward into a fresh slice. The dequeued elements are no longer reachable through the fresh slice's backing array.
- **Slot-zero on release.** When advancing the head, explicitly zero the released slot before adjusting the header, so the backing array no longer retains a reference at that index.
- **Ring buffer with bounded indices.** Keep a head index, a tail index, and a fixed-size or growth-bounded backing array. The retained prefix is bounded by the array's capacity, not by the number of enqueues performed.

The project pins the observable outcome: after a queue has logically consumed some prefix, the consumed elements are not retained through the container's backing storage. The same rule applies to the stack: the slot vacated by a pop must not retain the popped element's reference after the pop returns.

The project does not allow a "we'll leave it for the GC" or "we accept the trade-off" answer. The release strategy is explicit, documented in the package comment, and tested by the same-package tests that inspect the container's chosen representation. The implementation choice is open; the outcome is pinned.

The implementation does not rely on the runtime's garbage collector, on finalizers, or on weak references to detect that a slot has been released. The release is explicit and structural. Tests verify it by inspecting the container's chosen representation directly (for example the slice header and the slots within its range, or the head/tail indices and the array's reachable range), not by waiting for the GC to run.

#### Length transitions

Length is observable. After a push, length grows by one. After a pop that succeeds, length shrinks by one. After a pop on an empty container, length stays at zero. The empty-state report (`empty` or `length == 0`) is consistent with the not-found outcome of a read or removal. Tests pin the transitions directly.

#### Element type diversity

The containers work for many element types. Tests cover at least three distinct element types — for example `int`, `string`, and a small struct — to demonstrate that the implementation does not bake in a single type. Tests also cover pointer and interface element types where practical, observing that the container holds the values the caller enqueued, not copies or zero values.

#### Concurrency is out of scope

The containers are not safe for concurrent use. Concurrent push, pop, enqueue, dequeue, peek, or length reads are undefined behavior. The required scope does not test concurrency, and the project does not claim to be safe. A package comment states the concurrency-unsafety rule so a future reader does not assume otherwise.

## 7. Learning Objective

After completing this project the learner can:

- Declare a generic type with a type parameter and design its zero value to be immediately usable, with no constructor call required.
- Design an empty-state outcome for reads and removals that is explicit and not-found, distinct from a real value or a panic.
- Implement a LIFO stack with push, pop, peek, length, and empty-state behavior. Order across interleaved operations is preserved.
- Implement a FIFO queue with enqueue, dequeue, front, length, and empty-state behavior. Order across interleaved operations is preserved.
- Reason about slice backing storage: that re-slicing only changes the header; the backing array still holds the released values until overwritten or made unreachable. The same applies to a popped stack slot at the tail — the slot is still reachable through the backing array until the implementation clears it or replaces the backing array.
- Choose and pin a bounded-reclamation strategy for the queue (compact-on-dequeue, slot-zero on release, or ring buffer with bounded indices) and for the stack (slot-zero on pop or backing-array replacement). The strategy must be explicit, documented, and tested. The implementation does not rely on the runtime's garbage collector to release removed references.
- Use a sentinel error or a clearly-named boolean to signal empty-state outcomes so tests can assert the not-found contract directly.
- Use type parameters so the containers work for arbitrary element types, and confirm with tests across multiple distinct types.
- Confirm that the zero value of the container is usable without initialization, by exercising the empty-state contract on a freshly declared variable.
- Document the concurrency-unsafety rule in the package comment so the contract is honest.

## 8. Functional Requirements

1. The package defines a generic stack type and a generic queue type. Each container's zero value is usable immediately, with no `New` call and no constructor step.
2. The stack exposes push (add to top), pop (remove and return top), peek (read top without removing), length (current size), and an empty-state report. The package documents the order and the empty-state outcome.
3. The queue exposes enqueue (add to back), dequeue (remove and return front), front (read front without removing), length (current size), and an empty-state report. The package documents the order and the empty-state outcome.
4. The stack is LIFO: the most recently pushed element is the next peeked or popped element.
5. The queue is FIFO: the earliest enqueued element is the next fronted or dequeued element.
6. Reads and removals on an empty container return an explicit not-found outcome. The outcome is a sentinel error (`errors.Is`-matchable) or a clearly-named boolean paired with a zero-valued result. The package does not panic, does not return a fabricated zero value, and does not silently treat an empty container as having a default element.
7. The container's length is observable through a length accessor or equivalent. Length transitions are consistent with push/pop and enqueue/dequeue: each successful push or enqueue increments by one, each successful pop or dequeue decrements by one, and each empty-state operation leaves length at zero.
8. The container does not retain references to removed elements after a successful pop or dequeue. The chosen bounded-reclamation strategy (compact-on-dequeue, slot-zero on release, or ring buffer with bounded indices for the queue; slot-zero on pop or backing-array replacement for the stack) is implemented and verified by the same-package structural tests. The implementation does not rely on the runtime's garbage collector, on finalizers, or on weak references to release removed elements.
9. The container works for any element type supplied at the call site. Tests exercise at least three distinct element types.
10. The container is not safe for concurrent use. The package documentation states this rule.
11. The stack's pop returns the most recently pushed element, the queue's dequeue returns the earliest enqueued element. Order across interleaved push/enqueue and pop/dequeue operations is preserved.
12. The package documentation states the order rules (LIFO, FIFO), the empty-state outcome (sentinel error or boolean), the chosen bounded-reclamation strategy for both stack and queue, and the concurrency-unsafety rule.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- Elements supplied by the caller through push (stack) or enqueue (queue). Elements may be of any type the call site chooses, including `int`, `string`, structs, pointers, and interfaces.
- For length and empty-state queries, no element is supplied.
- For peek and front, no element is supplied; the operation only reads.
- For pop and dequeue, no element is supplied by the caller; the operation reads and removes.

#### Outputs

- For push and enqueue, no result value is returned beyond the implicit container state. Length increases by one.
- For pop and dequeue, the removed element and an outcome indicator. On success the indicator is "found" and the element is the most recently pushed (stack) or earliest enqueued (queue) value. On empty state the indicator is "not-found" and no real element is returned.
- For peek and front, the top/front element and an outcome indicator. Same not-found contract as pop/dequeue.
- For length, the current size. For empty-state, a true/false value indicating whether the container is empty.

#### Example text-only traces

Stack:

```
push a → [a]            length=1
push b → [a, b]         length=2
peek   → b, found       length=2 (unchanged)
pop    → b, found       length=1
push c → [a, c]         length=2
pop    → c, found       length=1
pop    → a, found       length=0
pop    → _, not-found   length=0
```

Queue:

```
enqueue a → [a]           length=1
enqueue b → [a, b]        length=2
front     → a, found      length=2 (unchanged)
dequeue   → a, found      length=1
enqueue c → [b, c]        length=2
dequeue   → b, found      length=1
dequeue   → c, found      length=0
dequeue   → _, not-found  length=0
```

#### Example text-only zero-value use

Stack (int element type):

```
declared with the zero value, no constructor
peek    → _, not-found            length=0
pop     → _, not-found            length=0
push 42                          length=1
peek    → 42, found               length=1
```

## 10. Rules and Edge Cases

- **Zero value.** A freshly declared stack or queue is empty, has length zero, and reports empty-state on peek, pop, dequeue, and front. No constructor, no `New`, no initialization step.
- **Empty pop.** Returns the not-found outcome. Length stays at zero. No panic, no fabricated value.
- **Empty dequeue.** Same as empty pop.
- **Empty peek / empty front.** Returns the not-found outcome. Length is unchanged.
- **Single element.** Pop returns the only element, the stack is empty afterward. Dequeue returns the only element, the queue is empty afterward.
- **Many elements.** Push and pop preserve LIFO order across many operations. Enqueue and dequeue preserve FIFO order across many operations.
- **Interleaved operations.** A sequence of pushes and pops (or enqueues and dequeues) preserves order. The test pins the order with a concrete script.
- **Length transitions.** Each successful push or enqueue increments length by one. Each successful pop or dequeue decrements length by one. Empty-state operations leave length at zero.
- **Empty-state report.** The container exposes an empty-state query (for example `IsEmpty` or a `len` method or equivalent). The report is consistent with the not-found outcome.
- **Multiple element types.** The same generic implementation works for `int`, `string`, and a small struct. The behavior is identical across types. No type-specific code paths.
- **Reference release.** Both the stack and the queue document a bounded-reclamation policy and implement it. After a successful pop or dequeue, the removed element's reference is not retained through the container's backing storage. The queue does not retain an unbounded consumed prefix. The stack does not retain a reference to a popped element through the vacated slot.
- **No concurrency.** Concurrent access is undefined behavior. The package documentation states this rule. The required scope does not test concurrency.
- **No panic.** Reads, removals, and length queries never panic, regardless of empty state.

## 11. Project Constraints

- Go standard library only. No third-party container libraries, no `container/list`, no `container/heap`. The package implements its own containers.
- The containers are generic through type parameters. The implementation works for any element type without per-type code.
- The containers are zero-value-usable. No constructor, no `New`, no setup step.
- Empty reads and removals return a not-found outcome through a sentinel error or a clearly-named boolean. The not-found outcome is observable in tests.
- The containers do not panic on empty state.
- The containers do not retain references to removed elements after a successful pop or dequeue. The stack and queue each document a bounded-reclamation policy and implement it. The queue does not retain an unbounded consumed prefix. The stack does not retain a reference to a popped element through the vacated slot.
- Concurrency safety is out of scope. The package documentation states this rule.
- Core logic is testable without terminal, real user directories, network, or any external service. Tests use only in-memory elements and direct API calls.
- The implementation does not rely on the runtime's garbage collector, on finalizers, or on weak references to release removed elements. The release is explicit, structural, and tested by same-package tests that inspect the container's chosen representation.

## 12. Design Questions Before Coding

- How is the not-found outcome exposed? As a sentinel error matched with `errors.Is`, as a boolean paired with a zero-valued result, or as a typed result the caller unpacks? Which choice is most discoverable for a caller reading the package for the first time?
- How is the zero-value-usable property enforced? Are all internal fields valid as their zero values, or is there a one-time initialization the runtime can perform lazily?
- How is length tracked? As a stored field, as `len(slice)`, or as both? Which choice makes the length transition tests straightforward and consistent?
- How is the queue's backing storage managed to honor the bounded-reclamation policy without unbounded retention of consumed elements? Compact-on-dequeue, slot-zero on release, or ring buffer with bounded indices? Which choice keeps the reclamation outcome testable through same-package structural inspection?
- How is the stack's backing storage managed so the vacated tail slot does not retain a reference to the popped element? Slot-zero on pop, backing-array replacement, or another strategy that the structural tests can verify?
- How is the order preserved across interleaved operations? Is the order a property of the slice itself, or is it maintained by an explicit head/tail index?
- How is the package documentation structured? Does it state LIFO and FIFO explicitly, the empty-state outcome, the chosen release strategy, and the concurrency-unsafety rule in the package comment or in the type's doc comment?

## 13. Implementation Milestones

1. Decide the package layout. Keep the stack and queue as separate generic types in the same package. Keep `main` as a thin driver that exercises a small script for each container.
2. Pin the public contract as named constants or sentinel errors: the LIFO and FIFO orders, the empty-state outcome, the length accessor, the empty-state query, and the chosen bounded-reclamation policy for both the stack and the queue.
3. Implement the stack type with a slice field. Push appends. Pop removes and returns the top element and clears the vacated tail slot so the backing array does not retain a reference. Peek returns the top element without removing. Length returns the current size. Empty-state query returns true when the container is empty. Empty reads and removals return the not-found outcome.
4. Implement the queue type with a slice field or a head/tail index. Enqueue appends. Dequeue removes and returns the front element and ensures the consumed prefix is not retained through the backing storage beyond the chosen bounded-reclamation threshold. Front returns the front element without removing. Length returns the current size. Empty-state query returns true when the container is empty. Empty reads and removals return the not-found outcome.
5. Implement the chosen bounded-reclamation strategy for both stack and queue. The strategies are documented in the package comment. The implementations match the documented strategies and are verifiable by same-package structural tests.
6. Verify zero-value usability. A test declares a stack and a queue with the zero value, exercises empty-state operations, and confirms the not-found outcome.
7. Verify order preservation across interleaved operations. A test runs a script of pushes and pops (or enqueues and dequeues) and asserts the exact order of returned elements.
8. Verify multiple element types. The same container code works for at least three distinct types. The tests exercise each type with a small script.
9. Wire `main`. The driver exercises a small script for each container and prints the traces. The driver is not part of the package's public contract.
10. Add tests for every verification case in section 14, with empty-state tests, ordering tests, length-transition tests, multiple-type tests, and bounded-reclamation tests separated.

## 14. Verification Cases the Learner Must Write

### Required Cases

Each case is described in natural language. Tests use only in-memory elements and direct API calls; no terminal, no real user directories.

#### Zero value usability

- A freshly declared stack with the zero value has length zero.
- A freshly declared stack with the zero value reports empty-state.
- A freshly declared stack with the zero value returns the not-found outcome on peek, pop, and equivalent read operations.
- A freshly declared queue with the zero value has length zero and returns the not-found outcome on front, dequeue, and equivalent read operations.
- After the first push (stack) or enqueue (queue), the container has length one and is no longer empty.

#### Empty-state operations

- Pop on an empty stack returns the not-found outcome. Length is unchanged.
- Peek on an empty stack returns the not-found outcome. Length is unchanged.
- Dequeue on an empty queue returns the not-found outcome. Length is unchanged.
- Front on an empty queue returns the not-found outcome. Length is unchanged.
- Repeated empty-state operations on an empty container all return the not-found outcome. The container does not enter a corrupt state.
- No empty-state operation panics.

#### Single and many elements

- Push a single element, then pop it. The popped element equals the pushed element. The stack is empty afterward.
- Enqueue a single element, then dequeue it. The dequeued element equals the enqueued element. The queue is empty afterward.
- Push many elements, then pop them all. The order is the reverse of the push order.
- Enqueue many elements, then dequeue them all. The order is the same as the enqueue order.

#### Interleaved operations

- Push `a`, push `b`, pop one (returns `b`), push `c`. Pop returns `c`. Pop returns `a`. Stack is empty.
- Enqueue `a`, enqueue `b`, dequeue one (returns `a`), enqueue `c`. Dequeue returns `b`. Dequeue returns `c`. Queue is empty.
- A test runs a longer interleaved script and asserts the exact order of returned elements.

#### Length transitions

- A test pushes N elements one by one and asserts length grows by one each time.
- A test pops N elements one by one from a stack of size N and asserts length shrinks by one each time.
- A test enqueues N elements one by one and asserts length grows by one each time.
- A test dequeues N elements one by one from a queue of size N and asserts length shrinks by one each time.
- After all elements are removed, length is zero and empty-state is true.
- An empty-state operation on an empty container leaves length at zero.

#### Multiple element types

- The same stack code works for `int`, `string`, and a small struct (for example a struct with two fields).
- The same queue code works for `int`, `string`, and a small struct.
- For each type, push and pop (or enqueue and dequeue) return values that compare equal to the originals.
- For struct element types, push and pop preserve field values.

#### Reference release (required, structural)

- The package documentation states the queue's bounded-reclamation policy and the stack's vacated-slot policy.
- A same-package structural test for the queue confirms that after a sequence of enqueues and dequeues, the consumed prefix is not retained through the container's backing storage beyond the documented threshold. The test inspects the chosen representation directly (the slice header and the slots within its range, or the head/tail indices and the array's reachable range). The test does not rely on the GC to clean up.
- A same-package structural test for the stack confirms that after a pop, the vacated tail slot does not retain a reference to the popped element. The test inspects the slot directly. The test does not rely on the GC to clean up.
- Tests use reference-typed elements (for example pointers, slices, or structs containing pointers) so the test can observe whether a reference is still reachable through the backing storage.
- Tests do not use finalizers, weak references, runtime read barriers, or any GC-coupled mechanism. The release is explicit and verified directly.

#### Multiple element types (reaffirmation)

- The same generic implementation works for `int`, `string`, and a small struct (for example a struct with two fields).
- For struct element types that contain reference fields, the bounded-reclamation policy still applies. The test confirms that the backing storage does not retain references through struct fields after a pop or dequeue.

#### Not-found outcome

- The not-found outcome is exposed through a sentinel error matched with `errors.Is` or through a boolean paired with a zero-valued result. The test confirms the chosen mechanism.
- The test confirms the not-found outcome is distinguishable from a real element of the same zero value (for example `0` for `int` or `""` for `string`) — that is, a caller can tell "I read an empty container" apart from "I read a real zero-valued element".

#### Concurrency-unsafety declaration

- The package documentation states that the container is not safe for concurrent use. The required scope does not test concurrency. The test does not run a concurrency check.

#### Process

- A test runs the driver with a small script for each container and confirms the printed output matches the expected text-only trace.

## 15. Common Mistakes to Watch For

- **Panicking on empty state.** The contract is a not-found outcome. `panic("empty stack")` or runtime index-out-of-range on an empty slice are wrong.
- **Returning a zero value as if it were a found value.** A peek that returns `0` for `int` on an empty container without distinguishing "found" from "not-found" is wrong. The not-found outcome must be explicit.
- **Baking in a single element type.** `type Stack struct { items []int }` instead of a type parameter is wrong. The container must work for arbitrary element types.
- **Constructing a separate version for each type.** Code duplication that exists only because type parameters were avoided is wrong. The implementation must be generic.
- **Treating the zero value as needing initialization.** A constructor that is the only correct way to use the container violates the zero-value-usable contract.
- **Releasing the queue's front by re-slicing without addressing reference retention.** Re-slicing the head leaves the released elements reachable through the backing array. The implementation must document and honor a bounded-reclamation policy.
- **Treating "the GC will handle it" or "we accept the trade-off" as the release strategy.** A comment that says the GC will clean up reachable slots is not a strategy. The implementation must explicitly release the references, the strategy must be documented, and the outcome must be verified by a same-package structural test.
- **Treating the stack's popped tail slot as automatically released.** The vacated slot still holds the popped element's reference until the implementation clears it or replaces the backing array. The structural test must verify it.
- **Losing order across interleaved operations.** A stack that returns elements in push order (instead of reverse) or a queue that returns elements in enqueue-reverse order (instead of enqueue order) is wrong. Order is part of the contract.
- **Drifting length.** Length that disagrees with the number of elements actually present (after push/pop or enqueue/dequeue) is wrong.
- **Adding locks, channels, or atomics "for safety".** Concurrency safety is out of scope. A mutex or atomic in the implementation is wrong for the required scope and changes the API surface.
- **Using `container/list`.** The project implements its own slice-backed containers. `container/list` is doubly-linked and is not the focus.
- **Returning the wrong element on peek or front.** Peek returns the top; front returns the front. Swapping them is a contract violation.
- **Treating an empty-state operation as changing state.** Pop on empty does not change length. Dequeue on empty does not change length. The not-found outcome leaves the container unchanged.

## 16. Topics and References for Study

- A Tour of Go: "Generics", "Slices", "Errors".
- Effective Go: "Generics", "Data", "Errors".
- Package documentation: `errors` (New, Is), `fmt` (Errorf, %w), `slices` (where helpful).
- Type parameter design patterns: search for "Go generics zero value usable", "Go generics container", "Go type parameter struct".
- Slice backing array and reference holding: search for "Go slice backing array retained reference", "Go slice header independence", "Go slice zero element release".
- Queue release strategies: search for "Go slice ring buffer", "Go queue head tail index", "Go queue compaction", "Go slice clear element release".
- Stack slot-zero and backing-array replacement patterns: search for "Go slice clear element", "Go slice nil out slot release".
- Container design philosophy: search for "Go zero value usable container", "Go optional result tuple", "Go sentinel error empty".
- Same-package structural testing of internal state: search for "Go test internal package", "Go same-package test access unexported", "Go slice backing array test".

## 17. Self-Assessment Questions

1. Why is the zero value of the stack and queue usable without a constructor, and what does this property buy the caller?
2. Why must an empty-container outcome be an explicit sentinel error or boolean distinguishable from a real zero-valued element, rather than a panic or fabricated value, and what does that distinction buy the caller?
3. Why must the stack be LIFO and the queue be FIFO, and what does an interleaved-operations test prove that a single-operation test does not?
4. Why does re-slicing the head of a queue leave the dequeued element reachable through the backing array, and what does this mean for the queue's bounded-reclamation policy?
5. Why does popping the tail of a stack leave the popped element reachable through the vacated slot in the backing array, and what does the stack's vacated-slot policy require the implementation to do?
6. Why is concurrency safety out of scope and documented explicitly, what would an implicit assumption hide, and what would change about the implementation and API if concurrency were required?
7. Why does the same generic implementation work for many element types, and what does a multi-type test prove about the type parameter design?
8. Why must length be consistent with the number of elements actually present after every push/pop or enqueue/dequeue, and what does a length-drift test catch?
9. Why is each container's bounded-reclamation policy part of the contract, and why must a same-package structural test inspect the chosen representation directly instead of relying on garbage collection, finalizers, or weak references?
10. Why is `container/list` not the right backing for this project, and what would using it change about the slice-based reasoning the project requires?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- [ ] The README's 19 sections are present in order; this file is the reference.
- [ ] Every functional requirement in section 8 is satisfied.
- [ ] Every verification case in section 14 has a corresponding test.
- [ ] The stack and queue are generic type-parameterized types. The same implementation works for `int`, `string`, and a small struct without per-type code.
- [ ] The zero value of each container is usable immediately, with no constructor call.
- [ ] Empty reads and removals return an explicit not-found outcome through a sentinel error or a clearly-named boolean. The outcome is `errors.Is`-matchable or boolean-true on empty state.
- [ ] LIFO and FIFO orders are preserved across interleaved operations. A test pins the order with a concrete script.
- [ ] Length is consistent with the number of elements actually present after every operation.
- [ ] Each container documents a bounded-reclamation policy in the package comment. A same-package structural test verifies the policy by inspecting the container's chosen representation directly. The implementation does not rely on the runtime's garbage collector, finalizers, or weak references.
- [ ] The queue does not retain an unbounded consumed prefix. The test confirms this with reference-typed elements and a structural inspection of the backing storage's reachable range.
- [ ] The stack does not retain a reference to a popped element through the vacated tail slot. The test confirms this by inspecting the slot directly.
- [ ] No empty-state operation panics.
- [ ] The package documentation states LIFO, FIFO, the empty-state outcome, each container's bounded-reclamation policy, and the concurrency-unsafety rule.
- [ ] The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Capacity hint.** Add an optional pre-allocation constructor that accepts an initial capacity hint and pre-allocates the backing slice to that capacity so the caller avoids early reallocations. The zero-value-usable contract is preserved: callers who do not use the hint still get a working container. Do not add a capacity query or a reallocate-on-demand method.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisite:** [Project 026 — Matrix Operations](../../02-data-structures/026_matrix_operations/README.md#20-prerequisite-based-documentation-guide).

Read the linked guide first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`slices`](https://pkg.go.dev/slices).
- **Standards and concept references:** [Go generics tutorial](https://go.dev/doc/tutorial/generics), [Go specification: type constraints](https://go.dev/ref/spec#Type_constraints).

### Project-specific learning focus

- **Learn now:** zero-value-friendly generic containers, LIFO and FIFO contracts, backing-array retention, slot clearing, queue compaction, and amortized cost.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
