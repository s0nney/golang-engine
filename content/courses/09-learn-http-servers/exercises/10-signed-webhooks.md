---
title: Signed Webhooks
difficulty: hard
after: authorization-and-webhooks
hints:
  - 'Order matters. Read the raw body first (with `http.MaxBytesReader`), parse the header, check the signature over `t + "." + body`, then the timestamp, and only then `json.Unmarshal` the body you already read. Never trust the JSON before the signature checks out.'
  - 'For each `v1` value and each secret: `mac := hmac.New(sha256.New, secret)`, write the signed payload, and compare `hex.EncodeToString(mac.Sum(nil))` with the value using `hmac.Equal([]byte(a), []byte(b))`. Any match is enough. Time is `time.Unix(t, 0)`; reject it if it''s more than `tolerance` before *or after* `r.now()`.'
  - 'Keep `seen map[string]time.Time` (event ID to when it was processed) behind a mutex, and hold the lock across check, `handle` and mark, so two copies of an event can''t both run. Only mark an event after `handle` succeeds, and treat an entry older than `dedupFor` as not seen.'
exercise:
  starter: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"time"
    )

    type Event struct {
    	ID   string            `json:"id"`
    	Type string            `json:"type"`
    	Data map[string]string `json:"data"`
    }

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, _ := json.Marshal(payload)
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, map[string]string{"error": msg})
    }

    const (
    	maxWebhookBody = 64 << 10
    	tolerance      = 5 * time.Minute
    	dedupFor       = 72 * time.Hour
    )

    type WebhookReceiver struct {
    }

    func NewWebhookReceiver(secrets [][]byte, now func() time.Time, handle func(Event) error) *WebhookReceiver {
    	return &WebhookReceiver{}
    }

    func (rc *WebhookReceiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	respondWithError(w, http.StatusNotImplemented, "not implemented")
    }

    func main() {
    	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    	rc := NewWebhookReceiver([][]byte{[]byte("whsec_brie")}, func() time.Time { return now }, func(ev Event) error {
    		fmt.Println("handling", ev.Type, "for", ev.Data["user_id"])
    		return nil
    	})
    	body := `{"id":"evt_1","type":"user.upgraded","data":{"user_id":"pip"}}`
    	req := httptest.NewRequest("POST", "/api/gouda/webhooks", strings.NewReader(body))
    	// HMAC-SHA256("whsec_brie", "1788264000." + body), in hex:
    	req.Header.Set("Gouda-Signature", "t=1788264000,v1=a6efcf60d4cf7676539c12124b597d4b9f6e163d409a4447e9da1362d3c1654e")
    	rec := httptest.NewRecorder()
    	rc.ServeHTTP(rec, req)
    	fmt.Println(rec.Code, rec.Body.String())
    }
  solution: |
    package main

    import (
    	"crypto/hmac"
    	"crypto/sha256"
    	"encoding/hex"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
    	"strings"
    	"sync"
    	"time"
    )

    type Event struct {
    	ID   string            `json:"id"`
    	Type string            `json:"type"`
    	Data map[string]string `json:"data"`
    }

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, _ := json.Marshal(payload)
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, map[string]string{"error": msg})
    }

    const (
    	maxWebhookBody = 64 << 10
    	tolerance      = 5 * time.Minute
    	dedupFor       = 72 * time.Hour
    )

    type WebhookReceiver struct {
    	secrets [][]byte
    	now     func() time.Time
    	handle  func(Event) error

    	mu   sync.Mutex
    	seen map[string]time.Time // event ID -> when it was processed
    }

    func NewWebhookReceiver(secrets [][]byte, now func() time.Time, handle func(Event) error) *WebhookReceiver {
    	return &WebhookReceiver{secrets: secrets, now: now, handle: handle, seen: make(map[string]time.Time)}
    }

    var errBadHeader = errors.New("malformed Gouda-Signature header")

    // parseSignatureHeader parses "t=<unix seconds>,v1=<hex>[,v1=<hex>...]".
    func parseSignatureHeader(h string) (int64, []string, error) {
    	var (
    		ts     int64
    		haveTS bool
    		sigs   []string
    	)
    	for item := range strings.SplitSeq(h, ",") {
    		k, v, ok := strings.Cut(strings.TrimSpace(item), "=")
    		if !ok {
    			return 0, nil, errBadHeader
    		}
    		switch k {
    		case "t":
    			n, err := strconv.ParseInt(v, 10, 64)
    			if err != nil || haveTS {
    				return 0, nil, errBadHeader
    			}
    			ts, haveTS = n, true
    		case "v1":
    			sigs = append(sigs, v)
    		}
    	}
    	if !haveTS || len(sigs) == 0 {
    		return 0, nil, errBadHeader
    	}
    	return ts, sigs, nil
    }

    func (rc *WebhookReceiver) validSignature(ts int64, body []byte, sigs []string) bool {
    	for _, secret := range rc.secrets {
    		mac := hmac.New(sha256.New, secret)
    		fmt.Fprintf(mac, "%d.", ts)
    		mac.Write(body)
    		want := []byte(hex.EncodeToString(mac.Sum(nil)))
    		for _, sig := range sigs {
    			if hmac.Equal([]byte(sig), want) {
    				return true
    			}
    		}
    	}
    	return false
    }

    func (rc *WebhookReceiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBody))
    	if err != nil {
    		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
    			respondWithError(w, http.StatusRequestEntityTooLarge, "payload too large")
    			return
    		}
    		respondWithError(w, http.StatusBadRequest, "couldn't read payload")
    		return
    	}
    	ts, sigs, err := parseSignatureHeader(r.Header.Get("Gouda-Signature"))
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, "missing or malformed Gouda-Signature header")
    		return
    	}
    	if !rc.validSignature(ts, body, sigs) {
    		respondWithError(w, http.StatusUnauthorized, "invalid signature")
    		return
    	}
    	age := rc.now().Sub(time.Unix(ts, 0))
    	if age > tolerance || age < -tolerance {
    		respondWithError(w, http.StatusUnauthorized, "timestamp outside the tolerance")
    		return
    	}
    	var ev Event
    	if err := json.Unmarshal(body, &ev); err != nil || ev.ID == "" {
    		respondWithError(w, http.StatusBadRequest, "invalid event")
    		return
    	}

    	rc.mu.Lock()
    	defer rc.mu.Unlock()
    	now := rc.now()
    	for id, at := range rc.seen {
    		if now.Sub(at) >= dedupFor {
    			delete(rc.seen, id)
    		}
    	}
    	if _, dup := rc.seen[ev.ID]; dup {
    		respondWithJSON(w, http.StatusOK, map[string]string{"status": "duplicate"})
    		return
    	}
    	if err := rc.handle(ev); err != nil {
    		respondWithError(w, http.StatusInternalServerError, "couldn't process event")
    		return
    	}
    	rc.seen[ev.ID] = now
    	respondWithJSON(w, http.StatusOK, map[string]string{"status": "processed"})
    }

    func sign(secret []byte, ts int64, body string) string {
    	mac := hmac.New(sha256.New, secret)
    	fmt.Fprintf(mac, "%d.%s", ts, body)
    	return hex.EncodeToString(mac.Sum(nil))
    }

    func main() {
    	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    	rc := NewWebhookReceiver([][]byte{[]byte("whsec_brie")}, func() time.Time { return now }, func(ev Event) error {
    		fmt.Println("handling", ev.Type, "for", ev.Data["user_id"])
    		return nil
    	})
    	body := `{"id":"evt_1","type":"user.upgraded","data":{"user_id":"pip"}}`
    	req := httptest.NewRequest("POST", "/api/gouda/webhooks", strings.NewReader(body))
    	req.Header.Set("Gouda-Signature", fmt.Sprintf("t=%d,v1=%s", now.Unix(), sign([]byte("whsec_brie"), now.Unix(), body)))
    	rec := httptest.NewRecorder()
    	rc.ServeHTTP(rec, req)
    	fmt.Println(rec.Code, rec.Body.String())
    }
  tests: |
    package main

    import (
    	"crypto/hmac"
    	"crypto/sha256"
    	"encoding/hex"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"net/http/httptest"
    	"os"
    	"strings"
    	"sync"
    	"testing"
    	"time"
    )

    var (
    	current = []byte("whsec_brie_2026")
    	old     = []byte("whsec_camembert_2025")
    	start   = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    )

    func hmacHex(secret []byte, ts int64, body string) string {
    	mac := hmac.New(sha256.New, secret)
    	fmt.Fprintf(mac, "%d.%s", ts, body)
    	return hex.EncodeToString(mac.Sum(nil))
    }

    type harness struct {
    	rc      *WebhookReceiver
    	now     time.Time
    	mu      sync.Mutex
    	handled []Event
    	fail    error
    }

    func newHarness() *harness {
    	h := &harness{now: start}
    	h.rc = NewWebhookReceiver([][]byte{current, old}, func() time.Time { return h.now }, func(ev Event) error {
    		h.mu.Lock()
    		defer h.mu.Unlock()
    		if h.fail != nil {
    			return h.fail
    		}
    		h.handled = append(h.handled, ev)
    		return nil
    	})
    	return h
    }

    func (h *harness) count() int {
    	h.mu.Lock()
    	defer h.mu.Unlock()
    	return len(h.handled)
    }

    func (h *harness) deliver(header, body string) (int, map[string]string, string) {
    	req := httptest.NewRequest("POST", "/api/gouda/webhooks", strings.NewReader(body))
    	if header != "" {
    		req.Header.Set("Gouda-Signature", header)
    	}
    	rec := httptest.NewRecorder()
    	h.rc.ServeHTTP(rec, req)
    	var m map[string]string
    	json.Unmarshal(rec.Body.Bytes(), &m)
    	return rec.Code, m, rec.Body.String()
    }

    func signed(ts time.Time, body string, secret []byte) string {
    	return fmt.Sprintf("t=%d,v1=%s", ts.Unix(), hmacHex(secret, ts.Unix(), body))
    }

    const upgrade = `{"id":"evt_1","type":"user.upgraded","data":{"user_id":"pip"}}`

    func TestValidWebhook(t *testing.T) {
    	h := newHarness()
    	code, m, raw := h.deliver(signed(start, upgrade, current), upgrade)
    	if code != 200 || m["status"] != "processed" {
    		t.Fatalf("correctly signed webhook: %d %s, want 200 {\"status\":\"processed\"}", code, raw)
    	}
    	if h.count() != 1 || h.handled[0].ID != "evt_1" || h.handled[0].Type != "user.upgraded" || h.handled[0].Data["user_id"] != "pip" {
    		t.Errorf("handle got %+v, want one event evt_1 user.upgraded for pip", h.handled)
    	}
    }

    func TestSecretRotationAndMultipleSignatures(t *testing.T) {
    	h := newHarness()
    	body := `{"id":"evt_old","type":"ping"}`
    	if code, _, raw := h.deliver(signed(start, body, old), body); code != 200 {
    		t.Errorf("signed with the previous secret: %d %s, want 200 (both secrets are valid during rotation)", code, raw)
    	}
    	body = `{"id":"evt_two","type":"ping"}`
    	header := fmt.Sprintf("t=%d,v1=%s,v1=%s", start.Unix(), strings.Repeat("0", 64), hmacHex(current, start.Unix(), body))
    	if code, _, raw := h.deliver(header, body); code != 200 {
    		t.Errorf("header with a bad v1 and a good v1: %d %s, want 200 (any matching v1 is enough)", code, raw)
    	}
    	body = `{"id":"evt_v0","type":"ping"}`
    	header = fmt.Sprintf("t=%d, v0=legacy, v1=%s", start.Unix(), hmacHex(current, start.Unix(), body))
    	if code, _, raw := h.deliver(header, body); code != 200 {
    		t.Errorf("header %q: %d %s, want 200 (ignore other schemes like v0, allow spaces after commas)", header, code, raw)
    	}
    }

    func TestBadSignatures(t *testing.T) {
    	ts := start.Unix()
    	tests := []struct {
    		name, header, body string
    	}{
    		{"tampered body", signed(start, upgrade, current), strings.Replace(upgrade, "pip", "whiskers", 1)},
    		{"unknown secret", signed(start, upgrade, []byte("whsec_guess")), upgrade},
    		{"signed without the timestamp", fmt.Sprintf("t=%d,v1=%x", ts, hmacSum(current, upgrade)), upgrade},
    		{"timestamp changed after signing", fmt.Sprintf("t=%d,v1=%s", ts+1, hmacHex(current, ts, upgrade)), upgrade},
    		{"signature is the secret", fmt.Sprintf("t=%d,v1=%s", ts, current), upgrade},
    		{"truncated signature", signed(start, upgrade, current)[:40], upgrade},
    		{"empty signature", fmt.Sprintf("t=%d,v1=", ts), upgrade},
    	}
    	for _, tt := range tests {
    		h := newHarness()
    		code, m, raw := h.deliver(tt.header, tt.body)
    		if code != 401 || m["error"] == "" {
    			t.Errorf("%s: %d %s, want 401 with a JSON error", tt.name, code, raw)
    		}
    		if h.count() != 0 {
    			t.Errorf("%s: handle was called for an unverified webhook", tt.name)
    		}
    		expected := hmacHex(current, ts, tt.body)
    		if strings.Contains(raw, expected) || strings.Contains(raw, string(current)) || strings.Contains(raw, string(old)) {
    			t.Errorf("%s: the response %s leaks the secret or the expected signature", tt.name, raw)
    		}
    	}
    }

    func hmacSum(secret []byte, body string) []byte {
    	mac := hmac.New(sha256.New, secret)
    	mac.Write([]byte(body))
    	return mac.Sum(nil)
    }

    func TestMalformedHeader(t *testing.T) {
    	sig := hmacHex(current, start.Unix(), upgrade)
    	for _, header := range []string{
    		"",
    		"garbage",
    		"v1=" + sig,
    		fmt.Sprintf("t=%d", start.Unix()),
    		"t=noon,v1=" + sig,
    		fmt.Sprintf("t=%d,t=%d,v1=%s", start.Unix(), start.Unix(), sig),
    	} {
    		h := newHarness()
    		if code, m, raw := h.deliver(header, upgrade); code != 400 || m["error"] == "" {
    			t.Errorf("Gouda-Signature %q: %d %s, want 400 with a JSON error", header, code, raw)
    		}
    	}
    }

    func TestTimestampTolerance(t *testing.T) {
    	for _, tt := range []struct {
    		offset time.Duration
    		want   int
    	}{
    		{0, 200}, {-5 * time.Minute, 200}, {5 * time.Minute, 200},
    		{-5*time.Minute - time.Second, 401}, {5*time.Minute + time.Second, 401},
    		{-24 * time.Hour, 401}, {24 * time.Hour, 401},
    	} {
    		h := newHarness()
    		ts := start.Add(tt.offset)
    		code, _, raw := h.deliver(signed(ts, upgrade, current), upgrade)
    		if code != tt.want {
    			t.Errorf("correctly signed, timestamp %v from now: %d %s, want %d (tolerance is 5 minutes either way)", tt.offset, code, raw, tt.want)
    		}
    	}
    }

    func TestReplayAndRetries(t *testing.T) {
    	h := newHarness()
    	header := signed(start, upgrade, current)
    	h.deliver(header, upgrade)
    	code, m, raw := h.deliver(header, upgrade)
    	if code != 200 || m["status"] != "duplicate" {
    		t.Errorf("exact replay of a processed webhook: %d %s, want 200 {\"status\":\"duplicate\"}", code, raw)
    	}
    	h.now = start.Add(time.Hour)
    	code, m, _ = h.deliver(signed(h.now, upgrade, current), upgrade)
    	if code != 200 || m["status"] != "duplicate" {
    		t.Errorf("Gouda retries evt_1 an hour later with a fresh signature: status %q, want duplicate", m["status"])
    	}
    	h.now = start.Add(dedupFor - time.Second)
    	if _, m, _ := h.deliver(signed(h.now, upgrade, current), upgrade); m["status"] != "duplicate" {
    		t.Errorf("retry of evt_1 just under 72h later: status %q, want duplicate", m["status"])
    	}
    	h.now = start.Add(dedupFor)
    	if _, m, _ := h.deliver(signed(h.now, upgrade, current), upgrade); m["status"] != "processed" {
    		t.Errorf("evt_1 again 72h after it was processed: status %q, want processed (IDs are remembered for 72h)", m["status"])
    	}
    	if n := h.count(); n != 2 {
    		t.Errorf("handle ran %d times, want 2", n)
    	}
    }

    func TestHandlerFailure(t *testing.T) {
    	h := newHarness()
    	h.fail = errors.New("pq: connection to 10.0.3.7 refused")
    	code, m, raw := h.deliver(signed(start, upgrade, current), upgrade)
    	if code != 500 || m["error"] == "" || strings.Contains(raw, "10.0.3.7") {
    		t.Errorf("handle fails: %d %s, want 500 with a generic JSON error (not the error text)", code, raw)
    	}
    	h.fail = nil
    	if _, m, raw := h.deliver(signed(start, upgrade, current), upgrade); m["status"] != "processed" {
    		t.Errorf("retry after a failure: %s, want processed (a failed event must not be marked as seen)", raw)
    	}
    }

    func TestBadPayloads(t *testing.T) {
    	for _, body := range []string{`{"id":`, `{"type":"ping"}`, `[1,2]`} {
    		h := newHarness()
    		if code, m, raw := h.deliver(signed(start, body, current), body); code != 400 || m["error"] == "" {
    			t.Errorf("correctly signed payload %s: %d %s, want 400 (bad JSON or no id)", body, code, raw)
    		}
    	}
    	h := newHarness()
    	big := `{"id":"evt_big","type":"ping","data":{"pad":"` + strings.Repeat("x", maxWebhookBody) + `"}}`
    	if code, _, _ := h.deliver(signed(start, big, current), big); code != 413 {
    		t.Errorf("payload over 64 KiB: status %d, want 413", code)
    	}
    }

    func TestConcurrentDeliveries(t *testing.T) {
    	h := newHarness()
    	header := signed(start, upgrade, current)
    	var wg sync.WaitGroup
    	for range 20 {
    		wg.Go(func() { h.deliver(header, upgrade) })
    	}
    	wg.Wait()
    	if n := h.count(); n != 1 {
    		t.Errorf("20 simultaneous deliveries of evt_1: handle ran %d times, want 1", n)
    	}
    }

    func TestConstantTime(t *testing.T) {
    	src, err := os.ReadFile("main.go")
    	if err != nil {
    		t.Skip("can't read main.go")
    	}
    	var code strings.Builder
    	for line := range strings.Lines(string(src)) {
    		before, _, _ := strings.Cut(line, "//")
    		code.WriteString(before)
    	}
    	if !strings.Contains(code.String(), "hmac.Equal") && !strings.Contains(code.String(), "subtle.ConstantTimeCompare") {
    		t.Errorf("compare signatures with hmac.Equal (or subtle.ConstantTimeCompare), never ==")
    	}
    }
