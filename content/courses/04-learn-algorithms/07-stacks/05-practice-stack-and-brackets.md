---
title: 'Practice: Stack and Brackets'
exercise:
  starter: |
    package main

    import "fmt"

    // Stack is a LIFO stack. The zero value is an empty stack.
    type Stack[T any] struct {
    	items []T
    }

    // Push puts v on top of the stack.
    func (s *Stack[T]) Push(v T) {
    	// ?
    }

    // Pop removes and returns the top item, or the zero value and false
    // if the stack is empty.
    func (s *Stack[T]) Pop() (T, bool) {
    	var zero T
    	// ?
    	return zero, false
    }

    // Peek returns the top item without removing it, or the zero value
    // and false if the stack is empty.
    func (s *Stack[T]) Peek() (T, bool) {
    	var zero T
    	// ?
    	return zero, false
    }

    // Len returns the number of items on the stack.
    func (s *Stack[T]) Len() int {
    	return len(s.items)
    }

    // isBalanced reports whether every (, [ and { in text is closed by the
    // matching bracket in the right order. Other characters are ignored.
    func isBalanced(text string) bool {
    	// ? use a Stack[rune]
    	return true
    }

    func main() {
    	var s Stack[string]
    	s.Push("add ava")
    	s.Push("budget 500")
    	top, ok := s.Pop()
    	fmt.Println(top, ok, s.Len()) // want: budget 500 true 1

    	fmt.Println(isBalanced("Hi {{handle}}! [Shop](https://x.example/{{code}})")) // want: true
    	fmt.Println(isBalanced("([)]"))                                              // want: false
    }
  solution: |
    package main

    import "fmt"

    type Stack[T any] struct {
    	items []T
    }

    func (s *Stack[T]) Push(v T) {
    	s.items = append(s.items, v)
    }

    func (s *Stack[T]) Pop() (T, bool) {
    	var zero T
    	if len(s.items) == 0 {
    		return zero, false
    	}
    	last := len(s.items) - 1
    	v := s.items[last]
    	s.items[last] = zero
    	s.items = s.items[:last]
    	return v, true
    }

    func (s *Stack[T]) Peek() (T, bool) {
    	var zero T
    	if len(s.items) == 0 {
    		return zero, false
    	}
    	return s.items[len(s.items)-1], true
    }

    func (s *Stack[T]) Len() int {
    	return len(s.items)
    }

    var openerFor = map[rune]rune{')': '(', ']': '[', '}': '{'}

    func isBalanced(text string) bool {
    	var stack Stack[rune]
    	for _, r := range text {
    		switch r {
    		case '(', '[', '{':
    			stack.Push(r)
    		case ')', ']', '}':
    			top, ok := stack.Pop()
    			if !ok || top != openerFor[r] {
    				return false
    			}
    		}
    	}
    	return stack.Len() == 0
    }

    func main() {
    	var s Stack[string]
    	s.Push("add ava")
    	s.Push("budget 500")
    	top, ok := s.Pop()
    	fmt.Println(top, ok, s.Len())

    	fmt.Println(isBalanced("Hi {{handle}}! [Shop](https://x.example/{{code}})"))
    	fmt.Println(isBalanced("([)]"))
    }
  tests: |
    package main

    import "testing"

    func TestStack(t *testing.T) {
    	var s Stack[int]
    	if _, ok := s.Pop(); ok {
    		t.Fatal("Pop on an empty stack returned ok = true, want false")
    	}
    	if _, ok := s.Peek(); ok {
    		t.Fatal("Peek on an empty stack returned ok = true, want false")
    	}
    	for i := range 5 {
    		s.Push(i * 10)
    	}
    	if s.Len() != 5 {
    		t.Fatalf("after 5 pushes, Len() = %d, want 5", s.Len())
    	}
    	if v, ok := s.Peek(); v != 40 || !ok {
    		t.Fatalf("Peek() = %d, %v; want 40, true", v, ok)
    	}
    	if s.Len() != 5 {
    		t.Fatalf("Peek changed the length to %d; it must not remove anything", s.Len())
    	}
    	for want := 40; want >= 0; want -= 10 {
    		if v, ok := s.Pop(); v != want || !ok {
    			t.Fatalf("Pop() = %d, %v; want %d, true (last in, first out)", v, ok, want)
    		}
    	}
    	if v, ok := s.Pop(); v != 0 || ok {
    		t.Fatalf("Pop on an emptied stack = %d, %v; want 0, false", v, ok)
    	}
    	s.Push(0)
    	if v, ok := s.Pop(); v != 0 || !ok {
    		t.Fatalf("pushing and popping a real 0 gave %d, %v; want 0, true", v, ok)
    	}
    }

    func TestIsBalanced(t *testing.T) {
    	tests := []struct {
    		text string
    		want bool
    	}{
    		{"", true},
    		{"no brackets at all", true},
    		{"()", true},
    		{"{[()()]}", true},
    		{"()[]{}", true},
    		{"Hi {{handle}}! [Shop](https://x.example/{{code}})", true},
    		{"🎉 (launch) 🎉", true},
    		{"(", false},
    		{")", false},
    		{")(", false},
    		{"((", false},
    		{"([)]", false},
    		{"Big news {{handle}! (link in bio)", false},
    		{"{[}", false},
    	}
    	for _, tt := range tests {
    		if got := isBalanced(tt.text); got != tt.want {
    			t.Errorf("isBalanced(%q) = %v, want %v", tt.text, got, tt.want)
    		}
    	}
    }
---

Time to build the stack yourself, then put it to work checking Clout's post
templates.

## Part 1: Stack[T]

Complete the generic `Stack[T]`, backed by a slice whose **end is the top**:

- `Push(v)` appends `v`.
- `Pop()` removes and returns the top item and `true`, or the zero value and
  `false` when the stack is empty. Clear the vacated slot with the zero value
  before shrinking, so the garbage collector can reclaim what it pointed to.
- `Peek()` returns the top item and `true` **without removing it**, or the zero
  value and `false`.

`Len` is done for you. The zero value `var s Stack[int]` must work with no
constructor.

## Part 2: isBalanced

Complete `isBalanced(text)` using a `Stack[rune]`:

1. Push every opening bracket: `(`, `[`, `{`.
2. On a closing bracket, pop. If the stack was empty, or the popped rune isn't
   the matching opener, return `false`.
3. Ignore every other character, including emoji.
4. At the end, the text is balanced only if the stack is empty.

A map from closer to opener keeps it tidy:

```go
var openerFor = map[rune]rune{')': '(', ']': '[', '}': '{'}
```

The tests cover the stack on its own (LIFO order, empty pops, a pushed `0`
versus an empty stack), then a table of balanced and unbalanced templates,
including `([)]`, `)(` and a missing `}`.
