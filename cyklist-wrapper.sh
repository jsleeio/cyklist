#!/bin/sh

docker run \
  --interactive \
  --tty \
  --rm \
  --env AWS_ACCESS_KEY_ID \
  --env AWS_SECRET_ACCESS_KEY \
  --env AWS_SECURITY_TOKEN \
  --env AWS_SESSION_TOKEN \
  --volume "$HOME/.kube:/.kube" \
  cyklist:latest \
  "$@"
