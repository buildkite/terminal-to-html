#!/bin/bash

# externalimg.sh path_or_url
if [ $# -eq 0 ]; then
  echo "Usage: externalimg.sh path_or_url"
  exit 1
fi

printf '\033]1338;path='
printf $1
printf '\a\n'

exit 0
