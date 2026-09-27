---
title: Encrypt-then-MAC, Properly
difficulty: easy
after: symmetric-encryption
hints:
  - 'In CTR mode the IV decides the keystream. If the tag doesn''t cover the IV, an attacker can change the IV freely and the tag still checks out, so you''d decrypt garbage and call it genuine. The tag must cover **iv || ciphertext**, everything except itself.'
  - 'In `open`: check the length first (at least 16 + 32 bytes), then compute the HMAC over `sealed[:len(sealed)-32]` and compare it with the last 32 bytes using `hmac.Equal`. Only after that succeeds should you decrypt, and on failure return `nil, ErrDecrypt`, never the plaintext.'
exercise:
  starter: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/hmac"
    	"crypto/rand"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: decryption failed")

    const (
    	ivSize  = aes.BlockSize // 16
    	tagSize = sha256.Size   // 32
    )

    // seal encrypts plaintext with AES-256-CTR under key[:32] and authenticates
    // it with HMAC-SHA256 under key[32:]. key is 64 bytes. Output layout:
    //
    //	iv (16) || ciphertext || tag (32)
    //
    // where tag = HMAC-SHA256(key[32:], iv || ciphertext).
    //
    // BUG: the tag only covers the ciphertext, not the IV.
    func seal(key, plaintext []byte) []byte {
    	block, err := aes.NewCipher(key[:32])
    	if err != nil {
    		panic(err)
    	}
    	iv := make([]byte, ivSize)
    	rand.Read(iv)
    	ct := make([]byte, len(plaintext))
    	cipher.NewCTR(block, iv).XORKeyStream(ct, plaintext)

    	mac := hmac.New(sha256.New, key[32:])
    	mac.Write(ct) // ? what else must the tag cover?
    	out := append(iv, ct...)
    	return mac.Sum(out)
    }

    // open reverses seal. Any problem (short input, wrong key, any modified
    // byte) returns nil and ErrDecrypt.
    //
    // BUG: it checks the wrong bytes, decrypts before verifying, and hands the
    // plaintext back even when the check fails. It also panics on short input.
    func open(key, sealed []byte) ([]byte, error) {
    	block, err := aes.NewCipher(key[:32])
    	if err != nil {
    		return nil, err
    	}
    	iv := sealed[:ivSize]
    	ct := sealed[ivSize : len(sealed)-tagSize]
    	tag := sealed[len(sealed)-tagSize:]

    	pt := make([]byte, len(ct))
    	cipher.NewCTR(block, iv).XORKeyStream(pt, ct)

    	mac := hmac.New(sha256.New, key[32:])
    	mac.Write(ct)
    	if !bytes.Equal(mac.Sum(nil), tag) {
    		return pt, ErrDecrypt
    	}
    	return pt, nil
    }

    func main() {
    	key := bytes.Repeat([]byte{7}, 64)
    	sealed := seal(key, []byte("pay alice $10"))
    	sealed[0] ^= 0x01 // an attacker flips one bit of the IV
    	pt, err := open(key, sealed)
    	fmt.Printf("%q %v\n", pt, err) // want: "" keybox: decryption failed
    }
  solution: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/hmac"
    	"crypto/rand"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: decryption failed")

    const (
    	ivSize  = aes.BlockSize // 16
    	tagSize = sha256.Size   // 32
    )

    // seal encrypts plaintext with AES-256-CTR under key[:32] and authenticates
    // it with HMAC-SHA256 under key[32:]. Output: iv || ciphertext || tag, where
    // tag = HMAC-SHA256(key[32:], iv || ciphertext).
    func seal(key, plaintext []byte) []byte {
    	block, err := aes.NewCipher(key[:32])
    	if err != nil {
    		panic(err)
    	}
    	out := make([]byte, ivSize+len(plaintext))
    	iv, ct := out[:ivSize], out[ivSize:]
    	rand.Read(iv)
    	cipher.NewCTR(block, iv).XORKeyStream(ct, plaintext)

    	mac := hmac.New(sha256.New, key[32:])
    	mac.Write(out) // iv || ciphertext
    	return mac.Sum(out)
    }

    // open reverses seal. Any problem returns nil and ErrDecrypt.
    func open(key, sealed []byte) ([]byte, error) {
    	if len(key) != 64 || len(sealed) < ivSize+tagSize {
    		return nil, ErrDecrypt
    	}
    	body, tag := sealed[:len(sealed)-tagSize], sealed[len(sealed)-tagSize:]
    	mac := hmac.New(sha256.New, key[32:])
    	mac.Write(body)
    	if !hmac.Equal(mac.Sum(nil), tag) {
    		return nil, ErrDecrypt
    	}
    	block, err := aes.NewCipher(key[:32])
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	pt := make([]byte, len(body)-ivSize)
    	cipher.NewCTR(block, body[:ivSize]).XORKeyStream(pt, body[ivSize:])
    	return pt, nil
    }

    func main() {
    	key := bytes.Repeat([]byte{7}, 64)
    	sealed := seal(key, []byte("pay alice $10"))
    	sealed[0] ^= 0x01
    	pt, err := open(key, sealed)
    	fmt.Printf("%q %v\n", pt, err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"errors"
    	"testing"
    )

    var (
    	testKey  = bytes.Repeat([]byte{0x42}, 64)
    	otherKey = append(bytes.Repeat([]byte{0x42}, 32), bytes.Repeat([]byte{0x43}, 32)...)
    )

    // safeOpen turns a panic into a test failure with a clear message.
    func safeOpen(t *testing.T, key, sealed []byte) (pt []byte, err error) {
    	t.Helper()
    	defer func() {
    		if r := recover(); r != nil {
    			t.Errorf("open panicked on %d-byte input: %v", len(sealed), r)
    			pt, err = nil, ErrDecrypt
    		}
    	}()
    	return open(key, sealed)
    }

    func TestRoundTrip(t *testing.T) {
    	for _, msg := range []string{"", "x", "pay alice $10", string(bytes.Repeat([]byte("vault"), 100))} {
    		sealed := seal(testKey, []byte(msg))
    		if len(sealed) != ivSize+len(msg)+tagSize {
    			t.Errorf("seal(%d bytes) returned %d bytes, want iv(16) + %d + tag(32) = %d", len(msg), len(sealed), len(msg), ivSize+len(msg)+tagSize)
    		}
    		pt, err := safeOpen(t, testKey, sealed)
    		if err != nil || string(pt) != msg {
    			t.Errorf("open(seal(%.20q)) = %.20q, %v, want the original message", msg, pt, err)
    		}
    	}
    }

    func TestEveryBitFlipRejected(t *testing.T) {
    	sealed := seal(testKey, []byte("pay alice $10"))
    	for i := range sealed {
    		for bit := range 8 {
    			bad := bytes.Clone(sealed)
    			bad[i] ^= 1 << bit
    			pt, err := safeOpen(t, testKey, bad)
    			if !errors.Is(err, ErrDecrypt) || pt != nil {
    				where := "ciphertext"
    				switch {
    				case i < ivSize:
    					where = "IV"
    				case i >= len(sealed)-tagSize:
    					where = "tag"
    				}
    				t.Fatalf("flipping bit %d of byte %d (in the %s) gave open = %q, %v, want nil, ErrDecrypt", bit, i, where, pt, err)
    			}
    		}
    	}
    }

    func TestWrongKeyRejected(t *testing.T) {
    	sealed := seal(testKey, []byte("pay alice $10"))
    	if pt, err := safeOpen(t, otherKey, sealed); !errors.Is(err, ErrDecrypt) || pt != nil {
    		t.Errorf("open with a different MAC key = %q, %v, want nil, ErrDecrypt", pt, err)
    	}
    }

    func TestShortInputRejected(t *testing.T) {
    	sealed := seal(testKey, []byte("hi"))
    	for n := range ivSize + tagSize {
    		if pt, err := safeOpen(t, testKey, sealed[:n]); !errors.Is(err, ErrDecrypt) || pt != nil {
    			t.Errorf("open of a %d-byte input = %q, %v, want nil, ErrDecrypt", n, pt, err)
    		}
    	}
    	if pt, err := safeOpen(t, testKey, sealed[:len(sealed)-1]); !errors.Is(err, ErrDecrypt) || pt != nil {
    		t.Errorf("open with the last byte cut off = %q, %v, want nil, ErrDecrypt", pt, err)
    	}
    }

    func TestFreshIVs(t *testing.T) {
    	seen := map[string]bool{}
    	for range 10_000 {
    		iv := string(seal(testKey, []byte("same message"))[:ivSize])
    		if seen[iv] {
    			t.Fatalf("seal reused an IV within 10,000 messages: CTR with a repeated IV leaks the XOR of the plaintexts")
    		}
    		seen[iv] = true
    	}
    }
---

Before Keybox switched to AES-GCM, it encrypted exported entries with a hand-built
**encrypt-then-MAC** scheme: AES-256-CTR for secrecy plus HMAC-SHA256 for integrity.
That's a sound design *if* every detail is right. The old importer still has to read
these files, and its code has several details wrong.

The sealed format is

```
iv (16 bytes) || ciphertext || tag (32 bytes)
tag = HMAC-SHA256(key[32:], iv || ciphertext)
```

with a 64-byte `key`: `key[:32]` for AES and `key[32:]` for HMAC (two keys, one per
purpose). Fix `seal` and `open` so that:

- the tag covers the IV **and** the ciphertext,
- `open` verifies the tag in **constant time** *before* decrypting anything,
- any failure (too short, wrong key, any modified bit) returns `nil, ErrDecrypt`,
  never a panic and never partial plaintext.

## Example

```go
sealed := seal(key, []byte("pay alice $10"))
open(key, sealed)     // "pay alice $10", nil
sealed[0] ^= 1        // flip one bit of the IV
open(key, sealed)     // nil, ErrDecrypt  (the buggy version returns garbage and no error!)
```

## Constraints

- The tests flip every single bit of a sealed message, cut inputs to every length
  below 48 bytes, use a wrong MAC key, and seal 10,000 messages checking that no IV
  repeats.
- New code shouldn't build this by hand: an AEAD like AES-GCM does all of it in one
  call. You'll meet this pattern when reading older formats.
