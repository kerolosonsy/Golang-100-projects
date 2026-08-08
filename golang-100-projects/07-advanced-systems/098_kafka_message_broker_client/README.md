# Project 098 — Kafka Message Broker Client

## 1. Project Name and Number

- Project 098, `098_kafka_message_broker_client`.
- Build a small producer and consumer service against Kafka using the maintained Go client `github.com/twmb/franz-go`, with the broker held behind an interface so unit tests use a deterministic fake and integration tests run against an externally started broker and are gated by a build tag plus an environment flag.
- This README is a learning guide only.
- It contains no implementation code, signatures, starter snippets, solution snippets, pseudocode, or implementation shell commands.
- Text-only input and output examples are permitted.

## 2. Project Idea

A bounded task-event envelope flows through a Kafka topic. The producer chooses the pinned client `github.com/twmb/franz-go`, supplies the entity identifier as the message key so events for the same entity land on the same partition in send order, requests acknowledgement from all in-sync replicas, and keeps the client's idempotent producer behavior enabled. The consumer joins a consumer group, reads records in partition order within each assigned partition, validates schema and version, performs an idempotent side effect keyed by record identifier, and commits the offset only after the side effect succeeds or after the original record is durably written to the dead-letter topic. The broker boundary is held behind an interface so unit tests use a deterministic fake. The integration suite is opt-in. Cancellation, rebalance, retry, dead-letter, and shutdown are defined at the boundary, and the project is honest about being at-least-once with per-entity ordering rather than exactly once or globally ordered.

## 3. Why This Project Now?

- Projects 037, 041, and 095 are the formal prerequisites: Project 037 contributes producer-consumer fundamentals, Project 041 contributes context and cancellation discipline, and Project 095 contributes event-driven delivery.
- Project 097 is optional immediate-catalog-predecessor context, while Project 086 is optional prior review for at-least-once vocabulary.

## 4. Prerequisites

- Projects 037, 041, and 095 are the formal prerequisites.
- Project 037 provides producer-consumer fundamentals; Project 041 provides context and cancellation discipline; Project 095 provides event-driven delivery.
- Project 097 is optional immediate-catalog-predecessor context.
- Optional prior review includes Project 086 for at-least-once vocabulary.
- Be comfortable with typed outcomes, contexts, fake interfaces for the broker boundary, build tags, environment-flag gating, the difference between a unit test and an integration test, and the honesty discipline of pinning a single client and not promising global behavior.

## 5. What You Must Know Before Starting

- A Kafka topic is split across one or more partitions. Records addressed to the same partition are read in send order by any consumer reading that partition. Records across partitions are not ordered relative to each other.
- A consumer group is a set of consumers that share work on a topic. A rebalance reassigns partition ownership between consumers in the group; in-flight records on a moved partition may be redelivered to the new owner.
- An offset is a per-partition position. Committing an offset after processing acknowledges that the records up to that offset have been processed; committing before processing risks losing work on a crash.
- An idempotent producer configuration is a client-side setting that de-duplicates retried sends within a producer session so the broker stores a record at most once per producer epoch and sequence. It is a property of the producer.
- Acknowledgement from all in-sync replicas is the acknowledgement class pinned by this project. The producer's idempotent producer behavior remains enabled at all times.
- A "poison message" is a record that the consumer cannot process because the envelope is malformed, the envelope exceeds the bound, or the schema version is unknown. Poison messages must not produce an infinite hot loop.
- The dead-letter destination is the source topic name with the suffix `.dlq`. The DLQ record retains safe metadata but not the original payload.
- Cancellation is cooperative. Shutdown must wait for in-flight work, abort cleanly on a moved partition, and not commit a partial record.
- The idempotency ledger is an in-memory boundary in the core learning implementation. Duplicate no-op is guaranteed only for the process lifetime. Restart durability requires an external transactional store and is outside this project.
- A fake broker is a deterministic in-process implementation behind the same interface the real client uses. A fake records sends, returns the configured outcomes, and lets the test simulate broker failure, latency, and rebalance without network I/O.

## 6. Explanation of New Concepts

### Concepts

