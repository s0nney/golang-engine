---
title: Reading Errors and Debugging
quiz:
  - question: |
      What does this compiler error mean?

      ```text
      ./send.go:14:9: cannot use count (variable of type int) as string value in return statement
      ```
    options:
      - text: Line 14 of `send.go` returns an `int` from a function whose return type is `string`
        correct: true
      - text: The program crashed on line 14 while it was running
      - text: The variable `count` was never declared
      - text: '`send.go` has 14 errors'
    explanation: |
      The format is `file:line:column: message`. The function is declared to
      return a `string`, but line 14 returns `count`, an `int`. Convert it,
      for example with `strconv.Itoa(count)`, or fix the return type.
  - question: |
      You add a debug print and see `"Alice "` in the output. What does
      the `%q` verb reveal that `%s` or `%v` wouldn't have?
    options:
      - text: That the string is in double quotes in your code
      - text: That there's a trailing space in the value
        correct: true
      - text: That the string is a constant
    explanation: |
      `%q` prints the value in quotes, with special characters escaped, so
      invisible problems like trailing spaces, tabs (`\t`) and newlines
      (`\n`) become visible. With `%s` you'd just see `Alice`.
exercise:
  starter: |
    package main

    import "fmt"

    // lastMessage returns the most recent message, or "" if there are none.
    func lastMessage(msgs []string) string {
    	return msgs[len(msgs)]
    }

    // countFailed returns how many statuses are "failed".
    func countFailed(statuses []string) int {
    	failed := 0
    	for _, s := range statuses {
    		failed := 0
    		if s == "failed" {
    			failed++
    		}
    	}
    	return failed
    }

    func main() {
    	fmt.Println(countFailed([]string{"sent", "failed", "failed"}))
    	fmt.Println(lastMessage([]string{"hi", "running late", "here!"}))
    	fmt.Println(lastMessage(nil))
    }
  solution: |
    package main

    import "fmt"

    func lastMessage(msgs []string) string {
    	if len(msgs) == 0 {
    		return ""
    	}
    	return msgs[len(msgs)-1]
    }

    func countFailed(statuses []string) int {
    	failed := 0
    	for _, s := range statuses {
    		if s == "failed" {
    			failed++
    		}
    	}
    	return failed
    }

    func main() {
    	fmt.Println(countFailed([]string{"sent", "failed", "failed"}))
    	fmt.Println(lastMessage([]string{"hi", "running late", "here!"}))
    	fmt.Println(lastMessage(nil))
    }
  tests: |
    package main

    import "testing"

    func TestLastMessage(t *testing.T) {
    	for _, tc := range []struct {
    		msgs []string
    		want string
    	}{
    		{[]string{"hi", "running late", "here!"}, "here!"},
    		{[]string{"only one"}, "only one"},
    		{nil, ""},
    		{[]string{}, ""},
    	} {
    		func() {
    			defer func() {
    				if r := recover(); r != nil {
    					t.Errorf("lastMessage(%q) panicked: %v", tc.msgs, r)
    				}
    			}()
    			if got := lastMessage(tc.msgs); got != tc.want {
    				t.Errorf("lastMessage(%q) = %q, want %q", tc.msgs, got, tc.want)
    			}
    		}()
    	}
    }

    func TestCountFailed(t *testing.T) {
    	for _, tc := range []struct {
    		statuses []string
    		want     int
    	}{
    		{[]string{"sent", "failed", "failed"}, 2},
    		{[]string{"failed"}, 1},
    		{[]string{"sent", "sent"}, 0},
    		{nil, 0},
    	} {
    		if got := countFailed(tc.statuses); got != tc.want {
    			t.Errorf("countFailed(%q) = %d, want %d", tc.statuses, got, tc.want)
    		}
    	}
    }
---

Every programmer spends a lot of time with code that doesn't work yet. The difference between a beginner and an experienced developer isn't that the expert makes fewer mistakes. It's that they find them faster. Here's how.

## Read the whole error message

Compiler errors look intimidating, but they follow one format:

```text
./main.go:12:15: undefined: sendMesage
```

