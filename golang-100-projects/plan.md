# Practical Plan: 100 Projects for Learning Go by Building

> This document is the source of truth for the entire curriculum. It is also a specification for a documentation model that prepares project folders and README guides only. The learner alone writes implementation and tests. Read it completely before changing any file.

## 1. The Real Goal

The goal is for **the learner to build every project**, not receive 100 finished solutions. Each project should add a small concept, reuse earlier learning, and end with objectively verifiable behavior.

Completing the curriculum with mastery provides a strong foundation for professional Go work, but does not automatically make someone senior. Production experience, reading other people's code, system design, operations, and incident handling remain necessary.

## 2. Current-State Review

Current reality:

- The curriculum contains exactly 100 numbered project folders, ordered from `001` through `100` under seven level folders.
- Every project folder contains a detailed English `README.md` learning guide using the same ordered 19-section template.
- These folders and guides are documentation only. They are not completed software projects and do not prove implementation progress.
- The learner has not implemented these projects through this documentation workflow.
- The curriculum tree contains no `.go` implementation files, learner scaffolds, tests, modules, configuration, migrations, Docker artifacts, CI artifacts, or shell scripts.
- No old Gemini implementation or placeholder is part of the current curriculum state and none should be preserved or described as a current file.
- Historically sensitive titles are already constrained by their guides: Project 077 teaches HTTP/2 multiplexing without Server Push, Project 082 permits only authorized localhost checks, and Project 085 parses offline fixtures without privileged packet capture.

Decision: **folder or README existence is not programming completion**. The documentation model prepares the curriculum only; the learner implements and tests every project independently.

## 3. Roles and Responsibility Boundaries

### Documentation Model Role

The documentation model is responsible for only these tasks:

1. Create a target project folder if it does not exist.
2. Create or improve that target project's `README.md` learning guide.
3. Update only that project's status cell in `plan.md` when appropriate.

It writes **learning maps and specifications only**. It never writes implementation, starter files, tests, or solution material.

### Learner Role

Only the learner:

- Decides the implementation file structure after analyzing requirements.
- Creates `main.go`, packages, modules, and any other project files.
- Writes every function and every line of implementation.
- Writes all unit and integration tests.
- Chooses libraries after understanding why they are needed.
- Creates configuration, migrations, Docker or CI artifacts, and commands when a guide requires them.
- Runs, debugs, measures, and improves the project.

### Repository Boundaries

- Curriculum root is `golang-100-projects/`.
- Each project lives in its level folder as `XX-level_name/NNN_project_name/`.
- Existing folder names and project order must remain exactly aligned with the curriculum table.
- Model write scope is strictly limited to a requested target project folder and its `README.md`, plus that project's status cell in `plan.md`.
- The model must never create or edit implementation code, tests, `go.mod`, `go.sum`, configuration, migrations, Docker or CI files, shell scripts, generated artifacts, or learner notes.
- Any learner-created file belongs to the learner and must never be overwritten, modified, renamed, or deleted.
- If any learner file or personal README addition already exists in a target folder, stop before writing and report the conflict to the user.

## 4. Learning Cycle for Each Project

1. Read the complete `README.md`.
2. Confirm mastery of the "What You Must Know Before Starting" checklist.
3. Study listed topics in official sources or conceptual explanations without copying a project solution.
4. Answer the README design questions on paper or in personal notes.
5. Create an appropriate structure and files independently.
6. Implement requirements one by one.
7. Write every verification case described by the README.
8. Exercise success, failure, boundary, and concurrency cases as applicable.
9. Review self-assessment questions and explain each answer aloud or in writing.
10. Mark the project `✅` only after satisfying its Definition of Completion.

### Project Statuses

- `📘` documentation folder and detailed guide are ready for study. All 100 rows currently have this status.
- `🟦` learner is currently implementing and testing the project.
- `✅` learner has implemented, tested, and understood the project.

Guide readiness is not implementation progress. The learner changes exactly one relevant row from `📘` to `🟦` while working, then to `✅` only after meeting the guide's completion definition.

## 5. Output Prepared by the Documentation Model

Permitted model-created content inside a target project is:

```text
XX-level_name/NNN_project_name/
└── README.md
```

If any other file already exists, stop and report it; do not read it for solutions or modify it. This tree describes only documentation-model output, not files the learner will later create.

Missing empty folders may be prepared separately, but detailed README work must occur in batches of one to five consecutive projects. This limit prevents generic, repetitive guidance. In the current repository, all 100 folders and detailed guides already exist.

## 6. Mandatory README Template for Every Project

Every README must be project-specific and contain these sections in this exact order:

> **Language policy:** This plan and every project README use clear natural English instructional prose. Preserve necessary Go and technical identifiers. Learners may use arbitrary input data, including any language or script, while implementing and testing; guide prose remains English.

