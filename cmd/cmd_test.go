package cmd

import (
	"strings"
	"testing"
)

func TestValidateRuntimeOptionsRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		opts RuntimeOptions
		want string
	}{
		{
			name: "thread must be positive",
			opts: RuntimeOptions{Thread: 0, Timeout: 5, Max: 1, Mode: 1, Fuzz: 0},
			want: "thread",
		},
		{
			name: "timeout must be positive",
			opts: RuntimeOptions{Thread: 1, Timeout: 0, Max: 1, Mode: 1, Fuzz: 0},
			want: "timeout",
		},
		{
			name: "mode must be known",
			opts: RuntimeOptions{Thread: 1, Timeout: 5, Max: 1, Mode: 4, Fuzz: 0},
			want: "mode",
		},
		{
			name: "fuzz must be known",
			opts: RuntimeOptions{Thread: 1, Timeout: 5, Max: 1, Mode: 1, Fuzz: 4},
			want: "fuzz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRuntimeOptions(tt.opts)
			if err == nil {
				t.Fatal("ValidateRuntimeOptions() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ValidateRuntimeOptions() error = %q, want it to mention %q", err, tt.want)
			}
		})
	}
}

func TestValidateRuntimeOptionsAcceptsValidValues(t *testing.T) {
	opts := RuntimeOptions{Thread: 1, Timeout: 5, Max: 100, Mode: 3, Fuzz: 3}
	if err := ValidateRuntimeOptions(opts); err != nil {
		t.Fatalf("ValidateRuntimeOptions() error = %v, want nil", err)
	}
}

func TestHelpRequested(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "short help", args: []string{"-h"}, want: true},
		{name: "long help", args: []string{"--help"}, want: true},
		{name: "flag value is not help", args: []string{"-u", "https://example.com/help"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HelpRequested(tt.args); got != tt.want {
				t.Fatalf("HelpRequested(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
