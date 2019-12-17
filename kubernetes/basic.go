package kubernetes

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// RestConfig from the kubeconfig
func RestConfig() (restConfig *rest.Config) {
	// Get a rest.Config from the kubeconfig file.  This will be passed into all
	// the client objects we create.
	restConfig, err := kubeconfig().ClientConfig()
	if err != nil {
		panic(err)
	}

	return
}

// Namespace from the kubeconfig
func Namespace() (namespace string) {
	namespace, _, err := kubeconfig().Namespace()
	if err != nil {
		panic(err)
	}
	return
}

func kubeconfig() clientcmd.ClientConfig {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
}
