# Project 100 — Production-Ready SaaS Backend

## 1. Project Name and Number

- Project 100, `100_production_ready_saas_backend`.
- Build a small production-oriented learning backend for team and task management using `net/http`, `database/sql` with the `pgx` standard-library adapter, PostgreSQL as the source of truth, and opaque server-side cookie sessions.
- The directory name is aspirational curriculum wording: completion of this project is not a claim that the result is production-ready.
- This README is a learning guide only.
- It contains no implementation code, signatures, starter snippets, solution snippets, pseudocode, or implementation shell commands.
- Text-only input and output examples are permitted.

## 2. Project Idea

A user registers, logs in, and operates inside a team. Each team has explicit member roles and owns projects. Each project owns tasks. Tasks carry status, priority, assignee, timestamps, and an optimistic-update discipline. The HTTP layer exposes a bounded versioned JSON API under `/v1` with consistent compact JSON errors, a request identifier header, pagination with a stable cursor and stable ordering, bounded validation, and idempotent task creation. Authentication uses bcrypt with cost `12` and opaque server-side cookie sessions stored in PostgreSQL. The capstone is bounded and milestone-driven; every layer is independently testable. External dependencies are gated by build tag and environment flag. Completion of this project is not a claim that the result is production-ready.

## 3. Why This Project Now?

- Project 099 is optional immediate-catalog-predecessor context only.
- The formal prerequisite is completion of all Level Gates 1 through 6, plus Projects 086, 091, 095, and 096.
- The six level gates establish readiness across the first six curriculum levels;
- Project 086 contributes deterministic advanced-system policy and testing discipline, Project 091 contributes an advanced API capstone, Project 095 contributes transactional event-driven delivery, and Project 096 contributes bounded Prometheus observability, injected-registry isolation, and deterministic metrics testing.

## 4. Prerequisites

- The formal prerequisite is completion of all Level Gates 1 through 6, plus Projects 086, 091, 095, and 096.
- The first six level gates establish readiness across the curriculum's first six levels.
- Project 086 provides deterministic advanced-system policy and testing discipline; Project 091 provides an advanced API capstone; Project 095 provides transactional event-driven delivery; Project 096 provides bounded Prometheus observability, injected-registry isolation, and deterministic metrics testing.
- Project 099 is optional immediate-catalog-predecessor context only.
- Be comfortable with PostgreSQL through its standard library and the `pgx` adapter, versioned migrations, transactional repositories, opaque cookie sessions, bcrypt, validation surfaces, optimistic concurrency, pagination cursors, idempotent create operations, structured logging with redaction, Prometheus metrics with bounded labels, graceful shutdown, and gated integration tests.

## 5. What You Must Know Before Starting

- Tenant isolation is enforced at two layers: at the authorization layer, which decides whether a request may operate on a given team or project, and at the data-access layer, which composes team-scoped predicates so a forgotten check at the call site does not leak rows.
- "Cross-team ID guessing" is when a caller supplies an identifier that lives in another team and hopes the access check lets them in. The defense is that every query is composed with a team-scoped predicate and every mutation requires authorization at the team boundary before the row is touched.
- "Mass assignment" is when a request body carries fields the server should not let the caller set directly. The defense is a typed input shape the server itself composes.
- "Role escalation" is when a member who is not an owner or admin performs an operation reserved for those roles. The defense is role checks at the authorization layer that consult the explicit role recorded on the team membership.
- Optimistic update is a discipline in which a row carries a version identifier and an update carries the version the caller saw; the update applies only when the row's current version equals the supplied version, otherwise the update reports `409 version_conflict`.
- Idempotent creation is a discipline in which a create request carries an `Idempotency-Key` header, the server records the key and the response, and a repeat of the same key replays the same outcome rather than creating a duplicate.
- Pagination with a stable cursor means the cursor encodes the last tuple and the current filters, the order is fixed, and inserts after page one do not appear on later pages.
- Password hashing is performed by `bcrypt` with cost `12`. Passwords are never plaintext, never reversible, and never logged.
- Generic login failure messages prevent enumeration: a wrong password and an unknown user produce the same `401 invalid_credentials` response.
- Sessions are opaque server-side cookie sessions. The cookie value is a 32-byte crypto-random opaque token; only the SHA-256 digest of the token is stored in PostgreSQL.
- Secret configuration comes from configuration sources that the deployment owns; secrets are never baked into the binary, never logged, and never returned in error responses.
- The service assumes TLS termination before it and does not trust forwarded client-IP headers in core scope. Remote IP is derived from the direct connection only.
- Readiness and liveness are separate signals. Liveness reports whether the process is alive. Readiness reports whether the process is ready to accept traffic, including database reachability.
- Redis is out of core. Background jobs are an optional extension only.

## 6. Explanation of New Concepts

### Concepts

