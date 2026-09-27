---
title: testdata and Golden Files
quiz:
  - question: What's special about a directory named `testdata`?
    options:
      - text: '`go test` deletes it after every run'
      - text: The go tool ignores it when looking for packages, so it can hold any files, even `.go` files that don't compile
        correct: true
      - text: It's embedded into the test binary automatically
      - text: Only `_test.go` files may read from it
    explanation: |
      Directories named `testdata` (and ones starting with `.` or `_`) are
      skipped by `./...` patterns and the build. Tests run with the
      package directory as their working directory, so they open
      `testdata/...` with a plain relative path.
  - question: |
      A golden-file test fails after you intentionally changed the
      statement layout. What's the right workflow?
    options:
      - text: Delete the golden file and the test
      - text: Run the test with `-update`, then review the diff of the golden file before committing it
        correct: true
      - text: Copy the "got" output from the failure message into the test source
      - text: Add `t.Skip` until the layout settles down
    explanation: |
      `-update` regenerates the expected output, and version control
      shows exactly what changed. The review step is the important part:
      a golden file that's updated without reading the diff tests nothing.
---

Some outputs are too big to write inline: a full monthly statement, an exported CSV, an HTML report. And some inputs are too awkward to inline: a real bank export with odd encodings and quirks. Both belong in files next to the test.

## The testdata directory

Put test fixtures in a directory called `testdata` inside the package:

```text
ledgerly/
├── import.go
├── import_test.go
└── testdata/
    ├── bank-export-2026-03.csv
    ├── semicolons.csv
    └── TestWriteStatement.golden
```

The go tool gives `testdata` special treatment: it's never treated as a package, so `go build ./...` and `go vet ./...` ignore it. And `go test` runs each package's tests **in that package's directory**, so a relative path just works:

```go
f, err := os.Open("testdata/bank-export-2026-03.csv")
```

Or, combining the last lesson: `os.DirFS("testdata")` turns the whole directory into an `fs.FS`.

## Data-driven tests from a directory

A neat pattern is to make every file in a directory a test case, so adding a case means adding a file:

```go
func TestImportExamples(t *testing.T) {
	files, err := filepath.Glob("testdata/import/*.csv")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			f, err := os.Open(file)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			if _, err := ImportCSV(f); err != nil {
				t.Errorf("ImportCSV(%s): %v", file, err)
			}
		})
	}
}
```

When a user reports "your importer chokes on my bank's file", you drop an anonymised copy into `testdata/import/` and you have a failing test.

## Golden files

A **golden file** holds the expected output of a test. The test produces output, compares it with the file, and fails if they differ. The trick that makes golden files pleasant is an `-update` flag that rewrites them:

```go
var update = flag.Bool("update", false, "rewrite golden files")

func TestWriteStatement(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteStatement(&buf, sampleTransactions); err != nil {
		t.Fatal(err)
	}

	golden := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		if err := os.WriteFile(golden, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("statement doesn't match %s\ngot:\n%s\nwant:\n%s", golden, buf.Bytes(), want)
	}
}
```

The flag is declared at package level in the test file. `go test` parses test flags before running tests, so `*update` is set by then:

```text
$ go test -run TestWriteStatement -update ./ledgerly
$ git diff testdata/
+2026-03-01  rent        -$1200.00
+2026-03-02  salary       $3500.00
+2026-03-05  groceries     -$87.34
+TOTAL                    $2212.66
```

Using `t.Name()` for the file name keeps one golden file per test. With subtests, `t.Name()` contains a `/`, so either create the subdirectory or replace the slash.

### Golden file discipline

- **Always review the diff.** `-update` makes tests pass by definition. The value comes from a human looking at the change and agreeing with it.
- **Keep output deterministic.** Anything that varies between runs (timestamps, map iteration order, temp paths) must be fixed or scrubbed before comparing, or the golden file will never match.
- **Watch line endings.** Git on Windows may convert `\n` to `\r\n` on checkout. A `.gitattributes` line like `testdata/** -text` stops it.
- **Don't overuse them.** A golden file says "the output is exactly this", which breaks on every cosmetic change. For "the report mentions the total", `strings.Contains` is a better test.

Golden files shine for output that humans read and care about the exact layout of: formatted reports, generated code, CLI help text.
