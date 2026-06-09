package db

import (
	"testing"
	"time"

	"github.com/justasandbox/my-todo-cli/todo"
)

// seed inserts a todo directly into the DB, optionally marking it done, with a
// controllable created_at so ordering tests are deterministic.
func seed(t *testing.T, repo *Repository, title string, dueDate *time.Time, done bool, createdAt time.Time) todo.Todo {
	t.Helper()
	task, err := repo.Create(title, dueDate)
	if err != nil {
		t.Fatalf("seed Create(%q): %v", title, err)
	}
	// Overwrite timestamps and done flag for ordering/filtering control.
	_, err = repo.db.Exec(
		`UPDATE todos SET done=?, created_at=?, updated_at=? WHERE id=?`,
		done, createdAt.UTC().Format(time.RFC3339), createdAt.UTC().Format(time.RFC3339), task.ID,
	)
	if err != nil {
		t.Fatalf("seed update(%q): %v", title, err)
	}
	task.Done = done
	task.CreatedAt = createdAt
	task.UpdatedAt = createdAt
	return task
}

func TestList_EmptyReturnsSlice(t *testing.T) {
	repo := openTestDB(t)
	for _, f := range []todo.Filter{todo.FilterAll, todo.FilterToday, todo.FilterDone} {
		todos, err := repo.List(f)
		if err != nil {
			t.Fatalf("List(%v): %v", f, err)
		}
		if todos == nil {
			t.Errorf("List(%v): expected non-nil empty slice, got nil", f)
		}
		if len(todos) != 0 {
			t.Errorf("List(%v): expected 0 items, got %d", f, len(todos))
		}
	}
}

func TestList_AllExcludesDone(t *testing.T) {
	repo := openTestDB(t)
	base := time.Now().UTC()

	seed(t, repo, "active", nil, false, base)
	seed(t, repo, "completed", nil, true, base.Add(time.Second))

	todos, err := repo.List(todo.FilterAll)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
	if todos[0].Title != "active" {
		t.Errorf("expected 'active', got %q", todos[0].Title)
	}
}

func TestList_AllOrderNullsLast(t *testing.T) {
	repo := openTestDB(t)
	base := time.Now().UTC()

	future := base.Add(48 * time.Hour)
	seed(t, repo, "no-date", nil, false, base)
	seed(t, repo, "has-date", &future, false, base.Add(time.Second))

	todos, err := repo.List(todo.FilterAll)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}
	if todos[0].Title != "has-date" {
		t.Errorf("expected 'has-date' first (nulls last), got %q", todos[0].Title)
	}
	if todos[1].Title != "no-date" {
		t.Errorf("expected 'no-date' second, got %q", todos[1].Title)
	}
}

func TestList_TodayFilter(t *testing.T) {
	repo := openTestDB(t)
	base := time.Now().UTC()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	seed(t, repo, "today-task", &today, false, base)
	seed(t, repo, "tomorrow-task", &tomorrow, false, base.Add(time.Second))
	seed(t, repo, "no-date-task", nil, false, base.Add(2*time.Second))

	todos, err := repo.List(todo.FilterToday)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
	if todos[0].Title != "today-task" {
		t.Errorf("expected 'today-task', got %q", todos[0].Title)
	}
}

func TestList_DoneFilter(t *testing.T) {
	repo := openTestDB(t)
	base := time.Now().UTC()

	seed(t, repo, "done-first", nil, true, base.Add(2*time.Second))
	seed(t, repo, "done-second", nil, true, base.Add(time.Second))
	seed(t, repo, "active", nil, false, base)

	todos, err := repo.List(todo.FilterDone)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}
	// updated_at DESC — done-first (later) should come first.
	if todos[0].Title != "done-first" {
		t.Errorf("expected 'done-first' first, got %q", todos[0].Title)
	}
	if todos[1].Title != "done-second" {
		t.Errorf("expected 'done-second' second, got %q", todos[1].Title)
	}
}
