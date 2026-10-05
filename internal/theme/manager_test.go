package theme

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeReplacesSymbolsTable(t *testing.T) {
	themeData := []byte(`
[os]
disabled = false

[os.symbols]
Arch = "old"
OnlyInTheme = "must-disappear"

[directory]
style = "blue"
`)
	override := []byte(`
[os.symbols]
Arch = "new"
CachyOS = "cachy"
`)

	merged, err := Merge(themeData, override)
	if err != nil {
		t.Fatal(err)
	}
	text := string(merged)
	for _, want := range []string{`Arch = 'new'`, `CachyOS = 'cachy'`, `disabled = false`, `style = 'blue'`} {
		if !strings.Contains(text, want) {
			t.Fatalf("merged config missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "OnlyInTheme") || strings.Contains(text, "old") {
		t.Fatalf("theme symbols leaked through full replacement:\n%s", text)
	}
}

func TestDiscoverCustomOverridesBuiltinWithSameName(t *testing.T) {
	dir := t.TempDir()
	starship := filepath.Join(dir, "starship")
	script := `#!/bin/sh
if [ "$1" = preset ] && [ "$2" = --list ]; then
  printf 'alpha\nbeta\n'
  exit 0
fi
exit 1
`
	if err := os.WriteFile(starship, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	themeDir := filepath.Join(dir, "themes")
	if err := os.MkdirAll(themeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	customPath := filepath.Join(themeDir, "alpha.toml")
	if err := os.WriteFile(customPath, []byte("format = '$directory'\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := &Manager{Starship: starship, Paths: Paths{ThemeDir: themeDir}}
	themes, err := m.Discover(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(themes) != 2 {
		t.Fatalf("got %d themes, want 2", len(themes))
	}
	if themes[0].Name != "alpha" || themes[0].Source != SourceCustom || themes[0].Path != customPath {
		t.Fatalf("alpha = %#v", themes[0])
	}
}

func TestPreviewDoesNotLeakShellPromptEscapes(t *testing.T) {
	dir := t.TempDir()
	starship := filepath.Join(dir, "starship")
	script := `#!/bin/sh
if [ "$1" = preset ] && [ "$2" = demo ]; then
  printf 'format = "$character"\n'
  exit 0
fi
if [ "$1" = prompt ]; then
  if [ -n "$STARSHIP_SHELL" ]; then
    printf '%%{wrapped%%}'
  else
    printf '\033[31mraw-preview\033[0m'
  fi
  exit 0
fi
exit 1
`
	if err := os.WriteFile(starship, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("STARSHIP_SHELL", "zsh")
	t.Setenv("STARSHIP_CONFIG", "/should/not/leak")

	m := &Manager{
		Starship: starship,
		Paths: Paths{
			SymbolsFile: filepath.Join(dir, "missing-symbols.toml"),
		},
		DefaultSymbols: []byte("[os.symbols]\nCachyOS = 'cachy'\n"),
		WorkDir:        dir,
	}

	preview, err := m.Preview(t.Context(), Theme{Name: "demo", Source: SourceBuiltin}, 80)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(preview, "%{") || strings.Contains(preview, "%}") || strings.Contains(preview, "wrapped") {
		t.Fatalf("preview leaked shell prompt escapes: %q", preview)
	}
	if !strings.Contains(preview, "\x1b[31mraw-preview\x1b[0m") {
		t.Fatalf("preview lost ANSI rendering: %q", preview)
	}
}

func TestCreatePreviewProjectProvidesGoAndGitContext(t *testing.T) {
	dir, cleanup, err := CreatePreviewProject(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if filepath.Base(dir) != "stheme-preview" {
		t.Fatalf("preview dir base = %q, want stheme-preview", filepath.Base(dir))
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		t.Fatalf("preview go.mod: %v", err)
	}

	if git, err := exec.LookPath("git"); err == nil {
		cmd := exec.Command(git, "-C", dir, "branch", "--show-current")
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("read preview git branch: %v", err)
		}
		if strings.TrimSpace(string(out)) != "main" {
			t.Fatalf("preview branch = %q, want main", strings.TrimSpace(string(out)))
		}
	}
}

func TestPreviewRunsStarshipInsidePreviewProject(t *testing.T) {
	dir := t.TempDir()
	previewDir := filepath.Join(dir, "preview")
	if err := os.Mkdir(previewDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(previewDir, "go.mod"), []byte("module example.com/preview\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	starship := filepath.Join(dir, "starship")
	script := `#!/bin/sh
if [ "$1" = preset ]; then
  printf 'format = "$character"\n'
  exit 0
fi
if [ "$1" = prompt ]; then
  test -f go.mod || exit 8
  printf 'cwd=%s' "$PWD"
  exit 0
fi
exit 1
`
	if err := os.WriteFile(starship, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &Manager{
		Starship:       starship,
		Paths:          Paths{SymbolsFile: filepath.Join(dir, "missing.toml")},
		DefaultSymbols: []byte("[os.symbols]\nCachyOS = 'cachy'\n"),
		WorkDir:        dir,
		PreviewDir:     previewDir,
	}
	preview, err := m.Preview(t.Context(), Theme{Name: "demo", Source: SourceBuiltin}, 80)
	if err != nil {
		t.Fatal(err)
	}
	if preview != "cwd="+previewDir {
		t.Fatalf("preview = %q, want cwd=%s", preview, previewDir)
	}
}
