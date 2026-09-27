---
title: Verifying MACs Safely
quiz:
  - question: A verifier checks `hmac.Equal(got, want[:len(got)])`, so that clients may send shortened tags. What can Mallory do?
    options:
      - text: Nothing; `hmac.Equal` is constant-time
      - text: Send an empty tag, which compares equal to `want[:0]`, and have any request accepted
        correct: true
      - text: Only replay old requests
      - text: Recover the key
    explanation: |
      Two empty slices are equal. The verifier lets the *attacker* choose how much of
      the tag to check, and the attacker chooses zero. The expected tag length must be
      fixed by the protocol, never by the input.
exercise:
  starter: |
    package main

    import (
    	"crypto/hmac"
    	"crypto/sha256"
    	"encoding/binary"
    	"encoding/hex"
    	"errors"
    	"fmt"
    	"time"
    )

    var (
    	ErrBadSignature = errors.New("bad signature")
    	ErrStale        = errors.New("request timestamp outside the allowed window")
    )

    const maxSkew = 5 * time.Minute

    type syncRequest struct {
    	Method    string
    	Path      string
    	Timestamp int64 // Unix seconds, set by the client
    	Body      []byte
    }

    // signRequest MACs every field with length prefixes, under a label.
    // It is correct; don't change it.
    func signRequest(key []byte, r syncRequest) string {
    	mac := hmac.New(sha256.New, key)
    	for _, f := range [][]byte{
    		[]byte("keybox sync request v1"),
    		[]byte(r.Method),
    		[]byte(r.Path),
    		binary.BigEndian.AppendUint64(nil, uint64(r.Timestamp)),
    		r.Body,
    	} {
    		mac.Write(binary.BigEndian.AppendUint64(nil, uint64(len(f))))
    		mac.Write(f)
    	}
    	return hex.EncodeToString(mac.Sum(nil))
    }

    // verifyRequest checks sig and that r.Timestamp is within maxSkew of now,
    // in either direction. It has two security bugs. Find and fix them.
    func verifyRequest(key []byte, r syncRequest, sig string, now time.Time) error {
    	got, err := hex.DecodeString(sig)
    	if err != nil {
    		return ErrBadSignature
    	}
    	want, _ := hex.DecodeString(signRequest(key, r))
    	if len(got) > len(want) || !hmac.Equal(got, want[:len(got)]) {
    		return ErrBadSignature
    	}
    	if now.Sub(time.Unix(r.Timestamp, 0)) > 24*time.Hour {
    		return ErrStale
    	}
    	return nil
    }

    func main() {
    	key := []byte("0123456789abcdef0123456789abcdef")
    	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
    	req := syncRequest{"PUT", "/v1/secrets/github-token", now.Unix(), []byte(`{"sealed":"..."}`)}
    	sig := signRequest(key, req)

    	fmt.Println("valid:", verifyRequest(key, req, sig, now))
    	fmt.Println("empty signature:", verifyRequest(key, req, "", now))
    	fmt.Println("replayed an hour later:", verifyRequest(key, req, sig, now.Add(time.Hour)))
    }
  solution: |
    package main

    import (
    	"crypto/hmac"
    	"crypto/sha256"
    	"encoding/binary"
    	"encoding/hex"
    	"errors"
    	"fmt"
    	"time"
    )

    var (
    	ErrBadSignature = errors.New("bad signature")
    	ErrStale        = errors.New("request timestamp outside the allowed window")
    )

    const maxSkew = 5 * time.Minute

    type syncRequest struct {
    	Method    string
    	Path      string
    	Timestamp int64
    	Body      []byte
    }

    func signRequest(key []byte, r syncRequest) string {
    	mac := hmac.New(sha256.New, key)
    	for _, f := range [][]byte{
    		[]byte("keybox sync request v1"),
    		[]byte(r.Method),
    		[]byte(r.Path),
    		binary.BigEndian.AppendUint64(nil, uint64(r.Timestamp)),
    		r.Body,
    	} {
    		mac.Write(binary.BigEndian.AppendUint64(nil, uint64(len(f))))
    		mac.Write(f)
    	}
    	return hex.EncodeToString(mac.Sum(nil))
    }

    func verifyRequest(key []byte, r syncRequest, sig string, now time.Time) error {
    	got, err := hex.DecodeString(sig)
    	if err != nil {
    		return ErrBadSignature
    	}
    	want, _ := hex.DecodeString(signRequest(key, r))
    	if !hmac.Equal(got, want) {
    		return ErrBadSignature
    	}
    	skew := now.Sub(time.Unix(r.Timestamp, 0))
    	if skew > maxSkew || skew < -maxSkew {
    		return ErrStale
    	}
    	return nil
    }

    func main() {
    	key := []byte("0123456789abcdef0123456789abcdef")
    	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
    	req := syncRequest{"PUT", "/v1/secrets/github-token", now.Unix(), []byte(`{"sealed":"..."}`)}
    	sig := signRequest(key, req)

    	fmt.Println("valid:", verifyRequest(key, req, sig, now))
    	fmt.Println("empty signature:", verifyRequest(key, req, "", now))
    	fmt.Println("replayed an hour later:", verifyRequest(key, req, sig, now.Add(time.Hour)))
    }
  tests: |
    package main

    import (
    	"errors"
    	"testing"
    	"time"
    )

    var (
    	testKey = []byte("test key for keybox sync, 32 by!")
    	testNow = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
    	testReq = syncRequest{"PUT", "/v1/secrets/aws", testNow.Unix(), []byte(`{"sealed":"AAAA"}`)}
    )

    func TestValidRequests(t *testing.T) {
    	sig := signRequest(testKey, testReq)
    	for _, d := range []time.Duration{0, time.Minute, 4 * time.Minute, -4 * time.Minute, maxSkew, -maxSkew} {
    		if err := verifyRequest(testKey, testReq, sig, testNow.Add(d)); err != nil {
    			t.Errorf("verifyRequest(valid request, checked %v after it was signed) = %v, want nil", d, err)
    		}
    	}
    }

    func TestBadSignatures(t *testing.T) {
    	sig := signRequest(testKey, testReq)
    	tampered := testReq
    	tampered.Body = []byte(`{"sealed":"BBBB"}`)
    	otherPath := testReq
    	otherPath.Path = "/v1/secrets/github"
    	for _, tt := range []struct {
    		name string
    		req  syncRequest
    		sig  string
    	}{
    		{"empty signature", testReq, ""},
    		{"first byte of the signature only", testReq, sig[:2]},
    		{"half the signature", testReq, sig[:32]},
    		{"signature with extra bytes", testReq, sig + "00"},
    		{"tampered body", tampered, sig},
    		{"different path", otherPath, sig},
    		{"signed with another key", testReq, signRequest([]byte("some other key, also 32 bytes!!!"), testReq)},
    		{"not hex", testReq, "zz"},
    	} {
    		if err := verifyRequest(testKey, tt.req, tt.sig, testNow); !errors.Is(err, ErrBadSignature) {
    			t.Errorf("%s: verifyRequest = %v, want ErrBadSignature", tt.name, err)
    		}
    	}
    }

    func TestStaleRequests(t *testing.T) {
    	sig := signRequest(testKey, testReq)
    	for _, d := range []time.Duration{maxSkew + time.Second, time.Hour, 23 * time.Hour, -maxSkew - time.Second, -time.Hour} {
    		if err := verifyRequest(testKey, testReq, sig, testNow.Add(d)); !errors.Is(err, ErrStale) {
    			t.Errorf("verifyRequest(valid signature, checked %v after signing) = %v, want ErrStale", d, err)
    		}
    	}
    }
