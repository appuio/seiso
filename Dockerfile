FROM docker.io/library/alpine:3.11 as runtime

ENTRYPOINT ["image-cleanup"]

RUN \
    apk add --no-cache curl bash

COPY image-cleanup /usr/bin/image-cleanup
USER 1000:0
