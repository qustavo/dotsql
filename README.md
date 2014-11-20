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

Usage
--

First of all, you need to define queries into a file:

```sql
-- name: migrate
CREATE TABLE users (
id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
name VARCHAR(255),
email VARCHAR(255)
);
-- name: create-user
INSERT INTO users (name, email) VALUES(?, ?)
-- name: find-one-user-by-email
SELECT id,name,email FROM users WHERE email = ? LIMIT 1
```

Notice that every query has a name tag (`--name:<some name>`),
this will be helpful for referring to a specific query

Then you should be able to run something like:

```go
// Get a database handle
db, err := sql.Open("sqlite3", ":memory:")

// Loads queries from file
dot, err := dotsql.Load("queries.sql")

// Run queries
res, err := dot.Exec(db, "migrate")
res, err := dot.Exec(db, "create-user", "User Name", "main@example.com")
rows, err := dot.Query(db, "find-one-user-by-email", "main@example.com")
```

For a complete example please refer to [dotsql_test.go](https://github.com/gchaincl/dotsql/blob/master/dotsql_test.go) and [test_schema.sql](https://github.com/gchaincl/dotsql/blob/master/test_schema.sql)

Development
--

Dotsql is in a very early stage so api may change. Contributions are welcome!
Integration tests are tagged with `+integration`, so if you want to run them you should:
```bash
go test -tags=integration
```

Otherwise just run:
```bash
go test
```
