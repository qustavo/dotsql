// +build integration

package dotsql

import (
	"database/sql"
	"testing"

	_ "github.com/mxk/go-sqlite/sqlite3"
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
		t.Fatal(err)
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
		t.Fatal(err)
	}

	rows, err := dotsql.Query(db, "find-one-user-by-email", "foo@bar.com")
	if err != nil {
		t.Fatal(err)
	}

	ok := rows.Next()
	if ok == false {
		t.Errorf("User with email 'foo@bar.com' cound not be found")
	}

	var id, name interface{}
	var email string
	err = rows.Scan(&id, &name, &email)
	if err != nil {
		t.Fatal(err)
	}

	if email != "foo@bar.com" {
		t.Errorf("Expect to find user with email == %s, got %s", "foo@bar.com", email)
	}
}

func TestSelectOne(t *testing.T) {
	db, dotsql := initDotSql()
	defer db.Close()

	_, err := db.Exec("INSERT INTO users(email) VALUES('foo@bar.com')")
	if err != nil {
		t.Fatal(err)
	}

	row, err := dotsql.QueryRow(db, "find-one-user-by-email", "foo@bar.com")
	if err != nil {
		t.Fatal(err)
	}

	var id, name interface{}
	var email string
	err = row.Scan(&id, &name, &email)

	if err != nil {
		t.Errorf("User with email 'foo@bar.com' cound not be found")
	}

	if email != "foo@bar.com" {
		t.Errorf("Expect to find user with email == %s, got %s", "foo@bar.com", email)
	}
}

func countUsers(t *testing.T, dot *DotSql, db QueryRower, name string, expected int) {
	t.Helper()
	row, err := dot.QueryRow(db, name)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Error(err)
	}
	if expected != count {
		t.Errorf("expected %d users but got %d", expected, count)
	}
}

func TestSelectWithData(t *testing.T) {
	db, dotsql := initDotSql()
	defer db.Close()

	_, err := dotsql.Exec(db, "create-user", "foo", "foo@bar.com")
	if err != nil {
		t.Fatal(err)
	}
	_, err = dotsql.Exec(db, "create-user", "bar", "bar@bar.com")
	if err != nil {
		t.Fatal(err)
	}

	countUsers(t, dotsql, db, "count-users", 2)
	excludeDeleted := dotsql.WithData(map[string]any{"exclude_deleted": true})
	countUsers(t, &excludeDeleted, db, "count-users", 2)

	_, err = dotsql.Exec(db, "soft-delete-user", "foo@bar.com")
	if err != nil {
		t.Fatal(err)
	}

	countUsers(t, dotsql, db, "count-users", 2)
	countUsers(t, &excludeDeleted, db, "count-users", 1)
}