- The project owns its domain layer, its repository layer, its authorization layer, its HTTP layer, its authentication layer, its observability layer, and its operational layer.
- Each layer's boundary is narrow so the unit test targets one layer at a time.

- The pinned tech core is `net/http` for the HTTP layer, `database/sql` with the `pgx` standard-library adapter for the storage layer, PostgreSQL as the source of truth, and opaque server-side cookie sessions for authentication.
- The learner selects currently supported library releases and pins them in their own module; this guide does not invent versions.

- The domain layer defines typed shapes for user, team, membership, project, and task.
- Resource identifiers are server-generated UUID version 4 strings.
- Registration accepts a unique normalized lowercase ASCII email of 3-254 characters, a display name of 1-120 trimmed UTF-8 characters, and the bounded password.
- Roles are exactly `owner`, `admin`, and `member`.
- Task status is exactly `todo`, `in_progress`, or `done`.
- Task priority is exactly `low`, `medium`, or `high`.
- Other names and titles are 1-120 trimmed UTF-8 characters.
- Task description is at most 4,000 characters.
- Project description is at most 1,000 characters.
- Team creator is the team's sole owner.
- There is exactly one active owner per team.
- An ownership transfer atomically changes the selected active member to owner and the previous owner to admin.
- Task assignee is either absent or an active membership in the same team.
- Server timestamps are injected UTC.
- Task version starts at `1` and increments exactly once per successful update.
- The domain layer is pure functions and does not touch a database, network, or filesystem.

- Authentication pins bcrypt with cost `12`.
- Passwords are 12-72 bytes.
- Login uses normalized email plus password and returns `401 invalid_credentials` for both wrong password and unknown user.
- The session token is a 32-byte crypto-random opaque token.
- Only its SHA-256 digest, user identifier, created time, and expiry are stored in PostgreSQL.
- The raw token is delivered only in a cookie named `tutorial_session` with `Secure`, `HttpOnly`, `SameSite=Strict`, `Path=/`, and a 24-hour absolute expiry. `POST /v1/logout` revokes the current session. `POST /v1/password/change` verifies the current password, stores the replacement hash, and revokes every session for the user in one transaction.
- The raw cookie, raw token, password, and password hash never appear in any log path.

- Permissions are pinned exactly.
- An `owner` can read and write all team resources, manage all memberships and roles, transfer ownership, and delete the team.
- An `admin` can read all team resources, add or remove only `member` memberships, and create, update, and delete projects and tasks; an admin cannot grant or revoke owner or admin, remove the owner, transfer ownership, or delete the team.
- A `member` can read team, project, and task resources, create an unassigned task or a task assigned to self, and update only `status` on tasks assigned to self; a member cannot manage memberships or projects, assign another user, reassign, delete tasks, or edit other task fields.
- Owner and admin may assign only active members of that team.
- All cross-team access returns the same `404 not_found` response as a nonexistent resource.

- The bounded API is exact: `POST /v1/register`, `POST /v1/login`, `POST /v1/logout`, and `POST /v1/password/change`; `POST /v1/teams`, `GET /v1/teams`, `GET /v1/teams/{team_id}`, `DELETE /v1/teams/{team_id}`, and `POST /v1/teams/{team_id}/ownership-transfer`; `POST /v1/teams/{team_id}/memberships`, `PATCH /v1/teams/{team_id}/memberships/{user_id}`, and `DELETE /v1/teams/{team_id}/memberships/{user_id}`; `POST` and `GET` on `/v1/teams/{team_id}/projects`, plus `PATCH` and `DELETE` on `/v1/teams/{team_id}/projects/{project_id}`; and `POST` and `GET` on that project's `/tasks` collection, plus `GET`, `PATCH`, and `DELETE` on its `/tasks/{task_id}` item.
- Password reset is absent because email delivery is out of scope.

- Errors are compact `application/json` with exactly `request_id`, `code`, and `message`.
- The pinned mappings are `400 validation_error` for malformed input, `401 authentication_required` for missing or expired session and `401 invalid_credentials` for login failures, `403 forbidden` only when membership exists but role is insufficient, `404 not_found` for missing or cross-team resources, `409 version_conflict` for stale optimistic updates, `409 idempotency_conflict` for `Idempotency-Key` reuse with a different request, `429 rate_limited` for rate-limit outcomes, `500 internal_error` for unhandled server errors, and `503 not_ready` when readiness fails.
- Internal details never reach clients. `X-Request-ID` accepts 1-64 ASCII characters drawn from letters, digits, period, underscore, colon, and hyphen; any value outside that set is replaced with a safe value, and the safe value is echoed and logged.

