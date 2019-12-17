package openshift

import (
	"github.com/appuio/image-cleanup/kubernetes"
	image "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
)

// NewImageV1Client for current kubeconfig
func NewImageV1Client() *image.ImageV1Client {
	return image.NewForConfigOrDie(kubernetes.RestConfig())
}
