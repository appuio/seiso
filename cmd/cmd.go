package cmd

import (
	"os"

	"github.com/appuio/image-cleanup/docker"
	"github.com/appuio/image-cleanup/git"
	"github.com/appuio/image-cleanup/openshift"
	"github.com/appuio/image-cleanup/version"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// Options is a struct holding the options of the root command
type Options struct {
	LogLevel string
}

// NewCleanupCommand creates the `image-cleanup` command
func NewCleanupCommand() *cobra.Command {
	o := Options{}
	cmds := &cobra.Command{
		Use:              "image-cleanup",
		Short:            "image-cleanup cleans up docker images",
		Long:             "image-cleanup cleans up docker images.",
		PersistentPreRun: o.init,
		Run:              runHelp,
	}

	cmds.PersistentFlags().StringVarP(&o.LogLevel, "verbosity", "v", "info", "Log level to use")

	cmds.AddCommand(docker.NewTagCommand())
	cmds.AddCommand(git.NewGitCommand())
	cmds.AddCommand(openshift.NewImageStreamCleanupCommand())
	cmds.AddCommand(version.NewVersionCommand())

	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func (o *Options) init(cmd *cobra.Command, args []string) {
	configureLogging(o.LogLevel)
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
