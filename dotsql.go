// Package dotsql provides a way to separate your code from SQL queries.
//
// It is not an ORM, it is not a query builder.
// Dotsql is a library that helps you keep sql files in one place and use it with ease.
//
// For more usage examples see https://github.com/qustavo/dotsql
package dotsql

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"text/template"
)

// Preparer is an interface used by Prepare.
type Preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

// PreparerContext is an interface used by PrepareContext.
type PreparerContext interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Queryer is an interface used by Query.
type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// QueryerContext is an interface used by QueryContext.
type QueryerContext interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// QueryRower is an interface used by QueryRow.
type QueryRower interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

// QueryRowerContext is an interface used by QueryRowContext.
type QueryRowerContext interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Execer is an interface used by Exec.
type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// ExecerContext is an interface used by ExecContext.
type ExecerContext interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// DotSql represents a dotSQL queries holder.
type DotSql struct {
	queries map[string]*template.Template
	data    any
}

func (d DotSql) WithData(data any) DotSql {
	return DotSql{queries: d.queries, data: data}
}

func (d DotSql) lookupQuery(name string, data any) (string, error) {
	template, ok := d.queries[name]
	if !ok {
		return "", fmt.Errorf("dotsql: '%s' could not be found", name)
	}
	if template == nil {
		return "", nil
	}
	buffer := bytes.NewBufferString("")
	err := template.Execute(buffer, data)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	return buffer.String(), nil
}

// Prepare is a wrapper for database/sql's Prepare(), using dotsql named query.
func (d DotSql) Prepare(db Preparer, name string) (*sql.Stmt, error) {
	query, err := d.lookupQuery(name, d.data)
	if err != nil {
		return nil, err
	}

	return db.Prepare(query)
}

// PrepareContext is a wrapper for database/sql's PrepareContext(), using dotsql named query.
func (d DotSql) PrepareContext(ctx context.Context, db PreparerContext, name string) (*sql.Stmt, error) {
	query, err := d.lookupQuery(name, d.data)
	if err != nil {
		return nil, err
	}

	return db.PrepareContext(ctx, query)
}

// Query is a wrapper for database/sql's Query(), using dotsql named query.
func (d DotSql) Query(db Queryer, name string, args ...interface{}) (*sql.Rows, error) {
	query, err := d.lookupQuery(name, d.data)
	if err != nil {
		return nil, err
	}

	return db.Query(query, args...)
}

// QueryContext is a wrapper for database/sql's QueryContext(), using dotsql named query.
func (d DotSql) QueryContext(ctx context.Context, db QueryerContext, name string, args ...interface{}) (*sql.Rows, error) {
	query, err := d.lookupQuery(name, d.data)
	if err != nil {
		return nil, err
	}

	return db.QueryContext(ctx, query, args...)
}

// QueryRow is a wrapper for database/sql's QueryRow(), using dotsql named query.
func (d DotSql) QueryRow(db QueryRower, name string, args ...interface{}) (*sql.Row, error) {
	query, err := d.lookupQuery(name, d.data)
	if err != nil {
		return nil, err
	}

	return db.QueryRow(query, args...), nil
}

// QueryRowContext is a wrapper for database/sql's QueryRowContext(), using dotsql named query.
func (d DotSql) QueryRowContext(ctx context.Context, db QueryRowerContext, name string, args ...interface{}) (*sql.Row, error) {
	query, err := d.lookupQuery(name, d.data)
	if err != nil {
		return nil, err
	}

	return db.QueryRowContext(ctx, query, args...), nil
}

// Exec is a wrapper for database/sql's Exec(), using dotsql named query.
func (d DotSql) Exec(db Execer, name string, args ...interface{}) (sql.Result, error) {
	query, err := d.lookupQuery(name, d.data)
	if err != nil {
		return nil, err
	}

	return db.Exec(query, args...)
}

// ExecContext is a wrapper for database/sql's ExecContext(), using dotsql named query.
func (d DotSql) ExecContext(ctx context.Context, db ExecerContext, name string, args ...interface{}) (sql.Result, error) {
	query, err := d.lookupQuery(name, d.data)
	if err != nil {
		return nil, err
	}

	return db.ExecContext(ctx, query, args...)
}

// Raw returns the query, everything after the --name tag
func (d DotSql) Raw(name string) (string, error) {
	return d.lookupQuery(name, d.data)
}

// QueryMap returns a map[string]string of loaded queries
func (d DotSql) QueryMap() map[string]*template.Template {
	return d.queries
}

// Load imports sql queries from any io.Reader.
func Load(r io.Reader) (*DotSql, error) {
	scanner := &Scanner{}
	queries := scanner.Run(bufio.NewScanner(r))

	templates := make(map[string]*template.Template)
	for k, v := range queries {
		tmpl, err := template.New(k).Parse(v)
		if err != nil {
			return nil, err
		}
		templates[k] = tmpl
	}

	return &DotSql{
		queries: templates,
	}, nil
}

// LoadSize imports sql queries from any io.Reader and use maxSize.
func LoadSize(r io.Reader, maxSize int) (*DotSql, error) {
	if maxSize < bufio.MaxScanTokenSize {
		maxSize = bufio.MaxScanTokenSize
	}
	// ref: bufio.Scan_test.go func TestHugeBuffer(t *testing.T)
	hugeBuffer := bufio.NewScanner(bufio.NewReader(r))
	hugeBuffer.Buffer(make([]byte, 4096), maxSize)
	scanner := &Scanner{}
	queries := scanner.Run(hugeBuffer)

	dotsql := &DotSql{
		queries: queries,
	}

	return dotsql, nil
}

// LoadFromFile imports SQL queries from the file.
func LoadFromFile(sqlFile string) (*DotSql, error) {
	f, err := os.Open(sqlFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Load(f)
}

// LoadFromFileSize imports SQL queries from the file and use maxSize.
func LoadFromFileSize(sqlFile string, maxSize int) (*DotSql, error) {
	f, err := os.Open(sqlFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return LoadSize(f, maxSize)
}

// LoadFromString imports SQL queries from the string.
func LoadFromString(sql string) (*DotSql, error) {
	buf := bytes.NewBufferString(sql)
	return Load(buf)
}

// Merge takes one or more *DotSql and merge its queries
// It's in-order, so the last source will override queries with the same name
// in the previous arguments if any.
func Merge(dots ...*DotSql) *DotSql {
	queries := make(map[string]*template.Template)

	for _, dot := range dots {
		for k, v := range dot.QueryMap() {
			queries[k] = v
		}
	}

	return &DotSql{
		queries: queries,
	}
}
