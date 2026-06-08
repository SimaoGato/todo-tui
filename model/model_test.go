package model

import (
	"testing"
	"time"
)

// mockRepo implements Repo for testing without a real database.
type mockRepo struct {
	todos []Todo
}

func (m *mockRepo) List(_ Filter) ([]Todo, error)                       { return m.todos, nil }
func (m *mockRepo) Create(_ string, _ *time.Time) (Todo, error)         { return Todo{}, nil }
func (m *mockRepo) ToggleDone(_ int) error                              { return nil }
func (m *mockRepo) Delete(_ int) error                                  { return nil }

func TestNew_Defaults(t *testing.T) {
	m := New(&mockRepo{})

	if m.Cursor != 0 {
		t.Errorf("cursor: got %d, want 0", m.Cursor)
	}
	if m.InputMode {
		t.Error("expected inputMode=false")
	}
	if m.ActiveTab != TabToday {
		t.Errorf("activeTab: got %d, want TabToday", m.ActiveTab)
	}
	if len(m.Tasks) != 0 {
		t.Errorf("expected empty tasks, got %d", len(m.Tasks))
	}
}

func TestInit_ReturnsCmdThatLoadsTodos(t *testing.T) {
	repo := &mockRepo{todos: []Todo{{ID: 1, Title: "test todo"}}}
	m := New(repo)

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil cmd")
	}

	msg := cmd()
	loaded, ok := msg.(todosLoadedMsg)
	if !ok {
		t.Fatalf("expected todosLoadedMsg, got %T", msg)
	}
	if len(loaded.todos) != 1 || loaded.todos[0].Title != "test todo" {
		t.Errorf("unexpected todos: %v", loaded.todos)
	}
}

func TestUpdate_TodosLoadedMsg_PopulatesTasks(t *testing.T) {
	m := New(&mockRepo{})

	todos := []Todo{{ID: 1, Title: "a"}, {ID: 2, Title: "b"}}
	next, _ := m.Update(todosLoadedMsg{todos: todos})

	am := next.(AppModel)
	if len(am.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(am.Tasks))
	}
	if am.Cursor != 0 {
		t.Errorf("cursor after load: got %d, want 0", am.Cursor)
	}
}

func TestTabToFilter(t *testing.T) {
	cases := []struct {
		tab    Tab
		filter Filter
	}{
		{TabToday, FilterToday},
		{TabAll, FilterAll},
		{TabCompleted, FilterDone},
	}
	for _, c := range cases {
		if got := tabToFilter(c.tab); got != c.filter {
			t.Errorf("tabToFilter(%d) = %d, want %d", c.tab, got, c.filter)
		}
	}
}

func TestInit_UsesTodayFilterByDefault(t *testing.T) {
	var calledWith Filter
	repo := &mockRepo{}
	// Override List to capture which filter is used.
	spy := &spyRepo{inner: repo, onList: func(f Filter) { calledWith = f }}

	m := New(spy)
	cmd := m.Init()
	cmd() // execute the command

	if calledWith != FilterToday {
		t.Errorf("Init() used filter %d, want FilterToday (%d)", calledWith, FilterToday)
	}
}

type spyRepo struct {
	inner  Repo
	onList func(Filter)
}

func (s *spyRepo) List(f Filter) ([]Todo, error) {
	s.onList(f)
	return s.inner.List(f)
}
func (s *spyRepo) Create(title string, d *time.Time) (Todo, error) { return s.inner.Create(title, d) }
func (s *spyRepo) ToggleDone(id int) error                         { return s.inner.ToggleDone(id) }
func (s *spyRepo) Delete(id int) error                             { return s.inner.Delete(id) }
