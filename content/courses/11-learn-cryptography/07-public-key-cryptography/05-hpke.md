---
title: HPKE
quiz:
  - question: What are HPKE's three configurable components?
    options:
      - text: A hash, a block cipher and a padding scheme
      - text: A KEM, a KDF and an AEAD
        correct: true
      - text: A signature scheme, a MAC and a cipher
      - text: A curve, a nonce and a salt
    explanation: |
      An HPKE ciphersuite is KEM + KDF + AEAD, for example
      MLKEM768X25519 + HKDF-SHA256 + AES-256-GCM. The KEM establishes a shared secret,
      the KDF turns it into keys and nonces, and the AEAD encrypts.
  - question: Keybox seals with `info = "keybox share v2"`. Bob's client opens with `info = "keybox share v1"`. What happens?
    options:
      - text: It works; `info` is only metadata
      - text: '`Open` fails, because `info` is mixed into the key schedule and both sides must use exactly the same value'
        correct: true
      - text: It decrypts to garbage without an error
    explanation: |
      `info` is bound into key derivation, so it acts like a label for the whole
      exchange. A mismatch gives different keys and the AEAD rejects the ciphertext.
      That's domain separation, for free.
exercise:
  starter: |
    package main

    import (
    	"crypto/hpke"
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: cannot open share")

    const shareInfo = "keybox share v2"

    // Keybox's HPKE ciphersuite: X-Wing (ML-KEM-768 + X25519), HKDF-SHA256,
    // AES-256-GCM.
    var (
    	kem  = hpke.MLKEM768X25519()
    	kdf  = hpke.HKDFSHA256()
    	aead = hpke.AES256GCM()
    )

    // shareHPKE parses the recipient's public key bytes with kem.NewPublicKey
    // and seals plaintext to it with hpke.Seal, using shareInfo as info.
    func shareHPKE(recipientPublic, plaintext []byte) ([]byte, error) {
    	// ?
    	return nil, errors.New("not implemented")
    }

    // openHPKE opens a share with the recipient's private key. Any failure
    // returns ErrDecrypt.
    func openHPKE(priv hpke.PrivateKey, share []byte) ([]byte, error) {
    	// ?
    	return nil, ErrDecrypt
    }

    func main() {
    	bob, err := kem.GenerateKey()
    	if err != nil {
    		panic(err)
    	}
    	published := bob.PublicKey().Bytes()
    	fmt.Println("bob's public key:", len(published), "bytes")

    	share, err := shareHPKE(published, []byte("db-password: s3cr3t-team-pw"))
    	fmt.Println("share:", len(share), "bytes", err)

    	pt, err := openHPKE(bob, share)
    	fmt.Printf("bob opens: %q %v\n", pt, err)
    }
  solution: |
    package main

    import (
    	"crypto/hpke"
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: cannot open share")

    const shareInfo = "keybox share v2"

    var (
    	kem  = hpke.MLKEM768X25519()
    	kdf  = hpke.HKDFSHA256()
    	aead = hpke.AES256GCM()
    )

    func shareHPKE(recipientPublic, plaintext []byte) ([]byte, error) {
    	pub, err := kem.NewPublicKey(recipientPublic)
    	if err != nil {
    		return nil, err
    	}
    	return hpke.Seal(pub, kdf, aead, []byte(shareInfo), plaintext)
    }

    func openHPKE(priv hpke.PrivateKey, share []byte) ([]byte, error) {
    	pt, err := hpke.Open(priv, kdf, aead, []byte(shareInfo), share)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func main() {
    	bob, err := kem.GenerateKey()
    	if err != nil {
    		panic(err)
    	}
    	published := bob.PublicKey().Bytes()
    	fmt.Println("bob's public key:", len(published), "bytes")

    	share, err := shareHPKE(published, []byte("db-password: s3cr3t-team-pw"))
    	fmt.Println("share:", len(share), "bytes", err)

    	pt, err := openHPKE(bob, share)
    	fmt.Printf("bob opens: %q %v\n", pt, err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/hpke"
    	"errors"
    	"testing"
    )

    func TestHPKERoundTrip(t *testing.T) {
    	bob, _ := kem.GenerateKey()
    	for _, msg := range []string{"", "db-password: s3cr3t-team-pw"} {
    		share, err := shareHPKE(bob.PublicKey().Bytes(), []byte(msg))
    		if err != nil {
    			t.Fatalf("shareHPKE error = %v", err)
    		}
    		if want := 1120 + len(msg) + 16; len(share) != want {
    			t.Errorf("share of %d bytes is %d bytes long, want %d (X-Wing enc + ciphertext + tag)", len(msg), len(share), want)
    		}
    		got, err := openHPKE(bob, share)
    		if err != nil || string(got) != msg {
    			t.Errorf("openHPKE = %q, %v; want %q", got, err, msg)
    		}
    	}
    }

    func TestHPKERejects(t *testing.T) {
    	bob, _ := kem.GenerateKey()
    	eve, _ := kem.GenerateKey()
    	share, _ := shareHPKE(bob.PublicKey().Bytes(), []byte("db-password"))

    	if _, err := openHPKE(eve, share); !errors.Is(err, ErrDecrypt) {
    		t.Errorf("openHPKE(wrong recipient) error = %v, want ErrDecrypt", err)
    	}
    	for _, i := range []int{0, 500, 1119, 1125, len(share) - 1} {
    		bad := bytes.Clone(share)
    		bad[i] ^= 1
    		if _, err := openHPKE(bob, bad); !errors.Is(err, ErrDecrypt) {
    			t.Errorf("openHPKE(byte %d flipped) error = %v, want ErrDecrypt", i, err)
    		}
    	}
    	if _, err := openHPKE(bob, share[:100]); !errors.Is(err, ErrDecrypt) {
    		t.Errorf("openHPKE(truncated) error = %v, want ErrDecrypt", err)
    	}
    	otherInfo, _ := hpke.Seal(bob.PublicKey(), kdf, aead, []byte("keybox share v1"), []byte("db-password"))
    	if _, err := openHPKE(bob, otherInfo); !errors.Is(err, ErrDecrypt) {
    		t.Errorf("openHPKE(share sealed with info \"keybox share v1\") error = %v, want ErrDecrypt; use shareInfo", err)
    	}
    }

    func TestBadPublicKey(t *testing.T) {
    	if _, err := shareHPKE(make([]byte, 32), []byte("x")); err == nil {
    		t.Errorf("shareHPKE(32-byte public key) error = nil, want an error (X-Wing keys are 1216 bytes)")
    	}
    }
---

You built hybrid encryption by hand: ECDH, HKDF with the right salt and label, AES-GCM,
a custom envelope format. It works, but it's *your* protocol, with *your* choices and no
external review. **HPKE** (Hybrid Public Key Encryption, RFC 9180) is the
standardized version of exactly this, and Go 1.26 added it as `crypto/hpke`.

## What HPKE standardizes

HPKE fixes every decision you made by hand, and a few you didn't:

- how the KEM's shared secret, the public keys and your `info` string feed into key
  derivation,
- how the AEAD key and nonces are derived (no random nonces to manage),
- the wire format: the KEM's encapsulated key followed by the ciphertext,
- modes for multiple messages, exporting extra secrets, and more.

It's used by TLS Encrypted Client Hello, the Messaging Layer Security protocol (MLS),
Oblivious HTTP and others, so it has had a lot of scrutiny and ships with test vectors.

## Ciphersuites

You choose one of each:

| Component | Options in `crypto/hpke` |
| --- | --- |
| KEM | `DHKEM(ecdh.X25519())`, `DHKEM(ecdh.P256())`, ..., `MLKEM768()`, `MLKEM1024()`, **`MLKEM768X25519()`** (X-Wing), `MLKEM768P256()`, `MLKEM1024P384()` |
| KDF | `HKDFSHA256()`, `HKDFSHA384()`, `HKDFSHA512()`, `SHAKE128()`, `SHAKE256()` |
| AEAD | `AES128GCM()`, `AES256GCM()`, `ChaCha20Poly1305()`, `ExportOnly()` |

Keybox v2 uses **MLKEM768X25519**, also called **X-Wing**: a hybrid KEM combining the
post-quantum ML-KEM-768 with X25519, so shares recorded today stay safe even against a
future quantum computer, *and* stay safe if ML-KEM turns out to have a classical flaw.

## Single-shot Seal and Open

For one message per recipient, which is Keybox's case, two functions do everything:

```go
kem := hpke.MLKEM768X25519()
bob, err := kem.GenerateKey()        // hpke.PrivateKey
published := bob.PublicKey().Bytes() // 1216 bytes, goes in the directory

pub, err := kem.NewPublicKey(published)
share, err := hpke.Seal(pub, hpke.HKDFSHA256(), hpke.AES256GCM(), []byte("keybox share v2"), plaintext)

plaintext, err = hpke.Open(bob, hpke.HKDFSHA256(), hpke.AES256GCM(), []byte("keybox share v2"), share)
```

`Seal` returns the encapsulated key (1,120 bytes for X-Wing) followed by the AEAD
ciphertext and tag, and `Open` takes that concatenation. The `info` argument is your
application label, mixed into key derivation: a mismatch makes `Open` fail. For many
messages to one recipient, `hpke.NewSender` and `hpke.NewRecipient` keep a context with
a message counter, so you encapsulate only once.

Private keys serialize with `priv.Bytes()` and load with `kem.NewPrivateKey`; wrap them
under the vault key like any other secret.

## What HPKE still doesn't do

Like your envelope, base-mode HPKE doesn't authenticate the **sender**. RFC 9180
defines authenticated modes for DH-based KEMs, but they don't work with KEMs like
ML-KEM, so the common pattern (and Keybox's) is: HPKE for confidentiality, plus a
**signature** over the result for authenticity. That's next chapter.

## Your task

Move Keybox's sharing to HPKE:

1. `shareHPKE(recipientPublic, plaintext)`: parse the key bytes with
   `kem.NewPublicKey` (returning any error) and return
   `hpke.Seal(pub, kdf, aead, []byte(shareInfo), plaintext)`.
2. `openHPKE(priv, share)`: call `hpke.Open` with the same suite and info, and return
   `ErrDecrypt` for any failure.
