package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const titleWidth = 40

var (
	styleTabActive   = lipgloss.NewStyle().Bold(true).Underline(true)
	styleTabInactive = lipgloss.NewStyle().Faint(true)
	styleBold        = lipgloss.NewStyle().Bold(true)
	styleOverdue     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleDueToday    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleDone        = lipgloss.NewStyle().Faint(true).Strikethrough(true)
	styleHelp        = lipgloss.NewStyle().Faint(true)
	styleError       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
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

func renderTaskRow(todo Todo, isCursor bool, now time.Time) string {
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

	var sb strings.Builder
	sb.WriteString(renderTabBar(m.ActiveTab))
	sb.WriteString("\n\n")

	if len(m.Tasks) == 0 {
		sb.WriteString(renderEmptyState(m.ActiveTab))
		sb.WriteString("\n")
	} else {
		now := time.Now()
		for i, task := range m.Tasks {
			sb.WriteString(renderTaskRow(task, i == m.Cursor, now))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(renderHelpBar(false))
	sb.WriteString("\n")
	return sb.String()
}
