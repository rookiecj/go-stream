package stream

import (
	"reflect"
	"testing"
)

func TestCollectAs(t *testing.T) {
	emptyArr := make([]myStruct, 0)
	var emptySlice []myStruct

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

	type args[T any] struct {
		s      Stream[T]
		target []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T // we can expect type T
	}
	tests := []testCase[myStruct]{
		{
			name: "array/empty",
			args: args[myStruct]{
				s:      FromSlice(emptyArr),
				target: []myStruct{},
			},
			want: []myStruct{},
		},
		{
			name: "slice/empty",
			args: args[myStruct]{
				s:      FromSlice(emptySlice),
				target: []myStruct{},
			},
			want: []myStruct{},
		},
		{
			name: "filter_map",
			args: args[myStruct]{
				s: FromSlice(arr).Filter(func(v myStruct) bool {
					return len(v.Name) == 1
				}).Map(func(v myStruct) myStruct {
					return myStruct{v.Name + "!"}
				}),
				target: []myStruct{},
			},
			want: []myStruct{{"a!"}, {"c!"}, {"e!"}, {"g!"}, {"i!"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CollectAs[myStruct](tt.args.s.AsSource(), tt.args.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CollectAs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestForEachAs(t *testing.T) {
	emptyArr := make([]myStruct, 0)
	var emptySlice []myStruct
	var collected []myStruct

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

	type args[T any] struct {
		s Stream[T]
		f func(T)
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "empty array",
			args: args[myStruct]{
				s: FromSlice(emptyArr),
				f: func(v myStruct) {
					collected = append(collected, v)
				},
			},
			want: []myStruct{},
		},
		{
			name: "empty slice",
			args: args[myStruct]{
				s: FromSlice(emptySlice),
				f: func(v myStruct) {
					collected = append(collected, v)
				},
			},
			want: []myStruct{},
		},
		{
			name: "filter_map",
			args: args[myStruct]{
				s: FromSlice(arr).
					Filter(func(v myStruct) bool {
						return len(v.Name) == 1
					}).
					Map(func(v myStruct) myStruct {
						return myStruct{v.Name + "!"}
					}),
				f: func(v myStruct) {
					collected = append(collected, v)
				},
			},
			want: []myStruct{{"a!"}, {"c!"}, {"e!"}, {"g!"}, {"i!"}},
		},
	}
	for _, tt := range tests {
		collected = []myStruct{}
		t.Run(tt.name, func(t *testing.T) {
			if ForEachAs[myStruct](tt.args.s.AsSource(), tt.args.f); !reflect.DeepEqual(collected, tt.want) {
				t.Errorf("ForEachAs() = %v, want %v", collected, tt.want)
			}
		})
	}
}
