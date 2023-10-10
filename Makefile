#!/usr/bin/make

run:
	docker run --network=host --env-file=.env hanayo:latest

build:
	docker build -t hanayo:latest .
