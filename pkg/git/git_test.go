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

func Test_GetCommitTags(t *testing.T) {
	commitLimit := 2
	commitHashes, err := GetCommitTags("../..", commitLimit) // Open repository from root dir

	expectInOrder := []string{"v0.1.0", "0.0.1"}

	assert.NoError(t, err)
	assert.Len(t, commitHashes, commitLimit)
	assert.EqualValues(t, commitHashes, expectInOrder)
}

func Test_GetCommitTagsAll(t *testing.T) {
	commitLimit := -1
	_, err := GetCommitTags("../..", commitLimit) // Open repository from root dir

	assert.NoError(t, err)
}

func Test_GetCommitTagsFail(t *testing.T) {
	commitLimit := 2
	_, err := GetCommitTags("not-a-repo", commitLimit)

	assert.Error(t, err)
}
