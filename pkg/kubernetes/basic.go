package kubernetes

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// RestConfig from the kubeconfig
func RestConfig() (*rest.Config, error) {
	// Get a rest.Config from the kubeconfig file.  This will be passed into all
	// the client objects we create.
	restConfig, err := kubeconfig().ClientConfig()
	if err != nil {
		return nil, err
	}

	return restConfig, nil
}

// Namespace from the kubeconfig
func Namespace() (string, error) {
	namespace, _, err := kubeconfig().Namespace()
	if err != nil {
		return "", err
	}
	return namespace, nil
}

func kubeconfig() clientcmd.ClientConfig {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
}
