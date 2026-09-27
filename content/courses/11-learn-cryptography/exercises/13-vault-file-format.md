---
title: Vault File with Password Change
difficulty: hard
after: symmetric-encryption
hints:
  - 'Two keys: a random 32-byte **data key** encrypts the contents, and a **key-encryption key** from PBKDF2(password, salt, iterations) encrypts only the data key. Changing the password then means unwrapping the data key with the old KEK and wrapping it under a new one, with a new salt, while the body bytes stay exactly as they were.'
  - 'Build the header first (`binary.BigEndian.AppendUint32([]byte(magic), uint32(iterations))` plus a fresh salt) and pass it as the associated data when sealing the data key; `cipher.NewGCMWithRandomNonce`''s `Seal(header, nil, dataKey, header)` appends nonce, ciphertext and tag straight after it.'
  - 'Order matters in `OpenVault`: check the length, the magic and the iteration range **before** running PBKDF2 (an attacker-chosen count of four billion would hang the app), then unwrap (`ErrWrongPassword`), then open the body (`ErrCorrupt`). `ChangePassword` should fully open the vault before rewrapping so it never blesses a damaged file.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var (
    	ErrBadHeader     = errors.New("keybox: not a valid vault file")
    	ErrWrongPassword = errors.New("keybox: wrong password or damaged key block")
    	ErrCorrupt       = errors.New("keybox: vault contents are damaged")
    	ErrWeakParams    = errors.New("keybox: iteration count out of range")
    )

    const (
    	magic         = "KBX1"
    	saltSize      = 16
    	headerSize    = 4 + 4 + saltSize // magic || iterations || salt
    	wrappedSize   = 12 + 32 + 16     // nonce || encrypted data key || tag
    	minIterations = 1_000            // low so the tests run fast; production: 600_000
    	maxIterations = 10_000_000
    )

    // CreateVault encrypts contents into a new vault file protected by password.
    func CreateVault(password string, iterations int, contents []byte) ([]byte, error) {
    	return nil, nil
    }

    // OpenVault decrypts a vault file.
    func OpenVault(password string, file []byte) ([]byte, error) {
    	return nil, ErrBadHeader
    }

    // ChangePassword re-protects file under newPassword without re-encrypting
    // the contents.
    func ChangePassword(file []byte, oldPassword, newPassword string, newIterations int) ([]byte, error) {
    	return nil, ErrBadHeader
    }

    func main() {
    	file, err := CreateVault("hunter2", 2_000, []byte(`{"github":"ghp_123"}`))
    	fmt.Println(len(file), err) // want: 132 <nil>
    	file, err = ChangePassword(file, "hunter2", "correct horse battery staple", 3_000)
    	fmt.Println(len(file), err) // want: 132 <nil>
    	pt, err := OpenVault("correct horse battery staple", file)
    	fmt.Printf("%s %v\n", pt, err) // want: {"github":"ghp_123"} <nil>
    	_, err = OpenVault("hunter2", file)
    	fmt.Println(err) // want: keybox: wrong password or damaged key block
    }
  solution: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/pbkdf2"
    	"crypto/rand"
    	"crypto/sha256"
    	"encoding/binary"
    	"errors"
    	"fmt"
    )

    var (
    	ErrBadHeader     = errors.New("keybox: not a valid vault file")
    	ErrWrongPassword = errors.New("keybox: wrong password or damaged key block")
    	ErrCorrupt       = errors.New("keybox: vault contents are damaged")
    	ErrWeakParams    = errors.New("keybox: iteration count out of range")
    )

    const (
    	magic         = "KBX1"
    	saltSize      = 16
    	headerSize    = 4 + 4 + saltSize // magic || iterations || salt
    	wrappedSize   = 12 + 32 + 16     // nonce || encrypted data key || tag
    	minIterations = 1_000            // low so the tests run fast; production: 600_000
    	maxIterations = 10_000_000
    )

    func newGCM(key []byte) cipher.AEAD {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		panic(err) // keys here are always 32 bytes
    	}
    	aead, err := cipher.NewGCMWithRandomNonce(block)
    	if err != nil {
    		panic(err)
    	}
    	return aead
    }

    // newHeader returns magic || iterations || fresh salt.
    func newHeader(iterations int) []byte {
    	h := binary.BigEndian.AppendUint32([]byte(magic), uint32(iterations))
    	salt := make([]byte, saltSize)
    	rand.Read(salt)
    	return append(h, salt...)
    }

    // kek derives the key-encryption key from password and a parsed header.
    func kek(password string, header []byte) []byte {
    	iterations := int(binary.BigEndian.Uint32(header[4:8]))
    	k, err := pbkdf2.Key(sha256.New, password, header[8:headerSize], iterations, 32)
    	if err != nil {
    		panic(err)
    	}
    	return k
    }

    // wrap builds header || wrapped data key for password.
    func wrap(password string, iterations int, dataKey []byte) []byte {
    	header := newHeader(iterations)
    	return newGCM(kek(password, header)).Seal(header, nil, dataKey, header)
    }

    // unwrap validates the header and recovers the data key.
    func unwrap(password string, file []byte) ([]byte, error) {
    	if len(file) < headerSize+wrappedSize+28 || string(file[:4]) != magic {
    		return nil, ErrBadHeader
    	}
    	iterations := binary.BigEndian.Uint32(file[4:8])
    	if iterations < minIterations || iterations > maxIterations {
    		return nil, ErrBadHeader
    	}
    	header := file[:headerSize]
    	dataKey, err := newGCM(kek(password, header)).Open(nil, nil, file[headerSize:headerSize+wrappedSize], header)
    	if err != nil {
    		return nil, ErrWrongPassword
    	}
    	return dataKey, nil
    }

    func checkIterations(n int) error {
    	if n < minIterations || n > maxIterations {
    		return ErrWeakParams
    	}
    	return nil
    }

    // CreateVault encrypts contents into a new vault file protected by password.
    func CreateVault(password string, iterations int, contents []byte) ([]byte, error) {
    	if err := checkIterations(iterations); err != nil {
    		return nil, err
    	}
    	dataKey := make([]byte, 32)
    	rand.Read(dataKey)
    	file := wrap(password, iterations, dataKey)
    	return newGCM(dataKey).Seal(file, nil, contents, []byte(magic)), nil
    }

    // OpenVault decrypts a vault file.
    func OpenVault(password string, file []byte) ([]byte, error) {
    	dataKey, err := unwrap(password, file)
    	if err != nil {
    		return nil, err
    	}
    	pt, err := newGCM(dataKey).Open(nil, nil, file[headerSize+wrappedSize:], []byte(magic))
    	if err != nil {
    		return nil, ErrCorrupt
    	}
    	return pt, nil
    }

    // ChangePassword re-protects file under newPassword without re-encrypting
    // the contents.
    func ChangePassword(file []byte, oldPassword, newPassword string, newIterations int) ([]byte, error) {
    	if err := checkIterations(newIterations); err != nil {
    		return nil, err
    	}
    	// Open fully first: never rewrap a vault whose contents are damaged.
    	if _, err := OpenVault(oldPassword, file); err != nil {
    		return nil, err
    	}
    	dataKey, _ := unwrap(oldPassword, file)
    	out := wrap(newPassword, newIterations, dataKey)
    	return append(out, bytes.Clone(file[headerSize+wrappedSize:])...), nil
    }

    func main() {
    	file, err := CreateVault("hunter2", 2_000, []byte(`{"github":"ghp_123"}`))
    	fmt.Println(len(file), err)
    	file, err = ChangePassword(file, "hunter2", "correct horse battery staple", 3_000)
    	fmt.Println(len(file), err)
    	pt, err := OpenVault("correct horse battery staple", file)
    	fmt.Printf("%s %v\n", pt, err)
    	_, err = OpenVault("hunter2", file)
    	fmt.Println(err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/pbkdf2"
    	"crypto/sha256"
    	"encoding/binary"
    	"errors"
    	"testing"
    	"time"
    )

    const (
    	pw       = "hunter2"
    	newPW    = "correct horse battery staple"
    	iter     = 1_000
    	contents = `{"github":"ghp_123","db":"s3cr3t"}`
    	bodyAt   = headerSize + wrappedSize
    )

    func mustCreate(t *testing.T, password string, iterations int, data string) []byte {
    	t.Helper()
    	file, err := CreateVault(password, iterations, []byte(data))
    	if err != nil {
    		t.Fatalf("CreateVault returned error %v", err)
    	}
    	if want := bodyAt + 12 + len(data) + 16; len(file) != want {
    		t.Fatalf("CreateVault(%d-byte contents) returned %d bytes, want %d (24 header + 60 key block + 12 + %d + 16 body)", len(data), len(file), want, len(data))
    	}
    	return file
    }

    func gcm(key []byte) cipher.AEAD {
    	b, _ := aes.NewCipher(key)
    	a, _ := cipher.NewGCMWithRandomNonce(b)
    	return a
    }

    func TestRoundTrip(t *testing.T) {
    	for _, data := range []string{"", "x", contents} {
    		file := mustCreate(t, pw, iter, data)
    		got, err := OpenVault(pw, file)
    		if err != nil || string(got) != data {
    			t.Errorf("OpenVault(CreateVault(%q)) = %q, %v", data, got, err)
    		}
    	}
    }

    func TestExactFormat(t *testing.T) {
    	file := mustCreate(t, pw, 1_234, contents)
    	if string(file[:4]) != "KBX1" {
    		t.Fatalf("file starts with %q, want magic \"KBX1\"", file[:4])
    	}
    	if n := binary.BigEndian.Uint32(file[4:8]); n != 1_234 {
    		t.Fatalf("iterations field = %d, want 1234 (uint32 big-endian at bytes 4-7)", n)
    	}
    	header := file[:headerSize]
    	kek, _ := pbkdf2.Key(sha256.New, pw, header[8:], 1_234, 32)
    	dataKey, err := gcm(kek).Open(nil, nil, file[headerSize:bodyAt], header)
    	if err != nil || len(dataKey) != 32 {
    		t.Fatalf("key block doesn't open as AES-GCM under PBKDF2(password, salt, iterations) with the 24-byte header as associated data: %v", err)
    	}
    	pt, err := gcm(dataKey).Open(nil, nil, file[bodyAt:], []byte("KBX1"))
    	if err != nil || string(pt) != contents {
    		t.Fatalf("body doesn't open under the data key with associated data \"KBX1\": %v", err)
    	}
    }

    func TestFreshSaltAndKeys(t *testing.T) {
    	a, b := mustCreate(t, pw, iter, contents), mustCreate(t, pw, iter, contents)
    	if bytes.Equal(a[8:headerSize], b[8:headerSize]) {
    		t.Errorf("two vaults with the same password got the same salt")
    	}
    	if bytes.Equal(a[bodyAt:], b[bodyAt:]) {
    		t.Errorf("two vaults with the same contents have identical bodies: use a fresh data key and nonce")
    	}
    }

    func TestWrongPassword(t *testing.T) {
    	file := mustCreate(t, pw, iter, contents)
    	for _, p := range []string{"", "Hunter2", "hunter2 ", "hunter"} {
    		if got, err := OpenVault(p, file); !errors.Is(err, ErrWrongPassword) || got != nil {
    			t.Errorf("OpenVault(%q) = %q, %v, want nil, ErrWrongPassword", p, got, err)
    		}
    	}
    }

    func TestEveryBitFlip(t *testing.T) {
    	file := mustCreate(t, pw, iter, "tiny")
    	for i := range file {
    		if i >= 4 && i < 8 {
    			continue // iterations field: see TestTamperedIterations (a flip can mean millions of iterations)
    		}
    		for bit := range 8 {
    			bad := bytes.Clone(file)
    			bad[i] ^= 1 << bit
    			got, err := OpenVault(pw, bad)
    			if err == nil || got != nil {
    				t.Fatalf("flipping bit %d of byte %d: OpenVault = %q, %v, want an error", bit, i, got, err)
    			}
    			if i >= bodyAt && !errors.Is(err, ErrCorrupt) {
    				t.Fatalf("flipping bit %d of body byte %d: err = %v, want ErrCorrupt", bit, i, err)
    			}
    			if i < 4 && !errors.Is(err, ErrBadHeader) {
    				t.Fatalf("flipping bit %d of magic byte %d: err = %v, want ErrBadHeader", bit, i, err)
    			}
    		}
    	}
    }

    func TestTamperedIterations(t *testing.T) {
    	file := mustCreate(t, pw, iter, contents)
    	for _, n := range []uint32{iter + 1, iter * 2, iter ^ 0x800} {
    		bad := bytes.Clone(file)
    		binary.BigEndian.PutUint32(bad[4:8], n)
    		if got, err := OpenVault(pw, bad); !errors.Is(err, ErrWrongPassword) || got != nil {
    			t.Errorf("iterations changed from %d to %d: OpenVault = %q, %v, want nil, ErrWrongPassword", iter, n, got, err)
    		}
    	}
    }

    func TestBadHeaders(t *testing.T) {
    	file := mustCreate(t, pw, iter, contents)
    	for n := range bodyAt + 28 {
    		if got, err := OpenVault(pw, file[:n]); !errors.Is(err, ErrBadHeader) || got != nil {
    			t.Fatalf("file cut to %d bytes: OpenVault = %q, %v, want nil, ErrBadHeader", n, got, err)
    		}
    	}
    	for _, n := range []uint32{0, 1, minIterations - 1, maxIterations + 1, 0xffffffff} {
    		bad := bytes.Clone(file)
    		binary.BigEndian.PutUint32(bad[4:8], n)
    		type result struct {
    			pt  []byte
    			err error
    		}
    		done := make(chan result, 1)
    		go func() {
    			pt, err := OpenVault(pw, bad)
    			done <- result{pt, err}
    		}()
    		var got []byte
    		var err error
    		select {
    		case r := <-done:
    			got, err = r.pt, r.err
    		case <-time.After(time.Second):
    			t.Fatalf("iterations field %d: OpenVault was still deriving a key after 1s: reject out-of-range counts before running PBKDF2", n)
    		}
    		if !errors.Is(err, ErrBadHeader) || got != nil {
    			t.Errorf("iterations field %d: OpenVault = %q, %v, want nil, ErrBadHeader (don't derive with it)", n, got, err)
    		}
    	}
    	if got, err := OpenVault(pw, file[:len(file)-1]); !errors.Is(err, ErrCorrupt) || got != nil {
    		t.Errorf("last byte cut off: OpenVault = %q, %v, want nil, ErrCorrupt", got, err)
    	}
    }

    func TestIterationLimits(t *testing.T) {
    	for _, n := range []int{0, minIterations - 1, maxIterations + 1} {
    		if f, err := CreateVault(pw, n, nil); !errors.Is(err, ErrWeakParams) {
    			t.Errorf("CreateVault(%d iterations) = %d bytes, %v, want ErrWeakParams", n, len(f), err)
    		}
    	}
    	file := mustCreate(t, pw, iter, contents)
    	if f, err := ChangePassword(file, pw, newPW, 10); !errors.Is(err, ErrWeakParams) {
    		t.Errorf("ChangePassword(10 iterations) = %d bytes, %v, want ErrWeakParams", len(f), err)
    	}
    }

    func TestChangePassword(t *testing.T) {
    	file := mustCreate(t, pw, iter, contents)
    	orig := bytes.Clone(file)
    	changed, err := ChangePassword(file, pw, newPW, 1_500)
    	if err != nil {
    		t.Fatalf("ChangePassword returned error %v", err)
    	}
    	if !bytes.Equal(file, orig) {
    		t.Errorf("ChangePassword modified its input slice")
    	}
    	if len(changed) != len(file) {
    		t.Fatalf("ChangePassword returned %d bytes, want %d", len(changed), len(file))
    	}
    	if !bytes.Equal(changed[bodyAt:], file[bodyAt:]) {
    		t.Errorf("ChangePassword re-encrypted the body: only the header and key block should change")
    	}
    	if bytes.Equal(changed[8:headerSize], file[8:headerSize]) {
    		t.Errorf("ChangePassword kept the old salt")
    	}
    	if n := binary.BigEndian.Uint32(changed[4:8]); n != 1_500 {
    		t.Errorf("after ChangePassword the iterations field is %d, want 1500", n)
    	}
    	if got, err := OpenVault(newPW, changed); err != nil || string(got) != contents {
    		t.Errorf("OpenVault(new password) = %q, %v, want the contents", got, err)
    	}
    	if got, err := OpenVault(pw, changed); !errors.Is(err, ErrWrongPassword) || got != nil {
    		t.Errorf("OpenVault(old password) on the new file = %q, %v, want nil, ErrWrongPassword", got, err)
    	}
    	if _, err := ChangePassword(file, "wrong", newPW, 1_500); !errors.Is(err, ErrWrongPassword) {
    		t.Errorf("ChangePassword with the wrong old password: err = %v, want ErrWrongPassword", err)
    	}
    	damaged := bytes.Clone(file)
    	damaged[len(damaged)-1] ^= 1
    	if _, err := ChangePassword(damaged, pw, newPW, 1_500); !errors.Is(err, ErrCorrupt) {
    		t.Errorf("ChangePassword on a vault with a damaged body: err = %v, want ErrCorrupt (check before rewrapping)", err)
    	}
    }
---

Keybox stores the vault on disk as a single file protected by the user's password.
Users change passwords, so the design must let them do it **without re-encrypting the
whole vault**, and the file has to carry its own KDF parameters so old files keep
opening when the defaults go up.

Implement `CreateVault`, `OpenVault` and `ChangePassword` for this format:

```
offset  size  field
0       4     magic "KBX1"
4       4     PBKDF2 iterations, uint32 big-endian
8       16    salt (random)
24      60    key block: AES-256-GCM seal of the 32-byte data key
                key   = PBKDF2-HMAC-SHA256(password, salt, iterations, 32)
                AD    = the 24 header bytes above
                bytes = nonce (12) || encrypted key (32) || tag (16)
84      ...   body: AES-256-GCM seal of the contents under the data key
                AD    = "KBX1"
                bytes = nonce (12) || ciphertext || tag (16)
```

- `CreateVault(password, iterations, contents)` uses a fresh salt, a fresh random data
  key and fresh nonces. It returns `ErrWeakParams` if `iterations` is outside
  `[minIterations, maxIterations]`.
- `OpenVault(password, file)` returns the contents, or
  - `ErrBadHeader` if the file is too short to hold header, key block and an empty
    body (112 bytes), has the wrong magic, or an iteration count outside the allowed
    range. Check these **before** deriving anything;
  - `ErrWrongPassword` if the key block doesn't open;
  - `ErrCorrupt` if the body doesn't open.
- `ChangePassword(file, oldPassword, newPassword, newIterations)` returns a new file
  with a new salt, the new iteration count and the **same data key**, rewrapped. The
  body bytes must be identical to the old file's. It fails with the errors above if
  the old file doesn't open completely, and with `ErrWeakParams` for a bad count. It
  must not modify `file`.

## Example

```go
file, _ := CreateVault("hunter2", 2000, []byte(`{"github":"ghp_123"}`))  // 132 bytes
file, _ = ChangePassword(file, "hunter2", "correct horse battery staple", 3000)
OpenVault("correct horse battery staple", file)  // {"github":"ghp_123"}, nil
OpenVault("hunter2", file)                       // nil, ErrWrongPassword
```

## Constraints

- `minIterations` is 1,000 so the tests run quickly; production code should use at
  least **600,000** PBKDF2-HMAC-SHA256 iterations (or a memory-hard KDF).
- The tests open your files with their own code to check the layout, flip every bit
  outside the iteration field, try tampered and absurd iteration counts, cut the file
  to every short length, and check that a password change leaves the body untouched.
- A password change doesn't protect old copies: anyone holding a backup of the old
  file and the old password can still open it. Rotating the data key (re-encrypting
  everything) is the fix when a password actually leaked.
