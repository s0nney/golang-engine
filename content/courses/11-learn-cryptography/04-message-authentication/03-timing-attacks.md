---
title: Timing Attacks
quiz:
  - question: Which comparison is appropriate for checking a user-supplied API key against the stored secret?
    options:
      - text: '`provided == stored`'
      - text: '`bytes.Equal([]byte(provided), []byte(stored))`'
      - text: '`subtle.ConstantTimeCompare([]byte(provided), []byte(stored)) == 1`'
        correct: true
      - text: '`strings.EqualFold(provided, stored)`'
    explanation: |
      `==`, `bytes.Equal` and `strings.EqualFold` may stop at the first mismatching
      byte, so their running time depends on how much of the guess was right.
      `subtle.ConstantTimeCompare` (and `hmac.Equal`, which uses it) looks at every byte
      regardless. Better still, store a hash of the key and compare hashes.
  - question: Does `subtle.ConstantTimeCompare` hide the *length* of the secret?
    options:
      - text: Yes, it's constant-time in every respect
      - text: No; it returns 0 immediately if the lengths differ, so use it on fixed-length values such as MACs or hashes
        correct: true
      - text: Only for strings, not byte slices
    explanation: |
      The docs say so: the time depends on the lengths, not the contents. That's fine
      for MACs and hashes, whose length is public anyway. For secrets of varying length,
      hash both sides first and compare the fixed-size digests.
exercise:
  starter: |
    package main

    import "fmt"

    // constantTimeEqual reports whether a and b are equal. Its running time
    // must depend only on the lengths, never on where the first difference is:
    // no early return inside the loop, and no branching on the data.
    func constantTimeEqual(a, b []byte) bool {
    	// ? This version returns as soon as it finds a difference.
    	if len(a) != len(b) {
    		return false
    	}
    	for i := range a {
    		if a[i] != b[i] {
    			return false
    		}
    	}
    	return len(a) > 0
    }

    func main() {
    	tag := []byte{0xde, 0xad, 0xbe, 0xef}
    	fmt.Println(constantTimeEqual(tag, []byte{0xde, 0xad, 0xbe, 0xef}))
    	fmt.Println(constantTimeEqual(tag, []byte{0xde, 0xad, 0xbe, 0xee}))
    	fmt.Println(constantTimeEqual(nil, []byte{}))
    }
  solution: |
    package main

    import "fmt"

    func constantTimeEqual(a, b []byte) bool {
    	if len(a) != len(b) {
    		return false
    	}
    	var diff byte
    	for i := range a {
    		diff |= a[i] ^ b[i]
    	}
    	return diff == 0
    }

    func main() {
    	tag := []byte{0xde, 0xad, 0xbe, 0xef}
    	fmt.Println(constantTimeEqual(tag, []byte{0xde, 0xad, 0xbe, 0xef}))
    	fmt.Println(constantTimeEqual(tag, []byte{0xde, 0xad, 0xbe, 0xee}))
    	fmt.Println(constantTimeEqual(nil, []byte{}))
    }
  tests: |
    package main

    import (
    	"go/ast"
    	"go/parser"
    	"go/token"
    	"testing"
    )

    func TestConstantTimeEqualResults(t *testing.T) {
    	for _, tt := range []struct {
    		a, b []byte
    		want bool
    	}{
    		{[]byte("keybox"), []byte("keybox"), true},
    		{[]byte("keybox"), []byte("keyboy"), false},
    		{[]byte("keybox"), []byte("Keybox"), false},
    		{[]byte("keybox"), []byte("keybo"), false},
    		{[]byte{}, []byte{}, true},
    		{nil, []byte{}, true},
    		{nil, []byte{0}, false},
    		{[]byte{0x80}, []byte{0x00}, false},
    		{[]byte{1, 2, 3, 4}, []byte{1, 2, 3, 4}, true},
    	} {
    		if got := constantTimeEqual(tt.a, tt.b); got != tt.want {
    			t.Errorf("constantTimeEqual(%x, %x) = %v, want %v", tt.a, tt.b, got, tt.want)
    		}
    	}
    }

    // The loop must not return early or branch on the data.
    func TestNoEarlyExit(t *testing.T) {
    	f, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
    	if err != nil {
    		t.Fatal(err)
    	}
    	for _, d := range f.Decls {
    		fn, ok := d.(*ast.FuncDecl)
    		if !ok || fn.Name.Name != "constantTimeEqual" {
    			continue
    		}
    		ast.Inspect(fn.Body, func(n ast.Node) bool {
    			loop, ok := n.(*ast.RangeStmt)
    			if !ok {
    				if f, ok := n.(*ast.ForStmt); ok {
    					loopBody(t, f.Body)
    				}
    				return true
    			}
    			loopBody(t, loop.Body)
    			return true
    		})
    	}
    }

    func loopBody(t *testing.T, body *ast.BlockStmt) {
    	ast.Inspect(body, func(n ast.Node) bool {
    		switch n.(type) {
    		case *ast.IfStmt, *ast.ReturnStmt, *ast.BranchStmt, *ast.SwitchStmt:
    			t.Errorf("constantTimeEqual's loop contains an if, return, break or switch; accumulate differences with |= and ^ instead")
    			return false
    		}
    		return true
    	})
    }
