---
title: Channels
quiz:
  - question: |
      What happens when this program runs?

      ```go
      package main

      import "fmt"

      func main() {
      	ch := make(chan string)
      	ch <- "hello"
      	fmt.Println(<-ch)
      }
      ```
    options:
      - text: It prints `hello`
      - text: 'It crashes with `fatal error: all goroutines are asleep - deadlock!`'
        correct: true
      - text: It prints nothing and exits normally
      - text: It doesn't compile
    explanation: |
      `ch` is unbuffered, so a send waits until another goroutine receives.
      But the only goroutine is `main`, which is stuck on the send and can
      never reach the receive. Go detects that nothing can ever make progress
      and stops with a deadlock error.
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	ch := make(chan int, 2)
      	ch <- 1
      	ch <- 2
      	fmt.Println(len(ch), cap(ch), <-ch)
      }
      ```
    options:
      - text: '`2 2 1`'
        correct: true
      - text: '`2 2 2`'
      - text: It deadlocks
      - text: '`0 2 1`'
    explanation: |
      A buffer of 2 lets both sends complete without a receiver. `len` is the
      number of values waiting (2) and `cap` is the buffer size (2). Channels
      are first in, first out, so the receive gets `1`.
exercise:
  starter: |
    package main

    import "fmt"

    func segments(body string) int {
    	return (len(body) + 159) / 160
    }

    // totalSegments measures every body in its own goroutine and adds up
    // the results it receives from a channel.
    func totalSegments(bodies []string) int {
    	counts := make(chan int)
    	for _, b := range bodies {
    		go func() {
    			counts <- segments(b)
    		}()
    	}
    	total := 0
    	// ?
    	return total
    }

    func main() {
    	fmt.Println(totalSegments([]string{"hi", "ok", string(make([]byte, 200))}))
    }
  solution: |
    package main

    import "fmt"

    func segments(body string) int {
    	return (len(body) + 159) / 160
    }

    func totalSegments(bodies []string) int {
    	counts := make(chan int)
    	for _, b := range bodies {
    		go func() {
    			counts <- segments(b)
    		}()
    	}
    	total := 0
    	for range bodies {
    		total += <-counts
    	}
    	return total
    }

    func main() {
    	fmt.Println(totalSegments([]string{"hi", "ok", string(make([]byte, 200))}))
    }
  tests: |
    package main

    import (
    	"runtime"
    	"strings"
    	"testing"
    	"time"
    )

    func TestTotalSegments(t *testing.T) {
    	for _, tc := range []struct {
    		bodies []string
    		want   int
    	}{
    		{[]string{"hi", "ok", strings.Repeat("x", 200)}, 4},
    		{[]string{strings.Repeat("x", 480), strings.Repeat("x", 481)}, 7},
    		{[]string{"one"}, 1},
    		{nil, 0},
    	} {
    		if got := totalSegments(tc.bodies); got != tc.want {
    			t.Errorf("totalSegments(%d bodies) = %d, want %d", len(tc.bodies), got, tc.want)
    		}
    	}
    }

    func TestNoLeakedGoroutines(t *testing.T) {
    	before := runtime.NumGoroutine()
    	bodies := make([]string, 50)
    	totalSegments(bodies)
    	time.Sleep(50 * time.Millisecond)
    	if after := runtime.NumGoroutine(); after > before {
    		t.Errorf("%d goroutines are still blocked after totalSegments returned; receive once per body", after-before)
    	}
    }
---

Goroutines run independently. **Channels** are how they talk to each other: one goroutine sends a value into a channel, and another receives it. Go's motto is:

> Don't communicate by sharing memory; share memory by communicating.

## Creating, sending and receiving

```go
package main

import "fmt"

func deliver(to string, results chan string) {
	results <- "delivered to " + to
}

func main() {
	results := make(chan string)

	go deliver("alice", results)

	msg := <-results
	fmt.Println(msg)
}
```

```text
delivered to alice
```

- `make(chan string)` creates a channel that carries `string` values. Like maps, channels must be made before use.
- `ch <- v` **sends** `v` into the channel. (The arrow points into the channel.)
- `v := <-ch` **receives** a value from the channel. (The arrow points out of it.)

## Channels synchronise

Here's the key property: on an **unbuffered** channel, a send waits until another goroutine receives, and a receive waits until another goroutine sends. They meet in the middle.

So `msg := <-results` above doesn't just get the value. It also **waits** for `deliver` to finish its work. No `WaitGroup` needed.

Collecting several results works the same way. Receive once per goroutine you started:

```go
package main

import (
	"fmt"
	"slices"
)

func segments(body string) int {
	return (len(body) + 159) / 160
}

func main() {
	bodies := []string{"hi", "a much longer message", "ok"}
	counts := make(chan int)

	for _, b := range bodies {
		go func() {
			counts <- segments(b)
		}()
	}

	var got []int
	for range bodies {
		got = append(got, <-counts)
	}
	slices.Sort(got) // results arrive in any order
	fmt.Println(got)
}
```

```text
[1 1 1]
```

## Deadlock

If a goroutine waits on a channel and nothing will ever send or receive on the other side, it's stuck forever. If *every* goroutine is stuck, Go's runtime notices and crashes the program:

```go
func main() {
	ch := make(chan string)
	ch <- "hello" // waits for a receiver that never comes
	fmt.Println(<-ch)
}
```

```text
fatal error: all goroutines are asleep - deadlock!
```

Whenever you use a channel, ask yourself: *who is on the other end?*

## Buffered channels

Give `make` a size and you get a **buffered** channel, with room to hold that many values:

```go
queue := make(chan string, 100)
```

- A send only waits when the buffer is **full**.
- A receive only waits when the buffer is **empty**.

```go
package main

import "fmt"

func main() {
	outbox := make(chan string, 3)
	outbox <- "reminder 1"
	outbox <- "reminder 2"
	fmt.Println(len(outbox), "queued")

	fmt.Println(<-outbox)
	fmt.Println(<-outbox)
}
```

```text
2 queued
reminder 1
reminder 2
```

No second goroutine needed here, because the buffer had room for both sends. Values come out in the order they went in.

Buffered channels are useful as queues and to smooth out bursts of work. But don't add a buffer just to "fix" a deadlock: that usually hides a design problem until the buffer fills up.

## Channel direction

A function parameter can say it only sends (`chan<- string`) or only receives (`<-chan string`) on a channel. The compiler then stops you from using it the wrong way:

```go
func deliver(to string, results chan<- string) { // may only send
	results <- "delivered to " + to
}
```

## Your turn

`totalSegments` already starts one goroutine per message body, and each one sends
its segment count into `counts`. But nobody receives, so the result is always `0`
and every goroutine is stuck forever on its send.

Receive **once per body** (a `for range bodies` loop works well) and add each value
to `total`. **Run** should print `4`.

## Further reading

- [Go by Example: Channels](https://gobyexample.com/channels)
- [Go by Example: Channel Buffering](https://gobyexample.com/channel-buffering)
- [A Tour of Go: Channels](https://go.dev/tour/concurrency/2)
