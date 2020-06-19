package cmd

import (
	"fmt"
	"github.com/appuio/seiso/pkg/kubernetes"

	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/resource"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	secretCommandLongDescription = `Sometimes secrets are left unused in the Kubernetes cluster.
This command deletes secrets that are not being used anymore.`
)

var (
	// secretCmd represents a cobra command to clean up unused secrets
	secretCmd = &cobra.Command{
		Use:          "secrets",
		Short:        "Cleans up your unused secrets in the Kubernetes cluster",
		Long:         secretCommandLongDescription,
		Aliases:      []string{"secret"},
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSecretCommandInput(); err != nil {
				cmd.Usage()
				return err
			}
			client, err := kubernetes.Init().NewCoreV1Client()
			if err != nil {
				return err
			}
			resources := resource.Secrets{
				&resource.GenericResources{
					Client: client,
					Namespace: config.Namespace,
				},
			}
			return executeSecretCleanupCommand(resources, config)
		},
	}
)

func init() {
	rootCmd.AddCommand(secretCmd)
	defaults := cfg.NewDefaultConfig()

	secretCmd.PersistentFlags().BoolP("delete", "d", defaults.Delete, "Effectively delete Secrets found")
	secretCmd.PersistentFlags().StringSliceP("label", "l", defaults.Resource.Labels, "Identify the Secret by these labels")
	secretCmd.PersistentFlags().IntP("keep", "k", defaults.History.Keep,
		"Keep most current <k> Secrets; does not include currently used secret (if detected)")
	secretCmd.PersistentFlags().String("older-than", defaults.Resource.OlderThan,
		"Delete Secrets that are older than the duration, e.g. [1y2mo3w4d5h6m7s]")
}

func validateSecretCommandInput() error {
	if _, err := parseCutOffDateTime(config.Resource.OlderThan); err != nil {
		return fmt.Errorf("Could not parse older-than flag: %w", err)
	}
	return nil
}

func executeSecretCleanupCommand(resources resource.SecretResources, config *cfg.Configuration) error {
	namespace := resources.GetNamespace()
	labels := config.Resource.Labels
	if len(labels) == 0 {
		secretNames, labels, err := resources.ListSecrets()
		if err != nil {
			return err
		}
		log.WithFields(log.Fields{
			"\n - namespace": namespace,
			"\n - 🔐 secrets": secretNames,
			"\n - 🎫 labels":  labels,
		}).Info("Please use labels to select Secrets. The following Secrets and Labels are available:")
		return nil
	}

	log.WithField("namespace", namespace).Debug("looking for secrets")

	err := resources.LoadSecrets(getListOptions(labels))
	if err != nil {
		return fmt.Errorf("could not retrieve secrets with labels '%s' for '%s': %w", labels, namespace, err)
	}

	err = resources.FilterUsed()
	if err != nil {
		return fmt.Errorf("could not retrieve unused secrets for '%s': %w", namespace, err)
	}

	olderThan := config.Resource.OlderThan
	keep := config.History.Keep
	cutOffDateTime, _ := parseCutOffDateTime(olderThan)
	resources.FilterByTime(cutOffDateTime)
	resources.FilterByMaxCount(keep)

	if config.Delete {
		err := resources.Delete(func(client kubernetes.CoreV1ClientInt) cfg.CoreObjectInterface {
			return client.Secrets(namespace)
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
