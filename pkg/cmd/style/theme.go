package style

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// GetTheme returns a theme for the huh package with custom CA styles.
func GetTheme() *huh.Theme {
	t := huh.ThemeBase()
	// CNJR-11090 we swap the button colors for light mode to improve visibility
	if !HasDarkBackground() {
		t.Focused.FocusedButton, t.Focused.BlurredButton = t.Focused.BlurredButton, t.Focused.FocusedButton
	}
	// CNJR-11090 make focused button bold for better visibility
	t.Focused.FocusedButton = t.Focused.FocusedButton.Bold(true)
	return t
}

// HasDarkBackground checks if the current terminal has a dark background.
func HasDarkBackground() bool {
	return lipgloss.HasDarkBackground()
}