1. **Project Name and Number**: exactly match the table and folder name.
2. **Project Idea**: clearly describe the product the learner will build without explaining the solution.
3. **Why This Project Now?**: connect it to earlier learning and state its curriculum contribution.
4. **Prerequisites**: list formal project prerequisites exactly as defined by the prerequisite map.
5. **What You Must Know Before Starting**: provide a precise knowledge checklist, not merely package names.
6. **Explanation of New Concepts**: explain each concept and why it matters in simple language, without code or a ready algorithm.
7. **Learning Objective**: state what the learner should be able to explain or implement afterward.
8. **Functional Requirements**: provide numbered, specific, objectively interpretable requirements.
9. **Inputs and Outputs**: define formats, expected values, and text-only success and error examples.
10. **Rules and Edge Cases**: cover empty values, invalid values, limits, and expected behavior.
11. **Project Constraints**: state allowed or prohibited libraries, security requirements, and scope boundaries.
12. **Design Questions Before Coding**: help the learner choose types, boundaries, and responsibilities; do not answer the questions.
13. **Implementation Milestones**: provide 6-12 ordered milestones describing what the learner delivers, not how to code it.
14. **Verification Cases the Learner Must Write**: list success, failure, boundary, and concurrency cases where relevant, without test code.
15. **Common Mistakes to Watch For**: explain project-specific risks without providing code fixes.
16. **Topics and References for Study**: name official Go pages, packages, and search terms; do not copy reference content.
17. **Self-Assessment Questions**: provide 5-10 understanding questions that do not depend on syntax memorization.
18. **Definition of Completion**: provide a checklist covering behavior, quality, testing, and understanding, not one successful run.
19. **Optional Extensions**: provide at most two extensions, clearly separated from core requirements.

Useful specificity matters more than word count. Advanced projects may divide work into phases within the same README.

## 7. No-Code Policy Inside README Files

This policy is strict:

- Never create or modify any `.go` file.
- Never create `*_test.go`, `go.mod`, `go.sum`, configuration, migration, Docker, CI, SQL, Protobuf, or shell-script files.
- Never place starter code, solution code, function bodies, or implementation snippets in a README.
- Never provide pseudocode that turns step by step into a direct solution.
- Never prescribe function names or signatures as a ready-made design.
- Never add "copy this code" sections or fenced blocks labeled `go`.
- Never generate learner commands or command scripts as project artifacts. A guide may name general verification commands the learner can run after creating their own files.
- Text-only input and output examples are allowed.
- Conceptual explanations, names of packages or APIs to study, and design questions are allowed.
- Verification cases must remain natural-language behavioral specifications; the learner converts them into test code.

If an explanation would reveal the solution, turn it into a design question or state only the required outcome.

## 8. Documentation Quality Gate

A project guide qualifies for `📘` only when:

- Folder name exactly matches the table.
- Only `README.md` was created or edited inside the target project by the documentation model.
- README contains all 19 sections in the required order.
- Formal prerequisites exactly match the prerequisite map; review-only references are not presented as prerequisites.
- README explains what must be learned instead of listing technology names.
- Functional requirements and edge cases are objectively verifiable.
- Milestones progress logically from the smallest deliverable to project completion.
- README contains no Go code, pseudocode, solution snippet, hidden answer, or ready-made function signature.
- Content is specific to the project rather than a reusable generic description.
- Instructional prose is English.
- No learner file or file outside the explicitly requested scope was changed.

All 100 current guides have passed documentation preparation and therefore use `📘`. This means **curriculum guide ready** only. Learners use `🟦` and `✅` for their own implementation progress.

## 9. Strict Documentation Model Protocol

When asked to document a project batch, execute these steps exactly:

1. Read this entire document.
2. Accept only one to five consecutive projects in one request.
3. Match each number to its exact folder, curriculum row, and formal prerequisites.
4. Inspect target folder names and file lists before writing.
5. If any learner-created file or personal README content exists, stop immediately, make no changes to that project, and report exact paths.
6. Create a missing target folder only when requested; create nothing inside it except `README.md`.
7. Read only documentation needed to align curriculum guidance. Never inspect learner code to derive or copy a solution.
8. Write a standalone, detailed English README using all 19 sections and the relevant curriculum row.
9. Enforce the no-code policy by removing every solution, pseudocode step, implementation snippet, or ready-made signature.
10. Verify the documentation quality gate separately for each target project.
11. Update only that project's status cell in `plan.md`: guide preparation may set it to `📘`; never set learner progress to `🟦` or `✅` on the learner's behalf.
12. Report exact folders and README files prepared, confirm no implementation or learner file changed, then stop.
13. Do not begin another batch without a new request.

### Explicit Prohibitions

- Do not implement any project or "help" by writing part of its code.
- Do not create starter files, even empty ones.
- Do not create ready-made tests; describe verification behavior only.
- Do not touch any learner implementation or personal file.
- Do not delete or rename existing files or folders.
- Do not process more than five README files per batch.
- Do not reuse a generic README with only project names changed.
- Do not treat README preparation as programming completion.
- Do not overwrite learner work. Stop and report instead.

