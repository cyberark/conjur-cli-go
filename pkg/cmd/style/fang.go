package style

import (
	"context"
	"errors"
	"io"
	"strings"

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

// errorHandler wraps fang's DefaultErrorHandler, stripping a trailing dot from
// the error message so that fang's own appended dot doesn't produce "..".
func errorHandler(w io.Writer, styles fang.Styles, err error) {
	msg := strings.TrimRight(err.Error(), ".")
	if msg != err.Error() {
		err = errors.New(msg)
	}
	fang.DefaultErrorHandler(w, styles, err)
}

func Execute(cmd *cobra.Command) error {
	return fang.Execute(
		context.Background(),
		cmd,
		fang.WithoutVersion(),
		fang.WithColorSchemeFunc(huhColorScheme),
		fang.WithErrorHandler(errorHandler),
	)
}
