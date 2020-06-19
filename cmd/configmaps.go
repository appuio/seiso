package cmd

import (
	"fmt"
	"github.com/appuio/seiso/pkg/kubernetes"
	"github.com/appuio/seiso/pkg/resource"

	"github.com/appuio/seiso/cfg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	configMapCommandLongDescription = `Sometimes ConfigMaps are left unused in the Kubernetes cluster.
This command deletes ConfigMaps that are not being used anymore.`
)

var (
	// configMapCmd represents a cobra command to clean up unused ConfigMaps
	configMapCmd = &cobra.Command{
		Use:          "configmaps",
		Short:        "Cleans up your unused ConfigMaps in the Kubernetes cluster",
		Long:         configMapCommandLongDescription,
		Aliases:      []string{"configmap", "cm"},
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
			if err := validateConfigMapCommandInput(args); err != nil {
				cmd.Usage()
				return err
			}
			client, err := kubernetes.Init().NewCoreV1Client()
			if err != nil {
				return err
			}
			resources := resource.ConfigMaps{
				&resource.GenericResources{
					Client: client,
					Namespace: config.Namespace,
				},
			}
			return executeConfigMapCleanupCommand(resources, config)
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

func validateConfigMapCommandInput(args []string) error {

	if _, err := parseCutOffDateTime(config.Resource.OlderThan); err != nil {
		return fmt.Errorf("Could not parse older-than flag: %w", err)
	}
	return nil
}

func executeConfigMapCleanupCommand(resources resource.ConfigMapResources, config *cfg.Configuration) error {
	namespace := resources.GetNamespace()
	labels := config.Resource.Labels
	if len(labels) == 0 {
		configMapNames, labels, err := resources.ListConfigMaps()
		if err != nil {
			return err
		}
		log.WithFields(log.Fields{
			"\n - namespace":    namespace,
			"\n - 🔓 configMaps": configMapNames,
			"\n - 🎫 labels":     labels,
		}).Info("Please use labels to select ConfigMaps. The following ConfigMaps and Labels are available:")
		return nil
	}

	log.WithField("namespace", namespace).Debug("Looking for ConfigMaps")

	err := resources.LoadConfigMaps(getListOptions(labels))
	if err != nil {
		return fmt.Errorf("could not retrieve ConfigMaps with labels '%s' for '%s': %w", labels, namespace, err)
	}

	err = resources.FilterUsed()
	if err != nil {
		return fmt.Errorf("could not retrieve unused configMaps for '%s': %w", namespace, err)
	}
	olderThan := config.Resource.OlderThan
	keep := config.History.Keep
	cutOffDateTime, _ := parseCutOffDateTime(olderThan)
	resources.FilterByTime(cutOffDateTime)
	resources.FilterByMaxCount(keep)

	if config.Delete {
		err := resources.Delete(func(client kubernetes.CoreV1ClientInt) cfg.CoreObjectInterface {
			return client.ConfigMaps(namespace)
		})
		if err != nil {
			log.WithError(err).Errorf("Failed to complete deletion in namespace %s", namespace)
		}
	} else {
		log.Infof("showing results for --keep=%d and --older-than=%s", keep, olderThan)
		resources.Print(config.Log.Batch)
	}

	return nil
}
