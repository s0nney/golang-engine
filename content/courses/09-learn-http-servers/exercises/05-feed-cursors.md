---
title: Feed Cursors
difficulty: hard
after: storage
hints:
  - 'An offset (`?page=3`) breaks as soon as someone posts: every squeak shifts down one place and the reader sees a duplicate. A **cursor** remembers *where you stopped* instead: the `(CreatedAt, ID)` of the last squeak on the page. The next page is every squeak that sorts strictly after that key, no matter what was added or deleted since.'
  - 'Encode the key as text, for example `fmt.Sprintf("%d_%s", sq.CreatedAt.UnixNano(), sq.ID)`, then wrap it in `base64.RawURLEncoding` so it''s opaque and URL-safe. Decoding reverses each step (`strings.Cut`, `strconv.ParseInt`, `uuid.Parse`, `time.Unix(0, n)`), and **any** failure is a 400.'
  - 'With `s.squeaks` sorted newest first, a squeak comes after key `k` when `sq.CreatedAt.Before(k.at)`, or the times are `Equal` and `sq.ID.Compare(k.id) < 0`. Take up to `limit+1` of those under `s.mu.RLock()`: if you got the extra one, there''s a next page, and its cursor is built from the last squeak you actually return.'
exercise:
  starter: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"sync"
    	"time"
    	"uuid"
    )

    type Squeak struct {
    	ID        uuid.UUID `json:"id"`
    	Author    string    `json:"author"`
    	Body      string    `json:"body"`
    	CreatedAt time.Time `json:"created_at"`
    }

    // newer reports the feed order: newest first, and by descending ID for
    // squeaks created at the same instant.
    func newer(a, b Squeak) int {
    	if c := b.CreatedAt.Compare(a.CreatedAt); c != 0 {
    		return c
    	}
    	return b.ID.Compare(a.ID)
    }

    // Store keeps squeaks sorted by newer. It's safe for concurrent use.
    type Store struct {
    	mu      sync.RWMutex
    	squeaks []Squeak
    }

    func (s *Store) Add(sq Squeak) {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	i, _ := slices.BinarySearchFunc(s.squeaks, sq, newer)
    	s.squeaks = slices.Insert(s.squeaks, i, sq)
    }

    func (s *Store) Delete(id uuid.UUID) {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	s.squeaks = slices.DeleteFunc(s.squeaks, func(sq Squeak) bool { return sq.ID == id })
    }

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
    		w.WriteHeader(http.StatusInternalServerError)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, map[string]string{"error": msg})
    }

    // handleList serves GET /api/squeaks?limit=N&cursor=C.
    func (s *Store) handleList(w http.ResponseWriter, r *http.Request) {
    	respondWithError(w, http.StatusNotImplemented, "not implemented")
    }

    func main() {
    	s := &Store{}
    	start := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    	for i := range 5 {
    		s.Add(Squeak{ID: uuid.NewV7(), Author: "pip", Body: fmt.Sprintf("squeak #%d", i+1), CreatedAt: start.Add(time.Duration(i) * time.Minute)})
    	}
    	rec := httptest.NewRecorder()
    	s.handleList(rec, httptest.NewRequest("GET", "/api/squeaks?limit=2", nil))
    	fmt.Println(rec.Code, rec.Body.String())
    }
  solution: |
    package main

    import (
    	"encoding/base64"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"strconv"
    	"strings"
    	"sync"
    	"time"
    	"uuid"
    )

    type Squeak struct {
    	ID        uuid.UUID `json:"id"`
    	Author    string    `json:"author"`
    	Body      string    `json:"body"`
    	CreatedAt time.Time `json:"created_at"`
    }

    func newer(a, b Squeak) int {
    	if c := b.CreatedAt.Compare(a.CreatedAt); c != 0 {
    		return c
    	}
    	return b.ID.Compare(a.ID)
    }

    type Store struct {
    	mu      sync.RWMutex
    	squeaks []Squeak
    }

    func (s *Store) Add(sq Squeak) {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	i, _ := slices.BinarySearchFunc(s.squeaks, sq, newer)
    	s.squeaks = slices.Insert(s.squeaks, i, sq)
    }

    func (s *Store) Delete(id uuid.UUID) {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	s.squeaks = slices.DeleteFunc(s.squeaks, func(sq Squeak) bool { return sq.ID == id })
    }

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
    		w.WriteHeader(http.StatusInternalServerError)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, map[string]string{"error": msg})
    }

    // cursorKey is the position of the last squeak a client has seen.
    type cursorKey struct {
    	at time.Time
    	id uuid.UUID
    }

    var errBadCursor = errors.New("invalid cursor")

    func encodeCursor(sq Squeak) string {
    	raw := fmt.Sprintf("%d_%s", sq.CreatedAt.UnixNano(), sq.ID)
    	return base64.RawURLEncoding.EncodeToString([]byte(raw))
    }

    func decodeCursor(s string) (cursorKey, error) {
    	raw, err := base64.RawURLEncoding.DecodeString(s)
    	if err != nil {
    		return cursorKey{}, errBadCursor
    	}
    	nanos, id, ok := strings.Cut(string(raw), "_")
    	if !ok {
    		return cursorKey{}, errBadCursor
    	}
    	n, err := strconv.ParseInt(nanos, 10, 64)
    	if err != nil {
    		return cursorKey{}, errBadCursor
    	}
    	u, err := uuid.Parse(id)
    	if err != nil {
    		return cursorKey{}, errBadCursor
    	}
    	return cursorKey{at: time.Unix(0, n), id: u}, nil
    }

    // after reports whether sq comes after k in feed order.
    func (k cursorKey) after(sq Squeak) bool {
    	return sq.CreatedAt.Before(k.at) || sq.CreatedAt.Equal(k.at) && sq.ID.Compare(k.id) < 0
    }

    type pageResponse struct {
    	Squeaks    []Squeak `json:"squeaks"`
    	NextCursor string   `json:"next_cursor,omitzero"`
    }

    func (s *Store) handleList(w http.ResponseWriter, r *http.Request) {
    	q := r.URL.Query()
    	limit := 20
    	if v := q.Get("limit"); v != "" {
    		n, err := strconv.Atoi(v)
    		if err != nil || n < 1 || n > 100 {
    			respondWithError(w, http.StatusBadRequest, "limit must be a number from 1 to 100")
    			return
    		}
    		limit = n
    	}
    	var key *cursorKey
    	if v := q.Get("cursor"); v != "" {
    		k, err := decodeCursor(v)
    		if err != nil {
    			respondWithError(w, http.StatusBadRequest, "invalid cursor")
    			return
    		}
    		key = &k
    	}

    	s.mu.RLock()
    	start := 0
    	if key != nil {
    		start = len(s.squeaks)
    		if i := slices.IndexFunc(s.squeaks, key.after); i >= 0 {
    			start = i
    		}
    	}
    	end := min(start+limit+1, len(s.squeaks))
    	page := slices.Clone(s.squeaks[start:end])
    	s.mu.RUnlock()

    	resp := pageResponse{Squeaks: page}
    	if len(page) > limit {
    		resp.Squeaks = page[:limit]
    		resp.NextCursor = encodeCursor(page[limit-1])
    	}
    	respondWithJSON(w, http.StatusOK, resp)
    }

    func main() {
    	s := &Store{}
    	start := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    	for i := range 5 {
    		s.Add(Squeak{ID: uuid.NewV7(), Author: "pip", Body: fmt.Sprintf("squeak #%d", i+1), CreatedAt: start.Add(time.Duration(i) * time.Minute)})
    	}
    	rec := httptest.NewRecorder()
    	s.handleList(rec, httptest.NewRequest("GET", "/api/squeaks?limit=2", nil))
    	fmt.Println(rec.Code, rec.Body.String())
    }
  tests: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"net/http/httptest"
    	"net/url"
    	"slices"
    	"strings"
    	"sync"
    	"testing"
    	"time"
    	"uuid"
    )

    var t0 = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

    type page struct {
    	code   int
    	raw    string
    	ids    []uuid.UUID
    	next   string
    	hasNxt bool
    }

    func get(t *testing.T, s *Store, limit, cursor string) page {
    	t.Helper()
    	q := url.Values{}
    	if limit != "" {
    		q.Set("limit", limit)
    	}
    	if cursor != "" {
    		q.Set("cursor", cursor)
    	}
    	target := "/api/squeaks"
    	if len(q) > 0 {
    		target += "?" + q.Encode()
    	}
    	rec := httptest.NewRecorder()
    	s.handleList(rec, httptest.NewRequest("GET", target, nil))
    	p := page{code: rec.Code, raw: rec.Body.String()}
    	if rec.Code != 200 {
    		return p
    	}
    	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
    		t.Fatalf("GET %s: Content-Type = %q, want application/json", target, ct)
    	}
    	var body struct {
    		Squeaks    []Squeak `json:"squeaks"`
    		NextCursor *string  `json:"next_cursor"`
    	}
    	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
    		t.Fatalf("GET %s: body %q isn't {\"squeaks\":[...],\"next_cursor\":\"...\"}: %v", target, p.raw, err)
    	}
    	for _, sq := range body.Squeaks {
    		p.ids = append(p.ids, sq.ID)
    	}
    	if body.NextCursor != nil {
    		p.next, p.hasNxt = *body.NextCursor, true
    	}
    	return p
    }

    // seed adds n squeaks; groups of `same` share a timestamp. It returns them in feed order.
    func seed(s *Store, n, same int) []Squeak {
    	var all []Squeak
    	for i := range n {
    		sq := Squeak{ID: uuid.NewV4(), Author: "pip", Body: fmt.Sprint("squeak ", i), CreatedAt: t0.Add(time.Duration(i/same) * time.Second)}
    		s.Add(sq)
    		all = append(all, sq)
    	}
    	slices.SortFunc(all, newer)
    	return all
    }

    func ids(sqs []Squeak) []uuid.UUID {
    	var out []uuid.UUID
    	for _, sq := range sqs {
    		out = append(out, sq.ID)
    	}
    	return out
    }

    func TestFirstPage(t *testing.T) {
    	s := &Store{}
    	all := seed(s, 30, 1)
    	p := get(t, s, "", "")
    	if p.code != 200 {
    		t.Fatalf("GET /api/squeaks = %d %s, want 200", p.code, p.raw)
    	}
    	if !slices.Equal(p.ids, ids(all[:20])) {
    		t.Errorf("GET /api/squeaks with no limit returned %d squeaks, want the 20 newest in order (newest first)", len(p.ids))
    	}
    	if !p.hasNxt || p.next == "" {
    		t.Errorf("first page of 30 squeaks: next_cursor missing, want one")
    	}
    	p = get(t, s, "5", "")
    	if !slices.Equal(p.ids, ids(all[:5])) {
    		t.Errorf("GET ?limit=5 returned the wrong squeaks: want the 5 newest, newest first")
    	}
    }

    func TestEmptyStore(t *testing.T) {
    	p := get(t, &Store{}, "", "")
    	if p.code != 200 || len(p.ids) != 0 || p.hasNxt {
    		t.Errorf("empty store: got %d %s, want 200 with no squeaks and no next_cursor", p.code, p.raw)
    	}
    	if !strings.Contains(p.raw, `"squeaks":[]`) {
    		t.Errorf("empty store: body = %s, want \"squeaks\":[] (an empty array, not null)", p.raw)
    	}
    }

    func TestWalkAllPages(t *testing.T) {
    	for _, tt := range []struct{ n, same, limit int }{{60, 3, 7}, {14, 1, 7}, {14, 14, 5}, {1, 1, 1}, {100, 100, 100}} {
    		s := &Store{}
    		all := seed(s, tt.n, tt.same)
    		var got []uuid.UUID
    		cursor, pages := "", 0
    		for {
    			p := get(t, s, fmt.Sprint(tt.limit), cursor)
    			if p.code != 200 {
    				t.Fatalf("%d squeaks, limit %d, page %d: status %d %s", tt.n, tt.limit, pages+1, p.code, p.raw)
    			}
    			pages++
    			got = append(got, p.ids...)
    			if !p.hasNxt {
    				break
    			}
    			if len(p.ids) != tt.limit {
    				t.Fatalf("%d squeaks, limit %d: page %d has %d squeaks and a next_cursor; only the last page may be short", tt.n, tt.limit, pages, len(p.ids))
    			}
    			if pages > tt.n+1 {
    				t.Fatalf("%d squeaks, limit %d: still paging after %d pages; the cursor doesn't move forward", tt.n, tt.limit, pages)
    			}
    			cursor = p.next
    		}
    		if !slices.Equal(got, ids(all)) {
    			t.Errorf("%d squeaks (%d per timestamp), limit %d: walking every page gave %d squeaks, want all %d exactly once, newest first then by descending ID", tt.n, tt.same, tt.limit, len(got), tt.n)
    		}
    		if want := (tt.n + tt.limit - 1) / tt.limit; pages != want {
    			t.Errorf("%d squeaks, limit %d: took %d requests, want %d (no next_cursor on the page that holds the last squeak)", tt.n, tt.limit, pages, want)
    		}
    	}
    }

    func TestStableUnderChanges(t *testing.T) {
    	s := &Store{}
    	all := seed(s, 30, 1)
    	p1 := get(t, s, "10", "")
    	if p1.code != 200 || !p1.hasNxt {
    		t.Fatalf("first page: %d %s", p1.code, p1.raw)
    	}
    	last := all[9] // the cursor's squeak

    	// While the reader looks at page 1: new squeaks arrive, and some old ones are deleted.
    	for i := range 5 {
    		s.Add(Squeak{ID: uuid.NewV4(), Body: "brand new", CreatedAt: t0.Add(time.Hour + time.Duration(i)*time.Second)})
    	}
    	s.Delete(last.ID)
    	s.Delete(all[10].ID)
    	sameTimeBefore := Squeak{ID: uuid.Max(), Body: "same time, sorts before the cursor", CreatedAt: last.CreatedAt}
    	sameTimeAfter := Squeak{ID: uuid.Nil(), Body: "same time, sorts after the cursor", CreatedAt: last.CreatedAt}
    	s.Add(sameTimeBefore)
    	s.Add(sameTimeAfter)

    	p2 := get(t, s, "10", p1.next)
    	if p2.code != 200 {
    		t.Fatalf("second page after changes: %d %s, want 200 (the cursor's own squeak was deleted, but its position still works)", p2.code, p2.raw)
    	}
    	want := append([]uuid.UUID{sameTimeAfter.ID}, ids(all[11:20])...)
    	if !slices.Equal(p2.ids, want) {
    		for i, id := range p2.ids {
    			if id == all[0].ID || slices.Contains(p1.ids, id) {
    				t.Fatalf("second page repeats squeak %d from the first page: use a (created_at, id) cursor, not an offset", i)
    			}
    		}
    		t.Errorf("second page after new posts and deletes:\n got %v\nwant %v\n(the squeak with the cursor's timestamp and a smaller ID comes first; squeaks newer than the cursor never appear)", p2.ids, want)
    	}
    }

    func TestBadParams(t *testing.T) {
    	s := &Store{}
    	seed(s, 30, 1)
    	p := get(t, s, "3", "")
    	good := p.next
    	if p.code != 200 || len(good) < 8 {
    		t.Fatalf("GET ?limit=3 on 30 squeaks = %d %s, want 200 with a next_cursor of at least 8 characters", p.code, p.raw)
    	}
    	for _, tt := range []struct{ limit, cursor string }{
    		{"0", ""}, {"-1", ""}, {"101", ""}, {"ten", ""}, {"2.5", ""},
    		{"", "not-a-cursor"}, {"", "!!!!"}, {"", good[:len(good)-5]}, {"", good + "AAAA"},
    	} {
    		p := get(t, s, tt.limit, tt.cursor)
    		if p.code != 400 {
    			t.Errorf("limit=%q cursor=%q: status %d, want 400", tt.limit, tt.cursor, p.code)
    			continue
    		}
    		var e map[string]string
    		if err := json.Unmarshal([]byte(p.raw), &e); err != nil || e["error"] == "" {
    			t.Errorf("limit=%q cursor=%q: body %s, want a JSON {\"error\":\"...\"}", tt.limit, tt.cursor, p.raw)
    		}
    	}
    	if p := get(t, s, "100", ""); p.code != 200 || len(p.ids) != 30 {
    		t.Errorf("limit=100 on 30 squeaks: %d with %d squeaks, want 200 with all 30", p.code, len(p.ids))
    	}
    }

    func TestConcurrentReadsAndWrites(t *testing.T) {
    	s := &Store{}
    	seed(s, 50, 2)
    	var wg sync.WaitGroup
    	for i := range 20 {
    		wg.Go(func() { s.Add(Squeak{ID: uuid.NewV4(), CreatedAt: t0.Add(time.Duration(i) * time.Millisecond)}) })
    		wg.Go(func() {
    			rec := httptest.NewRecorder()
    			s.handleList(rec, httptest.NewRequest("GET", "/api/squeaks?limit=7", nil))
    		})
    	}
    	wg.Wait()
    }
