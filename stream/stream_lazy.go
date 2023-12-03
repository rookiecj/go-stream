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

	// Filter returns a stream consisting of the elements of this stream that match the given predicate.
	Filter(filter func(T) bool) Stream[T]

	// Map returns a stream consisting of the results of applying the given function to the elements of this stream.
	Map(mapf func(T) T) Stream[T]
	// MapAny returns a stream consisting of the results of applying the given function to the elements of this stream.
	// the function return any type.
	MapAny(mapf func(T) any) Stream[any]
	// MapIndex returns a stream consisting of the results of applying the given function to the elements of this stream.
	// the function is given the index of the element.
	MapIndex(mapf func(int, T) T) Stream[T]
	// MapIndexAny returns a stream consisting of the results of applying the given function to the elements of this stream.
	// the function return any type and is given the index of the element.
	MapIndexAny(mapf func(int, T) any) Stream[any]

	// FlatMapConcat returns a stream consisting of the results of
	// replacing each element of this stream with the contents of a mapped stream produced by applying the provided mapping function to each element.
	FlatMapConcat(f func(T) Source[T]) Stream[T]
	// FlatMapConcatAny returns a stream consisting of the results of
	// replacing each element of this stream with the contents of a mapped stream produced by applying the provided mapping function to each element.
	// the function return any typed source.
	FlatMapConcatAny(f func(T) Source[any]) Stream[any]

	// Take returns a stream consisting of the first n elements of this stream.
	Take(n int) Stream[T]
	// Skip returns a stream consisting of the remaining elements of this stream after discarding the first n elements of the stream.
	Skip(n int) Stream[T]

	// Distinct returns a stream consisting of the subsequent distinct elements of this stream.
	Distinct() Stream[T]
	// DistinctBy returns a stream consisting of the subsequent distinct elements of this stream.
	DistinctBy(cmp func(old T, new T) bool) Stream[T]

	// ZipWith returns a stream consisting of appling the given function to the elements of this stream and another stream.
	ZipWith(other Source[T], zipf func(T, T) T) Stream[T]
	// ZipWithAny returns a stream consisting of appling the given function to the elements of this stream and another stream.
	// the function return any type.
	ZipWithAny(other Source[any], zipf func(T, any) any) Stream[any]
	// ZipWithPrev returns a stream consisting of applying the given function to
	// the elements of this stream and its previous element.
	// prev is nil for the first element.
	ZipWithPrev(zipf func(prev T, ele T) T) Stream[T]

	// Scan returns a stream consisting of the accumlated results of applying the given function to the elements of this stream.
	Scan(init T, accumf func(acc T, ele T) T) Stream[T]
	// ScanAny returns a stream consisting of the accumlated results of applying the given function to the elements of this stream.
	// an associative accumulation function that returns any type.
	ScanAny(init any, accumf func(acc any, ele T) any) Stream[any]

	// OnEach returns a stream do nothing but visit each element of this stream.
	OnEach(visit func(v T)) Stream[T]

	Collector[T]
}

// Source is source of a stream.
type Source[T any] interface {

	// Next returns true if the stream has more elements.
	Next() bool

	// Get returns the next element in the stream.
	// should return same value if called multiple times.
	Get() T
}

// terminal operations

