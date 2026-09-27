---
title: Keeping Secrets Out of Logs
quiz:
  - question: '`Secret` has a `String() string` method returning `"[REDACTED]"`. What does `fmt.Printf("%x", s)` print, if `Secret` is `struct{ value []byte }`?'
    options:
      - text: '`[REDACTED]`'
      - text: The hex encoding of `"[REDACTED]"`
        correct: true
      - text: The secret in hex
      - text: A compile error
    explanation: |
      For `%x`, `%s`, `%v` and `%q`, fmt uses `String()` if it exists, so `%x` prints
      the hex of the *redacted* string, which is safe but surprising. It's `%#v` that
      bypasses `String()` (it uses `GoString()` or prints the fields). Implementing
      `fmt.Formatter` takes control of every verb at once.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"log/slog"
    	"os"
    )

    const redacted = "[REDACTED]"

    // Secret holds sensitive bytes. Printing, logging or JSON-encoding it must
    // never reveal them; only Reveal does.
    type Secret struct {
    	value []byte
    }

    func NewSecret(b []byte) Secret { return Secret{value: b} }

    // Reveal returns the secret bytes. Call it only where they're needed.
    func (s Secret) Reveal() []byte { return s.value }

    // ? Implement:
    //   Format(f fmt.State, verb rune)       - fmt.Formatter: every verb prints redacted
    //   LogValue() slog.Value                - slog.LogValuer: logs as redacted
    //   MarshalJSON() ([]byte, error)        - json.Marshaler: encodes as "[REDACTED]"

    type Credential struct {
    	Name  string
    	Token Secret
    }

    func main() {
    	c := Credential{Name: "github-token", Token: NewSecret([]byte("ghp_TOPSECRET"))}
    	fmt.Printf("%v\n%+v\n%#v\n", c, c, c)
    	slog.New(slog.NewTextHandler(os.Stdout, nil)).Info("loaded", "token", c.Token)
    	fmt.Printf("revealed only on purpose: %s\n", c.Token.Reveal())
    }
  solution: |
    package main

    import (
    	"encoding/json"
    	"fmt"
    	"io"
    	"log/slog"
    	"os"
    )

    const redacted = "[REDACTED]"

    type Secret struct {
    	value []byte
    }

    func NewSecret(b []byte) Secret { return Secret{value: b} }

    func (s Secret) Reveal() []byte { return s.value }

    func (s Secret) Format(f fmt.State, verb rune) {
    	io.WriteString(f, redacted)
    }

    func (s Secret) LogValue() slog.Value {
    	return slog.StringValue(redacted)
    }

    func (s Secret) MarshalJSON() ([]byte, error) {
    	return json.Marshal(redacted)
    }

    type Credential struct {
    	Name  string
    	Token Secret
    }

    func main() {
    	c := Credential{Name: "github-token", Token: NewSecret([]byte("ghp_TOPSECRET"))}
    	fmt.Printf("%v\n%+v\n%#v\n", c, c, c)
    	slog.New(slog.NewTextHandler(os.Stdout, nil)).Info("loaded", "token", c.Token)
    	fmt.Printf("revealed only on purpose: %s\n", c.Token.Reveal())
    }
  tests: |
    package main

    import (
    	"bytes"
    	"encoding/json"
    	"fmt"
    	"log/slog"
    	"strings"
    	"testing"
    )

    const plain = "ghp_TOPSECRET"

    func leaks(s string) bool {
    	return strings.Contains(s, plain) || strings.Contains(s, fmt.Sprintf("%x", plain)) ||
    		strings.Contains(s, "103 104 112") // the bytes printed as numbers
    }

    func TestFmtVerbs(t *testing.T) {
    	s := NewSecret([]byte(plain))
    	c := Credential{Name: "github-token", Token: s}
    	for _, f := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%X", "%d"} {
    		for _, v := range []any{s, c, &c, []Secret{s}} {
    			out := fmt.Sprintf(f, v)
    			if leaks(out) {
    				t.Errorf("fmt.Sprintf(%q, %T) = %q, which leaks the secret", f, v, out)
    			}
    		}
    	}
    	if got := fmt.Sprint(s); got != redacted {
    		t.Errorf("fmt.Sprint(secret) = %q, want %q", got, redacted)
    	}
    	if err := fmt.Errorf("bad token %v", s); leaks(err.Error()) {
    		t.Errorf("error message leaks the secret: %v", err)
    	}
    }

    func TestSlog(t *testing.T) {
    	s := NewSecret([]byte(plain))
    	for name, h := range map[string]func(*bytes.Buffer) slog.Handler{
    		"text": func(b *bytes.Buffer) slog.Handler { return slog.NewTextHandler(b, nil) },
    		"json": func(b *bytes.Buffer) slog.Handler { return slog.NewJSONHandler(b, nil) },
    	} {
    		var buf bytes.Buffer
    		logger := slog.New(h(&buf))
    		logger.Info("loaded", "token", s, "cred", Credential{Name: "x", Token: s})
    		if leaks(buf.String()) {
    			t.Errorf("%s handler output leaks the secret: %s", name, buf.String())
    		}
    		if !strings.Contains(buf.String(), redacted) {
    			t.Errorf("%s handler output should show %s: %s", name, redacted, buf.String())
    		}
    	}
    }

    func TestJSON(t *testing.T) {
    	out, err := json.Marshal(Credential{Name: "github-token", Token: NewSecret([]byte(plain))})
    	if err != nil {
    		t.Fatal(err)
    	}
    	if want := `{"Name":"github-token","Token":"[REDACTED]"}`; string(out) != want {
    		t.Errorf("json.Marshal(credential) = %s, want %s", out, want)
    	}
    }

    func TestRevealStillWorks(t *testing.T) {
    	if got := NewSecret([]byte(plain)).Reveal(); string(got) != plain {
    		t.Errorf("Reveal() = %q, want %q", got, plain)
    	}
    }
