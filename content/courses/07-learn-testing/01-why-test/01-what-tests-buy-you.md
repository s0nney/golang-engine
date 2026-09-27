---
title: What Tests Buy You
quiz:
  - question: A teammate says "we don't need tests, the code works, I ran it." What is the strongest counter-argument?
    options:
      - text: Tests make the code run faster in production
      - text: Running it once proves it worked *today*; tests keep proving it after every future change, by anyone
        correct: true
      - text: The Go compiler refuses to build packages without tests
      - text: Tests replace the need for code review
    explanation: |
      Manual checking is a one-off. A test suite is a check that runs in
      seconds, forever, every time anyone touches the code. That's what makes
      refactoring and upgrades safe.
  - question: Which of these is *not* something a good test suite gives you?
    options:
      - text: Confidence to refactor
      - text: Executable documentation of how code is meant to be used
      - text: A proof that the program has no bugs
        correct: true
      - text: Faster feedback than deploying and clicking around
    explanation: |
      Tests can show the presence of bugs, never their absence. They check the
      cases you thought of. Fuzzing (later in this course) helps with the cases
      you didn't, but nothing proves a program bug-free.
---

Welcome to the testing course! You already know Go, you can write objects and closures, you've built trees and hashmaps, and you've wrangled goroutines. In course 01 you also met `go test`, `t.Errorf` and table-driven tests. This course goes much deeper, into Go's testing package and the tools around it.

Our running project is **Ledgerly**, a small accounting library. It stores money as integer **cents** (never `float64`, which can't represent `0.10` exactly), records transactions against accounts, works out balances and imports CSV files from the bank. Money code is a perfect place to learn testing: a bug is not just a crash, it's someone's rent going missing.

## So what do tests buy you?

Here's Ledgerly's core type:

```go
package ledgerly

// Cents is an amount of money in cents. $12.34 is Cents(1234).
type Cents int64
```

And here's a test, in `ledger_test.go`, next to the code:

```go
func TestBalanceSumsOneAccount(t *testing.T) {
	var l Ledger
	l.Add(Transaction{Account: "cash", Amount: 5000})
	l.Add(Transaction{Account: "cash", Amount: -1250})
	l.Add(Transaction{Account: "savings", Amount: 999})

	if got, want := l.Balance("cash"), Cents(3750); got != want {
		t.Errorf("Balance(%q) = %d, want %d", "cash", got, want)
	}
}
```

Those dozen lines give you a surprising amount:

- **Correctness, today.** The obvious one: the function does what you meant.
- **Safety, tomorrow.** Six months from now someone rewrites `Balance` to use a cache. They run `go test ./...` and know within a second whether they broke it. Without tests, every change is a gamble and people stop improving the code.
- **Documentation that can't go stale.** A test shows exactly how to call the API and what to expect. Unlike a comment, it fails when it becomes wrong.
- **Better design.** Code that is hard to test is usually hard to use: too many hidden dependencies, too much going on in one function. Writing the test first pushes you towards small, focused functions with clear inputs and outputs. We'll lean on this hard in the test doubles chapter.
- **Fast feedback.** Running a test takes milliseconds. Starting the app, importing a CSV and squinting at the output takes minutes.

## What tests cost

Tests aren't free. They're code, so they need maintaining, and a *bad* test suite can be worse than none:

- **Brittle tests** check *how* the code works instead of *what* it does, so every refactor breaks them even when behaviour is unchanged.
- **Slow tests** get skipped. If `go test` takes ten minutes, people stop running it.
- **Flaky tests** pass and fail at random. Once people learn to hit "retry", real failures hide among the noise.

A big part of this course is about writing tests that are fast, deterministic and focused on behaviour, so they stay an asset rather than a chore.

## Go makes it cheap

Go's attitude is that testing is part of the language toolchain, not an add-on:

- `go test` is built in. No framework to pick, no config file.
- Tests are plain Go. There's no assertion DSL to learn: you compare values with `==` and report with `t.Errorf`.
- The same tool runs benchmarks, fuzz tests, runnable examples and coverage, all of which you'll meet in this course.

## The recap

A quick reminder of the basics from course 01, since we'll build on them straight away:

- Test files end in `_test.go` and are only compiled by `go test`.
- Test functions look like `func TestXxx(t *testing.T)`.
- `t.Errorf` reports a failure and carries on; `t.Fatalf` reports and stops that test.
- `go test ./...` runs every package; `-v` shows each test and `-run` filters by name.

That's it. Let's think about *which* tests to write.
