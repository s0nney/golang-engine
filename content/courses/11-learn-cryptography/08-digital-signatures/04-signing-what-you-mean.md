---
title: Signing What You Mean
quiz:
  - question: 'Keybox signs `json.Marshal(bundle)`. The verifier parses the JSON into a struct, calls `json.Marshal` again, and verifies the signature against *that*. What''s the problem?'
    options:
      - text: Nothing; JSON encoding is deterministic
      - text: The re-encoded bytes may differ from what was signed (key order, whitespace, escaping, other encoders), so genuine bundles can fail, and fields the struct ignores were never checked at all
        correct: true
      - text: json.Marshal is too slow
      - text: Signatures can't cover JSON
    explanation: |
      Verify the exact bytes you received, then parse those same bytes. Don't
      re-serialize and hope for a match. Anything the parser sees must be covered by
      the signature, and the easiest way to guarantee it is to parse only verified bytes.
exercise:
  starter: |
    package main

    import (
    	"crypto/ed25519"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"time"
    )

    var (
    	ErrBadSignature   = errors.New("keybox: bad bundle signature")
    	ErrWrongRecipient = errors.New("keybox: bundle is for someone else")
    	ErrExpired        = errors.New("keybox: bundle is too old or from the future")
    )

    const (
    	bundleContext = "keybox bundle v1\n"
    	maxAge        = 7 * 24 * time.Hour
    	maxSkew       = 5 * time.Minute
    )

    type Bundle struct {
    	From    string   `json:"from"`
    	To      string   `json:"to"`
    	Created int64    `json:"created"` // Unix seconds
    	Shares  [][]byte `json:"shares"`  // HPKE shares, one per secret
    }

    // SignedBundle is what gets exported: the exact payload bytes and an
    // Ed25519 signature over bundleContext + payload.
    type SignedBundle struct {
    	Payload   []byte `json:"payload"`
    	Signature []byte `json:"signature"`
    }

    func exportBundle(priv ed25519.PrivateKey, b Bundle) (SignedBundle, error) {
    	payload, err := json.Marshal(b)
    	if err != nil {
    		return SignedBundle{}, err
    	}
    	sig := ed25519.Sign(priv, append([]byte(bundleContext), payload...))
    	return SignedBundle{Payload: payload, Signature: sig}, nil
    }

    // importBundle verifies sb against the sender's public key and returns the
    // bundle if, in this order:
    //  1. pub is a valid-length key and the signature over bundleContext +
    //     Payload verifies (else ErrBadSignature; also for bad JSON),
    //  2. the bundle's To is me (else ErrWrongRecipient),
    //  3. Created is no more than maxAge in the past and no more than maxSkew
    //     in the future, relative to now (else ErrExpired).
    func importBundle(pub ed25519.PublicKey, sb SignedBundle, me string, now time.Time) (Bundle, error) {
    	// ? This version trusts the payload without checking anything.
    	var b Bundle
    	err := json.Unmarshal(sb.Payload, &b)
    	return b, err
    }

    func main() {
    	alicePub, alicePriv, _ := ed25519.GenerateKey(nil)
    	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
    	sb, _ := exportBundle(alicePriv, Bundle{From: "alice", To: "bob", Created: now.Unix(), Shares: [][]byte{[]byte("share-1")}})

    	b, err := importBundle(alicePub, sb, "bob", now.Add(time.Hour))
    	fmt.Printf("bob imports: from %s, %d shares, err %v\n", b.From, len(b.Shares), err)

    	_, err = importBundle(alicePub, sb, "carol", now.Add(time.Hour))
    	fmt.Println("carol imports:", err)

    	forged := sb
    	forged.Payload = []byte(`{"from":"alice","to":"bob","created":1790510400,"shares":["bWFsbG9yeQ=="]}`)
    	_, err = importBundle(alicePub, forged, "bob", now.Add(time.Hour))
    	fmt.Println("forged payload:", err)
    }
  solution: |
    package main

    import (
    	"crypto/ed25519"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"time"
    )

    var (
    	ErrBadSignature   = errors.New("keybox: bad bundle signature")
    	ErrWrongRecipient = errors.New("keybox: bundle is for someone else")
    	ErrExpired        = errors.New("keybox: bundle is too old or from the future")
    )

    const (
    	bundleContext = "keybox bundle v1\n"
    	maxAge        = 7 * 24 * time.Hour
    	maxSkew       = 5 * time.Minute
    )

    type Bundle struct {
    	From    string   `json:"from"`
    	To      string   `json:"to"`
    	Created int64    `json:"created"`
    	Shares  [][]byte `json:"shares"`
    }

    type SignedBundle struct {
    	Payload   []byte `json:"payload"`
    	Signature []byte `json:"signature"`
    }

    func exportBundle(priv ed25519.PrivateKey, b Bundle) (SignedBundle, error) {
    	payload, err := json.Marshal(b)
    	if err != nil {
    		return SignedBundle{}, err
    	}
    	sig := ed25519.Sign(priv, append([]byte(bundleContext), payload...))
    	return SignedBundle{Payload: payload, Signature: sig}, nil
    }

    func importBundle(pub ed25519.PublicKey, sb SignedBundle, me string, now time.Time) (Bundle, error) {
    	if len(pub) != ed25519.PublicKeySize {
    		return Bundle{}, ErrBadSignature
    	}
    	signed := append([]byte(bundleContext), sb.Payload...)
    	if !ed25519.Verify(pub, signed, sb.Signature) {
    		return Bundle{}, ErrBadSignature
    	}
    	var b Bundle
    	if err := json.Unmarshal(sb.Payload, &b); err != nil {
    		return Bundle{}, ErrBadSignature
    	}
    	if b.To != me {
    		return Bundle{}, ErrWrongRecipient
    	}
    	age := now.Sub(time.Unix(b.Created, 0))
    	if age > maxAge || age < -maxSkew {
    		return Bundle{}, ErrExpired
    	}
    	return b, nil
    }

    func main() {
    	alicePub, alicePriv, _ := ed25519.GenerateKey(nil)
    	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
    	sb, _ := exportBundle(alicePriv, Bundle{From: "alice", To: "bob", Created: now.Unix(), Shares: [][]byte{[]byte("share-1")}})

    	b, err := importBundle(alicePub, sb, "bob", now.Add(time.Hour))
    	fmt.Printf("bob imports: from %s, %d shares, err %v\n", b.From, len(b.Shares), err)

    	_, err = importBundle(alicePub, sb, "carol", now.Add(time.Hour))
    	fmt.Println("carol imports:", err)

    	forged := sb
    	forged.Payload = []byte(`{"from":"alice","to":"bob","created":1790510400,"shares":["bWFsbG9yeQ=="]}`)
    	_, err = importBundle(alicePub, forged, "bob", now.Add(time.Hour))
    	fmt.Println("forged payload:", err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/ed25519"
    	"errors"
    	"testing"
    	"time"
    )

    var (
    	testSeed = bytes.Repeat([]byte{0xa1}, 32)
    	testPriv = ed25519.NewKeyFromSeed(testSeed)
    	testPub  = testPriv.Public().(ed25519.PublicKey)
    	testNow  = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
    )

    func testBundle(to string, created time.Time) SignedBundle {
    	sb, _ := exportBundle(testPriv, Bundle{From: "alice", To: to, Created: created.Unix(), Shares: [][]byte{[]byte("s1"), []byte("s2")}})
    	return sb
    }

    func TestImportValid(t *testing.T) {
    	for _, d := range []time.Duration{0, time.Hour, 6 * 24 * time.Hour, -4 * time.Minute} {
    		b, err := importBundle(testPub, testBundle("bob", testNow), "bob", testNow.Add(d))
    		if err != nil {
    			t.Fatalf("importBundle(valid bundle, %v after creation) error = %v", d, err)
    		}
    		if b.From != "alice" || b.To != "bob" || len(b.Shares) != 2 {
    			t.Errorf("importBundle returned %+v, want the exported bundle", b)
    		}
    	}
    }

    func TestImportBadSignatures(t *testing.T) {
    	sb := testBundle("bob", testNow)
    	_, otherPriv, _ := ed25519.GenerateKey(nil)
    	byOther, _ := exportBundle(otherPriv, Bundle{From: "alice", To: "bob", Created: testNow.Unix()})
    	noContext := SignedBundle{Payload: sb.Payload, Signature: ed25519.Sign(testPriv, sb.Payload)}
    	tampered := SignedBundle{Payload: bytes.Replace(sb.Payload, []byte(`"to":"bob"`), []byte(`"to":"eve"`), 1), Signature: sb.Signature}
    	garbage := SignedBundle{Payload: []byte("not json"), Signature: ed25519.Sign(testPriv, []byte(bundleContext+"not json"))}
    	for name, bad := range map[string]SignedBundle{
    		"tampered payload":             tampered,
    		"signed by another key":        byOther,
    		"signed without bundleContext": noContext,
    		"empty signature":              {Payload: sb.Payload},
    		"validly signed garbage":       garbage,
    	} {
    		if _, err := importBundle(testPub, bad, "bob", testNow); !errors.Is(err, ErrBadSignature) {
    			t.Errorf("%s: importBundle error = %v, want ErrBadSignature", name, err)
    		}
    	}
    	func() {
    		defer func() {
    			if r := recover(); r != nil {
    				t.Errorf("importBundle with a 5-byte public key panicked: %v", r)
    			}
    		}()
    		if _, err := importBundle(testPub[:5], sb, "bob", testNow); !errors.Is(err, ErrBadSignature) {
    			t.Errorf("importBundle with a 5-byte public key error = %v, want ErrBadSignature", err)
    		}
    	}()
    }

    func TestImportChecksClaims(t *testing.T) {
    	if _, err := importBundle(testPub, testBundle("bob", testNow), "carol", testNow); !errors.Is(err, ErrWrongRecipient) {
    		t.Errorf("carol importing bob's bundle: error = %v, want ErrWrongRecipient", err)
    	}
    	for _, d := range []time.Duration{maxAge + time.Second, 30 * 24 * time.Hour, -maxSkew - time.Second, -time.Hour} {
    		if _, err := importBundle(testPub, testBundle("bob", testNow), "bob", testNow.Add(d)); !errors.Is(err, ErrExpired) {
    			t.Errorf("importing %v after creation: error = %v, want ErrExpired", d, err)
    		}
    	}
    }
---

Ed25519 guarantees that a signature matches some bytes. Whether those bytes mean what
you think is up to you. Most signature bugs in real systems live here, not in the math.

## Verify the bytes you received

The number one rule: **verify the exact bytes, then parse those same bytes.** Keybox
exports a bundle as two fields, the JSON `Payload` exactly as it was signed, and the
`Signature`:

```go
type SignedBundle struct {
	Payload   []byte `json:"payload"`   // the signed bytes, untouched
	Signature []byte `json:"signature"`
}
```

The importer verifies `Payload` first and only then unmarshals it. Two tempting
alternatives are both wrong:

- **Re-encoding before verifying.** Parse the JSON, `json.Marshal` it again, verify
  that. JSON isn't canonical: key order, spacing, escaping and number formats can all
  differ between encoders, so genuine bundles fail. And fields your struct doesn't have
  are silently dropped before verification, so the verifier and some other consumer
  can disagree about what was signed.
- **Signing a struct's fields separately** and forgetting one. Every field you act on
  must be inside the signed bytes.

Until the signature checks out, don't log the payload's contents, don't look up the
"from" user it claims, don't do anything with it.

## Context: which protocol is this signature for?

Alice's Ed25519 key signs bundles. If later Keybox also uses it to sign, say, device
enrolment messages, an attacker could take a signature from one context and present it
in the other, whenever the byte strings happen to parse in both. Prevent this with a
**context prefix**: sign `"keybox bundle v1\n" + payload`, not just the payload.
It's the label idea from chapters 3 to 5 again, and since the prefix is never
transmitted, a signature made for any other context simply doesn't verify.

(Ed25519ctx and ML-DSA have a built-in context parameter for the same purpose. A prefix
works with plain Ed25519, which every library supports.)

## Claims to check after verifying

A valid signature only says Alice signed this. The bundle's contents then need checking
like any input:

- **Recipient.** `To` must be the importing user, or Bob could forward Alice's bundle
  to Carol as if Alice had sent it to her.
- **Freshness.** `Created` must be recent, and not in the future, or old bundles can be
  replayed indefinitely.
- **Sender.** The `From` field should match the identity whose public key you verified
  with, rather than being used to *choose* which key to verify with.

## Your task

Complete `importBundle(pub, sb, me, now)`. `exportBundle` shows how bundles are signed.

1. If `pub` isn't `ed25519.PublicKeySize` bytes, or the signature over
   `bundleContext + Payload` doesn't verify, return `ErrBadSignature`.
2. Only then unmarshal `Payload` (a JSON error is also `ErrBadSignature`).
3. If `To != me`, return `ErrWrongRecipient`.
4. If `Created` is more than `maxAge` before `now`, or more than `maxSkew` after it,
   return `ErrExpired`.

Return a zero `Bundle` with every error. Careful when building the signed message:
`append([]byte(bundleContext), sb.Payload...)` is safe because the conversion makes a
fresh slice.
