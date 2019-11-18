# Image Cleanup Client

## General

The image cleanup client is used to clean up Docker images in a Docker Registry when they are tagged using git SHA.

This helps to save space because obsolete images are being removed from the registry.

The respective licenses for attribution are placed in `/attribution`.

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
