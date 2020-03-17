package docker

import (
	"github.com/heroku/docker-registry-client/registry"
)

// GetImageTags returns the tags of a docker image
func GetImageTags(image string) ([]string, error) {
	url := "https://registry-1.docker.io/"
	username := "" // anonymous
	password := "" // anonymous
	hub, err := registry.New(url, username, password)
	if err != nil {
		return nil, err
	}

	return hub.Tags(image)
}
