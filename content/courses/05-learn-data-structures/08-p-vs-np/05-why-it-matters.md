---
title: Why P vs NP Matters
quiz:
  - question: Your raid scheduler times out on large guilds and you suspect the problem is NP-hard. What's usually the **best** first move?
    options:
      - text: Buy faster servers
      - text: Keep optimising the brute-force search until it's fast
      - text: Check whether your real inputs are small or specially structured, and otherwise switch to a heuristic or approximation that's good enough
        correct: true
      - text: Give up, because NP-hard problems can't be solved
    explanation: |
      Faster hardware barely dents exponential growth, and micro-optimising brute force
      won't change its complexity. But real inputs are often small or have structure you
      can exploit, and a good-enough answer found in milliseconds usually beats a perfect
      one found next week.
  - question: What does an approximation algorithm with a guarantee of 1.5 promise for a minimisation problem?
    options:
      - text: It's 1.5 times faster than the exact algorithm
      - text: Its answer is at most 1.5 times the optimal answer, on every input
        correct: true
      - text: It finds the optimum 1.5% of the time
      - text: It runs in O(n^1.5)
    explanation: |
      An approximation ratio bounds how far from optimal the answer can ever be. For
      metric TSP, the Christofides algorithm guarantees a tour at most 1.5 times the
      shortest. A plain heuristic like nearest neighbour gives no such promise.
---

P vs NP might sound like a puzzle for mathematicians. It's actually one of the most
practical ideas in this course, because it tells you when to **stop looking for a
perfect algorithm**.

## Recognise the smell

After a while you start to notice NP-hard problems by their shape: you're choosing a
**subset**, an **ordering** or an **assignment**, and the constraints interact so that
there's no obvious way to decide one piece at a time. In a game studio:

- Balancing matchmaking teams so their total ratings are as equal as possible
  (a cousin of subset sum).
- Packing items into a limited inventory for the best total value (knapsack).
- Scheduling raids so no player is double-booked (graph colouring).
- Planning the fastest route through every event zone (TSP).
- Laying out a level so every room is reachable under a pile of design rules
  (often SAT in disguise).

If your problem looks like one of these, search for it. Someone has probably proven it
NP-hard, and then you know not to waste a week hunting for a fast, exact, general
algorithm.

## What to do instead

**1. Check the size.** Exponential isn't a problem when n is small. A tour of 10 zones
is 9! = 362,880 orders, which brute force handles in milliseconds. Most guilds don't
have 10,000 members.

**2. Exploit structure.** Your inputs may not be the general case. Subset sum is fast
with dynamic programming when the values are small integers. Graph colouring is easy
on trees. TSP on points in a flat world has good approximation schemes. Restrict the
problem to what you actually need.

**3. Use a heuristic.** Greedy algorithms, like nearest neighbour, give a decent answer
fast without guarantees. Local search improves it: take a tour and try swapping pairs
of zones, keeping any swap that helps (2-opt). Real systems often stop after a fixed
time budget and use the best answer so far.

**4. Use an approximation algorithm.** Some have *proven* bounds. For TSP where the
direct route between zones is never longer than a detour (true for our walking times),
the Christofides algorithm always finds a tour at most **1.5 times** the optimum.

**5. Use a solver.** Modern SAT solvers and integer-programming solvers are remarkably
good on real-world instances, despite exponential worst cases. Encoding your problem for
one is often faster than writing your own search.

**6. Change the design.** Sometimes the best fix is in the game, not the code. If the
World Tour event only ever visits 7 zones, brute force is perfect forever.

## Why the answer matters to everyone

Most of the internet's security relies on problems that seem hard to solve but are easy
to check: finding a secret key from a public one, or breaking an encrypted message.
They're in NP, since a guessed key is easy to verify. If P = NP with a practical
algorithm, much of today's cryptography would collapse. On the bright side, scheduling,
logistics, chip design and protein folding would all get dramatically easier.

Most researchers expect P ≠ NP, which means hard problems stay hard, the secrets stay
secret, and engineers keep making smart trade-offs.

## Course wrap-up

Look at what you've built:

- **Trees**: a generic BST with iterators, and a red-black tree that stays O(log n)
  whatever you throw at it.
- **Hashmaps**: hashing, chaining, open addressing, resizing, and how Go's Swiss-table
  map really works.
- **Tries** for username search and autocomplete.
- **Heaps** for leaderboards, top-k and priority queues.
- **Graphs** with BFS, DFS, cycle detection and Dijkstra to navigate the world map.
- And the judgement to recognise when a problem has **no** efficient exact solution.

Picking the right data structure is most of the battle in real software. You now have a
full toolbox. Next up is [Learn Concurrency in Go](/courses/learn-concurrency), where you'll
put all those cores to work.
