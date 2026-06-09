package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTaskInput_NewIsInactive(t *testing.T) {
	ti := newTaskInput()
	if ti.Active {
		t.Error("new TaskInput should not be Active")
	}
	if ti.step != stepNone {
		t.Errorf("new TaskInput: step=%d, want stepNone", ti.step)
	}
}

func TestTaskInput_ActivateSetsTitleStep(t *testing.T) {
	ti := newTaskInput()
	ti, _ = ti.activate()
	if !ti.Active {
		t.Error("activate should set Active=true")
	}
	if ti.step != stepTitle {
		t.Errorf("activate: step=%d, want stepTitle", ti.step)
	}
}

func TestTaskInput_EscDeactivatesAndSignalsCancel(t *testing.T) {
	ti := newTaskInput()
	ti, _ = ti.activate()
	ti, _, signal, _ := ti.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if ti.Active {
		t.Error("Esc should set Active=false")
	}
	if ti.step != stepNone {
		t.Errorf("Esc: step=%d, want stepNone", ti.step)
	}
	if signal != taskInputSignalCancel {
		t.Errorf("Esc signal: got %d, want taskInputSignalCancel", signal)
	}
}

func TestTaskInput_EmptyTitleDoesNotAdvance(t *testing.T) {
	ti := newTaskInput()
	ti, _ = ti.activate()
	ti, _, signal, _ := ti.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ti.step != stepTitle {
		t.Error("empty title Enter should not advance to date step")
	}
	if signal != taskInputSignalNone {
		t.Errorf("empty title signal: got %d, want taskInputSignalNone", signal)
	}
}

func TestTaskInput_TitleEnterAdvancesToDateStep(t *testing.T) {
	ti := newTaskInput()
	ti, _ = ti.activate()
	for _, r := range "Buy milk" {
		ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ti.step != stepDate {
		t.Errorf("after title Enter: step=%d, want stepDate", ti.step)
	}
	if ti.pending != "Buy milk" {
		t.Errorf("pending: got %q, want %q", ti.pending, "Buy milk")
	}
}

func TestTaskInput_InvalidDateSetsErr(t *testing.T) {
	ti := newTaskInput()
	ti, _ = ti.activate()
	for _, r := range "Task" {
		ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyEnter}) // title → date
	for _, r := range "not-a-date" {
		ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	ti, _, signal, _ := ti.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ti.err == "" {
		t.Error("invalid date should set err")
	}
	if signal != taskInputSignalNone {
		t.Errorf("invalid date signal: got %d, want taskInputSignalNone", signal)
	}
	if !ti.Active {
		t.Error("should remain Active after invalid date")
	}
}

func TestTaskInput_SkipDateSignalsSubmit(t *testing.T) {
	ti := newTaskInput()
	ti, _ = ti.activate()
	for _, r := range "Buy milk" {
		ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyEnter}) // title → date
	_, _, signal, result := ti.Update(tea.KeyMsg{Type: tea.KeyEnter}) // skip date
	if signal != taskInputSignalSubmit {
		t.Errorf("skip-date signal: got %d, want taskInputSignalSubmit", signal)
	}
	if result == nil {
		t.Fatal("result should not be nil on submit")
	}
	if result.Title != "Buy milk" {
		t.Errorf("result.Title: got %q, want %q", result.Title, "Buy milk")
	}
	if result.DueDate != nil {
		t.Error("result.DueDate should be nil when date skipped")
	}
}

func TestTaskInput_ValidDateSignalsSubmitWithDate(t *testing.T) {
	ti := newTaskInput()
	ti, _ = ti.activate()
	for _, r := range "Task" {
		ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyEnter})
	for _, r := range "2026-12-25" {
		ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, _, signal, result := ti.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if signal != taskInputSignalSubmit {
		t.Errorf("valid-date signal: got %d, want taskInputSignalSubmit", signal)
	}
	if result == nil || result.DueDate == nil {
		t.Fatal("expected non-nil result with due date")
	}
	if got := result.DueDate.Format("2006-01-02"); got != "2026-12-25" {
		t.Errorf("result.DueDate: got %q, want %q", got, "2026-12-25")
	}
}

func TestTaskInput_TAutocompleteInDateStep(t *testing.T) {
	ti := newTaskInput()
	ti, _ = ti.activate()
	for _, r := range "Task" {
		ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyEnter}) // → stepDate
	ti, _, _, _ = ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	if got, want := ti.dateInput.Value(), time.Now().Format("2006-01-02"); got != want {
		t.Errorf("t autocomplete: got %q, want %q", got, want)
	}
}
