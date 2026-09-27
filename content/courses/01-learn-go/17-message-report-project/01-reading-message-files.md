---
title: Reading Message Files
quiz:
  - question: What does `os.ReadFile` return?
    options:
      - text: A `string` holding the whole file
      - text: 'A `[]byte` holding the whole file, and an `error`'
        correct: true
      - text: One line of the file at a time
      - text: 'An `*os.File` you must close'
    explanation: |
      `os.ReadFile` opens the file, reads *all* of it into a `[]byte`, and
      closes it for you. Convert with `string(data)` when you want text. If
      anything goes wrong (for example, the file doesn't exist), the error is
      non-nil and you shouldn't trust `data`.
  - question: |
      Why does this lesson write its sample file into `os.TempDir()` first,
      instead of just reading `messages.txt`?
    options:
      - text: Because Go can only read files inside the temp directory
      - text: Because `os.ReadFile` deletes the file after reading it
      - text: Because a program should never assume a file exists; creating it first makes the example work anywhere
        correct: true
    explanation: |
      Files outside your program's control might be missing, renamed or
      unreadable. The example creates its own file in the system's temporary
      directory so it runs the same on every machine. Real programs read
      whatever path the user gives them and **check the error**.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"os"
    	"path/filepath"
    )

    // loadMessages reads the file at path and returns its non-blank lines,
    // with surrounding spaces trimmed. If the file can't be read, it returns
    // nil and an error that starts with "load messages: " and wraps the original.
    func loadMessages(path string) ([]string, error) {
    	// ?
    	return nil, nil
    }

    func main() {
    	path := filepath.Join(os.TempDir(), "textio-load.txt")
    	os.WriteFile(path, []byte("hi there\n\n  see you soon  \n"), 0o644)
    	defer os.Remove(path)

    	messages, err := loadMessages(path)
    	fmt.Printf("%q %v\n", messages, err)

    	_, err = loadMessages(filepath.Join(os.TempDir(), "no-such-file.txt"))
    	fmt.Println(err != nil)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"os"
    	"path/filepath"
    	"strings"
    )

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

    func main() {
    	path := filepath.Join(os.TempDir(), "textio-load.txt")
    	os.WriteFile(path, []byte("hi there\n\n  see you soon  \n"), 0o644)
    	defer os.Remove(path)

    	messages, err := loadMessages(path)
    	fmt.Printf("%q %v\n", messages, err)

    	_, err = loadMessages(filepath.Join(os.TempDir(), "no-such-file.txt"))
    	fmt.Println(err != nil)
    }
  tests: |
    package main

    import (
    	"errors"
    	"io/fs"
    	"os"
    	"path/filepath"
    	"slices"
    	"strings"
    	"testing"
    )

    func TestLoadMessages(t *testing.T) {
    	for _, tc := range []struct {
    		content string
    		want    []string
    	}{
    		{"hi there\n\n  see you soon  \n", []string{"hi there", "see you soon"}},
    		{"one\ntwo\nthree", []string{"one", "two", "three"}},
    		{"\r\n  \t\nlunch?\r\n", []string{"lunch?"}},
    		{"", nil},
    	} {
    		path := filepath.Join(t.TempDir(), "messages.txt")
    		if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
    			t.Fatal(err)
    		}
    		got, err := loadMessages(path)
    		if err != nil {
    			t.Errorf("loadMessages on file %q returned error %v, want nil", tc.content, err)
    			continue
    		}
    		if !slices.Equal(got, tc.want) {
    			t.Errorf("loadMessages on file %q = %q, want %q", tc.content, got, tc.want)
    		}
    	}
    }

    func TestLoadMessagesMissingFile(t *testing.T) {
    	path := filepath.Join(t.TempDir(), "missing.txt")
    	got, err := loadMessages(path)
    	if err == nil {
    		t.Fatalf("loadMessages(missing file) error = nil, want an error")
    	}
    	if got != nil {
    		t.Errorf("loadMessages(missing file) messages = %q, want nil", got)
    	}
    	if !strings.HasPrefix(err.Error(), "load messages: ") {
    		t.Errorf("error %q should start with \"load messages: \"", err)
    	}
    	if !errors.Is(err, fs.ErrNotExist) {
    		t.Errorf("error %q should wrap the original error with %%w (errors.Is(err, fs.ErrNotExist) was false)", err)
    	}
    }