- The project draws a hard line between the broker adapter, the producer service, and the consumer service.
- The broker adapter is the only place where `github.com/twmb/franz-go` is constructed.
- Both the producer and the consumer depend on the adapter through an interface; in unit tests the interface is implemented by a deterministic fake and the real client is not imported into the unit test binary.

- The pinned envelope contract is compact UTF-8 JSON with no insignificant whitespace and no trailing newline.
- Its fields are in the fixed order `schema_version`, `event_id`, `entity_id`, `event_type`, `occurred_at`, then `payload`;
- Schema version is exactly `1`. `event_id` and `entity_id` are 1-64 ASCII characters drawn from letters, digits, period, underscore, and hyphen. `event_type` is exactly `task.created`, `task.updated`, or `task.deleted`. `occurred_at` is caller-injected UTC RFC 3339 with exactly millisecond precision and the `Z` suffix. `payload` is embedded as valid JSON, not as a quoted JSON string, and its encoded value is 2 bytes through 16 KiB.
- The full serialized envelope is at most 20 KiB. `event_id` is the idempotency key. `entity_id` is the Kafka message key.
- The payload and the raw key never appear in any log path.

- The producer returns a typed outcome drawn from exactly `Accepted`, `ValidationRejected`, and `BrokerError`. `Accepted` carries the topic, partition, and offset from the successful acknowledgement. `ValidationRejected` carries the validation class. `BrokerError` carries the available topic, partition, and offset, the full lowercase SHA-256 hex digest of the key, and the error class;
- It never carries the payload or the raw key.

- The consumer joins a consumer group and processes records sequentially within each assigned partition.
- Different assigned partitions may run concurrently.
- Sequential processing within a partition prevents a later offset from being committed past an unfinished earlier record.
- The consumer commits the offset only after the side effect completes successfully or after the original record is durably written to the dead-letter topic.

- The retry and dead-letter policy is pinned.
- Transient processing failures get at most three total processing attempts using injected `100 ms` then `200 ms` backoff.
- Unit tests never sleep; the backoff is injected.
- Malformed, oversize, and unknown-version records are permanent and receive exactly one processing attempt.
- The dead-letter destination is exactly the source topic name with the suffix `.dlq`.
- The DLQ record retains the safe source topic, partition, and offset, the key digest, the event ID when valid, and the error class, and never retains the original payload.
- If the DLQ acknowledgement fails, the consumer does not commit the source offset and returns an error so the record can be redelivered.

- Cancellation and rebalance are exercised together.
- A cancel signal stops the consumer from fetching new records immediately.
- The consumer finishes the current owned record within a five-second grace.
- The consumer commits the current owned record only if processing or acknowledged DLQ completed while ownership remained.
- Otherwise the consumer does not commit and allows redelivery.
- On partition revocation the consumer does not begin new work and never commits work completed after ownership was revoked.

- The idempotency ledger boundary records event IDs.
- The core learning implementation may be in-memory and guarantees duplicate no-op only for the process lifetime.
- State honestly that restart durability requires an external transactional store and is outside this project.
- This limitation is one reason there is no exactly-once claim.

- Integration tests run against an externally started broker.
- The integration gate is the build tag `kafka_integration` plus the environment flag `KAFKA_INTEGRATION=1`.
- Broker addresses and a unique topic prefix are supplied by the test environment.
- The test creates isolated source and `.dlq` topics or refuses to run if isolation cannot be guaranteed.
- The test uses a unique consumer group.
- No setup commands appear in this guide.

- Unit tests cover producer validation, producer acknowledgement failure, stable per-entity keying, same-key ordering assumptions inside one partition, consumer success then commit, consumer process failure without commit, idempotent side effect on duplicate delivery, retry to the pinned cap, dead-letter after the bound, malformed envelope rejection, unknown schema version rejection, cancellation, rebalance-safe shutdown, and redaction of payload and raw key.

## 7. Learning Objective

