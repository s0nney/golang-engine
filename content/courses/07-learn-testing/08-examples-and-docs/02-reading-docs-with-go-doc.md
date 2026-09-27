---
title: Reading Docs with go doc
quiz:
  - question: Which command shows the full source of the `Cut` function in the `strings` package?
    options:
      - text: '`go doc strings.Cut`'
      - text: '`go doc -src strings.Cut`'
        correct: true
      - text: '`go doc -all strings`'
      - text: '`go source strings.Cut`'
    explanation: |
      `-src` prints the declaration and body. Plain `go doc strings.Cut`
      shows only the signature and doc comment; `-all` shows every doc in
      the package.
  - question: |
      `go doc ledgerly` lists `func ParseAmount(s string) (Cents, error)`
      indented under `type Cents int64`. Why?
    options:
      - text: '`ParseAmount` is a method on `Cents`'
      - text: It returns a `Cents`, so go doc treats it like a constructor and groups it with the type
        correct: true
      - text: It's declared in the same file as `Cents`
      - text: It's a mistake in the package
    explanation: |
      Functions that return a package type (or a pointer to one) are listed
      with that type, which is how constructors like `NewLedger` end up next
      to `Ledger`.
---

Every API in this course was checked with `go doc`, and it should become one of your most-used commands. It reads documentation straight from source code, for the standard library, your dependencies and your own packages, without leaving the terminal.

## The basics

```text
go doc strings                 # package overview: doc comment and a list of symbols
go doc strings.Cut             # one function
go doc bytes.Buffer            # a type, with its methods listed
go doc bytes.Buffer.WriteString
go doc testing.B.Loop          # a method
go doc json.Decoder            # short package names work when they're unambiguous
```

Inside a module, `go doc` with no arguments documents the current package, and a bare capitalised name looks in it: `go doc ParseAmount`.

Lower-case letters match either case, so `go doc testing.b.loop` works too. Useful when you can't remember the capitalisation.

## Useful flags

| Flag | Shows |
| --- | --- |
| `-all` | every doc comment in the package, in full |
| `-short` | one line per symbol |
| `-src` | the full source of a symbol, including unexported details |
| `-u` | unexported symbols, methods and fields too |
| `-ex` | the executable examples for a symbol |
| `-cmd` | the exported symbols of a `package main` |
| `-http` | starts a local web server with rendered HTML docs |

`go doc -src` is underrated. When the docs don't answer a question ("does `slices.Sort` allocate?"), reading the source takes ten seconds.

## Reading your own package

Here's `go doc` on Ledgerly, with a package comment written in the next lesson's style:

```text
$ go doc .
package ledgerly // import "ledgerly"

Package ledgerly keeps track of money: accounts, transactions and balances,
with amounts stored as integer Cents.

# Amounts

Amounts are never floating point. Use ParseAmount to read one and Cents.String
to show one:

    c, err := ledgerly.ParseAmount("12.34")
    fmt.Println(c) // $12.34

var ErrBadAmount = errors.New("ledgerly: bad amount")
func WriteStatement(w io.Writer, txns []Transaction) error
type Cents int64
    func ParseAmount(s string) (Cents, error)
type Transaction struct{ ... }
    func ImportCSV(r io.Reader) ([]Transaction, error)
```

Two things to notice:

- Functions that return one of the package's types are grouped under that type, like constructors. `ParseAmount` returns `Cents`, so it's listed under `Cents`.
- Only exported names appear. That listing *is* your API. If something shows up that users shouldn't depend on, unexport it before someone does.

Running `go doc` on your own package before a release is a quick review: does every exported name have a comment? Does the overview make sense to someone who's never seen the code?

## pkg.go.dev

The same comments power [pkg.go.dev](https://pkg.go.dev), which renders documentation for every public module, with the examples runnable in the browser. `go doc -http` gives you a local preview of what your package will look like there before you publish.

## go doc versus reading the code

`go doc` answers "how do I use this?". When you need "how does this work?", use `-src`, or jump to the definition in your editor. When you need "what changed?", read the release notes: `go doc` shows the current version only.
