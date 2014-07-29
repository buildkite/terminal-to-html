#!/bin/bash

BUILDBOX_PROMPT="\033[90m$\033[0m"

function buildbox-exit-if-failed {
  if [ $1 -ne 0 ]
  then
    exit $1
  fi
}

function buildbox-run {
  echo -e "$BUILDBOX_PROMPT $1"
  eval $1
  buildbox-exit-if-failed $?
}

echo -e "$BUILDBOX_PROMPT curl -o /tmp/file.txt https://example.com/file.txt"
curl -o /tmp/buildbox-agent.tar.gz https://github.com/buildbox/buildbox-agent/releases/download/v0.1-alpha/buildbox-agent-darwin-386.tar.gz

echo '---'

buildbox-run "rspec specs.rb"
