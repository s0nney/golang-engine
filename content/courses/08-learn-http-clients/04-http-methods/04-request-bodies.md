---
title: Sending a Request Body
quiz:
  - question: 'You POST JSON but forget `req.Header.Set("Content-Type", "application/json")`. What''s the most likely result?'
    options:
      - text: Go detects the JSON and sets the header for you
      - text: The server may not parse the body as JSON and answers with an error like `415 Unsupported Media Type` or `400 Bad Request`
        correct: true
      - text: The request isn't sent at all
    explanation: |
      `net/http` doesn't sniff request bodies. The header is how the server knows which
      parser to use, and strict APIs reject requests without it.
  - question: Why pass a `*bytes.Buffer` (or `*bytes.Reader` or `*strings.Reader`) as the body instead of some other `io.Reader`?
    options:
      - text: Other readers aren't allowed
      - text: For these types, `NewRequestWithContext` knows the length up front and can rewind the body if it has to be resent
        correct: true
      - text: They're the only readers that support JSON
    explanation: |
      For those three types, `NewRequestWithContext` sets `ContentLength` and
      `GetBody`. The length lets it send a `Content-Length` header, and `GetBody` lets
      the transport resend the body after a redirect or a dropped connection.
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

    // NewIssue is what trackr sends to create an issue.
    type NewIssue struct {
    	Title  string   `json:"title"`
    	Labels []string `json:"labels,omitempty"`
    }

    // Issue is what the API sends back.
    type Issue struct {
    	ID     int      `json:"id"`
    	Title  string   `json:"title"`
    	State  string   `json:"state"`
    	Labels []string `json:"labels"`
    }

    // createIssue POSTs in as JSON to baseURL + "/projects/" + project + "/issues"
    // and returns the created issue. The API answers 201 Created on success.
    func createIssue(ctx context.Context, baseURL, project string, in NewIssue) (Issue, error) {
    	// ?
    	return Issue{}, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		var in NewIssue
    		if err := json.UnmarshalRead(r.Body, &in); err != nil {
    			http.Error(w, err.Error(), http.StatusBadRequest)
    			return
    		}
    		w.Header().Set("Content-Type", "application/json")
    		w.WriteHeader(http.StatusCreated)
    		json.MarshalWrite(w, Issue{ID: 101, Title: in.Title, State: "open", Labels: in.Labels})
    	}))
    	defer srv.Close()

    	iss, err := createIssue(context.Background(), srv.URL, "apollo", NewIssue{Title: "Add dark mode", Labels: []string{"ui"}})
    	fmt.Printf("%+v %v\n", iss, err)
    }
  solution: |
    package main

    import (
    	"bytes"
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    type NewIssue struct {
    	Title  string   `json:"title"`
    	Labels []string `json:"labels,omitempty"`
    }

    type Issue struct {
    	ID     int      `json:"id"`
    	Title  string   `json:"title"`
    	State  string   `json:"state"`
    	Labels []string `json:"labels"`
    }

    func createIssue(ctx context.Context, baseURL, project string, in NewIssue) (Issue, error) {
    	var body bytes.Buffer
    	if err := json.MarshalWrite(&body, in); err != nil {
    		return Issue{}, fmt.Errorf("encoding issue: %w", err)
    	}

    	url := baseURL + "/projects/" + project + "/issues"
    	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
    	if err != nil {
    		return Issue{}, err
    	}
    	req.Header.Set("Content-Type", "application/json")

    	resp, err := http.DefaultClient.Do(req)
    	if err != nil {
    		return Issue{}, err
    	}
    	defer resp.Body.Close()

    	if resp.StatusCode != http.StatusCreated {
    		return Issue{}, fmt.Errorf("creating issue: unexpected status %s", resp.Status)
    	}

    	var created Issue
    	if err := json.UnmarshalRead(resp.Body, &created); err != nil {
    		return Issue{}, fmt.Errorf("decoding created issue: %w", err)
    	}
    	return created, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		var in NewIssue
    		if err := json.UnmarshalRead(r.Body, &in); err != nil {
    			http.Error(w, err.Error(), http.StatusBadRequest)
    			return
    		}
    		w.Header().Set("Content-Type", "application/json")
    		w.WriteHeader(http.StatusCreated)
    		json.MarshalWrite(w, Issue{ID: 101, Title: in.Title, State: "open", Labels: in.Labels})
    	}))
    	defer srv.Close()

    	iss, err := createIssue(context.Background(), srv.URL, "apollo", NewIssue{Title: "Add dark mode", Labels: []string{"ui"}})
    	fmt.Printf("%+v %v\n", iss, err)
    }
  tests: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"strings"
    	"testing"
    )

    type seen struct {
    	method, path, contentType, body string
    }

    func fakeAPI(status int, got *seen) *httptest.Server {
    	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		b, _ := io.ReadAll(r.Body)
    		*got = seen{r.Method, r.URL.Path, r.Header.Get("Content-Type"), string(b)}
    		var in NewIssue
    		if err := json.Unmarshal(b, &in); err != nil {
    			http.Error(w, `{"error":"bad json"}`, http.StatusBadRequest)
    			return
    		}
    		w.Header().Set("Content-Type", "application/json")
    		w.WriteHeader(status)
    		if status == http.StatusCreated {
    			json.MarshalWrite(w, Issue{ID: 57, Title: in.Title, State: "open", Labels: in.Labels})
    			return
    		}
    		io.WriteString(w, `{"error":"title is required"}`)
    	}))
    }

    func TestCreateIssue(t *testing.T) {
    	var got seen
    	srv := fakeAPI(http.StatusCreated, &got)
    	defer srv.Close()

    	iss, err := createIssue(context.Background(), srv.URL, "apollo", NewIssue{Title: "Fix login", Labels: []string{"bug", "p1"}})
    	if err != nil {
    		t.Fatalf("createIssue returned error %v, want nil", err)
    	}
    	if got.method != http.MethodPost {
    		t.Errorf("server saw method %q, want POST", got.method)
    	}
    	if got.path != "/projects/apollo/issues" {
    		t.Errorf("server saw path %q, want /projects/apollo/issues", got.path)
    	}
    	if !strings.HasPrefix(got.contentType, "application/json") {
    		t.Errorf("server saw Content-Type %q, want application/json", got.contentType)
    	}
    	var sent map[string]any
    	if err := json.Unmarshal([]byte(got.body), &sent); err != nil {
    		t.Fatalf("request body %q isn't valid JSON: %v", got.body, err)
    	}
    	if sent["title"] != "Fix login" {
    		t.Errorf("request body %s: want \"title\":\"Fix login\"", got.body)
    	}
    	if iss.ID != 57 || iss.Title != "Fix login" || iss.State != "open" || !slices.Equal(iss.Labels, []string{"bug", "p1"}) {
    		t.Errorf("createIssue returned %+v, want the issue decoded from the response (ID 57)", iss)
    	}
    }

    func TestCreateIssueNoLabels(t *testing.T) {
    	var got seen
    	srv := fakeAPI(http.StatusCreated, &got)
    	defer srv.Close()

    	if _, err := createIssue(context.Background(), srv.URL, "apollo", NewIssue{Title: "Bare"}); err != nil {
    		t.Fatalf("createIssue returned error %v", err)
    	}
    	if strings.Contains(got.body, "labels") {
    		t.Errorf("request body %s: an issue without labels shouldn't send a labels field (the tag has omitempty, so encode NewIssue as-is)", got.body)
    	}
    }

    func TestCreateIssueRejected(t *testing.T) {
    	for _, status := range []int{http.StatusOK, http.StatusUnprocessableEntity, http.StatusInternalServerError} {
    		var got seen
    		srv := fakeAPI(status, &got)
    		iss, err := createIssue(context.Background(), srv.URL, "apollo", NewIssue{Title: "x"})
    		srv.Close()
    		if err == nil {
    			t.Errorf("server replied %d: createIssue = (%+v, nil), want an error (only 201 Created is success)", status, iss)
    		}
    	}
    }

    func TestCreateIssueUsesContext(t *testing.T) {
    	var got seen
    	srv := fakeAPI(http.StatusCreated, &got)
    	defer srv.Close()

    	ctx, cancel := context.WithCancel(context.Background())
    	cancel()
    	if _, err := createIssue(ctx, srv.URL, "apollo", NewIssue{Title: "x"}); err == nil {
    		t.Error("createIssue with a cancelled context returned nil error; build the request with http.NewRequestWithContext(ctx, ...)")
    	}
    	if got.method != "" {
    		t.Error("the server received a request even though the context was already cancelled")
    	}
    }
