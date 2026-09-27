---
title: Example Functions
quiz:
  - question: |
      What happens to this example when you run `go test`?

      ```go
      func ExampleParseAmount() {
          c, _ := ledgerly.ParseAmount("12.3")
          fmt.Println(c)
      }
      ```
    options:
      - text: It runs, and passes if it doesn't panic
      - text: It's compiled but not run, because it has no output comment
        correct: true
      - text: It fails because there's no `// Output:` comment
      - text: It isn't compiled at all
    explanation: |
      Examples without an output comment are compiled (so they can't rot
      into code that doesn't build) but not executed. Add `// Output:` to
      have `go test` run it and compare what it prints.
  - question: Which name attaches an example to the `Balance` method of `*Ledger`, as a second example called "empty"?
    options:
      - text: '`ExampleBalance_empty`'
      - text: '`ExampleLedger_Balance_empty`'
        correct: true
      - text: '`Example_Ledger_Balance_Empty`'
      - text: '`ExampleLedgerBalanceEmpty`'
    explanation: |
      The pattern is `Example` + type + `_` + method, plus an optional
      `_suffix` that starts with a lower-case letter. The pointer receiver
      doesn't appear in the name.
---

Documentation goes stale. Someone changes a function, forgets the comment, and now the docs lie. Go has a neat answer for the most valuable kind of documentation, the usage example: make it a test.

## An Example function

An **example** is a function in a `_test.go` file whose name starts with `Example`. It takes no arguments, returns nothing, and ends with an `// Output:` comment:

```go
package ledgerly_test

import (
	"fmt"

	"ledgerly"
)

func ExampleParseAmount() {
	c, err := ledgerly.ParseAmount("12.3")
	fmt.Println(c, err)
	_, err = ledgerly.ParseAmount("12.345")
	fmt.Println(err)
	// Output:
	// $12.30 <nil>
	// ledgerly: bad amount: "12.345"
}
```

`go test` runs it, captures everything it prints to standard output, and compares that with the comment (ignoring leading and trailing whitespace). If they differ, the example fails:

```text
=== RUN   ExampleParseAmount_negative
--- FAIL: ExampleParseAmount_negative (0.00s)
got:
-$0.05
want:
-$0.10
FAIL
```

And the same function appears in the package's documentation, on pkg.go.dev and in `go doc`, right under `ParseAmount`. One piece of code, two jobs: it shows readers how to use the API, and it proves the demonstration still works.

## The external test package

Notice `package ledgerly_test`. A `_test.go` file may declare the package name with a `_test` suffix. It's compiled as a **separate package** that imports `ledgerly` like any other user would, so it can only use exported names.

That's exactly what you want for examples: they show code a user can copy, with the `ledgerly.` prefix and no private helpers. It's also a good way to write "black-box" tests of your public API. Both kinds of test file can live side by side in the same directory.

## Naming examples

The name decides where the example shows up in the docs:

| Function | Documents |
| --- | --- |
| `Example()` | the package as a whole |
| `ExampleParseAmount()` | the function `ParseAmount` |
| `ExampleCents()` | the type `Cents` |
| `ExampleCents_String()` | the method `Cents.String` |
| `ExampleLedger_Balance()` | the method `(*Ledger).Balance` |

To have several examples for the same thing, add a suffix that starts with a lower-case letter: `ExampleParseAmount_negative`, `ExampleLedger_Balance_empty`, `Example_csvImport`. `go vet` checks these names and complains if an example refers to something that doesn't exist.

## Output comments

- `// Output:` followed by the expected lines. A short output can go on the same line: `// Output: -$0.05`.
- `// Unordered output:` when the lines can come in any order, like when ranging over a map. The lines are compared as a set.
- No output comment at all: the example is compiled but **not run**. Use this for examples that need a network or a real file, where running them in a test would fail.

```go
func Example_accounts() {
	balances := map[string]ledgerly.Cents{"rent": -120000, "cash": 5000}
	for name, c := range balances {
		fmt.Println(name, c)
	}
	// Unordered output:
	// cash $50.00
	// rent -$1200.00
}
```

## Whole-file examples

Sometimes an example needs a helper type or function, say a custom `Notifier` to show how Ledgerly's alerts work. If a file contains a single example function plus other declarations, and no tests or benchmarks, the documentation shows the *whole file* as the example. Name it something like `example_notifier_test.go`.

## When to write examples

Examples are for **readers**. Write them for the functions people reach for first (`ParseAmount`, `NewLedger`, `ImportCSV`) and for anything whose usage isn't obvious from its signature. Keep them short and realistic. They're not a replacement for tests: an example demonstrates one happy path, while a table test checks the edges.

## Further reading

- [Testable Examples in Go](https://go.dev/blog/examples)
