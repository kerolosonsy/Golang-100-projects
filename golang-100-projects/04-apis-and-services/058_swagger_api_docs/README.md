# Project 058 — Swagger API Docs

## 1. Project Name and Number

- Project 058 — `swagger_api_docs`.
- Folder: `04-apis-and-services/058_swagger_api_docs/`.
- README only; the learner writes all source and tests.

## 2. Project Idea

Hand-author an OpenAPI 3.1 document that exactly describes the Notes API built in Project 047, then add a contract test suite that loads the document, validates it with one pinned dependency, and asserts that the live HTTP routes and statuses match the document. The OpenAPI artifact is named exactly `openapi.json` and lives at a documented location in the project. The document declares `openapi: 3.1.0`, `info.title: Notes API`, `info.version: 1.0.0`, and exactly one `servers` entry whose URL is the relative path `/`. The only third-party dependency is `github.com/getkin/kin-openapi` pinned at `v0.145.0`. The project deliberately avoids generated servers, generated clients, UI hosting, authentication schemes, webhooks, callbacks, tags-only speculative surface, and any public server. The OpenAPI document is the source of truth for the documented surface, and the application must agree with it on every route, method, status, request body, and response envelope. Static schema validation proves only that the document is structurally valid; the live tests prove the behaviour of the routes and statuses that the document describes.

## 3. Why This Project Now?

- Projects 046 through 057 produced an HTTP service with growing surface area: CRUD endpoints, a JSON envelope, a rate limiter, CORS.
- None of those projects had a single place where the contract was described.
- Project 058 introduces the discipline of writing the contract, validating the contract, and asserting that the application still matches it.
- Project 059 will add session-cookie auth, and its contract additions belong in the same OpenAPI document.
- Project 060 will add graceful shutdown, which is operational and therefore out of scope for OpenAPI but worth understanding as a boundary.
- By the end of Project 058 the learner can ship a documented API whose contract is enforced by tests.

## 4. Prerequisites

- Required earlier projects: Project 057, Project 047, and Project 046.
- Earlier HTTP, JSON envelope, and middleware projects are useful review but are not formally required.
- The learner must already know the Notes API from Project 047 — its exact routes, methods, status codes, request fields, and response shapes — and must be comfortable reading and writing JSON.

## 5. What You Must Know Before Starting

- The Notes API built in Project 047: the exact routes, methods, status codes, request fields, response fields, and the rules for `405 Method Not Allowed` and `HEAD`.
- The difference between OpenAPI (the specification) and Swagger (the historical name and tooling family). OpenAPI is the spec; Swagger is one family of tools that consumes it.
- The OpenAPI 3.1 document structure at the level of `openapi`, `info`, `servers`, `paths`, `path-item`, `operation`, `parameter`, `requestBody`, `responses`, `components`, `schemas`, `examples`, and the unused `security` block.
- The meaning of `application/json` content type and why responses use it.
- The fact that JSON Schema 2020-12 is the schema language inside OpenAPI 3.1.
- How to load and validate an OpenAPI document from a file under test, and how to use `net/http/httptest` to drive the API for contract tests.
- The pinned module dependency allowed in this project: `github.com/getkin/kin-openapi` at version `v0.145.0`. The learner must pin the version in `go.mod` and import only that one dependency.

## 6. Explanation of New Concepts

### Concepts

- **OpenAPI versus Swagger.** OpenAPI is a specification language for HTTP APIs.
- A document written in OpenAPI describes routes, methods, parameters, request and response bodies, status codes, headers, and authentication schemes in a structured JSON or YAML file.
- Swagger is the historical brand name and the family of tools that work with OpenAPI, including Swagger Editor, Swagger UI, Swagger Codegen, and others.
- The two names are often used interchangeably in casual writing, but in this project the learner must use "OpenAPI" for the document and the spec, and refer to "Swagger tooling" only when discussing specific tools.

- **OpenAPI 3.1.** The 3.1 release aligns OpenAPI with JSON Schema 2020-12.
- The document starts with `openapi: 3.1.0`, then `info`, `servers`, `paths`, and `components`.
- Each path declares parameters and HTTP methods.
- Each operation declares `parameters`, `requestBody`, `responses`, and `operationId`.
- The schemas live under `components/schemas` and are referenced by `$ref`.

