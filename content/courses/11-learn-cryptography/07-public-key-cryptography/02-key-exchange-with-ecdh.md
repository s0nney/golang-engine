---
title: Key Exchange with ECDH
quiz:
  - question: Alice and Bob each run `ecdh.X25519().GenerateKey`, swap public keys, and call `ECDH`. What does Eve, who saw both public keys, learn about the shared secret?
    options:
      - text: The shared secret, since it's computed from the public keys
      - text: Nothing useful; computing it needs one of the private keys
        correct: true
      - text: Its first half
      - text: The private keys
    explanation: |
      That's the Diffie-Hellman magic: `alicePriv x bobPub == bobPriv x alicePub`, but
      from the two public values alone, recovering the shared secret is believed to be
      infeasible (the elliptic-curve Diffie-Hellman problem).
  - question: Why shouldn't Keybox use the 32 bytes from `priv.ECDH(peer)` directly as an AES key?
    options:
      - text: They're the wrong length for AES
      - text: The raw output isn't a uniformly random key and isn't bound to any context; run it through HKDF with the public keys and a label first
        correct: true
      - text: ECDH output is public
      - text: AES can't use keys from elliptic curves
    explanation: |
      32 bytes is a valid AES-256 length, which makes this mistake tempting. But the
      shared secret is a curve coordinate with structure, and using it directly skips
      binding it to the parties and the purpose. HKDF's extract and expand steps fix
      both.
exercise:
  starter: |
    package main

    import (
    	"crypto/ecdh"
    	"encoding/hex"
    	"errors"
    	"fmt"
    )

    // sharedSecret parses peerPublic as a 32-byte X25519 public key and returns
    // the raw X25519 shared secret with priv. It returns an error for a
    // malformed public key, or if ECDH fails (for example a low-order point).
    func sharedSecret(priv *ecdh.PrivateKey, peerPublic []byte) ([]byte, error) {
    	// ?
    	return nil, errors.New("not implemented")
    }

    // fingerprint returns the first 16 bytes of SHA-256(pub.Bytes()) as hex,
    // in space-separated groups of 4 characters, e.g. "300c 9c96 ... 4011".
    func fingerprint(pub *ecdh.PublicKey) string {
    	// ?
    	return ""
    }

    func mustHex(s string) []byte {
    	b, err := hex.DecodeString(s)
    	if err != nil {
    		panic(err)
    	}
    	return b
    }

    func main() {
    	// RFC 7748, section 6.1.
    	alice, _ := ecdh.X25519().NewPrivateKey(mustHex("77076d0a7318a57d3c16c17251b26645df4c2f87ebc0992ab177fba51db92c2a"))
    	bob, _ := ecdh.X25519().NewPrivateKey(mustHex("5dab087e624a8a4b79e17f8b83800ee66f3bb1292618b6fd1c2f8b27ff88e0eb"))

    	s1, err1 := sharedSecret(alice, bob.PublicKey().Bytes())
    	s2, err2 := sharedSecret(bob, alice.PublicKey().Bytes())
    	fmt.Printf("alice computes: %x %v\n", s1, err1)
    	fmt.Printf("bob computes:   %x %v\n", s2, err2)
    	fmt.Println("alice's fingerprint:", fingerprint(alice.PublicKey()))
    }
  solution: |
    package main

    import (
    	"crypto/ecdh"
    	"crypto/sha256"
    	"encoding/hex"
    	"fmt"
    	"strings"
    )

    func sharedSecret(priv *ecdh.PrivateKey, peerPublic []byte) ([]byte, error) {
    	pub, err := ecdh.X25519().NewPublicKey(peerPublic)
    	if err != nil {
    		return nil, err
    	}
    	return priv.ECDH(pub)
    }

    func fingerprint(pub *ecdh.PublicKey) string {
    	sum := sha256.Sum256(pub.Bytes())
    	h := hex.EncodeToString(sum[:16])
    	var groups []string
    	for i := 0; i < len(h); i += 4 {
    		groups = append(groups, h[i:i+4])
    	}
    	return strings.Join(groups, " ")
    }

    func mustHex(s string) []byte {
    	b, err := hex.DecodeString(s)
    	if err != nil {
    		panic(err)
    	}
    	return b
    }

    func main() {
    	alice, _ := ecdh.X25519().NewPrivateKey(mustHex("77076d0a7318a57d3c16c17251b26645df4c2f87ebc0992ab177fba51db92c2a"))
    	bob, _ := ecdh.X25519().NewPrivateKey(mustHex("5dab087e624a8a4b79e17f8b83800ee66f3bb1292618b6fd1c2f8b27ff88e0eb"))

    	s1, err1 := sharedSecret(alice, bob.PublicKey().Bytes())
    	s2, err2 := sharedSecret(bob, alice.PublicKey().Bytes())
    	fmt.Printf("alice computes: %x %v\n", s1, err1)
    	fmt.Printf("bob computes:   %x %v\n", s2, err2)
    	fmt.Println("alice's fingerprint:", fingerprint(alice.PublicKey()))
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/ecdh"
    	"encoding/hex"
    	"testing"
    )

    func TestRFC7748(t *testing.T) {
    	alice, _ := ecdh.X25519().NewPrivateKey(mustHex("77076d0a7318a57d3c16c17251b26645df4c2f87ebc0992ab177fba51db92c2a"))
    	bobPub := mustHex("de9edb7d7b7dc1b4d35b61c2ece435373f8343c85b78674dadfc7e146f882b4f")
    	want := "4a5d9d5ba4ce2de1728e3bf480350f25e07e21c947d19e3376f09b3c1e161742"
    	got, err := sharedSecret(alice, bobPub)
    	if err != nil || hex.EncodeToString(got) != want {
    		t.Errorf("sharedSecret(RFC 7748 Alice, Bob's public key) = %x, %v; want %s", got, err, want)
    	}
    }

    func TestRandomKeysAgree(t *testing.T) {
    	for range 5 {
    		a, _ := ecdh.X25519().GenerateKey(nil)
    		b, _ := ecdh.X25519().GenerateKey(nil)
    		s1, err1 := sharedSecret(a, b.PublicKey().Bytes())
    		s2, err2 := sharedSecret(b, a.PublicKey().Bytes())
    		if err1 != nil || err2 != nil || !bytes.Equal(s1, s2) || len(s1) != 32 {
    			t.Fatalf("both sides should compute the same 32-byte secret: %x (%v) vs %x (%v)", s1, err1, s2, err2)
    		}
    	}
    }

    func TestBadPublicKeys(t *testing.T) {
    	a, _ := ecdh.X25519().GenerateKey(nil)
    	for name, pub := range map[string][]byte{
    		"empty":             nil,
    		"31 bytes":          make([]byte, 31),
    		"33 bytes":          make([]byte, 33),
    		"low-order (zeros)": make([]byte, 32),
    	} {
    		if s, err := sharedSecret(a, pub); err == nil {
    			t.Errorf("sharedSecret with %s public key = %x, nil; want an error", name, s)
    		}
    	}
    }

    func TestFingerprint(t *testing.T) {
    	for _, tt := range []struct{ pub, want string }{
    		{"8520f0098930a754748b7ddcb43ef75a0dbf3a0d26381af4eba4a98eaa9b4e6a", "300c 9c96 03b9 2a4b 39ed 3958 bf92 4011"},
    		{"de9edb7d7b7dc1b4d35b61c2ece435373f8343c85b78674dadfc7e146f882b4f", "f35e 5616 160a 30bf 3c6e 79fa 73c5 76d4"},
    	} {
    		pub, _ := ecdh.X25519().NewPublicKey(mustHex(tt.pub))
    		if got := fingerprint(pub); got != tt.want {
    			t.Errorf("fingerprint(%s...) = %q, want %q", tt.pub[:8], got, tt.want)
    		}
    	}
    }
