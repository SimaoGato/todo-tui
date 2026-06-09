package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type inputStep int

const (
	stepNone  inputStep = iota
	stepTitle
	stepDate
)

type taskInputSignal int

const (
	taskInputSignalNone   taskInputSignal = iota
	taskInputSignalSubmit                 // valid title+date entered; result is non-nil
	taskInputSignalCancel                 // Esc pressed
)

// TaskInputResult carries the values produced by a completed add-task flow.
type TaskInputResult struct {
	Title   string
	DueDate *time.Time
}

// TaskInput manages the two-step add-task input flow (title → date).
type TaskInput struct {
	step       inputStep
	titleInput textinput.Model
	dateInput  textinput.Model
	pending    string
	err        string
	Active     bool
}

func newTaskInput() TaskInput {
	ti := textinput.New()
	ti.Placeholder = "Task title…"
	ti.CharLimit = 256

	di := textinput.New()
	di.Placeholder = "YYYY-MM-DD"
	di.CharLimit = 10

	return TaskInput{titleInput: ti, dateInput: di}
}

// activate resets state for a fresh session and returns a focus cmd for the title field.
func (ti TaskInput) activate() (TaskInput, tea.Cmd) {
	ti.Active = true
	ti.step = stepTitle
	ti.err = ""
	ti.pending = ""
	ti.titleInput.Reset()
	return ti, ti.titleInput.Focus()
}

// Update processes a key event during the active input flow.
// The returned Cmd is only for focus delegation; repo calls are never made here.
// On taskInputSignalSubmit the result is non-nil. On taskInputSignalCancel the TaskInput
// is deactivated. On taskInputSignalNone neither applies.
func (ti TaskInput) Update(msg tea.KeyMsg) (TaskInput, tea.Cmd, taskInputSignal, *TaskInputResult) {
	switch msg.Type {
	case tea.KeyEsc:
		ti.Active = false
		ti.step = stepNone
		ti.err = ""
		ti.pending = ""
		return ti, nil, taskInputSignalCancel, nil

	case tea.KeyEnter:
		if ti.step == stepTitle {
			title := strings.TrimSpace(ti.titleInput.Value())
			if title == "" {
				return ti, nil, taskInputSignalNone, nil
			}
			ti.pending = title
			ti.step = stepDate
			ti.dateInput.Reset()
			ti.titleInput.Blur()
			return ti, ti.dateInput.Focus(), taskInputSignalNone, nil
		}
		// stepDate: validate and signal submit (caller owns cmdCreate).
		dateStr := strings.TrimSpace(ti.dateInput.Value())
		var dueDate *time.Time
		if dateStr != "" {
			t, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				ti.err = "invalid date, use YYYY-MM-DD"
				return ti, nil, taskInputSignalNone, nil
			}
			dueDate = &t
		}
		return ti, nil, taskInputSignalSubmit, &TaskInputResult{Title: ti.pending, DueDate: dueDate}
	}

	// 't' fills today's date when the date field is empty.
	if ti.step == stepDate && msg.Type == tea.KeyRunes &&
		string(msg.Runes) == "t" && ti.dateInput.Value() == "" {
		ti.dateInput.SetValue(time.Now().Format("2006-01-02"))
		ti.dateInput.CursorEnd()
		return ti, nil, taskInputSignalNone, nil
	}

	// Delegate remaining keys to the active text field.
	var cmd tea.Cmd
	if ti.step == stepTitle {
		ti.titleInput, cmd = ti.titleInput.Update(msg)
	} else {
		ti.dateInput, cmd = ti.dateInput.Update(msg)
	}
	return ti, cmd, taskInputSignalNone, nil
}

// View renders the input form (title step or date step with optional error).
func (ti TaskInput) View() string {
	var sb strings.Builder
	if ti.step == stepTitle {
		sb.WriteString("New task: ")
		sb.WriteString(ti.titleInput.View())
	} else {
		sb.WriteString("Due date (YYYY-MM-DD, Enter to skip): ")
		sb.WriteString(ti.dateInput.View())
		if ti.err != "" {
			sb.WriteString("\n")
			sb.WriteString(styleError.Render(ti.err))
		}
	}
	sb.WriteString("\n\n")
	sb.WriteString(renderHelpBar(true))
	sb.WriteString("\n")
	return sb.String()
}
