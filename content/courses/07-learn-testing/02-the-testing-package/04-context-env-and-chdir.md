---
title: Context, Environment and Working Directory
quiz:
  - question: When is the context returned by `t.Context()` cancelled?
    options:
      - text: Never; you must cancel it yourself
      - text: Just before the test's `t.Cleanup` functions run
        correct: true
      - text: After all tests in the package finish
      - text: As soon as the test calls `t.Error`
    explanation: |
      The testing package cancels it when the test function (and its
      subtests) are done, *before* cleanups run. So a cleanup can wait for
      background goroutines that stop on `ctx.Done()`.
  - question: |
      What happens here?

      ```go
      func TestEuroFormatting(t *testing.T) {
          t.Parallel()
          t.Setenv("LEDGERLY_CURRENCY", "EUR")
          // ...
      }
      ```
    options:
      - text: The variable is set only for this test's goroutine
      - text: It works, but other parallel tests might see EUR
      - text: The test panics, because `t.Setenv` can't be used in parallel tests
        correct: true
      - text: '`t.Setenv` silently does nothing in parallel tests'
    explanation: |
      Environment variables belong to the whole process, so changing one while
      other tests run in parallel would be a race. `t.Setenv` (and `t.Chdir`)
      detect this and panic with "test using t.Setenv, t.Chdir, or
      cryptotest.SetGlobalRandom can not use t.Parallel".
---

Some code reaches out into its surroundings: it takes a `context.Context`, reads an environment variable, or opens files relative to the working directory. Each of these is process-wide state that a test must set up *and put back*. `*testing.T` has a method for each.

## t.Context (Go 1.24)

Ledgerly has a `Syncer` that polls the bank for new transactions until its context is cancelled. Testing it used to look like this:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

Since Go 1.24, the test hands you one:

```go
func TestSyncerStopsWithTest(t *testing.T) {
	s := NewSyncer(fakeBank{}, &Ledger{})
	done := make(chan struct{})
	go func() {
		s.Run(t.Context()) // returns when the context is cancelled
		close(done)
	}()
	t.Cleanup(func() { <-done }) // wait for Run to exit before the test ends
	// ... drive the syncer and check what it imported ...
}
```

`t.Context()` is cancelled **just before** cleanup functions run. That ordering is deliberate: in the cleanup, `Run` has already been told to stop, so waiting on `done` can't hang forever, and no goroutine outlives the test. `go fix` has a `testingcontext` fixer that rewrites the old `WithCancel` pattern for you.

## t.Setenv

Ledgerly reads its default currency from `LEDGERLY_CURRENCY`:

```go
func DefaultCurrency() string {
	if c := os.Getenv("LEDGERLY_CURRENCY"); c != "" {
		return c
	}
	return "USD"
}
```

To test the override, set the variable with `t.Setenv`, which restores the old value (or unsets it) in a cleanup:

```go
func TestDefaultCurrencyFromEnv(t *testing.T) {
	t.Setenv("LEDGERLY_CURRENCY", "EUR")
	if got := DefaultCurrency(); got != "EUR" {
		t.Errorf("DefaultCurrency() = %q, want %q", got, "EUR")
	}
}

func TestDefaultCurrencyFallback(t *testing.T) {
	t.Setenv("LEDGERLY_CURRENCY", "") // make sure the developer's shell doesn't leak in
	if got := DefaultCurrency(); got != "USD" {
		t.Errorf("DefaultCurrency() = %q, want %q", got, "USD")
	}
}
```

The second test shows a subtle point: a test that depends on an env var being *unset* should say so, or it'll fail on the one machine where someone exported it.

## t.Chdir (Go 1.24)

Ledgerly's CLI looks for `ledgerly.toml` in the current directory. `t.Chdir(dir)` changes directory for the duration of the test and changes back afterwards:

```go
func TestFindsConfigInWorkingDir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "ledgerly.toml"), []byte(`currency = "GBP"`), 0o644)
	t.Chdir(dir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Currency != "GBP" {
		t.Errorf("Currency = %q, want %q", cfg.Currency, "GBP")
	}
}
```

## The catch: no parallelism

Environment variables and the working directory belong to the **whole process**. If two parallel tests changed them, each would see the other's values. So `t.Setenv` and `t.Chdir` refuse to run in a test that called `t.Parallel()` (or whose parent did), and panic:

```text
panic: testing: test using t.Setenv, t.Chdir, or cryptotest.SetGlobalRandom can not use t.Parallel
```

## A design hint

Needing `t.Setenv` in lots of tests is a smell. Compare:

```go
func DefaultCurrency() string                           // hidden input: the environment
func DefaultCurrency(getenv func(string) string) string // explicit input
```

The second version can be tested in parallel with a plain map lookup, and production passes `os.Getenv`. Read the environment once, at the edge of your program, and pass plain values inward. You'll see this idea, **inject your dependencies**, all through the test doubles chapter.
