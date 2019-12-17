package kubernetes

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	log "github.com/sirupsen/logrus"
)

// RestConfig from the kubeconfig
func RestConfig() *rest.Config {
	// Get a rest.Config from the kubeconfig file.  This will be passed into all
	// the client objects we create.
	restConfig, err := kubeconfig().ClientConfig()
	if err != nil {
		log.WithError(err).Fatal("Could not create restConfig from kubeconfig")
	}

	return restConfig
}

// Namespace from the kubeconfig
func Namespace() string {
	namespace, _, err := kubeconfig().Namespace()
	if err != nil {
		log.WithError(err).Fatal("Could not determine namespace from kubeconfig")
	}
	return namespace
}

func kubeconfig() clientcmd.ClientConfig {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
}
