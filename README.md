# goland-engine

A minimal, boot.dev-style platform for learning to program through Go.
Text lessons with multiple-choice quizzes and coding exercises: a nine-course core roadmap plus two further courses.
No accounts: progress is tied to an anonymous cookie. JavaScript is optional: htmx (vendored)
makes Run/Submit update in place, and on exercise pages a CodeMirror editor with a toggleable
vim mode replaces the plain textarea. Lessons, quizzes, exercises, progress and the
heatmap work without JavaScript; automatic timezone detection and editor enhancements require it.

See the [documentation index](docs/README.md) for architecture, curriculum, operation,
data storage, authoring, and development instructions.

1. Learn Go
2. Learn Object-Oriented Programming in Go
3. Learn Functional Programming in Go
4. Data Structures and Algorithms 1
5. Data Structures and Algorithms 2
6. Learn Concurrency in Go
7. Learn Testing and Tooling in Go
8. Learn HTTP Clients in Go
9. Learn HTTP Servers in Go

Beyond the core:

10. Learn Generics and Advanced Types in Go
11. Learn Cryptography in Go

Stack: Go (module requirement: 1.27.1), Fiber v3, templ, htmx 2, SQLite
(modernc.org/sqlite, no cgo). The runner currently requires a Unix host.

## Security

Run and Submit compile and execute learners' code **on the server, as the server's user**,
limited only by timeouts, an output cap and a concurrency limit. That's why the server
listens on `127.0.0.1` by default. Public use needs isolated execution separate from the
application's data and credentials; see [runner security](docs/runner-and-security.md).

## Run

```sh
go run .                # http://localhost:3000
go run . -addr 127.0.0.1:8080 -db /tmp/goland.db
```

Course material lives in `content/courses` and is embedded into the binary. On startup it's
loaded and seeded into SQLite. Invalid content is logged and skipped at startup;
run the validator before shipping changes. Generated templates and editor assets are
committed, so a normal Go build does not require templ or Node.js.

## Exercise editor

`internal/web/editor/editor.js` (CodeMirror 6 + `@replit/codemirror-vim`) is bundled into
`internal/web/static/editor.js`, which is committed and embedded. To rebuild it:

```sh
cd internal/web/editor && npm ci && npm run build
```

## Writing lessons

See [content/AUTHORING.md](content/AUTHORING.md). Check your work with:

```sh
go run ./cmd/validate                               # every course
go run ./cmd/validate content/courses/01-learn-go   # one course
go run ./cmd/validate -exec                         # also run every exercise
./goland-engine -content content                    # a built binary serving lessons from disk
```

## Activity and streaks

The roadmap shows a 26-week heatmap, current streak, best streak, and today's activity.
Passing a lesson quiz or coding exercise records one completion; each counts once per
browser per calendar day. Reviews count again on a later day. Runs, failed submissions,
and resetting code do not count. Shades represent 0, 1–2, 3–5, 6–9, and 10+ completions.
A current streak includes consecutive active days ending today or yesterday, giving
learners the rest of today to continue it. Best streak includes all recorded history.

Activity uses the existing anonymous browser cookie and persists in SQLite across
restarts and content reseeding. Browser JavaScript supplies an IANA timezone; without
it, dates use UTC. Past activity keeps the local date recorded at the time, including
when the browser later changes timezone. Clearing browser cookies starts a new identity;
Reset my progress clears that browser's progress, activity, and streaks together.
Activity begins when this feature is installed; old progress timestamps are not
backfilled because they can represent code edits rather than completions.
