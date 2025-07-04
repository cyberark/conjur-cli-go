package style

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// GetTheme returns a theme for the huh package with custom CA styles.
func GetTheme() *huh.Theme {
	t := huh.ThemeBase()
	return t
}

// HasDarkBackground checks if the current terminal has a dark background.
func HasDarkBackground() bool {
	return lipgloss.HasDarkBackground()
}
