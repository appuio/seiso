package main

import (
	"fmt"
	"github.com/appuio/image-cleanup/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	version = "unknown"
	commit  = "dirty"
	date    = "today"
	rootCmd = &cobra.Command{
		Use:     "image-cleanup",
		Short:   "Cleans up images tags on remote registries",
		Version: fmt.Sprintf("%s, commit %s, date %s", version, commit, date),
	}
)

// Options is a struct holding the options of the root command
type Options struct {
	LogLevel string
}

func initializeCmd() {
	o := Options{}
	rootCmd.PersistentFlags().StringVarP(&o.LogLevel, "logLevel", "v", "info", "Log level to use")
	configureLogging(o.LogLevel)
}

func main() {
	rootCmd.AddCommand(
		cmd.NewHistoryCleanupCommand(),
		cmd.NewOrphanCleanupCommand(),
	)
	cobra.OnInitialize(initializeCmd)
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Error("Command failed with error.")
		os.Exit(1)
	}
}

func configureLogging(logLevel string) {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.SetOutput(os.Stderr)

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithField("error", err).Warn("Using info level.")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}
