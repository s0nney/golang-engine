---
title: 'Your Turn: Docs That Run'
quiz:
  - question: |
      This example's output comment is right, but `go test` reports it
      as failing with `got:` empty. What's the most likely cause?

      ```go
      func ExampleLedger_WriteStatement_empty() {
          var buf bytes.Buffer
          NewLedger().WriteStatement(&buf)
          // Output: no transactions
      }
      ```
    options:
      - text: Examples can't use `bytes.Buffer`
      - text: The statement went into `buf`, and only what's printed to standard output is compared
        correct: true
      - text: The suffix `_empty` must be capitalised
      - text: '`// Output:` must be on its own line'
    explanation: |
      An example's output is whatever it writes to `os.Stdout`. Write to
      `os.Stdout` directly, or print the buffer's contents at the end.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"os"
    )

    // Cents is an amount of money in hundredths of a dollar.
    // Negative amounts are debits.
    type Cents int64

    // String formats c like "$12.34" or "-$0.05".
    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    // Transaction is one entry in a Ledger.
    type Transaction struct {
    	Account string
    	Amount  Cents
    }

    type Ledger struct {
    	txns []Transaction
    }

    func NewLedger() *Ledger {
    	return &Ledger{}
    }

    // adds a transaction
    func (l *Ledger) Add(account string, amount Cents) {
    	l.txns = append(l.txns, Transaction{Account: account, Amount: amount})
    }

    // This returns the balance of one account.
    func (l *Ledger) Balance(account string) Cents {
    	var total Cents
    	for _, tx := range l.txns {
    		if tx.Account == account {
    			total += tx.Amount
    		}
    	}
    	return total
    }

    func (l *Ledger) Total() Cents {
    	var total Cents
    	for _, tx := range l.txns {
    		total += tx.Amount
    	}
    	return total
    }

    func (l *Ledger) WriteStatement(w io.Writer) error {
    	// ?
    	return nil
    }

    func main() {
    	l := NewLedger()
    	l.Add("rent", -120000)
    	l.Add("salary", 350000)
    	l.Add("groceries", -8734)
    	if err := l.WriteStatement(os.Stdout); err != nil {
    		fmt.Println("error:", err)
    	}
    	fmt.Println("---")
    	NewLedger().WriteStatement(os.Stdout)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"io"
    	"os"
    )

    // Cents is an amount of money in hundredths of a dollar.
    // Negative amounts are debits.
    type Cents int64

    // String formats c like "$12.34" or "-$0.05".
    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    // Transaction is one entry in a Ledger.
    type Transaction struct {
    	Account string
    	Amount  Cents
    }

    // Ledger records transactions in the order they were added.
    // The zero value is not ready to use; call [NewLedger].
    type Ledger struct {
    	txns []Transaction
    }

    // NewLedger returns an empty ledger.
    func NewLedger() *Ledger {
    	return &Ledger{}
    }

    // Add records a transaction of amount against account.
    func (l *Ledger) Add(account string, amount Cents) {
    	l.txns = append(l.txns, Transaction{Account: account, Amount: amount})
    }

    // Balance returns the sum of all transactions for account.
    // An account with no transactions has a zero balance.
    func (l *Ledger) Balance(account string) Cents {
    	var total Cents
    	for _, tx := range l.txns {
    		if tx.Account == account {
    			total += tx.Amount
    		}
    	}
    	return total
    }

    // Total returns the sum of every transaction in the ledger.
    //
    // Deprecated: Use [Ledger.Balance] for one account. Total mixes up
    // accounts and is rarely what you want.
    func (l *Ledger) Total() Cents {
    	var total Cents
    	for _, tx := range l.txns {
    		total += tx.Amount
    	}
    	return total
    }

    // WriteStatement writes one line per transaction to w, numbered from 1
    // and showing the running balance, or "no transactions" for an empty
    // ledger. It returns the first write error.
    func (l *Ledger) WriteStatement(w io.Writer) error {
    	if len(l.txns) == 0 {
    		_, err := fmt.Fprintln(w, "no transactions")
    		return err
    	}
    	var running Cents
    	for i, tx := range l.txns {
    		running += tx.Amount
    		if _, err := fmt.Fprintf(w, "#%d %-10s %10v %10v\n", i+1, tx.Account, tx.Amount, running); err != nil {
    			return err
    		}
    	}
    	return nil
    }

    func main() {
    	l := NewLedger()
    	l.Add("rent", -120000)
    	l.Add("salary", 350000)
    	l.Add("groceries", -8734)
    	if err := l.WriteStatement(os.Stdout); err != nil {
    		fmt.Println("error:", err)
    	}
    	fmt.Println("---")
    	NewLedger().WriteStatement(os.Stdout)
    }
  tests: |
    package main

    import (
    	"fmt"
    	"go/ast"
    	"go/parser"
    	"go/token"
    	"os"
    	"strings"
    	"testing"
    )

    func ExampleLedger_WriteStatement() {
    	l := NewLedger()
    	l.Add("rent", -120000)
    	l.Add("cash", 5000)
    	l.Add("rent", -1500)
    	l.Add("groceries", -8734)
    	l.WriteStatement(os.Stdout)
    	// Output:
    	// #1 rent        -$1200.00  -$1200.00
    	// #2 cash           $50.00  -$1150.00
    	// #3 rent          -$15.00  -$1165.00
    	// #4 groceries     -$87.34  -$1252.34
    }

    func ExampleLedger_WriteStatement_empty() {
    	NewLedger().WriteStatement(os.Stdout)
    	// Output: no transactions
    }

    func ExampleLedger_Balance() {
    	l := NewLedger()
    	l.Add("rent", -120000)
    	l.Add("cash", 5000)
    	l.Add("rent", -1500)
    	fmt.Println(l.Balance("rent"))
    	fmt.Println(l.Balance("savings"))
    	// Output:
    	// -$1215.00
    	// $0.00
    }

    // docs parses main.go and returns the doc comment text of each top-level
    // declaration, keyed by name ("Ledger", "NewLedger", "Ledger.Add", ...).
    func docs(t *testing.T) map[string]string {
    	t.Helper()
    	f, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, parser.ParseComments)
    	if err != nil {
    		t.Fatalf("parsing main.go: %v", err)
    	}
    	m := map[string]string{}
    	for _, d := range f.Decls {
    		switch d := d.(type) {
    		case *ast.FuncDecl:
    			name := d.Name.Name
    			if d.Recv != nil && len(d.Recv.List) == 1 {
    				typ := d.Recv.List[0].Type
    				if star, ok := typ.(*ast.StarExpr); ok {
    					typ = star.X
    				}
    				if id, ok := typ.(*ast.Ident); ok {
    					name = id.Name + "." + name
    				}
    			}
    			m[name] = d.Doc.Text()
    		case *ast.GenDecl:
    			for _, s := range d.Specs {
    				if ts, ok := s.(*ast.TypeSpec); ok {
    					doc := ts.Doc
    					if doc == nil {
    						doc = d.Doc
    					}
    					m[ts.Name.Name] = doc.Text()
    				}
    			}
    		}
    	}
    	return m
    }

    func TestDocComments(t *testing.T) {
    	m := docs(t)
    	for _, name := range []string{"Cents", "Transaction", "Ledger", "NewLedger", "Ledger.Add", "Ledger.Balance", "Ledger.Total", "Ledger.WriteStatement"} {
    		doc, ok := m[name]
    		if !ok {
    			t.Errorf("%s isn't declared in main.go any more", name)
    			continue
    		}
    		short := name[strings.LastIndex(name, ".")+1:]
    		if doc == "" {
    			t.Errorf("%s has no doc comment", name)
    		} else if !strings.HasPrefix(doc, short+" ") {
    			t.Errorf("%s's doc comment starts %q; it should start with %q", name, firstLine(doc), short+" ")
    		}
    	}
    }

    func TestTotalDeprecated(t *testing.T) {
    	doc := docs(t)["Ledger.Total"]
    	var para string
    	for p := range strings.SplitSeq(doc, "\n\n") {
    		if strings.HasPrefix(p, "Deprecated: ") {
    			para = p
    		}
    	}
    	if para == "" {
    		t.Fatalf("Total's doc comment has no paragraph starting with \"Deprecated: \" (it must be its own paragraph, after a blank // line):\n%s", doc)
    	}
    	if !strings.Contains(para, "[Ledger.Balance]") {
    		t.Errorf("Total's Deprecated paragraph should point to the replacement with the doc link [Ledger.Balance]: %q", para)
    	}
    }

    func firstLine(s string) string {
    	line, _, _ := strings.Cut(s, "\n")
    	return line
    }
