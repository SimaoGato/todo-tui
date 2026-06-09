package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/justasandbox/my-todo-cli/internal/todo"
)

// helper: run a confirm-delete sequence (d → y) and deliver the reload msg.
func doDeleteConfirmed(m AppModel) AppModel {
	m = sendKey(m, "d")
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m = next.(AppModel)
	if cmd != nil {
		next, _ = m.Update(cmd())
		m = next.(AppModel)
	}
	return m
}

// helper: press Space on the model and deliver any reload cmd.
func doToggle(m AppModel) AppModel {
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = next.(AppModel)
	if cmd != nil {
		next, _ = m.Update(cmd())
		m = next.(AppModel)
	}
	return m
}

// AC2: SessCompleted increments when toggling a task from done=false to done=true.
func TestSessionSummary_CompletedIncrements(t *testing.T) {
	tasks := []todo.Task{{ID: 1, Title: "task", Done: false}}
	repo := &testRepo{
		OnList:       func(_ todo.Filter) ([]todo.Task, error) { return tasks, nil },
		OnToggleDone: func(_ int) error { return nil },
	}
	m := New(repo)
	m.Tasks = tasks

	m = doToggle(m)

	if m.SessCompleted != 1 {
		t.Errorf("SessCompleted after toggling done=false: got %d, want 1", m.SessCompleted)
	}
}

// AC2 (edge): SessCompleted must NOT increment when toggling a task from done=true back to false.
func TestSessionSummary_CompletedNotIncrementedForUndo(t *testing.T) {
	tasks := []todo.Task{{ID: 1, Title: "task", Done: true}}
	repo := &testRepo{
		OnList:       func(_ todo.Filter) ([]todo.Task, error) { return tasks, nil },
		OnToggleDone: func(_ int) error { return nil },
	}
	m := New(repo)
	m.Tasks = tasks

	m = doToggle(m)

	if m.SessCompleted != 0 {
		t.Errorf("SessCompleted after toggling done=true: got %d, want 0", m.SessCompleted)
	}
}

// AC3: SessDeleted increments after a confirmed delete.
func TestSessionSummary_DeletedIncrements(t *testing.T) {
	tasks := []todo.Task{{ID: 1, Title: "task"}}
	repo := &testRepo{
		OnList:   func(_ todo.Filter) ([]todo.Task, error) { return tasks, nil },
		OnDelete: func(_ int) error { return nil },
	}
	m := New(repo)
	m.Tasks = tasks

	m = doDeleteConfirmed(m)

	if m.SessDeleted != 1 {
		t.Errorf("SessDeleted after confirmed delete: got %d, want 1", m.SessDeleted)
	}
}

// AC3 (edge): SessDeleted must NOT increment when delete is cancelled (n).
func TestSessionSummary_DeletedNotIncrementedOnCancel(t *testing.T) {
	tasks := []todo.Task{{ID: 1, Title: "task"}}
	repo := &testRepo{
		OnList:   func(_ todo.Filter) ([]todo.Task, error) { return tasks, nil },
		OnDelete: func(_ int) error { return nil },
	}
	m := New(repo)
	m.Tasks = tasks

	m = sendKey(m, "d") // show confirm prompt
	m = sendKey(m, "n") // cancel

	if m.SessDeleted != 0 {
		t.Errorf("SessDeleted after cancelled delete: got %d, want 0", m.SessDeleted)
	}
}

// Counters accumulate across multiple operations.
func TestSessionSummary_AccumulatesAcrossOperations(t *testing.T) {
	tasks := []todo.Task{
		{ID: 1, Title: "a", Done: false},
		{ID: 2, Title: "b", Done: false},
	}
	repo := &testRepo{
		OnList:       func(_ todo.Filter) ([]todo.Task, error) { return tasks, nil },
		OnToggleDone: func(_ int) error { return nil },
		OnDelete:     func(_ int) error { return nil },
	}
	m := New(repo)
	m.Tasks = tasks

	m = doToggle(m)           // complete task 1 → SessCompleted=1
	m.Tasks = tasks           // reload (simulate list refresh keeping tasks)
	m = doDeleteConfirmed(m)  // delete task 1 → SessDeleted=1

	if m.SessCompleted != 1 {
		t.Errorf("SessCompleted: got %d, want 1", m.SessCompleted)
	}
	if m.SessDeleted != 1 {
		t.Errorf("SessDeleted: got %d, want 1", m.SessDeleted)
	}
}

// GetSummary returns correct Completed and Deleted values.
func TestSessionSummary_GetSummaryReturnsCounters(t *testing.T) {
	m := New(&testRepo{})
	m.SessCompleted = 3
	m.SessDeleted = 2

	s := m.GetSummary()

	if s.Completed != 3 {
		t.Errorf("Summary.Completed: got %d, want 3", s.Completed)
	}
	if s.Deleted != 2 {
		t.Errorf("Summary.Deleted: got %d, want 2", s.Deleted)
	}
}
