package cmd

import (
	"fmt"

	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/cleanup"
	"github.com/appuio/seiso/pkg/openshift"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	configMapCommandLongDescription = `Sometimes ConfigMaps are left unused in the Kubernetes cluster.
This command deletes ConfigMaps that are not being used anymore.`
)

var configMapLog *log.Entry

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
			return executeConfigMapCleanupCommand(args)
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

func executeConfigMapCleanupCommand(args []string) error {
	if len(config.Resource.Labels) == 0 {
		if err := listConfigMaps(); err != nil {
			return err
		}
		return nil
	}

	c := config.Resource
	namespace := config.Namespace

	log.WithField("namespace", namespace).Debug("Looking for ConfigMaps")

	foundConfigMaps, err := openshift.ListConfigMaps(namespace, getListOptions(c.Labels))
	if err != nil {
		return fmt.Errorf("Could not retrieve ConfigMaps with labels '%s' for '%s': %w", c.Labels, namespace, err)
	}

	unusedConfigMaps, err := openshift.ListUnusedResources(namespace, foundConfigMaps)
	if err != nil {
		return fmt.Errorf("Could not retrieve unused ConfigMaps for '%s': %w", namespace, err)
	}

	cutOffDateTime, _ := parseCutOffDateTime(c.OlderThan)
	filteredConfigMaps := cleanup.FilterResourcesByTime(unusedConfigMaps, cutOffDateTime)
	filteredConfigMaps = cleanup.FilterResourcesByMaxCount(filteredConfigMaps, config.History.Keep)

	if config.Delete {
		DeleteResources(
			filteredConfigMaps,
			func(client *core.CoreV1Client) cfg.CoreObjectInterface {
				return client.ConfigMaps(namespace)
			})
	} else {
		log.Infof("Showing results for --keep=%d and --older-than=%s", config.History.Keep, c.OlderThan)
		PrintResources(filteredConfigMaps, namespace)
	}

	return nil
}
