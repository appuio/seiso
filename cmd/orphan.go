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
	CommitLimit         int
	GitRepoPath         string
	ImageRepository     string
	Tag                 bool
	SortCriteria        string
	OlderThan           string
	OrphanDeletionRegex string
}

const (
	orphanCommandLongDescription = `Sometimes images get tagged manually or by branches or force-pushed commits that do not exist anymore.
This command deletes images that are not found in the git history.`
	orphanDeletionPatternCliFlag = "orphan-deletion-pattern"
	orphanOlderThanCliFlag       = "older-than"
)

// NewOrphanCleanupCommand creates a cobra command to clean up images by comparing the git commit history. It removes any
// image tags that are not found in the git history by given criteria.
func NewOrphanCleanupCommand() *cobra.Command {
	o := OrphanCleanupOptions{}
	cmd := &cobra.Command{
		Use:     "orphans",
		Aliases: []string{"orph"},
		Short:   "Clean up unknown image tags",
		Long:    orphanCommandLongDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			validateOrphanCommandInput(&o)
			return ExecuteOrphanCleanupCommand(cmd, &o, args)
		},
	}
	cmd.Flags().BoolVarP(&o.Force, "force", "f", false, "Confirm deletion of image tags.")
	cmd.Flags().IntVarP(&o.CommitLimit, "git-commit-limit", "l", 0,
		"Only look at the first <l> commits to compare with tags. Use 0 (zero) for all commits. Limited effect if repo is a shallow clone.")
	cmd.Flags().StringVarP(&o.GitRepoPath, "git-repo-path", "p", ".", "Path to Git repository")
	cmd.Flags().StringVarP(&o.ImageRepository, imageRepositoryCliFlag, "i", "", "Image repository (e.g. namespace/repo)")
	cmd.Flags().BoolVarP(&o.Tag, "tags", "t", false,
		"Instead of comparing commit history, it will compare git tags with the existing image tags, removing any image tags that do not match")
	cmd.Flags().StringVar(&o.SortCriteria, "sort", string(git.SortOptionVersion),
		fmt.Sprintf("Sort git tags by criteria. Only effective with --tags. Allowed values: [%s, %s]", git.SortOptionVersion, git.SortOptionAlphabetic))
	cmd.Flags().StringVar(&o.OlderThan, orphanOlderThanCliFlag, "2mo",
		"delete images that are older than the duration. Ex.: [1y2mo3w4d5h6m7s]")
	cmd.Flags().StringVarP(&o.OrphanDeletionRegex, orphanDeletionPatternCliFlag, "r", "^[a-z0-9]{40}$",
		"Delete images that match the regex, defaults to matching Git SHA commits")
	cmd.MarkFlagRequired("image-repository")
	return cmd
}

func validateOrphanCommandInput(o *OrphanCleanupOptions) {

	if _, _, err := splitNamespaceAndImagestream(o.ImageRepository); err != nil {
		log.WithError(err).
			WithField(imageRepositoryCliFlag, o.ImageRepository).
			Fatal("Could not parse image repository.")
	}

	if _, err := parseOrphanDeletionRegex(o.OrphanDeletionRegex); err != nil {
		log.WithError(err).
			WithField(orphanDeletionPatternCliFlag, o.OrphanDeletionRegex).
			Fatal("Could not parse orphan deletion pattern.")
	}

	if _, err := parseCutOffDateTime(o.OlderThan); err != nil {
		log.WithError(err).
			WithField(orphanOlderThanCliFlag, o.OlderThan).
			Fatal("Could not parse cut off date.")
	}

	if o.Tag && !git.IsValidSortValue(o.SortCriteria) {
		log.WithFields(log.Fields{
			"error": "invalid sort criteria",
			"sort":  o.SortCriteria,
		}).Fatal("Could not parse sort criteria.")
	}

}

func ExecuteOrphanCleanupCommand(cmd *cobra.Command, o *OrphanCleanupOptions, args []string) error {

	gitCandidates := getGitCandidateList(o)

	namespace, imageName, err := splitNamespaceAndImagestream(o.ImageRepository)

	imageStreamObjectTags, err := openshift.GetImageStreamTags(namespace, imageName)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"ImageRepository": o.ImageRepository,
			}).
			Fatal("Could not retrieve image stream.")
	}

	cutOffDateTime, _ := parseCutOffDateTime(o.OlderThan)
	imageStreamTags := cleanup.FilterImageTagsByTime(&imageStreamObjectTags, cutOffDateTime)

	var matchOption cleanup.MatchOption
	if o.Tag {
		matchOption = cleanup.MatchOptionExact
	}

	orphanIncludeRegex, _ := parseOrphanDeletionRegex(o.OrphanDeletionRegex)
	var matchingTags []string
	matchingTags = cleanup.GetOrphanImageTags(&gitCandidates, &imageStreamTags, matchOption)
	matchingTags = cleanup.FilterByRegex(&imageStreamTags, orphanIncludeRegex)

	activeImageStreamTags, err := openshift.GetActiveImageStreamTags(namespace, imageName, imageStreamTags)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"ImageRepository": o.ImageRepository,
				"ImageName":       imageName,
				"imageStreamTags": imageStreamTags}).
			Fatal("Could not retrieve active image stream tags.")
	}

	log.WithField("activeTags", activeImageStreamTags).Debug("Currently found active image tags")
	inactiveImageTags := cleanup.GetInactiveImageTags(&activeImageStreamTags, &matchingTags)

	PrintImageTags(cmd, inactiveImageTags)

	if o.Force {
		DeleteImages(inactiveImageTags, imageName, namespace)
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
	return nil
}

func getGitCandidateList(o *OrphanCleanupOptions) []string {
	logEvent := log.WithFields(log.Fields{
		"GitRepoPath": o.GitRepoPath,
		"CommitLimit": o.CommitLimit,
	})
	if o.Tag {
		candidates, err := git.GetTags(o.GitRepoPath, o.CommitLimit, git.SortOption(o.SortCriteria))
		if err != nil {
			logEvent.WithError(err).Fatal("Retrieving commit tags failed.")
		}
		return candidates
	} else {
		candidates, err := git.GetCommitHashes(o.GitRepoPath, o.CommitLimit)
		if err != nil {
			logEvent.WithError(err).Fatal("Retrieving commit hashes failed.")
		}
		return candidates
	}
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
