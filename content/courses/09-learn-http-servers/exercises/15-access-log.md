---
title: Access Log
difficulty: medium
after: production-readiness
hints:
  - 'A handler never tells you its status: it just calls `w.WriteHeader` (or only `w.Write`, which means 200). Wrap `w` in your own struct that embeds `http.ResponseWriter`, remembers the first status, and adds up the bytes each `Write` returns. Pass the wrapper to `next`.'
  - '`ServeMux` fills in `r.Pattern` on the request it was given, so after `next.ServeHTTP(sw, r)` returns, the outer middleware can read `r.Pattern`. Measure the duration with `time.Now()` before and `time.Since(start)` after, and choose `slog.LevelError` for statuses of 500 and up.'
  - 'Embedding hides the real writer''s optional methods, so `http.NewResponseController(w).Flush()` fails through a plain wrapper. Add `func (sw *statusWriter) Unwrap() http.ResponseWriter { return sw.ResponseWriter }` and the controller will find the original.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"log/slog"
    	"net/http"
    	"net/http/httptest"
    	"os"
    )

    func accessLog(logger *slog.Logger, next http.Handler) http.Handler {
    	return next
    }

    func main() {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks/{id}", func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, `{"id":"42","body":"cheese"}`)
    	})
    	h := accessLog(slog.New(slog.NewJSONHandler(os.Stdout, nil)), mux)
    	for _, target := range []string{"/api/squeaks/42?token=sekrit", "/nope"} {
    		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", target, nil))
    	}
    }
  solution: |
    package main

    import (
    	"context"
    	"fmt"
    	"log/slog"
    	"net/http"
    	"net/http/httptest"
    	"os"
    	"time"
    )

    // statusWriter records the status and body size a handler writes.
    type statusWriter struct {
    	http.ResponseWriter
    	status int
    	bytes  int
    }

    func (sw *statusWriter) WriteHeader(code int) {
    	if sw.status == 0 {
    		sw.status = code
    	}
    	sw.ResponseWriter.WriteHeader(code)
    }

    func (sw *statusWriter) Write(b []byte) (int, error) {
    	if sw.status == 0 {
    		sw.status = http.StatusOK
    	}
    	n, err := sw.ResponseWriter.Write(b)
    	sw.bytes += n
    	return n, err
    }

    // Unwrap lets http.ResponseController reach the real writer.
    func (sw *statusWriter) Unwrap() http.ResponseWriter { return sw.ResponseWriter }

    func accessLog(logger *slog.Logger, next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		start := time.Now()
    		sw := &statusWriter{ResponseWriter: w}
    		next.ServeHTTP(sw, r)
    		if sw.status == 0 {
    			sw.status = http.StatusOK
    		}
    		route := r.Pattern
    		if route == "" {
    			route = "unmatched"
    		}
    		level := slog.LevelInfo
    		if sw.status >= 500 {
    			level = slog.LevelError
    		}
    		logger.LogAttrs(context.Background(), level, "request",
    			slog.String("method", r.Method),
    			slog.String("route", route),
    			slog.Int("status", sw.status),
    			slog.Int("bytes", sw.bytes),
    			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
    		)
    	})
    }

    func main() {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks/{id}", func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, `{"id":"42","body":"cheese"}`)
    	})
    	h := accessLog(slog.New(slog.NewJSONHandler(os.Stdout, nil)), mux)
    	for _, target := range []string{"/api/squeaks/42?token=sekrit", "/nope"} {
    		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", target, nil))
    	}
    }
  tests: |
    package main

    import (
    	"bytes"
    	"encoding/json/v2"
    	"io"
    	"log/slog"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    func squeakMux() *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks/{id}", func(w http.ResponseWriter, r *http.Request) {
    		io.WriteString(w, `{"id":"`+r.PathValue("id")+`"}`) // no WriteHeader: implicitly 200
    	})
    	mux.HandleFunc("POST /api/squeaks", func(w http.ResponseWriter, r *http.Request) {
    		w.WriteHeader(http.StatusCreated)
    		w.WriteHeader(http.StatusTeapot) // ignored by net/http: only the first status counts
    		io.WriteString(w, "created")
    		io.WriteString(w, "!")
    	})
    	mux.HandleFunc("DELETE /api/squeaks/{id}", func(w http.ResponseWriter, r *http.Request) {
    		w.WriteHeader(http.StatusNoContent)
    	})
    	mux.HandleFunc("GET /empty", func(w http.ResponseWriter, r *http.Request) {})
    	mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) {
    		http.Error(w, "database on fire", http.StatusInternalServerError)
    	})
    	mux.HandleFunc("GET /stream", func(w http.ResponseWriter, r *http.Request) {
    		io.WriteString(w, "data: 1\n\n")
    		if err := http.NewResponseController(w).Flush(); err != nil {
    			w.Header().Set("X-Flush-Error", err.Error())
    			io.WriteString(w, "flush failed")
    		}
    	})
    	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
    		time.Sleep(250 * time.Millisecond)
    		io.WriteString(w, "zzz")
    	})
    	return mux
    }

    type logEntry map[string]any

    func run(t *testing.T, method, target string, header map[string]string) (logEntry, *httptest.ResponseRecorder, string) {
    	t.Helper()
    	var buf bytes.Buffer
    	h := accessLog(slog.New(slog.NewJSONHandler(&buf, nil)), squeakMux())
    	req := httptest.NewRequest(method, target, nil)
    	for k, v := range header {
    		req.Header.Set(k, v)
    	}
    	rec := httptest.NewRecorder()
    	h.ServeHTTP(rec, req)
    	var lines []string
    	for line := range strings.Lines(buf.String()) {
    		lines = append(lines, line)
    	}
    	if len(lines) != 1 {
    		t.Fatalf("%s %s: logged %d lines, want exactly 1:\n%s", method, target, len(lines), buf.String())
    	}
    	var e logEntry
    	if err := json.Unmarshal([]byte(lines[0]), &e); err != nil {
    		t.Fatalf("%s %s: log line %q isn't JSON: %v", method, target, lines[0], err)
    	}
    	return e, rec, buf.String()
    }

    func TestLogFields(t *testing.T) {
    	tests := []struct {
    		method, target string
    		route, level   string
    		status, bytes  float64
    	}{
    		{"GET", "/api/squeaks/42", "GET /api/squeaks/{id}", "INFO", 200, 11},
    		{"POST", "/api/squeaks", "POST /api/squeaks", "INFO", 201, 8},
    		{"DELETE", "/api/squeaks/7", "DELETE /api/squeaks/{id}", "INFO", 204, 0},
    		{"GET", "/empty", "GET /empty", "INFO", 200, 0},
    		{"GET", "/boom", "GET /boom", "ERROR", 500, 17},
    		{"GET", "/nope", "unmatched", "INFO", 404, 19},
    		{"PUT", "/api/squeaks/42", "unmatched", "INFO", 405, 19},
    	}
    	for _, tt := range tests {
    		e, rec, raw := run(t, tt.method, tt.target, nil)
    		want := logEntry{"msg": "request", "level": tt.level, "method": tt.method, "route": tt.route, "status": tt.status, "bytes": tt.bytes}
    		for k, v := range want {
    			if e[k] != v {
    				t.Errorf("%s %s: log %q = %v, want %v\n  log line: %s", tt.method, tt.target, k, e[k], v, raw)
    			}
    		}
    		if _, ok := e["duration_ms"].(float64); !ok {
    			t.Errorf("%s %s: log has no numeric duration_ms: %s", tt.method, tt.target, raw)
    		}
    		if rec.Code != int(tt.status) {
    			t.Errorf("%s %s: client got status %d, want %d (the middleware must pass the response through)", tt.method, tt.target, rec.Code, int(tt.status))
    		}
    	}
    }

    func TestResponsePassesThrough(t *testing.T) {
    	_, rec, _ := run(t, "POST", "/api/squeaks", nil)
    	if rec.Body.String() != "created!" {
    		t.Errorf("client got body %q, want %q", rec.Body.String(), "created!")
    	}
    	_, rec, _ = run(t, "GET", "/stream", nil)
    	if msg := rec.Header().Get("X-Flush-Error"); msg != "" || !rec.Flushed {
    		t.Errorf("http.NewResponseController(w).Flush() through your wrapper failed (%q): give the wrapper an Unwrap method", msg)
    	}
    }

    func TestNoSecretsInLogs(t *testing.T) {
    	_, _, raw := run(t, "GET", "/api/squeaks/42?token=sq_live_sekrit&email=pip%40squeak.example", map[string]string{
    		"Authorization": "Bearer eyJhbGciOiJIUzI1NiJ9.sekrit",
    		"Cookie":        "session=sekrit-cookie",
    	})
    	for _, s := range []string{"sekrit", "token", "pip@", "eyJ", "/api/squeaks/42"} {
    		if strings.Contains(raw, s) {
    			t.Errorf("log line contains %q: log the route pattern, not the URL, and never headers\n  log line: %s", s, raw)
    		}
    	}
    }

    func TestDuration(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		e, _, raw := run(t, "GET", "/slow", nil)
    		if e["duration_ms"] != float64(250) {
    			t.Errorf("handler took 250ms (fake time): duration_ms = %v, want 250\n  log line: %s", e["duration_ms"], raw)
    		}
    		e, _, raw = run(t, "GET", "/empty", nil)
    		if e["duration_ms"] != float64(0) {
    			t.Errorf("instant handler: duration_ms = %v, want 0\n  log line: %s", e["duration_ms"], raw)
    		}
    	})
    }
