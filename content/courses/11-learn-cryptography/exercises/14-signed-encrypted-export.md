---
title: Signed, Sealed Export
difficulty: hard
after: digital-signatures
hints:
  - 'Seal first, then sign: `ct := hpke.Seal(pub, kdf, aead, info, plaintext)` with `info = infoPrefix || senderPub`, then sign `sigContext || sha256(recipientPub) || ct`. The file is `0x01 || sig || ct`. The sender''s public key is `sender.Public().(ed25519.PublicKey)`.'
  - 'In `openExport`, the recipient''s own public key bytes are `recipient.PublicKey().Bytes()`; you need them to rebuild the signed message. Check format, then sender key length, then the signature, and only then call `hpke.Open`. Tampered input then never reaches the decryptor at all.'
  - 'Why put the sender''s key in the HPKE info? Without it, anyone who intercepts the file can strip the signature and sign the same ciphertext themselves, and the recipient would credit them with secrets they can''t even read. With it, a ciphertext opens only under the key of the sender who made it.'
exercise:
  starter: |
    package main

    import (
    	"crypto/ed25519"
    	"crypto/hpke"
    	"errors"
    	"fmt"
    )

    var (
    	ErrBadFormat    = errors.New("keybox: not an export file")
    	ErrBadSignature = errors.New("keybox: export signature invalid")
    	ErrDecrypt      = errors.New("keybox: cannot decrypt export")
    )

    const (
    	formatV1   = 0x01
    	infoPrefix = "keybox export v1"     // HPKE info = infoPrefix || sender's Ed25519 public key
    	sigContext = "keybox export sig v1" // signed message = sigContext || SHA-256(recipient key) || ciphertext
    )

    var (
    	kem  = hpke.MLKEM768X25519()
    	kdf  = hpke.HKDFSHA256()
    	aead = hpke.AES256GCM()
    )

    // sealExport encrypts plaintext to recipientPub and signs the result as sender.
    func sealExport(sender ed25519.PrivateKey, recipientPub []byte, plaintext []byte) ([]byte, error) {
    	return nil, nil
    }

    // openExport verifies an export from senderPub and decrypts it with recipient.
    func openExport(recipient hpke.PrivateKey, senderPub ed25519.PublicKey, file []byte) ([]byte, error) {
    	return nil, ErrBadFormat
    }

    func main() {
    	alicePub, alicePriv, _ := ed25519.GenerateKey(nil)
    	bob, _ := kem.GenerateKey()
    	file, err := sealExport(alicePriv, bob.PublicKey().Bytes(), []byte("github: ghp_123"))
    	fmt.Println(len(file), err) // want: 1216 <nil>
    	pt, err := openExport(bob, alicePub, file)
    	fmt.Printf("%q %v\n", pt, err) // want: "github: ghp_123" <nil>
    	file[len(file)-1] ^= 1
    	_, err = openExport(bob, alicePub, file)
    	fmt.Println(err) // want: keybox: export signature invalid
    }
  solution: |
    package main

    import (
    	"crypto/ed25519"
    	"crypto/hpke"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    )

    var (
    	ErrBadFormat    = errors.New("keybox: not an export file")
    	ErrBadSignature = errors.New("keybox: export signature invalid")
    	ErrDecrypt      = errors.New("keybox: cannot decrypt export")
    )

    const (
    	formatV1   = 0x01
    	infoPrefix = "keybox export v1"     // HPKE info = infoPrefix || sender's Ed25519 public key
    	sigContext = "keybox export sig v1" // signed message = sigContext || SHA-256(recipient key) || ciphertext
    )

    var (
    	kem  = hpke.MLKEM768X25519()
    	kdf  = hpke.HKDFSHA256()
    	aead = hpke.AES256GCM()
    )

    func hpkeInfo(senderPub ed25519.PublicKey) []byte {
    	return append([]byte(infoPrefix), senderPub...)
    }

    func signedMessage(recipientPub, ciphertext []byte) []byte {
    	h := sha256.Sum256(recipientPub)
    	msg := append([]byte(sigContext), h[:]...)
    	return append(msg, ciphertext...)
    }

    // sealExport encrypts plaintext to recipientPub and signs the result as sender.
    func sealExport(sender ed25519.PrivateKey, recipientPub []byte, plaintext []byte) ([]byte, error) {
    	if len(sender) != ed25519.PrivateKeySize {
    		return nil, errors.New("keybox: bad signing key")
    	}
    	pub, err := kem.NewPublicKey(recipientPub)
    	if err != nil {
    		return nil, err
    	}
    	senderPub := sender.Public().(ed25519.PublicKey)
    	ct, err := hpke.Seal(pub, kdf, aead, hpkeInfo(senderPub), plaintext)
    	if err != nil {
    		return nil, err
    	}
    	sig := ed25519.Sign(sender, signedMessage(recipientPub, ct))
    	out := append([]byte{formatV1}, sig...)
    	return append(out, ct...), nil
    }

    // openExport verifies an export from senderPub and decrypts it with recipient.
    func openExport(recipient hpke.PrivateKey, senderPub ed25519.PublicKey, file []byte) ([]byte, error) {
    	if len(file) < 1+ed25519.SignatureSize+1 || file[0] != formatV1 {
    		return nil, ErrBadFormat
    	}
    	sig, ct := file[1:1+ed25519.SignatureSize], file[1+ed25519.SignatureSize:]
    	if len(senderPub) != ed25519.PublicKeySize {
    		return nil, ErrBadSignature
    	}
    	if !ed25519.Verify(senderPub, signedMessage(recipient.PublicKey().Bytes(), ct), sig) {
    		return nil, ErrBadSignature
    	}
    	pt, err := hpke.Open(recipient, kdf, aead, hpkeInfo(senderPub), ct)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func main() {
    	alicePub, alicePriv, _ := ed25519.GenerateKey(nil)
    	bob, _ := kem.GenerateKey()
    	file, err := sealExport(alicePriv, bob.PublicKey().Bytes(), []byte("github: ghp_123"))
    	fmt.Println(len(file), err)
    	pt, err := openExport(bob, alicePub, file)
    	fmt.Printf("%q %v\n", pt, err)
    	file[len(file)-1] ^= 1
    	_, err = openExport(bob, alicePub, file)
    	fmt.Println(err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/ed25519"
    	"crypto/hpke"
    	"crypto/sha256"
    	"errors"
    	"testing"
    )

    type party struct {
    	name string
    	sign ed25519.PrivateKey
    	pub  ed25519.PublicKey
    	kem  hpke.PrivateKey
    }

    func newParty(t *testing.T, name string, seed byte) party {
    	t.Helper()
    	sk := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{seed}, 32))
    	k, err := kem.DeriveKeyPair(bytes.Repeat([]byte{seed}, 32))
    	if err != nil {
    		t.Fatal(err)
    	}
    	return party{name, sk, sk.Public().(ed25519.PublicKey), k}
    }

    var secret = []byte(`{"github":"ghp_123","aws":"AKIA..."}`)

    func mustSeal(t *testing.T, from, to party) []byte {
    	t.Helper()
    	file, err := sealExport(from.sign, to.kem.PublicKey().Bytes(), secret)
    	if err != nil {
    		t.Fatalf("sealExport(%s -> %s) returned error %v", from.name, to.name, err)
    	}
    	if len(file) < 1+64+1 {
    		t.Fatalf("sealExport returned only %d bytes", len(file))
    	}
    	return file
    }

    func expectErr(t *testing.T, what string, to party, sender ed25519.PublicKey, file []byte, want error) {
    	t.Helper()
    	pt, err := openExport(to.kem, sender, file)
    	if !errors.Is(err, want) || pt != nil {
    		t.Errorf("%s: openExport = %q, %v, want nil, %v", what, pt, err, want)
    	}
    }

    func TestRoundTrip(t *testing.T) {
    	alice, bob := newParty(t, "alice", 1), newParty(t, "bob", 2)
    	file := mustSeal(t, alice, bob)
    	pt, err := openExport(bob.kem, alice.pub, file)
    	if err != nil || !bytes.Equal(pt, secret) {
    		t.Errorf("openExport(alice -> bob) = %q, %v, want the secret", pt, err)
    	}
    	empty, err := sealExport(alice.sign, bob.kem.PublicKey().Bytes(), nil)
    	if err != nil {
    		t.Fatal(err)
    	}
    	if pt, err := openExport(bob.kem, alice.pub, empty); err != nil || len(pt) != 0 {
    		t.Errorf("empty export: openExport = %q, %v", pt, err)
    	}
    }

    func TestExactFormat(t *testing.T) {
    	alice, bob := newParty(t, "alice", 1), newParty(t, "bob", 2)
    	file := mustSeal(t, alice, bob)
    	if file[0] != 0x01 {
    		t.Fatalf("first byte = %#x, want version 0x01", file[0])
    	}
    	sig, ct := file[1:65], file[65:]
    	rh := sha256.Sum256(bob.kem.PublicKey().Bytes())
    	msg := append(append([]byte("keybox export sig v1"), rh[:]...), ct...)
    	if !ed25519.Verify(alice.pub, msg, sig) {
    		t.Fatalf("bytes 1-64 aren't an Ed25519 signature over sigContext || SHA-256(recipient key) || ciphertext")
    	}
    	info := append([]byte("keybox export v1"), alice.pub...)
    	pt, err := hpke.Open(bob.kem, kdf, aead, info, ct)
    	if err != nil || !bytes.Equal(pt, secret) {
    		t.Fatalf("ciphertext doesn't open with hpke.Open and info = infoPrefix || sender key: %v", err)
    	}
    }

    func TestVerifyBeforeDecrypt(t *testing.T) {
    	alice, bob := newParty(t, "alice", 1), newParty(t, "bob", 2)
    	file := mustSeal(t, alice, bob)
    	// Tampering anywhere after the version byte must be caught by the
    	// signature, before any decryption is attempted.
    	for i := 1; i < len(file); i++ {
    		bad := bytes.Clone(file)
    		bad[i] ^= 1 << (i % 8)
    		pt, err := openExport(bob.kem, alice.pub, bad)
    		if !errors.Is(err, ErrBadSignature) || pt != nil {
    			t.Fatalf("byte %d modified: openExport = %q, %v, want nil, ErrBadSignature (verify first, then decrypt)", i, pt, err)
    		}
    	}
    	for _, v := range []byte{0x00, 0x02, 0xff} {
    		bad := bytes.Clone(file)
    		bad[0] = v
    		expectErr(t, "unknown version byte", bob, alice.pub, bad, ErrBadFormat)
    	}
    }

    func TestWrongSenderOrRecipient(t *testing.T) {
    	alice, bob, carol := newParty(t, "alice", 1), newParty(t, "bob", 2), newParty(t, "carol", 3)
    	file := mustSeal(t, alice, bob)
    	expectErr(t, "verified against carol's key instead of alice's", bob, carol.pub, file, ErrBadSignature)
    	toCarol := mustSeal(t, alice, carol)
    	expectErr(t, "bob opening an export alice addressed to carol", bob, alice.pub, toCarol, ErrBadSignature)
    	for _, bad := range []ed25519.PublicKey{nil, alice.pub[:31], append(bytes.Clone(alice.pub), 0)} {
    		func() {
    			defer func() {
    				if r := recover(); r != nil {
    					t.Errorf("openExport panicked on a %d-byte sender key: %v", len(bad), r)
    				}
    			}()
    			expectErr(t, "malformed sender key", bob, bad, file, ErrBadSignature)
    		}()
    	}
    }

    func TestResignedByMallory(t *testing.T) {
    	// Mallory intercepts alice's export to bob, strips alice's signature and
    	// signs the same ciphertext herself, so bob thinks the secrets came from
    	// her. The signature is valid, but the ciphertext is bound to alice's key.
    	alice, bob, mallory := newParty(t, "alice", 1), newParty(t, "bob", 2), newParty(t, "mallory", 4)
    	file := mustSeal(t, alice, bob)
    	ct := file[65:]
    	rh := sha256.Sum256(bob.kem.PublicKey().Bytes())
    	msg := append(append([]byte("keybox export sig v1"), rh[:]...), ct...)
    	resigned := append([]byte{0x01}, ed25519.Sign(mallory.sign, msg)...)
    	resigned = append(resigned, ct...)
    	expectErr(t, "alice's ciphertext re-signed by mallory", bob, mallory.pub, resigned, ErrDecrypt)
    }

    func TestTruncated(t *testing.T) {
    	alice, bob := newParty(t, "alice", 1), newParty(t, "bob", 2)
    	file := mustSeal(t, alice, bob)
    	for n := 0; n < len(file); n += 1 + n/8 {
    		pt, err := openExport(bob.kem, alice.pub, file[:n])
    		if err == nil || pt != nil {
    			t.Fatalf("file cut to %d of %d bytes: openExport = %q, %v, want an error", n, len(file), pt, err)
    		}
    		if n <= 65 && !errors.Is(err, ErrBadFormat) {
    			t.Fatalf("file cut to %d bytes (no room for version, signature and ciphertext): err = %v, want ErrBadFormat", n, err)
    		}
    	}
    }

    func TestSealRejectsBadRecipientKey(t *testing.T) {
    	alice, bob := newParty(t, "alice", 1), newParty(t, "bob", 2)
    	pub := bob.kem.PublicKey().Bytes()
    	for _, bad := range [][]byte{nil, pub[:100], append(bytes.Clone(pub), 0)} {
    		if f, err := sealExport(alice.sign, bad, secret); err == nil {
    			t.Errorf("sealExport to a %d-byte recipient key returned %d bytes and no error", len(bad), len(f))
    		}
    	}
    }
