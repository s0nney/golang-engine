---
title: Crypto Agility and Key Rotation
quiz:
  - question: Keybox rotates its data key from ID 1 to ID 2. What should happen to ciphertexts sealed under key 1?
    options:
      - text: They become unreadable immediately
      - text: They stay readable (the header says key 1, which stays in the keyring), get re-encrypted under key 2 in the background, and key 1 is retired once none remain
        correct: true
      - text: They're decrypted and stored in plaintext during the migration
      - text: Key 2 can decrypt them because it was derived from key 1
    explanation: |
      Rotation is a process: new writes use the new key immediately, old data is
      re-encrypted gradually, and the old key is destroyed only once nothing needs it.
      The key ID in each ciphertext's header is what makes this possible.
exercise:
  starter: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"encoding/binary"
    	"errors"
    	"fmt"
    )

    var (
    	ErrBadFormat  = errors.New("keybox: unknown ciphertext format")
    	ErrUnknownKey = errors.New("keybox: ciphertext uses a retired or unknown key")
    	ErrDecrypt    = errors.New("keybox: decryption failed")
    	ErrNoKey      = errors.New("keybox: no current key")
    )

    const (
    	formatV1   = 1
    	headerSize = 5 // 1-byte format version + 4-byte big-endian key ID
    )

    // Keyring holds every key that may still be needed to decrypt, and which
    // one new data is sealed with.
    type Keyring struct {
    	current uint32
    	keys    map[uint32]cipher.AEAD
    }

    func NewKeyring() *Keyring { return &Keyring{keys: map[uint32]cipher.AEAD{}} }

    // Add registers a 32-byte key under id, optionally making it current.
    func (k *Keyring) Add(id uint32, key []byte, makeCurrent bool) error {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return err
    	}
    	aead, err := cipher.NewGCMWithRandomNonce(block)
    	if err != nil {
    		return err
    	}
    	k.keys[id] = aead
    	if makeCurrent {
    		k.current = id
    	}
    	return nil
    }

    // Retire forgets a key. Anything still sealed under it becomes unreadable.
    func (k *Keyring) Retire(id uint32) { delete(k.keys, id) }

    // header returns formatV1 || id (big-endian).
    func header(id uint32) []byte {
    	return binary.BigEndian.AppendUint32([]byte{formatV1}, id)
    }

    // Seal encrypts with the current key and returns header || GCM output.
    // The associated data passed to GCM is header followed by ad, so the header
    // can't be altered. Returns ErrNoKey if the current key isn't in the keyring.
    func (k *Keyring) Seal(plaintext, ad []byte) ([]byte, error) {
    	// ?
    	return nil, ErrNoKey
    }

    // Open parses the header (ErrBadFormat if too short or not formatV1),
    // looks up the key ID (ErrUnknownKey), and opens with header+ad as
    // associated data (ErrDecrypt).
    func (k *Keyring) Open(sealed, ad []byte) ([]byte, error) {
    	// ?
    	return nil, ErrBadFormat
    }

    // Reencrypt opens sealed with whatever key it names and seals the result
    // with the current key.
    func (k *Keyring) Reencrypt(sealed, ad []byte) ([]byte, error) {
    	// ?
    	return sealed, nil
    }

    func main() {
    	kr := NewKeyring()
    	kr.Add(1, bytes.Repeat([]byte{1}, 32), true)
    	old, err := kr.Seal([]byte("ghp_alice"), []byte("github-token"))
    	fmt.Printf("sealed under key %x: %v\n", old[:min(len(old), headerSize)], err)

    	kr.Add(2, bytes.Repeat([]byte{2}, 32), true) // rotate
    	fresh, err := kr.Reencrypt(old, []byte("github-token"))
    	fmt.Printf("re-encrypted under %x: %v\n", fresh[:min(len(fresh), headerSize)], err)

    	kr.Retire(1)
    	pt, err := kr.Open(fresh, []byte("github-token"))
    	fmt.Printf("after retiring key 1: %q %v\n", pt, err)
    }
  solution: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"encoding/binary"
    	"errors"
    	"fmt"
    )

    var (
    	ErrBadFormat  = errors.New("keybox: unknown ciphertext format")
    	ErrUnknownKey = errors.New("keybox: ciphertext uses a retired or unknown key")
    	ErrDecrypt    = errors.New("keybox: decryption failed")
    	ErrNoKey      = errors.New("keybox: no current key")
    )

    const (
    	formatV1   = 1
    	headerSize = 5
    )

    type Keyring struct {
    	current uint32
    	keys    map[uint32]cipher.AEAD
    }

    func NewKeyring() *Keyring { return &Keyring{keys: map[uint32]cipher.AEAD{}} }

    func (k *Keyring) Add(id uint32, key []byte, makeCurrent bool) error {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return err
    	}
    	aead, err := cipher.NewGCMWithRandomNonce(block)
    	if err != nil {
    		return err
    	}
    	k.keys[id] = aead
    	if makeCurrent {
    		k.current = id
    	}
    	return nil
    }

    func (k *Keyring) Retire(id uint32) { delete(k.keys, id) }

    func header(id uint32) []byte {
    	return binary.BigEndian.AppendUint32([]byte{formatV1}, id)
    }

    func (k *Keyring) Seal(plaintext, ad []byte) ([]byte, error) {
    	aead, ok := k.keys[k.current]
    	if !ok {
    		return nil, ErrNoKey
    	}
    	h := header(k.current)
    	fullAD := append(bytes.Clone(h), ad...)
    	return aead.Seal(h, nil, plaintext, fullAD), nil
    }

    func (k *Keyring) Open(sealed, ad []byte) ([]byte, error) {
    	if len(sealed) < headerSize || sealed[0] != formatV1 {
    		return nil, ErrBadFormat
    	}
    	id := binary.BigEndian.Uint32(sealed[1:headerSize])
    	aead, ok := k.keys[id]
    	if !ok {
    		return nil, ErrUnknownKey
    	}
    	fullAD := append(bytes.Clone(sealed[:headerSize]), ad...)
    	pt, err := aead.Open(nil, nil, sealed[headerSize:], fullAD)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func (k *Keyring) Reencrypt(sealed, ad []byte) ([]byte, error) {
    	pt, err := k.Open(sealed, ad)
    	if err != nil {
    		return nil, err
    	}
    	return k.Seal(pt, ad)
    }

    func main() {
    	kr := NewKeyring()
    	kr.Add(1, bytes.Repeat([]byte{1}, 32), true)
    	old, err := kr.Seal([]byte("ghp_alice"), []byte("github-token"))
    	fmt.Printf("sealed under key %x: %v\n", old[:min(len(old), headerSize)], err)

    	kr.Add(2, bytes.Repeat([]byte{2}, 32), true)
    	fresh, err := kr.Reencrypt(old, []byte("github-token"))
    	fmt.Printf("re-encrypted under %x: %v\n", fresh[:min(len(fresh), headerSize)], err)

    	kr.Retire(1)
    	pt, err := kr.Open(fresh, []byte("github-token"))
    	fmt.Printf("after retiring key 1: %q %v\n", pt, err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"errors"
    	"testing"
    )

    func ring(t *testing.T) *Keyring {
    	t.Helper()
    	kr := NewKeyring()
    	if err := kr.Add(1, bytes.Repeat([]byte{1}, 32), true); err != nil {
    		t.Fatal(err)
    	}
    	return kr
    }

    var ad = []byte("github-token")

    func TestSealFormat(t *testing.T) {
    	kr := ring(t)
    	s, err := kr.Seal([]byte("ghp_alice"), ad)
    	if err != nil {
    		t.Fatalf("Seal error = %v", err)
    	}
    	if !bytes.Equal(s[:headerSize], []byte{1, 0, 0, 0, 1}) {
    		t.Errorf("header = %x, want 0100000001 (format 1, key ID 1)", s[:min(len(s), headerSize)])
    	}
    	if want := headerSize + 12 + len("ghp_alice") + 16; len(s) != want {
    		t.Errorf("sealed length = %d, want %d", len(s), want)
    	}
    	if pt, err := kr.Open(s, ad); err != nil || string(pt) != "ghp_alice" {
    		t.Errorf("Open(Seal(x)) = %q, %v", pt, err)
    	}
    	if _, err := NewKeyring().Seal([]byte("x"), nil); !errors.Is(err, ErrNoKey) {
    		t.Errorf("Seal with an empty keyring: error = %v, want ErrNoKey", err)
    	}
    }

    func TestRotation(t *testing.T) {
    	kr := ring(t)
    	old, _ := kr.Seal([]byte("ghp_alice"), ad)
    	kr.Add(2, bytes.Repeat([]byte{2}, 32), true)

    	if pt, err := kr.Open(old, ad); err != nil || string(pt) != "ghp_alice" {
    		t.Errorf("after rotating, old ciphertext: Open = %q, %v; want it still readable", pt, err)
    	}
    	s, _ := kr.Seal([]byte("new"), ad)
    	if !bytes.Equal(s[:headerSize], []byte{1, 0, 0, 0, 2}) {
    		t.Errorf("after rotating, new ciphertexts should use key 2; header = %x", s[:headerSize])
    	}
    	fresh, err := kr.Reencrypt(old, ad)
    	if err != nil || !bytes.Equal(fresh[:headerSize], []byte{1, 0, 0, 0, 2}) {
    		t.Fatalf("Reencrypt = header %x, %v; want key 2", fresh[:min(len(fresh), headerSize)], err)
    	}
    	kr.Retire(1)
    	if pt, err := kr.Open(fresh, ad); err != nil || string(pt) != "ghp_alice" {
    		t.Errorf("re-encrypted ciphertext after retiring key 1: Open = %q, %v", pt, err)
    	}
    	if _, err := kr.Open(old, ad); !errors.Is(err, ErrUnknownKey) {
    		t.Errorf("old ciphertext after retiring key 1: error = %v, want ErrUnknownKey", err)
    	}
    	if _, err := kr.Reencrypt(old, ad); !errors.Is(err, ErrUnknownKey) {
    		t.Errorf("Reencrypt of a retired-key ciphertext: error = %v, want ErrUnknownKey", err)
    	}
    }

    func TestHostileInput(t *testing.T) {
    	kr := ring(t)
    	kr.Add(2, bytes.Repeat([]byte{2}, 32), false)
    	s, _ := kr.Seal([]byte("ghp_alice"), ad)

    	relabelled := bytes.Clone(s)
    	relabelled[4] = 2 // claim key 2
    	wrongVersion := bytes.Clone(s)
    	wrongVersion[0] = 2
    	tampered := bytes.Clone(s)
    	tampered[len(tampered)-1] ^= 1

    	for name, tc := range map[string]struct {
    		in   []byte
    		ad   []byte
    		want error
    	}{
    		"empty":                 {nil, ad, ErrBadFormat},
    		"short header":          {s[:3], ad, ErrBadFormat},
    		"unknown version":       {wrongVersion, ad, ErrBadFormat},
    		"unknown key ID":        {append([]byte{1, 0, 0, 0, 9}, s[headerSize:]...), ad, ErrUnknownKey},
    		"key ID rewritten to 2": {relabelled, ad, ErrDecrypt},
    		"tampered body":         {tampered, ad, ErrDecrypt},
    		"wrong associated data": {s, []byte("wifi-password"), ErrDecrypt},
    		"header only":           {s[:headerSize], ad, ErrDecrypt},
    	} {
    		if _, err := kr.Open(tc.in, tc.ad); !errors.Is(err, tc.want) {
    			t.Errorf("%s: Open error = %v, want %v", name, err, tc.want)
    		}
    	}
    }
---

Every key eventually needs replacing, and every algorithm might one day need
replacing too. Systems that plan for both are **crypto-agile**. Systems that don't end
up stuck, like the many that couldn't leave SHA-1 or 1024-bit RSA for years.

## Why keys rotate

- **Compromise.** A key leaked (in a log, a laptop theft, a departing employee's
  backup). You must stop using it now and re-protect data.
- **Usage limits.** AES-GCM with random nonces should encrypt at most about 2^32
  messages per key. High-volume systems hit that.
- **Policy.** Many standards require rotation on a schedule, which also keeps the
  procedure rehearsed, so it works on the day it's needed.

## Make ciphertexts self-describing

The key idea: every ciphertext carries a small **header** saying how it was made.
Keybox's is five bytes:

```
format version (1 byte) || key ID (4 bytes, big-endian) || AES-GCM nonce || ciphertext || tag
```

- The **key ID** tells `Open` which key to use, so old and new keys can coexist.
- The **format version** says which algorithm and layout follow. Version 2 could switch
  to a different AEAD or a post-quantum wrapping, and old version 1 data still decodes.
- The header is fed into the AEAD's **associated data**, so it's authenticated:
  Mallory can't relabel a ciphertext as another key or version.

Unknown versions and unknown key IDs are rejected explicitly. Never "fall back" to
trying other keys or older algorithms: a downgrade path is an attack path.

You've already seen this pattern: the `pbkdf2-sha256$600000$salt$hash` strings in
[Password Hashing](/courses/learn-http-servers/authentication/password-hashing), the
vault header in chapter 5, HPKE's ciphersuite IDs, and TLS's version negotiation.

## A rotation, step by step

1. **Add** the new key to the keyring and make it current. New data uses it right away;
   old data stays readable.
2. **Re-encrypt** old data in the background: open with the key named in its header,
   seal with the current key. Use the same associated data.
3. **Retire** the old key once nothing uses it, and destroy every copy, including
   backups of the keyring.

For a *compromised* key, go through the steps urgently, and remember that anyone who
copied old ciphertext while the key was exposed can still read it. Rotation protects
the future, not the past.

With Keybox's key hierarchy from chapter 6, most rotations are cheap: rotating a KEK
means re-wrapping one data key; rotating the data key means re-encrypting each entry,
which is a few hundred small seals.

## Agility, with restraint

Agility doesn't mean supporting every algorithm. Each option is code to maintain and
a potential downgrade target. Support one current choice per format version, decode
older versions for migration only, and remove them when you can.

## Your task

Complete the `Keyring`. `Add`, `Retire` and `header` are provided.

1. `Seal(plaintext, ad)`: look up the current key (`ErrNoKey` if missing), build
   `h := header(k.current)`, and return `aead.Seal(h, nil, plaintext, h+ad)`. Build the
   combined associated data in a fresh slice (`append(bytes.Clone(h), ad...)`), not by
   appending to `h`, which `Seal` is also appending to.
2. `Open(sealed, ad)`: `ErrBadFormat` if shorter than `headerSize` or the first byte
   isn't `formatV1`; read the key ID with `binary.BigEndian.Uint32(sealed[1:5])`;
   `ErrUnknownKey` if it isn't in the keyring; open the rest with
   `sealed[:headerSize]+ad` as associated data, mapping failure to `ErrDecrypt`.
3. `Reencrypt(sealed, ad)`: `Open` then `Seal`, returning any error.
