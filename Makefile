VERSION := $(shell cat ./VERSION)
COMMIT_HASH := $(shell git rev-parse HEAD 2>/dev/null || true)
BUILD_TIME := $(shell date +%s)
DEFAULT_TEMPLATE := /etc/image-helpgen/template.tpl
# Used during all builds
LDFLAGS := -X main.version=${VERSION} -X main.commitHash=${COMMIT_HASH} -X main.buildTime=${BUILD_TIME}

PREFIX ?= /usr
CONFIG_DIR ?= /etc
BIN_DIR ?= ${PREFIX}/bin
SYSTEMD_UNIT_DIR ?= ${PREFIX}/lib/systemd/system

.PHONY: help build clean deps install lint static test

help:
	@echo "Targets:"
	@echo " - build: Build the target binary"
	@echo " - static: Build a static binary"
	@echo " - clean: Clean up after build"
	@echo " - deps: Install required tool and dependencies for building"
	@echo " - install: Install build results to the system"
	@echo " - lint: Run golint"
	@echo " - test: Run unittests"
	@echo " - changelog: Returns the changes from the last tag up till HEAD"
	@echo ""
	@echo "Variables:"
	@echo " - PREFIX: The root location to install. This prepends to all *_DIR variables. Set to: ${PREFIX}"
	@echo " - BIN_DIR: The directory that houses binaries. Set to: ${BIN_DIR}"
	@echo " - VERSION: Generally not overridden. The output of the VERSION file. Set to: ${VERSION}"
	@echo " - COMMIT_HASH: Generally not overridden. The git hash the code was built from. Set to: ${COMMIT_HASH}"
	@echo " - BUILD_TIME: Generally not overridden. The unix time of the build. Set to: ${BUILD_TIME}"

changelog:
	git log --format="- %s" `git tag | tail -n 1`..HEAD

systemd/pivot.service: systemd/pivot.service.in
	sed "s,@@PIVOT_BINARY_PATH@@,${BIN_DIR}/pivot,g" < systemd/pivot.service.in > systemd/pivot.service

pivot: Gopkg.* *.go cmd/*.go utils/*.go types/*.go
	go build -ldflags '${LDFLAGS}' -o pivot main.go
	strip pivot

build: pivot systemd/pivot.service

static: clean
	CGO_ENABLED=0 go build -ldflags '${LDFLAGS} -w -extldflags "-static"' -a -o pivot main.go

clean:
	rm -f pivot

deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure -v

install: build
	install -d ${DESTDIR}${BIN_DIR}
	install --mode 755 pivot ${DESTDIR}${BIN_DIR}/pivot
	install -d ${DESTDIR}${SYSTEMD_UNIT_DIR}
	install --mode 664 systemd/pivot.service ${DESTDIR}${SYSTEMD_UNIT_DIR}

lint:
	go get -u github.com/golang/lint/golint
	golint .


test:
	go list ./... | grep -v vendor | xargs go test -v
