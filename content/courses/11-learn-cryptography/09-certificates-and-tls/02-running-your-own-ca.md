---
title: Running Your Own CA
quiz:
  - question: 'Why should Keybox''s root CA certificate set `MaxPathLenZero: true` (with `MaxPathLen: 0`)?'
    options:
      - text: It makes the certificate smaller
      - text: It means certificates the root signs can't themselves act as CAs, so a leaked leaf key can't mint more certificates
        correct: true
      - text: It stops the root from expiring
      - text: It's required for ECDSA keys
    explanation: |
      Path length constraints limit how deep a chain under this CA may go. With 0, the
      root can only sign end-entity (leaf) certificates. Keybox's leaves also have
      `IsCA: false`, so either way they can't issue.
exercise:
  starter: |
    package main

    import (
    	"crypto"
    	"crypto/ecdsa"
    	"crypto/elliptic"
    	"crypto/rand"
    	"crypto/x509"
    	"crypto/x509/pkix"
    	"errors"
    	"fmt"
    	"time"
    )

    // CA is Keybox's private certificate authority.
    type CA struct {
    	Cert *x509.Certificate
    	Key  *ecdsa.PrivateKey
    }

    // newCA creates a self-signed root: CommonName name, valid from now-1h to
    // now+10 years, IsCA with BasicConstraintsValid, MaxPathLen 0 with
    // MaxPathLenZero, and KeyUsage CertSign|CRLSign. Its key is ECDSA P-256.
    func newCA(name string, now time.Time) (*CA, error) {
    	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    	if err != nil {
    		return nil, err
    	}
    	tmpl := &x509.Certificate{
    		Subject:   pkix.Name{CommonName: name},
    		NotBefore: now.Add(-time.Hour),
    		NotAfter:  now.AddDate(10, 0, 0),
    		// ? Mark it as a CA that may only sign leaf certificates.
    	}
    	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
    	if err != nil {
    		return nil, err
    	}
    	cert, err := x509.ParseCertificate(der)
    	if err != nil {
    		return nil, err
    	}
    	return &CA{Cert: cert, Key: key}, nil
    }

    // issueServer signs a server certificate for host and pub: CommonName and
    // DNSNames = host, valid from now-1h to now+90 days, KeyUsage
    // DigitalSignature, ExtKeyUsage ServerAuth, not a CA.
    func (ca *CA) issueServer(host string, pub crypto.PublicKey, now time.Time) (*x509.Certificate, error) {
    	// ?
    	return nil, errors.New("not implemented")
    }

    func main() {
    	now := time.Now()
    	ca, err := newCA("Keybox Root CA", now)
    	if err != nil {
    		fmt.Println(err)
    		return
    	}
    	serverKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    	leaf, err := ca.issueServer("sync.keybox.test", &serverKey.PublicKey, now)
    	if err != nil {
    		fmt.Println("issue:", err)
    		return
    	}
    	roots := x509.NewCertPool()
    	roots.AddCert(ca.Cert)
    	_, err = leaf.Verify(x509.VerifyOptions{Roots: roots, DNSName: "sync.keybox.test", CurrentTime: now})
    	fmt.Println("verify:", err)
    }
  solution: |
    package main

    import (
    	"crypto"
    	"crypto/ecdsa"
    	"crypto/elliptic"
    	"crypto/rand"
    	"crypto/x509"
    	"crypto/x509/pkix"
    	"fmt"
    	"time"
    )

    type CA struct {
    	Cert *x509.Certificate
    	Key  *ecdsa.PrivateKey
    }

    func newCA(name string, now time.Time) (*CA, error) {
    	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    	if err != nil {
    		return nil, err
    	}
    	tmpl := &x509.Certificate{
    		Subject:               pkix.Name{CommonName: name},
    		NotBefore:             now.Add(-time.Hour),
    		NotAfter:              now.AddDate(10, 0, 0),
    		IsCA:                  true,
    		BasicConstraintsValid: true,
    		MaxPathLen:            0,
    		MaxPathLenZero:        true,
    		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
    	}
    	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
    	if err != nil {
    		return nil, err
    	}
    	cert, err := x509.ParseCertificate(der)
    	if err != nil {
    		return nil, err
    	}
    	return &CA{Cert: cert, Key: key}, nil
    }

    func (ca *CA) issueServer(host string, pub crypto.PublicKey, now time.Time) (*x509.Certificate, error) {
    	tmpl := &x509.Certificate{
    		Subject:               pkix.Name{CommonName: host},
    		DNSNames:              []string{host},
    		NotBefore:             now.Add(-time.Hour),
    		NotAfter:              now.Add(90 * 24 * time.Hour),
    		KeyUsage:              x509.KeyUsageDigitalSignature,
    		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
    		BasicConstraintsValid: true,
    		IsCA:                  false,
    	}
    	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.Cert, pub, ca.Key)
    	if err != nil {
    		return nil, err
    	}
    	return x509.ParseCertificate(der)
    }

    func main() {
    	now := time.Now()
    	ca, err := newCA("Keybox Root CA", now)
    	if err != nil {
    		fmt.Println(err)
    		return
    	}
    	serverKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    	leaf, err := ca.issueServer("sync.keybox.test", &serverKey.PublicKey, now)
    	if err != nil {
    		fmt.Println("issue:", err)
    		return
    	}
    	roots := x509.NewCertPool()
    	roots.AddCert(ca.Cert)
    	_, err = leaf.Verify(x509.VerifyOptions{Roots: roots, DNSName: "sync.keybox.test", CurrentTime: now})
    	fmt.Println("verify:", err)
    }
  tests: |
    package main

    import (
    	"crypto/ecdsa"
    	"crypto/elliptic"
    	"crypto/rand"
    	"crypto/x509"
    	"crypto/x509/pkix"
    	"slices"
    	"testing"
    	"time"
    )

    var testNow = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

    func setup(t *testing.T) (*CA, *x509.Certificate, *ecdsa.PrivateKey, *x509.CertPool) {
    	t.Helper()
    	ca, err := newCA("Keybox Test Root", testNow)
    	if err != nil {
    		t.Fatalf("newCA error = %v", err)
    	}
    	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    	leaf, err := ca.issueServer("sync.keybox.test", &key.PublicKey, testNow)
    	if err != nil {
    		t.Fatalf("issueServer error = %v", err)
    	}
    	roots := x509.NewCertPool()
    	roots.AddCert(ca.Cert)
    	return ca, leaf, key, roots
    }

    func TestCAFields(t *testing.T) {
    	ca, _, _, _ := setup(t)
    	c := ca.Cert
    	if !c.IsCA || !c.BasicConstraintsValid {
    		t.Errorf("CA certificate: IsCA=%v BasicConstraintsValid=%v, want both true", c.IsCA, c.BasicConstraintsValid)
    	}
    	if c.MaxPathLen != 0 || !c.MaxPathLenZero {
    		t.Errorf("CA certificate: MaxPathLen=%d MaxPathLenZero=%v, want 0 and true", c.MaxPathLen, c.MaxPathLenZero)
    	}
    	if c.KeyUsage&x509.KeyUsageCertSign == 0 {
    		t.Errorf("CA certificate KeyUsage = %v, must include certSign", c.KeyUsage)
    	}
    }

    func TestLeafVerifies(t *testing.T) {
    	_, leaf, _, roots := setup(t)
    	if leaf.IsCA {
    		t.Errorf("server certificate has IsCA = true")
    	}
    	if !slices.Equal(leaf.DNSNames, []string{"sync.keybox.test"}) {
    		t.Errorf("server certificate DNSNames = %v, want [sync.keybox.test]", leaf.DNSNames)
    	}
    	for _, d := range []time.Duration{0, 24 * time.Hour, 89 * 24 * time.Hour} {
    		_, err := leaf.Verify(x509.VerifyOptions{Roots: roots, DNSName: "sync.keybox.test", CurrentTime: testNow.Add(d)})
    		if err != nil {
    			t.Errorf("Verify at issue time + %v: %v", d, err)
    		}
    	}
    }

    func TestLeafRejected(t *testing.T) {
    	_, leaf, _, roots := setup(t)
    	other, _ := newCA("Some Other CA", testNow)
    	otherRoots := x509.NewCertPool()
    	otherRoots.AddCert(other.Cert)
    	for name, opts := range map[string]x509.VerifyOptions{
    		"wrong host":           {Roots: roots, DNSName: "evil.test", CurrentTime: testNow},
    		"expired":              {Roots: roots, DNSName: "sync.keybox.test", CurrentTime: testNow.Add(91 * 24 * time.Hour)},
    		"not yet valid":        {Roots: roots, DNSName: "sync.keybox.test", CurrentTime: testNow.Add(-2 * time.Hour)},
    		"untrusted CA":         {Roots: otherRoots, DNSName: "sync.keybox.test", CurrentTime: testNow},
    		"used for client auth": {Roots: roots, DNSName: "sync.keybox.test", CurrentTime: testNow, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}},
    	} {
    		if _, err := leaf.Verify(opts); err == nil {
    			t.Errorf("%s: Verify succeeded, want an error", name)
    		}
    	}
    }

    func TestLeafCannotIssue(t *testing.T) {
    	_, leaf, leafKey, roots := setup(t)
    	tmpl := &x509.Certificate{
    		Subject:   pkix.Name{CommonName: "mallory.test"},
    		DNSNames:  []string{"mallory.test"},
    		NotBefore: testNow.Add(-time.Hour),
    		NotAfter:  testNow.Add(time.Hour),
    	}
    	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    	der, err := x509.CreateCertificate(rand.Reader, tmpl, leaf, &key.PublicKey, leafKey)
    	if err != nil {
    		return // refusing to sign is fine too
    	}
    	forged, _ := x509.ParseCertificate(der)
    	inter := x509.NewCertPool()
    	inter.AddCert(leaf)
    	if _, err := forged.Verify(x509.VerifyOptions{Roots: roots, Intermediates: inter, DNSName: "mallory.test", CurrentTime: testNow}); err == nil {
    		t.Errorf("a certificate signed by the server's key verified; the server certificate must not be a CA")
    	}
    }
