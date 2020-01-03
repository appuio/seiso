package git

import (
	"io"
	"sort"
	"strings"

	"gopkg.in/src-d/go-git.v4"
)

// GetCommitHashes returns the commit hashes of a given repository ordered by the `git.LogOrderCommitterTime`. If `commitLimit` is -1 all commits will be returned.
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

	for i := 0; i < commitLimit || commitLimit < 0; i++ {
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

// GetCommitTags returns the commit tags of a given repository ordered by the creation date. If `commitLimit` is -1 all tags will be returned.
func GetCommitTags(repoPath string, tagLimit int) ([]string, error) {
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

	for i := 0; i < tagLimit || tagLimit < 0; i++ {
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

	sort.Sort(sort.Reverse(sort.StringSlice(commitTags)))

	return commitTags, nil
}
