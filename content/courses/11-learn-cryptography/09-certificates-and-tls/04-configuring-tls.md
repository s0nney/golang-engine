---
title: Configuring crypto/tls
quiz:
  - question: What's the right way to choose TLS 1.3 cipher suites in a Go server?
    options:
      - text: List them in `tls.Config.CipherSuites`, strongest first
      - text: You don't; Go ignores `CipherSuites` for TLS 1.3 and picks from a small, fixed set of secure suites
        correct: true
      - text: 'Set `PreferServerCipherSuites: true`'
      - text: Disable AES to force ChaCha20
    explanation: |
      All TLS 1.3 suites are AEADs and all are safe, so Go doesn't make them
      configurable. `CipherSuites` only affects TLS 1.2 and below, and
      `PreferServerCipherSuites` has been ignored since Go 1.18.
  - question: What key exchange does a default Go 1.27 client and server negotiate with each other?
    options:
      - text: RSA key transport
      - text: Plain X25519
      - text: The hybrid post-quantum X25519MLKEM768
        correct: true
      - text: Finite-field Diffie-Hellman
    explanation: |
      Since Go 1.24, the default `CurvePreferences` includes X25519MLKEM768, and two Go
      endpoints prefer it. `ConnectionState().CurveID` reports what was used.
---

You've built a CA and verified chains. Now let's put certificates to work in
`crypto/tls`, the package behind every `https://` URL your Go programs touch.

## A handshake without a network

`tls.Server` and `tls.Client` wrap any `net.Conn`. `net.Pipe` gives you an in-memory
pair of connected conns, perfect for seeing TLS in isolation:

```go
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"time"
)

func main() {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{
		Subject:     pkix.Name{CommonName: "sync.keybox.test"},
		DNSNames:    []string{"sync.keybox.test"},
		NotBefore:   time.Now().Add(-time.Minute),
		NotAfter:    time.Now().Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	cert, _ := x509.ParseCertificate(der)
	roots := x509.NewCertPool()
	roots.AddCert(cert) // the client pins this self-signed certificate

	clientConn, serverConn := net.Pipe()
	go func() {
		srv := tls.Server(serverConn, &tls.Config{
			Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
			MinVersion:   tls.VersionTLS13,
		})
		srv.Write([]byte("hello from keybox"))
		srv.Close()
	}()

	cli := tls.Client(clientConn, &tls.Config{RootCAs: roots, ServerName: "sync.keybox.test", MinVersion: tls.VersionTLS13})
	if err := cli.Handshake(); err != nil {
		fmt.Println("handshake failed:", err)
		return
	}
	st := cli.ConnectionState()
	fmt.Println(tls.VersionName(st.Version), tls.CipherSuiteName(st.CipherSuite), st.CurveID)
	buf := make([]byte, 64)
	n, _ := cli.Read(buf)
	fmt.Printf("%q\n", buf[:n])
}
```

```
TLS 1.3 TLS_AES_128_GCM_SHA256 X25519MLKEM768
"hello from keybox"
```

(On a CPU without AES instructions, the suite may be `TLS_CHACHA20_POLY1305_SHA256`
instead; Go picks what's fast *and* constant-time on your hardware.)

Everything from this course shows up in that one line: an AEAD (AES-GCM) for records,
HKDF-SHA256 for the key schedule, a hybrid post-quantum KEM for key exchange, and an
ECDSA signature from the server's certificate key proving it holds the key.

## The fields that matter

**Server side:**

- `Certificates`: the chain (leaf first, then intermediates) and private key. Load
  from PEM files with `tls.LoadX509KeyPair(certFile, keyFile)`.
- `GetCertificate`: pick a certificate per connection (by SNI name), or reload
  renewed certificates without restarting.
- `ClientAuth` and `ClientCAs`: request and verify client certificates (next lesson).

**Client side:**

- `RootCAs`: roots to trust; `nil` means the system's roots. Use a private pool for a
  private CA, as above.
- `ServerName`: the hostname to verify. `http.Client` sets it from the URL for you;
  with raw `tls.Client` you must set it, or the handshake fails.
- `InsecureSkipVerify`: never in production. See
  [Never InsecureSkipVerify](/courses/learn-http-clients/https-and-security/never-insecure-skip-verify).

**Both:**

- `MinVersion`: the default minimum is TLS 1.2. Set `tls.VersionTLS13` when you
  control both ends, as Keybox does.
- `CurvePreferences`: leave it nil to get the defaults, including post-quantum
  hybrids. Setting it explicitly *replaces* the defaults.
- `CipherSuites`: only affects TLS 1.2. The default list is already restricted to
  modern AEAD suites with forward secrecy.
- `KeyLogWriter`: writes session secrets for debugging with Wireshark. It breaks the
  security of every logged connection; never enable it in production builds.

## Defaults are the configuration

Go's TLS defaults are chosen by the Go security team and updated with each release:
insecure versions, suites and curves get removed, post-quantum key exchange gets added,
and your program improves with nothing more than a Go upgrade. Most "hardening" guides
written for other stacks make Go *worse* by pinning yesterday's choices. Set what you
need (certificates, roots, `MinVersion`, client auth), and leave the rest alone.

## Testing HTTPS servers

For HTTP, `httptest.NewTLSServer` starts a server with a test certificate, and
`srv.Client()` returns a client that trusts it. For custom TLS settings, create it with
`httptest.NewUnstartedServer`, set `srv.TLS` to your `*tls.Config`, and call
`srv.StartTLS()`. That's exactly how the next exercise tests Keybox's mutual TLS.
