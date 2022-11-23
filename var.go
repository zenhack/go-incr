package incr

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

func (v Var[T]) Observe() Observer[T] {
	return v.Incr().Observe()
}

func (v Var[T]) Incr() Incr[T] {
	return Incr[T]{base: v.base}
}

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
