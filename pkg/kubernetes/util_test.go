package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type ObjectContainsTestCase struct {
	object   interface{}
	value    string
	expected bool
}

func Test_ObjectContains(t *testing.T) {
	testcases := []ObjectContainsTestCase{
		{
			object:   map[string]interface{}{"int": 20, "bool": true, "string": "foo"},
			value:    "foo",
			expected: true,
		},
		{
			object:   map[string]interface{}{"int": 20, "bool": true, "string": "foo"},
			value:    "bar",
			expected: false,
		},
		{
			object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata":   map[string]interface{}{"name": "seiso", "namespace": "seiso-test"},
				"spec": map[string]interface{}{
					"containers": []interface{}{map[string]interface{}{
						"serviceAccount": "default",
						"image":          "docker.io/appuio/oc:0b81a958f590ed7ed8be6ec0a2a87816228a482c",
					}},
				},
			},
			value:    "oc:0b81a958f590ed7ed8be6ec0a2a87816228a482c",
			expected: true,
		},
	}

	for _, testcase := range testcases {
		assert.Equal(t, testcase.expected, ObjectContains(testcase.object, testcase.value))
	}
}
