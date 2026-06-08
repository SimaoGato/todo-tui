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
		return m, nil

	case tea.KeyMsg:
		// Ctrl+C always quits (3.6).
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		// While in input mode only Esc is handled (input flow is Epic 4).
		if m.InputMode {
			if msg.Type == tea.KeyEsc {
				m.InputMode = false
			}
			return m, nil
		}

		switch msg.String() {

		// 3.6 – quit
		case "q":
			return m, tea.Quit

		// 3.2 – navigation
		case "j":
			if len(m.Tasks) > 0 {
				m.Cursor = (m.Cursor + 1) % len(m.Tasks)
			}

		case "k":
			if len(m.Tasks) > 0 {
				m.Cursor = (m.Cursor - 1 + len(m.Tasks)) % len(m.Tasks)
			}

		// 3.3 – tab switching
		case "1":
			m.ActiveTab = TabToday
			m.Cursor = 0
			return m, m.loadTodos()

		case "2":
			m.ActiveTab = TabAll
			m.Cursor = 0
			return m, m.loadTodos()

		case "3":
			m.ActiveTab = TabCompleted
			m.Cursor = 0
			return m, m.loadTodos()

		case "tab":
			m.ActiveTab = (m.ActiveTab + 1) % 3
			m.Cursor = 0
			return m, m.loadTodos()

		// 3.4 – toggle completion
		case " ":
			if len(m.Tasks) > 0 {
				_ = m.Repo.ToggleDone(m.Tasks[m.Cursor].ID)
				if m.Cursor >= len(m.Tasks)-1 {
					m.Cursor = max(0, len(m.Tasks)-2)
				}
				return m, m.loadTodos()
			}

		// 3.5 – delete
		case "d":
			if len(m.Tasks) > 0 {
				_ = m.Repo.Delete(m.Tasks[m.Cursor].ID)
				if m.Cursor >= len(m.Tasks)-1 {
					m.Cursor = max(0, len(m.Tasks)-2)
				}
				return m, m.loadTodos()
			}
		}
	}

	return m, nil
}

func (m AppModel) View() string {
	return "Hello, Todo\n"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
