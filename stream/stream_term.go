package stream

// CollectAs returns a slice containing the elements of this stream with slightly more type safety
func CollectAs[T any](s Source[any]) (target []T) {
	if s == nil {
		return []T{}
	}

	target = []T{}
	for s.Next() {
		v := s.Get()
		target = append(target, v.(T))
	}
	return target
}

// CollectAsSafe returns a slice containing the elements of this stream,
// and error handler will be installed to handle any error
func CollectAsSafe[T any](s Source[any], onerror func()) (target []T) {
	if s == nil {
		return []T{}
	}
	if onerror != nil {
		defer onerror()
	}

	return CollectAs[T](s)
}

// CollectTo returns a slice containing the elements of this stream into target slice
func CollectTo[T any](s Source[any], target []T) []T {
	if s == nil {
		return target
	}

	for s.Next() {
		v := s.Get()
		target = append(target, v.(T))
	}
	return target
}

// CollectToSafe returns a slice containing the elements of this stream into target slice,
// and error handler will be installed to handle any error
func CollectToSafe[T any](s Source[any], target []T, onerror func()) []T {
	if s == nil {
		return target
	}

	if onerror != nil {
		defer onerror()
	}

	return CollectTo(s, target)
}

// ForEachAs performs an action for each element of this stream.
func ForEachAs[T any](s Source[any], f func(T)) {
	if s == nil {
		return
	}

	for s.Next() {
		f(s.Get().(T))
	}
}

// ForEachIndex performs an action for each element of this stream.
func ForEachIndex[T any](s Source[any], visit func(int, T)) {
	if s == nil {
		return
	}

	idx := 0
	for s.Next() {
		visit(idx, s.Get().(T))
		idx++
	}
}

// ReduceAs performs a reduction on the elements of this stream,
// using the provided identity value and an associative accumulation function,
// and returns the reduced value.
func ReduceAs[T any](s Source[any], accumf func(T, T) T) T {
	var result T
	if s == nil {
		return result
	}

	for s.Next() {
		v := s.Get()
		result = accumf(result, v.(T))
	}
	return result
}

func FindOrAs[T any](s Source[any], predicate func(T) bool, defvalue T) T {
	if s == nil {
		return defvalue
	}

	for s.Next() {
		v := s.Get().(T)
		if predicate(v) {
			return v
		}
	}
	return defvalue
}
