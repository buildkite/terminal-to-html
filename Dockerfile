FROM golang:cross

RUN go get github.com/tools/godep

RUN mkdir -p /go/src/github.com/buildkite/terminal
ADD . /go/src/github.com/buildkite/terminal

WORKDIR /go/src/github.com/buildkite/terminal
