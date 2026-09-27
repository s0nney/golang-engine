---
title: Credit Transfer
difficulty: medium
after: pointers
hints:
  - 'Put all the "refuse" checks first as guard clauses, each returning `false`: a `nil` wallet, an amount that isn''t positive, the same wallet on both sides (`from == to` compares the pointers), or not enough credits.'
  - 'Order matters: check for `nil` **before** you read `from.Credits`, or a nil wallet will crash your program with a nil pointer dereference.'
  - 'In `topUpAll`, range over the slice and skip `nil` entries with `continue`. Because each element is a `*Wallet`, `w.Credits += amount` changes the real wallet, not a copy.'
exercise:
  starter: |
    package main

    import "fmt"

    type Wallet struct {
    	Owner   string
    	Credits int
    }

    func transfer(from, to *Wallet, amount int) bool {
    	return false
    }

    func topUpAll(wallets []*Wallet, amount int) int {
    	return 0
    }

    func main() {
    	mia := &Wallet{Owner: "Mia", Credits: 50}
    	sam := &Wallet{Owner: "Sam", Credits: 5}
    	fmt.Println(transfer(mia, sam, 20), *mia, *sam)
    	fmt.Println(topUpAll([]*Wallet{mia, nil, sam}, 10), *mia, *sam)
    }
  solution: |
    package main

    import "fmt"

    type Wallet struct {
    	Owner   string
    	Credits int
    }

    func transfer(from, to *Wallet, amount int) bool {
    	if from == nil || to == nil || from == to {
    		return false
    	}
    	if amount <= 0 || from.Credits < amount {
    		return false
    	}
    	from.Credits -= amount
    	to.Credits += amount
    	return true
    }

    func topUpAll(wallets []*Wallet, amount int) int {
    	if amount <= 0 {
    		return 0
    	}
    	count := 0
    	for _, w := range wallets {
    		if w == nil {
    			continue
    		}
    		w.Credits += amount
    		count++
    	}
    	return count
    }

    func main() {
    	mia := &Wallet{Owner: "Mia", Credits: 50}
    	sam := &Wallet{Owner: "Sam", Credits: 5}
    	fmt.Println(transfer(mia, sam, 20), *mia, *sam)
    	fmt.Println(topUpAll([]*Wallet{mia, nil, sam}, 10), *mia, *sam)
    }
  tests: |
    package main

    import "testing"

    func TestTransfer(t *testing.T) {
    	tests := []struct {
    		name                string
    		fromCredits, amount int
    		toCredits           int
    		want                bool
    		wantFrom, wantTo    int
    	}{
    		{"normal", 50, 20, 5, true, 30, 25},
    		{"everything", 20, 20, 0, true, 0, 20},
    		{"not enough", 10, 11, 0, false, 10, 0},
    		{"zero amount", 10, 0, 0, false, 10, 0},
    		{"negative amount", 10, -5, 0, false, 10, 0},
    	}
    	for _, tt := range tests {
    		from := &Wallet{Owner: "from", Credits: tt.fromCredits}
    		to := &Wallet{Owner: "to", Credits: tt.toCredits}
    		got := transfer(from, to, tt.amount)
    		if got != tt.want || from.Credits != tt.wantFrom || to.Credits != tt.wantTo {
    			t.Errorf("%s: transfer(%d credits -> %d credits, amount %d) = %v leaving %d and %d, want %v leaving %d and %d",
    				tt.name, tt.fromCredits, tt.toCredits, tt.amount, got, from.Credits, to.Credits, tt.want, tt.wantFrom, tt.wantTo)
    		}
    	}
    }

    func TestTransferNilAndSame(t *testing.T) {
    	w := &Wallet{Owner: "Mia", Credits: 50}
    	if transfer(nil, w, 10) || w.Credits != 50 {
    		t.Errorf("transfer(nil, wallet, 10) must return false and leave the wallet at 50 credits (got %d)", w.Credits)
    	}
    	if transfer(w, nil, 10) || w.Credits != 50 {
    		t.Errorf("transfer(wallet, nil, 10) must return false and leave the wallet at 50 credits (got %d)", w.Credits)
    	}
    	if transfer(nil, nil, 10) {
    		t.Errorf("transfer(nil, nil, 10) = true, want false")
    	}
    	if transfer(w, w, 10) || w.Credits != 50 {
    		t.Errorf("transfer(w, w, 10) (same wallet) must return false and leave it at 50 credits (got %d)", w.Credits)
    	}
    	twin := &Wallet{Owner: "Mia", Credits: 50}
    	if !transfer(w, twin, 10) || w.Credits != 40 || twin.Credits != 60 {
    		t.Errorf("two different wallets with equal contents are still different wallets: transfer should succeed, got %d and %d credits, want 40 and 60", w.Credits, twin.Credits)
    	}
    }

    func TestTopUpAll(t *testing.T) {
    	a := &Wallet{Owner: "a", Credits: 1}
    	b := &Wallet{Owner: "b", Credits: 0}
    	if got := topUpAll([]*Wallet{a, nil, b, nil}, 10); got != 2 || a.Credits != 11 || b.Credits != 10 {
    		t.Errorf("topUpAll([a, nil, b, nil], 10) = %d, credits now %d and %d; want 2, credits 11 and 10", got, a.Credits, b.Credits)
    	}
    	if got := topUpAll(nil, 10); got != 0 {
    		t.Errorf("topUpAll(nil, 10) = %d, want 0", got)
    	}
    	if got := topUpAll([]*Wallet{nil, nil}, 10); got != 0 {
    		t.Errorf("topUpAll([nil, nil], 10) = %d, want 0", got)
    	}
    	if got := topUpAll([]*Wallet{a, b}, 0); got != 0 || a.Credits != 11 || b.Credits != 10 {
    		t.Errorf("topUpAll(wallets, 0) = %d, credits now %d and %d; want 0 with nothing changed", got, a.Credits, b.Credits)
    	}
    	if got := topUpAll([]*Wallet{a, b}, -5); got != 0 || a.Credits != 11 || b.Credits != 10 {
    		t.Errorf("topUpAll(wallets, -5) = %d, credits now %d and %d; want 0 with nothing changed", got, a.Credits, b.Credits)
    	}
    	if got := topUpAll([]*Wallet{a, a}, 5); got != 2 || a.Credits != 21 {
    		t.Errorf("topUpAll([a, a], 5) = %d, credits now %d; want 2, credits 21 (a wallet listed twice is topped up twice)", got, a.Credits)
    	}
    }
