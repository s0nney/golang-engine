---
title: Monthly Invoices
difficulty: hard
after: errors
hints:
  - 'Split the job in two passes. First walk the usage rows, validate each one and add its messages to a per-customer running total. Only once every row is valid, turn the totals into invoices.'
  - 'Keep the invoices in first-appearance order with a slice of customer names plus maps from name to total messages and name to plan. Searching the slice for each row''s customer works but is O(n²): the large test has 100,000 rows.'
  - 'On a bad row, return `nil, &BillingError{Row: i, Customer: u.Customer, Err: ...}`. For an unknown plan, put the plan name in the inner error with `fmt.Errorf("%w: %q", ErrUnknownPlan, u.Plan)`, so `errors.Is` still finds `ErrUnknownPlan`.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    type Plan struct {
    	MonthlyCents int // flat fee
    	Included     int // messages included in the fee
    	OverageCents int // price of each message beyond Included
    }

    type Usage struct {
    	Customer string
    	Plan     string
    	Messages int
    }

    type Invoice struct {
    	Customer   string
    	Messages   int
    	TotalCents int
    }

    var (
    	ErrBadUsage     = errors.New("invalid usage row")
    	ErrUnknownPlan  = errors.New("unknown plan")
    	ErrPlanMismatch = errors.New("customer is on two plans")
    )

    // BillingError reports which usage row could not be billed, and why.
    type BillingError struct {
    	Row      int    // index into the usage slice
    	Customer string // the row's Customer, as written
    	Err      error  // the reason, wrapping one of the Err... sentinels
    }

    func (e *BillingError) Error() string {
    	return fmt.Sprintf("row %d (%q): %v", e.Row, e.Customer, e.Err)
    }

    // Unwrap lets errors.Is and errors.AsType look inside a BillingError
    // and find the sentinel in e.Err. You don't need to change it.
    func (e *BillingError) Unwrap() error {
    	return e.Err
    }

    func BillAll(plans map[string]Plan, usage []Usage) ([]Invoice, error) {
    	return nil, nil
    }

    func main() {
    	plans := map[string]Plan{
    		"starter": {MonthlyCents: 500, Included: 100, OverageCents: 4},
    		"pro":     {MonthlyCents: 2000, Included: 1000, OverageCents: 2},
    	}
    	invoices, err := BillAll(plans, []Usage{
    		{"Mia", "starter", 80},
    		{"Sam", "pro", 1500},
    		{"Mia", "starter", 70},
    	})
    	fmt.Println(invoices, err)

    	_, err = BillAll(plans, []Usage{{"Ana", "gold", 10}})
    	fmt.Println(err)
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strings"
    )

    type Plan struct {
    	MonthlyCents int // flat fee
    	Included     int // messages included in the fee
    	OverageCents int // price of each message beyond Included
    }

    type Usage struct {
    	Customer string
    	Plan     string
    	Messages int
    }

    type Invoice struct {
    	Customer   string
    	Messages   int
    	TotalCents int
    }

    var (
    	ErrBadUsage     = errors.New("invalid usage row")
    	ErrUnknownPlan  = errors.New("unknown plan")
    	ErrPlanMismatch = errors.New("customer is on two plans")
    )

    // BillingError reports which usage row could not be billed, and why.
    type BillingError struct {
    	Row      int    // index into the usage slice
    	Customer string // the row's Customer, as written
    	Err      error  // the reason, wrapping one of the Err... sentinels
    }

    func (e *BillingError) Error() string {
    	return fmt.Sprintf("row %d (%q): %v", e.Row, e.Customer, e.Err)
    }

    // Unwrap lets errors.Is and errors.AsType look inside a BillingError
    // and find the sentinel in e.Err. You don't need to change it.
    func (e *BillingError) Unwrap() error {
    	return e.Err
    }

    func BillAll(plans map[string]Plan, usage []Usage) ([]Invoice, error) {
    	var order []string
    	planOf := map[string]string{}
    	messages := map[string]int{}

    	for i, u := range usage {
    		var reason error
    		if strings.TrimSpace(u.Customer) == "" {
    			reason = fmt.Errorf("%w: no customer", ErrBadUsage)
    		} else if u.Messages < 0 {
    			reason = fmt.Errorf("%w: %d messages", ErrBadUsage, u.Messages)
    		} else if _, ok := plans[u.Plan]; !ok {
    			reason = fmt.Errorf("%w: %q", ErrUnknownPlan, u.Plan)
    		} else if p, seen := planOf[u.Customer]; seen && p != u.Plan {
    			reason = fmt.Errorf("%w: %q and %q", ErrPlanMismatch, p, u.Plan)
    		}
    		if reason != nil {
    			return nil, &BillingError{Row: i, Customer: u.Customer, Err: reason}
    		}

    		if _, seen := planOf[u.Customer]; !seen {
    			order = append(order, u.Customer)
    			planOf[u.Customer] = u.Plan
    		}
    		messages[u.Customer] += u.Messages
    	}

    	invoices := make([]Invoice, 0, len(order))
    	for _, c := range order {
    		p := plans[planOf[c]]
    		n := messages[c]
    		extra := max(0, n-p.Included)
    		invoices = append(invoices, Invoice{
    			Customer:   c,
    			Messages:   n,
    			TotalCents: p.MonthlyCents + extra*p.OverageCents,
    		})
    	}
    	return invoices, nil
    }

    func main() {
    	plans := map[string]Plan{
    		"starter": {MonthlyCents: 500, Included: 100, OverageCents: 4},
    		"pro":     {MonthlyCents: 2000, Included: 1000, OverageCents: 2},
    	}
    	invoices, err := BillAll(plans, []Usage{
    		{"Mia", "starter", 80},
    		{"Sam", "pro", 1500},
    		{"Mia", "starter", 70},
    	})
    	fmt.Println(invoices, err)

    	_, err = BillAll(plans, []Usage{{"Ana", "gold", 10}})
    	fmt.Println(err)
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"slices"
    	"strings"
    	"testing"
    	"time"
    )

    var testPlans = map[string]Plan{
    	"starter": {MonthlyCents: 500, Included: 100, OverageCents: 4},
    	"pro":     {MonthlyCents: 2000, Included: 1000, OverageCents: 2},
    	"payg":    {MonthlyCents: 0, Included: 0, OverageCents: 3},
    }

    func TestBillAll(t *testing.T) {
    	tests := []struct {
    		name  string
    		usage []Usage
    		want  []Invoice
    	}{
    		{"example", []Usage{{"Mia", "starter", 80}, {"Sam", "pro", 1500}, {"Mia", "starter", 70}},
    			[]Invoice{{"Mia", 150, 700}, {"Sam", 1500, 3000}}},
    		{"no usage", nil, nil},
    		{"zero messages still pays the fee", []Usage{{"Ana", "pro", 0}}, []Invoice{{"Ana", 0, 2000}}},
    		{"exactly the included amount", []Usage{{"Ana", "starter", 100}}, []Invoice{{"Ana", 100, 500}}},
    		{"one over", []Usage{{"Ana", "starter", 101}}, []Invoice{{"Ana", 101, 504}}},
    		{"pay as you go", []Usage{{"Bo", "payg", 7}}, []Invoice{{"Bo", 7, 21}}},
    		{"first appearance order", []Usage{{"Zed", "payg", 1}, {"Amy", "payg", 1}, {"Zed", "payg", 1}, {"Kim", "payg", 1}},
    			[]Invoice{{"Zed", 2, 6}, {"Amy", 1, 3}, {"Kim", 1, 3}}},
    		{"names are exact", []Usage{{"sam", "payg", 1}, {"Sam", "pro", 1}},
    			[]Invoice{{"sam", 1, 3}, {"Sam", 1, 2000}}},
    		{"unicode names", []Usage{{"Zoë 🐝", "payg", 2}, {"Zoë 🐝", "payg", 3}}, []Invoice{{"Zoë 🐝", 5, 15}}},
    	}
    	for _, tt := range tests {
    		got, err := BillAll(testPlans, tt.usage)
    		if err != nil {
    			t.Errorf("%s: BillAll(%v) returned error %v, want nil", tt.name, tt.usage, err)
    			continue
    		}
    		if len(got) != len(tt.want) || (len(got) > 0 && !slices.Equal(got, tt.want)) {
    			t.Errorf("%s: BillAll(%v) = %+v, want %+v", tt.name, tt.usage, got, tt.want)
    		}
    	}
    }

    func TestBillAllErrors(t *testing.T) {
    	tests := []struct {
    		name     string
    		usage    []Usage
    		sentinel error
    		row      int
    		customer string
    		mention  string
    	}{
    		{"unknown plan", []Usage{{"Mia", "starter", 1}, {"Ana", "gold", 10}}, ErrUnknownPlan, 1, "Ana", "gold"},
    		{"empty plan", []Usage{{"Ana", "", 10}}, ErrUnknownPlan, 0, "Ana", ""},
    		{"negative messages", []Usage{{"Ana", "pro", -1}}, ErrBadUsage, 0, "Ana", ""},
    		{"blank customer", []Usage{{"Mia", "pro", 1}, {"Sam", "pro", 1}, {"  ", "pro", 1}}, ErrBadUsage, 2, "  ", ""},
    		{"empty customer", []Usage{{"", "pro", 1}}, ErrBadUsage, 0, "", ""},
    		{"plan changes", []Usage{{"Mia", "starter", 1}, {"Sam", "pro", 1}, {"Mia", "pro", 1}}, ErrPlanMismatch, 2, "Mia", ""},
    		{"first bad row wins", []Usage{{"Mia", "pro", 1}, {"Sam", "pro", -4}, {"Ana", "gold", 1}}, ErrBadUsage, 1, "Sam", ""},
    		{"unknown plan on the first row", []Usage{{"Ana", "gold", 1}, {"Mia", "pro", -4}}, ErrUnknownPlan, 0, "Ana", "gold"},
    	}
    	for _, tt := range tests {
    		got, err := BillAll(testPlans, tt.usage)
    		if err == nil {
    			t.Errorf("%s: BillAll(%v) returned no error, want one wrapping %q", tt.name, tt.usage, tt.sentinel)
    			continue
    		}
    		if got != nil {
    			t.Errorf("%s: BillAll(%v) returned invoices %+v together with an error, want nil invoices", tt.name, tt.usage, got)
    		}
    		if !errors.Is(err, tt.sentinel) {
    			t.Errorf("%s: BillAll(%v) error = %v, want one wrapping %q", tt.name, tt.usage, err, tt.sentinel)
    		}
    		be, ok := errors.AsType[*BillingError](err)
    		if !ok {
    			t.Errorf("%s: BillAll(%v) error = %v, want a *BillingError", tt.name, tt.usage, err)
    			continue
    		}
    		if be.Row != tt.row || be.Customer != tt.customer {
    			t.Errorf("%s: BillAll(%v) BillingError has Row %d, Customer %q, want Row %d, Customer %q", tt.name, tt.usage, be.Row, be.Customer, tt.row, tt.customer)
    		}
    		if tt.mention != "" && !strings.Contains(err.Error(), tt.mention) {
    			t.Errorf("%s: BillAll(%v) error %q should mention the plan %q", tt.name, tt.usage, err.Error(), tt.mention)
    		}
    	}
    }

    func TestBillAllLarge(t *testing.T) {
    	const customers = 50_000
    	usage := make([]Usage, 0, 2*customers)
    	for round := range 2 {
    		for i := range customers {
    			usage = append(usage, Usage{fmt.Sprintf("cust-%05d", i), "payg", round + 1})
    		}
    	}
    	type result struct {
    		invoices []Invoice
    		err      error
    	}
    	done := make(chan result, 1)
    	go func() {
    		invoices, err := BillAll(testPlans, usage)
    		done <- result{invoices, err}
    	}()
    	var r result
    	select {
    	case r = <-done:
    	case <-time.After(time.Second):
    		t.Fatalf("BillAll(100,000 rows) took over a second: look customers up in a map instead of searching a slice")
    	}
    	if r.err != nil {
    		t.Fatalf("BillAll(100,000 valid rows) returned error %v", r.err)
    	}
    	got := r.invoices
    	if len(got) != customers || got[0] != (Invoice{"cust-00000", 3, 9}) || got[customers-1] != (Invoice{"cust-49999", 3, 9}) {
    		t.Errorf("BillAll(100,000 rows for 50,000 customers) gave %d invoices, want %d, each with 3 messages and 9 cents, in first-appearance order", len(got), customers)
    	}
    }
