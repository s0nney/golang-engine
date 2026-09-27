---
title: Wrapping Errors
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import (
      	"errors"
      	"fmt"
      )

      var ErrOptedOut = errors.New("recipient opted out")

      func main() {
      	a := fmt.Errorf("send to bob: %w", ErrOptedOut)
      	b := fmt.Errorf("send to bob: %v", ErrOptedOut)
      	fmt.Println(errors.Is(a, ErrOptedOut), errors.Is(b, ErrOptedOut))
      }
      ```
    options:
      - text: '`true true`'
      - text: '`false false`'
      - text: '`true false`'
        correct: true
      - text: '`false true`'
    explanation: |
      Both errors have the same message, but only `%w` *wraps* the original
      error so `errors.Is` can find it. `%v` just copies the text, so the
      link to `ErrOptedOut` is lost.
  - question: Why should you use `errors.Is(err, ErrOptedOut)` instead of `err == ErrOptedOut`?
    options:
      - text: '`==` doesn''t compile for errors'
      - text: '`errors.Is` also finds `ErrOptedOut` when it has been wrapped inside other errors'
        correct: true
      - text: '`errors.Is` is faster'
    explanation: |
      Once an error has been wrapped with `%w`, it's no longer equal to the
      original. `errors.Is` unwraps the chain, checking each error along the
      way.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var (
    	ErrEmptyBody    = errors.New("empty message body")
    	ErrInvalidPhone = errors.New("invalid phone number")
    )

    // validate checks a message before Textio sends it.
    func validate(to, body string) error {
    	// ?
    	return nil
    }

    func main() {
    	fmt.Println(validate("+1-555-0100", "hi")) // want <nil>
    	fmt.Println(validate("555-0100", "hi"))    // want to 555-0100: invalid phone number
    	fmt.Println(validate("+1-555-0100", ""))   // want to +1-555-0100: empty message body
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strings"
    )

    var (
    	ErrEmptyBody    = errors.New("empty message body")
    	ErrInvalidPhone = errors.New("invalid phone number")
    )

    func validate(to, body string) error {
    	if !strings.HasPrefix(to, "+") {
    		return fmt.Errorf("to %s: %w", to, ErrInvalidPhone)
    	}
    	if body == "" {
    		return fmt.Errorf("to %s: %w", to, ErrEmptyBody)
    	}
    	return nil
    }

    func main() {
    	fmt.Println(validate("+1-555-0100", "hi"))
    	fmt.Println(validate("555-0100", "hi"))
    	fmt.Println(validate("+1-555-0100", ""))
    }
  tests: |
    package main

    import (
    	"errors"
    	"testing"
    )

    func TestValidate(t *testing.T) {
    	tests := []struct {
    		to, body string
    		wantErr  error
    		wantMsg  string
    	}{
    		{"+1-555-0100", "hi", nil, ""},
    		{"555-0100", "hi", ErrInvalidPhone, "to 555-0100: invalid phone number"},
    		{"+1-555-0100", "", ErrEmptyBody, "to +1-555-0100: empty message body"},
    		{"", "", ErrInvalidPhone, "to : invalid phone number"},
    	}
    	for _, tt := range tests {
    		err := validate(tt.to, tt.body)
    		if tt.wantErr == nil {
    			if err != nil {
    				t.Errorf("validate(%q, %q) = %v, want nil", tt.to, tt.body, err)
    			}
    			continue
    		}
    		if !errors.Is(err, tt.wantErr) {
    			t.Errorf("validate(%q, %q) = %v, want an error wrapping %q", tt.to, tt.body, err, tt.wantErr)
    			continue
    		}
    		if err.Error() != tt.wantMsg {
    			t.Errorf("validate(%q, %q) error message = %q, want %q", tt.to, tt.body, err.Error(), tt.wantMsg)
    		}
    	}
    }
---

An error message like `connection refused` isn't very helpful by the time it reaches the top of your program. Refused *where*? While doing *what*? Good Go code adds context as errors travel up, by **wrapping** them.

## Adding context with `fmt.Errorf`

`fmt.Errorf` works like `fmt.Sprintf`, but returns an error. The special verb **`%w`** wraps another error inside the new one:

```go
package main

import (
	"errors"
	"fmt"
)

func dial(carrier string) error {
	return errors.New("connection refused")
}

func send(to string) error {
	if err := dial("Alpha Mobile"); err != nil {
		return fmt.Errorf("send to %s: %w", to, err)
	}
	return nil
}

func main() {
	if err := send("+1-555-0100"); err != nil {
		fmt.Println(err)
	}
}
```

```text
send to +1-555-0100: connection refused
```

Each layer adds its own context, separated by `: `. Read it like a trail of breadcrumbs: we were sending to this number, and that failed because the connection was refused. This is why error messages are lowercase with no trailing punctuation: they get joined into sentences like this.

## Sentinel errors

Sometimes a caller needs to react to a *specific* error. For example, if the recipient has opted out, Textio should stop retrying, but if the carrier is down, it should try again later.

A common approach is a **sentinel error**: a package-level error variable that callers can compare against. By convention its name starts with `Err`:

```go
var ErrOptedOut = errors.New("recipient opted out")
```

## Checking with `errors.Is`

Once an error has been wrapped, it's a *different* value from the original, so `==` no longer matches. `errors.Is` solves this: it **unwraps** the error chain, layer by layer, looking for a match:

```go
package main

import (
	"errors"
	"fmt"
)

var ErrOptedOut = errors.New("recipient opted out")

func deliver(to string) error {
	if to == "+1-555-0199" {
		return ErrOptedOut
	}
	return nil
}

func send(to string) error {
	if err := deliver(to); err != nil {
		return fmt.Errorf("send to %s: %w", to, err)
	}
	return nil
}

func main() {
	err := send("+1-555-0199")
	fmt.Println(err)
	fmt.Println(err == ErrOptedOut)
	fmt.Println(errors.Is(err, ErrOptedOut))

	if errors.Is(err, ErrOptedOut) {
		fmt.Println("removing from mailing list, no retry")
	}
}
```

```text
send to +1-555-0199: recipient opted out
false
true
removing from mailing list, no retry
```

Always use `errors.Is` rather than `==` to check for sentinel errors. It works whether or not the error was wrapped.

## `%w` versus `%v`

You can also put an error into `fmt.Errorf` with `%v`. The message looks identical, but the original error is **not** wrapped, so `errors.Is` can't find it. Use `%w` when callers might want to inspect the cause; use `%v` when you deliberately want to hide the details.

## Handle it or return it, not both

When your code gets an error, do **one** of two things:

1. **Handle it**: log it, retry, show the user a message, fall back to a default.
2. **Return it** (usually wrapped with context) and let the caller decide.

Don't log it *and* return it. Otherwise the same error gets logged at every level on its way up, and your logs fill with duplicates.

## Your turn

Complete `validate`, which checks a message before Textio sends it:

1. If `to` doesn't start with `"+"`, return `ErrInvalidPhone`, wrapped with the context `to <number>: `.
2. Otherwise, if `body` is empty, return `ErrEmptyBody`, wrapped the same way.
3. Otherwise, return `nil`.

For example, `validate("555-0100", "hi")` returns an error whose message is `to 555-0100: invalid phone number`, and `errors.Is(err, ErrInvalidPhone)` must be `true`. Use `fmt.Errorf` with `%w`, and `strings.HasPrefix` to check the `+`.

## Further reading

- [The Go Blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Go by Example: Errors](https://gobyexample.com/errors)
