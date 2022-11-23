package incr

// Implements heap.Interface on top of a slice of *incrBase.
type incrHeap []*incrBase

func (h *incrHeap) Len() int {
	return len(*h)
}

func (h *incrHeap) Less(i, j int) bool {
	s := *h
	return s[i].scheduledHeight < s[j].scheduledHeight
}

func (h *incrHeap) Pop() any {
	s := *h
	ret := s[len(s)-1]
	*h = s[:len(s)-1]
	return ret
}

func (h *incrHeap) Swap(i, j int) {
	s := *h
	tmp := s[i]
	s[i] = s[j]
	s[j] = tmp
}

func (h *incrHeap) Push(x any) {
	*h = append(*h, x.(*incrBase))
}
