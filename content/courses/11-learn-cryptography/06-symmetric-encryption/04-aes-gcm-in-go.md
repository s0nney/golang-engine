---
title: AES-GCM in Go
quiz:
  - question: 'Keybox seals a 100-byte secret with AES-GCM, a 12-byte random nonce stored in front, and the standard 16-byte tag. How long is the result?'
    options:
      - text: 100 bytes
      - text: 112 bytes
      - text: 128 bytes
        correct: true
      - text: 116 bytes
    explanation: |
      12 (nonce) + 100 (ciphertext, same length as the plaintext) + 16 (tag) = 128.
      GCM is a stream mode, so there's no padding; `gcm.Overhead()` reports the 16-byte
      tag.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: decryption failed")

    // seal encrypts plaintext with AES-256-GCM under key (32 bytes) and returns
    // nonce || ciphertext-and-tag, using a fresh random 12-byte nonce.
    func seal(key, plaintext []byte) ([]byte, error) {
    	// ?
    	return nil, errors.New("not implemented")
    }

    // open reverses seal. Any failure (short input, wrong key, tampering)
    // returns ErrDecrypt and no plaintext.
    func open(key, sealed []byte) ([]byte, error) {
    	// ?
    	return nil, ErrDecrypt
    }

    func main() {
    	key := []byte("an example 32-byte key for AES!!")
    	sealed, err := seal(key, []byte("ghp_hunter2hunter2"))
    	fmt.Printf("sealed: %x (%d bytes) %v\n", sealed, len(sealed), err)

    	plain, err := open(key, sealed)
    	fmt.Printf("opened: %q %v\n", plain, err)

    	if len(sealed) > 0 {
    		sealed[len(sealed)-1] ^= 1
    	}
    	_, err = open(key, sealed)
    	fmt.Println("tampered:", err)
    }
  solution: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/rand"
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: decryption failed")

    func newGCM(key []byte) (cipher.AEAD, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	return cipher.NewGCM(block)
    }

    func seal(key, plaintext []byte) ([]byte, error) {
    	gcm, err := newGCM(key)
    	if err != nil {
    		return nil, err
    	}
    	nonce := make([]byte, gcm.NonceSize())
    	rand.Read(nonce)
    	return gcm.Seal(nonce, nonce, plaintext, nil), nil
    }

    func open(key, sealed []byte) ([]byte, error) {
    	gcm, err := newGCM(key)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	if len(sealed) < gcm.NonceSize()+gcm.Overhead() {
    		return nil, ErrDecrypt
    	}
    	nonce, ciphertext := sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():]
    	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return plaintext, nil
    }

    func main() {
    	key := []byte("an example 32-byte key for AES!!")
    	sealed, err := seal(key, []byte("ghp_hunter2hunter2"))
    	fmt.Printf("sealed: %x (%d bytes) %v\n", sealed, len(sealed), err)

    	plain, err := open(key, sealed)
    	fmt.Printf("opened: %q %v\n", plain, err)

    	if len(sealed) > 0 {
    		sealed[len(sealed)-1] ^= 1
    	}
    	_, err = open(key, sealed)
    	fmt.Println("tampered:", err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"errors"
    	"testing"
    )

    var (
    	key1 = bytes.Repeat([]byte{1}, 32)
    	key2 = bytes.Repeat([]byte{2}, 32)
    )

    func mustSeal(t *testing.T, key, pt []byte) []byte {
    	t.Helper()
    	s, err := seal(key, pt)
    	if err != nil {
    		t.Fatalf("seal(%d-byte key, %q) error = %v", len(key), pt, err)
    	}
    	return s
    }

    func TestRoundTrip(t *testing.T) {
    	for _, pt := range []string{"", "x", "ghp_hunter2hunter2", string(bytes.Repeat([]byte("keybox"), 1000))} {
    		s := mustSeal(t, key1, []byte(pt))
    		if want := 12 + len(pt) + 16; len(s) != want {
    			t.Errorf("seal of %d bytes gave %d bytes, want %d (nonce + ciphertext + tag)", len(pt), len(s), want)
    		}
    		got, err := open(key1, s)
    		if err != nil || string(got) != pt {
    			t.Errorf("open(seal(%.20q...)) = %.20q, %v; want the plaintext back", pt, got, err)
    		}
    	}
    }

    func TestFreshNonces(t *testing.T) {
    	a := mustSeal(t, key1, []byte("same plaintext"))
    	b := mustSeal(t, key1, []byte("same plaintext"))
    	if bytes.Equal(a[:12], b[:12]) {
    		t.Errorf("two seals used the same nonce %x; generate a fresh random nonce every time", a[:12])
    	}
    	if bytes.Equal(a, b) {
    		t.Errorf("sealing the same plaintext twice gave identical output")
    	}
    }

    func TestTamperDetection(t *testing.T) {
    	s := mustSeal(t, key1, []byte("ghp_hunter2"))
    	for i := range len(s) * 8 {
    		bad := bytes.Clone(s)
    		bad[i/8] ^= 1 << (i % 8)
    		if got, err := open(key1, bad); !errors.Is(err, ErrDecrypt) || got != nil {
    			t.Fatalf("flipping bit %d: open = %q, %v; want nil, ErrDecrypt", i, got, err)
    		}
    	}
    }

    func TestWrongKeyAndGarbage(t *testing.T) {
    	s := mustSeal(t, key1, []byte("ghp_hunter2"))
    	if got, err := open(key2, s); !errors.Is(err, ErrDecrypt) || got != nil {
    		t.Errorf("open with the wrong key = %q, %v; want nil, ErrDecrypt", got, err)
    	}
    	for _, bad := range [][]byte{nil, {1, 2, 3}, s[:12], s[:27]} {
    		if _, err := open(key1, bad); !errors.Is(err, ErrDecrypt) {
    			t.Errorf("open(%d bytes) error = %v, want ErrDecrypt (and no panic)", len(bad), err)
    		}
    	}
    	if _, err := seal([]byte("short key"), []byte("x")); err == nil {
    		t.Errorf("seal with a 9-byte key: error = nil, want an error")
    	}
    }
---

Keybox is ready for real encryption. `crypto/cipher.NewGCM` wraps an AES block cipher
in Galois/Counter Mode and hands you a `cipher.AEAD`.

## Seal

```go
block, err := aes.NewCipher(key) // 32-byte key: AES-256
if err != nil {
	return nil, err
}
gcm, err := cipher.NewGCM(block)
if err != nil {
	return nil, err
}
nonce := make([]byte, gcm.NonceSize()) // 12 bytes
rand.Read(nonce)
sealed := gcm.Seal(nonce, nonce, plaintext, nil)
```

That last line packs a lot in. `Seal(dst, nonce, plaintext, additionalData)` **appends**
the ciphertext and 16-byte tag to `dst`. Passing `nonce` as `dst` produces
`nonce || ciphertext || tag` in one slice, which is the conventional layout: the
receiver needs the nonce to decrypt, and it isn't secret, so it travels in front.

## Open

Opening reverses it: split off the nonce and call `Open`:

```go
if len(sealed) < gcm.NonceSize()+gcm.Overhead() {
	return nil, ErrDecrypt
}
nonce, ciphertext := sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():]
plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
if err != nil {
	return nil, ErrDecrypt
}
```

- **Check the length first.** Slicing `sealed[:12]` on a 5-byte input panics, and a
  panic triggered by attacker input is a denial-of-service bug.
- **`Open` verifies the tag before returning anything.** If a single bit of the nonce,
  ciphertext, tag or associated data changed, or the key is wrong, you get an error and
  no plaintext.
- **Return one error for every failure.** Wrong key, tampering, truncation: to the
  caller they're all "decryption failed". Distinguishing them helps attackers, not
  users.

## Why the nonce is random here

Each seal needs a nonce that has never been used with this key. Keybox uses random
12-byte nonces from `crypto/rand`, which is safe as long as one key encrypts at most
about 2^32 (four billion) messages. Keybox's vault key encrypts a few hundred secrets
in its lifetime, so random nonces are plenty. The next lesson looks at what goes wrong
if a nonce repeats, and at a Go 1.24 helper that manages random nonces for you.

## Your task

Implement Keybox's `seal` and `open` for AES-256-GCM:

1. `seal(key, plaintext)`: create the cipher and GCM (return any error, such as a bad
   key length), generate a fresh random nonce of `gcm.NonceSize()` bytes, and return
   `nonce || ciphertext || tag`.
2. `open(key, sealed)`: return `ErrDecrypt` (and `nil` plaintext) if the key is
   invalid, if `sealed` is shorter than `NonceSize() + Overhead()`, or if `Open` fails.
   Otherwise return the plaintext.

A small `newGCM(key)` helper keeps the two functions short. The tests flip every single
bit of a sealed message and expect every one to be rejected.
