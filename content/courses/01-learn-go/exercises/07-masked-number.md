---
title: Masked Number
difficulty: medium
after: strings-bytes-and-runes
hints:
  - 'You can''t decide whether a digit stays visible until you know how many digits come **after** it. So make two passes: first count all the digits, then build the result.'
  - 'A rune `r` is a digit when `r >= ''0'' && r <= ''9''`. Ranging over a string gives you runes, so emoji and accented letters come through in one piece.'
  - 'In the second pass, count digits again as you go. While that count is at most `total - 4`, write `''*''` to a `strings.Builder` instead of the digit. Everything else is written unchanged.'
exercise:
  starter: |
    package main

    import "fmt"

    func maskPhone(phone string) string {
    	return phone
    }

    func main() {
    	fmt.Println(maskPhone("+1 (555) 123-4567"))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func isDigit(r rune) bool {
    	return r >= '0' && r <= '9'
    }

    func maskPhone(phone string) string {
    	total := 0
    	for _, r := range phone {
    		if isDigit(r) {
    			total++
    		}
    	}
    	var b strings.Builder
    	seen := 0
    	for _, r := range phone {
    		if isDigit(r) {
    			seen++
    			if seen <= total-4 {
    				b.WriteRune('*')
    				continue
    			}
    		}
    		b.WriteRune(r)
    	}
    	return b.String()
    }

    func main() {
    	fmt.Println(maskPhone("+1 (555) 123-4567"))
    }
  tests: |
    package main

    import "testing"

    func TestMaskPhone(t *testing.T) {
    	tests := []struct {
    		phone string
    		want  string
    	}{
    		{"+1 (555) 123-4567", "+* (***) ***-4567"},
    		{"5551234567", "******4567"},
    		{"", ""},
    		{"123", "123"},
    		{"1234", "1234"},
    		{"12345", "*2345"},
    		{"call me", "call me"},
    		{"📞 555-0199", "📞 ***-0199"},
    		{"ext. 99 ☎ 1234 5678", "ext. ** ☎ **** 5678"},
    		{"12-34-56", "**-34-56"},
    		{"Zoë: 0612 345 678", "Zoë: **** **5 678"},
    		{"٠١٢٣٤٥ 98765", "٠١٢٣٤٥ *8765"},
    	}
    	for _, tt := range tests {
    		if got := maskPhone(tt.phone); got != tt.want {
    			t.Errorf("maskPhone(%q) = %q, want %q", tt.phone, got, tt.want)
    		}
    	}
    }
---

Support agents at Textio can see a customer's phone number, but only the end
of it. Everything else is masked, the way banks show card numbers.

Write `maskPhone(phone string) string`. It returns `phone` with every digit
**except the last four** replaced by `*`. Characters that aren't digits
(spaces, `+`, brackets, dashes, letters, emoji) stay exactly where they are.
If `phone` has four digits or fewer, it's returned unchanged.

## Examples

```
maskPhone("+1 (555) 123-4567")   // "+* (***) ***-4567"
maskPhone("📞 555-0199")          // "📞 ***-0199"
maskPhone("123")                 // "123"
```

## Constraints

- Only the ASCII digits `0` to `9` count as digits. Other characters,
  including digits from other scripts such as `٣`, are left alone.
- `phone` is valid UTF-8 and may contain multi-byte characters, so don't mask
  by byte position.
