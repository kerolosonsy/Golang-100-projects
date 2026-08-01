# Project 089 — Raft Consensus Implementation

## 1. Project Name and Number
Project 089, raft_consensus_impl. Build a deterministic in-memory Raft simulation for a fixed cluster of three or five nodes, driven by a manual event loop, injected logical clock, scripted election-timeout source, and simulated message transport. This README is a learning guide only and contains no implementation code, signatures, snippets, pseudocode, or solution commands.

> **Scope.** This is not a production cluster. It has no sockets, goroutine-timing protocol, disk storage, snapshots, membership changes, authentication, client deduplication, or real process failure. Its purpose is to expose Raft state transitions and safety invariants under reproducible schedules.

## 2. Project Idea
Each node models follower, candidate, and leader roles; persistent-in-simulation term, vote, and log; volatile election, replication, commit, and apply state; and a bounded deterministic state machine. Tests advance logical time and choose which queued messages are delivered, delayed, duplicated, reordered, or dropped. The simulation must elect a majority leader, repair conflicting logs, commit safely, apply committed entries once in index order during one node incarnation, survive modeled restart, and preserve Raft's core safety invariants through partitions and healing.

## 3. Why This Project Now?
This project takes its required foundation from Projects 034 (worker_pool_basic), 041 (context_timeout_example), and 044 (fan_out_fan_in), which together supply bounded worker fan-out, context-scoped cancellation and timeouts, and many-to-many aggregation pipelines built on stable ordering of inputs. The catalog's immediate predecessor is Project 088 (lsm_tree_storage_engine); Project 088 is referenced here only as optional context for persistent versus volatile state distinctions, monotonically increasing sequence metadata, recovery ordering, corruption boundaries, and authoritative publication. This project carries those distinctions into a replicated-state-machine model, while deliberately keeping persistence and transport simulated so elections and message schedules remain deterministic.

## 4. Prerequisites
Projects 034 (worker_pool_basic), 041 (context_timeout_example), and 044 (fan_out_fan_in) are the required prerequisites. Project 088 (lsm_tree_storage_engine) is the immediate catalog predecessor and is useful for context but is not required. Understand majority quorums, monotonically increasing terms, log indexes and terms, finite-state transitions, deterministic pseudo-random sources, manual clocks, message queues, partitions, crash/restart models, and invariant-based tests. Read the extended Raft paper before designing states, but keep this project's fixed membership and in-memory limits explicit.

## 5. What You Must Know Before Starting
Know the difference between safety and liveness. Safety must hold under every tested delivery order; liveness requires eventual message delivery and suitable timeout separation. Understand why one vote per term prevents two majority winners, why log freshness is compared by last term before last index, why AppendEntries checks a predecessor, why a leader cannot directly commit an older-term entry merely by counting replicas, and why commit index and last applied are different. Also understand that repeated network messages must not repeat state-machine application.

## 6. Explanation of New Concepts
The cluster size is exactly three or exactly five, chosen at construction and fixed for its lifetime. Node IDs are distinct stable strings and all deterministic tie-breaking uses sorted node ID. Majority is more than half of the configured nodes: two in a three-node cluster and three in a five-node cluster. Commands are opaque UTF-8 strings of at most 4 KiB. Each log is capped at 10,000 real entries, and the simulated transport is capped at 100,000 queued envelopes so a faulty scenario remains bounded.

Logical time is an integer tick count that never moves backward. Heartbeat interval, minimum election timeout, and maximum election timeout are positive validated tick durations; heartbeat interval is strictly less than minimum election timeout, and minimum is no greater than maximum. Every election-deadline reset consumes exactly one value from an injected deterministic timeout source within the inclusive configured range. An out-of-range or exhausted scripted value is a simulation error, not a fallback to wall-clock randomness.

No timer or message runs automatically. Advancing the clock only makes due timer events eligible. The default event loop orders timer events by deadline and then node ID, and transport envelopes by delivery time and then monotonically assigned envelope sequence. Tests may explicitly select an eligible envelope out of order, move its delivery time later, duplicate it with a new envelope sequence, or drop it. Processing one event completes all of that node's synchronous state transition and resulting message enqueues before another event starts. There are no sleeps and no hidden goroutines.