## 10. Prepared Prompts for the Documentation Model

### Prepare the First Five Projects

```text
Read plan.md completely, then prepare documentation only for Projects 001 through 005.
Create a missing target folder if necessary, and create or improve README.md using the ordered 19 sections and documentation quality gate.
Write only inside each target project folder and its README.md, plus its status cell in plan.md.
If any learner files or personal README additions exist, stop and report them without overwriting anything.
Do not create, modify, or delete implementation, tests, go.mod, configuration, migrations, Docker/CI files, scripts, or any learner work. I will write every line myself.
Do not include starter code, solutions, pseudocode, implementation snippets, or ready-made function signatures.
Describe verification cases in natural language only. Keep guide prose English, set prepared guide rows to 📘, report exact changes, and stop.
```

### Prepare the Next Batch

```text
Read plan.md completely, then prepare documentation only for Projects 006 through 010 under the same rules.
Touch only target project folders and their README.md files, plus corresponding status cells in plan.md.
Stop and report if learner files or personal README additions exist. Never write code or overwrite learner work.
Verify that each guide is detailed, project-specific, English, and compliant with the 19-section template and no-code policy. Report exact changes and stop.
```

### Request More Explanation Without Code

```text
Explain the concepts required by Project 001 in more theoretical detail using its README.
Use English instructional prose. Do not write code or pseudocode, modify files, provide solution snippets, or give me a ready-made design.
Ask short questions to confirm my understanding before I begin implementing it myself.
```

---

# 11. The 100-Project Curriculum

The "Required Verification" column defines behavior the learner must test independently. The documentation model transfers these cases to README guides in natural language and never writes test code. Every table row currently uses `📘` because all 100 detailed guides are ready; the learner changes one row to `🟦` while implementing and to `✅` only after completion.

### Prerequisite Map

Do not infer dependencies. README files must state these formal prerequisites exactly. A project may discuss nearby work as review or context, but that does not make it a formal prerequisite.

- `001`: nothing beyond installing Go and knowing how to use a terminal.
- `002`: Project `001`. `003`: Project `002`. `004`: Project `003`. `005`: Project `004`. `006`: Project `005`. `007`: Project `006`. `008`: Project `007`. `009`: Project `008`. `010`: Project `009`. `011`: Project `010`. `012`: Project `011`. `013`: Project `012`. `014`: Project `013`. `015`: Project `014`. `016`: Project `015`. `017`: Project `016`. `018`: Project `017`. `019`: Project `018`. `020`: Project `019`. `021`: Project `020`.
- `022`: Projects `021`, `014`, and `017`. Project `022` validates contacts and persists JSON, so both earlier disciplines are formal prerequisites.
- `023`: Project `022`. `024`: Project `023`. `025`: Project `024`. `026`: Project `025`. `027`: Project `026`. `028`: Project `027`. `029`: Project `028`.
- `030`: Projects `029` and `017`. Project `030` uses the safe temporary-file publication discipline.
- `031`: Projects `015` and `027`. Projects `032–045`: immediately preceding project and `031`.
- `046`: Projects `014` and `017`.
- `047`: Project `046`.
- `048`: Projects `047` and `046`.
- `049`: Projects `048` and `046`.
- `050`: Projects `049` and `046`.
- `051`: Projects `050`, `046`, and `041`.
- `052`: Projects `051` and `046`.
- `053`: Projects `052` and `046`.
- `054`: Projects `053` and `046`.
- `055`: Projects `054` and `046`.
- `056`: Projects `055`, `046`, and `034`.
- `057`: Projects `056`, `046`, `034`, and `036`. Project `036` is formal because Project `057` directly reuses the token-bucket model.
- `058`: Projects `057`, `047`, and `046`. Project `047` is formal because Project `058` documents and contract-tests its Notes API.
- `059`: Projects `058`, `050`, and `046`. Project `050` is formal because Project `059` directly reuses bcrypt authentication discipline.
- `060`: Projects `059`, `046`, and `041`.
- `061`: Project `047`. Projects `062–070`: immediately preceding project and `061`; add `041` when using context.
- `071`: Projects `041` and `060`. Projects `072–085`: immediately preceding project and `071`; HTTP/RPC projects also require `060`.
- `086`: Projects `034`, `041`, and `065`. `087`: Project `043`. `088`: Projects `017`, `025`, and `087`. `089`: Projects `034`, `041`, and `044`.
- `090`: Projects `031`, `041`, and `043`. `091`: Projects `050`, `057`, `060`, and `078`. `092`: Projects `011` and `041`.
- `093`: Projects `006` and `028`. `094`: Projects `041`, `065`, and `066`. `095`: Projects `064`, `066`, and `086`.
- `096`: Projects `046` and `060`. `097`: Projects `025` and `030`. `098`: Projects `037`, `041`, and `095`.
- `099`: Projects `025`, `034`, and `087`. `100`: completion of Level Gates 1-6 and Projects `086`, `091`, `095`, and `096`. Project `099` is review-only immediate-catalog-predecessor context, not a formal prerequisite.

