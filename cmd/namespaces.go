package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/kubernetes"
	"github.com/appuio/seiso/pkg/namespace"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	nsCommandLongDescription = `Sometimes Namespaces are left empty in a Kubernetes cluster.
This command deletes Namespaces that are not being used anymore.
A Namespace is deemed empty if no Helm releases, Pods, Deployments, StatefulSets or DaemonSets can be found.`
)

var (
	nsCmd = &cobra.Command{
		Use:          "namespaces",
		Short:        "Cleans up your empty Namespaces",
		Long:         nsCommandLongDescription,
		Aliases:      []string{"namespace", "ns"},
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		PreRunE:      validateNsCommandInput,
		RunE:         executeNsCleanupCommand,
	}
)

func init() {
	rootCmd.AddCommand(nsCmd)
	defaults := cfg.NewDefaultConfig()

	nsCmd.PersistentFlags().BoolP("delete", "d", defaults.Delete, "Effectively delete Namespaces found")
	nsCmd.PersistentFlags().StringSliceP("label", "l", defaults.Resource.Labels,
		"Identify the Namespaces by these \"key=value\" labels")
	nsCmd.PersistentFlags().String("delete-after", defaults.Resource.DeleteAfter,
		"Only delete Namespaces after they were empty for this duration, e.g. [1y2mo3w4d5h6m7s]")
}

func validateNsCommandInput(cmd *cobra.Command, _ []string) (returnErr error) {
	defer showUsageOnError(cmd, returnErr)
	if len(config.Resource.Labels) == 0 {
		return missingLabelSelectorError(config.Namespace, "namespaces")
	}
	for _, label := range config.Resource.Labels {
		if !strings.Contains(label, "=") {
			return fmt.Errorf("incorrect label format does not match expected \"key=value\" format: %s", label)
		}
	}
	if _, err := parseCutOffDateTime(config.Resource.DeleteAfter); err != nil {
		return fmt.Errorf("could not parse delete-after flag %w", err)
	}
	return nil
}

func executeNsCleanupCommand(_ *cobra.Command, _ []string) error {
	coreClient, err := kubernetes.NewCoreV1Client()
	if err != nil {
		return fmt.Errorf("cannot initiate kubernetes client: %w", err)
	}

	dynamicClient, err := kubernetes.NewDynamicClient()
	if err != nil {
		return fmt.Errorf("cannot initiate kubernetes dynamic client: %w", err)
	}

	ctx := context.Background()
	c := config.Resource
	service := namespace.NewNamespacesService(
		coreClient.Namespaces(),
		dynamicClient,
		namespace.ServiceConfiguration{
			Batch: config.Log.Batch,
		})

	log.Debug("Getting Namespaces")
	allNamespaces, err := service.List(ctx, toListOptions(c.Labels))
	if err != nil {
		return fmt.Errorf("could not retrieve Namespaces with labels %q: %w", c.Labels, err)
	}

	emptyNamespaces, err := service.GetEmptyFor(ctx, allNamespaces, c.DeleteAfter)
	if err != nil {
		return fmt.Errorf("could not retrieve empty namespaces %w", err)
	}

	if config.Delete {
		err := service.Delete(ctx, emptyNamespaces)
		if err != nil {
			return fmt.Errorf("could not delete Namespaces %w", err)
		}
	} else {
		log.WithFields(log.Fields{
			"delete_after": c.DeleteAfter,
		}).Info("Showing results")
		service.Print(emptyNamespaces)
	}

	return nil
}
