---
title: Passwords Are Not Keys
quiz:
  - question: A thief copies Keybox's encrypted vault file. Compared with attacking a login page, what's different about guessing Alice's password now?
    options:
      - text: Nothing; the thief still has to go through the server's rate limiting
      - text: The attack is offline, so there's no rate limit, no lockout and no logging; only the cost of each guess slows them down
        correct: true
      - text: The thief can decrypt the file without guessing anything
      - text: Offline guessing is impossible for encrypted files
    explanation: |
      Online attacks can be throttled. Offline, the attacker runs guesses on their own
      hardware as fast as it can go. The only defence is making every guess
      expensive, which is exactly what a slow, salted KDF does.
  - question: Roughly how many bits of entropy does a 12-character password have if each character is chosen *uniformly at random* from 62 letters and digits?
    options:
      - text: '12'
      - text: About 71
        correct: true
      - text: '256'
      - text: '744'
    explanation: |
      Each character contributes log2(62), about 5.95 bits, so 12 characters give
      about 71 bits. A human-chosen password of the same length has far less, because
      people don't pick uniformly at random.
---

Keybox encrypts Alice's vault with a key derived from her password. That sounds like
"use the password as the key", but passwords and keys are very different things.

## Recap: password hashing on the server

In [Password Hashing](/courses/learn-http-servers/authentication/password-hashing) you
stored a *verifier*: a salted, deliberately slow PBKDF2 hash, so that a leaked user
table doesn't hand out passwords. Keybox needs the same ingredients for a different
reason. It doesn't store anything to compare against; it **derives an encryption key**
from the password, and only the right password yields a key that decrypts the vault.

Both jobs use the same tool, a **password-based key derivation function** (PBKDF),
because both face the same enemy: offline guessing.

## Entropy

A key is 256 bits of uniform randomness. A password is whatever a human came up with.
Security is measured by **entropy**: the number of bits of unpredictability, or
roughly log2 of the number of equally likely possibilities.

```go
package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Printf("random 12 chars of [a-zA-Z0-9]: %.0f bits\n", 12*math.Log2(62))
	fmt.Printf("random 6-word passphrase (7776-word list): %.0f bits\n", 6*math.Log2(7776))
	fmt.Printf("AES-256 key: %d bits\n", 256)
}
```

```
random 12 chars of [a-zA-Z0-9]: 71 bits
random 6-word passphrase (7776-word list): 78 bits
AES-256 key: 256 bits
```

Those are the *best* cases, for passwords picked randomly. Real, human-chosen passwords
are much weaker: `Summer2026!` technically has 11 characters from a big alphabet, but
it's in every cracking dictionary, and attackers try variations of common patterns
first. Studies of leaked databases find typical human passwords are worth perhaps 20
to 40 bits against a good cracker.

## Offline guessing

The thief with a stolen vault backup (from Keybox's threat model) can test a guess by
deriving a key and trying to decrypt. No server, no rate limit, no lockout. With a fast
hash like SHA-256, a single modern GPU tries around ten billion guesses per second,
2^33 or so. A 40-bit password falls in a couple of minutes.

You can't make the password stronger (though you can encourage passphrases and
password managers). What you *can* do is make each guess expensive:

- **Slow.** A KDF that takes about 100 ms per guess costs the attacker a billion times
  more than a single fast hash, and a legitimate user doesn't notice a tenth of a second
  at unlock time.
- **Salted.** A random salt per vault means the attacker must attack each vault
  separately, and precomputed tables are useless.
- **Memory-hard**, ideally. GPUs and custom chips have huge parallel compute but
  limited memory per core. A KDF that needs, say, 64 MiB per guess takes away most of
  their advantage.

The next lessons build Keybox's master key with PBKDF2, then look at the memory-hard
alternatives.

## Password strength checks

Make Keybox's password *checks* useful rather than annoying: require a decent minimum
length (12 or more for a vault password), allow long passphrases, reject passwords
found in breach lists, and skip composition rules like "must contain a symbol", which
mainly produce `Password1!`.
