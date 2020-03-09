package git

import (
	"github.com/appuio/image-cleanup/cfg"
	"github.com/sirupsen/logrus"
	"io"
	"strings"

	"gopkg.in/src-d/go-git.v4"
)

// GetCommitHashes returns the commit hashes of a given repository ordered by the `git.LogOrderCommitterTime`. If `commitLimit` is 0 all commits will be returned.
func GetCommitHashes(repoPath string, commitLimit int) ([]string, error) {
	var commitHashes []string

	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	commitIter, err := repository.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	defer commitIter.Close()
	if err != nil {
		return nil, err
	}

	for i := 0; i < commitLimit || commitLimit == 0; i++ {
		commit, err := commitIter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		commitHashes = append(commitHashes, commit.Hash.String())
	}

	return commitHashes, nil
}

// GetTags returns the commit tags of a given repository ordered alphabetically or by version. If `commitLimit` is 0 all tags will be returned.
func GetTags(repoPath string, tagLimit int, sortTagBy SortOption) ([]string, error) {
	var commitTags []string

	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	tagIter, err := repository.Tags()
	defer tagIter.Close()
	if err != nil {
		return nil, err
	}

	for i := 0; i < tagLimit || tagLimit == 0; i++ {
		tag, err := tagIter.Next()

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		splittedPath := strings.Split(tag.Name().String(), "/")
		tagName := splittedPath[len(splittedPath)-1]

		commitTags = append(commitTags, tagName)
	}

	return sortTags(commitTags, sortTagBy)
}

func GetGitCandidateList(o *cfg.GitConfig) []string {
	logEvent := logrus.WithFields(logrus.Fields{
		"GitRepoPath": o.RepoPath,
		"CommitLimit": o.CommitLimit,
	})
	if o.Tag {
		candidates, err := GetTags(o.RepoPath, o.CommitLimit, SortOption(o.SortCriteria))
		if err != nil {
			logEvent.WithError(err).Fatal("Retrieving commit tags failed.")
		}
		return candidates
	} else {
		candidates, err := GetCommitHashes(o.RepoPath, o.CommitLimit)
		if err != nil {
			logEvent.WithError(err).Fatal("Retrieving commit hashes failed.")
		}
		return candidates
	}
}
