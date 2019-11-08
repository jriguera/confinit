confinit
========

Boot configuration management for RaspberryPI and Linux, similar to Cloud-Init
but limited to copy files, render templates and execute commands cloning a folder
structure.


Usage
=====

Complete example: https://github.com/jriguera/raspbian-cloud/tree/master/stage8/99-confinit/config

Given a source folder structure (with files and sub-folders) the primary goal of
this program is replicate the same structure in a destination folder. The program
will scan the source folder and apply a list of operations (copy, render, execute)
in order. Each operation can define types of files/folders to process via regular
expressions and apply permissons on the new files.

There is also a global hooks to define startup and finish scripts.

Configuration
-------------

The configuration is defined in `internal/config/config.go`. You which
parameters are required and its defaults.

Global configuration parameters

```
#### confinit configuration file ###

# Log ouput can be:
# * stdout: dump logs to stdout
# * stderr: write logs to stderr
# * split: errors to stderr, rest of logs to stdout
# * - : discard all logs
# * path/to/file.log: dump logs to file
logoutput: split

# Log level defines the verbosity
loglevel: debug

# File (format json or yaml, by the extension) with additional variables
# accessible in templates
datafile: conf/data.yml

# Global environment variables accessible to programs and templates. Also the
# current environment variables are exported, here can be re-defined.
env:
    SYSTEM: raspberry
    LOCATION: home

# Startup command, non zero exit stops the execution.
# * timeout: defines how many seconds to wait for the execution (def)
# * dir: folder where the program will be executed (default is current dir)
# * env: key/value environment variables
start:
    cmd: ["pwd"]
    timeout: 60
    dir: /tmp
    env:
        A: a
        B: b

# Finish command, non zero exit stops the execution.
# * timeout: defines how many seconds to wait for the execution (def)
# * dir: folder where the program will be executed (default is current dir)
# * env: key/value environment variables
#
# There are two additional environment variables defined automatically:
# * CONFINIT_RC_START: stores the exit code of the `start` command.
# * CONFINIT_RC_PROCESS: stores the exit code of the `process` operations.
finish:
    cmd: ["env"]
    timeout: 600
    dir: /tmp
    env:
        A: a
        B: b
```

Processing files
----------------

The functionality of the program is defined in the field `process`:

```
# List of source folders. `source` is required and is the folder to clone.
# Optional filters can be set in `match` field using shell globs format, by
# default, it allows scanning all folders and files. The list of files is print
# out in debug `loglevel`. if `excludedone` is true (by default) when one file
# is processed by one operation, it will be ignored in other operations (is
# important the order of the operations).
process:
  - source: conf/templates
    excludedone: true
    match:
        folder:
           add: "*"
           skip: ".git"
        file:
           add: "*"
           skip: ".backup"
    operations: []
```

Operations:

* Copy files:
```
- destination: /
  regex: '.*'
  template: false
```

* Render templates (files wiht extension .template, which will be removed at destination)
```
- destination: /
  regex: '.*\.template'
  template: true
  delextension: true
  data:
    key1: value1
    key2: []
    key3:
       key4: {}
```

* Delete destination file, render template if condition renders to empty string
```
- destination: /
  regex: '.*\.template'
  template: true
  predelete: true
  condition: '{{ if not .Data.iface }}No interface{{ end }}'
```

* Execute script
```
- command:
    cmd: ["{{.SourceFullPath}}"]
    env:
      EXTRA_VAR: pepe
  regex: '.*\.sh'
```

* Copy and execute script
```
- destination: /tmp
  regex: '.*\.sh'
  default:
    mode:
      file: "0755"
  template: false
  command:
    cmd: ["{{.Destination}}"]
    env:
      EXTRA_VAR: pepe
```

* Render template at destination and execute it
```
- destination: /tmp
  regex: '.*\.sh'
  delextension: false
  default:
    mode:
      file: "0755"
  template: true
  command:
    cmd: ["{{.Destination}}"]
    env:
      EXTRA_VAR: pepe
  data:
    key: value
```

The template variables are defined in the file `pkg/fs/actions/templator.go:TemplateData`


Development
===========

Golang 1.11 . There is a `Makefile` to manage the development actions, releases
and binaries. `make build` generates binaries for Linux *amd64* and *arm32*

Golang Modules
--------------

Go 1.11 has a feature `vgo` which will replace `dep`. To use `vgo`,
see https://github.com/golang/go/wiki/Modules.

TLDR below:

```
export GO111MODULE=on
go mod init         # If you are not using git, type `go mod init $(basename `pwd`)`
go mod vendor       # if you have vendor/ folder, will automatically integrate
go build
```

This method creates a file called `go.mod` in your projects directory. You can
then build your project with `go build`. If `GO111MODULE=auto` is set, then your
project cannot be in `$GOPATH`


Author
======

(c) 2019 Jose Riguera

Apache 2.0
