package cli

import (
	"fmt"
	"io"
)

const nushellCompletion = `# Nushell completions for ssts

def "nu-complete ssts themes" [buffer, place] {
    ^ssts list | lines | where {|theme| $theme != "" }
}

def "nu-complete ssts shells" [buffer, place] {
    [bash zsh fish powershell nushell]
}

export extern "ssts" [
    theme?: string@"nu-complete ssts themes" # Starship theme name
    --help(-h)                                # Display help
    --version(-v)                             # Display version
]

export extern "ssts list" [
    --help(-h) # Display help
]

export extern "ssts version" [
    --help(-h) # Display help
]

export extern "ssts completion" [
    shell: string@"nu-complete ssts shells" # Shell to generate completion for
    --help(-h)                               # Display help
]

export extern "ssts help" [
    command?: string # Command to show help for
]
`

func genNushellCompletion(w io.Writer) error {
	if _, err := fmt.Fprint(w, nushellCompletion); err != nil {
		return fmt.Errorf("write Nushell completion: %w", err)
	}
	return nil
}
