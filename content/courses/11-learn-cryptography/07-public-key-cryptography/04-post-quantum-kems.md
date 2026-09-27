---
title: Post-Quantum KEMs
quiz:
  - question: Mallory flips a bit in an ML-KEM ciphertext before Bob decapsulates it. What does `dk.Decapsulate` return?
    options:
      - text: An error saying the ciphertext was modified
      - text: A panic
      - text: No error, but a shared key that doesn't match Alice's, so the AEAD layer on top fails
        correct: true
      - text: Alice's original shared key
    explanation: |
      ML-KEM uses *implicit rejection*: a malformed-but-right-length ciphertext yields
      a pseudorandom key instead of an error, which denies attackers a useful error
      oracle. The mismatch is caught when the AEAD's `Open` fails. A KEM alone never
      authenticates anything.
  - question: What is a "hybrid" key exchange such as X25519MLKEM768?
    options:
      - text: A key exchange that alternates between two algorithms on each connection
      - text: Running X25519 and ML-KEM-768 together and combining both shared secrets, so the result is secure as long as *either* one is unbroken
        correct: true
      - text: ML-KEM with a smaller key
      - text: X25519 with a longer key
    explanation: |
      X25519 has decades of analysis; ML-KEM resists quantum attacks but is newer.
      Combining them protects against both a quantum computer and an unexpected flaw
      in the new scheme. Go's TLS has used X25519MLKEM768 by default since Go 1.24.
