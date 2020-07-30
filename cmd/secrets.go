package cmd

import (
	"fmt"
	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/kubernetes"
	"github.com/appuio/seiso/pkg/secret"
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

			coreClient, err := kubernetes.NewCoreV1Client()
			if err != nil {
				return fmt.Errorf("cannot initiate kubernetes core client")
			}

			secretService := secret.NewSecretsService(
				coreClient.Secrets(config.Namespace),
				kubernetes.New(),
				secret.ServiceConfiguration{Batch: config.Log.Batch})
			return executeSecretCleanupCommand(secretService)
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

func executeSecretCleanupCommand(service secret.Service) error {
	c := config.Resource
	namespace := config.Namespace
	if len(config.Resource.Labels) == 0 {
		err := service.PrintNamesAndLabels(namespace)
		if err != nil {
			return err
		}
		return nil
	}

	log.WithField("namespace", namespace).Debug("Looking for secrets")

	foundSecrets, err := service.List(getListOptions(c.Labels))
	if err != nil {
		return fmt.Errorf("Could not retrieve secrets with labels '%s' for '%s': %w", c.Labels, namespace, err)
	}

	unusedSecrets, err := service.GetUnused(namespace, foundSecrets)
	if err != nil {
		return fmt.Errorf("Could not retrieve unused secrets for '%s': %w", namespace, err)
	}

	cutOffDateTime, _ := parseCutOffDateTime(c.OlderThan)

	filteredSecrets := service.FilterByTime(unusedSecrets, cutOffDateTime)
	filteredSecrets = service.FilterByMaxCount(filteredSecrets, config.History.Keep)

	if config.Delete {
		service.Delete(filteredSecrets)
	} else {
		log.Infof("Showing results for --keep=%d and --older-than=%s", config.History.Keep, c.OlderThan)
		service.Print(filteredSecrets)
	}

	return nil
}
