package stream

// CollectAs returns a slice containing the elements of this stream with slightly more type safety
func CollectAs[T any](s *Stream[any], target []T) []T {
	if s == nil {
		return nil
	}

	for s.next() {
		v := s.get().(T)
		target = append(target, v)
	}
	return target
}

// ForEachAs performs an action for each element of this stream.
func ForEachAs[T any](s *Stream[any], f func(T)) {
	if s == nil {
		return
	}

	for s.next() {
		f(s.get().(T))
	}
}

// ForEachIndex performs an action for each element of this stream.
func ForEachIndex[T any](s *Stream[any], f func(int, T)) {
	if s == nil {
		return
	}

	idx := 0
	for s.next() {
		f(idx, s.get().(T))
		idx++
	}
}

// ReduceAs performs a reduction on the elements of this stream,
// using the provided identity value and an associative accumulation function,
// and returns the reduced value.
func ReduceAs[T any](s *Stream[any], accumf func(T, T) T) T {
	var result T
	if s == nil {
		return result
	}

	for s.next() {
		v := s.get().(T)
		result = accumf(result, v)
	}
	return result
}

func FindOrAs[T any](s *Stream[any], f func(T) bool, defvalue T) T {
	if s == nil {
		return defvalue
	}

	for s.next() {
		v := s.get().(T)
		if f(v) {
			return v
		}
	}
	return defvalue
}