---

Diffie-Hellman key exchange lets two parties who share nothing agree on a secret over
a public channel. With elliptic curves it's called **ECDH**, and Go's `crypto/ecdh`
makes it a couple of calls.

## How it works (the short version)

X25519 defines a public "base point" `G` on a curve and a way to "multiply" a point by
a number. Multiplying is fast; undoing it (finding the number given the result) is
infeasible.

- Alice picks a random private number `a` and publishes `A = a·G`.
- Bob picks `b` and publishes `B = b·G`.
- Alice computes `a·B = a·b·G`. Bob computes `b·A = b·a·G`. Same point.

Eve sees `A` and `B` but can't compute `a·b·G` without `a` or `b`.

## In Go

```go
package main

import (
	"bytes"
	"crypto/ecdh"
	"fmt"
)

func main() {
	curve := ecdh.X25519()
	alice, _ := curve.GenerateKey(nil) // secure randomness is always used
	bob, _ := curve.GenerateKey(nil)

	// They exchange alice.PublicKey().Bytes() and bob.PublicKey().Bytes().
	s1, _ := alice.ECDH(bob.PublicKey())
	s2, _ := bob.ECDH(alice.PublicKey())
	fmt.Println(len(s1), bytes.Equal(s1, s2))
}
```

```
32 true
```

- `curve.GenerateKey` makes a key pair. Since Go 1.26 the `io.Reader` argument is
  ignored and system randomness is always used, so passing `nil` is fine.
- `curve.NewPrivateKey(b)` and `curve.NewPublicKey(b)` load keys from bytes. For X25519
  both are exactly 32 bytes; anything else is an error.
- `priv.ECDH(pub)` returns the 32-byte shared secret. It returns an error for
  **low-order points**, such as an all-zero public key, which would give an all-zero
  "shared secret" an attacker could predict. Always check that error.

You can check an implementation against RFC 7748, which publishes Alice's and Bob's
private keys and the shared secret they should reach. The exercise does exactly that.

## The raw secret needs HKDF

Don't use `ECDH`'s output directly as a key. Instead run it through HKDF with a label,
and include both public keys, so the key is bound to *this* exchange between *these*
keys:

```go
secret, err := priv.ECDH(peer)
// ...
salt := append(myPub.Bytes(), peerPub.Bytes()...) // a fixed order both sides agree on
key, err := hkdf.Key(sha256.New, secret, salt, "keybox share v1", 32)
```

Next lesson you'll build this into Keybox's sharing.

## Fingerprints

To defend against a man-in-the-middle, users can compare **fingerprints** of their
public keys over a channel they trust (in person, on a video call). A fingerprint is a
short hash of the public key. Keybox shows the first 16 bytes of SHA-256 in groups of 4
hex digits, a common human-friendly format:

```
300c 9c96 03b9 2a4b 39ed 3958 bf92 4011
```

Truncating to 128 bits is fine here: an attacker would need a *second preimage* for a
specific key, which at 128 bits is still out of reach.

## Your task

1. `sharedSecret(priv, peerPublic)`: parse `peerPublic` with
   `ecdh.X25519().NewPublicKey`, then return `priv.ECDH(pub)`. Return errors from
   either step.
2. `fingerprint(pub)`: SHA-256 of `pub.Bytes()`, take the first 16 bytes, hex-encode
   them, and join groups of 4 characters with single spaces.

The tests use the RFC 7748 vectors and try malformed and low-order public keys.
