package cmd

import (
	"fmt"

	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/git"
	"github.com/appuio/seiso/pkg/kubernetes"
	"github.com/appuio/seiso/pkg/openshift"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
func PrintImageTags(imageTags []string) {
	if config.Log.Batch {
		for _, tag := range imageTags {
			fmt.Println(tag)
		}
	} else {
		for _, tag := range imageTags {
			log.WithField("imageTag", tag).Info("Found image tag candidate.")
		}
	}
}

// addCommonFlagsForGit sets up the force flag, as well as the common git flags. Adding the flags to the root cmd would make those
// global, even for commands that do not need them, which might be overkill.
func addCommonFlagsForGit(cmd *cobra.Command, defaults *cfg.Configuration) {
	cmd.PersistentFlags().BoolP("force", "f", defaults.Force, "Confirm deletion of image tags.")
	cmd.PersistentFlags().IntP("commit-limit", "l", defaults.Git.CommitLimit,
		"Only look at the first <l> commits to compare with tags. Use 0 (zero) for all commits. Limited effect if repo is a shallow clone.")
	cmd.PersistentFlags().StringP("repo-path", "p", defaults.Git.RepoPath, "Path to Git repository")
	cmd.PersistentFlags().BoolP("tags", "t", defaults.Git.Tag,
		"Instead of comparing commit history, it will compare git tags with the existing image tags, removing any image tags that do not match")
	cmd.PersistentFlags().String("sort", defaults.Git.SortCriteria,
		fmt.Sprintf("Sort git tags by criteria. Only effective with --tags. Allowed values: [%s, %s]", git.SortOptionVersion, git.SortOptionAlphabetic))
}

func listImages() error {
	ns, err := kubernetes.Namespace()
	if err != nil {
		return err
	}
	imageStreams, err := openshift.ListImageStreams(ns)
	if err != nil {
		return err
	}
	imageNames := []string{}
	for _, image := range imageStreams {
		imageNames = append(imageNames, image.Name)
	}
	log.WithField("project", ns).WithField("images", imageNames).Info("Please select an image. The following images are available")
	return nil
}
