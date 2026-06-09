package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/justasandbox/my-todo-cli/internal/todo"
)

// modelWithHeight sets Height on a model pre-loaded with n tasks.
func modelWithHeight(n, h int) AppModel {
	m := modelWithTasks(n)
	m.Height = h
	return m
}

// TestViewport_OffsetFollowsCursorDown: navigating past the visible window advances offset.
func TestViewport_OffsetFollowsCursorDown(t *testing.T) {
	// Height=10, fixedRows=4 → visibleRows=6. With 10 tasks:
	// cursor 0..5 are visible at offset=0; cursor=6 must scroll offset to 1.
	m := modelWithHeight(10, 10)

	for i := 0; i < 5; i++ {
		m = sendKey(m, "j")
	}
	if m.Cursor != 5 {
		t.Fatalf("precondition: cursor=%d, want 5", m.Cursor)
	}
	if m.Offset != 0 {
		t.Errorf("offset should be 0 when cursor is within view, got %d", m.Offset)
	}

	m = sendKey(m, "j") // cursor → 6, beyond offset(0)+visibleRows(6)
	if m.Cursor != 6 {
		t.Fatalf("cursor after j: got %d, want 6", m.Cursor)
	}
	if m.Offset != 1 {
		t.Errorf("offset after scrolling cursor past viewport: got %d, want 1", m.Offset)
	}
}

// TestViewport_OffsetFollowsCursorUp: cursor above offset shrinks offset.
func TestViewport_OffsetFollowsCursorUp(t *testing.T) {
	m := modelWithHeight(10, 10)
	m.Cursor = 3
	m.Offset = 3

	m = sendKey(m, "k") // cursor → 2, above offset(3)
	if m.Cursor != 2 {
		t.Fatalf("cursor after k: got %d, want 2", m.Cursor)
	}
	if m.Offset != 2 {
		t.Errorf("offset after scrolling cursor above viewport: got %d, want 2", m.Offset)
	}
}

// TestViewport_WrapTopToBottom: wrapping cursor from top to bottom sets offset to show last row.
func TestViewport_WrapTopToBottom(t *testing.T) {
	// cursor=0, offset=0 → k → cursor=9 (wrap). visibleRows=6 → offset = 9-6+1 = 4.
	m := modelWithHeight(10, 10)
	m = sendKey(m, "k")
	if m.Cursor != 9 {
		t.Fatalf("cursor after wrap: got %d, want 9", m.Cursor)
	}
	if m.Offset != 4 {
		t.Errorf("offset after top-to-bottom wrap: got %d, want 4", m.Offset)
	}
}

// TestViewport_WrapBottomToTop: wrapping cursor from bottom to top resets offset to 0.
func TestViewport_WrapBottomToTop(t *testing.T) {
	// cursor=9, offset=4 → j → cursor=0 (wrap), offset must become 0.
	m := modelWithHeight(10, 10)
	m.Cursor = 9
	m.Offset = 4

	m = sendKey(m, "j")
	if m.Cursor != 0 {
		t.Fatalf("cursor after wrap: got %d, want 0", m.Cursor)
	}
	if m.Offset != 0 {
		t.Errorf("offset after bottom-to-top wrap: got %d, want 0", m.Offset)
	}
}

// TestViewport_ResizeShrink: shrinking terminal adjusts offset so cursor stays visible.
func TestViewport_ResizeShrink(t *testing.T) {
	// cursor=5, offset=0, Height=10 (visibleRows=6) → resize to Height=6 (visibleRows=2).
	// cursor=5 >= 0+2 → offset = 5-2+1 = 4.
	m := modelWithHeight(10, 10)
	m.Cursor = 5

	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 6})
	m = next.(AppModel)
	if m.Height != 6 {
		t.Fatalf("height not updated: got %d", m.Height)
	}
	if m.Offset != 4 {
		t.Errorf("offset after shrink: got %d, want 4", m.Offset)
	}
}

// TestViewport_ResizeGrow: growing terminal does not push offset out of bounds.
func TestViewport_ResizeGrow(t *testing.T) {
	// cursor=9, offset=6, Height=10 (visibleRows=6) → resize to Height=14 (visibleRows=10).
	// cursor=9 >= 6+10=16? No. cursor=9 < 6? No. → offset stays 6.
	m := modelWithHeight(10, 10)
	m.Cursor = 9
	m.Offset = 6

	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 14})
	m = next.(AppModel)
	if m.Offset != 6 {
		t.Errorf("offset after grow: got %d, want 6", m.Offset)
	}
}

