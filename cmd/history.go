package cmd

import (
	"fmt"
	"github.com/appuio/image-cleanup/cleanup"
	"github.com/appuio/image-cleanup/git"
	"github.com/appuio/image-cleanup/openshift"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// HistoryCleanupOptions is a struct to support the cleanup command
type HistoryCleanupOptions struct {
	Force           bool
	Keep            int
	ImageRepository string
}

var (
	historyCleanupOptions = HistoryCleanupOptions{}
	historyCmd            = &cobra.Command{
		Use:     "history",
		Aliases: []string{"hist"},
		Short:   "Clean up excessive image tags",
		Long:    `Clean up excessive image tags matching the commit hashes (prefix) of the git repository`,
		Run: func(cmd *cobra.Command, args []string) {
			validateFlagCombinationInput()
			ExecuteHistoryCleanupCommand(cmd, args)
		},
	}
)

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.Flags().BoolVarP(&historyCleanupOptions.Force, "force", "f", false, "Confirm deletion of image tags.")
	historyCmd.Flags().IntVarP(&gitOptions.CommitLimit, "git-commit-limit", "l", 0,
		"Only look at the first <l> commits to compare with image tags. Use 0 (zero) for all commits. Limited effect if repo is a shallow clone.")
	historyCmd.Flags().StringVarP(&gitOptions.RepoPath, "git-repo-path", "p", ".", "Path to Git repository.")
	historyCmd.Flags().StringVarP(&historyCleanupOptions.ImageRepository, imageRepositoryCliFlag, "i", "", "Image repository in form of <namespace/repo>.")
	historyCmd.Flags().IntVarP(&historyCleanupOptions.Keep, "keep", "k", 10, "Keep most current <k> images.")
	historyCmd.Flags().BoolVarP(&gitOptions.Tag, "tags", "t", false, "Compare with Git tags instead of commit hashes.")
	historyCmd.Flags().StringVar(&gitOptions.SortCriteria, "sort", string(git.SortOptionVersion),
		fmt.Sprintf("Sort git tags by criteria. Only effective with --tags. Allowed values: %s", []git.SortOption{git.SortOptionVersion, git.SortOptionAlphabetic}))
	historyCmd.MarkFlagRequired("image-repository")

}

func ExecuteHistoryCleanupCommand(cmd *cobra.Command, args []string) {

	namespace, image, _ := splitNamespaceAndImagestream(historyCleanupOptions.ImageRepository)

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
	if gitOptions.Tag {
		matchOption = cleanup.MatchOptionExact
	}

	gitCandidates := git.GetGitCandidateList(&gitOptions)
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
	inactiveTags = cleanup.LimitTags(&inactiveTags, historyCleanupOptions.Keep)

	PrintImageTags(cmd, inactiveTags)

	if historyCleanupOptions.Force {
		DeleteImages(inactiveTags, image, namespace)
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
}

func validateFlagCombinationInput() {

	if gitOptions.Tag && !git.IsValidSortValue(gitOptions.SortCriteria) {
		log.WithFields(log.Fields{
			"error": "invalid sort criteria",
			"sort":  gitOptions.SortCriteria,
		}).Fatal("Could not parse sort criteria.")
	}

}
