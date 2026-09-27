---
title: PUT, PATCH and DELETE
quiz:
  - question: |
      The API answers a DELETE with `204 No Content`. What happens here?

      ```go
      var result Issue
      err := json.UnmarshalRead(resp.Body, &result)
      ```
    options:
      - text: '`result` is the zero `Issue` and `err` is nil'
      - text: '`err` is non-nil, because an empty body isn''t valid JSON'
        correct: true
      - text: It blocks forever waiting for a body
    explanation: |
      A 204 has no body, and zero bytes isn't a JSON value, so decoding fails with an
      unexpected-EOF error. Check for 204 (or skip decoding for DELETE) instead.
  - question: |
      Why does `IssuePatch` use pointer fields?

      ```go
      type IssuePatch struct {
          Title *string `json:"title,omitzero"`
          State *string `json:"state,omitzero"`
      }
      ```
    options:
      - text: Pointers make the JSON smaller
      - text: A nil pointer means "don't change this field" and is left out, while a non-nil pointer is sent even if it points at `""`
        correct: true
      - text: '`encoding/json/v2` can''t encode plain strings'
    explanation: |
      A PATCH has to tell "not set" apart from "set to the empty value". With plain
      `string` and `omitzero`, clearing a title to `""` would be indistinguishable from
      not touching it.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    	State string `json:"state"`
    }

    // closeIssue sends PATCH baseURL/issues/<id> with the JSON body
    // {"state":"closed"} and returns the updated issue from the 200 OK response.
    func closeIssue(ctx context.Context, baseURL string, id int) (Issue, error) {
    	// ?
    	return Issue{}, nil
    }

    // deleteIssue sends DELETE baseURL/issues/<id>. Both 204 No Content and
    // 404 Not Found (it's already gone) count as success.
    func deleteIssue(ctx context.Context, baseURL string, id int) error {
    	// ?
    	return nil
    }

    func main() {
    	issues := map[string]*Issue{"42": {ID: 42, Title: "Login broken", State: "open"}}
    	mux := http.NewServeMux()
    	mux.HandleFunc("PATCH /issues/{id}", func(w http.ResponseWriter, r *http.Request) {
    		iss, ok := issues[r.PathValue("id")]
    		if !ok {
    			http.NotFound(w, r)
    			return
    		}
    		var patch struct {
    			State *string `json:"state"`
    		}
    		if err := json.UnmarshalRead(r.Body, &patch); err != nil {
    			http.Error(w, err.Error(), http.StatusBadRequest)
    			return
    		}
    		if patch.State != nil {
    			iss.State = *patch.State
    		}
    		json.MarshalWrite(w, iss)
    	})
    	mux.HandleFunc("DELETE /issues/{id}", func(w http.ResponseWriter, r *http.Request) {
    		if _, ok := issues[r.PathValue("id")]; !ok {
    			http.NotFound(w, r)
    			return
    		}
    		delete(issues, r.PathValue("id"))
    		w.WriteHeader(http.StatusNoContent)
    	})
    	srv := httptest.NewServer(mux)
    	defer srv.Close()
    	ctx := context.Background()

    	iss, err := closeIssue(ctx, srv.URL, 42)
    	fmt.Printf("close:    %+v %v\n", iss, err)
    	fmt.Println("delete 1:", deleteIssue(ctx, srv.URL, 42))
    	fmt.Println("delete 2:", deleteIssue(ctx, srv.URL, 42))
    }
  solution: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    )

    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    	State string `json:"state"`
    }

    func closeIssue(ctx context.Context, baseURL string, id int) (Issue, error) {
    	url := fmt.Sprintf("%s/issues/%d", baseURL, id)
    	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, strings.NewReader(`{"state":"closed"}`))
    	if err != nil {
    		return Issue{}, err
    	}
    	req.Header.Set("Content-Type", "application/json")

    	resp, err := http.DefaultClient.Do(req)
    	if err != nil {
    		return Issue{}, err
    	}
    	defer resp.Body.Close()

    	if resp.StatusCode != http.StatusOK {
    		return Issue{}, fmt.Errorf("closing issue %d: unexpected status %s", id, resp.Status)
    	}
    	var updated Issue
    	if err := json.UnmarshalRead(resp.Body, &updated); err != nil {
    		return Issue{}, fmt.Errorf("decoding issue %d: %w", id, err)
    	}
    	return updated, nil
    }

    func deleteIssue(ctx context.Context, baseURL string, id int) error {
    	url := fmt.Sprintf("%s/issues/%d", baseURL, id)
    	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
    	if err != nil {
    		return err
    	}
    	resp, err := http.DefaultClient.Do(req)
    	if err != nil {
    		return err
    	}
    	defer resp.Body.Close()

    	switch resp.StatusCode {
    	case http.StatusNoContent, http.StatusNotFound:
    		return nil
    	}
    	return fmt.Errorf("deleting issue %d: unexpected status %s", id, resp.Status)
    }

    func main() {
    	issues := map[string]*Issue{"42": {ID: 42, Title: "Login broken", State: "open"}}
    	mux := http.NewServeMux()
    	mux.HandleFunc("PATCH /issues/{id}", func(w http.ResponseWriter, r *http.Request) {
    		iss, ok := issues[r.PathValue("id")]
    		if !ok {
    			http.NotFound(w, r)
    			return
    		}
    		var patch struct {
    			State *string `json:"state"`
    		}
    		if err := json.UnmarshalRead(r.Body, &patch); err != nil {
    			http.Error(w, err.Error(), http.StatusBadRequest)
    			return
    		}
    		if patch.State != nil {
    			iss.State = *patch.State
    		}
    		json.MarshalWrite(w, iss)
    	})
    	mux.HandleFunc("DELETE /issues/{id}", func(w http.ResponseWriter, r *http.Request) {
    		if _, ok := issues[r.PathValue("id")]; !ok {
    			http.NotFound(w, r)
    			return
    		}
    		delete(issues, r.PathValue("id"))
    		w.WriteHeader(http.StatusNoContent)
    	})
    	srv := httptest.NewServer(mux)
    	defer srv.Close()
    	ctx := context.Background()

    	iss, err := closeIssue(ctx, srv.URL, 42)
    	fmt.Printf("close:    %+v %v\n", iss, err)
    	fmt.Println("delete 1:", deleteIssue(ctx, srv.URL, 42))
    	fmt.Println("delete 2:", deleteIssue(ctx, srv.URL, 42))
    }
  tests: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing"
    )

    type request struct {
    	method, path, contentType, body string
    }

    // fake records every request and answers with status and body.
    func fake(status int, body string, got *[]request) *httptest.Server {
    	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		b, _ := io.ReadAll(r.Body)
    		*got = append(*got, request{r.Method, r.URL.Path, r.Header.Get("Content-Type"), string(b)})
    		w.Header().Set("Content-Type", "application/json")
    		w.WriteHeader(status)
    		io.WriteString(w, body)
    	}))
    }

    func TestCloseIssue(t *testing.T) {
    	var got []request
    	srv := fake(http.StatusOK, `{"id": 7, "title": "Dark mode", "state": "closed"}`, &got)
    	defer srv.Close()

    	iss, err := closeIssue(context.Background(), srv.URL, 7)
    	if err != nil {
    		t.Fatalf("closeIssue returned error %v, want nil", err)
    	}
    	if len(got) != 1 {
    		t.Fatalf("server saw %d requests, want 1", len(got))
    	}
    	r := got[0]
    	if r.method != http.MethodPatch || r.path != "/issues/7" {
    		t.Errorf("server saw %s %s, want PATCH /issues/7", r.method, r.path)
    	}
    	if !strings.HasPrefix(r.contentType, "application/json") {
    		t.Errorf("server saw Content-Type %q, want application/json", r.contentType)
    	}
    	var sent map[string]any
    	if err := json.Unmarshal([]byte(r.body), &sent); err != nil {
    		t.Fatalf("request body %q isn't valid JSON: %v", r.body, err)
    	}
    	if len(sent) != 1 || sent["state"] != "closed" {
    		t.Errorf("request body %s, want exactly {\"state\":\"closed\"} (a PATCH sends only what changes)", r.body)
    	}
    	if iss != (Issue{ID: 7, Title: "Dark mode", State: "closed"}) {
    		t.Errorf("closeIssue returned %+v, want the issue decoded from the response", iss)
    	}
    }

    func TestCloseIssueErrors(t *testing.T) {
    	for _, status := range []int{http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError} {
    		var got []request
    		srv := fake(status, `{"error": "nope"}`, &got)
    		iss, err := closeIssue(context.Background(), srv.URL, 7)
    		srv.Close()
    		if err == nil {
    			t.Errorf("server replied %d: closeIssue = (%+v, nil), want an error", status, iss)
    		}
    	}
    }

    func TestDeleteIssue(t *testing.T) {
    	for _, status := range []int{http.StatusNoContent, http.StatusNotFound} {
    		var got []request
    		srv := fake(status, "", &got)
    		err := deleteIssue(context.Background(), srv.URL, 13)
    		srv.Close()
    		if err != nil {
    			t.Errorf("server replied %d: deleteIssue returned %v, want nil (204 and 404 both mean it's gone)", status, err)
    		}
    		if len(got) != 1 || got[0].method != http.MethodDelete || got[0].path != "/issues/13" {
    			t.Errorf("server saw %+v, want one DELETE /issues/13", got)
    			continue
    		}
    		if got[0].body != "" {
    			t.Errorf("DELETE sent the body %q, want no body", got[0].body)
    		}
    	}
    }

    func TestDeleteIssueErrors(t *testing.T) {
    	for _, status := range []int{http.StatusForbidden, http.StatusConflict, http.StatusInternalServerError} {
    		var got []request
    		srv := fake(status, `{"error": "nope"}`, &got)
    		err := deleteIssue(context.Background(), srv.URL, 13)
    		srv.Close()
    		if err == nil {
    			t.Errorf("server replied %d: deleteIssue returned nil, want an error", status)
    		}
    	}
    }

    func TestUsesContext(t *testing.T) {
    	var got []request
    	srv := fake(http.StatusNoContent, "", &got)
    	defer srv.Close()

    	ctx, cancel := context.WithCancel(context.Background())
    	cancel()
    	if _, err := closeIssue(ctx, srv.URL, 1); err == nil {
    		t.Error("closeIssue with a cancelled context returned nil error; build the request with http.NewRequestWithContext(ctx, ...)")
    	}
    	if err := deleteIssue(ctx, srv.URL, 1); err == nil {
    		t.Error("deleteIssue with a cancelled context returned nil error; build the request with http.NewRequestWithContext(ctx, ...)")
    	}
    	if len(got) != 0 {
    		t.Errorf("server received %d requests even though the context was already cancelled", len(got))
    	}
    }
