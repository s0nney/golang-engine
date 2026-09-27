---
title: PBKDF2 Master Keys
quiz:
  - question: Where should Keybox store the PBKDF2 salt and iteration count for Alice's vault?
    options:
      - text: Nowhere; they must be kept secret
      - text: In plain text in the vault file's header, next to the ciphertext
        correct: true
      - text: Encrypted with the master key
      - text: Hard-coded in the Keybox binary, the same for every user
    explanation: |
      The salt and iteration count aren't secret, and you need them *before* you have
      the key, so they can't be encrypted under it. Storing them in the header also
      lets you raise the iteration count for new vaults while old ones still open.
      A hard-coded salt would defeat the purpose of salting.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    const (
    	defaultIterations = 600_000
    	minIterations     = 100_000
    	saltSize          = 16
    	masterKeySize     = 32
    )

    var ErrWeakParams = errors.New("vault KDF parameters are too weak")

    // vaultHeader is stored unencrypted at the start of a Keybox vault file.
    type vaultHeader struct {
    	KDF        string // always "pbkdf2-sha256"
    	Salt       []byte
    	Iterations int
    }

    // newVaultHeader returns a header for a new vault: KDF "pbkdf2-sha256",
    // a fresh random salt of saltSize bytes and defaultIterations.
    func newVaultHeader() vaultHeader {
    	// ?
    	return vaultHeader{}
    }

    // checkHeader returns ErrWeakParams unless KDF is "pbkdf2-sha256", the
    // salt is at least saltSize bytes and Iterations is at least minIterations.
    func checkHeader(h vaultHeader) error {
    	// ?
    	return nil
    }

    // deriveMasterKey runs PBKDF2-HMAC-SHA256 over password with the header's
    // salt and iteration count, returning masterKeySize bytes. It does NOT
    // call checkHeader, so tests can use tiny iteration counts.
    func deriveMasterKey(password string, h vaultHeader) ([]byte, error) {
    	// ?
    	return nil, errors.New("not implemented")
    }

    func main() {
    	h := newVaultHeader()
    	fmt.Printf("header: %s, salt %x, %d iterations\n", h.KDF, h.Salt, h.Iterations)
    	fmt.Println("check:", checkHeader(h))

    	key, err := deriveMasterKey("correct horse battery staple", h)
    	fmt.Printf("master key: %x %v\n", key, err)
    }
  solution: |
    package main

    import (
    	"crypto/pbkdf2"
    	"crypto/rand"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    )

    const (
    	defaultIterations = 600_000
    	minIterations     = 100_000
    	saltSize          = 16
    	masterKeySize     = 32
    )

    var ErrWeakParams = errors.New("vault KDF parameters are too weak")

    type vaultHeader struct {
    	KDF        string
    	Salt       []byte
    	Iterations int
    }

    func newVaultHeader() vaultHeader {
    	salt := make([]byte, saltSize)
    	rand.Read(salt)
    	return vaultHeader{KDF: "pbkdf2-sha256", Salt: salt, Iterations: defaultIterations}
    }

    func checkHeader(h vaultHeader) error {
    	if h.KDF != "pbkdf2-sha256" || len(h.Salt) < saltSize || h.Iterations < minIterations {
    		return ErrWeakParams
    	}
    	return nil
    }

    func deriveMasterKey(password string, h vaultHeader) ([]byte, error) {
    	return pbkdf2.Key(sha256.New, password, h.Salt, h.Iterations, masterKeySize)
    }

    func main() {
    	h := newVaultHeader()
    	fmt.Printf("header: %s, salt %x, %d iterations\n", h.KDF, h.Salt, h.Iterations)
    	fmt.Println("check:", checkHeader(h))

    	key, err := deriveMasterKey("correct horse battery staple", h)
    	fmt.Printf("master key: %x %v\n", key, err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"encoding/hex"
    	"errors"
    	"testing"
    )

    // PBKDF2-HMAC-SHA256 test vectors from RFC 7914, section 11 (first 32 bytes).
    func TestDeriveMasterKeyRFC7914(t *testing.T) {
    	for _, v := range []struct {
    		password, salt string
    		iter           int
    		want           string
    	}{
    		{"passwd", "salt", 1, "55ac046e56e3089fec1691c22544b605f94185216dde0465e68b9d57c20dacbc"},
    		{"Password", "NaCl", 80000, "4ddcd8f60b98be21830cee5ef22701f9641a4418d04c0414aeff08876b34ab56"},
    	} {
    		key, err := deriveMasterKey(v.password, vaultHeader{KDF: "pbkdf2-sha256", Salt: []byte(v.salt), Iterations: v.iter})
    		if err != nil {
    			t.Fatalf("deriveMasterKey(%q, salt %q, %d iterations) error = %v", v.password, v.salt, v.iter, err)
    		}
    		if got := hex.EncodeToString(key); got != v.want {
    			t.Errorf("deriveMasterKey(%q, salt %q, %d iterations) = %s, want %s", v.password, v.salt, v.iter, got, v.want)
    		}
    	}
    }

    func TestNewVaultHeader(t *testing.T) {
    	a, b := newVaultHeader(), newVaultHeader()
    	if a.KDF != "pbkdf2-sha256" || a.Iterations != defaultIterations || len(a.Salt) != saltSize {
    		t.Fatalf("newVaultHeader() = {%q, %d-byte salt, %d}, want {\"pbkdf2-sha256\", %d-byte salt, %d}",
    			a.KDF, len(a.Salt), a.Iterations, saltSize, defaultIterations)
    	}
    	if bytes.Equal(a.Salt, b.Salt) {
    		t.Errorf("two calls to newVaultHeader() gave the same salt %x; salts must be random", a.Salt)
    	}
    	if err := checkHeader(a); err != nil {
    		t.Errorf("checkHeader(newVaultHeader()) = %v, want nil", err)
    	}
    }

    func TestCheckHeader(t *testing.T) {
    	salt := bytes.Repeat([]byte{7}, saltSize)
    	for _, tt := range []struct {
    		name string
    		h    vaultHeader
    		ok   bool
    	}{
    		{"good", vaultHeader{"pbkdf2-sha256", salt, 600_000}, true},
    		{"exactly the minimum", vaultHeader{"pbkdf2-sha256", salt, minIterations}, true},
    		{"too few iterations", vaultHeader{"pbkdf2-sha256", salt, 1000}, false},
    		{"zero iterations", vaultHeader{"pbkdf2-sha256", salt, 0}, false},
    		{"short salt", vaultHeader{"pbkdf2-sha256", salt[:8], 600_000}, false},
    		{"no salt", vaultHeader{"pbkdf2-sha256", nil, 600_000}, false},
    		{"unknown KDF", vaultHeader{"md5", salt, 600_000}, false},
    	} {
    		err := checkHeader(tt.h)
    		if tt.ok && err != nil {
    			t.Errorf("checkHeader(%s) = %v, want nil", tt.name, err)
    		}
    		if !tt.ok && !errors.Is(err, ErrWeakParams) {
    			t.Errorf("checkHeader(%s) = %v, want ErrWeakParams", tt.name, err)
    		}
    	}
    }
---

Time to give Keybox its master key. The standard library has had `crypto/pbkdf2` since
Go 1.24:

```go
func Key[Hash hash.Hash](h func() Hash, password string, salt []byte, iter, keyLength int) ([]byte, error)
```

## How PBKDF2 works

PBKDF2 (RFC 8018) runs HMAC over and over, using the password as the HMAC key:

```
U1 = HMAC(password, salt || 00000001)
U2 = HMAC(password, U1)
...
Uc = HMAC(password, Uc-1)
block = U1 ^ U2 ^ ... ^ Uc
```

`c` is the **iteration count**: the work factor. Each guess costs the attacker `c`
HMAC computations, exactly as it costs you. If you ask for more output than one hash
produces, PBKDF2 computes more blocks with counters 2, 3 and so on.

```go
package main

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"fmt"
)

func main() {
	key, err := pbkdf2.Key(sha256.New, "passwd", []byte("salt"), 1, 32)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%x\n", key)
}
```

```
55ac046e56e3089fec1691c22544b605f94185216dde0465e68b9d57c20dacbc
```

That matches RFC 7914's published vector for PBKDF2-HMAC-SHA256, so it doubles as a
known-answer test. `Key` returns an error only for invalid parameters, such as a key
length of zero, or (in FIPS 140-3 mode) inputs FIPS doesn't allow.

## Choosing parameters

- **Hash:** SHA-256. (PBKDF2-HMAC-SHA512 is also fine.)
- **Iterations:** OWASP currently recommends **600,000** for PBKDF2-HMAC-SHA256. On a
  typical laptop that's somewhere around 100 ms. Measure on your slowest supported
  device, and raise the count over the years as hardware gets faster.
- **Salt:** 16 random bytes from `crypto/rand`, fresh for every vault. Salts aren't
  secret.
- **Output length:** exactly what you need, here 32 bytes for one AES-256 key. Don't
  ask PBKDF2 for 64 bytes and split them into two keys: every extra 32-byte block costs
  *you* another full run of iterations, while an attacker only needs to compute the
  first block to test a guess. Derive one master key, then use HKDF (lesson 5) for more.

## Parameters travel with the data

The salt and iteration count must be available **before** you have the key, so they're
stored unencrypted in the vault header. That's fine, since they aren't secret, and it's
also what makes upgrades possible: new vaults get today's iteration count, old vaults
still open with theirs, and Keybox can re-derive with stronger settings the next time
Alice unlocks.

But a header read from disk or from the sync server is **attacker-controlled input**.
If Keybox re-saves a vault using whatever parameters its header claims, Mallory can
edit the header to 1 iteration and wait: the next save writes Alice's vault under a
key that's trivial to brute-force. So check parameters against minimums before
deriving a key you'll *encrypt* with, and cap the maximum too, so a malicious header
can't make Keybox spin for an hour.

## Your task

Complete Keybox's three KDF helpers:

1. `newVaultHeader()` returns `{KDF: "pbkdf2-sha256", Salt: <saltSize random bytes>,
   Iterations: defaultIterations}`.
2. `checkHeader(h)` returns `ErrWeakParams` unless `KDF` is `"pbkdf2-sha256"`, the salt
   is at least `saltSize` bytes, and `Iterations` is at least `minIterations`.
3. `deriveMasterKey(password, h)` returns `pbkdf2.Key` with SHA-256, the header's salt
   and iteration count, and `masterKeySize` bytes of output.

The tests use RFC 7914's vectors with low iteration counts, which is why
`deriveMasterKey` doesn't call `checkHeader` itself.
