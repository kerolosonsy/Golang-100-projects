# Project 096 — Metrics Prometheus Exporter

## 1. Project Name and Number
Project 096, `096_metrics_prometheus_exporter`. Build an HTTP metrics exporter that exposes a health endpoint and a Prometheus scrape endpoint through an injected Prometheus registry rather than the global default registry, and instrument a small deterministic Work service with a fixed set of pinned metric descriptors. This README is a learning guide only. It contains no implementation code, signatures, starter snippets, solution snippets, pseudocode, or implementation shell commands. Text-only input and output examples are permitted.

## 2. Project Idea
A caller submits a small Work item whose mode is one of a fixed enumeration. The Work service validates the submission, increments a request counter exactly once, decides an outcome, increments exactly one outcome counter, observes exactly one duration histogram sample, and keeps the in-progress gauge equal to the number of valid submissions currently executing. After all work completes, the gauge is zero. The HTTP layer exposes exactly two endpoints: a health endpoint at `/healthz` and a metrics endpoint at `/metrics`. The metrics endpoint uses the Prometheus text exposition handler bound to the injected registry. The service is constructed against an injected registry; no reference to the global default registerer or gatherer exists. The exporter is a learning surface for counters, gauges, histograms, label cardinality, registry isolation, and scrape content; it is not a production telemetry system.

## 3. Why This Project Now?
Projects 046 and 060 are the formal prerequisites: Project 046 contributes HTTP basics, and Project 060 contributes graceful-shutdown discipline. Project 095 is optional immediate-catalog-predecessor context, while Projects 064 and 086 are optional prior review for migration and deterministic-test discipline.

## 4. Prerequisites
Projects 046 and 060 are the formal prerequisites. Project 046 provides HTTP basics; Project 060 provides graceful-shutdown discipline. Project 095 is optional immediate-catalog-predecessor context. Optional prior review includes Project 064 for migration discipline and Project 086 for fake-clock and deterministic-test patterns. Be comfortable with `context`, the difference between counters, gauges, and histograms, the difference between an injected registry and the global default registry, the difference between validation rejection and execution rejection, and the boundary between a learning surface and a production monitoring system.

## 5. What You Must Know Before Starting
- A counter is a monotonically increasing value that resets to zero only on process restart. A counter never decreases during normal operation.
- A gauge is a value that may go up or down and represents a current observation. In-progress work is a gauge.
- A histogram observes a distribution of samples with a configured bucket layout; one histogram call observes exactly one sample.
- Labels are key-value pairs that multiply the metric series. Every unique label combination is one time series. Unbounded label values create scrape cost and storage cost.
- A Prometheus registry is the bookkeeping object that holds registered collectors. The global default registry is shared by every Prometheus call in a process; an injected registry is owned by the object that constructed it. The exporter here uses an injected registry only.
- The Prometheus text exposition format is the wire format a scrape returns. It is generated from the contents of a registry.
- The HTTP method discipline for both endpoints is pinned. Other methods on either path return `405`; unknown paths return `404`. Neither endpoint creates metrics dynamically.
- A panic in a worker must be recovered at the service boundary so a panic can never leave the in-progress gauge in a non-zero state.
- The deterministic observation seam is the only path through which a duration sample reaches the histogram in tests. Unit tests never depend on wall-clock time and never sleep.

## 6. Explanation of New Concepts
The exporter is constructed against an injected registry through the Prometheus Go client's registry constructor. The constructor is the only place where the four pinned descriptors are registered. Re-registering any of the pinned descriptors on the same registry returns a clean typed duplicate-registration outcome rather than panicking; two distinct registries may each hold the same descriptors and never leak values between them.

The health endpoint at `/healthz` responds to `GET` with status `200`, content type `application/json`, and the exact compact body `{"status":"ok"}` with no trailing newline. Other methods on `/healthz` return `405`. The metrics endpoint at `/metrics` responds to `GET` with the text exposition format gathered from the injected registry. Other methods on `/metrics` return `405`. Unknown paths return `404`. Neither endpoint creates metrics dynamically.

