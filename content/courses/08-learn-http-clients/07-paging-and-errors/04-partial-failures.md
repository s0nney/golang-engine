---
title: Partial Failures
quiz:
  - question: '`trackr close 11 12 13 20 21` fails on #13 (404). What''s the most useful behaviour?'
    options:
      - text: Stop at #13 and leave 20 and 21 untouched, without saying which ones were closed
      - text: Try every issue, then report which ones failed and why
        correct: true
      - text: Ignore the failure and exit successfully
    explanation: |
      The five closes are independent, so one bad ID shouldn't block the others. The
      user needs to know exactly what happened: which succeeded and which failed.
      `errors.Join` bundles the failures into one error.
  - question: |
      `err := errors.Join(e1, e2)`, where `e2` wraps an `*APIError`. What does
      `errors.AsType[*APIError](err)` return?
    options:
      - text: '`nil, false`, because `Join` hides the errors inside'
      - text: The `*APIError` from `e2` and `true`
        correct: true
      - text: It panics
    explanation: |
      A joined error has an `Unwrap() []error` method, and `errors.Is`, `errors.As` and
      `errors.AsType` search every branch. The first match found wins.
---

Talking to an API means many requests, and any one of them can fail. What should
`trackr` do when page 7 of 10 fails, or 2 of 5 updates? There's no single right
answer, but there are good patterns.

## Reads: page 7 of 10 failed

When `trackr list` is streaming pages, you have three reasonable choices:

1. **Fail the whole thing.** Print what you have, then the error, and exit non-zero.
   Simple and honest. Right for most interactive commands.
2. **Retry the page.** The failure was probably a 503 or a timeout. Your retry logic
   from chapter 6 applies to page requests like any other GET, and with cursor
   pagination retrying is exact: you resend the same cursor and get the same page.
3. **Resume later.** For a long export, save the last good cursor. If the run dies at
   page 7,000, the next run can start from there instead of from scratch:

```go
for iss, err := range allIssues(ctx, client, baseURL) {
	if err != nil {
		return fmt.Errorf("export stopped after issue #%d (resume with --after %d): %w",
			lastID, lastID, err)
	}
	lastID = iss.ID
	// write iss to the export file
}
```

What you should **not** do is skip the bad page and carry on silently. The output
would look complete while missing items, which is worse than an error.

## Writes: 2 of 5 updates failed

Batch commands like `trackr close 11 12 13 20 21` send independent requests. One bad
ID shouldn't stop the rest, and the user must learn *exactly* what happened.
`errors.Join` (Go 1.20) collects several errors into one:

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string { return fmt.Sprintf("trackr API: %d %s", e.StatusCode, e.Message) }

// closeIssue PATCHes one issue's state to closed.
func closeIssue(ctx context.Context, client *http.Client, baseURL string, id int) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		fmt.Sprintf("%s/issues/%d", baseURL, id), strings.NewReader(`{"state":"closed"}`))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return &APIError{resp.StatusCode, http.StatusText(resp.StatusCode)}
	}
	return nil
}

// closeAll tries every issue and reports all the failures together.
func closeAll(ctx context.Context, client *http.Client, baseURL string, ids []int) (closed []int, err error) {
	var errs []error
	for _, id := range ids {
		if err := closeIssue(ctx, client, baseURL, id); err != nil {
			errs = append(errs, fmt.Errorf("issue #%d: %w", id, err))
			continue
		}
		closed = append(closed, id)
	}
	return closed, errors.Join(errs...)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /issues/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.PathValue("id") {
		case "13":
			w.WriteHeader(http.StatusNotFound)
		case "21":
			w.WriteHeader(http.StatusForbidden)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	closed, err := closeAll(context.Background(), srv.Client(), srv.URL, []int{11, 12, 13, 20, 21})
	fmt.Println("closed:", closed)
	if err != nil {
		fmt.Printf("some issues failed:\n%v\n", err)
	}
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		fmt.Println("first API error status:", apiErr.StatusCode)
	}
}
```

Output:

```
closed: [11 12 20]
some issues failed:
issue #13: trackr API: 404 Not Found
issue #21: trackr API: 403 Forbidden
first API error status: 404
```

Things to notice:

- `errors.Join` returns **nil** when every error is nil (or there are none), so the
  success path needs no special case.
- Its message puts each error on its own line.
- `errors.Is` and `errors.AsType` look inside every joined error.
- Wrapping each error with the issue ID (`"issue #%d: %w"`) is what makes the report
  useful.

## Partial success is still success... partly

Return **both** the results and the error, as `closeAll` does. The caller can print
"closed 3 of 5" and still exit with a non-zero status, so scripts that call `trackr`
notice the failure:

```go
closed, err := closeAll(ctx, client, baseURL, ids)
fmt.Printf("closed %d of %d issues\n", len(closed), len(ids))
if err != nil {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
```

This bends the usual Go rule that results are meaningless when `err != nil`, so
**document it** in the function's comment when you do it.

## When the network is the problem

If the first error is a network error (DNS failure, connection refused), the next 500
requests will almost certainly fail the same way. A smart batch stops early after a
transport error rather than printing 500 identical messages:

```go
if _, isAPI := errors.AsType[*APIError](err); !isAPI {
	errs = append(errs, err)
	break // not the server saying no: we can't reach it at all
}
```

## Concurrency

Large batches often run requests concurrently (with a concurrency limit, as in
chapter 6). Collect the errors behind a mutex, or send them over a channel, and join
them at the end. The reporting stays the same.
