---
title: Endless Chant
difficulty: medium
after: abstraction
hints:
  - '`Read(p)` fills **as much of `p` as it can** (at most `len(p)` bytes), returns how many bytes it wrote, and returns `0, io.EOF` once there''s nothing left. It must never build the whole chant in memory: `times` can be astronomically large.'
  - 'Keep two pieces of state: how many repetitions are still to come, and your position inside the current line (`words + "\n"`). In a loop, `copy` from the current line at that position into `p[n:]`, advance, and move on to the next repetition when the line is used up.'
  - '`countLines` should only depend on `io.Reader`. Read into a buffer in a loop, count the `''\n''` bytes with `bytes.Count`, remember whether the last byte you saw was a newline, and stop at `io.EOF`. Any other error is returned.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"os"
    )

    type Chant struct {
    }

    func NewChant(words string, times int) *Chant {
    	return &Chant{}
    }

    func (c *Chant) Read(p []byte) (int, error) {
    	return 0, io.EOF
    }

    func countLines(r io.Reader) (int, error) {
    	return 0, nil
    }

    func main() {
    	io.Copy(os.Stdout, NewChant("Wake, O wyrm!", 3))
    	// want:
    	// Wake, O wyrm!
    	// Wake, O wyrm!
    	// Wake, O wyrm!
    	fmt.Println(countLines(NewChant("Om", 1_000_000))) // want 1000000 <nil>
    }
  solution: |
    package main

    import (
    	"bytes"
    	"errors"
    	"fmt"
    	"io"
    	"os"
    )

    // Chant is an io.Reader that yields words followed by a newline, times
    // times, without ever building the whole text.
    type Chant struct {
    	line []byte // words + "\n"
    	pos  int    // bytes of line already read in the current repetition
    	left int    // repetitions not yet finished
    }

    // NewChant returns a reader of times lines, each holding words.
    func NewChant(words string, times int) *Chant {
    	return &Chant{line: []byte(words + "\n"), left: max(0, times)}
    }

    // Read implements io.Reader.
    func (c *Chant) Read(p []byte) (int, error) {
    	if c.left == 0 {
    		return 0, io.EOF
    	}
    	n := 0
    	for n < len(p) && c.left > 0 {
    		k := copy(p[n:], c.line[c.pos:])
    		n += k
    		c.pos += k
    		if c.pos == len(c.line) {
    			c.pos = 0
    			c.left--
    		}
    	}
    	return n, nil
    }

    // countLines counts the lines r yields. A last line without a trailing
    // newline still counts.
    func countLines(r io.Reader) (int, error) {
    	buf := make([]byte, 32*1024)
    	lines, lastWasNewline, empty := 0, false, true
    	for {
    		n, err := r.Read(buf)
    		if n > 0 {
    			lines += bytes.Count(buf[:n], []byte{'\n'})
    			lastWasNewline = buf[n-1] == '\n'
    			empty = false
    		}
    		if errors.Is(err, io.EOF) {
    			break
    		}
    		if err != nil {
    			return 0, err
    		}
    	}
    	if !empty && !lastWasNewline {
    		lines++
    	}
    	return lines, nil
    }

    func main() {
    	io.Copy(os.Stdout, NewChant("Wake, O wyrm!", 3))
    	fmt.Println(countLines(NewChant("Om", 1_000_000)))
    }
  tests: |
    package main

    import (
    	"errors"
    	"io"
    	"strings"
    	"testing"
    	"testing/iotest"
    	"time"
    )

    func TestChantContent(t *testing.T) {
    	tests := []struct {
    		words string
    		times int
    	}{
    		{"Wake, O wyrm!", 3},
    		{"Om", 1},
    		{"x", 1000},
    		{"", 4},
    		{"never heard", 0},
    		{"never heard", -2},
    		{strings.Repeat("By scale and flame, ", 400), 3},
    	}
    	for _, tt := range tests {
    		want := strings.Repeat(tt.words+"\n", max(0, tt.times))
    		got, err := io.ReadAll(NewChant(tt.words, tt.times))
    		if err != nil || string(got) != want {
    			t.Errorf("reading NewChant(%q, %d) gave %q (%d bytes), %v, want %q (%d bytes), <nil>",
    				short(tt.words), tt.times, short(string(got)), len(got), err, short(want), len(want))
    			continue
    		}
    		// iotest.TestReader reads with many buffer sizes and checks io.Reader's rules.
    		if err := iotest.TestReader(NewChant(tt.words, tt.times), []byte(want)); err != nil {
    			t.Errorf("NewChant(%q, %d) breaks the io.Reader contract: %v", short(tt.words), tt.times, err)
    		}
    	}
    }

    func short(s string) string {
    	if len(s) > 30 {
    		return s[:30] + "..."
    	}
    	return s
    }

    func TestChantOneByteAtATime(t *testing.T) {
    	got, err := io.ReadAll(iotest.OneByteReader(NewChant("fire", 2)))
    	if string(got) != "fire\nfire\n" || err != nil {
    		t.Errorf("reading NewChant(\"fire\", 2) one byte at a time gave %q, %v, want \"fire\\nfire\\n\", <nil>", got, err)
    	}
    }

    func TestChantIsLazy(t *testing.T) {
    	done := make(chan string, 1)
    	go func() {
    		buf := make([]byte, 20)
    		n, _ := io.ReadFull(NewChant("abc", 1<<50), buf)
    		done <- string(buf[:n])
    	}()
    	select {
    	case got := <-done:
    		if want := "abc\nabc\nabc\nabc\nabc\n"; got != want {
    			t.Errorf("first 20 bytes of NewChant(\"abc\", 1<<50) = %q, want %q", got, want)
    		}
    	case <-time.After(2 * time.Second):
    		t.Fatal("reading 20 bytes of NewChant(\"abc\", 1<<50) didn't finish: don't build the whole chant up front")
    	}
    }

    type failingReader struct{ sent bool }

    var errScrollBurned = errors.New("the scroll burned")

    func (f *failingReader) Read(p []byte) (int, error) {
    	if f.sent {
    		return 0, errScrollBurned
    	}
    	f.sent = true
    	return copy(p, "line one\nline two\n"), nil
    }

    func TestCountLines(t *testing.T) {
    	tests := []struct {
    		name string
    		r    io.Reader
    		want int
    	}{
    		{"empty", strings.NewReader(""), 0},
    		{"one line with newline", strings.NewReader("hail\n"), 1},
    		{"one line without newline", strings.NewReader("hail"), 1},
    		{"last line without newline", strings.NewReader("a\nb\nc"), 3},
    		{"blank lines count", strings.NewReader("\n\n\n"), 3},
    		{"one byte at a time", iotest.OneByteReader(strings.NewReader("x\ny\nz")), 3},
    		{"half reads", iotest.HalfReader(strings.NewReader("ab\ncd\nef\n")), 3},
    		{"data then EOF together", iotest.DataErrReader(strings.NewReader("p\nq")), 2},
    		{"a chant", NewChant("Wake, O wyrm!", 7), 7},
    		{"a long chant", NewChant("Om", 1_000_000), 1_000_000},
    	}
    	for _, tt := range tests {
    		got, err := countLines(tt.r)
    		if got != tt.want || err != nil {
    			t.Errorf("%s: countLines = %d, %v, want %d, <nil>", tt.name, got, err, tt.want)
    		}
    	}
    }

    func TestCountLinesReportsErrors(t *testing.T) {
    	_, err := countLines(&failingReader{})
    	if !errors.Is(err, errScrollBurned) {
    		t.Errorf("countLines on a reader that fails with %q returned error %v, want that error", errScrollBurned, err)
    	}
    	_, err = countLines(iotest.TimeoutReader(strings.NewReader("a\nb\n")))
    	if !errors.Is(err, iotest.ErrTimeout) {
    		t.Errorf("countLines on a reader that times out returned error %v, want %v", err, iotest.ErrTimeout)
    	}
    }
