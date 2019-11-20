package main

import (
	"github.com/appuio/image-cleanup/commands"
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	version = "latest"
	commit  = "snapshot"
	date    = "unknown"
)

func main() {
	commands.Version = version
	commands.Commit = commit
	commands.Date = date

	ConfigureLogging()
	commands.Execute()
}

func ConfigureLogging() {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.SetOutput(os.Stdout)

	//TODO: To make this configurable via flag
	level, err := log.ParseLevel("debug")
	if err != nil {
		log.WithField("error", err).Warn("Using info level.")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}
