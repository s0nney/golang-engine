---
title: The strings Package
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import (
      	"fmt"
      	"strings"
      )

      func main() {
      	parts := strings.Split("a,,b", ",")
      	words := strings.Fields("  a   b  ")
      	fmt.Println(len(parts), len(words))
      }
      ```
    options:
      - text: '`2 2`'
      - text: '`3 2`'
        correct: true
      - text: '`3 6`'
      - text: '`2 6`'
    explanation: |
      `Split` cuts at *every* separator, so `"a,,b"` gives `["a" "" "b"]`,
      with an empty string between the two commas. `Fields` splits on runs
      of whitespace and never returns empty strings, so it gives `["a" "b"]`.
  - question: |
      What does this program print?

      ```go
      package main

      import (
      	"fmt"
      	"strings"
      )

      func main() {
      	cmd, arg, ok := strings.Cut("TOPUP 25", " ")
      	fmt.Printf("%q %q %v\n", cmd, arg, ok)
      }
      ```
    options:
      - text: '`"TOPUP" "25" true`'
        correct: true
      - text: '`"TOPUP " "25" true`'
      - text: '`"TOPUP" " 25" true`'
      - text: '`"TOPUP 25" "" false`'
    explanation: |
      `strings.Cut` splits around the **first** separator and drops the
      separator itself. `ok` reports whether the separator was found.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // sanitize cleans up a message before Textio sends it:
    //   - whitespace at the start and end is removed,
    //   - every run of spaces, tabs or newlines inside becomes a single space,
    //   - every banned word (compared case-insensitively) is replaced by
    //     asterisks, one per byte of the word.
    func sanitize(msg string, banned []string) string {
    	return strings.TrimSpace(msg) // ? do the rest
    }

    func main() {
    	banned := []string{"darn", "heck"}
    	fmt.Printf("%q\n", sanitize("  Oh   DARN,\tmy\ncode  ", banned))
    	fmt.Printf("%q\n", sanitize("what the heck", banned))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    	"strings"
    )

    func sanitize(msg string, banned []string) string {
    	words := strings.Fields(msg)
    	for i, w := range words {
    		if slices.Contains(banned, strings.ToLower(w)) {
    			words[i] = strings.Repeat("*", len(w))
    		}
    	}
    	return strings.Join(words, " ")
    }

    func main() {
    	banned := []string{"darn", "heck"}
    	fmt.Printf("%q\n", sanitize("  Oh   DARN,\tmy\ncode  ", banned))
    	fmt.Printf("%q\n", sanitize("what the heck", banned))
    }
  tests: |
    package main

    import "testing"

    func TestSanitize(t *testing.T) {
    	banned := []string{"darn", "heck"}
    	tests := []struct {
    		msg  string
    		want string
    	}{
    		{"hello", "hello"},
    		{"  hello  ", "hello"},
    		{"Your   code\tis\n4821", "Your code is 4821"},
    		{"what the heck", "what the ****"},
    		{"DARN it", "**** it"},
    		{"Heck, darn!", "Heck, darn!"},
    		{"  heck\n\nheck  ", "**** ****"},
    		{"checkout now", "checkout now"},
    		{"", ""},
    		{"   ", ""},
    	}
    	for _, tt := range tests {
    		if got := sanitize(tt.msg, banned); got != tt.want {
    			t.Errorf("sanitize(%q, %q) = %q, want %q", tt.msg, banned, got, tt.want)
    		}
    	}
    }
---

You could write loops over bytes and runes for every text job, but you'd be reinventing the wheel. The standard library's `strings` package already has well-tested functions for almost everything you'll need. Here's the toolkit Textio uses every day.

## Searching

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	msg := "Your Textio code is 4821"
	fmt.Println(strings.Contains(msg, "code"))
	fmt.Println(strings.HasPrefix(msg, "Your"))
	fmt.Println(strings.HasSuffix(msg, "!"))
	fmt.Println(strings.Index(msg, "code"))
	fmt.Println(strings.Count("banana", "a"))
}
```

