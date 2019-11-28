# Image Cleanup Client

![](https://img.shields.io/docker/pulls/appuio/image-cleanup)
![](https://img.shields.io/github/v/release/appuio/image-cleanup)
![](https://img.shields.io/github/license/appuio/image-cleanup)

## General

The image cleanup client is used to clean up Docker images in a Docker Registry when they are tagged using git SHA.

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
```

## License

This project is BSD 3-Clause licensed (see LICENSE file for details).

This project uses other OpenSource software as listed in `go.mod` and indirectly in `go.sum` files. All their licenses apply too.
