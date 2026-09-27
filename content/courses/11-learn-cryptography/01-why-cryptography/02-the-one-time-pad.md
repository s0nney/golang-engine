---
title: The One-Time Pad
quiz:
  - question: |
      Eve captures two messages that Keybox's prototype encrypted with the *same* pad:
      `c1 = p1 ^ pad` and `c2 = p2 ^ pad`. What does `c1 ^ c2` give her?
    options:
      - text: The pad
      - text: Random noise, because XOR with a secret pad is perfectly secure
      - text: '`p1 ^ p2`, the XOR of the two plaintexts, with the pad cancelled out'
        correct: true
      - text: Nothing, because XOR can't be applied to ciphertexts
    explanation: |
      `(p1 ^ pad) ^ (p2 ^ pad) = p1 ^ p2`, because `pad ^ pad` is zero. The XOR of two
      English texts is very easy to pull apart, especially if Eve can guess part of one
      of them. That's why it's a *one-time* pad.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var ErrShortPad = errors.New("pad is shorter than the message")

    // xorPad returns a new slice holding data XORed with pad, byte by byte.
    // It returns ErrShortPad if pad is shorter than data. Extra pad bytes
    // are ignored. data and pad are never modified.
    func xorPad(pad, data []byte) ([]byte, error) {
    	// ?
    	return nil, nil
    }

    func main() {
    	pad := []byte{0x13, 0x37, 0xc0, 0xff, 0xee, 0x42, 0x99, 0x01, 0x7a, 0x5c, 0x0d, 0xb8}
    	ciphertext, err := xorPad(pad, []byte("hunter2"))
    	fmt.Printf("ciphertext: %x %v\n", ciphertext, err)

    	plaintext, err := xorPad(pad, ciphertext)
    	fmt.Printf("plaintext:  %q %v\n", plaintext, err)

    	_, err = xorPad(pad[:3], []byte("hunter2"))
    	fmt.Println("short pad:", err)
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var ErrShortPad = errors.New("pad is shorter than the message")

    func xorPad(pad, data []byte) ([]byte, error) {
    	if len(pad) < len(data) {
    		return nil, ErrShortPad
    	}
    	out := make([]byte, len(data))
    	for i := range data {
    		out[i] = data[i] ^ pad[i]
    	}
    	return out, nil
    }

    func main() {
    	pad := []byte{0x13, 0x37, 0xc0, 0xff, 0xee, 0x42, 0x99, 0x01, 0x7a, 0x5c, 0x0d, 0xb8}
    	ciphertext, err := xorPad(pad, []byte("hunter2"))
    	fmt.Printf("ciphertext: %x %v\n", ciphertext, err)

    	plaintext, err := xorPad(pad, ciphertext)
    	fmt.Printf("plaintext:  %q %v\n", plaintext, err)

    	_, err = xorPad(pad[:3], []byte("hunter2"))
    	fmt.Println("short pad:", err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"errors"
    	"testing"
    )

    var testPad = []byte{0x13, 0x37, 0xc0, 0xff, 0xee, 0x42, 0x99, 0x01, 0x7a, 0x5c, 0x0d, 0xb8}

    func TestXorPadKnownValue(t *testing.T) {
    	got, err := xorPad(testPad, []byte("hunter2"))
    	if err != nil {
    		t.Fatalf("xorPad(pad, %q) error = %v, want nil", "hunter2", err)
    	}
    	want := []byte{0x7b, 0x42, 0xae, 0x8b, 0x8b, 0x30, 0xab}
    	if !bytes.Equal(got, want) {
    		t.Errorf("xorPad(pad, %q) = %x, want %x", "hunter2", got, want)
    	}
    }

    func TestXorPadRoundTrip(t *testing.T) {
    	for _, msg := range []string{"", "a", "hunter2", "correct hors"} {
    		c, err := xorPad(testPad, []byte(msg))
    		if err != nil {
    			t.Fatalf("xorPad(pad, %q) error = %v", msg, err)
    		}
    		if len(c) != len(msg) {
    			t.Fatalf("xorPad(pad, %q) returned %d bytes, want %d", msg, len(c), len(msg))
    		}
    		p, err := xorPad(testPad, c)
    		if err != nil {
    			t.Fatalf("decrypting %q: error = %v", msg, err)
    		}
    		if string(p) != msg {
    			t.Errorf("xorPad(pad, xorPad(pad, %q)) = %q, want the original message back", msg, p)
    		}
    	}
    }

    func TestXorPadShortPad(t *testing.T) {
    	_, err := xorPad(testPad[:3], []byte("hunter2"))
    	if !errors.Is(err, ErrShortPad) {
    		t.Errorf("xorPad with a 3-byte pad and a 7-byte message: error = %v, want ErrShortPad", err)
    	}
    }

    func TestXorPadDoesNotModifyInputs(t *testing.T) {
    	pad := bytes.Clone(testPad)
    	msg := []byte("hunter2")
    	if _, err := xorPad(pad, msg); err != nil {
    		t.Fatal(err)
    	}
    	if !bytes.Equal(pad, testPad) || string(msg) != "hunter2" {
    		t.Errorf("xorPad modified its inputs: pad=%x msg=%q; it must return a new slice", pad, msg)
    	}
    }
---

Let's start with the one cipher that's *provably* unbreakable, and see why nobody uses
it for Keybox.

## XOR

XOR (`^` in Go) compares two bits and outputs 1 when they differ. It has two
properties that make it the workhorse of symmetric encryption:

- `x ^ k ^ k == x`. XORing with the same key twice undoes it, so encryption and
  decryption are the *same* operation.
- If `k` is uniformly random, `x ^ k` is uniformly random too, whatever `x` is.

```go
package main

import "fmt"

func main() {
	secret := byte('K')     // 0100 1011
	key := byte(0b10110110) // random-looking key byte
	c := secret ^ key
	fmt.Printf("%08b ^ %08b = %08b\n", secret, key, c)
	fmt.Printf("%08b ^ %08b = %08b (%q)\n", c, key, c^key, c^key)
}
```

```
01001011 ^ 10110110 = 11111101
11111101 ^ 10110110 = 01001011 ('K')
```

## The one-time pad

The **one-time pad** XORs each byte of the message with a byte of a truly random key,
the *pad*, that's at least as long as the message. Claude Shannon proved in 1949 that
it has **perfect secrecy**: the ciphertext `7b42ae8b8b30ab` is equally likely to be
`hunter2`, `letmein`, or any other seven-byte message, because for every candidate
plaintext there's a pad that produces it. No amount of computing power helps.

So why isn't every system built on it? Look at the rules:

1. The pad must be **truly random**. (Next chapter: where randomness comes from.)
2. The pad must be **as long as all the data** you'll ever encrypt with it.
3. It must be shared **securely in advance**, and kept secret forever.
4. It must **never be reused**, not even partly.

Rules 2 and 3 are the killers. If Alice can securely hand Bob a pad as long as her
secrets, she could have handed him the secrets. Rule 4 is the one people break in
practice: reusing a pad (a "two-time pad") cancels it out, as the quiz shows. Soviet
spies reused pads in the 1940s, and the US Venona project read their messages for
decades.

Modern ciphers keep the XOR idea but replace the enormous pad with a **keystream**
generated from a short key (32 bytes) and a **nonce** (a number used once). You'll use
exactly that in the symmetric encryption chapter, and the "never reuse" rule comes
along with it.

## Your task

Complete `xorPad(pad, data)` so it returns a **new** slice holding `data[i] ^ pad[i]`
for every byte of `data`. If the pad is shorter than the data, return `ErrShortPad`.
Don't modify either argument.

Because XOR undoes itself, the same function both encrypts and decrypts.

The standard library has this built in as `subtle.XORBytes(dst, x, y)` in
`crypto/subtle`, which XORs as many bytes as the shorter input has. Write the loop
yourself this time; it's worth seeing how little is going on.
