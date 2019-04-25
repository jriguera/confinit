// Copyright © 2019 Jose Riguera <jriguera@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package fs

import (
	"fmt"
	"os"
	"regexp"

	log "confinit/pkg/log"
)

type FsItemType int

const (
	FsItemAll  FsItemType = 0
	FsItemFile FsItemType = 1
	FsItemDir  FsItemType = 2
)

// Process is an interface to define a configurator factory
type Process interface {
	Function(base string, path string, i os.FileMode) error
	Match(path string, i os.FileMode) bool
	AddProcessed(path string, i os.FileMode)
	AddError(path string, err error)
	Type(t FsItemType) bool
	ListErrors() []string
	ListMapErrors() map[string]error
	ListProcessed() []string
	ListMapProcessed() map[string]os.FileMode
}

func (fs *Fs) Run(f Process) error {
	e := false
	if f.Type(FsItemAll) || f.Type(FsItemDir) {
		for path := range fs.dirs {
			if f.Match(path, fs.dirs[path]) {
				if err := f.Function(fs.BasePath, path, fs.dirs[path]); err != nil {
					f.AddError(path, err)
					log.Errorf("Could not complete process with folder '%s': %s", path, err)
					e = true
				}
				f.AddProcessed(path, fs.dirs[path])
			}
		}
	}
	if f.Type(FsItemAll) || f.Type(FsItemFile) {
		for path := range fs.files {
			if f.Match(path, fs.files[path]) {
				if err := f.Function(fs.BasePath, path, fs.files[path]); err != nil {
					f.AddError(path, err)
					log.Errorf("Could not complete process with file '%s': %s", path, err)
					e = true
				}
				f.AddProcessed(path, fs.dirs[path])
			}
		}
	}
	if e {
		return fmt.Errorf("There were errors running processes on some items: %v", f.ListErrors())
	}
	return nil
}

//

// Processor implments Process interface and is the base class for all its subclasses
// defined in actions package
type Processor struct {
	Regex     *regexp.Regexp
	FsType    FsItemType
	Exclude   []string
	Processed map[string]os.FileMode
	Errors    map[string]error
}

func NewProcessor(regex string, t FsItemType, exclude []string) (*Processor, error) {
	pattern, err := regexp.Compile(regex)
	if err != nil {
		err = fmt.Errorf("Invalid pattern '%s', %s", regex, err)
		return nil, err
	}
	p := Processor{
		Regex:     pattern,
		Processed: make(map[string]os.FileMode),
		Errors:    make(map[string]error),
		Exclude:   exclude,
		FsType:    t,
	}
	return &p, nil
}

func (p *Processor) Type(t FsItemType) bool {
	return p.FsType == t
}

func (p *Processor) Match(path string, i os.FileMode) bool {
	for _, exc := range p.Exclude {
		if exc == path {
			return false
		}
	}
	return p.Regex.MatchString(path)
}

func (p *Processor) AddProcessed(path string, i os.FileMode) {
	p.Processed[path] = i
}

func (p *Processor) AddError(path string, err error) {
	p.Errors[path] = err
}

func (p *Processor) ListErrors() []string {
	v := make([]string, 0, len(p.Errors))
	for k := range p.Errors {
		v = append(v, k)
	}
	return v
}

func (p *Processor) ListProcessed() []string {
	v := make([]string, 0, len(p.Processed))
	for k := range p.Processed {
		v = append(v, k)
	}
	return v
}

func (p *Processor) ListMapErrors() map[string]error {
	return p.Errors
}

func (p *Processor) ListMapProcessed() map[string]os.FileMode {
	return p.Processed
}

func (p *Processor) Function(base string, path string, i os.FileMode) error {
	return nil
}
