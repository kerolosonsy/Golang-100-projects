# 100 Go Projects: Learn by Building

A documentation-first challenge for learning Go by building 100 progressively harder projects yourself.

This repository does **not** contain finished solutions, starter implementations, or generated project code. Each project has a detailed learning guide; you are responsible for designing the solution, writing every line of Go, creating the tests, and verifying the result.

## What Is Included?

- 100 project guides numbered `001` through `100`.
- Seven levels covering language foundations, data structures, concurrency, APIs, databases, networking, and advanced systems.
- A consistent 19-section README for every project.
- Project-specific requirements, edge cases, design questions, milestones, verification cases, study references, and completion criteria.
- An explicit prerequisite map and level gates in the curriculum plan.

The guides are ready, but the projects are intentionally not implemented.

## Curriculum

| Level | Projects | Focus |
|---|---:|---|
| 1 | 001-015 | Go and CLI foundations |
| 2 | 016-030 | Data, files, and algorithms |
| 3 | 031-045 | Concurrency and cancellation |
| 4 | 046-060 | HTTP and REST services |
| 5 | 061-070 | Databases and caching |
| 6 | 071-085 | Networking and protocols |
| 7 | 086-100 | Advanced systems and the capstone |

Read the complete [curriculum plan](golang-100-projects/plan.md) for the project order, prerequisite rules, level gates, and learner workflow.

## How to Start

1. Install a supported Go release and become comfortable using a terminal.
2. Read the [curriculum plan](golang-100-projects/plan.md) completely.
3. Open [Project 001: Hello CLI](golang-100-projects/01-foundations/001_hello_cli/README.md).
4. Confirm that you understand the prerequisite checklist and study the listed concepts.
5. Answer the design questions before coding.
6. Create the implementation and tests yourself.
7. Complete every required verification case and the project's Definition of Completion.
8. Continue in order; do not skip formal prerequisites or level gates.

## Project Status

- `📘` — the learning guide is ready.
- `🟦` — you are currently implementing and testing the project.
- `✅` — you implemented, tested, and understood the project.

All 100 guides currently use `📘`. Guide readiness is not implementation progress.

## Learning Rules

- Write every line of implementation and every test yourself.
- Use the README as a specification and learning map, not as a solution.
- Prefer official documentation and conceptual explanations when studying.
- Test success, failure, boundary, security, and concurrency behavior where applicable.
- Mark a project complete only when you can explain the design and satisfy its completion criteria.
- Keep optional extensions separate from the required core project.

## AI Assistance Policy

AI tools are welcome only as read-only mentors and reviewers. They may explain concepts, ask guiding questions, clarify requirements, and evaluate work you have already written. They must not change repository files or provide code, tests, snippets, patches, pseudocode, commands, or other copyable solutions.

Every supported coding agent is directed to the canonical [Mandatory AI Mentor-Only Policy](AI_POLICY.md). Tool-specific instruction files and read-only settings reinforce the same boundary.

## Repository Structure

```text
.
├── README.md
├── LICENSE
└── golang-100-projects/
    ├── plan.md
    ├── 01-foundations/
    ├── 02-data-structures/
    ├── 03-concurrency/
    ├── 04-apis-and-services/
    ├── 05-databases/
    ├── 06-networking/
    └── 07-advanced-systems/
```

Each numbered project folder contains its own `README.md`. Learner-created implementation files belong inside the relevant project folder and must not replace the guide.

## Contributing

Improvements to project explanations, correctness, sequencing, or verification coverage are welcome. Do not add finished solutions or starter implementations to the curriculum guides.

## License

Licensed under the [MIT License](LICENSE).
