package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type UnstructuredListContainsTestCase struct {
	objectlist *unstructured.UnstructuredList
	value      string
	expected   bool
}

func Test_UnstructuredListContains(t *testing.T) {
	testcases := []UnstructuredListContainsTestCase{
		// Successful lookup string in map[string]interface{}
		{
			objectlist: &unstructured.UnstructuredList{
				Object: map[string]interface{}{"kind": "List", "apiVersion": "v1"},
				Items: []unstructured.Unstructured{
					{Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"int":  20,
							"bool": true,
							"[]interface": []interface{}{map[string]interface{}{
								"serviceAccount": "bar",
								"name":           "bar",
							}},
							"name": "foo",
						},
					}},
				},
			},
			value:    "foo",
			expected: true,
		},
		// Successful lookup string in []interface{}
		{
			objectlist: &unstructured.UnstructuredList{
				Object: map[string]interface{}{"kind": "List", "apiVersion": "v1"},
				Items: []unstructured.Unstructured{
					{Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"int":  20,
							"bool": true,
							"[]interface": []interface{}{map[string]interface{}{
								"serviceAccount": "foo",
								"name":           "bar",
							}},
							"name": "foo",
						},
					}},
				},
			},
			value:    "bar",
			expected: true,
		},
		// Lookup for non-existing value
		{
			objectlist: &unstructured.UnstructuredList{
				Object: map[string]interface{}{"kind": "List", "apiVersion": "v1"},
				Items: []unstructured.Unstructured{
					{Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"int":  20,
							"bool": true,
							"[]interface": []interface{}{map[string]interface{}{
								"serviceAccount": "foo",
								"name":           "foo",
							}},
							"name": "foo",
						},
					}},
				},
			},
			value:    "bar",
			expected: false,
		},
		// Successful lookup for value in second element
		{
			objectlist: &unstructured.UnstructuredList{
				Object: map[string]interface{}{"kind": "List", "apiVersion": "v1"},
				Items: []unstructured.Unstructured{
					{Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"int":  20,
							"bool": true,
							"[]interface": []interface{}{map[string]interface{}{
								"serviceAccount": "foo",
								"name":           "foo",
							}},
							"name": "foo",
						},
					}},
					{Object: map[string]interface{}{
						"kind":       "Pod",
						"apiVersion": "v1",
						"metadata": map[string]interface{}{
							"int":  20,
							"bool": true,
							"[]interface": []interface{}{map[string]interface{}{
								"serviceAccount": "foo",
								"name":           "foo",
							}},
							"name": "bar",
						},
					}},
				},
			},
			value:    "bar",
			expected: true,
		},
	}

	for _, testcase := range testcases {
		assert.Equal(t, testcase.expected, UnstructuredListContains(testcase.objectlist, testcase.value))
	}
}
