---
title: Idempotency
quiz:
  - question: Which of these operations is naturally idempotent?
    options:
      - text: '`POST /api/squeaks` with `{"body":"hi"}`'
      - text: '"Add one month of Squeak Gold to user X"'
      - text: '"Set user X''s plan to Gold"'
        correct: true
      - text: '"Increment the hit counter"'
    explanation: |
      Setting something to a value gives the same end state however many times you do it.
      Creating a new squeak, adding a month or incrementing a counter changes the state
      again on every repeat.
  - question: |
      Squeak's webhook handler records the event ID as processed, then crashes before
      crediting the user. Gouda Pay retries. What happens?
    options:
      - text: The retry is processed normally
      - text: The retry is skipped as a duplicate, so the user never gets their Gold
        correct: true
      - text: The user gets Gold twice
      - text: Gouda Pay stops sending webhooks
    explanation: |
      Marking "done" before the work is done loses the work if you crash in between.
      Mark the event processed in the *same* transaction (or under the same lock) as the
      change itself, or after it succeeds.
exercise:
  starter: |
    package main

    import (
    	"crypto/subtle"
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"uuid"
    )

    type goudaEvent struct {
    	ID    string `json:"id"`
    	Event string `json:"event"`
    	Data  struct {
    		UserID uuid.UUID `json:"user_id"`
    	} `json:"data"`
    }

    type apiConfig struct {
    	goudaKey string

    	mu         sync.Mutex
    	goldMonths map[uuid.UUID]int // user ID -> months of Squeak Gold
    	processed  map[string]bool   // Gouda Pay event IDs already handled
    }

    func newAPIConfig(goudaKey string) *apiConfig {
    	return &apiConfig{
    		goudaKey:   goudaKey,
    		goldMonths: make(map[uuid.UUID]int),
    		processed:  make(map[string]bool),
    	}
    }

    // hasAPIKey reports whether h carries "Authorization: ApiKey <want>".
    func hasAPIKey(h http.Header, want string) bool {
    	scheme, key, ok := strings.Cut(h.Get("Authorization"), " ")
    	if !ok || !strings.EqualFold(scheme, "ApiKey") {
    		return false
    	}
    	return subtle.ConstantTimeCompare([]byte(key), []byte(want)) == 1
    }

    // handleGoudaWebhook handles POST /api/gouda/webhooks. See the lesson for the rules.
    func (cfg *apiConfig) handleGoudaWebhook(w http.ResponseWriter, r *http.Request) {
    	var ev goudaEvent
    	json.UnmarshalRead(r.Body, &ev)
    	// ?
    	cfg.goldMonths[ev.Data.UserID]++
    	w.WriteHeader(http.StatusNoContent)
    }

    func main() {
    	cfg := newAPIConfig("gouda-secret")
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	send := func(auth, body string) int {
    		req := httptest.NewRequest("POST", "/api/gouda/webhooks", strings.NewReader(body))
    		if auth != "" {
    			req.Header.Set("Authorization", auth)
    		}
    		rec := httptest.NewRecorder()
    		cfg.handleGoudaWebhook(rec, req)
    		return rec.Code
    	}
    	evt1 := `{"id":"evt_1","event":"user.upgraded","data":{"user_id":"` + pip.String() + `"}}`
    	evt2 := `{"id":"evt_2","event":"user.upgraded","data":{"user_id":"` + pip.String() + `"}}`

    	fmt.Println("no key:        ", send("", evt1))
    	fmt.Println("evt_1:         ", send("ApiKey gouda-secret", evt1))
    	fmt.Println("evt_1 again:   ", send("ApiKey gouda-secret", evt1))
    	fmt.Println("evt_2:         ", send("ApiKey gouda-secret", evt2))
    	fmt.Println("other event:   ", send("ApiKey gouda-secret", `{"id":"evt_3","event":"invoice.created"}`))
    	fmt.Println("pip's months:  ", cfg.goldMonths[pip])
    }
  solution: |
    package main

    import (
    	"crypto/subtle"
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"uuid"
    )

    type goudaEvent struct {
    	ID    string `json:"id"`
    	Event string `json:"event"`
    	Data  struct {
    		UserID uuid.UUID `json:"user_id"`
    	} `json:"data"`
    }

    type apiConfig struct {
    	goudaKey string

    	mu         sync.Mutex
    	goldMonths map[uuid.UUID]int // user ID -> months of Squeak Gold
    	processed  map[string]bool   // Gouda Pay event IDs already handled
    }

    func newAPIConfig(goudaKey string) *apiConfig {
    	return &apiConfig{
    		goudaKey:   goudaKey,
    		goldMonths: make(map[uuid.UUID]int),
    		processed:  make(map[string]bool),
    	}
    }

    // hasAPIKey reports whether h carries "Authorization: ApiKey <want>".
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
    	if err := json.UnmarshalRead(r.Body, &ev); err != nil || ev.ID == "" {
    		w.WriteHeader(http.StatusBadRequest)
    		return
    	}
    	if ev.Event != "user.upgraded" {
    		w.WriteHeader(http.StatusNoContent)
    		return
    	}

    	cfg.mu.Lock()
    	if !cfg.processed[ev.ID] {
    		cfg.goldMonths[ev.Data.UserID]++
    		cfg.processed[ev.ID] = true
    	}
    	cfg.mu.Unlock()
    	w.WriteHeader(http.StatusNoContent)
    }

    func main() {
    	cfg := newAPIConfig("gouda-secret")
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	send := func(auth, body string) int {
    		req := httptest.NewRequest("POST", "/api/gouda/webhooks", strings.NewReader(body))
    		if auth != "" {
    			req.Header.Set("Authorization", auth)
    		}
    		rec := httptest.NewRecorder()
    		cfg.handleGoudaWebhook(rec, req)
    		return rec.Code
    	}
    	evt1 := `{"id":"evt_1","event":"user.upgraded","data":{"user_id":"` + pip.String() + `"}}`
    	evt2 := `{"id":"evt_2","event":"user.upgraded","data":{"user_id":"` + pip.String() + `"}}`

    	fmt.Println("no key:        ", send("", evt1))
    	fmt.Println("evt_1:         ", send("ApiKey gouda-secret", evt1))
    	fmt.Println("evt_1 again:   ", send("ApiKey gouda-secret", evt1))
    	fmt.Println("evt_2:         ", send("ApiKey gouda-secret", evt2))
    	fmt.Println("other event:   ", send("ApiKey gouda-secret", `{"id":"evt_3","event":"invoice.created"}`))
    	fmt.Println("pip's months:  ", cfg.goldMonths[pip])
    }
  tests: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"testing"
    	"uuid"
    )

    const testKey = "test-gouda-key"

    var (
    	pip      = uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	whiskers = uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-00000000beef")
    )

    func upgrade(eventID string, user uuid.UUID) string {
    	return fmt.Sprintf(`{"id":%q,"event":"user.upgraded","data":{"user_id":%q}}`, eventID, user)
    }

    func send(cfg *apiConfig, auth, body string) int {
    	req := httptest.NewRequest("POST", "/api/gouda/webhooks", strings.NewReader(body))
    	if auth != "" {
    		req.Header.Set("Authorization", auth)
    	}
    	rec := httptest.NewRecorder()
    	cfg.handleGoudaWebhook(rec, req)
    	return rec.Code
    }

    func months(cfg *apiConfig, user uuid.UUID) int {
    	cfg.mu.Lock()
    	defer cfg.mu.Unlock()
    	return cfg.goldMonths[user]
    }

    func TestAPIKey(t *testing.T) {
    	cfg := newAPIConfig(testKey)
    	for _, auth := range []string{"", "ApiKey wrong", "Bearer " + testKey, testKey, "ApiKey " + testKey + "x"} {
    		if code := send(cfg, auth, upgrade("evt_1", pip)); code != http.StatusUnauthorized {
    			t.Errorf("Authorization %q: status = %d, want 401", auth, code)
    		}
    	}
    	if got := months(cfg, pip); got != 0 {
    		t.Errorf("rejected webhooks changed pip's Gold months to %d, want 0", got)
    	}
    	if code := send(cfg, "apikey "+testKey, upgrade("evt_1", pip)); code != http.StatusNoContent {
    		t.Errorf("valid key with a lowercase scheme: status = %d, want 204", code)
    	}
    }

    func TestBadBodies(t *testing.T) {
    	cfg := newAPIConfig(testKey)
    	for _, body := range []string{
    		`{"id":"evt_1","event":`,
    		`not json`,
    		`{"event":"user.upgraded","data":{"user_id":"` + pip.String() + `"}}`,
    		`{"id":"evt_2","event":"user.upgraded","data":{"user_id":"` + pip.String() + `"},"pad":"` + strings.Repeat("x", 70<<10) + `"}`,
    	} {
    		if code := send(cfg, "ApiKey "+testKey, body); code != http.StatusBadRequest {
    			t.Errorf("POST %.50s...: status = %d, want 400", body, code)
    		}
    	}
    	if got := months(cfg, pip); got != 0 {
    		t.Errorf("bad bodies changed pip's Gold months to %d, want 0", got)
    	}
    }

    func TestIgnoredEvent(t *testing.T) {
    	cfg := newAPIConfig(testKey)
    	body := `{"id":"evt_9","event":"invoice.created","data":{"user_id":"` + pip.String() + `"}}`
    	if code := send(cfg, "ApiKey "+testKey, body); code != http.StatusNoContent {
    		t.Errorf("unhandled event type: status = %d, want 204 (acknowledge it so Gouda Pay stops retrying)", code)
    	}
    	if got := months(cfg, pip); got != 0 {
    		t.Errorf("an invoice.created event gave pip %d months of Gold, want 0", got)
    	}
    }

    func TestDuplicates(t *testing.T) {
    	cfg := newAPIConfig(testKey)
    	for _, tt := range []struct {
    		event string
    		user  uuid.UUID
    	}{{"evt_1", pip}, {"evt_1", pip}, {"evt_2", pip}, {"evt_3", whiskers}, {"evt_2", pip}, {"evt_1", pip}} {
    		if code := send(cfg, "ApiKey "+testKey, upgrade(tt.event, tt.user)); code != http.StatusNoContent {
    			t.Errorf("delivery of %s: status = %d, want 204 (duplicates are acknowledged too)", tt.event, code)
    		}
    	}
    	if got := months(cfg, pip); got != 2 {
    		t.Errorf("pip got %d months from events evt_1 and evt_2 (each retried), want 2", got)
    	}
    	if got := months(cfg, whiskers); got != 1 {
    		t.Errorf("whiskers got %d months, want 1", got)
    	}
    }

    func TestConcurrentRetries(t *testing.T) {
    	cfg := newAPIConfig(testKey)
    	var wg sync.WaitGroup
    	for i := range 100 {
    		wg.Go(func() {
    			send(cfg, "ApiKey "+testKey, upgrade(fmt.Sprintf("evt_%d", i%5), pip))
    		})
    	}
    	wg.Wait()
    	if got := months(cfg, pip); got != 5 {
    		t.Errorf("100 simultaneous deliveries of 5 distinct events gave pip %d months, want 5", got)
    	}
    }
