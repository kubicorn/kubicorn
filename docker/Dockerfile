FROM golang:1.8

WORKDIR /go/src/

RUN apt-get update && apt-get install -y libncurses5-dev
RUN go get google.golang.org/api/compute/v1
RUN go get cloud.google.com/go/compute/metadata