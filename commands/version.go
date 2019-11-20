package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	Version string
	Commit string
	Date string
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and exit",
	Long:  `Print the version and exit`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("version =", Version)
		fmt.Println("commit =", Commit)
		fmt.Println("date =", Date)
		os.Exit(0)
	},
}
