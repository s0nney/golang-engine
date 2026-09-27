---
title: 'Your Turn: A Fake Clock and a Spy'
quiz:
  - question: |
      Why must `FakeClock`'s methods use a pointer receiver
      (`func (c *FakeClock) Advance(...)`)?
    options:
      - text: Pointer receivers are faster
      - text: With a value receiver, `Advance` would change a copy, and the code holding the clock would never see time move
        correct: true
      - text: Interfaces can only hold pointers
      - text: It's required by the `time` package
    explanation: |
      The test and the `Reminder` must share one clock. A value receiver
      gets a copy of the struct, so `Advance` would move the copy's time
      and throw it away.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"time"
    )

    // FakeClock is a Clock for tests. It starts at the time given to
    // NewFakeClock and only moves when Advance is called.
    type FakeClock struct {
    	// ?
    }

    func NewFakeClock(t time.Time) *FakeClock {
    	// ?
    	return &FakeClock{}
    }

    func (c *FakeClock) Now() time.Time {
    	// ?
    	return time.Time{}
    }

    func (c *FakeClock) Advance(d time.Duration) {
    	// ?
    }

    // SpyNotifier is a Notifier for tests. It records every message, in
    // order, in Messages.
    type SpyNotifier struct {
    	Messages []string
    }

    func (s *SpyNotifier) Notify(msg string) {
    	// ?
    }

    // ---- Ledgerly code ----

    type Cents int64

    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    // Clock tells the time. Production code uses realClock.
    type Clock interface {
    	Now() time.Time
    }

    type realClock struct{}

    func (realClock) Now() time.Time { return time.Now() }

    // Notifier sends a message to the user.
    type Notifier interface {
    	Notify(msg string)
    }

    type Bill struct {
    	Name   string
    	Amount Cents
    	Due    time.Time
    }

    // Reminder warns about bills that are due soon.
    type Reminder struct {
    	clock    Clock
    	notifier Notifier
    	bills    []Bill
    	sent     map[string]bool
    }

    func NewReminder(clock Clock, n Notifier) *Reminder {
    	return &Reminder{clock: clock, notifier: n, sent: map[string]bool{}}
    }

    func (r *Reminder) Add(b Bill) { r.bills = append(r.bills, b) }

    // Window is how far ahead Check looks for bills.
    const Window = 72 * time.Hour

    // Check notifies once about every bill that is overdue or due within
    // Window of now, and returns how many notifications it sent. A bill is
    // never notified twice.
    func (r *Reminder) Check() int {
    	now := time.Now()
    	sent := 0
    	for _, b := range r.bills {
    		if r.sent[b.Name] || b.Due.Sub(now) > Window {
    			continue
    		}
    		if b.Due.Before(now) {
    			r.notifier.Notify(fmt.Sprintf("%s (%v) is OVERDUE", b.Name, b.Amount))
    		} else {
    			r.notifier.Notify(fmt.Sprintf("%s (%v) is due %s", b.Name, b.Amount, b.Due.Format("Mon Jan 2")))
    		}
    		r.sent[b.Name] = true
    		sent++
    	}
    	return sent
    }

    func main() {
    	start := time.Date(2026, time.March, 1, 9, 0, 0, 0, time.UTC)
    	clock := NewFakeClock(start)
    	spy := &SpyNotifier{}
    	r := NewReminder(clock, spy)
    	r.Add(Bill{Name: "phone", Amount: 4500, Due: start.AddDate(0, 0, -2)})
    	r.Add(Bill{Name: "rent", Amount: 120000, Due: start.AddDate(0, 0, 2)})
    	r.Add(Bill{Name: "insurance", Amount: 8999, Due: start.AddDate(0, 0, 9)})

    	fmt.Println("clock:", clock.Now().Format(time.DateTime))
    	fmt.Println("sent:", r.Check())
    	clock.Advance(7 * 24 * time.Hour)
    	fmt.Println("clock:", clock.Now().Format(time.DateTime))
    	fmt.Println("sent:", r.Check())
    	fmt.Printf("messages: %q\n", spy.Messages)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"time"
    )

    // FakeClock is a Clock for tests. It starts at the time given to
    // NewFakeClock and only moves when Advance is called.
    type FakeClock struct {
    	now time.Time
    }

    func NewFakeClock(t time.Time) *FakeClock {
    	return &FakeClock{now: t}
    }

    func (c *FakeClock) Now() time.Time {
    	return c.now
    }

    func (c *FakeClock) Advance(d time.Duration) {
    	c.now = c.now.Add(d)
    }

    // SpyNotifier is a Notifier for tests. It records every message, in
    // order, in Messages.
    type SpyNotifier struct {
    	Messages []string
    }

    func (s *SpyNotifier) Notify(msg string) {
    	s.Messages = append(s.Messages, msg)
    }

    // ---- Ledgerly code ----

    type Cents int64

    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    // Clock tells the time. Production code uses realClock.
    type Clock interface {
    	Now() time.Time
    }

    type realClock struct{}

    func (realClock) Now() time.Time { return time.Now() }

    // Notifier sends a message to the user.
    type Notifier interface {
    	Notify(msg string)
    }

    type Bill struct {
    	Name   string
    	Amount Cents
    	Due    time.Time
    }

    // Reminder warns about bills that are due soon.
    type Reminder struct {
    	clock    Clock
    	notifier Notifier
    	bills    []Bill
    	sent     map[string]bool
    }

    func NewReminder(clock Clock, n Notifier) *Reminder {
    	return &Reminder{clock: clock, notifier: n, sent: map[string]bool{}}
    }

    func (r *Reminder) Add(b Bill) { r.bills = append(r.bills, b) }

    // Window is how far ahead Check looks for bills.
    const Window = 72 * time.Hour

    // Check notifies once about every bill that is overdue or due within
    // Window of now, and returns how many notifications it sent. A bill is
    // never notified twice.
    func (r *Reminder) Check() int {
    	now := r.clock.Now()
    	sent := 0
    	for _, b := range r.bills {
    		if r.sent[b.Name] || b.Due.Sub(now) > Window {
    			continue
    		}
    		if b.Due.Before(now) {
    			r.notifier.Notify(fmt.Sprintf("%s (%v) is OVERDUE", b.Name, b.Amount))
    		} else {
    			r.notifier.Notify(fmt.Sprintf("%s (%v) is due %s", b.Name, b.Amount, b.Due.Format("Mon Jan 2")))
    		}
    		r.sent[b.Name] = true
    		sent++
    	}
    	return sent
    }

    func main() {
    	start := time.Date(2026, time.March, 1, 9, 0, 0, 0, time.UTC)
    	clock := NewFakeClock(start)
    	spy := &SpyNotifier{}
    	r := NewReminder(clock, spy)
    	r.Add(Bill{Name: "phone", Amount: 4500, Due: start.AddDate(0, 0, -2)})
    	r.Add(Bill{Name: "rent", Amount: 120000, Due: start.AddDate(0, 0, 2)})
    	r.Add(Bill{Name: "insurance", Amount: 8999, Due: start.AddDate(0, 0, 9)})

    	fmt.Println("clock:", clock.Now().Format(time.DateTime))
    	fmt.Println("sent:", r.Check())
    	clock.Advance(7 * 24 * time.Hour)
    	fmt.Println("clock:", clock.Now().Format(time.DateTime))
    	fmt.Println("sent:", r.Check())
    	fmt.Printf("messages: %q\n", spy.Messages)
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"time"
    )

    var start = time.Date(2026, time.March, 1, 9, 0, 0, 0, time.UTC)

    func TestFakeClock(t *testing.T) {
    	c := NewFakeClock(start)
    	if got := c.Now(); !got.Equal(start) {
    		t.Fatalf("NewFakeClock(%v).Now() = %v, want %v", start, got, start)
    	}
    	if got := c.Now(); !got.Equal(start) {
    		t.Fatalf("calling Now() twice: second call = %v, want %v (a fake clock must not move on its own)", got, start)
    	}
    	c.Advance(90 * time.Minute)
    	if want := start.Add(90 * time.Minute); !c.Now().Equal(want) {
    		t.Errorf("after Advance(90m), Now() = %v, want %v", c.Now(), want)
    	}
    	c.Advance(48 * time.Hour)
    	if want := start.Add(90*time.Minute + 48*time.Hour); !c.Now().Equal(want) {
    		t.Errorf("after Advance(90m) and Advance(48h), Now() = %v, want %v", c.Now(), want)
    	}
    	var clock Clock = c // FakeClock must satisfy Clock
    	_ = clock
    }

    func TestSpyNotifier(t *testing.T) {
    	s := &SpyNotifier{}
    	if len(s.Messages) != 0 {
    		t.Fatalf("new SpyNotifier has Messages %q, want none", s.Messages)
    	}
    	var n Notifier = s
    	n.Notify("first")
    	n.Notify("second")
    	if want := []string{"first", "second"}; !slices.Equal(s.Messages, want) {
    		t.Errorf("after Notify(\"first\") and Notify(\"second\"), Messages = %q, want %q", s.Messages, want)
    	}
    }

    func TestReminderUsesClock(t *testing.T) {
    	clock := NewFakeClock(start)
    	spy := &SpyNotifier{}
    	r := NewReminder(clock, spy)
    	r.Add(Bill{Name: "phone", Amount: 4500, Due: start.AddDate(0, 0, -2)})
    	r.Add(Bill{Name: "rent", Amount: 120000, Due: start.AddDate(0, 0, 2)})
    	r.Add(Bill{Name: "insurance", Amount: 8999, Due: start.AddDate(0, 0, 9)})

    	if got := r.Check(); got != 2 {
    		t.Fatalf("on Mar 1, Check() = %d, want 2 (phone is overdue, rent is due in 2 days, insurance isn't due for 9). Does Check use r.clock?", got)
    	}
    	want := []string{"phone ($45.00) is OVERDUE", "rent ($1200.00) is due Tue Mar 3"}
    	if !slices.Equal(spy.Messages, want) {
    		t.Fatalf("on Mar 1, messages = %q, want %q", spy.Messages, want)
    	}
    	if got := r.Check(); got != 0 {
    		t.Errorf("second Check() on Mar 1 = %d, want 0 (bills are only notified once)", got)
    	}

    	clock.Advance(5 * 24 * time.Hour) // Mar 6: insurance due Mar 10 is 4 days away
    	if got := r.Check(); got != 0 {
    		t.Errorf("on Mar 6, Check() = %d, want 0 (insurance is 4 days away)", got)
    	}
    	clock.Advance(2 * 24 * time.Hour) // Mar 8
    	if got := r.Check(); got != 1 {
    		t.Fatalf("on Mar 8, Check() = %d, want 1 (insurance is due in 2 days)", got)
    	}
    	if got := spy.Messages[len(spy.Messages)-1]; got != "insurance ($89.99) is due Tue Mar 10" {
    		t.Errorf("last message = %q, want %q", got, "insurance ($89.99) is due Tue Mar 10")
    	}
    }
---

Ledgerly's `Reminder` warns you about bills that are overdue or due within the next 72 hours. Someone tried to make it testable: `NewReminder` already takes a `Clock` and a `Notifier`. But nobody wrote the test doubles, and there's a bug that means the clock doesn't actually help yet.

## Your task

1. **Write `FakeClock`.** `NewFakeClock(t)` returns a clock whose `Now()` returns `t` until `Advance(d)` moves it forward by `d`. Store the current time in the struct.
2. **Write `SpyNotifier`.** `Notify(msg)` appends `msg` to the `Messages` slice, so a test can check what was sent and in what order.
3. **Fix `Reminder.Check`.** Read it carefully. It was handed a clock and... ignores it. Make it use the injected clock.

The grader tests your `FakeClock` and `SpyNotifier` on their own, then uses them to drive `Reminder` through a week of fake time: on March 1st the overdue phone bill and the rent (due March 3rd) get reminders, the insurance (due March 10th) doesn't, and a week later it does.

**Run** prints what happens on March 1st and March 8th. Once your doubles work, look at the `sent:` numbers before and after fixing `Check`. With the real clock, every bill from March 2026 is overdue, whenever you happen to run it. That's the kind of test that silently changes meaning over time.
