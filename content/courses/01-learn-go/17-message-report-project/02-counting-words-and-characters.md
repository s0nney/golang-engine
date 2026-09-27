---
title: Counting Words and Characters
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
      	words := strings.Fields("  Go   go\tGO!  ")
      	fmt.Println(len(words), words[2])
      }
      ```
    options:
      - text: '`5 GO!`'
      - text: '`3 GO!`'
        correct: true
      - text: '`3 go`'
      - text: '`1 Go   go\tGO!`'
    explanation: |
      `strings.Fields` splits on *any* run of whitespace (spaces, tabs,
      newlines) and ignores leading and trailing whitespace, so you get three
      words. It doesn't change case or strip punctuation, so the third word is
      still `GO!`.
  - question: 'Why count characters with `utf8.RuneCountInString(s)` instead of `len(s)`?'
    options:
      - text: '`len` doesn''t work on strings'
      - text: '`len` counts bytes, and characters like `é` or emoji take more than one byte'
        correct: true
      - text: '`utf8.RuneCountInString` also skips spaces'
    explanation: |
      `len` returns the number of **bytes**. Characters outside plain ASCII are
      encoded with 2 to 4 bytes in UTF-8, so `len("café")` is 5 while
      `utf8.RuneCountInString("café")` is 4.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    )

    // countWords returns how many times each normalized word appears in messages.
    func countWords(messages []string) map[string]int {
    	counts := make(map[string]int)
    	// ?
    	return counts
    }

    // countChars returns the number of non-whitespace characters (runes) in messages.
    func countChars(messages []string) int {
    	// ?
    	return 0
    }

    func main() {
    	messages := []string{"Lunch? Yes, lunch!", "café at noon 👋"}
    	counts := countWords(messages)
    	fmt.Println(counts["lunch"], counts["café"], len(counts))
    	fmt.Println(countChars(messages))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    	"unicode"
    )

    func countWords(messages []string) map[string]int {
    	counts := make(map[string]int)
    	for _, msg := range messages {
    		for _, word := range strings.Fields(msg) {
    			word = strings.ToLower(strings.Trim(word, ".,!?"))
    			if word == "" {
    				continue
    			}
    			counts[word]++
    		}
    	}
    	return counts
    }

    func countChars(messages []string) int {
    	chars := 0
    	for _, msg := range messages {
    		for _, r := range msg {
    			if !unicode.IsSpace(r) {
    				chars++
    			}
    		}
    	}
    	return chars
    }

    func main() {
    	messages := []string{"Lunch? Yes, lunch!", "café at noon 👋"}
    	counts := countWords(messages)
    	fmt.Println(counts["lunch"], counts["café"], len(counts))
    	fmt.Println(countChars(messages))
    }
  tests: |
    package main

    import (
    	"maps"
    	"testing"
    )

    func TestCountWords(t *testing.T) {
    	for _, tc := range []struct {
    		messages []string
    		want     map[string]int
    	}{
    		{[]string{"Lunch? Yes, lunch!", "café at noon 👋"},
    			map[string]int{"lunch": 2, "yes": 1, "café": 1, "at": 1, "noon": 1, "👋": 1}},
    		{[]string{"GO go\tGo.", "!!! ?"}, map[string]int{"go": 3}},
    		{[]string{"don't stop", "Don't!"}, map[string]int{"don't": 2, "stop": 1}},
    		{nil, map[string]int{}},
    	} {
    		got := countWords(tc.messages)
    		if got == nil {
    			t.Errorf("countWords(%q) returned a nil map, want an empty map", tc.messages)
    			continue
    		}
    		if !maps.Equal(got, tc.want) {
    			t.Errorf("countWords(%q) = %v, want %v", tc.messages, got, tc.want)
    		}
    	}
    }

    func TestCountChars(t *testing.T) {
    	for _, tc := range []struct {
    		messages []string
    		want     int
    	}{
    		{[]string{"Lunch? Yes, lunch!", "café at noon 👋"}, 27},
    		{[]string{"a b\tc\n"}, 3},
    		{[]string{"héllo"}, 5},
    		{[]string{"   "}, 0},
    		{nil, 0},
    	} {
    		if got := countChars(tc.messages); got != tc.want {
    			t.Errorf("countChars(%q) = %d, want %d", tc.messages, got, tc.want)
    		}
    	}
    }
---

Our tool can load messages. Now let's count what's in them. Textio's report needs
two numbers per file: how many **words** there are and how many **characters**, plus
how often each word appears.

## Splitting into words

`strings.Fields` splits a string around runs of whitespace. That's exactly what we
want for words, because people type double spaces, tabs and stray newlines all the
time:

```go
strings.Fields("see  you\tsoon") // ["see" "you" "soon"]
```

## Normalizing words

To a human, `Lunch`, `lunch` and `lunch?` are the same word. To a map they're three
different keys. So before counting, **normalize** each word:

- `strings.ToLower` makes it lowercase.
- `strings.Trim(word, ".,!?")` removes those characters from both ends (but not the
  middle, so `don't` survives).

A word made only of punctuation, like `!!!`, becomes the empty string. Skip it.

## Counting with a map

Remember that reading a missing key from a map gives the zero value. That makes
counting a one-liner, `counts[word]++`:

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	text := "Lunch? Yes, lunch! LUNCH at noon !!!"
	counts := make(map[string]int)
	for _, word := range strings.Fields(text) {
		word = strings.ToLower(strings.Trim(word, ".,!?"))
		if word == "" {
			continue
		}
		counts[word]++
	}
	fmt.Println(counts["lunch"], counts["noon"], len(counts))
}
```

```text
3 1 4
```

Four distinct words: `lunch`, `yes`, `at` and `noon`. The `!!!` became empty and was
skipped. Printing `len(counts)` while you develop is a good habit, because it
catches surprises like a stray `""` key.

## Counting characters

Textio charges by the character, so the report should count them too. As you saw in
the strings chapter, `len` counts bytes, not characters. Loop over the string with
`range` to visit each **rune** instead, and use `unicode.IsSpace` to leave out
whitespace:

```go
chars := 0
for _, r := range "hé 👋" {
	if !unicode.IsSpace(r) {
		chars++
	}
}
fmt.Println(chars) // 3
```

## Your turn

Complete both functions. They take the `[]string` that `loadMessages` returned in the
previous lesson.

1. `countWords` splits every message with `strings.Fields`, trims `.,!?` from both
   ends with `strings.Trim`, lowercases the result, skips empty words, and counts the
   rest in the map. It must return an empty (not `nil`) map when there are no words.
2. `countChars` counts every rune in every message except whitespace
   (`unicode.IsSpace`).

**Run** should print `2 1 6` and then `27`.
