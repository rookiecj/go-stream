package stream

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type myStruct struct {
	Name string
}

func toInterface[T any](arr []T) (result []any) {
	for _, a := range arr {
		result = append(result, a)
	}
	return
}

func TestStream_Collect(t *testing.T) {

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
		//want []T
		want []interface{}
	}
	tests := []testCase[myStruct]{
		{
			name: "empty array",
			args: args[myStruct]{
				s: FromSlice(emptyArr),
			},
			want: toInterface([]myStruct{}),
		},
		{
			name: "empty slice",
			args: args[myStruct]{
				s: FromSlice(emptySlice),
			},
			want: toInterface([]myStruct{}),
		},
		{
			name: "filter_map",
			args: args[myStruct]{
				s: FromSlice(arr).Filter(func(v any) bool {
					return len(v.(myStruct).Name) == 1
				}).Map(func(v any) any {
					return myStruct{v.(myStruct).Name + "!"}
				}),
			},
			want: toInterface([]myStruct{{"a!"}, {"c!"}, {"e!"}, {"g!"}, {"i!"}}),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.s.Collect(); got != nil {
				gotValue := reflect.ValueOf(got)
				wantValue := reflect.ValueOf(tt.want)
				gotValueT := gotValue.Type()
				wantValueT := wantValue.Type()
				fmt.Printf("%v, %v\n", gotValueT, wantValueT)
				//[]interface {}, []stream.myStruct
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Collect() = %v, want %v", got, tt.want)
				}
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
			name: "slice slice",
			s:    FromSlice(arr),
			args: args{
				f: func(v any) *Stream[any] {
					ms := v.(myStruct)
					stream := FromSlice[myStruct]([]myStruct{ms, ms})
					return stream
				},
			},
			want: []myStruct{{"a"}, {"a"}, {"b"}, {"b"}, {"c"}, {"c"}, {"d"}, {"d"}},
		},
		{
			name: "slice variatic",
			s:    FromSlice(arr),
			args: args{
				f: func(v any) *Stream[any] {
					ms := v.(myStruct)
					stream := FromVar[myStruct]([]myStruct{ms, ms}...)
					return stream
				},
			},
			want: []myStruct{{"a"}, {"a"}, {"b"}, {"b"}, {"c"}, {"c"}, {"d"}, {"d"}},
		},
		{
			name: "slice chan",
			s:    FromSlice(arr),
			args: args{
				f: func(v any) *Stream[any] {
					ch := make(chan myStruct, 0)
					go func() {
						ch <- v.(myStruct)
						ch <- v.(myStruct)
						close(ch)
					}()
					time.Sleep(100 * time.Millisecond)
					stream := FromChan[myStruct](ch)
					return stream
				},
			},
			want: []myStruct{{"a"}, {"a"}, {"b"}, {"b"}, {"c"}, {"c"}, {"d"}, {"d"}},
		},
		{
			name: "chan slice",
			s: func() *Stream[any] {
				ch := make(chan myStruct, 0)
				go func() {
					for _, ele := range arr {
						ch <- ele
					}
					close(ch)
				}()
				time.Sleep(100 * time.Millisecond)
				stream := FromChan[myStruct](ch)
				return stream
			}(),
			args: args{
				f: func(v any) *Stream[any] {
					ms := v.(myStruct)
					stream := FromSlice[myStruct]([]myStruct{ms, ms})
					return stream
				},
			},
			want: []myStruct{{"a"}, {"a"}, {"b"}, {"b"}, {"c"}, {"c"}, {"d"}, {"d"}},
		},
		{
			name: "chan chan",
			s: func() *Stream[any] {
				ch := make(chan myStruct, 0)
				go func() {
					for _, ele := range arr {
						ch <- ele
					}
					close(ch)
				}()
				time.Sleep(100 * time.Millisecond)
				stream := FromChan[myStruct](ch)
				return stream
			}(),
			args: args{
				f: func(v any) *Stream[any] {
					ch := make(chan myStruct, 0)
					go func() {
						ch <- v.(myStruct)
						ch <- v.(myStruct)
						close(ch)
					}()
					time.Sleep(100 * time.Millisecond)
					stream := FromChan[myStruct](ch)
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
			s:    FromSlice(arr),
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
			s:    FromSlice(arr),
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
			s:    FromSlice(arr1),
			args: args{
				other: FromSlice(arr2),
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

func TestStream_Scan(t *testing.T) {
	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
	}

	init := myStruct{"!"}

	type args[T any] struct {
		init   T
		accumf func(acc any, ele any) any
	}
	type testCase[T any] struct {
		name string
		s    *Stream[any]
		args args[T]
		want []myStruct
	}
	tests := []testCase[myStruct]{
		{
			name: "Scan/add",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				init: init,
				accumf: func(acc, ele any) any {
					return myStruct{acc.(myStruct).Name + ele.(myStruct).Name}
				},
			},
			want: []myStruct{{"!a"}, {"!ab"}, {"!abc"}, {"!abcd"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CollectAs(tt.s.Scan(tt.args.init, tt.args.accumf), []myStruct{}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_FindIndex(t *testing.T) {

	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
	}

	type args struct {
		predicate func(any) bool
	}
	type testCase[T any] struct {
		name string
		s    *Stream[any]
		args args
		want int
	}
	tests := []testCase[myStruct]{
		{
			name: "find/0",
			s:    FromSlice(arr1),
			args: args{
				predicate: func(ele any) bool {
					return ele.(myStruct) == myStruct{"a"}
				},
			},
			want: 0,
		},
		{
			name: "find/3",
			s:    FromSlice(arr1),
			args: args{
				predicate: func(ele any) bool {
					return ele.(myStruct) == myStruct{"d"}
				},
			},
			want: 3,
		},
		{
			name: "find/-1(not found)",
			s:    FromSlice(arr1),
			args: args{
				predicate: func(ele any) bool {
					return false
				},
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.FindIndex(tt.args.predicate); got != tt.want {
				t.Errorf("FindIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_FindLastIndex(t *testing.T) {

	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"b"},
		{"d"},
	}

	type args struct {
		predicate func(any) bool
	}
	type testCase[T any] struct {
		name string
		s    *Stream[any]
		args args
		want int
	}
	tests := []testCase[myStruct]{
		{
			name: "find/0",
			s:    FromSlice(arr1),
			args: args{
				predicate: func(ele any) bool {
					return ele.(myStruct) == myStruct{"a"}
				},
			},
			want: 0,
		},
		{
			name: "find/3",
			s:    FromSlice(arr1),
			args: args{
				predicate: func(ele any) bool {
					return ele.(myStruct) == myStruct{"d"}
				},
			},
			want: 3,
		},
		{
			name: "find/2",
			s:    FromSlice(arr1),
			args: args{
				predicate: func(ele any) bool {
					return ele.(myStruct) == myStruct{"b"}
				},
			},
			want: 2,
		},
		{
			name: "find/-1(not found)",
			s:    FromSlice(arr1),
			args: args{
				predicate: func(ele any) bool {
					return false
				},
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.FindLastIndex(tt.args.predicate); got != tt.want {
				t.Errorf("FindLastIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_Fold(t *testing.T) {
	var emptySlice []myStruct

	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
	}

	type args[T any] struct {
		init    T
		reducer func(acc any, ele any) any
	}
	type testCase[T any] struct {
		name string
		s    *Stream[any]
		args args[T]
		want T
	}
	tests := []testCase[myStruct]{
		{
			name: "fold empty",
			s:    FromSlice(emptySlice),
			args: args[myStruct]{
				init: myStruct{"!"},
				reducer: func(acc any, ele any) any {
					return myStruct{acc.(myStruct).Name + ele.(myStruct).Name}
				},
			},
			want: myStruct{"!"},
		},
		{
			name: "fold",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				init: myStruct{"!"},
				reducer: func(acc any, ele any) any {
					return myStruct{acc.(myStruct).Name + ele.(myStruct).Name}
				},
			},
			want: myStruct{"!abcd"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.s.Fold(tt.args.init, tt.args.reducer); !reflect.DeepEqual(gotResult, tt.want) {
				t.Errorf("Fold() = %v, want %v", gotResult, tt.want)
			}
		})
	}
}

func TestStream_Fold_Reduce_As_Array(t *testing.T) {
	var emptySlice []myStruct

	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
	}

	type args[T any] struct {
		init    []T
		reducer func(acc any, ele any) any
	}
	type testCase[T any] struct {
		name string
		s    *Stream[any]
		args args[T]
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "fold empty list",
			s:    FromSlice(emptySlice),
			args: args[myStruct]{
				init: []myStruct{},
				reducer: func(acc any, ele any) any {
					return append(acc.([]myStruct), ele.(myStruct))
				},
			},
			want: []myStruct{},
		},
		{
			name: "fold empty list with init value",
			s:    FromSlice(emptySlice),
			args: args[myStruct]{
				init: []myStruct{{"!"}},
				reducer: func(acc any, ele any) any {
					return append(acc.([]myStruct), ele.(myStruct))
				},
			},
			want: []myStruct{{"!"}},
		},
		{
			name: "fold",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				init: []myStruct{},
				reducer: func(acc any, ele any) any {
					return append(acc.([]myStruct), ele.(myStruct))
				},
			},
			want: []myStruct{{"a"}, {"b"}, {"c"}, {"d"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.s.Fold(tt.args.init, tt.args.reducer); !reflect.DeepEqual(gotResult, tt.want) {
				t.Errorf("Fold() = %v, want %v", gotResult, tt.want)
			}
		})
	}
}