- Idempotent task creation requires an `Idempotency-Key` header of 16-128 printable ASCII non-space characters.
- The key is scoped to authenticated user, team, and endpoint.
- PostgreSQL stores the key, the normalized request hash, the response status, and the task identifier in the same transaction as task creation.
- The same key with the same request replays the original response.
- The same key with a different request returns `409 idempotency_conflict`.
- Records live for at least 24 hours.

- Task optimistic update requires the caller to supply the current integer version.
- The conditional update increments the version on match.
- A stale value returns `409 version_conflict` and changes nothing.

- Pagination has a default page size of `20` and an allowed range of `1-100`.
- Ordering is `created_at` descending then identifier descending, both server-generated.
- The cursor is opaque and encodes the last tuple and the current filters.
- Inserts after page one sort before the page-one cursor and therefore do not appear on later pages; no duplicates and no skips are caused by such inserts.
- A malformed or mismatched cursor returns `400 validation_error`.

- Migrations are versioned and bring the schema up.
- Both `up` and `down` migrations are required and are tested only on an isolated integration database.
- Forward-safety checks are not part of the pinned discipline.
- Integration tests are gated by the build tag `postgres_integration` plus the environment flag `POSTGRES_INTEGRATION=1`.

- Observability binds to the run.
- Structured logs include a request identifier, operation name, outcome class, duration, and permitted contextual identifiers, with sensitive fields redacted.
- An injected registry exposes exactly `saas_http_requests_total`, help text `Total HTTP requests by method, route, and status class.`, and `saas_http_request_duration_seconds`, help text `Duration of HTTP requests in seconds by method, route, and status class.` Both use only labels `method`, `route`, and `status_class`; route is the bounded route template, never the raw path.
- The histogram buckets are `0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5`. `GET /livez` returns status 200, `application/json`, and exact body `{"status":"alive"}` with no trailing newline. `GET /readyz` returns the same content type and exact body `{"status":"ready"}` only after migrations and a current database ping succeed; otherwise it returns the standard `503 not_ready` error.
- Graceful shutdown drains in-flight requests within 10 seconds.
- Database and handler work use explicit contexts.

- Rate limits use an injected clock.
- For each remote IP, the first three registration attempts in an hour are allowed and the fourth is rejected; the first five login attempts in a minute are allowed and the sixth is rejected.
- For each authenticated user, the first 60 task-create attempts in a minute are allowed and the 61st is rejected.
- A rejection returns `429 rate_limited` without handler mutation.
- Remote IP comes from the direct connection only.

- Audit events record membership changes, role changes, ownership changes, denied role-escalation attempts, login success and failure, logout, password change, and team deletion in PostgreSQL.
- Each successful domain change and its audit row share one transaction; standalone authentication outcomes write their audit event transactionally.
- Audit records are append-only through the public API and redact secrets.

## 7. Learning Objective

- After completing this project you must be able to explain in your own words: why tenant isolation is enforced at both the authorization and the data-access layer;
- Why mass assignment is blocked by a typed input shape;
- Why role escalation is blocked by an explicit membership record;
- Why the role permissions are pinned exactly and the same-team and cross-team matrix is exercised;
- Why optimistic update requires both a row-side version and a caller-supplied version;
- Why idempotent task creation requires an `Idempotency-Key` header and a server-side record;
- Why pagination uses a stable cursor and a stable ordering;
- Why bcrypt with cost `12` is the chosen password-hashing discipline and why passwords are never plaintext, never reversible, and never logged;
- Why generic login failure messages are the only correct response;
- Why sessions are opaque server-side cookie sessions and why only the token digest is stored;
- Why TLS termination is assumed before the service and why forwarded client-IP headers are not trusted in core scope;
- Why readiness and liveness are separate signals and why readiness includes database reachability;
- Why Prometheus metrics use bounded labels and never unbounded user data;
- Why logs are redacted before emission;
- Why Redis is out of core;
- Why background jobs are an optional extension only;
- Why audit events share a transaction with their domain change;
- Why the capstone is bounded and how each milestone keeps the scope honest;
- And why completion of this project is not a claim that the result is production-ready.

## 8. Functional Requirements

