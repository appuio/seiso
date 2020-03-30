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
		Use:          "history [PROJECT/IMAGE]",
		Aliases:      []string{"hist"},
		Short:        "Clean up excessive image tags",
		Long:         `Clean up excessive image tags matching the commit hashes (prefix) of the git repository`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateHistoryCommandInput(args); err != nil {
				cmd.Usage()
				return err
			}
			return ExecuteHistoryCleanupCommand(args)
		},
	}
)

func init() {
	imagesCmd.AddCommand(historyCmd)
	defaults := cfg.NewDefaultConfig()

	addCommonFlagsForGit(historyCmd, defaults)
	historyCmd.PersistentFlags().IntP("keep", "k", defaults.History.Keep,
		"Keep most current <k> images. Does not include currently used image tags (if detected).")

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

// ExecuteHistoryCleanupCommand executes the history cleanup command
func ExecuteHistoryCleanupCommand(args []string) error {

	c := config.History
	namespace, image, _ := splitNamespaceAndImagestream(args[0])

	imageStreamObjectTags, err := openshift.GetImageStreamTags(namespace, image)
	if err != nil {
		return fmt.Errorf("could not retrive image stream '%s/%s': %w", namespace, image, err)
	}

	var imageStreamTags []string
	for _, imageTag := range imageStreamObjectTags {
		imageStreamTags = append(imageStreamTags, imageTag.Tag)
	}

	matchOption := cleanup.MatchOptionDefault
	if config.Git.Tag {
		matchOption = cleanup.MatchOptionExact
	}

	gitCandidates, err := git.GetGitCandidateList(&config.Git)
	if err != nil {
		return err
	}
	var matchingTags = cleanup.GetMatchingTags(&gitCandidates, &imageStreamTags, matchOption)

	activeImageStreamTags, err := openshift.GetActiveImageStreamTags(namespace, image, imageStreamTags)
	if err != nil {
		return fmt.Errorf("could not retrieve active image stream tags for '%s/%s': %w", namespace, image, err)
	}

	inactiveTags := cleanup.GetInactiveImageTags(&activeImageStreamTags, &matchingTags)
	inactiveTags = cleanup.LimitTags(&inactiveTags, c.Keep)

	PrintImageTags(inactiveTags)

	if config.Force {
		DeleteImages(inactiveTags, image, namespace)
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
	return nil
}
