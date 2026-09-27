---
title: Keys, Tokens and Nonces
quiz:
  - question: Keybox needs a share link token. Which is the best choice?
    options:
      - text: '`fmt.Sprintf("%d", time.Now().UnixNano())`'
      - text: '`uuid.NewV7().String()`'
      - text: 32 bytes from `crypto/rand`, encoded with `base64.RawURLEncoding`
        correct: true
      - text: '`sha256.Sum256` of the user''s email and the secret name'
    explanation: |
      Only the `crypto/rand` option is unguessable. The timestamp is guessable, the
      hash of known inputs is computable by anyone, and UUIDv7 is mostly timestamp plus
      a modest random part: great for IDs, not designed to be a secret.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    )

    const (
    	keySize   = 32 // AES-256 and HMAC-SHA256 keys
    	nonceSize = 12 // AES-GCM nonces
    	tokenSize = 32 // random bytes behind a share-link token
    )

    // newKey returns keySize bytes from crypto/rand.
    func newKey() []byte {
    	// ?
    	return make([]byte, keySize)
    }

    // newNonce returns nonceSize bytes from crypto/rand.
    func newNonce() []byte {
    	// ?
    	return make([]byte, nonceSize)
    }

    // newShareToken returns tokenSize random bytes encoded with
    // base64.RawURLEncoding, safe to put in a URL.
    func newShareToken() string {
    	// ?
    	return ""
    }

    // newRecoveryCode returns rand.Text() split into dash-separated groups
    // of 4 characters (the last group may be shorter), e.g.
    // "7CMZ-KHFS-SHA7-AP23-ZICE-ATWJ-K4".
    func newRecoveryCode() string {
    	// ?
    	return ""
    }

    func main() {
    	fmt.Printf("key:      %x\n", newKey())
    	fmt.Printf("nonce:    %x\n", newNonce())
    	fmt.Println("token:   ", newShareToken())
    	fmt.Println("recovery:", newRecoveryCode())
    }
  solution: |
    package main

    import (
    	"crypto/rand"
    	"encoding/base64"
    	"fmt"
    	"strings"
    )

    const (
    	keySize   = 32
    	nonceSize = 12
    	tokenSize = 32
    )

    func newKey() []byte {
    	key := make([]byte, keySize)
    	rand.Read(key)
    	return key
    }

    func newNonce() []byte {
    	nonce := make([]byte, nonceSize)
    	rand.Read(nonce)
    	return nonce
    }

    func newShareToken() string {
    	b := make([]byte, tokenSize)
    	rand.Read(b)
    	return base64.RawURLEncoding.EncodeToString(b)
    }

    func newRecoveryCode() string {
    	text := rand.Text()
    	var groups []string
    	for len(text) > 4 {
    		groups = append(groups, text[:4])
    		text = text[4:]
    	}
    	groups = append(groups, text)
    	return strings.Join(groups, "-")
    }

    func main() {
    	fmt.Printf("key:      %x\n", newKey())
    	fmt.Printf("nonce:    %x\n", newNonce())
    	fmt.Println("token:   ", newShareToken())
    	fmt.Println("recovery:", newRecoveryCode())
    }
  tests: |
    package main

    import (
    	"bytes"
    	"encoding/base64"
    	"strings"
    	"testing"
    	"testing/cryptotest"
    )

    func TestSizes(t *testing.T) {
    	if got := len(newKey()); got != keySize {
    		t.Errorf("len(newKey()) = %d, want %d", got, keySize)
    	}
    	if got := len(newNonce()); got != nonceSize {
    		t.Errorf("len(newNonce()) = %d, want %d", got, nonceSize)
    	}
    	tok := newShareToken()
    	raw, err := base64.RawURLEncoding.DecodeString(tok)
    	if err != nil {
    		t.Fatalf("newShareToken() = %q, which isn't valid base64.RawURLEncoding: %v", tok, err)
    	}
    	if len(raw) != tokenSize {
    		t.Errorf("newShareToken() decodes to %d bytes, want %d", len(raw), tokenSize)
    	}
    }

    func TestDistinct(t *testing.T) {
    	keys := map[string]bool{}
    	nonces := map[string]bool{}
    	tokens := map[string]bool{}
    	codes := map[string]bool{}
    	for range 1000 {
    		keys[string(newKey())] = true
    		nonces[string(newNonce())] = true
    		tokens[newShareToken()] = true
    		codes[newRecoveryCode()] = true
    	}
    	for name, n := range map[string]int{"newKey": len(keys), "newNonce": len(nonces), "newShareToken": len(tokens), "newRecoveryCode": len(codes)} {
    		if n != 1000 {
    			t.Errorf("1000 calls to %s gave only %d distinct values; is it random?", name, n)
    		}
    	}
    }

    func TestRecoveryCodeFormat(t *testing.T) {
    	for range 50 {
    		code := newRecoveryCode()
    		groups := strings.Split(code, "-")
    		joined := strings.Join(groups, "")
    		if len(joined) < 26 {
    			t.Fatalf("newRecoveryCode() = %q has %d characters without dashes; rand.Text gives at least 26", code, len(joined))
    		}
    		for i, g := range groups {
    			if len(g) == 0 || len(g) > 4 || (i < len(groups)-1 && len(g) != 4) {
    				t.Fatalf("newRecoveryCode() = %q: group %d is %q; want groups of 4 with only the last one shorter", code, i, g)
    			}
    		}
    		for _, r := range joined {
    			if !strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZ234567", r) {
    				t.Fatalf("newRecoveryCode() = %q contains %q, which isn't in rand.Text's base32 alphabet", code, r)
    			}
    		}
    	}
    }

    // Under cryptotest.SetGlobalRandom, crypto/rand becomes deterministic. If
    // these values don't repeat after resetting the seed, they didn't come
    // from crypto/rand.
    func TestUsesCryptoRand(t *testing.T) {
    	cryptotest.SetGlobalRandom(t, 11)
    	k1, n1, t1, c1 := newKey(), newNonce(), newShareToken(), newRecoveryCode()
    	cryptotest.SetGlobalRandom(t, 11)
    	k2, n2, t2, c2 := newKey(), newNonce(), newShareToken(), newRecoveryCode()
    	if !bytes.Equal(k1, k2) || !bytes.Equal(n1, n2) || t1 != t2 || c1 != c2 {
    		t.Errorf("values didn't repeat under a fixed crypto/rand seed; make sure all four functions use crypto/rand")
    	}
    }
