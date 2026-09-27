---
title: Signing Releases
quiz:
  - question: An attacker compromises the Keybox download server and replaces both `keybox-linux-amd64` and `SHA256SUMS`. Which check still catches it?
    options:
      - text: Comparing the binary's SHA-256 with the downloaded `SHA256SUMS`
      - text: Verifying `SHA256SUMS.sig` with the release public key that shipped inside the *previous* Keybox version
        correct: true
      - text: Checking the file size
      - text: Downloading over HTTPS
    explanation: |
      The attacker can recompute any checksum and serve anything over HTTPS from the
      server they own. They can't produce a valid signature without the release private
      key, which never touches the download server. The public key must come from
      somewhere the attacker doesn't control, such as the already-installed version.
exercise:
  starter: |
    package main

    import (
    	"crypto/ed25519"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    )

    var (
    	ErrBadSignature = errors.New("release: SHA256SUMS signature is invalid")
    	ErrUnknownFile  = errors.New("release: file not listed in SHA256SUMS")
    	ErrChecksum     = errors.New("release: checksum mismatch")
    )

    const releaseContext = "keybox release v1\n"

    // verifyDownload checks a downloaded file against a signed SHA256SUMS:
    //  1. sig must be a valid Ed25519 signature by pub over releaseContext + sums
    //     (a public key of the wrong length is ErrBadSignature, not a panic),
    //  2. sums has one "<64 hex chars>  <filename>" entry per line; find the
    //     line whose filename is exactly name (else ErrUnknownFile),
    //  3. the SHA-256 of data, as lowercase hex, must equal that line's hash
    //     (else ErrChecksum).
    func verifyDownload(pub ed25519.PublicKey, sums, sig []byte, name string, data []byte) error {
    	// ?
    	return nil
    }

    func main() {
    	pub, priv, _ := ed25519.GenerateKey(nil)
    	binary := []byte("pretend this is the keybox binary")
    	sums := []byte(fmt.Sprintf("%x  keybox-linux-amd64\n%x  keybox-darwin-arm64\n",
    		sha256.Sum256(binary), sha256.Sum256([]byte("mac build"))))
    	sig := ed25519.Sign(priv, append([]byte(releaseContext), sums...))

    	fmt.Println("genuine:", verifyDownload(pub, sums, sig, "keybox-linux-amd64", binary))
    	fmt.Println("trojaned:", verifyDownload(pub, sums, sig, "keybox-linux-amd64", []byte("evil")))
    	fmt.Println("unknown:", verifyDownload(pub, sums, sig, "keybox-windows.exe", binary))
    }
  solution: |
    package main

    import (
    	"crypto/ed25519"
    	"crypto/sha256"
    	"encoding/hex"
    	"errors"
    	"fmt"
    	"strings"
    )

    var (
    	ErrBadSignature = errors.New("release: SHA256SUMS signature is invalid")
    	ErrUnknownFile  = errors.New("release: file not listed in SHA256SUMS")
    	ErrChecksum     = errors.New("release: checksum mismatch")
    )

    const releaseContext = "keybox release v1\n"

    func verifyDownload(pub ed25519.PublicKey, sums, sig []byte, name string, data []byte) error {
    	if len(pub) != ed25519.PublicKeySize || !ed25519.Verify(pub, append([]byte(releaseContext), sums...), sig) {
    		return ErrBadSignature
    	}
    	for line := range strings.Lines(string(sums)) {
    		hash, file, ok := strings.Cut(strings.TrimRight(line, "\n"), "  ")
    		if !ok || file != name {
    			continue
    		}
    		sum := sha256.Sum256(data)
    		if hex.EncodeToString(sum[:]) != hash {
    			return ErrChecksum
    		}
    		return nil
    	}
    	return ErrUnknownFile
    }

    func main() {
    	pub, priv, _ := ed25519.GenerateKey(nil)
    	binary := []byte("pretend this is the keybox binary")
    	sums := []byte(fmt.Sprintf("%x  keybox-linux-amd64\n%x  keybox-darwin-arm64\n",
    		sha256.Sum256(binary), sha256.Sum256([]byte("mac build"))))
    	sig := ed25519.Sign(priv, append([]byte(releaseContext), sums...))

    	fmt.Println("genuine:", verifyDownload(pub, sums, sig, "keybox-linux-amd64", binary))
    	fmt.Println("trojaned:", verifyDownload(pub, sums, sig, "keybox-linux-amd64", []byte("evil")))
    	fmt.Println("unknown:", verifyDownload(pub, sums, sig, "keybox-windows.exe", binary))
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/ed25519"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    	"testing"
    )

    var (
    	relPriv = ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x0f}, 32))
    	relPub  = relPriv.Public().(ed25519.PublicKey)
    	linux   = []byte("keybox for linux")
    	mac     = []byte("keybox for mac")
    	sums    = []byte(fmt.Sprintf("%x  keybox-linux-amd64\n%x  keybox-darwin-arm64\n", sha256.Sum256(linux), sha256.Sum256(mac)))
    	sumsSig = ed25519.Sign(relPriv, append([]byte(releaseContext), sums...))
    )

    func TestGenuineDownloads(t *testing.T) {
    	if err := verifyDownload(relPub, sums, sumsSig, "keybox-linux-amd64", linux); err != nil {
    		t.Errorf("genuine linux build: error = %v, want nil", err)
    	}
    	if err := verifyDownload(relPub, sums, sumsSig, "keybox-darwin-arm64", mac); err != nil {
    		t.Errorf("genuine mac build (second line): error = %v, want nil", err)
    	}
    }

    func TestTamperedDownloads(t *testing.T) {
    	if err := verifyDownload(relPub, sums, sumsSig, "keybox-linux-amd64", []byte("trojan")); !errors.Is(err, ErrChecksum) {
    		t.Errorf("trojaned binary: error = %v, want ErrChecksum", err)
    	}
    	if err := verifyDownload(relPub, sums, sumsSig, "keybox-linux-amd64", mac); !errors.Is(err, ErrChecksum) {
    		t.Errorf("mac binary under the linux name: error = %v, want ErrChecksum", err)
    	}
    	for _, name := range []string{"keybox-windows.exe", "keybox-linux", "keybox-linux-amd64.old", ""} {
    		if err := verifyDownload(relPub, sums, sumsSig, name, linux); !errors.Is(err, ErrUnknownFile) {
    			t.Errorf("verifyDownload(name %q) error = %v, want ErrUnknownFile", name, err)
    		}
    	}
    }

    func TestReplacedSums(t *testing.T) {
    	evil := []byte("evil")
    	evilSums := []byte(fmt.Sprintf("%x  keybox-linux-amd64\n", sha256.Sum256(evil)))
    	_, attackerPriv, _ := ed25519.GenerateKey(nil)
    	for name, tc := range map[string]struct{ sums, sig []byte }{
    		"replaced sums, old signature":      {evilSums, sumsSig},
    		"replaced sums, attacker signature": {evilSums, ed25519.Sign(attackerPriv, append([]byte(releaseContext), evilSums...))},
    		"replaced sums, no context":         {evilSums, ed25519.Sign(relPriv, evilSums)},
    		"genuine sums, missing signature":   {sums, nil},
    	} {
    		if err := verifyDownload(relPub, tc.sums, tc.sig, "keybox-linux-amd64", evil); !errors.Is(err, ErrBadSignature) {
    			t.Errorf("%s: error = %v, want ErrBadSignature", name, err)
    		}
    	}
    	if err := verifyDownload(relPub[:10], sums, sumsSig, "keybox-linux-amd64", linux); !errors.Is(err, ErrBadSignature) {
    		t.Errorf("10-byte public key: error = %v, want ErrBadSignature", err)
    	}
    }
