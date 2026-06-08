package db

import (
	"os"
	"testing"
)

func TestOpen_CreatesDBAndTable(t *testing.T) {
	path := t.TempDir() + "/test.db"

	conn, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer conn.Close()

	// Verify the file exists on disk.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}

	// Verify the todos table has the expected columns.
	rows, err := conn.Query("PRAGMA table_info(todos)")
	if err != nil {
		t.Fatalf("PRAGMA table_info error: %v", err)
	}
	defer rows.Close()

	want := map[string]bool{
		"id": false, "title": false, "done": false,
		"due_date": false, "created_at": false, "updated_at": false,
	}
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("scan error: %v", err)
		}
		want[name] = true
	}
	for col, found := range want {
		if !found {
			t.Errorf("missing column: %s", col)
		}
	}
}

func TestOpen_IdempotentOnSecondCall(t *testing.T) {
	path := t.TempDir() + "/test.db"

	conn1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open() error: %v", err)
	}
	conn1.Close()

	conn2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open() error: %v", err)
	}
	conn2.Close()
}
