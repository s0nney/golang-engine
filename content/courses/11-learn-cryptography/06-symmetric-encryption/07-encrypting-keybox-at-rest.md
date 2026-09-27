---
title: Encrypting Keybox at Rest
quiz:
  - question: Alice changes her Keybox password. With key wrapping, what has to be re-encrypted?
    options:
      - text: Every secret in the vault
      - text: Only the wrapped data key, sealed under a key derived from the new password
        correct: true
      - text: Nothing; the old password keeps working
      - text: The vault header's salt, but nothing else
    explanation: |
      Secrets are sealed with a random data key that never changes. The password only
      protects that one 32-byte key, so a password change re-wraps 32 bytes instead of
      re-encrypting everything.
exercise:
  starter: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/pbkdf2"
    	"crypto/rand"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    )

    var ErrWrongPassword = errors.New("keybox: wrong password or corrupted vault")

    const wrapAD = "keybox v1 wrapped data key"

    type vaultHeader struct {
    	Salt       []byte
    	Iterations int
    }

    // vaultFile is what Keybox writes to disk: plaintext KDF parameters and the
    // data key, sealed under a key-encryption key (KEK) derived from the password.
    type vaultFile struct {
    	Header     vaultHeader
    	WrappedKey []byte
    }

    // deriveKEK turns a password into a 32-byte key-encryption key.
    func deriveKEK(password string, h vaultHeader) ([]byte, error) {
    	return pbkdf2.Key(sha256.New, password, h.Salt, h.Iterations, 32)
    }

    // newAEAD returns AES-256-GCM with managed random nonces.
    func newAEAD(key []byte) (cipher.AEAD, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	return cipher.NewGCMWithRandomNonce(block)
    }

    // newHeader returns a header with a fresh 16-byte random salt.
    func newHeader(iterations int) vaultHeader {
    	salt := make([]byte, 16)
    	rand.Read(salt)
    	return vaultHeader{Salt: salt, Iterations: iterations}
    }

    // createVault generates a random 32-byte data key and wraps it under the
    // password: seal it with the KEK, using wrapAD as associated data.
    func createVault(password string, iterations int) (vaultFile, []byte, error) {
    	// ?
    	return vaultFile{}, nil, errors.New("not implemented")
    }

    // unlockVault derives the KEK and unwraps the data key. Any failure to
    // open the wrapped key returns ErrWrongPassword.
    func unlockVault(password string, f vaultFile) ([]byte, error) {
    	// ?
    	return nil, ErrWrongPassword
    }

    // changePassword unlocks f with oldPassword and returns a new vaultFile
    // with a fresh header (new salt, the given iterations) whose WrappedKey
    // holds the SAME data key, wrapped under newPassword.
    func changePassword(f vaultFile, oldPassword, newPassword string, iterations int) (vaultFile, error) {
    	// ?
    	return f, nil
    }

    func main() {
    	f, dataKey, err := createVault("correct horse battery staple", 600_000)
    	fmt.Printf("created: %d-byte wrapped key, err %v\n", len(f.WrappedKey), err)

    	got, err := unlockVault("correct horse battery staple", f)
    	fmt.Println("unlock:", err == nil && string(got) == string(dataKey), err)

    	_, err = unlockVault("Tr0ub4dor&3", f)
    	fmt.Println("wrong password:", err)

    	f2, err := changePassword(f, "correct horse battery staple", "new passphrase for alice", 600_000)
    	fmt.Println("changed:", err)
    	got, err = unlockVault("new passphrase for alice", f2)
    	fmt.Println("same data key:", err == nil && string(got) == string(dataKey))
    }
  solution: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/pbkdf2"
    	"crypto/rand"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    )

    var ErrWrongPassword = errors.New("keybox: wrong password or corrupted vault")

    const wrapAD = "keybox v1 wrapped data key"

    type vaultHeader struct {
    	Salt       []byte
    	Iterations int
    }

    type vaultFile struct {
    	Header     vaultHeader
    	WrappedKey []byte
    }

    func deriveKEK(password string, h vaultHeader) ([]byte, error) {
    	return pbkdf2.Key(sha256.New, password, h.Salt, h.Iterations, 32)
    }

    func newAEAD(key []byte) (cipher.AEAD, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	return cipher.NewGCMWithRandomNonce(block)
    }

    func newHeader(iterations int) vaultHeader {
    	salt := make([]byte, 16)
    	rand.Read(salt)
    	return vaultHeader{Salt: salt, Iterations: iterations}
    }

    func wrap(password string, h vaultHeader, dataKey []byte) (vaultFile, error) {
    	kek, err := deriveKEK(password, h)
    	if err != nil {
    		return vaultFile{}, err
    	}
    	aead, err := newAEAD(kek)
    	if err != nil {
    		return vaultFile{}, err
    	}
    	return vaultFile{Header: h, WrappedKey: aead.Seal(nil, nil, dataKey, []byte(wrapAD))}, nil
    }

    func createVault(password string, iterations int) (vaultFile, []byte, error) {
    	dataKey := make([]byte, 32)
    	rand.Read(dataKey)
    	f, err := wrap(password, newHeader(iterations), dataKey)
    	if err != nil {
    		return vaultFile{}, nil, err
    	}
    	return f, dataKey, nil
    }

    func unlockVault(password string, f vaultFile) ([]byte, error) {
    	kek, err := deriveKEK(password, f.Header)
    	if err != nil {
    		return nil, ErrWrongPassword
    	}
    	aead, err := newAEAD(kek)
    	if err != nil {
    		return nil, ErrWrongPassword
    	}
    	dataKey, err := aead.Open(nil, nil, f.WrappedKey, []byte(wrapAD))
    	if err != nil {
    		return nil, ErrWrongPassword
    	}
    	return dataKey, nil
    }

    func changePassword(f vaultFile, oldPassword, newPassword string, iterations int) (vaultFile, error) {
    	dataKey, err := unlockVault(oldPassword, f)
    	if err != nil {
    		return vaultFile{}, err
    	}
    	return wrap(newPassword, newHeader(iterations), dataKey)
    }

    func main() {
    	f, dataKey, err := createVault("correct horse battery staple", 600_000)
    	fmt.Printf("created: %d-byte wrapped key, err %v\n", len(f.WrappedKey), err)

    	got, err := unlockVault("correct horse battery staple", f)
    	fmt.Println("unlock:", err == nil && string(got) == string(dataKey), err)

    	_, err = unlockVault("Tr0ub4dor&3", f)
    	fmt.Println("wrong password:", err)

    	f2, err := changePassword(f, "correct horse battery staple", "new passphrase for alice", 600_000)
    	fmt.Println("changed:", err)
    	got, err = unlockVault("new passphrase for alice", f2)
    	fmt.Println("same data key:", err == nil && string(got) == string(dataKey))
    }
  tests: |
    package main

    import (
    	"bytes"
    	"errors"
    	"testing"
    )

    const testIter = 1000 // low so the tests are fast; production uses 600,000

    func TestCreateAndUnlock(t *testing.T) {
    	f, dataKey, err := createVault("alice's passphrase", testIter)
    	if err != nil {
    		t.Fatalf("createVault error = %v", err)
    	}
    	if len(dataKey) != 32 {
    		t.Fatalf("createVault returned a %d-byte data key, want 32", len(dataKey))
    	}
    	if len(f.Header.Salt) != 16 || f.Header.Iterations != testIter {
    		t.Errorf("header = %d-byte salt, %d iterations; want 16-byte salt, %d", len(f.Header.Salt), f.Header.Iterations, testIter)
    	}
    	if want := 12 + 32 + 16; len(f.WrappedKey) != want {
    		t.Errorf("WrappedKey is %d bytes, want %d (nonce + sealed 32-byte key + tag)", len(f.WrappedKey), want)
    	}
    	if bytes.Contains(f.WrappedKey, dataKey) {
    		t.Fatalf("the data key appears in plaintext inside WrappedKey")
    	}
    	got, err := unlockVault("alice's passphrase", f)
    	if err != nil || !bytes.Equal(got, dataKey) {
    		t.Errorf("unlockVault(right password) = %x, %v; want the data key", got, err)
    	}
    	_, other, _ := createVault("alice's passphrase", testIter)
    	if bytes.Equal(other, dataKey) {
    		t.Errorf("two vaults got the same data key; generate it with crypto/rand")
    	}
    }

    func TestWrongPasswordAndTampering(t *testing.T) {
    	f, _, _ := createVault("alice's passphrase", testIter)
    	if _, err := unlockVault("alice's passphrasE", f); !errors.Is(err, ErrWrongPassword) {
    		t.Errorf("unlockVault(wrong password) error = %v, want ErrWrongPassword", err)
    	}
    	bad := f
    	bad.WrappedKey = bytes.Clone(f.WrappedKey)
    	bad.WrappedKey[20] ^= 1
    	if _, err := unlockVault("alice's passphrase", bad); !errors.Is(err, ErrWrongPassword) {
    		t.Errorf("unlockVault(tampered wrapped key) error = %v, want ErrWrongPassword", err)
    	}
    	bad = f
    	bad.Header.Iterations = 1
    	if _, err := unlockVault("alice's passphrase", bad); !errors.Is(err, ErrWrongPassword) {
    		t.Errorf("unlockVault(header iterations changed) error = %v, want ErrWrongPassword", err)
    	}
    }

    func TestChangePassword(t *testing.T) {
    	f, dataKey, _ := createVault("old passphrase", testIter)
    	f2, err := changePassword(f, "old passphrase", "new passphrase", testIter+1)
    	if err != nil {
    		t.Fatalf("changePassword error = %v", err)
    	}
    	if bytes.Equal(f2.Header.Salt, f.Header.Salt) {
    		t.Errorf("changePassword kept the old salt; use a fresh header")
    	}
    	if f2.Header.Iterations != testIter+1 {
    		t.Errorf("changePassword header iterations = %d, want %d", f2.Header.Iterations, testIter+1)
    	}
    	got, err := unlockVault("new passphrase", f2)
    	if err != nil || !bytes.Equal(got, dataKey) {
    		t.Errorf("unlock with the new password = %x, %v; want the original data key", got, err)
    	}
    	if _, err := unlockVault("old passphrase", f2); !errors.Is(err, ErrWrongPassword) {
    		t.Errorf("the old password still unlocks the new vault file: error = %v", err)
    	}
    	if _, err := changePassword(f, "not the password", "new passphrase", testIter); !errors.Is(err, ErrWrongPassword) {
    		t.Errorf("changePassword with the wrong old password error = %v, want ErrWrongPassword", err)
    	}
    }
