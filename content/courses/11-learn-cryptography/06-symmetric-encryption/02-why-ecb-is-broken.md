---
title: Why ECB Is Broken
quiz:
  - question: 'Keybox''s legacy importer finds ciphertext whose 1st and 5th 16-byte blocks are identical. What can you conclude, without the key?'
    options:
      - text: Nothing; identical ciphertext blocks happen by chance all the time
      - text: The data was almost certainly encrypted in ECB mode, and plaintext blocks 1 and 5 are identical
        correct: true
      - text: The key is weak
      - text: The data is corrupted
    explanation: |
      With a proper mode, a repeated 16-byte block has a chance of around 2^-128 per
      pair, which is never. In ECB, equal plaintext blocks always give equal ciphertext
      blocks, so repeats are a fingerprint of ECB and leak which plaintext blocks match.
exercise:
  starter: |
    package main

    import (
    	"crypto/aes"
    	"fmt"
    )

    // legacyECBEncrypt is how Keybox's ancestor, "PassBin", encrypted data.
    // DO NOT USE: it exists only so the importer can recognise old files.
    // plaintext must be a multiple of 16 bytes.
    func legacyECBEncrypt(key, plaintext []byte) []byte {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		panic(err)
    	}
    	out := make([]byte, len(plaintext))
    	for i := 0; i+aes.BlockSize <= len(plaintext); i += aes.BlockSize {
    		block.Encrypt(out[i:i+aes.BlockSize], plaintext[i:i+aes.BlockSize])
    	}
    	return out
    }

    // looksLikeECB reports whether any two complete 16-byte blocks of data are
    // identical. A trailing partial block is ignored.
    func looksLikeECB(data []byte) bool {
    	// ?
    	return false
    }

    func main() {
    	key := []byte("0123456789abcdef0123456789abcdef")
    	file := []byte("password=hunter2" + "email=alice@kb.x" + "password=hunter2")
    	ct := legacyECBEncrypt(key, file)
    	fmt.Printf("%x\n%x\n%x\n", ct[0:16], ct[16:32], ct[32:48])
    	fmt.Println("looks like ECB:", looksLikeECB(ct))
    }
  solution: |
    package main

    import (
    	"crypto/aes"
    	"fmt"
    )

    func legacyECBEncrypt(key, plaintext []byte) []byte {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		panic(err)
    	}
    	out := make([]byte, len(plaintext))
    	for i := 0; i+aes.BlockSize <= len(plaintext); i += aes.BlockSize {
    		block.Encrypt(out[i:i+aes.BlockSize], plaintext[i:i+aes.BlockSize])
    	}
    	return out
    }

    func looksLikeECB(data []byte) bool {
    	seen := make(map[[aes.BlockSize]byte]bool)
    	for i := 0; i+aes.BlockSize <= len(data); i += aes.BlockSize {
    		b := [aes.BlockSize]byte(data[i : i+aes.BlockSize])
    		if seen[b] {
    			return true
    		}
    		seen[b] = true
    	}
    	return false
    }

    func main() {
    	key := []byte("0123456789abcdef0123456789abcdef")
    	file := []byte("password=hunter2" + "email=alice@kb.x" + "password=hunter2")
    	ct := legacyECBEncrypt(key, file)
    	fmt.Printf("%x\n%x\n%x\n", ct[0:16], ct[16:32], ct[32:48])
    	fmt.Println("looks like ECB:", looksLikeECB(ct))
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"testing"
    )

    var testKey = bytes.Repeat([]byte{0x5a}, 32)

    func TestDetectsECB(t *testing.T) {
    	for _, pt := range []string{
    		"password=hunter2password=hunter2",
    		"aaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbccccccccccccccccaaaaaaaaaaaaaaaa",
    		"0000000000000000111111111111111122222222222222223333333333333333" + "1111111111111111",
    	} {
    		if !looksLikeECB(legacyECBEncrypt(testKey, []byte(pt))) {
    			t.Errorf("looksLikeECB(ECB encryption of %q) = false, want true", pt)
    		}
    	}
    }

    func TestNoFalsePositives(t *testing.T) {
    	distinct := "0000000000000000111111111111111122222222222222223333333333333333"
    	if looksLikeECB(legacyECBEncrypt(testKey, []byte(distinct))) {
    		t.Errorf("looksLikeECB(ECB of 4 distinct blocks) = true, want false")
    	}
    	block, _ := aes.NewCipher(testKey)
    	gcm, _ := cipher.NewGCMWithRandomNonce(block)
    	repeated := bytes.Repeat([]byte("password=hunter2"), 8)
    	if looksLikeECB(gcm.Seal(nil, nil, repeated, nil)) {
    		t.Errorf("looksLikeECB(AES-GCM encryption of a repeated block) = true, want false")
    	}
    	for _, data := range [][]byte{nil, []byte("short"), bytes.Repeat([]byte{1}, 16), append(bytes.Repeat([]byte{1}, 16), 1)} {
    		if looksLikeECB(data) {
    			t.Errorf("looksLikeECB(%d bytes with fewer than two full blocks) = true, want false", len(data))
    		}
    	}
    }

    func TestIgnoresPartialBlock(t *testing.T) {
    	data := append(bytes.Repeat([]byte{9}, 16), bytes.Repeat([]byte{7}, 16)...)
    	data = append(data, bytes.Repeat([]byte{9}, 15)...) // partial block matching block 0's prefix
    	if looksLikeECB(data) {
    		t.Errorf("looksLikeECB compared a trailing partial block; ignore it")
    	}
    }
