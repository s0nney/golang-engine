---
title: Mutual TLS
quiz:
  - question: 'A server sets `ClientAuth: tls.RequestClientCert`. What does it get?'
    options:
      - text: Clients must present a certificate signed by `ClientCAs`
      - text: Clients are asked for a certificate, but it isn't required *or verified*, so the handler can't trust it
        correct: true
      - text: The same as `RequireAndVerifyClientCert`
      - text: Clients must present any certificate, verified against the system roots
    explanation: |
      Only `VerifyClientCertIfGiven` and `RequireAndVerifyClientCert` verify the chain
      against `ClientCAs`. For "only enrolled devices may connect", use
      `RequireAndVerifyClientCert`.
exercise:
  starter: |
    package main

    import (
    	"crypto/tls"
    	"crypto/x509"
    	"fmt"
    	"net/http"
    )

    // serverTLSConfig configures Keybox's sync server: present serverCert, and
    // require every client to present a certificate that verifies against
    // deviceCAs. TLS 1.3 only.
    func serverTLSConfig(serverCert tls.Certificate, deviceCAs *x509.CertPool) *tls.Config {
    	// ?
    	return &tls.Config{Certificates: []tls.Certificate{serverCert}}
    }

    // clientTLSConfig configures a Keybox device: present deviceCert, trust
    // only roots, and expect the server to be serverName. TLS 1.3 only.
    func clientTLSConfig(deviceCert tls.Certificate, roots *x509.CertPool, serverName string) *tls.Config {
    	// ?
    	return &tls.Config{}
    }

    // deviceName returns the Common Name of the verified client certificate,
    // or "" if the request has no verified client certificate.
    func deviceName(r *http.Request) string {
    	// ?
    	return ""
    }

    func main() {
    	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "hello %s", deviceName(r))
    	})
    	fmt.Println("the tests start a real mTLS server with httptest and connect several devices")
    }
  solution: |
    package main

    import (
    	"crypto/tls"
    	"crypto/x509"
    	"fmt"
    	"net/http"
    )

    func serverTLSConfig(serverCert tls.Certificate, deviceCAs *x509.CertPool) *tls.Config {
    	return &tls.Config{
    		Certificates: []tls.Certificate{serverCert},
    		ClientAuth:   tls.RequireAndVerifyClientCert,
    		ClientCAs:    deviceCAs,
    		MinVersion:   tls.VersionTLS13,
    	}
    }

    func clientTLSConfig(deviceCert tls.Certificate, roots *x509.CertPool, serverName string) *tls.Config {
    	return &tls.Config{
    		Certificates: []tls.Certificate{deviceCert},
    		RootCAs:      roots,
    		ServerName:   serverName,
    		MinVersion:   tls.VersionTLS13,
    	}
    }

    func deviceName(r *http.Request) string {
    	if r.TLS == nil || len(r.TLS.VerifiedChains) == 0 || len(r.TLS.VerifiedChains[0]) == 0 {
    		return ""
    	}
    	return r.TLS.VerifiedChains[0][0].Subject.CommonName
    }

    func main() {
    	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "hello %s", deviceName(r))
    	})
    	fmt.Println("the tests start a real mTLS server with httptest and connect several devices")
    }
  tests: |
    package main

    import (
    	"crypto/ecdsa"
    	"crypto/elliptic"
    	"crypto/rand"
    	"crypto/tls"
    	"crypto/x509"
    	"crypto/x509/pkix"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"testing"
    	"time"
    )

    type ca struct {
    	cert *x509.Certificate
    	key  *ecdsa.PrivateKey
    	pool *x509.CertPool
    }

    func newTestCA(t *testing.T, name string) *ca {
    	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    	tmpl := &x509.Certificate{Subject: pkix.Name{CommonName: name}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
    		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
    	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
    	if err != nil {
    		t.Fatal(err)
    	}
    	cert, _ := x509.ParseCertificate(der)
    	pool := x509.NewCertPool()
    	pool.AddCert(cert)
    	return &ca{cert, key, pool}
    }

    func (c *ca) issue(t *testing.T, cn string, usage x509.ExtKeyUsage) tls.Certificate {
    	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    	tmpl := &x509.Certificate{Subject: pkix.Name{CommonName: cn}, DNSNames: []string{cn}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
    		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{usage}}
    	der, err := x509.CreateCertificate(rand.Reader, tmpl, c.cert, &key.PublicKey, c.key)
    	if err != nil {
    		t.Fatal(err)
    	}
    	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
    }

    func startServer(t *testing.T, serverCA, deviceCA *ca) *httptest.Server {
    	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "hello %s", deviceName(r))
    	}))
    	srv.TLS = serverTLSConfig(serverCA.issue(t, "sync.keybox.test", x509.ExtKeyUsageServerAuth), deviceCA.pool)
    	srv.StartTLS()
    	t.Cleanup(srv.Close)
    	return srv
    }

    func get(srv *httptest.Server, cfg *tls.Config) (string, error) {
    	client := &http.Client{Transport: &http.Transport{TLSClientConfig: cfg}, Timeout: 3 * time.Second}
    	resp, err := client.Get(srv.URL + "/hello")
    	if err != nil {
    		return "", err
    	}
    	defer resp.Body.Close()
    	body, err := io.ReadAll(resp.Body)
    	return string(body), err
    }

    func TestEnrolledDevice(t *testing.T) {
    	keybox := newTestCA(t, "Keybox Root")
    	srv := startServer(t, keybox, keybox)
    	cfg := clientTLSConfig(keybox.issue(t, "laptop-alice", x509.ExtKeyUsageClientAuth), keybox.pool, "sync.keybox.test")
    	body, err := get(srv, cfg)
    	if err != nil {
    		t.Fatalf("enrolled device: request failed: %v", err)
    	}
    	if body != "hello laptop-alice" {
    		t.Errorf("enrolled device: body = %q, want %q (deviceName should return the client cert's Common Name)", body, "hello laptop-alice")
    	}
    	if cfg.MinVersion != tls.VersionTLS13 || srv.TLS.MinVersion != tls.VersionTLS13 {
    		t.Errorf("MinVersion should be tls.VersionTLS13 on both sides")
    	}
    }

    func TestRejectedDevices(t *testing.T) {
    	keybox := newTestCA(t, "Keybox Root")
    	mallory := newTestCA(t, "Mallory Root")
    	srv := startServer(t, keybox, keybox)

    	noCert := clientTLSConfig(tls.Certificate{}, keybox.pool, "sync.keybox.test")
    	noCert.Certificates = nil
    	for name, cfg := range map[string]*tls.Config{
    		"no client certificate":        noCert,
    		"certificate from another CA":  clientTLSConfig(mallory.issue(t, "laptop-mallory", x509.ExtKeyUsageClientAuth), keybox.pool, "sync.keybox.test"),
    		"server-auth-only certificate": clientTLSConfig(keybox.issue(t, "laptop-eve", x509.ExtKeyUsageServerAuth), keybox.pool, "sync.keybox.test"),
    	} {
    		if body, err := get(srv, cfg); err == nil {
    			t.Errorf("%s: request succeeded with body %q; the server must require and verify client certificates", name, body)
    		}
    	}
    }

    func TestClientVerifiesServer(t *testing.T) {
    	keybox := newTestCA(t, "Keybox Root")
    	mallory := newTestCA(t, "Mallory Root")
    	srv := startServer(t, mallory, keybox) // server cert from the wrong CA
    	cfg := clientTLSConfig(keybox.issue(t, "laptop-alice", x509.ExtKeyUsageClientAuth), keybox.pool, "sync.keybox.test")
    	if _, err := get(srv, cfg); err == nil {
    		t.Errorf("client connected to a server whose certificate isn't from the Keybox root; set RootCAs")
    	}
    	good := startServer(t, keybox, keybox)
    	cfg = clientTLSConfig(keybox.issue(t, "laptop-alice", x509.ExtKeyUsageClientAuth), keybox.pool, "api.keybox.test")
    	if _, err := get(good, cfg); err == nil {
    		t.Errorf("client accepted a server certificate for sync.keybox.test while expecting api.keybox.test; set ServerName")
    	}
    }

    func TestDeviceNameWithoutTLS(t *testing.T) {
    	r := httptest.NewRequest("GET", "/hello", nil)
    	if got := deviceName(r); got != "" {
    		t.Errorf("deviceName(plain HTTP request) = %q, want \"\"", got)
    	}
    }
