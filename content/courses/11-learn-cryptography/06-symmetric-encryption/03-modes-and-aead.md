---
title: Modes and AEAD
quiz:
  - question: What does "authenticated encryption" add over plain CTR or CBC mode?
    options:
      - text: It makes the ciphertext shorter
      - text: A tag that `Open` checks, so any modification of the ciphertext (or the associated data) makes decryption fail instead of returning altered plaintext
        correct: true
      - text: It proves who encrypted the data, like a signature
      - text: It removes the need for a nonce
    explanation: |
      AEAD = confidentiality + integrity in one primitive. It doesn't prove *which* key
      holder encrypted something (both sides share the key), and it still needs a unique
      nonce per message.
  - question: When is ChaCha20-Poly1305 a better choice than AES-GCM?
    options:
      - text: Always; AES is broken
      - text: On hardware without AES instructions (some older phones and small devices), where ChaCha20 is faster in software and naturally constant-time
        correct: true
      - text: When you don't want to use a nonce
      - text: When you need non-repudiation
    explanation: |
      Both are excellent AEADs. AES-GCM is fastest where the CPU accelerates AES and
      carry-less multiplication; ChaCha20-Poly1305 shines without that hardware. TLS
      1.3 supports both, and Go picks based on the hardware.
---

A mode of operation turns a block cipher into something that can encrypt a message of
any length. There are many modes; you need to know why two popular ones aren't enough,
and why modern code uses an **AEAD**.

## CTR: a stream cipher from a block cipher

**Counter mode** encrypts a sequence of counter blocks, `nonce||0`, `nonce||1`,
`nonce||2`..., to produce a **keystream**, and XORs it with the plaintext. It's the
one-time pad idea with the pad generated from a short key. CTR is fast, parallel and
needs no padding.

It also inherits the one-time pad's two weaknesses:

- **Malleable.** Flip a ciphertext bit and the same plaintext bit flips (chapter 1).
- **Nonce reuse is fatal.** Reuse a nonce with the same key and you reuse the keystream:
  the two-time pad. XORing the ciphertexts gives the XOR of the plaintexts.

## CBC: chaining blocks

**Cipher Block Chaining** XORs each plaintext block with the previous ciphertext block
before encrypting it, starting from a random IV. It hides repeated blocks, unlike ECB,
but it needs **padding** to a multiple of 16 bytes, and it's malleable in its own way.

Worse, CBC gave us the **padding oracle**: if a server reveals, by error message or by
timing, whether a tampered ciphertext had valid padding, an attacker can decrypt data
without the key. Variants of this broke TLS (Lucky Thirteen, POODLE), ASP.NET and many
custom protocols. The root cause is always the same: **the receiver processed data
before checking it was authentic.**

Go still provides `cipher.NewCTR` and `cipher.NewCBCEncrypter` for compatibility with
existing formats. Don't reach for them in new designs.

## AEAD: authenticated encryption with associated data

An **AEAD** encrypts *and* authenticates in one operation, and decryption is
all-or-nothing: either you get the exact original plaintext, or you get an error and no
plaintext at all. Go's interface is `cipher.AEAD`:

```go
type AEAD interface {
	NonceSize() int // bytes of nonce Seal and Open expect
	Overhead() int  // max bytes the ciphertext is longer than the plaintext

	// Seal encrypts and authenticates plaintext, authenticates additionalData,
	// and appends the result to dst.
	Seal(dst, nonce, plaintext, additionalData []byte) []byte

	// Open authenticates and decrypts. On any mismatch it returns an error.
	Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error)
}
```

The "associated data" is extra context that's **authenticated but not encrypted**:
Keybox will use the secret's name, so a ciphertext can't be moved to a different entry.
That's lesson 6.

Two AEADs matter in practice:

- **AES-GCM**: AES in counter mode plus a polynomial MAC called GHASH (the "Galois"
  part). Standardized by NIST, FIPS-approved, hardware-accelerated on nearly every
  server and laptop CPU. In Go: `cipher.NewGCM`, or `cipher.NewGCMWithRandomNonce`.
- **ChaCha20-Poly1305** (RFC 8439): the ChaCha20 stream cipher plus the Poly1305 MAC.
  Fast in pure software and constant-time without special instructions. In Go it's in
  `golang.org/x/crypto/chacha20poly1305`, which also offers **XChaCha20-Poly1305** with
  a 24-byte nonce that's safe to choose at random for practically unlimited messages.
  (The standard library uses ChaCha20-Poly1305 internally in TLS and exposes it through
  `crypto/hpke`, but has no general-purpose package for it.)

Both return a `cipher.AEAD`, so code written against the interface can switch between
them.

## The rules for any AEAD

1. **Never reuse a nonce with the same key.** For GCM, the consequences are worse than
   for CTR: reuse leaks the authentication key too, allowing forgeries (lesson 5).
2. **Treat `Open` errors as final.** Don't retry with other settings, don't log the
   "partial" plaintext (there is none), and return one generic error.
3. **Keep messages to a sensible size.** GCM can encrypt up to about 64 GiB per
   message. For huge files, split them into chunks, each sealed separately with its
   position in the associated data, so chunks can't be reordered.