---

The simplest way to encrypt a long message with a block cipher is to chop it into
16-byte blocks and encrypt each one separately. That's **Electronic Codebook** (ECB)
mode, and it's the textbook example of how not to use a block cipher.

## The penguin

Since AES is deterministic, ECB encrypts equal plaintext blocks to **equal ciphertext
blocks**, wherever they appear. The famous illustration is a bitmap of the Linux
penguin, Tux, encrypted in ECB mode: every pixel changes colour, but areas of one colour
stay one (different) colour, and the penguin is perfectly recognisable.

The same happens to structured data. Suppose Keybox's ancestor, PassBin, stored records
as 16-byte fields and encrypted the file with ECB:

```
block 0: password=hunter2  ->  c2 9e 41 ...
block 1: email=alice@kb.x  ->  7f 03 d8 ...
block 2: password=hunter2  ->  c2 9e 41 ...   <- same as block 0
```

Eve learns, without the key, that two records share a password. Across many users she
learns which accounts share passwords, which values are common, and where the structure
of the file changes. In 2013, a leak of Adobe's user database showed exactly this: 150
million passwords encrypted with 3DES in ECB mode, and the password *hints* stored in
the clear. Identical ciphertexts grouped users with the same password, and the hints
often gave it away. It was nicknamed the greatest crossword puzzle in history.

ECB has more problems: an attacker can **cut and paste** ciphertext blocks to build new
valid-looking messages (swap the "role=reader" block for a "role=admin" block from
someone else's record), and, like all unauthenticated modes, it detects no tampering.

## Go doesn't offer it

The Go standard library deliberately has no ECB mode. `crypto/cipher` offers CBC, CTR
and GCM, and you'd have to write the loop yourself, as the starter code below does for
the legacy importer. If you ever see a hand-written loop calling `block.Encrypt` on
consecutive chunks, you've found ECB, and a bug.

## Detecting ECB

Since repeated blocks essentially never occur in properly encrypted data (the chance is
about 2^-128 per pair of blocks), repetition is a strong signal. It's a handy check for
a security audit or a migration tool: scan existing ciphertext, and flag files that were
clearly encrypted with ECB so they can be re-encrypted properly.

## Your task

Keybox's importer needs to recognise legacy PassBin files. Complete
`looksLikeECB(data)`: return `true` if any two **complete** 16-byte blocks of `data`
are identical, and `false` otherwise. Ignore a trailing partial block.

A map keyed by `[16]byte` makes this easy. Since Go 1.20 you can convert a slice to an
array directly: `[16]byte(data[i : i+16])` (it panics if the slice is shorter, so only
convert complete blocks).
