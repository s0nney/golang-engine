---
title: Spell Cooldowns
difficulty: hard
after: design-in-practice
hints:
  - 'Never call `time.Now()` inside `Caster`: store the injected `Clock` and ask it. Cooldowns are then easy: on a successful cast, remember `readyAt[name] = now.Add(spell.Cooldown)`; the spell is on cooldown while `readyAt[name].Sub(now) > 0`, and that difference is the `Remaining` time.'
  - 'Regenerate lazily. Keep `mana` and `regenFrom` (when progress towards the next point began). Before reading or spending mana, compute `ticks := int(now.Sub(regenFrom) / regenEvery)`, add them (capped at the maximum) and move `regenFrom` forward by exactly `ticks * regenEvery`, so the leftover progress is kept.'
  - 'The "a full pool stores no progress" rule: whenever the pool is full after catching up (or regen is off), set `regenFrom = now`. In `Cast`, catch up first, then check in order: unknown spell, cooldown, mana. Only a cast that passes all three may touch the mana or the cooldown.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"time"
    )

    // Clock tells the time. Inject a fake one in tests.
    type Clock interface {
    	Now() time.Time
    }

    // SystemClock is the real clock.
    type SystemClock struct{}

    func (SystemClock) Now() time.Time { return time.Now() }

    type Spell struct {
    	Name     string
    	Cost     int           // mana
    	Cooldown time.Duration // how long after a cast before it can be cast again
    }

    var (
    	ErrKnown        = errors.New("spell already known")
    	ErrInvalidSpell = errors.New("invalid spell")
    )

    type UnknownSpellError struct {
    	Name string
    }

    func (e *UnknownSpellError) Error() string {
    	return fmt.Sprintf("unknown spell %q", e.Name)
    }

    type CooldownError struct {
    	Spell     string
    	Remaining time.Duration
    }

    func (e *CooldownError) Error() string {
    	return fmt.Sprintf("%s is on cooldown for %v", e.Spell, e.Remaining)
    }

    type ManaError struct {
    	Need, Have int
    }

    func (e *ManaError) Error() string {
    	return fmt.Sprintf("need %d mana, have %d", e.Need, e.Have)
    }

    type Caster struct {
    }

    func NewCaster(clock Clock, maxMana int, regenEvery time.Duration) *Caster {
    	return &Caster{}
    }

    func (c *Caster) Learn(s Spell) error {
    	return nil
    }

    func (c *Caster) Cast(name string) error {
    	return nil
    }

    func (c *Caster) Mana() int {
    	return 0
    }

    func (c *Caster) ReadyIn(name string) (time.Duration, bool) {
    	return 0, false
    }

    // fakeClock is a Clock you move by hand.
    type fakeClock struct{ now time.Time }

    func (f *fakeClock) Now() time.Time          { return f.now }
    func (f *fakeClock) Advance(d time.Duration) { f.now = f.now.Add(d) }

    func main() {
    	clock := &fakeClock{now: time.Date(1200, 3, 1, 12, 0, 0, 0, time.UTC)}
    	mage := NewCaster(clock, 50, 2*time.Second)
    	mage.Learn(Spell{Name: "fireball", Cost: 30, Cooldown: 5 * time.Second})

    	fmt.Println(mage.Cast("fireball"), mage.Mana()) // want <nil> 20
    	clock.Advance(3 * time.Second)
    	fmt.Println(mage.Cast("fireball"), mage.Mana()) // want fireball is on cooldown for 2s 21
    	clock.Advance(2 * time.Second)
    	fmt.Println(mage.Cast("fireball"), mage.Mana()) // want need 30 mana, have 22 22
    	fmt.Println(mage.Cast("frostbolt"))             // want unknown spell "frostbolt"
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"time"
    )

    // Clock tells the time. Inject a fake one in tests.
    type Clock interface {
    	Now() time.Time
    }

    // SystemClock is the real clock.
    type SystemClock struct{}

    func (SystemClock) Now() time.Time { return time.Now() }

    type Spell struct {
    	Name     string
    	Cost     int           // mana
    	Cooldown time.Duration // how long after a cast before it can be cast again
    }

    var (
    	ErrKnown        = errors.New("spell already known")
    	ErrInvalidSpell = errors.New("invalid spell")
    )

    type UnknownSpellError struct {
    	Name string
    }

    func (e *UnknownSpellError) Error() string {
    	return fmt.Sprintf("unknown spell %q", e.Name)
    }

    type CooldownError struct {
    	Spell     string
    	Remaining time.Duration
    }

    func (e *CooldownError) Error() string {
    	return fmt.Sprintf("%s is on cooldown for %v", e.Spell, e.Remaining)
    }

    type ManaError struct {
    	Need, Have int
    }

    func (e *ManaError) Error() string {
    	return fmt.Sprintf("need %d mana, have %d", e.Need, e.Have)
    }

    // Caster is a mage with a mana pool and a spellbook.
    type Caster struct {
    	clock      Clock
    	maxMana    int
    	mana       int
    	regenEvery time.Duration
    	regenFrom  time.Time // when progress towards the next mana point started
    	spells     map[string]Spell
    	readyAt    map[string]time.Time
    }

    var _ Clock = SystemClock{}

    // NewCaster returns a caster with a full pool of maxMana that regains one
    // point every regenEvery (never, if regenEvery <= 0), reading time from clock.
    func NewCaster(clock Clock, maxMana int, regenEvery time.Duration) *Caster {
    	maxMana = max(0, maxMana)
    	return &Caster{
    		clock:      clock,
    		maxMana:    maxMana,
    		mana:       maxMana,
    		regenEvery: regenEvery,
    		regenFrom:  clock.Now(),
    		spells:     map[string]Spell{},
    		readyAt:    map[string]time.Time{},
    	}
    }

    // regen brings the mana up to date at time now.
    func (c *Caster) regen(now time.Time) {
    	if c.mana >= c.maxMana || c.regenEvery <= 0 {
    		c.regenFrom = now // a full pool stores no progress
    		return
    	}
    	ticks := int(now.Sub(c.regenFrom) / c.regenEvery)
    	c.regenFrom = c.regenFrom.Add(time.Duration(ticks) * c.regenEvery)
    	c.mana = min(c.maxMana, c.mana+ticks)
    	if c.mana == c.maxMana {
    		c.regenFrom = now
    	}
    }

    // Learn adds s to the spellbook.
    func (c *Caster) Learn(s Spell) error {
    	if s.Name == "" || s.Cost < 0 || s.Cooldown < 0 {
    		return fmt.Errorf("learn %+v: %w", s, ErrInvalidSpell)
    	}
    	if _, ok := c.spells[s.Name]; ok {
    		return fmt.Errorf("learn %q: %w", s.Name, ErrKnown)
    	}
    	c.spells[s.Name] = s
    	return nil
    }

    // Cast casts the named spell, or explains why it can't.
    func (c *Caster) Cast(name string) error {
    	s, ok := c.spells[name]
    	if !ok {
    		return &UnknownSpellError{Name: name}
    	}
    	now := c.clock.Now()
    	c.regen(now)
    	if wait := c.readyAt[name].Sub(now); wait > 0 {
    		return &CooldownError{Spell: name, Remaining: wait}
    	}
    	if c.mana < s.Cost {
    		return &ManaError{Need: s.Cost, Have: c.mana}
    	}
    	c.mana -= s.Cost
    	c.readyAt[name] = now.Add(s.Cooldown)
    	return nil
    }

    // Mana returns the current mana.
    func (c *Caster) Mana() int {
    	c.regen(c.clock.Now())
    	return c.mana
    }

    // ReadyIn returns how long until the named spell can be cast again (0 if it
    // can be cast now, ignoring mana), and false if the spell isn't known.
    func (c *Caster) ReadyIn(name string) (time.Duration, bool) {
    	if _, ok := c.spells[name]; !ok {
    		return 0, false
    	}
    	return max(0, c.readyAt[name].Sub(c.clock.Now())), true
    }

    // fakeClock is a Clock you move by hand.
    type fakeClock struct{ now time.Time }

    func (f *fakeClock) Now() time.Time          { return f.now }
    func (f *fakeClock) Advance(d time.Duration) { f.now = f.now.Add(d) }

    func main() {
    	clock := &fakeClock{now: time.Date(1200, 3, 1, 12, 0, 0, 0, time.UTC)}
    	mage := NewCaster(clock, 50, 2*time.Second)
    	mage.Learn(Spell{Name: "fireball", Cost: 30, Cooldown: 5 * time.Second})

    	fmt.Println(mage.Cast("fireball"), mage.Mana()) // <nil> 20
    	clock.Advance(3 * time.Second)
    	fmt.Println(mage.Cast("fireball"), mage.Mana()) // fireball is on cooldown for 2s 21
    	clock.Advance(2 * time.Second)
    	fmt.Println(mage.Cast("fireball"), mage.Mana()) // need 30 mana, have 22
    	fmt.Println(mage.Cast("frostbolt"))             // unknown spell "frostbolt"
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"testing"
    	"time"
    )

    // stubClock is the test's own Clock, set far from the real time on purpose.
    type stubClock struct{ now time.Time }

    func (s *stubClock) Now() time.Time          { return s.now }
    func (s *stubClock) Advance(d time.Duration) { s.now = s.now.Add(d) }

    func newStub() *stubClock {
    	return &stubClock{now: time.Date(1066, 10, 14, 9, 0, 0, 0, time.UTC)}
    }

    const sec = time.Second

    func mustLearn(t *testing.T, c *Caster, spells ...Spell) {
    	t.Helper()
    	for _, s := range spells {
    		if err := c.Learn(s); err != nil {
    			t.Fatalf("Learn(%+v) = %v, want nil", s, err)
    		}
    	}
    }

    // outcome summarises Cast's result as "ok", "unknown <name>",
    // "cooldown <spell> <remaining>" or "mana <need>/<have>".
    func outcome(err error) string {
    	if err == nil {
    		return "ok"
    	}
    	if e, ok := errors.AsType[*UnknownSpellError](err); ok {
    		return "unknown " + e.Name
    	}
    	if e, ok := errors.AsType[*CooldownError](err); ok {
    		return fmt.Sprintf("cooldown %s %v", e.Spell, e.Remaining)
    	}
    	if e, ok := errors.AsType[*ManaError](err); ok {
    		return fmt.Sprintf("mana %d/%d", e.Need, e.Have)
    	}
    	return "unexpected error " + err.Error()
    }

    type step struct {
    	advance time.Duration // move the clock first
    	cast    string        // then cast this ("" to only read Mana)
    	want    string        // outcome of the cast
    	mana    int           // Mana() afterwards
    }

    func play(t *testing.T, name string, c *Caster, clock *stubClock, steps []step) {
    	t.Helper()
    	elapsed := time.Duration(0)
    	for _, s := range steps {
    		clock.Advance(s.advance)
    		elapsed += s.advance
    		if s.cast != "" {
    			if got := outcome(c.Cast(s.cast)); got != s.want {
    				t.Errorf("%s, t=%v: Cast(%q) gave %q, want %q", name, elapsed, s.cast, got, s.want)
    			}
    		}
    		if got := c.Mana(); got != s.mana {
    			t.Errorf("%s, t=%v: Mana() = %d, want %d", name, elapsed, got, s.mana)
    			return
    		}
    	}
    }

    func TestCastScenario(t *testing.T) {
    	clock := newStub()
    	mage := NewCaster(clock, 50, 2*sec)
    	mustLearn(t, mage, Spell{"fireball", 30, 5 * sec}, Spell{"spark", 5, 0})
    	play(t, "50 mana, +1 every 2s", mage, clock, []step{
    		{0, "fireball", "ok", 20},
    		{3 * sec, "fireball", "cooldown fireball 2s", 21},
    		{2 * sec, "fireball", "mana 30/22", 22}, // off cooldown, but short of mana
    		{0, "frostbolt", "unknown frostbolt", 22},
    		{0, "spark", "ok", 17},
    		{0, "spark", "ok", 12}, // no cooldown at all
    		{35 * sec, "fireball", "ok", 0},
    	})
    }

    func TestRegen(t *testing.T) {
    	clock := newStub()
    	mage := NewCaster(clock, 10, 3*sec)
    	mustLearn(t, mage, Spell{"bolt", 4, 0})
    	play(t, "10 mana, +1 every 3s", mage, clock, []step{
    		{time.Hour, "", "", 10}, // a full pool stores no progress...
    		{0, "bolt", "ok", 6},    // ...so the first point comes 3s after this cast
    		{2999 * time.Millisecond, "", "", 6},
    		{time.Millisecond, "", "", 7},
    		{2 * sec, "bolt", "ok", 3}, // spending doesn't reset the 2s of progress
    		{sec, "", "", 4},
    		{7 * sec, "", "", 6},
    		{20 * sec, "", "", 10}, // capped at the maximum
    		{sec, "bolt", "ok", 6},
    		{2 * sec, "", "", 6},
    		{sec, "", "", 7},
    	})

    	noRegen := NewCaster(clock, 8, 0)
    	mustLearn(t, noRegen, Spell{"bolt", 4, 0})
    	play(t, "8 mana, no regen", noRegen, clock, []step{
    		{0, "bolt", "ok", 4},
    		{time.Hour, "bolt", "ok", 0},
    		{time.Hour, "bolt", "mana 4/0", 0},
    	})
    }

    func TestCooldowns(t *testing.T) {
    	clock := newStub()
    	mage := NewCaster(clock, 100, 0)
    	mustLearn(t, mage, Spell{"meteor", 60, 10 * sec}, Spell{"ward", 10, 3 * sec})
    	if d, ok := mage.ReadyIn("meteor"); d != 0 || !ok {
    		t.Errorf("before any cast: ReadyIn(\"meteor\") = %v, %v, want 0s, true", d, ok)
    	}
    	play(t, "100 mana, no regen", mage, clock, []step{
    		{0, "meteor", "ok", 40},
    		{sec, "ward", "ok", 30}, // cooldowns are per spell
    		{sec, "meteor", "cooldown meteor 8s", 30},
    		{0, "ward", "cooldown ward 2s", 30},
    		{2 * sec, "ward", "ok", 20}, // exactly when the cooldown ends
    		{6 * sec, "meteor", "mana 60/20", 20},
    	})
    	if d, ok := mage.ReadyIn("meteor"); d != 0 || !ok {
    		t.Errorf("after the cooldown: ReadyIn(\"meteor\") = %v, %v, want 0s, true", d, ok)
    	}
    	if err := mage.Cast("ward"); err != nil {
    		t.Fatalf("t=10s: Cast(\"ward\") = %v, want nil", err)
    	}
    	if d, ok := mage.ReadyIn("ward"); d != 3*sec || !ok {
    		t.Errorf("right after casting ward: ReadyIn(\"ward\") = %v, %v, want 3s, true", d, ok)
    	}
    	clock.Advance(500 * time.Millisecond)
    	if d, _ := mage.ReadyIn("ward"); d != 2500*time.Millisecond {
    		t.Errorf("0.5s later: ReadyIn(\"ward\") = %v, want 2.5s", d)
    	}
    	if d, ok := mage.ReadyIn("teleport"); d != 0 || ok {
    		t.Errorf("ReadyIn of an unknown spell = %v, %v, want 0s, false", d, ok)
    	}
    }

    func TestFailedCastChangesNothing(t *testing.T) {
    	clock := newStub()
    	mage := NewCaster(clock, 20, 0)
    	mustLearn(t, mage, Spell{"nova", 15, 4 * sec}, Spell{"blast", 25, sec})
    	play(t, "20 mana, no regen", mage, clock, []step{
    		{0, "blast", "mana 25/20", 20}, // doesn't start blast's cooldown or spend mana
    		{0, "nova", "ok", 5},
    		{sec, "nova", "cooldown nova 3s", 5}, // cooldown is checked before mana
    		{3 * sec, "nova", "mana 15/5", 5},
    	})
    	if d, _ := mage.ReadyIn("blast"); d != 0 {
    		t.Errorf("a failed cast of blast started its cooldown: ReadyIn = %v, want 0s", d)
    	}
    }

    func TestLearn(t *testing.T) {
    	clock := newStub()
    	mage := NewCaster(clock, 30, 0)
    	mustLearn(t, mage, Spell{"heal", 10, 2 * sec})
    	mage.Cast("heal")
    	if err := mage.Learn(Spell{"heal", 0, 0}); !errors.Is(err, ErrKnown) {
    		t.Errorf("learning \"heal\" twice = %v, want an error wrapping ErrKnown", err)
    	}
    	if got := outcome(mage.Cast("heal")); got != "cooldown heal 2s" {
    		t.Errorf("after a rejected re-Learn, Cast(\"heal\") gave %q, want the original spell still on cooldown", got)
    	}
    	for _, bad := range []Spell{{"", 1, 0}, {"drain", -5, 0}, {"stall", 1, -sec}} {
    		if err := mage.Learn(bad); !errors.Is(err, ErrInvalidSpell) {
    			t.Errorf("Learn(%+v) = %v, want an error wrapping ErrInvalidSpell", bad, err)
    		}
    	}
    	if _, ok := mage.ReadyIn("drain"); ok {
    		t.Errorf("a rejected spell was added to the spellbook")
    	}
    	mustLearn(t, mage, Spell{"free", 0, 0})
    	for range 3 {
    		if err := mage.Cast("free"); err != nil {
    			t.Errorf("casting a free spell with no cooldown = %v, want nil", err)
    		}
    	}
    }

    func TestErrorMessages(t *testing.T) {
    	clock := newStub()
    	mage := NewCaster(clock, 5, 0)
    	mustLearn(t, mage, Spell{"bolt", 3, 90 * sec}, Spell{"big", 9, 0})
    	mage.Cast("bolt")
    	for _, tt := range []struct{ cast, want string }{
    		{"bolt", "bolt is on cooldown for 1m30s"},
    		{"big", "need 9 mana, have 2"},
    		{"nope", `unknown spell "nope"`},
    	} {
    		err := mage.Cast(tt.cast)
    		if err == nil || err.Error() != tt.want {
    			t.Errorf("Cast(%q) error = %v, want %q", tt.cast, err, tt.want)
    		}
    	}
    }

    func TestCastersShareAClockButNothingElse(t *testing.T) {
    	clock := newStub()
    	a, b := NewCaster(clock, 10, sec), NewCaster(clock, 10, sec)
    	mustLearn(t, a, Spell{"zap", 6, 5 * sec})
    	mustLearn(t, b, Spell{"zap", 6, 5 * sec})
    	a.Cast("zap")
    	if got := outcome(b.Cast("zap")); got != "ok" || a.Mana() != 4 || b.Mana() != 4 {
    		t.Errorf("two casters: b.Cast(\"zap\") gave %q, manas %d and %d, want ok, 4 and 4", got, a.Mana(), b.Mana())
    	}
    	clock.Advance(2 * sec)
    	if a.Mana() != 6 || b.Mana() != 6 {
    		t.Errorf("2s later, manas %d and %d, want 6 and 6 (time comes from the injected clock)", a.Mana(), b.Mana())
    	}
    }

    func TestNegativeMax(t *testing.T) {
    	mage := NewCaster(newStub(), -10, sec)
    	mustLearn(t, mage, Spell{"free", 0, 0})
    	if mage.Mana() != 0 || mage.Cast("free") != nil {
    		t.Errorf("NewCaster with a negative max: Mana() = %d, want 0, and free spells still work", mage.Mana())
    	}
    }
