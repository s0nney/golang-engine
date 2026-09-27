---
title: Which Device Signed It?
difficulty: easy
after: digital-signatures
hints:
  - 'Loop over `devices` in order. `continue` past a device that is `Revoked` or whose `len(d.Key) != ed25519.PublicKeySize`, and return `d.Name, true` for the first one where `ed25519.Verify(d.Key, msg, sig)` is true.'
  - '`ed25519.Verify` panics on a public key of the wrong length (it treats that as a programming error). Keys that arrive over the network aren''t your program''s invariants, so check the length yourself first. A wrong-length *signature* just returns false.'
exercise:
  starter: |
    package main

    import (
    	"crypto/ed25519"
    	"fmt"
    )

    // Device is one of a user's enrolled devices.
    type Device struct {
    	Name    string
    	Key     ed25519.PublicKey // should be 32 bytes, but it came from a sync server
    	Revoked bool              // a lost or retired device
    }

    // whoSigned returns the name of the first non-revoked device in devices
    // whose key verifies sig over msg, and true. If none does, it returns "", false.
    // A key of the wrong length can't verify anything: skip it (ed25519.Verify
    // would panic on it).
    func whoSigned(devices []Device, msg, sig []byte) (string, bool) {
    	// For each device: skip revoked ones and keys that aren't
    	// ed25519.PublicKeySize bytes, then try ed25519.Verify.
    	return "", false
    }

    func main() {
    	pub, priv, _ := ed25519.GenerateKey(nil)
    	devices := []Device{
    		{Name: "old phone", Key: make([]byte, 31)}, // corrupted entry
    		{Name: "laptop", Key: pub},
    	}
    	sig := ed25519.Sign(priv, []byte("add secret: github-token"))
    	fmt.Println(whoSigned(devices, []byte("add secret: github-token"), sig)) // want: laptop true
    }
  solution: |
    package main

    import (
    	"crypto/ed25519"
    	"fmt"
    )

    // Device is one of a user's enrolled devices.
    type Device struct {
    	Name    string
    	Key     ed25519.PublicKey
    	Revoked bool
    }

    func whoSigned(devices []Device, msg, sig []byte) (string, bool) {
    	for _, d := range devices {
    		if d.Revoked || len(d.Key) != ed25519.PublicKeySize {
    			continue
    		}
    		if ed25519.Verify(d.Key, msg, sig) {
    			return d.Name, true
    		}
    	}
    	return "", false
    }

    func main() {
    	pub, priv, _ := ed25519.GenerateKey(nil)
    	devices := []Device{
    		{Name: "old phone", Key: make([]byte, 31)},
    		{Name: "laptop", Key: pub},
    	}
    	sig := ed25519.Sign(priv, []byte("add secret: github-token"))
    	fmt.Println(whoSigned(devices, []byte("add secret: github-token"), sig))
    }
  tests: |
    package main

    import (
    	"bytes"
    	"encoding/hex"
    	"testing"
    )

    func unhex(s string) []byte {
    	b, err := hex.DecodeString(s)
    	if err != nil {
    		panic(err)
    	}
    	return b
    }

    // RFC 8032, section 7.1, tests 1-3: public key, message, signature.
    var (
    	pub1 = unhex("d75a980182b10ab7d54bfed3c964073a0ee172f3daa62325af021a68f707511a")
    	msg1 = []byte{}
    	sig1 = unhex("e5564300c360ac729086e2cc806e828a84877f1eb8e5d974d873e065224901555fb8821590a33bacc61e39701cf9b46bd25bf5f0595bbe24655141438e7a100b")
    	pub2 = unhex("3d4017c3e843895a92b70aa74d1b7ebc9c982ccf2ec4968cc0cd55f12af4660c")
    	msg2 = unhex("72")
    	sig2 = unhex("92a009a9f0d4cab8720e820b5f642540a2b27b5416503f8fb3762223ebdb69da085ac1e43e15996e458f3613d0f11d8c387b2eaeb4302aeeb00d291612bb0c00")
    	pub3 = unhex("fc51cd8e6218a1a38da47ed00230f0580816ed13ba3303ac5deb911548908025")
    	msg3 = unhex("af82")
    	sig3 = unhex("6291d657deec24024827e69c3abe01a30ce548a284743a445e3680d7db5ac3ac18ff9b538d16f290ae67f760984dc6594a7c15e9716ed28dc027beceea1ec40a")
    )

    func devices() []Device {
    	return []Device{
    		{Name: "phone", Key: pub1},
    		{Name: "laptop", Key: pub2},
    		{Name: "tablet", Key: pub3},
    	}
    }

    // safeWhoSigned reports a panic as a test failure.
    func safeWhoSigned(t *testing.T, ds []Device, msg, sig []byte) (name string, ok bool) {
    	t.Helper()
    	defer func() {
    		if r := recover(); r != nil {
    			t.Errorf("whoSigned panicked: %v", r)
    		}
    	}()
    	return whoSigned(ds, msg, sig)
    }

    func TestFindsSigner(t *testing.T) {
    	tests := []struct {
    		msg, sig []byte
    		want     string
    	}{{msg1, sig1, "phone"}, {msg2, sig2, "laptop"}, {msg3, sig3, "tablet"}}
    	for _, tt := range tests {
    		if got, ok := safeWhoSigned(t, devices(), tt.msg, tt.sig); got != tt.want || !ok {
    			t.Errorf("whoSigned(RFC 8032 message %x) = %q, %v, want %q, true", tt.msg, got, ok, tt.want)
    		}
    	}
    }

    func TestRejectsMismatches(t *testing.T) {
    	ds := devices()
    	if got, ok := safeWhoSigned(t, ds, msg3, sig2); ok {
    		t.Errorf("signature 2 over message 3 was attributed to %q, want no signer", got)
    	}
    	if got, ok := safeWhoSigned(t, ds, []byte{0x73}, sig2); ok {
    		t.Errorf("signature over 0x72 accepted for message 0x73 (signed by %q)", got)
    	}
    	for i := range sig2 {
    		bad := bytes.Clone(sig2)
    		bad[i] ^= 0x10
    		if got, ok := safeWhoSigned(t, ds, msg2, bad); ok {
    			t.Fatalf("signature with byte %d modified was accepted as %q", i, got)
    		}
    	}
    	for _, bad := range [][]byte{nil, sig2[:63], append(bytes.Clone(sig2), 0)} {
    		if got, ok := safeWhoSigned(t, ds, msg2, bad); ok {
    			t.Errorf("%d-byte signature accepted as %q", len(bad), got)
    		}
    	}
    	if got, ok := safeWhoSigned(t, nil, msg2, sig2); ok {
    		t.Errorf("whoSigned with no devices = %q, true, want \"\", false", got)
    	}
    }

    func TestSkipsRevoked(t *testing.T) {
    	ds := devices()
    	ds[1].Revoked = true
    	if got, ok := safeWhoSigned(t, ds, msg2, sig2); ok {
    		t.Errorf("signature from the revoked laptop was accepted as %q", got)
    	}
    	// The same key enrolled again as a new device counts.
    	ds = append(ds, Device{Name: "laptop (re-enrolled)", Key: pub2})
    	if got, ok := safeWhoSigned(t, ds, msg2, sig2); got != "laptop (re-enrolled)" || !ok {
    		t.Errorf("whoSigned = %q, %v, want %q, true", got, ok, "laptop (re-enrolled)")
    	}
    }

    func TestMalformedKeysDontPanic(t *testing.T) {
    	ds := []Device{
    		{Name: "empty", Key: nil},
    		{Name: "short", Key: pub2[:31]},
    		{Name: "long", Key: append(bytes.Clone(pub2), 0)},
    		{Name: "laptop", Key: pub2},
    	}
    	if got, ok := safeWhoSigned(t, ds, msg2, sig2); got != "laptop" || !ok {
    		t.Errorf("with malformed keys listed first, whoSigned = %q, %v, want %q, true", got, ok, "laptop")
    	}
    }
