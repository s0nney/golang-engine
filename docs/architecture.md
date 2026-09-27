# Architecture

goland-engine is a single Go binary. It embeds the course material, the static assets
and the HTML templates, keeps state in one SQLite file, and shells out to the local Go
toolchain to run learners' code. There are no accounts, no external services and no
build step for the frontend beyond `templ generate` (and an optional esbuild bundle for
the code editor).

```
                 ┌─────────────── goland-engine binary ────────────────┐
 browser ──HTTP──▶ Fiber app (internal/web)                            │
   ▲             │   ├─ session middleware (anonymous cookie)          │
   │ HTML        │   ├─ handlers ──▶ templ views (internal/web/views)  │
   │ (htmx       │   ├─ store (internal/store) ──▶ SQLite goland.db    │
   │  fragments) │   └─ runner (internal/runner) ──▶ go build / test ──┼─▶ temp dir, child process
   │             │                                                     │
   └─────────────│ embedded: content/courses/**, web/static/**         │
                 └─────────────────────────────────────────────────────┘
```

## Packages

| Package | Role |
|---|---|
| `main` (`main.go`) | Parses flags, loads courses, opens and seeds the store, creates the runner, starts the web app. |
| `content` | `//go:embed all:courses`: the course tree compiled into the binary. |
| `internal/course` | Loads the course tree from any `fs.FS`, validates it and renders Markdown to HTML (goldmark + GFM). Defines `Course`, `Chapter`, `Lesson`, `Question`, `Option`, `Exercise`. |
| `internal/store` | SQLite access: content tables (rebuilt on every start), progress and activity tables (kept). |
| `internal/runner` | Compiles and runs learner code with `go build` / `go test -c` in a temp dir, with timeouts, an output cap and a concurrency limit. Unix only. |
| `internal/web` | Fiber v3 app: routes, session cookie, handlers, quiz grading, error handling, static files. |
| `internal/web/views` | templ components (`views.templ`, generated into `views_templ.go`). |
| `internal/web/static` | `style.css`, vendored `htmx.min.js`, bundled `editor.js`, `activity.js`, editor licences. |
| `internal/web/editor` | Source of the CodeMirror 6 + vim editor bundle (npm/esbuild; not part of the Go build). |
| `cmd/validate` | Authoring tool: validates content and, with `-exec`, runs every exercise. |

## Startup

`main.run` (`main.go`):

1. **Load content.** `course.Load(fsys, "courses")` walks `courses/NN-slug/NN-slug/NN-slug.md`.
   `fsys` is the embedded `content.FS`, or `os.DirFS(-content)` when authoring from disk.
   Broken lessons and empty chapters are skipped with a warning rather than stopping the
   server; the server only refuses to start if no course is valid.
2. **Open the store.** `store.Open` opens SQLite in WAL mode with foreign keys on, a 5 s
   busy timeout and immediate write transactions, and runs
   `CREATE TABLE IF NOT EXISTS` for the `progress` and `activity` tables.
3. **Seed.** `Store.Seed` drops and recreates the content tables (`courses`, `chapters`,
   `lessons`, `questions`, `options`, `exercises`) and inserts everything in one
   transaction. The database therefore always matches the binary. Progress and activity
   are keyed by slugs, not row IDs, so they survive reseeding.
4. **Runner.** `runner.New` looks up `go` on `PATH`. If it isn't there the site still
   works, but Run/Submit report that code can't be run.
5. **Serve.** `web.New(store, runner).Listen(addr)`; default `127.0.0.1:3000`.

Flags:

| Flag | Default | Meaning |
|---|---|---|
| `-addr` | `127.0.0.1:3000` | Listen address. Keep it on loopback (see [runner-and-security.md](runner-and-security.md)). |
| `-db` | `goland.db` | SQLite file path. |
| `-content` | (embedded) | Serve courses from this directory instead, e.g. `-content content`. |

## Routes

All routes are in `internal/web/web.go`.

| Method | Path | Handler | What it does |
|---|---|---|---|
| GET | `/static/*` | Fiber static | Embedded assets, `Cache-Control: max-age=3600`, compressed. |
| GET | `/` | `home` | Roadmap: activity heatmap and streaks, all eleven courses with per-course progress. |
| POST | `/progress/reset` | `resetProgress` | Deletes this browser's progress and activity, then 303s to `/`. |
| GET | `/courses/:course` | `course` | Course outline with ✓ marks and a "continue" link to the first unfinished lesson. |
| GET | `/courses/:course/:chapter/:lesson` | `lesson` | Lesson page: body, quiz, exercise, prev/next. |
| POST | `/courses/:course/:chapter/:lesson` | `lesson` | Form field `action` selects `quiz`, `run`, `submit` or `reset`. |
| any | anything else | — | 404 page. |

## Request flow: lesson POST

1. `session` middleware reads the `goland_session` cookie (a UUID), or mints a new
   `uuid.NewV4()`; the cookie is refreshed for a year on every request.
2. `lesson` loads the course outline (for prev/next and "lesson N of M"), the lesson, and
   this session's progress on it. On a course's first or last lesson, `neighbourCourses`
   links the previous course's last lesson or the next course's first lesson, so paging
   never dead-ends.
3. Depending on `action`:
   - **quiz**: `grade` reads `q0`, `q1`, … option indexes from the form and checks each
     against the correct option. All correct → `Store.PassQuiz` (also records activity).
     The whole page re-renders with per-question right/wrong marks and explanations.
   - **run**: `Runner.Run` builds and runs the code; output is shown, never graded.
     The code is saved.
   - **submit**: `Runner.Submit` grades with the hidden tests or by comparing output.
     The code is saved with `passed`; a pass records activity.
   - **reset**: code goes back to the starter; saved code is cleared.
4. htmx requests (`HX-Request: true`) for run/submit get only the `Output` fragment
   (swapped into `#output`), plus an out-of-band `#lesson-status` update when the lesson
   just became complete. Reset returns the `Exercise` fragment. Without JavaScript the
   same POSTs return the full page.

Grading happens on the server. Initial quiz HTML does not include correctness flags;
after submission, feedback and explanations are shown. Hidden test source is not
included in lesson HTML, but test output can reveal test names and expected values.
Exercise solutions aren't stored in SQLite (the validator uses them); they still exist
in the repository and embedded course files, so this is not a secret-exam system.

## Error handling

`errorHandler` maps `store.ErrNotFound` and Fiber 404s to the templ `NotFound` page, logs
500s with `slog`, and sends the status text for other errors. Request bodies are capped at
256 KiB by Fiber and submitted code at 64 KiB by the handler.

## Tech stack

Go 1.27.1 · Fiber v3 · templ · htmx 2 (vendored) · SQLite via `modernc.org/sqlite` (pure
Go, no cgo) · goldmark (Markdown, GFM) · goccy/go-yaml (frontmatter) · CodeMirror 6 +
`@replit/codemirror-vim` (bundled with esbuild). The imported `uuid` package generates
session IDs and `time/tzdata` is embedded so timezone names resolve on any host.

## Boundaries

There is no JSON API, authentication service, background worker, websocket, or external
database. Slugs are public navigation identifiers; the session cookie is the bearer
identifier for learner state. Course prerequisites guide learning but are not enforced:
every lesson URL is accessible immediately. The server reads content at startup, not
on each request, even with `-content`.
