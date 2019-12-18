package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var object = map[string]interface{}{"int": 20, "bool": true, "string": "foo"}

var foo = "foo"
var bar = "bar"

func Test_ObjectContainsTrue(t *testing.T) {
	assert.True(t, ObjectContains(object, foo))
}

func Test_ObjectContainsFalse(t *testing.T) {
	assert.False(t, ObjectContains(object, bar))
}
