package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// 6.1 – empty list guards (navigation, toggle, delete already tested in update_test.go)
// These tests verify the full empty-list no-op contract.

func TestEmptyList_SpaceNoOp(t *testing.T) {
	m := modelWithTasks(0)
	m2 := sendKey(m, " ")
	if m2.Cursor != 0 {
		t.Errorf("space on empty list: cursor got %d, want 0", m2.Cursor)
	}
}

func TestEmptyList_DeleteNoOp(t *testing.T) {
	m := modelWithTasks(0)
	m2 := sendKey(m, "d")
	if m2.Cursor != 0 {
		t.Errorf("d on empty list: cursor got %d, want 0", m2.Cursor)
	}
}

// 6.2 – cursor persistence after mutations
// These tests drive the todosLoadedMsg handler directly because loadTodos is async.

func TestCursorPersistence_LoadPreservesPosition(t *testing.T) {
	m := modelWithTasks(3)
	m.Cursor = 1
	next, _ := m.Update(todosLoadedMsg{
		todos:  []Todo{{ID: 1}, {ID: 2}, {ID: 3}},
		cursor: 1,
	})
	am := next.(AppModel)
	if am.Cursor != 1 {
		t.Errorf("cursor should stay at 1, got %d", am.Cursor)
	}
}

func TestCursorPersistence_ClampedWhenListShrinks(t *testing.T) {
	m := modelWithTasks(3)
	m.Cursor = 2
	// List shrinks to 1 item (e.g. after toggle removes task from view).
	next, _ := m.Update(todosLoadedMsg{
		todos:  []Todo{{ID: 1}},
		cursor: 2,
	})
	am := next.(AppModel)
	if am.Cursor != 0 {
		t.Errorf("cursor should clamp to 0, got %d", am.Cursor)
	}
}

func TestCursorPersistence_EmptyListAfterMutation(t *testing.T) {
	m := modelWithTasks(1)
	m.Cursor = 0
	next, _ := m.Update(todosLoadedMsg{
		todos:  []Todo{},
		cursor: 0,
	})
	am := next.(AppModel)
	if am.Cursor != 0 {
		t.Errorf("cursor on empty list should be 0, got %d", am.Cursor)
	}
}

func TestCursorPersistence_TabSwitchResetsCursorToZero(t *testing.T) {
	m := modelWithTasks(3)
	m.Cursor = 2
	m = sendKey(m, "2") // explicitly resets cursor to 0 before async load
	if m.Cursor != 0 {
		t.Errorf("tab switch should reset cursor to 0, got %d", m.Cursor)
	}
}

// 6.3 – terminal resize handling

func TestWindowResize_StoresDimensions(t *testing.T) {
	m := New(&testRepo{})
	next, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	am := next.(AppModel)
	if am.Width != 120 || am.Height != 40 {
		t.Errorf("expected 120x40, got %dx%d", am.Width, am.Height)
	}
	if cmd != nil {
		t.Error("resize should return no cmd")
	}
}

func TestWindowResize_UpdatedDimensionsReflected(t *testing.T) {
	m := New(&testRepo{})
	next1, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	am1 := next1.(AppModel)
	next2, _ := am1.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	am2 := next2.(AppModel)
	if am2.Width != 200 || am2.Height != 50 {
		t.Errorf("expected 200x50 after second resize, got %dx%d", am2.Width, am2.Height)
	}
}

func TestTitleColWidth_DefaultWhenNoSize(t *testing.T) {
	m := New(&testRepo{})
	if got := m.titleColWidth(); got != defaultTitleWidth {
		t.Errorf("expected default %d, got %d", defaultTitleWidth, got)
	}
}

func TestTitleColWidth_DynamicWhenWidthKnown(t *testing.T) {
	m := New(&testRepo{})
	m.Width = 100
	got := m.titleColWidth()
	if got <= defaultTitleWidth {
		t.Errorf("100-wide terminal should give title col > %d, got %d", defaultTitleWidth, got)
	}
}

func TestTitleColWidth_MinimumEnforced(t *testing.T) {
	m := New(&testRepo{})
	m.Width = 10 // very small
	got := m.titleColWidth()
	if got < 20 {
		t.Errorf("title col should never be below 20, got %d", got)
	}
}
