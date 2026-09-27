---
title: strconv and fmt Verbs
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import (
      	"fmt"
      	"strconv"
      )

      func main() {
      	n, err := strconv.Atoi("12a")
      	fmt.Println(n, err != nil)
      }
      ```
    options:
      - text: '`12 false`'
      - text: '`12 true`'
      - text: '`0 true`'
        correct: true
      - text: It panics, because `"12a"` isn't a number
    explanation: |
      `"12a"` isn't a valid integer, so `Atoi` returns an error and the
      zero value `0`. Parsing doesn't panic: it hands you an error to check.
  - question: 'Which verb prints a string **with quotes around it**, so stray spaces show up?'
    options:
      - text: '`%s`'
      - text: '`%v`'
      - text: '`%q`'
        correct: true
      - text: '`%T`'
    explanation: |
      `%q` prints `"Alice "` with quotes (and escapes like `\n`), so you
      can see exactly where the string starts and ends. It's a great
      debugging verb.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // parseTopUp parses an SMS command like "TOPUP 25". It returns the
    // amount and true if cmd is a valid top-up, or 0 and false if not.
    func parseTopUp(cmd string) (int, bool) {
    	_, amount, found := strings.Cut(cmd, " ")
    	if !found {
    		return 0, false
    	}
    	// ? check the command word, and turn amount into an int
    	return len(amount), true
    }

    func main() {
    	fmt.Println(parseTopUp("TOPUP 25"))   // want 25 true
    	fmt.Println(parseTopUp("topup 5"))    // want 5 true
    	fmt.Println(parseTopUp("TOPUP lots")) // want 0 false
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strconv"
    	"strings"
    )

    func parseTopUp(cmd string) (int, bool) {
    	word, amount, found := strings.Cut(strings.TrimSpace(cmd), " ")
    	if !found || !strings.EqualFold(word, "TOPUP") {
    		return 0, false
    	}
    	n, err := strconv.Atoi(amount)
    	if err != nil || n <= 0 {
    		return 0, false
    	}
    	return n, true
    }

    func main() {
    	fmt.Println(parseTopUp("TOPUP 25"))
    	fmt.Println(parseTopUp("topup 5"))
    	fmt.Println(parseTopUp("TOPUP lots"))
    }
  tests: |
    package main

    import "testing"

    func TestParseTopUp(t *testing.T) {
    	tests := []struct {
    		cmd    string
    		want   int
    		wantOK bool
    	}{
    		{"TOPUP 25", 25, true},
    		{"topup 5", 5, true},
    		{"TopUp 1000", 1000, true},
    		{"  TOPUP 100  ", 100, true},
    		{"TOPUP", 0, false},
    		{"TOPUP lots", 0, false},
    		{"TOPUP 2.5", 0, false},
    		{"TOPUP -5", 0, false},
    		{"TOPUP 0", 0, false},
    		{"SEND 25", 0, false},
    		{"STOP", 0, false},
    		{"", 0, false},
    	}
    	for _, tt := range tests {
    		got, ok := parseTopUp(tt.cmd)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("parseTopUp(%q) = %d, %v; want %d, %v", tt.cmd, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }
---

Customers text commands to Textio, like `TOPUP 25` to buy 25 credits. That `25` arrives as **text**, and you need it as a **number**. Going between numbers and text is the job of the `strconv` ("string conversion") package, and going the other way with style is the job of `fmt`.

## Text to numbers

You met `strconv.Atoi` ("ASCII to integer") in the type conversion lesson. Parsing text can fail, so every parsing function returns an **error** as its second result:

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	n, err := strconv.Atoi("160")
	fmt.Println(n, err)

	n, err = strconv.Atoi("12a")
	fmt.Println(n, err)

	price, err := strconv.ParseFloat("0.05", 64)
	fmt.Println(price*100, err)

	optIn, err := strconv.ParseBool("true")
	fmt.Println(optIn, err)
}
```