- **Path ID parameter.** The Notes API has the path parameter `/notes/{id}`.
- The OpenAPI document declares it under `parameters` for that path with `in: path`, `name: id`, `required: true`, and a schema that declares `type: integer` with `minimum: 1`.
- The runtime's stricter rule — that the id must be a canonical decimal integer with no sign, no leading zeros, and no whitespace — is not captured by the JSON Schema `minimum` alone and remains in the description of the parameter and is exercised by the live tests.

- **Schema pins.** `NoteWriteRequest` declares `title` as a required string and `body` as an optional string.
- The document records, in prose attached to the schema description, that the runtime trims whitespace from `title` and rejects an empty trimmed title, and that the runtime preserves `body` exactly as provided. `additionalProperties` is `false`. `Note` declares exactly the required fields `id`, `title`, `body`, `created_at`, `updated_at`, with `additionalProperties: false`. `id` is `type: integer` with `minimum: 1`. `created_at` and `updated_at` are `type: string` with `format: date-time`. `Error` declares exactly the required fields `code` and `message`, both strings, with `additionalProperties: false`.

- **Responses and content types.** Each operation declares `responses` keyed by status code or `default`.
- Each body-bearing response declares `content` with the `application/json` media type.
- OpenAPI models the media type as `application/json`; the runtime emits the exact header value `application/json; charset=utf-8`.
- The document declares the model as `application/json`, and the live tests assert the runtime header value exactly.
- The `204 No Content` response declares no content map and the runtime emits no `Content-Type` header. `Content-Type` is owned by the content map and is not modelled separately as a response header.
- The statuses pinned in this project are: `listNotes` returns `200`; `createNote` returns `201` plus the `Location` response header and the error statuses `400`, `413`, `415`, `500`; `getNote` returns `200` plus `400`, `404`; `replaceNote` returns `200` plus `400`, `404`, `413`, `415`; `deleteNote` returns `204` with no content plus `400`, `404`.

- **Stable error codes.** Every body-bearing error in the live API returns one of the stable codes `invalid_request`, `invalid_id`, `not_found`, `method_not_allowed`, `unsupported_media_type`, `payload_too_large`, `internal_error`, each with a documented exact message.
- The OpenAPI document declares the codes used by supported operations as `Error` examples for each applicable status; runtime-only `405` tests pin `method_not_allowed` separately.
- The exact mappings are: `invalid_request` → `invalid request`; `invalid_id` → `invalid note id`; `not_found` → `note not found`; `method_not_allowed` → `method not allowed`; `unsupported_media_type` → `unsupported media type`; `payload_too_large` → `payload too large`; `internal_error` → `internal server error`.
- Item-operation invalid-ID responses use `invalid_id`; invalid create or replacement documents use `invalid_request`.

- **Examples.** The document declares at least one valid `Note` example and at least one valid `NoteWriteRequest` example.
- Every `example` declared in the document is validated against its referenced schema by the test suite.

- **Validation dependency.** The pinned `kin-openapi` module loads and validates the document.
- The learner uses it to load `openapi.json`, validate it against the OpenAPI 3.1 meta-schema, validate representative request and response bodies against the schemas declared in the document, and validate every `example` against its referenced schema.

- **Static versus dynamic validation.** Static schema validation proves that the document is structurally valid and that the examples conform to their schemas.
- It cannot prove that the running API actually returns the documented statuses for every input.
- The contract test therefore has two layers: the static layer loads and validates the document; the dynamic layer drives the live API via `httptest` and asserts that each documented route, method, status, and content type matches the document.
- Both layers are required.
- The honest statement in the test comments and in this README is that even the dynamic layer cannot prove every possible runtime behaviour, only the cases the tests exercise.

## 7. Learning Objective

- After finishing this project, the learner can write an OpenAPI 3.1 document by hand, validate it with one pinned dependency, drive the live API to confirm the contract on every documented status, and articulate the boundary between static schema validation and dynamic behaviour tests.
- The learner can also state the difference between OpenAPI and Swagger without conflating them.

