PROJECTNAME = $(shell basename "$(PWD)")
BINARY = ${PROJECTNAME}

#BUILD := $(shell git rev-parse HEAD)
#VERSION := $(shell cat VERSION)
BUILD = aaaa
VERSION = 1
#CHANGES := $(shell test -n "$$(git status --porcelain)" && echo '+CHANGES' || true)
PKGS := $(shell go list ./... | grep -v /vendor)
LDFLAGS := -X main.Build=$(BUILD) -X main.Version=$(VERSION)

# Go mod
export GO111MODULE=on

# Define architectures
AMD64 = linux
ARM32 = 6 7

# Go paths and tools
GOBIN = $(GOPATH)/bin
GOCMD = go
GOVET = $(GOCMD) tool vet
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOLINT = $(GOBIN)/golint
ERRCHECK = $(GOBIN)/errcheck
STATICCHECK = $(GOBIN)/staticcheck

.PHONY: clean-all
clean-all: clean clean-vendor clean-release

.PHONY: clean
clean:
	$(GOCLEAN) -i ./...
	rm -Rf build/*

.PHONY: clean-vendor
clean-vendor:
	find $(CURDIR)/vendor -type d -print0 2>/dev/null | xargs -0 rm -Rf

.PHONY: clean-release
clean-releases:
	rm -Rf releases/*

.PHONY: test
test:
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

.PHONY: $(AMD64)
$(AMD64):
	@mkdir -p build
	$(eval OS := $(word 1, $@))
	GOOS=${OS} GOARCH=amd64 ${GOBUILD} -ldflags "${LDFLAGS}" -o build/${BINARY}-${VERSION}-${OS}-amd64

.PHONY: $(ARM32)
$(ARM32):
	@mkdir -p build
	$(eval ARM := $(word 1, $@))
	GOOS=linux GOARCH=arm GOARM=${ARM} ${GOBUILD} -ldflags "${LDFLAGS}" -o build/${BINARY}-${VERSION}-linux-arm${ARM}

.PHONY: build
build: $(ARM32) $(AMD64)
