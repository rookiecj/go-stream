package stream

// FromSlice build a Stream from given slice
func FromSlice[T any](arr []T) *Stream[any] {
	return FromVar(arr...)
}

// FromVar build a Stream from variatic
func FromVar[T any](arr ...T) *Stream[any] {
	stream := new(Stream[any])
	stream.idx = -1
	stream.next = func() bool {
		if stream.idx+1 == len(arr) {
			return false
		}
		stream.idx++
		return true
	}

	stream.get = func() any {
		return arr[stream.idx]
	}
	return stream
}
