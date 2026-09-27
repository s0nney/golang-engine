---
title: Global Top
difficulty: medium
after: heaps-and-priority-queues
hints:
  - 'Each board is already sorted, so the best remaining entry overall is always at the **front** of some board. Keep one "cursor" per board (board index + position) in a heap ordered by the entry it points at: higher score first, then smaller name.'
  - 'Pop the best cursor, output its entry, then push the cursor back advanced by one (if that board has entries left). Stop after `k` entries or when the heap is empty. With `container/heap` you need `Len`, `Less`, `Swap`, `Push` and `Pop`; or write a small sift-up / sift-down heap yourself.'
  - 'A player on several boards pops out first with their **best** entry (the ordering guarantees it). Remember output players in a `map[string]bool` and skip any later entry for a player you''ve already output. Skipped entries don''t count towards `k`.'
exercise:
  starter: |
    package main

    import "fmt"

    type Entry struct {
    	Player string
    	Score  int
    }

    func globalTop(boards [][]Entry, k int) []Entry {
    	return nil
    }

    func main() {
    	eu := []Entry{{"mira", 980}, {"kai", 870}, {"lux", 500}}
    	na := []Entry{{"zed", 990}, {"kai", 860}, {"ava", 700}}
    	fmt.Println(globalTop([][]Entry{eu, na}, 4)) // want [{zed 990} {mira 980} {kai 870} {ava 700}]
    }
  solution: |
    package main

    import (
    	"container/heap"
    	"fmt"
    )

    type Entry struct {
    	Player string
    	Score  int
    }

    // better reports whether a ranks above b: higher score, then smaller name.
    func better(a, b Entry) bool {
    	if a.Score != b.Score {
    		return a.Score > b.Score
    	}
    	return a.Player < b.Player
    }

    type cursor struct {
    	board, pos int
    }

    type cursorHeap struct {
    	boards [][]Entry
    	items  []cursor
    }

    func (h *cursorHeap) at(c cursor) Entry { return h.boards[c.board][c.pos] }
    func (h *cursorHeap) Len() int          { return len(h.items) }
    func (h *cursorHeap) Less(i, j int) bool {
    	return better(h.at(h.items[i]), h.at(h.items[j]))
    }
    func (h *cursorHeap) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }
    func (h *cursorHeap) Push(x any)    { h.items = append(h.items, x.(cursor)) }
    func (h *cursorHeap) Pop() any {
    	last := h.items[len(h.items)-1]
    	h.items = h.items[:len(h.items)-1]
    	return last
    }

    func globalTop(boards [][]Entry, k int) []Entry {
    	h := &cursorHeap{boards: boards}
    	for b, board := range boards {
    		if len(board) > 0 {
    			h.items = append(h.items, cursor{b, 0})
    		}
    	}
    	heap.Init(h)
    	var top []Entry
    	seen := map[string]bool{}
    	for len(top) < k && h.Len() > 0 {
    		c := heap.Pop(h).(cursor)
    		if e := h.at(c); !seen[e.Player] {
    			seen[e.Player] = true
    			top = append(top, e)
    		}
    		if c.pos+1 < len(boards[c.board]) {
    			heap.Push(h, cursor{c.board, c.pos + 1})
    		}
    	}
    	return top
    }

    func main() {
    	eu := []Entry{{"mira", 980}, {"kai", 870}, {"lux", 500}}
    	na := []Entry{{"zed", 990}, {"kai", 860}, {"ava", 700}}
    	fmt.Println(globalTop([][]Entry{eu, na}, 4))
    }
  tests: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"slices"
    	"testing"
    	"time"
    )

    func TestGlobalTop(t *testing.T) {
    	eu := []Entry{{"mira", 980}, {"kai", 870}, {"lux", 500}}
    	na := []Entry{{"zed", 990}, {"kai", 860}, {"ava", 700}}
    	asia := []Entry{{"hana", 870}, {"lux", 600}, {"rin", 100}}
    	tests := []struct {
    		name   string
    		boards [][]Entry
    		k      int
    		want   []Entry
    	}{
    		{"example", [][]Entry{eu, na}, 4, []Entry{{"zed", 990}, {"mira", 980}, {"kai", 870}, {"ava", 700}}},
    		{"k = 1", [][]Entry{eu, na}, 1, []Entry{{"zed", 990}}},
    		{"k = 0", [][]Entry{eu, na}, 0, nil},
    		{"k negative", [][]Entry{eu, na}, -3, nil},
    		{"no boards", nil, 5, nil},
    		{"only empty boards", [][]Entry{{}, nil, {}}, 5, nil},
    		{"one board", [][]Entry{eu}, 2, []Entry{{"mira", 980}, {"kai", 870}}},
    		{"k bigger than everyone", [][]Entry{eu, na}, 99, []Entry{{"zed", 990}, {"mira", 980}, {"kai", 870}, {"ava", 700}, {"lux", 500}}},
    		{
    			"tie on score: name order",
    			[][]Entry{eu, na, asia}, 4,
    			[]Entry{{"zed", 990}, {"mira", 980}, {"hana", 870}, {"kai", 870}},
    		},
    		{
    			"duplicate player keeps best entry only",
    			[][]Entry{eu, na, asia}, 7,
    			[]Entry{{"zed", 990}, {"mira", 980}, {"hana", 870}, {"kai", 870}, {"ava", 700}, {"lux", 600}, {"rin", 100}},
    		},
    		{
    			"empty board among others",
    			[][]Entry{{}, {{"solo", 5}}, nil, {{"duo", 5}, {"trio", -2}}}, 3,
    			[]Entry{{"duo", 5}, {"solo", 5}, {"trio", -2}},
    		},
    	}
    	for _, tt := range tests {
    		got := globalTop(tt.boards, tt.k)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("%s: globalTop(%v, %d) = %v, want %v", tt.name, tt.boards, tt.k, got, tt.want)
    		}
    	}
    	if !slices.Equal(eu, []Entry{{"mira", 980}, {"kai", 870}, {"lux", 500}}) {
    		t.Errorf("globalTop changed a board to %v; it must not modify its input", eu)
    	}
    }

    func TestGlobalTopManyBoards(t *testing.T) {
    	// 2,000 regional boards of 100 entries each. Player p%50,000 appears
    	// on several boards with different scores.
    	var boards [][]Entry
    	var all []Entry
    	for b := range 2_000 {
    		board := make([]Entry, 100)
    		for i := range board {
    			p := (b*100 + i) * 7919 % 50_000
    			board[i] = Entry{fmt.Sprintf("p%05d", p), (b*131 + i*977) % 10_000}
    		}
    		slices.SortFunc(board, func(a, b Entry) int {
    			return cmp.Or(cmp.Compare(b.Score, a.Score), cmp.Compare(a.Player, b.Player))
    		})
    		boards = append(boards, board)
    		all = append(all, board...)
    	}
    	// Reference answer: sort everything, keep each player's first (best) entry.
    	slices.SortFunc(all, func(a, b Entry) int {
    		return cmp.Or(cmp.Compare(b.Score, a.Score), cmp.Compare(a.Player, b.Player))
    	})
    	seen := map[string]bool{}
    	var want []Entry
    	for _, e := range all {
    		if !seen[e.Player] && len(want) < 500 {
    			seen[e.Player] = true
    			want = append(want, e)
    		}
    	}
    	start := time.Now()
    	got := globalTop(boards, 500)
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("globalTop(2,000 boards, 500) took %v", d)
    	}
    	if !slices.Equal(got, want) {
    		i := 0
    		for i < min(len(got), len(want)) && got[i] == want[i] {
    			i++
    		}
    		t.Errorf("globalTop(2,000 boards, 500): got %d entries, first difference at position %d", len(got), i)
    	}
    }
---

Every region has its own leaderboard. For the season finale, the studio wants a
**global top k**: the best `k` players across all regions.

Write `globalTop(boards, k)`. Each board is sorted best first: **higher score
first**, and equal scores in **alphabetical order** of player name. Return the
best `k` entries across all boards, in that same order.

Some players compete in several regions, so they appear on more than one board.
A player appears **at most once** in the result, with their best entry.

- If `k` is 0 or negative, return an empty (or nil) slice.
- If there are fewer than `k` distinct players, return them all.
- Boards may be empty or nil. Don't modify the boards.

## Example

```go
eu := []Entry{{"mira", 980}, {"kai", 870}, {"lux", 500}}
na := []Entry{{"zed", 990}, {"kai", 860}, {"ava", 700}}
globalTop([][]Entry{eu, na}, 4)
// [{zed 990} {mira 980} {kai 870} {ava 700}]
```

`kai` is on both boards; only his best entry (870) counts, and his 860 on `na`
is skipped.

## Constraints

- Up to 2,000 boards and 200,000 entries in total.
- The boards are already sorted, so you don't need to sort everything. A heap of
  one cursor per board finds each next entry in O(log b), where b is the number
  of boards. (This is a *k-way merge*.)
