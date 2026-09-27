---
title: Hybrid Encryption
quiz:
  - question: Why does Alice generate a new *ephemeral* key pair for every share, instead of using her long-term key pair?
    options:
      - text: Long-term keys can't be used with ECDH
      - text: A fresh key pair per message gives a fresh shared secret and AES key every time, and the ephemeral private key is thrown away right after, so later compromise of Alice's device can't reveal it
        correct: true
      - text: It makes the envelope smaller
      - text: So Bob can verify it came from Alice
    explanation: |
      Ephemeral keys make every share independent. Note what they *don't* do: since
      anyone can generate an ephemeral key, the envelope doesn't prove who sent it.
      That's what signatures add in the next chapter.
exercise:
  starter: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/ecdh"
    	"crypto/hkdf"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: cannot open share")

    const shareLabel = "keybox share v1"

    // Envelope is what Alice uploads for Bob.
    type Envelope struct {
    	EphemeralPublic []byte // 32-byte X25519 public key, fresh for every share
    	Sealed          []byte // AES-256-GCM nonce || ciphertext || tag
    }

    // shareKey derives the AES key for a share from the X25519 shared secret,
    // binding it to both public keys and the label.
    func shareKey(secret, ephemeralPub, recipientPub []byte) ([]byte, error) {
    	salt := append(append([]byte{}, ephemeralPub...), recipientPub...)
    	return hkdf.Key(sha256.New, secret, salt, shareLabel, 32)
    }

    func newAEAD(key []byte) (cipher.AEAD, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	return cipher.NewGCMWithRandomNonce(block)
    }

    // shareSecret encrypts plaintext so that only the holder of recipient's
    // private key can open it.
    func shareSecret(recipient *ecdh.PublicKey, plaintext []byte) (Envelope, error) {
    	// ?
    	return Envelope{}, errors.New("not implemented")
    }

    // openEnvelope decrypts env with the recipient's private key. Every failure
    // returns ErrDecrypt.
    func openEnvelope(priv *ecdh.PrivateKey, env Envelope) ([]byte, error) {
    	// ?
    	return nil, ErrDecrypt
    }

    func main() {
    	bob, _ := ecdh.X25519().GenerateKey(nil)
    	env, err := shareSecret(bob.PublicKey(), []byte("db-password: s3cr3t-team-pw"))
    	fmt.Printf("envelope: %d-byte ephemeral key, %d-byte sealed secret, err %v\n", len(env.EphemeralPublic), len(env.Sealed), err)

    	pt, err := openEnvelope(bob, env)
    	fmt.Printf("bob opens: %q %v\n", pt, err)

    	eve, _ := ecdh.X25519().GenerateKey(nil)
    	_, err = openEnvelope(eve, env)
    	fmt.Println("eve opens:", err)
    }
  solution: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/ecdh"
    	"crypto/hkdf"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: cannot open share")

    const shareLabel = "keybox share v1"

    type Envelope struct {
    	EphemeralPublic []byte
    	Sealed          []byte
    }

    func shareKey(secret, ephemeralPub, recipientPub []byte) ([]byte, error) {
    	salt := append(append([]byte{}, ephemeralPub...), recipientPub...)
    	return hkdf.Key(sha256.New, secret, salt, shareLabel, 32)
    }

    func newAEAD(key []byte) (cipher.AEAD, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	return cipher.NewGCMWithRandomNonce(block)
    }

    func shareSecret(recipient *ecdh.PublicKey, plaintext []byte) (Envelope, error) {
    	eph, err := ecdh.X25519().GenerateKey(nil)
    	if err != nil {
    		return Envelope{}, err
    	}
    	secret, err := eph.ECDH(recipient)
    	if err != nil {
    		return Envelope{}, err
    	}
    	ephPub := eph.PublicKey().Bytes()
    	key, err := shareKey(secret, ephPub, recipient.Bytes())
    	if err != nil {
    		return Envelope{}, err
    	}
    	aead, err := newAEAD(key)
    	if err != nil {
    		return Envelope{}, err
    	}
    	return Envelope{EphemeralPublic: ephPub, Sealed: aead.Seal(nil, nil, plaintext, []byte(shareLabel))}, nil
    }

    func openEnvelope(priv *ecdh.PrivateKey, env Envelope) ([]byte, error) {
    	ephPub, err := ecdh.X25519().NewPublicKey(env.EphemeralPublic)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	secret, err := priv.ECDH(ephPub)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	key, err := shareKey(secret, env.EphemeralPublic, priv.PublicKey().Bytes())
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	aead, err := newAEAD(key)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	pt, err := aead.Open(nil, nil, env.Sealed, []byte(shareLabel))
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func main() {
    	bob, _ := ecdh.X25519().GenerateKey(nil)
    	env, err := shareSecret(bob.PublicKey(), []byte("db-password: s3cr3t-team-pw"))
    	fmt.Printf("envelope: %d-byte ephemeral key, %d-byte sealed secret, err %v\n", len(env.EphemeralPublic), len(env.Sealed), err)

    	pt, err := openEnvelope(bob, env)
    	fmt.Printf("bob opens: %q %v\n", pt, err)

    	eve, _ := ecdh.X25519().GenerateKey(nil)
    	_, err = openEnvelope(eve, env)
    	fmt.Println("eve opens:", err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/ecdh"
    	"errors"
    	"testing"
    )

    func TestShareRoundTrip(t *testing.T) {
    	bob, _ := ecdh.X25519().GenerateKey(nil)
    	for _, msg := range []string{"", "db-password: s3cr3t-team-pw", string(bytes.Repeat([]byte("k"), 5000))} {
    		env, err := shareSecret(bob.PublicKey(), []byte(msg))
    		if err != nil {
    			t.Fatalf("shareSecret error = %v", err)
    		}
    		if len(env.EphemeralPublic) != 32 {
    			t.Errorf("EphemeralPublic is %d bytes, want a 32-byte X25519 public key", len(env.EphemeralPublic))
    		}
    		if want := 12 + len(msg) + 16; len(env.Sealed) != want {
    			t.Errorf("Sealed is %d bytes for a %d-byte secret, want %d", len(env.Sealed), len(msg), want)
    		}
    		got, err := openEnvelope(bob, env)
    		if err != nil || string(got) != msg {
    			t.Errorf("openEnvelope(bob, share of %.20q) = %.20q, %v", msg, got, err)
    		}
    	}
    }

    func TestFreshEphemeralKeys(t *testing.T) {
    	bob, _ := ecdh.X25519().GenerateKey(nil)
    	a, _ := shareSecret(bob.PublicKey(), []byte("same"))
    	b, _ := shareSecret(bob.PublicKey(), []byte("same"))
    	if bytes.Equal(a.EphemeralPublic, b.EphemeralPublic) {
    		t.Errorf("two shares used the same ephemeral public key; generate a new key pair per share")
    	}
    	if bytes.Equal(a.EphemeralPublic, bob.PublicKey().Bytes()) {
    		t.Errorf("EphemeralPublic is the recipient's public key; it should be a fresh key of Alice's")
    	}
    }

    func TestOnlyRecipientCanOpen(t *testing.T) {
    	bob, _ := ecdh.X25519().GenerateKey(nil)
    	eve, _ := ecdh.X25519().GenerateKey(nil)
    	env, _ := shareSecret(bob.PublicKey(), []byte("db-password"))
    	if got, err := openEnvelope(eve, env); !errors.Is(err, ErrDecrypt) || got != nil {
    		t.Errorf("openEnvelope(eve, share for bob) = %q, %v; want nil, ErrDecrypt", got, err)
    	}
    }

    func TestTamperedEnvelopes(t *testing.T) {
    	bob, _ := ecdh.X25519().GenerateKey(nil)
    	env, _ := shareSecret(bob.PublicKey(), []byte("db-password"))
    	mallory, _ := ecdh.X25519().GenerateKey(nil)
    	for name, bad := range map[string]Envelope{
    		"flipped ciphertext bit":  {env.EphemeralPublic, flip(env.Sealed, 15)},
    		"flipped ephemeral bit":   {flip(env.EphemeralPublic, 3), env.Sealed},
    		"swapped ephemeral key":   {mallory.PublicKey().Bytes(), env.Sealed},
    		"low-order ephemeral key": {make([]byte, 32), env.Sealed},
    		"short ephemeral key":     {env.EphemeralPublic[:31], env.Sealed},
    		"truncated sealed data":   {env.EphemeralPublic, env.Sealed[:10]},
    	} {
    		if got, err := openEnvelope(bob, bad); !errors.Is(err, ErrDecrypt) || got != nil {
    			t.Errorf("%s: openEnvelope = %q, %v; want nil, ErrDecrypt", name, got, err)
    		}
    	}
    }

    func flip(b []byte, i int) []byte {
    	c := bytes.Clone(b)
    	c[i] ^= 1
    	return c
    }
---

ECDH gives two parties a shared secret, but it needs *both* of them online to swap
public keys. Keybox's sharing is asynchronous: Bob might be asleep when Alice shares.
The fix is to let Alice play both roles on her side, with a throwaway key pair.

## Ephemeral-static Diffie-Hellman

Bob's public key is already in the Keybox directory (his **static** key). To share with
him, Alice:

1. generates a fresh **ephemeral** X25519 key pair,
2. computes `ECDH(ephemeralPrivate, bobPublic)`,
3. derives an AES-256 key from that with HKDF, binding in both public keys and the
   label `"keybox share v1"`,
4. seals the secret with AES-GCM under that key,
5. uploads an **envelope**: the ephemeral *public* key and the sealed secret,
6. throws the ephemeral private key away.

Bob computes `ECDH(bobPrivate, ephemeralPublic)`, which is the same secret, derives
the same AES key, and opens the envelope.

```
Envelope {
    EphemeralPublic: 32 bytes
    Sealed:          nonce(12) || ciphertext || tag(16)
}
```

This construction is sometimes called **ECIES** (Elliptic Curve Integrated Encryption
Scheme). The overhead is 60 bytes per share, and a share costs one key generation and
one ECDH on each side: well under a millisecond.

## Details that matter

- **Bind the public keys into the KDF.** Keybox's `shareKey` uses the ephemeral and
  recipient public keys as the HKDF salt. That ties the AES key to this exact exchange.
- **Check ECDH's error.** A malicious envelope with a low-order ephemeral key makes
  `ECDH` fail; treat that like any other decryption failure.
- **One error for everything.** Malformed key, failed ECDH, failed `Open`: all return
  `ErrDecrypt`.
- **Associated data.** Pass the label (or more context, such as the secret's name and
  the recipient's user ID) as the AEAD's associated data.

## What it doesn't give you

The envelope proves **nothing about the sender**. Anyone can generate an ephemeral key
and encrypt something to Bob's public key, including Mallory or the server. Bob knows
only that the envelope was made for him. Authenticity needs a **signature** from Alice's
long-term signing key over the envelope, which is exactly what chapter 8 adds. And as
lesson 1 warned, Alice must also be sure the public key she encrypted to is really
Bob's.

## Should you build this yourself?

You're building it here to see how the parts fit. In production, prefer a
**standardized** construction with test vectors and security analysis. For this exact
job, that's **HPKE** (RFC 9180), which is in Go's standard library as `crypto/hpke` and
is lesson 5 of this chapter.

## Your task

Complete `shareSecret` and `openEnvelope`. `shareKey` and `newAEAD` are provided.

- `shareSecret(recipient, plaintext)`: generate an ephemeral key with
  `ecdh.X25519().GenerateKey(nil)`, compute `ECDH` with `recipient`, derive the key with
  `shareKey(secret, ephemeralPub, recipient.Bytes())`, and seal with
  `newAEAD(key).Seal(nil, nil, plaintext, []byte(shareLabel))`.
- `openEnvelope(priv, env)`: parse `env.EphemeralPublic`, compute `ECDH` with `priv`,
  derive the key with `shareKey(secret, env.EphemeralPublic, priv.PublicKey().Bytes())`,
  and open with the same associated data. Any failure returns `ErrDecrypt`.
