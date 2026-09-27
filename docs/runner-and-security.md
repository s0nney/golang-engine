# Code runner and security

Coding exercises are compiled and executed **on the server** by `internal/runner`.

## How a run works

`Runner.Run(code)`, `Runner.Test(code, tests)` and `Runner.Submit(code, tests, expected)` all
go through `buildAndRun`:

1. **Acquire a slot.** A semaphore channel sized to `runtime.NumCPU()` limits concurrent
   jobs. A request waiting for a slot gives up if its context is cancelled.
2. **Write files** into a fresh `os.MkdirTemp("", "goland-run-")` directory (removed
   afterwards): `main.go`, optionally the hidden `main_test.go`, and
   `go.mod` (`module exercise`, `go 1.27`).
3. **Build** with `go build -o prog .`, or `go test -c -o prog .` for tests, with
   `GOTOOLCHAIN=local GOPROXY=off GOFLAGS=-mod=mod CGO_ENABLED=0 GOWORK=off`. Standard library
   exercises are authored without third-party dependencies and module-proxy downloads
   are disabled. This configuration is not network isolation. Build timeout: **60 s**.
4. **Run** `./prog` (tests with `-test.v -test.count=1`) with a minimal environment
   (`HOME` and `TMPDIR` set to the temp dir, `PATH=/usr/bin:/bin`). Run timeout: **5 s**.
5. Stdout and stderr are combined into a buffer capped at **64 KiB** (`(output truncated)`
   is appended when it overflows). A timeout appends
   `Killed: took longer than 5s. Is there an infinite loop?`.

Processes start in their own process group (`Setpgid`), and on timeout the whole group is
sent `SIGKILL`, including ordinary descendants that remain in that group. Hostile code
can detach from the group; this is not a guarantee that all descendants are contained.
No shell is involved:
the go tool and the program are started with fixed argument lists, and learner code only
reaches them as files.

## Grading

- **`tests`**: Submit builds with the hidden `main_test.go` and passes when `go test` exits 0.
  Learners see the verbose test output, so test failure messages are the feedback.
- **`expected_output`**: Submit runs the program and passes when its output equals the
  expected text, ignoring trailing whitespace on each line and at the end
  (`runner.SameOutput`). On a mismatch the page shows Expected and Your output side by side.

Run is never graded. Only a passing Submit marks an exercise as passed.

## Threat model: read this before deploying

The runner is **not a sandbox**. Learner code runs as the server's OS user and can read
and write any file that user can, open network connections, and use CPU and memory up to
the timeout. The limits above only protect against accidents (infinite loops, huge output,
too many simultaneous runs), not against a hostile user.

That is why the server listens on `127.0.0.1` by default: it's designed to be run by a
learner on their own machine. To expose it to other people you must put the runner inside
real isolation, for example:

- isolate each job in a disposable execution environment (for example a hardened
  container, sandbox, or microVM), with network restrictions and CPU/memory/process limits;
- keep the application database, session state, host credentials and other learners'
  work outside that environment;
- run the application itself as an unprivileged user and apply request rate limits.

Wrapping the entire application in one container does not isolate learner code from
its SQLite database or other jobs. Hidden tests are placed beside submitted code and
run in the same process for test grading. A hostile program can inspect or bypass them;
passing is instructional feedback, not a security boundary.

Other things to add before a public deployment: CSRF protection on the POST routes
(currently only `SameSite=Lax` cookies), `Secure` cookies behind HTTPS, and per-session
rate limits on Run/Submit.

## Platform

The runner file is built only on Unix (`//go:build unix`) because it uses process groups.
The Go toolchain must be on `PATH` when the server starts. The same runner powers
`go run ./cmd/validate -exec`.
