package cmd

import (
	"fmt"
	"github.com/appuio/image-cleanup/pkg/git"
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

func getGitCandidateList(o *GitOptions) []string {
	logEvent := log.WithFields(log.Fields{
		"GitRepoPath": o.RepoPath,
		"CommitLimit": o.CommitLimit,
	})
	if o.Tag {
		candidates, err := git.GetTags(o.RepoPath, o.CommitLimit, git.SortOption(o.SortCriteria))
		if err != nil {
			logEvent.WithError(err).Fatal("Retrieving commit tags failed.")
		}
		return candidates
	} else {
		candidates, err := git.GetCommitHashes(o.RepoPath, o.CommitLimit)
		if err != nil {
			logEvent.WithError(err).Fatal("Retrieving commit hashes failed.")
		}
		return candidates
	}
}
