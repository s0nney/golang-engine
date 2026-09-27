---
title: HMAC in Go
quiz:
  - question: |
      What's wrong with this verifier?

      ```go
      func verify(key, msg, tag []byte) bool {
      	mac := hmac.New(sha256.New, key)
      	mac.Write(msg)
      	return hmac.Equal(tag, mac.Sum(tag))
      }
      ```
    options:
      - text: Nothing
      - text: '`mac.Sum(tag)` *appends* the computed MAC to `tag`, so it compares a 32-byte slice with a 64-byte one and always returns false'
        correct: true
      - text: '`hmac.Equal` should be `bytes.Equal`'
      - text: HMAC can't be used with SHA-256
    explanation: |
      `Sum(b)` appends to `b`. Always write `mac.Sum(nil)` unless you really want to
      append. As a bonus bug, if `tag` had spare capacity, `Sum` could even overwrite
      memory the caller shares.
exercise:
  starter: |
    package main

    import (
    	"encoding/hex"
    	"fmt"
    )

    // tagMessage returns HMAC-SHA256(key, msg).
    func tagMessage(key, msg []byte) []byte {
    	// ?
    	return nil
    }

    // checkTag reports whether tag is the correct HMAC-SHA256 of msg under key,
    // comparing in constant time.
    func checkTag(key, msg, tag []byte) bool {
    	// ?
    	return false
    }

    func main() {
    	// RFC 4231, test case 2.
    	tag := tagMessage([]byte("Jefe"), []byte("what do ya want for nothing?"))
    	fmt.Println(hex.EncodeToString(tag))
    	fmt.Println(checkTag([]byte("Jefe"), []byte("what do ya want for nothing?"), tag))
    	fmt.Println(checkTag([]byte("Jeff"), []byte("what do ya want for nothing?"), tag))
    }
  solution: |
    package main

    import (
    	"crypto/hmac"
    	"crypto/sha256"
    	"encoding/hex"
    	"fmt"
    )

    func tagMessage(key, msg []byte) []byte {
    	mac := hmac.New(sha256.New, key)
    	mac.Write(msg)
    	return mac.Sum(nil)
    }

    func checkTag(key, msg, tag []byte) bool {
    	return hmac.Equal(tag, tagMessage(key, msg))
    }

    func main() {
    	tag := tagMessage([]byte("Jefe"), []byte("what do ya want for nothing?"))
    	fmt.Println(hex.EncodeToString(tag))
    	fmt.Println(checkTag([]byte("Jefe"), []byte("what do ya want for nothing?"), tag))
    	fmt.Println(checkTag([]byte("Jeff"), []byte("what do ya want for nothing?"), tag))
    }
  tests: |
    package main

    import (
    	"bytes"
    	"encoding/hex"
    	"testing"
    )

    // Test vectors from RFC 4231 (HMAC-SHA-256).
    var rfc4231 = []struct {
    	name string
    	key  []byte
    	msg  string
    	want string
    }{
    	{"case 1", bytes.Repeat([]byte{0x0b}, 20), "Hi There",
    		"b0344c61d8db38535ca8afceaf0bf12b881dc200c9833da726e9376c2e32cff7"},
    	{"case 2", []byte("Jefe"), "what do ya want for nothing?",
    		"5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843"},
    	{"case 6 (131-byte key)", bytes.Repeat([]byte{0xaa}, 131), "Test Using Larger Than Block-Size Key - Hash Key First",
    		"60e431591ee0b67f0d8a26aacbf5b77f8e0bc6213728c5140546040f0ee37f54"},
    }

    func TestTagMessageRFC4231(t *testing.T) {
    	for _, v := range rfc4231 {
    		got := hex.EncodeToString(tagMessage(v.key, []byte(v.msg)))
    		if got != v.want {
    			t.Errorf("RFC 4231 %s: tagMessage = %s, want %s", v.name, got, v.want)
    		}
    	}
    }

    func TestCheckTag(t *testing.T) {
    	for _, v := range rfc4231 {
    		tag, _ := hex.DecodeString(v.want)
    		if !checkTag(v.key, []byte(v.msg), tag) {
    			t.Errorf("RFC 4231 %s: checkTag(correct tag) = false, want true", v.name)
    		}
    		if checkTag(v.key, []byte(v.msg+"!"), tag) {
    			t.Errorf("RFC 4231 %s: checkTag(modified message) = true, want false", v.name)
    		}
    		if checkTag(append(bytes.Clone(v.key), 1), []byte(v.msg), tag) {
    			t.Errorf("RFC 4231 %s: checkTag(wrong key) = true, want false", v.name)
    		}
    		if checkTag(v.key, []byte(v.msg), tag[:16]) {
    			t.Errorf("RFC 4231 %s: checkTag(truncated tag) = true, want false", v.name)
    		}
    		if checkTag(v.key, []byte(v.msg), nil) {
    			t.Errorf("RFC 4231 %s: checkTag(empty tag) = true, want false", v.name)
    		}
    	}
    }
