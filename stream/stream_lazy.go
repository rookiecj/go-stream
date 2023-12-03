package stream

import (
	"errors"
	"reflect"
)

var (
	nilAnyStream *baseStream[any]
)

// Stream is lazy, the operations are not executed until terminal operationa are called.
// and it is not thread safe
type Stream[T any] interface {
	Source[T]

	Filter(filter func(T) bool) Stream[T]

	Map(mapf func(T) T) Stream[T]
	MapAny(mapf func(T) any) Stream[any]
	MapIndex(mapf func(int, T) T) Stream[T]
	MapIndexAny(mapf func(int, T) any) Stream[any]

	FlatMapConcat(f func(T) Source[T]) Stream[T]
	FlatMapConcatAny(f func(T) Source[any]) Stream[any]

	Take(n int) Stream[T]
	Skip(n int) Stream[T]

	Distinct() Stream[T]
	DistinctBy(cmp func(old T, new T) bool) Stream[T]

	ZipWith(other Source[T], zipf func(T, T) T) Stream[T]
	ZipWithAny(other Source[any], zipf func(T, any) any) Stream[any]
	ZipWithPrev(zipf func(prev T, ele T) T) Stream[T]

	Scan(init T, accumf func(acc T, ele T) T) Stream[T]
	ScanAny(init any, accumf func(acc any, ele T) any) Stream[any]

	OnEach(visit func(v T)) Stream[T]

	Collector[T]
}

// Source is source of a stream.
type Source[T any] interface {

	// Next는 element가 존재하는지 여부를 반환한다.
	Next() bool

	// Next에 의해서 확인한 element을 반환한다.
	// 호출시마다 동일한 값을 리턴한다.
	Get() T
}

// terminal operations
type Collector[T any] interface {
	ForEach(visit func(ele T))
	ForEachIndex(visit func(int, T))

	Collect() (target []T)

	Reduce(reducer func(acc T, ele T) T) (result T)
	ReduceAny(reducer func(acc any, ele T) any) (result any)
	Fold(init T, reducer func(acc T, ele T) T) (result T)
	FoldAny(init any, reducer func(acc any, ele T) any) (result any)

	Find(predicate func(T) bool) (found T, err error)
	FindOr(predicate func(ele T) bool, defvalue T) T
	FindIndex(predicate func(T) bool) int
	FindLast(predicate func(T) bool) (found T, err error)
	FindLastOr(predicate func(T) bool, defvalue T) (found T)
	FindLastIndex(predicate func(T) bool) (found int)
}

type baseStream[T any] struct {
	idx   int // full index
	limit int // -1, unlimited or unknown
	next  func() bool
	get   func() any
}

// source operations
func (s *baseStream[T]) Next() bool {
	if s == nil {
		return false
	}
	if s.next() {
		return true
	}
	return false
}

func (s *baseStream[T]) Get() (result T) {
	if s == nil {
		return
	}
	return s.get().(T)
}

//
// stream operations
//

func (s *baseStream[T]) Filter(filter func(T) bool) Stream[T] {
	if s == nil {
		return s
	}
	downstream := new(baseStream[T])
	downstream.idx = -1
	downstream.next = func() bool {
		for s.next() {
			downstream.idx++
			if filter(s.get().(T)) {
				return true
			}
		}
		return false
	}

	downstream.get = func() any {
		return s.get()
	}
	return downstream
}

//
//func (stream *baseStream[T]) FilterAny(filter func(any) bool) Stream[any] {
//	downstream := new(baseStream[any])
//	downstream.idx = -1
//	downstream.next = func() bool {
//		for stream.next() {
//			downstream.idx++
//			if filter(stream.get()) {
//				return true
//			}
//		}
//		return true
//	}
//
//	downstream.get = func() any {
//		return stream.get()
//	}
//	return downstream
//}

func (s *baseStream[T]) Map(mapf func(T) T) Stream[T] {
	if s == nil {
		return s
	}

	downstream := new(baseStream[T])
	downstream.idx = -1
	downstream.next = func() bool {
		if s.next() {
			downstream.idx++
			return true
		}
		return false
	}

	downstream.get = func() any {
		return mapf(s.get().(T))
	}
	return downstream
}

