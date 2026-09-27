---
title: Naming Things
quiz:
  - question: In Go, which variable name is most idiomatic for a loop index used in a three-line loop?
    options:
      - text: '`theCurrentIndexIntoTheHeroesSlice`'
      - text: '`i`'
        correct: true
      - text: '`indexCounterVariable`'
      - text: '`INDEX`'
    explanation: |
      Go favours short names for things with a small scope. A loop index that
      lives for three lines is perfectly clear as `i`. Longer names are for things
      that live longer and are used far from where they're declared.
  - question: |
      You have a package called `hero`. Which exported constructor name reads best at the call site?
    options:
      - text: '`hero.NewHeroObject`'
      - text: '`hero.HeroNew`'
      - text: '`hero.New`'
        correct: true
      - text: '`hero.CreateNewHeroInstance`'
    explanation: |
      Callers always write the package name first, so `hero.New(...)` already says
      "a new hero". Repeating the package name (`hero.NewHero`) is called stutter.
      It's not wrong, but Go style avoids it when the package makes the meaning clear.
---

There's an old joke: *there are only two hard things in computer science: cache invalidation and naming things.* Good names are the cheapest way to make code readable, so let's learn how Go programmers pick them.

## Names should say what, not how

Compare:

```go
func calc(a []int) int
```

with:

```go
func totalDamage(hits []int) int
```

The second tells you what it computes and what goes in. You barely need to read the body.

## Short names for short scopes

Go is famous for short names like `i`, `r`, `w` and `err`. That's not laziness. The rule is:

> The further a name is used from where it's declared, the more descriptive it should be.

```go
for i, h := range heroes {
	fmt.Println(i, h.Name)
}
```

`h` is used one line after it's declared, so `h` is plenty. But a package-level variable used in twenty files deserves a full name like `defaultSpawnRate`.

## Go naming conventions

- **MixedCaps**, never underscores: `maxHealth`, `DragonKing`, not `max_health`.
- **Acronyms stay one case**: `userID`, `HTTPServer`, `xmlParser`, not `UserId` or `HttpServer`.
- **Capital first letter means exported** (visible outside the package). We'll dig into this in the encapsulation chapter.
- **Package names** are short, lower-case, single words: `combat`, `hero`, `inventory`. No `utils` or `helpers`, because they say nothing.
- **Don't stutter**: inside package `hero`, call the type `Hero` and the constructor `New`, so callers write `hero.New()`, not `hero.NewHero()` or `hero.HeroStruct`.
- **Booleans read like questions**: `isAlive`, `hasShield`, `canFly`.
- **Interfaces with one method** are often named after the method plus `-er`: `Attacker`, `Healer`, `Stringer`.

## Putting it together

Here's a function with bad names:

```go
func proc(l []int, x int) (int, bool) {
	t := 0
	for _, v := range l {
		t += v
	}
	return t, t >= x
}
```

And the same function, renamed:

```go
package main

import "fmt"

func totalDamage(hits []int, dragonHP int) (total int, slain bool) {
	for _, hit := range hits {
		total += hit
	}
	return total, total >= dragonHP
}

func main() {
	total, slain := totalDamage([]int{40, 55, 80}, 150)
	fmt.Println(total, slain)
}
```

```
175 true
```

Same logic, but now the code documents itself. The named results `total` and `slain` even show up in the function signature, so anyone hovering over it in their editor knows what they get back.

## Names are design

When you struggle to name something, that's often a sign the *thing itself* is muddled. A function called `handleStuffAndSave` probably does two jobs and wants to be split. Keep that instinct: in the next chapter we'll start designing types, and a good name is the first test of a good type.

## Further reading

- [Effective Go: Names](https://go.dev/doc/effective_go#names)