---

Textio customers can gift message credits to each other. Wallets are passed
around as pointers, so your functions change the real wallets, and some
entries may be `nil` (a deleted account).

Write two functions:

- `transfer(from, to *Wallet, amount int) bool` moves `amount` credits from
  `from` to `to` and returns `true`. It refuses (returns `false` and changes
  **nothing**) if either wallet is `nil`, `from` and `to` are the **same**
  wallet, `amount` isn't positive, or `from` doesn't have enough credits.
- `topUpAll(wallets []*Wallet, amount int) int` adds `amount` credits to every
  non-`nil` wallet and returns how many wallets it topped up. If `amount`
  isn't positive, it changes nothing and returns `0`.

## Example

```go
mia := &Wallet{Owner: "Mia", Credits: 50}
sam := &Wallet{Owner: "Sam", Credits: 5}

transfer(mia, sam, 20)                  // true: Mia 30, Sam 25
transfer(sam, mia, 100)                 // false: Sam only has 25
transfer(mia, mia, 5)                   // false: same wallet
topUpAll([]*Wallet{mia, nil, sam}, 10)  // 2: Mia 40, Sam 35
```

## Constraints

- "Same wallet" means the same pointer. Two different wallets that happen to
  hold the same owner and credits are still different wallets.
- A wallet listed twice in `topUpAll` is topped up twice.
