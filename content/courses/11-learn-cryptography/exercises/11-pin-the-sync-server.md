---
title: Pin the Sync Server
difficulty: medium
after: certificates-and-tls
hints:
  - '`spkiPin` hashes `cert.RawSubjectPublicKeyInfo` (the DER of the public key, not the whole certificate) with `sha256.Sum256` and returns `"sha256/" + base64.StdEncoding.EncodeToString(sum[:])`. Pinning the key rather than the certificate means routine renewals with the same key keep working.'
  - 'Keep `RootCAs` and `ServerName`, and add a `VerifyConnection` callback. Go calls it only *after* the chain and hostname have been verified; in it, compute the pin of `cs.PeerCertificates[0]` and return an error unless `slices.Contains(pins, thatPin)`. Never reach for `InsecureSkipVerify`.'
exercise:
  starter: |
    package main

    import (
    	"crypto/tls"
    	"crypto/x509"
    	"errors"
    	"fmt"
    	"io"
    	"log"
    	"net/http"
    	"net/http/httptest"
    )

    var ErrPinMismatch = errors.New("keybox: server key doesn't match any pin")

    // spkiPin returns the pin for cert's public key.
    func spkiPin(cert *x509.Certificate) string {
    	return ""
    }

    // pinnedConfig returns a client TLS config that verifies the server normally
    // and then also requires its leaf key to match one of pins.
    func pinnedConfig(roots *x509.CertPool, serverName string, pins []string) *tls.Config {
    	return &tls.Config{RootCAs: roots, ServerName: serverName}
    }

    func main() {
    	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "vault synced")
    	}))
    	ts.Config.ErrorLog = log.New(io.Discard, "", 0)
    	ts.StartTLS()
    	defer ts.Close()

    	roots := x509.NewCertPool()
    	roots.AddCert(ts.Certificate())
    	pin := spkiPin(ts.Certificate())
    	fmt.Println("server pin:", pin) // sha256/<44 base64 chars>

    	for _, pins := range [][]string{{pin}, {"sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}} {
    		client := &http.Client{Transport: &http.Transport{TLSClientConfig: pinnedConfig(roots, "example.com", pins)}}
    		resp, err := client.Get(ts.URL)
    		if err != nil {
    			fmt.Println("refused:", err) // want: the second (wrong) pin is refused
    			continue
    		}
    		resp.Body.Close()
    		fmt.Println("connected with pins", pins)
    	}
    }
  solution: |
    package main

    import (
    	"crypto/sha256"
    	"crypto/tls"
    	"crypto/x509"
    	"encoding/base64"
    	"errors"
    	"fmt"
    	"io"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    )

    var ErrPinMismatch = errors.New("keybox: server key doesn't match any pin")

    // spkiPin returns the pin for cert's public key.
    func spkiPin(cert *x509.Certificate) string {
    	sum := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
    	return "sha256/" + base64.StdEncoding.EncodeToString(sum[:])
    }

    // pinnedConfig returns a client TLS config that verifies the server normally
    // and then also requires its leaf key to match one of pins.
    func pinnedConfig(roots *x509.CertPool, serverName string, pins []string) *tls.Config {
    	pins = slices.Clone(pins)
    	return &tls.Config{
    		RootCAs:    roots,
    		ServerName: serverName,
    		MinVersion: tls.VersionTLS13,
    		VerifyConnection: func(cs tls.ConnectionState) error {
    			if len(cs.PeerCertificates) == 0 {
    				return ErrPinMismatch
    			}
    			if !slices.Contains(pins, spkiPin(cs.PeerCertificates[0])) {
    				return ErrPinMismatch
    			}
    			return nil
    		},
    	}
    }

    func main() {
    	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "vault synced")
    	}))
    	ts.Config.ErrorLog = log.New(io.Discard, "", 0)
    	ts.StartTLS()
    	defer ts.Close()

    	roots := x509.NewCertPool()
    	roots.AddCert(ts.Certificate())
    	pin := spkiPin(ts.Certificate())
    	fmt.Println("server pin:", pin)

    	for _, pins := range [][]string{{pin}, {"sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}} {
    		client := &http.Client{Transport: &http.Transport{TLSClientConfig: pinnedConfig(roots, "example.com", pins)}}
    		resp, err := client.Get(ts.URL)
    		if err != nil {
    			fmt.Println("refused:", err)
    			continue
    		}
    		resp.Body.Close()
    		fmt.Println("connected with pins", pins)
    	}
    }
  tests: |
    package main

    import (
    	"crypto/ecdsa"
    	"crypto/elliptic"
    	"crypto/rand"
    	"crypto/sha256"
    	"crypto/tls"
    	"crypto/x509"
    	"crypto/x509/pkix"
    	"encoding/base64"
    	"fmt"
    	"io"
    	"log"
    	"math/big"
    	"net"
    	"net/http"
    	"net/http/httptest"
    	"testing"
    	"time"
    )

    func startServer(t *testing.T, cert *tls.Certificate) *httptest.Server {
    	t.Helper()
    	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprint(w, "vault synced")
    	}))
    	ts.Config.ErrorLog = log.New(io.Discard, "", 0)
    	if cert != nil {
    		ts.TLS = &tls.Config{Certificates: []tls.Certificate{*cert}}
    	}
    	ts.StartTLS()
    	t.Cleanup(ts.Close)
    	return ts
    }

    // caIssued returns a CA certificate and a leaf for example.com signed by it,
    // with a different key from httptest's built-in certificate.
    func caIssued(t *testing.T) (*x509.Certificate, *tls.Certificate) {
    	t.Helper()
    	caKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    	now := time.Now()
    	caTmpl := &x509.Certificate{
    		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Test CA"},
    		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
    		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
    	}
    	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
    	if err != nil {
    		t.Fatal(err)
    	}
    	ca, _ := x509.ParseCertificate(caDER)
    	leafKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    	leafTmpl := &x509.Certificate{
    		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "example.com"},
    		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
    		DNSNames: []string{"example.com"}, IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
    		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
    	}
    	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, ca, &leafKey.PublicKey, caKey)
    	if err != nil {
    		t.Fatal(err)
    	}
    	return ca, &tls.Certificate{Certificate: [][]byte{leafDER}, PrivateKey: leafKey}
    }

    func connect(t *testing.T, ts *httptest.Server, cfg *tls.Config) error {
    	t.Helper()
    	if cfg == nil {
    		t.Fatal("pinnedConfig returned nil")
    	}
    	tr := &http.Transport{TLSClientConfig: cfg}
    	defer tr.CloseIdleConnections()
    	resp, err := (&http.Client{Transport: tr}).Get(ts.URL)
    	if err != nil {
    		return err
    	}
    	resp.Body.Close()
    	return nil
    }

    func TestSPKIPin(t *testing.T) {
    	ts := startServer(t, nil)
    	cert := ts.Certificate()
    	sum := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
    	want := "sha256/" + base64.StdEncoding.EncodeToString(sum[:])
    	if got := spkiPin(cert); got != want {
    		t.Errorf("spkiPin(httptest certificate) = %q, want %q (SHA-256 of RawSubjectPublicKeyInfo, standard base64)", got, want)
    	}
    }

    func TestPinnedServerConnects(t *testing.T) {
    	ts := startServer(t, nil)
    	roots := x509.NewCertPool()
    	roots.AddCert(ts.Certificate())
    	pin := spkiPin(ts.Certificate())
    	if err := connect(t, ts, pinnedConfig(roots, "example.com", []string{pin})); err != nil {
    		t.Errorf("connecting with the server's own pin failed: %v", err)
    	}
    	backup := "sha256/" + base64.StdEncoding.EncodeToString(make([]byte, 32))
    	if err := connect(t, ts, pinnedConfig(roots, "example.com", []string{backup, pin})); err != nil {
    		t.Errorf("connecting with [backup pin, current pin] failed: %v (any listed pin should do)", err)
    	}
    }

    func TestWrongPinRefused(t *testing.T) {
    	ts := startServer(t, nil)
    	roots := x509.NewCertPool()
    	roots.AddCert(ts.Certificate())
    	other := "sha256/" + base64.StdEncoding.EncodeToString(make([]byte, 32))
    	for _, pins := range [][]string{{other}, nil, {""}} {
    		if err := connect(t, ts, pinnedConfig(roots, "example.com", pins)); err == nil {
    			t.Errorf("connected with pins %q, which don't include the server's key", pins)
    		}
    	}
    }

    func TestTrustedCAButUnpinnedKeyRefused(t *testing.T) {
    	// The attack pinning exists for: a CA you trust issues a valid certificate
    	// for your hostname, with someone else's key.
    	genuine := startServer(t, nil)
    	ca, rogueCert := caIssued(t)
    	rogue := startServer(t, rogueCert)
    	roots := x509.NewCertPool()
    	roots.AddCert(genuine.Certificate())
    	roots.AddCert(ca)
    	cfg := pinnedConfig(roots, "example.com", []string{spkiPin(genuine.Certificate())})
    	if err := connect(t, genuine, cfg); err != nil {
    		t.Fatalf("connecting to the genuine server failed: %v", err)
    	}
    	if err := connect(t, rogue, pinnedConfig(roots, "example.com", []string{spkiPin(genuine.Certificate())})); err == nil {
    		t.Errorf("connected to a server with a CA-valid certificate but an unpinned key")
    	}
    }

    func TestNormalVerificationStillRuns(t *testing.T) {
    	ts := startServer(t, nil)
    	pin := spkiPin(ts.Certificate())
    	if err := connect(t, ts, pinnedConfig(x509.NewCertPool(), "example.com", []string{pin})); err == nil {
    		t.Errorf("connected with an empty root pool: the pin must be checked in addition to the chain, not instead of it")
    	}
    	roots := x509.NewCertPool()
    	roots.AddCert(ts.Certificate())
    	if err := connect(t, ts, pinnedConfig(roots, "evil.example", []string{pin})); err == nil {
    		t.Errorf("connected with ServerName evil.example: the hostname must still be checked")
    	}
    	cfg := pinnedConfig(roots, "example.com", []string{pin})
    	if cfg.InsecureSkipVerify {
    		t.Errorf("InsecureSkipVerify is set: use VerifyConnection, which runs after normal verification")
    	}
    }
