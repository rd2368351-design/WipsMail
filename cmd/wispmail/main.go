package main

import (
	"os"

	"github.com/wispmail/wispmail/cmd/wispmail/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}