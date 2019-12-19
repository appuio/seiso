package openshift

import (
	"github.com/appuio/image-cleanup/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	resources = []schema.GroupVersionResource{
		schema.GroupVersionResource{Version: "v1", Resource: "pods"},
		schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"},
		schema.GroupVersionResource{Group: "apps.openshift.io", Version: "v1", Resource: "deploymentconfigs"},
		schema.GroupVersionResource{Group: "batch", Version: "v1beta1", Resource: "cronjobs"},
		schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "daemonsets"},
		schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "deployments"},
		schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "replicasets"},
	}
)

// GetActiveImageStreamTags retrieves the image streams tags referenced in some Kubernetes resources
func GetActiveImageStreamTags(namespace, imageStream string, imageStreamTags []string) ([]string, error) {
	var activeImageStreamTags []string

	for _, resource := range resources {
		for _, imageStreamTag := range imageStreamTags {
			image := BuildImageStreamTagName(imageStream, imageStreamTag)

			contains, err := kubernetes.ResourceContains(namespace, image, resource)
			if err != nil {
				return nil, err
			}

			if contains {
				activeImageStreamTags = append(activeImageStreamTags, imageStreamTag)
			}
		}
	}

	return activeImageStreamTags, nil
}

// GetImageStreams returns the image streams for a namespace
func GetImageStreams(namespace string) ([]string, error) {
	var imageStreams []string

	imageClient, err := NewImageV1Client()
	if err != nil {
		return nil, err
	}

	imageStreamList, err := imageClient.ImageStreams(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	imageStreams = make([]string, len(imageStreamList.Items))

	for i, imageStream := range imageStreamList.Items {
		imageStreams[i] = imageStream.Name
	}

	return imageStreams, nil
}

// GetImageStreamTags returns the tags of an image stream
func GetImageStreamTags(namespace, imageStreamName string) ([]string, error) {
	var imageStreamTags []string

	imageClient, err := NewImageV1Client()
	if err != nil {
		return nil, err
	}

	imageStream, err := imageClient.ImageStreams(namespace).Get(imageStreamName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	imageStreamTags = make([]string, len(imageStream.Status.Tags))

	for i, imageStreamTag := range imageStream.Status.Tags {
		imageStreamTags[i] = imageStreamTag.Tag
	}

	return imageStreamTags, nil
}

// DeleteImageStreamTag deletes the image stream tag
func DeleteImageStreamTag(namespace, name string) error {
	imageclient, err := NewImageV1Client()
	if err != nil {
		return err
	}

	return imageclient.ImageStreamTags(namespace).Delete(name, &metav1.DeleteOptions{})
}

// BuildImageStreamTagName combines a name of an image stream and a tag
func BuildImageStreamTagName(imageStream string, imageStreamTag string) string {
	return imageStream + ":" + imageStreamTag
}
