package dotsql

import (
	"context"
	"reflect"
	"testing"
)

type PrepareCalls []struct {
	Query string
}

func comparePrepareCalls(t *testing.T, ff PrepareCalls, template string) bool {
	t.Helper()
	if len(ff) != 1 {
		t.Errorf("prepare was expected to be called only once, but was called %d times", len(ff))
		return false
	} else if ff[0].Query != template {
		t.Errorf("prepare was expected to be called with %q query, got %q", template, ff[0].Query)
		return false
	}
	return true
}

type PrepareContextCalls []struct {
	Ctx   context.Context
	Query string
}

func comparePrepareContextCalls(t *testing.T, ff PrepareContextCalls, ctx context.Context, template string) bool {
	t.Helper()
	if len(ff) != 1 {
		t.Errorf("prepare was expected to be called only once, but was called %d times", len(ff))
		return false
	} else if ff[0].Query != template {
		t.Errorf("prepare was expected to be called with %q query, got %q", template, ff[0].Query)
		return false
	} else if !reflect.DeepEqual(ff[0].Ctx, ctx) {
		t.Error("prepare context does not match")
		return false
	}
	return true
}

type QueryCalls []struct {
	Query string
	Args  []interface{}
}

func compareCalls(t *testing.T, ff QueryCalls, command, template, testArg string) bool {
	t.Helper()
	if len(ff) != 1 {
		t.Errorf("%s was expected to be called only once, but was called %d times", command, len(ff))
		return false
	} else if ff[0].Query != template {
		t.Errorf("%s was expected to be called with %q query, got %q", command, template, ff[0].Query)
		return false
	} else if len(ff[0].Args) != 1 {
		t.Errorf("%s was expected to be called with 1 argument, got %d", command, len(ff[0].Args))
		return false
	} else if !reflect.DeepEqual(ff[0].Args[0], testArg) {
		t.Errorf("%s was expected to be called with %q argument, got %v", command, testArg, ff[0].Args[0])
		return false
	}
	return true
}

type QueryContextCalls []struct {
	Ctx   context.Context
	Query string
	Args  []interface{}
}

func compareContextCalls(t *testing.T, ff QueryContextCalls, ctx context.Context, command, template, testArg string) bool {
	t.Helper()
	if len(ff) != 1 {
		t.Errorf("%s was expected to be called only once, but was called %d times", command, len(ff))
		return false
	} else if ff[0].Query != template {
		t.Errorf("%s was expected to be called with %q query, got %q", command, template, ff[0].Query)
		return false
	} else if len(ff[0].Args) != 1 {
		t.Errorf("%s was expected to be called with 1 argument, got %d", command, len(ff[0].Args))
		return false
	} else if !reflect.DeepEqual(ff[0].Args[0], testArg) {
		t.Errorf("%s was expected to be called with %q argument, got %v", command, testArg, ff[0].Args[0])
		return false
	} else if !reflect.DeepEqual(ff[0].Ctx, ctx) {
		t.Errorf("%s context does not match", command)
		return false
	}
	return true
}
