VERSION?=$(shell cat VERSION | tr -d " \t\n\r")
# Image URL to use all building/pushing image targets
REGISTRY?=birdhk
AMD64 ?= -amd64
EW_IMG ?= birdhk/conversions:$(VERSION)


.PHONY: build
build:
	@echo "Building..."
	docker buildx build --push --platform linux/amd64,linux/arm64 -f Dockerfile . -t ${EW_IMG}