---

Welcome to the project chapter! Over the next five lessons you'll build a small
command-line tool for Textio: it reads a file full of text messages and prints a
report of the most common words. It's the same idea as boot.dev's "Bookbot", but
for SMS.

Every useful program starts with input. So far all of our data has been written
directly into the code. Now it's time to read it from a **file**.

## os.ReadFile

The simplest way to read a file is `os.ReadFile`. It reads the *whole* file into
memory and hands you a byte slice:

```go
data, err := os.ReadFile("messages.txt")
if err != nil {
	fmt.Println("could not read messages:", err)
	return
}
fmt.Println(string(data))
```

Two things to notice:

1. It returns an `error`. Files go missing all the time, so **always check it**.
2. It returns `[]byte`, not `string`. You met bytes in the strings chapter;
   `string(data)` converts them to text.

## Never assume a file exists

Our examples can't rely on a `messages.txt` sitting on your computer. Instead,
they create one first. `os.TempDir()` returns the operating system's scratch
directory, `filepath.Join` glues path pieces together with the right separator,
and `os.WriteFile` writes bytes to a file (the `0o644` is the usual "owner can
write, everyone can read" permission):

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	path := filepath.Join(os.TempDir(), "textio-messages.txt")
	sample := "Hey, are you free?\n\nYes! Lunch at noon?\n"
	if err := os.WriteFile(path, []byte(sample), 0o644); err != nil {
		fmt.Println("write failed:", err)
		return
	}
	defer os.Remove(path) // tidy up when main returns

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("read failed:", err)
		return
	}
	for i, line := range strings.Split(string(data), "\n") {
		fmt.Printf("%d: %q\n", i, line)
	}
}
```

```text
0: "Hey, are you free?"
1: ""
2: "Yes! Lunch at noon?"
3: ""
```

Look closely at the output. Splitting on `"\n"` gives an empty string for the
blank line *and* one after the final newline. Real message files are messy like
this, so our tool will skip blank lines.

## bufio.Scanner: one line at a time

`os.ReadFile` loads everything at once. That's fine for a few thousand messages,
but for huge files you can read line by line with `os.Open` and a
`bufio.Scanner`:

```go
file, err := os.Open(path)
if err != nil {
	return err
}
defer file.Close()

scanner := bufio.NewScanner(file)
for scanner.Scan() {
	fmt.Println(scanner.Text()) // one line, without the "\n"
}
if err := scanner.Err(); err != nil {
	return err
}
```

`Scan` returns `false` when the input runs out *or* something breaks, so check
`scanner.Err()` after the loop. Unlike `os.ReadFile`, `os.Open` hands you an
open file, and you must `Close` it; `defer` makes that impossible to forget.

## Missing files

When a file doesn't exist, the error wraps `fs.ErrNotExist` (from the `io/fs`
package), so you can test for it with `errors.Is`, just like in the errors chapter:

```go
_, err := os.ReadFile("nope.txt")
fmt.Println(errors.Is(err, fs.ErrNotExist)) // true
```

## Your turn

Complete `loadMessages`. It should:

1. Read the whole file with `os.ReadFile`.
2. If that fails, return `nil` and an error made with
   `fmt.Errorf("load messages: %w", err)`, so callers can still use `errors.Is`.
3. Otherwise split the text into lines, trim each one with `strings.TrimSpace`,
   skip lines that end up empty, and return the rest in order.

An empty file should give back a `nil` slice and a `nil` error. **Run** should print
`["hi there" "see you soon"] <nil>` and then `true`.
