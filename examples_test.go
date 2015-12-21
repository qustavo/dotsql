package dotsql_test

import (
	"database/sql"
	"log"

	"github.com/gchaincl/dotsql"
)

func ExampleLoadFromFile() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	dot, err := dotsql.LoadFromFile("queries.sql")
	if err != nil {
		panic(err)
	}

	/* queries.sql looks like:
	-- name: create-users-table
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		name VARCHAR(255),
		email VARCHAR(255)
	);

	-- name: create-user
	INSERT INTO users (name, email) VALUES(?, ?)

	-- name: find-users-by-email
	SELECT id,name,email FROM users WHERE email = ?

	-- name: find-one-user-by-email
	SELECT id,name,email FROM users WHERE email = ? LIMIT 1

	--name: drop-users-table
	DROP TABLE users
	*/

	// Exec
	if _, err := dot.Exec(db, "create-users-table"); err != nil {
		panic(err)
	}

	if _, err := dot.Exec(db, "create-user", "user@example.com"); err != nil {
		panic(err)
	}

	// Query
	rows, err := dot.Query(db, "find-user-by-email", "user@example.com")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string

		err = rows.Scan(&id, &name, &email)

		if err != nil {
			panic(err)
		}

		log.Println(email)
	}

	// QueryRow
	row, err := dot.QueryRow(db, "find-one-user-by-email", "user@example.com")
	if err != nil {
		panic(err)
	}

	var id int
	var name, email string
	row.Scan(&id, &name, &email)

	log.Println(email)
}