---

With `NewRequestWithContext` and a JSON body you can already send any method. This
lesson collects the patterns `trackr` uses for updates and deletes, plus the one
status code that trips people up.

## PATCH: send only the changes

A PATCH body should contain exactly the fields you're changing. Pointer fields plus
`omitzero` make that easy: nil means "leave it alone", so it's omitted, and anything
else is sent.

```go
type IssuePatch struct {
	Title *string `json:"title,omitzero"`
	State *string `json:"state,omitzero"`
}

patch := IssuePatch{State: new("closed")} // new(expr), Go 1.26+
// encodes as {"state":"closed"}
```

## PUT: send the whole thing

A PUT replaces the resource, so send the complete issue. The usual flow is
read-modify-write: GET the issue, change the fields, PUT it back. (If someone else
edits the issue in between, you'll overwrite their change. That's one reason APIs
prefer PATCH for small edits.)

## DELETE: no body, maybe no response body

A DELETE usually sends no body (`nil`) and gets back `204 No Content`. **204 means
there's no body at all**, so don't try to decode one. Just check the status and close.

## Try it

This fake API supports PATCH and DELETE on `/issues/{id}`. It uses Go's method-aware
`ServeMux` patterns (`"PATCH /issues/{id}"`) and `r.PathValue`. You'll build more
servers like it in chapter 9.

