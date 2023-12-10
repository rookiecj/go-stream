package stream

import (
	"reflect"
	"testing"
)

type myStructSource struct {
	index     int
	myStructs []myStruct
}

func newMyStructSource(myStructs []myStruct) Source[Indexed[myStruct]] {
	return &myStructSource{
		index:     -1,
		myStructs: myStructs,
	}
}

func (c *myStructSource) Next() bool {
	if c.index+1 == len(c.myStructs) {
		return false
	}
	c.index++
	return true
}

func (c *myStructSource) Get() Indexed[myStruct] {
	return Indexed[myStruct]{
		Index: c.index,
		Value: c.myStructs[c.index],
	}
}

func TestFromSource_Indexed(t *testing.T) {
	var emptySlice []myStruct

	arr := []myStruct{
		{"a"},
		{"bb"},
		{"c"},
		{"ddd"},
		{"e"},
	}

	type args[T any] struct {
		s      Stream[Indexed[T]]
		target []Indexed[T]
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []Indexed[T] // we can expect type T
	}
	tests := []testCase[myStruct]{
		{
			name: "empty",
			args: args[myStruct]{
				s:      FromSource(newMyStructSource(emptySlice)),
				target: []Indexed[myStruct]{},
			},
			want: []Indexed[myStruct]{},
		},
		{
			name: "slice",
			args: args[myStruct]{
				s:      FromSource(newMyStructSource(arr)),
				target: []Indexed[myStruct]{},
			},
			want: []Indexed[myStruct]{
				{0, myStruct{"a"}},
				{1, myStruct{"bb"}},
				{2, myStruct{"c"}},
				{3, myStruct{"ddd"}},
				{4, myStruct{"e"}},
			},
		},
		{
			name: "slice filter map",
			args: args[myStruct]{
				s: FromSource(newMyStructSource(arr)).
					Filter(func(ele Indexed[myStruct]) bool {
						return len(ele.Value.Name) == 1
					}).Map(func(ele Indexed[myStruct]) Indexed[myStruct] {
					return Indexed[myStruct]{ele.Index,
						myStruct{ele.Value.Name + "!"},
					}
				}),
				target: []Indexed[myStruct]{},
			},
			want: []Indexed[myStruct]{
				{0, myStruct{"a!"}},
				{2, myStruct{"c!"}},
				{4, myStruct{"e!"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.s.Collect(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromSource() = %v, want %v", got, tt.want)
			}
		})
	}
}
