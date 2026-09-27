---
title: Variadic Functions
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func count(prefix string, names ...string) {
      	fmt.Println(prefix, len(names))
      }

      func main() {
      	count("recipients:")
      	count("recipients:", "Alice", "Bob")
      }
      ```
    options:
      - text: |
          ```text
          recipients: 0
          recipients: 2
          ```
        correct: true
      - text: It doesn't compile, because the first call passes no names
      - text: |
          ```text
          recipients: 1
          recipients: 3
          ```
    explanation: |
      A variadic parameter accepts zero or more values. With none, `names` is
      empty (length 0). The regular `prefix` parameter isn't part of the
      variadic list, so the second call has 2 names, not 3.
  - question: |
      You have `list := []string{"Alice", "Bob"}` and a function
      `func sendAll(names ...string)`. How do you pass `list` to it?
    options:
      - text: '`sendAll(list)`'
      - text: '`sendAll(...list)`'
      - text: '`sendAll(list...)`'
        correct: true
    explanation: |
      Putting `...` *after* a slice spreads its elements out as the variadic
      arguments. `sendAll(list)` doesn't compile, because `list` is a slice,
      not a `string`.
---

You've been calling a function that takes any number of arguments since your very first program:

```go
fmt.Println("one")
fmt.Println("one", "two", "three")
```

How does `Println` accept one argument, three, or none at all? It's a **variadic** function, and you can write your own.

## Declaring a variadic function

Put `...` before the type of the **last** parameter:

```go
package main

import "fmt"

func totalLength(messages ...string) int {
	total := 0
	for _, m := range messages {
		total += len(m)
	}
	return total
}

func main() {
	fmt.Println(totalLength())
	fmt.Println(totalLength("hi"))
	fmt.Println(totalLength("hi", "hello", "hey"))
}
```

```text
0
2
10
```

Inside the function, `messages` is a **slice** of strings: a list holding however many arguments were passed. You'll learn all about slices and the `for ... range` loop in later chapters. For now, just read the loop as "for each message, add its length to the total".

`total += len(m)` is shorthand for `total = total + len(m)`. There are similar shortcuts for other operators: `-=`, `*=` and `/=`.

## Rules

- Only the **last** parameter can be variadic.
- You can have regular parameters before it.
- Callers can pass zero values, and then the slice is empty.

```go
package main

import "fmt"

func notify(sender string, recipients ...string) {
	for _, r := range recipients {
		fmt.Printf("%s -> %s\n", sender, r)
	}
}

func main() {
	notify("Textio", "Alice", "Bob")
}
```

```text
Textio -> Alice
Textio -> Bob
```

## Spreading a slice into a variadic call

If you already have a slice, you can't pass it directly, because the function expects individual strings. Add `...` **after** the slice to spread its elements out:

```go
package main

import "fmt"

func notify(sender string, recipients ...string) {
	fmt.Println(sender, "is messaging", len(recipients), "people")
}

func main() {
	team := []string{"Alice", "Bob", "Carol"}
	notify("Textio", team...)
}
```

```text
Textio is messaging 3 people
```

Easy way to remember:

- `...string` in a **declaration** means "gather arguments into a slice".
- `team...` in a **call** means "spread this slice out into arguments".

## Built-ins that are variadic

Several of Go's built-in functions work the same way. `min` and `max` take any number of values:

```go
fmt.Println(max(3, 9, 4))    // 9
fmt.Println(min(2.5, 1.0))   // 1
```

And `append`, which you'll meet in the slices chapter, is variadic too. Once you recognise the pattern, you'll see it everywhere in Go's standard library.

## Further reading

- [Go by Example: Variadic Functions](https://gobyexample.com/variadic-functions)
