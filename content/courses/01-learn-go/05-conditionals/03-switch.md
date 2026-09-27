---
title: Switch
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	carrier := "verizon"
      	switch carrier {
      	case "att":
      		fmt.Println("AT&T")
      	case "verizon", "tmobile":
      		fmt.Println("supported")
      	case "sprint":
      		fmt.Println("legacy")
      	default:
      		fmt.Println("unknown")
      	}
      }
      ```
    options:
      - text: '`supported`'
        correct: true
      - text: '`supported`, then `legacy`, then `unknown`'
      - text: '`unknown`'
      - text: '`supported` and then `legacy`'
    explanation: |
      `"verizon"` matches the second case, which lists two values. Go runs
      only that case and then leaves the switch. Unlike C or JavaScript,
      there's no automatic fall-through, so no `break` is needed.
  - question: When does the `default` case of a `switch` run?
    options:
      - text: Always, after the matching case
      - text: Only when no other case matches
        correct: true
      - text: Only if it's the first case in the switch
    explanation: |
      `default` is the fallback. It runs only if none of the other cases
      match, no matter where it appears in the switch (though by convention
      it goes last).
exercise:
  starter: |
    package main

    import "fmt"

    // carrierFor returns the carrier Textio uses for a country calling code.
    func carrierFor(code string) string {
    	// ?
    	return ""
    }

    func main() {
    	fmt.Println(carrierFor("+1"))  // want Alpha Mobile
    	fmt.Println(carrierFor("+49")) // want Gamma Net
    	fmt.Println(carrierFor("+81")) // want unsupported
    }
  solution: |
    package main

    import "fmt"

    func carrierFor(code string) string {
    	switch code {
    	case "+1":
    		return "Alpha Mobile"
    	case "+44":
    		return "Beta Tel"
    	case "+33", "+49":
    		return "Gamma Net"
    	default:
    		return "unsupported"
    	}
    }

    func main() {
    	fmt.Println(carrierFor("+1"))
    	fmt.Println(carrierFor("+49"))
    	fmt.Println(carrierFor("+81"))
    }
  tests: |
    package main

    import "testing"

    func TestCarrierFor(t *testing.T) {
    	tests := []struct {
    		code string
    		want string
    	}{
    		{"+1", "Alpha Mobile"},
    		{"+44", "Beta Tel"},
    		{"+33", "Gamma Net"},
    		{"+49", "Gamma Net"},
    		{"+81", "unsupported"},
    		{"", "unsupported"},
    	}
    	for _, tt := range tests {
    		if got := carrierFor(tt.code); got != tt.want {
    			t.Errorf("carrierFor(%q) = %q, want %q", tt.code, got, tt.want)
    		}
    	}
    }
---

A long chain of `if`/`else if` comparing one value against many options gets repetitive. Go's `switch` statement is a cleaner way to write it.

## Switching on a value

```go
package main

import "fmt"

func countryName(code string) string {
	switch code {
	case "US":
		return "United States"
	case "GB":
		return "United Kingdom"
	case "FR":
		return "France"
	default:
		return "Unknown"
	}
}

func main() {
	fmt.Println(countryName("GB"))
	fmt.Println(countryName("XX"))
}
```

```text
United Kingdom
Unknown
```

Go compares `code` against each `case` from top to bottom and runs the **first** one that matches. `default` runs if none of them match. It's optional.

## No fall-through

In C, Java and JavaScript, a `switch` case "falls through" into the next case unless you remember to write `break`. Forgetting `break` is a classic bug.

Go flips this around: **each case stops automatically**. Only one case ever runs. (If you really want the C behaviour, Go has a `fallthrough` keyword, but you'll almost never need it.)

## Several values in one case

A case can list multiple values, separated by commas:

```go
package main

import "fmt"

func isWeekend(day string) bool {
	switch day {
	case "Saturday", "Sunday":
		return true
	default:
		return false
	}
}

func main() {
	fmt.Println(isWeekend("Sunday"), isWeekend("Monday"))
}
```

```text
true false
```

Textio uses rules like this to avoid sending marketing messages at the weekend.

## Switch with no value

If you leave out the value after `switch`, each case is a full boolean condition. The first case that's `true` wins. This is a tidy replacement for a long `if`/`else if` chain:

```go
package main

import "fmt"

func segments(length int) int {
	switch {
	case length == 0:
		return 0
	case length <= 160:
		return 1
	case length <= 306:
		return 2
	default:
		return 3
	}
}

func main() {
	fmt.Println(segments(0), segments(42), segments(200), segments(1000))
}
```

```text
0 1 2 3
```

## Switch with an init statement

Just like `if`, a `switch` can start with a short statement:

```go
switch length := len(message); {
case length > 160:
	fmt.Println("long")
default:
	fmt.Println("short")
}
```

The `;` after the init statement is required even when there's no value to switch on.

## Switching on constants

`switch` pairs beautifully with the `iota` constants you met earlier:

```go
const (
	Queued = iota
	Sending
	Delivered
	Failed
)

func describe(status int) string {
	switch status {
	case Queued, Sending:
		return "in progress"
	case Delivered:
		return "done"
	case Failed:
		return "needs retry"
	}
	return "unknown status"
}
```

When a function has a return type, Go needs a `return` after the switch unless every path already returns, which is why `"unknown status"` is there even though it's an unusual case.

## Your turn

Textio routes each message through a carrier based on the recipient's country calling code. Complete `carrierFor` using a `switch`:

| Code | Carrier |
|------|---------|
| `"+1"` | `"Alpha Mobile"` |
| `"+44"` | `"Beta Tel"` |
| `"+33"` or `"+49"` | `"Gamma Net"` |
| anything else | `"unsupported"` |

Try to handle `+33` and `+49` with a single `case`.

## Further reading

- [Go by Example: Switch](https://gobyexample.com/switch)
- [A Tour of Go: Switch with no condition](https://go.dev/tour/flowcontrol/11)
