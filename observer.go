package incr

type Observer[T any] struct {
	base *incrBase
	obs  *observeIncr
}

func Observe[T any](parent Incr[T]) Observer[T] {
	obs := &observeIncr{
		parent: parent.base,
	}
	base := newBase(parent.base.reactor, obs)
	parent.base.subscribe(base)
	return Observer[T]{
		base: base,
		obs:  obs,
	}
}

func (o Observer[T]) Release() {
	o.obs.parent.unsubscribe(o.base)
}

func (o Observer[T]) Watch(f func(T) bool) {
	o.obs.watch(func(v any) bool {
		return f(v.(T))
	})
}

func (o Observer[T]) Get() T {
	return o.base.get().(T)
}

type observeIncr struct {
	parent   *incrBase
	watchers []func(any) bool
}

func (o *observeIncr) get() any {
	return o.parent.get()
}

func (o *observeIncr) recompute(b *incrBase) {
	v := o.get()
	j := 0
	for i := 0; i < len(o.watchers); i++ {
		f := o.watchers[i]
		if f(v) {
			o.watchers[j] = f
			j++
		}
	}
	o.watchers = o.watchers[:j]
	b.notify()
}

func (o *observeIncr) watch(f func(any) bool) {
	o.watchers = append(o.watchers, f)
}

func (*observeIncr) activate(*incrBase)   {}
func (*observeIncr) deactivate(*incrBase) {}
