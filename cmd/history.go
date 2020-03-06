package cmd

import (
	"fmt"
	"github.com/appuio/image-cleanup/pkg/cleanup"
	"github.com/appuio/image-cleanup/pkg/git"
	"github.com/appuio/image-cleanup/pkg/openshift"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// HistoryCleanupOptions is a struct to support the cleanup command
type HistoryCleanupOptions struct {
	Force           bool
	Keep            int
	ImageRepository string
	Namespace       string
}

// NewHistoryCleanupCommand creates a cobra command to clean up images by comparing the git commit history.
func NewHistoryCleanupCommand() *cobra.Command {
	historyCleanupOptions := HistoryCleanupOptions{}
	gitOptions := GitOptions{}
	cmd := &cobra.Command{
		Use:     "history",
		Aliases: []string{"hist"},
		Short:   "Clean up excessive image tags",
		Long:    `Clean up excessive image tags matching the commit hashes (prefix) of the git repository`,
		Run: func(cmd *cobra.Command, args []string) {
			validateFlagCombinationInput(&gitOptions)
			ExecuteHistoryCleanupCommand(cmd, &historyCleanupOptions, &gitOptions, args)
		},
	}
	cmd.Flags().BoolVarP(&historyCleanupOptions.Force, "force", "f", false, "Confirm deletion of image tags.")
	cmd.Flags().IntVarP(&gitOptions.CommitLimit, "git-commit-limit", "l", 0,
		"Only look at the first <l> commits to compare with image tags. Use 0 (zero) for all commits. Limited effect if repo is a shallow clone.")
	cmd.Flags().StringVarP(&gitOptions.RepoPath, "git-repo-path", "p", ".", "Path to Git repository.")
	cmd.Flags().StringVarP(&historyCleanupOptions.ImageRepository, imageRepositoryCliFlag, "i", "", "Image repository in form of <namespace/repo>.")
	cmd.Flags().IntVarP(&historyCleanupOptions.Keep, "keep", "k", 10, "Keep most current <k> images.")
	cmd.Flags().BoolVarP(&gitOptions.Tag, "tags", "t", false, "Compare with Git tags instead of commit hashes.")
	cmd.Flags().StringVar(&gitOptions.SortCriteria, "sort", string(git.SortOptionVersion),
		fmt.Sprintf("Sort git tags by criteria. Only effective with --tags. Allowed values: %s", []git.SortOption{git.SortOptionVersion, git.SortOptionAlphabetic}))
	cmd.MarkFlagRequired("image-repository")
	return cmd
}

func ExecuteHistoryCleanupCommand(cmd *cobra.Command, o *HistoryCleanupOptions, gitOptions *GitOptions, args []string) {

	imageStreamObjectTags, err := openshift.GetImageStreamTags(o.Namespace, o.ImageRepository)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"ImageRepository": o.Namespace,
				"ImageName":       o.ImageRepository}).
			Fatal("Could not retrieve image stream.")
	}

	var imageStreamTags []string
	for _, imageTag := range imageStreamObjectTags {
		imageStreamTags = append(imageStreamTags, imageTag.Tag)
	}

	var matchOption cleanup.MatchOption
	if gitOptions.Tag {
		matchOption = cleanup.MatchOptionExact
	}

	gitCandidates := getGitCandidateList(gitOptions)
	var matchingTags = cleanup.GetMatchingTags(&gitCandidates, &imageStreamTags, matchOption)

	activeImageStreamTags, err := openshift.GetActiveImageStreamTags(o.Namespace, o.ImageRepository, imageStreamTags)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"ImageRepository": o.Namespace,
				"ImageName":       o.ImageRepository,
				"imageStreamTags": imageStreamTags}).
			Fatal("Could not retrieve active image stream tags.")
	}

	inactiveTags := cleanup.GetInactiveImageTags(&activeImageStreamTags, &matchingTags)

	inactiveTags = cleanup.LimitTags(&inactiveTags, o.Keep)

	PrintImageTags(cmd, inactiveTags)

	if o.Force {
		DeleteImages(inactiveTags, o.ImageRepository, o.Namespace)
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
}

func validateFlagCombinationInput(gitOptions *GitOptions) {

	if gitOptions.Tag && !git.IsValidSortValue(gitOptions.SortCriteria) {
		log.WithField("sort_criteria", gitOptions.SortCriteria).Fatal("Invalid sort criteria.")
	}

}