---

Networks are unreliable. A webhook from Gouda Pay reaches Squeak, Squeak upgrades the
user, and the `204` response gets lost on the way back. From Gouda Pay's point of view the
delivery failed, so it sends the event **again**. Every webhook provider documents it:
*delivery is at least once*. You will see duplicates.

The same happens with your own clients. A mouse on a train taps "Squeak!", the app times
out waiting for the response and retries. Did the first request make it? The app can't
know.

An operation is **idempotent** if doing it twice has the same effect as doing it once.
The whole trick of handling retries is making your handlers idempotent.

## HTTP methods and idempotency

The HTTP spec already classifies methods:

| Method | Idempotent? | Why |
|---|---|---|
| `GET`, `HEAD` | yes | they don't change anything |
| `PUT` | yes | "set this resource to exactly this" |
| `DELETE` | yes | deleting twice leaves it deleted |
| `POST` | **no** | "create a new thing" or "do this action" |

(A second `DELETE` may answer 404 instead of 204. The *response* differs, but the *state*
is the same, and state is what idempotency is about.)

Clients and proxies may retry idempotent requests on their own. `POST` is the one that
needs care.

## Deduplicating webhooks by event ID

Well-behaved webhook senders give every event a unique ID (`"id": "evt_8Ybf2"`) and
reuse it on retries. So remember which IDs you've processed, and skip repeats. Here's
why it matters for an operation that isn't naturally idempotent, "add one month of Gold":