---

Squeak is in production, and when a mouse reports "my squeak didn't post", the
first question is: what did the server see? Add an **access log**: one
structured `log/slog` entry per request.

Write `accessLog(logger, next)`. After `next` has handled the request, log one
entry with the message `"request"` and these attributes:

| attribute | value |
|---|---|
| `method` | the request method |
| `route` | the `ServeMux` pattern that matched, like `GET /api/squeaks/{id}`, or `"unmatched"` if none did |
| `status` | the status the handler sent (`200` if it never called `WriteHeader`) |
| `bytes` | how many body bytes the handler wrote |
| `duration_ms` | how long `next` took, in whole milliseconds |

Log at **info** level, or **error** level when the status is `500` or more.

## Example

```
GET /api/squeaks/42?token=sekrit
{"time":"…","level":"INFO","msg":"request","method":"GET","route":"GET /api/squeaks/{id}","status":200,"bytes":27,"duration_ms":0}

GET /nope
{"time":"…","level":"INFO","msg":"request","method":"GET","route":"unmatched","status":404,"bytes":19,"duration_ms":0}
```

## Constraints

- The response must reach the client unchanged, and handlers that stream must
  still be able to flush with `http.NewResponseController(w).Flush()`.
- Log the **route pattern**, never the raw URL: query strings and paths can
  hold tokens and email addresses. Never log request headers.
- The duration test runs in a `testing/synctest` bubble with a handler that
  sleeps 250ms of fake time, so use `time.Now()` and `time.Since`.
