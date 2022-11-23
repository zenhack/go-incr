package incr

func Map[A, B any](a Incr[A], f func(a A) B) Incr[B] {
	return Map2(a, a.base.reactor.constEmpty, func(a A, _ struct{}) B {
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