---

`crypto/hmac` turns any hash constructor into a MAC. It's the same package you used for
[Signed Access Tokens](/courses/learn-http-servers/authentication/signed-tokens) and
[webhook signatures](/courses/learn-http-servers/authorization-and-webhooks/webhooks-and-api-keys);
this time let's look at it properly.

## Computing a tag

```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func main() {
	key := []byte("Jefe")
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("what do ya want for nothing?"))
	fmt.Printf("%x\n", mac.Sum(nil))
}
```

```
5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843
```

That exact value appears in **RFC 4231**, which publishes test vectors for HMAC-SHA-224,
256, 384 and 512. It's the known-answer test for your HMAC code.

- `hmac.New(h, key)` takes the hash's *constructor*, `sha256.New`, not a hash value.
  HMAC needs to create fresh hash states internally.
- The result is a `hash.Hash`, so it's an `io.Writer`: you can `io.Copy` a whole file
  into it, exactly as you did with SHA-256 in the last chapter. `Write` never fails.
- `Sum(nil)` returns the tag, 32 bytes for SHA-256. Like any hash, `Sum` *appends*, so
  pass `nil`.
- `Reset()` gets you back to the keyed initial state, ready for the next message with
  the same key.

## Keys

An HMAC key should be random and as long as the hash output: **32 bytes** from
`crypto/rand` for HMAC-SHA256. Shorter keys weaken it, and longer ones don't help much.
(Keys longer than the hash's 64-byte block size are hashed first, as RFC 4231's test
case 6 checks.) Passwords are not keys; run them through a KDF first (next chapter).

## Verifying a tag

To verify, recompute the tag and compare it with the one you received, using
**`hmac.Equal`**:

```go
func checkTag(key, msg, tag []byte) bool {
	return hmac.Equal(tag, tagMessage(key, msg))
}
```

`hmac.Equal` compares in **constant time**: it takes the same time whether the tags
differ in the first byte or the last. With `bytes.Equal` or `==`, the comparison stops
at the first difference, and that timing can leak how many leading bytes of a guess were
right. The next lesson explains why that's dangerous. `hmac.Equal` returns false for
slices of different lengths, so a truncated or empty tag can't match.

## Truncated tags

Some protocols send only part of the tag (say 16 bytes) to save space. That's
acceptable if it's in the spec and you check exactly that many bytes, but never accept
"whatever length the sender provided". If your verifier compared only
`len(receivedTag)` bytes, an empty tag would pass.

## Your task

Complete Keybox's two MAC helpers:

1. `tagMessage(key, msg)` returns HMAC-SHA256 of `msg` under `key`.
2. `checkTag(key, msg, tag)` recomputes the tag and compares it with `hmac.Equal`.

The tests use RFC 4231's vectors, and try modified messages, wrong keys, and truncated
and empty tags.
