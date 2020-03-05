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
	"os"

	log "confinit/pkg/log"
)

type MapFile map[string]os.FileMode

type Fs struct {
	CurrentPath  string
	BasePath     string
	SkipDirGlob  *Glob
	SkipFileGlob *Glob
	FileGlob     *Glob
	DirGlob      *Glob
	files        MapFile
	dirs         MapFile
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
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	fglob, _ := NewGlob("*")
	dglob, _ := NewGlob("*")
	f := &Fs{
		CurrentPath:  dir,
		SkipDirGlob:  nil,
		SkipFileGlob: nil,
		FileGlob:     fglob,
		DirGlob:      dglob,
		files:        make(MapFile),
		dirs:         make(MapFile),
	}
	// call option functions on instance to set options on it
	for _, opt := range opts {
		opt(f)
	}
	return f
}
