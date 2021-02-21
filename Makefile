#!/bin/bash

ifndef IMAGE_NAME
override IMAGE_NAME=quay.io/munnerz/kube-plex
endif

build:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/kube-plex_linux_amd64 main.go
	docker build -t ${IMAGE_NAME} .

push: build
	echo "Building ${IMAGE_NAME}"
	docker push ${IMAGE_NAME}