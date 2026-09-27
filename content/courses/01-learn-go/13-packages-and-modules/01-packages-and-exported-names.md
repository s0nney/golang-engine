---
title: Packages and Exported Names
quiz:
  - question: |
      A package `billing` declares these functions:

      ```go
      func Cost(body string) float64 { ... }
      func segments(body string) int { ... }
      ```

      Which call works from `package main`?
    options:
      - text: '`billing.segments("hi")`'
      - text: '`billing.Cost("hi")`'
        correct: true
      - text: Both
      - text: Neither, because `main` can't import other packages
    explanation: |
      Only names that start with a capital letter are exported (visible
      outside their package). `Cost` is exported; `segments` is private to
      `billing`, and calling it from `main` fails with
      `name segments not exported by package billing`.
  - question: Two files in the same folder, `send.go` and `bill.go`, both start with `package main`. Can a function in `send.go` call an unexported function declared in `bill.go`?
    options:
      - text: Yes, because they're in the same package
        correct: true
      - text: No, each file is its own scope
      - text: Only if `send.go` imports `bill.go`
    explanation: |
      A package is all the `.go` files in one directory. Package-level names,
      exported or not, are shared across every file in the package. Files
      never import each other.
exercise:
  starter: |
    package main

    import "fmt"

    // Imagine this file is the billing package. Fix the capital letters so
    // only the API that other packages need is exported.

    // maxSegments is the most segments Textio will send for one message.
    const maxSegments = 3

    // cost returns the price in cents of sending body, or -1 if it's too long.
    func cost(body string) int {
    	n := Segments(body)
    	if n > maxSegments {
    		return -1
    	}
    	return n * CentsPerSegment
    }

    // Segments reports how many 160-character SMS segments body needs.
    func Segments(body string) int {
    	return (len(body) + 159) / 160
    }

    // CentsPerSegment is an internal price detail.
    const CentsPerSegment = 2

    func main() {
    	fmt.Println(cost("Your code is 4821"), maxSegments)
    }
  solution: |
    package main

    import "fmt"

    // MaxSegments is the most segments Textio will send for one message.
    const MaxSegments = 3

    // Cost returns the price in cents of sending body, or -1 if it's too long.
    func Cost(body string) int {
    	n := segments(body)
    	if n > MaxSegments {
    		return -1
    	}
    	return n * centsPerSegment
    }

    // segments reports how many 160-character SMS segments body needs.
    func segments(body string) int {
    	return (len(body) + 159) / 160
    }

    // centsPerSegment is an internal price detail.
    const centsPerSegment = 2

    func main() {
    	fmt.Println(Cost("Your code is 4821"), MaxSegments)
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    )

    // This test only compiles if the exported and unexported names are right.
    func TestAPI(t *testing.T) {
    	if MaxSegments != 3 {
    		t.Errorf("MaxSegments = %d, want 3", MaxSegments)
    	}
    	if centsPerSegment != 2 {
    		t.Errorf("centsPerSegment = %d, want 2", centsPerSegment)
    	}
    	for _, tc := range []struct {
    		body string
    		want int
    	}{
    		{"Your code is 4821", 2},
    		{strings.Repeat("a", 160), 2},
    		{strings.Repeat("a", 161), 4},
    		{strings.Repeat("a", 480), 6},
    		{strings.Repeat("a", 481), -1},
    	} {
    		if got := Cost(tc.body); got != tc.want {
    			t.Errorf("Cost(<%d characters>) = %d, want %d", len(tc.body), got, tc.want)
    		}
    		if got := segments(tc.body); got != (len(tc.body)+159)/160 {
    			t.Errorf("segments(<%d characters>) = %d", len(tc.body), got)
    		}
    	}
    }
---

Real programs are too big for one file. Textio's backend has code for billing, carriers, phone number validation and much more. Go organises code into **packages**.

## A package is a folder

A **package** is a directory of `.go` files that all start with the same `package` line. Everything declared at the top level of any of those files (functions, types, variables, constants) belongs to the whole package, so the files can use each other's names freely. Files never import each other.

Here's a small Textio project laid out as two packages:

```text
smsapp/
├── go.mod
├── main.go            (package main)
└── billing/
    └── billing.go     (package billing)
```

`billing/billing.go`:

```go
// Package billing works out what Textio charges for messages.
package billing

const costPerSegment = 0.01

// Cost returns the price in dollars of sending body as an SMS.
func Cost(body string) float64 {
	return float64(segments(body)) * costPerSegment
}

func segments(body string) int {
	return (len(body) + 159) / 160
}
```

`main.go`:

```go
package main

import (
	"fmt"

	"github.com/textio/smsapp/billing"
)

func main() {
	fmt.Printf("$%.2f\n", billing.Cost("Your code is 4821"))
}
```

```text
$ go run .
$0.01
```

(`go run .` means "run the package in the current directory". The `github.com/textio/smsapp` part comes from the `go.mod` file, which is next lesson's topic.)

## Exported names start with a capital letter

Go has no `public` or `private` keywords. Instead, it uses the **first letter** of a name:

- **Capitalised** names like `Cost` are **exported**: code in other packages can use them.
- **Lowercase** names like `segments` and `costPerSegment` are **unexported**: only code inside the same package can see them.

Try calling `billing.segments` from `main` and the compiler stops you:

```text
./main.go:10:32: name segments not exported by package billing
```

This applies to everything: functions, types, constants, variables, struct fields and methods. That's why every function you've used from the standard library starts with a capital letter: `fmt.Println`, `strconv.Itoa`, `slices.Sort`.

Keep as much as possible unexported. The exported names are your package's **API**, a promise to everyone who uses it. Unexported names are implementation details that you're free to change.

## Importing packages

- Standard library packages are imported by their short path: `"fmt"`, `"strings"`, `"errors"`.
- Your own and third-party packages are imported by their full path: `"github.com/textio/smsapp/billing"`.
- Group imports in one `import ( ... )` block. By convention, the standard library comes first, then a blank line, then everything else.
- You refer to a package by the **last element** of its path: `billing.Cost`.
- Importing a package you don't use is a compile error, just like an unused variable.

If two packages have the same name, you can give one an alias:

```go
import (
	mrand "math/rand/v2"
)
```

## `package main` is special

A package named `main` with a `func main()` builds into a **program**. Every other package is a **library**: code for other packages to import. A library package's name should be short, lowercase and a single word, and usually matches its folder name: `billing`, not `billingUtils` or `billing_helpers`.

## Doc comments

Notice the comments right above `package billing` and `func Cost`. A comment directly before a declaration is its **doc comment**. Tools like `go doc` and pkg.go.dev display them as documentation. Start each one with the name it describes: "Cost returns...". Every exported name should have one.

## Your turn

The exercise editor only holds one file, so imagine the code above `main` is the
`billing` package. Whoever wrote it got the capital letters backwards: the helpers
are exported and the real API isn't.

Rename things so that:

- `Cost` and `MaxSegments` are **exported**: other packages need them.
- `segments` and `centsPerSegment` are **unexported**: they're implementation details.

Update every use of each name, and the doc comments too (a doc comment starts with
the name it documents). Your editor would do this with a "rename symbol" command;
here, do it by hand. **Run** should still print `2 3`.

## Further reading

- [A Tour of Go: Exported names](https://go.dev/tour/basics/3)
- [How to Write Go Code](https://go.dev/doc/code)
