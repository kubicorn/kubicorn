FROM golang:latest
MAINTAINER Kris Nova <kris@nivenly.com>

ADD rootfs /go/src/github.com/kris-nova/kubicorn/rootfs
RUN cd /go/src/github.com/kris-nova/terraformctl/rootfs && \
    mv /go/src/github.com/kris-nova/terraformctl/rootfs/.azure ~/.azure




ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go





ENTRYPOINT ./kubicorn controller -v 4
