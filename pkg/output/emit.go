package output

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// terminalSizeFunc is a test-overridable seam. It checks real os.Stdout
// rather than cmd.OutOrStdout(), which is a bytes.Buffer in tests and cannot
// report a size. Tests override it to exercise fitColumns' width logic
// headlessly.
var terminalSizeFunc = func() (int, int, error) {
	return term.GetSize(int(os.Stdout.Fd()))
}

// isRealTerminal is a test-overridable seam, mirroring
// prompts.PromptingAvailable's check. Resolve decides isInteractive purely
// from the --interactive/--output interactive flag, with no TTY check —
// Emit verifies one itself before launching the alt-screen bubbletea
// program, since bubbletea opening /dev/tty hard-crashes in any headless
// context (CI, a pipe, ssh without a pty) rather than failing gracefully.
var isRealTerminal = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

type formatContextKey struct{}

type formatContextValue struct {
	format      Format
	interactive bool
}

type explicitContextKey struct{}

// SetExplicit records whether the user actually typed --output, as opposed
// to it defaulting to plain because the flag was omitted. PlainOrUnix
// call sites use this to distinguish "no flag" (keep pre-existing legacy
// plain rendering, e.g. pretty JSON) from "--output plain" explicitly
// (render via PlainUnix instead). Commands that never call SetExplicit
// (e.g. most unit tests) get the safe default of false.
func SetExplicit(cmd *cobra.Command, explicit bool) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, explicitContextKey{}, explicit))
}

// IsExplicit reads back the value SetExplicit stored, defaulting to false
// if it was never set.
func IsExplicit(cmd *cobra.Command) bool {
	ctx := cmd.Context()
	if ctx == nil {
		return false
	}
	explicit, _ := ctx.Value(explicitContextKey{}).(bool)
	return explicit
}

// Resolve turns a raw --output flag value (possibly empty) into a concrete
// Format and whether the caller should use the interactive view instead of
// Render. Interactive mode is opt-in only, via --output interactive: an
// empty spec defaults to plain, and any other explicit value is passed
// straight to ParseFormat. Format.Type is FormatPlain either way; whether
// the flag was actually passed (as opposed to defaulting) is tracked
// separately via SetExplicit/IsExplicit, since an omitted --output renders
// differently from an explicit --output plain (see PlainOrUnix).
func Resolve(raw string) (Format, bool, error) {
	if raw == "interactive" {
		return Format{}, true, nil
	}
	if raw == "" {
		raw = "plain"
	}
	f, err := ParseFormat(raw)
	return f, false, err
}

// SetContext stores the resolved format/interactive decision on cmd's
// context so Emit (and any command needing early access, e.g. policy
// fetch's allowlist check) can read it back via FromContext.
func SetContext(cmd *cobra.Command, format Format, isInteractive bool) {
	// A cobra.Command constructed directly (as unit tests commonly do,
	// rather than via Execute()) has a nil Context() until cobra sets one;
	// context.WithValue panics on a nil parent, so fall back explicitly.
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, formatContextKey{}, formatContextValue{format, isInteractive})
	cmd.SetContext(ctx)
}

// FromContext reads back the value SetContext stored. It errors if called
// before the root PersistentPreRunE has run (a programmer error, not a
// runtime condition users can trigger), and also guards the same nil-context
// case SetContext does, for the same reason.
func FromContext(cmd *cobra.Command) (Format, bool, error) {
	ctx := cmd.Context()
	if ctx == nil {
		return Format{}, false, fmt.Errorf("output format not resolved on context; root PersistentPreRunE must run first")
	}
	v, ok := ctx.Value(formatContextKey{}).(formatContextValue)
	if !ok {
		return Format{}, false, fmt.Errorf("output format not resolved on context; root PersistentPreRunE must run first")
	}
	return v.format, v.interactive, nil
}

// RequireFormat rejects a resolved format that isn't in allowed, for
// commands whose data shape doesn't support the full renderer set (e.g.
// policy fetch rejecting custom-columns/go-template).
func RequireFormat(format Format, allowed ...FormatType) error {
	for _, a := range allowed {
		if format.Type == a {
			return nil
		}
	}
	names := make([]string, len(allowed))
	for i, a := range allowed {
		names[i] = string(a)
	}
	return fmt.Errorf("format %q is not supported for this command (supported formats: %s)", format.Type, strings.Join(names, ", "))
}

// Emit is the single call every migrated command's RunE uses to produce its
// output: it dispatches to the interactive view or Render based on what
// Resolve decided.
func Emit(cmd *cobra.Command, data any) error {
	format, isInteractive, err := FromContext(cmd)
	if err != nil {
		return err
	}

	if !isInteractive {
		return Render(cmd.OutOrStdout(), format, data)
	}

	rows, isSingleObject, err := normalizeForInteractive(data)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No resources found.")
		return nil
	}

	if !isRealTerminal() {
		return Render(cmd.OutOrStdout(), Format{Type: FormatJSON}, data)
	}

	var model tea.Model
	if isSingleObject {
		dm := newDetailModel(rows[0], true)
		model = dm
	} else {
		model = newTableModel(rows)
	}

	p := tea.NewProgram(model)
	_, err = p.Run()
	return err
}

// PlainJSON pretty-prints data as 2-space-indented JSON to cmd's stdout —
// the plainFallback body shared by every Plain/PlainOrUnix call site whose
// pre-existing plain output was already pretty JSON (as opposed to one
// item per line, or a bare scalar; see PlainBare for the latter).
//
// HTML escaping is disabled to match utils.PrettyPrintToJSON's pre-existing
// behavior (CNJR-4574): the default encoder would otherwise mangle values
// containing '<', '>', or '&'.
func PlainJSON(cmd *cobra.Command, data any) error {
	var buf strings.Builder
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return err
	}
	cmd.Print(buf.String())
	return nil
}

