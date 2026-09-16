// Command yup-base64 is the CLI wrapper around github.com/gloo-foo/cmd-base64.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-base64"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name              = "base64"
	flagDecode        = "decode"
	flagIgnoreGarbage = "ignore-garbage"
	flagWrap          = "wrap"
)

// synopsis is the multi-line --help usage block. urfave/cli indents the whole
// block three spaces, so the lines stay flush-left.
const synopsis = `base64 [OPTION]... [FILE]

Base64 encode or decode FILE, or standard input, to standard output.
With no FILE, or when FILE is -, read standard input.`

// spec declares the base64 wrapper: a file-or-stdin filter with encode/decode,
// garbage-tolerance, and line-wrapping flags.
var spec = clix.Spec{
	Name:     name,
	Summary:  "base64 encode/decode data and print on the standard output",
	Synopsis: synopsis,
	Build:    build,
	Flags:    flags(),
}

// build maps the invocation to base64's pipeline: a file-or-stdin source into
// the base64 command configured by the parsed flags.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	return clix.OperandsOrStdin(inv), command.Base64(options(inv.Args)...), nil
}

// options folds the parsed flags into base64's option values.
func options(c *urf.Command) []any {
	var opts []any
	if c.Bool(flagDecode) {
		opts = append(opts, command.Base64Decode)
	}
	if c.Bool(flagIgnoreGarbage) {
		opts = append(opts, command.Base64IgnoreGarbage)
	}
	if c.IsSet(flagWrap) {
		opts = append(opts, command.Base64Wrap(c.Int(flagWrap)))
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }

// flags builds a FRESH flag set on every call.
//
// urfave/cli flags carry mutable parse state: once a flag is supplied, its
// hasBeenSet stays true for the life of the value. A package-level slice shared
// across two parses therefore reports the second parse as having flags only the
// FIRST one set, and IsSet stops meaning "the user asked for this".
//
// The shared gate runs every suite with -count=2 in one process, which is
// exactly that shape — the first pass sets the flag, the second sees it already
// set. Building fresh keeps one definition while giving each parse its own
// state.
func flags() []urf.Flag {
	return []urf.Flag{
		&urf.BoolFlag{
			Name:    flagDecode,
			Aliases: []string{"d"},
			Usage:   "decode data",
			Sources: urf.EnvVars("YUP_BASE64_DECODE"),
		},
		&urf.BoolFlag{
			Name:    flagIgnoreGarbage,
			Aliases: []string{"i"},
			Usage:   "when decoding, ignore non-alphabet characters",
			Sources: urf.EnvVars("YUP_BASE64_IGNORE_GARBAGE"),
		},
		&urf.IntFlag{
			Name:    flagWrap,
			Aliases: []string{"w"},
			Value:   76,
			Usage:   "wrap encoded lines after COLS character; use 0 to disable wrapping",
			Sources: urf.EnvVars("YUP_BASE64_WRAP"),
		},
	}
}
