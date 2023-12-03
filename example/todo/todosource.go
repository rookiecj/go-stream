package main

import (
	"fmt"
	s "github.com/rookiecj/go-stream/stream"
)

// From https://jsonplaceholder.typicode.com/posts

type Todo struct {
	userId int    `json:"userId"`
	id     int    `json:"id"`
	title  string `json:"title"`
	body   string `json:"body"`
	done   bool
}

var Todos = []Todo{
	{
		userId: 1,
		id:     1,
		title:  "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
		body:   "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nreprehenderit molestiae ut ut quas totam\nnostrum rerum est autem sunt rem eveniet architecto",
	},
	{
		userId: 1,
		id:     2,
		title:  "qui est esse",
		body:   "est rerum tempore vitae\nsequi sint nihil reprehenderit dolor beatae ea dolores neque\nfugiat blanditiis voluptate porro vel nihil molestiae ut reiciendis\nqui aperiam non debitis possimus qui neque nisi nulla",
	},
	{
		userId: 1,
		id:     3,
		title:  "ea molestias quasi exercitationem repellat qui ipsa sit aut",
		body:   "et iusto sed quo iure\nvoluptatem occaecati omnis eligendi aut ad\nvoluptatem doloribus vel accusantium quis pariatur\nmolestiae porro eius odio et labore et velit aut",
	},
	{
		userId: 1,
		id:     4,
		title:  "eum et est occaecati",
		body:   "ullam et saepe reiciendis voluptatem adipisci\nsit amet autem assumenda provident rerum culpa\nquis hic commodi nesciunt rem tenetur doloremque ipsam iure\nquis sunt voluptatem rerum illo velit",
	},
	{
		userId: 1,
		id:     5,
		title:  "nesciunt quas odio",
		body:   "repudiandae veniam quaerat sunt sed\nalias aut fugiat sit autem sed est\nvoluptatem omnis possimus esse voluptatibus quis\nest aut tenetur dolor neque",
	},
	{
		userId: 2,
		id:     11,
		title:  "et ea vero quia laudantium autem",
		body:   "delectus reiciendis molestiae occaecati non minima eveniet qui voluptatibus\naccusamus in eum beatae sit\nvel qui neque voluptates ut commodi qui incidunt\nut animi commodi",
	},
	{
		userId: 2,
		id:     12,
		title:  "in quibusdam tempore odit est dolorem",
		body:   "itaque id aut magnam\npraesentium quia et ea odit et ea voluptas et\nsapiente quia nihil amet occaecati quia id voluptatem\nincidunt ea est distinctio odio",
	},
	{
		userId: 2,
		id:     13,
		title:  "dolorum ut in voluptas mollitia et saepe quo animi",
		body:   "aut dicta possimus sint mollitia voluptas commodi quo doloremque\niste corrupti reiciendis voluptatem eius rerum\nsit cumque quod eligendi laborum minima\nperferendis recusandae assumenda consectetur porro architecto ipsum ipsam",
	},
	{
		userId: 2,
		id:     14,
		title:  "voluptatem eligendi optio",
		body:   "fuga et accusamus dolorum perferendis illo voluptas\nnon doloremque neque facere\nad qui dolorum molestiae beatae\nsed aut voluptas totam sit illum",
	},
}

type todoSource struct {
	idx   int
	todos []Todo
}

func newTodoSource(todos []Todo) *todoSource {
	return &todoSource{
		idx:   -1,
		todos: todos,
	}
}

func (c *todoSource) Next() bool {
	if c.idx+1 < len(c.todos) {
		c.idx++
		return true
	}
	return false
}

func (c *todoSource) Get() Todo {
	return c.todos[c.idx]
}

func main() {

	todoS := newTodoSource(Todos)
	length := s.FromSource[Todo](todoS).
		Filter(func(todo Todo) bool {
			return todo.userId == 1
		}).
		Count()
	fmt.Println("todos for user1 count=", length)

	todosource := newTodoSource(Todos)
	todos1 := s.FromSource[Todo](todosource).
		Filter(func(ele Todo) bool {
			return ele.userId == 1
		}).Collect()
	fmt.Println("len(todos for user1)=", len(todos1))

	todosource = newTodoSource(Todos)
	titlestream := s.FromSource[Todo](todosource).
		Filter(func(ele Todo) bool {
			return ele.userId == 2
		}).
		MapAny(func(ele Todo) any {
			return ele.title
		})
	titles := s.CollectAs[string](titlestream)
	for _, title := range titles {
		fmt.Println(title)
	}
}