The four pinned descriptors are:
- `tutorial_work_requests_total`, help text `Total number of work submissions received.`, no labels.
- `tutorial_work_outcomes_total`, help text `Total number of work submissions by outcome.`, one label named `outcome`.
- `tutorial_work_in_progress`, help text `Number of work submissions currently executing.`, no labels.
- `tutorial_work_duration_seconds`, help text `Duration of work submissions in seconds by outcome.`, one label named `outcome`, buckets `0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5`.

The exact set of `outcome` label values is `success`, `validation_rejected`, `transient_failure`, and `canceled`. No other label values exist for either the outcome counter or the duration histogram. The constructor initializes all four labeled counter and histogram children so a scrape before the first submission exposes each outcome with a zero count. The Work input field that drives behavior, called mode here, is never a label and is never recorded in any metric.

The lifecycle is coherent. Every submission attempt increments `tutorial_work_requests_total` exactly once, increments exactly one of the `tutorial_work_outcomes_total` series exactly once, and observes exactly one sample on `tutorial_work_duration_seconds`. Validation runs before execution. A submission whose mode is not in the fixed mode set is rejected as `validation_rejected` and the in-progress gauge is never incremented for that submission. A valid submission increments the gauge immediately before execution, runs the deterministic state machine, and decrements the gauge exactly once on every outcome: `success`, `transient_failure`, and `canceled`. Under concurrency the gauge equals the number of valid submissions still executing; it reaches zero after all of them finish.

The mode set that drives behavior is exactly `success`, `transient_failure`, and `block_until_canceled`. Any other mode value is `validation_rejected`. The label set on the outcome metric and the duration histogram never contains a mode value.

A worker panic is recovered at the service boundary. The recovered submission is classified as `transient_failure` without leaking panic detail into any error or log path; exactly one duration sample is observed; the gauge is decremented; the service returns an error. After recovery, the gauge is zero. There is no execution path that leaves the gauge non-zero.

The deterministic observation seam accepts the start time and the end time of a submission and emits the observed duration. The seam is the only path through which a duration sample reaches the histogram in tests. The seam is injected through the Work service constructor. A unit test that exercises duration assertions reads the observed sample from the registry, not from wall-clock time.

Counters never decrease. The request counter and the outcome counter never drop during operation. The completion path asserts counter monotonicity explicitly.

The injected observation seam, the injected clock, and the injected registry are the three seams that make the unit test deterministic. The unit test owns all three. The test constructs the service, drives every outcome path, gathers from the injected registry, and asserts the exact label set and bucket distribution.

## 7. Learning Objective
After completing this project you must be able to explain in your own words: why every metric descriptor is pinned in code rather than constructed at runtime; why the `outcome` label set is a fixed enumeration and never a free-form string; why mode is a Work input field and never a metric label; why validation runs before execution and why a validation rejection never increments the in-progress gauge; why a worker panic must be recovered at the service boundary so the gauge returns to zero on every outcome; why a duplicate registration against an injected registry must return a clean typed outcome rather than panic; why an injected registry is used instead of the global default registry; why counters never decrease during operation; how a scrape reflects the registered descriptors and not the contents of the global default registry; why the `/healthz` body is the exact compact form with no trailing newline; why the HTTP method discipline returns `405` on other methods and `404` on unknown paths; why the duration histogram uses an injected observation seam in deterministic tests; why this learning project is not a production monitoring system; and what production monitoring concerns such as retention, alert evaluation, federation, push gateway, service-level objectives, and host-level metric scrapes are deliberately outside scope.

