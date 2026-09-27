---
title: What Cryptography Is For
quiz:
  - question: Keybox stores your secrets on a server. Which property stops a curious server admin from *reading* them?
    options:
      - text: Integrity
      - text: Authenticity
      - text: Confidentiality
        correct: true
      - text: Availability
    explanation: |
      Confidentiality means only holders of the right key can read the data. Integrity
      (nobody changed it) and authenticity (it came from who you think) are separate
      promises, and you need separate tools, or a combined tool, to get them.
  - question: Bob receives a secret that Keybox says Alice shared. Which *two* properties does he need so he can trust it wasn't planted or altered by the server?
    options:
      - text: Confidentiality and availability
      - text: Integrity and authenticity
        correct: true
      - text: Only confidentiality, since encrypted data can't be changed
      - text: None, since the connection uses HTTPS
    explanation: |
      Bob needs to know the bytes weren't modified (integrity) and that they really came
      from Alice (authenticity). Encryption on its own doesn't give either, as you'll see
      in a couple of lessons. HTTPS only protects the hop to the server, and the server is
      exactly who Bob can't fully trust here.
---

Welcome to the last course on the roadmap. You already *use* cryptography: you've hashed
passwords with PBKDF2, signed tokens with HMAC and refused to set `InsecureSkipVerify`.
This course opens the box. You'll learn what each primitive actually promises, where
the sharp edges are, and how to combine the pieces without cutting yourself.

## Meet Keybox

The running project is **Keybox**, a small end-to-end-encrypted secrets vault. By the
end of the course it will:

1. **Store secrets encrypted at rest**, under a key derived from your password, so a
   stolen disk or database dump reveals nothing.
2. **Share a secret with another user** by encrypting it to their *public* key, so the
   Keybox server relays it without ever being able to read it.
3. **Sign exported bundles**, so the recipient can check who made them and that nobody
   changed a byte on the way.

Keybox's users are the traditional cast of cryptography: **Alice** and **Bob** want to
communicate, **Eve** eavesdrops, and **Mallory** actively tampers with messages.

## The three big promises

Almost everything cryptography does serves one of three goals:

| Goal | Question it answers | Keybox example |
| --- | --- | --- |
| **Confidentiality** | Who can read it? | Only Alice can read her stored API keys. |
| **Integrity** | Has it been changed? | A flipped bit in the vault file is detected, not silently accepted. |
| **Authenticity** | Who made it? | Bob can tell a bundle really came from Alice, not from Mallory. |

A fourth, **non-repudiation**, is a strong form of authenticity: the author can't later
deny having made something, because only they could have produced the proof. Digital
signatures give you that; shared-secret MACs don't, because both sides hold the key.

These goals are independent. Encryption hides content but, on its own, doesn't stop
anyone from changing it. A checksum detects accidental changes but not deliberate
ones, because the attacker can just recompute it. A lot of real-world breakage comes
from assuming one property implies another.

## What cryptography doesn't do

Cryptography turns a big secret (your data) into a small secret (a key). That's
enormously useful, but it moves the problem rather than solving it:

- If the key leaks, the protection is gone. Key management is most of the real work.
- It doesn't protect data while your program is using it. Malware on Alice's laptop
  can read her secrets after Keybox decrypts them.
- It doesn't hide *metadata* unless you design for it: who talked to whom, when, and
  roughly how much.
- It doesn't fix bugs in the code around it. Most breaches of "encrypted" systems come
  from mistakes in using the primitives, not from breaking them.

## Keybox's API, as a sketch

Here's the shape of what you'll build, as a set of Go signatures. Don't worry about the
types yet; every one of them gets a chapter.

```go
// At rest: a key from Alice's password encrypts her vault.
func masterKeyFromPassword(password string, salt []byte, iterations int) ([]byte, error)
func sealSecret(key []byte, name string, plaintext []byte) ([]byte, error)
func openSecret(key []byte, name string, sealed []byte) ([]byte, error)

// Sharing: encrypt to Bob's public key, so only Bob's private key can open it.
func shareSecret(bobPublic *ecdh.PublicKey, plaintext []byte) (Envelope, error)

// Exporting: sign a bundle so anyone with Alice's public key can check it.
func signBundle(alicePrivate ed25519.PrivateKey, bundle []byte) []byte
```

Every function above uses only Go's standard library. That's a big deal: in many
languages, doing this safely means picking between a dozen third-party packages.
