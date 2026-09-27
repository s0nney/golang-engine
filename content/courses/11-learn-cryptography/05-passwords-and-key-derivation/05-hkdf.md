---
title: Deriving Keys with HKDF
quiz:
  - question: Keybox needs a 32-byte AES key from the raw output of an X25519 key exchange. What should it do?
    options:
      - text: Use the X25519 output directly as the AES key
      - text: Run it through PBKDF2 with 600,000 iterations
      - text: Run it through HKDF with a label describing the purpose
        correct: true
      - text: Hash it with MD5 to get 16 bytes
    explanation: |
      Key-exchange outputs are secret but not uniformly random bit strings, so they need
      HKDF's extract step, and the label binds the key to its purpose. PBKDF2's slowness
      only helps for low-entropy *passwords*; a 256-bit shared secret doesn't need it.
  - question: What is HKDF's `info` parameter for?
    options:
      - text: It's a second secret key
      - text: It's a public, application-specific label, so different purposes derive different, independent keys from the same secret
        correct: true
      - text: It must be a random nonce
      - text: It's the output length
    explanation: |
      `info` provides context and domain separation. It isn't secret. Different `info`
      strings give unrelated keys, which is how one master key safely becomes many.
exercise:
  starter: |
    package main

    import (
    	"bytes"
    	"errors"
    	"fmt"
    )

    const subkeySize = 32

    // vaultKeys are Keybox's per-purpose keys, all derived from one master key.
    type vaultKeys struct {
    	Vault []byte // info "keybox v1 vault encryption"
    	Sync  []byte // info "keybox v1 sync request mac"
    	Share []byte // info "keybox v1 share tokens"
    }

    // deriveVaultKeys derives the three keys with HKDF-SHA256, using masterKey
    // (from PBKDF2) as the secret and the device's accountKey as the salt.
    // Each key is subkeySize bytes. It returns an error if masterKey isn't
    // exactly 32 bytes or accountKey isn't exactly 16 bytes.
    func deriveVaultKeys(masterKey, accountKey []byte) (vaultKeys, error) {
    	// ?
    	return vaultKeys{}, errors.New("not implemented")
    }

    func main() {
    	master := bytes.Repeat([]byte{1}, 32)
    	account := bytes.Repeat([]byte{2}, 16)
    	keys, err := deriveVaultKeys(master, account)
    	if err != nil {
    		fmt.Println("error:", err)
    		return
    	}
    	fmt.Printf("vault: %x\nsync:  %x\nshare: %x\n", keys.Vault, keys.Sync, keys.Share)
    }
  solution: |
    package main

    import (
    	"bytes"
    	"crypto/hkdf"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    )

    const subkeySize = 32

    type vaultKeys struct {
    	Vault []byte
    	Sync  []byte
    	Share []byte
    }

    func deriveVaultKeys(masterKey, accountKey []byte) (vaultKeys, error) {
    	if len(masterKey) != 32 || len(accountKey) != 16 {
    		return vaultKeys{}, errors.New("master key must be 32 bytes and account key 16 bytes")
    	}
    	var keys vaultKeys
    	for _, k := range []struct {
    		dst  *[]byte
    		info string
    	}{
    		{&keys.Vault, "keybox v1 vault encryption"},
    		{&keys.Sync, "keybox v1 sync request mac"},
    		{&keys.Share, "keybox v1 share tokens"},
    	} {
    		key, err := hkdf.Key(sha256.New, masterKey, accountKey, k.info, subkeySize)
    		if err != nil {
    			return vaultKeys{}, err
    		}
    		*k.dst = key
    	}
    	return keys, nil
    }

    func main() {
    	master := bytes.Repeat([]byte{1}, 32)
    	account := bytes.Repeat([]byte{2}, 16)
    	keys, err := deriveVaultKeys(master, account)
    	if err != nil {
    		fmt.Println("error:", err)
    		return
    	}
    	fmt.Printf("vault: %x\nsync:  %x\nshare: %x\n", keys.Vault, keys.Sync, keys.Share)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"encoding/hex"
    	"testing"
    )

    func TestDeriveVaultKeysKnownAnswers(t *testing.T) {
    	keys, err := deriveVaultKeys(bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 16))
    	if err != nil {
    		t.Fatalf("deriveVaultKeys(valid keys) error = %v", err)
    	}
    	for _, tt := range []struct {
    		name string
    		got  []byte
    		want string
    	}{
    		{"Vault (info \"keybox v1 vault encryption\")", keys.Vault, "8c5954a1ef6216467487d6cd9155e1f58579d4171f11e9c86d8b895360ce7c6d"},
    		{"Sync (info \"keybox v1 sync request mac\")", keys.Sync, "83340789dfcaca31bc641e7e32ee6027208bd35cf99d452a8d7eff2cf69d743b"},
    		{"Share (info \"keybox v1 share tokens\")", keys.Share, "3f05d664e5899b79fff9f82f8bf85b2f0c4c67facaeeb0e6e6e30c41197f0d14"},
    	} {
    		if got := hex.EncodeToString(tt.got); got != tt.want {
    			t.Errorf("%s = %s, want %s (check the hash, the secret/salt order and the info string)", tt.name, got, tt.want)
    		}
    	}
    }

    func TestAccountKeyMatters(t *testing.T) {
    	master := bytes.Repeat([]byte{1}, 32)
    	a, _ := deriveVaultKeys(master, bytes.Repeat([]byte{2}, 16))
    	b, _ := deriveVaultKeys(master, bytes.Repeat([]byte{3}, 16))
    	if bytes.Equal(a.Vault, b.Vault) {
    		t.Errorf("different account keys gave the same vault key; use accountKey as the HKDF salt")
    	}
    }

    func TestBadInputs(t *testing.T) {
    	for _, tt := range []struct{ m, a int }{{31, 16}, {32, 15}, {0, 16}, {32, 0}, {64, 16}} {
    		if _, err := deriveVaultKeys(make([]byte, tt.m), make([]byte, tt.a)); err == nil {
    			t.Errorf("deriveVaultKeys(%d-byte master key, %d-byte account key) error = nil, want an error", tt.m, tt.a)
    		}
    	}
    }
