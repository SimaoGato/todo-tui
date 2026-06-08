package model

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// spyCreateRepo captures Create calls for assertions.
type spyCreateRepo struct {
	todos        []Todo
	createCalled bool
	createdTitle string
	createdDate  *time.Time
}

func (s *spyCreateRepo) List(_ Filter) ([]Todo, error)  { return s.todos, nil }
func (s *spyCreateRepo) ToggleDone(_ int) error          { return nil }
func (s *spyCreateRepo) Delete(_ int) error              { return nil }
func (s *spyCreateRepo) Create(title string, dueDate *time.Time) (Todo, error) {
	s.createCalled = true
	s.createdTitle = title
	s.createdDate = dueDate
	return Todo{ID: 99, Title: title}, nil
}

// typeString sends each rune as a separate KeyRunes message.
func typeString(m AppModel, s string) AppModel {
	for _, r := range s {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(AppModel)
	}
	return m
}

// 4.1 – enter add mode

func TestInput_APressEntersInputMode(t *testing.T) {
	m := New(&mockRepo{})
	m = sendKey(m, "a")
	if !m.InputMode {
		t.Error("pressing 'a' should set InputMode=true")
	}
	if m.inputStep != stepTitle {
		t.Errorf("expected inputStep=stepTitle, got %d", m.inputStep)
	}
}

func TestInput_EscExitsInputMode(t *testing.T) {
	m := New(&mockRepo{})
	m = sendKey(m, "a")
	m = sendKeyType(m, tea.KeyEsc)
	if m.InputMode {
		t.Error("Esc should exit input mode")
	}
	if m.inputStep != stepNone {
		t.Errorf("expected inputStep=stepNone after Esc, got %d", m.inputStep)
	}
}

func TestInput_NavigationIgnoredInInputMode(t *testing.T) {
	m := modelWithTasks(3)
	m = sendKey(m, "a")
	m = sendKey(m, "j")
	if m.Cursor != 0 {
		t.Error("navigation keys should be ignored while in input mode")
	}
}

func TestInput_CharactersAddedToBuffer(t *testing.T) {
	m := New(&mockRepo{})
	m = sendKey(m, "a")
	m = typeString(m, "hello")
	if m.titleInput.Value() != "hello" {
		t.Errorf("title buffer: got %q, want %q", m.titleInput.Value(), "hello")
	}
}

// 4.2 – due date prompt

func TestInput_EmptyTitleDoesNotAdvance(t *testing.T) {
	m := New(&mockRepo{})
	m = sendKey(m, "a")
	m = sendKeyType(m, tea.KeyEnter)
	if m.inputStep != stepTitle {
		t.Error("empty title should not advance to date step")
	}
}

func TestInput_TitleEnterMovesToDateStep(t *testing.T) {
	m := New(&mockRepo{})
	m = sendKey(m, "a")
	m = typeString(m, "Buy milk")
	m = sendKeyType(m, tea.KeyEnter)
	if m.inputStep != stepDate {
		t.Errorf("after title Enter, expected stepDate, got %d", m.inputStep)
	}
	if m.pendingTitle != "Buy milk" {
		t.Errorf("pendingTitle: got %q, want %q", m.pendingTitle, "Buy milk")
	}
}

func TestInput_InvalidDateShowsError(t *testing.T) {
	m := New(&mockRepo{})
	m = sendKey(m, "a")
	m = typeString(m, "Task")
	m = sendKeyType(m, tea.KeyEnter) // confirm title
	m = typeString(m, "not-a-date")
	m = sendKeyType(m, tea.KeyEnter) // invalid date
	if m.inputErr == "" {
		t.Error("expected inputErr to be set for invalid date")
	}
	if !m.InputMode {
		t.Error("should remain in input mode after invalid date")
	}
}

func TestInput_EscDuringDateStepCancels(t *testing.T) {
	repo := &spyCreateRepo{}
	m := New(repo)
	m = sendKey(m, "a")
	m = typeString(m, "Task")
	m = sendKeyType(m, tea.KeyEnter) // confirm title
	m = sendKeyType(m, tea.KeyEsc)   // cancel during date step
	if m.InputMode {
		t.Error("Esc should exit input mode from date step")
	}
	if repo.createCalled {
		t.Error("Create should not be called after Esc")
	}
}

// 4.3 – save and return to list

func TestInput_EmptyDateCreatesTaskWithNoDueDate(t *testing.T) {
	repo := &spyCreateRepo{}
	m := New(repo)
	m = sendKey(m, "a")
	m = typeString(m, "Buy milk")
	m = sendKeyType(m, tea.KeyEnter) // confirm title
	m = sendKeyType(m, tea.KeyEnter) // skip date
	if !repo.createCalled {
		t.Fatal("Create should have been called")
	}
	if repo.createdTitle != "Buy milk" {
		t.Errorf("createdTitle: got %q, want %q", repo.createdTitle, "Buy milk")
	}
	if repo.createdDate != nil {
		t.Error("expected nil due date when date skipped")
	}
	if m.InputMode {
		t.Error("should have exited input mode after save")
	}
}

func TestInput_ValidDateCreatesTaskWithDueDate(t *testing.T) {
	repo := &spyCreateRepo{}
	m := New(repo)
	m = sendKey(m, "a")
	m = typeString(m, "Task")
	m = sendKeyType(m, tea.KeyEnter) // confirm title
	m = typeString(m, "2026-12-25")
	m = sendKeyType(m, tea.KeyEnter) // confirm date
	if !repo.createCalled {
		t.Fatal("Create should have been called")
	}
	if repo.createdDate == nil {
		t.Fatal("expected non-nil due date")
	}
	if got := repo.createdDate.Format("2006-01-02"); got != "2026-12-25" {
		t.Errorf("createdDate: got %q, want %q", got, "2026-12-25")
	}
	if m.InputMode {
		t.Error("should have exited input mode after save")
	}
}

func TestInput_SaveExitsAndReloads(t *testing.T) {
	repo := &spyCreateRepo{}
	m := New(repo)
	m = sendKey(m, "a")
	m = typeString(m, "Task")
	m = sendKeyType(m, tea.KeyEnter)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// cmd should be a loadTodos cmd (not nil)
	if cmd == nil {
		t.Error("saving should return a reload cmd")
	}
}
