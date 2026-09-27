---
title: Decoding Responses
quiz:
  - question: |
      What's wrong with this?

      ```go
      var issues []Issue
      err := json.NewDecoder(resp.Body).Decode(issues)
      ```
    options:
      - text: Nothing
      - text: '`Decode` needs a pointer: `&issues`'
        correct: true
      - text: You must read the body with `io.ReadAll` first
      - text: '`NewDecoder` only works on files'
    explanation: |
      Like `Unmarshal`, `Decode` has to change your variable, so it needs its address.
      Passing the slice itself returns an `InvalidUnmarshalError`.
  - question: The server replies `500` with an HTML error page. What should `trackr` do first?
    options:
      - text: Decode it as JSON and let the decoder fail
      - text: Check `resp.StatusCode` and return an error before decoding
        correct: true
      - text: Retry the decode with `json.Unmarshal`
    explanation: |
      Check the status before you trust the body. Decoding an HTML page gives a
      confusing "invalid character '<'" error that hides the real problem: the server
      said 500.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    // Issue is one Trackr issue. Add struct tags so it matches the API's JSON.
    type Issue struct {
    	ID          int
    	Title       string
    	State       string
    	StoryPoints int
    	Labels      []string
    }

    // listIssues fetches baseURL + "/issues" and decodes the JSON array.
    func listIssues(baseURL string) ([]Issue, error) {
    	// ?
    	return nil, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("Content-Type", "application/json")
    		fmt.Fprint(w, `[
    			{"id": 1, "title": "Set up CI", "state": "closed", "story_points": 2, "labels": []},
    			{"id": 7, "title": "Dark mode", "state": "open", "story_points": 5, "labels": ["ui"]}
    		]`)
    	}))
    	defer srv.Close()

    	issues, err := listIssues(srv.URL)
    	if err != nil {
    		fmt.Println("error:", err)
    		return
    	}
    	fmt.Printf("%d issues\n", len(issues))
    	for _, iss := range issues {
    		fmt.Printf("#%d [%s] %s (%d pts) %v\n", iss.ID, iss.State, iss.Title, iss.StoryPoints, iss.Labels)
    	}
    }
  solution: |
    package main

    import (
    	"encoding/json"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    type Issue struct {
    	ID          int      `json:"id"`
    	Title       string   `json:"title"`
    	State       string   `json:"state"`
    	StoryPoints int      `json:"story_points"`
    	Labels      []string `json:"labels"`
    }

    func listIssues(baseURL string) ([]Issue, error) {
    	resp, err := http.Get(baseURL + "/issues")
    	if err != nil {
    		return nil, err
    	}
    	defer resp.Body.Close()

    	if resp.StatusCode != http.StatusOK {
    		return nil, fmt.Errorf("listing issues: unexpected status %s", resp.Status)
    	}

    	var issues []Issue
    	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
    		return nil, fmt.Errorf("decoding issues: %w", err)
    	}
    	return issues, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("Content-Type", "application/json")
    		fmt.Fprint(w, `[
    			{"id": 1, "title": "Set up CI", "state": "closed", "story_points": 2, "labels": []},
    			{"id": 7, "title": "Dark mode", "state": "open", "story_points": 5, "labels": ["ui"]}
    		]`)
    	}))
    	defer srv.Close()

    	issues, err := listIssues(srv.URL)
    	if err != nil {
    		fmt.Println("error:", err)
    		return
    	}
    	fmt.Printf("%d issues\n", len(issues))
    	for _, iss := range issues {
    		fmt.Printf("#%d [%s] %s (%d pts) %v\n", iss.ID, iss.State, iss.Title, iss.StoryPoints, iss.Labels)
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"testing"
    )

    func serve(status int, body string) *httptest.Server {
    	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		if r.URL.Path != "/issues" || r.Method != http.MethodGet {
    			http.Error(w, "not found", http.StatusNotFound)
    			return
    		}
    		w.Header().Set("Content-Type", "application/json")
    		w.WriteHeader(status)
    		fmt.Fprint(w, body)
    	}))
    }

    func TestListIssues(t *testing.T) {
    	srv := serve(http.StatusOK, `[
    		{"id": 3, "title": "Flaky login test", "state": "open", "story_points": 8, "labels": ["bug", "ci"], "reporter": "bo"},
    		{"id": 11, "title": "Rename project", "state": "closed", "story_points": 1, "labels": []}
    	]`)
    	defer srv.Close()

    	got, err := listIssues(srv.URL)
    	if err != nil {
    		t.Fatalf("listIssues returned error %v, want nil", err)
    	}
    	if len(got) != 2 {
    		t.Fatalf("listIssues returned %d issues, want 2", len(got))
    	}
    	first := got[0]
    	if first.ID != 3 || first.Title != "Flaky login test" || first.State != "open" {
    		t.Errorf("first issue = %+v, want ID 3, Title %q, State %q", first, "Flaky login test", "open")
    	}
    	if first.StoryPoints != 8 {
    		t.Errorf("first issue StoryPoints = %d, want 8 (does the tag match \"story_points\"?)", first.StoryPoints)
    	}
    	if !slices.Equal(first.Labels, []string{"bug", "ci"}) {
    		t.Errorf("first issue Labels = %q, want [bug ci]", first.Labels)
    	}
    	if got[1].ID != 11 || got[1].StoryPoints != 1 {
    		t.Errorf("second issue = %+v, want ID 11 with 1 story point", got[1])
    	}
    }

    func TestListIssuesEmpty(t *testing.T) {
    	srv := serve(http.StatusOK, `[]`)
    	defer srv.Close()
    	got, err := listIssues(srv.URL)
    	if err != nil || len(got) != 0 {
    		t.Errorf("listIssues on [] = (%v, %v), want an empty list and nil error", got, err)
    	}
    }

    func TestListIssuesBadStatus(t *testing.T) {
    	srv := serve(http.StatusInternalServerError, `<html>oops</html>`)
    	defer srv.Close()
    	if got, err := listIssues(srv.URL); err == nil {
    		t.Errorf("server replied 500: listIssues = (%v, nil), want an error", got)
    	}
    }

    func TestListIssuesBadJSON(t *testing.T) {
    	srv := serve(http.StatusOK, `[{"id": "seven"}]`)
    	defer srv.Close()
    	if got, err := listIssues(srv.URL); err == nil {
    		t.Errorf("server sent an id that's a string: listIssues = (%v, nil), want the decode error", got)
    	}
    }
