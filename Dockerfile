FROM public.ecr.aws/docker/library/golang:1.23.4@sha256:574185e5c6b9d09873f455a7c205ea0514bfd99738c5dc7750196403a44ed4b7

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
