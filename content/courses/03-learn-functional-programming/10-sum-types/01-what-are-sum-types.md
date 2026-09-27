---
title: What Are Sum Types?
quiz:
  - question: A Doc2Doc `Block` is either a heading, a paragraph or a code block, and never two at once. What kind of type is that?
    options:
      - text: A product type
      - text: A pointer type
      - text: A sum type
        correct: true
      - text: A generic type
    explanation: |
      A sum type is "one of" several alternatives. A product type (a struct) is
      "all of" its fields at once.
  - question: |
      How many valid values does this struct allow for `Kind` and `Level` together, and how many make sense?

      ```go
      type Block struct {
          Kind  string // "heading" or "paragraph"
          Level int    // only meaningful for headings, 1 to 6
      }
      ```
    options:
      - text: The type allows exactly the 7 meaningful combinations
      - text: The compiler rejects paragraphs with a level
      - text: 'The type allows endlessly many combinations, such as `{Kind: "paragraph", Level: 4}` or `{Kind: "banana"}`, but only 7 make sense'
        correct: true
    explanation: |
      A struct allows every combination of its fields, including nonsense. Six
      heading levels plus one paragraph make 7 sensible values. A sum type lets you
      describe exactly those, so invalid states can't be built.
---

You've spent the whole course passing functions around. This last chapter is about
**data**, specifically a kind of type that functional languages love and Go doesn't
have built in: the **sum type**.

## Product types: "all of"

A struct is a **product type**. A value holds *all* of its fields at once:

```go
type Doc struct {
	Title string
	Words int
}
```

Every `Doc` has a title **and** a word count. The possible values are every title
combined with every count, which is where the name "product" comes from.

## Sum types: "one of"

A **sum type** holds exactly *one* of several alternatives, and each alternative
can carry its own data. When Doc2Doc parses Markdown, each block of the document is
one of:

- a **Heading** with a level and some text,
- a **Paragraph** with some text,
- a **CodeBlock** with a language and some code.

A block is never a heading *and* a code block. In a language with sum types, such as
Rust, Haskell, Swift or OCaml, you'd write something like:

```text
type Block =
  | Heading   of level: int * text: string
  | Paragraph of text: string
  | CodeBlock of lang: string * code: string
```

And then the compiler forces every `match` on a `Block` to handle all three cases.
Add a fourth, like `Image`, and it points you at every place you forgot to update.

## The struct-with-a-kind-field approach

Without sum types, a common first attempt is one struct with every possible field
and a "kind" tag:

```go
type Block struct {
	Kind  string // "heading", "paragraph" or "code"
	Level int    // headings only
	Text  string // headings and paragraphs
	Lang  string // code only
	Code  string // code only
}
```

It works, but it's leaky:

- **Invalid states are representable.** `Block{Kind: "paragraph", Level: 3, Lang: "go"}`
  compiles happily. So does `Block{Kind: "hedaing"}`.
- **The fields don't say which ones matter.** You have to remember that `Lang` is
  only meaningful for code blocks.
- **Nothing checks your `switch b.Kind`.** Forget a case and you find out at runtime.

Functional programmers have a slogan for this: **make illegal states
unrepresentable**. Design your types so that nonsense values can't even be
constructed, and a whole category of bugs disappears.

## What Go offers

Go has no `|` for types. Instead, the next lessons show the idiomatic emulation:

1. A **sealed interface**: one interface, with one struct per alternative.
2. A **type switch** to handle each alternative.
3. **Generic** `Option` and `Result` types, and why Go usually prefers `(T, bool)` and
   `(T, error)` instead.

You'll also see honestly where the emulation falls short, chiefly that the compiler
can't check that you've handled every case.
