---
title: Threat Models and Kerckhoffs's Principle
quiz:
  - question: A teammate proposes making Keybox safer by keeping its encryption algorithm secret. What does Kerckhoffs's principle say about that?
    options:
      - text: Good idea, because attackers can't break what they can't see
      - text: The system must stay secure even if everything except the key is public, so secrecy of the algorithm adds nothing you can rely on
        correct: true
      - text: The algorithm and the key should both be secret, and rotated together
      - text: It only applies to public-key cryptography
    explanation: |
      Algorithms leak: binaries get decompiled, employees leave, source code gets
      stolen. A key is small and easy to replace; an algorithm isn't. Design as if the
      attacker has your source code, because eventually they will.
  - question: Which of these is part of Keybox's threat model, meaning something the design must defend against?
    options:
      - text: Malware running as Alice on her unlocked laptop
      - text: An attacker who steals a copy of the Keybox server's database
        correct: true
      - text: Alice choosing to share her secret with Eve
      - text: Someone watching Alice type her password over her shoulder
    explanation: |
      A stolen database is exactly what end-to-end encryption is for: the server only
      ever holds ciphertext. The others are real risks but out of scope. Once malware
      controls Alice's device, or Alice herself hands a secret over, no encryption
      scheme can help.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // caesar shifts every lowercase letter in s forward by shift places,
    // wrapping z around to a. Everything else is left alone.
    // caesar(-shift, caesar(shift, s)) == s.
    func caesar(shift int, s string) string {
    	shift = ((shift % 26) + 26) % 26
    	return strings.Map(func(r rune) rune {
    		if r < 'a' || r > 'z' {
    			return r
    		}
    		return 'a' + (r-'a'+rune(shift))%26
    	}, s)
    }

    // crack tries every possible key and returns the first shift in 0..25
    // whose decryption of ciphertext contains crib, along with that
    // decryption. ok is false if no shift works.
    func crack(ciphertext, crib string) (shift int, plaintext string, ok bool) {
    	// ?
    	return 0, "", false
    }

    func main() {
    	intercepted := "wkh nhbera pdvwhu sdvvzrug lv fkhhvh"
    	shift, plain, ok := crack(intercepted, "keybox")
    	fmt.Println(shift, plain, ok)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func caesar(shift int, s string) string {
    	shift = ((shift % 26) + 26) % 26
    	return strings.Map(func(r rune) rune {
    		if r < 'a' || r > 'z' {
    			return r
    		}
    		return 'a' + (r-'a'+rune(shift))%26
    	}, s)
    }

    func crack(ciphertext, crib string) (shift int, plaintext string, ok bool) {
    	for shift := range 26 {
    		plain := caesar(-shift, ciphertext)
    		if strings.Contains(plain, crib) {
    			return shift, plain, true
    		}
    	}
    	return 0, "", false
    }

    func main() {
    	intercepted := "wkh nhbera pdvwhu sdvvzrug lv fkhhvh"
    	shift, plain, ok := crack(intercepted, "keybox")
    	fmt.Println(shift, plain, ok)
    }
  tests: |
    package main

    import "testing"

    func TestCrack(t *testing.T) {
    	for _, tt := range []struct {
    		plain, crib string
    		shift       int
    	}{
    		{"the keybox master password is cheese", "keybox", 3},
    		{"attack at dawn", "dawn", 13},
    		{"keybox", "keybox", 0},
    		{"meet bob at the keybox office", "bob", 25},
    		{"zebras zigzag", "zebra", 7},
    	} {
    		c := caesar(tt.shift, tt.plain)
    		shift, plain, ok := crack(c, tt.crib)
    		if !ok {
    			t.Errorf("crack(%q, %q) ok = false, want true (it was encrypted with shift %d)", c, tt.crib, tt.shift)
    			continue
    		}
    		if shift != tt.shift || plain != tt.plain {
    			t.Errorf("crack(%q, %q) = %d, %q; want %d, %q", c, tt.crib, shift, plain, tt.shift, tt.plain)
    		}
    	}
    }

    func TestCrackNoMatch(t *testing.T) {
    	shift, plain, ok := crack(caesar(5, "hello world"), "keybox")
    	if ok {
    		t.Errorf("crack found shift %d (%q) for a crib that isn't in the message, want ok = false", shift, plain)
    	}
    }
---

Before choosing any primitive, you need to answer one question: **who are you
defending against, and what can they do?** That answer is your **threat model**, and
without one, "is this secure?" has no meaning.

## Keybox's threat model

Write it down, even informally. For Keybox:

**Assets.** The secret values (passwords, API keys), and to a lesser degree the secret
*names*, which leak what services Alice uses.

**Attackers and their powers.**

- *Eve*, a passive network observer. Sees every byte between clients and server.
- *Mallory*, an active network attacker. Can drop, replay, reorder and modify traffic.
- *The server itself*, or whoever breaks into it. Has the full database and can serve
  modified data. Keybox is **end-to-end** encrypted precisely so the server is
  untrusted for confidentiality.
- *A thief with a stolen backup.* Can run offline password guessing for as long as
  they like.

**Out of scope.** Malware on Alice's device, Alice deliberately leaking a secret, and
attacks on the operating system. Saying so out loud matters: it stops you from
pretending the design covers them.

Every decision in the rest of the course traces back to this list. The thief with a
backup is why Keybox uses a slow password KDF. Mallory is why every ciphertext is
authenticated. The untrusted server is why sharing uses Bob's public key, and why
bundles are signed.

## Kerckhoffs's principle

In 1883 Auguste Kerckhoffs wrote down design rules for military ciphers. The one that
survived says: **a cryptosystem should be secure even if everything about it, except
the key, is public knowledge.** Claude Shannon's version is blunter: "the enemy knows
the system."

The alternative, *security through obscurity*, fails because secrets that are hard to
change eventually leak. Keys are 32 random bytes; you can rotate one in a second.
An algorithm is baked into every client you've shipped.

Here's the Caesar cipher, the ancient "secret algorithm" whose only key is a shift
from 0 to 25:

```go
package main

import (
	"fmt"
	"strings"
)

func caesar(shift int, s string) string {
	shift = ((shift % 26) + 26) % 26
	return strings.Map(func(r rune) rune {
		if r < 'a' || r > 'z' {
			return r
		}
		return 'a' + (r-'a'+rune(shift))%26
	}, s)
}

func main() {
	c := caesar(3, "the keybox master password is cheese")
	fmt.Println(c)
	fmt.Println(caesar(-3, c))
}
```

```
wkh nhbera pdvwhu sdvvzrug lv fkhhvh
the keybox master password is cheese
```

Once the algorithm is known, and it always becomes known, 26 keys is nothing. Modern
ciphers have 2^128 or 2^256 possible keys. Trying them all isn't "slow"; it's more work
than the energy output of the sun could pay for.

## Don't roll your own

Kerckhoffs's principle has a corollary for you as a developer: **use public, heavily
analysed algorithms and well-reviewed implementations.** Designing a new cipher, or a
new way of combining primitives, is a research problem. Even experts' designs get
broken, which is why they publish them and wait years before trusting them.

"Don't roll your own crypto" doesn't mean "don't learn how it works". It means: in
production code, reach for AES-GCM, not your clever XOR scheme; for HPKE, not a
home-made hybrid protocol; for TLS, not a custom handshake. This course builds some
things from primitives *so you understand them*, and says clearly when a
higher-level tool should replace them.

## Your task

Show that the Caesar cipher's key space is hopeless. Complete `crack(ciphertext, crib)`:
try every shift from 0 to 25, decrypt with `caesar(-shift, ciphertext)`, and return the
**first** shift whose plaintext contains `crib` (a word you expect to appear, which is
exactly what codebreakers call a *crib*). If none match, return `ok == false`.
