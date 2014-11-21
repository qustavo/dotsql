package dotsql

import (
	"strings"

	"testing"
)

func failIfError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("err == nil, got '%s'", err)
	}
}

func TestLoad(t *testing.T) {
	_, err := Load(strings.NewReader(""))
	failIfError(t, err)
}

func TestLoadFromString(t *testing.T) {
	_, err := LoadFromString("")
	failIfError(t, err)
}
