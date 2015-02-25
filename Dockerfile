FROM golang:cross

# The golang image sets the $GOPATH to be /go, and copies all the source files
# into the WORKDIR

RUN mkdir -p /go/src/github.com/buildkite/terminal
WORKDIR /go/src/github.com/buildkite/terminal
