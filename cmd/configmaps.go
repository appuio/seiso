package cmd

import (
	"fmt"
	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/configmap"
	"github.com/appuio/seiso/pkg/kubernetes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	configMapCommandLongDescription = `Sometimes ConfigMaps are left unused in the Kubernetes cluster.
This command deletes ConfigMaps that are not being used anymore.`
)

var (
	configMapCmd = &cobra.Command{
		Use:          "configmaps",
		Short:        "Cleans up your unused ConfigMaps in the Kubernetes cluster",
		Long:         configMapCommandLongDescription,
		Aliases:      []string{"configmap", "cm"},
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		PreRunE:      validateConfigMapCommandInput,
		RunE:         executeConfigMapCleanupCommand,
	}
)

func init() {
	rootCmd.AddCommand(configMapCmd)
	defaults := cfg.NewDefaultConfig()

	configMapCmd.PersistentFlags().BoolP("delete", "d", defaults.Delete, "Effectively delete ConfigMaps found")
	configMapCmd.PersistentFlags().StringSliceP("label", "l", defaults.Resource.Labels,
		"Identify the ConfigMap by these \"key=value\" labels")
	configMapCmd.PersistentFlags().IntP("keep", "k", defaults.History.Keep,
		"Keep most current <k> ConfigMaps; does not include currently used ConfigMaps (if detected)")
	configMapCmd.PersistentFlags().String("older-than", defaults.Resource.OlderThan,
		"Delete ConfigMaps that are older than the duration, e.g. [1y2mo3w4d5h6m7s]")
}

func validateConfigMapCommandInput(cmd *cobra.Command, args []string) (returnErr error) {
	defer showUsageOnError(cmd, returnErr)
	if len(config.Resource.Labels) == 0 {
		return missingLabelSelectorError(config.Namespace, "configmaps")
	}
	for _, label := range config.Resource.Labels {
		if !strings.Contains(label, "=") {
			return fmt.Errorf("incorrect label format does not match expected \"key=value\" format: %s", label)
		}
	}
	if _, err := parseCutOffDateTime(config.Resource.OlderThan); err != nil {
		return fmt.Errorf("could not parse older-than flag: %w", err)
	}
	return nil
}

func executeConfigMapCleanupCommand(cmd *cobra.Command, args []string) error {
	coreClient, err := kubernetes.NewCoreV1Client()
	if err != nil {
		return fmt.Errorf("cannot initiate kubernetes client: %w", err)
	}

	c := config.Resource
	namespace := config.Namespace
	service := configmap.NewConfigMapsService(
		coreClient.ConfigMaps(namespace),
		kubernetes.New(),
		configmap.ServiceConfiguration{Batch: config.Log.Batch})

	log.WithField("namespace", namespace).Debug("Getting ConfigMaps")
	foundConfigMaps, err := service.List(toListOptions(c.Labels))
	if err != nil {
		return fmt.Errorf("could not retrieve ConfigMaps with labels '%s' for '%s': %w", c.Labels, namespace, err)
	}

	unusedConfigMaps, err := service.GetUnused(namespace, foundConfigMaps)
	if err != nil {
		return fmt.Errorf("could not retrieve unused config maps for '%s': %w", namespace, err)
	}

	cutOffDateTime, _ := parseCutOffDateTime(c.OlderThan)
	filteredConfigMaps := service.FilterByTime(unusedConfigMaps, cutOffDateTime)
	filteredConfigMaps = service.FilterByMaxCount(filteredConfigMaps, config.History.Keep)

	if config.Delete {
		err := service.Delete(filteredConfigMaps)
		if err != nil {
			return fmt.Errorf("could not delete ConfigMaps for '%s': %s", namespace, err)
		}
	} else {
		log.WithFields(log.Fields{
			"namespace":  namespace,
			"keep":       config.History.Keep,
			"older_than": c.OlderThan,
		}).Info("Showing results")
		service.Print(filteredConfigMaps)
	}

	return nil
}
