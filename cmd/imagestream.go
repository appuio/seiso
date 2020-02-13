package cmd

import (
	log "github.com/sirupsen/logrus"
	"time"
	"regexp"

	"github.com/appuio/image-cleanup/pkg/cleanup"
	"github.com/appuio/image-cleanup/pkg/git"
	"github.com/appuio/image-cleanup/pkg/kubernetes"
	"github.com/appuio/image-cleanup/pkg/openshift"
	"github.com/spf13/cobra"
	"github.com/karrick/tparse"
)

// ImageStreamCleanupOptions is a struct to support the cleanup command
type ImageStreamCleanupOptions struct {
	Force       			 bool
	CommitLimit 			 int
	RepoPath    			 string
	Keep        			 int
	ImageStream 			 string
	Namespace   			 string
	Tag         			 bool
	Sorted      			 string
	Orphan      			 bool
	OlderThan    			 string
	OrphanIncludeRegex       string
}

// NewImageStreamCleanupCommand creates a cobra command to clean up an imagestream based on commits
func NewImageStreamCleanupCommand() *cobra.Command {
	o := ImageStreamCleanupOptions{}
	cmd := &cobra.Command{
		Use:     "imagestream",
		Aliases: []string{"is"},
		Short:   "Clean up excessive image tags",
		Long:    `Clean up excessive image tags matching the commit hashes (prefix) of the git repository`,
		Run:     o.cleanupImageStreamTags,
	}
	cmd.Flags().BoolVarP(&o.Force, "force", "f", false, "delete image stream tags")
	cmd.Flags().IntVarP(&o.CommitLimit, "git-commit-limit", "l", 0, "only look at the first <n> commits to compare with tags or use 0 for all commits")
	cmd.Flags().StringVarP(&o.RepoPath, "git-repo-path", "p", ".", "absolute path to Git repository (for current dir use: $PWD)")
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().IntVarP(&o.Keep, "keep", "k", 10, "keep most current <n> images")
	cmd.Flags().BoolVarP(&o.Tag, "tag", "t", false, "use tags instead of commit hashes")
	cmd.Flags().StringVar(&o.Sorted, "sort", string(git.SortOptionVersion), "sort tags by criteria. Allowed values: [version, alphabetical]")
	cmd.Flags().BoolVarP(&o.Orphan, "orphan", "o", false, "delete images that do not match any git commit")
	cmd.Flags().StringVar(&o.OlderThan, "older-than", "", "delete images that are older than the duration. Ex.: [1y2mo3w4d5h6m7s]")
	cmd.Flags().StringVarP(&o.OrphanIncludeRegex, "orphan-deletion-pattern", "i", "^[a-z0-9]{40}$", "delete images that match the regex, works only with the -o flag, defaults to matching Git SHA commits")
	return cmd
}

func (o *ImageStreamCleanupOptions) cleanupImageStreamTags(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		o.ImageStream = args[0]
	}

	validateFlagCombinationInput(o)

	orphanIncludeRegex := parseOrphanIncludeRegex(o.OrphanIncludeRegex)

	cutOffDateTime := parseCutOffDateTime(o.OlderThan)

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
				"Namespace":   o.Namespace,
				"ImageStream": o.ImageStream}).
			Fatal("Could not retrieve image stream.")
	}

	imageStreamTags := cleanup.FilterImageTagsByTime(&imageStreamObjectTags, cutOffDateTime)

	var matchOption cleanup.MatchOption
	if o.Tag {
		matchOption = cleanup.MatchOptionExact
	}

	var matchingTags []string
	if o.Orphan {
		matchingTags = cleanup.GetOrphanImageTags(&matchValues, &imageStreamTags, matchOption)
		matchingTags = cleanup.FilterByRegex(&imageStreamTags, orphanIncludeRegex)
	} else {
		matchingTags = cleanup.GetMatchingTags(&matchValues, &imageStreamTags, matchOption)
	}

	activeImageStreamTags, err := openshift.GetActiveImageStreamTags(o.Namespace, o.ImageStream, imageStreamTags)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"Namespace":       o.Namespace,
				"ImageStream":     o.ImageStream,
				"imageStreamTags": imageStreamTags}).
			Fatal("Could not retrieve active image stream tags.")
	}

	inactiveTags := cleanup.GetInactiveTags(&activeImageStreamTags, &matchingTags)

	inactiveTags = cleanup.LimitTags(&inactiveTags, o.Keep)

	log.WithField("inactiveTags", inactiveTags).Info("Tags for deletion")

	if o.Force {
		for _, inactiveTag := range inactiveTags {
			openshift.DeleteImageStreamTag(o.Namespace, openshift.BuildImageStreamTagName(o.ImageStream, inactiveTag))
			log.WithField("inactiveTag", inactiveTag).Info("Deleted image stream tag")
		}
	} else {
		log.Info("--force was not specified. Nothing has been deleted.")
	}
}

func validateFlagCombinationInput(o *ImageStreamCleanupOptions) {

	if o.Orphan == false && o.OrphanIncludeRegex != "^[a-z0-9]{40}$" {
		log.WithFields(log.Fields{"Orphan": o.Orphan, "Regex": o.OrphanIncludeRegex}).
			Fatal("Missing Orphan flag")
	}

	if o.Tag && !git.IsValidSortValue(o.Sorted) {
		log.WithField("sort_criteria", o.Sorted).Fatal("Invalid sort criteria.")
	}

	if o.CommitLimit !=0 && o.Orphan == true {
		log.WithFields(log.Fields{"CommitLimit": o.CommitLimit, "Orphan": o.Orphan}).
			Fatal("Mutually exclusive flags")
	}
}

func parseOrphanIncludeRegex(orphanIncludeRegex string) *regexp.Regexp {
	regexp, err := regexp.Compile(orphanIncludeRegex)
	if err != nil {
		log.WithField("orphanIncludeRegex", orphanIncludeRegex).
			Fatal("Invalid orphan include regex.")
	}

	return regexp
}

func parseCutOffDateTime(olderThan string) time.Time {
	if len(olderThan) > 0 {
		cutOffDateTime, err := tparse.ParseNow(time.RFC3339, "now-" + olderThan)
		if err != nil {
			log.WithError(err).
				WithField("--older-than", olderThan).
				Fatal("Could not parse --older-than flag.")
		}
		return cutOffDateTime;
	}

	return time.Now()
}