## 8. Functional Requirements

1. The project ships a single OpenAPI 3.1 document file named `openapi.json`, located at a documented path in the project tree.
2. The document declares `openapi: 3.1.0`, `info.title: Notes API`, `info.version: 1.0.0`, and exactly one `servers` entry whose URL is `/`. There are no other server entries.
3. The document declares exactly five operations: `GET /notes` with `operationId: listNotes`, `POST /notes` with `operationId: createNote`, `GET /notes/{id}` with `operationId: getNote`, `PUT /notes/{id}` with `operationId: replaceNote`, `DELETE /notes/{id}` with `operationId: deleteNote`. No other path or operation is declared. There is no `security` block, no `webhooks`, no `callbacks`, no Swagger UI, no code generator, no additional public server, and no tags-only speculative surface.
4. `components.schemas` declares exactly three schemas: `NoteWriteRequest`, `Note`, and `Error`. No other schemas are declared. There is no `Problem` schema.
5. `NoteWriteRequest` declares `title` as a required string, `body` as an optional string, and `additionalProperties: false`. The schema description records that the runtime trims whitespace from `title` and rejects an empty trimmed title, and that the runtime preserves `body` exactly as provided.
6. `Note` declares `id` (required integer, `minimum: 1`), `title` (required string), `body` (required string), `created_at` (required string with `format: date-time`), `updated_at` (required string with `format: date-time`), and `additionalProperties: false`.
7. `Error` declares `code` (required string) and `message` (required string), with `additionalProperties: false`.
8. `GET /notes/{id}`, `PUT /notes/{id}`, and `DELETE /notes/{id}` declare the path parameter `id` under `parameters` with `in: path`, `required: true`, `type: integer`, `minimum: 1`. The runtime's stricter canonical-decimal rule is recorded in the parameter description.
9. `listNotes` declares `responses.200` with `application/json` content whose schema is an array of `Note` references. The operation declares at least two array-shaped examples under the response: one non-empty array example containing a valid `Note`, and one empty-array example (`[]`) to pin the empty-list response shape. No singular `Note` example is described as the response example.
10. `createNote` declares a required `requestBody` with `application/json` content whose schema references `NoteWriteRequest` and at least one `NoteWriteRequest` example. `createNote` declares `responses.201` with `application/json` content whose schema references `Note` and at least one `Note` example, plus a `Location` response header whose schema is `type: string`. `createNote` declares `responses.400`, `responses.413`, `responses.415`, `responses.500` with `application/json` content whose schema references `Error` and at least one `Error` example per status using the documented stable codes (`invalid_request`, `payload_too_large`, `unsupported_media_type`, `internal_error`).
11. `getNote` declares `responses.200` with `application/json` content whose schema references `Note` and at least one `Note` example. `getNote` declares `responses.400` with `application/json` content whose schema references `Error` and an `Error` example with code `invalid_id` and the documented exact message `invalid note id`. `getNote` declares `responses.404` with `application/json` content whose schema references `Error` and an `Error` example with code `not_found`.
12. `replaceNote` declares a required `requestBody` with `application/json` content whose schema references `NoteWriteRequest`. `replaceNote` declares `responses.200` with `application/json` content whose schema references `Note` and at least one `Note` example. Its `400` response contains separate examples for `invalid_id` and `invalid_request`, because either the path ID or the replacement document can be invalid. Its `404`, `413`, and `415` responses use `not_found`, `payload_too_large`, and `unsupported_media_type` respectively.
13. `deleteNote` declares `responses.204` with no content. `deleteNote` declares `responses.400` and `responses.404` with `application/json` content whose schema references `Error` and at least one `Error` example per status using the documented codes.
14. The application code does not import a Swagger UI package, a code generator, an authentication scheme library, a `Problem` schema package, or any second OpenAPI-related package. The only third-party dependency is `github.com/getkin/kin-openapi` at version `v0.145.0`.
15. The application exposes an explicit route/status manifest as a documented function or table that lists every route, method, and status the API actually serves. The contract tests read this manifest and compare it bidirectionally against the OpenAPI document's operations and statuses.
16. The contract tests load `openapi.json` from disk, validate it with `kin-openapi`, and assert that every `$ref` resolves, every `operationId` is unique, every declared `example` validates against its schema, every declared `Note` example validates against the `Note` schema, every declared `NoteWriteRequest` example validates against the `NoteWriteRequest` schema, every declared `Error` example validates against the `Error` schema, and the document contains no `Problem`, no `webhooks`, no `callbacks`, no `security`, no second `servers` entry.
17. The contract tests drive the live API via `httptest`, send representative request bodies for each operation, and assert that the response status and body match the document for every documented status of every operation. The tests also assert the `Location` header on `createNote` `201` and the empty body on `deleteNote` `204`.
18. The contract tests separately cover the Project 047 runtime behaviour for unsupported methods on a path (`405 Method Not Allowed` plus the `Allow` response header) and for `HEAD` requests, because these are part of the runtime contract even though they are not OpenAPI operations.

