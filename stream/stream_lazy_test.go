package stream

import (
	"reflect"
	"testing"
)

func TestCollectAs(t *testing.T) {

	type myStruct struct {
		Name string
	}
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
			want: emptyArr,
		},
		{
			name: "empty slice",
			args: args[myStruct]{
				s:      ToStream(emptySlice),
				target: []myStruct{},
			},
			want: emptySlice,
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
