VERSION := $(shell cat ./VERSION)
COMMIT_HASH := $(shell git rev-parse HEAD 2>/dev/null || true)
BUILD_TIME := $(shell date +%s)
DEFAULT_TEMPLATE := /etc/image-helpgen/template.tpl

# Used during all builds
LDFLAGS := -X main.version=${VERSION} -X main.commitHash=${COMMIT_HASH} -X main.buildTime=${BUILD_TIME}

CONFIG_DIR ?= /etc
BIN_DIR ?= /usr/bin

.PHONY: help build clean deps install lint test

help:
	@echo "Targets:"
	@echo " - build: Build the target binary"
	@echo " - clean: Clean up after build"
	@echo " - deps: Install required tool and dependencies for building"
	@echo " - install: Install build results to the system"
	@echo " - lint: Run golint"
	@echo " - test: Run unittests"
	@echo ""
	@echo "Variables:"
	@echo " - PREFIX: The root location to install. This prepends to all *_DIR variables. Set to: ${PREFIX}"
	@echo " - BIN_DIR: The directory that houses binaries. Set to: ${BIN_DIR}"
	@echo " - VERSION: Generally not overridden. The output of the VERSION file. Set to: ${VERSION}"
	@echo " - COMMIT_HASH: Generally not overridden. The git hash the code was built from. Set to: ${COMMIT_HASH}"
	@echo " - BUILD_TIME: Generally not overridden. The unix time of the build. Set to: ${BUILD_TIME}"

build: clean
	go build -ldflags '${LDFLAGS}' -o pivot main.go
	strip pivot

clean:
	rm -f pivot

deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure -v

install: clean build
	install -d ${PREFIX}${BIN_DIR}
	install --mode 755 pivot ${PREFIX}${BIN_DIR}/pivot

lint:
	go get -u github.com/golang/lint/golint
	golint .


test:
	go list ./... | grep -v vendor | xargs go test -v