func (s *baseStream[T]) MapAny(mapf func(T) any) Stream[any] {
	if s == nil {
		return nilAnyStream
	}

	// T -> any
	downstream := new(baseStream[any])
	downstream.idx = -1
	downstream.next = func() bool {
		if s.next() {
			downstream.idx++
			return true
		}
		return false
	}

	downstream.get = func() any {
		return mapf(s.get().(T))
	}
	return downstream
}

// MapIndex returns a stream consisting of the results of applying the given function to the elements of this stream.
func (s *baseStream[T]) MapIndex(mapf func(int, T) T) Stream[T] {
	if s == nil {
		return s
	}

	downstream := new(baseStream[T])
	downstream.idx = -1
	downstream.next = func() bool {
		if s.next() {
			downstream.idx++
			return true
		}
		return false
	}
	downstream.get = func() any {
		return mapf(downstream.idx, s.get().(T))
	}
	return downstream
}

// MapIndexAny returns a stream consisting of the results of applying the given function to the elements of this stream.
func (s *baseStream[T]) MapIndexAny(mapf func(int, T) any) Stream[any] {
	if s == nil {
		return nilAnyStream
	}

	downstream := new(baseStream[any])
	downstream.idx = -1
	downstream.next = func() bool {
		if s.next() {
			downstream.idx++
			return true
		}
		return false
	}
	downstream.get = func() any {
		return mapf(downstream.idx, s.get().(T))
	}
	return downstream
}

// FlatMapConcat returns a stream consisting of the results of
// replacing each element of this stream with the contents of
// a mapped stream produced by applying the provided mapping function to each element.
func (s *baseStream[T]) FlatMapConcat(fmap func(ele T) Source[T]) Stream[T] {
	if s == nil {
		return s
	}

	downstream := new(baseStream[T])
	downstream.idx = -1
	var cursource Source[T]
	downstream.next = func() bool {
	loop:
		if cursource != nil && cursource.Next() {
			downstream.idx++
			return true
		}
		for s.next() {
			cursource = fmap(s.get().(T))
			goto loop
		}
		return false
	}

	downstream.get = func() any {
		return cursource.Get()
	}
	return downstream
}

func (s *baseStream[T]) FlatMapConcatAny(fmap func(ele T) Source[any]) Stream[any] {
	if s == nil {
		return nilAnyStream
	}

	downstream := new(baseStream[any])
	downstream.idx = -1
	var astream Source[any]
	downstream.next = func() bool {
	loop:
		if astream != nil && astream.Next() {
			downstream.idx++
			return true
		}
		for s.next() {
			astream = fmap(s.get().(T))
			goto loop
		}
		return false
	}

	downstream.get = func() any {
		return astream.Get()
	}
	return downstream
}

// Take returns a stream consisting of the first n elements of this stream.
func (s *baseStream[T]) Take(n int) Stream[T] {
	if s == nil {
		return s
	}

	downstream := new(baseStream[T])
	downstream.idx = -1
	downstream.next = func() bool {
		if downstream.idx+1 == n {
			return false
		}
		downstream.idx++
		return s.next()
	}
	downstream.get = func() any {
		return s.get()
	}
	return downstream
}

// Skip returns a stream consisting of the remaining elements of this stream after discarding the first n elements of the stream.
func (s *baseStream[T]) Skip(n int) Stream[T] {
	if s == nil {
		return s
	}

	downstream := new(baseStream[T])
	downstream.idx = -1
	downstream.next = func() bool {
		for downstream.idx+1 < n {
			if !s.next() {
				return false
			}
			downstream.idx++
		}
		return s.next()
	}
	downstream.get = func() any {
		return s.get()
	}
	return downstream
}

// Distinct returns a stream consisting of the subsequent distinct elements of this stream.
// [a, a, b, c, a] => [a, b, c, a]
func (s *baseStream[T]) Distinct() Stream[T] {
	if s == nil {
		return s
	}

	return s.DistinctBy(func(old, v T) bool {
		return reflect.DeepEqual(old, v)
	})
}

