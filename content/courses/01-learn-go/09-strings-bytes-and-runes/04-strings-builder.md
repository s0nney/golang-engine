---
title: Building Strings
quiz:
  - question: Why is `s += piece` inside a long loop slow?
    options:
      - text: Because `+=` isn't allowed on strings, so Go converts to `[]byte` every time
      - text: Because strings are immutable, so every `+=` copies everything built so far into a brand new string
        correct: true
      - text: Because the loop variable is copied on every iteration
    explanation: |
      A string can't grow in place. Each `+=` allocates a new string and
      copies all the old bytes plus the new ones. For many pieces, that's a
      lot of copying. `strings.Builder` grows a buffer instead.
  - question: |
      What does this program print?

      ```go
      package main

      import (
      	"fmt"
      	"strings"
      )

      func main() {
      	var b strings.Builder
      	b.WriteString("Hi")
      	b.WriteByte(' ')
      	b.WriteRune('👋')
      	fmt.Println(b.Len(), b.String())
      }
      ```
    options:
      - text: '`4 Hi 👋`'
      - text: '`7 Hi 👋`'
        correct: true
      - text: It panics, because the builder was never created with `make`
    explanation: |
      A zero-value `strings.Builder` is ready to use. `Len` counts **bytes**
      written so far: 2 for `Hi`, 1 for the space, and 4 for the emoji.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // numberedList formats items as a numbered list, one per line,
    // each line ending in a newline. An empty list gives "".
    func numberedList(items []string) string {
    	var b strings.Builder
    	// ?
    	return b.String()
    }

    func main() {
    	fmt.Print(numberedList([]string{"Your code is 4821", "Your order shipped"}))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func numberedList(items []string) string {
    	var b strings.Builder
    	for i, item := range items {
    		fmt.Fprintf(&b, "%d. %s\n", i+1, item)
    	}
    	return b.String()
    }

    func main() {
    	fmt.Print(numberedList([]string{"Your code is 4821", "Your order shipped"}))
    }
  tests: |
    package main

    import "testing"

    func TestNumberedList(t *testing.T) {
    	tests := []struct {
    		items []string
    		want  string
    	}{
    		{nil, ""},
    		{[]string{"hi"}, "1. hi\n"},
    		{[]string{"Your code is 4821", "Your order shipped"}, "1. Your code is 4821\n2. Your order shipped\n"},
    		{[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}, "1. a\n2. b\n3. c\n4. d\n5. e\n6. f\n7. g\n8. h\n9. i\n10. j\n"},
    	}
    	for _, tt := range tests {
    		if got := numberedList(tt.items); got != tt.want {
    			t.Errorf("numberedList(%q) = %q, want %q", tt.items, got, tt.want)
    		}
    	}
    }
---

Textio builds a lot of text: receipts, daily digests, reports. You already know `+` joins strings. So why not just do this?

```go
report := ""
for _, line := range lines {
	report += line + "\n"
}
```

For a handful of lines it's fine. For thousands it gets slow, and here's why.

## Why `+=` is slow in a loop

Strings are immutable. `report += line` can't add to the end of the existing string. It has to allocate a **new** string big enough for both, and copy **all** of the old bytes and the new ones into it. The next iteration copies all of *that* again. The more you've built, the more each step copies.

## `strings.Builder`

A `strings.Builder` keeps a growing buffer of bytes, like a slice you `append` to, and only turns it into a string at the end:

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	var b strings.Builder
	b.WriteString("Daily digest\n")
	for i := range 3 {
		b.WriteString("- message ")
		b.WriteByte('A' + byte(i))
		b.WriteRune('\n')
	}
	fmt.Print(b.String())
	fmt.Println(b.Len(), "bytes")
}
```

```text
Daily digest
- message A
- message B
- message C
49 bytes
```

- The **zero value** is ready to use: `var b strings.Builder`. No `make` needed.
- `WriteString` adds a string, `WriteByte` adds one byte, and `WriteRune` adds one rune (encoded as UTF-8).
- `String()` returns everything written so far. `Len()` returns its length in bytes.

`'A' + byte(i)` is a neat trick: characters are numbers, so `'A' + 1` is `'B'`.

## `fmt.Fprintf` writes into a builder

You already know `Printf` (print to the screen) and `Sprintf` (return a string). There's a third sibling, `Fprintf`, which writes formatted text **into** something you give it, such as a builder:

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	names := []string{"alice", "bob"}
	costs := []float64{0.42, 1.5}

	var b strings.Builder
	for i, name := range names {
		fmt.Fprintf(&b, "%-6s $%.2f\n", name, costs[i])
	}
	fmt.Print(b.String())
}
```

```text
alice  $0.42
bob    $1.50
```

Notice `&b`: `Fprintf` needs a **pointer** to the builder so it can add to the real one, not a copy. (Pointers get their own chapter soon. For now, remember the `&`.) For the same reason, never copy a `strings.Builder` value after you start writing to it. Pass `*strings.Builder` around instead.

## When to use what

- Joining a slice with a separator? `strings.Join`.
- A few pieces, once? `+` or `fmt.Sprintf` are perfectly clear.
- Building text in a loop? `strings.Builder`.

## Your turn

Textio's weekly digest shows each message as a numbered list. Complete `numberedList` so that it returns one line per item, numbered from `1`, in this format:

```text
1. Your code is 4821
2. Your order shipped
```

Every line ends with a newline (including the last one), and an empty list gives an empty string. Use a `strings.Builder`.

## Further reading

- [Package strings: Builder](https://pkg.go.dev/strings#Builder)
