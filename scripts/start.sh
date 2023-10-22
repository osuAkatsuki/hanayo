#!/usr/bin/env bash
set -eo pipefail

if [ -z "$APP_COMPONENT" ]; then
  echo "Please set APP_COMPONENT"
  exit 1
fi

if [ -z "$APP_ENV" ]; then
  echo "Please set APP_ENV"
  exit 1
fi

if [[ $PULL_SECRETS_FROM_VAULT -eq 1 ]]; then
  # TODO: is there a better way to deal with this?
  pip install --break-system-packages -i $PYPI_INDEX_URL akatsuki-cli
  akatsuki vault get hanayo $APP_ENV -o .env
  source .env
fi

if [ "$APP_COMPONENT" = "api" ]; then
    exec ./hanayo
else
    echo "Unknown component: $APP_COMPONENT"
    exit 1
fi
