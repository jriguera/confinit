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
	//log "confinit/pkg/log"
)

// ActionTemplate is an interface to define a configurator factory
type ActionTemplate interface {
	Command(command []string)
	SetEnv(env map[string]string)
	SetTimeout(t int)
	SetDir(d string)
	String() string
	Run() (int, error)
}

type Runner struct {
	*Templator
	Render  bool
	Cmd     string
	ActionT ActionTemplate
	Dir     string
}

func NewRunner(glob, dst string, force, skipext, render bool) (*Runner, error) {
	tpl, err := NewTemplator(glob, dst, force, skipext)
	if err != nil {
		return nil, err
	}
	r := Runner{
		Templator: tpl,
		Render:    render,
	}
	return &r, nil
}

func (tr *Runner) SetRunner(ar ActionTemplate, env map[string]string, timeout int, dir string) {
	tr.Cmd = ar.String()
	tr.ActionT = ar
	tr.Env = env
	tr.Dir = dir
	tr.ActionT.SetTimeout(timeout)
	tr.ActionT.SetEnv(env)
}

func (tr *Runner) AddEnv(env map[string]string) {
	for key, value := range env {
		tr.Env[key] = value
	}
	tr.ActionT.SetEnv(tr.Env)
}

func (tr *Runner) Function(base string, path string, i os.FileInfo) (err error) {
	tpldata := tr.NewTemplateData(base, path)
	if tr.DstPath != "" {
		if tr.Render {
			if err = tr.Templator.Function(base, path, i); err != nil {
				return
			}
		}
	}
	arg, errarg := tr.renderTemplateString("arg", tr.Cmd, tpldata)
	if errarg != nil {
		return fmt.Errorf("Cannot render process arg '%s', %s", tr.Cmd, errarg)
	}
	command := strings.Fields(arg)
	// run
	homedir := tr.Dir
	if homedir == "" {
		homedir = tpldata.DestinationPath
		if tr.DstPath == "" {
			homedir = tpldata.SourcePath
		}
	}
	tr.ActionT.SetDir(homedir)
	tr.ActionT.Command(command)
	_, err = tr.ActionT.Run()
	return
}
