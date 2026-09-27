---
title: Assemble Keybox
quiz:
  - question: Bob's `ImportBundle` verifies Alice's signature, then opens and stores shares one at a time, returning on the first failure. What can go wrong?
    options:
      - text: Nothing, since the signature was valid
      - text: A bundle with a bad share halfway through leaves Bob's vault partly updated; open everything first, then store
        correct: true
      - text: HPKE can't open more than one share per key
      - text: The signature becomes invalid after the first share
    explanation: |
      Even a validly signed bundle can contain a share Bob can't open (a bug, or a
      share sealed to someone else). Validating everything before changing any state
      keeps imports all-or-nothing.
exercise:
  starter: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/ed25519"
    	"crypto/hpke"
    	"crypto/pbkdf2"
    	"crypto/rand"
    	"crypto/sha256"
    	"encoding/binary"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"time"
    )

    var (
    	ErrNotFound       = errors.New("keybox: no such secret")
    	ErrDecrypt        = errors.New("keybox: decryption failed")
    	ErrBadSignature   = errors.New("keybox: bad bundle signature")
    	ErrWrongRecipient = errors.New("keybox: bundle is for someone else")
    	ErrExpired        = errors.New("keybox: bundle is too old or from the future")
    )

    const (
    	bundleContext = "keybox bundle v1\n"
    	shareInfo     = "keybox share v2"
    	maxAge        = 7 * 24 * time.Hour
    	maxSkew       = 5 * time.Minute
    )

    // ======== keys and identities (chapters 7 and 8) ========

    var (
    	kem  = hpke.MLKEM768X25519()
    	kdf  = hpke.HKDFSHA256()
    	aead = hpke.AES256GCM()
    )

    // Identity is a user's private keys. It never leaves their devices.
    type Identity struct {
    	Name       string
    	Encryption hpke.PrivateKey
    	Signing    ed25519.PrivateKey
    }

    // DirectoryEntry is what the Keybox directory publishes about a user. Clients
    // check the SigningKey fingerprint out of band.
    type DirectoryEntry struct {
    	Name          string
    	EncryptionKey []byte
    	SigningKey    ed25519.PublicKey
    }

    func NewIdentity(name string) (*Identity, error) {
    	enc, err := kem.GenerateKey()
    	if err != nil {
    		return nil, err
    	}
    	_, sign, err := ed25519.GenerateKey(nil)
    	if err != nil {
    		return nil, err
    	}
    	return &Identity{Name: name, Encryption: enc, Signing: sign}, nil
    }

    func (id *Identity) Entry() DirectoryEntry {
    	return DirectoryEntry{
    		Name:          id.Name,
    		EncryptionKey: id.Encryption.PublicKey().Bytes(),
    		SigningKey:    id.Signing.Public().(ed25519.PublicKey),
    	}
    }

    // ======== the vault at rest (chapters 5 and 6) ========

    type Vault struct {
    	id      string
    	gcm     cipher.AEAD
    	entries map[string][]byte
    }

    // OpenVault derives the vault key from a password with PBKDF2. (A full
    // Keybox wraps a random data key under it, as in chapter 6.)
    func OpenVault(id, password string, salt []byte, iterations int) (*Vault, error) {
    	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, 32)
    	if err != nil {
    		return nil, err
    	}
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	gcm, err := cipher.NewGCMWithRandomNonce(block)
    	if err != nil {
    		return nil, err
    	}
    	return &Vault{id: id, gcm: gcm, entries: map[string][]byte{}}, nil
    }

    func entryAD(vaultID, name string) []byte {
    	var ad []byte
    	for _, f := range []string{"keybox v1 vault entry", vaultID, name} {
    		ad = binary.BigEndian.AppendUint64(ad, uint64(len(f)))
    		ad = append(ad, f...)
    	}
    	return ad
    }

    func (v *Vault) Put(name string, secret []byte) {
    	v.entries[name] = v.gcm.Seal(nil, nil, secret, entryAD(v.id, name))
    }

    func (v *Vault) Get(name string) ([]byte, error) {
    	sealed, ok := v.entries[name]
    	if !ok {
    		return nil, ErrNotFound
    	}
    	pt, err := v.gcm.Open(nil, nil, sealed, entryAD(v.id, name))
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    // ======== sharing (this lesson) ========

    // Share is one secret, sealed with HPKE to the recipient.
    type Share struct {
    	Name   string `json:"name"`
    	Sealed []byte `json:"sealed"`
    }

    type Bundle struct {
    	From    string  `json:"from"`
    	To      string  `json:"to"`
    	Created int64   `json:"created"`
    	Shares  []Share `json:"shares"`
    }

    type SignedBundle struct {
    	Payload   []byte `json:"payload"`
    	Signature []byte `json:"signature"`
    }

    func encodeBundle(b Bundle) ([]byte, error) { return json.Marshal(b) }

    func decodeBundle(payload []byte) (Bundle, error) {
    	var b Bundle
    	err := json.Unmarshal(payload, &b)
    	return b, err
    }

    // shareInfoFor binds an HPKE share to the secret's name.
    func shareInfoFor(name string) []byte {
    	return []byte(shareInfo + "\x00" + name)
    }

    // ExportBundle seals each named secret from vault to the recipient with
    // HPKE (info: shareInfoFor(name)), and returns the bundle signed by from.
    func ExportBundle(from *Identity, vault *Vault, names []string, to DirectoryEntry, now time.Time) (SignedBundle, error) {
    	// ?
    	return SignedBundle{}, errors.New("not implemented")
    }

    // ImportBundle verifies sb against the sender's directory entry, checks it's
    // addressed to me and fresh, opens every share, and only then stores them
    // all in vault. It returns the imported names in bundle order.
    func ImportBundle(me *Identity, vault *Vault, sb SignedBundle, from DirectoryEntry, now time.Time) ([]string, error) {
    	// ?
    	return nil, errors.New("not implemented")
    }

    func main() {
    	alice, _ := NewIdentity("alice")
    	bob, _ := NewIdentity("bob")
    	salt := make([]byte, 16)
    	rand.Read(salt)
    	aliceVault, _ := OpenVault("vault-alice", "alice's passphrase", salt, 1000)
    	bobVault, _ := OpenVault("vault-bob", "bob's passphrase", salt, 1000)

    	aliceVault.Put("db-password", []byte("s3cr3t-team-pw"))
    	aliceVault.Put("github-token", []byte("ghp_alice"))

    	now := time.Now()
    	sb, err := ExportBundle(alice, aliceVault, []string{"db-password"}, bob.Entry(), now)
    	fmt.Println("export:", len(sb.Payload), "byte payload,", err)

    	names, err := ImportBundle(bob, bobVault, sb, alice.Entry(), now.Add(time.Minute))
    	fmt.Println("import:", names, err)
    	got, err := bobVault.Get("db-password")
    	fmt.Printf("bob's vault: %q %v\n", got, err)
    }
  solution: |
    package main

    import (
    	"crypto/aes"
    	"crypto/cipher"
    	"crypto/ed25519"
    	"crypto/hpke"
    	"crypto/pbkdf2"
    	"crypto/rand"
    	"crypto/sha256"
    	"encoding/binary"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"time"
    )

    var (
    	ErrNotFound       = errors.New("keybox: no such secret")
    	ErrDecrypt        = errors.New("keybox: decryption failed")
    	ErrBadSignature   = errors.New("keybox: bad bundle signature")
    	ErrWrongRecipient = errors.New("keybox: bundle is for someone else")
    	ErrExpired        = errors.New("keybox: bundle is too old or from the future")
    )

    const (
    	bundleContext = "keybox bundle v1\n"
    	shareInfo     = "keybox share v2"
    	maxAge        = 7 * 24 * time.Hour
    	maxSkew       = 5 * time.Minute
    )

    // ======== keys and identities (chapters 7 and 8) ========

    var (
    	kem  = hpke.MLKEM768X25519()
    	kdf  = hpke.HKDFSHA256()
    	aead = hpke.AES256GCM()
    )

    // Identity is a user's private keys. It never leaves their devices.
    type Identity struct {
    	Name       string
    	Encryption hpke.PrivateKey
    	Signing    ed25519.PrivateKey
    }

    // DirectoryEntry is what the Keybox directory publishes about a user. Clients
    // check the SigningKey fingerprint out of band.
    type DirectoryEntry struct {
    	Name          string
    	EncryptionKey []byte
    	SigningKey    ed25519.PublicKey
    }

    func NewIdentity(name string) (*Identity, error) {
    	enc, err := kem.GenerateKey()
    	if err != nil {
    		return nil, err
    	}
    	_, sign, err := ed25519.GenerateKey(nil)
    	if err != nil {
    		return nil, err
    	}
    	return &Identity{Name: name, Encryption: enc, Signing: sign}, nil
    }

    func (id *Identity) Entry() DirectoryEntry {
    	return DirectoryEntry{
    		Name:          id.Name,
    		EncryptionKey: id.Encryption.PublicKey().Bytes(),
    		SigningKey:    id.Signing.Public().(ed25519.PublicKey),
    	}
    }

    // ======== the vault at rest (chapters 5 and 6) ========

    type Vault struct {
    	id      string
    	gcm     cipher.AEAD
    	entries map[string][]byte
    }

    // OpenVault derives the vault key from a password with PBKDF2. (A full
    // Keybox wraps a random data key under it, as in chapter 6.)
    func OpenVault(id, password string, salt []byte, iterations int) (*Vault, error) {
    	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, 32)
    	if err != nil {
    		return nil, err
    	}
    	block, err := aes.NewCipher(key)
    	if err != nil {
    		return nil, err
    	}
    	gcm, err := cipher.NewGCMWithRandomNonce(block)
    	if err != nil {
    		return nil, err
    	}
    	return &Vault{id: id, gcm: gcm, entries: map[string][]byte{}}, nil
    }

    func entryAD(vaultID, name string) []byte {
    	var ad []byte
    	for _, f := range []string{"keybox v1 vault entry", vaultID, name} {
    		ad = binary.BigEndian.AppendUint64(ad, uint64(len(f)))
    		ad = append(ad, f...)
    	}
    	return ad
    }

    func (v *Vault) Put(name string, secret []byte) {
    	v.entries[name] = v.gcm.Seal(nil, nil, secret, entryAD(v.id, name))
    }

    func (v *Vault) Get(name string) ([]byte, error) {
    	sealed, ok := v.entries[name]
    	if !ok {
    		return nil, ErrNotFound
    	}
    	pt, err := v.gcm.Open(nil, nil, sealed, entryAD(v.id, name))
    	if err != nil {
    		return nil, ErrDecrypt
    	}
    	return pt, nil
    }

    // ======== sharing (this lesson) ========

    // Share is one secret, sealed with HPKE to the recipient.
    type Share struct {
    	Name   string `json:"name"`
    	Sealed []byte `json:"sealed"`
    }

    type Bundle struct {
    	From    string  `json:"from"`
    	To      string  `json:"to"`
    	Created int64   `json:"created"`
    	Shares  []Share `json:"shares"`
    }

    type SignedBundle struct {
    	Payload   []byte `json:"payload"`
    	Signature []byte `json:"signature"`
    }

    func encodeBundle(b Bundle) ([]byte, error) { return json.Marshal(b) }

    func decodeBundle(payload []byte) (Bundle, error) {
    	var b Bundle
    	err := json.Unmarshal(payload, &b)
    	return b, err
    }

    // shareInfoFor binds an HPKE share to the secret's name.
    func shareInfoFor(name string) []byte {
    	return []byte(shareInfo + "\x00" + name)
    }

    func ExportBundle(from *Identity, vault *Vault, names []string, to DirectoryEntry, now time.Time) (SignedBundle, error) {
    	pub, err := kem.NewPublicKey(to.EncryptionKey)
    	if err != nil {
    		return SignedBundle{}, err
    	}
    	b := Bundle{From: from.Name, To: to.Name, Created: now.Unix()}
    	for _, name := range names {
    		secret, err := vault.Get(name)
    		if err != nil {
    			return SignedBundle{}, err
    		}
    		sealed, err := hpke.Seal(pub, kdf, aead, shareInfoFor(name), secret)
    		if err != nil {
    			return SignedBundle{}, err
    		}
    		b.Shares = append(b.Shares, Share{Name: name, Sealed: sealed})
    	}
    	payload, err := encodeBundle(b)
    	if err != nil {
    		return SignedBundle{}, err
    	}
    	sig := ed25519.Sign(from.Signing, append([]byte(bundleContext), payload...))
    	return SignedBundle{Payload: payload, Signature: sig}, nil
    }

    func ImportBundle(me *Identity, vault *Vault, sb SignedBundle, from DirectoryEntry, now time.Time) ([]string, error) {
    	if len(from.SigningKey) != ed25519.PublicKeySize ||
    		!ed25519.Verify(from.SigningKey, append([]byte(bundleContext), sb.Payload...), sb.Signature) {
    		return nil, ErrBadSignature
    	}
    	b, err := decodeBundle(sb.Payload)
    	if err != nil || b.From != from.Name {
    		return nil, ErrBadSignature
    	}
    	if b.To != me.Name {
    		return nil, ErrWrongRecipient
    	}
    	age := now.Sub(time.Unix(b.Created, 0))
    	if age > maxAge || age < -maxSkew {
    		return nil, ErrExpired
    	}
    	opened := make(map[string][]byte, len(b.Shares))
    	var names []string
    	for _, s := range b.Shares {
    		pt, err := hpke.Open(me.Encryption, kdf, aead, shareInfoFor(s.Name), s.Sealed)
    		if err != nil {
    			return nil, ErrDecrypt
    		}
    		opened[s.Name] = pt
    		names = append(names, s.Name)
    	}
    	for _, name := range names {
    		vault.Put(name, opened[name])
    	}
    	return names, nil
    }

    func main() {
    	alice, _ := NewIdentity("alice")
    	bob, _ := NewIdentity("bob")
    	salt := make([]byte, 16)
    	rand.Read(salt)
    	aliceVault, _ := OpenVault("vault-alice", "alice's passphrase", salt, 1000)
    	bobVault, _ := OpenVault("vault-bob", "bob's passphrase", salt, 1000)

    	aliceVault.Put("db-password", []byte("s3cr3t-team-pw"))
    	aliceVault.Put("github-token", []byte("ghp_alice"))

    	now := time.Now()
    	sb, err := ExportBundle(alice, aliceVault, []string{"db-password"}, bob.Entry(), now)
    	fmt.Println("export:", len(sb.Payload), "byte payload,", err)

    	names, err := ImportBundle(bob, bobVault, sb, alice.Entry(), now.Add(time.Minute))
    	fmt.Println("import:", names, err)
    	got, err := bobVault.Get("db-password")
    	fmt.Printf("bob's vault: %q %v\n", got, err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/ed25519"
    	"crypto/hpke"
    	"encoding/json/v2"
    	"errors"
    	"slices"
    	"testing"
    	"time"
    )

    var testNow = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

    type world struct {
    	alice, bob, carol *Identity
    	aliceVault        *Vault
    	bobVault          *Vault
    }

    func setup(t *testing.T) *world {
    	t.Helper()
    	w := &world{}
    	var err error
    	for _, p := range []struct {
    		dst  **Identity
    		name string
    	}{{&w.alice, "alice"}, {&w.bob, "bob"}, {&w.carol, "carol"}} {
    		if *p.dst, err = NewIdentity(p.name); err != nil {
    			t.Fatal(err)
    		}
    	}
    	salt := bytes.Repeat([]byte{9}, 16)
    	w.aliceVault, _ = OpenVault("vault-alice", "alice pw", salt, 1000)
    	w.bobVault, _ = OpenVault("vault-bob", "bob pw", salt, 1000)
    	w.aliceVault.Put("db-password", []byte("s3cr3t-team-pw"))
    	w.aliceVault.Put("github-token", []byte("ghp_alice"))
    	w.aliceVault.Put("wifi", []byte("hunter2"))
    	return w
    }

    func (w *world) export(t *testing.T, names ...string) SignedBundle {
    	t.Helper()
    	sb, err := ExportBundle(w.alice, w.aliceVault, names, w.bob.Entry(), testNow)
    	if err != nil {
    		t.Fatalf("ExportBundle(%v) error = %v", names, err)
    	}
    	return sb
    }

    // resign lets the tests build bundles that Alice's key really signed.
    func resign(t *testing.T, id *Identity, b Bundle) SignedBundle {
    	t.Helper()
    	payload, err := json.Marshal(b)
    	if err != nil {
    		t.Fatal(err)
    	}
    	return SignedBundle{Payload: payload, Signature: ed25519.Sign(id.Signing, append([]byte(bundleContext), payload...))}
    }

    func decode(t *testing.T, sb SignedBundle) Bundle {
    	t.Helper()
    	var b Bundle
    	if err := json.Unmarshal(sb.Payload, &b); err != nil {
    		t.Fatalf("payload isn't a JSON Bundle: %v", err)
    	}
    	return b
    }

    func TestEndToEnd(t *testing.T) {
    	w := setup(t)
    	sb := w.export(t, "db-password", "github-token")

    	b := decode(t, sb)
    	if b.From != "alice" || b.To != "bob" || b.Created != testNow.Unix() || len(b.Shares) != 2 {
    		t.Fatalf("exported bundle = from %q to %q created %d with %d shares; want alice, bob, %d, 2",
    			b.From, b.To, b.Created, len(b.Shares), testNow.Unix())
    	}
    	if bytes.Contains(sb.Payload, []byte("s3cr3t-team-pw")) || bytes.Contains(sb.Payload, []byte("ghp_alice")) {
    		t.Fatalf("the bundle payload contains a plaintext secret; seal each one with HPKE")
    	}

    	names, err := ImportBundle(w.bob, w.bobVault, sb, w.alice.Entry(), testNow.Add(time.Hour))
    	if err != nil {
    		t.Fatalf("ImportBundle error = %v", err)
    	}
    	if !slices.Equal(names, []string{"db-password", "github-token"}) {
    		t.Errorf("ImportBundle names = %v, want [db-password github-token]", names)
    	}
    	for name, want := range map[string]string{"db-password": "s3cr3t-team-pw", "github-token": "ghp_alice"} {
    		if got, err := w.bobVault.Get(name); err != nil || string(got) != want {
    			t.Errorf("bob's vault %q = %q, %v; want %q", name, got, err, want)
    		}
    	}
    	if _, err := w.bobVault.Get("wifi"); !errors.Is(err, ErrNotFound) {
    		t.Errorf("bob's vault has %q, which wasn't shared", "wifi")
    	}
    }

    func TestExportErrors(t *testing.T) {
    	w := setup(t)
    	if _, err := ExportBundle(w.alice, w.aliceVault, []string{"nope"}, w.bob.Entry(), testNow); !errors.Is(err, ErrNotFound) {
    		t.Errorf("exporting a missing secret: error = %v, want ErrNotFound", err)
    	}
    	bad := w.bob.Entry()
    	bad.EncryptionKey = bad.EncryptionKey[:32]
    	if _, err := ExportBundle(w.alice, w.aliceVault, []string{"wifi"}, bad, testNow); err == nil {
    		t.Errorf("exporting to a malformed encryption key: error = nil, want an error")
    	}
    }

    func TestSignatureChecks(t *testing.T) {
    	w := setup(t)
    	sb := w.export(t, "db-password")
    	mallory, _ := NewIdentity("alice") // claims to be alice, but her own keys

    	tampered := SignedBundle{Payload: bytes.Replace(sb.Payload, []byte(`"to":"bob"`), []byte(`"to":"bob" `), 1), Signature: sb.Signature}
    	noContext := SignedBundle{Payload: sb.Payload, Signature: ed25519.Sign(w.alice.Signing, sb.Payload)}
    	fromCarol := decode(t, sb)
    	fromCarol.From = "carol"

    	for name, bad := range map[string]SignedBundle{
    		"tampered payload":                  tampered,
    		"signed by an impostor":             resign(t, mallory, decode(t, sb)),
    		"signed without the context":        noContext,
    		"alice signing a bundle from carol": resign(t, w.alice, fromCarol),
    		"no signature":                      {Payload: sb.Payload},
    	} {
    		if _, err := ImportBundle(w.bob, w.bobVault, bad, w.alice.Entry(), testNow); !errors.Is(err, ErrBadSignature) {
    			t.Errorf("%s: ImportBundle error = %v, want ErrBadSignature", name, err)
    		}
    	}
    	broken := w.alice.Entry()
    	broken.SigningKey = broken.SigningKey[:5]
    	if _, err := ImportBundle(w.bob, w.bobVault, sb, broken, testNow); !errors.Is(err, ErrBadSignature) {
    		t.Errorf("5-byte signing key: ImportBundle error = %v, want ErrBadSignature (and no panic)", err)
    	}
    }

    func TestClaims(t *testing.T) {
    	w := setup(t)
    	sb := w.export(t, "db-password")
    	if _, err := ImportBundle(w.carol, w.bobVault, sb, w.alice.Entry(), testNow); !errors.Is(err, ErrWrongRecipient) {
    		t.Errorf("carol importing bob's bundle: error = %v, want ErrWrongRecipient", err)
    	}
    	for _, d := range []time.Duration{maxAge + time.Second, -maxSkew - time.Second} {
    		if _, err := ImportBundle(w.bob, w.bobVault, sb, w.alice.Entry(), testNow.Add(d)); !errors.Is(err, ErrExpired) {
    			t.Errorf("importing %v after creation: error = %v, want ErrExpired", d, err)
    		}
    	}
    }

    func TestSharesBoundAndAtomic(t *testing.T) {
    	w := setup(t)
    	b := decode(t, w.export(t, "db-password", "github-token"))

    	// Alice's signature is valid, but the share names are swapped.
    	swapped := b
    	swapped.Shares = []Share{{Name: "db-password", Sealed: b.Shares[1].Sealed}, {Name: "github-token", Sealed: b.Shares[0].Sealed}}
    	if _, err := ImportBundle(w.bob, w.bobVault, resign(t, w.alice, swapped), w.alice.Entry(), testNow); !errors.Is(err, ErrDecrypt) {
    		t.Errorf("shares with swapped names: error = %v, want ErrDecrypt (bind each share to its name with the HPKE info)", err)
    	}

    	// The first share is fine, the second was sealed to carol.
    	carolPub, _ := hpke.MLKEM768X25519().NewPublicKey(w.carol.Entry().EncryptionKey)
    	forCarol, _ := hpke.Seal(carolPub, kdf, aead, shareInfoFor("github-token"), []byte("ghp_alice"))
    	mixed := b
    	mixed.Shares = []Share{b.Shares[0], {Name: "github-token", Sealed: forCarol}}
    	if _, err := ImportBundle(w.bob, w.bobVault, resign(t, w.alice, mixed), w.alice.Entry(), testNow); !errors.Is(err, ErrDecrypt) {
    		t.Errorf("a share sealed to someone else: error = %v, want ErrDecrypt", err)
    	}
    	if _, err := w.bobVault.Get("db-password"); !errors.Is(err, ErrNotFound) {
    		t.Errorf("a failed import still stored db-password in bob's vault; open every share before storing any")
    	}
    }
---

Everything comes together. Keybox's full flow has three parts, and you've built each
piece already:

1. **At rest** (chapters 5 and 6): a vault whose entries are sealed with AES-256-GCM
   under a key derived from a password, each bound to its vault and name with
   associated data.
2. **Sharing** (chapter 7): each secret sealed to the recipient's public key with HPKE,
   using the hybrid post-quantum X-Wing KEM, so the server relays only ciphertext.
3. **Authenticity** (chapter 8): the bundle of shares signed with the sender's Ed25519
   key, with a context prefix, a recipient and a timestamp inside the signed bytes.

## The flow

Alice shares `db-password` with Bob:

```
Alice's vault --Get--> plaintext --HPKE Seal to Bob (info binds the name)--> Share
Shares + from/to/created --JSON--> payload --Ed25519 sign (context prefix)--> SignedBundle
                                      ... Keybox server relays it ...
Bob: verify signature with Alice's directory key --> parse --> check to, created
     --> HPKE Open every share --> Put all into Bob's vault
```

Each user has an `Identity` (their private keys, which never leave their devices) and a
public `DirectoryEntry` (name, HPKE public key, Ed25519 public key). The starter code
provides identities, the vault, bundle encoding, and `shareInfoFor(name)`, which binds
each HPKE share to the secret's name the same way associated data bound vault entries.

## Design decisions worth noticing

- **The server learns nothing useful.** It sees the names of shared secrets and who
  shares with whom (metadata), but never a secret or a key. A real Keybox would also
  encrypt the names.
- **Authenticity doesn't depend on the server.** Bob checks the signature with Alice's
  *signing* key, whose fingerprint he verified out of band. A malicious directory
  handing out a fake key is the remaining risk, which key transparency addresses.
- **The `From` field must match the key you verified with.** Otherwise Alice's valid
  signature could vouch for a bundle claiming to come from Carol.
- **All or nothing.** Bob opens *every* share before storing *any*. A bundle with one bad
  share is rejected whole, so a half-imported bundle can't leave his vault in a state
  nobody intended.
- **One error per failure class**, and no plaintext ever in a log or error message.

## Your task

Implement the two halves of sharing.

**`ExportBundle(from, vault, names, to, now)`**

1. Parse the recipient's key: `kem.NewPublicKey(to.EncryptionKey)` (return any error).
2. Start a `Bundle{From: from.Name, To: to.Name, Created: now.Unix()}`.
3. For each name, in order: `vault.Get(name)` (return its error), then
   `hpke.Seal(pub, kdf, aead, shareInfoFor(name), secret)`, and append
   `Share{Name: name, Sealed: sealed}`.
4. Encode with `encodeBundle`, sign `bundleContext + payload` with `from.Signing`, and
   return the `SignedBundle`.

**`ImportBundle(me, vault, sb, from, now)`**

1. If `from.SigningKey` isn't `ed25519.PublicKeySize` bytes or the signature over
   `bundleContext + sb.Payload` doesn't verify, return `ErrBadSignature`.
2. `decodeBundle(sb.Payload)`. A decode error, or `b.From != from.Name`, is also
   `ErrBadSignature`.
3. `b.To != me.Name` is `ErrWrongRecipient`. `Created` more than `maxAge` in the past or
   more than `maxSkew` in the future is `ErrExpired`.
4. Open every share with `hpke.Open(me.Encryption, kdf, aead, shareInfoFor(s.Name),
   s.Sealed)`. Any failure returns `ErrDecrypt` and stores nothing.
5. Only then `vault.Put` each one, and return the names in bundle order.

The tests play the whole story: Alice exports, Bob imports, and then Mallory tries an
impostor key, a tampered payload, a missing context, a bundle for someone else, an
expired bundle, swapped share names and a share sealed to Carol.
