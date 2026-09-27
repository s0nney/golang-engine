---
title: Test Coverage
quiz:
  - question: A package has 100% statement coverage. What does that guarantee?
    options:
      - text: The package has no bugs
      - text: Every statement ran at least once during the tests, and nothing more
        correct: true
      - text: Every branch combination was tested
      - text: Every function's result was checked by an assertion
    explanation: |
      Coverage records what *executed*, not what was *checked*. A test that
      calls every function and asserts nothing gets 100% too. Coverage is
      great at showing what's untested, and weak evidence that anything is
      tested well.
  - question: Which command shows coverage per function in the terminal?
    options:
      - text: '`go test -cover -v`'
      - text: '`go test -coverprofile=cover.out` then `go tool cover -func=cover.out`'
        correct: true
      - text: '`go tool cover -html`'
      - text: '`go vet -cover`'
    explanation: |
      `-coverprofile` writes the raw data; `go tool cover -func` summarises
      it per function, and `-html` opens an annotated view of the source.
---

How do you know which parts of Ledgerly your tests actually exercise? Go can measure it for you. **Coverage** instruments the code, runs the tests, and records which statements executed.

## A percentage

```text
$ go test -cover ./ledgerly
ok  	ledgerly	0.002s	coverage: 70.8% of statements
```

That's the fraction of statements that ran at least once. On its own, a single number isn't very useful. What's useful is finding out *which* 29.2% never ran.

## A profile

Write the raw data to a file, then ask for a per-function breakdown:

```text
$ go test -coverprofile=cover.out ./ledgerly
ok  	ledgerly	0.002s	coverage: 70.8% of statements
$ go tool cover -func=cover.out
ledgerly/amount.go:15:	ParseAmount	92.9%
ledgerly/amount.go:37:	Format		0.0%
ledgerly/amount.go:45:	allDigits	66.7%
total:			(statements)	70.8%
```

Two findings straight away:

- `Format` is never called by any test. Every bug in it is invisible.
- `allDigits` is only two-thirds covered. The tests check `"12.345"`, which is rejected by the length check before `allDigits` ever sees a letter. So no test proves that `"12.ab"` or `""` is rejected.

## The HTML view

```text
$ go tool cover -html=cover.out
```

This opens your browser with the source code coloured green (covered) and red (not covered). It's the fastest way to see exactly which branches are missing. Here it would show the two `return false` lines in `allDigits` in red, and the whole of `Format`.

Add `-o coverage.html` to write the file instead of opening a browser, which is handy in CI.

## Modes

`-covermode` picks what's recorded:

- `set` (the default): did each statement run?
- `count`: how many times did it run? The HTML view then shades hot code more brightly.
- `atomic`: like `count`, but safe with parallel goroutines. It's the default when you use `-race`.

## Coverage across packages

By default each package's coverage only counts its own tests. If `ledgerly/cli` has tests that exercise `ledgerly`, that doesn't count for `ledgerly`. `-coverpkg` changes that:

```text
go test -coverpkg=./... -coverprofile=cover.out ./...
```

Now every test run counts towards every package in the module, which gives a more honest picture of what the whole suite exercises.

## What coverage doesn't tell you

Coverage is a **flashlight, not a grade**. It's excellent at finding code nobody tests. It's bad at telling you whether code is tested *well*:

```go
func TestParseAmountRuns(t *testing.T) {
	ParseAmount("12.34")
	ParseAmount("abc")
	ParseAmount("-0.05")
}
```

That test asserts nothing, and it still pushes coverage up. The table-catches-bugs exercise in chapter 3 used a stronger idea, mutation testing: break the code on purpose and check that some test fails.

Some practical advice:

- Look at the **red lines**, not the percentage. Ask "would a bug here matter?" For error handling in an importer, yes. For a `String` method used in debug logs, maybe not.
- Don't set a hard 100% target. The last few percent are usually impossible-to-trigger error paths, and chasing them produces tests that are all setup and no insight. Many teams watch that coverage doesn't *drop* in a change instead.
- Coverage of a line tells you it ran with *some* input. Boundary bugs (`<` versus `<=`) live on fully covered lines. That's what table tests and fuzzing are for.

## Coverage for programs, too

`go build -cover` builds a binary that records coverage while it runs (set `GOCOVERDIR` to a directory for the data). That lets you measure what end-to-end or integration tests exercise, not only unit tests. `go tool covdata` then converts the data into the same profile format as above.