## 8. Functional Requirements
1. The exporter exposes exactly two endpoints: `GET /healthz` and `GET /metrics`. Other methods on either path return `405`. Unknown paths return `404`. Neither endpoint creates metrics dynamically.
2. The exporter is constructed against an injected registry through the Prometheus Go client's registry constructor. No reference to the global default registerer or global default gatherer exists in the exporter code paths.
3. The four metric descriptors are registered with the exact names, exact help text, exact label sets, and exact histogram buckets pinned in this guide. No other descriptors exist.
4. The exact `outcome` label values are `success`, `validation_rejected`, `transient_failure`, and `canceled`. No other label values exist on the outcome counter or the duration histogram. The constructor preinitializes all four values for both labeled collectors so each appears with zero count before the first submission.
5. The Work service accepts a submission whose mode is one of the fixed set `success`, `transient_failure`, and `block_until_canceled`. Any other mode is rejected as `validation_rejected`.
6. Every submission attempt, including `validation_rejected`, increments `tutorial_work_requests_total` exactly once.
7. Every submission attempt increments exactly one series of `tutorial_work_outcomes_total` and observes exactly one sample on `tutorial_work_duration_seconds`.
8. Validation runs before execution. A `validation_rejected` submission never increments `tutorial_work_in_progress`, and the gauge remains zero across the validation rejection path.
9. A valid submission increments `tutorial_work_in_progress` immediately before execution. The gauge is decremented exactly once on `success`, `transient_failure`, and `canceled` outcomes. Under concurrency it equals the number of valid submissions still executing and is zero after all submissions finish.
10. A worker panic is recovered at the service boundary. The recovered submission is classified as `transient_failure`. Exactly one duration sample is observed. The gauge is decremented to zero. The service returns an error. Panic detail never appears in any return value or log path.
11. The duration histogram observes through an injected observation seam. The seam is the only path through which a duration sample reaches the histogram in tests.
12. The exporter constructor returns a clean typed duplicate-registration outcome when the same registry is reused. Two distinct registries may each hold the same descriptors and never leak values between them.
13. The `/healthz` endpoint responds to `GET` with status `200`, content type `application/json`, and the exact compact body `{"status":"ok"}` with no trailing newline.
14. The `/metrics` endpoint responds to `GET` with the Prometheus text exposition format gathered from the injected registry.
15. Counters never decrease during operation. No code path drops a counter to zero.
16. The exporter is a learning surface; no push gateway, alert manager, dashboards, host-level metrics, retention model, retention backend, multi-instance federation, or production availability claim is part of this project.

## 9. Inputs and Outputs
- Inputs are Work submissions carrying a mode drawn from the fixed set `success`, `transient_failure`, and `block_until_canceled`, plus a context.
- Outputs are an outcome chosen from the fixed enumeration, an observed duration, a changed request count, a changed outcome count, and a returned-to-zero in-progress gauge for the valid path. HTTP output from `/healthz` is the exact compact body `{"status":"ok"}` with no trailing newline. HTTP output from `/metrics` is the Prometheus text exposition format gathered from the injected registry.
- Text-only behaviour example. Submit one valid submission whose mode is `success`. After the call, `tutorial_work_requests_total` has increased by one; `tutorial_work_outcomes_total{outcome="success"}` has increased by one; `tutorial_work_in_progress` is zero; and `tutorial_work_duration_seconds{outcome="success"}` reports one observation through the seam.
- Text-only behaviour example. Submit one submission whose mode is `block_until_canceled`, then cancel its context. The in-progress gauge returns to zero. The outcome counter increases by one under `outcome="canceled"`. The duration histogram reports one observation under `outcome="canceled"`. No other `outcome` value increases.
- Text-only behaviour example. Submit one submission whose mode is `unknown`. The validation rejection path recorded one increment on `tutorial_work_requests_total`, one increment on `tutorial_work_outcomes_total{outcome="validation_rejected"}`, one observation on `tutorial_work_duration_seconds{outcome="validation_rejected"}`, and the in-progress gauge never moved from zero.
- Text-only behaviour example. Drive a worker panic inside the service. The recovered submission recorded one request increment, one `transient_failure` outcome, one duration sample, and the gauge is zero. Panic detail does not appear in the return value or in any log path.

