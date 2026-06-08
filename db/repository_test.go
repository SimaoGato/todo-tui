package db

import (
	"testing"
	"time"
)

func openTestDB(t *testing.T) *Repository {
	t.Helper()
	conn, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return NewRepository(conn)
}

func TestCreate_NoDate(t *testing.T) {
	repo := openTestDB(t)

	todo, err := repo.Create("buy milk", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if todo.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if todo.Title != "buy milk" {
		t.Errorf("title: got %q, want %q", todo.Title, "buy milk")
	}
	if todo.Done {
		t.Error("expected done=false")
	}
	if todo.DueDate != nil {
		t.Errorf("expected nil DueDate, got %v", todo.DueDate)
	}
	if todo.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if todo.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestCreate_WithDate(t *testing.T) {
	repo := openTestDB(t)

	due := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	todo, err := repo.Create("new year task", &due)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if todo.DueDate == nil {
		t.Fatal("expected non-nil DueDate")
	}
	if !todo.DueDate.Equal(due) {
		t.Errorf("DueDate: got %v, want %v", todo.DueDate, due)
	}
}

func TestCreate_QueryBack(t *testing.T) {
	repo := openTestDB(t)

	created, err := repo.Create("verify roundtrip", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	row := repo.db.QueryRow(
		`SELECT id, title, done, created_at FROM todos WHERE id = ?`, created.ID,
	)
	var id int
	var title string
	var done bool
	var createdAt string
	if err := row.Scan(&id, &title, &done, &createdAt); err != nil {
		t.Fatalf("QueryRow: %v", err)
	}

	if id != created.ID {
		t.Errorf("id: got %d, want %d", id, created.ID)
	}
	if title != "verify roundtrip" {
		t.Errorf("title: got %q", title)
	}
	if done {
		t.Error("expected done=false in DB")
	}
}
