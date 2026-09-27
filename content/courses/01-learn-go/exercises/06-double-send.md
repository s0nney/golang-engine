---
title: Double Send
difficulty: easy
after: maps
hints:
  - 'Checking each number against every earlier one works, but it''s slow for big lists. A map can remember which numbers you''ve already seen.'
  - 'Use `seen := map[string]bool{}`. For each phone, if `seen[phone]` is already `true`, you''ve found the repeat; otherwise set `seen[phone] = true` and keep going.'
  - 'Return as soon as you find the first repeat. If the loop finishes without one, return `"", false`.'
exercise:
  starter: |
    package main

    import "fmt"

    // firstRepeat returns the first phone number in phones that has already
    // appeared earlier in the list, and true. If no number repeats, it
    // returns "" and false.
    func firstRepeat(phones []string) (string, bool) {
    	// Remember the numbers you've seen in a map.
    	return "", false
    }

    func main() {
    	batch := []string{"555-0101", "555-0199", "555-0142", "555-0199", "555-0101"}
    	fmt.Println(firstRepeat(batch)) // want: 555-0199 true
    }
  solution: |
    package main

    import "fmt"

    func firstRepeat(phones []string) (string, bool) {
    	seen := map[string]bool{}
    	for _, p := range phones {
    		if seen[p] {
    			return p, true
    		}
    		seen[p] = true
    	}
    	return "", false
    }

    func main() {
    	batch := []string{"555-0101", "555-0199", "555-0142", "555-0199", "555-0101"}
    	fmt.Println(firstRepeat(batch))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    	"time"
    )

    func TestFirstRepeat(t *testing.T) {
    	tests := []struct {
    		phones []string
    		want   string
    		wantOK bool
    	}{
    		{[]string{"555-0101", "555-0199", "555-0142", "555-0199", "555-0101"}, "555-0199", true},
    		{nil, "", false},
    		{[]string{"555-0101"}, "", false},
    		{[]string{"555-0101", "555-0101"}, "555-0101", true},
    		{[]string{"a", "b", "c"}, "", false},
    		{[]string{"a", "b", "b", "a"}, "b", true},
    		{[]string{"a", "b", "c", "a", "b"}, "a", true},
    		{[]string{"", "x", ""}, "", true},
    		{[]string{"+44 20", "+4420", "+44 20"}, "+44 20", true},
    	}
    	for _, tt := range tests {
    		got, ok := firstRepeat(tt.phones)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("firstRepeat(%q) = %q, %v, want %q, %v", tt.phones, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }

    func TestFirstRepeatLarge(t *testing.T) {
    	phones := make([]string, 100_000)
    	for i := range phones {
    		phones[i] = fmt.Sprintf("555-%06d", i)
    	}
    	phones = append(phones, "555-099999")
    	type result struct {
    		phone string
    		ok    bool
    	}
    	done := make(chan result, 1)
    	go func() {
    		phone, ok := firstRepeat(phones)
    		done <- result{phone, ok}
    	}()
    	select {
    	case r := <-done:
    		if r.phone != "555-099999" || !r.ok {
    			t.Errorf("firstRepeat(100,001 numbers) = %q, %v, want %q, true", r.phone, r.ok, "555-099999")
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("firstRepeat(100,001 numbers) took over a second: use a map instead of comparing every pair")
    	}
    }
---

Sending the same campaign twice to one person is a great way to get
unsubscribed. Before a batch goes out, Textio checks it for duplicates.

Complete `firstRepeat(phones)`. Walk through the list in order and return the
first phone number that has **already appeared earlier**, plus `true`. If
every number is unique, return `""` and `false`.

## Examples

```
firstRepeat([]string{"555-0101", "555-0199", "555-0142", "555-0199", "555-0101"})
// "555-0199", true   ("555-0199" repeats at index 3, before "555-0101" repeats at index 4)

firstRepeat([]string{"a", "b", "c"})   // "", false
```

## Constraints

- Numbers are compared exactly as written: `"+44 20"` and `"+4420"` are different.
- A batch can hold 100,000 numbers. One test uses a batch that big, where
  comparing every number with every other one takes far too long.
