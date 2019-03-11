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
	Function(base string, path string, i os.FileInfo) error
	Match(path string, i os.FileInfo) bool
	AddItem(path string, i os.FileInfo)
	AddError(path string, err error)
	Type(t FsItemType) bool
	ListItemErrors() []string
	ListMapErrors() map[string]error
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
				} else {
					f.AddItem(path, fs.files[path])
				}
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
				} else {
					f.AddItem(path, fs.files[path])
				}
			}
		}
	}
	if e {
		return fmt.Errorf("There were errors running processes on some items: %v", f.ListItemErrors())
	}
	return nil
}

//

// Processor implments Process interface and is the base class for all its subclasses
// defined in actions package
type Processor struct {
	Regex  *regexp.Regexp
	FsType FsItemType
	Items  map[string]os.FileInfo
	Errors map[string]error
}

func NewProcessor(regex string, t FsItemType) (*Processor, error) {
	pattern, err := regexp.Compile(regex)
	if err != nil {
		err = fmt.Errorf("Invalid pattern '%s', %s", regex, err)
		return nil, err
	}
	p := Processor{
		Regex:  pattern,
		Items:  make(map[string]os.FileInfo),
		Errors: make(map[string]error),
		FsType: t,
	}
	return &p, nil
}

func (p *Processor) Type(t FsItemType) bool {
	return p.FsType == t
}

func (p *Processor) Match(path string, i os.FileInfo) bool {
	return p.Regex.MatchString(path)
}

func (p *Processor) AddItem(path string, i os.FileInfo) {
	p.Items[path] = i
}

func (p *Processor) AddError(path string, err error) {
	p.Errors[path] = err
}

func (p *Processor) ListItemErrors() []string {
	v := make([]string, 0, len(p.Errors))
	for k, _ := range p.Errors {
		v = append(v, k)
	}
	return v
}

func (p *Processor) ListMapErrors() map[string]error {
	return p.Errors
}

func (p *Processor) Function(base string, path string, i os.FileInfo) error {
	return nil
}