---

The grader for this exercise is mostly *documentation*: example functions whose output comments specify what your code must print, and a check that reads your doc comments. Documentation that fails the build when it's wrong is the whole point of this chapter.

## Part 1: make the examples pass

`WriteStatement` isn't written yet. Its specification is these two examples, which the grader runs:

```go
func ExampleLedger_WriteStatement() {
	l := NewLedger()
	l.Add("rent", -120000)
	l.Add("cash", 5000)
	l.Add("rent", -1500)
	l.Add("groceries", -8734)
	l.WriteStatement(os.Stdout)
	// Output:
	// #1 rent        -$1200.00  -$1200.00
	// #2 cash           $50.00  -$1150.00
	// #3 rent          -$15.00  -$1165.00
	// #4 groceries     -$87.34  -$1252.34
}

func ExampleLedger_WriteStatement_empty() {
	NewLedger().WriteStatement(os.Stdout)
	// Output: no transactions
}
```

Each line is `#`, the transaction number counting from 1, the account padded to 10 characters, the amount, and the **running balance** after that transaction. Work out a format string from the example; `%-10s` and `%10v` will get you most of the way. An empty ledger prints `no transactions`. Return the first error from writing, like the last few exercises.

## Part 2: document the API

Run `go doc` on this package in your head. `Ledger`, `NewLedger`, `Add`, `Balance`, `Total` and `WriteStatement` either have no comment or one that doesn't follow the conventions. Fix that:

1. Every one of them needs a doc comment whose first word is the name it documents (`Add records...`, not `adds a transaction` or `This returns...`). `Cents`, `String` and `Transaction` are already done; use them as a model.
2. `Total` adds up every account together, which is almost never what anyone wants. Mark it deprecated: after its normal description, add a blank `//` line and a paragraph that starts with `Deprecated: ` and points to the replacement with the doc link `[Ledger.Balance]`.

The grader parses `main.go` with `go/parser` (the same package `go doc` builds on) to check these. **Run** prints a statement for a small ledger and for an empty one.
