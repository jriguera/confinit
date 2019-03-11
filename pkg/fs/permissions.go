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
	per, err := NewPerm(uid, gid, mode)
	if err != nil {
		return nil, err
	}
	proc, err := NewProcessor(glob, typ)
	if err != nil {
		return nil, err
	}
	p := Permissions{
		Processor: proc,
		perm:      per,
	}
	return &p, nil
}

func (fp *Permissions) Function(base string, path string, i os.FileInfo) error {
	err := fp.perm.Set(filepath.Join(base, path))
	if err == nil {
		log.Debugf("Successfully applied permissions to '%s'", filepath.Join(base, path))
	}
	return err
}
