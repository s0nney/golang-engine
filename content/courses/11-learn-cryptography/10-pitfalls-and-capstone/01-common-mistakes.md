---
title: Common Mistakes
quiz:
  - question: A code review finds `key := sha256.Sum256([]byte(password))` used as an AES-GCM key. What's the main problem?
    options:
      - text: SHA-256 output is the wrong length for AES
      - text: A fast, unsalted hash lets anyone who steals ciphertext test billions of password guesses per second; passwords need a slow, salted KDF
        correct: true
      - text: AES-GCM can't use hashed keys
      - text: Nothing; hashing the password makes it a proper key
    explanation: |
      32 bytes is a fine AES-256 key length, which is why this bug survives review. The
      key only has as much entropy as the password, and SHA-256 makes guessing cheap.
      Use PBKDF2 or Argon2id with a random salt, as in chapter 5.
  - question: Which of these is *not* a reason to avoid designing your own protocol from primitives?
    options:
      - text: Primitives are easy to combine in ways that are individually fine but jointly broken
      - text: Standard constructions come with security proofs, test vectors and years of analysis
      - text: The standard library forbids calling low-level primitives directly
        correct: true
      - text: Other implementations can interoperate with a standard
    explanation: |
      Go happily lets you call `aes.NewCipher` and `ecdh` directly; you did throughout
      this course. The reasons to prefer HPKE, TLS and friends are proofs, review and
      interoperability, not prohibition.
  - question: 'Keybox''s error handler logs `slog.Error("decrypt failed", "key", hex.EncodeToString(key), "err", err)`. Why is this serious?'
    options:
      - text: It's too verbose
      - text: Logs are copied to many systems and people; the key is now exposed far beyond the process that needed it
        correct: true
      - text: slog can't log hex strings
      - text: It isn't serious if the logs are on the same server
    explanation: |
      Logs flow to aggregators, backups, support tickets and third-party services. A
      key in a log line should be treated as leaked: rotate it. Log key *IDs*, never
      key material.
---

You've now used every major primitive in Go's standard library. Here's the catalog of
ways real systems still get them wrong, with pointers back to where each was covered.
Use it as a code-review checklist.

## Randomness

- **`math/rand` for secrets**, or seeding anything from the clock. Keys, nonces, salts,
  tokens: `crypto/rand` only. ([Predictable Randomness](/courses/learn-cryptography/randomness/predictable-randomness))
- **Modulo bias** when mapping random bytes onto a range.
  ([Modulo Bias](/courses/learn-cryptography/randomness/modulo-bias))

## Hashing and MACs

- **MD5 or SHA-1** anywhere collision resistance matters.
- **Ambiguous encodings** fed to a hash, MAC or signature: concatenating fields without
  lengths. ([Hashing Structured Data](/courses/learn-cryptography/hashing/hashing-structured-data))
- **`hash(key || message)` as a MAC.** Use HMAC.
- **`==` or `bytes.Equal` on MACs and tokens.** Use `hmac.Equal` or
  `subtle.ConstantTimeCompare`. ([Timing Attacks](/courses/learn-cryptography/message-authentication/timing-attacks))
- **Verifiers that check too little:** truncated tags, missing timestamps, fields
  outside the MAC. ([Verifying MACs Safely](/courses/learn-cryptography/message-authentication/verifying-macs-safely))

## Passwords and keys

- **Fast or unsalted password hashing**, or a password used directly as a key.
- **One key for several purposes.** Derive per-purpose keys with HKDF and labels.
- **Raw ECDH output as a key.** Run it through HKDF.
- **Hard-coded keys** in source code or container images. Load them from a secrets
  manager, and scan repositories for leaks.

## Encryption

- **Unauthenticated encryption:** ECB, CBC or CTR without a MAC. Use an AEAD.
  ([Modes and AEAD](/courses/learn-cryptography/symmetric-encryption/modes-and-aead))
- **Nonce reuse**, often through counters that reset or are shared. Prefer
  `cipher.NewGCMWithRandomNonce` and per-context keys.
  ([Nonce Reuse](/courses/learn-cryptography/symmetric-encryption/nonce-reuse))
- **No associated data**, so ciphertexts can be swapped between records.
- **Detailed decryption errors**, or "trying another way" after `Open` fails.
- **Panicking on short input** because a slice was taken before checking its length.

## Public keys, signatures and TLS

- **Unauthenticated public keys:** encrypting to whatever key the server hands you.
- **Parsing before verifying**, or verifying re-encoded data instead of the bytes
  received. ([Signing What You Mean](/courses/learn-cryptography/digital-signatures/signing-what-you-mean))
- **Signatures without context, audience or freshness.**
- **`InsecureSkipVerify: true`**, custom verification that only checks names, or peer
  certificates added to `RootCAs`.
  ([Verifying Chains](/courses/learn-cryptography/certificates-and-tls/verifying-chains))
- **Over-configured TLS** that pins yesterday's cipher suites and curves.

## Homemade protocols

Most of the list above comes from one root cause: combining sound primitives into a
new protocol. Each step looks reasonable; the combination leaks. Before building,
look for a standard that already does the job:

| Need | Use |
| --- | --- |
| Secure channel | TLS 1.3 (`crypto/tls`), mTLS for machines |
| Encrypt to a public key | HPKE (`crypto/hpke`) |
| Encrypt data at rest | AES-GCM with managed nonces, keys from HKDF |
| Tokens | A well-reviewed format, verified strictly |
| Password storage | Argon2id, or PBKDF2 when FIPS or stdlib-only applies |

## Secrets in the wrong places

Finally, the bug that isn't cryptographic at all: **secret material leaking out of the
program.** Keys and plaintext end up in logs, panic messages, error strings, crash
dumps, metrics labels, URLs (and therefore proxy logs and browser history) and
`%+v`-printed structs. You met this from the client side in
[Secrets in Logs](/courses/learn-http-clients/https-and-security/secrets-in-logs);
lesson 3 builds a type that makes it hard to do by accident.

The next exercise hides several of these bugs in one function.
