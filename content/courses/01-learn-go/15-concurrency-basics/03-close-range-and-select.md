---
title: Close, Range and Select
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	ch := make(chan int, 3)
      	ch <- 1
      	ch <- 2
      	close(ch)

      	for n := range ch {
      		fmt.Print(n, " ")
      	}
      	v, ok := <-ch
      	fmt.Println(v, ok)
      }
      ```
    options:
      - text: '`1 2 0 false`'
        correct: true
      - text: '`1 2` and then it deadlocks'
      - text: It panics, because you can't receive from a closed channel
      - text: '`1 2 2 true`'
    explanation: |
      Values sent before `close` can still be received, so the loop prints
      `1 2` and then ends because the channel is closed and empty. Receiving
      from a closed, empty channel returns the zero value immediately, with
      `ok` set to `false`.
  - question: Who should close a channel?
    options:
      - text: The receiver, once it has read everything
      - text: The sender, when it has nothing more to send
        correct: true
      - text: Every goroutine that uses it
      - text: Nobody; channels must always be closed by the runtime
    explanation: |
      Closing means "no more values are coming", which only the sender knows.
      Sending on a closed channel panics, so a receiver that closes a channel
      can crash a sender that's still running.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"time"
    )

    // queueReminders sends "Reminder for <name>" into out for every name,
    // in order, and then tells the receiver there's nothing more to come.
    func queueReminders(names []string, out chan<- string) {
    	for _, n := range names {
    		out <- "Reminder for " + n
    	}
    	// ?
    }

    func main() {
    	reminders := make(chan string)
    	go queueReminders([]string{"alice", "bob"}, reminders)

    	timeout := time.After(time.Second)
    	for {
    		select {
    		case r, ok := <-reminders:
    			if !ok {
    				fmt.Println("queue empty")
    				return
    			}
    			fmt.Println(r)
    		case <-timeout:
    			fmt.Println("still waiting after 1s: did you forget something?")
    			return
    		}
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"time"
    )

    func queueReminders(names []string, out chan<- string) {
    	for _, n := range names {
    		out <- "Reminder for " + n
    	}
    	close(out)
    }

    func main() {
    	reminders := make(chan string)
    	go queueReminders([]string{"alice", "bob"}, reminders)

    	timeout := time.After(time.Second)
    	for {
    		select {
    		case r, ok := <-reminders:
    			if !ok {
    				fmt.Println("queue empty")
    				return
    			}
    			fmt.Println(r)
    		case <-timeout:
    			fmt.Println("still waiting after 1s: did you forget something?")
    			return
    		}
    	}
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"time"
    )

    func TestQueueReminders(t *testing.T) {
    	tests := [][]string{
    		{"alice", "bob", "carol"},
    		{"dave"},
    		{},
    	}
    	for _, names := range tests {
    		out := make(chan string)
    		go queueReminders(names, out)

    		var got []string
    		timeout := time.After(500 * time.Millisecond)
    	receive:
    		for {
    			select {
    			case r, ok := <-out:
    				if !ok {
    					break receive
    				}
    				got = append(got, r)
    			case <-timeout:
    				t.Fatalf("queueReminders(%q): channel was never closed (received %q so far)", names, got)
    			}
    		}

    		var want []string
    		for _, n := range names {
    			want = append(want, "Reminder for "+n)
    		}
    		if !slices.Equal(got, want) {
    			t.Errorf("queueReminders(%q) sent %q, want %q", names, got, want)
    		}
    	}
    }
---

Receiving exactly N values works when you know N. But what if a goroutine produces an unknown number of messages? The sender needs a way to say "that's everything".

## `close`

The sender calls `close(ch)` to signal that no more values will be sent. The receiver can find out with the comma-ok form of receive:

```go
v, ok := <-ch // ok is false once ch is closed and empty
```

After a channel is closed:

- Values already in the buffer can still be received.
- Once it's empty, receives return **immediately** with the zero value and `ok == false`.
- **Sending** on it **panics**. So only the sender should close a channel, never the receiver.

You don't have to close every channel. Close one only when a receiver needs to know that the values have stopped.

## `range` over a channel

`for v := range ch` receives values until the channel is closed and empty. It's the cleanest way to consume a stream of work:

```go
package main

import "fmt"

func queueReminders(names []string, out chan<- string) {
	for _, n := range names {
		out <- "Reminder for " + n
	}
	close(out) // tell the receiver we're done
}

func main() {
	reminders := make(chan string)
	go queueReminders([]string{"alice", "bob", "carol"}, reminders)

	for r := range reminders {
		fmt.Println(r)
	}
	fmt.Println("queue empty")
}
```

```text
Reminder for alice
Reminder for bob
Reminder for carol
queue empty
```

Forget the `close` and the `range` loop waits forever for a fourth reminder. That's a deadlock.

## `select`: waiting on several channels

`select` is like a `switch` for channel operations. It waits until **one** of its cases can proceed, then runs that case. If several are ready at once, it picks one at random:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	delivered := make(chan string)

	go func() {
		time.Sleep(50 * time.Millisecond) // a slow carrier
		delivered <- "alice"
	}()

	select {
	case who := <-delivered:
		fmt.Println("delivered to", who)
	case <-time.After(10 * time.Millisecond):
		fmt.Println("carrier timed out, retrying later")
	}
}
```

```text
carrier timed out, retrying later
```

`time.After(d)` returns a channel that receives a value once `d` has passed. The carrier takes 50ms, but the timeout fires after 10ms, so the second case wins. This pattern is how Go programs put time limits on slow operations.

## `default`: don't wait at all

A `select` with a `default` case never blocks. If no other case is ready right now, `default` runs:

```go
package main

import "fmt"

func main() {
	queue := make(chan string, 1)
	queue <- "urgent message"

	for range 2 {
		select {
		case msg := <-queue:
			fmt.Println("processing", msg)
		default:
			fmt.Println("nothing to do")
		}
	}
}
```

```text
processing urgent message
nothing to do
```

## Putting it together

These pieces combine into Go's classic concurrency patterns: a producer sends work into a channel and closes it, several workers `range` over it, and `select` adds timeouts and cancellation. You'll build these patterns properly in later courses, but you now know every building block.

## Your turn

`main` receives reminders until the channel is closed, with a one-second safety timeout. Press **Run**: the reminders arrive, but "queue empty" never does.

Fix `queueReminders` so the receiver knows when the queue is finished. Remember who is responsible for closing a channel.

## Further reading

- [Go by Example: Closing Channels](https://gobyexample.com/closing-channels)
- [Go by Example: Range over Channels](https://gobyexample.com/range-over-channels)
- [Go by Example: Select](https://gobyexample.com/select)
- [A Tour of Go: Range and Close](https://go.dev/tour/concurrency/4)
