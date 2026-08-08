# Project 092 — CLI Docker Manager

## 1. Project Name and Number

- Project 092, `092_cli_docker_manager`.
- The folder name is fixed by the curriculum table; do not rename the directory.

## 2. Project Idea

A small, safety-limited command-line tool that lists, starts, and stops containers on the local Docker daemon. The tool talks to Docker through a small client interface that the rest of the program depends on. The tool refuses to create, remove, prune, pull images, exec into containers, relabel containers, or to launch any subprocess or shell command. Every operation is gated on the exact label `go-tutorial.project=092`, so the tool only ever touches containers that the learner has marked out of band. Stop requires an interactive confirmation read from an injected reader, or an explicit `--yes` flag; cancellation or a "no" answer leaves the container untouched.

## 3. Why This Project Now?

- This project is the CLI dependency-inversion capstone.
- It pulls the discipline of "small interface, real implementation, fake implementation for tests" into a tool whose target happens to be a real, dangerous integration surface.
- The previous project pulled together middleware composition and a JSON error contract; this project pulls together injection, argument resolution, and a strict refusal surface.
- The formal prerequisites — project 011 (interactive menu CLI) and project 041 (context cancellation and timeout) — each contribute one ingredient that this tool combines.
- The immediate catalog predecessor, project 091, is optional context rather than a formal prerequisite.

## 4. Prerequisites

- The formal prerequisites are projects 011 and 041; project 091 is the immediate catalog predecessor and remains useful as optional context rather than a formal prerequisite.

## 5. What You Must Know Before Starting

- That the Docker Go client is large and the right move is to define a tiny interface that exposes only `List`, `Start`, and `Stop`, then to depend only on that interface everywhere else.
- That "local" Docker in this project means the local Unix socket on Linux and macOS, or the named pipe on Windows; the tool refuses any other connection, including remote TCP daemons.
- That container lifecycle states exposed by the daemon include created, running, paused, restarting, exited, and dead. The tool's view is restricted to "running" and "stopped" for start and stop.
- That an interactive confirmation prompt must read from an injected reader so tests can simulate "yes", "no", and end-of-file without a TTY.
- That context cancellation must propagate into the Docker call and out of the program with a clean exit code and a stable error message.
- That the tool logs names and identifiers but never logs image labels, environment variables, or other potentially sensitive fields.
- That the tool never shells out to a `docker` binary; the entire point of this project is to depend on the Docker API directly.

## 6. Explanation of New Concepts

### Concepts

- Dependency inversion over the Docker API: the program defines an interface with exactly three methods, `List`, `Start`, and `Stop`, and depends only on that interface. The real Docker client and the fake test client both implement that interface. No other Docker type appears in the rest of the program.
- Allowed set as a filter, not a security boundary: the tool requires the exact label `go-tutorial.project=092` on every container it operates on. The label is a safety filter against accidental operations on the learner's other containers, not a security mechanism.
- Exact identifier matching within the allowed set: the resolver matches an argument within the labeled set by either an exact normalized name match or an exact full identifier match. Fuzzy matching is not allowed and silent fuzzy matching is forbidden. The resolver reports a distinct outcome for unmanaged, missing, ambiguous, and invalid input. An ambiguous outcome occurs when a single input matches more than one allowed record; the command returns the ambiguous outcome and makes zero lifecycle calls.
- Idempotent start: starting an already-running container is not an error. The command reports the running status and exits successfully without calling `Start`.
- Stop with explicit confirmation: stopping a container is destructive enough to warrant a confirmation. The tool reads one line from the injected reader, trims surrounding ASCII whitespace, and compares case-insensitively. Only `y` and `yes` proceed. `n`, `no`, an empty line, end-of-file, or any unrecognized answer aborts without calling `Stop`.
- Stop outcome on cancellation or timeout: when the stop API call has been issued and the context is canceled or the timeout elapses, the outcome is unknown with respect to the daemon. The tool never claims a rollback occurred; it reports only what the daemon returned.
- Interface narrowness: the interface has exactly three methods. Adding `Remove`, `Prune`, `Pull`, `Exec`, `Kill`, `Pause`, `Unpause`, `Restart`, `Inspect`, `Logs`, or a relabel method is an unauthorized scope expansion and is not present in this project.
- Integration safety model: the optional integration test uses a pre-existing disposable local container that is explicitly named by an environment variable. The container must already carry the exact pinned learning label. The test refuses to run otherwise. The test may list, start, and stop that container and must restore its original running or stopped state on exit. The test never creates, removes, or relabels a container and never invokes a shell command.

