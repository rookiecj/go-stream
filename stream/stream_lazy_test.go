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

func TestStream_Distinct(t *testing.T) {

	arr := []myStruct{
		{"a"},
		{"a"},
		{"b"},
		{"c"},
		{"a"},
		{"b"},
	}

	type testCase[T any] struct {
		name string
		s    *Stream[any]
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "distinct",
			s:    ToStream(arr),
			want: []myStruct{{"a"}, {"b"}, {"c"}, {"a"}, {"b"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target []myStruct
			if got := CollectAs(tt.s.Distinct(), target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Distinct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_DistinctBy(t *testing.T) {
	arr := []myStruct{
		{"a"},
		{"a"},
		{"b"},
		{"c"},
		{"a"},
		{"b"},
	}

	type args struct {
		cmp func(any, any) bool
	}
	type testCase[T any] struct {
		name string
		s    *Stream[any]
		args args
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "distinctBy with deepequal",
			s:    ToStream(arr),
			args: args{
				cmp: func(old, v any) bool {
					return reflect.DeepEqual(old, v)
				},
			},
			want: []myStruct{{"a"}, {"b"}, {"c"}, {"a"}, {"b"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target []myStruct
			if got := CollectAs(tt.s.DistinctBy(tt.args.cmp), target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DistinctBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_ZipWith(t *testing.T) {
	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
	}

	arr2 := []myStruct{
		{"1"},
		{"2"},
		{"3"},
		{"4"},
	}

	type args struct {
		other *Stream[any]
		f     func(any, any) any
	}
	type testCase[T any] struct {
		name string
		s    *Stream[any]
		args args
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "zipwith",
			s:    ToStream(arr1),
			args: args{
				other: ToStream(arr2),
				f: func(ele1 any, ele2 any) any {
					result := myStruct{
						Name: ele1.(myStruct).Name + ele2.(myStruct).Name,
					}
					return result
				},
			},
			want: []myStruct{{"a1"}, {"b2"}, {"c3"}, {"d4"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target []myStruct
			if got := CollectAs(tt.s.ZipWith(tt.args.other, tt.args.f), target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ZipWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_Intermediates_NilReceiver(t *testing.T) {

	type args struct {
		f func(s *Stream[any]) *Stream[any]
	}
	type testCase[T any] struct {
		name string
		s    *Stream[any]
		args args
		want reflect.Type
	}

	var s *Stream[any]
	typeOfNilStream := reflect.TypeOf(s)

	tests := []testCase[myStruct]{
		{
			name: "filter",
			s:    s,
			args: args{
				f: func(s *Stream[any]) *Stream[any] {
					return s.Filter(nil)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "map",
			s:    s,
			args: args{
				f: func(s *Stream[any]) *Stream[any] {
					return s.Map(nil)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "mapIndex",
			s:    s,
			args: args{
				f: func(s *Stream[any]) *Stream[any] {
					return s.MapIndex(nil)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "flatmapconcat",
			s:    s,
			args: args{
				f: func(s *Stream[any]) *Stream[any] {
					return s.FlatMapConcat(nil)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "Take",
			s:    s,
			args: args{
				f: func(s *Stream[any]) *Stream[any] {
					return s.Take(0)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "Skip",
			s:    s,
			args: args{
				f: func(s *Stream[any]) *Stream[any] {
					return s.Skip(0)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "Distinct",
			s:    s,
			args: args{
				f: func(s *Stream[any]) *Stream[any] {
					return s.Distinct()
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "DistinctBy",
			s:    s,
			args: args{
				f: func(s *Stream[any]) *Stream[any] {
					return s.DistinctBy(nil)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "ZipWith",
			s:    s,
			args: args{
				f: func(s *Stream[any]) *Stream[any] {
					return s.ZipWith(nil, nil)
				},
			},
			want: typeOfNilStream,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.f(tt.s); reflect.TypeOf(got) != tt.want {
				t.Errorf("got = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}
