---
title: Fix the Webhook Check
difficulty: easy
after: message-authentication
hints:
  - '`strings.HasPrefix(want, got)` is true whenever `got` is *any* prefix of the right answer, including the empty string. An attacker who sends `sha256=` gets in without knowing the secret.'
  - 'Don''t compare hex strings at all. `strings.CutPrefix` gives you the hex and whether the prefix was there; `hex.DecodeString` turns it into bytes (and accepts upper case); `hmac.Equal` then compares the raw 32-byte MACs in constant time and returns false for any length mismatch.'
exercise:
  starter: |
    package main

    import (
    	"crypto/hmac"
    	"crypto/sha256"
    	"encoding/hex"
    	"fmt"
    	"strings"
    )

    // verifyWebhook reports whether header is a valid signature of body under
    // secret. header must look like "sha256=<hex>", where <hex> is the full
    // HMAC-SHA256(secret, body) in upper or lower case.
    //
    // BUG: this version "works" for honest callers but accepts forged headers.
    // Rewrite the comparison:
    //  1. Require the "sha256=" prefix (strings.CutPrefix tells you if it was there).
    //  2. Decode the hex into bytes; invalid hex is a bad signature.
    //  3. Compare the decoded bytes with the expected MAC using hmac.Equal.
    func verifyWebhook(secret, body []byte, header string) bool {
    	mac := hmac.New(sha256.New, secret)
    	mac.Write(body)
    	want := hex.EncodeToString(mac.Sum(nil))
    	got := strings.TrimPrefix(header, "sha256=")
    	return strings.HasPrefix(want, got)
    }

    func main() {
    	secret := []byte("Jefe")
    	body := []byte("what do ya want for nothing?")
    	good := "sha256=5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843"
    	fmt.Println("genuine:", verifyWebhook(secret, body, good))      // want: true
    	fmt.Println("empty:  ", verifyWebhook(secret, body, "sha256=")) // want: false
    }
  solution: |
    package main

    import (
    	"crypto/hmac"
    	"crypto/sha256"
    	"encoding/hex"
    	"fmt"
    	"strings"
    )

    // verifyWebhook reports whether header is a valid signature of body under
    // secret. header must look like "sha256=<hex>", where <hex> is the full
    // HMAC-SHA256(secret, body) in upper or lower case.
    func verifyWebhook(secret, body []byte, header string) bool {
    	hexTag, ok := strings.CutPrefix(header, "sha256=")
    	if !ok {
    		return false
    	}
    	tag, err := hex.DecodeString(hexTag)
    	if err != nil {
    		return false
    	}
    	mac := hmac.New(sha256.New, secret)
    	mac.Write(body)
    	return hmac.Equal(tag, mac.Sum(nil))
    }

    func main() {
    	secret := []byte("Jefe")
    	body := []byte("what do ya want for nothing?")
    	good := "sha256=5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843"
    	fmt.Println("genuine:", verifyWebhook(secret, body, good))
    	fmt.Println("empty:  ", verifyWebhook(secret, body, "sha256="))
    }
  tests: |
    package main

    import (
    	"bytes"
    	"strings"
    	"testing"
    )

    // RFC 4231 test cases 1-3 for HMAC-SHA-256.
    var vectors = []struct {
    	name      string
    	key, data []byte
    	hexTag    string
    }{
    	{"RFC 4231 case 1", bytes.Repeat([]byte{0x0b}, 20), []byte("Hi There"),
    		"b0344c61d8db38535ca8afceaf0bf12b881dc200c9833da726e9376c2e32cff7"},
    	{"RFC 4231 case 2", []byte("Jefe"), []byte("what do ya want for nothing?"),
    		"5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843"},
    	{"RFC 4231 case 3", bytes.Repeat([]byte{0xaa}, 20), bytes.Repeat([]byte{0xdd}, 50),
    		"773ea91e36800e46854db8ebd09181a72959098b3ef8c122d9635514ced565fe"},
    }

    func TestAcceptsGenuine(t *testing.T) {
    	for _, v := range vectors {
    		for _, h := range []string{"sha256=" + v.hexTag, "sha256=" + strings.ToUpper(v.hexTag)} {
    			if !verifyWebhook(v.key, v.data, h) {
    				t.Errorf("%s: verifyWebhook rejected the genuine header %q", v.name, h)
    			}
    		}
    	}
    }

    func TestRejectsForgeries(t *testing.T) {
    	v := vectors[1]
    	full := v.hexTag
    	bad := []struct{ name, header string }{
    		{"empty tag", "sha256="},
    		{"empty header", ""},
    		{"missing prefix", full},
    		{"wrong algorithm", "sha1=" + full},
    		{"truncated to 32 hex", "sha256=" + full[:32]},
    		{"truncated by one byte", "sha256=" + full[:62]},
    		{"odd length", "sha256=" + full[:63]},
    		{"extra byte", "sha256=" + full + "00"},
    		{"not hex", "sha256=" + strings.Repeat("zz", 32)},
    		{"last bit flipped", "sha256=" + full[:63] + "2"},
    		{"prefix twice", "sha256=sha256=" + full},
    		{"space before", "sha256= " + full},
    	}
    	for _, b := range bad {
    		if verifyWebhook(v.key, v.data, b.header) {
    			t.Errorf("%s: verifyWebhook accepted %q", b.name, b.header)
    		}
    	}
    }

    func TestRejectsWrongKeyOrBody(t *testing.T) {
    	v := vectors[1]
    	h := "sha256=" + v.hexTag
    	if verifyWebhook([]byte("Jeff"), v.data, h) {
    		t.Errorf("verifyWebhook accepted a header made with a different secret")
    	}
    	if verifyWebhook(v.key, []byte("what do ya want for nothing!"), h) {
    		t.Errorf("verifyWebhook accepted a header for a modified body")
    	}
    	if verifyWebhook(v.key, nil, h) {
    		t.Errorf("verifyWebhook accepted a header for an empty body")
    	}
    }
---

Keybox's billing provider calls a Keybox **webhook** whenever a subscription changes.
Each request carries a header `X-Signature: sha256=<hex>`, the HMAC-SHA256 of the
request body under a secret the two companies share. Anyone on the internet can reach
the webhook, so the signature is the only thing standing between an attacker and a
free lifetime plan.

The current `verifyWebhook` accepts every genuine request, and plenty of forged ones.
Fix it so that it returns `true` **only** when `header` is exactly `sha256=` followed by
the full HMAC-SHA256(`secret`, `body`) in hex (upper or lower case), and compares the
MAC in **constant time**.

## Examples

With `secret = "Jefe"` and `body = "what do ya want for nothing?"` (RFC 4231, test case 2):

```
sha256=5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843  true
sha256=5BDCC146BF60754E6A042426089575C75A003F089D2739839DEC58B964EC3843  true
sha256=5bdcc146bf60754e6a042426089575c7                                  false (truncated)
sha256=                                                                  false
5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843         false (no prefix)
```

## Constraints

- The tests use the RFC 4231 HMAC-SHA-256 vectors, then try truncated, extended,
  non-hex, wrong-prefix and bit-flipped headers, a wrong secret and a modified body.
- No test can measure timing reliably, so it's on you: never compare MACs with `==`,
  `bytes.Equal` or string comparisons, which stop at the first differing byte.
