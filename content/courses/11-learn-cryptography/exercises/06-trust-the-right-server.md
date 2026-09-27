---
title: Trust the Right Server
difficulty: easy
after: certificates-and-tls
hints:
  - '`InsecureSkipVerify: true` doesn''t skip *one* check; it skips the chain, the expiry and the hostname. A man-in-the-middle with any certificate at all gets your vault traffic. Delete it rather than trying to fix it up with callbacks.'
  - 'Three fields do the job: `RootCAs: roots` (trust this pool instead of the system roots), `ServerName: serverName` (the name the certificate must list in its SANs) and `MinVersion: tls.VersionTLS13`. With those set, Go verifies everything during the handshake.'
exercise:
  starter: |
    package main

    import (
    	"crypto/tls"
    	"crypto/x509"
    	"fmt"
    	"io"
    	"log"
    	"net/http"
    	"net/http/httptest"
    )

    // syncClientConfig returns the TLS config a Keybox device uses to reach its
    // sync server. It must trust only the certificates in roots, and only a
    // certificate valid for serverName.
    //
    // BUG: someone "fixed" a certificate error during development like this,
    // which turns off every check TLS does. Remove InsecureSkipVerify and set:
    //   - RootCAs to roots (the CA pool to trust instead of the system's),
    //   - ServerName to serverName (the name the certificate must be valid for),
    //   - MinVersion to tls.VersionTLS13.
    func syncClientConfig(roots *x509.CertPool, serverName string) *tls.Config {
    	return &tls.Config{InsecureSkipVerify: true}
    }

    func main() {
    	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "vault synced")
    	}))
    	ts.Config.ErrorLog = log.New(io.Discard, "", 0) // hide the server's handshake errors
    	ts.StartTLS()
    	defer ts.Close()

    	roots := x509.NewCertPool()
    	roots.AddCert(ts.Certificate()) // httptest's certificate is valid for example.com

    	for _, name := range []string{"example.com", "evil.example"} {
    		client := &http.Client{Transport: &http.Transport{TLSClientConfig: syncClientConfig(roots, name)}}
    		resp, err := client.Get(ts.URL)
    		if err != nil {
    			fmt.Printf("%-12s refused: %v\n", name, err) // want: evil.example refused
    			continue
    		}
    		body, _ := io.ReadAll(resp.Body)
    		resp.Body.Close()
    		fmt.Printf("%-12s %s\n", name, body) // want: example.com vault synced
    	}
    }
  solution: |
    package main

    import (
    	"crypto/tls"
    	"crypto/x509"
    	"fmt"
    	"io"
    	"log"
    	"net/http"
    	"net/http/httptest"
    )

    // syncClientConfig returns the TLS config a Keybox device uses to reach its
    // sync server: trust only roots, and only a certificate valid for serverName.
    func syncClientConfig(roots *x509.CertPool, serverName string) *tls.Config {
    	return &tls.Config{
    		RootCAs:    roots,
    		ServerName: serverName,
    		MinVersion: tls.VersionTLS13,
    	}
    }

    func main() {
    	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "vault synced")
    	}))
    	ts.Config.ErrorLog = log.New(io.Discard, "", 0) // hide the server's handshake errors
    	ts.StartTLS()
    	defer ts.Close()

    	roots := x509.NewCertPool()
    	roots.AddCert(ts.Certificate())

    	for _, name := range []string{"example.com", "evil.example"} {
    		client := &http.Client{Transport: &http.Transport{TLSClientConfig: syncClientConfig(roots, name)}}
    		resp, err := client.Get(ts.URL)
    		if err != nil {
    			fmt.Printf("%-12s refused: %v\n", name, err)
    			continue
    		}
    		body, _ := io.ReadAll(resp.Body)
    		resp.Body.Close()
    		fmt.Printf("%-12s %s\n", name, body)
    	}
    }
  tests: |
    package main

    import (
    	"crypto/tls"
    	"crypto/x509"
    	"fmt"
    	"io"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"testing"
    )

    func newServer(t *testing.T) *httptest.Server {
    	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "vault synced")
    	}))
    	ts.Config.ErrorLog = log.New(io.Discard, "", 0)
    	ts.StartTLS()
    	t.Cleanup(ts.Close)
    	return ts
    }

    // get fetches ts.URL with cfg and returns the body or the error.
    func get(t *testing.T, ts *httptest.Server, cfg *tls.Config) (string, error) {
    	t.Helper()
    	if cfg == nil {
    		t.Fatal("syncClientConfig returned nil")
    	}
    	tr := &http.Transport{TLSClientConfig: cfg}
    	defer tr.CloseIdleConnections()
    	resp, err := (&http.Client{Transport: tr}).Get(ts.URL)
    	if err != nil {
    		return "", err
    	}
    	defer resp.Body.Close()
    	b, err := io.ReadAll(resp.Body)
    	return string(b), err
    }

    func TestConnectsToGenuineServer(t *testing.T) {
    	ts := newServer(t)
    	roots := x509.NewCertPool()
    	roots.AddCert(ts.Certificate())
    	body, err := get(t, ts, syncClientConfig(roots, "example.com"))
    	if err != nil || body != "vault synced" {
    		t.Errorf("connecting to the genuine server as example.com = %q, %v, want %q, nil", body, err, "vault synced")
    	}
    }

    func TestChecksHostname(t *testing.T) {
    	ts := newServer(t)
    	roots := x509.NewCertPool()
    	roots.AddCert(ts.Certificate())
    	for _, name := range []string{"evil.example", "sync.keybox.test", "example.co"} {
    		if body, err := get(t, ts, syncClientConfig(roots, name)); err == nil {
    			t.Errorf("connected with ServerName %q to a server whose certificate is only for example.com (got %q): the hostname isn't being checked", name, body)
    		}
    	}
    }

    func TestChecksIssuer(t *testing.T) {
    	ts := newServer(t)
    	if body, err := get(t, ts, syncClientConfig(x509.NewCertPool(), "example.com")); err == nil {
    		t.Errorf("connected (got %q) although roots is empty: the certificate chain isn't being checked", body)
    	}
    }

    func TestConfigFields(t *testing.T) {
    	roots := x509.NewCertPool()
    	cfg := syncClientConfig(roots, "sync.keybox.test")
    	if cfg == nil {
    		t.Fatal("syncClientConfig returned nil")
    	}
    	if cfg.InsecureSkipVerify {
    		t.Error("InsecureSkipVerify is still true: it disables all certificate checks")
    	}
    	if cfg.RootCAs != roots {
    		t.Error("RootCAs is not the roots pool passed in")
    	}
    	if cfg.ServerName != "sync.keybox.test" {
    		t.Errorf("ServerName = %q, want %q", cfg.ServerName, "sync.keybox.test")
    	}
    	if cfg.MinVersion != tls.VersionTLS13 {
    		t.Errorf("MinVersion = %s, want TLS 1.3", tls.VersionName(cfg.MinVersion))
    	}
    }
---

Keybox devices sync over TLS with a server whose certificate comes from Keybox's own
private CA. During development someone hit `x509: certificate signed by unknown
authority` and "fixed" it with `InsecureSkipVerify: true`, which shipped. Now the
client happily talks to **any** server presenting **any** certificate.

Fix `syncClientConfig(roots, serverName)` so the returned `*tls.Config`:

- trusts only the CAs in `roots`,
- accepts only a certificate valid for `serverName`, even though the client may dial
  an IP address,
- requires TLS 1.3, and
- does **not** set `InsecureSkipVerify`.

## Example

The tests (and `main`) run a real TLS server in-process with `httptest`. Its built-in
certificate is valid for `example.com`:

```
syncClientConfig(roots, "example.com")   -> connects, reads "vault synced"
syncClientConfig(roots, "evil.example")  -> x509: certificate is valid for example.com, *.example.com, not evil.example
syncClientConfig(emptyPool, ...)         -> x509: certificate signed by unknown authority
```

## Constraints

- No network access: everything runs over loopback.
- When you do need extra checks (such as pinning a key), add them with
  `VerifyConnection`, which runs *after* Go's normal verification, instead of
  switching verification off.