## Level 1: Language and CLI Foundations — Projects 001 to 015

| Status | Project | Required Outcome | Core Concepts | Required Verification |
|---|---|---|---|---|
| ✅ | `001_hello_cli` | Read a name and age, then print years remaining until age 100 | `package`, `main`, variables, `fmt`, functions | Ordinary name, empty name, age 100, and invalid age |
| ✅ | `002_calculator` | Four-operation calculator with a text interface | Functions, `switch`, numbers, errors | Every operation, division by zero, and unknown operator |
| ✅ | `003_unit_converter` | Convert temperature, length, and weight in both directions | Constants, decimal precision, logic separation | Conversion table with tolerance and valid negative inputs |
| ✅ | `004_number_guessing` | Guessing game with limited attempts and higher/lower hints | Loops, conditions, injectable randomness | Win, exhausted attempts, and nonnumeric input |
| 📘 | `005_simple_quiz` | Struct-based quiz that calculates score and percentage | `struct`, `slice`, iteration | All correct, all incorrect, question without choices, and percentage calculation |
| 📘 | `006_string_reverser` | Reverse Unicode and check palindromes after simple normalization | `string`, `[]rune`, `strings` | Arabic and emoji input, empty text, spaces, and case differences |
| 📘 | `007_bmi_calculator` | Calculate and classify BMI after validation | `float64`, multiple functions, errors | Category boundaries, zero height or weight, and display-only rounding |
| 📘 | `008_rock_paper_scissors` | Play one or more rounds with a clear result | Custom types, constants, `iota`, randomness | Every win/loss/draw combination and invalid input |
| 📘 | `009_system_info_cli` | Display OS, architecture, and selected environment data through flags | `os`, `runtime`, `flag` | Default flags, present/missing variable, and stable formatting |
| 📘 | `010_pass_generator` | Generate a secure password with required length and character classes | `crypto/rand`, options, errors | Length, required classes, impossible options, and no `math/rand` |
| 📘 | `011_interactive_menu` | Extensible menu of actions and selections | Functions as values, maps, I/O separation | Valid choice, missing choice, exit, and injected reader/writer |
| 📘 | `012_bill_splitter` | Split bill and tip using a declared rounding policy | Structs, methods, simple monetary arithmetic | One/multiple people, zero tip, and invalid people count |
| 📘 | `013_time_world_clock` | Convert a fixed time between time zones | `time`, `LoadLocation`, format/parse | Two valid zones, unknown zone, and date-boundary change |
| 📘 | `014_input_validator` | Independent validators for email, phone, and date | Moderate regex use, parsing, error aggregation | Tables of valid/invalid values, Unicode, and impossible date |
| 📘 | `015_cli_counter` | Cancellable counter with injectable interval | `time.Ticker`, channels, `context` | Correct count, cancellation, and no ticker leak or real-time wait |

### Level 1 Gate

Do not start 016 until you can explain the difference between `string` and `[]rune`, why functions return `error`, and how to separate logic from user input. Projects 001-015 must each pass their learner-written tests.

## Level 2: Data, Files, and Algorithms — Projects 016 to 030

| Status | Project | Required Outcome | Core Concepts | Required Verification |
|---|---|---|---|---|
| 📘 | `016_todo_cli` | In-memory task CRUD using subcommands | Structs, IDs, packages, table tests | Add/list/complete/delete, missing ID, and stable ordering |
| 📘 | `017_json_todo_persister` | Atomically save and restore Project 016 tasks in JSON | `encoding/json`, files, atomic rename | Round trip, missing/corrupt file, and no partial file |
| 📘 | `018_csv_data_parser` | Read CSV and produce a statistical report | `encoding/csv`, `io.Reader`, parsing | Header, malformed row, quotes, empty file, and correct totals |
| 📘 | `019_word_frequency_counter` | Report most frequent words in deterministic order | `bufio.Scanner`, maps, sorting | Punctuation, case differences, tied counts, and large text |
| 📘 | `020_file_organizer` | Plan organization by extension with `--dry-run`, then execute safely | `filepath`, `WalkDir`, file modes | Dry run changes nothing, name collision, and unknown extension |
| 📘 | `021_log_file_analyzer` | Stream-parse logs and extract counts and errors | Scanner, parsing, structs, reports | Multiple levels, malformed line, long line, and report ordering |
| 📘 | `022_contact_book` | Contact book with search and JSON persistence | Small interfaces, validation, persistence | Reject duplicate ID, case-insensitive search, and round trip |
| 📘 | `023_mark_to_html_converter` | Limited parser for headings, lists, and emphasis with escaping | State machine, `strings.Builder`, HTML escaping | Supported elements, plain text, malicious HTML, and unclosed syntax |
| 📘 | `024_directory_tree_printer` | Print a sorted directory tree with depth limit | Recursion, `io/fs`, `filepath` | Ordering, zero depth, empty directory, and symlink without loop |
| 📘 | `025_file_duplicate_finder` | Group identical files by size, then SHA-256 | Hashing, streaming I/O, maps | Same/different files, empty file, and no whole-file memory load |
| 📘 | `026_matrix_operations` | Add, multiply, and transpose matrices | Two-dimensional slices, errors, table tests | Valid/invalid dimensions, empty matrix, and no input mutation |
| 📘 | `027_custom_stack_queue` | Generic stack and queue safe in empty state | Generics, zero value, API design | Multiple types, LIFO/FIFO order, empty state, and usable zero value |
| 📘 | `028_binary_search_tree` | Generic BST supporting insert, search, and in-order traversal | Recursion, comparator, invariants | Insert/search, duplicate policy, empty tree, and traversal order |
| 📘 | `029_linked_list_impl` | Singly linked list with insert, delete, and find | Pointers, nodes, edge cases | Head/middle/tail, empty list, and missing value |
| 📘 | `030_file_encryptor_decryptor` | Encrypt files with AES-GCM and random nonce | `crypto/aes`, GCM, streaming boundaries, errors | Round trip, wrong key, tampering, and no nonce reuse |