- After completing this project you must be able to explain in your own words: why the client is pinned to `github.com/twmb/franz-go` and the acknowledgement class is pinned to all in-sync replicas;
- Why the idempotent producer behavior remains enabled;
- Why per-entity ordering depends on the message key and the broker's per-partition order;
- Why processing must be idempotent under at-least-once delivery;
- Why offset commit must happen after the side effect and not before;
- Why sequential processing within a partition prevents a later offset from being committed past an unfinished earlier record;
- Why a rebalance can cause redelivery;
- Why a poison message must be moved aside after a bounded number of attempts and must never loop;
- Why error messages carry contextual metadata but never the payload or the raw key;
- Why the broker boundary is held behind an interface;
- Why a unit test never imports the real broker client;
- Why integration tests are gated by a build tag and an environment flag;
- Why the idempotency ledger boundary is in-memory and why restart durability requires an external transactional store;
- And why this project does not promise exactly-once end-to-end behavior or global ordering.

## 8. Functional Requirements

1. The project uses exactly one Kafka client: `github.com/twmb/franz-go`. The learner selects and pins a currently supported release in their own module. This guide does not invent a version.
2. The broker is held behind an interface. Unit tests use a deterministic fake. The real client is imported only in the production code path and in the integration test files.
3. The producer requests acknowledgement from all in-sync replicas and keeps the client's idempotent producer behavior enabled.
4. The producer sends compact UTF-8 JSON with no insignificant whitespace or trailing newline. Fields are ordered exactly as `schema_version`, `event_id`, `entity_id`, `event_type`, `occurred_at`, then `payload`; schema version is exactly `1`. Validation runs before send.
5. `event_id` and `entity_id` are 1-64 ASCII characters drawn from letters, digits, period, underscore, and hyphen. `event_type` is exactly `task.created`, `task.updated`, or `task.deleted`. `occurred_at` is caller-injected UTC RFC 3339 with exactly millisecond precision and the `Z` suffix. `payload` is embedded as valid JSON rather than a quoted JSON string and is 2 bytes through 16 KiB. The full serialized envelope is at most 20 KiB.
6. `event_id` is the idempotency key. `entity_id` is the Kafka message key. The payload and the raw key never appear in any log path.
7. The producer returns a typed outcome drawn from exactly `Accepted`, `ValidationRejected`, and `BrokerError`. `Accepted` carries the topic, partition, and offset. `BrokerError` carries the available topic, partition, and offset, the full lowercase SHA-256 hex digest of the key, and the error class.
8. The consumer joins a consumer group and processes records sequentially within each assigned partition. Different assigned partitions may run concurrently.
9. The consumer commits the offset only after the idempotent side effect completes successfully or after the original record is durably written to the dead-letter topic. The consumer never commits a partial record.
10. The side effect is idempotent by `event_id`. A duplicate delivery performs the side effect at most once within the consumer's in-memory idempotency ledger for the process lifetime. Restart durability of the ledger requires an external transactional store and is outside this project.
11. Transient processing failures get at most three total processing attempts using injected `100 ms` then `200 ms` backoff. Unit tests never sleep.
12. Malformed, oversize, and unknown-version records are permanent and receive exactly one processing attempt.
13. The dead-letter destination is exactly the source topic name with the suffix `.dlq`. The DLQ record retains the safe source topic, partition, and offset, the key digest, the event ID when valid, and the error class, and never retains the original payload. If the DLQ acknowledgement fails, the consumer does not commit the source offset and returns an error so the record can be redelivered.
14. Cancellation stops the consumer from fetching new records immediately. The consumer finishes the current owned record within a five-second grace. The consumer commits the current owned record only if processing or acknowledged DLQ completed while ownership remained. Otherwise the consumer does not commit and allows redelivery.
15. On partition revocation the consumer does not begin new work and never commits work completed after ownership was revoked.
16. The integration test suite is gated by the build tag `kafka_integration` and the environment flag `KAFKA_INTEGRATION=1`. Broker addresses and a unique topic prefix are supplied by the test environment. The test creates isolated source and `.dlq` topics or refuses to run if isolation cannot be guaranteed. The test uses a unique consumer group.
17. No secrets, payloads, raw keys, or environmental credentials appear in logs. Errors carry contextual metadata: topic, partition, offset, key digest, and an error class.

## 9. Inputs and Outputs

### Interface Contract

