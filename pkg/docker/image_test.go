package docker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const IMAGE = "appuio/oc"

func Test_GetImageTags(t *testing.T) {
	imageTags, err := GetImageTags(IMAGE)

	assert.NoError(t, err)
	assert.NotEmpty(t, imageTags)
}
