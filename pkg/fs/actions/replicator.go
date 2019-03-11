package actions

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	fs "confinit/pkg/fs"
	log "confinit/pkg/log"
)

type Replicator struct {
	*Permissions
	Force    bool
	DirMode  os.FileMode
	FileMode os.FileMode
}

func NewReplicator(glob, dst string, typ fs.FsItemType, force bool) (*Replicator, error) {
	perm, err := NewPermissions(glob, dst, typ)
	if err != nil {
		return nil, err
	}
	r := Replicator{
		Permissions: perm,
		Force:       force,
		DirMode:     os.FileMode(0),
		FileMode:    os.FileMode(0),
	}
	return &r, nil
}

func (fr *Replicator) SetDefaultModes(dirmode, filemode os.FileMode) {
	fr.DirMode = dirmode
	fr.FileMode = filemode
}

func (fr *Replicator) mkdir(dst string, mode os.FileMode) error {
	if fr.DirMode != 0 {
		mode = fr.DirMode
	}
	if _, err := os.Stat(dst); os.IsNotExist(err) && fr.Force {
		if err := os.MkdirAll(dst, mode); err != nil {
			return err
		}
		log.Debugf("Folder %s created successfully", dst)
	}
	return nil
}

func (fr *Replicator) copyfile(src, dst string, dirmode, filemode os.FileMode) (int64, error) {
	if err := fr.mkdir(filepath.Dir(dst), dirmode); err != nil {
		return 0, err
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) && !fr.Force {
		// File exists and no force, skip
		log.Debugf("Skipped file %s, exists", dst)
		return 0, nil
	}
	if fr.FileMode != 0 {
		filemode = fr.FileMode
	}
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()
	destination, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, filemode)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	bytes, err := io.Copy(destination, source)
	if err == nil {
		log.Debugf("Successfully copied '%s' to '%s': %d bytes", src, dst, bytes)
		return bytes, err
	}
	err = fmt.Errorf("Cannot copy to '%s': %s", dst, err)
	return bytes, err
}

func (fr *Replicator) Function(base string, path string, i os.FileInfo) (err error) {
	srcpath := filepath.Join(base, path)
	dstpath := filepath.Join(fr.DstPath, path)
	if i.IsDir() {
		err = fr.mkdir(dstpath, i.Mode())
	} else {
		_, err = fr.copyfile(srcpath, dstpath, os.FileMode(0755), i.Mode())
		if err == nil {
			err = fr.applyPermissions(dstpath)
		}
	}
	return
}
