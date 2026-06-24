package main

import (
	"context"
	"fmt"
	"io"

	command "github.com/gloo-foo/cmd-base64"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const (
	flagDecode        = "decode"
	flagIgnoreGarbage = "ignore-garbage"
	flagWrap          = "wrap"
)

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `base64 [OPTION]... [FILE]

Base64 encode or decode FILE, or standard input, to standard output.
With no FILE, or when FILE is -, read standard input.`

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags while still exposing
// the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// run builds and executes the base64 CLI against the injected version, I/O, and
// filesystem, returning the process exit code.
func run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer, fs afero.Fs) int {
	cmd := newCommand(version, stdin, stdout, fs)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, "base64: %v\n", err)
		return 1
	}
	return 0
}

func newCommand(version string, stdin io.Reader, stdout io.Writer, fs afero.Fs) *cli.Command {
	return &cli.Command{
		Name:            "base64",
		Version:         version,
		Usage:           "base64 encode/decode data and print on the standard output",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: flagDecode, Aliases: []string{"d"}, Usage: "decode data"},
			&cli.BoolFlag{Name: flagIgnoreGarbage, Aliases: []string{"i"}, Usage: "when decoding, ignore non-alphabet characters"},
			&cli.IntFlag{Name: flagWrap, Aliases: []string{"w"}, Value: 76, Usage: "wrap encoded lines after COLS character; use 0 to disable wrapping"},
		},
		Action: action(stdin, stdout, fs),
	}
}

func action(stdin io.Reader, stdout io.Writer, fs afero.Fs) cli.ActionFunc {
	return func(_ context.Context, cmd *cli.Command) error {
		_, err := gloo.Run(source(cmd, stdin, fs), gloo.ByteWriteTo(stdout), command.Base64(options(cmd)...))
		return err
	}
}

func source(cmd *cli.Command, stdin io.Reader, fs afero.Fs) any {
	if cmd.NArg() == 0 {
		return gloo.ByteReaderSource([]io.Reader{stdin})
	}
	files := make([]gloo.File, cmd.NArg())
	for i := range files {
		files[i] = gloo.File(cmd.Args().Get(i))
	}
	return gloo.ByteFileSource(fs, files)
}

func options(cmd *cli.Command) []any {
	var opts []any
	if cmd.Bool(flagDecode) {
		opts = append(opts, command.Base64Decode)
	}
	if cmd.Bool(flagIgnoreGarbage) {
		opts = append(opts, command.Base64IgnoreGarbage)
	}
	if cmd.IsSet(flagWrap) {
		opts = append(opts, command.Base64Wrap(cmd.Int(flagWrap)))
	}
	return opts
}
