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
func (s *Stream[any]) Filter(filter func(any) bool) *Stream[any] {
	if s == nil {
		return s
	}

	var stream Stream[any]
	stream.next = func() bool {
		for s.next() {
			if filter(s.get()) {
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
func (s *Stream[any]) Map(mapf func(any) any) *Stream[any] {
	if s == nil {
		return s
	}

	var stream Stream[any]
	stream.next = func() bool {
		return s.next()
	}
	stream.get = func() any {
		return mapf(s.get())
	}
	return &stream
}

// MapIndex returns a stream consisting of the results of applying the given function to the elements of this stream.
func (s *Stream[any]) MapIndex(mapf func(int, any) any) *Stream[any] {
	if s == nil {
		return s
	}

	var stream Stream[any]

	stream.idx = -1
	stream.next = func() bool {
		if s.next() {
			stream.idx++
			return true
		}
		return false
	}
	stream.get = func() any {
		return mapf(stream.idx, s.get())
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
	stream.idx = -1
	stream.next = func() bool {
		if stream.idx+1 == n {
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
	stream.idx = -1
	stream.next = func() bool {
		for stream.idx+1 < n {
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

// ZipWithPrev returns a stream consisting of applying the given function to
// the elements of this stream and its previous element.
// [a, b, c, d] => [f(nil, a), f(a, b), f(b, c), f(c, d)]
func (s *Stream[any]) ZipWithPrev(f func(prev any, ele any) any) *Stream[any] {
	if s == nil {
		return s
	}

	var stream Stream[any]
	var prev any
	stream.next = func() bool {
		return s.next()
	}
	stream.get = func() any {
		v := s.get()
		result := f(prev, v)
		prev = v
		return result
	}
	return &stream
}

// Scan returns a stream consisting of the accumlated results of applying the given function to the elements of this stream.
//
//	with source=[1, 2, 3, 4], init=0 and func(acc, i) { acc + i } produces [1, 3, 6, 10]
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

// OnEach returns a stream do nothing but visit each element of this stream.
func (s *Stream[any]) OnEach(visit func(v any)) *Stream[any] {
	if s == nil {
		return s
	}

	var stream Stream[any]
	stream.next = func() bool {
		return s.next()
	}
	stream.get = func() any {
		v := s.get()
		visit(v)
		return v
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

// Reduce performs a reduction on the elements of this stream,
// using the provided identity value and an associative accumulation function,
// and returns the reduced value.
// [a, b, c, d] => f(f(f(a, b), c), d)
// if the stream is empty returns nil
func (s *Stream[any]) Reduce(reducer func(acc any, ele any) any) (result any) {
	if s == nil {
		return
	}
	if s.next() {
		result = s.get()
	} else {
		// nothing to return
		return
	}
	for s.next() {
		v := s.get()
		result = reducer(result, v)
	}
	return result
}

// Fold performs a reduction on the elements of this stream, using the provided identity value
// and an associative accumulation function, and returns the reduced value.
func (s *Stream[any]) Fold(init any, reducer func(acc any, ele any) any) (result any) {
	if s == nil {
		return init
	}

	result = init
	for s.next() {
		v := s.get()
		result = reducer(result, v)
	}
	return result
}

// Find returns the first element of this stream matching the given predicate
func (s *Stream[any]) Find(predicate func(any) bool) (found any) {
	return s.FindOr(predicate, found)
}

// FindOr returns the first element of this stream matching the given predicate, or defvalue if no such element exists.
func (s *Stream[any]) FindOr(predicate func(any) bool, defvalue any) any {
	if s == nil {
		return defvalue
	}

	for s.next() {
		v := s.get()
		if predicate(v) {
			return v
		}
	}
	return defvalue
}

// FindIndex returns the index of the first element of this stream matching the given predicate,
// or -1 if no such element exists.
func (s *Stream[any]) FindIndex(predicate func(any) bool) int {
	if s == nil {
		return -1
	}

	idx := -1
	for s.next() {
		v := s.get()
		idx++
		if predicate(v) {
			return idx
		}
	}
	return -1
}

// FindLast returns the last element of this stream matching the given predicate
func (s *Stream[any]) FindLast(predicate func(any) bool) (found any) {
	if s == nil {
		return
	}

	for s.next() {
		v := s.get()
		if predicate(v) {
			found = v
		}
	}
	return
}

// FindLastOr returns the last element of this stream matching the given predicate,
// or defvalue if no such element exists.
func (s *Stream[any]) FindLastOr(predicate func(any) bool, defvalue any) (found any) {
	if s == nil {
		return defvalue
	}

	found = defvalue
	for s.next() {
		v := s.get()
		if predicate(v) {
			found = v
		}
	}
	return
}

// FindLastIndex returns the index of the last element of this stream matching the given predicate.
func (s *Stream[any]) FindLastIndex(predicate func(any) bool) (found int) {
	if s == nil {
		return -1
	}

	idx := -1
	found = idx
	for s.next() {
		v := s.get()
		idx++
		if predicate(v) {
			found = idx
		}
	}
	return found
}
