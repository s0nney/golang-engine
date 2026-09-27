---
title: gofmt and go fix
quiz:
  - question: What does `gofmt -l .` do?
    options:
      - text: Formats every file in place
      - text: Lists the files whose formatting differs from gofmt's, without changing them
        correct: true
      - text: Prints line numbers of long lines
      - text: Lints the code for bugs
    explanation: |
      `-l` lists files that need formatting, which makes it perfect for a
      CI check (fail if the output isn't empty). `-w` rewrites files in
      place, and `-d` prints a diff.
  - question: |
      You run `go fix ./...` on a Go 1.27 module. What happens to this code?

      ```go
      r := Report{Stats: Stats{Count: 3}, Title: "March"}
      ```
    options:
      - text: Nothing, it's already modern
      - text: 'The `embedlit` modernizer rewrites it to `Report{Count: 3, Title: "March"}`, using Go 1.27''s promoted fields in struct literals'
        correct: true
      - text: It's deleted as dead code
      - text: '`go fix` reports an error because `Stats` is embedded'
    explanation: |
      Go 1.27 lets struct literals set promoted fields of embedded structs
      directly, and `go fix` has a modernizer to use it. Modernizers only
      apply when the module's `go` version allows the newer feature.
---

The last two tools in this chapter don't find bugs. They keep code *uniform* and *current*, which matters more than it sounds when a codebase lives for years and has dozens of authors. You met both in [Learn Go](/courses/learn-go/packages-and-modules/go-fmt-vet-and-fix); this lesson covers the flags and the workflow you'll use on a real project.

## gofmt

There is exactly one way to format Go code: the way `gofmt` does it. Tabs for indentation, aligned comments and struct fields, imports sorted, no arguments about brace style.

```text
gofmt -l .        # list files that aren't formatted
gofmt -d file.go  # show the diff it would apply
gofmt -w .        # rewrite files in place
go fmt ./...      # runs gofmt -l -w on packages
```

Your editor almost certainly runs it on save. In CI, check it with `test -z "$(gofmt -l .)"` so unformatted code can't be merged. It also reformats doc comments, as you saw in the Examples and Docs chapter.

`gofmt -s` applies a few simplifications, such as removing redundant types in composite literals (`[]Transaction{Transaction{Account: "rent"}}` to `[]Transaction{{Account: "rent"}}`). `goimports`, a separate tool from `golang.org/x/tools`, does everything gofmt does and also adds and removes imports. Most editors use it.

## go fix and modernizers

Go keeps adding better ways to write common things: `any` for `interface{}`, `min` and `max`, `range` over integers, `slices.Contains`, `errors.AsType`. Old code still works, since Go's compatibility promise means it'll compile forever. But reading a mix of old and new idioms is tiring.

Since Go 1.26, `go fix` runs **modernizers**: analyzers that find old patterns and rewrite them to the current idiom. Here's `go fix -diff` on an older part of Ledgerly (trimmed):

```text
$ go fix -diff ./...
-var imported int64
+var imported atomic.Int64

 func HasAccount(names []string, want string) bool {
-	for _, n := range names {
-		if n == want {
-			return true
-		}
-	}
-	return false
+	return slices.Contains(names, want)
 }

-	for i := 0; i < len(lines); i++ {
+	for i := range lines {
 		line := lines[i]
-		wg.Add(1)
-		go func() {
-			defer wg.Done()
+		wg.Go(func() {
 			process(line)
-			atomic.AddInt64(&imported, 1)
-		}()
+			imported.Add(1)
+		})

-	var le *LineError
-	if errors.As(err, &le) {
+	if le, ok := errors.AsType[*LineError](err); ok {

-	for _, f := range strings.Split(s, ",") {
+	for f := range strings.SplitSeq(s, ",") {

-	return Report{Stats: Stats{Count: 0}, Title: title}
+	return Report{Count: 0, Title: title}

-func Opt(v interface{}) *int {
+func Opt(v any) *int {
```

`-diff` shows what would change without touching anything. Plain `go fix ./...` applies it. Each fix is designed to keep behaviour identical, and a modernizer only fires when the module's `go` version in `go.mod` allows the newer feature, so a module that says `go 1.21` won't get `wg.Go` (Go 1.25) suggestions.

## Two newer modernizers

`go tool fix help` lists every modernizer, and the list grows with most releases. Two recent ones:

- **`embedlit`** uses 1.27's new struct literal rule: promoted fields of an embedded struct can be set directly, so `Report{Stats: Stats{Count: 0}, Title: title}` becomes `Report{Count: 0, Title: title}`. (That's the last hunk above.)
- **`atomictypes`** replaces the old `sync/atomic` functions on plain integers (`atomic.AddInt64(&imported, 1)`) with the typed wrappers (`var imported atomic.Int64` and `imported.Add(1)`). The typed versions make it impossible to accidentally read the variable without an atomic operation.

Others include `any`, `minmax`, `rangeint`, `slicescontains`, `stringsseq`, `stringscut`, `waitgroupgo`, `newexpr` (use `new(expr)`), `errorsastype` and `testingcontext` (replace a hand-made `context.WithCancel` in tests with `t.Context()`).

To run just one: `go fix -errorsastype ./...`. `go tool fix help errorsastype` explains exactly what it rewrites and when.

## A workflow

1. Upgrade the `go` line in `go.mod`, since that's what unlocks newer modernizers.
2. Run `go fix -diff ./...` and skim the result.
3. Run `go fix ./...`, then `go test ./...`.
4. Commit it **on its own**, separately from any behaviour change, so reviewers can see it's purely mechanical.

The same analyzers power the "this can be simplified" hints in editors that use gopls, so you'll often see these suggestions while typing.
