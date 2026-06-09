package ui

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/justasandbox/my-todo-cli/internal/todo"
)

// AC1 + AC7 — ToggleDone is not called synchronously; cmd dispatches it.

func TestAsyncToggle_SpaceDoesNotCallRepoSynchronously(t *testing.T) {
	called := false
	todos := []todo.Task{{ID: 1, Title: "task"}}
	repo := &testRepo{OnToggleDone: func(_ int) error { called = true; return nil }}
	m := New(repo)
	m.Tasks = todos

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	if called {
		t.Error("ToggleDone must not be called synchronously")
	}
	if cmd == nil {
		t.Fatal("space should return a non-nil cmd")
	}
	cmd()
	if !called {
		t.Error("ToggleDone should be called when cmd is executed")
	}
}

func TestAsyncToggle_SuccessMsgIncrementsCompletedAndReloads(t *testing.T) {
	m := New(&testRepo{OnList: func(_ todo.Filter) ([]todo.Task, error) { return nil, nil }})
	m.Tasks = []todo.Task{{ID: 1}}

	next, cmd := m.Update(toggleDoneMsg{wasIncomplete: true, err: nil})
	m = next.(AppModel)
	if m.SessCompleted != 1 {
		t.Errorf("SessCompleted: got %d, want 1", m.SessCompleted)
	}
	if cmd == nil {
		t.Error("success should return loadTodos cmd")
	}
}

func TestAsyncToggle_SuccessMsgNoIncrementWhenAlreadyDone(t *testing.T) {
	m := New(&testRepo{OnList: func(_ todo.Filter) ([]todo.Task, error) { return nil, nil }})

	next, _ := m.Update(toggleDoneMsg{wasIncomplete: false, err: nil})
	m = next.(AppModel)
	if m.SessCompleted != 0 {
		t.Errorf("SessCompleted should not increment for undo, got %d", m.SessCompleted)
	}
}

// AC3 + AC6 — error path for ToggleDone sets errorMsg, no loadTodos.

func TestAsyncToggle_ErrorMsgSetsErrorMsg(t *testing.T) {
	m := New(&testRepo{})
	m.Tasks = []todo.Task{{ID: 1}}

	next, cmd := m.Update(toggleDoneMsg{err: errors.New("disk full")})
	m = next.(AppModel)
	if m.errorMsg == "" {
		t.Error("errorMsg should be set on toggle failure")
	}
	if cmd != nil {
		t.Error("failed toggle should not trigger a reload")
	}
}

// AC1 + AC7 — Delete is not called synchronously; cmd dispatches it.

func TestAsyncDelete_YDoesNotCallRepoSynchronously(t *testing.T) {
	called := false
	todos := []todo.Task{{ID: 1, Title: "task"}}
	repo := &testRepo{OnDelete: func(_ int) error { called = true; return nil }}
	m := New(repo)
	m.Tasks = todos
	m.ConfirmDelete = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if called {
		t.Error("Delete must not be called synchronously")
	}
	if cmd == nil {
		t.Fatal("y should return a non-nil cmd")
	}
	cmd()
	if !called {
		t.Error("Delete should be called when cmd is executed")
	}
}

func TestAsyncDelete_SuccessMsgIncrementsDeletedAndReloads(t *testing.T) {
	m := New(&testRepo{OnList: func(_ todo.Filter) ([]todo.Task, error) { return nil, nil }})

	next, cmd := m.Update(deleteDoneMsg{err: nil})
	m = next.(AppModel)
	if m.SessDeleted != 1 {
		t.Errorf("SessDeleted: got %d, want 1", m.SessDeleted)
	}
	if cmd == nil {
		t.Error("success should return loadTodos cmd")
	}
}

// AC3 + AC6 — error path for Delete sets errorMsg, no loadTodos.

func TestAsyncDelete_ErrorMsgSetsErrorMsg(t *testing.T) {
	m := New(&testRepo{})

	next, cmd := m.Update(deleteDoneMsg{err: errors.New("locked")})
	m = next.(AppModel)
	if m.errorMsg == "" {
		t.Error("errorMsg should be set on delete failure")
	}
	if cmd != nil {
		t.Error("failed delete should not trigger a reload")
	}
}

// AC1 + AC7 — Create is not called synchronously; cmd dispatches it.

func TestAsyncCreate_EnterOnDateStepDoesNotCallRepoSynchronously(t *testing.T) {
	called := false
	repo := &testRepo{
		OnCreate: func(_ string, _ *time.Time) (todo.Task, error) {
			called = true
			return todo.Task{}, nil
		},
	}
	m := New(repo)
	m.InputMode = true
	m.inputStep = stepDate
	m.pendingTitle = "My Task"

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if called {
		t.Error("Create must not be called synchronously")
	}
	if cmd == nil {
		t.Fatal("Enter on stepDate should return a non-nil cmd")
	}
	cmd()
	if !called {
		t.Error("Create should be called when cmd is executed")
	}
}

// AC7 — success path: exits input mode and triggers reload.

func TestAsyncCreate_SuccessMsgExitsInputMode(t *testing.T) {
	m := New(&testRepo{OnList: func(_ todo.Filter) ([]todo.Task, error) { return nil, nil }})
	m.InputMode = true
	m.inputStep = stepDate
	m.pendingTitle = "Task"

	next, cmd := m.Update(createDoneMsg{err: nil})
	m = next.(AppModel)
	if m.InputMode {
		t.Error("InputMode should be false after successful create")
	}
	if m.pendingTitle != "" {
		t.Errorf("pendingTitle should be cleared, got %q", m.pendingTitle)
	}
	if cmd == nil {
		t.Error("success should return loadTodos cmd")
	}
}

// AC4 + AC6 — failed Create keeps input mode with text preserved; sets inputErr.

func TestAsyncCreate_ErrorMsgKeepsInputModeWithTextPreserved(t *testing.T) {
	m := New(&testRepo{})
	m.InputMode = true
	m.inputStep = stepDate
	m.pendingTitle = "My Task"

	next, cmd := m.Update(createDoneMsg{err: errors.New("db error")})
	m = next.(AppModel)
	if !m.InputMode {
		t.Error("InputMode should remain true on create failure")
	}
	if m.pendingTitle != "My Task" {
		t.Errorf("pendingTitle should be preserved, got %q", m.pendingTitle)
	}
	if m.inputErr == "" {
		t.Error("inputErr should be set on create failure")
	}
	if cmd != nil {
		t.Error("failed create should not trigger a reload")
	}
}
