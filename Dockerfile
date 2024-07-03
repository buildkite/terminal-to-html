FROM public.ecr.aws/docker/library/golang:1.22.5@sha256:45c516d2adb456d6a0ca3cc5172013abbf65d5152dbd5f9edffaa63fbf96a96d

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
