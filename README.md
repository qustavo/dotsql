dotsql 
[![GoDoc](https://godoc.org/github.com/gchaincl/dotsql?status.svg)](https://godoc.org/github.com/gchaincl/dotsql)
[![Build Status](https://travis-ci.org/gchaincl/dotsql.svg)](https://travis-ci.org/gchaincl/dotsql)
[![Test coverage](http://gocover.io/_badge/github.com/gchaincl/dotsql)](https://gocover.io/github.com/gchaincl/dotsql)
[![Go Report Card](https://goreportcard.com/badge/github.com/gchaincl/dotsql)](https://goreportcard.com/report/github.com/gchaincl/dotsql)
======

A Golang library for using SQL.

It is not an ORM, it is not a query builder. Dotsql is a library that helps you
keep sql files in one place and use it with ease.

_Dotsql is heavily inspired by_ [yesql](https://github.com/krisajenkins/yesql).

Installation
--
```bash
$ go get github.com/gchaincl/dotsql
```

Usage 
--

First of all, you need to define queries inside your sql file:

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
this is needed to be able to uniquely identify each query
inside dotsql.

With your sql file prepared, you can load it up and start utilizing your queries:

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

You can also merge multiple dotsql instances created from different sql file inputs:
```go
dot1, err := dotsql.LoadFromFile("queries1.sql")
dot2, err := dotsql.LoadFromFile("queries2.sql")
dot := dotsql.Merge(dot1, dot2)
```

Embeding
--
To avoid distributing `sql` files alongside the binary file, you will need to use tools like 
[gotic](https://github.com/gchaincl/gotic) to embed / pack everything into one file.

TODO
--
- [ ] Enable text interpolation inside queries using `text/template`
- [ ] `sqlx` integration
