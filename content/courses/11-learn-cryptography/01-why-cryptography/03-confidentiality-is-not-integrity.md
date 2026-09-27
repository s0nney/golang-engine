---
title: Confidentiality Is Not Integrity
quiz:
  - question: Mallory flips bit 3 of byte 10 in a message encrypted with a stream cipher (plain XOR with a keystream, no MAC). What happens when Bob decrypts it?
    options:
      - text: Decryption fails with an error
      - text: The whole message decrypts to garbage
      - text: Bit 3 of byte 10 of the plaintext flips, and everything else decrypts normally
        correct: true
      - text: Nothing, because Mallory doesn't know the key
    explanation: |
      Each plaintext bit is XORed with one keystream bit, so flipping a ciphertext bit
      flips exactly that plaintext bit. There's no check anywhere to notice. This is
      called **malleability**, and it's why modern encryption is always *authenticated*.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var ErrLengthMismatch = errors.New("known and wanted plaintexts must be the same length as the ciphertext")

    // xorPad is the one-time pad from the previous lesson.
    func xorPad(pad, data []byte) ([]byte, error) {
    	if len(pad) < len(data) {
    		return nil, errors.New("pad is shorter than the message")
    	}
    	out := make([]byte, len(data))
    	for i := range data {
    		out[i] = data[i] ^ pad[i]
    	}
    	return out, nil
    }

    // forge is Mallory's attack. Given a ciphertext, the plaintext it is
    // known to contain, and the plaintext Mallory wants instead, it returns
    // a new ciphertext that decrypts to wanted under the same (unknown!)
    // pad. It returns ErrLengthMismatch unless all three are the same length.
    func forge(ciphertext, known, wanted []byte) ([]byte, error) {
    	// ?
    	return ciphertext, nil
    }

    func main() {
    	// Only Alice's Keybox client and the sync server know this pad.
    	pad := []byte("9f2c8a1e4b7d6035c1a2e8f94d3b7a60")

    	order, _ := xorPad(pad, []byte("share db-password with bob"))

    	// Mallory sees only the ciphertext and guesses the plaintext format.
    	forged, err := forge(order, []byte("share db-password with bob"), []byte("share db-password with eve"))
    	if err != nil {
    		fmt.Println("forge:", err)
    		return
    	}

    	decrypted, _ := xorPad(pad, forged)
    	fmt.Printf("server decrypts: %q\n", decrypted)
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var ErrLengthMismatch = errors.New("known and wanted plaintexts must be the same length as the ciphertext")

    func xorPad(pad, data []byte) ([]byte, error) {
    	if len(pad) < len(data) {
    		return nil, errors.New("pad is shorter than the message")
    	}
    	out := make([]byte, len(data))
    	for i := range data {
    		out[i] = data[i] ^ pad[i]
    	}
    	return out, nil
    }

    func forge(ciphertext, known, wanted []byte) ([]byte, error) {
    	if len(known) != len(ciphertext) || len(wanted) != len(ciphertext) {
    		return nil, ErrLengthMismatch
    	}
    	out := make([]byte, len(ciphertext))
    	for i := range ciphertext {
    		out[i] = ciphertext[i] ^ known[i] ^ wanted[i]
    	}
    	return out, nil
    }

    func main() {
    	pad := []byte("9f2c8a1e4b7d6035c1a2e8f94d3b7a60")

    	order, _ := xorPad(pad, []byte("share db-password with bob"))

    	forged, err := forge(order, []byte("share db-password with bob"), []byte("share db-password with eve"))
    	if err != nil {
    		fmt.Println("forge:", err)
    		return
    	}

    	decrypted, _ := xorPad(pad, forged)
    	fmt.Printf("server decrypts: %q\n", decrypted)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"errors"
    	"testing"
    )

    func TestForgeChangesRecipient(t *testing.T) {
    	pad := []byte("an entirely different secret pad!")
    	for _, tt := range []struct{ known, wanted string }{
    		{"share db-password with bob", "share db-password with eve"},
    		{"share api-key with bob", "share api-key with eve"},
    		{"delete nothing", "delete all!!!!"},
    		{"", ""},
    	} {
    		c, _ := xorPad(pad, []byte(tt.known))
    		orig := bytes.Clone(c)
    		forged, err := forge(c, []byte(tt.known), []byte(tt.wanted))
    		if err != nil {
    			t.Fatalf("forge(ciphertext of %q, ..., %q) error = %v", tt.known, tt.wanted, err)
    		}
    		got, _ := xorPad(pad, forged)
    		if string(got) != tt.wanted {
    			t.Errorf("forged ciphertext decrypts to %q, want %q", got, tt.wanted)
    		}
    		if !bytes.Equal(c, orig) {
    			t.Errorf("forge modified the original ciphertext; return a new slice")
    		}
    	}
    }

    func TestForgeLengthMismatch(t *testing.T) {
    	c := []byte("0123456789")
    	for _, tt := range []struct{ known, wanted string }{
    		{"short", "0123456789"},
    		{"0123456789", "short"},
    		{"0123456789ab", "0123456789ab"},
    	} {
    		if _, err := forge(c, []byte(tt.known), []byte(tt.wanted)); !errors.Is(err, ErrLengthMismatch) {
    			t.Errorf("forge(10-byte ciphertext, %d-byte known, %d-byte wanted) error = %v, want ErrLengthMismatch",
    				len(tt.known), len(tt.wanted), err)
    		}
    	}
    }
---

The last lesson's one-time pad has perfect secrecy. Eve, who only *reads*, learns
nothing. Now meet Mallory, who can *change* messages in transit. Against her, the
perfect cipher is useless.

## Flipping bits you can't read

Suppose an early Keybox prototype sends sharing orders to its server as XOR-encrypted
text, using a pad only the client and server know:

```
plaintext:  share db-password with bob
ciphertext: 4a0e53115d4155071912561745475c470711165b1150465b5b06
```

Mallory can't read the ciphertext. But she can **guess** it. The format is predictable,
and she knows Alice shares with Bob every Monday. Since `c = p ^ pad`, she can compute:

```
c ^ known ^ wanted = (p ^ pad) ^ p ^ wanted = pad ^ wanted
```

That's a perfectly valid encryption of `wanted` under the pad she never saw. The server
decrypts `share db-password with eve` and does exactly what it's told.

Even without guessing the whole message, Mallory can XOR any pattern into any position
she likes. Flip the right bits in a bank transfer's amount field and `0100` becomes
`9100`. The cipher is **malleable**: predictable changes to the ciphertext cause
predictable changes to the plaintext.

## Encryption needs a seal

This isn't a quirk of the one-time pad. Every stream cipher and the popular CTR mode
behave the same way, and CBC mode, the old default, has its own flavour of the problem.
The fix is to pair encryption with an **integrity check the attacker can't forge**: a
MAC computed with a secret key (chapter 4). Modern designs bundle the two into one
primitive called an **AEAD**, authenticated encryption with associated data
(chapter 6). If anyone changes a single bit, decryption fails loudly instead of
returning altered data.

Two rules to take away:

1. **Never use unauthenticated encryption.** If a library gives you "encrypt" without a
   tag or MAC, it's the wrong tool.
2. **Decryption failing is a feature.** Code that handles the error by "trying to
   decrypt anyway" or falling back to an unauthenticated path reintroduces the attack.

You'll meet a nastier version of this later too: the **padding oracle**, where the mere
fact that a server reports "bad padding" versus "bad data" lets an attacker decrypt
CBC ciphertexts byte by byte. The root cause is the same: the server processed data
nobody had authenticated.

## Your task

Play Mallory. Complete `forge(ciphertext, known, wanted)` so that it returns a **new**
ciphertext which decrypts to `wanted` under the same unknown pad:

- If `known` or `wanted` isn't exactly as long as `ciphertext`, return
  `ErrLengthMismatch`.
- Otherwise XOR each ciphertext byte with the matching bytes of `known` and `wanted`.

Notice that `forge` never takes the pad as a parameter. That's the whole point.
