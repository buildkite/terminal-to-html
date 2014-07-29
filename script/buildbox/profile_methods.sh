#!/bin/bash
set -e

echo '--- bundling'
bundle

echo '--- profilling methods'
bundle exec script/profile_methods
