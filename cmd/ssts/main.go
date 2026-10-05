package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/Elliot-32/starship-theme-selector/internal/cli"
)

var version = "dev"

func normalizeVersion(injected, moduleVersion string) string {
	if injected != "" && injected != "dev" {
		return strings.TrimPrefix(injected, "v")
	}
	if moduleVersion != "" && moduleVersion != "(devel)" {
		return strings.TrimPrefix(moduleVersion, "v")
	}
	return "dev"
}

func resolvedVersion(injected string) string {
	moduleVersion := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		moduleVersion = info.Main.Version
	}
	return normalizeVersion(injected, moduleVersion)
}

func main() {
	root, err := cli.New(resolvedVersion(version))
	if err != nil {
		fmt.Fprintln(os.Stderr, "ssts:", err)
		os.Exit(1)
	}
	os.Exit(cli.Execute(root))
}
