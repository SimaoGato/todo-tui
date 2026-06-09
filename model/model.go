package model

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/justasandbox/my-todo-cli/todo"
)

// Repo is the data access interface the TUI model depends on.
// db.Repository satisfies it implicitly.
type Repo interface {
	List(filter todo.Filter) ([]todo.Todo, error)
	Create(title string, dueDate *time.Time) (todo.Todo, error)
	ToggleDone(id int) error
	Delete(id int) error
}

type Tab int

const (
	TabToday Tab = iota
	TabAll
	TabCompleted
	tabCount
)

type inputStep int

const (
	stepNone  inputStep = iota
	stepTitle
	stepDate
)

// todosLoadedMsg carries the result of an async List() call.
// cursor is the desired cursor position after the load (will be clamped).
type todosLoadedMsg struct {
	todos  []todo.Todo
	cursor int
	err    error
}

type AppModel struct {
	Tasks         []todo.Todo
	Cursor        int
	ActiveTab     Tab
	InputMode     bool
	inputStep     inputStep
	titleInput    textinput.Model
	dateInput     textinput.Model
	pendingTitle  string
	inputErr      string
	errorMsg      string
	Repo          Repo
	Width         int
	Height        int
	ConfirmDelete bool
	SessCompleted int
	SessDeleted   int
}

// Summary holds per-session counters exposed to main.go after the program exits.
type Summary struct {
	Completed      int
	Deleted        int
	TodayRemaining int
}

func (m AppModel) GetSummary() Summary {
	return Summary{
		Completed: m.SessCompleted,
		Deleted:   m.SessDeleted,
	}
}

func New(repo Repo) AppModel {
	ti := textinput.New()
	ti.Placeholder = "Task title…"
	ti.CharLimit = 256

	di := textinput.New()
	di.Placeholder = "YYYY-MM-DD"
	di.CharLimit = 10

	return AppModel{
		ActiveTab:  TabToday,
		Repo:       repo,
		titleInput: ti,
		dateInput:  di,
	}
}

func tabToFilter(tab Tab) todo.Filter {
	switch tab {
	case TabAll:
		return todo.FilterAll
	case TabCompleted:
		return todo.FilterDone
	default:
		return todo.FilterToday
	}
}

func (m AppModel) loadTodos() tea.Cmd {
	cursor := m.Cursor
	return func() tea.Msg {
		todos, err := m.Repo.List(tabToFilter(m.ActiveTab))
		return todosLoadedMsg{todos: todos, cursor: cursor, err: err}
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.loadTodos()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case todosLoadedMsg:
		if msg.err != nil {
			m.errorMsg = "load failed: " + msg.err.Error()
			return m, nil
		}
		m.errorMsg = ""
		m.Tasks = msg.todos
		switch {
		case len(m.Tasks) == 0:
			m.Cursor = 0
		case msg.cursor >= len(m.Tasks):
			m.Cursor = len(m.Tasks) - 1
		default:
			m.Cursor = msg.cursor
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.InputMode {
			return m.updateInputMode(msg)
		}
		if m.ConfirmDelete {
			return m.updateConfirmDelete(msg)
		}
		return m.updateNormalMode(msg)
	}

	return m, nil
}

func (m AppModel) updateInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.InputMode = false
		m.inputStep = stepNone
		m.inputErr = ""
		m.pendingTitle = ""
		return m, nil

	case tea.KeyEnter:
		if m.inputStep == stepTitle {
			title := strings.TrimSpace(m.titleInput.Value())
			if title == "" {
				return m, nil
			}
			m.pendingTitle = title
			m.inputStep = stepDate
			m.dateInput.Reset()
			m.titleInput.Blur()
			focusCmd := m.dateInput.Focus()
			return m, focusCmd
		}
		// stepDate: validate and save
		dateStr := strings.TrimSpace(m.dateInput.Value())
		var dueDate *time.Time
		if dateStr != "" {
			t, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				m.inputErr = "invalid date, use YYYY-MM-DD"
				return m, nil
			}
			dueDate = &t
		}
		if _, err := m.Repo.Create(m.pendingTitle, dueDate); err != nil {
			m.errorMsg = "create failed: " + err.Error()
			m.InputMode = false
			m.inputStep = stepNone
			m.inputErr = ""
			m.pendingTitle = ""
			return m, nil
		}
		m.InputMode = false
		m.inputStep = stepNone
		m.inputErr = ""
		m.pendingTitle = ""
		return m, m.loadTodos()
	}

	// 6.7 – 't' fills today's date when the date field is empty
	if m.inputStep == stepDate && msg.Type == tea.KeyRunes &&
		string(msg.Runes) == "t" && m.dateInput.Value() == "" {
		m.dateInput.SetValue(time.Now().Format("2006-01-02"))
		m.dateInput.CursorEnd()
		return m, nil
	}

	// Delegate remaining keys to the active input field.
	var cmd tea.Cmd
	if m.inputStep == stepTitle {
		m.titleInput, cmd = m.titleInput.Update(msg)
	} else {
		m.dateInput, cmd = m.dateInput.Update(msg)
	}
	return m, cmd
}

func (m AppModel) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "y":
		if err := m.Repo.Delete(m.Tasks[m.Cursor].ID); err != nil {
			m.errorMsg = "delete failed: " + err.Error()
			m.ConfirmDelete = false
			return m, nil
		}
		m.SessDeleted++
		m.ConfirmDelete = false
		return m, m.loadTodos()
	case msg.String() == "n", msg.Type == tea.KeyEsc:
		m.ConfirmDelete = false
	}
	return m, nil
}

func (m AppModel) updateNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear any displayed error on the next key press in normal mode.
	m.errorMsg = ""

	switch msg.String() {

	case "a":
		m.InputMode = true
		m.inputStep = stepTitle
		m.inputErr = ""
		m.titleInput.Reset()
		focusCmd := m.titleInput.Focus()
		return m, focusCmd

	case "q":
		return m, tea.Quit

	case "j":
		if len(m.Tasks) > 0 {
			m.Cursor = (m.Cursor + 1) % len(m.Tasks)
		}

	case "k":
		if len(m.Tasks) > 0 {
			m.Cursor = (m.Cursor - 1 + len(m.Tasks)) % len(m.Tasks)
		}

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
		m.ActiveTab = (m.ActiveTab + 1) % tabCount
		m.Cursor = 0
		return m, m.loadTodos()

	case " ":
		if len(m.Tasks) > 0 {
			wasIncomplete := !m.Tasks[m.Cursor].Done
			if err := m.Repo.ToggleDone(m.Tasks[m.Cursor].ID); err != nil {
				m.errorMsg = "toggle failed: " + err.Error()
				return m, nil
			}
			if wasIncomplete {
				m.SessCompleted++
			}
			return m, m.loadTodos()
		}

	case "d":
		if len(m.Tasks) > 0 {
			m.ConfirmDelete = true
		}

	case "up":
		if len(m.Tasks) > 0 {
			m.Cursor = (m.Cursor - 1 + len(m.Tasks)) % len(m.Tasks)
		}

	case "down":
		if len(m.Tasks) > 0 {
			m.Cursor = (m.Cursor + 1) % len(m.Tasks)
		}

	case "right":
		m.ActiveTab = (m.ActiveTab + 1) % tabCount
		m.Cursor = 0
		return m, m.loadTodos()

	case "left":
		m.ActiveTab = (m.ActiveTab + 2) % tabCount
		m.Cursor = 0
		return m, m.loadTodos()
	}

	return m, nil
}
