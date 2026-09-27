---
title: Writing Benchmarks with b.Loop
quiz:
  - question: |
      How many times does the setup line run in each version, for one
      measurement?

      ```go
      func BenchmarkA(b *testing.B) {
          input := makeCSV(1000) // setup
          for b.Loop() {
              ImportCSV(strings.NewReader(input))
          }
      }

      func BenchmarkB(b *testing.B) {
          input := makeCSV(1000) // setup
          for i := 0; i < b.N; i++ {
              ImportCSV(strings.NewReader(input))
          }
      }
      ```
    options:
      - text: Once in both
      - text: Once in A; several times in B, because the function is called again for each growing value of `b.N`
        correct: true
      - text: Once per iteration in both
      - text: Several times in A; once in B
    explanation: |
      A `b.N` benchmark is called repeatedly with larger `N` until it runs
      long enough, so its setup repeats. A `b.Loop` benchmark is called once
      per measurement and ramps up inside the loop. `b.Loop` also resets the
      timer on its first call, so setup is never counted.
  - question: 'What does `BenchmarkImportCSV/1000-12   2479   101430 ns/op` tell you?'
    options:
      - text: The benchmark took 101 ms in total
      - text: The loop body ran 2479 times, averaging about 101 µs per iteration, with GOMAXPROCS set to 12
        correct: true
      - text: It imported 1000 files using 12 goroutines
      - text: 12 iterations took 2479 ns each
    explanation: |
      The second column is the iteration count (`b.N`), the third the
      average time per iteration. The `-12` suffix is GOMAXPROCS, and
      `/1000` is the sub-benchmark name.
---

"Is it fast enough?" is a question you should answer with measurements, not intuition. Go has benchmarks built into the `testing` package, right next to your tests.

## A first benchmark

A benchmark is a function in a `_test.go` file named `BenchmarkXxx` that takes a `*testing.B`:

```go
func BenchmarkCentsString(b *testing.B) {
	c := Cents(123456)
	for b.Loop() {
		_ = c.String()
	}
}
```

The body of `for b.Loop()` is what gets measured. (The `_ =` is only there to keep `go vet` quiet: it warns about calling a `String` method and ignoring the result, which is usually a bug.) `b.Loop` (Go 1.24) keeps returning `true` until the testing package has enough iterations for a stable measurement, then `false`.

Benchmarks don't run with a plain `go test`. Ask for them with `-bench`, a regular expression like `-run`:

```text
$ go test -run '^$' -bench . ./ledgerly
goos: linux
goarch: amd64
pkg: ledgerly
cpu: AMD Ryzen 5 7500F 6-Core Processor
BenchmarkCentsString-12    	 2784759	        85.10 ns/op
PASS
ok  	ledgerly	1.800s
```

`-run '^$'` matches no tests, so only benchmarks run. The result line reads: name, a `-12` suffix for GOMAXPROCS, the number of iterations, and the average time per iteration. Formatting one amount takes about 85 nanoseconds on this machine. Your numbers will differ, and that's fine; benchmarks compare things on *one* machine.

By default each benchmark runs for about a second. `-benchtime=3s` runs longer, and `-benchtime=500x` runs exactly 500 iterations.

## What b.Loop does for you

- **Setup before the loop runs once, and isn't timed.** `b.Loop` resets the timer the first time it's called, and stops it when it returns `false`, so cleanup after the loop isn't timed either.
- **The work can't be optimised away.** Inside a `for b.Loop()` body, the compiler keeps function arguments and results alive, so it can't delete the `c.String()` call just because the result is thrown away. (More on this trap in lesson 3.)
- **The function runs once per measurement.** It ramps up the iteration count inside the loop.

## The old way: b.N

Before Go 1.24, benchmarks looped to `b.N`, and you'll see this form in lots of existing code:

```go
func BenchmarkCentsString(b *testing.B) {
	c := Cents(123456)
	for i := 0; i < b.N; i++ {
		_ = c.String()
	}
}
```

The testing package calls this function several times with growing values of `b.N` (1, 100, 10000, ...) until it runs long enough. That has consequences:

- Setup before the loop runs every time the function is called. If it's expensive, call `b.ResetTimer()` after it, or it pollutes the measurement.
- Nothing stops the compiler from eliminating work whose result is unused.

Use `b.Loop` for new code. Use one or the other: a benchmark shouldn't contain both. (`go fix` deliberately doesn't rewrite `b.N` loops for you, because the rewrite can change what an old benchmark measures. Convert them by hand, and compare before and after.)

## Sub-benchmarks

Performance often depends on input size. `b.Run` creates sub-benchmarks, just like `t.Run`:

```go
func BenchmarkImportCSV(b *testing.B) {
	for _, n := range []int{10, 1000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			input := makeCSV(n)
			b.SetBytes(int64(len(input)))
			for b.Loop() {
				if _, err := ImportCSV(strings.NewReader(input)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
```

```text
BenchmarkImportCSV/10-12      	  103116	      2412 ns/op	  99.50 MB/s
BenchmarkImportCSV/1000-12    	    2479	    101430 ns/op	 235.73 MB/s
```

`b.SetBytes` tells the testing package how many bytes each iteration processes, and it adds a throughput column. Throughput is higher for the bigger input: fixed per-call costs (creating the CSV reader, reading the header) get spread over more rows. Sizes that differ by 100x or more are a good way to spot costs that don't scale the way you expect.

Notice the `b.Fatal` too. A benchmark that measures an error path by accident is worse than no benchmark. Check errors, even in the loop.

## Further reading

- [Learn Go with Tests: Iteration (benchmarking)](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/iteration)
- [More predictable benchmarking with testing.B.Loop](https://go.dev/blog/testing-b-loop)