---

It's the end of the month, and Textio needs to bill its customers. The usage
log has one row per sending batch, so a customer can appear many times.

Write `BillAll(plans, usage)`, which returns one `Invoice` per customer.

**Billing.** Add up each customer's messages across all their rows. Their
invoice total is their plan's `MonthlyCents`, plus `OverageCents` for every
message beyond the plan's `Included` messages. Invoices come out in the order
in which each customer **first appears** in `usage`. Customer names are
compared exactly (`"sam"` and `"Sam"` are two customers).

**Validation.** A row is bad if:

| Problem | Sentinel to wrap |
| --- | --- |
| `Customer` is empty or only whitespace | `ErrBadUsage` |
| `Messages` is negative | `ErrBadUsage` |
| `Plan` isn't a key in `plans` (the message must name the plan) | `ErrUnknownPlan` |
| the customer already appeared with a **different** plan | `ErrPlanMismatch` |

Check the rules in that order. At the **first** bad row, return `nil`
invoices and a `*BillingError` holding the row's index, its `Customer`
exactly as written, and an `Err` that wraps the right sentinel. The
`BillingError` type is already written for you.

## Example

```go
plans := map[string]Plan{
	"starter": {MonthlyCents: 500, Included: 100, OverageCents: 4},
	"pro":     {MonthlyCents: 2000, Included: 1000, OverageCents: 2},
}
BillAll(plans, []Usage{{"Mia", "starter", 80}, {"Sam", "pro", 1500}, {"Mia", "starter", 70}})
// [{Mia 150 700} {Sam 1500 3000}], nil
//   Mia: 500 + (150-100)×4 = 700    Sam: 2000 + (1500-1000)×2 = 3000

BillAll(plans, []Usage{{"Ana", "gold", 10}})
// nil, row 0 ("Ana"): unknown plan: "gold"
```

## Constraints

- A customer with 0 messages still pays the monthly fee. No usage at all
  gives an empty (or `nil`) invoice list and no error.
- Up to 100,000 rows and 50,000 customers. One test bills that many under a
  one-second limit, so don't search the invoice list for every row.
