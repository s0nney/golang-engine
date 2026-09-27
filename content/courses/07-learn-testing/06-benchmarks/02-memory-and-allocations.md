---
title: Memory and Allocations
quiz:
  - question: '`BenchmarkCentsString-12   4105833   83.80 ns/op   16 B/op   2 allocs/op`. What do the last two columns mean?'
    options:
      - text: The benchmark used 16 bytes of stack and 2 goroutines
      - text: Each iteration allocated 16 bytes on the heap, in 2 separate allocations, on average
        correct: true
      - text: The whole benchmark allocated 16 bytes in total
      - text: 2 garbage collections ran, freeing 16 bytes
    explanation: |
      `-benchmem` (or `b.ReportAllocs()`) adds heap bytes and heap
      allocation counts, both averaged per iteration. Stack memory isn't
      counted, and it's free to reclaim.
  - question: Why might a test call `testing.AllocsPerRun` instead of relying on a benchmark?
    options:
      - text: It's more accurate at measuring time
      - text: It runs during a normal `go test`, so a change that adds an allocation to a hot path fails the build instead of going unnoticed
        correct: true
      - text: Benchmarks can't measure allocations
      - text: It disables the garbage collector
    explanation: |
      Benchmarks only run when someone asks for them, and nobody reads the
      numbers every day. An `AllocsPerRun` check turns "this doesn't
      allocate" into an ordinary test that CI enforces.
---

Time per operation is only half the story. In Go, allocating memory on the heap costs time twice: once when you allocate, and again when the garbage collector has to find and free it. Code that allocates less usually runs faster, and makes the whole program's GC pauses shorter.

## -benchmem

Add `-benchmem` to see allocations:

```text
$ go test -run '^$' -bench 'CentsString|AppendCents' -benchmem ./ledgerly
BenchmarkCentsString-12    	 4105833	        83.80 ns/op	      16 B/op	       2 allocs/op
BenchmarkAppendCents-12    	39711495	         9.285 ns/op	       0 B/op	       0 allocs/op
```

- **B/op**: heap bytes allocated per iteration.
- **allocs/op**: number of heap allocations per iteration.

To always report them for one benchmark, call `b.ReportAllocs()` at the start of it.

## Where do allocations come from?

`Cents.String` is written as `fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)`. Two allocations per call:

1. The **result string**. `Sprintf` has to return a new string, and it lives on the heap.
2. **Boxing** an `int64` into the `...any` parameter. Putting a value in an interface may need a heap copy. (Go avoids it for small integers, which is why `n%100` doesn't count here but the `1234` dollars do.)

The compiler can tell you what escapes to the heap:

```text
$ go build -gcflags=-m ./ledgerly
./ledgerly.go:24:42: n / int64(100) escapes to heap
./ledgerly.go:24:49: n % int64(100) escapes to heap
```

The "escapes" means the compiler couldn't prove the value stays on the stack. (It's a compile-time answer; the small-integer trick happens at run time.) `-gcflags=-m` output is noisy, so `grep` for the file you care about.

## The append pattern

The standard library's answer to formatting without allocating is **append-style** functions, like `strconv.AppendInt` and `fmt.Appendf`: the caller passes in a buffer and gets back the extended slice.

```go
// AppendCents appends c formatted like c.String() to dst.
func AppendCents(dst []byte, c Cents) []byte {
	n := int64(c)
	if n < 0 {
		dst = append(dst, '-')
		n = -n
	}
	dst = append(dst, '$')
	dst = strconv.AppendInt(dst, n/100, 10)
	return append(dst, '.', byte('0'+n%100/10), byte('0'+n%10))
}
```

The caller reuses one buffer across many calls with `buf = AppendCents(buf[:0], c)`. Once the buffer is big enough, there are zero allocations. That's the 9 ns version above: about nine times faster, with no garbage at all.

You don't need this everywhere. `String()` is fine for a report printed once. The append version matters in hot paths, such as writing a million-line CSV export.

## Preallocate slices

The most common easy win is giving a slice its final capacity up front:

```go
out := make([]string, 0, len(txns)) // instead of: var out []string
for _, tx := range txns {
	out = append(out, tx.Amount.String())
}
```

```text
BenchmarkFormatAll/append-12     3380    91149 ns/op   49140 B/op   1755 allocs/op
BenchmarkFormatAll/prealloc-12   4442    87307 ns/op   30327 B/op   1745 allocs/op
```

Growing a slice by appending reallocates and copies whenever it runs out of room. Preallocating removed 10 allocations and 40% of the bytes. The time barely moved, because the thousand-plus `String` calls dominate. That's a useful lesson in itself: **measure before optimising**, since the obvious fix isn't always where the time goes.

## Lock it in with AllocsPerRun

Once you've made a hot path allocation-free, keep it that way with an ordinary test:

```go
func TestAppendCentsDoesNotAllocate(t *testing.T) {
	buf := make([]byte, 0, 32)
	allocs := testing.AllocsPerRun(100, func() {
		buf = AppendCents(buf[:0], -123456)
	})
	if allocs != 0 {
		t.Errorf("AppendCents allocates %v times per call, want 0", allocs)
	}
}
```

`testing.AllocsPerRun(runs, f)` calls `f` once to warm up, then `runs` times, and returns the average number of allocations. Unlike a benchmark, it runs on every `go test`, so the day someone slips a `fmt.Sprintf` back in, CI goes red.

Allocation counts are far more stable than timings, which makes them safe to assert on. Never assert on nanoseconds in a test: timing depends on the machine and what else it's doing.
