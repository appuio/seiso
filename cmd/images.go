package cmd

import (
	"github.com/spf13/cobra"
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