Each log has a permanent sentinel at index zero with term zero and no client command. Real entries begin at index one and contain a term and command. Index is its position, not a separately trusted input. Two entries at one index conflict when their terms differ. If equal index and term carry different commands, the simulator reports an invariant violation rather than silently choosing one.

The fields current term, voted-for node or none, and the complete log are persistent in the simulation. A node updates current term and clears voted-for before processing any higher-term message further. Granting a vote records voted-for before the positive response is queued. Appending or truncating log entries completes before a successful AppendEntries response is queued. These are atomic in-memory transitions, not claims about disk durability.

Role, known leader, election deadline, collected votes, commit index, last applied, next indexes, match indexes, heartbeat deadline, and in-flight request bookkeeping are volatile. A fresh node starts as follower in term zero with no vote and only the sentinel. A stopped node processes no timers or messages. Restart retains current term, voted-for, and log, resets all volatile state, starts as follower with no leader hint, draws a new election timeout, and uses a fresh empty in-memory state-machine view.

A follower or candidate whose election deadline becomes due starts an election. It increments current term, becomes candidate, votes for itself, clears stale leader knowledge, draws a new deadline, and queues RequestVote messages carrying its current term and last-log identity. Vote responses are counted once per peer and only for the candidate's current term. A current-term majority makes it leader. A candidate that times out starts a newer election rather than continuing the old tally.

A voter first handles term comparison. A lower-term RequestVote is rejected. A higher term causes step-down, term update, and vote reset before evaluation. In the current term, a vote is grantable only if voted-for is empty or already names that candidate and the candidate's log is at least as up to date. Log freshness compares last-log term first; only equal last terms compare last-log index. Granting a vote resets the election deadline. Denial and stale requests do not reset it. An exact duplicate request from the already chosen candidate returns the same grant without creating a second vote.

On election, a leader initializes each follower's next index to one beyond its own last log index and match index to zero, records its own match at its last index, and sends immediate empty AppendEntries heartbeats. Later heartbeat timer events send replication to every follower and schedule the next heartbeat. A leader does not run an election timer while it remains leader.

Every protocol request and response carries a term. A node receiving any message with a higher term becomes follower, stores that term, clears its vote and leader-only volatile state, and then processes a request if applicable. Lower-term requests are rejected and lower-term responses are ignored. A candidate or leader receiving a current-term AppendEntries request from another node steps down before evaluating log consistency. A valid current-term leader contact records the leader hint and resets the election timeout even when predecessor consistency fails.

AppendEntries identifies a predecessor index and term, zero or more consecutive entries after it, and the leader's commit index. A follower rejects without changing its log when the predecessor index is beyond its last index or its term differs. A missing predecessor reports the follower's next available index. A term mismatch reports the conflicting term and the first index carrying that term. These deterministic hints accelerate repair but do not alter safety.

When the predecessor matches, the follower compares incoming entries in order. Existing entries with the same index and term remain. At the first different term, the follower deletes that entry and every later entry, then appends the incoming suffix. A repeated identical AppendEntries is therefore harmless. A successful response reports the highest replicated index. The follower never truncates merely because a heartbeat has fewer entries than its own log.

The leader uses successful current-term responses to advance that peer's match index monotonically and set next index to one after it. A rejection moves next index toward the supplied conflict boundary but never below one and never below a prefix already confirmed by a later success. Responses whose term, peer, or recorded request range is stale cannot undo newer replication knowledge. The leader retries from the revised predecessor on a later explicit send event.

A leader advances commit index to the highest index replicated on a majority, including itself, only when the entry at that index belongs to the leader's current term. Committing such an entry also commits every earlier index. Replication count alone cannot directly commit an entry from an older term. A follower advances commit index monotonically to the smaller of leader commit and its own last index after a successful AppendEntries; it never commits an index it does not hold.

After any event that advances commit index, a node applies each newly committed command in strictly increasing index order until last applied equals commit index. Duplicate messages and repeated commit values do not apply an index again during that node incarnation. The required educational state machine is an ordered applied-command record, making order and duplication observable.

