---
title: Tables and Subtests
quiz:
  - question: |
      A `TestImport` table has grown fields like `checkReport bool` and
      `skipHeader bool`, and its loop body is full of `if tt.checkReport`
      branches. What's the best fix?
    options:
      - text: Add a `name` to every branch so failures are easier to read
      - text: Split it into two or more smaller tables (or plain tests), each with a simple loop body
        correct: true
      - text: Switch from a slice of cases to a map of cases
      - text: Move the branches into a helper that takes every flag as a parameter
    explanation: |
      A table should be data plus one straightforward loop body. Once cases
      need flags that change what the loop *does*, the test is harder to
      read than the code under test. Several small tables are clearer than
      one clever one.
  - question: Why do most Go table tests use a slice of structs rather than a `map[string]struct{...}` keyed by name?
    options:
      - text: Maps can't hold structs
      - text: A slice runs the cases in the order they're written; a map's order is random on every run
        correct: true
      - text: Slices are required by `t.Run`
      - text: Maps can't be ranged over in tests
    explanation: |
      Both work, and a map guarantees unique names. But a map runs cases in a
      different order each time, which makes output harder to compare. Some
      people like that it exposes order dependencies. Slices are the common
      default.
---

You already know the shape from course 01: a slice of anonymous structs, a loop, and `t.Run` for each case. Here's a quick recap in Ledgerly terms, then we'll go past the basics, because tables are *the* Go testing idiom. Nearly every test in the standard library is one.

## The shape, with Ledgerly

Ledgerly's `ParseAmount` turns what a person typed into `Cents`. The rules: an optional `-`, some digits, and optionally a `.` followed by **one or two** digits. Anything else is an error.

```go
func TestParseAmount(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Cents
	}{
		{name: "dollars and cents", in: "12.34", want: 1234},
		{name: "whole dollars", in: "12", want: 1200},
		{name: "one decimal digit", in: "12.3", want: 1230},
		{name: "negative", in: "-0.05", want: -5},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmount(tt.in)
			if err != nil {
				t.Fatalf("ParseAmount(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseAmount(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
```

One habit worth adopting now: **keyed fields** in each case (`in: "12.34"`). With more than two or three fields, positional literals like `{"negative", "-0.05", -5}` get hard to read, and adding a field breaks every line.

## What t.Run gives you

You saw the headline benefits in course 01: each case is named in the output, one failing case doesn't hide the others (even `t.Fatalf` only stops its own subtest), and `-run` can pick out a single case. Two more matter once suites grow:

- **Per-case features.** `t.Parallel()`, `t.Cleanup`, `t.TempDir` and helpers all work per case, because each subtest has its own `t`.
- **A result.** `t.Run` returns `false` if the subtest failed, which you can use to stop early when later steps depend on an earlier one:

```go
if !t.Run("import", testImport) {
	t.FailNow() // no point checking the report if the import failed
}
t.Run("report", testReport)
```

## Keep the table dumb

The table should be **data**. The moment cases grow `if tt.special` branches in the loop body, the test gets harder to read than the code. Signs you need a second table (or a plain test):

- Half the fields are only used by some cases.
- The loop body has conditionals on case fields other than expected results.
- A case's name needs a paragraph to explain it.

Two small tables, `TestParseAmountValid` and `TestParseAmountInvalid`, are often clearer than one clever one.

## Maps as tables

You'll sometimes see:

```go
cases := map[string]struct {
	in   string
	want Cents
}{
	"dollars and cents": {in: "12.34", want: 1234},
	"negative":          {in: "-0.05", want: -5},
}
for name, tt := range cases {
	t.Run(name, func(t *testing.T) { /* ... */ })
}
```

Map keys can't collide, and the random order will expose cases that accidentally depend on each other. The slice form is more common because output order is stable. Either is fine; be consistent within a package.

## Further reading

- [Go Wiki: Table-driven tests](https://go.dev/wiki/TableDrivenTests)
