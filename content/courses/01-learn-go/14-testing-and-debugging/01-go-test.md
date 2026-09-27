---
title: Your First Test
quiz:
  - question: Which of these is a valid Go test?
    options:
      - text: '`func testSegments(t *testing.T)` in `billing.go`'
      - text: '`func TestSegments(t *testing.T)` in `billing_test.go`'
        correct: true
      - text: '`func TestSegments()` in `billing_test.go`'
      - text: '`func Segments_test(t testing.T)` in `test_billing.go`'
    explanation: |
      Test files must end in `_test.go`, and test functions must start with
      `Test` (capital T) and take exactly one parameter, `t *testing.T`.
  - question: What is the difference between `t.Errorf` and `t.Fatalf`?
    options:
      - text: '`t.Errorf` marks the test as failed but keeps running it; `t.Fatalf` marks it failed and stops it immediately'
        correct: true
      - text: '`t.Fatalf` crashes the whole program, including other tests'
      - text: '`t.Errorf` only prints a warning and the test still passes'
    explanation: |
      Use `t.Errorf` when later checks are still meaningful, so you see every
      failure at once. Use `t.Fatalf` when carrying on makes no sense, for
      example when setup failed or a value you're about to use is nil.
---

How do you know your code works? You could run it and eyeball the output, but that gets old fast, and you'd have to do it again after every change. Instead, you write code that checks your code: **tests**. Go has testing built right in.

## A function to test

Here's a function from Textio's `billing` package. It's supposed to work out how many 160-character SMS segments a message needs:

```go
package billing

// Segments returns how many SMS segments a message body needs.
func Segments(body string) int {
	return len(body) / 160
}
```

Can you spot the bug? Let's write a test and let Go spot it for us.

## Writing the test

Tests live next to the code, in a file whose name ends in **`_test.go`**. Put this in `billing/billing_test.go`:

```go
package billing

import "testing"

func TestSegments(t *testing.T) {
	got := Segments("hello")
	want := 1
	if got != want {
		t.Errorf("Segments(%q) = %d, want %d", "hello", got, want)
	}
}
```

The rules:

- The file name ends in `_test.go`. These files are only compiled when you run tests, never into your program.
- Each test is a function whose name starts with **`Test`** followed by a capital letter, taking one parameter: `t *testing.T`.
- The test calls your code, compares what it **got** with what you **want**, and reports a failure with `t.Errorf` if they differ.

The `got`/`want` naming and the `Segments(input) = got, want want` message format are Go conventions. When a test fails, you immediately see what was called, what came back and what was expected.

## Running tests

Run `go test` with the packages to test. `./...` means every package in the module:

```text
$ go test ./...
--- FAIL: TestSegments (0.00s)
    billing_test.go:9: Segments("hello") = 0, want 1
FAIL
FAIL	github.com/textio/smsapp/billing	0.001s
FAIL
```

Caught it! `5 / 160` is `0` in integer division. A 5-character message still needs one segment. We need to round **up**, and a classic trick for that is to add `159` before dividing:

```go
func Segments(body string) int {
	return (len(body) + 159) / 160
}
```

```text
$ go test ./...
ok  	github.com/textio/smsapp/billing	0.001s
```

Add `-v` (verbose) to see each test as it runs, and `-run` with a pattern to run only matching tests: `go test -v -run Segments ./billing`.

## Red, green, refactor

Many Go programmers write the test **first**:

1. **Red:** write a test for behaviour that doesn't exist yet, and watch it fail. (That proves the test can actually fail.)
2. **Green:** write the simplest code that makes it pass.
3. **Refactor:** tidy up the code, with the test making sure you didn't break anything.

This is called **test-driven development** (TDD). You don't have to work this way all the time, but it's a great habit, especially when fixing bugs: first write a test that reproduces the bug, then fix it.

## `t.Errorf` versus `t.Fatalf`

- `t.Errorf` (and `t.Error`) records a failure and **keeps going**, so one run can report several problems.
- `t.Fatalf` (and `t.Fatal`) records a failure and **stops this test** immediately. Use it when continuing makes no sense.

## `t.Context`

Much of Textio's code takes a `context.Context`, a value that carries deadlines and cancellation signals (so a slow carrier call can be abandoned when a user gives up). Since Go 1.24, tests get one for free with `t.Context()`. It's cancelled automatically when the test finishes, so nothing a test starts is left running afterwards:

```go
func TestSend(t *testing.T) {
	ctx := t.Context()

	if err := Send(ctx, "+1-555-0100", "hi"); err != nil {
		t.Fatalf("Send returned an unexpected error: %v", err)
	}
	if err := Send(ctx, "", "hi"); err == nil {
		t.Error("Send with no recipient: got nil error, want an error")
	}
}
```

Before Go 1.24, you'd write `context.WithCancel(context.Background())` and remember to cancel it yourself. `go fix` can update old tests for you.

## Further reading

- [Learn Go with Tests: Hello, World](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/hello-world)
- [Go by Example: Testing and Benchmarking](https://gobyexample.com/testing-and-benchmarking)
