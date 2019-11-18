# Image Cleanup Client

## General
The image cleanup client is used to clean up Docker images in a Docker Registry when they are tagged using git SHA.

This helps to save space because not anymore needed images are being removed from the registry.

The respective licenses for attribution are placed in `/attribution`.

## Development
Build and run:
```
go build -o client
./client
```
