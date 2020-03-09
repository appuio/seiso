package cmd

import (
	"fmt"
	"github.com/appuio/image-cleanup/cfg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "image-cleanup",
		Short: "Cleans up images tags on remote registries",
	}
	config = cfg.NewDefaultConfig()
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
	defaults := cfg.NewDefaultConfig()
	rootCmd.PersistentFlags().String("logLevel", defaults.Log.LogLevel, "Log level to use")
	rootCmd.PersistentFlags().BoolP("verbose", "v", defaults.Log.Verbose, "Shorthand for --logLevel debug")
	rootCmd.PersistentFlags().BoolP("batch", "b", defaults.Log.Batch, "Use Batch mode (disables logging, prints deleted images only)")
	cobra.OnInitialize(initConfig)

}

// initConfig reads in cfg file and ENV variables if set.
func initConfig() {

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.WithError(err).Fatal("Could not bind flags.")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&config); err != nil {
		log.WithError(err).Fatal("Could not read config")
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if config.Log.Batch {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}
	if config.Log.Verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		level, err := log.ParseLevel(config.Log.LogLevel)
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