---

Time to use `crypto/rand` for Keybox's own random values. There are three kinds, and
they differ in what they need to be.

## Keys: secret and random

A **key** must be secret and uniformly random. For symmetric crypto, just take bytes
straight from `crypto/rand`:

```go
key := make([]byte, 32) // 256 bits
rand.Read(key)
```

Never derive a key from something guessable (a username, a timestamp, a hash of either),
and never use a *password* directly as a key. Passwords are low-entropy; they need a
slow key derivation function first (chapter 5).

## Nonces: unique, not secret

A **nonce** ("number used once") isn't secret at all. It's sent right alongside the
ciphertext. Its one job is to be **unique per key**: using the same nonce twice with
the same key breaks AES-GCM badly (chapter 6). Two ways to get uniqueness:

- **Random**, from `crypto/rand`. Simple and stateless, but with 12-byte nonces you must
  stop after about 2^32 messages per key, or the chance of a random collision gets
  uncomfortable.
- **A counter**, incremented for every message. No collision risk, but you must never
  lose or roll back the counter, which is surprisingly hard across restarts and
  multiple servers.

Keybox encrypts a handful of secrets per user, so random nonces are the right choice.

## Tokens: secret, and printable

A **token** is a secret that travels as text: a share link, a reset link, an API key.
Take enough random bytes (32 is generous) and encode them for the medium:

- `base64.RawURLEncoding` is compact and URL-safe, with no `+`, `/` or `=` to escape.
- `hex.EncodeToString` is twice as long but uses only `0`-`9` and `a`-`f`.
- `rand.Text()` gives base32 (A-Z and 2-7), which avoids look-alike characters like
  `0`/`O` and `1`/`l`. Humans can read it aloud.

Encoding doesn't add or remove randomness. 32 random bytes are 256 bits whether you
print them as 43 base64 characters or 64 hex digits.

Store only a **hash** of long-lived tokens on the server, the same way you hashed
refresh tokens in [Refresh Tokens](/courses/learn-http-servers/authentication/refresh-tokens).
A fast hash like SHA-256 is fine here, because unlike passwords the tokens are already
256 bits of randomness.

## What about UUIDs?

Go 1.27's `uuid` package (see
[IDs with UUID](/courses/learn-http-servers/storage/ids-with-uuid)) makes great **identifiers**. A version 4 UUID has
122 random bits, but UUIDs are designed to be unique, not secret: they get logged,
put in URLs and shown in dashboards. Version 7 is mostly a timestamp. Use a UUID to
*name* a secret, and a token to *grant access* to it.

## Your task

Fill in Keybox's four generators:

- `newKey()`: `keySize` bytes from `crypto/rand`.
- `newNonce()`: `nonceSize` bytes from `crypto/rand`.
- `newShareToken()`: `tokenSize` random bytes, encoded with `base64.RawURLEncoding`.
- `newRecoveryCode()`: `rand.Text()` split into groups of 4 characters joined with
  `-`. The last group may be shorter. Don't assume `rand.Text` returns exactly 26
  characters; its docs say future versions may return more.

You'll need to add `crypto/rand`, `encoding/base64` and probably `strings` to the
imports. The tests check sizes, formats and that 1,000 calls give 1,000 different
values. They also use `testing/cryptotest` (next-but-one lesson) to confirm your values
really come from `crypto/rand`.