## 9. Inputs and Outputs

### Interface Contract

Inputs: the `openapi.json` file, the `kin-openapi` validator, the running Notes API under `httptest`, the route/status manifest, and representative request and response bodies. Outputs: a contract test report that either passes or fails with a clear message naming the operation, route, status, or schema that does not match. Example textual inputs and expected textual outputs:

- `openapi.json` is loaded, parsed, and validated against the OpenAPI 3.1 meta-schema. Expected: no error from `kin-openapi`.
- Live request: `POST /notes` with body `{"title":"hello","body":"world"}` to the test API. Expected: status `201`, body conforms to the `Note` schema, `Location` header is set to the path of the created note, `Content-Type` is exactly `application/json; charset=utf-8`.
- Live request: `GET /notes/{id}` for an existing note with the canonical id `1`. Expected: status `200`, body conforms to the `Note` schema, `Content-Type` is exactly `application/json; charset=utf-8`.
- Live request: `GET /notes` against an empty collection. Expected: status `200`, body is `[]` (empty JSON array), body validates against an array of `Note` references, `Content-Type` is exactly `application/json; charset=utf-8`.
- Live request: `GET /notes/{id}` for a missing note. Expected: status `404`, body conforms to the `Error` schema with code `not_found` and the documented message, `Content-Type` is exactly `application/json; charset=utf-8`.
- Live request: `GET /notes/{id}` with the id `abc`. Expected: status `400`, body conforms to the `Error` schema with code `invalid_id` and the documented message `invalid note id`.
- Live request: `PUT /notes/{id}` with body `{"title":"new","body":"body"}`. Expected: status `200`, body conforms to the `Note` schema, `Content-Type` is exactly `application/json; charset=utf-8`.
- Live request: `PUT /notes/{id}` with body `{"title":"","body":"x"}`. Expected: status `400`, body conforms to the `Error` schema with code `invalid_request` and the documented message `invalid request`.
- Live request: `POST /notes` with body whose length is exactly 1,048,576 bytes (the boundary) and is otherwise valid. Expected: status `201`; the body is eligible for decoding.
- Live request: `POST /notes` with body whose length is exactly 1,048,577 bytes (one byte over). Expected: status `413`, body conforms to the `Error` schema with code `payload_too_large` and the documented message `payload too large`.
- Live request: `POST /notes` with `Content-Type: text/plain`. Expected: status `415`, body conforms to the `Error` schema with code `unsupported_media_type` and the documented message `unsupported media type`.
- Live request: `DELETE /notes/{id}` for an existing note. Expected: status `204`, empty body, no `Content-Type` header.
- Live request: `DELETE /notes/{id}` for a missing note. Expected: status `404`, body conforms to the `Error` schema with code `not_found`, `Content-Type` is exactly `application/json; charset=utf-8`.
- Schema conformance: every `example` declared in the document validates against its schema. Expected: no validation error.
- Route/status manifest comparison: the operations declared in the document exactly match the operations in the manifest; the statuses declared per operation exactly match the statuses in the manifest for that operation. Expected: no operation in the document is missing from the manifest, no operation in the manifest is missing from the document, no status is missing in either direction.
- Runtime `405` on an item path: `PATCH /notes/{id}` returns `405`, the exact JSON error `method_not_allowed` / `method not allowed`, exact `Allow: DELETE, GET, PUT`, and exact `Content-Type: application/json; charset=utf-8`.
- Runtime `405` on the collection path: an unsupported method on `/notes` returns `405`, the same exact JSON error, exact `Allow: GET, POST`, and the exact JSON content type.
- Runtime `HEAD` on an item path: `HEAD /notes/{id}` returns the corresponding `405` policy and headers, including exact sorted `Allow: DELETE, GET, PUT` and the JSON content type, but the wire body has zero bytes because the method is `HEAD`.
- Runtime `HEAD` on the collection path: `HEAD /notes` returns the corresponding `405` headers, including exact sorted `Allow: GET, POST` and the JSON content type, with zero wire-body bytes.
- Runtime `HEAD` on an unknown path: `HEAD` against an unknown sub-path returns the corresponding `404` headers, including the JSON content type, with zero wire-body bytes.

