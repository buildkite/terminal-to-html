#!/bin/bash

echo -e "\033[90m$\033[0m curl -o /tmp/file.txt https://example.com/file.txt"
curl -o /tmp/buildkite-agent.tar.gz https://github.com/buildkite/buildkite-agent/releases/download/v0.1-alpha/buildkite-agent-darwin-386.tar.gz
