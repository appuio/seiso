package cleanup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// Contains one value not present in `tags` and one shortend value
	prefixes = []string{
		"0b81a958f590ed7ed8",
		"108f2be974f8e1e5fec8bc759ecf824e81565747",
		"4cb7de27c985216b8888ff6049294dae02f3282e",
		"fa617c0bbf84f09c569870653729aab82766e549",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
	}
	// Contains one value not present in `prefixes` and one extended value
	tags = []string{
		"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
		"108f2be974f8e1e5fec8bc759ecf824e81565747",
		"4cb7de27c985216b8888ff6049294dae02f3282e",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
		"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
	}

	// Contains two values not present in `tags`
	activeTags = []string{
		"108f2be974f8e1e5fec8bc759ecf824e81565747",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
		"fa617c0bbf84f09c569870653729aab82766e549",
		"v3.0.0",
	}
)

func Test_GetTagsMatchingPrefixes(t *testing.T) {
	expected := []string{
		"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
		"108f2be974f8e1e5fec8bc759ecf824e81565747",
		"4cb7de27c985216b8888ff6049294dae02f3282e",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
	}

	matchingTags := GetTagsMatchingPrefixes(prefixes, tags)

	assert.Equal(t, expected, matchingTags)
}

func Test_GetInactiveTags(t *testing.T) {
	expected := []string{
		"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
		"4cb7de27c985216b8888ff6049294dae02f3282e",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
		"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
	}

	inactiveTags := GetInactiveTags(activeTags, tags)

	assert.Equal(t, expected, inactiveTags)
}

func Test_LimitTags(t *testing.T) {
	expected := []string{
		"4cb7de27c985216b8888ff6049294dae02f3282e",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
		"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
		"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
	}

	limitedTags := LimitTags(tags, 2)

	assert.Equal(t, expected, limitedTags)
}
