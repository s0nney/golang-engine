---
title: Comparisons and Logic
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	credits := 0
      	isAdmin := true
      	fmt.Println(credits > 0 || isAdmin, credits > 0 && isAdmin)
      }
      ```
    options:
      - text: '`true true`'
      - text: '`false false`'
      - text: '`false true`'
      - text: '`true false`'
        correct: true
    explanation: |
      `credits > 0` is `false`. With `||` (OR), only one side needs to be
      true, and `isAdmin` is, so the first result is `true`. With `&&` (AND),
      both sides must be true, so the second result is `false`.
  - question: Which expression is `true` when `length` is between 1 and 160, inclusive?
    options:
      - text: '`1 <= length <= 160`'
      - text: '`length >= 1 && length <= 160`'
        correct: true
      - text: '`length >= 1 || length <= 160`'
      - text: '`length = 1..160`'
    explanation: |
      You need both conditions to hold, so join them with `&&`. Go doesn't
      support chained comparisons like `1 <= length <= 160`, and with `||`
      the expression would be true for almost every number.
---

Before a program can make decisions, it needs to ask yes-or-no questions. "Is this message too long?" "Does the user have credits?" In Go, the answer to a yes-or-no question is a `bool`: `true` or `false`.

## Comparison operators

Comparison operators compare two values and produce a `bool`:

| Operator | Meaning |
|----------|---------|
| `==` | equal to |
| `!=` | not equal to |
| `<` | less than |
| `>` | greater than |
| `<=` | less than or equal to |
| `>=` | greater than or equal to |

```go
package main

import "fmt"

func main() {
	length := 172
	fmt.Println(length > 160)
	fmt.Println(length == 172)
	fmt.Println(length != 172)
	fmt.Println("alice" == "Alice")
}
```

```text
true
true
false
false
```

String comparison is exact and case-sensitive, so `"alice"` and `"Alice"` are different.

**Watch out:** `=` and `==` are completely different. `=` *assigns* a value; `==` *compares* two values. Go won't let you mix them up inside an `if`, which saves you from a classic bug in other languages.

Both sides of a comparison must have the same type. `length == "172"` won't compile, because you can't compare an `int` with a `string`.

## Logical operators

To combine several yes-or-no questions, use **logical operators**:

| Operator | Name | True when... |
|----------|------|---------------|
| `&&` | AND | **both** sides are true |
| `\|\|` | OR | **at least one** side is true |
| `!` | NOT | the value is false (it flips it) |

```go
package main

import "fmt"

func main() {
	hasCredits := true
	optedOut := false
	length := 90

	canSend := hasCredits && !optedOut && length <= 160
	fmt.Println(canSend)

	needsReview := length > 1000 || optedOut
	fmt.Println(needsReview)
}
```

```text
true
false
```

Read `!optedOut` as "not opted out".

## Short-circuit evaluation

Go evaluates `&&` and `||` from left to right, and **stops as soon as it knows the answer**:

- With `a && b`, if `a` is false, the whole thing is false, so `b` is never evaluated.
- With `a || b`, if `a` is true, the whole thing is true, so `b` is never evaluated.

This is called **short-circuiting**, and it's genuinely useful. You can put a safety check first and rely on it:

```go
// If count is 0, the division never happens.
ok := count > 0 && total/count > 5
```

Without short-circuiting, `total/count` would crash the program when `count` is `0` (integer division by zero panics).

## No chained comparisons

In maths you might write `1 ≤ x ≤ 160`. Go doesn't allow `1 <= x <= 160`. Spell it out with `&&`:

```go
valid := length >= 1 && length <= 160
```

## Precedence

`!` binds tightest, then comparisons, then `&&`, then `||`. So `a || b && c` means `a || (b && c)`. When in doubt, add parentheses. They cost nothing and make your intent obvious.
