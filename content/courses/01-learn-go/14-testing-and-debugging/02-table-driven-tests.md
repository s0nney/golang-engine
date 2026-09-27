---
title: Table-Driven Tests
quiz:
  - question: |
      A table-driven test has cases named `"empty"`, `"short"` and
      `"exactly one segment"`, each run with `t.Run(tt.name, ...)`. Which
      command runs only the `"exactly one segment"` case?
    options:
      - text: '`go test -run "exactly one segment"`'
      - text: '`go test -run TestSegments/exactly_one_segment`'
        correct: true
      - text: '`go test -case 3`'
    explanation: |
      Subtests are named `TestName/subtest`, with spaces replaced by
      underscores. `-run` takes a pattern that matches those names, with `/`
      separating the levels.
  - question: What's the main advantage of a table-driven test over writing one test function per case?
    options:
      - text: It runs faster
      - text: Adding a new case is a single line of data, and every case shares the same checking logic
        correct: true
      - text: Go requires tests to be table-driven
      - text: It hides failures from cases you don't care about
    explanation: |
      The logic is written once; the cases are just data. That makes it
      cheap to add edge cases, and easy to see at a glance which inputs are
      covered.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Segments returns how many 160-character SMS segments a message body
    // needs. An empty body needs 0 segments.
    func Segments(body string) int {
    	return len(body)/160 + 1
    }

    func main() {
    	for _, n := range []int{0, 5, 160, 161, 320} {
    		fmt.Printf("%d chars -> %d segment(s)\n", n, Segments(strings.Repeat("a", n)))
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func Segments(body string) int {
    	return (len(body) + 159) / 160
    }

    func main() {
    	for _, n := range []int{0, 5, 160, 161, 320} {
    		fmt.Printf("%d chars -> %d segment(s)\n", n, Segments(strings.Repeat("a", n)))
    	}
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    )

    func TestSegments(t *testing.T) {
    	tests := []struct {
    		name string
    		body string
    		want int
    	}{
    		{name: "empty", body: "", want: 0},
    		{name: "short", body: "hello", want: 1},
    		{name: "exactly one segment", body: strings.Repeat("a", 160), want: 1},
    		{name: "just over", body: strings.Repeat("a", 161), want: 2},
    		{name: "two full segments", body: strings.Repeat("a", 320), want: 2},
    		{name: "long", body: strings.Repeat("a", 1000), want: 7},
    	}
    	for _, tt := range tests {
    		t.Run(tt.name, func(t *testing.T) {
    			if got := Segments(tt.body); got != tt.want {
    				t.Errorf("Segments(%d chars) = %d, want %d", len(tt.body), got, tt.want)
    			}
    		})
    	}
    }
---

One test case is good. But `Segments` has several interesting inputs: an empty message, a short one, exactly 160 characters, 161 characters. Writing a separate test function for each would mean copying the same checking code four times. Go programmers have a favourite pattern for this: the **table-driven test**.

## A table of cases

Describe each case as data in a slice of anonymous structs, then loop over it:

```go
package billing

import (
	"strings"
	"testing"
)

func TestSegments(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{name: "empty", body: "", want: 0},
		{name: "short", body: "hello", want: 1},
		{name: "exactly one segment", body: strings.Repeat("a", 160), want: 1},
		{name: "just over", body: strings.Repeat("a", 161), want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Segments(tt.body)
			if got != tt.want {
				t.Errorf("Segments(%d chars) = %d, want %d", len(tt.body), got, tt.want)
			}
		})
	}
}
```

`strings.Repeat("a", 160)` builds a string of 160 `a`s. `tt` is the conventional name for "this test case".

Remember the anonymous structs from the structs chapter? This is where they shine. The struct type only matters inside this one test, so it doesn't need a name.

Adding a case is now one line. Want to test 320 characters? Add `{name: "two full segments", body: strings.Repeat("a", 320), want: 2},` and you're done.

## Subtests with `t.Run`

`t.Run(name, func)` runs each case as a named **subtest**. Suppose someone "simplifies" `Segments` to `len(body)/160 + 1`. Here's what you'd see:

```text
$ go test ./billing
--- FAIL: TestSegments (0.00s)
    --- FAIL: TestSegments/empty (0.00s)
        billing_test.go:24: Segments(0 chars) = 1, want 0
    --- FAIL: TestSegments/exactly_one_segment (0.00s)
        billing_test.go:24: Segments(160 chars) = 2, want 1
FAIL
FAIL	github.com/textio/smsapp/billing	0.001s
FAIL
```

Subtests give you three big wins:

1. **Clear failures.** Each failure names its case, so you know exactly which input broke.
2. **Every case runs.** One failing case doesn't hide the others. (Even `t.Fatalf` inside a subtest only stops *that* subtest.)
3. **Run one case.** Subtest names are `TestName/case_name`, with spaces turned into underscores, so you can focus on one:

```text
$ go test -run 'TestSegments/just_over' -v ./billing
=== RUN   TestSegments
=== RUN   TestSegments/just_over
--- PASS: TestSegments (0.00s)
    --- PASS: TestSegments/just_over (0.00s)
PASS
ok  	github.com/textio/smsapp/billing	0.001s
```

The inner function takes its own `t *testing.T`, which belongs to the subtest. It shadows the outer `t` on purpose, so failures are reported against the right case.

## Testing errors in a table

Tables work just as well for functions that return errors. Add a field saying whether an error is expected:

```go
tests := []struct {
	name    string
	phone   string
	wantErr bool
}{
	{name: "valid", phone: "+1-555-0100", wantErr: false},
	{name: "too short", phone: "555", wantErr: true},
	{name: "no plus", phone: "15550100000", wantErr: true},
}

for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		err := validatePhone(tt.phone)
		if (err != nil) != tt.wantErr {
			t.Errorf("validatePhone(%q) error = %v, wantErr %v", tt.phone, err, tt.wantErr)
		}
	})
}
```

`(err != nil) != tt.wantErr` reads as "whether we got an error doesn't match whether we wanted one".

## Choosing good cases

The best test cases sit on the **edges**: zero, one, exactly at a limit, one past it, negative numbers, empty strings, nil slices. Bugs love boundaries. The `Segments` bug above was invisible for a "normal" message and only showed up at 0 and exactly 160.

## Your turn

Here's the "simplified" `Segments` from this lesson, bug and all. The hidden tests are a table-driven test much like the one above, with a few extra cases.

Press **Run** to see what the function returns now, then fix `Segments` so every subtest passes. When you **Submit**, look at the `-v` output: each subtest is listed by name.

## Further reading

- [Learn Go with Tests: Structs, methods & interfaces (table driven tests)](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/structs-methods-and-interfaces)
- [Go Wiki: Table-driven tests](https://go.dev/wiki/TableDrivenTests)
