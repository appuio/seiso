package openshift

import (
	log "github.com/sirupsen/logrus"

	"github.com/appuio/image-cleanup/kubernetes"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewImageStreamCommand creates a cobra command to print imagestreams of a namespace
func NewImageStreamCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "imagestream",
		Short: "Print imagestreams from namespace",
		Long:  `tbd`,
		Run:   printImageStreamsFromNamespace,
	}

	return cmd
}

func printImageStreamsFromNamespace(cmd *cobra.Command, args []string) {
	imageClient := NewImageV1Client()

	imageStreams, err := imageClient.ImageStreams(kubernetes.Namespace()).List(metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Fatal("Could not retrieve list of image streams.")
	}

	for _, imageStream := range imageStreams.Items {
		log.Println(imageStream.ObjectMeta.Name)
	}
}

// GetImageStreamTags retrieves the tags for an image stream
func GetImageStreamTags(imageStreamName string) []string {
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

// DeleteImageStreamTag deletes an image stream tag
func DeleteImageStreamTag(name string) {
	imageclient := NewImageV1Client()

	err := imageclient.ImageStreamTags(kubernetes.Namespace()).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		log.WithError(err).WithField("name", name).Fatal("Could not delete image stream.")
	}
}

// BuildImageStreamTagName builds the name of an image stream tag
func BuildImageStreamTagName(imageStream string, imageStreamTag string) string {
	return imageStream + ":" + imageStreamTag
}
