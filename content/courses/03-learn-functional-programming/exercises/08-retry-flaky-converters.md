---
title: Retry Flaky Converters
difficulty: medium
after: decorators-and-middleware
hints:
  - '`Retry(...)` returns a `Middleware`, which is itself a function that takes `next Converter` and returns a new `Converter`. That''s two nested `return func` literals.'
  - 'Inside the innermost function, loop up to `attempts` times (at least once). Call `next(doc)`; on success return straight away; if the error isn''t `retryable`, return it straight away; otherwise remember it and try again.'
  - 'Wrap the final error with `%w` so `errors.Is` still finds the original: `fmt.Errorf("gave up after %d attempts: %w", n, lastErr)`. Keep the attempt counter **inside** the innermost function, so every document gets a fresh set of attempts.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    type Converter func(doc string) (string, error)

    type Middleware func(Converter) Converter

    var ErrBusy = errors.New("renderer busy")

    func Retry(attempts int, retryable func(error) bool) Middleware {
    	return func(next Converter) Converter {
    		return next
    	}
    }

    func main() {
    	calls := 0
    	flaky := func(doc string) (string, error) {
    		calls++
    		if calls < 3 {
    			return "", ErrBusy
    		}
    		return "<p>" + doc + "</p>", nil
    	}
    	isBusy := func(err error) bool { return errors.Is(err, ErrBusy) }

    	convert := Retry(5, isBusy)(flaky)
    	out, err := convert("hello")
    	fmt.Printf("%q %v after %d calls\n", out, err, calls) // want: "<p>hello</p>" <nil> after 3 calls
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    type Converter func(doc string) (string, error)

    type Middleware func(Converter) Converter

    var ErrBusy = errors.New("renderer busy")

    func Retry(attempts int, retryable func(error) bool) Middleware {
    	attempts = max(attempts, 1)
    	return func(next Converter) Converter {
    		return func(doc string) (string, error) {
    			var lastErr error
    			for range attempts {
    				out, err := next(doc)
    				if err == nil {
    					return out, nil
    				}
    				if !retryable(err) {
    					return "", err
    				}
    				lastErr = err
    			}
    			return "", fmt.Errorf("gave up after %d attempts: %w", attempts, lastErr)
    		}
    	}
    }

    func main() {
    	calls := 0
    	flaky := func(doc string) (string, error) {
    		calls++
    		if calls < 3 {
    			return "", ErrBusy
    		}
    		return "<p>" + doc + "</p>", nil
    	}
    	isBusy := func(err error) bool { return errors.Is(err, ErrBusy) }

    	convert := Retry(5, isBusy)(flaky)
    	out, err := convert("hello")
    	fmt.Printf("%q %v after %d calls\n", out, err, calls)
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"testing"
    )

    var errBroken = errors.New("broken markup")

    func isBusy(err error) bool { return errors.Is(err, ErrBusy) }

    // failing returns a converter that fails with err the first n times it is
    // called, then succeeds. *calls counts every call.
    func failing(n int, err error, calls *int) Converter {
    	return func(doc string) (string, error) {
    		*calls++
    		if *calls <= n {
    			return "", err
    		}
    		return "<p>" + doc + "</p>", nil
    	}
    }

    func TestRetrySucceeds(t *testing.T) {
    	tests := []struct {
    		name      string
    		failures  int
    		attempts  int
    		wantCalls int
    	}{
    		{"works first time", 0, 3, 1},
    		{"one busy", 1, 3, 2},
    		{"busy until the last attempt", 2, 3, 3},
    		{"many attempts left over", 1, 10, 2},
    	}
    	for _, tt := range tests {
    		calls := 0
    		conv := Retry(tt.attempts, isBusy)(failing(tt.failures, ErrBusy, &calls))
    		out, err := conv("hi")
    		if out != "<p>hi</p>" || err != nil {
    			t.Errorf("%s: Retry(%d) around a converter busy %d time(s) returned %q, %v, want %q, nil", tt.name, tt.attempts, tt.failures, out, err, "<p>hi</p>")
    		}
    		if calls != tt.wantCalls {
    			t.Errorf("%s: the converter was called %d times, want %d (stop retrying once it succeeds)", tt.name, calls, tt.wantCalls)
    		}
    	}
    }

    func TestRetryGivesUp(t *testing.T) {
    	calls := 0
    	conv := Retry(3, isBusy)(failing(100, ErrBusy, &calls))
    	out, err := conv("hi")
    	if calls != 3 {
    		t.Errorf("Retry(3) around an always-busy converter called it %d times, want 3", calls)
    	}
    	if out != "" || err == nil {
    		t.Fatalf("Retry(3) around an always-busy converter returned %q, %v, want \"\" and an error", out, err)
    	}
    	if !errors.Is(err, ErrBusy) {
    		t.Errorf("errors.Is(err, ErrBusy) = false for %q: wrap the last error with %%w", err)
    	}
    	if want := "gave up after 3 attempts: renderer busy"; err.Error() != want {
    		t.Errorf("error = %q, want %q", err, want)
    	}
    }

    func TestRetryDoesNotRetryOtherErrors(t *testing.T) {
    	calls := 0
    	conv := Retry(5, isBusy)(failing(100, errBroken, &calls))
    	out, err := conv("hi")
    	if calls != 1 {
    		t.Errorf("a converter failing with a non-retryable error was called %d times, want 1", calls)
    	}
    	if out != "" || err != errBroken {
    		t.Errorf("Retry returned %q, %v, want \"\" and the converter's own error %q unchanged", out, err, errBroken)
    	}
    }

    func TestRetryAtLeastOnce(t *testing.T) {
    	for _, attempts := range []int{0, -3, 1} {
    		calls := 0
    		conv := Retry(attempts, isBusy)(failing(0, ErrBusy, &calls))
    		if out, err := conv("x"); out != "<p>x</p>" || err != nil || calls != 1 {
    			t.Errorf("Retry(%d) around a working converter returned %q, %v with %d call(s), want %q, nil with 1 call", attempts, out, err, calls, "<p>x</p>")
    		}
    	}
    }

    func TestRetryFreshAttemptsPerDocument(t *testing.T) {
    	// Busy on every odd-numbered call: each document needs 2 calls.
    	calls := 0
    	alternating := func(doc string) (string, error) {
    		calls++
    		if calls%2 == 1 {
    			return "", ErrBusy
    		}
    		return doc, nil
    	}
    	conv := Retry(2, isBusy)(alternating)
    	for i := range 4 {
    		doc := fmt.Sprint("doc", i)
    		if out, err := conv(doc); out != doc || err != nil {
    			t.Fatalf("document #%d through Retry(2): got %q, %v, want %q, nil: each call of the converter needs its own attempt count", i+1, out, err, doc)
    		}
    	}
    }

    func TestRetryPassesDocThrough(t *testing.T) {
    	var seen []string
    	record := func(doc string) (string, error) {
    		seen = append(seen, doc)
    		if len(seen) < 2 {
    			return "", ErrBusy
    		}
    		return doc + "!", nil
    	}
    	out, _ := Retry(3, isBusy)(record)("same")
    	if fmt.Sprint(seen) != "[same same]" || out != "same!" {
    		t.Errorf("the converter saw %q and Retry returned %q, want [same same] and %q", seen, out, "same!")
    	}
    }
---

Doc2Doc sends some conversions to a remote renderer that sometimes answers
`ErrBusy`. Trying again a moment later usually works. Other errors, like
broken markup, will fail every time, so retrying them is pointless.

Write `Retry(attempts, retryable)`. It returns a **middleware** that wraps a
`Converter` in retry logic. The wrapped converter:

1. calls `next(doc)` and returns its result straight away if it succeeds;
2. returns `""` and the error **unchanged**, without retrying, if
   `retryable(err)` is false;
3. otherwise tries again, calling `next` at most `attempts` times in total;
4. if every attempt fails, returns `""` and the error
   `gave up after N attempts: <last error>`, wrapped with `%w` so that
   `errors.Is(err, ErrBusy)` still works.

`attempts` below 1 counts as 1. Every call of the wrapped converter gets a
fresh set of attempts.

## Example

```go
calls := 0
flaky := func(doc string) (string, error) {
	calls++
	if calls < 3 {
		return "", ErrBusy
	}
	return "<p>" + doc + "</p>", nil
}
isBusy := func(err error) bool { return errors.Is(err, ErrBusy) }

convert := Retry(5, isBusy)(flaky)
convert("hello") // "<p>hello</p>", nil after 3 calls of flaky

_, err := Retry(2, isBusy)(alwaysBusy)("hi")
// err.Error() == "gave up after 2 attempts: renderer busy"
```

## Constraints

- No sleeping or timers: retry immediately.
- `Retry` must not call `next` until the wrapped converter is called.