## 7. Learning Objective

- After completing this project you must be able to explain in your own words: why the interface is exactly three methods, why the label is mandatory, why start is idempotent and stop is not, why the confirmation comes from an injected reader, why the tool refuses any non-local daemon, why the implementation never opens a subprocess, why ambiguous identifiers return without lifecycle calls, why cancellation after the API call left the daemon is reported as unknown rather than as rollback, and how the integration test stays safe.

## 8. Functional Requirements

1. The tool exposes three subcommands: `list`, `start <id|name>`, and `stop <id|name>`.
2. The tool connects only to `unix:///var/run/docker.sock` on Unix-like systems or `npipe:////./pipe/docker_engine` on Windows. Any other endpoint, including every TCP endpoint, is refused before constructing the real client.
3. The tool defines a small interface for the three operations and depends only on that interface throughout the program. The real Docker client and the fake test client both implement this interface.
4. The tool requires the exact label `go-tutorial.project=092`. It filters every operation through that key-value pair. A container without that exact label is invisible to `list` and cannot be started or stopped.
5. `list` prints, in deterministic order, every container that carries the label, with its identifier, name, and running status. Order is name ascending as primary and identifier ascending as tiebreaker.
6. `start <id|name>` resolves the argument within the allowed set. The resolver accepts an exact normalized name match or an exact full identifier match. If the input matches more than one allowed record, the command returns an ambiguous outcome and makes zero lifecycle calls. If the input is unmanaged, missing, or syntactically invalid, the command returns the corresponding distinct outcome and makes zero lifecycle calls. If the container is already running, the command reports the running status and exits successfully without calling `Start`. If the container is stopped, the command calls `Start` with a context that has a bounded timeout. Cancellation returns the cancellation outcome.
7. `stop <id|name>` resolves the argument within the allowed set using the same rules. The command reads one confirmation line from an injected reader. After trimming surrounding ASCII whitespace and folding case, only `y` and `yes` proceed; `n`, `no`, empty input, end-of-file, and every unrecognized answer abort without calling `Stop`. If `--yes` is passed, the confirmation is skipped and the command proceeds. Stop calls `Stop` with a bounded timeout. Cancellation or timeout after the API call has been issued returns the cancellation or timeout outcome with the daemon state reported as unknown; the tool does not claim rollback.
8. The tool logs each operation with its arguments and outcome. Logs never include environment variables, image labels, inspect output, or other fields outside the narrow container summary.
9. Errors from the Docker daemon are mapped to a small set of stable user-readable messages. The daemon error detail is logged but not printed verbatim.
10. The tool exits zero on success, non-zero with one class on user error (unmanaged, missing, ambiguous, refused confirmation), and non-zero with another class on daemon error.
11. The implementation does not create, remove, prune, pull images, exec into containers, relabel containers, or launch any subprocess or shell command.

## 9. Inputs and Outputs

### Interface Contract

- `list` output: one line per labeled container, in deterministic order, formatted as `identifier  name  status`. Example: `a1b2c3d4  web-1  running`.
- `start <id|name>` output on success: a single line `started <identifier> <name>`. On already-running without a `Start` call: a single line `already-running <identifier> <name>`.
- `start <id|name>` output on unmanaged: `unmanaged <input>`.
- `start <id|name>` output on ambiguous: `ambiguous <input> <count>` listing the candidates.
- `start <id|name>` output on invalid input: `invalid <input>`.
- `stop <id|name>` confirmed: `stopped <identifier> <name>`. Refused: `aborted <identifier> <name>`. Timeout or cancellation: `unknown <identifier> <name>` with the cause class.

## 10. Rules and Edge Cases

- A container without the label does not appear in `list` and cannot be referenced by `start` or `stop`. The argument resolver runs against the filtered set.
- An argument that matches multiple containers within the allowed set is ambiguous; the command returns the ambiguous outcome and makes zero lifecycle calls.
- An argument that does not resolve within the allowed set is unmanaged; the command returns the unmanaged outcome and makes zero lifecycle calls.
- An argument that fails the exact normalized name or exact full identifier rule is invalid; the command returns the invalid outcome and makes zero lifecycle calls.
- `stop` with end-of-file on the confirmation reader aborts without a `Stop` call.
- `stop` with an unrecognized answer aborts without a `Stop` call.
- `stop --yes` skips the prompt and calls `Stop`.
- `start` and `stop` honor context cancellation. Cancellation before the API call returns the cancellation outcome with zero lifecycle calls. Cancellation after the API call has been issued returns the cancellation outcome with the daemon state reported as unknown; the tool does not claim rollback.
- The optional integration test refuses to run unless the environment variable identifies a pre-existing local container that already carries the pinned label. On exit, the integration test restores the original running or stopped state of that container.
- The integration test never creates, removes, or relabels a container and never invokes a shell command.

