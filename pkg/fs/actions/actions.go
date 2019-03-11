package actions

import (
	"os"
)

type ActionRouter struct {
	*Runner
}

func NewActionRouter(glob, dst string, force, skipext, render bool) (*ActionRouter, error) {
	r, err := NewRunner(glob, dst, force, skipext, render)
	if err != nil {
		return nil, err
	}
	a := ActionRouter{
		Runner: r,
	}
	return &a, nil
}

func (a *ActionRouter) Function(base string, path string, i os.FileInfo) (err error) {
	if a.Cmd != "" {
		err = a.Runner.Function(base, path, i)
	} else {
		if a.DstPath != "" {
			if a.Render {
				err = a.Templator.Function(base, path, i)
			} else {
				err = a.Replicator.Function(base, path, i)
			}
		}
	}
	return
}
