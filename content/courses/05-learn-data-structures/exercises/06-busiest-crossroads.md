---
title: Busiest Crossroads
difficulty: easy
after: graphs
hints:
  - 'The number of paths touching a zone is its **degree**. Each path is undirected, so it adds 1 to the degree of **both** of its zones. A `map[string]int` holds every degree.'
  - 'Then loop over the map and keep the best zone. Map order is random, so the tie rule must be part of the comparison: take a zone if its degree is bigger, **or** equal and its name is alphabetically smaller.'
exercise:
  starter: |
    package main

    import "fmt"

    // busiestZone returns the zone touched by the most paths, and how many
    // paths touch it. Ties go to the alphabetically first zone. With no
    // paths, it returns "", 0.
    func busiestZone(paths [][2]string) (string, int) {
    	// 1. Count each zone's degree: every path touches two zones.
    	// 2. Pick the zone with the highest degree (ties: smallest name).
    	return "", 0
    }

    func main() {
    	paths := [][2]string{
    		{"Village", "Forest"},
    		{"Village", "Lake"},
    		{"Lake", "Forest"},
    		{"Forest", "Caves"},
    	}
    	fmt.Println(busiestZone(paths)) // want Forest 3
    }
  solution: |
    package main

    import "fmt"

    func busiestZone(paths [][2]string) (string, int) {
    	degree := map[string]int{}
    	for _, p := range paths {
    		degree[p[0]]++
    		degree[p[1]]++
    	}
    	best, bestDeg := "", 0
    	for zone, d := range degree {
    		if d > bestDeg || d == bestDeg && zone < best {
    			best, bestDeg = zone, d
    		}
    	}
    	return best, bestDeg
    }

    func main() {
    	paths := [][2]string{
    		{"Village", "Forest"},
    		{"Village", "Lake"},
    		{"Lake", "Forest"},
    		{"Forest", "Caves"},
    	}
    	fmt.Println(busiestZone(paths))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    func TestBusiestZone(t *testing.T) {
    	tests := []struct {
    		name  string
    		paths [][2]string
    		want  string
    		deg   int
    	}{
    		{"no paths", nil, "", 0},
    		{"one path", [][2]string{{"Village", "Forest"}}, "Forest", 1},
    		{"example", [][2]string{{"Village", "Forest"}, {"Village", "Lake"}, {"Lake", "Forest"}, {"Forest", "Caves"}}, "Forest", 3},
    		{"hub", [][2]string{{"Castle", "Hub"}, {"Hub", "Arena"}, {"Docks", "Hub"}, {"Arena", "Docks"}}, "Hub", 3},
    		{"counts both ends", [][2]string{{"A", "Z"}, {"B", "Z"}, {"Z", "C"}}, "Z", 3},
    		{"triangle tie", [][2]string{{"Mine", "Lake"}, {"Lake", "Keep"}, {"Keep", "Mine"}}, "Keep", 2},
    		{"two islands", [][2]string{{"Isle", "Reef"}, {"Isle", "Cove"}, {"Peak", "Pass"}, {"Pass", "Gate"}}, "Isle", 2},
    	}
    	for _, tt := range tests {
    		// Run each case several times: map iteration order changes between runs.
    		for range 20 {
    			got, deg := busiestZone(tt.paths)
    			if got != tt.want || deg != tt.deg {
    				t.Errorf("%s: busiestZone(%v) = %q, %d, want %q, %d", tt.name, tt.paths, got, deg, tt.want, tt.deg)
    				break
    			}
    		}
    	}
    }

    func TestBusiestZoneLarge(t *testing.T) {
    	var paths [][2]string
    	for i := range 50_000 {
    		paths = append(paths, [2]string{fmt.Sprint("zone", i), fmt.Sprint("zone", i+1)})
    	}
    	// Every zone except zone0 and zone50000 has degree 2; "zone1" is the
    	// alphabetically first of those.
    	if got, deg := busiestZone(paths); got != "zone1" || deg != 2 {
    		t.Errorf("busiestZone(a 50,000-path chain) = %q, %d, want %q, 2", got, deg, "zone1")
    	}
    }
---

The world map is an undirected graph: zones connected by paths. The design team
wants to know which zone is the busiest crossroads, so they can put a fast-travel
shrine there.

Complete `busiestZone(paths)`. Each element of `paths` is one undirected path
between two zones. Return the zone that the **most paths** touch (its degree), and
that number. If several zones tie, return the one whose name comes first
alphabetically. With no paths, return `"", 0`.

## Example

```
Village ── Forest ── Caves
    \        /
     \      /
       Lake
```

```go
paths := [][2]string{{"Village", "Forest"}, {"Village", "Lake"}, {"Lake", "Forest"}, {"Forest", "Caves"}}
busiestZone(paths) // "Forest", 3
```

## Constraints

- No path connects a zone to itself, and no pair of zones is listed twice.
- O(V + E): count the degrees in one pass.
