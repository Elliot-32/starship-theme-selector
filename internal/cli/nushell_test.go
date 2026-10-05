package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestNushellCompletion(t *testing.T) {
	var out bytes.Buffer
	if err := genNushellCompletion(&out); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		`extern "ssts"`,
		`string@"nu-complete ssts themes"`,
		`^ssts list`,
		`extern "ssts completion"`,
		`[bash zsh fish powershell nushell]`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Nushell completion missing %q", want)
		}
	}
}
