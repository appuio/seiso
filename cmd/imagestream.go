package cmd

import (
	log "github.com/sirupsen/logrus"

	"github.com/appuio/image-cleanup/pkg/cleanup"
	"github.com/appuio/image-cleanup/pkg/git"
	"github.com/appuio/image-cleanup/pkg/kubernetes"
	"github.com/appuio/image-cleanup/pkg/openshift"
	"github.com/spf13/cobra"
)

// ImageStreamCleanupOptions is a struct to support the cleanup command
type ImageStreamCleanupOptions struct {
	Force       bool
	CommitLimit int
	RepoPath    string
	Keep        int
	ImageStream string
	Namespace   string
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
	cmd.Flags().IntVarP(&o.CommitLimit, "git-commit-limit", "l", 100, "only look at the first <n> commits to compare with tags or use -1 for all commits")
	cmd.Flags().StringVarP(&o.RepoPath, "git-repo-path", "p", ".", "absolute path to Git repository (for current dir use: $PWD)")
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().IntVarP(&o.Keep, "keep", "k", 10, "keep most current <n> images")
	return cmd
}

func (o *ImageStreamCleanupOptions) cleanupImageStreamTags(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		o.ImageStream = args[0]
	}
	
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
			log.WithError(err).WithField("namespace", o.Namespace).Fatal("Could not retrieve image streams.")
		}

		log.Printf("No image stream provided as argument. Available image streams for namespace %s: %s", o.Namespace, imageStreams)

		return
	}

	commitHashes, err := git.GetCommitHashes(o.RepoPath, o.CommitLimit)
	if err != nil {
		log.WithError(err).WithField("RepoPath", o.RepoPath).WithField("CommitLimit", o.CommitLimit).Fatal("Retrieving commit hashes failed.")
	}

	imageStreamTags, err := openshift.GetImageStreamTags(o.Namespace, o.ImageStream)
	if err != nil {
		log.WithError(err).WithField("Namespace", o.Namespace).WithField("ImageStream", o.ImageStream).Fatal("Could not retrieve image stream.")
	}

	matchingTags := cleanup.GetTagsMatchingPrefixes(&commitHashes, &imageStreamTags)

	activeImageStreamTags, err := openshift.GetActiveImageStreamTags(o.Namespace, o.ImageStream, imageStreamTags)
	if err != nil {
		log.WithError(err).WithField("Namespace", o.Namespace).WithField("ImageStream", o.ImageStream).WithField("imageStreamTags", imageStreamTags).Fatal("Could not retrieve active image stream tags.")
	}

	inactiveTags := cleanup.GetInactiveTags(&activeImageStreamTags, &matchingTags)

	inactiveTags = cleanup.LimitTags(&inactiveTags, o.Keep)

	log.Printf("Tags for deletion: %s", inactiveTags)

	if o.Force {
		for _, inactiveTag := range inactiveTags {
			openshift.DeleteImageStreamTag(o.Namespace, openshift.BuildImageStreamTagName(o.ImageStream, inactiveTag))
			log.Printf("Deleted image stream tag: %s", inactiveTag)
		}
	} else {
		log.Println("--force was not specified. Nothing has been deleted.")
	}
}
