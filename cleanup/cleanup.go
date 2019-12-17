package cleanup

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/appuio/image-cleanup/git"
	"github.com/appuio/image-cleanup/kubernetes"
	"github.com/appuio/image-cleanup/openshift"
	"github.com/spf13/cobra"
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

// NewCleanupCommand creates a cobra command to clean up an imagestream
func NewCleanupCommand() *cobra.Command {
	o := Options{}
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up excessive image tags",
		Long:  `Clean up excessive (stale) image tags`,
		Run:   o.cleanupImageStreamTags,
	}
	cmd.Flags().BoolVarP(&o.Force, "force", "f", false, "don't ask for confirmation to delete image tags")
	cmd.Flags().IntVarP(&o.CommitLimit, "git-commit-limit", "l", 100, "only look at the first <n> commits to compare with tags")
	cmd.Flags().StringVarP(&o.RepoPath, "git-repo-path", "p", ".", "absolute path to Git repository (for current dir use: $PWD)")
	cmd.Flags().IntVarP(&o.Keep, "keep", "k", 10, "keep most current <n> images")
	return cmd
}

func (o *Options) cleanupImageStreamTags(cmd *cobra.Command, args []string) {
	o.ImageStream = args[0]

	commitHashes := git.GetCommitHashes(o.RepoPath, o.CommitLimit)

	imageStreamTags := openshift.GetImageStreamTags(o.ImageStream)

	deletionCandidates := getDeletionCandidates(commitHashes, imageStreamTags, o.Keep)

	activeImageStreamTags := getActiveImageStreamTags(o.ImageStream, imageStreamTags)

	// Remove the activeImageStreamTags from the deletionCandidates
	for _, activeImageStreamTag := range activeImageStreamTags {
		for i, deletionCandidate := range deletionCandidates {
			if activeImageStreamTag == deletionCandidate {
				deletionCandidates[i] = deletionCandidates[len(deletionCandidates)-1]
				deletionCandidates = deletionCandidates[:len(deletionCandidates)-1]
			}
		}
	}

	log.Printf("Candidates for deletion: %s", deletionCandidates)

	if o.Force {
		for _, deletionCandidate := range deletionCandidates {
			openshift.DeleteImageStreamTag(openshift.BuildImageStreamTagName(o.ImageStream, deletionCandidate))
			log.Printf("Deleted image stream tag: %s", deletionCandidate)
		}
	}
}

func getActiveImageStreamTags(imageStream string, imageStreamTags []string) []string {
	var activeImageStreamTags []string

	for _, resource := range resources {
		for _, imageStreamTag := range imageStreamTags {
			image := openshift.BuildImageStreamTagName(imageStream, imageStreamTag)

			if kubernetes.ResourceContains(image, resource) {
				activeImageStreamTags = append(activeImageStreamTags, imageStreamTag)
			}
		}
	}

	return activeImageStreamTags
}

func getDeletionCandidates(commitHashes []string, imageStreamTags []string, keep int) []string {
	var deletionCandidates []string

	if len(commitHashes) > 0 && len(imageStreamTags) > 0 {
		for _, commitHash := range commitHashes {
			for _, candidate := range imageStreamTags {
				if strings.HasPrefix(candidate, commitHash) {
					deletionCandidates = append(deletionCandidates, candidate)
				}
			}
		}
	}

	if len(deletionCandidates) > keep {
		deletionCandidates = deletionCandidates[keep:]
	} else {
		deletionCandidates = []string{}
	}

	return deletionCandidates
}
