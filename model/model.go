package model

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
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

type inputStep int

const (
	stepNone  inputStep = iota
	stepTitle
	stepDate
)

// todosLoadedMsg carries the result of an async List() call.
// cursor is the desired cursor position after the load (will be clamped).
type todosLoadedMsg struct {
	todos  []Todo
	cursor int
}

type AppModel struct {
	Tasks         []Todo
	Cursor        int
	ActiveTab     Tab
	InputMode     bool
	inputStep     inputStep
	titleInput    textinput.Model
	dateInput     textinput.Model
	pendingTitle  string
	inputErr      string
	Repo          Repo
	Width         int
	Height        int
	ConfirmDelete bool
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
	cursor := m.Cursor
	return func() tea.Msg {
		todos, _ := m.Repo.List(tabToFilter(m.ActiveTab))
		return todosLoadedMsg{todos: todos, cursor: cursor}
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.loadTodos()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case todosLoadedMsg:
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
		// Ctrl+C always quits (3.6).
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		if m.InputMode {
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

		// 6.5 – confirm-delete gate: only y / n / esc are processed.
		if m.ConfirmDelete {
			switch {
			case msg.String() == "y":
				_ = m.Repo.Delete(m.Tasks[m.Cursor].ID)
				m.ConfirmDelete = false
				return m, m.loadTodos()
			case msg.String() == "n", msg.Type == tea.KeyEsc:
				m.ConfirmDelete = false
			}
			return m, nil
		}

		switch msg.String() {

		// 4.1 – enter add mode
		case "a":
			m.InputMode = true
			m.inputStep = stepTitle
			m.inputErr = ""
			m.titleInput.Reset()
			focusCmd := m.titleInput.Focus()
			return m, focusCmd

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
				return m, m.loadTodos()
			}

		// 3.5 / 6.5 – delete: show confirmation prompt first
		case "d":
			if len(m.Tasks) > 0 {
				m.ConfirmDelete = true
			}

		// 6.6 – arrow key navigation
		case "up":
			if len(m.Tasks) > 0 {
				m.Cursor = (m.Cursor - 1 + len(m.Tasks)) % len(m.Tasks)
			}

		case "down":
			if len(m.Tasks) > 0 {
				m.Cursor = (m.Cursor + 1) % len(m.Tasks)
			}

		case "right":
			m.ActiveTab = (m.ActiveTab + 1) % 3
			m.Cursor = 0
			return m, m.loadTodos()

		case "left":
			m.ActiveTab = (m.ActiveTab + 2) % 3
			m.Cursor = 0
			return m, m.loadTodos()
		}
	}

	return m, nil
}

