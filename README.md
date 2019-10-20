# Image Cleanup Client

## General
The image cleanup client is used to clean up Docker images in a Docker Registry when they are tagged using git SHA.

This helps to save space because not anymore needed images are being removed from the registry.

## Dependencies

|Name|Usage|
|---|---|
|[openshift/client-go](https://github.com/openshift/client-go)|Interaction with OpenShift clusters|
|[heroku/docker-registry-client](https://github.com/heroku/docker-registry-client)|Interaction with Docker registries|
|[spf13/cobra](https://github.com/spf13/cobra)|CLI helper|
|[src-d/go-git](https://github.com/src-d/go-git)|Interaction with git repositories|
|||

The respective licenses for attribution are placed in `/attribution`.

## Development
Build and run:
```
go build -o client
./client
```