exercise:
  starter: |
    package main

    import (
    	"bytes"
    	"crypto/mlkem"
    	"errors"
    	"fmt"
    )

    // loadDeviceKey rebuilds a device's ML-KEM-768 decapsulation key from its
    // stored 64-byte seed. It returns an error for a seed of the wrong size.
    func loadDeviceKey(seed []byte) (*mlkem.DecapsulationKey768, error) {
    	// ?
    	return nil, errors.New("not implemented")
    }

    // encapsulateTo parses a recipient's 1184-byte encapsulation key and
    // returns a fresh 32-byte shared key plus the 1088-byte ciphertext to send.
    func encapsulateTo(encapsulationKey []byte) (sharedKey, ciphertext []byte, err error) {
    	// ?
    	return nil, nil, errors.New("not implemented")
    }

    func main() {
    	bob, err := loadDeviceKey(bytes.Repeat([]byte{0x42}, 64))
    	if err != nil {
    		fmt.Println("load:", err)
    		return
    	}
    	published := bob.EncapsulationKey().Bytes()
    	fmt.Println("bob's encapsulation key:", len(published), "bytes")

    	key, ct, err := encapsulateTo(published)
    	fmt.Printf("alice: %d-byte key, %d-byte ciphertext, err %v\n", len(key), len(ct), err)

    	bobsKey, err := bob.Decapsulate(ct)
    	fmt.Println("keys match:", bytes.Equal(key, bobsKey), err)
    }
  solution: |
    package main

    import (
    	"bytes"
    	"crypto/mlkem"
    	"fmt"
    )

    func loadDeviceKey(seed []byte) (*mlkem.DecapsulationKey768, error) {
    	return mlkem.NewDecapsulationKey768(seed)
    }

    func encapsulateTo(encapsulationKey []byte) (sharedKey, ciphertext []byte, err error) {
    	ek, err := mlkem.NewEncapsulationKey768(encapsulationKey)
    	if err != nil {
    		return nil, nil, err
    	}
    	sharedKey, ciphertext = ek.Encapsulate()
    	return sharedKey, ciphertext, nil
    }

    func main() {
    	bob, err := loadDeviceKey(bytes.Repeat([]byte{0x42}, 64))
    	if err != nil {
    		fmt.Println("load:", err)
    		return
    	}
    	published := bob.EncapsulationKey().Bytes()
    	fmt.Println("bob's encapsulation key:", len(published), "bytes")

    	key, ct, err := encapsulateTo(published)
    	fmt.Printf("alice: %d-byte key, %d-byte ciphertext, err %v\n", len(key), len(ct), err)

    	bobsKey, err := bob.Decapsulate(ct)
    	fmt.Println("keys match:", bytes.Equal(key, bobsKey), err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/mlkem"
    	"testing"
    )

    func TestLoadDeviceKey(t *testing.T) {
    	seed := bytes.Repeat([]byte{0x42}, mlkem.SeedSize)
    	a, err := loadDeviceKey(seed)
    	if err != nil {
    		t.Fatalf("loadDeviceKey(64-byte seed) error = %v", err)
    	}
    	b, _ := loadDeviceKey(seed)
    	if !bytes.Equal(a.EncapsulationKey().Bytes(), b.EncapsulationKey().Bytes()) {
    		t.Errorf("loading the same seed twice gave different keys; use mlkem.NewDecapsulationKey768(seed)")
    	}
    	if !bytes.Equal(a.Bytes(), seed) {
    		t.Errorf("key.Bytes() should give back the seed")
    	}
    	for _, n := range []int{0, 32, 63, 65} {
    		if _, err := loadDeviceKey(make([]byte, n)); err == nil {
    			t.Errorf("loadDeviceKey(%d-byte seed) error = nil, want an error", n)
    		}
    	}
    }

    func TestEncapsulateTo(t *testing.T) {
    	bob, _ := mlkem.GenerateKey768()
    	key, ct, err := encapsulateTo(bob.EncapsulationKey().Bytes())
    	if err != nil {
    		t.Fatalf("encapsulateTo(valid key) error = %v", err)
    	}
    	if len(key) != mlkem.SharedKeySize || len(ct) != mlkem.CiphertextSize768 {
    		t.Fatalf("encapsulateTo gave a %d-byte key and %d-byte ciphertext, want %d and %d",
    			len(key), len(ct), mlkem.SharedKeySize, mlkem.CiphertextSize768)
    	}
    	got, err := bob.Decapsulate(ct)
    	if err != nil || !bytes.Equal(got, key) {
    		t.Errorf("Bob's Decapsulate = %x, %v; want Alice's key %x", got, err, key)
    	}
    	key2, ct2, _ := encapsulateTo(bob.EncapsulationKey().Bytes())
    	if bytes.Equal(key, key2) || bytes.Equal(ct, ct2) {
    		t.Errorf("two encapsulations gave the same key or ciphertext; each must be fresh")
    	}
    	for _, n := range []int{0, 32, 1183, 1185} {
    		if _, _, err := encapsulateTo(make([]byte, n)); err == nil {
    			t.Errorf("encapsulateTo(%d-byte key) error = nil, want an error", n)
    		}
    	}
    }
---

Everything in this chapter so far, like all widely deployed public-key cryptography,
would fall to a large quantum computer. Go's standard library already has the
replacement.

## Harvest now, decrypt later

Shor's algorithm, run on a big enough fault-tolerant quantum computer, solves the math
problems behind RSA, finite-field Diffie-Hellman and elliptic curves. Nobody knows when
(or whether) such a machine will exist. But an adversary can **record** encrypted
traffic today and decrypt it once one does. For Keybox secrets that must stay secret for
decades, that's a present-day problem for key exchange.

(Symmetric crypto is in much better shape: Grover's algorithm only halves the effective
key length, which AES-256 shrugs off. Signatures are less urgent, since a forged
signature in 2040 can't retroactively fool a check made today, though long-lived keys
such as root CAs need migrating too.)

## KEMs

The post-quantum replacement for key exchange is a **key encapsulation mechanism**
(KEM). Instead of both sides contributing a key pair, it works like this:

- Bob publishes an **encapsulation key** and keeps a **decapsulation key**.
- Alice calls `Encapsulate` on Bob's encapsulation key and gets two things: a fresh
  random **shared key** and a **ciphertext** that encapsulates it.
- Bob calls `Decapsulate` on the ciphertext and gets the same shared key.

That's exactly the shape of the ephemeral-static ECDH from last lesson, which is really
a KEM built from Diffie-Hellman (HPKE calls it "DHKEM").

## ML-KEM in Go

NIST standardized **ML-KEM** (FIPS 203, based on CRYSTALS-Kyber) in 2024, and Go 1.24
added `crypto/mlkem`:

```go
package main

import (
	"bytes"
	"crypto/mlkem"
	"fmt"
)

func main() {
	dk, err := mlkem.GenerateKey768() // Bob's decapsulation (private) key
	if err != nil {
		panic(err)
	}
	ek := dk.EncapsulationKey() // Bob's encapsulation (public) key

	sharedKey, ciphertext := ek.Encapsulate()  // Alice
	bobsKey, err := dk.Decapsulate(ciphertext) // Bob
	fmt.Println(len(ek.Bytes()), len(ciphertext), len(sharedKey), bytes.Equal(sharedKey, bobsKey), err)
}
```

```
1184 1088 32 true <nil>
```

The sizes are the main practical difference from X25519: a 1,184-byte public key and a
1,088-byte ciphertext instead of 32 bytes each. ML-KEM is fast, though, faster than
X25519 in many benchmarks.

- **ML-KEM-768** targets roughly AES-192-level security and is the recommended default.
  **ML-KEM-1024** (`GenerateKey1024`) is available for higher margins.
- A decapsulation key serializes as a 64-byte **seed** (`dk.Bytes()`), and
  `mlkem.NewDecapsulationKey768(seed)` rebuilds it. Store the seed, wrapped like any
  other private key.
- The 32-byte shared key is ready to use as a symmetric key, or to feed into HKDF with
  a label.
- **Implicit rejection:** decapsulating a tampered (right-length) ciphertext returns a
  different, pseudorandom key rather than an error. Wrong-length input does return an
  error. Either way, the AEAD layer on top is what detects tampering.

## Hybrids

ML-KEM is much younger than elliptic curves, so current practice is **hybrid**: run
X25519 and ML-KEM together and combine the two shared secrets. An attacker must break
both. Since Go 1.24, `crypto/tls` offers the hybrid **X25519MLKEM768** by default, and
Go 1.26 added `SecP256r1MLKEM768` and `SecP384r1MLKEM1024`. If you've made an HTTPS
request with a recent Go to a server that supports it, you've already used post-quantum
key exchange.

## Your task

Give Keybox's devices ML-KEM keys:

1. `loadDeviceKey(seed)`: rebuild an ML-KEM-768 decapsulation key from its 64-byte seed
   with `mlkem.NewDecapsulationKey768`, returning its error for a bad seed.
2. `encapsulateTo(encapsulationKey)`: parse the recipient's key bytes with
   `mlkem.NewEncapsulationKey768` (returning any error), then return `Encapsulate()`'s
   shared key and ciphertext.
