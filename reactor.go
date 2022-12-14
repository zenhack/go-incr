package incr

import "container/heap"

type Reactor struct {
	allocHeight uint64
	dirty       incrHeap
	constEmpty  Incr[struct{}] // Used as a dummy value in Map()
}

func NewReactor() *Reactor {
	r := &Reactor{}
	r.constEmpty = Const(r, struct{}{})
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
