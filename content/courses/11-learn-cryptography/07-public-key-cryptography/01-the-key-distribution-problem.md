---
title: The Key Distribution Problem
quiz:
  - question: Alice wants to share a secret with Bob through the untrusted Keybox server. Which key does she encrypt to?
    options:
      - text: Her own private key
      - text: Bob's public key, which the server can relay freely because it only allows encrypting to Bob
        correct: true
      - text: Bob's private key
      - text: The server's public key
    explanation: |
      Anyone may encrypt to Bob's public key; only Bob's private key can decrypt. The
      server can store and forward both the public key and the ciphertext without
      learning anything. (It could still lie about *which* public key is Bob's, a
      problem this chapter ends on.)
  - question: Why do real systems use "hybrid" encryption instead of encrypting the whole message with public-key operations?
    options:
      - text: Public-key operations are slow and only handle small inputs, so they're used to agree on or transport a symmetric key, and AES-GCM encrypts the data
        correct: true
      - text: Public-key encryption is less secure than AES
      - text: Hybrid encryption is required by law
      - text: Public keys can't encrypt anything
    explanation: |
      Asymmetric operations cost far more than AES and are designed around small
      values such as keys. So the public-key part establishes a fresh symmetric key,
      and a fast AEAD does the bulk work. TLS, HPKE, PGP and Signal all work this way.
---

Everything so far has been **symmetric**: the same key encrypts and decrypts, or
creates and checks a MAC. That leaves a chicken-and-egg problem: Alice and Bob need a
shared secret key before they can communicate securely, but how do they agree on one
over a channel Eve is listening to?

For Keybox this is the sharing feature. Alice wants to give Bob her team's database
password. They've never met, and every byte between them goes through Keybox's
server, which the threat model says can't be trusted with plaintext.

## Key pairs

**Public-key** (asymmetric) cryptography splits the key in two:

- a **private key**, kept secret by its owner, and
- a **public key**, derived from it and safe to publish anywhere.

It's computationally infeasible to recover the private key from the public one. What
you can do with the pair depends on the scheme:

| Scheme | Public key lets anyone... | Private key lets the owner... | Chapter |
| --- | --- | --- | --- |
| Key exchange (X25519) | combine it with their own private key to get a shared secret | compute the same shared secret | this one |
| KEM (ML-KEM) | *encapsulate*: produce a fresh shared secret plus a ciphertext | *decapsulate* the ciphertext to get the secret | this one |
| Signatures (Ed25519) | verify signatures | sign | 8 |

In 1976, Whitfield Diffie and Martin Hellman published the first practical public key
exchange, a result that made secure communication between strangers possible and
underlies everything from HTTPS to messaging apps.

## Keybox's sharing, in outline

1. When Bob creates his account, his device generates an X25519 key pair. The private
   key stays on his device (wrapped under his vault key, as in chapter 6). The public
   key is uploaded to the Keybox directory.
2. Alice's client fetches Bob's public key and uses it to derive a fresh symmetric key
   that only Bob can also derive.
3. Alice seals the secret with AES-GCM under that key and uploads the result.
4. Bob downloads it and derives the same key with his private key.

That's **hybrid encryption**: public-key crypto to establish a key, symmetric crypto
for the data.

## Elliptic curves

Diffie and Hellman's original scheme uses arithmetic modulo a large prime and needs
2048-bit or larger numbers. Modern systems use **elliptic curves**, which give the
same security with much smaller keys: an X25519 public key is **32 bytes**, and so is
the private key. Go's `crypto/ecdh` supports X25519 and the NIST curves P-256, P-384 and
P-521. X25519, designed by Daniel Bernstein, is the usual first choice: fast, easy to
implement safely, and supported everywhere. P-256 is the choice when FIPS compliance
matters.

```go
priv, err := ecdh.X25519().GenerateKey(rand.Reader) // rand is ignored since Go 1.26
pub := priv.PublicKey()
fmt.Println(len(priv.Bytes()), len(pub.Bytes())) // 32 32
```

## The catch: who owns this key?

Public keys solve the eavesdropper, not the impostor. If Mallory controls the server,
she can hand Alice *her own* public key and claim it's Bob's. Alice encrypts to Mallory,
who decrypts, reads, and re-encrypts to Bob: a **man-in-the-middle** attack.

Public-key crypto always needs a way to **authenticate public keys**:

- comparing **fingerprints** (hashes of the public key) in person or over a trusted
  channel, like Signal's safety numbers,
- **signatures** from someone you already trust, which is chapter 8,
- **certificates** from an authority, which is chapter 9.

Keep that in mind through this chapter: the code here keeps Eve out, and chapters 8 and
9 deal with Mallory.

## Quantum computers

A large enough quantum computer running Shor's algorithm would break X25519, the NIST
curves and RSA. None exists yet, but ciphertext recorded today could be decrypted
later ("harvest now, decrypt later"). This chapter ends with the **post-quantum** KEM
ML-KEM and with HPKE, both in Go's standard library.
