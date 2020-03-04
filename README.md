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

# File (format json or yaml, by the extension) or HTTP url (GET) with additional
# variables/structures accessible in templates
datafile: conf/data.yml

# Global environment variables accessible to programs and templates. Also the
# current environment variables are exported, here can be re-defined.
env:
    SYSTEM: raspberry
    LOCATION: home

# Startup command, non zero exit stops the execution.
# * timeout: defines how many seconds to wait for the execution (def)
# * dir: folder where the program will be executed (default is current dir)
# * env: key/value environment variables.
# This command can perform operations on the datafile (see above), like
# getting from a database/url and building it. Data from datafile is loaded
# after this command runs and before the list of operations.
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
# * CONFINIT_RC_LOAD_DATA: stores an exit code of the result of loading the
# datafile.
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

Descriptive examples of operations:

1. Copy all files to a destination (even binaries):
```
- destination: /
  regex: '.*'
  template: false
```

2. Render templates, files with extension `.template` (which will be removed at
destination, after rendering it because of `delextension: true`). Extra data,
added on top of `datafile` setting, will be used to process these templates:
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

3. Render template to determine what to do with the file (`delete` in this case)
```
- destination: /
  regex: '.*\.template'
  template: true
  condition: '{{if not .Data.iface}} delete {{end}}'
  delete:
    ifconfition: true
```

Other options to echo to `condition` field are (case insesitive, trim spaces):

  * with `""`, `render` or `continue`: continue processing file/template.
  * case `skip`: stop rendering template file (do not delete destination file if exists).
  * `delete` or `delete-file` : stop rendering and delete current file (if exists).
  * `delete-if-empty`: delete the file only if template results in an empty file.
  * `delete-if-fail`: delete if template calls `fail "<msg>"` function or renderting template fails.
  * `delete-after-exec`: delete a file (when is a command/script template, see below) after its execution.

The setting `delete.ifcondition` (default `true`) controls if rendering templates
can define delete actions. If is `false` it will NOT render the template if the 
condition generates an output string, the output script will be used as informational
message in the logs.

Apart from the conditional ouput, `delete` parameter operates by its own and has
these options with default values:

  * `prestart` (default `false`): always delete the file before processing it.
  * `ifempty` (default `true`): deletes if renders to an empty file.
  * `ifconfition` (default `true`): "do what the condition says".
  * `ifrenderfail` (default `true`): delete if the template does not render (or it calls `fail` function).
  * `afterexec` (default `true`): delete after executing the file (see below).


4. Execute script (do not copy or render it, only execute it from the source)
```
- command:
    cmd: ["{{.SourceFullPath}}"]
    env:
      EXTRA_VAR: pepe
  regex: '.*\.sh'
```

5. Copy and execute script, `command.cmd` points to the destination of the file being rendered.
Delete destination file after running it.
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
  delete:
    afterexec: true
```

6. Render template (keeping the full extension of the filename) and execute it (do not delete it):
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
  delete:
    afterexec: false
```

Templates
---------

Templates are implemented using `golang/text.template`. There is more information
about how to write them in the official documentation: https://golang.org/pkg/text/template/

Confinit defines template variables in the file `pkg/fs/actions/templator.go:TemplateData`
```
	IsDir           bool
	Mode            string
	SourceBaseDir   string
	Source          string
	Filename        string
	SourceFile      string
	Path            string
	SourceFullPath  string
	SourcePath      string
	Ext             string
	DstBaseDir      string
	Destination     string
	DestinationPath string
	Data            interface{}
	Env             map[string]string
```

So, for example, in order to get a variable defined in `datafile` you have 
define `{{ .Data.VARIABLE }} and to get the destination path of the current template
`{{ .Destination }}`. Those getters can also be used in the `condition` parameter.

There are a lot of template functions defined in the file `pkg/tplfunctions/tfunctions.go`
ready to be used in template files, for example:

```
UUID: {{ uuid }}
randomString: {{ randomString 6 }}
randominteger: {{ random "0123456789" 6 }}
env: {{ .Env.EEEE  }}
env: {{ env "EEEE" }}
now: {{ now | date "2006-01-02" }}
now: {{ now | epoch }}
remove_spaces ({{ .Data.D }}): {{ trim .Data.D }}
```


Development
===========

Golang 1.11 . There is a `Makefile` to manage the development actions, releases
and binaries. `make build` generates binaries for: `linux-amd64`, `linux-arm-6`,
`linux-arm-7` and `make deb` generates debian packages for `deb-amd64`, `deb-armhf`

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


Debian package
--------------

The usual call to build a binary package is `dpkg-buildpackage -us -uc`.
You might call debuild for other purposes, like `debuild clean` for instance.

```
# -us -uc skips package signing.
dpkg-buildpackage -rfakeroot -us -uc
```

Author
======

(c) 2019,2020 Jose Riguera

Apache 2.0
