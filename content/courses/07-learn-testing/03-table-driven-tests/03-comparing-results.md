---
title: Comparing Structs, Slices and Maps
quiz:
  - question: |
      What does this print?

      ```go
      var a []string
      b := []string{}
      fmt.Println(reflect.DeepEqual(a, b), slices.Equal(a, b))
      ```
    options:
      - text: '`true true`'
      - text: '`false false`'
      - text: '`false true`'
        correct: true
      - text: '`true false`'
    explanation: |
      `reflect.DeepEqual` treats a nil slice and an empty non-nil slice as
      different. `slices.Equal` only compares length and elements, so two
      empty slices are equal whether or not they're nil.
  - question: |
      `got` and `want` are `Transaction` values whose fields are `Account string`,
      `Amount Cents` and `Tags []string`. What happens with `if got != want`?
    options:
      - text: It compares every field, including the slice's elements
      - text: It compares only the first field
      - text: It doesn't compile, because a struct with a slice field isn't comparable
        correct: true
      - text: It compares the slice by pointer
    explanation: |
      `==` works on structs only when every field is comparable. Slices, maps
      and funcs aren't, so the struct isn't either, and the compiler rejects
      the comparison. Use `reflect.DeepEqual` or compare field by field.
---

So far every `want` in our tables has been a number or a string, where `!=` just works. Real functions return structs, slices and maps. Comparing those is where many Go tests get sloppy.

## Comparable structs: just use ==

If every field of a struct is comparable (numbers, strings, bools, pointers, arrays, other comparable structs), `==` compares them all:

```go
type Posting struct {
	Account string
	Amount  Cents
}

got := SplitBill("dinner", 1000, 3)[0]
want := Posting{Account: "dinner", Amount: 334}
if got != want {
	t.Errorf("SplitBill(...)[0] = %+v, want %+v", got, want)
}
```

`%+v` prints field names (`{Account:dinner Amount:334}`), which makes failure messages much easier to read than a bare `%v`.

## Slices: slices.Equal

Slices can't be compared with `==` (only against `nil`). For slices of comparable elements, use `slices.Equal`:

```go
got := l.Accounts()
want := []string{"cash", "rent", "savings"}
if !slices.Equal(got, want) {
	t.Errorf("Accounts() = %q, want %q", got, want)
}
```

`%q` on a `[]string` quotes each element, so `["cash" "rent "]` shows the stray space you'd never spot otherwise. There's also `slices.EqualFunc` when elements need custom comparison, and `maps.Equal` / `maps.EqualFunc` for maps:

```go
want := map[string]Cents{"cash": 1250, "savings": 9900}
if got := l.Balances(); !maps.Equal(got, want) {
	t.Errorf("Balances() = %v, want %v", got, want)
}
```

Printing a map with `%v` sorts its keys, so the output is stable even though map order isn't.

## Everything else: reflect.DeepEqual

A `Transaction` with a `Tags []string` field isn't comparable, so `got != want` won't compile. `reflect.DeepEqual` compares anything, recursively:

```go
if !reflect.DeepEqual(got, want) {
	t.Errorf("ParseLine(%q) =\n%#v\nwant\n%#v", line, got, want)
}
```

It has gotchas, and they bite in tests:

- **nil vs empty.** `DeepEqual([]string(nil), []string{})` is `false`. Worse, `%+v` prints both as `[]`, so the failure message shows two identical-looking values. Use `%#v`, which prints `[]string(nil)` vs `[]string{}`, or decide which one your API returns and document it.
- **Unexported fields are compared too.** A cache or a mutex inside a struct can make two "equal" values differ.
- **Funcs are never equal** (unless both are nil), and neither are `NaN` floats.
- **No explanation.** It returns a bare `bool`. On a 20-field struct you're left squinting at two long lines.

## Compare what matters

Often the best comparison isn't "the whole value". Check the fields the test is about, one `if` each, so the message says exactly what's wrong:

```go
tx, err := ParseLine("2026-03-01,rent,-1200.00")
if err != nil {
	t.Fatalf("ParseLine: %v", err)
}
if tx.Account != "rent" {
	t.Errorf("Account = %q, want %q", tx.Account, "rent")
}
if tx.Amount != -120000 {
	t.Errorf("Amount = %v, want %v", tx.Amount, Cents(-120000))
}
```

That's more lines but far better failures, and the test doesn't break when someone adds an unrelated field.

## What about go-cmp?

Many projects use the third-party `github.com/google/go-cmp/cmp` package, whose `cmp.Diff(want, got)` prints a readable diff of any two values. It's excellent, and you'll see it in lots of codebases. This course sticks to the standard library, and for most tests `==`, `slices.Equal`, `maps.Equal` and good messages are enough.

## Got before want

Whatever you compare, follow the Go convention in messages: **got first, then want**, as in `Balance("cash") = $12.50, want $20.00`. Every Go developer reads failures in that order, so swapping them causes real confusion.
