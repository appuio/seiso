package cmd

import (
	"fmt"
	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/configmap"
	"github.com/appuio/seiso/pkg/kubernetes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
			if err := validateConfigMapCommandInput(); err != nil {
				cmd.Usage()
				return err
			}

			coreClient, err := kubernetes.NewCoreV1Client()
			if err != nil {
				return fmt.Errorf("cannot initiate kubernetes core client")
			}

			configMapService := configmap.NewConfigMapsService(
				coreClient.ConfigMaps(config.Namespace),
				kubernetes.New(),
				configmap.ServiceConfiguration{Batch: config.Log.Batch})
			return executeConfigMapCleanupCommand(configMapService)
		},
	}
)

func init() {
	rootCmd.AddCommand(configMapCmd)
	defaults := cfg.NewDefaultConfig()

	configMapCmd.PersistentFlags().BoolP("delete", "d", defaults.Delete, "Effectively delete ConfigMaps found")
	configMapCmd.PersistentFlags().StringSliceP("label", "l", defaults.Resource.Labels, "Identify the ConfigMap by these labels")
	configMapCmd.PersistentFlags().IntP("keep", "k", defaults.History.Keep,
		"Keep most current <k> ConfigMaps; does not include currently used ConfigMaps (if detected)")
	configMapCmd.PersistentFlags().String("older-than", defaults.Resource.OlderThan,
		"Delete ConfigMaps that are older than the duration, e.g. [1y2mo3w4d5h6m7s]")
}

func validateConfigMapCommandInput() error {
	if len(config.Resource.Labels) == 0 {
		return missingLabelSelectorError(config.Namespace, "configmaps")
	}
	if _, err := parseCutOffDateTime(config.Resource.OlderThan); err != nil {
		return fmt.Errorf("could not parse older-than flag: %w", err)
	}
	return nil
}

func executeConfigMapCleanupCommand(service configmap.Service) error {
	c := config.Resource
	namespace := config.Namespace

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