```go
package main

import (
	"fmt"
	"sync"
)

type billing struct {
	mu        sync.Mutex
	months    map[string]int      // user -> months of Gold
	processed map[string]struct{} // event IDs already handled
}

// naiveCredit adds a month every time it's called.
func (b *billing) naiveCredit(eventID, user string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.months[user]++
}

// credit adds a month once per event ID. The check, the change and the
// "processed" mark all happen under one lock, so they're a single step.
func (b *billing) credit(eventID, user string) (duplicate bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, seen := b.processed[eventID]; seen {
		return true
	}
	b.months[user]++
	b.processed[eventID] = struct{}{}
	return false
}

func main() {
	deliveries := []string{"evt_1", "evt_1", "evt_2", "evt_1"} // evt_1 retried twice

	naive := &billing{months: map[string]int{}}
	for _, id := range deliveries {
		naive.naiveCredit(id, "pip")
	}
	fmt.Println("naive: pip has", naive.months["pip"], "months of Gold")

	safe := &billing{months: map[string]int{}, processed: map[string]struct{}{}}
	for _, id := range deliveries {
		if safe.credit(id, "pip") {
			fmt.Println("skipped duplicate", id)
		}
	}
	fmt.Println("safe:  pip has", safe.months["pip"], "months of Gold")
}
```

