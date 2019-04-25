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
package actions

import (
	"fmt"
	"os"
	"path/filepath"

	fs "confinit/pkg/fs"
	log "confinit/pkg/log"
)

type Permissions struct {
	*fs.Processor
	perms   map[string]*fs.Perm
	DstPath string
}

func NewPermissions(glob, dst string, typ fs.FsItemType, excludes []string) (*Permissions, error) {
	proc, err := fs.NewProcessor(glob, typ, excludes)
	if err != nil {
		return nil, err
	}
	p := Permissions{
		Processor: proc,
		perms:     make(map[string]*fs.Perm),
		DstPath:   dst,
	}
	return &p, nil
}

func (fp *Permissions) SetPermissions(glob, uid, gid string, mode os.FileMode) error {
	if _, err := fs.NewGlob(glob); err != nil {
		err = fmt.Errorf("Invalid glob pattern '%s' for permissions: %s", glob, err)
		return err
	}
	per, err := fs.NewPerm(uid, gid, mode)
	if err != nil {
		return err
	}
	fp.perms[glob] = per
	return nil
}

func (fp *Permissions) applyPermissions(dst string) error {
	e := false
	for glob, p := range fp.perms {
		pattern, _ := fs.NewGlob(glob)
		if pattern.MatchString(dst) {
			if err := p.Set(dst); err != nil {
				e = true
				log.Errorf("Cannot apply pemissions '%s' to '%s'", glob, dst)
			} else {
				log.Debugf("Successfully applied permissions to '%s': %s", dst, p)
			}
		}
	}
	if e {
		return fmt.Errorf("Cannot apply all pemissions to '%s'", dst)
	}
	return nil
}

func (fp *Permissions) Function(base string, path string, i os.FileMode) (err error) {
	return fp.applyPermissions(filepath.Join(fp.DstPath, path))
}
