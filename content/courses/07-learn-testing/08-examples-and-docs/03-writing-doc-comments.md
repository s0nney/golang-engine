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
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"go/ast"
    	"go/doc/comment"
    	"go/parser"
    	"go/token"
    )

    // Cents is an amount of money in hundredths of a dollar.
    type Cents int64

    // ErrSameAccount is returned by [Ledger.Transfer] when from and to are
    // the same account.
    var ErrSameAccount = errors.New("ledgerly: transfer to the same account")

    // ErrNotPositive is returned by [Ledger.Transfer] when amount is zero or
    // negative.
    var ErrNotPositive = errors.New("ledgerly: transfer amount must be positive")

    // Ledger holds a balance for each account.
    type Ledger struct {
    	balances map[string]Cents
    }

    // NewLedger returns an empty Ledger.
    func NewLedger() *Ledger {
    	return &Ledger{balances: map[string]Cents{}}
    }

    // Balance returns the balance of account. Unknown accounts have a zero
    // balance.
    func (l *Ledger) Balance(account string) Cents {
    	return l.balances[account]
    }

    // Transfer moves amount from one account to another. It fails if:
    // - from and to are the same account
    // - amount is not positive
    // For example, to move $50 from cash into savings:
    // err := l.Transfer("cash", "savings", 5000)
    // The errors it returns are ErrSameAccount and ErrNotPositive.
    func (l *Ledger) Transfer(from, to string, amount Cents) error {
    	// ?
    	return nil
    }

    // showDoc prints Transfer's doc comment the way go doc would show it.
    func showDoc() {
    	f, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, parser.ParseComments)
    	if err != nil {
    		fmt.Println("(can't read main.go here, so no doc preview)")
    		return
    	}
    	for _, d := range f.Decls {
    		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "Transfer" && fn.Doc != nil {
    			// Treat every [Name] as a doc link to something that exists.
    			p := comment.Parser{LookupSym: func(recv, name string) bool { return true }}
    			fmt.Printf("%s", new(comment.Printer).Text(p.Parse(fn.Doc.Text())))
    		}
    	}
    }

    func main() {
    	l := NewLedger()
    	fmt.Println(l.Transfer("cash", "savings", 5000), l.Balance("cash"), l.Balance("savings"))
    	fmt.Println(l.Transfer("cash", "cash", 100))
    	fmt.Println(l.Transfer("cash", "savings", 0))
    	fmt.Println("--- go doc Ledger.Transfer ---")
    	showDoc()
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"go/ast"
    	"go/doc/comment"
    	"go/parser"
    	"go/token"
    )

    // Cents is an amount of money in hundredths of a dollar.
    type Cents int64

    // ErrSameAccount is returned by [Ledger.Transfer] when from and to are
    // the same account.
    var ErrSameAccount = errors.New("ledgerly: transfer to the same account")

    // ErrNotPositive is returned by [Ledger.Transfer] when amount is zero or
    // negative.
    var ErrNotPositive = errors.New("ledgerly: transfer amount must be positive")

    // Ledger holds a balance for each account.
    type Ledger struct {
    	balances map[string]Cents
    }

    // NewLedger returns an empty Ledger.
    func NewLedger() *Ledger {
    	return &Ledger{balances: map[string]Cents{}}
    }

    // Balance returns the balance of account. Unknown accounts have a zero
    // balance.
    func (l *Ledger) Balance(account string) Cents {
    	return l.balances[account]
    }

    // Transfer moves amount from one account to another. It fails if:
    //   - from and to are the same account
    //   - amount is not positive
    //
    // For example, to move $50 from cash into savings:
    //
    //	err := l.Transfer("cash", "savings", 5000)
    //
    // The errors it returns are [ErrSameAccount] and [ErrNotPositive].
    func (l *Ledger) Transfer(from, to string, amount Cents) error {
    	if from == to {
    		return ErrSameAccount
    	}
    	if amount <= 0 {
    		return ErrNotPositive
    	}
    	l.balances[from] -= amount
    	l.balances[to] += amount
    	return nil
    }

    // showDoc prints Transfer's doc comment the way go doc would show it.
    func showDoc() {
    	f, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, parser.ParseComments)
    	if err != nil {
    		fmt.Println("(can't read main.go here, so no doc preview)")
    		return
    	}
    	for _, d := range f.Decls {
    		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "Transfer" && fn.Doc != nil {
    			// Treat every [Name] as a doc link to something that exists.
    			p := comment.Parser{LookupSym: func(recv, name string) bool { return true }}
    			fmt.Printf("%s", new(comment.Printer).Text(p.Parse(fn.Doc.Text())))
    		}
    	}
    }

    func main() {
    	l := NewLedger()
    	fmt.Println(l.Transfer("cash", "savings", 5000), l.Balance("cash"), l.Balance("savings"))
    	fmt.Println(l.Transfer("cash", "cash", 100))
    	fmt.Println(l.Transfer("cash", "savings", 0))
    	fmt.Println("--- go doc Ledger.Transfer ---")
    	showDoc()
    }
  tests: |
    package main

    import (
    	"errors"
    	"go/ast"
    	"go/doc/comment"
    	"go/parser"
    	"go/token"
    	"strings"
    	"testing"
    )

    func TestTransfer(t *testing.T) {
    	l := NewLedger()
    	if err := l.Transfer("cash", "savings", 5000); err != nil {
    		t.Fatalf(`Transfer("cash", "savings", 5000) = %v, want nil`, err)
    	}
    	if err := l.Transfer("savings", "rent", 1200); err != nil {
    		t.Fatalf(`Transfer("savings", "rent", 1200) = %v, want nil`, err)
    	}
    	for account, want := range map[string]Cents{"cash": -5000, "savings": 3800, "rent": 1200} {
    		if got := l.Balance(account); got != want {
    			t.Errorf("after two transfers, Balance(%q) = %d, want %d", account, got, want)
    		}
    	}
    	for _, tc := range []struct {
    		from, to string
    		amount   Cents
    		want     error
    	}{
    		{"cash", "cash", 100, ErrSameAccount},
    		{"cash", "savings", 0, ErrNotPositive},
    		{"cash", "savings", -100, ErrNotPositive},
    	} {
    		if err := l.Transfer(tc.from, tc.to, tc.amount); !errors.Is(err, tc.want) {
    			t.Errorf("Transfer(%q, %q, %d) = %v, want %v", tc.from, tc.to, tc.amount, err, tc.want)
    		}
    	}
    	if got := l.Balance("cash"); got != -5000 {
    		t.Errorf("failed transfers changed Balance(\"cash\") to %d; a failed transfer must not move any money", got)
    	}
    }

    // transferDoc parses main.go and returns Transfer's doc comment, parsed the
    // way go doc parses it.
    func transferDoc(t *testing.T) *comment.Doc {
    	t.Helper()
    	f, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, parser.ParseComments)
    	if err != nil {
    		t.Fatalf("parsing main.go: %v", err)
    	}
    	names := map[string]bool{}
    	for _, d := range f.Decls {
    		switch d := d.(type) {
    		case *ast.GenDecl:
    			for _, s := range d.Specs {
    				switch s := s.(type) {
    				case *ast.ValueSpec:
    					for _, n := range s.Names {
    						names[n.Name] = true
    					}
    				case *ast.TypeSpec:
    					names[s.Name.Name] = true
    				}
    			}
    		case *ast.FuncDecl:
    			names[d.Name.Name] = true
    		}
    	}
    	for _, d := range f.Decls {
    		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "Transfer" {
    			if fn.Doc == nil {
    				t.Fatal("Transfer has no doc comment")
    			}
    			p := comment.Parser{LookupSym: func(recv, name string) bool { return names[name] }}
    			return p.Parse(fn.Doc.Text())
    		}
    	}
    	t.Fatal("Transfer not found in main.go")
    	return nil
    }

    func plain(text []comment.Text) string {
    	var b strings.Builder
    	for _, t := range text {
    		switch t := t.(type) {
    		case comment.Plain:
    			b.WriteString(string(t))
    		case comment.Italic:
    			b.WriteString(string(t))
    		case *comment.DocLink:
    			b.WriteString(plain(t.Text))
    		case *comment.Link:
    			b.WriteString(plain(t.Text))
    		}
    	}
    	return b.String()
    }

    func TestTransferDocStartsWithName(t *testing.T) {
    	d := transferDoc(t)
    	if len(d.Content) == 0 {
    		t.Fatal("Transfer's doc comment is empty")
    	}
    	p, ok := d.Content[0].(*comment.Paragraph)
    	if !ok || !strings.HasPrefix(plain(p.Text), "Transfer ") {
    		t.Error(`Transfer's doc comment should start with a paragraph beginning "Transfer "`)
    	}
    }

    func TestTransferDocList(t *testing.T) {
    	d := transferDoc(t)
    	for _, b := range d.Content {
    		if l, ok := b.(*comment.List); ok {
    			if len(l.Items) != 2 {
    				t.Fatalf("Transfer's doc comment has a list with %d items, want 2 (the two ways Transfer fails)", len(l.Items))
    			}
    			return
    		}
    	}
    	t.Error("Transfer's doc comment has no list: go doc shows the \"- ...\" lines as part of the paragraph. Indent them: //   - from and to are the same account")
    }

    func TestTransferDocCodeBlock(t *testing.T) {
    	d := transferDoc(t)
    	for _, b := range d.Content {
    		if c, ok := b.(*comment.Code); ok {
    			if !strings.Contains(c.Text, `l.Transfer("cash", "savings", 5000)`) {
    				t.Errorf("Transfer's doc comment has a code block, but it doesn't contain the example call:\n%s", c.Text)
    			}
    			return
    		}
    	}
    	t.Error("Transfer's doc comment has no code block: the example call is re-wrapped into the paragraph. Put a blank // line before it and indent it with a tab: //\terr := l.Transfer(...)")
    }

    func TestTransferDocLinks(t *testing.T) {
    	d := transferDoc(t)
    	found := map[string]bool{}
    	var walk func([]comment.Text)
    	walk = func(text []comment.Text) {
    		for _, x := range text {
    			if dl, ok := x.(*comment.DocLink); ok {
    				found[dl.Name] = true
    			}
    		}
    	}
    	for _, b := range d.Content {
    		switch b := b.(type) {
    		case *comment.Paragraph:
    			walk(b.Text)
    		case *comment.List:
    			for _, it := range b.Items {
    				for _, ib := range it.Content {
    					if p, ok := ib.(*comment.Paragraph); ok {
    						walk(p.Text)
    					}
    				}
    			}
    		}
    	}
    	for _, name := range []string{"ErrSameAccount", "ErrNotPositive"} {
    		if !found[name] {
    			t.Errorf("Transfer's doc comment doesn't link to %s: write it as a doc link, [%s]", name, name)
    		}
    	}
    }
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

## Your turn: fix Transfer's docs

`Ledger.Transfer` in the editor has two problems. It doesn't do anything yet, and its doc comment was written by someone who'd never seen the doc comment syntax. `go doc` squashes the whole comment into one paragraph.

1. **Implement `Transfer`.** Return `ErrSameAccount` if `from == to`, and `ErrNotPositive` if `amount <= 0`, without changing any balance. Otherwise subtract `amount` from `from` and add it to `to`.
2. **Fix its doc comment**, keeping the same words:
   - the two failure conditions become a real **list**;
   - the example call `err := l.Transfer("cash", "savings", 5000)` becomes a **code block**;
   - `ErrSameAccount` and `ErrNotPositive` become **doc links**.

The grader parses your comment with [`go/doc/comment`](https://pkg.go.dev/go/doc/comment), the same parser `go doc` and pkg.go.dev use, and checks for a two-item list, a code block containing the call, and both doc links. **Run** shows the transfers, then your comment rendered the way `go doc` would print it. Watch it change shape as you fix it.

## Further reading

- [Go Doc Comments](https://go.dev/doc/comment), the full reference for the syntax.
