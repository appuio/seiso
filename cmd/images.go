package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"strings"
)

// imagesCmd represents the images command
var imagesCmd = &cobra.Command{
	Use:     "images",
	Short:   "Cleans up your image registry from unused image tags",
	Aliases: []string{"image", "img"},
}

func init() {
	rootCmd.AddCommand(imagesCmd)
}

func splitNamespaceAndImagestream(repo string) (namespace string, image string, err error) {
	if !strings.Contains(repo, "/") {
		namespace = config.Namespace
		image = repo
	} else {
		paths := strings.SplitAfter(repo, "/")
		if len(paths) >= 3 {
			namespace = paths[1]
			image = paths[2]
		} else {
			namespace = paths[0]
			image = paths[1]
		}
	}
	if namespace == "" {
		return "", "", errors.New("missing or invalid namespace")
	}
	if image == "" {
		return "", "", errors.New("missing or invalid image name")
	}
	return strings.TrimSuffix(namespace, "/"), image, nil
}