---

Everything in Keybox depends on users running the real Keybox binary. A trojaned build
could simply upload every secret after decryption. **Signed releases** make sure the
code users install is the code the Keybox team built.

## Checksums aren't enough

Release pages often publish a `SHA256SUMS` file:

```
3f1c...a9e0  keybox-linux-amd64
7b22...04d1  keybox-darwin-arm64
```

That catches corrupted downloads, but not attacks: whoever can replace the binary on the
download server can usually replace `SHA256SUMS` too. You need something the attacker
can't recompute.

## Sign the sums file

The release process signs `SHA256SUMS` with a **release signing key** and publishes the
signature next to it. Signing the sums file instead of each binary keeps one signature
per release, and the check is two steps:

1. verify the signature on `SHA256SUMS` with the **release public key**,
2. hash the downloaded file and compare with its line in the (now trusted) sums file.

The release public key must come from somewhere the attacker can't touch: embedded in
the previous version of Keybox (for auto-updates), published in the source repository,
pinned in a package manager. If it comes from the same server as the download, it
proves nothing.

## Guarding the release key

- Keep the private key **offline** or in a hardware security module (HSM) or cloud KMS,
  and sign in a dedicated release job, never on a developer laptop.
- Use a **context prefix** (`"keybox release v1\n"`), so a signature for some other
  purpose can't be passed off as a release.
- Plan for **rotation**: ship the next public key inside the current release, so users
  can move to it if the old key must be retired.

## Transparency: catching misuse of the key

A signature proves the key holder signed; it doesn't prove they signed *only* what
they published. If the release key is stolen, or coerced, an attacker can sign a
malicious build and serve it to one targeted user. **Transparency logs** make that
visible: every signed artifact is appended to a public, append-only log (built on
Merkle trees of hashes), and clients refuse artifacts that aren't in it. Anyone can
monitor the log for releases they didn't expect.

You've used one already: Go modules. When `go` downloads a module version that isn't
in your `go.sum` yet, it checks the hash against the **Go checksum database**
(`sum.golang.org`), a transparency log, so everyone gets the same code for the same
version.
Sigstore does the same for container images and other artifacts, and Certificate
Transparency does it for TLS certificates (next chapter).

Finally, **reproducible builds** (which Go supports well) let independent parties rebuild
a release from source and confirm the binary matches, so a single compromised build
machine can't go unnoticed.

## Your task

Complete `verifyDownload(pub, sums, sig, name, data)`:

1. If `pub` isn't 32 bytes or the signature over `releaseContext + sums` doesn't verify,
   return `ErrBadSignature`. Do this before reading `sums` at all.
2. Go through `sums` line by line (`strings.Lines` is handy). Each line is
   `<hash>  <filename>` with **two** spaces; `strings.Cut(line, "  ")` splits it.
   Trim the trailing newline first. Find the line whose filename equals `name` exactly;
   if there's none, return `ErrUnknownFile`.
3. Compare the lowercase hex SHA-256 of `data` with that line's hash, and return
   `ErrChecksum` if they differ.
