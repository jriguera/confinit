#!/usr/bin/make -f

PROJECTNAME := $(shell basename "$(PWD)")
BINARY := ${PROJECTNAME}
BUILD := $(shell git rev-parse --short HEAD)
VERSION := $(shell head -n1 VERSION)
CHANGES := $(shell test -n "$$(git status --porcelain)" && echo '+CHANGES' || true)
PKGS := $(shell go list ./... | grep -v /vendor)
LDFLAGS := -X main.Build=$(BUILD) -X main.Version=$(VERSION)

# Go mod
export GO111MODULE=on

# Define architectures
BUILDER := linux-amd64 linux-arm-6 linux-arm-7
DEBPKG := deb-amd64 deb-armhf

# Go paths and tools
GOBIN := $(GOPATH)/bin
GOCMD := go
GOVET := $(GOCMD) tool vet
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOLINT := $(GOBIN)/golint
ERRCHECK := $(GOBIN)/errcheck
STATICCHECK := $(GOBIN)/staticcheck


.PHONY: all
all: test build deb

.PHONY: clean-all
clean-all: clean clean-vendor clean-build

.PHONY: clean
clean:
	@echo "*** Deleting go resources ***"
	$(GOCLEAN) -i ./...

.PHONY: clean-vendor
clean-vendor:
	@echo "*** Deleting vendor packages ***"
	find $(CURDIR)/vendor -type d -print0 2>/dev/null | xargs -0 rm -Rf

.PHONY: clean-build
clean-build:
	@echo "*** Deleting builds ***"
	@rm -Rf build/*
	@rm -Rf deb/*

.PHONY: test
test:
	@echo "*** Running tests ***"
	$(GOTEST) -v ./...

.PHONY: lint
lint: golint vet errcheck staticcheck unused checklicense

$(GOLINT):
	go get -u -v github.com/golang/lint/golint

$(ERRCHECK):
	go get -u github.com/kisielk/errcheck

$(STATICCHECK):
	go get -u honnef.co/go/tools/cmd/staticcheck

$(UNUSED):
	go get -u honnef.co/go/tools/cmd/unused

.PHONY: golint
golint: $(GOLINT)
	$(GOLINT) $(PKGS)

.PHONY: vet
vet:
	$(GOVET) -v $(PKGS)

.PHONY: errcheck
errcheck: $(ERRCHECK)
	$(ERRCHECK) ./...

.PHONY: staticcheck
staticcheck: $(STATICCHECK)
	$(STATICCHECK) ./...

.PHONY: unused
unused: $(UNUSED)
	$(UNUSED) ./...

.PHONY: $(GOMETALINTER)
$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

# linux-amd64, linux-arm-6, linux-arm-7, linux-arm64
.PHONY: $(BUILDER)
$(BUILDER):
	@echo "*** Building binary for $@ ***"
	$(eval OS := $(word 1,$(subst -, ,$@)))
	$(eval OSARCH := $(word 2,$(subst -, ,$@)))
	$(eval ARCHV := $(word 3,$(subst -, ,$@)))
	@if [ "$(OSARCH)" = "arm" ]; then export GOARM=${ARCHV}; fi
	@mkdir -p build
	GOOS=${OS} GOARCH=${OSARCH} ${GOBUILD} -ldflags "${LDFLAGS}" -o build/${BINARY}-${VERSION}-${OS}-${OSARCH}${ARCHV}

# deb-amd64 deb-armhf
.PHONY: $(DEBPKG)
$(DEBPKG):
	@echo "*** Building debian package for $@ ***"
	$(eval ARCH := $(word 2,$(subst -, ,$@)))
	@mkdir -p deb
	dpkg-buildpackage -rfakeroot -us -uc --host-arch=${ARCH} --target-arch=${${ARCH}}
	@mv -f ../confinit_* deb/

# from all
.PHONY: build
build: linux-amd64 linux-arm-6

# from all
.PHONY: deb
deb: deb-amd64 deb-armhf

