---
title: 'Practice: Rendering Blocks'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Block is one of: Heading, Paragraph, CodeBlock, List.
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

    // toMarkdown renders one block as Markdown:
    //
    //	Heading{2, "Setup"}             -> "## Setup"
    //	Paragraph{"Hi."}                -> "Hi."
    //	CodeBlock{"go", "x := 1"}       -> "```go\nx := 1\n```"
    //	List{[]string{"a", "b"}}        -> "- a\n- b"
    //
    // Any other value, including a nil Block, returns an error.
    func toMarkdown(b Block) (string, error) {
    	// ?
    	return "", nil
    }

    // wordCount counts the words in headings, paragraphs and list items.
    // Code blocks don't count: code isn't prose.
    func wordCount(blocks []Block) int {
    	// ?
    	return 0
    }

    func main() {
    	doc := []Block{
    		Heading{Level: 1, Text: "Doc2Doc"},
    		Paragraph{Text: "Converts your documents."},
    		CodeBlock{Lang: "sh", Code: "doc2doc report.md"},
    		List{Items: []string{"Markdown in", "HTML out"}},
    	}
    	var parts []string
    	for _, b := range doc {
    		md, err := toMarkdown(b)
    		if err != nil {
    			fmt.Println("error:", err)
    			continue
    		}
    		parts = append(parts, md)
    	}
    	fmt.Println(strings.Join(parts, "\n\n"))
    	fmt.Println("\nwords:", wordCount(doc)) // 8
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // Block is one of: Heading, Paragraph, CodeBlock, List.
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

    // toMarkdown renders one block as Markdown:
    //
    //	Heading{2, "Setup"}             -> "## Setup"
    //	Paragraph{"Hi."}                -> "Hi."
    //	CodeBlock{"go", "x := 1"}       -> "```go\nx := 1\n```"
    //	List{[]string{"a", "b"}}        -> "- a\n- b"
    //
    // Any other value, including a nil Block, returns an error.
    func toMarkdown(b Block) (string, error) {
    	switch b := b.(type) {
    	case Heading:
    		return strings.Repeat("#", b.Level) + " " + b.Text, nil
    	case Paragraph:
    		return b.Text, nil
    	case CodeBlock:
    		return "```" + b.Lang + "\n" + b.Code + "\n```", nil
    	case List:
    		lines := make([]string, len(b.Items))
    		for i, item := range b.Items {
    			lines[i] = "- " + item
    		}
    		return strings.Join(lines, "\n"), nil
    	default:
    		return "", fmt.Errorf("toMarkdown: unknown block %T", b)
    	}
    }

    // wordCount counts the words in headings, paragraphs and list items.
    // Code blocks don't count: code isn't prose.
    func wordCount(blocks []Block) int {
    	total := 0
    	for _, b := range blocks {
    		switch b := b.(type) {
    		case Heading:
    			total += len(strings.Fields(b.Text))
    		case Paragraph:
    			total += len(strings.Fields(b.Text))
    		case List:
    			for _, item := range b.Items {
    				total += len(strings.Fields(item))
    			}
    		}
    	}
    	return total
    }

    func main() {
    	doc := []Block{
    		Heading{Level: 1, Text: "Doc2Doc"},
    		Paragraph{Text: "Converts your documents."},
    		CodeBlock{Lang: "sh", Code: "doc2doc report.md"},
    		List{Items: []string{"Markdown in", "HTML out"}},
    	}
    	var parts []string
    	for _, b := range doc {
    		md, err := toMarkdown(b)
    		if err != nil {
    			fmt.Println("error:", err)
    			continue
    		}
    		parts = append(parts, md)
    	}
    	fmt.Println(strings.Join(parts, "\n\n"))
    	fmt.Println("\nwords:", wordCount(doc)) // 8
    }
  tests: |
    package main

    import "testing"

    type Image struct{ URL string }

    func (Image) isBlock() {}

    func TestToMarkdown(t *testing.T) {
    	for _, tt := range []struct {
    		b    Block
    		want string
    	}{
    		{Heading{Level: 1, Text: "Doc2Doc"}, "# Doc2Doc"},
    		{Heading{Level: 3, Text: "Linux"}, "### Linux"},
    		{Paragraph{Text: "Plain text."}, "Plain text."},
    		{CodeBlock{Lang: "go", Code: "x := 1"}, "```go\nx := 1\n```"},
    		{List{Items: []string{"a", "b", "c"}}, "- a\n- b\n- c"},
    		{List{}, ""},
    	} {
    		got, err := toMarkdown(tt.b)
    		if err != nil {
    			t.Errorf("toMarkdown(%#v) returned error %v, want %q", tt.b, err, tt.want)
    			continue
    		}
    		if got != tt.want {
    			t.Errorf("toMarkdown(%#v) = %q, want %q", tt.b, got, tt.want)
    		}
    	}
    }

    func TestToMarkdownUnknown(t *testing.T) {
    	if _, err := toMarkdown(Image{URL: "cat.png"}); err == nil {
    		t.Errorf("toMarkdown(Image{...}) returned no error; unknown block types need a default case that returns an error")
    	}
    	var nothing Block
    	if _, err := toMarkdown(nothing); err == nil {
    		t.Errorf("toMarkdown(nil) returned no error; a nil Block should be an error too")
    	}
    }

    func TestWordCount(t *testing.T) {
    	doc := []Block{
    		Heading{Level: 1, Text: "Doc2Doc Guide"},
    		Paragraph{Text: "It converts documents quickly."},
    		CodeBlock{Lang: "sh", Code: "doc2doc --all files here"},
    		List{Items: []string{"fast", "pure functions"}},
    		Image{URL: "logo.png"},
    	}
    	if got := wordCount(doc); got != 9 {
    		t.Errorf("wordCount = %d, want 9 (2 heading + 4 paragraph + 3 list; code and images don't count)", got)
    	}
    	if got := wordCount(nil); got != 0 {
    		t.Errorf("wordCount(nil) = %d, want 0", got)
    	}
    }
---

Doc2Doc's parser produces a slice of `Block`s, a sealed interface with four
alternatives: `Heading`, `Paragraph`, `CodeBlock` and `List`. Now write the operations
that take them apart.

## Your task

1. `toMarkdown(b)` uses a **type switch** to render one block:
   - `Heading{Level: 2, Text: "Setup"}` becomes `## Setup`
   - `Paragraph` becomes its text
   - `CodeBlock{Lang: "go", Code: "x := 1"}` becomes a fenced block: three backticks
     and the language, a newline, the code, a newline, three backticks
   - `List{Items: []string{"a", "b"}}` becomes `- a` and `- b` on separate lines

   Anything else, including a nil `Block`, must return an **error** that names the type
   (`%T`). The tests sneak in an `Image` type your switch has never heard of.
2. `wordCount(blocks)` counts the words in headings, paragraphs and list items. Code
   blocks and unknown types count as zero.

## Tips

- In `switch b := b.(type)`, each single-type case gives you `b` as the concrete type,
  so `b.Level` and `b.Items` just work.
- A nil interface matches no concrete case and lands in `default`.
- Go won't warn you about missing cases, so the `default` branch is your safety net.
