package style

import (
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// GetTheme returns a theme for the huh package with custom CA styles.
func GetTheme() huh.ThemeFunc {
	return func(isDark bool) *huh.Styles {
		t := huh.ThemeBase(isDark)
		// CNJR-11090 we swap the button colors for light mode to improve visibility
		if !isDark {
			t.Focused.FocusedButton, t.Focused.BlurredButton = t.Focused.BlurredButton, t.Focused.FocusedButton
		}
		// CNJR-11090 make focused button bold for better visibility
		t.Focused.FocusedButton = t.Focused.FocusedButton.Bold(true)
		return t
	}
}

// HasDarkBackground checks if the current terminal has a dark background.
func HasDarkBackground() bool {
	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
}