- Producer inputs are the bounded task-event envelope, the entity identifier used as the message key, and a context.
- Producer outputs are a typed outcome drawn from exactly `Accepted`, `ValidationRejected`, and `BrokerError`, and the contextual metadata on the `BrokerError` path.
- Consumer inputs are records from assigned partitions, each carrying a key, an offset, a partition, and a payload.
- Consumer outputs are side-effect mutations through the in-memory idempotency ledger for the process lifetime, offset commits, dead-letter records to the `.dlq` destination, and a typed termination outcome.
- Text-only behaviour example. Send two events with different `event_id` values for the same entity in order and read them back through a consumer assigned to the same partition. The consumer observes both records in send order and each side effect runs once. Redeliver either event with the same `event_id`; that delivery becomes a process-lifetime no-op.
- Text-only behaviour example. Drive the producer with a malformed envelope. The producer's typed outcome is `ValidationRejected`; the broker receives nothing.
- Text-only behaviour example. Drive the consumer with an unknown schema version. The consumer rejects the record before any side effect and records it on the `.dlq` destination after the single permanent-attempt policy.
- Text-only behaviour example. Cancel the consumer mid-read while a record is owned and in progress. The current record finishes within the five-second grace. If processing or acknowledged DLQ completed while ownership remained, the record is committed. Otherwise the record is not committed and is redelivered.

## 10. Rules and Edge Cases

- A producer validation failure is rejected before send. The broker never observes the envelope.
- A broker error returns a typed outcome that carries topic, partition, offset, key digest, and error class. Payloads and raw keys never appear in error logs.
- A consumer side-effect failure classified as transient retries with backoff up to the bound; at the bound the record is moved to the `.dlq` destination. The retry schedule is the injected `100 ms` then `200 ms` backoff and is never a wall-clock sleep in unit tests.
- A consumer side-effect failure classified as permanent, including an unknown schema version, an oversize envelope, and a malformed envelope, receives exactly one processing attempt and is then moved to the `.dlq` destination.
- A duplicate delivery is consumed through the in-memory idempotency ledger by `event_id`. The ledger guarantees duplicate no-op only for the process lifetime.
- A rebalance during in-flight processing does not commit a partial record's offset. The record remains eligible for redelivery to the partition's new owner.
- A cancellation stops the consumer's fetch loop immediately. The current owned record terminates within the five-second grace. The commit policy above is observed.
- An integration test running against an externally started broker uses isolated source and `.dlq` topic names and a unique consumer group identifier; the test does not depend on global names. If isolation cannot be guaranteed the test refuses to run.
- A DLQ acknowledgement failure leaves the source offset uncommitted and returns an error so the record can be redelivered.

## 11. Project Constraints

- The unit test binary does not import `github.com/twmb/franz-go`. The broker boundary is held behind an interface.
- The Kafka client is exactly `github.com/twmb/franz-go`. The version is pinned in the learner's own module; this guide does not invent a version.
- The acknowledgement class is all in-sync replicas. The idempotent producer behavior is enabled at all times.
- The project explicitly does not promise exactly-once end-to-end behavior, global ordering, transactional exactly-once pipelines, a schema registry, multi-cluster replication, a Kafka administration interface, a custom broker, a custom protocol, or a production deployment claim.
- The integration suite is opt-in. The build tag `kafka_integration` and the environment flag `KAFKA_INTEGRATION=1` are required. The broker is started outside the program.
- No shell setup commands or Compose files appear in this guide. The integration suite's environment is described; the learner wires the start command at their module boundary.

## 12. Design Questions Before Coding

- Why is the Kafka client pinned to `github.com/twmb/franz-go` and the acknowledgement class pinned to all in-sync replicas?
- Why is the idempotent producer behavior kept enabled at all times?
- How is the broker boundary expressed as an interface so the unit test never imports the real client?
- How does the schema version of the envelope look in code, and at which boundary is it validated on receive?
- How does the message key interact with partition assignment to produce per-entity ordering, and what does the consumer assume about same-partition order?
- Where in the consumer flow does the offset commit live relative to the idempotent side-effect ledger update?
- Why does sequential processing within a partition prevent a later offset from being committed past an unfinished earlier record?
- What backoff schedule, cap, and classification rules separate transient from permanent processing failures?
- How is a rebalance modeled in the unit test without a real broker, and what does the consumer do when ownership of an in-flight partition is revoked?
- What contextual metadata does an error carry, and what is excluded from every error log?
- What build tag and environment flag gate the integration test, and how are source and `.dlq` topics and the consumer group isolated per test run?
- Why does the project not promise exactly-once, and how does that non-claim appear in the documentation?

