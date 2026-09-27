---
title: Portal Dungeon
difficulty: medium
after: bfs-and-dfs
hints:
  - 'Every move costs exactly 1 step, so this is an unweighted shortest path: **BFS**. The vertices are cells `(row, col)`, and the edges are the four neighbouring floor cells plus, on a portal, its twin.'
  - 'Before the search, scan the grid once and build a map from each portal letter to its two cells, such as `map[byte][]cell` with `type cell struct{ r, c int }`. Then a cell''s twin is whichever of the two positions isn''t the cell itself.'
  - 'Store each cell''s distance in a 2D slice initialised to -1 (meaning "not seen"). Set a neighbour''s distance when you **enqueue** it, not when you dequeue it, so no cell is queued twice. The first time you reach `E`, its distance is the answer.'
exercise:
  starter: |
    package main

    import "fmt"

    func escapeSteps(grid []string) int {
    	return 0
    }

    func main() {
    	dungeon := []string{
    		"S.#a.",
    		"..#.#",
    		"a.#.E",
    	}
    	fmt.Println(escapeSteps(dungeon)) // want 6
    }
  solution: |
    package main

    import "fmt"

    type cell struct{ r, c int }

    func escapeSteps(grid []string) int {
    	portals := map[byte][]cell{}
    	var start cell
    	dist := make([][]int, len(grid))
    	for r, row := range grid {
    		dist[r] = make([]int, len(row))
    		for c := range len(row) {
    			dist[r][c] = -1
    			switch ch := row[c]; {
    			case ch == 'S':
    				start = cell{r, c}
    			case ch >= 'a' && ch <= 'z':
    				portals[ch] = append(portals[ch], cell{r, c})
    			}
    		}
    	}
    	dist[start.r][start.c] = 0
    	queue := []cell{start}
    	for len(queue) > 0 {
    		cur := queue[0]
    		queue = queue[1:]
    		ch := grid[cur.r][cur.c]
    		if ch == 'E' {
    			return dist[cur.r][cur.c]
    		}
    		next := []cell{{cur.r - 1, cur.c}, {cur.r + 1, cur.c}, {cur.r, cur.c - 1}, {cur.r, cur.c + 1}}
    		for _, p := range portals[ch] {
    			if p != cur {
    				next = append(next, p)
    			}
    		}
    		for _, n := range next {
    			if n.r < 0 || n.r >= len(grid) || n.c < 0 || n.c >= len(grid[n.r]) {
    				continue
    			}
    			if grid[n.r][n.c] == '#' || dist[n.r][n.c] != -1 {
    				continue
    			}
    			dist[n.r][n.c] = dist[cur.r][cur.c] + 1
    			queue = append(queue, n)
    		}
    	}
    	return -1
    }

    func main() {
    	dungeon := []string{
    		"S.#a.",
    		"..#.#",
    		"a.#.E",
    	}
    	fmt.Println(escapeSteps(dungeon))
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    	"time"
    )

    func TestEscapeSteps(t *testing.T) {
    	tests := []struct {
    		name string
    		grid []string
    		want int
    	}{
    		{"example", []string{"S.#a.", "..#.#", "a.#.E"}, 6},
    		{"side by side", []string{"SE"}, 1},
    		{"straight corridor", []string{"S....E"}, 5},
    		{"walled in", []string{"S#E"}, -1},
    		{"exit sealed off", []string{"S..", "###", "..E"}, -1},
    		{"around a wall", []string{"S.#", "..#", "..E"}, 4},
    		{"portal not worth it", []string{"Sa.E", "####", "a..."}, 3},
    		{"portal chain", []string{"Sa#b.", "####.", "a.b#E"}, 8},
    		{"portal into a dead end", []string{"Sa#E", "####", "a#.."}, -1},
    		{"two portals, choose well", []string{"S.a#E", "####b", "b.a##"}, 7},
    		{"walk onto a portal without jumping", []string{"S", "a", ".", "E", "#", "a"}, 3},
    	}
    	for _, tt := range tests {
    		if got := escapeSteps(tt.grid); got != tt.want {
    			t.Errorf("%s: escapeSteps(%q) = %d, want %d", tt.name, tt.grid, got, tt.want)
    		}
    	}
    }

    func TestEscapeStepsBigMaze(t *testing.T) {
    	// A 401x401 serpentine maze: long corridors joined at alternating ends.
    	// Walking it takes about 80,000 steps; a portal pair skips most of it.
    	const size = 401
    	rows := make([][]byte, size)
    	for r := range rows {
    		if r%2 == 0 {
    			rows[r] = []byte(strings.Repeat(".", size))
    		} else {
    			rows[r] = []byte(strings.Repeat("#", size))
    			if r%4 == 1 {
    				rows[r][size-1] = '.'
    			} else {
    				rows[r][0] = '.'
    			}
    		}
    	}
    	rows[0][0] = 'S'
    	rows[size-1][size-1] = 'E'
    	grid := make([]string, size)
    	for r := range rows {
    		grid[r] = string(rows[r])
    	}
    	start := time.Now()
    	if got := escapeSteps(grid); got != 80_800 {
    		t.Errorf("escapeSteps(401x401 serpentine maze) = %d, want 80800", got)
    	}
    	rows[0][3] = 'q'
    	rows[size-1][5] = 'q'
    	for r := range rows {
    		grid[r] = string(rows[r])
    	}
    	if got := escapeSteps(grid); got != 3+1+(size-1-5) {
    		t.Errorf("escapeSteps(401x401 maze with a portal) = %d, want %d", got, 3+1+(size-1-5))
    	}
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("two searches of a 401x401 maze took %v: each cell should be queued at most once", d)
    	}
    }
---

Dungeon maps are grids of characters:

- `S` is where the party starts and `E` is the exit (exactly one of each),
- `#` is a wall, and `.` is floor,
- a lowercase letter `a`–`z` is a **portal**. Each letter appears exactly twice.
  Portal cells are floor too.

Each step moves up, down, left or right onto a cell that isn't a wall. Standing
on a portal, the party can also spend **one step** to jump to the other cell with
the same letter. Using a portal is optional.

Write `escapeSteps(grid)`. It returns the fewest steps from `S` to `E`, or `-1` if
the exit can't be reached. All rows have the same length.

## Example

```
S.#a.
..#.#
a.#.E
```

The wall in the middle column cuts the map in two, so the party must use the
portal. Walk down twice to the `a` in the bottom-left corner (2 steps), jump to
the other `a` (1 step), then go down, down and right to `E` (3 steps). So
`escapeSteps` returns **6**.

## Constraints

- Grids up to 401 × 401 (about 160,000 cells).
- BFS visits each cell at most once: O(rows × cols). One test uses a 401 × 401
  serpentine maze whose only path is about 80,000 steps long.
