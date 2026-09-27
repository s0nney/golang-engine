---
title: Mana Pool
difficulty: easy
after: encapsulation
hints:
  - 'Keep both numbers in **unexported** fields (`current`, `max`). Code outside the package can then only change them through your methods, and every method can enforce `0 <= Mana() <= Max()`.'
  - 'In `Spend`, check everything that could make the spell fail **before** you touch `current`. A failed cast must leave the pool exactly as it was.'
  - '`min(p.max, p.current+n)` restores without overflowing the pool, and `max(0, limit)` turns a negative limit into 0.'
exercise:
  starter: |
    package main

    import "fmt"

    // ManaPool holds a mage's mana. Its mana is always between 0 and its
    // maximum, and code outside the type can't break that rule.
    type ManaPool struct {
    	// your (unexported) fields here
    }

    // NewManaPool returns a full pool holding limit mana.
    // A negative limit counts as 0.
    func NewManaPool(limit int) *ManaPool {
    	return &ManaPool{}
    }

    // Mana returns the mana currently in the pool.
    func (p *ManaPool) Mana() int {
    	return 0
    }

    // Max returns the pool's maximum.
    func (p *ManaPool) Max() int {
    	return 0
    }

    // Spend removes cost mana and returns true. If cost is negative or the pool
    // holds less than cost, it returns false and leaves the pool unchanged.
    func (p *ManaPool) Spend(cost int) bool {
    	return false
    }

    // Restore adds n mana, never going above Max. A negative n does nothing.
    func (p *ManaPool) Restore(n int) {
    }

    func main() {
    	pool := NewManaPool(50)
    	fmt.Println(pool.Spend(30), pool.Mana()) // want true 20
    	fmt.Println(pool.Spend(30), pool.Mana()) // want false 20
    	pool.Restore(100)
    	fmt.Println(pool.Mana(), pool.Max()) // want 50 50
    }
  solution: |
    package main

    import "fmt"

    // ManaPool holds a mage's mana. Its mana is always between 0 and its
    // maximum, and code outside the type can't break that rule.
    type ManaPool struct {
    	current int
    	limit   int
    }

    // NewManaPool returns a full pool holding limit mana.
    // A negative limit counts as 0.
    func NewManaPool(limit int) *ManaPool {
    	limit = max(0, limit)
    	return &ManaPool{current: limit, limit: limit}
    }

    // Mana returns the mana currently in the pool.
    func (p *ManaPool) Mana() int {
    	return p.current
    }

    // Max returns the pool's maximum.
    func (p *ManaPool) Max() int {
    	return p.limit
    }

    // Spend removes cost mana and returns true. If cost is negative or the pool
    // holds less than cost, it returns false and leaves the pool unchanged.
    func (p *ManaPool) Spend(cost int) bool {
    	if cost < 0 || cost > p.current {
    		return false
    	}
    	p.current -= cost
    	return true
    }

    // Restore adds n mana, never going above Max. A negative n does nothing.
    func (p *ManaPool) Restore(n int) {
    	if n < 0 {
    		return
    	}
    	p.current = min(p.limit, p.current+n)
    }

    func main() {
    	pool := NewManaPool(50)
    	fmt.Println(pool.Spend(30), pool.Mana())
    	fmt.Println(pool.Spend(30), pool.Mana())
    	pool.Restore(100)
    	fmt.Println(pool.Mana(), pool.Max())
    }
  tests: |
    package main

    import (
    	"reflect"
    	"strconv"
    	"strings"
    	"testing"
    )

    // run applies a script like "s30 r10" (spend 30, restore 10) to a new pool
    // and records the result of each step as "ok:mana" or "mana".
    func run(limit int, script string) string {
    	p := NewManaPool(limit)
    	out := []string{strconv.Itoa(p.Mana())}
    	for _, op := range strings.Fields(script) {
    		n, _ := strconv.Atoi(op[1:])
    		if op[0] == 's' {
    			out = append(out, strconv.FormatBool(p.Spend(n))+":"+strconv.Itoa(p.Mana()))
    		} else {
    			p.Restore(n)
    			out = append(out, strconv.Itoa(p.Mana()))
    		}
    		if p.Mana() < 0 || p.Mana() > p.Max() {
    			out = append(out, "BROKEN")
    		}
    	}
    	return strings.Join(out, " ")
    }

    func TestManaPool(t *testing.T) {
    	tests := []struct {
    		limit  int
    		script string
    		want   string
    	}{
    		{50, "s30 s30 r100", "50 true:20 false:20 50"},
    		{50, "s50 s1 r1 s1", "50 true:0 false:0 1 true:0"},
    		{10, "s0 s-5 r-5", "10 true:10 false:10 10"},
    		{10, "s4 r2 r2 r2", "10 true:6 8 10 10"},
    		{0, "s0 s1 r5", "0 true:0 false:0 0"},
    		{-7, "r3 s1", "0 0 false:0"},
    		{100, "s11 s11 s11 r0 s67 s0", "100 true:89 true:78 true:67 67 true:0 true:0"},
    	}
    	for _, tt := range tests {
    		if got := run(tt.limit, tt.script); got != tt.want {
    			t.Errorf("NewManaPool(%d) then %q:\n got  %s\n want %s", tt.limit, tt.script, got, tt.want)
    		}
    	}
    }

    func TestMax(t *testing.T) {
    	for _, tt := range []struct{ limit, want int }{{50, 50}, {0, 0}, {-3, 0}} {
    		p := NewManaPool(tt.limit)
    		p.Spend(1)
    		if got := p.Max(); got != tt.want {
    			t.Errorf("NewManaPool(%d).Max() = %d, want %d", tt.limit, got, tt.want)
    		}
    	}
    }

    func TestFieldsAreHidden(t *testing.T) {
    	typ := reflect.TypeFor[ManaPool]()
    	if typ.NumField() == 0 {
    		t.Fatal("ManaPool has no fields: store the mana and the maximum inside it")
    	}
    	for f := range typ.Fields() {
    		if f.IsExported() {
    			t.Errorf("ManaPool.%s is exported: other packages could set it to anything. Unexport it", f.Name)
    		}
    	}
    }
---

A mage's mana pool fills up to a maximum, drains when spells are cast, and must
never go negative or overflow. Build `ManaPool` so that **no code outside the
package can break those rules**: keep its fields unexported and let the methods
guard every change.

- `NewManaPool(limit)` returns a **full** pool with maximum `limit`. A negative
  limit counts as 0.
- `Mana()` and `Max()` report the current mana and the maximum.
- `Spend(cost)` removes `cost` mana and returns `true`. If `cost` is negative or
  bigger than the current mana, it returns `false` and changes **nothing**.
  Spending 0 always succeeds.
- `Restore(n)` adds `n` mana, stopping at the maximum. A negative `n` does
  nothing (use `Spend` to remove mana).

## Example

```go
pool := NewManaPool(50)
pool.Spend(30)  // true, mana 20
pool.Spend(30)  // false, mana still 20
pool.Restore(100)
pool.Mana()     // 50: capped at the maximum
pool.Spend(-10) // false: that would have been a sneaky refill
```

## Constraints

- All fields of `ManaPool` must be unexported. A test checks.
