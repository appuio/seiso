package cmd

import (
	"context"
	"fmt"
	"strings"

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
	secretCmd = &cobra.Command{
		Use:          "secrets",
		Short:        "Cleans up your unused Secrets in the Kubernetes cluster",
		Long:         secretCommandLongDescription,
		Aliases:      []string{"secret"},
		SilenceUsage: true,
		PreRunE:      validateSecretCommandInput,
		RunE:         executeSecretCleanupCommand,
	}
)

func init() {
	rootCmd.AddCommand(secretCmd)
	defaults := cfg.NewDefaultConfig()

	secretCmd.PersistentFlags().BoolP("delete", "d", defaults.Delete, "Effectively delete Secrets found")
	secretCmd.PersistentFlags().StringSliceP("label", "l", defaults.Resource.Labels,
		"Identify the Secrets by these \"key=value\" labels")
	secretCmd.PersistentFlags().IntP("keep", "k", defaults.History.Keep,
		"Keep most current <k> Secrets; does not include currently used secret (if detected)")
	secretCmd.PersistentFlags().String("older-than", defaults.Resource.OlderThan,
		"Delete Secrets that are older than the duration, e.g. [1y2mo3w4d5h6m7s]")
}

func validateSecretCommandInput(cmd *cobra.Command, args []string) (returnErr error) {
	defer showUsageOnError(cmd, returnErr)
	if len(config.Resource.Labels) == 0 {
		return missingLabelSelectorError(config.Namespace, "secrets")
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

func executeSecretCleanupCommand(cmd *cobra.Command, args []string) error {
	coreClient, err := kubernetes.NewCoreV1Client()
	if err != nil {
		return fmt.Errorf("cannot initiate kubernetes client: %w", err)
	}

	ctx := context.Background()
	c := config.Resource
	namespace := config.Namespace
	service := secret.NewSecretsService(
		coreClient.Secrets(namespace),
		kubernetes.New(),
		secret.ServiceConfiguration{Batch: config.Log.Batch})

	log.WithField("namespace", namespace).Debug("Getting Secrets")
	foundSecrets, err := service.List(ctx, toListOptions(c.Labels))
	if err != nil {
		return fmt.Errorf("could not retrieve Secrets with labels '%s' for '%s': %w", c.Labels, namespace, err)
	}

	unusedSecrets, err := service.GetUnused(ctx, namespace, foundSecrets)
	if err != nil {
		return fmt.Errorf("could not retrieve unused Secrets for '%s': %w", namespace, err)
	}

	cutOffDateTime, _ := parseCutOffDateTime(c.OlderThan)

	filteredSecrets := service.FilterByTime(unusedSecrets, cutOffDateTime)
	filteredSecrets = service.FilterByMaxCount(filteredSecrets, config.History.Keep)

	if config.Delete {
		err := service.Delete(ctx, filteredSecrets)
		if err != nil {
			return fmt.Errorf("could not delete Secrets for '%s': %s", namespace, err)
		}
	} else {
		log.WithFields(log.Fields{
			"namespace":  namespace,
			"keep":       config.History.Keep,
			"older_than": c.OlderThan,
		}).Info("Showing results")
		service.Print(filteredSecrets)
	}

	return nil
}
