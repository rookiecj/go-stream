# go-stream

`go-stream` is a `Go` package to help the stream processing of elements in more functional manner.

`Stream` uses iterator pattern for visiting each element of a stream.
It is lazy, which means the operations are not executed until terminal operation is called.
and it is not thread safe.

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

	stream := s.FromSlice(arr)
	imm1 := stream.Filter(func(v any) bool {
		return len(v.(myStruct).Name) == 1
	}).Map(func(v any) any {
		return myStruct{v.(myStruct).Name + "!"}
	})

	myStructs := s.CollectAs(imm1, []myStruct{})
	fmt.Println(myStructs)

	arrempty := []myStruct{}
	imm2 := s.FromSlice(arrempty)
	myStructs2 := s.CollectAs(imm2, []myStruct{})
	fmt.Println(myStructs2)
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

### Intermediate operations

Intermediate operations generate new stream which consume data from upstream and apply operator on it.

- [X] Filter
- [x] Map
- [x] FlatMapConcat
- [ ] FlatMapConcurrent
- [X] Take, Skip
- [X] Distinct, DistinctBy
- [ ] Zip(with Pair)
- [X] ZipWith
- [ ] ZipWithPrev (Experimental)
- [X] Scan
- [ ] Window

### Terminal operations

Terminal operations are collectors which trigger streams to work. and return the result of the stream.

- [X] ForEach, ForEachIndex
- [X] Collect
- [ ] Reduce (Experimental)
- [X] Fold
- [X] Find/FindLast
- [ ] All, Any

Slightly more type safe functions are:
- [X] ForEachAs, ForEachAsIndex
- [X] CollectAs
- [ ] ReduceAs
- [ ] FoldAs

## TODO

- [X] make ToStream lazy
- [X] add more Builders 
- [ ] Stream to interface
- [ ] add more intermediate operations
- [ ] add more terminal operations
- [ ] add doc
- [ ] add unittest

