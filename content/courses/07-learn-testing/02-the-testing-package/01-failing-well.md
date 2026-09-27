---
title: Failing Well
quiz:
  - question: |
      What happens when this test runs?

      ```go
      func TestImport(t *testing.T) {
          l, err := Import(strings.NewReader("bad csv"))
          if err != nil {
              t.Errorf("Import: %v", err)
          }
          if got := l.Balance("cash"); got != 0 {
              t.Errorf("cash = %d, want 0", got)
          }
      }
      ```

      Assume `Import` returns `nil, err` for bad input.
    options:
      - text: It reports the Import error and stops
      - text: It reports the Import error, then panics with a nil pointer dereference on `l.Balance`
        correct: true
      - text: It passes, because `t.Errorf` doesn't fail the test
      - text: It reports two errors
    explanation: |
      `t.Errorf` marks the test failed but keeps executing. The next line calls
      a method on a nil `*Ledger`, which panics. When the rest of the test
      depends on the result, use `t.Fatalf` so the test stops right there.
  - question: Why is calling `t.Fatal` from a goroutine you started a bug?
    options:
      - text: '`t.Fatal` only stops the goroutine that calls it (via `runtime.Goexit`), not the test function, so the test keeps running in a confused state'
        correct: true
      - text: It crashes the whole test binary immediately
      - text: '`t.Fatal` is not safe to call concurrently at all'
      - text: It silently does nothing
    explanation: |
      `FailNow` works by calling `runtime.Goexit`, which exits the *current*
      goroutine. From another goroutine, the test function carries on. `go vet`
      flags it ("call to (*testing.T).Fatal from a non-test goroutine"). Use
      `t.Error` there, or send the error back over a channel.
---

You met `t.Errorf` and `t.Fatalf` in course 01. Let's look at what they actually do, because choosing well makes the difference between a failure that explains itself and one that makes you reach for a debugger.

## The whole family

`*testing.T` has a small set of reporting methods, and they're all combinations of two ideas: **log a message** and **mark the test as failed**, optionally **stopping** it.

| Method | Logs | Marks failed | Stops the test |
| --- | --- | --- | --- |
| `t.Log`, `t.Logf` | yes | no | no |
| `t.Fail` | no | yes | no |
| `t.Error`, `t.Errorf` | yes | yes | no |
| `t.FailNow` | no | yes | yes |
| `t.Fatal`, `t.Fatalf` | yes | yes | yes |

`t.Errorf` is just `t.Logf` followed by `t.Fail`, and `t.Fatalf` is `t.Logf` followed by `t.FailNow`. `t.Failed()` tells you whether the test has failed so far.

`t.Log` output is only printed if the test fails or you ran `go test -v`, so it's a great place for context that's noise on a pass.

## Error or Fatal?

The rule of thumb: **use `Fatal` when the rest of the test can't make sense, `Error` otherwise.**

```go
func TestImportMarch(t *testing.T) {
	l, err := Import(strings.NewReader(marchCSV))
	if err != nil {
		t.Fatalf("Import: %v", err) // nothing below works without l
	}

	// These are independent checks: report all of them.
	if got := l.Balance("cash"); got != 48210 {
		t.Errorf("cash = %v, want $482.10", got)
	}
	if got := l.Balance("savings"); got != 100000 {
		t.Errorf("savings = %v, want $1,000.00", got)
	}
}
```

If both balances are wrong, `Errorf` shows you both in one run, which is often the clue you need ("everything is off by exactly 100, so I'm parsing the amounts wrong").

## How Fatal stops the test

`FailNow` (and so `Fatal`) calls `runtime.Goexit`, which stops the **current goroutine** after running its deferred calls. Two consequences:

1. Your `defer`s and `t.Cleanup` functions still run, so it's safe to `Fatal` after opening a file.
2. It must be called from the goroutine running the test. From any other goroutine it just ends *that* goroutine and the test function carries on. `go vet` catches this:

```text
$ go vet ./...
./import_test.go:13:3: call to (*testing.T).Fatal from a non-test goroutine
```

Call `t.Error` from goroutines instead (it's safe for concurrent use), or send errors back to the test goroutine over a channel.

## Messages that explain themselves

When a test fails, the message is all you get. Compare:

```text
import_test.go:18: wrong balance
import_test.go:18: Balance("cash") after importing march.csv = $482.00, want $482.10
```

The second says **what was called, with what, what came back and what was expected**. The Go convention is `Func(args) = got, want want`, with *got before want*. Some habits:

- Use `%q` for strings, so empty strings and stray spaces are visible: `""` vs `" "`.
- Use `%v` on types with a `String` method (like `Cents`) so the output reads naturally.
- Don't write `"expected X but got Y"` in one test and `"got Y, want X"` in the next. Pick the Go style and stick to it.
- Don't reach for an assertion library. `if got != want { t.Errorf(...) }` is idiomatic Go, and it lets you write the exact message each check deserves.

## Further reading

- [Go Wiki: Test Comments](https://go.dev/wiki/TestComments) collects the conventions Go's own reviewers use.