---

Alice wants to hand Bob a bundle of team secrets as a file, sent over email or chat,
channels neither of them trusts. Bob must be sure of two things: only he can read it,
and it really came from Alice. Keybox combines **HPKE** (to encrypt to Bob's public
key) with an **Ed25519 signature** (Alice's identity key), and Bob checks the signature
**before** decrypting anything.

Implement:

- `sealExport(sender, recipientPub, plaintext)`:
  1. parse `recipientPub` with `kem.NewPublicKey` (return its error),
  2. `ct = hpke.Seal(...)` with Keybox's suite and
     `info = infoPrefix || sender's Ed25519 public key (32 bytes)`,
  3. `sig = ed25519.Sign(sender, sigContext || SHA-256(recipientPub) || ct)`,
  4. return `0x01 || sig (64 bytes) || ct`.
- `openExport(recipient, senderPub, file)` returns the plaintext, or:
  - `ErrBadFormat` if `file` is 65 bytes or shorter, or doesn't start with `0x01`;
  - `ErrBadSignature` if `senderPub` isn't 32 bytes (no panics) or the signature
    doesn't verify over `sigContext || SHA-256(recipient's public key) || ct`;
  - `ErrDecrypt` if the signature is fine but HPKE decryption fails.

## Example

```go
file, _ := sealExport(alicePriv, bobPubBytes, secrets)  // 1 + 64 + 1151 bytes for 15 bytes of secrets
openExport(bob, alicePub, file)       // secrets, nil
openExport(bob, carolPub, file)       // nil, ErrBadSignature (not from carol)
openExport(bob, alicePub, toCarol)    // nil, ErrBadSignature (addressed to someone else)
```

## Constraints

- The tests check the layout with `ed25519.Verify` and `hpke.Open` directly.
- Every modified byte after the version must give `ErrBadSignature`, which only
  happens if you verify before decrypting. They also try unknown versions, wrong and
  malformed sender keys, a file addressed to someone else, truncations and bad
  recipient keys.
- One test takes Alice's file to Bob and re-signs the ciphertext with Mallory's key.
  The signature is valid, but `openExport(bob, malloryPub, …)` must fail with
  `ErrDecrypt`, because the ciphertext is bound to Alice's key through the HPKE info.
