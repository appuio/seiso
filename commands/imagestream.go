package commands

import (
	"fmt"

	imagev1client "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	rootCmd.AddCommand(imagestreamCmd)
}

var imagestreamCmd = &cobra.Command{
	Use:   "imagestream",
	Short: "Print imagestreams from namespace",
	Long:  `tbd`,
	Run: printImageStreamsFromNamespace,
}

func printImageStreamsFromNamespace(cmd *cobra.Command, args []string) {
	// Instantiate loader for kubeconfig file.
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	namespace := resolveNamespace(kubeconfig)

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
}

// Get the namespace defined in the kubeconfig
func resolveNamespace(kubeconfig clientcmd.ClientConfig) (namespace string) {
	namespace, _, err := kubeconfig.Namespace()
	if err != nil {
		panic(err)
	}
	return
}
