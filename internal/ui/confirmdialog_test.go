package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfirmDialog_DefaultIsInactive(t *testing.T) {
	cd := ConfirmDialog{}
	if cd.Active {
		t.Error("default ConfirmDialog should not be Active")
	}
}

func TestConfirmDialog_YSignalsYesAndDeactivates(t *testing.T) {
	cd := ConfirmDialog{Active: true, prompt: "task title"}
	updated, signal := cd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if signal != confirmSignalYes {
		t.Errorf("y signal: got %d, want confirmSignalYes", signal)
	}
	if updated.Active {
		t.Error("y should deactivate ConfirmDialog")
	}
}

func TestConfirmDialog_NSignalsNoAndDeactivates(t *testing.T) {
	cd := ConfirmDialog{Active: true, prompt: "task title"}
	updated, signal := cd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if signal != confirmSignalNo {
		t.Errorf("n signal: got %d, want confirmSignalNo", signal)
	}
	if updated.Active {
		t.Error("n should deactivate ConfirmDialog")
	}
}

func TestConfirmDialog_EscSignalsNoAndDeactivates(t *testing.T) {
	cd := ConfirmDialog{Active: true, prompt: "task title"}
	updated, signal := cd.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if signal != confirmSignalNo {
		t.Errorf("Esc signal: got %d, want confirmSignalNo", signal)
	}
	if updated.Active {
		t.Error("Esc should deactivate ConfirmDialog")
	}
}

func TestConfirmDialog_OtherKeysNoSignal(t *testing.T) {
	cd := ConfirmDialog{Active: true, prompt: "task title"}
	updated, signal := cd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if signal != confirmSignalNone {
		t.Errorf("irrelevant key signal: got %d, want confirmSignalNone", signal)
	}
	if !updated.Active {
		t.Error("irrelevant key should not deactivate")
	}
}

func TestConfirmDialog_ViewContainsPrompt(t *testing.T) {
	cd := ConfirmDialog{Active: true, prompt: "my important task"}
	v := cd.View()
	if !strings.Contains(v, "my important task") {
		t.Error("View should contain the prompt text")
	}
}