1. The pinned tech core is `net/http`, `database/sql` with the `pgx` standard-library adapter, PostgreSQL as the source of truth, and opaque server-side cookie sessions. The learner selects currently supported library releases and pins them in their own module. This guide does not invent versions.
2. Registration uses a unique normalized lowercase ASCII email of 3-254 characters, a display name of 1-120 trimmed UTF-8 characters, and a 12-72-byte password. Login uses normalized email and password. Bcrypt cost is `12`, and unknown user and wrong password both return `401 invalid_credentials`.
3. Sessions are 32-byte crypto-random opaque tokens whose SHA-256 digest, user identifier, created time, and expiry are stored in PostgreSQL. The raw token appears only in `tutorial_session` with `Secure`, `HttpOnly`, `SameSite=Strict`, `Path=/`, and 24-hour absolute expiry. Logout revokes the current session. Password change verifies the current password, changes the hash, and revokes all sessions in one transaction. Raw cookie, token, password, and password hash never appear in logs.
4. Resource identifiers are server-generated UUID version 4 strings. Roles are exactly `owner`, `admin`, and `member`; status is `todo`, `in_progress`, or `done`; priority is `low`, `medium`, or `high`. The pinned text bounds, sole-owner rule, atomic ownership-transfer rule, assignee-membership rule, injected UTC timestamps, and task version lifecycle all apply.
5. Permissions use the exact matrix in this guide: owner controls all resources; admin manages only member memberships plus projects and tasks; member creates only unassigned or self-assigned tasks and updates only status on tasks assigned to self. Owner/admin assignment targets an active member. Cross-team access returns `404 not_found`.
6. The bounded `/v1` surface is exactly the auth, team, ownership-transfer, membership, project, and task method/path set pinned in the concepts section. No password reset or unlisted route exists.
7. Errors are compact `application/json` with exactly `request_id`, `code`, and `message`. The pinned mappings are `400 validation_error`, `401 authentication_required`, `401 invalid_credentials`, `403 forbidden` only when membership exists but role is insufficient, `404 not_found` for missing or cross-team resources, `409 version_conflict`, `409 idempotency_conflict`, `429 rate_limited`, `500 internal_error`, and `503 not_ready`. Internal details never reach clients. `X-Request-ID` accepts one value of 1-64 ASCII letters, digits, period, underscore, colon, or hyphen; empty, repeated, or invalid values are replaced by one generated safe value that is echoed and logged.
8. Idempotent task creation requires the pinned `Idempotency-Key` and transactional record. Same key and request replay; same key with a different normalized request returns `409 idempotency_conflict`; records remain for at least 24 hours.
9. Task update requires the current positive integer version, applies one conditional update, increments once on success, and returns `409 version_conflict` without mutation on a stale value.
10. Pagination defaults to 20, accepts 1-100, orders by server-generated creation time descending then identifier descending, and uses an opaque cursor holding the last tuple and current filters. Later inserts sort before the cursor and do not appear on subsequent pages. Invalid or mismatched cursors return `400 validation_error`.
11. PostgreSQL is the source of truth. Both versioned `up` and `down` migrations are required and tested only on the isolated integration database gated by `postgres_integration` and `POSTGRES_INTEGRATION=1`.
12. Every domain query and mutation enforces tenant isolation. The same-team and cross-team matrix is exercised in tests.
13. The HTTP layer enforces the consistent compact error shape, validation, pagination, idempotency, rate limits, and explicit contexts for database and handler work.
14. Rate limits use an injected clock. The first 3 registration attempts per remote IP per hour, first 5 login attempts per remote IP per minute, and first 60 authenticated task-create attempts per user per minute are allowed; the next attempt is rejected with `429 rate_limited` and no mutation. Remote IP comes only from the direct connection.
15. Structured logs include a request identifier, operation name, outcome class, duration, and permitted contextual identifiers. Sensitive fields are redacted before emission.
16. An injected registry exposes exactly the two HTTP metric families, help text, labels, and duration buckets pinned in this guide. Labels use method, bounded route template, and status class; raw paths and user data never appear.
17. `GET /livez` returns 200 and exact body `{"status":"alive"}`. `GET /readyz` returns 200 and exact body `{"status":"ready"}` only after migrations and a current database ping; otherwise it returns the standard `503 not_ready` error. Success bodies are compact `application/json` with no trailing newline.
18. Graceful shutdown drains in-flight requests within a 10-second window.
19. Audit events cover membership, role, ownership, denied escalation, login, logout, password change, and team deletion. Successful domain change and audit row share one transaction; audit records are append-only through the public API and redact secrets.
20. Redis is out of core. Background jobs are an optional extension only and are not a core requirement, milestone, or completion condition.
21. The capstone is bounded by the milestones. Each milestone produces its own deliverable before the next milestone starts. The learner writes every line of implementation.
22. The integration test suite is gated by `postgres_integration` and `POSTGRES_INTEGRATION=1` and runs against an isolated PostgreSQL database.

## 9. Inputs and Outputs

### Interface Contract

