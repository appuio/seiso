package cleanup

import (
	"regexp"
	"testing"
	"time"

	imagev1 "github.com/openshift/api/image/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetMatchingTagsTestCase struct {
	matchValues, tags, expected []string
	matchOption                 MatchOption
}

type GetOrphanTagsTestCase struct {
	matchValues, tags, expected []string
	matchOption                 MatchOption
}

type GetInactiveTagsTestCase struct {
	tags, activeTags, expected []string
}

type LimitTagsTestCase struct {
	tags, expected []string
	limit          int
}

type FilterByRegexTestCase struct {
	tags, expected []string
}

type TagsOlderThanTestCase struct {
	tags      []imagev1.NamedTagEventList
	expected  []string
	olderThan time.Time
}

func Test_GetMatchingTags(t *testing.T) {
	testcases := []GetMatchingTagsTestCase{
		{
			matchOption: MatchOptionPrefix,
			matchValues: []string{
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
		assert.Equal(t, testcase.expected, GetMatchingTags(&testcase.matchValues, &testcase.tags, testcase.matchOption))
	}
}

func Test_GetOrphanTags(t *testing.T) {
	testcases := []GetOrphanTagsTestCase{
		{
			matchValues: []string{},
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
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
			},
			matchOption: MatchOptionPrefix,
		},
		{
			matchValues: []string{
				"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
			},
			tags: []string{
				"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
			},
			expected:    []string{},
			matchOption: MatchOptionPrefix,
		},
		{
			matchValues: []string{
				"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
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
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
			},
			matchOption: MatchOptionPrefix,
		},
		{
			matchValues: []string{
				"3.4",
				"0.0.1",
				"0.0.2",
				"v2.3.0",
			},
			tags: []string{
				"1.0",
				"3.4",
				"v1.0.2",
				"0.0.1",
				"0.0.2",
				"v2.3.0",
			},
			expected: []string{
				"1.0",
				"v1.0.2",
			},
			matchOption: MatchOptionExact,
		},
	}

	for _, testcase := range testcases {
		assert.Equal(t, testcase.expected, FilterOrphanImageTags(&testcase.matchValues, &testcase.tags, testcase.matchOption))
	}
}

func Test_GetInactiveTags(t *testing.T) {
	testcases := []GetInactiveTagsTestCase{
		{
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
		assert.Equal(t, testcase.expected, GetInactiveImageTags(&testcase.activeTags, &testcase.tags))
	}
}

func Test_FilterByRegex(t *testing.T) {
	reg, err := regexp.Compile("^[a-z0-9]{40}$")
	testcases := []FilterByRegexTestCase{
		{
			tags: []string{
				"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
				"v2.0",
				"v2.0-4",
			},
			expected: []string{
				"0b81a958f590ed7ed8be6ec0a2a87816228a482c",
				"108f2be974f8e1e5fec8bc759ecf824e81565747",
				"4cb7de27c985216b8888ff6049294dae02f3282e",
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd",
				"4b35e092ad45a626d9a43b7bc7b03e7f7c3c8037",
			},
		},
	}

	assert.NoError(t, err)
	for _, testcase := range testcases {
		assert.Equal(t, testcase.expected, FilterByRegex(&testcase.tags, reg))
	}
}

func Test_LimitTags(t *testing.T) {
	testcases := []LimitTagsTestCase{
		{
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
		{
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

func Test_TagsOlderThan(t *testing.T) {
	testcases := []TagsOlderThanTestCase{
		{
			tags: []imagev1.NamedTagEventList{
				{
					Tag: "0b81a958f590ed7ed8be6ec0a2a87816228a482c",
					Items: []imagev1.TagEvent{
						{
							Created: metav1.Time{
								Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
							},
						},
						{
							Created: metav1.Time{
								Time: time.Date(2020, 5, 5, 0, 0, 0, 0, time.Local),
							},
						},
					},
				},
				{
					Tag: "108f2be974f8e1e5fec8bc759ecf824e81565747",
					Items: []imagev1.TagEvent{
						{
							Created: metav1.Time{
								Time: time.Date(2020, 4, 4, 0, 0, 0, 0, time.Local),
							},
						},
					},
				},
				{
					Tag: "4cb7de27c985216b8888ff6049294dae02f3282e",
					Items: []imagev1.TagEvent{
						{
							Created: metav1.Time{
								Time: time.Date(2020, 3, 3, 0, 0, 0, 0, time.Local),
							},
						},
					},
				},
				{
					Tag: "c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
					Items: []imagev1.TagEvent{
						{
							Created: metav1.Time{
								Time: time.Date(2020, 2, 2, 0, 0, 0, 0, time.Local),
							},
						},
					},
				},
			},
			expected: []string{
				"c8a693ad89e7069674eda512c553ff56d3ca2ffd-debug",
			},
			olderThan: time.Date(2020, 3, 3, 0, 0, 0, 0, time.Local),
		},
	}

	for _, testcase := range testcases {
		assert.Equal(t, testcase.expected, FilterImageTagsByTime(&testcase.tags, testcase.olderThan))
	}
}

func TestGetInactiveImageTags(t *testing.T) {
	type args struct {
		activeTags   *[]string
		allImageTags *[]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "ShouldFilterOut_ActiveTag",
			args: args{
				activeTags:   &[]string{"active"},
				allImageTags: &[]string{"active", "inactive"},
			},
			want: []string{"inactive"},
		},
		{
			name: "ShouldIgnore",
			args: args{
				activeTags:   &[]string{"active"},
				allImageTags: &[]string{"inactive"},
			},
			want: []string{"inactive"},
		},
		{
			name: "ShouldIgnore_NoActive",
			args: args{
				activeTags:   &[]string{},
				allImageTags: &[]string{"inactive"},
			},
			want: []string{"inactive"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetInactiveImageTags(tt.args.activeTags, tt.args.allImageTags)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFilterOrphanImageTags(t *testing.T) {
	type args struct {
		gitValues   *[]string
		imageTags   *[]string
		matchOption MatchOption
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "ShouldNotFilterAnything_IfEmpty",
			args: args{
				gitValues:   &[]string{},
				imageTags:   &[]string{"a1"},
				matchOption: MatchOptionPrefix,
			},
			want: []string{"a1"},
		},
		{
			name: "ShouldReturn_NotFoundImageTag",
			args: args{
				gitValues:   &[]string{"a1", "a2", "a3"},
				imageTags:   &[]string{"a1", "b2", "b3"},
				matchOption: MatchOptionPrefix,
			},
			want: []string{"b2", "b3"},
		},
		{
			name: "ShouldReturn_Empty_IfNoImageOrphaned",
			args: args{
				gitValues:   &[]string{"a1", "a2", "a3"},
				imageTags:   &[]string{"a1", "a2", "a3"},
				matchOption: MatchOptionPrefix,
			},
			want: []string{},
		},
		{
			name: "ShouldReturn_Empty_IfNoImages",
			args: args{
				gitValues:   &[]string{"a1", "a2", "a3"},
				imageTags:   &[]string{},
				matchOption: MatchOptionPrefix,
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterOrphanImageTags(tt.args.gitValues, tt.args.imageTags, tt.args.matchOption)
			assert.Equal(t, tt.want, result)
		})
	}
}
