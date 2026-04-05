package main

import (
	"os"

	"github.com/charmbracelet/x/term"
)

func isTTY() bool {
	return term.IsTerminal(os.Stdout.Fd())
}
