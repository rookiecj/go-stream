package stream

//
// stream builders
//

// FromSlice build a Stream from given slice
func FromSlice[T any](arr []T) Stream[T] {

	stream := new(baseStream[T])
	stream.idx = -1
	stream.limit = len(arr)
	stream.next = func() bool {
		if stream.idx+1 == stream.limit {
			return false
		}
		stream.idx++
		return true
	}

	stream.get = func() any {
		return arr[stream.idx]
	}
	stream.getonrecover = func() RecoverFunc {
		return nil
	}
	return stream
}

// FromVar build a Stream from variatic
func FromVar[T any](arr ...T) Stream[T] {
	stream := new(baseStream[T])
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
	stream.getonrecover = func() RecoverFunc {
		return nil
	}
	return stream
}

// FromChan build a Stream from given channel
func FromChan[T any](ch <-chan T) Stream[T] {
	stream := new(baseStream[T])
	stream.idx = -1
	var v T
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
			if ok {
				stream.idx++
			}
			return ok
		}
	}

	stream.get = func() any {
		return v
	}
	stream.getonrecover = func() RecoverFunc {
		return nil
	}
	return stream
}

func FromSource[T any](source Source[T]) Stream[T] {
	stream := new(baseStream[T])
	stream.idx = -1
	stream.next = func() bool {
		if source.Next() {
			stream.idx++
			return true
		}
		return false
	}

	stream.get = func() any {
		return source.Get()
	}
	stream.getonrecover = func() RecoverFunc {
		return nil
	}
	return stream
}
