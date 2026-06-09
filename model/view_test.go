package model

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/justasandbox/my-todo-cli/todo"
)

// 5.1 – tab bar

func TestRenderTabBar_ContainsAllLabels(t *testing.T) {
	for _, tab := range []Tab{TabToday, TabAll, TabCompleted} {
		bar := renderTabBar(tab)
		for _, label := range []string{"Today", "All", "Completed"} {
			if !strings.Contains(bar, label) {
				t.Errorf("renderTabBar(%d): missing label %q", tab, label)
			}
		}
	}
}

func TestRenderTabBar_ActiveLabelPresent(t *testing.T) {
	cases := []struct {
		tab   Tab
		label string
	}{
		{TabToday, "Today"},
		{TabAll, "All"},
		{TabCompleted, "Completed"},
	}
	for _, c := range cases {
		bar := renderTabBar(c.tab)
		if !strings.Contains(bar, c.label) {
			t.Errorf("renderTabBar(%d): active label %q not in output", c.tab, c.label)
		}
	}
}

// 5.2 – task row structure

func TestRenderTaskRow_CursorIndicator(t *testing.T) {
	task := todo.Todo{ID: 1, Title: "task"}
	now := time.Now()
	with := renderTaskRow(task, true, now, defaultTitleWidth)
	without := renderTaskRow(task, false, now, defaultTitleWidth)
	if !strings.Contains(with, ">") {
		t.Error("cursor row should contain '>'")
	}
	if strings.Contains(without, ">") {
		t.Error("non-cursor row should not contain '>'")
	}
}

func TestRenderTaskRow_CheckboxPending(t *testing.T) {
	task := todo.Todo{ID: 1, Title: "pending"}
	row := renderTaskRow(task, false, time.Now(), defaultTitleWidth)
	if !strings.Contains(row, "[ ]") {
		t.Error("pending task should show '[ ]'")
	}
}

func TestRenderTaskRow_CheckboxDone(t *testing.T) {
	task := todo.Todo{ID: 1, Title: "done task", Done: true}
	row := renderTaskRow(task, false, time.Now(), defaultTitleWidth)
	if !strings.Contains(row, "[x]") {
		t.Error("done task should show '[x]'")
	}
}

func TestRenderTaskRow_NoDueDateShowsDash(t *testing.T) {
	task := todo.Todo{ID: 1, Title: "no date"}
	row := renderTaskRow(task, false, time.Now(), defaultTitleWidth)
	if !strings.Contains(row, "—") {
		t.Error("task without due date should show '—'")
	}
}

func TestRenderTaskRow_DueDateFormatted(t *testing.T) {
	due := time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC)
	task := todo.Todo{ID: 1, Title: "xmas", DueDate: &due}
	row := renderTaskRow(task, false, time.Now(), defaultTitleWidth)
	if !strings.Contains(row, "2026-12-25") {
		t.Error("task with due date should show formatted date")
	}
}

func TestRenderTaskRow_TitleTruncated(t *testing.T) {
	long := strings.Repeat("a", 50)
	task := todo.Todo{ID: 1, Title: long}
	row := renderTaskRow(task, false, time.Now(), defaultTitleWidth)
	if strings.Contains(row, long) {
		t.Error("long title should be truncated")
	}
	if !strings.Contains(row, "...") {
		t.Error("truncated title should end with '...'")
	}
}

// 5.3 – color coding (structural: verify no panic, correct content still present)

func TestRenderTaskRow_OverdueNoPanic(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	task := todo.Todo{ID: 1, Title: "overdue", DueDate: &yesterday}
	row := renderTaskRow(task, false, time.Now(), defaultTitleWidth)
	if !strings.Contains(row, "overdue") {
		t.Error("overdue row should contain task title")
	}
}

func TestRenderTaskRow_DueTodayNoPanic(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	task := todo.Todo{ID: 1, Title: "due today", DueDate: &today}
	row := renderTaskRow(task, false, now, defaultTitleWidth)
	if !strings.Contains(row, "due today") {
		t.Error("due-today row should contain task title")
	}
}

func TestRenderTaskRow_FutureDateNoPanic(t *testing.T) {
	future := time.Now().AddDate(0, 0, 7)
	task := todo.Todo{ID: 1, Title: "future", DueDate: &future}
	row := renderTaskRow(task, false, time.Now(), defaultTitleWidth)
	if !strings.Contains(row, "future") {
		t.Error("future row should contain task title")
	}
}

// 5.4 – help bar

