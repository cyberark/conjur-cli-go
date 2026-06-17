package style

import (
	"context"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

func huhColorScheme(lightDark lipgloss.LightDarkFunc) fang.ColorScheme {
	base := lightDark(lipgloss.Black, lipgloss.White)
	t := GetTheme()(base == lipgloss.Black)
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