- Inputs are the bounded JSON bodies for the pinned auth, team, membership, project, and task routes, plus `Idempotency-Key` on task creation, the current task version on update, and optional `X-Request-ID` on every request.
- Outputs are JSON errors with exactly the pinned three fields or bounded resource responses. Pagination includes a page and opaque cursor. Login sets `tutorial_session` with the pinned attributes. Every response echoes the safe `X-Request-ID` value.
- Text-only behaviour example. A user registers, logs in, creates a team, and becomes owner. The owner adds an existing second user as member, creates a project and task, assigns the task to that member, and updates it with the current version. A nonmember receives `404 not_found` for every guessed team, project, task, or membership identifier.
- Text-only behaviour example. The owner creates a task with an `Idempotency-Key`. A replay of the same `Idempotency-Key` with the same request returns the same task. A replay of the same `Idempotency-Key` with a different request returns `409 idempotency_conflict`. The owner updates the task with a stale version; the response is `409 version_conflict` and the row's version is unchanged. A repeat of the update with the corrected version succeeds.
- Text-only behaviour example. Pagination lists tasks with `created_at` descending then identifier descending. An insert between page one and page two sorts before the page-one cursor and therefore does not appear on later pages; no duplicates and no skips are caused by such inserts. A malformed cursor returns `400 validation_error`.
- Text-only behaviour example. The process starts, the database becomes reachable, the up migration succeeds, and readiness flips from `503 not_ready` to `200`. The process receives a shutdown signal, stops accepting new traffic, drains in-flight requests within the 10-second window, and exits with a documented exit code.

## 10. Rules and Edge Cases

- A request whose body contains a field the server does not accept is rejected with `400 validation_error`. A field the server does accept is mapped into the typed input shape the server composes.
- A team-scoped read or mutation that arrives without a valid team membership at the time of the request is rejected with `404 not_found`.
- A request that references a team identifier the caller does not belong to returns `404 not_found`, even when the resource under the team matches a guessed identifier.
- A role escalation attempt by a non-owner, non-admin member is rejected with `403 forbidden` when the caller is a member with insufficient role, and recorded as an audit event.
- A task update with a stale version returns `409 version_conflict` and does not mutate the row.
- A repeated idempotent task creation with the same `Idempotency-Key` and the same request returns the existing task and records no audit event of consequence. A repeated `Idempotency-Key` with a different request returns `409 idempotency_conflict`.
- A pagination request whose cursor is malformed or mismatched returns `400 validation_error`. A pagination request that returns no additional items returns an empty page and a terminal cursor.
- A login with a wrong password and a login with an unknown user return the same `401 invalid_credentials` response and the same audit event class.
- A rate-limited request returns `429 rate_limited` without mutating handler state.
- A shutdown that exceeds the 10-second drain logs a bounded diagnostic and proceeds with the remaining close sequence.
- An empty, repeated, overlong, or invalid `X-Request-ID` is replaced with one generated safe value, and every response and log uses that value.
- Ownership transfer changes the selected active member to owner and the previous owner to admin in one transaction; no operation may leave zero or multiple active owners.
- Audit events are append-only through the public API. Successful domain changes and their audit rows share one transaction.

## 11. Project Constraints

- The pinned tech core is `net/http`, `database/sql` with the `pgx` standard-library adapter, PostgreSQL as the source of truth, and opaque server-side cookie sessions.
- The project defines a small team and task management domain. It does not introduce billing, email delivery, social login, file uploads, Kubernetes manifests, multi-region deployment, a custom identity provider, a full frontend, Redis in core, or background jobs in core.
- The capstone is bounded by the milestones. Each milestone produces its own deliverable before the next milestone starts.
- The integration test suite is gated by the build tag `postgres_integration` and the environment flag `POSTGRES_INTEGRATION=1` and uses an isolated PostgreSQL database.
- The project documentation includes the consistent compact error shape, the pagination cursor, the rate limits, the timeouts, the versioned `up` and `down` migrations, the pinned authentication approach, the bcrypt cost, the `tutorial_session` cookie attributes, the idempotency contract, the explicit non-claims about production readiness, and the explicit scope of the capstone.
- The project name is aspirational curriculum wording. Completion is not a claim that the result is production-ready. Production deployment requires configuration, transport security at the deployment boundary, secret management, operational monitoring, alerting, capacity planning, and compliance review that are outside this project.

## 12. Design Questions Before Coding

- How is the typed input shape for each endpoint composed to block mass assignment without rejecting legitimate fields?
- How is tenant isolation expressed at the repository layer so a forgotten authorization check does not leak rows?
- How is the bcrypt cost pinned and how is the generic login failure rendered?
- How is the session token generated, stored as its SHA-256 digest, and delivered through the `tutorial_session` cookie with the pinned attributes?
- How is the role matrix pinned and exercised by the same-team and cross-team tests?
- How is the optimistic-update version surfaced in the API and recorded on the row, and how does the update path handle a stale version?
- How is the `Idempotency-Key` request replay path keyed, and how is a replay distinguished from a conflict?
- How is the pagination cursor encoded so that two callers at different times see the same ordering for the same input, and how is an insert after page one excluded from later pages?
- How is the consistent compact error shape defined in code so that every endpoint uses it?
- How is rate limiting enforced at sensitive endpoints and mapped to the consistent error shape?
- How is redaction performed in the structured logging path so that passwords, tokens, secrets, and payloads never appear?
- How is the readiness signal composed of liveness, the up migration, and a database ping?
- How is the integration gate expressed in the test layer so that unit tests run without Docker or external services?
- What does the project promise to be, what does it explicitly not promise, and how does that non-claim appear in the documentation?

