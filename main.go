package main

import (
	"os"

	"github.com/appuio/image-cleanup/cmd"
	"github.com/appuio/image-cleanup/version"

	log "github.com/sirupsen/logrus"
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

	configureLogging()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

func configureLogging() {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.SetOutput(os.Stderr)

	//TODO: To make this configurable via flag
	level, err := log.ParseLevel("debug")
	if err != nil {
		log.WithField("error", err).Warn("Using info level.")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}
