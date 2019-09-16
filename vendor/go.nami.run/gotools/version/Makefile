BUILD_DIR ?= ./artifacts

all:
	@$(MAKE) get-deps
	@$(MAKE) vet
	@$(MAKE) lint
	@$(MAKE) cover
	@$(MAKE) build

get-build-deps:
	@echo "+ Downloading build dependencies"
	@go get golang.org/x/tools/cmd/goimports
	@go get github.com/golang/lint/golint

build:
	@go build

include ./common.mk
