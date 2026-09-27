---
title: SHA-256 in Go
quiz:
  - question: |
      What does this print?

      ```go
      sum := sha256.Sum256([]byte("keybox"))
      fmt.Println(len(sum), len(hex.EncodeToString(sum[:])))
      ```
    options:
      - text: '`32 32`'
      - text: '`32 64`'
        correct: true
      - text: '`64 64`'
      - text: '`256 64`'
    explanation: |
      `Sum256` returns a `[32]byte` array (256 bits). Hex uses two characters per byte,
      so the string is 64 characters long. Note the `sum[:]`: `EncodeToString` wants a
      slice, and `sum` is an array.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var (
    	ErrBadChecksum = errors.New("checksum must be 64 hex characters")
    	ErrMismatch    = errors.New("checksum mismatch")
    )

    // fingerprint returns the SHA-256 digest of data as 64 lowercase hex characters.
    func fingerprint(data []byte) string {
    	// ?
    	return ""
    }

    // verifyChecksum checks data against a SHA-256 checksum given as hex
    // (upper or lower case). It returns ErrBadChecksum if want isn't valid
    // hex for exactly 32 bytes, ErrMismatch if the digest differs, and nil
    // if it matches.
    func verifyChecksum(data []byte, want string) error {
    	// ?
    	return nil
    }

    func main() {
    	release := []byte("keybox v1.4.0 linux/amd64 binary contents")
    	sum := fingerprint(release)
    	fmt.Println("sha256:", sum)
    	fmt.Println("verify:", verifyChecksum(release, sum))
    	fmt.Println("tampered:", verifyChecksum(append(release, '!'), sum))
    	fmt.Println("garbage:", verifyChecksum(release, "not-hex"))
    }
  solution: |
    package main

    import (
    	"bytes"
    	"crypto/sha256"
    	"encoding/hex"
    	"errors"
    	"fmt"
    )

    var (
    	ErrBadChecksum = errors.New("checksum must be 64 hex characters")
    	ErrMismatch    = errors.New("checksum mismatch")
    )

    func fingerprint(data []byte) string {
    	sum := sha256.Sum256(data)
    	return hex.EncodeToString(sum[:])
    }

    func verifyChecksum(data []byte, want string) error {
    	wantSum, err := hex.DecodeString(want)
    	if err != nil || len(wantSum) != sha256.Size {
    		return ErrBadChecksum
    	}
    	got := sha256.Sum256(data)
    	if !bytes.Equal(got[:], wantSum) {
    		return ErrMismatch
    	}
    	return nil
    }

    func main() {
    	release := []byte("keybox v1.4.0 linux/amd64 binary contents")
    	sum := fingerprint(release)
    	fmt.Println("sha256:", sum)
    	fmt.Println("verify:", verifyChecksum(release, sum))
    	fmt.Println("tampered:", verifyChecksum(append(release, '!'), sum))
    	fmt.Println("garbage:", verifyChecksum(release, "not-hex"))
    }
  tests: |
    package main

    import (
    	"errors"
    	"strings"
    	"testing"
    )

    // Known answers: FIPS 180-2's "abc" example, the empty string, and "keybox".
    var vectors = []struct{ in, want string }{
    	{"abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
    	{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
    	{"keybox", "bf3d3b21c23b934be88eb74d9a756545963ec7e8f803fac4ec949ee5ec21140f"},
    }

    func TestFingerprint(t *testing.T) {
    	for _, v := range vectors {
    		if got := fingerprint([]byte(v.in)); got != v.want {
    			t.Errorf("fingerprint(%q) = %q, want %q", v.in, got, v.want)
    		}
    	}
    }

    func TestVerifyChecksum(t *testing.T) {
    	for _, v := range vectors {
    		if err := verifyChecksum([]byte(v.in), v.want); err != nil {
    			t.Errorf("verifyChecksum(%q, correct sum) = %v, want nil", v.in, err)
    		}
    		if err := verifyChecksum([]byte(v.in), strings.ToUpper(v.want)); err != nil {
    			t.Errorf("verifyChecksum(%q, correct sum in UPPER case) = %v, want nil", v.in, err)
    		}
    		if err := verifyChecksum([]byte(v.in+"x"), v.want); !errors.Is(err, ErrMismatch) {
    			t.Errorf("verifyChecksum(%q, sum of %q) = %v, want ErrMismatch", v.in+"x", v.in, err)
    		}
    	}
    	sum := vectors[0].want
    	for _, bad := range []string{"", "not-hex", sum[:62], sum + "00", sum[:63] + "g", sum[:32]} {
    		if err := verifyChecksum([]byte("abc"), bad); !errors.Is(err, ErrBadChecksum) {
    			t.Errorf("verifyChecksum(\"abc\", %q) = %v, want ErrBadChecksum", bad, err)
    		}
    	}
    }
---

Keybox publishes a SHA-256 checksum next to every release download, and its CLI prints
fingerprints of keys and bundles. Both start with `crypto/sha256`.

## One-shot hashing

For data already in memory, `sha256.Sum256` is all you need:

```go
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func main() {
	sum := sha256.Sum256([]byte("abc"))
	fmt.Printf("%T\n", sum)
	fmt.Printf("%x\n", sum)
	fmt.Println(hex.EncodeToString(sum[:]))
}
```

```
[32]uint8
ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
```

That digest for `"abc"` is the worked example from the SHA-2 standard, FIPS 180-2. If
your code prints anything else for `"abc"`, something is broken, which makes it a
handy known-answer test.

A few details worth noticing:

- `Sum256` returns an **array**, `[32]byte`, not a slice. Arrays are comparable, so
  `sha256.Sum256(a) == sha256.Sum256(b)` compiles and works. To pass it where a slice is
  expected, write `sum[:]`.
- `%x` formats bytes as lowercase hex directly. `hex.EncodeToString` does the same and
  returns a string.
- `sha256.Size` is 32, handy for validating input lengths.

## Verifying a checksum

Checking a download against a published checksum means decoding the expected hex and
comparing digests. `hex.DecodeString` accepts both upper and lower case and returns an
error for anything that isn't hex, so you get validation for free.

Should the comparison be constant-time? Here it doesn't matter: a public checksum isn't
a secret, and timing can't help an attacker produce a matching file. When you compare
*secret* values such as MACs and tokens it matters a great deal, as chapter 4 shows.
Get into the habit of asking "is either side of this comparison secret?" every time.

## Your task

Complete Keybox's two helpers:

1. `fingerprint(data)` returns the SHA-256 of `data` as 64 lowercase hex characters.
2. `verifyChecksum(data, want)`:
   - decode `want` with `hex.DecodeString`; if that fails, or it isn't exactly
     `sha256.Size` bytes, return `ErrBadChecksum`,
   - compute the digest of `data` and return `ErrMismatch` if it differs (`bytes.Equal`
     is fine, since neither side is secret),
   - otherwise return `nil`.

You'll need to import `crypto/sha256`, `encoding/hex` and `bytes`.