Restart requires a narrow interpretation of exactly once. Because commit index, last applied, and application state are volatile in this simulation, restart creates a fresh state-machine view. After the restarted node relearns a committed prefix, it rebuilds that fresh view from index one in order; it replaces rather than appends to any pre-restart observable view. The project guarantees no duplicate application within one incarnation and one entry per index in the rebuilt view. It does not guarantee crash-proof exactly-once external effects.

Client commands are accepted only by a running leader. A follower or candidate rejects with its current leader hint when known; otherwise it returns Unavailable. A leader validates the bound, appends one current-term entry, records its own match, and queues replication. Acceptance identifies a pending log index; client success is not final until that index is committed and locally applied. Retrying a client request may append another command because client request IDs and deduplication are outside scope.

A partition is modeled only by dropping or withholding selected transport envelopes. A leader in a minority may append entries but cannot commit them. A majority partition may elect a higher-term leader and commit. When communication heals, higher terms force stale leaders to step down, and predecessor checks repair uncommitted conflicting suffixes. Committed entries must never be replaced.

Text-only scenarios are permitted. A normal election moves one follower to candidate and then leader after two votes in a three-node cluster. A split vote ends with no majority until a later timeout starts a higher term. An isolated old leader can append but not commit. After healing, it observes the newer term, becomes follower, deletes only its uncommitted conflicting suffix, and converges on the committed log.

## 7. Learning Objective
Design and verify a deterministic fixed-membership Raft simulation that models elections, term changes, heartbeats, log consistency and repair, majority replication, the current-term commit rule, ordered non-duplicate application, partitions, and volatile restart behavior while preserving safety invariants and making all production omissions explicit.

## 8. Functional Requirements
1. Support exactly three or five distinct fixed node IDs and compute majority from configured membership.
2. Use a non-decreasing logical clock, validated heartbeat and election ranges, and one injected deterministic timeout draw per deadline reset.
3. Process timers and transport envelopes only through a manual one-event-at-a-time loop with deterministic default ordering and explicit drop, delay, duplicate, and reorder controls.
4. Model follower, candidate, and leader roles plus persistent-in-simulation current term, voted-for, and log.
5. Start elections on due timeouts, increment term, self-vote, send RequestVote, deduplicate responses, and require a current-term majority.
6. Grant at most one candidate per term and only when its last log is at least as up to date by term-then-index comparison.
7. Send immediate and periodic heartbeats and step down on higher terms or a current-term leader AppendEntries.
8. Enforce predecessor index and term consistency, deterministic conflict hints, suffix repair, and harmless duplicate messages.
9. Maintain per-follower next and match indexes without allowing stale responses to undo newer confirmed progress.
10. Let a leader advance commit only through a majority index whose entry is from its current term.
11. Let followers advance commit only after successful leader contact and only through entries they hold.
12. Apply committed commands exactly once per node incarnation in increasing index order and rebuild a fresh view deterministically after restart.
13. Accept client commands only on leaders; return a known leader hint or Unavailable elsewhere; distinguish accepted-pending from committed success.
14. Retain only term, vote, and log across simulated restart and reset every documented volatile field.
15. Enforce bounded commands, logs, queued messages, and monotonic term, index, commit, and apply values.
16. Continuously check election safety, log matching, committed-prefix agreement, and state-machine ordering invariants in scenarios.

## 9. Inputs and Outputs
Inputs are cluster configuration, logical-time advances, scripted timeout draws, client command submissions, node stop and restart events, and explicit transport actions that deliver, drop, delay, duplicate, partition, or heal messages. Outputs are deterministic node states, queued envelopes, vote and replication responses, leader or Unavailable client outcomes, accepted log indexes, commit notifications, applied-command records, and invariant failures. No input or output uses sockets, disk, DNS, wall-clock sleeps, or goroutine timing.

