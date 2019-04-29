ARG ARCH
FROM ${ARCH}/alpine:latest

ADD qemu-statics /usr/bin/
RUN apk add --update curl ca-certificates jq
RUN rm -f /usr/bin/qemu-*-static

MAINTAINER th-docker@thzn.de
ARG ARCH

add suntrigger-$ARCH /suntrigger

ENTRYPOINT [ "/suntrigger" ]


