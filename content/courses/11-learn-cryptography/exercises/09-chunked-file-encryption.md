---
title: Chunked Attachments
difficulty: medium
after: symmetric-encryption
hints:
  - 'Use `cipher.NewGCM` (you choose the nonces here, so not the random-nonce variant). Chunk `i`''s nonce is the 8-byte prefix followed by `binary.BigEndian.AppendUint32(..., uint32(i))`: 12 bytes, unique per chunk within a file, and unique across files because each file gets a fresh random prefix.'
  - 'The index in the nonce stops reordering, dropping and duplicating chunks, but not cutting the file off after a whole chunk: those chunks are all genuine. The final flag in the associated data fixes that. A chunk sealed as "not final" will never open as the last chunk.'
  - 'When decrypting, a chunk is `min(chunkSize+tagSize, len(rest))` bytes, and it is the final one exactly when it uses up the rest of the input. Return `nil, ErrDecrypt` for any failure, never the chunks you had already opened.'
exercise:
  starter: |
    package main

    import (
    	"bytes"
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: attachment is corrupt or incomplete")

    const (
    	chunkSize  = 64 // plaintext bytes per chunk; small for the tests (production: 64 KiB)
    	prefixSize = 8  // random nonce prefix at the start of the file
    	tagSize    = 16 // AES-GCM tag per chunk
    )

    // encryptAttachment seals plaintext chunk by chunk with AES-256-GCM.
    func encryptAttachment(key, plaintext []byte) ([]byte, error) {
    	return nil, nil
    }

    // decryptAttachment reverses encryptAttachment.
    func decryptAttachment(key, sealed []byte) ([]byte, error) {
    	return nil, ErrDecrypt
    }

    func main() {
    	key := bytes.Repeat([]byte{9}, 32)
    	doc := bytes.Repeat([]byte("recovery-kit "), 12) // 156 bytes: 3 chunks
    	sealed, err := encryptAttachment(key, doc)
    	fmt.Println(len(sealed), err) // want: 212 <nil>  (8 + 156 + 3*16)
    	pt, err := decryptAttachment(key, sealed)
    	fmt.Println(bytes.Equal(pt, doc), err) // want: true <nil>
    	_, err = decryptAttachment(key, sealed[:prefixSize+2*(chunkSize+tagSize)])
    	fmt.Println(err) // want: keybox: attachment is corrupt or incomplete
    }
  solution: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/rand"
    	"encoding/binary"
    	"errors"
    	"fmt"
    )

    var ErrDecrypt = errors.New("keybox: attachment is corrupt or incomplete")

    const (
    	chunkSize  = 64 // plaintext bytes per chunk; small for the tests (production: 64 KiB)
    	prefixSize = 8  // random nonce prefix at the start of the file
    	tagSize    = 16 // AES-GCM tag per chunk
    )

    var (
    	adMiddle = []byte{0}
    	adFinal  = []byte{1}
    )

    func chunkNonce(prefix []byte, i uint32) []byte {
    	return binary.BigEndian.AppendUint32(bytes.Clone(prefix), i)
    }

    // encryptAttachment seals plaintext chunk by chunk with AES-256-GCM.
    func encryptAttachment(key, plaintext []byte) ([]byte, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	aead, err := cipher.NewGCM(block)
    	if err != nil {
    		return nil, err
    	}
    	out := make([]byte, prefixSize, prefixSize+len(plaintext)+(len(plaintext)/chunkSize+1)*tagSize)
    	rand.Read(out)
    	prefix := bytes.Clone(out)
    	for i := uint32(0); ; i++ {
    		n := min(chunkSize, len(plaintext))
    		chunk := plaintext[:n]
    		plaintext = plaintext[n:]
    		ad := adMiddle
    		if len(plaintext) == 0 {
    			ad = adFinal
    		}
    		out = aead.Seal(out, chunkNonce(prefix, i), chunk, ad)
    		if len(plaintext) == 0 {
    			return out, nil
    		}
    	}
    }

    // decryptAttachment reverses encryptAttachment.
    func decryptAttachment(key, sealed []byte) ([]byte, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	aead, err := cipher.NewGCM(block)
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	if len(sealed) < prefixSize+tagSize {
    		return nil, ErrDecrypt
    	}
    	prefix, rest := sealed[:prefixSize], sealed[prefixSize:]
    	var out []byte
    	for i := uint32(0); ; i++ {
    		n := min(chunkSize+tagSize, len(rest))
    		ad := adMiddle
    		if n == len(rest) {
    			ad = adFinal
    		}
    		out, err = aead.Open(out, chunkNonce(prefix, i), rest[:n], ad)
    		if err != nil {
    			return nil, ErrDecrypt
    		}
    		rest = rest[n:]
    		if len(rest) == 0 {
    			if out == nil {
    				out = []byte{}
    			}
    			return out, nil
    		}
    		if len(rest) < tagSize {
    			return nil, ErrDecrypt
    		}
    	}
    }

    func main() {
    	key := bytes.Repeat([]byte{9}, 32)
    	doc := bytes.Repeat([]byte("recovery-kit "), 12)
    	sealed, err := encryptAttachment(key, doc)
    	fmt.Println(len(sealed), err)
    	pt, err := decryptAttachment(key, sealed)
    	fmt.Println(bytes.Equal(pt, doc), err)
    	_, err = decryptAttachment(key, sealed[:prefixSize+2*(chunkSize+tagSize)])
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
    	"fmt"
    	"testing"
    )

    var testKey = bytes.Repeat([]byte{0x5a}, 32)

    func doc(n int) []byte {
    	b := make([]byte, n)
    	for i := range b {
    		b[i] = byte('a' + i%26)
    	}
    	return b
    }

    func mustEncrypt(t *testing.T, pt []byte) []byte {
    	t.Helper()
    	sealed, err := encryptAttachment(testKey, pt)
    	if err != nil {
    		t.Fatalf("encryptAttachment(%d bytes) returned error %v", len(pt), err)
    	}
    	chunks := max(1, (len(pt)+chunkSize-1)/chunkSize)
    	if want := prefixSize + len(pt) + chunks*tagSize; len(sealed) != want {
    		t.Fatalf("encryptAttachment(%d bytes) returned %d bytes, want %d (8-byte prefix + plaintext + %d chunk tags)", len(pt), len(sealed), want, chunks)
    	}
    	return sealed
    }

    // chunks splits a sealed attachment into its prefix and sealed chunks.
    func chunks(sealed []byte) (prefix []byte, cs [][]byte) {
    	prefix, rest := sealed[:prefixSize], sealed[prefixSize:]
    	for len(rest) > 0 {
    		n := min(chunkSize+tagSize, len(rest))
    		cs = append(cs, rest[:n])
    		rest = rest[n:]
    	}
    	return prefix, cs
    }

    func join(prefix []byte, cs ...[]byte) []byte {
    	return bytes.Join(append([][]byte{prefix}, cs...), nil)
    }

    func expectReject(t *testing.T, what string, sealed []byte) {
    	t.Helper()
    	pt, err := decryptAttachment(testKey, sealed)
    	if !errors.Is(err, ErrDecrypt) || pt != nil {
    		t.Errorf("%s: decryptAttachment = %d bytes, %v, want nil, ErrDecrypt", what, len(pt), err)
    	}
    }

    func TestRoundTrip(t *testing.T) {
    	for _, n := range []int{0, 1, 63, 64, 65, 128, 129, 1000} {
    		pt := doc(n)
    		got, err := decryptAttachment(testKey, mustEncrypt(t, pt))
    		if err != nil || !bytes.Equal(got, pt) {
    			t.Errorf("round trip of %d bytes: got %d bytes, %v, want the original", n, len(got), err)
    		}
    	}
    }

    func TestExactFormat(t *testing.T) {
    	// Decrypt one chunk ourselves: nonce = prefix || uint32 index, AD = final flag.
    	sealed := mustEncrypt(t, doc(100))
    	prefix, cs := chunks(sealed)
    	block, _ := aes.NewCipher(testKey)
    	aead, _ := cipher.NewGCM(block)
    	for i, c := range cs {
    		nonce := binary.BigEndian.AppendUint32(bytes.Clone(prefix), uint32(i))
    		final := byte(0)
    		if i == len(cs)-1 {
    			final = 1
    		}
    		pt, err := aead.Open(nil, nonce, c, []byte{final})
    		if want := doc(100)[i*chunkSize : min(100, (i+1)*chunkSize)]; err != nil || !bytes.Equal(pt, want) {
    			t.Errorf("chunk %d doesn't open with nonce prefix||%d and AD {%d}: %v", i, i, final, err)
    		}
    	}
    }

    func TestReorderDropDuplicate(t *testing.T) {
    	prefix, cs := chunks(mustEncrypt(t, doc(3*chunkSize+10))) // 4 chunks
    	expectReject(t, "chunks 0 and 1 swapped", join(prefix, cs[1], cs[0], cs[2], cs[3]))
    	expectReject(t, "middle chunk dropped", join(prefix, cs[0], cs[2], cs[3]))
    	expectReject(t, "first chunk duplicated", join(prefix, cs[0], cs[0], cs[1], cs[2], cs[3]))
    	expectReject(t, "final chunk dropped", join(prefix, cs[0], cs[1], cs[2]))
    	expectReject(t, "only the first chunk", join(prefix, cs[0]))
    	expectReject(t, "final chunk repeated", join(prefix, cs[0], cs[1], cs[2], cs[3], cs[3]))
    }

    func TestSpliceAcrossFiles(t *testing.T) {
    	pa, a := chunks(mustEncrypt(t, doc(200)))
    	_, b := chunks(mustEncrypt(t, doc(200)))
    	expectReject(t, "chunk 1 taken from another file under the same key", join(pa, a[0], b[1], a[2], a[3]))
    }

    func TestEveryBitFlip(t *testing.T) {
    	sealed := mustEncrypt(t, doc(150))
    	for i := range sealed {
    		for bit := range 8 {
    			bad := bytes.Clone(sealed)
    			bad[i] ^= 1 << bit
    			pt, err := decryptAttachment(testKey, bad)
    			if !errors.Is(err, ErrDecrypt) || pt != nil {
    				t.Fatalf("flipping bit %d of byte %d: decryptAttachment = %d bytes, %v, want nil, ErrDecrypt", bit, i, len(pt), err)
    			}
    		}
    	}
    }

    func TestTruncationAndWrongKey(t *testing.T) {
    	sealed := mustEncrypt(t, doc(150))
    	for n := range len(sealed) {
    		pt, err := decryptAttachment(testKey, sealed[:n])
    		if !errors.Is(err, ErrDecrypt) || pt != nil {
    			t.Fatalf("cut to %d of %d bytes: decryptAttachment = %d bytes, %v, want nil, ErrDecrypt", n, len(sealed), len(pt), err)
    		}
    	}
    	other := bytes.Repeat([]byte{0x5b}, 32)
    	if pt, err := decryptAttachment(other, sealed); !errors.Is(err, ErrDecrypt) || pt != nil {
    		t.Errorf("wrong key: decryptAttachment = %d bytes, %v, want nil, ErrDecrypt", len(pt), err)
    	}
    	if _, err := encryptAttachment(make([]byte, 7), doc(10)); err == nil {
    		t.Errorf("encryptAttachment with a 7-byte key returned no error")
    	}
    }

    func TestFreshPrefixes(t *testing.T) {
    	seen := map[string]bool{}
    	for i := range 10_000 {
    		sealed, err := encryptAttachment(testKey, []byte("same attachment"))
    		if err != nil || len(sealed) < prefixSize {
    			t.Fatalf("encryptAttachment = %d bytes, %v", len(sealed), err)
    		}
    		p := fmt.Sprintf("%x", sealed[:prefixSize])
    		if seen[p] {
    			t.Fatalf("encryption %d reused nonce prefix %s: with the same key that repeats every chunk nonce", i, p)
    		}
    		seen[p] = true
    	}
    }
