package dotsql

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"text/template"
)

func failIfError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("err == nil, got '%s'", err)
	}
}

func failIfNotError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("err expected to be non-nil, got nil")
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

	tmpl := extractTemplate(t, dot, "select")

	t.Run("query not found", func(t *testing.T) {
		p := preparerStub(nil)
		stmt, err := dot.Prepare(p, "insert")
		if stmt != nil {
			t.Error("stmt expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		ff := p.PrepareCalls()
		if len(ff) > 0 {
			t.Error("prepare was not expected to be called")
		}
	})

	t.Run("error returned by DB", func(t *testing.T) {
		p := preparerStub(errors.New("critical error"))
		stmt, err := dot.Prepare(p, "select")
		if stmt != nil {
			t.Error("stmt expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		comparePrepareCalls(t, p.PrepareCalls(), tmpl)
	})

	t.Run("successful query preparation", func(t *testing.T) {
		p := preparerStub(nil)
		stmt, err := dot.Prepare(p, "select")
		if stmt == nil {
			t.Error("stmt expected to be non-nil, got nil")
		}

		failIfError(t, err)
		comparePrepareCalls(t, p.PrepareCalls(), tmpl)
	})
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
	tmpl := extractTemplate(t, dot, "select")

	t.Run("query not found", func(t *testing.T) {
		p := preparerContextStub(nil)
		stmt, err := dot.PrepareContext(ctx, p, "insert")
		if stmt != nil {
			t.Error("stmt expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		ff := p.PrepareContextCalls()
		if len(ff) > 0 {
			t.Error("prepare was not expected to be called")
		}
	})

	t.Run("error returned by DB", func(t *testing.T) {
		p := preparerContextStub(errors.New("critical error"))
		stmt, err := dot.PrepareContext(ctx, p, "select")
		if stmt != nil {
			t.Error("stmt expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		comparePrepareContextCalls(t, p.PrepareContextCalls(), ctx, tmpl)
	})

	t.Run("successful query preparation", func(t *testing.T) {
		p := preparerContextStub(nil)
		stmt, err := dot.PrepareContext(ctx, p, "select")
		if stmt == nil {
			t.Error("stmt expected to be non-nil, got nil")
		}

		failIfError(t, err)
		comparePrepareContextCalls(t, p.PrepareContextCalls(), ctx, tmpl)
	})
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

	testArg := "test"
	tmpl := extractTemplate(t, dot, "select")

	t.Run("query not found", func(t *testing.T) {
		q := queryerStub(nil)
		rows, err := dot.Query(q, "insert", testArg)
		if rows != nil {
			t.Error("rows expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		ff := q.QueryCalls()
		if len(ff) > 0 {
			t.Error("query was not expected to be called")
		}
	})

	t.Run("error returned by DB", func(t *testing.T) {
		q := queryerStub(errors.New("critical error"))
		rows, err := dot.Query(q, "select", testArg)
		if rows != nil {
			t.Error("rows expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		compareCalls(t, q.QueryCalls(), "query", tmpl, testArg)
	})

	t.Run("successful query", func(t *testing.T) {
		q := queryerStub(nil)
		rows, err := dot.Query(q, "select", testArg)
		if rows == nil {
			t.Error("rows expected to be non-nil, got nil")
		}

		failIfError(t, err)
		compareCalls(t, q.QueryCalls(), "query", tmpl, testArg)
	})
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
	testArg := "test"
	tmpl := extractTemplate(t, dot, "select")

	t.Run("query not found", func(t *testing.T) {
		q := queryerContextStub(nil)
		rows, err := dot.QueryContext(ctx, q, "insert", testArg)
		if rows != nil {
			t.Error("rows expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		ff := q.QueryContextCalls()
		if len(ff) > 0 {
			t.Error("query was not expected to be called")
		}
	})

	t.Run("error returned by DB", func(t *testing.T) {
		q := queryerContextStub(errors.New("critical error"))
		rows, err := dot.QueryContext(ctx, q, "select", testArg)
		if rows != nil {
			t.Error("rows expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		compareContextCalls(t, q.QueryContextCalls(), ctx, "query", tmpl, testArg)
	})

	t.Run("successful query", func(t *testing.T) {
		q := queryerContextStub(nil)
		rows, err := dot.QueryContext(ctx, q, "select", testArg)
		if rows == nil {
			t.Error("rows expected to be non-nil, got nil")
		}

		failIfError(t, err)
		compareContextCalls(t, q.QueryContextCalls(), ctx, "query", tmpl, testArg)
	})
}

func TestQueryWithData(t *testing.T) {
	dot := DotSql{
		queries: map[string]*template.Template{
			"select": createTemplate(t, "SELECT * from users WHERE name = ?{{if  .exclude_deleted}} AND deleted IS NULL{{end}}"),
		},
	}
	dotExcludeDeleted := dot.WithData(map[string]any{"exclude_deleted": true})

	queryerStub := func(err error) *QueryerMock {
		return &QueryerMock{
			QueryFunc: func(_ string, _ ...any) (*sql.Rows, error) {
				if err != nil {
					return nil, err
				}
				return &sql.Rows{}, nil
			},
		}
	}

	testArg := "test"
	criticalError := errors.New("critical error")

	t.Run("query not found", func(t *testing.T) {
		q := queryerStub(nil)
		rows, err := dot.Query(q, "insert", testArg)
		if rows != nil {
			t.Error("rows expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		ff := q.QueryCalls()
		if len(ff) > 0 {
			t.Error("query was not expected to be called")
		}
	})

	t.Run("error returned by DB", func(t *testing.T) {
		q := queryerStub(criticalError)
		rows, err := dot.Query(q, "select", testArg)
		if rows != nil {
			t.Error("rows expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		compareCalls(t, q.QueryCalls(), "query", extractTemplate(t, dot, "select"), testArg)
	})

	t.Run("error returned by DB (with data)", func(t *testing.T) {
		q := queryerStub(criticalError)
		rows, err := dotExcludeDeleted.Query(q, "select", testArg)
		if rows != nil {
			t.Error("rows expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		compareCalls(t, q.QueryCalls(), "query", extractTemplate(t, dotExcludeDeleted, "select"), testArg)
	})

	t.Run("successful query", func(t *testing.T) {
		q := queryerStub(nil)
		rows, err := dot.Query(q, "select", testArg)
		if rows == nil {
			t.Error("rows expected to be non-nil, got nil")
		}

		failIfError(t, err)
		compareCalls(t, q.QueryCalls(), "query", extractTemplate(t, dot, "select"), testArg)
	})

	t.Run("successful query (with data)", func(t *testing.T) {
		q := queryerStub(nil)
		rows, err := dotExcludeDeleted.Query(q, "select", testArg)
		if rows == nil {
			t.Error("rows expected to be non-nil, got nil")
		}

		failIfError(t, err)
		compareCalls(t, q.QueryCalls(), "query", extractTemplate(t, dotExcludeDeleted, "select"), testArg)
	})
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

	testArg := "test"

	t.Run("query not found", func(t *testing.T) {
		q := queryRowerStub(true)
		row, err := dot.QueryRow(q, "insert", testArg)
		if row != nil {
			t.Error("row expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		ff := q.QueryRowCalls()
		if len(ff) > 0 {
			t.Error("query was not expected to be called")
		}
	})

	t.Run("error returned by DB", func(t *testing.T) {
		q := queryRowerStub(false)
		row, _ := dot.QueryRow(q, "select", testArg)
		if row != nil {
			t.Error("row expected to be nil, got non-nil")
		}

		compareCalls(t, q.QueryRowCalls(), "query", extractTemplate(t, dot, "select"), testArg)
	})

	t.Run("successful query", func(t *testing.T) {
		q := queryRowerStub(true)
		row, err := dot.QueryRow(q, "select", testArg)
		if row == nil {
			t.Error("row expected to be non-nil, got nil")
		}

		failIfError(t, err)
		compareCalls(t, q.QueryRowCalls(), "query", extractTemplate(t, dot, "select"), testArg)
	})
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
	testArg := "test"
	template := extractTemplate(t, dot, "select")

	t.Run("query not found", func(t *testing.T) {
		q := queryRowerContextStub(true)
		row, err := dot.QueryRowContext(ctx, q, "insert", testArg)
		if row != nil {
			t.Error("row expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		ff := q.QueryRowContextCalls()
		if len(ff) > 0 {
			t.Error("query was not expected to be called")
		}
	})

	t.Run("error returned by DB", func(t *testing.T) {
		q := queryRowerContextStub(false)
		row, _ := dot.QueryRowContext(ctx, q, "select", testArg)
		if row != nil {
			t.Error("row expected to be nil, got non-nil")
		}

		compareContextCalls(t, q.QueryRowContextCalls(), ctx, "query", template, testArg)
	})

	t.Run("successful query", func(t *testing.T) {
		q := queryRowerContextStub(true)
		row, _ := dot.QueryRowContext(ctx, q, "select", testArg)
		if row == nil {
			t.Error("row expected to be non-nil, got nil")
		}

		compareContextCalls(t, q.QueryRowContextCalls(), ctx, "query", template, testArg)
	})
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

	testArg := "test"
	template := extractTemplate(t, dot, "select")

	t.Run("query not found", func(t *testing.T) {
		q := execerStub(nil)
		result, err := dot.Exec(q, "insert", testArg)
		if result != nil {
			t.Error("result expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		ff := q.ExecCalls()
		if len(ff) > 0 {
			t.Error("exec was not expected to be called")
		}
	})

	t.Run("error returned by DB", func(t *testing.T) {
		q := execerStub(errors.New("critical error"))
		result, err := dot.Exec(q, "select", testArg)
		if result != nil {
			t.Error("result expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		compareCalls(t, q.ExecCalls(), "exec", template, testArg)
	})

	t.Run("successful query exec", func(t *testing.T) {
		q := execerStub(nil)
		result, err := dot.Exec(q, "select", testArg)
		if result == nil {
			t.Error("result expected to be non-nil, got nil")
		}

		failIfError(t, err)
		compareCalls(t, q.ExecCalls(), "exec", template, testArg)
	})
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
	testArg := "test"
	tmpl := extractTemplate(t, dot, "select")

	t.Run("query not found", func(t *testing.T) {
		q := execerContextStub(nil)
		result, err := dot.ExecContext(ctx, q, "insert", testArg)
		if result != nil {
			t.Error("result expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		ff := q.ExecContextCalls()
		if len(ff) > 0 {
			t.Error("exec was not expected to be called")
		}
	})

	t.Run("error returned by DB", func(t *testing.T) {
		q := execerContextStub(errors.New("critical error"))
		result, err := dot.ExecContext(ctx, q, "select", testArg)
		if result != nil {
			t.Error("result expected to be nil, got non-nil")
		}

		failIfNotError(t, err)
		compareContextCalls(t, q.ExecContextCalls(), ctx, "exec", tmpl, testArg)
	})

	t.Run("successful query exec", func(t *testing.T) {
		q := execerContextStub(nil)
		result, err := dot.ExecContext(ctx, q, "select", testArg)
		if result == nil {
			t.Error("result expected to be non-nil, got nil")
		}

		failIfError(t, err)

		compareContextCalls(t, q.ExecContextCalls(), ctx, "exec", tmpl, testArg)
	})
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

func TestTemplatesFailQuickly(t *testing.T) {
	_, err := LoadFromString("-- name: bad-query\nSELECT * FROM {{if .user}}users{{else}}clients")
	expectedErr := "template: bad-query:1: unexpected EOF"
	if err == nil {
		t.Fatalf("expected '%s' error, but didn't get one", expectedErr)
	} else if err.Error() != expectedErr {
		t.Errorf("expected '%s' error, but got '%v'", expectedErr, err)
	}
}
