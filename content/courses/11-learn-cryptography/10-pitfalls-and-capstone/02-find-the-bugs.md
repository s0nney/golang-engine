---
title: Find the Bugs
quiz:
  - question: A function slices `sealed[:12]` before checking `len(sealed)`. Mallory sends a 3-byte body. What's the impact?
    options:
      - text: None; Go bounds-checks slices
      - text: A panic from attacker-controlled input, which can crash the process (or the goroutine's request) on demand, a denial of service
        correct: true
      - text: Mallory learns the key
      - text: The GCM tag check is skipped
    explanation: |
      Go's bounds check turns memory corruption into a panic, which is much better than
      C, but a remotely triggerable panic is still a bug. `net/http` recovers panics
      per request, yet other servers and background workers may simply die. Validate
      lengths first.
exercise:
  starter: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"errors"
    	"fmt"
    	"log/slog"
    	"os"
    )

    var ErrDecrypt = errors.New("keybox: cannot open export")

    // Exporter seals individual secrets for export files. There are FOUR
    // security bugs in Seal and Open. Find and fix them all.
    type Exporter struct {
    	aead   cipher.AEAD
    	logger *slog.Logger
    }

    func NewExporter(key []byte, logger *slog.Logger) (*Exporter, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	aead, err := cipher.NewGCM(block)
    	if err != nil {
    		return nil, err
    	}
    	return &Exporter{aead: aead, logger: logger}, nil
    }

    // Seal encrypts secret, bound to name, as nonce || ciphertext || tag.
    func (e *Exporter) Seal(name string, secret []byte) []byte {
    	nonce := make([]byte, e.aead.NonceSize())
    	e.logger.Info("sealing secret", "name", name, "secret", string(secret))
    	return e.aead.Seal(nonce, nonce, secret, nil)
    }

    // Open reverses Seal. Every failure returns ErrDecrypt.
    func (e *Exporter) Open(name string, sealed []byte) ([]byte, error) {
    	nonce, ct := sealed[:e.aead.NonceSize()], sealed[e.aead.NonceSize():]
    	pt, err := e.aead.Open(nil, nonce, ct, nil)
    	if err != nil {
    		e.logger.Error("open failed", "name", name, "err", err)
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func main() {
    	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    	ex, _ := NewExporter(bytes.Repeat([]byte{7}, 32), logger)
    	a := ex.Seal("github-token", []byte("ghp_alice"))
    	b := ex.Seal("wifi-password", []byte("hunter2"))
    	fmt.Printf("nonces: %x %x\n", a[:12], b[:12])
    	pt, err := ex.Open("github-token", b) // swapped!
    	fmt.Printf("swapped open: %q %v\n", pt, err)
    }
  solution: |
    package main

    import (
    	"bytes"
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/rand"
    	"errors"
    	"fmt"
    	"log/slog"
    	"os"
    )

    var ErrDecrypt = errors.New("keybox: cannot open export")

    type Exporter struct {
    	aead   cipher.AEAD
    	logger *slog.Logger
    }

    func NewExporter(key []byte, logger *slog.Logger) (*Exporter, error) {
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	aead, err := cipher.NewGCM(block)
    	if err != nil {
    		return nil, err
    	}
    	return &Exporter{aead: aead, logger: logger}, nil
    }

    func exportAD(name string) []byte {
    	return []byte("keybox export v1\x00" + name)
    }

    func (e *Exporter) Seal(name string, secret []byte) []byte {
    	nonce := make([]byte, e.aead.NonceSize())
    	rand.Read(nonce)                                                    // fix 1: a fresh random nonce every time
    	e.logger.Info("sealing secret", "name", name, "bytes", len(secret)) // fix 2: never log the secret
    	return e.aead.Seal(nonce, nonce, secret, exportAD(name))            // fix 3: bind to the name
    }

    func (e *Exporter) Open(name string, sealed []byte) ([]byte, error) {
    	if len(sealed) < e.aead.NonceSize()+e.aead.Overhead() { // fix 4: check before slicing
    		return nil, ErrDecrypt
    	}
    	nonce, ct := sealed[:e.aead.NonceSize()], sealed[e.aead.NonceSize():]
    	pt, err := e.aead.Open(nil, nonce, ct, exportAD(name))
    	if err != nil {
    		e.logger.Error("open failed", "name", name, "err", err)
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    func main() {
    	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    	ex, _ := NewExporter(bytes.Repeat([]byte{7}, 32), logger)
    	a := ex.Seal("github-token", []byte("ghp_alice"))
    	b := ex.Seal("wifi-password", []byte("hunter2"))
    	fmt.Printf("nonces: %x %x\n", a[:12], b[:12])
    	pt, err := ex.Open("github-token", b)
    	fmt.Printf("swapped open: %q %v\n", pt, err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"errors"
    	"log/slog"
    	"strings"
    	"testing"
    )

    func newTestExporter(t *testing.T) (*Exporter, *bytes.Buffer) {
    	t.Helper()
    	var logs bytes.Buffer
    	ex, err := NewExporter(bytes.Repeat([]byte{0x5c}, 32), slog.New(slog.NewJSONHandler(&logs, nil)))
    	if err != nil {
    		t.Fatal(err)
    	}
    	return ex, &logs
    }

    func TestRoundTrip(t *testing.T) {
    	ex, _ := newTestExporter(t)
    	s := ex.Seal("github-token", []byte("ghp_alice"))
    	if pt, err := ex.Open("github-token", s); err != nil || string(pt) != "ghp_alice" {
    		t.Fatalf("Open(Seal(x)) = %q, %v; want \"ghp_alice\", nil", pt, err)
    	}
    }

    func TestBug1FreshNonces(t *testing.T) {
    	ex, _ := newTestExporter(t)
    	seen := map[string]bool{}
    	for range 100 {
    		n := string(ex.Seal("github-token", []byte("same"))[:12])
    		if seen[n] {
    			t.Fatalf("Seal reused nonce %x; every seal needs a fresh random nonce", n)
    		}
    		seen[n] = true
    	}
    }

    func TestBug2NoSecretsInLogs(t *testing.T) {
    	ex, logs := newTestExporter(t)
    	s := ex.Seal("github-token", []byte("ghp_TOPSECRET"))
    	ex.Open("github-token", s)
    	ex.Open("github-token", []byte("garbage that is long enough to not be short!!"))
    	if strings.Contains(logs.String(), "TOPSECRET") {
    		t.Errorf("the plaintext secret appears in the logs:\n%s", logs.String())
    	}
    }

    func TestBug3BoundToName(t *testing.T) {
    	ex, _ := newTestExporter(t)
    	wifi := ex.Seal("wifi-password", []byte("hunter2"))
    	if pt, err := ex.Open("github-token", wifi); !errors.Is(err, ErrDecrypt) {
    		t.Errorf("opening wifi-password's ciphertext as github-token = %q, %v; want ErrDecrypt (use the name as associated data)", pt, err)
    	}
    }

    func TestBug4ShortInput(t *testing.T) {
    	ex, _ := newTestExporter(t)
    	for _, n := range []int{0, 3, 11, 12, 27} {
    		func() {
    			defer func() {
    				if r := recover(); r != nil {
    					t.Errorf("Open(%d-byte input) panicked: %v", n, r)
    				}
    			}()
    			if _, err := ex.Open("x", make([]byte, n)); !errors.Is(err, ErrDecrypt) {
    				t.Errorf("Open(%d-byte input) error = %v, want ErrDecrypt", n, err)
    			}
    		}()
    	}
    }

    func TestTamperStillDetected(t *testing.T) {
    	ex, _ := newTestExporter(t)
    	s := ex.Seal("github-token", []byte("ghp_alice"))
    	s[len(s)-1] ^= 1
    	if _, err := ex.Open("github-token", s); !errors.Is(err, ErrDecrypt) {
    		t.Errorf("Open(tampered) error = %v, want ErrDecrypt", err)
    	}
    }
---

Time for a code review. Keybox's export feature seals individual secrets so they can be
written to a file and imported elsewhere. It uses AES-256-GCM, a real AEAD, from the
standard library, with a proper 32-byte key. And it still has **four** security bugs.

## Reading crypto code like an attacker

When reviewing code that uses cryptography, don't just check that the right algorithm
is named. Walk through each input and ask:

1. **Where does every random or unique value come from?** Keys, nonces, salts, IDs.
   Is it `crypto/rand`? Could it ever repeat?
2. **What is bound to what?** Does the ciphertext or signature cover the context it's
   used in: the record's name, the user, the version, the purpose?
3. **What happens with hostile input?** Empty, short, huge, malformed, swapped,
   replayed. Is every length checked before slicing? Is every error treated as final?
4. **Where does secret data go?** Logs, errors, metrics, panics, responses. Only
   ciphertext and non-sensitive metadata should leave the function.
5. **Is the comparison constant-time** wherever a secret is compared?

Running the starter shows two of the problems immediately: both exports print the same
nonce, and opening Wi-Fi's ciphertext under the name `github-token` succeeds. The log
output shows a third.

## Your task

Fix all four bugs in `Seal` and `Open`, keeping the output layout
`nonce || ciphertext || tag`:

1. **Nonce.** Every seal must use a fresh random nonce from `crypto/rand`.
2. **Logging.** The logger must never receive the secret. Logging the name and the
   secret's *length* is fine.
3. **Binding.** A ciphertext sealed for one name must not open under another. Pass
   associated data that includes the name, the same in `Seal` and `Open`. (A fixed
   label such as `"keybox export v1\x00"` in front of the name is good practice.)
4. **Hostile input.** `Open` must return `ErrDecrypt`, never panic, when `sealed` is
   shorter than `NonceSize() + Overhead()`.

Each bug has its own test, named after it, so you can see which ones you've fixed.