## 10. Rules and Edge Cases

- The document must not include a `security` section. There is no documented authentication at this point; Project 059 will add it.
- The document must not include a `webhooks` or `callbacks` block.
- The document must not include a second `servers` entry, a public hostname, or any URL other than `/`.
- The document must not declare a `Problem` schema or any schema other than `NoteWriteRequest`, `Note`, and `Error`.
- The document must not include a Swagger UI page, a generated client, a generated server, or a tags-only speculative path.
- A route or status present in the application but absent from the document is a contract failure. A route or status present in the document but absent from the application is also a contract failure.
- Example bodies must conform to their declared schemas. The contract tests assert this.
- The `Content-Type` response header is owned by the content map. It is not modelled separately as a response header.
- The dependency list in `go.mod` contains `github.com/getkin/kin-openapi` at version `v0.145.0` and no other OpenAPI- or Swagger-related package.
- The document is loaded from `openapi.json` on disk, not from a string literal embedded in test code.
- No `go generate` step is required. The document is hand-authored.
- No network access is required at any point. The OpenAPI validator is local; the API under test is local.
- `405` and `HEAD` are runtime behaviours of Project 047 and are not OpenAPI operations. They are tested separately and documented in the route/status manifest as runtime-only entries.

## 11. Project Constraints

- One third-party dependency is allowed: `github.com/getkin/kin-openapi` at version `v0.145.0`. Its version is pinned in `go.mod`.
- No Swagger UI, no code generators, no annotation-based generators.
- No documentation frameworks that auto-generate OpenAPI from code. The document is the source of truth.
- No public hostnames, no cloud dependencies, no Docker.
- Standard library everywhere else.

## 12. Design Questions Before Coding

- Where does `openapi.json` live in the project tree? Beside `main.go`, under `docs/`, or somewhere else? How does the test find it?
- How is the route/status manifest produced? A hand-maintained table, a function that walks the router, or a struct that lists each operation and its statuses? The manifest must list every status the API can return for each operation, including `405` and `HEAD` as runtime-only entries.
- How are the schemas referenced? Inline or under `components/schemas`? The required pattern uses `components/schemas` so that they can be referenced by multiple operations and by examples.
- How is `info.version` declared? It is pinned to `1.0.0` in this project.
- How is `operationId` chosen? Stable and unique. The convention is a verb-noun form.
- How is `Location` declared on `createNote` `201`? As a response header under the response, with a schema of `type: string`.
- How is `204 No Content` declared on `deleteNote`? With no content map. The empty body is part of the contract.
- How does the test fail clearly when the contract is broken? Which file, which operation, which status, which schema? The failure message must point the learner at the problem.

## 13. Implementation Milestones

