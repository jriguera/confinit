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

type ActionRouter struct {
	*Runner
	Condition string
	PreDelete bool
}

func NewActionRouter(glob, dst string, force, skipext, render, predelete bool, excludes []string) (*ActionRouter, error) {
	r, err := NewRunner(glob, dst, force, skipext, render, excludes)
	if err != nil {
		return nil, err
	}
	a := ActionRouter{
		Runner:    r,
		PreDelete: predelete,
	}
	return &a, nil
}

func (a *ActionRouter) SetCondition(c string) {
	a.Condition = c
}

func (a *ActionRouter) condition(data *TemplateData) (bool, string, error) {
	if a.Condition != "" {
		c, err := a.renderTemplateString("condition", a.Condition, data)
		if err != nil {
			return false, "", fmt.Errorf("Cannot render condition '%s', %s", a.Condition, err)
		}
		c = strings.TrimSpace(c)
		if c != "" {
			return false, c, nil
		}
	}
	return true, "", nil
}

func (a *ActionRouter) Function(base string, path string, i os.FileInfo) (err error) {
	tpldata := a.NewTemplateData(base, path, i)
	if _, err = os.Stat(tpldata.Destination); !os.IsNotExist(err) {
		if a.PreDelete && !i.IsDir() {
			if err = os.Remove(tpldata.Destination); err != nil {
				return
			}
		}
	}
	if c, msg, errc := a.condition(tpldata); errc != nil {
		return errc
	} else if !c {
		log.Infof("Skipping %s: %s", tpldata.SourceFullPath, msg)
		return nil
	}
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
