package cmd

import (
	"github.com/appuio/image-cleanup/cfg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:              "image-cleanup",
		Short:            "Cleans up images tags on remote registries",
		PersistentPreRun: parseConfig,
	}
	config = cfg.NewDefaultConfig()
)

// Execute is the main entrypoint of the CLI, it executes child commands as given by the user-defined flags and arguments.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Command aborted")
	}
}

func init() {
	rootCmd.PersistentFlags().String("log.level", config.Log.LogLevel, "Log level, one of [debug info warn error fatal]")
	rootCmd.PersistentFlags().BoolP("log.verbose", "v", config.Log.Verbose, "Shorthand for --log.level debug")
	rootCmd.PersistentFlags().BoolP("log.batch", "b", config.Log.Batch, "Use Batch mode (disables logging, prints deleted images only)")
	cobra.OnInitialize(initRootConfig)

}

func initRootConfig() {
	bindFlags(rootCmd.Flags())
}

// parseConfig reads the flags and ENV vars
func parseConfig(cmd *cobra.Command, args []string) {
	bindFlags(cmd.PersistentFlags())
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
		config.Log.LogLevel = "debug"
	}
	level, err := log.ParseLevel(config.Log.LogLevel)
	if err != nil {
		log.WithField("error", err).Warn("Using info level.")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}

	log.WithField("config", config).Debug("Parsed configuration.")
}

func bindFlags(flagSet *pflag.FlagSet) {
	if err := viper.BindPFlags(flagSet); err != nil {
		log.WithError(err).Fatal("Could not bind flags")
	}
}

// SetVersion sets the version string in the help messages
func SetVersion(version string) {
	rootCmd.Version = version
}
