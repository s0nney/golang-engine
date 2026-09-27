---
title: The Testing Pyramid
quiz:
  - question: Ledgerly's `ParseAmount("12.34")` turns a string into `Cents(1234)`. Where should most of its edge-case tests live?
    options:
      - text: In an end-to-end test that uploads a CSV through the web app
      - text: In unit tests that call `ParseAmount` directly
        correct: true
      - text: In a manual QA checklist
      - text: Nowhere; parsing is too simple to test
    explanation: |
      Unit tests are fast and precise: when one fails, it points straight at
      `ParseAmount`. Testing twenty edge cases through the whole app would be
      slow, and a failure could come from anywhere in the stack.
  - question: Why does the pyramid have *fewer* end-to-end tests than unit tests?
    options:
      - text: End-to-end tests are less valuable
      - text: Go can't run end-to-end tests
      - text: They are slower, flakier and harder to debug, so you use a few of them to check the pieces fit together
        correct: true
      - text: The pyramid is upside down; you should have mostly end-to-end tests
    explanation: |
      End-to-end tests catch wiring mistakes nothing else can, but each one is
      expensive to run and a failure says "something is broken somewhere". A
      handful covering the main paths gives you most of their value.
---

Not all tests are the same size. A useful mental model is the **testing pyramid**: lots of small, fast tests at the bottom, fewer bigger ones on top.

```text
            /\
           /  \        end-to-end: a few
          /----\
         /      \      integration: some
        /--------\
       /          \    unit: lots
      /------------\
```

## Unit tests

A **unit test** checks one small piece, usually a function or a type, in isolation. For Ledgerly:

```go
func TestParseAmount(t *testing.T) {
	got, err := ParseAmount("12.34")
	if err != nil {
		t.Fatalf("ParseAmount(%q) returned error: %v", "12.34", err)
	}
	if got != 1234 {
		t.Errorf("ParseAmount(%q) = %d, want 1234", "12.34", got)
	}
}
```

Unit tests are fast (microseconds), deterministic (no disk, network or clock), and precise: when `TestParseAmount` fails, you know exactly where to look. You can afford hundreds of them, so this is where edge cases go: negative amounts, missing decimals, garbage input.

## Integration tests

An **integration test** checks that several pieces work together, often including a real resource: a file on disk, a database, an HTTP server. For Ledgerly that might be "import this CSV file from disk into a ledger and check the balances":

```go
func TestImportFile(t *testing.T) {
	f, err := os.Open("testdata/march.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	l, err := Import(f)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if got := l.Balance("cash"); got != 48210 {
		t.Errorf("cash balance = %d, want 48210", got)
	}
}
```

Integration tests catch mistakes in the *seams*: the CSV reader and the ledger each work alone but disagree about column order, for example. They're slower and need more setup, so you write fewer.

## End-to-end tests

An **end-to-end** (E2E) test drives the whole system the way a user would: start the Ledgerly web app, upload a file through HTTP, read the report page. It's the only kind of test that proves the entire thing works, but it's slow, has many moving parts that can fail for boring reasons, and when it fails you have to dig to find out why. Keep a few for the critical paths.

## Go doesn't make you choose names

Go has no separate "unit" and "integration" test types. They're all `func TestXxx(t *testing.T)`. Teams usually separate them in one of two ways:

- **`testing.Short()`**: slow tests call `t.Skip` when you run `go test -short`. You'll meet this in the next chapter.
- **Build tags**: put integration tests in files starting with `//go:build integration` and run them with `go test -tags integration ./...`.

## A pyramid, not a law

The shape is a guideline. A library like Ledgerly is almost all unit tests. A service that mostly glues a database to HTTP might have more integration tests, because that's where its bugs live. The real rule underneath is: **test each behaviour at the lowest level that can catch its bugs**, because lower means faster and more precise.

## Further reading

- [Martin Fowler: The Practical Test Pyramid](https://martinfowler.com/articles/practical-test-pyramid.html)
