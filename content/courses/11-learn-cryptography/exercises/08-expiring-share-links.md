---
title: Expiring Share Links
difficulty: medium
after: message-authentication
hints:
  - 'Build the payload with `[]byte{0x01}`, `binary.BigEndian.AppendUint64(payload, uint64(expires.Unix()))` and `append(payload, secretID...)`. `mac.Sum(payload)` appends the tag to it, then base64url-encode the lot. The ID needs no length prefix: it''s whatever sits between the fixed 9-byte header and the fixed 32-byte tag.'
  - 'In `checkLink`, decode with `base64.RawURLEncoding`, require at least 9 + 32 bytes, split off the last 32 as the tag, and `hmac.Equal` it against a fresh MAC of the rest **before** reading the version, the expiry or the ID. Until the tag checks out, every field is attacker-controlled.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"time"
    )

    var (
    	ErrInvalidLink = errors.New("keybox: invalid share link")
    	ErrExpired     = errors.New("keybox: share link expired")
    )

    // makeLink returns a share-link token for secretID that expires at expires.
    func makeLink(key []byte, secretID string, expires time.Time) string {
    	return ""
    }

    // checkLink verifies token at time now and returns the secret ID it grants.
    func checkLink(key []byte, token string, now time.Time) (string, error) {
    	return "", ErrInvalidLink
    }

    func main() {
    	key := []byte("0123456789abcdef0123456789abcdef")
    	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    	tok := makeLink(key, "vault-7/github-token", now.Add(time.Hour))
    	fmt.Println("token:", tok)
    	fmt.Println(checkLink(key, tok, now))                  // want: vault-7/github-token <nil>
    	fmt.Println(checkLink(key, tok, now.Add(2*time.Hour))) // want: keybox: share link expired
    }
  solution: |
    package main

    import (
    	"crypto/hmac"
    	"crypto/sha256"
    	"encoding/base64"
    	"encoding/binary"
    	"errors"
    	"fmt"
    	"time"
    )

    var (
    	ErrInvalidLink = errors.New("keybox: invalid share link")
    	ErrExpired     = errors.New("keybox: share link expired")
    )

    const (
    	linkV1     = 0x01
    	headerSize = 1 + 8
    	tagSize    = sha256.Size
    )

    func linkTag(key, payload []byte) []byte {
    	mac := hmac.New(sha256.New, key)
    	mac.Write(payload)
    	return mac.Sum(nil)
    }

    // makeLink returns a share-link token for secretID that expires at expires.
    func makeLink(key []byte, secretID string, expires time.Time) string {
    	payload := []byte{linkV1}
    	payload = binary.BigEndian.AppendUint64(payload, uint64(expires.Unix()))
    	payload = append(payload, secretID...)
    	return base64.RawURLEncoding.EncodeToString(append(payload, linkTag(key, payload)...))
    }

    // checkLink verifies token at time now and returns the secret ID it grants.
    func checkLink(key []byte, token string, now time.Time) (string, error) {
    	raw, err := base64.RawURLEncoding.DecodeString(token)
    	if err != nil || len(raw) < headerSize+tagSize {
    		return "", ErrInvalidLink
    	}
    	payload, tag := raw[:len(raw)-tagSize], raw[len(raw)-tagSize:]
    	if !hmac.Equal(tag, linkTag(key, payload)) {
    		return "", ErrInvalidLink
    	}
    	if payload[0] != linkV1 {
    		return "", ErrInvalidLink
    	}
    	expires := int64(binary.BigEndian.Uint64(payload[1:headerSize]))
    	if now.Unix() >= expires {
    		return "", ErrExpired
    	}
    	return string(payload[headerSize:]), nil
    }

    func main() {
    	key := []byte("0123456789abcdef0123456789abcdef")
    	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    	tok := makeLink(key, "vault-7/github-token", now.Add(time.Hour))
    	fmt.Println("token:", tok)
    	fmt.Println(checkLink(key, tok, now))
    	fmt.Println(checkLink(key, tok, now.Add(2*time.Hour)))
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/hmac"
    	"crypto/sha256"
    	"encoding/base64"
    	"encoding/binary"
    	"errors"
    	"testing"
    	"time"
    )

    var (
    	key     = []byte("0123456789abcdef0123456789abcdef")
    	now     = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    	expires = now.Add(time.Hour)
    )

    // reference builds a token exactly as specified, with any version byte.
    func reference(key []byte, version byte, id string, exp time.Time) string {
    	payload := []byte{version}
    	payload = binary.BigEndian.AppendUint64(payload, uint64(exp.Unix()))
    	payload = append(payload, id...)
    	mac := hmac.New(sha256.New, key)
    	mac.Write(payload)
    	return base64.RawURLEncoding.EncodeToString(mac.Sum(payload))
    }

    func TestMakeLinkFormat(t *testing.T) {
    	for _, id := range []string{"vault-7/github-token", "x", ""} {
    		want := reference(key, 1, id, expires)
    		if got := makeLink(key, id, expires); got != want {
    			t.Errorf("makeLink(key, %q, %v) =\n  %q\nwant\n  %q", id, expires, got, want)
    		}
    	}
    }

    func TestRoundTrip(t *testing.T) {
    	for _, id := range []string{"vault-7/github-token", "a", "ünïcødé/🔑", string(bytes.Repeat([]byte("z"), 500))} {
    		tok := makeLink(key, id, expires)
    		if got, err := checkLink(key, tok, now); got != id || err != nil {
    			t.Errorf("checkLink(makeLink(%.30q)) = %.30q, %v, want the same ID and nil", id, got, err)
    		}
    	}
    }

    func TestExpiry(t *testing.T) {
    	tok := makeLink(key, "vault-7/db", expires)
    	if _, err := checkLink(key, tok, expires.Add(-time.Second)); err != nil {
    		t.Errorf("one second before expiry: checkLink error = %v, want nil", err)
    	}
    	for _, at := range []time.Time{expires, expires.Add(time.Second), expires.Add(24 * time.Hour)} {
    		if got, err := checkLink(key, tok, at); !errors.Is(err, ErrExpired) || got != "" {
    			t.Errorf("checkLink at %v for a link expiring at %v = %q, %v, want \"\", ErrExpired", at, expires, got, err)
    		}
    	}
    }

    func TestEveryByteTampered(t *testing.T) {
    	raw := mustDecode(t, makeLink(key, "vault-7/db", expires))
    	if len(raw) != 1+8+len("vault-7/db")+32 {
    		t.Fatalf("decoded token is %d bytes, want 1+8+%d+32", len(raw), len("vault-7/db"))
    	}
    	for i := range raw {
    		bad := bytes.Clone(raw)
    		bad[i] ^= 0x01
    		tok := base64.RawURLEncoding.EncodeToString(bad)
    		if got, err := checkLink(key, tok, now); !errors.Is(err, ErrInvalidLink) {
    			t.Fatalf("flipping one bit in byte %d of the token: checkLink = %q, %v, want ErrInvalidLink", i, got, err)
    		}
    	}
    }

    func TestForgedExtensionIsInvalidNotExpired(t *testing.T) {
    	// An old, expired link with its expiry pushed forward but the tag unchanged.
    	raw := mustDecode(t, makeLink(key, "vault-7/db", now.Add(-time.Hour)))
    	binary.BigEndian.PutUint64(raw[1:9], uint64(now.Add(time.Hour).Unix()))
    	if _, err := checkLink(key, base64.RawURLEncoding.EncodeToString(raw), now); !errors.Is(err, ErrInvalidLink) {
    		t.Errorf("expired link with a forged later expiry: err = %v, want ErrInvalidLink", err)
    	}
    	// A forged link that is also expired must say invalid, not expired:
    	// check the tag before trusting any field.
    	raw = mustDecode(t, makeLink(key, "vault-7/db", now.Add(-time.Hour)))
    	raw[len(raw)-1] ^= 0x80
    	if _, err := checkLink(key, base64.RawURLEncoding.EncodeToString(raw), now); !errors.Is(err, ErrInvalidLink) {
    		t.Errorf("expired link with a bad tag: err = %v, want ErrInvalidLink (verify the tag first)", err)
    	}
    }

    func TestWrongKeyAndVersion(t *testing.T) {
    	other := bytes.Clone(key)
    	other[0] ^= 1
    	if got, err := checkLink(other, makeLink(key, "vault-7/db", expires), now); !errors.Is(err, ErrInvalidLink) {
    		t.Errorf("checkLink with the wrong key = %q, %v, want ErrInvalidLink", got, err)
    	}
    	for _, v := range []byte{0, 2, 0xff} {
    		tok := reference(key, v, "vault-7/db", expires) // correctly MACed, unknown version
    		if got, err := checkLink(key, tok, now); !errors.Is(err, ErrInvalidLink) {
    			t.Errorf("token with version byte %d: checkLink = %q, %v, want ErrInvalidLink", v, got, err)
    		}
    	}
    }

    func TestTruncatedAndMalformed(t *testing.T) {
    	tok := makeLink(key, "vault-7/db", expires)
    	for n := range len(tok) {
    		if got, err := checkLink(key, tok[:n], now); !errors.Is(err, ErrInvalidLink) {
    			t.Fatalf("token cut to %d of %d characters: checkLink = %q, %v, want ErrInvalidLink", n, len(tok), got, err)
    		}
    	}
    	std := base64.StdEncoding.EncodeToString(mustDecode(t, tok))
    	for _, bad := range []string{"", "!!!!", tok + "A", tok + "=", std + "=", " " + tok} {
    		if got, err := checkLink(key, bad, now); !errors.Is(err, ErrInvalidLink) {
    			t.Errorf("checkLink(%q) = %q, %v, want ErrInvalidLink", bad, got, err)
    		}
    	}
    }

    func mustDecode(t *testing.T, s string) []byte {
    	t.Helper()
    	b, err := base64.RawURLEncoding.DecodeString(s)
    	if err != nil {
    		t.Fatalf("makeLink returned %q, which isn't unpadded base64url: %v", s, err)
    	}
    	if len(b) < 1+8+32 {
    		t.Fatalf("makeLink returned a %d-byte token, want at least 1+8+32", len(b))
    	}
    	return b
    }
