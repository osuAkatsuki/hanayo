#!/usr/bin/env bash
set -eo pipefail

if [ -n "$KUBERNETES" ]; then
    source /vault/secrets/secrets.txt
fi

if [ -z "$APP_COMPONENT" ]; then
  echo "Please set APP_COMPONENT"
  exit 1
fi

if [ -z "$APP_ENV" ]; then
  echo "Please set APP_ENV"
  exit 1
fi

if [ "$APP_COMPONENT" = "api" ]; then
    exec ./hanayo
else
    echo "Unknown component: $APP_COMPONENT"
    exit 1
fi
