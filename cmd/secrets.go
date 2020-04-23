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
	secretCommandLongDescription = `Sometimes secrets are left unused in the Kubernetes cluster.
This command deletes secrets that are not being used anymore.`
)

var (
	// secretCmd represents a cobra command to clean up unused secrets
	secretCmd = &cobra.Command{
		Use:          "secrets [PROJECT]",
		Short:        "Cleans up your unused secrets in the Kubernetes cluster",
		Long:         secretCommandLongDescription,
		Aliases:      []string{"secret"},
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSecretCommandInput(args); err != nil {
				cmd.Usage()
				return err
			}
			return executeSecretCleanupCommand(args)
		},
	}
)

func init() {
	rootCmd.AddCommand(secretCmd)
	defaults := cfg.NewDefaultConfig()

	secretCmd.PersistentFlags().BoolP("force", "f", defaults.Force, "Confirm deletion of secrets.")
	secretCmd.PersistentFlags().StringSliceP("label", "l", defaults.Resource.Labels, "Identify the secret by these labels")
	secretCmd.PersistentFlags().IntP("keep", "k", defaults.History.Keep,
		"Keep most current <k> secrets. Does not include currently used secret (if detected).")
	secretCmd.PersistentFlags().String("older-than", defaults.Resource.OlderThan,
		"Delete secrets that are older than the duration. Ex.: [1y2mo3w4d5h6m7s]")
}

func validateSecretCommandInput(args []string) error {

	if _, err := parseCutOffDateTime(config.Resource.OlderThan); err != nil {
		return fmt.Errorf("Could not parse older-than flag: %w", err)
	}
	return nil
}

func executeSecretCleanupCommand(args []string) error {
	if len(args) == 0 || len(config.Resource.Labels) == 0 {
		if err := listSecrets(args); err != nil {
			return err
		}
		return nil
	}

	c := config.Resource
	namespace := args[0]

	log.WithField("namespace", namespace).Debug("Looking for secrets")

	foundSecrets, err := openshift.ListSecrets(namespace, getListOptions(c.Labels))
	if err != nil {
		return fmt.Errorf("Could not retrieve secrets with labels '%s' for '%s': %w", c.Labels, namespace, err)
	}

	unusedSecrets, err := openshift.ListUnusedResources(namespace, foundSecrets)
	if err != nil {
		return fmt.Errorf("Could not retrieve unused secrets for '%s': %w", namespace, err)
	}

	if len(unusedSecrets) == 0 {
		log.WithField("namespace", namespace).Info("No unused secret found")
		return nil
	}

	cutOffDateTime, _ := parseCutOffDateTime(c.OlderThan)
	filteredSecrets := cleanup.FilterResourcesByTime(unusedSecrets, cutOffDateTime)
	filteredSecrets = cleanup.FilterResourcesByMaxCount(filteredSecrets, config.History.Keep)

	PrintResources(filteredSecrets)
	DeleteResources(
		filteredSecrets,
		config.Force,
		func(client *core.CoreV1Client) cfg.CoreObjectInterface {
			return client.Secrets(namespace)
		})

	return nil
}