---

The cultists of the Ember Peak chant the same words over and over to wake their
dragon, sometimes for a million verses. Model the chant as an **`io.Reader`**,
Go's one-method abstraction for "a stream of bytes", so it plugs into
everything in the standard library: `io.Copy`, `io.ReadAll`, `bufio.Scanner`,
files, network connections...

1. `NewChant(words, times)` returns a `*Chant` whose content is `words`
   followed by a newline, repeated `times` times. A `times` of 0 or less is an
   empty chant.
2. `(*Chant).Read(p)` implements `io.Reader`: it copies the next bytes of the
   chant into `p` (at most `len(p)` of them), returns how many it copied, and
   returns `0, io.EOF` when the chant is over. It must work with **any** buffer
   size, including 1 byte, and must **not** build the whole chant in memory.
3. `countLines(r)` counts the lines in **any** `io.Reader`. A final line
   without a trailing newline still counts, and an empty stream has 0 lines.
   If `r` fails with an error other than `io.EOF`, return that error.

## Example

```go
io.Copy(os.Stdout, NewChant("Wake, O wyrm!", 3))
// Wake, O wyrm!
// Wake, O wyrm!
// Wake, O wyrm!

countLines(NewChant("Om", 1_000_000))       // 1000000, nil
countLines(strings.NewReader("a\nb\nc"))    // 3, nil
countLines(strings.NewReader(""))           // 0, nil
```

## Constraints

- A test reads the first 20 bytes of `NewChant("abc", 1<<50)`. Building that
  string would need a petabyte.
- The tests use `testing/iotest` to read your chant with awkward buffer sizes,
  and to feed `countLines` readers that return data in dribs and drabs or fail.
