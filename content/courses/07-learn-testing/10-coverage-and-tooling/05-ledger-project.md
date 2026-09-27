---
title: 'Build It: Ledgerly Import Checks'
exercise:
  starter: |
    package main

    import "fmt"

    // An importer accepts account,cents records and sums cents by account.
    type Importer func(string) (map[string]int64, error)

    func checkImporter(importer Importer) error {
    	return nil
    }

    func main() {
    	err := checkImporter(func(string) (map[string]int64, error) {
    		return map[string]int64{}, nil
    	})
    	fmt.Println("broken importer detected:", err != nil)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"maps"
    )

    type Importer func(string) (map[string]int64, error)

    func checkImporter(importer Importer) error {
    	for _, tc := range []struct {
    		input string
    		want  map[string]int64
    	}{
    		{"", map[string]int64{}},
    		{"cash,125\ncash,-25\nrent,-40\n", map[string]int64{"cash": 100, "rent": -40}},
    		{"cash,0\n", map[string]int64{"cash": 0}},
    	} {
    		got, err := importer(tc.input)
    		if err != nil || !maps.Equal(got, tc.want) {
    			return fmt.Errorf("import(%q) = %v, %v; want %v, nil", tc.input, got, err, tc.want)
    		}
    	}
    	for _, input := range []string{"cash,nope\n", "cash\n", "cash,1,extra\n"} {
    		if _, err := importer(input); err == nil {
    			return fmt.Errorf("import(%q) accepted invalid record", input)
    		}
    	}
    	return nil
    }

    func main() {
    	err := checkImporter(func(string) (map[string]int64, error) {
    		return map[string]int64{}, nil
    	})
    	fmt.Println("broken importer detected:", err != nil)
    }
  tests: |
    package main

    import (
    	"encoding/csv"
    	"errors"
    	"io"
    	"strconv"
    	"strings"
    	"testing"
    )

    func reference(mode string) Importer {
    	return func(input string) (map[string]int64, error) {
    		r := csv.NewReader(strings.NewReader(input))
    		r.FieldsPerRecord = -1
    		out := map[string]int64{}
    		for {
    			record, err := r.Read()
    			if err == io.EOF {
    				break
    			}
    			if err != nil {
    				return nil, err
    			}
    			if len(record) != 2 {
    				if mode == "ignore-fields" {
    					continue
    				}
    				return nil, errors.New("two fields required")
    			}
    			n, err := strconv.ParseInt(record[1], 10, 64)
    			if err != nil && mode != "ignore-amount" {
    				return nil, err
    			}
    			if mode == "positive-only" && n < 0 {
    				continue
    			}
    			if mode == "drop-zero" && n == 0 {
    				continue
    			}
    			if mode == "overwrite" {
    				out[record[0]] = n
    			} else {
    				out[record[0]] += n
    			}
    		}
    		if mode == "reject-empty" && input == "" {
    			return nil, errors.New("empty")
    		}
    		return out, nil
    	}
    }

    func TestChecks(t *testing.T) {
    	if err := checkImporter(reference("correct")); err != nil {
    		t.Fatalf("rejected correct importer: %v", err)
    	}
    	for _, mode := range []string{"ignore-fields", "ignore-amount", "positive-only", "drop-zero", "overwrite", "reject-empty"} {
    		t.Run(mode, func(t *testing.T) {
    			if err := checkImporter(reference(mode)); err == nil {
    				t.Errorf("checks missed %s bug", mode)
    			}
    		})
    	}
    }
---

Ledgerly has an importer and a worrying test suite: the tests pass even when debits
disappear. Your final task is to test a contract, including implementations you
haven't seen. Combine tables, injected dependencies, map comparison, and error cases.

## The contract

An importer receives CSV text containing `account,cents` records. Amounts are signed
base-10 `int64` values. It sums repeated accounts, keeps separate accounts separate,
preserves zero-balance accounts, and accepts empty input. Malformed amounts and
records with any field count other than two return an error. Inputs in this exercise
do not overflow their account totals.

Complete `checkImporter`. Return `nil` for a correct implementation and a helpful
error for a contract violation. Injecting the importer makes it a test double seam:
your checks must discover bugs by calling the function, without inspecting its code.

Use a table of successful cases and compare maps without depending on iteration
order. Add error cases for both too few and too many fields and an invalid amount.
Avoid asserting exact error messages: the contract promises an error, not its prose.
For empty input, a nil map and an empty map are both acceptable.

```go
if err != nil || !maps.Equal(got, want) {
	return fmt.Errorf("import(%q) = %v, %v; want %v, nil", input, got, err, want)
}
```

The browser editor supplies `main.go`, so these checks return errors for the hidden
test harness to inspect. In a local project, put the same cases in `import_test.go`,
use named `t.Run` subtests, and report each failure with `t.Errorf`.

**Run** should report `broken importer detected: true`. **Submit** tries your suite
against a correct importer and versions with deliberately introduced faults: missing
debits, overwritten totals, dropped zero balances, and swallowed parse errors.

After it passes, take one of those bugs to a local Ledgerly project: watch your test
fail, fix the bug, and run `go test ./...` and `go vet ./...`. Add a benchmark before
trying to optimize the importer; use fuzzing to explore more malformed CSV inputs.

Next, [Learn HTTP Clients in Go](/courses/learn-http-clients) applies dependency
injection and repeatable tests to APIs, timeouts, retries, and pagination.
