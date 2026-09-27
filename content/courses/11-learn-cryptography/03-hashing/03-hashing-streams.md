---
title: Hashing Streams
quiz:
  - question: |
      What does this print?

      ```go
      h := sha256.New()
      h.Write([]byte("key"))
      a := h.Sum(nil)
      h.Write([]byte("box"))
      b := h.Sum(nil)
      fmt.Println(bytes.Equal(b, sha256.New().Sum(nil)), hex.EncodeToString(b) == fmt.Sprintf("%x", sha256.Sum256([]byte("keybox"))), len(a))
      ```
    options:
      - text: '`false false 32`'
      - text: '`false true 32`'
        correct: true
      - text: '`true false 64`'
      - text: '`false true 64`'
    explanation: |
      `Sum` doesn't reset or finalize the hash; you can keep writing afterwards. So `b`
      is the hash of everything written so far, `"keybox"`, and it matches `Sum256`.
      `Sum(nil)` appends the 32-byte digest to a nil slice, so `len(a)` is 32.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"strings"
    )

    // hashStream reads r to the end, computing its SHA-256 and SHA-512 digests
    // in a single pass without holding the whole input in memory. It returns
    // both digests as lowercase hex and the number of bytes read. If reading
    // fails, it returns the error.
    func hashStream(r io.Reader) (sum256, sum512 string, n int64, err error) {
    	// ?
    	return "", "", 0, nil
    }

    func main() {
    	backup := strings.NewReader(strings.Repeat("keybox vault backup block\n", 100_000))
    	s256, s512, n, err := hashStream(backup)
    	fmt.Println("bytes: ", n, err)
    	fmt.Println("sha256:", s256)
    	fmt.Println("sha512:", s512)
    }
  solution: |
    package main

    import (
    	"crypto/sha256"
    	"crypto/sha512"
    	"encoding/hex"
    	"fmt"
    	"io"
    	"strings"
    )

    func hashStream(r io.Reader) (sum256, sum512 string, n int64, err error) {
    	h256 := sha256.New()
    	h512 := sha512.New()
    	n, err = io.Copy(io.MultiWriter(h256, h512), r)
    	if err != nil {
    		return "", "", n, err
    	}
    	return hex.EncodeToString(h256.Sum(nil)), hex.EncodeToString(h512.Sum(nil)), n, nil
    }

    func main() {
    	backup := strings.NewReader(strings.Repeat("keybox vault backup block\n", 100_000))
    	s256, s512, n, err := hashStream(backup)
    	fmt.Println("bytes: ", n, err)
    	fmt.Println("sha256:", s256)
    	fmt.Println("sha512:", s512)
    }
  tests: |
    package main

    import (
    	"crypto/sha256"
    	"crypto/sha512"
    	"errors"
    	"fmt"
    	"io"
    	"strings"
    	"testing"
    	"testing/iotest"
    )

    func TestHashStream(t *testing.T) {
    	for _, s := range []string{"", "abc", "keybox", strings.Repeat("0123456789abcdef", 300_000)} {
    		// HalfReader forces many small reads, like a slow network.
    		for _, r := range []io.Reader{strings.NewReader(s), iotest.HalfReader(strings.NewReader(s))} {
    			got256, got512, n, err := hashStream(r)
    			if err != nil {
    				t.Fatalf("hashStream(%d bytes) error = %v", len(s), err)
    			}
    			want256 := fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
    			want512 := fmt.Sprintf("%x", sha512.Sum512([]byte(s)))
    			if got256 != want256 {
    				t.Errorf("hashStream(%d bytes) sha256 = %q, want %q", len(s), got256, want256)
    			}
    			if got512 != want512 {
    				t.Errorf("hashStream(%d bytes) sha512 = %q, want %q", len(s), got512, want512)
    			}
    			if n != int64(len(s)) {
    				t.Errorf("hashStream(%d bytes) n = %d, want %d", len(s), n, len(s))
    			}
    		}
    	}
    }

    func TestHashStreamError(t *testing.T) {
    	boom := errors.New("disk on fire")
    	r := io.MultiReader(strings.NewReader("partial data"), iotest.ErrReader(boom))
    	_, _, _, err := hashStream(r)
    	if !errors.Is(err, boom) {
    		t.Errorf("hashStream(reader that fails) error = %v, want %v", err, boom)
    	}
    }
---

`sha256.Sum256` needs the whole input in memory. Keybox's encrypted backups can be
gigabytes, arriving over the network or from disk. For those you hash a **stream**.

## `hash.Hash` is an `io.Writer`

`sha256.New()` returns a `hash.Hash`. Its key methods:

```go
type Hash interface {
	io.Writer             // Write(p []byte) (n int, err error): feed it more data
	Sum(b []byte) []byte  // append the digest of everything written so far to b
	Reset()               // start over
	Size() int            // digest length in bytes
	BlockSize() int       // internal block size
}
```

Because it's an `io.Writer`, everything in the `io` toolbox works with it. The idiom for
hashing a file or any reader is a single `io.Copy`:

```go
package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
)

func main() {
	h := sha256.New()
	n, err := io.Copy(h, strings.NewReader("abc"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d bytes, sha256 %x\n", n, h.Sum(nil))
}
```

```
3 bytes, sha256 ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
```

`io.Copy` moves data in 32 KB chunks, so memory use stays flat whatever the input size.
Hashing data in pieces gives exactly the same digest as hashing it all at once; how
the input was split doesn't matter.

For a file you'd `os.Open` it and pass the `*os.File` as the reader. For an HTTP
download, pass `resp.Body`.

## `Write` never fails

`hash.Hash`'s documentation promises that `Write` never returns an error. That's why
code like `h.Write(data)` without checking is fine for hashes (and MACs), even though
ignoring errors from other writers would be a bug. The only errors in the example above
come from the **reader**.

## `Sum` appends, and doesn't reset

Two gotchas in `Sum(b)`:

1. It **appends** the digest to `b` and returns the result. `h.Sum(nil)` gives you a
   fresh 32-byte slice. `h.Sum(prefix)` gives `prefix` followed by the digest, which is
   occasionally handy and occasionally a nasty surprise.
2. It **doesn't change the state**. You can call `Sum`, keep writing, and call it again
   for a running digest. Call `Reset()` to start from scratch.

## Hashing with several functions at once

`io.MultiWriter` fans every write out to several writers. That lets you compute
SHA-256 and SHA-512 in a single pass over the data, reading a large file only once:

```go
h256, h512 := sha256.New(), sha512.New()
n, err := io.Copy(io.MultiWriter(h256, h512), r)
```

Package repositories often publish several digests for the same file, so tools that
verify them do exactly this.

## Your task

Complete `hashStream(r)`. In **one pass** over `r`, compute the SHA-256 and SHA-512 of
everything it produces, and return both as lowercase hex along with the byte count from
`io.Copy`. If `io.Copy` returns an error, return it (the digest strings don't matter
then).

Don't use `io.ReadAll`: it would work for the tests, but the point is to handle inputs
bigger than memory. The tests feed data through deliberately awkward readers and
finish with one that fails partway.
