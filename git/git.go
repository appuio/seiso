package git

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
)

// NewGitCommand creates a new cobra command to print the git head
func NewGitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "Print the git HEAD",
		Long:  `tbd`,
		Args:  cobra.ExactValidArgs(1),
		Run:   printGitHEAD,
	}

	return cmd
}

func printGitHEAD(cmd *cobra.Command, args []string) {
	path := args[0]

	repository, err := git.PlainOpen(path)

	if err != nil {
		log.WithError(err).WithField("path", path).Fatal("Could not open repository")
	}

	fmt.Println(repository.Head())
}

// GetCommitHashes returns the commit hashes of a given repository ordered by the `git.LogOrderCommitterTime`
func GetCommitHashes(repoPath string, commitLimit int) (commits []string, err error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		panic(err)
	}

	commitIter, err := r.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, err
	}

	for i := 0; i < commitLimit; i++ {
		commit, err := commitIter.Next()
		if err != nil {
			break
		}

		commits = append(commits, commit.Hash.String())
	}

	return
}
