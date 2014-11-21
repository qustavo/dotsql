// +build integration

package dotsql

import (
	"database/sql"
	"testing"

	_ "code.google.com/p/go-sqlite/go1/sqlite3"
)

func initDotSql() (*sql.DB, *DotSql) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	dotsql, err := LoadFile("test_schema.sql")
	if err != nil {
		panic(err)
	}

	_, err = dotsql.Exec(db, "migrate")
	if err != nil {
		panic(err)
	}

	return db, dotsql
}

func TestCreateTable(t *testing.T) {
	db, _ := initDotSql()
	defer db.Close()

	row := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='users'",
	)

	var table string
	row.Scan(&table)

	if table != "users" {
		t.Errorf("Table 'users' has not been created")
	}
}

func TestInserts(t *testing.T) {
	db, dotsql := initDotSql()
	defer db.Close()

	_, err := dotsql.Exec(db, "create-user", "Foo Bar", "foo@bar.com")
	if err != nil {
		panic(err)
	}

	row := db.QueryRow(
		"SELECT email FROM users LIMIT 1",
	)

	var email string
	row.Scan(&email)

	if email != "foo@bar.com" {
		t.Errorf("Expect to find user with email == %s, got %s", "foo@bar.com", email)
	}
}

func TestSelect(t *testing.T) {
	db, dotsql := initDotSql()
	defer db.Close()

	_, err := db.Exec("INSERT INTO users(email) VALUES('foo@bar.com')")
	if err != nil {
		panic(err)
	}

	rows, err := dotsql.Query(db, "find-one-user-by-email", "foo@bar.com")
	if err != nil {
		panic(err)
	}

	ok := rows.Next()
	if ok == false {
		t.Errorf("User with email 'foo@bar.com' cound not be found")
	}

	var id, name interface{}
	var email string
	err = rows.Scan(&id, &name, &email)
	if err != nil {
		panic(err)
	}

	if email != "foo@bar.com" {
		t.Errorf("Expect to find user with email == %s, got %s", "foo@bar.com", email)
	}
}

func TestLoad(t *testing.T) {
	dotsql1, err := LoadFile("test_schema.sql")
	if err != nil {
		t.Errorf("Failed to load queries from file")
	}
	if len(dotsql1.queries) != 3 {
		t.Errorf("Wrong number of queries loaded from file")
	}

	dotsql2, err := LoadString(testSchemaStr)
	if err != nil {
		t.Errorf("Failed to load queries from string")
	}
	if len(dotsql2.queries) != 3 {
		t.Errorf("Wrong number of queries loaded from string")
	}

	dotsql3, err := LoadBytes(testSchemaBytes)
	if err != nil {
		t.Errorf("Failed to load queries from file")
	}
	if len(dotsql3.queries) != 3 {
		t.Errorf("Wrong number of queries loaded from bytes")
	}
}

var (
	testSchemaStr = `-- name: migrate
CREATE TABLE users (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	name VARCHAR(255),
	email VARCHAR(255)
);

-- name: create-user
INSERT INTO users (name, email) VALUES(?, ?)

-- name: find-one-user-by-email
SELECT id,name,email FROM users WHERE email = ? LIMIT 1
`

	testSchemaBytes = []byte(testSchemaStr)
)
