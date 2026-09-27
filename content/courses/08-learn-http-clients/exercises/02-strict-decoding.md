---
title: Strict Decoding
difficulty: easy
after: json
hints:
  - '`encoding/json/v2` already rejects duplicate names, invalid UTF-8 and wrong-case names by default. `json.UnmarshalRead(r, &v)` also reads to EOF, so trailing junk is an error too.'
  - 'Unknown keys are the one rule that isn''t on by default. Pass the option `json.RejectUnknownMembers(true)` as the last argument.'
  - 'After a failed decode `iss` may be half-filled. Return `Issue{}` with the error, not `iss`.'
exercise:
  starter: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"io"
    	"strings"
    )

    // Issue is one Trackr issue as the webhook sends it.
    type Issue struct {
    	ID     int      `json:"id"`
    	Title  string   `json:"title"`
    	State  string   `json:"state"`
    	Labels []string `json:"labels"`
    }

    // decodeIssue reads exactly one JSON issue from r and decodes it strictly.
    // It returns Issue{} and an error if the JSON:
    //
    //   - has a key Issue doesn't know about, like "titel" or "ID",
    //   - repeats a key,
    //   - is followed by anything other than whitespace,
    //   - or isn't valid JSON for an Issue at all.
    func decodeIssue(r io.Reader) (Issue, error) {
    	var iss Issue
    	// ?: decode r into iss with encoding/json/v2, strictly.
    	_ = json.UnmarshalRead
    	return iss, nil
    }

    func main() {
    	fmt.Println(decodeIssue(strings.NewReader(`{"id": 7, "title": "Dark mode", "state": "open"}`)))
    	// want: {7 Dark mode open []} <nil>
    	fmt.Println(decodeIssue(strings.NewReader(`{"id": 7, "titel": "Dark mode"}`)))
    	// want: {0   []} and an error about "titel"
    }
  solution: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"io"
    	"strings"
    )

    // Issue is one Trackr issue as the webhook sends it.
    type Issue struct {
    	ID     int      `json:"id"`
    	Title  string   `json:"title"`
    	State  string   `json:"state"`
    	Labels []string `json:"labels"`
    }

    func decodeIssue(r io.Reader) (Issue, error) {
    	var iss Issue
    	if err := json.UnmarshalRead(r, &iss, json.RejectUnknownMembers(true)); err != nil {
    		return Issue{}, fmt.Errorf("decoding issue: %w", err)
    	}
    	return iss, nil
    }

    func main() {
    	fmt.Println(decodeIssue(strings.NewReader(`{"id": 7, "title": "Dark mode", "state": "open"}`)))
    	fmt.Println(decodeIssue(strings.NewReader(`{"id": 7, "titel": "Dark mode"}`)))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"strings"
    	"testing"
    )

    // fromServer serves body from a fake API and decodes the response with decodeIssue.
    func fromServer(t *testing.T, body string) (Issue, error) {
    	t.Helper()
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("Content-Type", "application/json")
    		fmt.Fprint(w, body)
    	}))
    	defer srv.Close()
    	resp, err := srv.Client().Get(srv.URL + "/webhooks/issue")
    	if err != nil {
    		t.Fatalf("GET: %v", err)
    	}
    	defer resp.Body.Close()
    	return decodeIssue(resp.Body)
    }

    func TestDecodeIssueValid(t *testing.T) {
    	tests := []struct {
    		body string
    		want Issue
    	}{
    		{`{"id": 7, "title": "Dark mode", "state": "open", "labels": ["ui", "good first issue"]}`,
    			Issue{7, "Dark mode", "open", []string{"ui", "good first issue"}}},
    		{`{"title": "Crash on start", "id": 12}`, Issue{ID: 12, Title: "Crash on start"}},
    		{"  {\"id\": 3, \"title\": \"Émojis 🐛 break search\"}\n\n", Issue{ID: 3, Title: "Émojis 🐛 break search"}},
    		{`{}`, Issue{}},
    	}
    	for _, tt := range tests {
    		got, err := fromServer(t, tt.body)
    		if err != nil {
    			t.Errorf("decodeIssue(%s) returned error %v, want %+v", tt.body, err, tt.want)
    			continue
    		}
    		if got.ID != tt.want.ID || got.Title != tt.want.Title || got.State != tt.want.State || !slices.Equal(got.Labels, tt.want.Labels) {
    			t.Errorf("decodeIssue(%s) = %+v, want %+v", tt.body, got, tt.want)
    		}
    	}
    }

    func TestDecodeIssueStrict(t *testing.T) {
    	tests := []struct {
    		why, body string
    	}{
    		{"misspelt key", `{"id": 7, "titel": "Dark mode"}`},
    		{"extra key", `{"id": 7, "title": "Dark mode", "priority": "high"}`},
    		{"wrong case", `{"ID": 7, "title": "Dark mode"}`},
    		{"duplicate key", `{"id": 7, "title": "Dark mode", "title": "Light mode"}`},
    		{"trailing data", `{"id": 7, "title": "Dark mode"} {"id": 8}`},
    		{"trailing junk", `{"id": 7, "title": "Dark mode"} oops`},
    		{"wrong type", `{"id": "7", "title": "Dark mode"}`},
    		{"not an object", `[{"id": 7}]`},
    		{"truncated", `{"id": 7, "title": "Dark`},
    		{"empty body", ``},
    	}
    	for _, tt := range tests {
    		got, err := fromServer(t, tt.body)
    		if err == nil {
    			t.Errorf("%s: decodeIssue(%s) = %+v, nil, want an error", tt.why, tt.body, got)
    			continue
    		}
    		if got.ID != 0 || got.Title != "" || got.State != "" || got.Labels != nil {
    			t.Errorf("%s: decodeIssue(%s) returned %+v with its error, want Issue{} (don't return a half-filled issue)", tt.why, tt.body, got)
    		}
    	}
    }

    func TestDecodeIssueReader(t *testing.T) {
    	got, err := decodeIssue(strings.NewReader(`{"id": 1, "state": "closed"}`))
    	if err != nil || got.ID != 1 || got.State != "closed" {
    		t.Errorf("decodeIssue(strings.Reader) = %+v, %v, want issue 1, state closed", got, err)
    	}
    }
---

Trackr's webhook sends `trackr` a JSON issue whenever something changes. A
client that quietly accepts `"titel"` or two different `"title"` keys will
someday act on data nobody meant to send, so `trackr` decodes webhooks
**strictly**.

Complete `decodeIssue(r)` with `encoding/json/v2`. It reads exactly one JSON
issue from `r` and returns it. It returns `Issue{}` and an error if the JSON:

- has a key `Issue` doesn't know about (`"titel"`, `"priority"`, or `"ID"`,
  since names match case-sensitively),
- repeats a key,
- is followed by anything other than whitespace,
- or isn't a valid JSON object for an `Issue` (wrong types, truncated, empty).

## Examples

```
{"id": 7, "title": "Dark mode", "state": "open"}   -> {7 Dark mode open []}, nil
{"id": 7, "titel": "Dark mode"}                    -> Issue{}, error (unknown key)
{"id": 7, "title": "a", "title": "b"}              -> Issue{}, error (duplicate)
{"id": 7} {"id": 8}                                -> Issue{}, error (trailing data)
```

## Constraints

- The tests serve each body from a fake API and pass you `resp.Body`.
- Keys may appear in any order, and missing keys are fine: they keep their zero
  values.
- One line of decoding does almost all of it. The only rule you have to switch on
  yourself is rejecting unknown keys.
