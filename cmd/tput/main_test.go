package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xp-runners/evolution/internal/version"
)

func TestRun(t *testing.T) {
	for _, tt := range []struct {
		name       string
		args       []string
		input      string
		wantOut    string
		wantErr    string
		wantStatus int
	}{
		{name: "usage", wantErr: "usage: tput [-V] [-S] capname\n", wantStatus: 1},
		{name: "version", args: []string{"-V"}, wantOut: "xp-runners " + version.Number + "\n"},

		// Arguments
		{name: "lines", args: []string{"lines"}, wantOut: "28\n"},
		{name: "cols", args: []string{"cols"}, wantOut: "91\n"},
		{name: "bell", args: []string{"bel"}, wantOut: "\a\n"},
		{name: "unknown", args: []string{"clear"}, wantErr: "tput: unknown terminfo capability 'clear'\n", wantStatus: 1},

		// Batch
		{name: "empty batch", args: []string{"-S"}},
		{name: "batch", args: []string{"-S"}, input: " cols \r\n\tbel\nlines", wantOut: "91\n\a\n28\n"},
		{
			name:       "batch stops on error",
			args:       []string{"-S"},
			input:      "cols\nbad\nlines\n",
			wantOut:    "91\n",
			wantErr:    "tput: unknown terminfo capability 'bad'\n",
			wantStatus: 1,
		},
		{
			name:       "blank line in batch",
			args:       []string{"-S"},
			input:      "\n",
			wantErr:    "tput: unknown terminfo capability ''\n",
			wantStatus: 1,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out, stderr bytes.Buffer
			status := run(tt.args, strings.NewReader(tt.input), &out, &stderr, func() (int, int) { return 91, 28 })

			if status != tt.wantStatus || out.String() != tt.wantOut || stderr.String() != tt.wantErr {
				t.Fatalf("got (%d, %q, %q); want (%d, %q, %q)", status, &out, &stderr, tt.wantStatus, tt.wantOut, tt.wantErr)
			}
		})
	}
}
