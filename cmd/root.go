package cmd

import (
	"fmt"
	"github.com/appuio/seiso/pkg/kubernetes"
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
		Use:               "seiso",
		Short:             "Keeps your Kubernetes projects clean",
		PersistentPreRunE: parseConfig,
	}
	config        = cfg.NewDefaultConfig()
	koanfInstance = koanf.New(".")
	version       = "undefined"
)

// Execute is the main entrypoint of the CLI, it executes child commands as given by the user-defined flags and arguments.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("namespace", "n", config.Namespace, "Cluster namespace of current context")
	rootCmd.PersistentFlags().String("log.level", config.Log.LogLevel, "Log level, one of [debug info warn error fatal]")
	rootCmd.PersistentFlags().BoolP("log.verbose", "v", config.Log.Verbose, "Shorthand for \"--log.level debug\"")
	rootCmd.PersistentFlags().BoolP("log.batch", "b", config.Log.Batch,
		"Use Batch mode (Prints error to StdErr, StdOut is used to just print resource names, useful for piping)")
	cobra.OnInitialize(initRootConfig)
}

func initRootConfig() {
	bindFlags(rootCmd.Flags())
}

// parseConfig reads the flags and ENV vars
func parseConfig(cmd *cobra.Command, args []string) error {

	loadEnvironmentVariables()
	bindFlags(cmd.PersistentFlags())

	if err := koanfInstance.Unmarshal("", &config); err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}

	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	if config.Log.Verbose {
		config.Log.LogLevel = "debug"
	}
	if config.Log.Batch {
		log.SetOutput(os.Stderr)
		config.Log.LogLevel = "error"
	} else {
		log.SetOutput(os.Stdout)
	}
	level, err := log.ParseLevel(config.Log.LogLevel)
	if err != nil {
		log.WithError(err).Warn("Could not parse log level, fallback to info level")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
	if config.Namespace == "" {
		namespace, err := kubernetes.Namespace()
		if err != nil {
			return fmt.Errorf("unable to determine default namespace from Kubeconfig and --namespace not given: %w", err)
		}
		config.Namespace = namespace
	}
	log.Infof("Seiso %s", version)
	log.WithFields(log.Fields{
		"namespace": config.Namespace,
		"git":       config.Git,
		"log":       config.Log,
		"history":   config.History,
		"orphan":    config.Orphan,
		"resource":  config.Resource,
	}).Debug("Using config")
	return nil
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
func SetVersion(v string) {
	// We need to set both properties in order to break an initialization loop
	rootCmd.Version = v
	version = v
}
