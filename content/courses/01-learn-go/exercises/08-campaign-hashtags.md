---
title: Campaign Hashtags
difficulty: medium
after: strings-bytes-and-runes
hints:
  - '`strings.Fields` splits a message into words on any run of spaces, tabs or newlines. Then clean each word: `strings.TrimRight(word, ".,!?")` drops trailing punctuation and `strings.ToLower` normalises the case.'
  - 'A cleaned word is a hashtag if it starts with `#` (`strings.HasPrefix`) and is longer than just `"#"`.'
  - 'To drop repeats but keep the order, append a tag to the result only if a `seen` map doesn''t have it yet, then mark it as seen.'
exercise:
  starter: |
    package main

    import "fmt"

    func hashtags(msg string) []string {
    	return nil
    }

    func main() {
    	fmt.Println(hashtags("Summer sale! #Deals on #shoes, more #deals soon."))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func hashtags(msg string) []string {
    	var tags []string
    	seen := map[string]bool{}
    	for _, word := range strings.Fields(msg) {
    		tag := strings.ToLower(strings.TrimRight(word, ".,!?"))
    		if !strings.HasPrefix(tag, "#") || tag == "#" || seen[tag] {
    			continue
    		}
    		seen[tag] = true
    		tags = append(tags, tag)
    	}
    	return tags
    }

    func main() {
    	fmt.Println(hashtags("Summer sale! #Deals on #shoes, more #deals soon."))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestHashtags(t *testing.T) {
    	tests := []struct {
    		msg  string
    		want []string
    	}{
    		{"Summer sale! #Deals on #shoes, more #deals soon.", []string{"#deals", "#shoes"}},
    		{"", nil},
    		{"no tags here", nil},
    		{"#", nil},
    		{"# #!! #?", nil},
    		{"#Go", []string{"#go"}},
    		{"#DEALS #deals #Deals", []string{"#deals"}},
    		{"#b #a #b #c #a", []string{"#b", "#a", "#c"}},
    		{"x#y and mid#tag stay out", nil},
    		{"  #one\t#two\n\n#three  ", []string{"#one", "#two", "#three"}},
    		{"Wow!!! #launch!!! #launch?", []string{"#launch"}},
    		{"#Café opens, see #CAFÉ.", []string{"#café"}},
    		{"#🎉 party time #🎉", []string{"#🎉"}},
    		{"#go#fast", []string{"#go#fast"}},
    		{"#a,b", []string{"#a,b"}},
    	}
    	for _, tt := range tests {
    		got := hashtags(tt.msg)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("hashtags(%q) = %q, want %q", tt.msg, got, tt.want)
    		}
    	}
    }
---

Textio's analytics page groups campaigns by the hashtags in their messages.
Your job is to pull the hashtags out of one message.

Write `hashtags(msg string) []string`. It returns the message's hashtags in
the order they first appear, with no repeats.

1. Split `msg` into words on whitespace (spaces, tabs, newlines).
2. Remove any trailing `.`, `,`, `!` and `?` characters from each word, and
   make it lowercase.
3. The cleaned word is a hashtag if it starts with `#` and has at least one
   character after the `#`.
4. Keep only the first copy of each hashtag. `#Deals` and `#deals` are the
   same tag.

If there are no hashtags, return an empty (or `nil`) slice.

## Examples

```
hashtags("Summer sale! #Deals on #shoes, more #deals soon.")  // ["#deals", "#shoes"]
hashtags("#Café opens, see #CAFÉ.")                           // ["#café"]
hashtags("x#y and # alone")                                   // []
```

## Constraints

- Only **trailing** punctuation is removed: `#a,b` stays `#a,b`.
- A `#` in the middle of a word doesn't start a tag.
- Messages can contain any Unicode, including accented letters and emoji.