---

`trackr new "Add dark mode" --label ui` has to send data *to* the server. That data
travels in the **request body**, and for Trackr, as for most APIs, it's JSON.

## Encode, then send

The body argument of `http.NewRequestWithContext` is an `io.Reader`. Encode your value
into a `bytes.Buffer` and pass a pointer to it:

```go
type NewIssue struct {
	Title  string   `json:"title"`
	Labels []string `json:"labels,omitempty"`
}

var body bytes.Buffer
if err := json.MarshalWrite(&body, NewIssue{Title: "Add dark mode", Labels: []string{"ui"}}); err != nil {
	return err
}

req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/projects/apollo/issues", &body)
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/json")
```

(`json` here is `encoding/json/v2`. `json.Marshal` plus `bytes.NewReader(data)` works
just as well.)

### Content-Type is your job

The server reads the `Content-Type` header to decide how to parse the body. `net/http`
won't guess it for a request, so **always** set it when you send JSON. Forgetting it
is the classic cause of a mysterious `415 Unsupported Media Type`.

### Content-Length comes free

When the body is a `*bytes.Buffer`, `*bytes.Reader` or `*strings.Reader`,
`NewRequestWithContext` knows the size, so it sets `req.ContentLength` and sends a
`Content-Length` header. It also sets `req.GetBody`, a function that returns a fresh
copy of the body. The client needs that to resend the body after a `307` or `308`
redirect, or to retry on a new connection. With any other reader (say, a file or a
pipe) the length is unknown, so the body is streamed in chunks instead.