---

Squeak Gold payments come from **Gouda Pay**, which now signs its webhooks
instead of sending a static API key. Every webhook carries a header like this:

```
Gouda-Signature: t=1788264000,v1=5d1f3a…,v1=91bc0e…
```

- `t` is the Unix time (seconds) when Gouda Pay sent it.
- Each `v1` is a hex HMAC-SHA256 of **`t + "." + raw body`** with a shared
  secret. During secret rotation Gouda Pay sends one `v1` per secret, and
  Squeak accepts both its current and its previous secret.

Implement `WebhookReceiver`. `NewWebhookReceiver(secrets, now, handle)` takes the
accepted secrets, a clock, and the function that does the actual work.
`ServeHTTP` then:

1. Reads the raw body, at most `maxWebhookBody` bytes (else `413`).
2. Parses `Gouda-Signature`: comma-separated `key=value` items (spaces around
   them are fine), exactly one integer `t`, at least one `v1`. Other keys such
   as `v0` are ignored. Missing or malformed: `400`.
3. Accepts the webhook only if **some** `v1` matches the HMAC under **some**
   secret. Otherwise `401`. Compare in constant time.
4. Rejects a timestamp more than `tolerance` (5 minutes) away from `now()`,
   in either direction, with `401`. This stops an attacker who recorded a
   webhook from replaying it later.
5. Decodes the body as an `Event`. Bad JSON or an empty `id`: `400`.
6. If the event ID was processed less than `dedupFor` (72 hours) ago, answers
   `200 {"status":"duplicate"}` without calling `handle`. Gouda Pay retries for
   days, each time with a fresh timestamp and signature.
7. Calls `handle(ev)`. On error: `500`, and the event is **not** remembered, so
   the retry can succeed. On success, remember the ID and answer
   `200 {"status":"processed"}`.

All errors use `respondWithError`, and none of them may reveal a secret or the
signature you expected.

## Example

```
POST /api/gouda/webhooks
Gouda-Signature: t=1788264000,v1=<valid>
{"id":"evt_1","type":"user.upgraded","data":{"user_id":"pip"}}

200 {"status":"processed"}     (the same request again: {"status":"duplicate"})
```

## Constraints

- Verify the signature before parsing the JSON, and sign the **raw bytes**
  exactly as received.
- Two copies of an event arriving at the same moment must not both be handled.
- The tests control time through `now`, and check for `hmac.Equal` (or
  `subtle.ConstantTimeCompare`).