---

Keybox's sync servers talk only to Keybox clients, so there's no need to pay a public
CA or depend on one. Keybox runs its own **private CA**: a root certificate baked into
every client, and a signing key kept offline.

## The root

A CA certificate is an ordinary certificate with a few extra fields:

```go
tmpl := &x509.Certificate{
	Subject:               pkix.Name{CommonName: "Keybox Root CA"},
	NotBefore:             now.Add(-time.Hour),
	NotAfter:              now.AddDate(10, 0, 0),
	IsCA:                  true,
	BasicConstraintsValid: true,
	MaxPathLen:            0,
	MaxPathLenZero:        true,
	KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
}
der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
```

- **`IsCA` with `BasicConstraintsValid`** marks it as allowed to sign certificates.
  Without the basic constraints extension, verifiers won't accept it as an issuer.
- **`MaxPathLen: 0` with `MaxPathLenZero: true`** means "no CAs below me". (A plain
  `MaxPathLen: 0` means "unset", because 0 is also Go's zero value; that's why the
  second field exists. It's a classic Go gotcha.)
- **`KeyUsageCertSign`** is the key usage that permits signing certificates.
- **Backdate `NotBefore` slightly** to tolerate clocks that run a little slow.

Public CAs keep their roots offline and sign day-to-day certificates with an
**intermediate** CA, so a compromised intermediate can be revoked without replacing
the root in every client. Keybox's root is small enough to skip that, which is also why
it can forbid all sub-CAs.

## Issuing a server certificate

A leaf is signed by passing the **CA's certificate as parent** and the **CA's private
key** as the signer, along with the *server's* public key:

```go
der, err := x509.CreateCertificate(rand.Reader, leafTmpl, ca.Cert, serverPublicKey, ca.Key)
```

The leaf template sets what the server may do: `DNSNames` for the names it serves,
`KeyUsageDigitalSignature` (TLS 1.3 servers sign the handshake), and
`ExtKeyUsageServerAuth`. Setting `BasicConstraintsValid: true` with `IsCA: false`
states explicitly that it's not a CA. Keep leaf lifetimes short, 90 days or less, and
automate renewal; a short-lived certificate limits the damage of a stolen server key
without relying on revocation, which is notoriously unreliable.

Notice the server's private key never goes near the CA. In production the server
generates its key pair and sends a **certificate signing request** (CSR,
`x509.CreateCertificateRequest`); the CA checks it and returns a certificate.

## Checking the result

`leaf.Verify` builds and checks a chain to the roots you trust:

```go
roots := x509.NewCertPool()
roots.AddCert(ca.Cert)
chains, err := leaf.Verify(x509.VerifyOptions{
	Roots:       roots,
	DNSName:     "sync.keybox.test",
	CurrentTime: now, // zero means time.Now()
})
```

It checks every signature, validity periods, the hostname, the CA constraints and
(by default) that the leaf allows server authentication. Next lesson looks at it
more closely.

## Your task

1. Finish `newCA`: add `IsCA`, `BasicConstraintsValid`, `MaxPathLen: 0` with
   `MaxPathLenZero: true`, and `KeyUsage: CertSign | CRLSign` to the template.
2. Write `ca.issueServer(host, pub, now)`: a template with `CommonName` and `DNSNames`
   set to `host`, valid from `now - 1h` to `now + 90 days`,
   `KeyUsageDigitalSignature`, `ExtKeyUsageServerAuth`, and
   `BasicConstraintsValid: true` with `IsCA: false`. Sign it with
   `x509.CreateCertificate(rand.Reader, tmpl, ca.Cert, pub, ca.Key)` and return the
   parsed certificate.

The tests verify your leaf, then try wrong hostnames, expired times, another CA, the
wrong usage, and using the leaf's key to mint a certificate of its own.
