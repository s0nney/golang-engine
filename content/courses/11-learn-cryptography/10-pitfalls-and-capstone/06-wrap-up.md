---
title: Wrap-Up
quiz:
  - question: You need to encrypt data to a colleague's public key in a new Go service. What's the best first choice from this course?
    options:
      - text: Hand-rolled ECDH + HKDF + AES-GCM, since you've now built it
      - text: '`crypto/hpke` with a standard ciphersuite, plus a signature if the recipient must know who sent it'
        correct: true
      - text: RSA PKCS #1 v1.5 encryption
      - text: AES-GCM with a key derived from the colleague's email address
    explanation: |
      Building the hybrid scheme yourself taught you how it works; in production, the
      standardized, analysed construction wins. HPKE gives confidentiality to a public
      key, and a signature adds sender authenticity.
  - question: Which habit from this course applies to *every* primitive you used?
    options:
      - text: Always use the largest key size available
      - text: Bind context into what you hash, MAC, encrypt or sign, with an unambiguous encoding and a versioned label
        correct: true
      - text: Always compare with `==`
      - text: Avoid the standard library in favour of third-party packages
    explanation: |
      Labels and unambiguous encodings showed up in hashing, MACs, KDF info strings,
      associated data, HPKE info and signature contexts. Most real-world breaks of
      sound primitives are context confusions, and this habit prevents them.
---

That's Keybox done: an end-to-end-encrypted vault that derives keys from passwords,
encrypts at rest with authenticated encryption, shares secrets with post-quantum
public-key encryption, and signs what it exports. Every line of it used Go's standard
library.

## What you learned

- **Goals and threat models.** Confidentiality, integrity and authenticity are separate
  promises, and a design only means something relative to the attacker it's for.
- **Randomness.** `crypto/rand` for everything secret; rejection sampling for ranges;
  `testing/cryptotest` to make tests reproducible.
- **Hashing.** SHA-2 and SHA-3, streaming with `io.Copy`, and unambiguous encodings.
- **MACs.** HMAC, RFC test vectors, constant-time comparison, and verifiers without
  gaps.
- **Key derivation.** PBKDF2 for passwords (and why Argon2id is better when you can use
  it), salts and peppers, HKDF for turning one key into many.
- **Symmetric encryption.** Why ECB and unauthenticated modes fail, AES-GCM, nonce
  management with `NewGCMWithRandomNonce`, associated data and key wrapping.
- **Public-key cryptography.** X25519, hybrid encryption, ML-KEM and HPKE.
- **Signatures.** Ed25519, ECDSA, RSA and ML-DSA; signing exact bytes with context;
  signed releases and transparency logs.
- **Certificates and TLS.** Running a CA, verifying chains, `crypto/tls` defaults and
  mutual TLS.
- **Operations.** A review checklist of real mistakes, crypto agility, key rotation,
  and keeping secrets out of logs.

Here's the whole of Keybox's cryptography, as the calls you made:

```go
masterKey, _ := pbkdf2.Key(sha256.New, password, salt, 600_000, 32) // password -> key
vaultKey, _ := hkdf.Key(sha256.New, masterKey, accountKey, "keybox v1 vault encryption", 32)
gcm, _ := cipher.NewGCMWithRandomNonce(block)                       // at rest
sealed := gcm.Seal(nil, nil, secret, entryAD(vaultID, name))
share, _ := hpke.Seal(bobKey, hpke.HKDFSHA256(), hpke.AES256GCM(), info, secret) // sharing
sig := ed25519.Sign(aliceSigning, append([]byte(bundleContext), payload...))      // authenticity
```

Six lines, each standing on a chapter's worth of reasons.

## Where to go next with cryptography

1. **Read the Go source.** `crypto/internal/fips140` holds the implementations, written
   for readability and constant-time behaviour. `crypto/hpke` is a compact, readable
   protocol implementation.
2. **Read *Real-World Cryptography*** by David Wong, or Dan Boneh and Victor Shoup's free
   *A Graduate Course in Applied Cryptography* if you want the proofs.
3. **Study a protocol end to end.** RFC 8446 (TLS 1.3), RFC 9180 (HPKE) or the Signal
   protocol documentation, now that you know every building block they use.
4. **Try breaking things, legally.** The Cryptopals challenges walk you through attacks
   on the very mistakes in this course's checklist.
5. **Harden a real project.** Add Argon2id from `golang.org/x/crypto`, a proper key
   hierarchy, and key rotation to something you run.

## The end of the roadmap

This is the last course on the roadmap. You started with variables and `fmt.Println`;
since then you've learned object-oriented and functional design, algorithms and data
structures, concurrency, testing, both sides of HTTP, the depths of
[Go's type system](/courses/learn-advanced-types), and now how to protect data with
cryptography. That's the toolkit of a working Go engineer.

From here, what you build is up to you. Pick a project you care about, ship it, keep it
running, and come back to any course when a topic becomes real. It'll make more sense
the second time, now that you've built things with it.

Thanks for learning with goland-engine. Go build something, and keep your nonces unique.
