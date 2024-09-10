package git

import (
	"fmt"
	"io"
	"strings"

	"github.com/appuio/seiso/cfg"

	"github.com/go-git/go-git/v5"
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

// GetGitCandidateList returns either git tags or git commit SHAs
func GetGitCandidateList(o *cfg.GitConfig) ([]string, error) {
	if o.Tag {
		candidates, err := GetTags(o.RepoPath, o.CommitLimit, SortOption(o.SortCriteria))
		if err != nil {
			return []string{}, fmt.Errorf("retrieving commit tags failed: %w", err)
		}
		return candidates, nil
	}
	candidates, err := GetCommitHashes(o.RepoPath, o.CommitLimit)
	if err != nil {
		return []string{}, fmt.Errorf("retrieving commit hashes failed: %w", err)
	}
	return candidates, nil

}
