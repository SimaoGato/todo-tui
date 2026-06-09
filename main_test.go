package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDBPath_DefaultEndsWithTodoDB(t *testing.T) {
	os.Unsetenv("TODO_DB_PATH")
	p := dbPath()
	if !strings.HasSuffix(p, ".todo.db") {
		t.Errorf("default path should end with .todo.db, got %s", p)
	}
}

func TestDBPath_AbsoluteEnvVar(t *testing.T) {
	os.Setenv("TODO_DB_PATH", "/tmp/custom.db")
	defer os.Unsetenv("TODO_DB_PATH")
	p := dbPath()
	if p != "/tmp/custom.db" {
		t.Errorf("expected /tmp/custom.db, got %s", p)
	}
}

func TestDBPath_RelativePathIsExpanded(t *testing.T) {
	os.Setenv("TODO_DB_PATH", "relative/test.db")
	defer os.Unsetenv("TODO_DB_PATH")
	p := dbPath()
	if !filepath.IsAbs(p) {
		t.Errorf("relative path should be expanded to absolute, got %s", p)
	}
}

func TestDBPath_EnvVarOverridesDefault(t *testing.T) {
	os.Setenv("TODO_DB_PATH", "/custom/path.db")
	defer os.Unsetenv("TODO_DB_PATH")
	p := dbPath()
	if strings.HasSuffix(p, ".todo.db") {
		t.Errorf("env var should override default, but got default path: %s", p)
	}
}
