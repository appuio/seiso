package commands

import (
	"fmt"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/spf13/cobra"
)

var image string

func init() {
	rootCmd.AddCommand(tagCmd)
}

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Print the available tags",
	Long:  `tbd`,
	Args:  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		image := args[0]
		url := "https://registry-1.docker.io/"
		username := "" // anonymous
		password := "" // anonymous
		hub, err := registry.New(url, username, password)

		tags, err := hub.Tags(image)
		if err != nil {
			panic(err)
		}

		for _, tag := range tags {
			fmt.Println(tag)
		}

	},
}