---

Ordinary TLS authenticates only the server. Keybox's sync server also wants to know
**which device** is connecting, and to refuse anything that isn't an enrolled Keybox
device before a single HTTP byte is parsed. That's **mutual TLS** (mTLS): the client
presents a certificate too.

## How it works

During the handshake the server sends a `CertificateRequest`. The client answers with
its certificate chain and signs the handshake transcript with its private key, the same
proof of possession the server gives. The server verifies the chain against its own
pool of trusted **client CAs**.

Keybox enrols a device like this: the device generates a key pair and a CSR, Alice
approves it in an already-enrolled app, and the Keybox device CA issues a certificate
with `CommonName: "laptop-alice"` and `ExtKeyUsageClientAuth`. From then on the device
connects with it. Service meshes and internal APIs use the same pattern to authenticate
machines to each other without passwords or API keys.

## Server configuration

```go
&tls.Config{
	Certificates: []tls.Certificate{serverCert},
	ClientAuth:   tls.RequireAndVerifyClientCert,
	ClientCAs:    deviceCAs, // trust only Keybox's device CA
	MinVersion:   tls.VersionTLS13,
}
```

`ClientAuth` has five levels. `RequireAndVerifyClientCert` is the one you want when
every client must be enrolled; `VerifyClientCertIfGiven` suits servers that also accept
other forms of login. The two "request" and "require any" levels don't verify the chain
at all, so a certificate they receive proves nothing.

