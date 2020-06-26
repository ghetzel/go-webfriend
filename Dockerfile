FROM alpine:3.12
MAINTAINER PerformLine Engineering <tech+docker@performline.com>

RUN apk update && apk add bash git make curl chromium jq openssl msttcorefonts-installer
RUN update-ms-fonts
RUN fc-cache -f
ADD bin/webfriend-linux-amd64 /usr/bin/webfriend

ENV LOGLEVEL debug
ENTRYPOINT "webfriend"