## 13. Implementation Milestones

1. Write the API contract as a typed shape that fixes every endpoint, the request shape, the response shape, the consistent compact error shape, the pagination cursor format, the idempotency contract, the optimistic-update contract, the authentication contract, and the role matrix. The contract is the deliverable; implementation follows.
2. Implement the domain layer with the typed shapes and the policies for tenant isolation, role enforcement, optimistic-update decisions, and idempotency decisions. The unit tests are pure-function tables.
3. Implement the schema and versioned `up` and `down` migrations. The migrations bring the schema up; the integration suite exercises the `up` and `down` migrations on an isolated integration database.
4. Implement the repository layer with team-scoped predicates and transactional multi-row mutations. The integration tests run against an isolated PostgreSQL database behind the build tag `postgres_integration` and the environment flag `POSTGRES_INTEGRATION=1`.
5. Implement authentication: registration, login, bcrypt with cost `12`, opaque 32-byte session tokens, the SHA-256 digest storage, the `tutorial_session` cookie with the pinned attributes, the generic `401 invalid_credentials` failure, session expiry, the logout path that revokes the current session, and the password-change path that revokes all sessions.
6. Implement team membership: team creation with the creator as the sole `owner`, explicit roles, role enforcement at every team-scoped authorization check, and the same-team and cross-team matrix.
7. Implement projects and tasks, including the optimistic-update discipline on tasks, the assignee membership rule, and the audit events for security-sensitive changes.
8. Implement the HTTP layer: routing under `/v1`, the consistent compact `application/json` error response, validation, pagination with a stable cursor, idempotent task creation through `Idempotency-Key`, rate limits at sensitive endpoints, and explicit contexts for database and handler work.
9. Implement observability and operations: structured logs with redaction, Prometheus metrics through `GET /metrics` with bounded labels, `GET /livez`, `GET /readyz`, graceful shutdown with the 10-second drain window, and the `X-Request-ID` echo.
10. Verify under integration tests for storage and HTTP, under unit tests for every layer, and under the race detector. Reproduce the explicit non-claims about production readiness in the project documentation.
11. Document the local-development surface, the consistent error shape, the pagination cursor, the rate limits, the timeouts, the versioned `up` and `down` migrations, the pinned authentication approach, the bcrypt cost, the `tutorial_session` cookie attributes, the idempotency contract, and the explicit non-claims.

## 14. Verification Cases the Learner Must Write

### Required Cases

- Pure domain and policy tables: each policy decision is exercised through a pure-function test in the domain layer.
- Handler tests with `httptest`: each endpoint is exercised through the HTTP layer against an in-process handler stack with fakes for the lower layers.
- Repository integration tests: each repository is exercised against the isolated PostgreSQL integration database behind the build tag and environment flag.
- Migration `up` and `down`: the migration suite runs the `up` migration and the `down` migration on the isolated integration database.
- Tenant-isolation matrix: the same-team and cross-team cases are exercised for every team-scoped operation, including guessed identifiers, and the cross-team case returns `404 not_found`.
- Role matrix: every row of the pinned role permissions matrix is exercised against the same-team and cross-team boundaries.
- Authentication and session cases: registration, login with correct credentials, login with wrong password, login with unknown user, expiry, logout that revokes the current session, password change that revokes all sessions, replay protection, and the generic `401 invalid_credentials` response.
- `X-Request-ID` echo: a request with a valid `X-Request-ID` echoes and logs the safe value; a request with a value outside the accepted character set is replaced with a safe value and the safe value is echoed and logged.
- Pagination stability: the cursor produces a stable order across two consecutive callers, across an insert between page one and page two, and across no duplicates and no skips caused by such inserts.
- Idempotent create replay and conflict: the same `Idempotency-Key` with the same request returns the existing task; the same `Idempotency-Key` with a different request returns `409 idempotency_conflict`; a different `Idempotency-Key` creates a new task.
- Optimistic update conflict: a task update with a stale version returns `409 version_conflict` and changes nothing; the row's version increments exactly once per successful update.
- Transaction rollback: a repository operation that fails mid-transaction leaves the database at its pre-operation state and writes no audit row.
- Concurrent update conflict: two callers updating the same row with the same version result in one success and one `409 version_conflict`.
- Rate limits: with an injected clock, assert registration attempts 1-3 are allowed and 4 is rejected per IP per hour, login attempts 1-5 are allowed and 6 is rejected per IP per minute, and task-create attempts 1-60 are allowed and 61 is rejected per user per minute; rejected attempts do not mutate handler state.
- Graceful shutdown: a shutdown signal drains in-flight requests within the 10-second window and exits with the documented code.
- Redaction: structured logs never contain the raw cookie, the raw session token, the password, the password hash, or the request payload.
- Liveness and readiness: assert the exact success statuses, content type, bodies, and no-trailing-newline policy; readiness returns the standard `503 not_ready` error until migrations and a current database ping succeed.
- Registry isolation: the metrics endpoint gathers exactly the two pinned metric families from the injected registry, with their exact help text, labels, and buckets, and never reads a global registry.
- Audit: membership, role, ownership, denied escalation, login, logout, password change, and team deletion outcomes are recorded with the pinned transactional behavior and redaction; audit records are append-only through the public API.
- End-to-end happy path and end-to-end error path: a full request from registration to authenticated team, project, and task operations succeeds; an error path through tenant isolation produces the consistent compact error shape with `request_id`, `code`, and `message`.

