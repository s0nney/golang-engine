---
title: NP-Complete and NP-Hard
quiz:
  - question: What would happen if someone found a polynomial-time algorithm for one NP-complete problem?
    options:
      - text: Only that one problem would become easy
      - text: Every problem in NP would become solvable in polynomial time, proving P = NP
        correct: true
      - text: Nothing, because NP-complete problems can't be solved
      - text: All NP-hard problems, including undecidable ones, would become easy
    explanation: |
      Every NP problem reduces to every NP-complete problem in polynomial time. Solve
      one NP-complete problem fast and you can translate any NP problem into it and
      solve that fast too. Undecidable problems like the halting problem are NP-hard but
      not in NP, so they'd stay unsolvable.
  - question: What's the difference between NP-hard and NP-complete?
    options:
      - text: There is none; they're synonyms
      - text: NP-complete problems are NP-hard **and** in NP; NP-hard problems don't have to be in NP
        correct: true
      - text: NP-hard problems are easier
      - text: NP-complete problems have been proven unsolvable
    explanation: |
      NP-hard means "at least as hard as everything in NP". NP-complete adds "and it's
      in NP itself", so answers can be checked quickly. Optimisation problems like
      "find the *shortest* tour" are NP-hard but not in NP, since you can't quickly
      check that a tour is the shortest.
  - question: You want to show that a new "raid scheduling" problem is NP-hard. Which reduction do you need?
    options:
      - text: Translate raid scheduling into a known NP-complete problem
      - text: Translate a known NP-complete problem into raid scheduling
        correct: true
      - text: Translate raid scheduling into a problem in P
      - text: Run it on bigger and bigger inputs and time it
    explanation: |
      To show your problem is *at least as hard* as a known hard one, show that any
      instance of the known problem can be solved using your problem. Reducing the
      other way only shows your problem is no harder than the known one.
---

Some NP problems are special: they're the hardest ones in the whole class. Understanding
why starts with an idea you use all the time without naming it: **reductions**.

## Reductions

A **reduction** from problem A to problem B is a fast (polynomial-time) way to
translate any instance of A into an instance of B with the same answer. If you have a
fast solver for B, you get a fast solver for A for free: translate, then solve.

You've already done this. "Can a player walk from zone X to zone Y?" reduces to "does
BFS from X reach Y?". "Can the quests all be completed?" reduces to "does the
prerequisite graph have no cycle?". In each case, a question about the game became a
question about a graph that we know how to answer.

Reductions also compare difficulty. If A reduces to B, then **B is at least as hard as
A**: any fast algorithm for B would give a fast algorithm for A.

## NP-complete

In 1971, Stephen Cook (and, independently, Leonid Levin) proved something remarkable:
there's a problem in NP that **every** NP problem reduces to. It's called SAT: given a
boolean formula like `(a || !b) && (b || c) && (!a || !c)`, is there an assignment of
true/false to the variables that makes it true?

A problem is **NP-complete** if:

1. It's **in NP** (answers can be checked quickly), and
2. **Every** problem in NP reduces to it.

Soon after, Richard Karp showed 21 famous problems were NP-complete too, and today
there are thousands. Subset sum, graph 3-colouring, Sudoku, scheduling, the decision
version of the travelling-salesman problem, and even questions about puzzle video
games like generalised Tetris, Minesweeper and Candy Crush are all NP-complete.

That gives NP-complete problems an all-or-nothing property:

- If **any one** NP-complete problem has a polynomial-time algorithm, then **all** of
  NP does, and P = NP.
- If **any one** is proven to need more than polynomial time, then **none** of them has
  a polynomial algorithm, and P ≠ NP.

They all stand or fall together.

## NP-hard

A problem is **NP-hard** if every NP problem reduces to it, whether or not it's in NP
itself. So NP-complete = NP-hard ∩ NP. Some NP-hard problems aren't in NP because:

- **They're optimisation problems.** "Find the shortest tour of all zones" is NP-hard.
  But if someone hands you a tour and claims it's the shortest, you can't verify that
  quickly; you'd have to compare it with every other tour.
- **They're even harder.** The halting problem ("does this Go program ever stop?") is
  NP-hard but undecidable: no algorithm solves it at all, fast or slow.

## The map of the classes

If P ≠ NP, which is what nearly everyone believes, the picture looks like this:

```
|<------------------- NP ------------------->|
                          |<--------------------- NP-hard --------------------->|
+-------------+-----------+------------------+-----------------------------------+
| P           | other NP  | NP-complete      | NP-hard but not in NP             |
| sorting     | problems  | SAT, subset sum, | shortest tour (optimisation),     |
| BFS, DFS    |           | 3-colouring,     | halting problem (undecidable)     |
| Dijkstra    |           | Sudoku           |                                   |
+-------------+-----------+------------------+-----------------------------------+
```

NP-complete is exactly where NP and NP-hard overlap. If P = NP, the first three boxes
collapse into one: everything checkable quickly would be solvable quickly.

## Proving a problem is NP-hard

Say you're asked to write a raid scheduler that fits every raid into the weekend with
no player double-booked, and it keeps timing out. To show it's NP-hard, reduce a
**known** NP-complete problem, such as graph colouring, **to** your problem. Turn any
graph into a raid schedule: each vertex becomes a raid, each edge becomes a player
signed up for both of its raids, and each colour becomes a time slot. A valid schedule
is then exactly a valid colouring, so if you could schedule raids fast, you could colour
graphs fast. The direction matters:
reducing *your* problem to a hard one proves nothing about your problem.
