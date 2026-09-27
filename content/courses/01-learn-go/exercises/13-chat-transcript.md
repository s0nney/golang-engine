---
title: Chat Transcript
difficulty: hard
after: structs
hints:
  - 'Clean each message first: trim the sender (use `"Unknown"` if nothing is left), split the body with `strings.Split(body, "\n")`, and keep only the lines that aren''t blank after `strings.TrimSpace`. A message with no lines left is skipped entirely.'
  - 'Remember the sender of the last group you wrote. When a kept message has a different sender, start a new group: write a blank line first (unless this is the very first group), then `Name:`. Then write each of its lines as `"  " + line + "\n"`.'
  - 'Building the result with `out += ...` copies the whole transcript on every step, which is O(n²) and takes seconds for 50,000 messages. Write into a `strings.Builder` and call `.String()` once at the end.'
exercise:
  starter: |
    package main

    import "fmt"

    type Message struct {
    	From string
    	Body string
    }

    func renderThread(msgs []Message) string {
    	return ""
    }

    func main() {
    	thread := []Message{
    		{From: "Mia", Body: "hey"},
    		{From: "Mia", Body: "you there?"},
    		{From: "Sam", Body: "yes!\nsorry, was driving"},
    		{From: "Mia", Body: "   "},
    		{From: "Sam", Body: "what's up"},
    	}
    	fmt.Print(renderThread(thread))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    type Message struct {
    	From string
    	Body string
    }

    // bodyLines returns the non-blank lines of body, trimmed.
    func bodyLines(body string) []string {
    	var lines []string
    	for _, line := range strings.Split(body, "\n") {
    		if line = strings.TrimSpace(line); line != "" {
    			lines = append(lines, line)
    		}
    	}
    	return lines
    }

    func renderThread(msgs []Message) string {
    	var b strings.Builder
    	current := ""
    	for _, m := range msgs {
    		lines := bodyLines(m.Body)
    		if len(lines) == 0 {
    			continue
    		}
    		from := strings.TrimSpace(m.From)
    		if from == "" {
    			from = "Unknown"
    		}
    		if b.Len() == 0 || from != current {
    			if b.Len() > 0 {
    				b.WriteString("\n")
    			}
    			b.WriteString(from + ":\n")
    			current = from
    		}
    		for _, line := range lines {
    			b.WriteString("  " + line + "\n")
    		}
    	}
    	return b.String()
    }

    func main() {
    	thread := []Message{
    		{From: "Mia", Body: "hey"},
    		{From: "Mia", Body: "you there?"},
    		{From: "Sam", Body: "yes!\nsorry, was driving"},
    		{From: "Mia", Body: "   "},
    		{From: "Sam", Body: "what's up"},
    	}
    	fmt.Print(renderThread(thread))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"strings"
    	"testing"
    	"time"
    )

    func TestRenderThread(t *testing.T) {
    	tests := []struct {
    		name string
    		msgs []Message
    		want string
    	}{
    		{"example", []Message{
    			{"Mia", "hey"},
    			{"Mia", "you there?"},
    			{"Sam", "yes!\nsorry, was driving"},
    			{"Mia", "   "},
    			{"Sam", "what's up"},
    		}, "Mia:\n  hey\n  you there?\n\nSam:\n  yes!\n  sorry, was driving\n  what's up\n"},
    		{"empty thread", nil, ""},
    		{"only blank messages", []Message{{"Mia", ""}, {"Sam", " \n\t\n"}}, ""},
    		{"one message", []Message{{"Mia", "hi"}}, "Mia:\n  hi\n"},
    		{"alternating", []Message{{"A", "1"}, {"B", "2"}, {"A", "3"}},
    			"A:\n  1\n\nB:\n  2\n\nA:\n  3\n"},
    		{"trimmed sender joins the group", []Message{{"Mia", "a"}, {"  Mia ", "b"}}, "Mia:\n  a\n  b\n"},
    		{"names are case-sensitive", []Message{{"mia", "a"}, {"Mia", "b"}}, "mia:\n  a\n\nMia:\n  b\n"},
    		{"blank sender", []Message{{"", "who is this"}, {"  ", "hello?"}}, "Unknown:\n  who is this\n  hello?\n"},
    		{"trims lines and drops blank ones", []Message{{"Sam", "  first  \n\n   \n\tsecond\n"}}, "Sam:\n  first\n  second\n"},
    		{"windows line endings", []Message{{"Sam", "one\r\ntwo\r\n"}}, "Sam:\n  one\n  two\n"},
    		{"unicode", []Message{{"Zoë 🐝", "héllo 👋"}, {"Zoë 🐝", "ça va?"}}, "Zoë 🐝:\n  héllo 👋\n  ça va?\n"},
    		{"leading blank message", []Message{{"Sam", ""}, {"Mia", "hi"}}, "Mia:\n  hi\n"},
    	}
    	for _, tt := range tests {
    		if got := renderThread(tt.msgs); got != tt.want {
    			t.Errorf("%s: renderThread(%q)\n got: %q\nwant: %q", tt.name, tt.msgs, got, tt.want)
    		}
    	}
    }

    func TestRenderThreadLarge(t *testing.T) {
    	msgs := make([]Message, 50_000)
    	for i := range msgs {
    		msgs[i] = Message{fmt.Sprintf("user%d", i/3%7), fmt.Sprintf("message number %d", i)}
    	}
    	done := make(chan string, 1)
    	go func() { done <- renderThread(msgs) }()
    	select {
    	case got := <-done:
    		if !strings.HasPrefix(got, "user0:\n  message number 0\n  message number 1\n  message number 2\n\nuser1:\n") {
    			t.Errorf("renderThread(50,000 messages) starts %q, want it to start with user0's group of 3 messages", got[:min(80, len(got))])
    		}
    		if !strings.HasSuffix(got, "  message number 49999\n") {
    			t.Errorf("renderThread(50,000 messages) should end with the last message, %q", "  message number 49999\n")
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("renderThread(50,000 messages) took over a second: build the text with a strings.Builder instead of += on a string")
    	}
    }
---

Customers can download a conversation from Textio as plain text. Your job is
to turn a thread of messages into that transcript.

Write `renderThread(msgs []Message) string`:

1. **Clean each message.** Trim spaces around `From`; if nothing is left, the
   sender is `"Unknown"`. Split `Body` into lines on `"\n"`, trim spaces
   (including tabs and `\r`) around each line, and drop the blank ones. A
   message with no lines left is **skipped entirely**, as if it had never been
   sent.
2. **Group by sender.** Consecutive kept messages from the same sender share
   one group. A group is the sender's name followed by a colon, then each of
   its lines indented by two spaces. Every line ends with `"\n"`.
3. **Separate groups** with one blank line.

An empty thread (or one where every message is skipped) gives `""`.

## Example

```go
renderThread([]Message{
	{From: "Mia", Body: "hey"},
	{From: "Mia", Body: "you there?"},
	{From: "Sam", Body: "yes!\nsorry, was driving"},
	{From: "Mia", Body: "   "},
	{From: "Sam", Body: "what's up"},
})
```

returns this text:

```
Mia:
  hey
  you there?

Sam:
  yes!
  sorry, was driving
  what's up
```

Mia's blank message is skipped, so both of Sam's messages end up in one group.

## Constraints

- Sender names are compared after trimming and are case-sensitive.
- Threads can have 50,000 messages. One test renders that many with a
  one-second limit, and building the text with `+=` takes several seconds.
