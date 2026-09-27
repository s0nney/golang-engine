---
title: Password Upgrade
difficulty: medium
after: authentication
hints:
  - 'Parse with `strings.Split(stored, "$")`: you need exactly 4 parts, `parts[0] == "pbkdf2-sha256"`, a positive iteration count from `strconv.Atoi`, and two parts that decode with `base64.RawStdEncoding`. Every failure returns `ErrMalformedHash` (wrap it with `%w` if you like).'
  - 'Re-derive the key with **the stored** salt and iteration count: `pbkdf2.Key(sha256.New, password, salt, iter, len(want))`, and compare with `subtle.ConstantTimeCompare`. Only after the password checks out do you look at whether `iter < minIterations`.'
  - 'The upgrade is just `hashPassword(password, minIterations)`: a fresh salt and the new cost. You can only do it at login, because it''s the one moment you have the plain password.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var ErrMalformedHash = errors.New("malformed password hash")

    func hashPassword(password string, iterations int) (string, error) {
    	return "", nil
    }

    func verifyAndUpgrade(password, stored string, minIterations int) (ok bool, upgraded string, err error) {
    	return false, "", nil
    }

    func main() {
    	old, _ := hashPassword("cheese123", 1_000)
    	fmt.Println(old)
    	ok, upgraded, err := verifyAndUpgrade("cheese123", old, 5_000)
    	fmt.Println(ok, upgraded, err)
    	ok, upgraded, err = verifyAndUpgrade("cheese124", old, 5_000)
    	fmt.Println(ok, upgraded, err)
    }
  solution: |
    package main

    import (
    	"crypto/pbkdf2"
    	"crypto/rand"
    	"crypto/sha256"
    	"crypto/subtle"
    	"encoding/base64"
    	"errors"
    	"fmt"
    	"strconv"
    	"strings"
    )

    var ErrMalformedHash = errors.New("malformed password hash")

    var b64 = base64.RawStdEncoding

    const algorithm = "pbkdf2-sha256"

    func hashPassword(password string, iterations int) (string, error) {
    	salt := make([]byte, 16)
    	rand.Read(salt)
    	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, 32)
    	if err != nil {
    		return "", err
    	}
    	return fmt.Sprintf("%s$%d$%s$%s", algorithm, iterations, b64.EncodeToString(salt), b64.EncodeToString(key)), nil
    }

    func verifyAndUpgrade(password, stored string, minIterations int) (ok bool, upgraded string, err error) {
    	parts := strings.Split(stored, "$")
    	if len(parts) != 4 || parts[0] != algorithm {
    		return false, "", ErrMalformedHash
    	}
    	iter, err := strconv.Atoi(parts[1])
    	if err != nil || iter < 1 {
    		return false, "", ErrMalformedHash
    	}
    	salt, err := b64.DecodeString(parts[2])
    	if err != nil || len(salt) == 0 {
    		return false, "", ErrMalformedHash
    	}
    	want, err := b64.DecodeString(parts[3])
    	if err != nil || len(want) == 0 {
    		return false, "", ErrMalformedHash
    	}
    	got, err := pbkdf2.Key(sha256.New, password, salt, iter, len(want))
    	if err != nil {
    		return false, "", err
    	}
    	if subtle.ConstantTimeCompare(got, want) != 1 {
    		return false, "", nil
    	}
    	if iter >= minIterations {
    		return true, "", nil
    	}
    	upgraded, err = hashPassword(password, minIterations)
    	if err != nil {
    		// The password was right; keep the old hash rather than fail the login.
    		return true, "", nil
    	}
    	return true, upgraded, nil
    }

    func main() {
    	old, _ := hashPassword("cheese123", 1_000)
    	fmt.Println(old)
    	ok, upgraded, err := verifyAndUpgrade("cheese123", old, 5_000)
    	fmt.Println(ok, upgraded, err)
    	ok, upgraded, err = verifyAndUpgrade("cheese124", old, 5_000)
    	fmt.Println(ok, upgraded, err)
    }
  tests: |
    package main

    import (
    	"crypto/pbkdf2"
    	"crypto/sha256"
    	"encoding/base64"
    	"errors"
    	"fmt"
    	"os"
    	"strings"
    	"testing"
    )

    var b64test = base64.RawStdEncoding

    // makeHash builds a stored hash the way Squeak's old code did.
    func makeHash(password string, iter int, salt string) string {
    	key, _ := pbkdf2.Key(sha256.New, password, []byte(salt), iter, 32)
    	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", iter, b64test.EncodeToString([]byte(salt)), b64test.EncodeToString(key))
    }

    func TestHashPassword(t *testing.T) {
    	a, err := hashPassword("cheese123", 1500)
    	if err != nil {
    		t.Fatalf("hashPassword returned error %v", err)
    	}
    	parts := strings.Split(a, "$")
    	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" || parts[1] != "1500" {
    		t.Fatalf("hashPassword(\"cheese123\", 1500) = %q, want pbkdf2-sha256$1500$<salt>$<key>", a)
    	}
    	salt, err1 := b64test.DecodeString(parts[2])
    	key, err2 := b64test.DecodeString(parts[3])
    	if err1 != nil || err2 != nil || len(salt) != 16 || len(key) != 32 {
    		t.Fatalf("hashPassword: salt and key must be unpadded standard base64 of 16 and 32 bytes, got %q", a)
    	}
    	want, _ := pbkdf2.Key(sha256.New, "cheese123", salt, 1500, 32)
    	if string(key) != string(want) {
    		t.Errorf("hashPassword: key isn't pbkdf2.Key(sha256.New, password, salt, 1500, 32)")
    	}
    	b, _ := hashPassword("cheese123", 1500)
    	if a == b {
    		t.Errorf("hashPassword gave the same result twice: every hash needs a fresh random salt from crypto/rand")
    	}
    }

    func TestVerify(t *testing.T) {
    	stored := makeHash("cheese123", 2000, "0123456789abcdef")
    	tests := []struct {
    		password string
    		wantOK   bool
    	}{
    		{"cheese123", true},
    		{"cheese124", false},
    		{"Cheese123", false},
    		{"cheese12", false},
    		{"", false},
    	}
    	for _, tt := range tests {
    		ok, upgraded, err := verifyAndUpgrade(tt.password, stored, 2000)
    		if ok != tt.wantOK || upgraded != "" || err != nil {
    			t.Errorf("verifyAndUpgrade(%q, <2000 iterations>, 2000) = %v, %q, %v, want %v, \"\", nil", tt.password, ok, upgraded, err, tt.wantOK)
    		}
    	}
    	if ok, up, err := verifyAndUpgrade("cheese123", stored, 1000); !ok || up != "" || err != nil {
    		t.Errorf("stored with 2000 iterations, minimum 1000: got %v, %q, %v, want true, \"\", nil (no upgrade needed)", ok, up, err)
    	}
    }

    func TestUpgrade(t *testing.T) {
    	stored := makeHash("cheese123", 1000, "saltysaltysalty!")
    	ok, up, err := verifyAndUpgrade("cheese123", stored, 3000)
    	if !ok || err != nil {
    		t.Fatalf("right password, weak hash: got ok=%v err=%v, want true, nil", ok, err)
    	}
    	if !strings.HasPrefix(up, "pbkdf2-sha256$3000$") {
    		t.Fatalf("right password, 1000 < 3000 iterations: upgraded = %q, want a new pbkdf2-sha256$3000$... hash", up)
    	}
    	if ok, again, err := verifyAndUpgrade("cheese123", up, 3000); !ok || again != "" || err != nil {
    		t.Errorf("verifying the upgraded hash: got %v, %q, %v, want true, \"\", nil", ok, again, err)
    	}
    	if ok, _, _ := verifyAndUpgrade("cheese124", up, 3000); ok {
    		t.Errorf("the upgraded hash accepts the wrong password")
    	}
    	if ok, wrongUp, err := verifyAndUpgrade("nope", stored, 3000); ok || wrongUp != "" || err != nil {
    		t.Errorf("wrong password, weak hash: got %v, %q, %v, want false, \"\", nil (never upgrade without a correct password)", ok, wrongUp, err)
    	}
    }

    func TestMalformed(t *testing.T) {
    	good := makeHash("cheese123", 1000, "0123456789abcdef")
    	parts := strings.Split(good, "$")
    	bad := map[string]string{
    		"empty":           "",
    		"plain text":      "cheese123",
    		"three parts":     strings.Join(parts[:3], "$"),
    		"five parts":      good + "$extra",
    		"other algorithm": "bcrypt$" + strings.Join(parts[1:], "$"),
    		"zero iterations": parts[0] + "$0$" + parts[2] + "$" + parts[3],
    		"negative":        parts[0] + "$-5$" + parts[2] + "$" + parts[3],
    		"not a number":    parts[0] + "$lots$" + parts[2] + "$" + parts[3],
    		"bad salt":        parts[0] + "$1000$!!!$" + parts[3],
    		"bad key":         parts[0] + "$1000$" + parts[2] + "$***",
    		"empty key":       parts[0] + "$1000$" + parts[2] + "$",
    	}
    	for name, stored := range bad {
    		ok, up, err := verifyAndUpgrade("cheese123", stored, 1000)
    		if ok || up != "" || !errors.Is(err, ErrMalformedHash) {
    			t.Errorf("%s (%q): got %v, %q, %v, want false, \"\", ErrMalformedHash", name, stored, ok, up, err)
    		}
    		if err != nil && strings.Contains(err.Error(), "cheese123") {
    			t.Errorf("%s: the error %q contains the password; errors end up in logs", name, err)
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
    		t.Errorf("compare the derived key with subtle.ConstantTimeCompare (or hmac.Equal), not == or bytes.Equal")
    	}
    }
---

Squeak launched with 1,000 PBKDF2 iterations per password hash. That was far too
few, and the new setting is hundreds of times higher. But you can't rehash the
old passwords without knowing them. The standard trick: **upgrade on login**.
When a mouse logs in successfully with a weak hash, you have their plain
password for a moment, so you hash it again at the new cost and store that.

Stored hashes look like this (salt and key in unpadded standard base64):

```
pbkdf2-sha256$<iterations>$<salt>$<key>
```

Implement two functions:

**`hashPassword(password, iterations)`** returns a new stored hash: a fresh
16-byte salt from `crypto/rand`, and a 32-byte key from
`pbkdf2.Key(sha256.New, password, salt, iterations, 32)`.

**`verifyAndUpgrade(password, stored, minIterations)`** returns:

| Situation | `ok` | `upgraded` | `err` |
|---|---|---|---|
| `stored` can't be parsed | `false` | `""` | `ErrMalformedHash` |
| wrong password | `false` | `""` | `nil` |
| right password, stored iterations ≥ `minIterations` | `true` | `""` | `nil` |
| right password, stored iterations < `minIterations` | `true` | a new hash with `minIterations` | `nil` |

## Example

```go
old, _ := hashPassword("cheese123", 1_000)
verifyAndUpgrade("cheese123", old, 5_000)  // true, "pbkdf2-sha256$5000$…", nil
verifyAndUpgrade("cheese124", old, 5_000)  // false, "", nil
verifyAndUpgrade("cheese123", "hunter2", 5_000)  // false, "", ErrMalformedHash
```

## Constraints

- A stored hash is malformed unless it has exactly 4 `$`-separated parts, the
  algorithm `pbkdf2-sha256`, a positive iteration count, and a salt and key
  that decode as base64 and aren't empty. `errors.Is(err, ErrMalformedHash)`
  must hold.
- Re-derive the key with the stored salt and iteration count, and compare in
  constant time. The tests check for `subtle.ConstantTimeCompare` (or
  `hmac.Equal`).
- Never upgrade after a wrong password, and never put the password in an error.
- The tests use small iteration counts so they run fast. Production would use
  600,000.
