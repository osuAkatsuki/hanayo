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
  pip install -i $PYPI_INDEX_URL akatsuki-cli
  akatsuki vault get bancho-service $APP_ENV -o .env
  source .env
fi

if [ "$APP_COMPONENT" = "api" ]; then
    exec ./hanayo
else
    echo "Unknown component: $APP_COMPONENT"
    exit 1
fi