---

You know the pieces: HMAC over an unambiguous encoding, compared with `hmac.Equal`. Yet
real verifiers still get broken, almost always in the code *around* the MAC. You met
one such design in [Webhooks and API Keys](/courses/learn-http-servers/authorization-and-webhooks/webhooks-and-api-keys),
where a timestamp inside the signed data stopped replays. Here's a checklist for
reviewing any verifier.

## 1. Cover everything that matters

The tag must cover every field the receiver acts on: method, path, query, relevant
headers, body, timestamp. Anything outside the MAC is attacker-controlled. A classic bug
signs the body but not the path, so Mallory replays a valid `PUT` body against a
different URL. Encode the fields unambiguously (length prefixes, as in chapter 3), and
start with a label like `"keybox sync request v1"`.

## 2. The protocol fixes the tag length

Compare the *whole* expected tag. Code that "supports truncated tags" by comparing only
as many bytes as the client sent lets the client send zero bytes. `hmac.Equal` already
returns false for different lengths, so just don't slice the expected value.

## 3. Verify before you trust

Check the MAC **before** parsing, logging, or acting on the payload. Until the tag
verifies, every field is attacker-controlled, including the timestamp and any
"algorithm" or "key ID" header.

## 4. Reject replays

A valid request stays valid forever unless you stop it. Include a timestamp in the
signed data and reject requests outside a small window, in **both** directions: a
timestamp far in the future is as suspicious as one far in the past, since an attacker
who got a future-dated request signed could replay it for a long time. For stronger
protection, also include a random request ID and remember the IDs you've seen within the
window.

## 5. Fail closed, with one error

On *any* problem (bad hex, wrong length, wrong tag) return the same error. Don't fall
back to an unsigned code path, and don't tell the client *which* check failed. Detailed
reasons belong in your server logs.

```go
if !hmac.Equal(got, want) {
	return ErrBadSignature // same error for every signature failure
}
```

## Your task

Keybox's sync server verifies client requests with `verifyRequest`. `signRequest` is
correct. `verifyRequest` has **two** security bugs:

1. It compares only as many bytes of the tag as the client sent.
2. Its freshness check allows requests up to a day old, and any amount in the future.

Fix both. Compare the full tag with `hmac.Equal`, then return `ErrStale` if the
timestamp is more than `maxSkew` away from `now` in either direction (exactly
`maxSkew` is still allowed). Signature failures must return `ErrBadSignature`.
