package cmd

import (
	"fmt"
	"github.com/appuio/image-cleanup/cfg"
	"github.com/appuio/image-cleanup/cleanup"
	"github.com/appuio/image-cleanup/git"
	"github.com/appuio/image-cleanup/openshift"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	historyCmd = &cobra.Command{
		Use:     "history [IMAGE]",
		Aliases: []string{"hist"},
		Short:   "Clean up excessive image tags",
		Long:    `Clean up excessive image tags matching the commit hashes (prefix) of the git repository`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateHistoryCommandInput(args); err != nil {
				return err
			}
			ExecuteHistoryCleanupCommand(args)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(historyCmd)
	defaults := cfg.NewDefaultConfig()

	addCommonFlagsForGit(historyCmd, defaults)
	historyCmd.PersistentFlags().IntP("keep", "k", defaults.History.Keep, "Keep most current <k> images.")

	bindFlags(historyCmd.PersistentFlags())

}

func validateHistoryCommandInput(args []string) error {
	if _, _, err := splitNamespaceAndImagestream(args[0]); err != nil {
		return fmt.Errorf("could not parse image name: %w", err)
	}
	if config.Git.Tag && !git.IsValidSortValue(config.Git.SortCriteria) {
		return fmt.Errorf("invalid sort flag provided: %v", config.Git.SortCriteria)
	}
	return nil
}

func ExecuteHistoryCleanupCommand(args []string) {

	c := config.History
	namespace, image, _ := splitNamespaceAndImagestream(args[0])

	imageStreamObjectTags, err := openshift.GetImageStreamTags(namespace, image)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"ImageRepository": namespace,
				"ImageName":       image,
			}).
			Fatal("Could not retrieve image stream.")
	}

	var imageStreamTags []string
	for _, imageTag := range imageStreamObjectTags {
		imageStreamTags = append(imageStreamTags, imageTag.Tag)
	}

	matchOption := cleanup.MatchOptionDefault
	if config.Git.Tag {
		matchOption = cleanup.MatchOptionExact
	}

	gitCandidates := git.GetGitCandidateList(&config.Git)
	var matchingTags = cleanup.GetMatchingTags(&gitCandidates, &imageStreamTags, matchOption)

	activeImageStreamTags, err := openshift.GetActiveImageStreamTags(namespace, image, imageStreamTags)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"ImageRepository": namespace,
				"ImageName":       image,
				"imageStreamTags": imageStreamTags}).
			Fatal("Could not retrieve active image stream tags.")
	}

	inactiveTags := cleanup.GetInactiveImageTags(&activeImageStreamTags, &matchingTags)
	inactiveTags = cleanup.LimitTags(&inactiveTags, c.Keep)

	PrintImageTags(inactiveTags)

	if config.Force {
		DeleteImages(inactiveTags, image, namespace)
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
}