---

PBKDF2 turned a weak password into one strong master key. Keybox actually needs several
keys (one per purpose, as chapter 4 argued), and later it'll need keys from other
sources, like key exchanges. The tool for turning one good secret into many keys is
**HKDF**, in `crypto/hkdf` since Go 1.24.

## Extract, then expand

HKDF (RFC 5869) is built from HMAC in two steps:

1. **Extract**: `PRK = HMAC(salt, secret)`. This concentrates whatever randomness
   the input secret has into a uniformly random 32-byte pseudorandom key. Inputs like a
   Diffie-Hellman shared secret are hard to guess but have structure; extract cleans
   that up. The **salt** is optional and non-secret, ideally random.
2. **Expand**: `OKM = HMAC(PRK, info || 0x01) || HMAC(PRK, T1 || info || 0x02) || ...`,
   producing as many bytes as you ask for. The **info** string is a public label that
   makes each derived key independent.

Go gives you both halves (`hkdf.Extract`, `hkdf.Expand`) and the combination
`hkdf.Key`, which is what you'll normally use:

```go
package main

import (
	"bytes"
	"crypto/hkdf"
	"crypto/sha256"
	"fmt"
)

func main() {
	// RFC 5869, test case 1.
	ikm := bytes.Repeat([]byte{0x0b}, 22)
	salt := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c}
	info := "\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9"

	okm, err := hkdf.Key(sha256.New, ikm, salt, info, 42)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%x\n", okm)
}
```

```
3cb25f25faacd57a90434f64d0362f2a2d2d0a90cf1a5a4c5db02d56ecc4c5bf34007208d5b887185865
```

That's RFC 5869's first test vector. Note the argument order: **secret, then salt, then
info**, then length. `info` is a `string`, which nudges you toward human-readable labels.

## HKDF is not a password KDF

HKDF is fast: it's a couple of HMACs. That's right for inputs that are already strong
(random keys, key-exchange outputs), and badly wrong for passwords, which need the slow,
salted functions from earlier in this chapter. The usual pipeline chains them:

```
password --PBKDF2/Argon2 (slow)--> master key --HKDF (fast)--> per-purpose keys
```

## Keybox's key hierarchy

Putting the chapter together, Keybox derives its keys like this:

- `masterKey = PBKDF2-SHA256(password, header.Salt, header.Iterations, 32)`
- `Vault = HKDF-SHA256(masterKey, salt: accountKey, info: "keybox v1 vault encryption")`
- `Sync = HKDF-SHA256(masterKey, salt: accountKey, info: "keybox v1 sync request mac")`
- `Share = HKDF-SHA256(masterKey, salt: accountKey, info: "keybox v1 share tokens")`

The 16-byte **account key** from the salts-and-peppers lesson goes in as the HKDF salt.
HKDF's extract step is itself an HMAC keyed with the salt, so a secret salt is fine and
mixes the account key's 128 bits into every derived key. Without it, a thief holding the
vault still can't derive anything by guessing the password alone.

## Your task

Complete `deriveVaultKeys(masterKey, accountKey)`:

1. Return an error unless `masterKey` is exactly 32 bytes and `accountKey` exactly 16.
2. Derive each of `Vault`, `Sync` and `Share` with `hkdf.Key(sha256.New, masterKey,
   accountKey, info, subkeySize)`, using the `info` strings in the struct's comments,
   spelled exactly.

The tests check exact output bytes, so a typo in a label shows up immediately. That's a
feature: a label is part of the protocol.
