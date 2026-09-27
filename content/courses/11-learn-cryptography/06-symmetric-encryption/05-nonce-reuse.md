---
title: Nonce Reuse
quiz:
  - question: Two Keybox messages were sealed with AES-GCM under the same key and the same nonce. Which consequence is *not* true?
    options:
      - text: Eve learns the XOR of the two plaintexts
      - text: The GCM authentication key can be recovered, letting an attacker forge messages that pass `Open`
      - text: The AES key itself is revealed directly
        correct: true
    explanation: |
      Reuse doesn't reveal the AES key, but it doesn't need to. The shared keystream
      leaks the XOR of the plaintexts, and the two tags give enough equations to solve
      for GHASH's authentication key, after which forgeries under that nonce are easy.
      Confidentiality and integrity are both gone.
  - question: '`cipher.NewGCMWithRandomNonce` reports a `NonceSize()` of 0. Why?'
    options:
      - text: It doesn't use a nonce at all
      - text: It generates a random 12-byte nonce inside `Seal`, prepends it to the output, and reads it back in `Open`, so the caller passes no nonce
        correct: true
      - text: It uses a counter starting at zero
      - text: It's a bug in Go 1.24
    explanation: |
      The nonce is still there; it's just managed for you. `Overhead()` is 28: 12 bytes
      of nonce plus the 16-byte tag. You call `Seal(dst, nil, plaintext, ad)`.
