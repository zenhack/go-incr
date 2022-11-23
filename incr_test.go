package incr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstWatcher(t *testing.T) {
	// Watchers should trigger once on constants
	r := NewReactor()
	seen := 0
	Const(r, 4).Observe().Watch(func(value int) bool {
		seen = value
		return true
	})
	r.Stabilize()
	assert.Equal(t, seen, 4)
}

func TestVarWatcher(t *testing.T) {
	// Watchers should trigger when vars change
	r := NewReactor()
	v := NewVar(r, 0)
	last := 0
	v.Observe().Watch(func(x int) bool {
		last = x
		return true
	})
	v.Set(2)
	r.Stabilize()
	assert.Equal(t, last, 2)
	v.Set(4)
	r.Stabilize()
	assert.Equal(t, last, 4)
}

func TestDisableWatcher(t *testing.T) {
	// Watchers should be disabled when they return false.
	r := NewReactor()
	last := 0
	v := NewVar(r, last)
	v.Observe().Watch(func(x int) bool {
		last = x
		return false
	})
	v.Set(1)
	r.Stabilize()
	assert.Equal(t, last, 1)
	v.Set(2)
	r.Stabilize()
	assert.Equal(t, last, 1)
}

func TestMapObservers(t *testing.T) {
	// Map should trigger observers.
	r := NewReactor()
	last := 0
	v := NewVar(r, last)
	Map(v.Incr(), func(x int) int { return x + 1 }).Observe().Watch(func(x int) bool {
		last = x
		return true
	})
	v.Set(1)
	r.Stabilize()
	assert.Equal(t, last, 2)
	v.Set(7)
	r.Stabilize()
	assert.Equal(t, last, 8)
}
