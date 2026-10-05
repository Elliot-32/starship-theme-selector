package main

import (
	"fmt"
	"os"

	"github.com/Elliot-32/stheme/internal/cli"
)

var version = "dev"

func main() {
	root, err := cli.New(version)
	if err != nil {
		fmt.Fprintln(os.Stderr, "stheme:", err)
		os.Exit(1)
	}
	os.Exit(cli.Execute(root))
}
