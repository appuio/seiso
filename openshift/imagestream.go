package openshift

import (
	"fmt"

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
	imageclient := NewImageV1Client()

	imagestreamlist, err := imageclient.ImageStreams(kubernetes.Namespace()).List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, imagestream := range imagestreamlist.Items {
		fmt.Println(imagestream.ObjectMeta.Name)
	}
}

// GetImageStreamTags retrieves the tags for an image stream
func GetImageStreamTags(imageStreamName string) []string {
	var imageStreamTags []string

	imageclient := NewImageV1Client()

	imagestream, err := imageclient.ImageStreams(kubernetes.Namespace()).Get(imageStreamName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	for _, imageStreamTag := range imagestream.Status.Tags {
		imageStreamTags = append(imageStreamTags, imageStreamTag.Tag)
	}

	return imageStreamTags
}

// DeleteImageStreamTag deletes an image stream tag
func DeleteImageStreamTag(name string) {
	imageclient := NewImageV1Client()

	err := imageclient.ImageStreamTags(kubernetes.Namespace()).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}

	return
}

// BuildImageStreamTagName builds the name of an image stream tag
func BuildImageStreamTagName(imageStream string, imageStreamTag string) string {
	return imageStream + ":" + imageStreamTag
}
