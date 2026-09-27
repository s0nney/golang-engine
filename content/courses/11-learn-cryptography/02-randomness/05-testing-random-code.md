---
title: Testing Random Code
quiz:
  - question: 'A test for `sealSecret` (which uses a random nonce) says `if sealed != wantHex { t.Fatal(...) }`, with `wantHex` copied from one run. What''s wrong?'
    options:
      - text: Nothing, as long as it passes once
      - text: It will fail on almost every run, because the nonce, and so the ciphertext, changes each time; test properties like round-tripping and tamper detection instead
        correct: true
      - text: It should use `bytes.Equal` instead of `!=`
      - text: It needs `t.Parallel()`
    explanation: |
      Randomized encryption is *supposed* to give a different output every time. Test
      what must always hold (decrypt gives the plaintext back, a flipped bit fails, the
      wrong key fails), and use published test vectors for the deterministic parts.
  - question: Why can't a test that calls `cryptotest.SetGlobalRandom` use `t.Parallel()`?
    options:
      - text: Because `SetGlobalRandom` replaces the randomness source for the whole process, which would leak into other tests running at the same time
        correct: true
      - text: Because crypto code isn't safe for concurrent use
      - text: Because parallel tests run in separate processes
    explanation: |
      The setting is global. Any test running in parallel would suddenly get
      deterministic "randomness" too, and the order of reads between tests would make
      the values unpredictable anyway. The docs forbid it for both reasons.
---

Randomness is great for security and awkward for tests. You can't assert
`ciphertext == "3f9a..."` if the ciphertext changes every run. Here are the four
techniques this course's exercises use, and you'll want them in your own code too.

## 1. Test properties, not bytes

Most crypto code has properties that hold whatever the random values are:

- **Round trip:** `open(seal(x)) == x`.
- **Tamper detection:** flip any bit of the output and `open` fails.
- **Wrong key fails:** opening with a different key returns an error.
- **Freshness:** sealing the same plaintext twice gives different outputs, and the
  nonces differ.

```go
func TestSealTamper(t *testing.T) {
	key := newKey()
	sealed, err := sealSecret(key, "github-token", []byte("ghp_hunter2"))
	if err != nil {
		t.Fatal(err)
	}
	for i := range len(sealed) * 8 {
		bad := bytes.Clone(sealed)
		bad[i/8] ^= 1 << (i % 8)
		if _, err := openSecret(key, "github-token", bad); err == nil {
			t.Fatalf("flipping bit %d wasn't detected", i)
		}
	}
}
```

## 2. Known-answer tests for the deterministic parts

Hashes, MACs, key derivation and Ed25519 signatures are deterministic. Standards
publish **test vectors** for them: fixed inputs with the exact expected output. Use
them. A vector from an RFC catches mistakes that round-trip tests can't, like using
the wrong hash function consistently on both sides. You'll use vectors from RFC 4231
(HMAC), RFC 5869 (HKDF), RFC 7748 (X25519) and RFC 8032 (Ed25519) later in the course.

## 3. Inject the reader

When *you* consume random bytes, as `randomString` did last lesson, accept an
`io.Reader`. Production passes `rand.Reader`; tests pass a `bytes.Reader` with exactly
the bytes they want. (The standard library used to work this way too, but since Go
1.26 most `crypto/...` functions ignore the reader you pass and always use secure
randomness, so this trick only works for your own code.)

## 4. `testing/cryptotest.SetGlobalRandom`

For everything else, Go 1.26 added `testing/cryptotest`. One call makes `crypto/rand`
**and** all the randomness used internally by `crypto/...` packages deterministic for
the rest of the test:

```go
func TestKeyGenerationIsSeeded(t *testing.T) {
	cryptotest.SetGlobalRandom(t, 42)
	a := newKey()
	cryptotest.SetGlobalRandom(t, 42) // reset the stream
	b := newKey()
	if !bytes.Equal(a, b) {
		t.Error("newKey didn't use crypto/rand")
	}
}
```

The tests in [Keys, Tokens and Nonces](/courses/learn-cryptography/randomness/keys-tokens-and-nonces)
used exactly this trick to prove your code called
`crypto/rand` and not `math/rand`.

Some caveats from the docs:

- It affects the **whole process**, so it can't be used in parallel tests.
- How algorithms consume randomness isn't specified and can change between Go
  versions, so don't hard-code, say, an expected ECDSA signature produced under a
  fixed seed. Use it for reproducibility, and combine it with property checks.
- It's for **tests only**. There's no way to make production randomness
  deterministic, and that's deliberate.

## Where `math/rand/v2` still belongs

Non-security randomness is fine, and reproducible seeds are an asset there:
simulations, load-test traffic, retry jitter, shuffling test cases. A seeded
`rand.New(rand.NewChaCha8(seed))` in a fuzzer or property test lets you print the seed
on failure and replay it exactly. Just never let those values become keys, nonces,
salts or tokens.
