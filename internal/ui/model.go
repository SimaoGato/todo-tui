package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/justasandbox/my-todo-cli/internal/todo"
)

// Repo is the data access interface the TUI model depends on.
// db.Repository satisfies it implicitly.
type Repo interface {
	List(filter todo.Filter) ([]todo.Task, error)
	Create(title string, dueDate *time.Time) (todo.Task, error)
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

// todosLoadedMsg carries the result of an async List() call.
// cursor is the desired cursor position after the load (will be clamped).
type todosLoadedMsg struct {
	tasks  []todo.Task
	cursor int
	err    error
}

type toggleDoneMsg struct {
	err          error
	wasIncomplete bool
}

type deleteDoneMsg struct {
	err error
}

type createDoneMsg struct {
	err error
}

// fixedRows is the number of non-task lines in the normal view:
// tab bar + blank line + blank line before help + help bar.
const fixedRows = 4

type AppModel struct {
	Tasks         []todo.Task
	Cursor        int
	Offset        int
	ActiveTab     Tab
	InputMode     bool
	taskInput     TaskInput
	confirmDialog ConfirmDialog
	errorMsg      string
	Repo          Repo
	Width         int
	Height        int
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
	return AppModel{
		ActiveTab: TabToday,
		Repo:      repo,
		taskInput: newTaskInput(),
	}
}

// clampViewport adjusts m.Offset so that m.Cursor is always within the visible window.
// It is a no-op when Height==0 (no terminal size received yet) or when there are no tasks.
func clampViewport(m AppModel) AppModel {
	if m.Height == 0 || len(m.Tasks) == 0 {
		return m
	}
	visibleRows := m.Height - fixedRows
	if visibleRows < 1 {
		visibleRows = 1
	}
	if m.Cursor < m.Offset {
		m.Offset = m.Cursor
	}
	if m.Cursor >= m.Offset+visibleRows {
		m.Offset = m.Cursor - visibleRows + 1
	}
	return m
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
		tasks, err := m.Repo.List(tabToFilter(m.ActiveTab))
		return todosLoadedMsg{tasks: tasks, cursor: cursor, err: err}
	}
}

func cmdToggle(repo Repo, id int, wasIncomplete bool) tea.Cmd {
	return func() tea.Msg {
		return toggleDoneMsg{err: repo.ToggleDone(id), wasIncomplete: wasIncomplete}
	}
}

func cmdDelete(repo Repo, id int) tea.Cmd {
	return func() tea.Msg {
		return deleteDoneMsg{err: repo.Delete(id)}
	}
}

func cmdCreate(repo Repo, title string, dueDate *time.Time) tea.Cmd {
	return func() tea.Msg {
		_, err := repo.Create(title, dueDate)
		return createDoneMsg{err: err}
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
		m.Tasks = msg.tasks
		switch {
		case len(m.Tasks) == 0:
			m.Cursor = 0
		case msg.cursor >= len(m.Tasks):
			m.Cursor = len(m.Tasks) - 1
		default:
			m.Cursor = msg.cursor
		}
		m = clampViewport(m)
		return m, nil

	case toggleDoneMsg:
		if msg.err != nil {
			m.errorMsg = "toggle failed: " + msg.err.Error()
			return m, nil
		}
		if msg.wasIncomplete {
			m.SessCompleted++
		}
		return m, m.loadTodos()

	case deleteDoneMsg:
		if msg.err != nil {
			m.errorMsg = "delete failed: " + msg.err.Error()
			return m, nil
		}
		m.SessDeleted++
		return m, m.loadTodos()

	case createDoneMsg:
		if msg.err != nil {
			m.taskInput.err = "create failed: " + msg.err.Error()
			return m, nil
		}
		m.InputMode = false
		m.taskInput.Active = false
		m.taskInput.step = stepNone
		m.taskInput.err = ""
		m.taskInput.pending = ""
		return m, m.loadTodos()

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m = clampViewport(m)
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.InputMode {
			return m.updateInputMode(msg)
		}
		if m.confirmDialog.Active {
			return m.updateConfirmDelete(msg)
		}
		return m.updateNormalMode(msg)
	}

	return m, nil
}

func (m AppModel) updateInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	updated, cmd, signal, result := m.taskInput.Update(msg)
	m.taskInput = updated
	switch signal {
	case taskInputSignalCancel:
		m.InputMode = false
		return m, nil
	case taskInputSignalSubmit:
		return m, cmdCreate(m.Repo, result.Title, result.DueDate)
	}
	return m, cmd
}

func (m AppModel) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	updated, signal := m.confirmDialog.Update(msg)
	m.confirmDialog = updated
	if signal == confirmSignalYes && len(m.Tasks) > 0 {
		id := m.Tasks[m.Cursor].ID
		return m, cmdDelete(m.Repo, id)
	}
	return m, nil
}

func (m AppModel) updateNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear any displayed error on the next key press in normal mode.
	m.errorMsg = ""

	switch msg.String() {

	case "a":
		m.InputMode = true
		var focusCmd tea.Cmd
		m.taskInput, focusCmd = m.taskInput.activate()
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
		m.Offset = 0
		return m, m.loadTodos()

	case "2":
		m.ActiveTab = TabAll
		m.Cursor = 0
		m.Offset = 0
		return m, m.loadTodos()

	case "3":
		m.ActiveTab = TabCompleted
		m.Cursor = 0
		m.Offset = 0
		return m, m.loadTodos()

	case "tab":
		m.ActiveTab = (m.ActiveTab + 1) % tabCount
		m.Cursor = 0
		m.Offset = 0
		return m, m.loadTodos()

	case " ":
		if len(m.Tasks) > 0 {
			id := m.Tasks[m.Cursor].ID
			wasIncomplete := !m.Tasks[m.Cursor].Done
			return m, cmdToggle(m.Repo, id, wasIncomplete)
		}

	case "d":
		if len(m.Tasks) > 0 {
			m.confirmDialog.Active = true
			m.confirmDialog.prompt = m.Tasks[m.Cursor].Title
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
		m.Offset = 0
		return m, m.loadTodos()

	case "left":
		m.ActiveTab = (m.ActiveTab + 2) % tabCount
		m.Cursor = 0
		m.Offset = 0
		return m, m.loadTodos()
	}

	m = clampViewport(m)
	return m, nil
}
