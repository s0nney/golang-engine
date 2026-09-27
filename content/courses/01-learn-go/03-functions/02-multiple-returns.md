---
title: Multiple Return Values
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func splitCost(total float64, people int) (float64, float64) {
      	each := total / float64(people)
      	return each, total - each
      }

      func main() {
      	_, rest := splitCost(9, 3)
      	fmt.Println(rest)
      }
      ```
    options:
      - text: '`3`'
      - text: '`6`'
        correct: true
      - text: '`3 6`'
      - text: It doesn't compile, because the first return value isn't used
    explanation: |
      `each` is `9 / 3 = 3`, and the second return value is `9 - 3 = 6`. The
      blank identifier `_` throws away the first value, so only `rest` (6)
      is stored and printed.
  - question: 'Why won''t `name := getUser()` compile when `getUser` returns `(string, int)`?'
    options:
      - text: Because `getUser` has no parameters
      - text: Because the function returns two values, so the caller must receive two
        correct: true
      - text: Because `:=` can't be used with function calls
    explanation: |
      When a function returns two values, you must accept both:
      `name, age := getUser()`. If you don't need one of them, discard it with
      `_`, as in `name, _ := getUser()`.
exercise:
  starter: |
    package main

    import "fmt"

    // bill returns how many of this month's messages are charged,
    // and their total cost in cents. The first 5 messages each month
    // are free, and every message after that costs 2 cents.
    func bill(messages int) (int, int) {
    	// ?
    	return 0, 0
    }

    func main() {
    	charged, cost := bill(12)
    	fmt.Println(charged, "charged messages cost", cost, "cents") // want: 7 charged messages cost 14 cents
    }
  solution: |
    package main

    import "fmt"

    func bill(messages int) (int, int) {
    	charged := max(messages-5, 0)
    	return charged, charged * 2
    }

    func main() {
    	charged, cost := bill(12)
    	fmt.Println(charged, "charged messages cost", cost, "cents")
    }
  tests: |
    package main

    import "testing"

    func TestBill(t *testing.T) {
    	tests := []struct {
    		messages    int
    		wantCharged int
    		wantCost    int
    	}{
    		{12, 7, 14},
    		{5, 0, 0},
    		{0, 0, 0},
    		{3, 0, 0},
    		{6, 1, 2},
    		{100, 95, 190},
    	}
    	for _, tt := range tests {
    		charged, cost := bill(tt.messages)
    		if charged != tt.wantCharged || cost != tt.wantCost {
    			t.Errorf("bill(%d) = %d, %d; want %d, %d", tt.messages, charged, cost, tt.wantCharged, tt.wantCost)
    		}
    	}
    }
---

Most languages let a function return one value. Go lets a function return **several**, and Go code uses this everywhere.

## Returning more than one value

List the return types in parentheses, and return the values separated by commas:

```go
package main

import "fmt"

func getContact() (string, string) {
	return "Alice", "+1-555-0100"
}

func main() {
	name, phone := getContact()
	fmt.Println(name, "can be reached at", phone)
}
```

```text
Alice can be reached at +1-555-0100
```

The caller receives the values in the same order the function returns them. You must receive *all* of them: `name := getContact()` won't compile, because there are two values and only one variable.

## A real use: a result and a status

Textio splits long messages into 160-character chunks. A function can report both how many chunks are needed and how many characters are left over in the last one:

```go
package main

import "fmt"

func chunks(length int) (int, int) {
	full := length / 160
	leftover := length % 160
	return full, leftover
}

func main() {
	full, leftover := chunks(350)
	fmt.Println(full, "full chunks and", leftover, "extra characters")
}
```

```text
2 full chunks and 30 extra characters
```

The most common pairing in Go is **a value and an error**: "here's the result, and here's whether something went wrong". You met this already with `strconv.Atoi`, and you'll use it constantly once you reach the errors chapter.

## Ignoring values with `_`

Sometimes you only care about one of the values. Remember that Go refuses to compile unused variables. So how do you ignore one?

With the **blank identifier**, `_`. It's a write-only placeholder: anything assigned to it is thrown away.

```go
_, leftover := chunks(350)
fmt.Println(leftover) // 30
```

You can use `_` as many times as you like, and it never counts as an unused variable.

## Named return values

You can give return values names. They become variables inside the function, starting at their zero values, and a bare `return` returns their current values:

```go
func chunks(length int) (full int, leftover int) {
	full = length / 160
	leftover = length % 160
	return
}
```

Named returns are useful as documentation: the signature now tells you what each `int` means. But a bare `return` in a long function makes it hard to see what's being returned. Most Go programmers name results when it helps readability, and still write `return full, leftover` explicitly.

## Returning early with several values

When a function has multiple return values, every `return` statement must supply all of them:

```go
func chunks(length int) (int, int) {
	if length == 0 {
		return 0, 0
	}
	return length / 160, length % 160
}
```

That brings us to the next lesson: returning early.

## Your turn

Textio gives every customer 5 free messages a month. After that, each message costs 2 cents.

Complete `bill` so that it returns **two** values: how many messages are charged, and their total cost in cents. For example, `bill(12)` returns `7, 14`, and `bill(3)` returns `0, 0` (never a negative number!).

Hint: the built-in `max` function returns the larger of its arguments.

## Further reading

- [Go by Example: Multiple Return Values](https://gobyexample.com/multiple-return-values)
- [A Tour of Go: Named return values](https://go.dev/tour/basics/7)
