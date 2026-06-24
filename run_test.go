package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name       string
		version    string
		args       []string
		stdin      string
		files      map[string]string
		wantOut    string
		wantCode   int
		wantErrSub string
	}{
		{
			name:    "encode stdin",
			args:    []string{"base64"},
			stdin:   "hello\n",
			wantOut: "aGVsbG8=\n",
		},
		{
			name:    "decode stdin",
			args:    []string{"base64", "-d"},
			stdin:   "aGVsbG8=\n",
			wantOut: "hello\n",
		},
		{
			// -i discards non-alphabet bytes before decoding, so the spaces
			// embedded in the encoded text do not abort the decode.
			name:    "decode ignores garbage",
			args:    []string{"base64", "-d", "-i"},
			stdin:   "aGV s bG8=\n",
			wantOut: "hello\n",
		},
		{
			// -w 4 wraps the 8-character encoding into two lines of four.
			name:    "wrap at four columns",
			args:    []string{"base64", "-w", "4"},
			stdin:   "hello\n",
			wantOut: "aGVs\nbG8=\n",
		},
		{
			// -w 0 disables wrapping, emitting one unbroken line regardless of
			// the input length.
			name:    "wrap zero disables wrapping",
			args:    []string{"base64", "-w", "0"},
			stdin:   strings.Repeat("a", 60),
			wantOut: "YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFh\n",
		},
		{
			// The file source yields the file's bytes without its trailing
			// newline, so "one\ntwo" base64-encodes to b25lCnR3bw==.
			name:    "encode file source",
			args:    []string{"base64", "/in.txt"},
			files:   map[string]string{"/in.txt": "one\ntwo\n"},
			wantOut: "b25lCnR3bw==\n",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"base64", "--version"},
			wantOut: "base64 version 1.2.3\n",
		},
		{
			name:       "decode rejects garbage without ignore flag",
			args:       []string{"base64", "-d"},
			stdin:      "not valid base64 !!!\n",
			wantCode:   1,
			wantErrSub: "base64:",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"base64", "--nope"},
			wantCode:   1,
			wantErrSub: "base64:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for path, content := range tc.files {
				if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
					t.Fatalf("write fixture %s: %v", path, err)
				}
			}

			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(tc.stdin), &out, &errOut, fs)

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
