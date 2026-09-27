---
title: Password Records
difficulty: medium
after: passwords-and-key-derivation
hints:
  - 'Store everything verification needs *in the record*: algorithm, iterations, salt, key. `verifyPassword` must derive with the record''s own iterations and salt, not your current defaults, or raising the default would lock everyone out.'
  - 'Parse strictly, in order: exactly 4 fields from `strings.Split(record, "$")`, the algorithm name, `strconv.Atoi` for the count (1 to `maxIterations`, else malformed), `base64.RawStdEncoding` for salt and key, and a key of exactly 32 bytes. Only then apply the policy checks (count below `minIterations`, salt under 16 bytes), then derive.'
  - 'Compare the derived key with the stored one using `hmac.Equal` (or `subtle.ConstantTimeCompare`), never `==` or `bytes.Equal`.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var (
    	ErrMismatch   = errors.New("keybox: wrong password")
    	ErrMalformed  = errors.New("keybox: malformed password record")
    	ErrWeakParams = errors.New("keybox: password record parameters too weak")
    )

    const (
    	algorithm     = "pbkdf2-sha256"
    	saltSize      = 16
    	keySize       = 32
    	minIterations = 1_000 // kept low so the tests run fast; production: 600_000
    	maxIterations = 10_000_000
    )

    // hashPassword returns a record for password in the form
    //
    //	pbkdf2-sha256$<iterations>$<salt>$<key>
    func hashPassword(password string, iterations int) (string, error) {
    	return "", nil
    }

    // verifyPassword checks password against a record made by hashPassword.
    func verifyPassword(record, password string) error {
    	return nil
    }

    func main() {
    	rec, err := hashPassword("correct horse battery staple", 2_000)
    	fmt.Println(rec, err)
    	fmt.Println("right password:", verifyPassword(rec, "correct horse battery staple"))
    	fmt.Println("wrong password:", verifyPassword(rec, "Correct horse battery staple"))
    }
  solution: |
    package main

    import (
    	"crypto/hmac"
    	"crypto/pbkdf2"
    	"crypto/rand"
    	"crypto/sha256"
    	"encoding/base64"
    	"errors"
    	"fmt"
    	"strconv"
    	"strings"
    )

    var (
    	ErrMismatch   = errors.New("keybox: wrong password")
    	ErrMalformed  = errors.New("keybox: malformed password record")
    	ErrWeakParams = errors.New("keybox: password record parameters too weak")
    )

    const (
    	algorithm     = "pbkdf2-sha256"
    	saltSize      = 16
    	keySize       = 32
    	minIterations = 1_000 // kept low so the tests run fast; production: 600_000
    	maxIterations = 10_000_000
    )

    var b64 = base64.RawStdEncoding

    // hashPassword returns a record for password in the form
    //
    //	pbkdf2-sha256$<iterations>$<salt>$<key>
    func hashPassword(password string, iterations int) (string, error) {
    	if iterations < minIterations || iterations > maxIterations {
    		return "", ErrWeakParams
    	}
    	salt := make([]byte, saltSize)
    	rand.Read(salt)
    	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, keySize)
    	if err != nil {
    		return "", err
    	}
    	return strings.Join([]string{algorithm, strconv.Itoa(iterations), b64.EncodeToString(salt), b64.EncodeToString(key)}, "$"), nil
    }

    // verifyPassword checks password against a record made by hashPassword.
    func verifyPassword(record, password string) error {
    	parts := strings.Split(record, "$")
    	if len(parts) != 4 || parts[0] != algorithm {
    		return ErrMalformed
    	}
    	iterations, err := strconv.Atoi(parts[1])
    	if err != nil || iterations < 1 || iterations > maxIterations {
    		return ErrMalformed
    	}
    	salt, err := b64.DecodeString(parts[2])
    	if err != nil {
    		return ErrMalformed
    	}
    	want, err := b64.DecodeString(parts[3])
    	if err != nil || len(want) != keySize {
    		return ErrMalformed
    	}
    	if iterations < minIterations || len(salt) < saltSize {
    		return ErrWeakParams
    	}
    	got, err := pbkdf2.Key(sha256.New, password, salt, iterations, keySize)
    	if err != nil {
    		return ErrMalformed
    	}
    	if !hmac.Equal(got, want) {
    		return ErrMismatch
    	}
    	return nil
    }

    func main() {
    	rec, err := hashPassword("correct horse battery staple", 2_000)
    	fmt.Println(rec, err)
    	fmt.Println("right password:", verifyPassword(rec, "correct horse battery staple"))
    	fmt.Println("wrong password:", verifyPassword(rec, "Correct horse battery staple"))
    }
  tests: |
    package main

    import (
    	"crypto/pbkdf2"
    	"crypto/sha256"
    	"encoding/base64"
    	"errors"
    	"strconv"
    	"strings"
    	"testing"
    )

    const pw = "correct horse battery staple"

    func mustHash(t *testing.T, password string, iterations int) string {
    	t.Helper()
    	rec, err := hashPassword(password, iterations)
    	if err != nil {
    		t.Fatalf("hashPassword(%q, %d) returned error %v", password, iterations, err)
    	}
    	return rec
    }

    // fields splits a record into its four fields or stops the test.
    func fields(t *testing.T, rec string) []string {
    	t.Helper()
    	p := strings.Split(rec, "$")
    	if len(p) != 4 {
    		t.Fatalf("record %q should have 4 fields separated by $", rec)
    	}
    	return p
    }

    func TestRecordFormat(t *testing.T) {
    	rec := mustHash(t, pw, 1_500)
    	parts := strings.Split(rec, "$")
    	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" || parts[1] != "1500" {
    		t.Fatalf("hashPassword(%q, 1500) = %q, want pbkdf2-sha256$1500$<salt>$<key>", pw, rec)
    	}
    	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
    	if err != nil || len(salt) != saltSize {
    		t.Fatalf("salt field %q should be %d bytes in unpadded standard base64 (err %v, %d bytes)", parts[2], saltSize, err, len(salt))
    	}
    	want, _ := pbkdf2.Key(sha256.New, pw, salt, 1_500, keySize)
    	if parts[3] != base64.RawStdEncoding.EncodeToString(want) {
    		t.Errorf("key field = %q, want PBKDF2-HMAC-SHA256(password, salt, 1500, 32) = %q", parts[3], base64.RawStdEncoding.EncodeToString(want))
    	}
    }

    func TestFreshSalt(t *testing.T) {
    	a, b := mustHash(t, pw, 1_000), mustHash(t, pw, 1_000)
    	if a == b {
    		t.Errorf("hashing the same password twice gave the same record %q: use a fresh random salt each time", a)
    	}
    }

    func TestVerify(t *testing.T) {
    	for _, iter := range []int{1_000, 1_500, 3_000} {
    		rec := mustHash(t, pw, iter)
    		if err := verifyPassword(rec, pw); err != nil {
    			t.Errorf("verifyPassword(record with %d iterations, right password) = %v, want nil", iter, err)
    		}
    		for _, wrong := range []string{"", "Correct horse battery staple", pw + " ", pw[:len(pw)-1]} {
    			if err := verifyPassword(rec, wrong); !errors.Is(err, ErrMismatch) {
    				t.Errorf("verifyPassword(record, %q) = %v, want ErrMismatch", wrong, err)
    			}
    		}
    	}
    	if err := verifyPassword(mustHash(t, "", 1_000), ""); err != nil {
    		t.Errorf("an empty password should still round-trip, got %v", err)
    	}
    }

    func TestTamperedRecord(t *testing.T) {
    	rec := mustHash(t, pw, 2_000)
    	parts := fields(t, rec)
    	flip := func(field string) string {
    		b, _ := base64.RawStdEncoding.DecodeString(field)
    		b[0] ^= 0x01
    		return base64.RawStdEncoding.EncodeToString(b)
    	}
    	tampered := []struct {
    		name   string
    		fields []string
    	}{
    		{"iterations changed", []string{parts[0], "2001", parts[2], parts[3]}},
    		{"salt changed", []string{parts[0], parts[1], flip(parts[2]), parts[3]}},
    		{"key changed", []string{parts[0], parts[1], parts[2], flip(parts[3])}},
    	}
    	for _, tt := range tampered {
    		if err := verifyPassword(strings.Join(tt.fields, "$"), pw); !errors.Is(err, ErrMismatch) {
    			t.Errorf("%s: verifyPassword = %v, want ErrMismatch", tt.name, err)
    		}
    	}
    }

    func TestMalformed(t *testing.T) {
    	rec := mustHash(t, pw, 1_000)
    	p := fields(t, rec)
    	short, _ := base64.RawStdEncoding.DecodeString(p[3])
    	bad := []struct{ name, record string }{
    		{"empty", ""},
    		{"too few fields", strings.Join(p[:3], "$")},
    		{"too many fields", rec + "$extra"},
    		{"unknown algorithm", "pbkdf2-sha1$" + strings.Join(p[1:], "$")},
    		{"iterations not a number", strings.Join([]string{p[0], "lots", p[2], p[3]}, "$")},
    		{"negative iterations", strings.Join([]string{p[0], "-1000", p[2], p[3]}, "$")},
    		{"absurd iterations", strings.Join([]string{p[0], "999999999999", p[2], p[3]}, "$")},
    		{"salt not base64", strings.Join([]string{p[0], p[1], "!!notbase64!!", p[3]}, "$")},
    		{"key not base64", strings.Join([]string{p[0], p[1], p[2], "%%%"}, "$")},
    		{"truncated key", strings.Join([]string{p[0], p[1], p[2], base64.RawStdEncoding.EncodeToString(short[:16])}, "$")},
    		{"empty key", strings.Join([]string{p[0], p[1], p[2], ""}, "$")},
    	}
    	for _, b := range bad {
    		if err := verifyPassword(b.record, pw); !errors.Is(err, ErrMalformed) {
    			t.Errorf("%s: verifyPassword(%q) = %v, want ErrMalformed", b.name, b.record, err)
    		}
    	}
    }

    func TestWeakParams(t *testing.T) {
    	for _, iter := range []int{0, 1, 999, maxIterations + 1} {
    		if rec, err := hashPassword(pw, iter); !errors.Is(err, ErrWeakParams) {
    			t.Errorf("hashPassword(%d iterations) = %q, %v, want ErrWeakParams", iter, rec, err)
    		}
    	}
    	// A record an attacker downgraded to 1 iteration and a 4-byte salt,
    	// complete with a matching key, must still be refused.
    	for _, c := range []struct {
    		iter int
    		salt string
    	}{{1, strings.Repeat("s", 16)}, {1_000, "salt"}} {
    		key, _ := pbkdf2.Key(sha256.New, pw, []byte(c.salt), c.iter, keySize)
    		rec := strings.Join([]string{"pbkdf2-sha256", strconv.Itoa(c.iter), base64.RawStdEncoding.EncodeToString([]byte(c.salt)), base64.RawStdEncoding.EncodeToString(key)}, "$")
    		if err := verifyPassword(rec, pw); !errors.Is(err, ErrWeakParams) {
    			t.Errorf("record with %d iterations and a %d-byte salt: verifyPassword = %v, want ErrWeakParams", c.iter, len(c.salt), err)
    		}
    	}
    }
