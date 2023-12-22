package main

import (
	"database/sql"
	"fmt"
	"log"

	// To import a package solely for its side-effects (initialization), use the blank identifier as explicit package name:
	// In the case of go-sqlite3, the underscore import is used for the side-effect of registering the sqlite3 driver as a database driver in the init() function, without importing any other functions:
	// Once it's registered in this way, sqlite3 can be used with the standard library's sql interface in your code like in the example:
	_ "github.com/mattn/go-sqlite3"
)

type row struct {
	_type       string
	coordinates string
}

type sqliteSource struct {
	conn string
	db   *sql.DB
	rows *sql.Rows
}

func (c *sqliteSource) Next() bool {
	if c == nil || c.db == nil {
		return false
	}

	if c.rows == nil {
		rows, err := c.db.Query("SELECT * FROM geometry")
		if err != nil {
			return false
		}
		if rows == nil {
			return false
		}
		c.rows = rows
		//return rows.Next()
	}

	return c.rows.Next()
}

func (c *sqliteSource) Get() *row {
	if c == nil || c.db == nil || c.rows == nil {
		return nil
	}
	// {"type":"Point","coordinates":[127.0604505,37.5079355]}
	var row row

	if err := c.rows.Scan(&row._type, &row.coordinates); err == sql.ErrNoRows {
		log.Printf("Id not found")
		return nil
	}
	//log.Printf("query: %v, %s\n", _type, coordinates)
	return &row
}

func (c *sqliteSource) Close() {
	if c == nil || c.db == nil {
		return
	}
	_ = c.db.Close()
	c.db = nil
}

func FromConnectString(conn string) *sqliteSource {
	source := &sqliteSource{}

	db, err := sql.Open("sqlite3", conn)
	if err != nil {
		panic(err)
	}
	source.conn = conn
	source.db = db
	return source
}

func main() {
	fmt.Println("hello")
	source := FromConnectString("geojson.db")
	defer source.Close()

	for source.Next() {
		log.Println("element", source.Get())
	}
}
