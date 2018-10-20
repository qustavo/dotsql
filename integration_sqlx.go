// +build integrationwithsqlx

package dotsql

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/jmoiron/sqlx"
)

type Getter interface {
	Get(dst interface{}, query string, args ...interface{}) error
}

type Selecter interface {
	Select(dst interface{}, query string, args ...interface{}) error
}

type Queryerx interface {
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
}

type QueryRowerx interface {
	QueryRowx(query string, args ...interface{}) *sqlx.Row
}

type MustExecer interface {
	MustExec(query string, args ...interface{}) sql.Result
}

type NameBinder interface {
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
}
type Rebinder interface {
	Rebind(query string) string
}

type NamedQueryer interface {
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
}

type NamedExecer interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
}

type NamedPreparer interface {
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
}

func (d DotSql) Get(dbx Getter, dst interface{}, name string, args ...interface{}) error {
	query, err := d.lookupQuery(name)
	if err != nil {
		return err
	}
	return dbx.Get(dst, query, args...)
}

func (d DotSql) Select(dbx Selecter, dst interface{}, name string, args ...interface{}) error {
	query, err := d.lookupQuery(name)
	if err != nil {
		return err
	}
	return dbx.Select(dst, query, args...)
}

func (d DotSql) Queryx(dbx Queryerx, name string, args ...interface{}) (*sqlx.Rows, error) {
	query, err := d.lookupQuery(name)
	if err != nil {
		return nil, err
	}
	return dbx.Queryx(query, args...)
}

func (d DotSql) QueryRowx(dbx QueryRowerx, name string, args ...interface{}) (*sqlx.Row, error) {
	query, err := d.lookupQuery(name)
	if err != nil {
		return nil, err
	}
	return dbx.QueryRowx(query, args...), nil
}

func (d DotSql) MustExec(dbx MustExecer, name string, args ...interface{}) sql.Result {
	query, err := d.lookupQuery(name)
	if err != nil {
		panic(err) //
	}
	return dbx.MustExec(query, args...)
}

func (d DotSql) BindNamed(dbx NameBinder, name string, arg interface{}) (string, []interface{}, error) {
	query, err := d.lookupQuery(name)
	if err != nil {
		return "", []interface{}{}, err
	}
	return dbx.BindNamed(query, arg)
}

func (d DotSql) NamedQuery(dbx NamedQueryer, name string, arg interface{}) (*sqlx.Rows, error) {
	query, err := d.lookupQuery(name)
	if err != nil {
		return nil, err
	}
	return dbx.NamedQuery(query, arg)
}

func (d DotSql) NamedExec(dbx NamedExecer, name string, arg interface{}) (sql.Result, error) {
	query, err := d.lookupQuery(name)
	if err != nil {
		return nil, err
	}
	return dbx.NamedExec(query, arg)
}

func (d DotSql) PrepareNamed(dbx NamedPreparer, name string) (*sqlx.NamedStmt, error) {
	query, err := d.lookupQuery(name)
	if err != nil {
		return nil, err
	}
	return dbx.PrepareNamed(query)
}

func (d DotSql) In(name string, args ...interface{}) (string, []interface{}, error) {
	query, err := d.lookupQuery(name)
	if err != nil {
		return "", []interface{}{}, err
	}
	return sqlx.In(query, args...)
}

func (d DotSql) Rebind(dbx Rebinder, name string) string {
	query, err := d.lookupQuery(name)
	if err != nil {
		return ""
	}
	return dbx.Rebind(query)
}