func TestRenderHelpBar_NormalMode(t *testing.T) {
	h := renderHelpBar(false)
	for _, want := range []string{"j/k", "a: add", "d: delete", "q: quit"} {
		if !strings.Contains(h, want) {
			t.Errorf("normal help bar missing %q", want)
		}
	}
}

func TestRenderHelpBar_InputMode(t *testing.T) {
	h := renderHelpBar(true)
	for _, want := range []string{"enter: confirm", "esc: cancel"} {
		if !strings.Contains(h, want) {
			t.Errorf("input help bar missing %q", want)
		}
	}
}

// 5.5 – empty state

func TestRenderEmptyState_Messages(t *testing.T) {
	cases := []struct {
		tab  Tab
		want string
	}{
		{TabToday, "No tasks for today"},
		{TabAll, "No tasks"},
		{TabCompleted, "No completed tasks"},
	}
	for _, c := range cases {
		got := renderEmptyState(c.tab)
		if !strings.Contains(got, c.want) {
			t.Errorf("empty state tab %d: want %q, got %q", c.tab, c.want, got)
		}
	}
}

// full View() integration

func TestView_NormalModeContainsTabsAndHelp(t *testing.T) {
	m := New(&testRepo{})
	v := m.View()
	for _, want := range []string{"Today", "All", "Completed", "a: add", "q: quit"} {
		if !strings.Contains(v, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestView_EmptyStateShownWhenNoTasks(t *testing.T) {
	m := New(&testRepo{})
	v := m.View()
	if !strings.Contains(v, "No tasks for today") {
		t.Error("empty Today view should show 'No tasks for today'")
	}
}

func TestView_TaskTitleShown(t *testing.T) {
	m := New(&testRepo{})
	m.Tasks = []todo.Todo{{ID: 1, Title: "my important task"}}
	v := m.View()
	if !strings.Contains(v, "my important task") {
		t.Error("view should show task title")
	}
}

func TestView_InputMode_TitlePrompt(t *testing.T) {
	m := New(&testRepo{})
	m = sendKey(m, "a")
	v := m.View()
	if !strings.Contains(v, "New task:") {
		t.Error("input mode title step should show 'New task:'")
	}
	if !strings.Contains(v, "esc: cancel") {
		t.Error("input mode should show input help bar")
	}
}

func TestView_InputMode_DatePrompt(t *testing.T) {
	m := New(&testRepo{})
	m = sendKey(m, "a")
	m = typeString(m, "Buy milk")
	m = sendKeyType(m, tea.KeyEnter)
	v := m.View()
	if !strings.Contains(v, "Due date") {
		t.Error("date step should show due date prompt")
	}
}

// Timezone-aware date classification (TT-78)

func TestTaskDateClass_DueTodayInNegativeOffsetZone(t *testing.T) {
	loc := time.FixedZone("UTC-4", -4*60*60)
	// 10pm June 9 in UTC-4 = 2am June 10 UTC — "today" is still June 9 locally.
	now := time.Date(2026, 6, 9, 22, 0, 0, 0, loc)
	// Midnight June 9 in UTC-4 = 4am June 9 UTC.
	// With Truncate(24h): UTC June 9 < UTC June 10 → "overdue" (bug).
	// With Year/Month/Day: both June 9 locally → "today" (correct).
	due := time.Date(2026, 6, 9, 0, 0, 0, 0, loc)
	if got := taskDateClass(&due, now); got != "today" {
		t.Errorf("taskDateClass = %q, want %q", got, "today")
	}
}

func TestTaskDateClass_OverdueInNegativeOffsetZone(t *testing.T) {
	loc := time.FixedZone("UTC-4", -4*60*60)
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, loc)
	due := time.Date(2026, 6, 8, 0, 0, 0, 0, loc)
	if got := taskDateClass(&due, now); got != "overdue" {
		t.Errorf("taskDateClass = %q, want %q", got, "overdue")
	}
}

func TestTaskDateClass_FutureNotClassified(t *testing.T) {
	loc := time.FixedZone("UTC-4", -4*60*60)
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, loc)
	due := time.Date(2026, 6, 10, 0, 0, 0, 0, loc)
	if got := taskDateClass(&due, now); got != "" {
		t.Errorf("taskDateClass = %q, want %q", got, "")
	}
}

func TestTaskDateClass_NilDueDate(t *testing.T) {
	if got := taskDateClass(nil, time.Now()); got != "" {
		t.Errorf("taskDateClass(nil) = %q, want %q", got, "")
	}
}