// Collector is a terminal operation that collects the stream elements into a container.
type Collector[T any] interface {

	// ForEach performs an action for each element of this stream.
	ForEach(visit func(ele T))
	// ForEachIndex performs an action for each element of this stream.
	ForEachIndex(visit func(int, T))

	// Collect returns a slice containing the elements of this stream.
	Collect() (target []T)
	// CollectTo collects the stream elements into a given slice.
	CollectTo(target []T) []T

	// Reduce performs a reduction on the elements of this stream,
	Reduce(reducer func(acc T, ele T) T) (result T)
	// ReduceAny performs a reduction on the elements of this stream,
	// an associative accumulation function that returns any type,
	// and returns the reduced value as any type.
	ReduceAny(reducer func(acc any, ele T) any) (result any)

	// Fold performs a reduction on the elements of this stream,
	Fold(init T, reducer func(acc T, ele T) T) (result T)
	// FoldAny performs a reduction on the elements of this stream,
	// an associative accumulation function that returns any type,
	// and returns the reduced value as any type.
	FoldAny(init any, reducer func(acc any, ele T) any) (result any)

	// Find returns the first element of this stream matching the given predicate
	Find(predicate func(T) bool) (found T, err error)
	// FindOr returns the first element of this stream matching the given predicate, or default value if no such element exists.
	FindOr(predicate func(ele T) bool, defvalue T) T
	// FindIndex returns the index of the first element of this down stream, not source, matching the given predicate,
	FindIndex(predicate func(T) bool) int
	// FindLast returns the last element of this stream matching the given predicate.
	FindLast(predicate func(T) bool) (found T, err error)
	// FindLastOr returns the last element of this stream matching the given predicate, or default value if no such element exists.
	FindLastOr(predicate func(T) bool, defvalue T) (found T)
	// FindLastIndex returns the index of the last element of this stream matching the given predicate.
	FindLastIndex(predicate func(T) bool) (found int)

	// Count returns the count of elements of this stream.
	// it does not consume the stream, specifically, it does not call Get().
	Count() int

	// All returns true if all elements of this stream match the given predicate.
	All(predicate func(T) bool) bool

	// Any returns true if any elements of this stream match the given predicate.
	Any(predicate func(T) bool) bool
}

type baseStream[T any] struct {
	idx   int // full index
	limit int // -1, unlimited or unknown
	next  func() bool
	get   func() any
}

//
// source operations
//

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
	filterstream := new(baseStream[T])
	filterstream.idx = -1
	filterstream.next = func() bool {
		for s.next() {
			filterstream.idx++
			if filter(s.get().(T)) {
				return true
			}
		}
		return false
	}

	filterstream.get = func() any {
		return s.get()
	}
	return filterstream
}

func (s *baseStream[T]) Map(mapf func(T) T) Stream[T] {
	if s == nil {
		return s
	}

	mapstream := new(baseStream[T])
	mapstream.idx = -1
	mapstream.next = func() bool {
		if s.next() {
			mapstream.idx++
			return true
		}
		return false
	}

	mapstream.get = func() any {
		return mapf(s.get().(T))
	}
	return mapstream
}

func (s *baseStream[T]) MapAny(mapf func(T) any) Stream[any] {
	if s == nil {
		return nilAnyStream
	}

	// T -> any
	mapstream := new(baseStream[any])
	mapstream.idx = -1
	mapstream.next = func() bool {
		if s.next() {
			mapstream.idx++
			return true
		}
		return false
	}

	mapstream.get = func() any {
		return mapf(s.get().(T))
	}
	return mapstream
}

// MapIndex returns a stream consisting of the results of applying the given function to the elements of this stream.
func (s *baseStream[T]) MapIndex(mapf func(int, T) T) Stream[T] {
	if s == nil {
		return s
	}

	mapstream := new(baseStream[T])
	mapstream.idx = -1
	mapstream.next = func() bool {
		if s.next() {
			mapstream.idx++
			return true
		}
		return false
	}
	mapstream.get = func() any {
		return mapf(mapstream.idx, s.get().(T))
	}
	return mapstream
}

// MapIndexAny returns a stream consisting of the results of applying the given function to the elements of this stream.
func (s *baseStream[T]) MapIndexAny(mapf func(int, T) any) Stream[any] {
	if s == nil {
		return nilAnyStream
	}

	mapstream := new(baseStream[any])
	mapstream.idx = -1
	mapstream.next = func() bool {
		if s.next() {
			mapstream.idx++
			return true
		}
		return false
	}
	mapstream.get = func() any {
		return mapf(mapstream.idx, s.get().(T))
	}
	return mapstream
}

