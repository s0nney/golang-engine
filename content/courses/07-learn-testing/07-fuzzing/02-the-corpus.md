---
title: The Corpus and Failing Inputs
quiz:
  - question: The fuzzer wrote `testdata/fuzz/FuzzParseAmount/029cabfca8431a6c`. What should you do with it after fixing the bug?
    options:
      - text: Delete it, since the bug is fixed
      - text: Commit it, so every normal `go test` run replays that input as a regression test
        correct: true
      - text: Move it to `$GOCACHE`
      - text: Add it to `.gitignore`
    explanation: |
      Files in `testdata/fuzz/<FuzzName>/` are part of the seed corpus. A
      plain `go test` runs each one as a subtest, so the bug can never come
      back unnoticed.
  - question: Where does the fuzzer keep the *interesting* inputs it generates while fuzzing (the ones that found new coverage but didn't fail)?
    options:
      - text: In `testdata/fuzz`, next to the failing inputs
      - text: In the build cache (`$GOCACHE/fuzz`), outside your repository
        correct: true
      - text: Nowhere; they're discarded when fuzzing stops
      - text: In memory only, for the current process
    explanation: |
      The generated corpus lives in the cache so it doesn't clutter your
      repo, and later fuzzing runs pick up where earlier ones left off.
      `go clean -fuzzcache` deletes it.
---

Fuzzing found `"+0"`. Now what? This lesson covers what the fuzzer leaves behind and the workflow for turning a crash into a fixed bug with a permanent regression test.

## The failing input file

When a fuzz target fails, the fuzzer first **minimises** the input (tries shorter and simpler variations that still fail), then writes it to a file:

```text
$ cat testdata/fuzz/FuzzParseAmount/0a2a901998b6f11e
go test fuzz v1
string("+0")
```

The first line is a format version. Then there's one line per fuzz argument, written as a Go literal with its type, so multi-argument targets look like:

```text
go test fuzz v1
string("rent")
int64(-9223372036854775808)
```

You can write these files by hand, too, if you want a seed that's awkward to express with `f.Add`.

## Replaying it

That file is now part of the **seed corpus**. Every plain `go test` runs it, as a subtest named after the file:

```text
$ go test -run FuzzParseAmount -v ./ledgerly
=== RUN   FuzzParseAmount
=== RUN   FuzzParseAmount/0a2a901998b6f11e
    amount_test.go:18: ParseAmount("+0") error = <nil>, but the grammar says valid = false
--- FAIL: FuzzParseAmount (0.00s)
    --- FAIL: FuzzParseAmount/0a2a901998b6f11e (0.00s)
```

So the workflow is:

1. The fuzzer fails and writes a file.
2. Run `go test -run=FuzzParseAmount/0a2a901998b6f11e` to reproduce it without fuzzing. Debug it like any failing test.
3. Fix the code (here: check that the dollars and cents are all digits instead of trusting `ParseInt`).
4. Commit the fix **and** the testdata file. It's now a regression test.
5. Fuzz again. There's often a second bug hiding behind the first.

Doing exactly that with Ledgerly, the second run fails almost immediately on something new:

```text
--- FAIL: FuzzParseAmount (0.39s)
    --- FAIL: FuzzParseAmount (0.00s)
        amount_test.go:24: ParseAmount("100000000000000000") = -8446744073709551616: wrong sign
```

A hundred quadrillion dollars, times 100 cents, overflows `int64` and wraps round to a negative number. Silent integer overflow in a money library! Nobody would have written that test case by hand.

If you'd rather keep the case in your table test than in a file, copy it into `f.Add` or your table. Either works; the file is just less effort.

## The generated corpus

While fuzzing, the fuzzer keeps every input that reached new code ("new interesting" in the progress lines). Those go in the build cache, not your repository:

```text
fuzz: elapsed: 3s, execs: 1053481 (351151/sec), new interesting: 131 (total: 138)
```

Later runs start from that corpus, so fuzzing effort accumulates across sessions on the same machine. `go clean -fuzzcache` throws it away.

## Useful flags

| Flag | Meaning |
| --- | --- |
| `-fuzz FuzzName` | fuzz the one fuzz test matching the pattern |
| `-fuzztime 30s` / `-fuzztime 50000x` | stop after a time or a number of inputs (default: forever) |
| `-fuzzminimizetime 10s` | time limit for minimising a failing input |
| `-parallel 4` | number of fuzzing worker processes (default GOMAXPROCS) |
| `-run '^$'` | skip the normal tests, which run first otherwise |

Only one package and one fuzz test can be fuzzed at a time.

## Fuzzing in CI

Seed corpora run on every `go test`, so they're always in CI. Actual fuzzing is open-ended, so teams usually either run it for a fixed time on a schedule (`-fuzztime 10m` nightly) or use a continuous fuzzing service. For a small library like Ledgerly, a minute of fuzzing after changing the parser catches most of what it'll ever find.
