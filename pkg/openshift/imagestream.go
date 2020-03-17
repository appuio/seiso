package openshift

import (
	"github.com/appuio/image-cleanup/pkg/kubernetes"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/thoas/go-funk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	resources = []schema.GroupVersionResource{
		{Version: "v1", Resource: "pods"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps.openshift.io", Version: "v1", Resource: "deploymentconfigs"},
		{Group: "batch", Version: "v1beta1", Resource: "cronjobs"},
		{Group: "extensions", Version: "v1beta1", Resource: "daemonsets"},
		{Group: "extensions", Version: "v1beta1", Resource: "deployments"},
		{Group: "extensions", Version: "v1beta1", Resource: "replicasets"},
	}
	helper = kubernetes.New()
)

// GetActiveImageStreamTags retrieves the image streams tags referenced in some Kubernetes resources
func GetActiveImageStreamTags(namespace, imageStream string, imageStreamTags []string) (activeImageStreamTags []string, funcError error) {
	funk.ForEach(resources, func(resource schema.GroupVersionResource) {
		funk.ForEach(imageStreamTags, func(imageStreamTag string) {
			if funk.ContainsString(activeImageStreamTags, imageStreamTag) {
				// already marked as existing, skip this
				return
			}
			image := BuildImageStreamTagName(imageStream, imageStreamTag)
			contains, err := helper.ResourceContains(namespace, image, resource)
			if err != nil {
				funcError = err
				return
			}

			if contains {
				activeImageStreamTags = append(activeImageStreamTags, imageStreamTag)
			}
		})
	})
	return activeImageStreamTags, funcError
}

// GetImageStreamTags returns the tags of an image stream older than the specified time
func GetImageStreamTags(namespace, imageStreamName string) ([]imagev1.NamedTagEventList, error) {

	imageClient, err := NewImageV1Client()
	if err != nil {
		return nil, err
	}

	imageStream, err := imageClient.ImageStreams(namespace).Get(imageStreamName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return imageStream.Status.Tags, nil
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