---

You know how to `GET` a body and how to decode JSON. Put them together and `trackr` can
list issues.

## Stream straight from the body

`resp.Body` is an `io.Reader`, and `json.NewDecoder` reads JSON from any reader. So you
don't need to load the body into a `[]byte` first:

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type Issue struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	State string `json:"state"`
}

func main() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":1,"title":"Set up CI","state":"closed"},
			{"id":2,"title":"Write README","state":"open"}]`)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/issues")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("unexpected status:", resp.Status)
		return
	}

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		fmt.Println("decoding issues:", err)
		return
	}
	for _, iss := range issues {
		fmt.Printf("#%d [%s] %s\n", iss.ID, iss.State, iss.Title)
	}
}
```

Output:

```
#1 [closed] Set up CI
#2 [open] Write README
```

The order of the checks matters:

1. network error? return it;
2. `defer` the close;
3. bad status? return an error, *without* decoding;
4. decode into a pointer to your variable, and wrap the error with some context.

## ReadAll + Unmarshal works too

```go
body, err := io.ReadAll(resp.Body)
if err != nil {
	return nil, err
}
var issues []Issue
err = json.Unmarshal(body, &issues)
```

That's handy when you also want the raw bytes, for example to include a snippet of
the body in an error message. For plain decoding, the streaming `Decoder` is shorter.

## A Decoder gotcha

`Decoder.Decode` reads **one** JSON value and stops. If the body is
`[...] garbage`, `Decode` happily returns the array and never notices the junk after
it. `json.Unmarshal` would reject it. For API responses that rarely matters, but the
v2 package you'll meet next fixes it with `UnmarshalRead`, which insists on exactly
one value.

## Decode errors are useful

When the JSON doesn't match your types, the error says where:

```
decoding issues: json: cannot unmarshal string into Go struct field Issue.id of type int
```

And when the server sends something that isn't JSON at all, like an HTML error page
from a proxy:

```
decoding issues: invalid character '<' looking for beginning of value
```

If you see `'<'`, you almost certainly decoded HTML. Check the status and the
`Content-Type` header.

## Your turn

The Trackr API serves `GET /issues` as a JSON array. Each issue looks like this:

```json
{"id": 7, "title": "Dark mode", "state": "open", "story_points": 5, "labels": ["ui"]}
```

1. Add struct tags to `Issue` so every field is filled in, including `StoryPoints`
   from `story_points`.
2. Complete `listIssues` so it `GET`s `baseURL + "/issues"`, closes the body,
   returns an error for any status other than `200 OK`, and decodes the array.
   Return decode errors too (wrapping them with `fmt.Errorf("...: %w", err)` is nice).
