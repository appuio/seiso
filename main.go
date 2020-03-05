package main

import (
	"fmt"
	"github.com/appuio/image-cleanup/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
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
	options = &Options{}
)

// Options is a struct holding the options of the root command
type Options struct {
	LogLevel string
	Batch    bool
}

func initializeCmd() {
	configureLogging()
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&options.LogLevel, "logLevel", "v", "info", "Log level to use")
	rootCmd.PersistentFlags().BoolVarP(&options.Batch, "batch", "b", false, "Use Batch mode (disables logging, prints deleted images only)")
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

func configureLogging() {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if options.Batch {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}

	level, err := log.ParseLevel(options.LogLevel)
	if err != nil {
		log.WithField("error", err).Warn("Using info level.")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}
