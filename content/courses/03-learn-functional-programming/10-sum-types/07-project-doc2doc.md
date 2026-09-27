---
title: 'Project: Assemble Doc2Doc'
quiz:
  - question: 'In the finished pipeline `Chain(markdownToHTML, trimSpace, wordLimit(500))`, which pieces are pure functions?'
    options:
      - text: Only `trimSpace`
      - text: None of them, because they're wrapped in middleware
      - text: All of them, and so is the `Converter` that `Chain` returns
        correct: true
      - text: Only `markdownToHTML`
    explanation: |
      Every piece takes a string and returns a string without touching anything
      else, and wrapping pure functions in pure wrappers gives another pure
      function. That's why the whole pipeline can be tested with plain
      input/output tables.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    	"strings"
    )

    // ---- Blocks: a sum type (chapter 10) ----

    type Block interface{ isBlock() }

    type Heading struct {
    	Level int
    	Text  string
    }
    type Paragraph struct{ Text string }
    type CodeBlock struct{ Lang, Code string }
    type List struct{ Items []string }

    func (Heading) isBlock()   {}
    func (Paragraph) isBlock() {}
    func (CodeBlock) isBlock() {}
    func (List) isBlock()      {}

    // ---- The parser: a lazy iterator (chapter 9), already written ----

    // parse turns Markdown into blocks, one at a time.
    func parse(doc string) iter.Seq[Block] {
    	return func(yield func(Block) bool) {
    		var items, code []string
    		inCode, lang := false, ""
    		flushList := func() bool {
    			if len(items) == 0 {
    				return true
    			}
    			b := List{Items: items}
    			items = nil
    			return yield(b)
    		}
    		for line := range strings.Lines(doc) {
    			line = strings.TrimRight(line, "\r\n")
    			if inCode {
    				if line == "```" {
    					inCode = false
    					if !yield(CodeBlock{Lang: lang, Code: strings.Join(code, "\n")}) {
    						return
    					}
    					code = nil
    				} else {
    					code = append(code, line)
    				}
    				continue
    			}
    			if item, ok := strings.CutPrefix(line, "- "); ok {
    				items = append(items, item)
    				continue
    			}
    			if !flushList() {
    				return
    			}
    			var b Block
    			switch {
    			case strings.HasPrefix(line, "```"):
    				inCode, lang = true, line[3:]
    			case strings.HasPrefix(line, "#"):
    				level := len(line) - len(strings.TrimLeft(line, "#"))
    				b = Heading{Level: level, Text: strings.TrimSpace(line[level:])}
    			case strings.TrimSpace(line) != "":
    				b = Paragraph{Text: line}
    			}
    			if b != nil && !yield(b) {
    				return
    			}
    		}
    		flushList()
    	}
    }

    // ---- Middleware (chapter 8), already written ----

    type Converter func(string) string

    type Middleware func(Converter) Converter

    func Chain(c Converter, mws ...Middleware) Converter {
    	for _, mw := range slices.Backward(mws) {
    		c = mw(c)
    	}
    	return c
    }

    func trimSpace(next Converter) Converter {
    	return func(doc string) string {
    		return next(strings.TrimSpace(doc))
    	}
    }

    // ---- Your part ----

    // toHTML renders one block (chapter 10). Escape every piece of text with
    // html.EscapeString:
    //
    //	Heading{2, "Setup"}        -> <h2>Setup</h2>
    //	Paragraph{"a & b"}         -> <p>a &amp; b</p>
    //	CodeBlock{"go", "x := 1"}  -> <pre><code class="language-go">x := 1</code></pre>
    //	CodeBlock{"", "x := 1"}    -> <pre><code>x := 1</code></pre>
    //	List{[]string{"a", "b"}}   -> <ul><li>a</li><li>b</li></ul>
    //	anything else              -> <!-- unknown block -->
    func toHTML(b Block) string {
    	// ?
    	return ""
    }

    // Map lazily applies f to every value of seq (chapter 9). Stop as soon as
    // yield returns false.
    func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
    	return func(yield func(U) bool) {
    		// ?
    	}
    }

    // markdownToHTML parses doc, renders every block with toHTML and joins
    // the results with "\n" (chapters 2 and 5).
    func markdownToHTML(doc string) string {
    	// ?
    	return doc
    }

    // wordLimit returns a middleware (chapters 7 and 8): documents with more
    // than n words (strings.Fields) become "<p>document too long</p>" without
    // calling next; shorter ones pass straight through.
    func wordLimit(n int) Middleware {
    	return func(next Converter) Converter {
    		// ?
    		return next
    	}
    }

    func main() {
    	convert := Chain(markdownToHTML, trimSpace, wordLimit(50))

    	doc := `
    # Doc2Doc
    Converts **all** your documents & more.

    ## Install
    ` + "```sh\ngo install doc2doc@latest\n```" + `
    - fast
    - pure <functions>
    `
    	fmt.Println(convert(doc))
    	fmt.Println(convert(strings.Repeat("word ", 51)))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"html"
    	"iter"
    	"slices"
    	"strings"
    )

    // ---- Blocks: a sum type (chapter 10) ----

    type Block interface{ isBlock() }

    type Heading struct {
    	Level int
    	Text  string
    }
    type Paragraph struct{ Text string }
    type CodeBlock struct{ Lang, Code string }
    type List struct{ Items []string }

    func (Heading) isBlock()   {}
    func (Paragraph) isBlock() {}
    func (CodeBlock) isBlock() {}
    func (List) isBlock()      {}

    // ---- The parser: a lazy iterator (chapter 9), already written ----

    // parse turns Markdown into blocks, one at a time.
    func parse(doc string) iter.Seq[Block] {
    	return func(yield func(Block) bool) {
    		var items, code []string
    		inCode, lang := false, ""
    		flushList := func() bool {
    			if len(items) == 0 {
    				return true
    			}
    			b := List{Items: items}
    			items = nil
    			return yield(b)
    		}
    		for line := range strings.Lines(doc) {
    			line = strings.TrimRight(line, "\r\n")
    			if inCode {
    				if line == "```" {
    					inCode = false
    					if !yield(CodeBlock{Lang: lang, Code: strings.Join(code, "\n")}) {
    						return
    					}
    					code = nil
    				} else {
    					code = append(code, line)
    				}
    				continue
    			}
    			if item, ok := strings.CutPrefix(line, "- "); ok {
    				items = append(items, item)
    				continue
    			}
    			if !flushList() {
    				return
    			}
    			var b Block
    			switch {
    			case strings.HasPrefix(line, "```"):
    				inCode, lang = true, line[3:]
    			case strings.HasPrefix(line, "#"):
    				level := len(line) - len(strings.TrimLeft(line, "#"))
    				b = Heading{Level: level, Text: strings.TrimSpace(line[level:])}
    			case strings.TrimSpace(line) != "":
    				b = Paragraph{Text: line}
    			}
    			if b != nil && !yield(b) {
    				return
    			}
    		}
    		flushList()
    	}
    }

    // ---- Middleware (chapter 8), already written ----

    type Converter func(string) string

    type Middleware func(Converter) Converter

    func Chain(c Converter, mws ...Middleware) Converter {
    	for _, mw := range slices.Backward(mws) {
    		c = mw(c)
    	}
    	return c
    }

    func trimSpace(next Converter) Converter {
    	return func(doc string) string {
    		return next(strings.TrimSpace(doc))
    	}
    }

    // ---- Your part ----

    func toHTML(b Block) string {
    	switch b := b.(type) {
    	case Heading:
    		return fmt.Sprintf("<h%d>%s</h%d>", b.Level, html.EscapeString(b.Text), b.Level)
    	case Paragraph:
    		return "<p>" + html.EscapeString(b.Text) + "</p>"
    	case CodeBlock:
    		if b.Lang == "" {
    			return "<pre><code>" + html.EscapeString(b.Code) + "</code></pre>"
    		}
    		return `<pre><code class="language-` + b.Lang + `">` + html.EscapeString(b.Code) + "</code></pre>"
    	case List:
    		var sb strings.Builder
    		sb.WriteString("<ul>")
    		for _, item := range b.Items {
    			sb.WriteString("<li>" + html.EscapeString(item) + "</li>")
    		}
    		sb.WriteString("</ul>")
    		return sb.String()
    	default:
    		return "<!-- unknown block -->"
    	}
    }

    func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
    	return func(yield func(U) bool) {
    		for v := range seq {
    			if !yield(f(v)) {
    				return
    			}
    		}
    	}
    }

    func markdownToHTML(doc string) string {
    	return strings.Join(slices.Collect(Map(parse(doc), toHTML)), "\n")
    }

    func wordLimit(n int) Middleware {
    	return func(next Converter) Converter {
    		return func(doc string) string {
    			if len(strings.Fields(doc)) > n {
    				return "<p>document too long</p>"
    			}
    			return next(doc)
    		}
    	}
    }

    func main() {
    	convert := Chain(markdownToHTML, trimSpace, wordLimit(50))

    	doc := `
    # Doc2Doc
    Converts **all** your documents & more.

    ## Install
    ` + "```sh\ngo install doc2doc@latest\n```" + `
    - fast
    - pure <functions>
    `
    	fmt.Println(convert(doc))
    	fmt.Println(convert(strings.Repeat("word ", 51)))
    }
  tests: |
    package main

    import (
    	"slices"
    	"strings"
    	"testing"
    )

    type Image struct{ Src string }

    func (Image) isBlock() {}

    func TestToHTML(t *testing.T) {
    	for _, tt := range []struct {
    		b    Block
    		want string
    	}{
    		{Heading{Level: 1, Text: "Doc2Doc"}, "<h1>Doc2Doc</h1>"},
    		{Heading{Level: 3, Text: "Q&A"}, "<h3>Q&amp;A</h3>"},
    		{Paragraph{Text: "a < b & c"}, "<p>a &lt; b &amp; c</p>"},
    		{CodeBlock{Lang: "go", Code: "if a < b {\n}"}, `<pre><code class="language-go">if a &lt; b {` + "\n" + `}</code></pre>`},
    		{CodeBlock{Code: "ls"}, "<pre><code>ls</code></pre>"},
    		{List{Items: []string{"fast", "<pure>"}}, "<ul><li>fast</li><li>&lt;pure&gt;</li></ul>"},
    		{Image{Src: "cat.png"}, "<!-- unknown block -->"},
    		{nil, "<!-- unknown block -->"},
    	} {
    		if got := toHTML(tt.b); got != tt.want {
    			t.Errorf("toHTML(%#v) = %q, want %q", tt.b, got, tt.want)
    		}
    	}
    }

    func TestMap(t *testing.T) {
    	double := func(n int) int { return n * 2 }
    	got := slices.Collect(Map(slices.Values([]int{1, 2, 3}), double))
    	if !slices.Equal(got, []int{2, 4, 6}) {
    		t.Errorf("Map([1 2 3], double) = %v, want [2 4 6]", got)
    	}
    	defer func() {
    		if r := recover(); r != nil {
    			t.Fatalf("Map kept yielding after the loop broke: %v", r)
    		}
    	}()
    	calls := 0
    	for range Map(slices.Values([]int{1, 2, 3, 4}), func(n int) int { calls++; return n }) {
    		break
    	}
    	if calls != 1 {
    		t.Errorf("Map called f %d times for a loop that stopped after one value, want 1 (be lazy)", calls)
    	}
    }

    func TestMarkdownToHTML(t *testing.T) {
    	doc := "# Title\nHello & welcome.\n\n```go\nx := 1\n```\n- a\n- b\n## End"
    	want := strings.Join([]string{
    		"<h1>Title</h1>",
    		"<p>Hello &amp; welcome.</p>",
    		`<pre><code class="language-go">x := 1</code></pre>`,
    		"<ul><li>a</li><li>b</li></ul>",
    		"<h2>End</h2>",
    	}, "\n")
    	if got := markdownToHTML(doc); got != want {
    		t.Errorf("markdownToHTML(%q) =\n%s\nwant\n%s", doc, got, want)
    	}
    	if got := markdownToHTML(""); got != "" {
    		t.Errorf("markdownToHTML(\"\") = %q, want \"\"", got)
    	}
    }

    func TestWordLimit(t *testing.T) {
    	calls := 0
    	echo := func(doc string) string { calls++; return "[" + doc + "]" }
    	c := wordLimit(3)(echo)
    	if got := c("one two three"); got != "[one two three]" {
    		t.Errorf("wordLimit(3) on 3 words = %q, want the next converter's result %q", got, "[one two three]")
    	}
    	if got := c("one two three four"); got != "<p>document too long</p>" {
    		t.Errorf("wordLimit(3) on 4 words = %q, want %q", got, "<p>document too long</p>")
    	}
    	if calls != 1 {
    		t.Errorf("next converter ran %d times, want 1 (don't call it for long documents)", calls)
    	}
    }

    func TestFullPipeline(t *testing.T) {
    	convert := Chain(markdownToHTML, trimSpace, wordLimit(10))
    	if got, want := convert("\n\n  # Hi\nThere\n\n"), "<h1>Hi</h1>\n<p>There</p>"; got != want {
    		t.Errorf("full pipeline = %q, want %q", got, want)
    	}
    	if got := convert(strings.Repeat("w ", 11)); got != "<p>document too long</p>" {
    		t.Errorf("full pipeline on 11 words = %q, want %q", got, "<p>document too long</p>")
    	}
    }