## 11. Project Constraints

- The Docker Go client version is selected by the learner at implementation time from the currently supported official client versions, and is pinned in the module. No invented version is committed to the documentation.
- The interface has exactly three methods: `List`, `Start`, `Stop`. Tests must assert that no other method exists on the interface.
- The only accepted connection targets are `unix:///var/run/docker.sock` and `npipe:////./pipe/docker_engine`, selected for the host platform. Every other endpoint is rejected.
- No configuration file format. The only configuration is the pinned label, the accepted affirmative token, and the timeout; constants or flags only.
- Unit tests must run locally with no Docker daemon. They use a fake implementation that records call counts and returns programmable errors.
- The optional integration test is gated by a build tag and an explicit environment opt-in. Both gates are required.
- Logs do not include image labels, environment variables, or any field the user did not place in source.
- No subprocess or shell path exists in the implementation; tests assert behaviorally that no subprocess of the binary is created.

## 12. Design Questions Before Coding

- What is your interface exactly, and which fields does each method return?
- How do you express exact normalized name and exact full identifier matching without falling back to fuzzy matching?
- How do you inject the reader and the writer so the prompt works in tests?
- How will you implement the pinned `y` and `yes` confirmation policy, including whitespace, case, end-of-file, and unrecognized input?
- How do you express the timeout, and what is the default and the override flag?
- How do you map unmanaged, missing, ambiguous, invalid input, and daemon error to distinct exit codes?
- How does the integration test identify its disposable local container, and how does it restore the original running or stopped state on exit?
- How does the test suite assert that no subprocess was created during any command?

## 13. Implementation Milestones

1. Define the three-method interface and a small container value type carrying identifier, name, and running status.
2. Implement the fake implementation that records call counts and can return programmable errors.
3. Implement the argument resolver with exact normalized name match and exact full identifier match within the labeled set, with distinct outcomes for unmanaged, missing, ambiguous, and invalid input.
4. Implement the `list` subcommand with deterministic ordering and label filtering.
5. Implement `start <id|name>` with distinct outcomes and idempotent already-running handling.
6. Implement the prompt helper that reads from the injected reader and treats end-of-file and unrecognized answers as "no".
7. Implement `stop <id|name>` and `stop --yes <id|name>`, propagating context and reporting an unknown outcome if cancellation or timeout arrives after the API call was issued.
8. Implement the real Docker client wrapping the official Go client. The client refuses any non-local endpoint at construction. It exposes only the three interface methods.
9. Implement the CLI entry point that selects the fake or real client, dispatches subcommands, and maps outcomes to exit codes.
10. Implement logging with the narrow log policy.
11. Write the unit test suite covering every command, every argument class, every confirmation outcome, the cancellation outcome, and the no-destructive-method invariant.
12. Write the opt-in integration test gated by build tag and environment opt-in that uses a pre-existing labeled local container, exercises the three commands, and restores the original state on exit without creating, removing, or relabeling any container.

## 14. Verification Cases the Learner Must Write

### Required Cases

- `list`: deterministic ordering across calls; a container without the label does not appear; an empty allowed set produces an empty list and exit code zero.
- `start` resolution: exact name match starts the container; exact identifier match starts the container; ambiguous within the allowed set returns ambiguous with zero lifecycle calls; unmanaged returns unmanaged with zero lifecycle calls; invalid input returns invalid with zero lifecycle calls.
- `start` idempotency: an already-running container does not call `Start` and reports `already-running`.
- `stop` confirmation: accepted affirmative calls `Stop` once and reports `stopped`; accepted negative does not call `Stop` and reports `aborted`; end-of-file does not call `Stop` and reports `aborted`; unrecognized answer does not call `Stop` and reports `aborted`.
- `stop --yes`: prompt is skipped and `Stop` is called once.
- Daemon error: a fake returning a daemon error yields the daemon error exit code and a stable user-facing message; the daemon error detail is logged only.
- Cancellation: cancellation before the API call returns cancellation with zero lifecycle calls; cancellation after the API call has been issued returns the cancellation outcome with the daemon state reported as unknown and never claims rollback.
- Call counts: across all subcommands the fake records exactly the expected call sequence and no others; tests assert that `Remove`, `Prune`, `Pull`, `Exec`, `Kill`, and a relabel method are not on the interface.
- Log content: captured log lines never contain image labels, environment variables, or any field other than the container identifier, name, status, and command outcome.
- No subprocess: a behavioral test asserts that no subprocess of the binary is created during any command.
- Integration opt-in: with the gating flag and the gating build tag set, a pre-existing labeled local container is listed, started, stopped, and its original running or stopped state is restored on exit; the test never creates, removes, or relabels a container and never invokes a shell command.