```text
true
true
false
12
3
```

`Index` returns the byte position of the first match, or `-1` if there isn't one. All of these are **case-sensitive**: `"Code"` and `"code"` are different strings.

## Changing case and trimming

Remember, strings are immutable, so these all return a **new** string:

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.ToUpper("stop"))
	fmt.Println(strings.ToLower("STOP"))
	fmt.Printf("%q\n", strings.TrimSpace("  \t hi there \n"))
	fmt.Println(strings.Trim("!!sale!!", "!"))
	fmt.Println(strings.EqualFold("Stop", "STOP"))
}
```

```text
STOP
stop
"hi there"
sale
true
```

`TrimSpace` removes whitespace (spaces, tabs, newlines) from both ends. `Trim` removes any of the characters you list. `EqualFold` compares two strings, ignoring case.

## Replacing

```go
fmt.Println(strings.ReplaceAll("SALE SALE SALE", "SALE", "deal")) // deal deal deal
fmt.Println(strings.Replace("SALE SALE SALE", "SALE", "deal", 1)) // deal SALE SALE
```

`Replace` takes a count of how many matches to replace. `ReplaceAll` replaces every one.

## Splitting and joining

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	recipients := strings.Split("alice,bob,carol", ",")
	fmt.Println(len(recipients), recipients)

	words := strings.Fields("  your   code\tis 4821 ")
	fmt.Println(len(words), words)

	fmt.Println(strings.Join(recipients, " & "))
}
```

```text
3 [alice bob carol]
4 [your code is 4821]
alice & bob & carol
```

- `Split(s, sep)` cuts at every `sep`. Two separators in a row give an empty string between them.
- `Fields(s)` splits on any run of whitespace and ignores whitespace at the ends. It's the right tool for "give me the words".
- `Join(parts, sep)` is the opposite of `Split`: it glues the pieces together with `sep` between them.

## Cutting in two

A very common job is splitting a string at the first separator, like `"alice: see you at 9"` into a name and a message. `strings.Cut` does exactly that:

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	user, body, ok := strings.Cut("alice: see you at 9", ": ")
	fmt.Printf("%q %q %v\n", user, body, ok)

	_, _, ok = strings.Cut("no colon here", ": ")
	fmt.Println(ok)
}
```

```text
"alice" "see you at 9" true
false
```

Go 1.27 added `strings.CutLast`, which cuts around the **last** separator instead.

## Iterating over lines and parts

`strings.Split` builds a whole slice. When you only want to loop over the pieces, `strings.SplitSeq` and `strings.Lines` give you an **iterator** that you can `range` over, without building a slice first:

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	log := "alice: hi\nbob: hey\n"
	for line := range strings.Lines(log) {
		fmt.Printf("%q\n", line)
	}
	for part := range strings.SplitSeq("a-b-c", "-") {
		fmt.Print(part, " ")
	}
	fmt.Println()
}
```

```text
"alice: hi\n"
"bob: hey\n"
a b c
```

Notice that `strings.Lines` keeps the `\n` at the end of each line. Trim it off with `strings.TrimSpace` or `strings.TrimSuffix(line, "\n")` if you don't want it.

## Your turn

Before a message goes out, Textio cleans it up. Complete `sanitize` so that it:

1. Removes whitespace at the start and end.
2. Turns every run of spaces, tabs or newlines inside the message into a single space.
3. Replaces each **word** that is in `banned` (comparing in lowercase) with asterisks, one per byte of the word. `"heck"` becomes `"****"` and so does `"HECK"`.

A "word" here is whatever `strings.Fields` gives you, so `"Heck,"` (with its comma) is not a banned word, and neither is `"checkout"`.

Hint: `strings.Fields` and `strings.Join` together handle steps 1 and 2 in one go. `slices.Contains`, `strings.ToLower` and `strings.Repeat` handle step 3.

## Further reading

- [Package strings](https://pkg.go.dev/strings)
- [Go by Example: String Functions](https://gobyexample.com/string-functions)
