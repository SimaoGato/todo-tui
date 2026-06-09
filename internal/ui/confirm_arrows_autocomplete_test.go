package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/justasandbox/my-todo-cli/internal/todo"
)

// ── 6.5 Delete confirmation ──────────────────────────────────────────────────

func TestConfirm_DSetsFlagButDoesNotDelete(t *testing.T) {
	deleted := false
	todos := []todo.Task{{ID: 1, Title: "task"}}
	repo := &testRepo{
		OnList:   func(_ todo.Filter) ([]todo.Task, error) { return todos, nil },
		OnDelete: func(_ int) error { deleted = true; return nil },
	}
	m := New(repo)
	m.Tasks = todos

	m = sendKey(m, "d")
	if !m.ConfirmDelete {
		t.Error("d should set ConfirmDelete=true")
	}
	if deleted {
		t.Error("d should not delete immediately")
	}
}

func TestConfirm_YConfirmsDelete(t *testing.T) {
	deleted := false
	todos := []todo.Task{{ID: 5, Title: "bye"}}
	repo := &testRepo{
		OnList:   func(_ todo.Filter) ([]todo.Task, error) { return todos, nil },
		OnDelete: func(id int) error { deleted = (id == 5); return nil },
	}
	m := New(repo)
	m.Tasks = todos

	m = sendKey(m, "d")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatal("y should return a cmd")
	}
	cmd()
	if !deleted {
		t.Error("y should confirm and call Delete")
	}
}

func TestConfirm_NCancels(t *testing.T) {
	deleted := false
	todos := []todo.Task{{ID: 3, Title: "keep"}}
	repo := &testRepo{
		OnList:   func(_ todo.Filter) ([]todo.Task, error) { return todos, nil },
		OnDelete: func(_ int) error { deleted = true; return nil },
	}
	m := New(repo)
	m.Tasks = todos

	m = sendKey(m, "d")
	m = sendKey(m, "n")
	if deleted {
		t.Error("n should cancel without deleting")
	}
	if m.ConfirmDelete {
		t.Error("n should clear ConfirmDelete flag")
	}
}

func TestConfirm_EscCancels(t *testing.T) {
	deleted := false
	todos := []todo.Task{{ID: 2, Title: "keep"}}
	repo := &testRepo{
		OnList:   func(_ todo.Filter) ([]todo.Task, error) { return todos, nil },
		OnDelete: func(_ int) error { deleted = true; return nil },
	}
	m := New(repo)
	m.Tasks = todos

	m = sendKey(m, "d")
	m = sendKeyType(m, tea.KeyEsc)
	if deleted {
		t.Error("esc should cancel without deleting")
	}
	if m.ConfirmDelete {
		t.Error("esc should clear ConfirmDelete flag")
	}
}

func TestConfirm_OtherKeysIgnored(t *testing.T) {
	deleted := false
	todos := []todo.Task{{ID: 1, Title: "t"}}
	repo := &testRepo{
		OnList:   func(_ todo.Filter) ([]todo.Task, error) { return todos, nil },
		OnDelete: func(_ int) error { deleted = true; return nil },
	}
	m := New(repo)
	m.Tasks = todos

	m = sendKey(m, "d")
	m = sendKey(m, "j") // navigation key — should be swallowed
	m = sendKey(m, "x") // random key — should be swallowed
	if deleted {
		t.Error("irrelevant keys should not trigger delete")
	}
	if !m.ConfirmDelete {
		t.Error("ConfirmDelete should still be set after irrelevant keys")
	}
}

func TestConfirm_EmptyListNoOp(t *testing.T) {
	m := modelWithTasks(0)
	m = sendKey(m, "d")
	if m.ConfirmDelete {
		t.Error("d on empty list should not set ConfirmDelete")
	}
}

func TestConfirm_ViewShowsPrompt(t *testing.T) {
	m := modelWithTasks(1)
	m.Tasks[0].Title = "my task"
	m = sendKey(m, "d")
	v := m.View()
	if !strings.Contains(v, "my task") {
		t.Error("confirm view should contain the task title")
	}
	if !strings.Contains(v, "[y/n]") {
		t.Error("confirm view should show [y/n] prompt")
	}
}

func TestConfirm_ViewShowsHelpBarWhenNotConfirming(t *testing.T) {
	m := modelWithTasks(1)
	v := m.View()
	if !strings.Contains(v, "q: quit") {
		t.Error("normal view should show help bar")
	}
}

// ── 6.6 Arrow key navigation ─────────────────────────────────────────────────

func TestArrow_DownMovesDown(t *testing.T) {
	m := modelWithTasks(3)
	m = sendKeyType(m, tea.KeyDown)
	if m.Cursor != 1 {
		t.Errorf("↓ cursor: got %d, want 1", m.Cursor)
	}
}

