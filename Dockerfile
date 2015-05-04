FROM golang:cross

# To get zip
RUN apt-get update

# For dealing with Go deps
RUN go get github.com/tools/godep

# For creating Github releases
RUN go get github.com/buildkite/github-release

# Zip for win and osx releases
RUN apt-get install -y zip

RUN mkdir -p /go/src/github.com/buildkite/terminal
ADD . /go/src/github.com/buildkite/terminal

WORKDIR /go/src/github.com/buildkite/terminal
