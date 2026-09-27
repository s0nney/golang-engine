---
title: API Key Check
difficulty: easy
after: authentication
hints:
  - '`sha256.Sum256([]byte(presented))` returns a `[32]byte`. Turn the stored hex string back into bytes with `hex.DecodeString`; if that fails, the stored value is broken and nothing can match it.'
  - 'Compare the two byte slices with `subtle.ConstantTimeCompare(a, b) == 1` (from `crypto/subtle`), not `bytes.Equal` or `==`. It takes the same time wherever the first difference is, so the response time gives nothing away.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    )

    // keyMatches reports whether presented is the API key whose SHA-256
    // hash is storedHash (64 lowercase hex characters).
    // An empty presented key never matches, and neither does a storedHash
    // that isn't valid hex.
    func keyMatches(presented, storedHash string) bool {
    	// 1. Hash presented with SHA-256.
    	// 2. Decode storedHash from hex.
    	// 3. Compare them in constant time.
    	return false
    }

    func main() {
    	stored := "3c456a58a1ca490e0340299563556fa04c1cbecae409bd6a47f666a091c71ef4"
    	fmt.Println(keyMatches("sqk_live_7Hc2", stored)) // want: true
    	fmt.Println(keyMatches("sqk_live_7Hc3", stored)) // want: false
    }
  solution: |
    package main

    import (
    	"crypto/sha256"
    	"crypto/subtle"
    	"encoding/hex"
    	"fmt"
    )

    func keyMatches(presented, storedHash string) bool {
    	if presented == "" {
    		return false
    	}
    	want, err := hex.DecodeString(storedHash)
    	if err != nil {
    		return false
    	}
    	got := sha256.Sum256([]byte(presented))
    	return subtle.ConstantTimeCompare(got[:], want) == 1
    }

    func main() {
    	stored := "3c456a58a1ca490e0340299563556fa04c1cbecae409bd6a47f666a091c71ef4"
    	fmt.Println(keyMatches("sqk_live_7Hc2", stored))
    	fmt.Println(keyMatches("sqk_live_7Hc3", stored))
    }
  tests: |
    package main

    import (
    	"crypto/sha256"
    	"encoding/hex"
    	"os"
    	"strings"
    	"testing"
    )

    func hashOf(key string) string {
    	sum := sha256.Sum256([]byte(key))
    	return hex.EncodeToString(sum[:])
    }

    func TestKeyMatches(t *testing.T) {
    	stored := hashOf("sqk_live_7Hc2")
    	tests := []struct {
    		presented, stored string
    		want              bool
    	}{
    		{"sqk_live_7Hc2", stored, true},
    		{"sqk_live_7Hc3", stored, false},
    		{"sqk_live_7Hc", stored, false},
    		{"sqk_live_7Hc2 ", stored, false},
    		{"SQK_LIVE_7HC2", stored, false},
    		{"", stored, false},
    		{"", hashOf(""), false},
    		{"sqk_live_7Hc2", stored[:62], false},
    		{"sqk_live_7Hc2", "zz" + stored[2:], false},
    		{"sqk_live_7Hc2", "", false},
    		{stored, stored, false},
    		{"another-key", hashOf("another-key"), true},
    	}
    	for _, tt := range tests {
    		if got := keyMatches(tt.presented, tt.stored); got != tt.want {
    			t.Errorf("keyMatches(%q, %q) = %v, want %v", tt.presented, tt.stored, got, tt.want)
    		}
    	}
    }

    func TestConstantTime(t *testing.T) {
    	src, err := os.ReadFile("main.go")
    	if err != nil {
    		t.Skip("can't read main.go")
    	}
    	var code strings.Builder
    	for line := range strings.Lines(string(src)) {
    		before, _, _ := strings.Cut(line, "//")
    		code.WriteString(before)
    	}
    	if !strings.Contains(code.String(), "subtle.ConstantTimeCompare") && !strings.Contains(code.String(), "hmac.Equal") {
    		t.Errorf("keyMatches must compare hashes with subtle.ConstantTimeCompare (or hmac.Equal), so timing can't leak how much of a guess was right")
    	}
    }
---

Squeak lets developers build bots with **API keys** like `sqk_live_7Hc2...`.
Like passwords, keys are secrets, so Squeak never stores them. It shows a new
key to the developer once and keeps only its **SHA-256 hash**, as 64 hex
characters. If the database leaks, the hashes are useless to an attacker.

Complete `keyMatches(presented, storedHash)`. It returns `true` only if the
SHA-256 hash of `presented` equals the hash that `storedHash` encodes:

- An empty `presented` key never matches.
- If `storedHash` isn't valid hex, nothing matches it.
- Compare the hashes in **constant time**.

## Examples

```go
stored := hex.EncodeToString(sha256Of("sqk_live_7Hc2"))
keyMatches("sqk_live_7Hc2", stored)  // true
keyMatches("sqk_live_7Hc3", stored)  // false
keyMatches("", stored)               // false
keyMatches("sqk_live_7Hc2", "zz…")   // false: storedHash isn't hex
```

## Constraints

- Why hash with plain SHA-256 here, when passwords need a slow hash? API keys
  are long random strings, not words a human picked, so there's no dictionary
  to try. A fast hash is enough, and it keeps every request quick.
- The tests check that your code uses `subtle.ConstantTimeCompare` (or
  `hmac.Equal`).