---

Cryptography can't protect a key that's printed to a log file. Logs are copied to
aggregators, kept in backups for years, searched by many people, and pasted into
tickets. The same goes for error messages, panics, metrics and debug output. So let's
make leaking secrets hard to do by accident, using Go's formatting interfaces.

## How secrets escape

None of these look like "log the password":

```go
log.Printf("loaded credential %+v", cred)            // struct includes the token
return fmt.Errorf("decrypt %s with key %x: %w", name, key, err)
slog.Info("request", "headers", r.Header)            // Authorization header
panic(fmt.Sprintf("bad config: %#v", cfg))           // config holds the DB password
```

The fix is a type for secret values that **can't be printed**. Code that needs the
value calls an explicit method, which stands out in review.

## The interfaces to implement

Go has one interface per output path:

- **`fmt.Formatter`**: `Format(f fmt.State, verb rune)`. If a type implements it, the
  `fmt` package calls it for **every** verb (`%v`, `%+v`, `%#v`, `%s`, `%x`, `%d`...)
  instead of doing its own formatting. That includes when the secret is a field inside
  a struct being printed. `String()` alone isn't enough: `%#v` bypasses it and prints
  the raw fields.
- **`slog.LogValuer`**: `LogValue() slog.Value`. slog calls it to find out what to log
  for a value. Return `slog.StringValue("[REDACTED]")`.
- **`json.Marshaler`**: `MarshalJSON() ([]byte, error)`, so APIs and JSON logs encode it
  as `"[REDACTED]"`.

```go
func (s Secret) Format(f fmt.State, verb rune) {
	io.WriteString(f, "[REDACTED]")
}
```

Keep the field **unexported**, so other packages can't reach in, and provide a
deliberately named accessor like `Reveal()`.

## Beyond redaction

- **Log key IDs, not keys**: `"key_id", 2` is useful when debugging a rotation;
  `"key", key` is an incident.
- **Log lengths and types, not contents**: `"secret_bytes", len(v)`.
- **Scrub request logging**: drop `Authorization`, `Cookie` and `Set-Cookie` headers
  and any query parameter named like `token` or `key`.
- **Error messages**: say *what* failed (`decrypting vault entry "github-token"`), never
  include the key or plaintext.
- **Memory**: Go doesn't guarantee that zeroing a slice removes every copy, since the
  garbage collector may have moved or copied data. The experimental `runtime/secret`
  package (built only with `GOEXPERIMENT=runtimesecret`) explores better support. Don't
  build a security argument on wiping memory; limit how long secrets live and where
  they go instead.

## Your task

Make Keybox's `Secret` type impossible to print by accident. Add three methods:

1. `Format(f fmt.State, verb rune)` that writes `redacted` for every verb
   (`io.WriteString(f, redacted)`).
2. `LogValue() slog.Value` returning `slog.StringValue(redacted)`.
3. `MarshalJSON() ([]byte, error)` returning the JSON string `"[REDACTED]"`
   (`json.Marshal(redacted)` does it).

You'll need `io` and `encoding/json` imports. The tests print a `Secret` and a struct
containing one with every common verb, log them through slog's text and JSON handlers,
and JSON-encode them. `Reveal()` must still return the real bytes.
