---
title: Chat Commands
difficulty: easy
after: tries
hints:
  - 'Walk down the trie one rune of `msg` at a time, starting at `&t.root`, just like `Contains` does. Stop as soon as the next rune has no child.'
  - 'Every time you reach a node with `end == true`, the runes you''ve read so far form a stored command. Remember how many **bytes** of `msg` that was (the `i` from `for i, r := range msg` plus the rune''s length, `utf8.RuneLen(r)`), and return `msg[:that]` at the end.'
exercise:
  starter: |
    package main

    import "fmt"

    type trieNode struct {
    	children map[rune]*trieNode
    	end      bool
    }

    type Trie struct {
    	root trieNode
    }

    func (t *Trie) Insert(word string) {
    	n := &t.root
    	for _, r := range word {
    		if n.children == nil {
    			n.children = map[rune]*trieNode{}
    		}
    		next, ok := n.children[r]
    		if !ok {
    			next = &trieNode{}
    			n.children[r] = next
    		}
    		n = next
    	}
    	n.end = true
    }

    // LongestPrefixOf returns the longest stored word that is a prefix of
    // msg, and true. If no stored word is a prefix of msg, it returns "", false.
    func (t *Trie) LongestPrefixOf(msg string) (string, bool) {
    	// Walk down from the root following msg's runes. Remember the
    	// last position where a stored word ended.
    	return "", false
    }

    func main() {
    	var cmds Trie
    	for _, c := range []string{"/g", "/guild", "/guildinvite", "/w"} {
    		cmds.Insert(c)
    	}
    	fmt.Println(cmds.LongestPrefixOf("/guildchat gg")) // want /guild true
    	fmt.Println(cmds.LongestPrefixOf("hello"))         // want  false
    }
  solution: |
    package main

    import (
    	"fmt"
    	"unicode/utf8"
    )

    type trieNode struct {
    	children map[rune]*trieNode
    	end      bool
    }

    type Trie struct {
    	root trieNode
    }

    func (t *Trie) Insert(word string) {
    	n := &t.root
    	for _, r := range word {
    		if n.children == nil {
    			n.children = map[rune]*trieNode{}
    		}
    		next, ok := n.children[r]
    		if !ok {
    			next = &trieNode{}
    			n.children[r] = next
    		}
    		n = next
    	}
    	n.end = true
    }

    func (t *Trie) LongestPrefixOf(msg string) (string, bool) {
    	n := &t.root
    	best, found := 0, n.end
    	for i, r := range msg {
    		n = n.children[r]
    		if n == nil {
    			break
    		}
    		if n.end {
    			best, found = i+utf8.RuneLen(r), true
    		}
    	}
    	return msg[:best], found
    }

    func main() {
    	var cmds Trie
    	for _, c := range []string{"/g", "/guild", "/guildinvite", "/w"} {
    		cmds.Insert(c)
    	}
    	fmt.Println(cmds.LongestPrefixOf("/guildchat gg"))
    	fmt.Println(cmds.LongestPrefixOf("hello"))
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    	"time"
    )

    func TestLongestPrefixOf(t *testing.T) {
    	var cmds Trie
    	for _, c := range []string{"/g", "/guild", "/guildinvite", "/w", "/émote", "!roll"} {
    		cmds.Insert(c)
    	}
    	tests := []struct {
    		msg  string
    		want string
    		ok   bool
    	}{
    		{"/guildchat gg", "/guild", true},
    		{"/guild", "/guild", true},
    		{"/guildinvite mira", "/guildinvite", true},
    		{"/guildinv", "/guild", true}, // "/guildinvite" isn't a prefix of this
    		{"/gx", "/g", true},
    		{"/g", "/g", true},
    		{"/w hi", "/w", true},
    		{"/émote dance", "/émote", true},
    		{"/émo", "", false},
    		{"!roll 20", "!roll", true},
    		{"/", "", false},
    		{"hello", "", false},
    		{"", "", false},
    		{"g/guild", "", false}, // only prefixes count
    	}
    	for _, tt := range tests {
    		got, ok := cmds.LongestPrefixOf(tt.msg)
    		if got != tt.want || ok != tt.ok {
    			t.Errorf("LongestPrefixOf(%q) = %q, %v, want %q, %v", tt.msg, got, ok, tt.want, tt.ok)
    		}
    	}
    }

    func TestLongestPrefixOfEmptyTrie(t *testing.T) {
    	var empty Trie
    	if got, ok := empty.LongestPrefixOf("/guild"); got != "" || ok {
    		t.Errorf("LongestPrefixOf(%q) on an empty trie = %q, %v, want %q, false", "/guild", got, ok, "")
    	}
    }

    func TestLongestPrefixOfLongMessages(t *testing.T) {
    	var cmds Trie
    	cmds.Insert("/shout")
    	msg := "/shout " + strings.Repeat("a", 100_000)
    	start := time.Now()
    	for range 2_000 {
    		if got, ok := cmds.LongestPrefixOf(msg); got != "/shout" || !ok {
    			t.Fatalf("LongestPrefixOf(a 100,007-byte message) = %q, %v, want %q, true", got, ok, "/shout")
    		}
    	}
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("2,000 lookups of a long message took %v: stop walking as soon as the trie has no child for the next rune", d)
    	}
    }
---

Chat messages that start with a command, like `/guild` or `/w`, are routed to a
command handler. Commands can be prefixes of each other (`/g` is short for
`/guild`), so the router picks the **longest** command that the message starts with.

The editor has the `Trie` from the tries chapter, already filled with every
command. Complete `LongestPrefixOf(msg)`. It returns the longest stored word that
is a prefix of `msg`, and `true`. If no stored word is a prefix of `msg`, it
returns `"", false`.

## Examples

```
commands: /g  /guild  /guildinvite  /w

LongestPrefixOf("/guildchat gg")  // "/guild", true
LongestPrefixOf("/guildinv")      // "/guild", true ("/guildinvite" is longer than the message)
LongestPrefixOf("/gx")            // "/g", true
LongestPrefixOf("hello")          // "", false
```

## Constraints

- Messages may contain any Unicode text, including multi-byte runes like `é`.
- Walk **only as far as the trie goes**: the cost should be O(length of the
  longest matching path), not O(length of the message). One test sends 100,000-byte
  messages.
