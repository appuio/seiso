package cmd

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/appuio/seiso/cfg"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:              "seiso",
		Short:            "Keeps your Kubernetes projects clean",
		PersistentPreRun: parseConfig,
	}
	config        = cfg.NewDefaultConfig()
	koanfInstance = koanf.New(".")
)

// Execute is the main entrypoint of the CLI, it executes child commands as given by the user-defined flags and arguments.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("namespace", "n", config.Namespace, "Cluster namespace of current context")
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

	loadEnvironmentVariables()
	bindFlags(cmd.PersistentFlags())

	if err := koanfInstance.Unmarshal("", &config); err != nil {
		log.WithError(err).Fatal("Could not read config")
	}

	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
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
		log.WithError(err).Warn("Could not parse log level, fallback to info level")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
}

func loadEnvironmentVariables() {
	prefix := "SEISO_"
	err := koanfInstance.Load(env.Provider(prefix, ".", func(s string) string {
		/*
			Configuration can contain hierarchies (YAML, etc.) and CLI flags dashes. To read environment variables with
			hierarchies and dashes we replace the hierarchy delimiter with double underscore and dashes with single underscore,
			so that parent.child-with-dash becomes PARENT__CHILD_WITH_DASH
		*/
		s = strings.TrimPrefix(s, prefix)
		s = strings.Replace(strings.ToLower(s), "__", ".", -1)
		s = strings.Replace(strings.ToLower(s), "_", "-", -1)
		return s
	}), nil)
	if err != nil {
		log.WithError(err).Fatal("Could not load environment variables")
	}
}

func bindFlags(flagSet *pflag.FlagSet) {
	err := koanfInstance.Load(posflag.Provider(flagSet, ".", koanfInstance), nil)
	if err != nil {
		log.WithError(err).Fatal("Could not bind flags")
	}
}

// SetVersion sets the version string in the help messages
func SetVersion(version string) {
	rootCmd.Version = version
}
