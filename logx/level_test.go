package logx

import (
	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    slog.Level
		wantErr bool
	}{
		{name: "debug", input: "debug", want: slog.LevelDebug},
		{name: "info", input: "INFO", want: slog.LevelInfo},
		{name: "warn", input: "Warn", want: slog.LevelWarn},
		{name: "warning", input: "warning", want: slog.LevelWarn},
		{name: "error", input: "ERROR", want: slog.LevelError},
		{name: "trim-space", input: "  info  ", want: slog.LevelInfo},
		{name: "invalid", input: "trace", want: slog.LevelInfo, wantErr: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseLogLevel(tc.input)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for input %q", tc.input)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("unexpected level for input %q: got=%v want=%v", tc.input, got, tc.want)
			}
		})
	}
}
