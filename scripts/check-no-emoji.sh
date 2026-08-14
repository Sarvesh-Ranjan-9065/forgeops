#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

# check-no-emoji.sh fails if any Go source contains non-ASCII characters, which
# catches the emojis the project style forbids in source.

if grep -RPn "[^\x00-\x7F]" --include="*.go" . ; then
  echo "error: non-ASCII characters found in Go sources (emojis are not allowed)"
  exit 1
fi
echo "no-emoji check passed"