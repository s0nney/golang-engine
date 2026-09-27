---
title: Migrate the Envelope
difficulty: hard
after: pitfalls-and-capstone
hints:
  - 'Store one `cipher.AEAD` per key ID (from `cipher.NewGCMWithRandomNonce`) in a map. `Seal` builds the 5-byte header with `binary.BigEndian.AppendUint32([]byte{0x02}, id)` and calls `aead.Seal(header, nil, plaintext, header+context)`, so nonce, ciphertext and tag land right after the header.'
  - '`Open` switches on the first byte. v1: the rest is `nonce || ct || tag` under key 1 with **nil** associated data (refused once legacy is disabled). v2: read the key ID, look it up (`ErrUnknownKey`), and open `blob[5:]` with `blob[:5] + context` as associated data. Check lengths before slicing, and map every AEAD failure to `ErrDecrypt`.'
  - 'Why can''t an attacker strip a v2 header and present the rest as v1? Because every v2 blob was sealed with non-empty associated data, and GCM refuses to open it with none. `Upgrade` is simply `Open`, then (unless it''s already v2 under the current key) `Seal` under the current key with the same context.'
exercise:
  starter: |
    package main

    import (
    	"bytes"
    	"errors"
    	"fmt"
    )

    var (
    	ErrUnsupportedVersion = errors.New("keybox: unsupported ciphertext version")
    	ErrUnknownKey         = errors.New("keybox: unknown key")
    	ErrDecrypt            = errors.New("keybox: decryption failed")
    	ErrNoCurrentKey       = errors.New("keybox: no current key")
    )

    const (
    	v1          = 0x01
    	v2          = 0x02
    	legacyKeyID = 1 // v1 ciphertexts were all made with the key now registered as ID 1
    )

    // Envelope seals vault records in format v2 and still opens old v1 records.
    type Envelope struct {
    	// your fields here
    }

    func NewEnvelope() *Envelope {
    	return &Envelope{}
    }

    func (e *Envelope) AddKey(id uint32, key []byte) error { return nil }

    func (e *Envelope) SetCurrent(id uint32) error { return nil }

    func (e *Envelope) DisableLegacy() {}

    func (e *Envelope) Seal(plaintext []byte, context string) ([]byte, error) {
    	return nil, ErrNoCurrentKey
    }

    func (e *Envelope) Open(blob []byte, context string) ([]byte, error) {
    	return nil, ErrDecrypt
    }

    func (e *Envelope) Upgrade(blob []byte, context string) ([]byte, bool, error) {
    	return nil, false, ErrDecrypt
    }

    func main() {
    	e := NewEnvelope()
    	fmt.Println(e.AddKey(1, bytes.Repeat([]byte{1}, 32)), e.AddKey(2, bytes.Repeat([]byte{2}, 32)))
    	fmt.Println(e.SetCurrent(2))
    	blob, err := e.Seal([]byte("ghp_123"), "vault-7/github")
    	fmt.Printf("%x %v\n", blob[:min(5, len(blob))], err) // want: 0200000002 <nil>
    	pt, err := e.Open(blob, "vault-7/github")
    	fmt.Printf("%q %v\n", pt, err) // want: "ghp_123" <nil>
    	_, err = e.Open(blob, "vault-7/aws")
    	fmt.Println(err) // want: keybox: decryption failed
    }
  solution: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"encoding/binary"
    	"errors"
    	"fmt"
    )

    var (
    	ErrUnsupportedVersion = errors.New("keybox: unsupported ciphertext version")
    	ErrUnknownKey         = errors.New("keybox: unknown key")
    	ErrDecrypt            = errors.New("keybox: decryption failed")
    	ErrNoCurrentKey       = errors.New("keybox: no current key")
    )

    const (
    	v1          = 0x01
    	v2          = 0x02
    	legacyKeyID = 1 // v1 ciphertexts were all made with the key now registered as ID 1
    	v2Header    = 1 + 4
    	gcmOverhead = 12 + 16
    )

    // Envelope seals vault records in format v2 and still opens old v1 records.
    type Envelope struct {
    	keys         map[uint32]cipher.AEAD
    	current      uint32
    	legacyClosed bool
    }

    func NewEnvelope() *Envelope {
    	return &Envelope{keys: map[uint32]cipher.AEAD{}}
    }

    func (e *Envelope) AddKey(id uint32, key []byte) error {
    	if id == 0 {
    		return errors.New("keybox: key ID 0 is reserved")
    	}
    	if _, ok := e.keys[id]; ok {
    		return fmt.Errorf("keybox: key %d already registered", id)
    	}
    	if len(key) != 32 {
    		return errors.New("keybox: keys must be 32 bytes")
    	}
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return err
    	}
    	aead, err := cipher.NewGCMWithRandomNonce(block)
    	if err != nil {
    		return err
    	}
    	e.keys[id] = aead
    	return nil
    }

    func (e *Envelope) SetCurrent(id uint32) error {
    	if _, ok := e.keys[id]; !ok {
    		return ErrUnknownKey
    	}
    	e.current = id
    	return nil
    }

    func (e *Envelope) DisableLegacy() { e.legacyClosed = true }

    // v2AD is the associated data for a v2 blob: its 5-byte header, then context.
    func v2AD(header []byte, context string) []byte {
    	return append(bytes.Clone(header), context...)
    }

    func (e *Envelope) Seal(plaintext []byte, context string) ([]byte, error) {
    	aead, ok := e.keys[e.current]
    	if !ok {
    		return nil, ErrNoCurrentKey
    	}
    	header := binary.BigEndian.AppendUint32([]byte{v2}, e.current)
    	return aead.Seal(header, nil, plaintext, v2AD(header, context)), nil
    }

    func (e *Envelope) Open(blob []byte, context string) ([]byte, error) {
    	if len(blob) == 0 {
    		return nil, ErrUnsupportedVersion
    	}
    	var (
    		id     uint32
    		body   []byte
    		ad     []byte
    		minLen int
    	)
    	switch blob[0] {
    	case v1:
    		if e.legacyClosed {
    			return nil, ErrUnsupportedVersion
    		}
    		id, body, minLen = legacyKeyID, blob[1:], 1+gcmOverhead
    	case v2:
    		if len(blob) < v2Header {
    			return nil, ErrDecrypt
    		}
    		id, body, minLen = binary.BigEndian.Uint32(blob[1:v2Header]), blob[v2Header:], v2Header+gcmOverhead
    		ad = v2AD(blob[:v2Header], context)
    	default:
    		return nil, ErrUnsupportedVersion
    	}
    	if len(blob) < minLen {
    		return nil, ErrDecrypt
    	}
    	aead, ok := e.keys[id]
    	if !ok {
    		return nil, ErrUnknownKey
    	}
    	pt, err := aead.Open(nil, nil, body, ad)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func (e *Envelope) Upgrade(blob []byte, context string) ([]byte, bool, error) {
    	pt, err := e.Open(blob, context)
    	if err != nil {
    		return nil, false, err
    	}
    	if blob[0] == v2 && binary.BigEndian.Uint32(blob[1:v2Header]) == e.current {
    		return blob, false, nil
    	}
    	out, err := e.Seal(pt, context)
    	if err != nil {
    		return nil, false, err
    	}
    	return out, true, nil
    }

    func main() {
    	e := NewEnvelope()
    	fmt.Println(e.AddKey(1, bytes.Repeat([]byte{1}, 32)), e.AddKey(2, bytes.Repeat([]byte{2}, 32)))
    	fmt.Println(e.SetCurrent(2))
    	blob, err := e.Seal([]byte("ghp_123"), "vault-7/github")
    	fmt.Printf("%x %v\n", blob[:min(5, len(blob))], err)
    	pt, err := e.Open(blob, "vault-7/github")
    	fmt.Printf("%q %v\n", pt, err)
    	_, err = e.Open(blob, "vault-7/aws")
    	fmt.Println(err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"encoding/binary"
    	"errors"
    	"testing"
    )

    var (
    	key1 = bytes.Repeat([]byte{0x01}, 32) // the legacy key
    	key2 = bytes.Repeat([]byte{0x02}, 32)
    	key3 = bytes.Repeat([]byte{0x03}, 32)
    )

    const ctx = "vault-7/github"

    func gcm(key []byte) cipher.AEAD {
    	b, _ := aes.NewCipher(key)
    	a, _ := cipher.NewGCMWithRandomNonce(b)
    	return a
    }

    // legacy makes a v1 blob the way the old app did: 0x01 || nonce || ct || tag,
    // no associated data.
    func legacy(pt string) []byte {
    	return gcm(key1).Seal([]byte{0x01}, nil, []byte(pt), nil)
    }

    // envelope returns an Envelope with keys 1 and 2, current = current.
    func envelope(t *testing.T, current uint32) *Envelope {
    	t.Helper()
    	e := NewEnvelope()
    	if err := e.AddKey(1, key1); err != nil {
    		t.Fatalf("AddKey(1) = %v", err)
    	}
    	if err := e.AddKey(2, key2); err != nil {
    		t.Fatalf("AddKey(2) = %v", err)
    	}
    	if err := e.SetCurrent(current); err != nil {
    		t.Fatalf("SetCurrent(%d) = %v", current, err)
    	}
    	return e
    }

    func mustSeal(t *testing.T, e *Envelope, pt, context string) []byte {
    	t.Helper()
    	blob, err := e.Seal([]byte(pt), context)
    	if err != nil {
    		t.Fatalf("Seal returned error %v", err)
    	}
    	if len(blob) != 5+12+len(pt)+16 {
    		t.Fatalf("Seal(%d bytes) returned %d bytes, want 5 header + 12 nonce + %d + 16 tag", len(pt), len(blob), len(pt))
    	}
    	return blob
    }

    func expectErr(t *testing.T, what string, e *Envelope, blob []byte, context string, want error) {
    	t.Helper()
    	pt, err := e.Open(blob, context)
    	if !errors.Is(err, want) || pt != nil {
    		t.Errorf("%s: Open = %q, %v, want nil, %v", what, pt, err, want)
    	}
    }

    func TestSealFormat(t *testing.T) {
    	e := envelope(t, 2)
    	blob := mustSeal(t, e, "ghp_123", ctx)
    	if blob[0] != 0x02 || binary.BigEndian.Uint32(blob[1:5]) != 2 {
    		t.Fatalf("header = %x, want 02 00000002 (version 2, key ID 2)", blob[:5])
    	}
    	ad := append(bytes.Clone(blob[:5]), ctx...)
    	pt, err := gcm(key2).Open(nil, nil, blob[5:], ad)
    	if err != nil || string(pt) != "ghp_123" {
    		t.Fatalf("v2 body doesn't open under key 2 with AD = header || context: %v", err)
    	}
    }

    func TestOpenBothVersions(t *testing.T) {
    	e := envelope(t, 2)
    	if pt, err := e.Open(legacy("old-secret"), "anything"); err != nil || string(pt) != "old-secret" {
    		t.Errorf("Open(v1 blob) = %q, %v, want \"old-secret\", nil", pt, err)
    	}
    	if pt, err := e.Open(legacy(""), ctx); err != nil || len(pt) != 0 {
    		t.Errorf("Open(empty v1 blob) = %q, %v, want empty, nil", pt, err)
    	}
    	blob := mustSeal(t, e, "ghp_123", ctx)
    	if pt, err := e.Open(blob, ctx); err != nil || string(pt) != "ghp_123" {
    		t.Errorf("Open(v2 blob) = %q, %v", pt, err)
    	}
    }

    func TestRotation(t *testing.T) {
    	e := envelope(t, 1)
    	old := mustSeal(t, e, "sealed under key 1", ctx)
    	if err := e.AddKey(3, key3); err != nil {
    		t.Fatal(err)
    	}
    	if err := e.SetCurrent(3); err != nil {
    		t.Fatal(err)
    	}
    	fresh := mustSeal(t, e, "sealed under key 3", ctx)
    	if id := binary.BigEndian.Uint32(fresh[1:5]); id != 3 {
    		t.Errorf("after SetCurrent(3), Seal used key ID %d", id)
    	}
    	for _, b := range [][]byte{old, fresh} {
    		if _, err := e.Open(b, ctx); err != nil {
    			t.Errorf("after rotation, Open(blob under key %d) = %v, want nil", binary.BigEndian.Uint32(b[1:5]), err)
    		}
    	}
    	if err := e.SetCurrent(9); !errors.Is(err, ErrUnknownKey) {
    		t.Errorf("SetCurrent(9) = %v, want ErrUnknownKey", err)
    	}
    	other := NewEnvelope()
    	other.AddKey(2, key2)
    	expectErr(t, "blob under key 1 in an envelope without key 1", other, old, ctx, ErrUnknownKey)
    	expectErr(t, "v1 blob in an envelope without the legacy key", other, legacy("x"), ctx, ErrUnknownKey)
    	if _, err := NewEnvelope().Seal([]byte("x"), ctx); !errors.Is(err, ErrNoCurrentKey) {
    		t.Errorf("Seal with no current key = %v, want ErrNoCurrentKey", err)
    	}
    }

    func TestContextAndKeyIDBound(t *testing.T) {
    	e := envelope(t, 2)
    	blob := mustSeal(t, e, "ghp_123", ctx)
    	expectErr(t, "opened under a different context", e, blob, "vault-7/aws", ErrDecrypt)
    	expectErr(t, "opened with an empty context", e, blob, "", ErrDecrypt)
    	swapped := bytes.Clone(blob)
    	binary.BigEndian.PutUint32(swapped[1:5], 1)
    	expectErr(t, "key ID changed from 2 to 1", e, swapped, ctx, ErrDecrypt)
    }

    func TestNoDowngrade(t *testing.T) {
    	// v1 had no associated data. A v2 blob must never open as v1, even when
    	// sealed under the legacy key.
    	e := envelope(t, 1)
    	blob := mustSeal(t, e, "ghp_123", ctx)
    	stripped := append([]byte{0x01}, blob[5:]...) // v1 header, v2 nonce||ct||tag
    	expectErr(t, "v2 blob relabelled as v1", e, stripped, ctx, ErrDecrypt)
    	relabelled := bytes.Clone(blob)
    	relabelled[0] = 0x01
    	expectErr(t, "v2 version byte changed to 1", e, relabelled, ctx, ErrDecrypt)
    }

    func TestEveryBitFlip(t *testing.T) {
    	e := envelope(t, 2)
    	blobs := []struct {
    		name string
    		blob []byte
    	}{{"v1", legacy("old")}, {"v2", mustSeal(t, e, "new", ctx)}}
    	for _, b := range blobs {
    		name, blob := b.name, b.blob
    		for i := range blob {
    			for bit := range 8 {
    				bad := bytes.Clone(blob)
    				bad[i] ^= 1 << bit
    				if pt, err := e.Open(bad, ctx); err == nil || pt != nil {
    					t.Fatalf("%s blob, bit %d of byte %d flipped: Open = %q, %v, want an error", name, bit, i, pt, err)
    				}
    			}
    		}
    	}
    }

    func TestTruncatedAndUnknown(t *testing.T) {
    	e := envelope(t, 2)
    	for _, blob := range [][]byte{legacy("old"), mustSeal(t, e, "new", ctx)} {
    		for n := range len(blob) {
    			if pt, err := e.Open(blob[:n], ctx); err == nil || pt != nil {
    				t.Fatalf("v%d blob cut to %d bytes: Open = %q, %v, want an error", blob[0], n, pt, err)
    			}
    		}
    	}
    	expectErr(t, "empty blob", e, nil, ctx, ErrUnsupportedVersion)
    	for _, v := range []byte{0x00, 0x03, 0xff} {
    		blob := mustSeal(t, e, "x", ctx)
    		blob[0] = v
    		expectErr(t, "unknown version", e, blob, ctx, ErrUnsupportedVersion)
    	}
    }

    func TestUpgrade(t *testing.T) {
    	e := envelope(t, 2)
    	current := mustSeal(t, e, "already current", ctx)
    	if out, changed, err := e.Upgrade(current, ctx); err != nil || changed || !bytes.Equal(out, current) {
    		t.Errorf("Upgrade(blob under the current key) = %d bytes, %v, %v, want the same blob, false, nil", len(out), changed, err)
    	}
    	e1 := envelope(t, 1)
    	underOld := mustSeal(t, e1, "under key 1", ctx)
    	olds := []struct {
    		name string
    		blob []byte
    	}{{"v1 blob", legacy("legacy secret")}, {"v2 blob under key 1", underOld}}
    	for _, o := range olds {
    		name, blob := o.name, o.blob
    		want, _ := e.Open(blob, ctx)
    		out, changed, err := e.Upgrade(blob, ctx)
    		if err != nil || !changed {
    			t.Errorf("Upgrade(%s) = changed %v, %v, want true, nil", name, changed, err)
    			continue
    		}
    		if len(out) < 5 || out[0] != 0x02 || binary.BigEndian.Uint32(out[1:5]) != 2 {
    			t.Errorf("Upgrade(%s) header = %x, want 02 00000002", name, out[:min(5, len(out))])
    		}
    		if pt, err := e.Open(out, ctx); err != nil || !bytes.Equal(pt, want) {
    			t.Errorf("Open(Upgrade(%s)) = %q, %v, want %q", name, pt, err, want)
    		}
    		expectErr(t, "upgraded "+name+" under another context", e, out, "vault-9/other", ErrDecrypt)
    	}
    	bad := legacy("x")
    	bad[len(bad)-1] ^= 1
    	if out, _, err := e.Upgrade(bad, ctx); err == nil || out != nil {
    		t.Errorf("Upgrade(tampered v1 blob) = %d bytes, %v, want nil and an error: never re-seal what didn't authenticate", len(out), err)
    	}
    }

    func TestDisableLegacy(t *testing.T) {
    	e := envelope(t, 2)
    	old := legacy("legacy secret")
    	upgraded, _, err := e.Upgrade(old, ctx)
    	if err != nil {
    		t.Fatalf("Upgrade(v1) = %v", err)
    	}
    	e.DisableLegacy()
    	expectErr(t, "v1 blob after DisableLegacy", e, old, ctx, ErrUnsupportedVersion)
    	if pt, err := e.Open(upgraded, ctx); err != nil || string(pt) != "legacy secret" {
    		t.Errorf("upgraded blob after DisableLegacy: Open = %q, %v", pt, err)
    	}
    }

    func TestAddKeyValidation(t *testing.T) {
    	e := NewEnvelope()
    	if err := e.AddKey(0, key1); err == nil {
    		t.Errorf("AddKey(0, ...) succeeded: ID 0 is reserved")
    	}
    	if err := e.AddKey(5, key1[:16]); err == nil {
    		t.Errorf("AddKey with a 16-byte key succeeded, want an error (AES-256 only)")
    	}
    	if err := e.AddKey(5, key1); err != nil {
    		t.Fatalf("AddKey(5) = %v", err)
    	}
    	if err := e.AddKey(5, key2); err == nil {
    		t.Errorf("AddKey(5) twice succeeded: silently replacing a key would make its old blobs unreadable")
    	}
    }

    func TestFreshNonces(t *testing.T) {
    	e := envelope(t, 2)
    	seen := map[string]bool{}
    	for i := range 10_000 {
    		blob := mustSeal(t, e, "same", ctx)
    		n := string(blob[5:17])
    		if seen[n] {
    			t.Fatalf("Seal repeated a nonce after %d messages under one key", i)
    		}
    		seen[n] = true
    	}
    }
---

Keybox 1.0 sealed every vault record with a single AES-256-GCM key and **no associated
data**, as `0x01 || nonce || ciphertext || tag`. That meant the server could swap two
records' ciphertexts and nobody would notice. Keybox 2.0 fixes it with a new format
that names its key (so keys can rotate) and binds each record to its context, while
still reading every 1.0 record until the migration finishes.

```
v1 (legacy, read-only):  0x01 || nonce(12) || ct || tag(16)
                         key = key ID 1, associated data = none
v2:                      0x02 || key ID (uint32 big-endian) || nonce(12) || ct || tag(16)
                         associated data = the 5 header bytes || context
```

Implement `Envelope`:

- `AddKey(id, key)` registers a 32-byte AES key. It returns an error for ID 0, a key
  that isn't 32 bytes, or an ID that's already registered (never silently replace one).
- `SetCurrent(id)` picks the key new records are sealed with (`ErrUnknownKey` if missing).
- `Seal(plaintext, context)` always writes **v2** under the current key with a fresh
  random nonce (`ErrNoCurrentKey` if none is set).
- `Open(blob, context)` reads v1 and v2:
  - `ErrUnsupportedVersion` for an empty blob, an unknown version byte, or a v1 blob
    after `DisableLegacy`;
  - `ErrUnknownKey` if the blob's key isn't registered (for v1, key 1);
  - `ErrDecrypt` for anything else: too short, tampered, or wrong context.
- `Upgrade(blob, context)` returns the record re-sealed as v2 under the current key and
  `true`. If the blob is *already* v2 under the current key it returns it unchanged and
  `false`. A blob that doesn't open returns that error and no output.
- `DisableLegacy()` turns off v1 for good, once every record has been upgraded.

## Example

```go
e := NewEnvelope()
e.AddKey(1, legacyKey); e.AddKey(2, newKey); e.SetCurrent(2)
e.Open(oldV1Record, "vault-7/github")        // plaintext, nil (context can't be checked for v1)
blob, _ := e.Seal(pt, "vault-7/github")      // 02 00000002 ...
e.Open(blob, "vault-7/aws")                  // nil, ErrDecrypt
up, changed, _ := e.Upgrade(oldV1Record, "vault-7/github")  // v2 blob, true
e.DisableLegacy()
e.Open(oldV1Record, "vault-7/github")        // nil, ErrUnsupportedVersion
```

## Constraints

- The tests check the v2 layout with their own AES-GCM code, rotate keys, bind context
  and key ID, flip every bit of v1 and v2 blobs, truncate them, try unknown versions and
  keys, relabel v2 blobs as v1 (a **downgrade**), upgrade, disable legacy, and check
  10,000 nonces for repeats.
- Random 96-bit nonces are safe for about 2³² messages per key, which is one more good
  reason to rotate keys.
