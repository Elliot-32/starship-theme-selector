package main

import "testing"

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name          string
		injected      string
		moduleVersion string
		want          string
	}{
		{name: "release ldflags win", injected: "v1.2.3", moduleVersion: "v9.9.9", want: "1.2.3"},
		{name: "go install module version", injected: "dev", moduleVersion: "v1.2.3", want: "1.2.3"},
		{name: "local build", injected: "dev", moduleVersion: "(devel)", want: "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeVersion(tt.injected, tt.moduleVersion); got != tt.want {
				t.Fatalf("normalizeVersion(%q, %q) = %q, want %q", tt.injected, tt.moduleVersion, got, tt.want)
			}
		})
	}
}
