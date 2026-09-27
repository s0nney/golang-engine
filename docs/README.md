# goland-engine documentation

goland-engine is a boot.dev-style platform for learning to program through Go: a nine-course
core plus two further courses of short text lessons, each with a quiz and/or a coding exercise that runs on
the server. No accounts; progress lives against an anonymous cookie in SQLite.

| Document | Read it when you want to… |
|---|---|
| [architecture.md](architecture.md) | Understand the packages, startup sequence, routes and request flow. |
| [data-model.md](data-model.md) | Know what's in `goland.db`, what's rebuilt on start and what's kept. |
| [runner-and-security.md](runner-and-security.md) | Understand how learner code is compiled and run, and why the server stays on localhost. |
| [frontend.md](frontend.md) | Work on templates, CSS, htmx behaviour or the CodeMirror/vim editor. |
| [activity-and-streaks.md](activity-and-streaks.md) | Understand the heatmap, streak rules and timezone handling. |
| [curriculum.md](curriculum.md) | See what the eleven courses cover, their running projects and how they hand off. |
| [development.md](development.md) | Build, test, validate content and rebuild the editor bundle. |
| [operations.md](operations.md) | Configure a local instance, back up and restore data, troubleshoot failures, and understand deployment limits. |
| [../content/AUTHORING.md](../content/AUTHORING.md) | Write or edit lessons, quizzes and exercises (the content spec). |
| [sources.md](sources.md) | See the reference material the curriculum draws on. |

Quick start:

```sh
go run .                       # http://127.0.0.1:3000
go run ./cmd/validate -exec    # check every lesson and run every exercise
```

The module currently requires Go 1.27.1 and the runner requires Unix. Node.js and the
templ CLI are only needed to regenerate their respective assets. See
[development.md](development.md) for exact commands.

For learners, the recommended sequence is the nine-course core roadmap, then the two further courses. For contributors,
start with architecture and authoring; for operators, start with operations and runner
security. These documents describe the checked-in implementation, not a hosted service:
there are no accounts, synchronization, billing, or public-execution sandbox.