---

You now have every piece to encrypt Keybox's vault at rest: PBKDF2 for the password,
AES-GCM with managed nonces, and associated data. One design question remains: should
the password-derived key encrypt the secrets directly?

## Two keys: KEK and DEK

If the key from the password encrypts every secret, changing the password means
decrypting and re-encrypting the whole vault, on every device, atomically. Instead,
almost every real system uses **key wrapping** with two keys:

- A **data encryption key** (DEK): 32 random bytes from `crypto/rand`, generated once
  when the vault is created. It seals every secret, as in the last lesson.
- A **key encryption key** (KEK): derived from the password with PBKDF2. Its only job is
  to seal ("wrap") the DEK.

The vault file stores the KDF header, the wrapped DEK, and the sealed secrets:

```
header:      salt, iterations        (plaintext)
wrapped key: AES-GCM(KEK, DEK)       (60 bytes)
entries:     AES-GCM(DEK, secret, AD: vault ID + name) ...
```

Unlocking derives the KEK (the slow step), unwraps the DEK, and keeps it in memory
while the vault is open.

## What this buys

- **Cheap password changes.** Re-wrap 32 bytes under a new KEK, with a new salt. The
  entries don't change.
- **Multiple ways in.** The same DEK can be wrapped several times: under the password,
  under a recovery key printed on paper, under a hardware key. Each wrapping is
  independent. (Chapter 7 adds one more: wrapped to a *public* key, for sharing.)
