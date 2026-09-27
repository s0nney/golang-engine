---
title: Signed Access Tokens
quiz:
  - question: Why should `validateToken` check the signature *before* looking at `exp`?
    options:
      - text: It's faster
      - text: Until the signature is verified, every claim (including `exp`) is attacker-controlled and means nothing
        correct: true
      - text: The `exp` claim is inside the signature
      - text: The order doesn't matter at all
    explanation: |
      An unverified payload could say anything. Reporting "expired" for a forged token also
      tells an attacker something about how you parse tokens. Verify first, then trust the
      claims.
exercise:
  starter: |
    package main

    import (
    	"crypto/hmac"
    	"crypto/sha256"
    	"encoding/base64"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"strings"
    	"time"
    	"uuid"
    )

    var (
    	ErrInvalidToken = errors.New("invalid token")
    	ErrExpiredToken = errors.New("token expired")
    )

    var b64 = base64.RawURLEncoding

    const issuer = "squeak"

    type tokenHeader struct {
    	Alg string `json:"alg"`
    	Typ string `json:"typ"`
    }

    type claims struct {
    	Issuer    string `json:"iss"`
    	Subject   string `json:"sub"`
    	IssuedAt  int64  `json:"iat"`
    	ExpiresAt int64  `json:"exp"`
    }

    // sign returns HMAC-SHA256(secret, msg).
    func sign(secret []byte, msg string) []byte {
    	mac := hmac.New(sha256.New, secret)
    	mac.Write([]byte(msg))
    	return mac.Sum(nil)
    }

    // makeToken issues an HS256 JWT for userID, valid for ttl from now.
    func makeToken(userID uuid.UUID, secret []byte, now time.Time, ttl time.Duration) (string, error) {
    	header, err := json.Marshal(tokenHeader{Alg: "HS256", Typ: "JWT"})
    	if err != nil {
    		return "", err
    	}
    	payload, err := json.Marshal(claims{
    		Issuer:    issuer,
    		Subject:   userID.String(),
    		IssuedAt:  now.Unix(),
    		ExpiresAt: now.Add(ttl).Unix(),
    	})
    	if err != nil {
    		return "", err
    	}
    	unsigned := b64.EncodeToString(header) + "." + b64.EncodeToString(payload)
    	return unsigned + "." + b64.EncodeToString(sign(secret, unsigned)), nil
    }

    // validateToken checks token and returns the user ID it was issued for.
    // Forged, tampered or malformed tokens give ErrInvalidToken; genuine
    // tokens whose exp is not after now give ErrExpiredToken.
    func validateToken(token string, secret []byte, now time.Time) (uuid.UUID, error) {
    	// ?
    	return uuid.Nil(), nil
    }

    func main() {
    	secret := []byte("squeak-dev-secret-change-me")
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	issued := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)

    	token, _ := makeToken(pip, secret, issued, time.Hour)
    	fmt.Println("token:", token)

    	id, err := validateToken(token, secret, issued.Add(30*time.Minute))
    	fmt.Println("after 30m:", id, err)

    	_, err = validateToken(token, secret, issued.Add(2*time.Hour))
    	fmt.Println("after 2h:", err)

    	_, err = validateToken(token, []byte("wrong-secret"), issued)
    	fmt.Println("wrong secret:", err)

    	_, _ = strings.Split, hmac.Equal
    }
  solution: |
    package main

    import (
    	"crypto/hmac"
    	"crypto/sha256"
    	"encoding/base64"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"strings"
    	"time"
    	"uuid"
    )

    var (
    	ErrInvalidToken = errors.New("invalid token")
    	ErrExpiredToken = errors.New("token expired")
    )

    var b64 = base64.RawURLEncoding

    const issuer = "squeak"

    type tokenHeader struct {
    	Alg string `json:"alg"`
    	Typ string `json:"typ"`
    }

    type claims struct {
    	Issuer    string `json:"iss"`
    	Subject   string `json:"sub"`
    	IssuedAt  int64  `json:"iat"`
    	ExpiresAt int64  `json:"exp"`
    }

    func sign(secret []byte, msg string) []byte {
    	mac := hmac.New(sha256.New, secret)
    	mac.Write([]byte(msg))
    	return mac.Sum(nil)
    }

    func makeToken(userID uuid.UUID, secret []byte, now time.Time, ttl time.Duration) (string, error) {
    	header, err := json.Marshal(tokenHeader{Alg: "HS256", Typ: "JWT"})
    	if err != nil {
    		return "", err
    	}
    	payload, err := json.Marshal(claims{
    		Issuer:    issuer,
    		Subject:   userID.String(),
    		IssuedAt:  now.Unix(),
    		ExpiresAt: now.Add(ttl).Unix(),
    	})
    	if err != nil {
    		return "", err
    	}
    	unsigned := b64.EncodeToString(header) + "." + b64.EncodeToString(payload)
    	return unsigned + "." + b64.EncodeToString(sign(secret, unsigned)), nil
    }

    func validateToken(token string, secret []byte, now time.Time) (uuid.UUID, error) {
    	parts := strings.Split(token, ".")
    	if len(parts) != 3 {
    		return uuid.Nil(), ErrInvalidToken
    	}

    	headerJSON, err := b64.DecodeString(parts[0])
    	if err != nil {
    		return uuid.Nil(), ErrInvalidToken
    	}
    	var header tokenHeader
    	if err := json.Unmarshal(headerJSON, &header); err != nil || header.Alg != "HS256" {
    		return uuid.Nil(), ErrInvalidToken
    	}

    	got, err := b64.DecodeString(parts[2])
    	if err != nil {
    		return uuid.Nil(), ErrInvalidToken
    	}
    	if !hmac.Equal(got, sign(secret, parts[0]+"."+parts[1])) {
    		return uuid.Nil(), ErrInvalidToken
    	}

    	payload, err := b64.DecodeString(parts[1])
    	if err != nil {
    		return uuid.Nil(), ErrInvalidToken
    	}
    	var c claims
    	if err := json.Unmarshal(payload, &c); err != nil || c.Issuer != issuer {
    		return uuid.Nil(), ErrInvalidToken
    	}
    	if !now.Before(time.Unix(c.ExpiresAt, 0)) {
    		return uuid.Nil(), ErrExpiredToken
    	}
    	id, err := uuid.Parse(c.Subject)
    	if err != nil {
    		return uuid.Nil(), ErrInvalidToken
    	}
    	return id, nil
    }

    func main() {
    	secret := []byte("squeak-dev-secret-change-me")
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	issued := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)

    	token, _ := makeToken(pip, secret, issued, time.Hour)
    	fmt.Println("token:", token)

    	id, err := validateToken(token, secret, issued.Add(30*time.Minute))
    	fmt.Println("after 30m:", id, err)

    	_, err = validateToken(token, secret, issued.Add(2*time.Hour))
    	fmt.Println("after 2h:", err)

    	_, err = validateToken(token, []byte("wrong-secret"), issued)
    	fmt.Println("wrong secret:", err)
    }
  tests: |
    package main

    import (
    	"encoding/json/v2"
    	"errors"
    	"strings"
    	"testing"
    	"time"
    	"uuid"
    )

    var (
    	testSecret = []byte("test-secret-for-squeak")
    	testUser   = uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	testNow    = time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
    )

    func mustToken(t *testing.T) string {
    	t.Helper()
    	tok, err := makeToken(testUser, testSecret, testNow, time.Hour)
    	if err != nil {
    		t.Fatal(err)
    	}
    	return tok
    }

    func encodeJSON(v any) string {
    	data, _ := json.Marshal(v)
    	return b64.EncodeToString(data)
    }

    func TestValidToken(t *testing.T) {
    	tok := mustToken(t)
    	for _, after := range []time.Duration{0, 30 * time.Minute, 59 * time.Minute} {
    		id, err := validateToken(tok, testSecret, testNow.Add(after))
    		if err != nil {
    			t.Fatalf("validateToken(valid token, %v after issue) error = %v, want nil", after, err)
    		}
    		if id != testUser {
    			t.Errorf("validateToken returned user %s, want %s", id, testUser)
    		}
    	}
    }

    func TestExpiredToken(t *testing.T) {
    	tok := mustToken(t)
    	for _, after := range []time.Duration{time.Hour, time.Hour + time.Second, 48 * time.Hour} {
    		_, err := validateToken(tok, testSecret, testNow.Add(after))
    		if !errors.Is(err, ErrExpiredToken) {
    			t.Errorf("validateToken(%v after issue, ttl 1h) error = %v, want ErrExpiredToken", after, err)
    		}
    	}
    }

    func TestInvalidTokens(t *testing.T) {
    	tok := mustToken(t)
    	parts := strings.Split(tok, ".")
    	mallory := uuid.MustParse("0192f1e2-0000-7000-8000-000000000bad")

    	forgedPayload := encodeJSON(claims{Issuer: issuer, Subject: mallory.String(), IssuedAt: testNow.Unix(), ExpiresAt: testNow.Add(time.Hour).Unix()})
    	tampered := parts[0] + "." + forgedPayload + "." + parts[2]

    	attackerToken, _ := makeToken(mallory, []byte("attacker-secret"), testNow, time.Hour)

    	noneHeader := encodeJSON(tokenHeader{Alg: "none", Typ: "JWT"})
    	algNone := noneHeader + "." + parts[1] + "."

    	noneSigned := noneHeader + "." + parts[1] + "." + b64.EncodeToString(sign(testSecret, noneHeader+"."+parts[1]))

    	wrongIssuerPayload := encodeJSON(claims{Issuer: "cheddar-corp", Subject: testUser.String(), IssuedAt: testNow.Unix(), ExpiresAt: testNow.Add(time.Hour).Unix()})
    	wrongIssuer := parts[0] + "." + wrongIssuerPayload + "." + b64.EncodeToString(sign(testSecret, parts[0]+"."+wrongIssuerPayload))

    	badSubjectPayload := encodeJSON(claims{Issuer: issuer, Subject: "pip", IssuedAt: testNow.Unix(), ExpiresAt: testNow.Add(time.Hour).Unix()})
    	badSubject := parts[0] + "." + badSubjectPayload + "." + b64.EncodeToString(sign(testSecret, parts[0]+"."+badSubjectPayload))

    	expiredForgery := parts[0] + "." + encodeJSON(claims{Issuer: issuer, Subject: mallory.String(), ExpiresAt: 1}) + "." + parts[2]

    	for _, tt := range []struct {
    		name, token string
    	}{
    		{"empty string", ""},
    		{"two parts", parts[0] + "." + parts[1]},
    		{"four parts", tok + ".extra"},
    		{"signature not base64", parts[0] + "." + parts[1] + ".!!!"},
    		{"header not base64", "%%%." + parts[1] + "." + parts[2]},
    		{"tampered payload", tampered},
    		{"signed with another secret", attackerToken},
    		{"alg none, no signature", algNone},
    		{"alg none header", noneSigned},
    		{"wrong issuer", wrongIssuer},
    		{"subject not a UUID", badSubject},
    		{"forged token claiming to be expired", expiredForgery},
    	} {
    		id, err := validateToken(tt.token, testSecret, testNow)
    		if !errors.Is(err, ErrInvalidToken) {
    			t.Errorf("%s: validateToken error = %v, want ErrInvalidToken", tt.name, err)
    		}
    		if id != uuid.Nil() {
    			t.Errorf("%s: validateToken returned user %s, want uuid.Nil()", tt.name, id)
    		}
    	}
    }
