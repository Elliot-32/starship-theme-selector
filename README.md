# stheme

A Starship theme picker with live preview.

## Install

```sh
mise use -g github:Elliot-32/stheme
```

Or with Go:

```sh
go install github.com/Elliot-32/stheme/cmd/stheme@latest
```

`starship` must be available in `PATH`.

## Usage

```sh
# Pick a theme interactively
stheme

# Apply a theme directly
stheme tokyo-night

# List available themes
stheme list
```

## Custom themes

Put custom Starship configs in:

```text
~/.config/stheme/themes/*.toml
```

The filename becomes the theme name. For example:

```text
~/.config/stheme/themes/my-theme.toml
```

can be applied with:

```sh
stheme my-theme
```

Custom themes take precedence over built-in presets with the same name.

## Shell completion

Generate completion with:

```sh
stheme completion zsh
```

For mise-completions-sync:

```toml
stheme = "standard"
```