func TestArrow_UpMovesUp(t *testing.T) {
	m := modelWithTasks(3)
	m.Cursor = 2
	m = sendKeyType(m, tea.KeyUp)
	if m.Cursor != 1 {
		t.Errorf("↑ cursor: got %d, want 1", m.Cursor)
	}
}

func TestArrow_DownWraps(t *testing.T) {
	m := modelWithTasks(3)
	m.Cursor = 2
	m = sendKeyType(m, tea.KeyDown)
	if m.Cursor != 0 {
		t.Errorf("↓ wrap: got %d, want 0", m.Cursor)
	}
}

func TestArrow_UpWraps(t *testing.T) {
	m := modelWithTasks(3)
	m.Cursor = 0
	m = sendKeyType(m, tea.KeyUp)
	if m.Cursor != 2 {
		t.Errorf("↑ wrap: got %d, want 2", m.Cursor)
	}
}

func TestArrow_RightAdvancesTab(t *testing.T) {
	m := modelWithTasks(0) // starts at TabToday
	m = sendKeyType(m, tea.KeyRight)
	if m.ActiveTab != TabAll {
		t.Errorf("→ tab: got %d, want TabAll", m.ActiveTab)
	}
}

func TestArrow_LeftGoesBackTab(t *testing.T) {
	m := modelWithTasks(0)
	m.ActiveTab = TabAll
	m = sendKeyType(m, tea.KeyLeft)
	if m.ActiveTab != TabToday {
		t.Errorf("← tab: got %d, want TabToday", m.ActiveTab)
	}
}

func TestArrow_RightWrapsTab(t *testing.T) {
	m := modelWithTasks(0)
	m.ActiveTab = TabCompleted
	m = sendKeyType(m, tea.KeyRight)
	if m.ActiveTab != TabToday {
		t.Errorf("→ wrap: got %d, want TabToday", m.ActiveTab)
	}
}

func TestArrow_LeftWrapsTab(t *testing.T) {
	m := modelWithTasks(0)
	m.ActiveTab = TabToday
	m = sendKeyType(m, tea.KeyLeft)
	if m.ActiveTab != TabCompleted {
		t.Errorf("← wrap: got %d, want TabCompleted", m.ActiveTab)
	}
}

func TestArrow_IgnoredInInputMode(t *testing.T) {
	m := modelWithTasks(3)
	m.InputMode = true
	m = sendKeyType(m, tea.KeyDown)
	if m.Cursor != 0 {
		t.Error("↓ should be ignored in input mode")
	}
}

func TestArrow_IgnoredInConfirmMode(t *testing.T) {
	m := modelWithTasks(3)
	m.ConfirmDelete = true
	m = sendKeyType(m, tea.KeyDown)
	if m.Cursor != 0 {
		t.Error("↓ should be ignored in confirm mode")
	}
}

// ── 6.7 Due date autocomplete ─────────────────────────────────────────────────

func TestAutocomplete_TFillsTodayWhenEmpty(t *testing.T) {
	m := New(&testRepo{})
	m = sendKey(m, "a")
	m = typeString(m, "My task")
	m = sendKeyType(m, tea.KeyEnter) // advance to date step

	m = sendKey(m, "t") // autocomplete
	got := m.dateInput.Value()
	want := time.Now().Format("2006-01-02")
	if got != want {
		t.Errorf("autocomplete: got %q, want %q", got, want)
	}
}

func TestAutocomplete_TIgnoredWhenFieldNotEmpty(t *testing.T) {
	m := New(&testRepo{})
	m = sendKey(m, "a")
	m = typeString(m, "My task")
	m = sendKeyType(m, tea.KeyEnter)

	m = typeString(m, "2026") // field has content
	m = sendKey(m, "t")       // t should be appended as text, not autocomplete
	if m.dateInput.Value() == time.Now().Format("2006-01-02") {
		t.Error("t should not autocomplete when field already has text")
	}
}

func TestAutocomplete_TIgnoredInTitleStep(t *testing.T) {
	m := New(&testRepo{})
	m = sendKey(m, "a") // enter title step
	before := m.titleInput.Value()
	m = sendKey(m, "t")
	after := m.titleInput.Value()
	if after == time.Now().Format("2006-01-02") {
		t.Error("t autocomplete should not fire in title step")
	}
	_ = before
}

func TestConfirmBar_RendersTaskTitle(t *testing.T) {
	bar := renderConfirmBar("Buy groceries")
	if !strings.Contains(bar, "Buy groceries") {
		t.Error("confirm bar should contain task title")
	}
	if !strings.Contains(bar, "[y/n]") {
		t.Error("confirm bar should contain [y/n]")
	}
}

func TestConfirmBar_TruncatesLongTitle(t *testing.T) {
	long := strings.Repeat("x", 50)
	bar := renderConfirmBar(long)
	if strings.Contains(bar, long) {
		t.Error("confirm bar should truncate very long titles")
	}
}