// FlatMapConcat returns a stream consisting of the results of
// replacing each element of this stream with the contents of
// a mapped stream produced by applying the provided mapping function to each element.
func (s *baseStream[T]) FlatMapConcat(fmap func(ele T) Source[T]) Stream[T] {
	if s == nil {
		return s
	}

	fmapstream := new(baseStream[T])
	fmapstream.idx = -1
	var source Source[T]
	fmapstream.next = func() bool {
	loop:
		if source != nil && source.Next() {
			fmapstream.idx++
			return true
		}
		for s.next() {
			source = fmap(s.get().(T))
			goto loop
		}
		return false
	}

	fmapstream.get = func() any {
		return source.Get()
	}
	return fmapstream
}

func (s *baseStream[T]) FlatMapConcatAny(fmap func(ele T) Source[any]) Stream[any] {
	if s == nil {
		return nilAnyStream
	}

	fmapstream := new(baseStream[any])
	fmapstream.idx = -1
	var source Source[any]
	fmapstream.next = func() bool {
	loop:
		if source != nil && source.Next() {
			fmapstream.idx++
			return true
		}
		for s.next() {
			source = fmap(s.get().(T))
			goto loop
		}
		return false
	}

	fmapstream.get = func() any {
		return source.Get()
	}
	return fmapstream
}

// Take returns a stream consisting of the first n elements of this stream.
func (s *baseStream[T]) Take(n int) Stream[T] {
	if s == nil {
		return s
	}

	takestream := new(baseStream[T])
	takestream.idx = -1
	takestream.next = func() bool {
		if takestream.idx+1 == n {
			return false
		}
		takestream.idx++
		return s.next()
	}
	takestream.get = func() any {
		return s.get()
	}
	return takestream
}

// Skip returns a stream consisting of the remaining elements of this stream after discarding the first n elements of the stream.
func (s *baseStream[T]) Skip(n int) Stream[T] {
	if s == nil {
		return s
	}

	skipstream := new(baseStream[T])
	skipstream.idx = -1
	skipstream.next = func() bool {
		for skipstream.idx+1 < n {
			if !s.next() {
				return false
			}
			skipstream.idx++
		}
		return s.next()
	}
	skipstream.get = func() any {
		return s.get()
	}
	return skipstream
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

	distinctstream := new(baseStream[T])
	var old T
	distinctstream.idx = -1
	distinctstream.next = func() bool {
		for s.next() {
			distinctstream.idx++
			v := s.get().(T)
			if !cmp(old, v) {
				old = v
				return true
			}
			old = v
		}
		return false
	}
	distinctstream.get = func() any {
		return old
	}
	return distinctstream
}

// ZipWith returns a stream consisting of appling the given function to the elements of this stream and another stream.
func (s *baseStream[T]) ZipWith(other Source[T], zipf func(T, T) T) Stream[T] {
	if s == nil {
		return s
	}

	zipstream := new(baseStream[T])
	zipstream.idx = -1
	zipstream.next = func() bool {
		if s.next() && other.Next() {
			zipstream.idx++
			return true
		}
		return false
	}
	zipstream.get = func() any {
		return zipf(s.get().(T), other.Get())
	}
	return zipstream
}

// ZipWithAny returns a stream consisting of appling the given function to the elements of this stream and another stream.
func (s *baseStream[T]) ZipWithAny(other Source[any], zipf func(T, any) any) Stream[any] {
	if s == nil {
		return nilAnyStream
	}

	zipstream := new(baseStream[any])
	zipstream.idx = -1
	zipstream.next = func() bool {
		if s.next() && other.Next() {
			zipstream.idx++
			return true
		}
		return false
	}
	zipstream.get = func() any {
		return zipf(s.get().(T), other.Get())
	}
	return zipstream
}

