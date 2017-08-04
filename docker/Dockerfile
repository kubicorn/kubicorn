FROM golang:1.8

WORKDIR /go/src/

RUN apt-get update && apt-get install -y libncurses5-dev
RUN go get github.com/kris-nova/kubicorn/cmd