---

Squeak's home feed is infinite scroll: the app loads 20 squeaks, and when the
reader reaches the bottom it asks for the next 20. The first version used
`?page=2`, and readers kept seeing the same squeak twice: every new post pushed
everything down one place between requests.

Replace it with **cursor pagination**. Complete `handleList` for
`GET /api/squeaks?limit=N&cursor=C`:

- Squeaks are listed in feed order: newest `CreatedAt` first, and by
  **descending ID** among squeaks created at the same instant. `Store` already
  keeps `s.squeaks` in this order (see `newer`).
- `limit` is optional (default `20`) and must be a whole number from `1` to
  `100`.
- Without `cursor`, return the first `limit` squeaks. With a cursor, return the
  first `limit` squeaks that come **strictly after** the squeak the cursor
  points to.
- Respond `200` with `{"squeaks": [...], "next_cursor": "..."}`. Include
  `next_cursor` only if more squeaks follow this page. It points at the last
  squeak on the page. `squeaks` is `[]`, never `null`, when there are none.
- A bad `limit` or a cursor you can't decode is `400` with a JSON
  `{"error": "..."}` body.

The cursor format is up to you, but it must keep working when things change
between requests: new squeaks arrive at the top, and squeaks (even the one the
cursor points at) get deleted.

## Example

```
GET /api/squeaks?limit=2
200 {"squeaks":[{"id":"…","body":"squeak #5",…},{"id":"…","body":"squeak #4",…}],
     "next_cursor":"MTc1Njcy…"}

GET /api/squeaks?limit=2&cursor=MTc1Njcy…
200 {"squeaks":[{…"squeak #3"…},{…"squeak #2"…}],"next_cursor":"…"}
```

## Constraints

- Many squeaks may share a timestamp. Walking every page must return each
  squeak exactly once, in feed order, and the page holding the last squeak has
  no `next_cursor`.
- Handlers run concurrently with `Add` and `Delete`: read `s.squeaks` only
  while holding `s.mu`.