- `./main.go` is the file.
- `12` is the line and `15` is the column.
- After that comes the actual problem.

Always start with the **first** error. Later errors are often knock-on effects of the first one, and fixing it can make the rest disappear.

Some errors you'll see all the time:

| Error | What it usually means |
|-------|------------------------|
| `undefined: x` | A typo, a missing import, or `x` is out of scope |
| `declared and not used: x` | Use the variable, or delete it |
| `"strings" imported and not used` | Remove the import (your editor can do this on save) |
| `cannot use x (variable of type int) as string value` | A type mismatch: convert it or fix the types |
| `missing return` | Some path through the function doesn't return |
| `assignment mismatch: 1 variable but f returns 2 values` | Receive every return value, or use `_` |
| `x.foo undefined (type T has no field or method foo)` | A typo, or the name isn't exported (check the capital letter) |

## Read the panic

Runtime crashes print a **panic message** and a **stack trace**. Look for the message first, then the first line of the trace that points into your own code:

```text
panic: runtime error: index out of range [3] with length 3

goroutine 1 [running]:
main.lastMessage(...)
	/home/you/smsapp/main.go:8
main.main()
	/home/you/smsapp/main.go:13 +0x1d
exit status 2
```

Read it bottom-up to see the chain of calls: `main` (line 13) called `lastMessage`, which crashed on line 8 by reading index 3 of a 3-element slice. Off-by-one error! The last valid index is `len(s) - 1`.

## Print debugging

The simplest debugging tool is also one of the most effective: **print what's actually happening**. When the result is wrong, print the values along the way and find the first place where reality differs from what you expected:

```go
package main

import "fmt"

func totalCost(lengths []int) float64 {
	total := 0.0
	for i, n := range lengths {
		segments := n / 160
		fmt.Printf("DEBUG i=%d n=%d segments=%d\n", i, n, segments)
		total += float64(segments) * 0.01
	}
	return total
}

func main() {
	fmt.Println(totalCost([]int{42, 200}))
}
```

```text
DEBUG i=0 n=42 segments=0
DEBUG i=1 n=200 segments=1
0.01
```

There it is: a 42-character message counts as 0 segments. It's the same rounding bug from the testing lesson.

Tips:

- Label every debug line (`DEBUG i=...`) so you can tell them apart and find them later to delete.
- Use `%q` for strings, so stray spaces and newlines are visible: `"Alice "` versus `"Alice"`.
- Use `%+v` for structs, to see field names.
- Use `%T` when you're not sure what type something is.
- Remove debug prints before committing. Better yet, turn the bug into a **test** so it can never come back.

## Narrow it down

When you have no idea where a bug is, shrink the search space:

1. **Reproduce it reliably.** Find an input that always triggers it. Ideally, write a failing test.
2. **Divide and conquer.** Check the value halfway through the process. Right there? The bug is later. Wrong already? It's earlier. Repeat.
3. **Question your assumptions.** The bug is usually in the line you were sure was fine.
4. **Explain it out loud.** Describing the code line by line to someone (or a rubber duck) is surprisingly effective at making the bug jump out.

## Debuggers

For tricky bugs, a **debugger** lets you pause a running program, step through it line by line and inspect every variable. The standard Go debugger is **Delve** (`dlv`), and most editors, such as VS Code and GoLand, integrate it so you can click to set breakpoints. Print debugging and tests will take you a long way, but it's worth learning a debugger once you're comfortable with the basics.

## Your turn

This program has two bugs. Debug it the way this lesson describes:

1. Press **Run**. `countFailed` prints `0` even though two messages failed. Add a
   `fmt.Printf("DEBUG ...")` line inside the loop if you need to, and look closely
   at every `:=`. (Hint: the shadowing chapter.)
2. Once that's fixed, **Run** panics. Read the panic message and the stack trace
   to find the line in `lastMessage`, then fix it. It must also return `""` for an
   empty or nil slice instead of crashing.

Remove any debug prints before you submit. When everything works, **Run** prints
`2`, `here!` and an empty line.

## Further reading

- [Delve debugger](https://github.com/go-delve/delve)
