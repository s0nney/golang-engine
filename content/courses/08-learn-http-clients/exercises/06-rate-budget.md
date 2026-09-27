---
title: Rate Budget
difficulty: easy
after: http-clients-in-practice
hints:
  - '`h.Get("X-Ratelimit-Remaining")` returns `""` for a missing header, and header names are case-insensitive, so `Get` finds `x-ratelimit-remaining` too.'
  - 'Write a tiny helper that parses one header with `strconv.Atoi` and reports false for errors or negative numbers, then call it three times. `time.Unix(reset, 0).UTC()` turns the Unix seconds into a `time.Time`.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"time"
    )

    // Budget is where we stand against the API's rate limit.
    type Budget struct {
    	Limit     int       // X-Ratelimit-Limit: calls allowed per window
    	Remaining int       // X-Ratelimit-Remaining: calls left in this window
    	Reset     time.Time // X-Ratelimit-Reset: when the window restarts (Unix seconds), in UTC
    }

    // readBudget reads the three X-Ratelimit-* headers from h.
    // It returns false if any of them is missing, isn't a whole number or is negative.
    func readBudget(h http.Header) (Budget, bool) {
    	// ?: parse each header with strconv.Atoi, then build the Budget.
    	return Budget{}, false
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("X-Ratelimit-Limit", "5000")
    		w.Header().Set("X-Ratelimit-Remaining", "4987")
    		w.Header().Set("X-Ratelimit-Reset", "1790503200")
    	}))
    	defer srv.Close()

    	resp, err := srv.Client().Get(srv.URL + "/issues")
    	if err != nil {
    		fmt.Println(err)
    		return
    	}
    	resp.Body.Close()
    	b, ok := readBudget(resp.Header)
    	fmt.Println(b.Limit, b.Remaining, b.Reset, ok)
    	// want: 5000 4987 2026-09-27 10:00:00 +0000 UTC true
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
    	"time"
    )

    // Budget is where we stand against the API's rate limit.
    type Budget struct {
    	Limit     int       // X-Ratelimit-Limit: calls allowed per window
    	Remaining int       // X-Ratelimit-Remaining: calls left in this window
    	Reset     time.Time // X-Ratelimit-Reset: when the window restarts (Unix seconds), in UTC
    }

    func readBudget(h http.Header) (Budget, bool) {
    	limit, ok1 := headerInt(h, "X-Ratelimit-Limit")
    	remaining, ok2 := headerInt(h, "X-Ratelimit-Remaining")
    	reset, ok3 := headerInt(h, "X-Ratelimit-Reset")
    	if !ok1 || !ok2 || !ok3 {
    		return Budget{}, false
    	}
    	return Budget{limit, remaining, time.Unix(int64(reset), 0).UTC()}, true
    }

    func headerInt(h http.Header, name string) (int, bool) {
    	n, err := strconv.Atoi(h.Get(name))
    	return n, err == nil && n >= 0
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("X-Ratelimit-Limit", "5000")
    		w.Header().Set("X-Ratelimit-Remaining", "4987")
    		w.Header().Set("X-Ratelimit-Reset", "1790503200")
    	}))
    	defer srv.Close()

    	resp, err := srv.Client().Get(srv.URL + "/issues")
    	if err != nil {
    		fmt.Println(err)
    		return
    	}
    	resp.Body.Close()
    	b, ok := readBudget(resp.Header)
    	fmt.Println(b.Limit, b.Remaining, b.Reset, ok)
    }
  tests: |
    package main

    import (
    	"net/http"
    	"net/http/httptest"
    	"testing"
    	"time"
    )

    // testHeaders sends raw header lines from a fake API and returns what the client received.
    func testHeaders(t *testing.T, headers map[string]string) http.Header {
    	t.Helper()
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		for k, v := range headers {
    			w.Header()[k] = []string{v} // raw: keeps lower-case names as sent
    		}
    	}))
    	defer srv.Close()
    	resp, err := srv.Client().Get(srv.URL + "/issues")
    	if err != nil {
    		t.Fatalf("GET: %v", err)
    	}
    	resp.Body.Close()
    	return resp.Header
    }

    func TestReadBudget(t *testing.T) {
    	tests := []struct {
    		headers map[string]string
    		want    Budget
    	}{
    		{map[string]string{"X-Ratelimit-Limit": "5000", "X-Ratelimit-Remaining": "4987", "X-Ratelimit-Reset": "1790503200"},
    			Budget{5000, 4987, time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)}},
    		{map[string]string{"x-ratelimit-limit": "60", "x-ratelimit-remaining": "0", "x-ratelimit-reset": "1790503265"},
    			Budget{60, 0, time.Date(2026, 9, 27, 10, 1, 5, 0, time.UTC)}},
    		{map[string]string{"X-RateLimit-Limit": "10", "X-RateLimit-Remaining": "10", "X-RateLimit-Reset": "0"},
    			Budget{10, 10, time.Unix(0, 0).UTC()}},
    	}
    	for _, tt := range tests {
    		got, ok := readBudget(testHeaders(t, tt.headers))
    		if !ok || got.Limit != tt.want.Limit || got.Remaining != tt.want.Remaining || !got.Reset.Equal(tt.want.Reset) {
    			t.Errorf("headers %v: readBudget = %+v, %v\n\twant %+v, true", tt.headers, got, ok, tt.want)
    		}
    		if ok && got.Reset.Location() != time.UTC {
    			t.Errorf("headers %v: Reset is in %v, want UTC", tt.headers, got.Reset.Location())
    		}
    	}
    }

    func TestReadBudgetInvalid(t *testing.T) {
    	good := map[string]string{"X-Ratelimit-Limit": "5000", "X-Ratelimit-Remaining": "4987", "X-Ratelimit-Reset": "1790503200"}
    	tests := []struct {
    		name, value string // "" value means: leave the header out
    	}{
    		{"X-Ratelimit-Limit", ""},
    		{"X-Ratelimit-Remaining", ""},
    		{"X-Ratelimit-Reset", ""},
    		{"X-Ratelimit-Remaining", "lots"},
    		{"X-Ratelimit-Remaining", "-1"},
    		{"X-Ratelimit-Limit", "50.5"},
    		{"X-Ratelimit-Reset", "Sun, 27 Sep 2026 08:00:00 GMT"},
    		{"X-Ratelimit-Reset", "-5"},
    	}
    	for _, tt := range tests {
    		headers := map[string]string{}
    		for k, v := range good {
    			headers[k] = v
    		}
    		delete(headers, tt.name)
    		if tt.value != "" {
    			headers[tt.name] = tt.value
    		}
    		got, ok := readBudget(testHeaders(t, headers))
    		if ok {
    			t.Errorf("headers %v: readBudget = %+v, true, want false", headers, got)
    		}
    	}
    	if got, ok := readBudget(http.Header{}); ok {
    		t.Errorf("readBudget(no headers) = %+v, true, want false", got)
    	}
    }
---

Before a big import, `trackr` wants to know how many API calls it has left, so it
can slow down before the server starts answering `429 Too Many Requests`. The
Trackr API reports that on every response:

```
X-Ratelimit-Limit: 5000
X-Ratelimit-Remaining: 4987
X-Ratelimit-Reset: 1790503200
```

Complete `readBudget(h)`. It parses the three headers into a `Budget`:

- `Limit` and `Remaining` are plain whole numbers.
- `Reset` is a Unix timestamp in seconds. Turn it into a `time.Time` in **UTC**.

It returns `Budget{}, false` if any of the three headers is missing, isn't a
whole number, or is negative.

## Examples

```
X-Ratelimit-Limit: 5000, Remaining: 4987, Reset: 1790503200
    -> Budget{5000, 4987, 2026-09-27 10:00:00 UTC}, true
x-ratelimit-limit: 60, x-ratelimit-remaining: 0, x-ratelimit-reset: 1790503265
    -> Budget{60, 0, 2026-09-27 10:01:05 UTC}, true
Remaining: lots
    -> Budget{}, false
```

## Constraints

- The tests read the headers from real responses sent by a fake API, some with
  lower-case names. Header names are case-insensitive.
- `Remaining: 0` is valid (and the most important value to get right).
