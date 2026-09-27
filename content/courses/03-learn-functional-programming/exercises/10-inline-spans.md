---
title: Inline Spans
difficulty: medium
after: sum-types
hints:
  - 'Use a type switch: `switch s := span.(type) { case Text: ... case Bold: ... default: ... }`. A `nil` span lands in `default` too, and `%T` of a nil interface prints `<nil>`.'
  - 'Write a helper `renderChildren(children []Span) (string, error)` that calls `render` on each child and joins the results. `Bold`, `Italic` and `Link` all use it, which is where the recursion happens.'
  - 'If any child returns an error, return `"", err` straight away so errors from deep inside the tree reach the caller. Escape text with `html.EscapeString`.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"html"
    )

    // Span is one of: Text, Code, Bold, Italic, Link.
    type Span interface{ isSpan() }

    type Text struct{ S string }
    type Code struct{ S string }
    type Bold struct{ Children []Span }
    type Italic struct{ Children []Span }
    type Link struct {
    	Href     string
    	Children []Span
    }

    func (Text) isSpan()   {}
    func (Code) isSpan()   {}
    func (Bold) isSpan()   {}
    func (Italic) isSpan() {}
    func (Link) isSpan()   {}

    func render(span Span) (string, error) {
    	_ = html.EscapeString
    	return "", nil
    }

    func main() {
    	span := Bold{[]Span{
    		Text{"Run "},
    		Code{"doc2doc -o out.html"},
    		Text{" & see the "},
    		Link{"https://doc2doc.dev/docs", []Span{Italic{[]Span{Text{"docs"}}}}},
    	}}
    	fmt.Println(render(span))
    	// want: <strong>Run <code>doc2doc -o out.html</code> &amp; see the <a href="https://doc2doc.dev/docs"><em>docs</em></a></strong> <nil>
    }
  solution: |
    package main

    import (
    	"fmt"
    	"html"
    	"strings"
    )

    type Span interface{ isSpan() }

    type Text struct{ S string }
    type Code struct{ S string }
    type Bold struct{ Children []Span }
    type Italic struct{ Children []Span }
    type Link struct {
    	Href     string
    	Children []Span
    }

    func (Text) isSpan()   {}
    func (Code) isSpan()   {}
    func (Bold) isSpan()   {}
    func (Italic) isSpan() {}
    func (Link) isSpan()   {}

    func render(span Span) (string, error) {
    	switch s := span.(type) {
    	case Text:
    		return html.EscapeString(s.S), nil
    	case Code:
    		return "<code>" + html.EscapeString(s.S) + "</code>", nil
    	case Bold:
    		return wrap("<strong>", s.Children, "</strong>")
    	case Italic:
    		return wrap("<em>", s.Children, "</em>")
    	case Link:
    		return wrap(`<a href="`+html.EscapeString(s.Href)+`">`, s.Children, "</a>")
    	default:
    		return "", fmt.Errorf("unknown span type %T", span)
    	}
    }

    func wrap(open string, children []Span, end string) (string, error) {
    	var b strings.Builder
    	b.WriteString(open)
    	for _, child := range children {
    		out, err := render(child)
    		if err != nil {
    			return "", err
    		}
    		b.WriteString(out)
    	}
    	b.WriteString(end)
    	return b.String(), nil
    }

    func main() {
    	span := Bold{[]Span{
    		Text{"Run "},
    		Code{"doc2doc -o out.html"},
    		Text{" & see the "},
    		Link{"https://doc2doc.dev/docs", []Span{Italic{[]Span{Text{"docs"}}}}},
    	}}
    	fmt.Println(render(span))
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    )

    // Strike is a span type that render doesn't know about.
    type Strike struct{ Children []Span }

    func (Strike) isSpan() {}

    func TestRender(t *testing.T) {
    	tests := []struct {
    		name string
    		span Span
    		want string
    	}{
    		{"text", Text{"hello"}, "hello"},
    		{"empty text", Text{""}, ""},
    		{"escaped text", Text{`a < b & "c"`}, "a &lt; b &amp; &#34;c&#34;"},
    		{"code", Code{"x := <-ch"}, "<code>x := &lt;-ch</code>"},
    		{"bold", Bold{[]Span{Text{"hi"}}}, "<strong>hi</strong>"},
    		{"italic", Italic{[]Span{Text{"hi"}}}, "<em>hi</em>"},
    		{"empty bold", Bold{}, "<strong></strong>"},
    		{"link", Link{"/a?x=1&y=2", []Span{Text{"here"}}}, `<a href="/a?x=1&amp;y=2">here</a>`},
    		{"several children", Italic{[]Span{Text{"a"}, Code{"b"}, Text{"c"}}}, "<em>a<code>b</code>c</em>"},
    		{"nested", Bold{[]Span{Italic{[]Span{Link{"/x", []Span{Code{"deep"}}}}}}}, `<strong><em><a href="/x"><code>deep</code></a></em></strong>`},
    		{
    			"example",
    			Bold{[]Span{
    				Text{"Run "},
    				Code{"doc2doc -o out.html"},
    				Text{" & see the "},
    				Link{"https://doc2doc.dev/docs", []Span{Italic{[]Span{Text{"docs"}}}}},
    			}},
    			`<strong>Run <code>doc2doc -o out.html</code> &amp; see the <a href="https://doc2doc.dev/docs"><em>docs</em></a></strong>`,
    		},
    	}
    	for _, tt := range tests {
    		got, err := render(tt.span)
    		if got != tt.want || err != nil {
    			t.Errorf("%s: render(%#v) = %q, %v, want %q, nil", tt.name, tt.span, got, err, tt.want)
    		}
    	}
    }

    func TestRenderDeep(t *testing.T) {
    	var span Span = Text{"core"}
    	for range 200 {
    		span = Bold{[]Span{span}}
    	}
    	got, err := render(span)
    	want := strings.Repeat("<strong>", 200) + "core" + strings.Repeat("</strong>", 200)
    	if got != want || err != nil {
    		t.Errorf("render(200 nested Bolds around \"core\") returned %d bytes, %v, want %d bytes, nil", len(got), err, len(want))
    	}
    }

    func TestRenderErrors(t *testing.T) {
    	tests := []struct {
    		name    string
    		span    Span
    		wantErr string
    	}{
    		{"nil span", nil, "unknown span type <nil>"},
    		{"unknown span", Strike{}, "unknown span type main.Strike"},
    		{"nil child", Bold{[]Span{Text{"ok"}, nil}}, "unknown span type <nil>"},
    		{"unknown grandchild", Link{"/", []Span{Italic{[]Span{Strike{}}}}}, "unknown span type main.Strike"},
    	}
    	for _, tt := range tests {
    		got, err := render(tt.span)
    		if err == nil || err.Error() != tt.wantErr || got != "" {
    			t.Errorf("%s: render(%#v) = %q, %v, want \"\" and the error %q", tt.name, tt.span, got, err, tt.wantErr)
    		}
    	}
    }
