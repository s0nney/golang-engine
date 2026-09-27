---
title: Scheduled Send
difficulty: medium
after: errors
hints:
  - '`strings.Cut(s, ":")` splits `s` at the first colon and tells you whether there was one. If the minute part still contains a `:`, there were too many.'
  - 'Check the shape before converting: the hour has 1 or 2 characters, the minute exactly 2, and every character is between `''0''` and `''9''`. That rules out signs like `+1` and `-5`, which `strconv.Atoi` would happily accept.'
  - 'Wrap the sentinel so callers can still find it: `fmt.Errorf("parse %q: %w", s, ErrFormat)`. `%w` (not `%v`) is what lets `errors.Is` see through the wrapping.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var (
    	ErrFormat = errors.New("time must look like HH:MM")
    	ErrRange  = errors.New("time out of range")
    )

    func parseSendTime(s string) (int, int, error) {
    	return 0, 0, nil
    }

    func main() {
    	for _, s := range []string{"9:05", "24:00", "12-30"} {
    		h, m, err := parseSendTime(s)
    		fmt.Println(h, m, err)
    	}
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strconv"
    	"strings"
    )

    var (
    	ErrFormat = errors.New("time must look like HH:MM")
    	ErrRange  = errors.New("time out of range")
    )

    func allDigits(s string) bool {
    	for _, r := range s {
    		if r < '0' || r > '9' {
    			return false
    		}
    	}
    	return true
    }

    func parseSendTime(s string) (int, int, error) {
    	hh, mm, ok := strings.Cut(strings.TrimSpace(s), ":")
    	if !ok || len(hh) < 1 || len(hh) > 2 || len(mm) != 2 || !allDigits(hh) || !allDigits(mm) {
    		return 0, 0, fmt.Errorf("parse %q: %w", s, ErrFormat)
    	}
    	hour, _ := strconv.Atoi(hh)
    	minute, _ := strconv.Atoi(mm)
    	if hour > 23 || minute > 59 {
    		return 0, 0, fmt.Errorf("parse %q: %w", s, ErrRange)
    	}
    	return hour, minute, nil
    }

    func main() {
    	for _, s := range []string{"9:05", "24:00", "12-30"} {
    		h, m, err := parseSendTime(s)
    		fmt.Println(h, m, err)
    	}
    }
  tests: |
    package main

    import (
    	"errors"
    	"strings"
    	"testing"
    )

    func TestParseSendTimeValid(t *testing.T) {
    	tests := []struct {
    		s            string
    		hour, minute int
    	}{
    		{"9:05", 9, 5},
    		{"09:05", 9, 5},
    		{"0:00", 0, 0},
    		{"00:00", 0, 0},
    		{"23:59", 23, 59},
    		{"  12:30\n", 12, 30},
    	}
    	for _, tt := range tests {
    		h, m, err := parseSendTime(tt.s)
    		if err != nil || h != tt.hour || m != tt.minute {
    			t.Errorf("parseSendTime(%q) = %d, %d, %v, want %d, %d, nil", tt.s, h, m, err, tt.hour, tt.minute)
    		}
    	}
    }

    func TestParseSendTimeErrors(t *testing.T) {
    	tests := []struct {
    		s    string
    		want error
    	}{
    		{"24:00", ErrRange},
    		{"12:60", ErrRange},
    		{"99:99", ErrRange},
    		{"", ErrFormat},
    		{"   ", ErrFormat},
    		{"12-30", ErrFormat},
    		{"12:5", ErrFormat},
    		{"12:305", ErrFormat},
    		{"123:00", ErrFormat},
    		{":30", ErrFormat},
    		{"12:", ErrFormat},
    		{"1:2:3", ErrFormat},
    		{"12:30:00", ErrFormat},
    		{"+1:30", ErrFormat},
    		{"-1:30", ErrFormat},
    		{"1:-5", ErrFormat},
    		{"ab:cd", ErrFormat},
    		{"12 :30", ErrFormat},
    		{"１２:３０", ErrFormat},
    		{"9:05pm", ErrFormat},
    	}
    	for _, tt := range tests {
    		h, m, err := parseSendTime(tt.s)
    		if !errors.Is(err, tt.want) {
    			t.Errorf("parseSendTime(%q) error = %v, want an error wrapping %q", tt.s, err, tt.want)
    			continue
    		}
    		if h != 0 || m != 0 {
    			t.Errorf("parseSendTime(%q) = %d, %d with an error, want 0, 0", tt.s, h, m)
    		}
    		if tt.s != "" && !strings.Contains(err.Error(), tt.s) {
    			t.Errorf("parseSendTime(%q) error %q should mention the text that failed", tt.s, err.Error())
    		}
    		if err == tt.want {
    			t.Errorf("parseSendTime(%q) returned the bare sentinel %q; wrap it with fmt.Errorf and %%w to add the input", tt.s, err)
    		}
    	}
    }
---

Textio customers can schedule a message for later by typing a time like
`9:05` or `18:30`. Your parser turns that text into numbers, or explains what
went wrong.

Write `parseSendTime(s string) (hour, minute int, err error)`:

1. Ignore spaces and newlines around `s`.
2. The text must be `H:MM` or `HH:MM`: one or two digits, a colon, then
   exactly two digits. Only the characters `0` to `9` count as digits. If it
   isn't, return an error wrapping `ErrFormat`.
3. The hour must be `0` to `23` and the minute `0` to `59`. If not, return an
   error wrapping `ErrRange`.
4. Otherwise return the hour, the minute and a `nil` error.

When there's an error, return `0, 0` with it. The error must **wrap** the
sentinel (so `errors.Is` finds it) and its message must include the
original text `s`, for example `parse "24:00": time out of range`.

## Examples

```
parseSendTime("9:05")    // 9, 5, nil
parseSendTime(" 23:59 ") // 23, 59, nil
parseSendTime("24:00")   // 0, 0, parse "24:00": time out of range
parseSendTime("12:5")    // 0, 0, parse "12:5": time must look like HH:MM
```

## Constraints

- Keep `ErrFormat` and `ErrRange` as they are: the tests check for them with
  `errors.Is`.
- Signs aren't allowed: `+1:30` and `1:-5` are format errors.
