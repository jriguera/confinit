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
	"path/filepath"
	"sort"

	log "confinit/pkg/log"
)

func (fs *Fs) Scan(p string) error {
	fs.skippedPaths = nil
	fs.skippedFiles = nil
	fs.files = make(MapFileInfo)
	fs.dirs = make(MapFileInfo)
	fs.BasePath = p
	return filepath.Walk(fs.BasePath, fs.scan)
}

func (fs *Fs) scan(p string, i os.FileInfo, err error) error {
	if err != nil {
		err = fmt.Errorf("Cannot scan path '%s', %s", p, err.Error())
		return err
	}
	relp, err := filepath.Rel(fs.BasePath, p)
	if err != nil {
		return err
	}
	if i.IsDir() {
		if fs.SkipDirGlob != nil && fs.SkipDirGlob.MatchString(relp) {
			fs.skippedPaths = append(fs.skippedPaths, relp)
			log.Debugf("Skipping folder due to glob '%s': %s", fs.SkipDirGlob.String(), relp)
			return filepath.SkipDir
		} else if fs.DirGlob != nil && !fs.DirGlob.MatchString(relp) {
			fs.skippedPaths = append(fs.skippedPaths, relp)
			log.Debugf("Skipping folder due to not matching glob '%s': %s", fs.DirGlob.String(), relp)
			return filepath.SkipDir
		}
		log.Debugf("Adding folder: %s", p)
		fs.dirs[relp] = i
	} else if i.Mode().IsRegular() || i.Mode()&os.ModeSymlink != 0 {
		if fs.SkipFileGlob != nil && fs.SkipFileGlob.MatchString(relp) {
			fs.skippedFiles = append(fs.skippedFiles, relp)
			log.Debugf("Skipping file due to glob '%s': %s", fs.SkipFileGlob.String(), relp)
			return nil
		} else if fs.FileGlob != nil && !fs.FileGlob.MatchString(relp) {
			fs.skippedFiles = append(fs.skippedFiles, relp)
			log.Debugf("Skipping file due to not matching glob '%s': %s", fs.FileGlob.String(), relp)
			return nil
		}
		log.Debugf("Adding file: %s", p)
		fs.files[relp] = i
	} else {
		log.Debugf("Skipping non regular file: %s", p)
	}
	return nil
}

func (fs *Fs) ListSkipped(dirs bool) (items []string) {
	if dirs {
		items = fs.skippedPaths
	} else {
		items = fs.skippedFiles
	}
	sort.Strings(items)
	return
}

func (fs *Fs) ListDirs() (items []string) {
	for p := range fs.dirs {
		items = append(items, p)
	}
	sort.Strings(items)
	return
}

func (fs *Fs) ListFiles() (items []string) {
	for p := range fs.files {
		items = append(items, p)
	}
	sort.Strings(items)
	return
}
