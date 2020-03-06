package cmd

import (
	"fmt"
	"github.com/appuio/image-cleanup/pkg/openshift"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	imageRepositoryCliFlag = "image-repository"
)

type (
	GitOptions struct {
		CommitLimit  int
		RepoPath     string
		Tag          bool
		SortCriteria string
	}
)
var (
	gitOptions = GitOptions{}
)

func DeleteImages(imageTags []string, imageName string, namespace string) {
	for _, inactiveTag := range imageTags {
		err := openshift.DeleteImageStreamTag(namespace, openshift.BuildImageStreamTagName(imageName, inactiveTag))
		if err == nil {
			log.WithField("imageTag", inactiveTag).Info("Deleted image tag")
		} else {
			log.WithError(err).WithField("imageTag", inactiveTag).Error("Could not delete image")
		}
	}
}

// PrintImageTags prints the given image tags line by line. In batch mode, only the tag name is printed, otherwise default
// log with info level
func PrintImageTags(cmd *cobra.Command, imageTags []string) {
	if cmd.Parent().PersistentFlags().Lookup("batch").Value.String() == "true" {
		for _, tag := range imageTags {
			fmt.Println(tag)
		}
	} else {
		for _, tag := range imageTags {
			log.WithField("image_tag", tag).Info("image tag candidate")
		}
	}
}