```go
package main

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type Issue struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	State string `json:"state"`
}

type IssuePatch struct {
	Title *string `json:"title,omitzero"`
	State *string `json:"state,omitzero"`
}

func main() {
	issues := map[string]*Issue{"42": {ID: 42, Title: "Login broken", State: "open"}}
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /issues/{id}", func(w http.ResponseWriter, r *http.Request) {
		iss, ok := issues[r.PathValue("id")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		var p IssuePatch
		if err := json.UnmarshalRead(r.Body, &p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if p.Title != nil {
			iss.Title = *p.Title
		}
		if p.State != nil {
			iss.State = *p.State
		}
		json.MarshalWrite(w, iss)
	})
	mux.HandleFunc("DELETE /issues/{id}", func(w http.ResponseWriter, r *http.Request) {
		delete(issues, r.PathValue("id"))
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx := context.Background()

	// PATCH: send only what changes.
	var body bytes.Buffer
	json.MarshalWrite(&body, IssuePatch{State: new("closed")})
	fmt.Println("PATCH body:", body.String())
	req, _ := http.NewRequestWithContext(ctx, http.MethodPatch, srv.URL+"/issues/42", &body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	var updated Issue
	json.UnmarshalRead(resp.Body, &updated)
	resp.Body.Close()
	fmt.Printf("%s -> %+v\n", resp.Status, updated)

	// DELETE: no request body, and 204 means no response body either.
	for range 2 {
		req, _ = http.NewRequestWithContext(ctx, http.MethodDelete, srv.URL+"/issues/42", nil)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		resp.Body.Close()
		fmt.Println("DELETE:", resp.Status)
	}
}
```

