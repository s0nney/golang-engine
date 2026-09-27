---
title: 'Assemble Stash'
quiz:
  - question: |
      Two repositories share one `Bus`: a `Repo[string, Entry]` and a `Repo[int, Config]`. A handler subscribed with `func(e Stored[string, Entry])` receives which events?
    options:
      - text: Every `Stored` event from both repositories
      - text: 'Only `Stored[string, Entry]` events, because each instantiation of a generic type is a distinct type with its own `typeKey`'
        correct: true
      - text: None, because generic event types can't be used with the bus
      - text: It panics when a `Stored[int, Config]` arrives
    explanation: |
      `Stored[string, Entry]` and `Stored[int, Config]` are different types, so
      `(*E)(nil)` gives different keys and the handlers are filed separately. The type
      assertion to `func(E)` can never see the wrong instantiation.
  - question: Why does `Put` wrap the validation error with `%w` instead of formatting it with `%v`?
    options:
      - text: '`%v` doesn''t work with errors'
      - text: 'So callers can still find the `*FieldError` values inside with `errors.AsType`, for example to highlight the bad field'
        correct: true
      - text: '`%w` is faster'
      - text: '`errors.Join` requires it'
    explanation: |
      `%w` keeps the original error in the chain. `%v` would flatten it into text, and the
      structured `FieldError`s would be lost to callers.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"iter"
    	"reflect"
    	"slices"
    	"strconv"
    	"strings"
    )

    // ---- From chapter 5: OrderedMap ----

    // OrderedMap is a map that remembers insertion order.
    type OrderedMap[K comparable, V any] struct {
    	keys []K
    	vals map[K]V
    }

    func (m *OrderedMap[K, V]) Set(k K, v V) {
    	if m.vals == nil {
    		m.vals = make(map[K]V)
    	}
    	if _, ok := m.vals[k]; !ok {
    		m.keys = append(m.keys, k)
    	}
    	m.vals[k] = v
    }

    func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
    	v, ok := m.vals[k]
    	return v, ok
    }

    func (m *OrderedMap[K, V]) Delete(k K) bool {
    	if _, ok := m.vals[k]; !ok {
    		return false
    	}
    	delete(m.vals, k)
    	i := slices.Index(m.keys, k)
    	m.keys = slices.Delete(m.keys, i, i+1)
    	return true
    }

    func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
    	return func(yield func(K, V) bool) {
    		for _, k := range m.keys {
    			if !yield(k, m.vals[k]) {
    				return
    			}
    		}
    	}
    }

    // ---- From chapter 5: LRU (now with Remove) ----

    type node[K comparable, V any] struct {
    	key        K
    	val        V
    	prev, next *node[K, V]
    }

    // LRU is a fixed-capacity cache that evicts the least recently used entry.
    type LRU[K comparable, V any] struct {
    	capacity int
    	items    map[K]*node[K, V]
    	root     node[K, V]
    }

    func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
    	c := &LRU[K, V]{capacity: max(capacity, 1), items: make(map[K]*node[K, V])}
    	c.root.next, c.root.prev = &c.root, &c.root
    	return c
    }

    func (c *LRU[K, V]) unlink(n *node[K, V]) {
    	n.prev.next = n.next
    	n.next.prev = n.prev
    }

    func (c *LRU[K, V]) pushFront(n *node[K, V]) {
    	n.prev = &c.root
    	n.next = c.root.next
    	c.root.next.prev = n
    	c.root.next = n
    }

    func (c *LRU[K, V]) Get(k K) (V, bool) {
    	n, ok := c.items[k]
    	if !ok {
    		var zero V
    		return zero, false
    	}
    	c.unlink(n)
    	c.pushFront(n)
    	return n.val, true
    }

    func (c *LRU[K, V]) Put(k K, v V) {
    	if n, ok := c.items[k]; ok {
    		n.val = v
    		c.unlink(n)
    		c.pushFront(n)
    		return
    	}
    	if len(c.items) >= c.capacity {
    		oldest := c.root.prev
    		c.unlink(oldest)
    		delete(c.items, oldest.key)
    	}
    	n := &node[K, V]{key: k, val: v}
    	c.pushFront(n)
    	c.items[k] = n
    }

    // Remove drops k from the cache and reports whether it was there.
    func (c *LRU[K, V]) Remove(k K) bool {
    	n, ok := c.items[k]
    	if !ok {
    		return false
    	}
    	c.unlink(n)
    	delete(c.items, k)
    	return true
    }

    // ---- From this chapter: Store and Cached ----

    // Store is anything that stores values by key. *OrderedMap satisfies it.
    type Store[K comparable, V any] interface {
    	Get(k K) (V, bool)
    	Set(k K, v V)
    	Delete(k K) bool
    	All() iter.Seq2[K, V]
    }

    // Cached wraps a Store with an LRU cache for reads. It's a Store too.
    type Cached[K comparable, V any] struct {
    	store        Store[K, V]
    	cache        *LRU[K, V]
    	hits, misses int
    }

    // NewCached returns a Cached in front of store, caching up to capacity values.
    func NewCached[K comparable, V any](store Store[K, V], capacity int) *Cached[K, V] {
    	return &Cached[K, V]{store: store, cache: NewLRU[K, V](capacity)}
    }

    // Get serves k from the cache (a hit) or else from the store (a miss),
    // caching what the store returns.
    func (c *Cached[K, V]) Get(k K) (V, bool) {
    	if v, ok := c.cache.Get(k); ok {
    		c.hits++
    		return v, true
    	}
    	c.misses++
    	v, ok := c.store.Get(k)
    	if ok {
    		c.cache.Put(k, v)
    	}
    	return v, ok
    }

    // Set writes through: to the store first, then the cache.
    func (c *Cached[K, V]) Set(k K, v V) {
    	c.store.Set(k, v)
    	c.cache.Put(k, v)
    }

    // Delete removes k from the store and the cache.
    func (c *Cached[K, V]) Delete(k K) bool {
    	c.cache.Remove(k)
    	return c.store.Delete(k)
    }

    // All iterates the underlying store.
    func (c *Cached[K, V]) All() iter.Seq2[K, V] { return c.store.All() }

    // Stats reports cache hits and misses so far.
    func (c *Cached[K, V]) Stats() (hits, misses int) { return c.hits, c.misses }

    // ---- From chapter 6: Bus ----

    // Bus delivers events to handlers subscribed to their type.
    type Bus struct {
    	handlers map[any][]any
    }

    func typeKey[E any]() any { return (*E)(nil) }

    func (b *Bus) Subscribe[E any](fn func(E)) {
    	if b.handlers == nil {
    		b.handlers = make(map[any][]any)
    	}
    	k := typeKey[E]()
    	b.handlers[k] = append(b.handlers[k], fn)
    }

    func (b *Bus) Publish[E any](e E) int {
    	hs := b.handlers[typeKey[E]()]
    	for _, h := range hs {
    		h.(func(E))(e)
    	}
    	return len(hs)
    }

    // ---- From chapter 7: Validate ----

    // FieldError reports a field that failed a validation rule.
    type FieldError struct {
    	Field string
    	Rule  string
    }

    func (e *FieldError) Error() string { return e.Field + ": failed " + e.Rule }

    // Validate checks a struct (or pointer to one) against its `validate` tags.
    // Each field gets at most one error: the first rule it fails, in tag order.
    // Errors are joined with errors.Join, in field order.
    func Validate(v any) error {
    	rv := reflect.ValueOf(v)
    	for rv.Kind() == reflect.Pointer && !rv.IsNil() {
    		rv = rv.Elem()
    	}
    	if rv.Kind() != reflect.Struct {
    		return fmt.Errorf("stash: Validate needs a struct, got %T", v)
    	}
    	var errs []error
    	for f, fv := range rv.Fields() {
    		tag, ok := f.Tag.Lookup("validate")
    		if !ok || !f.IsExported() {
    			continue
    		}
    		for rule := range strings.SplitSeq(tag, ",") {
    			ok, err := checkRule(fv, rule)
    			if err != nil {
    				errs = append(errs, fmt.Errorf("%s: %w", f.Name, err))
    				break
    			}
    			if !ok {
    				errs = append(errs, &FieldError{Field: f.Name, Rule: rule})
    				break
    			}
    		}
    	}
    	return errors.Join(errs...)
    }

    // checkRule reports whether v passes one rule: "required", "min=N" or "max=N".
    // It returns an error for unknown rules, bad numbers, or kinds a rule
    // doesn't support.
    func checkRule(v reflect.Value, rule string) (bool, error) {
    	name, arg, _ := strings.Cut(rule, "=")
    	switch name {
    	case "required":
    		return !v.IsZero(), nil
    	case "min", "max":
    		limit, err := strconv.ParseFloat(arg, 64)
    		if err != nil {
    			return false, fmt.Errorf("bad number in rule %q", rule)
    		}
    		var x float64
    		switch v.Kind() {
    		case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
    			x = float64(v.Len())
    		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    			x = float64(v.Int())
    		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
    			x = float64(v.Uint())
    		case reflect.Float32, reflect.Float64:
    			x = v.Float()
    		default:
    			return false, fmt.Errorf("rule %q doesn't support kind %s", rule, v.Kind())
    		}
    		if name == "min" {
    			return x >= limit, nil
    		}
    		return x <= limit, nil
    	}
    	return false, fmt.Errorf("unknown rule %q", rule)
    }

    // ---- New: the repository ----

    // Stored is published after a value is stored.
    type Stored[K comparable, V any] struct {
    	Key   K
    	Value V
    }

    // Deleted is published after a key is deleted.
    type Deleted[K comparable] struct{ Key K }

    // Repo is a typed, validated repository. V must be a struct type (or a
    // pointer to one) so that Validate can check it.
    type Repo[K comparable, V any] struct {
    	store  Store[K, V]
    	Events Bus
    }

    // NewRepo returns a repository backed by store.
    func NewRepo[K comparable, V any](store Store[K, V]) *Repo[K, V] {
    	return &Repo[K, V]{store: store}
    }

    // NewMemoryRepo returns an in-memory repository with an LRU read cache.
    func NewMemoryRepo[K comparable, V any](cacheSize int) *Repo[K, V] {
    	return NewRepo(NewCached[K, V](&OrderedMap[K, V]{}, cacheSize))
    }

    // Put validates v, stores it under k and publishes a Stored event.
    // If v is invalid, nothing is stored or published, and the error
    // ("stash: put <k>: <validation error>") wraps Validate's error.
    func (r *Repo[K, V]) Put(k K, v V) error {
    	// ?
    	return nil
    }

    // Get returns the value stored under k.
    func (r *Repo[K, V]) Get(k K) (V, bool) {
    	// ?
    	var zero V
    	return zero, false
    }

    // Delete removes k. If it existed, it publishes a Deleted event and
    // returns true.
    func (r *Repo[K, V]) Delete(k K) bool {
    	// ?
    	return false
    }

    // All yields every key and value, in the store's order.
    func (r *Repo[K, V]) All() iter.Seq2[K, V] {
    	// ?
    	return func(yield func(K, V) bool) {}
    }

    // Where lazily yields the pairs whose value satisfies keep.
    func (r *Repo[K, V]) Where(keep func(V) bool) iter.Seq2[K, V] {
    	// ?
    	return func(yield func(K, V) bool) {}
    }

    // ---- Try it ----

    // Entry is what Stash stores.
    type Entry struct {
    	Key  string   `validate:"required,min=3"`
    	Size int64    `validate:"min=1"`
    	Tags []string `validate:"max=3"`
    }

    func main() {
    	repo := NewMemoryRepo[string, Entry](2)
    	repo.Events.Subscribe(func(e Stored[string, Entry]) { fmt.Println("stored", e.Key, e.Value.Size) })
    	repo.Events.Subscribe(func(e Deleted[string]) { fmt.Println("deleted", e.Key) })

    	fmt.Println(repo.Put("logo", Entry{Key: "logo.png", Size: 2048, Tags: []string{"img"}}))
    	fmt.Println(repo.Put("app", Entry{Key: "app.js", Size: 512}))
    	fmt.Println(repo.Put("bad", Entry{Key: "x"}))

    	e, ok := repo.Get("logo")
    	fmt.Println(e.Key, ok)
    	for k, v := range repo.Where(func(e Entry) bool { return e.Size > 1000 }) {
    		fmt.Println("big:", k, v.Size)
    	}
    	fmt.Println(repo.Delete("app"), repo.Delete("app"))
    	for k := range repo.All() {
    		fmt.Println("left:", k)
    	}
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"iter"
    	"reflect"
    	"slices"
    	"strconv"
    	"strings"
    )

    // ---- From chapter 5: OrderedMap ----

    // OrderedMap is a map that remembers insertion order.
    type OrderedMap[K comparable, V any] struct {
    	keys []K
    	vals map[K]V
    }

    func (m *OrderedMap[K, V]) Set(k K, v V) {
    	if m.vals == nil {
    		m.vals = make(map[K]V)
    	}
    	if _, ok := m.vals[k]; !ok {
    		m.keys = append(m.keys, k)
    	}
    	m.vals[k] = v
    }

    func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
    	v, ok := m.vals[k]
    	return v, ok
    }

    func (m *OrderedMap[K, V]) Delete(k K) bool {
    	if _, ok := m.vals[k]; !ok {
    		return false
    	}
    	delete(m.vals, k)
    	i := slices.Index(m.keys, k)
    	m.keys = slices.Delete(m.keys, i, i+1)
    	return true
    }

    func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
    	return func(yield func(K, V) bool) {
    		for _, k := range m.keys {
    			if !yield(k, m.vals[k]) {
    				return
    			}
    		}
    	}
    }

    // ---- From chapter 5: LRU (now with Remove) ----

    type node[K comparable, V any] struct {
    	key        K
    	val        V
    	prev, next *node[K, V]
    }

    // LRU is a fixed-capacity cache that evicts the least recently used entry.
    type LRU[K comparable, V any] struct {
    	capacity int
    	items    map[K]*node[K, V]
    	root     node[K, V]
    }

    func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
    	c := &LRU[K, V]{capacity: max(capacity, 1), items: make(map[K]*node[K, V])}
    	c.root.next, c.root.prev = &c.root, &c.root
    	return c
    }

    func (c *LRU[K, V]) unlink(n *node[K, V]) {
    	n.prev.next = n.next
    	n.next.prev = n.prev
    }

    func (c *LRU[K, V]) pushFront(n *node[K, V]) {
    	n.prev = &c.root
    	n.next = c.root.next
    	c.root.next.prev = n
    	c.root.next = n
    }

    func (c *LRU[K, V]) Get(k K) (V, bool) {
    	n, ok := c.items[k]
    	if !ok {
    		var zero V
    		return zero, false
    	}
    	c.unlink(n)
    	c.pushFront(n)
    	return n.val, true
    }

    func (c *LRU[K, V]) Put(k K, v V) {
    	if n, ok := c.items[k]; ok {
    		n.val = v
    		c.unlink(n)
    		c.pushFront(n)
    		return
    	}
    	if len(c.items) >= c.capacity {
    		oldest := c.root.prev
    		c.unlink(oldest)
    		delete(c.items, oldest.key)
    	}
    	n := &node[K, V]{key: k, val: v}
    	c.pushFront(n)
    	c.items[k] = n
    }

    // Remove drops k from the cache and reports whether it was there.
    func (c *LRU[K, V]) Remove(k K) bool {
    	n, ok := c.items[k]
    	if !ok {
    		return false
    	}
    	c.unlink(n)
    	delete(c.items, k)
    	return true
    }

    // ---- From this chapter: Store and Cached ----

    // Store is anything that stores values by key. *OrderedMap satisfies it.
    type Store[K comparable, V any] interface {
    	Get(k K) (V, bool)
    	Set(k K, v V)
    	Delete(k K) bool
    	All() iter.Seq2[K, V]
    }

    // Cached wraps a Store with an LRU cache for reads. It's a Store too.
    type Cached[K comparable, V any] struct {
    	store        Store[K, V]
    	cache        *LRU[K, V]
    	hits, misses int
    }

    // NewCached returns a Cached in front of store, caching up to capacity values.
    func NewCached[K comparable, V any](store Store[K, V], capacity int) *Cached[K, V] {
    	return &Cached[K, V]{store: store, cache: NewLRU[K, V](capacity)}
    }

    // Get serves k from the cache (a hit) or else from the store (a miss),
    // caching what the store returns.
    func (c *Cached[K, V]) Get(k K) (V, bool) {
    	if v, ok := c.cache.Get(k); ok {
    		c.hits++
    		return v, true
    	}
    	c.misses++
    	v, ok := c.store.Get(k)
    	if ok {
    		c.cache.Put(k, v)
    	}
    	return v, ok
    }

    // Set writes through: to the store first, then the cache.
    func (c *Cached[K, V]) Set(k K, v V) {
    	c.store.Set(k, v)
    	c.cache.Put(k, v)
    }

    // Delete removes k from the store and the cache.
    func (c *Cached[K, V]) Delete(k K) bool {
    	c.cache.Remove(k)
    	return c.store.Delete(k)
    }

    // All iterates the underlying store.
    func (c *Cached[K, V]) All() iter.Seq2[K, V] { return c.store.All() }

    // Stats reports cache hits and misses so far.
    func (c *Cached[K, V]) Stats() (hits, misses int) { return c.hits, c.misses }

    // ---- From chapter 6: Bus ----

    // Bus delivers events to handlers subscribed to their type.
    type Bus struct {
    	handlers map[any][]any
    }

    func typeKey[E any]() any { return (*E)(nil) }

    func (b *Bus) Subscribe[E any](fn func(E)) {
    	if b.handlers == nil {
    		b.handlers = make(map[any][]any)
    	}
    	k := typeKey[E]()
    	b.handlers[k] = append(b.handlers[k], fn)
    }

    func (b *Bus) Publish[E any](e E) int {
    	hs := b.handlers[typeKey[E]()]
    	for _, h := range hs {
    		h.(func(E))(e)
    	}
    	return len(hs)
    }

    // ---- From chapter 7: Validate ----

    // FieldError reports a field that failed a validation rule.
    type FieldError struct {
    	Field string
    	Rule  string
    }

    func (e *FieldError) Error() string { return e.Field + ": failed " + e.Rule }

    // Validate checks a struct (or pointer to one) against its `validate` tags.
    // Each field gets at most one error: the first rule it fails, in tag order.
    // Errors are joined with errors.Join, in field order.
    func Validate(v any) error {
    	rv := reflect.ValueOf(v)
    	for rv.Kind() == reflect.Pointer && !rv.IsNil() {
    		rv = rv.Elem()
    	}
    	if rv.Kind() != reflect.Struct {
    		return fmt.Errorf("stash: Validate needs a struct, got %T", v)
    	}
    	var errs []error
    	for f, fv := range rv.Fields() {
    		tag, ok := f.Tag.Lookup("validate")
    		if !ok || !f.IsExported() {
    			continue
    		}
    		for rule := range strings.SplitSeq(tag, ",") {
    			ok, err := checkRule(fv, rule)
    			if err != nil {
    				errs = append(errs, fmt.Errorf("%s: %w", f.Name, err))
    				break
    			}
    			if !ok {
    				errs = append(errs, &FieldError{Field: f.Name, Rule: rule})
    				break
    			}
    		}
    	}
    	return errors.Join(errs...)
    }

    // checkRule reports whether v passes one rule: "required", "min=N" or "max=N".
    // It returns an error for unknown rules, bad numbers, or kinds a rule
    // doesn't support.
    func checkRule(v reflect.Value, rule string) (bool, error) {
    	name, arg, _ := strings.Cut(rule, "=")
    	switch name {
    	case "required":
    		return !v.IsZero(), nil
    	case "min", "max":
    		limit, err := strconv.ParseFloat(arg, 64)
    		if err != nil {
    			return false, fmt.Errorf("bad number in rule %q", rule)
    		}
    		var x float64
    		switch v.Kind() {
    		case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
    			x = float64(v.Len())
    		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    			x = float64(v.Int())
    		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
    			x = float64(v.Uint())
    		case reflect.Float32, reflect.Float64:
    			x = v.Float()
    		default:
    			return false, fmt.Errorf("rule %q doesn't support kind %s", rule, v.Kind())
    		}
    		if name == "min" {
    			return x >= limit, nil
    		}
    		return x <= limit, nil
    	}
    	return false, fmt.Errorf("unknown rule %q", rule)
    }

    // ---- New: the repository ----

    // Stored is published after a value is stored.
    type Stored[K comparable, V any] struct {
    	Key   K
    	Value V
    }

    // Deleted is published after a key is deleted.
    type Deleted[K comparable] struct{ Key K }

    // Repo is a typed, validated repository. V must be a struct type (or a
    // pointer to one) so that Validate can check it.
    type Repo[K comparable, V any] struct {
    	store  Store[K, V]
    	Events Bus
    }

    // NewRepo returns a repository backed by store.
    func NewRepo[K comparable, V any](store Store[K, V]) *Repo[K, V] {
    	return &Repo[K, V]{store: store}
    }

    // NewMemoryRepo returns an in-memory repository with an LRU read cache.
    func NewMemoryRepo[K comparable, V any](cacheSize int) *Repo[K, V] {
    	return NewRepo(NewCached[K, V](&OrderedMap[K, V]{}, cacheSize))
    }

    // Put validates v, stores it under k and publishes a Stored event.
    // If v is invalid, nothing is stored or published, and the error
    // ("stash: put <k>: <validation error>") wraps Validate's error.
    func (r *Repo[K, V]) Put(k K, v V) error {
    	if err := Validate(v); err != nil {
    		return fmt.Errorf("stash: put %v: %w", k, err)
    	}
    	r.store.Set(k, v)
    	r.Events.Publish(Stored[K, V]{Key: k, Value: v})
    	return nil
    }

    // Get returns the value stored under k.
    func (r *Repo[K, V]) Get(k K) (V, bool) {
    	return r.store.Get(k)
    }

    // Delete removes k. If it existed, it publishes a Deleted event and
    // returns true.
    func (r *Repo[K, V]) Delete(k K) bool {
    	if !r.store.Delete(k) {
    		return false
    	}
    	r.Events.Publish(Deleted[K]{Key: k})
    	return true
    }

    // All yields every key and value, in the store's order.
    func (r *Repo[K, V]) All() iter.Seq2[K, V] {
    	return r.store.All()
    }

    // Where lazily yields the pairs whose value satisfies keep.
    func (r *Repo[K, V]) Where(keep func(V) bool) iter.Seq2[K, V] {
    	return func(yield func(K, V) bool) {
    		for k, v := range r.store.All() {
    			if keep(v) && !yield(k, v) {
    				return
    			}
    		}
    	}
    }

    // ---- Try it ----

    // Entry is what Stash stores.
    type Entry struct {
    	Key  string   `validate:"required,min=3"`
    	Size int64    `validate:"min=1"`
    	Tags []string `validate:"max=3"`
    }

    func main() {
    	repo := NewMemoryRepo[string, Entry](2)
    	repo.Events.Subscribe(func(e Stored[string, Entry]) { fmt.Println("stored", e.Key, e.Value.Size) })
    	repo.Events.Subscribe(func(e Deleted[string]) { fmt.Println("deleted", e.Key) })

    	fmt.Println(repo.Put("logo", Entry{Key: "logo.png", Size: 2048, Tags: []string{"img"}}))
    	fmt.Println(repo.Put("app", Entry{Key: "app.js", Size: 512}))
    	fmt.Println(repo.Put("bad", Entry{Key: "x"}))

    	e, ok := repo.Get("logo")
    	fmt.Println(e.Key, ok)
    	for k, v := range repo.Where(func(e Entry) bool { return e.Size > 1000 }) {
    		fmt.Println("big:", k, v.Size)
    	}
    	fmt.Println(repo.Delete("app"), repo.Delete("app"))
    	for k := range repo.All() {
    		fmt.Println("left:", k)
    	}
    }
  tests: |
    package main

    import (
    	"errors"
    	"slices"
    	"strings"
    	"testing"
    )

    func newTestRepo(t *testing.T) (*Repo[string, Entry], *[]string) {
    	t.Helper()
    	repo := NewMemoryRepo[string, Entry](2)
    	var events []string
    	repo.Events.Subscribe(func(e Stored[string, Entry]) { events = append(events, "stored:"+e.Key+":"+e.Value.Key) })
    	repo.Events.Subscribe(func(e Deleted[string]) { events = append(events, "deleted:"+e.Key) })
    	return repo, &events
    }

    func TestPutAndGet(t *testing.T) {
    	repo, events := newTestRepo(t)
    	if err := repo.Put("logo", Entry{Key: "logo.png", Size: 2048}); err != nil {
    		t.Fatalf("Put(valid entry) = %v, want nil", err)
    	}
    	e, ok := repo.Get("logo")
    	if !ok || e.Key != "logo.png" || e.Size != 2048 {
    		t.Errorf(`Get("logo") = (%+v, %v), want the stored entry`, e, ok)
    	}
    	if _, ok := repo.Get("missing"); ok {
    		t.Errorf(`Get("missing") found something`)
    	}
    	if want := []string{"stored:logo:logo.png"}; !slices.Equal(*events, want) {
    		t.Errorf("events = %q, want %q", *events, want)
    	}
    }

    func TestPutRejectsInvalid(t *testing.T) {
    	repo, events := newTestRepo(t)
    	err := repo.Put("bad", Entry{Key: "x", Size: 5})
    	if err == nil {
    		t.Fatalf("Put(invalid entry) = nil, want a validation error")
    	}
    	if !strings.HasPrefix(err.Error(), "stash: put bad: ") {
    		t.Errorf("error %q should start with %q", err, "stash: put bad: ")
    	}
    	fe, ok := errors.AsType[*FieldError](err)
    	if !ok || fe.Field != "Key" {
    		t.Errorf("errors.AsType[*FieldError](err) = (%v, %v), want the Key FieldError (wrap with %%w)", fe, ok)
    	}
    	if _, ok := repo.Get("bad"); ok {
    		t.Errorf("an invalid entry was stored")
    	}
    	if len(*events) != 0 {
    		t.Errorf("an invalid Put published events: %q", *events)
    	}
    }

    func TestDelete(t *testing.T) {
    	repo, events := newTestRepo(t)
    	repo.Put("a", Entry{Key: "aaa", Size: 1})
    	if !repo.Delete("a") {
    		t.Errorf(`Delete("a") = false, want true`)
    	}
    	if repo.Delete("a") {
    		t.Errorf(`second Delete("a") = true, want false`)
    	}
    	if _, ok := repo.Get("a"); ok {
    		t.Errorf(`Get("a") after Delete still finds it (check the cache is invalidated)`)
    	}
    	want := []string{"stored:a:aaa", "deleted:a"}
    	if !slices.Equal(*events, want) {
    		t.Errorf("events = %q, want %q (Deleted only when something was deleted)", *events, want)
    	}
    }

    func TestAllAndWhere(t *testing.T) {
    	repo, _ := newTestRepo(t)
    	for i, k := range []string{"c", "a", "b", "d"} {
    		if err := repo.Put(k, Entry{Key: k + k + k, Size: int64(i+1) * 100}); err != nil {
    			t.Fatalf("Put(%q) = %v", k, err)
    		}
    	}
    	var keys []string
    	for k := range repo.All() {
    		keys = append(keys, k)
    	}
    	if !slices.Equal(keys, []string{"c", "a", "b", "d"}) {
    		t.Errorf("All() keys = %v, want insertion order [c a b d]", keys)
    	}
    	var big []string
    	for k, v := range repo.Where(func(e Entry) bool { return e.Size >= 200 }) {
    		big = append(big, k+":"+v.Key)
    	}
    	if !slices.Equal(big, []string{"a:aaa", "b:bbb", "d:ddd"}) {
    		t.Errorf("Where(Size >= 200) = %v, want [a:aaa b:bbb d:ddd]", big)
    	}
    	var first []string
    	for k := range repo.Where(func(Entry) bool { return true }) {
    		first = append(first, k)
    		break
    	}
    	if !slices.Equal(first, []string{"c"}) {
    		t.Errorf("breaking out of Where after one value gave %v, want [c]", first)
    	}
    }
