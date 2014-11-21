package dotsql

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
)

type Preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type DotSql struct {
	filePath string
	queries  map[string]string
}

func (self DotSql) lookupQuery(name string) (query string, err error) {
	query, ok := self.queries[name]
	if !ok {
		err = fmt.Errorf("dotsql: '%s' could not be found on '%s'",
			name, self.filePath)
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

func (self DotSql) Query(db Queryer, name string, args ...interface{}) (*sql.Rows, error) {
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

func Load(sqlFile string) (dotsql *DotSql, err error) {
	f, err := os.Open(sqlFile)
	if err != nil {
		return
	}
	defer f.Close()

	dotsql = &DotSql{
		filePath: sqlFile,
		queries:  scanQueries(f),
	}

	return
}

func scanQueries(f *os.File) map[string]string {
	scanner := &Scanner{}
	return scanner.Run(bufio.NewScanner(f))
}