---

Time to put Doc2Doc together. Over ten chapters you've built every piece of a
Markdown-to-HTML converter; this exercise wires them into one working pipeline:

```text
Markdown ─► trimSpace ─► wordLimit ─► parse ─► Map(toHTML) ─► join ─► HTML
            └──── middleware (ch 7-8) ────┘    └ iterator (ch 9) ┘
```

The editor already contains:

- the **`Block` sum type** from the last lesson: `Heading`, `Paragraph`, `CodeBlock`
  and `List` behind a sealed interface;
- **`parse(doc)`**, a lazy `iter.Seq[Block]` that reads Markdown line by line (read it,
  it uses a closure over local state to collect list items);
- **`Converter`, `Middleware`, `Chain`** and the **`trimSpace`** middleware from
  chapter 8.

## Your task

1. **`toHTML(b Block) string`**: a type switch that renders each kind of block, with
   every piece of text passed through `html.EscapeString` (so `a & b` becomes
   `a &amp; b`). The exact formats are in the comment above the function. Unknown
   blocks, including `nil`, become `<!-- unknown block -->`.
2. **`Map(seq, f)`**: the lazy iterator adapter from chapter 9. It must stop the moment
   `yield` returns false.
3. **`markdownToHTML(doc)`**: glue. Parse, map every block through `toHTML`, collect
   with `slices.Collect`, and join with `"\n"`. One line is enough.
