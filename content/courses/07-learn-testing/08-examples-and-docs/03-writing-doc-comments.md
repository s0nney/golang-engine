---
title: Writing Good Doc Comments
quiz:
  - question: |
      How does `go doc` render this comment?

      ```go
      // Transfer moves amount between accounts. It fails if:
      // - the accounts are the same
      // - amount is not positive
      func Transfer(...)
      ```
    options:
      - text: As a sentence followed by a two-item bulleted list
      - text: 'As one paragraph: "It fails if: - the accounts are the same - amount is not positive"'
        correct: true
      - text: As a code block
      - text: '`gofmt` rejects the file'
    explanation: |
      In Go doc comments a list must be **indented**. Unindented `-` lines
      are just more paragraph text and get re-wrapped. Write
      `//   - the accounts are the same` instead.
  - question: What's the conventional first word of the doc comment on `func NewLedger() *Ledger`?
    options:
      - text: '`This`'
      - text: '`Returns`'
      - text: '`NewLedger`'
        correct: true
      - text: '`Function`'
    explanation: |
      Doc comments are complete sentences that start with the name being
      declared: "NewLedger returns an empty ledger." Tools rely on this, and
      it reads well in `go doc`'s one-symbol output.
---

A doc comment is the comment directly above a top-level declaration, with no blank line in between. Every exported name should have one. Go's doc comments are plain text with a small, deliberately limited syntax, and `gofmt` knows it well enough to tidy some of it for you.

## Start with the name

A doc comment is one or more complete sentences, and the first starts with the name it documents:

```go
// Cents is an amount of money in hundredths of a dollar.
// Negative amounts are debits.
type Cents int64

// ParseAmount parses a decimal amount like "12.34" or "-0.05".
// It returns an error wrapping ErrBadAmount if s is malformed or too large.
func ParseAmount(s string) (Cents, error)

// Balance returns the sum of all transactions for account.
// An account with no transactions has a zero balance.
func (l *Ledger) Balance(account string) Cents
```

Say what it does and what the caller needs to know: units, edge cases, which errors come back, whether it's safe for concurrent use. Don't describe the implementation ("loops over the transactions"). That's what the code is for.

The package gets one comment too, above the `package` clause in one file (often `doc.go`), starting with "Package ledgerly ...".

## The syntax

**Paragraphs** are separated by blank `//` lines. Lines within a paragraph are re-wrapped, so line breaks don't matter.

**Headings** are a line starting with `# `, with blank lines around it:

```go
// # Amounts
```

**Code blocks** are indented lines (a tab after the `//`, which `gofmt` inserts for you). They're shown verbatim:

```go
// Use it like this:
//
//	c, err := ledgerly.ParseAmount("12.34")
//	fmt.Println(c) // $12.34
```

**Lists** must be indented too. Unindented `-` lines are just paragraph text, which is the most common mistake:

```go
// Transfer fails if:
//   - the accounts are the same
//   - amount is not positive
```

`go doc` shows the difference clearly:

```text
func Transfer()
    Transfer fails if:
      - the accounts are the same
      - amount is not positive
```

Numbered lists work the same way: `//  1. first step`.

**Doc links** put a Go name in square brackets: `[ParseAmount]`, `[Cents.String]`, `[io.Reader]`, `[encoding/csv.Reader]`. They become hyperlinks on pkg.go.dev, and `go doc` shows them as plain names.

**URL links** use a reference at the end of the comment:

```go
// ImportCSV reads a bank export in the [RFC 4180] CSV format.
//
// [RFC 4180]: https://www.rfc-editor.org/rfc/rfc4180
```

## Deprecated

When an API is being replaced, add a paragraph that begins with `Deprecated: `:

```go
// Total returns the sum of every transaction in the ledger.
//
// Deprecated: Use [Ledger.Balance] for one account, or [Ledger.Balances]
// for all of them. Total mixes up accounts and is rarely what you want.
func (l *Ledger) Total() Cents
```

pkg.go.dev hides deprecated names by default, editors strike them through, and linters such as staticcheck warn when code uses them. The paragraph must start with exactly `Deprecated: `.

## Let gofmt help

`gofmt` reformats doc comments: it indents code blocks with a tab, puts blank lines around them and around headings, and normalises list markers. Run it (your editor probably does on save), and if a comment changes shape unexpectedly, the syntax probably isn't what you meant.

## Checklist

- Every exported name has a comment starting with that name.
- The package has a "Package x ..." comment.
- Errors, units, edge cases and concurrency safety are mentioned where relevant.
- Examples for the important entry points (last lesson).
- `go doc` output reads well.

## Further reading

- [Go Doc Comments](https://go.dev/doc/comment), the full reference for the syntax.
