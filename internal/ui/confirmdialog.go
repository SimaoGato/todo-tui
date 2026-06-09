package ui

import tea "github.com/charmbracelet/bubbletea"

type confirmSignal int

const (
	confirmSignalNone confirmSignal = iota
	confirmSignalYes
	confirmSignalNo
)

// ConfirmDialog manages the delete-confirmation prompt.
type ConfirmDialog struct {
	prompt string
	Active bool
}

// Update processes a key event for the confirm dialog.
// Returns the updated ConfirmDialog and a signal; the caller handles the yes path.
func (cd ConfirmDialog) Update(msg tea.KeyMsg) (ConfirmDialog, confirmSignal) {
	if msg.String() == "y" {
		cd.Active = false
		return cd, confirmSignalYes
	}
	if msg.String() == "n" || msg.Type == tea.KeyEsc {
		cd.Active = false
		return cd, confirmSignalNo
	}
	return cd, confirmSignalNone
}

// View renders the confirm prompt bar.
func (cd ConfirmDialog) View() string {
	return renderConfirmBar(cd.prompt)
}
