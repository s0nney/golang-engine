---
title: Share with Every Device
difficulty: medium
after: public-key-cryptography
hints:
  - 'Generate one fresh 32-byte **data key** with `crypto/rand`. Seal the secret once with AES-256-GCM under it (`cipher.NewGCMWithRandomNonce`, `Seal(nil, nil, secret, nil)`), then seal the *data key* to each device with `hpke.Seal(pub, kdf, aead, []byte(wrapInfo), dataKey)`. Parse each device''s bytes with `kem.NewPublicKey` and return its error.'
  - '`openShare` is the same in reverse: look up `s.Wrapped[name]` (missing means `ErrNotRecipient`), `hpke.Open` it to get the data key, then open `s.Body`. Turn every failure after the lookup into `ErrDecrypt`.'
exercise:
  starter: |
    package main

    import (
    	"crypto/hpke"
    	"errors"
    	"fmt"
    )

    var (
    	ErrDecrypt      = errors.New("keybox: cannot open share")
    	ErrNotRecipient = errors.New("keybox: share is not addressed to this device")
    )

    // Keybox's HPKE ciphersuite and the info string for wrapping data keys.
    const wrapInfo = "keybox multi-device share v1"

    var (
    	kem  = hpke.MLKEM768X25519()
    	kdf  = hpke.HKDFSHA256()
    	aead = hpke.AES256GCM()
    )

    // Share is one secret shared with several devices.
    type Share struct {
    	Body    []byte            // the secret, sealed once under a fresh data key
    	Wrapped map[string][]byte // device name -> the data key, sealed to that device
    }

    // shareToDevices seals secret so that every device in devices can open it.
    func shareToDevices(devices map[string][]byte, secret []byte) (Share, error) {
    	return Share{}, nil
    }

    // openShare opens s as the device called name, holding priv.
    func openShare(name string, priv hpke.PrivateKey, s Share) ([]byte, error) {
    	return nil, ErrDecrypt
    }

    func main() {
    	laptop, _ := kem.GenerateKey()
    	phone, _ := kem.GenerateKey()
    	devices := map[string][]byte{
    		"laptop": laptop.PublicKey().Bytes(),
    		"phone":  phone.PublicKey().Bytes(),
    	}
    	s, err := shareToDevices(devices, []byte("wifi password: hunter2"))
    	fmt.Println("wrapped for", len(s.Wrapped), "devices, err:", err) // want: 2 devices
    	pt, err := openShare("phone", phone, s)
    	fmt.Printf("phone opens: %q %v\n", pt, err) // want: "wifi password: hunter2" <nil>
    }
  solution: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/hpke"
    	"crypto/rand"
    	"errors"
    	"fmt"
    )

    var (
    	ErrDecrypt      = errors.New("keybox: cannot open share")
    	ErrNotRecipient = errors.New("keybox: share is not addressed to this device")
    )

    // Keybox's HPKE ciphersuite and the info string for wrapping data keys.
    const wrapInfo = "keybox multi-device share v1"

    var (
    	kem  = hpke.MLKEM768X25519()
    	kdf  = hpke.HKDFSHA256()
    	aead = hpke.AES256GCM()
    )

    // Share is one secret shared with several devices.
    type Share struct {
    	Body    []byte            // the secret, sealed once under a fresh data key
    	Wrapped map[string][]byte // device name -> the data key, sealed to that device
    }

    func newGCM(key []byte) (cipher.AEAD, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	return cipher.NewGCMWithRandomNonce(block)
    }

    // shareToDevices seals secret so that every device in devices can open it.
    func shareToDevices(devices map[string][]byte, secret []byte) (Share, error) {
    	if len(devices) == 0 {
    		return Share{}, errors.New("keybox: no devices to share with")
    	}
    	dataKey := make([]byte, 32)
    	rand.Read(dataKey)
    	s := Share{Wrapped: make(map[string][]byte, len(devices))}
    	for name, pubBytes := range devices {
    		pub, err := kem.NewPublicKey(pubBytes)
    		if err != nil {
    			return Share{}, fmt.Errorf("keybox: public key for %q: %w", name, err)
    		}
    		wrapped, err := hpke.Seal(pub, kdf, aead, []byte(wrapInfo), dataKey)
    		if err != nil {
    			return Share{}, err
    		}
    		s.Wrapped[name] = wrapped
    	}
    	gcm, err := newGCM(dataKey)
    	if err != nil {
    		return Share{}, err
    	}
    	s.Body = gcm.Seal(nil, nil, secret, nil)
    	return s, nil
    }

    // openShare opens s as the device called name, holding priv.
    func openShare(name string, priv hpke.PrivateKey, s Share) ([]byte, error) {
    	wrapped, ok := s.Wrapped[name]
    	if !ok {
    		return nil, ErrNotRecipient
    	}
    	dataKey, err := hpke.Open(priv, kdf, aead, []byte(wrapInfo), wrapped)
    	if err != nil || len(dataKey) != 32 {
    		return nil, ErrDecrypt
    	}
    	gcm, err := newGCM(dataKey)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	pt, err := gcm.Open(nil, nil, s.Body, nil)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func main() {
    	laptop, _ := kem.GenerateKey()
    	phone, _ := kem.GenerateKey()
    	devices := map[string][]byte{
    		"laptop": laptop.PublicKey().Bytes(),
    		"phone":  phone.PublicKey().Bytes(),
    	}
    	s, err := shareToDevices(devices, []byte("wifi password: hunter2"))
    	fmt.Println("wrapped for", len(s.Wrapped), "devices, err:", err)
    	pt, err := openShare("phone", phone, s)
    	fmt.Printf("phone opens: %q %v\n", pt, err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/hpke"
    	"errors"
    	"maps"
    	"slices"
    	"testing"
    )

    // device derives a deterministic key pair for a test device.
    func device(t *testing.T, seed byte) hpke.PrivateKey {
    	t.Helper()
    	k, err := kem.DeriveKeyPair(bytes.Repeat([]byte{seed}, 32))
    	if err != nil {
    		t.Fatal(err)
    	}
    	return k
    }

    type fleet struct {
    	privs map[string]hpke.PrivateKey
    	pubs  map[string][]byte
    }

    func newFleet(t *testing.T) fleet {
    	f := fleet{map[string]hpke.PrivateKey{}, map[string][]byte{}}
    	for i, name := range []string{"laptop", "phone", "tablet"} {
    		f.privs[name] = device(t, byte(i+1))
    		f.pubs[name] = f.privs[name].PublicKey().Bytes()
    	}
    	return f
    }

    var secret = []byte("db password: correct-horse-battery-staple")

    func mustShare(t *testing.T, pubs map[string][]byte) Share {
    	t.Helper()
    	s, err := shareToDevices(pubs, secret)
    	if err != nil {
    		t.Fatalf("shareToDevices returned error %v", err)
    	}
    	if got, want := slices.Sorted(maps.Keys(s.Wrapped)), slices.Sorted(maps.Keys(pubs)); !slices.Equal(got, want) {
    		t.Fatalf("Share.Wrapped has entries for %v, want exactly %v", got, want)
    	}
    	return s
    }

    func expectFail(t *testing.T, what string, name string, priv hpke.PrivateKey, s Share, want error) {
    	t.Helper()
    	pt, err := openShare(name, priv, s)
    	if !errors.Is(err, want) || pt != nil {
    		t.Errorf("%s: openShare = %q, %v, want nil, %v", what, pt, err, want)
    	}
    }

    func TestEveryDeviceOpens(t *testing.T) {
    	f := newFleet(t)
    	s := mustShare(t, f.pubs)
    	for name, priv := range f.privs {
    		pt, err := openShare(name, priv, s)
    		if err != nil || !bytes.Equal(pt, secret) {
    			t.Errorf("openShare as %s = %q, %v, want the secret", name, pt, err)
    		}
    	}
    }

    func TestFormat(t *testing.T) {
    	// Each wrapped entry is HPKE (info wrapInfo) of a 32-byte data key, and
    	// Body is AES-256-GCM (random nonce || ciphertext || tag) under that key.
    	f := newFleet(t)
    	s := mustShare(t, f.pubs)
    	var first []byte
    	for name, priv := range f.privs {
    		dk, err := hpke.Open(priv, kdf, aead, []byte(wrapInfo), s.Wrapped[name])
    		if err != nil || len(dk) != 32 {
    			t.Fatalf("Wrapped[%q] doesn't open as an HPKE-sealed 32-byte key with info %q: %d bytes, %v", name, wrapInfo, len(dk), err)
    		}
    		if first == nil {
    			first = dk
    		} else if !bytes.Equal(dk, first) {
    			t.Fatalf("devices received different data keys: the body is sealed once, so every device needs the same key")
    		}
    	}
    	block, _ := aes.NewCipher(first)
    	gcm, _ := cipher.NewGCMWithRandomNonce(block)
    	if pt, err := gcm.Open(nil, nil, s.Body, nil); err != nil || !bytes.Equal(pt, secret) {
    		t.Errorf("Body doesn't open as nonce||ciphertext||tag under the data key: %v", err)
    	}
    	if bytes.Contains(s.Body, secret) {
    		t.Errorf("Body contains the secret in plaintext")
    	}
    }

    func TestFreshDataKeys(t *testing.T) {
    	f := newFleet(t)
    	a, b := mustShare(t, f.pubs), mustShare(t, f.pubs)
    	if bytes.Equal(a.Body, b.Body) {
    		t.Errorf("sharing the same secret twice gave identical bodies")
    	}
    	// A body from one share must not open with the key from another.
    	mixed := Share{Body: b.Body, Wrapped: a.Wrapped}
    	expectFail(t, "body swapped in from another share", "laptop", f.privs["laptop"], mixed, ErrDecrypt)
    }

    func TestOutsidersAndWrongKeys(t *testing.T) {
    	f := newFleet(t)
    	s := mustShare(t, f.pubs)
    	outsider := device(t, 99)
    	expectFail(t, "device not in the share", "desktop", outsider, s, ErrNotRecipient)
    	expectFail(t, "outsider claiming to be the laptop", "laptop", outsider, s, ErrDecrypt)
    	expectFail(t, "phone's key used for the laptop's entry", "laptop", f.privs["phone"], s, ErrDecrypt)
    }

    func TestTampering(t *testing.T) {
    	f := newFleet(t)
    	s := mustShare(t, f.pubs)
    	laptop := f.privs["laptop"]
    	clone := func() Share {
    		return Share{Body: bytes.Clone(s.Body), Wrapped: maps.Clone(s.Wrapped)}
    	}
    	for i := range s.Body {
    		c := clone()
    		c.Body[i] ^= 0x01
    		if pt, err := openShare("laptop", laptop, c); !errors.Is(err, ErrDecrypt) || pt != nil {
    			t.Fatalf("Body byte %d flipped: openShare = %q, %v, want nil, ErrDecrypt", i, pt, err)
    		}
    	}
    	w := s.Wrapped["laptop"]
    	for i := 0; i < len(w); i += 7 {
    		c := clone()
    		c.Wrapped["laptop"] = bytes.Clone(w)
    		c.Wrapped["laptop"][i] ^= 0x80
    		if pt, err := openShare("laptop", laptop, c); !errors.Is(err, ErrDecrypt) || pt != nil {
    			t.Fatalf("wrapped key byte %d flipped: openShare = %q, %v, want nil, ErrDecrypt", i, pt, err)
    		}
    	}
    	for _, n := range []int{0, 1, 16, len(w) - 1} {
    		c := clone()
    		c.Wrapped["laptop"] = w[:n]
    		expectFail(t, "wrapped key truncated", "laptop", laptop, c, ErrDecrypt)
    	}
    	for _, n := range []int{0, 11, 12, 27, len(s.Body) - 1} {
    		c := clone()
    		c.Body = s.Body[:n]
    		expectFail(t, "body truncated", "laptop", laptop, c, ErrDecrypt)
    	}
    }

    func TestBadPublicKeys(t *testing.T) {
    	f := newFleet(t)
    	for _, bad := range [][]byte{nil, make([]byte, 10), f.pubs["phone"][:len(f.pubs["phone"])-1]} {
    		pubs := maps.Clone(f.pubs)
    		pubs["broken"] = bad
    		if _, err := shareToDevices(pubs, secret); err == nil {
    			t.Errorf("shareToDevices accepted a %d-byte public key", len(bad))
    		}
    	}
    	if _, err := shareToDevices(nil, secret); err == nil {
    		t.Errorf("shareToDevices with no devices returned no error: nobody could ever open it")
    	}
    }
---

Keybox users often have three or four devices. When Alice shares the office Wi-Fi
password with Bob, every one of Bob's devices must be able to read it, but encrypting
the (possibly large) secret separately for each device is wasteful. The standard
answer is **key wrapping**: encrypt the secret once under a random data key, then
encrypt only that small key to each device's public key.

Implement:

- `shareToDevices(devices, secret)`: `devices` maps a device name to its HPKE public
  key bytes (Keybox's suite: `MLKEM768X25519`, `HKDFSHA256`, `AES256GCM`, given as
  `kem`, `kdf` and `aead`). Return a `Share` whose
  - `Body` is `secret` sealed with AES-256-GCM under a **fresh random 32-byte data
    key**, as `nonce || ciphertext || tag` with no associated data, and
  - `Wrapped[name]` is the data key sealed to that device with `hpke.Seal`, using
    `wrapInfo` as the info.

  Return an error if `devices` is empty or any public key doesn't parse.
- `openShare(name, priv, s)` returns the secret. It returns `ErrNotRecipient` if `s`
  has no entry for `name`, and `ErrDecrypt` for any other failure: wrong key, tampered
  or truncated body or wrapped key.

## Example

```go
s, _ := shareToDevices(map[string][]byte{"laptop": laptopPub, "phone": phonePub}, secret)
openShare("phone", phonePriv, s)      // secret, nil
openShare("desktop", desktopPriv, s)  // nil, ErrNotRecipient
openShare("laptop", phonePriv, s)     // nil, ErrDecrypt
```

## Constraints

- The tests open your `Wrapped` entries and `Body` with their own code, so follow the
  format exactly.
- They check that each share uses a new data key, that bodies can't be swapped between
  shares, and that flipped or truncated bytes in the body or a wrapped key are
  rejected, as are malformed public keys.
- HPKE with `MLKEM768X25519` stays secure if *either* X25519 or ML-KEM holds up,
  including against a future quantum computer.
