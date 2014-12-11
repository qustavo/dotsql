package dotsql_test

import (
	"database/sql"
	"github.com/gchaincl/dotsql"
)

func ExampleLoadFromFile() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	dot, err := dotsql.LoadFromFile("queries/users.sql")
	if err != nil {
		panic(err)
	}

	/* users.sql looks like:
	--name: create-user
	INSERT INTO users (email) VALUES(?)

	--name: find-user-by-email
	SELECT email FROM users WHERE email = ?
	*/

	if _, err := dot.Exec(db, "create-user", "user@example.com"); err != nil {
		panic(err)
	}

	rows, err := dot.Query(db, "find-user-by-email", "user@example.com")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
}