// DistinctBy returns a stream consisting of the subsequent distinct elements of this stream.
// [a(0), a(1), b, c, a] => [a(0), b, c, a]
// cmd is a function to compare two elements, return true if they are equal.
// With the first element, nil is given to old.
func (s *baseStream[T]) DistinctBy(cmp func(old, new T) bool) Stream[T] {
	if s == nil {
		return s
	}

	downstream := new(baseStream[T])
	var old T
	downstream.idx = -1
	downstream.next = func() bool {
		for s.next() {
			downstream.idx++
			v := s.get().(T)
			if !cmp(old, v) {
				old = v
				return true
			}
			old = v
		}
		return false
	}
	downstream.get = func() any {
		return old
	}
	return downstream
}

// ZipWith returns a stream consisting of appling the given function to the elements of this stream and another stream.
func (s *baseStream[T]) ZipWith(other Source[T], zipf func(T, T) T) Stream[T] {
	if s == nil {
		return s
	}

	downstream := new(baseStream[T])
	downstream.idx = -1
	downstream.next = func() bool {
		if s.next() && other.Next() {
			downstream.idx++
			return true
		}
		return false
	}
	downstream.get = func() any {
		return zipf(s.get().(T), other.Get())
	}
	return downstream
}

// ZipWithAny returns a stream consisting of appling the given function to the elements of this stream and another stream.
func (s *baseStream[T]) ZipWithAny(other Source[any], zipf func(T, any) any) Stream[any] {
	if s == nil {
		return nilAnyStream
	}

	downstream := new(baseStream[any])
	downstream.idx = -1
	downstream.next = func() bool {
		if s.next() && other.Next() {
			downstream.idx++
			return true
		}
		return false
	}
	downstream.get = func() any {
		return zipf(s.get().(T), other.Get())
	}
	return downstream
}

// ZipWithPrev returns a stream consisting of applying the given function to
// the elements of this stream and its previous element.
// [a, b, c, d] => [f(nil, a), f(a, b), f(b, c), f(c, d)]
func (s *baseStream[T]) ZipWithPrev(zipf func(prev T, ele T) T) Stream[T] {
	if s == nil {
		return s
	}

	downstream := new(baseStream[T])
	downstream.idx = -1
	var prev T
	downstream.next = func() bool {
		if s.next() {
			downstream.idx++
			return true
		}
		return false
	}
	downstream.get = func() any {
		v := s.get().(T)
		result := zipf(prev, v)
		prev = v
		return result
	}
	return downstream
}

// Scan returns a stream consisting of the accumlated results of applying the given function to the elements of this stream.
//
//	with source=[1, 2, 3, 4], init=0 and func(acc, i) { acc + i } produces [1, 3, 6, 10]
func (s *baseStream[T]) Scan(init T, accumf func(acc T, ele T) T) Stream[T] {
	if s == nil {
		return s
	}

	downstream := new(baseStream[T])
	downstream.idx = -1
	acc := init
	downstream.next = func() bool {
		if s.next() {
			downstream.idx++
			return true
		}
		return false
	}
	downstream.get = func() any {
		acc = accumf(acc, s.get().(T))
		return acc
	}
	return downstream
}

func (s *baseStream[T]) ScanAny(init any, accumf func(acc any, ele T) any) Stream[any] {
	if s == nil {
		return nilAnyStream
	}

	downstream := new(baseStream[any])
	downstream.idx = -1
	acc := init
	downstream.next = func() bool {
		if s.next() {
			downstream.idx++
			return true
		}
		return false
	}
	downstream.get = func() any {
		acc = accumf(acc, s.get().(T))
		return acc
	}
	return downstream
}

// OnEach returns a stream do nothing but visit each element of this stream.
func (s *baseStream[T]) OnEach(visit func(v T)) Stream[T] {
	if s == nil {
		return s
	}

	downstream := new(baseStream[T])
	downstream.idx = -1
	downstream.next = func() bool {
		if s.next() {
			downstream.idx++
			return true
		}
		return false
	}
	downstream.get = func() any {
		v := s.get()
		visit(v.(T))
		return v
	}
	return downstream
}

//
// terminal operations
//

