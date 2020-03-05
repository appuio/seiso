# Image Cleanup Client

[![dockeri.co](https://dockeri.co/image/appuio/image-cleanup)](https://hub.docker.com/r/appuio/image-cleanup)

![](https://img.shields.io/github/workflow/status/appuio/image-cleanup/Build)
![](https://img.shields.io/github/v/release/appuio/image-cleanup?include_prereleases)
![](https://img.shields.io/github/issues-raw/appuio/image-cleanup)
![](https://img.shields.io/github/issues-pr-raw/appuio/image-cleanup)
![](https://img.shields.io/github/license/appuio/image-cleanup)

## General

The image cleanup client is used to clean up container images in an image registry when they are tagged using git SHA. The cleaning is done either using git commit hashes or tags. Defaults to hashes otherwise ```--tag``` flag should be used. The tool also allows to clean orphan image stream tags using ```--orphan``` flag, the orphan image stream tags are images that do not have any Git commit/tag. There are secondary flags which help to norrow the cleaning process, for more information use ```--help```.

This helps to save space because obsolete images are being removed from the registry.

## Development

### Requirements

* go 1.13 (`sudo snap install --classic go`)
* goreleaser (`sudo snap install --classic goreleaser`)

### Build

```
goreleaser --snapshot --rm-dist
```

### Test

```
go test ./...
```

### Run
```
./dist/image-cleanup_linux_amd64/image-cleanup --help
# or
docker run --rm -it appuio/image-cleanup:<tag>
```

## Migrate from legacy cleanup plugin

Projects using the legacy `oc` cleanup plugin can be migrated to `image-cleanup` as follows

```console
oc -n "$OPENSHIFT_PROJECT" plugin cleanup "$APP_NAME" -p "$PWD" -f=y
```
becomes:
```console
image-cleanup imagestream "$APP_NAME" --namespace="$OPENSHIFT_PROJECT" --force --git-commit-limit=0
```

## Release

Push a git tag with the scheme `vX.Y.Z`.

## License

This project is BSD 3-Clause licensed (see LICENSE file for details).

This project uses other OpenSource software as listed in `go.mod` and indirectly in `go.sum` files. All their licenses apply too.
