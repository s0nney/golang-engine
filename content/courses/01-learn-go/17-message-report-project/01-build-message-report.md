---
title: 'Build It: Textio Message Report'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"strings"
    )

    type Report struct {
    	Messages int
    	Words int
    	Frequency map[string]int
    }

    func summarize(r io.Reader) (Report, error) {
    	return Report{Frequency: make(map[string]int)}, nil
    }

    func main() {
    	report, err := summarize(strings.NewReader("Go go!\n\nShip it\n"))
    	fmt.Println(report.Messages, report.Words, report.Frequency["go"], err)
    }
  solution: |
    package main

    import (
    	"bufio"
    	"fmt"
    	"io"
    	"strings"
    )

    type Report struct {
    	Messages int
    	Words int
    	Frequency map[string]int
    }

    func summarize(r io.Reader) (Report, error) {
    	report := Report{Frequency: make(map[string]int)}
    	scanner := bufio.NewScanner(r)
    	for scanner.Scan() {
    		words := strings.Fields(scanner.Text())
    		if len(words) == 0 { continue }
    		report.Messages++
    		for _, word := range words {
    			word = strings.ToLower(strings.Trim(word, ".,!?;:"))
    			if word == "" { continue }
    			report.Words++
    			report.Frequency[word]++
    		}
    	}
    	return report, scanner.Err()
    }

    func main() {
    	report, err := summarize(strings.NewReader("Go go!\n\nShip it\n"))
    	fmt.Println(report.Messages, report.Words, report.Frequency["go"], err)
    }
  tests: |
    package main

    import (
    	"errors"
    	"reflect"
    	"strings"
    	"testing"
    )

    var errRead = errors.New("read failed")
    type brokenReader struct{}
    func (brokenReader) Read([]byte) (int, error) { return 0, errRead }

    func TestReport(t *testing.T) {
    	for _, tc := range []struct {
    		input string
    		want Report
    	}{
    		{"Go go!\n\nShip it\n", Report{2, 4, map[string]int{"go": 2, "ship": 1, "it": 1}}},
    		{" \t\n", Report{0, 0, map[string]int{}}},
    		{"Hi,\t世界!\n!!!", Report{2, 2, map[string]int{"hi": 1, "世界": 1}}},
    	} {
    		got, err := summarize(strings.NewReader(tc.input))
    		if err != nil || !reflect.DeepEqual(got, tc.want) {
    			t.Errorf("summarize(%q) = %#v, %v; want %#v, nil", tc.input, got, err, tc.want)
    		}
    	}
    }

    func TestReadError(t *testing.T) {
    	if _, err := summarize(brokenReader{}); !errors.Is(err, errRead) {
    		t.Errorf("reader error = %v, want %v", err, errRead)
    	}
    }
---

Textio wants a report of the messages it sends. Assemble loops, strings, maps,
structs, and error handling into one small program. Start with the reader-based
function here; the same function can read a real file without changing its logic.

## Your assignment

Complete `summarize`. Read lines with `bufio.NewScanner(r)` and return a `Report`:

1. Count each line containing at least one whitespace-separated token as a message.
   Empty lines and lines containing only whitespace don't count.
2. Split each line with `strings.Fields`. Trim leading and trailing `. , ! ? ; :`
   punctuation using `strings.Trim(word, ".,!?;:")`, then lowercase it.
3. Skip tokens that become empty. Count the remaining words and their frequencies.
4. Always initialize `Frequency`, including for empty input. Return the scanner's
   error if reading fails, alongside whatever report you have built so far.

`bufio.Scanner` reads one line per successful `Scan()` call. `Text()` gives you that
line, and `Err()` tells you whether scanning stopped because of an error:

```go
scanner := bufio.NewScanner(r)
for scanner.Scan() {
	line := scanner.Text()
	_ = line // replace this with your report logic
}
return report, scanner.Err()
```

An `io.Reader` describes anything you can read bytes from. A `strings.Reader` is
useful for the browser demonstration; an `*os.File` also satisfies that interface.
Scanner limits line size by default, so an extremely long message produces an error;
don't silently report it as a successfully processed file.

**Run** should print `2 4 2 <nil>` once your implementation is finished. **Submit**
also checks empty input, tabs, punctuation-only tokens, Unicode, and reader errors.

## Take it to your terminal

After passing, replace `strings.NewReader` in your local `main` with a file opened
using `os.Open`. Check the open error, `defer file.Close()`, and pass the file to
`summarize`. Print the counts only after checking the returned error. That's a useful
separation: `main` chooses the input and handles failures; `summarize` does the work.

Next, take the final quiz. Then [Learn Object-Oriented Programming in Go](/courses/learn-oop)
will show how methods and interfaces help organize programs as they grow.
