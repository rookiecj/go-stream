package stream

// FromSlice build a Stream from given slice
func FromSlice[T any](arr []T) *Stream[any] {
	// Stream should be work on empty slice with type safe manner
	//if arr == nil {
	//	var nilReceiver *Stream[any]
	//	return nilReceiver
	//}

	stream := new(Stream[any])
	stream.idx = -1
	limit := len(arr)
	stream.next = func() bool {
		if stream.idx+1 == limit {
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

// FromChan build a Stream from given channel
func FromChan[T any](ch <-chan T) (stream *Stream[any]) {
	stream = new(Stream[any])
	var v any
	var ok = true
	if ch == nil {
		stream.next = func() bool {
			return false
		}
	} else {
		stream.next = func() bool {
			//if !ok {
			//	done <- true
			//	return false
			//}
			v, ok = <-ch
			return ok
		}
	}

	stream.get = func() any {
		return v
	}
	return
}
