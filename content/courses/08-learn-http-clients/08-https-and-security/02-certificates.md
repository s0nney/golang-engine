---
title: Certificates
quiz:
  - question: '`trackr` connects to `https://api.trackr.dev`, and the certificate is valid and signed by a trusted CA, but it was issued for `www.trackr.dev` only. What happens?'
    options:
      - text: The connection works, because the certificate is valid
      - text: The handshake fails with a hostname error
        correct: true
      - text: Go asks the user whether to continue
    explanation: |
      A certificate proves identity for specific names. A perfectly valid certificate
      for a different name proves nothing about this server, so Go rejects it with
      "certificate is valid for www.trackr.dev, not api.trackr.dev".
  - question: Why does `http.Get(srv.URL)` fail against an `httptest.NewTLSServer`, when `srv.Client().Get(srv.URL)` works?
    options:
      - text: The test server only accepts HTTP/2
      - text: The test server's certificate is self-signed by a test CA your system doesn't trust, and `srv.Client()` is configured to trust exactly that CA
        correct: true
      - text: '`http.Get` doesn''t support HTTPS'
    explanation: |
      Certificate verification works exactly as it would in production: the default
      client checks the system roots and doesn't find the test CA. `srv.Client()` has
      a transport whose `RootCAs` includes it.
---

TLS's third guarantee, "you're talking to the real `api.trackr.dev`", rests on
**certificates**.

## What a certificate says

A certificate is a signed statement:

> "The public key *K* belongs to `api.trackr.dev` and `*.trackr.dev`. Valid from
> 2026-06-01 to 2026-08-30. Signed, *Some Certificate Authority*."

During the handshake the server sends its certificate and proves it holds the private
key that matches *K*. The client then checks:

1. **Chain of trust.** Was it signed by a **Certificate Authority** (CA) the client
   trusts? Usually the server's certificate is signed by an intermediate CA, which is
   signed by a root CA. The client follows the chain up to a root in its trust store.
2. **Hostname.** Does the certificate list the name the client asked for?
3. **Dates.** Is today within the validity period?

Any failure aborts the connection before a single byte of HTTP is sent.

## Where the trusted roots come from

Go uses the operating system's root store: the system certificate bundle on Linux,
Keychain on macOS, and the system store on Windows. On Linux you can point Go at a
different bundle with the `SSL_CERT_FILE` or `SSL_CERT_DIR` environment variables.

Public CAs (Let's Encrypt and others) are in every root store. Companies often run
their own **private CA** for internal services, and its root has to be added
explicitly. Next lesson.

## Seeing it in Go

`httptest.NewTLSServer` (or `NewUnstartedServer` plus `StartTLS`) runs an HTTPS test
server with a certificate from a built-in test CA. That's perfect for watching
verification happen:

```go
package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
)

func main() {
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "secure hello")
	}))
	srv.Config.ErrorLog = log.New(io.Discard, "", 0) // keep the demo output tidy
	srv.StartTLS()
	defer srv.Close()

	// 1. An ordinary client doesn't trust the test server's certificate.
	_, err := http.Get(srv.URL)
	fmt.Println("default client:", err)

	// 2. srv.Client() trusts exactly that certificate.
	resp, err := srv.Client().Get(srv.URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	state := resp.TLS // non-nil for HTTPS responses
	cert := state.PeerCertificates[0]
	fmt.Println("status:      ", resp.Status)
	fmt.Println("version:     ", tls.VersionName(state.Version))
	fmt.Println("cipher:      ", tls.CipherSuiteName(state.CipherSuite))
	fmt.Println("key exchange:", state.CurveID)
	fmt.Println("issuer:      ", cert.Issuer.Organization)
	fmt.Println("valid for:   ", cert.DNSNames, cert.IPAddresses)
	fmt.Println("expires:     ", cert.NotAfter.Format("2006-01-02"))
}
```

Output (your port will differ):

```
default client: Get "https://127.0.0.1:42121": tls: failed to verify certificate: x509: certificate signed by unknown authority
status:       200 OK
version:      TLS 1.3
cipher:       TLS_AES_128_GCM_SHA256
key exchange: X25519MLKEM768
issuer:       [Acme Co]
valid for:    [example.com *.example.com] [127.0.0.1 ::1]
expires:      2084-01-29
```

- The default client **refused** the connection, as it should. It has never heard of
  "Acme Co", the test CA.
- `srv.Client()` returns an `*http.Client` whose transport trusts that test CA, so the
  request succeeds with full verification.
- `resp.TLS` is a `*tls.ConnectionState` describing the connection. There's TLS 1.3,
  and the post-quantum hybrid key exchange from the last lesson.
- The certificate is valid for `127.0.0.1`, which is why `srv.URL` works. Ask for
  `https://localhost:.../` instead and you get the hostname check failing:

```
x509: certificate is valid for example.com, *.example.com, not localhost
```

## Handling certificate errors

These errors mean "this might not be the server you think it is". The right response
is to **stop** and tell the user, never to retry without verification. You can detect
the specific cases with `errors.AsType`:

```go
if _, ok := errors.AsType[x509.UnknownAuthorityError](err); ok {
	return fmt.Errorf("can't verify %s's certificate (behind a corporate proxy? see `trackr help tls`): %w", host, err)
}
if _, ok := errors.AsType[x509.HostnameError](err); ok {
	return fmt.Errorf("certificate doesn't match %s: check the API URL: %w", host, err)
}
```

(Both are value types, not pointers, which is why the type parameter has no `*`.)

## Certificates expire

Public certificates now last months, not years, and they're renewed automatically. When
renewal breaks, clients start failing with `x509: certificate has expired or is not yet
valid`. That's the server's problem to fix, but it's good to recognise the message.
Also check your machine's clock: a badly wrong system time produces the same error.
