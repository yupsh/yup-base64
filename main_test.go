package main

import (
	"context"
	"testing"

	clix "github.com/gloo-foo/cli"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// parse runs args through a bare command carrying the wrapper's flags and
// returns the parsed accessor, so flag-dependent helpers are tested against real
// parsed flags.
func parse(t *testing.T, args ...string) *urf.Command {
	t.Helper()
	var got *urf.Command
	app := &urf.Command{
		Name:   name,
		Flags:  flags(),
		Action: func(_ context.Context, c *urf.Command) error { got = c; return nil },
	}
	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return got
}

func TestOptions(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"none", []string{name}, 0},
		{"decode", []string{name, "-d"}, 1},
		{"ignore", []string{name, "-i"}, 1},
		{"wrap", []string{name, "-w", "80"}, 1},
		{"all", []string{name, "-d", "-i", "-w", "0"}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(options(parse(t, tc.args...))); got != tc.want {
				t.Fatalf("options len=%d, want %d", got, tc.want)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	src, filter, err := build(clix.Invocation{Args: parse(t, name), Fs: afero.NewMemMapFs()})
	if err != nil || src == nil || filter == nil {
		t.Fatalf("build: src=%v filter=%v err=%v", src, filter, err)
	}
}

func Test_main(t *testing.T) {
	orig := runMain
	t.Cleanup(func() { runMain = orig })
	var gotName clix.Name
	runMain = func(s clix.Spec, _ clix.Version) { gotName = s.Name }
	main()
	if gotName != name {
		t.Fatalf("main used spec %q, want %s", gotName, name)
	}
}

// TestFlagsAreFreshOnEveryCall names flags's claim: each call yields its own
// values, so parse state cannot leak from one run to the next.
//
// urfave/cli flags carry mutable state — once supplied, a flag's hasBeenSet
// stays true for the life of that value. Sharing one slice across parses made
// IsSet report flags that only an EARLIER parse had set, which is how the
// shared gate's -count=2 (two passes in one process) turned a passing suite
// red: the first pass set the flag, the second saw it already set.
func TestFlagsAreFreshOnEveryCall(t *testing.T) {
	first, second := flags(), flags()
	if len(first) != len(second) {
		t.Fatalf("flag count differs between calls: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] == second[i] {
			t.Fatalf("flag %d is the same value on both calls; parse state would leak between runs", i)
		}
	}
}
