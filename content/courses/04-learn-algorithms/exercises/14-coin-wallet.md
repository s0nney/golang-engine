---
title: Coin Wallet
difficulty: hard
after: stacks
hints:
  - 'Undo always reverts the **most recent** transaction that is still in effect: last in, first out. That''s a stack of transactions.'
  - 'Recomputing the lowest balance by replaying the whole history is O(n) per call. Instead, when you push a transaction, also store the balance **after** it and the lowest balance **so far** (the smaller of the previous entry''s lowest and the new balance).'
  - 'Then `Balance` and `Lowest` just read the top entry (or 0 when the stack is empty), and `Undo` just pops. Everything is O(1), and the zero `Wallet` works if an empty stack means "balance 0, lowest 0".'
exercise:
  starter: |
    package main

    import "fmt"

    // Wallet tracks a creator's Clout Coin balance. The zero Wallet is
    // ready to use and has a balance of 0.
    type Wallet struct {
    	// your fields here
    }

    // Apply adds delta (which may be negative) to the balance.
    func (w *Wallet) Apply(delta int) {
    }

    // Undo reverts the most recent transaction that hasn't been undone yet.
    // It returns false if there is nothing to undo.
    func (w *Wallet) Undo() bool {
    	return false
    }

    // Balance returns the current balance.
    func (w *Wallet) Balance() int {
    	return 0
    }

    // Lowest returns the lowest balance the wallet has had, counting the
    // starting 0 and only the transactions that are still in effect.
    func (w *Wallet) Lowest() int {
    	return 0
    }

    func main() {
    	var w Wallet
    	w.Apply(50)
    	w.Apply(-80)
    	w.Apply(100)
    	fmt.Println(w.Balance(), w.Lowest()) // want 70 -30
    	w.Undo()
    	w.Undo()
    	fmt.Println(w.Balance(), w.Lowest()) // want 50 0
    }
  solution: |
    package main

    import "fmt"

    type entry struct {
    	balance, lowest int
    }

    // Wallet tracks a creator's Clout Coin balance. The zero Wallet is
    // ready to use and has a balance of 0.
    type Wallet struct {
    	history []entry // a stack: the top is the current state
    }

    func (w *Wallet) top() entry {
    	if len(w.history) == 0 {
    		return entry{}
    	}
    	return w.history[len(w.history)-1]
    }

    // Apply adds delta (which may be negative) to the balance.
    func (w *Wallet) Apply(delta int) {
    	t := w.top()
    	b := t.balance + delta
    	w.history = append(w.history, entry{balance: b, lowest: min(t.lowest, b)})
    }

    // Undo reverts the most recent transaction that hasn't been undone yet.
    // It returns false if there is nothing to undo.
    func (w *Wallet) Undo() bool {
    	if len(w.history) == 0 {
    		return false
    	}
    	w.history = w.history[:len(w.history)-1]
    	return true
    }

    // Balance returns the current balance.
    func (w *Wallet) Balance() int {
    	return w.top().balance
    }

    // Lowest returns the lowest balance the wallet has had, counting the
    // starting 0 and only the transactions that are still in effect.
    func (w *Wallet) Lowest() int {
    	return w.top().lowest
    }

    func main() {
    	var w Wallet
    	w.Apply(50)
    	w.Apply(-80)
    	w.Apply(100)
    	fmt.Println(w.Balance(), w.Lowest())
    	w.Undo()
    	w.Undo()
    	fmt.Println(w.Balance(), w.Lowest())
    }
  tests: |
    package main

    import (
    	"strconv"
    	"strings"
    	"testing"
    	"time"
    )

    // run applies a script like "+50 -80 u +10" and returns the
    // Balance/Lowest after each step, formatted as "b/l".
    func run(w *Wallet, script string) string {
    	var out []string
    	for _, op := range strings.Fields(script) {
    		if op == "u" {
    			ok := w.Undo()
    			out = append(out, "u:"+strconv.FormatBool(ok))
    		} else {
    			d, _ := strconv.Atoi(op)
    			w.Apply(d)
    		}
    		out = append(out, strconv.Itoa(w.Balance())+"/"+strconv.Itoa(w.Lowest()))
    	}
    	return strings.Join(out, " ")
    }

    func TestWallet(t *testing.T) {
    	tests := []struct{ script, want string }{
    		{"+50 -80 +100 u u", "50/0 -30/-30 70/-30 u:true -30/-30 u:true 50/0"},
    		{"u", "u:false 0/0"},
    		{"+10 u u +5", "10/0 u:true 0/0 u:false 0/0 5/0"},
    		{"-5 -5 -5 u", "-5/-5 -10/-10 -15/-15 u:true -10/-10"},
    		{"+100 -30 -30 -30 u u u u", "100/0 70/0 40/0 10/0 u:true 40/0 u:true 70/0 u:true 100/0 u:true 0/0"},
    		{"-20 +100 -150 +500 u u -10", "-20/-20 80/-20 -70/-70 430/-70 u:true -70/-70 u:true 80/-20 70/-20"},
    		{"+0 +0 u", "0/0 0/0 u:true 0/0"},
    	}
    	for _, tt := range tests {
    		var w Wallet
    		if got := run(&w, tt.script); got != tt.want {
    			t.Errorf("script %q:\n got  %s\n want %s", tt.script, got, tt.want)
    		}
    	}
    }

    func TestNewWallet(t *testing.T) {
    	var w Wallet
    	if w.Balance() != 0 || w.Lowest() != 0 {
    		t.Errorf("zero Wallet: Balance() = %d, Lowest() = %d, want 0 and 0", w.Balance(), w.Lowest())
    	}
    }

    func TestTwoWallets(t *testing.T) {
    	var a, b Wallet
    	a.Apply(-40)
    	b.Apply(7)
    	if a.Balance() != -40 || b.Balance() != 7 || b.Lowest() != 0 {
    		t.Errorf("two wallets share state: balances %d and %d, want -40 and 7", a.Balance(), b.Balance())
    	}
    }

    func TestWalletLarge(t *testing.T) {
    	n := 200_000
    	done := make(chan [2]int, 1)
    	go func() {
    		var w Wallet
    		sum := 0
    		for i := range n {
    			// Mostly rising, with a dip every 1,000 steps and an undo every 3.
    			if i%1000 == 999 {
    				w.Apply(-5000)
    			} else {
    				w.Apply(3)
    			}
    			if i%3 == 2 {
    				w.Undo()
    			}
    			sum += w.Lowest()
    		}
    		done <- [2]int{w.Balance(), sum}
    	}()
    	select {
    	case got := <-done:
    		want := [2]int{-270_400, -27_036_938_400}
    		if got != want {
    			t.Errorf("after %d operations: balance and sum of Lowest() = %v, want %v", n, got, want)
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("%d Apply/Undo/Lowest calls took over a second: Lowest must be O(1), not a scan of the history", n)
    	}
    }
---

Creators on Clout earn and spend **Clout Coins**. Support staff often need to
reverse a mistaken transaction, and the fraud team watches for wallets that
dipped deep into the negative. Build the `Wallet` type that backs both.

- `Apply(delta)` adds `delta` (positive or negative) to the balance.
- `Undo()` reverts the most recent transaction that is still in effect and
  returns `true`, or returns `false` if there's nothing to undo. Undo can be
  called repeatedly, stepping further back each time.
- `Balance()` returns the current balance.
- `Lowest()` returns the lowest balance the wallet has ever had, counting the
  starting balance of 0 and **only the transactions still in effect**: an
  undone transaction is as if it never happened.

The zero `Wallet` (`var w Wallet`) must be ready to use.

## Example

```go
var w Wallet
w.Apply(50)   // balance 50,  lowest 0
w.Apply(-80)  // balance -30, lowest -30
w.Apply(100)  // balance 70,  lowest -30
w.Undo()      // back to -30, lowest -30
w.Undo()      // back to 50,  lowest 0: the dip was undone
w.Undo()      // back to 0,   lowest 0
w.Undo()      // false: nothing left to undo
```

## Constraints

- Up to 200,000 operations; every balance fits in an `int`.
- Every method must be **O(1)** (amortized). The performance test calls
  `Lowest` after each of 200,000 operations under a one-second limit, so
  replaying the history on each call won't finish.
