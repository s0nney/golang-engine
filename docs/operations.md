# Operating a local instance

This application is intended for a learner's own machine or a trusted development
environment. It executes submitted Go code with the server user's permissions. Read
[runner-and-security.md](runner-and-security.md) before considering shared hosting.

## Requirements and configuration

Use a Unix host with the toolchain required by `go.mod` (currently Go 1.27.1).
The application and the exercise runner both need Go; a compiled server still needs
the `go` executable on `PATH` to run exercises. The Go build cache and temporary
directory must be writable. Disk usage includes the database, Go caches, and concurrent
temporary builds.

Configuration is through command-line flags:

```sh
go run . -addr 127.0.0.1:3000 -db ./goland.db
go run . -content content -db /tmp/goland-authoring.db
```

`-content` names a directory containing `courses/`, not a single course. Relative
database and content paths resolve from the working directory. There is no `.env`
loader or application configuration file. The runner inherits the server environment
for compilation, then gives the learner program a minimal environment; see the runner
document for the exact differences.

Keep a stable hostname and port for a predictable browser origin. Cookies are scoped
to host and path rather than TCP port: two instances on the same hostname can share
the browser's cookie while writing to different databases.

## Startup, updates and shutdown

On startup the loader reports invalid content and skips it. The server exits if no
courses are usable. A successful listen alone does not establish that all nine courses
loaded: validate the source and inspect the startup log's course count.

Every startup replaces the content tables in a transaction. Progress and activity
survive because their keys use course/chapter/lesson slugs. An authoring instance also
needs a restart after edits. Embedded content and static assets require rebuilding.
Treat a slug change as a data migration; renumbering a filename prefix preserves its
slug, while renaming the slug disconnects existing progress from the new lesson.

The server currently has no application-level graceful shutdown hook, maintenance
mode, scheduled cleanup, or dedicated health endpoint. Stop a local instance when no
exercise is running, especially before restoring or replacing its database. Do not
start different content versions concurrently against the same database: each can
replace the content tables.

## Backup and restore

Back up learner state before replacing a database or changing schemas. With the
optional SQLite command-line tool installed, make a consistent online backup:

```sh
sqlite3 goland.db ".backup '/absolute/path/to/goland-backup.db'"
```

For a file-based backup, stop the server first and copy the database together with any
remaining `goland.db-wal` and `goland.db-shm` files as one set. Copying just the main file
while it is live can omit recently committed progress.

To test or restore a SQLite `.backup` file, stop the server, preserve the old files,
and start with the backup's path using `-db`. Using a new filename avoids mixing a
restored database with stale WAL files. Startup will reseed content from the selected
binary or content directory; learner rows remain. Check a known browser's progress.

A database backup does not restore browser identity. The learner also needs their
original `goland_session` cookie. There is no account recovery or export/import UI.

## Browser identity and privacy

The server sets a random UUID cookie with `HttpOnly`, `SameSite=Lax`, path `/`, and a
one-year expiry refreshed on application requests. It is persistent, not a tab-only
session. A second tab shares progress. Private browsing, another browser profile, or
clearing the cookie creates a separate identity. Blocking cookies prevents stable
progress across requests.

SQLite holds submitted source code, pass flags, timestamps, daily activity and the
anonymous identifier. The timezone cookie is readable by JavaScript; the editor's Vim
preference is kept in localStorage. No account details or analytics service are part
of the implementation. Treat the session cookie as a bearer credential: someone with
its value can use that learner's state. Identifiers are not signed or recoverable.

Reset my progress deletes that identity's progress and activity in one transaction.
Reset code only replaces the current exercise's saved source and does not undo passes.
Clearing cookies loses access to old state but does not delete its database rows.
There is no automatic expiry or pruning of abandoned learner rows.

## Troubleshooting

| Symptom | Check |
|---|---|
| A course or lesson is absent | Run `go run ./cmd/validate`; inspect startup warnings and YAML quoting. |
| Edits do not appear | Restart after content edits; rebuild embedded assets and templates as appropriate. Browser assets are cached for an hour, so hard-reload after a frontend rebuild. |
| Run says the toolchain is unavailable | Check `go version` and `PATH` in the environment that starts the server. Restart after fixing them. |
| A build fails for newer APIs | Confirm the runner uses the required local Go version. It disables toolchain auto-download. |
| Code is killed after five seconds | The execution limit is fixed; a server example that blocks forever belongs in a local terminal, not the exercise runner. |
| Progress disappeared | Check hostname, browser profile, cookie presence, selected database path, and whether slugs changed. |
| Heatmap day looks wrong | Check the displayed timezone, browser clock configuration and timezone cookie. Server time determines the timestamp; without detection the fallback is UTC. |
| Saved code reverted | Only Run/Submit save edits; switching pages with unsubmitted edits discards them. Another tab can overwrite the same exercise. |
| SQLite reports a lock | Check concurrent instances, permissions and disk health. Writers wait up to the configured five-second busy timeout; that is not unlimited retry. |
| Server cannot bind its port | Choose an unused loopback address with `-addr`, or stop the process already listening. |

Logs use Go's `slog` default handler. Startup logs include seed count and content/runner
warnings; request failures with status 500 log path and error. There is no bundled
metrics collector, trace exporter, request access log, or operations dashboard.
