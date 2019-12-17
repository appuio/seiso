package main

import (
	"os"

	"github.com/appuio/image-cleanup/cmd"
	"github.com/appuio/image-cleanup/version"
)

// CLI Version constants
const (
	VERSION = "latest"
	COMMIT  = "snapshot"
	DATE    = "unknown"
)

func main() {
	version.Version = VERSION
	version.Commit = COMMIT
	version.Date = DATE

	command := cmd.NewCleanupCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
