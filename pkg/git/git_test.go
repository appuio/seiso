package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetCommitHashes(t *testing.T) {
	commitLimit := 2
	commitHashes, err := GetCommitHashes("../..", commitLimit) // Open repository from root dir

	assert.NoError(t, err)
	assert.Len(t, commitHashes, commitLimit)
}

func Test_GetCommitHashesAll(t *testing.T) {
	commitLimit := -1
	_, err := GetCommitHashes("../..", commitLimit) // Open repository from root dir

	assert.NoError(t, err)
}

func Test_GetCommitHashesFail(t *testing.T) {
	commitLimit := 2
	_, err := GetCommitHashes("not-a-repo", commitLimit)

	assert.Error(t, err)
}
