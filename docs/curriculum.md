# The nine-course core

The roadmap teaches programming through Go, from first variables to tested concurrent
HTTP services. The boot.dev influence is the learning pattern: short text lessons,
frequent retrieval questions, code written by the learner, immediate feedback, and
recurring project themes. It is not a copy of boot.dev's course content or service.

The sequence is recommended rather than locked. Learners can open any lesson. Course
01 introduces testing and concurrency lightly; courses 06 and 07 revisit them in
depth. Learning to test early supports the algorithm and design courses in between.

## Course map

| Order | Course and source | What the learner practices | Handoff |
|---|---|---|---|
| 01 | [Learn Go](../content/courses/01-learn-go/) | Textio message-processing examples; variables, functions, scope, decisions, loops, slices, maps, text, structs, pointers, errors, packages, basic tests, concurrency and generics; a message-report project. | Read and write small Go programs before studying design. |
| 02 | [Object-Oriented Programming](../content/courses/02-learn-oop/) | Structs, receivers, method sets, package-level encapsulation, small interfaces, embedding, composition, polymorphism, generics and dependency injection. | Model behavior without imposing a class hierarchy on Go. |
| 03 | [Functional Programming](../content/courses/03-learn-functional-programming/) | Doc2Doc document transformations; functions as values, purity, recursion, closures, currying, decorators, iterators and sum-type representations. | Build transformations and understand callbacks used in later courses. |
| 04 | [Data Structures and Algorithms 1](../content/courses/04-learn-algorithms/) | Cost models, logarithms, Big O, sorting, exponential search, arrays/slices, stacks, queues and linked lists. | Reason about correctness and cost before larger structures. |
| 05 | [Data Structures and Algorithms 2](../content/courses/05-learn-data-structures/) | Game-world and leaderboard examples; binary trees, balancing, hashmaps, tries, heaps, graphs, BFS/DFS, shortest paths and P vs NP. | Choose and explain structures used in applications. |
| 06 | [Concurrency](../content/courses/06-learn-concurrency/) | Dispatchly dispatch work; goroutine lifetimes, channels, select, timers, context, synchronization, worker pools, pipelines, coordination, cancellation, race/leak diagnosis and deterministic tests. | Manage parallel work with explicit ownership and shutdown. |
| 07 | [Testing and Tooling](../content/courses/07-learn-testing/) | Ledgerly accounting-library examples; red/green/refactor, testing APIs, table tests, doubles, I/O, benchmarks, fuzzing, examples, coverage, profiling and Go tooling. | Design observable, testable boundaries before network integrations. |
| 08 | [HTTP Clients](../content/courses/08-learn-http-clients/) | Trackr issue-tracking client; HTTP, URLs, DNS, JSON, methods, headers/status, transport reuse, timeouts, retries, pagination, TLS and isolated client tests. | Understand a client's contract before implementing the server side. |
| 09 | [HTTP Servers](../content/courses/09-learn-http-servers/) | Squeak microblogging API; serving, routing, middleware, JSON, in-memory storage, authentication concepts, authorization, webhooks, handler tests, shutdown and deployment concepts. | Combine Go fundamentals into a small service and identify the next operational topics. |

The metadata and lesson files are the source of truth for counts and ordering. Get
current chapter, lesson, question and exercise totals with:

```sh
go run ./cmd/validate
```

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

The core is a foundation, not a complete production-backend certification. SQL and
database-backed persistence, distributed systems, full deployment automation and
advanced security would require further courses and additional exercise infrastructure.
The Squeak teaching API and the learning platform itself are different applications:
Squeak's authentication lessons do not mean goland-engine has accounts or safe public
code execution.

See [sources.md](sources.md) for reference material and the concurrency coverage audit.
