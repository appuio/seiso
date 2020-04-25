package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/cleanup"
	"github.com/appuio/seiso/pkg/git"
	"github.com/appuio/seiso/pkg/openshift"
	"github.com/karrick/tparse"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	orphanCommandLongDescription = `Sometimes images get tagged manually or by branches or force-pushed commits that do not exist anymore.
This command deletes images that are not found in the git history.`
	orphanDeletionPatternCliFlag = "deletion-pattern"
	orphanOlderThanCliFlag       = "older-than"
)

var (
	// orphanCmd represents a cobra command to clean up images by comparing the git commit history. It removes any
	// image tags that are not found in the git history by given criteria.
	orphanCmd = &cobra.Command{
		Use:          "orphans [NAMESPACE/IMAGE]",
		Short:        "Clean up unknown image tags",
		Long:         orphanCommandLongDescription,
		Aliases:      []string{"orph", "orphan"},
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOrphanCommandInput(args); err != nil {
				cmd.Usage()
				return err
			}
			return ExecuteOrphanCleanupCommand(args)
		},
	}
)

func init() {
	imagesCmd.AddCommand(orphanCmd)
	defaults := cfg.NewDefaultConfig()

	addCommonFlagsForGit(orphanCmd, defaults)
	orphanCmd.PersistentFlags().String(orphanOlderThanCliFlag, defaults.Orphan.OlderThan,
		"Delete images that are older than the duration. Ex.: [1y2mo3w4d5h6m7s]")
	orphanCmd.PersistentFlags().StringP(orphanDeletionPatternCliFlag, "r", defaults.Orphan.OrphanDeletionRegex,
		"Delete images that match the regex, defaults to matching Git SHA commits")
}

func validateOrphanCommandInput(args []string) error {
	if len(args) == 0 {
		return nil
	}
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

// ExecuteOrphanCleanupCommand executes the orphan cleanup command
func ExecuteOrphanCleanupCommand(args []string) error {
	if len(args) == 0 {
		return listImages()
	}
	o := config.Orphan
	namespace, imageName, _ := splitNamespaceAndImagestream(args[0])

	allImageTags, err := openshift.GetImageStreamTags(namespace, imageName)
	if err != nil {
		return fmt.Errorf("could not retrieve image stream '%v/%v': %w", namespace, imageName, err)
	}

	cutOffDateTime, _ := parseCutOffDateTime(o.OlderThan)
	orphanIncludeRegex, _ := parseOrphanDeletionRegex(o.OrphanDeletionRegex)

	matchOption := cleanup.MatchOptionPrefix
	if config.Git.Tag {
		matchOption = cleanup.MatchOptionExact
	}

	gitCandidates, err := git.GetGitCandidateList(&config.Git)
	if err != nil {
		return err
	}
	imageTagList := cleanup.FilterImageTagsByTime(&allImageTags, cutOffDateTime)
	imageTagList = cleanup.FilterOrphanImageTags(&gitCandidates, &imageTagList, matchOption)
	imageTagList = cleanup.FilterByRegex(&imageTagList, orphanIncludeRegex)
	imageTagList, err = cleanup.FilterActiveImageTags(namespace, imageName, imageTagList, &imageTagList)
	if err != nil {
		return err
	}
	if len(imageTagList) == 0 {
		log.WithFields(log.Fields{
			"\n - namespace": namespace,
			"\n - 📺 image":   imageName,
		}).Info("No orphaned image stream tags found")
		return nil
	}

	if config.Delete {
		DeleteImages(imageTagList, imageName, namespace)
	} else {
		PrintImageTags(imageTagList, imageName, namespace)
	}

	return nil
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
	if !strings.Contains(repo, "/") {
		namespace = config.Namespace
		image = repo
	} else {
		paths := strings.SplitAfter(repo, "/")
		if len(paths) >= 3 {
			namespace = paths[1]
			image = paths[2]
		} else {
			namespace = paths[0]
			image = paths[1]
		}
	}
	if namespace == "" {
		return "", "", errors.New("missing or invalid namespace")
	}
	if image == "" {
		return "", "", errors.New("missing or invalid image name")
	}
	return strings.TrimSuffix(namespace, "/"), image, nil
}
