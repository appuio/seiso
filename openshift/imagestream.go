package openshift

import (
	log "github.com/sirupsen/logrus"

	"github.com/appuio/image-cleanup/cleanup"
	"github.com/appuio/image-cleanup/git"
	"github.com/appuio/image-cleanup/kubernetes"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	resources = []schema.GroupVersionResource{
		schema.GroupVersionResource{Version: "v1", Resource: "pods"},
		schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"},
		schema.GroupVersionResource{Group: "apps.openshift.io", Version: "v1", Resource: "deploymentconfigs"},
		schema.GroupVersionResource{Group: "batch", Version: "v1beta1", Resource: "cronjobs"},
		schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "deployments"},
		schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "replicasets"},
	}
)

// Options is a struct to support the cleanup command
type Options struct {
	Force       bool
	CommitLimit int
	RepoPath    string
	Keep        int
	ImageStream string
}

// NewImageStreamCleanupCommand creates a cobra command to clean up an imagestream based on commits
func NewImageStreamCleanupCommand() *cobra.Command {
	o := Options{}
	cmd := &cobra.Command{
		Use:     "imagestream",
		Aliases: []string{"is"},
		Short:   "Clean up excessive image tags",
		Long:    `Clean up excessive image tags matching the commit hashes (prefix) of the git repository`,
		Run:     o.cleanupImageStreamTags,
	}
	cmd.Flags().BoolVarP(&o.Force, "force", "f", false, "delete image stream tags")
	cmd.Flags().IntVarP(&o.CommitLimit, "git-commit-limit", "l", 100, "only look at the first <n> commits to compare with tags")
	cmd.Flags().StringVarP(&o.RepoPath, "git-repo-path", "p", ".", "absolute path to Git repository (for current dir use: $PWD)")
	cmd.Flags().IntVarP(&o.Keep, "keep", "k", 10, "keep most current <n> images")
	return cmd
}

func (o *Options) cleanupImageStreamTags(cmd *cobra.Command, args []string) {
	o.ImageStream = args[0]

	commitHashes := git.GetCommitHashes(o.RepoPath, o.CommitLimit)

	imageStreamTags := getImageStreamTags(o.ImageStream)

	matchingTags := cleanup.GetTagsMatchingPrefixes(commitHashes, imageStreamTags)

	activeImageStreamTags := getActiveImageStreamTags(o.ImageStream, imageStreamTags)

	inactiveTags := cleanup.GetInactiveTags(activeImageStreamTags, matchingTags)

	log.Printf("Tags for deletion: %s", inactiveTags)

	if o.Force {
		for _, inactiveTag := range inactiveTags {
			deleteImageStreamTag(buildImageStreamTagName(o.ImageStream, inactiveTag))
			log.Printf("Deleted image stream tag: %s", inactiveTag)
		}
	}
}

func getActiveImageStreamTags(imageStream string, imageStreamTags []string) []string {
	var activeImageStreamTags []string

	for _, resource := range resources {
		for _, imageStreamTag := range imageStreamTags {
			image := buildImageStreamTagName(imageStream, imageStreamTag)

			if kubernetes.ResourceContains(image, resource) {
				activeImageStreamTags = append(activeImageStreamTags, imageStreamTag)
			}
		}
	}

	return activeImageStreamTags
}

func getImageStreamTags(imageStreamName string) []string {
	var imageStreamTags []string

	imageClient := NewImageV1Client()

	imageStream, err := imageClient.ImageStreams(kubernetes.Namespace()).Get(imageStreamName, metav1.GetOptions{})
	if err != nil {
		log.WithError(err).WithField("imageStreamName", imageStreamName).Fatal("Could not retrieve image stream.")
	}

	for _, imageStreamTag := range imageStream.Status.Tags {
		imageStreamTags = append(imageStreamTags, imageStreamTag.Tag)
	}

	return imageStreamTags
}

func deleteImageStreamTag(name string) {
	imageclient := NewImageV1Client()

	err := imageclient.ImageStreamTags(kubernetes.Namespace()).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		log.WithError(err).WithField("name", name).Fatal("Could not delete image stream.")
	}
}

func buildImageStreamTagName(imageStream string, imageStreamTag string) string {
	return imageStream + ":" + imageStreamTag
}
