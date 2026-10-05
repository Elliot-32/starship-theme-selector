package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Elliot-32/stheme/internal/theme"
)

func TestInitRendersSelectedThemePreview(t *testing.T) {
	dir := t.TempDir()
	starship := filepath.Join(dir, "starship")
	script := `#!/bin/sh
case "$1" in
  prompt)
    grep -q 'CachyOS' "$STARSHIP_CONFIG" || exit 9
    printf '\033[32mPREVIEW-MARKER\033[0m'
    ;;
  *) exit 1 ;;
esac
`
	if err := os.WriteFile(starship, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	custom := filepath.Join(dir, "custom.toml")
	if err := os.WriteFile(custom, []byte("format = 'preview'\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := &theme.Manager{
		Starship: starship,
		Paths: theme.Paths{
			ThemeDir:    dir,
			SymbolsFile: filepath.Join(dir, "does-not-exist.toml"),
			Target:      filepath.Join(dir, "starship.toml"),
		},
		DefaultSymbols: []byte("[os.symbols]\nCachyOS = '\uf385'\n"),
		WorkDir:        dir,
	}
	model := New(m, []theme.Theme{{Name: "custom", Source: theme.SourceCustom, Path: custom}})
	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Init returned nil preview command")
	}
	msg, ok := cmd().(previewMsg)
	if !ok {
		t.Fatalf("preview command returned %T", cmd())
	}
	if msg.err != nil {
		t.Fatal(msg.err)
	}
	if msg.name != "custom" || !strings.Contains(msg.text, "PREVIEW-MARKER") {
		t.Fatalf("preview = %#v", msg)
	}
}
