#!/bin/bash

if [[ $GITHUB_REF == refs/tags/* ]]; then
  # Extract tag name from GITHUB_REF
  tag_name=${GITHUB_REF#refs/tags/}
  echo "$tag_name"
else
  if [ -n "$GITHUB_SHA" ]; then
    short_sha=${GITHUB_SHA::7}
  elif [ -n "$CI_COMMIT_SHORT_SHA" ]; then
    short_sha=$CI_COMMIT_SHORT_SHA
  else
    git config --global --add safe.directory /app
    short_sha=$(git rev-parse --short HEAD)
  fi
  echo "${short_sha}"
fi