func (s *baseStream[T]) ForEach(visit func(T)) {
	if s == nil {
		return
	}

	for s.next() {
		visit(s.get().(T))
	}
}

// ForEachIndex performs an action for each element of this stream.
func (s *baseStream[T]) ForEachIndex(visit func(int, T)) {
	if s == nil {
		return
	}

	idx := -1
	for s.next() {
		idx++
		visit(idx, s.get().(T))
	}
}

func (s *baseStream[T]) Collect() (target []T) {
	if s == nil {
		return
	}

	for s.next() {
		v := s.get()
		target = append(target, v.(T))
	}
	return
}

// Reduce performs a reduction on the elements of this stream,
// using the provided identity value and an associative accumulation function,
// and returns the reduced value.
// [a, b, c, d] => f(f(f(a, b), c), d)
// if the stream is empty returns nil
func (s *baseStream[T]) Reduce(reducer func(acc T, ele T) T) (result T) {
	if s == nil {
		return
	}
	if s.next() {
		result = s.get().(T)
	} else {
		// nothing to return
		return
	}
	for s.next() {
		result = reducer(result, s.get().(T))
	}
	return result
}

func (s *baseStream[T]) ReduceAny(reducer func(acc any, ele T) any) (result any) {
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
		result = reducer(result, v.(T))
	}
	return result
}

// Fold performs a reduction on the elements of this stream, using the provided identity value
// and an associative accumulation function, and returns the reduced value.
func (s *baseStream[T]) Fold(init T, reducer func(acc T, ele T) T) (result T) {
	if s == nil {
		return init
	}

	result = init
	for s.next() {
		v := s.get()
		result = reducer(result, v.(T))
	}
	return result
}

func (s *baseStream[T]) FoldAny(init any, reducer func(acc any, ele T) any) (result any) {
	if s == nil {
		return init
	}

	result = init
	for s.next() {
		v := s.get()
		result = reducer(result, v.(T))
	}
	return result
}

// Find returns the first element of this stream matching the given predicate
func (s *baseStream[T]) Find(predicate func(T) bool) (found T, err error) {
	if s == nil {
		err = errors.New("upstream is nil")
		return
	}

	for s.next() {
		v := s.get().(T)
		if predicate(v) {
			return v, nil
		}
	}
	err = nil
	return
}

// FindOr returns the first element of this stream matching the given predicate, or defvalue if no such element exists.
func (s *baseStream[T]) FindOr(predicate func(ele T) bool, defvalue T) T {
	if s == nil {
		return defvalue
	}

	for s.next() {
		v := s.get().(T)
		if predicate(v) {
			return v
		}
	}
	return defvalue
}

// FindIndex returns the index of the first element of this down stream, not source, matching the given predicate,
// or -1 if no such element exists.
func (s *baseStream[T]) FindIndex(predicate func(T) bool) int {
	if s == nil {
		return -1
	}

	idx := -1
	for s.next() {
		idx++
		v := s.get()
		if predicate(v.(T)) {
			return idx
		}
	}
	return -1
}

// FindLast returns the last element of this stream matching the given predicate, otherwise -1.
func (s *baseStream[T]) FindLast(predicate func(T) bool) (found T, err error) {
	if s == nil {
		err = errors.New("upstream is nil")
		return
	}

	for s.next() {
		v := s.get().(T)
		if predicate(v) {
			found = v
		}
	}
	return found, nil
}

// FindLastOr returns the last element of this stream matching the given predicate,
// or defvalue if no such element exists.
func (s *baseStream[T]) FindLastOr(predicate func(T) bool, defvalue T) (found T) {
	if s == nil {
		return defvalue
	}

	found = defvalue
	for s.next() {
		v := s.get().(T)
		if predicate(v) {
			found = v
		}
	}
	return
}

// FindLastIndex returns the index of the last element of this stream matching the given predicate, otherwise -1.
func (s *baseStream[T]) FindLastIndex(predicate func(T) bool) (found int) {
	if s == nil {
		return -1
	}

	idx := -1
	found = idx
	for s.next() {
		idx++
		v := s.get()
		if predicate(v.(T)) {
			found = idx
		}
	}
	return found
}