## 10. Rules and Edge Cases
Election timeout equality is due. Timeout draws outside the inclusive range fail the scenario. Votes from old terms, unknown nodes, duplicate peers, or a prior candidacy are ignored. A node never votes for two candidates in one term. A stale message cannot reduce term, commit index, or last applied. A heartbeat still performs predecessor consistency checks. A follower never truncates a suffix without an incoming conflicting entry. Duplicate AppendEntries never duplicate log entries or application. Leader commit count includes itself. An older-term entry cannot be directly committed by majority counting, but becomes committed when a later current-term entry is safely committed. A stopped node ignores messages and timers. Restart does not retain leader role, commit index, last applied, timer, leader hint, votes received, next index, or match index. A minority cannot commit. Healing may remove uncommitted conflicts but never a committed entry.

## 11. Project Constraints
Deterministic in-memory simulation only; fixed membership of three or five nodes; bounded opaque commands; manual event loop; no real network, disk, goroutine protocol, real clock, snapshots, log compaction, pre-vote, leadership transfer, joint consensus, dynamic membership, read leases, linearizable read protocol, authentication, encryption, client sessions, or exactly-once client effects. The simulator demonstrates Raft reasoning, not production availability or durability. Real deployments require stable storage ordering, transport security, backpressure, operational metrics, snapshot installation, membership management, and extensive fault testing.

## 12. Design Questions Before Coding
Which fields survive restart and why? Exactly when is a new election timeout drawn? Why does log freshness compare term before index? Why may a repeated vote request receive the same grant? Why does a heartbeat need predecessor validation? When may a follower delete a suffix? How do stale responses interact with next and match indexes? Why can a leader directly commit only a current-term entry? How does that rule commit older entries indirectly? What does exactly-once application mean before and after this simulation's restart? Why can an isolated leader accept a pending command but not report committed success? Which assumptions are needed for liveness but not safety?

## 13. Implementation Milestones
1. Establish fixed cluster bounds, sentinel log indexing, persistent and volatile fields, and invariant checks.
2. Establish logical time, scripted timeout draws, deterministic timer ordering, and manual transport controls.
3. Establish term comparison, follower behavior, election deadline resets, and RequestVote freshness and one-vote rules.
4. Establish candidate elections, deduplicated majority counting, split-vote retries, and leader initialization.
5. Establish heartbeat scheduling, predecessor consistency, deterministic rejection hints, and follower suffix repair.
6. Establish leader next and match tracking that remains correct under duplicated and reordered responses.
7. Establish current-term majority commit, follower commit advancement, and ordered non-duplicate application.
8. Establish leader-only client acceptance, pending versus committed outcomes, and nonleader hints.
9. Establish stop, restart, fresh state-machine rebuild, partition, and healing behavior.
10. Complete table-driven deterministic scenarios and invariant checks with no sleeps or goroutine timing.

## 14. Verification Cases the Learner Must Write
- A three-node cluster elects one leader after a deterministic timeout and majority vote.
- A five-node cluster requires three votes and does not elect with two.
- A split vote yields no leader, then a later deterministic timeout elects in a higher term.
- Duplicate vote requests and responses never create extra votes.
- A stale candidate loses to a candidate with a fresher last-log term or, when terms tie, index.
- Heartbeats prevent follower elections while valid leader contact continues.
- Leader failure permits a new leader; restart returns the old node as follower with persistent fields retained and volatile fields reset.
- Lower-term requests and responses cannot change current state; every higher-term message causes required step-down and term persistence.
- A current-term AppendEntries causes a candidate or competing leader to step down.
- Missing predecessors and term mismatches produce deterministic conflict repair without deleting an unrelated matching prefix.
- Duplicate and reordered AppendEntries and responses converge without duplicate log entries or regressed confirmed progress.
- A current-term entry commits only after majority replication, then applies exactly once in index order.
- A leader in a minority partition cannot commit even though it may append locally.
- An entry from an older term is not directly committed by replica count; committing a later current-term entry commits the prefix.
- Followers never commit beyond their local last index and apply only through commit index.
- Repeated leader commit values and duplicate messages never reapply an index in one incarnation.
- Restart builds a fresh state-machine view in order after commit information returns and does not append to its old observable view.
- Nonleaders return a current hint when known or Unavailable when unknown; only leaders accept pending commands.
- Partition healing forces stale leaders to step down, repairs uncommitted conflicts, and preserves committed entries.
- Runtime invariants hold after every event: at most one leader per term, committed entries never conflict, matching log entries share their prefix, commit and apply indexes are monotonic within an incarnation, and applied order matches committed log order.
- Every scenario uses scripted time and message actions with no sleeps, goroutine scheduling, sockets, or disk.

