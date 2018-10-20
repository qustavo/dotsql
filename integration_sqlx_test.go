// +build integrationwithsqlx
// +build integrationwithsqlite

package dotsql

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mxk/go-sqlite/sqlite3"
	"testing"
)

type user struct {
	Id    int
	Name  string
	Email string
}

func initDotSqlx() (*sqlx.DB, *DotSql) {
	dbx, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	dot, err := LoadFromFile("test_schema.sql")
	if err != nil {
		panic(err)
	}

	dot.MustExec(dbx, "create-users-table")

	return dbx, dot
}

func TestCreateTable2(t *testing.T) {
	dbx, _ := initDotSqlx()
	defer dbx.Close()

	row := dbx.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='users'",
	)

	var table string
	row.Scan(&table)

	if table != "users" {
		t.Errorf("Table 'users' has not been created")
	}
}

func TestInserts2(t *testing.T) {
	dbx, dotsql := initDotSqlx()
	defer dbx.Close()

	dotsql.MustExec(dbx, "create-user", "Foo Bar", "foo@bar.com")

	row := dbx.QueryRow(
		"SELECT email FROM users LIMIT 1",
	)

	var email string
	row.Scan(&email)

	if email != "foo@bar.com" {
		t.Errorf("Expect to find user with email == %s, got %s", "foo@bar.com", email)
	}
}

func TestDotSql_Get(t *testing.T) {
	dbx, dotsql := initDotSqlx()
	defer dbx.Close()

	dbx.MustExec("INSERT INTO users (name, email) VALUES('Foo Bar', 'foo@bar.com')")

	user := struct {
		Id    int
		Name  sql.NullString
		Email sql.NullString
	}{}

	if err := dotsql.Get(dbx, &user, "find-one-user-by-email", "foo@bar.com"); err != nil {
		panic(err)
	}

	if !user.Email.Valid || user.Email.String != "foo@bar.com" {
		t.Errorf("Expect to find user with email == %s, got %s", "foo@bar.com", user.Email.String)
	}
}

func TestDotSql_Select(t *testing.T) {
	dbx, dotsql := initDotSqlx()
	defer dbx.Close()

	rawUsers := map[string]user{"Foo Bar": user{Name: "Foo Bar", Email: "foo@bar.com"}, "Foo Bar2": user{Name: "Foo Bar2", Email: "foo2@bar.com"}}

	containAllUser := func(users []user) bool {
		if len(users) != len(rawUsers) {
			return false
		}
		for _, user := range users {
			rawUser, ok := rawUsers[user.Name]
			if !ok {
				return false
			}
			if rawUser.Name != user.Name || rawUser.Email != user.Email {
				return false
			}
		}
		return true
	}

	dbx.MustExec("INSERT INTO users (name, email) VALUES('Foo Bar', 'foo@bar.com')")
	dbx.MustExec("INSERT INTO users (name, email) VALUES('Foo Bar2', 'foo2@bar.com')")

	users := make([]user, 0)

	if err := dotsql.Select(dbx, &users, "get-user-list"); err != nil {
		panic(err)
	}

	if !containAllUser(users) {
		t.Errorf("Expect to get all users, but got %#v", users)
	}

}
