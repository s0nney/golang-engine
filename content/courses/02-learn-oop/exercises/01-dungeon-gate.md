---
title: Dungeon Gate
difficulty: easy
after: clean-code
hints:
  - 'Don''t try to patch the pyramid. Start a fresh function body and turn each rule into a **guard clause**: `if <rule broken> { return "<reason>" }`. The happy path, `return ""`, goes at the very end.'
  - 'Write the guards in the order the rules are listed. The old code had two off-by-one bugs: a hero with exactly 0 HP is down, and exactly 4 heroes is allowed.'
  - 'Give the average a name (`avgLevel`) and compute it with integer division **after** the empty-party guard, so you never divide by zero.'
exercise:
  starter: |
    package main

    import "fmt"

    type Hero struct {
    	Name  string
    	Level int
    	HP    int
    }

    // gateVerdict explains why a party may NOT enter a dungeon whose minimum
    // level is minLevel, or returns "" if the gate opens.
    //
    // Somebody wrote this in a hurry. It's hard to follow and it has bugs.
    // Rewrite it with guard clauses so each rule is one early return.
    func gateVerdict(party []Hero, minLevel int) string {
    	r := ""
    	if len(party) > 0 {
    		if len(party) < 4 {
    			x := 0
    			for i := 0; i < len(party); i++ {
    				if party[i].HP < 0 {
    					if r == "" {
    						r = party[i].Name + " is down"
    					}
    				}
    				x = x + party[i].Level
    			}
    			if r == "" {
    				if x/len(party) < minLevel {
    					r = fmt.Sprintf("too weak: average level %d, need %d", x/len(party), minLevel)
    				}
    			}
    		} else {
    			r = fmt.Sprintf("too many heroes: %d, max 4", len(party))
    		}
    	} else {
    		r = "no heroes"
    	}
    	return r
    }

    func main() {
    	party := []Hero{{"Ayla", 5, 30}, {"Bram", 3, 0}, {"Cora", 6, 22}}
    	fmt.Printf("%q\n", gateVerdict(party, 4)) // want "Bram is down"
    	party[1].HP = 12
    	fmt.Printf("%q\n", gateVerdict(party, 4)) // want ""
    	fmt.Printf("%q\n", gateVerdict(party, 5)) // want "too weak: average level 4, need 5"
    }
  solution: |
    package main

    import "fmt"

    type Hero struct {
    	Name  string
    	Level int
    	HP    int
    }

    const maxPartySize = 4

    // gateVerdict explains why a party may NOT enter a dungeon whose minimum
    // level is minLevel, or returns "" if the gate opens.
    func gateVerdict(party []Hero, minLevel int) string {
    	if len(party) == 0 {
    		return "no heroes"
    	}
    	if len(party) > maxPartySize {
    		return fmt.Sprintf("too many heroes: %d, max %d", len(party), maxPartySize)
    	}
    	totalLevel := 0
    	for _, h := range party {
    		if h.HP <= 0 {
    			return h.Name + " is down"
    		}
    		totalLevel += h.Level
    	}
    	if avgLevel := totalLevel / len(party); avgLevel < minLevel {
    		return fmt.Sprintf("too weak: average level %d, need %d", avgLevel, minLevel)
    	}
    	return ""
    }

    func main() {
    	party := []Hero{{"Ayla", 5, 30}, {"Bram", 3, 0}, {"Cora", 6, 22}}
    	fmt.Printf("%q\n", gateVerdict(party, 4))
    	party[1].HP = 12
    	fmt.Printf("%q\n", gateVerdict(party, 4))
    	fmt.Printf("%q\n", gateVerdict(party, 5))
    }
  tests: |
    package main

    import "testing"

    func TestGateVerdict(t *testing.T) {
    	ayla, bram, cora := Hero{"Ayla", 5, 30}, Hero{"Bram", 3, 12}, Hero{"Cora", 6, 22}
    	tests := []struct {
    		name     string
    		party    []Hero
    		minLevel int
    		want     string
    	}{
    		{"empty party", nil, 1, "no heroes"},
    		{"empty party, no minimum", []Hero{}, 0, "no heroes"},
    		{"strong enough", []Hero{ayla, bram, cora}, 4, ""},
    		{"average rounds down", []Hero{ayla, bram, cora}, 5, "too weak: average level 4, need 5"},
    		{"exactly the minimum", []Hero{ayla}, 5, ""},
    		{"solo too weak", []Hero{bram}, 4, "too weak: average level 3, need 4"},
    		{"four heroes is fine", []Hero{ayla, bram, cora, ayla}, 1, ""},
    		{"five heroes", []Hero{ayla, bram, cora, ayla, bram}, 1, "too many heroes: 5, max 4"},
    		{"too many beats a downed hero", []Hero{ayla, {"Dax", 9, 0}, cora, ayla, bram}, 1, "too many heroes: 5, max 4"},
    		{"0 HP means down", []Hero{ayla, {"Dax", 9, 0}, cora}, 1, "Dax is down"},
    		{"negative HP means down", []Hero{{"Eli", 2, -4}}, 1, "Eli is down"},
    		{"first downed hero is named", []Hero{ayla, {"Fen", 7, 0}, {"Gus", 7, -1}}, 1, "Fen is down"},
    		{"downed beats too weak", []Hero{{"Hal", 1, 0}}, 10, "Hal is down"},
    	}
    	for _, tt := range tests {
    		if got := gateVerdict(tt.party, tt.minLevel); got != tt.want {
    			t.Errorf("%s: gateVerdict(%v, %d) = %q, want %q", tt.name, tt.party, tt.minLevel, got, tt.want)
    		}
    	}
    }
---

The gatekeeper of the Sunken Keep checks every party before raising the
portcullis. Whoever wrote `gateVerdict` built a pyramid of nested `if`s, reused
`r` and `x` for everything, and hid two bugs in the tangle.

Rewrite `gateVerdict(party, minLevel)` so it's easy to read **and** correct.
It returns the reason the party is turned away, or `""` if the gate opens.
Check the rules **in this order** and report only the first one that fails:

1. The party is empty: `"no heroes"`.
2. There are more than 4 heroes: `"too many heroes: N, max 4"`.
3. A hero has 0 HP or less: `"<Name> is down"`, naming the **first** such hero.
4. The party's average level (integer division, rounded down) is below
   `minLevel`: `"too weak: average level A, need M"`.

Otherwise return `""`.

## Examples

```go
party := []Hero{{"Ayla", 5, 30}, {"Bram", 3, 0}, {"Cora", 6, 22}}
gateVerdict(party, 4) // "Bram is down"

party[1].HP = 12
gateVerdict(party, 4) // "" (average level 14/3 = 4)
gateVerdict(party, 5) // "too weak: average level 4, need 5"
gateVerdict(nil, 1)   // "no heroes"
```

## Constraints

- Only the returned string is graded, but aim for guard clauses, clear names
  and no magic numbers: that's the point of the exercise.
