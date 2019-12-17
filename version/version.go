package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version string
	Commit  string
	Date    string
)

// NewVersionCommand returns a cobra command to print the CLI version
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version and exit",
		Long:  `Print the version and exit`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("version = %s\ncommit = %s\ndate = %s", Version, Commit, Date)
		},
	}

	return cmd
}
