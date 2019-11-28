#______________________________________________________________________________
#### Base Image, to save build time on local dev machine
ARG GOARCH=amd64
ARG ARCH=amd64-3.10
FROM docker.io/library/golang:1.13-alpine as builder

WORKDIR /go/src/app

COPY ["go.mod", "go.sum", "./"]

RUN \
    go mod download

ARG VERSION=unspecified
ARG GIT_COMMIT=unspecified

COPY / .
RUN \
    go build -ldflags "-X main.version=${VERSION} -X main.commit=${GIT_COMMIT} -X main.date=$(date -u '+%Y-%m-%dT%TZ')"

#______________________________________________________________________________
#### Runtime Image
ARG ARCH=amd64-3.10
FROM docker.io/multiarch/alpine:${ARCH} as runtime

ENV ALPINE_MIRROR=http://dl-cdn.alpinelinux.org/alpine
ENTRYPOINT ["image-cleanup"]

RUN \
    echo "${ALPINE_MIRROR}/${ALPINE_REL}/main" > /etc/apk/repositories && \
    echo "${ALPINE_MIRROR}/${ALPINE_REL}/community" >> /etc/apk/repositories && \
    apk update && \
    apk upgrade && \
    apk add --no-cache curl bash

COPY --from=builder /go/src/app/image-cleanup /usr/bin/image-cleanup
USER 1000:0