---

Every change to a Keybox vault is signed by the device that made it. When a change
arrives through the sync server, the other devices look up which enrolled device
signed it, both to accept it and to show "Added by laptop" in the history.

Complete `whoSigned(devices, msg, sig)`. It returns the `Name` of the **first**
device in `devices` that

- isn't `Revoked`, and
- has a `Key` whose Ed25519 signature check passes for `sig` over `msg`,

together with `true`. If no device qualifies, return `"", false`.

The device list comes from the sync server, so a `Key` may be malformed (empty, too
short or too long). Such a device can't have signed anything: skip it, and **never
panic**.

## Example

Using public keys and signatures from RFC 8032's Ed25519 test vectors:

```go
devices := []Device{{"phone", pub1, false}, {"laptop", pub2, false}}
whoSigned(devices, []byte{0x72}, sig2)  // "laptop", true
whoSigned(devices, []byte{0x73}, sig2)  // "", false (different message)
devices[1].Revoked = true
whoSigned(devices, []byte{0x72}, sig2)  // "", false (revoked)
```

## Constraints

- The tests use RFC 8032 section 7.1, tests 1 to 3, plus every modified signature
  byte, signatures of the wrong length, and keys of length 0, 31 and 33.
- A valid signature proves the holder of that private key signed these exact bytes.
  It says nothing about whether the device should still be trusted, which is why
  revocation is checked separately.
