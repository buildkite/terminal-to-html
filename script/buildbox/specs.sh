#!/bin/bash
set -e

echo '--- bundling'
bundle

echo '--- specs'
bundle exec rake
