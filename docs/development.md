# Development

## Prerequisites

- **Go 1.27.1** (the module requires it, and exercises are built with the local toolchain).
- **templ** CLI, matching the module version (`go install github.com/a-h/templ/cmd/templ@v0.3.1001`),
  only if you edit `.templ` files.
- **Node.js + npm**, only if you change the code editor bundle.
- Linux or macOS: the runner uses Unix process groups.

## Run

```sh
go run .                                  # http://127.0.0.1:3000, database ./goland.db
go run . -addr 127.0.0.1:8080 -db /tmp/goland.db
go run . -content content                 # serve lessons from disk instead of the embedded copy
```

Content is embedded at build time, so after editing lessons either restart `go run .` or
use `-content content` (still needs a restart to reseed, since loading happens at startup).

## Build

```sh
templ generate        # regenerate internal/web/views/views_templ.go after editing views.templ
go build -o goland-engine .
```

The binary is self-contained (content, templates and static files are embedded), but it
needs `go` on `PATH` at runtime to run exercises.

## Tests

```sh
go test ./...
go test -race ./...
```

| Package | Covers |
|---|---|
| `internal/runner` | Running programs and hidden tests, compile errors, panics, timeouts, output truncation. |
| `internal/store` | Activity recording, calendar streaks, local-date handling. |
| `internal/web` | Page rendering, quiz progress, exercise Run/Submit/Reset (with and without htmx), the activity dashboard, browser timezone handling. |

## Content validation

```sh
go run ./cmd/validate                                   # every course: structure and quiz rules
go run ./cmd/validate content/courses/06-learn-concurrency
go run ./cmd/validate -exec                             # also run every exercise
```

`-exec` runs every exercise's solution (must pass) and starter (must not pass) through the
real runner, in parallel. It takes a few minutes for the whole roadmap. See
[content/AUTHORING.md](../content/AUTHORING.md) for the lesson format and
[curriculum.md](curriculum.md) for what each course covers.

## Editor bundle

```sh
cd internal/web/editor
npm ci
npm run build     # writes ../static/editor.js and ../static/editor-licenses.txt
```

`node_modules/` is git-ignored; the built `static/editor.js` is committed and embedded.

## Project layout

```
main.go                     entry point: flags, load, seed, serve
cmd/validate/               content validator
content/
  AUTHORING.md              how to write lessons (the content spec)
  audits/                   coverage audits against reference material
  content.go                embeds courses/
  courses/NN-course/NN-chapter/NN-lesson.md
docs/                       this documentation
internal/
  course/                   load, validate and render content
  runner/                   compile and run learner code
  store/                    SQLite: content, progress, activity
  web/                      Fiber app and handlers
    views/                  templ components
    static/                 css, htmx, editor bundle, activity.js
    editor/                 editor bundle source (npm)
```

## Conventions

- Modern Go (1.27): `errors.AsType`, `for range n`, `WaitGroup.Go`, `slices`/`maps`, the
  stdlib `uuid` package.
- Keep the site working without JavaScript; any JS must be an enhancement.
- Keep dependencies minimal. Exercises may only use the standard library.
- Don't put learner state in the content tables; they're dropped on every start.