## 10. Rules and Edge Cases
- A submission whose mode is not in the fixed mode set is rejected as `validation_rejected`. The mode never appears as a label value.
- A submission that enters the service and runs to completion with `mode="success"` records `outcome="success"`. A submission whose execution returns a transient error records `outcome="transient_failure"`. A submission whose execution is cancelled through `context` records `outcome="canceled"`.
- A worker panic is recovered and classified as `transient_failure` at the service boundary; the gauge is returned to zero before the service returns an error.
- A scrape endpoint called before any submission exposes the four registered descriptors, including every preinitialized outcome label with zero count, rather than panicking or returning `404`.
- A health endpoint called concurrently with submissions returns the exact compact body `{"status":"ok"}` with no trailing newline and does not block on metric computation.
- Other HTTP methods on `/healthz` or `/metrics` return `405`. Unknown paths return `404`.
- The exporter constructor returns a clean typed duplicate-registration outcome when the same registry is reused. Distinct registries do not leak values between them.
- The deterministic observation seam is the only path through which a duration sample reaches the histogram in tests. Unit tests never sleep and never depend on wall-clock time.

## 11. Project Constraints
- The exporter is constructed against an injected registry. No global Prometheus registration or global gather paths exist in the project. Two exporter instances in one test may each register the same descriptors on distinct registries without conflict, and no descriptor crosses registries.
- The Prometheus Go client is the chosen dependency. The learner selects and pins a currently maintained module release in their own implementation. This guide does not invent a version.
- No push gateway, alert manager, dashboards, host metrics, multi-process aggregation, remote write, federation, or service-level objective evaluation is part of this project.
- Test execution uses `httptest`, the injected registry, and the deterministic observation seam. No Docker, no network scraper, no external Prometheus server, no wall-clock sleeps, and no race-prone shared global state.
- The exporter is a learning surface and explicitly is not a production monitoring system. Production monitoring concerns such as retention, alerting, security, capacity, and multi-tenant scraping are outside scope.

## 12. Design Questions Before Coding
- How are the four descriptors registered against the injected registry exactly once, and what does the constructor return on a duplicate registration?
- How does the Work service distinguish validation from execution, and where in the code path is each increment, gauge step, and observation placed?
- How does the service boundary recover a worker panic, classify it as `transient_failure`, observe duration, and return the gauge to zero?
- How does the deterministic observation seam feed the duration histogram in tests, and what does the test read from the registry to assert duration?
- How does the test prove that two distinct registries never leak values between them?
- How does the test prove that counters never decrease, and how does the test prove that a `validation_rejected` submission never increments the in-progress gauge?
- How does the `/healthz` endpoint return the exact compact body with no trailing newline, and how does the test assert it byte for byte?
- How does the `/metrics` endpoint return the text exposition format gathered from the injected registry only, and how does the test assert scrape content without depending on metric registration order?
- Why is the mode set a fixed enumeration, and why is mode never recorded as a metric label?
- Why does the project not promise production monitoring, and which production concerns are deliberately out of scope?

## 13. Implementation Milestones
1. Pin the four descriptors, the exact `outcome` label enumeration, the exact mode set, and the deterministic observation seam interface.
2. Implement the Work service skeleton with the validation step, the deterministic execution step, and the per-outcome branches.
3. Wire the request counter, the outcome counter, the duration histogram, and the in-progress gauge so the lifecycle is coherent across every outcome including validation rejection.
4. Wire the service-boundary panic recovery so a recovered submission classifies as `transient_failure`, observes one duration sample, decrements the gauge, and returns an error.
5. Construct the exporter against an injected registry. The constructor returns a clean typed duplicate-registration outcome on reuse.
6. Implement the `/healthz` endpoint with the exact compact body and no trailing newline. Implement the `/metrics` endpoint with the text exposition handler bound to the injected registry.
7. Implement the HTTP method discipline: other methods on either path return `405`; unknown paths return `404`.
8. Write the unit test suite that exercises every pinned descriptor, every outcome, the panic recovery, the duplicate-registration outcome, the registry-isolation invariant, the byte-exact `/healthz` body, the scrape content, and the deterministic observation seam.
9. Verify under the race detector and reproduce the honest non-production statement in the project documentation.

