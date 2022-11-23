package incr

func Const[T any](r *Reactor, value T) Incr[T] {
	return Incr[T]{base: newBase(r, constIncr{value: value})}
}

type constIncr struct {
	value any
}

func (c constIncr) get() any              { return c.value }
func (c constIncr) recompute(b *incrBase) { b.notify() }
func (c constIncr) activate(b *incrBase)  { b.setDirty() }
func (constIncr) deactivate(*incrBase)    {}