### Level 2 Gate

Write an appropriate benchmark for Project 019 or 025 and run every level test. Explain `io.Reader`, value versus pointer semantics, and why encryption without authentication is insufficient.

## Level 3: Concurrency and Cancellation — Projects 031 to 045

| Status | Project | Required Outcome | Core Concepts | Required Verification |
|---|---|---|---|---|
| 📘 | `031_concurrent_timer` | Run multiple timers and collect their results | Goroutines, channels, `WaitGroup` | Every result exactly once, deterministic display order, and injected timing |
| 📘 | `032_ping_pong_channels` | Exchange ping/pong a fixed number of times, then close cleanly | Unbuffered channels, ownership, close | Counts 0 and N, no send after close, and no goroutine leak |
| 📘 | `033_concurrent_url_checker` | Check multiple URLs with concurrency limit and timeout | HTTP client, semaphore, context | `httptest`, success/failure/timeout, and actual concurrency maximum |
| 📘 | `034_worker_pool_basic` | General worker pool processing jobs and returning results/errors | Worker pool, channels, backpressure | Every job once, preserved error, 1 and N workers, and cancellation |
| 📘 | `035_file_downloader_concurrency` | Download ranges from a local server and merge them | Range requests, `WriterAt`, context | Matching content, server without range, failed part, and temporary cleanup |
| 📘 | `036_rate_limiter_token_bucket` | Token bucket with injectable clock | Mutex, ticker/clock, burst | Capacity, refill, excess request, and tests without sleep |
| 📘 | `037_producer_consumer` | Multiple producers/consumers with correct shutdown | Buffered channels, fan-in, ownership | No loss/duplication, zero producers, early cancellation, and race test |
| 📘 | `038_mutex_bank_account` | Concurrent bank account that prevents negative balance | `Mutex`/`RWMutex`, invariants | Parallel deposit/withdrawal, final balance, rejected withdrawal, and `-race` |
| 📘 | `039_concurrent_image_resizer` | Resize images with bounded workers while preserving format | `image`, worker pool, file I/O | Valid dimensions/content, corrupt file, concurrency limit, and resource cleanup |
| 📘 | `040_race_condition_detector` | Documented race example and corrected version | Data race, happens-before, race detector | Safe version passes `-race` and README explains defect |
| 📘 | `041_context_timeout_example` | Operation chain respecting deadline and cancellation | Context propagation, `select`, errors | Success, timeout, parent cancellation, and no context stored in struct |
| 📘 | `042_pub_sub_event_bus` | Internal event bus with subscribe and unsubscribe | Channels, mutex, slow-consumer policy | Multiple subscribers, unsubscribe, slow subscriber, and idempotent close |
| 📘 | `043_thread_safe_cache` | Concurrent TTL cache with stoppable cleanup | RWMutex, clock, lifecycle | Hit/miss, expiry, concurrent access, and leak-free stop |
| 📘 | `044_fan_out_fan_in` | Parallel stream pipeline preserving cancellation | Fan-out/fan-in, merge, context | Complete results, worker error, cancellation, and channels close once |
| 📘 | `045_atomic_operations` | Benchmark mutex and atomic counters | `sync/atomic`, benchmark, trade-offs | Concurrent correctness, `-race`, and documented benchmark without absolute claim |

### Level 3 Gate

All Projects 031-045 pass `go test -race`. Explain who owns closing each channel, how every goroutine stops, and when to choose a mutex instead of a channel.

## Level 4: HTTP and REST — Projects 046 to 060

