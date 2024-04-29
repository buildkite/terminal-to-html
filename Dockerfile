FROM public.ecr.aws/docker/library/golang:1.22.2@sha256:d5302d40dc5fbbf38ec472d1848a9d2391a13f93293a6a5b0b87c99dc0eaa6ae

ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US:en \
    LC_ALL=en_US.UTF-8

RUN apt-get update -q && \
  apt-get install -y zip ruby ruby-dev rpm locales

RUN go install github.com/buildkite/github-release@latest

RUN gem install --no-document rake fpm package_cloud

RUN echo "en_US UTF-8" > /etc/locale.gen && \
  locale-gen en_US.UTF-8

WORKDIR /go/src/github.com/buildkite/terminal-to-html
ADD . /go/src/github.com/buildkite/terminal-to-html

CMD [ "make", "dist" ]
