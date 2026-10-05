# stheme

A small Starship theme picker and applier with an interactive Bubble Tea preview.

## Features

- `stheme` opens an interactive theme picker.
- The selected theme is previewed with a real `starship prompt` render before it is applied.
- Preview rendering uses a stable synthetic `main` Git repository with a Python project so Git and language modules are visible consistently.
- The Bubble Tea UI uses a fixed-size Lip Gloss layout and alternate screen to avoid redraw artifacts while previews update.
- `stheme THEME` applies a theme directly.
- Built-in Starship presets and custom TOML themes are shown together.
- `[os.symbols]` is replaced as a complete table, using bundled Nerd Font distro symbols by default.
- User overrides can replace the bundled symbol table.
- Native shell completion generation works with mise-completions-sync's `standard` convention.
- No fzf, Python, dasel, or shell wrapper is required.

## Requirements

- `starship` available in `PATH`.
- `git` is optional; when present, the interactive preview includes a real `main` branch.

## Install with mise

```sh
mise use -g github:Elliot-32/stheme
```

Or with Go:

```sh
go install github.com/Elliot-32/stheme/cmd/stheme@latest
```

## Usage

```sh
# Interactive picker with live preview
stheme

# Apply directly
stheme tokyo-night

# List themes
stheme list

# Generate completions
stheme completion zsh
```

`stheme <TAB>` supports dynamic completion when the generated shell completion is installed.
For mise-completions-sync, use:

```toml
stheme = "standard"
```

## Custom themes

Put custom themes in:

```text
~/.config/stheme/themes/*.toml
```

A custom theme with the same name as a built-in Starship preset takes precedence.

Environment overrides:

- `STHEME_CONFIG_DIR`: stheme config directory.
- `STHEME_THEME_DIR`: custom theme directory.
- `STHEME_SYMBOLS`: path to an alternative `[os.symbols]` TOML file.
- `STARSHIP_CONFIG`: destination Starship config path, matching Starship itself.

## OS symbols

The bundled `os-symbols.toml` is applied by replacing the entire theme `[os.symbols]` table. To customize it, create:

```text
~/.config/stheme/os-symbols.toml
```

with the same structure:

```toml
[os.symbols]
Arch = ""
CachyOS = ""
```
