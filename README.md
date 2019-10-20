# Image Cleanup Client

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