| Status | Project | Required Outcome | Core Concepts | Required Verification |
|---|---|---|---|---|
| 📘 | `046_basic_http_server` | `net/http` server with health endpoint and HTML page | Handler, server timeouts, templates | `httptest` for both, disallowed method, and content type |
| 📘 | `047_rest_api_crud` | In-memory CRUD for a notes resource | REST, status codes, JSON, mutex | Every endpoint, 404/400, stable ID, and concurrent requests |
| 📘 | `048_custom_router_middleware` | Small router and middleware chain | `ServeMux`/patterns, middleware, context | Middleware order, route parameters, 404/405, and request ID |
| 📘 | `049_json_api_response_formatter` | Consistent envelope and error mapping | JSON encoding, typed errors, logging boundaries | Success/error, header before body, and no internal error leak |
| 📘 | `050_jwt_auth_server` | Login and protected route using a short-lived signed token | Password hashing, JWT claims, auth middleware | Valid/invalid login, signature/expiry/claims, and environment secret |
| 📘 | `051_file_upload_server` | Stream file upload with size and type limits | Multipart, `MaxBytesReader`, safe names | Valid, oversized, rejected type, path traversal, and partial-file cleanup |
| 📘 | `052_static_file_server` | Serve static assets with caching and custom 404 | `FileServer`, embed, headers | Present/missing file, no directory listing, and ETag or Cache-Control |
| 📘 | `053_url_shortener_api` | Create Base62 code and redirect through a store interface | Encoding, interfaces, collision handling | Create/redirect, invalid URL, injected collision, and 404 |
| 📘 | `054_gin_framework_crud` | Project 047 contract implemented with Gin | Framework binding, middleware, testing | Same contract and validation cases as 047 |
| 📘 | `055_fiber_framework_crud` | Project 047 contract implemented with Fiber, then compared | Framework lifecycle, adapters, trade-offs | Same contract and honest comparison of size, testing, and API |
| 📘 | `056_cors_header_middleware` | CORS allowlist and security headers | Preflight, origins, headers | Allowed/rejected origin, OPTIONS, credentials, and `Vary: Origin` |
| 📘 | `057_rate_limited_api` | Per-client rate limiting with state cleanup | Middleware, client identity, token bucket | 429, burst, client isolation, proxy trust policy, and memory cleanup |
| 📘 | `058_swagger_api_docs` | Small OpenAPI contract matching a real API | OpenAPI, schema, contract testing | Valid file, matching endpoints/schemas, and correct examples |
| 📘 | `059_session_cookie_auth` | Login/logout using server-side sessions and secure cookie | Cookies, CSRF basics, session rotation | Flags, expiry, logout, rotation, and missing CSRF where required |
| 📘 | `060_graceful_shutdown_web` | Service combining API, timeouts, and orderly shutdown | Signals, shutdown, dependency cleanup | In-flight request finishes within deadline, reject new work, and close once |

### Level 4 Gate

Choose one API and run its complete contract tests. Explain authentication versus authorization, when to return 400/401/403/404/409, and why server timeouts are bounded.

## Level 5: Databases and Caching — Projects 061 to 070

| Status | Project | Required Outcome | Core Concepts | Required Verification |
|---|---|---|---|---|
| 📘 | `061_sqlite_crud` | Real CRUD with `database/sql` and temporary SQLite | SQL, scan, context, constraints | Create/read/update/delete, duplicate, not found, and temporary database |
| 📘 | `062_postgres_sqlx` | PostgreSQL repository using sqlx | Named queries, struct scan, integration tests | Unit tests with fake store and integration tests for CRUD |
| 📘 | `063_gorm_orm_basics` | Users/projects with explicit GORM relationships | ORM, relations, preload, N+1 | Create relationships, constraints, related loading, and documented query count |
| 📘 | `064_database_migrations` | Up/down tool with version table | Migrations, transactions, idempotency | Fresh up, down, mid-migration failure, and duplicate-version prevention |
| 📘 | `065_redis_caching_layer` | Cache-aside around repository with TTL | Redis, cache miss, invalidation, serialization | Hit/miss, update invalidation, and Redis-unavailable fallback |
| 📘 | `066_db_transaction_manager` | Atomic transfer between accounts | Transactions, isolation, rollback, idempotency | Success, insufficient funds, mid-operation failure, and repeated request ID |
| 📘 | `067_mongodb_nosql_crud` | Document CRUD with unique index and pagination | BSON, context, indexes, ObjectID | CRUD, duplicate, cursor pagination, and context timeout |
| 📘 | `068_connection_pooling` | Controlled pool-configuration experiment with statistics | `SetMaxOpenConns`, DBStats, small load test | Limit never exceeded, WaitCount visible, and rows always closed |
| 📘 | `069_search_text_indexing` | Simple text search with ranking and pagination | Index, simple ranking, query normalization | Multiple words, no results, stable order, and pages without duplicates |
| 📘 | `070_read_write_splitting` | Router directing reads/writes with declared consistency policy | Primary/replica, interfaces, read-after-write | Correct routing, replica failure, and sticky read after write |

### Level 5 Gate

Unit tests must pass without Docker or external services. Integration tests are documented, gated, and run only when required services are available. Explain transaction isolation, why rows must be closed, and cache invalidation difficulty.

