---
title: Keybox Key Hierarchy
difficulty: hard
after: passwords-and-key-derivation
hints:
  - 'Write one private helper that turns a list of labels into the `info` string (`infoPrefix`, one count byte, then a length byte and the bytes of each label) and one that validates the labels and calls `hkdf.Key(sha256.New, root, []byte(accountID), info, 32)`. `Derive` and `Child` are then a few lines each.'
  - 'Why length prefixes? With a separator such as `/`, the paths `["a/b", "c"]` and `["a", "b/c"]` produce the same info and so the same key. Prefixing each label with its length (and the list with its count) makes every distinct list encode differently.'
  - '`Child` uses the purpose `"subtree"` internally, so `Derive` must refuse it: otherwise `Derive("subtree", "vault-7")` would hand out the *root key* of that child tree as if it were an ordinary key. Also clone the root in `NewKeyTree`.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var (
    	ErrBadRoot        = errors.New("keybox: root key must be 32 bytes")
    	ErrBadLabel       = errors.New("keybox: bad label")
    	ErrUnknownPurpose = errors.New("keybox: unknown key purpose")
    )

    const (
    	rootSize   = 32
    	keySize    = 32
    	maxPath    = 4
    	infoPrefix = "keybox/kdf/v1"
    	subtree    = "subtree" // reserved purpose, used only by Child
    )

    // purposes are the key purposes Derive accepts.
    var purposes = map[string]bool{"vault-enc": true, "vault-mac": true, "search-index": true, "export": true}

    // KeyTree derives every key an account needs from one root key.
    type KeyTree struct {
    	// your fields here
    }

    func NewKeyTree(root []byte, accountID string) (*KeyTree, error) {
    	return &KeyTree{}, nil
    }

    func (t *KeyTree) Derive(purpose string, path ...string) ([]byte, error) {
    	return nil, nil
    }

    func (t *KeyTree) Child(label string) (*KeyTree, error) {
    	return &KeyTree{}, nil
    }

    func main() {
    	root := make([]byte, rootSize) // all zeros: for demonstration only!
    	tree, err := NewKeyTree(root, "acct-42")
    	fmt.Println(err)
    	enc, _ := tree.Derive("vault-enc", "vault-7")
    	mac, _ := tree.Derive("vault-mac", "vault-7")
    	fmt.Printf("enc %x\nmac %x\n", enc, mac) // two different 32-byte keys
    	_, err = tree.Derive(subtree, "vault-7")
    	fmt.Println(err) // want: keybox: unknown key purpose
    }
  solution: |
    package main

    import (
    	"bytes"
    	"crypto/hkdf"
    	"crypto/sha256"
    	"errors"
    	"fmt"
    )

    var (
    	ErrBadRoot        = errors.New("keybox: root key must be 32 bytes")
    	ErrBadLabel       = errors.New("keybox: bad label")
    	ErrUnknownPurpose = errors.New("keybox: unknown key purpose")
    )

    const (
    	rootSize   = 32
    	keySize    = 32
    	maxPath    = 4
    	infoPrefix = "keybox/kdf/v1"
    	subtree    = "subtree" // reserved purpose, used only by Child
    )

    // purposes are the key purposes Derive accepts.
    var purposes = map[string]bool{"vault-enc": true, "vault-mac": true, "search-index": true, "export": true}

    // KeyTree derives every key an account needs from one root key.
    type KeyTree struct {
    	root      []byte
    	accountID string
    }

    func validLabel(s string) bool { return len(s) >= 1 && len(s) <= 255 }

    func NewKeyTree(root []byte, accountID string) (*KeyTree, error) {
    	if len(root) != rootSize {
    		return nil, ErrBadRoot
    	}
    	if !validLabel(accountID) {
    		return nil, ErrBadLabel
    	}
    	return &KeyTree{root: bytes.Clone(root), accountID: accountID}, nil
    }

    // info encodes labels unambiguously: prefix, count, then each label with a
    // one-byte length.
    func info(labels ...string) string {
    	b := []byte(infoPrefix)
    	b = append(b, byte(len(labels)))
    	for _, l := range labels {
    		b = append(b, byte(len(l)))
    		b = append(b, l...)
    	}
    	return string(b)
    }

    func (t *KeyTree) derive(labels ...string) ([]byte, error) {
    	for _, l := range labels {
    		if !validLabel(l) {
    			return nil, ErrBadLabel
    		}
    	}
    	return hkdf.Key(sha256.New, t.root, []byte(t.accountID), info(labels...), keySize)
    }

    func (t *KeyTree) Derive(purpose string, path ...string) ([]byte, error) {
    	if !purposes[purpose] {
    		return nil, ErrUnknownPurpose
    	}
    	if len(path) > maxPath {
    		return nil, ErrBadLabel
    	}
    	return t.derive(append([]string{purpose}, path...)...)
    }

    func (t *KeyTree) Child(label string) (*KeyTree, error) {
    	root, err := t.derive(subtree, label)
    	if err != nil {
    		return nil, err
    	}
    	return &KeyTree{root: root, accountID: t.accountID}, nil
    }

    func main() {
    	root := make([]byte, rootSize)
    	tree, err := NewKeyTree(root, "acct-42")
    	fmt.Println(err)
    	enc, _ := tree.Derive("vault-enc", "vault-7")
    	mac, _ := tree.Derive("vault-mac", "vault-7")
    	fmt.Printf("enc %x\nmac %x\n", enc, mac)
    	_, err = tree.Derive(subtree, "vault-7")
    	fmt.Println(err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/hkdf"
    	"crypto/sha256"
    	"encoding/hex"
    	"errors"
    	"fmt"
    	"strings"
    	"testing"
    )

    // ---- a reference implementation of the spec ----

    func refInfo(labels ...string) string {
    	b := []byte("keybox/kdf/v1")
    	b = append(b, byte(len(labels)))
    	for _, l := range labels {
    		b = append(b, byte(len(l)))
    		b = append(b, l...)
    	}
    	return string(b)
    }

    func ref(root []byte, account string, labels ...string) []byte {
    	k, err := hkdf.Key(sha256.New, root, []byte(account), refInfo(labels...), 32)
    	if err != nil {
    		panic(err)
    	}
    	return k
    }

    var testRoot = bytes.Repeat([]byte{0x11}, 32)

    func mustTree(t *testing.T, root []byte, account string) *KeyTree {
    	t.Helper()
    	tree, err := NewKeyTree(root, account)
    	if err != nil || tree == nil {
    		t.Fatalf("NewKeyTree(32-byte root, %q) = %v, %v", account, tree, err)
    	}
    	return tree
    }

    func mustDerive(t *testing.T, tree *KeyTree, purpose string, path ...string) []byte {
    	t.Helper()
    	k, err := tree.Derive(purpose, path...)
    	if err != nil {
    		t.Fatalf("Derive(%q, %q) returned error %v", purpose, path, err)
    	}
    	if len(k) != keySize {
    		t.Fatalf("Derive(%q, %q) returned %d bytes, want %d", purpose, path, len(k), keySize)
    	}
    	return k
    }

    func TestReferenceMatchesRFC5869(t *testing.T) {
    	// Sanity check for the reference: RFC 5869, test case 1.
    	salt, _ := hex.DecodeString("000102030405060708090a0b0c")
    	info, _ := hex.DecodeString("f0f1f2f3f4f5f6f7f8f9")
    	okm, _ := hkdf.Key(sha256.New, bytes.Repeat([]byte{0x0b}, 22), salt, string(info), 42)
    	if want := "3cb25f25faacd57a90434f64d0362f2a2d2d0a90cf1a5a4c5db02d56ecc4c5bf34007208d5b887185865"; hex.EncodeToString(okm) != want {
    		t.Fatalf("crypto/hkdf doesn't match RFC 5869 test case 1")
    	}
    }

    func TestDeriveMatchesSpec(t *testing.T) {
    	tree := mustTree(t, testRoot, "acct-42")
    	cases := [][]string{
    		{"vault-enc"},
    		{"vault-enc", "vault-7"},
    		{"vault-mac", "vault-7"},
    		{"search-index", "vault-7", "2026"},
    		{"export", "a", "b", "c", "d"},
    		{"vault-enc", strings.Repeat("x", 255)},
    	}
    	for _, labels := range cases {
    		got := mustDerive(t, tree, labels[0], labels[1:]...)
    		if want := ref(testRoot, "acct-42", labels...); !bytes.Equal(got, want) {
    			t.Errorf("Derive(%.40q) = %x, want %x", labels, got, want)
    		}
    	}
    }

    func TestChildMatchesSpec(t *testing.T) {
    	tree := mustTree(t, testRoot, "acct-42")
    	child, err := tree.Child("vault-7")
    	if err != nil || child == nil {
    		t.Fatalf("Child(\"vault-7\") = %v, %v", child, err)
    	}
    	childRoot := ref(testRoot, "acct-42", "subtree", "vault-7")
    	if got, want := mustDerive(t, child, "vault-enc"), ref(childRoot, "acct-42", "vault-enc"); !bytes.Equal(got, want) {
    		t.Errorf("Child(\"vault-7\").Derive(\"vault-enc\") = %x, want %x", got, want)
    	}
    	grand, err := child.Child("item-3")
    	if err != nil || grand == nil {
    		t.Fatalf("Child(\"vault-7\").Child(\"item-3\") = %v, %v", grand, err)
    	}
    	grandRoot := ref(childRoot, "acct-42", "subtree", "item-3")
    	if got, want := mustDerive(t, grand, "export"), ref(grandRoot, "acct-42", "export"); !bytes.Equal(got, want) {
    		t.Errorf("grandchild Derive(\"export\") = %x, want %x", got, want)
    	}
    }

    func TestDomainSeparation(t *testing.T) {
    	tree := mustTree(t, testRoot, "acct-42")
    	child, _ := tree.Child("vault-7")
    	pairs := []struct {
    		name string
    		a, b []byte
    	}{
    		{"purposes", mustDerive(t, tree, "vault-enc", "v"), mustDerive(t, tree, "vault-mac", "v")},
    		{"label boundaries ab|c vs a|bc", mustDerive(t, tree, "vault-enc", "ab", "c"), mustDerive(t, tree, "vault-enc", "a", "bc")},
    		{"separator inside labels", mustDerive(t, tree, "vault-enc", "a/b", "c"), mustDerive(t, tree, "vault-enc", "a", "b/c")},
    		{"path prefix", mustDerive(t, tree, "vault-enc", "v"), mustDerive(t, tree, "vault-enc", "v", "v")},
    		{"child vs path", mustDerive(t, child, "vault-enc"), mustDerive(t, tree, "vault-enc", "vault-7")},
    		{"accounts", mustDerive(t, tree, "vault-enc"), mustDerive(t, mustTree(t, testRoot, "acct-43"), "vault-enc")},
    	}
    	for _, p := range pairs {
    		if bytes.Equal(p.a, p.b) {
    			t.Errorf("%s: two different derivations gave the same key %x", p.name, p.a)
    		}
    	}
    	seen := map[string]string{}
    	for i := range 200 {
    		for _, purpose := range []string{"vault-enc", "vault-mac", "search-index", "export"} {
    			k := string(mustDerive(t, tree, purpose, fmt.Sprint(i)))
    			id := purpose + "/" + fmt.Sprint(i)
    			if prev, ok := seen[k]; ok {
    				t.Fatalf("%s and %s derived the same key", prev, id)
    			}
    			seen[k] = id
    		}
    	}
    }

    func TestReservedAndUnknownPurposes(t *testing.T) {
    	tree := mustTree(t, testRoot, "acct-42")
    	for _, p := range []string{"subtree", "", "vault", "VAULT-ENC", "vault-enc "} {
    		if k, err := tree.Derive(p, "vault-7"); !errors.Is(err, ErrUnknownPurpose) {
    			t.Errorf("Derive(%q, \"vault-7\") = %x, %v, want ErrUnknownPurpose", p, k, err)
    		}
    	}
    }

    func TestBadLabels(t *testing.T) {
    	tree := mustTree(t, testRoot, "acct-42")
    	bad := [][]string{
    		{""},
    		{"vault-7", ""},
    		{strings.Repeat("x", 256)},
    		{"a", "b", "c", "d", "e"},
    	}
    	for _, path := range bad {
    		if k, err := tree.Derive("vault-enc", path...); !errors.Is(err, ErrBadLabel) {
    			t.Errorf("Derive(\"vault-enc\", %.30q) = %x, %v, want ErrBadLabel", path, k, err)
    		}
    	}
    	for _, l := range []string{"", strings.Repeat("y", 256)} {
    		if c, err := tree.Child(l); !errors.Is(err, ErrBadLabel) {
    			t.Errorf("Child(%.30q) = %v, %v, want ErrBadLabel", l, c, err)
    		}
    	}
    }

    func TestNewKeyTreeValidation(t *testing.T) {
    	for _, n := range []int{0, 16, 31, 33, 64} {
    		if tree, err := NewKeyTree(make([]byte, n), "acct-42"); !errors.Is(err, ErrBadRoot) {
    			t.Errorf("NewKeyTree(%d-byte root) = %v, %v, want ErrBadRoot", n, tree, err)
    		}
    	}
    	for _, a := range []string{"", strings.Repeat("a", 256)} {
    		if tree, err := NewKeyTree(testRoot, a); !errors.Is(err, ErrBadLabel) {
    			t.Errorf("NewKeyTree(root, %.20q) = %v, %v, want ErrBadLabel", a, tree, err)
    		}
    	}
    }

    func TestNoAliasing(t *testing.T) {
    	root := bytes.Clone(testRoot)
    	tree := mustTree(t, root, "acct-42")
    	before := mustDerive(t, tree, "vault-enc")
    	clear(root) // the caller wipes its copy of the root key
    	if after := mustDerive(t, tree, "vault-enc"); !bytes.Equal(before, after) {
    		t.Errorf("wiping the caller's root slice changed derived keys: NewKeyTree must keep its own copy")
    	}
    	k := mustDerive(t, tree, "export")
    	clear(k)
    	if again := mustDerive(t, tree, "export"); !bytes.Equal(again, ref(testRoot, "acct-42", "export")) {
    		t.Errorf("wiping a returned key changed the next Derive result")
    	}
    }
---

Keybox needs many keys per account: an encryption key and a MAC key per vault, a key
for the encrypted search index, a key for exports, and later ones nobody has thought
of yet. Storing each separately is a backup nightmare. Instead, every key is
**derived** from one 32-byte account root with HKDF, along a path of labels, so that
learning one derived key reveals nothing about any other.

That only holds if two different paths can **never** produce the same HKDF input.
Implement `KeyTree` to this spec:

- `NewKeyTree(root, accountID)`: `root` must be exactly 32 bytes (`ErrBadRoot`);
  `accountID` must be a valid label (`ErrBadLabel`). Keep a **copy** of `root`.
- `Derive(purpose, path...)` returns the 32-byte key
  `HKDF-SHA256(secret = root, salt = accountID, info = encode(purpose, path...))`.
  - `purpose` must be one of `purposes`, else `ErrUnknownPurpose`. The reserved
    `"subtree"` is **not** in that list.
  - At most `maxPath` (4) path labels, and every label 1 to 255 bytes, else `ErrBadLabel`.
- `Child(label)` returns a new `KeyTree` for the same account whose root is
  `HKDF-SHA256(root, accountID, encode("subtree", label))`, or `ErrBadLabel`.

where

```
encode(l1, ..., ln) = "keybox/kdf/v1" || byte(n) || byte(len(l1)) || l1 || ... || byte(len(ln)) || ln
```

## Example

```go
tree, _ := NewKeyTree(bytes.Repeat([]byte{0x11}, 32), "acct-42")
tree.Derive("vault-enc", "vault-7")   // 9926b0392ab1a70bdc0aa53ad2e992dfef8b6cf41b3c2b9927a4471fe5339a9a
tree.Derive("vault-mac", "vault-7")   // a different, independent key
tree.Derive("subtree", "vault-7")     // nil, ErrUnknownPurpose
vault, _ := tree.Child("vault-7")
vault.Derive("vault-enc")             // yet another key, unrelated to both above
```

## Constraints

- The tests compare your keys with a reference implementation of this spec (checked
  against RFC 5869's HKDF test vector), then check domain separation directly:
  different purposes, `["ab","c"]` vs `["a","bc"]`, labels containing `/`, path
  prefixes, child trees vs paths, and different accounts must all give different keys.
- They also try unknown and reserved purposes, empty, 256-byte and too many labels,
  bad root sizes, and wiping the caller's root slice after `NewKeyTree`.
- HKDF is the right tool here because the root is already a uniformly random key. A
  password would first need a slow KDF such as PBKDF2.
