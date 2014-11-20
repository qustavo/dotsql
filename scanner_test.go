package dotsql

import (
	"bufio"
	"strings"
	"testing"
)

func TestGetTag(t *testing.T) {
	var tests = []struct {
		line string
		want string
	}{
		{"SELECT 1+1", ""},
		{"-- Some Comment", ""},
		{"-- name:  ", ""},
		{"-- name: find-users-by-name", "find-users-by-name"},
		{"  --  name:  save-user ", "save-user"},
	}

	for _, c := range tests {
		got := getTag(c.line)
		if got != c.want {
			t.Errorf("isTag('%s') == %s, expect %v", c.line, got, c.want)
		}
	}
}

func TestScannerRun(t *testing.T) {
	sqlFile := `
	-- name: all-users
	-- Finds all users
	SELECT * from USER

	-- name: empty-query-should-not-be-stored
	-- name: save-user
	INSERT INTO users (?, ?, ?)
	`

	scanner := &Scanner{}
	queries := scanner.Run(bufio.NewScanner(strings.NewReader(sqlFile)))

	numberOfQueries := len(queries)
	expectedQueries := 2
	if numberOfQueries != expectedQueries {
		t.Errorf("Scanner/Run() has %d queries instead of %d",
			numberOfQueries, expectedQueries)
	}
}
