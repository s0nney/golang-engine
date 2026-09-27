---
title: File Systems with fs.FS and fstest.MapFS
quiz:
  - question: |
      Which paths are valid for an `fs.FS`?
    options:
      - text: '`/statements/march.csv`'
      - text: '`./statements/march.csv`'
      - text: '`statements/march.csv`'
        correct: true
      - text: '`statements\march.csv` on Windows'
    explanation: |
      `fs.FS` paths are always slash-separated and unrooted: no leading
      `/`, no `./` or `..` elements, and `/` even on Windows. `fs.ValidPath`
      checks this.
  - question: |
      What does `LoadDir` see here?

      ```go
      fsys := fstest.MapFS{
          "statements/march.csv": {Data: []byte("...")},
      }
      entries, _ := fs.ReadDir(fsys, "statements")
      ```
    options:
      - text: An error, because the `statements` directory was never created
      - text: One entry, `march.csv`; MapFS invents parent directories automatically
        correct: true
      - text: One entry, `statements/march.csv`
      - text: Nothing, because MapFS has no directories
    explanation: |
      MapFS synthesises any directory implied by a file path, so you only
      list files. `ReadDir` returns entries with base names.
---

`io.Reader` covers one file. When code works with a whole directory, finding statement files, reading several, skipping the ones it doesn't recognise, take an `fs.FS` instead.

## fs.FS

`io/fs.FS` is a one-method interface for a read-only file tree:

```go
type FS interface {
	Open(name string) (fs.File, error)
}
```

Helper functions in `io/fs` build on it: `fs.ReadFile`, `fs.ReadDir`, `fs.Glob`, `fs.WalkDir`, `fs.Stat`. Here's Ledgerly loading every statement in a folder:

```go
// LoadDir imports every *.csv file at the root of fsys.
func LoadDir(fsys fs.FS) ([]Transaction, error) {
	names, err := fs.Glob(fsys, "*.csv")
	if err != nil {
		return nil, err
	}
	var all []Transaction
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
		all = append(all, txns...)
	}
	return all, nil
}
```

In production, `os.DirFS("/home/you/statements")` gives you an `fs.FS` backed by a real directory. Other implementations include `embed.FS` (files compiled into the binary) and `zip.Reader`.

## fstest.MapFS

In a test, build the file tree in memory with `testing/fstest.MapFS`, a `map[string]*fstest.MapFile`:

```go
func TestLoadDir(t *testing.T) {
	fsys := fstest.MapFS{
		"march.csv": {Data: []byte("date,account,amount\n2026-03-01,rent,-1200.00\n")},
		"april.csv": {Data: []byte("date,account,amount\n2026-04-01,rent,-1200.00\n")},
		"notes.txt": {Data: []byte("not a statement")},
	}

	txns, err := LoadDir(fsys)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(txns) != 2 {
		t.Errorf("LoadDir loaded %d transactions, want 2 (notes.txt should be skipped)", len(txns))
	}
}
```

Things to know:

- **Directories are implied.** `"2026/march.csv"` makes a `2026` directory automatically.
- **`MapFile` has more fields**: `Mode`, `ModTime` and `Sys`, if your code cares about permissions or timestamps.
- **Sorted listings.** `fs.ReadDir` and `fs.Glob` return names in sorted order, so `april.csv` comes before `march.csv`. If your code's output depends on order, that's now deterministic and testable.
- **Adding a file is one line**, so testing the error path is cheap:

```go
fsys["broken.csv"] = &fstest.MapFile{Data: []byte("date,account,amount\n2026-03-01,rent,x\n")}
_, err = LoadDir(fsys)
// err: broken.csv: line 2: ledgerly: bad amount: "x"
```

## fstest.TestFS

If you ever write your *own* `fs.FS` (say, one that reads statements out of a bank's export archive), `fstest.TestFS` checks that it behaves like a proper file system: that `Open`, `ReadDir`, `Stat` and friends agree with each other.

```go
if err := fstest.TestFS(myFS, "march.csv", "april.csv"); err != nil {
	t.Fatal(err)
}
```

The listed names must exist. It walks the whole tree, so it's thorough.

## When you need real files

`fs.FS` is read-only. For code that *writes* files, or that needs `os`-level behaviour (permissions, renames, locking), use real files in `t.TempDir()`, as you did in chapter 2. A good split is: code that reads a tree takes an `fs.FS`, code that writes takes an `io.Writer` or a directory path, and only the latter touches the disk in tests.
