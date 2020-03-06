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
	CommitLimit     int
	RepoPath        string
	Keep            int
	ImageRepository string
	Namespace       string
	Tag             bool
	SortCriteria    string
}

// NewHistoryCleanupCommand creates a cobra command to clean up images by comparing the git commit history.
func NewHistoryCleanupCommand() *cobra.Command {
	o := HistoryCleanupOptions{}
	cmd := &cobra.Command{
		Use:     "history",
		Aliases: []string{"hist"},
		Short:   "Clean up excessive image tags",
		Long:    `Clean up excessive image tags matching the commit hashes (prefix) of the git repository`,
		Run: func(cmd *cobra.Command, args []string) {
			validateFlagCombinationInput(&o)
			ExecuteHistoryCleanupCommand(cmd, &o, args)
		},
	}
	cmd.Flags().BoolVarP(&o.Force, "force", "f", false, "Confirm deletion of image tags.")
	cmd.Flags().IntVarP(&o.CommitLimit, "git-commit-limit", "l", 0,
		"Only look at the first <l> commits to compare with image tags. Use 0 (zero) for all commits. Limited effect if repo is a shallow clone.")
	cmd.Flags().StringVarP(&o.RepoPath, "git-repo-path", "p", ".", "Path to Git repository.")
	cmd.Flags().StringVarP(&o.ImageRepository, imageRepositoryCliFlag, "i", "", "Image repository in form of <namespace/repo>.")
	cmd.Flags().IntVarP(&o.Keep, "keep", "k", 10, "Keep most current <k> images.")
	cmd.Flags().BoolVarP(&o.Tag, "tags", "t", false, "Compare with Git tags instead of commit hashes.")
	cmd.Flags().StringVar(&o.SortCriteria, "sort", string(git.SortOptionVersion),
		fmt.Sprintf("Sort git tags by criteria. Only effective with --tags. Allowed values: %s", []git.SortOption{git.SortOptionVersion, git.SortOptionAlphabetic}))
	cmd.MarkFlagRequired("image-repository")
	return cmd
}

func ExecuteHistoryCleanupCommand(cmd *cobra.Command, o *HistoryCleanupOptions, args []string) {

	var matchValues []string
	if o.Tag {
		var err error
		matchValues, err = git.GetTags(o.RepoPath, o.CommitLimit, git.SortOption(o.SortCriteria))
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{
					"RepoPath":    o.RepoPath,
					"CommitLimit": o.CommitLimit}).
				Fatal("Retrieving commit tags failed.")
		}
	} else {
		var err error
		matchValues, err = git.GetCommitHashes(o.RepoPath, o.CommitLimit)
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{
					"RepoPath":    o.RepoPath,
					"CommitLimit": o.CommitLimit}).
				Fatal("Retrieving commit hashes failed.")
		}
	}

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
	if o.Tag {
		matchOption = cleanup.MatchOptionExact
	}

	var matchingTags = cleanup.GetMatchingTags(&matchValues, &imageStreamTags, matchOption)

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

	log.WithField("inactiveTags", inactiveTags).Info("Compiled list of image tags to delete.")

	if o.Force {
		DeleteImages(inactiveTags, o.ImageRepository, o.Namespace)
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
}

func validateFlagCombinationInput(o *HistoryCleanupOptions) {

	if o.Tag && !git.IsValidSortValue(o.SortCriteria) {
		log.WithField("sort_criteria", o.SortCriteria).Fatal("Invalid sort criteria.")
	}

}
