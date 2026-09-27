---
title: Naming Cases and Picking Them with -run
quiz:
  - question: |
      A table has two cases both named `"negative"`. What are the subtests called?
    options:
      - text: '`TestParseAmount/negative` twice; the second overwrites the first'
      - text: '`TestParseAmount/negative` and `TestParseAmount/negative#01`'
        correct: true
      - text: The test fails with a "duplicate subtest" error
      - text: '`TestParseAmount/negative_1` and `TestParseAmount/negative_2`'
    explanation: |
      Go makes names unique by appending `#01`, `#02` and so on. It works,
      but `negative#01` tells you nothing, so give each case a distinct,
      descriptive name.
  - question: Which command runs *only* the case named `"whole dollars"` in `TestParseAmount`?
    options:
      - text: '`go test -run "whole dollars"`'
      - text: '`go test -run TestParseAmount -case whole_dollars`'
      - text: '`go test -run ''TestParseAmount/^whole_dollars$''`'
        correct: true
      - text: '`go test -run TestParseAmount.whole`'
    explanation: |
      Spaces in subtest names become underscores. `-run` splits its pattern on
      `/` and matches each part against one level of the name, as an unanchored
      regular expression, so `^...$` pins it to exactly that case.
---

Once a table has thirty cases, you'll want to run just the one that's failing, and you'll want its name to tell you what broke without opening the file.

## Naming cases

A case name ends up in the failure output, so write it for the person reading a CI log at 2am:

```text
--- FAIL: TestParseAmount (0.00s)
    --- FAIL: TestParseAmount/three_decimal_places_is_an_error (0.00s)
        amount_test.go:41: ParseAmount("12.345") = 1234, nil; want ErrBadAmount
```

Good names describe **the situation**, not the data or the expected result:

| Instead of | Prefer |
| --- | --- |
| `"test1"`, `"case 2"` | `"one decimal digit"` |
| `"12.345"` | `"three decimal places"` |
| `"should work"` | `"negative with leading zero"` |

Some tips:

- Spaces become underscores in the reported name: `"one decimal digit"` runs as `TestParseAmount/one_decimal_digit`.
- **Avoid `/` in names.** It's the subtest separator, so it makes `-run` patterns confusing.
- Keep names unique. Duplicates get `#01`, `#02` appended, which is legal but unhelpful.
- If the input *is* the interesting part, it's fine to use it directly as the name, especially for short inputs. `t.Run(tt.in, ...)` is common for parser tests.

## Picking tests with -run

`-run` takes a regular expression. For subtests, the pattern is split on `/` and each piece is matched, **unanchored**, against the matching level of the name:

```text
go test -run ParseAmount ./...                       # any test whose name contains ParseAmount
go test -run '^TestParseAmount$' ./...               # exactly that test (and all its cases)
go test -run 'TestParseAmount/negative' ./...        # its cases containing "negative"
go test -run 'TestParseAmount/^negative$' ./...      # exactly the "negative" case
go test -run '/decimal' ./...                        # "decimal" cases in every test
```

Because matching is unanchored, `-run Balance` also runs `TestBalanceEmpty` and `TestNegativeBalanceAlert`. Anchor with `^` and `$` when that matters.

## Skipping and listing

- `-skip` is the opposite of `-run`, with the same syntax: `go test -skip 'TestImport/huge' ./...`.
- `-list regexp` prints the names of top-level tests, benchmarks, fuzz tests and examples that match, without running them. `go test -list .` is a quick table of contents. It can't list subtests, since those only exist once the test runs.
- `-run '^$'` matches no tests, which is how you run *only* benchmarks: `go test -run '^$' -bench . ./...`.

## Combining with -v and -count

```text
go test -v -count=1 -run 'TestParseAmount/one_decimal' ./ledgerly
```

`-count=1` bypasses the test cache. Normally `go test` caches passing results for packages whose code and inputs haven't changed, and prints `(cached)`. That's great in general, but when you're poking at a single case you usually want it to really run.

## A workflow that works

1. A CI run fails with `TestImport/semicolon_separated_file`.
2. Copy the name straight into `-run`: `go test -v -run 'TestImport/semicolon_separated_file' ./ledgerly`.
3. Fix, rerun just that, then run the whole package.

That loop only works if names are unique and descriptive, which is the whole point of this lesson.