1. List the exact routes, methods, statuses, request fields, and response fields of the Notes API from Project 047. This list is the contract.
2. Hand-author `openapi.json` with `info`, `servers`, `paths`, and `components/schemas`. Include the `NoteWriteRequest`, `Note`, and `Error` schemas, the path parameter, the `Location` response header on `createNote`, the empty content on `deleteNote`, and at least one example for every body-bearing response.
3. Add `github.com/getkin/kin-openapi` at version `v0.145.0` to `go.mod`. Do not add any other OpenAPI or Swagger dependency.
4. Write the application's route/status manifest as a documented function or table that lists every route, every method, every status, plus the runtime-only `405` and `HEAD` behaviours.
5. Write the static-validation tests together: load `openapi.json`, validate it with `kin-openapi`, assert the schemas parse and every `$ref` resolves, assert the document declares exactly the five operations with the exact `operationId`s and the exact statuses per operation, and assert the document declares no `security`, no `webhooks`, no `callbacks`, no second `servers` entry, and no `Problem` schema.
6. Write a test that compares the document's operations and statuses to the route/status manifest in both directions.
7. Write a test that asserts every declared `example` validates against its schema, and a separate test that asserts every body-bearing response validates against its referenced schema.
8. Write a test that drives the live API under `httptest` and asserts every documented status for every operation, including strict/malformed input, media type, the 1,048,576-byte boundary and an oversized body, canonical/invalid/missing IDs, missing resources, success bodies, the `Location` header on `201`, the empty body on `204`, and the stable error codes.
9. Write a test that separately covers `405 Method Not Allowed` plus the `Allow` header and the `HEAD` behaviour.
10. Run the full test suite and confirm every test passes.
11. Review the verification list and confirm every item is covered before declaring the project complete.

## 14. Verification Cases the Learner Must Write

### Required Cases

Each item is a behavioural specification. The learner writes the corresponding `go test` code.

