package incr

type Incr[T any] struct {
	base *incrBase
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

type incr interface {
	get() any
	recompute(*incrBase)
	activate(*incrBase)
	deactivate(*incrBase)
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
