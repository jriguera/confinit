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
	"strings"

	log "confinit/pkg/log"
)

type DeleteType uint32

const (
	DeleteNever DeleteType = 1 << iota
	DeletePreStart
	DeleteIfCondition
	DeleteIfEmpty
	DeleteIfRenderFail
	DeleteAfterExec
)

func (d DeleteType) Has(flag DeleteType) bool {
	return d&flag != 0
}

func (d *DeleteType) Set(flag DeleteType) {
	*d |= flag
}

func (d *DeleteType) Unset(flag DeleteType) {
	*d &= ^flag
}

func (d *DeleteType) Names(sep string) (result string) {
	var bits []string
	var names = map[DeleteType]string{
		DeleteNever:        "no",
		DeletePreStart:     "pre-start",
		DeleteIfCondition:  "if-condition",
		DeleteIfEmpty:      "if-empty",
		DeleteIfRenderFail: "if-fail",
		DeleteAfterExec:    "after-exec",
	}
	for k, v := range names {
		if d.Has(k) {
			bits = append(bits, v)
		}
	}
	result = strings.Join(bits, sep)
	return
}

type ActionRouter struct {
	*Runner
	Condition string
	Delete    DeleteType
}

func NewActionRouter(glob, dst string, force, skipext, render bool, excludes []string) (*ActionRouter, error) {
	r, err := NewRunner(glob, dst, force, skipext, render, excludes)
	if err != nil {
		return nil, err
	}
	a := ActionRouter{
		Runner: r,
		Delete: DeleteNever,
	}
	return &a, nil
}

func (a *ActionRouter) SetCondition(c string) {
	a.Condition = c
}

func (a *ActionRouter) SetDelete(delete DeleteType) {
	a.Delete.Set(delete)
}

func (a *ActionRouter) condition(data *TemplateData) (bool, string, error) {
	if a.Condition != "" {
		c, err := a.renderTemplateString("condition", a.Condition, data)
		if err != nil {
			return false, "", fmt.Errorf("Cannot render condition '%s', %s", a.Condition, err)
		}
		c = strings.TrimSpace(c)
		if !a.Delete.Has(DeleteIfCondition) {
			if c != "" {
				return false, c, nil
			}
			return true, "render", nil
		}
		switch output := strings.ToLower(c); output {
		case "":
			// continue
			return true, "render", nil
		case "render":
			// continue
			return true, c, nil
		case "skip":
			// no render
			return false, c, nil
		case "delete":
			// no render
			a.Delete.Set(DeletePreStart)
			return false, c, nil
		case "delete-if-empty":
			a.Delete.Set(DeleteIfEmpty)
			return true, c, nil
		case "delete-if-fail":
			a.Delete.Set(DeleteIfRenderFail)
			return true, c, nil
		case "delete-after-exec":
			a.Delete.Set(DeleteAfterExec)
			return true, c, nil
		}
		// no render, show message
		return false, c, nil
	}
	return true, "render", nil
}

func (a *ActionRouter) Function(base string, path string, i os.FileMode) (err error) {
	tpldata := a.NewTemplateData(base, path, i)
	action := ""
	c, msg, errc := a.condition(tpldata)
	if errc != nil {
		return errc
	} else if !c {
		log.Infof("Skipping render %s, condition reported: %s", tpldata.SourceFullPath, msg)
		return nil
	}
	if a.DstPath != "" {
		if _, err = os.Stat(tpldata.Destination); !os.IsNotExist(err) {
			if a.Delete.Has(DeletePreStart) && !i.IsDir() {
				if err = os.Remove(tpldata.Destination); err != nil {
					return
				}
			}
		}
	}
	if a.Cmd != "" {
		action, err = a.Runner.Function(base, path, i)
		if a.DstPath != "" && a.Delete.Has(DeleteAfterExec) {
			os.Remove(tpldata.Destination)
			log.Infof("Condition delete-after-exec triggered for %s, deleted", tpldata.Destination)
		}
	} else {
		if a.DstPath != "" {
			if a.Render {
				action, err = a.Templator.Function(base, path, i)
				if err != nil && a.Delete.Has(DeleteIfRenderFail) {
					os.Remove(action)
					log.Infof("Condition delete-if-error triggered for %s, deleted", action)
				}
			} else {
				action, err = a.Replicator.Function(base, path, i)
			}
			if err == nil && a.Delete.Has(DeleteIfEmpty) {
				if fi, err := os.Stat(action); err == nil {
					size := fi.Size()
					if size <= 0 {
						log.Infof("Condition delete-if-empty triggered for %s, deleted", action)
						err = os.Remove(action)
					}
				}
			}
		}
	}
	return
}
