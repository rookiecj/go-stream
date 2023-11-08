# go-stream

`go-stream` is a `Go` package to easily handle the stream of slice elements.

`Stream` uses iterator pattern for visit each elements of a stream.
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

	stream := s.ToStream(arr)
	imm1 := stream.Filter(func(v any) bool {
		return len(v.(myStruct).Name) == 1
	}).Map(func(v any) any {
		return myStruct{v.(myStruct).Name + "!"}
	})

	myStructs := s.CollectAs(imm1, []myStruct{})
	fmt.Println(myStructs)

	arrempty := []myStruct{}
	imm2 := s.ToStream(arrempty)
	myStructs2 := s.CollectAs(imm2, []myStruct{})
	fmt.Println(myStructs2)
}

```

## Operations

There 3 main operations:

- Builders
- Intermediates
- Termimals

### Builders

Builder operations build a `Stream` from various sources like slice and array.

- [X] ToStream
- [ ] FromSlice 
- [ ] FromChan

### Intermediate operations

Intermediate operations generate new stream which consume data from upstream and apply operator on it.

- [X] Filter
- [x] Map
- [x] FlatMapConcat
- [ ] FlatMapConcurrent
- [X] Take, Skip
- [X] Distinct
- [ ] Zip with Pair 
- [X] ZipWith

### Terminal operations

Terminal operations are collectors which trigger streams to work. and return the result of the stream.

- [X] ForEach, ForEachIndex
- [X] CollectAs
- [X] Reduce
- [ ] Fold

## TODO

- [X] make ToStream lazy
- [ ] add more Builders 
- [ ] Stream to interface
- [ ] add more intermediate operations
- [ ] add more terminal operations
- [ ] add doc
- [ ] add unittest