## 13. Implementation Milestones

1. Pin the Kafka client to `github.com/twmb/franz-go`. Pin the acknowledgement class to all in-sync replicas and the idempotent producer behavior to enabled. Define the broker interface the producer and consumer depend on.
2. Define the versioned task-event envelope with schema version `1`, the fixed field order, the field validation rules, and the size bounds. Implement validation before send.
3. Implement the producer wrapper: entity-key message key, all in-sync-replica acknowledgement, idempotent producer behavior, typed outcomes, and contextual metadata on `BrokerError`.
4. Implement the broker fake for unit tests: record sends, simulate acknowledgement and failure, simulate the same-partition read order for shared keys, simulate rebalance events.
5. Implement the consumer read loop with sequential processing per partition and concurrency across partitions, schema and version validation, in-memory idempotency ledger update by `event_id`, and post-effect offset commit.
6. Implement retry, dead-letter, poison-message handling, and shutdown semantics on top of the broker boundary.
7. Implement cancellation handling: stop fetching immediately, finish the current owned record within the five-second grace, commit only if processing or acknowledged DLQ completed while ownership remained.
8. Implement contextual-error metadata. Ensure the payload, raw key, secrets, and environmental credentials never appear in any log path.
9. Write the unit test suite behind the non-integration build configuration.
10. Write the opt-in integration test suite behind the build tag `kafka_integration` and the environment flag `KAFKA_INTEGRATION=1`. Use isolated source and `.dlq` topic names and a unique consumer group identifier. Refuse to run if isolation cannot be guaranteed.
11. Verify under the race detector and reproduce the honest at-least-once statement in the documentation.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Producer validation: drive the producer with a malformed envelope, an oversize envelope, an unknown `event_type`, an `event_id` or `entity_id` outside the pinned character set, and a payload outside the pinned JSON size bound; assert each results in `ValidationRejected` and the broker records zero sends.
- Producer acknowledgement failure: configure the broker fake to return a leader-not-available class, send a record, and assert the typed outcome is `BrokerError` with the available topic, partition, offset, key digest, and error class. Assert the payload and raw key never appear in any log.
- Stable keying: send two records for the same entity and assert the broker fake records them with the same partition assignment and in send order.
- Same-key ordering assumption: through the broker fake, expose two records on the same partition in send order and assert the consumer reads them in that order before commit.
- Consumer success then commit: process a record, perform the side effect through the in-memory idempotency ledger, and assert the offset commit happens only after the ledger update.
- Process failure without commit: drive the consumer with a side effect that returns a transient failure, assert no offset commit, and assert the retry behavior with the injected backoff.
- Duplicate idempotency: deliver the same record twice and assert the ledger effect runs once across both deliveries. Assert the no-op-on-duplicate guarantee is process-lifetime only.
- Retry and dead-letter: drive the consumer with a transient failure and assert it retries up to the bound and then writes to the `.dlq` destination; drive it with a permanent failure and assert it writes to the `.dlq` destination after the single processing attempt and never retries.
- DLQ record shape: assert the `.dlq` record retains the safe source topic, partition, offset, key digest, event ID when valid, and error class, and never retains the original payload.
- DLQ acknowledgement failure: simulate a DLQ acknowledgement failure and assert the source offset is not committed and the consumer returns an error so the record can be redelivered.
- Malformed envelope: deliver a record whose envelope fails validation and assert the consumer rejects it without side effect, performs one processing attempt, and writes to the `.dlq` destination.
- Unknown schema version: deliver a record with a schema version the consumer does not know and assert the consumer rejects it as a permanent failure, performs one processing attempt, and writes to the `.dlq` destination.
- Cancellation: cancel through `context` mid-read; assert the current owned record terminates within the five-second grace; assert the commit happens only if processing or acknowledged DLQ completed while ownership remained; otherwise assert the offset is not committed and the record is redelivered.
- Rebalance-safe shutdown: simulate a rebalance event during processing and assert no partial offset commit, no new work after revocation, and that the record remains eligible for redelivery.
- Sequential processing per partition: simulate a slow side effect on one record and a faster side effect on the next; assert the next record is not committed before the earlier record in the same partition.
- Concurrent across partitions: simulate work in two assigned partitions and assert each partition's records are processed sequentially within that partition while the two partitions run concurrently.
- Opt-in broker round trip: with the build tag and environment flag set, run an isolated-topic consumer group against an externally started broker; assert records land in send order on the assigned partition; assert committed offsets advance; assert a rebalance does not corrupt offsets; assert isolation of source and `.dlq` topics and the consumer group identifier.
- Redaction: assert that error logs and structured error records never include the payload, a secret, a raw key, or an environmental credential.