## 15. Common Mistakes to Watch For

- Composing team-scoped queries at the call site rather than inside the repository layer, where a forgotten check leaks rows.
- Trusting the body for the team identifier or the role, which permits cross-team ID guessing and role escalation.
- Letting a request body set fields the server does not accept, which is mass assignment.
- Treating a database-side `UPDATE` without an optimistic version check as a safe update.
- Treating a repeated idempotent create as a brand new create, which duplicates rows.
- Returning a stable order that depends on map iteration, raw query order, or caller-controlled timestamps, so pagination drifts across callers.
- Returning distinct messages for wrong password and unknown user, which leaks user enumeration.
- Logging the raw cookie, the raw session token, the password, the password hash, or the request payload in any log path, even under debug level.
- Mixing integration tests with unit tests so that unit tests require Docker or external services.
- Pinning a library version in this guide rather than at the learner module boundary.
- Inventing cryptography: a custom hashing scheme, a custom token format, a custom session encoding. The project delegates to established libraries.
- Treating a Redis integration test as required. Redis is out of core.
- Calling the result production-ready. Completion is a learning milestone, not a deployment claim.
- Letting the capstone scope creep into billing, email delivery, social login, file uploads, Kubernetes, multi-region deployment, a custom identity provider, a full frontend, Redis in core, or background jobs in core.
- Trusting forwarded client-IP headers in core scope. Remote IP is derived from the direct connection only.
- Returning internal details in `500 internal_error` responses. Internal details never reach clients.
- Reusing a request identifier outside the accepted character set without replacement. The safe value is the only value echoed and logged.
- Treating pagination inserts as able to appear on later pages. Inserts after page one sort before the page-one cursor and do not appear on later pages.

## 16. Topics and References for Study

- The `net/http` documentation for routing, middleware composition, request identifiers, cookie attributes, and shutdown hooks.
- The `database/sql` documentation for the chosen `pgx` adapter, transactional boundaries, and the pinned migration discipline.
- The `bcrypt` documentation for parameter selection, including the pinned cost `12`.
- The PostgreSQL documentation for the chosen adapter, the chosen `up` and `down` migrations, and the chosen isolation discipline.
- The Prometheus client documentation for the bounded-metrics discipline with route-template, method, and status-class labels.
- Formal prerequisites: completion of all Level Gates 1 through 6, plus Projects 086, 091, 095, and 096. The level gates establish readiness across the first six curriculum levels; Project 086 contributes deterministic advanced-system policy and testing discipline, Project 091 an advanced API capstone, Project 095 transactional event-driven delivery, and Project 096 bounded Prometheus observability, injected-registry isolation, and deterministic metrics testing. Project 099 is optional immediate-catalog-predecessor context only.

## 17. Self-Assessment Questions

1. Why must tenant isolation exist in both authorization and data access, and how does the same-team/cross-team matrix prevent leaks?
2. Why does the typed input shape block mass assignment and role escalation?
3. Why are roles pinned exactly, and how does the authorization matrix constrain each role's resource operations?
4. Why does optimistic concurrency require both a row version and caller-supplied version, and what happens on conflict?
5. Why does idempotent task creation require a scoped `Idempotency-Key` record, and how are replay and conflict distinguished?
6. Why do stable pagination cursors and ordering prevent duplicates or skips after inserts?
7. Why are bcrypt cost `12`, generic login failures, opaque digest-backed sessions, and strict cookie attributes security requirements?
8. Why are readiness and liveness separate, and what do migrations, database reachability, graceful shutdown, and explicit contexts contribute?
9. Why must metrics labels be bounded and logs redact cookies, tokens, passwords, hashes, payloads, secrets, and internal details?
10. Why are integration tests gated, Redis and background jobs optional, and completion explicitly not a production-readiness claim?

## 18. Definition of Completion

