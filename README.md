confinit
========

Local configuration management for RaspberryPI and Linux


Development
===========

Golang


Makefile
--------

Manages releases and binaries. `make build` generates binaries for Linux *amd64* and *arm32*


Modules
-------

Go 1.11 has a feature `vgo` which will replace `dep`. To use `vgo`,
see https://github.com/golang/go/wiki/Modules.

TLDR below:

```
export GO111MODULE=on
go mod init         # If you are not using git, type `go mod init $(basename `pwd`)`
go mod vendor       # if you have vendor/ folder, will automatically integrate
go build
```

This method creates a file called `go.mod` in your projects directory.
You can then build your project with `go build`.
If `GO111MODULE=auto` is set, then your project cannot be in `$GOPATH`
