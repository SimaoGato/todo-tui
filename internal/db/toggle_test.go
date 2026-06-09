package db

import (
	"testing"
	"time"
)

func TestToggleDone_FalseToTrue(t *testing.T) {
	repo := openTestDB(t)
	todo, _ := repo.Create("toggle me", nil)

	if err := repo.ToggleDone(todo.ID); err != nil {
		t.Fatalf("ToggleDone: %v", err)
	}

	var done bool
	row := repo.db.QueryRow(`SELECT done FROM todos WHERE id = ?`, todo.ID)
	if err := row.Scan(&done); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if !done {
		t.Error("expected done=true after first toggle")
	}
}

func TestToggleDone_TrueToFalse(t *testing.T) {
	repo := openTestDB(t)
	todo, _ := repo.Create("toggle twice", nil)

	repo.ToggleDone(todo.ID)
	if err := repo.ToggleDone(todo.ID); err != nil {
		t.Fatalf("second ToggleDone: %v", err)
	}

	var done bool
	row := repo.db.QueryRow(`SELECT done FROM todos WHERE id = ?`, todo.ID)
	row.Scan(&done)
	if done {
		t.Error("expected done=false after second toggle")
	}
}

func TestToggleDone_UpdatesTimestamp(t *testing.T) {
	repo := openTestDB(t)
	todo, _ := repo.Create("timestamp check", nil)

	before := time.Now().UTC().Add(-time.Second)
	if err := repo.ToggleDone(todo.ID); err != nil {
		t.Fatalf("ToggleDone: %v", err)
	}

	var updatedAtStr string
	row := repo.db.QueryRow(`SELECT updated_at FROM todos WHERE id = ?`, todo.ID)
	row.Scan(&updatedAtStr)
	updatedAt, _ := time.Parse(time.RFC3339, updatedAtStr)
	if !updatedAt.After(before) {
		t.Errorf("expected updated_at to be refreshed, got %v", updatedAt)
	}
}

func TestToggleDone_NonexistentID(t *testing.T) {
	repo := openTestDB(t)
	if err := repo.ToggleDone(99999); err == nil {
		t.Error("expected error for nonexistent ID, got nil")
	}
}
