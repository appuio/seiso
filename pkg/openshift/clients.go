package openshift

import (
	"github.com/appuio/image-cleanup/pkg/kubernetes"
	image "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
)

// NewImageV1Client for current kubeconfig
func NewImageV1Client() (*image.ImageV1Client, error) {
	restConfig, err := kubernetes.RestConfig()
	if err != nil {
		return nil, err
	}

	return image.NewForConfig(restConfig)
}
