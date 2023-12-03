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
	//{a!}
	//{c!}
	//{e!}
	//{g!}
	//{i!}
}
