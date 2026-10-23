package main

import (
	"os"

	// Embeds the time zone database, so consumption --from/--to can be read in
	// the home's time zone on systems without one installed (Windows).
	_ "time/tzdata"

	"github.com/kristofferrisa/powerctl-cli/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