exercise:
  starter: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"encoding/binary"
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: decryption failed")

    // sealer encrypts Keybox sync messages. Output layout:
    // 12-byte nonce || ciphertext || 16-byte tag.
    type sealer struct {
    	gcm     cipher.AEAD
    	counter uint64
    }

    func newSealer(key []byte) (*sealer, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	gcm, err := cipher.NewGCM(block)
    	if err != nil {
    		return nil, err
    	}
    	return &sealer{gcm: gcm}, nil
    }

    // BUG: the counter lives in memory and starts at 0 in every process, so
    // every restart, and every other device sharing the key, reuses nonces.
    func (s *sealer) seal(plaintext []byte) []byte {
    	nonce := make([]byte, 12)
    	binary.BigEndian.PutUint64(nonce[4:], s.counter)
    	s.counter++
    	return s.gcm.Seal(nonce, nonce, plaintext, nil)
    }

    func (s *sealer) open(sealed []byte) ([]byte, error) {
    	if len(sealed) < 12+16 {
    		return nil, ErrDecrypt
    	}
    	pt, err := s.gcm.Open(nil, sealed[:12], sealed[12:], nil)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func main() {
    	key := []byte("an example 32-byte key for AES!!")
    	laptop, _ := newSealer(key)
    	phone, _ := newSealer(key) // another device, or the laptop after a restart
    	a := laptop.seal([]byte("set github-token"))
    	b := phone.seal([]byte("set aws-password"))
    	fmt.Printf("laptop nonce: %x\nphone nonce:  %x\n", a[:12], b[:12])
    	pt, err := laptop.open(b)
    	fmt.Printf("laptop opens phone's message: %q %v\n", pt, err)
    }
  solution: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: decryption failed")

    type sealer struct {
    	gcm cipher.AEAD
    }

    func newSealer(key []byte) (*sealer, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	gcm, err := cipher.NewGCMWithRandomNonce(block)
    	if err != nil {
    		return nil, err
    	}
    	return &sealer{gcm: gcm}, nil
    }

    func (s *sealer) seal(plaintext []byte) []byte {
    	return s.gcm.Seal(nil, nil, plaintext, nil)
    }

    func (s *sealer) open(sealed []byte) ([]byte, error) {
    	pt, err := s.gcm.Open(nil, nil, sealed, nil)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func main() {
    	key := []byte("an example 32-byte key for AES!!")
    	laptop, _ := newSealer(key)
    	phone, _ := newSealer(key)
    	a := laptop.seal([]byte("set github-token"))
    	b := phone.seal([]byte("set aws-password"))
    	fmt.Printf("laptop nonce: %x\nphone nonce:  %x\n", a[:12], b[:12])
    	pt, err := laptop.open(b)
    	fmt.Printf("laptop opens phone's message: %q %v\n", pt, err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"errors"
    	"testing"
    )

    var testKey = bytes.Repeat([]byte{0x33}, 32)

    func TestNoNonceReuseAcrossSealers(t *testing.T) {
    	seen := map[string]bool{}
    	for device := range 20 {
    		s, err := newSealer(testKey)
    		if err != nil {
    			t.Fatal(err)
    		}
    		for i := range 50 {
    			out := s.seal([]byte("sync message"))
    			n := string(out[:12])
    			if seen[n] {
    				t.Fatalf("device %d, message %d reused nonce %x; nonces must never repeat under one key", device, i, out[:12])
    			}
    			seen[n] = true
    		}
    	}
    }

    func TestLayoutAndRoundTrip(t *testing.T) {
    	a, _ := newSealer(testKey)
    	b, _ := newSealer(testKey)
    	msg := []byte("set github-token")
    	out := a.seal(msg)
    	if len(out) != 12+len(msg)+16 {
    		t.Fatalf("seal(%d bytes) returned %d bytes, want nonce(12) + ciphertext + tag(16) = %d", len(msg), len(out), 12+len(msg)+16)
    	}
    	block, _ := aes.NewCipher(testKey)
    	gcm, _ := cipher.NewGCM(block)
    	if pt, err := gcm.Open(nil, out[:12], out[12:], nil); err != nil || !bytes.Equal(pt, msg) {
    		t.Errorf("output isn't nonce || ciphertext || tag: standard GCM Open = %q, %v", pt, err)
    	}
    	if pt, err := b.open(out); err != nil || !bytes.Equal(pt, msg) {
    		t.Errorf("another sealer with the same key: open = %q, %v; want %q", pt, err, msg)
    	}
    	out[len(out)-1] ^= 1
    	if _, err := b.open(out); !errors.Is(err, ErrDecrypt) {
    		t.Errorf("open(tampered) error = %v, want ErrDecrypt", err)
    	}
    	if _, err := b.open([]byte("short")); !errors.Is(err, ErrDecrypt) {
    		t.Errorf("open(5 bytes) error = %v, want ErrDecrypt", err)
    	}
    }
---

The last lesson's rule was "a fresh random nonce every time". This lesson is about what
happens when that rule breaks, and how to make breaking it hard.

## What reuse costs

With AES-GCM, sealing two messages under the **same key and nonce** is catastrophic:

- **Confidentiality.** GCM encrypts with CTR mode, so the same nonce means the same
  keystream. XORing the two ciphertexts cancels the keystream and leaves the XOR of the
  two plaintexts: the two-time pad from chapter 1. With any guessable structure (JSON
  field names, a known secret format), both messages unravel.
- **Integrity.** GCM's tag is computed with an authentication key `H` derived from the
  AES key. Two messages with the same nonce give an attacker enough algebra to recover
  `H`, and with it, forge tags for modified messages under that nonce. This is sometimes
  called the "forbidden attack".

The AES key itself isn't revealed, but it doesn't matter: both of AEAD's promises are
gone. Researchers scanning the internet in 2016 found about 180 HTTPS servers that
repeated GCM nonces, some every time.

## How nonces end up repeating

Nobody writes `nonce := []byte("fixed nonce!")` on purpose (well, rarely). Reuse creeps
in through state:

- **Counters that reset.** A counter in memory starts at 0 after every restart.
- **Counters shared across machines.** Two servers or devices with the same key and
  their own counters produce the same sequence.
- **Snapshots and rollbacks.** Restore a VM or a database backup and the counter goes
  backwards.
- **Too many random nonces.** With 96-bit random nonces the chance of any collision
  becomes non-negligible after about 2^32 messages under one key. Rotate keys well
  before that.
- **Bugs.** Generating the nonce once at startup, or reusing a buffer that's filled
  once.

## Designing it out

- **Random nonces, managed by the library.** Go 1.24 added
  `cipher.NewGCMWithRandomNonce(block)`. Its `Seal` generates a fresh random 12-byte
  nonce every call and prepends it; `Open` reads it back. You pass `nil` as the nonce.
  There's no nonce for you to get wrong:

  ```go
  gcm, err := cipher.NewGCMWithRandomNonce(block)
  // ...
  sealed := gcm.Seal(nil, nil, plaintext, additionalData) // nonce || ciphertext || tag
  plaintext, err := gcm.Open(nil, nil, sealed, additionalData)
  ```

  The output layout is exactly the `nonce || ciphertext || tag` you built by hand last
  lesson, so the two are compatible. The 2^32-messages-per-key limit still applies.
- **Fresh keys per context.** Derive a key per file or per session with HKDF, so no
  single key sees enough messages to worry.
- **Bigger nonces.** XChaCha20-Poly1305 (in `x/crypto`) has 192-bit nonces that are
  safe to choose at random practically without limit.
- **Nonce-misuse-resistant modes** such as AES-GCM-SIV degrade gracefully: reusing a
  nonce only reveals whether two messages were identical. They're not in Go's standard
  library.

Counters remain a fine choice for *protocols* that already keep strict per-connection
state, which is exactly what TLS does: every record gets the next sequence number, and
each connection has fresh keys.

## Your task

Keybox's sync client builds nonces from an in-memory counter, so the laptop and the
phone (and the laptop after every restart) reuse the same nonces with the same key.
Run the starter to see both devices print the identical nonce.

Fix `sealer` so nonces can't repeat: build it with `cipher.NewGCMWithRandomNonce`,
remove the counter, and make `seal` and `open` pass `nil` nonces. Keep the output
layout `nonce || ciphertext || tag`, and keep returning `ErrDecrypt` for any failure
in `open`.
