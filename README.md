dotsql [![Build Status](https://travis-ci.org/gchaincl/dotsql.svg)](https://travis-ci.org/gchaincl/dotsql)
======

A Golang library for using SQL.

It is not an ORM, it is not a query builder. Dotsql is a library that helps you
keep sql files in one place and use it with ease.

_Dotsql is heavily inspired by_ [yesql](https://github.com/krisajenkins/yesql).

Installation
--
Simple install the package to your `$GOPATH` with the `go` tool from shell:
```bash
$ go get github.com/gchaincl/dotsql
```
Make sure Git is installed on your machine and in your system's `$PATH`

Usage [![GoDoc](https://godoc.org/github.com/gchaincl/dotsql?status.svg)](https://godoc.org/github.com/gchaincl/dotsql)
--

First of all, you need to define queries into a file:

```sql
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
```

Notice that every query has a name tag (`--name:<some name>`),
this will be helpful for referring to a specific query

Then you should be able to run something like:

```go
// Get a database handle
db, err := sql.Open("sqlite3", ":memory:")

// Loads queries from file
dot, err := dotsql.LoadFromFile("queries.sql")

// Run queries
res, err := dot.Exec(db, "create-users-table")
res, err := dot.Exec(db, "create-user", "User Name", "main@example.com")
rows, err := dot.Query(db, "find-users-by-email", "main@example.com")
row, err := dot.QueryRow(db, "find-one-user-by-email", "user@example.com")

stmt, err := dot.Prepare(db, "drop-users-table")
result, err := stmt.Exec()
```

For a complete example please refer to [integration_test.go](https://github.com/gchaincl/dotsql/blob/master/integration_test.go) and [test_schema.sql](https://github.com/gchaincl/dotsql/blob/master/test_schema.sql)

Development
--

Dotsql is in a very early stage so api may change. Contributions are welcome!
Integration tests are tagged with `+integration`, so if you want to run them you should:
```bash
go test -tags=integration
```
_If  integration tests takes too long remember to_ `go install github.com/mxk/go-sqlite/sqlite3`

Otherwise just run:
```bash
go test
```

Embeding
--
To avoid distributing `sql` files with the binary, you will need to embed into it, tools like [gotic](https://github.com/gchaincl/gotic) may help

TODO
----

- [ ] Enable text interpolation inside queries using `text/template`
- [ ] `sqlx` integration
