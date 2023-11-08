package stream

// Stream uses iterator pattern
// which means it is lazy, and the operations are not executed until terminal operation is called
// Stream is not thread safe
type Stream[T any] struct {
	slice []T
	idx   int // full index
	next  func() bool
	get   func() T
}

// ToStream converts slice to Stream
func ToStream[T any](arr []T) *Stream[any] {
	stream := Stream[any]{
		slice: func() []any {
			var result []any
			for _, v := range arr {
				result = append(result, v)
			}
			return result
		}(),
		idx: -1,
	}

	stream.next = func() bool {
		if stream.idx+1 == len(stream.slice) {
			return false
		}
		stream.idx++
		return true
	}

	stream.get = func() any {
		return stream.slice[stream.idx]
	}
	return &stream
}

//
// intermediate operations
//

// Filter returns a stream consisting of the elements of this stream that match the given predicate.
func (s *Stream[any]) Filter(f func(any) bool) *Stream[any] {
	var stream Stream[any]

	stream.next = func() bool {
		for s.next() {
			if f(s.get()) {
				return true
			}
		}
		return false
	}

	stream.get = func() any {
		return s.get()
	}

	return &stream
}

// Map returns a stream consisting of the results of applying the given function to the elements of this stream.
func (s *Stream[any]) Map(f func(any) any) *Stream[any] {
	var stream Stream[any]
	stream.next = func() bool {
		return s.next()
	}
	stream.get = func() any {
		return f(s.get())
	}
	return &stream
}

// MapIndex returns a stream consisting of the results of applying the given function to the elements of this stream.
func (s *Stream[any]) MapIndex(f func(int, any) any) *Stream[any] {
	var stream Stream[any]
	stream.idx = 0
	stream.next = func() bool {
		return s.next()
	}
	stream.get = func() any {
		idx := stream.idx
		stream.idx++
		return f(idx, s.get())
	}
	return &stream
}

// FlatMapConcat returns a stream consisting of the results of
// replacing each element of this stream with the contents of
// a mapped stream produced by applying the provided mapping function to each element.
func (s *Stream[any]) FlatMapConcat(f func(any) *Stream[any]) *Stream[any] {
	var stream Stream[any]
	var astream *Stream[any]
	stream.next = func() bool {
		if astream != nil && astream.next() {
			return true
		} else {
			astream = nil
			for astream == nil {
				if !s.next() {
					return false
				}
				astream = f(s.get())
				if astream.next() {
					return true
				}
				astream = nil
			}
			return false
		}
	}

	stream.get = func() any {
		return astream.get()
	}
	return &stream
}

// Take returns a stream consisting of the first n elements of this stream.
func (s *Stream[any]) Take(n int) *Stream[any] {
	var stream Stream[any]
	stream.idx = 0
	stream.next = func() bool {
		if stream.idx == n {
			return false
		}
		stream.idx++
		return s.next()
	}
	stream.get = func() any {
		return s.get()
	}
	return &stream
}

// Skip returns a stream consisting of the remaining elements of this stream after discarding the first n elements of the stream.
func (s *Stream[any]) Skip(n int) *Stream[any] {
	var stream Stream[any]
	stream.idx = 0
	stream.next = func() bool {
		for stream.idx < n {
			if !s.next() {
				return false
			}
			stream.idx++
		}
		return s.next()
	}
	stream.get = func() any {
		return s.get()
	}
	return &stream
}

//
// terminal operations
//

// ForEach performs an action for each element of this stream.
func (s *Stream[any]) ForEach(f func(any)) {
	for s.next() {
		f(s.get())
	}
}

// ForEachIndex performs an action for each element of this stream.
func (s *Stream[any]) ForEachIndex(f func(int, any)) {
	idx := 0
	for s.next() {
		f(idx, s.get())
		idx++
	}
}

// Reduce performs a reduction on the elements of this stream, using the provided identity value and an associative accumulation function, and returns the reduced value.
func (s *Stream[any]) Reduce(f func(any, any) any) any {
	var result any
	for s.next() {
		v := s.get()
		result = f(result, v)
	}
	return result
}

// FindOr returns the first element of this stream matching the given predicate, or defvalue if no such element exists.
func (s *Stream[any]) FindOr(f func(any) bool, defvalue any) any {
	for s.next() {
		v := s.get()
		if f(v) {
			return v
		}
	}
	return defvalue
}

// CollectAs returns a slice containing the elements of this stream.
func CollectAs[T any](s *Stream[any], target []T) []T {
	for s.next() {
		v := s.get()
		target = append(target, v.(T))
	}
	return target
}