## 15. Common Mistakes to Watch For

- Importing `github.com/twmb/franz-go` from a unit test file. The unit test binary must remain free of the real client.
- Treating processing as exactly once. The honest model is at-least-once with idempotent side effects, and the in-memory ledger guarantees duplicate no-op only for the process lifetime.
- Committing the offset before the side effect. The side effect must author the commit.
- Forgetting to validate the schema version on receive. An unknown version is a permanent failure and never a side effect.
- Letting the same-partition guarantee become a global ordering claim. Per-entity ordering is the strongest statement the project makes.
- Looping forever on a poison message. The bound and the dead-letter destination are the only correct response.
- Logging the payload, a raw key, a secret, or an environmental credential in any error path. Errors carry contextual metadata only.
- Mixing integration tests with unit tests by accident. The build tag `kafka_integration` and the environment flag `KAFKA_INTEGRATION=1` are the only gates.
- Reusing a global topic or group name across runs and corrupting offsets. The integration suite uses isolated source and `.dlq` topic names and a unique consumer group identifier per run.
- Pinning a Kafka client version in this guide. The learner picks the version.
- Promising exactly-once end-to-end behavior or transactional exactly-once. The project does not.
- Implementing a custom broker, a custom protocol, or a different Kafka client. The project pins exactly `github.com/twmb/franz-go`.
- Implementing a real sleep for backoff in unit tests. The backoff is injected.
- Letting the source offset commit when the DLQ acknowledgement fails. The source offset must remain uncommitted so the record can be redelivered.

## 16. Topics and References for Study

- The `github.com/twmb/franz-go` documentation covering producer and consumer construction, idempotent producer configuration, all in-sync replica acknowledgement, partition assignment, rebalances, and offset commit semantics.
- The Kafka documentation covering topics, partitions, message keys, consumer groups, offsets, and acknowledgement classes, treated as background.
- The Kafka documentation covering producer idempotence, the producer epoch and sequence model, and the broker's de-duplication window. Idempotence is a producer-side property.
- The Kafka documentation covering dead-letter patterns and poison-message handling. The project uses bounded retry and dead-letter recording without infinite loops.
- Projects 037, 041, and 095 are the formal prerequisites: Project 037 for producer-consumer fundamentals, Project 041 for context and cancellation discipline, and Project 095 for event-driven delivery. Project 097 is optional immediate-catalog-predecessor context; Project 086 is optional study for at-least-once vocabulary.

## 17. Self-Assessment Questions

1. Why does per-entity ordering require a stable message key, and what ordering does Kafka not promise?
2. Why are the client, all-in-sync-replica acknowledgement, and idempotent-producer settings pinned?
3. Why must the side effect be idempotent by `event_id` under at-least-once delivery, and what is the restart limitation of the in-memory ledger?
4. Why does the offset commit happen only after the side effect or durable DLQ write?
5. Why does sequential processing per partition matter, and how can a rebalance cause redelivery?
6. How do transient and permanent failures differ in attempt count, backoff, and DLQ handling?
7. Why must poison messages be bounded rather than retried forever?
8. Why are payloads, raw keys, secrets, and environmental credentials excluded from errors and logs?
9. Why is the broker boundary an interface, and what exact build-tag, environment-flag, topic, and group isolation gates integration tests?
10. Why does the project make no exactly-once end-to-end or global-ordering claim?

