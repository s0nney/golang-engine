---
title: Why Hashes Aren't MACs
quiz:
  - question: Which of these tags is the right way to authenticate a message with a secret key?
    options:
      - text: '`sha256(key || message)`'
      - text: '`sha512(key || message)`'
      - text: '`hmac.New(sha256.New, key)` over the message'
        correct: true
      - text: '`sha256(message)`, sent over HTTPS'
    explanation: |
      SHA-256 and SHA-512 use the Merkle-Damgård construction, and their digest *is*
      their final internal state, so `hash(key || message)` is vulnerable to length
      extension. A plain hash has no key at all. HMAC is designed and proven for exactly
      this job.
  - question: What does a MAC give you that a plain SHA-256 checksum doesn't?
    options:
      - text: Confidentiality of the message
      - text: Assurance that the message was produced by someone holding the secret key and hasn't been modified since
        correct: true
      - text: Non-repudiation, so the sender can't deny sending it
      - text: A shorter output
    explanation: |
      Anyone can recompute a checksum after changing the data. Only a key holder can
      compute a valid MAC. Because *both* sides hold the key, though, a MAC can't prove
      to a third party which of them made it; that's what signatures are for.
---

Chapter 3 ended with a warning: a hash alone can't prove who made something, because
anyone can compute one. What you need is a **message authentication code** (MAC): a
tag that only someone holding a secret key can produce.

## What a MAC promises

A MAC takes a key and a message and returns a short tag:

```
tag = MAC(key, message)
```

The receiver, who shares the key, recomputes the tag and compares. If they match, the
message came from a key holder and wasn't changed. The security goal is
**unforgeability**: even after seeing tags for as many messages as she likes (even
messages she chose), Mallory can't produce a valid tag for any *new* message.

Keybox's sync server and client share a key and MAC every request, so the server can
reject anything Mallory altered in transit.

## The tempting shortcut

"Just hash the key together with the message":

```go
// DON'T: vulnerable to length extension.
func naiveMAC(key, msg []byte) []byte {
	h := sha256.New()
	h.Write(key)
	h.Write(msg)
	return h.Sum(nil)
}
```

It looks fine. Without the key, you can't compute the hash of `key || msg`. Yet this
construction is broken for SHA-256, SHA-512 and SHA-1.

## Length extension, conceptually

SHA-2 uses the **Merkle-Damgård** construction: it pads the input, splits it into
blocks, and runs a compression function that updates an internal state one block at a
time. The final digest **is** that internal state, written out.

That means a published digest of `key || msg` hands out the hash's full internal state
at the end of the (padded) message. Someone who knows only the *length* of the key can
pick up where the hash left off and compute a valid digest for
`key || msg || padding || more`, a longer message that starts with the original, without
ever learning the key. In 2009, Flickr's API signed requests as `md5(secret || params)`
and researchers showed exactly this kind of forgery against it.

You don't need to know how to carry out the attack to avoid it. Remember the rule:
**never build a MAC by hashing a key and message together yourself.**

Hashes that don't expose their full state are immune: SHA-384, SHA-512/256 and SHA-3
(which is why SHA-3's designers say a keyed SHA-3 hash *is* a fine MAC, called KMAC).
But you don't need to remember which hashes are which, because there's a standard
construction that's safe with all of them.

## HMAC

**HMAC** (RFC 2104, 1997) wraps any hash function in two nested passes, with the key
mixed in at both ends:

```
HMAC(K, m) = H( (K ^ opad) || H( (K ^ ipad) || m ) )
```

`ipad` and `opad` are fixed constants (bytes `0x36` and `0x5c` repeated). The inner hash
processes the message; the outer hash hides the inner hash's state behind a second
keyed hash, so there's nothing an attacker can extend. HMAC comes with a security proof
that relies only on mild properties of the underlying hash, which is why HMAC-MD5 and
HMAC-SHA1 remained unbroken long after MD5 and SHA-1 collisions were found (don't use
them anyway).

HMAC is everywhere: JWT's `HS256`, TLS 1.2's record MACs, AWS request signing, webhook
signatures, and inside HKDF and PBKDF2, which you'll meet in chapter 5.

## Other MACs

- **KMAC** (SHA-3 based), mentioned above.
- **Poly1305** and **GMAC** are fast one-time MACs built into the authenticated
  encryption modes ChaCha20-Poly1305 and AES-GCM. You'll use them indirectly in
  chapter 6, and never on their own.

In Go, the standard MAC is `crypto/hmac`, and it's next.