---

Keybox can share a single secret through a link such as
`https://keybox.test/s/<token>`. The server keeps no table of issued links: the token
itself says which secret it grants and until when, and an HMAC under a server key
makes sure nobody can change either.

Implement:

- `makeLink(key, secretID, expires)` returns the token: the unpadded base64url
  (`base64.RawURLEncoding`) encoding of

  ```
  0x01 (version) || expires as uint64 Unix seconds, big-endian (8 bytes) || secretID || tag
  tag = HMAC-SHA256(key, everything before the tag)
  ```

- `checkLink(key, token, now)` returns the secret ID and `nil` if the token is
  genuine and `now` is **before** its expiry. It returns `ErrExpired` for a genuine
  token at or after its expiry, and `ErrInvalidLink` for everything else: bad
  base64, too short, wrong tag (tampering, wrong key, truncation), or an unknown
  version byte.

## Example

```go
now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
tok := makeLink(key, "vault-7/github-token", now.Add(time.Hour))
checkLink(key, tok, now)                     // "vault-7/github-token", nil
checkLink(key, tok, now.Add(time.Hour))      // "", ErrExpired
checkLink(otherKey, tok, now)                // "", ErrInvalidLink
```

## Constraints

- The tests compare `makeLink`'s output with a reference implementation of the format
  byte for byte, so follow it exactly.
- They flip every bit of a token, cut it to every shorter length, try a wrong key,
  bad base64 and correctly MACed tokens with an unknown version, and push an expired
  link's expiry forward.
- A **forged** token that also happens to be expired must report `ErrInvalidLink`:
  verify the tag, in constant time, before you trust any field.
