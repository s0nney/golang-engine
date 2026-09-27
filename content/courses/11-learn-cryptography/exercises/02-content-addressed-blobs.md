---
title: Content-Addressed Blobs
difficulty: easy
after: hashing
hints:
  - '`sha256.Sum256(data)` returns a `[32]byte` array; slice it with `sum[:]` before passing it to `hex.EncodeToString`. Write a small `blobID(data)` helper, since `Put` and `Get` both need it.'
  - 'Content addressing means the ID *is* the checksum. In `Get`, recompute `blobID` of the stored bytes and compare it with `id`: that''s what catches a flipped bit or a swapped blob.'
  - '`Put` must store `bytes.Clone(data)`. Storing `data` itself keeps a reference to the caller''s buffer, so their later writes would silently change your "immutable" blob.'
exercise:
  starter: |
    package main

    import "fmt"

    // BlobStore keeps encrypted vault blobs under the hash of their bytes.
    type BlobStore struct {
    	// blobs maps an ID ("sha256:" + 64 lowercase hex chars) to the blob's bytes.
    	// Imagine it lives on a sync server you don't fully trust.
    	blobs map[string][]byte
    }

    func NewBlobStore() *BlobStore {
    	return &BlobStore{blobs: map[string][]byte{}}
    }

    // Put stores a private copy of data and returns its ID:
    // "sha256:" followed by the lowercase hex SHA-256 of data.
    // Putting the same bytes twice stores them only once.
    func (s *BlobStore) Put(data []byte) string {
    	// 1. Hash data with sha256.Sum256 and build the ID with hex.EncodeToString.
    	// 2. Store a copy (bytes.Clone), so later changes to data can't reach the store.
    	return ""
    }

    // Get returns the blob stored under id. Before returning it, Get re-hashes
    // the stored bytes: if they no longer match id (the storage was corrupted
    // or tampered with), or id is unknown, it returns nil, false.
    func (s *BlobStore) Get(id string) ([]byte, bool) {
    	// Look the blob up, then check that its ID is still id.
    	return nil, false
    }

    // Len reports how many distinct blobs are stored.
    func (s *BlobStore) Len() int { return len(s.blobs) }

    func main() {
    	s := NewBlobStore()
    	id := s.Put([]byte("abc"))
    	fmt.Println(id) // want: sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
    	data, ok := s.Get(id)
    	fmt.Printf("%q %v\n", data, ok) // want: "abc" true
    }
  solution: |
    package main

    import (
    	"bytes"
    	"crypto/sha256"
    	"encoding/hex"
    	"fmt"
    )

    // BlobStore keeps encrypted vault blobs under the hash of their bytes.
    type BlobStore struct {
    	blobs map[string][]byte
    }

    func NewBlobStore() *BlobStore {
    	return &BlobStore{blobs: map[string][]byte{}}
    }

    func blobID(data []byte) string {
    	sum := sha256.Sum256(data)
    	return "sha256:" + hex.EncodeToString(sum[:])
    }

    func (s *BlobStore) Put(data []byte) string {
    	id := blobID(data)
    	if _, ok := s.blobs[id]; !ok {
    		s.blobs[id] = bytes.Clone(data)
    	}
    	return id
    }

    func (s *BlobStore) Get(id string) ([]byte, bool) {
    	data, ok := s.blobs[id]
    	if !ok || blobID(data) != id {
    		return nil, false
    	}
    	return bytes.Clone(data), true
    }

    func (s *BlobStore) Len() int { return len(s.blobs) }

    func main() {
    	s := NewBlobStore()
    	id := s.Put([]byte("abc"))
    	fmt.Println(id)
    	data, ok := s.Get(id)
    	fmt.Printf("%q %v\n", data, ok)
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    )

    func TestPutIDs(t *testing.T) {
    	// Published SHA-256 test vectors (FIPS 180-2 / NIST examples).
    	tests := []struct {
    		data, want string
    	}{
    		{"", "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
    		{"abc", "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
    		{"abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq", "sha256:248d6a61d20638b8e5c026930c3e6039a33ce45964ff2167f6ecedd419db06c1"},
    	}
    	for _, tt := range tests {
    		s := NewBlobStore()
    		if got := s.Put([]byte(tt.data)); got != tt.want {
    			t.Errorf("Put(%q) = %q, want %q", tt.data, got, tt.want)
    		}
    	}
    }

    func TestGetRoundTrip(t *testing.T) {
    	s := NewBlobStore()
    	a := s.Put([]byte("vault page 1"))
    	b := s.Put([]byte("vault page 2"))
    	if a == b {
    		t.Fatalf("different blobs got the same ID %q", a)
    	}
    	for id, want := range map[string]string{a: "vault page 1", b: "vault page 2"} {
    		got, ok := s.Get(id)
    		if !ok || string(got) != want {
    			t.Errorf("Get(%q) = %q, %v, want %q, true", id, got, ok, want)
    		}
    	}
    	if got, ok := s.Get("sha256:" + strings.Repeat("0", 64)); ok {
    		t.Errorf("Get(unknown ID) = %q, true, want nil, false", got)
    	}
    }

    func TestPutDeduplicates(t *testing.T) {
    	s := NewBlobStore()
    	id1 := s.Put([]byte("same bytes"))
    	id2 := s.Put([]byte("same bytes"))
    	s.Put([]byte("other bytes"))
    	if id1 != id2 {
    		t.Errorf("Put of identical bytes gave IDs %q and %q, want the same", id1, id2)
    	}
    	if s.Len() != 2 {
    		t.Errorf("after putting 2 distinct blobs (one twice), Len() = %d, want 2", s.Len())
    	}
    }

    func TestPutCopies(t *testing.T) {
    	s := NewBlobStore()
    	data := []byte("secret-v1")
    	id := s.Put(data)
    	data[0] = 'X' // the caller reuses its buffer
    	got, ok := s.Get(id)
    	if !ok || string(got) != "secret-v1" {
    		t.Errorf("after the caller modified its slice, Get = %q, %v, want %q, true: Put must store a copy", got, ok, "secret-v1")
    	}
    }

    func TestGetDetectsCorruption(t *testing.T) {
    	s := NewBlobStore()
    	id := s.Put([]byte("encrypted vault blob"))
    	if len(s.blobs[id]) == 0 {
    		t.Fatalf("Put returned %q but s.blobs has no blob under that ID", id)
    	}
    	s.blobs[id][3] ^= 0x01 // one flipped bit on the server's disk
    	if got, ok := s.Get(id); ok {
    		t.Errorf("Get returned %q, true for a blob whose stored bytes no longer hash to its ID, want nil, false", got)
    	}
    	id2 := s.Put([]byte("another blob"))
    	s.blobs[id2] = []byte("swapped in by the server")
    	if got, ok := s.Get(id2); ok {
    		t.Errorf("Get returned %q, true after the server swapped the blob, want nil, false", got)
    	}
    }
---

Keybox syncs encrypted vault pages through a server you don't fully trust. Each page is
stored under its **content address**: the SHA-256 of its bytes. Two devices uploading
the same page store it once, and anyone holding an ID can check that the bytes they got
back are exactly the bytes that were put in.

Complete the `BlobStore` methods:

- `Put(data)` returns the blob's ID, `"sha256:"` followed by the 64-character lowercase
  hex SHA-256 of `data`, and stores a **private copy** of `data` under it. Putting the
  same bytes again stores nothing new.
- `Get(id)` returns the stored bytes and `true`, but only after **re-hashing** them and
  checking they still match `id`. For an unknown ID, or bytes that no longer match
  (corrupted or swapped by the server), return `nil, false`.

`Len` and the `blobs` field are given; keep the field, because the tests tamper with it.

## Examples

```go
s := NewBlobStore()
s.Put([]byte("abc"))  // "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
s.Put([]byte(""))     // "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
s.Put([]byte("abc"))  // same ID again; s.Len() is still 2
```

## Constraints

- IDs use the published SHA-256 test vectors above, so the format must match exactly:
  lowercase hex, `sha256:` prefix, no spaces.
- A hash proves the bytes match the ID, not *who* wrote them. Anyone can compute a
  SHA-256, so this protects against corruption and swaps, not against someone who can
  choose which ID you ask for. That's what MACs and signatures are for.
