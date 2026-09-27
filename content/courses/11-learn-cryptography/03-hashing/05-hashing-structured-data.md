---
title: Hashing Structured Data
quiz:
  - question: 'Keybox fingerprints a share as `sha256(owner + recipient + secretName)`. Which pair of shares collides?'
    options:
      - text: '`("alice", "bob", "db")` and `("alice", "bob", "api")`'
      - text: '`("alice", "bob", "db")` and `("alice", "bo", "bdb")`'
        correct: true
      - text: '`("alice", "bob", "db")` and `("bob", "alice", "db")`'
      - text: None, because SHA-256 is collision resistant
    explanation: |
      Both concatenate to `"alicebobdb"`, so they hash identically. SHA-256 did its job
      perfectly; the *encoding* threw away the field boundaries before the hash ever
      saw the data. Length-prefix each field to fix it.
exercise:
  starter: |
    package main

    import (
    	"crypto/sha256"
    	"fmt"
    )

    // hashFields hashes a list of fields unambiguously. For each field, in
    // order, it writes the field's length as an 8-byte big-endian integer
    // followed by the field's bytes, and returns the SHA-256 of the result.
    func hashFields(fields ...string) [32]byte {
    	// ? This version is ambiguous: ("ab", "c") and ("a", "bc") collide.
    	h := sha256.New()
    	for _, f := range fields {
    		h.Write([]byte(f))
    	}
    	var out [32]byte
    	h.Sum(out[:0])
    	return out
    }

    func main() {
    	fmt.Printf("%x\n", hashFields("alice", "bob", "db"))
    	fmt.Printf("%x\n", hashFields("alice", "bo", "bdb"))
    }
  solution: |
    package main

    import (
    	"crypto/sha256"
    	"encoding/binary"
    	"fmt"
    )

    func hashFields(fields ...string) [32]byte {
    	h := sha256.New()
    	for _, f := range fields {
    		var n [8]byte
    		binary.BigEndian.PutUint64(n[:], uint64(len(f)))
    		h.Write(n[:])
    		h.Write([]byte(f))
    	}
    	var out [32]byte
    	h.Sum(out[:0])
    	return out
    }

    func main() {
    	fmt.Printf("%x\n", hashFields("alice", "bob", "db"))
    	fmt.Printf("%x\n", hashFields("alice", "bo", "bdb"))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    func TestHashFieldsKnownAnswers(t *testing.T) {
    	for _, tt := range []struct {
    		fields []string
    		want   string
    	}{
    		{[]string{"keybox", "v1"}, "a7b1cdcb2e85dfb5887632fdb36908c7a11b1d138f04876f3855c2d252691845"},
    		{[]string{"ab", "c"}, "601d5476e2ccfe2c87a2bba7a322659734a05749d5b5aa781f513e4912db0d5f"},
    		{[]string{""}, "af5570f5a1810b7af78caf4bc70a660f0df51e42baf91d4de5b2328de0e83dfc"},
    		{nil, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
    	} {
    		if got := fmt.Sprintf("%x", hashFields(tt.fields...)); got != tt.want {
    			t.Errorf("hashFields(%q) = %s, want %s", tt.fields, got, tt.want)
    		}
    	}
    }

    func TestHashFieldsUnambiguous(t *testing.T) {
    	for _, pair := range [][2][]string{
    		{{"ab", "c"}, {"a", "bc"}},
    		{{"alice", "bob", "db"}, {"alice", "bo", "bdb"}},
    		{{"abc"}, {"abc", ""}},
    		{{""}, {}},
    		{{"", ""}, {""}},
    		{{"a\x00", "b"}, {"a", "\x00b"}},
    	} {
    		if hashFields(pair[0]...) == hashFields(pair[1]...) {
    			t.Errorf("hashFields(%q) == hashFields(%q); different field lists must hash differently", pair[0], pair[1])
    		}
    	}
    }
---

SHA-256 is collision resistant. So how did Mallory make two different Keybox shares
with the same fingerprint? She didn't attack the hash. She attacked the **encoding**.

## Concatenation is ambiguous

When you hash several fields, the obvious code glues them together:

```go
sum := sha256.Sum256([]byte(owner + recipient + secretName))
```

But `"alice" + "bob" + "db"` and `"alice" + "bo" + "bdb"` are the same string. The
boundaries between fields are lost *before* hashing, so no hash function can save you.
If that digest is later signed or MACed, Mallory can take a signature Alice made for
one share and present it for the other.

This is a real, recurring class of bug. The first version of AWS request signing
(Signature Version 1) concatenated parameter names and values without separators, so
`?a=1&b=2` and `?a=1b2` both became `a1b2` and shared one signature. Amazon had to
replace the whole scheme.

Separators like `|` help only if fields can never contain the separator. They usually
can, eventually.

## Length-prefixing

The robust fix is to write each field's **length** before its bytes. A fixed-size
length field means the decoder always knows where one field ends and the next begins:

```
("ab", "c")  ->  00000000 00000002 'a' 'b'  00000000 00000001 'c'
("a", "bc")  ->  00000000 00000001 'a'      00000000 00000002 'b' 'c'
```

The two encodings differ, so their hashes differ. `encoding/binary` writes the length:

```go
var n [8]byte
binary.BigEndian.PutUint64(n[:], uint64(len(field)))
h.Write(n[:])
h.Write([]byte(field))
```

Other unambiguous options you'll meet in the wild: TLS's encoding (length prefixes of
fixed sizes), ASN.1 DER in certificates, Protocol Buffers with deterministic
serialization, and JSON in a *canonical* form. JSON as your application normally
produces it isn't canonical: key order, whitespace and number formatting can all vary,
so the same data can have many encodings. Chapter 8 returns to this when Keybox signs
JSON bundles.

## Domain separation

A related trick: start every hashed structure with a fixed **label** saying what it
is and which version, such as `"keybox share fingerprint v1"`. Then a hash computed for
one purpose can never be mistaken for a hash of a different kind of object that happens
to have the same fields. You'll use the same idea for MACs, key derivation and
signatures. With `hashFields`, the label is just the first field:

```go
id := hashFields("keybox share fingerprint v1", owner, recipient, secretName)
```

## Your task

The starter's `hashFields` concatenates fields, so it's ambiguous. Fix it: for each
field, write its length as an **8-byte big-endian** unsigned integer, then its bytes,
and return the SHA-256 of the whole thing.

The tests check exact digests, so the encoding must match precisely. Remember to import
`encoding/binary`. With no fields at all, you'll get the SHA-256 of the empty string.
