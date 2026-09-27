---
title: Associated Data
quiz:
  - question: Keybox passes the secret's name as associated data when sealing. Is the name encrypted?
    options:
      - text: Yes, associated data is encrypted along with the plaintext
      - text: No; it isn't stored in the ciphertext at all, but it's authenticated, so `Open` fails unless the caller supplies exactly the same name
        correct: true
      - text: No, and it isn't authenticated either
      - text: Only its first 16 bytes are encrypted
    explanation: |
      Associated data is bound into the tag but never encrypted or included in the
      output. The receiver must already know it (here, from the map key or database
      row) and pass it to `Open`.
exercise:
  starter: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"encoding/binary"
    	"errors"
    	"fmt"
    )

    var (
    	ErrNotFound = errors.New("keybox: no such secret")
    	ErrDecrypt  = errors.New("keybox: decryption failed")
    )

    // Vault stores sealed secrets by name. entries is what gets written to disk
    // and synced to the (untrusted) server.
    type Vault struct {
    	id      string
    	gcm     cipher.AEAD
    	entries map[string][]byte
    }

    func NewVault(id string, key []byte) (*Vault, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	gcm, err := cipher.NewGCMWithRandomNonce(block)
    	if err != nil {
    		return nil, err
    	}
    	return &Vault{id: id, gcm: gcm, entries: map[string][]byte{}}, nil
    }

    // entryAD returns the associated data for a secret: a label, the vault ID
    // and the secret's name, each length-prefixed.
    func entryAD(vaultID, name string) []byte {
    	var ad []byte
    	for _, f := range []string{"keybox v1 vault entry", vaultID, name} {
    		ad = binary.BigEndian.AppendUint64(ad, uint64(len(f)))
    		ad = append(ad, f...)
    	}
    	return ad
    }

    // Put seals secret and stores it under name.
    func (v *Vault) Put(name string, secret []byte) {
    	// ? Bind the ciphertext to this vault and name.
    	v.entries[name] = v.gcm.Seal(nil, nil, secret, nil)
    }

    // Get returns the secret stored under name.
    func (v *Vault) Get(name string) ([]byte, error) {
    	sealed, ok := v.entries[name]
    	if !ok {
    		return nil, ErrNotFound
    	}
    	// ?
    	pt, err := v.gcm.Open(nil, nil, sealed, nil)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func main() {
    	key := []byte("an example 32-byte key for AES!!")
    	v, _ := NewVault("vault-alice", key)
    	v.Put("github-token", []byte("ghp_alice"))
    	v.Put("wifi-password", []byte("hunter2"))

    	// The malicious server swaps two entries in the synced file.
    	v.entries["github-token"], v.entries["wifi-password"] = v.entries["wifi-password"], v.entries["github-token"]

    	got, err := v.Get("github-token")
    	fmt.Printf("github-token after swap: %q %v\n", got, err)
    }
  solution: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"encoding/binary"
    	"errors"
    	"fmt"
    )

    var (
    	ErrNotFound = errors.New("keybox: no such secret")
    	ErrDecrypt  = errors.New("keybox: decryption failed")
    )

    type Vault struct {
    	id      string
    	gcm     cipher.AEAD
    	entries map[string][]byte
    }

    func NewVault(id string, key []byte) (*Vault, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	gcm, err := cipher.NewGCMWithRandomNonce(block)
    	if err != nil {
    		return nil, err
    	}
    	return &Vault{id: id, gcm: gcm, entries: map[string][]byte{}}, nil
    }

    func entryAD(vaultID, name string) []byte {
    	var ad []byte
    	for _, f := range []string{"keybox v1 vault entry", vaultID, name} {
    		ad = binary.BigEndian.AppendUint64(ad, uint64(len(f)))
    		ad = append(ad, f...)
    	}
    	return ad
    }

    func (v *Vault) Put(name string, secret []byte) {
    	v.entries[name] = v.gcm.Seal(nil, nil, secret, entryAD(v.id, name))
    }

    func (v *Vault) Get(name string) ([]byte, error) {
    	sealed, ok := v.entries[name]
    	if !ok {
    		return nil, ErrNotFound
    	}
    	pt, err := v.gcm.Open(nil, nil, sealed, entryAD(v.id, name))
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func main() {
    	key := []byte("an example 32-byte key for AES!!")
    	v, _ := NewVault("vault-alice", key)
    	v.Put("github-token", []byte("ghp_alice"))
    	v.Put("wifi-password", []byte("hunter2"))

    	v.entries["github-token"], v.entries["wifi-password"] = v.entries["wifi-password"], v.entries["github-token"]

    	got, err := v.Get("github-token")
    	fmt.Printf("github-token after swap: %q %v\n", got, err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"errors"
    	"testing"
    )

    var testKey = bytes.Repeat([]byte{0x44}, 32)

    func newTestVault(t *testing.T, id string) *Vault {
    	t.Helper()
    	v, err := NewVault(id, testKey)
    	if err != nil {
    		t.Fatal(err)
    	}
    	return v
    }

    func TestRoundTrip(t *testing.T) {
    	v := newTestVault(t, "vault-alice")
    	v.Put("github-token", []byte("ghp_alice"))
    	v.Put("empty", nil)
    	if got, err := v.Get("github-token"); err != nil || string(got) != "ghp_alice" {
    		t.Errorf("Get(github-token) = %q, %v; want \"ghp_alice\", nil", got, err)
    	}
    	if got, err := v.Get("empty"); err != nil || len(got) != 0 {
    		t.Errorf("Get(empty) = %q, %v; want empty, nil", got, err)
    	}
    	if _, err := v.Get("missing"); !errors.Is(err, ErrNotFound) {
    		t.Errorf("Get(missing) error = %v, want ErrNotFound", err)
    	}
    }

    func TestSwappedEntriesRejected(t *testing.T) {
    	v := newTestVault(t, "vault-alice")
    	v.Put("github-token", []byte("ghp_alice"))
    	v.Put("wifi-password", []byte("hunter2"))
    	v.entries["github-token"], v.entries["wifi-password"] = v.entries["wifi-password"], v.entries["github-token"]
    	for _, name := range []string{"github-token", "wifi-password"} {
    		if got, err := v.Get(name); !errors.Is(err, ErrDecrypt) {
    			t.Errorf("after swapping entries, Get(%q) = %q, %v; want ErrDecrypt (pass entryAD(v.id, name) to Seal and Open)", name, got, err)
    		}
    	}
    }

    func TestCopiedEntryRejected(t *testing.T) {
    	v := newTestVault(t, "vault-alice")
    	v.Put("github-token", []byte("ghp_alice"))
    	v.entries["backup-of-token"] = v.entries["github-token"]
    	if got, err := v.Get("backup-of-token"); !errors.Is(err, ErrDecrypt) {
    		t.Errorf("entry copied to a new name: Get = %q, %v; want ErrDecrypt", got, err)
    	}
    }

    func TestMovedBetweenVaultsRejected(t *testing.T) {
    	alice := newTestVault(t, "vault-alice")
    	shared := newTestVault(t, "vault-team") // same key, different vault
    	alice.Put("github-token", []byte("ghp_alice"))
    	shared.entries["github-token"] = alice.entries["github-token"]
    	if got, err := shared.Get("github-token"); !errors.Is(err, ErrDecrypt) {
    		t.Errorf("entry moved to another vault: Get = %q, %v; want ErrDecrypt", got, err)
    	}
    }
---

AES-GCM stops Mallory from changing a single bit of Keybox's ciphertexts. But it
doesn't stop her from **moving whole ciphertexts around**. Associated data fixes that.

## The swap attack

Keybox syncs Alice's vault through its server, which Keybox's threat model says is
untrusted. The synced file maps names to sealed values:

```
"github-token"  -> 9c1e...  (sealed "ghp_alice")
"wifi-password" -> 51d0...  (sealed "hunter2")
```

Each value is a perfectly valid AES-GCM ciphertext under Alice's key. A malicious server
can't read or edit them, but it can **swap** them. Now Keybox decrypts the
`github-token` entry successfully and gets `hunter2`. Depending on what the client does
next, that could paste the Wi-Fi password into a GitHub login form, or send the
"github-token" to a script that posts it somewhere less trusted. Similar tricks: copy
an entry to a new name, move an entry from Alice's personal vault into a shared team
vault encrypted with the same key, or replay an old version of an entry.

The tag proves "this ciphertext was made by a key holder". It doesn't say **for what**.

## Binding context with associated data

The AEAD's last parameter, `additionalData`, fixes this:

```go
sealed := gcm.Seal(nil, nil, secret, entryAD(vaultID, name))
secret, err := gcm.Open(nil, nil, sealed, entryAD(vaultID, name))
```

Associated data (AD) is **authenticated but not encrypted**, and it's **not stored** in
the output. The tag covers it, so `Open` succeeds only if the caller passes the exact
same bytes. Swap two entries and each one is opened with the wrong name, so the tag
check fails.

What belongs in the AD? Anything that defines where this ciphertext is *supposed* to be
and what it means, that the reader already knows:

- a **label and version**: `"keybox v1 vault entry"`,
- the **vault ID** (so entries can't move between vaults),
- the **entry name** or database row ID (so they can't move between entries),
- sometimes a **version number** of the entry, to stop rollback to an older value
  (that also needs the reader to know which version is current).

Encode it unambiguously, with length prefixes, for the same reason as in chapter 3: the
AD `("vault-a", "lice/x")` must never equal `("vault-al", "ice/x")`.

## Things AD doesn't do

- It doesn't hide the name. If names are sensitive (they reveal which services Alice
  uses), encrypt them too, perhaps as part of the plaintext, and look entries up by a
  random ID or an HMAC of the name.
- It doesn't prevent **deleting** entries or rolling back the whole vault to an older,
  consistent state. Detecting that needs something like a signed, versioned manifest.

## Your task

Keybox's `Vault` seals entries with no associated data, so the swap attack works; run the
starter to see `github-token` come back as `hunter2`. Fix `Put` and `Get` to pass
`entryAD(v.id, name)` as the associated data to `Seal` and `Open`. Everything else stays
the same. The tests swap entries, copy one to a new name, and move one to another vault
with the same key, and expect `ErrDecrypt` every time.
