package incr

type DiffFunc[T any] func(T, T) bool

func ComparableDifferent[T comparable](x, y T) bool {
	return x != y
}
