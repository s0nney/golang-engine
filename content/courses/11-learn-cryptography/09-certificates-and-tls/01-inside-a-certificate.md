---
title: Inside a Certificate
quiz:
  - question: What does a CA's signature on a certificate actually vouch for?
    options:
      - text: That the server is run by honest people
      - text: That the public key in the certificate belongs to the names listed in it, for the stated validity period and uses
        correct: true
      - text: That the connection is encrypted
      - text: That the server's software has no vulnerabilities
    explanation: |
      A certificate is a signed statement binding a public key to identities (DNS
      names, IPs, emails) with constraints. Nothing more. It says nothing about the
      trustworthiness of whoever holds the key.
  - question: Which field decides which hostnames a TLS server certificate is valid for, in Go and modern browsers?
    options:
      - text: '`Subject.CommonName`'
      - text: '`Issuer`'
      - text: The Subject Alternative Name extension (`DNSNames`, `IPAddresses`)
        correct: true
      - text: '`SerialNumber`'
    explanation: |
      Go has ignored the Common Name for hostname checks since Go 1.15. Put every name
      the server answers to in `DNSNames` (or `IPAddresses`).
---

You met certificates from the client's side in
[Certificates](/courses/learn-http-clients/https-and-security/certificates): a server
presents one, Go checks it chains to a trusted CA and matches the hostname, and you
[never set `InsecureSkipVerify`](/courses/learn-http-clients/https-and-security/never-insecure-skip-verify).
Now that you know signatures, let's open one up and then build our own.

## A signed statement

An X.509 certificate is a **public key plus claims about it, signed by an issuer**.
The claims Go exposes on `*x509.Certificate`:

| Field | Meaning |
| --- | --- |
| `Subject` | Who the certificate is about (a distinguished name like `CN=keybox.test`) |
| `DNSNames`, `IPAddresses`, `EmailAddresses`, `URIs` | The Subject Alternative Names: identities the key is valid for |
| `PublicKey`, `PublicKeyAlgorithm` | The subject's public key |
| `Issuer` | Who signed it |
| `NotBefore`, `NotAfter` | Validity period |
| `SerialNumber` | Unique per issuer, used for revocation |
| `IsCA`, `BasicConstraintsValid`, `MaxPathLen` | Whether it may sign other certificates |
| `KeyUsage`, `ExtKeyUsage` | What the key may be used for (signing, server auth, client auth...) |
| `SignatureAlgorithm`, `Signature` | The issuer's signature over all of the above |

The encoding is **ASN.1 DER**, a strict binary format: exactly the kind of unambiguous
encoding chapter 3 recommended for signed data. PEM files (`-----BEGIN CERTIFICATE-----`)
are just base64 of the DER bytes with a header.

## Making one in Go

`x509.CreateCertificate(rand, template, parent, pub, priv)` builds and signs a
certificate. If `parent` is the template itself, it's **self-signed**:

```go
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"time"
)

func main() {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	tmpl := &x509.Certificate{
		Subject:     pkix.Name{CommonName: "keybox.test"},
		DNSNames:    []string{"keybox.test", "sync.keybox.test"},
		NotBefore:   start,
		NotAfter:    start.Add(90 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	cert, _ := x509.ParseCertificate(der)

	fmt.Println("subject:  ", cert.Subject)
	fmt.Println("issuer:   ", cert.Issuer)
	fmt.Println("names:    ", cert.DNSNames)
	fmt.Println("valid:    ", cert.NotBefore.Format(time.DateOnly), "to", cert.NotAfter.Format(time.DateOnly))
	fmt.Println("algorithm:", cert.PublicKeyAlgorithm, "signed with", cert.SignatureAlgorithm)
	fmt.Println("usage:    ", cert.KeyUsage, cert.ExtKeyUsage[0])
	fmt.Println("is CA:    ", cert.IsCA)
	fmt.Println("signature ok:", cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature) == nil)
	fmt.Println("may sign certs:", cert.CheckSignatureFrom(cert))
}
```

```
subject:   CN=keybox.test
issuer:    CN=keybox.test
names:     [keybox.test sync.keybox.test]
valid:     2026-09-01 to 2026-11-30
algorithm: ECDSA signed with ECDSA-SHA256
usage:     digitalSignature serverAuth
is CA:     false
signature ok: true
may sign certs: x509: invalid signature: parent certificate cannot sign this kind of certificate
```

Notes:

- `CreateCertificate` returns **DER bytes**; `x509.ParseCertificate` turns them into a
  struct. Only the fields in the template's documented list are used.
- **Pass `rand.Reader`**, not `nil`. With no `SerialNumber` in the template Go generates
  a random one, and it reads from the reader you pass; `nil` crashes.
- `KeyUsage` and `ExtKeyUsage` got `String` methods in Go 1.27, which is why they print
  readably.
- The signature over the certificate's body (`RawTBSCertificate`, "to be signed")
  checks out, but `CheckSignatureFrom(cert)` still fails: that asks whether `cert` may
  act as an *issuer*, and without `IsCA` it may not. Constraints are part of what gets
  verified, not decoration.
- Supported key types: ECDSA, Ed25519, RSA and, since Go 1.27, ML-DSA.

## Self-signed isn't "insecure", it's "unvouched"

Nothing is wrong with the cryptography of a self-signed certificate. The problem is
that nobody else vouches for the binding: a client can only trust it by already having
that exact certificate, which is called **pinning**. That's fine inside a system you
control, like Keybox's own sync protocol, and it's how private CAs work: you create one
self-signed **root**, distribute it to your clients, and sign everything else with it.
Next lesson, you build that CA.

## The public Web PKI

For public websites, browsers and operating systems ship a list of about 150 trusted
root CAs. Those CAs must follow strict rules, verify domain control before issuing
(Let's Encrypt automates this with the ACME protocol), and log every certificate they
issue to public **Certificate Transparency** logs, the same idea as the transparency
logs of the last chapter. Public certificate lifetimes are shrinking, too: the industry
has agreed to cut the maximum from 398 days down to 47 days by 2029, pushing everyone
toward automated renewal.
