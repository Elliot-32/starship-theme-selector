package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Elliot-32/starship-theme-selector/internal/theme"
)

func TestVersionCommand(t *testing.T) {
	root := newRoot("1.2.3", &theme.Manager{})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != "ssts version 1.2.3" {
		t.Fatalf("version output = %q", got)
	}
}
