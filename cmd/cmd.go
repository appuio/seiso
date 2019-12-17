package cmd

import (
	"github.com/appuio/image-cleanup/cleanup"
	"github.com/appuio/image-cleanup/docker"
	"github.com/appuio/image-cleanup/git"
	"github.com/appuio/image-cleanup/openshift"
	"github.com/appuio/image-cleanup/version"
	"github.com/spf13/cobra"
)

// NewCleanupCommand creates the `image-cleanup` command
func NewCleanupCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "image-cleanup",
		Short: "image-cleanup cleans up docker images",
		Long:  "image-cleanup cleans up docker images.",
		Run:   runHelp,
	}

	cmds.AddCommand(cleanup.NewCleanupCommand())
	cmds.AddCommand(docker.NewTagCommand())
	cmds.AddCommand(git.NewGitCommand())
	cmds.AddCommand(openshift.NewImageStreamCommand())
	cmds.AddCommand(version.NewVersionCommand())

	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}
