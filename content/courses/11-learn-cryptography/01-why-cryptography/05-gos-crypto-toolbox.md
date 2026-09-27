---
title: Go's Crypto Toolbox
quiz:
  - question: You need bcrypt or Argon2 for a project. Where do they live?
    options:
      - text: '`crypto/bcrypt` and `crypto/argon2` in the standard library'
      - text: In `golang.org/x/crypto`, maintained by the Go team but outside the standard library
        correct: true
      - text: Nowhere; Go doesn't support them
      - text: In `crypto/pbkdf2`, as options
    explanation: |
      `golang.org/x/crypto` holds algorithms the Go team maintains but hasn't put in the
      standard library, including `bcrypt`, `argon2`, `scrypt` and `chacha20poly1305`.
      You add it to `go.mod` like any other module.
  - question: |
      What does this print when run with `GODEBUG=fips140=only`?

      ```go
      defer func() { fmt.Println("recovered:", recover()) }()
      fmt.Printf("%x\n", md5.Sum([]byte("keybox")))
      ```
    options:
      - text: The MD5 hash, because MD5 is still in the standard library
      - text: 'A panic, recovered as `crypto/md5: use of MD5 is not allowed in FIPS 140-only mode`'
        correct: true
      - text: A compile error
      - text: An empty hash
    explanation: |
      `fips140=on` switches the approved algorithms to the FIPS 140-3 module's
      implementation. `fips140=only` goes further and makes non-approved algorithms
      return errors or panic. MD5 isn't approved, and `md5.Sum` has no error to return,
      so it panics.
---

Go has one of the best standard-library crypto collections of any language. Before
diving in, here's the map, and the philosophy that shaped it.

## The packages you'll use

| Package | What it's for | Chapter |
| --- | --- | --- |
| `crypto/rand` | Secure random bytes, keys, tokens | 2 |
| `crypto/sha256`, `crypto/sha512`, `crypto/sha3` | Hash functions | 3 |
| `crypto/hmac`, `crypto/subtle` | MACs and constant-time comparison | 4 |
| `crypto/pbkdf2`, `crypto/hkdf` | Key derivation from passwords and from keys | 5 |
| `crypto/aes`, `crypto/cipher` | Block ciphers and AES-GCM authenticated encryption | 6 |
| `crypto/ecdh`, `crypto/mlkem`, `crypto/hpke` | Key exchange, post-quantum KEMs, hybrid public-key encryption | 7 |
| `crypto/ed25519`, `crypto/ecdsa`, `crypto/mldsa` | Digital signatures | 8 |
| `crypto/x509`, `crypto/tls` | Certificates and TLS | 9 |

A lot of this is recent. `crypto/hkdf`, `crypto/pbkdf2`, `crypto/sha3` and `crypto/mlkem`
arrived in Go 1.24, `crypto/hpke` in Go 1.26, and `crypto/mldsa` (post-quantum
signatures) in Go 1.27. Older tutorials send you to `golang.org/x/crypto` for things
that are now built in.

That module, `golang.org/x/crypto`, is still where you'll find `bcrypt`, `argon2`,
`scrypt`, `chacha20poly1305` and `ssh`. It's maintained by the Go team and perfectly
reasonable to depend on. This course's exercises only use the standard library, so
those get explained but not run.

You'll also see packages you should **not** use for new designs: `crypto/md5`,
`crypto/sha1`, `crypto/des` and `crypto/rc4` exist for compatibility with old
protocols and files. MD5 and SHA-1 have practical collision attacks, DES has a key
small enough to brute-force, and RC4's keystream is biased.

## The design philosophy

Go's crypto was designed by people who spent years cleaning up after other libraries'
footguns, and it shows:

- **Safe defaults, few knobs.** `tls.Config{}` is secure as it is. `ed25519` has no
  parameters to get wrong. There's no API for AES in ECB mode, the textbook-broken way
  to use a block cipher.
- **Hard-to-misuse APIs.** `cipher.NewGCMWithRandomNonce` (Go 1.24) handles nonces for
  you. `crypto/rand.Read` never returns an error (Go 1.24), so there's no error branch
  to accidentally ignore.
- **Randomness you can't mess up.** Since Go 1.26, most functions that take an
  `io.Reader` for randomness, like `ecdh.X25519().GenerateKey(rand)` or
  `ecdsa.SignASN1(rand, ...)`, ignore it and always use the secure system source. For deterministic tests there's
  `testing/cryptotest.SetGlobalRandom` instead.
- **Constant-time by default.** Implementations avoid branches and memory lookups that
  depend on secret data, so timing doesn't leak keys.

## FIPS 140-3 mode

Some customers (US government agencies, many regulated companies) require crypto that's
validated under **FIPS 140-3**, a US standard for cryptographic modules. Since Go 1.24,
the standard library contains the **Go Cryptographic Module**, and you can switch it on
without cgo or a special toolchain:

- `GODEBUG=fips140=on` runs the approved algorithms through the module and its
  self-tests.
- `GODEBUG=fips140=only` additionally makes non-approved algorithms (like MD5) fail.
- `GOFIPS140=v1.0.0` at **build** time selects a specific frozen, validated version of
  the module instead of the latest source.

`crypto/fips140` reports what's going on at run time:

```go
package main

import (
	"crypto/fips140"
	"fmt"
)

func main() {
	fmt.Println("FIPS 140-3 mode:", fips140.Enabled())
	fmt.Println("enforced:", fips140.Enforced())
	fmt.Println("module version:", fips140.Version())
}
```

```
FIPS 140-3 mode: false
enforced: false
module version: latest
```

Run it with the setting and the answers change:

```
$ GODEBUG=fips140=only go run .
FIPS 140-3 mode: true
enforced: true
module version: latest
```

FIPS mode is a compliance tool, not a security upgrade. Go's normal crypto is already
the same code; FIPS mode mostly adds self-tests and removes non-approved options. Use it
when a contract or regulation requires it, and read
[the FIPS 140-3 documentation](https://go.dev/doc/security/fips140) first.
