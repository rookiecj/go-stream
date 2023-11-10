package stream

import "reflect"

// Stream uses iterator pattern
// which means it is lazy, and the operations are not executed until terminal operation is called
// Stream is not thread safe
type Stream[T any] struct {
	idx  int // full index
	next func() bool
	get  func() T
}

//
// intermediate operations
//

// Filter returns a stream consisting of the elements of this stream that match the given predicate.
func (s *Stream[any]) Filter(f func(any) bool) *Stream[any] {
	if s == nil {
		return s
	}

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
	if s == nil {
		return s
	}

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
	if s == nil {
		return s
	}

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
	if s == nil {
		return s
	}

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
	if s == nil {
		return s
	}

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
	if s == nil {
		return s
	}

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

// Distinct returns a stream consisting of the subsequent distinct elements of this stream.
// [a, a, b, c, a] => [a, b, c, a]
func (s *Stream[any]) Distinct() *Stream[any] {
	if s == nil {
		return s
	}

	return s.DistinctBy(func(old, v any) bool {
		return reflect.DeepEqual(old, v)
	})
}

// DistinctBy returns a stream consisting of the subsequent distinct elements of this stream.
// [a, a, b, c, a] => [a, b, c, a]
// cmd is a function to compare two elements, return true if they are equal
func (s *Stream[any]) DistinctBy(cmp func(any, any) bool) *Stream[any] {
	if s == nil {
		return s
	}

	var stream Stream[any]
	var old any
	var first bool = true
	stream.next = func() bool {
		for s.next() {
			v := s.get()
			if first || !cmp(old, v) {
				old = v
				first = false
				return true
			}
			old = v
		}
		return false
	}
	stream.get = func() any {
		return s.get()
	}
	return &stream
}

// // Zip returns a stream consisting of the elements of this stream and another stream.
// // returns a stream of Pair[any, any]
// func (s *Stream[any]) Zip(other *Stream[any]) *Stream[any] {
// 	var stream Stream[any]
// 	stream.next = func() bool {
// 		return s.next() && other.next()
// 	}
// 	stream.get = func() any {
// 		return []any{s.get(), other.get()}
// 	}
// 	return &stream
// }

// ZipWith returns a stream consisting of appling the given function to the elements of this stream and another stream.
func (s *Stream[any]) ZipWith(other *Stream[any], f func(any, any) any) *Stream[any] {
	if s == nil {
		return s
	}

	var stream Stream[any]
	stream.next = func() bool {
		return s.next() && other.next()
	}
	stream.get = func() any {
		return f(s.get(), other.get())
	}
	return &stream
}

// Scan returns a stream consisting of the accumlated results of applying the given function to the elements of this stream.
//
//	with [1, 2, 3, 4], init 0, accumf func(acc, i) { acc + i } produces [1, 3, 6, 10]
func (s *Stream[any]) Scan(init any, accumf func(acc any, ele any) any) *Stream[any] {
	if s == nil {
		return s
	}

	var stream Stream[any]
	acc := init
	stream.next = func() bool {
		return s.next()
	}
	stream.get = func() any {
		acc = accumf(acc, s.get())
		return acc
	}
	return &stream
}

//
// terminal operations
//

// ForEach performs an action for each element of this stream.
func (s *Stream[any]) ForEach(f func(any)) {
	if s == nil {
		return
	}

	for s.next() {
		f(s.get())
	}
}

// ForEachIndex performs an action for each element of this stream.
func (s *Stream[any]) ForEachIndex(f func(int, any)) {
	if s == nil {
		return
	}

	idx := 0
	for s.next() {
		f(idx, s.get())
		idx++
	}
}

func (s *Stream[any]) Collect() (target []any) {
	if s == nil {
		return
	}

	for s.next() {
		v := s.get()
		target = append(target, v)
	}
	return
}

// Reduce performs a reduction on the elements of this stream, using the provided identity value and an associative accumulation function, and returns the reduced value.
func (s *Stream[any]) Reduce(f func(any, any) any) any {
	var result any
	if s == nil {
		return result
	}

	for s.next() {
		v := s.get()
		result = f(result, v)
	}
	return result
}

// FindOr returns the first element of this stream matching the given predicate, or defvalue if no such element exists.
func (s *Stream[any]) FindOr(f func(any) bool, defvalue any) any {
	if s == nil {
		return defvalue
	}

	for s.next() {
		v := s.get()
		if f(v) {
			return v
		}
	}
	return defvalue
}