4. **`wordLimit(n)`**: a configurable middleware. Documents with more than `n` words
   (`strings.Fields`) become `<p>document too long</p>` **without** calling the next
   converter. Anything shorter passes through.

When it all works, `main` prints real HTML:

```text
<h1>Doc2Doc</h1>
<p>Converts **all** your documents &amp; more.</p>
<h2>Install</h2>
<pre><code class="language-sh">go install doc2doc@latest</code></pre>
<ul><li>fast</li><li>pure &lt;functions&gt;</li></ul>
<p>document too long</p>
```

(Inline formatting such as `**bold**` stays as it is. A real converter would parse
that too, and it's a nice extension if you want to keep going.)

## What you've learned

Look at how little of Doc2Doc is "program" and how much is small, pure functions
snapped together:

- **First-class functions** let you pass `toHTML` to `Map` and `markdownToHTML` to
  `Chain`, like any other value.
- **Pure functions** make each piece testable with a table of inputs and outputs.
- **Closures and currying** give you configurable pieces like `wordLimit(500)`.
- **Middleware** adds cross-cutting behaviour without touching the converter.
- **Iterators** keep the pipeline lazy: nothing is parsed until someone asks for it.
- **Sum types** model "exactly one of these shapes" and the type switch takes them
  apart.

And you've seen where Go pushes back: no method-level generics on interfaces, no
exhaustive switches, and plain `for` loops that are often clearer than a clever
chain. Knowing when *not* to be functional is part of the skill.

## What's next

Functions that transform data are exactly how algorithms are described. In
[Learn Algorithms](/courses/learn-algorithms) you'll measure how running time grows
with Big O notation, implement the classic sorting algorithms (leaning on the
recursion from chapter 4 for the divide-and-conquer ones), and build stacks, queues
and linked lists from scratch. The habits from this course (small
pure functions, table-driven tests) will make every one of them easier to get right.
