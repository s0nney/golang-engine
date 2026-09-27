---
title: Block Ciphers
quiz:
  - question: Which key lengths does `aes.NewCipher` accept?
    options:
      - text: Any length; it hashes the key first
      - text: 16, 24 or 32 bytes, selecting AES-128, AES-192 or AES-256
        correct: true
      - text: Only 32 bytes
      - text: 8 to 64 bytes
    explanation: |
      AES is defined for 128-, 192- and 256-bit keys. Anything else returns an error
      such as `crypto/aes: invalid key size 9`. It never stretches or hashes a key for
      you, which is one more reason passwords must go through a KDF.
  - question: Why can't you use a `cipher.Block` on its own to encrypt a Keybox secret?
    options:
      - text: It only transforms exactly one 16-byte block, is deterministic, and provides no integrity; you need a mode of operation, ideally an AEAD like GCM
        correct: true
      - text: It's too slow for real data
      - text: It only works on text, not bytes
    explanation: |
      A block cipher is a building block. Encrypting longer messages safely needs a
      mode that handles many blocks, randomizes with a nonce, and (for AEAD modes)
      authenticates the result.
---

Chapter 1's one-time pad needed a key as long as the message. Real symmetric
encryption gets by with a 32-byte key, thanks to a primitive called a **block cipher**.

## A keyed permutation

A block cipher takes a key and one fixed-size block of plaintext and produces a block of
ciphertext the same size. For each key it's a **permutation**: every possible 16-byte
input maps to a different 16-byte output, and with the key you can run it backwards.
Without the key, it should be indistinguishable from a randomly chosen permutation.

**AES** (the Advanced Encryption Standard, 2001) is the block cipher everyone uses. Its
block size is 16 bytes, and its key is 16, 24 or 32 bytes (AES-128, AES-192, AES-256).
Most CPUs have dedicated AES instructions, and Go uses them, so AES is both fast and
free of timing leaks on those machines.

## AES in Go

`aes.NewCipher` returns a `cipher.Block`:

```go
type Block interface {
	BlockSize() int
	Encrypt(dst, src []byte) // exactly one block
	Decrypt(dst, src []byte)
}
```

Here it is on the AES-256 example from the standard, FIPS 197:

```go
package main

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
)

func main() {
	key, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	pt, _ := hex.DecodeString("00112233445566778899aabbccddeeff")
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ct := make([]byte, aes.BlockSize)
	block.Encrypt(ct, pt)
	fmt.Printf("ciphertext: %x\n", ct)
	back := make([]byte, aes.BlockSize)
	block.Decrypt(back, ct)
	fmt.Printf("decrypted:  %x\n", back)
	_, err = aes.NewCipher([]byte("too short"))
	fmt.Println(err)
}
```

```
ciphertext: 8ea2b7ca516745bfeafc49904b496089
decrypted:  00112233445566778899aabbccddeeff
crypto/aes: invalid key size 9
```

## Why a block isn't enough

`Encrypt` handles exactly one 16-byte block, and it's **deterministic**: the same key
and block always give the same output. Three problems follow:

1. **Messages are longer than 16 bytes.** You need a rule for chaining many blocks,
   called a **mode of operation**.
2. **Determinism leaks.** If encrypting `"password=hunter2"` twice gives the same
   ciphertext, Eve learns that two stored secrets are equal without decrypting
   anything. Secure encryption must be **randomized**, usually with a nonce, so the same
   plaintext encrypts differently every time.
3. **No integrity.** A block cipher will happily "decrypt" any 16 bytes you hand it.
   As chapter 1 showed, Mallory then gets to edit data she can't read.

The rest of this chapter walks from the worst mode (ECB) to the one Keybox uses
(AES-GCM), which solves all three.

## AES-128 or AES-256?

AES-128 has no practical attacks and is still fine for most purposes. AES-256 costs a
few percent more and keeps a large margin, including against quantum computers
(Grover's algorithm would, in theory, cut AES-256 to roughly 128-bit strength). Keybox
uses **AES-256** because its master key is 32 bytes anyway and vault data may need to
stay secret for decades.

You'll almost never call `Block.Encrypt` directly in real code. It's shown here so you
know what's inside the higher-level APIs.
