package cmd

import (
	"fmt"
	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/cleanup"
	"github.com/appuio/seiso/pkg/git"
	"github.com/appuio/seiso/pkg/openshift"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	historyCmd = &cobra.Command{
		Use:          "history [NAMESPACE/IMAGE]",
		Aliases:      []string{"hist"},
		Short:        "Clean up excessive image tags",
		Long:         `Clean up excessive image tags matching the commit hashes (prefix) of the git repository`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		PreRunE:      validateHistoryCommandInput,
		RunE:         ExecuteHistoryCleanupCommand,
	}
)

func init() {
	imagesCmd.AddCommand(historyCmd)
	defaults := cfg.NewDefaultConfig()

	addCommonFlagsForGit(historyCmd, defaults)
	historyCmd.PersistentFlags().IntP("keep", "k", defaults.History.Keep,
		"Keep most current <k> images. Does not include currently used image tags (if detected).")

}

func validateHistoryCommandInput(cmd *cobra.Command, args []string) (returnErr error) {
	defer showUsageOnError(cmd, returnErr)
	if len(args) == 0 {
		return missingImageNameError(config.Namespace)
	}
	namespace, image, err := splitNamespaceAndImagestream(args[0])
	if err != nil {
		return fmt.Errorf("could not parse image name: %w", err)
	}
	if config.Git.Tag && !git.IsValidSortValue(config.Git.SortCriteria) {
		return fmt.Errorf("invalid sort flag provided: %v", config.Git.SortCriteria)
	}
	log.WithFields(log.Fields{
		"namespace": namespace,
		"image":     image,
	}).Debug("Using image config")
	config.Namespace = namespace
	return nil
}

// ExecuteHistoryCleanupCommand executes the history cleanup command
func ExecuteHistoryCleanupCommand(cmd *cobra.Command, args []string) error {
	c := config.History
	namespace, imageName, _ := splitNamespaceAndImagestream(args[0])

	imageStreamObjectTags, err := openshift.GetImageStreamTags(namespace, imageName)
	if err != nil {
		return fmt.Errorf("could not retrieve image stream '%s/%s': %w", namespace, imageName, err)
	}

	var imageStreamTags []string
	for _, imageTag := range imageStreamObjectTags {
		imageStreamTags = append(imageStreamTags, imageTag.Tag)
	}

	matchOption := cleanup.MatchOptionPrefix
	if config.Git.Tag {
		matchOption = cleanup.MatchOptionExact
	}

	gitCandidates, err := git.GetGitCandidateList(&config.Git)
	if err != nil {
		return err
	}
	var matchingTags = cleanup.GetMatchingTags(&gitCandidates, &imageStreamTags, matchOption)

	activeImageStreamTags, err := openshift.GetActiveImageStreamTags(namespace, imageName, matchingTags)
	if err != nil {
		return fmt.Errorf("could not retrieve active image stream tags for '%s/%s': %w", namespace, imageName, err)
	}

	inactiveTags := cleanup.GetInactiveImageTags(&activeImageStreamTags, &matchingTags)
	inactiveTags = cleanup.LimitTags(&inactiveTags, c.Keep)
	if len(inactiveTags) == 0 {
		log.WithFields(log.Fields{
			"\n - namespace": namespace,
			"\n - 📺 image":   imageName,
		}).Info("No inactive image stream tags found")
		return nil
	}
	if config.Delete {
		DeleteImages(inactiveTags, imageName, namespace)
	} else {
		log.Infof("Showing results for --commit-limit=%d and --keep=%d", config.Git.CommitLimit, c.Keep)
		PrintImageTags(inactiveTags, imageName, namespace)
	}
	return nil
}
