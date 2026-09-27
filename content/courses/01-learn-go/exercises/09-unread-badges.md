---
title: Unread Badges
difficulty: medium
after: structs
hints:
  - 'Skip read messages straight away with `continue`. For the rest, work out the sender name, using `strings.TrimSpace` to spot blank names and replacing them with `"unknown"`.'
  - 'You need the result in first-appearance order, and a map''s order is random. So build the result slice as you go, and use a map from sender name to that sender''s **index** in the slice.'
  - 'If the sender is already in the index map, do `result[i].Unread++`. Otherwise append `SenderCount{From: name, Unread: 1}` and store its index, `len(result)-1`.'
exercise:
  starter: |
    package main

    import "fmt"

    type Message struct {
    	From string
    	Body string
    	Read bool
    }

    type SenderCount struct {
    	From   string
    	Unread int
    }

    func unreadBySender(inbox []Message) []SenderCount {
    	return nil
    }

    func main() {
    	inbox := []Message{
    		{From: "Mia", Body: "lunch?", Read: true},
    		{From: "Sam", Body: "call me"},
    		{From: "Mia", Body: "hello??"},
    		{From: "Sam", Body: "urgent"},
    		{From: "", Body: "you won a prize"},
    	}
    	fmt.Println(unreadBySender(inbox))
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
    	Read bool
    }

    type SenderCount struct {
    	From   string
    	Unread int
    }

    func unreadBySender(inbox []Message) []SenderCount {
    	var result []SenderCount
    	index := map[string]int{}
    	for _, m := range inbox {
    		if m.Read {
    			continue
    		}
    		name := m.From
    		if strings.TrimSpace(name) == "" {
    			name = "unknown"
    		}
    		if i, ok := index[name]; ok {
    			result[i].Unread++
    			continue
    		}
    		result = append(result, SenderCount{From: name, Unread: 1})
    		index[name] = len(result) - 1
    	}
    	return result
    }

    func main() {
    	inbox := []Message{
    		{From: "Mia", Body: "lunch?", Read: true},
    		{From: "Sam", Body: "call me"},
    		{From: "Mia", Body: "hello??"},
    		{From: "Sam", Body: "urgent"},
    		{From: "", Body: "you won a prize"},
    	}
    	fmt.Println(unreadBySender(inbox))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestUnreadBySender(t *testing.T) {
    	tests := []struct {
    		name  string
    		inbox []Message
    		want  []SenderCount
    	}{
    		{"example", []Message{
    			{From: "Mia", Body: "lunch?", Read: true},
    			{From: "Sam", Body: "call me"},
    			{From: "Mia", Body: "hello??"},
    			{From: "Sam", Body: "urgent"},
    			{From: "", Body: "you won a prize"},
    		}, []SenderCount{{"Sam", 2}, {"Mia", 1}, {"unknown", 1}}},
    		{"empty inbox", nil, nil},
    		{"everything read", []Message{{From: "Mia", Read: true}, {From: "Sam", Read: true}}, nil},
    		{"one sender", []Message{{From: "Ana"}, {From: "Ana"}, {From: "Ana"}}, []SenderCount{{"Ana", 3}}},
    		{"blank names are unknown", []Message{{From: "  "}, {From: ""}, {From: "\t"}}, []SenderCount{{"unknown", 3}}},
    		{"names are case-sensitive", []Message{{From: "sam"}, {From: "Sam"}}, []SenderCount{{"sam", 1}, {"Sam", 1}}},
    		{"order of first unread", []Message{
    			{From: "Zed", Read: true},
    			{From: "Amy"},
    			{From: "Zed"},
    			{From: "Amy"},
    		}, []SenderCount{{"Amy", 2}, {"Zed", 1}}},
    		{"unicode names", []Message{{From: "Zoë 🐝"}, {From: "Zoë 🐝"}}, []SenderCount{{"Zoë 🐝", 2}}},
    	}
    	for _, tt := range tests {
    		got := unreadBySender(tt.inbox)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("%s: unreadBySender(%+v) = %+v, want %+v", tt.name, tt.inbox, got, tt.want)
    		}
    	}
    }
---

Textio's inbox shows a badge next to each contact with how many of their
messages you haven't read yet.

Write `unreadBySender(inbox []Message) []SenderCount`. It returns one
`SenderCount` for every sender with **at least one unread** message, holding
the sender's name and how many of their messages are unread.

- Results are in the order in which each sender's **first unread** message
  appears in `inbox`.
- Messages whose `From` is empty or only whitespace are counted under the
  name `"unknown"`.
- Names are compared exactly: `"sam"` and `"Sam"` are different senders.
- If nothing is unread, return an empty (or `nil`) slice.

## Example

```go
inbox := []Message{
	{From: "Mia", Body: "lunch?", Read: true},
	{From: "Sam", Body: "call me"},
	{From: "Mia", Body: "hello??"},
	{From: "Sam", Body: "urgent"},
	{From: "", Body: "you won a prize"},
}
unreadBySender(inbox)
// [{Sam 2} {Mia 1} {unknown 1}]
```

Mia's first message was already read, so her badge is ordered by her
second message, which comes after Sam's first.

## Constraints

- Keep the `Message` and `SenderCount` types as they are; the tests use them.