---

Doc2Doc's parser splits each paragraph into **inline spans**: plain text,
inline code, bold, italic and links. Bold, italic and links contain other spans,
so a span is a small tree, and `Span` is a **sum type**: every span is exactly
one of `Text`, `Code`, `Bold`, `Italic` or `Link`.

Write `render(span)`, which turns a span tree into HTML:

| Span | HTML |
|---|---|
| `Text{S}` | `S`, escaped |
| `Code{S}` | `<code>S</code>`, with `S` escaped |
| `Bold{Children}` | `<strong>…</strong>` around the rendered children |
| `Italic{Children}` | `<em>…</em>` around the rendered children |
| `Link{Href, Children}` | `<a href="Href">…</a>`, with `Href` escaped |

Escape with `html.EscapeString`. Children are rendered in order and joined with
nothing in between.

Any other span, including `nil`, is an error: return `""` and
`fmt.Errorf("unknown span type %T", span)`. An error anywhere inside the tree
makes the whole `render` call return `""` and that error.

## Example

```go
span := Bold{[]Span{
	Text{"Run "},
	Code{"doc2doc -o out.html"},
	Text{" & see the "},
	Link{"https://doc2doc.dev/docs", []Span{Italic{[]Span{Text{"docs"}}}}},
}}
render(span)
// <strong>Run <code>doc2doc -o out.html</code> &amp; see the <a href="https://doc2doc.dev/docs"><em>docs</em></a></strong>, nil

render(Bold{[]Span{Text{"ok"}, nil}})
// "", unknown span type <nil>
```

## Constraints

- Span trees can be a couple of hundred levels deep.
- `render` must not modify the tree.