## 15. Common Mistakes to Watch For

- Adding a method to the interface because "it would be useful"; every added method widens the blast radius and is an unauthorized scope expansion.
- Treating "already running" as an error.
- Reading confirmation from `os.Stdin` directly.
- Treating any non-affirmative answer as anything other than "no" for safety.
- Connecting to a remote Docker daemon because an environment variable pointed there.
- Logging inspect output or other sensitive fields.
- Letting `Stop` block forever when the daemon is unhealthy.
- Running the integration test by default in CI.
- Letting the integration test create, remove, or relabel a container; the test must use a pre-existing labeled container and must restore its state.
- Letting the implementation open a subprocess to a `docker` binary.
- Failing to distinguish unmanaged, missing, ambiguous, and invalid input as separate outcomes.
- Falling back to fuzzy matching when an exact match would resolve.

## 16. Topics and References for Study

- The official Docker Go client documentation and the `ContainerList`, `ContainerStart`, and `ContainerStop` methods on `Client`.
- The Docker engine API documentation for the corresponding REST endpoints.
- The Go blog post on interfaces and dependency inversion, with the standard library's narrow interfaces (`io.Reader` and friends) as a model.
- The Go `testing` package and the `flag` package for CLI parsing and exit code control.
- The Go `context` package for cancellation and timeout.
- The project documentation for the Docker Go client version you pin.

## 17. Self-Assessment Questions

1. Why is the interface exactly three methods?
2. Why is the label mandatory, and what would go wrong if the label were optional?
3. Why is `start` idempotent and `stop` not idempotent?
4. Why must the confirmation come from an injected reader?
5. Why must the tool refuse any non-local daemon?
6. Why must the implementation never open a subprocess?
7. How does the resolver ensure exact matching and avoid fuzzy fallback?
8. How does the tool distinguish unmanaged, missing, ambiguous, and invalid input?
9. How does the tool report an unknown outcome when cancellation or timeout arrives after the API call was issued?
10. What guarantees does the integration test rely on, and what does it forbid?

## 18. Definition of Completion

- [ ] The project is complete when the tool exposes exactly three subcommands, depends only on its own three-method interface, refuses any non-local endpoint, refuses to operate on a container without the pinned learning label, runs `list` deterministically, starts an already-running container without a `Start` call and a stopped container via the interface, prompts for confirmation on `stop` unless `--yes` is passed, returns distinct outcomes for unmanaged, missing, ambiguous, invalid, and daemon error inputs, reports an unknown outcome rather than a rollback when cancellation or timeout arrives after the API call was issued, never opens a subprocess, and passes its unit tests with no Docker daemon running and no interface methods beyond `List`, `Start`, and `Stop`.
- [ ] The opt-in integration test passes only when the gating build tag is selected, the gating environment variable names a pre-existing labeled local container, every command against that container succeeds, the original running or stopped state is restored on exit, and the test never creates, removes, or relabels any container or invokes a shell command.

## 19. Optional Extensions

- A `status <id|name>` subcommand that returns the current running or stopped status without starting or stopping, subject to the same resolver rules. The interface gains a fourth method only for this extension, and the unit tests must then assert the new method count.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 011 — Interactive Menu](../../01-foundations/011_interactive_menu/README.md#20-prerequisite-based-documentation-guide), [Project 041 — Context Timeout Example](../../03-concurrency/041_context_timeout_example/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`github.com/docker/docker/client`](https://pkg.go.dev/github.com/docker/docker/client).
- **Standards and concept references:** [Docker Engine API](https://docs.docker.com/reference/api/engine/), [Docker Go SDK guide](https://docs.docker.com/reference/api/engine/sdk/).

### Project-specific learning focus

- **Learn now:** narrow client interfaces, labeled resource scoping, exact argument resolution, context deadlines, exit-code contracts, test fakes, dry runs, and avoiding destructive Docker operations.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