Output:

```
PATCH body: {"state":"closed"}
200 OK -> {ID:42 Title:Login broken State:closed}
DELETE: 204 No Content
DELETE: 204 No Content
```

Notice that the second DELETE also returned 204. This API treats "already gone" as
success, which makes DELETE truly idempotent from the client's point of view. Other
APIs return `404` the second time. If `trackr delete` should be repeat-proof, it can
treat a 404 on DELETE as success too.

## A helper for "any method with JSON"

`trackr` ends up writing the same few lines for every call, so it's worth one small
helper:

```go
// newJSONRequest builds a request whose body is v encoded as JSON.
// Pass v == nil for no body.
func newJSONRequest(ctx context.Context, method, url string, v any) (*http.Request, error) {
	var body io.Reader
	if v != nil {
		var buf bytes.Buffer
		if err := json.MarshalWrite(&buf, v); err != nil {
			return nil, err
		}
		body = &buf
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if v != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}
```

Look closely at `var body io.Reader`. It's declared as the **interface** type and only
assigned when there's a body. If you wrote `var buf *bytes.Buffer` and passed `buf`
while it's nil, you'd hand `NewRequestWithContext` a non-nil interface holding a nil
pointer, the classic Go nil-interface gotcha. `NewRequestWithContext` would see a
`*bytes.Buffer`, ask it for its length, and panic with a nil pointer dereference.
Declaring the interface type avoids it.

## Your turn

Give `trackr` its `close` and `delete` commands. Complete:

1. `closeIssue(ctx, baseURL, id)`: send `PATCH baseURL/issues/<id>` whose body is
   exactly `{"state":"closed"}` (a `strings.NewReader` is fine for a fixed body), with
   `Content-Type: application/json`. Return an error unless the status is `200 OK`,
   and otherwise decode and return the updated `Issue`.
2. `deleteIssue(ctx, baseURL, id)`: send `DELETE baseURL/issues/<id>` with no body.
   Treat `204 No Content` **and** `404 Not Found` as success, so that deleting twice is
   harmless. Any other status is an error.

Build both requests with `http.NewRequestWithContext` and the `ctx` you're given, send
them with `http.DefaultClient.Do`, and close every response body. Run the program
first: the second delete should print `<nil>` too.

In the next chapter you'll set headers like `Accept` and `Authorization` properly and
learn what all those status codes mean.
