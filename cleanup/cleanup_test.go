package cleanup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getDeletionCandidates(t *testing.T) {
	commitHashes := []string{
		"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
		"108f2be974f8e1e5fec8bc759ecf824e81565747",
		"4cb7de27c985216b8888ff6049294dae02f3282e",
		"fa617c0bbf84f09c569870653729aab82766e549",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
		"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
	}
	imageStreamTags := []string{
		"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
		"108f2be974f8e1e5fec8bc759ecf824e81565747",
		"4cb7de27c985216b8888ff6049294dae02f3282e",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
		"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
	}
	keep := 2
	expected := []string{
		"4cb7de27c985216b8888ff6049294dae02f3282e",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
		"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
	}

	deletionCandidates := getDeletionCandidates(commitHashes, imageStreamTags, keep)

	assert.Equal(t, expected, deletionCandidates)
}
