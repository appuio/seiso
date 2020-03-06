package cmd

import (
	"errors"
	"fmt"
	"github.com/appuio/image-cleanup/pkg/cleanup"
	"github.com/appuio/image-cleanup/pkg/git"
	"github.com/appuio/image-cleanup/pkg/openshift"
	"github.com/karrick/tparse"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
	"time"
)

// OrphanCleanupOptions holds the user-defined settings
type OrphanCleanupOptions struct {
	Force               bool
	ImageRepository     string
	OlderThan           string
	OrphanDeletionRegex string
}

const (
	orphanCommandLongDescription = `Sometimes images get tagged manually or by branches or force-pushed commits that do not exist anymore.
This command deletes images that are not found in the git history.`
	orphanDeletionPatternCliFlag = "orphan-deletion-pattern"
	orphanOlderThanCliFlag       = "older-than"
)

var (
	orphanCleanupOptions = OrphanCleanupOptions{}
	// orphanCmd represents a cobra command to clean up images by comparing the git commit history. It removes any
	// image tags that are not found in the git history by given criteria.
	orphanCmd = &cobra.Command{
		Use:     "orphans",
		Short:   "Clean up unknown image tags",
		Long:    orphanCommandLongDescription,
		Aliases: []string{"orph"},
		RunE: func(cmd *cobra.Command, args []string) error {
			validateOrphanCommandInput()
			return ExecuteOrphanCleanupCommand(cmd, args)
		},
	}
)

func init() {
	rootCmd.AddCommand(orphanCmd)

	orphanCmd.Flags().BoolVarP(&orphanCleanupOptions.Force, "force", "f", false, "Confirm deletion of image tags.")
	orphanCmd.Flags().IntVarP(&gitOptions.CommitLimit, "git-commit-limit", "l", 0,
		"Only look at the first <l> commits to compare with tags. Use 0 (zero) for all commits. Limited effect if repo is a shallow clone.")
	orphanCmd.Flags().StringVarP(&gitOptions.RepoPath, "git-repo-path", "p", ".", "Path to Git repository")
	orphanCmd.Flags().StringVarP(&orphanCleanupOptions.ImageRepository, imageRepositoryCliFlag, "i", "", "Image repository (e.g. namespace/repo)")
	orphanCmd.Flags().BoolVarP(&gitOptions.Tag, "tags", "t", false,
		"Instead of comparing commit history, it will compare git tags with the existing image tags, removing any image tags that do not match")
	orphanCmd.Flags().StringVar(&gitOptions.SortCriteria, "sort", string(git.SortOptionVersion),
		fmt.Sprintf("Sort git tags by criteria. Only effective with --tags. Allowed values: [%s, %s]", git.SortOptionVersion, git.SortOptionAlphabetic))
	orphanCmd.Flags().StringVar(&orphanCleanupOptions.OlderThan, orphanOlderThanCliFlag, "2mo",
		"Delete images that are older than the duration. Ex.: [1y2mo3w4d5h6m7s]")
	orphanCmd.Flags().StringVarP(&orphanCleanupOptions.OrphanDeletionRegex, orphanDeletionPatternCliFlag, "r", "^[a-z0-9]{40}$",
		"Delete images that match the regex, defaults to matching Git SHA commits")
	orphanCmd.MarkFlagRequired("image-repository")

}

func validateOrphanCommandInput() {

	if _, _, err := splitNamespaceAndImagestream(orphanCleanupOptions.ImageRepository); err != nil {
		log.WithError(err).
			WithField(imageRepositoryCliFlag, orphanCleanupOptions.ImageRepository).
			Fatal("Could not parse image repository.")
	}

	if _, err := parseOrphanDeletionRegex(orphanCleanupOptions.OrphanDeletionRegex); err != nil {
		log.WithError(err).
			WithField(orphanDeletionPatternCliFlag, orphanCleanupOptions.OrphanDeletionRegex).
			Fatal("Could not parse orphan deletion pattern.")
	}

	if _, err := parseCutOffDateTime(orphanCleanupOptions.OlderThan); err != nil {
		log.WithError(err).
			WithField(orphanOlderThanCliFlag, orphanCleanupOptions.OlderThan).
			Fatal("Could not parse cut off date.")
	}

	if gitOptions.Tag && !git.IsValidSortValue(gitOptions.SortCriteria) {
		log.WithFields(log.Fields{
			"error": "invalid sort criteria",
			"sort":  gitOptions.SortCriteria,
		}).Fatal("Could not parse sort criteria.")
	}

}

