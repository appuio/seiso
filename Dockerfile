FROM docker.io/library/alpine:3.20 as runtime

ENTRYPOINT ["seiso"]

RUN \
    apk add --no-cache curl bash

COPY seiso /usr/bin/
USER 1000:0