---

Keybox lets you attach files (a scanned recovery kit, an SSH key) to a vault entry.
Attachments can be large, so they're encrypted in **chunks** that can be processed
one at a time. Sealing each chunk separately is easy; the hard part is making sure
nobody can **reorder, drop, duplicate or splice** chunks, or **cut the file short**
at a chunk boundary, without being caught.

Implement `encryptAttachment(key, plaintext)` and `decryptAttachment(key, sealed)`
with AES-256-GCM (`key` is 32 bytes) and this format:

```
prefix (8 random bytes) || chunk 0 || chunk 1 || ... || chunk n-1
chunk i = AES-GCM seal of plaintext[i*64 : (i+1)*64]
          nonce = prefix || uint32(i) big-endian        (12 bytes)
          associated data = {1} for the last chunk, {0} for all others
```

Every chunk holds `chunkSize` (64) plaintext bytes except the last, which holds the
remaining 1 to 64 bytes. Empty plaintext is a single, empty, final chunk. So the
output is always `8 + len(plaintext) + 16 × (number of chunks)` bytes.

`decryptAttachment` returns the plaintext, or `nil, ErrDecrypt` for **any** problem.
`encryptAttachment` returns an error for a key of the wrong size.

## Example

```go
sealed, _ := encryptAttachment(key, doc)       // doc is 156 bytes: chunks of 64, 64, 28
len(sealed)                                    // 212 = 8 + 156 + 3*16
decryptAttachment(key, sealed)                 // doc, nil
decryptAttachment(key, sealed[:8+2*80])        // nil, ErrDecrypt (final chunk missing)
```

## Constraints

- `chunkSize` is 64 bytes so the tests can build multi-chunk files cheaply. Real
  formats use chunks of around 64 KiB.
- The tests decrypt your chunks with their own code to check the nonce and associated
  data layout, and then try swapped, dropped, duplicated and spliced chunks, every
  single-bit flip, every truncation, a wrong key, and 10,000 encryptions of the same
  file (no nonce prefix may repeat).