---

This is it: the course's final exercise. Everything Stash needs is already in the starter, collected from earlier chapters: the `OrderedMap` backend, the `LRU` with its new `Remove`, the `Store` interface and `Cached` decorator from the last lesson, the typed `Bus`, and `Validate`. Your job is the part that ties them together: `Repo[K, V]`.

## What's provided

```go
type Stored[K comparable, V any] struct {
	Key   K
	Value V
}

type Deleted[K comparable] struct{ Key K }

type Repo[K comparable, V any] struct {
	store  Store[K, V]
	Events Bus
}

func NewRepo[K comparable, V any](store Store[K, V]) *Repo[K, V]
func NewMemoryRepo[K comparable, V any](cacheSize int) *Repo[K, V]
```

`NewMemoryRepo` builds the full stack: an `OrderedMap` wrapped in a `Cached` with an LRU of the given size, wrapped in a `Repo`. `Events` is an exported `Bus` field, so users subscribe with `repo.Events.Subscribe(func(e Stored[string, Entry]) {...})`, and the zero `Bus` is ready to use.

## What you write

- **`Put(k, v) error`**: run `Validate(v)`. If it fails, return `fmt.Errorf("stash: put %v: %w", k, err)` and do nothing else: no store write, no event. Otherwise store the value and publish a `Stored[K, V]{Key: k, Value: v}`.
- **`Get(k) (V, bool)`**: straight from the store (which is the cache, when built by `NewMemoryRepo`).
- **`Delete(k) bool`**: delete from the store. Only if something was deleted, publish `Deleted[K]{Key: k}` and return `true`.
- **`All() iter.Seq2[K, V]`**: the store's pairs, in its order.
- **`Where(keep) iter.Seq2[K, V]`**: a **lazy** filtered view of `All`. Yield only the pairs whose value passes `keep`, and stop as soon as the loop body breaks.

One thing to watch: `r.Events.Publish(Stored[K, V]{...})` infers `E` from its argument, so it publishes exactly the type subscribers asked for. Publishing, say, a plain `Entry` would reach nobody.

When the tests pass, take a moment to read the whole file top to bottom. Every line uses something from this course: named types and constraints, inference, zero values and comma-ok, pointer receivers, generic methods, iterators, interface decorators, the typed bus with `any` hidden inside, and a reflection-based validator.
