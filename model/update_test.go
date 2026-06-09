package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/justasandbox/my-todo-cli/todo"
)

// 3.2 – navigation

func TestNav_JMovesDown(t *testing.T) {
	m := modelWithTasks(3)
	m = sendKey(m, "j")
	if m.Cursor != 1 {
		t.Errorf("cursor after j: got %d, want 1", m.Cursor)
	}
}

func TestNav_KMovesUp(t *testing.T) {
	m := modelWithTasks(3)
	m.Cursor = 2
	m = sendKey(m, "k")
	if m.Cursor != 1 {
		t.Errorf("cursor after k: got %d, want 1", m.Cursor)
	}
}

func TestNav_JWrapsAround(t *testing.T) {
	m := modelWithTasks(3)
	m.Cursor = 2
	m = sendKey(m, "j")
	if m.Cursor != 0 {
		t.Errorf("cursor wrap-down: got %d, want 0", m.Cursor)
	}
}

func TestNav_KWrapsAround(t *testing.T) {
	m := modelWithTasks(3)
	m.Cursor = 0
	m = sendKey(m, "k")
	if m.Cursor != 2 {
		t.Errorf("cursor wrap-up: got %d, want 2", m.Cursor)
	}
}

func TestNav_EmptyListNoOp(t *testing.T) {
	m := modelWithTasks(0)
	m = sendKey(m, "j")
	m = sendKey(m, "k")
	if m.Cursor != 0 {
		t.Errorf("cursor on empty list: got %d, want 0", m.Cursor)
	}
}

func TestNav_IgnoredInInputMode(t *testing.T) {
	m := modelWithTasks(3)
	m.InputMode = true
	m = sendKey(m, "j")
	if m.Cursor != 0 {
		t.Error("j should be ignored in input mode")
	}
}

// 3.3 – tab switching

func TestTab_NumberKeySwitchesTab(t *testing.T) {
	m := modelWithTasks(0)
	cases := []struct {
		key string
		tab Tab
	}{
		{"1", TabToday},
		{"2", TabAll},
		{"3", TabCompleted},
	}
	for _, c := range cases {
		m2 := sendKey(m, c.key)
		if m2.ActiveTab != c.tab {
			t.Errorf("key %q: activeTab got %d, want %d", c.key, m2.ActiveTab, c.tab)
		}
	}
}

func TestTab_TabKeyCycles(t *testing.T) {
	m := modelWithTasks(0) // starts at TabToday
	tabs := []Tab{TabAll, TabCompleted, TabToday}
	for _, want := range tabs {
		m = sendKey(m, "tab")
		if m.ActiveTab != want {
			t.Errorf("after tab key: got %d, want %d", m.ActiveTab, want)
		}
	}
}

func TestTab_SwitchResetsCursor(t *testing.T) {
	m := modelWithTasks(3)
	m.Cursor = 2
	m = sendKey(m, "2")
	if m.Cursor != 0 {
		t.Errorf("cursor after tab switch: got %d, want 0", m.Cursor)
	}
}

// 3.4 – toggle

func TestToggle_SpaceCallsToggleDone(t *testing.T) {
	todos := []todo.Todo{{ID: 42, Title: "toggle me"}}
	toggled := false
	repo := &testRepo{
		OnList:       func(_ todo.Filter) ([]todo.Todo, error) { return todos, nil },
		OnToggleDone: func(id int) error { toggled = (id == 42); return nil },
	}
	m := New(repo)
	m.Tasks = todos

	sendKey(m, " ")
	if !toggled {
		t.Error("ToggleDone not called with correct ID")
	}
}

func TestToggle_EmptyListNoOp(t *testing.T) {
	m := modelWithTasks(0)
	// Should not panic.
	sendKey(m, " ")
}

// 3.5 – delete

func TestDelete_DCallsDelete(t *testing.T) {
	todos := []todo.Todo{{ID: 7, Title: "delete me"}}
	deleted := false
	repo := &testRepo{
		OnList:   func(_ todo.Filter) ([]todo.Todo, error) { return todos, nil },
		OnDelete: func(id int) error { deleted = (id == 7); return nil },
	}
	m := New(repo)
	m.Tasks = todos

	m = sendKey(m, "d") // shows confirm prompt
	sendKey(m, "y")     // confirms
	if !deleted {
		t.Error("Delete not called with correct ID")
	}
}

func TestDelete_EmptyListNoOp(t *testing.T) {
	m := modelWithTasks(0)
	sendKey(m, "d")
}

// 3.6 – quit

func TestQuit_QReturnsQuitCmd(t *testing.T) {
	m := modelWithTasks(0)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("q should return a cmd")
	}
	if cmd() != tea.Quit() {
		t.Error("q should return tea.Quit cmd")
	}
}

func TestQuit_CtrlCAlwaysQuits(t *testing.T) {
	m := modelWithTasks(0)
	m.InputMode = true
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil || cmd() != tea.Quit() {
		t.Error("Ctrl+C should always quit regardless of input mode")
	}
}

func TestQuit_QIgnoredInInputMode(t *testing.T) {
	m := modelWithTasks(0)
	m.InputMode = true
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd != nil {
		t.Error("q should be ignored in input mode")
	}
}
