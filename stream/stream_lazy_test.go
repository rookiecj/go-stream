package stream

import (
	"reflect"
	"testing"
)

type myStruct struct {
	Name string
}

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
		s      *Stream[any]
		target []T
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
				s:      ToStream(emptyArr),
				target: []myStruct{},
			},
			want: []myStruct{},
		},
		{
			name: "empty slice",
			args: args[myStruct]{
				s:      ToStream(emptySlice),
				target: []myStruct{},
			},
			want: []myStruct{},
		},
		{
			name: "filter_map",
			args: args[myStruct]{
				s: ToStream(arr).Filter(func(v any) bool {
					return len(v.(myStruct).Name) == 1
				}).Map(func(v any) any {
					return myStruct{v.(myStruct).Name + "!"}
				}),
				target: []myStruct{},
			},
			want: []myStruct{{"a!"}, {"c!"}, {"e!"}, {"g!"}, {"i!"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CollectAs(tt.args.s, tt.args.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CollectAs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_FlatMapConCat(t *testing.T) {

	arr := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
	}

	type args struct {
		f func(any) *Stream[any]
	}
	type testCase[T any] struct {
		name string
		s    *Stream[any]
		args args
		want []myStruct
	}
	tests := []testCase[myStruct]{
		{
			name: "flatmap_concat",
			s:    ToStream(arr),
			args: args{
				f: func(v any) *Stream[any] {
					ms := v.(myStruct)
					stream := ToStream[myStruct]([]myStruct{ms, ms})
					return stream
				},
			},
			want: []myStruct{{"a"}, {"a"}, {"b"}, {"b"}, {"c"}, {"c"}, {"d"}, {"d"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := []myStruct{}
			if got := CollectAs(tt.s.FlatMapConcat(tt.args.f), target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FlatMapConCat() = %v, want %v", got, tt.want)
			}
		})
	}
}
