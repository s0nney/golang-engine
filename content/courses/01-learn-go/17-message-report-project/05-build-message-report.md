---
title: 'Build It: Textio Message Report'
quiz:
  - question: 'Why does `run` take an `out io.Writer` instead of printing straight to the screen with `fmt.Println`?'
    options:
      - text: Because `fmt.Println` doesn't work inside functions that return an error
      - text: So a test can pass a `strings.Builder` and check exactly what was written, while `main` passes `os.Stdout`
        correct: true
      - text: Because `io.Writer` output is faster than `fmt.Println`
    explanation: |
      `os.Stdout`, files and `strings.Builder` are all `io.Writer`s. Taking
      the writer as a parameter lets the caller decide where the report
      goes, which makes the function easy to test.
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"errors"
    	"flag"
    	"fmt"
    	"io"
    	"os"
    	"path/filepath"
    	"slices"
    	"strings"
    	"unicode"
    )

    // ---- Built in the previous lessons ----

    var ErrUsage = errors.New("usage: report [-top N] FILE")

    type WordCount struct {
    	Word  string
    	Count int
    }

    func parseArgs(args []string) (string, int, error) {
    	fs := flag.NewFlagSet("report", flag.ContinueOnError)
    	fs.SetOutput(io.Discard)
    	top := fs.Int("top", 3, "how many words to show")
    	if err := fs.Parse(args); err != nil {
    		return "", 0, err
    	}
    	if fs.NArg() != 1 {
    		return "", 0, ErrUsage
    	}
    	if *top < 1 {
    		return "", 0, fmt.Errorf("-top must be at least 1, got %d", *top)
    	}
    	return fs.Arg(0), *top, nil
    }

    func loadMessages(path string) ([]string, error) {
    	data, err := os.ReadFile(path)
    	if err != nil {
    		return nil, fmt.Errorf("load messages: %w", err)
    	}
    	var messages []string
    	for _, line := range strings.Split(string(data), "\n") {
    		line = strings.TrimSpace(line)
    		if line == "" {
    			continue
    		}
    		messages = append(messages, line)
    	}
    	return messages, nil
    }

    func countWords(messages []string) map[string]int {
    	counts := make(map[string]int)
    	for _, msg := range messages {
    		for _, word := range strings.Fields(msg) {
    			word = strings.ToLower(strings.Trim(word, ".,!?"))
    			if word == "" {
    				continue
    			}
    			counts[word]++
    		}
    	}
    	return counts
    }

    func countChars(messages []string) int {
    	chars := 0
    	for _, msg := range messages {
    		for _, r := range msg {
    			if !unicode.IsSpace(r) {
    				chars++
    			}
    		}
    	}
    	return chars
    }

    func topWords(counts map[string]int, n int) []WordCount {
    	var entries []WordCount
    	for word, count := range counts {
    		entries = append(entries, WordCount{word, count})
    	}
    	slices.SortFunc(entries, func(a, b WordCount) int {
    		if c := cmp.Compare(b.Count, a.Count); c != 0 {
    			return c
    		}
    		return cmp.Compare(a.Word, b.Word)
    	})
    	return entries[:min(n, len(entries))]
    }

    // ---- Your code ----

    // run parses args, reads the file and writes the report to out.
    func run(args []string, out io.Writer) error {
    	// ?
    	return errors.New("run is not finished yet")
    }

    func main() {
    	// Create a sample file so Run works anywhere.
    	path := filepath.Join(os.TempDir(), "messages.txt")
    	sample := "Lunch at noon?\nYes! Lunch at noon.\n\nSee you at lunch!\n"
    	if err := os.WriteFile(path, []byte(sample), 0o644); err != nil {
    		fmt.Println(err)
    		return
    	}
    	defer os.Remove(path)

    	// In your terminal you'd pass os.Args[1:] instead.
    	if err := run([]string{"-top", "3", path}, os.Stdout); err != nil {
    		fmt.Println("error:", err)
    	}
    }
  solution: |
    package main

    import (
    	"cmp"
    	"errors"
    	"flag"
    	"fmt"
    	"io"
    	"os"
    	"path/filepath"
    	"slices"
    	"strings"
    	"unicode"
    )

    // ---- Built in the previous lessons ----

    var ErrUsage = errors.New("usage: report [-top N] FILE")

    type WordCount struct {
    	Word  string
    	Count int
    }

    func parseArgs(args []string) (string, int, error) {
    	fs := flag.NewFlagSet("report", flag.ContinueOnError)
    	fs.SetOutput(io.Discard)
    	top := fs.Int("top", 3, "how many words to show")
    	if err := fs.Parse(args); err != nil {
    		return "", 0, err
    	}
    	if fs.NArg() != 1 {
    		return "", 0, ErrUsage
    	}
    	if *top < 1 {
    		return "", 0, fmt.Errorf("-top must be at least 1, got %d", *top)
    	}
    	return fs.Arg(0), *top, nil
    }

    func loadMessages(path string) ([]string, error) {
    	data, err := os.ReadFile(path)
    	if err != nil {
    		return nil, fmt.Errorf("load messages: %w", err)
    	}
    	var messages []string
    	for _, line := range strings.Split(string(data), "\n") {
    		line = strings.TrimSpace(line)
    		if line == "" {
    			continue
    		}
    		messages = append(messages, line)
    	}
    	return messages, nil
    }

    func countWords(messages []string) map[string]int {
    	counts := make(map[string]int)
    	for _, msg := range messages {
    		for _, word := range strings.Fields(msg) {
    			word = strings.ToLower(strings.Trim(word, ".,!?"))
    			if word == "" {
    				continue
    			}
    			counts[word]++
    		}
    	}
    	return counts
    }

    func countChars(messages []string) int {
    	chars := 0
    	for _, msg := range messages {
    		for _, r := range msg {
    			if !unicode.IsSpace(r) {
    				chars++
    			}
    		}
    	}
    	return chars
    }

    func topWords(counts map[string]int, n int) []WordCount {
    	var entries []WordCount
    	for word, count := range counts {
    		entries = append(entries, WordCount{word, count})
    	}
    	slices.SortFunc(entries, func(a, b WordCount) int {
    		if c := cmp.Compare(b.Count, a.Count); c != 0 {
    			return c
    		}
    		return cmp.Compare(a.Word, b.Word)
    	})
    	return entries[:min(n, len(entries))]
    }

    // ---- Your code ----

    // run parses args, reads the file and writes the report to out.
    func run(args []string, out io.Writer) error {
    	path, top, err := parseArgs(args)
    	if err != nil {
    		return err
    	}
    	messages, err := loadMessages(path)
    	if err != nil {
    		return err
    	}
    	counts := countWords(messages)
    	words := 0
    	for _, n := range counts {
    		words += n
    	}

    	fmt.Fprintf(out, "=== Textio report: %s ===\n", filepath.Base(path))
    	fmt.Fprintf(out, "%d messages, %d words, %d characters\n", len(messages), words, countChars(messages))
    	for _, wc := range topWords(counts, top) {
    		fmt.Fprintf(out, "%s: %d\n", wc.Word, wc.Count)
    	}
    	fmt.Fprintln(out, "=== end ===")
    	return nil
    }

    func main() {
    	// Create a sample file so Run works anywhere.
    	path := filepath.Join(os.TempDir(), "messages.txt")
    	sample := "Lunch at noon?\nYes! Lunch at noon.\n\nSee you at lunch!\n"
    	if err := os.WriteFile(path, []byte(sample), 0o644); err != nil {
    		fmt.Println(err)
    		return
    	}
    	defer os.Remove(path)

    	// In your terminal you'd pass os.Args[1:] instead.
    	if err := run([]string{"-top", "3", path}, os.Stdout); err != nil {
    		fmt.Println("error:", err)
    	}
    }
  tests: |
    package main

    import (
    	"errors"
    	"io/fs"
    	"os"
    	"path/filepath"
    	"strings"
    	"testing"
    )

    func writeFile(t *testing.T, name, content string) string {
    	t.Helper()
    	path := filepath.Join(t.TempDir(), name)
    	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
    		t.Fatal(err)
    	}
    	return path
    }

    func TestRunReport(t *testing.T) {
    	for _, tc := range []struct {
    		name, content string
    		flags         []string
    		want          string
    	}{
    		{
    			"messages.txt",
    			"Lunch at noon?\nYes! Lunch at noon.\n\nSee you at lunch!\n",
    			[]string{"-top", "3"},
    			"=== Textio report: messages.txt ===\n" +
    				"3 messages, 11 words, 42 characters\n" +
    				"at: 3\nlunch: 3\nnoon: 2\n" +
    				"=== end ===\n",
    		},
    		{
    			"texts.txt",
    			"Go go GO!\n  café ☕ time  \n",
    			nil,
    			"=== Textio report: texts.txt ===\n" +
    				"2 messages, 6 words, 16 characters\n" +
    				"go: 3\ncafé: 1\ntime: 1\n" +
    				"=== end ===\n",
    		},
    		{
    			"one.txt",
    			"hi\n",
    			[]string{"-top", "10"},
    			"=== Textio report: one.txt ===\n" +
    				"1 messages, 1 words, 2 characters\n" +
    				"hi: 1\n" +
    				"=== end ===\n",
    		},
    		{
    			"empty.txt",
    			"\n\n",
    			nil,
    			"=== Textio report: empty.txt ===\n" +
    				"0 messages, 0 words, 0 characters\n" +
    				"=== end ===\n",
    		},
    	} {
    		path := writeFile(t, tc.name, tc.content)
    		var out strings.Builder
    		args := append(tc.flags, path)
    		if err := run(args, &out); err != nil {
    			t.Errorf("run(%q) on file %q returned error: %v", args, tc.content, err)
    			continue
    		}
    		if out.String() != tc.want {
    			t.Errorf("run on file %q with flags %q wrote:\n%s\nwant:\n%s", tc.content, tc.flags, out.String(), tc.want)
    		}
    	}
    }

    func TestRunErrors(t *testing.T) {
    	var out strings.Builder
    	if err := run(nil, &out); !errors.Is(err, ErrUsage) {
    		t.Errorf("run with no arguments: error = %v, want ErrUsage", err)
    	}
    	missing := filepath.Join(t.TempDir(), "missing.txt")
    	if err := run([]string{missing}, &out); !errors.Is(err, fs.ErrNotExist) {
    		t.Errorf("run on a missing file: error = %v, want one wrapping fs.ErrNotExist", err)
    	}
    	if out.Len() != 0 {
    		t.Errorf("run wrote %q before failing, want nothing written when there's an error", out.String())
    	}
    }