---

In 2010, researchers showed that many OAuth and OpenID libraries compared HMAC tags
with an ordinary string comparison, and that the difference in response time could be
measured over a network. Comparing secrets the normal way leaks information.

## How an ordinary comparison leaks

`bytes.Equal`, `==` on strings and most hand-written loops stop at the **first
difference**. Comparing two 32-byte tags that differ in byte 0 is a little faster than
comparing two that differ in byte 31.

The difference is nanoseconds, but an attacker can send the same request many times and
average out the noise. If the time reveals how many leading bytes of a guessed tag were
correct, the attacker no longer has to guess all 32 bytes at once: they can make
progress **one byte at a time**, which turns an impossible search into a feasible one.

This is a **side channel**: information leaking not through the output but through how
the computation behaved. Timing is the classic one; others include power draw, cache
state and even sound.

## Constant-time comparison

The fix is to make the running time independent of the secret contents. Instead of
returning at the first mismatch, look at every byte and **accumulate** differences:

```go
var diff byte
for i := range a {
	diff |= a[i] ^ b[i] // nonzero bits appear wherever the bytes differ
}
return diff == 0
```

XOR gives zero only for equal bytes, and OR-ing everything together remembers any
difference. The loop runs the same number of iterations with the same operations
whatever the data.

The standard library has this ready-made, and you should use it rather than your own:

- **`hmac.Equal(a, b)`**: for MAC tags.
- **`subtle.ConstantTimeCompare(a, b)`**: returns `1` if equal, `0` otherwise. `crypto/subtle`
  also has constant-time selection and byte comparison helpers for writing
  constant-time code.
- **`subtle.WithDataIndependentTiming(f)`** (Go 1.24) runs `f` with CPU features enabled
  that guarantee data-independent instruction timing, on the architectures that support
  them (arm64's DIT, for example). It doesn't make variable-time code constant-time;
  it's for code that's already written to be.

Both comparison functions return early if the **lengths** differ. Lengths of MACs and
hashes are public, so that's fine. If a secret's length is itself sensitive, hash
both sides to fixed-size digests first.

## Where timing leaks hide

- **Comparing tokens, API keys, MACs, password hashes.** Use constant-time comparison,
  or better, look up by a hash (`sha256(token)`) in a map or database and compare
  digests.
- **Early exits in verification.** "Wrong user" answering faster than "wrong password"
  tells an attacker which usernames exist. Do the same work in both cases.
- **Your own crypto arithmetic.** Branches and table lookups indexed by secret data leak
  through timing and caches. This is a big reason not to implement primitives yourself;
  Go's implementations are written to avoid it.

Constant time is also not a license to be sloppy elsewhere: rate-limit verification
endpoints too, so nobody gets millions of attempts in the first place.

## Your task

Rewrite `constantTimeEqual(a, b)` so its loop has **no early exit and no branches on the
data**. Return `false` immediately if the lengths differ (that's public), then
accumulate `a[i] ^ b[i]` with `|=` and compare the result with zero after the loop. Two
empty slices are equal.

The tests check the results, and parse your code to make sure the loop contains no
`if`, `return`, `break` or `switch`. In real code you'd simply call
`subtle.ConstantTimeCompare`, but it's worth writing once to see there's no magic.
