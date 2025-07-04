package style

import (
	"context"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/spf13/cobra"
)

func huhColorScheme(lightDark lipgloss.LightDarkFunc) fang.ColorScheme {
	base := lightDark(lipgloss.Black, lipgloss.White)
	t := GetTheme()
	return fang.ColorScheme{
		Base:         base,
		Description:  base,
		Comment:      lightDark(lipgloss.BrightWhite, lipgloss.BrightBlack),
		Argument:     base,
		Help:         base,
		Dash:         base,
		ErrorDetails: t.Focused.ErrorMessage.GetForeground(),
	}
}

func Execute(cmd *cobra.Command) error {
	return fang.Execute(
		context.Background(),
		cmd,
		fang.WithoutVersion(),
		fang.WithColorSchemeFunc(huhColorScheme),
	)
}
