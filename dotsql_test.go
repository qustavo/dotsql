package dotsql

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"text/template"
)

func failIfError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("err == nil, got '%s'", err)
	}
}

func createTemplate(t *testing.T, content string) *template.Template {
	t.Helper()
	tmpl, err := template.New("get-users").Parse(content)
	if err != nil {
		t.Fatal(err)
	}
	return tmpl
}

func extractTemplate(t *testing.T, dot DotSql, name string) string {
	t.Helper()
	query, err := dot.Raw(name)
	if err != nil {
		t.Fatal(err)
	}
	return query
}

func executeTemplate(t *testing.T, tmpl *template.Template) string {
	t.Helper()
	buffer := bytes.NewBufferString("")
	err := tmpl.Execute(buffer, nil)
	if err != nil {
		t.Fatal(err)
	}
	return buffer.String()
}

func TestPrepare(t *testing.T) {
	dot := DotSql{
		queries: map[string]*template.Template{
			"select": createTemplate(t, "SELECT * from users"),
		},
	}

	preparerStub := func(err error) *PreparerMock {
		return &PreparerMock{
			PrepareFunc: func(_ string) (*sql.Stmt, error) {
				if err != nil {
					return nil, err
				}
				return &sql.Stmt{}, nil
			},
		}
	}

	// query not found
	p := preparerStub(nil)
	stmt, err := dot.Prepare(p, "insert")
	if stmt != nil {
		t.Error("stmt expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff := p.PrepareCalls()
	if len(ff) > 0 {
		t.Error("prepare was not expected to be called")
	}

	// error returned by db
	p = preparerStub(errors.New("critical error"))
	stmt, err = dot.Prepare(p, "select")
	if stmt != nil {
		t.Error("stmt expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff = p.PrepareCalls()
	tmpl := extractTemplate(t, dot, "select")
	if len(ff) != 1 {
		t.Errorf("prepare was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("prepare was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	}

	// successful query preparation
	p = preparerStub(nil)
	stmt, err = dot.Prepare(p, "select")
	if stmt == nil {
		t.Error("stmt expected to be non-nil, got nil")
	}

	failIfError(t, err)

	ff = p.PrepareCalls()
	if len(ff) != 1 {
		t.Errorf("prepare was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("prepare was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	}
}

func TestPrepareContext(t *testing.T) {
	dot := DotSql{
		queries: map[string]*template.Template{
			"select": createTemplate(t, "SELECT * from users"),
		},
	}

	preparerContextStub := func(err error) *PreparerContextMock {
		return &PreparerContextMock{
			PrepareContextFunc: func(_ context.Context, _ string) (*sql.Stmt, error) {
				if err != nil {
					return nil, err
				}
				return &sql.Stmt{}, nil
			},
		}
	}

	ctx := context.Background()

	// query not found
	p := preparerContextStub(nil)
	stmt, err := dot.PrepareContext(ctx, p, "insert")
	if stmt != nil {
		t.Error("stmt expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff := p.PrepareContextCalls()
	if len(ff) > 0 {
		t.Error("prepare was not expected to be called")
	}

	// error returned by db
	p = preparerContextStub(errors.New("critical error"))
	stmt, err = dot.PrepareContext(ctx, p, "select")
	if stmt != nil {
		t.Error("stmt expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff = p.PrepareContextCalls()
	tmpl := extractTemplate(t, dot, "select")
	if len(ff) != 1 {
		t.Errorf("prepare was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("prepare was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if !reflect.DeepEqual(ff[0].Ctx, ctx) {
		t.Errorf("prepare context does not match")
	}

	// successful query preparation
	p = preparerContextStub(nil)
	stmt, err = dot.PrepareContext(ctx, p, "select")
	if stmt == nil {
		t.Error("stmt expected to be non-nil, got nil")
	}

	failIfError(t, err)

	ff = p.PrepareContextCalls()
	if len(ff) != 1 {
		t.Errorf("prepare was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("prepare was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if !reflect.DeepEqual(ff[0].Ctx, ctx) {
		t.Errorf("prepare context does not match")
	}
}

func TestQuery(t *testing.T) {
	dot := DotSql{
		queries: map[string]*template.Template{
			"select": createTemplate(t, "SELECT * from users WHERE name = ?"),
		},
	}

	queryerStub := func(err error) *QueryerMock {
		return &QueryerMock{
			QueryFunc: func(_ string, _ ...interface{}) (*sql.Rows, error) {
				if err != nil {
					return nil, err
				}
				return &sql.Rows{}, nil
			},
		}
	}

	arg := "test"

	// query not found
	q := queryerStub(nil)
	rows, err := dot.Query(q, "insert", arg)
	if rows != nil {
		t.Error("rows expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff := q.QueryCalls()
	if len(ff) > 0 {
		t.Error("query was not expected to be called")
	}

	// error returned by db
	q = queryerStub(errors.New("critical error"))
	rows, err = dot.Query(q, "select", arg)
	if rows != nil {
		t.Error("rows expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff = q.QueryCalls()
	tmpl := extractTemplate(t, dot, "select")
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	}

	// successful query
	q = queryerStub(nil)
	rows, err = dot.Query(q, "select", arg)
	if rows == nil {
		t.Error("rows expected to be non-nil, got nil")
	}

	failIfError(t, err)

	ff = q.QueryCalls()
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	}
}

func TestQueryContext(t *testing.T) {
	dot := DotSql{
		queries: map[string]*template.Template{
			"select": createTemplate(t, "SELECT * from users WHERE name = ?"),
		},
	}

	queryerContextStub := func(err error) *QueryerContextMock {
		return &QueryerContextMock{
			QueryContextFunc: func(_ context.Context, _ string, _ ...interface{}) (*sql.Rows, error) {
				if err != nil {
					return nil, err
				}
				return &sql.Rows{}, nil
			},
		}
	}

	ctx := context.Background()
	arg := "test"

	// query not found
	q := queryerContextStub(nil)
	rows, err := dot.QueryContext(ctx, q, "insert", arg)
	if rows != nil {
		t.Error("rows expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff := q.QueryContextCalls()
	if len(ff) > 0 {
		t.Error("query was not expected to be called")
	}

	// error returned by db
	q = queryerContextStub(errors.New("critical error"))
	rows, err = dot.QueryContext(ctx, q, "select", arg)
	if rows != nil {
		t.Error("rows expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff = q.QueryContextCalls()
	tmpl := extractTemplate(t, dot, "select")
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	} else if !reflect.DeepEqual(ff[0].Ctx, ctx) {
		t.Errorf("query context does not match")
	}

	// successful query
	q = queryerContextStub(nil)
	rows, err = dot.QueryContext(ctx, q, "select", arg)
	if rows == nil {
		t.Error("rows expected to be non-nil, got nil")
	}

	failIfError(t, err)

	ff = q.QueryContextCalls()
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	} else if !reflect.DeepEqual(ff[0].Ctx, ctx) {
		t.Errorf("query context does not match")
	}
}

func TestQueryOptions(t *testing.T) {
	dot := DotSql{
		queries: map[string]*template.Template{
			"select": createTemplate(t, "SELECT * from users WHERE name = ?{{if  .exclude_deleted}} AND deleted IS NULL{{end}}"),
		},
	}
	dotExcludeDeleted := dot.WithOptions(map[string]interface{}{"exclude_deleted": true})

	queryerStub := func(err error) *QueryerMock {
		return &QueryerMock{
			QueryFunc: func(_ string, _ ...interface{}) (*sql.Rows, error) {
				if err != nil {
					return nil, err
				}
				return &sql.Rows{}, nil
			},
		}
	}

	arg := "test"

	// query not found
	q := queryerStub(nil)
	rows, err := dot.Query(q, "insert", arg)
	if rows != nil {
		t.Error("rows expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff := q.QueryCalls()
	if len(ff) > 0 {
		t.Error("query was not expected to be called")
	}

	// error returned by db
	q = queryerStub(errors.New("critical error"))
	rows, err = dot.Query(q, "select", arg)
	if rows != nil {
		t.Error("rows expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff = q.QueryCalls()
	tmpl := extractTemplate(t, dot, "select")
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	}

	// error returned by db (with options)
	q = queryerStub(errors.New("critical error"))
	rows, err = dotExcludeDeleted.Query(q, "select", arg)
	if rows != nil {
		t.Error("rows expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff = q.QueryCalls()
	tmpl = extractTemplate(t, dotExcludeDeleted, "select")
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	}

	// successful query
	q = queryerStub(nil)
	rows, err = dot.Query(q, "select", arg)
	if rows == nil {
		t.Error("rows expected to be non-nil, got nil")
	}

	failIfError(t, err)

	ff = q.QueryCalls()
	tmpl = extractTemplate(t, dot, "select")
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	}

	// successful query (with options)
	q = queryerStub(nil)
	rows, err = dotExcludeDeleted.Query(q, "select", arg)
	if rows == nil {
		t.Error("rows expected to be non-nil, got nil")
	}

	failIfError(t, err)

	ff = q.QueryCalls()
	tmpl = extractTemplate(t, dotExcludeDeleted, "select")
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	}
}

func TestQueryRow(t *testing.T) {
	dot := DotSql{
		queries: map[string]*template.Template{
			"select": createTemplate(t, "SELECT * from users WHERE name = ?"),
		},
	}

	queryRowerStub := func(success bool) *QueryRowerMock {
		return &QueryRowerMock{
			QueryRowFunc: func(_ string, _ ...interface{}) *sql.Row {
				if !success {
					return nil
				}
				return &sql.Row{}
			},
		}
	}

	arg := "test"

	// query not found
	q := queryRowerStub(true)
	row, err := dot.QueryRow(q, "insert", arg)
	if row != nil {
		t.Error("row expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff := q.QueryRowCalls()
	if len(ff) > 0 {
		t.Error("query was not expected to be called")
	}

	err = nil
	// error returned by db
	q = queryRowerStub(false)
	row, err = dot.QueryRow(q, "select", arg)
	if row != nil {
		t.Error("row expected to be nil, got non-nil")
	}

	ff = q.QueryRowCalls()
	tmpl := extractTemplate(t, dot, "select")
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	}

	// successful query
	q = queryRowerStub(true)
	row, err = dot.QueryRow(q, "select", arg)
	if row == nil {
		t.Error("row expected to be non-nil, got nil")
	}

	ff = q.QueryRowCalls()
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	}
}

func TestQueryRowContext(t *testing.T) {
	dot := DotSql{
		queries: map[string]*template.Template{
			"select": createTemplate(t, "SELECT * from users WHERE name = ?"),
		},
	}

	queryRowerContextStub := func(success bool) *QueryRowerContextMock {
		return &QueryRowerContextMock{
			QueryRowContextFunc: func(_ context.Context, _ string, _ ...interface{}) *sql.Row {
				if !success {
					return nil
				}
				return &sql.Row{}
			},
		}
	}

	ctx := context.Background()
	arg := "test"

	// query not found
	q := queryRowerContextStub(true)
	row, err := dot.QueryRowContext(ctx, q, "insert", arg)
	if row != nil {
		t.Error("row expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff := q.QueryRowContextCalls()
	if len(ff) > 0 {
		t.Error("query was not expected to be called")
	}

	err = nil
	// error returned by db
	q = queryRowerContextStub(false)
	row, err = dot.QueryRowContext(ctx, q, "select", arg)
	if row != nil {
		t.Error("row expected to be nil, got non-nil")
	}

	ff = q.QueryRowContextCalls()
	tmpl := extractTemplate(t, dot, "select")
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	} else if !reflect.DeepEqual(ff[0].Ctx, ctx) {
		t.Errorf("query context does not match")
	}

	// successful query
	q = queryRowerContextStub(true)
	row, err = dot.QueryRowContext(ctx, q, "select", arg)
	if row == nil {
		t.Error("row expected to be non-nil, got nil")
	}

	ff = q.QueryRowContextCalls()
	if len(ff) != 1 {
		t.Errorf("query was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("query was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("query was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("query was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	} else if !reflect.DeepEqual(ff[0].Ctx, ctx) {
		t.Errorf("query context does not match")
	}
}

type sqlResult struct{}

func (s sqlResult) LastInsertId() (int64, error) {
	return 1, nil
}

func (s sqlResult) RowsAffected() (int64, error) {
	return 1, nil
}

func TestExec(t *testing.T) {
	dot := DotSql{
		queries: map[string]*template.Template{
			"select": createTemplate(t, "SELECT * from users WHERE name = ?"),
		},
	}

	execerStub := func(err error) *ExecerMock {
		return &ExecerMock{
			ExecFunc: func(_ string, _ ...interface{}) (sql.Result, error) {
				if err != nil {
					return nil, err
				}
				return sqlResult{}, nil
			},
		}
	}

	arg := "test"

	// query not found
	q := execerStub(nil)
	result, err := dot.Exec(q, "insert", arg)
	if result != nil {
		t.Error("result expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff := q.ExecCalls()
	if len(ff) > 0 {
		t.Error("exec was not expected to be called")
	}

	// error returned by db
	q = execerStub(errors.New("critical error"))
	result, err = dot.Exec(q, "select", arg)
	if result != nil {
		t.Error("result expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff = q.ExecCalls()
	tmpl := extractTemplate(t, dot, "select")
	if len(ff) != 1 {
		t.Errorf("exec was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("exec was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("exec was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("exec was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	}

	// successful query exec
	q = execerStub(nil)
	result, err = dot.Exec(q, "select", arg)
	if result == nil {
		t.Error("result expected to be non-nil, got nil")
	}

	failIfError(t, err)

	ff = q.ExecCalls()
	if len(ff) != 1 {
		t.Errorf("exec was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("exec was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("exec was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("exec was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	}
}

func TestExecContext(t *testing.T) {
	dot := DotSql{
		queries: map[string]*template.Template{
			"select": createTemplate(t, "SELECT * from users WHERE name = ?"),
		},
	}

	execerContextStub := func(err error) *ExecerContextMock {
		return &ExecerContextMock{
			ExecContextFunc: func(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
				if err != nil {
					return nil, err
				}
				return sqlResult{}, nil
			},
		}
	}

	ctx := context.Background()
	arg := "test"

	// query not found
	q := execerContextStub(nil)
	result, err := dot.ExecContext(ctx, q, "insert", arg)
	if result != nil {
		t.Error("result expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff := q.ExecContextCalls()
	if len(ff) > 0 {
		t.Error("exec was not expected to be called")
	}

	// error returned by db
	q = execerContextStub(errors.New("critical error"))
	result, err = dot.ExecContext(ctx, q, "select", arg)
	if result != nil {
		t.Error("result expected to be nil, got non-nil")
	}

	if err == nil {
		t.Error("err expected to be non-nil, got nil")
	}

	ff = q.ExecContextCalls()
	tmpl := extractTemplate(t, dot, "select")
	if len(ff) != 1 {
		t.Errorf("exec was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("exec was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("exec was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("exec was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	} else if !reflect.DeepEqual(ff[0].Ctx, ctx) {
		t.Errorf("exec context does not match")
	}

	// successful query exec
	q = execerContextStub(nil)
	result, err = dot.ExecContext(ctx, q, "select", arg)
	if result == nil {
		t.Error("result expected to be non-nil, got nil")
	}

	failIfError(t, err)

	ff = q.ExecContextCalls()
	if len(ff) != 1 {
		t.Errorf("exec was expected to be called only once, but was called %d times", len(ff))
	} else if ff[0].Query != tmpl {
		t.Errorf("exec was expected to be called with %q query, got %q", tmpl, ff[0].Query)
	} else if len(ff[0].Args) != 1 {
		t.Errorf("exec was expected to be called with 1 argument, got %d", len(ff[0].Args))
	} else if !reflect.DeepEqual(ff[0].Args[0], arg) {
		t.Errorf("exec was expected to be called with %q argument, got %v", arg, ff[0].Args[0])
	} else if !reflect.DeepEqual(ff[0].Ctx, ctx) {
		t.Errorf("exec context does not match")
	}
}

func TestLoad(t *testing.T) {
	_, err := Load(strings.NewReader(""))
	failIfError(t, err)
}

func TestLoadFromFile(t *testing.T) {
	dot, err := LoadFromFile("./non-existent.sql")
	if err == nil {
		t.Error("error expected to be non-nil, got nil")
	}

	if dot != nil {
		t.Error("dotsql instance expected to be nil, got non-nil")
	}

	dot, err = LoadFromFile("./test_schema.sql")
	failIfError(t, err)

	if dot == nil {
		t.Error("dotsql instance expected to be non-nil, got nil")
	}
}

func TestLoadFromString(t *testing.T) {
	_, err := LoadFromString("")
	failIfError(t, err)
}

func TestRaw(t *testing.T) {
	expectedQuery := "SELECT 1+1"

	dot, err := LoadFromString("--name: my-query\n" + expectedQuery)
	failIfError(t, err)

	got, err := dot.Raw("my-query")
	failIfError(t, err)

	got = strings.TrimSpace(got)
	if got != expectedQuery {
		t.Errorf("Raw() == '%s', expected '%s'", got, expectedQuery)
	}
}

func TestQueries(t *testing.T) {
	expectedQueryMap := map[string]string{
		"select": "SELECT * from users",
		"insert": "INSERT INTO users (?, ?, ?)",
	}

	dot, err := LoadFromString(`
	-- name: select
	SELECT * from users

	-- name: insert
	INSERT INTO users (?, ?, ?)
	`)
	failIfError(t, err)

	got := dot.QueryMap()

	if len(got) != len(expectedQueryMap) {
		t.Errorf("QueryMap() len (%d) differ from expected (%d)", len(got), len(expectedQueryMap))
	}

	for name, query := range got {
		tmpl := executeTemplate(t, query)
		if tmpl != expectedQueryMap[name] {
			t.Errorf("QueryMap()[%s] == '%s', expected '%s'", name, tmpl, expectedQueryMap[name])
		}
	}
}

func TestMergeHaveBothQueries(t *testing.T) {
	expectedQueryMap := map[string]string{
		"query-a": "SELECT * FROM a",
		"query-b": "SELECT * FROM b",
	}

	a, err := LoadFromString("--name: query-a\nSELECT * FROM a")
	failIfError(t, err)

	b, err := LoadFromString("--name: query-b\nSELECT * FROM b")
	failIfError(t, err)

	c := Merge(a, b)

	got := c.QueryMap()
	if len(got) != len(expectedQueryMap) {
		t.Errorf("QueryMap() len (%d) differ from expected (%d)", len(got), len(expectedQueryMap))
	}
}

func TestMergeTakesPresecendeFromLastArgument(t *testing.T) {
	expectedQuery := "SELECT * FROM c"

	a, err := LoadFromString("--name: query\nSELECT * FROM a")
	failIfError(t, err)

	b, err := LoadFromString("--name: query\nSELECT * FROM b")
	failIfError(t, err)

	c, err := LoadFromString("--name: query\nSELECT * FROM c")
	failIfError(t, err)

	x := Merge(a, b, c)

	got := executeTemplate(t, x.QueryMap()["query"])
	if expectedQuery != got {
		t.Errorf("Expected query: '%s', got: '%s'", expectedQuery, got)
	}
}