---

Build the spell system for the game's mages. Anything involving time is
notoriously hard to test, so the `Caster` never asks the operating system for
the time: it gets a `Clock` **injected** and asks that. The game passes
`SystemClock{}`, and the tests pass a clock they move by hand.

```go
type Clock interface {
	Now() time.Time
}
```

`NewCaster(clock, maxMana, regenEvery)` returns a caster with a **full** pool of
`maxMana` (a negative value counts as 0) and an empty spellbook.

1. **Regeneration.** The caster regains 1 mana for every full `regenEvery` that
   passes, up to `maxMana`. If `regenEvery <= 0` it never regenerates. Progress
   towards the next point survives spending, but a **full pool stores no
   progress**: after spending from a full pool, the first point comes back a
   whole `regenEvery` later. `Mana()` returns the up-to-date amount.
2. **`Learn(spell)`** adds a spell. A duplicate name returns an error wrapping
   `ErrKnown` (and the original spell stays as it was). An empty name, a
   negative cost or a negative cooldown returns an error wrapping
   `ErrInvalidSpell`.
3. **`Cast(name)`** checks, in this order:
   - the spell isn't known: return a `*UnknownSpellError`;
   - the spell's cooldown hasn't finished: return a `*CooldownError` with the
     remaining time;
   - there isn't enough mana: return a `*ManaError` with the cost and the
     current mana.

   Otherwise it spends the mana, starts that spell's cooldown (it can be cast
   again exactly `Cooldown` later) and returns `nil`. A failed cast changes
   nothing. Cooldowns are per spell.
4. **`ReadyIn(name)`** returns the time left on the spell's cooldown (0 if it's
   ready, ignoring mana) and `true`, or `0, false` for an unknown spell.

The error types and their messages are already written for you.

## Example

```go
clock := &fakeClock{now: time.Date(1200, 3, 1, 12, 0, 0, 0, time.UTC)}
mage := NewCaster(clock, 50, 2*time.Second)
mage.Learn(Spell{Name: "fireball", Cost: 30, Cooldown: 5 * time.Second})

mage.Cast("fireball")          // nil, mana 20
clock.Advance(3 * time.Second) // mana 21 (one point per 2s, 1s of progress kept)
mage.Cast("fireball")          // fireball is on cooldown for 2s
clock.Advance(2 * time.Second) // mana 22
mage.Cast("fireball")          // need 30 mana, have 22
mage.Cast("frostbolt")         // unknown spell "frostbolt"
```

## Constraints

- The tests' clock is set to the year 1066, so any call to `time.Now()` or
  `time.Since` inside `Caster` shows up as absurd cooldowns.
- Everything is exact: no rounding, and a cooldown that ends at time `t` allows
  a cast at exactly `t`.
