package main

import (
	"os"

	"github.com/appuio/image-cleanup/cmd"
)

var (
	version string
	commit  string
	date    string
)

func main() {

	command := cmd.NewCleanupCommand(cmd.Build{Version: version, Commit: commit, Date: date})

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
