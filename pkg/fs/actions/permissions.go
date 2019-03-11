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

func NewPermissions(glob, dst string, typ fs.FsItemType) (*Permissions, error) {
	proc, err := fs.NewProcessor(glob, typ)
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

func (fp *Permissions) Function(base string, path string, i os.FileInfo) (err error) {
	return fp.applyPermissions(filepath.Join(fp.DstPath, path))
}
