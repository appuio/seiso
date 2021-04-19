IMG_TAG ?= latest

BIN_FILENAME ?= $(PROJECT_ROOT_DIR)/seiso

SHASUM ?= $(shell command -v sha1sum > /dev/null && echo "sha1sum" || echo "shasum -a1")

# Image URL to use all building/pushing image targets
DOCKER_IMG ?= docker.io/appuio/seiso:$(IMG_TAG)
QUAY_IMG ?= quay.io/appuio/seiso:$(IMG_TAG)
