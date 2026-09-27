---
title: What Is a Closure?
quiz:
  - question: What is a closure?
    options:
      - text: A function that has finished running
      - text: The closing brace of a function
      - text: A function value that refers to variables from the scope where it was created
        correct: true
      - text: A function that closes open files
    explanation: |
      A closure "closes over" variables from its surrounding scope. It can read and
      change them even after the surrounding function has returned.
  - question: |
      What does this print?

      ```go
      ext := ".md"
      addExt := func(name string) string { return name + ext }
      ext = ".txt"
      fmt.Println(addExt("notes"))
      ```
    options:
      - text: '`notes.txt`'
        correct: true
      - text: '`notes.md`'
      - text: '`notes`'
    explanation: |
      A closure captures the *variable* `ext`, not a snapshot of its value. By the
      time `addExt` runs, `ext` holds `".txt"`.
---

You've already written several closures without being told. Remember `surround`?

```go
func surround(left, right string) func(string) string {
	return func(s string) string {
		return left + s + right
	}
}
```

The inner function uses `left` and `right`, which aren't its parameters. They belong
to `surround`. And yet the inner function keeps working long after `surround` has
returned. How?

## Closing over variables

A **closure** is a function value bundled with the variables it references from the
surrounding scope. The function "closes over" those variables and keeps them alive
for as long as the function value exists.

Normally, a local variable goes away when its function returns. But if a closure
still refers to it, Go keeps it around. The compiler moves it to the heap if needed,
which is called *escaping*. You don't have to manage any of this yourself.

## Doc2Doc: a link builder

```go
package main

import "fmt"

func linker(baseURL string) func(page string) string {
	return func(page string) string {
		return fmt.Sprintf("[%s](%s/%s.html)", page, baseURL, page)
	}
}

func main() {
	docs := linker("https://docs.example.com")
	blog := linker("https://blog.example.com")

	fmt.Println(docs("install"))
	fmt.Println(blog("launch"))
	fmt.Println(docs("faq"))
}
```

```text
[install](https://docs.example.com/install.html)
[launch](https://blog.example.com/launch.html)
[faq](https://docs.example.com/faq.html)
```

`docs` and `blog` are made by the same code, but each closure has its **own**
`baseURL`. Calling `linker` twice created two separate variables, one per closure.

## Variables, not values

This is the most important rule about closures, so here it is in bold:
**a closure captures variables, not values.** It doesn't take a snapshot. It holds
on to the variable itself, and reads its *current* value every time it runs.

```go
package main

import "fmt"

func main() {
	style := "plain"
	describe := func() string { return "style is " + style }

	fmt.Println(describe())
	style = "fancy"
	fmt.Println(describe())
}
```

```text
style is plain
style is fancy
```

The closure saw the change, because it and `main` share the same `style` variable.
It works the other way too: a closure can *assign* to a captured variable, and the
change is visible outside. That's the basis for the stateful closures in the next
lesson.

## Closures and purity

A closure that only *reads* captured variables that never change (like `baseURL`
above) behaves like a pure function. It's a function with some settings baked in. A
closure that *changes* captured variables has state, and state is a side effect. Both
are useful, as long as you know which kind you're writing.

## Further reading

- [A Tour of Go: Function closures](https://go.dev/tour/moretypes/25)
