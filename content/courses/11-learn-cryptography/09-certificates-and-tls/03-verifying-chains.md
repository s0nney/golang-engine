---
title: Verifying Chains
quiz:
  - question: 'A custom `VerifyPeerCertificate` callback parses the server''s leaf certificate and accepts it if `leaf.Subject.CommonName == "sync.keybox.test"`, with `InsecureSkipVerify: true` set so the callback runs instead of normal verification. What can Mallory do?'
    options:
      - text: Nothing, she can't get a certificate with that name
      - text: Create her own self-signed certificate with that Common Name and impersonate the server
        correct: true
      - text: Only read traffic, not modify it
      - text: Only attack if the root CA is expired
    explanation: |
      Anyone can put any name in a certificate they sign themselves. Without checking
      the signature chain back to a trusted root, the name means nothing. If you
      customize verification, call `x509.Certificate.Verify` with proper options
      yourself.
exercise:
  starter: |
    package main

    import (
    	"crypto/x509"
    	"errors"
    	"fmt"
    	"time"
    )

    // verifyServerChain checks the certificates a server presented (DER, leaf
    // first, then any intermediates) against roots, for host at time now, for
    // server authentication. It returns the parsed leaf on success.
    //
    // BUG: this version only compares the Common Name.
    func verifyServerChain(rawCerts [][]byte, roots *x509.CertPool, host string, now time.Time) (*x509.Certificate, error) {
    	if len(rawCerts) == 0 {
    		return nil, errors.New("no certificates")
    	}
    	leaf, err := x509.ParseCertificate(rawCerts[0])
    	if err != nil {
    		return nil, err
    	}
    	if leaf.Subject.CommonName != host {
    		return nil, fmt.Errorf("certificate is for %q, not %q", leaf.Subject.CommonName, host)
    	}
    	return leaf, nil
    }

    func main() {
    	_, err := verifyServerChain(nil, x509.NewCertPool(), "sync.keybox.test", time.Now())
    	fmt.Println("empty chain:", err)
    	fmt.Println("(the tests build real chains: root -> intermediate -> leaf)")
    }
  solution: |
    package main

    import (
    	"crypto/x509"
    	"errors"
    	"fmt"
    	"time"
    )

    func verifyServerChain(rawCerts [][]byte, roots *x509.CertPool, host string, now time.Time) (*x509.Certificate, error) {
    	if len(rawCerts) == 0 {
    		return nil, errors.New("no certificates")
    	}
    	leaf, err := x509.ParseCertificate(rawCerts[0])
    	if err != nil {
    		return nil, err
    	}
    	intermediates := x509.NewCertPool()
    	for _, raw := range rawCerts[1:] {
    		c, err := x509.ParseCertificate(raw)
    		if err != nil {
    			return nil, err
    		}
    		intermediates.AddCert(c)
    	}
    	_, err = leaf.Verify(x509.VerifyOptions{
    		Roots:         roots,
    		Intermediates: intermediates,
    		DNSName:       host,
    		CurrentTime:   now,
    		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
    	})
    	if err != nil {
    		return nil, err
    	}
    	return leaf, nil
    }

    func main() {
    	_, err := verifyServerChain(nil, x509.NewCertPool(), "sync.keybox.test", time.Now())
    	fmt.Println("empty chain:", err)
    	fmt.Println("(the tests build real chains: root -> intermediate -> leaf)")
    }
  tests: |
    package main

    import (
    	"crypto/ecdsa"
    	"crypto/elliptic"
    	"crypto/rand"
    	"crypto/x509"
    	"crypto/x509/pkix"
    	"testing"
    	"time"
    )

    var now = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

    type issuer struct {
    	cert *x509.Certificate
    	key  *ecdsa.PrivateKey
    }

    func mk(t *testing.T, tmpl *x509.Certificate, parent *issuer) ([]byte, *issuer) {
    	t.Helper()
    	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    	p, signer := tmpl, key
    	if parent != nil {
    		p, signer = parent.cert, parent.key
    	}
    	der, err := x509.CreateCertificate(rand.Reader, tmpl, p, &key.PublicKey, signer)
    	if err != nil {
    		t.Fatal(err)
    	}
    	c, _ := x509.ParseCertificate(der)
    	return der, &issuer{c, key}
    }

    func caTmpl(name string) *x509.Certificate {
    	return &x509.Certificate{Subject: pkix.Name{CommonName: name}, NotBefore: now.Add(-time.Hour), NotAfter: now.AddDate(5, 0, 0),
    		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
    }

    func leafTmpl(cn string, names ...string) *x509.Certificate {
    	return &x509.Certificate{Subject: pkix.Name{CommonName: cn}, DNSNames: names, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(90 * 24 * time.Hour),
    		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
    }

    func TestChains(t *testing.T) {
    	_, root := mk(t, caTmpl("Keybox Root"), nil)
    	interDER, inter := mk(t, caTmpl("Keybox Issuing CA"), root)
    	leafDER, _ := mk(t, leafTmpl("sync.keybox.test", "sync.keybox.test"), inter)
    	sanOnlyDER, _ := mk(t, leafTmpl("server-17", "sync.keybox.test"), inter)
    	cnOnlyDER, _ := mk(t, leafTmpl("sync.keybox.test"), inter)
    	selfDER, _ := mk(t, leafTmpl("sync.keybox.test", "sync.keybox.test"), nil)
    	_, otherRoot := mk(t, caTmpl("Mallory Root"), nil)
    	otherDER, _ := mk(t, leafTmpl("sync.keybox.test", "sync.keybox.test"), otherRoot)
    	clientTmpl := leafTmpl("sync.keybox.test", "sync.keybox.test")
    	clientTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
    	clientDER, _ := mk(t, clientTmpl, inter)

    	roots := x509.NewCertPool()
    	roots.AddCert(root.cert)

    	good := map[string][][]byte{
    		"leaf + intermediate":         {leafDER, interDER},
    		"name only in DNSNames (SAN)": {sanOnlyDER, interDER},
    	}
    	for name, chain := range good {
    		if leaf, err := verifyServerChain(chain, roots, "sync.keybox.test", now); err != nil || leaf == nil {
    			t.Errorf("%s: verifyServerChain = %v, %v; want the leaf and nil", name, leaf, err)
    		}
    	}

    	bad := map[string]struct {
    		chain [][]byte
    		host  string
    		at    time.Time
    	}{
    		"self-signed impostor":              {[][]byte{selfDER}, "sync.keybox.test", now},
    		"signed by an untrusted root":       {[][]byte{otherDER}, "sync.keybox.test", now},
    		"missing intermediate":              {[][]byte{leafDER}, "sync.keybox.test", now},
    		"wrong host":                        {[][]byte{leafDER, interDER}, "api.keybox.test", now},
    		"name only in Common Name (no SAN)": {[][]byte{cnOnlyDER, interDER}, "sync.keybox.test", now},
    		"expired":                           {[][]byte{leafDER, interDER}, "sync.keybox.test", now.Add(100 * 24 * time.Hour)},
    		"client-auth-only certificate":      {[][]byte{clientDER, interDER}, "sync.keybox.test", now},
    		"garbage":                           {[][]byte{[]byte("not a certificate")}, "sync.keybox.test", now},
    		"empty":                             {nil, "sync.keybox.test", now},
    	}
    	for name, tc := range bad {
    		if _, err := verifyServerChain(tc.chain, roots, tc.host, tc.at); err == nil {
    			t.Errorf("%s: verifyServerChain succeeded, want an error", name)
    		}
    	}
    }
---

`crypto/tls` verifies certificates for you, and you should let it. But sometimes you
must verify a chain yourself: a certificate uploaded through an API, a signed document,
or custom rules in `tls.Config.VerifyConnection`. That code is famously easy to get
wrong.

## What `Verify` checks

`(*x509.Certificate).Verify(opts)` searches for a **chain** from the leaf, through any
intermediates, to one of the roots, and checks at every step:

- each signature against its issuer's public key,
- each certificate's validity period against `CurrentTime` (zero means now),
- that each issuer is a CA (`IsCA`, key usage for signing) within its path length and
  any name constraints,
- that the leaf matches `DNSName` (or an IP), using **only** the Subject Alternative
  Names,
- that the chain allows the requested `KeyUsages` (the default is `ServerAuth`; use
  `ExtKeyUsageAny` only if you mean it).

It returns every valid chain it found, or an error explaining why none was. Some errors
you'll see: `x509: certificate signed by unknown authority`,
`x509: certificate is valid for sync.keybox.test, not api.keybox.test`,
`x509: certificate has expired or is not yet valid`.

## Roots vs intermediates

`VerifyOptions` has two pools, and mixing them up is a vulnerability:

- **`Roots`**: what you *trust*. Only put CA certificates here that you obtained
  out of band (shipped with your app, or `x509.SystemCertPool()`).
- **`Intermediates`**: *untrusted* helpers, usually whatever the peer sent after its
  leaf. They're only used to build a path to a root.

Adding the peer's certificates to `Roots` "to make the error go away" makes Mallory's
self-signed certificate trusted, because she sent it.

## Classic custom-verification bugs

- **Checking names without checking signatures**, like the starter code below. Anyone
  can create a certificate claiming any name.
- **`InsecureSkipVerify: true` plus a callback that checks too little.** Setting
  `InsecureSkipVerify` turns off *all* of Go's checks; your `VerifyPeerCertificate`
  callback must then redo all of them. Prefer `VerifyConnection`, which runs *after*
  normal verification, for extra checks such as pinning a key.
- **Using `CommonName` for hostnames.** Go ignores it for verification. Put names in
  `DNSNames`.
- **Forgetting the time.** Verifying signed documents "as of now" is right for TLS;
  for archived documents you may need the signing time, which needs a trusted timestamp.
- **Ignoring key usage.** A certificate issued for client authentication shouldn't be
  accepted as a server.

## Revocation

Certificates can be revoked before they expire, through CRLs or OCSP. In practice
revocation checking is unreliable and many clients skip it; Go's `Verify` doesn't do
it. Short certificate lifetimes are the modern, dependable answer.

## Your task

Keybox's certificate uploader has a `verifyServerChain` that only compares the Common
Name. Replace its check with real chain verification:

1. Parse `rawCerts[0]` as the leaf (return errors, including for an empty slice).
2. Parse the rest into an **intermediates** pool (return parse errors).
3. Call `leaf.Verify` with `Roots: roots`, your intermediates, `DNSName: host`,
   `CurrentTime: now` and `KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}`.
4. Return the leaf if that succeeds, or the error.

The tests build a real root, intermediate and several leaves, including a self-signed
impostor and a certificate from a different root.
