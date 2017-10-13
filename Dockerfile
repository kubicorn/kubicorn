FROM golang:latest

RUN mkdir -p /go/src/github.com/kris-nova/kubicorn/

WORKDIR /go/src/github.com/kris-nova/kubicorn/

COPY apis		apis
COPY cmd		cmd
COPY Makefile	Makefile
COPY state		state
COPY profiles	profiles
COPY test		test
COPY bootstrap	bootstrap
COPY docs		docs
COPY vendor		vendor
COPY cloud		cloud
COPY cutil		cutil
COPY examples	examples
COPY scripts	scripts
COPY Gopkg.lock Gopkg.lock
COPY Gopkg.toml Gopkg.toml
COPY main.go		.

RUN CGO_ENABLED=0 GOOS=linux  make docker-build-linux-amd64

FROM alpine:latest

MAINTAINER Kris Nova <kris@nivenly.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

WORKDIR /root/
COPY --from=0 /go/src/github.com/kris-nova/kubicorn .

RUN echo "Image build complete."


CMD [ "./kubicorn" ]
