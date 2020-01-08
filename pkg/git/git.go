package git

import (
	"errors"
	"io"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
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

// GetCommitTags returns the commit tags of a given repository ordered alphabetically or by version. If `commitLimit` is -1 all tags will be returned.
func GetCommitTags(repoPath string, tagLimit int, sortTagBy SortTagBy) ([]string, error) {
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

	return Sort(commitTags, sortTagBy)
}

// Sort function sorts the slice according to the sort type
func Sort(tags []string, sortTagBy SortTagBy) ([]string, error) {
	switch sortTagBy {

	case Version:
		var versionTags []*version.Version
		for _, raw := range tags {
			version, err := version.NewVersion(raw)
			if err != nil {
				log.WithError(err).WithField("tag", raw).Warn("Skipped invalid version")
			} else {
				versionTags = append(versionTags, version)
			}
		}

		sort.Sort(sort.Reverse(version.Collection(versionTags)))

		sortedTags := make([]string, len(versionTags))
		for i, sortedVersion := range versionTags {
			sortedTags[i] = sortedVersion.Original()
		}

		return sortedTags, nil

	case Alphabetic:
		sort.Strings(tags)
		return tags, nil

	default:
		return nil, errors.New("Undefined sort type")
	}
}
