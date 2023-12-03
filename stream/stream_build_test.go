package stream

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestFromSlice(t *testing.T) {
	var emptySlice []myStruct

	arr := []myStruct{
		{"a"},
		{"bb"},
		{"c"},
		{"ddd"},
		{"e"},
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
			name: "slice empty",
			args: args[myStruct]{
				s:      FromSlice(emptySlice),
				target: []myStruct{},
			},
			want: []myStruct{},
		},
		{
			name: "slice",
			args: args[myStruct]{
				s:      FromSlice(arr),
				target: []myStruct{},
			},
			want: []myStruct{{"a"}, {"bb"}, {"c"}, {"ddd"}, {"e"}},
		},
		{
			name: "slice fiter map",
			args: args[myStruct]{
				s: FromSlice(arr).Filter(func(v myStruct) bool {
					return len(v.Name) == 1
				}).Map(func(v myStruct) myStruct {
					return myStruct{v.Name + "!"}
				}),
				target: []myStruct{},
			},
			want: []myStruct{{"a!"}, {"c!"}, {"e!"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.s.Collect(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromVar(t *testing.T) {
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
			name: "variatic/empty",
			args: args[myStruct]{
				s:      FromVar[myStruct](),
				target: []myStruct{},
			},
			want: []myStruct{},
		},
		{
			name: "variatic",
			args: args[myStruct]{
				s:      FromVar(myStruct{"a"}, myStruct{"bb"}, myStruct{"c"}, myStruct{"ddd"}, myStruct{"e"}),
				target: []myStruct{},
			},
			want: []myStruct{{"a"}, {"bb"}, {"c"}, {"ddd"}, {"e"}},
		},
		{
			name: "variatic fiter map",
			args: args[myStruct]{
				s: FromVar(myStruct{"a"}, myStruct{"bb"}, myStruct{"c"}, myStruct{"ddd"}, myStruct{"e"}).
					Filter(func(v myStruct) bool {
						return len(v.Name) == 1
					}).Map(func(v myStruct) myStruct {
					return myStruct{v.Name + "!"}
				}),
				target: []myStruct{},
			},
			want: []myStruct{{"a!"}, {"c!"}, {"e!"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.s.Collect(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromVar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromChan(t *testing.T) {

	type args[T any] struct {
		ch     chan T
		source []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[myStruct]{
		{
			name: "chan nil",
			args: args[myStruct]{
				ch: nil,
			},
			want: []myStruct{},
		},
		{
			name: "chan empty",
			args: args[myStruct]{
				ch: make(chan myStruct, 0),
				source: func() []myStruct {
					return nil
				}(),
			},
			want: []myStruct{},
		},
		{
			name: "chan 1",
			args: args[myStruct]{
				ch:     make(chan myStruct, 0),
				source: []myStruct{{"a"}},
			},
			want: []myStruct{{"a"}},
		},
		{
			name: "chan n",
			args: args[myStruct]{
				ch:     make(chan myStruct, 0),
				source: []myStruct{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}},
			},
			want: []myStruct{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				if tt.args.ch != nil {
					for _, ele := range tt.args.source {
						tt.args.ch <- ele
					}
					close(tt.args.ch)
				}
			}()

			// github action filed
			// panic: send on closed chann
			// Give time to the goroutine to run before the collection
			time.Sleep(100 * time.Millisecond)

			if got := FromChan(tt.args.ch).Collect(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromChan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromChan_ReadWaitForSource(t *testing.T) {

	//var jobs chan int
	jobs := make(chan int, 0)
	done := make(chan bool, 0)

	go func() {
		s := FromChan(jobs)
		s.OnEach(func(ele int) {
			fmt.Printf("ele: %d -> ", ele)
		}).Map(func(ele int) int {
			return ele * 100
		}).ForEach(func(ele int) {
			fmt.Printf("got: %d\n", ele)
		})
		done <- true
	}()

	fmt.Println("Source: Start")
	for i := 0; i < 5; i++ {
		//time.Sleep(100 * time.Millisecond)
		fmt.Printf("Source: %d\n", i)
		jobs <- i
	}
	fmt.Println("Source: Done")
	// close job before waiting, to let receivers know the channel closed
	close(jobs)

	<-done
}

func TestFromChan_WriteReadSequentially(t *testing.T) {

	//var jobs chan int
	worker := 5
	jobs := make(chan int, worker)

	fmt.Println("Source: Start")
	for i := 0; i < worker; i++ {
		//time.Sleep(100 * time.Millisecond)
		fmt.Printf("Source: %d\n", i)
		jobs <- i
	}
	fmt.Println("Source: Done")

	// close channel before move on
	close(jobs)

	s := FromChan(jobs)
	s.OnEach(func(ele int) {
		fmt.Printf("ele: %d -> ", ele)
	}).Map(func(ele int) int {
		return ele * 100
	}).ForEach(func(ele int) {
		fmt.Printf("got: %d\n", ele)
	})

}

func TestFromChan_ReadOnClosedChannel(t *testing.T) {

	worker := 5
	jobs := make(chan int, 0)
	done := make(chan bool, 0)

	go func() {
		fmt.Println("Source: Start")
		for i := 0; i < worker; i++ {
			//time.Sleep(100 * time.Millisecond)
			fmt.Printf("Source: %d\n", i)
			jobs <- i
		}
		fmt.Println("Source: Done")
		// close channel before move on
		close(jobs)
	}()

	time.Sleep(1 * time.Second)

	go func() {
		s := FromChan(jobs)
		s.OnEach(func(ele int) {
			fmt.Printf("ele: %d -> ", ele)
		}).Map(func(ele int) int {
			return ele * 100
		}).ForEach(func(ele int) {
			fmt.Printf("got: %d\n", ele)
		})
		done <- true
	}()

	<-done
}

func TestFromChan_ReadWriteConcurrently(t *testing.T) {

	worker := 5
	jobs := make(chan int, 0)
	done := make(chan bool, 0)

	go func() {
		fmt.Println("Source: Start")
		for i := 0; i < worker; i++ {
			//time.Sleep(100 * time.Millisecond)
			fmt.Printf("Source: %d\n", i)
			jobs <- i
		}
		fmt.Println("Source: Done")
		// close channel before move on
		close(jobs)
	}()

	go func() {
		s := FromChan(jobs)
		s.OnEach(func(ele int) {
			fmt.Printf("ele: %d -> ", ele)
		}).Map(func(ele int) int {
			return ele * 100
		}).ForEach(func(ele int) {
			fmt.Printf("got: %d\n", ele)
		})
		done <- true
	}()

	<-done
}