## 18. Definition of Completion

- [ ] The project is complete when `github.com/twmb/franz-go` is the pinned Kafka client and the version is chosen by the learner;
- [ ] When the broker boundary is held behind an interface so unit tests use a fake and the real client is not imported into the unit test binary;
- [ ] When the producer requests acknowledgement from all in-sync replicas, keeps the idempotent producer behavior enabled, validates the envelope before send, keys messages by `entity_id`, and returns a typed outcome drawn from exactly `Accepted`, `ValidationRejected`, and `BrokerError` with the contextual metadata pinned in this guide;
- [ ] When the consumer joins a group, processes records sequentially within each assigned partition, validates schema and version before any side effect, performs an idempotent side effect keyed by `event_id` through the in-memory ledger for the process lifetime, and commits offsets only after a successful side effect or after a durable write to the `.dlq` destination;
- [ ] When transient failures retry with the injected `100 ms` then `200 ms` backoff up to the bound, when permanent failures receive exactly one processing attempt and are written to the `.dlq` destination, and when the DLQ record retains the safe metadata and never retains the original payload;
- [ ] When a DLQ acknowledgement failure leaves the source offset uncommitted and returns an error for redelivery;
- [ ] When cancellation stops fetching immediately, finishes the current owned record within the five-second grace, commits only if processing or acknowledged DLQ completed while ownership remained, and otherwise does not commit and allows redelivery;
- [ ] When a rebalance is treated as a soft cancel on affected partitions and never commits a partial record or work completed after ownership was revoked;
- [ ] When errors carry contextual metadata only and never log the payload, raw key, secret, or environmental credential;
- [ ] When the integration suite is gated by the build tag `kafka_integration` and the environment flag `KAFKA_INTEGRATION=1`, uses isolated source and `.dlq` topic names and a unique consumer group identifier, refuses to run if isolation cannot be guaranteed, and runs against an externally started broker;
- [ ] When the unit tests pass with the broker fake and cover producer validation, acknowledgement failure, stable keying, same-key ordering, success-then-commit, transient failure without commit, duplicate idempotency, retry and dead-letter, DLQ record shape, DLQ acknowledgement failure, malformed envelope, unknown schema version, cancellation, rebalance-safe shutdown, sequential processing per partition, concurrent processing across partitions, and contextual metadata redaction;
- [ ] When the race detector is clean;
- [ ] When the project documentation reproduces the at-least-once statement, the per-entity ordering statement, the in-memory idempotency ledger limitation, and the explicit non-claims about exactly-once and global ordering;
- [ ] And when this guide contains no implementation code, signatures, starter snippets, solution snippets, pseudocode, or implementation shell commands.

## 19. Optional Extensions

- A second bounded envelope event type sharing the same topic and consumer group, with the consumer dispatching on `event_type` only after schema-version validation, without changing the per-entity key or offset discipline.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 037 — Producer Consumer](../../03-concurrency/037_producer_consumer/README.md#20-prerequisite-based-documentation-guide), [Project 041 — Context Timeout Example](../../03-concurrency/041_context_timeout_example/README.md#20-prerequisite-based-documentation-guide), [Project 095 — Microservice with Event-Driven Outbox](../../07-advanced-systems/095_microservice_event_driven/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`github.com/twmb/franz-go`](https://pkg.go.dev/github.com/twmb/franz-go).
- **Standards and concept references:** [Apache Kafka design documentation](https://kafka.apache.org/documentation/#design), [Kafka producer configuration](https://kafka.apache.org/documentation/#producerconfigs), [Kafka consumer configuration](https://kafka.apache.org/documentation/#consumerconfigs).

### Project-specific learning focus

- **Learn now:** topics and partitions, keyed ordering, consumer groups and rebalances, manual offset commits, producer idempotence boundaries, retries, poison messages, dead letters, and observability.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
