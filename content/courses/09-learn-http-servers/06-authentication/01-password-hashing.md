---
title: Password Hashing
quiz:
  - question: Why is a plain `sha256.Sum256([]byte(password))` a bad way to store passwords?
    options:
      - text: SHA-256 can be reversed to recover the password
      - text: It's far too fast; attackers with a leaked database can try billions of guesses per second, and identical passwords get identical hashes
        correct: true
      - text: SHA-256 produces hashes that are too long to store
      - text: It isn't in the standard library
    explanation: |
      Cryptographic hashes like SHA-256 are designed to be *fast*, which is exactly wrong for
      passwords. Password hashing functions are deliberately slow and use a random salt, so
      each guess is expensive and the same password gives a different hash for every user.
  - question: What is the salt for?
    options:
      - text: It encrypts the hash so only the server can read it
      - text: It's a random value stored next to the hash so identical passwords hash differently and precomputed tables are useless
        correct: true
      - text: It's a secret key that must never be stored
      - text: It makes the hash shorter
    explanation: |
      The salt isn't secret, and it's stored with the hash. Its job is to make every hash
      unique, so an attacker can't crack all the `cheese123` users at once or use a
      precomputed "rainbow table".
---

Squeak is about to get user accounts, which means it's about to hold passwords. This is
the part of the course where mistakes end up in the news, so let's be careful.

## Never store the password

Rule one: the server never stores a password, not even encrypted. It stores a **hash**:
the output of a one-way function. At login you hash what the user typed and compare.
If the database leaks, the attacker gets hashes, not passwords.

But not just any hash. SHA-256 is designed to be *fast*, and a GPU can compute
billions of them per second. An attacker with your leaked table just hashes every word
in a dictionary and looks for matches. Password hashing needs two extra ingredients:

1. **A salt.** A random value, different for every user, mixed into the hash and stored
   next to it. Now two users with the password `cheese123` get different hashes, and
   precomputed tables are useless.
2. **A work factor.** The function is deliberately slow (tens of milliseconds), so each
   guess costs the attacker the same. A tunable iteration count lets you make it slower
   as hardware gets faster.

## bcrypt, argon2 and PBKDF2

The usual choices are:

- **argon2id**, the modern recommendation, which is memory-hard so GPUs don't help much.
- **bcrypt**, older and still perfectly respectable. The most common choice in Go.
- **scrypt**, also memory-hard.
- **PBKDF2**, the oldest, which isn't memory-hard but is standardised and FIPS-approved.

In a real project you'd most likely use `golang.org/x/crypto/bcrypt` or
`golang.org/x/crypto/argon2`. Those live in the `x/crypto` module, not the standard
library, and this course's exercises can only use the standard library. Since Go 1.24,
though, the standard library has **`crypto/pbkdf2`**, and with enough iterations it's
an acceptable password hash. (OWASP currently recommends 600,000 iterations of
HMAC-SHA256.) Squeak uses it.

```go
func Key[Hash hash.Hash](h func() Hash, password string, salt []byte, iter, keyLength int) ([]byte, error)
```

## Hashing and checking

```go
package main

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const iterations = 600_000

var b64 = base64.RawStdEncoding

// hashPassword returns a self-describing string: algorithm$iterations$salt$hash.
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	rand.Read(salt) // crypto/rand never fails on supported platforms
	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", iterations, b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

func checkPassword(password, stored string) (bool, error) {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false, errors.New("unknown hash format")
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, err
	}
	salt, err := b64.DecodeString(parts[2])
	if err != nil {
		return false, err
	}
	want, err := b64.DecodeString(parts[3])
	if err != nil {
		return false, err
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, iter, len(want))
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

func main() {
	a, _ := hashPassword("cheese123")
	b, _ := hashPassword("cheese123")
	fmt.Println("same password, same hash?", a == b)
	fmt.Println("prefix:", a[:21])

	ok, _ := checkPassword("cheese123", a)
	fmt.Println("right password:", ok)
	ok, _ = checkPassword("cheese124", a)
	fmt.Println("wrong password:", ok)
}
```

```
same password, same hash? false
prefix: pbkdf2-sha256$600000$
right password: true
wrong password: false
```

The details that matter:

- **The stored string describes itself.** Algorithm, iterations and salt travel with
  the hash, so you can raise the iteration count later: old hashes still verify with
  their old count, and you rehash on the user's next successful login. bcrypt and argon2
  strings (`$2a$10$...`, `$argon2id$v=19$...`) work the same way.
- **`crypto/rand`** makes the salt. Never `math/rand`.
- **`subtle.ConstantTimeCompare`** compares without stopping at the first different byte,
  so response timing doesn't reveal how close a guess was. (`hmac.Equal` does the same.)

## Cost and denial of service

At 600,000 iterations, one hash takes tens of milliseconds of CPU. That's the point, but
it also means someone hammering `/api/login` can burn your CPU. Rate-limit login
attempts (you'll build a rate limiter in the testing chapter), and cap password length
(say 72 to 256 bytes) so nobody sends a 10 MB password.

## Password rules

Keep them simple and modern: a minimum length (8 or more), a generous maximum, and no
arbitrary "must contain a symbol" rules. Checking new passwords against lists of breached
passwords helps more than composition rules.
