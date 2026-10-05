package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Elliot-32/starship-theme-selector/internal/theme"
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

func TestPreviewPanelKeepsFixedOuterSizeAcrossPreviewStates(t *testing.T) {
	for _, width := range []int{72, 140} {
		model := New(nil, []theme.Theme{{Name: "demo", Source: theme.SourceBuiltin}})
		model.width = width
		model.height = 24
		model.resize()

		loadingPanel := model.previewPanel()
		loadingWidth, loadingHeight := lipgloss.Size(loadingPanel)
		if loadingWidth != model.previewPanelWidth() {
			t.Fatalf("width %d: loading panel width = %d, want %d", width, loadingWidth, model.previewPanelWidth())
		}
		if loadingHeight != previewFrameHeight {
			t.Fatalf("width %d: loading panel height = %d, want %d", width, loadingHeight, previewFrameHeight)
		}

		loadingViewHeight := lipgloss.Height(model.viewString())
		model.previewFor = "demo"
		model.preview = "\n\x1b[31mline one\x1b[0m\nline two\nline three\nline four"

		loadedPanel := model.previewPanel()
		loadedWidth, loadedHeight := lipgloss.Size(loadedPanel)
		if loadedWidth != loadingWidth || loadedHeight != loadingHeight {
			t.Fatalf("width %d: panel size changed from %dx%d to %dx%d", width, loadingWidth, loadingHeight, loadedWidth, loadedHeight)
		}
		if strings.Contains(loadedPanel, "line three") || strings.Contains(loadedPanel, "line four") {
			t.Fatalf("width %d: preview viewport leaked lines beyond %d rows", width, previewPromptLines)
		}
		if loadedViewHeight := lipgloss.Height(model.viewString()); loadedViewHeight != loadingViewHeight {
			t.Fatalf("width %d: view height changed from %d to %d", width, loadingViewHeight, loadedViewHeight)
		}

		view := model.View()
		if !view.AltScreen {
			t.Fatal("TUI view must use the alternate screen")
		}
	}
}

func TestClampPreviewNeverExceedsSafeWidth(t *testing.T) {
	const width = 24
	preview := "\x1b[38;5;212m  elliot \x1b[0m\ue0b0 \x1b[38;5;220m\uf126 main\x1b[0m \ue0b0 \x1b[32m\ue73c v3.14.7\x1b[0m \ue0b0 19:23 \ue0b0"

	got := clampPreview(preview, width, previewPromptLines)
	lines := strings.Split(got, "\n")
	if len(lines) > previewPromptLines {
		t.Fatalf("preview has %d lines, want at most %d", len(lines), previewPromptLines)
	}
	for i, line := range lines {
		if gotWidth := ansi.StringWidth(line); gotWidth > width {
			t.Fatalf("line %d width = %d, want <= %d: %q", i, gotWidth, width, line)
		}
	}
}

func TestViewReservesRightmostTerminalColumn(t *testing.T) {
	for _, width := range []int{72, 90, 140} {
		model := New(nil, []theme.Theme{{Name: "catppuccin-powerline", Source: theme.SourceBuiltin}})
		model.width = width
		model.height = 24
		model.resize()
		model.previewFor = "catppuccin-powerline"
		model.preview = "\x1b[38;5;212m  elliot \x1b[0m\ue0b0 \x1b[38;5;220m\uf126 main\x1b[0m \ue0b0 \x1b[32m\ue73c v3.14.7\x1b[0m \ue0b0 19:23 \ue0b0"

		gotWidth := lipgloss.Width(model.viewString())
		wantMax := width - layoutRightMargin
		if gotWidth > wantMax {
			t.Fatalf("terminal width %d: view width = %d, want <= %d", width, gotWidth, wantMax)
		}
	}
}
