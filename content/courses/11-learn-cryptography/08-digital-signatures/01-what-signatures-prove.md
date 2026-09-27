---
title: What Signatures Prove
quiz:
  - question: Alice and the Keybox server share an HMAC key. The server shows Bob a message with a valid HMAC tag and says Alice wrote it. Why can't Bob be sure?
    options:
      - text: HMAC tags can be forged by anyone
      - text: The server holds the same key, so it could have produced the tag itself; a MAC can't prove *which* key holder made it
        correct: true
      - text: HMAC doesn't provide integrity
      - text: Bob doesn't have the key, but he could still be sure
    explanation: |
      MACs are symmetric: anyone who can verify can also create. A signature is made
      with a private key only Alice holds and verified with her public key, so a valid
      signature points at Alice (or at whoever stole her key).
  - question: A valid Ed25519 signature from Alice on a Keybox bundle proves which of these?
    options:
      - text: The bundle is recent
      - text: Alice intended Bob, specifically, to receive it
      - text: Someone holding Alice's private key signed exactly these bytes
        correct: true
      - text: The contents are true
    explanation: |
      That's all a signature says. Freshness, the intended recipient and meaning have
      to be *inside* the signed bytes (timestamps, recipient IDs, labels) to be covered
      at all.
---

Chapter 7 ended with two gaps: an HPKE share doesn't say who sent it, and a public key
from the directory might be Mallory's. **Digital signatures** close both.

## Signatures vs MACs

A signature scheme has three operations:

- `GenerateKey()` gives a **private signing key** and a **public verification key**.
- `Sign(private, message)` gives a signature.
- `Verify(public, message, signature)` gives true or false.

Like a MAC, it proves the message wasn't modified and came from a key holder. Unlike a
MAC, **verifying requires only the public key**, so anyone can check, and nobody who can
check can forge. That enables three things MACs can't do:

1. **Third parties can verify.** Bob checks Alice's bundle without sharing any secret
   with her.
2. **Non-repudiation.** Only Alice's private key could have produced the signature, so
   she can't plausibly claim the server made it up.
3. **One-to-many.** Keybox's release key signs once; millions of installs verify.

## Using them in Go

```go
package main

import (
	"crypto/ed25519"
	"fmt"
)

func main() {
	pub, priv, err := ed25519.GenerateKey(nil) // nil: use crypto/rand
	if err != nil {
		panic(err)
	}
	msg := []byte("keybox bundle: 3 secrets for bob")
	sig := ed25519.Sign(priv, msg)

	fmt.Println(len(pub), len(sig))
	fmt.Println(ed25519.Verify(pub, msg, sig))
	fmt.Println(ed25519.Verify(pub, []byte("keybox bundle: 3 secrets for eve"), sig))
}
```

```
32 64
true
false
```

## What a signature doesn't prove

A signature binds a key to a sequence of bytes. It says nothing else, so everything else
must go **into** the signed bytes:

- **Freshness.** A signature from 2024 is still valid in 2030. Include a timestamp or
  sequence number and check it, or Mallory replays old bundles.
- **Audience.** "Share with Bob" signed by Alice can be forwarded to Carol unless the
  recipient is inside the signed data.
- **Meaning.** A signature over bytes that could be parsed two ways endorses both
  readings. Sign a precise, unambiguous encoding with a label (lesson 4).
- **Identity.** The signature proves "the holder of *this* public key". Whether that
  key is really Alice's is a separate question.

## Binding keys to people

That last point is the man-in-the-middle problem from chapter 7, and signatures are the
standard answer, in a few flavours:

- **Trust on first use (TOFU):** remember the key you saw first, warn loudly if it ever
  changes. SSH does this.
- **Out-of-band verification:** compare fingerprints in person. Signal's safety numbers.
- **Someone you trust signs the binding.** A certificate authority signs "this key
  belongs to keybox.example" (chapter 9). Or Keybox's own directory signs each user's
  keys, and clients publish those signatures to a transparency log so a misbehaving
  directory gets caught.

Keybox users each get two key pairs: an HPKE key for receiving shares, and an **Ed25519
key** for signing what they send. The signing key's fingerprint is what users verify,
and it signs the encryption key, so one verification covers both.
