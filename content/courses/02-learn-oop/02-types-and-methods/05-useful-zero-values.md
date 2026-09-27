---
title: Make the Zero Value Useful
quiz:
  - question: |
      What does this print?

      ```go
      type Inventory struct {
      	items []string
      }

      func (inv *Inventory) Add(item string) {
      	inv.items = append(inv.items, item)
      }

      func (inv *Inventory) Len() int { return len(inv.items) }

      func main() {
      	var inv Inventory
      	inv.Add("potion")
      	inv.Add("rope")
      	fmt.Println(inv.Len())
      }
      ```
    options:
      - text: It panics because `items` is nil
      - text: '`2`'
        correct: true
      - text: '`0`'
      - text: It doesn't compile without a constructor
    explanation: |
      `append` works fine on a nil slice, and `len` of a nil slice is 0. So the
      zero `Inventory` is ready to use with no constructor at all.
  - question: |
      Which field would break a "useful zero value" if you wrote to it without initialising it first?
    options:
      - text: '`gold int`'
      - text: '`items []string` (written with `append`)'
      - text: '`counts map[string]int` (written with `counts[k]++`)'
        correct: true
      - text: '`mu sync.Mutex`'
    explanation: |
      Writing to a nil map panics. Ints start at 0, `append` handles nil slices and
      a zero `sync.Mutex` is an unlocked mutex. For maps, create them lazily inside
      the method (`if m == nil { m = make(...) }`).
---

Every Go type has a **zero value**: `0` for numbers, `""` for strings, `nil` for slices, maps and pointers, and for structs, every field at its own zero value. When you write `var inv Inventory`, that's what you get.

Another Go proverb says:

> Make the zero value useful.

It means: design your types so that the zero value is **ready to use**, with no constructor required.

## The standard library does this everywhere

```go
var sb strings.Builder
sb.WriteString("Dragon ")
sb.WriteString("slain!")
fmt.Println(sb.String()) // Dragon slain!

var mu sync.Mutex
mu.Lock() // a zero Mutex is an unlocked mutex
mu.Unlock()

var buf bytes.Buffer
buf.WriteString("loot") // a zero Buffer is an empty buffer
```

None of these need a `New...` call. You can embed them in your own structs and they just work.

## Designing an RPG type this way

Let's build a hero's inventory. The trick is to lean on things that already have useful zero values, and to handle `nil` lazily where they don't:

```go
package main

import "fmt"

type Inventory struct {
	items []string
	gold  int
	count map[string]int
}

func (inv *Inventory) Add(item string) {
	inv.items = append(inv.items, item) // append handles a nil slice
	if inv.count == nil {
		inv.count = make(map[string]int) // create the map on first use
	}
	inv.count[item]++
}

func (inv *Inventory) AddGold(n int) { inv.gold += n }

func (inv *Inventory) Has(item string) bool {
	return inv.count[item] > 0 // reading a nil map is fine
}

func main() {
	var inv Inventory // no constructor!
	inv.Add("potion")
	inv.Add("potion")
	inv.Add("rope")
	inv.AddGold(50)

	fmt.Println(len(inv.items), inv.gold)
	fmt.Println(inv.Has("potion"), inv.Has("sword"))
}
```

```
3 50
true false
```

The gotcha to remember: **reading** a nil map returns the zero value, but **writing** to one panics. That's why `Add` checks `inv.count == nil` before writing.

## Why bother?

- **Fewer ways to get it wrong.** Nobody can forget to call the constructor, because there isn't one to forget.
- **Easy composition.** A `Party` struct can contain an `Inventory` field and it's ready immediately.
- **Simpler tests.** `var inv Inventory` is the whole setup.

## When you still need a constructor

Some types can't have a sensible zero value. A `Hero` with no name and 0 HP isn't a hero, it's a corpse. A dragon needs a real element. That's fine: use a `NewHero` function, as in the constructors lesson, and document that the zero value isn't meant to be used.

The rule of thumb:

- If the zero value *can* be useful, make it useful and skip the constructor.
- If it can't, provide `NewX` and, in the next chapter, use unexported fields so other packages can't build a broken one by hand.

## Further reading

- [Effective Go: Allocation with new](https://go.dev/doc/effective_go#allocation_new)