func ExecuteOrphanCleanupCommand(cmd *cobra.Command, args []string) error {

	gitCandidates := git.GetGitCandidateList(&gitOptions)

	namespace, imageName, _ := splitNamespaceAndImagestream(orphanCleanupOptions.ImageRepository)

	imageStreamObjectTags, err := openshift.GetImageStreamTags(namespace, imageName)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"ImageRepository": orphanCleanupOptions.ImageRepository,
			}).
			Fatal("Could not retrieve image stream.")
	}

	cutOffDateTime, _ := parseCutOffDateTime(orphanCleanupOptions.OlderThan)
	imageStreamTags := cleanup.FilterImageTagsByTime(&imageStreamObjectTags, cutOffDateTime)

	var matchOption cleanup.MatchOption
	if gitOptions.Tag {
		matchOption = cleanup.MatchOptionExact
	}

	orphanIncludeRegex, _ := parseOrphanDeletionRegex(orphanCleanupOptions.OrphanDeletionRegex)
	var matchingTags []string
	matchingTags = cleanup.GetOrphanImageTags(&gitCandidates, &imageStreamTags, matchOption)
	matchingTags = cleanup.FilterByRegex(&imageStreamTags, orphanIncludeRegex)

	activeImageStreamTags, err := openshift.GetActiveImageStreamTags(namespace, imageName, imageStreamTags)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"ImageRepository": orphanCleanupOptions.ImageRepository,
				"ImageName":       imageName,
				"imageStreamTags": imageStreamTags}).
			Fatal("Could not retrieve active image stream tags.")
	}

	log.WithField("activeTags", activeImageStreamTags).Debug("Currently found active image tags")
	inactiveImageTags := cleanup.GetInactiveImageTags(&activeImageStreamTags, &matchingTags)

	PrintImageTags(cmd, inactiveImageTags)

	if orphanCleanupOptions.Force {
		DeleteImages(inactiveImageTags, imageName, namespace)
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
	return nil
}

func parseOrphanDeletionRegex(orphanIncludeRegex string) (*regexp.Regexp, error) {
	r, err := regexp.Compile(orphanIncludeRegex)
	if err != nil {
		log.WithError(err).
			WithField(orphanDeletionPatternCliFlag, orphanIncludeRegex).
			Fatal("Invalid orphan include regex.")
	}
	return r, err
}

func parseCutOffDateTime(olderThan string) (time.Time, error) {
	if len(olderThan) > 0 {
		cutOffDateTime, err := tparse.ParseNow(time.RFC3339, "now-"+olderThan)
		if err != nil {
			log.WithError(err).
				WithField(orphanOlderThanCliFlag, olderThan).
				Fatal("Could not parse --older-than flag.")
			return time.Now(), err
		}
		return cutOffDateTime, nil
	}
	return time.Now(), nil
}

func splitNamespaceAndImagestream(repo string) (namespace string, image string, err error) {
	if repo == "" || !strings.Contains(repo, "/") {
		return "", "", errors.New("missing or invalid image repository name")
	}
	paths := strings.SplitAfter(repo, "/")
	if len(paths) >= 3 {
		namespace = paths[1]
		image = paths[2]
	} else {
		namespace = paths[0]
		image = paths[1]
	}
	if image == "" {
		return "", "", errors.New("missing or invalid image repository name")
	}
	return strings.TrimSuffix(namespace, "/"), image, nil
}