## Level 6: Networking and Protocols — Projects 071 to 085

| Status | Project | Required Outcome | Core Concepts | Required Verification |
|---|---|---|---|---|
| 📘 | `071_tcp_echo_server` | TCP echo with line protocol and explicit limits | `net.Conn`, deadlines, framing | Multiple messages/clients, oversized message, timeout, and client close |
| 📘 | `072_tcp_chat_room` | Local chat room with names and broadcast | Hub loop, connections, backpressure | Join/leave, broadcast, duplicate name, slow client, and `-race` |
| 📘 | `073_udp_broadcast_server` | Local discovery using UDP on loopback | Datagrams, packet loss, deadlines | Encode/decode, timeout, malformed packet, and loopback-only test |
| 📘 | `074_websocket_live_chat` | WebSocket chat with heartbeat and message limit | Upgrade, read/write pumps, ping/pong | Two connections, broadcast, oversized message, disconnect, and no leak |
| 📘 | `075_grpc_user_service` | Unary gRPC user service | Protobuf, status codes, interceptors | Bufconn without port, basic CRUD, validation, and NotFound |
| 📘 | `076_grpc_streaming_api` | Bidirectional stream processing events | Streaming, flow control, context | Multiple messages, client cancel, server error, and clean EOF |
| 📘 | `077_http2_push_server` | **Despite legacy name:** HTTP/2 TLS server demonstrating multiplexing, without Server Push | TLS, HTTP/2 configuration, protocol negotiation | Client negotiates `h2`, parallel requests, and explicit fallback/TLS failure |
| 📘 | `078_reverse_proxy` | Reverse proxy with allowlist, timeout, and header policy | `httputil.ReverseProxy`, director/rewrite, errors | Routing, backend down, hop-by-hop headers, and no arbitrary target |
| 📘 | `079_load_balancer_round_robin` | Round-robin balancer with health state | Reverse proxy, atomics/mutex, health checks | Rotation, unhealthy backend, all down, and concurrent requests |
| 📘 | `080_dns_lookup_tool` | Lookup selected record types through injectable resolver | `net.Resolver`, context, formatting | Fake A/AAAA/CNAME, timeout, NXDOMAIN, and deterministic order |
| 📘 | `081_ssh_remote_executor` | Educational executor for `localhost` or test server with command allowlist | SSH client, host-key verification, context | Correct/wrong key, host-key mismatch, rejected command, timeout, and no public hosts |
| 📘 | `082_simple_port_scanner` | Health checker for explicit **localhost-only** port list | Dial timeout, worker limit, authorization boundary | Local open/closed port, timeout, concurrency limit, and reject nonlocal host |
| 📘 | `083_custom_http_client` | Client with bounded retry and injectable exponential backoff/jitter | RoundTripper, idempotency, context | Retry selected 5xx, no retry for POST without key, cancel, and body close |
| 📘 | `084_tls_ssl_server` | Local TLS service and optional mTLS with test certificates | `crypto/tls`, CA, verification | Trusted/expired/wrong-name certificate, mTLS without client cert, and TLS minimum |
| 📘 | `085_packet_sniffer_basics` | **Despite legacy name:** offline parser for frames in `testdata`, without raw capture | Binary encoding, bounds checks, protocol fields | Valid/truncated/unknown frames, fuzz test, and no elevated privileges |

### Level 6 Gate

Every server starts on a random local port and closes during its tests. Never scan public devices or disable certificate verification. Explain framing, deadlines, and backpressure.

## Level 7: Advanced Systems and Capstone — Projects 086 to 100

