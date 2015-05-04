FROM golang:cross

# buildkite-agent for artifact management
RUN echo deb https://apt.buildkite.com/buildkite-agent unstable main > /etc/apt/sources.list.d/buildkite-agent.list
RUN apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 32A37959C2FA5C3C99EFBC32A79206696452D198

# To get zip
RUN apt-get update

# Install buildkite-agent
RUN apt-get install -y buildkite-agent

# For dealing with Go deps
RUN go get github.com/tools/godep

# For creating Github releases
RUN go get github.com/buildkite/github-release

# Zip for win and osx releases
RUN apt-get install -y zip

RUN mkdir -p /go/src/github.com/buildkite/terminal
ADD . /go/src/github.com/buildkite/terminal

WORKDIR /go/src/github.com/buildkite/terminal
