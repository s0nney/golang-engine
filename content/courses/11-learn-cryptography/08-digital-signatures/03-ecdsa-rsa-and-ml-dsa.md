---
title: ECDSA, RSA and ML-DSA
quiz:
  - question: In 2010, researchers extracted Sony's PlayStation 3 code-signing key. What had Sony done wrong?
    options:
      - text: Used a key that was too short
      - text: Used the same ECDSA per-signature nonce for every signature, which let anyone solve for the private key from two signatures
        correct: true
      - text: Published the private key by accident
      - text: Used Ed25519 instead of ECDSA
    explanation: |
      ECDSA needs a fresh secret nonce for every signature. Sony used a constant, so two
      signatures gave two equations in two unknowns. Ed25519's deterministic nonces,
      and Go's hedged ECDSA nonces, exist to make this impossible.
  - question: Which signature scheme would you pick for a *new* protocol where every peer runs recent Go and you want to prepare for quantum computers?
    options:
      - text: RSA-1024
      - text: ECDSA with P-256 alone
      - text: ML-DSA, possibly alongside Ed25519 in a hybrid
        correct: true
      - text: HMAC-SHA256
    explanation: |
      ML-DSA (FIPS 204) is the standardized post-quantum signature, in Go since 1.27.
      Signatures are large (3,309 bytes for ML-DSA-65), and many deployments pair it
      with Ed25519 during the transition. RSA-1024 is too weak, and HMAC isn't a
      signature.
---

Ed25519 is Keybox's choice, but you'll meet three other signature families in
certificates, protocols and file formats. Here's what to know about each.

## ECDSA

**ECDSA** (Elliptic Curve Digital Signature Algorithm) with the NIST curve **P-256** is
the most common signature in TLS certificates, and it's FIPS-approved.

```go
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

func main() {
	msg := []byte("keybox v1.4.0 SHA256SUMS")

	ec, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	digest := sha256.Sum256(msg)
	sig, _ := ecdsa.SignASN1(rand.Reader, ec, digest[:])
	sig2, _ := ecdsa.SignASN1(rand.Reader, ec, digest[:])
	fmt.Println("ECDSA P-256 valid:", ecdsa.VerifyASN1(&ec.PublicKey, digest[:], sig), "randomized:", string(sig) != string(sig2))

	ml, _ := mldsa.GenerateKey(mldsa.MLDSA65())
	msig, err := ml.Sign(nil, msg, &mldsa.Options{})
	if err != nil {
		panic(err)
	}
	fmt.Println("ML-DSA-65 valid:", mldsa.Verify(ml.PublicKey(), msg, msig, nil) == nil,
		"public key:", len(ml.PublicKey().Bytes()), "signature:", len(msig))
}
```

```
ECDSA P-256 valid: true randomized: true
ML-DSA-65 valid: true public key: 1952 signature: 3309
```

Differences from Ed25519 to keep in mind:

- **You hash first.** `SignASN1` takes a *digest*, not the message. Pick the hash to
  match the curve (SHA-256 for P-256).
- **The nonce is the danger.** Every ECDSA signature needs a secret random nonce `k`.
  Reuse `k` for two messages and the private key can be computed from the two
  signatures; that's how Sony's PS3 signing key fell in 2010. Even a few *biased* bits
  of `k` across many signatures can leak the key. Go derives `k` from the private key,
  the message **and** fresh randomness ("hedged" nonces), so neither a bad RNG nor a
  repeated message alone can cause reuse. It still means signatures are randomized, as
  the output shows.
- **Encoding.** `SignASN1` returns a variable-length ASN.1 DER signature (70 to 72
  bytes for P-256). Some protocols (JWS, WebAuthn in places) use a fixed 64-byte `r||s`
  form instead; convert carefully.
- **Malleability.** For any valid ECDSA signature `(r, s)`, `(r, -s)` is also valid.
  Don't use signature bytes as unique IDs.

## RSA

RSA (`crypto/rsa`) is the oldest scheme still in wide use: many root CAs, older
certificates, and JWT's RS256. If you must use it:

- keys of at least **2048 bits** (3072 for long-term); Go refuses to generate keys
  under 1024 bits.
- **RSA-PSS** (`rsa.SignPSS`) for new signatures; PKCS #1 v1.5 (`rsa.SignPKCS1v15`)
  only for compatibility.
- RSA *encryption* is a different, even more error-prone matter; in Go 1.26 the
  PKCS #1 v1.5 encryption functions were deprecated. Use HPKE instead.

RSA keys and signatures are large (256 bytes at 2048 bits) and key generation is slow,
so there's little reason to choose it for anything new.

## ML-DSA: post-quantum signatures

**ML-DSA** (FIPS 204, formerly CRYSTALS-Dilithium) is the standardized post-quantum
signature scheme, and Go 1.27 added it as `crypto/mldsa`, with support in `crypto/x509`
and `crypto/tls` too.

- Three parameter sets: `MLDSA44()`, `MLDSA65()` and `MLDSA87()`. ML-DSA-65 is a common
  default.
- It's big: a 1,952-byte public key and a 3,309-byte signature for ML-DSA-65, compared
  with 32 and 64 bytes for Ed25519.
- `mldsa.Options{Context: "..."}` gives built-in domain separation, and
  `SignDeterministic` exists when you need reproducible output.

As with KEMs, early deployments often use **hybrid** signatures: sign with both Ed25519
and ML-DSA and require both to verify.

## Which one?

| Situation | Choose |
| --- | --- |
| Default for new designs | Ed25519 |
| FIPS or WebPKI certificate requirements | ECDSA P-256 |
| Interop with legacy systems | RSA-PSS, 2048+ bits |
| Long-lived signatures that must survive quantum computers | ML-DSA (with Ed25519 as a hybrid) |
