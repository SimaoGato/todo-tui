package db

import (
	"strings"
	"testing"
	"time"

	"github.com/justasandbox/my-todo-cli/internal/todo"
)

// TestCreate_DueDatePreservesLocalDate verifies that Create stores the local calendar
// date, not the UTC date. In UTC+ zones, local midnight is UTC previous-day evening;
// dueDate.UTC().Format("2006-01-02") would give the wrong UTC day.
func TestCreate_DueDatePreservesLocalDate(t *testing.T) {
	repo := openTestDB(t)

	loc := time.FixedZone("UTC+5", 5*60*60)
	// Midnight June 8 in UTC+5 = June 7 19:00 UTC.
	// Buggy .UTC().Format() stores "2026-06-07"; fixed .Format() stores "2026-06-08".
	due := time.Date(2026, 6, 8, 0, 0, 0, 0, loc)

	task, err := repo.Create("date test", &due)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var stored string
	if err := repo.db.QueryRow(`SELECT due_date FROM todos WHERE id = ?`, task.ID).Scan(&stored); err != nil {
		t.Fatalf("QueryRow: %v", err)
	}

	// Driver stores "YYYY-MM-DDT00:00:00Z". We check the date prefix is "2026-06-08"
	// (local date), not "2026-06-07" (UTC date of the same instant).
	if !strings.HasPrefix(stored, "2026-06-08") {
		t.Errorf("stored due_date = %q, want prefix %q (local date, not UTC date)", stored, "2026-06-08")
	}
}

// TestList_DueDateCalendarDayRoundtrip verifies that a due date created in a UTC+ timezone
// round-trips through Create → List with the correct local calendar day.
// Without the Create fix, UTC().Format() shifts the stored day back one; List() then
// returns a DueDate whose Date() is the wrong calendar day.
func TestList_DueDateCalendarDayRoundtrip(t *testing.T) {
	repo := openTestDB(t)

	loc := time.FixedZone("UTC+5", 5*60*60)
	// Midnight June 8 UTC+5 = June 7 19:00 UTC → buggy code stores June 7, correct stores June 8.
	due := time.Date(2026, 6, 8, 0, 0, 0, 0, loc)

	_, err := repo.Create("roundtrip test", &due)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	todos, err := repo.List(todo.FilterAll)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}

	task := todos[0]
	if task.DueDate == nil {
		t.Fatal("expected non-nil DueDate")
	}

	// DueDate is stored/parsed as UTC midnight; its calendar Date() should be June 8.
	gotY, gotM, gotD := task.DueDate.Date()
	if gotY != 2026 || gotM != time.June || gotD != 8 {
		t.Errorf("DueDate.Date() = %d-%02d-%02d, want 2026-06-08 (local calendar day)",
			gotY, int(gotM), gotD)
	}
}
