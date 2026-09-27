---
title: Never InsecureSkipVerify
quiz:
  - question: 'What does `tls.Config{InsecureSkipVerify: true}` actually turn off?'
    options:
      - text: Encryption, so traffic is sent in plain text
      - text: Checking the server's certificate chain and hostname, so any server (including an attacker's) is accepted
        correct: true
      - text: Only the expiry date check
    explanation: |
      The connection is still encrypted, but you no longer know *who* is on the other
      end. A machine-in-the-middle can present its own certificate, decrypt everything
      (tokens included) and pass it along.
  - question: Your company's internal Trackr server uses a certificate from a private CA, and `trackr` fails with "certificate signed by unknown authority". What's the right fix?
    options:
      - text: Set `InsecureSkipVerify` just for that server
      - text: Add the company's CA certificate to `RootCAs` (or the system trust store)
        correct: true
      - text: Switch to `http://`
    explanation: |
      The problem is that the client doesn't know your CA. Teach it that CA, and you
      keep full verification against everything else.
exercise:
  starter: |
    package main

    import (
    	"encoding/pem"
    	"fmt"
    	"io"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"time"
    )

    // newTrustingClient returns a client that trusts the system's root CAs plus
    // the CA certificate(s) in caPEM, with full verification left on. It uses a
    // clone of http.DefaultTransport and a 10-second timeout. It returns an error
    // if caPEM holds no certificates.
    func newTrustingClient(caPEM []byte) (*http.Client, error) {
    	// ?
    	return &http.Client{Timeout: 10 * time.Second}, nil
    }

    func main() {
    	// Stand-in for your company's internal Trackr server and its private CA.
    	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "hello from the internal Trackr server")
    	}))
    	srv.Config.ErrorLog = log.New(io.Discard, "", 0)
    	srv.StartTLS()
    	defer srv.Close()
    	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})

    	client, err := newTrustingClient(caPEM)
    	if err != nil {
    		fmt.Println("building client:", err)
    		return
    	}
    	resp, err := client.Get(srv.URL)
    	if err != nil {
    		fmt.Println("request failed:", err)
    		return
    	}
    	defer resp.Body.Close()
    	body, _ := io.ReadAll(resp.Body)
    	fmt.Println(resp.Status, string(body))
    }
  solution: |
    package main

    import (
    	"crypto/tls"
    	"crypto/x509"
    	"encoding/pem"
    	"errors"
    	"fmt"
    	"io"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"time"
    )

    func newTrustingClient(caPEM []byte) (*http.Client, error) {
    	pool, err := x509.SystemCertPool()
    	if err != nil {
    		pool = x509.NewCertPool()
    	}
    	if !pool.AppendCertsFromPEM(caPEM) {
    		return nil, errors.New("no certificates found in the CA file")
    	}

    	t := http.DefaultTransport.(*http.Transport).Clone()
    	t.TLSClientConfig = &tls.Config{RootCAs: pool}
    	return &http.Client{Transport: t, Timeout: 10 * time.Second}, nil
    }

    func main() {
    	// Stand-in for your company's internal Trackr server and its private CA.
    	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "hello from the internal Trackr server")
    	}))
    	srv.Config.ErrorLog = log.New(io.Discard, "", 0)
    	srv.StartTLS()
    	defer srv.Close()
    	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})

    	client, err := newTrustingClient(caPEM)
    	if err != nil {
    		fmt.Println("building client:", err)
    		return
    	}
    	resp, err := client.Get(srv.URL)
    	if err != nil {
    		fmt.Println("request failed:", err)
    		return
    	}
    	defer resp.Body.Close()
    	body, _ := io.ReadAll(resp.Body)
    	fmt.Println(resp.Status, string(body))
    }
  tests: |
    package main

    import (
    	"encoding/pem"
    	"io"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing"
    	"time"
    )

    func tlsServer(t *testing.T) (*httptest.Server, []byte) {
    	t.Helper()
    	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		io.WriteString(w, "internal")
    	}))
    	srv.Config.ErrorLog = log.New(io.Discard, "", 0)
    	srv.StartTLS()
    	t.Cleanup(srv.Close)
    	return srv, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
    }

    func TestTrustsPrivateCA(t *testing.T) {
    	srv, caPEM := tlsServer(t)
    	client, err := newTrustingClient(caPEM)
    	if err != nil {
    		t.Fatalf("newTrustingClient returned error %v", err)
    	}
    	resp, err := client.Get(srv.URL)
    	if err != nil {
    		t.Fatalf("GET %s with the private CA trusted: %v", srv.URL, err)
    	}
    	defer resp.Body.Close()
    	if body, _ := io.ReadAll(resp.Body); string(body) != "internal" {
    		t.Errorf("body = %q, want %q", body, "internal")
    	}
    }

    func TestStillVerifies(t *testing.T) {
    	srv, caPEM := tlsServer(t)
    	client, err := newTrustingClient(caPEM)
    	if err != nil {
    		t.Fatalf("newTrustingClient returned error %v", err)
    	}
    	// The test certificate is valid for 127.0.0.1 but not for "localhost".
    	wrongName := strings.Replace(srv.URL, "127.0.0.1", "localhost", 1)
    	if resp, err := client.Get(wrongName); err == nil {
    		resp.Body.Close()
    		t.Errorf("GET %s succeeded, but the certificate isn't valid for localhost: verification must stay on (no InsecureSkipVerify!)", wrongName)
    	}

    	tr, ok := client.Transport.(*http.Transport)
    	if !ok {
    		t.Fatalf("client.Transport is %T, want a *http.Transport cloned from http.DefaultTransport", client.Transport)
    	}
    	if tr.TLSClientConfig == nil || tr.TLSClientConfig.RootCAs == nil {
    		t.Error("TLSClientConfig.RootCAs isn't set: put your CertPool there")
    	} else if tr.TLSClientConfig.InsecureSkipVerify {
    		t.Error("InsecureSkipVerify is true. Never.")
    	}
    	if tr == http.DefaultTransport {
    		t.Error("the client uses http.DefaultTransport itself: Clone it instead of changing the shared one")
    	}
    	if tr.Proxy == nil || tr.IdleConnTimeout != 90*time.Second {
    		t.Error("the transport lost DefaultTransport's settings (Proxy, IdleConnTimeout): start from http.DefaultTransport.(*http.Transport).Clone()")
    	}
    	if client.Timeout != 10*time.Second {
    		t.Errorf("client.Timeout = %v, want 10s", client.Timeout)
    	}
    }

    func TestBadPEM(t *testing.T) {
    	for _, in := range []string{"", "not a certificate", "-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----\n"} {
    		if _, err := newTrustingClient([]byte(in)); err == nil {
    			t.Errorf("newTrustingClient(%q) returned nil error, want an error (no certificates in it)", in)
    		}
    	}
    }
