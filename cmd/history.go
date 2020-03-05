package cmd

import (
	"github.com/appuio/image-cleanup/pkg/cleanup"
	"github.com/appuio/image-cleanup/pkg/git"
	"github.com/appuio/image-cleanup/pkg/kubernetes"
	"github.com/appuio/image-cleanup/pkg/openshift"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// HistoryCleanupOptions is a struct to support the cleanup command
type HistoryCleanupOptions struct {
	Force       bool
	CommitLimit int
	RepoPath    string
	Keep        int
	ImageStream string
	Namespace   string
	Tag         bool
	Sorted      string
}

// NewHistoryCleanupCommand creates a cobra command to clean up images by comparing the git commit history.
func NewHistoryCleanupCommand() *cobra.Command {
	o := HistoryCleanupOptions{}
	cmd := &cobra.Command{
		Use:     "history",
		Aliases: []string{"hist"},
		Short:   "Clean up excessive image tags",
		Long:    `Clean up excessive image tags matching the commit hashes (prefix) of the git repository`,
		Run:     o.cleanupImageStreamTags,
	}
	cmd.Flags().BoolVarP(&o.Force, "force", "f", false, "Confirm deletion of image tags.")
	cmd.Flags().IntVarP(&o.CommitLimit, "git-commit-limit", "l", 0,
		"Only look at the first <l> commits to compare with tags. Use 0 (zero) for all commits. Limited effect if repo is a shallow clone.")
	cmd.Flags().StringVarP(&o.RepoPath, "git-repo-path", "p", ".", "Path to Git repository")
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().IntVarP(&o.Keep, "keep", "k", 10, "keep most current <k> images")
	cmd.Flags().BoolVarP(&o.Tag, "tags", "t", false, "use tags instead of commit hashes")
	cmd.Flags().StringVar(&o.Sorted, "sort", string(git.SortOptionVersion), "sort tags by criteria. Allowed values: [version, alphabetical]")
	return cmd
}

func (o *HistoryCleanupOptions) cleanupImageStreamTags(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		o.ImageStream = args[0]
	}

	validateFlagCombinationInput(o)

	if len(o.Namespace) == 0 {
		namespace, err := kubernetes.Namespace()
		if err != nil {
			log.WithError(err).Fatal("Could not retrieve default namespace from kubeconfig.")
		}

		o.Namespace = namespace
	}

	if len(o.ImageStream) == 0 {
		imageStreams, err := openshift.GetImageStreams(o.Namespace)
		if err != nil {
			log.WithError(err).
				WithField("namespace", o.Namespace).
				Fatal("Could not retrieve image streams.")
		}

		log.Printf("No image stream provided as argument. Available image streams for namespace %s: %s", o.Namespace, imageStreams)

		return
	}

	var matchValues []string
	if o.Tag {
		var err error
		matchValues, err = git.GetTags(o.RepoPath, o.CommitLimit, git.SortOption(o.Sorted))
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

	imageStreamObjectTags, err := openshift.GetImageStreamTags(o.Namespace, o.ImageStream)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"ImageRepository": o.Namespace,
				"ImageStream":     o.ImageStream}).
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

	activeImageStreamTags, err := openshift.GetActiveImageStreamTags(o.Namespace, o.ImageStream, imageStreamTags)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"ImageRepository": o.Namespace,
				"ImageStream":     o.ImageStream,
				"imageStreamTags": imageStreamTags}).
			Fatal("Could not retrieve active image stream tags.")
	}

	inactiveTags := cleanup.GetInactiveImageTags(&activeImageStreamTags, &matchingTags)

	inactiveTags = cleanup.LimitTags(&inactiveTags, o.Keep)

	log.WithField("inactiveTags", inactiveTags).Info("Compiled list of image tags to delete.")

	if o.Force {
		DeleteImages(inactiveTags, o.ImageStream, o.Namespace)
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
}

func validateFlagCombinationInput(o *HistoryCleanupOptions) {

	if o.Tag && !git.IsValidSortValue(o.Sorted) {
		log.WithField("sort_criteria", o.Sorted).Fatal("Invalid sort criteria.")
	}

}