---

Keybox's sync server has a certificate from a public CA. But *any* of the CAs in a
trust store can issue a certificate for `sync.keybox.test`, and CAs have been tricked
or compromised before. Because Keybox ships both the client and the server, it can
go further and **pin** the server's public key: a connection succeeds only if the
certificate is valid *and* carries one of a few known keys.

Implement:

- `spkiPin(cert)` returns `"sha256/"` followed by the standard base64 encoding of the
  SHA-256 of `cert.RawSubjectPublicKeyInfo`.
- `pinnedConfig(roots, serverName, pins)` returns a client `*tls.Config` that
  - does all the normal verification against `roots` for `serverName` (TLS 1.3), and
  - **additionally** refuses the connection unless the pin of the server's leaf
    certificate is in `pins`. An empty `pins` refuses everything.

## Example

```go
pin := spkiPin(ts.Certificate())       // "sha256/OJ+e3lINvDPSrrxIkkatieIh0ewV9pPDSMWLCCGTZ6o="
pinnedConfig(roots, "example.com", []string{pin})           // connects
pinnedConfig(roots, "example.com", []string{backup, pin})   // connects (any pin matches)
pinnedConfig(roots, "example.com", []string{backup})        // refused: key doesn't match
pinnedConfig(emptyPool, "example.com", []string{pin})       // refused: chain untrusted
```

## Constraints

- The tests run real TLS servers in-process with `httptest`, including one whose
  certificate is issued by a CA the client trusts, but with a different key. That's
  exactly what pinning must stop.
- The pin is checked **in addition to** normal verification: an untrusted chain or a
  wrong hostname must still fail, even with a matching pin.
- Always ship at least one **backup pin** (a key you've generated but not deployed),
  or losing the server's key locks every client out.
