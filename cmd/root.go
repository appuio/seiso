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
		Use:   "image-cleanup",
		Short: "Cleans up images tags on remote registries",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.WithField("config", config).Debug("Parsed configuration.")
		},
	}
	config = cfg.NewDefaultConfig()
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Command aborted")
	}
}

func init() {
	defaults := cfg.NewDefaultConfig()
	rootCmd.PersistentFlags().String("log.level", defaults.Log.LogLevel, "Log level, one of [debug info warn error fatal]")
	rootCmd.PersistentFlags().BoolP("log.verbose", "v", defaults.Log.Verbose, "Shorthand for --log.level debug")
	rootCmd.PersistentFlags().BoolP("log.batch", "b", defaults.Log.Batch, "Use Batch mode (disables logging, prints deleted images only)")
	cobra.OnInitialize(initRootConfig)

}

// initRootConfig reads in cfg file and ENV variables if set.
func initRootConfig() {

	bindFlags(rootCmd.Flags())
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

func bindFlags(flagSet *pflag.FlagSet) {
	if err := viper.BindPFlags(flagSet); err != nil {
		log.WithError(err).Fatal("Could not bind flags")
	}
}

// SetVersion sets the version string in the help messages
func SetVersion(version string) {
	rootCmd.Version = version
}
