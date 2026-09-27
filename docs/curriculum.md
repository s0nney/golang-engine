# Curriculum: the nine-course core and beyond

The roadmap teaches programming through Go, from first variables to tested concurrent
HTTP services. The boot.dev influence is the learning pattern: short text lessons,
frequent retrieval questions, code written by the learner, immediate feedback, and
recurring project themes. It is not a copy of boot.dev's course content or service.

The sequence is recommended rather than locked. Learners can open any lesson. Course
01 introduces testing and concurrency lightly; courses 06 and 07 revisit them in
depth. Learning to test early supports the algorithm and design courses in between.

## Course map

Counts as of 2026-09-27 (regenerate with `go run ./cmd/validate`): 11 courses, 111
chapters, 548 lessons, 1046 quiz questions, 250 coding exercises. Every chapter has at
least two graded exercises, and every course ends with an "assemble it" exercise that
combines the course's earlier pieces, followed by a link to the next course.

Courses 01–09 are the **core**: the path from zero to a tested HTTP service. Courses
10–11 go **beyond the core** (`track: beyond` in their `course.yaml`), and the roadmap
page lists them under their own heading. They come after it on the roadmap (the last core lesson
points to them) but aren't prerequisites for anything.

| # | Course (slug) | Running project | Chapters | Ends with |
|---|---|---|---|---|
| 01 | [Learn Go](../content/courses/01-learn-go/) (`learn-go`) | Textio, an SMS startup | Introduction, variables, functions, scope, conditionals, loops, arrays and slices, maps, **strings, bytes and runes**, structs, pointers, errors, packages and modules, testing and debugging, concurrency basics, generics basics, **Project: Message Report** (files, `bufio.Scanner`, counting, sorting, `os.Args`/`flag`), final review | Message-report tool (`run(args, out)`) |
| 02 | [Learn OOP in Go](../content/courses/02-learn-oop/) (`learn-oop`) | A dragon/fantasy game | Clean code, types and methods, encapsulation, abstraction, inheritance vs composition, polymorphism (incl. errors as interfaces), generics and OOP, design in practice | Dragon fight: embedding, interfaces, injected dice, custom errors |
| 03 | [Learn Functional Programming in Go](../content/courses/03-learn-functional-programming/) (`learn-functional-programming`) | Doc2Doc, a document converter | FP ideas, first-class functions, pure functions, recursion, function transformations, closures, currying, decorators and middleware (over converters), iterators, sum types | Doc2Doc pipeline: middleware + iterator + sum-type renderer |
| 04 | [DSA 1](../content/courses/04-learn-algorithms/) (`learn-algorithms`) | Clout, an influencer app | Algorithms intro, math, Big O, sorting, exponential time, data structures intro, stacks, queues, linked lists | Moderation desk: list-backed queue, undo stack, stable top-N |
| 05 | [DSA 2](../content/courses/05-learn-data-structures/) (`learn-data-structures`) | A game studio | Binary trees, red-black trees, hashmaps, tries, heaps and priority queues, graphs, BFS and DFS, P vs NP | Matchmaking: hashmap + BFS + heap top-k |
| 06 | [Learn Concurrency in Go](../content/courses/06-learn-concurrency/) (`learn-concurrency`) | Dispatchly, food delivery | Why concurrency, goroutines and WaitGroups, channels in depth, select and time, context, sync primitives, concurrency patterns, errors and coordination, races and deadlocks, **testing concurrent code** (the roadmap's home for `testing/synctest`) | Dispatchly batch engine: bounded workers, first-error cancellation |
| 07 | [Learn Testing and Tooling in Go](../content/courses/07-learn-testing/) (`learn-testing`) | Ledgerly, an accounting library | Why test, the testing package, table tests, test doubles, I/O and files, benchmarks, fuzzing, examples and docs, **the terminal, Git and the go command**, coverage and tooling (vet, staticcheck, `go fix`, **pprof**) | A contract checker that catches buggy CSV importers (the learner writes the tests) |
| 08 | [Learn HTTP Clients in Go](../content/courses/08-learn-http-clients/) (`learn-http-clients`) | Trackr, an issue-tracker CLI | Why HTTP, URLs and DNS, JSON (incl. `encoding/json/v2`), methods, headers and status codes, clients in practice (timeouts, reuse, retries, Retry-After, rate limits), paging, HTTPS and security, testing clients | Trackr client: requests, typed errors, retries, `iter.Seq2` paging |
| 09 | [Learn HTTP Servers in Go](../content/courses/09-learn-http-servers/) (`learn-http-servers`) | Squeak, a microblogging API | Servers, routing, handlers and middleware, JSON APIs, storage, authentication, authorization and webhooks, testing servers, **SQL and databases** (`database/sql`, injection, transactions, migrations), production readiness | Assembled Squeak router; wrap-up of the whole roadmap |
| 10 | [Learn Generics and Advanced Types in Go](../content/courses/10-learn-advanced-types/) (`learn-advanced-types`) | Stash, a generic collections library | The type system, type sets and constraints, type inference, designing generic APIs, generic data structures, interfaces in depth, reflection, generics performance and limits (GC-shape stenciling), capstone | `Repo[K, V]`: typed, validated, LRU-cached repository |
| 11 | [Learn Cryptography in Go](../content/courses/11-learn-cryptography/) (`learn-cryptography`) | Keybox, an encrypted secrets vault | Why cryptography, randomness, hashing, message authentication, passwords and key derivation, symmetric encryption (AES-GCM), public-key crypto (ECDH, ML-KEM, HPKE), digital signatures (Ed25519), certificates and TLS, pitfalls | Keybox: password-derived vault, HPKE sharing, signed bundles; wrap-up of the whole roadmap |

### Where shared topics live

Some topics appear in more than one course on purpose. One course owns the full
treatment; the others recap briefly and link to it.

| Topic | Introduced | Taught in depth | Reused |
|---|---|---|---|
| Testing basics, table tests, TDD | 01 Testing and Debugging | 07 | everywhere after 07 |
| Goroutines, channels, select | 01 Concurrency Basics | 06 | 08, 09 |
| `testing/synctest` | — | 06 Testing Concurrent Code | 07, 08, 09 (back-references) |
| Benchmarks, `b.Loop` | 04 Big O Analysis | 07 Benchmarks | — |
| `pprof` | 06 (pointer only) | 07 Coverage and Tooling | — |
| `go fmt` / `vet` / `fix` | 01 Packages and Modules | 07 Coverage and Tooling | — |
| `os.Args` / `flag` | 01 Message Report project | 07 Terminal, Git and go | 08 (Trackr flags) |
| Middleware / decorators | 03 (over Doc2Doc converters) | 09 Handlers and Middleware | — |
| JSON | 08 JSON (incl. json/v2) | 08 | 09 JSON APIs |
| Errors, `errors.AsType` | 01 Errors | 02 (errors as interfaces) | everywhere |
| Generics | 01 Generics Basics, 02 Generics and OOP | 10 | 03, 04, 05 |
| Reflection, interface internals | 02 Abstraction | 10 | — |
| Hashing, HMAC, password hashing, TLS | 05 Hashmaps, 08 HTTPS, 09 Authentication | 11 | — |

## Lesson and completion model

Each lesson has prose and at least one assessment: a quiz, an exercise, or both. A quiz
passes when every question is answered correctly in one submission. Run executes
code without grading; Submit uses hidden tests or an expected-output comparison.
Passing every assessment present completes the lesson. Prior passes are preserved
when a learner reviews a quiz, edits code, or resets its starter.

The activity heatmap measures daily passed assessments, not reading time or unique
lifetime lessons. A lesson containing both assessments can contribute twice on one
day, and a successful review can contribute again on a later day. See
[activity-and-streaks.md](activity-and-streaks.md).

## In-browser practice and local work

Each browser exercise is one `package main` source file. The server adds a hidden test
file when needed and grades it with the local Go toolchain. Dependencies are limited
to the standard-library curriculum; there is no package-install UI, project filesystem,
interactive terminal, or database service provisioned for students.

Some lessons teach commands, benchmarks, profiling, fuzzing, multiple-package design,
or long-running HTTP servers. Those activities belong in a local project as explained
by the lesson. The five-second browser execution limit cannot host a persistent web
server or perform an extended fuzzing campaign. In-browser exercises practice bounded
pieces of the same concepts; a recurring story does not imply a single persistent
workspace accumulated across all lessons.

## Maintaining the learning progression

Use [AUTHORING.md](../content/AUTHORING.md) for format and exercise rules. A useful
review checks both executable correctness and the learner's path:

1. Introduce terminology and prerequisites before using them in an assessment.
2. Give code-writing practice where implementation matters, including meaningful edge
   cases; a correct reference solution alone cannot demonstrate a strong grader.
3. Keep starter code compilable, but require learner changes to satisfy grading.
4. Distinguish a complete runnable program from a fragment, an intentionally broken
   example, or a program meant to run locally.
5. End a course by combining earlier ideas and pointing to the next course.
6. Preserve URL slugs when adding or renumbering lessons so learner history survives.

The core is a foundation, not a complete production-backend certification. SQL is
covered as a chapter of course 09, but the runner is standard-library-only and has no
database driver, so SQL is assessed with quizzes and with exercises on query building,
row scanning and error mapping against fakes rather than a live database. Distributed
systems, deployment automation and advanced security are beyond these courses. The
Squeak teaching API and the learning platform itself are different applications:
Squeak's authentication lessons do not mean goland-engine has accounts or safe public
code execution.

See [sources.md](sources.md) for reference material and the concurrency coverage audit.
