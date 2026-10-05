package theme

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Source string

const (
	SourceBuiltin Source = "built-in"
	SourceCustom  Source = "custom"
)

type Theme struct {
	Name   string
	Source Source
	Path   string
}

type Paths struct {
	ConfigDir   string
	ThemeDir    string
	SymbolsFile string
	Target      string
}

type Manager struct {
	Starship       string
	Paths          Paths
	DefaultSymbols []byte
	WorkDir        string
	PreviewDir     string
}

func DefaultPaths() (Paths, error) {
	configHome, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve user config dir: %w", err)
	}

	configDir := os.Getenv("STHEME_CONFIG_DIR")
	if configDir == "" {
		configDir = filepath.Join(configHome, "stheme")
	}

	themeDir := os.Getenv("STHEME_THEME_DIR")
	if themeDir == "" {
		themeDir = filepath.Join(configDir, "themes")
	}

	symbolsFile := os.Getenv("STHEME_SYMBOLS")
	if symbolsFile == "" {
		symbolsFile = filepath.Join(configDir, "os-symbols.toml")
	}

	target := os.Getenv("STARSHIP_CONFIG")
	if target == "" {
		target = filepath.Join(configHome, "starship.toml")
	}

	return Paths{
		ConfigDir:   configDir,
		ThemeDir:    themeDir,
		SymbolsFile: symbolsFile,
		Target:      target,
	}, nil
}

func New(starship string, paths Paths, defaultSymbols []byte) (*Manager, error) {
	if starship == "" {
		var err error
		starship, err = exec.LookPath("starship")
		if err != nil {
			return nil, errors.New("starship was not found in PATH")
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get current directory: %w", err)
	}
	return &Manager{
		Starship:       starship,
		Paths:          paths,
		DefaultSymbols: append([]byte(nil), defaultSymbols...),
		WorkDir:        cwd,
	}, nil
}

func (m *Manager) Discover(ctx context.Context) ([]Theme, error) {
	byName := map[string]Theme{}

	cmd := exec.CommandContext(ctx, m.Starship, "preset", "--list")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list Starship presets: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		byName[name] = Theme{Name: name, Source: SourceBuiltin}
	}

	entries, err := os.ReadDir(m.Paths.ThemeDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read custom theme directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".toml" {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		byName[name] = Theme{
			Name:   name,
			Source: SourceCustom,
			Path:   filepath.Join(m.Paths.ThemeDir, entry.Name()),
		}
	}

	themes := make([]Theme, 0, len(byName))
	for _, item := range byName {
		themes = append(themes, item)
	}
	sort.Slice(themes, func(i, j int) bool {
		return strings.ToLower(themes[i].Name) < strings.ToLower(themes[j].Name)
	})
	return themes, nil
}

func (m *Manager) Find(ctx context.Context, name string) (Theme, error) {
	themes, err := m.Discover(ctx)
	if err != nil {
		return Theme{}, err
	}
	for _, item := range themes {
		if item.Name == name {
			return item, nil
		}
	}
	return Theme{}, fmt.Errorf("theme %q not found", name)
}

func (m *Manager) themeBytes(ctx context.Context, item Theme) ([]byte, error) {
	if item.Source == SourceCustom {
		data, err := os.ReadFile(item.Path)
		if err != nil {
			return nil, fmt.Errorf("read custom theme %q: %w", item.Name, err)
		}
		return data, nil
	}

	cmd := exec.CommandContext(ctx, m.Starship, "preset", item.Name)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("render Starship preset %q: %w", item.Name, err)
	}
	return out, nil
}

func (m *Manager) symbolsBytes() ([]byte, error) {
	data, err := os.ReadFile(m.Paths.SymbolsFile)
	if err == nil {
		return data, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read os.symbols override: %w", err)
	}
	return append([]byte(nil), m.DefaultSymbols...), nil
}

func Merge(themeData, symbolsData []byte) ([]byte, error) {
	var doc map[string]any
	if err := toml.Unmarshal(themeData, &doc); err != nil {
		return nil, fmt.Errorf("parse theme TOML: %w", err)
	}
	if doc == nil {
		doc = map[string]any{}
	}

	var override struct {
		OS struct {
			Symbols map[string]any `toml:"symbols"`
		} `toml:"os"`
	}
	if err := toml.Unmarshal(symbolsData, &override); err != nil {
		return nil, fmt.Errorf("parse os.symbols override: %w", err)
	}
	if len(override.OS.Symbols) == 0 {
		return nil, errors.New("os.symbols override is empty")
	}

	osTable, ok := doc["os"].(map[string]any)
	if !ok || osTable == nil {
		osTable = map[string]any{}
	}
	// Intentional full replacement: keys from the theme's original [os.symbols]
	// must not survive unless they are present in the override table.
	osTable["symbols"] = override.OS.Symbols
	doc["os"] = osTable

	out, err := toml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("encode merged Starship config: %w", err)
	}
	return out, nil
}

