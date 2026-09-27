---
title: Key Separation
quiz:
  - question: Keybox uses one HMAC key both for session tokens (`HMAC(key, userID)`) and for share-link tokens (`HMAC(key, secretID)`). Why is that dangerous?
    options:
      - text: HMAC can only be used once per key
      - text: If a secret ID ever equals a user ID's bytes, a share-link token doubles as a valid session token for that user
        correct: true
      - text: It makes HMAC slower
      - text: It isn't dangerous; HMAC is secure
    explanation: |
      HMAC guarantees nobody *without the key* can forge a tag. It doesn't stop the
      system from treating a legitimate tag made for one purpose as valid for another.
      Separate keys per purpose (or at least a purpose label inside the MAC) close
      that cross-protocol hole.
  - question: Why should key-derivation labels include a version, as in `"keybox v1 vault encryption"`?
    options:
      - text: HMAC requires a version number in its input
      - text: So that when the format changes, keys and data for `v2` can never be confused with `v1`
        correct: true
      - text: To make the derived key longer
      - text: So attackers can't guess the label
    explanation: |
      Labels aren't secret; their job is to keep contexts apart. Putting the version in
      the label means a new format automatically gets new, unrelated keys, and old
      data can't be misread under new rules.
---

One key, one purpose. It's one of the most useful rules in applied cryptography, and
one of the most often broken.

## Why reuse hurts

Every primitive's security proof assumes the key is used only as that primitive
specifies. Use the same key for two things and you're outside the proof, where odd
interactions live:

- **Cross-protocol confusion.** A tag the server issued for a share link happens to be
  valid input to the session-token checker. Nothing was "broken"; the two uses simply
  weren't kept apart.
- **Different algorithms, same key.** Using one key for both AES-CBC encryption and a
  CBC-MAC, or for both HMAC and encryption, has led to real attacks where one operation
  becomes an oracle for the other.
- **Blast radius.** If the webhook secret, the token key and the encryption key are all
  the same bytes, one leak (say, the webhook secret pasted into a support ticket)
  compromises everything. Separate keys can be rotated and revoked separately.

## Keybox's keys

Keybox needs several symmetric keys:

| Purpose | Used with |
| --- | --- |
| Encrypting vault entries | AES-256-GCM |
| MACing sync requests | HMAC-SHA256 |
| Share-link tokens | HMAC-SHA256 |

It would be annoying to generate, store and back up three random keys per user. The
usual answer is to keep **one master key** and **derive** each purpose's key from it,
using a label that names the purpose.

## Deriving subkeys

HMAC is a pseudorandom function: with a secret key, its outputs look random and
unrelated for different inputs. So `HMAC(master, label)` gives a fresh, independent key
per label:

```go
package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func subkey(master []byte, label string) []byte {
	mac := hmac.New(sha256.New, master)
	mac.Write([]byte(label))
	return mac.Sum(nil)
}

func main() {
	master := bytes.Repeat([]byte{0x42}, 32)
	enc := subkey(master, "keybox v1 vault encryption")
	sync := subkey(master, "keybox v1 sync request mac")
	fmt.Println(bytes.Equal(enc, sync))
	fmt.Printf("%x\n%x\n", enc[:8], sync[:8])
}
```

```
false
14a30987c2c86755
d301a8b55952788e
```

This is the core of **HKDF**, the standard key derivation function built from HMAC,
which you'll use properly in the next chapter. HKDF adds a proper "extract" step for
input keys that aren't uniformly random (like the output of a Diffie-Hellman exchange),
and it lets you ask for any output length. Don't ship the one-liner above; use
`crypto/hkdf`.

## Labels everywhere

Even with separate keys, labels are cheap insurance. You've now seen the same idea three
times:

- in **hashes**, a label as the first field (`"keybox share fingerprint v1"`),
- in **MACs**, a label in the signed data (`"keybox sync request v1"`),
- in **key derivation**, a label per derived key (`"keybox v1 vault encryption"`).

Include a **version** in each label. When you change the format in a year, `v2` data
can never be confused with `v1`. Signatures get the same treatment in chapter 8.
