package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/Elliot-32/starship-theme-selector/internal/assets"
	"github.com/Elliot-32/starship-theme-selector/internal/theme"
	"github.com/Elliot-32/starship-theme-selector/internal/tui"
)

func New(version string) (*cobra.Command, error) {
	paths, err := theme.DefaultPaths()
	if err != nil {
		return nil, err
	}
	manager, err := theme.New("", paths, assets.DefaultSymbols)
	if err != nil {
		return nil, err
	}
	return newRoot(version, manager), nil
}

func newRoot(version string, manager *theme.Manager) *cobra.Command {
	root := &cobra.Command{
		Use:           "ssts [theme]",
		Short:         "Pick and apply Starship themes",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return applyByName(cmd.Context(), manager, args[0], cmd)
			}
			return runTUI(cmd.Context(), manager, cmd)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			themes, err := manager.Discover(cmd.Context())
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			values := make([]string, 0, len(themes))
			for _, item := range themes {
				values = append(values, fmt.Sprintf("%s\t%s", item.Name, item.Source))
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		},
	}
	root.CompletionOptions.DisableDefaultCmd = true

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "ssts version %s\n", version)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List built-in and custom themes",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			themes, err := manager.Discover(cmd.Context())
			if err != nil {
				return err
			}
			for _, item := range themes {
				fmt.Fprintln(cmd.OutOrStdout(), item.Name)
			}
			return nil
		},
	})

	completion := &cobra.Command{
		Use:       "completion [zsh|bash|fish|powershell|nushell]",
		Short:     "Generate shell completion",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"zsh", "bash", "fish", "powershell", "nushell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "zsh":
				return root.GenZshCompletion(cmd.OutOrStdout())
			case "bash":
				return root.GenBashCompletion(cmd.OutOrStdout())
			case "fish":
				return root.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return root.GenPowerShellCompletion(cmd.OutOrStdout())
			case "nushell":
				return genNushellCompletion(cmd.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell %q", args[0])
			}
		},
	}
	root.AddCommand(completion)
	return root
}

func applyByName(ctx context.Context, manager *theme.Manager, name string, cmd *cobra.Command) error {
	item, err := manager.Find(ctx, name)
	if err != nil {
		return err
	}
	if err := manager.Apply(ctx, item); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Applied %s Starship theme: %s\n", item.Source, item.Name)
	return nil
}

func runTUI(ctx context.Context, manager *theme.Manager, cmd *cobra.Command) error {
	if !isTerminal(os.Stdin) || !isTerminal(os.Stdout) {
		return fmt.Errorf("no theme specified and no interactive terminal is available")
	}
	themes, err := manager.Discover(ctx)
	if err != nil {
		return err
	}
	if len(themes) == 0 {
		return fmt.Errorf("no Starship themes found")
	}

	previewDir, cleanupPreview, err := theme.CreatePreviewProject(ctx)
	if err != nil {
		return err
	}
	defer cleanupPreview()
	manager.PreviewDir = previewDir
	defer func() { manager.PreviewDir = "" }()

	model := tui.New(manager, themes)
	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return fmt.Errorf("run theme picker: %w", err)
	}
	final, ok := result.(tui.Model)
	if !ok || final.Choice() == nil {
		return nil
	}
	item := *final.Choice()
	if err := manager.Apply(ctx, item); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Applied %s Starship theme: %s\n", item.Source, item.Name)
	return nil
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func formatError(err error) string {
	return "ssts: " + strings.TrimSpace(err.Error())
}

func Execute(root *cobra.Command) int {
	if err := root.Execute(); err != nil {
		fmt.Fprintln(root.ErrOrStderr(), formatError(err))
		return 1
	}
	return 0
}
