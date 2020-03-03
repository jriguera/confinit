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
)

// Execute is an interface to define a configurator factory
type Execute interface {
	Command(command []string)
	SetEnv(env map[string]string)
	SetTimeout(t int)
	SetDir(d string)
	String() string
	Run() (int, error)
}

type Runner struct {
	*Templator
	Render bool
	Cmd    string
	Exec   Execute
	Dir    string
}

func NewRunner(glob, dst string, force, skipext, render bool, excludes []string) (*Runner, error) {
	tpl, err := NewTemplator(glob, dst, force, skipext, excludes)
	if err != nil {
		return nil, err
	}
	r := Runner{
		Templator: tpl,
		Render:    render,
	}
	return &r, nil
}

func (tr *Runner) SetRunner(exec Execute, env map[string]string, timeout int, dir string) {
	tr.Cmd = exec.String()
	tr.Exec = exec
	tr.Env = env
	tr.Dir = dir
	tr.Exec.SetTimeout(timeout)
	tr.Exec.SetEnv(env)
}

func (tr *Runner) AddEnv(env map[string]string) {
	for key, value := range env {
		tr.Env[key] = value
	}
	tr.Exec.SetEnv(tr.Env)
}

func (tr *Runner) Function(base string, path string, i os.FileMode) (dst string, err error) {
	tpldata := tr.NewTemplateData(base, path, i)
	if tr.DstPath != "" {
		if tr.Render {
			if dst, err = tr.Templator.Function(base, path, i); err != nil {
				return
			}
		}
	}
	arg, errarg := tr.renderTemplateString("arg", tr.Cmd, tpldata)
	if errarg != nil {
		err = fmt.Errorf("Cannot render process arg '%s', %s", tr.Cmd, errarg)
		return
	}
	command := strings.Fields(arg)
	dst = arg
	// run
	homedir := tr.Dir
	if homedir == "" {
		homedir = tpldata.DestinationPath
		if tr.DstPath == "" {
			homedir = tpldata.SourcePath
		}
	}
	tr.Exec.SetDir(homedir)
	tr.Exec.Command(command)
	_, err = tr.Exec.Run()
	return
}
