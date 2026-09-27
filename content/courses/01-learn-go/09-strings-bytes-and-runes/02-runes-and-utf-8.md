---
title: Runes and UTF-8
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	for i, r := range "hé!" {
      		fmt.Print(i, ":", string(r), " ")
      	}
      	fmt.Println()
      }
      ```
    options:
      - text: '`0:h 1:é 2:! `'
      - text: '`0:h 1:é 3:! `'
        correct: true
      - text: '`0:h 1:� 2:� 3:! `'
      - text: '`0:104 1:233 3:33 `'
    explanation: |
      Ranging over a string decodes it rune by rune, but `i` is the **byte**
      index where each rune starts. `é` takes two bytes (1 and 2), so `!`
      starts at byte 3. `string(r)` turns each rune back into a string.
  - question: 'How many characters does `utf8.RuneCountInString("Hi 👋")` report?'
    options:
      - text: '`4`'
        correct: true
      - text: '`7`'
      - text: '`3`'
    explanation: |
      `H`, `i`, the space and `👋` are four runes. `len` would say 7,
      because the emoji alone takes 4 bytes.
exercise:
  starter: |
    package main

    import "fmt"

    // charCount returns the number of characters (runes) in msg.
    func charCount(msg string) int {
    	return len(msg) // ? this counts bytes
    }

    // truncate returns msg cut down to at most limit characters,
    // without ever cutting a character in half.
    func truncate(msg string, limit int) string {
    	if len(msg) <= limit {
    		return msg
    	}
    	return msg[:limit] // ? this slices bytes
    }

    func main() {
    	fmt.Println(charCount("Hi 👋"))                   // want 4
    	fmt.Printf("%q\n", truncate("Olá, mundo! 👋", 5)) // want "Olá, "
    }
  solution: |
    package main

    import (
    	"fmt"
    	"unicode/utf8"
    )

    func charCount(msg string) int {
    	return utf8.RuneCountInString(msg)
    }

    func truncate(msg string, limit int) string {
    	runes := []rune(msg)
    	if len(runes) <= limit {
    		return msg
    	}
    	return string(runes[:limit])
    }

    func main() {
    	fmt.Println(charCount("Hi 👋"))
    	fmt.Printf("%q\n", truncate("Olá, mundo! 👋", 5))
    }
  tests: |
    package main

    import "testing"

    func TestCharCount(t *testing.T) {
    	tests := []struct {
    		msg  string
    		want int
    	}{
    		{"", 0},
    		{"hello", 5},
    		{"héllo", 5},
    		{"Hi 👋", 4},
    		{"👋👋👋", 3},
    		{"こんにちは", 5},
    	}
    	for _, tt := range tests {
    		if got := charCount(tt.msg); got != tt.want {
    			t.Errorf("charCount(%q) = %d, want %d", tt.msg, got, tt.want)
    		}
    	}
    }

    func TestTruncate(t *testing.T) {
    	tests := []struct {
    		msg   string
    		limit int
    		want  string
    	}{
    		{"hello", 10, "hello"},
    		{"hello", 5, "hello"},
    		{"hello", 3, "hel"},
    		{"héllo", 2, "hé"},
    		{"👋 hi", 1, "👋"},
    		{"Olá, mundo! 👋", 5, "Olá, "},
    		{"Olá, mundo! 👋", 13, "Olá, mundo! 👋"},
    		{"こんにちは", 3, "こんに"},
    		{"", 5, ""},
    		{"abc", 0, ""},
    	}
    	for _, tt := range tests {
    		if got := truncate(tt.msg, tt.limit); got != tt.want {
    			t.Errorf("truncate(%q, %d) = %q, want %q", tt.msg, tt.limit, got, tt.want)
    		}
    	}
    }
---

If `len` counts bytes, how do you count *characters*? Meet the **rune**.

## Runes

Unicode gives every character a number, called a **code point**: `A` is 65, `é` is 233, `👋` is 128075. Go's `rune` type holds one code point. It's just another name for `int32`, and you write rune literals in **single quotes**:

```go
package main

import "fmt"

func main() {
	r := 'é'
	fmt.Println(r)
	fmt.Printf("%c %U %T\n", r, r, r)
}
```

```text
233
é U+00E9 int32
```

`%c` prints the character, `%U` prints its Unicode code point in the standard `U+` form, and `%T` confirms that a rune really is an `int32`.

Double quotes make a `string`, single quotes make a `rune`. `"a"` and `'a'` are different types.

UTF-8 is the recipe that turns each rune into 1 to 4 bytes. Go does the decoding for you whenever you ask for runes.

## Ranging over a string gives runes

Indexing a string gives bytes, but a `for ... range` loop over a string **decodes UTF-8** and gives you one rune per iteration:

```go
package main

import "fmt"

func main() {
	for i, r := range "hé👋" {
		fmt.Printf("%d %c %U\n", i, r, r)
	}
}
```

```text
0 h U+0068
1 é U+00E9
3 👋 U+1F44B
```

The first value, `i`, is the **byte index** where each rune starts. Notice it jumps from 1 to 3, because `é` takes two bytes. The second value, `r`, is the rune.

Compare that with counting through the bytes yourself:

```go
s := "hé👋"
for i := range len(s) {
	fmt.Print(s[i], " ") // 104 195 169 240 159 145 139
}
```

Seven bytes, three runes. When you care about characters, range over the string.

## Counting characters

The `unicode/utf8` package counts runes without you writing a loop:

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	msg := "Olá 👋"
	fmt.Println(len(msg), "bytes")
	fmt.Println(utf8.RuneCountInString(msg), "characters")
}
```

```text
9 bytes
5 characters
```

## Converting between strings, runes and bytes

| Conversion | Result |
|------------|--------|
| `[]byte(s)` | The string's bytes (a copy) |
| `[]rune(s)` | The string decoded into runes (a copy) |
| `string(b)` for a `[]byte` | A string made from those bytes |
| `string(rs)` for a `[]rune` | A string made from those runes, encoded as UTF-8 |
| `string(r)` for a `rune` | A one-character string |

`[]rune` is handy when you need to work *by character*, because then indexes and lengths count characters:

```go
package main

import "fmt"

func main() {
	runes := []rune("¡Hola!")
	fmt.Println(len(runes))
	fmt.Println(string(runes[0]))
	fmt.Println(string(runes[1:5]))
}
```

```text
6
¡
Hola
```

Remember the trap from the type conversion lesson: `string(65)` gives `"A"`, not `"65"`. Converting a `rune` to a `string` gives you a character, and that's exactly what you want here. For numbers as digits, use `strconv`.

## A gotcha: what's a "character"?

Some things that *look* like one character are several runes. A flag like 🇬🇧 is two runes, and 👍🏽 (thumbs up plus a skin tone) is two as well. Counting what a human sees as one character is surprisingly hard, and needs packages outside the standard library. For Textio, counting runes is close enough.

## Your turn

Textio's message composer shows a character counter, and trims messages that are too long. Both are broken for anyone who uses emoji, because they work with bytes.

Fix them:

- `charCount` returns the number of **runes** in `msg`.
- `truncate` returns `msg` cut down to at most `limit` **characters**. If `msg` is already short enough, return it unchanged. It must never cut a character in half.

Hint: `utf8.RuneCountInString` and `[]rune` are your friends.

## Further reading

- [Strings, bytes, runes and characters in Go (Go blog)](https://go.dev/blog/strings)
- [Go by Example: Strings and Runes](https://gobyexample.com/strings-and-runes)
