---
title: Webhooks and API Keys
quiz:
  - question: Gouda Pay sends Squeak a webhook for an event type Squeak doesn't care about (`"invoice.created"`). What should Squeak respond?
    options:
      - text: '`400 Bad Request`, because the event is unknown'
      - text: '`2xx` (such as 204), acknowledging receipt and doing nothing'
        correct: true
      - text: '`404 Not Found`'
      - text: Nothing; close the connection
    explanation: |
      Any non-2xx answer tells the sender "delivery failed, try again". It will keep
      retrying an event you'll never handle. Acknowledge events you ignore with a 2xx.
  - question: Why is an HMAC signature over the request body stronger than a static API key in a header?
    options:
      - text: HMACs are longer than API keys
      - text: The signature proves the *body* came from the sender unmodified, and the secret itself never travels over the wire
        correct: true
      - text: API keys can't be compared in constant time
      - text: HMAC signatures never expire
    explanation: |
      A static key authenticates the caller but says nothing about the body, and it's sent
      with every request, so anything that logs headers can leak it. An HMAC ties the
      signature to the exact bytes, and the shared secret stays on both servers.
---

So far every request to Squeak has come from a mouse with an app. Now a *company* wants to
talk to you.

Squeak is launching **Squeak Gold**: a paid tier with a shiny badge. Payments are
handled by a (fictional) provider called **Gouda Pay**. When a user pays, Gouda Pay needs
to tell Squeak "user X just upgraded". It does that with a **webhook**.

## What's a webhook?

A webhook is an HTTP request that *another service* sends to *your* server when
something happens on their side. Instead of Squeak polling Gouda Pay every minute ("any
new payments? how about now?"), Gouda Pay calls you:

```
POST /api/gouda/webhooks
Authorization: ApiKey f271c81ff7084ee5b99a5091b42d486e
Content-Type: application/json

{"id": "evt_8Ybf2", "event": "user.upgraded", "data": {"user_id": "0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6"}}
```

You register the URL once in the provider's dashboard. From then on, your server is the
one receiving requests from a machine, which brings three questions:

1. How do you know it's really Gouda Pay? (Anyone can POST to a URL.)
2. What should you answer?
3. What happens when it sends the same event twice? (Next lesson.)

## API keys

The simplest answer to question 1: Gouda Pay gives you a secret **API key**, and sends it
in a header with every webhook. Your handler compares it with the key from your config:

```go
func hasAPIKey(h http.Header, want string) bool {
	scheme, key, ok := strings.Cut(h.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "ApiKey") {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(key), []byte(want)) == 1
}
```

It's the same shape as the bearer token parser, with a different scheme. Compare with
`subtle.ConstantTimeCompare` rather than `==`, so response timing can't be used to guess
the key byte by byte. (It returns early only when the *lengths* differ, and the length of
a key isn't much of a secret.)

API keys are also how *Squeak* could let other programs call it: generate one with
`crypto/rand.Text()`, show it to the user once, and store only its SHA-256 hash, exactly
like refresh tokens.

## A webhook handler

```go
package main

import (
	"crypto/subtle"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

type goudaEvent struct {
	ID    string `json:"id"`
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

type apiConfig struct {
	goudaKey string
	mu       sync.Mutex
	gold     map[string]bool // user ID -> has Squeak Gold
}

func hasAPIKey(h http.Header, want string) bool {
	scheme, key, ok := strings.Cut(h.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "ApiKey") {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(key), []byte(want)) == 1
}

func (cfg *apiConfig) handleGoudaWebhook(w http.ResponseWriter, r *http.Request) {
	if !hasAPIKey(r.Header, cfg.goudaKey) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var ev goudaEvent
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := json.UnmarshalRead(r.Body, &ev); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if ev.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent) // not interested, but got it
		return
	}
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if _, ok := cfg.gold[ev.Data.UserID]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	cfg.gold[ev.Data.UserID] = true
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	cfg := &apiConfig{goudaKey: "f271c81ff7084ee5", gold: map[string]bool{"pip": false}}

	send := func(auth, body string) int {
		req := httptest.NewRequest("POST", "/api/gouda/webhooks", strings.NewReader(body))
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		rec := httptest.NewRecorder()
		cfg.handleGoudaWebhook(rec, req)
		return rec.Code
	}
	upgrade := `{"id":"evt_1","event":"user.upgraded","data":{"user_id":"pip"}}`

	fmt.Println("no key:      ", send("", upgrade))
	fmt.Println("wrong key:   ", send("ApiKey guess", upgrade))
	fmt.Println("other event: ", send("ApiKey f271c81ff7084ee5", `{"id":"evt_2","event":"invoice.created"}`))
	fmt.Println("unknown user:", send("ApiKey f271c81ff7084ee5", `{"id":"evt_3","event":"user.upgraded","data":{"user_id":"nobody"}}`))
	fmt.Println("upgrade:     ", send("ApiKey f271c81ff7084ee5", upgrade))
	fmt.Println("pip is gold: ", cfg.gold["pip"])
}
```

```
no key:       401
wrong key:    401
other event:  204
unknown user: 404
upgrade:      204
pip is gold:  true
```

## Answering webhooks

The sender only cares about one thing: **did you get it?**

- **2xx** means "received, stop sending". Use it for events you handled *and* for events
  you deliberately ignore. Otherwise the sender retries them forever.
- **Anything else** (4xx, 5xx, a timeout) means "failed". Most providers retry with
  exponential backoff for hours or days.
- **Answer fast.** Providers time out after a few seconds. If handling an event is slow
  (sending emails, generating PDFs), record it and do the work in the background.

Notice that the unknown user got a 404. That's a judgement call: it means Gouda Pay will
retry, which helps if the user record just hasn't been written *yet*. If retrying could
never help, a 2xx and a log line is kinder.

## Signed webhooks

A static API key has weaknesses. It travels with every request, and it proves who sent
the request but not that the *body* wasn't changed on the way. Big providers (Stripe,
GitHub, Slack) therefore **sign** each webhook: they compute an HMAC-SHA256 of the raw
body with a shared secret and send it in a header, such as
`X-Gouda-Signature: sha256=5d1f...`. You verify it the same way you verified JWTs:

```go
body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64<<10))
// ... handle err ...
mac := hmac.New(sha256.New, cfg.goudaSecret)
mac.Write(body)
want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
if !hmac.Equal([]byte(r.Header.Get("X-Gouda-Signature")), []byte(want)) {
	w.WriteHeader(http.StatusUnauthorized)
	return
}
// only now: json.Unmarshal(body, &ev)
```

Two gotchas. You must sign the **exact raw bytes**, so read the body into a `[]byte`
first and decode from that (re-encoding the JSON would change whitespace and break the
signature). And signatures usually include a **timestamp** in the signed data, so an
attacker who records a valid webhook can't replay it next week.
