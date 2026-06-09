package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/justasandbox/my-todo-cli/internal/todo"
)

const defaultTitleWidth = 40

var (
	styleTabActive   = lipgloss.NewStyle().Bold(true).Underline(true)
	styleTabInactive = lipgloss.NewStyle().Faint(true)
	styleBold        = lipgloss.NewStyle().Bold(true)
	styleOverdue     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleDueToday    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleDone        = lipgloss.NewStyle().Faint(true).Strikethrough(true)
	styleHelp        = lipgloss.NewStyle().Faint(true)
	styleError       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleConfirm     = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
)

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-3]) + "..."
}

// 5.1 – tab bar

func renderTabBar(active Tab) string {
	entries := []struct {
		name string
		tab  Tab
	}{
		{"Today", TabToday},
		{"All", TabAll},
		{"Completed", TabCompleted},
	}
	parts := make([]string, len(entries))
	for i, e := range entries {
		label := "[ " + e.name + " ]"
		if e.tab == active {
			parts[i] = styleTabActive.Render(label)
		} else {
			parts[i] = styleTabInactive.Render(label)
		}
	}
	return strings.Join(parts, "  ")
}

// taskDateClass returns "overdue", "today", or "" by comparing calendar dates in their
// own timezones. Using Year/Month/Day avoids the UTC-aligned Truncate(24h) bug that
// misclassifies tasks in negative-UTC-offset zones near midnight.
func taskDateClass(dueDate *time.Time, now time.Time) string {
	if dueDate == nil {
		return ""
	}
	dY, dM, dD := dueDate.Date()
	nY, nM, nD := now.Date()
	if dY == nY && dM == nM && dD == nD {
		return "today"
	}
	if dY < nY || (dY == nY && (dM < nM || (dM == nM && dD < nD))) {
		return "overdue"
	}
	return ""
}

// 5.2 + 5.3 – task row with color coding

func renderTaskRow(task todo.Todo, isCursor bool, now time.Time, titleWidth int) string {
	cursor := "  "
	if isCursor {
		cursor = "> "
	}
	check := "[ ]"
	if task.Done {
		check = "[x]"
	}
	dueStr := "—"
	if task.DueDate != nil {
		dueStr = task.DueDate.Format("2006-01-02")
	}
	title := fmt.Sprintf("%-*s", titleWidth, truncate(task.Title, titleWidth))
	line := cursor + check + " " + title + "  " + dueStr

	if task.Done {
		return styleDone.Render(line)
	}
	switch taskDateClass(task.DueDate, now) {
	case "overdue":
		return styleOverdue.Render(line)
	case "today":
		return styleDueToday.Render(line)
	}
	if isCursor {
		return styleBold.Render(line)
	}
	return line
}

// 6.5 – confirm-delete bar

func renderConfirmBar(title string) string {
	label := "Delete \"" + truncate(title, 30) + "\"? [y/n]"
	return styleConfirm.Render(label)
}

// 5.4 – help bar

func renderHelpBar(inputMode bool) string {
	if inputMode {
		return styleHelp.Render("enter: confirm  esc: cancel")
	}
	return styleHelp.Render("j/k: navigate  a: add  space: toggle  d: delete  q: quit")
}

// 5.5 – empty state

func renderEmptyState(tab Tab) string {
	msgs := map[Tab]string{
		TabToday:     "No tasks for today",
		TabAll:       "No tasks",
		TabCompleted: "No completed tasks",
	}
	return styleHelp.Render(msgs[tab])
}

// viewportBounds returns the [start, end) index range of Tasks to render.
// When Height==0, all tasks are returned (fallback for startup before WindowSizeMsg).
func (m AppModel) viewportBounds() (start, end int) {
	start = m.Offset
	end = len(m.Tasks)
	if m.Height > 0 {
		visibleRows := m.Height - fixedRows
		if visibleRows < 1 {
			visibleRows = 1
		}
		if end > start+visibleRows {
			end = start + visibleRows
		}
	}
	// Guard against a stale offset (e.g. tasks shrunk since last clampViewport).
	if start < 0 {
		start = 0
	}
	if start > end {
		start = 0
	}
	return
}

// titleColWidth returns the title column width based on terminal width.
func (m AppModel) titleColWidth() int {
	// prefix: "> [x] " = 6, suffix: "  YYYY-MM-DD" = 12, min title = 20
	if m.Width > 0 {
		if w := m.Width - 20; w >= 20 {
			return w
		}
		return 20
	}
	return defaultTitleWidth
}

// View renders the full TUI screen.
func (m AppModel) View() string {
	if m.InputMode {
		var sb strings.Builder
		if m.inputStep == stepTitle {
			sb.WriteString("New task: ")
			sb.WriteString(m.titleInput.View())
		} else {
			sb.WriteString("Due date (YYYY-MM-DD, Enter to skip): ")
			sb.WriteString(m.dateInput.View())
			if m.inputErr != "" {
				sb.WriteString("\n")
				sb.WriteString(styleError.Render(m.inputErr))
			}
		}
		sb.WriteString("\n\n")
		sb.WriteString(renderHelpBar(true))
		sb.WriteString("\n")
		return sb.String()
	}

	tw := m.titleColWidth()
	var sb strings.Builder
	sb.WriteString(renderTabBar(m.ActiveTab))
	sb.WriteString("\n\n")

	if len(m.Tasks) == 0 {
		sb.WriteString(renderEmptyState(m.ActiveTab))
		sb.WriteString("\n")
	} else {
		now := time.Now()
		start, end := m.viewportBounds()
		for i := start; i < end; i++ {
			sb.WriteString(renderTaskRow(m.Tasks[i], i == m.Cursor, now, tw))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	if m.ConfirmDelete && len(m.Tasks) > 0 {
		sb.WriteString(renderConfirmBar(m.Tasks[m.Cursor].Title))
	} else if m.errorMsg != "" {
		sb.WriteString(styleError.Render(m.errorMsg))
	} else {
		sb.WriteString(renderHelpBar(false))
	}
	// No trailing \n: view occupies Height-1 newlines when full, preventing the alt-screen
	// from scrolling the tab bar off the top of the terminal.
	return sb.String()
}