```text
160 <nil>
0 strconv.Atoi: parsing "12a": invalid syntax
5 <nil>
true <nil>
```

- `strconv.Atoi(s)` parses an `int`.
- `strconv.ParseFloat(s, 64)` parses a `float64`. The `64` is the bit size you want.
- `strconv.ParseBool(s)` accepts `"true"`, `"false"`, `"1"`, `"0"` and a few more.

When parsing fails, you get the zero value and a non-`nil` error that explains what went wrong. Always check it before using the number. (The errors chapter coming up covers errors in depth. For now, `err != nil` means "it failed".)

User input is messy, so parsing is usually combined with the `strings` functions from earlier: trim the input, cut it into pieces, then parse the piece that should be a number.

## Numbers to text

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	credits := 25
	fmt.Println("Credits: " + strconv.Itoa(credits))
	fmt.Println(strconv.FormatFloat(0.125, 'f', 2, 64))
	fmt.Println("Credits: " + fmt.Sprint(credits))
}
```

```text
Credits: 25
0.12
Credits: 25
```

`strconv.Itoa` turns an `int` into its digits. `strconv.FormatFloat(f, 'f', 2, 64)` formats a float with 2 decimal places (0.125 rounds to `0.12` because floats can't store it exactly). `fmt.Sprint` converts anything, and is fine when speed doesn't matter.

## `fmt` verbs, a recap

You've been using `Printf` verbs since the variables chapter. Here are the ones worth knowing by heart, including a few new ones for strings, bytes and runes:

| Verb | Prints | Example |
|------|--------|---------|
| `%v` | Any value, default format | `42`, `[a b]`, `{alice 3}` |
| `%+v` | Structs with field names | `{name:alice credits:3}` |
| `%T` | The value's type | `[]string` |
| `%d` | An integer | `42` |
| `%5d` / `%-5d` | Padded to 5, right / left aligned | `   42` / `42   ` |
| `%f` / `%.2f` | A float / with 2 decimals | `4.500000` / `4.50` |
| `%s` | A string | `hi` |
| `%q` | A quoted string (or rune) | `"hi\n"`, `'👋'` |
| `%c` | A rune as a character | `é` |
| `%U` | A rune as a Unicode code point | `U+00E9` |
| `%x` | Hexadecimal (bytes of a string, too) | `2a`, `6869` |
| `%t` | A bool | `true` |
| `%%` | A literal percent sign | `%` |

```go
package main

import "fmt"

func main() {
	msg := "hé"
	fmt.Printf("%s|%q|%x|%d\n", msg, msg, msg, len(msg))
	for _, r := range msg {
		fmt.Printf("%c %U %q\n", r, r, r)
	}
	fmt.Printf("[%6.2f] [%-8s] [%3d%%]\n", 4.5, "alice", 98)
}
```

```text
hé|"hé"|68c3a9|3
h U+0068 'h'
é U+00E9 'é'
[  4.50] [alice   ] [ 98%]
```

`%x` on a string shows you its raw bytes in hexadecimal: `68` is `h`, and `c3 a9` are the two bytes of `é`. That's a handy way to see UTF-8 at work.

## Your turn

Complete `parseTopUp`, which reads a customer's top-up command. A valid command:

- is the word `TOPUP` (in any mix of upper and lower case), a single space, then a whole number of credits, like `TOPUP 25` or `topup 5`;
- may have extra whitespace at the very start or end, which you should ignore;
- must ask for **at least 1** credit.

Return the amount and `true` for a valid command, and `0, false` for anything else, including `TOPUP lots`, `TOPUP 2.5` and `TOPUP -5`.

Hint: `strings.TrimSpace`, `strings.Cut`, `strings.EqualFold` and `strconv.Atoi`.

## Further reading

- [Go by Example: Number Parsing](https://gobyexample.com/number-parsing)
- [Package fmt](https://pkg.go.dev/fmt)
