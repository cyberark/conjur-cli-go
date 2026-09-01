package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// outputBoolAliases are persistent boolean flags that are sugar for
// --output <name> — e.g. --json is exactly --output json. They exist
// purely for convenience; --output remains the source of truth and these
// aliases are rejected if they'd conflict with an explicit --output value.
var outputBoolAliases = []string{"json", "yaml", "plain", "table", "interactive"}

// resolveOutputSpec merges --output with its --json/--yaml/--plain/--table/
// --interactive/--columns/--template sugar flags into the single raw spec
// output.Resolve expects, and reports whether the result came from an
// explicit user choice (as opposed to --output's implicit default).
//
// Each alias is only honored if it's actually the persistent flag in
// effect for cmd — a subcommand defining its own local flag of the same
// name (e.g. variable get's --interactive, which means something
// unrelated) silently shadows the persistent one by cobra/pflag design,
// and must keep its original, sole meaning rather than being reinterpreted
// as an output-format request.
func resolveOutputSpec(cmd *cobra.Command) (spec string, explicit bool, err error) {
	raw, err := cmd.Flags().GetString("output")
	if err != nil {
		return "", false, err
	}
	explicit = cmd.Flags().Changed("output")

	var aliasName, aliasSpec string
	setAlias := func(name, spec string) error {
		if aliasName != "" {
			return fmt.Errorf("cannot combine --%s with --%s", aliasName, name)
		}
		aliasName, aliasSpec = name, spec
		return nil
	}

	for _, name := range outputBoolAliases {
		if !isPersistentFlagInEffect(cmd, name) {
			continue
		}
		v, err := cmd.Flags().GetBool(name)
		if err != nil {
			return "", false, err
		}
		if v {
			if err := setAlias(name, name); err != nil {
				return "", false, err
			}
		}
	}

	if isPersistentFlagInEffect(cmd, "columns") {
		columns, err := cmd.Flags().GetString("columns")
		if err != nil {
			return "", false, err
		}
		if columns != "" {
			colSpec, err := columnsAliasSpec(columns)
			if err != nil {
				return "", false, err
			}
			if err := setAlias("columns", colSpec); err != nil {
				return "", false, err
			}
		}
	}

	if isPersistentFlagInEffect(cmd, "template") {
		tmpl, err := cmd.Flags().GetString("template")
		if err != nil {
			return "", false, err
		}
		if tmpl != "" {
			if err := setAlias("template", "go-template="+tmpl); err != nil {
				return "", false, err
			}
		}
	}

	if aliasName == "" {
		return raw, explicit, nil
	}
	if explicit && raw != aliasSpec {
		return "", false, fmt.Errorf("cannot combine --output %s with --%s", raw, aliasName)
	}
	return aliasSpec, true, nil
}

// isPersistentFlagInEffect reports whether the root's persistent flag
// named name is the one cmd will actually read — false if a local flag of
// the same name shadows it.
func isPersistentFlagInEffect(cmd *cobra.Command, name string) bool {
	persistent := cmd.Root().PersistentFlags().Lookup(name)
	return persistent != nil && cmd.Flags().Lookup(name) == persistent
}

// columnsAliasSpec turns a plain comma-separated field-name list (e.g.
// "id,kind") into a custom-columns= spec selecting each field by its
// top-level JSON key (e.g. "id:.id,kind:.kind"). For nested paths, use
// --output custom-columns=NAME:.path directly.
func columnsAliasSpec(columns string) (string, error) {
	names := strings.Split(columns, ",")
	pairs := make([]string, 0, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			return "", fmt.Errorf("--columns: empty column name in %q", columns)
		}
		pairs = append(pairs, fmt.Sprintf("%s:.%s", n, n))
	}
	return "custom-columns=" + strings.Join(pairs, ","), nil
}
