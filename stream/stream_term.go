package stream

// CollectAs returns a slice containing the elements of this stream with slightly more type safety
func CollectAs[T any](s Source[T], target []T) []T {
	if s == nil {
		return nil
	}

	for s.Next() {
		v := s.Get()
		target = append(target, v)
	}
	return target
}

// ForEachAs performs an action for each element of this stream.
func ForEachAs[T any](s Source[T], f func(T)) {
	if s == nil {
		return
	}

	for s.Next() {
		f(s.Get())
	}
}

// ForEachIndex performs an action for each element of this stream.
func ForEachIndex[T any](s Source[T], visit func(int, T)) {
	if s == nil {
		return
	}

	idx := -1
	for s.Next() {
		idx++
		visit(idx, s.Get())
	}
}

// ReduceAs performs a reduction on the elements of this stream,
// using the provided identity value and an associative accumulation function,
// and returns the reduced value.
func ReduceAs[T any](s Source[T], accumf func(T, T) T) T {
	var result T
	if s == nil {
		return result
	}

	for s.Next() {
		v := s.Get()
		result = accumf(result, v)
	}
	return result
}

func FindOrAs[T any](s Source[T], predicate func(T) bool, defvalue T) T {
	if s == nil {
		return defvalue
	}

	for s.Next() {
		v := s.Get()
		if predicate(v) {
			return v
		}
	}
	return defvalue
}
