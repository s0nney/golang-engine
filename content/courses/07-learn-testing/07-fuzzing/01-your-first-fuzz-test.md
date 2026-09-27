---
title: Your First Fuzz Test
quiz:
  - question: What does a plain `go test` (without `-fuzz`) do with `FuzzParseAmount`?
    options:
      - text: Skips it
      - text: Fuzzes it for one second
      - text: Runs the fuzz function once for each seed input (and each file in its testdata corpus), like a table test
        correct: true
      - text: Fails, because fuzz tests need `-fuzz`
    explanation: |
      Without `-fuzz`, a fuzz test is a regular test over its seed corpus:
      the `f.Add` values plus any files in `testdata/fuzz/FuzzParseAmount/`.
      Generating new inputs only happens with `-fuzz`.
  - question: |
      Which of these fuzz function signatures is invalid?
    options:
      - text: '`f.Fuzz(func(t *testing.T, s string) { ... })`'
      - text: '`f.Fuzz(func(t *testing.T, data []byte, n int64) { ... })`'
      - text: '`f.Fuzz(func(t *testing.T, tx Transaction) { ... })`'
        correct: true
      - text: '`f.Fuzz(func(t *testing.T, ok bool, r rune) { ... })`'
    explanation: |
      Fuzz arguments must be basic types the fuzzer knows how to mutate:
      `string`, `[]byte`, the integer and float types, `bool` and `rune`/`byte`.
      To fuzz a struct, take its fields as separate arguments and build it
      inside the function.
---

Every test you've written so far checks inputs *you* thought of. The bugs that reach production are usually in inputs nobody thought of. **Fuzzing** has the computer generate inputs for you, millions of them, steered towards code paths that haven't been explored yet.

## The shape of a fuzz test

A fuzz test lives in a `_test.go` file, is named `FuzzXxx`, and takes a `*testing.F`:

```go
func FuzzParseAmount(f *testing.F) {
	for _, s := range []string{"12.34", "12", "12.3", "-0.05", "", ".50", "12.345"} {
		f.Add(s) // seed corpus
	}
	f.Fuzz(func(t *testing.T, s string) {
		c, err := ParseAmount(s)
		if valid := validAmount.MatchString(s); valid != (err == nil) {
			t.Fatalf("ParseAmount(%q) error = %v, but the grammar says valid = %v", s, err, valid)
		}
		if err != nil {
			return
		}
		if neg := strings.HasPrefix(s, "-"); (c < 0) != (neg && c != 0) {
			t.Fatalf("ParseAmount(%q) = %d: wrong sign", s, c)
		}
	})
}
```

It has two parts:

- **The seed corpus.** `f.Add` provides example inputs. Their types must match the fuzz function's parameters after `t`. Good seeds are valid inputs and near misses; the fuzzer mutates them to make new inputs.
- **The fuzz target.** `f.Fuzz` takes a function whose first parameter is `*testing.T` and whose other parameters are the fuzzed values. It's called once per input and uses `t` exactly like a normal test.

## What does it check?

Here's the interesting part. With random input you don't know the expected result, so you can't write `want`. Instead you check **properties** that must hold for *every* input. This one checks two:

1. `ParseAmount` accepts exactly the strings that match the documented grammar. `validAmount` is a regular expression, `^-?[0-9]+(\.[0-9]{1,2})?$`, which is slow but obviously correct. It's an **oracle**.
2. If a string parses, the result is negative exactly when the string starts with `-` (and isn't zero).

On top of that, the fuzzer checks one property for free: **the code must not panic** (or hang, or call `os.Exit`). A crash is a failure.

## Running it

A plain `go test` runs the fuzz target once per seed, like a table test:

```text
$ go test -run FuzzParseAmount -v ./ledgerly
=== RUN   FuzzParseAmount
=== RUN   FuzzParseAmount/seed#0
...
--- PASS: FuzzParseAmount (0.00s)
```

To actually fuzz, pass `-fuzz` with a pattern that matches **exactly one** fuzz test:

```text
$ go test -run '^$' -fuzz FuzzParseAmount ./ledgerly
fuzz: elapsed: 0s, gathering baseline coverage: 0/7 completed
fuzz: elapsed: 0s, gathering baseline coverage: 7/7 completed, now fuzzing with 12 workers
fuzz: elapsed: 0s, execs: 64 (2884/sec), new interesting: 0 (total: 7)
--- FAIL: FuzzParseAmount (0.02s)
    --- FAIL: FuzzParseAmount (0.00s)
        amount_test.go:18: ParseAmount("+0") error = <nil>, but the grammar says valid = false

    Failing input written to testdata/fuzz/FuzzParseAmount/0a2a901998b6f11e
    To re-run:
    go test -run=FuzzParseAmount/0a2a901998b6f11e
FAIL
```

Sixty-four inputs in, it found a bug. `ParseAmount` uses `strconv.ParseInt`, which happily accepts a leading `+`, so `"+0"` and `"+12.50"` are accepted even though the documented format doesn't allow them. Ledgerly's CSV export would never produce that, but a user's spreadsheet might, and now two parts of the system disagree about what a valid amount is.

Fuzzing runs until it finds a failure or you stop it with Ctrl+C. Use `-fuzztime 30s` (or `-fuzztime 100000x` for a number of inputs) to put a limit on it.

## Rules for fuzz targets

- **Deterministic.** The same input must give the same result, or failures can't be reproduced. No clocks, no randomness, no global state carried between calls.
- **Fast.** Each call should take microseconds. The fuzzer's power comes from volume.
- **Don't modify the arguments.** Especially `[]byte` slices: the fuzzer may reuse their memory.
- **Supported types only**: `string`, `[]byte`, `bool`, `byte`, `rune`, the sized and unsized `int` and `uint` types, `float32` and `float64`.

## Further reading

- [Go Fuzzing](https://go.dev/doc/security/fuzz/)
- [Tutorial: Getting started with fuzzing](https://go.dev/doc/tutorial/fuzz)
