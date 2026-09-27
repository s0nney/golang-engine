---
title: Command-Line Arguments
quiz:
  - question: 'You run `go run . -top 5 texts.txt`. What is `os.Args[0]`?'
    options:
      - text: '`-top`'
      - text: '`texts.txt`'
      - text: The path of the program itself
        correct: true
      - text: '`5`'
    explanation: |
      `os.Args[0]` is always the program's own name or path. The arguments
      the user typed start at `os.Args[1]`, which is why programs usually work
      with `os.Args[1:]`.
  - question: |
      After this code parses `[]string{"-top", "5", "texts.txt"}`, what are
      `*top` and `fs.Args()`?

      ```go
      fs := flag.NewFlagSet("report", flag.ContinueOnError)
      top := fs.Int("top", 3, "how many words to show")
      err := fs.Parse(args)
      ```
    options:
      - text: '`3` and `["-top" "5" "texts.txt"]`'
      - text: '`5` and `["texts.txt"]`'
        correct: true
      - text: '`5` and `[]`'
    explanation: |
      `fs.Int` returns a *pointer* whose value `Parse` fills in, so `*top`
      is 5 instead of the default 3. Whatever is left after the flags
      (the positional arguments) comes back from `fs.Args()`.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"flag"
    	"fmt"
    	"io"
    )

    var ErrUsage = errors.New("usage: report [-top N] FILE")

    // parseArgs parses the command-line arguments (without the program name).
    // It returns the file to read and how many top words to show.
    func parseArgs(args []string) (string, int, error) {
    	fs := flag.NewFlagSet("report", flag.ContinueOnError)
    	fs.SetOutput(io.Discard)
    	top := fs.Int("top", 3, "how many words to show")
    	if err := fs.Parse(args); err != nil {
    		return "", 0, err
    	}
    	// ?
    	return fs.Arg(0), *top, nil
    }

    func main() {
    	fmt.Println(parseArgs([]string{"-top", "5", "texts.txt"}))
    	fmt.Println(parseArgs([]string{"texts.txt"}))
    	fmt.Println(parseArgs([]string{}))
    	fmt.Println(parseArgs([]string{"-top", "0", "texts.txt"}))
    }
  solution: |
    package main

    import (
    	"errors"
    	"flag"
    	"fmt"
    	"io"
    )

    var ErrUsage = errors.New("usage: report [-top N] FILE")

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

    func main() {
    	fmt.Println(parseArgs([]string{"-top", "5", "texts.txt"}))
    	fmt.Println(parseArgs([]string{"texts.txt"}))
    	fmt.Println(parseArgs([]string{}))
    	fmt.Println(parseArgs([]string{"-top", "0", "texts.txt"}))
    }
  tests: |
    package main

    import (
    	"errors"
    	"testing"
    )

    func TestParseArgsValid(t *testing.T) {
    	for _, tc := range []struct {
    		args    []string
    		path    string
    		wantTop int
    	}{
    		{[]string{"-top", "5", "texts.txt"}, "texts.txt", 5},
    		{[]string{"texts.txt"}, "texts.txt", 3},
    		{[]string{"-top=1", "/tmp/a b.txt"}, "/tmp/a b.txt", 1},
    	} {
    		path, top, err := parseArgs(tc.args)
    		if err != nil || path != tc.path || top != tc.wantTop {
    			t.Errorf("parseArgs(%q) = %q, %d, %v; want %q, %d, nil", tc.args, path, top, err, tc.path, tc.wantTop)
    		}
    	}
    }

    func TestParseArgsUsage(t *testing.T) {
    	for _, args := range [][]string{{}, {"-top", "2"}, {"a.txt", "b.txt"}, {"a.txt", "-top", "2"}} {
    		if _, _, err := parseArgs(args); !errors.Is(err, ErrUsage) {
    			t.Errorf("parseArgs(%q) error = %v, want ErrUsage", args, err)
    		}
    	}
    }

    func TestParseArgsBadTop(t *testing.T) {
    	for _, args := range [][]string{{"-top", "0", "a.txt"}, {"-top", "-4", "a.txt"}, {"-top", "lots", "a.txt"}} {
    		if _, _, err := parseArgs(args); err == nil {
    			t.Errorf("parseArgs(%q) error = nil, want an error", args)
    		}
    	}
    }
---

Our report tool shouldn't have a file name baked into it. Real command-line tools
take **arguments**: `go run . messages.txt`. Let's read them.

## os.Args

`os.Args` is a `[]string` holding the command line. The first element is the
program's own path; everything the user typed comes after it:

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:] // skip the program name
	if len(args) != 1 {
		fmt.Println("usage: report FILE")
		os.Exit(1)
	}
	fmt.Println("reading", args[0])
}
```

```sh
$ go run . messages.txt
reading messages.txt
$ go run .
usage: report FILE
exit status 1
```

`os.Exit(1)` ends the program immediately with a non-zero **exit status**, which is
how command-line tools tell the shell "that didn't work". Deferred functions don't
run when you call `os.Exit`, so only use it at the very end of `main`.

## Flags with the flag package

Options like `-top 5` are called **flags**. You could pick them out of `os.Args`
by hand, but the standard `flag` package does it for you, including default values
and a generated `-h` help message.

We'll use a `FlagSet`, a named group of flags that parses whatever slice you give
it. That makes the parsing a normal function you can test, instead of something
glued to the real `os.Args`:

```go
func parseArgs(args []string) (string, int, error) {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	top := fs.Int("top", 3, "how many words to show")
	if err := fs.Parse(args); err != nil {
		return "", 0, err
	}
	// fs.Args() holds what's left after the flags.
	return fs.Arg(0), *top, nil
}
```

A few things to know:

- `fs.Int` returns an `*int`. `Parse` fills it in, so read it with `*top` **after**
  parsing.
- `flag.ContinueOnError` makes `Parse` return an error for a bad flag like
  `-top lots`, instead of exiting the program.
- Flags must come **before** positional arguments. `Parse` stops at the first
  argument that isn't a flag, so `texts.txt -top 5` leaves `-top` and `5` in
  `fs.Args()`.
- `fs.SetOutput(io.Discard)` silences the usage text `Parse` prints on errors,
  which is handy when you return the error yourself.

In `main`, you'd call `parseArgs(os.Args[1:])`.

The `flag` package can do much more (string and bool flags, custom types, and so on).
You'll go deeper into writing testable command-line programs in
[Learn Testing in Go](/courses/learn-testing). For now, a single flag is plenty.

## Your turn

`parseArgs` already defines and parses the `-top` flag. Finish it so that it
validates what it parsed:

1. If there isn't **exactly one** positional argument (`fs.NArg()` tells you how
   many there are), return `ErrUsage`.
2. If `*top` is less than 1, return an error such as
   `fmt.Errorf("-top must be at least 1, got %d", *top)`.
3. Otherwise return the file name, the top value and `nil`.

When you return an error, return `""` and `0` for the other values. **Run** prints
one line per call, and only the first two should end in `<nil>`.
