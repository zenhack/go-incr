package incr

import "container/heap"

type Reactor struct {
	allocHeight uint64
	dirty       incrHeap
	constNil    Incr[any]
}

func NewReactor() *Reactor {
	r := &Reactor{}
	r.constNil = Const[any](r, nil)
	return r
}

func (r *Reactor) Stabilize() {
	for len(r.dirty) != 0 {
		b := heap.Pop(&r.dirty).(*incrBase)
		b.dirty = false
		if b.scheduledHeight < b.height {
			// This one has been de-prioritized since being scheduled;
			// throw it back in the heap for later.
			b.setDirty()
			continue
		}
		r.allocHeight = b.height + 1
		b.recompute(b)
	}
	r.allocHeight = 0
}

func (r *Reactor) addDirty(b *incrBase) {
	heap.Push(&r.dirty, b)
}

type incr interface {
	get() any
	recompute(*incrBase)
	activate(*incrBase)
	deactivate(*incrBase)
}

type incrBase struct {
	reactor                 *Reactor
	height, scheduledHeight uint64
	subscribers             map[*incrBase]struct{}
	refCount                int
	dirty                   bool
	differentFn             func(any, any) bool
	incr
}

func newBase(r *Reactor, incr incr) *incrBase {
	ret := &incrBase{
		reactor:     r,
		height:      r.allocHeight,
		subscribers: make(map[*incrBase]struct{}),
		incr:        incr,
	}
	if ret.height < r.allocHeight {
		ret.height = r.allocHeight
	}
	return ret
}

func (b *incrBase) different(x, y any) bool {
	if b.differentFn == nil {
		return x != y
	}
	return b.differentFn(x, y)
}

func (b *incrBase) subscribe(sub *incrBase) {
	b.subscribers[sub] = struct{}{}
	b.refCount++
	if b.refCount == 1 {
		b.activate(b)
	}
	if b.active() {
		sub.setMinHeight(b.height + 1)
	}
}

func (b *incrBase) unsubscribe(sub *incrBase) {
	delete(b.subscribers, sub)
	b.refCount--
	if b.refCount == 0 {
		b.deactivate(b)
	}
}

func (b *incrBase) active() bool {
	return b.refCount > 0
}

func (b *incrBase) notify() {
	for k, _ := range b.subscribers {
		k.setDirty()
	}
}

func (b *incrBase) setMinHeight(height uint64) {
	if height > b.height {
		b.height = height
		for k, _ := range b.subscribers {
			k.setMinHeight(height + 1)
		}
	}
}

func (b *incrBase) setDirty() {
	if !b.dirty {
		b.dirty = true
		b.scheduledHeight = b.height
		b.reactor.addDirty(b)
	}
}

type constIncr struct {
	value any
}

func (c constIncr) get() any              { return c.value }
func (c constIncr) recompute(b *incrBase) { b.notify() }
func (c constIncr) activate(b *incrBase)  { b.setDirty() }
func (constIncr) deactivate(*incrBase)    {}

type varIncr struct {
	value    any
	modified bool
}

func (v *varIncr) set(b *incrBase, value any) {
	if !b.different(v.value, value) {
		return
	}
	v.value = value
	v.modified = true
	if b.active() {
		b.setDirty()
	}
}

func (v *varIncr) get() any { return v.value }

func (v *varIncr) recompute(b *incrBase) {
	v.modified = false
	b.notify()
}

func (v *varIncr) activate(b *incrBase) {
	if v.modified {
		b.setDirty()
	}
}

func (*varIncr) deactivate(*incrBase) {}

type Var[T any] struct {
	base *incrBase
	v    *varIncr
}

func NewVar[T any](r *Reactor, value T) Var[T] {
	v := &varIncr{
		value:    value,
		modified: true,
	}
	return Var[T]{
		v:    v,
		base: newBase(r, v),
	}
}

func (v Var[T]) Set(value T) {
	v.v.set(v.base, value)
}

func (v Var[T]) Incr() Incr[T] {
	return Incr[T]{base: v.base}
}

func Const[T any](r *Reactor, value T) Incr[T] {
	return Incr[T]{base: newBase(r, constIncr{value: value})}
}

type DiffFunc[T any] func(T, T) bool

func ComparableDifferent[T comparable](x, y T) bool {
	return x != y
}

type Incr[T any] struct {
	base *incrBase
}

func Map[A, B any](a Incr[A], f func(a A) B) Incr[B] {
	return Map2(a, a.base.reactor.constNil, func(a A, _ any) B {
		return f(a)
	})
}

func Map2[A, B, C any](a Incr[A], b Incr[B], f func(a A, b B) C) Incr[C] {
	if a.base.reactor != b.base.reactor {
		panic("Tried to mix Incrs from different reactors.")
	}
	return Incr[C]{
		base: newBase(a.base.reactor, &map2Incr{
			a: a.base,
			b: b.base,
			f: func(x, y any) any {
				return f(x.(A), y.(B))
			},
		}),
	}
}

type map2Incr struct {
	a, b  *incrBase
	f     func(any, any) any
	value any
}

func (m *map2Incr) get() any {
	return m.value
}

func (m *map2Incr) activate(b *incrBase) {
	m.a.subscribe(b)
	m.b.subscribe(b)
}

func (m *map2Incr) deactivate(b *incrBase) {
	m.a.unsubscribe(b)
	m.b.unsubscribe(b)
}

func (m *map2Incr) recompute(b *incrBase) {
	last := m.value
	next := m.f(m.a.get(), m.b.get())
	if b.different(last, next) {
		m.value = next
		b.notify()
	}
}