- **A strong key for the bulk data**, whatever the password quality. The password's
  weakness is contained in one place, behind the slow KDF.

Cloud key management services (AWS KMS, Google Cloud KMS) are built on the same idea,
called **envelope encryption**: a master key in the KMS wraps data keys that you use
locally.

One honest caveat: a password change doesn't protect against someone who already copied
the *old* vault file and knows the *old* password, because the DEK is the same. If you
suspect the vault key itself was exposed, you need a new DEK and a full re-encryption,
which is key *rotation* (chapter 10).

## Wrong password or corrupted file?

`Open` fails identically for a wrong password and for a tampered wrapped key, so Keybox
can't tell which happened. That's fine; the message "wrong password or corrupted vault"
is honest. Don't add a fast "password check" value to the header (such as an unsalted
hash of the password): it would let attackers test guesses without running the slow
KDF at all.

## Your task

Implement Keybox's key wrapping. `deriveKEK`, `newAEAD` and `newHeader` are provided.

1. `createVault(password, iterations)`: generate a random 32-byte data key and a fresh
   header, derive the KEK, and seal the data key with `newAEAD(kek)` using
   `[]byte(wrapAD)` as associated data. Return the file and the data key.
2. `unlockVault(password, f)`: derive the KEK from `f.Header` and open `f.WrappedKey`
   with the same associated data. Return `ErrWrongPassword` for any failure.
3. `changePassword(f, old, new, iterations)`: unlock with `old` (returning its error),
   then wrap the **same** data key under `new` with a **fresh** header.

A small `wrap(password, header, dataKey)` helper, used by both `createVault` and
`changePassword`, keeps things tidy. The tests use 1,000 iterations to stay fast.
