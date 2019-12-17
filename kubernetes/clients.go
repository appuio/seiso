package kubernetes

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	apps "k8s.io/client-go/kubernetes/typed/apps/v1"
	batch "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
)

// NewAppsV1Client for current kubeconfig
func NewAppsV1Client() *apps.AppsV1Client {
	return apps.NewForConfigOrDie(RestConfig())
}

// NewBatchV1beta1Client creates a new BatchV1beta1Client for the current kubeconfig
func NewBatchV1beta1Client() *batch.BatchV1beta1Client {
	return batch.NewForConfigOrDie(RestConfig())
}

// NewDynamicClient creates a new dynamic client
func NewDynamicClient() dynamic.Interface {
	dynamicInterface, err := dynamic.NewForConfig(RestConfig())
	if err != nil {
		log.WithError(err).Fatal("Could not create dynamic client for rest config.")
	}

	return dynamicInterface
}