func (m *Manager) Build(ctx context.Context, item Theme) ([]byte, error) {
	themeData, err := m.themeBytes(ctx, item)
	if err != nil {
		return nil, err
	}
	symbolsData, err := m.symbolsBytes()
	if err != nil {
		return nil, err
	}
	return Merge(themeData, symbolsData)
}

func previewEnvironment(configPath string) []string {
	env := make([]string, 0, len(os.Environ())+1)
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "STARSHIP_SHELL=") || strings.HasPrefix(entry, "STARSHIP_CONFIG=") {
			continue
		}
		env = append(env, entry)
	}
	return append(env, "STARSHIP_CONFIG="+configPath)
}

func CreatePreviewProject(ctx context.Context) (string, func(), error) {
	root, err := os.MkdirTemp("", "stheme-preview-*")
	if err != nil {
		return "", nil, fmt.Errorf("create preview project: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(root) }
	dir := filepath.Join(root, "stheme-preview")
	if err := os.Mkdir(dir, 0o755); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("create preview project directory: %w", err)
	}
	projectFiles := map[string][]byte{
		"pyproject.toml": []byte("[project]\nname = 'stheme-preview'\nversion = '0.1.0'\nrequires-python = '>=3.11'\n"),
		"main.py":        []byte("print('stheme preview')\n"),
	}
	for name, data := range projectFiles {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("write preview %s: %w", name, err)
		}
	}

	// Git is optional for stheme itself, but when available we make the preview
	// a real repository so Starship renders git_branch/git_status exactly as it
	// would in a project. The committed Python project files keep the preview
	// status clean.
	if git, err := exec.LookPath("git"); err == nil {
		commands := [][]string{
			{"-C", dir, "init", "-q", "-b", "main"},
			{"-C", dir, "add", "pyproject.toml", "main.py"},
			{"-C", dir, "-c", "user.name=stheme", "-c", "user.email=preview@invalid", "commit", "-qm", "preview"},
		}
		for _, args := range commands {
			cmd := exec.CommandContext(ctx, git, args...)
			if out, err := cmd.CombinedOutput(); err != nil {
				cleanup()
				return "", nil, fmt.Errorf("prepare preview git repository: %s", strings.TrimSpace(string(out)))
			}
		}
	}

	return dir, cleanup, nil
}

func (m *Manager) render(ctx context.Context, config []byte, width int, workDir string) (string, error) {
	tmp, err := os.CreateTemp("", "stheme-preview-*.toml")
	if err != nil {
		return "", fmt.Errorf("create preview config: %w", err)
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.Write(config); err != nil {
		tmp.Close()
		return "", fmt.Errorf("write preview config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close preview config: %w", err)
	}

	if width < 20 {
		width = 80
	}
	if workDir == "" {
		workDir = m.WorkDir
	}
	cmd := exec.CommandContext(ctx, m.Starship,
		"prompt",
		"--status", "0",
		"--terminal-width", fmt.Sprint(width),
		"--path", workDir,
	)
	cmd.Dir = workDir
	cmd.Env = previewEnvironment(name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(out))
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("render preview: %s", message)
	}
	return strings.Trim(string(out), "\r\n"), nil
}

func (m *Manager) Preview(ctx context.Context, item Theme, width int) (string, error) {
	config, err := m.Build(ctx, item)
	if err != nil {
		return "", err
	}
	workDir := m.PreviewDir
	if workDir == "" {
		workDir = m.WorkDir
	}
	return m.render(ctx, config, width, workDir)
}

func resolveWriteTarget(path string) (string, error) {
	seen := make(map[string]struct{})
	current := filepath.Clean(path)
	for {
		if _, ok := seen[current]; ok {
			return "", fmt.Errorf("resolve Starship config target: symlink loop at %s", current)
		}
		seen[current] = struct{}{}

		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return current, nil
		}
		if err != nil {
			return "", fmt.Errorf("inspect Starship config target: %w", err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			return current, nil
		}

		next, err := os.Readlink(current)
		if err != nil {
			return "", fmt.Errorf("read Starship config symlink: %w", err)
		}
		if !filepath.IsAbs(next) {
			next = filepath.Join(filepath.Dir(current), next)
		}
		current = filepath.Clean(next)
	}
}

func atomicWrite(path string, data []byte) error {
	target, err := resolveWriteTarget(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create Starship config directory: %w", err)
	}

	mode := os.FileMode(0o644)
	if info, err := os.Stat(target); err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat existing Starship config: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".starship.toml.*")
	if err != nil {
		return fmt.Errorf("create temporary Starship config: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return fmt.Errorf("set temporary config mode: %w", err)
	}
	if _, err := bytes.NewReader(data).WriteTo(tmp); err != nil {
		tmp.Close()
		return fmt.Errorf("write temporary Starship config: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temporary Starship config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary Starship config: %w", err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("replace Starship config: %w", err)
	}
	cleanup = false
	return nil
}

func (m *Manager) Apply(ctx context.Context, item Theme) error {
	config, err := m.Build(ctx, item)
	if err != nil {
		return err
	}
	if _, err := m.render(ctx, config, 80, m.WorkDir); err != nil {
		return err
	}
	return atomicWrite(m.Paths.Target, config)
}
