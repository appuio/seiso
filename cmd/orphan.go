package cmd

import (
	"github.com/appuio/image-cleanup/pkg/cleanup"
	"github.com/appuio/image-cleanup/pkg/git"
	"github.com/appuio/image-cleanup/pkg/kubernetes"
	"github.com/appuio/image-cleanup/pkg/openshift"
	"github.com/karrick/tparse"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"regexp"
	"time"
)

// OrphanCleanupOptions holds the user-defined settings
type OrphanCleanupOptions struct {
	Force              bool
	CommitLimit        int
	RepoPath           string
	ImageStream        string
	Namespace          string
	Tag                bool
	SortCriteria       string
	OlderThan          string
	OrphanIncludeRegex string
}

const (
	orphanCommandLongDescription = `Sometimes images get tagged manually or by branches or force-pushed commits that do not exist anymore.This command deletes images that are not found in the git history.`
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
			err := validateOrphanCommandInput(&o)
			return ExecuteOrphanCleanupCommand(cmd, &o, args)
				return err
			}
			return ExecuteOrphanCleanupCommand(&o, args)
		},
	}
	cmd.Flags().BoolVarP(&o.Force, "force", "f", false, "Confirm deletion of image tags.")
	cmd.Flags().IntVarP(&o.CommitLimit, "git-commit-limit", "l", 0,
		"Only look at the first <l> commits to compare with tags. Use 0 (zero) for all commits. Limited effect if repo is a shallow clone.")
	cmd.Flags().StringVarP(&o.RepoPath, "git-repo-path", "p", ".", "Path to Git repository")
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().BoolVarP(&o.Tag, "tags", "t", false,
		"Instead of comparing commit history, it will compare git tags with the existing image tags, removing any image tags that do not match")
	cmd.Flags().StringVar(&o.SortCriteria, "sort", string(git.SortOptionVersion), "sort tags by criteria. Allowed values: [version, alphabetical]")
	cmd.Flags().StringVar(&o.OlderThan, "older-than", "2mo",
		"delete images that are older than the duration. Ex.: [1y2mo3w4d5h6m7s]")
	cmd.Flags().StringVarP(&o.OrphanIncludeRegex, "orphan-deletion-pattern", "i", "^[a-z0-9]{40}$",
		"Delete images that match the regex, defaults to matching Git SHA commits")
	return cmd
}

func validateOrphanCommandInput(o *OrphanCleanupOptions) error {

	if _, err := parseOrphanIncludeRegex(o.OrphanIncludeRegex); err != nil {
		return err
	}

	if _, err := parseCutOffDateTime(o.OlderThan); err != nil {
		return err
	}

	if o.Tag && !git.IsValidSortValue(o.SortCriteria) {
		log.WithField("sort_criteria", o.SortCriteria).Fatal("Invalid sort criteria.")
	}

	return nil
}

func ExecuteOrphanCleanupCommand(cmd *cobra.Command, o *OrphanCleanupOptions, args []string) error {

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

	}

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

	imageStreamObjectTags, err := openshift.GetImageStreamTags(o.Namespace, o.ImageStream)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"Namespace":   o.Namespace,
				"ImageStream": o.ImageStream}).
			Fatal("Could not retrieve image stream.")
	}

	cutOffDateTime, _ := parseCutOffDateTime(o.OlderThan)
	imageStreamTags := cleanup.FilterImageTagsByTime(&imageStreamObjectTags, cutOffDateTime)

	var matchOption cleanup.MatchOption
	if o.Tag {
		matchOption = cleanup.MatchOptionExact
	}

	orphanIncludeRegex, _ := parseOrphanIncludeRegex(o.OrphanIncludeRegex)
	var matchingTags []string
	matchingTags = cleanup.GetOrphanImageTags(&matchValues, &imageStreamTags, matchOption)
	matchingTags = cleanup.FilterByRegex(&imageStreamTags, orphanIncludeRegex)

	activeImageStreamTags, err := openshift.GetActiveImageStreamTags(o.Namespace, o.ImageStream, imageStreamTags)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"Namespace":       o.Namespace,
				"ImageStream":     o.ImageStream,
				"imageStreamTags": imageStreamTags}).
			Fatal("Could not retrieve active image stream tags.")
	}

	inactiveImageTags := cleanup.GetInactiveImageTags(&activeImageStreamTags, &matchingTags)

	log.WithField("inactiveImageTags", inactiveImageTags).Info("Tags for deletion")

	if o.Force {
		DeleteImages(inactiveImageTags, imageName, namespace)
			err:= openshift.DeleteImageStreamTag(o.Namespace, openshift.BuildImageStreamTagName(o.ImageStream, inactiveTag))
			if err == nil {
				log.WithField("imageTag", inactiveTag).Info("Deleted image tag")
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
	return nil
}
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
	return nil
}

func parseOrphanIncludeRegex(orphanIncludeRegex string) (*regexp.Regexp, error) {
	r, err := regexp.Compile(orphanIncludeRegex)
	if err != nil {
		log.WithError(err).
			WithField("orphanIncludeRegex", orphanIncludeRegex).
			Fatal("Invalid orphan include regex.")
	}
	return r, err
}

func parseCutOffDateTime(olderThan string) (time.Time, error) {
	if len(olderThan) > 0 {
		cutOffDateTime, err := tparse.ParseNow(time.RFC3339, "now-"+olderThan)
		if err != nil {
			log.WithError(err).
				WithField("older-than", olderThan).
				Fatal("Could not parse --older-than flag.")
			return time.Now(), err
		}
		return cutOffDateTime, nil
	}
	return time.Now(), nil
}
