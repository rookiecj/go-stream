# go-stream

`go-stream` is a `Go` package to help processing stream of elements in more functional manner.

`Stream` uses iterator pattern for visiting each element of a stream.
It is lazy, which means the operations are not executed until terminal operation is called.

## How to use

Install as follows:
```
go get github.com/rookiecj/go-stream
```

and uses as follows:
```go
package main 

import (
	"fmt"

	s "github.com/rookiecj/go-stream/stream"
)

type myStruct struct {
	Name string
}

func main() {
	arr := []myStruct{
		{"a"},
		{"bbb"},

		{"c"},
		{"dddd"},
		{"e"},
		{"fffff"},
		{"g"},
		{"hhhhh"},
		{"i"},
	}

	s.FromSlice(arr).
	Filter(func(v myStruct) bool {
		return len(v.Name) == 1
	}).
	Map(func(v myStruct) myStruct {
		return myStruct{v.Name + "!"}
	}).
	ForEach(func(ele myStruct) {
		fmt.Println(ele)
	})
}

```

## Operations

There 3 main operations:

- Builders
- Intermediates
- Terminals

### Builders

Builder operations build a `Stream` from various sources like slice and array.

- [X] FromSlice 
- [X] FromVar
- [X] FromChan (Experimental)
- [X] FromSource

### Intermediate operations

Intermediate operations generate new stream which consume data from upstream and apply operator on it.

- [X] Filter
- [x] Map/MapAny,MapIndex/MapIndexAny
- [x] FlatMapConcat/FlatMapConcatAny(Experimental)
- [ ] FlatMapConcurrent
- [X] Take, Skip
- [X] Distinct, DistinctBy
- [X] ZipWith/ZipWithAny
- [X] ZipWithPrev (Experimental)
- [X] Scan/ScanAny
- [ ] Window
- [ ] OnRecover (Experimental)

### Terminal operations

Terminal operations are collectors which trigger streams to work. and return the result of the stream.

- [X] ForEach, ForEachIndex
- [X] Collect/CollectTo
- [X] Reduce/ReduceAny
- [X] Fold/FoldAny
- [X] Find/FindIndex/FindLast
- [X] All,Any
- [X] Count

Slightly more type safe functions are:
- [X] ForEachAs, ForEachIndex
- [X] CollectAs, CollectTo
- [ ] CollectAsSafe, CollectionToSafe
- [X] ReduceAs
- [ ] FoldAs
- [X] FindOrAs


## TODO

- [X] make ToStream lazy
- [X] add more Builders 
- [X] Stream to interface
- [ ] add more intermediate operations
- [ ] add more terminal operations
- [X] add doc
- [X] add unittest
- [ ] remove Source/Collector from Stream
- [ ] add more safe terminal operations