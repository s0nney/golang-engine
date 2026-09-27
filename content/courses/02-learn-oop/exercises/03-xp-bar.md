---
title: XP Bar
difficulty: easy
after: types-and-methods
hints:
  - '`Gain` has to change the bar, so it needs a **pointer receiver** (`func (b *XPBar) Gain(...)`). With a value receiver it would level up a copy and throw it away.'
  - 'For a useful zero value, don''t store the level itself (the zero value would be level 0). Store how many levels have been **gained**, and have `Level()` return `1 + gained`.'
  - 'In `Gain`, add `n` to the XP, then loop: while the XP is at least `Level()*100`, subtract that amount and gain a level. A single huge gain can cross several levels.'
exercise:
  starter: |
    package main

    import "fmt"

    // XPBar tracks a hero's level and experience. The zero XPBar is a fresh
    // level 1 hero with 0 XP, so `var bar XPBar` must be ready to use.
    //
    // Reaching level L+1 from level L costs L*100 XP. Leftover XP carries over.
    type XPBar struct {
    	// your fields here
    }

    // Gain adds n XP and returns how many levels the hero gained.
    // Gaining 0 or negative XP does nothing and returns 0.
    func (b XPBar) Gain(n int) int {
    	return 0
    }

    // Level returns the hero's current level (1 for a new hero).
    func (b XPBar) Level() int {
    	return 0
    }

    // XP returns the XP earned towards the next level.
    func (b XPBar) XP() int {
    	return 0
    }

    // ToNext returns how much more XP the hero needs to level up.
    func (b XPBar) ToNext() int {
    	return 0
    }

    func main() {
    	var bar XPBar
    	fmt.Println(bar.Level(), bar.XP(), bar.ToNext()) // want 1 0 100
    	fmt.Println(bar.Gain(250))                       // want 1
    	fmt.Println(bar.Level(), bar.XP(), bar.ToNext()) // want 2 150 50
    }
  solution: |
    package main

    import "fmt"

    // XPBar tracks a hero's level and experience. The zero XPBar is a fresh
    // level 1 hero with 0 XP, so `var bar XPBar` must be ready to use.
    //
    // Reaching level L+1 from level L costs L*100 XP. Leftover XP carries over.
    type XPBar struct {
    	gained int // levels gained so far; the level is 1 + gained
    	xp     int // XP towards the next level
    }

    // Gain adds n XP and returns how many levels the hero gained.
    // Gaining 0 or negative XP does nothing and returns 0.
    func (b *XPBar) Gain(n int) int {
    	if n <= 0 {
    		return 0
    	}
    	b.xp += n
    	ups := 0
    	for b.xp >= b.cost() {
    		b.xp -= b.cost()
    		b.gained++
    		ups++
    	}
    	return ups
    }

    // cost is the XP needed to go from the current level to the next.
    func (b *XPBar) cost() int {
    	return b.Level() * 100
    }

    // Level returns the hero's current level (1 for a new hero).
    func (b *XPBar) Level() int {
    	return 1 + b.gained
    }

    // XP returns the XP earned towards the next level.
    func (b *XPBar) XP() int {
    	return b.xp
    }

    // ToNext returns how much more XP the hero needs to level up.
    func (b *XPBar) ToNext() int {
    	return b.cost() - b.xp
    }

    func main() {
    	var bar XPBar
    	fmt.Println(bar.Level(), bar.XP(), bar.ToNext())
    	fmt.Println(bar.Gain(250))
    	fmt.Println(bar.Level(), bar.XP(), bar.ToNext())
    }
  tests: |
    package main

    import "testing"

    func state(b *XPBar) [3]int {
    	return [3]int{b.Level(), b.XP(), b.ToNext()}
    }

    func TestZeroValue(t *testing.T) {
    	var b XPBar
    	if got, want := state(&b), [3]int{1, 0, 100}; got != want {
    		t.Errorf("new XPBar: [Level XP ToNext] = %v, want %v", got, want)
    	}
    }

    func TestGain(t *testing.T) {
    	tests := []struct {
    		gains []int
    		ups   []int  // what each Gain returns
    		want  [3]int // Level, XP, ToNext afterwards
    	}{
    		{[]int{50}, []int{0}, [3]int{1, 50, 50}},
    		{[]int{100}, []int{1}, [3]int{2, 0, 200}},
    		{[]int{99, 1}, []int{0, 1}, [3]int{2, 0, 200}},
    		{[]int{250}, []int{1}, [3]int{2, 150, 50}},
    		{[]int{300}, []int{2}, [3]int{3, 0, 300}},
    		{[]int{1000}, []int{4}, [3]int{5, 0, 500}},
    		{[]int{1499}, []int{4}, [3]int{5, 499, 1}},
    		{[]int{40, 40, 40}, []int{0, 0, 1}, [3]int{2, 20, 180}},
    		{[]int{0}, []int{0}, [3]int{1, 0, 100}},
    		{[]int{70, -500}, []int{0, 0}, [3]int{1, 70, 30}},
    		{[]int{-1, 0, 100}, []int{0, 0, 1}, [3]int{2, 0, 200}},
    	}
    	for _, tt := range tests {
    		var b XPBar
    		for i, n := range tt.gains {
    			if got := b.Gain(n); got != tt.ups[i] {
    				t.Errorf("gains %v: Gain(%d) returned %d, want %d", tt.gains, n, got, tt.ups[i])
    			}
    		}
    		if got := state(&b); got != tt.want {
    			t.Errorf("after gains %v: [Level XP ToNext] = %v, want %v", tt.gains, got, tt.want)
    		}
    	}
    }

    func TestBarsAreIndependent(t *testing.T) {
    	var a, b XPBar
    	a.Gain(300)
    	b.Gain(10)
    	if a.Level() != 3 || b.Level() != 1 || b.XP() != 10 {
    		t.Errorf("two bars: a.Level() = %d, b.Level() = %d, b.XP() = %d, want 3, 1, 10", a.Level(), b.Level(), b.XP())
    	}
    }
---

Every hero has an XP bar under their portrait. Build the `XPBar` type behind it.

- A new hero is **level 1 with 0 XP**, and the zero value `var bar XPBar` must
  be exactly that hero, with no constructor needed.
- Going from level `L` to level `L+1` costs `L*100` XP (100 to reach level 2,
  200 more to reach level 3, and so on). XP left over after a level-up carries
  over, so one big reward can grant several levels.
- `Gain(n)` adds `n` XP and returns the number of levels gained. `n <= 0` does
  nothing and returns 0.
- `Level()` is the current level, `XP()` the XP earned towards the next level,
  and `ToNext()` how much more XP the next level needs.

The starter's methods all have **value receivers**. Decide which ones must change.

## Example

```go
var bar XPBar
bar.Level(), bar.XP(), bar.ToNext() // 1, 0, 100
bar.Gain(250)                       // 1 (100 XP buys level 2; 150 carries over)
bar.Level(), bar.XP(), bar.ToNext() // 2, 150, 50
bar.Gain(350)                       // 2 (50 for level 3, then 300 for level 4)
bar.Level(), bar.XP()               // 4, 0
```

## Constraints

- Gains fit comfortably in an `int`.
