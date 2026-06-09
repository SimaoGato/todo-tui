package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/justasandbox/my-todo-cli/internal/todo"
)

type testRepo struct {
	OnList       func(todo.Filter) ([]todo.Task, error)
	OnCreate     func(string, *time.Time) (todo.Task, error)
	OnToggleDone func(int) error
	OnDelete     func(int) error
}

func (r *testRepo) List(f todo.Filter) ([]todo.Task, error) {
	if r.OnList != nil {
		return r.OnList(f)
	}
	return nil, nil
}

func (r *testRepo) Create(title string, d *time.Time) (todo.Task, error) {
	if r.OnCreate != nil {
		return r.OnCreate(title, d)
	}
	return todo.Task{}, nil
}

func (r *testRepo) ToggleDone(id int) error {
	if r.OnToggleDone != nil {
		return r.OnToggleDone(id)
	}
	return nil
}

func (r *testRepo) Delete(id int) error {
	if r.OnDelete != nil {
		return r.OnDelete(id)
	}
	return nil
}

func modelWithTasks(n int) AppModel {
	todos := make([]todo.Task, n)
	for i := range todos {
		todos[i] = todo.Task{ID: i + 1, Title: "task"}
	}
	repo := &testRepo{OnList: func(_ todo.Filter) ([]todo.Task, error) { return todos, nil }}
	m := New(repo)
	m.Tasks = todos
	return m
}

func sendKey(m AppModel, key string) AppModel {
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return next.(AppModel)
}

func sendKeyType(m AppModel, kt tea.KeyType) AppModel {
	next, _ := m.Update(tea.KeyMsg{Type: kt})
	return next.(AppModel)
}

func typeString(m AppModel, s string) AppModel {
	for _, r := range s {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(AppModel)
	}
	return m
}
