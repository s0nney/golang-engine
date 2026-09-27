---
title: Cleanup and TempDir
quiz:
  - question: |
      In what order are the lines logged?

      ```go
      func TestOrder(t *testing.T) {
          t.Cleanup(func() { t.Log("A") })
          t.Cleanup(func() { t.Log("B") })
          defer t.Log("C")
          t.Log("D")
      }
      ```
    options:
      - text: A, B, C, D
      - text: D, C, B, A
        correct: true
      - text: D, A, B, C
      - text: D, C, A, B
    explanation: |
      The body runs first (D). When the function returns, its deferred calls
      run (C). Then the testing package runs cleanup functions in last-added,
      first-called order: B, then A.
  - question: What's the main advantage of `t.Cleanup` over `defer` in a helper?
    options:
      - text: '`t.Cleanup` runs faster'
      - text: A `defer` in a helper runs when the *helper* returns; `t.Cleanup` runs when the *test* finishes, so the helper can hand back a resource that stays valid
        correct: true
      - text: '`defer` doesn''t run if the test calls `t.Fatal`'
      - text: '`t.Cleanup` also runs after the whole package''s tests finish'
    explanation: |
      That's the killer feature. A `newTestDB(t)` helper can open something,
      register `t.Cleanup(db.Close)` and return it. With `defer`, the
      resource would be closed before the test ever used it.
---

Tests often create things that must be torn down afterwards: temp files, servers, goroutines, global settings. Go gives you two tools that make this painless and hard to get wrong.

## t.Cleanup

`t.Cleanup(f)` registers `f` to run when the test (and all its subtests) finishes, whether it passed, failed or called `t.Fatal`. Cleanups run in **last-in, first-out** order, like `defer`:

```go
package ledgerly

import "testing"

func TestOrder(t *testing.T) {
	t.Cleanup(func() { t.Log("cleanup 1") })
	t.Cleanup(func() { t.Log("cleanup 2") })
	defer t.Log("defer")
	t.Log("body")
}
```

```text
=== RUN   TestOrder
    order_test.go:9: body
    order_test.go:10: defer
    order_test.go:7: cleanup 2
    order_test.go:6: cleanup 1
--- PASS: TestOrder (0.00s)
```

So why not just `defer`? Because `defer` is tied to the *function* it's written in. A setup helper that `defer`s would tear its resource down before returning it. `t.Cleanup` is tied to the *test*:

```go
// newTestServer starts a fake bank API and stops it when the test ends.
func newTestServer(t *testing.T) *BankServer {
	t.Helper()
	srv, err := StartBankServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("starting fake bank: %v", err)
	}
	t.Cleanup(srv.Stop)
	return srv
}

func TestSync(t *testing.T) {
	srv := newTestServer(t) // no teardown to remember
	// ...
}
```

The caller can't forget to clean up, because there's nothing to remember. This is the standard shape for Go test fixtures.

## t.TempDir

Ledgerly imports and exports CSV files, so its tests need somewhere to write. Never write into the package directory or a fixed path like `/tmp/ledgerly`: parallel tests would trample each other, and failures would leave junk behind.

`t.TempDir()` creates a fresh, empty directory that is **deleted automatically** (with a cleanup) when the test finishes. Each call returns a new directory:

```go
func TestExportWritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "march.csv")

	l := sampleLedger(t)
	if err := l.ExportFile(path); err != nil {
		t.Fatalf("ExportFile(%q): %v", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "date,account,amount\n") {
		t.Errorf("export starts with %q, want the CSV header", data[:min(len(data), 40)])
	}
}
```

Under the hood the directory has a name like `/tmp/TestExportWritesFile3765857736/001`, built from the test name, so it's easy to spot if you go looking. If anything in it can't be removed, the test fails (on Windows, that catches files you forgot to close).

## Cleanup and subtests

Cleanups registered on a subtest's `t` run when that subtest ends. Cleanups on the parent run after *all* its subtests finish, including parallel ones. That makes the parent the right place for a shared, expensive fixture:

```go
func TestReports(t *testing.T) {
	l := loadBigLedger(t) // parent registers its cleanup
	t.Run("monthly", func(t *testing.T) { /* uses l */ })
	t.Run("yearly", func(t *testing.T) { /* uses l */ })
}
```

## TestMain, briefly

If a whole *package* needs setup (starting a database container, say), define `func TestMain(m *testing.M)`. It runs instead of the tests; you do setup, call `m.Run()`, then tear down. Most packages never need it. Reach for `t.Cleanup` and helpers first.