## 15. Common Mistakes to Watch For
Using wall-clock timers; letting timer callbacks run automatically; drawing timeouts at inconsistent reset points; counting duplicate votes; comparing last index before last term; granting a second candidate in one term; ignoring higher terms in responses; treating a heartbeat as exempt from log checks; truncating a follower merely because the leader sent fewer entries; searching for conflicts without bounds; allowing stale rejection responses to undo later success; committing an old-term entry directly by replica count; applying through last log rather than commit index; applying one index twice after duplicate messages; retaining leader role across restart; calling volatile state durable; letting a minority commit; or treating accepted-pending as client success.

## 16. Topics and References for Study
Study Diego Ongaro and John Ousterhout's extended Raft paper, especially election safety, leader completeness, log matching, AppendEntries conflict handling, and the current-term commit restriction. Study deterministic simulation testing, logical clocks, event queues, state-machine safety, quorum intersection, partitions, and crash/restart models. Review Go documentation for deterministic random sources, sorting, container queues, and table-driven tests, but keep the protocol single-threaded. Review Projects 034 (worker_pool_basic), 041 (context_timeout_example), and 044 (fan_out_fan_in) for bounded worker fan-out, context-scoped cancellation and timeouts, and many-to-many aggregation discipline, and Project 088 (lsm_tree_storage_engine) for persistent-versus-volatile and recovery-ordering discipline.

## 17. Self-Assessment Questions
Which fields survive simulated restart, and which guarantees disappear because persistence and transport are simulated rather than real? What starts an election, and which timeout resets consume a scripted draw within the configured range? When may a voter grant a vote in the current term, and how is log freshness compared term-then-index when last terms differ? Which messages force step-down, and how is the new term stored before any further processing of that message? What exact predecessor conditions permit a follower to append incoming entries, and what happens at the first different term? When may a follower delete a suffix of its log, and why does a heartbeat with fewer entries never truncate an already matching log? How does a leader use next and match indexes, and why is the current-term commit rule necessary for safety? How does a follower learn commit, and what rule keeps a follower from committing an index it does not locally hold? Why do duplicate messages, duplicate AppendEntries, and repeated commit values not duplicate application within one node incarnation? What can a leader in a minority partition still do locally, and what does it fail to commit until communication heals?

## 18. Definition of Completion
- [ ] Projects 034 (worker_pool_basic), 041 (context_timeout_example), and 044 (fan_out_fan_in) are treated as the required prerequisites.
- [ ] Cluster has exactly three or five fixed nodes and a correct majority threshold.
- [ ] Manual logical time, scripted timeout randomness, and transport controls make every scenario deterministic.
- [ ] Persistent-in-simulation and volatile fields follow the documented restart contract.
- [ ] RequestVote enforces term handling, one vote per term, and term-then-index log freshness.
- [ ] Majority election, heartbeats, and higher-term or current-leader step-down are correct.
- [ ] AppendEntries enforces predecessor consistency, safe suffix repair, and duplicate-message idempotence.
- [ ] Next and match indexes remain correct under delayed, duplicated, and reordered responses.
- [ ] Leaders directly commit only current-term entries after majority replication; followers commit only held entries.
- [ ] State machines apply committed entries once per incarnation in exact index order and rebuild fresh after restart.
- [ ] Only leaders accept pending client commands; nonleaders return a hint or Unavailable.
- [ ] Election, split vote, failure, stale term, conflict, majority, minority, old-term, duplicate, partition, heal, and apply-order scenarios pass.
- [ ] Safety invariants are checked after every event.
- [ ] No production cluster, network, disk persistence, snapshots, or membership changes are claimed.
- [ ] Guide contains no implementation code, signatures, snippets, pseudocode, or solution commands.

## 19. Optional Extensions
- Add a deterministic event-trace visualizer that records terms, roles, messages, commit movement, and invariant checks without changing protocol state.
- Add bounded exhaustive exploration of short event schedules to search automatically for invariant violations and emit the minimal failing schedule.
