#!/usr/bin/make

build:
	docker build -t hanayo:latest -t registry.digitalocean.com/akatsuki/hanayo:latest .

push:
	docker push registry.digitalocean.com/akatsuki/hanayo:latest

install:
	helm install --values chart/values.yaml hanayo-staging ../common-helm-charts/microservice-base/

uninstall:
	helm uninstall hanayo-staging

diff-upgrade:
	helm diff upgrade --allow-unreleased --values chart/values.yaml hanayo-staging ../common-helm-charts/microservice-base/

upgrade:
	helm upgrade --atomic --values chart/values.yaml hanayo-staging ../common-helm-charts/microservice-base/
