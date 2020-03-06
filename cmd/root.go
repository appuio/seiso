package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

// Options is a struct holding the options of the root command
type (
	Options struct {
		LogLevel string
		Batch    bool
		Verbose  bool
	}
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "image-cleanup",
		Short: "Cleans up images tags on remote registries",
	}
	options = &Options{}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&options.LogLevel, "logLevel", "info", "Log level to use")
	rootCmd.PersistentFlags().BoolVarP(&options.Verbose, "verbose", "v", false, "Shorthand for --logLevel debug")
	rootCmd.PersistentFlags().BoolVarP(&options.Batch, "batch", "b", false, "Use Batch mode (disables logging, prints deleted images only)")
	cobra.OnInitialize(initConfig)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if options.Batch {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	if options.Verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		level, err := log.ParseLevel(options.LogLevel)
		if err != nil {
			log.WithField("error", err).Warn("Using info level.")
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(level)
		}
	}
}

// SetVersion sets the version string in the help messages
func SetVersion(version string) {
	rootCmd.Version = version
}
