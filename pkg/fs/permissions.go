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
	"path/filepath"

	log "confinit/pkg/log"
)

func (fs *Fs) SetPermissionsFiles(glob, uid, gid string, mode os.FileMode) error {
	perms, err := NewPermissions(glob, uid, gid, mode, FsItemFile)
	if err == nil {
		return fs.Run(perms)
	}
	return err
}

func (fs *Fs) SetPermissionsDirs(glob, uid, gid string, mode os.FileMode) error {
	perms, err := NewPermissions(glob, uid, gid, mode, FsItemDir)
	if err == nil {
		return fs.Run(perms)
	}
	return err
}

func (fs *Fs) SetPermissions(glob, uid, gid string, mode os.FileMode) error {
	perms, err := NewPermissions(glob, uid, gid, mode, FsItemAll)
	if err == nil {
		return fs.Run(perms)
	}
	return err
}

//

type Permissions struct {
	*Processor
	perm *Perm
}

func NewPermissions(glob, uid, gid string, mode os.FileMode, typ FsItemType) (*Permissions, error) {
	var exclude []string
	per, err := NewPerm(uid, gid, mode)
	if err != nil {
		return nil, err
	}
	proc, err := NewProcessor(glob, typ, exclude)
	if err != nil {
		return nil, err
	}
	p := Permissions{
		Processor: proc,
		perm:      per,
	}
	return &p, nil
}

func (fp *Permissions) Function(base string, path string, i os.FileMode) error {
	err := fp.perm.Set(filepath.Join(base, path))
	if err == nil {
		log.Debugf("Successfully applied permissions to '%s'", filepath.Join(base, path))
	}
	return err
}
