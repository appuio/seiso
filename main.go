package main

import (
	"flag"
	"fmt"

	imagev1client "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/heroku/docker-registry-client/registry"
)

func main() {
	// Instantiate loader for kubeconfig file.
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	namespace := resolveNamespace(kubeconfig)

	flag.StringVar(&namespace, "n", namespace, "Namespace to use, defaults to the namespace defined in kubeconfig")
	flag.Parse()

	// Get a rest.Config from the kubeconfig file.  This will be passed into all
	// the client objects we create.
	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		panic(err)
	}

	// Create an OpenShift image/v1 client.
	imageclient, err := imagev1client.NewForConfig(restconfig)
	if err != nil {
		panic(err)
	}

	imagestreamlist, err := imageclient.ImageStreams(namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, imagestream := range imagestreamlist.Items {
		fmt.Println(imagestream.ObjectMeta.Name)
	}

	url := "https://registry-1.docker.io/"
	username := "" // anonymous
	password := "" // anonymous
	hub, err := registry.New(url, username, password)

	tags, err := hub.Tags("appuio/oc")
	if err != nil {
		panic(err)
	}

	for _, tag := range tags {
		fmt.Println(tag)
	}
}

// Get the namespace defined in the kubeconfig
func resolveNamespace(kubeconfig clientcmd.ClientConfig) (namespace string) {
	namespace, _, err := kubeconfig.Namespace()
	if err != nil {
		panic(err)
	}
	return
}
