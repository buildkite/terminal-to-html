#!/bin/bash

echo -e "\033[90m$\033[0m curl -o /tmp/file.txt https://example.com/file.txt"
curl -o /tmp/buildbox-agent.tar.gz https://github.com/buildbox/buildbox-agent/releases/download/v0.1-alpha/buildbox-agent-darwin-386.tar.gz