// PlainBare prints v as-is via cmd.Println — the plainFallback body shared
// by call sites whose pre-existing plain output is a single unwrapped
// value (e.g. an API key, or exists' bare true/false), even though the
// data passed to Emit for --output json/yaml is a wrapping map so those
// formats can label the value.
func PlainBare(cmd *cobra.Command, v any) error {
	cmd.Println(v)
	return nil
}

// PlainOneLine prints each item on its own line via cmd.Println — the
// plainFallback body shared by call sites whose pre-existing plain output
// is one raw value per line rather than pretty JSON (see PlainJSON).
func PlainOneLine(cmd *cobra.Command, items []string) error {
	for _, item := range items {
		cmd.Println(item)
	}
	return nil
}

// emitPlain is the shared core of Plain and PlainOrUnix: dispatch to Emit
// unless the resolved format is non-interactive plain, in which case use
// plainFallback — except when unixOnExplicit is true and --output plain
// was passed explicitly (as opposed to defaulting to plain because
// --output was omitted), where it uses PlainUnix instead. Plain and
// PlainOrUnix exist only to name that one true/false at the call site
// instead of leaving callers to infer it from which of several
// similarly-named functions they picked.
func emitPlain(cmd *cobra.Command, data any, plainFallback func() error, unixOnExplicit bool) error {
	format, isInteractive, err := FromContext(cmd)
	if err != nil {
		return err
	}
	if isInteractive || format.Type != FormatPlain {
		return Emit(cmd, data)
	}
	if unixOnExplicit && IsExplicit(cmd) {
		return PlainUnix(cmd, data)
	}
	return plainFallback()
}

// Plain is Emit, except in the default non-interactive plain case — flag
// omitted or explicit --output plain, treated identically — it calls
// plainFallback instead, for commands whose plain-mode rendering predates
// the output package (e.g. printing a raw value or one item per line) and
// isn't just Emit's generic JSON/YAML/table rendering of data.
func Plain(cmd *cobra.Command, data any, plainFallback func() error) error {
	return emitPlain(cmd, data, plainFallback, false)
}

// PlainOrUnix is Plain, except explicit --output plain renders via
// PlainUnix instead of plainFallback: omitting --output preserves
// plainFallback's pre-existing legacy output (e.g. pretty JSON) byte-for-
// byte, while explicit --output plain produces unix-style textual output
// (bare values, one-per-line lists, key=value objects) — useful for
// commands whose pre-migration default plain output isn't already
// unix-style, so the two need to diverge. (Plain's plainFallback already
// is unix-style-equivalent at every current call site, so it never needs
// this split.)
//
// One further wrinkle plainFallback exists to dodge: PlainUnix's rule of
// "print data as JSON.Marshal would" is unsuitable as a *default* body for
// data that started life as a typed struct's response bytes rather than a
// literal Go value — e.g. policy load's dry-run response, whose field
// order (status, created, updated, deleted, errors) callers depend on.
// json.Marshal always sorts map keys alphabetically once data has been
// round-tripped through a generic map[string]any, but the original
// response bytes (marshaled server-side from a typed struct) preserve
// that order; such callers pass a plainFallback that prints those original
// bytes instead of re-marshaling data.
func PlainOrUnix(cmd *cobra.Command, data any, plainFallback func() error) error {
	return emitPlain(cmd, data, plainFallback, true)
}

// PlainJSONOrUnix is PlainOrUnix for the common case where the plain
// fallback is just pretty-printing the same data Emit would otherwise
// json/yaml/table-render — i.e. PlainOrUnix(cmd, data, func() error {
// return PlainJSON(cmd, data) }), spelled as one call.
func PlainJSONOrUnix(cmd *cobra.Command, data any) error {
	return PlainOrUnix(cmd, data, func() error {
		return PlainJSON(cmd, data)
	})
}

// normalizeForInteractive round-trips data through JSON like normalizeRows
// (columns.go) does, but also reports whether the *unmarshaled* top-level
// value was a JSON object rather than an array, so Emit can open the detail
// view directly instead of a one-row table. It inspects the round-tripped
// shape rather than data's original Go type — a typed slice such as
// []conjurapi.HostFactoryTokenResponse or []string is not literally []any,
// so switching on data itself would misclassify every such collection as a
// single object.
func normalizeForInteractive(data any) (rows []any, isSingleObject bool, err error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, false, fmt.Errorf("interactive: marshaling data: %w", err)
	}
	var unmarshaled any
	if err := json.Unmarshal(raw, &unmarshaled); err != nil {
		return nil, false, fmt.Errorf("interactive: unmarshaling data: %w", err)
	}

	switch v := unmarshaled.(type) {
	case nil:
		return nil, false, nil
	case []any:
		return wrapScalarRows(v), false, nil
	case map[string]any:
		return []any{v}, true, nil
	default:
		return nil, false, fmt.Errorf("interactive view requires a struct, map, or slice, got %T", unmarshaled)
	}
}

// wrapScalarRows synthesizes a single "value" column for any row that isn't
// already a map[string]any, so scalar slices (e.g. the default []string
// result of `conjur list`) render as a one-column table instead of an empty,
// column-less one. deriveColumns/cellValue only understand map[string]any
// rows; without this, every scalar-slice command's default TTY output would
// be a blank box.
func wrapScalarRows(rows []any) []any {
	wrapped := make([]any, len(rows))
	for i, row := range rows {
		if m, ok := row.(map[string]any); ok {
			wrapped[i] = m
			continue
		}
		wrapped[i] = map[string]any{"value": row}
	}
	return wrapped
}
