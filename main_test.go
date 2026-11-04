package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunRejectsBadUsage(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "no command", args: []string{}},
		{name: "unknown command", args: []string{"nope"}},
		{name: "block with extra argument", args: []string{"block", "extra"}},
		{name: "balance without address", args: []string{"balance"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			if code := run(tc.args, &stdout, &stderr); code != 2 {
				t.Errorf("run(%q) = %d, want 2", tc.args, code)
			}
			if !strings.Contains(stderr.String(), "usage:") {
				t.Errorf("run(%q) stderr = %q, want it to contain %q", tc.args, stderr.String(), "usage:")
			}
		})
	}
}