- Document loads: `openapi.json` is loaded from disk and parsed; no parse error.
- Document validates: `kin-openapi` reports no validation errors against the OpenAPI 3.1 meta-schema.
- Document version and title: `openapi` is `3.1.0`, `info.title` is `Notes API`, `info.version` is `1.0.0`.
- Servers: exactly one `servers` entry whose URL is `/`.
- Operation set: exactly the five operations with the exact `operationId`s `listNotes`, `createNote`, `getNote`, `replaceNote`, `deleteNote`.
- Operation uniqueness: every `operationId` is unique.
- Schema set: exactly `NoteWriteRequest`, `Note`, and `Error`; no `Problem` schema.
- `NoteWriteRequest`: required `title` string, optional `body` string, `additionalProperties: false`.
- `Note`: required `id` integer with `minimum: 1`, required `title` string, required `body` string, required `created_at` string with `format: date-time`, required `updated_at` string with `format: date-time`, `additionalProperties: false`.
- `Error`: required `code` string, required `message` string, `additionalProperties: false`.
- Path parameter: `id` declared as required path parameter with integer type and `minimum: 1` on the routes that use it.
- Statuses per operation: `listNotes` declares exactly `200`; `createNote` declares exactly `201`, `400`, `413`, `415`, `500`; `getNote` declares exactly `200`, `400`, `404`; `replaceNote` declares exactly `200`, `400`, `404`, `413`, `415`; `deleteNote` declares exactly `204`, `400`, `404`.
- `createNote` `201` declares the `Location` response header.
- `deleteNote` `204` declares no content map.
- Examples conform: every `example` in the document validates against its referenced schema.
- `Note` and `NoteWriteRequest` examples conform.
- Stable operation error codes appear with documented messages: `invalid_request` / `invalid request`, `invalid_id` / `invalid note id`, `not_found` / `note not found`, `unsupported_media_type` / `unsupported media type`, `payload_too_large` / `payload too large`, and `internal_error` / `internal server error`. Each appears in an applicable OpenAPI `Error` example. Runtime-only `405` tests separately assert `method_not_allowed` / `method not allowed`.
- Live `POST /notes` happy path: status `201`, body validates against `Note`, `Location` header is set, `Content-Type` is exactly `application/json; charset=utf-8`.
- Live `POST /notes` malformed body: status `400`, body validates against `Error` with code `invalid_request`.
- Live `POST /notes` oversized body: status `413`, body validates against `Error` with code `payload_too_large`.
- Live `POST /notes` 1,048,576-byte boundary: a body whose total request-body size is exactly 1,048,576 bytes and is otherwise valid is eligible for decoding and returns `201`. The boundary succeeds; the test pins `201` and asserts the body validates.
- Live `POST /notes` 1,048,577-byte body: returns `413` with code `payload_too_large`. The boundary fails by one byte.
- Live `POST /notes` unsupported media type: status `415`, body validates against `Error` with code `unsupported_media_type`.
- Live `GET /notes` happy path: status `200`, body is an array of `Note` references and validates, `Content-Type` is exactly `application/json; charset=utf-8`.
- Live `GET /notes` empty collection: status `200`, body is exactly `[]`, `Content-Type` is exactly `application/json; charset=utf-8`. The response example in the OpenAPI document includes this empty-array example.
- Live `GET /notes/{id}` happy path: status `200`, body validates against `Note`.
- Live `GET /notes/{id}` invalid id: status `400`, body validates against `Error` with code `invalid_id`.
- Live `GET /notes/{id}` missing: status `404`, body validates against `Error` with code `not_found`.
- Live `PUT /notes/{id}` happy path: status `200`, body validates against `Note`.
- Live `PUT /notes/{id}` invalid id: status `400`, body validates against `Error` with code `invalid_id`.
- Live `PUT /notes/{id}` missing: status `404`, body validates against `Error` with code `not_found`.
- Live `PUT /notes/{id}` oversized body: status `413`, body validates against `Error` with code `payload_too_large`.
- Live `PUT /notes/{id}` unsupported media type: status `415`, body validates against `Error` with code `unsupported_media_type`.
- Live `DELETE /notes/{id}` happy path: status `204`, empty body, no `Content-Type` header.
- Live `DELETE /notes/{id}` invalid id: status `400`, body validates against `Error` with code `invalid_id`.
- Live `DELETE /notes/{id}` missing: status `404`, body validates against `Error` with code `not_found`.
- Route/status manifest comparison: every operation in the document is in the manifest; every operation in the manifest is in the document; the statuses per operation match in both directions.
- Runtime `405` on an item path: `PATCH /notes/{id}` returns `405`, the exact `method_not_allowed` JSON error, `Allow: DELETE, GET, PUT`, and the exact JSON content type.
- Runtime `405` on the collection path: an unsupported method on `/notes` returns `405`, the exact `method_not_allowed` JSON error, `Allow: GET, POST`, and the exact JSON content type.
- Runtime `HEAD` on a known item path: `HEAD /notes/{id}` returns `405`, zero wire-body bytes, `Allow: DELETE, GET, PUT`, and the same content type that the corresponding error response declares.
- Runtime `HEAD` on the collection path: `HEAD /notes` returns `405`, zero wire-body bytes, `Allow: GET, POST`, and the corresponding error content type.
- Runtime `HEAD` on an unknown path: `HEAD` against an unknown sub-path returns `404`, zero wire-body bytes, and the corresponding error content type.
- Live schema validation: every body-bearing response validates against its referenced schema.
- Honest boundary statement: the test file or the README includes the prose statement that static validation proves document structure only, that live tests prove selected runtime cases, and that even live tests cannot prove all runtime behaviour.

## 15. Common Mistakes to Watch For

- Confusing OpenAPI (the spec) with Swagger (the tools). The README and the test names use the correct name.
- Importing a second OpenAPI or Swagger library. Only `kin-openapi` at `v0.145.0` is allowed.
- Pinning a different version of `kin-openapi`. The version is exactly `v0.145.0`.
- Adding a `security` section before authentication is implemented. Project 059 will add it.
- Adding a `webhooks` or `callbacks` block. Out of scope.
- Adding a `Problem` schema. Out of scope.
- Adding UI hosting. Out of scope.
- Adding generated client or server code. Out of scope.
- Modelling `Content-Type` as a response header. The content map owns it.
- Letting the document drift from the application. The route/status manifest comparison test catches this; it must run on every change.
- Forgetting required fields in the `Note` schema. Required fields are explicit and the tests assert this.
- Exposing internal fields in the public schema. Only public fields appear.
- Putting a public hostname in `servers`. The base URL is the relative path `/`.
- Skipping the live tests because the static validation passes. Static validation cannot prove runtime behaviour.
- Asserting that static validation proves runtime behaviour. It does not. The README and the test comments say so honestly.
- Treating `405` and `HEAD` as OpenAPI operations. They are runtime behaviours of Project 047 and are tested separately.
- Asserting the wire header `Content-Type` is exactly `application/json` without the `; charset=utf-8` suffix. The runtime emits `application/json; charset=utf-8`; the OpenAPI document models `application/json`. Both forms are correct in their context; the live tests assert the exact runtime header value.
- Modelling `Content-Type` on the `204 No Content` response. The runtime emits no `Content-Type` on `204`, and the document must not declare content on `204`.