// TestViewport_ZeroHeightRendersAll: when Height==0 (no WindowSizeMsg yet), all tasks are rendered.
func TestViewport_ZeroHeightRendersAll(t *testing.T) {
	m := modelWithHeight(10, 0)
	view := m.View()
	count := strings.Count(view, "task")
	if count != 10 {
		t.Errorf("Height=0 should render all 10 tasks, got %d occurrences of 'task'", count)
	}
}

// TestViewport_FewerTasksThanViewport: when task count < visibleRows, all tasks are shown.
func TestViewport_FewerTasksThanViewport(t *testing.T) {
	m := modelWithHeight(3, 10) // visibleRows=6, only 3 tasks
	view := m.View()
	count := strings.Count(view, "task")
	if count != 3 {
		t.Errorf("expected 3 task rows, got %d", count)
	}
}

// TestViewport_ViewCapsRenderedRows: View() only renders visibleRows rows from offset.
func TestViewport_ViewCapsRenderedRows(t *testing.T) {
	// 10 tasks, Height=10 (visibleRows=6), offset=0 → only 6 rendered.
	m := modelWithHeight(10, 10)
	view := m.View()
	count := strings.Count(view, "task")
	if count != 6 {
		t.Errorf("expected 6 visible task rows, got %d", count)
	}
}

// TestViewport_ViewShowsCorrectWindow: after scrolling, View() shows the right slice.
func TestViewport_ViewShowsCorrectWindow(t *testing.T) {
	// 10 tasks with distinct titles; navigate to cursor=7 so offset=2.
	// View should contain tasks 2..7, not task 0 or task 9.
	todos := make([]todo.Task, 10)
	for i := range todos {
		todos[i] = todo.Task{ID: i + 1, Title: fmt.Sprintf("task%d", i)}
	}
	repo := &testRepo{OnList: func(_ todo.Filter) ([]todo.Task, error) { return todos, nil }}
	m := New(repo)
	m.Tasks = todos
	m.Height = 10 // visibleRows=6

	// Navigate to cursor=7 → offset=2
	for i := 0; i < 7; i++ {
		m = sendKey(m, "j")
	}
	if m.Cursor != 7 || m.Offset != 2 {
		t.Fatalf("precondition failed: cursor=%d offset=%d, want cursor=7 offset=2", m.Cursor, m.Offset)
	}

	view := m.View()
	if strings.Contains(view, "task0") {
		t.Error("view should not contain task0 (scrolled out)")
	}
	if strings.Contains(view, "task1") {
		t.Error("view should not contain task1 (scrolled out)")
	}
	if !strings.Contains(view, "task2") {
		t.Error("view should contain task2 (first visible)")
	}
	if !strings.Contains(view, "task7") {
		t.Error("view should contain task7 (cursor)")
	}
	if strings.Contains(view, "task8") {
		t.Error("view should not contain task8 (scrolled out)")
	}
}

// TestViewport_TabSwitchResetsOffset: switching tabs resets offset to 0.
func TestViewport_TabSwitchResetsOffset(t *testing.T) {
	m := modelWithHeight(10, 10)
	m.Cursor = 7
	m.Offset = 4

	m = sendKey(m, "2") // switch to All tab, resets cursor+offset
	if m.Offset != 0 {
		t.Errorf("offset after tab switch: got %d, want 0", m.Offset)
	}
	if m.Cursor != 0 {
		t.Errorf("cursor after tab switch: got %d, want 0", m.Cursor)
	}
}

// TestViewport_ArrowKeysScrollViewport: up/down arrows obey the same viewport logic.
func TestViewport_ArrowKeysScrollViewport(t *testing.T) {
	m := modelWithHeight(10, 10) // visibleRows=6
	m.Cursor = 5

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(AppModel)
	if m.Cursor != 6 {
		t.Fatalf("down arrow cursor: got %d, want 6", m.Cursor)
	}
	if m.Offset != 1 {
		t.Errorf("offset after down past viewport: got %d, want 1", m.Offset)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = next.(AppModel)
	if m.Cursor != 5 {
		t.Fatalf("up arrow cursor: got %d, want 5", m.Cursor)
	}
	if m.Offset != 1 {
		t.Errorf("offset after up (still in view): got %d, want 1", m.Offset)
	}
}
