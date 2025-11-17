FROM public.ecr.aws/docker/library/golang:1.25.4@sha256:e68f6a00e88586577fafa4d9cefad1349c2be70d21244321321c407474ff9bf2

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
