package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Repo is the data access interface the TUI model depends on.
// db.Repository satisfies it implicitly.
type Repo interface {
	List(filter Filter) ([]Todo, error)
	Create(title string, dueDate *time.Time) (Todo, error)
	ToggleDone(id int) error
	Delete(id int) error
}

type Tab int

const (
	TabToday Tab = iota
	TabAll
	TabCompleted
)

// todosLoadedMsg carries the result of an async List() call.
type todosLoadedMsg struct {
	todos []Todo
}

type AppModel struct {
	Tasks     []Todo
	Cursor    int
	ActiveTab Tab
	InputMode bool
	Repo      Repo
}

func New(repo Repo) AppModel {
	return AppModel{
		ActiveTab: TabToday,
		Repo:      repo,
	}
}

func tabToFilter(tab Tab) Filter {
	switch tab {
	case TabAll:
		return FilterAll
	case TabCompleted:
		return FilterDone
	default:
		return FilterToday
	}
}

func (m AppModel) loadTodos() tea.Cmd {
	return func() tea.Msg {
		todos, _ := m.Repo.List(tabToFilter(m.ActiveTab))
		return todosLoadedMsg{todos: todos}
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.loadTodos()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case todosLoadedMsg:
		m.Tasks = msg.todos
		m.Cursor = 0
	}
	return m, nil
}

func (m AppModel) View() string {
	return "Hello, Todo\n"
}