## 14. Verification Cases the Learner Must Write
- Descriptor content: assert the injected registry gathers exactly four metric families with the exact names, exact help text, exact label sets, and exact histogram buckets pinned in this guide, and no others.
- Request counter increment: submit one valid submission, one `validation_rejected` submission, one `canceled` submission, one `transient_failure` submission, and one panic-recovered submission; assert `tutorial_work_requests_total` increased by exactly five and never decreased.
- Outcome counter labels: assert that each of the four label values can appear in a single test sequence and that no other label value appears.
- In-progress gauge zero: for each of `success`, `transient_failure`, `canceled`, and `validation_rejected`, assert `tutorial_work_in_progress` is exactly zero after the call.
- Validation before execution: assert that a `validation_rejected` submission never incremented the gauge and that the gauge was zero before, during, and after the rejection path.
- Histogram observation through the seam: assert that the histogram reports exactly one observation per submission with the seam-supplied duration; assert that no test depends on wall-clock time or sleeps.
- Panic recovery: drive a worker panic and assert the recovered submission classified as `transient_failure`, observed one duration sample, returned the gauge to zero, and returned an error; assert panic detail does not appear in the return value or log path.
- Method discipline: assert `PUT /healthz`, `DELETE /metrics`, `GET /unknown` return `405` and `404` respectively, and that `/unknown` returns `404`.
- Health response: assert `GET /healthz` returns status `200`, content type `application/json`, and the exact compact body `{"status":"ok"}` with no trailing newline.
- Scrape content: assert `GET /metrics` returns the Prometheus text exposition format gathered from the injected registry and exposes the four registered descriptors, including every preinitialized outcome label with zero count, before any submission.
- Duplicate registration: assert that a second construction attempt on the same injected registry returns a clean typed duplicate-registration outcome rather than panicking.
- Registry isolation: assert that two exporter instances on two distinct registries do not share values and that neither scrape exposes the other's descriptors.
- Counter monotonicity: assert that no code path drops a counter to zero and that counters never decrease during operation.
- Determinism: assert that every duration assertion uses the observation seam and that no test relies on wall-clock time.

## 15. Common Mistakes to Watch For
- Reaching for the global default registerer or global default gatherer in any code path. The exporter must use the injected registry throughout.
- Treating the request counter or the outcome counter as a gauge. Counters never decrease.
- Letting a `validation_rejected` submission increment the in-progress gauge. Validation runs before execution.
- Letting any outcome branch forget to decrement the in-progress gauge, or letting two branches decrement it for the same submission.
- Letting a worker panic escape the service boundary so the gauge ends non-zero. Recovery at the boundary is required.
- Letting the histogram miss the observation seam and reading wall-clock time inside the assertion path.
- Letting an unbounded value reach a label set. The `outcome` label set is the fixed enumeration and nothing else.
- Letting the histogram receive an observation with a missing label. The label set is fixed per observation.
- Letting a second constructor call on the same registry panic. The constructor must return a clean typed outcome.
- Hiding race-prone shared state behind a global. The project uses an injected registry precisely so global state does not leak.
- Dropping a counter to zero through any test path. Counters never decrease during operation.
- Adding a `/metrics` response that reads from the global default registry. The endpoint gathers from the injected registry only.
- Returning a `/healthz` body with a trailing newline, with whitespace, or with a different shape than the exact compact form.
- Calling the result a production monitoring stack. It is a learning surface, and the documentation must say so.
- Pinning a Prometheus client version in this guide. The learner chooses and pins a currently maintained release at implementation time.
- Adding a push gateway, alert manager, dashboards, retention model, or host-level scrape that the project explicitly excludes.