## 16. Topics and References for Study

- The OpenAPI Initiative specification, version 3.1.1 (or the latest 3.1.x patch). The structure of `openapi`, `info`, `servers`, `paths`, `components`.
- JSON Schema 2020-12, because OpenAPI 3.1 aligns with it.
- The `github.com/getkin/kin-openapi` documentation for loading and validating OpenAPI 3.1 documents in Go.
- The Notes API contract from Project 047, particularly the request and response shapes, the status codes, the `405` and `HEAD` runtime behaviours, the canonical id rule, and the 1,048,576-byte body limit.
- The Fetch Living Standard for HTTP semantics and content types.
- The `kin-openapi` release notes for `v0.145.0` if the learner needs to confirm the API surface.

## 17. Self-Assessment Questions

1. What is the difference between OpenAPI and Swagger? Give one sentence for each.
2. Why is OpenAPI 3.1 aligned with JSON Schema 2020-12? What does this buy the learner?
3. What can static schema validation prove, and what can it not prove? Give one example of each.
4. Why must the application's route/status manifest match the document's operations and statuses exactly? What is the failure mode if they drift?
5. Why is a `security` section absent from this document? When will it be added, and by whom?
6. Why is a Swagger UI not part of this project? Where might it live in a future project?
7. What does it mean for an `example` in the document to "conform" to its schema? How is that checked?
8. Why are `405` and `HEAD` tested separately from the OpenAPI operations? Why is `HEAD` on a known path `405` rather than `200` in Project 047?
9. Why does the document model the media type as `application/json` while the runtime emits `application/json; charset=utf-8`? Are the two forms compatible?

## 18. Definition of Completion

The project is complete when, in addition to the rules above:

- [ ] Every item in the verification list is a passing test that the learner wrote themselves.
- [ ] The tests pass under `go test ./...` from the project folder.
- [ ] The only third-party dependency in `go.mod` is `github.com/getkin/kin-openapi` at version `v0.145.0`.
- [ ] `openapi.json` lives at a documented location, is hand-authored, and is loaded from disk in the tests.
- [ ] The learner can answer every self-assessment question without rereading the README.
- [ ] The README or the test comments include the honest statement that static validation proves document structure only, that live tests prove selected runtime cases, and that even live tests cannot prove all runtime behaviour.

## 19. Optional Extensions

At most two. Pick one only if the core project is already complete and tested. Optional extensions must not add speculative OpenAPI paths or another dependency.

- Add a coverage report per operation that prints, after the contract tests run, which documented statuses were exercised by live tests and which were only verified statically.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisites:** [Project 057 — Rate Limited API](../../04-apis-and-services/057_rate_limited_api/README.md#20-prerequisite-based-documentation-guide), [Project 047 — REST API CRUD](../../04-apis-and-services/047_rest_api_crud/README.md#20-prerequisite-based-documentation-guide), [Project 046 — Basic HTTP Server](../../04-apis-and-services/046_basic_http_server/README.md#20-prerequisite-based-documentation-guide).

Read the linked guides first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`github.com/getkin/kin-openapi`](https://pkg.go.dev/github.com/getkin/kin-openapi).
- **Standards and concept references:** [OpenAPI 3.1 specification](https://spec.openapis.org/oas/v3.1.1.html), [JSON Schema 2020-12](https://json-schema.org/draft/2020-12).

### Project-specific learning focus

- **Learn now:** paths and operations, reusable schemas, request and response examples, error contracts, validation, runtime-versus-documentation parity, and pinned tooling.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
