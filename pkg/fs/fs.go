package fs

import (
	"os"

	log "confinit/pkg/log"
)

type MapFileInfo map[string]os.FileInfo

type Fs struct {
	BasePath     string
	SkipDirGlob  *Glob
	SkipFileGlob *Glob
	FileGlob     *Glob
	DirGlob      *Glob
	files        MapFileInfo
	dirs         MapFileInfo
	skippedPaths []string
	skippedFiles []string
}

// Option to pass to the constructor using Functional Options
type Option func(*Fs)

// SkipDirGlob is a function used by users to set options.
func SkipDirGlob(s string) Option {
	return func(f *Fs) {
		if s != "" {
			if pattern, err := NewGlob(s); err != nil {
				log.Errorf("Invalid SkipDir glob pattern '%s', %s", s, err.Error())
			} else {
				f.SkipDirGlob = pattern
			}
		}
	}
}

// SkipFileGlob is a function used by users to set options.
func SkipFileGlob(s string) Option {
	return func(f *Fs) {
		if s != "" {
			if pattern, err := NewGlob(s); err != nil {
				log.Errorf("Invalid SkipFile glob pattern '%s', %s", s, err.Error())
			} else {
				f.SkipFileGlob = pattern
			}
		}
	}
}

// FileGlob is a function used by users to set options.
func FileGlob(s string) Option {
	return func(f *Fs) {
		if s != "" {
			if pattern, err := NewGlob(s); err != nil {
				log.Errorf("Invalid File glob pattern '%s', %s", s, err.Error())
			} else {
				f.FileGlob = pattern
			}
		}
	}
}

// DirGlob is a function used by users to set options.
func DirGlob(s string) Option {
	return func(f *Fs) {
		if s != "" {
			if pattern, err := NewGlob(s); err != nil {
				log.Errorf("Invalid Dir glob pattern '%s', %s", s, err.Error())
			} else {
				f.DirGlob = pattern
			}
		}
	}
}

// New is the contructor
func New(opts ...Option) *Fs {
	fglob, _ := NewGlob("*")
	dglob, _ := NewGlob("*")
	f := &Fs{
		SkipDirGlob:  nil,
		SkipFileGlob: nil,
		FileGlob:     fglob,
		DirGlob:      dglob,
		files:        make(MapFileInfo),
		dirs:         make(MapFileInfo),
	}
	// call option functions on instance to set options on it
	for _, opt := range opts {
		opt(f)
	}
	return f
}