---

Time to build Squeak's access tokens: real, standard HS256 JWTs, using nothing but
`crypto/hmac`, `crypto/sha256`, `encoding/base64` and `encoding/json/v2`.

`makeToken` is written for you. Read it first. It:

1. encodes the header `{"alg":"HS256","typ":"JWT"}` and the claims as JSON,
2. base64url-encodes each (`base64.RawURLEncoding`, no `=` padding),
3. joins them with a dot and signs *that string* with HMAC-SHA256,
4. appends the encoded signature after another dot.

Both functions take `now` as a parameter rather than calling `time.Now()`, which makes
them easy to test at any moment in time. The handler passes `time.Now()`.

## Your task

Write `validateToken(token, secret, now)`. It returns the user ID from `sub`, or an error:

1. Split the token on `.`. Anything other than exactly 3 parts is `ErrInvalidToken`.
2. Decode and parse the header. If it isn't valid, or `alg` isn't exactly `HS256`,
   return `ErrInvalidToken`. **Never** accept `none`.
3. Recompute `sign(secret, parts[0] + "." + parts[1])`, decode the token's signature,
   and compare them with **`hmac.Equal`**. A mismatch is `ErrInvalidToken`.
4. Only now decode and parse the claims. Bad JSON, or `iss` not equal to `issuer`, is
   `ErrInvalidToken`.
5. If `now` is **not before** `exp` (as `time.Unix(exp, 0)`), return `ErrExpiredToken`.
6. Parse `sub` with `uuid.Parse`. Failure is `ErrInvalidToken`.

Return `uuid.Nil()` alongside every error.

The tests throw a dozen hostile tokens at you: tampered payloads, other secrets,
`alg: none`, wrong issuers, garbage. Every one must come back as `ErrInvalidToken`.

## Hints

- `b64.DecodeString` returns an error for invalid base64. Treat any decode error as
  `ErrInvalidToken`.
- Comparing signatures with `==` or `bytes.Equal` *works*, but leaks timing
  information. `hmac.Equal` is the right tool.
- Delete the starter's `_, _ = strings.Split, hmac.Equal` line once you use both.