| Status | Project | Required Outcome | Core Concepts | Required Verification |
|---|---|---|---|---|
| 📘 | `086_distributed_task_queue` | Queue with two workers, retry, lease, and dead-letter handling | Delivery semantics, idempotency, backoff | Success, retry, worker crash/lease expiry, duplicate, and DLQ |
| 📘 | `087_kv_store_in_memory` | Concurrent TTL key-value store with snapshot | API design, locks, persistence, expiry | Set/get/delete, TTL, concurrent access, and snapshot round trip |
| 📘 | `088_lsm_tree_storage_engine` | Simplified WAL, memtable, SSTable, and compaction | Storage layout, checksums, tombstones | Restart recovery, overwrite/delete, corrupt file, and compaction preserves results |
| 📘 | `089_raft_consensus_impl` | Deterministic in-memory Raft simulation, not production cluster | Terms, elections, log replication, state machine | Election, leader failure, stale term, majority commit, and injected clock tests |
| 📘 | `090_cron_job_scheduler` | Scheduler with simplified schedules and overlap prevention | Priority queue, clock, lifecycle, misfire policy | Ordering, missed time, long job, cancellation, and restart state |
| 📘 | `091_api_gateway_service` | Gateway combining routing, auth, rate limit, and request IDs | Composition, configuration, proxy, observability | Correct route, unauthorized, limited, backend down, and correlation ID |
| 📘 | `092_cli_docker_manager` | Bounded CLI listing/starting/stopping local containers through interface | Docker API, CLI, dependency inversion | Fake client for each command, confirmation before stop, and no deletion |
| 📘 | `093_custom_interpreter_eval` | Lexer/parser/AST/evaluator for arithmetic language with variables | Tokens, recursive descent, AST, errors | Precedence, parentheses, variable, invalid token, and error position |
| 📘 | `094_distributed_lock_redis` | Redis lock using token, TTL, and safe release with limits explained | Compare-and-delete, lease, failure modes | Single owner, nonowner cannot release, expiry, Redis failure, and integration |
| 📘 | `095_microservice_event_driven` | Orders service publishing events through an outbox | Event schemas, outbox, idempotent consumer | Atomic commit, republish, duplicate event, consumer failure, and schema version |
| 📘 | `096_metrics_prometheus_exporter` | Service exposing useful metrics without high cardinality | Counters, histograms, labels, instrumentation | Correct names/values, bounded labels, endpoint, and concurrent updates |
| 📘 | `097_git_cli_mini_clone` | Educational local model for init/hash object/index/commit | Content addressing, compression, binary formats | Known hash, object round trip, index, commit, and temporary repository only |
| 📘 | `098_kafka_message_broker_client` | Producer/consumer with keys, offsets, and retry policy | Partitions, ordering, consumer groups, delivery | Fake unit tests, gated integration, duplicate handling, and graceful close |
| 📘 | `099_distributed_file_storage` | Local storage nodes splitting files into replicated chunks | Hashing, replication, metadata, repair | Put/get, node loss, corrupt chunk, deduplication, and byte-identical rebuild |
| 📘 | `100_production_ready_saas_backend` | Bounded team/project/task JSON API; aspirational name, not a production-readiness claim | `net/http`, PostgreSQL source of truth, opaque server-side cookie sessions, tenant isolation, idempotency, stable cursor pagination, observability, audit events, graceful shutdown | Unit and gated PostgreSQL integration tests, API behavior, migrations, tenant isolation, idempotency, cursor stability, liveness/readiness, redaction, audit behavior, graceful shutdown, and race checks |

## 12. Project 100 Scope and Boundaries

To keep the capstone concrete, its product is a **bounded team, project, and task management API**:

- Users register, log in, and operate inside teams with explicit `owner`, `admin`, and `member` roles.
- Teams own memberships and projects; projects own tasks with bounded status, priority, assignment, timestamps, and optimistic versions.
- Core stack is `net/http`, `database/sql` through the `pgx` standard-library adapter, PostgreSQL as the source of truth, and opaque server-side cookie sessions.
- Sessions use crypto-random opaque tokens; only token digests are stored server-side. Password handling, cookie attributes, expiry, logout, and password-change revocation follow the Project 100 guide.
- Tenant isolation is enforced in authorization and data access. Cross-team resource access is indistinguishable from not found.
- Bounded `/v1` JSON routes cover authentication, teams, ownership transfer, memberships, projects, and tasks. A compact stable error shape and safe request identifiers apply throughout.
- Task creation is idempotent through a scoped `Idempotency-Key`; task updates use optimistic versions and report stable conflicts.
- List operations use stable opaque cursor pagination and deterministic ordering.
- Versioned `up` and `down` PostgreSQL migrations are core. PostgreSQL integration tests are gated by build tag `postgres_integration` and environment flag `POSTGRES_INTEGRATION=1` and use an isolated database.
- Structured redacted logs, bounded-label Prometheus metrics, `GET /livez`, database-aware `GET /readyz`, audit events, explicit contexts, rate limits with injected clock, and graceful shutdown with a bounded drain window are core.
- Audit events cover membership, role, ownership, denied escalation, login, logout, password change, and team deletion with the transactional behavior defined by the guide.
- Redis is outside core scope. Background jobs are optional extensions only, never a required phase or completion condition.
- Dockerfile, Compose, CI, Kubernetes, cloud deployment, billing, email delivery, social login, file uploads, custom identity providers, multi-region deployment, and a frontend are not required core artifacts.
- Project name is aspirational curriculum wording. Completion is a learning milestone and **not a claim that the result is production-ready**.

Project 100 milestones remain aligned with its README: contract, domain policy, schema and migrations, repository, authentication, memberships, projects and tasks, HTTP behavior, observability and operations, verification, and documentation. Each milestone has its own deliverable and tests before the next begins. Optional background jobs are not part of this sequence.

## 13. Suggested Study Pace

- Small project: one to three sessions.
- Medium project: three to six sessions.
- Advanced project: divide it into milestones; do not compress it into one file or one session.
- After every five projects: rewrite a small part without consulting your prior solution and record what remains unclear.
- After each level: choose one project and review its guide and your own tests instead of adding random features.

Most important rule: **one correct, understood project is better than ten folders that merely look complete.**
