#!/bin/bash

BUILDKITE_PROMPT="\033[90m$\033[0m"

function buildkite-run {
  echo -e "$BUILDKITE_PROMPT $1"
  eval $1
}

echo 'curl'
echo -e "$BUILDKITE_PROMPT curl -o /tmp/file.txt https://example.com/file.txt"
curl -L -o /tmp/buildkite.html https://buildkite.com/

echo '---'
echo 'specs'

buildkite-run "rspec -c specs.rb"

echo '---'
echo 'image'

buildkite-run "./print_image.sh"

echo '---'
echo 'pikachu'

buildkite-run "cat pikachu.ansi"
