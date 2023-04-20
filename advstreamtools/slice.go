package advstreamtools

// functions for manipulating slices in common ways. note the rare
// appearance of the pointer-to-slice!

// advance moves the slice forward the given number. As a special case, if
// n is equal to the entire slice, the slice is trimmed with [:0] rather
// than [n:], which has the effect of making the slice reusable for other
// tasks without a reallocation.
//
// n must not exceed the slice or this will just go ahead and panic.
func advance[T any](s *[]T, n int) {
	if n == 0 {
		return
	}
	if len(*s) == n {
		*s = (*s)[:0]
		return
	}
	*s = (*s)[n:]
}

// move will take as many values out of src as possible, copy them to dst,
// and then advance src by the amount consumed, returning how many values
// it so copyied.
func move[T any](dst []T, src *[]T) int {
	n := copy(dst, *src)
	advance(src, n)
	return n
}

// pop pulls a value off the front of the slice and advances the
// slice. Note that it does no length checking.
func pop[T any](dst *[]T) T {
	val := (*dst)[0]
	advance(dst, 1)
	return val
}