---

Sooner or later, a certificate error blocks someone, and they search the error
message. The top answer will say:

```go
// DON'T
t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
```

It makes the error go away. It also makes TLS nearly pointless.

## What it really does

`InsecureSkipVerify: true` tells Go to accept **any** certificate for **any**
hostname. The connection is still encrypted, but you've given up the guarantee that
you're talking to the right server. The `crypto/tls` docs say it plainly: in this mode
TLS "is susceptible to machine-in-the-middle attacks".

Here's the attack. You're on a hotel Wi-Fi, and the access point intercepts
connections to `api.trackr.dev`. It presents its own certificate. A verifying client
refuses. A client with `InsecureSkipVerify` happily completes the handshake and sends:

```
Authorization: Bearer tk_live_8f3a...
```

to the attacker, who forwards it to the real API so nothing looks wrong. Your token is
now theirs.

## "It's only for testing"

That's how it always starts. Then the flag is made configurable "for staging", someone
sets it in production to get past an outage, and it never gets unset. Code that can
skip verification eventually does. Tests don't need it anyway:

- For an `httptest` TLS server, use **`srv.Client()`**. It trusts exactly the test
  server's certificate and keeps verification on.
- Chapter 9's `httptest.NewTestServer` (Go 1.27) gives you a client already wired to
  the server.

## Fix the real problem

Certificate errors have real causes, and each has a proper fix:

| Error | Likely cause | Fix |
| --- | --- | --- |
| `certificate signed by unknown authority` | private or corporate CA, or a TLS-inspecting proxy | trust that CA (below) |
| `certificate is valid for X, not Y` | wrong hostname in the URL, or a misconfigured server | fix the URL, or ask for a correct certificate |
| `certificate has expired or is not yet valid` | expired certificate, or your clock is wrong | the server must renew; check your clock |

### Trusting a private CA

If your company runs its own CA, add **its** certificate to the client's roots, on top
of the system ones:

```go
caPEM, err := os.ReadFile("trackr-ca.pem") // the CA certificate your IT team gave you
if err != nil {
	return nil, err
}

pool, err := x509.SystemCertPool() // start from the normal roots...
if err != nil {
	pool = x509.NewCertPool()
}
if !pool.AppendCertsFromPEM(caPEM) { // ...and add the private CA
	return nil, errors.New("trackr-ca.pem holds no certificates")
}

t := http.DefaultTransport.(*http.Transport).Clone()
t.TLSClientConfig = &tls.Config{RootCAs: pool}
client := &http.Client{Transport: t, Timeout: 10 * time.Second}
```

The important part is `tls.Config{RootCAs: pool}`. Verification stays fully on. It
just knows about one more CA. Start from `x509.SystemCertPool()` so public servers
keep working, and clone `DefaultTransport` so you keep its other settings. The path
to the PEM file is something the user configures.

Often you don't even need code. On Linux, pointing `SSL_CERT_FILE` at a bundle that
includes the company CA, or installing the CA into the system store, fixes every Go
program on the machine at once.

## Other settings to leave alone

The TLS defaults are good, so resist tuning them:

- Don't lower `MinVersion` below its default (TLS 1.2).
- Don't set `CipherSuites` or `CurvePreferences`. The defaults are secure and include
  post-quantum key exchange, and setting the list yourself can silently turn that off.
- Don't set `KeyLogWriter` outside a debugging session. It writes session secrets that
  let anyone decrypt captured traffic.

## If you truly must

There are rare, legitimate cases (a device with a self-signed certificate you
**pin** manually). The safe way is still not blanket skipping: keep verification and
add a `VerifyConnection` callback, or compare the certificate's fingerprint against a
known value. That's advanced territory. For `trackr`, the answer is simply: never.

## Your turn

Package the private-CA setup as a function `trackr` can call at startup. The program
in the editor starts an HTTPS test server whose certificate plays the part of your
company's CA, and hands you that certificate as PEM. Right now the request fails with
`certificate signed by unknown authority`.

Complete `newTrustingClient(caPEM)` so that it:

1. starts from `x509.SystemCertPool()` (falling back to `x509.NewCertPool()` if that
   fails) and adds `caPEM` with `AppendCertsFromPEM`;
2. returns an error if `caPEM` doesn't contain a single certificate;
3. clones `http.DefaultTransport` and sets `TLSClientConfig` to a `tls.Config` whose
   `RootCAs` is your pool, with verification left **on**;
4. returns an `*http.Client` using that transport, with a 10-second `Timeout`.

The tests also check that the client still rejects a certificate issued for a
different hostname.
