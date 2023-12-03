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
		s      Stream[T]
		target []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		//want []T
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "empty array",
			args: args[myStruct]{
				s: FromSlice(emptyArr),
			},
			//want: toInterface([]myStruct{}),
			want: []myStruct{},
		},
		{
			name: "empty slice",
			args: args[myStruct]{
				s: FromSlice(emptySlice),
			},
			//want: toInterface([]myStruct{}),
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
			},
			//want: toInterface([]myStruct{{"a!"}, {"c!"}, {"e!"}, {"g!"}, {"i!"}}),
			want: []myStruct{
				{"a!"},
				{"c!"},
				{"e!"},
				{"g!"},
				{"i!"},
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.s.Collect(); got != nil {
				gotValue := reflect.ValueOf(got)
				wantValue := reflect.ValueOf(tt.want)
				gotValueT := gotValue.Type()
				wantValueT := wantValue.Type()
				fmt.Printf("type %v, %v\n", gotValueT, wantValueT)
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

	type args[T any] struct {
		f func(T) Source[T]
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want []myStruct
	}
	tests := []testCase[myStruct]{
		{
			name: "slice slice",
			s:    FromSlice(arr),
			args: args[myStruct]{
				f: func(v myStruct) Source[myStruct] {
					ms := v
					stream := FromSlice[myStruct]([]myStruct{ms, ms})
					return stream
				},
			},
			want: []myStruct{{"a"}, {"a"}, {"b"}, {"b"}, {"c"}, {"c"}, {"d"}, {"d"}},
		},
		{
			name: "slice variatic",
			s:    FromSlice(arr),
			args: args[myStruct]{
				f: func(v myStruct) Source[myStruct] {
					ms := v
					stream := FromVar[myStruct]([]myStruct{ms, ms}...)
					return stream
				},
			},
			want: []myStruct{{"a"}, {"a"}, {"b"}, {"b"}, {"c"}, {"c"}, {"d"}, {"d"}},
		},
		{
			name: "slice chan",
			s:    FromSlice(arr),
			args: args[myStruct]{
				f: func(v myStruct) Source[myStruct] {
					ch := make(chan myStruct, 0)
					go func() {
						ch <- v
						ch <- v
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
			s: func() Stream[myStruct] {
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
			args: args[myStruct]{
				f: func(v myStruct) Source[myStruct] {
					ms := v
					stream := FromSlice[myStruct]([]myStruct{ms, ms})
					return stream
				},
			},
			want: []myStruct{{"a"}, {"a"}, {"b"}, {"b"}, {"c"}, {"c"}, {"d"}, {"d"}},
		},
		{
			name: "chan chan",
			s: func() Stream[myStruct] {
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
			args: args[myStruct]{
				f: func(v myStruct) Source[myStruct] {
					ch := make(chan myStruct, 0)
					go func() {
						ch <- v
						ch <- v
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
			if got := tt.s.FlatMapConcat(tt.args.f).Collect(); !reflect.DeepEqual(got, tt.want) {
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
		s    Stream[T]
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
			if got := tt.s.Distinct().Collect(); !reflect.DeepEqual(got, tt.want) {
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
		cmp func(myStruct, myStruct) bool
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "distinctBy with deepequal",
			s:    FromSlice(arr),
			args: args{
				cmp: func(old, v myStruct) bool {
					return reflect.DeepEqual(old, v)
				},
			},
			want: []myStruct{{"a"}, {"b"}, {"c"}, {"a"}, {"b"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.DistinctBy(tt.args.cmp).Collect(); !reflect.DeepEqual(got, tt.want) {
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
		{"e"},
	}

	arr2 := []myStruct{
		{"1"},
		{"2"},
		{"3"},
		{"4"},
	}

	type args[T any] struct {
		other Stream[T]
		f     func(T, T) T
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "zipwith",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				other: FromSlice(arr2),
				f: func(ele1, ele2 myStruct) myStruct {
					result := myStruct{
						Name: ele1.Name + ele2.Name,
					}
					return result
				},
			},
			want: []myStruct{{"a1"}, {"b2"}, {"c3"}, {"d4"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.ZipWith(tt.args.other, tt.args.f).Collect(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ZipWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_Intermediates_NilReceiver(t *testing.T) {

	type args[T any] struct {
		f func(s Stream[T]) Stream[T]
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want reflect.Type
	}

	//var s Stream[myStruct]
	var s *baseStream[myStruct]
	typeOfNilStream := reflect.TypeOf(s)

	tests := []testCase[myStruct]{
		{
			name: "filter",
			s:    s,
			args: args[myStruct]{
				f: func(s Stream[myStruct]) Stream[myStruct] {
					return s.Filter(nil)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "map",
			s:    s,
			args: args[myStruct]{
				f: func(s Stream[myStruct]) Stream[myStruct] {
					return s.Map(nil)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "mapIndex",
			s:    s,
			args: args[myStruct]{
				f: func(s Stream[myStruct]) Stream[myStruct] {
					return s.MapIndex(nil)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "flatmapconcat",
			s:    s,
			args: args[myStruct]{
				f: func(s Stream[myStruct]) Stream[myStruct] {
					return s.FlatMapConcat(nil)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "Take",
			s:    s,
			args: args[myStruct]{
				f: func(s Stream[myStruct]) Stream[myStruct] {
					return s.Take(0)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "Skip",
			s:    s,
			args: args[myStruct]{
				f: func(s Stream[myStruct]) Stream[myStruct] {
					return s.Skip(0)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "Distinct",
			s:    s,
			args: args[myStruct]{
				f: func(s Stream[myStruct]) Stream[myStruct] {
					return s.Distinct()
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "DistinctBy",
			s:    s,
			args: args[myStruct]{
				f: func(s Stream[myStruct]) Stream[myStruct] {
					return s.DistinctBy(nil)
				},
			},
			want: typeOfNilStream,
		},
		{
			name: "ZipWith",
			s:    s,
			args: args[myStruct]{
				f: func(s Stream[myStruct]) Stream[myStruct] {
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
		accumf func(acc T, ele T) T
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want []myStruct
	}
	tests := []testCase[myStruct]{
		{
			name: "Scan/add",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				init: init,
				accumf: func(acc, ele myStruct) myStruct {
					return myStruct{acc.Name + ele.Name}
				},
			},
			want: []myStruct{{"!a"}, {"!ab"}, {"!abc"}, {"!abcd"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Scan(tt.args.init, tt.args.accumf).Collect(); !reflect.DeepEqual(got, tt.want) {
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

	type args[T any] struct {
		predicate func(T) bool
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want int
	}
	tests := []testCase[myStruct]{
		{
			name: "find/0",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return ele == myStruct{"a"}
				},
			},
			want: 0,
		},
		{
			name: "find/3",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return ele == myStruct{"d"}
				},
			},
			want: 3,
		},
		{
			name: "find/-1(not found)",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
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

	type args[T any] struct {
		predicate func(T) bool
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want int
	}
	tests := []testCase[myStruct]{
		{
			name: "find/0",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return ele == myStruct{"a"}
				},
			},
			want: 0,
		},
		{
			name: "find/3",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return ele == myStruct{"d"}
				},
			},
			want: 3,
		},
		{
			name: "find/2",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return ele == myStruct{"b"}
				},
			},
			want: 2,
		},
		{
			name: "find/-1(not found)",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
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
		reducer func(acc, ele T) T
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want T
	}
	tests := []testCase[myStruct]{
		{
			name: "fold empty",
			s:    FromSlice(emptySlice),
			args: args[myStruct]{
				init: myStruct{"!"},
				reducer: func(acc, ele myStruct) myStruct {
					return myStruct{acc.Name + ele.Name}
				},
			},
			want: myStruct{"!"},
		},
		{
			name: "fold",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				init: myStruct{"!"},
				reducer: func(acc, ele myStruct) myStruct {
					return myStruct{acc.Name + ele.Name}
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

func TestStream_FoldAny_As_Array(t *testing.T) {
	var emptySlice []myStruct

	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
	}

	type args[T any] struct {
		init    []T
		reducer func(acc any, ele T) any
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "fold empty list",
			s:    FromSlice(emptySlice),
			args: args[myStruct]{
				init: []myStruct{},
				reducer: func(acc any, ele myStruct) any {
					return append(acc.([]myStruct), ele)
				},
			},
			want: []myStruct{},
		},
		{
			name: "fold empty list with init value",
			s:    FromSlice(emptySlice),
			args: args[myStruct]{
				init: []myStruct{{"!"}},
				reducer: func(acc any, ele myStruct) any {
					return append(acc.([]myStruct), ele)
				},
			},
			want: []myStruct{{"!"}},
		},
		{
			name: "fold",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				init: []myStruct{},
				reducer: func(acc any, ele myStruct) any {
					return append(acc.([]myStruct), ele)
				},
			},
			want: []myStruct{{"a"}, {"b"}, {"c"}, {"d"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.s.FoldAny(tt.args.init, tt.args.reducer); !reflect.DeepEqual(gotResult, tt.want) {
				t.Errorf("FoldAny() = %v, want %v", gotResult, tt.want)
			}
		})
	}
}

func TestStream_Reduce(t *testing.T) {
	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
	}

	type args[T any] struct {
		reducer func(acc, ele T) T
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want T
	}

	tests := []testCase[myStruct]{
		//{
		//	name: "reduce empty return nil with no type",
		//	s:    FromSlice(emptySlice),
		//	args: args[myStruct]{
		//		reducer: func(acc any, ele any) any {
		//			// won't be called
		//			return myStruct{acc.(myStruct).Name + ele.(myStruct).Name}
		//		},
		//	},
		//	want: nilStruct, // this should be nil to interface{}
		//},
		{
			name: "reduce 1",
			s:    FromSlice(arr1[:1]),
			args: args[myStruct]{
				reducer: func(acc, ele myStruct) myStruct {
					return myStruct{acc.Name + ele.Name}
				},
			},
			want: myStruct{"a"},
		},
		{
			name: "reduce",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				reducer: func(acc, ele myStruct) myStruct {
					return myStruct{acc.Name + ele.Name}
				},
			},
			want: myStruct{"abcd"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.s.Reduce(tt.args.reducer); !reflect.DeepEqual(gotResult, tt.want) {
				t.Errorf("Reduce() = %v, want %v", gotResult, tt.want)
			}
		})
	}
}

func TestStream_Reduce_empty(t *testing.T) {
	var emptySlice []myStruct

	type args[T any] struct {
		reducer func(acc, ele T) T
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want T
	}

	tests := []testCase[myStruct]{
		{
			name: "reduce empty return empty",
			s:    FromSlice(emptySlice),
			args: args[myStruct]{
				reducer: func(acc, ele myStruct) myStruct {
					// won't be called
					return myStruct{acc.Name + ele.Name}
				},
			},
			want: myStruct{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.s.Reduce(tt.args.reducer); !reflect.DeepEqual(gotResult, tt.want) {
				t.Errorf("Reduce() = %v, want %v", gotResult, tt.want)
			}
		})
	}
}

func Test_baseStream_Map(t *testing.T) {

	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
		{"e"},
	}

	type args[T any] struct {
		mapf func(T) T
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "map zero",
			s:    FromSlice([]myStruct{}),
			args: args[myStruct]{
				mapf: func(ele myStruct) myStruct {
					return myStruct{
						Name: ele.Name + "!",
					}
				},
			},
			want: []myStruct{},
		},
		{
			name: "map as-is",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				mapf: func(ele myStruct) myStruct {
					return ele
				},
			},
			want: arr1,
		},
		{
			name: "map copy",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				mapf: func(ele myStruct) myStruct {
					return myStruct{
						Name: ele.Name,
					}
				},
			},
			want: arr1,
		},
		{
			name: "map title!",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				mapf: func(ele myStruct) myStruct {
					return myStruct{
						Name: ele.Name + "!",
					}
				},
			},
			want: []myStruct{
				{"a!"},
				{"b!"},
				{"c!"},
				{"d!"},
				{"e!"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Map(tt.args.mapf).Collect(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_baseStream_MapAny(t *testing.T) {
	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
		{"e"},
	}

	arr1Names := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
	}

	type args[T any] struct {
		mapf func(T) any
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want []string
	}
	tests := []testCase[myStruct]{
		{
			name: "mapany zero",
			s:    FromSlice([]myStruct{}),
			args: args[myStruct]{
				mapf: func(ele myStruct) any {
					return ele.Name
				},
			},
			want: []string{},
		},
		{
			name: "map as-is",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				mapf: func(ele myStruct) any {
					return ele.Name
				},
			},
			want: arr1Names,
		},
		{
			name: "map title!",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				mapf: func(ele myStruct) any {
					return ele.Name + "!"
				},
			},
			want: []string{
				"a!",
				"b!",
				"c!",
				"d!",
				"e!",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CollectAs[string](tt.s.MapAny(tt.args.mapf)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapAny() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_baseStream_CollectTo(t *testing.T) {

	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
		{"e"},
	}

	type args[T any] struct {
		target []T
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "nil target",
			s:    FromSlice([]myStruct{}),
			args: args[myStruct]{
				target: nil,
			},
			want: []myStruct{},
		},
		{
			name: "zero/empty",
			s:    FromSlice([]myStruct{}),
			args: args[myStruct]{
				target: []myStruct{},
			},
			want: []myStruct{},
		},
		{
			name: "collect as-is",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				target: []myStruct{},
			},
			want: arr1,
		},
		{
			name: "collect to not empty target",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				target: []myStruct{
					{"a"},
				},
			},
			want: func() (result []myStruct) {
				result = append(result, myStruct{"a"})
				result = append(result, arr1...)
				return
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.CollectTo(tt.args.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CollectTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_baseStream_All(t *testing.T) {
	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
		{"e"},
	}

	type args[T any] struct {
		predicate func(T) bool
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want bool
	}
	tests := []testCase[myStruct]{
		{
			name: "empty source - false",
			s:    FromSlice([]myStruct{}),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return true
				},
			},
			want: false,
		},
		{
			name: "all true - true",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return true
				},
			},
			want: true,
		},
		{
			name: "some true - false",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return ele.Name == "e"
				},
			},
			want: false,
		},
		{
			name: "some false - false",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return ele.Name == "e"
				},
			},
			want: false,
		},
		{
			name: "all false - false",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return false
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.All(tt.args.predicate); got != tt.want {
				t.Errorf("All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_baseStream_Any(t *testing.T) {
	arr1 := []myStruct{
		{"a"},
		{"b"},
		{"c"},
		{"d"},
		{"e"},
	}

	type args[T any] struct {
		predicate func(T) bool
	}
	type testCase[T any] struct {
		name string
		s    Stream[T]
		args args[T]
		want bool
	}
	tests := []testCase[myStruct]{
		{
			name: "empty source - false",
			s:    FromSlice([]myStruct{}),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return true
				},
			},
			want: false,
		},
		{
			name: "all true - true",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return true
				},
			},
			want: true,
		},
		{
			name: "some true - true",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return ele.Name == "e"
				},
			},
			want: true,
		},
		{
			name: "some false - true",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return ele.Name == "e"
				},
			},
			want: true,
		},
		{
			name: "all false - false",
			s:    FromSlice(arr1),
			args: args[myStruct]{
				predicate: func(ele myStruct) bool {
					return false
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Any(tt.args.predicate); got != tt.want {
				t.Errorf("Any() = %v, want %v", got, tt.want)
			}
		})
	}
}
