package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// adjustOutputShorthand decides, once per process and before cobra parses
// anything, whether --output may claim the -o shorthand for this specific
// invocation. cobra/pflag has no notion of "this shorthand applies to
// every command except one": merging persistent and local flags into one
// FlagSet per command is unconditional, and pflag panics the instant two
// flags in that merged set claim the same shorthand byte (see AddFlag in
// pflag's flag.go). list's local -o/--offset flag makes a global
// -o/--output impossible in general.
//
// What IS possible: each CLI invocation is a fresh process running
// exactly one command, so we can predict which command is about to run by
// walking args ourselves — using only flag sets that already exist
// pre-merge (each visited command's own local flags, plus every ancestor's
// PersistentFlags up to that point) — and decide the shorthand just for
// this run, before Execute() ever triggers cobra's real, merging flag
// resolution.
//
// This deliberately does not call cobra's own Command.Find/stripFlags:
// those merge persistent flags into whatever command they visit as a side
// effect (mergePersistentFlags), which would itself panic on list before
// we ever got a chance to decide anything.
func adjustOutputShorthand(root *cobra.Command, args []string) {
	outputFlag := root.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		return
	}

	all := collectCommands(root)
	target := predictTargetCommand(root, args, all)
	if target.Flags().ShorthandLookup("o") != nil {
		outputFlag.Shorthand = ""
		return
	}
	outputFlag.Shorthand = "o"
}

// predictTargetCommand is a deliberately narrow, non-merging stand-in for
// Command.Find. Any input it can't confidently interpret just stops the
// walk where it is; predictTargetCommand always returns a valid command
// (root, or wherever it got to), so adjustOutputShorthand's decision is
// always based on a real command in the tree — worst case it's the wrong
// one, and -o silently isn't available for that run, which is the
// existing, always-safe status quo.
func predictTargetCommand(cur *cobra.Command, args []string, all []*cobra.Command) *cobra.Command {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			break
		}
		if isFlagToken(a) {
			if flagConsumesNextArg(a, all) && i+1 < len(args) {
				i++
			}
			continue
		}
		next := findSubCommand(cur, a)
		if next == nil {
			break
		}
		cur = next
	}
	return cur
}

func isFlagToken(a string) bool {
	return len(a) > 1 && a[0] == '-'
}

// flagConsumesNextArg reports whether token a (a flag, without an inline
// "=value") needs the following argv token as its value. It searches
// every command's own local flags across the whole tree (all already
// fully registered — via each command's own init(), long before
// Execute() runs — so this never needs cobra's lazy, merging accessors)
// rather than only flags visible from cur's current position in the
// walk: a flag local to a not-yet-reached descendant (e.g. list's
// --offset, encountered before "list" itself appears in args) is
// otherwise invisible from an ancestor's vantage point, which would
// misjudge it as taking no value and misinterpret its actual value as
// the next subcommand name instead.
func flagConsumesNextArg(a string, all []*cobra.Command) bool {
	if strings.Contains(a, "=") {
		return false
	}

	var flag *pflag.Flag
	if strings.HasPrefix(a, "--") {
		name := strings.TrimPrefix(a, "--")
		if name == "output" {
			// --output's own value-taking-ness can't be looked up the
			// normal way here: --output is exactly the flag whose
			// shorthand we're trying to resolve, and it may not be
			// registered as anything yet at this point. It always takes
			// a value regardless, so short-circuit.
			return true
		}
		flag = lookupFlagByName(all, name)
	} else {
		// Shorthand clusters (e.g. -di) never need a trailing value for
		// our purposes: this codebase never defines a clustered boolean
		// shorthand followed by a value-taking one in the same cluster.
		short := strings.TrimPrefix(a, "-")
		if len(short) != 1 {
			return false
		}
		if short == "o" {
			// Same reasoning as --output above: "-o" means either
			// --offset (int) or --output (string) depending on the
			// eventual target, and both take a value, so its
			// value-taking-ness doesn't depend on resolving that
			// ambiguity — treat it as value-consuming unconditionally.
			return true
		}
		flag = lookupFlagByShorthand(all, short)
	}
	if flag == nil {
		// Unknown to us: assume no value, matching cobra's own handling
		// of an unrecognized flag (it errors rather than consuming a
		// following token speculatively).
		return false
	}
	return flag.NoOptDefVal == ""
}

// collectCommands returns root and every command reachable from it.
func collectCommands(root *cobra.Command) []*cobra.Command {
	all := []*cobra.Command{root}
	for _, c := range root.Commands() {
		all = append(all, collectCommands(c)...)
	}
	return all
}

// lookupFlagByName searches every command's own local flags plus root's
// (or any command's) PersistentFlags for name, returning the first match.
// Ambiguous only if the same name existed twice in the tree with
// different value-taking-ness, which this codebase doesn't do.
func lookupFlagByName(all []*cobra.Command, name string) *pflag.Flag {
	for _, c := range all {
		if f := c.Flags().Lookup(name); f != nil {
			return f
		}
		if f := c.PersistentFlags().Lookup(name); f != nil {
			return f
		}
	}
	return nil
}

func lookupFlagByShorthand(all []*cobra.Command, shorthand string) *pflag.Flag {
	for _, c := range all {
		if f := c.Flags().ShorthandLookup(shorthand); f != nil {
			return f
		}
		if f := c.PersistentFlags().ShorthandLookup(shorthand); f != nil {
			return f
		}
	}
	return nil
}

func findSubCommand(cur *cobra.Command, name string) *cobra.Command {
	for _, c := range cur.Commands() {
		if c.Name() == name || c.HasAlias(name) {
			return c
		}
	}
	return nil
}
