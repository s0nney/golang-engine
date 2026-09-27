---
title: What Is an Algorithm?
quiz:
  - question: Which of these best describes an algorithm?
    options:
      - text: A program written in a fast language such as Go
      - text: A finite, step-by-step procedure that turns an input into an output
        correct: true
      - text: Any function with a loop in it
      - text: A mathematical formula with no steps
    explanation: |
      An algorithm is a recipe: a finite list of unambiguous steps that turns an
      input into an output. It exists independently of any language. Go is just
      one way to write it down, and plenty of algorithms have no loops at all.
  - question: |
      Clout's `totalFollowers` function takes a slice of follower counts. What are
      its input and output?
    options:
      - text: The input is the total, the output is the slice
      - text: There is no input, because the slice is a global
      - text: The input is the slice of counts, the output is their sum
        correct: true
    explanation: |
      Every algorithm maps an input to an output. Here the input is the slice of
      follower counts and the output is a single number, their sum.
---

Welcome to your new job! You've just joined **Clout**, a scrappy startup that
crunches social-media data for influencers and the brands that pay them. Your
first week's tickets sound simple: "find the smallest account", "rank these
creators by followers", "keep a history of undo actions". Every one of them is
an *algorithm* problem, and by the end of this course you'll know how to solve
them and, more importantly, how to tell whether your solution will survive
when Clout goes from 100 influencers to 100 million.

## So what is an algorithm?

An **algorithm** is a finite sequence of well-defined steps that takes some
input and produces some output. That's it. A cooking recipe is an algorithm:

1. Input: eggs, butter, salt.
2. Crack the eggs into a bowl.
3. Whisk.
4. Melt the butter in a pan, add the eggs, stir until set.
5. Output: scrambled eggs.

For something to count as an algorithm, each step has to be unambiguous
("add some salt" is too vague for a computer), and it must eventually stop.

## Algorithms vs programs

An algorithm is an *idea*. A program is that idea written in a specific
language. The algorithm "add every number in a list to a running total" is the
same whether you write it in Go, Python or on a napkin. Here it is in Go:

```go
package main

import "fmt"

func totalFollowers(counts []int) int {
	total := 0
	for _, c := range counts {
		total += c
	}
	return total
}

func main() {
	counts := []int{1200, 56000, 830, 4400}
	fmt.Println(totalFollowers(counts))
}
```

This prints `62430`. The input is a slice of follower counts, the output is
their sum, and the steps are: start at zero, add each count, return the result.

## Why study them?

You already know how to write Go. So why a whole course on algorithms? Because
two programs that give the *same answer* can take wildly different amounts of
time and memory to get there. At Clout's scale, one approach might answer in a
millisecond while another takes longer than the heat death of the universe.
We'll learn to:

- **Design** step-by-step solutions to problems.
- **Analyse** how their cost grows as the input grows (Big O).
- **Choose** the right *data structure* (slices, stacks, queues, linked lists)
  to hold the data the algorithm works on.

Algorithms and data structures go hand in hand. An algorithm is the *how*, and
the data structure is the *where*.
