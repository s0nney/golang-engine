---
title: 'Your Turn: Totals from a Folder of Statements'
quiz:
  - question: |
      `WriteSummary` ends with `fmt.Fprintf(w, "%-10s %10v\n", "TOTAL", total)`
      and then `return nil`. What's the bug?
    options:
      - text: '`%-10s` pads on the wrong side'
      - text: It ignores `Fprintf`'s error, so a failed final write is reported as success
        correct: true
      - text: '`Fprintf` can''t write to an `io.Writer`'
      - text: There's no bug
    explanation: |
      `Fprintf` returns `(n int, err error)`. Returning `err` from the last
      call (and checking the earlier ones) is what lets callers notice a
      full disk or a closed connection.
exercise:
  starter: |
    package main

    import (
    	"encoding/csv"
    	"errors"
    	"fmt"
    	"io"
    	"io/fs"
    	"os"
    	"strconv"
    	"strings"
    	"testing/fstest"
    )

    // LoadTotals imports every *.csv file at the root of fsys (not in
    // subdirectories) and returns the total amount per account. If a file
    // fails to import, the error must start with the file name, e.g.
    // "broken.csv: line 2: ...", and wrap the original error.
    func LoadTotals(fsys fs.FS) (map[string]Cents, error) {
    	// ?
    	return nil, nil
    }

    // WriteSummary writes one line per account, sorted by name, then a TOTAL
    // line, each formatted with "%-10s %10v\n". It returns the first write
    // error, if any.
    func WriteSummary(w io.Writer, totals map[string]Cents) error {
    	// ?
    	return nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    var ErrBadAmount = errors.New("ledgerly: bad amount")

    // ParseAmount parses "12", "12.3", "12.34" or "-0.05" into cents.
    func ParseAmount(s string) (Cents, error) {
    	digits, neg := strings.CutPrefix(s, "-")
    	whole, frac, hasDot := strings.Cut(digits, ".")
    	if whole == "" || (hasDot && (frac == "" || len(frac) > 2)) {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	for len(frac) < 2 {
    		frac += "0"
    	}
    	d, err1 := strconv.ParseUint(whole, 10, 40)
    	c, err2 := strconv.ParseUint(frac, 10, 8)
    	if err1 != nil || err2 != nil {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	n := Cents(d*100 + c)
    	if neg {
    		n = -n
    	}
    	return n, nil
    }

    type Transaction struct {
    	Date    string
    	Account string
    	Amount  Cents
    }

    // ImportCSV reads "date,account,amount" rows after a header line.
    func ImportCSV(r io.Reader) ([]Transaction, error) {
    	cr := csv.NewReader(r)
    	cr.FieldsPerRecord = 3
    	if _, err := cr.Read(); err != nil {
    		return nil, fmt.Errorf("reading header: %w", err)
    	}
    	var txns []Transaction
    	for {
    		rec, err := cr.Read()
    		if err == io.EOF {
    			return txns, nil
    		}
    		if err != nil {
    			return nil, err
    		}
    		amt, err := ParseAmount(rec[2])
    		if err != nil {
    			line, _ := cr.FieldPos(2)
    			return nil, fmt.Errorf("line %d: %w", line, err)
    		}
    		txns = append(txns, Transaction{Date: rec[0], Account: rec[1], Amount: amt})
    	}
    }

    func main() {
    	statements := fstest.MapFS{
    		"march.csv":        {Data: []byte("date,account,amount\n2026-03-01,rent,-1200.00\n2026-03-02,cash,500\n")},
    		"april.csv":        {Data: []byte("date,account,amount\n2026-04-01,rent,-1200.00\n2026-04-02,cash,750\n")},
    		"readme.txt":       {Data: []byte("not a statement")},
    		"archive/2025.csv": {Data: []byte("date,account,amount\n2025-12-01,rent,-9999.99\n")},
    	}
    	totals, err := LoadTotals(statements)
    	if err != nil {
    		fmt.Println("LoadTotals:", err)
    		return
    	}
    	fmt.Println("totals:", totals)
    	if err := WriteSummary(os.Stdout, totals); err != nil {
    		fmt.Println("WriteSummary:", err)
    	}
    }
  solution: |
    package main

    import (
    	"encoding/csv"
    	"errors"
    	"fmt"
    	"io"
    	"io/fs"
    	"maps"
    	"os"
    	"slices"
    	"strconv"
    	"strings"
    	"testing/fstest"
    )

    // LoadTotals imports every *.csv file at the root of fsys (not in
    // subdirectories) and returns the total amount per account. If a file
    // fails to import, the error must start with the file name, e.g.
    // "broken.csv: line 2: ...", and wrap the original error.
    func LoadTotals(fsys fs.FS) (map[string]Cents, error) {
    	names, err := fs.Glob(fsys, "*.csv")
    	if err != nil {
    		return nil, err
    	}
    	totals := map[string]Cents{}
    	for _, name := range names {
    		f, err := fsys.Open(name)
    		if err != nil {
    			return nil, err
    		}
    		txns, err := ImportCSV(f)
    		f.Close()
    		if err != nil {
    			return nil, fmt.Errorf("%s: %w", name, err)
    		}
    		for _, tx := range txns {
    			totals[tx.Account] += tx.Amount
    		}
    	}
    	return totals, nil
    }

    // WriteSummary writes one line per account, sorted by name, then a TOTAL
    // line, each formatted with "%-10s %10v\n". It returns the first write
    // error, if any.
    func WriteSummary(w io.Writer, totals map[string]Cents) error {
    	var total Cents
    	for _, account := range slices.Sorted(maps.Keys(totals)) {
    		if _, err := fmt.Fprintf(w, "%-10s %10v\n", account, totals[account]); err != nil {
    			return err
    		}
    		total += totals[account]
    	}
    	_, err := fmt.Fprintf(w, "%-10s %10v\n", "TOTAL", total)
    	return err
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    var ErrBadAmount = errors.New("ledgerly: bad amount")

    // ParseAmount parses "12", "12.3", "12.34" or "-0.05" into cents.
    func ParseAmount(s string) (Cents, error) {
    	digits, neg := strings.CutPrefix(s, "-")
    	whole, frac, hasDot := strings.Cut(digits, ".")
    	if whole == "" || (hasDot && (frac == "" || len(frac) > 2)) {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	for len(frac) < 2 {
    		frac += "0"
    	}
    	d, err1 := strconv.ParseUint(whole, 10, 40)
    	c, err2 := strconv.ParseUint(frac, 10, 8)
    	if err1 != nil || err2 != nil {
    		return 0, fmt.Errorf("%w: %q", ErrBadAmount, s)
    	}
    	n := Cents(d*100 + c)
    	if neg {
    		n = -n
    	}
    	return n, nil
    }

    type Transaction struct {
    	Date    string
    	Account string
    	Amount  Cents
    }

    // ImportCSV reads "date,account,amount" rows after a header line.
    func ImportCSV(r io.Reader) ([]Transaction, error) {
    	cr := csv.NewReader(r)
    	cr.FieldsPerRecord = 3
    	if _, err := cr.Read(); err != nil {
    		return nil, fmt.Errorf("reading header: %w", err)
    	}
    	var txns []Transaction
    	for {
    		rec, err := cr.Read()
    		if err == io.EOF {
    			return txns, nil
    		}
    		if err != nil {
    			return nil, err
    		}
    		amt, err := ParseAmount(rec[2])
    		if err != nil {
    			line, _ := cr.FieldPos(2)
    			return nil, fmt.Errorf("line %d: %w", line, err)
    		}
    		txns = append(txns, Transaction{Date: rec[0], Account: rec[1], Amount: amt})
    	}
    }

    func main() {
    	statements := fstest.MapFS{
    		"march.csv":        {Data: []byte("date,account,amount\n2026-03-01,rent,-1200.00\n2026-03-02,cash,500\n")},
    		"april.csv":        {Data: []byte("date,account,amount\n2026-04-01,rent,-1200.00\n2026-04-02,cash,750\n")},
    		"readme.txt":       {Data: []byte("not a statement")},
    		"archive/2025.csv": {Data: []byte("date,account,amount\n2025-12-01,rent,-9999.99\n")},
    	}
    	totals, err := LoadTotals(statements)
    	if err != nil {
    		fmt.Println("LoadTotals:", err)
    		return
    	}
    	fmt.Println("totals:", totals)
    	if err := WriteSummary(os.Stdout, totals); err != nil {
    		fmt.Println("WriteSummary:", err)
    	}
    }
  tests: |
    package main

    import (
    	"bytes"
    	"errors"
    	"maps"
    	"strings"
    	"testing"
    	"testing/fstest"
    )

    func statements() fstest.MapFS {
    	return fstest.MapFS{
    		"march.csv": {Data: []byte("date,account,amount\n2026-03-01,rent,-1200.00\n2026-03-02,cash,500\n2026-03-09,groceries,-87.34\n")},
    		"april.csv": {Data: []byte("date,account,amount\n2026-04-01,rent,-1200.00\n2026-04-02,cash,750.5\n")},
    	}
    }

    func TestLoadTotals(t *testing.T) {
    	got, err := LoadTotals(statements())
    	if err != nil {
    		t.Fatalf("LoadTotals: unexpected error: %v", err)
    	}
    	want := map[string]Cents{"rent": -240000, "cash": 125050, "groceries": -8734}
    	if !maps.Equal(got, want) {
    		t.Errorf("LoadTotals = %v, want %v", got, want)
    	}
    }

    func TestLoadTotalsOnlyRootCSV(t *testing.T) {
    	fsys := statements()
    	fsys["notes.txt"] = &fstest.MapFile{Data: []byte("remember to pay rent")}
    	fsys["archive/2025.csv"] = &fstest.MapFile{Data: []byte("date,account,amount\n2025-12-01,rent,-9999.99\n")}
    	got, err := LoadTotals(fsys)
    	if err != nil {
    		t.Fatalf("LoadTotals with notes.txt and archive/2025.csv present: unexpected error: %v", err)
    	}
    	if got["rent"] != -240000 {
    		t.Errorf("rent total = %v, want -$2400.00 (only *.csv files at the root count; skip notes.txt and archive/)", got["rent"])
    	}
    }

    func TestLoadTotalsEmpty(t *testing.T) {
    	got, err := LoadTotals(fstest.MapFS{})
    	if err != nil {
    		t.Fatalf("LoadTotals(empty FS): unexpected error: %v", err)
    	}
    	if len(got) != 0 {
    		t.Errorf("LoadTotals(empty FS) = %v, want no accounts", got)
    	}
    }

    func TestLoadTotalsBadFile(t *testing.T) {
    	fsys := statements()
    	fsys["broken.csv"] = &fstest.MapFile{Data: []byte("date,account,amount\n2026-05-01,rent,12.345\n")}
    	_, err := LoadTotals(fsys)
    	if err == nil {
    		t.Fatal("LoadTotals with a broken broken.csv returned no error")
    	}
    	if !errors.Is(err, ErrBadAmount) {
    		t.Errorf("LoadTotals error = %v, want it to wrap ErrBadAmount (use %%w)", err)
    	}
    	if !strings.HasPrefix(err.Error(), "broken.csv: ") {
    		t.Errorf("LoadTotals error = %q, want it to start with \"broken.csv: \"", err)
    	}
    }

    func TestWriteSummary(t *testing.T) {
    	var buf bytes.Buffer
    	err := WriteSummary(&buf, map[string]Cents{"rent": -240000, "cash": 125050, "groceries": -8734})
    	if err != nil {
    		t.Fatalf("WriteSummary: unexpected error: %v", err)
    	}
    	want := "" +
    		"cash         $1250.50\n" +
    		"groceries     -$87.34\n" +
    		"rent        -$2400.00\n" +
    		"TOTAL       -$1236.84\n"
    	if got := buf.String(); got != want {
    		t.Errorf("WriteSummary wrote:\n%s\nwant:\n%s", got, want)
    	}
    }

    func TestWriteSummaryEmpty(t *testing.T) {
    	var buf bytes.Buffer
    	if err := WriteSummary(&buf, nil); err != nil {
    		t.Fatalf("WriteSummary(nil): unexpected error: %v", err)
    	}
    	if got, want := buf.String(), "TOTAL           $0.00\n"; got != want {
    		t.Errorf("WriteSummary(nil) wrote %q, want %q", got, want)
    	}
    }

    // failWriter accepts n writes, then fails.
    type failWriter struct{ n int }

    var errDiskFull = errors.New("disk full")

    func (w *failWriter) Write(p []byte) (int, error) {
    	if w.n == 0 {
    		return 0, errDiskFull
    	}
    	w.n--
    	return len(p), nil
    }

    func TestWriteSummaryWriteError(t *testing.T) {
    	totals := map[string]Cents{"rent": -240000, "cash": 125050, "groceries": -8734}
    	for n := range 4 {
    		err := WriteSummary(&failWriter{n: n}, totals)
    		if !errors.Is(err, errDiskFull) {
    			t.Errorf("WriteSummary to a writer that fails after %d writes returned %v, want the write error", n, err)
    		}
    	}
    }
---

Time to write I/O code that's testable from the start. Ledgerly users keep their monthly bank statements in a folder, and they want a summary of every account across all of them.

`ImportCSV` (from this chapter's first lesson) is already written and takes an `io.Reader`. You'll write the two functions around it, and the grader will test them without touching a single real file.

## Your task

1. **`LoadTotals(fsys fs.FS)`** imports every file matching `*.csv` at the **root** of `fsys` and returns the total amount per account. Other files (like `readme.txt`) and anything in subdirectories are ignored. `fs.Glob` does the matching for you. Remember to close each file.

   If `ImportCSV` fails, return an error that starts with the file name and wraps the original with `%w`: `broken.csv: line 2: ledgerly: bad amount: "12.345"`. An empty file system is not an error; it just has no accounts.

2. **`WriteSummary(w io.Writer, totals)`** writes one line per account, **sorted by account name**, then a `TOTAL` line, each with the format `"%-10s %10v\n"`:

   ```text
   cash         $1250.50
   groceries     -$87.34
   rent        -$2400.00
   TOTAL       -$1236.84
   ```

   It must return the first error from writing. `slices.Sorted(maps.Keys(totals))` gives you the sorted names (you'll need to import `maps` and `slices`).

The grader feeds `LoadTotals` a `fstest.MapFS`, captures `WriteSummary`'s output in a `bytes.Buffer`, and uses a writer that fails after a few writes to check your error handling. **Run** tries both functions on a small `MapFS` and prints to `os.Stdout`.
