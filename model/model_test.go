package model

import (
	"testing"

	"github.com/justasandbox/my-todo-cli/todo"
)

func TestNew_Defaults(t *testing.T) {
	m := New(&testRepo{})

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
	todos := []todo.Todo{{ID: 1, Title: "test todo"}}
	repo := &testRepo{OnList: func(_ todo.Filter) ([]todo.Todo, error) { return todos, nil }}
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
	m := New(&testRepo{})

	todos := []todo.Todo{{ID: 1, Title: "a"}, {ID: 2, Title: "b"}}
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
		filter todo.Filter
	}{
		{TabToday, todo.FilterToday},
		{TabAll, todo.FilterAll},
		{TabCompleted, todo.FilterDone},
	}
	for _, c := range cases {
		if got := tabToFilter(c.tab); got != c.filter {
			t.Errorf("tabToFilter(%d) = %d, want %d", c.tab, got, c.filter)
		}
	}
}

func TestInit_UsesTodayFilterByDefault(t *testing.T) {
	var calledWith todo.Filter
	repo := &testRepo{OnList: func(f todo.Filter) ([]todo.Todo, error) {
		calledWith = f
		return nil, nil
	}}

	m := New(repo)
	cmd := m.Init()
	cmd()

	if calledWith != todo.FilterToday {
		t.Errorf("Init() used filter %d, want FilterToday (%d)", calledWith, todo.FilterToday)
	}
}
