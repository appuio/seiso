package cmd

import (
	"errors"
	"fmt"
	"github.com/appuio/image-cleanup/cfg"
	"github.com/appuio/image-cleanup/cleanup"
	"github.com/appuio/image-cleanup/git"
	"github.com/appuio/image-cleanup/openshift"
	"github.com/karrick/tparse"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
	"time"
)

const (
	orphanCommandLongDescription = `Sometimes images get tagged manually or by branches or force-pushed commits that do not exist anymore.
This command deletes images that are not found in the git history.`
	orphanDeletionPatternCliFlag = "orphan-deletion-pattern"
	orphanOlderThanCliFlag       = "older-than"
)

var (
	// orphanCmd represents a cobra command to clean up images by comparing the git commit history. It removes any
	// image tags that are not found in the git history by given criteria.
	orphanCmd = &cobra.Command{
		Use:     "orphans [IMAGE]",
		Short:   "Clean up unknown image tags",
		Long:    orphanCommandLongDescription,
		Aliases: []string{"orph"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOrphanCommandInput(args); err != nil {
				return err
			}
			ExecuteOrphanCleanupCommand(cmd, args)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(orphanCmd)
	defaults := cfg.NewDefaultConfig()

	addCommonFlagsForGit(orphanCmd, defaults)
	orphanCmd.PersistentFlags().String(orphanOlderThanCliFlag, defaults.Orphan.OlderThan,
		"Delete images that are older than the duration. Ex.: [1y2mo3w4d5h6m7s]")
	orphanCmd.PersistentFlags().StringP(orphanDeletionPatternCliFlag, "r", defaults.Orphan.OrphanDeletionRegex,
		"Delete images that match the regex, defaults to matching Git SHA commits")

}

func validateOrphanCommandInput(args []string) error {

	o := config.Orphan
	if _, _, err := splitNamespaceAndImagestream(args[0]); err != nil {
		return err
	}
	if _, err := parseOrphanDeletionRegex(o.OrphanDeletionRegex); err != nil {
		return fmt.Errorf("could not parse orphan deletion pattern: %w", err)
	}

	if _, err := parseCutOffDateTime(o.OlderThan); err != nil {
		return fmt.Errorf("could not parse older-than flag: %w", err)
	}

	if config.Git.Tag && !git.IsValidSortValue(config.Git.SortCriteria) {
		return fmt.Errorf("invalid sort flag provided: %v", config.Git.SortCriteria)
	}
	return nil
}

func ExecuteOrphanCleanupCommand(cmd *cobra.Command, args []string) {

	gitCandidates := git.GetGitCandidateList(&config.Git)

	o := config.Orphan
	namespace, imageName, _ := splitNamespaceAndImagestream(args[0])

	imageStreamObjectTags, err := openshift.GetImageStreamTags(namespace, imageName)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"ImageRepository": namespace,
				"ImageName":       imageName,
			}).
			Fatal("Could not retrieve image stream.")
	}

	cutOffDateTime, _ := parseCutOffDateTime(o.OlderThan)
	imageStreamTags := cleanup.FilterImageTagsByTime(&imageStreamObjectTags, cutOffDateTime)

	matchOption := cleanup.MatchOptionDefault
	if config.Git.Tag {
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
				"ImageRepository": namespace,
				"ImageName":       imageName,
				"imageStreamTags": imageStreamTags}).
			Fatal("Could not retrieve active image stream tags.")
	}

	log.WithField("activeTags", activeImageStreamTags).Debug("Found currently active image tags")
	inactiveImageTags := cleanup.GetInactiveImageTags(&activeImageStreamTags, &matchingTags)

	PrintImageTags(cmd, inactiveImageTags)

	if config.Force {
		DeleteImages(inactiveImageTags, imageName, namespace)
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
}

func parseOrphanDeletionRegex(orphanIncludeRegex string) (*regexp.Regexp, error) {
	return regexp.Compile(orphanIncludeRegex)
}

func parseCutOffDateTime(olderThan string) (time.Time, error) {
	if len(olderThan) == 0 {
		return time.Now(), nil
	}
	cutOffDateTime, err := tparse.ParseNow(time.RFC3339, "now-"+olderThan)
	if err != nil {
		return time.Now(), err
	}
	return cutOffDateTime, nil
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