// ZipWithPrev returns a stream consisting of applying the given function to
// the elements of this stream and its previous element.
// [a, b, c, d] => [f(nil, a), f(a, b), f(b, c), f(c, d)]
func (s *baseStream[T]) ZipWithPrev(zipf func(prev T, ele T) T) Stream[T] {
	if s == nil {
		return s
	}

	zipstream := new(baseStream[T])
	zipstream.idx = -1
	var prev T
	zipstream.next = func() bool {
		if s.next() {
			zipstream.idx++
			return true
		}
		return false
	}
	zipstream.get = func() any {
		v := s.get().(T)
		result := zipf(prev, v)
		prev = v
		return result
	}
	return zipstream
}

// Scan returns a stream consisting of the accumlated results of applying the given function to the elements of this stream.
//
//	with source=[1, 2, 3, 4], init=0 and func(acc, i) { acc + i } produces [1, 3, 6, 10]
func (s *baseStream[T]) Scan(init T, accumf func(acc T, ele T) T) Stream[T] {
	if s == nil {
		return s
	}

	scanstream := new(baseStream[T])
	scanstream.idx = -1
	acc := init
	scanstream.next = func() bool {
		if s.next() {
			scanstream.idx++
			return true
		}
		return false
	}
	scanstream.get = func() any {
		acc = accumf(acc, s.get().(T))
		return acc
	}
	return scanstream
}

func (s *baseStream[T]) ScanAny(init any, accumf func(acc any, ele T) any) Stream[any] {
	if s == nil {
		return nilAnyStream
	}

	scanstream := new(baseStream[any])
	scanstream.idx = -1
	acc := init
	scanstream.next = func() bool {
		if s.next() {
			scanstream.idx++
			return true
		}
		return false
	}
	scanstream.get = func() any {
		acc = accumf(acc, s.get().(T))
		return acc
	}
	return scanstream
}

// OnEach returns a stream do nothing but visit each element of this stream.
func (s *baseStream[T]) OnEach(visit func(v T)) Stream[T] {
	if s == nil {
		return s
	}

	eachstream := new(baseStream[T])
	eachstream.idx = -1
	eachstream.next = func() bool {
		if s.next() {
			eachstream.idx++
			return true
		}
		return false
	}
	eachstream.get = func() any {
		v := s.get()
		visit(v.(T))
		return v
	}
	return eachstream
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

// Collect returns a slice containing the elements of this stream.
func (s *baseStream[T]) Collect() (target []T) {
	if s == nil {
		return []T{}
	}

	target = []T{}
	for s.next() {
		v := s.get()
		target = append(target, v.(T))
	}
	return
}

// CollectTo collects the stream elements into a given slice.
func (s *baseStream[T]) CollectTo(target []T) []T {
	if s == nil {
		return target
	}

	// nil target
	if target == nil {
		return []T{}
	}

	for s.next() {
		v := s.get()
		target = append(target, v.(T))
	}
	return target
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

// ReduceAny performs a reduction on the elements of this stream,
// an associative accumulation function that returns any type,
// and returns the reduced value as any type.
// [a, b, c, d] => f(f(f(a, b), c), d)
// if the stream is empty returns nil
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

// Count returns the count of elements of this stream.
// it does not consume the stream, specifically, it does not call Get().
func (s *baseStream[T]) Count() (count int) {
	if s == nil {
		return 0
	}

	for s.next() {
		count++
	}
	return
}

func (s *baseStream[T]) All(predicate func(T) bool) bool {
	if s == nil {
		return false
	}

	// for empty
	result := false
	for s.next() {
		result = predicate(s.get().(T))
		if !result {
			return false
		}
	}
	return result
}

func (s *baseStream[T]) Any(predicate func(T) bool) bool {
	if s == nil {
		return false
	}

	// for empty
	result := false
	for s.next() {
		result = predicate(s.get().(T))
		if result {
			return true
		}
	}
	return result
}