## 16. Topics and References for Study
- The Prometheus Go client documentation covering the registry, the four metric types, label cardinality warnings, the text exposition format, and the HTTP handler integration.
- The Prometheus naming and label convention documentation covering counter, gauge, and histogram naming and stable label sets.
- The `net/http` documentation covering method routing and the `405` and `404` responses.
- The `httptest` documentation used in the test layer.
- Projects 046 and 060 are the formal prerequisites: Project 046 for HTTP basics and Project 060 for graceful-shutdown discipline. Project 095 is optional immediate-catalog-predecessor context; Projects 064 and 086 are optional study for migration and deterministic-test patterns.

## 17. Self-Assessment Questions
- Which four descriptor names, help texts, label sets, and histogram buckets are pinned?
- Why is the request counter a counter, why is the in-progress value a gauge, and why must the gauge return to zero on every outcome?
- Why does validation run before execution, and why does `validation_rejected` never increment the gauge?
- Why is the duration histogram fed through an injected observation seam instead of wall-clock time in tests?
- Why is the registry injected instead of using the global default registerer or gatherer?
- How does service-boundary panic recovery preserve the outcome, duration observation, gauge, and error contract without leaking panic detail?
- Why are `outcome` values fixed while mode is never a metric label?
- How do tests prove registry isolation, duplicate-registration handling, exact scrape content, and zero-valued preinitialized series?
- How do tests prove counter monotonicity and the exact `/healthz` response and HTTP method/path status rules?
- Which production monitoring concerns are outside scope, and how is the non-production claim documented?

## 18. Definition of Completion
The project is complete when the exporter is constructed against an injected registry and contains no reference to the global default registerer or global default gatherer; when the four descriptors are registered with the exact names, exact help text, exact label sets, and exact histogram buckets pinned in this guide; when the `outcome` label set is exactly `success`, `validation_rejected`, `transient_failure`, and `canceled`; when the Work service accepts only the mode set `success`, `transient_failure`, and `block_until_canceled`, rejects any other mode as `validation_rejected`, and never records mode as a metric label; when validation runs before execution so a `validation_rejected` submission never increments the in-progress gauge; when every submission increments `tutorial_work_requests_total` exactly once, increments exactly one series of `tutorial_work_outcomes_total`, observes exactly one sample on `tutorial_work_duration_seconds`, and on the valid path returns the gauge to zero on every outcome; when a worker panic is recovered at the service boundary so the recovered submission classifies as `transient_failure`, observes one duration sample, decrements the gauge, and returns an error; when the duration histogram observes through an injected observation seam and the test never depends on wall-clock time; when the constructor returns a clean typed duplicate-registration outcome on registry reuse and two distinct registries do not leak values between them; when `GET /healthz` returns status `200`, content type `application/json`, and the exact compact body `{"status":"ok"}` with no trailing newline, when `GET /metrics` returns the text exposition format gathered from the injected registry, when other methods on either path return `405`, and when unknown paths return `404`; when counters never decrease during operation; when the unit tests pass with `httptest`, covering descriptor content, counter increments, label values, histogram observations, gauge zero on every outcome, validation-before-execution, panic recovery, registry isolation, byte-exact health body, scrape content, method discipline, duplicate registration, and counter monotonicity; when the race detector is clean; when the project documentation reproduces the honest statement that the result is a learning surface and not a production monitoring system; and when this guide contains no implementation code, signatures, starter snippets, solution snippets, pseudocode, or implementation shell commands.

## 19. Optional Extensions
- A bounded second exporter instance type that distinguishes two operation kinds through a second bounded label set, registered on its own injected registry, demonstrating that label cardinality lives in the metric family and not in the registry identity.
