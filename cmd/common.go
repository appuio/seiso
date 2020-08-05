package cmd

import (
	"fmt"
	"strings"

	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/git"
	"github.com/appuio/seiso/pkg/openshift"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteImages deletes a list of image tags
func DeleteImages(imageTags []string, imageName string, namespace string) {
	for _, inactiveTag := range imageTags {
		log.Infof("Deleting %s/%s:%s", namespace, imageName, inactiveTag)

		if err := openshift.DeleteImageStreamTag(namespace, openshift.BuildImageStreamTagName(imageName, inactiveTag)); err != nil {
			log.WithError(err).Errorf("Failed to delete %s/%s:%s", namespace, imageName, inactiveTag)
		}
	}
}

// PrintImageTags prints the given image tags line by line. In batch mode, only the tag name is printed, otherwise default
// log with info level
func PrintImageTags(imageTags []string, imageName string, namespace string) {
	if config.Log.Batch {
		for _, tag := range imageTags {
			fmt.Println(tag)
		}
	} else {
		for _, tag := range imageTags {
			log.Infof("Found image tag candidate: %s/%s:%s", namespace, imageName, tag)
		}
	}
}

// addCommonFlagsForGit sets up the delete flag, as well as the common git flags. Adding the flags to the root cmd would make those
// global, even for commands that do not need them, which might be overkill.
func addCommonFlagsForGit(cmd *cobra.Command, defaults *cfg.Configuration) {
	cmd.PersistentFlags().BoolP("delete", "d", defaults.Delete, "Effectively delete image tags found")
	cmd.PersistentFlags().IntP("commit-limit", "l", defaults.Git.CommitLimit,
		"Only look at the first <l> commits to compare with tags. Use 0 (zero) for all commits. Limited effect if repo is a shallow clone.")
	cmd.PersistentFlags().StringP("repo-path", "p", defaults.Git.RepoPath, "Path to Git repository")
	cmd.PersistentFlags().BoolP("tags", "t", defaults.Git.Tag,
		"Instead of comparing commit history, it will compare git tags with the existing image tags, removing any image tags that do not match")
	cmd.PersistentFlags().String("sort", defaults.Git.SortCriteria,
		fmt.Sprintf("Sort git tags by criteria. Only effective with --tags. Allowed values: [%s, %s]", git.SortOptionVersion, git.SortOptionAlphabetic))
}

// toListOptions converts "key=value"-labels to Kubernetes LabelSelector
func toListOptions(labels []string) metav1.ListOptions {
	labelSelector := fmt.Sprintf(strings.Join(labels, ","))
	return metav1.ListOptions{
		LabelSelector: labelSelector,
	}
}

func missingLabelSelectorError(namespace, resource string) error {
	return fmt.Errorf("label selector with --label expected. You can print out available labels with \"kubectl -n %s get %s --show-labels\"", namespace, resource)
}

func missingImageNameError(namespace string) error {
	return fmt.Errorf("no image name given. On OpenShift, you can print available image streams with \"oc -n %s get imagestreams\"", namespace)
}
