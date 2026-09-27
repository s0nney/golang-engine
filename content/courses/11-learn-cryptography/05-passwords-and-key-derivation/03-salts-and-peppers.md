---
title: Salts and Peppers
quiz:
  - question: What's the key difference between a salt and a pepper?
    options:
      - text: A salt is secret and a pepper isn't
      - text: A salt is random per user and stored with the hash; a pepper is one secret key kept *outside* the database, so a database-only leak can't be cracked at all
        correct: true
      - text: They're the same thing with different names
      - text: A pepper replaces the need for a slow KDF
    explanation: |
      Salts defeat precomputation and make each hash a separate target, but they're
      stored right next to the hash. A pepper lives somewhere else (a secrets manager,
      an HSM, the app's config), so an attacker who only has the database can't even
      start guessing. It's an addition to a slow KDF, never a replacement.
  - question: Keybox generates a random 128-bit "account key" on each new device and mixes it with the password when deriving the vault key. What does that buy?
    options:
      - text: Nothing, since the account key is stored on the device
      - text: A stolen *server-side* copy of the vault can't be brute-forced, because guessing the password alone isn't enough without the 128-bit account key
        correct: true
      - text: It lets Alice forget her password
      - text: It makes PBKDF2 faster
    explanation: |
      That's the design 1Password calls a Secret Key: a client-side pepper. An attacker
      who breaches the server gets vaults protected by password *and* 128 random bits.
      The cost is usability: Alice must keep the account key (or its printed recovery
      kit) to set up a new device.
---

Salts and peppers sound like a recipe, and they're often mixed up. They solve different
problems.

## Salt: every hash is its own target

Without a salt, identical passwords give identical hashes, and an attacker can
precompute hashes of a dictionary once (a *rainbow table*) and look up every leaked
hash instantly. A **salt** is a random value, unique per user or per vault, mixed into
the KDF and stored in the clear next to the result:

- The same password gives different results for different users.
- Precomputed tables are useless: the attacker would need a separate table per salt.
- Cracking a million leaked hashes costs a million times as much as cracking one.

Salts must be **unique** and **unpredictable**, which 16 bytes from `crypto/rand` give
you for free. A username is a poor salt: it's the same across sites and across
password changes, so tables can be built for popular usernames like `admin`.

## Pepper: a secret the database doesn't have

A **pepper** is a secret key, the same for all users, kept **outside** the database: in
a secrets manager, a hardware security module, or at least the application's
configuration. Before (or after) the slow KDF, the server mixes it in with HMAC:

```go
package main

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/sha256"
	"fmt"
)

// peppered derives a login verifier. pepper comes from the secrets manager,
// never from the database.
func peppered(pepper []byte, password string, salt []byte, iter int) ([]byte, error) {
	slow, err := pbkdf2.Key(sha256.New, password, salt, iter, 32)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, pepper)
	mac.Write(slow)
	return mac.Sum(nil), nil
}

func main() {
	pepper := []byte("from-the-secrets-manager-32bytes")
	v, _ := peppered(pepper, "cheese123", []byte("per-user-salt-16"), 1000)
	fmt.Printf("%x\n", v)
}
```

```
37f9851104ff264710af3c8bb1d017edc26992cf31201596cba050d4e0a1fef4
```

(The example uses 1,000 iterations so it runs instantly; use the real count in
production.)

If only the database leaks, which is the most common kind of breach thanks to SQL
injection and misplaced backups, the attacker has hashes they can't test a single guess
against. If the application server is also compromised, the pepper is gone and you're
back to relying on the salt and the slow KDF. So a pepper is **defence in depth**, not
a substitute.

Operational notes:

- Rotating a pepper means you can't recompute old hashes. Store a pepper **version**
  with each hash, keep old peppers until users have logged in again, and re-pepper on
  successful login.
- Apply HMAC with the pepper, as above. Don't just concatenate it to the password; that's
  back to rolling your own MAC.

## A client-side pepper for Keybox

Keybox's encryption happens on Alice's devices, not on the server, so a server-side
pepper doesn't help protect her vault from the server. But the same idea works on the
client. When Alice creates her account, Keybox generates a random 128-bit **account
key**, stores it on her devices and prints it on a recovery sheet. The vault key is then
derived from **both** the password (through the slow KDF) and the account key (through
HKDF, coming up in lesson 5).

A thief who steals vaults from the Keybox server now faces 2^128 possibilities on top
of the password. 1Password uses exactly this design, calling it the Secret Key. The
trade-off is real: lose every device and the recovery sheet, and nobody, including
Keybox, can open the vault. For an end-to-end-encrypted product, that's the point.
