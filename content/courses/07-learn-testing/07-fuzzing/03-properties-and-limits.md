---
title: Good Properties, and What Fuzzing Can't Find
quiz:
  - question: You have `Format(c Cents) string` and `ParseAmount(s string) (Cents, error)`. Which property makes a good fuzz target for them?
    options:
      - text: '`Format(c)` is not empty'
      - text: For every `c`, `ParseAmount(Format(c))` returns `c` and no error
        correct: true
      - text: '`Format(c)` equals `"12.34"`'
      - text: '`ParseAmount` never returns an error'
    explanation: |
      A round trip checks that two functions agree with each other for
      every input, without you knowing any expected output. It catches
      bugs in either function.
  - question: |
      `Cents(math.MinInt64).String()` returns garbage. Fuzzing an `int64`
      for 90 seconds (33 million inputs) didn't find it. Why not?
    options:
      - text: The fuzzer can't generate negative numbers
      - text: One bad value out of 2⁶⁴, with no branch leading towards it, is effectively impossible to hit at random
        correct: true
      - text: Fuzzing only works on strings
      - text: The fuzzer skips values that would fail
    explanation: |
      Coverage guidance helps the fuzzer get past `if` statements, but this
      bug has no special branch; it's an arithmetic edge. Put known boundary
      values like `math.MinInt64` and `math.MaxInt64` into the seed corpus.
---

A fuzz test is only as good as its properties. "Doesn't panic" is a start, but most bugs return a wrong answer rather than crashing. Here are the patterns that catch them.

## Pattern 1: round trips

If you have an encoder and a decoder, decoding what you encoded must give back the original:

```go
func FuzzFormatRoundTrip(f *testing.F) {
	f.Add(int64(1234))
	f.Add(int64(-5))
	f.Fuzz(func(t *testing.T, n int64) {
		c := Cents(n)
		got, err := ParseAmount(Format(c))
		if err != nil {
			t.Fatalf("ParseAmount(Format(%d)) = error %v; Format gave %q", n, err, Format(c))
		}
		if got != c {
			t.Fatalf("ParseAmount(Format(%d)) = %d", n, got)
		}
	})
}
```

Round trips apply everywhere: JSON marshal/unmarshal, CSV write/read, compress/decompress, `String`/`Parse`.

## Pattern 2: an oracle

Compare against a second implementation that's slow or limited but obviously right. In the first lesson, the oracle was a regular expression for the amount grammar. Other oracles:

- the old implementation, when you rewrite something for speed (**differential** fuzzing)
- a brute-force version of a clever algorithm
- the standard library (your `AppendCents` should agree with `fmt.Sprintf` for non-negative inputs)

## Pattern 3: invariants

Things that must be true of any output, whatever the input:

- A statement's `TOTAL` line equals the sum of its rows.
- Sorting returns a permutation of the input, in order.
- `ImportCSV` never returns more transactions than the input has lines.
- A parsed amount's sign matches the input's sign (this is what caught the overflow).

## Pattern 4: no crash on garbage

For code that handles untrusted input (file imports, network protocols), just calling the function with arbitrary bytes is valuable. `ImportCSV(bytes.NewReader(data))` must return an error for junk, never panic, never hang.

## What coverage guidance does

Go's fuzzer is **coverage-guided**. It instruments your code, watches which branches each input reaches, and keeps inputs that reach new ones. It then mutates those. That's how it finds `"+0"` in 64 tries: it doesn't guess randomly, it climbs through the `if` statements of `ParseAmount` and `strconv.ParseInt` one branch at a time.

## What it can't find

Remember the bug left in `Cents.String` back in chapter 1? For negative amounts it computes `n = -n`, and for the smallest `int64`, `-9223372036854775808`, negation overflows and gives back the same negative number. The result is `-$-92,233,720,368,547,758.-8`.

Here's a fuzz test that checks the output looks like a well-formed amount:

```go
var formatted = regexp.MustCompile(`^-?\$[0-9]{1,3}(,[0-9]{3})*\.[0-9]{2}$`)

func FuzzCentsString(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(123456))
	f.Add(int64(-5))
	f.Fuzz(func(t *testing.T, n int64) {
		if s := Cents(n).String(); !formatted.MatchString(s) {
			t.Fatalf("Cents(%d).String() = %q, which isn't a well-formed amount", n, s)
		}
	})
}
```

Run it for a minute and a half:

```text
fuzz: elapsed: 1m30s, execs: 32964539 (0/sec), new interesting: 0 (total: 3)
PASS
```

Thirty-three million inputs, no failure. The property is right and the bug is real, but there's exactly one bad value in 2⁶⁴, and no branch in `String` points the fuzzer towards it. "New interesting: 0" is the tell: the fuzzer found nothing new to explore.

The fix is to **seed the boundaries yourself**:

```go
f.Add(int64(math.MinInt64))
f.Add(int64(math.MaxInt64))
```

Now a plain `go test` fails immediately, no fuzzing needed. Fuzzing and hand-picked cases complement each other: you know where the cliffs are (zero, one, max, min, empty, huge), and the fuzzer explores everything in between.

## So what should String do?

Once found, you still have to decide. For Ledgerly, `math.MinInt64` cents is about -92 quadrillion dollars, so no real ledger will contain it. Reasonable fixes are handling it specially, or working with `uint64` for the magnitude (`uint64(-n)` is correct even for the minimum value). What matters is that the behaviour is now a *decision*, pinned by a test, rather than an accident.
