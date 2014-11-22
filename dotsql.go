package dotsql

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
)

type Preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

type Querier interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type DotSql struct {
	queries map[string]string
}

func (self DotSql) lookupQuery(name string) (query string, err error) {
	query, ok := self.queries[name]
	if !ok {
		err = fmt.Errorf("dotsql: '%s' could not be found", name)
	}

	return
}

func (self DotSql) Prepare(db Preparer, name string) (*sql.Stmt, error) {
	query, err := self.lookupQuery(name)
	if err != nil {
		return nil, err
	}

	return db.Prepare(query)
}

func (self DotSql) Query(db Querier, name string, args ...interface{}) (*sql.Rows, error) {
	query, err := self.lookupQuery(name)
	if err != nil {
		return nil, err
	}

	return db.Query(query, args...)
}

func (self DotSql) Exec(db Execer, name string, args ...interface{}) (sql.Result, error) {
	query, err := self.lookupQuery(name)
	if err != nil {
		return nil, err
	}

	return db.Exec(query, args...)
}

func Load(r io.Reader) (*DotSql, error) {
	scanner := &Scanner{}
	queries := scanner.Run(bufio.NewScanner(r))

	dotsql := &DotSql{
		queries: queries,
	}

	return dotsql, nil
}

func LoadFromFile(sqlFile string) (*DotSql, error) {
	f, err := os.Open(sqlFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Load(f)
}

func LoadFromString(sql string) (*DotSql, error) {
	buf := bytes.NewBufferString(sql)
	return Load(buf)
}