---

Time to put it all together. Over the last four lessons you built every piece of the
Textio message report:

| Lesson | Function | Job |
| --- | --- | --- |
| Reading Message Files | `loadMessages` | read a file into non-blank lines |
| Counting Words and Characters | `countWords`, `countChars` | tally what's in them |
| Sorting the Results | `topWords` | most common words first |
| Command-Line Arguments | `parseArgs` | get the file name and `-top` |

They're already in the starter code, exactly as you wrote them. Your job is the
function that connects them.

## A testable main

It's tempting to write the whole program inside `main`. But `main` can't return
an error and can't be called from a test. So experienced Go programmers keep `main`
tiny and move the real work into a function like this:

```go
func run(args []string, out io.Writer) error
```

- `args` is the command line **without** the program name, so a test can pass any
  arguments it likes.
- `out` is where the report goes. `fmt.Fprintf(out, ...)` works just like
  `fmt.Printf`, but writes to `out`.
- Returning an `error` means one place decides how to report failures.

Then `main` in your terminal version is just:

```go
func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

Errors go to **standard error** (`os.Stderr`) so they don't get mixed into the
report, and the non-zero exit status tells the shell that something went wrong.

## Your assignment

Complete `run`:

1. Call `parseArgs(args)` to get the path and the `top` value.
2. Call `loadMessages(path)`.
3. If either returns an error, return it straight away, **before writing anything**
   to `out`.
4. Count the words with `countWords`. The total number of words is the sum of all
   the counts in the map.
5. Write the report to `out` in exactly this format:

```text
=== Textio report: messages.txt ===
3 messages, 11 words, 42 characters
at: 3
lunch: 3
noon: 2
=== end ===
```

The header uses `filepath.Base(path)`, which gives just the file name without its
directories. The second line gives the number of messages (lines), the total word
count and `countChars`. Then comes one `word: count` line for each entry from
`topWords(counts, top)`. Keep the wording fixed, even for a single message
(`1 messages`); pluralizing properly is a nice extra for your terminal version.

**Run** creates a sample file in your temp directory and should print the report
above. **Submit** also tries other files, the default `-top` of 3, an empty file,
missing arguments and a missing file.

## Take it to your terminal

This is a real tool, so try it for real. On your machine, make a folder, run
`go mod init report`, paste in your solution, and replace the body of `main` with the
`run(os.Args[1:], os.Stdout)` version above. Then point it at any text file:

```sh
$ go run . -top 5 notes.txt
```

Try it with no arguments and with a file that doesn't exist, and check that you get
a helpful error each time.

Congratulations: you've just built your first complete Go program, from reading input
to printing sorted results! One more stop, the final review, and you're done with
the course.