- [ ] The project is complete when the API contract is written as a typed shape that fixes every endpoint, the request shape, the response shape, the consistent compact error shape, the pagination cursor format, the idempotency contract, the optimistic-update contract, the authentication contract, and the role matrix;
- [ ] When the domain layer is implemented as pure functions with policies for tenant isolation, role enforcement, optimistic-update decisions, and idempotency decisions;
- [ ] When the schema is brought up by a versioned `up` migration and a `down` migration, and both migrations are tested on the isolated integration database;
- [ ] When the repository layer composes team-scoped predicates and runs every multi-row mutation inside a transaction;
- [ ] When authentication uses bcrypt with cost `12`, generates 32-byte crypto-random opaque session tokens, stores only the SHA-256 digest in PostgreSQL, delivers the raw token only through the `tutorial_session` cookie with `Secure`, `HttpOnly`, and `SameSite=Strict`, exposes the generic `401 invalid_credentials` response for wrong password and unknown user, expires sessions after 24 hours, revokes the current session on logout, revokes all sessions on password change, and never logs the raw cookie, raw token, password, or password hash;
- [ ] When team membership exposes the pinned `owner`, `admin`, and `member` roles and the authorization layer consults the membership record for every team-scoped read and mutation;
- [ ] When projects and tasks support the optimistic-update discipline on tasks and the assignee membership rule;
- [ ] When the HTTP layer enforces the consistent compact `application/json` error response, validation, pagination with a stable cursor, idempotent task creation through `Idempotency-Key`, rate limits at sensitive endpoints, and explicit contexts for database and handler work;
- [ ] When observability emits bounded Prometheus metrics through the injected registry and structured logs with redaction;
- [ ] When `GET /livez` reports process liveness with the exact compact JSON body and `GET /readyz` reports `200` only after the `up` migration and a database ping succeed;
- [ ] When graceful shutdown drains in-flight requests within the 10-second window;
- [ ] When audit events record membership changes, role changes, ownership changes, login success and failure, logout, and team deletion, share a transaction with their domain change, and are append-only through the public API;
- [ ] When the pure-function policy tables, the handler tests with `httptest`, the integration tests against the isolated PostgreSQL integration database behind the build tag and environment flag, the migration `up` and `down`, the same-team and cross-team matrix, the role matrix, the authentication and session cases, the pagination stability test, the idempotent create replay and conflict test, the transaction rollback test, the concurrent update conflict test, the rate-limit test with the injected clock, the graceful shutdown test, the redaction test, the liveness and readiness test, the registry isolation test, and the end-to-end happy and error path tests all pass;
- [ ] When the race detector is clean;
- [ ] When the project documentation reproduces the consistent compact error shape, the pagination cursor, the rate limits, the timeouts, the versioned `up` and `down` migrations, the pinned authentication approach, the bcrypt cost, the `tutorial_session` cookie attributes, the idempotency contract, the explicit non-claims about production readiness, and the explicit scope of the capstone;
- [ ] And when this guide contains no implementation code, signatures, starter snippets, solution snippets, pseudocode, or implementation shell commands.

## 19. Optional Extensions

- A bounded optional read-through cache that may cache project metadata only, never caches authorization, membership, audit, session, or task list results, falls back to PostgreSQL on miss or error, and may be stale under a documented TTL, demonstrating the discipline that a cache is not the source of truth and never caches authorization.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Level Gate 1](../../plan.md#level-1-gate), [Level Gate 2](../../plan.md#level-2-gate), [Level Gate 3](../../plan.md#level-3-gate), [Level Gate 4](../../plan.md#level-4-gate), [Level Gate 5](../../plan.md#level-5-gate), [Level Gate 6](../../plan.md#level-6-gate), [Project 086 — Distributed Task Queue](../../07-advanced-systems/086_distributed_task_queue/README.md#20-prerequisite-based-documentation-guide), [Project 091 — API Gateway Service](../../07-advanced-systems/091_api_gateway_service/README.md#20-prerequisite-based-documentation-guide), [Project 095 — Microservice with Event-Driven Outbox](../../07-advanced-systems/095_microservice_event_driven/README.md#20-prerequisite-based-documentation-guide), [Project 096 — Metrics Prometheus Exporter](../../07-advanced-systems/096_metrics_prometheus_exporter/README.md#20-prerequisite-based-documentation-guide).

Everything introduced by Projects 001–085 and by the linked advanced prerequisites is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **Standards and concept references:** [OWASP Application Security Verification Standard](https://owasp.org/www-project-application-security-verification-standard/).
- **Testing references:** [Go vulnerability scanning](https://go.dev/doc/tutorial/govulncheck).

### Project-specific learning focus

- **Learn now:** layered architecture, configuration and secrets, authentication and tenancy, transactional boundaries, migrations, background work, bounded observability, graceful shutdown, and production-focused tests.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
