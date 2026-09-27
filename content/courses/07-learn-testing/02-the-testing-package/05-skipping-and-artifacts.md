---
title: Skipping Tests and Saving Artifacts
quiz:
  - question: |
      A test calls `t.Skip("needs Postgres")` halfway through, after one
      `t.Errorf` has already been reported. What's the outcome?
    options:
      - text: The test is reported as skipped and the error is ignored
      - text: The test is reported as failed; skipping doesn't undo an earlier failure
        correct: true
      - text: The test panics
      - text: '`t.Skip` is ignored because the test already failed'
    explanation: |
      `t.Skip` stops the test like `t.FailNow` does, but a test that already
      failed stays failed. Put skip checks at the very top of the test, before
      any real work.
  - question: You run `go test -artifacts ./...` without `-outputdir`. Where does `t.ArtifactDir()` point?
    options:
      - text: A temporary directory that is deleted when the test ends
      - text: A directory under `_artifacts` in the directory you ran `go test` from
        correct: true
      - text: The user's home directory
      - text: It panics without `-outputdir`
    explanation: |
      With `-artifacts`, files are kept under `<outputdir>/_artifacts/...`,
      and `-outputdir` defaults to the directory you ran `go test` from.
      Without `-artifacts`,
      `t.ArtifactDir()` behaves like `t.TempDir()`: it's deleted afterwards.
---

Two more tools for real-world test suites: skipping tests that can't or shouldn't run right now, and keeping files a test produced so you can look at them after it fails.

## t.Skip

`t.Skip(args...)` (and `t.Skipf`) logs a reason and stops the test, marking it **skipped** rather than passed or failed. Use it when a test can't run in this environment:

```go
func TestImportFromPostgres(t *testing.T) {
	dsn := os.Getenv("LEDGERLY_TEST_DSN")
	if dsn == "" {
		t.Skip("LEDGERLY_TEST_DSN not set; skipping database test")
	}
	// ...
}
```

`go test -v` shows it clearly:

```text
=== RUN   TestImportFromPostgres
    db_test.go:12: LEDGERLY_TEST_DSN not set; skipping database test
--- SKIP: TestImportFromPostgres (0.00s)
```

A skipped test is not a passing test. It's easy for skips to pile up until half your suite never runs. Skip for a *reason* the reader can act on, and make sure CI runs the skipped tests somewhere.

## testing.Short

`go test -short` sets a flag that tests can check with `testing.Short()`. It's the standard way to split fast and slow tests:

```go
func TestImportMillionRows(t *testing.T) {
	if testing.Short() {
		t.Skip("slow: imports 1M rows; run without -short")
	}
	// ...
}
```

Developers run `go test -short ./...` constantly while editing; CI runs the whole thing. The flag doesn't skip anything on its own. Tests have to opt in.

## Skipping a flaky test is a last resort

If a test is flaky, `t.Skip("flaky, see issue #123")` stops it blocking everyone. But it also stops it catching anything. Treat it as a tourniquet and fix the cause (the last lesson in this chapter is all about flakiness).

## t.ArtifactDir (Go 1.26)

When Ledgerly's report renderer produces the wrong HTML, you want to *see* the HTML. Writing it to `t.TempDir()` is useless, because that's deleted when the test ends. `t.ArtifactDir()` returns a directory for **test artifacts**, files meant for humans to inspect afterwards:

```go
func TestRenderReport(t *testing.T) {
	html := RenderReport(sampleLedger(t))

	out := filepath.Join(t.ArtifactDir(), "report.html")
	if err := os.WriteFile(out, html, 0o644); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(html, []byte("<h1>March 2026</h1>")) {
		t.Errorf("report is missing its heading; see %s", out)
	}
}
```

What happens to the files depends on how you run the tests:

- **Plain `go test`**: `ArtifactDir` works like `TempDir`. The directory is deleted when the test ends, so your tests don't litter the disk.
- **`go test -artifacts`**: the files are kept, under `_artifacts/` in the directory you ran `go test` from, one subdirectory per test (plus the package's path when you test several packages with `./...`). Add `-outputdir dir` (an existing directory) to put them somewhere else, like a CI job's upload folder.

With `-v` you also get a line telling you where they went:

```text
=== RUN   TestRenderReport
=== ARTIFACTS TestRenderReport /home/you/ledgerly/_artifacts/TestRenderReport/3642752042
```

Each test and subtest gets its own directory, so parallel tests never collide. Add `_artifacts/` to your `.gitignore`.

## Extra output: t.Output and t.Attr (Go 1.25)

Two smaller additions round out the reporting tools:

- `t.Output()` returns an `io.Writer` that writes into the test's log, indented like `t.Log` but without file:line prefixes. Handy for passing to code that wants a writer, like a logger: `slog.New(slog.NewTextHandler(t.Output(), nil))`.
- `t.Attr(key, value)` attaches a key/value attribute to the test, printed as `=== ATTR  TestName key value`. It's meant for CI systems that parse test output, for example to link a test to an issue number.

Both are also on `testing.TB`, so helpers can use them.