Go also checks the client certificate's extended key usage: it must allow
`ClientAuth`, so a *server* certificate from the same CA can't be used to log in as a
device.

## Client configuration

```go
&tls.Config{
	Certificates: []tls.Certificate{deviceCert},
	RootCAs:      keyboxRoots, // trust only the Keybox root for the server
	ServerName:   "sync.keybox.test",
	MinVersion:   tls.VersionTLS13,
}
```

For HTTP, put it in `http.Transport{TLSClientConfig: cfg}`. The device still verifies
the *server* normally; mutual means both directions.

## Who's calling?

After a successful handshake, `r.TLS.VerifiedChains` holds the chains Go verified. The
first certificate of the first chain is the client's own:

```go
leaf := r.TLS.VerifiedChains[0][0]
device := leaf.Subject.CommonName
```

Use `VerifiedChains`, not `PeerCertificates`. Both contain the client's leaf, but only
`VerifiedChains` is guaranteed to be verified, and if someone later changes `ClientAuth`
to a weaker level, `PeerCertificates` will happily hold unverified certificates. Guard
against `r.TLS` being `nil` (plain HTTP, or a test request) and against empty chains.

## Operating mTLS

- **Short-lived client certificates** (hours to days), renewed automatically, make
  revocation mostly unnecessary: removing a device just means not renewing it.
- **Separate CAs** for servers and devices, so a device certificate can never pass for
  a server, even if usages were misconfigured.
- **Authorization is still your job.** mTLS tells you *which device* is calling. Whether
  `laptop-alice` may read the team vault is a separate check, like the owner checks in
  [Learn HTTP Servers](/courses/learn-http-servers/authorization-and-webhooks/owner-only-actions).

## Your task

1. `serverTLSConfig(serverCert, deviceCAs)`: present `serverCert`, set
   `ClientAuth: tls.RequireAndVerifyClientCert` and `ClientCAs: deviceCAs`, and
   `MinVersion: tls.VersionTLS13`.
2. `clientTLSConfig(deviceCert, roots, serverName)`: present `deviceCert`, set
   `RootCAs` and `ServerName`, and TLS 1.3 minimum.
3. `deviceName(r)`: return the Common Name of `r.TLS.VerifiedChains[0][0]`, or `""` if
   `r.TLS` is nil or there's no verified chain.

The tests start a real HTTPS server on the loopback interface with
`httptest.NewUnstartedServer` and `StartTLS`, then connect an enrolled laptop, a device
with no certificate, one from Mallory's CA, and one with the wrong key usage.
