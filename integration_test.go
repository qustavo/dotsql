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

	dotsql, err := LoadFromFile("test_schema.sql")
	if err != nil {
		panic(err)
	}

	_, err = dotsql.Exec(db, "create-users-table")
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
