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

func Test_GetCommitTagsSortedInAlphabeticalOrder(t *testing.T) {
	commitLimit := 2
	commitHashes, err := GetCommitTags("../..", commitLimit, "alphabetic") // Open repository from root dir

	expectInOrder := []string{"0.0.1", "v0.1.0"}

	assert.NoError(t, err)
	assert.Len(t, commitHashes, commitLimit)
	assert.EqualValues(t, commitHashes, expectInOrder)
}

func Test_GetCommitTagsSortedByVersion(t *testing.T) {
	commitLimit := 2
	commitHashes, err := GetCommitTags("../..", commitLimit, "version") // Open repository from root dir

	expectInOrder := []string{"v0.1.0", "0.0.1"}

	assert.NoError(t, err)
	assert.Len(t, commitHashes, commitLimit)
	assert.EqualValues(t, commitHashes, expectInOrder)
}

func Test_GetCommitAllTags(t *testing.T) {
	commitLimit := -1
	_, err := GetCommitTags("../..", commitLimit, "version") // Open repository from root dir

	assert.NoError(t, err)
}

func Test_GetCommitTagsFail(t *testing.T) {
	commitLimit := 2
	_, err := GetCommitTags("not-a-repo", commitLimit, "version")

	assert.Error(t, err)
}

func Test_SortByVersion(t *testing.T) {

	unsortedTags := []string{"v3.0.1", "0.3", "v2.1.1", "0.0.1", "v5.0.2", "4.0.1-beta", "v3.0.0-alpha", "v3", "0.0.2", "v0.2.0", "3.0.0", "random"}
	expectedSortedTags := []string{"v5.0.2", "4.0.1-beta", "v3.0.1", "v3", "3.0.0", "v3.0.0-alpha", "v2.1.1", "0.3", "v0.2.0", "0.0.2", "0.0.1"}

	sortedTags, err := Sort(unsortedTags, Version)

	assert.NoError(t, err)
	assert.EqualValues(t, expectedSortedTags, sortedTags)
}

func Test_SortBInAlphabeticalOrder(t *testing.T) {

	unsortedTags := []string{"v3.0.1", "0.3", "v2.1.1", "0.0.1", "v5.0.2", "4.0.1-beta", "v3.0.0-alpha", "v3", "0.0.2", "v0.2.0", "3.0.0", "random"}
	expectedSortedTags := []string{"0.0.1", "0.0.2", "0.3", "3.0.0", "4.0.1-beta", "random", "v0.2.0", "v2.1.1", "v3", "v3.0.0-alpha", "v3.0.1", "v5.0.2"}

	sortedTags, err := Sort(unsortedTags, Alphabetic)

	assert.NoError(t, err)
	assert.EqualValues(t, expectedSortedTags, sortedTags)
}
