package cmd

import (
	"github.com/appuio/image-cleanup/pkg/docker"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// NewTagsCommand creates a cobra command to print the tags of a docker image
func NewTagsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Print the available tags",
		Long:  `Print the available tags for a Docker image`,
		Args:  cobra.ExactValidArgs(1),
		Run:   printImageTags,
	}

	return cmd
}

func printImageTags(cmd *cobra.Command, args []string) {
	image := args[0]
	
	imageTags, err := docker.GetImageTags(image)
	if err != nil {
		log.WithError(err).WithField("image", image).Fatal("Retrieving image tags failed.")
	}

	log.Println(imageTags)
}