```
naive: pip has 4 months of Gold
skipped duplicate evt_1
skipped duplicate evt_1
safe:  pip has 2 months of Gold
```

Things to get right:

- **Check, act and record atomically.** Under one lock in memory, and in one transaction
  in a database (a unique index on `event_id` makes the database enforce it for you).
  Recording "processed" *before* the work succeeds loses events when something crashes
  in between.
- **Answer 2xx for duplicates.** From the sender's side, the delivery succeeded.
- **Expire old IDs eventually.** Providers stop retrying after a few days, so you
  don't need to remember event IDs forever.

Even better is to design the operation itself to be idempotent where you can. "Set
Pip's Gold to expire on 2026-10-01" (from the event's data) is naturally safe to repeat,
whereas "add a month" isn't.

## Idempotency keys for your own API

For `POST` requests from your clients, the common pattern (popularised by Stripe) is an
**`Idempotency-Key`** header. The client generates a random key for each *logical*
action and sends the same key on every retry of it:

```
POST /api/squeaks
Authorization: Bearer eyJ...
Idempotency-Key: 5f0c3e0a-6a2b-4b8e-9c0d-2f4a1e7b9c11

{"body": "on the train, hope this sends"}
```

The server stores `(user, key) -> response` for a day or so. The first request runs
normally and saves its status and body. A retry with the same key gets the *saved*
response without creating a second squeak. Scope keys to the user, so one user can't
replay another's response, and if a retry arrives while the first attempt is still
running, answer `409 Conflict` rather than running it twice.

## Idempotency and authorization together

Webhook endpoints are a place where authorization, idempotency and speed all meet.
A solid Gouda Pay handler:

1. verifies the API key or signature (**401** if wrong),
2. parses the event, and answers **2xx** for types it doesn't handle,
3. applies the change and records the event ID in one atomic step,
4. answers **204** quickly, for new events and duplicates alike.

## Your task

Build that handler. Gouda Pay's `user.upgraded` event now means "add one month of
Squeak Gold", which is *not* naturally idempotent, so deduplicate by event ID.

`handleGoudaWebhook` must:

1. Respond **401** unless `hasAPIKey(r.Header, cfg.goudaKey)` is true.
2. Limit the body to 64 KiB with `http.MaxBytesReader`, and decode it with
   `json.UnmarshalRead`. A decode error, or an event with an empty `id`: **400**.
3. For any `event` other than `user.upgraded`: **204** and change nothing.
4. Otherwise, under `cfg.mu`, skip events whose ID is already in `cfg.processed`. For a
   new one, add 1 to `cfg.goldMonths[userID]` and record the ID.
5. Answer **204**, for new events and duplicates alike.

The tests replay events, send 100 deliveries at once from many goroutines, and expect
exactly one month per distinct event ID.

That completes Squeak's feature set. Next up: proving it all works with tests.
