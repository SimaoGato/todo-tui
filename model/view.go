package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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

// 5.2 + 5.3 – task row with color coding

func renderTaskRow(todo Todo, isCursor bool, now time.Time, titleWidth int) string {
	cursor := "  "
	if isCursor {
		cursor = "> "
	}
	check := "[ ]"
	if todo.Done {
		check = "[x]"
	}
	dueStr := "—"
	if todo.DueDate != nil {
		dueStr = todo.DueDate.Format("2006-01-02")
	}
	title := fmt.Sprintf("%-*s", titleWidth, truncate(todo.Title, titleWidth))
	line := cursor + check + " " + title + "  " + dueStr

	if todo.Done {
		return styleDone.Render(line)
	}
	if todo.DueDate != nil {
		due := todo.DueDate.In(now.Location()).Truncate(24 * time.Hour)
		today := now.Truncate(24 * time.Hour)
		if due.Before(today) {
			return styleOverdue.Render(line)
		}
		if due.Equal(today) {
			return styleDueToday.Render(line)
		}
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
		for i, task := range m.Tasks {
			sb.WriteString(renderTaskRow(task, i == m.Cursor, now, tw))
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
	sb.WriteString("\n")
	return sb.String()
}
