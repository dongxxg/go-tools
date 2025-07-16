SHELL := /bin/bash
CURRENT_DIR := $(shell pwd)
DIRS=$(shell ls)
DEBUG ?= 0
GIT_TAG := $(shell git describe --exact-match --tags --abbrev=0  2> /dev/null || echo untagged)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD || echo "0.0.0")
BUILD_DATE=$(shell date +%FT%T%z)

PLATFORMS ?= linux_arm64 linux_amd64

# only support linux
GOOS=linux
# set a specific PLATFORM, defaults to the host platform
ifeq ($(origin PLATFORM), undefined)
	ifeq ($(origin GOARCH), undefined)
		GOARCH := $(shell go env GOARCH)
	endif
	PLATFORM := $(GOOS)_$(GOARCH)
endif