## Read the answer

A create usually answers `201 Created`, with the new resource in the body and often a
`Location` header pointing at it:

```go
resp, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusCreated {
	return fmt.Errorf("creating issue: unexpected status %s", resp.Status)
}

var created Issue
if err := json.UnmarshalRead(resp.Body, &created); err != nil {
	return err
}
fmt.Println("created issue", created.ID, "at", resp.Header.Get("Location"))
```

Checking for exactly `201` (rather than "any 2xx") is fine when the API documents it.
Chapter 5 builds a more general status check.

## Who closes the request body?

You don't. The client closes the request body (if it's an `io.Closer`) once it's done
with it, even when `Do` returns an error. **Response** bodies are still yours to close.

## The shortcut

For a quick POST there's `http.Post(url, contentType, body)`:

```go
resp, err := http.Post(baseURL+"/projects/apollo/issues", "application/json", &body)
```

It sets `Content-Type` for you, but it has no context and no way to add other
headers, so `trackr` builds its requests by hand.

## Your turn

Complete `createIssue` so it:

1. encodes `in` as JSON (use `json.MarshalWrite` into a `bytes.Buffer`; the starter
   already imports `encoding/json/v2`);
2. builds a `POST` to `baseURL + "/projects/" + project + "/issues"` with
   `http.NewRequestWithContext`, using the `ctx` it's given;
3. sets `Content-Type: application/json`;
4. sends it with `http.DefaultClient.Do` and closes the response body;
5. returns an error unless the status is `201 Created`;
6. decodes and returns the created `Issue`.

Encode `in` exactly as it is. Its `omitempty` tag already leaves out an empty label
list, and the tests check that.