---

Keybox's web dashboard lets team admins sign in with a password. The server must
never store the password itself, only a **password record** it can check a login
against. Records stay in the database for years, so each one carries its own
parameters: a record made with today's iteration count must still verify after
you raise the default.

Implement:

- `hashPassword(password, iterations)` returns a record
  `pbkdf2-sha256$<iterations>$<salt>$<key>` where `salt` is 16 fresh random bytes,
  `key` is PBKDF2-HMAC-SHA256(password, salt, iterations, 32 bytes), and both are
  encoded with `base64.RawStdEncoding`. If `iterations` is outside
  `[minIterations, maxIterations]`, return `ErrWeakParams`.
- `verifyPassword(record, password)` returns `nil` if the password matches, or:
  - `ErrMalformed` if the record isn't exactly four `$`-separated fields, the
    algorithm isn't `pbkdf2-sha256`, the iteration count isn't a whole number from 1 to
    `maxIterations`, salt or key isn't valid base64, or the key isn't exactly 32 bytes;
  - `ErrWeakParams` if it's well-formed but has fewer than `minIterations` iterations
    or a salt shorter than 16 bytes;
  - `ErrMismatch` if the password is wrong.

## Example

```go
rec, _ := hashPassword("correct horse battery staple", 2000)
// pbkdf2-sha256$2000$ttbYLNBioPmK1q14ePQzAA$GfKq9Jv3Xf9uuKgZE7xpQQl2...  (random salt)
verifyPassword(rec, "correct horse battery staple")  // nil
verifyPassword(rec, "Correct horse battery staple")  // ErrMismatch
verifyPassword("pbkdf2-sha256$1$...", "...")          // ErrWeakParams, even if the key matches
```

## Constraints

- `minIterations` is 1,000 here so the tests run quickly. A production deployment of
  PBKDF2-HMAC-SHA256 should use **at least 600,000** iterations (OWASP's current
  recommendation), or a memory-hard KDF.
- `maxIterations` caps the work a single record can demand, so a planted record with
  a billion iterations can't tie up your login server.
- The tests check the record format against `pbkdf2.Key`, fresh salts, records at
  several iteration counts, tampered iterations/salt/key, truncated and malformed
  records, and downgraded records with a matching key.
- Compare keys in constant time.
