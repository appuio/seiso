package cleanup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type GetTagsMatchingPrefixesTestCase struct {
	prefixes, tags, expected []string
}

type GetInactiveTagsTestCase struct {
	tags, activeTags, expected []string
}

type LimitTagsTestCase struct {
	tags, expected []string
	limit          int
}

func Test_GetTagsMatchingPrefixes(t *testing.T) {
	testcases := []GetTagsMatchingPrefixesTestCase{
		GetTagsMatchingPrefixesTestCase{
			prefixes: []string{
				"0b81a958f590ed7ed8",
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"fa617c0bbf84f09c569870653729aab82766e549",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
			},
			tags: []string{
				"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
			},
			expected: []string{
				"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
			},
		},
	}

	for _, testcase := range testcases {
		assert.Equal(t, testcase.expected, GetTagsMatchingPrefixes(&testcase.prefixes, &testcase.tags))
	}
}

func Test_GetInactiveTags(t *testing.T) {
	testcases := []GetInactiveTagsTestCase{
		GetInactiveTagsTestCase{
			tags: []string{
				"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
			},
			activeTags: []string{
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
				"fa617c0bbf84f09c569870653729aab82766e549",
				"v3.0.0",
			},
			expected: []string{
				"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
			},
		},
	}

	for _, testcase := range testcases {
		assert.Equal(t, testcase.expected, GetInactiveTags(&testcase.activeTags, &testcase.tags))
	}
}

func Test_LimitTags(t *testing.T) {
	testcases := []LimitTagsTestCase{
		LimitTagsTestCase{
			tags: []string{
				"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
			},
			limit: 2,
			expected: []string{
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
			},
		},
		LimitTagsTestCase{
			tags: []string{
				"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
			},
			limit:    6,
			expected: []string{},
		},
	}

	for _, testcase := range testcases {
		assert.Equal(t, testcase.expected, LimitTags(&testcase.tags, testcase.limit))
	}
}
